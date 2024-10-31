mod negation {
    use crate::match_stage::NegativeNormalize;
    use agg_ast::definitions::{Expression, MatchExpression};

    macro_rules! test_negation {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                let input: MatchExpression = serde_json::from_str($input).unwrap();
                let expected: MatchExpression = serde_json::from_str($expected).unwrap();
                let result = input.get_negation();
                assert_eq!(result, expected);
            }
        };
    }

    macro_rules! test_expression_negation {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                let input: Expression = serde_json::from_str($input).unwrap();
                let expected: Expression = serde_json::from_str($expected).unwrap();
                let result = input.get_negation();
                assert_eq!(result, expected);
            }
        };
    }

    test_negation!(
        lt,
        expected = r#"{"$expr": {"$gte": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$lt": ["$foo", true]}}"#
    );

    test_negation!(
        lte,
        expected = r#"{"$expr": {"$gt": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$lte": ["$foo", true]}}"#
    );

    test_negation!(
        gt,
        expected = r#"{"$expr": {"$lte": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$gt": ["$foo", true]}}"#
    );

    test_negation!(
        gte,
        expected = r#"{"$expr": {"$lt": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$gte": ["$foo", true]}}"#
    );

    test_negation!(
        eq,
        expected = r#"{"$expr": {"$ne": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$eq": ["$foo", true]}}"#
    );

    test_negation!(
        ne,
        expected = r#"{"$expr": {"$eq": ["$foo", true]}}"#,
        input = r#"{"$expr": {"$ne": ["$foo", true]}}"#
    );

    test_negation!(
        and,
        expected = r#"{"$expr": {"$or": [{"$gt": ["$foo", 10]}, {"$lt": ["$foo", 5]}]}}"#,
        input = r#"{"$expr": {"$and": [{"$lte": ["$foo", 10]}, {"$gte": ["$foo", 5]}]}}"#
    );

    test_negation!(
        or,
        expected = r#"{"$expr": {"$and": [{"$gte": ["$foo", 10]}, {"$lte": ["$foo", 5]}]}}"#,
        input = r#"{"$expr": {"$or": [{"$lt": ["$foo", 10]}, {"$gt": ["$foo", 5]}]}}"#
    );

    // MatchField negation tests
    test_negation!(
        m_eq,
        expected = r#"{"x": {"$ne": 10}}"#,
        input = r#"{"x": 10}"#
    );
    test_negation!(
        m_gt,
        expected = r#"{"x": {"$lte": 10}}"#,
        input = r#"{"x": {"$gt": 10}}"#
    );
    test_negation!(
        m_gte,
        expected = r#"{"x": {"$lt": 10}}"#,
        input = r#"{"x": {"$gte": 10}}"#
    );
    test_negation!(
        m_lt,
        expected = r#"{"x": {"$gte": 10}}"#,
        input = r#"{"x": {"$lt": 10}}"#
    );
    test_negation!(
        m_lte,
        expected = r#"{"x": {"$gt": 10}}"#,
        input = r#"{"x": {"$lte": 10}}"#
    );
    test_negation!(
        m_ne,
        expected = r#"{"x": 10}"#,
        input = r#"{"x": {"$ne": 10}}"#
    );
    test_negation!(
        m_in,
        expected = r#"{"x": {"$nin": [10, "hello"]}}"#,
        input = r#"{"x": {"$in": [10, "hello"]}}"#
    );
    test_negation!(
        m_nin,
        expected = r#"{"x": {"$in": [10, "hello"]}}"#,
        input = r#"{"x": {"$nin": [10, "hello"]}}"#
    );
    test_negation!(
        m_exists_true,
        expected = r#"{"x": {"$exists": false}}"#,
        input = r#"{"x": {"$exists": true}}"#
    );
    test_negation!(
        m_exists_false,
        expected = r#"{"x": {"$exists": true}}"#,
        input = r#"{"x": {"$exists": false}}"#
    );
    test_negation!(
        m_exists_10,
        expected = r#"{"x": {"$exists": false}}"#,
        input = r#"{"x": {"$exists": 10}}"#
    );
    test_negation!(
        m_exists_0,
        expected = r#"{"x": {"$exists": true}}"#,
        input = r#"{"x": {"$exists": 0}}"#
    );
    test_negation!(
        m_exists_string,
        expected = r#"{"x": {"$exists": false}}"#,
        input = r#"{"x": {"$exists": "hello"}}"#
    );
    test_negation!(
        m_exists_array,
        expected = r#"{"x": {"$exists": false}}"#,
        input = r#"{"x": {"$exists": []}}"#
    );
    test_negation!(
        m_exists_object,
        expected = r#"{"x": {"$exists": false}}"#,
        input = r#"{"x": {"$exists": {}}}"#
    );
    test_negation!(
        m_exists_null,
        expected = r#"{"x": {"$exists": true}}"#,
        input = r#"{"x": {"$exists": null}}"#
    );
    test_negation!(
        m_type,
        expected = r#"{"x": {"$type": [
            "string",
            "object",
            "array",
            "binData",
            "undefined",
            "objectId",
            "bool",
            "date",
            "null",
            "regex",
            "dbPointer",
            "javascript",
            "symbol",
            "javascriptWithScope",
            "timestamp",
            "long",
            "decimal",
            "minKey",
            "maxKey"
        ]}}"#,
        input = r#"{"x": {"$type": ["int", "double"]}}"#
    );
    test_negation!(
        m_size,
        expected = r#"{"x": {"$not": {"$size": 5}}}"#,
        input = r#"{"x": {"$size": 5}}"#
    );
    test_negation!(
        m_mod,
        expected = r#"{"x": {"$not": {"$mod": [3, 5]}}}"#,
        input = r#"{"x": {"$mod": [3, 5]}}"#
    );
    test_negation!(
        m_bits_any_set,
        expected = r#"{"x": {"$bitsAllClear": "$x"}}"#,
        input = r#"{"x": {"$bitsAnySet": "$x"}}"#
    );
    test_negation!(
        m_bits_all_set,
        expected = r#"{"x": {"$bitsAnyClear": "$x"}}"#,
        input = r#"{"x": {"$bitsAllSet": "$x"}}"#
    );
    test_negation!(
        m_bits_any_clear,
        expected = r#"{"x": {"$bitsAllSet": "$x"}}"#,
        input = r#"{"x": {"$bitsAnyClear": "$x"}}"#
    );
    test_negation!(
        m_bits_all_clear,
        expected = r#"{"x": {"$bitsAnySet": "$x"}}"#,
        input = r#"{"x": {"$bitsAllClear": "$x"}}"#
    );
    test_negation!(
        m_all,
        expected = r#"{"x": {"$not": {"$all": [5, 4]}}}"#,
        input = r#"{"x": {"$all": [5, 4]}}"#
    );
    test_negation!(
        m_geo_intersects,
        expected = r#"{"x": {"$not": {"$geoIntersects": {
            "$geometry": {
                "type": "Polygon" ,
                "coordinates" :  [[ 0, 0 ], [ 3, 6 ], [ 6, 1 ], [ 0, 0 ]]
            }
        }}}}"#,
        input = r#"{"x": {"$geoIntersects": {
            "$geometry": {
                "type": "Polygon" ,
                "coordinates" :  [[ 0, 0 ], [ 3, 6 ], [ 6, 1 ], [ 0, 0 ]]
            }
        }}}"#
    );
    test_negation!(
        m_geo_within,
        expected = r#"{"x": {"$not": {"$geoWithin": {
            "$geometry": {
                "type": "Polygon" ,
                "coordinates" :  [[ 0, 0 ], [ 3, 6 ], [ 6, 1 ], [ 0, 0 ]]
            }
        }}}}"#,
        input = r#"{"x": {"$geoWithin": {
            "$geometry": {
                "type": "Polygon" ,
                "coordinates" :  [[ 0, 0 ], [ 3, 6 ], [ 6, 1 ], [ 0, 0 ]]
            }
        }}}"#
    );
    test_negation!(
        m_near,
        expected = r#"{"x": {"$not": {"$near": {
            "$geometry": {
                "type": "Polygon" ,
                "coordinates" :  [[ 0, 0 ], [ 3, 6 ], [ 6, 1 ], [ 0, 0 ]]
            }
        }}}}"#,
        input = r#"{"x": {"$near": {
            "$geometry": {
                "type": "Polygon" ,
                "coordinates" :  [[ 0, 0 ], [ 3, 6 ], [ 6, 1 ], [ 0, 0 ]]
            }
        }}}"#
    );
    test_negation!(
        m_near_sphere,
        expected = r#"{"x": {"$not": {"$nearSphere": {
            "$geometry": {
                "type": "Polygon" ,
                "coordinates" :  [[ 0, 0 ], [ 3, 6 ], [ 6, 1 ], [ 0, 0 ]]
            }
        }}}}"#,
        input = r#"{"x": {"$nearSphere": {
            "$geometry": {
                "type": "Polygon" ,
                "coordinates" :  [[ 0, 0 ], [ 3, 6 ], [ 6, 1 ], [ 0, 0 ]]
            }
        }}}"#
    );
    test_negation!(
        m_multi_operators,
        expected = r#"{"$or": [
            {"x": {"$eq": 10}},
            {"x": {"$lte": 5}},
            {"x": {"$gt": 500}},
            {"x": {"$exists": false}},
            {"x": {"$not": {"$size": 3}}}
        ]}"#,
        input = r#"{"x": {
                "$ne": 10,
                "$gt": 5,
                "$lte": 500,
                "$exists": true,
                "$size": 3
            }
        }"#
    );

    test_expression_negation!(literal, expected = r#"2"#, input = r#"2"#);

    test_expression_negation!(array, expected = r#"[2]"#, input = r#"[2]"#);

    test_expression_negation!(document, expected = r#"{"a": 2}"#, input = r#"{"a": 2}"#);

    test_expression_negation!(
        field_ref,
        expected = r#"{"$or": [{"$lte": ["$foo", null]}, {"$eq": ["$foo", 0]}]}"#,
        input = r#""$foo""#
    );
}

mod negative_normal_form {
    use crate::match_stage::NegativeNormalize;
    use agg_ast::definitions::MatchExpression;

    macro_rules! test_nnf {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                let input: MatchExpression = serde_json::from_str($input).unwrap();
                let expected: MatchExpression = serde_json::from_str($expected).unwrap();
                let result = input.get_negative_normal_form();
                assert_eq!(result, expected);
            }
        };
    }

    test_nnf!(
        expr_func_noop,
        expected = r#"{"$expr": {"$eq": ["$foo", 0]}}"#,
        input = r#"{"$expr": {"$eq": ["$foo", 0]}}"#
    );
}
