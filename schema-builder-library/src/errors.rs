use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("JsonSchemaFailure")]
    JsonSchemaFailure,
    #[error("BsonFailure")]
    BsonFailure,
    #[error("Unable to get collection stats for {0}")]
    NoCollectionStats(String),
    #[error("Unable to get bounds for collection: {0}")]
    NoBounds(String),
    #[error("Collection {0} appears to be empty")]
    EmptyCollection(String),
    #[error("NoIdInSample")]
    NoIdInSample,
    #[error("Driver Error {0}")]
    DriverError(mongodb::error::Error),
    #[error("NoCollection {0}")]
    NoCollection(String),
    #[error("Execution Error {0}")]
    TokioError(tokio::task::JoinError),
}

impl From<mongodb::error::Error> for Error {
    fn from(value: mongodb::error::Error) -> Self {
        Self::DriverError(value)
    }
}
