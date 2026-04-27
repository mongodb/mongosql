use bson::Document;
use serde::Serialize;
use wasm_bindgen::prelude::*;

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
        hint: Option<JsValue>,
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
    type Error = WasmDataServiceError;

    async fn list_databases(&self) -> std::result::Result<Vec<String>, Self::Error> {
        let js_result = self
            .js_service
            .list_databases()
            .await
            .map_err(|e| WasmDataServiceError::Query(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| WasmDataServiceError::Deserialization(e.to_string()))
    }

    async fn list_collections(
        &self,
        db_name: &str,
    ) -> std::result::Result<Vec<CollectionInfo>, Self::Error> {
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
        hint: Option<Document>,
    ) -> std::result::Result<Vec<Document>, Self::Error> {
        let serializer = serde_wasm_bindgen::Serializer::new().serialize_maps_as_objects(true);
        let pipeline_js = pipeline
            .serialize(&serializer)
            .map_err(|e| WasmDataServiceError::Serialization(e.to_string()))?;
        let hint_js = hint
            .map(|h| {
                h.serialize(&serializer)
                    .map_err(|e| WasmDataServiceError::Serialization(e.to_string()))
            })
            .transpose()?;

        let js_result = self
            .js_service
            .aggregate(db_name, coll_name, pipeline_js, hint_js)
            .await
            .map_err(|e| WasmDataServiceError::Query(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| WasmDataServiceError::Deserialization(e.to_string()))
    }

    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        filter: Document,
    ) -> std::result::Result<Vec<Document>, Self::Error> {
        let serializer = serde_wasm_bindgen::Serializer::new().serialize_maps_as_objects(true);
        let filter_js = filter
            .serialize(&serializer)
            .map_err(|e| WasmDataServiceError::Serialization(e.to_string()))?;

        let js_result = self
            .js_service
            .find(db_name, coll_name, filter_js)
            .await
            .map_err(|e| WasmDataServiceError::Query(format!("{e:?}")))?;

        serde_wasm_bindgen::from_value(js_result)
            .map_err(|e| WasmDataServiceError::Deserialization(e.to_string()))
    }
}
