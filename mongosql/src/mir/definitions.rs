use crate::{
    mir::{
        binding_tuple::{BindingTuple, Key},
        schema::SchemaCache,
        Error,
    },
    schema::{ResultSet, Schema},
    util::unique_linked_hash_map::UniqueLinkedHashMap,
};

use derive_new::new;

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

impl Stage {
    pub fn is_filter(&self) -> bool {
        matches!(self, Stage::Filter(_))
    }
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
    // is_add_fields is true if the project stage is adding fields to the schema without
    // removing any. This is currently used only to allow ordering by a column not in the select
    // list.
    pub is_add_fields: bool,
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
    pub path: FieldPath,
    pub index: Option<String>,
    pub outer: bool,
    pub cache: SchemaCache<ResultSet>,

    // This field is relevant for optimization.
    // It should be set to false in all contexts
    // except PrefilterUnwindsOptimizer.
    pub is_prefiltered: bool,
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
    pub local_field: Box<FieldPath>,
    pub foreign_field: Box<FieldPath>,
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
    Asc(FieldPath),
    Desc(FieldPath),
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
    Literal(LiteralValue),
    Reference(ReferenceExpr),
    ScalarFunction(ScalarFunctionApplication),
    SearchedCase(SearchedCaseExpr),
    SimpleCase(SimpleCaseExpr),
    Subquery(SubqueryExpr),
    SubqueryComparison(SubqueryComparison),
    TypeAssertion(TypeAssertionExpr),

    // Special variants that only exists for optimization purposes;
    // these do not represent actual MongoSQL constructs.
    MQLIntrinsicFieldExistence(FieldAccess),
}

impl Expression {
    pub fn is_nullable(&self) -> bool {
        match self {
            Expression::Array(_) => false,
            Expression::Cast(x) => x.is_nullable,
            Expression::DateFunction(x) => x.is_nullable,
            Expression::Document(_) => false,
            Expression::Exists(_) => false,
            Expression::FieldAccess(x) => x.is_nullable,
            Expression::Is(_) => false,
            Expression::Like(_) => false,
            Expression::Literal(LiteralValue::Null) => true,
            Expression::Literal(_) => false,
            Expression::MQLIntrinsicFieldExistence(_) => false,
            Expression::Reference(_) => false,
            Expression::ScalarFunction(x) => x.is_nullable,
            Expression::SearchedCase(x) => x.is_nullable,
            Expression::SimpleCase(x) => x.is_nullable,
            Expression::Subquery(x) => x.is_nullable,
            Expression::SubqueryComparison(x) => x.is_nullable,
            Expression::TypeAssertion(x) => x.expr.is_nullable(),
        }
    }

    pub fn set_is_nullable(&mut self, value: bool) {
        match self {
            Expression::Cast(x) => x.is_nullable = value,
            Expression::DateFunction(x) => x.is_nullable = value,
            Expression::FieldAccess(x) => x.is_nullable = value,
            Expression::ScalarFunction(x) => x.is_nullable = x.function.is_always_nullable() || value,
            Expression::SearchedCase(x) => x.is_nullable = value,
            Expression::SimpleCase(x) => x.is_nullable = value,
            Expression::SubqueryComparison(x) => x.is_nullable = value,
            Expression::Array(_) => (),
            Expression::Document(_) => (),
            Expression::Exists(_) => (),
            Expression::Is(_) => (),
            Expression::Like(_) => (),
            Expression::Literal(_) => (),
            Expression::MQLIntrinsicFieldExistence(_) => (),
            Expression::Reference(_) => (),
            Expression::Subquery(_) => (),
            Expression::TypeAssertion(_) => (),
        }
    }
}

impl From<bson::Bson> for Expression {
    fn from(bson: bson::Bson) -> Expression {
        match bson {
            bson::Bson::Array(a) => Expression::Array(ArrayExpr {
                array: a.into_iter().map(|expr| expr.into()).collect()
            }),
            bson::Bson::Binary(b) => Expression::Literal(LiteralValue::Binary(b)),
            bson::Bson::Boolean(b) => Expression::Literal(LiteralValue::Boolean(b)),
            bson::Bson::DateTime(d) => Expression::Literal(LiteralValue::DateTime(d)),
            bson::Bson::DbPointer(d) => Expression::Literal(LiteralValue::DbPointer(d)),
            bson::Bson::Decimal128(d) => Expression::Literal(LiteralValue::Decimal128(d)),
            bson::Bson::Document(d) => Expression::Document(UniqueLinkedHashMap::from(
                d.into_iter().map(|(key, expr)| (key, expr.into())).collect::<linked_hash_map::LinkedHashMap<String, Expression>>()
            ).into()),
            bson::Bson::Double(d) => Expression::Literal(LiteralValue::Double(d)),
            bson::Bson::Int32(i) => Expression::Literal(LiteralValue::Integer(i)),
            bson::Bson::Int64(i) => Expression::Literal(LiteralValue::Long(i)),
            bson::Bson::JavaScriptCode(j) => Expression::Literal(LiteralValue::JavaScriptCode(j)),
            bson::Bson::JavaScriptCodeWithScope(j) => Expression::Literal(LiteralValue::JavaScriptCodeWithScope(j)),
            bson::Bson::MinKey => Expression::Literal(LiteralValue::MinKey),
            bson::Bson::MaxKey => Expression::Literal(LiteralValue::MaxKey),
            bson::Bson::Null => Expression::Literal(LiteralValue::Null),
            bson::Bson::ObjectId(o) => Expression::Literal(LiteralValue::ObjectId(o)),
            bson::Bson::RegularExpression(r) => Expression::Literal(LiteralValue::RegularExpression(r)),
            bson::Bson::String(s) => Expression::Literal(LiteralValue::String(s)),
            bson::Bson::Symbol(s) => Expression::Literal(LiteralValue::Symbol(s)),
            bson::Bson::Timestamp(t) => Expression::Literal(LiteralValue::Timestamp(t)),
            bson::Bson::Undefined => Expression::Literal(LiteralValue::Undefined),
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub enum LiteralValue {
    Null,
    Boolean(bool),
    String(String),
    Integer(i32),
    Long(i64),
    Double(f64),
    RegularExpression(bson::Regex),
    JavaScriptCode(String),
    JavaScriptCodeWithScope(bson::JavaScriptCodeWithScope),
    Timestamp(bson::Timestamp),
    Binary(bson::Binary),
    ObjectId(bson::oid::ObjectId),
    DateTime(bson::DateTime),
    Symbol(String),
    Decimal128(bson::Decimal128),
    Undefined,
    MaxKey,
    MinKey,
    DbPointer(bson::DbPointer),
}

#[derive(PartialEq, Eq, Debug, Clone, Hash)]
pub struct ReferenceExpr {
    pub key: Key,
}

impl<T: Into<Key>> From<T> for ReferenceExpr {
    fn from(k: T) -> Self {
        ReferenceExpr {
            key: k.into(),
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct ArrayExpr {
    pub array: Vec<Expression>,
}

impl From<Vec<Expression>> for ArrayExpr {
    fn from(array: Vec<Expression>) -> Self {
        ArrayExpr {
            array,
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct DocumentExpr {
    pub document: UniqueLinkedHashMap<String, Expression>,
}

impl<T: Into<UniqueLinkedHashMap<String, Expression>>> From<T> for DocumentExpr {
    fn from(document: T) -> Self {
        DocumentExpr {
            document: document.into(),
        }
    }
}

#[derive(PartialEq, Debug, Clone)]
pub struct ExistsExpr {
    pub stage: Box<Stage>,
}

impl From<Box<Stage>> for ExistsExpr {
    fn from(stage: Box<Stage>) -> Self {
        ExistsExpr {
            stage,
        }
    }
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

#[derive(PartialEq, Debug, Clone, new)]
pub struct ScalarFunctionApplication {
    pub function: ScalarFunction,
    pub args: Vec<Expression>,
    #[new(value = "true")]
    pub is_nullable: bool,

}

#[derive(PartialEq, Debug, Clone, new)]
pub struct FieldAccess {
    pub expr: Box<Expression>,
    pub field: String,
    #[new(value = "true")]
    pub is_nullable: bool,
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
    Replace,
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
    Millisecond,
    Week,
    DayOfWeek,
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
            ScalarFunction::Millisecond => "Millisecond",
            ScalarFunction::Week => "Week",
            ScalarFunction::IsoWeek => "IsoWeek",
            ScalarFunction::IsoWeekday => "IsoWeekday",
            ScalarFunction::DayOfWeek => "DayOfWeek",
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
            ScalarFunction::Replace => "Replace",
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

    // Some functions are always nullable regardless of argument nullablity,
    // we list them here and return true for them. We enumate all functions
    // so this will fail to compile when new functions are added, so that
    // we must consider the semantics.
    pub fn is_always_nullable(&self) -> bool {
        match self {
            // COALESCE(v1, v2, ..., vn) : Returns the first non-NULL argument, or NULL if there are no non-NULL arguments.
            ScalarFunction::Coalesce
            // COS( number ) : If the argument evaluates to negative or positive Infinity, the result of the operation is NULL.ce
            | ScalarFunction::Cos
            // SIN( number ) : If the argument evaluates to negative or positive Infinity, the result of the operation is NULL.
            | ScalarFunction::Sin
            // number / divisor : A divisor of 0 will result in NULL.
            | ScalarFunction::Div
            // MOD( number, divisor ) : A divisor of 0 will result in NULL.
            | ScalarFunction::Mod
            // NULLIF (arg1, arg2) returns NULL if arg1 = arg2
            | ScalarFunction::NullIf
            // ROUND(number, decimals) : The decimals argument must be an integer between -20 and 100. Arguments outside of the supported ranges will return NULL.
            | ScalarFunction::Round
            // TAN( number ) : If the argument evaluates to negative or positive Infinity, the result of the operation is NULL
            | ScalarFunction::Tan
            // SPLIT(string, delimiter, token number) : If any argument is NULL or MISSING, or a delimiter evaluates to an empty string, the result is NULL.
            | ScalarFunction::Split => true,

            ScalarFunction::Add
            | ScalarFunction::And
            | ScalarFunction::BitLength
            | ScalarFunction::CharLength
            | ScalarFunction::ComputedFieldAccess
            | ScalarFunction::Concat
            | ScalarFunction::CurrentTimestamp
            | ScalarFunction::Year
            | ScalarFunction::Month
            | ScalarFunction::Day
            | ScalarFunction::Hour
            | ScalarFunction::Minute
            | ScalarFunction::Second
            | ScalarFunction::Millisecond
            | ScalarFunction::Week
            | ScalarFunction::IsoWeek
            | ScalarFunction::IsoWeekday
            | ScalarFunction::DayOfWeek
            | ScalarFunction::DayOfYear
            | ScalarFunction::Abs
            | ScalarFunction::Ceil
            | ScalarFunction::Degrees
            | ScalarFunction::Eq
            | ScalarFunction::Floor
            | ScalarFunction::Gt
            | ScalarFunction::Gte
            | ScalarFunction::Between
            | ScalarFunction::Log
            | ScalarFunction::Lower
            | ScalarFunction::Lt
            | ScalarFunction::Lte
            | ScalarFunction::Mul
            | ScalarFunction::Neq
            | ScalarFunction::Neg
            | ScalarFunction::Not
            | ScalarFunction::OctetLength
            | ScalarFunction::Or
            | ScalarFunction::Pos
            | ScalarFunction::Position
            | ScalarFunction::Pow
            | ScalarFunction::Radians
            | ScalarFunction::Replace
            | ScalarFunction::Size
            | ScalarFunction::Slice
            | ScalarFunction::Sqrt
            | ScalarFunction::Sub
            | ScalarFunction::Substring
            | ScalarFunction::LTrim
            | ScalarFunction::RTrim
            | ScalarFunction::BTrim
            | ScalarFunction::Upper
            | ScalarFunction::MergeObjects => false
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
    Millisecond,
}

#[derive(PartialEq, Eq, Debug, Clone)]
pub enum DateFunction {
    Add,
    Diff,
    Trunc,
}

#[derive(PartialEq, Debug, Clone, new)]
pub struct DateFunctionApplication {
    pub function: DateFunction,
    pub date_part: DatePart,
    pub args: Vec<Expression>,
    #[new(value = "true")]
    pub is_nullable: bool,

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

#[derive(PartialEq, Debug, Clone, new)]
pub struct SearchedCaseExpr {
    pub when_branch: Vec<WhenBranch>,
    pub else_branch: Box<Expression>,
    #[new(value = "true")]
    pub is_nullable: bool,
}

#[derive(PartialEq, Debug, Clone, new)]
pub struct SimpleCaseExpr {
    pub expr: Box<Expression>,
    pub when_branch: Vec<WhenBranch>,
    pub else_branch: Box<Expression>,
    #[new(value = "true")]
    pub is_nullable: bool,
}

#[derive(PartialEq, Debug, Clone, new)]
pub struct WhenBranch {
    pub when: Box<Expression>,
    pub then: Box<Expression>,
    #[new(value = "true")]
    pub is_nullable: bool,
}

#[derive(PartialEq, Debug, Clone, new)]
pub struct CastExpr {
    pub expr: Box<Expression>,
    pub to: Type,
    pub on_null: Box<Expression>,
    pub on_error: Box<Expression>,
    #[new(value = "true")]
    pub is_nullable: bool,
}

#[derive(PartialEq, Debug, Clone)]
pub struct TypeAssertionExpr {
    pub expr: Box<Expression>,
    pub target_type: Type,
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

#[derive(PartialEq, Debug, Clone, new)]
pub struct SubqueryExpr {
    pub output_expr: Box<Expression>,
    pub subquery: Box<Stage>,
    #[new(value = "true")]
    pub is_nullable: bool,
}

#[derive(PartialEq, Debug, Clone, new)]
pub struct SubqueryComparison {
    pub operator: SubqueryComparisonOp,
    pub modifier: SubqueryModifier,
    pub argument: Box<Expression>,
    pub subquery_expr: SubqueryExpr,
    #[new(value = "true")]
    pub is_nullable: bool,
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
pub enum MatchQuery {
    Logical(MatchLanguageLogical),
    Type(MatchLanguageType),
    Regex(MatchLanguageRegex),
    ElemMatch(ElemMatch),
    Comparison(MatchLanguageComparison),
}

#[derive(Eq, Debug, Clone, new)]
pub struct FieldPath {
    pub key: Key,
    pub fields: Vec<String>,
    #[new(value = "true")]
    pub is_nullable: bool,
}

impl std::hash::Hash for FieldPath {
    fn hash<H: std::hash::Hasher>(&self, state: &mut H) {
        self.key.hash(state);
        self.fields.hash(state);
    }
}

impl PartialEq for FieldPath {
    fn eq(&self, other: &Self) -> bool {
        // ignore is_nullable for equality
        // because we generate FieldPaths in some places where nullability
        // is unknown.
        self.key == other.key
            && self.fields == other.fields
    }
}

impl TryFrom<Expression> for FieldPath {
    type Error = ();

    fn try_from(value: Expression) -> Result<Self, Self::Error> {
        if let Expression::FieldAccess(f) = value {
            f.try_into()
        } else {
            Err(())
        }
    }
}

impl TryFrom<&Expression> for FieldPath {
    type Error = ();

    fn try_from(value: &Expression) -> Result<Self, Self::Error> {
        if let Expression::FieldAccess(ref f) = value {
            f.try_into()
        } else {
            Err(())
        }
    }
}

impl TryFrom<FieldAccess> for FieldPath {
    type Error = ();

    fn try_from(value: FieldAccess) -> Result<Self, Self::Error> {
        let mut cur = *value.expr;
        let mut fields = vec![value.field];
        loop {
            match cur {
                Expression::Reference(ReferenceExpr {key}) => {
                    fields.reverse();
                    return Ok(
                        FieldPath {
                            key,
                            fields,
                            is_nullable: value.is_nullable,
                        });
                }
                Expression::FieldAccess(FieldAccess {expr, field, ..}) => {
                    fields.push(field);
                    cur = *expr;
                }
                _ => {
                    return Err(());
                }
            }
        }
    }
}

impl TryFrom<&FieldAccess> for FieldPath {
    type Error = ();

    fn try_from(value: &FieldAccess) -> Result<Self, Self::Error> {
        let mut cur = value.expr.as_ref();
        let mut fields = vec![value.field.clone()];
        loop {
            match cur {
                Expression::Reference(ReferenceExpr {key}) => {
                    fields.reverse();
                    return Ok(
                        FieldPath {
                            key: key.clone(),
                            fields,
                            is_nullable: value.is_nullable,
                        });
                }
                Expression::FieldAccess(FieldAccess {expr, field, ..}) => {
                    fields.push(field.clone());
                    cur = expr.as_ref();
                }
                _ => {
                    return Err(());
                }
            }
        }
    }
}

#[derive(PartialEq, Eq, Debug, Clone, Hash)]
pub struct Field {
    pub parent: Box<FieldPath>,
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
    pub input: Option<FieldPath>,
    pub target_type: TypeOrMissing,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchLanguageRegex {
    pub input: Option<FieldPath>,
    pub regex: String,
    pub options: String,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct ElemMatch {
    pub input: FieldPath,
    pub condition: Box<MatchQuery>,
    pub cache: SchemaCache<Schema>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct MatchLanguageComparison {
    pub function: MatchLanguageComparisonOp,
    pub input: Option<FieldPath>,
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
