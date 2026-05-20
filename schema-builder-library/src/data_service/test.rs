use std::{collections::HashMap, convert::Infallible};

use bson::Document;
use futures::Stream;

use crate::{
    ViewOptions,
    data_service::{
        AggregateOptions, Collection, CollectionInfo, DataService, TimeSeriesOptions,
        TimeseriesInfo, TimeseriesOptions, ViewInfo,
    },
};

/// A configurable in-memory implementation of [`DataService`] for use in tests.
///
/// `databases` and `collections` drive `list_databases` and `list_collections`.
/// `documents` drives both `aggregate` and `find`, keyed by `"db.collection"`.
#[allow(dead_code)]
pub(crate) struct MockDataService {
    pub databases: Vec<String>,
    pub collections: HashMap<String, Vec<Collection>>,
    pub documents: HashMap<String, Vec<Document>>,
}

impl DataService for MockDataService {
    type Error = Infallible;

    async fn list_database_names(&self) -> Result<Vec<String>, Self::Error> {
        Ok(self.databases.clone())
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<Collection>, Self::Error> {
        Ok(self.collections.get(db_name).cloned().unwrap_or_default())
    }

    async fn aggregate(
        &self,
        db_name: &str,
        coll_name: &str,
        _pipeline: Vec<Document>,
        _options: AggregateOptions,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        let results = self
            .documents
            .get(&format!("{db_name}.{coll_name}"))
            .cloned()
            .unwrap_or_default();

        // The cursor returns results, so we map the infallible documents to a
        // wrapped `Result` here.
        let results: Vec<_> = results.into_iter().map(Ok).collect();

        Ok(futures::stream::iter(results))
    }

    async fn find(
        &self,
        db_name: &str,
        coll_name: &str,
        _filter: Document,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        let results = self
            .documents
            .get(&format!("{db_name}.{coll_name}"))
            .cloned()
            .unwrap_or_default();

        // The cursor returns results, so we map the infallible documents to a
        // wrapped `Result` here.
        let results: Vec<_> = results.into_iter().map(Ok).collect();

        Ok(futures::stream::iter(results))
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
    let collection: Collection = bson::from_document(doc).unwrap();
    assert_eq!(
        collection,
        Collection::Collection(CollectionInfo {
            name: "myCollection".to_string(),
        })
    );
}

#[test]
fn test_view_missing_view_on_defaults_to_empty_string() {
    let doc = bson::doc! {
        "name": "myView",
        "type": "view",
        "options": {}
    };
    let collection: Collection = bson::from_document(doc).unwrap();
    assert_eq!(
        collection,
        Collection::View(ViewInfo {
            name: "myView".to_string(),
            options: ViewOptions {
                view_on: "".to_string(),
                ..Default::default()
            }
        })
    );
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
    let collection: Collection = bson::from_document(doc).unwrap();
    assert_eq!(
        collection,
        Collection::View(ViewInfo {
            name: "myView".to_string(),
            options: ViewOptions {
                view_on: "sourceCollection".to_string(),
                pipeline: vec![bson::doc! { "$match": { "active": true } }],
            },
        })
    );
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
    let collection: Collection = bson::from_document(doc).unwrap();
    assert_eq!(
        collection,
        Collection::Timeseries(TimeseriesInfo {
            name: "myTimeseries".to_string(),
            options: TimeseriesOptions {
                timeseries: TimeSeriesOptions {
                    time_field: "timestamp".to_string(),
                    meta_field: Some("metadata".to_string()),
                },
            },
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
    let collection: Collection = bson::from_document(doc).unwrap();
    assert_eq!(
        collection,
        Collection::Timeseries(TimeseriesInfo {
            name: "myTimeseries".to_string(),
            options: TimeseriesOptions {
                timeseries: TimeSeriesOptions {
                    time_field: "timestamp".to_string(),
                    meta_field: None,
                }
            },
        })
    );
}
