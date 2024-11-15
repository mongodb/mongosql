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

test_negation!(
    get_field,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$getField": {"input": "$x", "field": "foo"}}, null]}, {"$eq": [{"$getField": {"input": "$x", "field": "foo"}}, 0]}, {"$eq": [{"$getField": {"input": "$x", "field": "foo"}}, false]}]}}"#,
    input = r#"{"$expr": {"$getField": {"input": "$x", "field": "foo"}}}"#
);

test_negation!(
    reduce,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$reduce": {"input": "$x", "initialValue": {"sum": 1}, "in": {"sum": { "$add" : ["$$value.sum", "$$this"] }}}}, null]}, {"$eq": [{"$reduce": {"input": "$x", "initialValue": {"sum": 1}, "in": {"sum": { "$add" : ["$$value.sum", "$$this"] }}}}, 0]}, {"$eq": [{"$reduce": {"input": "$x", "initialValue": {"sum": 1}, "in": {"sum": { "$add" : ["$$value.sum", "$$this"] }}}}, false]}]}}"#,
    input = r#"{"$expr": {"$reduce": {"input": "$x", "initialValue": {"sum": 1}, "in": {"sum": { "$add" : ["$$value.sum", "$$this"] }}}}}"#
);

test_negation!(
    set_field,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$setField": {"input": "$x", "field": "foo", "value": "bar"}}, null]}, {"$eq": [{"$setField": {"input": "$x", "field": "foo", "value": "bar"}}, 0]}, {"$eq": [{"$setField": {"input": "$x", "field": "foo", "value": "bar"}}, false]}]}}"#,
    input = r#"{"$expr": {"$setField": {"input": "$x", "field": "foo", "value": "bar"}}}"#
);

test_negation!(
    switch,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$switch": {"branches": [{ "case": { "$eq": [ 0, 5 ] }, "then": "equals" }], "default": "Did not match"}}, null]}, {"$eq": [{"$switch": {"branches": [{ "case": { "$eq": [ 0, 5 ] }, "then": "equals" }], "default": "Did not match"}}, 0]}, {"$eq": [{"$switch": {"branches": [{ "case": { "$eq": [ 0, 5 ] }, "then": "equals" }], "default": "Did not match"}}, false]}]}}"#,
    input = r#"{"$expr": {"$switch": {"branches": [{ "case": { "$eq": [ 0, 5 ] }, "then": "equals" }], "default": "Did not match"}}}"#
);

test_negation!(
    unset_field,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$unsetField": {"input": "$x", "field": "foo"}}, null]}, {"$eq": [{"$unsetField": {"input": "$x", "field": "foo"}}, 0]}, {"$eq": [{"$unsetField": {"input": "$x", "field": "foo"}}, false]}]}}"#,
    input = r#"{"$expr": {"$unsetField": {"input": "$x", "field": "foo"}}}"#
);

test_negation!(
    first_n,
    expected = r#"{"$expr": {"$lte": [{"$firstN": {"input": "$x", "n": 2}}, null]}}"#,
    input = r#"{"$expr": {"$firstN": {"input": "$x", "n": 2}}}"#
);

test_negation!(
    last_n,
    expected = r#"{"$expr": {"$lte": [{"$lastN": {"input": "$x", "n": 2}}, null]}}"#,
    input = r#"{"$expr": {"$lastN": {"input": "$x", "n": 2}}}"#
);

test_negation!(
    max_n_array_element,
    expected = r#"{"$expr": {"$lte": [{"$maxN": {"input": "$x", "n": 2}}, null]}}"#,
    input = r#"{"$expr": {"$maxN": {"input": "$x", "n": 2}}}"#
);

test_negation!(
    min_n_array_element,
    expected = r#"{"$expr": {"$lte": [{"$minN": {"input": "$x", "n": 2}}, null]}}"#,
    input = r#"{"$expr": {"$minN": {"input": "$x", "n": 2}}}"#
);

test_negation!(
    let_op,
    expected = r#"{"$expr": {"$let": {"vars": {"a": 1, "b": 2}, "in": {"$ne": ["$$a", "$$b"]}}}}"#,
    input = r#"{"$expr": {"$let": {"vars": {"a": 1, "b": 2}, "in": {"$eq": ["$$a", "$$b"]}}}}"#
);

test_negation!(
    day_of_month,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$dayOfMonth": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$dayOfMonth": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$dayOfMonth": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    day_of_week,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$dayOfWeek": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$dayOfWeek": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$dayOfWeek": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    day_of_year,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$dayOfYear": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$dayOfYear": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$dayOfYear": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    hour,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$hour": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$hour": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$hour": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    millisecond,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$millisecond": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$millisecond": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$millisecond": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    minute,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$minute": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$minute": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$minute": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    month,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$month": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$month": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$month": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    second,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$second": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$second": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$second": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    week,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$week": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$week": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$week": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    year,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$year": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$year": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$year": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    iso_day_of_week,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$isoDayOfWeek": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$isoDayOfWeek": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$isoDayOfWeek": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    iso_week,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$isoWeek": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$isoWeek": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$isoWeek": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    iso_week_year,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$isoWeekYear": {"date": "$x", "timezone": "$y"}}, null]}, {"$eq": [{"$isoWeekYear": {"date": "$x", "timezone": "$y"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$isoWeekYear": {"date": "$x", "timezone": "$y"}}}"#
);

test_negation!(
    median,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$median": {"input": "$x", "method": "approximate"}}, null]}, {"$eq": [{"$median": {"input": "$x", "method": "approximate"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$median": {"input": "$x", "method": "approximate"}}}"#
);

test_negation!(
    percentile,
    expected = r#"{"$expr": {"$or": [{"$lte": [{"$percentile": {"input": "$x", "p": ["$y", "$z"], "method": "approximate"}}, null]}, {"$eq": [{"$percentile": {"input": "$x", "p": ["$y", "$z"], "method": "approximate"}}, 0]}]}}"#,
    input = r#"{"$expr": {"$percentile": {"input": "$x", "p": ["$y", "$z"], "method": "approximate"}}}"#
);
