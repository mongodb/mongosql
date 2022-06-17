use crate::{
    ir::{
        binding_tuple::{self, BindingTuple, DatasourceName, DuplicateKeyError, Key},
        Type, TypeOrMissing,
    },
    json_schema, map,
    schema::Schema::{AnyOf, Unsat},
    set,
};
use enum_iterator::IntoEnumIterator;
use itertools::Itertools;
use lazy_static::lazy_static;
use std::{
    collections::{BTreeMap, BTreeSet},
    iter::once,
    str::FromStr,
};
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("is not a valid BSON type")]
    InvalidBSONType(),
    #[error("invalid combination of fields")]
    InvalidCombinationOfFields(),
    #[error("cannot exhaustively enumerate all field paths in schema {0:?}")]
    CannotEnumerateAllFieldPaths(Schema),
}

#[derive(PartialEq, Debug, Clone, Default)]
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

    pub fn with_merged_mappings(
        self,
        mappings: SchemaEnvironment,
    ) -> Result<Self, DuplicateKeyError> {
        self.0
            .with_merged_mappings(mappings.0)
            .map(SchemaEnvironment)
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

#[derive(PartialEq, Debug, Clone, Default)]
pub struct ResultSet {
    pub schema_env: SchemaEnvironment,
    pub min_size: u64,
    pub max_size: Option<u64>,
}

#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone)]
pub enum Schema {
    Any,
    Unsat,
    Missing,
    Atomic(Atomic),
    AnyOf(BTreeSet<Schema>),
    Array(Box<Schema>),
    Document(Document),
}

#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone, Copy, IntoEnumIterator)]
pub enum Atomic {
    String,
    Integer,
    Long,
    Double,
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
    SqlDate,
    SqlTime,
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
            _ => Err(Error::InvalidBSONType()),
        }
    }
}

#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone)]
pub struct Document {
    pub keys: BTreeMap<String, Schema>,
    pub required: BTreeSet<String>,
    pub additional_properties: bool,
}

impl Document {
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

impl TryFrom<Document> for json_schema::Schema {
    type Error = Error;

    fn try_from(v: Document) -> Result<Self, Self::Error> {
        Ok(json_schema::Schema {
            bson_type: Some(json_schema::BsonType::Single("object".to_string())),
            properties: Some(
                v.keys
                    .into_iter()
                    .map(|(k, v)| match json_schema::Schema::try_from(v) {
                        Ok(s) => Ok((k, s)),
                        Err(e) => Err(e),
                    })
                    .collect::<Result<_, _>>()?,
            ),
            required: Some(v.required.into_iter().collect()),
            additional_properties: Some(v.additional_properties),
            items: None,
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
            any_of: None,
            one_of: None,
        }
    }
}

impl From<Atomic> for String {
    fn from(v: Atomic) -> Self {
        use self::Atomic::*;
        match v {
            Decimal => "decimal",
            Double => "double",
            Integer => "int",
            Long => "long",
            String => "string",
            BinData => "binData",
            ObjectId => "objectId",
            Boolean => "bool",
            Date => "date",
            Null => "null",
            Regex => "regex",
            DbPointer => "dbPointer",
            Javascript => "javascript",
            Symbol => "symbol",
            JavascriptWithScope => "javascriptWithScope",
            Timestamp => "timestamp",
            MinKey => "minKey",
            MaxKey => "maxKey",
            SqlDate => "sqlDate",
            SqlTime => "sqlTime",
        }
        .to_string()
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
                any_of: None,
                one_of: None,
            },
            Schema::Unsat => json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
                any_of: Some(vec![]),
                one_of: None,
            },
            Schema::Missing => return Err(Error::InvalidBSONType()),
            Schema::Atomic(a) => a.into(),
            Schema::AnyOf(ao) => json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
                any_of: Some(
                    ao.into_iter()
                        .map(json_schema::Schema::try_from)
                        .collect::<Result<_, _>>()?,
                ),
                one_of: None,
            },
            Schema::Array(a) => json_schema::Schema {
                bson_type: Some(json_schema::BsonType::Single("array".to_string())),
                properties: None,
                required: None,
                additional_properties: None,
                items: Some(json_schema::Items::Single(Box::new(
                    json_schema::Schema::try_from(*a)?,
                ))),
                any_of: None,
                one_of: None,
            },
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

    // Nullish Schemas (Schemas that additionally allow for Null or Missing).
    pub static ref NULLISH: Schema =
        Schema::AnyOf(set![Schema::Atomic(Atomic::Null), Schema::Missing,]);
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

            (Array(self_arr), Array(other_arr)) => self_arr.satisfies(&*other_arr),
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

    /// get_single_field_name returns `Some(fieldName)` if every value matched
    /// by `self` which is not the empty document is a document containing a single field called `fieldName`.
    /// If this is not the case, or `self` doesn't match any values, it returns `None`.
    pub fn get_single_field_name(&self) -> Option<&str> {
        match self {
            AnyOf(any_of) => {
                let field_names = any_of
                    .iter()
                    .filter(|schema| !matches!(schema, Unsat))
                    .map(|schema| schema.get_single_field_name())
                    .collect::<BTreeSet<Option<&str>>>();
                match field_names.len() {
                    1 => field_names.into_iter().next().unwrap_or(None),
                    _ => None,
                }
            }
            Schema::Document(d) => match d.num_keys() {
                (num_required, Some(1)) if num_required <= 1 => {
                    Some(d.keys.iter().next().unwrap().0.as_str())
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
    /// an error.
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
    ) -> Result<BTreeSet<Vec<String>>, Error> {
        match self {
            Schema::Document(d) => {
                // Error if we do not have complete schema information
                if d.additional_properties {
                    return Err(Error::CannotEnumerateAllFieldPaths(self.clone()));
                }
                d.keys
                    .clone()
                    .into_iter()
                    .fold(Ok(BTreeSet::new()), |acc, (key, schema)| match max_length {
                        Some(0) => acc,
                        _ => {
                            let mut new_paths =
                                schema.enumerate_field_paths(max_length.map(|l| l - 1))?;
                            if new_paths.is_empty() {
                                new_paths = set![vec![]];
                            }
                            let mut acc = acc?;
                            acc.extend(
                                new_paths
                                    .into_iter()
                                    .map(|path| {
                                        vec![key.clone()]
                                            .into_iter()
                                            .chain(path.into_iter())
                                            .collect_vec()
                                    })
                                    .collect::<BTreeSet<Vec<String>>>(),
                            );
                            Ok(acc)
                        }
                    })
            }
            AnyOf(a) => a.iter().fold(
                Ok(BTreeSet::new()),
                |acc: Result<BTreeSet<Vec<String>>, Error>, schema| {
                    let mut new_paths = schema.enumerate_field_paths(max_length)?;
                    // Propagate an empty vector for recursive cases
                    if new_paths.is_empty() {
                        new_paths = set![vec![]]
                    }
                    Ok(acc?
                        .union(&new_paths)
                        .cloned()
                        .collect::<BTreeSet<Vec<String>>>())
                },
            ),
            Schema::Any => Err(Error::CannotEnumerateAllFieldPaths(Schema::Any)),
            Schema::Array(_) | Schema::Atomic(_) | Schema::Missing | Schema::Unsat => Ok(set![]),
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
            Date => Schema::Atomic(Atomic::SqlDate),
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
            Time => Schema::Atomic(Atomic::SqlTime),
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
                    "array" => Ok(Schema::Array(Box::new(match items {
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
                        None => Schema::Any,
                    }))),
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
                            .collect::<Result<BTreeSet<Schema>, _>>()?,
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
                    .collect::<Result<BTreeSet<Schema>, _>>()?,
            )),
            json_schema::Schema {
                bson_type: None,
                properties: None,
                required: None,
                additional_properties: None,
                items: None,
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
    fn satisfies(&self, other: &Self) -> Satisfaction {
        if self == other {
            Satisfaction::Must
        } else {
            Satisfaction::Not
        }
    }

    /// is_comparable_with returns whether or not two atomics are comparable (Must or Not only).
    /// Atomics are comparable if they are both numeric, if either is null,
    /// or otherwise both equal.
    fn is_comparable_with(&self, other: &Self) -> Satisfaction {
        use self::Atomic::*;
        use Satisfaction::*;

        match (self, other) {
            (Null, _) | (_, Null) => Must,
            (l, r) if l == r || l.is_numeric() && r.is_numeric() => Must,
            _ => Not,
        }
    }

    /// is_numeric returns whether or not the atomic value is numeric.
    pub fn is_numeric(&self) -> bool {
        use self::Atomic::*;
        match self {
            Decimal | Double | Integer | Long => true,
            String | BinData | ObjectId | Boolean | Date | Null | Regex | DbPointer
            | Javascript | Symbol | JavascriptWithScope | Timestamp | MinKey | MaxKey | SqlDate
            | SqlTime => false,
        }
    }
}

impl Document {
    /// any returns an Any Document, that is a Document that may contain any
    /// keys of Any Schema
    pub fn any() -> Document {
        Document {
            keys: map! {},
            required: set! {},
            additional_properties: true,
        }
    }

    /// empty returns an Empty Document
    pub fn empty() -> Document {
        Document {
            keys: map! {},
            required: set! {},
            additional_properties: false,
        }
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
    /// Those keys that overlap have their Schemata joined in an AnyOf.
    fn union_keys(
        mut m1: BTreeMap<String, Schema>,
        m2: BTreeMap<String, Schema>,
    ) -> BTreeMap<String, Schema> {
        for (key2, schema2) in m2.into_iter() {
            if let Some(old_schema) = m1.remove(&key2) {
                m1.insert(key2, Schema::AnyOf(set![old_schema.clone(), schema2]));
            } else {
                m1.insert(key2, schema2);
            }
        }
        m1
    }

    /// intersect_keys constructs a key map that is the intersection of the
    /// two passed maps.
    fn intersect_keys(
        m1: BTreeMap<String, Schema>,
        mut m2: BTreeMap<String, Schema>,
    ) -> BTreeMap<String, Schema> {
        let mut out = BTreeMap::new();
        for (key, s1) in m1.into_iter() {
            if let Some(s2) = m2.remove(&key) {
                out.insert(key, Schema::AnyOf(set![s1, s2]));
            }
        }
        out
    }

    /// retain_keys retains keys from the m1 map argument, creating an AnyOf for the Schema of any
    /// that overlap, and ignoring the keys from the m1 map that are not overlapping with m1.
    fn retain_keys(
        mut m1: BTreeMap<String, Schema>,
        m2: BTreeMap<String, Schema>,
    ) -> BTreeMap<String, Schema> {
        for (key, s1) in m2.into_iter() {
            if let Some(s2) = m1.remove(&key) {
                m1.insert(key, Schema::AnyOf(set![s1, s2]));
            }
        }
        m1
    }

    /// union unions together two schema::Documents returning a single Document schema that matches
    /// all document values matched by either `self` or `other`.  The Document returned is not
    /// necessarily as tight a bound as `AnyOf([Schema::Document(self), Schema::Document(other)])`;
    /// in other words, it may match additional document values not matched by `self` or `other`.
    fn union(self, other: Document) -> Document {
        let additional_properties = self.additional_properties || other.additional_properties;
        let keys = match (self.additional_properties, other.additional_properties) {
            (true, true) => Document::intersect_keys(self.keys, other.keys),
            (false, false) => Document::union_keys(self.keys, other.keys),
            (true, false) => Document::retain_keys(self.keys, other.keys),
            (false, true) => Document::retain_keys(other.keys, self.keys),
        };
        Document {
            keys,
            required: self
                .required
                .intersection(&other.required)
                .cloned()
                .collect(),
            additional_properties,
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
        Document {
            keys: Document::union_keys(self.keys, other.keys),
            required: self.required.into_iter().chain(other.required).collect(),
            additional_properties: self.additional_properties || other.additional_properties,
        }
    }

    #[inline]
    fn is_empty(&self) -> bool {
        self.keys.is_empty() && self.required.is_empty() && !self.additional_properties
    }
}
