use crate::{
    catalog::*,
    map,
    mir::{
        binding_tuple,
        schema::util::{lift_array_schemas, merge_bot_any_of_document_schemas, set_field_schema},
        *,
    },
    schema::{
        Atomic, Document, ResultSet, Satisfaction, Schema, SchemaEnvironment, ANY_ARRAY,
        ANY_ARRAY_OR_NULLISH, ANY_DOCUMENT, BOOLEAN_OR_NULLISH, DATE_OR_NULLISH, EMPTY_DOCUMENT,
        INTEGER_LONG_OR_NULLISH, INTEGER_OR_NULLISH, NULLISH, NUMERIC, NUMERIC_OR_NULLISH,
        STRING_OR_NULLISH,
    },
    set,
    util::unique_linked_hash_map::UniqueLinkedHashMap,
};
use bson::spec::BinarySubtype;
use std::{
    cell::RefCell,
    cmp::min,
    collections::{BTreeMap, BTreeSet},
    hash::{Hash, Hasher},
};

mod errors;
pub use errors::Error;
#[cfg(test)]
pub(crate) use errors::ANY_SCHEMA_ADDENDUM;
mod util;

#[cfg(test)]
mod test;

#[derive(PartialEq, Eq, Hash, Clone, Debug)]
struct SchemaCacheContents<T: Clone> {
    result: T,
}

impl<T: Clone> SchemaCacheContents<T> {
    fn new(result: T) -> SchemaCacheContents<T> {
        SchemaCacheContents { result }
    }
}

/// SchemaCache caches a `Result` because `schema()` methods are fallible and always return a Result.
#[derive(Clone, Eq)]
pub struct SchemaCache<T: Clone>(RefCell<Option<SchemaCacheContents<Result<T, Error>>>>);

// impl Debug for SchemaCache<T>

impl<T: Clone> std::fmt::Debug for SchemaCache<T> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        // We don't want to print the contents of the cache, so we just print the type.
        write!(f, "<SchemaCache...>")
    }
}

impl<T: Clone> PartialEq for SchemaCache<T> {
    fn eq(&self, _other: &Self) -> bool {
        // Always return true since stages and expressions derive PartialEq, and we shouldn't care
        // about the contents of the SchemaCache of the stage or expression when comparing structs
        // which contain caches.
        true
    }
}

impl<T: Clone> Hash for SchemaCache<T> {
    fn hash<H: Hasher>(&self, state: &mut H) {
        0.hash(state);
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
    type ReturnType: Clone + std::fmt::Debug;

    /// Get the stored RefCell from the underlying struct this trait is implemented on if it exists.
    /// Some types don't benefit from caching (literal and reference expressions specifically),
    /// so this function returns an Option that may or may not contain a cache.
    fn get_cache(&self) -> &SchemaCache<Self::ReturnType>;

    /// Get the cached result if the cache exists and it contains a value.
    #[allow(unused)]
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
        self.check_schema(state)
    }
}

#[derive(Debug, Copy, Clone, Default)]
pub enum SchemaCheckingMode {
    // In strict mode, schema checking will fail unless all type
    // constraints are satisfied.
    Strict,
    // In relaxed mode, schema checking will pass unless a type constraint
    // is violated.
    #[default]
    Relaxed,
}

#[derive(Debug, Clone)]
pub struct SchemaInferenceState<'a> {
    pub scope_level: u16,
    pub env: SchemaEnvironment,
    pub catalog: &'a Catalog,
    pub schema_checking_mode: SchemaCheckingMode,
}

impl<'a> SchemaInferenceState<'a> {
    pub fn empty(catalog: &'a Catalog) -> Self {
        Self::new(
            0u16,
            SchemaEnvironment::default(),
            catalog,
            SchemaCheckingMode::default(),
        )
    }

    pub fn new(
        scope_level: u16,
        env: SchemaEnvironment,
        catalog: &'a Catalog,
        schema_checking_mode: SchemaCheckingMode,
    ) -> SchemaInferenceState<'a> {
        SchemaInferenceState {
            scope_level,
            env,
            catalog,
            schema_checking_mode,
        }
    }

    pub fn with_merged_schema_env(&self, env: SchemaEnvironment) -> SchemaInferenceState {
        SchemaInferenceState {
            env: env.with_merged_mappings(self.env.clone()),
            catalog: self.catalog,
            scope_level: self.scope_level,
            schema_checking_mode: self.schema_checking_mode,
        }
    }

    pub fn subquery_state(&self) -> SchemaInferenceState {
        SchemaInferenceState {
            scope_level: self.scope_level + 1,
            env: self.env.clone(),
            catalog: self.catalog,
            schema_checking_mode: self.schema_checking_mode,
        }
    }

    /// check_satisfies computes the satisfaction level of [`schema1`] -> [`schema2`],
    /// and checks if it's in line with the schema checking mode.
    pub fn check_satisfies(&self, schema1: &Schema, schema2: &Schema) -> bool {
        self.check_satisfaction(schema1.satisfies(schema2))
    }

    /// check_comparable_with computes the satisfaction level
    /// for comparing [`schema1`] to [`schema2`], and checks if it's in line
    /// with the schema checking mode.
    pub fn check_comparable_with(&self, schema1: &Schema, schema2: &Schema) -> bool {
        self.check_satisfaction(schema1.is_comparable_with(schema2))
    }

    /// check_self_comparable computes the satisfaction level
    /// for comparing [`schema`] to itself, and checks if it's in line
    /// with the schema checking mode.
    pub fn check_self_comparable(&self, schema: &Schema) -> bool {
        self.check_satisfaction(schema.is_self_comparable())
    }

    pub fn check_satisfaction(&self, satisfaction: Satisfaction) -> bool {
        matches!(
            (self.schema_checking_mode, satisfaction),
            (SchemaCheckingMode::Relaxed, Satisfaction::May)
                | (SchemaCheckingMode::Relaxed, Satisfaction::Must)
                | (SchemaCheckingMode::Strict, Satisfaction::Must)
        )
    }
}

impl CachedSchema for Stage {
    type ReturnType = ResultSet;

    /// Cached call to `check_schema()`.
    fn schema(&self, state: &SchemaInferenceState) -> Result<Self::ReturnType, Error> {
        // This mutable borrow of the cache will be released when schema() is complete.
        let mut cache = self.get_cache().0.borrow_mut();
        match &*cache {
            Some(contents) => contents.result.clone(),
            _ => {
                let schema_result = self.check_schema(state);
                cache.replace(SchemaCacheContents::new(schema_result.clone()));
                schema_result
            }
        }
    }

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
            Stage::Derived(s) => &s.cache,
            Stage::Unwind(s) => &s.cache,
            Stage::MQLIntrinsic(MQLStage::EquiJoin(s)) => &s.cache,
            Stage::MQLIntrinsic(MQLStage::LateralJoin(s)) => &s.cache,
            Stage::MQLIntrinsic(MQLStage::MatchFilter(s)) => &s.cache,
            Stage::Sentinel => unreachable!(),
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
                let state = state.with_merged_schema_env(source_result_set.schema_env.clone());
                let cond_schema = f.condition.schema(&state)?;
                if !state.check_satisfies(&cond_schema, &BOOLEAN_OR_NULLISH) {
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
                let state = state.with_merged_schema_env(source_result_set.schema_env);
                let schema_env = p
                    .expression
                    .iter()
                    .map(|(k, e)| match e.schema(&state) {
                        Ok(s) => {
                            if !state.check_satisfies(&s, &ANY_DOCUMENT) {
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
                let schema_env = if p.is_add_fields {
                    state.env.union(schema_env)
                } else {
                    schema_env
                };
                // We want AddFields keys under bottom to be added to the incoming bottom schema.
                let schema_env = merge_bot_any_of_document_schemas(state.scope_level, schema_env);
                Ok(ResultSet {
                    schema_env: schema_env.simplify(),
                    min_size,
                    max_size,
                })
            }
            #[allow(clippy::manual_try_fold)]
            Stage::Group(g) => {
                let source_result_set = g.source.schema(state)?;
                let state = state.with_merged_schema_env(source_result_set.schema_env.clone());

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
                        ..Default::default()
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

                        // If schema checking is in strict mode, the group key must have a
                        // schema that is self-comparable, otherwise it must or may have one.
                        if !state.check_self_comparable(&group_key_schema) {
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
                                     _ => panic!("group key at position {0} is an unaliased field access with no datasource reference", index),
                                },
                                _ => panic!("group key at position {0} is an unaliased non-field access expression", index),
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

                let schema_env = g
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

                // We want group keys and aggregations under bottom to be combined under a single
                // document schema.
                let schema_env = merge_bot_any_of_document_schemas(state.scope_level, schema_env);

                Ok(ResultSet {
                    schema_env: schema_env.simplify(),
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

                // Note that the parser parses the OFFSET value as a u32, which
                // we convert to an i64 during algebrization. This means we can
                // be sure the o.offset i64 value is actually always positive.
                // Therefore, converting to u64 for result set size arithmetic
                // is safe to do.
                Ok(ResultSet {
                    schema_env: source_result_set.schema_env,
                    min_size: source_result_set.min_size.saturating_sub(o.offset as u64),
                    max_size: source_result_set
                        .max_size
                        .map(|x| x.saturating_sub(o.offset as u64)),
                })
            }
            Stage::Sort(s) => {
                let source_result_set = s.source.schema(state)?;
                let state = state.with_merged_schema_env(source_result_set.schema_env.clone());

                // Ensure that each sort key is statically comparable to itself.
                s.specs
                    .clone()
                    .iter()
                    .enumerate()
                    .try_for_each(|(index, spec)| match spec {
                        SortSpecification::Asc(a) | SortSpecification::Desc(a) => {
                            let schema = a.schema(&state)?;
                            if !state.check_self_comparable(&schema) {
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
                    None => {
                        return Err(Error::CollectionNotFound(
                            c.db.clone(),
                            c.collection.clone(),
                        ))
                    }
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
                if state.check_satisfies(&array_items_schema, &ANY_DOCUMENT) {
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

                let state = state.with_merged_schema_env(left_result_set.schema_env.clone());
                let state = state.with_merged_schema_env(right_result_set.schema_env.clone());
                if let Some(e) = &j.condition {
                    let cond_schema = e.schema(&state)?;
                    if !state.check_satisfies(&cond_schema, &BOOLEAN_OR_NULLISH) {
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
                            schema_env: left_result_set.schema_env.with_merged_mappings(
                                right_result_set
                                    .schema_env
                                    .into_iter()
                                    .map(|(key, schema)| {
                                        (key, Schema::AnyOf(set![Schema::Missing, schema]))
                                    })
                                    .collect::<SchemaEnvironment>(),
                            ),
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
                                .with_merged_mappings(right_result_set.schema_env),
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
                Ok(ResultSet {
                    schema_env: left_result_set
                        .schema_env
                        .union(right_result_set.schema_env),
                    min_size: left_result_set.min_size + right_result_set.min_size,
                    max_size,
                })
            }
            Stage::Derived(s) => s.source.schema(&SchemaInferenceState {
                scope_level: state.scope_level + 1,
                env: state.env.clone(),
                catalog: state.catalog,
                schema_checking_mode: state.schema_checking_mode,
            }),
            Stage::Unwind(u) => {
                let source_result_set = u.source.schema(state)?;
                let mut state = SchemaInferenceState::new(
                    state.scope_level,
                    source_result_set.schema_env.clone(),
                    state.catalog,
                    state.schema_checking_mode,
                );

                let path_datasource = u.path.key.clone();
                let path_schema = u.path.schema(&state)?;
                let is_path_not_array = path_schema.satisfies(&ANY_ARRAY) == Satisfaction::Not;

                // If the Unwind specifies an INDEX name, ensure it does not
                // conflict with an existing field in the schema. If it does
                // not conflict, then include the INDEX field in the result
                // schema.
                if let Some(index) = &u.index {
                    let sat = state.env.get(&path_datasource).map_or(
                        Err(Error::DatasourceNotFoundInSchemaEnv(
                            path_datasource.clone(),
                        )),
                        |k| Ok(k.contains_field(index.as_str())),
                    )?;

                    if matches!(
                        (sat, state.schema_checking_mode),
                        (Satisfaction::Must, _) | (Satisfaction::May, SchemaCheckingMode::Strict)
                    ) {
                        return Err(Error::UnwindIndexNameConflict(index.clone()));
                    }

                    let path_is_exactly_nullish_array =
                        path_schema.satisfies(&ANY_ARRAY_OR_NULLISH);

                    // The schema of INDEX depends on two factors: the schema of the PATH argument
                    // and the value of OUTER.
                    //
                    // If the schema of the PATH argument indicates it is never an array, the INDEX
                    // schema is exactly NULL. This is because there will never be a value of PATH
                    // that is an array, so there will never be data to populate the INDEX field.
                    //
                    // After that, we know the PATH schema MAY or MUST be an array. If it MUST be
                    // an array (or null or missing) and if OUTER is set to false, the INDEX schema
                    // is exactly LONG. This is because the value of PATH will always be a non-empty
                    // array, so there will always be data to populate the INDEX field.
                    //
                    // In all other cases, the INDEX schema is either NULL or LONG. This is because
                    // the values of PATH may be non-arrays (including null or missing) or may be
                    // empty arrays. For any such values, there is no corresponding INDEX data.
                    let index_result_schema = if is_path_not_array {
                        Schema::Atomic(Atomic::Null)
                    } else if path_is_exactly_nullish_array == Satisfaction::Must && !u.outer {
                        Schema::Atomic(Atomic::Long)
                    } else {
                        Schema::AnyOf(set![
                            Schema::Atomic(Atomic::Long),
                            Schema::Atomic(Atomic::Null)
                        ])
                    };

                    match state.env.get(&path_datasource) {
                        None => return Err(Error::DatasourceNotFoundInSchemaEnv(path_datasource)),
                        Some(path_datasource_schema) => {
                            let updated_path_datasource_schema = set_field_schema(
                                path_datasource_schema.clone(),
                                &mut vec![index.clone()],
                                index_result_schema,
                                true, // the index field is always required
                            );
                            state
                                .env
                                .insert(path_datasource.clone(), updated_path_datasource_schema);
                        }
                    }
                };

                // The result schema of PATH depends on two factors: the schema of the PATH argument
                // and the value of OUTER.
                //
                // If the schema of the PATH argument indicates it is never an array, and if OUTER
                // is true, then the result PATH schema is the same as the input PATH schema. This
                // is because the UNWIND operation will not alter any input documents since
                // non-arrays are not altered and null and missing values are retained.
                //
                // If the input PATH schema MAY or MUST be an array, and if OUTER is true, then the
                // result PATH schema consists of any inner-array schemas or other non-array schemas
                // from the input PATH schema, and may be missing. For example, consider the schema
                // AnyOf(Array(Int), Array(String), Null) as the input PATH schema. The result PATH
                // schema would be AnyOf(Int, String, Null, Missing).
                //
                // In all other cases, the result PATH schema consists of any inner-array schemas or
                // other non-array schemas from the input PATH schema, and is not nullish. Here, the
                // "inner-array" and "other non-array" schemas are the same as the ones described
                // above. The reason the result is not nullish is because OUTER is false, so empty
                // arrays, NULL, and MISSING values are ignored.
                let (path_result_schema, path_required) = if is_path_not_array && u.outer {
                    (
                        path_schema.clone(),
                        path_schema.satisfies(&Schema::Missing) == Satisfaction::Not,
                    )
                } else {
                    let temp_path_schema = lift_array_schemas(path_schema, u.outer);
                    (
                        // We call simplify since the lift() and add_missing() functions
                        // may create schemas with duplicate information in an AnyOf, or
                        // an AnyOf with just one member.
                        Schema::simplify(
                            &(if !is_path_not_array && u.outer {
                                match temp_path_schema {
                                    Schema::AnyOf(mut ao) => {
                                        ao.insert(Schema::Missing);
                                        Schema::AnyOf(ao)
                                    }
                                    _ => Schema::AnyOf(set![temp_path_schema, Schema::Missing]),
                                }
                            } else {
                                temp_path_schema
                            }),
                        ),
                        !u.outer,
                    )
                };

                match state.env.get(&(path_datasource.clone())) {
                    None => return Err(Error::DatasourceNotFoundInSchemaEnv(path_datasource)),
                    Some(path_datasource_schema) => {
                        // the set_field_schema method expects paths to be in reverse order
                        // of access. Rather than clone and then reverse in place, we just reverse
                        // as we clone to avoid one iterative pass.
                        let mut path_fields = u.path.fields.iter().cloned().rev().collect();
                        let updated_path_datasource_schema = set_field_schema(
                            path_datasource_schema.clone(),
                            &mut path_fields,
                            path_result_schema,
                            path_required,
                        );
                        state
                            .env
                            .insert(path_datasource, updated_path_datasource_schema);
                    }
                }

                Ok(ResultSet {
                    schema_env: state.env,
                    min_size: source_result_set.min_size,
                    max_size: source_result_set.max_size,
                })
            }
            Stage::MQLIntrinsic(MQLStage::EquiJoin(j)) => {
                let source_result_set = j.source.schema(state)?;
                let from_result_set = j.from.schema(state)?;

                let state = state.with_merged_schema_env(source_result_set.schema_env.clone());
                let state = state.with_merged_schema_env(from_result_set.schema_env.clone());

                let local_field_schema = j.local_field.schema(&state)?;
                let foreign_field_schema = j.foreign_field.schema(&state)?;

                if !state.check_comparable_with(&local_field_schema, &foreign_field_schema) {
                    return Err(Error::InvalidComparison(
                        "equijoin comparison",
                        local_field_schema,
                        foreign_field_schema,
                    ));
                }

                match j.join_type {
                    JoinType::Left => {
                        let min_size = source_result_set.min_size;
                        let max_size = source_result_set
                            .max_size
                            .and_then(|l| from_result_set.max_size.map(|r| l * r));
                        Ok(ResultSet {
                            schema_env: source_result_set.schema_env.with_merged_mappings(
                                from_result_set
                                    .schema_env
                                    .into_iter()
                                    .map(|(key, schema)| {
                                        (key, Schema::AnyOf(set![Schema::Missing, schema]))
                                    })
                                    .collect::<SchemaEnvironment>(),
                            ),
                            min_size,
                            max_size,
                        })
                    }
                    JoinType::Inner => {
                        // A EquiJoin Inner Join can potentially return 0 results.
                        let min_size = 0;
                        let max_size = source_result_set
                            .max_size
                            .and_then(|l| from_result_set.max_size.map(|r| l * r));
                        Ok(ResultSet {
                            schema_env: source_result_set
                                .schema_env
                                .with_merged_mappings(from_result_set.schema_env),
                            min_size,
                            max_size,
                        })
                    }
                }
            }
            Stage::MQLIntrinsic(MQLStage::LateralJoin(j)) => {
                let source_result_set = j.source.schema(state)?;
                let state = state.with_merged_schema_env(source_result_set.schema_env.clone());
                let subquery_result_set = j.subquery.schema(&state)?;

                match j.join_type {
                    JoinType::Left => {
                        let min_size = source_result_set.min_size;
                        let max_size = source_result_set
                            .max_size
                            .and_then(|l| subquery_result_set.max_size.map(|r| l * r));
                        Ok(ResultSet {
                            schema_env: source_result_set.schema_env.with_merged_mappings(
                                subquery_result_set
                                    .schema_env
                                    .into_iter()
                                    .map(|(key, schema)| {
                                        (key, Schema::AnyOf(set![Schema::Missing, schema]))
                                    })
                                    .collect::<SchemaEnvironment>(),
                            ),
                            min_size,
                            max_size,
                        })
                    }
                    JoinType::Inner => {
                        // A EquiJoin Inner Join can potentially return 0 results.
                        let min_size = 0;
                        let max_size = source_result_set
                            .max_size
                            .and_then(|l| subquery_result_set.max_size.map(|r| l * r));
                        Ok(ResultSet {
                            schema_env: source_result_set
                                .schema_env
                                .with_merged_mappings(subquery_result_set.schema_env),
                            min_size,
                            max_size,
                        })
                    }
                }
            }
            Stage::MQLIntrinsic(MQLStage::MatchFilter(f)) => {
                let source_result_set = f.source.schema(state)?;
                let state = state.with_merged_schema_env(source_result_set.schema_env.clone());
                let cond_schema = f.condition.schema(&state)?;
                if !state.check_satisfies(&cond_schema, &BOOLEAN_OR_NULLISH) {
                    return Err(Error::SchemaChecking {
                        name: "match filter condition",
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
            Stage::Sentinel => unreachable!(),
        }
    }
}

impl AggregationExpr {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        match self {
            AggregationExpr::CountStar(_) => Ok(Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long)
            ])),
            AggregationExpr::Function(a) => a.schema(state),
        }
    }
}

impl AggregationFunctionApplication {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let arg_schema = self.arg.schema(state)?;
        if self.distinct && !state.check_self_comparable(&arg_schema) {
            return Err(Error::AggregationArgumentMustBeSelfComparable(
                format!("{} DISTINCT", self.function.as_str()),
                arg_schema,
            ));
        }
        self.function.schema(state, arg_schema)
    }
}

impl AggregationFunction {
    pub fn schema(
        &self,
        state: &SchemaInferenceState,
        arg_schema: Schema,
    ) -> Result<Schema, Error> {
        use crate::mir::AggregationFunction::*;
        use Satisfaction::*;
        Ok(match self {
            AddToArray => Schema::Array(Box::new(arg_schema)),
            Avg | StddevPop | StddevSamp => {
                self.schema_check_fixed_args(
                    state,
                    &[arg_schema.clone()],
                    &[NUMERIC_OR_NULLISH.clone()],
                )?;
                // we cannot use get_arithmetic_schema for Avg, StddevPop, StddevSamp
                // because they never return Long or Integer results, even for Long
                // or Integer inputs.
                let get_numeric_schema = || match (
                    arg_schema.satisfies(&Schema::Atomic(Atomic::Decimal)),
                    arg_schema.satisfies(&Schema::AnyOf(set![
                        Schema::Atomic(Atomic::Integer),
                        Schema::Atomic(Atomic::Double),
                        Schema::Atomic(Atomic::Long),
                    ])),
                ) {
                    (Must, _) => Schema::Atomic(Atomic::Decimal),
                    (May, Not) => Schema::Atomic(Atomic::Decimal),
                    (May, May) => Schema::AnyOf(set![
                        Schema::Atomic(Atomic::Double),
                        Schema::Atomic(Atomic::Decimal),
                    ]),
                    _ => Schema::Atomic(Atomic::Double),
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
                if !state.check_self_comparable(&arg_schema) {
                    return Err(Error::AggregationArgumentMustBeSelfComparable(
                        self.as_str().to_string(),
                        arg_schema,
                    ));
                }
                arg_schema
            }
            MergeDocuments => {
                self.schema_check_fixed_args(
                    state,
                    &[arg_schema.clone()],
                    &[ANY_DOCUMENT.clone()],
                )?;
                arg_schema
            }
            Sum => self.get_arithmetic_schema(state, &[arg_schema])?,
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
            schema_checking_mode: state.schema_checking_mode,
        })?;
        Ok((schema, min_size, max_size))
    }

    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let (schema, min_size, max_size) = self.schema_helper(state)?;

        // The subquery's result set MUST have a cardinality of 0 or 1. The returned schema
        // MUST include MISSING if the cardinality of the result set MAY be 0. We can exclude
        // MISSING from the returned schema if the cardinality of the result set MUST be 1.
        if max_size.is_none() || max_size.unwrap() > 1 {
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
        self.get_comparison_schema(state, &[argument_schema, subquery_schema])
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

    fn array_schema(state: &SchemaInferenceState, a: &[Expression]) -> Result<Schema, Error> {
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
        state: &SchemaInferenceState,
        d: &UniqueLinkedHashMap<String, Expression>,
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
            ..Default::default()
        }))
    }
}

impl Expression {
    /// Recursively schema checks this expression, its arguments, and
    /// all contained expressions/stages. If schema checking succeeds,
    /// returns a [`Schema`] describing this expression's schema. The
    /// provided [`SchemaInferenceState`] should include schema
    /// information for all datasources in scope.
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        match self {
            Expression::Literal(lit) => lit.schema(state),
            Expression::Reference(ReferenceExpr { key, .. }) => state
                .env
                .get(key)
                .cloned()
                .ok_or_else(|| Error::DatasourceNotFoundInSchemaEnv(key.clone())),
            Expression::Array(ArrayExpr { array, .. }) => Expression::array_schema(state, array),
            Expression::Document(DocumentExpr { document, .. }) => {
                Expression::document_schema(state, document)
            }
            Expression::FieldAccess(fa) => fa.schema(state),
            Expression::DateFunction(d) => d.schema(state),
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
                if !state.check_satisfies(&expr_schema, &STRING_OR_NULLISH) {
                    return Err(Error::SchemaChecking {
                        name: "Like",
                        required: STRING_OR_NULLISH.clone(),
                        found: expr_schema,
                    });
                }
                let pattern_schema = l.pattern.schema(state)?;
                if !state.check_satisfies(&pattern_schema, &STRING_OR_NULLISH) {
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
            Expression::MQLIntrinsicFieldExistence(_) => Ok(Schema::Atomic(Atomic::Boolean)),
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
            RegularExpression(_) => Schema::Atomic(Atomic::Regex),
            JavaScriptCode(_) => Schema::Atomic(Atomic::Javascript),
            JavaScriptCodeWithScope(_) => Schema::Atomic(Atomic::JavascriptWithScope),
            Timestamp(_) => Schema::Atomic(Atomic::Timestamp),
            Binary(b) => match b.subtype {
                BinarySubtype::UuidOld => return Err(Error::InvalidBinaryDataType),
                _ => Schema::Atomic(Atomic::BinData),
            },
            ObjectId(_) => Schema::Atomic(Atomic::ObjectId),
            DateTime(_) => Schema::Atomic(Atomic::Date),
            Symbol(_) => Schema::Atomic(Atomic::Symbol),
            MaxKey => Schema::Atomic(Atomic::MaxKey),
            MinKey => Schema::Atomic(Atomic::MinKey),
            DbPointer(_) => Schema::Atomic(Atomic::DbPointer),
            Undefined => Schema::Atomic(Atomic::Null),
            Decimal128(_) => Schema::Atomic(Atomic::Decimal),
        })
    }
}

impl FieldAccess {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let accessee_schema = self.expr.schema(state)?;
        if accessee_schema.satisfies(&ANY_DOCUMENT) == Satisfaction::Not {
            return Err(Error::SchemaChecking {
                name: "FieldAccess",
                required: ANY_DOCUMENT.clone(),
                found: accessee_schema,
            });
        }
        if accessee_schema.contains_field(&self.field) == Satisfaction::Not {
            let found_keys = accessee_schema.keys();
            if found_keys.is_empty() {
                return Err(Error::AccessMissingField(self.field.clone(), None));
            } else {
                return Err(Error::AccessMissingField(
                    self.field.clone(),
                    Some(found_keys),
                ));
            }
        }
        Ok(Expression::get_field_schema(&accessee_schema, &self.field))
    }
}

impl FieldPath {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let key_schema = state
            .env
            .get(&self.key)
            .cloned()
            .ok_or_else(|| Error::DatasourceNotFoundInSchemaEnv(self.key.clone()))?;

        let mut cur_schema = key_schema;

        // Note this will not check anything, if fields is empty, except that the datasource exists.
        // Our code should never generate empty fields, and even if it did, this is technically correct.
        // There is no reason the datasource has to be a Document if no fields are accessed.
        for field in self.fields.iter() {
            if cur_schema.satisfies(&ANY_DOCUMENT) == Satisfaction::Not {
                return Err(Error::SchemaChecking {
                    name: "FieldPath",
                    required: ANY_DOCUMENT.clone(),
                    found: cur_schema,
                });
            }
            if cur_schema.contains_field(field) == Satisfaction::Not {
                return Err(Error::AccessMissingField(
                    field.clone(),
                    Some(cur_schema.keys()),
                ));
            }
            cur_schema = Expression::get_field_schema(&cur_schema, field);
        }
        Ok(cur_schema)
    }
}

impl DateFunctionApplication {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let args = self
            .args
            .iter()
            .map(|x| x.schema(state))
            .collect::<Result<Vec<_>, _>>()?;
        self.function.schema(state, &args)
    }
}

impl ScalarFunctionApplication {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let args = self
            .args
            .iter()
            .map(|x| x.schema(state))
            .collect::<Result<Vec<_>, _>>()?;
        self.function.schema(state, &args)
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
    fn get_arithmetic_schema(
        &self,
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
    ) -> Result<Schema, Error> {
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
                            .filter(|sch| sch.satisfies(&NUMERIC.clone()) == Satisfaction::Must)
                            .collect::<BTreeSet<_>>(),
                    ),
                    Atomic(a) if a.is_numeric() => s.clone(),
                    Any => NUMERIC.clone(),
                    _ => unreachable!(),
                })
                .collect::<BTreeSet<_>>();
            retain(&schemas_collected)
        }

        match self.schema_check_variadic_args(state, arg_schemas, NUMERIC_OR_NULLISH.clone())? {
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
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
        required_schemas: &[Schema],
    ) -> Result<Satisfaction, Error> {
        self.ensure_arg_count(arg_schemas.len(), required_schemas.len())?;
        let mut total_null_sat = Satisfaction::Not;
        for (i, arg) in arg_schemas.iter().enumerate() {
            if !state.check_satisfies(arg, &required_schemas[i]) {
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
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
        required_schema: Schema,
    ) -> Result<Satisfaction, Error> {
        self.schema_check_fixed_args(
            state,
            arg_schemas,
            &std::iter::repeat_n(required_schema, arg_schemas.len()).collect::<Vec<Schema>>(),
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
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
        required_schemas: &[Schema],
        base_return_schema: Schema,
    ) -> Result<Schema, Error> {
        Ok(self.propagate_null_arguments_helper(
            self.schema_check_fixed_args(state, arg_schemas, required_schemas)?,
            base_return_schema,
        ))
    }

    /// Many scalar functions will return `NULL` if one of their arguments is `NULL`.
    /// This function factors out the schema checking for these scalar functions for variadic
    /// functions.
    fn propagate_variadic_null_arguments(
        &self,
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
        required_schema: Schema,
        base_return_schema: Schema,
    ) -> Result<Schema, Error> {
        Ok(self.propagate_null_arguments_helper(
            self.schema_check_variadic_args(state, arg_schemas, required_schema)?,
            base_return_schema,
        ))
    }

    /// Returns the boolean and/or null schema for a valid comparison, or an error
    /// if the number of operands is not two or the comparison is not valid.
    fn get_comparison_schema(
        &self,
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
    ) -> Result<Schema, Error> {
        let ret_schema = self.propagate_fixed_null_arguments(
            state,
            arg_schemas,
            &[Schema::Any, Schema::Any],
            Schema::Atomic(Atomic::Boolean),
        )?;

        if state.check_comparable_with(&arg_schemas[0], &arg_schemas[1]) {
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

impl SQLFunction for DateFunction {
    fn as_str(&self) -> &'static str {
        self.as_str()
    }
}

impl DateFunction {
    pub fn schema(
        &self,
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
    ) -> Result<Schema, Error> {
        use DateFunction::*;
        match self {
            Add => {
                self.ensure_arg_count(arg_schemas.len(), 2)?;
                self.propagate_fixed_null_arguments(
                    state,
                    arg_schemas,
                    &[INTEGER_LONG_OR_NULLISH.clone(), DATE_OR_NULLISH.clone()],
                    Schema::Atomic(Atomic::Date),
                )
            }
            Diff => {
                self.ensure_arg_count(arg_schemas.len(), 3)?;
                self.propagate_fixed_null_arguments(
                    state,
                    arg_schemas,
                    &[
                        DATE_OR_NULLISH.clone(),
                        DATE_OR_NULLISH.clone(),
                        STRING_OR_NULLISH.clone(),
                    ],
                    Schema::Atomic(Atomic::Integer),
                )
            }
            Trunc => {
                self.ensure_arg_count(arg_schemas.len(), 2)?;
                self.propagate_fixed_null_arguments(
                    state,
                    arg_schemas,
                    &[DATE_OR_NULLISH.clone(), STRING_OR_NULLISH.clone()],
                    Schema::Atomic(Atomic::Date),
                )
            }
        }
    }
}

impl ScalarFunction {
    pub fn schema(
        &self,
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
    ) -> Result<Schema, Error> {
        use ScalarFunction::*;
        match self {
            // String operators.
            Concat => self.propagate_fixed_null_arguments(
                state,
                arg_schemas,
                &[STRING_OR_NULLISH.clone(), STRING_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::String),
            ),
            // Unary arithmetic operators.
            Pos | Neg | Abs | Ceil | Floor => {
                self.ensure_arg_count(arg_schemas.len(), 1)?;
                self.get_arithmetic_schema(state, arg_schemas)
            }
            // Arithmetic operators with variadic arguments.
            Add | Mul => self.get_arithmetic_schema(state, arg_schemas),
            // Arithmetic operators with fixed (two) arguments.
            Sub | Div => {
                self.ensure_arg_count(arg_schemas.len(), 2)?;
                self.get_arithmetic_schema(state, arg_schemas)
            }

            Round => {
                self.ensure_arg_count(arg_schemas.len(), 2)?;
                if arg_schemas[0].satisfies(&NUMERIC_OR_NULLISH) != Satisfaction::Must {
                    Err(Error::SchemaChecking {
                        name: Round.as_str(),
                        required: NUMERIC_OR_NULLISH.clone(),
                        found: arg_schemas[0].clone(),
                    })
                } else if arg_schemas[1].satisfies(&INTEGER_LONG_OR_NULLISH) != Satisfaction::Must {
                    Err(Error::SchemaChecking {
                        name: Round.as_str(),
                        required: INTEGER_LONG_OR_NULLISH.clone(),
                        found: arg_schemas[1].clone(),
                    })
                } else {
                    Ok(arg_schemas[0].clone())
                }
            }

            Cos | Degrees | Radians | Sin | Sqrt | Tan => {
                self.ensure_arg_count(arg_schemas.len(), 1)?;
                self.get_arithmetic_schema(
                    state,
                    // default the schema to Double in case the actual arg schema is INT or LONG
                    &[arg_schemas[0].clone(), Schema::Atomic(Atomic::Double)],
                )
            }

            Log | Mod | Pow => {
                self.ensure_arg_count(arg_schemas.len(), 2)?;
                self.get_arithmetic_schema(
                    state,
                    &[
                        arg_schemas[0].clone(),
                        arg_schemas[1].clone(),
                        // default the schema to Double in case the actual arg schema is INT or LONG
                        Schema::Atomic(Atomic::Double),
                    ],
                )
            }

            // Comparison operators.
            Lt | Lte | Neq | Eq | Gt | Gte => self.get_comparison_schema(state, arg_schemas),
            Between => {
                self.schema_check_fixed_args(
                    state,
                    arg_schemas,
                    &[Schema::Any, Schema::Any, Schema::Any],
                )?;
                Ok(Schema::AnyOf(set![
                    self.get_comparison_schema(
                        state,
                        &[arg_schemas[0].clone(), arg_schemas[1].clone()],
                    )?,
                    self.get_comparison_schema(
                        state,
                        &[arg_schemas[0].clone(), arg_schemas[2].clone()],
                    )?,
                ]))
            }
            // Boolean operators.
            Not => self.propagate_fixed_null_arguments(
                state,
                arg_schemas,
                &[BOOLEAN_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::Boolean),
            ),
            And | Or => self.propagate_variadic_null_arguments(
                state,
                arg_schemas,
                BOOLEAN_OR_NULLISH.clone(),
                Schema::Atomic(Atomic::Boolean),
            ),
            // Computed Field Access operator when the field is not known until runtime.
            ComputedFieldAccess => {
                self.schema_check_fixed_args(
                    state,
                    arg_schemas,
                    &[ANY_DOCUMENT.clone(), Schema::Atomic(Atomic::String)],
                )?;
                Ok(Schema::Any)
            }
            // Conditional scalar functions.
            NullIf => {
                self.get_comparison_schema(state, arg_schemas)?;
                Ok(Schema::AnyOf(set![
                    arg_schemas[0].clone().upconvert_missing_to_null(),
                    Schema::Atomic(Atomic::Null),
                ]))
            }
            Coalesce => self.get_coalesce_schema(arg_schemas),
            // Array scalar functions.
            Slice => self.get_slice_schema(state, arg_schemas),
            Size => self.propagate_fixed_null_arguments(
                state,
                arg_schemas,
                &[ANY_ARRAY_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::Integer),
            ),
            // Numeric value scalar functions.
            Position => self.propagate_fixed_null_arguments(
                state,
                arg_schemas,
                &[STRING_OR_NULLISH.clone(), STRING_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::Integer),
            ),
            CharLength | OctetLength | BitLength => self.propagate_fixed_null_arguments(
                state,
                arg_schemas,
                &[STRING_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::Integer),
            ),
            Year | Month | Day | Hour | Minute | Second | Millisecond | Week | IsoWeek
            | IsoWeekday | DayOfWeek | DayOfYear => self.propagate_fixed_null_arguments(
                state,
                arg_schemas,
                &[DATE_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::Integer),
            ),
            // String value scalar functions.
            Replace => self.propagate_fixed_null_arguments(
                state,
                arg_schemas,
                &[
                    STRING_OR_NULLISH.clone(),
                    STRING_OR_NULLISH.clone(),
                    STRING_OR_NULLISH.clone(),
                ],
                Schema::Atomic(Atomic::String),
            ),
            Substring => self.get_substring_schema(state, arg_schemas),
            Split => self.propagate_fixed_null_arguments(
                state,
                arg_schemas,
                &[
                    STRING_OR_NULLISH.clone(),
                    STRING_OR_NULLISH.clone(),
                    INTEGER_OR_NULLISH.clone(),
                ],
                Schema::AnyOf(set![
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Null),
                ]),
            ),
            Upper | Lower => self.propagate_fixed_null_arguments(
                state,
                arg_schemas,
                &[STRING_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::String),
            ),
            LTrim | RTrim | BTrim => self.propagate_fixed_null_arguments(
                state,
                arg_schemas,
                &[STRING_OR_NULLISH.clone(), STRING_OR_NULLISH.clone()],
                Schema::Atomic(Atomic::String),
            ),
            // Datetime value scalar function.
            CurrentTimestamp => {
                self.schema_check_fixed_args(state, arg_schemas, &[])?;
                Ok(Schema::Atomic(Atomic::Date))
            }
            MergeObjects => self.schema_check_merge_objects(state, arg_schemas),
        }
    }

    #[allow(clippy::manual_try_fold)]
    fn schema_check_merge_objects(
        &self,
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
    ) -> Result<Schema, Error> {
        self.schema_check_variadic_args(state, arg_schemas, ANY_DOCUMENT.clone())?;
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
                let sat = acc.has_overlapping_keys_with(&curr);
                if sat != Satisfaction::Not {
                    return Err(Error::CannotMergeObjects(acc, curr, sat));
                }
                if let (Schema::Document(mut d1), Schema::Document(d2)) = (acc, curr) {
                    d1.keys.extend(d2.keys);
                    d1.required.extend(d2.required);
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
    fn get_substring_schema(
        &self,
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
    ) -> Result<Schema, Error> {
        Ok(self.propagate_null_arguments_helper(
            self.schema_check_fixed_args(
                state,
                arg_schemas,
                &[
                    STRING_OR_NULLISH.clone(),
                    INTEGER_OR_NULLISH.clone(),
                    INTEGER_OR_NULLISH.clone(),
                ],
            )
            .or_else(|err| match err {
                Error::IncorrectArgumentCount { .. } => self.schema_check_fixed_args(
                    state,
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
            Schema::AnyOf(arg_schemas.iter().take(idx + 1).cloned().collect()).subtract_nullish()
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
    fn get_slice_schema(
        &self,
        state: &SchemaInferenceState,
        arg_schemas: &[Schema],
    ) -> Result<Schema, Error> {
        Ok(self.propagate_null_arguments_helper(
            self.schema_check_fixed_args(
                state,
                arg_schemas,
                &[
                    ANY_ARRAY.clone(),
                    INTEGER_OR_NULLISH.clone(),
                    INTEGER_OR_NULLISH.clone(),
                ],
            )
            .or_else(|err| match err {
                Error::IncorrectArgumentCount { .. } => self.schema_check_fixed_args(
                    state,
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
        if state.check_satisfies(&expr_schema, &type_schema) {
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
            if !state.check_satisfies(when_schema, &BOOLEAN_OR_NULLISH) {
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
            if !state.check_comparable_with(&case_operand_schema, when_schema) {
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

impl CachedSchema for MatchQuery {
    type ReturnType = Schema;

    fn get_cache(&self) -> &SchemaCache<Self::ReturnType> {
        match self {
            MatchQuery::Logical(s) => &s.cache,
            MatchQuery::Type(s) => &s.cache,
            MatchQuery::Regex(s) => &s.cache,
            MatchQuery::ElemMatch(s) => &s.cache,
            MatchQuery::Comparison(s) => &s.cache,
            MatchQuery::False(f) => &f.cache,
        }
    }

    fn check_schema(&self, _: &SchemaInferenceState) -> Result<Self::ReturnType, Error> {
        Ok(BOOLEAN_OR_NULLISH.clone())
    }
}
