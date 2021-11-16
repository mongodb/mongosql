use crate::{
    catalog::*,
    ir::{binding_tuple, *},
    map,
    schema::{
        Atomic, Document, ResultSet, Satisfaction, Schema, SchemaEnvironment, ANY_ARRAY,
        ANY_ARRAY_OR_NULLISH, ANY_DOCUMENT, BOOLEAN_OR_NULLISH, DATE_OR_NULLISH, EMPTY_DOCUMENT,
        INTEGER_OR_NULLISH, NULLISH, NUMERIC_OR_NULLISH, STRING_OR_NULLISH,
    },
    set,
    util::unique_linked_hash_map::UniqueLinkedHashMap,
};
use std::{
    cell::RefCell,
    cmp::min,
    collections::{BTreeMap, BTreeSet},
};
use thiserror::Error;

#[derive(Debug, Error, PartialEq, Clone)]
pub enum Error {
    #[error("datasource {0:?} not found in schema environment")]
    DatasourceNotFoundInSchemaEnv(binding_tuple::Key),
    #[error("incorrect argument count for {name}: required {required}, found {found}")]
    IncorrectArgumentCount {
        name: &'static str,
        required: usize,
        found: usize,
    },
    #[error("schema checking failed for {name}: required {required:?}, found {found:?}")]
    SchemaChecking {
        name: &'static str,
        required: Schema,
        found: Schema,
    },
    #[error(
        "cannot have {0:?} aggregations over the schema: {1:?} as it is not comparable to itself"
    )]
    AggregationArgumentMustBeSelfComparable(String, Schema),
    #[error("COUNT(DISTINCT *) is not supported")]
    CountDistinctStarNotSupported,
    #[error("invalid comparison for {0}: {1:?} cannot be compared to {2:?}")]
    InvalidComparison(&'static str, Schema, Schema),
    #[error("cannot merge objects {0:?} and {1:?} as they {2:?} have overlapping keys")]
    CannotMergeObjects(Schema, Schema, Satisfaction),
    #[error("cannot access field {0} because it does not exist")]
    AccessMissingField(String),
    #[error("cardinality of the subquery's result set may be greater than 1")]
    InvalidSubqueryCardinality,
    #[error("cannot create schema environment with duplicate datasource: {0:?}")]
    DuplicateKey(binding_tuple::Key),
    #[error("sort key at position {0} is not statically comparable to itself because it has the schema {1:?}")]
    SortKeyNotSelfComparable(usize, Schema),
    #[error("group key at position {0} is not statically comparable to itself because it has the schema {1:?}")]
    GroupKeyNotSelfComparable(usize, Schema),
    #[error("group key at position {0} is an unaliased field access with no datasource reference")]
    UnaliasedFieldAccessWithNoReference(usize),
    #[error("group key at position {0} is an unaliased non-field access expression")]
    UnaliasedNonFieldAccessExpression(usize),
}

#[derive(PartialEq, Clone, Debug)]
struct SchemaCacheContents<T: Clone> {
    // SchemaCacheContents stores everything from a SchemaInferenceState but the Catalog reference
    // in order to avoid specifying a lifetime.
    env: SchemaEnvironment,
    scope_level: u16,
    result: T,
}

impl<T: Clone> SchemaCacheContents<T> {
    fn new(state: &SchemaInferenceState, result: T) -> SchemaCacheContents<T> {
        SchemaCacheContents {
            env: state.env.clone(),
            scope_level: state.scope_level,
            result,
        }
    }
    /// Check if the passed-in state will result in a cache hit.
    fn matches(&self, state: &SchemaInferenceState) -> bool {
        self.env == state.env && self.scope_level == state.scope_level
    }
}

/// SchemaCache caches a `Result` because `schema()` methods are fallible and always return a Result.
#[derive(Clone, Debug)]
pub struct SchemaCache<T: Clone>(RefCell<Option<SchemaCacheContents<Result<T, Error>>>>);

impl<T: Clone> PartialEq for SchemaCache<T> {
    fn eq(&self, _other: &Self) -> bool {
        // Always return true since stages and expressions derive PartialEq, and we shouldn't care
        // about the contents of the SchemaCache of the stage or expression when comparing structs
        // which contain caches.
        true
    }
}

impl<T: Clone> SchemaCache<T> {
    pub fn new() -> Self {
        Self(RefCell::new(None))
    }
}

impl<T: Clone> Default for SchemaCache<T> {
    fn default() -> Self {
        Self::new()
    }
}

pub trait CachedSchema {
    type ReturnType: Clone;

    /// Get the stored RefCell from the underlying struct this trait is implemented on if it exists.
    /// Some types don't benefit from caching (literal and reference expressions specifically),
    /// so this function returns an Option that may or may not contain a cache.
    fn get_cache(&self) -> &SchemaCache<Self::ReturnType>;

    /// Get the cached result if the cache exists and it contains a value.
    fn get_cached_schema(&self) -> Option<Result<Self::ReturnType, Error>> {
        match self.get_cache().0.borrow().clone() {
            Some(c) => Some(c.result),
            None => None,
        }
    }

    /// Run schema checking algorithm for a node.
    fn check_schema(&self, state: &SchemaInferenceState) -> Result<Self::ReturnType, Error>;

    /// Cached call to `check_schema()`.
    fn schema(&self, state: &SchemaInferenceState) -> Result<Self::ReturnType, Error> {
        // This mutable borrow of the cache will be released when schema() is complete.
        let mut cache = self.get_cache().0.borrow_mut();
        match &*cache {
            Some(contents) if contents.matches(state) => contents.result.clone(),
            _ => {
                let schema_result = self.check_schema(state);
                cache.replace(SchemaCacheContents::new(state, schema_result.clone()));
                schema_result
            }
        }
    }
}

#[derive(Debug, Clone)]
pub struct SchemaInferenceState<'a> {
    pub scope_level: u16,
    pub env: SchemaEnvironment,
    pub catalog: &'a Catalog,
}

impl<'a> SchemaInferenceState<'a> {
    pub fn new(
        scope_level: u16,
        env: SchemaEnvironment,
        catalog: &'a Catalog,
    ) -> SchemaInferenceState {
        SchemaInferenceState {
            scope_level,
            env,
            catalog,
        }
    }

    pub fn with_merged_schema_env(
        &self,
        env: SchemaEnvironment,
    ) -> Result<SchemaInferenceState, Error> {
        Ok(SchemaInferenceState {
            env: env
                .with_merged_mappings(self.env.clone())
                .map_err(|e| Error::DuplicateKey(e.key))?,
            catalog: self.catalog,
            scope_level: self.scope_level,
        })
    }

    pub fn subquery_state(&self) -> SchemaInferenceState {
        SchemaInferenceState {
            scope_level: self.scope_level + 1,
            env: self.env.clone(),
            catalog: self.catalog,
        }
    }
}

impl CachedSchema for Stage {
    type ReturnType = ResultSet;

    fn get_cache(&self) -> &SchemaCache<Self::ReturnType> {
        match self {
            Stage::Filter(s) => &s.cache,
            Stage::Project(s) => &s.cache,
            Stage::Group(s) => &s.cache,
            Stage::Limit(s) => &s.cache,
            Stage::Offset(s) => &s.cache,
            Stage::Sort(s) => &s.cache,
            Stage::Collection(s) => &s.cache,
            Stage::Array(s) => &s.cache,
            Stage::Join(s) => &s.cache,
            Stage::Set(s) => &s.cache,
        }
    }

    /// Recursively schema checks this stage, its sources, and all
    /// contained expressions. If schema checking succeeds, returns a
    /// [`ResultSet`] describing the schema of the result set returned
    /// by this stage. The provided [`SchemaInferenceState`] should
    /// include schema information for any datasources from outer
    /// scopes. Schema information for the current scope will be
    /// obtained by calling [`Stage::schema`] on source stages.
    fn check_schema(&self, state: &SchemaInferenceState) -> Result<ResultSet, Error> {
        match self {
            Stage::Filter(f) => {
                let source_result_set = f.source.schema(state)?;
                let state = state.with_merged_schema_env(source_result_set.schema_env.clone())?;
                let cond_schema = f.condition.schema(&state)?;
                if cond_schema.satisfies(&BOOLEAN_OR_NULLISH) != Satisfaction::Must {
                    return Err(Error::SchemaChecking {
                        name: "filter condition",
                        required: BOOLEAN_OR_NULLISH.clone(),
                        found: cond_schema,
                    });
                }
                Ok(ResultSet {
                    schema_env: source_result_set.schema_env,
                    min_size: 0,
                    max_size: source_result_set.max_size,
                })
            }
            Stage::Project(p) => {
                let source_result_set = p.source.schema(state)?;
                let (min_size, max_size) = (source_result_set.min_size, source_result_set.max_size);
                let state = state.with_merged_schema_env(source_result_set.schema_env)?;
                let schema_env = p
                    .expression
                    .iter()
                    .map(|(k, e)| match e.schema(&state) {
                        Ok(s) => {
                            if s.satisfies(&ANY_DOCUMENT) != Satisfaction::Must {
                                Err(Error::SchemaChecking {
                                    name: "project datasource",
                                    required: ANY_DOCUMENT.clone(),
                                    found: s,
                                })
                            } else {
                                Ok((k.clone(), s))
                            }
                        }
                        Err(e) => Err(e),
                    })
                    .collect::<Result<SchemaEnvironment, _>>()?;
                Ok(ResultSet {
                    schema_env,
                    min_size,
                    max_size,
                })
            }
            Stage::Group(g) => {
                let source_result_set = g.source.schema(state)?;
                let state = state.with_merged_schema_env(source_result_set.schema_env.clone())?;

                // If all group keys are literals, the result set max size is 1.
                // Otherwise, the max is derived from the source's result set.
                let max_size = if g
                    .keys
                    .iter()
                    .all(|key| matches!(key.get_expr(), Expression::Literal(_)))
                {
                    Some(1)
                } else {
                    source_result_set.max_size
                };

                // A helper to bind a field/alias to a schema.
                // Used for embedding a group key or aggregation function under a datasource.
                let schema_binding_doc = |field_or_alias: String, schema: Schema| {
                    Schema::Document(Document {
                        keys: map! { field_or_alias.clone() => schema },
                        required: set! { field_or_alias },
                        additional_properties: false,
                    })
                };

                let schema_env = g
                    .keys
                    .iter()
                    .enumerate()
                    .map(|(index, key)| {
                        // Schema check the group key and upconvert Missing to Null.
                        let group_key_schema = key
                            .get_expr()
                            .schema(&state)
                            .map(|s| s.upconvert_missing_to_null())?;

                        // The group key must have a schema that is self-comparable.
                        if group_key_schema.is_self_comparable() != Satisfaction::Must {
                            return Err(Error::GroupKeyNotSelfComparable(index, group_key_schema));
                        }

                        // Get the SchemaEnvironment BindingTuple key and its associated schema.
                        let (schema_env_key, schema_env_schema) = match key {
                            // If the group key has an alias, bind a document containing that alias
                            // and the group key's schema to the Bottom datasource.
                            OptionallyAliasedExpr::Aliased(AliasedExpr { expr: _, ref alias }) => (
                                binding_tuple::Key::bot(state.scope_level),
                                schema_binding_doc(alias.clone(), group_key_schema),
                            ),
                            // Otherwise for a field access group key, bind a document containing the
                            // field access string and the group key's schema to the Reference datasource.
                            OptionallyAliasedExpr::Unaliased(ref expr) => match expr {
                                Expression::FieldAccess(f) => match f.expr.as_ref() {
                                    Expression::Reference(r) => (
                                        r.key.clone(),
                                        schema_binding_doc(f.field.clone(), group_key_schema),
                                    ),
                                    _ => {
                                        return Err(Error::UnaliasedFieldAccessWithNoReference(
                                            index,
                                        ))
                                    }
                                },
                                _ => return Err(Error::UnaliasedNonFieldAccessExpression(index)),
                            },
                        };

                        Ok((schema_env_key, schema_env_schema))
                    })
                    // Since we may have multiple tuples with the same datasource key (e.g. two aliased group
                    // keys will both belong to the Bottom datasource), it is incorrect to collect() into a
                    // SchemaEnvironment here, as collect() will overwrite any tuples with the same key.
                    // We handle this using a fold operation, which allows us to handle duplicate tuples by
                    // manually unioning them with an accumulating SchemaEnvironment.
                    .fold(
                        Ok(SchemaEnvironment::default()),
                        |acc_schema_env, tuple_result| {
                            let (schema_env_key, schema_env_schema) = tuple_result?;
                            Ok(acc_schema_env?
                                .union_schema_for_datasource(schema_env_key, schema_env_schema))
                        },
                    )?;

                let mut schema_env = g
                    .aggregations
                    .iter()
                    // Bind a document containing each aggregation function's alias
                    // and the aggregation function schema to the Bottom datasource.
                    .map(|aliased_agg| {
                        let agg_schema = aliased_agg.agg_expr.schema(&state)?;
                        Ok((
                            binding_tuple::Key::bot(state.scope_level),
                            schema_binding_doc(aliased_agg.alias.clone(), agg_schema),
                        ))
                    })
                    // See the comment above the group key folding; the same requirement
                    // applies to aggregation functions.
                    .fold(
                        Ok(SchemaEnvironment::default()),
                        |acc_schema_env, tuple_result| {
                            let (schema_env_key, schema_env_schema) = tuple_result?;
                            Ok(acc_schema_env?
                                .union_schema_for_datasource(schema_env_key, schema_env_schema))
                        },
                    )?
                    // Union the aggregation function bindings with the group key bindings.
                    .union(schema_env);

                // The calls to union() and union_schema_for_datasource() when constructing the schema environment
                // combines schemas with the same Key under an AnyOf. However, we want group keys and aggregations under
                // bottom to be combined under a single document. To accomplish this, we pull out the schema for bottom,
                // and if it's an AnyOf set of documents, use Document::merge() to combine the documents.
                let schema_env = if let Some(bot_schema) =
                    schema_env.remove(&binding_tuple::Key::bot(state.scope_level))
                {
                    // Call simplify() to flatten the AnyOfs into a single set.
                    if let Schema::AnyOf(schemas) = Schema::simplify(&bot_schema) {
                        let merged_document =
                            schemas.into_iter().fold(Document::empty(), |acc, schema| {
                                acc.merge(match schema {
                                    Schema::Document(doc) => doc,
                                    // Bottom is always a document according to the spec, and so
                                    // its schema can only be Document or AnyOf(documents). After
                                    // simplification, we know that we're inside the only AnyOf,
                                    // so there should only be documents.
                                    _ => unreachable!(),
                                })
                            });
                        schema_env.union_schema_for_datasource(
                            binding_tuple::Key::bot(state.scope_level),
                            Schema::Document(merged_document),
                        )
                    } else {
                        // Put the value back unchanged if the top-level bottom schema wasn't AnyOf.
                        schema_env.union_schema_for_datasource(
                            binding_tuple::Key::bot(state.scope_level),
                            bot_schema,
                        )
                    }
                } else {
                    schema_env
                };

                Ok(ResultSet {
                    schema_env,
                    min_size: source_result_set.min_size,
                    max_size,
                })
            }
            Stage::Limit(l) => {
                let source_result_set = l.source.schema(state)?;
                Ok(ResultSet {
                    schema_env: source_result_set.schema_env,
                    min_size: min(l.limit, source_result_set.min_size),
                    max_size: source_result_set
                        .max_size
                        .map_or(Some(l.limit), |x| Some(min(l.limit, x))),
                })
            }
            Stage::Offset(o) => {
                let source_result_set = o.source.schema(state)?;
                Ok(ResultSet {
                    schema_env: source_result_set.schema_env,
                    min_size: source_result_set.min_size.saturating_sub(o.offset),
                    max_size: source_result_set
                        .max_size
                        .map(|x| x.saturating_sub(o.offset)),
                })
            }
            Stage::Sort(s) => {
                let source_result_set = s.source.schema(state)?;
                let state = state.with_merged_schema_env(source_result_set.schema_env.clone())?;

                // Ensure that each sort key is statically comparable to itself.
                s.specs
                    .clone()
                    .iter()
                    .enumerate()
                    .try_for_each(|(index, spec)| match spec {
                        SortSpecification::Asc(a) | SortSpecification::Desc(a) => {
                            let schema = a.schema(&state)?;
                            if schema.is_self_comparable() != Satisfaction::Must {
                                return Err(Error::SortKeyNotSelfComparable(index, schema));
                            }
                            Ok(())
                        }
                    })?;
                Ok(source_result_set)
            }
            Stage::Collection(c) => {
                let schema = match state.catalog.get_schema_for_namespace(&Namespace {
                    db: c.db.clone(),
                    collection: c.collection.clone(),
                }) {
                    Some(s) => s.clone(),
                    None => ANY_DOCUMENT.clone(),
                };
                Ok(ResultSet {
                    schema_env: map! {
                        (c.collection.clone(), state.scope_level).into() => schema,
                    },
                    min_size: 0,
                    max_size: None,
                })
            }
            Stage::Array(a) => {
                let array_items_schema = Expression::array_items_schema(&a.array, state)?;
                if array_items_schema.satisfies(&ANY_DOCUMENT) == Satisfaction::Must {
                    Ok(ResultSet {
                        schema_env: map! {
                            (a.alias.clone(), state.scope_level).into() => array_items_schema,
                        },
                        min_size: a.array.len() as u64,
                        max_size: Some(a.array.len() as u64),
                    })
                } else {
                    Err(Error::SchemaChecking {
                        name: "array datasource items",
                        required: ANY_DOCUMENT.clone(),
                        found: array_items_schema,
                    })
                }
            }
            Stage::Join(j) => {
                let left_result_set = j.left.schema(state)?;
                let right_result_set = j.right.schema(state)?;

                let state = state.with_merged_schema_env(left_result_set.schema_env.clone())?;
                let state = state.with_merged_schema_env(right_result_set.schema_env.clone())?;
                if let Some(e) = &j.condition {
                    let cond_schema = e.schema(&state)?;
                    if cond_schema.satisfies(&BOOLEAN_OR_NULLISH) != Satisfaction::Must {
                        return Err(Error::SchemaChecking {
                            name: "join condition",
                            required: BOOLEAN_OR_NULLISH.clone(),
                            found: cond_schema,
                        });
                    }
                };
                match j.join_type {
                    JoinType::Left => {
                        let min_size = left_result_set.min_size;
                        let max_size = left_result_set
                            .max_size
                            .and_then(|l| right_result_set.max_size.map(|r| l * r));
                        Ok(ResultSet {
                            schema_env: left_result_set
                                .schema_env
                                .with_merged_mappings(
                                    right_result_set
                                        .schema_env
                                        .into_iter()
                                        .map(|(key, schema)| {
                                            (key, Schema::AnyOf(set![Schema::Missing, schema]))
                                        })
                                        .collect::<SchemaEnvironment>(),
                                )
                                .map_err(|e| Error::DuplicateKey(e.key))?,
                            min_size,
                            max_size,
                        })
                    }
                    JoinType::Inner => {
                        // If the join is a cross join, set min_size to the cardinality of the
                        // Cartesian product of the left and right result sets. Set max_size to this
                        // value for both cross and inner joins.
                        let min_size = match j.condition {
                            None => left_result_set.min_size * right_result_set.min_size,
                            Some(_) => 0,
                        };
                        let max_size = left_result_set
                            .max_size
                            .and_then(|l| right_result_set.max_size.map(|r| l * r));
                        Ok(ResultSet {
                            schema_env: left_result_set
                                .schema_env
                                .with_merged_mappings(right_result_set.schema_env)
                                .map_err(|e| Error::DuplicateKey(e.key))?,
                            min_size,
                            max_size,
                        })
                    }
                }
            }
            Stage::Set(s) => {
                let left_result_set = s.left.schema(state)?;
                let right_result_set = s.right.schema(state)?;
                let max_size = left_result_set
                    .max_size
                    .and_then(|l| right_result_set.max_size.map(|r| l + r));
                match s.operation {
                    SetOperation::Union => Ok(ResultSet {
                        schema_env: left_result_set
                            .schema_env
                            .union(right_result_set.schema_env),
                        min_size: min(left_result_set.min_size + right_result_set.min_size, 1),
                        max_size,
                    }),
                    SetOperation::UnionAll => Ok(ResultSet {
                        schema_env: left_result_set
                            .schema_env
                            .union(right_result_set.schema_env),
                        min_size: left_result_set.min_size + right_result_set.min_size,
                        max_size,
                    }),
                }
            }
        }
    }
}

impl AggregationExpr {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        match self {
            AggregationExpr::CountStar(distinct) => {
                if *distinct {
                    Err(Error::CountDistinctStarNotSupported)
                } else {
                    Ok(Schema::AnyOf(set![
                        Schema::Atomic(Atomic::Integer),
                        Schema::Atomic(Atomic::Long)
                    ]))
                }
            }
            AggregationExpr::Function(a) => a.schema(state),
        }
    }
}

impl AggregationFunctionApplication {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let arg_schema = self.arg.schema(state)?;
        if self.distinct && arg_schema.is_self_comparable() != Satisfaction::Must {
            return Err(Error::AggregationArgumentMustBeSelfComparable(
                format!("{} DISTINCT", self.function.as_str()),
                arg_schema,
            ));
        }
        self.function.schema(arg_schema)
    }
}

impl AggregationFunction {
    pub fn schema(&self, arg_schema: Schema) -> Result<Schema, Error> {
        use crate::ir::AggregationFunction::*;
        Ok(match self {
            AddToArray => Schema::Array(Box::new(arg_schema)),
            Avg | StddevPop | StddevSamp => {
                self.schema_check_fixed_args(&[arg_schema.clone()], &[NUMERIC_OR_NULLISH.clone()])?;
                // we cannot use get_arithmetic_schema for Avg, StddevPop, StddevSamp
                // because they never return Long or Integer results, even for Long
                // or Integer inputs.
                let get_numeric_schema =
                    || match arg_schema.satisfies(&Schema::Atomic(Atomic::Decimal)) {
                        Satisfaction::Not => Schema::Atomic(Atomic::Double),
                        Satisfaction::Must => Schema::Atomic(Atomic::Decimal),
                        Satisfaction::May => Schema::AnyOf(set![
                            Schema::Atomic(Atomic::Double),
                            Schema::Atomic(Atomic::Decimal),
                        ]),
                    };
                match arg_schema.satisfies(&NULLISH) {
                    Satisfaction::Not => get_numeric_schema(),
                    Satisfaction::Must => Schema::Atomic(Atomic::Null),
                    Satisfaction::May => {
                        Schema::AnyOf(set![get_numeric_schema(), Schema::Atomic(Atomic::Null)])
                    }
                }
            }
            Count => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long)
            ]),
            First | Last => arg_schema,
            Min | Max => {
                if arg_schema.is_self_comparable() != Satisfaction::Must {
                    return Err(Error::AggregationArgumentMustBeSelfComparable(
                        self.as_str().to_string(),
                        arg_schema,
                    ));
                }
                arg_schema
            }
            MergeDocuments => {
                self.schema_check_fixed_args(&[arg_schema.clone()], &[ANY_DOCUMENT.clone()])?;
                arg_schema
            }
            Sum => self.get_arithmetic_schema(&[arg_schema])?,
        })
    }
}

impl SubqueryExpr {
    /// Subquery schema helper used by both subquery expressions and subquery comparisons.
    /// Returns the schema of the subquery's output expression along with the min and max sizes of
    /// the result set of the subquery.
    fn schema_helper(
        &self,
        state: &SchemaInferenceState,
    ) -> Result<(Schema, u64, Option<u64>), Error> {
        let result_set = self.subquery.schema(&state.subquery_state())?;
        let max_size = result_set.max_size;
        let min_size = result_set.min_size;
        let schema = self.output_expr.schema(&SchemaInferenceState {
            scope_level: state.scope_level + 1,
            env: result_set.schema_env,
            catalog: state.catalog,
        })?;
        Ok((schema, min_size, max_size))
    }

    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let (schema, min_size, max_size) = self.schema_helper(state)?;

        // The subquery's result set MUST have a cardinality of 0 or 1. The returned schema
        // MUST include MISSING if the cardinality of the result set MAY be 0. We can exclude
        // MISSING from the returned schema if the cardinality of the result set MUST be 1.
        if max_size == None || max_size.unwrap() > 1 {
            return Err(Error::InvalidSubqueryCardinality);
        }
        match min_size {
            0 => Ok(Schema::AnyOf(set![schema, Schema::Missing])),
            1 => Ok(schema),
            _ => Err(Error::InvalidSubqueryCardinality),
        }
    }
}

/// While a `SubqueryComparison` isn't strictly a `SQLFunction`, the actual comparison operation re-uses
/// the same schema-checking that any binary comparison operator (Lt, Gt, Eq, etc.) uses. Those comparisons
/// are represented as `ScalarFunction`s, and so that's where the schema checking lives.
impl SQLFunction for SubqueryComparison {
    fn as_str(&self) -> &'static str {
        "subquery comparison"
    }
}

impl SubqueryComparison {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let argument_schema = self.argument.schema(state)?;
        let (subquery_schema, _, _) = self.subquery_expr.schema_helper(state)?;
        self.get_comparison_schema(&[argument_schema, subquery_schema])
    }
}

impl Expression {
    // get_field_schema returns the Schema for a known field name retrieved
    // from the argument Schema. It follows the MongoSQL semantics for
    // path access.
    //
    // If it is possible for the argument Schema
    // to be a non-Document or the key does not exist in the Document, Missing will
    // be part of the returned Schema.
    pub fn get_field_schema(s: &Schema, field: &str) -> Schema {
        // If self is Any, it may contain the field, but we don't know the
        // Schema. If it's a Document we need to do more investigation.
        // If it's AnyOf we need to apply get_field_schema to each
        // sub-schema and apply AnyOf to the results as appropriate.
        // If it's anything else it will Not contain the field, so we return Missing.
        let d = match s {
            Schema::Any => return Schema::Any,
            Schema::Document(d) => d,
            Schema::AnyOf(vs) => {
                return Schema::AnyOf(
                    vs.iter()
                        .map(|s| Expression::get_field_schema(s, field))
                        .collect(),
                );
            }
            Schema::Missing | Schema::Array(_) | Schema::Atomic(_) => return Schema::Missing,
            Schema::Unsat => return Schema::Unsat,
        };
        // If we find the field in the Document, we just return
        // the Schema for that field, unless the field is not required,
        // then we return AnyOf(Schema, Missing).
        if let Some(s) = d.keys.get(field) {
            if d.required.contains(field) {
                return s.clone();
            }
            return Schema::AnyOf(set![s.clone(), Schema::Missing]);
        }
        // If the schema allows additional_properties, it May exist,
        // regardless of its presence or absence in keys, but we don't know the Schema.
        // If the key is in required, it Must exist, but we don't know the
        // Schema because it wasn't in keys.
        // Either way, the resulting Schema must be Any.
        if d.additional_properties || d.required.contains(field) {
            return Schema::Any;
        }
        // We return Missing because the field Must not exist.
        Schema::Missing
    }

    fn array_schema(a: &[Expression], state: &SchemaInferenceState) -> Result<Schema, Error> {
        Ok(Schema::Array(Box::new(Expression::array_items_schema(
            a, state,
        )?)))
    }

    /// For array literals, we return Array(Unsat) (isomorphic with Array(AnyOf([])) for an empty array.
    /// For other arrays we return Array(AnyOf(S1...SN)), for Schemata S1...SN inferred from the element
    /// Expressions of the array literal.
    fn array_items_schema(a: &[Expression], state: &SchemaInferenceState) -> Result<Schema, Error> {
        Ok(Schema::AnyOf(
            a.iter()
                .map(|e| e.schema(state).map(|s| s.upconvert_missing_to_null()))
                .collect::<Result<_, _>>()?,
        ))
    }

    /// For document literals, we infer the most restrictive schema possible. This means
    /// that additional_properties are not allowed.
    fn document_schema(
        d: &UniqueLinkedHashMap<String, Expression>,
        state: &SchemaInferenceState,
    ) -> Result<Schema, Error> {
        let (mut keys, mut required) = (BTreeMap::new(), BTreeSet::new());
        for (key, e) in d.iter() {
            let key_schema = e.schema(state)?;
            match key_schema.satisfies(&Schema::Missing) {
                Satisfaction::Not => {
                    required.insert(key.clone());
                    keys.insert(key.clone(), key_schema);
                }
                Satisfaction::May => {
                    keys.insert(key.clone(), key_schema);
                }
                Satisfaction::Must => (),
            }
        }
        Ok(Schema::Document(Document {
            keys,
            required,
            additional_properties: false,
        }))
    }
}

impl CachedSchema for Expression {
    type ReturnType = Schema;

    fn get_cache(&self) -> &SchemaCache<Self::ReturnType> {
        match self {
            Expression::Literal(s) => &s.cache,
            Expression::Reference(r) => &r.cache,
            Expression::Array(a) => &a.cache,
            Expression::Document(d) => &d.cache, // TODO: investigate how to add a cache on the hashmap
            Expression::ScalarFunction(s) => &s.cache,
            Expression::Cast(s) => &s.cache,
            Expression::SearchedCase(s) => &s.cache,
            Expression::SimpleCase(s) => &s.cache,
            Expression::TypeAssertion(s) => &s.cache,
            Expression::Is(s) => &s.cache,
            Expression::Like(s) => &s.cache,
            Expression::FieldAccess(s) => &s.cache,
            Expression::Subquery(s) => &s.cache,
            Expression::SubqueryComparison(s) => &s.cache,
            Expression::Exists(s) => &s.cache, // schema is always always boolean after checking the boxed stage, which is cached
        }
    }

    /// Recursively schema checks this expression, its arguments, and
    /// all contained expressions/stages. If schema checking succeeds,
    /// returns a [`Schema`] describing this expression's schema. The
    /// provided [`SchemaInferenceState`] should include schema
    /// information for all datasources in scope.
    fn check_schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        match self {
            Expression::Literal(lit) => lit.value.schema(state),
            Expression::Reference(ReferenceExpr { key, .. }) => state
                .env
                .get(key)
                .cloned()
                .ok_or_else(|| Error::DatasourceNotFoundInSchemaEnv(key.clone())),
            Expression::Array(ArrayExpr { array, .. }) => Expression::array_schema(array, state),
            Expression::Document(DocumentExpr { document, .. }) => {
                Expression::document_schema(document, state)
            }
            Expression::FieldAccess(FieldAccess {
                expr,
                field,
                cache: _cache,
            }) => {
                let accessee_schema = expr.schema(state)?;
                if accessee_schema.satisfies(&ANY_DOCUMENT) == Satisfaction::Not {
                    return Err(Error::SchemaChecking {
                        name: "FieldAccess",
                        required: ANY_DOCUMENT.clone(),
                        found: accessee_schema,
                    });
                }
                if accessee_schema.contains_field(field) == Satisfaction::Not {
                    return Err(Error::AccessMissingField(field.clone()));
                }
                Ok(Expression::get_field_schema(&accessee_schema, field))
            }
            Expression::ScalarFunction(f) => f.schema(state),
            Expression::Cast(c) => c.schema(state),
            Expression::SearchedCase(sc) => sc.schema(state),
            Expression::SimpleCase(sc) => sc.schema(state),
            Expression::TypeAssertion(t) => t.schema(state),
            Expression::Subquery(s) => s.schema(state),
            Expression::SubqueryComparison(sc) => sc.schema(state),
            Expression::Exists(ExistsExpr { stage, .. }) => {
                stage.schema(&state.subquery_state())?;
                Ok(Schema::Atomic(Atomic::Boolean))
            }
            Expression::Is(i) => {
                i.expr.schema(state)?;
                Ok(Schema::Atomic(Atomic::Boolean))
            }
            Expression::Like(l) => {
                let expr_schema = l.expr.schema(state)?;
                if expr_schema.satisfies(&STRING_OR_NULLISH) != Satisfaction::Must {
                    return Err(Error::SchemaChecking {
                        name: "Like",
                        required: STRING_OR_NULLISH.clone(),
                        found: expr_schema,
                    });
                }
                let pattern_schema = l.pattern.schema(state)?;
                if pattern_schema.satisfies(&STRING_OR_NULLISH) != Satisfaction::Must {
                    return Err(Error::SchemaChecking {
                        name: "Like",
                        required: STRING_OR_NULLISH.clone(),
                        found: pattern_schema,
                    });
                }
                match expr_schema
                    .satisfies(&NULLISH)
                    .max(pattern_schema.satisfies(&NULLISH))
                {
                    Satisfaction::Not => Ok(Schema::Atomic(Atomic::Boolean)),
                    Satisfaction::May => Ok(Schema::AnyOf(set![
                        Schema::Atomic(Atomic::Boolean),
                        Schema::Atomic(Atomic::Null),
                    ])),
                    Satisfaction::Must => Ok(Schema::Atomic(Atomic::Null)),
                }
            }
        }
    }
}

impl LiteralValue {
    pub fn schema(&self, _state: &SchemaInferenceState) -> Result<Schema, Error> {
        use LiteralValue::*;
        Ok(match self {
            Null => Schema::Atomic(Atomic::Null),
            Boolean(_) => Schema::Atomic(Atomic::Boolean),
            String(_) => Schema::Atomic(Atomic::String),
            Integer(_) => Schema::Atomic(Atomic::Integer),
            Long(_) => Schema::Atomic(Atomic::Long),
            Double(_) => Schema::Atomic(Atomic::Double),
        })
    }
}

impl ScalarFunctionApplication {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let args = self
            .args
            .iter()
            .map(|x| x.schema(state))
            .collect::<Result<Vec<_>, _>>()?;
        self.function.schema(&args)
    }
}

/// retain is responsible for iterating through all schemas and retaining
/// the dominant numerics. Iteration allows for early termination if
/// certain heuristics are met such as encountering Atomic(Decimal) yields
/// Atomic(Decimal)
pub(crate) fn retain(schemas: &BTreeSet<Schema>) -> Result<Schema, Error> {
    use schema::{Atomic::*, Schema::*};
    fn retain_aux(atomic: &Schema, anyof: &BTreeSet<Schema>) -> Schema {
        let mut anyof = anyof.clone();
        let mut retained = anyof.split_off(atomic);
        if retained.is_empty() {
            return atomic.clone();
        }
        if !anyof.is_empty() {
            retained.insert(atomic.clone());
        }
        Schema::simplify(&AnyOf(retained))
    }

    if schemas.is_empty() {
        return Ok(Schema::Atomic(Integer));
    }
    let mut result = schemas.iter().next().unwrap().clone();

    for schema in schemas.iter() {
        if *schema == Schema::Atomic(Decimal) {
            return Ok(Schema::Atomic(Decimal));
        }
        match (result, schema) {
            (Atomic(a1), Atomic(a2)) => {
                result = max_numeric(&Schema::Atomic(a1), &Schema::Atomic(*a2)).unwrap();
            }
            (Atomic(at), AnyOf(ao)) => {
                result = retain_aux(&Schema::Atomic(at), ao);
            }
            (AnyOf(ao), Atomic(at)) => {
                result = retain_aux(&Schema::Atomic(*at), &ao);
            }
            (AnyOf(ao1), AnyOf(ao2)) => {
                let mut current_result = <BTreeSet<Schema>>::new();
                for anyof in ao1.iter() {
                    if let Atomic(at) = anyof {
                        current_result.insert(retain_aux(&Schema::Atomic(*at), ao2));
                    }
                }
                result = Schema::simplify(&AnyOf(current_result));
            }
            _ => unreachable!(),
        }
    }
    Ok(result)
}

/// Compares two atomics and returns max numeric type
pub(crate) fn max_numeric(a1: &Schema, a2: &Schema) -> Result<Schema, Error> {
    use schema::{Atomic::*, Schema::*};
    Ok(match (a1, a2) {
        (Atomic(Decimal), _) | (_, Atomic(Decimal)) => Atomic(Decimal),
        (Atomic(Double), _) | (_, Atomic(Double)) => Atomic(Double),
        (Atomic(Long), _) | (_, Atomic(Long)) => Atomic(Long),
        (Atomic(Integer), _) | (_, Atomic(Integer)) => Atomic(Integer),
        _ => unreachable!(),
    })
}

trait SQLFunction {
    /// Returns the schema for an arithmetic function, maximizing the Numeric type
    /// based on the total order: Integer < Long < Double < Decimal.
    ///
    /// If any argument May be Nullish we return AnyOf(Maxed Numeric Schema, Null). If
    /// any argument Must be NULLISH we return Null. If no argument may be NULLISH we return
    /// Maxed Numeric Schema.
    fn get_arithmetic_schema(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        use schema::{Atomic::*, Schema::*};

        fn get_arithmetic_schema_aux(schemas: &[Schema]) -> Result<Schema, Error> {
            let schemas_collected: BTreeSet<_> = schemas
                .iter()
                .map(Schema::simplify)
                // Remove Missing and Null from the schemas since the inclusion of Null
                // in the output is handled by schema_check_variadic_args below.
                .filter(|s| !matches!(s, Missing) && !matches!(s, Atomic(Null)))
                .map(|s| match s {
                    AnyOf(ao) => AnyOf(
                        ao.into_iter()
                            .filter(|sch| !matches!(sch, Missing) && !matches!(sch, Atomic(Null)))
                            .collect::<BTreeSet<_>>(),
                    ),
                    Atomic(a) if a.is_numeric() => s.clone(),
                    _ => unreachable!(),
                })
                .collect::<BTreeSet<_>>();
            retain(&schemas_collected)
        }

        match self.schema_check_variadic_args(arg_schemas, NUMERIC_OR_NULLISH.clone())? {
            Satisfaction::Must => Ok(Atomic(Null)),
            Satisfaction::Not => get_arithmetic_schema_aux(arg_schemas),
            Satisfaction::May => Ok(Schema::simplify(&AnyOf(set![
                get_arithmetic_schema_aux(arg_schemas)?,
                Atomic(Null),
            ]))),
        }
    }

    fn ensure_arg_count(&self, found: usize, required: usize) -> Result<(), Error> {
        if found != required {
            return Err(Error::IncorrectArgumentCount {
                name: self.as_str(),
                required,
                found,
            });
        }
        Ok(())
    }

    /// Checks a function's argument count and its arguments' types against the required schemas.
    /// Used for functions with a fixed (predetermined) number of arguments, including 0.
    ///
    /// Since the argument count is fixed, the slice of required schemas must correspond 1-to-1
    /// with the slice of argument schemas.
    ///
    /// The satisfaction result returns whether a NULLISH argument is a possibility.
    fn schema_check_fixed_args(
        &self,
        arg_schemas: &[Schema],
        required_schemas: &[Schema],
    ) -> Result<Satisfaction, Error> {
        self.ensure_arg_count(arg_schemas.len(), required_schemas.len())?;
        let mut total_null_sat = Satisfaction::Not;
        for (i, arg) in arg_schemas.iter().enumerate() {
            if arg.satisfies(&required_schemas[i]) != Satisfaction::Must {
                return Err(Error::SchemaChecking {
                    name: self.as_str(),
                    required: required_schemas[i].clone(),
                    found: arg.clone(),
                });
            }
            let sat = arg.satisfies(&NULLISH);
            if sat > total_null_sat {
                total_null_sat = sat;
            }
        }
        Ok(total_null_sat)
    }

    /// Checks a function's arguments' types against the required schema.
    /// Used for functions with a variadic number of arguments, including 0.
    ///
    /// Since the argument count can vary, the required schema is a single
    /// value that's compared against each of the arguments.
    fn schema_check_variadic_args(
        &self,
        arg_schemas: &[Schema],
        required_schema: Schema,
    ) -> Result<Satisfaction, Error> {
        self.schema_check_fixed_args(
            arg_schemas,
            &std::iter::repeat(required_schema)
                .take(arg_schemas.len())
                .collect::<Vec<Schema>>(),
        )
    }

    /// Helper function which takes a `Satisfaction` from either `schema_check_fixed_arguments()` or
    /// `schema_check_variadic_arguments()` and propagates a `NULL` return value if necessary.
    fn propagate_null_arguments_helper(
        &self,
        sat: Satisfaction,
        base_return_schema: Schema,
    ) -> Schema {
        match sat {
            Satisfaction::May => {
                Schema::AnyOf(set![base_return_schema, Schema::Atomic(Atomic::Null)])
            }
            Satisfaction::Not => base_return_schema,
            Satisfaction::Must => Schema::Atomic(Atomic::Null),
        }
    }

    /// Many scalar functions will return `NULL` if one of their arguments is `NULL`.
    /// This function factors out the schema checking for these scalar functions for a fixed
    /// number of arguments.
    fn propagate_fixed_null_arguments(
        &self,
        arg_schemas: &[Schema],
        required_schemas: &[Schema],
        base_return_schema: Schema,
    ) -> Result<Schema, Error> {
        Ok(self.propagate_null_arguments_helper(
            self.schema_check_fixed_args(arg_schemas, required_schemas)?,
            base_return_schema,
        ))
    }

    /// Many scalar functions will return `NULL` if one of their arguments is `NULL`.
    /// This function factors out the schema checking for these scalar functions for variadic
    /// functions.
    fn propagate_variadic_null_arguments(
        &self,
        arg_schemas: &[Schema],
        required_schema: Schema,
        base_return_schema: Schema,
    ) -> Result<Schema, Error> {
        Ok(self.propagate_null_arguments_helper(
            self.schema_check_variadic_args(arg_schemas, required_schema)?,
            base_return_schema,
        ))
    }

    /// Returns the boolean and/or null schema for a valid comparison, or an error
    /// if the number of operands is not two or the comparison is not valid.
    fn get_comparison_schema(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        let ret_schema = self.propagate_fixed_null_arguments(
            arg_schemas,
            &[Schema::Any, Schema::Any],
            Schema::Atomic(Atomic::Boolean),
        )?;

        if arg_schemas[0].is_comparable_with(&arg_schemas[1]) == Satisfaction::Must {
            Ok(ret_schema)
        } else {
            Err(Error::InvalidComparison(
                self.as_str(),
                arg_schemas[0].clone(),
                arg_schemas[1].clone(),
            ))
        }
    }

    /// as_str returns a static str representation for the SQLFunction.
    fn as_str(&self) -> &'static str;
}

impl SQLFunction for ScalarFunction {
    fn as_str(&self) -> &'static str {
        self.as_str()
    }
}

impl SQLFunction for AggregationFunction {
    fn as_str(&self) -> &'static str {
        self.as_str()
    }
}

impl ScalarFunction {
    pub fn schema(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        use ScalarFunction::*;
        match self {
            // String operators.
            Concat => self.propagate_fixed_null_arguments(
                arg_schemas,
                &[STRING_OR_NULLISH.clone(), STRING_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::String),
            ),
            // Unary arithmetic operators.
            Pos | Neg => {
                self.ensure_arg_count(arg_schemas.len(), 1)?;
                self.get_arithmetic_schema(arg_schemas)
            }
            // Arithmetic operators with variadic arguments.
            Add | Mul => self.get_arithmetic_schema(arg_schemas),
            // Arithmetic operators with fixed (two) arguments.
            Sub | Div => {
                self.ensure_arg_count(arg_schemas.len(), 2)?;
                self.get_arithmetic_schema(arg_schemas)
            }
            // Comparison operators.
            Lt | Lte | Neq | Eq | Gt | Gte => self.get_comparison_schema(arg_schemas),
            Between => {
                self.schema_check_fixed_args(
                    arg_schemas,
                    &[Schema::Any, Schema::Any, Schema::Any],
                )?;
                Ok(Schema::AnyOf(set![
                    self.get_comparison_schema(&[arg_schemas[0].clone(), arg_schemas[1].clone()])?,
                    self.get_comparison_schema(&[arg_schemas[0].clone(), arg_schemas[2].clone()])?,
                ]))
            }
            // Boolean operators.
            Not => self.propagate_fixed_null_arguments(
                arg_schemas,
                &[BOOLEAN_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::Boolean),
            ),
            And | Or => self.propagate_variadic_null_arguments(
                arg_schemas,
                BOOLEAN_OR_NULLISH.clone(),
                Schema::Atomic(Atomic::Boolean),
            ),
            // Computed Field Access operator when the field is not known until runtime.
            ComputedFieldAccess => {
                self.schema_check_fixed_args(
                    arg_schemas,
                    &[ANY_DOCUMENT.clone(), Schema::Atomic(Atomic::String)],
                )?;
                Ok(Schema::Any)
            }
            // Conditional scalar functions.
            NullIf => {
                self.get_comparison_schema(arg_schemas)?;
                Ok(Schema::AnyOf(set![
                    arg_schemas[0].clone().upconvert_missing_to_null(),
                    Schema::Atomic(Atomic::Null),
                ]))
            }
            Coalesce => self.get_coalesce_schema(arg_schemas),
            // Array scalar functions.
            Slice => self.get_slice_schema(arg_schemas),
            Size => self.propagate_fixed_null_arguments(
                arg_schemas,
                &[ANY_ARRAY_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::Integer),
            ),
            // Numeric value scalar functions.
            Position => self.propagate_fixed_null_arguments(
                arg_schemas,
                &[STRING_OR_NULLISH.clone(), STRING_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::Integer),
            ),
            CharLength | OctetLength | BitLength => self.propagate_fixed_null_arguments(
                arg_schemas,
                &[STRING_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::Integer),
            ),
            Year | Month | Day | Hour | Minute | Second => self.propagate_fixed_null_arguments(
                arg_schemas,
                &[DATE_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::Integer),
            ),
            // String value scalar functions.
            Substring => self.get_substring_schema(arg_schemas),
            Upper | Lower => self.propagate_fixed_null_arguments(
                arg_schemas,
                &[STRING_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::String),
            ),
            LTrim | RTrim | BTrim => self.propagate_fixed_null_arguments(
                arg_schemas,
                &[STRING_OR_NULLISH.clone(), STRING_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::String),
            ),
            // Datetime value scalar function.
            CurrentTimestamp => {
                self.schema_check_fixed_args(arg_schemas, &[])?;
                Ok(Schema::Atomic(Atomic::Date))
            }
            MergeObjects => self.schema_check_merge_objects(arg_schemas),
        }
    }

    fn schema_check_merge_objects(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        self.schema_check_variadic_args(arg_schemas, ANY_DOCUMENT.clone())?;
        // union all AnyOf arg_schemas into union Document schemata
        arg_schemas
            .iter()
            .cloned()
            .map(|s| match s {
                Schema::AnyOf(ao) => ao
                    .into_iter()
                    .reduce(Schema::document_union)
                    .unwrap_or_else(|| EMPTY_DOCUMENT.clone()),
                Schema::Document(d) => Schema::Document(d),
                Schema::Any
                | Schema::Unsat
                | Schema::Missing
                | Schema::Atomic(_)
                | Schema::Array(_) => {
                    unreachable!()
                }
            })
            .fold(Ok(EMPTY_DOCUMENT.clone()), |acc, curr| {
                let acc = acc?;
                let curr = curr;
                let sat = acc.has_overlapping_keys_with(&curr);
                if sat != Satisfaction::Not {
                    return Err(Error::CannotMergeObjects(acc, curr, sat));
                }
                if let (Schema::Document(mut d1), Schema::Document(d2)) = (acc, curr) {
                    d1.keys.extend(d2.keys.into_iter());
                    d1.required.extend(d2.required.into_iter());
                    d1.additional_properties |= d2.additional_properties;
                    Ok(Schema::Document(d1))
                } else {
                    unreachable!();
                }
            })
    }

    /// Returns the string and/or null schema for the substring function.
    ///
    /// The error checks include special handling for an optional third argument.
    /// That is, a valid substring function can only be one of:
    ///   - SUBSTRING(<string> FROM <start>)
    ///   - SUBSTRING(<string> FROM <start> FOR <length>)
    ///
    /// We first check the schema for 3 args. If the check fails specifically due
    /// to an incorrect argument count, we check the schema for 2 args instead.
    fn get_substring_schema(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        Ok(self.propagate_null_arguments_helper(
            self.schema_check_fixed_args(
                arg_schemas,
                &[
                    STRING_OR_NULLISH.clone(),
                    INTEGER_OR_NULLISH.clone(),
                    INTEGER_OR_NULLISH.clone(),
                ],
            )
            .or_else(|err| match err {
                Error::IncorrectArgumentCount { .. } => self.schema_check_fixed_args(
                    arg_schemas,
                    &[STRING_OR_NULLISH.clone(), INTEGER_OR_NULLISH.clone()],
                ),
                e => Err(e),
            })?,
            Schema::Atomic(Atomic::String),
        ))
    }

    /// Returns the schema for the `COALESCE()` function, or an error if no arguments are provided.
    /// If there is a certainly non-nullish argument, then the result schema will be the set of
    /// all non-nullish schema possibilities for the arguments up to and including the first
    /// certainly non-nullish argument. Otherwise, all argument schemas (including `NULL`) are possible.
    fn get_coalesce_schema(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        // Coalesce requires at least one argument.
        if arg_schemas.is_empty() {
            return Err(Error::IncorrectArgumentCount {
                name: self.as_str(),
                required: 1,
                found: 0,
            });
        }

        let schema = if let Some(idx) = arg_schemas
            .iter()
            // Search for a certainly non-null argument.
            .position(|s| s.satisfies(&NULLISH.clone()) == Satisfaction::Not)
        {
            // If one exists, only consider schemas up to and including it and remove nullish schemas
            // as a possibility.
            Schema::AnyOf(arg_schemas.iter().cloned().take(idx + 1).collect()).subtract_nullish()
        } else {
            Schema::AnyOf(arg_schemas.iter().cloned().collect())
        };
        let schema = Schema::simplify(&schema.upconvert_missing_to_null());
        Ok(schema)
    }

    /// Returns the array and/or null schema for the slice function.
    ///
    /// The error checks include special handling for an optional third argument.
    /// That is, a valid slice function can only be one of:
    ///   - SLICE(<array>, <length>)
    ///   - SLICE(<array>, <start>, <length>)
    ///
    /// We first check the schema for 3 args. If the check fails specifically due
    /// to an incorrect argument count, we check the schema for 2 args instead.
    fn get_slice_schema(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        Ok(self.propagate_null_arguments_helper(
            self.schema_check_fixed_args(
                arg_schemas,
                &[
                    ANY_ARRAY.clone(),
                    INTEGER_OR_NULLISH.clone(),
                    INTEGER_OR_NULLISH.clone(),
                ],
            )
            .or_else(|err| match err {
                Error::IncorrectArgumentCount { .. } => self.schema_check_fixed_args(
                    arg_schemas,
                    &[ANY_ARRAY.clone(), INTEGER_OR_NULLISH.clone()],
                ),
                e => Err(e),
            })?,
            ANY_ARRAY.clone(),
        ))
    }
}

impl CastExpr {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        // The schemas of the original expression and the type being casted to.
        let expr_schema = self.expr.schema(state)?;
        let type_schema = Schema::from(self.to);

        // The schemas of the `on_null` and `on_error` fields as set during algebrization.
        let on_null_schema = self.on_null.schema(state)?;
        let on_error_schema = self.on_error.schema(state)?;

        // If the original expression is definitely null or missing, return the `on_null` schema.
        if expr_schema.satisfies(&Schema::AnyOf(set![
            Schema::Atomic(Atomic::Null),
            Schema::Missing,
        ])) == Satisfaction::Must
        {
            return Ok(on_null_schema);
        }

        // If the original expression is being cast to its own type, return that type.
        if expr_schema.satisfies(&type_schema) == Satisfaction::Must {
            return Ok(type_schema);
        }

        Ok(Schema::AnyOf(set![
            type_schema,
            on_null_schema,
            on_error_schema,
        ]))
    }
}

/// A trait with the shared schema checking behavior for SearchedCaseExpr and SimpleCaseExpr.
/// Both case expressions specify their check on the WHEN branch by implementing `schema()`.
/// The trait's `schema_aux()` function then performs that check and returns the resulting schema.
trait SchemaCheckCaseExpr {
    fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error>;
    fn schema_aux(
        state: &SchemaInferenceState,
        when_branches: &[WhenBranch],
        else_branch: &Expression,
        check_when_schema: &dyn Fn(&Schema) -> Result<(), Error>,
    ) -> Result<Schema, Error> {
        // Check each WHEN condition/operand and collect each THEN result in the output.
        let mut schemas = when_branches
            .iter()
            .map(|wb| {
                Ok({
                    check_when_schema(&wb.when.schema(state)?)?;
                    wb.then.schema(state)?
                })
            })
            .collect::<Result<BTreeSet<_>, _>>()?;

        // The resulting schema for a case expression is AnyOf the THEN results
        // from each when_branch, along with the ELSE branch result.
        schemas.insert(else_branch.schema(state)?);
        Ok(Schema::AnyOf(schemas))
    }
}

impl SchemaCheckCaseExpr for SearchedCaseExpr {
    fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        // All WHEN conditions for a SearchedCaseExpr must be boolean or nullish.
        let check_when_schema = &|when_schema: &Schema| {
            if when_schema.satisfies(&BOOLEAN_OR_NULLISH) != Satisfaction::Must {
                return Err(Error::SchemaChecking {
                    name: "SearchedCase",
                    required: BOOLEAN_OR_NULLISH.clone(),
                    found: when_schema.clone(),
                });
            }
            Ok(())
        };

        Self::schema_aux(
            state,
            &self.when_branch,
            &self.else_branch,
            check_when_schema,
        )
    }
}

impl SchemaCheckCaseExpr for SimpleCaseExpr {
    fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        // The case operand for a SimpleCaseExpr must be comparable with all WHEN operands.
        let case_operand_schema = self.expr.schema(state)?;
        let check_when_schema = &|when_schema: &Schema| {
            if case_operand_schema.is_comparable_with(when_schema) != Satisfaction::Must {
                return Err(Error::InvalidComparison(
                    "SimpleCase",
                    case_operand_schema.clone(),
                    when_schema.clone(),
                ));
            }
            Ok(())
        };

        Self::schema_aux(
            state,
            &self.when_branch,
            &self.else_branch,
            check_when_schema,
        )
    }
}

impl TypeAssertionExpr {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        // The schemas of the original expression and the type being asserted to.
        let expr_schema = self.expr.schema(state)?;
        let target_schema = Schema::from(self.target_type);

        // Return an error if the target type is not one of the
        // original expression's possible types.
        if expr_schema.satisfies(&target_schema) == Satisfaction::Not {
            return Err(Error::SchemaChecking {
                name: "::!",
                required: target_schema,
                found: expr_schema,
            });
        }

        Ok(target_schema)
    }
}
