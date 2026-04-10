use bson::Document;
use serde::Serialize;
use wasm_bindgen::prelude::*;

use super::{CollectionInfo, DataService};
use crate::{Error, Result};

#[wasm_bindgen(start)]
pub fn init() {
    console_error_panic_hook::set_once();
}

// Embed TypeScript interface definitions directly in the generated .d.ts file.
#[wasm_bindgen(typescript_custom_section)]
const TS_INTERFACE: &str = r#"
/**
 * BSON Document type - a generic key-value object.
 * This is equivalent to MongoDB's Document type.
 */
export type BsonDocument = Record<string, unknown>;

/**
 * Information about a collection.
 */
export interface CollectionInfo {
    /** The name of the collection */
    name: string;
    /** The type: "collection", "view", or "timeseries" */
    type: string;
    options?: {
        /** For views, the source collection name */
        viewOn?: string;
        /** For views, the aggregation pipeline */
        pipeline?: BsonDocument[];
        /** For timeseries, the timeseries options */
        timeseries?: {
            timeField: string;
            metaField?: string;
        };
    };
}

/**
 * SqlDataService interface for database operations.
 *
 * Implement this interface in your JavaScript/TypeScript code to provide
 * database access to the schema builder.
 */
export interface SqlDataService {
    /** List all database names */
    listDatabases(): Promise<string[]>;
    /** List all collections in a database */
    listCollections(dbName: string): Promise<CollectionInfo[]>;
    /** Execute an aggregation pipeline on a collection */
    aggregate(dbName: string, collName: string, pipeline: BsonDocument[]): Promise<BsonDocument[]>;
    /** Execute a find query on a collection */
    find(dbName: string, collName: string, filter: BsonDocument): Promise<BsonDocument[]>;
}
"#;

// Declare the expected JavaScript object shape.
#[wasm_bindgen]
extern "C" {
    /// JavaScript SqlDataService object type.
    #[wasm_bindgen(typescript_type = "SqlDataService")]
    pub type JsDataService;

    #[wasm_bindgen(method, js_name = "listDatabases", catch)]
    async fn list_databases(this: &JsDataService) -> std::result::Result<JsValue, JsValue>;

    #[wasm_bindgen(method, js_name = "listCollections", catch)]
    async fn list_collections(
        this: &JsDataService,
        db_name: &str,
    ) -> std::result::Result<JsValue, JsValue>;

    #[wasm_bindgen(method, catch)]
    async fn aggregate(
        this: &JsDataService,
        db_name: &str,
        coll_name: &str,
        pipeline: JsValue,
    ) -> std::result::Result<JsValue, JsValue>;

    #[wasm_bindgen(method, catch)]
    async fn find(
        this: &JsDataService,
        db_name: &str,
        coll_name: &str,
        filter: JsValue,
    ) -> std::result::Result<JsValue, JsValue>;
}

/// [`DataService`] implementation backed by a JavaScript DataService object.
pub struct WasmDataService {
    js_service: JsDataService,
}

impl WasmDataService {
    pub fn new(js_service: JsDataService) -> Self {
        Self { js_service }
    }
}

#[async_trait::async_trait(?Send)]
impl DataService for WasmDataService {
    async fn list_databases(&self) -> Result<Vec<String>> {
        let js_result = self
            .js_service
            .list_databases()
            .await
            .map_err(|e| Error::JsError(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| Error::JsError(format!("Deserialization error: {e}")))
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>> {
        let js_result = self
            .js_service
            .list_collections(db_name)
            .await
            .map_err(|e| Error::JsError(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| Error::JsError(format!("Deserialization error: {e}")))
    }

    async fn aggregate(
        &self,
        db_name: &str,
        coll_name: &str,
        pipeline: Vec<Document>,
    ) -> Result<Vec<Document>> {
        let serializer = serde_wasm_bindgen::Serializer::new().serialize_maps_as_objects(true);
        let pipeline_js = pipeline
            .serialize(&serializer)
            .map_err(|e| Error::JsError(format!("Serialization error: {e}")))?;

        let js_result = self
            .js_service
            .aggregate(db_name, coll_name, pipeline_js)
            .await
            .map_err(|e| Error::JsError(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| Error::JsError(format!("Deserialization error: {e}")))
    }

    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        filter: Document,
    ) -> Result<Vec<Document>> {
        let serializer = serde_wasm_bindgen::Serializer::new().serialize_maps_as_objects(true);
        let filter_js = filter
            .serialize(&serializer)
            .map_err(|e| Error::JsError(format!("Serialization error: {e}")))?;

        let js_result = self
            .js_service
            .find(db_name, coll_name, filter_js)
            .await
            .map_err(|e| Error::JsError(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| Error::JsError(format!("Deserialization error: {e}")))
    }
}
