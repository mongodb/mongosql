use catalog::load_catalog;
use criterion::{black_box, criterion_group, criterion_main, Criterion};
use lazy_static::lazy_static;
use mongosql::{
    catalog::Catalog,
    options::{ExcludeNamespacesOption, SqlOptions},
    translate_sql, SchemaCheckingMode, Translation,
};
use pprof::criterion::{Output, PProfProfiler};

// load our massive catalog
lazy_static! {
    static ref CATALOG: Catalog =
        load_catalog("./src/catalog/catalogs/sample_analytics.json").unwrap();
}

fn translate(sql: &str) -> Translation {
    translate_sql(
        "sample_analytics",
        sql,
        &CATALOG,
        SqlOptions::new(
            ExcludeNamespacesOption::IncludeNamespaces,
            SchemaCheckingMode::Relaxed,
        ),
    )
    .unwrap()
}

fn criterion_benchmark(c: &mut Criterion) {
    c.bench_function("translate_complex", |b| {
        b.iter(|| {
            translate(black_box(
                "SELECT * FROM FLATTEN(customers_tiers_and_details_view)",
            ))
        })
    });
}

// this runs our "criterion_benchmark" function 10 times to give us an average execution time
criterion_group!(
    name = benches;
    config = Criterion::default().sample_size(10).with_profiler(PProfProfiler::new(10, Output::Flamegraph(None)));
    targets = criterion_benchmark
);
criterion_main!(benches);
