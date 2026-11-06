//! A scriptable in-memory [`DataService`] for unit tests.
//!
//! [`ScriptedDataService`] answers each pipeline shape issued by partitioning and schema
//! derivation with a scripted [`Response`], and records every pipeline it receives so tests can
//! assert which queries ran.

use std::sync::{Mutex, PoisonError};

use bson::{Bson, Document, doc};
use futures::Stream;

use crate::data_service::{AggregateOptions, CollectionInfo, DataService};

/// The database name every [`ScriptedDataService`] answers for.
pub(crate) const DB: &str = "db";
/// The collection name every [`ScriptedDataService`] lists.
pub(crate) const COLL: &str = "coll";
/// One mebibyte, for building `$collStats` sizes.
pub(crate) const MIB: i64 = 1024 * 1024;

/// A scripted outcome for one aggregation.
#[derive(Debug, Clone)]
pub(crate) enum Response {
    /// The cursor opens and yields these documents.
    Docs(Vec<Document>),
    /// The `aggregate` call itself fails.
    FailOnOpen,
    /// The cursor opens, yields these documents, then fails (e.g. a killed `getMore`).
    FailMidStream(Vec<Document>),
}

/// The error returned by [`ScriptedDataService`] for scripted failures.
#[derive(Debug, thiserror::Error)]
#[error("scripted failure: {0}")]
pub(crate) struct MockError(&'static str);

/// A [`DataService`] whose aggregations are answered by pipeline shape.
///
/// | Pipeline                                         | Answered by      |
/// |--------------------------------------------------|------------------|
/// | `$collStats`                                     | `coll_stats`     |
/// | `$sort` + `$limit` (bounds)                      | `min_bound` / `max_bound`, by sort direction |
/// | `$sort` + `$project` + `$match.$sampleRate`      | `sample_rate`    |
/// | `$sample`                                        | `sample`         |
/// | `$match` (per-partition scan)                    | always empty     |
#[derive(Debug)]
pub(crate) struct ScriptedDataService {
    pub collections: Vec<CollectionInfo>,
    /// One `$collStats` document per shard.
    pub coll_stats: Vec<Document>,
    pub min_bound: Document,
    pub max_bound: Document,
    pub sample_rate: Response,
    pub sample: Response,
    pipelines: Mutex<Vec<Vec<Document>>>,
}

impl ScriptedDataService {
    /// Creates a service for [`DB`].[`COLL`] whose `$collStats` returns `coll_stats`.
    ///
    /// Bounds default to `_id` 0 and 100, and both sampling queries return no documents.
    pub(crate) fn new(coll_stats: Vec<Document>) -> Self {
        Self {
            collections: vec![CollectionInfo {
                name: COLL.to_string(),
                ..Default::default()
            }],
            coll_stats,
            min_bound: doc! { "_id": 0 },
            max_bound: doc! { "_id": 100 },
            sample_rate: Response::Docs(vec![]),
            sample: Response::Docs(vec![]),
            pipelines: Mutex::new(vec![]),
        }
    }

    /// Returns every pipeline received so far, in order.
    pub(crate) fn pipelines(&self) -> Vec<Vec<Document>> {
        self.pipelines
            .lock()
            .unwrap_or_else(PoisonError::into_inner)
            .clone()
    }

    /// Returns true if a `$sample` pipeline was issued.
    pub(crate) fn issued_sample(&self) -> bool {
        self.pipelines().iter().any(|p| is_sample(p))
    }

    fn respond(&self, pipeline: &[Document]) -> Response {
        let first = pipeline.first();
        let first_has = |key: &str| first.is_some_and(|stage| stage.contains_key(key));

        if first_has("$collStats") {
            Response::Docs(self.coll_stats.clone())
        } else if is_sample(pipeline) {
            self.sample.clone()
        } else if is_sample_rate(pipeline) {
            self.sample_rate.clone()
        } else if first_has("$sort")
            && pipeline
                .get(1)
                .is_some_and(|stage| stage.contains_key("$limit"))
        {
            let ascending = first
                .and_then(|stage| stage.get_document("$sort").ok())
                .and_then(|sort| sort.values().next().cloned())
                == Some(Bson::Int32(1));
            let bound = if ascending {
                &self.min_bound
            } else {
                &self.max_bound
            };
            Response::Docs(vec![bound.clone()])
        } else if first_has("$match") {
            Response::Docs(vec![])
        } else {
            Response::FailOnOpen
        }
    }
}

fn is_sample(pipeline: &[Document]) -> bool {
    pipeline
        .first()
        .is_some_and(|stage| stage.contains_key("$sample"))
}

fn is_sample_rate(pipeline: &[Document]) -> bool {
    pipeline.iter().any(|stage| {
        stage
            .get_document("$match")
            .is_ok_and(|m| m.contains_key("$sampleRate"))
    })
}

/// Builds a per-shard `$collStats` document.
pub(crate) fn stats(size: i64, count: i64) -> Document {
    doc! { "storageStats": { "size": size, "count": count } }
}

/// Builds a per-shard `$collStats` document without `count`, as timeseries collections report.
pub(crate) fn stats_without_count(size: i64) -> Document {
    doc! { "storageStats": { "size": size } }
}

#[expect(
    clippy::unused_async_trait_impl,
    reason = "the trait is async; scripted responses are ready immediately"
)]
impl DataService for ScriptedDataService {
    type Error = MockError;

    async fn list_database_names(&self) -> Result<Vec<String>, Self::Error> {
        Ok(vec![DB.to_string()])
    }

    async fn list_collections(&self, _db_name: &str) -> Result<Vec<CollectionInfo>, Self::Error> {
        Ok(self.collections.clone())
    }

    async fn aggregate(
        &self,
        _db_name: &str,
        _coll_name: &str,
        pipeline: Vec<Document>,
        _options: AggregateOptions,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        let response = self.respond(&pipeline);
        self.pipelines
            .lock()
            .unwrap_or_else(PoisonError::into_inner)
            .push(pipeline);

        let items: Vec<Result<Document, MockError>> = match response {
            Response::FailOnOpen => return Err(MockError("aggregate failed to open")),
            Response::Docs(docs) => docs.into_iter().map(Ok).collect(),
            Response::FailMidStream(docs) => docs
                .into_iter()
                .map(Ok)
                .chain(std::iter::once(Err(MockError("cursor killed mid-stream"))))
                .collect(),
        };
        Ok(futures::stream::iter(items))
    }

    async fn find(
        &self,
        _db_name: &str,
        _coll_name: &str,
        _filter: Document,
    ) -> Result<impl Stream<Item = Result<Document, Self::Error>>, Self::Error> {
        Ok(futures::stream::iter(vec![Err(MockError(
            "find is not scripted",
        ))]))
    }
}
