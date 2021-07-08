use crate::{
    ir::{binding_tuple::Key, *},
    map,
    schema::{
        Atomic, Document, ResultSet, Satisfaction, Schema, SchemaEnvironment, ANY_DOCUMENT,
        BOOLEAN_OR_NULLISH, DATE_OR_NULLISH, INTEGER_OR_NULLISH, NULLISH, NUMERIC_OR_NULLISH,
        STRING_OR_NULLISH,
    },
};
use linked_hash_map::LinkedHashMap;
use std::cmp::min;
use std::collections::{BTreeMap, BTreeSet};
use thiserror::Error;

#[allow(dead_code)]
#[derive(Debug, Error, PartialEq)]
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
    #[error("cannot access field {0} because it does not exist")]
    AccessMissingField(String),
    #[error("invalid JSON schema: {0}")]
    InvalidJsonSchema(String),
    #[error("cardinality of the subquery's result set may be greater than 1")]
    InvalidSubqueryCardinality,
    #[error("cannot create schema environment with duplicate datasource: {0:?}")]
    DuplicateKeyError(Key),
}

#[derive(Debug, Clone)]
pub struct SchemaInferenceState {
    pub scope_level: u16,
    pub env: SchemaEnvironment,
}

impl SchemaInferenceState {
    #[allow(dead_code)]
    pub fn new(scope_level: u16, env: SchemaEnvironment) -> SchemaInferenceState {
        SchemaInferenceState { scope_level, env }
    }

    pub fn with_merged_schema_env(
        &self,
        env: SchemaEnvironment,
    ) -> Result<SchemaInferenceState, Error> {
        Ok(SchemaInferenceState {
            env: env
                .with_merged_mappings(self.env.clone())
                .map_err(|e| Error::DuplicateKeyError(e.key))?,
            scope_level: self.scope_level,
        })
    }

    pub fn subquery_state(&self) -> SchemaInferenceState {
        SchemaInferenceState {
            scope_level: self.scope_level + 1,
            env: self.env.clone(),
        }
    }
}

impl Stage {
    /// Recursively schema checks this stage, its sources, and all
    /// contained expressions. If schema checking succeeds, returns a
    /// [`ResultSet`] describing the schema of the result set returned
    /// by this stage. The provided [`SchemaInferenceState`] should
    /// include schema information for any datasources from outer
    /// scopes. Schema information for the current scope will be
    /// obtained by calling [`Stage::schema`] on source stages.
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<ResultSet, Error> {
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
            Stage::Group(_) => unimplemented!(),
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
            Stage::Sort(_) => unimplemented!(),
            Stage::Collection(c) => Ok(ResultSet {
                schema_env: map! {
                    // we know the top level elements of a collection must be a Document,
                    // but we do not know what kind, so we return ANY_DOCUMENT.
                    (c.collection.clone(), state.scope_level).into() => ANY_DOCUMENT.clone(),
                },
                min_size: 0,
                max_size: None,
            }),
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
                let state = state
                    .with_merged_schema_env(left_result_set.schema_env.clone())?
                    .with_merged_schema_env(right_result_set.schema_env.clone())?;
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
                                            (key, Schema::AnyOf(vec![Schema::Missing, schema]))
                                        })
                                        .collect::<SchemaEnvironment>(),
                                )
                                .map_err(|e| Error::DuplicateKeyError(e.key))?,
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
                                .map_err(|e| Error::DuplicateKeyError(e.key))?,
                            min_size,
                            max_size,
                        })
                    }
                }
            }
            Stage::Set(_) => unimplemented!(),
        }
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
            return Schema::AnyOf(vec![s.clone(), Schema::Missing]);
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

    /// Recursively schema checks this expression, its arguments, and
    /// all contained expressions/stages. If schema checking succeeds,
    /// returns a [`Schema`] describing this expression's schema. The
    /// provided [`SchemaInferenceState`] should include schema
    /// information for all datasources in scope.
    #[allow(dead_code)]
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        match self {
            Expression::Literal(lit) => lit.schema(state),
            Expression::Reference(key) => state
                .env
                .get(key)
                .cloned()
                .ok_or_else(|| Error::DatasourceNotFoundInSchemaEnv(key.clone())),
            Expression::Array(a) => Expression::array_schema(a, state),
            Expression::Document(d) => Expression::document_schema(d, state),
            Expression::FieldAccess(FieldAccess { expr, field }) => {
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
                Ok(Expression::get_field_schema(&accessee_schema, &field))
            }
            Expression::ScalarFunction(f) => f.schema(state),
            Expression::Cast(c) => c.schema(state),
            Expression::SearchedCase(sc) => sc.schema(state),
            Expression::SimpleCase(_) => unimplemented!(),
            Expression::TypeAssertion(t) => t.schema(state),
            Expression::SubqueryExpression(SubqueryExpr {
                output_expr,
                subquery,
            }) => {
                let result_set = subquery.schema(&state.subquery_state())?;
                let schema = output_expr.schema(&SchemaInferenceState {
                    scope_level: state.scope_level + 1,
                    env: result_set.schema_env,
                })?;

                // The subquery's result set MUST have a cardinality of 0 or 1. The returned schema
                // MUST include MISSING if the cardinality of the result set MAY be 0. We can exclude
                // MISSING from the returned schema if the cardinality of the result set MUST be 1.
                if result_set.max_size == None || result_set.max_size.unwrap() > 1 {
                    return Err(Error::InvalidSubqueryCardinality);
                }
                match result_set.min_size {
                    0 => Ok(Schema::AnyOf(vec![schema, Schema::Missing])),
                    1 => Ok(schema),
                    _ => Err(Error::InvalidSubqueryCardinality),
                }
            }
            Expression::SubqueryComparison(_) => unimplemented!(),
            Expression::Exists(stage) => {
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
                Ok(Schema::AnyOf(vec![
                    Schema::Atomic(Atomic::Null),
                    Schema::Atomic(Atomic::Boolean),
                ]))
            }
        }
    }

    fn array_schema(a: &[Expression], state: &SchemaInferenceState) -> Result<Schema, Error> {
        Ok(Schema::Array(Box::new(Expression::array_items_schema(
            a, state,
        )?)))
    }

    /// For array literals, we return Array(Any) for an empty array. For other arrays
    /// we return Array(AnyOf(S1...SN)), for Schemata S1...SN inferred from the element
    /// Expressions of the array literal.
    fn array_items_schema(a: &[Expression], state: &SchemaInferenceState) -> Result<Schema, Error> {
        Ok(Schema::AnyOf(
            a.iter()
                .map(|e| e.schema(state).map(|s| s.upconvert_missing_to_null()))
                .collect::<Result<Vec<_>, _>>()?,
        ))
    }

    /// For document literals, we infer the most restrictive schema possible. This means
    /// that additional_properties are not allowed.
    fn document_schema(
        d: &LinkedHashMap<String, Expression>,
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

impl Literal {
    pub fn schema(&self, _state: &SchemaInferenceState) -> Result<Schema, Error> {
        use Literal::*;
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

impl ScalarFunction {
    pub fn schema(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        use ScalarFunction::*;
        match self {
            // String operators.
            Concat => Ok(
                match self.schema_check_fixed_args(
                    arg_schemas,
                    &[STRING_OR_NULLISH.clone(), STRING_OR_NULLISH.clone()],
                )? {
                    Satisfaction::May => Schema::AnyOf(vec![
                        Schema::Atomic(Atomic::String),
                        Schema::Atomic(Atomic::Null),
                    ]),
                    Satisfaction::Not => Schema::Atomic(Atomic::String),
                    Satisfaction::Must => Schema::Atomic(Atomic::Null),
                },
            ),
            // Unary arithmetic operators.
            Pos | Neg => {
                self.schema_check_fixed_args(arg_schemas, &[NUMERIC_OR_NULLISH.clone()])?;
                Ok(arg_schemas[0].clone().upconvert_missing_to_null())
            }
            // Arithmetic operators with variadic arguments.
            Add | Mul => {
                self.schema_check_variadic_args(arg_schemas, NUMERIC_OR_NULLISH.clone())?;
                Ok(ScalarFunction::get_arithmetic_schema(arg_schemas))
            }
            // Arithmetic operators with fixed (two) arguments.
            Sub | Div => {
                self.schema_check_fixed_args(
                    arg_schemas,
                    &[NUMERIC_OR_NULLISH.clone(), NUMERIC_OR_NULLISH.clone()],
                )?;
                Ok(ScalarFunction::get_arithmetic_schema(arg_schemas))
            }
            // Comparison operators.
            Lt => unimplemented!(),
            Lte => unimplemented!(),
            Neq => unimplemented!(),
            Eq => unimplemented!(),
            Gt => unimplemented!(),
            Gte => unimplemented!(),
            Between => unimplemented!(),
            // Boolean operators.
            Not => unimplemented!(),
            And => unimplemented!(),
            Or => unimplemented!(),
            // Computed Field Access operator when the field is not known until runtime.
            ComputedFieldAccess => {
                self.schema_check_fixed_args(
                    arg_schemas,
                    &[ANY_DOCUMENT.clone(), Schema::Atomic(Atomic::String)],
                )?;
                Ok(Schema::Any)
            }
            // Conditional scalar functions.
            NullIf => unimplemented!(),
            Coalesce => unimplemented!(),
            // Array scalar functions
            Slice => unimplemented!(),
            Size => unimplemented!(),
            // Numeric value scalar functions.
            Position => unimplemented!(),
            CharLength => unimplemented!(),
            OctetLength => unimplemented!(),
            BitLength => unimplemented!(),
            Year | Month | Day | Hour | Minute | Second => Ok(
                match self.schema_check_fixed_args(arg_schemas, &[DATE_OR_NULLISH.clone()])? {
                    Satisfaction::May => Schema::AnyOf(vec![
                        Schema::Atomic(Atomic::Integer),
                        Schema::Atomic(Atomic::Null),
                    ]),
                    Satisfaction::Not => Schema::Atomic(Atomic::Integer),
                    Satisfaction::Must => Schema::Atomic(Atomic::Null),
                },
            ),
            // String value scalar functions.
            Substring => self.get_substring_schema(arg_schemas),
            Upper | Lower => Ok(
                match self.schema_check_fixed_args(arg_schemas, &[STRING_OR_NULLISH.clone()])? {
                    Satisfaction::May => Schema::AnyOf(vec![
                        Schema::Atomic(Atomic::String),
                        Schema::Atomic(Atomic::Null),
                    ]),
                    Satisfaction::Not => Schema::Atomic(Atomic::String),
                    Satisfaction::Must => Schema::Atomic(Atomic::Null),
                },
            ),
            LTrim | RTrim | BTrim => Ok(
                match self.schema_check_fixed_args(
                    arg_schemas,
                    &[STRING_OR_NULLISH.clone(), STRING_OR_NULLISH.clone()],
                )? {
                    Satisfaction::May => Schema::AnyOf(vec![
                        Schema::Atomic(Atomic::String),
                        Schema::Atomic(Atomic::Null),
                    ]),
                    Satisfaction::Not => Schema::Atomic(Atomic::String),
                    Satisfaction::Must => Schema::Atomic(Atomic::Null),
                },
            ),
            // Datetime value scalar function.
            CurrentTimestamp => {
                self.schema_check_fixed_args(arg_schemas, &[])?;
                Ok(Schema::Atomic(Atomic::Date))
            }
        }
    }

    /// Checks a function's argument count and its arguments' types against the required schemas.
    /// Used for functions with a fixed (predetermined) number of arguments, including 0.
    ///
    /// Since the argument count is fixed, the slice of required schemas must correspond 1-to-1
    /// with the slice of argument schemas.
    ///
    /// The boolean result returns whether a NULLISH argument is a possibility.
    fn schema_check_fixed_args(
        &self,
        arg_schemas: &[Schema],
        required_schemas: &[Schema],
    ) -> Result<Satisfaction, Error> {
        if arg_schemas.len() != required_schemas.len() {
            return Err(Error::IncorrectArgumentCount {
                name: self.as_str(),
                required: required_schemas.len(),
                found: arg_schemas.len(),
            });
        }
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

    /// Returns the schema for the substring function.
    ///
    /// The error checks include special handling for an optional third argument.
    /// That is, a valid substring function can only be one of:
    ///   - SUBSTRING(<string> FROM <start>)
    ///   - SUBSTRING(<string> FROM <start> FOR <length>)
    ///
    /// We first check the schema for 3 args. If the check fails specifically due
    /// to an incorrect argument count, we check the schema for 2 args instead.
    fn get_substring_schema(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        Ok(
            match self
                .schema_check_fixed_args(
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
                })? {
                Satisfaction::May => Schema::AnyOf(vec![
                    Schema::Atomic(Atomic::String),
                    Schema::Atomic(Atomic::Null),
                ]),
                Satisfaction::Not => Schema::Atomic(Atomic::String),
                Satisfaction::Must => Schema::Atomic(Atomic::Null),
            },
        )
    }

    /// Returns the schema type for an arithmetic (Add, Sub, Mult, Div)
    /// function in accordance with the following table:
    /// +-------------------------+-------------------+---------+
    /// |    Type of Operand 1    | Type of Operand 2 | Result  |
    /// +-------------------------+-------------------+---------+
    /// | INT                     | INT               | INT     |
    /// | INT or LONG             | LONG              | LONG    |
    /// | Any numeric non-DECIMAL | DOUBLE            | DOUBLE  |
    /// | Any numeric             | DECIMAL           | DECIMAL |
    /// +-------------------------+-------------------+---------+
    ///
    /// For `Add` and `Mult` which have variadic arguments:
    /// - If no operand is provided, return an integer type.
    /// - If one operand is provided, return the type of that operand.
    /// - Otherwise operands op1, op2, ..., opN are reduced in a
    ///   pairwise fashion, with the order of priority from
    ///   'largest' to 'smallest' being:
    ///
    ///    null | missing > decimal > double > long > int
    ///
    /// The provided `arg_schemas` must have passed schema checking prior.
    fn get_arithmetic_schema(arg_schemas: &[Schema]) -> Schema {
        const NULL: &Schema = &Schema::Atomic(Atomic::Null);
        const MISSING: &Schema = &Schema::Missing;
        const DECIMAL: &Schema = &Schema::Atomic(Atomic::Decimal);
        const DOUBLE: &Schema = &Schema::Atomic(Atomic::Double);
        const LONG: &Schema = &Schema::Atomic(Atomic::Long);
        const INTEGER: &Schema = &Schema::Atomic(Atomic::Integer);

        arg_schemas
            .iter()
            .reduce(|op1, op2| match (op1, op2) {
                (NULL, _) | (_, NULL) | (MISSING, _) | (_, MISSING) => NULL,
                (DECIMAL, _) | (_, DECIMAL) => DECIMAL,
                (DOUBLE, _) | (_, DOUBLE) => DOUBLE,
                (LONG, _) | (_, LONG) => LONG,
                (INTEGER, _) | (_, INTEGER) => INTEGER,
                (l, r) => unreachable!(
                    "argument schemas {:?} and {:?} should be disallowed by schema checking",
                    l, r
                ),
            })
            .unwrap_or(INTEGER)
            .clone()
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
        if expr_schema.satisfies(&Schema::AnyOf(vec![
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

        Ok(Schema::AnyOf(vec![
            type_schema,
            on_null_schema,
            on_error_schema,
        ]))
    }
}

impl SearchedCaseExpr {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let mut schemas = self
            .when_branch
            .iter()
            .map(|wb| {
                Ok({
                    // the when branch must evaluate to BOOLEAN_OR_NULLISH.
                    let when_schema = wb.when.schema(state)?;
                    if when_schema.satisfies(&BOOLEAN_OR_NULLISH) != Satisfaction::Must {
                        return Err(Error::SchemaChecking {
                            name: "SearchedCase",
                            required: BOOLEAN_OR_NULLISH.clone(),
                            found: when_schema,
                        });
                    }
                    // the Schema of the when branch is collected to the output.
                    wb.then.schema(state)?
                })
            })
            .collect::<Result<Vec<_>, _>>()?;
        schemas.push(self.else_branch.schema(state)?);
        // The resulting Schema is the AnyOf every WHEN branch with the ELSE branch.
        Ok(Schema::AnyOf(schemas))
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
