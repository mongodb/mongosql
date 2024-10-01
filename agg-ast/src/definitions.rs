use serde::{
    de::{Error as serde_err, MapAccess, Visitor},
    Deserialize, Deserializer,
};
use std::{collections::HashMap, fmt};

// This module contains an aggregation pipeline syntax tree that implements
// serde::Deserialize. This allows us to deserialize aggregation pipelines from
// test YAML files into structured data and then transform that structured data
// into air structs so that we can run desugarer passes and therefore test the
// desugarers.

/// Stage represents an aggregation pipeline stage. This is not
/// a complete representation of all of MQL. Only stages relevant
/// for desugarer testing are supported here.
#[derive(Debug, PartialEq, Deserialize)]
pub enum Stage {
    #[serde(skip)]
    Collection(Collection),
    #[serde(rename = "$documents")]
    Documents(Vec<HashMap<String, Expression>>),
    #[serde(rename = "$project")]
    Project(HashMap<String, ProjectItem>),
    #[serde(rename = "$replaceWith")]
    ReplaceWith(Expression),
    #[serde(rename = "$match")]
    Match(MatchExpression),
    #[serde(rename = "$limit")]
    Limit(i64),
    #[serde(rename = "$skip")]
    Skip(i64),
    #[serde(rename = "$sort")]
    Sort(HashMap<String, i8>),
    #[serde(rename = "$group")]
    Group(Group),
    #[serde(rename = "$join")]
    Join(Box<Join>),
    #[serde(rename = "$equiJoin")]
    EquiJoin(EquiJoin),
    #[serde(rename = "$unwind")]
    Unwind(Unwind),
    #[serde(rename = "$lookup")]
    Lookup(Lookup),
    #[serde(rename = "$equiLookup")]
    EquiLookup(EquiLookup),
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct Collection {
    pub db: String,
    pub collection: String,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(from = "Expression")]
pub enum ProjectItem {
    Exclusion,
    Inclusion,
    Assignment(Expression),
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct MatchExpression {
    #[serde(rename = "$expr")]
    pub expr: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct Group {
    #[serde(rename = "_id")]
    pub keys: Expression,
    #[serde(flatten)]
    pub aggregations: HashMap<String, GroupAccumulator>,
}

#[derive(Debug, PartialEq)]
pub struct GroupAccumulator {
    pub function: String,
    pub expr: GroupAccumulatorExpr,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub enum GroupAccumulatorExpr {
    SqlAccumulator {
        distinct: bool,
        var: Box<Expression>,
    },
    NonSqlAccumulator(Expression),
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct Join {
    pub database: Option<String>,
    pub collection: Option<String>,
    #[serde(rename = "joinType")]
    pub join_type: JoinType,
    #[serde(rename = "let")]
    pub let_body: Option<HashMap<String, Expression>>,
    pub pipeline: Vec<Stage>,
    pub condition: Option<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct EquiJoin {
    // Note: At the moment equijoin are only supported on collections of the same DB
    pub database: Option<String>,
    pub collection: Option<String>,
    pub join_type: JoinType,
    pub local_field: String,
    pub foreign_field: String,
    #[serde(rename = "as")]
    pub as_var: String,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum JoinType {
    Inner,
    Left,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub enum Unwind {
    Document(UnwindExpr),
    FieldPath(Expression),
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct UnwindExpr {
    pub path: Box<Expression>,
    pub include_array_index: Option<String>,
    pub preserve_null_and_empty_arrays: Option<bool>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct Lookup {
    pub from: Option<LookupFrom>,
    #[serde(rename = "let")]
    pub let_body: Option<HashMap<String, Expression>>,
    pub pipeline: Vec<Stage>,
    #[serde(rename = "as")]
    pub as_var: String,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct EquiLookup {
    pub from: LookupFrom,
    pub local_field: String,
    pub foreign_field: String,
    #[serde(rename = "as")]
    pub as_var: String,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub enum LookupFrom {
    Collection(String),
    Namespace(Namespace),
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct Namespace {
    pub db: String,
    pub coll: String,
}

/// Expression represents an aggregation pipeline expression. This is not
/// a complete representation of all of MQL. Only expressions relevant for
/// desugarer testing are supported here. Order of these variants matters
/// since we use custom deserialization for several expression types.
#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub enum Expression {
    // Non-string literal values
    Literal(LiteralValue),

    // String literal values, or variable or field refs
    #[serde(deserialize_with = "deserialize_string_or_ref")]
    StringOrRef(StringOrRef),

    // Operators with structured arguments
    TaggedOperator(TaggedOperator),

    // Operators with unstructured arguments
    #[serde(deserialize_with = "deserialize_mql_operator")]
    UntaggedOperator(UntaggedOperator),

    // Array literal expressions
    Array(Vec<Expression>),

    // Document literal expressions
    Document(HashMap<String, Expression>),
}

/// StringOrRef represents string constants in the serialized pipelines.
/// String literals, field references, and variable references are all represented
/// by values in double quotes. The only difference is that variables are prefixed
/// with "$$" and field references are prefixed with "$", while all other values
/// are string literals. We need a custom deserializer that distinguishes these
/// types by inspecting the actual string data.
#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub enum StringOrRef {
    String(String),
    FieldRef(String),
    Variable(String),
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub enum LiteralValue {
    Null,
    Boolean(bool),
    Integer(i32),
    Long(i64),
    Double(f64),
}

/// UntaggedOperators are operators that follow the general format:
///   { "$<op_name>": [<args>] }
/// We need a custom deserializer that turns the key "$op_name" into
/// the field "op" in the struct.
#[derive(Debug, PartialEq, Deserialize)]
pub struct UntaggedOperator {
    pub op: String,
    pub args: Vec<Expression>,
}

/// TaggedOperators are operators that have named arguments. We can utilize
/// serde directly for these by using the enum names as the keys (operator names).
#[derive(Debug, PartialEq, Deserialize)]
pub enum TaggedOperator {
    #[serde(rename = "$getField")]
    GetField(GetField),
    #[serde(rename = "$setField")]
    SetField(SetField),
    #[serde(rename = "$unsetField")]
    UnsetField(UnsetField),
    #[serde(rename = "$switch")]
    Switch(Switch),
    #[serde(rename = "$let")]
    Let(Let),
    #[serde(rename = "$sqlConvert")]
    SqlConvert(SqlConvert),
    #[serde(rename = "$convert")]
    Convert(Convert),
    #[serde(rename = "$like")]
    Like(Like),
    #[serde(rename = "$regexMatch")]
    RegexMatch(RegexMatch),
    #[serde(rename = "$sqlDivide")]
    SqlDivide(SqlDivide),
    #[serde(rename = "$trim")]
    Trim(Trim),
    #[serde(rename = "$ltrim")]
    LTrim(Trim),
    #[serde(rename = "$rtrim")]
    RTrim(Trim),
    #[serde(rename = "$subquery")]
    Subquery(Subquery),
    #[serde(rename = "$subqueryComparison")]
    SubqueryComparison(SubqueryComparison),
    #[serde(rename = "$subqueryExists")]
    SubqueryExists(SubqueryExists),

    // Array Operators
    #[serde(rename = "$firstN")]
    FirstN(FirstN),
    #[serde(rename = "$lastN")]
    LastN(LastN),
    #[serde(rename = "$filter")]
    Filter(Filter),
    #[serde(rename = "$map")]
    Map(Map),
    #[serde(rename = "$maxN")]
    MaxNArrayElement(MaxNArrayElement),
    #[serde(rename = "$minN")]
    MinNArrayElement(MinNArrayElement),
    #[serde(rename = "$reduce")]
    Reduce(Reduce),
    #[serde(rename = "$sortArray")]
    SortArray(SortArray),
    #[serde(rename = "$zip")]
    Zip(Zip),
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct GetField {
    pub field: String,
    pub input: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct SetField {
    pub field: String,
    pub input: Box<Expression>,
    pub value: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct UnsetField {
    pub field: String,
    pub input: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct Switch {
    pub branches: Vec<SwitchCase>,
    pub default: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct SwitchCase {
    pub case: Box<Expression>,
    pub then: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct Let {
    pub vars: HashMap<String, Expression>,
    #[serde(rename = "in")]
    pub inside: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SqlConvert {
    pub input: Box<Expression>,
    pub to: String,
    pub on_null: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Convert {
    pub input: Box<Expression>,
    pub to: String,
    pub on_null: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Filter {
    pub input: Box<Expression>,
    #[serde(rename = "as")]
    pub _as: String,
    pub cond: Box<Expression>,
    pub limit: Option<Box<Expression>>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct FirstN {
    pub input: Box<Expression>,
    pub n: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct LastN {
    pub input: Box<Expression>,
    pub n: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct Like {
    pub input: Box<Expression>,
    pub pattern: Box<Expression>,
    pub escape: Option<char>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Map {
    pub input: Box<Expression>,
    #[serde(rename = "as")]
    pub _as: String,
    #[serde(rename = "in")]
    pub inside: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct MaxNArrayElement {
    pub input: Box<Expression>,
    pub n: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct MinNArrayElement {
    pub input: Box<Expression>,
    pub n: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct RegexMatch {
    pub input: Box<Expression>,
    pub regex: Box<Expression>,
    pub options: Option<Box<Expression>>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SqlDivide {
    pub dividend: Box<Expression>,
    pub divisor: Box<Expression>,
    pub on_error: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct Trim {
    pub input: Box<Expression>,
    pub chars: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Reduce {
    pub input: Box<Expression>,
    pub initial_value: Box<Expression>,
    #[serde(rename = "in")]
    pub inside: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SortArray {
    pub input: Box<Expression>,
    pub sort_by: SortArraySpec,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub enum SortArraySpec {
    Value(i8),
    Keys(HashMap<String, i8>),
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Subquery {
    pub db: Option<String>,
    pub collection: Option<String>,
    #[serde(rename = "let")]
    pub let_bindings: Option<HashMap<String, Expression>>,
    pub output_path: Option<Vec<String>>,
    pub pipeline: Vec<Stage>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct SubqueryComparison {
    pub op: String,
    pub modifier: String,
    pub arg: Box<Expression>,
    pub subquery: Box<Subquery>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub struct SubqueryExists {
    pub db: Option<String>,
    pub collection: Option<String>,
    #[serde(rename = "let")]
    pub let_bindings: Option<HashMap<String, Expression>>,
    pub pipeline: Vec<Stage>,
}

fn default_zip_defaults() -> bool {
    false
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Zip {
    pub inputs: Box<Expression>,
    #[serde(default = "default_zip_defaults")]
    pub use_longest_length: bool,
    pub defaults: Option<Box<Expression>>,
}

/// Custom map visitor for identifying and deserializing UntaggedOperators.
struct UntaggedOperatorVisitor {}

impl UntaggedOperatorVisitor {
    fn new() -> Self {
        Self {}
    }
}

/// VecOrSingleExpr represents the argument to UntaggedOperators.
/// Either of the following is valid MQL:
///   { "$sqrt": "$a" }, or
///   { "$sqrt": ["$a"] }
/// So we need to be able to parse either while deserializing an
/// UntaggedOperator. This struct enables that.
#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
enum VecOrSingleExpr {
    Vec(Vec<Expression>),
    Single(Expression),
}

impl VecOrSingleExpr {
    fn get_as_vec(self) -> Vec<Expression> {
        match self {
            VecOrSingleExpr::Vec(v) => v,
            VecOrSingleExpr::Single(e) => vec![e],
        }
    }
}

impl<'de> Visitor<'de> for UntaggedOperatorVisitor {
    type Value = UntaggedOperator;

    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        formatter.write_str("{\"$op\": [args]}")
    }

    fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
    where
        M: MapAccess<'de>,
    {
        let kv = access.next_entry::<String, VecOrSingleExpr>()?;
        if let Some((key, value)) = kv {
            // If the key does not start with a "$", then it is not an agg operator.
            // Ignore this map and stop attempting to deserialize with this function.
            if !key.starts_with('$') {
                return Err(serde_err::custom("ignoring key that does not start with $"));
            }

            // Immediately return when we see one key that starts with a "$".
            // In a general environment, this would be very brittle, however in this
            // controlled test environment, we safely make the assumption that
            // a single key that starts with a "$" is present and indicates an operator.
            return Ok(UntaggedOperator {
                op: key,
                args: value.get_as_vec(),
            });
        }

        Err(serde_err::custom(
            "fail when there are no keys; this lets empty doc be parsed as Document",
        ))
    }
}

/// Custom deserialization function for untagged aggregation operators.
fn deserialize_mql_operator<'de, D>(deserializer: D) -> Result<UntaggedOperator, D::Error>
where
    D: Deserializer<'de>,
{
    deserializer.deserialize_map(UntaggedOperatorVisitor::new())
}

impl<'de> Deserialize<'de> for GroupAccumulator {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        /// Custom map visitor for identifying and deserializing Accumulators.
        struct AccumulatorVisitor;

        impl<'de> Visitor<'de> for AccumulatorVisitor {
            type Value = GroupAccumulator;

            fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
                formatter.write_str("{\"$op\": <expression or struct>}")
            }

            fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
            where
                M: MapAccess<'de>,
            {
                let kv = access.next_entry::<String, GroupAccumulatorExpr>()?;
                if let Some((key, value)) = kv {
                    // If the key does not start with a "$", then it is not an accumulator function.
                    // Ignore this map and stop attempting to deserialize with this function.
                    if !key.starts_with('$') {
                        return Err(serde_err::custom("ignoring key that does not start with $"));
                    }

                    // Immediately return when we see one key that starts with a "$".
                    // In a general environment, this would be very brittle, however in this
                    // controlled test environment, we safely make the assumption that
                    // a single key that starts with a "$" is present and indicates an operator.
                    // let value = value.get_as_vec();
                    return Ok(GroupAccumulator {
                        function: key,
                        expr: value,
                    });
                }

                Err(serde_err::custom("no accumulator could be parsed"))
            }
        }

        const FIELDS: &[&str] = &["function", "expr"];
        deserializer.deserialize_struct("GroupAccumulator", FIELDS, AccumulatorVisitor)
    }
}

/// Custom deserialization function for string constants in agg pipelines.
fn deserialize_string_or_ref<'de, D>(deserializer: D) -> Result<StringOrRef, D::Error>
where
    D: Deserializer<'de>,
{
    let s = String::deserialize(deserializer)?;

    if s.starts_with("$$") {
        Ok(StringOrRef::Variable(s.chars().skip(2).collect()))
    } else if s.starts_with('$') {
        Ok(StringOrRef::FieldRef(s.chars().skip(1).collect()))
    } else {
        Ok(StringOrRef::String(s))
    }
}

impl From<Expression> for ProjectItem {
    fn from(e: Expression) -> Self {
        match e {
            Expression::Literal(LiteralValue::Integer(0)) => ProjectItem::Exclusion,
            Expression::Literal(LiteralValue::Integer(1)) => ProjectItem::Inclusion,
            _ => ProjectItem::Assignment(e),
        }
    }
}
