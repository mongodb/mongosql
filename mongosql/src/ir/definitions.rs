use crate::{
    ir::{
        binding_tuple::{BindingTuple, Key},
        schema::SchemaCache,
    },
    schema::{ResultSet, Schema},
    util::unique_linked_hash_map::UniqueLinkedHashMap,
};

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
pub struct Limit {
    pub source: Box<Stage>,
    pub limit: u64,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Offset {
    pub source: Box<Stage>,
    pub offset: u64,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Sort {
    pub source: Box<Stage>,
    pub specs: Vec<SortSpecification>,
    pub cache: SchemaCache<ResultSet>,
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
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SetOperation {
    UnionAll,
    Union,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Derived {
    pub source: Box<Stage>,
    pub cache: SchemaCache<ResultSet>,
}

#[derive(PartialEq, Debug, Clone)]
pub enum Expression {
    Literal(LiteralExpr),
    Reference(ReferenceExpr),
    Array(ArrayExpr),
    Document(DocumentExpr),
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
    Exists(ExistsExpr),
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
    pub fn as_str(&self) -> &'static str {
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
