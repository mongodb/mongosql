pub mod catalog;
pub mod metrics;
pub mod service;
pub use service::{PanicHandlingTranslateSqlService, TranslateSqlService};

pub mod trace;
pub mod version;

pub mod translator {
    include!("translator.v1.rs");
}

pub mod logger {
    use std::env;
    use tracing::Level;
    use tracing_subscriber::{filter::EnvFilter, fmt::time};

    pub fn init_logger() {
        let log_level_str = env::var("LOG_LEVEL").unwrap_or_else(|_| "info".into());
        let log_level = log_level_str.parse::<Level>().unwrap_or(Level::INFO);
        let env_filter = EnvFilter::new(log_level_str);

        tracing_subscriber::fmt()
            .with_env_filter(env_filter)
            .with_target(true)
            .with_timer(time::ChronoLocal::new("%Y-%m-%d %H:%M:%S".to_string()))
            .init();

        tracing::info!("Logger initialized with level: {}", log_level);
    }
}
