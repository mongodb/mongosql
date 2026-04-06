use bson::{Bson, Document};
use mongosql::schema::Schema;

// An initial schema, parsed from an existing entry
pub struct InitialSchema {
    pub collection: String,
    pub schema: Schema,
}

#[derive(Debug, thiserror::Error)]
pub enum Error {
    #[error("The document is missing the required _id field")]
    MissingId,

    #[error("The document is missing the required schema field")]
    MissingSchema,

    #[error("The schema is invalid: {0}")]
    InvalidSchema(#[from] mongosql::schema::Error),
}

impl TryFrom<Document> for InitialSchema {
    type Error = Error;

    fn try_from(mut doc: Document) -> Result<Self, Self::Error> {
        let Some(Bson::String(collection)) = doc.remove("_id") else {
            return Err(Error::MissingId);
        };
        let Some(Bson::Document(schema)) = doc.remove("schema") else {
            return Err(Error::MissingSchema);
        };

        Ok(InitialSchema {
            collection,
            schema: schema.try_into()?,
        })
    }
}
