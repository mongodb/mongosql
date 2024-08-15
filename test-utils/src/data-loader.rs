use fake::{Fake, Faker};
use mongodb::{
    bson::{datetime, doc, Bson, Document},
    results::InsertManyResult,
    Collection, Database,
};
use mongosql::{json_schema, schema::Schema};
use rand::{rngs::StdRng, SeedableRng};
use std::{env, str::FromStr};

mod schema_builder_library_integration_test_consts;
use schema_builder_library_integration_test_consts::{
    DATA_DOC_SIZE_IN_BYTES, LARGE_COLL_NAME, LARGE_COLL_SIZE_IN_MB, LARGE_ID_MIN,
    NONUNIFORM_DB_NAME, NONUNIFORM_LARGE_SCHEMA, NONUNIFORM_SMALL_SCHEMA, NONUNIFORM_VIEW_SCHEMA,
    NUM_DOCS_PER_LARGE_PARTITION, SMALL_COLL_NAME, SMALL_COLL_SIZE_IN_MB, SMALL_ID_MIN,
    UNIFORM_COLL_SCHEMA, UNIFORM_DB_NAME, UNIFORM_VIEW_SCHEMA, UNITARY_COLL_NAME, VIEW_NAME,
};

const SEED: [u8; 32] = [
    1, 0, 0, 0, 23, 0, 0, 0, 200, 1, 0, 0, 210, 30, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0,
];
const ASCII: &str = r#"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 !@#$%^&*()_+-=,.<>;':/?[]{}\""#;

macro_rules! handle_write_result {
    ($db: expr, $coll: expr, $res: expr) => {
        match $res {
            Err(err) => println!("failed to write data for `{}.{}`: {err}", $db, $coll),
            Ok(res) => println!(
                "data loaded for `{}.{}`: {} docs inserted",
                $db,
                $coll,
                res.inserted_ids.iter().len()
            ),
        }
    };
}

/// The data-loader tool is used to load data for integration testing the schema-builder-library.
/// This tool loads arbitrary data of preset nature. It creates 2 databases, each with 2 collections
/// and one view. One database contains "uniform" data while the other contains "nonuniform" data.
///
/// Here, "uniform" means the schema is consistent across all documents in the collection, whereas
/// "nonuniform" means the schema is not consistent across all documents, some may differ.
///
/// In each database, one of the collections is 90 MB and the other is 350 MB. This means the former
/// is represented by 1 partition and the latter is represented by 4 partitions.
///
/// All data is arbitrary except for the _id fields, which need to be known for testing purposes.
/// The _id values in each collection are integers, starting at 12345 for the small collections and
/// 0 for the large collections, and increasing by 1 for each document. Additionally, all documents
/// are the same size, 400 bytes. This ensures known partition boundaries.
#[tokio::main]
async fn main() {
    let mdb_uri = format!(
        "mongodb://localhost:{}/",
        env::var("MDB_TEST_LOCAL_PORT").unwrap_or("27017".to_string())
    );
    let client = mongodb::Client::with_uri_str(mdb_uri).await.unwrap();

    let rng = &mut StdRng::from_seed(SEED);

    create_test_user(&client).await;

    // Generate uniform data.
    let uniform_db = client.database(UNIFORM_DB_NAME);
    generate_db_data(&uniform_db, rng, generate_uniform_data_doc).await;

    // Create unit data.
    let unit_coll = uniform_db.collection::<Document>(UNITARY_COLL_NAME);
    let unit_data_res = unit_coll
        .insert_many(vec![generate_uniform_data_doc(0, rng)])
        .await;
    handle_write_result!(UNIFORM_DB_NAME, UNITARY_COLL_NAME, unit_data_res);

    // Generate nonuniform data.
    let nonuniform_db = client.database(NONUNIFORM_DB_NAME);
    generate_db_data(&nonuniform_db, rng, generate_nonuniform_data_doc).await;

    // Create uniform view.
    create_view(
        &uniform_db,
        SMALL_COLL_NAME,
        vec![
            doc! { "$project": { "_id": "$_id", "array_field": "$array_field" } },
            doc! { "$unwind": { "path": "$array_field", "includeArrayIndex": "idx" } },
        ],
    )
    .await;

    // Create nonuniform view.
    create_view(
        &nonuniform_db,
        LARGE_COLL_NAME,
        vec![
            doc! { "$match": { "$expr": { "$gt": ["$var", Bson::Null] } } },
            doc! { "$project": {
                "_id": "$_id",
                "var": "$var",
                "second": { "$convert": { "input": "$second", "to": "long", "onNull": Bson::Int64(-1) } },
                "third": { "$ifNull": ["$third", "unspecified"] },
                "padding_len": { "$strLenBytes": "$padding" }
            }},
        ]
    ).await;

    // Write schemas.
    write_schema(
        &uniform_db,
        SMALL_COLL_NAME,
        "Collection",
        UNIFORM_COLL_SCHEMA.clone(),
    )
    .await;

    write_schema(
        &uniform_db,
        LARGE_COLL_NAME,
        "Collection",
        UNIFORM_COLL_SCHEMA.clone(),
    )
    .await;

    write_schema(
        &uniform_db,
        UNITARY_COLL_NAME,
        "Collection",
        UNIFORM_COLL_SCHEMA.clone(),
    )
    .await;

    write_schema(&uniform_db, VIEW_NAME, "View", UNIFORM_VIEW_SCHEMA.clone()).await;

    write_schema(
        &nonuniform_db,
        SMALL_COLL_NAME,
        "Collection",
        NONUNIFORM_SMALL_SCHEMA.clone(),
    )
    .await;

    write_schema(
        &nonuniform_db,
        LARGE_COLL_NAME,
        "Collection",
        NONUNIFORM_LARGE_SCHEMA.clone(),
    )
    .await;

    write_schema(
        &nonuniform_db,
        VIEW_NAME,
        "View",
        NONUNIFORM_VIEW_SCHEMA.clone(),
    )
    .await;
}

/// create_test_user creates a test user in the admin database
async fn create_test_user(client: &mongodb::Client) {
    let admin_db = client.database("admin");
    // If the user already exists, MongoDB will return a duplicate user error
    // this is unimportant for the purposes of this tool.
    let _ = admin_db
        .run_command(
            doc! { "createUser": "test", "pwd": "test", "roles": ["readWriteAnyDatabase"] },
        )
        .await;
}

/// generate_db_data generates the large and small collection data in the provided database, using
/// the provided generator.
async fn generate_db_data(
    db: &Database,
    rng: &mut StdRng,
    generator: fn(i64, &mut StdRng) -> Document,
) {
    // Generate small data.
    let small_coll = db.collection::<Document>(SMALL_COLL_NAME);
    let small_data_res = generate_collection_data(
        SMALL_COLL_SIZE_IN_MB,
        SMALL_ID_MIN,
        small_coll,
        rng,
        generator,
    )
    .await;
    handle_write_result!(db.name(), SMALL_COLL_NAME, small_data_res);

    // Generate large data.
    let large_coll = db.collection::<Document>(LARGE_COLL_NAME);
    let large_data_res = generate_collection_data(
        LARGE_COLL_SIZE_IN_MB,
        LARGE_ID_MIN,
        large_coll,
        rng,
        generator,
    )
    .await;
    handle_write_result!(db.name(), LARGE_COLL_NAME, large_data_res);
}

/// generate_collection_data generates data of the provided size using the provided generator
/// function and inserts that data into the provided collection. It creates as many documents as
/// needed to reach the specified data size.
async fn generate_collection_data(
    size_in_mb: i64,
    id_min: i64,
    coll: Collection<Document>,
    rng: &mut StdRng,
    generator: fn(i64, &mut StdRng) -> Document,
) -> mongodb::error::Result<InsertManyResult> {
    let size_in_bytes = size_in_mb * 1024 * 1024;
    let num_docs = (size_in_bytes / DATA_DOC_SIZE_IN_BYTES) + id_min;

    let range = id_min..num_docs;
    let data = range.map(move |idx| generator(idx, rng));

    coll.insert_many(data).await
}

/// generate_uniform_data_doc generates an arbitrary data document for use in the uniform dataset.
/// The schema for this document is evident in the body of the function. Each document is 400 bytes.
fn generate_uniform_data_doc(idx: i64, rng: &mut StdRng) -> Document {
    let string_faker = fake::StringFaker::with(Vec::from(ASCII), 92);

    let oid = mongodb::bson::oid::ObjectId::new();
    let uuid = mongodb::bson::uuid::Uuid::new();
    let double: f64 = Faker.fake_with_rng(rng);
    let long: i64 = Faker.fake_with_rng(rng);
    let long = Bson::Int64(long);
    let string: String = string_faker.fake_with_rng(rng);
    let date = mongodb::bson::datetime::DateTime::now();
    let bool: bool = Faker.fake_with_rng(rng);
    let decimal =
        mongodb::bson::decimal128::Decimal128::from_str(format!("{idx}").as_str()).unwrap();
    let int: i32 = Faker.fake_with_rng(rng);
    let arr_int_1: i32 = Faker.fake_with_rng(rng);
    let arr_int_2: i32 = Faker.fake_with_rng(rng);

    let doc = doc! {
        "_id": Bson::Int64(idx),
        "oid_field": oid,
        "uuid_field": uuid,
        "double_field": double,
        "long_field": long,
        "string_field": string,
        "date_field": date,
        "document_field": {
            "sub_bool_field": bool,
            "sub_decimal_field": decimal,
            "sub_document_field": {
                "sub_sub_int_field": int,
            },
        },
        "array_field": vec![arr_int_1, arr_int_2],
    };

    assert_data_doc_size(doc.clone(), format!("uniform data doc (idx {idx}"));

    doc
}

/// generate_nonuniform_data_doc generates an arbitrary data document for use in the nonuniform
/// dataset. The intent of these documents is to have inconsistent shape across them, so subsequent
/// calls may not produce similar documents. To facilitate testing, this function produces documents
/// of the same shape for certain ranges. Specifically, it produces different document shapes along
/// partition boundaries:
///     1. idx values in the first partition range look like:
///         { _id: <long>, var: <int | string | null>, padding: <string> }
///     2. idx values in the second partition range look like:
///         { _id: <long>, var: <int | string | null>, second: <int | long>, padding: <string> }
///     3. idx values in the third partition range look like:
///         { _id: <long>, var: <int | string | null>, third: oid, padding: <string> }
///     4. idx values in the fourth partition range look like:
///         { _id: <long>, var: <int | string | null>, second: <int | long>,
///           third: oid, padding: <string> }
///
/// Each document is 400 bytes -- the "padding" field ensures we make it to that size.
fn generate_nonuniform_data_doc(idx: i64, rng: &mut StdRng) -> Document {
    let small_string_faker = fake::StringFaker::with(Vec::from(ASCII), 4);
    let (var_val, var_val_size) = match idx % 3 {
        0 => (Bson::Int32(Faker.fake_with_rng(rng)), 4),
        1 => (Bson::String(small_string_faker.fake_with_rng(rng)), 9),
        2 => (Bson::Null, 0),
        _ => panic!("invalid result from idx % 3"),
    };

    let (second_val, second_val_size) = match idx % 2 {
        0 => (Bson::Int32(Faker.fake_with_rng(rng)), 4),
        1 => (Bson::Int64(Faker.fake_with_rng(rng)), 8),
        _ => panic!("invalid result from idx % 2"),
    };

    let (third_val, third_val_size) = (mongodb::bson::oid::ObjectId::new(), 12);

    let mut doc = doc! { "_id": Bson::Int64(idx), "var": var_val };

    let partition_desc = if idx < *NUM_DOCS_PER_LARGE_PARTITION {
        // First partition
        let padding_size = DATA_DOC_SIZE_IN_BYTES
            - 8 /* size of _id value */
            - var_val_size
            - 16 /* size of field names */
            - 5 /* size of document infrastructure */
            - 5 /* size of string infrastructure */
            - 3 /* # elements */;
        let padding_faker = fake::StringFaker::with(Vec::from(ASCII), padding_size as usize);
        doc.insert("padding", Bson::String(padding_faker.fake_with_rng(rng)));

        "partition 1"
    } else if idx < *NUM_DOCS_PER_LARGE_PARTITION * 2 {
        // Second partition
        doc.insert("second", second_val);

        let padding_size = DATA_DOC_SIZE_IN_BYTES
            - 8 /* size of _id value */
            - var_val_size
            - second_val_size
            - 23 /* size of field names */
            - 5 /* size of document infrastructure */
            - 5 /* size of string infrastructure */
            - 4 /* # elements */;
        let padding_faker = fake::StringFaker::with(Vec::from(ASCII), padding_size as usize);
        doc.insert("padding", Bson::String(padding_faker.fake_with_rng(rng)));

        "partition 2"
    } else if idx < *NUM_DOCS_PER_LARGE_PARTITION * 3 {
        // Third partition
        doc.insert("third", third_val);

        let padding_size = DATA_DOC_SIZE_IN_BYTES
            - 8 /* size of _id value */
            - var_val_size
            - third_val_size
            - 22 /* size of field names */
            - 5 /* size of document infrastructure */
            - 5 /* size of string infrastructure */
            - 4 /* # elements */;
        let padding_faker = fake::StringFaker::with(Vec::from(ASCII), padding_size as usize);
        doc.insert("padding", Bson::String(padding_faker.fake_with_rng(rng)));

        "partition 3"
    } else {
        // Fourth partition
        doc.insert("second", second_val);
        doc.insert("third", third_val);

        let padding_size = DATA_DOC_SIZE_IN_BYTES
            - 8 /* size of _id value */
            - var_val_size
            - second_val_size
            - third_val_size
            - 29 /* size of field names */
            - 5 /* size of document infrastructure */
            - 5 /* size of string infrastructure */
            - 5 /* # elements */;
        let padding_faker = fake::StringFaker::with(Vec::from(ASCII), padding_size as usize);
        doc.insert("padding", Bson::String(padding_faker.fake_with_rng(rng)));

        "partition 4"
    };

    let desc = format!("nonuniform data doc (idx {}) ({})", idx, partition_desc);
    assert_data_doc_size(doc.clone(), desc);

    doc
}

fn assert_data_doc_size(doc: Document, desc: String) {
    let mut bytes: Vec<u8> = vec![];
    let _ = doc.to_writer(&mut bytes);
    assert_eq!(
        DATA_DOC_SIZE_IN_BYTES,
        bytes.len() as i64,
        "{desc} is incorrect size"
    )
}

async fn create_view(db: &Database, coll_name: &str, pipeline: Vec<Document>) {
    let view_res = db
        .create_collection(VIEW_NAME)
        .view_on(coll_name)
        .pipeline(pipeline)
        .await;

    match view_res {
        Err(err) => println!("failed to create view in db `{}`: {}", db.name(), err),
        Ok(()) => println!("view created in db `{}`", db.name()),
    }
}

async fn write_schema(
    db: &Database,
    coll_or_view_name: &str,
    coll_or_view_type: &str,
    schema: Schema,
) {
    let schema_collection = db.collection::<Document>("__sql_schemas");

    let as_json_schema: json_schema::Schema = schema.try_into().unwrap_or_else(|_| {
        panic!(
            "failed to convert Schema to JSON schema for {}.{coll_or_view_name}",
            db.name()
        )
    });
    let as_bson_schema = as_json_schema.to_bson().unwrap_or_else(|_| {
        panic!(
            "failed to convert JSON schema to BSON for {}.{coll_or_view_name}",
            db.name()
        )
    });

    let schema_doc = doc! {
        "_id": coll_or_view_name,
        "type": coll_or_view_type,
        "schema": as_bson_schema,
        "lastUpdated": datetime::DateTime::now(),
    };

    let write_res = schema_collection.insert_one(schema_doc).await;
    match write_res {
        Err(err) => println!(
            "failed to write schema for {}.{coll_or_view_name}: {err}",
            db.name()
        ),
        Ok(res) => println!(
            "successfully wrote schema for {}.{coll_or_view_name}: {res:?}",
            db.name()
        ),
    }
}
