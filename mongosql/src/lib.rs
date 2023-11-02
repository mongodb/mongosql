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
    json_schema::{BsonType, BsonTypeName},
    mir::schema::CachedSchema,
    options::{ExcludeNamespacesOption, SqlOptions},
    result::Result,
    schema::{Schema, SchemaEnvironment},
    translator::MqlTranslator,
};
use serde::{Deserialize, Serialize};
use std::collections::{BTreeSet, HashMap};

/// Contains all the information needed to execute the MQL translation of a SQL query.
#[derive(Debug)]
pub struct Translation {
    pub target_db: String,
    pub target_collection: Option<String>,
    pub pipeline: bson::Bson,
    pub result_set_schema: json_schema::Schema,
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
    let result_set_schema =
        order_result_set_metadata(result_set_schema, select_order, sql_options)?;

    Ok(Translation {
        target_db,
        target_collection,
        pipeline,
        result_set_schema,
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

// order_result_set_metadata takes the select body parsed from the AST, and uses it to
// order the result_set_metadata by select order (result set metadata is in an arbitrary order at this point).
pub fn order_result_set_metadata(
    result_set_schema: json_schema::Schema,
    select_order: Option<ast::SelectBody>,
    sql_options: SqlOptions,
) -> Result<json_schema::Schema> {
    let body = if let Some(body) = select_order {
        body
    } else {
        return Ok(result_set_schema);
    };
    match body {
        ast::SelectBody::Values(v) => {
            // to simplify lookups, create two maps: one mapping namespaces to schemas (helpful for substars), and one mapping the bottom fields to their schemas (helpful for expressions)
            let (mut namespace_to_schema, mut bottom_fields) = match sql_options.exclude_namespaces
            {
                ExcludeNamespacesOption::ExcludeNamespaces => {
                    // if we are not using namespaces, all fields will be at the top level, and we won't need a map from namespaces to schema
                    (
                        HashMap::new(),
                        HashMap::from_iter(result_set_schema.properties),
                    )
                }
                ExcludeNamespacesOption::IncludeNamespaces => {
                    let mut schema_map = HashMap::from_iter(result_set_schema.properties);
                    let bottom_fields = match schema_map.remove("") {
                        None => HashMap::new(),
                        Some(s) => HashMap::from_iter(s.properties),
                    };
                    (schema_map, bottom_fields)
                }
            };
            let mut ordered_properties: Vec<(String, json_schema::Schema)> = Vec::new();
            for expr in v {
                match expr {
                    // if we have an expression, get it from bottom and add it to the ordered metadata
                    ast::SelectValuesExpression::Expression(e) => {
                        if let ast::Expression::Document(d) = e {
                            for document_pair in d {
                                let schema = bottom_fields.remove(&document_pair.key).ok_or(
                                    result::Error::JsonSchemaConversion(
                                        schema::Error::InvalidBottomField(
                                            document_pair.key.clone(),
                                        ),
                                    ),
                                )?;
                                match sql_options.exclude_namespaces {
                                    ExcludeNamespacesOption::ExcludeNamespaces => {
                                        ordered_properties.push((document_pair.key, schema));
                                    }
                                    ExcludeNamespacesOption::IncludeNamespaces => {
                                        ordered_properties.push((
                                            "".to_string(),
                                            json_schema::Schema {
                                                bson_type: Some(BsonType::Single(
                                                    BsonTypeName::Object,
                                                )),
                                                properties: vec![(document_pair.key, schema)],
                                                ..Default::default()
                                            },
                                        ));
                                    }
                                }
                            }
                        }
                    }
                    // for substars, take the whole schema for the namespaces and add it to the metadata
                    ast::SelectValuesExpression::Substar(s) => {
                        let schema = namespace_to_schema.remove(&s.datasource).ok_or(
                            result::Error::JsonSchemaConversion(schema::Error::InvalidNamespace(
                                s.datasource.clone(),
                            )),
                        )?;
                        ordered_properties.push((s.datasource, schema));
                    }
                }
            }
            Ok(json_schema::Schema {
                properties: ordered_properties,
                ..result_set_schema
            })
        }
        // if we get a Standard select body after AST rewrites, this should just be select *. No need to update
        _ => Ok(result_set_schema),
    }
}
