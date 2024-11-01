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

test_negation!(
    array_to_object,
    expected = r#"{"$expr": {"$lte": [{"$arrayToObject": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$arrayToObject": "$x"}}"#
);

test_negation!(
    first,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$first": "$x"}, null]}, {"$eq": [{"$first": "$x"}, 0]}, {"$eq": [{"$first": "$x"}, false]}]}}"#,
    input = r#"{"$expr": {"$first": "$x"}}"#
);

test_negation!(
    last,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$last": "$x"}, null]}, {"$eq": [{"$last": "$x"}, 0]}, {"$eq": [{"$last": "$x"}, false]}]}}"#,
    input = r#"{"$expr": {"$last": "$x"}}"#
);

test_negation!(
    object_to_array,
    expected = r#"{"$expr": {"$lte": [{"$objectToArray": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$objectToArray": "$x"}}"#
);

test_negation!(
    reverse_array,
    expected = r#"{"$expr": {"$lte": [{"$reverseArray": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$reverseArray": "$x"}}"#
);

test_negation!(
    to_date,
    expected = r#"{"$expr": {"$lte": [{"$toDate": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$toDate": "$x"}}"#
);

test_negation!(
    to_object_id,
    expected = r#"{"$expr": {"$lte": [{"$toObjectId": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$toObjectId": "$x"}}"#
);

test_negation!(
    to_string,
    expected = r#"{"$expr": {"$lte": [{"$toString": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$toString": "$x"}}"#
);

test_negation!(
    ts_second,
    expected = r#"{"$expr": {"$lte": [{"$tsSecond": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$tsSecond": "$x"}}"#
);

test_negation!(
    ts_increment,
    expected = r#"{"$expr": {"$lte": [{"$tsIncrement": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$tsIncrement": "$x"}}"#
);

test_negation!(
    concat,
    expected = r#"{"$expr": {"$lte": [{"$concat": ["$x", "$y", "$z"]}, null]}}"#,
    input = r#"{"$expr": {"$concat": ["$x", "$y", "$z"]}}"#
);

test_negation!(
    concat_arrays,
    expected = r#"{"$expr": {"$lte": [{"$concatArrays": ["$x", "$y"]}, null]}}"#,
    input = r#"{"$expr": {"$concatArrays": ["$x", "$y"]}}"#
);

test_negation!(
    set_difference,
    expected = r#"{"$expr": {"$lte": [{"$setDifference": ["$x", "$y"]}, null]}}"#,
    input = r#"{"$expr": {"$setDifference": ["$x", "$y"]}}"#
);

test_negation!(
    set_intersection,
    expected = r#"{"$expr": {"$lte": [{"$setIntersection": ["$x", "$y", "$z"]}, null]}}"#,
    input = r#"{"$expr": {"$setIntersection": ["$x", "$y", "$z"]}}"#
);

test_negation!(
    set_union,
    expected = r#"{"$expr": {"$lte": [{"$setUnion": ["$x", "$y"]}, null]}}"#,
    input = r#"{"$expr": {"$setUnion": ["$x", "$y"]}}"#
);

test_negation!(
    slice,
    expected = r#"{"$expr": {"$lte": [{"$slice": ["$x", "$y", "$z"]}, null]}}"#,
    input = r#"{"$expr": {"$slice": ["$x", "$y", "$z"]}}"#
);

test_negation!(
    split,
    expected = r#"{"$expr": {"$lte": [{"$split": ["$x", "$y"]}, null]}}"#,
    input = r#"{"$expr": {"$split": ["$x", "$y"]}}"#
);
