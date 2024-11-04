test_negation!(
    m_or,
    expected = r#"{"$and": [{"foo": {"$ne": 10}}, {"bar": {"$eq": 5}}]}"#,
    input = r#"{"$or": [{"foo": {"$eq": 10}}, {"bar": {"$ne": 5}}]}"#
);

test_negation!(
    m_and,
    expected = r#"{"$or": [{"foo": {"$gte": 10}}, {"foo": {"$lte": 5}}]}"#,
    input = r#"{"$and": [{"foo": {"$lt": 10}}, {"foo": {"$gt": 5}}]}"#
);

test_negation!(
    m_nor,
    expected = r#"{"$or": [{"foo": {"$lt": 10}}, {"foo": {"$gt": 5}}]}"#,
    input = r#"{"$nor": [{"foo": {"$lt": 10}}, {"foo": {"$gt": 5}}]}"#
);

test_negation!(
    m_not_multiple,
    expected = r#"{"foo": {"$lt": 42, "$gte": 0}}"#,
    input = r#"{"foo": {"$not": {"$lt": 42, "$gte": 0}}}"#
);
