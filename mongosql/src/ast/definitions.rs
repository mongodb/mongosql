use variant_count::VariantCount;

#[macro_export]
macro_rules! multimap {
	($($key:expr => $val:expr),* $(,)?) => {
		std::iter::Iterator::collect([
			$({
				$crate::ast::DocumentPair {
                                    key: $key,
                                    value: $val,
                                }
			},)*
		].into_iter())
	};
}

visitgen::generate_visitors! {

#[allow(clippy::large_enum_variant)]
#[derive(PartialEq, Debug, Clone, VariantCount)]
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

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum SetOperator {
    Union,
    UnionAll,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SelectClause {
    pub set_quantifier: SetQuantifier,
    pub body: SelectBody,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum SetQuantifier {
    All,
    Distinct,
}

#[derive(PartialEq, Debug, Clone, VariantCount)]
pub enum SelectBody {
    Standard(Vec<SelectExpression>),
    Values(Vec<SelectValuesExpression>),
}

#[derive(PartialEq, Debug, Clone, VariantCount)]
pub enum SelectValuesExpression {
    Expression(Expression),
    Substar(SubstarExpr),
}

#[derive(PartialEq, Debug, Clone, VariantCount)]
pub enum SelectExpression {
    Star,
    Substar(SubstarExpr),
    Expression(OptionallyAliasedExpr),
}

#[derive(PartialEq, Eq, Debug, Clone)]
pub struct SubstarExpr {
    pub datasource: String,
}

impl<S> From<S> for SubstarExpr
where
    S: Into<String>,
{
    fn from(s: S) -> Self {
        Self {
            datasource: s.into(),
        }
    }
}

#[derive(PartialEq, Debug, Clone, VariantCount)]
pub enum Datasource {
    Array(ArraySource),
    Collection(CollectionSource),
    Derived(DerivedSource),
    Join(JoinSource),
    Flatten(FlattenSource),
    Unwind(UnwindSource),
}

#[derive(PartialEq, Debug, Clone)]
pub struct ArraySource {
    pub array: Vec<Expression>,
    pub alias: String,
}

#[derive(PartialEq, Eq, Debug, Clone)]
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
    pub alias: String,
}

#[derive(PartialEq, Debug, Clone, VariantCount)]
pub enum OptionallyAliasedExpr {
    Aliased(AliasedExpr),
    Unaliased(Expression),
}

impl OptionallyAliasedExpr {
    pub fn get_expr(&self) -> &Expression {
        match self {
            OptionallyAliasedExpr::Aliased(AliasedExpr { alias: _, expr }) => expr,
            OptionallyAliasedExpr::Unaliased(expr) => expr,
        }
    }

    pub fn get_alias(&self) -> Option<&str> {
        match self {
            OptionallyAliasedExpr::Aliased(AliasedExpr { alias, expr: _ }) => Some(alias.as_str()),
            OptionallyAliasedExpr::Unaliased(_) => None,
        }
    }

    pub fn take_fields(self) -> (Expression, Option<String>) {
        match self {
            OptionallyAliasedExpr::Aliased(AliasedExpr { alias, expr }) => (expr, Some(alias)),
            OptionallyAliasedExpr::Unaliased(expr) => (expr, None),
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct JoinSource {
    pub join_type: JoinType,
    pub left: Box<Datasource>,
    pub right: Box<Datasource>,
    pub condition: Option<Expression>,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum JoinType {
    Left,
    Right,
    Cross,
    Inner,
}

#[derive(PartialEq, Debug, Clone)]
pub struct FlattenSource {
    pub datasource: Box<Datasource>,
    pub options: Vec<FlattenOption>,
}

#[derive(PartialEq, Eq, Debug, Clone, VariantCount)]
pub enum FlattenOption {
    Separator(String),
    Depth(u32),
}

#[derive(PartialEq, Debug, Clone)]
pub struct UnwindSource {
    pub datasource: Box<Datasource>,
    pub options: Vec<UnwindOption>,
}

#[derive(PartialEq, Debug, Clone, VariantCount)]
pub enum UnwindOption {
    Path(Expression),
    Index(String),
    Outer(bool),
}

#[derive(PartialEq, Debug, Clone, VariantCount)]
pub enum Expression {
    Binary(BinaryExpr),
    Unary(UnaryExpr),
    Between(BetweenExpr),
    Case(CaseExpr),
    Function(FunctionExpr),
    Trim(TrimExpr),
    DateFunction(DateFunctionExpr),
    Extract(ExtractExpr),
    Cast(CastExpr),
    Array(Vec<Expression>),
    Subquery(Box<Query>),
    Exists(Box<Query>),
    SubqueryComparison(SubqueryComparisonExpr),
    Document(Vec<DocumentPair>),
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
pub struct DocumentPair {
    pub key: String,
    pub value: Expression,
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

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum SubqueryQuantifier {
    All,
    Any,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryComparisonExpr {
    pub expr: Box<Expression>,
    pub op: ComparisonOp,
    pub quantifier: SubqueryQuantifier,
    pub subquery: Box<Query>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct FunctionExpr {
    pub function: FunctionName,
    pub args: FunctionArguments,
    pub set_quantifier: Option<SetQuantifier>,
}

#[derive(PartialEq, Eq, Debug, Clone, VariantCount)]
pub enum DateFunctionName {
    Add,
    Diff,
    Trunc,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum DatePart {
    Year,
    Quarter,
    Month,
    Week,
    Day,
    Hour,
    Minute,
    Second,
    DayOfYear,
    IsoWeek,
    IsoWeekday,
}

#[derive(PartialEq, Debug, Clone)]
pub struct DateFunctionExpr {
    pub function: DateFunctionName,
    pub date_part: DatePart,
    pub args: Vec<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct ExtractExpr {
    pub extract_spec: DatePart,
    pub arg: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct TrimExpr {
    pub trim_spec: TrimSpec,
    pub trim_chars: Box<Expression>,
    pub arg: Box<Expression>,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
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
    Abs,
    BitLength,
    Ceil,
    CharLength,
    Coalesce,
    Cos,
    CurrentTimestamp,
    Degrees,
    Floor,
    Log,
    Lower,
    Mod,
    NullIf,
    OctetLength,
    Position,
    Pow,
    Radians,
    Round,
    Sin,
    Size,
    Slice,
    Split,
    Sqrt,
    Substring,
    Tan,
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
            "ABS" => Ok(FunctionName::Abs),
            "ADD_TO_ARRAY" => Ok(FunctionName::AddToArray),
            "ADD_TO_SET" => Ok(FunctionName::AddToSet),
            "BIT_LENGTH" => Ok(FunctionName::BitLength),
            "AVG" => Ok(FunctionName::Avg),
            "CEIL" => Ok(FunctionName::Ceil),
            "CHAR_LENGTH" => Ok(FunctionName::CharLength),
            "CHARACTER_LENGTH" => Ok(FunctionName::CharLength),
            "COALESCE" => Ok(FunctionName::Coalesce),
            "COUNT" => Ok(FunctionName::Count),
            "COS" => Ok(FunctionName::Cos),
            "CURRENT_TIMESTAMP" => Ok(FunctionName::CurrentTimestamp),
            "DEGREES" => Ok(FunctionName::Degrees),
            "FIRST" => Ok(FunctionName::First),
            "FLOOR" => Ok(FunctionName::Floor),
            "LAST" => Ok(FunctionName::Last),
            "LOG" => Ok(FunctionName::Log),
            "LOWER" => Ok(FunctionName::Lower),
            "MAX" => Ok(FunctionName::Max),
            "MERGE_DOCUMENTS" => Ok(FunctionName::MergeDocuments),
            "MIN" => Ok(FunctionName::Min),
            "MOD" => Ok(FunctionName::Mod),
            "NULLIF" => Ok(FunctionName::NullIf),
            "OCTET_LENGTH" => Ok(FunctionName::OctetLength),
            "POSITION" => Ok(FunctionName::Position),
            "POW" => Ok(FunctionName::Pow),
            "RADIANS" => Ok(FunctionName::Radians),
            "ROUND" => Ok(FunctionName::Round),
            "SIN" => Ok(FunctionName::Sin),
            "SIZE" => Ok(FunctionName::Size),
            "SLICE" => Ok(FunctionName::Slice),
            "SPLIT" => Ok(FunctionName::Split),
            "SQRT" => Ok(FunctionName::Sqrt),
            "STDDEV_POP" => Ok(FunctionName::StddevPop),
            "STDDEV_SAMP" => Ok(FunctionName::StddevSamp),
            "SUBSTRING" => Ok(FunctionName::Substring),
            "SUM" => Ok(FunctionName::Sum),
            "TAN" => Ok(FunctionName::Tan),
            "UPPER" => Ok(FunctionName::Upper),
            _ => Err(format!("unknown function {name}")),
        }
    }
}

impl FunctionName {
    /// Returns a capitalized string representing the function name enum.
    pub fn as_str(&self) -> &'static str {
        match self {
            // Keep in sync with `FunctionName::try_from` above.
            FunctionName::Abs => "ABS",
            FunctionName::AddToArray => "ADD_TO_ARRAY",
            FunctionName::AddToSet => "ADD_TO_SET",
            FunctionName::BitLength => "BIT_LENGTH",
            FunctionName::Avg => "AVG",
            FunctionName::Ceil => "CEIL",
            FunctionName::CharLength => "CHAR_LENGTH",
            FunctionName::Coalesce => "COALESCE",
            FunctionName::Cos => "COS",
            FunctionName::Count => "COUNT",
            FunctionName::CurrentTimestamp => "CURRENT_TIMESTAMP",
            FunctionName::Degrees => "DEGREES",
            FunctionName::First => "FIRST",
            FunctionName::Floor => "FLOOR",
            FunctionName::Last => "LAST",
            FunctionName::Log => "LOG",
            FunctionName::Lower => "LOWER",
            FunctionName::Max => "MAX",
            FunctionName::MergeDocuments => "MERGE_DOCUMENTS",
            FunctionName::Min => "MIN",
            FunctionName::Mod => "MOD",
            FunctionName::NullIf => "NULLIF",
            FunctionName::OctetLength => "OCTET_LENGTH",
            FunctionName::Position => "POSITION",
            FunctionName::Pow => "POW",
            FunctionName::Radians => "RADIANS",
            FunctionName::Round => "ROUND",
            FunctionName::Size => "SIZE",
            FunctionName::Sin => "SIN",
            FunctionName::Slice => "SLICE",
            FunctionName::Split => "SPLIT",
            FunctionName::Sqrt => "SQRT",
            FunctionName::StddevPop => "STDDEV_POP",
            FunctionName::StddevSamp => "STDDEV_SAMP",
            FunctionName::Substring => "SUBSTRING",
            FunctionName::Sum => "SUM",
            FunctionName::Tan => "TAN",
            FunctionName::Upper => "UPPER",
        }
    }

    /// Returns true if the `FunctionName` is any of the aggregation functions, and false otherwise.
    pub fn is_aggregation_function(&self) -> bool {
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

            FunctionName::Abs
            | FunctionName::BitLength
            | FunctionName::Ceil
            | FunctionName::CharLength
            | FunctionName::Coalesce
            | FunctionName::Cos
            | FunctionName::CurrentTimestamp
            | FunctionName::Degrees
            | FunctionName::Floor
            | FunctionName::Log
            | FunctionName::Lower
            | FunctionName::Mod
            | FunctionName::NullIf
            | FunctionName::OctetLength
            | FunctionName::Position
            | FunctionName::Pow
            | FunctionName::Radians
            | FunctionName::Round
            | FunctionName::Sin
            | FunctionName::Size
            | FunctionName::Slice
            | FunctionName::Split
            | FunctionName::Sqrt
            | FunctionName::Substring
            | FunctionName::Tan
            | FunctionName::Upper => false,
        }
    }
}

#[derive(PartialEq, Debug, Clone, VariantCount)]
pub enum FunctionArguments {
    Star,
    Args(Vec<Expression>),
}

impl FunctionArguments {
    pub fn is_empty(&self) -> bool {
        match self {
            FunctionArguments::Star => false,
            FunctionArguments::Args(a) => a.is_empty(),
        }
    }
}

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
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

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum TypeOrMissing {
    Type(Type),
    Number,
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

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum UnaryOp {
    Pos,
    Neg,
    Not,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum BinaryOp {
    Add,
    And,
    Concat,
    Div,
    In,
    Mul,
    NotIn,
    Or,
    Sub,
    Comparison(ComparisonOp),
}

impl BinaryOp {
    pub fn as_str(&self) -> &'static str {
        use BinaryOp::*;
        match self {
            Add => "Add",
            And => "And",
            Concat => "Concat",
            Div => "Div",
            In => "In",
            Mul => "Mul",
            NotIn => "NotIn",
            Or => "Or",
            Sub => "Sub",
            Comparison(co) => co.as_str(),
        }
    }
}

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum ComparisonOp {
    Eq,
    Gt,
    Gte,
    Lt,
    Lte,
    Neq,
}

impl ComparisonOp {
    pub fn as_str(&self) -> &'static str {
        use ComparisonOp::*;
        match self {
            Eq => "Eq",
            Gt => "Gt",
            Gte => "Gte",
            Lt => "Lt",
            Lte => "Lte",
            Neq => "Neq",
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct GroupByClause {
    pub keys: Vec<OptionallyAliasedExpr>,
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

#[derive(PartialEq, Debug, Clone, VariantCount)]
pub enum SortKey {
    Simple(Expression),
    Positional(u32),
}

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum SortDirection {
    Asc,
    Desc,
}

#[derive(PartialEq, Debug, Clone, VariantCount)]
pub enum Literal {
    Null,
    Boolean(bool),
    String(String),
    Integer(i32),
    Long(i64),
    Double(f64),
}

#[derive(PartialEq, Eq, Debug, Clone, Copy, VariantCount)]
pub enum Type {
    Array,
    BinData,
    Boolean,
    Date,
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
    Time,
    Timestamp,
    Undefined,
}

} // end of generate_visitors! block
