use crate::{codegen::generate_mql, ir};

#[test]
fn codegen_collection() {
    let current_db = "test".to_string();
    let plan = ir::Stage::Collection(ir::Collection {
        db: current_db.clone(),
        collection: "col".to_string(),
    });

    let actual = generate_mql(current_db, plan).expect("codegen failed");
    assert_eq!(actual.database, "test".to_string());
    assert_eq!(actual.collection, Some("col".to_string()));
    assert_eq!(
        actual.bson,
        bson::bson!([{"$project": {"_id": 0, "col": "$$ROOT"}}]),
    );
}
