mod parser;
mod result;

pub fn translate_sql(sql: &str) -> String {
    let bson = translate_sql_to_bson(sql);
    serde_json::to_string(&bson).unwrap()
}

fn translate_sql_to_bson(sql: &str) -> bson::Bson {
    parser::parse(sql).expect("provided sql query should parse");
    bson::bson!([
        {"$sql": {
            "statement": sql,
            "format": "odbc",
            "formatVersion": 1,
        }},
    ])
}
