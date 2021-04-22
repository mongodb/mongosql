#[derive(PartialEq, Debug, Clone)]
pub enum Query {
    Select(SelectQuery),
    Set(SetQuery),
}

#[derive(PartialEq, Debug, Clone)]
pub struct SelectQuery {
    pub select_clause: SelectClause,
    pub where_clause: Option<Box<Expression>>,
    pub order_by_clause: Option<OrderByClause>,
    pub limit: Option<u32>,
    pub offset: Option<u32>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SetQuery {
    pub left: Box<Query>,
    pub op: SetOperator,
    pub right: Box<Query>,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SetOperator {
    Union,
    UnionAll,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SelectClause {
    pub set_quantifier: SetQuantifier,
    pub body: SelectBody,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SetQuantifier {
    All,
    Distinct,
}

#[derive(PartialEq, Debug, Clone)]
pub enum SelectBody {
    Standard(Vec<SelectExpression>),
    Values(Vec<SelectValuesExpression>),
}

#[derive(PartialEq, Debug, Clone)]
pub enum SelectValuesExpression {
    Expression(Expression),
    Substar(SubstarExpression),
}

#[derive(PartialEq, Debug, Clone)]
pub enum SelectExpression {
    Star,
    Substar(SubstarExpression),
    Aliased(AliasedExpression),
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubstarExpression {
    pub datasource: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct AliasedExpression {
    pub expression: Expression,
    pub alias: Option<String>,
}

#[derive(PartialEq, Debug, Clone)]
pub enum Expression {
    Binary(BinaryExpr),
    Unary(UnaryExpr),
    Between(BetweenExpr),
    Case(CaseExpr),
    Function(FunctionExpr),
    Array(Vec<Expression>),
    Subquery(Box<SelectQuery>),
    Exists(Box<SelectQuery>),
    SubqueryComparison(SubqueryComparisonExpr),
    Identifier(Identifier),
    Literal(Literal),
    Tuple(Vec<Expression>),
}

#[derive(PartialEq, Debug, Clone)]
pub struct BinaryExpr {
    pub left: Box<Expression>,
    pub op: BinaryOp,
    pub right: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct UnaryExpr {
    pub op: UnaryOp,
    pub expr: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct BetweenExpr {
    pub expr: Box<Expression>,
    pub min: Box<Expression>,
    pub max: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct CaseExpr {
    pub expr: Option<Box<Expression>>,
    pub when_branch: Vec<WhenBranch>,
    pub else_branch: Option<Box<Expression>>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct WhenBranch {
    pub when: Box<Expression>,
    pub then: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SubqueryQuantifier {
    All,
    Any,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryComparisonExpr {
    pub expr: Box<Expression>,
    pub op: BinaryOp,
    pub quantifier: SubqueryQuantifier,
    pub subquery: Box<SelectQuery>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct FunctionExpr {
    pub function: FunctionName,
    pub args: Vec<FunctionArg>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct FunctionName(pub String);

#[derive(PartialEq, Debug, Clone)]
pub enum FunctionArg {
    Expr(Expression),
    Extract(ExtractSpec),
    Fold(Casing),
    Trim(TrimSpec),
    Cast(Type),
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum ExtractSpec {
    TimezoneHour,
    TimezoneMinute,
    Year,
    Month,
    Day,
    Hour,
    Minute,
    Second,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum Casing {
    Upper,
    Lower,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum TrimSpec {
    Leading,
    Trailing,
    Both,
}

#[derive(PartialEq, Debug, Clone)]
pub enum Identifier {
    Simple(String),
    Compound(Vec<String>),
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum UnaryOp {
    Pos,
    Neg,
    Not,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum BinaryOp {
    Add,
    And,
    Concat,
    Div,
    Eq,
    Gt,
    Gte,
    In,
    Lt,
    Lte,
    Mul,
    Neq,
    NotIn,
    Or,
    Sub,
}

#[derive(PartialEq, Debug, Clone)]
pub struct OrderByClause {
    pub sort_specs: Vec<SortSpec>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SortSpec {
    pub key: SortKey,
    pub direction: SortDirection,
}

#[derive(PartialEq, Debug, Clone)]
pub enum SortKey {
    Simple(Identifier),
    Positional(u32),
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SortDirection {
    Asc,
    Desc,
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

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum Type {
    Array,
    BinData,
    Boolean,
    Datetime,
    DbPointer,
    Decimal128(Option<u32>),
    Document,
    Double(Option<u32>),
    Int32,
    Int64,
    Javascript,
    JavascriptWithScope,
    MaxKey,
    MinKey,
    Null,
    ObjectId,
    RegularExpression,
    String(Option<u32>),
    Symbol,
    Timestamp,
    Undefined,
}
