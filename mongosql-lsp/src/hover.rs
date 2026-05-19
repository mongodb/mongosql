//! Static hover-tooltip map for MIR and AIR node names.

use std::collections::HashMap;
use std::sync::LazyLock;

/// Maps a node name (as it appears in `{:#?}` output) to a Markdown tooltip.
pub static HOVER_MAP: LazyLock<HashMap<&'static str, &'static str>> = LazyLock::new(|| {
    let mut m = HashMap::new();

    // ── MIR (Mid-level Intermediate Representation) nodes ────────────────────
    m.insert(
        "Project",
        "**MIR Project** — Evaluates the SELECT-list expressions and produces one output \
         tuple per input row. The `expression` field is a `BindingTuple` whose keys become \
         the output column names.",
    );
    m.insert(
        "Filter",
        "**MIR Filter** — Retains only rows for which `condition` evaluates to a truthy \
         value. Corresponds to a SQL `WHERE` or `HAVING` clause.",
    );
    m.insert(
        "Collection",
        "**MIR Collection** — Scans every document in a MongoDB collection. \
         The `db` and `collection` fields identify the target namespace.",
    );
    m.insert(
        "Group",
        "**MIR Group** — Partitions input rows into groups (SQL `GROUP BY`). \
         The `keys` field lists the grouping expressions; `aggregations` contains \
         the per-group aggregate functions.",
    );
    m.insert(
        "Sort",
        "**MIR Sort** — Orders the result set (SQL `ORDER BY`). \
         Each `SortSpecification` carries an expression and a direction.",
    );
    m.insert(
        "Limit",
        "**MIR Limit** — Restricts the number of output rows (SQL `LIMIT n`).",
    );
    m.insert(
        "Offset",
        "**MIR Offset** — Skips the first `n` rows (SQL `OFFSET n`).",
    );
    m.insert(
        "Unwind",
        "**MIR Unwind** — Unnests an array-valued field into individual rows, one per \
         array element. Equivalent to MongoDB `$unwind`.",
    );
    m.insert(
        "Join",
        "**MIR Join** — Combines rows from two sub-trees (SQL `JOIN`). \
         The `join_type` field indicates `Inner`, `Left`, `Cross`, etc.",
    );
    m.insert(
        "Subquery",
        "**MIR Subquery** — A correlated or uncorrelated sub-query used as an expression \
         or a data source.",
    );
    m.insert(
        "Set",
        "**MIR Set** — A set operation node (`UNION ALL`, `UNION`, `INTERSECT`, `EXCEPT`).",
    );
    m.insert(
        "Derived",
        "**MIR Derived** — An inline view (sub-select in a `FROM` clause).",
    );
    m.insert(
        "ExpressionCollection",
        "**MIR ExpressionCollection** — A virtual collection built from an inline array \
         expression rather than a stored MongoDB collection.",
    );

    // ── AIR (Aggregation Intermediate Representation) nodes ──────────────────
    m.insert(
        "ReplaceWith",
        "**AIR ReplaceWith** — Emits a MongoDB `$replaceWith` aggregation stage that \
         replaces each document with the result of an expression.",
    );
    m.insert(
        "Lookup",
        "**AIR Lookup** — Emits a `$lookup` aggregation stage to implement a SQL `JOIN`.",
    );
    m.insert(
        "Unset",
        "**AIR Unset** — Emits a `$unset` aggregation stage to remove fields from \
         documents.",
    );
    m.insert(
        "AddFields",
        "**AIR AddFields** — Emits a `$addFields` aggregation stage to add or overwrite \
         document fields.",
    );
    m.insert(
        "Match",
        "**AIR Match** — Emits a `$match` stage (corresponds to a MIR `Filter`).",
    );
    m.insert(
        "Project", // AIR also has Project
        "**AIR / MIR Project** — Emits a `$project` aggregation stage that reshapes \
         documents to the desired output fields.",
    );
    m.insert(
        "Group", // AIR also has Group
        "**AIR Group** — Emits a `$group` aggregation stage (SQL `GROUP BY`).",
    );
    m.insert(
        "Sort", // AIR also has Sort
        "**AIR Sort** — Emits a `$sort` aggregation stage (SQL `ORDER BY`).",
    );
    m.insert(
        "Limit", // AIR also has Limit
        "**AIR Limit** — Emits a `$limit` aggregation stage.",
    );
    m.insert(
        "Skip",
        "**AIR Skip** — Emits a `$skip` aggregation stage (SQL `OFFSET`).",
    );
    m.insert(
        "Unwind", // AIR also has Unwind
        "**AIR Unwind** — Emits a `$unwind` aggregation stage.",
    );
    m.insert(
        "Documents",
        "**AIR Documents** — Emits a `$documents` stage for inline data.",
    );
    m.insert(
        "Collection", // AIR root source
        "**AIR Collection** — The root source of an aggregation pipeline, identifying the \
         MongoDB collection to run the pipeline against.",
    );

    m
});
