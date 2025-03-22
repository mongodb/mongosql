use crate::custom_serde::{deserialize_mql_operator, serialize_mql_operator};
use bson::Bson;
use linked_hash_map::LinkedHashMap;
use serde::{Deserialize, Serialize};

// This module contains an aggregation pipeline syntax tree that implements
// serde::Deserialize. This allows us to deserialize aggregation pipelines from
// test YAML files into structured data and then transform that structured data
// into air structs so that we can run desugarer passes and therefore test the
// desugarers.

// This module contains an aggregation pipeline syntax tree that implements serde::Deserialize and
// serde::Serialize. This syntax tree has two primary use cases: 1) for desugarer testing, and 2)
// for schema-derivation for the mongodb-schema-manager.
//
// The desugarer tests are specified in YAML files. The test pipelines are parsed into the types
// in this module, and then converted into air structs in the air/agg_ast module.
//
// The schema-derivation module is a sibling to this module, and is used by the mongodb-schema-
// manager.

/// Stage represents an aggregation pipeline stage.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum Stage {
    #[serde(rename = "$collection")]
    Collection(Collection),
    #[serde(rename = "$documents")]
    Documents(Vec<LinkedHashMap<String, Expression>>),
    #[serde(rename = "$project")]
    Project(ProjectStage),
    #[serde(rename = "$replaceWith", alias = "$replaceRoot")]
    ReplaceWith(ReplaceStage),
    #[serde(rename = "$match")]
    Match(MatchStage),
    #[serde(rename = "$limit")]
    Limit(i64),
    #[serde(rename = "$skip")]
    Skip(i64),
    #[serde(rename = "$sort")]
    Sort(LinkedHashMap<String, i8>),
    #[serde(rename = "$sortByCount")]
    SortByCount(Box<Expression>),
    #[serde(rename = "$group")]
    Group(Group),
    #[serde(rename = "$join")]
    Join(Box<Join>),
    #[serde(rename = "$equiJoin")]
    EquiJoin(EquiJoin),
    #[serde(rename = "$unwind")]
    Unwind(Unwind),
    #[serde(rename = "$lookup")]
    Lookup(Lookup),
    #[serde(rename = "$addFields", alias = "$set")]
    AddFields(LinkedHashMap<String, Expression>),
    #[serde(rename = "$redact")]
    Redact(Box<Expression>),
    #[serde(rename = "$unset")]
    Unset(Unset),
    #[serde(rename = "$setWindowFields")]
    SetWindowFields(SetWindowFields),
    #[serde(rename = "$bucket")]
    Bucket(Bucket),
    #[serde(rename = "$bucketAuto")]
    BucketAuto(BucketAuto),
    #[serde(rename = "$count")]
    Count(String),
    #[serde(rename = "$densify")]
    Densify(Densify),
    #[serde(rename = "$facet")]
    Facet(LinkedHashMap<String, Vec<Stage>>),
    #[serde(rename = "$fill")]
    Fill(Fill),
    #[serde(rename = "$geoNear")]
    GeoNear(GeoNear),
    #[serde(rename = "$sample")]
    Sample(Sample),
    #[serde(rename = "$unionWith")]
    UnionWith(UnionWith),

    // Search stages
    #[serde(rename = "$graphLookup")]
    GraphLookup(GraphLookup),
    #[serde(untagged)]
    AtlasSearchStage(AtlasSearchStage),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Collection {
    pub db: String,
    pub collection: String,
}

#[derive(Clone, Debug, PartialEq, Default)]
pub struct ProjectStage {
    pub items: LinkedHashMap<String, ProjectItem>,
}

impl ProjectStage {
    pub fn with_capacity(capacity: usize) -> ProjectStage {
        ProjectStage {
            items: LinkedHashMap::with_capacity(capacity),
        }
    }

    pub fn into_inner(self) -> LinkedHashMap<String, ProjectItem> {
        self.items
    }

    pub fn push(&mut self, items: (String, ProjectItem)) {
        self.items.insert(items.0, items.1);
    }

    pub fn is_empty(&self) -> bool {
        self.items.is_empty()
    }

    pub fn len(&self) -> usize {
        self.items.len()
    }
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum ProjectItem {
    Exclusion,
    Inclusion,
    Assignment(Expression),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub enum ReplaceStage {
    NewRoot(Expression),
    #[serde(untagged)]
    Expression(Expression),
}

impl ReplaceStage {
    pub fn expression(self) -> Expression {
        match self {
            ReplaceStage::NewRoot(expr) => expr,
            ReplaceStage::Expression(expr) => expr,
        }
    }
}

#[derive(Clone, Debug, PartialEq, Default)]
pub struct MatchStage {
    pub expr: Vec<MatchExpression>,
}

impl MatchStage {
    pub fn with_capacity(capacity: usize) -> MatchStage {
        MatchStage {
            expr: Vec::with_capacity(capacity),
        }
    }

    pub fn into_inner(self) -> Vec<MatchExpression> {
        self.expr
    }

    pub fn push(&mut self, expr: MatchExpression) {
        self.expr.push(expr);
    }

    pub fn is_empty(&self) -> bool {
        self.expr.is_empty()
    }

    pub fn len(&self) -> usize {
        self.expr.len()
    }
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum MatchExpression {
    Expr(MatchExpr),
    Logical(MatchLogical),
    Misc(MatchMisc),
    Field(MatchField),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum MatchMisc {
    Regex(MatchRegex),
    Element(MatchElement),
    Where(MatchWhere),
    JsonSchema(MatchJsonSchema),
    Text(MatchText),
    Comment(MatchComment),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct MatchExpr {
    #[serde(rename = "$expr")]
    pub expr: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum MatchLogical {
    #[serde(rename = "$and")]
    And(Vec<MatchExpression>),
    #[serde(rename = "$or")]
    Or(Vec<MatchExpression>),
    #[serde(rename = "$nor")]
    Nor(Vec<MatchExpression>),
    #[serde(untagged)]
    Not(MatchNot),
}

/// MatchElement represents $elemMatch expressions.
#[derive(Clone, Debug, PartialEq)]
pub struct MatchElement {
    pub field: Ref,
    pub query: MatchArrayExpression,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum MatchArrayExpression {
    Value(LinkedHashMap<MatchBinaryOp, bson::Bson>),
    Query(MatchArrayQuery),
}

#[derive(Clone, Debug, PartialEq)]
pub struct MatchArrayQuery {
    pub query: Vec<MatchExpression>,
}

impl MatchArrayQuery {
    pub fn with_capacity(capacity: usize) -> MatchArrayQuery {
        MatchArrayQuery {
            query: Vec::with_capacity(capacity),
        }
    }

    pub fn into_inner(self) -> Vec<MatchExpression> {
        self.query
    }

    pub fn push(&mut self, query: MatchExpression) {
        self.query.push(query);
    }

    pub fn is_empty(&self) -> bool {
        self.query.is_empty()
    }

    pub fn len(&self) -> usize {
        self.query.len()
    }
}

#[derive(Clone, Debug, PartialEq)]
pub struct MatchNot {
    pub field: Ref,
    pub expr: MatchNotExpression,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum MatchNotExpression {
    Query(LinkedHashMap<MatchBinaryOp, bson::Bson>),
    // technically, this needs to be a String or Regex, but this does not need
    // to be encoded in the AST, it can be enforced semantically.
    Regex(bson::Bson),
    // this is functionally unreachable because bson::Bson will capture everything deserializing; however,
    // the custom serde explicitly pulls out $elemMatch operators and deserializes to this variant.
    Element(MatchArrayExpression),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct MatchWhere {
    #[serde(rename = "$where")]
    // This is technically supposed to be String or javascript code, but this does not need to
    // be encoded in the AST, it can be enforced semantically.
    pub code: bson::Bson,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct MatchJsonSchema {
    // At some point it may make sense to fully support JsonSchema rather than just defaulting to
    // bson.
    #[serde(rename = "$jsonSchema")]
    pub schema: bson::Bson,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct MatchText {
    #[serde(rename = "$text")]
    pub expr: MatchTextContents,
}

#[derive(Clone, Debug, PartialEq)]
pub struct MatchRegex {
    pub field: Ref,
    pub pattern: bson::Bson,
    pub options: Option<bson::Bson>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct MatchComment {
    #[serde(rename = "$comment")]
    pub comment: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct MatchTextContents {
    #[serde(rename = "$search")]
    pub search: String,
    #[serde(rename = "$language")]
    pub language: Option<String>,
    #[serde(rename = "$caseSensitive")]
    pub case_sensitive: Option<bool>,
    #[serde(rename = "$diacriticSensitive")]
    pub diacritic_sensitive: Option<bool>,
}

#[derive(Clone, Debug, PartialEq)]
pub struct MatchField {
    pub field: Ref,
    pub ops: LinkedHashMap<MatchBinaryOp, bson::Bson>,
}

#[derive(Clone, Debug, PartialEq, Copy, Eq, Hash, Serialize, Deserialize)]
pub enum MatchBinaryOp {
    // Typical $match binary operators with standard format of {field: {operator: value}}
    #[serde(rename = "$eq")]
    Eq,
    #[serde(rename = "$gt")]
    Gt,
    #[serde(rename = "$gte")]
    Gte,
    #[serde(rename = "$in")]
    In,
    #[serde(rename = "$lt")]
    Lt,
    #[serde(rename = "$lte")]
    Lte,
    #[serde(rename = "$ne")]
    Ne,
    #[serde(rename = "$nin")]
    Nin,
    #[serde(rename = "$exists")]
    Exists,
    #[serde(rename = "$type")]
    Type,
    #[serde(rename = "$size")]
    Size,
    #[serde(rename = "$mod")]
    Mod,
    #[serde(rename = "$bitsAnySet")]
    BitsAnySet,
    #[serde(rename = "$bitsAnyClear")]
    BitsAnyClear,
    #[serde(rename = "$bitsAllSet")]
    BitsAllSet,
    #[serde(rename = "$bitsAllClear")]
    BitsAllClear,
    #[serde(rename = "$all")]
    All,

    // Geospatial operators have the same issue as $regex, where the fields are
    // just stuck in a bson document as the value.
    #[serde(rename = "$geoIntersects")]
    GeoIntersects,
    #[serde(rename = "$geoWithin")]
    GeoWithin,
    #[serde(rename = "$near")]
    Near,
    #[serde(rename = "$nearSphere")]
    NearSphere,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Group {
    #[serde(rename = "_id")]
    pub keys: Expression,
    #[serde(flatten)]
    pub aggregations: LinkedHashMap<String, Expression>,
}

#[derive(Clone, Copy, Debug, PartialEq, Serialize, Deserialize)]
pub enum GroupAccumulatorName {
    #[serde(rename = "$addToSet")]
    AddToSet,
    #[serde(rename = "$avg")]
    Avg,
    #[serde(rename = "$first")]
    First,
    #[serde(rename = "$last")]
    Last,
    #[serde(rename = "$max")]
    Max,
    #[serde(rename = "$mergeObjects")]
    MergeObjects,
    #[serde(rename = "$min")]
    Min,
    #[serde(rename = "$push")]
    Push,
    #[serde(rename = "$sqlAvg")]
    SQLAvg,
    #[serde(rename = "$sqlCount")]
    SQLCount,
    #[serde(rename = "$sqlFirst")]
    SQLFirst,
    #[serde(rename = "$sqlLast")]
    SQLLast,
    #[serde(rename = "$sqlMax")]
    SQLMax,
    #[serde(rename = "$sqlMergeObjects")]
    SQLMergeObjects,
    #[serde(rename = "$sqlMin")]
    SQLMin,
    #[serde(rename = "$sqlStdDevPop")]
    SQLStdDevPop,
    #[serde(rename = "$sqlStdDevSamp")]
    SQLStdDevSamp,
    #[serde(rename = "$sqlSum")]
    SQLSum,
    #[serde(rename = "$stdDevPop")]
    StdDevPop,
    #[serde(rename = "$stdDevSamp")]
    StdDevSamp,
    #[serde(rename = "$sum")]
    Sum,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Join {
    pub database: Option<String>,
    pub collection: Option<String>,
    #[serde(rename = "joinType")]
    pub join_type: JoinType,
    #[serde(rename = "let")]
    pub let_body: Option<LinkedHashMap<String, Expression>>,
    pub pipeline: Vec<Stage>,
    pub condition: Option<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct EquiJoin {
    // Note: At the moment equijoin are only supported on collections of the same DB
    pub database: Option<String>,
    pub collection: Option<String>,
    pub join_type: JoinType,
    pub local_field: String,
    pub foreign_field: String,
    #[serde(rename = "as")]
    pub as_var: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum JoinType {
    Inner,
    Left,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Unwind {
    Document(UnwindExpr),
    FieldPath(Expression),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct UnwindExpr {
    pub path: Box<Expression>,
    pub include_array_index: Option<String>,
    pub preserve_null_and_empty_arrays: Option<bool>,
}

#[derive(Clone, Debug, PartialEq, Deserialize, Serialize)]
#[serde(untagged)]
pub enum Lookup {
    ConciseSubquery(ConciseSubqueryLookup),
    Equality(EqualityLookup),
    Subquery(SubqueryLookup),
}

#[derive(Clone, Debug, PartialEq, Deserialize, Serialize)]
#[serde(untagged)]
pub enum LookupFrom {
    Collection(String),
    Namespace(Namespace),
}

#[derive(Clone, Debug, PartialEq, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct EqualityLookup {
    pub from: LookupFrom,
    pub local_field: String,
    pub foreign_field: String,
    #[serde(rename = "as")]
    pub as_var: String,
}

#[derive(Clone, Debug, PartialEq, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct ConciseSubqueryLookup {
    pub from: Option<LookupFrom>,
    pub local_field: String,
    pub foreign_field: String,
    #[serde(rename = "let")]
    pub let_body: Option<LinkedHashMap<String, Expression>>,
    pub pipeline: Vec<Stage>,
    #[serde(rename = "as")]
    pub as_var: String,
}

#[derive(Clone, Debug, PartialEq, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SubqueryLookup {
    pub from: Option<LookupFrom>,
    #[serde(rename = "let")]
    pub let_body: Option<LinkedHashMap<String, Expression>>,
    pub pipeline: Vec<Stage>,
    #[serde(rename = "as")]
    pub as_var: String,
}

#[derive(Clone, Debug, PartialEq, Deserialize, Serialize)]
pub struct Namespace {
    pub db: String,
    pub coll: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Unset {
    Single(String),
    Multiple(Vec<String>),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SetWindowFields {
    pub partition_by: Option<Box<Expression>>,
    pub sort_by: Option<LinkedHashMap<String, i8>>,
    pub output: LinkedHashMap<String, SetWindowFieldsOutput>,
}

#[derive(Clone, Debug, PartialEq, Serialize)]
pub struct SetWindowFieldsOutput {
    #[serde(flatten)]
    pub window_func: Box<Expression>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub window: Option<Window>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Window {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub documents: Option<[Bson; 2]>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub range: Option<[Bson; 2]>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub unit: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Bucket {
    pub group_by: Box<Expression>,
    pub boundaries: Vec<Bson>,
    pub default: Option<Bson>,
    pub output: Option<LinkedHashMap<String, Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct BucketAuto {
    pub group_by: Box<Expression>,
    pub buckets: i32,
    pub output: Option<LinkedHashMap<String, Expression>>,
    pub granularity: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Densify {
    pub field: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub partition_by_fields: Option<Vec<String>>,
    pub range: DensifyRange,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct DensifyRange {
    pub step: Bson,
    pub bounds: DensifyRangeBounds,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub unit: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub enum DensifyRangeBounds {
    Full,
    Partition,
    #[serde(untagged)]
    Array(Box<[Bson; 2]>),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Fill {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub partition_by: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub partition_by_fields: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub sort_by: Option<LinkedHashMap<String, i8>>,
    pub output: LinkedHashMap<String, FillOutput>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub enum FillOutput {
    Value(Expression),
    Method(FillOutputMethod),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub enum FillOutputMethod {
    Linear,
    Locf,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct GeoNear {
    pub distance_field: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub distance_multiplier: Option<Bson>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub include_locs: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub key: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_distance: Option<Bson>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub min_distance: Option<Bson>,
    pub near: GeoNearPoint,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub query: Option<MatchExpression>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub spherical: Option<bool>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum GeoNearPoint {
    GeoJSON(GeoJSON),
    Legacy([Bson; 2]),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct GeoJSON {
    pub r#type: String,
    pub coordinates: [Bson; 2],
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Sample {
    pub size: u32,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum UnionWith {
    Collection(String),
    Pipeline(UnionWithPipeline),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct UnionWithPipeline {
    pub coll: Option<String>,
    pub pipeline: Option<Vec<Stage>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct GraphLookup {
    pub from: String,
    pub start_with: Box<Expression>,
    pub connect_from_field: String,
    pub connect_to_field: String,
    #[serde(rename = "as")]
    pub as_var: String,
    pub max_depth: Option<i32>,
    pub depth_field: Option<String>,
    pub restrict_search_with_match: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum AtlasSearchStage {
    #[serde(rename = "$search")]
    Search(Box<Expression>),
    #[serde(rename = "$searchMeta")]
    SearchMeta(Box<Expression>),
    #[serde(rename = "$vectorSearch")]
    VectorSearch(Box<Expression>),
}

/// Expression represents an aggregation pipeline expression. Order of these variants matters
/// since we use custom deserialization for several expression types.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Expression {
    // Variable or field refs
    Ref(Ref),

    // Literal values including non-Ref strings
    Literal(LiteralValue),

    // Operators with structured arguments
    TaggedOperator(TaggedOperator),

    // Operators with unstructured arguments
    #[serde(
        deserialize_with = "deserialize_mql_operator",
        serialize_with = "serialize_mql_operator"
    )]
    UntaggedOperator(UntaggedOperator),

    // Array literal expressions
    Array(Vec<Expression>),

    // Document literal expressions
    Document(LinkedHashMap<String, Expression>),
}

/// Ref represents field references and variable references. Variable references are prefixed with
/// "$$" and field references are prefixed with "$".
#[derive(Clone, Debug, PartialEq)]
pub enum Ref {
    FieldRef(String),
    VariableRef(String),
}

impl Ref {
    pub fn as_str(&self) -> &str {
        match self {
            Ref::FieldRef(s) => s,
            Ref::VariableRef(s) => s,
        }
    }

    pub fn is_variable(&self) -> bool {
        matches!(self, Ref::VariableRef(_))
    }

    pub fn is_field_ref(&self) -> bool {
        matches!(self, Ref::FieldRef(_))
    }
}

// Literal values are atomic types that cannot contain sub-expressions that must be evaluated.
// Becuase of this we do not treat Arrays or Documents as literals.
// This can be thought of as an identical enum to the bson::Bson enum with Array and Document
// removed.
#[derive(Clone, Debug, PartialEq)]
pub enum LiteralValue {
    // specified in order of bson spec
    Double(f64),
    String(String),
    // Array is supported as an expression and not as a literal since it can contain expressions that must be evaluated.
    // Document is supported as expression and not as a literal since it can contain expressions that must be evaluated.
    Boolean(bool),
    Null,
    RegularExpression(bson::Regex),
    JavaScriptCode(String),
    JavaScriptCodeWithScope(bson::JavaScriptCodeWithScope),
    Int32(i32),
    Int64(i64),
    Timestamp(bson::Timestamp),
    Binary(bson::Binary),
    ObjectId(bson::oid::ObjectId),
    DateTime(bson::DateTime),
    Symbol(String),
    Decimal128(bson::Decimal128),
    Undefined,
    MaxKey,
    MinKey,
    DbPointer(bson::DbPointer),
}

/// UntaggedOperators are operators that follow the general format:
///   { "$<op_name>": [<args>] }
/// We need a custom deserializer that turns the key "$op_name" into
/// the field "op" in the struct.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct UntaggedOperator {
    pub op: UntaggedOperatorName,
    pub args: Vec<Expression>,
}

#[derive(Clone, Copy, Debug, PartialEq, Serialize, Deserialize)]
pub enum UntaggedOperatorName {
    #[serde(rename = "$abs")]
    Abs,
    #[serde(rename = "$acos")]
    Acos,
    #[serde(rename = "$acosh")]
    Acosh,
    #[serde(rename = "$asin")]
    Asin,
    #[serde(rename = "$asinh")]
    Asinh,
    #[serde(rename = "$atan")]
    Atan,
    #[serde(rename = "$atan2")]
    Atan2,
    #[serde(rename = "$atanh")]
    Atanh,
    #[serde(rename = "$add")]
    Add,
    #[serde(rename = "$addToSet")]
    AddToSet,
    #[serde(rename = "$allElementsTrue")]
    AllElementsTrue,
    #[serde(rename = "$and")]
    And,
    #[serde(rename = "$anyElementTrue")]
    AnyElementTrue,
    #[serde(rename = "$arrayElemAt")]
    ArrayElemAt,
    #[serde(rename = "$arrayToObject")]
    ArrayToObject,
    #[serde(rename = "$avg")]
    Avg,
    #[serde(rename = "$binarySize")]
    BinarySize,
    #[serde(rename = "$bitAnd")]
    BitAnd,
    #[serde(rename = "$bitNot")]
    BitNot,
    #[serde(rename = "$bitOr")]
    BitOr,
    #[serde(rename = "$bitXor")]
    BitXor,
    #[serde(rename = "$bsonSize")]
    BsonSize,
    #[serde(rename = "$ceil")]
    Ceil,
    #[serde(rename = "$cmp")]
    Cmp,
    #[serde(rename = "$coalesce")]
    Coalesce,
    #[serde(rename = "$concat")]
    Concat,
    #[serde(rename = "$concatArrays")]
    ConcatArrays,
    #[serde(rename = "$cond")]
    Cond,
    #[serde(rename = "$cos")]
    Cos,
    #[serde(rename = "$cosh")]
    Cosh,
    #[serde(rename = "$covariancePop")]
    CovariancePop,
    #[serde(rename = "$covarianceSamp")]
    CovarianceSamp,
    #[serde(rename = "$degreesToRadians")]
    DegreesToRadians,
    #[serde(rename = "$count")]
    Count,
    #[serde(rename = "$divide")]
    Divide,
    #[serde(rename = "$eq")]
    Eq,
    #[serde(rename = "$exp")]
    Exp,
    #[serde(rename = "$first")]
    First,
    #[serde(rename = "$floor")]
    Floor,
    #[serde(rename = "$gt")]
    Gt,
    #[serde(rename = "$gte")]
    Gte,
    #[serde(rename = "$ifNull")]
    IfNull,
    #[serde(rename = "$in")]
    In,
    #[serde(rename = "$indexOfArray")]
    IndexOfArray,
    #[serde(rename = "$indexOfBytes")]
    IndexOfBytes,
    #[serde(rename = "$indexOfCP")]
    IndexOfCP,
    #[serde(rename = "$is")]
    Is,
    #[serde(rename = "$isArray")]
    IsArray,
    #[serde(rename = "$isNumber")]
    IsNumber,
    #[serde(rename = "$last")]
    Last,
    #[serde(rename = "$literal")]
    Literal,
    #[serde(rename = "$locf")]
    Locf,
    #[serde(rename = "$log")]
    Log,
    #[serde(rename = "$log10")]
    Log10,
    #[serde(rename = "$ln")]
    Ln,
    #[serde(rename = "$lt")]
    Lt,
    #[serde(rename = "$lte")]
    Lte,
    #[serde(rename = "$max")]
    Max,
    #[serde(rename = "$meta")]
    Meta,
    #[serde(rename = "$min")]
    Min,
    #[serde(rename = "$mergeObjects")]
    MergeObjects,
    #[serde(rename = "$mod")]
    Mod,
    #[serde(rename = "$mqlBetween")]
    MQLBetween,
    #[serde(rename = "$multiply")]
    Multiply,
    #[serde(rename = "$ne")]
    Ne,
    #[serde(rename = "$not")]
    Not,
    #[serde(rename = "$nullIf")]
    NullIf,
    #[serde(rename = "$numberDouble")]
    NumberDouble,
    #[serde(rename = "$objectToArray")]
    ObjectToArray,
    #[serde(rename = "$or")]
    Or,
    #[serde(rename = "$pow")]
    Pow,
    #[serde(rename = "$push")]
    Push,
    #[serde(rename = "$radiansToDegrees")]
    RadiansToDegrees,
    #[serde(rename = "$rand")]
    Rand,
    #[serde(rename = "$range")]
    Range,
    #[serde(rename = "$reverseArray")]
    ReverseArray,
    #[serde(rename = "$round")]
    Round,
    #[serde(rename = "$sampleRate")]
    SampleRate,
    #[serde(rename = "$setEquals")]
    SetEquals,
    #[serde(rename = "$setIntersection")]
    SetIntersection,
    #[serde(rename = "$setDifference")]
    SetDifference,
    #[serde(rename = "$setUnion")]
    SetUnion,
    #[serde(rename = "$setIsSubset")]
    SetIsSubset,
    #[serde(rename = "$sin")]
    Sin,
    #[serde(rename = "$sinh")]
    Sinh,
    #[serde(rename = "$size")]
    Size,
    #[serde(rename = "$slice")]
    Slice,
    #[serde(rename = "$split")]
    Split,
    #[serde(rename = "$sqlAnd")]
    SQLAnd,
    #[serde(rename = "$sqlBetween")]
    SQLBetween,
    #[serde(rename = "$sqlBitLength")]
    SQLBitLength,
    #[serde(rename = "$sqlCos")]
    SQLCos,
    #[serde(rename = "$sqlEq")]
    SQLEq,
    #[serde(rename = "$sqlGt")]
    SQLGt,
    #[serde(rename = "$sqlGte")]
    SQLGte,
    #[serde(rename = "$sqlIndexOfCP")]
    SQLIndexOfCP,
    #[serde(rename = "$sqlIs")]
    SQLIs,
    #[serde(rename = "$sqlLog")]
    SQLLog,
    #[serde(rename = "$sqlLt")]
    SQLLt,
    #[serde(rename = "$sqlLte")]
    SQLLte,
    #[serde(rename = "$sqlMod")]
    SQLMod,
    #[serde(rename = "$sqlNe")]
    SQLNe,
    #[serde(rename = "$sqlNeg")]
    SQLNeg,
    #[serde(rename = "$sqlNot")]
    SQLNot,
    #[serde(rename = "$sqlOr")]
    SQLOr,
    #[serde(rename = "$sqlPos")]
    SQLPos,
    #[serde(rename = "$sqlRound")]
    SQLRound,
    #[serde(rename = "$sqlSin")]
    SQLSin,
    #[serde(rename = "$sqlSlice")]
    SQLSlice,
    #[serde(rename = "$sqlSize")]
    SQLSize,
    #[serde(rename = "$sqlSplit")]
    SQLSplit,
    #[serde(rename = "$sqlSqrt")]
    SQLSqrt,
    #[serde(rename = "$sqlStrLenCP")]
    SQLStrLenCP,
    #[serde(rename = "$sqlStrLenBytes")]
    SQLStrLenBytes,
    #[serde(rename = "$sqlSubstrCP")]
    SQLSubstrCP,
    #[serde(rename = "$sqlSum")]
    SQLSum,
    #[serde(rename = "$sqlTan")]
    SQLTan,
    #[serde(rename = "$sqlToLower")]
    SQLToLower,
    #[serde(rename = "$sqlToUpper")]
    SQLToUpper,
    #[serde(rename = "$sqrt")]
    Sqrt,
    #[serde(rename = "$stdDevPop")]
    StdDevPop,
    #[serde(rename = "$stdDevSamp")]
    StdDevSamp,
    #[serde(rename = "$strcasecmp")]
    Strcasecmp,
    #[serde(rename = "$strLenBytes")]
    StrLenBytes,
    #[serde(rename = "$strLenCP")]
    StrLenCP,
    #[serde(rename = "$substr")]
    Substr,
    #[serde(rename = "$substrBytes")]
    SubstrBytes,
    #[serde(rename = "$substrCP")]
    SubstrCP,
    #[serde(rename = "$subtract")]
    Subtract,
    #[serde(rename = "$sum")]
    Sum,
    #[serde(rename = "$tan")]
    Tan,
    #[serde(rename = "$tanh")]
    Tanh,
    #[serde(rename = "$toBool")]
    ToBool,
    #[serde(rename = "$toDate")]
    ToDate,
    #[serde(rename = "$toDecimal")]
    ToDecimal,
    #[serde(rename = "$toDouble")]
    ToDouble,
    #[serde(rename = "$toHashedIndexKey")]
    ToHashedIndexKey,
    #[serde(rename = "$toInt")]
    ToInt,
    #[serde(rename = "$toLong")]
    ToLong,
    #[serde(rename = "$toObjectId")]
    ToObjectId,
    #[serde(rename = "$toString")]
    ToString,
    #[serde(rename = "$toLower")]
    ToLower,
    #[serde(rename = "$toUpper")]
    ToUpper,
    #[serde(rename = "$trunc")]
    Trunc,
    #[serde(rename = "$tsIncrement")]
    TSIncrement,
    #[serde(rename = "$tsSecond")]
    TSSecond,
    #[serde(rename = "$type")]
    Type,
}

impl std::fmt::Display for UntaggedOperatorName {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(f, "{}", serde_json::to_string(self).unwrap())
    }
}

impl TryFrom<&str> for UntaggedOperatorName {
    type Error = String;

    fn try_from(value: &str) -> Result<Self, String> {
        serde_json::from_str(value)
            .map_err(|e| format!("Failed to deserialize operator name: {}", e))
    }
}

/// TaggedOperators are operators that have named arguments. We can utilize
/// serde directly for these by using the enum names as the keys (operator names).
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum TaggedOperator {
    #[serde(rename = "$accumulator")]
    Accumulator(Accumulator),
    #[serde(rename = "$function")]
    Function(Function),
    #[serde(rename = "$getField")]
    GetField(GetField),
    #[serde(rename = "$setField")]
    SetField(SetField),
    #[serde(rename = "$unsetField")]
    UnsetField(UnsetField),
    #[serde(rename = "$switch")]
    Switch(Switch),
    #[serde(rename = "$let")]
    Let(Let),
    #[serde(rename = "$sqlConvert")]
    SQLConvert(SQLConvert),
    #[serde(rename = "$convert")]
    Convert(Convert),
    #[serde(rename = "$like")]
    Like(Like),
    #[serde(rename = "$regexMatch")]
    Regex(RegexAggExpression),
    #[serde(rename = "$regexFind")]
    RegexFind(RegexAggExpression),
    #[serde(rename = "$regexFindAll")]
    RegexFindAll(RegexAggExpression),
    #[serde(rename = "$sqlDivide")]
    SQLDivide(SQLDivide),
    #[serde(rename = "$trim")]
    Trim(Trim),
    #[serde(rename = "$ltrim")]
    LTrim(Trim),
    #[serde(rename = "$rtrim")]
    RTrim(Trim),
    #[serde(rename = "$replaceAll")]
    ReplaceAll(Replace),
    #[serde(rename = "$replaceOne")]
    ReplaceOne(Replace),

    // Subquery Operators (extended from MQL)
    #[serde(rename = "$subquery")]
    Subquery(Subquery),
    #[serde(rename = "$subqueryComparison")]
    SubqueryComparison(SubqueryComparison),
    #[serde(rename = "$subqueryExists")]
    SubqueryExists(SubqueryExists),

    // Accumulator exprs
    #[serde(rename = "$bottom")]
    Bottom(Bottom),
    #[serde(rename = "$bottomN")]
    BottomN(BottomN),
    #[serde(rename = "$median")]
    Median(Median),
    #[serde(rename = "$percentile")]
    Percentile(Percentile),
    #[serde(rename = "$top")]
    Top(Top),
    #[serde(rename = "$topN")]
    TopN(TopN),

    // Array Operators
    #[serde(rename = "$firstN")]
    FirstN(NArrayOp),
    #[serde(rename = "$lastN")]
    LastN(NArrayOp),
    #[serde(rename = "$filter")]
    Filter(Filter),
    #[serde(rename = "$map")]
    Map(Map),
    #[serde(rename = "$maxN")]
    MaxNArrayElement(NArrayOp),
    #[serde(rename = "$minN")]
    MinNArrayElement(NArrayOp),
    #[serde(rename = "$reduce")]
    Reduce(Reduce),
    #[serde(rename = "$sortArray")]
    SortArray(SortArray),
    #[serde(rename = "$zip")]
    Zip(Zip),

    // date operators
    #[serde(rename = "$hour")]
    Hour(DateExpression),
    #[serde(rename = "$minute")]
    Minute(DateExpression),
    #[serde(rename = "$second")]
    Second(DateExpression),
    #[serde(rename = "$millisecond")]
    Millisecond(DateExpression),
    #[serde(rename = "$dayOfWeek")]
    DayOfWeek(DateExpression),
    #[serde(rename = "$dayOfMonth")]
    DayOfMonth(DateExpression),
    #[serde(rename = "$dayOfYear")]
    DayOfYear(DateExpression),
    #[serde(rename = "$isoDayOfWeek")]
    IsoDayOfWeek(DateExpression),
    #[serde(rename = "$isoWeek")]
    IsoWeek(DateExpression),
    #[serde(rename = "$isoWeekYear")]
    IsoWeekYear(DateExpression),
    #[serde(rename = "$week")]
    Week(DateExpression),
    #[serde(rename = "$month")]
    Month(DateExpression),
    #[serde(rename = "$year")]
    Year(DateExpression),
    #[serde(rename = "$dateToParts")]
    DateToParts(DateToParts),
    #[serde(rename = "$dateFromParts")]
    DateFromParts(DateFromParts),
    #[serde(rename = "$dateFromString")]
    DateFromString(DateFromString),
    #[serde(rename = "$dateToString")]
    DateToString(DateToString),
    #[serde(rename = "$dateAdd")]
    DateAdd(DateAdd),
    #[serde(rename = "$dateSubtract")]
    DateSubtract(DateSubtract),
    #[serde(rename = "$dateDiff")]
    DateDiff(DateDiff),
    #[serde(rename = "$dateTrunc")]
    DateTrunc(DateTrunc),

    // Window Functions (note: $covariance[Pop | Samp] are UntaggedOperators)
    #[serde(rename = "$denseRank")]
    DenseRank(EmptyDoc),
    #[serde(rename = "$derivative")]
    Derivative(Derivative),
    #[serde(rename = "$documentNumber")]
    DocumentNumber(EmptyDoc),
    #[serde(rename = "$expMovingAvg")]
    ExpMovingAvg(ExpMovingAvg),
    #[serde(rename = "$integral")]
    Integral(Integral),
    #[serde(rename = "$rank")]
    Rank(EmptyDoc),
    #[serde(rename = "$shift")]
    Shift(Shift),
    #[serde(rename = "$cond")]
    Cond(Cond),

    // SQL Group Accumulators
    #[serde(rename = "$sqlAvg")]
    SQLAvg(SQLAccumulator),
    #[serde(rename = "$sqlCount")]
    SQLCount(SQLAccumulator),
    #[serde(rename = "$sqlFirst")]
    SQLFirst(SQLAccumulator),
    #[serde(rename = "$sqlLast")]
    SQLLast(SQLAccumulator),
    #[serde(rename = "$sqlMax")]
    SQLMax(SQLAccumulator),
    #[serde(rename = "$sqlMergeObjects")]
    SQLMergeObjects(SQLAccumulator),
    #[serde(rename = "$sqlMin")]
    SQLMin(SQLAccumulator),
    #[serde(rename = "$sqlStdDevPop")]
    SQLStdDevPop(SQLAccumulator),
    #[serde(rename = "$sqlStdDevSamp")]
    SQLStdDevSamp(SQLAccumulator),
    #[serde(rename = "$sqlSum")]
    SQLSum(SQLAccumulator),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Accumulator {
    pub init: Box<Expression>,
    pub init_args: Option<Vec<Expression>>,
    pub accumulate: Box<Expression>,
    pub accumulate_args: Vec<Expression>,
    pub merge: Box<Expression>,
    pub finalize: Option<Box<Expression>>,
    pub lang: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Function {
    pub body: Box<Expression>,
    pub args: Vec<Expression>,
    pub lang: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct GetField {
    pub field: String,
    pub input: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct SetField {
    pub field: String,
    pub input: Box<Expression>,
    pub value: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct UnsetField {
    pub field: String,
    pub input: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Switch {
    pub branches: Vec<SwitchCase>,
    pub default: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct SwitchCase {
    pub case: Box<Expression>,
    pub then: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Let {
    pub vars: LinkedHashMap<String, Expression>,
    #[serde(rename = "in")]
    pub inside: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SQLConvert {
    pub input: Box<Expression>,
    pub to: String,
    pub on_null: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct Convert {
    pub input: Box<Expression>,
    pub to: Box<Expression>,
    pub format: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub on_null: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub on_error: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Filter {
    pub input: Box<Expression>,
    #[serde(skip_serializing_if = "Option::is_none")]
    #[serde(rename = "as")]
    pub _as: Option<String>,
    pub cond: Box<Expression>,
    pub limit: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct NArrayOp {
    pub input: Box<Expression>,
    pub n: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Like {
    pub input: Box<Expression>,
    pub pattern: Box<Expression>,
    pub escape: Option<char>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Map {
    pub input: Box<Expression>,
    #[serde(skip_serializing_if = "Option::is_none")]
    #[serde(rename = "as")]
    pub _as: Option<String>,
    #[serde(rename = "in")]
    pub inside: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct RegexAggExpression {
    pub input: Box<Expression>,
    pub regex: Box<Expression>,
    pub options: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Replace {
    pub input: Box<Expression>,
    pub find: Box<Expression>,
    pub replacement: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SQLDivide {
    pub dividend: Box<Expression>,
    pub divisor: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize)]
pub struct Trim {
    pub input: Box<Expression>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub chars: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Reduce {
    pub input: Box<Expression>,
    pub initial_value: Box<Expression>,
    #[serde(rename = "in")]
    pub inside: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SortArray {
    pub input: Box<Expression>,
    pub sort_by: SortArraySpec,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum SortArraySpec {
    Value(i8),
    Keys(LinkedHashMap<String, i8>),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Subquery {
    pub db: Option<String>,
    pub collection: Option<String>,
    #[serde(rename = "let")]
    pub let_bindings: Option<LinkedHashMap<String, Expression>>,
    pub output_path: Option<Vec<String>>,
    pub pipeline: Vec<Stage>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct SubqueryComparison {
    pub op: String,
    pub modifier: String,
    pub arg: Box<Expression>,
    pub subquery: Box<Subquery>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct SubqueryExists {
    pub db: Option<String>,
    pub collection: Option<String>,
    #[serde(rename = "let")]
    pub let_bindings: Option<LinkedHashMap<String, Expression>>,
    pub pipeline: Vec<Stage>,
}

fn default_zip_defaults() -> bool {
    false
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Zip {
    pub inputs: Box<Expression>,
    #[serde(default = "default_zip_defaults")]
    pub use_longest_length: bool,
    pub defaults: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq)]
pub struct DateExpression {
    pub date: Box<Expression>,
    pub timezone: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DateAdd {
    pub start_date: Box<Expression>,
    pub unit: Box<Expression>,
    pub amount: Box<Expression>,
    pub timezone: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DateDiff {
    pub start_date: Box<Expression>,
    pub end_date: Box<Expression>,
    pub unit: Box<Expression>,
    pub timezone: Option<Box<Expression>>,
    pub start_of_week: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct DateFromParts {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub year: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub iso_week_year: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub month: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub iso_week: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub day: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub iso_day_of_week: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub hour: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub minute: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub second: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub millisecond: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub timezone: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct DateFromString {
    pub date_string: Box<Expression>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub format: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub timezone: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub on_error: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub on_null: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DateSubtract {
    pub start_date: Box<Expression>,
    pub unit: Box<Expression>,
    pub amount: Box<Expression>,
    pub timezone: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct DateToParts {
    pub date: Box<Expression>,
    pub timezone: Option<Box<Expression>>,
    pub iso8601: Option<bool>,
}

#[derive(Clone, Debug, PartialEq, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct DateToString {
    pub date: Box<Expression>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub format: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub timezone: Option<Box<Expression>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub on_null: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DateTrunc {
    pub date: Box<Expression>,
    pub unit: Box<Expression>,
    pub bin_size: Option<Box<Expression>>,
    pub timezone: Option<Box<Expression>>,
    pub start_of_week: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Bottom {
    pub sort_by: Box<Expression>,
    pub output: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Top {
    pub sort_by: Box<Expression>,
    pub output: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct BottomN {
    pub sort_by: Box<Expression>,
    pub output: Box<Expression>,
    pub n: i64,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TopN {
    pub sort_by: Box<Expression>,
    pub output: Box<Expression>,
    pub n: i64,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Median {
    pub input: Box<Expression>,
    pub method: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Percentile {
    pub input: Box<Expression>,
    pub p: Vec<Expression>,
    pub method: String,
}

// This is useful for operators that accept an empty document as an argument.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct EmptyDoc {}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Derivative {
    pub input: Box<Expression>,
    pub unit: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct ExpMovingAvg {
    pub input: Box<Expression>,
    #[serde(flatten)]
    pub opt: ExpMovingAvgOpt,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum ExpMovingAvgOpt {
    N(i32),
    #[serde(rename = "alpha")]
    Alpha(f64),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Integral {
    pub input: Box<Expression>,
    pub unit: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Shift {
    pub output: Box<Expression>,
    pub by: i32,
    pub default: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize)]
pub struct Cond {
    pub r#if: Box<Expression>,
    pub then: Box<Expression>,
    pub r#else: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct SQLAccumulator {
    pub distinct: bool,
    pub var: Box<Expression>,
    pub arg_is_possibly_doc: Option<String>,
}

/// VecOrSingleExpr represents the argument to UntaggedOperators.
///
/// Either of the following is valid MQL:
///   { "$sqrt": "$a" }, or
///   { "$sqrt": ["$a"] }
/// So we need to be able to parse either while deserializing an
/// UntaggedOperator. This struct enables that.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum VecOrSingleExpr {
    Vec(Vec<Expression>),
    Single(Expression),
}

impl VecOrSingleExpr {
    pub fn get_as_vec(self) -> Vec<Expression> {
        match self {
            VecOrSingleExpr::Vec(v) => v,
            VecOrSingleExpr::Single(e) => vec![e],
        }
    }
}

impl Stage {
    pub fn name(&self) -> &str {
        match self {
            Stage::Collection(_) => "<collection>",
            Stage::Documents(_) => "$documents",
            Stage::Project(_) => "$project",
            Stage::ReplaceWith(_) => "$replaceWith",
            Stage::Match(_) => "$match",
            Stage::Limit(_) => "$limit",
            Stage::Skip(_) => "$skip",
            Stage::Sort(_) => "$sort",
            Stage::SortByCount(_) => "$sortByCount",
            Stage::Group(_) => "$group",
            Stage::Join(_) => "$join",
            Stage::EquiJoin(_) => "$equiJoin",
            Stage::Unwind(_) => "$unwind",
            Stage::Lookup(_) => "$lookup",
            Stage::AddFields(_) => "$addFields",
            Stage::Redact(_) => "$redact",
            Stage::Unset(_) => "$unset",
            Stage::SetWindowFields(_) => "$setWindowFields",
            Stage::Bucket(_) => "$bucket",
            Stage::BucketAuto(_) => "$bucketAuto",
            Stage::Count(_) => "$count",
            Stage::Densify(_) => "$densify",
            Stage::Facet(_) => "$facet",
            Stage::Fill(_) => "$fill",
            Stage::GeoNear(_) => "$geoNear",
            Stage::Sample(_) => "$sample",
            Stage::UnionWith(_) => "$unionWith",
            Stage::GraphLookup(_) => "$graphLookup",
            Stage::AtlasSearchStage(_) => "<Atlas search stage>",
        }
    }
}
