use bson::Bson;
use mongosql::schema::{Schema, definitions::Document};

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

impl TryFrom<bson::Document> for InitialSchema {
    type Error = Error;

    fn try_from(mut doc: bson::Document) -> Result<Self, Self::Error> {
        let Some(Bson::String(collection)) = doc.remove("_id") else {
            return Err(Error::MissingId);
        };
        let Some(Bson::Document(schema)) = doc.remove("schema") else {
            return Err(Error::MissingSchema);
        };

        let mut schema: Schema = schema.try_into()?;

        // Old runs might not have an unstable flag, so we default to false if
        // the field isn't found (or is something that isn't a boolean). If it
        // _is_ found, then we need to make sure that the schema itself has the
        // unstable flag.
        let is_unstable = doc.get("unstable").and_then(Bson::as_bool).unwrap_or(false);
        if is_unstable && let Schema::Document(d) = schema {
            schema = Schema::Document(Document {
                unstable: true,
                ..d
            });
        }

        Ok(InitialSchema { collection, schema })
    }
}
