use std::fmt;
use std::fmt::{Display, Formatter};

#[macro_export]
macro_rules! notify {
    ($channel:expr, $notification:expr $(,)?) => {{
        match $notification.action {
            SamplerAction::Warning { .. } => {
                tracing::warn!("{}", $notification);
                $channel.send($notification).unwrap_or_default();
            }
            SamplerAction::Error { .. } => {
                tracing::error!("{}", $notification);
                $channel.send($notification).unwrap_or_default();
            }
            SamplerAction::Info { .. } => {
                // Log but do not send to avoid flooding the channel.
                tracing::info!("{}", $notification);
            }
            SamplerAction::Querying { .. }
            | SamplerAction::Processing { .. }
            | SamplerAction::Partitioning { .. }
            | SamplerAction::SamplingView
            | SamplerAction::UsingInitialSchema => {
                // Log at trace level but do not send to avoid unbounded memory growth.
                tracing::trace!("{}", $notification);
            }
        }
    }};
}

#[macro_export]
macro_rules! notify_error {
    ($channel:expr, $db:expr, $coll:expr, $message:expr $(,)?) => {{
        let notification = SamplerNotification {
            db: $db.to_string(),
            collection_or_view: $coll.to_string(),
            action: SamplerAction::Error {
                message: $message.to_string(),
            },
        };
        notify!($channel, notification);
    }};
}

#[macro_export]
macro_rules! notify_warning {
    ($channel:expr, $db:expr, $coll:expr, $message:expr $(,)?) => {{
        let notification = SamplerNotification {
            db: $db.to_string(),
            collection_or_view: $coll.to_string(),
            action: SamplerAction::Warning {
                message: $message.to_string(),
            },
        };
        notify!($channel, notification);
    }};
}

#[macro_export]
macro_rules! notify_info {
    ($channel:expr, $db:expr, $coll:expr, $message:expr $(,)?) => {{
        let notification = SamplerNotification {
            db: $db.to_string(),
            collection_or_view: $coll.to_string(),
            action: SamplerAction::Info {
                message: $message.to_string(),
            },
        };
        notify!($channel, notification);
    }};
}

#[derive(Debug, Clone)]
pub enum SamplerAction {
    Querying { partition: u16 },
    Processing { partition: u16 },
    Partitioning { partitions: u16 },
    UsingInitialSchema,
    Warning { message: String },
    Error { message: String },
    Info { message: String },
    SamplingView,
}

#[derive(Debug, Clone)]
pub struct SamplerNotification {
    pub db: String,
    pub collection_or_view: String,
    pub action: SamplerAction,
}

impl Display for SamplerAction {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        match self {
            SamplerAction::Querying { partition } => {
                write!(f, "Querying partition {partition}")
            }
            SamplerAction::Processing { partition } => {
                write!(f, "Processing partition {partition}")
            }
            SamplerAction::Partitioning { partitions } => {
                write!(f, "Partitioning into {partitions} parts")
            }
            SamplerAction::UsingInitialSchema => {
                write!(f, "Using initial schema")
            }
            SamplerAction::Warning { message } => write!(f, "Warning: {message}"),
            SamplerAction::Error { message } => write!(f, "Error: {message}"),
            SamplerAction::Info { message } => write!(f, "Info: {message}"),
            SamplerAction::SamplingView => write!(f, "Sampling view "),
        }
    }
}

impl Display for SamplerNotification {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        write!(
            f,
            "{} in collection: {} in database: {}",
            self.action, self.collection_or_view, self.db
        )
    }
}
