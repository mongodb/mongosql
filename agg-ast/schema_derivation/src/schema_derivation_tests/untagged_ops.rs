use crate::schema_derivation::{DeriveSchema, ResultSetState};
use agg_ast::definitions::Expression;
use mongosql::{
    map,
    schema::{Atomic, Document, Schema},
    set,
};
use std::collections::BTreeMap;

macro_rules! test_type_conversion_op {
        ($func_name:ident, expected = $expected:expr, op = $op:expr) => {
            #[test]
            fn $func_name() {
                let mut state = ResultSetState {
                    catalog: &BTreeMap::new(),
                    variables: &BTreeMap::new(),
                    result_set_schema: Schema::Document(Document {
                        keys: map! {"foo".to_string() => Schema::Atomic(Atomic::Null)},
                        ..Default::default()
                    })
                };
                let input: Expression = serde_json::from_str(format!("{{\"{0}\":\"$foo\"}}", $op).as_str()).unwrap();
                // if the input schema is null, we should return null
                let result = input.derive_schema(&mut state);
                assert_eq!(result, Ok(Schema::Atomic(Atomic::Null)));
                // if the input schema is not nullable, return the expected type
                state.result_set_schema = Schema::Document(Document {
                    keys: map! {"foo".to_string() => Schema::Atomic(Atomic::Integer)},
                    ..Default::default()
                });
                let result = input.derive_schema(&mut state);
                assert_eq!(result, Ok($expected));
                // if the input schema is null or some, return expected type, nullable
                state.result_set_schema = Schema::Document(Document {
                    keys: map! {"foo".to_string() => Schema::AnyOf(set!(Schema::Missing, Schema::Atomic(Atomic::Null), Schema::Atomic(Atomic::Integer)))},
                    ..Default::default()
                });
                let result = input.derive_schema(&mut state);
                assert_eq!(result, Ok(Schema::AnyOf(set!(Schema::Atomic(Atomic::Null), $expected))));
            }
        };
    }

mod array_ops {
    use super::*;
    test_derive_schema!(
        array_elem_at_missing_field,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$arrayElemAt": ["$food", 0]}"#
    );
    test_derive_schema!(
        array_elem_at_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$arrayElemAt": [null, 0]}"#
    );
    test_derive_schema!(
        array_elem_at_array_field,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{"$arrayElemAt": ["$foo", 0]}"#,
        ref_schema = Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))
    );
    test_derive_schema!(
        array_elem_at_literal_array,
        expected = Ok(Schema::AnyOf(
            set! {Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::String)}
        )),
        input = r#"{"$arrayElemAt": [[1, "hello", 3], 0]}"#
    );
    test_derive_schema!(
        array_to_object,
        expected = Ok(Schema::Document(Document::any())),
        input = r#"{"$arrayToObject": [["foo", 1], ["bar", "hello"]]}"#
    );
    test_derive_schema!(
        concat_arrays,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::String)
        ))))),
        input = r#"{"$concatArrays": ["$foo", ["hello", "world"]]}"#,
        ref_schema = Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double)
        ))))
    );
    test_derive_schema!(
        set_union,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::String)
        ))))),
        input = r#"{"$setUnion": ["$foo", ["hello", "world"]]}"#,
        ref_schema = Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double)
        ))))
    );
    test_derive_schema!(
        set_intersection_multi,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ))))),
        input = r#"{"$setIntersection": ["$foo", ["hello", "world", {"$numberDecimal": "42"}]]}"#,
        ref_schema = Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double)
        ))))
    );
    test_derive_schema!(
        set_intersection_multi_int_only,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ))))),
        input = r#"{"$setIntersection": ["$foo", [42, 54]]}"#,
        ref_schema = Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double)
        ))))
    );
    test_derive_schema!(
        set_intersection_empty,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Null),))),
        input = r#"{"$setIntersection": ["$foo", []]}"#
    );
    test_derive_schema!(
        set_intersection_empty_set,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Null),))),
        input = r#"{"$setIntersection": ["$foo", ["hello", "world"]]}"#,
        ref_schema = Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double)
        ))))
    );
    test_derive_schema!(
        set_difference,
        expected = Ok(Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double),
        ))))),
        input = r#"{"$setDifference": ["$foo", [42, 53]]}"#,
        ref_schema = Schema::Array(Box::new(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Double)
        ))))
    );
}

mod group_ops {
    use super::*;
    test_derive_schema!(
        max,
        expected = Ok(Schema::Atomic(Atomic::MaxKey)),
        input = r#"{"$max": "$foo"}"#,
        ref_schema = Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Timestamp),
            Schema::Atomic(Atomic::MaxKey),
        })
    );
    test_derive_schema!(
        min,
        expected = Ok(Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        })),
        input = r#"{"$min": "$foo"}"#,
        ref_schema = Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Date),
            Schema::Atomic(Atomic::Timestamp),
            Schema::Atomic(Atomic::MaxKey),
        })
    );
}

mod constant_ops {
    use super::*;
    test_derive_schema!(
        no_ops,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ))),
        input = r#"{"$abs": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Decimal),
        ))
    );

    test_derive_schema!(
        constant_boolean,
        expected = Ok(Schema::Atomic(Atomic::Boolean)),
        input = r#"{"$eq": ["foo", "$bar"]}"#
    );

    test_derive_schema!(
        constant_int,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{"$strLenBytes": "hello world"}"#
    );

    test_derive_schema!(
        constant_integral,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ))),
        input = r#"{"$count": {}}"#
    );

    test_derive_schema!(
        constant_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{"$rand": {}}"#
    );

    test_derive_schema!(
        constant_array_int,
        expected = Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))),
        input = r#"{"$range": [ 0, "$distance", 25 ]}"#
    );

    test_derive_schema!(
        constant_string,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{"$substr": [ "$quarter", 2, -1 ]}"#
    );

    test_derive_schema!(
        constant_long,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = r#"{"$toHashedIndexKey": "$val"}"#
    );
}
mod conversion_ops {
    use super::*;
    test_type_conversion_op!(
        convert_integer,
        expected = Schema::Atomic(Atomic::Integer),
        op = "$toInt"
    );

    test_type_conversion_op!(
        convert_string,
        expected = Schema::Atomic(Atomic::String),
        op = "$toString"
    );

    test_type_conversion_op!(
        convert_long,
        expected = Schema::Atomic(Atomic::Long),
        op = "$toLong"
    );

    test_type_conversion_op!(
        convert_boolean,
        expected = Schema::Atomic(Atomic::Boolean),
        op = "$toBool"
    );

    test_type_conversion_op!(
        convert_date,
        expected = Schema::Atomic(Atomic::Date),
        op = "$toDate"
    );

    test_type_conversion_op!(
        convert_decimal,
        expected = Schema::Atomic(Atomic::Decimal),
        op = "$toDecimal"
    );

    test_type_conversion_op!(
        convert_double,
        expected = Schema::Atomic(Atomic::Double),
        op = "$toDouble"
    );

    test_type_conversion_op!(
        convert_object_id,
        expected = Schema::Atomic(Atomic::ObjectId),
        op = "$toObjectid"
    );
}
mod bit_ops {
    use super::*;

    test_derive_schema!(
        bitwise_op_long,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = r#"{"$bitAnd": [1, {"$numberLong": "1"}]}"#
    );
    test_derive_schema!(
        bitwise_op_int,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{"$bitAnd": [1, 1]}"#
    );
    test_derive_schema!(
        bitwise_op_integral,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ))),
        input = r#"{"$bitAnd": [1, "$foo"]}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
            Schema::Atomic(Atomic::Decimal),
        ))
    );
}
mod window_ops {
    use super::*;
    test_derive_schema!(
        window_func_decimal_or_null,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$covariancePop": [1, "$foo"]}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))
    );
    test_derive_schema!(
        window_func_double_or_null,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$covariancePop": [1, "$foo"]}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
        ))
    );
}
mod numeric_ops {
    use super::*;

    test_derive_schema!(
        math_op_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = r#"{"$log": "$foo"}"#,
        ref_schema = Schema::Atomic(Atomic::Decimal)
    );

    test_derive_schema!(
        math_op_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{"$log": [1, 2.1]}"#
    );

    test_derive_schema!(
        math_op_long,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{"$log": [1, {"$numberLong": "1"}]}"#
    );

    test_derive_schema!(
        math_op_int,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{"$log": [1, 2]}"#
    );

    test_derive_schema!(
        math_op_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$log": [null, 1]}"#
    );

    test_derive_schema!(
        math_op_nullish,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$log": "$foo"}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Decimal),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Null),
        ))
    );
    test_derive_schema!(
        multiply_integral,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ))),
        input = r#"{"$multiply": [1, 1]}"#
    );
    test_derive_schema!(
        multiply_long,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = r#"{"$multiply": [1, {"$numberLong": "1"}]}"#
    );
    test_derive_schema!(
        multiply_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = r#"{"$multiply": [1, "$foo"]}"#,
        ref_schema = Schema::Atomic(Atomic::Decimal)
    );
    test_derive_schema!(
        multiply_double_or_null,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))),
        input = r#"{"$multiply": [1, "$foo"]}"#,
        ref_schema = Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Null),
        ))
    );
    test_derive_schema!(
        pow_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = r#"{"$pow": [1, "$foo"]}"#,
        ref_schema = Schema::Atomic(Atomic::Decimal)
    );
    test_derive_schema!(
        pow_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$pow": [1, null]}"#
    );
    test_derive_schema!(
        pow_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{"$pow": [1, 2.0]}"#
    );
    test_derive_schema!(
        pow_long,
        expected = Ok(Schema::Atomic(Atomic::Long),),
        input = r#"{"$pow": [1, {"$numberLong": "1"}]}"#
    );
    test_derive_schema!(
        pow_integers,
        expected = Ok(Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        ))),
        input = r#"{"$pow": [1, 123]}"#
    );
    test_derive_schema!(
        mod_decimal,
        expected = Ok(Schema::Atomic(Atomic::Decimal)),
        input = r#"{"$mod": [1, "$foo"]}"#,
        ref_schema = Schema::Atomic(Atomic::Decimal)
    );
    test_derive_schema!(
        mod_null,
        expected = Ok(Schema::Atomic(Atomic::Null)),
        input = r#"{"$mod": [1, null]}"#
    );
    test_derive_schema!(
        mod_double,
        expected = Ok(Schema::Atomic(Atomic::Double)),
        input = r#"{"$mod": [1, 2.1]}"#
    );
    test_derive_schema!(
        mod_long,
        expected = Ok(Schema::Atomic(Atomic::Long)),
        input = r#"{"$mod": [3, {"$numberLong": "2"}]}"#
    );
    test_derive_schema!(
        mod_int,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{"$mod": [1, 123]}"#
    );
}

mod supremum_ops {
    use super::*;
    test_derive_schema!(
        max,
        expected = Ok(Schema::Atomic(Atomic::String)),
        input = r#"{"$max": ["$foo", 1, "hello"]}"#,
        ref_schema = Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        })
    );
    test_derive_schema!(
        min,
        expected = Ok(Schema::Atomic(Atomic::Integer)),
        input = r#"{"$min": ["$foo", 1, "hello"]}"#,
        ref_schema = Schema::AnyOf(set! {
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::String),
        })
    );
}
