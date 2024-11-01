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
