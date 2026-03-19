use std::{
    collections::{BTreeSet, HashMap, HashSet},
    sync::Arc,
};

use crate::NamespaceInfoWithSchema;
use agg_ast::Namespace;
use mongosql::schema::Schema;
use tokio::sync::RwLock;
use tracing::debug;
mod catalog;

pub use catalog::Catalog;

/// A shareable handle to a ResultSet, allowing for concurrent reads,
/// but single writers.
pub type ShareableResultSet = Arc<RwLock<ResultSet>>;

/// ResultSet holds the catalog of schemas and tracks which namespaces changed
/// (or were marked unstable) during a run.
#[derive(Debug)]
pub struct ResultSet {
    /// Mapping of Database name to catalog
    schemas: HashMap<String, Catalog>,

    /// Set of changed database:namespace pairs
    ///
    /// The downstream consumer of this library only cares about schemas updated
    /// as part of the reconciliation process, so we keep track of the namespaces
    /// updated to later use as a filter when returning back changed schemas.
    changed: HashSet<(String, String)>,

    /// Namespaces whose initial schema was marked unstable (skipped for updates).
    /// Use `mark_unstable_initial_schema` and `into_unstable_initial_schemas` to write and read.
    unstable_initial_schemas: HashSet<(String, String)>,
}

/// Result of consuming a [`ResultSet`]: schemas that were updated this run,
/// and namespaces that were skipped because their initial schema was unstable.
#[derive(Debug)]
pub struct SchemaBuildOutput {
    /// Schemas that changed during the run and should be written.
    pub changed_schemas: Vec<NamespaceInfoWithSchema>,
    /// Namespaces `(db_name, collection)` skipped because the initial schema was unstable.
    pub skipped_unstable_namespaces: HashSet<(String, String)>,
}

impl ResultSet {
    /// Creates a new ResultSet with the given initial schemas.
    pub fn new(initial_schemas: HashMap<String, Catalog>) -> Self {
        // Log the initial schemas that we see upon construction.
        initial_schemas.iter().for_each(|(database, catalog)| {
            catalog.iter().for_each(|(namespace, _)| {
                debug!("Adding initial schema for {database}.{namespace}")
            })
        });

        Self {
            schemas: initial_schemas,
            changed: Default::default(),
            unstable_initial_schemas: Default::default(),
        }
    }

    /// Get a specific schema by database and collection/view name
    pub fn get_schema_for_database(&self, db_name: &str) -> Option<&Catalog> {
        self.schemas.get(db_name)
    }

    /// Get schemas for specific collections in a database
    pub fn get_schemas_for_namespaces<'a>(
        &self,
        namespaces: &'a BTreeSet<Namespace>,
    ) -> Option<HashMap<&'a str, Arc<Schema>>> {
        namespaces
            .iter()
            .filter_map(|namespace| {
                self.schemas
                    .get(&namespace.database)
                    .and_then(|db_catalog: &Catalog| db_catalog.get(&namespace.collection))
                    .map(|schema| {
                        (
                            namespace.collection.as_str(),
                            Arc::clone(&schema.namespace_schema),
                        )
                    })
            })
            .collect::<HashMap<_, _>>()
            .into()
    }

    /// Add a full schema to the result set
    pub(crate) fn add_schema(&mut self, schema: NamespaceInfoWithSchema) {
        let db_name = schema.namespace_info.db_name.as_str();
        let coll_or_view_name = schema.namespace_info.coll_or_view_name.as_str();

        debug!("Adding schema for {}.{}", db_name, coll_or_view_name);
        let db_catalog = self.schemas.entry(db_name.to_string()).or_default();

        // We only care about updating the schema if it either has not been seen
        // or has not changed since the last invocation.
        //
        // NOTE: This equality operation is NOT cheap, and might cause performance
        // issues on large enough schemas with multiple keys.
        if let Some(old) = db_catalog.get(coll_or_view_name)
            && old == &schema
        {
            debug!(
                "Skipping insertion of schema in {}:{} as it has not changed",
                db_name, coll_or_view_name
            );
        } else {
            debug!("Adding schema for {}.{}", db_name, coll_or_view_name);

            self.changed
                .insert((db_name.to_string(), coll_or_view_name.to_string()));
            db_catalog.insert(coll_or_view_name.to_string(), schema);
        }
    }
    pub(crate) fn mark_unstable_initial_schema(&mut self, db: String, collection: String) {
        self.unstable_initial_schemas.insert((db, collection));
    }

    /// Mark a namespace as having been changed
    ///
    /// This is mostly to allow dry-runs to short-circuit before doing
    /// any computationally expensive tasks while still notifying the downstream
    /// consumer to see which namespaces would have been modified.
    pub(crate) fn mark_as_changed(&mut self, db: String, collection: String) {
        self.changed.insert((db, collection));
    }

    /// Return the set of changed (database, namespace) pairs during this run.
    ///
    /// If you want the schemas instead, refer to `into_inner`
    pub fn into_changed(self) -> HashSet<(String, String)> {
        self.changed
    }

    /// Return the set of unstable initial schemas (namespaces skipped because
    /// their initial schema was marked unstable).
    pub fn into_unstable_initial_schemas(self) -> HashSet<(String, String)> {
        self.unstable_initial_schemas
    }

    /// Collects schemas whose (db_name, namespace) is in `changed` into a vec.
    fn collect_changed_schemas(
        schemas: HashMap<String, Catalog>,
        changed: HashSet<(String, String)>,
    ) -> Vec<NamespaceInfoWithSchema> {
        schemas
            .into_iter()
            .flat_map(|(_, catalog)| {
                catalog.into_iter().filter_map(|(namespace, schema)| {
                    changed
                        .contains(&(schema.namespace_info.db_name.clone(), namespace))
                        .then_some(schema)
                })
            })
            .collect()
    }

    /// Consumes the result set and returns the build output: changed schemas to write
    /// and namespaces that were skipped because their initial schema was unstable.
    pub fn into_build_output(self) -> SchemaBuildOutput {
        SchemaBuildOutput {
            changed_schemas: Self::collect_changed_schemas(self.schemas, self.changed),
            skipped_unstable_namespaces: self.unstable_initial_schemas,
        }
    }

    /// Convert the result set into a list of namespace, schema pairs that have
    /// been changed through the course of this program.
    ///
    /// Note that this does not include any initial schemas that have not been
    /// processed during execution.
    pub fn into_inner(self) -> Vec<NamespaceInfoWithSchema> {
        Self::collect_changed_schemas(self.schemas, self.changed)
    }
}
