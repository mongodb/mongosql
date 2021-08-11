use crate::ir::binding_tuple::{BindingTuple, Key};
use linked_hash_map::LinkedHashMap;

#[derive(PartialEq, Debug, Clone)]
pub enum Stage {
    Filter(Filter),
    Project(Project),
    Group(Group),
    Limit(Limit),
    Offset(Offset),
    Sort(Sort),
    Collection(Collection),
    Array(Array),
    Join(Join),
    Set(Set),
}

#[derive(PartialEq, Debug, Clone)]
pub struct Filter {
    pub source: Box<Stage>,
    pub condition: Expression,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Project {
    pub source: Box<Stage>,
    pub expression: BindingTuple<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Group {
    pub source: Box<Stage>,
    pub keys: Vec<AliasedExpression>,
    pub aggregations: Vec<AliasedAggregation>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct AliasedExpression {
    pub alias: Option<String>,
    pub inner: Expression,
}

#[derive(PartialEq, Debug, Clone)]
pub struct AliasedAggregation {
    pub alias: String,
    pub inner: AggregationExpr,
}

#[derive(PartialEq, Debug, Clone)]
pub enum AggregationExpr {
    CountStar(bool), // true = distinct, false = not distinct
    Function(AggregationFunctionApplication),
}

#[derive(PartialEq, Debug, Clone)]
pub struct AggregationFunctionApplication {
    pub function: AggregationFunction,
    pub distinct: bool,
    pub arg: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Limit {
    pub source: Box<Stage>,
    pub limit: u64,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Offset {
    pub source: Box<Stage>,
    pub offset: u64,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Sort {
    pub source: Box<Stage>,
    pub specs: Vec<SortSpecification>,
}

#[derive(PartialEq, Debug, Clone)]
pub enum SortSpecification {
    Asc(Box<Expression>),
    Desc(Box<Expression>),
}

#[derive(PartialEq, Debug, Clone)]
pub struct Collection {
    pub db: String,
    pub collection: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Array {
    pub array: Vec<Expression>,
    pub alias: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Join {
    pub join_type: JoinType,
    pub left: Box<Stage>,
    pub right: Box<Stage>,
    pub condition: Option<Expression>,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum JoinType {
    Left,
    Inner,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Set {
    pub operation: SetOperation,
    pub left: Box<Stage>,
    pub right: Box<Stage>,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SetOperation {
    UnionAll,
    Union,
}

#[derive(PartialEq, Debug, Clone)]
pub enum Expression {
    Literal(Literal),
    Reference(Key),
    Array(Vec<Expression>),
    Document(LinkedHashMap<String, Expression>),
    ScalarFunction(ScalarFunctionApplication),
    Cast(CastExpr),
    SearchedCase(SearchedCaseExpr),
    SimpleCase(SimpleCaseExpr),
    TypeAssertion(TypeAssertionExpr),
    Is(IsExpr),
    Like(LikeExpr),
    FieldAccess(FieldAccess),
    Subquery(SubqueryExpr),
    SubqueryComparison(SubqueryComparison),
    Exists(Box<Stage>),
}

#[derive(PartialEq, Debug, Clone)]
pub enum Literal {
    Null,
    Boolean(bool),
    String(String),
    Integer(i32),
    Long(i64),
    Double(f64),
}

#[derive(PartialEq, Debug, Clone)]
pub struct IsExpr {
    pub expr: Box<Expression>,
    pub target_type: TypeOrMissing,
}

#[derive(PartialEq, Debug, Clone)]
pub struct LikeExpr {
    pub expr: Box<Expression>,
    pub pattern: Box<Expression>,
    pub escape: Option<String>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct ScalarFunctionApplication {
    pub function: ScalarFunction,
    pub args: Vec<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct FieldAccess {
    pub expr: Box<Expression>,
    pub field: String,
}

#[derive(PartialEq, Debug, Clone, Copy)]
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

impl AggregationFunction {
    /// Returns a string of the function enum.
    pub fn as_str(&self) -> &'static str {
        match self {
            AggregationFunction::AddToArray => "AddToArray",
            AggregationFunction::Avg => "Avg",
            AggregationFunction::Count => "Count",
            AggregationFunction::First => "First",
            AggregationFunction::Last => "Last",
            AggregationFunction::Max => "Max",
            AggregationFunction::MergeDocuments => "MergeDocuments",
            AggregationFunction::Min => "Min",
            AggregationFunction::StddevPop => "StddevPop",
            AggregationFunction::StddevSamp => "StddevSamp",
            AggregationFunction::Sum => "Sum",
        }
    }
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum ScalarFunction {
    // String operators
    Concat,

    // Unary arithmetic operators
    Pos,
    Neg,

    // Arithmetic operators
    Add,
    Sub,
    Mul,
    Div,

    // Comparison operators
    Lt,
    Lte,
    Neq,
    Eq,
    Gt,
    Gte,
    Between,

    // Boolean operators
    Not,
    And,
    Or,

    // Computed Field Access operator
    // when the field is not known until runtime.
    ComputedFieldAccess,

    // Conditional scalar functions
    NullIf,
    Coalesce,

    // Array scalar functions
    Slice,
    Size,

    // Numeric value scalar functions
    Position,
    CharLength,
    OctetLength,
    BitLength,

    // String value scalar functions
    Substring,
    Upper,
    Lower,
    BTrim,
    LTrim,
    RTrim,

    // Datetime value scalar function
    CurrentTimestamp,
    Year,
    Month,
    Day,
    Hour,
    Minute,
    Second,

    // MergeObjects merges an array of objects
    MergeObjects,
}

impl ScalarFunction {
    /// Returns a string of the function enum.
    pub(crate) fn as_str(&self) -> &'static str {
        match self {
            ScalarFunction::Add => "Add",
            ScalarFunction::And => "And",
            ScalarFunction::BitLength => "BitLength",
            ScalarFunction::CharLength => "CharLength",
            ScalarFunction::Coalesce => "Coalesce",
            ScalarFunction::ComputedFieldAccess => "ComputedFieldAccess",
            ScalarFunction::Concat => "Concat",
            ScalarFunction::CurrentTimestamp => "CurrentTimestamp",
            ScalarFunction::Year => "Year",
            ScalarFunction::Month => "Month",
            ScalarFunction::Day => "Day",
            ScalarFunction::Hour => "Hour",
            ScalarFunction::Minute => "Minute",
            ScalarFunction::Second => "Second",
            ScalarFunction::Div => "Div",
            ScalarFunction::Eq => "Eq",
            ScalarFunction::Gt => "Gt",
            ScalarFunction::Gte => "Gte",
            ScalarFunction::Between => "Between",
            ScalarFunction::Lower => "Lower",
            ScalarFunction::Lt => "Lt",
            ScalarFunction::Lte => "Lte",
            ScalarFunction::Mul => "Mul",
            ScalarFunction::Neq => "Neq",
            ScalarFunction::Neg => "Neg",
            ScalarFunction::Not => "Not",
            ScalarFunction::NullIf => "NullIf",
            ScalarFunction::OctetLength => "OctetLength",
            ScalarFunction::Or => "Or",
            ScalarFunction::Pos => "Pos",
            ScalarFunction::Position => "Position",
            ScalarFunction::Size => "Size",
            ScalarFunction::Slice => "Slice",
            ScalarFunction::Sub => "Sub",
            ScalarFunction::Substring => "Substring",
            ScalarFunction::LTrim => "LTrim",
            ScalarFunction::RTrim => "RTrim",
            ScalarFunction::BTrim => "BTrim",
            ScalarFunction::Upper => "Upper",
            ScalarFunction::MergeObjects => "MergeObjects",
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct SearchedCaseExpr {
    pub when_branch: Vec<WhenBranch>,
    pub else_branch: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SimpleCaseExpr {
    pub expr: Box<Expression>,
    pub when_branch: Vec<WhenBranch>,
    pub else_branch: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct WhenBranch {
    pub when: Box<Expression>,
    pub then: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct CastExpr {
    pub expr: Box<Expression>,
    pub to: Type,
    pub on_null: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct TypeAssertionExpr {
    pub expr: Box<Expression>,
    pub target_type: Type,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum TypeOrMissing {
    Missing,
    Number,
    Type(Type),
}

#[derive(PartialEq, Debug, Clone, Copy)]
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
pub struct SubqueryExpr {
    pub output_expr: Box<Expression>,
    pub subquery: Box<Stage>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryComparison {
    pub output_expr: Box<Expression>,
    pub operator: SubqueryComparisonOp,
    pub modifier: SubqueryModifier,
    pub argument: Box<Expression>,
    pub subquery: Box<Stage>,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SubqueryComparisonOp {
    Lt,
    Lte,
    Neq,
    Eq,
    Gt,
    Gte,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SubqueryModifier {
    Any,
    All,
}

impl From<crate::ast::UnaryOp> for ScalarFunction {
    fn from(op: crate::ast::UnaryOp) -> ScalarFunction {
        use crate::ast;
        match op {
            ast::UnaryOp::Pos => ScalarFunction::Pos,
            ast::UnaryOp::Neg => ScalarFunction::Neg,
            ast::UnaryOp::Not => ScalarFunction::Not,
        }
    }
}

impl From<SubqueryComparisonOp> for ScalarFunction {
    fn from(op: SubqueryComparisonOp) -> ScalarFunction {
        match op {
            SubqueryComparisonOp::Lt => ScalarFunction::Lt,
            SubqueryComparisonOp::Lte => ScalarFunction::Lte,
            SubqueryComparisonOp::Neq => ScalarFunction::Neq,
            SubqueryComparisonOp::Eq => ScalarFunction::Eq,
            SubqueryComparisonOp::Gt => ScalarFunction::Gt,
            SubqueryComparisonOp::Gte => ScalarFunction::Gte,
        }
    }
}

impl From<crate::ast::TypeOrMissing> for TypeOrMissing {
    fn from(ty: crate::ast::TypeOrMissing) -> Self {
        match ty {
            crate::ast::TypeOrMissing::Missing => TypeOrMissing::Missing,
            crate::ast::TypeOrMissing::Type(ty) => TypeOrMissing::Type(Type::from(ty)),
            crate::ast::TypeOrMissing::Number => TypeOrMissing::Number,
        }
    }
}

impl From<crate::ast::Type> for Type {
    fn from(ty: crate::ast::Type) -> Self {
        use Type::*;
        match ty {
            crate::ast::Type::Array => Array,
            crate::ast::Type::BinData => BinData,
            crate::ast::Type::Boolean => Boolean,
            crate::ast::Type::Datetime => Datetime,
            crate::ast::Type::DbPointer => DbPointer,
            crate::ast::Type::Decimal128 => Decimal128,
            crate::ast::Type::Document => Document,
            crate::ast::Type::Double => Double,
            crate::ast::Type::Int32 => Int32,
            crate::ast::Type::Int64 => Int64,
            crate::ast::Type::Javascript => Javascript,
            crate::ast::Type::JavascriptWithScope => JavascriptWithScope,
            crate::ast::Type::MaxKey => MaxKey,
            crate::ast::Type::MinKey => MinKey,
            crate::ast::Type::Null => Null,
            crate::ast::Type::ObjectId => ObjectId,
            crate::ast::Type::RegularExpression => RegularExpression,
            crate::ast::Type::String => String,
            crate::ast::Type::Symbol => Symbol,
            crate::ast::Type::Timestamp => Timestamp,
            crate::ast::Type::Undefined => Undefined,
        }
    }
}
