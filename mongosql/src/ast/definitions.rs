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
    StringConstructor(String),
    Tuple(Vec<Expression>),
    TypeAssertion(TypeAssertionExpr),
}

impl Expression {
    pub fn into_date_part(self) -> Option<DatePart> {
        match self {
            Expression::Identifier(i) => i.as_str().try_into().ok(),
            _ => None
        }
    }
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
    pub arg: Box<Expression>,
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
    Millisecond,
    DayOfYear,
    DayOfWeek,
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
    Log10,
    Lower,
    LTrim,
    Mod,
    NullIf,
    OctetLength,
    Position,
    Pow,
    Radians,
    Replace,
    Round,
    RTrim,
    Sin,
    Size,
    Slice,
    Split,
    Sqrt,
    Substring,
    Tan,
    Upper,

    // Date funcs
    DateAdd,
    DateDiff,
    DateTrunc,
    Year,
    Month,
    Week,
    DayOfWeek,
    DayOfMonth,
    DayOfYear,
    Hour,
    Minute,
    Second,
    Millisecond,
}

impl TryFrom<FunctionName> for TrimSpec {
    type Error = ();

    fn try_from(name: FunctionName) -> Result<Self, Self::Error> {
        match name {
            FunctionName::LTrim => Ok(TrimSpec::Leading),
            FunctionName::RTrim => Ok(TrimSpec::Trailing),
            _ => Err(()),
        }
    }
}

impl TryFrom<FunctionName> for DatePart {
    type Error = ();

    fn try_from(name: FunctionName) -> Result<Self, Self::Error> {
        match name {
            FunctionName::Year => Ok(DatePart::Year),
            FunctionName::Month => Ok(DatePart::Month),
            FunctionName::Week => Ok(DatePart::Week),
            FunctionName::DayOfWeek => Ok(DatePart::DayOfWeek),
            FunctionName::DayOfMonth => Ok(DatePart::Day),
            FunctionName::DayOfYear => Ok(DatePart::DayOfYear),
            FunctionName::Hour => Ok(DatePart::Hour),
            FunctionName::Minute => Ok(DatePart::Minute),
            FunctionName::Second => Ok(DatePart::Second),
            FunctionName::Millisecond => Ok(DatePart::Millisecond),
            _ => Err(()),
        }
    }
}

impl TryFrom<&str> for DatePart {
    type Error = String;

    fn try_from(name: &str) -> Result<Self, Self::Error> {
        match name.to_uppercase().as_str() {
            "SQL_TSI_FRAC_SECOND" => Ok(DatePart::Millisecond),
            "SQL_TSI_SECOND" => Ok(DatePart::Second),
            "SQL_TSI_MINUTE" => Ok(DatePart::Minute),
            "SQL_TSI_HOUR" => Ok(DatePart::Hour),
            "SQL_TSI_DAY" => Ok(DatePart::Day),
            "SQL_TSI_DAYOFYEAR" => Ok(DatePart::DayOfYear),
            "SQL_TSI_WEEK" => Ok(DatePart::Week),
            "SQL_TSI_MONTH" => Ok(DatePart::Month),
            "SQL_TSI_QUARTER" => Ok(DatePart::Quarter),
            "SQL_TSI_YEAR" => Ok(DatePart::Year),
            "MILLISECOND" => Ok(DatePart::Millisecond),
            "SECOND" => Ok(DatePart::Second),
            "MINUTE" => Ok(DatePart::Minute),
            "HOUR" => Ok(DatePart::Hour),
            "DAY" => Ok(DatePart::Day),
            "DAY_OF_YEAR" => Ok(DatePart::DayOfYear),
            "DAYOFYEAR" => Ok(DatePart::DayOfYear),
            "DAY_OF_WEEK" => Ok(DatePart::DayOfWeek),
            "DAYOFWEEK" => Ok(DatePart::DayOfWeek),
            "ISO_WEEKDAY" => Ok(DatePart::IsoWeekday),
            "WEEK" => Ok(DatePart::Week),
            "ISO_WEEK" => Ok(DatePart::IsoWeek),
            "MONTH" => Ok(DatePart::Month),
            "QUARTER" => Ok(DatePart::Quarter),
            "YEAR" => Ok(DatePart::Year),
            _ => Err(format!("unknown date part {name}")),
        }
    }
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
            "CEILING" => Ok(FunctionName::Ceil),
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
            "LCASE" => Ok(FunctionName::Lower),
            "LOG" => Ok(FunctionName::Log),
            "LOG10" => Ok(FunctionName::Log10),
            "LOWER" => Ok(FunctionName::Lower),
            "LTRIM" => Ok(FunctionName::LTrim),
            "MAX" => Ok(FunctionName::Max),
            "MERGE_DOCUMENTS" => Ok(FunctionName::MergeDocuments),
            "MIN" => Ok(FunctionName::Min),
            "MOD" => Ok(FunctionName::Mod),
            "NOW" => Ok(FunctionName::CurrentTimestamp),
            "NULLIF" => Ok(FunctionName::NullIf),
            "OCTET_LENGTH" => Ok(FunctionName::OctetLength),
            "POSITION" => Ok(FunctionName::Position),
            "POW" => Ok(FunctionName::Pow),
            "POWER" => Ok(FunctionName::Pow),
            "RADIANS" => Ok(FunctionName::Radians),
            "REPLACE" => Ok(FunctionName::Replace),
            "ROUND" => Ok(FunctionName::Round),
            "RTRIM" => Ok(FunctionName::RTrim),
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
            "UCASE" => Ok(FunctionName::Upper),
            "UPPER" => Ok(FunctionName::Upper),

            "DATEADD" => Ok(FunctionName::DateAdd),
            "DATEDIFF" => Ok(FunctionName::DateDiff),
            "DATETRUNC" => Ok(FunctionName::DateTrunc),
            "TIMESTAMPADD" => Ok(FunctionName::DateAdd),
            "TIMESTAMPDIFF" => Ok(FunctionName::DateDiff),
            "TIMESTAMPTRUNC" => Ok(FunctionName::DateTrunc),
            "YEAR" => Ok(FunctionName::Year),
            "MONTH" => Ok(FunctionName::Month),
            "WEEK" => Ok(FunctionName::Week),
            "DAYOFWEEK" => Ok(FunctionName::DayOfWeek),
            "DAYOFMONTH" => Ok(FunctionName::DayOfMonth),
            "DAYOFYEAR" => Ok(FunctionName::DayOfYear),
            "HOUR" => Ok(FunctionName::Hour),
            "MINUTE" => Ok(FunctionName::Minute),
            "SECOND" => Ok(FunctionName::Second),
            "MILLISECOND" => Ok(FunctionName::Millisecond),
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
            FunctionName::Log10 => "LOG10",
            FunctionName::Lower => "LOWER",
            FunctionName::LTrim => "LTRIM",
            FunctionName::Max => "MAX",
            FunctionName::MergeDocuments => "MERGE_DOCUMENTS",
            FunctionName::Min => "MIN",
            FunctionName::Mod => "MOD",
            FunctionName::NullIf => "NULLIF",
            FunctionName::OctetLength => "OCTET_LENGTH",
            FunctionName::Position => "POSITION",
            FunctionName::Pow => "POW",
            FunctionName::Radians => "RADIANS",
            FunctionName::Replace => "REPLACE",
            FunctionName::Round => "ROUND",
            FunctionName::RTrim => "RTRIM",
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
            FunctionName::DateAdd => "DATEADD",
            FunctionName::DateDiff => "DATEDIFF",
            FunctionName::DateTrunc => "DATETRUNC",
            FunctionName::Year => "YEAR",
            FunctionName::Month => "MONTH",
            FunctionName::Week => "WEEK",
            FunctionName::DayOfWeek => "DAYOFWEEK",
            FunctionName::DayOfMonth => "DAYOFMONTH",
            FunctionName::DayOfYear => "DAYOFYEAR",
            FunctionName::Hour => "HOUR",
            FunctionName::Minute => "MINUTE",
            FunctionName::Second => "SECOND",
            FunctionName::Millisecond => "MILLISECOND",
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
            | FunctionName::Log10
            | FunctionName::Lower
            | FunctionName::LTrim
            | FunctionName::Mod
            | FunctionName::NullIf
            | FunctionName::OctetLength
            | FunctionName::Position
            | FunctionName::Pow
            | FunctionName::Radians
            | FunctionName::Replace
            | FunctionName::Round
            | FunctionName::RTrim
            | FunctionName::Sin
            | FunctionName::Size
            | FunctionName::Slice
            | FunctionName::Split
            | FunctionName::Sqrt
            | FunctionName::Substring
            | FunctionName::Tan
            | FunctionName::Upper
            | FunctionName::DateAdd
            | FunctionName::DateDiff
            | FunctionName::DateTrunc
            | FunctionName::Year
            | FunctionName::Month
            | FunctionName::Week
            | FunctionName::DayOfWeek
            | FunctionName::DayOfMonth
            | FunctionName::DayOfYear
            | FunctionName::Hour
            | FunctionName::Minute
            | FunctionName::Second
            | FunctionName::Millisecond => false,
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
    pub escape: Option<char>,
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
