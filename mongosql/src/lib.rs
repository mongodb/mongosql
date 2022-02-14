mod algebrizer;
mod ast;
pub mod catalog;
mod codegen;
#[cfg(test)]
mod internal_spec_test;
mod ir;
pub use ir::schema::SchemaCheckingMode;
pub mod json_schema;
mod parser;
mod result;
pub mod schema;
mod util;
use crate::{
    algebrizer::Algebrizer,
    catalog::Catalog,
    ir::schema::CachedSchema,
    parser::Parser,
    result::Result,
    schema::{Schema, SchemaEnvironment},
};

/// Contains all the information needed to execute the MQL translation of a SQL query.
#[derive(Debug)]
pub struct Translation {
    pub target_db: String,
    pub target_collection: Option<String>,
    pub pipeline: bson::Bson,
    pub result_set_schema: json_schema::Schema,
    pub namespaces: Vec<ir::namespace::Namespace>,
}

/// Returns the MQL translation for the provided SQL query in the
/// specified db.
pub fn translate_sql(
    current_db: &str,
    sql: &str,
    catalog: &Catalog,
    schema_checking_mode: SchemaCheckingMode,
) -> Result<Translation> {
    // parse the query and apply syntactic rewrites
    let p = Parser::new();
    let ast = p.parse_query(sql)?;
    let ast = ast::rewrites::rewrite_query(ast)?;

    // construct the algebrizer and use it to build an ir plan
    let algebrizer = Algebrizer::new(current_db, catalog, 0u16, schema_checking_mode);
    let plan = algebrizer.algebrize_query(ast)?;

    // flatten variadic function
    let plan = ir::flatten::flatten_variadic_functions(plan);

    // constant fold stages
    let plan = ir::constant_folding::fold_constants(plan, schema_checking_mode);

    let namespaces = ir::namespace::get_namespaces(plan.clone());

    // get the schema_env for the plan
    let schema_env = plan
        .schema(&algebrizer.schema_inference_state())?
        .schema_env;

    // generate mql from the ir plan
    let mql_translation = codegen::generate_mql(plan)?;

    // A non-empty database value is needed for mongoast
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
        mql_schema_env_to_json_schema(schema_env, &mql_translation.mapping_registry)?;

    Ok(Translation {
        target_db,
        target_collection,
        pipeline,
        result_set_schema,
        namespaces,
    })
}

// mql_schema_env_to_json_schema converts a SchemaEnvironment to a json_schema::Schema with an
// MqlMappingRegistry.  It will not work with any other codegen backends
fn mql_schema_env_to_json_schema(
    schema_env: SchemaEnvironment,
    mapping_registry: &codegen::MqlMappingRegistry,
) -> Result<json_schema::Schema> {
    let keys: std::collections::BTreeMap<String, Schema> = schema_env
        .into_iter()
        .map(|(k, v)| {
            let mql_name = mapping_registry.get(&k);
            match mql_name {
                Some(mql_name) => Ok((mql_name.clone(), v)),
                None => Err(result::Error::Codegen(codegen::Error::ReferenceNotFound(k))),
            }
        })
        .collect::<Result<_>>()?;
    json_schema::Schema::try_from(Schema::simplify(&Schema::Document(schema::Document {
        required: keys.keys().cloned().collect(),
        keys,
        additional_properties: false,
    })))
    .map_err(result::Error::JsonSchemaConversion)
}
