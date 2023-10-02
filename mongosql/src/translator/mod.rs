use crate::{
    air,
    mapping_registry::{Key, MqlMappingRegistry, MqlMappingRegistryValue, MqlReferenceType},
    mir,
    options::{ExcludeNamespacesOption, SqlOptions},
    util::ROOT,
};
use itertools::Itertools;
use mongosql_datastructures::{
    binding_tuple::DatasourceName, unique_linked_hash_map::DuplicateKeyError,
};
use thiserror::Error;

#[cfg(test)]
mod test;

mod expressions;
mod match_query;
mod stages;
mod utils;

type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("invalid document key '{0}': document keys may not be empty, contain dots, or start with dollars")]
    InvalidDocumentKey(String),
    #[error("binding tuple key {0:?} not found in mapping registry")]
    ReferenceNotFound(Key),
    #[error("duplicate key found: {0}")]
    DuplicateKey(#[from] DuplicateKeyError),
    #[error("project fields may not be empty")]
    InvalidProjectField,
    #[error("invalid group key, unaliased key must be field ref")]
    InvalidGroupKey,
    #[error("invalid sqlConvert target type: {0:?}")]
    InvalidSqlConvertToType(air::Type),
    #[error("unexpected expr type for sort: not a FieldAccess or Reference")]
    ExprNotReferenceOrFieldAccess,
    #[error("LIMIT ({0}) cannot be converted to i64")]
    LimitOutOfI64Range(u64),
    #[error("expected FieldRef for subquery output path")]
    SubqueryOutputPathNotFieldRef,
    #[error("expected FieldRef")]
    ExpectedFieldRef,
    #[error("expected Collection source on RHS of EquiJoin")]
    ExpectedCollection,
    #[error("expected FieldRef for match-language input, but got Variable")]
    InvalidMatchLanguageInputRef,
    #[error("expected ROOT-based Variable for EquiJoin foreign field, but got {0}")]
    InvalidEquiJoinForeignFieldRef(String),
}

#[derive(Clone)]
pub struct MqlTranslator {
    pub mapping_registry: MqlMappingRegistry,
    pub scope_level: u16,
    pub sql_options: SqlOptions,
    is_join: bool,
}

impl MqlTranslator {
    pub fn new(sql_options: SqlOptions) -> Self {
        Self {
            mapping_registry: Default::default(),
            scope_level: 0u16,
            sql_options,
            is_join: false,
        }
    }

    /// translate_plan is the entry point, it mostly just calls translate_stage,
    /// but depending on desired namespace results will change how data is returned
    /// to users.
    ///
    /// If sql_options contains a [`mongosql::options::ExcludeNamespacesOption`]
    /// that is set to [`mongosql::options::ExcludeNamespacesOption::ExcludeNamespaces`],
    /// it will merge all documents togther.
    /// If the value is [`mongosql::options::ExcludeNamespacesOption::IncludeNamespaces`],
    /// it will set up a ReplaceWith to replace __bot with the empty key: ''.
    pub fn translate_plan(&mut self, mir_stage: mir::Stage) -> Result<air::Stage> {
        let source = self.translate_stage(mir_stage)?;
        if self.sql_options.exclude_namespaces == ExcludeNamespacesOption::ExcludeNamespaces {
            self.append_unnest_stage(source)
        } else {
            self.append_name_replacements(source)
        }
    }

    /// append_name_replacements adds a stage that replaces any datasource
    /// aliases that used different names in the translation. Specifically,
    /// Bottom uses a unique name that typically looks like "__bot" with
    /// potentially more preceding '_'s. It must be replaced with empty string
    /// "" if present in the mapping registry. Additionally, Named datasources
    /// that start with '$' or contain '.'s use sanitized names that replace
    /// those characters with '_'. They must be replaced with their original
    /// names. Any other Named datasources that encountered conflicts as a
    /// result of sanitization must also be replaced with their original names.
    ///
    /// Example: SELECT * FROM `$foo`, `bar.baz`, `$_foo`, `_foo`, `bar`
    /// - `$foo` is mapped to `_foo`
    /// - `bar.baz` is mapped to `bar_baz`
    /// - `$_foo` is mapped to `__foo`
    /// - `_foo` is mapped to `___foo`
    /// - `bar` is mapped to `bar`
    /// So the first 4 datasources must have their original names restored since
    /// the translation otherwise referred to them using their mapped names.
    fn append_name_replacements(&mut self, source: air::Stage) -> Result<air::Stage> {
        // We must create a new registry since we do a mutable borrow of
        // self.mapping_registry below for iteration and update the registry
        // during that iteration.
        let mut new_registry = MqlMappingRegistry::new();
        let mut replacement_expr = ROOT.clone();
        let mut needs_replacement = false;

        // Iterate through in descending alphabetical order. This is because a
        // mapped name either just replaces '$'/'.'s with '_' without conflict
        // OR fixes conflicts by prepending '_'. More '_' characters indicate
        // possible conflicts with other mapped or original names. Names with
        // possible conflicts must be replaced last so that the conflicting
        // original name is not overwritten. Again, refer to the example in
        // the function-level comment to see how this works.
        for (og_name, mapped_name) in self
            .mapping_registry
            .get_registry()
            .iter()
            .sorted_by(|(_, v1), (_, v2)| Ord::cmp(&v2.name, &v1.name))
        {
            let (name_to_unset, name_to_set) = match &og_name.datasource {
                DatasourceName::Bottom => {
                    needs_replacement = true;
                    (mapped_name.name.clone(), "".to_string())
                }
                DatasourceName::Named(name) => {
                    if *name != *mapped_name.name {
                        needs_replacement = true;
                        (mapped_name.name.clone(), name.clone())
                    } else {
                        // Retain all mappings, even ones without conflicts.
                        new_registry.insert(og_name.clone(), mapped_name.clone());
                        continue;
                    }
                }
            };

            // Update the mapping registry.
            new_registry.insert(
                og_name.clone(),
                MqlMappingRegistryValue::new(name_to_set.clone(), mapped_name.ref_type.clone()),
            );

            // Update the replacement expression to unset the mapped named
            // and set the "new" name.
            replacement_expr = air::Expression::UnsetField(air::UnsetField {
                field: name_to_unset.clone(),
                input: Box::new(air::Expression::SetField(air::SetField {
                    field: name_to_set,
                    input: Box::new(replacement_expr),
                    value: Box::new(air::Expression::FieldRef(name_to_unset.into())),
                })),
            })
        }

        if needs_replacement {
            self.mapping_registry = new_registry;
            Ok(air::Stage::ReplaceWith(air::ReplaceWith {
                source: Box::new(source),
                new_root: Box::new(replacement_expr),
            }))
        } else {
            Ok(source)
        }
    }

    /// append_unnest_stage goes through the mapping registry, getting all variables and references
    /// representing datasources and flattening them into a vec.
    /// If the length of the resultant vec is 1, then we only had one datasource and we generate a flat
    /// ReplaceWith. If there was more than one, we generate a ReplaceWith with a MergeObjects.
    /// This ultimately results in all namespaces being removed from the results the users will see
    fn append_unnest_stage(
        &mut self,
        source: air::Stage,
    ) -> std::result::Result<air::Stage, Error> {
        let namespaces = self
            .mapping_registry
            .get_registry()
            .iter()
            .map(|(_, v)| match v.ref_type {
                MqlReferenceType::FieldRef => air::Expression::FieldRef(v.name.clone().into()),
                MqlReferenceType::Variable => air::Expression::Variable(v.name.clone().into()),
            })
            .collect::<Vec<air::Expression>>();

        Ok(air::Stage::ReplaceWith(air::ReplaceWith {
            source: Box::new(source),
            new_root: Box::new(if namespaces.len() == 1 {
                namespaces.get(0).unwrap().clone()
            } else {
                air::Expression::MQLSemanticOperator(air::MQLSemanticOperator {
                    op: air::MQLOperator::MergeObjects,
                    args: namespaces,
                })
            }),
        }))
    }
}
