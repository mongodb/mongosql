use crate::{
    get_schema_for_path_mut,
    negative_normalize::{NegativeNormalize, DECIMAL_ZERO},
    schema_for_bson, DeriveSchema, Result, ResultSetState,
};
use agg_ast::definitions::{MatchBinaryOp, MatchExpression, MatchField, MatchStage};
use bson::Bson;
use mongosql::{
    map,
    schema::{Atomic, Document, Schema, NUMERIC},
    set,
};
use std::collections::BTreeSet;

fn is_unitary_schema(schema: &Schema) -> bool {
    matches!(
        schema,
        Schema::Atomic(Atomic::Null)
            | Schema::Missing
            | Schema::Atomic(Atomic::MinKey)
            | Schema::Atomic(Atomic::MaxKey)
            | Schema::Atomic(Atomic::Undefined)
    )
}

macro_rules! maybe_any_of {
    ($schemas:expr) => {
        if $schemas.len() == 1 {
            $schemas.into_iter().next().unwrap()
        } else {
            Schema::AnyOf($schemas)
        }
    };
}

macro_rules! any_expand {
    () => {
        Schema::AnyOf(set![
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::String),
            Schema::Document(Document::any()),
            Schema::Array(Box::new(Schema::Any)),
            Schema::Atomic(Atomic::BinData),
            Schema::Atomic(Atomic::Undefined),
            Schema::Atomic(Atomic::ObjectId),
            Schema::Atomic(Atomic::Boolean),
            Schema::Atomic(Atomic::Date),
            Schema::Missing,
            Schema::Atomic(Atomic::Null),
            Schema::Atomic(Atomic::Regex),
            Schema::Atomic(Atomic::DbPointer),
            Schema::Atomic(Atomic::Javascript),
            Schema::Atomic(Atomic::Symbol),
            Schema::Atomic(Atomic::JavascriptWithScope),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Timestamp),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::MinKey),
            Schema::Atomic(Atomic::MaxKey),
        ])
    };
}

macro_rules! bits_schema {
    () => {
        Schema::AnyOf(set![
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::BinData),
        ])
    };
}

macro_rules! geo_schema {
    () => {
        Schema::AnyOf(set![
            Schema::Document(Document {
                keys: map! {
                    "type".to_string() => Schema::Atomic(Atomic::String),
                    "coordinates".to_string() => Schema::Array(Box::new(NUMERIC.clone())),
                },
                required: set!["coordinates".to_string()],
                additional_properties: false,
                jaccard_index: None,
            }),
            Schema::Array(Box::new(NUMERIC.clone())),
        ])
    };
}

// is_exists_true_bson returns true if the BSON value is considered to truthy in exist in a match
// operation.
fn is_exists_true_bson(bson: &Bson) -> bool {
    match bson {
        Bson::Null => false,
        Bson::Boolean(b) => *b,
        Bson::Int32(i) => *i != 0,
        Bson::Int64(i) => *i != 0,
        Bson::Double(d) => *d != 0.0,
        Bson::Decimal128(d) => *d != *DECIMAL_ZERO,
        _ => true,
    }
}

// schema_for_match_bson_literal generates the proper schema for a Bson object used in a $match.
// This is a bit complicated in that mongodb will allow a 1 integer, for instance, to match all
// numeric types but something like 1.5 can only match double and decimal128, but a 1.0 Decimal128
// or Double can match an integer 1.
fn schema_for_match_bson_literal(bson: &Bson, include_missing: bool) -> Schema {
    // helper to handle doubles
    let schema_for_match_double = |d: f64| {
        if d.fract() == 0.0 {
            NUMERIC.clone()
        } else {
            Schema::AnyOf(set![
                Schema::Atomic(Atomic::Double),
                Schema::Atomic(Atomic::Decimal),
            ])
        }
    };
    match bson {
        Bson::Int32(_) | Bson::Int64(_) => NUMERIC.clone(),
        Bson::Double(d) => schema_for_match_double(*d),
        Bson::Decimal128(d) => {
            let d = d.to_string().parse::<f64>();
            match d {
                Ok(d) => schema_for_match_double(d),
                // If this can't be parsed into a double, we can clearly only match Decimal128
                Err(_) => Schema::Atomic(Atomic::Decimal),
            }
        }
        // Missing is considered equivalent to Null in $match
        Bson::Null => {
            if include_missing {
                Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing])
            } else {
                Schema::Atomic(Atomic::Null)
            }
        }
        b => schema_for_bson(b),
    }
}

// schema_for_type_str generates a schema for a type name string used in a $type match operation.
fn schema_for_type_str(type_str: &str) -> Schema {
    match type_str {
        "double" => Schema::Atomic(Atomic::Double),
        "string" => Schema::Atomic(Atomic::String),
        "object" => Schema::Document(Document::any()),
        "array" => Schema::Array(Box::new(Schema::Any)),
        "binData" => Schema::Atomic(Atomic::BinData),
        "undefined" => Schema::Atomic(Atomic::Undefined),
        "objectId" => Schema::Atomic(Atomic::ObjectId),
        "bool" => Schema::Atomic(Atomic::Boolean),
        "date" => Schema::Atomic(Atomic::Date),
        "null" => Schema::Atomic(Atomic::Null),
        "regex" => Schema::Atomic(Atomic::Regex),
        "dbPointer" => Schema::Atomic(Atomic::DbPointer),
        "javascript" => Schema::Atomic(Atomic::Javascript),
        "symbol" => Schema::Atomic(Atomic::Symbol),
        "javascriptWithScope" => Schema::Atomic(Atomic::JavascriptWithScope),
        "int" => Schema::Atomic(Atomic::Integer),
        "timestamp" => Schema::Atomic(Atomic::Timestamp),
        "long" => Schema::Atomic(Atomic::Long),
        "decimal" => Schema::Atomic(Atomic::Decimal),
        "minKey" => Schema::Atomic(Atomic::MinKey),
        "maxKey" => Schema::Atomic(Atomic::MaxKey),
        _ => unreachable!(),
    }
}

// schema_for_type_expr generates a schema for a type expression used in a $type match operation.
fn schema_for_type_expr(bson: &Bson) -> Schema {
    match bson {
        Bson::String(s) => schema_for_type_str(s),
        Bson::Array(values) => {
            let schemas = values
                .iter()
                .map(|v| match v {
                    Bson::String(s) => schema_for_type_str(s),
                    _ => unreachable!(),
                })
                .collect::<BTreeSet<_>>();
            maybe_any_of!(schemas)
        }
        // this should never happen because this should result in a statically unacceptable query
        _ => Schema::Any,
    }
}

// schema_for_array_bson_values generates a schema for an array of BSON values used in a match
// operation. This currently only happens in the cases of $in, $nin, and $all. While $type takes
// an array, it is array of type name strings and must be handled separately.
fn schema_for_array_bson_values(bson: &Bson, include_missing: bool) -> Schema {
    match bson {
        Bson::Array(values) => {
            let schemas = values
                .iter()
                .map(|x| schema_for_match_bson_literal(x, include_missing))
                .collect::<BTreeSet<_>>();
            maybe_any_of!(schemas)
        }
        // this should never happen because this should result in a statically unacceptable query
        _ => Schema::Any,
    }
}

// promote_missing adds Schema::Missing to any non-required key in a Document schema. This is
// necessary for our operations to work correctly, since they operate on paths, leaving no way
// to check or modify required. We will rely on Schema::simply to lower Schema::Missing back to
// removing from required, since Schema::Missing cannot be serialized.
fn promote_missing(schema: &Schema) -> Schema {
    match schema {
        Schema::AnyOf(schemas) => {
            let schemas = schemas.iter().map(promote_missing).collect::<BTreeSet<_>>();
            maybe_any_of!(schemas)
        }
        Schema::Array(schema) => Schema::Array(Box::new(promote_missing(schema))),
        Schema::Document(doc) => {
            let mut doc = doc.clone();
            for (key, schema) in doc.keys.iter_mut() {
                if !doc.required.contains(key) {
                    *schema = schema.union(&Schema::Missing);
                    doc.required.insert(key.clone());
                }
            }
            Schema::Document(doc)
        }
        _ => schema.clone(),
    }
}

impl DeriveSchema for MatchStage {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        state.result_set_schema = promote_missing(&state.result_set_schema);
        for expr in self.expr.iter() {
            let expr = expr.get_negative_normal_form();
            expr.derive_schema(state)?;
        }
        Ok(state.result_set_schema.clone())
    }
}

impl DeriveSchema for MatchExpression {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        match self {
            MatchExpression::Expr(_) => todo!(),
            MatchExpression::Misc(_) => todo!(),
            MatchExpression::Logical(_) => todo!(),
            MatchExpression::Field(f) => f.derive_schema(state),
        }
    }
}

impl DeriveSchema for MatchField {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        let path: Vec<_> = self
            .field
            .as_str()
            .split('.')
            .map(|s| s.to_string())
            .collect();
        self.ops.iter().for_each(|(op, b)| {
            derive_schema_for_match_op(path.clone(), *op, b, state);
        });
        // This result will not be used, but this is the most accurate Schema.
        Ok(Schema::Atomic(Atomic::Boolean))
    }
}

// schema_difference removes a set of Schema from another Schema. This differs from
// Schema::intersection in that it does not use two Schemas as operands. Part of this is that
// schema_difference only ever happens with Atomic Schemas (and Missing, which is rather isomorphic
// to Atomic) and this is expedient. If we ever need to expand this to more complex Schemas, it may
// make sense to make this a real operator on two Schemas in the schema module.
//
// Note that this could also be achieved by complementing the Schema to be removed and intersecting
// it with the Schema to be modified, but this would be quite a bit less efficient.
fn schema_difference(schema: &mut Schema, to_remove: BTreeSet<Schema>) {
    match schema {
        Schema::Any => {
            *schema = any_expand!();
            schema_difference(schema, to_remove);
        }
        Schema::AnyOf(schemas) => {
            let any_of_schemas = schemas
                .difference(&to_remove)
                .cloned()
                .collect::<BTreeSet<_>>();
            *schema = maybe_any_of!(any_of_schemas);
        }
        _ => (),
    }
}

// derive_schema_for_match_op derives the schema for a single match operation, since a given field
// may be nested above multiple match operations. This is a helper function for MatchField::derive_schema.
fn derive_schema_for_match_op(
    path: Vec<String>,
    op: MatchBinaryOp,
    b: &Bson,
    state: &mut ResultSetState,
) {
    macro_rules! schema_intersect {
        ($path:expr, $state:expr, $schema:expr) => {
            let field_schema = get_schema_for_path_mut(&mut $state.result_set_schema, $path);
            match field_schema {
                Some(field_schema) => {
                    *field_schema = $schema.intersection(field_schema);
                }
                None => (),
            }
        };
    }
    macro_rules! schema_difference {
        ($path:expr, $state:expr, $schemas:expr) => {
            let field_schema = get_schema_for_path_mut(&mut $state.result_set_schema, $path);
            match field_schema {
                Some(field_schema) => {
                    schema_difference(field_schema, $schemas);
                }
                None => (),
            }
        };
    }
    match op {
        MatchBinaryOp::Eq | MatchBinaryOp::Gte | MatchBinaryOp::Lte => {
            let schema = schema_for_match_bson_literal(b, true);
            schema_intersect!(path, state, schema);
        }
        MatchBinaryOp::Gt | MatchBinaryOp::Lt => {
            let schema = schema_for_match_bson_literal(b, false);
            schema_intersect!(path, state, schema);
        }
        // Ne actually does not tell us anything about Schema except for types that are
        // unitary-valued: Null, MinKey, MaxKey, Undefined.
        MatchBinaryOp::Ne => {
            let schema = schema_for_match_bson_literal(b, true);
            if is_unitary_schema(&schema) {
                schema_difference!(path, state, set![schema]);
            } else if let Schema::AnyOf(schemas) = schema {
                let to_remove = schemas.into_iter().filter(is_unitary_schema).collect();
                schema_difference!(path, state, to_remove);
            }
        }
        MatchBinaryOp::Exists => {
            if is_exists_true_bson(b) {
                schema_difference!(path, state, set![Schema::Missing]);
            } else {
                schema_intersect!(path, state, Schema::Missing);
            }
        }
        MatchBinaryOp::Type => {
            let schema = schema_for_type_expr(b);
            schema_intersect!(path, state, schema);
        }
        MatchBinaryOp::In => {
            let schema = schema_for_array_bson_values(b, true);
            schema_intersect!(path, state, schema);
        }
        MatchBinaryOp::Nin => {
            let schema = schema_for_array_bson_values(b, true);
            if is_unitary_schema(&schema) {
                schema_difference!(path, state, set![schema]);
            } else if let Schema::AnyOf(schemas) = schema {
                let to_remove = schemas
                    .iter()
                    .cloned()
                    .flat_map(|schema| {
                        if let Schema::AnyOf(schemas) = schema {
                            schemas
                        } else {
                            set! {schema}
                        }
                    })
                    .filter(is_unitary_schema)
                    .collect();
                schema_difference!(path, state, to_remove);
            }
        }
        MatchBinaryOp::Mod => {
            schema_intersect!(path, state, NUMERIC.clone());
        }
        MatchBinaryOp::Size => {
            let schema = Schema::Array(Box::new(Schema::Any));
            schema_intersect!(path, state, schema);
        }
        MatchBinaryOp::All => {
            let schema = Schema::Array(Box::new(schema_for_array_bson_values(b, false)));
            schema_intersect!(path, state, schema);
        }
        MatchBinaryOp::BitsAllClear
        | MatchBinaryOp::BitsAnyClear
        | MatchBinaryOp::BitsAllSet
        | MatchBinaryOp::BitsAnySet => {
            schema_intersect!(path, state, bits_schema!());
        }
        MatchBinaryOp::Near
        | MatchBinaryOp::GeoWithin
        | MatchBinaryOp::NearSphere
        | MatchBinaryOp::GeoIntersects => {
            schema_intersect!(path, state, geo_schema!());
        }
    }
}
