mod agg_ir;
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
mod translator;
mod util;

use crate::{
    algebrizer::Algebrizer,
    catalog::Catalog,
    ir::schema::CachedSchema,
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
    let ast = parser::parse_query(sql)?;
    let ast = ast::rewrites::rewrite_query(ast)?;

    // construct the algebrizer and use it to build an ir plan
    let algebrizer = Algebrizer::new(current_db, catalog, 0u16, schema_checking_mode);
    let plan = algebrizer.algebrize_query(ast)?;

    // flatten variadic function
    let plan = ir::flatten::flatten_variadic_functions(plan);

    // constant fold stages
    let plan = ir::constant_folding::fold_constants(plan, schema_checking_mode);

    // get the schema_env for the plan
    let schema_env = plan
        .schema(&algebrizer.schema_inference_state())?
        .schema_env;

    let translator = MqlTranslator::new();
    let agg_plan = translator.translate_stage(plan.clone());

    // generate mql from the ir plan
    let mql_translation = match agg_plan {
        Err(translator::Error::UnimplementedStruct) => codegen::generate_mql_from_ir(plan)?,
        Err(err) => return Err(result::Error::Translator(err)),
        Ok(agg_plan) => codegen::generate_mql_from_agg_ir(agg_plan)?.into(),
    };

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
                None => Err(result::Error::CodegenIR(
                    codegen::ir_to_mql::Error::ReferenceNotFound(k),
                )),
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

#[cfg(test)]
mod test_get_namespaces {
    macro_rules! test_get_namespaces {
        ($func_name:ident, $(expected = $expected:expr,)? $(expected_pat = $expected_pat:pat,)? db = $current_db:expr, query = $sql:expr,) => {
            #[test]
            fn $func_name() {
                #[allow(unused_imports)]
                use crate::{get_namespaces, set, Namespace};
                let current_db = $current_db;
                let sql = $sql;
                let actual = get_namespaces(current_db, sql);
                $(assert!(matches!(actual, $expected_pat));)?
                $(assert_eq!(actual, $expected);)?
            }
        };
    }

    test_get_namespaces!(
        no_collections,
        expected = Ok(set![]),
        db = "mydb",
        query = "select * from [] as arr",
    );

    test_get_namespaces!(
        implicit,
        expected = Ok(set![Namespace {
            database: "mydb".into(),
            collection: "foo".into()
        }]),
        db = "mydb",
        query = "select * from foo",
    );

    test_get_namespaces!(
        explicit,
        expected = Ok(set![Namespace {
            database: "bar".into(),
            collection: "baz".into()
        }]),
        db = "mydb",
        query = "select * from bar.baz",
    );

    test_get_namespaces!(
        duplicates,
        expected = Ok(set![Namespace {
            database: "mydb".into(),
            collection: "foo".into()
        }]),
        db = "mydb",
        query = "select * from foo a join foo b",
    );

    test_get_namespaces!(
        semantically_invalid,
        expected = Ok(set![
            Namespace {
                database: "mydb".into(),
                collection: "foo".into()
            },
            Namespace {
                database: "mydb".into(),
                collection: "bar".into()
            }
        ]),
        db = "mydb",
        query = "select a from foo join bar",
    );

    test_get_namespaces!(
        syntactically_invalid,
        expected_pat = Err(_),
        db = "mydb",
        query = "not a valid query",
    );
}
