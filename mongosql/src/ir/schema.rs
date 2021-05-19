use crate::{
    ir::*,
    schema::{Atomic, Document, ResultSet, Satisfaction, Schema, SchemaEnvironment, ANY_DOCUMENT},
};
use common_macros::hash_map;
use linked_hash_map::LinkedHashMap;
use std::collections::{BTreeMap, BTreeSet};
use thiserror::Error;

#[allow(dead_code)]
#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("datasource {0:?} not found in schema environment")]
    DatasourceNotFoundInSchemaEnv(binding_tuple::Key),
    #[error("incorrect argument count for {0}: expected {1}, found {2}")]
    IncorrectArgumentCount(&'static str, usize, usize),
    #[error("schema checking failed for {0}: required {1:?}, found {2:?}")]
    SchemaChecking(&'static str, Schema, Schema),
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
                let source_state = state.with_merged_schema_env(source_result_set.schema);
                let schema = p
                    .expression
                    .iter()
                    .map(|(k, e)| match e.schema(&source_state) {
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
                schema: hash_map! {
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
                    return Err(Error::SchemaChecking(
                        "FieldAccess",
                        accessee_schema,
                        ANY_DOCUMENT.clone(),
                    ));
                }
                if accessee_schema.contains_field(field) == Satisfaction::Not {
                    return Err(Error::AccessMissingField(field.clone()));
                }
                Ok(Expression::get_field_schema(&accessee_schema, &field))
            }
            Expression::Function(f) => f.schema(state),
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
    // schema inference for the ComputedFieldAccess function.
    pub fn computed_field_access_schema(arg_schemas: &[Schema]) -> Result<Schema, Error> {
        if arg_schemas.len() != 2 {
            return Err(Error::IncorrectArgumentCount(
                "ComputedFieldAccess",
                2,
                arg_schemas.len(),
            ));
        }
        if arg_schemas[0].satisfies(&ANY_DOCUMENT) == Satisfaction::Not {
            return Err(Error::SchemaChecking(
                "ComputedFieldAccess",
                arg_schemas[0].clone(),
                ANY_DOCUMENT.clone(),
            ));
        }
        if arg_schemas[1].satisfies(&Schema::AnyOf(vec![
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Null),
            Schema::Missing,
        ])) == Satisfaction::Not
        {
            return Err(Error::SchemaChecking(
                "ComputedFieldAccess",
                arg_schemas[1].clone(),
                Schema::Atomic(Atomic::String),
            ));
        }
        Ok(Schema::Any)
    }

    pub fn schema(&self, arg_schemas: &[Schema]) -> Result<Schema, Error> {
        use Function::*;
        match self {
            ComputedFieldAccess => Function::computed_field_access_schema(arg_schemas),
            _ => unimplemented!(),
        }
    }
}
