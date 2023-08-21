use crate::{
    mir::{
        binding_tuple::{BindingTuple, Key},
        schema::SchemaCache,
        Error,
    },
    schema::{ResultSet, Schema},
    util::unique_linked_hash_map::UniqueLinkedHashMap,
};

visitgen::generate_visitors! {

#[derive(PartialEq, Debug, Clone)]
pub enum Stage {
    Filter(Filter),
    Project(Project),
    Group(Group),
    Limit(Limit),
    Offset(Offset),
    Sort(Sort),
    Collection(Collection),
    Array(ArraySource),
    Join(Join),
    Set(Set),
    Derived(Derived),
    Unwind(Unwind),
    MQLIntrinsic(MQLStage),
    // We need this to handle source swapping. It is not a real stage.
    // We could change source to be Option<Box<Stage>> for all nodes,
    // but that would be more difficult a refactor.
    Sentinel,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Filter {
    pub source: Box<Stage>,
    pub condition: Expression,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Project {
    pub source: Box<Stage>,
    pub expression: BindingTuple<Expression>,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Group {
    pub source: Box<Stage>,
    pub keys: Vec<OptionallyAliasedExpr>,
    pub aggregations: Vec<AliasedAggregation>,
    pub cache: SchemaCache<ResultSet>,
    pub scope: u16,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Limit {
    pub source: Box<Stage>,
    pub limit: u64,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Offset {
    pub source: Box<Stage>,
    pub offset: i64,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Sort {
    pub source: Box<Stage>,
    pub specs: Vec<SortSpecification>,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Collection {
    pub db: String,
    pub collection: String,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct ArraySource {
    pub array: Vec<Expression>,
    pub alias: String,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Join {
    pub join_type: JoinType,
    pub left: Box<Stage>,
    pub right: Box<Stage>,
    pub condition: Option<Expression>,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Set {
    pub operation: SetOperation,
    pub left: Box<Stage>,
    pub right: Box<Stage>,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Derived {
    pub source: Box<Stage>,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Unwind {
    pub source: Box<Stage>,
    pub path: Box<Expression>,
    pub index: Option<String>,
    pub outer: bool,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub enum MQLStage {
    EquiJoin(EquiJoin),
    LateralJoin(LateralJoin),
    MatchFilter(MatchFilter),
}

#[derive(PartialEq, Debug, Clone)]
pub struct EquiJoin {
    pub join_type: JoinType,
    pub source: Box<Stage>,
    pub from: Box<Stage>,
    pub local_field: Box<Expression>,
    pub foreign_field: Box<Expression>,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct LateralJoin {
    pub join_type: JoinType,
    pub source: Box<Stage>,
    pub subquery: Box<Stage>,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchFilter {
    pub source: Box<Stage>,
    pub condition: MatchQuery,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct AliasedExpr {
    pub alias: String,
    pub expr: Expression,
}

#[derive(PartialEq, Debug, Clone)]
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
}

#[derive(PartialEq, Debug, Clone)]
pub struct AliasedAggregation {
    pub alias: String,
    pub agg_expr: AggregationExpr,
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
pub enum SortSpecification {
    Asc(Box<Expression>),
    Desc(Box<Expression>),
}

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum JoinType {
    Left,
    Inner,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum SetOperation {
    UnionAll,
    Union,
}

#[derive(PartialEq, Debug, Clone)]
pub enum Expression {
    Array(ArrayExpr),
    Cast(CastExpr),
    DateFunction(DateFunctionApplication),
    Document(DocumentExpr),
    Exists(ExistsExpr),
    FieldAccess(FieldAccess),
    Is(IsExpr),
    Like(LikeExpr),
    Literal(LiteralExpr),
    Reference(ReferenceExpr),
    ScalarFunction(ScalarFunctionApplication),
    SearchedCase(SearchedCaseExpr),
    SimpleCase(SimpleCaseExpr),
    Subquery(SubqueryExpr),
    SubqueryComparison(SubqueryComparison),
    TypeAssertion(TypeAssertionExpr),

    // Special variants that only exists for optimization purposes;
    // these do not represent actual MongoSQL constructs.
    OptimizedMatchExists(OptimizedMatchExists),
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
pub struct LiteralExpr {
    pub value: LiteralValue,
    pub cache: SchemaCache<Schema>,
}

impl From<LiteralValue> for LiteralExpr {
    fn from(value: LiteralValue) -> Self {
        LiteralExpr {
            value,
            cache: SchemaCache::new(),
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct ReferenceExpr {
    pub key: Key,
    pub cache: SchemaCache<Schema>,
}

impl<T: Into<Key>> From<T> for ReferenceExpr {
    fn from(k: T) -> Self {
        ReferenceExpr {
            key: k.into(),
            cache: SchemaCache::new(),
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct ArrayExpr {
    pub array: Vec<Expression>,
    pub cache: SchemaCache<Schema>,
}

impl From<Vec<Expression>> for ArrayExpr {
    fn from(array: Vec<Expression>) -> Self {
        ArrayExpr {
            array,
            cache: SchemaCache::new(),
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct DocumentExpr {
    pub document: UniqueLinkedHashMap<String, Expression>,
    pub cache: SchemaCache<Schema>,
}

impl<T: Into<UniqueLinkedHashMap<String, Expression>>> From<T> for DocumentExpr {
    fn from(document: T) -> Self {
        DocumentExpr {
            document: document.into(),
            cache: SchemaCache::new(),
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct ExistsExpr {
    pub stage: Box<Stage>,
    pub cache: SchemaCache<Schema>,
}

impl From<Box<Stage>> for ExistsExpr {
    fn from(stage: Box<Stage>) -> Self {
        ExistsExpr {
            stage,
            cache: SchemaCache::new(),
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct IsExpr {
    pub expr: Box<Expression>,
    pub target_type: TypeOrMissing,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct LikeExpr {
    pub expr: Box<Expression>,
    pub pattern: Box<Expression>,
    pub escape: Option<String>,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct ScalarFunctionApplication {
    pub function: ScalarFunction,
    pub args: Vec<Expression>,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct FieldAccess {
    pub expr: Box<Expression>,
    pub field: String,
    pub cache: SchemaCache<Schema>,
}

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

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
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
    Abs,
    Ceil,
    Cos,
    Degrees,
    Floor,
    Log,
    Mod,
    Pow,
    Radians,
    Round,
    Sin,
    Sqrt,
    Tan,

    // String value scalar functions
    Substring,
    Upper,
    Lower,
    BTrim,
    LTrim,
    RTrim,
    Split,

    // Datetime value scalar function
    CurrentTimestamp,
    Year,
    Month,
    Day,
    Hour,
    Minute,
    Second,
    Week,
    DayOfYear,
    IsoWeek,
    IsoWeekday,

    // MergeObjects merges an array of objects
    MergeObjects,
}

impl ScalarFunction {
    /// Returns a string of the function enum.
    pub fn as_str(&self) -> &'static str {
        match self {
            ScalarFunction::Add => "Add",
            ScalarFunction::And => "And",
            ScalarFunction::BitLength => "BitLength",
            ScalarFunction::CharLength => "CharLength",
            ScalarFunction::Coalesce => "Coalesce",
            ScalarFunction::ComputedFieldAccess => "ComputedFieldAccess",
            ScalarFunction::Concat => "Concat",
            ScalarFunction::Cos => "Cos",
            ScalarFunction::CurrentTimestamp => "CurrentTimestamp",
            ScalarFunction::Year => "Year",
            ScalarFunction::Month => "Month",
            ScalarFunction::Day => "Day",
            ScalarFunction::Hour => "Hour",
            ScalarFunction::Minute => "Minute",
            ScalarFunction::Second => "Second",
            ScalarFunction::Week => "Week",
            ScalarFunction::IsoWeek => "IsoWeek",
            ScalarFunction::IsoWeekday => "IsoWeekday",
            ScalarFunction::DayOfYear => "DayOfYear",
            ScalarFunction::Abs => "Abs",
            ScalarFunction::Ceil => "Ceil",
            ScalarFunction::Degrees => "Degrees",
            ScalarFunction::Div => "Div",
            ScalarFunction::Eq => "Eq",
            ScalarFunction::Floor => "Floor",
            ScalarFunction::Gt => "Gt",
            ScalarFunction::Gte => "Gte",
            ScalarFunction::Between => "Between",
            ScalarFunction::Log => "Log",
            ScalarFunction::Lower => "Lower",
            ScalarFunction::Lt => "Lt",
            ScalarFunction::Lte => "Lte",
            ScalarFunction::Mod => "Mod",
            ScalarFunction::Mul => "Mul",
            ScalarFunction::Neq => "Neq",
            ScalarFunction::Neg => "Neg",
            ScalarFunction::Not => "Not",
            ScalarFunction::NullIf => "NullIf",
            ScalarFunction::OctetLength => "OctetLength",
            ScalarFunction::Or => "Or",
            ScalarFunction::Pos => "Pos",
            ScalarFunction::Position => "Position",
            ScalarFunction::Pow => "Pow",
            ScalarFunction::Radians => "Radians",
            ScalarFunction::Round => "Round",
            ScalarFunction::Sin => "Sin",
            ScalarFunction::Size => "Size",
            ScalarFunction::Slice => "Slice",
            ScalarFunction::Split => "Split",
            ScalarFunction::Sqrt => "Sqrt",
            ScalarFunction::Sub => "Sub",
            ScalarFunction::Substring => "Substring",
            ScalarFunction::Tan => "Tan",
            ScalarFunction::LTrim => "LTrim",
            ScalarFunction::RTrim => "RTrim",
            ScalarFunction::BTrim => "BTrim",
            ScalarFunction::Upper => "Upper",
            ScalarFunction::MergeObjects => "MergeObjects",
        }
    }
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
pub struct DateFunctionApplication {
    pub function: DateFunction,
    pub date_part: DatePart,
    pub args: Vec<Expression>,
    pub cache: SchemaCache<Schema>,
}

impl DateFunction {
    /// Returns a string of the function enum.
    pub fn as_str(&self) -> &'static str {
        match self {
            DateFunction::Add => "DateAdd",
            DateFunction::Diff => "DateDiff",
            DateFunction::Trunc => "DateTrunc",
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct SearchedCaseExpr {
    pub when_branch: Vec<WhenBranch>,
    pub else_branch: Box<Expression>,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SimpleCaseExpr {
    pub expr: Box<Expression>,
    pub when_branch: Vec<WhenBranch>,
    pub else_branch: Box<Expression>,
    pub cache: SchemaCache<Schema>,
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
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct TypeAssertionExpr {
    pub expr: Box<Expression>,
    pub target_type: Type,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum TypeOrMissing {
    Missing,
    Number,
    Type(Type),
}

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
pub struct SubqueryExpr {
    pub output_expr: Box<Expression>,
    pub subquery: Box<Stage>,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubqueryComparison {
    pub operator: SubqueryComparisonOp,
    pub modifier: SubqueryModifier,
    pub argument: Box<Expression>,
    pub subquery_expr: SubqueryExpr,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum SubqueryComparisonOp {
    Lt,
    Lte,
    Neq,
    Eq,
    Gt,
    Gte,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum SubqueryModifier {
    Any,
    All,
}

#[derive(PartialEq, Debug, Clone)]
pub struct OptimizedMatchExists {
    pub field_access: FieldAccess,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub enum MatchQuery {
    Logical(MatchLanguageLogical),
    Type(MatchLanguageType),
    Regex(MatchLanguageRegex),
    ElemMatch(ElemMatch),
    Comparison(MatchLanguageComparison),
}

#[derive(PartialEq, Debug, Clone)]
pub enum MatchPath {
    MatchReference(ReferenceExpr),
    MatchFieldAccess(MatchFieldAccess),
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchFieldAccess {
    pub parent: Box<MatchPath>,
    pub field: String,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchLanguageLogical {
    pub op: MatchLanguageLogicalOp,
    pub args: Vec<MatchQuery>,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Eq, Debug, Clone, Copy)]
pub enum MatchLanguageLogicalOp {
    Or,
    And,
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchLanguageType {
    pub input: Option<MatchPath>,
    pub target_type: TypeOrMissing,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchLanguageRegex {
    pub input: Option<MatchPath>,
    pub regex: String,
    pub options: String,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct ElemMatch {
    pub input: MatchPath,
    pub condition: Box<MatchQuery>,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchLanguageComparison {
    pub function: MatchLanguageComparisonOp,
    pub input: Option<MatchPath>,
    pub arg: LiteralValue,
    pub cache: SchemaCache<Schema>
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

impl TryFrom<crate::ast::TypeOrMissing> for TypeOrMissing {
    type Error = Error;
    fn try_from(ty: crate::ast::TypeOrMissing) -> Result<Self, Self::Error> {
        match ty {
            crate::ast::TypeOrMissing::Missing => Ok(TypeOrMissing::Missing),
            crate::ast::TypeOrMissing::Type(ty) => Ok(TypeOrMissing::Type(Type::try_from(ty)?)),
            crate::ast::TypeOrMissing::Number => Ok(TypeOrMissing::Number),
        }
    }
}

impl TryFrom<crate::ast::Type> for Type {
    type Error = Error;
    fn try_from(ty: crate::ast::Type) -> Result<Self, Self::Error> {
        use Type::*;
        match ty {
            crate::ast::Type::Array => Ok(Array),
            crate::ast::Type::BinData => Ok(BinData),
            crate::ast::Type::Boolean => Ok(Boolean),
            crate::ast::Type::Date => Err(Error::InvalidType(ty)),
            crate::ast::Type::Datetime => Ok(Datetime),
            crate::ast::Type::DbPointer => Ok(DbPointer),
            crate::ast::Type::Decimal128 => Ok(Decimal128),
            crate::ast::Type::Document => Ok(Document),
            crate::ast::Type::Double => Ok(Double),
            crate::ast::Type::Int32 => Ok(Int32),
            crate::ast::Type::Int64 => Ok(Int64),
            crate::ast::Type::Javascript => Ok(Javascript),
            crate::ast::Type::JavascriptWithScope => Ok(JavascriptWithScope),
            crate::ast::Type::MaxKey => Ok(MaxKey),
            crate::ast::Type::MinKey => Ok(MinKey),
            crate::ast::Type::Null => Ok(Null),
            crate::ast::Type::ObjectId => Ok(ObjectId),
            crate::ast::Type::RegularExpression => Ok(RegularExpression),
            crate::ast::Type::String => Ok(String),
            crate::ast::Type::Symbol => Ok(Symbol),
            crate::ast::Type::Time => Err(Error::InvalidType(ty)),
            crate::ast::Type::Timestamp => Ok(Timestamp),
            crate::ast::Type::Undefined => Ok(Undefined),
        }
    }
}

} // end of generate_visitors! block
