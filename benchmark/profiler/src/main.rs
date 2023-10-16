use std::collections::HashMap;

use catalog::load_catalog;
use mongosql::{
    map,
    options::{ExcludeNamespacesOption, SqlOptions},
    translate_sql, SchemaCheckingMode,
};

struct Query {
    db: &'static str,
    query: &'static str,
}

const TPCH: Query = Query {
    db: "tpch",
    query: r#"SELECT
          l_returnflag,
          l_linestatus,
          sum(l_quantity) as sum_qty,
          sum(l_extendedprice) as sum_base_price,
          sum(l_extendedprice * (1 - l_discount)) as sum_disc_price,
          sum(l_extendedprice * (1 - l_discount) * (1 + l_tax)) as sum_charge,
          avg(l_quantity) as avg_qty,
          avg(l_extendedprice) as avg_price,
          avg(l_discount) as avg_disc,
          count(*) as count_order
      FROM
          lineitem
      WHERE
          l_shipdate <= '1998-12-01T00:00:00.000Z'::TIMESTAMP
      GROUP BY
          l_returnflag,
          l_linestatus
      ORDER BY
          l_returnflag,
          l_linestatus"#,
};

const SAMPLE_ANALYTICS: Query = Query {
    db: "sample_analytics",
    query: "SELECT * FROM FLATTEN(customers_tiers_and_details_view)",
};

fn main() {
    let queries: HashMap<&str, Query> =
        map!("tpch" => TPCH, "sample_analytics" => SAMPLE_ANALYTICS);
    if cfg!(feature = "tpch") {
        profile(queries.get("tpch").unwrap())
    } else {
        profile(queries.get("sample_analytics").unwrap())
    }
}

fn profile(query: &Query) {
    let base_dir = env!("CARGO_MANIFEST_DIR");
    let catalogs_dir = format!("{}/src/catalog/catalogs/{}.jsonc", base_dir, query.db);
    let catalog = load_catalog(&catalogs_dir).unwrap();
    let res = translate_sql(
        query.db,
        query.query,
        &catalog,
        SqlOptions::new(
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::Relaxed,
        ),
    )
    .unwrap();
    let _ = res;
}
