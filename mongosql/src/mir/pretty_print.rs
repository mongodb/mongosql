//! Pretty-printer for MIR pipeline trees.
//!
//! Renders MIR stages as an indented tree — the same shape as database EXPLAIN output —
//! so that optimizer pass changes are easy to audit visually.
//!
//! # Example
//!
//! ```text
//! FILTER [Lt(foo.int, 42i32)]
//!   COLLECTION [db.foo]
//! ```

// This module is a diagnostic utility with no production callers; suppress the
// dead_code lint so the module compiles cleanly until callers are added.
#![allow(dead_code)]

use super::{
    AggregationExpr, DatePart, ElemMatch, Expression, FieldAccess, FieldPath, JoinType,
    LiteralValue, MatchFalse, MatchLanguageComparison, MatchLanguageComparisonOp,
    MatchLanguageLogical, MatchLanguageLogicalOp, MatchLanguageRegex, MatchLanguageType,
    MatchQuery, MqlStage, OptionallyAliasedExpr, ReferenceExpr, SetOperation, SortSpecification,
    Stage, SubqueryComparisonOp, SubqueryExpr, SubqueryModifier, Type, TypeOrMissing,
};
use mongosql_datastructures::binding_tuple::{DatasourceName, Key};

/// Pretty-prints a MIR data structure to a human-readable indented tree.
pub trait PrettyPrint {
    /// Returns a pretty-printed representation of this value rooted at column 0.
    ///
    /// # Errors
    ///
    /// MIR pretty-printing is infallible; this method never returns `Err`.
    fn pretty_print(&self) -> Result<String>;
}

/// Error type for MIR pretty-printing.
///
/// MIR pretty-printing is infallible, so this enum has no variants.
#[derive(Debug, Clone, Copy, thiserror::Error, PartialEq, Eq)]
pub enum Error {}

type Result<T> = std::result::Result<T, Error>;

/// Indents every line of `s` by `depth` levels (2 spaces each).
fn indent_lines(s: &str, depth: usize) -> String {
    let prefix = "  ".repeat(depth);
    s.lines()
        .map(|line| format!("{prefix}{line}"))
        .collect::<Vec<_>>()
        .join("\n")
}

fn type_str(ty: Type) -> &'static str {
    match ty {
        Type::Array => "Array",
        Type::BinData => "BinData",
        Type::Boolean => "Boolean",
        Type::Datetime => "Datetime",
        Type::DbPointer => "DbPointer",
        Type::Decimal128 => "Decimal128",
        Type::Document => "Document",
        Type::Double => "Double",
        Type::Int32 => "Int32",
        Type::Int64 => "Int64",
        Type::Javascript => "Javascript",
        Type::JavascriptWithScope => "JavascriptWithScope",
        Type::MaxKey => "MaxKey",
        Type::MinKey => "MinKey",
        Type::Null => "Null",
        Type::ObjectId => "ObjectId",
        Type::RegularExpression => "RegularExpression",
        Type::String => "String",
        Type::Symbol => "Symbol",
        Type::Timestamp => "Timestamp",
        Type::Undefined => "Undefined",
    }
}

fn type_or_missing_str(tom: &TypeOrMissing) -> String {
    match tom {
        TypeOrMissing::Missing => "Missing".to_string(),
        TypeOrMissing::Number => "Number".to_string(),
        TypeOrMissing::Type(ty) => type_str(*ty).to_string(),
    }
}

fn date_part_str(dp: &DatePart) -> &'static str {
    match dp {
        DatePart::Year => "Year",
        DatePart::Quarter => "Quarter",
        DatePart::Month => "Month",
        DatePart::Week => "Week",
        DatePart::Day => "Day",
        DatePart::Hour => "Hour",
        DatePart::Minute => "Minute",
        DatePart::Second => "Second",
        DatePart::Millisecond => "Millisecond",
    }
}

fn agg_expr_pp(agg: &AggregationExpr) -> Result<String> {
    match agg {
        AggregationExpr::CountStar(distinct) => {
            if *distinct {
                Ok("CountStarDistinct".to_string())
            } else {
                Ok("CountStar".to_string())
            }
        }
        AggregationExpr::Function(f) => {
            let arg = f.arg.pretty_print()?;
            if f.distinct {
                Ok(format!("{}Distinct({})", f.function.as_str(), arg))
            } else {
                Ok(format!("{}({})", f.function.as_str(), arg))
            }
        }
    }
}

fn subquery_expr_pp(se: &SubqueryExpr) -> Result<String> {
    let output = se.output_expr.pretty_print()?;
    let stage = indent_lines(&se.subquery.pretty_print()?, 1);
    Ok(format!("Subquery({output})\n{stage}"))
}

fn join_type_str(jt: JoinType) -> &'static str {
    match jt {
        JoinType::Inner => "Inner",
        JoinType::Left => "Left",
    }
}

fn subquery_comparison_op_str(op: SubqueryComparisonOp) -> &'static str {
    match op {
        SubqueryComparisonOp::Lt => "Lt",
        SubqueryComparisonOp::Lte => "Lte",
        SubqueryComparisonOp::Neq => "Neq",
        SubqueryComparisonOp::Eq => "Eq",
        SubqueryComparisonOp::Gt => "Gt",
        SubqueryComparisonOp::Gte => "Gte",
    }
}

fn subquery_modifier_str(m: SubqueryModifier) -> &'static str {
    match m {
        SubqueryModifier::Any => "Any",
        SubqueryModifier::All => "All",
    }
}

fn match_comparison_op_str(op: MatchLanguageComparisonOp) -> &'static str {
    match op {
        MatchLanguageComparisonOp::Lt => "Lt",
        MatchLanguageComparisonOp::Lte => "Lte",
        MatchLanguageComparisonOp::Ne => "Ne",
        MatchLanguageComparisonOp::Eq => "Eq",
        MatchLanguageComparisonOp::Gt => "Gt",
        MatchLanguageComparisonOp::Gte => "Gte",
    }
}

impl PrettyPrint for Key {
    fn pretty_print(&self) -> Result<String> {
        let name = match &self.datasource {
            DatasourceName::Bottom => "__bot__",
            DatasourceName::Named(n) => n.as_str(),
        };
        if self.scope == 0 {
            Ok(name.to_string())
        } else {
            Ok(format!("{name}@{}", self.scope))
        }
    }
}

impl PrettyPrint for LiteralValue {
    fn pretty_print(&self) -> Result<String> {
        let s = match self {
            LiteralValue::Null => "null".to_string(),
            LiteralValue::Undefined => "undefined".to_string(),
            LiteralValue::Boolean(b) => b.to_string(),
            LiteralValue::Integer(i) => format!("{i}i32"),
            LiteralValue::Long(l) => format!("{l}i64"),
            LiteralValue::Double(d) => format!("{d}f64"),
            LiteralValue::String(s) => format!("\"{s}\""),
            LiteralValue::Decimal128(d) => format!("Decimal128({d})"),
            LiteralValue::RegularExpression(r) => {
                format!("Regex({}, {})", r.pattern, r.options)
            }
            LiteralValue::JavaScriptCode(code) => format!("Javascript({code})"),
            LiteralValue::JavaScriptCodeWithScope(j) => {
                format!("JavascriptWithScope({})", j.code)
            }
            LiteralValue::Timestamp(t) => format!("Timestamp({}, {})", t.time, t.increment),
            LiteralValue::Binary(b) => {
                format!("Binary({:?}, <{} bytes>)", b.subtype, b.bytes.len())
            }
            LiteralValue::ObjectId(oid) => format!("ObjectId({oid})"),
            LiteralValue::DateTime(dt) => format!("DateTime({})", dt.timestamp_millis()),
            LiteralValue::Symbol(s) => format!("Symbol({s})"),
            LiteralValue::MaxKey => "MaxKey".to_string(),
            LiteralValue::MinKey => "MinKey".to_string(),
            LiteralValue::DbPointer(dp) => format!("DbPointer({dp:?})"),
        };
        Ok(s)
    }
}

impl PrettyPrint for FieldPath {
    fn pretty_print(&self) -> Result<String> {
        let key = self.key.pretty_print()?;
        if self.fields.is_empty() {
            Ok(key)
        } else {
            Ok(format!("{key}.{}", self.fields.join(".")))
        }
    }
}

impl PrettyPrint for SortSpecification {
    fn pretty_print(&self) -> Result<String> {
        match self {
            SortSpecification::Asc(fp) => Ok(format!("{} ASC", fp.pretty_print()?)),
            SortSpecification::Desc(fp) => Ok(format!("{} DESC", fp.pretty_print()?)),
        }
    }
}

impl PrettyPrint for Expression {
    fn pretty_print(&self) -> Result<String> {
        match self {
            Expression::Literal(lit) => lit.pretty_print(),
            Expression::Reference(ReferenceExpr { key }) => Ok(format!("${}", key.pretty_print()?)),
            Expression::FieldAccess(FieldAccess { expr, field, .. }) => {
                Ok(format!("{}.{field}", expr.pretty_print()?))
            }
            Expression::MqlIntrinsicFieldExistence(FieldAccess { expr, field, .. }) => {
                Ok(format!("FieldExists({}.{field})", expr.pretty_print()?))
            }
            Expression::Array(arr) => {
                let elems = arr
                    .array
                    .iter()
                    .map(|e| e.pretty_print())
                    .collect::<Result<Vec<_>>>()?
                    .join(", ");
                Ok(format!("[{elems}]"))
            }
            Expression::Document(doc) => {
                let fields = doc
                    .document
                    .iter()
                    .map(|(k, v)| v.pretty_print().map(|vpp| format!("{k}: {vpp}")))
                    .collect::<Result<Vec<_>>>()?
                    .join(", ");
                Ok(format!("{{{fields}}}"))
            }
            Expression::ScalarFunction(sf) => {
                let args = sf
                    .args
                    .iter()
                    .map(|a| a.pretty_print())
                    .collect::<Result<Vec<_>>>()?
                    .join(", ");
                Ok(format!("{}({args})", sf.function.as_str()))
            }
            Expression::DateFunction(df) => {
                let args = df
                    .args
                    .iter()
                    .map(|a| a.pretty_print())
                    .collect::<Result<Vec<_>>>()?
                    .join(", ");
                Ok(format!(
                    "{}({}, {args})",
                    df.function.as_str(),
                    date_part_str(&df.date_part)
                ))
            }
            Expression::Is(is_expr) => Ok(format!(
                "Is({}, {})",
                is_expr.expr.pretty_print()?,
                type_or_missing_str(&is_expr.target_type)
            )),
            Expression::Like(like_expr) => {
                let expr = like_expr.expr.pretty_print()?;
                let pattern = like_expr.pattern.pretty_print()?;
                if let Some(escape) = like_expr.escape {
                    Ok(format!("Like({expr}, {pattern}, '{escape}')"))
                } else {
                    Ok(format!("Like({expr}, {pattern})"))
                }
            }
            Expression::Cast(cast) => Ok(format!(
                "Cast({}, {}, on_null={}, on_error={})",
                cast.expr.pretty_print()?,
                type_str(cast.to),
                cast.on_null.pretty_print()?,
                cast.on_error.pretty_print()?
            )),
            Expression::TypeAssertion(ta) => Ok(format!(
                "TypeAssertion({}, {})",
                ta.expr.pretty_print()?,
                type_str(ta.target_type)
            )),
            Expression::SearchedCase(sc) => {
                let branches = sc
                    .when_branch
                    .iter()
                    .map(|wb| {
                        let when = wb.when.pretty_print()?;
                        let then = wb.then.pretty_print()?;
                        Ok(format!("When({when}) Then({then})"))
                    })
                    .collect::<Result<Vec<_>>>()?
                    .join(", ");
                let else_br = sc.else_branch.pretty_print()?;
                Ok(format!("SearchedCase({branches}, Else({else_br}))"))
            }
            Expression::SimpleCase(sc) => {
                let operand = sc.expr.pretty_print()?;
                let branches = sc
                    .when_branch
                    .iter()
                    .map(|wb| {
                        let when = wb.when.pretty_print()?;
                        let then = wb.then.pretty_print()?;
                        Ok(format!("When({when}) Then({then})"))
                    })
                    .collect::<Result<Vec<_>>>()?
                    .join(", ");
                let else_br = sc.else_branch.pretty_print()?;
                Ok(format!(
                    "SimpleCase({operand}, {branches}, Else({else_br}))"
                ))
            }
            Expression::Exists(exists) => {
                let stage = indent_lines(&exists.stage.pretty_print()?, 1);
                Ok(format!("Exists\n{stage}"))
            }
            Expression::Subquery(se) => subquery_expr_pp(se),
            Expression::SubqueryComparison(sc) => {
                let op = subquery_comparison_op_str(sc.operator);
                let modifier = subquery_modifier_str(sc.modifier);
                let arg = sc.argument.pretty_print()?;
                let subquery = indent_lines(&sc.subquery_expr.subquery.pretty_print()?, 1);
                let output = sc.subquery_expr.output_expr.pretty_print()?;
                Ok(format!(
                    "SubqueryComparison({op}, {modifier}, {arg}, {output})\n{subquery}"
                ))
            }
        }
    }
}

impl PrettyPrint for MatchQuery {
    fn pretty_print(&self) -> Result<String> {
        match self {
            MatchQuery::Logical(ml) => ml.pretty_print(),
            MatchQuery::Type(mt) => mt.pretty_print(),
            MatchQuery::Regex(mr) => mr.pretty_print(),
            MatchQuery::ElemMatch(em) => em.pretty_print(),
            MatchQuery::Comparison(mc) => mc.pretty_print(),
            MatchQuery::False(mf) => mf.pretty_print(),
        }
    }
}

impl PrettyPrint for MatchLanguageLogical {
    fn pretty_print(&self) -> Result<String> {
        let op = match self.op {
            MatchLanguageLogicalOp::And => "And",
            MatchLanguageLogicalOp::Or => "Or",
        };
        let args = self
            .args
            .iter()
            .map(|a| a.pretty_print())
            .collect::<Result<Vec<_>>>()?
            .join(", ");
        Ok(format!("{op}({args})"))
    }
}

impl PrettyPrint for MatchLanguageType {
    fn pretty_print(&self) -> Result<String> {
        let ty = type_or_missing_str(&self.target_type);
        match &self.input {
            Some(fp) => Ok(format!("MatchType({}, {ty})", fp.pretty_print()?)),
            None => Ok(format!("MatchType({ty})")),
        }
    }
}

impl PrettyPrint for MatchLanguageRegex {
    fn pretty_print(&self) -> Result<String> {
        match &self.input {
            Some(fp) => Ok(format!(
                "MatchRegex({}, {}, {})",
                fp.pretty_print()?,
                self.regex,
                self.options
            )),
            None => Ok(format!("MatchRegex({}, {})", self.regex, self.options)),
        }
    }
}

impl PrettyPrint for ElemMatch {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "ElemMatch({}, {})",
            self.input.pretty_print()?,
            self.condition.pretty_print()?
        ))
    }
}

impl PrettyPrint for MatchLanguageComparison {
    fn pretty_print(&self) -> Result<String> {
        let op = match_comparison_op_str(self.function);
        let arg = self.arg.pretty_print()?;
        match &self.input {
            Some(fp) => Ok(format!("{op}({}, {arg})", fp.pretty_print()?)),
            None => Ok(format!("{op}({arg})")),
        }
    }
}

impl PrettyPrint for MatchFalse {
    fn pretty_print(&self) -> Result<String> {
        Ok("MatchFalse".to_string())
    }
}

impl PrettyPrint for Stage {
    fn pretty_print(&self) -> Result<String> {
        match self {
            Stage::Collection(c) => Ok(format!("COLLECTION [{}.{}]", c.db, c.collection)),
            Stage::Array(arr) => {
                let alias = &arr.alias;
                let elems = arr
                    .array
                    .iter()
                    .map(|e| e.pretty_print())
                    .collect::<Result<Vec<_>>>()?
                    .join(", ");
                Ok(format!("ARRAY [{elems}] AS {alias}"))
            }
            Stage::Sentinel => Ok("SENTINEL".to_string()),
            Stage::Filter(f) => {
                let cond = f.condition.pretty_print()?;
                let source = indent_lines(&f.source.pretty_print()?, 1);
                Ok(format!("FILTER [{cond}]\n{source}"))
            }
            Stage::Project(p) => {
                let header = if p.is_add_fields {
                    "ADD_FIELDS"
                } else {
                    "PROJECT"
                };
                let bindings = p
                    .expression
                    .0
                    .iter()
                    .map(|(k, v)| {
                        let kpp = k.pretty_print()?;
                        let vpp = v.pretty_print()?;
                        Ok(format!("  {kpp} => {vpp}"))
                    })
                    .collect::<Result<Vec<_>>>()?
                    .join("\n");
                let source = indent_lines(&p.source.pretty_print()?, 1);
                Ok(format!("{header}\n{bindings}\n{source}"))
            }
            Stage::Group(g) => {
                let keys_str = if g.keys.is_empty() {
                    "    (none)".to_string()
                } else {
                    g.keys
                        .iter()
                        .map(|oae| match oae {
                            OptionallyAliasedExpr::Aliased(ae) => ae
                                .expr
                                .pretty_print()
                                .map(|e| format!("    {} = {e}", ae.alias)),
                            OptionallyAliasedExpr::Unaliased(e) => {
                                e.pretty_print().map(|e| format!("    {e}"))
                            }
                        })
                        .collect::<Result<Vec<_>>>()?
                        .join("\n")
                };
                let aggs_str = if g.aggregations.is_empty() {
                    "    (none)".to_string()
                } else {
                    g.aggregations
                        .iter()
                        .map(|aa| {
                            agg_expr_pp(&aa.agg_expr).map(|a| format!("    {} = {a}", aa.alias))
                        })
                        .collect::<Result<Vec<_>>>()?
                        .join("\n")
                };
                let source = indent_lines(&g.source.pretty_print()?, 1);
                Ok(format!(
                    "GROUP [scope={}]\n  keys:\n{keys_str}\n  aggs:\n{aggs_str}\n{source}",
                    g.scope
                ))
            }
            Stage::Limit(l) => {
                let source = indent_lines(&l.source.pretty_print()?, 1);
                Ok(format!("LIMIT [{}]\n{source}", l.limit))
            }
            Stage::Offset(o) => {
                let source = indent_lines(&o.source.pretty_print()?, 1);
                Ok(format!("OFFSET [{}]\n{source}", o.offset))
            }
            Stage::Sort(s) => {
                let specs = s
                    .specs
                    .iter()
                    .map(|sp| sp.pretty_print())
                    .collect::<Result<Vec<_>>>()?
                    .join(", ");
                let source = indent_lines(&s.source.pretty_print()?, 1);
                Ok(format!("SORT [{specs}]\n{source}"))
            }
            Stage::Unwind(u) => {
                let path = u.path.pretty_print()?;
                let index = match &u.index {
                    Some(i) => i.as_str(),
                    None => "none",
                };
                let source = indent_lines(&u.source.pretty_print()?, 1);
                Ok(format!(
                    "UNWIND [path={path}, outer={}, index={index}, prefiltered={}]\n{source}",
                    u.outer, u.is_prefiltered
                ))
            }
            Stage::Join(j) => {
                let jt = join_type_str(j.join_type);
                let cond_str = match &j.condition {
                    Some(c) => format!(", condition={}", c.pretty_print()?),
                    None => String::new(),
                };
                let left = indent_lines(&j.left.pretty_print()?, 2);
                let right = indent_lines(&j.right.pretty_print()?, 2);
                Ok(format!(
                    "JOIN [{jt}{cond_str}]\n  LEFT:\n{left}\n  RIGHT:\n{right}"
                ))
            }
            Stage::Set(s) => {
                let op = match s.operation {
                    SetOperation::UnionAll => "UNION_ALL",
                };
                let left = indent_lines(&s.left.pretty_print()?, 2);
                let right = indent_lines(&s.right.pretty_print()?, 2);
                Ok(format!("{op}\n  LEFT:\n{left}\n  RIGHT:\n{right}"))
            }
            Stage::Derived(d) => {
                let source = indent_lines(&d.source.pretty_print()?, 1);
                Ok(format!("DERIVED\n{source}"))
            }
            Stage::MqlIntrinsic(mql) => mql.pretty_print(),
        }
    }
}

impl PrettyPrint for MqlStage {
    fn pretty_print(&self) -> Result<String> {
        match self {
            MqlStage::MatchFilter(mf) => {
                let cond = mf.condition.pretty_print()?;
                let source = indent_lines(&mf.source.pretty_print()?, 1);
                Ok(format!("MATCH_FILTER [{cond}]\n{source}"))
            }
            MqlStage::EquiJoin(ej) => {
                let jt = join_type_str(ej.join_type);
                let local = ej.local_field.pretty_print()?;
                let foreign = ej.foreign_field.pretty_print()?;
                let source = indent_lines(&ej.source.pretty_print()?, 2);
                let from = indent_lines(&ej.from.pretty_print()?, 2);
                Ok(format!(
                    "EQUI_JOIN [{jt}, local={local}, foreign={foreign}]\n  SOURCE:\n{source}\n  FROM:\n{from}"
                ))
            }
            MqlStage::LateralJoin(lj) => {
                let jt = join_type_str(lj.join_type);
                let source = indent_lines(&lj.source.pretty_print()?, 2);
                let subquery = indent_lines(&lj.subquery.pretty_print()?, 2);
                Ok(format!(
                    "LATERAL_JOIN [{jt}]\n  SOURCE:\n{source}\n  SUBQUERY:\n{subquery}"
                ))
            }
        }
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use crate::{
        map,
        mir::{
            self, schema::SchemaCache, AggregationFunction, AggregationFunctionApplication,
            ArrayExpr, ArraySource, Derived, DocumentExpr, ElemMatch, ExistsExpr, FieldAccess,
            Filter, Group, IsExpr, Join, JoinType, LateralJoin, LikeExpr, Limit, MatchFalse,
            MatchFilter, MatchLanguageComparison, MatchLanguageComparisonOp, MatchLanguageLogical,
            MatchLanguageLogicalOp, MatchLanguageRegex, MatchLanguageType, MatchQuery, MqlStage,
            Offset, OptionallyAliasedExpr, Project, ReferenceExpr, ScalarFunction,
            ScalarFunctionApplication, SearchedCaseExpr, Set, SetOperation, SimpleCaseExpr, Sort,
            SortSpecification, Stage, TypeAssertionExpr, TypeOrMissing, Unwind, WhenBranch,
        },
        schema::Satisfaction,
    };
    use mongosql_datastructures::binding_tuple::{BindingTuple, Key};

    fn collection(db: &str, coll: &str) -> Box<Stage> {
        Box::new(Stage::Collection(mir::Collection {
            db: db.to_string(),
            collection: coll.to_string(),
            cache: SchemaCache::new(),
        }))
    }

    fn ref_expr(name: &str) -> Expression {
        Expression::Reference(ReferenceExpr {
            key: Key::named(name, 0),
        })
    }

    fn field_access(name: &str, field: &str) -> Expression {
        Expression::FieldAccess(FieldAccess {
            expr: Box::new(ref_expr(name)),
            field: field.to_string(),
            is_nullable: false,
        })
    }

    // ── Literals ──────────────────────────────────────────────────────────────

    mod literal {
        use super::*;

        #[test]
        fn null_renders_without_type_suffix() {
            assert_eq!(LiteralValue::Null.pretty_print().unwrap(), "null");
        }

        #[test]
        fn undefined_renders() {
            assert_eq!(LiteralValue::Undefined.pretty_print().unwrap(), "undefined");
        }

        #[test]
        fn boolean_true() {
            assert_eq!(LiteralValue::Boolean(true).pretty_print().unwrap(), "true");
        }

        #[test]
        fn boolean_false() {
            assert_eq!(
                LiteralValue::Boolean(false).pretty_print().unwrap(),
                "false"
            );
        }

        #[test]
        fn integer_has_i32_suffix() {
            assert_eq!(LiteralValue::Integer(42).pretty_print().unwrap(), "42i32");
        }

        #[test]
        fn long_has_i64_suffix() {
            assert_eq!(LiteralValue::Long(42).pretty_print().unwrap(), "42i64");
        }

        #[test]
        fn double_has_f64_suffix() {
            assert_eq!(
                LiteralValue::Double(3.14).pretty_print().unwrap(),
                "3.14f64"
            );
        }

        #[test]
        fn string_is_double_quoted() {
            assert_eq!(
                LiteralValue::String("hello".to_string())
                    .pretty_print()
                    .unwrap(),
                r#""hello""#
            );
        }

        #[test]
        fn max_key_renders() {
            assert_eq!(LiteralValue::MaxKey.pretty_print().unwrap(), "MaxKey");
        }

        #[test]
        fn min_key_renders() {
            assert_eq!(LiteralValue::MinKey.pretty_print().unwrap(), "MinKey");
        }
    }

    // ── Key ───────────────────────────────────────────────────────────────────

    mod key {
        use super::*;

        #[test]
        fn named_scope_zero_omits_scope() {
            let k = Key::named("foo", 0);
            assert_eq!(k.pretty_print().unwrap(), "foo");
        }

        #[test]
        fn named_scope_nonzero_shows_at_suffix() {
            let k = Key::named("foo", 1);
            assert_eq!(k.pretty_print().unwrap(), "foo@1");
        }

        #[test]
        fn bottom_key_renders_as_dunder_bot() {
            let k = Key::bot(0);
            assert_eq!(k.pretty_print().unwrap(), "__bot__");
        }
    }

    // ── FieldPath ─────────────────────────────────────────────────────────────

    mod field_path {
        use super::*;
        use crate::mir::FieldPath;

        #[test]
        fn single_field() {
            let fp = FieldPath {
                key: Key::named("orders", 0),
                fields: vec!["status".to_string()],
                is_nullable: false,
            };
            assert_eq!(fp.pretty_print().unwrap(), "orders.status");
        }

        #[test]
        fn multi_field_chain() {
            let fp = FieldPath {
                key: Key::named("orders", 0),
                fields: vec!["customer".to_string(), "id".to_string()],
                is_nullable: false,
            };
            assert_eq!(fp.pretty_print().unwrap(), "orders.customer.id");
        }

        #[test]
        fn bot_key_with_field() {
            let fp = FieldPath {
                key: Key::bot(0),
                fields: vec!["x".to_string()],
                is_nullable: false,
            };
            assert_eq!(fp.pretty_print().unwrap(), "__bot__.x");
        }
    }

    // ── Sort ──────────────────────────────────────────────────────────────────

    mod sort_spec {
        use super::*;
        use crate::mir::FieldPath;

        fn fp(name: &str, field: &str) -> FieldPath {
            FieldPath {
                key: Key::named(name, 0),
                fields: vec![field.to_string()],
                is_nullable: false,
            }
        }

        #[test]
        fn asc_spec() {
            assert_eq!(
                SortSpecification::Asc(fp("orders", "status"))
                    .pretty_print()
                    .unwrap(),
                "orders.status ASC"
            );
        }

        #[test]
        fn desc_spec() {
            assert_eq!(
                SortSpecification::Desc(fp("orders", "amount"))
                    .pretty_print()
                    .unwrap(),
                "orders.amount DESC"
            );
        }
    }

    // ── Expression::Reference ─────────────────────────────────────────────────

    mod reference {
        use super::*;

        #[test]
        fn named_reference_has_dollar_sigil() {
            let expr = ref_expr("foo");
            assert_eq!(expr.pretty_print().unwrap(), "$foo");
        }

        #[test]
        fn bot_reference_uses_dunder_bot() {
            let expr = Expression::Reference(ReferenceExpr { key: Key::bot(0) });
            assert_eq!(expr.pretty_print().unwrap(), "$__bot__");
        }
    }

    // ── Expression::FieldAccess ───────────────────────────────────────────────

    mod field_access_expr {
        use super::*;

        #[test]
        fn single_level() {
            let expr = field_access("foo", "bar");
            assert_eq!(expr.pretty_print().unwrap(), "$foo.bar");
        }

        #[test]
        fn multi_level() {
            let expr = Expression::FieldAccess(FieldAccess {
                expr: Box::new(field_access("foo", "bar")),
                field: "baz".to_string(),
                is_nullable: false,
            });
            assert_eq!(expr.pretty_print().unwrap(), "$foo.bar.baz");
        }
    }

    // ── ScalarFunction ────────────────────────────────────────────────────────

    mod scalar_function {
        use super::*;

        #[test]
        fn unary_function() {
            let expr = Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Upper,
                args: vec![field_access("foo", "name")],
                is_nullable: false,
            });
            assert_eq!(expr.pretty_print().unwrap(), "Upper($foo.name)");
        }

        #[test]
        fn binary_function() {
            let expr = Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    field_access("foo", "int"),
                    Expression::Literal(LiteralValue::Integer(10)),
                ],
                is_nullable: false,
            });
            assert_eq!(expr.pretty_print().unwrap(), "Lt($foo.int, 10i32)");
        }
    }

    // ── Stage::Collection ─────────────────────────────────────────────────────

    mod collection_stage {
        use super::*;

        #[test]
        fn renders_db_dot_collection() {
            let stage = collection("mydb", "mycoll");
            assert_eq!(stage.pretty_print().unwrap(), "COLLECTION [mydb.mycoll]");
        }
    }

    // ── Stage::Filter ─────────────────────────────────────────────────────────

    mod filter_stage {
        use super::*;

        #[test]
        fn literal_condition() {
            let stage = Stage::Filter(Filter {
                condition: Expression::Literal(LiteralValue::Boolean(true)),
                source: collection("db", "foo"),
                cache: SchemaCache::new(),
            });
            assert_eq!(
                stage.pretty_print().unwrap(),
                "FILTER [true]\n  COLLECTION [db.foo]"
            );
        }

        #[test]
        fn scalar_function_condition() {
            let stage = Stage::Filter(Filter {
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Lt,
                    args: vec![
                        field_access("foo", "int"),
                        Expression::Literal(LiteralValue::Integer(10)),
                    ],
                    is_nullable: false,
                }),
                source: collection("db", "foo"),
                cache: SchemaCache::new(),
            });
            assert_eq!(
                stage.pretty_print().unwrap(),
                "FILTER [Lt($foo.int, 10i32)]\n  COLLECTION [db.foo]"
            );
        }

        #[test]
        fn nested_filters() {
            let inner = Box::new(Stage::Filter(Filter {
                condition: Expression::Literal(LiteralValue::Boolean(true)),
                source: collection("db", "foo"),
                cache: SchemaCache::new(),
            }));
            let stage = Stage::Filter(Filter {
                condition: Expression::Literal(LiteralValue::Boolean(false)),
                source: inner,
                cache: SchemaCache::new(),
            });
            assert_eq!(
                stage.pretty_print().unwrap(),
                "FILTER [false]\n  FILTER [true]\n    COLLECTION [db.foo]"
            );
        }
    }

    // ── Stage::Project ────────────────────────────────────────────────────────

    mod project_stage {
        use super::*;

        #[test]
        fn single_binding() {
            let stage = Stage::Project(Project {
                is_add_fields: false,
                source: collection("db", "foo"),
                expression: BindingTuple(map! {
                    Key::named("foo", 0) => ref_expr("foo")
                }),
                cache: SchemaCache::new(),
            });
            assert_eq!(
                stage.pretty_print().unwrap(),
                "PROJECT\n  foo => $foo\n  COLLECTION [db.foo]"
            );
        }

        #[test]
        fn add_fields_uses_different_header() {
            let stage = Stage::Project(Project {
                is_add_fields: true,
                source: collection("db", "foo"),
                expression: BindingTuple(map! {
                    Key::named("foo", 0) => ref_expr("foo")
                }),
                cache: SchemaCache::new(),
            });
            assert!(stage.pretty_print().unwrap().starts_with("ADD_FIELDS"));
        }
    }

    // ── Stage::Group ──────────────────────────────────────────────────────────

    mod group_stage {
        use super::*;
        use crate::mir::{AliasedAggregation, AliasedExpr, FieldPath};

        fn fp(name: &str, field: &str) -> FieldPath {
            FieldPath {
                key: Key::named(name, 0),
                fields: vec![field.to_string()],
                is_nullable: false,
            }
        }

        #[test]
        fn no_keys_count_star() {
            let stage = Stage::Group(Group {
                source: collection("db", "orders"),
                keys: vec![],
                aggregations: vec![AliasedAggregation {
                    alias: "total".to_string(),
                    agg_expr: AggregationExpr::CountStar(false),
                }],
                cache: SchemaCache::new(),
                scope: 0,
            });
            let pp = stage.pretty_print().unwrap();
            assert!(pp.contains("keys:\n    (none)"), "got: {pp}");
            assert!(pp.contains("total = CountStar"), "got: {pp}");
        }

        #[test]
        fn aliased_key_and_sum_aggregation() {
            let stage = Stage::Group(Group {
                source: collection("db", "orders"),
                keys: vec![OptionallyAliasedExpr::Aliased(AliasedExpr {
                    alias: "status".to_string(),
                    expr: Expression::FieldAccess(FieldAccess {
                        expr: Box::new(ref_expr("orders")),
                        field: "status".to_string(),
                        is_nullable: false,
                    }),
                })],
                aggregations: vec![AliasedAggregation {
                    alias: "total".to_string(),
                    agg_expr: AggregationExpr::Function(AggregationFunctionApplication {
                        function: AggregationFunction::Sum,
                        distinct: false,
                        arg: Box::new(Expression::FieldAccess(FieldAccess {
                            expr: Box::new(ref_expr("orders")),
                            field: "amount".to_string(),
                            is_nullable: false,
                        })),
                        arg_is_possibly_doc: Satisfaction::Not,
                    }),
                }],
                cache: SchemaCache::new(),
                scope: 0,
            });
            let pp = stage.pretty_print().unwrap();
            assert!(pp.contains("status = $orders.status"), "got: {pp}");
            assert!(pp.contains("total = Sum($orders.amount)"), "got: {pp}");
        }
    }

    // ── Stage::Join ───────────────────────────────────────────────────────────

    mod join_stage {
        use super::*;

        #[test]
        fn inner_join_no_condition() {
            let stage = Stage::Join(Join {
                join_type: JoinType::Inner,
                left: collection("db", "orders"),
                right: collection("db", "customers"),
                condition: None,
                cache: SchemaCache::new(),
            });
            let pp = stage.pretty_print().unwrap();
            assert_eq!(
                pp,
                "JOIN [Inner]\n  LEFT:\n    COLLECTION [db.orders]\n  RIGHT:\n    COLLECTION [db.customers]"
            );
        }

        #[test]
        fn left_join_with_condition() {
            let stage = Stage::Join(Join {
                join_type: JoinType::Left,
                left: collection("db", "orders"),
                right: collection("db", "customers"),
                condition: Some(Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Eq,
                    args: vec![
                        field_access("orders", "id"),
                        field_access("customers", "order_id"),
                    ],
                    is_nullable: false,
                })),
                cache: SchemaCache::new(),
            });
            let pp = stage.pretty_print().unwrap();
            assert!(pp.starts_with("JOIN [Left, condition=Eq("), "got: {pp}");
        }
    }

    // ── MqlStage::MatchFilter ─────────────────────────────────────────────────

    mod match_filter_stage {
        use super::*;

        #[test]
        fn renders_match_filter_with_condition() {
            let stage = Stage::MqlIntrinsic(MqlStage::MatchFilter(Box::new(MatchFilter {
                source: collection("db", "orders"),
                condition: MatchQuery::Comparison(MatchLanguageComparison {
                    function: MatchLanguageComparisonOp::Gt,
                    input: Some(crate::mir::FieldPath {
                        key: Key::named("orders", 0),
                        fields: vec!["amount".to_string()],
                        is_nullable: false,
                    }),
                    arg: LiteralValue::Integer(100),
                    cache: SchemaCache::new(),
                }),
                cache: SchemaCache::new(),
            })));
            let pp = stage.pretty_print().unwrap();
            assert!(
                pp.starts_with("MATCH_FILTER [Gt(orders.amount, 100i32)]"),
                "got: {pp}"
            );
        }
    }

    // ── Integration: multi-stage pipelines ────────────────────────────────────

    mod integration {
        use super::*;

        #[test]
        fn filter_over_collection() {
            let stage = Stage::Filter(Filter {
                condition: Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Lt,
                    args: vec![
                        field_access("foo", "int"),
                        Expression::Literal(LiteralValue::Integer(10)),
                    ],
                    is_nullable: false,
                }),
                source: collection("db", "foo"),
                cache: SchemaCache::new(),
            });
            assert_eq!(
                stage.pretty_print().unwrap(),
                "FILTER [Lt($foo.int, 10i32)]\n  COLLECTION [db.foo]"
            );
        }

        #[test]
        fn sort_over_filter_over_collection() {
            let filter = Box::new(Stage::Filter(Filter {
                condition: Expression::Literal(LiteralValue::Boolean(true)),
                source: collection("db", "orders"),
                cache: SchemaCache::new(),
            }));
            let stage = Stage::Sort(Sort {
                specs: vec![SortSpecification::Asc(crate::mir::FieldPath {
                    key: Key::named("orders", 0),
                    fields: vec!["status".to_string()],
                    is_nullable: false,
                })],
                source: filter,
                cache: SchemaCache::new(),
            });
            assert_eq!(
                stage.pretty_print().unwrap(),
                "SORT [orders.status ASC]\n  FILTER [true]\n    COLLECTION [db.orders]"
            );
        }
    }
}
