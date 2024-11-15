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
    if_null,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$ifNull": ["$x", "foo"]}, null]}, {"$eq": [{"$ifNull": ["$x", "foo"]}, 0]}, {"$eq": [{"$ifNull": ["$x", "foo"]}, false]}]}}"#,
    input = r#"{"$expr": {"$ifNull": ["$x", "foo"]}}"#
);

test_negation!(
    last,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$last": "$x"}, null]}, {"$eq": [{"$last": "$x"}, 0]}, {"$eq": [{"$last": "$x"}, false]}]}}"#,
    input = r#"{"$expr": {"$last": "$x"}}"#
);

test_negation!(
    literal,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$literal": "$foo"}, null]}, {"$eq": [{"$literal": "$foo"}, 0]}, {"$eq": [{"$literal": "$foo"}, false]}]}}"#,
    input = r#"{"$expr": {"$literal": "$foo"}}"#
);

test_negation!(
    max,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$max": ["$x", "foo"]}, null]}, {"$eq": [{"$max": ["$x", "foo"]}, 0]}, {"$eq": [{"$max": ["$x", "foo"]}, false]}]}}"#,
    input = r#"{"$expr": {"$max": ["$x", "foo"]}}"#
);

test_negation!(
    min,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$min": "$x"}, null]}, {"$eq": [{"$min": "$x"}, 0]}, {"$eq": [{"$min": "$x"}, false]}]}}"#,
    input = r#"{"$expr": {"$min": "$x"}}"#
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

test_negation!(
    meta,
    expected = r#"{"$expr": {"$lte": [{"$meta": "textScore"}, null]}}"#,
    input = r#"{"$expr": {"$meta": "textScore"}}"#
);

test_negation!(
    merge_objects,
    expected = r#"{"$expr": {"$lte": [{"$mergeObjects": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$mergeObjects": "$x"}}"#
);

test_negation!(
    rand,
    expected = r#"{"$expr": {"$lte": [{"$rand": {}}, null]}}"#,
    input = r#"{"$expr": {"$rand": {}}}"#
);

test_negation!(
    range,
    expected = r#"{"$expr": {"$lte": [{"$range": [1, 2, 3]}, null]}}"#,
    input = r#"{"$expr": {"$range": [1,2,3]}}"#
);

test_negation!(
    substr,
    expected = r#"{"$expr": {"$lte": [{"$substr": ["$x", 1, 2]}, null]}}"#,
    input = r#"{"$expr": {"$substr": ["$x", 1, 2]}}"#
);

test_negation!(
    substr_bytes,
    expected = r#"{"$expr": {"$lte": [{"$substrBytes": ["$x", 1, 2]}, null]}}"#,
    input = r#"{"$expr": {"$substrBytes": ["$x", 1, 2]}}"#
);

test_negation!(
    substr_cp,
    expected = r#"{"$expr": {"$lte": [{"$substrCP": ["$x", 1, 2]}, null]}}"#,
    input = r#"{"$expr": {"$substrCP": ["$x", 1, 2]}}"#
);

test_negation!(
    to_hashed_index_key,
    expected = r#"{"$expr": {"$lte": [{"$toHashedIndexKey": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$toHashedIndexKey": "$x"}}"#
);

test_negation!(
    to_lower,
    expected = r#"{"$expr": {"$lte": [{"$toLower": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$toLower": "$x"}}"#
);

test_negation!(
    to_upper,
    expected = r#"{"$expr": {"$lte": [{"$toUpper": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$toUpper": "$x"}}"#
);

test_negation!(
    type_of,
    expected = r#"{"$expr": {"$lte": [{"$type": "$x"}, null]}}"#,
    input = r#"{"$expr": {"$type": "$x"}}"#
);

test_negation!(
    not,
    expected = r#"{"$expr": {"$eq": ["$x", "$y"]}}"#,
    input = r#"{"$expr": {"$not": {"$eq": ["$x", "$y"]}}}"#
);

test_negation!(
    all_elements_true,
    expected = r#"{"$expr": {"$allElementsTrue": "$x"}}"#,
    input = r#"{"$expr": {"$allElementsTrue": "$x"}}"#
);

test_negation!(
    any_element_true,
    expected = r#"{"$expr": {"$anyElementTrue": "$x"}}"#,
    input = r#"{"$expr": {"$anyElementTrue": "$x"}}"#
);

test_negation!(
    cmp,
    expected = r#"{"$expr": {"$cmp": ["$x", "$y"]}}"#,
    input = r#"{"$expr": {"$cmp": ["$x", "$y"]}}"#
);

test_negation!(
    in_,
    expected = r#"{"$expr": {"$in": ["$x", [1, 2, 3]]}}"#,
    input = r#"{"$expr": {"$in": ["$x", [1, 2, 3]]}}"#
);

test_negation!(
    size,
    expected = r#"{"$expr": {"$size": "$x"}}"#,
    input = r#"{"$expr": {"$size": "$x"}}"#
);

test_negation!(
    str_len_bytes,
    expected = r#"{"$expr": {"$strLenBytes": "$x"}}"#,
    input = r#"{"$expr": {"$strLenBytes": "$x"}}"#
);

test_negation!(
    str_len_cp,
    expected = r#"{"$expr": {"$strLenCP": "$x"}}"#,
    input = r#"{"$expr": {"$strLenCP": "$x"}}"#
);

test_negation!(
    strcasecmp,
    expected = r#"{"$expr": {"$strcasecmp": ["$x", "$y"]}}"#,
    input = r#"{"$expr": {"$strcasecmp": ["$x", "$y"]}}"#
);

test_negation!(
    set_equals,
    expected = r#"{"$expr": {"$setEquals": ["$x", "$y"]}}"#,
    input = r#"{"$expr": {"$setEquals": ["$x", "$y"]}}"#
);

test_negation!(
    set_is_subset,
    expected = r#"{"$expr": {"$setIsSubset": ["$x", "$y"]}}"#,
    input = r#"{"$expr": {"$setIsSubset": ["$x", "$y"]}}"#
);

test_negation!(
    sum,
    expected = r#"{"$expr": {"$sum": ["$x", "$y"]}}"#,
    input = r#"{"$expr": {"$sum": ["$x", "$y"]}}"#
);

test_negation!(
    to_bool,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$toBool": "$x"}, null]}, {"$eq": [{"$toBool": "$x"}, false]}]}}"#,
    input = r#"{"$expr": {"$toBool": "$x"}}"#
);

test_negation!(
    is_array,
    expected = r#"{"$expr": {"$eq": [{"$isArray": "$x"}, false]}}"#,
    input = r#"{"$expr": {"$isArray": "$x"}}"#
);

test_negation!(
    is_number,
    expected = r#"{"$expr": {"$eq": [{"$isNumber": "$x"}, false]}}"#,
    input = r#"{"$expr": {"$isNumber": "$x"}}"#
);

test_negation!(
    abs,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$abs": ["$x"]}, null]}, {"$eq": [{"$abs": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$abs": ["$x"]}}"#
);

test_negation!(
    acos,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$acos": ["$x"]}, null]}, {"$eq": [{"$acos": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$acos": ["$x"]}}"#
);

test_negation!(
    acosh,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$acosh": ["$x"]}, null]}, {"$eq": [{"$acosh": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$acosh": ["$x"]}}"#
);

test_negation!(
    asin,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$asin": ["$x"]}, null]}, {"$eq": [{"$asin": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$asin": ["$x"]}}"#
);

test_negation!(
    asinh,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$asinh": ["$x"]}, null]}, {"$eq": [{"$asinh": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$asinh": ["$x"]}}"#
);

test_negation!(
    atan,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$atan": ["$x"]}, null]}, {"$eq": [{"$atan": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$atan": ["$x"]}}"#
);

test_negation!(
    atan2,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$atan2": ["$x"]}, null]}, {"$eq": [{"$atan2": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$atan2": ["$x"]}}"#
);

test_negation!(
    atanh,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$atanh": ["$x"]}, null]}, {"$eq": [{"$atanh": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$atanh": ["$x"]}}"#
);

test_negation!(
    avg,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$avg": ["$x"]}, null]}, {"$eq": [{"$avg": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$avg": ["$x"]}}"#
);

test_negation!(
    cos,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$cos": ["$x"]}, null]}, {"$eq": [{"$cos": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$cos": ["$x"]}}"#
);

test_negation!(
    cosh,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$cosh": ["$x"]}, null]}, {"$eq": [{"$cosh": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$cosh": ["$x"]}}"#
);

test_negation!(
    degrees_to_radians,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$degreesToRadians": ["$x"]}, null]}, {"$eq": [{"$degreesToRadians": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$degreesToRadians": ["$x"]}}"#
);

test_negation!(
    divide,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$divide": ["$x"]}, null]}, {"$eq": [{"$divide": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$divide": ["$x"]}}"#
);

test_negation!(
    exp,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$exp": ["$x"]}, null]}, {"$eq": [{"$exp": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$exp": ["$x"]}}"#
);

test_negation!(
    ln,
    expected =
        r#"{"$expr": {"$or": [{"$lte": [{"$ln": ["$x"]}, null]}, {"$eq": [{"$ln": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$ln": ["$x"]}}"#
);

test_negation!(
    log,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$log": ["$x"]}, null]}, {"$eq": [{"$log": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$log": ["$x"]}}"#
);

test_negation!(
    log10,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$log10": ["$x"]}, null]}, {"$eq": [{"$log10": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$log10": ["$x"]}}"#
);

test_negation!(
    r#mod,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$mod": ["$x"]}, null]}, {"$eq": [{"$mod": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$mod": ["$x"]}}"#
);

test_negation!(
    multiply,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$multiply": ["$x"]}, null]}, {"$eq": [{"$multiply": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$multiply": ["$x"]}}"#
);

test_negation!(
    pow,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$pow": ["$x"]}, null]}, {"$eq": [{"$pow": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$pow": ["$x"]}}"#
);

test_negation!(
    radians_to_degrees,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$radiansToDegrees": ["$x"]}, null]}, {"$eq": [{"$radiansToDegrees": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$radiansToDegrees": ["$x"]}}"#
);

test_negation!(
    sin,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$sin": ["$x"]}, null]}, {"$eq": [{"$sin": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$sin": ["$x"]}}"#
);

test_negation!(
    sinh,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$sinh": ["$x"]}, null]}, {"$eq": [{"$sinh": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$sinh": ["$x"]}}"#
);

test_negation!(
    sqrt,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$sqrt": ["$x"]}, null]}, {"$eq": [{"$sqrt": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$sqrt": ["$x"]}}"#
);

test_negation!(
    tan,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$tan": ["$x"]}, null]}, {"$eq": [{"$tan": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$tan": ["$x"]}}"#
);

test_negation!(
    tanh,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$tanh": ["$x"]}, null]}, {"$eq": [{"$tanh": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$tanh": ["$x"]}}"#
);

test_negation!(
    trunc,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$trunc": ["$x"]}, null]}, {"$eq": [{"$trunc": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$trunc": ["$x"]}}"#
);

test_negation!(
    ceil,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$ceil": ["$x"]}, null]}, {"$eq": [{"$ceil": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$ceil": ["$x"]}}"#
);

test_negation!(
    floor,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$floor": ["$x"]}, null]}, {"$eq": [{"$floor": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$floor": ["$x"]}}"#
);

test_negation!(
    index_of_array,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$indexOfArray": ["$x"]}, null]}, {"$eq": [{"$indexOfArray": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$indexOfArray": ["$x"]}}"#
);

test_negation!(
    index_of_bytes,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$indexOfBytes": ["$x"]}, null]}, {"$eq": [{"$indexOfBytes": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$indexOfBytes": ["$x"]}}"#
);

test_negation!(
    index_of_cp,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$indexOfCP": ["$x"]}, null]}, {"$eq": [{"$indexOfCP": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$indexOfCP": ["$x"]}}"#
);

test_negation!(
    to_int,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$toInt": ["$x"]}, null]}, {"$eq": [{"$toInt": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$toInt": ["$x"]}}"#
);

test_negation!(
    add,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$add": ["$x"]}, null]}, {"$eq": [{"$add": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$add": ["$x"]}}"#
);

test_negation!(
    subtract,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$subtract": ["$x"]}, null]}, {"$eq": [{"$subtract": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$subtract": ["$x"]}}"#
);

test_negation!(
    array_elem_at,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$arrayElemAt": ["$x"]}, null]}, {"$eq": [{"$arrayElemAt": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$arrayElemAt": ["$x"]}}"#
);

test_negation!(
    binary_size,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$binarySize": ["$x"]}, null]}, {"$eq": [{"$binarySize": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$binarySize": ["$x"]}}"#
);

test_negation!(
    bit_and,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$bitAnd": ["$x"]}, null]}, {"$eq": [{"$bitAnd": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$bitAnd": ["$x"]}}"#
);

test_negation!(
    bit_not,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$bitNot": ["$x"]}, null]}, {"$eq": [{"$bitNot": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$bitNot": ["$x"]}}"#
);

test_negation!(
    bit_or,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$bitOr": ["$x"]}, null]}, {"$eq": [{"$bitOr": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$bitOr": ["$x"]}}"#
);

test_negation!(
    bit_xor,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$bitXor": ["$x"]}, null]}, {"$eq": [{"$bitXor": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$bitXor": ["$x"]}}"#
);

test_negation!(
    bson_size,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$bsonSize": ["$x"]}, null]}, {"$eq": [{"$bsonSize": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$bsonSize": ["$x"]}}"#
);

test_negation!(
    covariance_pop,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$covariancePop": ["$x"]}, null]}, {"$eq": [{"$covariancePop": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$covariancePop": ["$x"]}}"#
);

test_negation!(
    covariance_samp,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$covarianceSamp": ["$x"]}, null]}, {"$eq": [{"$covarianceSamp": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$covarianceSamp": ["$x"]}}"#
);

test_negation!(
    std_dev_pop,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$stdDevPop": ["$x"]}, null]}, {"$eq": [{"$stdDevPop": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$stdDevPop": ["$x"]}}"#
);

test_negation!(
    std_dev_samp,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$stdDevSamp": ["$x"]}, null]}, {"$eq": [{"$stdDevSamp": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$stdDevSamp": ["$x"]}}"#
);

test_negation!(
    round,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$round": ["$x"]}, null]}, {"$eq": [{"$round": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$round": ["$x"]}}"#
);

test_negation!(
    to_decimal,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$toDecimal": ["$x"]}, null]}, {"$eq": [{"$toDecimal": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$toDecimal": ["$x"]}}"#
);

test_negation!(
    to_double,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$toDouble": ["$x"]}, null]}, {"$eq": [{"$toDouble": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$toDouble": ["$x"]}}"#
);

test_negation!(
    to_long,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$toLong": ["$x"]}, null]}, {"$eq": [{"$toLong": ["$x"]}, 0]}]}}"#,
    input = r#"{"$expr": {"$toLong": ["$x"]}}"#
);
