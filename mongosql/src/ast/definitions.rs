use linked_hash_map::LinkedHashMap;
use std::convert::TryFrom;

#[derive(PartialEq, Debug, Clone)]
pub enum Query {
    Select(SelectQuery),
    Set(SetQuery),
}

#[derive(PartialEq, Debug, Clone)]
pub struct SelectQuery {
    pub select_clause: SelectClause,
    pub from_clause: Option<Datasource>,
    pub where_clause: Option<Expression>,
    pub group_by_clause: Option<GroupByClause>,
    pub having_clause: Option<Expression>,
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
    Substar(SubstarExpr),
}

#[derive(PartialEq, Debug, Clone)]
pub enum SelectExpression {
    Star,
    Substar(SubstarExpr),
    Aliased(AliasedExpr),
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubstarExpr {
    pub datasource: String,
}

#[derive(PartialEq, Debug, Clone)]
pub enum Datasource {
    Array(ArraySource),
    Collection(CollectionSource),
    Derived(DerivedSource),
    Join(JoinSource),
}

#[derive(PartialEq, Debug, Clone)]
pub struct ArraySource {
    pub array: Vec<Expression>,
    pub alias: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct CollectionSource {
    pub database: Option<String>,
    pub collection: String,
    pub alias: Option<String>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct DerivedSource {
    pub query: Box<Query>,
    pub alias: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct AliasedExpr {
    pub expr: Expression,
    pub alias: Option<String>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct JoinSource {
    pub join_type: JoinType,
    pub left: Box<Datasource>,
    pub right: Box<Datasource>,
    pub condition: Option<Expression>,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum JoinType {
    Left,
    Right,
    Cross,
    Inner,
}

#[derive(PartialEq, Debug, Clone)]
pub enum Expression {
    Binary(BinaryExpr),
    Unary(UnaryExpr),
    Between(BetweenExpr),
    Case(CaseExpr),
    Function(FunctionExpr),
    Trim(TrimExpr),
    Extract(ExtractExpr),
    Cast(CastExpr),
    Array(Vec<Expression>),
    Subquery(Box<Query>),
    Exists(Box<Query>),
    SubqueryComparison(SubqueryComparisonExpr),
    Document(LinkedHashMap<String, Expression>),
    Access(AccessExpr),
    Subpath(SubpathExpr),
    Identifier(String),
    Is(IsExpr),
    Like(LikeExpr),
    Literal(Literal),
    Tuple(Vec<Expression>),
    TypeAssertion(TypeAssertionExpr),
}

#[derive(PartialEq, Debug, Clone)]
pub struct CastExpr {
    pub expr: Box<Expression>,
    pub to: Type,
    pub on_null: Option<Box<Expression>>,
    pub on_error: Option<Box<Expression>>,
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
    pub subquery: Box<Query>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct FunctionExpr {
    pub function: FunctionName,
    pub args: FunctionArguments,
    pub set_quantifier: Option<SetQuantifier>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct ExtractExpr {
    pub extract_spec: ExtractSpec,
    pub arg: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct TrimExpr {
    pub trim_spec: TrimSpec,
    pub trim_chars: Box<Expression>,
    pub arg: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum FunctionName {
    // Aggregation functions.
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

    // Scalar functions.
    BitLength,
    CharLength,
    Coalesce,
    CurrentTimestamp,
    Lower,
    NullIf,
    OctetLength,
    Position,
    Size,
    Slice,
    Substring,
    Upper,
}

impl TryFrom<&str> for FunctionName {
    type Error = String;

    /// Takes a case-insensitive string of a function name and tries to return the
    /// corresponding enum. Returns an error string if the name is not recognized.
    ///
    /// The reciprocal `try_into` method on the `&str` type is implicitly defined.
    fn try_from(name: &str) -> Result<Self, Self::Error> {
        match name.to_uppercase().as_str() {
            // Keep in sync with `FunctionName::as_str` below.
            "ADD_TO_ARRAY" => Ok(FunctionName::AddToArray),
            "ADD_TO_SET" => Ok(FunctionName::AddToSet),
            "BIT_LENGTH" => Ok(FunctionName::BitLength),
            "AVG" => Ok(FunctionName::Avg),
            "CHAR_LENGTH" => Ok(FunctionName::CharLength),
            "CHARACTER_LENGTH" => Ok(FunctionName::CharLength),
            "COALESCE" => Ok(FunctionName::Coalesce),
            "COUNT" => Ok(FunctionName::Count),
            "CURRENT_TIMESTAMP" => Ok(FunctionName::CurrentTimestamp),
            "FIRST" => Ok(FunctionName::First),
            "LAST" => Ok(FunctionName::Last),
            "LOWER" => Ok(FunctionName::Lower),
            "MAX" => Ok(FunctionName::Max),
            "MERGE_DOCUMENTS" => Ok(FunctionName::MergeDocuments),
            "MIN" => Ok(FunctionName::Min),
            "NULLIF" => Ok(FunctionName::NullIf),
            "OCTET_LENGTH" => Ok(FunctionName::OctetLength),
            "POSITION" => Ok(FunctionName::Position),
            "SIZE" => Ok(FunctionName::Size),
            "SLICE" => Ok(FunctionName::Slice),
            "STDDEV_POP" => Ok(FunctionName::StddevPop),
            "STDDEV_SAMP" => Ok(FunctionName::StddevSamp),
            "SUBSTRING" => Ok(FunctionName::Substring),
            "SUM" => Ok(FunctionName::Sum),
            "UPPER" => Ok(FunctionName::Upper),
            _ => Err(format!("unknown function {}", name)),
        }
    }
}

impl FunctionName {
    /// Returns a capitalized string representing the function name enum.
    pub(crate) fn as_str(&self) -> &'static str {
        match self {
            // Keep in sync with `FunctionName::try_from` above.
            FunctionName::AddToArray => "ADD_TO_ARRAY",
            FunctionName::AddToSet => "ADD_TO_SET",
            FunctionName::BitLength => "BIT_LENGTH",
            FunctionName::Avg => "AVG",
            FunctionName::CharLength => "CHARACTER_LENGTH",
            FunctionName::Coalesce => "COALESCE",
            FunctionName::Count => "COUNT",
            FunctionName::CurrentTimestamp => "CURRENT_TIMESTAMP",
            FunctionName::First => "FIRST",
            FunctionName::Last => "LAST",
            FunctionName::Lower => "LOWER",
            FunctionName::Max => "MAX",
            FunctionName::MergeDocuments => "MERGE_DOCUMENTS",
            FunctionName::Min => "MIN",
            FunctionName::NullIf => "NULLIF",
            FunctionName::OctetLength => "OCTET_LENGTH",
            FunctionName::Position => "POSITION",
            FunctionName::Size => "SIZE",
            FunctionName::Slice => "SLICE",
            FunctionName::StddevPop => "STDDEV_POP",
            FunctionName::StddevSamp => "STDDEV_SAMP",
            FunctionName::Substring => "SUBSTRING",
            FunctionName::Sum => "SUM",
            FunctionName::Upper => "UPPER",
        }
    }

    /// Returns true if the `FunctionName` is any of the aggregation functions, and false otherwise.
    pub(crate) fn is_aggregation_function(&self) -> bool {
        match self {
            FunctionName::AddToArray
            | FunctionName::AddToSet
            | FunctionName::Avg
            | FunctionName::Count
            | FunctionName::First
            | FunctionName::Last
            | FunctionName::Max
            | FunctionName::MergeDocuments
            | FunctionName::Min
            | FunctionName::StddevPop
            | FunctionName::StddevSamp
            | FunctionName::Sum => true,

            FunctionName::BitLength
            | FunctionName::CharLength
            | FunctionName::Coalesce
            | FunctionName::CurrentTimestamp
            | FunctionName::Lower
            | FunctionName::NullIf
            | FunctionName::OctetLength
            | FunctionName::Position
            | FunctionName::Size
            | FunctionName::Slice
            | FunctionName::Substring
            | FunctionName::Upper => false,
        }
    }

    /// Returns true if the `FunctionName` is any of the scalar functions, and false otherwise.
    #[allow(dead_code)]
    pub(crate) fn is_scalar_function(&self) -> bool {
        !self.is_aggregation_function()
    }
}

#[derive(PartialEq, Debug, Clone)]
pub enum FunctionArguments {
    Star,
    Args(Vec<Expression>),
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum ExtractSpec {
    Year,
    Month,
    Day,
    Hour,
    Minute,
    Second,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum TrimSpec {
    Leading,
    Trailing,
    Both,
}

#[derive(PartialEq, Debug, Clone)]
pub struct AccessExpr {
    pub expr: Box<Expression>,
    pub subfield: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubpathExpr {
    pub expr: Box<Expression>,
    pub subpath: String,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum TypeOrMissing {
    Type(Type),
    Missing,
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
pub struct TypeAssertionExpr {
    pub expr: Box<Expression>,
    pub target_type: Type,
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
pub struct GroupByClause {
    pub keys: Vec<AliasedExpr>,
    pub aggregations: Vec<AliasedExpr>,
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
    Simple(Expression),
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
