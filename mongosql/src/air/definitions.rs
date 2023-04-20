use crate::util::unique_linked_hash_map::UniqueLinkedHashMap;

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
}

#[derive(PartialEq, Debug, Clone)]
pub struct Project {
    pub source: Box<Stage>,
    pub specifications: UniqueLinkedHashMap<String, Expression>,
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
    pub path: Box<Expression>,
    pub index: Option<String>,
    pub outer: bool,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Lookup {
    pub source: Box<Stage>,
    pub from_db: Option<String>,
    pub from_coll: Option<String>,
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
pub struct Match {
    pub source: Box<Stage>,
    pub expr: Box<Expression>,
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
    Trim(Trim),
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

    // Boolean operators
    Not,
    And,
    Or,

    // Array scalar functions
    Slice,
    Size,

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
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone)]
pub enum SQLOperator {
    // Arithmetic operators
    Divide,
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

#[derive(PartialEq, Debug, Clone)]
pub enum LiteralValue {
    Null,
    Boolean(bool),
    String(String),
    Integer(i32),
    Long(i64),
    Double(f64),
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

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone)]
pub enum SubqueryComparisonOp {
    Lt,
    Lte,
    Neq,
    Eq,
    Gt,
    Gte,
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum SubqueryModifier {
    Any,
    All,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryComparison {
    pub op: SubqueryComparisonOp,
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
pub struct Trim{
    pub op: TrimOperator,
    pub input: Box<Expression>,
    pub chars: Box<Expression>,
}


} // end of generate_visitors! block
