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
    #[error("NoIdInSample")]
    NoIdInSample,
    #[error("Driver Error {0}")]
    DriverError(mongodb::error::Error),
    #[error("NoCollection {0}")]
    NoCollection(String),
}

impl From<mongodb::error::Error> for Error {
    fn from(value: mongodb::error::Error) -> Self {
        Self::DriverError(value)
    }
}
