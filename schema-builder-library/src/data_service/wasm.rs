use bson::Document;
use futures::Stream;
use serde::Serialize;
use wasm_bindgen::prelude::*;

use crate::data_service::AggregateOptions;

use super::{CollectionInfo, DataService};

/// Error type for [`WasmDataService`] operations.
#[derive(Debug, thiserror::Error)]
pub enum WasmDataServiceError {
    #[error("Serialization error: {0}")]
    Serialization(String),
    #[error("Query error: {0}")]
    Query(String),
    #[error("Deserialization error: {0}")]
    Deserialization(String),
}

#[wasm_bindgen(start)]
pub fn init() {
    console_error_panic_hook::set_once();
}

// Embed TypeScript interface definitions directly in the generated .d.ts file.
#[wasm_bindgen(typescript_custom_section)]
const TS_INTERFACE: &str = include_str!("types.d.ts");

// Declare the expected JavaScript object shape.
#[wasm_bindgen]
extern "C" {
    /// JavaScript SqlDataService object type.
    #[wasm_bindgen(typescript_type = "SqlDataService")]
    pub type JsDataService;

    /// JavaScript SqlCursor object type.
    #[wasm_bindgen(typescript_type = "SqlCursor")]
    pub type JsCursor;

    #[wasm_bindgen(method, js_name = "listDatabases", catch)]
    async fn list_database_names(this: &JsDataService) -> Result<JsValue, JsValue>;

    #[wasm_bindgen(method, js_name = "listCollections", catch)]
    async fn list_collections(this: &JsDataService, db_name: &str) -> Result<JsValue, JsValue>;

    #[wasm_bindgen(method, catch)]
    async fn aggregate(
        this: &JsDataService,
        db_name: &str,
        coll_name: &str,
        pipeline: JsValue,
        options: JsValue,
    ) -> Result<JsCursor, JsValue>;

    #[wasm_bindgen(method, catch)]
    async fn find(
        this: &JsDataService,
        db_name: &str,
        coll_name: &str,
        filter: JsValue,
    ) -> Result<JsCursor, JsValue>;

    #[wasm_bindgen(method, catch)]
    async fn next(this: &JsCursor) -> Result<JsValue, JsValue>;
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
    type Error = WasmDataServiceError;

    async fn list_database_names(&self) -> Result<Vec<String>, Self::Error> {
        let js_result = self
            .js_service
            .list_database_names()
            .await
            .map_err(|e| WasmDataServiceError::Query(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| WasmDataServiceError::Deserialization(e.to_string()))
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>, Self::Error> {
        let js_result = self
            .js_service
            .list_collections(db_name)
            .await
            .map_err(|e| WasmDataServiceError::Query(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| WasmDataServiceError::Deserialization(e.to_string()))
    }

    async fn aggregate(
        &self,
        db_name: &str,
        coll_name: &str,
        pipeline: Vec<Document>,
        options: AggregateOptions,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        let serializer = serde_wasm_bindgen::Serializer::new()
            .serialize_maps_as_objects(true)
            .serialize_missing_as_null(true);
        let pipeline_js = pipeline
            .serialize(&serializer)
            .map_err(|e| WasmDataServiceError::Serialization(e.to_string()))?;
        let options_js = options
            .serialize(&serializer)
            .map_err(|e| WasmDataServiceError::Serialization(e.to_string()))?;

        let js_cursor = self
            .js_service
            .aggregate(db_name, coll_name, pipeline_js, options_js)
            .await
            .map_err(|e| WasmDataServiceError::Query(format!("{e:?}")))?;

        Ok(futures::stream::try_unfold(
            js_cursor,
            |cursor| async move {
                let next = cursor
                    .next()
                    .await
                    .map_err(|e| WasmDataServiceError::Query(format!("{e:?}")))?;
                let deserialized: Option<Document> = serde_wasm_bindgen::from_value(next)
                    .map_err(|e| WasmDataServiceError::Deserialization(e.to_string()))?;

                Ok(deserialized.map(|doc| (doc, cursor)))
            },
        ))
    }

    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        filter: Document,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        let serializer = serde_wasm_bindgen::Serializer::new().serialize_maps_as_objects(true);
        let filter_js = filter
            .serialize(&serializer)
            .map_err(|e| WasmDataServiceError::Serialization(e.to_string()))?;

        let js_cursor = self
            .js_service
            .find(db_name, coll_name, filter_js)
            .await
            .map_err(|e| WasmDataServiceError::Query(format!("{e:?}")))?;

        Ok(futures::stream::try_unfold(
            js_cursor,
            |cursor| async move {
                let next = cursor
                    .next()
                    .await
                    .map_err(|e| WasmDataServiceError::Query(format!("{e:?}")))?;
                let deserialized: Option<Document> = serde_wasm_bindgen::from_value(next)
                    .map_err(|e| WasmDataServiceError::Deserialization(e.to_string()))?;

                Ok(deserialized.map(|doc| (doc, cursor)))
            },
        ))
    }
}
