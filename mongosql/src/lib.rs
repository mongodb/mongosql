mod algebrizer;
mod ast;
mod codegen;
mod ir;
mod json_schema;
mod parser;
mod result;
mod schema;
mod util;
use crate::{algebrizer::Algebrizer, parser::Parser, result::Result};

/// Contains all the information needed to execute the MQL translation of a SQL query.
#[derive(Debug)]
pub struct Translation {
    pub target_db: Option<String>,
    pub target_collection: Option<String>,
    pub pipeline: bson::Bson,
}

impl From<codegen::MqlTranslation> for Translation {
    fn from(t: codegen::MqlTranslation) -> Self {
        let pipeline =
            bson::Bson::Array(t.pipeline.into_iter().map(bson::Bson::Document).collect());
        Self {
            target_db: t.database,
            target_collection: t.collection,
            pipeline,
        }
    }
}

/// Returns the MQL translation for the provided SQL query in the
/// specified db. Currently a stub implementation that returns a
/// hard-coded result.
pub fn translate_sql(current_db: &str, sql: &str) -> Result<Translation> {
    // parse the query and apply syntactic rewrites
    let p = Parser::new();
    let ast = p.parse_query(sql)?;
    let ast = ast::rewrites::rewrite_query(ast)?;

    // construct the algebrizer and use it to build an ir plan
    let algebrizer = Algebrizer::new(current_db.to_string(), 0u16);
    let plan = algebrizer.algebrize_query(ast)?;

    // constant fold stages
    let plan = ir::constant_folding::fold_constants(plan);

    // generate mql from the ir plan
    let translation = codegen::generate_mql(plan)?;
    Ok(translation.into())
}
