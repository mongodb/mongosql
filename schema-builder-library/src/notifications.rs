use std::fmt;
use std::fmt::{Display, Formatter};

#[derive(Debug, Clone)]
pub enum SamplerAction {
    Querying { partition: u16 },
    Processing { partition: u16 },
    Partitioning { partitions: u16 },
    Error { message: String },
    CriticalError { message: String },
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
            SamplerAction::Querying { partition } => write!(f, "Querying partition {}", partition),
            SamplerAction::Processing { partition } => {
                write!(f, "Processing partition {}", partition)
            }
            SamplerAction::Partitioning { partitions } => {
                write!(f, "Partitioning into {} parts", partitions)
            }
            SamplerAction::Error { message } => write!(f, "Error: {}", message),
            SamplerAction::CriticalError { message } => write!(f, "Critical Error: {}", message),
            SamplerAction::SamplingView => write!(f, "Sampling"),
        }
    }
}

impl Display for SamplerNotification {
    fn fmt(&self, f: &mut Formatter) -> fmt::Result {
        write!(
            f,
            "{} for collection/view: {} in database: {}",
            self.action, self.collection_or_view, self.db
        )
    }
}
