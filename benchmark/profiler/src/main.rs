use config_loader::load_query_and_catalog;
use mongosql::{
    options::{ExcludeNamespacesOption, SqlOptions},
    translate_sql, SchemaCheckingMode,
};
use std::env;

fn main() {
    let args: Vec<String> = env::args().collect();
    let default_query_name = "sample_analytics".to_string();
    let query_name = args.get(1).unwrap_or(&default_query_name);

    profile(query_name);
}

fn profile(query_name: &str) {
    let query_and_catalog = load_query_and_catalog(query_name);
    match query_and_catalog {
        Ok((query, catalog)) => {
            if let Some(skip_reason) = query.skip_reason {
                println!("skipping {}: {}", query_name, skip_reason);
                return;
            }

            let res = translate_sql(
                &query.db,
                &query.query,
                &catalog,
                SqlOptions::new(
                    ExcludeNamespacesOption::IncludeNamespaces,
                    SchemaCheckingMode::Relaxed,
                ),
            )
            .unwrap();
            let _ = res;
        }
        Err(e) => {
            println!("Encountered Error: {}", e);
            std::process::exit(1);
        }
    }
}
