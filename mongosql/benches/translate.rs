use criterion::{black_box, criterion_group, criterion_main, Criterion};
use mongosql::{
    catalog::{Catalog, Namespace},
    map,
    schema::ANY_DOCUMENT,
    translate_sql, SchemaCheckingMode, Translation,
};
use pprof::criterion::{Output, PProfProfiler};

fn translate(sql: &str) -> Translation {
    let current_db = "mydb";
    let catalog = Catalog::new(map! {
        Namespace {db: "mydb".into(), collection: "foo".into()} => ANY_DOCUMENT.clone(),
    });
    let schema_checking_mode = SchemaCheckingMode::Relaxed;
    translate_sql(current_db, sql, &catalog, schema_checking_mode).unwrap()
}

fn criterion_benchmark(c: &mut Criterion) {
    c.bench_function("translate_simple", |b| {
        b.iter(|| translate(black_box("select * from foo where a > 5")))
    });
}

criterion_group!(
    name = benches;
    config = Criterion::default().with_profiler(PProfProfiler::new(100, Output::Flamegraph(None)));
    targets = criterion_benchmark
);
criterion_main!(benches);
