use crate::util::unique_linked_hash_map::UniqueLinkedHashMap;

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
    ReplaceRoot(ReplaceRoot),
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
    pub limit: u64,
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
pub struct ReplaceRoot {
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
    pub skip: u64,
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
    GetField(GetField),
    FieldRef(FieldRef),
    Switch(Switch),
    Let(Let),
    SqlConvert(SqlConvert),
    Convert(Convert),
    Like(Like),
    Subquery(Subquery),
    SubqueryComparison(SubqueryComparison),
    SubqueryExists(SubqueryExists),
    Is(Is),
    Variable(String),
    Array(Vec<Expression>),
    Document(UniqueLinkedHashMap<String, Expression>),
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone)]
pub enum MQLOperator {
    // String operators
    Concat,

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
    Trim,
    LTrim,
    RTrim,
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

    // Comparison operators
    Lt,
    Lte,
    Ne,
    Eq,
    Gt,
    Gte,
    Between,
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
    pub to: Type,
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
    pub db: Option<String>,
    pub collection: Option<String>,
    pub let_bindings: bson::Document,
    pub output_path: Option<Vec<String>>,
    pub pipeline: Vec<bson::Document>,
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
    pub subquery: Box<Stage>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryExists {
    pub db: Option<String>,
    pub collection: Option<String>,
    pub let_bindings: bson::Document,
    pub pipeline: Vec<bson::Document>,
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Debug, Clone)]
pub enum TypeOrMissing {
    Missing,
    Number,
    Type(Type),
}

#[derive(PartialEq, Debug, Clone)]
pub struct Is {
    pub expr: Box<Expression>,
    pub target_type: TypeOrMissing,
}
