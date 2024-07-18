use mongodb::bson::{doc, Document};

#[tokio::main]
async fn main() {
    // TODO: SQL-2208: Make this more robust, according to the design doc
    let uri = "mongodb://localhost:27017/";
    let client = mongodb::Client::with_uri_str(uri).await.unwrap();

    let my_db = client.database("my_db");
    let my_db_foo = my_db.collection::<Document>("foo");
    let my_db_bar = my_db.collection::<Document>("bar");

    let test_db = client.database("test_db");
    let test_db_foo = test_db.collection::<Document>("foo");
    let test_db_baz = test_db.collection::<Document>("baz");

    let _ = my_db_foo.insert_one(doc! {"_id": 1, "x": 10}).await;
    let _ = my_db_bar.insert_one(doc! {"_id": 1, "y": false}).await;
    let _ = test_db_foo.insert_one(doc! {"_id": 1, "x": "10"}).await;
    let _ = test_db_baz.insert_one(doc! {"_id": 1, "z": 2.5}).await;

    println!("data loaded")
}
