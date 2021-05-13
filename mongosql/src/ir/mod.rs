pub mod binding_tuple;
pub mod schema;
#[cfg(test)]
mod test;

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
    Sort(Vec<SortSpecification>),
    Collection(Collection),
    Array(Vec<Expression>),
    Join(Join),
    Set(Set),
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Filter {
    source: Box<Stage>,
    condition: Expression,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Project {
    source: Box<Stage>,
    expression: BindingTuple<Expression>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct Group {
    source: Box<Stage>,
    keys: Vec<Aliased<Expression>>,
    aggregates: Vec<Aliased<AggregationFunction>>,
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
pub struct Aliased<T> {
    alias: String,
    inner: T,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct AggregationFunction {
    function: String,
    distinct: bool,
    args: Vec<Expression>,
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
pub struct Join {
    join_type: JoinType,
    left: Box<Stage>,
    right: Box<Stage>,
    condition: Option<Box<Expression>>,
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
    operation: SetOperation,
    left: Box<Stage>,
    right: Box<Stage>,
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
    function: Function,
    args: Vec<Expression>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum Function {
    // String operators
    Concat,
    Like,

    // Arithmetic operators
    Add,
    Sub,
    Mult,
    Div,

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

    // Control-flow operator
    Case,

    // Type operator
    Is,

    // Field Access operator
    FieldAccess,

    // Conditional scalar functions
    Nullif,
    Coalesce,

    // Type conversion scalar function
    Cast,

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

    // Datetime value scalar functions
    CurrentTimestamp,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryExpression {
    output_expr: Box<Expression>,
    subquery: Box<Stage>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryComparison {
    output_expr: Box<Expression>,
    operator: Function,
    modifier: SubqueryModifier,
    argument: Box<Expression>,
    subquery: Box<Stage>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SubqueryModifier {
    Any,
    All,
}
