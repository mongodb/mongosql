mod parser;
mod result;

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
    parser::parse(sql).expect("provided sql query should parse");
    Translation {
        target_db: current_db.into(),
        target_collection: "foo".into(),
        pipeline: bson::bson!([
            {"$project": {"_id": 0, "a": "$a"}}
        ]),
    }
}

/// A variant of translate_sql that encodes the resulting Translation
/// as extended JSON. This function will become unnecessary once we
/// switch the mongosql-c API to encode its return values as BSON
/// instead of extended JSON.
pub fn translate_sql_extjson(current_db: &str, sql: &str) -> String {
    let translation = translate_sql(current_db, sql);
    let translation = bson::bson!({
        "target_db": translation.target_db,
        "target_collection": translation.target_collection,
        "pipeline": translation.pipeline,
    });
    serde_json::to_string(&translation.into_canonical_extjson()).unwrap()
}
