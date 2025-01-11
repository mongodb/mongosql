use crate::{
    json_schema, map,
    mir::{
        binding_tuple::{self, BindingTuple, DatasourceName, DuplicateKeyError, Key},
        Type, TypeOrMissing,
    },
    schema::Schema::{AnyOf, Unsat},
    set,
};
use enum_iterator::IntoEnumIterator;
use itertools::Itertools;
use lazy_static::lazy_static;
use std::collections::{HashMap, HashSet};
use std::{
    collections::{BTreeMap, BTreeSet},
    fmt::{Display, Formatter},
    iter::once,
};
use thiserror::Error;

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("{0:?} is not a valid BSON type")]
    InvalidBSONType(String),
    #[error("invalid combination of fields")]
    InvalidCombinationOfFields(),
    #[error("cannot exhaustively enumerate all field paths in schema {0:?}")]
    CannotEnumerateAllFieldPaths(Schema),
    #[error("cannot convert bson type {0:?} to atomic")]
    CannotConvertBsonTypeToAtomic(json_schema::BsonTypeName),
    #[error("{0}")]
    FieldConflictInNonNamespacedResult(String),
    #[error("{0}")]
    InvalidBottomField(String),
    #[error("{0}")]
    InvalidNamespace(String),
    #[error("{0}")]
    UnsupportedBsonType(String),
    #[error("JsonSchemaFailure")]
    JsonSchemaFailure,
    #[error("bson failure: {0}")]
    BsonFailure(#[from] json_schema::Error),
}

impl From<user_schema_error::Error> for Error {
    fn from(value: user_schema_error::Error) -> Self {
        match value {
            user_schema_error::Error::FieldConflictInNonNamespacedResult(_, _) => {
                Error::FieldConflictInNonNamespacedResult(value.to_string())
            }
            user_schema_error::Error::UnsupportedBsonType(_) => {
                Error::UnsupportedBsonType(value.to_string())
            }
        }
    }
}

mod user_schema_error {
    // namespace_error implements the UserError trait and is separate from schema::Error
    // as this error could be directly returned to end-users
    use crate::schema::Schema;
    use crate::usererror::{UserError, UserErrorDisplay};

    #[derive(Debug, UserErrorDisplay, PartialEq)]
    pub enum Error {
        FieldConflictInNonNamespacedResult(Vec<String>, Vec<Schema>),
        UnsupportedBsonType(String),
    }

    impl UserError for Error {
        fn code(&self) -> u32 {
            match self {
                Error::FieldConflictInNonNamespacedResult(_, _) => 4000,
                Error::UnsupportedBsonType(_) => 1018,
            }
        }

        fn user_message(&self) -> Option<String> {
            match self {
                Error::FieldConflictInNonNamespacedResult(names, _) => Some(format!(
                    "Consider aliasing the following conflicting field(s) to unique names: {}",
                    names.join(", ")
                )),
                Error::UnsupportedBsonType(s) => Some(format!("Unsupported BSON type: {s}.")),
            }
        }

        fn technical_message(&self) -> String {
            match self {
                Error::FieldConflictInNonNamespacedResult(_, schemas) => {
                    format!(
                        "Cannot return non-namespaced result set due to field name \
                             conflict(s), schema {:?}",
                        schemas
                    )
                }
                Error::UnsupportedBsonType(t) => format!("Deprecated BSON type {t}"),
            }
        }
    }
}

#[derive(PartialEq, Eq, Debug, Clone, Default)]
pub struct SchemaEnvironment(BindingTuple<Schema>);

impl SchemaEnvironment {
    /// Takes all Datasource-Schema key-value pairs from a SchemaEnvironment
    /// and adds them to the current SchemaEnvironment, returning the modified
    /// SchemaEnvironment.
    ///
    /// Schema values with duplicate Datasource keys are bundled in an AnyOf under
    /// that same Datasource key.
    pub fn union(self, other: SchemaEnvironment) -> Self {
        let mut out = self;
        for (k, v) in other.0.into_iter() {
            out = out.union_schema_for_datasource(k, v);
        }
        out
    }

    /// check_for_non_namespaced_collisions iterates through the SchemaEnvironment
    /// and checks for duplicate field names across datasources. If duplicates are found,
    /// it will return an error with a list of conflicting names and schemas.
    /// The `additional_properties` flag is checked and an error is returned if true
    pub fn check_for_non_namespaced_collisions(&self) -> Result<(), Error> {
        let mut column_names_and_key = HashMap::new();
        let mut name_duplicates = Vec::new();
        let mut keys_with_duplicates = BTreeSet::new();
        for (key, schema) in self.iter() {
            if let Schema::Document(d) = schema {
                if d.additional_properties {
                    return Err(Error::CannotEnumerateAllFieldPaths(schema.clone()));
                }
            }
            for column in schema.keys() {
                if let Some(key_with_duplicate) =
                    column_names_and_key.insert(column.clone(), key.clone())
                {
                    name_duplicates.push(column.clone());
                    // Insert the existing key with duplicate column and current key
                    keys_with_duplicates.insert(key_with_duplicate);
                    keys_with_duplicates.insert(key.clone());
                }
            }
        }
        if !name_duplicates.is_empty() {
            return Err(
                user_schema_error::Error::FieldConflictInNonNamespacedResult(
                    name_duplicates,
                    keys_with_duplicates
                        .iter()
                        .filter_map(|key| self.get(key))
                        .cloned()
                        .collect(),
                )
                .into(),
            );
        }
        Ok(())
    }

    /// Inserts a Datasource-Schema key-value pair into the current
    /// SchemaEnvironment, returning the modified SchemaEnvironment.
    ///
    /// If inserting a key with Schema value V, but the SchemaEnvironment already
    /// contains the key with existing Schema value W, the existing Schema value
    /// is overwritten to AnyOf(V, W).
    pub fn union_schema_for_datasource(
        mut self,
        datasource_key: binding_tuple::Key,
        schema_value: Schema,
    ) -> Self {
        if let Some(s) = self.0.remove(&datasource_key) {
            self.0
                .insert(datasource_key, Schema::AnyOf(set![s, schema_value]));
        } else {
            self.0.insert(datasource_key, schema_value);
        }
        self
    }

    pub fn nearest_scope_for_datasource(&self, d: &DatasourceName, scope: u16) -> Option<u16> {
        self.0.nearest_scope_for_datasource(d, scope)
    }

    pub fn new() -> Self {
        Self(BindingTuple::new())
    }

    pub fn len(&self) -> usize {
        self.0.len()
    }

    pub fn get(&self, k: &Key) -> Option<&Schema> {
        self.0.get(k)
    }

    pub fn insert(&mut self, k: Key, v: Schema) -> Option<Schema> {
        self.0.insert(k, v)
    }

    pub fn remove(&mut self, k: &Key) -> Option<Schema> {
        self.0.remove(k)
    }

    pub fn contains_key(&self, k: &Key) -> bool {
        self.0.contains_key(k)
    }

    pub fn is_empty(&self) -> bool {
        self.0.is_empty()
    }

    pub fn keys(&self) -> impl Iterator<Item = &Key> {
        self.0.keys()
    }

    pub fn iter(&self) -> impl Iterator<Item = (&Key, &Schema)> {
        self.0.iter()
    }

    pub fn merge(&mut self, other: SchemaEnvironment) -> Result<(), DuplicateKeyError> {
        self.0.merge(other.0)
    }

    pub fn with_merged_mappings(self, mappings: SchemaEnvironment) -> Self {
        self.0
            .with_merged_mappings(mappings.0)
            .map(SchemaEnvironment)
            .expect("cannot create schema environment with duplicate datasource")
    }
}

impl IntoIterator for SchemaEnvironment {
    type Item = (Key, Schema);
    type IntoIter = <BindingTuple<Schema> as IntoIterator>::IntoIter;

    fn into_iter(self) -> Self::IntoIter {
        self.0.into_iter()
    }
}

impl FromIterator<(Key, Schema)> for SchemaEnvironment {
    fn from_iter<I: IntoIterator<Item = (Key, Schema)>>(iter: I) -> Self {
        let mut bt = SchemaEnvironment(BindingTuple::new());
        for (k, v) in iter {
            bt.0.insert(k, v);
        }
        bt
    }
}

#[derive(PartialEq, Eq, Debug, Clone, Default)]
pub struct ResultSet {
    pub schema_env: SchemaEnvironment,
    pub min_size: u64,
    pub max_size: Option<u64>,
}

impl ResultSet {
    pub fn has_datasource(&self, datasource: &Key) -> bool {
        self.schema_env.contains_key(datasource)
    }

    /// contains_field determines if this ResultSet contains the field,
    /// specified by `path`, in the provided datasource. If the ResultSet
    /// does include the datasource and contains the full path, returns true.
    /// Otherwise returns false.
    pub fn contains_field(&self, datasource: &Key, path: &[String]) -> bool {
        if let Some(s) = self.schema_env.get(datasource) {
            let mut cur_schema = s;

            for field in path.iter() {
                match cur_schema {
                    Schema::Document(d) => {
                        // If the Document schema contains the next part of the path,
                        // continue searching down the Document schema chain, otherwise
                        // return false since part of the path is not included.
                        if let Some(field_s) = d.keys.get(field) {
                            cur_schema = field_s;
                        } else {
                            return false;
                        }
                    }
                    // If the current schema is not a Document but there are
                    // more field path components, the ResultSet must not contain
                    // the full path.
                    _ => return false,
                }
            }

            // If we get through iteration, we know the field must be included in
            // this ResultSet.
            true
        } else {
            // Does not contain the datasource.
            false
        }
    }
}

#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone)]
pub enum Schema {
    Unsat,
    Missing,
    Atomic(Atomic),
    Array(Box<Schema>),
    Document(Document),
    AnyOf(BTreeSet<Schema>),
    Any,
}

impl TryFrom<Schema> for bson::Document {
    type Error = Error;
    fn try_from(schema: Schema) -> std::result::Result<Self, Self::Error> {
        Ok(bson::doc! {
            "$jsonSchema": bson::Bson::try_from(schema)?
        })
    }
}

impl TryFrom<Schema> for bson::Bson {
    type Error = Error;
    fn try_from(schema: Schema) -> std::result::Result<Self, Self::Error> {
        let json_schema: json_schema::Schema = schema
            .clone()
            .try_into()
            .map_err(|_| Error::JsonSchemaFailure)?;
        let bson_schema = json_schema.to_bson().map_err(Error::BsonFailure)?;

        Ok(bson_schema)
    }
}

impl TryFrom<bson::Document> for Schema {
    type Error = Error;
    fn try_from(schema_doc: bson::Document) -> std::result::Result<Self, Self::Error> {
        let json_schema: json_schema::Schema = match schema_doc.get_document("$jsonSchema") {
            Ok(doc) => json_schema::Schema::from_document(doc).map_err(Error::BsonFailure),
            Err(_) => json_schema::Schema::from_document(&schema_doc).map_err(Error::BsonFailure),
        }?;
        let sampler_schema: Schema = json_schema
            .try_into()
            .map_err(|_| Error::JsonSchemaFailure)?;
        Ok(sampler_schema)
    }
}

#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone, Copy, IntoEnumIterator)]
pub enum Atomic {
    MinKey,
    Null,
    Integer,
    Long,
    Double,
    Decimal,
    Symbol,
    String,
    BinData,
    Undefined,
    ObjectId,
    Boolean,
    Date,
    Timestamp,
    Regex,
    DbPointer,
    Javascript,
    JavascriptWithScope,
    MaxKey,
}

impl Display for Schema {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        match self {
            Schema::Any => write!(f, "any type"),
            Unsat => write!(f, "Unsat"),
            Schema::Missing => write!(f, "missing value"),
            Schema::Atomic(atomic) => atomic.fmt(f),
            AnyOf(schema_set) => {
                let self_copy = AnyOf(schema_set.clone());

                if schema_set.len() == 3
                    && schema_set.contains(&Schema::Missing)
                    && schema_set.contains(&Schema::Atomic(Atomic::Null))
                {
                    let self_copy_without_nullish = self_copy.clone().subtract_nullish();
                    match self_copy_without_nullish {
                        AnyOf(mut set) => {
                            let anyof_contents = set.pop_first().unwrap();
                            if let AnyOf(_) = anyof_contents {
                                write!(f, "polymorphic type")
                            } else {
                                write!(f, "nullable {}", anyof_contents)
                            }
                        }
                        _ => unreachable!(),
                    }
                } else if self_copy.eq(&INTEGER_LONG_OR_NULLISH) {
                    write!(f, "nullable long or integer")
                } else if self_copy.eq(&NUMERIC_OR_NULLISH) {
                    write!(f, "nullable numeric type")
                } else if self_copy.eq(&NUMERIC) {
                    write!(f, "numeric type")
                } else if self_copy.eq(&NULLISH) {
                    write!(f, "null type")
                } else if self_copy.eq(&NON_NULLISH) {
                    write!(f, "non-null type")
                } else {
                    write!(f, "polymorphic type")
                }
            }
            Schema::Array(_) => write!(f, "array type"),
            Schema::Document(_) => write!(f, "object type"),
        }
    }
}

impl Display for Atomic {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        match self {
            Atomic::String => write!(f, "string"),
            Atomic::Integer => write!(f, "int"),
            Atomic::Long => write!(f, "long"),
            Atomic::Double => write!(f, "double"),
            Atomic::Decimal => write!(f, "decimal"),
            Atomic::BinData => write!(f, "binData"),
            Atomic::Undefined => write!(f, "undefined"),
            Atomic::ObjectId => write!(f, "objectId"),
            Atomic::Boolean => write!(f, "boolean"),
            Atomic::Date => write!(f, "date"),
            Atomic::Null => write!(f, "null"),
            Atomic::Regex => write!(f, "regex"),
            Atomic::DbPointer => write!(f, "dbPointer"),
            Atomic::Javascript => write!(f, "javascript"),
            Atomic::Symbol => write!(f, "symbol"),
            Atomic::JavascriptWithScope => write!(f, "javascriptWithScope"),
            Atomic::Timestamp => write!(f, "timestamp"),
            Atomic::MinKey => write!(f, "minKey"),
            Atomic::MaxKey => write!(f, "maxKey"),
        }
    }
}

impl TryFrom<json_schema::BsonTypeName> for Atomic {
    type Error = Error;

    fn try_from(t: json_schema::BsonTypeName) -> Result<Self, Self::Error> {
        use json_schema::BsonTypeName;
        match t {
            BsonTypeName::String => Ok(Atomic::String),
            BsonTypeName::Int => Ok(Atomic::Integer),
            BsonTypeName::Double => Ok(Atomic::Double),
            BsonTypeName::Long => Ok(Atomic::Long),
            BsonTypeName::Decimal => Ok(Atomic::Decimal),
            BsonTypeName::BinData => Ok(Atomic::BinData),
            BsonTypeName::ObjectId => Ok(Atomic::ObjectId),
            BsonTypeName::Bool => Ok(Atomic::Boolean),
            BsonTypeName::Date => Ok(Atomic::Date),
            BsonTypeName::Null => Ok(Atomic::Null),
            BsonTypeName::Regex => Ok(Atomic::Regex),
            BsonTypeName::DbPointer => Ok(Atomic::DbPointer),
            BsonTypeName::Javascript => Ok(Atomic::Javascript),
            BsonTypeName::Symbol => Ok(Atomic::Symbol),
            BsonTypeName::JavascriptWithScope => Ok(Atomic::JavascriptWithScope),
            BsonTypeName::Timestamp => Ok(Atomic::Timestamp),
            BsonTypeName::MinKey => Ok(Atomic::MinKey),
            BsonTypeName::MaxKey => Ok(Atomic::MaxKey),
            BsonTypeName::Undefined => Ok(Atomic::Undefined),
            BsonTypeName::Object | BsonTypeName::Array => {
                Err(Error::CannotConvertBsonTypeToAtomic(t))
            }
        }
    }
}

#[derive(Debug, Clone, Copy)]
pub struct JaccardIndex {
    pub avg_ji: f64,
    pub num_unions: u32,
    pub stability_limit: f64,
}

impl Default for JaccardIndex {
    fn default() -> Self {
        JaccardIndex {
            avg_ji: 1.0,
            num_unions: 0,
            stability_limit: 0.8,
        }
    }
}

impl JaccardIndex {
    pub fn new(stability_limit: f64) -> Self {
        JaccardIndex {
            stability_limit,
            ..Default::default()
        }
    }
}

impl PartialEq for JaccardIndex {
    fn eq(&self, _: &Self) -> bool {
        true
    }
}

impl Eq for JaccardIndex {}

impl PartialOrd for JaccardIndex {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        Some(self.cmp(other))
    }
}

impl Ord for JaccardIndex {
    fn cmp(&self, _other: &Self) -> std::cmp::Ordering {
        std::cmp::Ordering::Equal
    }
}

#[derive(Eq, PartialOrd, Ord, Clone, Default)]
pub struct Document {
    pub keys: BTreeMap<String, Schema>,
    pub required: BTreeSet<String>,
    pub additional_properties: bool,
    // JaccardIndex is an optional field that is used to track the stability of the schema
    pub jaccard_index: Option<JaccardIndex>,
}

impl PartialEq for Document {
    fn eq(&self, other: &Self) -> bool {
        self.keys == other.keys
            && self.required == other.required
            && self.additional_properties == other.additional_properties
    }
}

impl Document {
    /// Check equality including jaccard_index.
    ///
    /// This exists for easy recursive comparison of documents
    /// at multiple levels of nesting.
    pub fn eq_with_jaccard_index(&self, other: &Self) -> bool {
        self.eq(other) && self.jaccard_index == other.jaccard_index
    }

    /// num_keys returns the min and max number of keys that a document matching
    /// this schema could contain.
    pub fn num_keys(&self) -> (usize, Option<usize>) {
        let min = self.required.len();
        let max = match self.additional_properties {
            true => None,
            false => Some(self.keys.len()),
        };
        (min, max)
    }
}

impl std::fmt::Debug for Document {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let mut debug = f.debug_struct("Document");
        debug.field("keys", &self.keys);
        debug.field("required", &self.required);
        debug.field("additional_properties", &self.additional_properties);
        debug.finish()
    }
}

impl TryFrom<json_schema::Schema> for Document {
    type Error = Error;

    /// try_from tries to construct a Schema::Document from the passed-in JSON schema,
    /// and returns an error if the call to Schema::try_from fails.
    fn try_from(v: json_schema::Schema) -> Result<Self, Self::Error> {
        Ok(Document {
            keys: v
                .properties
                .unwrap_or_default()
                .into_iter()
                .map(|(key, schema)| {
                    Ok::<(std::string::String, Schema), Self::Error>((
                        key,
                        Schema::try_from(schema)?,
                    ))
                })
                .collect::<Result<_, _>>()?,
            required: v
                .required
                .unwrap_or_default()
                .into_iter()
                .collect::<BTreeSet<String>>(),
            additional_properties: v.additional_properties.unwrap_or(true),
            ..Default::default()
        })
    }
}

impl TryFrom<Document> for json_schema::Schema {
    type Error = Error;

    fn try_from(v: Document) -> Result<Self, Self::Error> {
        Ok(json_schema::Schema {
            bson_type: Some(json_schema::BsonType::Single(
                json_schema::BsonTypeName::Object,
            )),
            properties: Some(
                v.keys
                    .into_iter()
                    .map(|(k, v)| match json_schema::Schema::try_from(v) {
                        Ok(s) => Ok((k, s)),
                        Err(e) => Err(e),
                    })
                    .collect::<Result<_, _>>()?,
            ),
            required: if v.required.is_empty() {
                None
            } else {
                Some(v.required.into_iter().collect())
            },
            additional_properties: Some(v.additional_properties),
            items: None,
            max_items: None,
            any_of: None,
            one_of: None,
        })
    }
}

impl From<Atomic> for json_schema::Schema {
    fn from(v: Atomic) -> Self {
        json_schema::Schema {
            bson_type: Some(json_schema::BsonType::Single(v.into())),
            properties: None,
            required: None,
            additional_properties: None,
            items: None,
            max_items: None,
            any_of: None,
            one_of: None,
        }
    }
}

impl From<Atomic> for json_schema::BsonTypeName {
    fn from(v: Atomic) -> Self {
        use self::Atomic::*;
        match v {
            Decimal => json_schema::BsonTypeName::Decimal,
            Double => json_schema::BsonTypeName::Double,
            Integer => json_schema::BsonTypeName::Int,
            Long => json_schema::BsonTypeName::Long,
            String => json_schema::BsonTypeName::String,
            BinData => json_schema::BsonTypeName::BinData,
            Undefined => json_schema::BsonTypeName::Undefined,
            ObjectId => json_schema::BsonTypeName::ObjectId,
            Boolean => json_schema::BsonTypeName::Bool,
            Date => json_schema::BsonTypeName::Date,
            Null => json_schema::BsonTypeName::Null,
            Regex => json_schema::BsonTypeName::Regex,
            DbPointer => json_schema::BsonTypeName::DbPointer,
            Javascript => json_schema::BsonTypeName::Javascript,
            Symbol => json_schema::BsonTypeName::Symbol,
            JavascriptWithScope => json_schema::BsonTypeName::JavascriptWithScope,
            Timestamp => json_schema::BsonTypeName::Timestamp,
            MinKey => json_schema::BsonTypeName::MinKey,
            MaxKey => json_schema::BsonTypeName::MaxKey,
        }
    }
}

impl TryFrom<Schema> for json_schema::Schema {
    type Error = Error;

    fn try_from(v: Schema) -> Result<Self, Self::Error> {
        Ok(match v {
            Schema::Any => json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
                max_items: None,
                any_of: None,
                one_of: None,
            },
            Schema::Unsat => json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
                max_items: None,
                any_of: Some(vec![]),
                one_of: None,
            },
            Schema::Missing => return Err(Error::InvalidBSONType("missing".to_string())),
            Schema::Atomic(a) => a.into(),
            Schema::AnyOf(ao) => json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
                max_items: None,
                any_of: Some(
                    ao.into_iter()
                        .map(json_schema::Schema::try_from)
                        .collect::<Result<_, _>>()?,
                ),
                one_of: None,
            },
            Schema::Array(a) => {
                // `a` gets moved below (i.e., at this part `json_schema::Schema::try_from(*a)?` ),
                // so doing this equality check here prevents needing to clone `a`.
                let unsat_array = *a == Schema::Unsat;

                let mut array_json_schema = json_schema::Schema {
                    bson_type: Some(json_schema::BsonType::Single(
                        json_schema::BsonTypeName::Array,
                    )),
                    properties: None,
                    required: None,
                    additional_properties: None,
                    items: Some(json_schema::Items::Single(Box::new(
                        json_schema::Schema::try_from(*a)?,
                    ))),
                    max_items: None,
                    any_of: None,
                    one_of: None,
                };

                if unsat_array {
                    array_json_schema.max_items = Some(0);
                    array_json_schema.items = None;
                }

                array_json_schema
            }
            Schema::Document(d) => json_schema::Schema::try_from(d)?,
        })
    }
}

lazy_static! {
    // The types represented by Schema::Any, unrolled into an AnyOf().
    pub static ref UNFOLDED_ANY: Schema = Schema::AnyOf(
        Atomic::into_enum_iter() // All atomic schemas.
            .map(Schema::Atomic)
            .chain(once(ANY_DOCUMENT.clone())) // Any document.
            .chain(once(ANY_ARRAY.clone())) // Any array.
            .chain(once(Schema::Missing.clone())) // Or missing.
            .collect()
    );
    // Special Document Schemas.
    pub static ref ANY_DOCUMENT: Schema = Schema::Document(Document::any());
    pub static ref EMPTY_DOCUMENT: Schema = Schema::Document(Document::empty());

    // Special Array Schemas.
    pub static ref ANY_ARRAY: Schema = Schema::Array(Box::new(Schema::Any));
    pub static ref EMPTY_ARRAY: Schema = Schema::Array(Box::new(Schema::Unsat));

    // these types
    pub static ref FALSIFIABLE_TYPES: Schema = Schema::AnyOf(set![
        Schema::Atomic(Atomic::Boolean),
        Schema::Atomic(Atomic::Integer),
        Schema::Atomic(Atomic::Long),
        Schema::Atomic(Atomic::Double),
        Schema::Atomic(Atomic::Decimal),
        Schema::Atomic(Atomic::Null),
        Schema::Missing
    ]);

    // Nullish Schemas (Schemas that additionally allow for Null or Missing).
    pub static ref NULLISH: Schema =
        Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing,]);
    pub static ref NULLISH_OR_UNDEFINED: Schema =
        Schema::AnyOf(set![Schema::Missing, Schema::Atomic(Atomic::Null), Schema::Atomic(Atomic::Undefined),]);
    pub static ref NON_NULLISH: Schema = Schema::Any.subtract_nullish();

    pub static ref ANY_ARRAY_OR_NULLISH: Schema = Schema::AnyOf(set![
        Schema::Array(Box::new(Schema::Any)),
        Schema::Atomic(Atomic::Null),
        Schema::Missing,
    ]);
    pub static ref BOOLEAN_OR_NULLISH: Schema = Schema::AnyOf(set![
        Schema::Atomic(Atomic::Boolean),
        Schema::Atomic(Atomic::Null),
        Schema::Missing,
    ]);
    pub static ref DATE_OR_NULLISH: Schema = Schema::AnyOf(set![
        Schema::Atomic(Atomic::Date),
        Schema::Atomic(Atomic::Null),
        Schema::Missing,
    ]);
    pub static ref INTEGER_OR_NULLISH: Schema = Schema::AnyOf(set![
        Schema::Atomic(Atomic::Integer),
        Schema::Atomic(Atomic::Null),
        Schema::Missing,
    ]);
    pub static ref NUMERIC: Schema = Schema::AnyOf(set![
        Schema::Atomic(Atomic::Integer),
        Schema::Atomic(Atomic::Long),
        Schema::Atomic(Atomic::Double),
        Schema::Atomic(Atomic::Decimal),
    ]);
    pub static ref NUMERIC_OR_NULLISH: Schema = Schema::AnyOf(set![
        Schema::Atomic(Atomic::Integer),
        Schema::Atomic(Atomic::Long),
        Schema::Atomic(Atomic::Double),
        Schema::Atomic(Atomic::Decimal),
        Schema::Atomic(Atomic::Null),
        Schema::Missing,
    ]);
    pub static ref INTEGER_LONG_OR_NULLISH: Schema = Schema::AnyOf(set![
        Schema::Atomic(Atomic::Integer),
        Schema::Atomic(Atomic::Long),
        Schema::Atomic(Atomic::Null),
        Schema::Missing,
    ]);
    pub static ref STRING_OR_NULLISH: Schema = Schema::AnyOf(set![
        Schema::Atomic(Atomic::String),
        Schema::Atomic(Atomic::Null),
        Schema::Missing,
    ]);
}

#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone, Copy)]
pub enum Satisfaction {
    Not,
    May,
    Must,
}

impl Satisfaction {
    /// equal_or_may returns self if other is equal, otherwise it
    /// returns Satisfaction::May.
    fn equal_or_may(self, other: Self) -> Self {
        if self == other {
            self
        } else {
            Satisfaction::May
        }
    }
}

impl Schema {
    /// returns the available schema field names.
    pub fn keys(&self) -> Vec<String> {
        match self {
            Schema::Any => vec![],
            Unsat => vec![],
            Schema::Missing => vec![],
            Schema::Atomic(_) => vec![],
            AnyOf(a) => a.iter().flat_map(|s| s.keys()).collect(),
            Schema::Array(_) => vec![],
            Schema::Document(d) => d.clone().keys.into_keys().collect(),
        }
    }

    /// returns a simplified version of this schema.
    pub fn simplify(schema: &Schema) -> Schema {
        // remove_missing removes all Missing types from the given Schema. It should
        // never be called on a Schema that Must satisfy Missing, and the argument Schema
        // must always be pre-simplified, so there is never a nested AnyOf.
        fn remove_missing(s: Schema) -> Schema {
            match s {
                Schema::Missing => unreachable!(),
                Schema::Any
                | Schema::Unsat
                | Schema::Document(_)
                | Schema::Array(_)
                | Schema::Atomic(_) => s,
                Schema::AnyOf(ao) => {
                    Schema::AnyOf(ao.into_iter().filter(|x| x != &Schema::Missing).collect())
                }
            }
        }
        match schema {
            Schema::AnyOf(a) => {
                let ret: BTreeSet<Schema> = a
                    .iter()
                    .map(Schema::simplify)
                    .flat_map(|x| match x {
                        Schema::AnyOf(vs) => vs,
                        _ => set![x],
                    })
                    .collect();
                if ret.is_empty() {
                    Unsat
                } else if ret.contains(&Schema::Any) {
                    Schema::Any
                } else if ret.len() == 1 {
                    ret.into_iter().next().unwrap()
                } else {
                    Schema::AnyOf(ret)
                }
            }
            Schema::Array(arr) => Schema::Array(Box::new(Schema::simplify(arr))),
            Schema::Document(d) => {
                let mut missing_keys = BTreeSet::new();
                Schema::Document(Document {
                    keys: d
                        .keys
                        .iter()
                        .filter_map(|(k, s)| {
                            let s = Schema::simplify(s);
                            match s.satisfies(&Schema::Missing) {
                                Satisfaction::Not => Some((k.clone(), s)),
                                Satisfaction::May => {
                                    missing_keys.insert(k.clone());
                                    Some((k.clone(), remove_missing(s)))
                                }
                                Satisfaction::Must => {
                                    missing_keys.insert(k.clone());
                                    None
                                }
                            }
                        })
                        .collect(),
                    required: d.required.difference(&missing_keys).cloned().collect(),
                    ..*d
                })
            }
            Schema::Atomic(_) => schema.clone(),
            Schema::Any => Schema::Any,
            Schema::Unsat => Schema::Unsat,
            Schema::Missing => Schema::Missing,
        }
    }

    /// schema_predicate_meet applies a schema_predicate to all passed Schemata,
    /// and takes the meet of the Satisfaction lattice defined as:
    ///
    /// Must  Not
    ///    \ /
    ///    May
    ///
    /// Thus returning:
    ///
    /// Must if the predicate returns Must for all Schemata.
    /// Not if the predicate returns Not for all Schemata.
    /// May if the predcate is not Must or Not for all Schemata.
    ///
    /// If looked at as a binary operator we get:
    /// meet(X, Y) = May
    /// meet(X, X) = X
    /// where X != Y; X, Y in {May, Must, Not}
    ///
    fn schema_predicate_meet(
        vs: &BTreeSet<Schema>,
        predicate: &dyn Fn(&Schema) -> Satisfaction,
    ) -> Satisfaction {
        vs.iter()
            .fold(None, |so_far, schema| {
                let satisfaction = predicate(schema);
                match (so_far, satisfaction) {
                    (None, s) => Some(s),
                    (Some(s1), s2) if s1 == s2 => Some(s1),
                    _ => Some(Satisfaction::May),
                }
            })
            .unwrap_or(Satisfaction::Must)
    }

    /// Satisfies AnyOf the passed set of Schemata.
    fn satisfies_any_of(&self, vs: &BTreeSet<Schema>) -> Satisfaction {
        use Satisfaction::*;
        let mut ret = Not;
        for s in vs.iter() {
            match self.satisfies(s) {
                Must => return Must,
                May => ret = May,
                Not => (),
            }
        }
        ret
    }

    /// Returns if all the possible values satisfying the self Schema also satisfy the argument
    /// other Schema.
    ///
    /// returns:
    /// Must: any value that satisfies the self Schema Must satisfy other
    /// May: any value that satisfies the self Schema May or May Not satisfy other
    /// Not: any value that satisfies the self Schema must Not satisfy the other
    pub fn satisfies(&self, other: &Schema) -> Satisfaction {
        use Satisfaction::*;
        use Schema::*;
        match (self, &other) {
            // other is Unsat or self is Unsat
            (Unsat, _) => Must,
            (AnyOf(x), Unsat) if x.is_empty() => Must,
            (_, Unsat) => Not,

            // other is Any or self is Any
            (_, Any) => Must,
            (Any, _) => May,

            // self is AnyOf
            (AnyOf(self_vs), other_s) => {
                Schema::schema_predicate_meet(self_vs, &|s: &Schema| s.satisfies(other_s))
            }

            // other is AnyOf
            (_, AnyOf(vs)) => self.satisfies_any_of(vs),

            (Atomic(self_a), Atomic(other_a)) => self_a.satisfies(other_a),
            (Atomic(_), _) => Not,

            (Array(self_arr), Array(other_arr)) => self_arr.satisfies(other_arr),
            (Array(_), _) => Not,

            (Document(self_d), Document(other_d)) => self_d.satisfies(other_d),
            (Document(_), _) => Not,

            // self is Missing
            (Missing, Missing) => Must,
            (Missing, _) => Not,
        }
    }

    /// Returns if this Schema Must, May, or must Not contain the passed field.
    pub fn contains_field(&self, field: &str) -> Satisfaction {
        self.satisfies(&Schema::Document(Document {
            keys: map! {
                field.to_string() => Schema::Any
            },
            required: set![field.to_string()],
            additional_properties: true,
            ..Default::default()
        }))
    }

    /// Returns the satisfaction result for comparing two operands.
    pub fn is_comparable_with(&self, other: &Schema) -> Satisfaction {
        use Satisfaction::*;
        use Schema::*;

        match (&self, &other) {
            // We currently disallow arrays and documents in comparisons.
            (Array(_), _) | (_, Array(_)) => Not,
            (Document(_), _) | (_, Document(_)) => Not,

            // Comparing with an Any schema will always result in a May.
            (Any, _) | (_, Any) => May,

            // Missing behaves like null, in that any type is comparable to it.
            (Missing, _) | (_, Missing) => Must,

            // Unsat behaves like null, in that any type is comparable to it.
            // However, this likely does not matter at the moment given that Unsat
            // can only be found in arrays, for which we currently disallow comparisons.
            (Unsat, _) | (_, Unsat) => Must,

            // Atomics have their own criteria for comparability involving numerics and null.
            (Atomic(a1), Atomic(a2)) => a1.is_comparable_with(a2),

            // Use the meet logic if we have an AnyOf regardless of which side the AnyOf is on,
            // since comparison is a commutative operation that type satisfaction must reflect.
            (AnyOf(anyof_vs), v) | (v, AnyOf(anyof_vs)) => {
                Schema::schema_predicate_meet(anyof_vs, &|s: &Schema| s.is_comparable_with(v))
            }
        }
    }

    /// Returns the satisfaction result for comparing a schema to itself.
    pub fn is_self_comparable(&self) -> Satisfaction {
        self.is_comparable_with(self)
    }

    // has_overlapping_keys_with returns whether any value satisfying the self Schema May, Must, or
    // must Not have overlapping keys with any value satisfying the other Schema. Either Schema may
    // be any kind of Schema, and if one of them does Not satisfy ANY_DOCUMENT, it will return Not.
    // Additionally, the EMPTY_DOCUMENT must Not have overlapping keys with any other Schema, since
    // it allows no keys.
    pub fn has_overlapping_keys_with(&self, other: &Schema) -> Satisfaction {
        match self {
            Schema::AnyOf(ao) => ao
                .iter()
                .map(|s| s.has_overlapping_keys_with(other))
                .reduce(Satisfaction::equal_or_may)
                .unwrap_or(Satisfaction::Not),
            Schema::Any => std::cmp::min(Satisfaction::May, other.satisfies(&ANY_DOCUMENT)),
            Schema::Unsat | Schema::Missing | Schema::Atomic(_) | Schema::Array(_) => {
                Satisfaction::Not
            }
            Schema::Document(d1) => match other {
                Schema::AnyOf(ao) => ao
                    .iter()
                    .map(|s| self.has_overlapping_keys_with(s))
                    .reduce(Satisfaction::equal_or_may)
                    .unwrap_or(Satisfaction::Not),
                Schema::Any => Satisfaction::May,
                Schema::Unsat | Schema::Missing | Schema::Atomic(_) | Schema::Array(_) => {
                    Satisfaction::Not
                }
                Schema::Document(d2) => d1.has_overlapping_keys_with(d2),
            },
        }
    }

    /// upconvert_missing_to_null upconverts Missing to Null in the current level
    /// of the schema including nested AnyOf's. It does not recurse into Documents or Arrays.
    /// This is used to properly handle array items Schemata, where Missing is not possible.
    pub fn upconvert_missing_to_null(self) -> Self {
        match self {
            Schema::Missing => Schema::Atomic(Atomic::Null),
            Schema::AnyOf(vs) => Schema::AnyOf(
                vs.into_iter()
                    .map(|e| e.upconvert_missing_to_null())
                    .collect(),
            ),
            // Any implicitly contains both missing and null, subtract missing from the schema to
            // upconvert missing to null in an Any schema.
            Schema::Any => Schema::Any.subtract_nullish(),
            Schema::Atomic(_) | Schema::Document(_) | Schema::Array(_) | Schema::Unsat => self,
        }
    }

    /// document_union unions together two Schemata returning a single Schema guaranteed to have
    /// variant Schema::Document.  The return Schema matches all document values matched by either
    /// `self` or `other`.  The Schema returned is not necessarily as tight a bound as
    /// `AnyOf([self, other])`; in other words, it may match additional document values not matched
    /// by `self` or `other`.
    pub fn document_union(self, other: Schema) -> Schema {
        match self {
            Schema::AnyOf(ao) => ao.into_iter().fold(other, Schema::document_union),
            Schema::Any
            | Schema::Unsat
            | Schema::Missing
            | Schema::Atomic(_)
            | Schema::Array(_) => EMPTY_DOCUMENT.clone(),
            Schema::Document(ref d1) => match other {
                Schema::AnyOf(_) => other.document_union(self),
                Schema::Document(d2) => Schema::Document(d1.clone().union(d2)),
                Schema::Any
                | Schema::Unsat
                | Schema::Missing
                | Schema::Atomic(_)
                | Schema::Array(_) => EMPTY_DOCUMENT.clone(),
            },
        }
    }

    /// get_single_field_name_and_schema returns `Some((fieldName, fieldSchema))` if every value
    /// matched by `self` which is not the empty document is a document containing a single field
    /// called `fieldName`.
    /// If this is not the case, or `self` doesn't match any values, it returns `None`.
    pub fn get_single_field_name_and_schema(&self) -> Option<(&str, Schema)> {
        match self {
            AnyOf(any_of) => {
                let mut fields = BTreeMap::<&str, Schema>::new();
                let found_schema_without_single_field = any_of
                    .iter()
                    .filter(|schema| !matches!(schema, Unsat))
                    .map(|schema| schema.get_single_field_name_and_schema())
                    .fold(false, |acc, field_info| match field_info {
                        Some((field_name, field_schema)) => {
                            let sch = fields
                                .get(field_name)
                                .map(|s| s.union(&field_schema))
                                .unwrap_or(field_schema);
                            fields.insert(field_name, sch);
                            acc
                        }
                        None => true,
                    });
                if found_schema_without_single_field {
                    return None;
                }
                match fields.len() {
                    1 => fields.into_iter().next(),
                    _ => None,
                }
            }
            Schema::Document(d) => match d.num_keys() {
                (num_required, Some(1)) if num_required <= 1 => {
                    d.keys.iter().next().map(|(field_name, field_schema)| {
                        (field_name.as_str(), field_schema.clone())
                    })
                }
                _ => None,
            },
            _ => None,
        }
    }

    /// Set-subtracts Null and Missing from the given schema. Ensures that for every schema `S1`,
    /// `(S1.subtract_nullish() == Unsat) or S1.subtract_nullish().satisfies(AnyOf(Null, Missing)) == Not`
    pub fn subtract_nullish(self) -> Schema {
        let nullish = &NULLISH.clone();
        match self {
            AnyOf(schemas) => AnyOf(
                schemas
                    .into_iter()
                    .filter(|schema| schema.satisfies(nullish) != Satisfaction::Must)
                    .map(|schema| schema.subtract_nullish())
                    .collect(),
            ),
            Schema::Any => UNFOLDED_ANY.clone().subtract_nullish(),
            schema => {
                // If the schemas overlap fully, then their difference is empty.
                if schema.satisfies(nullish) == Satisfaction::Must {
                    Unsat
                } else {
                    schema
                }
            }
        }
    }

    /// enumerate_field_paths exhaustively enumerates all field paths
    /// of length <= `max_length` that could exist in a value matched
    /// by the schema `self`, which can be any kind of Schema. If it
    /// cannot exhaustively enumerate all field paths (e.g. `self` is the
    /// `Any` Schema, or additional properties are allowed), it returns
    /// an error. Additionally, a boolean value is returned indicating if
    /// the schema has only nullable polymorphism, if at all polymorphic. A value of `true`
    /// indicates that the schema represents only an Object or an Object
    /// that is possibly Null and/or Missing, i.e. "nullable".
    ///
    /// Example:
    ///
    /// If `self` describes documents with the shape {'a': {'b': {'c': 1}}},
    /// enumerate_field_paths will return one of the following:
    ///
    /// - if max_length = Some(0), set{}
    /// - if max_length = Some(1), set{['a']}
    /// - if max_length = Some(2), set{['a', 'b']}
    /// - if max_length = Some(d), where d >= 3, or None, set{['a', 'b', 'c']}
    pub fn enumerate_field_paths(
        &self,
        max_length: Option<u32>,
    ) -> Result<(BTreeSet<Vec<String>>, bool), Error> {
        // Call auxiliary function with parameter `inside_document` set to false because
        // function is not yet in the context of a document.
        self.enumerate_field_paths_aux(max_length, false)
    }

    /// enumerate_field_paths_aux is the auxiliary function called by
    /// enumerate_field_paths to exhaustively enumerate all field paths
    /// of length <= `max_length` that could exist in a value matched
    /// by the schema `self`. Additionally, this function determines what
    /// the boolean value returned by enumerate_field_paths should be.
    /// If `inside_document` is true, `self` is a schema within a
    /// Schema::Document.
    ///
    /// Examples:
    ///
    /// If `self` is a schema AnyOf([Int, Double]) and `inside_document` is false, then
    /// enumerate_field_paths_aux returns the empty set `set![].
    ///
    /// If `self` is a schema AnyOf([Int, Double]) and `inside_document` is true, then
    /// enumerate_field_paths_aux returns the set containing an empty vector `set![vec![]]``.
    #[allow(clippy::manual_try_fold)]
    fn enumerate_field_paths_aux(
        &self,
        max_length: Option<u32>,
        inside_document: bool,
    ) -> Result<(BTreeSet<Vec<String>>, bool), Error> {
        match self {
            Schema::Document(d) => {
                // Error if we do not have complete schema information
                if d.additional_properties {
                    return Err(Error::CannotEnumerateAllFieldPaths(self.clone()));
                }
                d.keys.clone().into_iter().fold(
                    Ok((BTreeSet::new(), true)),
                    |acc, (key, schema)| match max_length {
                        Some(0) => acc,
                        _ => {
                            let (mut new_paths, has_only_nullable_polymorphism) = schema
                                .enumerate_field_paths_aux(max_length.map(|l| l - 1), true)?;
                            if new_paths.is_empty() {
                                new_paths = set![vec![]];
                            }
                            let mut acc = acc?;
                            acc.0.extend(
                                new_paths
                                    .into_iter()
                                    .map(|path| {
                                        vec![key.clone()].into_iter().chain(path).collect_vec()
                                    })
                                    .collect::<BTreeSet<Vec<String>>>(),
                            );
                            Ok((acc.0, acc.1 && has_only_nullable_polymorphism))
                        }
                    },
                )
            }
            AnyOf(a) => {
                let mut mut_copy_of_a = a.clone();
                let mut object_found = false;

                let combined_doc = a.iter().fold(EMPTY_DOCUMENT.clone(), |acc, schema| {
                    let acc = acc.document_union(schema.clone());
                    if let Schema::Document(_) = schema {
                        object_found = true;
                        // After combining all the documents into one, it is no longer necessary to deal with each individual document,
                        // so we can remove them after using them in the document_union above.
                        mut_copy_of_a.remove(schema);
                    }
                    acc
                });

                // We don't want to add a document unless there already is one in the AnyOf because we could potentially create a polymorphic object.
                // If there already is a document in the AnyOf, adding another one will not create any new consequences.
                if object_found {
                    mut_copy_of_a.insert(combined_doc);
                }

                let mut non_null_or_missing_found = false;
                mut_copy_of_a.iter().fold(
                    Ok((BTreeSet::new(), true)),
                    |acc: Result<(BTreeSet<Vec<String>>, bool), Error>, schema| {
                        let acc = acc?;

                        if !matches!(
                            schema,
                            Schema::Missing | Schema::Atomic(Atomic::Null) | Schema::Document(_)
                        ) {
                            non_null_or_missing_found = true;
                        }

                        let (mut new_paths, has_only_nullable_polymorphism) =
                            schema.enumerate_field_paths_aux(max_length, inside_document)?;
                        // If we are in the context of a document, propagate an empty vector to
                        // recursively build document field paths
                        if new_paths.is_empty() && inside_document {
                            new_paths = set![vec![]]
                        }
                        Ok((
                            acc.0
                                .union(&new_paths)
                                .cloned()
                                .collect::<BTreeSet<Vec<String>>>(),
                            // If we encountered an object and any non-nullable schema in this AnyOf,
                            // then we know the schema at this field path is polymorphic with non-null values.
                            !(object_found && non_null_or_missing_found)
                                && has_only_nullable_polymorphism
                                && acc.1,
                        ))
                    },
                )
            }
            Schema::Any => Err(Error::CannotEnumerateAllFieldPaths(Schema::Any)),
            Schema::Array(_) | Schema::Atomic(_) | Schema::Missing | Schema::Unsat => {
                Ok((set![], true))
            }
        }
    }

    /// union unions two schemata. The idea is that the two schema both represent data in a given
    /// collection, and that by combining them we have a more clear picture of the possibilities
    /// for a given collection. The uniond schema must match all values matched by the two original
    /// schemata, conceptually this is a set union, so a safe fall back is AnyOf of the two
    /// original schemata, but we can do better for specific cases, for example, two documents can
    /// simply union the keys (and the types of the keys when keys overlap), intersect the
    /// required, and do a lattice join over additional_properties (if one document is true, and
    /// the other is false, the solution is true).
    pub fn union(&self, other: &Schema) -> Schema {
        use std::cmp::Ordering;
        use Schema::*;
        let (left, right) = (Self::simplify(self), Self::simplify(other));
        let ordering = left.cmp(&right);
        let (left, right) = match ordering {
            Ordering::Greater => (right, left),
            Ordering::Less => (left, right),
            Ordering::Equal => {
                return left;
            }
        };
        // Schema types ordered least to greatest. We use the order to reduce the number
        // of cases needed to match. For instance, Unsat will always be leftmost and Any will
        // always be rightmost, so we do not need to check symmetric cases (and catchall will
        // get the reflexive case).
        //
        // Unsat
        // Missing
        // Atomic(Atomic)
        // Array(Box<Schema>)
        // Document(Document)
        // AnyOf(BTreeSet<Schema>)
        // Any
        match (left, right) {
            (Unsat, s) => s,
            (_, Any) => Any,
            (AnyOf(mut b1), AnyOf(b2)) => {
                // this is equivalent to a destructive union
                b1.extend(b2);
                AnyOf(b1)
            }
            (Array(s1), Array(s2)) => Array(Box::new(s1.union(s2.as_ref()))),
            (Array(s1), AnyOf(schemas)) => {
                let (arrays, mut rest): (BTreeSet<_>, BTreeSet<_>) =
                    schemas.into_iter().partition(|s| matches!(s, Array(_)));
                if arrays.is_empty() {
                    rest.insert(Array(s1));
                } else if arrays.len() > 1 {
                    rest.insert(Array(Box::new(Schema::Any)));
                } else if let Some(Array(old_s)) = arrays.into_iter().next() {
                    rest.insert(Array(Box::new(old_s.as_ref().union(s1.as_ref()))));
                } else {
                    unreachable!();
                }
                AnyOf(rest)
            }
            (Document(d1), Document(d2)) => Document(d1.union(d2)),
            (Document(d), AnyOf(schemas)) => {
                let (documents, mut rest): (BTreeSet<_>, BTreeSet<_>) =
                    schemas.into_iter().partition(|s| matches!(s, Document(_)));
                if documents.is_empty() {
                    rest.insert(Document(d));
                } else if let Some(Document(old_d)) = documents.into_iter().next() {
                    rest.insert(Document(old_d.union(d)));
                } else {
                    unreachable!();
                }
                AnyOf(rest)
            }
            // x (strictly) < AnyOf
            (x, AnyOf(mut b)) => {
                b.insert(x);
                AnyOf(b)
            }
            (s1, s2) => AnyOf(set! {s1, s2}),
        }
    }

    /// This helper takes a BTreeSet -- the intersection of each member of an AnyOf with another schema --
    /// and returns a schema
    fn schema_from_anyof_intersection(&self, intersection: BTreeSet<Schema>) -> Schema {
        if intersection.is_empty() {
            Schema::Unsat
        } else {
            Schema::simplify(&Schema::AnyOf(intersection))
        }
    }

    /// The intersection of two Schemas S and T is the maximal schema R
    /// such that R.satisfies(S) == R.satisfies(T) == must. Knowing this is useful for schema
    /// derivation, where we can use intersection to combine the implied schemas of $match filters
    pub fn intersection(&self, other: &Schema) -> Schema {
        let self_schema = Schema::simplify(self);
        let other_schema = Schema::simplify(other);
        match (self_schema, other_schema) {
            (Schema::Any, schema) | (schema, Schema::Any) => schema,
            (Schema::Missing, Schema::Missing) => Schema::Missing,
            (Schema::Atomic(a), Schema::Atomic(b)) => {
                if a == b {
                    self.clone()
                } else {
                    Schema::Unsat
                }
            }
            (atomic @ Schema::Atomic(_), Schema::AnyOf(anyof))
            | (Schema::AnyOf(anyof), atomic @ Schema::Atomic(_)) => {
                if anyof.contains(&atomic) {
                    atomic.clone()
                } else {
                    Schema::Unsat
                }
            }
            (Schema::Missing, Schema::AnyOf(anyof)) | (Schema::AnyOf(anyof), Schema::Missing) => {
                if anyof.contains(&Schema::Missing) {
                    Schema::Missing
                } else {
                    Schema::Unsat
                }
            }
            (Schema::Array(a), Schema::Array(b)) => match a.intersection(&b) {
                Schema::Unsat => Schema::Unsat,
                schema => Schema::Array(Box::new(schema)),
            },
            (schema @ Schema::Array(_), Schema::AnyOf(anyof))
            | (schema @ Schema::Document(_), Schema::AnyOf(anyof))
            | (Schema::AnyOf(anyof), schema @ Schema::Array(_))
            | (Schema::AnyOf(anyof), schema @ Schema::Document(_)) => {
                let ret: BTreeSet<Schema> = anyof
                    .iter()
                    .map(|anyof_item| schema.intersection(anyof_item))
                    .filter(|intersection: &Schema| intersection != &Schema::Unsat)
                    .collect();
                self.schema_from_anyof_intersection(ret)
            }
            (Schema::Document(a), Schema::Document(b)) => {
                let mut doc_intersection = Document::default();
                a.keys.clone().into_iter().for_each(|(key, schema)| {
                    if let Some(b_schema) = b.keys.get(&key) {
                        match schema.intersection(b_schema) {
                            Schema::Unsat => {}
                            intersection => {
                                doc_intersection.keys.insert(key.clone(), intersection);
                                if a.required.contains(&key) && b.required.contains(&key) {
                                    doc_intersection.required.insert(key);
                                }
                            }
                        }
                    }
                });
                if doc_intersection.keys.is_empty() {
                    Schema::Unsat
                } else {
                    Schema::Document(doc_intersection)
                }
            }
            (Schema::AnyOf(a), Schema::AnyOf(b)) => {
                let ret: BTreeSet<Schema> = a
                    .iter()
                    .map(|anyof_item| {
                        let ret: BTreeSet<Schema> = b
                            .iter()
                            .map(|item| anyof_item.intersection(item))
                            .filter(|intersection: &Schema| intersection != &Schema::Unsat)
                            .collect();
                        self.schema_from_anyof_intersection(ret)
                    })
                    .filter(|intersection: &Schema| intersection != &Schema::Unsat)
                    .collect();
                self.schema_from_anyof_intersection(ret)
            }
            // if there is no intersection between the two schemas, we will use Unsat to represent
            // that no such schema R exists.
            _ => Schema::Unsat,
        }
    }
}

impl From<Type> for Schema {
    fn from(t: Type) -> Self {
        use Type::*;
        match t {
            Array => ANY_ARRAY.clone(),
            BinData => Schema::Atomic(Atomic::BinData),
            Boolean => Schema::Atomic(Atomic::Boolean),
            Datetime => Schema::Atomic(Atomic::Date),
            DbPointer => Schema::Atomic(Atomic::DbPointer),
            Decimal128 => Schema::Atomic(Atomic::Decimal),
            Document => ANY_DOCUMENT.clone(),
            Double => Schema::Atomic(Atomic::Double),
            Int32 => Schema::Atomic(Atomic::Integer),
            Int64 => Schema::Atomic(Atomic::Long),
            Javascript => Schema::Atomic(Atomic::Javascript),
            JavascriptWithScope => Schema::Atomic(Atomic::JavascriptWithScope),
            MaxKey => Schema::Atomic(Atomic::MaxKey),
            MinKey => Schema::Atomic(Atomic::MinKey),
            Null => Schema::Atomic(Atomic::Null),
            ObjectId => Schema::Atomic(Atomic::ObjectId),
            RegularExpression => Schema::Atomic(Atomic::Regex),
            String => Schema::Atomic(Atomic::String),
            Symbol => Schema::Atomic(Atomic::Symbol),
            Timestamp => Schema::Atomic(Atomic::Timestamp),
            Undefined => Schema::Atomic(Atomic::Null),
        }
    }
}

impl From<TypeOrMissing> for Schema {
    fn from(t: TypeOrMissing) -> Self {
        use TypeOrMissing::*;
        match t {
            Missing => Schema::Missing,
            Type(t) => Schema::from(t),
            Number => Schema::AnyOf(set![
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
                Schema::Atomic(Atomic::Double)
            ]),
        }
    }
}

impl TryFrom<json_schema::Schema> for Schema {
    type Error = Error;

    /// from converts a json schema into a MongoSQL schema by following these rules:
    ///      - BsonType::Single => Schema::Atomic
    ///      - BsonType::Multiple => Schema::AnyOf
    ///      - properties, required, and additional_properties => Schema::Document
    ///      - items => Schema::Array
    ///      - any_of => Schema::AnyOf
    ///
    /// any_of and one_of are the only fields that are mutually exclusive with the rest.
    fn try_from(v: json_schema::Schema) -> Result<Self, Self::Error> {
        // Explicitly match the valid combinations of JSON schema fields
        match v {
            // The empty JSON schema is equivalent to `Any`. This would
            // technically be handled correctly by the following branch, but
            // this special case makes for a cleaner conversion (`Any` instead
            // of an `AnyOf` representing all possible values of each type).
            json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
                max_items: None,
                any_of: None,
                one_of: None,
            } => Ok(Schema::Any),

            // This branch handles all other JSON schema validators that don't
            // use `one_of` or `any_of`. We must explicitly filter out unsupported
            // BsonTypeName values we don't support.
            json_schema::Schema {
                bson_type,
                properties,
                required,
                additional_properties,
                items,
                max_items,
                any_of: None,
                one_of: None,
            } => {
                let bson_type = bson_type.unwrap_or_else(|| {
                    json_schema::BsonType::Multiple(
                        json_schema::BsonTypeName::into_enum_iter()
                            .filter(|&t| t != json_schema::BsonTypeName::Undefined)
                            .collect(),
                    )
                });
                match bson_type {
                    json_schema::BsonType::Single(json_schema::BsonTypeName::Array) => {
                        Ok(Schema::Array(Box::new(match items {
                            // The single-schema variant of the `items`
                            // field constrains all elements of the array.
                            Some(json_schema::Items::Single(i)) => Schema::try_from(*i)?,
                            // The multiple-schema variant of the `items`
                            // field only asserts the schemas for the
                            // array items at specified indexes, and
                            // imposes no constraint on items at larger
                            // indexes. As such, the only schema that can
                            // describe all elements of the array is
                            // `Any`.
                            Some(json_schema::Items::Multiple(_)) => Schema::Any,
                            // No `items` field means no constraints on
                            // array elements.
                            None => {
                                if max_items == Some(0) {
                                    Schema::Array(Box::new(Schema::Unsat))
                                } else {
                                    Schema::Any
                                }
                            }
                        })))
                    }
                    json_schema::BsonType::Single(json_schema::BsonTypeName::Object) => {
                        Ok(Schema::Document(Document::try_from(json_schema::Schema {
                            properties,
                            required,
                            additional_properties,
                            ..Default::default()
                        })?))
                    }
                    json_schema::BsonType::Single(typ) => {
                        Ok(Schema::Atomic(Atomic::try_from(typ)?))
                    }
                    json_schema::BsonType::Multiple(m) => {
                        // For each value in `bson_type`, construct a json_schema::Schema that only
                        // contains the single type and any relevant fields and recursively call
                        // Schema::try_from on it. Then, wrap the resulting vector in a Schema::AnyOf
                        // and call `simplify` in order to remove any unnecessary AnyOf wrappings.
                        Ok(Schema::simplify(&AnyOf(
                            m.into_iter()
                                .map(|bson_type| match bson_type {
                                    json_schema::BsonTypeName::Array => {
                                        Schema::try_from(json_schema::Schema {
                                            bson_type: Some(json_schema::BsonType::Single(
                                                bson_type,
                                            )),
                                            items: items.clone(),
                                            max_items,
                                            ..Default::default()
                                        })
                                    }
                                    json_schema::BsonTypeName::Object => {
                                        Schema::try_from(json_schema::Schema {
                                            bson_type: Some(json_schema::BsonType::Single(
                                                bson_type,
                                            )),
                                            properties: properties.clone(),
                                            required: required.clone(),
                                            additional_properties,
                                            ..Default::default()
                                        })
                                    }
                                    _ => Schema::try_from(json_schema::Schema {
                                        bson_type: Some(json_schema::BsonType::Single(bson_type)),
                                        ..Default::default()
                                    }),
                                })
                                .collect::<Result<BTreeSet<Schema>, _>>()?,
                        )))
                    }
                }
            }
            json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
                max_items: None,
                any_of: Some(any_of),
                one_of: None,
            } => Ok(Schema::AnyOf(
                any_of
                    .into_iter()
                    .map(Schema::try_from)
                    .collect::<Result<BTreeSet<Schema>, _>>()?,
            )),
            json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
                max_items: None,
                any_of: None,
                one_of: Some(one_of),
                // convert one_of to any_of
            } => Ok(Schema::AnyOf(
                one_of
                    .into_iter()
                    .map(Schema::try_from)
                    .collect::<Result<BTreeSet<Schema>, _>>()?,
            )),
            _ => Err(Error::InvalidCombinationOfFields()),
        }
    }
}

impl Atomic {
    /// satisfies returns whether one atomic satisfies another atomic (Must or Not only).
    pub fn satisfies(&self, other: &Self) -> Satisfaction {
        if self == other {
            Satisfaction::Must
        } else {
            Satisfaction::Not
        }
    }

    /// is_comparable_with returns whether or not two atomics are comparable (Must or Not only).
    /// Atomics are comparable if they are both numeric, if either is null,
    /// or otherwise both equal.
    pub fn is_comparable_with(&self, other: &Self) -> Satisfaction {
        use self::Atomic::*;
        use Satisfaction::*;

        match (self, other) {
            (Null, _) | (_, Null) => Must,
            // DbPointer, Javascript, and JavascriptWithScope are not comparable with any other type except NULL
            (DbPointer, _) | (_, DbPointer) => Not,
            (Javascript, _) | (_, Javascript) => Not,
            (JavascriptWithScope, _) | (_, JavascriptWithScope) => Not,
            (l, r) if l == r || l.is_numeric() && r.is_numeric() => Must,
            _ => Not,
        }
    }

    /// is_numeric returns whether or not the atomic value is numeric.
    pub fn is_numeric(&self) -> bool {
        use self::Atomic::*;
        match self {
            Decimal | Double | Integer | Long => true,
            String | BinData | Undefined | ObjectId | Boolean | Date | Null | Regex | DbPointer
            | Javascript | Symbol | JavascriptWithScope | Timestamp | MinKey | MaxKey => false,
        }
    }
}

impl Document {
    /// any returns an Any Document, that is a Document that may contain any
    /// keys of Any Schema
    pub fn any() -> Document {
        Document {
            additional_properties: true,
            ..Default::default()
        }
    }

    /// empty returns an Empty Document
    pub fn empty() -> Document {
        Document::default()
    }

    /// satisfies returns whether one Document Schema satisfies another Document Schema.
    fn satisfies(&self, other: &Self) -> Satisfaction {
        use Satisfaction::*;
        let mut ret = Must;
        // First if the other Schema does not allow additional_properties, we must make
        // sure self does not allow properties not allowed by other Schema.
        if !other.additional_properties {
            if self
                .required
                .iter()
                .any(|key| !(other.keys.contains_key(key) || other.required.contains(key)))
            {
                return Not;
            }
            if self.additional_properties
                || self
                    .keys
                    .iter()
                    .any(|(key, _)| !(other.keys.contains_key(key) || other.required.contains(key)))
            {
                ret = May;
            }
        }

        // Next check the Schema for the key in self satisfies the
        // Schema for that key in other, for all the keys in other.
        for (key, other_key_schema) in other.keys.iter() {
            let self_key_schema = match self.keys.get(key) {
                None => {
                    if !self.additional_properties {
                        &Schema::Missing
                    } else {
                        &Schema::Any
                    }
                }
                Some(schema) => schema,
            };
            match self_key_schema.satisfies(other_key_schema) {
                Not => return Not,
                May => ret = May,
                Must => (),
            }
        }
        // At this point, all the key Schemata either Must or May satisfy, now
        // we must check that all the required keys must be present.
        for key in other.required.iter() {
            if !self.required.contains(key) {
                if !(self.keys.contains_key(key) || self.additional_properties) {
                    // It is impossible to satisfy one of the required keys.
                    return Not;
                }
                // One of the required keys is not required in self, so
                // the best we can say is that self May satisfy.
                ret = May;
            }
        }
        ret
    }

    /// union_keys constructs a key map where all the keys from both maps are kept.
    /// Those keys that overlap have their Schemata merged.
    fn union_keys(
        mut m1: BTreeMap<String, Schema>,
        m2: BTreeMap<String, Schema>,
    ) -> BTreeMap<String, Schema> {
        for (key2, schema2) in m2.into_iter() {
            if let Some(old_schema) = m1.remove(&key2) {
                m1.insert(key2, old_schema.union(&schema2));
            } else {
                m1.insert(key2, schema2);
            }
        }
        m1
    }

    /// intersect_keys constructs a key map that is the intersection of the
    /// two passed maps.
    #[allow(dead_code)]
    fn intersect_keys(
        m1: BTreeMap<String, Schema>,
        mut m2: BTreeMap<String, Schema>,
    ) -> BTreeMap<String, Schema> {
        let mut out = BTreeMap::new();
        for (key, s1) in m1.into_iter() {
            if let Some(s2) = m2.remove(&key) {
                out.insert(key, s1.union(&s2));
            }
        }
        out
    }

    /// retain_keys retains keys from the m1 map argument, creating an AnyOf for the Schema of any
    /// that overlap, and ignoring the keys from the m2 map that are not overlapping with m1.
    #[allow(dead_code)]
    fn retain_keys(
        mut m1: BTreeMap<String, Schema>,
        m2: BTreeMap<String, Schema>,
    ) -> BTreeMap<String, Schema> {
        for (key, s1) in m2.into_iter() {
            if let Some(s2) = m1.remove(&key) {
                m1.insert(key, s1.union(&s2));
            }
        }
        m1
    }

    /// determines whether there is a jaccard index for the document(s),
    /// and returns the index
    fn get_jaccard_index(left: &Document, right: &Document) -> Option<JaccardIndex> {
        // no work needed if neither document has a jaccard index
        if left.jaccard_index.is_none() && right.jaccard_index.is_none() {
            return None;
        }
        // as best we can, we will average the two rates
        let rate1 = left.jaccard_index.unwrap_or_default();
        let rate2 = right.jaccard_index.unwrap_or_default();
        let num_unions = rate1.num_unions + rate2.num_unions;
        let avg_ji = (rate1.avg_ji * rate1.num_unions as f64
            + rate2.avg_ji * rate2.num_unions as f64)
            / num_unions as f64;
        // if we have a NaN, we've divided by zero. We will set the initial average
        // to 1.0
        let avg_ji = if avg_ji.is_finite() { avg_ji } else { 1.0 };

        JaccardIndex {
            avg_ji,
            num_unions,
            // the stability limit is going to be constant for documents we see based on the initial
            // stability rate the user chooses, or the default, so we just use rate1
            stability_limit: rate1.stability_limit,
        }
        .into()
    }

    // https://en.wikipedia.org/wiki/Jaccard_index
    // This might be further improved by using minhash
    fn update_jaccard_index(
        ji: JaccardIndex,
        union_size: usize,
        intersection_size: usize,
    ) -> JaccardIndex {
        if ji.num_unions == 0 {
            return JaccardIndex {
                num_unions: 1,
                avg_ji: intersection_size as f64 / union_size as f64,
                stability_limit: ji.stability_limit,
            };
        }
        let new_jaccard_index = intersection_size as f64 / union_size as f64;
        let new_avg_ji =
            (ji.avg_ji * ji.num_unions as f64 + new_jaccard_index) / (ji.num_unions + 1) as f64;
        JaccardIndex {
            num_unions: ji.num_unions + 1,
            avg_ji: new_avg_ji,
            stability_limit: ji.stability_limit,
        }
    }

    /// union unions together two Schema::Documents returning a single Document schema that matches
    /// all document values matched by either `self` or `other`. Additional properties will
    /// be allowed if either `self` or `other` allows them.
    ///
    /// If either of the two documents has a JaccardIndex, the union calculates the moving
    /// average. Once the 5th union is reached, each union will inspect the index
    /// and return a Document that allows any properties if the index is less than the
    /// stability_limit specified in the JaccardIndex struct. Prior to comparison, the inverse
    /// of the numer of unions is added to the average Jaccard index to allow for some variance,
    /// becoming more sensitive as more unions are processed. Documents that are a subset or superset
    /// of each other will be considered equivalent and will always have a JaccardIndex of 1.
    pub fn union(self, other: Document) -> Document {
        if self == Document::any() || other == Document::any() {
            return Document::any();
        }
        if let Some(jaccard_index) = Document::get_jaccard_index(&self, &other) {
            let union = Document::union_keys(self.keys.clone(), other.keys.clone());
            let left_keys = self.keys.keys().collect::<HashSet<_>>();
            let right_keys = other.keys.keys().collect::<HashSet<_>>();
            let intersection_count = left_keys.intersection(&right_keys).count();
            // if the left keys are a subset of the right keys, or vice versa, then we will consider
            // then equivalent documents
            let intersection_size =
                if left_keys.is_subset(&right_keys) || left_keys.is_superset(&right_keys) {
                    union.len()
                } else {
                    intersection_count
                };

            let jaccard_index =
                Document::update_jaccard_index(jaccard_index, union.len(), intersection_size);

            // as the number of unions grows, this number will become smaller and smaller
            let stabilization_rate = 1.0 / jaccard_index.num_unions as f64;
            if jaccard_index.num_unions >= 5
                && (jaccard_index.avg_ji + stabilization_rate) < jaccard_index.stability_limit
            {
                Document::any()
            } else {
                Document {
                    keys: union,
                    required: self
                        .required
                        .intersection(&other.required)
                        .cloned()
                        .collect(),
                    additional_properties: self.additional_properties
                        || other.additional_properties,
                    jaccard_index: Some(jaccard_index),
                }
            }
        } else {
            Document {
                keys: Document::union_keys(self.keys, other.keys),
                required: self
                    .required
                    .intersection(&other.required)
                    .cloned()
                    .collect(),
                additional_properties: self.additional_properties || other.additional_properties,
                ..Default::default()
            }
        }
    }

    /// has_overlapping_keys_with returns whether any Document value satisfying the self Document
    /// Schema May, Must, or must Not have overlapping keys with any value satisfying the other
    /// Document Schema.
    fn has_overlapping_keys_with(&self, other: &Document) -> Satisfaction {
        // the empty document schema cannot overlap with any other document, even if the other
        // document allows additional_properties.
        if self.is_empty() || other.is_empty() {
            return Satisfaction::Not;
        }
        if self.required.intersection(&other.required).next().is_some() {
            return Satisfaction::Must;
        }
        if self.additional_properties || other.additional_properties {
            return Satisfaction::May;
        }
        if self
            .keys
            .keys()
            .collect::<BTreeSet<_>>()
            .intersection(&other.keys.keys().collect::<BTreeSet<_>>())
            .next()
            .is_some()
        {
            return Satisfaction::May;
        }
        Satisfaction::Not
    }

    /// Merge two documents to produce a new document. Unlike `union()`, documents which
    /// satisfy one of the input schemas will not satisfy the resulting schema unless one is a
    /// subset of the other.
    pub fn merge(self, other: Document) -> Document {
        let jaccard_index = Document::get_jaccard_index(&self, &other);
        Document {
            keys: Document::union_keys(self.keys, other.keys),
            required: self.required.into_iter().chain(other.required).collect(),
            additional_properties: self.additional_properties || other.additional_properties,
            jaccard_index,
        }
    }

    #[inline]
    fn is_empty(&self) -> bool {
        self.keys.is_empty() && self.required.is_empty() && !self.additional_properties
    }
}
