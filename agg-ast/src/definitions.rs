use crate::custom_serde::{deserialize_mql_operator, serialize_mql_operator};
use bson::Bson;
use linked_hash_map::LinkedHashMap;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

// This module contains an aggregation pipeline syntax tree that implements
// serde::Deserialize. This allows us to deserialize aggregation pipelines from
// test YAML files into structured data and then transform that structured data
// into air structs so that we can run desugarer passes and therefore test the
// desugarers.

/// Stage represents an aggregation pipeline stage. This is not
/// a complete representation of all of MQL. Only stages relevant
/// for desugarer testing are supported here.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum Stage {
    #[serde(skip)]
    Collection(Collection),
    #[serde(rename = "$documents")]
    Documents(Vec<LinkedHashMap<String, Expression>>),
    #[serde(rename = "$project")]
    Project(ProjectStage),
    #[serde(rename = "$replaceWith")]
    ReplaceWith(Expression),
    #[serde(rename = "$match")]
    Match(MatchStage),
    #[serde(rename = "$limit")]
    Limit(i64),
    #[serde(rename = "$skip")]
    Skip(i64),
    #[serde(rename = "$sort")]
    Sort(LinkedHashMap<String, i8>),
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
    #[serde(rename = "$equiLookup")]
    EquiLookup(EquiLookup),
    #[serde(rename = "$bucket")]
    Bucket(Bucket),
    #[serde(rename = "$bucketAuto")]
    BucketAuto(BucketAuto),
    #[serde(rename = "$count")]
    Count(String),

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
    pub options: Option<String>,
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
    pub aggregations: LinkedHashMap<String, GroupAccumulator>,
}

#[derive(Clone, Debug, PartialEq)]
pub struct GroupAccumulator {
    pub function: String,
    pub expr: GroupAccumulatorExpr,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum GroupAccumulatorExpr {
    SqlAccumulator {
        distinct: bool,
        var: Box<Expression>,
    },
    NonSqlAccumulator(Expression),
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

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Lookup {
    pub from: Option<LookupFrom>,
    #[serde(rename = "let")]
    pub let_body: Option<LinkedHashMap<String, Expression>>,
    pub pipeline: Vec<Stage>,
    #[serde(rename = "as")]
    pub as_var: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct EquiLookup {
    pub from: LookupFrom,
    pub local_field: String,
    pub foreign_field: String,
    #[serde(rename = "as")]
    pub as_var: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum LookupFrom {
    Collection(String),
    Namespace(Namespace),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Namespace {
    pub db: String,
    pub coll: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Bucket {
    pub group_by: Box<Expression>,
    pub boundaries: Vec<Bson>,
    pub default: Option<Bson>,
    pub output: Option<HashMap<String, Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct BucketAuto {
    pub group_by: Box<Expression>,
    pub buckets: i32,
    pub output: Option<HashMap<String, Expression>>,
    pub granularity: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct GraphLookup {
    pub from: String,
    pub start_with: Box<Expression>,
    pub connect_from_field: String,
    pub connect_to_field: String,
    pub r#as: String,
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

/// Expression represents an aggregation pipeline expression. This is not
/// a complete representation of all of MQL. Only expressions relevant for
/// desugarer testing are supported here. Order of these variants matters
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

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum LiteralValue {
    Null,
    Boolean(bool),
    Integer(i32),
    Long(i64),
    Double(f64),
    Decimal128(bson::Decimal128),
    String(String),
}

/// UntaggedOperators are operators that follow the general format:
///   { "$<op_name>": [<args>] }
/// We need a custom deserializer that turns the key "$op_name" into
/// the field "op" in the struct.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct UntaggedOperator {
    pub op: String,
    pub args: Vec<Expression>,
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
    SqlConvert(SqlConvert),
    #[serde(rename = "$convert")]
    Convert(Convert),
    #[serde(rename = "$like")]
    Like(Like),
    #[serde(rename = "$regexMatch")]
    Regex(RegexAggExpression),
    #[serde(rename = "$sqlDivide")]
    SqlDivide(SqlDivide),
    #[serde(rename = "$trim")]
    Trim(Trim),
    #[serde(rename = "$ltrim")]
    LTrim(Trim),
    #[serde(rename = "$rtrim")]
    RTrim(Trim),

    // Subquery Operators (extended from MQL)
    #[serde(rename = "$regexFind")]
    RegexFind(RegexFind),
    #[serde(rename = "$regexFindAll")]
    RegexFindAll(RegexFindAll),
    #[serde(rename = "$replaceAll")]
    ReplaceAll(ReplaceAll),
    #[serde(rename = "$replaceOne")]
    ReplaceOne(ReplaceOne),
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
    FirstN(FirstN),
    #[serde(rename = "$lastN")]
    LastN(LastN),
    #[serde(rename = "$filter")]
    Filter(Filter),
    #[serde(rename = "$map")]
    Map(Map),
    #[serde(rename = "$maxN")]
    MaxNArrayElement(MaxNArrayElement),
    #[serde(rename = "$minN")]
    MinNArrayElement(MinNArrayElement),
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
pub struct SqlConvert {
    pub input: Box<Expression>,
    pub to: String,
    pub on_null: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Convert {
    pub input: Box<Expression>,
    pub to: String,
    pub on_null: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Filter {
    pub input: Box<Expression>,
    #[serde(rename = "as")]
    pub _as: String,
    pub cond: Box<Expression>,
    pub limit: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct FirstN {
    pub input: Box<Expression>,
    pub n: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct LastN {
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
    #[serde(rename = "as")]
    pub _as: String,
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
#[serde(rename_all = "camelCase")]
pub struct MaxNArrayElement {
    pub input: Box<Expression>,
    pub n: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct MinNArrayElement {
    pub input: Box<Expression>,
    pub n: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct RegexMatch {
    pub input: Box<Expression>,
    pub regex: Box<Expression>,
    pub options: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct RegexFind {
    pub input: Box<Expression>,
    pub regex: Box<Expression>,
    pub options: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct RegexFindAll {
    pub input: Box<Expression>,
    pub regex: Box<Expression>,
    pub options: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct ReplaceAll {
    pub input: Box<Expression>,
    pub find: Box<Expression>,
    pub replacement: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct ReplaceOne {
    pub input: Box<Expression>,
    pub find: Box<Expression>,
    pub replacement: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SqlDivide {
    pub dividend: Box<Expression>,
    pub divisor: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Trim {
    pub input: Box<Expression>,
    pub chars: Box<Expression>,
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

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DateFromParts {
    pub year: Option<Box<Expression>>,
    pub iso_week_year: Option<Box<Expression>>,
    pub month: Option<Box<Expression>>,
    pub iso_week: Option<Box<Expression>>,
    pub day: Option<Box<Expression>>,
    pub iso_day_of_week: Option<Box<Expression>>,
    pub hour: Option<Box<Expression>>,
    pub minute: Option<Box<Expression>>,
    pub second: Option<Box<Expression>>,
    pub millisecond: Option<Box<Expression>>,
    pub timezone: Option<Box<Expression>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DateFromString {
    pub date_string: Box<Expression>,
    pub format: Option<Box<Expression>>,
    pub timezone: Option<Box<Expression>>,
    pub on_error: Option<Box<Expression>>,
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

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DateToString {
    pub date: Box<Expression>,
    pub format: Option<String>,
    pub timezone: Option<Box<Expression>>,
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

/// VecOrSingleExpr represents the argument to UntaggedOperators.
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
            Stage::Group(_) => "$group",
            Stage::Join(_) => "$join",
            Stage::EquiJoin(_) => "$equiJoin",
            Stage::Unwind(_) => "$unwind",
            Stage::Lookup(_) => "$lookup",
            Stage::EquiLookup(_) => "$equiLookup",
            Stage::Bucket(_) => "$bucket",
            Stage::BucketAuto(_) => "$bucketAuto",
            Stage::Count(_) => "$count",
            Stage::GraphLookup(_) => "$graphLookup",
            Stage::AtlasSearchStage(_) => "<Atlas search stage>",
        }
    }
}
