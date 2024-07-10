use std::fmt;
use std::fmt::{Display, Formatter};

#[macro_export]
macro_rules! notify {
    ($channel:expr, $notification:expr $(,)?) => {{
        match $notification.action {
            SamplerAction::Warning { .. } => {
                tracing::warn!("{}", $notification);
            }
            SamplerAction::Error { .. } => {
                tracing::error!("{}", $notification);
            }
            SamplerAction::Querying { .. } => tracing::trace!("{}", $notification),
            SamplerAction::Processing { .. } => tracing::trace!("{}", $notification),
            SamplerAction::Partitioning { .. } => tracing::trace!("{}", $notification),
            SamplerAction::SamplingView => tracing::trace!("{}", $notification),
        }

        $channel.send($notification).unwrap_or_default();
    }};
}

#[derive(Debug, Clone)]
pub enum SamplerAction {
    Querying { partition: u16 },
    Processing { partition: u16 },
    Partitioning { partitions: u16 },
    Warning { message: String },
    Error { message: String },
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
                write!(f, "Querying partition {} in collection", partition)
            }
            SamplerAction::Processing { partition } => {
                write!(f, "Processing partition {} in collection", partition)
            }
            SamplerAction::Partitioning { partitions } => {
                write!(f, "Partitioning into {} parts for collection", partitions)
            }
            SamplerAction::Warning { message } => write!(f, "Warning: {}", message),
            SamplerAction::Error { message } => write!(f, "Error: {}", message),
            SamplerAction::SamplingView => write!(f, "Sampling view "),
        }
    }
}

impl Display for SamplerNotification {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        write!(
            f,
            "{} {} in database: {}",
            self.action, self.collection_or_view, self.db
        )
    }
}
