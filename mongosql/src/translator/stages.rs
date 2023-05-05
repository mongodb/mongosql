use crate::{
    air,
    mapping_registry::{Key, MqlMappingRegistry, MqlMappingRegistryValue, MqlReferenceType},
    mir,
    translator::{utils::ROOT, Error, MqlTranslator, Result},
};
use mongosql_datastructures::{
    unique_linked_hash_map, unique_linked_hash_map::UniqueLinkedHashMap,
};
use std::collections::BTreeSet;

impl MqlTranslator {
    pub(crate) fn translate_stage(&mut self, mir_stage: mir::Stage) -> Result<air::Stage> {
        match mir_stage {
            mir::Stage::Filter(f) => self.translate_filter(f),
            mir::Stage::Project(p) => self.translate_project(p),
            mir::Stage::Group(g) => self.translate_group(g),
            mir::Stage::Limit(l) => self.translate_limit(l),
            mir::Stage::Offset(o) => self.translate_offset(o),
            mir::Stage::Sort(s) => self.translate_sort(s),
            mir::Stage::Collection(c) => self.translate_collection(c),
            mir::Stage::Array(arr) => self.translate_array_stage(arr),
            mir::Stage::Join(j) => self.translate_join(j),
            mir::Stage::Set(s) => self.translate_set(s),
            mir::Stage::Derived(d) => self.translate_derived(d),
            mir::Stage::Unwind(u) => self.translate_unwind(u),
        }
    }

    fn translate_filter(&mut self, mir_filter: mir::Filter) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_filter.source)?;
        let expr_translation = self.translate_expression(mir_filter.condition)?;

        Ok(air::Stage::Match(air::Match {
            source: Box::new(source_translation),
            expr: Box::new(expr_translation),
        }))
    }

    fn translate_project(&mut self, mir_project: mir::Project) -> Result<air::Stage> {
        // Store the incoming mapping registry so we can maintain references
        // to outer scopes in the output registry.
        let mut outer_scope_mr = self.mapping_registry.clone();

        // When we translate expressions in this project, we'll also want to
        // maintain references to outer scopes, so we clone the translator.
        //
        // The reason we need expr_translator AND outer_scope_mr is that the
        // expr_translator must be extended with the mapping registry from the
        // source of this Project, which are only in scope for the translation
        // of this stage, whereas the outer_scope_mr is not extended with those
        // mappings. Instead, outer_scope_mr is extended with the new mappings
        // created by this Project's translation.
        let mut expr_translator = self.clone();

        // Translate the source stage.
        let source_translation = self.translate_stage(*mir_project.source)?;

        // Extend the expr_translator with mappings from the source. This means
        // we can use expr_translator to translate expressions in this Project
        // because expr_translator contains mappings from outer scopes as well
        // as from this Project's source.
        expr_translator
            .mapping_registry
            .merge(self.mapping_registry.clone());

        // We will add mappings to the mapping registry introduced by this Project Stage, which
        // is all of the keys. Previous bindings are removed after the subexpressions are
        // translated because Project kills all its inputs.
        let unique_bot_name = Self::get_unique_bot_name(&mir_project.expression);
        let mut project_body = UniqueLinkedHashMap::new();
        let mut output_registry = MqlMappingRegistry::new();
        for (k, e) in mir_project.expression.into_iter() {
            let mapped_k = Self::get_datasource_name(&k.datasource, &unique_bot_name);
            if mapped_k.starts_with('$') || mapped_k.contains('.') || mapped_k.as_str() == "" {
                return Err(Error::InvalidProjectField);
            }

            project_body.insert(
                mapped_k.clone(),
                air::ProjectItem::Assignment(expr_translator.translate_expression(e)?),
            )?;
            output_registry.insert(
                k,
                MqlMappingRegistryValue::new(mapped_k, MqlReferenceType::FieldRef),
            );
        }

        // Extend the outer_scope_mr with the output mappings for this stage's
        // translation.
        outer_scope_mr.merge(output_registry);

        self.mapping_registry = outer_scope_mr;
        Ok(air::Stage::Project(air::Project {
            source: Box::new(source_translation),
            specifications: project_body,
        }))
    }

    fn translate_group(&mut self, mir_group: mir::Group) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_group.source)?;

        // specifications are the top level projection fields. Every key will
        // be a Datasource. Unaliased group keys will need to be inserted
        // directly into the specifications with the proper Datasource name.
        let mut specifications = UniqueLinkedHashMap::new();
        // bot_body is a Document containing all the field mappings under the Bot Datasource.
        // These will be all aliased group keys and all the aggregations.
        let mut bot_body = UniqueLinkedHashMap::new();

        // map the group key aliases and translate the expressions. Unliased keys will be Projected
        // straight into the specifications
        let keys = self.translate_group_keys(mir_group.keys, &mut specifications, &mut bot_body)?;

        // map the group aggregation aliases and translate the expresisons.
        let aggregations =
            self.translate_group_aggregations(mir_group.aggregations, &mut bot_body)?;

        // Fixup mapping_registry to contain only the output of the final Project Stage for this
        // GROUP BY.
        self.mapping_registry = MqlMappingRegistry::new();

        for key in specifications.keys() {
            self.mapping_registry.insert(
                Key::named(key, self.scope_level),
                MqlMappingRegistryValue::new(key.clone(), MqlReferenceType::FieldRef),
            );
        }
        // bot_body will only be empty if all Group keys are references and there are not
        // aggregation functions. If it is not empty, we need to add it to the output Project
        // Stage under the unique bot name.
        if !bot_body.is_empty() {
            let unique_bot_name =
                Self::generate_unique_bot_name(|s| specifications.contains_key(s));
            specifications.insert(
                unique_bot_name.clone(),
                air::ProjectItem::Assignment(air::Expression::Document(bot_body)),
            )?;
            self.mapping_registry.insert(
                Key::bot(self.scope_level),
                MqlMappingRegistryValue::new(unique_bot_name, MqlReferenceType::FieldRef),
            );
        }

        // Return the proper Stages, which is a Project with specifications set as above and the
        // source being the generated Group Stage.
        Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Group(air::Group {
                source: Box::new(source_translation),
                keys,
                aggregations,
            })),
            specifications,
        }))
    }

    fn translate_limit(&mut self, mir_limit: mir::Limit) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_limit.source)?;
        let limit =
            i64::try_from(mir_limit.limit).or(Err(Error::LimitOutOfI64Range(mir_limit.limit)))?;

        Ok(air::Stage::Limit(air::Limit {
            source: Box::new(source_translation),
            limit,
        }))
    }

    fn translate_offset(&mut self, mir_offset: mir::Offset) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_offset.source)?;

        Ok(air::Stage::Skip(air::Skip {
            source: Box::new(source_translation),
            skip: mir_offset.offset,
        }))
    }

    fn translate_sort(&mut self, mir_sort: mir::Sort) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_sort.source)?;
        let specifications = mir_sort
            .specs
            .into_iter()
            .map(|spec| {
                Ok(match spec {
                    mir::SortSpecification::Asc(mir_expr) => {
                        air::SortSpecification::Asc(self.get_reference_key_name(*mir_expr)?)
                    }
                    mir::SortSpecification::Desc(mir_expr) => {
                        air::SortSpecification::Desc(self.get_reference_key_name(*mir_expr)?)
                    }
                })
            })
            .collect::<Result<Vec<air::SortSpecification>>>()?;
        Ok(air::Stage::Sort(air::Sort {
            source: source_translation.into(),
            specs: specifications,
        }))
    }

    fn translate_collection(&mut self, mir_collection: mir::Collection) -> Result<air::Stage> {
        let coll_stage = air::Stage::Collection(air::Collection {
            db: mir_collection.db,
            collection: mir_collection.collection.clone(),
        });

        self.mapping_registry.insert(
            Key::named(&mir_collection.collection, self.scope_level),
            MqlMappingRegistryValue::new(
                mir_collection.collection.clone(),
                MqlReferenceType::FieldRef,
            ),
        );
        Ok(air::Stage::Project(air::Project {
            source: Box::new(coll_stage),
            specifications: unique_linked_hash_map! {
                mir_collection.collection => air::ProjectItem::Assignment(ROOT.clone()),
            },
        }))
    }

    fn translate_array_stage(&mut self, mir_arr: mir::ArraySource) -> Result<air::Stage> {
        let doc_stage = air::Stage::Documents(air::Documents {
            array: mir_arr
                .array
                .iter()
                .map(|mir_expr| self.translate_expression(mir_expr.clone()))
                .collect::<Result<Vec<air::Expression>>>()?,
        });

        self.mapping_registry.insert(
            Key::named(&mir_arr.alias, self.scope_level),
            MqlMappingRegistryValue::new(mir_arr.alias.clone(), MqlReferenceType::FieldRef),
        );

        Ok(air::Stage::Project(air::Project {
            source: Box::new(doc_stage),
            specifications: unique_linked_hash_map! {
                mir_arr.alias => air::ProjectItem::Assignment(ROOT.clone()),
            },
        }))
    }

    fn translate_join(&mut self, mir_join: mir::Join) -> Result<air::Stage> {
        let join_type = match mir_join.join_type {
            mir::JoinType::Inner => air::JoinType::Inner,
            mir::JoinType::Left => air::JoinType::Left,
        };

        // We need to store the left_registry to generate the let bindings
        // since only datasources from the left will be bound to variables.
        let left = self.translate_stage(*mir_join.left)?;
        let left_registry = self.mapping_registry.clone();
        let right = self.translate_stage(*mir_join.right)?;

        let mut let_vars = None;
        let condition = mir_join
            .condition
            .map(|x| {
                let_vars = Some(self.generate_let_bindings(left_registry));
                self.translate_expression(x)
            })
            .transpose()?;

        Ok(air::Stage::Join(air::Join {
            join_type,
            left: Box::new(left),
            right: Box::new(right),
            let_vars,
            condition,
        }))
    }

    fn translate_set(&mut self, mir_set: mir::Set) -> Result<air::Stage> {
        let source = self.translate_stage(*mir_set.left)?;
        let pipeline = self.translate_stage(*mir_set.right)?;

        Ok(air::Stage::UnionWith(air::UnionWith {
            source: Box::new(source),
            pipeline: Box::new(pipeline),
        }))
    }

    fn translate_derived(&mut self, mir_derived: mir::Derived) -> Result<air::Stage> {
        self.scope_level += 1;
        let derived_translation = self.translate_stage(*mir_derived.source)?;
        self.scope_level -= 1;
        Ok(derived_translation)
    }

    fn translate_unwind(&mut self, mir_unwind: mir::Unwind) -> Result<air::Stage> {
        Ok(air::Stage::Unwind(air::Unwind {
            source: Box::new(self.translate_stage(*mir_unwind.source)?),
            path: Box::new(self.translate_expression(*mir_unwind.path)?),
            index: mir_unwind.index,
            outer: mir_unwind.outer,
        }))
    }

    // Utility functions for Group translation
    fn get_unique_alias(existing_aliases: &BTreeSet<String>, mut alias: String) -> String {
        while existing_aliases.contains(&alias) {
            alias.insert(0, '_')
        }
        alias
    }

    fn translate_agg_function(afa: mir::AggregationFunction) -> air::AggregationFunction {
        match afa {
            mir::AggregationFunction::AddToArray => air::AggregationFunction::AddToArray,
            mir::AggregationFunction::Avg => air::AggregationFunction::Avg,
            mir::AggregationFunction::Count => air::AggregationFunction::Count,
            mir::AggregationFunction::First => air::AggregationFunction::First,
            mir::AggregationFunction::Last => air::AggregationFunction::Last,
            mir::AggregationFunction::Max => air::AggregationFunction::Max,
            mir::AggregationFunction::MergeDocuments => air::AggregationFunction::MergeDocuments,
            mir::AggregationFunction::Min => air::AggregationFunction::Min,
            mir::AggregationFunction::StddevPop => air::AggregationFunction::StddevPop,
            mir::AggregationFunction::StddevSamp => air::AggregationFunction::StddevSamp,
            mir::AggregationFunction::Sum => air::AggregationFunction::Sum,
        }
    }

    fn get_datasource_and_field_for_unaliased_group_key<'a>(
        &'a self,
        e: &'a mir::Expression,
    ) -> Result<(&'a MqlMappingRegistryValue, &'a String)> {
        let (key, field) = match e {
            mir::Expression::FieldAccess(mir::FieldAccess {
                ref expr,
                ref field,
                ..
            }) => match **expr {
                mir::Expression::Reference(mir::ReferenceExpr { ref key, .. }) => (key, field),
                _ => return Err(Error::InvalidGroupKey),
            },
            _ => return Err(Error::InvalidGroupKey),
        };
        Ok((
            self.mapping_registry
                .get(key)
                .ok_or_else(|| Error::ReferenceNotFound(key.clone()))?,
            field,
        ))
    }

    fn translate_group_keys(
        &self,
        keys: Vec<mir::OptionallyAliasedExpr>,
        specifications: &mut UniqueLinkedHashMap<String, air::ProjectItem>,
        bot_body: &mut UniqueLinkedHashMap<String, air::Expression>,
    ) -> Result<Vec<air::NameExprPair>> {
        let unique_aliases = keys
            .iter()
            .filter_map(|k| k.get_alias().map(String::from))
            .collect::<BTreeSet<_>>();

        let mut translated_keys = Vec::new();

        let make_key_ref = |name| air::Expression::FieldRef(format!("_id.{}", name).into());
        for (i, k) in keys.into_iter().enumerate() {
            match k {
                mir::OptionallyAliasedExpr::Aliased(ae) => {
                    // an aliased key will be projected under Bot
                    translated_keys.push(air::NameExprPair {
                        name: ae.alias.clone(),
                        expr: self.translate_expression(ae.expr)?,
                    });
                    bot_body.insert(ae.alias.clone(), make_key_ref(ae.alias))?;
                }
                // An unaliased key will be projected under its originating Datasource.
                // After the Aliasing rewrite pass, an unaliased key can only exist when it is a
                // FieldAccess, and the parent of the FieldAccess *must* be a Datasource.
                mir::OptionallyAliasedExpr::Unaliased(e) => {
                    let position_counter = i + 1;
                    let unique_name = Self::get_unique_alias(
                        &unique_aliases,
                        format!("__unaliasedKey{position_counter}"),
                    );
                    let (datasource, field) =
                        self.get_datasource_and_field_for_unaliased_group_key(&e)?;
                    match specifications.get_mut(&datasource.name) {
                        // If we have already put something under this Datasource, we just update
                        // the document.
                        Some(air::ProjectItem::Assignment(air::Expression::Document(ref mut d))) => {
                            d.insert(field.clone(), make_key_ref(unique_name.clone()))?
                        }
                        // We have nothing under this Datasource, so we need to create a new
                        // Document with one key/value pair.
                        None => specifications.insert(
                            datasource.name.clone(),
                            air::ProjectItem::Assignment(air::Expression::Document(unique_linked_hash_map! {field.clone() => make_key_ref(unique_name.clone())})),
                        )?,
                        // We only generate Project fields where the values are Documents because the keys of
                        // the Project are Datasources. If the following case is hit, it is a bug in
                        // the code.
                        _ => unreachable!(),
                    }
                    translated_keys.push(air::NameExprPair {
                        name: unique_name,
                        expr: self.translate_expression(e)?,
                    });
                }
            }
        }

        Ok(translated_keys)
    }

    fn translate_group_aggregations(
        &self,
        aggregations: Vec<mir::AliasedAggregation>,
        bot_body: &mut UniqueLinkedHashMap<String, air::Expression>,
    ) -> Result<Vec<air::AccumulatorExpr>> {
        let mut unique_aliases = aggregations
            .iter()
            .map(|k| k.alias.clone())
            .collect::<BTreeSet<_>>();
        // the Group keys will be under _id in MQL, so we need to rename any alias that is _id.
        unique_aliases.insert("_id".to_string());

        let mut translated_aggregations = Vec::new();

        for a in aggregations.into_iter() {
            let alias = a.alias;
            let unique_alias = if alias.as_str() == "_id" {
                Self::get_unique_alias(&unique_aliases, "_id".to_string())
            } else {
                alias.clone()
            };
            let (function, distinct, arg) = match a.agg_expr {
                mir::AggregationExpr::CountStar(b) => (
                    air::AggregationFunction::Count,
                    b,
                    Box::new(air::Expression::Literal(air::LiteralValue::Integer(1))),
                ),
                mir::AggregationExpr::Function(afa) => (
                    Self::translate_agg_function(afa.function),
                    afa.distinct,
                    Box::new(self.translate_expression(*afa.arg)?),
                ),
            };
            bot_body.insert(
                alias.clone(),
                air::Expression::FieldRef(unique_alias.clone().into()),
            )?;
            translated_aggregations.push(air::AccumulatorExpr {
                alias: unique_alias,
                function,
                distinct,
                arg,
            });
        }

        Ok(translated_aggregations)
    }
}
