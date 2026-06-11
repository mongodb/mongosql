// index_utilization_experimentation.js
// Run with: mongosh <connection-string> index_utilization_experimentation.js

use("index_utilization_experiment");

// ── truth_tables ──────────────────────────────────────────────────────────────

function loadTruthTables() {

    db.truth_tables.drop();
    db.not_truth_table.drop();


    db.truth_tables.insertMany([
        {_id: 0, a: true, b: true, a_and_b: true, a_or_b: true},
        {_id: 1, a: true, b: false, a_and_b: false, a_or_b: true},
        {_id: 2, a: true, b: null, a_and_b: null, a_or_b: true},
        {_id: 3, a: true, a_and_b: null, a_or_b: true},
        {_id: 4, a: false, b: true, a_and_b: false, a_or_b: true},
        {_id: 5, a: false, b: false, a_and_b: false, a_or_b: false},
        {_id: 6, a: false, b: null, a_and_b: false, a_or_b: null},
        {_id: 7, a: false, a_and_b: false, a_or_b: null},
        {_id: 8, a: null, b: true, a_and_b: null, a_or_b: true},
        {_id: 9, a: null, b: false, a_and_b: false, a_or_b: null},
        {_id: 10, a: null, b: null, a_and_b: null, a_or_b: null},
        {_id: 11, a: null, a_and_b: null, a_or_b: null},
        {_id: 12, b: true, a_and_b: null, a_or_b: true},
        {_id: 13, b: false, a_and_b: false, a_or_b: null},
        {_id: 14, b: null, a_and_b: null, a_or_b: null},
        {_id: 15, a_and_b: null, a_or_b: null},
    ]);

    db.truth_tables.createIndex({a: 1});
    db.truth_tables.createIndex({b: 1});
    db.truth_tables.createIndex({a: 1, b: 1});

    db.not_truth_table.insertMany([
        {_id: 0, a: true, not_a: false},
        {_id: 1, a: false, not_a: true},
        {_id: 2, a: null, not_a: null},
        {_id: 3, not_a: null},
    ]);

    db.not_truth_table.createIndex({a: 1});
    db.not_truth_table.createIndex({not_a: 1});




// ── __sql_schemas ─────────────────────────────────────────────────────────────
// createCollection is required because the leading underscores prevent mongosh
// from resolving db.__sql_schemas as a collection reference.

    db.createCollection("__sql_schemas");

    db.getCollection("__sql_schemas").deleteOne({_id: "truth_tables"});
    db.getCollection("__sql_schemas").deleteOne({_id: "not_truth_table"});

    db.getCollection("__sql_schemas").insertOne({
        _id: "truth_tables",
        schema: {
            bsonType: "object",
            required: ["_id", "a", "b"],
            additionalProperties: false,
            properties: {
                _id: {bsonType: "int"},
                a: {anyOf: [{bsonType: "bool"}, {bsonType: "null"}]},
                b: {anyOf: [{bsonType: "bool"}, {bsonType: "null"}]},
                a_and_b: {anyOf: [{bsonType: "bool"}, {bsonType: "null"}]},
                a_or_b: {anyOf: [{bsonType: "bool"}, {bsonType: "null"}]},
            },
        },
    });
    print("Done — truth_tables loaded with 16 docs, 3 indexes, and schema registered.");

    db.getCollection("__sql_schemas").insertOne({
        _id: "not_truth_table",
        schema: {
            bsonType: "object",
            required: ["_id", "a", "not_a"],
            additionalProperties: false,
            properties: {
                _id: {bsonType: "int"},
                a: {anyOf: [{bsonType: "bool"}, {bsonType: "null"}]},
                not_a: {anyOf: [{bsonType: "bool"}, {bsonType: "null"}]},
            },
        },
    });
    print("Done — not truth table loaded with 4 docs, 1 indexes, and schema registered.");

}

function testOrOperatorIndexUtilization() {
    // MongoSQL Aggregation Pipeline for:"SELECT _id FROM truth_tables WHERE a OR b"
    const mongosql_aggregation_pipeline = [
        { "$match": { "$expr": { "$let": { "vars": { "desugared_sqlOr_input0": "$a", "desugared_sqlOr_input1": "$b" }, "in": { "$cond": [{ "$or": [{ "$eq": ["$$desugared_sqlOr_input0", { "$literal": true }] }, { "$eq": ["$$desugared_sqlOr_input1", { "$literal": true }] }] }, { "$literal": true }, { "$cond": [{ "$or": [{ "$lte": ["$$desugared_sqlOr_input0", { "$literal": null }] }, { "$lte": ["$$desugared_sqlOr_input1", { "$literal": null }] }] }, { "$literal": null }, { "$literal": false }] }] } } } } },
        { "$project": { "truth_tables": "$$ROOT", "_id": 0 } },
        { "$project": { "__bot": { "_id": "$truth_tables._id" }, "_id": 0 } },
        { "$replaceWith": { "$unsetField": { "field": "__bot", "input": { "$setField": { "field": "", "input": "$$ROOT", "value": "$__bot" } } } } },
    ];

    // Aggregation pipeline to test restoring index utilization for $or queries by adding null checks
    const proposed_desugaring_aggregation_pipeline = [
        {
            $match: {
                $expr: {
                    $or: [
                        {
                            $and: [
                                { $gte: ["$a", null] },
                                "$a"
                            ]
                        },
                        {
                            $and: [
                                { $gte: ["$b", null] },
                                "$b"
                            ]
                        }
                    ]
                }
            }
        }
    ]

    const mongosql_result_set = db.truth_tables.aggregate(mongosql_aggregation_pipeline).toArray();
    const proposed_desugaring_result_set = db.truth_tables.aggregate(proposed_desugaring_aggregation_pipeline).toArray();
    const proposed_desugaring_explain_plan = db.truth_tables.explain("executionStats").aggregate(proposed_desugaring_aggregation_pipeline);

    const mongosql_ids = new Set(mongosql_result_set.map(doc => doc[""]._id));
    const proposed_ids = new Set(proposed_desugaring_result_set.map(doc => doc._id));

    const only_in_mongosql = mongosql_ids.difference(proposed_ids);
    const only_in_proposed = proposed_ids.difference(mongosql_ids);

    if (only_in_mongosql.size === 0 && only_in_proposed.size === 0) {
        print(`[Or Operator Index Utilization] PASS: result sets match (${mongosql_ids.size} docs each). Proposed desugaring produces the same results as MongoSQL's desugaring for this query.`);
    } else {
        print(`FAIL: result sets differ`);
        if (only_in_mongosql.size > 0) print(`  only in MongoSQL:  ${JSON.stringify([...only_in_mongosql])}`);
        if (only_in_proposed.size > 0) print(`  only in proposed:  ${JSON.stringify([...only_in_proposed])}`);
    }
}

function testAndOperatorIndexUtilization() {
    // MongoSQL Aggregation Pipeline for:"SELECT _id FROM truth_tables WHERE a AND b"
    const mongosql_aggregation_pipeline = [
        { "$match": { "$expr": { "$let": { "vars": { "desugared_sqlAnd_input0": "$a", "desugared_sqlAnd_input1": "$b" }, "in": { "$cond": [{ "$or": [{ "$eq": ["$$desugared_sqlAnd_input0", { "$literal": false }] }, { "$eq": ["$$desugared_sqlAnd_input1", { "$literal": false }] }] }, { "$literal": false }, { "$cond": [{ "$or": [{ "$lte": ["$$desugared_sqlAnd_input0", { "$literal": null }] }, { "$lte": ["$$desugared_sqlAnd_input1", { "$literal": null }] }] }, { "$literal": null }, { "$literal": true }] }] } } } } },
        { "$project": { "truth_tables": "$$ROOT", "_id": 0 } },
        { "$project": { "__bot": { "_id": "$truth_tables._id" }, "_id": 0 } },
        { "$replaceWith": { "$unsetField": { "field": "__bot", "input": { "$setField": { "field": "", "input": "$$ROOT", "value": "$__bot" } } } } },
    ];

    const proposed_desugaring_aggregation_pipeline = [
        {
            $match: {
                $expr: {
                    $and: [
                        "$a",
                        { $gt: ["$a", null] },
                        "$b",
                        { $gt: ["$b", null] },
                    ]
                }
            }
        }
    ];

    const mongosql_result_set = db.truth_tables.aggregate(mongosql_aggregation_pipeline).toArray();
    const proposed_desugaring_result_set = db.truth_tables.aggregate(proposed_desugaring_aggregation_pipeline).toArray();
    const proposed_desugaring_explain_plan = db.truth_tables.explain("executionStats").aggregate(proposed_desugaring_aggregation_pipeline);

    const mongosql_ids = new Set(mongosql_result_set.map(doc => doc[""]._id));
    const proposed_ids = new Set(proposed_desugaring_result_set.map(doc => doc._id));

    const only_in_mongosql = mongosql_ids.difference(proposed_ids);
    const only_in_proposed = proposed_ids.difference(mongosql_ids);

    if (only_in_mongosql.size === 0 && only_in_proposed.size === 0) {
        print(`[And Operator Index Utilization] PASS: result sets match (${mongosql_ids.size} docs each). Proposed desugaring produces the same results as MongoSQL's desugaring for this query.`);
    } else {
        print(`FAIL: result sets differ`);
        if (only_in_mongosql.size > 0) print(`  only in MongoSQL:  ${JSON.stringify([...only_in_mongosql])}`);
        if (only_in_proposed.size > 0) print(`  only in proposed:  ${JSON.stringify([...only_in_proposed])}`);
    }
}


/**
 *
 * NOT operator: It seems like the existing query gets index utilization so not sure if we need to optimize anything?
 *
 * */
function testNotOperatorIndexUtilization() {
    // MongoSQL Aggregation Pipeline for:"SELECT _id FROM not_truth_table WHERE NOT a"
    const mongosql_aggregation_pipeline = [
        {"$match": {"$expr": {"$and": [{"$gt": ["$a", {"$literal": null}]}, {"$not": ["$a"]}]}}},
        {"$project": {"not_truth_table": "$$ROOT", "_id": 0}},
        {"$project": {"__bot": {"_id": "$not_truth_table._id", "a": "$not_truth_table.a"}, "_id": 0}},
        {
            "$replaceWith": {
                "$unsetField": {
                    "field": "__bot",
                    "input": {"$setField": {"field": "", "input": "$$ROOT", "value": "$__bot"}}
                }
            }
        },
    ];

    const proposed_desugaring_aggregation_pipeline = [
        {
            $match: {
                $expr: {
                    $and: [
                        {$gt: ["$a", null]},
                        {$not: "$a"}
                    ]
                }
            }
        }
    ];

    const mongosql_ids = new Set(mongosql_aggregation_pipeline.map(doc => doc[""]._id));
    const proposed_ids = new Set(proposed_desugaring_aggregation_pipeline.map(doc => doc._id));

    const only_in_mongosql = mongosql_ids.difference(proposed_ids);
    const only_in_proposed = proposed_ids.difference(mongosql_ids);

    if (only_in_mongosql.size === 0 && only_in_proposed.size === 0) {
        print(`[NOT Operator Index Utilization] PASS: result sets match (${mongosql_ids.size} docs each). Proposed desugaring produces the same results as MongoSQL's desugaring for this query.`);
    } else {
        print(`FAIL: result sets differ`);
        if (only_in_mongosql.size > 0) print(`  only in MongoSQL:  ${JSON.stringify([...only_in_mongosql])}`);
        if (only_in_proposed.size > 0) print(`  only in proposed:  ${JSON.stringify([...only_in_proposed])}`);
    }
}

loadTruthTables();
testOrOperatorIndexUtilization();
testAndOperatorIndexUtilization();
