use std::{
    collections::HashMap,
    convert::Infallible,
    sync::{Mutex, PoisonError},
};

use bson::Document;
use futures::Stream;

use crate::data_service::{
    AggregateOptions, CollectionInfo, CollectionOptions, CollectionType, DataService,
    TimeSeriesOptions,
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

impl DataService for MockDataService {
    type Error = Infallible;

    async fn list_database_names(&self) -> Result<Vec<String>, Self::Error> {
        Ok(self.databases.clone())
    }

    async fn list_collections(&self, db_name: &str) -> Result<Vec<CollectionInfo>, Self::Error> {
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

/// The kind of aggregation a [`ScriptedDataService`] receives, classified by its stages.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub(crate) enum PipelineKind {
    /// `$collStats`, issued by `get_size_counts`.
    CollStats,
    /// `$sort` + `$limit`, issued by `get_bounds`.
    Bound { ascending: bool },
    /// `$sort` + `$project` + `$match { $sampleRate }`, the partitioning pipeline.
    Partitioning,
    /// `$sample`, the large-collection sampling pipeline.
    Sample,
    /// `$match` + `$sort` + `$limit`, the range scan in `derive_schema_for_partition`.
    RangeScan,
}

impl PipelineKind {
    fn classify(pipeline: &[Document]) -> Self {
        let first = pipeline.first().and_then(|stage| stage.keys().next());
        match first.map(String::as_str) {
            Some("$collStats") => Self::CollStats,
            Some("$sample") => Self::Sample,
            Some("$match") => Self::RangeScan,
            Some("$sort") if pipeline.len() == 3 && pipeline[1].contains_key("$limit") => {
                let ascending = pipeline[0]
                    .get_document("$sort")
                    .ok()
                    .and_then(|sort| sort.values().next())
                    .and_then(bson::Bson::as_i32)
                    == Some(1);
                Self::Bound { ascending }
            }
            Some("$sort") => Self::Partitioning,
            other => panic!("unexpected pipeline in test: {other:?}"),
        }
    }
}

/// The scripted outcome of one `aggregate` call on a [`ScriptedDataService`].
#[derive(Debug, Clone)]
pub(crate) enum Scripted {
    /// The cursor opens and yields these documents.
    Docs(Vec<Document>),
    /// The `aggregate` call itself fails.
    FailOpen,
    /// The cursor yields these documents, then fails (e.g. `CursorKilled` on `getMore`).
    FailMidStream(Vec<Document>),
}

/// The error type of [`ScriptedDataService`].
#[derive(Debug, thiserror::Error)]
#[error("scripted failure: {0}")]
pub(crate) struct ScriptedError(&'static str);

/// Handler deciding the outcome of an `aggregate` call from its kind and how many earlier calls
/// of that kind there were.
type Script = Box<dyn Fn(PipelineKind, usize) -> Scripted + Send + Sync>;

/// A [`DataService`] whose `aggregate` results are scripted per [`PipelineKind`], and which
/// records every kind it receives so tests can assert which pipelines ran.
pub(crate) struct ScriptedDataService {
    script: Script,
    calls: Mutex<Vec<PipelineKind>>,
}

impl ScriptedDataService {
    pub(crate) fn new(
        script: impl Fn(PipelineKind, usize) -> Scripted + Send + Sync + 'static,
    ) -> Self {
        Self {
            script: Box::new(script),
            calls: Mutex::new(Vec::new()),
        }
    }

    /// The kinds of every `aggregate` call made so far, in order.
    pub(crate) fn calls(&self) -> Vec<PipelineKind> {
        self.calls
            .lock()
            .unwrap_or_else(PoisonError::into_inner)
            .clone()
    }
}

impl DataService for ScriptedDataService {
    type Error = ScriptedError;

    async fn list_database_names(&self) -> Result<Vec<String>, Self::Error> {
        Ok(Vec::new())
    }

    async fn list_collections(&self, _db_name: &str) -> Result<Vec<CollectionInfo>, Self::Error> {
        Ok(Vec::new())
    }

    async fn aggregate(
        &self,
        _db_name: &str,
        _coll_name: &str,
        pipeline: Vec<Document>,
        _options: AggregateOptions,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        let kind = PipelineKind::classify(&pipeline);
        let prior_calls = {
            let mut calls = self.calls.lock().unwrap_or_else(PoisonError::into_inner);
            let prior = calls.iter().filter(|k| **k == kind).count();
            calls.push(kind);
            prior
        };

        let results: Vec<Result<Document, ScriptedError>> = match (self.script)(kind, prior_calls) {
            Scripted::Docs(docs) => docs.into_iter().map(Ok).collect(),
            Scripted::FailOpen => return Err(ScriptedError("aggregate failed to open")),
            Scripted::FailMidStream(docs) => docs
                .into_iter()
                .map(Ok)
                .chain(std::iter::once(Err(ScriptedError("cursor killed"))))
                .collect(),
        };
        Ok(futures::stream::iter(results))
    }

    async fn find(
        &self,
        _db_name: &str,
        _coll_name: &str,
        _filter: Document,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        Ok(futures::stream::iter(Vec::new()))
    }
}

// CollectionType deserialization tests.
//
// These verify that CollectionType correctly maps the raw strings from MongoDB's
// listCollections wire format and rejects unknown values.

#[test]
fn test_collection_type_deserialization() {
    assert_eq!(
        bson::from_bson::<CollectionType>(bson::Bson::String("collection".to_string())).unwrap(),
        CollectionType::Collection
    );
    assert_eq!(
        bson::from_bson::<CollectionType>(bson::Bson::String("view".to_string())).unwrap(),
        CollectionType::View
    );
    assert_eq!(
        bson::from_bson::<CollectionType>(bson::Bson::String("timeseries".to_string())).unwrap(),
        CollectionType::Timeseries
    );
}

#[test]
fn test_unknown_collection_type_deserialization_errors() {
    let result = bson::from_bson::<CollectionType>(bson::Bson::String("unknown".to_string()));
    assert!(result.is_err());
    let err = result.unwrap_err().to_string();
    assert!(
        err.contains("unknown"),
        "error should mention the unknown variant, got: {err}"
    );
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
            collection_type: CollectionType::Collection,
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
fn test_view_missing_view_on_defaults_to_empty_string() {
    let doc = bson::doc! {
        "name": "myView",
        "type": "view",
        "options": {}
    };
    let info: CollectionInfo = bson::from_document(doc).unwrap();
    assert_eq!(info.options.view_on, "");
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
    assert_eq!(
        info,
        CollectionInfo {
            name: "myView".to_string(),
            collection_type: CollectionType::View,
            options: CollectionOptions {
                view_on: "sourceCollection".to_string(),
                pipeline: vec![bson::doc! { "$match": { "active": true } }],
                timeseries: None,
            },
        }
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
    let info: CollectionInfo = bson::from_document(doc).unwrap();
    assert_eq!(
        info,
        CollectionInfo {
            name: "myTimeseries".to_string(),
            collection_type: CollectionType::Timeseries,
            options: CollectionOptions {
                timeseries: Some(TimeSeriesOptions {
                    time_field: "timestamp".to_string(),
                    meta_field: Some("metadata".to_string()),
                }),
                ..Default::default()
            },
        }
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
    assert_eq!(
        info,
        CollectionInfo {
            name: "myTimeseries".to_string(),
            collection_type: CollectionType::Timeseries,
            options: CollectionOptions {
                timeseries: Some(TimeSeriesOptions {
                    time_field: "timestamp".to_string(),
                    meta_field: None,
                }),
                ..Default::default()
            },
        }
    );
}
