mod ast;
mod codegen;
mod ir;
mod parser;
mod result;
mod schema;
use crate::{parser::Parser, result::Result};

/// Contains all the information needed to execute the MQL translation of a SQL query.
#[derive(Debug)]
pub struct Translation {
    pub target_db: String,
    pub target_collection: Option<String>,
    pub pipeline: bson::Bson,
}

impl From<codegen::MqlTranslation> for Translation {
    fn from(t: codegen::MqlTranslation) -> Self {
        Self {
            target_db: t.database,
            target_collection: t.collection,
            pipeline: t.bson,
        }
    }
}

/// Returns the MQL translation for the provided SQL query in the
/// specified db. Currently a stub implementation that returns a
/// hard-coded result.
pub fn translate_sql(current_db: &str, sql: &str) -> Result<Translation> {
    let p = Parser::new();
    let _ast = p.parse_query(sql)?;
    let _ast = ast::rewrites::rewrite_query(_ast)?;

    let plan = ir::Stage::Collection(ir::Collection {
        db: current_db.to_string(),
        collection: "foo".to_string(),
    });

    let translation = codegen::generate_mql(current_db.to_string(), plan)?;
    Ok(translation.into())
}
