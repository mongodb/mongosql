use crate::ir::binding_tuple::{BindingTuple, Key};
use linked_hash_map::LinkedHashMap;

#[allow(dead_code)]
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

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Filter {
    pub source: Box<Stage>,
    pub condition: Expression,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Project {
    pub source: Box<Stage>,
    pub expression: BindingTuple<Expression>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Group {
    pub source: Box<Stage>,
    pub keys: Vec<Aliased<Expression>>,
    pub aggregations: Vec<Aliased<AggregationFunction>>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Aliased<T> {
    pub alias: String,
    pub inner: T,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct AggregationFunction {
    pub function: String,
    pub distinct: bool,
    pub args: Vec<Expression>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Limit {
    pub source: Box<Stage>,
    pub limit: u64,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Offset {
    pub source: Box<Stage>,
    pub offset: u64,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Sort {
    pub source: Box<Stage>,
    pub specs: Vec<SortSpecification>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub enum SortSpecification {
    Asc(Box<Expression>),
    Dsc(Box<Expression>),
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Collection {
    pub db: String,
    pub collection: String,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Array {
    pub array: Vec<Expression>,
    pub alias: String,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Join {
    pub join_type: JoinType,
    pub left: Box<Stage>,
    pub right: Box<Stage>,
    pub condition: Option<Box<Expression>>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum JoinType {
    Left,
    Inner,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Set {
    pub operation: SetOperation,
    pub left: Box<Stage>,
    pub right: Box<Stage>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SetOperation {
    UnionAll,
    Union,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub enum Expression {
    Literal(Literal),
    Reference(Key),
    Array(Vec<Expression>),
    Document(LinkedHashMap<String, Expression>),
    Function(FunctionApplication),
    SearchedCase(SearchedCaseExpr),
    SimpleCase(SimpleCaseExpr),
    Cast(CastExpression),
    TypeAssertion(TypeAssertionExpression),
    FieldAccess(FieldAccess),
    SubqueryExpression(SubqueryExpression),
    SubqueryComparison(SubqueryComparison),
    Exists(Box<Stage>),
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub enum Literal {
    Null,
    Boolean(bool),
    String(String),
    Integer(i32),
    Long(i64),
    Double(f64),
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct FunctionApplication {
    pub function: Function,
    pub args: Vec<Expression>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct FieldAccess {
    pub expr: Box<Expression>,
    pub field: String,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum Function {
    // String operators
    Concat,
    Like,

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

    // Control-flow operator

    // Type operator
    Is,

    // Computed Field Access operator
    // when the field is not known until runtime.
    ComputedFieldAccess,

    // Conditional scalar functions
    Nullif,
    Coalesce,

    // Array scalar functions
    Slice,
    Size,

    // Numeric value scalar functions
    Position,
    CharLen,
    OctetLen,
    BitLen,
    Extract,

    // String value scalar functions
    Substring,
    Upper,
    Lower,
    Trim,

    // Datetime value scalar function
    CurrentTimestamp,
}

impl Function {
    /// Returns a string of the function enum.
    pub(crate) fn as_str(&self) -> &'static str {
        match self {
            Function::Add => "Add",
            Function::And => "And",
            Function::Between => "Between",
            Function::BitLen => "BitLen",
            Function::CharLen => "CharLen",
            Function::Coalesce => "Coalesce",
            Function::ComputedFieldAccess => "ComputedFieldAccess",
            Function::Concat => "Concat",
            Function::CurrentTimestamp => "CurrentTimestamp",
            Function::Div => "Div",
            Function::Eq => "Eq",
            Function::Extract => "Extract",
            Function::Gt => "Gt",
            Function::Gte => "Gte",
            Function::Is => "Is",
            Function::Like => "Like",
            Function::Lower => "Lower",
            Function::Lt => "Lt",
            Function::Lte => "Lte",
            Function::Mul => "Mul",
            Function::Neq => "Neq",
            Function::Neg => "Neg",
            Function::Not => "Not",
            Function::Nullif => "Nullif",
            Function::OctetLen => "OctetLen",
            Function::Or => "Or",
            Function::Pos => "Pos",
            Function::Position => "Position",
            Function::Size => "Size",
            Function::Slice => "Slice",
            Function::Sub => "Sub",
            Function::Substring => "Substring",
            Function::Trim => "Trim",
            Function::Upper => "Upper",
        }
    }
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct CastExpression {
    pub expr: Box<Expression>,
    pub to: Type,
    pub on_null: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct TypeAssertionExpression {
    pub expr: Box<Expression>,
    pub target_type: Type,
}

#[allow(dead_code)]
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

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryExpression {
    pub output_expr: Box<Expression>,
    pub subquery: Box<Stage>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryComparison {
    pub output_expr: Box<Expression>,
    pub operator: Function,
    pub modifier: SubqueryModifier,
    pub argument: Box<Expression>,
    pub subquery: Box<Stage>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SubqueryModifier {
    Any,
    All,
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
