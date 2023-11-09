// we have a false positive on this because of FieldPath which does not actually Hash its Cache.
// There appears to be no way to turn this off other than globally. Putting it on the struct
// does not fix things.
#![allow(clippy::mutable_key_type)]

// air module (read as: the word "air", or "A - I - R"; stands for "Aggregation IR")
mod air;
mod algebrizer;
mod ast;
pub mod catalog;
mod codegen;
#[cfg(test)]
mod internal_spec_test;
// mir module (read as: the word "mir", or "M - I -R"; stands for "MongoSQl abstract model IR")
mod mir;
pub use mir::schema::SchemaCheckingMode;
pub mod json_schema;
mod mapping_registry;
pub mod options;
mod parser;
pub mod result;
pub mod schema;
#[cfg(test)]
mod test;
mod translator;
pub mod usererror;
mod util;

use crate::{
    algebrizer::Algebrizer,
    catalog::Catalog,
    mir::schema::CachedSchema,
    options::{ExcludeNamespacesOption, SqlOptions},
    result::Result,
    schema::{Schema, SchemaEnvironment},
    translator::MqlTranslator,
};
use serde::{Deserialize, Serialize};
use std::collections::BTreeSet;

/// Contains all the information needed to execute the MQL translation of a SQL query.
#[derive(Debug)]
pub struct Translation {
    pub target_db: String,
    pub target_collection: Option<String>,
    pub pipeline: bson::Bson,
    pub result_set_schema: json_schema::Schema,
    pub select_order: Vec<Vec<String>>,
}

/// Returns the MQL translation for the provided SQL query in the
/// specified db.
pub fn translate_sql(
    current_db: &str,
    sql: &str,
    catalog: &Catalog,
    sql_options: SqlOptions,
) -> Result<Translation> {
    // parse the query and apply syntactic rewrites
    let ast = parser::parse_query(sql)?;
    let ast = ast::rewrites::rewrite_query(ast)?;
    let select_order = get_select_order(&ast);

    // construct the algebrizer and use it to build an mir plan
    let algebrizer = Algebrizer::new(current_db, catalog, 0u16, sql_options.schema_checking_mode);
    let plan = algebrizer.algebrize_query(ast)?;

    // optimizer runs
    let plan = mir::optimizer::optimize_plan(
        plan,
        sql_options.schema_checking_mode,
        &algebrizer.schema_inference_state(),
    );

    // get the schema_env for the plan
    let schema_env = plan
        .schema(&algebrizer.schema_inference_state())?
        .schema_env;

    // check for non-namespaced field name collisions if namespaces are excluded
    if sql_options.exclude_namespaces == ExcludeNamespacesOption::ExcludeNamespaces {
        schema_env.check_for_non_namespaced_collisions()?;
    }

    // construct the translator and use it to build an air plan
    let mut translator = MqlTranslator::new(sql_options);
    let agg_plan = translator.translate_plan(plan)?;

    // desugar the air plan
    let agg_plan = air::desugarer::desugar_pipeline(agg_plan)?;

    // codegen the plan into MQL
    let mql_translation = codegen::generate_mql(agg_plan)?;

    // A non-empty database value is needed for ADF
    let target_db = mql_translation
        .database
        .clone()
        .unwrap_or_else(|| current_db.to_string());

    let target_collection = mql_translation.collection;

    let pipeline = bson::Bson::Array(
        mql_translation
            .pipeline
            .into_iter()
            .map(bson::Bson::Document)
            .collect(),
    );

    let result_set_schema =
        mql_schema_env_to_json_schema(schema_env, &translator.mapping_registry, sql_options)?;
    let select_order =
        parse_select_list_order(select_order, result_set_schema.clone(), sql_options);

    Ok(Translation {
        target_db,
        target_collection,
        pipeline,
        result_set_schema,
        select_order,
    })
}

#[derive(Serialize, Deserialize, PartialEq, Eq, Clone, Debug, PartialOrd, Ord)]
pub struct Namespace {
    pub database: String,
    pub collection: String,
}

pub fn get_namespaces(current_db: &str, sql: &str) -> Result<BTreeSet<Namespace>> {
    let ast = parser::parse_query(sql)?;
    let namespaces = ast::visitors::get_collection_sources(ast)
        .into_iter()
        .map(|cs| Namespace {
            database: cs.database.unwrap_or_else(|| current_db.to_string()),
            collection: cs.collection,
        })
        .collect();
    Ok(namespaces)
}

// get_select_order uses pattern matching to parse the select body from the rewritten AST.
// Only parse if it is a non-distinct SelectQuery, otherwise return none (we won't attempt reordering).
pub fn get_select_order(ast: &ast::Query) -> Option<ast::SelectBody> {
    match ast {
        ast::Query::Select(ast::SelectQuery {
            select_clause:
                ast::SelectClause {
                    set_quantifier: ast::SetQuantifier::All,
                    body: b,
                },
            ..
        }) => Some(b.clone()),
        _ => None,
    }
}

// given the select order, produce a list representing each field of the result set.
// if we are including namespaces, we will model the fields as a list of [namespace, field name].
// if we are excluding namespaces, we will model the fields as a list of [field name].
pub fn parse_select_list_order(
    select_body: Option<ast::SelectBody>,
    result_set_schema: json_schema::Schema,
    sql_options: SqlOptions,
) -> Vec<Vec<String>> {
    let mut select_order: Vec<Vec<String>> = vec![];
    if let Some(body) = select_body {
        match body {
            ast::SelectBody::Values(values) => {
                let mut properties = result_set_schema.properties.unwrap_or_default();
                for value in values {
                    match value {
                        ast::SelectValuesExpression::Expression(e) => {
                            if let ast::Expression::Document(d) = e {
                                for document_pair in d {
                                    select_order.push(match sql_options.exclude_namespaces {
                                        ExcludeNamespacesOption::ExcludeNamespaces => {
                                            vec![document_pair.key]
                                        }
                                        ExcludeNamespacesOption::IncludeNamespaces => {
                                            vec!["".into(), document_pair.key]
                                        }
                                    });
                                }
                            }
                        }
                        ast::SelectValuesExpression::Substar(s) => {
                            // for substar, get the whole namespace from the result set schema, and parse out all field names
                            let mut fields = properties
                                .remove(&s.datasource)
                                .unwrap_or_default()
                                .properties
                                .unwrap_or_default()
                                .keys()
                                .map(|field| match sql_options.exclude_namespaces {
                                    ExcludeNamespacesOption::ExcludeNamespaces => {
                                        vec![field.clone()]
                                    }
                                    ExcludeNamespacesOption::IncludeNamespaces => {
                                        vec![s.datasource.clone(), field.clone()]
                                    }
                                })
                                .collect::<Vec<_>>();
                            // we sort substar keys to ensure they have a deterministic order
                            fields.sort();
                            select_order.append(&mut fields);
                        }
                    }
                }
            }
            // this option should only occur if we have a select *. Return all fields in sorted order
            ast::SelectBody::Standard(_) => {
                result_set_schema
                    .properties
                    .unwrap_or_default()
                    .into_iter()
                    .for_each(|(name, schema)| match sql_options.exclude_namespaces {
                        ExcludeNamespacesOption::ExcludeNamespaces => select_order.push(vec![name]),
                        ExcludeNamespacesOption::IncludeNamespaces => schema
                            .properties
                            .unwrap_or_default()
                            .into_iter()
                            .for_each(|(field, _)| select_order.push(vec![name.clone(), field])),
                    });
                select_order.sort();
            }
        }
    }
    select_order
}

// mql_schema_env_to_json_schema converts a SchemaEnvironment to a json_schema::Schema with an
// MqlMappingRegistry. It uses `SqlOptions` to determine whether to include namespaces in the schema or
// to remove them and instead pull the keys in the namespaces up a level in the schema.
// It will not work with any other codegen backends.
fn mql_schema_env_to_json_schema(
    schema_env: SchemaEnvironment,
    mapping_registry: &codegen::MqlMappingRegistry,
    sql_options: SqlOptions,
) -> Result<json_schema::Schema> {
    let keys: std::collections::BTreeMap<String, Schema> =
        if sql_options.exclude_namespaces == ExcludeNamespacesOption::IncludeNamespaces {
            include_namespace_in_result_set_schema_keys(schema_env, mapping_registry)?
        } else {
            exclude_namespace_in_result_set_schema_keys(schema_env, mapping_registry)?
        };

    json_schema::Schema::try_from(Schema::simplify(&Schema::Document(schema::Document {
        required: keys.keys().cloned().collect(),
        keys,
        additional_properties: false,
    })))
    .map_err(result::Error::JsonSchemaConversion)
}

fn include_namespace_in_result_set_schema_keys(
    schema_env: SchemaEnvironment,
    mapping_registry: &codegen::MqlMappingRegistry,
) -> Result<std::collections::BTreeMap<String, Schema>> {
    schema_env
        .into_iter()
        .map(|(k, v)| {
            let registry_value = mapping_registry.get(&k);
            match registry_value {
                Some(registry_value) => Ok((registry_value.name.clone(), v)),
                None => Err(result::Error::Translator(
                    translator::Error::ReferenceNotFound(k),
                )),
            }
        })
        .collect::<Result<std::collections::BTreeMap<String, Schema>>>()
}

fn exclude_namespace_in_result_set_schema_keys(
    schema_env: SchemaEnvironment,
    mapping_registry: &codegen::MqlMappingRegistry,
) -> Result<std::collections::BTreeMap<String, Schema>> {
    schema_env
        .into_iter()
        .flat_map(|(k, v)| {
            if mapping_registry.get(&k).is_some() {
                if let Schema::Document(doc) = v {
                    doc.keys
                        .into_iter()
                        .map(|(key, schema)| Ok((key, schema)))
                        .collect::<Vec<_>>()
                } else {
                    vec![Err(result::Error::Translator(
                        translator::Error::DocumentSchemaTypeNotFound(v),
                    ))]
                }
            } else {
                vec![Err(result::Error::Translator(
                    translator::Error::ReferenceNotFound(k),
                ))]
            }
        })
        .collect::<Result<std::collections::BTreeMap<String, Schema>>>()
}
