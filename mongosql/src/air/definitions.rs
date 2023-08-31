use crate::util::unique_linked_hash_map::UniqueLinkedHashMap;
use bson::{oid::ObjectId, DateTime, Decimal128};

visitgen::generate_visitors! {

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub enum Stage {
    Project(Project),
    Group(Group),
    Limit(Limit),
    Sort(Sort),
    Collection(Collection),
    Join(Join),
    Unwind(Unwind),
    Lookup(Lookup),
    ReplaceWith(ReplaceWith),
    Match(Match),
    UnionWith(UnionWith),
    Skip(Skip),
    Documents(Documents),
    EquiJoin(EquiJoin),
    EquiLookup(EquiLookup),
    Sentinel
}

#[derive(PartialEq, Debug, Clone)]
pub struct Project {
    pub source: Box<Stage>,
    pub specifications: UniqueLinkedHashMap<String, ProjectItem>,
}

#[derive(PartialEq, Debug, Clone)]
pub enum ProjectItem {
    Exclusion,
    Inclusion,
    Assignment(Expression)
}

#[derive(PartialEq, Debug, Clone)]
pub struct Group {
    pub source: Box<Stage>,
    pub keys: Vec<NameExprPair>,
    pub aggregations: Vec<AccumulatorExpr>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct NameExprPair {
    pub name: String,
    pub expr: Expression,
}

#[derive(PartialEq, Debug, Clone)]
pub struct AccumulatorExpr {
    pub alias: String,
    pub function: AggregationFunction,
    pub distinct: bool,
    pub arg: Box<Expression>,
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum AggregationFunction {
    AddToArray,
    AddToSet,
    Avg,
    Count,
    First,
    Last,
    Max,
    MergeDocuments,
    Min,
    StddevPop,
    StddevSamp,
    Sum,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Limit {
    pub source: Box<Stage>,
    pub limit: i64,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Sort {
    pub source: Box<Stage>,
    pub specs: Vec<SortSpecification>,
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone)]
pub enum SortSpecification {
    Asc(String),
    Desc(String),
}

#[derive(PartialEq, Eq, Debug, Clone)]
pub struct Collection {
    pub db: String,
    pub collection: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Join {
    pub join_type: JoinType,
    pub left: Box<Stage>,
    pub right: Box<Stage>,
    pub let_vars: Option<Vec<LetVariable>>,
    pub condition: Option<Expression>,
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum JoinType {
    Left,
    Inner,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Unwind {
    pub source: Box<Stage>,
    pub path: Expression,
    pub index: Option<String>,
    pub outer: bool,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Lookup {
    pub source: Box<Stage>,
    pub let_vars: Option<Vec<LetVariable>>,
    pub pipeline: Box<Stage>,
    pub as_var: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct LetVariable {
    pub name: String,
    pub expr: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct ReplaceWith {
    pub source: Box<Stage>,
    pub new_root: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub enum Match {
    ExprLanguage(ExprLanguage),
    MatchLanguage(MatchLanguage),
}

#[derive(PartialEq, Debug, Clone)]
pub struct ExprLanguage {
    pub source: Box<Stage>,
    pub expr: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchLanguage {
    pub source: Box<Stage>,
    pub expr: Box<MatchQuery>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct UnionWith {
    pub source: Box<Stage>,
    pub pipeline: Box<Stage>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Skip {
    pub source: Box<Stage>,
    pub skip: i64,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Documents {
    pub array: Vec<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct EquiJoin {
    pub join_type: JoinType,
    pub source: Box<Stage>,
    pub from: Collection,
    pub local_field: FieldRef,
    pub foreign_field: FieldRef,
}

#[derive(PartialEq, Debug, Clone)]
pub struct EquiLookup {
    pub source: Box<Stage>,
    pub from: Collection,
    pub local_field: FieldRef,
    pub foreign_field: FieldRef,
    pub as_var: String,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub enum Expression {
    MQLSemanticOperator(MQLSemanticOperator),
    SQLSemanticOperator(SQLSemanticOperator),
    Literal(LiteralValue),
    FieldRef(FieldRef),
    Variable(Variable),
    GetField(GetField),
    SetField(SetField),
    UnsetField(UnsetField),
    Switch(Switch),
    Let(Let),
    SqlConvert(SqlConvert),
    Convert(Convert),
    Like(Like),
    Is(Is),
    DateFunction(DateFunctionApplication),
    RegexMatch(RegexMatch),
    SqlDivide(SqlDivide),
    Trim(Trim),
    Reduce(Reduce),
    Subquery(Subquery),
    SubqueryComparison(SubqueryComparison),
    SubqueryExists(SubqueryExists),
    Array(Vec<Expression>),
    Document(UniqueLinkedHashMap<String, Expression>),
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone)]
pub enum MQLOperator {
    // String operators
    Concat,

    // Conditional operators
    Cond,
    IfNull,

    // Arithmetic operators
    Add,
    Subtract,
    Multiply,
    Divide,

    // Comparison operators
    Lt,
    Lte,
    Ne,
    Eq,
    Gt,
    Gte,
    Between,

    // Boolean operators
    Not,
    And,
    Or,

    // Array scalar functions
    Slice,
    Size,
    ElemAt,
    In,
    First,
    Last,

    // Numeric value scalar functions
    IndexOfCP,
    IndexOfBytes,
    StrLenCP,
    StrLenBytes,
    Abs,
    Ceil,
    Cos,
    DegreesToRadians,
    Floor,
    Log,
    Mod,
    Pow,
    RadiansToDegrees,
    Round,
    Sin,
    Tan,
    Sqrt,
    Avg,
    Max,
    Min,
    Sum,
    StddevPop,
    StddevSamp,

    // String value scalar functions
    SubstrCP,
    SubstrBytes,
    ToUpper,
    ToLower,
    Split,

    // Datetime value scalar function
    Year,
    Month,
    DayOfMonth,
    Hour,
    Minute,
    Second,
    Week,
    DayOfYear,
    IsoWeek,
    IsoDayOfWeek,
    DateAdd,
    DateDiff,
    DateTrunc,

    // MergeObjects merges an array of objects
    MergeObjects,

    // Type operators
    IsArray,
    IsNumber,
    Type,
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone)]
pub enum SQLOperator {
    // Arithmetic operators
    Pos,
    Neg,

    // Comparison operators
    Lt,
    Lte,
    Ne,
    Eq,
    Gt,
    Gte,
    Between,

    // Conditional scalar functions
    NullIf,
    Coalesce,

    // Boolean operators
    Not,
    And,
    Or,

    // Array scalar functions
    Slice,
    Size,

    // Numeric value scalar functions
    IndexOfCP,
    StrLenCP,
    StrLenBytes,
    BitLength,
    Cos,
    Log,
    Mod,
    Round,
    Sin,
    Sqrt,
    Tan,

    // String value scalar functions
    SubstrCP,
    ToUpper,
    ToLower,
    Split,

    // Extended Operators
    ComputedFieldAccess,
    CurrentTimestamp,
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum Type {
    Array,
    BinData,
    Boolean,
    Datetime,
    DbPointer,
    Decimal128,
    Document,
    Double,
    Int32,
    Int64,
    Javascript,
    JavascriptWithScope,
    MaxKey,
    MinKey,
    Null,
    ObjectId,
    RegularExpression,
    String,
    Symbol,
    Timestamp,
    Undefined,
}

impl Type {
    pub fn to_str(self) -> &'static str {
        use Type::*;
        match self {
            Array => "array",
            BinData => "binData",
            Boolean => "bool",
            Datetime => "date",
            DbPointer => "dbPointer",
            Decimal128 => "decimal",
            Document => "object",
            Double => "double",
            Int32 => "int",
            Int64 => "long",
            Javascript => "javascript",
            JavascriptWithScope => "javascriptWithScope",
            MaxKey => "maxKey",
            MinKey => "minKey",
            Null => "null",
            ObjectId => "objectId",
            RegularExpression => "regex",
            String => "string",
            Symbol => "symbol",
            Timestamp => "timestamp",
            Undefined => "undefined",
        }
    }
}


#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum SqlConvertTargetType {
    Array,
    Document,
}

impl SqlConvertTargetType {
    pub fn to_str(self) -> &'static str {
        use SqlConvertTargetType::*;
        match self {
            Array => "array",
            Document => "object",
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct MQLSemanticOperator {
    pub op: MQLOperator,
    pub args: Vec<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SQLSemanticOperator {
    pub op: SQLOperator,
    pub args: Vec<Expression>,
}

#[derive(Debug, Clone)]
pub enum LiteralValue {
    Null,
    Boolean(bool),
    String(String),
    Integer(i32),
    Long(i64),
    Double(f64),
    // Only derived from folded Converts
    ObjectId(ObjectId),
    Decimal128(Decimal128),
    DateTime(DateTime),
}

#[derive(PartialEq, Debug, Clone)]
pub struct GetField {
    pub field: String,
    pub input: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SetField {
    pub field: String,
    pub input: Box<Expression>,
    pub value: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct UnsetField {
    pub field: String,
    pub input: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct FieldRef {
    pub parent: Option<Box<FieldRef>>,
    pub name: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SwitchCase {
    pub case: Box<Expression>,
    pub then: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Switch {
    pub branches: Vec<SwitchCase>,
    pub default: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Let {
    pub vars: Vec<LetVariable>,
    pub inside: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SqlConvert {
    pub input: Box<Expression>,
    pub to: SqlConvertTargetType,
    pub on_null: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Convert {
    pub input: Box<Expression>,
    pub to: Type,
    pub on_null: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Like {
    pub expr: Box<Expression>,
    pub pattern: Box<Expression>,
    pub escape: Option<String>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Subquery {
    pub let_bindings: Vec<LetVariable>,
    pub output_path: Vec<String>,
    pub pipeline: Box<Stage>,
}

#[derive(PartialEq, Eq, Debug, Clone)]
pub enum SubqueryComparisonOp {
    Lt,
    Lte,
    Neq,
    Eq,
    Gt,
    Gte,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum SubqueryComparisonOpType {
    Mql,
    Sql,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum SubqueryModifier {
    Any,
    All,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryComparison {
    pub op: SubqueryComparisonOp,
    pub op_type: SubqueryComparisonOpType,
    pub modifier: SubqueryModifier,
    pub arg: Box<Expression>,
    pub subquery: Box<Subquery>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryExists {
    pub let_bindings: Vec<LetVariable>,
    pub pipeline: Box<Stage>,
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone)]
pub enum TypeOrMissing {
    Missing,
    Number,
    Type(Type),
}

impl TypeOrMissing {
    pub fn to_str(&self) -> &'static str {
        use TypeOrMissing::*;
        match self {
            Missing => "missing",
            Number => "number",
            Type(ty) => ty.to_str(),
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct Is {
    pub expr: Box<Expression>,
    pub target_type: TypeOrMissing,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Variable {
    pub parent: Option<Box<Variable>>,
    pub name: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct RegexMatch {
    pub input: Box<Expression>,
    pub regex: Box<Expression>,
    pub options: Option<Box<Expression>>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SqlDivide {
    pub dividend: Box<Expression>,
    pub divisor: Box<Expression>,
    pub on_error: Box<Expression>
}

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum DatePart {
    Year,
    Quarter,
    Month,
    Week,
    Day,
    Hour,
    Minute,
    Second,
}

#[derive(PartialEq, Eq, Debug, Clone)]
pub enum DateFunction {
    Add,
    Diff,
    Trunc,
}

#[derive(PartialEq, Debug, Clone)]
pub struct DateFunctionApplication{
    pub function: DateFunction,
    pub unit: DatePart,
    pub args: Vec<Expression>,
}

#[derive(PartialEq, Eq, Debug, Clone)]
pub enum TrimOperator {
    Trim,
    LTrim,
    RTrim,
}


#[derive(PartialEq, Debug, Clone)]
pub struct Trim {
    pub op: TrimOperator,
    pub input: Box<Expression>,
    pub chars: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Reduce {
    pub input: Box<Expression>,
    pub init_value: Box<Expression>,
    pub inside: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub enum MatchQuery {
    Or(Vec<MatchQuery>),
    And(Vec<MatchQuery>),
    Type(MatchLanguageType),
    Regex(MatchLanguageRegex),
    ElemMatch(ElemMatch),
    Comparison(MatchLanguageComparison),
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchLanguageType {
    pub input: Option<FieldRef>,
    pub target_type: TypeOrMissing,
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchLanguageRegex {
    pub input: Option<FieldRef>,
    pub regex: String,
    pub options: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct ElemMatch {
    pub input: FieldRef,
    pub condition: Box<MatchQuery>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchLanguageComparison {
    pub function: MatchLanguageComparisonOp,
    pub input: Option<FieldRef>,
    pub arg: LiteralValue,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum MatchLanguageComparisonOp {
    Lt,
    Lte,
    Ne,
    Eq,
    Gt,
    Gte,
}


} // end of generate_visitors! block
