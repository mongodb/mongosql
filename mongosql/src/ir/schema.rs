use crate::{
    ir::*,
    map,
    schema::{Atomic, Document, ResultSet, Satisfaction, Schema, SchemaEnvironment, ANY_DOCUMENT},
};
use linked_hash_map::LinkedHashMap;
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

    pub fn with_merged_schema_env(&self, env: SchemaEnvironment) -> SchemaInferenceState {
        let mut merged_env = env;
        for (k, v) in self.env.iter() {
            merged_env.insert(k.clone(), v.clone());
        }
        SchemaInferenceState {
            env: merged_env,
            scope_level: self.scope_level,
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
            Stage::Filter(_) => unimplemented!(),
            Stage::Project(p) => {
                let source_result_set = p.source.schema(state)?;
                let (min_size, max_size) = (source_result_set.min_size, source_result_set.max_size);
                let state = state.with_merged_schema_env(source_result_set.schema);
                let schema = p
                    .expression
                    .iter()
                    .map(|(k, e)| match e.schema(&state) {
                        Ok(s) => Ok((k.clone(), s)),
                        Err(e) => Err(e),
                    })
                    .collect::<Result<SchemaEnvironment, _>>()?;
                Ok(ResultSet {
                    schema,
                    min_size,
                    max_size,
                })
            }
            Stage::Group(_) => unimplemented!(),
            Stage::Limit(_) => unimplemented!(),
            Stage::Offset(_) => unimplemented!(),
            Stage::Sort(_) => unimplemented!(),
            Stage::Collection(c) => Ok(ResultSet {
                schema: map! {
                    (c.collection.clone(), state.scope_level).into() => Schema::Any,
                },
                min_size: None,
                max_size: None,
            }),
            Stage::Array(_) => unimplemented!(),
            Stage::Join(_) => unimplemented!(),
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
        // If it's AnyOf or OneOf we need to apply get_field_schema to each
        // sub-schema and apply AnyOf or OneOf to the results as appropriate.
        // If it's anything else it will Not contain the field, so we return Missing.
        let d = match s {
            Schema::Any => return Schema::Any,
            Schema::Document(d) => d,
            Schema::AnyOf(vs) | Schema::OneOf(vs) => {
                return Schema::AnyOf(
                    vs.iter()
                        .map(|s| Expression::get_field_schema(s, field))
                        .collect(),
                )
            }
            Schema::Missing | Schema::Array(_) | Schema::Atomic(_) => return Schema::Missing,
        };
        // If we find the field in the Document, we just return
        // the Schema for that field, unless the field is not required,
        // then we return OneOf(Schema, Missing).
        if let Some(s) = d.keys.get(field) {
            if d.required.contains(field) {
                return s.clone();
            }
            return Schema::OneOf(vec![s.clone(), Schema::Missing]);
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
            Expression::Function(f) => f.schema(state),
            Expression::Cast(c) => c.schema(state),
            Expression::TypeAssertion(t) => t.schema(state),
            Expression::SubqueryExpression(_) => unimplemented!(),
            Expression::SubqueryComparison(_) => unimplemented!(),
            Expression::Exists(_) => unimplemented!(),
        }
    }

    /// For array literals, we return Array(Any) for an empty array. For other arrays
    /// we return Array(OneOf(S1...SN)), for Schemata S1...SN inferred from the element
    /// Expressions of the array literal.
    fn array_schema(a: &[Expression], state: &SchemaInferenceState) -> Result<Schema, Error> {
        let types = a
            .iter()
            .map(|x| x.schema(state))
            .collect::<Result<Vec<_>, _>>()?;
        Ok(Schema::Array(Box::new(if types.is_empty() {
            Schema::Any
        } else {
            Schema::OneOf(types)
        })))
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
            Integer(_) => Schema::Atomic(Atomic::Int),
            Long(_) => Schema::Atomic(Atomic::Long),
            Double(_) => Schema::Atomic(Atomic::Double),
        })
    }
}

impl FunctionApplication {
    pub fn schema(&self, state: &SchemaInferenceState) -> Result<Schema, Error> {
        let args = self
            .args
            .iter()
            .map(|x| x.schema(state))
            .collect::<Result<Vec<_>, _>>()?;
        self.function.schema(&args)
    }
}

impl Function {
    pub fn schema(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        use Function::*;
        match self {
            // String operators.
            Concat => unimplemented!(),
            Like => unimplemented!(),
            // Unary arithmetic operators.
            Pos | Neg => {
                self.schema_check_fixed_args(
                    arg_schemas,
                    &[Schema::AnyOf(vec![
                        Schema::Atomic(Atomic::Int),
                        Schema::Atomic(Atomic::Long),
                        Schema::Atomic(Atomic::Double),
                        Schema::Atomic(Atomic::Decimal),
                        Schema::Atomic(Atomic::Null),
                        Schema::Missing,
                    ])],
                )?;
                Ok(arg_schemas[0].clone())
            }
            // Arithmetic operators with variadic arguments.
            Add | Mult => {
                self.schema_check_variadic_args(
                    arg_schemas,
                    Schema::AnyOf(vec![
                        Schema::Atomic(Atomic::Int),
                        Schema::Atomic(Atomic::Long),
                        Schema::Atomic(Atomic::Double),
                        Schema::Atomic(Atomic::Decimal),
                        Schema::Atomic(Atomic::Null),
                        Schema::Missing,
                    ]),
                )?;
                Ok(Function::get_arithmetic_schema(arg_schemas))
            }
            // Arithmetic operators with fixed (two) arguments.
            Sub | Div => {
                self.schema_check_fixed_args(
                    arg_schemas,
                    &[
                        Schema::AnyOf(vec![
                            Schema::Atomic(Atomic::Int),
                            Schema::Atomic(Atomic::Long),
                            Schema::Atomic(Atomic::Double),
                            Schema::Atomic(Atomic::Decimal),
                            Schema::Atomic(Atomic::Null),
                            Schema::Missing,
                        ]),
                        Schema::AnyOf(vec![
                            Schema::Atomic(Atomic::Int),
                            Schema::Atomic(Atomic::Long),
                            Schema::Atomic(Atomic::Double),
                            Schema::Atomic(Atomic::Decimal),
                            Schema::Atomic(Atomic::Null),
                            Schema::Missing,
                        ]),
                    ],
                )?;
                Ok(Function::get_arithmetic_schema(arg_schemas))
            }
            // Comparison operators.
            Lt => unimplemented!(),
            Lte => unimplemented!(),
            Ne => unimplemented!(),
            Eq => unimplemented!(),
            Gt => unimplemented!(),
            Gte => unimplemented!(),
            Between => unimplemented!(),
            // Boolean operators.
            Not => unimplemented!(),
            And => unimplemented!(),
            Or => unimplemented!(),
            // Control-flow operator.
            Case => unimplemented!(),
            // Type operator.
            Is => unimplemented!(),
            // Computed Field Access operator when the field is not known until runtime.
            ComputedFieldAccess => {
                self.schema_check_fixed_args(
                    arg_schemas,
                    &[
                        ANY_DOCUMENT.clone(),
                        Schema::AnyOf(vec![
                            Schema::Atomic(Atomic::String),
                            Schema::Atomic(Atomic::Null),
                            Schema::Missing,
                        ]),
                    ],
                )?;
                Ok(Schema::Any)
            }
            // Conditional scalar functions.
            Nullif => unimplemented!(),
            Coalesce => unimplemented!(),
            // Array scalar functions
            Slice => unimplemented!(),
            Size => unimplemented!(),
            // Numeric value scalar functions.
            Position => unimplemented!(),
            CharLen => unimplemented!(),
            OctetLen => unimplemented!(),
            BitLen => unimplemented!(),
            Extract => unimplemented!(),
            // String value scalar functions.
            Substring => unimplemented!(),
            Upper => unimplemented!(),
            Lower => unimplemented!(),
            Trim => unimplemented!(),
            // Datetime value scalar functions.
            CurrentTimestamp => unimplemented!(),
        }
    }

    /// Checks a function's argument count and its arguments' types against the required schemas.
    /// Used for functions with a fixed (predetermined) number of arguments.
    ///
    /// Since the argument count is fixed, the vector of required schemas must correspond 1-to-1
    /// with the array of argument schemas.
    fn schema_check_fixed_args(
        &self,
        arg_schemas: &[Schema],
        required_schemas: &[Schema],
    ) -> Result<(), Error> {
        if arg_schemas.len() != required_schemas.len() {
            return Err(Error::IncorrectArgumentCount {
                name: self.as_str(),
                required: required_schemas.len(),
                found: arg_schemas.len(),
            });
        }
        for (i, arg) in arg_schemas.iter().enumerate() {
            if arg.satisfies(&required_schemas[i]) != Satisfaction::Must {
                return Err(Error::SchemaChecking {
                    name: self.as_str(),
                    required: required_schemas[i].clone(),
                    found: arg.clone(),
                });
            }
        }
        Ok(())
    }

    /// Checks a function's arguments' types against the required schema.
    /// Used for functions with a variadic number of arguments.
    ///
    /// Since the argument count can vary, the required schema is a single
    /// value that's compared against each of the arguments.
    fn schema_check_variadic_args(
        &self,
        arg_schemas: &[Schema],
        required_schema: Schema,
    ) -> Result<(), Error> {
        self.schema_check_fixed_args(
            arg_schemas,
            &std::iter::repeat(required_schema)
                .take(arg_schemas.len())
                .collect::<Vec<Schema>>(),
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
        const INTEGER: &Schema = &Schema::Atomic(Atomic::Int);

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

impl CastExpression {
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

impl TypeAssertionExpression {
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
