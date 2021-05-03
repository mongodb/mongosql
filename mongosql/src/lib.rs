mod ast;
mod ir;
mod json_schema;
mod parser;
mod result;
mod schema;
use crate::parser::Parser;

/// Contains all the information needed to execute the MQL translation of a SQL query.
#[derive(Debug)]
pub struct Translation {
    pub target_db: String,
    pub target_collection: String,
    pub pipeline: bson::Bson,
}

/// Returns the MQL translation for the provided SQL query in the
/// specified db. Currently a stub implementation that returns a
/// hard-coded result.
pub fn translate_sql(current_db: &str, sql: &str) -> Translation {
    let p = Parser::new();
    p.parse_query(sql).expect("provided sql query should parse");
    Translation {
        target_db: current_db.into(),
        target_collection: "foo".into(),
        pipeline: bson::bson!([
            {"$project": {"_id": 0, "a": "$a"}}
        ]),
    }
}

/// A variant of translate_sql that encodes the resulting Translation
/// as BSON and then base64-encodes it.
pub fn translate_sql_bson_base64(current_db: &str, sql: &str) -> String {
    let translation = translate_sql(current_db, sql);
    let translation = bson::doc! {
        "target_db": translation.target_db,
        "target_collection": translation.target_collection,
        "pipeline": translation.pipeline,
    };

    let mut buf = Vec::new();
    translation
        .to_writer(&mut buf)
        .expect("failed to serialize Translation");

    base64::encode(buf)
}
