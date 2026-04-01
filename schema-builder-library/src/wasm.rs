//! WASM bridge for the schema builder library.
//!
//! This module provides WASM bindings for the schema builder, allowing it to be
//! called from JavaScript/TypeScript. It defines the JavaScript interface using
//! `wasm_bindgen` and provides a `WasmDataService` wrapper that implements the
//! Rust `DataService` trait by delegating to JavaScript callbacks.
//!
//! Note: The WASM entry point (`build_schema_wasm`) will be added in a future
//! ticket once `BuilderOptions` is made generic over `DataService`.

#![cfg(target_arch = "wasm32")]

use bson::Document;
use serde::Serialize;
use wasm_bindgen::prelude::*;

use crate::{Error, Result, data_service::{CollectionInfo, DataService}};

/// Initialize the WASM module. Sets up a panic hook for better error messages in the browser.
#[wasm_bindgen(start)]
pub fn init() {
    #[cfg(feature = "console_error_panic_hook")]
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
    /** Options for views */
    options?: {
        /** For views, the source collection name */
        viewOn?: string;
        /** For views, the aggregation pipeline */
        pipeline?: BsonDocument[];
    };
}

/**
 * DataService interface for database operations.
 *
 * Implement this interface in your JavaScript/TypeScript code to provide
 * database access to the schema builder.
 *
 * @example
 * ```typescript
 * class MyDataService implements DataService {
 *     async listDatabases(): Promise<string[]> {
 *         return ['db1', 'db2'];
 *     }
 *     async listCollections(dbName: string): Promise<CollectionInfo[]> {
 *         return [{ name: 'coll1', type: 'collection' }];
 *     }
 *     async aggregate(ns: string, pipeline: BsonDocument[]): Promise<BsonDocument[]> {
 *         return await myDriver.aggregate(ns, pipeline);
 *     }
 *     async find(ns: string, filter: BsonDocument): Promise<BsonDocument[]> {
 *         return await myDriver.find(ns, filter);
 *     }
 * }
 * ```
 */
export interface DataService {
    /** List all database names */
    listDatabases(): Promise<string[]>;
    /** List all collections in a database */
    listCollections(dbName: string): Promise<CollectionInfo[]>;
    /** Execute an aggregation pipeline on a namespace (format: "database.collection") */
    aggregate(ns: string, pipeline: BsonDocument[]): Promise<BsonDocument[]>;
    /** Execute a find query on a namespace (format: "database.collection") */
    find(ns: string, filter: BsonDocument): Promise<BsonDocument[]>;
}
"#;

// Declare the expected JavaScript object shape.
#[wasm_bindgen]
extern "C" {
    /// JavaScript DataService object type.
    #[wasm_bindgen(typescript_type = "DataService")]
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
        ns: &str,
        pipeline: JsValue,
    ) -> std::result::Result<JsValue, JsValue>;

    #[wasm_bindgen(method, catch)]
    async fn find(
        this: &JsDataService,
        ns: &str,
        filter: JsValue,
    ) -> std::result::Result<JsValue, JsValue>;
}

/// Wrapper around a JavaScript DataService that implements the Rust `DataService` trait.
///
/// This type bridges the JS world and the Rust schema builder by deserializing
/// JS return values into Rust types and serializing Rust inputs (pipelines, filters)
/// into plain JS objects.
pub struct WasmDataService {
    js_service: JsDataService,
}

impl WasmDataService {
    /// Create a new `WasmDataService` wrapping a JavaScript DataService object.
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

    async fn aggregate(&self, ns: &str, pipeline: Vec<Document>) -> Result<Vec<Document>> {
        let serializer = serde_wasm_bindgen::Serializer::new().serialize_maps_as_objects(true);
        let pipeline_js = pipeline
            .serialize(&serializer)
            .map_err(|e| Error::JsError(format!("Serialization error: {e}")))?;

        let js_result = self
            .js_service
            .aggregate(ns, pipeline_js)
            .await
            .map_err(|e| Error::JsError(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| Error::JsError(format!("Deserialization error: {e}")))
    }

    async fn find(&self, ns: &str, filter: Document) -> Result<Vec<Document>> {
        let serializer = serde_wasm_bindgen::Serializer::new().serialize_maps_as_objects(true);
        let filter_js = filter
            .serialize(&serializer)
            .map_err(|e| Error::JsError(format!("Serialization error: {e}")))?;

        let js_result = self
            .js_service
            .find(ns, filter_js)
            .await
            .map_err(|e| Error::JsError(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| Error::JsError(format!("Deserialization error: {e}")))
    }
}
