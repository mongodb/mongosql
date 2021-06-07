use crate::{
    ir::{binding_tuple::BindingTuple, schema::Error, Type},
    json_schema, map,
    schema::Schema::AnyOf,
    set,
};
use lazy_static::lazy_static;
use std::collections::{BTreeMap, BTreeSet};
use std::convert::TryFrom;
use std::str::FromStr;

pub type SchemaEnvironment = BindingTuple<Schema>;

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct ResultSet {
    pub schema_env: SchemaEnvironment,
    pub min_size: Option<u64>,
    pub max_size: Option<u64>,
}

impl Default for ResultSet {
    fn default() -> Self {
        Self {
            schema_env: SchemaEnvironment::default(),
            min_size: None,
            max_size: None,
        }
    }
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone)]
pub enum Schema {
    Any,
    Missing,
    Atomic(Atomic),
    OneOf(Vec<Schema>),
    AnyOf(Vec<Schema>),
    Array(Box<Schema>),
    Document(Document),
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone, Copy)]
pub enum Atomic {
    String,
    Integer,
    Double,
    Long,
    Decimal,
    BinData,
    ObjectId,
    Boolean,
    Date,
    Null,
    Regex,
    DbPointer,
    Javascript,
    Symbol,
    JavascriptWithScope,
    Timestamp,
    MinKey,
    MaxKey,
}

impl FromStr for Atomic {
    type Err = Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "string" => Ok(Atomic::String),
            "int" => Ok(Atomic::Integer),
            "double" => Ok(Atomic::Double),
            "long" => Ok(Atomic::Long),
            "decimal" => Ok(Atomic::Decimal),
            "binData" => Ok(Atomic::BinData),
            "objectId" => Ok(Atomic::ObjectId),
            "bool" => Ok(Atomic::Boolean),
            "date" => Ok(Atomic::Date),
            "null" => Ok(Atomic::Null),
            "regex" => Ok(Atomic::Regex),
            "dbPointer" => Ok(Atomic::DbPointer),
            "javascript" => Ok(Atomic::Javascript),
            "symbol" => Ok(Atomic::Symbol),
            "javascriptWithScope" => Ok(Atomic::JavascriptWithScope),
            "timestamp" => Ok(Atomic::Timestamp),
            "minKey" => Ok(Atomic::MinKey),
            "maxKey" => Ok(Atomic::MaxKey),
            _ => Err(Error::InvalidJsonSchema(format!(
                "{} is not a valid BSON type",
                s
            ))),
        }
    }
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone)]
pub struct Document {
    pub keys: BTreeMap<String, Schema>,
    pub required: BTreeSet<String>,
    pub additional_properties: bool,
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
                .map(|(key, schema)| Ok((key, Schema::try_from(schema)?)))
                .collect::<Result<_, _>>()?,
            required: v
                .required
                .unwrap_or_default()
                .into_iter()
                .collect::<BTreeSet<String>>(),
            additional_properties: v.additional_properties.unwrap_or(true),
        })
    }
}

lazy_static! {
    pub static ref ANY_DOCUMENT: Schema = Schema::Document(Document {
        keys: BTreeMap::new(),
        required: BTreeSet::new(),
        additional_properties: true
    });
    pub static ref ANY_ARRAY: Schema = Schema::Array(Box::new(Schema::Any));
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum Satisfaction {
    Must,
    Not,
    May,
}

#[allow(dead_code, unused_variables)]
impl Schema {
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
        vs: &[Schema],
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

    /// satisfies AnyOf the passed set of Schemata.
    fn satisfies_any_of(&self, vs: &[Schema]) -> Satisfaction {
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

    /// satisfies exactly OneOf the passed set of Schemata.
    fn satisfies_one_of(&self, vs: &[Schema]) -> Satisfaction {
        use Satisfaction::*;
        let mut ret = Not;
        for s in vs.iter() {
            match self.satisfies(s) {
                May => return May,
                Must => {
                    if ret == Must {
                        return Not;
                    } else {
                        ret = Must;
                    }
                }
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
            // other is Any or self is Any
            (_, Any) => Must,
            (Any, _) => May,

            // self is AnyOf or OneOf
            (AnyOf(self_vs), other_s) => {
                Schema::schema_predicate_meet(self_vs, &|s: &Schema| s.satisfies(other_s))
            }
            (OneOf(self_vs), other_s) => {
                Schema::schema_predicate_meet(self_vs, &|s: &Schema| s.satisfies(other_s))
            }
            // other is AnyOf or OneOf
            (_, AnyOf(vs)) => self.satisfies_any_of(vs),
            (_, OneOf(vs)) => self.satisfies_one_of(vs),

            (Atomic(self_a), Atomic(other_a)) => self_a.satisfies(&other_a),
            (Atomic(_), _) => Not,

            (Array(self_arr), Array(other_arr)) => self_arr.satisfies(&*other_arr),
            (Array(_), _) => Not,

            (Document(self_d), Document(other_d)) => self_d.satisfies(&other_d),
            (Document(_), _) => Not,

            // self is Missing
            (Missing, Missing) => Must,
            (Missing, _) => Not,
        }
    }

    /// returns if this Schema Must, May, or must Not contain the passed field.
    pub fn contains_field(&self, field: &str) -> Satisfaction {
        self.satisfies(&Schema::Document(Document {
            keys: map! {
                field.to_string() => Schema::Any
            },
            required: set![field.to_string()],
            additional_properties: true,
        }))
    }

    /// upconvert_missing_to_null upconverts Missing to Null in the current level
    /// of the schema including nested Any/OneOf. It does not recurse into Documents or Arrays.
    /// This is used to properly handle array items Schemata, where Missing is not possible.
    pub fn upconvert_missing_to_null(self) -> Self {
        match self {
            Schema::Missing => Schema::Atomic(Atomic::Null),
            Schema::AnyOf(vs) => Schema::AnyOf(
                vs.into_iter()
                    .map(|e| e.upconvert_missing_to_null())
                    .collect(),
            ),
            Schema::OneOf(vs) => Schema::OneOf(
                vs.into_iter()
                    .map(|e| e.upconvert_missing_to_null())
                    .collect(),
            ),
            Schema::Any | Schema::Atomic(_) | Schema::Document(_) | Schema::Array(_) => self,
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

impl TryFrom<json_schema::Schema> for Schema {
    type Error = Error;

    /// from converts a json schema into a MongoSQL schema by following these rules:
    ///      - BsonType::Single => Schema::Atomic
    ///      - BsonType::Multiple => Schema::AnyOf
    ///      - properties, required, and additional_properties => Schema::Document
    ///      - items => Schema::Array
    ///      - any_of => Schema::AnyOf
    ///      - one_of => Schema::OneOf
    ///
    /// any_of and one_of are the only fields that are mutually exclusive with the rest.
    fn try_from(v: json_schema::Schema) -> Result<Self, Self::Error> {
        // Explicitly match the valid combinations of JSON schema fields
        match v {
            json_schema::Schema {
                bson_type: Some(bson_type),
                properties,
                required,
                additional_properties,
                items,
                any_of: None,
                one_of: None,
            } => match bson_type {
                json_schema::BsonType::Single(s) => match s.as_str() {
                    "array" => match items {
                        Some(i) => Ok(Schema::Array(Box::new(Schema::try_from(*i)?))),
                        None => Ok(Schema::Array(Box::new(Schema::Any))),
                    },
                    "object" => Ok(Schema::Document(Document::try_from(json_schema::Schema {
                        properties,
                        required,
                        additional_properties,
                        ..Default::default()
                    })?)),
                    _ => Ok(Schema::Atomic(Atomic::from_str(s.as_str())?)),
                },
                json_schema::BsonType::Multiple(m) => {
                    // For each value in `bson_type`, construct a json_schema::Schema that only
                    // contains the single type and any relevant fields and recursively call
                    // Schema::try_from on it. Wrap the resulting vector in a Schema::AnyOf
                    Ok(AnyOf(
                        m.into_iter()
                            .map(|bson_type| match bson_type.as_str() {
                                "array" => Schema::try_from(json_schema::Schema {
                                    bson_type: Some(json_schema::BsonType::Single(bson_type)),
                                    items: items.clone(),
                                    ..Default::default()
                                }),
                                "object" => Schema::try_from(json_schema::Schema {
                                    bson_type: Some(json_schema::BsonType::Single(bson_type)),
                                    properties: properties.clone(),
                                    required: required.clone(),
                                    additional_properties,
                                    ..Default::default()
                                }),
                                _ => Schema::try_from(json_schema::Schema {
                                    bson_type: Some(json_schema::BsonType::Single(bson_type)),
                                    ..Default::default()
                                }),
                            })
                            .collect::<Result<Vec<Schema>, _>>()?,
                    ))
                }
            },
            json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
                any_of: Some(any_of),
                one_of: None,
            } => Ok(Schema::AnyOf(
                any_of
                    .into_iter()
                    .map(Schema::try_from)
                    .collect::<Result<Vec<Schema>, _>>()?,
            )),
            json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
                any_of: None,
                one_of: Some(one_of),
            } => Ok(Schema::OneOf(
                one_of
                    .into_iter()
                    .map(Schema::try_from)
                    .collect::<Result<Vec<Schema>, _>>()?,
            )),
            _ => Err(Error::InvalidJsonSchema(
                "invalid combination of fields".into(),
            )),
        }
    }
}

impl Atomic {
    fn satisfies(&self, other: &Self) -> Satisfaction {
        if self == other {
            Satisfaction::Must
        } else {
            Satisfaction::Not
        }
    }
}

impl Document {
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
}
