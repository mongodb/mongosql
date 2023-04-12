use crate::air;
use itertools::Itertools;
use linked_hash_map::LinkedHashMap;
use mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap;
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
pub(crate) enum Stage {
    #[serde(skip)]
    Collection(Collection),
    #[serde(rename = "$documents")]
    Documents(Vec<HashMap<String, Expression>>),
    #[serde(rename = "$project")]
    Project(HashMap<String, Expression>),
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
    #[serde(rename = "$unwind")]
    Unwind(Unwind),
    #[serde(rename = "$lookup")]
    Lookup(Lookup),
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct Collection {
    pub(crate) db: String,
    pub(crate) collection: String,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub(crate) enum MatchExpression {
    Expr(MatchExpr),
    NonExpr(Expression),
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct MatchExpr {
    #[serde(rename = "$expr")]
    pub(crate) expr: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct Group {
    #[serde(rename = "_id")]
    pub(crate) keys: HashMap<String, Expression>,
    #[serde(flatten)]
    pub(crate) aggregations: HashMap<String, AccumulatorExpr>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct AccumulatorExpr {
    pub(crate) function: String,
    pub(crate) distinct: bool,
    pub(crate) arg: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct Join {
    pub(crate) database: Option<String>,
    pub(crate) collection: Option<String>,
    #[serde(rename = "joinType")]
    pub(crate) join_type: JoinType,
    #[serde(rename = "let")]
    pub(crate) let_body: Option<HashMap<String, Expression>>,
    pub(crate) pipeline: Vec<Stage>,
    pub(crate) condition: Option<Stage>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) enum JoinType {
    Inner,
    Left,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub(crate) struct Unwind {
    pub(crate) path: String,
    pub(crate) include_array_index: Option<String>,
    pub(crate) preserve_null_and_empty_arrays: Option<bool>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct Lookup {
    pub(crate) from: Option<LookupFrom>,
    #[serde(rename = "let")]
    pub(crate) let_body: Option<HashMap<String, Expression>>,
    pub(crate) pipeline: Vec<Stage>,
    #[serde(rename = "as")]
    pub(crate) as_var: String,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub(crate) enum LookupFrom {
    Collection(String),
    Namespace(Namespace),
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct Namespace {
    pub(crate) db: String,
    pub(crate) coll: String,
}

/// Expression represents an aggregation pipeline expression. This is not
/// a complete representation of all of MQL. Only expressions relevant for
/// desugarer testing are supported here. Order of these variants matters
/// since we use custom deserialization for several expression types.
#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub(crate) enum Expression {
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
pub(crate) enum StringOrRef {
    String(String),
    FieldRef(String),
    Variable(String),
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(untagged)]
pub(crate) enum LiteralValue {
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
pub(crate) struct UntaggedOperator {
    pub(crate) op: String,
    pub(crate) args: Vec<Expression>,
}

/// TaggedOperators are operators that have named arguments. We can utilize
/// serde directly for these by using the enum names as the keys (operator names).
#[derive(Debug, PartialEq, Deserialize)]
pub(crate) enum TaggedOperator {
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
    #[serde(rename = "$sqlDivide")]
    SqlDivide(SqlDivide),
    #[serde(rename = "$subquery")]
    Subquery(Subquery),
    #[serde(rename = "$subqueryComparison")]
    SubqueryComparison(SubqueryComparison),
    #[serde(rename = "$subqueryExists")]
    SubqueryExists(SubqueryExists),
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct GetField {
    pub(crate) field: String,
    pub(crate) input: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct SetField {
    pub(crate) field: String,
    pub(crate) input: Box<Expression>,
    pub(crate) value: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct UnsetField {
    pub(crate) field: String,
    pub(crate) input: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct Switch {
    pub(crate) branches: Vec<SwitchCase>,
    pub(crate) default: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct SwitchCase {
    pub(crate) case: Box<Expression>,
    pub(crate) then: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct Let {
    pub(crate) vars: HashMap<String, Expression>,
    #[serde(rename = "in")]
    pub(crate) inside: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub(crate) struct SqlConvert {
    pub(crate) input: Box<Expression>,
    pub(crate) to: String,
    pub(crate) on_null: Box<Expression>,
    pub(crate) on_error: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub(crate) struct Convert {
    pub(crate) input: Box<Expression>,
    pub(crate) to: String,
    pub(crate) on_null: Box<Expression>,
    pub(crate) on_error: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct Like {
    pub(crate) input: Box<Expression>,
    pub(crate) pattern: Box<Expression>,
    pub(crate) escape: Option<String>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub(crate) struct SqlDivide {
    pub(crate) dividend: Box<Expression>,
    pub(crate) divisor: Box<Expression>,
    pub(crate) on_error: Box<Expression>,
}

#[derive(Debug, PartialEq, Deserialize)]
#[serde(rename_all = "camelCase")]
pub(crate) struct Subquery {
    pub(crate) db: Option<String>,
    pub(crate) collection: Option<String>,
    #[serde(rename = "let")]
    pub(crate) let_bindings: Option<HashMap<String, Expression>>,
    pub(crate) output_path: Option<Vec<String>>,
    pub(crate) pipeline: Vec<Stage>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct SubqueryComparison {
    pub(crate) op: String,
    pub(crate) modifier: String,
    pub(crate) arg: Box<Expression>,
    pub(crate) subquery: Box<Subquery>,
}

#[derive(Debug, PartialEq, Deserialize)]
pub(crate) struct SubqueryExists {
    pub(crate) db: Option<String>,
    pub(crate) collection: Option<String>,
    #[serde(rename = "let")]
    pub(crate) let_bindings: Option<HashMap<String, Expression>>,
    pub(crate) pipeline: Vec<Stage>,
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

impl From<(air::Stage, Stage)> for air::Stage {
    fn from((source, ast_stage): (air::Stage, Stage)) -> Self {
        match ast_stage {
            Stage::Collection(c) => air::Stage::Collection(air::Collection {
                db: c.db,
                collection: c.collection,
            }),
            Stage::Documents(d) => {
                let array: Vec<air::Expression> = d
                    .into_iter()
                    .map(|m| Expression::Document(m).into())
                    .collect();

                air::Stage::Documents(air::Documents { array })
            }
            Stage::Project(p) => {
                let specs = to_unique_linked_hash_map_of_air_exprs(p);

                air::Stage::Project(air::Project {
                    source: Box::new(source),
                    specifications: specs,
                })
            }
            Stage::ReplaceWith(r) => air::Stage::ReplaceWith(air::ReplaceWith {
                source: Box::new(source),
                new_root: Box::new(r.into()),
            }),
            Stage::Match(m) => {
                let expr = match m {
                    MatchExpression::Expr(e) => *e.expr,
                    MatchExpression::NonExpr(e) => e,
                };

                air::Stage::Match(air::Match {
                    source: Box::new(source),
                    expr: Box::new(expr.into()),
                })
            }
            Stage::Limit(l) => air::Stage::Limit(air::Limit {
                source: Box::new(source),
                limit: l,
            }),
            Stage::Skip(s) => air::Stage::Skip(air::Skip {
                source: Box::new(source),
                skip: s,
            }),
            Stage::Sort(s) => {
                let specs = s
                    .into_iter()
                    // sort alphabetically for testing purposes -- this obviously changes sort semantics, however that doesn't matter for desugarer tests
                    .sorted_by(|a, b| Ord::cmp(&a.0, &b.0))
                    .map(|(k, v)| {
                        if v >= 1 {
                            air::SortSpecification::Asc(k)
                        } else if v <= -1 {
                            air::SortSpecification::Desc(k)
                        } else {
                            panic!("sort spec cannot be 0 but was for '{}'", k)
                        }
                    })
                    .collect::<Vec<air::SortSpecification>>();

                air::Stage::Sort(air::Sort {
                    source: Box::new(source),
                    specs,
                })
            }
            // TODO: SQL-1318 finish implementing stages
            Stage::Group(_) => todo!(),
            Stage::Join(_) => todo!(),
            Stage::Unwind(_) => todo!(),
            Stage::Lookup(_) => todo!(),
        }
    }
}

fn to_unique_linked_hash_map_of_air_exprs(
    hm: HashMap<String, Expression>,
) -> UniqueLinkedHashMap<String, air::Expression> {
    let lhm: LinkedHashMap<String, air::Expression> = hm
        .into_iter()
        // sort alphabetically for testing purposes
        .sorted_by(|a, b| Ord::cmp(&a.0, &b.0))
        .map(|(k, v)| (k, air::Expression::from(v)))
        .collect::<_>();

    // We do not worry about DuplicateKeyErrors in this controlled test environment.
    UniqueLinkedHashMap::from(lhm)
}

impl From<Expression> for air::Expression {
    fn from(ast_expr: Expression) -> Self {
        match ast_expr {
            Expression::Literal(lv) => air::Expression::Literal(air::LiteralValue::from(lv)),
            Expression::StringOrRef(v) => match v {
                StringOrRef::String(s) => air::Expression::Literal(air::LiteralValue::String(s)),
                StringOrRef::FieldRef(s) => {
                    air::Expression::FieldRef(ref_from_dotted_field_path(s))
                }
                StringOrRef::Variable(s) => {
                    air::Expression::Variable(var_from_dotted_field_path(s))
                }
            },
            Expression::TaggedOperator(to) => to.into(),
            Expression::UntaggedOperator(uo) => uo.into(),
            Expression::Array(a) => {
                air::Expression::Array(a.into_iter().map(air::Expression::from).collect::<Vec<_>>())
            }
            Expression::Document(m) => {
                let ulhm = to_unique_linked_hash_map_of_air_exprs(m);

                air::Expression::Document(ulhm)
            }
        }
    }
}

impl From<LiteralValue> for air::LiteralValue {
    fn from(ast_lv: LiteralValue) -> Self {
        match ast_lv {
            LiteralValue::Null => air::LiteralValue::Null,
            LiteralValue::Boolean(v) => air::LiteralValue::Boolean(v),
            LiteralValue::Integer(v) => air::LiteralValue::Integer(v),
            LiteralValue::Long(v) => air::LiteralValue::Long(v),
            LiteralValue::Double(v) => air::LiteralValue::Double(v),
        }
    }
}

fn ref_from_dotted_field_path(p: String) -> air::FieldRef {
    let r = p.split('.').fold(None, |acc, n| {
        Some(Box::new(air::FieldRef {
            parent: acc,
            name: n.to_string(),
        }))
    });

    // The fold operation will always result in a non-None value, so unwrapping is safe.
    *r.unwrap()
}

fn var_from_dotted_field_path(p: String) -> air::Variable {
    let v = p.split('.').fold(None, |acc, n| {
        Some(Box::new(air::Variable {
            parent: acc,
            name: n.to_string(),
        }))
    });

    // The fold operation will always result in a non-None value, so unwrapping is safe.
    *v.unwrap()
}

impl From<Box<Expression>> for Box<air::Expression> {
    fn from(e: Box<Expression>) -> Self {
        Box::new((*e).into())
    }
}

impl From<TaggedOperator> for air::Expression {
    fn from(ast_op: TaggedOperator) -> Self {
        match ast_op {
            TaggedOperator::GetField(gf) => air::Expression::GetField(air::GetField {
                field: gf.field,
                input: gf.input.into(),
            }),
            TaggedOperator::SetField(sf) => air::Expression::SetField(air::SetField {
                field: sf.field,
                input: sf.input.into(),
                value: sf.value.into(),
            }),
            TaggedOperator::UnsetField(uf) => air::Expression::UnsetField(air::UnsetField {
                field: uf.field,
                input: uf.input.into(),
            }),
            TaggedOperator::Switch(s) => air::Expression::Switch(air::Switch {
                branches: s
                    .branches
                    .into_iter()
                    .map(|b| air::SwitchCase {
                        case: b.case.into(),
                        then: b.then.into(),
                    })
                    .collect(),
                default: s.default.into(),
            }),
            TaggedOperator::Let(l) => air::Expression::Let(air::Let {
                vars: l
                    .vars
                    .into_iter()
                    // sort alphabetically for testing purposes
                    .sorted_by(|a, b| Ord::cmp(&a.0, &b.0))
                    .map(|(k, v)| air::LetVariable {
                        name: k,
                        expr: Box::new(v.into()),
                    })
                    .collect(),
                inside: l.inside.into(),
            }),
            TaggedOperator::SqlConvert(c) => {
                let target_type = match c.to.as_str() {
                    "array" => air::SqlConvertTargetType::Array,
                    "object" => air::SqlConvertTargetType::Document,
                    _ => panic!("invalid '$sqlConvert' target type '{}'", c.to),
                };

                air::Expression::SqlConvert(air::SqlConvert {
                    input: c.input.into(),
                    to: target_type,
                    on_null: c.on_null.into(),
                    on_error: c.on_error.into(),
                })
            }
            TaggedOperator::Convert(c) => air::Expression::Convert(air::Convert {
                input: c.input.into(),
                to: str_to_air_type(c.to),
                on_null: c.on_null.into(),
                on_error: c.on_error.into(),
            }),
            TaggedOperator::Like(l) => air::Expression::Like(air::Like {
                expr: l.input.into(),
                pattern: l.pattern.into(),
                escape: l.escape,
            }),
            TaggedOperator::SqlDivide(d) => {
                air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                    op: air::SQLOperator::Divide,
                    args: vec![(*d.dividend).into(), (*d.divisor).into()],
                })
            }
            // TODO: SQL-1319 finish implementing expressions
            TaggedOperator::Subquery(_) => todo!(),
            TaggedOperator::SubqueryComparison(_) => todo!(),
            TaggedOperator::SubqueryExists(_) => todo!(),
        }
    }
}

fn str_to_air_type(t: String) -> air::Type {
    match t.as_str() {
        "array" => air::Type::Array,
        "binData" => air::Type::BinData,
        "bool" => air::Type::Boolean,
        "date" => air::Type::Datetime,
        "dbPointer" => air::Type::DbPointer,
        "decimal" => air::Type::Decimal128,
        "object" => air::Type::Document,
        "double" => air::Type::Double,
        "int" => air::Type::Int32,
        "long" => air::Type::Int64,
        "javascript" => air::Type::Javascript,
        "javascriptWithScope" => air::Type::JavascriptWithScope,
        "maxKey" => air::Type::MaxKey,
        "minKey" => air::Type::MinKey,
        "null" => air::Type::Null,
        "objectId" => air::Type::ObjectId,
        "regex" => air::Type::RegularExpression,
        "string" => air::Type::String,
        "symbol" => air::Type::Symbol,
        "timestamp" => air::Type::Timestamp,
        "undefined" => air::Type::Undefined,
        _ => panic!("invalid $convert or $is target type '{}'", t),
    }
}

impl From<UntaggedOperator> for air::Expression {
    fn from(ast_op: UntaggedOperator) -> Self {
        let args: Vec<air::Expression> =
            ast_op.args.into_iter().map(air::Expression::from).collect();

        // Special cases:
        //   - $literal becomes a Literal
        //   - $is/$sqlIs become an Is
        //   - $nullIf and $coalesce are SQL operators but don't start with $sql
        match ast_op.op.as_str() {
            "$literal" => {
                let arg = args.get(0).unwrap();
                match arg {
                    air::Expression::Literal(_) => return arg.clone(),
                    _ => panic!("invalid $literal"),
                }
            }
            "$is" | "$sqlIs" => {
                let expr = Box::new(args.get(0).unwrap().clone());

                let arg2 = match args.get(1).unwrap() {
                    air::Expression::Literal(air::LiteralValue::String(s)) => s,
                    _ => panic!("invalid second arg to $is/$sqlIs"),
                };
                let target_type = to_type_or_missing(arg2.clone());

                return air::Expression::Is(air::Is { expr, target_type });
            }
            "$nullIf" => {
                return air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                    op: air::SQLOperator::NullIf,
                    args,
                })
            }
            "$coalesce" => {
                return air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                    op: air::SQLOperator::Coalesce,
                    args,
                })
            }
            _ => (),
        }

        if ast_op.op.starts_with("$sql") {
            let op = match ast_op.op.as_str() {
                "$sqlDivide" => air::SQLOperator::Divide,
                "$sqlPos" => air::SQLOperator::Pos,
                "$sqlNeg" => air::SQLOperator::Neg,
                "$sqlLt" => air::SQLOperator::Lt,
                "$sqlLte" => air::SQLOperator::Lte,
                "$sqlNe" => air::SQLOperator::Ne,
                "$sqlEq" => air::SQLOperator::Eq,
                "$sqlGt" => air::SQLOperator::Gt,
                "$sqlGte" => air::SQLOperator::Gte,
                "$sqlBetween" => air::SQLOperator::Between,
                "$sqlNot" => air::SQLOperator::Not,
                "$sqlAnd" => air::SQLOperator::And,
                "$sqlOr" => air::SQLOperator::Or,
                "$sqlSlice" => air::SQLOperator::Slice,
                "$sqlSize" => air::SQLOperator::Size,
                "$sqlIndexOfCP" => air::SQLOperator::IndexOfCP,
                "$sqlStrLenCP" => air::SQLOperator::StrLenCP,
                "$sqlStrLenBytes" => air::SQLOperator::StrLenBytes,
                "$sqlBitLength" => air::SQLOperator::BitLength,
                "$sqlCos" => air::SQLOperator::Cos,
                "$sqlLog" => air::SQLOperator::Log,
                "$sqlMod" => air::SQLOperator::Mod,
                "$sqlRound" => air::SQLOperator::Round,
                "$sqlSin" => air::SQLOperator::Sin,
                "$sqlSqrt" => air::SQLOperator::Sqrt,
                "$sqlTan" => air::SQLOperator::Tan,
                "$sqlSubstrCP" => air::SQLOperator::SubstrCP,
                "$sqlToUpper" => air::SQLOperator::ToUpper,
                "$sqlToLower" => air::SQLOperator::ToLower,
                "$sqlSplit" => air::SQLOperator::Split,
                _ => panic!("invalid sql operator '{}'", ast_op.op),
            };

            air::Expression::SQLSemanticOperator(air::SQLSemanticOperator { op, args })
        } else {
            let op = match ast_op.op.as_str() {
                "$concat" => air::MQLOperator::Concat,
                "$cond" => air::MQLOperator::Cond,
                "$add" => air::MQLOperator::Add,
                "$subtract" => air::MQLOperator::Subtract,
                "$multiply" => air::MQLOperator::Multiply,
                "$divide" => air::MQLOperator::Divide,
                "$lt" => air::MQLOperator::Lt,
                "$lte" => air::MQLOperator::Lte,
                "$ne" => air::MQLOperator::Ne,
                "$eq" => air::MQLOperator::Eq,
                "$gt" => air::MQLOperator::Gt,
                "$gte" => air::MQLOperator::Gte,
                "$not" => air::MQLOperator::Not,
                "$and" => air::MQLOperator::And,
                "$or" => air::MQLOperator::Or,
                "$slice" => air::MQLOperator::Slice,
                "$size" => air::MQLOperator::Size,
                "$indexOfCP" => air::MQLOperator::IndexOfCP,
                "$indexOfBytes" => air::MQLOperator::IndexOfBytes,
                "$strLenCP" => air::MQLOperator::StrLenCP,
                "$strLenBytes" => air::MQLOperator::StrLenBytes,
                "$abs" => air::MQLOperator::Abs,
                "$ceil" => air::MQLOperator::Ceil,
                "$cos" => air::MQLOperator::Cos,
                "$degreesToRadians" => air::MQLOperator::DegreesToRadians,
                "$floor" => air::MQLOperator::Floor,
                "$log" => air::MQLOperator::Log,
                "$mod" => air::MQLOperator::Mod,
                "$pow" => air::MQLOperator::Pow,
                "$radiansToDegrees" => air::MQLOperator::RadiansToDegrees,
                "$round" => air::MQLOperator::Round,
                "$sin" => air::MQLOperator::Sin,
                "$tan" => air::MQLOperator::Tan,
                "$sqrt" => air::MQLOperator::Sqrt,
                "$substrCP" => air::MQLOperator::SubstrCP,
                "$substrBytes" => air::MQLOperator::SubstrBytes,
                "$toUpper" => air::MQLOperator::ToUpper,
                "$toLower" => air::MQLOperator::ToLower,
                "$trim" => air::MQLOperator::Trim,
                "$ltrim" => air::MQLOperator::LTrim,
                "$rtrim" => air::MQLOperator::RTrim,
                "$split" => air::MQLOperator::Split,
                "$year" => air::MQLOperator::Year,
                "$month" => air::MQLOperator::Month,
                "$dayOfMonth" => air::MQLOperator::DayOfMonth,
                "$hour" => air::MQLOperator::Hour,
                "$minute" => air::MQLOperator::Minute,
                "$second" => air::MQLOperator::Second,
                "$week" => air::MQLOperator::Week,
                "$dayOfYear" => air::MQLOperator::DayOfYear,
                "$isoWeek" => air::MQLOperator::IsoWeek,
                "$isoDayOfWeek" => air::MQLOperator::IsoDayOfWeek,
                "$dateAdd" => air::MQLOperator::DateAdd,
                "$dateDiff" => air::MQLOperator::DateDiff,
                "$dateTrunc" => air::MQLOperator::DateTrunc,
                "$mergeObjects" => air::MQLOperator::MergeObjects,
                _ => panic!("invalid mql operator '{}'", ast_op.op),
            };

            air::Expression::MQLSemanticOperator(air::MQLSemanticOperator { op, args })
        }
    }
}

fn to_type_or_missing(s: String) -> air::TypeOrMissing {
    match s.as_str() {
        "missing" => air::TypeOrMissing::Missing,
        "number" => air::TypeOrMissing::Number,
        _ => air::TypeOrMissing::Type(str_to_air_type(s)),
    }
}
