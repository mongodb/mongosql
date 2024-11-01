test_negation!(
    date_add,
    expected = r#"{"$expr": {"$lte": [{"$dateAdd": {"startDate": "$x", "unit": "$y", "amount": "$z", "timezone": "$a"}}, null]}}"#,
    input = r#"{"$expr": {"$dateAdd": {"startDate": "$x", "unit": "$y", "amount": "$z", "timezone": "$a"}}}"#
);

test_negation!(
    date_diff,
    expected = r#"{"$expr": {"$lte": [{"$dateDiff": {"startDate": "$x", "endDate": "$y", "unit": "$z", "timezone": "$a", "startOfWeek": "$b"}}, null]}}"#,
    input = r#"{"$expr": {"$dateDiff": {"startDate": "$x", "endDate": "$y", "unit": "$z", "timezone": "$a", "startOfWeek": "$b"}}}"#
);

test_negation!(
    date_from_string,
    expected = r#"{"$expr": {"$lte": [{"$dateFromString": {"dateString": "$x", "format": "$y", "timezone": "$z", "onError": "$a", "onNull": "$b"}}, null]}}"#,
    input = r#"{"$expr": {"$dateFromString": {"dateString": "$x", "format": "$y", "timezone": "$z", "onError": "$a", "onNull": "$b"}}}"#
);

test_negation!(
    date_from_parts,
    expected = r#"{"$expr": {"$lte": [{"$dateFromParts": {"year": "$x", "month": "$y", "day": "$z", "hour": "$a", "minute": "$b", "second": "$c", "millisecond": "$n", "timezone": "$m"}}, null]}}"#,
    input = r#"{"$expr": {"$dateFromParts": {"year": "$x", "month": "$y", "day": "$z", "hour": "$a", "minute": "$b", "second": "$c", "millisecond": "$n", "timezone": "$m"}}}"#
);

test_negation!(
    date_subtract,
    expected = r#"{"$expr": {"$lte": [{"$dateSubtract": {"startDate": "$x", "unit": "$y", "amount": "$z", "timezone": "$a"}}, null]}}"#,
    input = r#"{"$expr": {"$dateSubtract": {"startDate": "$x", "unit": "$y", "amount": "$z", "timezone": "$a"}}}"#
);

test_negation!(
    date_trunc,
    expected = r#"{"$expr": {"$lte": [{"$dateTrunc": {"date": "$x", "unit": "$y", "binSize": "$z", "timezone": "$a", "startOfWeek": "$b"}}, null]}}"#,
    input = r#"{"$expr": {"$dateTrunc": {"date": "$x", "unit": "$y", "binSize": "$z", "timezone": "$a", "startOfWeek": "$b"}}}"#
);

test_negation!(
    date_to_parts,
    expected = r#"{"$expr": {"$lte": [{"$dateToParts": {"date": "$x", "timezone": "$y", "iso8601": true}}, null]}}"#,
    input = r#"{"$expr": {"$dateToParts": {"date": "$x", "timezone": "$y", "iso8601": true}}}"#
);

test_negation!(
    date_to_string,
    expected = r#"{"$expr": {"$lte": [{"$dateToString": {"date": "$x", "format": "$y", "timezone": "$z", "onNull": "$a"}}, null]}}"#,
    input = r#"{"$expr": {"$dateToString": {"date": "$x", "format": "$y", "timezone": "$z", "onNull": "$a"}}}"#
);

test_negation!(
    filter,
    expected = r#"{"$expr": {"$lte": [{"$filter": {"input": "$x", "as": "this", "cond": "$y", "limit": "$z"}}, null]}}"#,
    input = r#"{"$expr": {"$filter": {"input": "$x", "as": "this", "cond": "$y", "limit": "$z"}}}"#
);

test_negation!(
    ltrim,
    expected = r#"{"$expr": {"$lte": [{"$ltrim": {"input": "$x", "chars": "$y"}}, null]}}"#,
    input = r#"{"$expr": {"$ltrim": {"input": "$x", "chars": "$y"}}}"#
);

test_negation!(
    map,
    expected =
        r#"{"$expr": {"$lte": [{"$map": {"input": "$x", "as": "this", "in": "$y"}}, null]}}"#,
    input = r#"{"$expr": {"$map": {"input": "$x", "as": "this", "in": "$y"}}}"#
);

test_negation!(
    regex_find,
    expected = r#"{"$expr": {"$lte": [{"$regexFind": {"input": "$x", "regex": "$y", "options": "$z"}}, null]}}"#,
    input = r#"{"$expr": {"$regexFind": {"input": "$x", "regex": "$y", "options": "$z"}}}"#
);

test_negation!(
    regex_find_all,
    expected = r#"{"$expr": {"$lte": [{"$regexFindAll": {"input": "$x", "regex": "$y", "options": "$z"}}, null]}}"#,
    input = r#"{"$expr": {"$regexFindAll": {"input": "$x", "regex": "$y", "options": "$z"}}}"#
);

test_negation!(
    regex_match,
    expected = r#"{"$expr": {"$eq": [{"$regexMatch": {"input": "$x", "regex": "$y", "options": "$z"}}, false]}}"#,
    input = r#"{"$expr": {"$regexMatch": {"input": "$x", "regex": "$y", "options": "$z"}}}"#
);

test_negation!(
    replace_one,
    expected = r#"{"$expr": {"$lte": [{"$replaceOne": {"input": "$x", "find": "$y", "replacement": "$z"}}, null]}}"#,
    input = r#"{"$expr": {"$replaceOne": {"input": "$x", "find": "$y", "replacement": "$z"}}}"#
);

test_negation!(
    replace_all,
    expected = r#"{"$expr": {"$lte": [{"$replaceAll": {"input": "$x", "find": "$y", "replacement": "$z"}}, null]}}"#,
    input = r#"{"$expr": {"$replaceAll": {"input": "$x", "find": "$y", "replacement": "$z"}}}"#
);

test_negation!(
    rtrim,
    expected = r#"{"$expr": {"$lte": [{"$rtrim": {"input": "$x", "chars": "$y"}}, null]}}"#,
    input = r#"{"$expr": {"$rtrim": {"input": "$x", "chars": "$y"}}}"#
);

test_negation!(
    sort_array,
    expected = r#"{"$expr": {"$lte": [{"$sortArray": {"input": "$x", "sortBy": 1}}, null]}}"#,
    input = r#"{"$expr": {"$sortArray": {"input": "$x", "sortBy": 1}}}"#
);

test_negation!(
    trim,
    expected = r#"{"$expr": {"$lte": [{"$trim": {"input": "$x", "chars": "$y"}}, null]}}"#,
    input = r#"{"$expr": {"$trim": {"input": "$x", "chars": "$y"}}}"#
);

test_negation!(
    zip,
    expected = r#"{"$expr": {"$lte": [{"$zip": {"inputs": ["$x", "$y"], "useLongestLength": true, "defaults": [1, 2]}}, null]}}"#,
    input = r#"{"$expr": {"$zip": {"inputs": ["$x", "$y"], "useLongestLength": true, "defaults": [1, 2]}}}"#
);
