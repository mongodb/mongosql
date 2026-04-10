use std::collections::HashMap;

use bson::Document;

use crate::{
    Result,
    data_service::{CollectionInfo, CollectionOptions, DataService, TimeSeriesOptions},
};

/// A configurable in-memory implementation of [`DataService`] for use in tests.
///
/// `databases` and `collections` drive `list_databases` and `list_collections`.
/// `documents` drives both `aggregate` and `find`, keyed by `"db.collection"`.
#[allow(dead_code)]
pub(crate) struct MockDataService {
    pub databases: Vec<String>,
    pub collections: HashMap<String, Vec<CollectionInfo>>,
    pub documents: HashMap<String, Vec<Document>>,
}

#[cfg_attr(not(target_arch = "wasm32"), async_trait::async_trait)]
#[cfg_attr(target_arch = "wasm32", async_trait::async_trait(?Send))]
impl DataService for MockDataService {
    async fn list_databases(&self) -> Result<Vec<String>> {
        Ok(self.databases.clone())
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>> {
        Ok(self.collections.get(db_name).cloned().unwrap_or_default())
    }

    async fn aggregate(
        &self,
        db_name: &str,
        coll_name: &str,
        _pipeline: Vec<Document>,
    ) -> Result<Vec<Document>> {
        Ok(self
            .documents
            .get(&format!("{db_name}.{coll_name}"))
            .cloned()
            .unwrap_or_default())
    }

    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        _filter: Document,
    ) -> Result<Vec<Document>> {
        Ok(self
            .documents
            .get(&format!("{db_name}.{coll_name}"))
            .cloned()
            .unwrap_or_default())
    }
}

// CollectionInfo deserialization tests.
//
// These verify that the serde annotations on CollectionInfo, CollectionOptions, and
// TimeSeriesOptions correctly map MongoDB's listCollections wire format to our types.
// Unknown fields returned by the server (e.g. idIndex, info) must be silently ignored.

#[test]
fn test_plain_collection_deserialization() {
    let doc = bson::doc! {
        "name": "myCollection",
        "type": "collection",
        "options": {},
        // Server fields not represented in our type — must be silently ignored.
        "idIndex": { "v": 2, "key": { "_id": 1 }, "name": "_id_" },
        "info": { "readOnly": false }
    };
    let info: CollectionInfo = bson::from_document(doc).unwrap();
    assert_eq!(
        info,
        CollectionInfo {
            name: "myCollection".to_string(),
            collection_type: "collection".to_string(),
            options: CollectionOptions::default(),
        }
    );
}

#[test]
fn test_missing_options_field_uses_default() {
    let doc = bson::doc! {
        "name": "myCollection",
        "type": "collection"
    };
    let info: CollectionInfo = bson::from_document(doc).unwrap();
    assert_eq!(info.options, CollectionOptions::default());
}

#[test]
fn test_view_deserialization() {
    let doc = bson::doc! {
        "name": "myView",
        "type": "view",
        "options": {
            "viewOn": "sourceCollection",
            "pipeline": [{ "$match": { "active": true } }]
        }
    };
    let info: CollectionInfo = bson::from_document(doc).unwrap();
    assert_eq!(info.name, "myView");
    assert_eq!(info.collection_type, "view");
    assert_eq!(info.options.view_on, "sourceCollection");
    assert_eq!(
        info.options.pipeline,
        vec![bson::doc! { "$match": { "active": true } }]
    );
    assert_eq!(info.options.timeseries, None);
}

#[test]
fn test_timeseries_deserialization() {
    let doc = bson::doc! {
        "name": "myTimeseries",
        "type": "timeseries",
        "options": {
            "timeseries": {
                "timeField": "timestamp",
                "metaField": "metadata",
                // Extra server fields not in our struct — must be silently ignored.
                "granularity": "seconds",
                "bucketMaxSpanSeconds": 3600_i32
            }
        }
    };
    let info: CollectionInfo = bson::from_document(doc).unwrap();
    assert_eq!(info.name, "myTimeseries");
    assert_eq!(info.collection_type, "timeseries");
    assert_eq!(
        info.options.timeseries,
        Some(TimeSeriesOptions {
            time_field: "timestamp".to_string(),
            meta_field: Some("metadata".to_string()),
        })
    );
}

#[test]
fn test_timeseries_without_meta_field_deserialization() {
    let doc = bson::doc! {
        "name": "myTimeseries",
        "type": "timeseries",
        "options": {
            "timeseries": {
                "timeField": "timestamp"
            }
        }
    };
    let info: CollectionInfo = bson::from_document(doc).unwrap();
    let ts_opts = info.options.timeseries.unwrap();
    assert_eq!(ts_opts.time_field, "timestamp");
    assert_eq!(ts_opts.meta_field, None);
}
