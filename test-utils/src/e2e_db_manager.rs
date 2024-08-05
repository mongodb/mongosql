use mongodb::{
    bson::{doc, Document},
    error::{CommandError, ErrorKind},
    Client,
};
use std::cell::LazyCell;

#[allow(clippy::declare_interior_mutable_const)]
pub const SCHEMA: LazyCell<Document> = LazyCell::new(|| {
    doc! {
        "bsonType": "object",
        "properties": {
            "_id": { "bsonType": "objectId" },
            "a": { "bsonType": "int" },
        },
        "required": ["_id", "a"],
        "additionalProperties": false,
    }
});

#[allow(clippy::declare_interior_mutable_const)]
pub const ALT_DOC: LazyCell<Document> = LazyCell::new(|| {
    doc! { "b": true }
});

#[allow(clippy::declare_interior_mutable_const)]
pub const MERGED_ALT_SCHEMA: LazyCell<Document> = LazyCell::new(|| {
    doc! {
        "bsonType": "object",
        "properties": {
            "_id": { "bsonType": "objectId" },
            "a": { "bsonType": "int" },
            "b": { "bsonType": "bool" },
        },
        "required": ["_id"],
        "additionalProperties": false,
    }
});

const USER_ALREADY_EXISTS_ERROR_CODE: i32 = 51003;

pub struct TestDatabaseManager {
    pub client: Client,
    dbs: Vec<String>,
}

impl TestDatabaseManager {
    pub async fn new(dbs: Vec<String>, collections: Vec<String>, view: Option<String>) -> Self {
        let mdb_uri = format!(
            "mongodb://localhost:{}/",
            std::env::var("MDB_TEST_LOCAL_PORT").unwrap_or("27017".to_string())
        );
        let client = Client::with_uri_str(mdb_uri)
            .await
            .expect("Failed to create client");

        let admin_db = client.database("admin");

        let create_user_res = admin_db
            .run_command(doc! {
                "createUser": "test",
                "pwd": "test",
                "roles": [ { "role": "readWrite", "db": "test" } ]
            })
            .await;

        match create_user_res {
            Ok(_) => {}
            Err(e) => match *e.kind {
                ErrorKind::Command(CommandError { code, message, .. }) => {
                    // 51003 is the error code for user already exists
                    if code != USER_ALREADY_EXISTS_ERROR_CODE {
                        panic!("Failed to create user with error: {message}");
                    }
                }
                _ => panic!("Failed to create user with error: {e}"),
            },
        }

        for db in &dbs {
            let db = client.database(db);
            // drop the database in case it exists
            db.drop().await.expect("Failed to drop databases");
            for coll in &collections {
                db.collection(coll)
                    .insert_one(doc! { "a": 1 })
                    .await
                    .expect("Failed to insert document");
            }
            if let Some(view_name) = view.clone() {
                db.create_collection(view_name)
                    .view_on(collections[0].clone())
                    .pipeline(vec![doc! { "$match": { "a": 1 } }])
                    .await
                    .expect("Failed to create view");
            }
        }

        Self { client, dbs }
    }

    pub async fn cleanup(&self) {
        for db in &self.dbs {
            let db = self.client.database(db);
            db.drop().await.expect("Failed to drop databases");
        }
        self.client
            .database("admin")
            .run_command(doc! { "dropUser": "test" })
            .await
            .expect("Failed to drop user");
    }
}
