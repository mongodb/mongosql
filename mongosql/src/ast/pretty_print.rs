use crate::ast::*;
use lazy_static::lazy_static;
use regex::RegexSet;
use thiserror::Error;

lazy_static! {
    pub static ref KEYWORDS: RegexSet = RegexSet::new(&[
        r"(?i)aggregate$",
        r"(?i)all$",
        r"(?i)and$",
        r"(?i)any$",
        r"(?i)array$",
        r"(?i)as$",
        r"(?i)asc$",
        r"(?i)between$",
        r"(?i)bindata$",
        r"(?i)bit$",
        r"(?i)bool$",
        r"(?i)boolean$",
        r"(?i)both$",
        r"(?i)bson_date$",
        r"(?i)bson_timestamp$",
        r"(?i)by$",
        r"(?i)case$",
        r"(?i)cast$",
        r"(?i)char$",
        r"(?i)character$",
        r"(?i)char\s+varying$",
        r"(?i)character\s+varying$",
        r"(?i)cross$",
        r"(?i)current_timestamp$",
        r"(?i)day$",
        r"(?i)dbpointer$",
        r"(?i)dec$",
        r"(?i)decimal$",
        r"(?i)desc$",
        r"(?i)distinct$",
        r"(?i)document$",
        r"(?i)double$",
        r"(?i)else$",
        r"(?i)end$",
        r"(?i)error$",
        r"(?i)escape$",
        r"(?i)exists$",
        r"(?i)extract$",
        r"(?i)false$",
        r"(?i)fetch\s+first$",
        r"(?i)fetch\s+last$",
        r"(?i)float$",
        r"(?i)for$",
        r"(?i)from$",
        r"(?i)group$",
        r"(?i)having$",
        r"(?i)hour$",
        r"(?i)in$",
        r"(?i)inner$",
        r"(?i)int$",
        r"(?i)integer$",
        r"(?i)is$",
        r"(?i)javascript$",
        r"(?i)javascriptwithscope$",
        r"(?i)join$",
        r"(?i)leading$",
        r"(?i)like$",
        r"(?i)left$",
        r"(?i)limit$",
        r"(?i)long$",
        r"(?i)maxkey$",
        r"(?i)minkey$",
        r"(?i)minute$",
        r"(?i)missing$",
        r"(?i)month$",
        r"(?i)natural$",
        r"(?i)not$",
        r"(?i)not\s+in$",
        r"(?i)not\s+like$",
        r"(?i)null$",
        r"(?i)numeric$",
        r"(?i)number$",
        r"(?i)objectid$",
        r"(?i)offset$",
        r"(?i)on$",
        r"(?i)or$",
        r"(?i)order$",
        r"(?i)outer$",
        r"(?i)position$",
        r"(?i)precision$",
        r"(?i)real$",
        r"(?i)regex$",
        r"(?i)right$",
        r"(?i)rows?\s+only$",
        r"(?i)second$",
        r"(?i)smallint$",
        r"(?i)some$",
        r"(?i)string$",
        r"(?i)substring$",
        r"(?i)symbol$",
        r"(?i)then$",
        r"(?i)timestamp$",
        r"(?i)trailing$",
        r"(?i)trim$",
        r"(?i)true$",
        r"(?i)undefined$",
        r"(?i)union$",
        r"(?i)value$",
        r"(?i)values$",
        r"(?i)varchar$",
        r"(?i)when$",
        r"(?i)where$",
        r"(?i)year$",
    ])
    .unwrap();
}

fn identifier_to_string(s: &str) -> String {
    if ident_needs_delimiters(s) {
        format!("`{}`", s.replace('`', "``"))
    } else {
        s.to_string()
    }
}

#[derive(Debug, Clone, Copy, Error, PartialEq, Eq)]
pub enum Error {
    #[error("set queries cannot be right associative")]
    RightAssociativeSetQuery,
    #[error("joins cannot be right associative")]
    RightAssociativeJoin,
}

type Result<T> = std::result::Result<T, Error>;

// Escape single quotes in string literal by doubling them.
fn escape_string_literal(s: &str) -> String {
    s.replace('\'', "''")
}

/// A trait for any data structure that supports pretty printing to a String
/// that has the posssibility of failing.
pub trait PrettyPrint {
    fn pretty_print(&self) -> Result<String>;
}

impl PrettyPrint for Query {
    fn pretty_print(&self) -> Result<String> {
        match self {
            Query::Select(q) => q.pretty_print(),
            Query::Set(q) => q.pretty_print(),
        }
    }
}

impl PrettyPrint for SetQuery {
    fn pretty_print(&self) -> Result<String> {
        if matches!(self.right.as_ref(), Query::Set(_)) {
            return Err(Error::RightAssociativeSetQuery);
        }
        Ok(format!(
            "{} {} {}",
            self.left.pretty_print()?,
            self.op.pretty_print()?,
            self.right.pretty_print()?
        ))
    }
}

impl PrettyPrint for SetOperator {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            SetOperator::Union => "UNION",
            SetOperator::UnionAll => "UNION ALL",
        }
        .to_string())
    }
}

impl PrettyPrint for SelectQuery {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "{}{}{}{}{}{}{}{}",
            self.select_clause.pretty_print()?,
            self.from_clause
                .as_ref()
                .map_or(Ok("".to_string()), |x| Ok(format!(
                    " FROM {}",
                    x.pretty_print()?
                )))?,
            self.where_clause
                .as_ref()
                .map_or(Ok("".to_string()), |x| Ok(format!(
                    " WHERE {}",
                    x.pretty_print()?
                )))?,
            self.group_by_clause
                .as_ref()
                .map_or(Ok("".to_string()), |x| x.pretty_print())?,
            self.having_clause
                .as_ref()
                .map_or(Ok("".to_string()), |x| Ok(format!(
                    " HAVING {}",
                    x.pretty_print()?
                )))?,
            self.order_by_clause
                .as_ref()
                .map_or(Ok("".to_string()), |x| x.pretty_print())?,
            self.limit
                .map_or("".to_string(), |x| format!(" LIMIT {}", x)),
            self.offset
                .map_or("".to_string(), |x| format!(" OFFSET {}", x)),
        ))
    }
}

impl PrettyPrint for SelectClause {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "SELECT{}{}",
            self.set_quantifier.pretty_print()?,
            self.body.pretty_print()?
        ))
    }
}

impl PrettyPrint for SetQuantifier {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            SetQuantifier::All => "",
            SetQuantifier::Distinct => " DISTINCT",
        }
        .to_string())
    }
}

impl PrettyPrint for SelectBody {
    fn pretty_print(&self) -> Result<String> {
        match self {
            SelectBody::Standard(v) => Ok(format!(
                " {}",
                v.iter()
                    .map(|x| x.pretty_print())
                    .collect::<Result<Vec<_>>>()?
                    .join(", ")
            )),
            SelectBody::Values(v) => {
                let kwd = if v.len() > 1 { "VALUES" } else { "VALUE" };
                Ok(format!(
                    " {} {}",
                    kwd,
                    v.iter()
                        .map(|x| x.pretty_print())
                        .collect::<Result<Vec<_>>>()?
                        .join(", ")
                ))
            }
        }
    }
}

impl PrettyPrint for SelectExpression {
    fn pretty_print(&self) -> Result<String> {
        match self {
            SelectExpression::Star => Ok("*".to_string()),
            SelectExpression::Substar(s) => s.pretty_print(),
            SelectExpression::Expression(e) => e.pretty_print(),
        }
    }
}

impl PrettyPrint for SubstarExpr {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "{}.*",
            identifier_to_string(self.datasource.as_str())
        ))
    }
}

impl PrettyPrint for OptionallyAliasedExpr {
    fn pretty_print(&self) -> Result<String> {
        match self {
            OptionallyAliasedExpr::Aliased(ae) => ae.pretty_print(),
            OptionallyAliasedExpr::Unaliased(e) => e.pretty_print(),
        }
    }
}

impl PrettyPrint for AliasedExpr {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "{} AS {}",
            self.expr.pretty_print()?,
            identifier_to_string(self.alias.as_ref())
        ))
    }
}

impl PrettyPrint for SelectValuesExpression {
    fn pretty_print(&self) -> Result<String> {
        match self {
            SelectValuesExpression::Substar(s) => s.pretty_print(),
            SelectValuesExpression::Expression(e) => e.pretty_print(),
        }
    }
}

impl PrettyPrint for Datasource {
    fn pretty_print(&self) -> Result<String> {
        match self {
            Datasource::Array(a) => a.pretty_print(),
            Datasource::Collection(c) => c.pretty_print(),
            Datasource::Derived(d) => d.pretty_print(),
            Datasource::Join(j) => j.pretty_print(),
        }
    }
}

impl PrettyPrint for ArraySource {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "[{}] AS {}",
            self.array
                .iter()
                .map(|x| x.pretty_print())
                .collect::<Result<Vec<_>>>()?
                .join(", "),
            identifier_to_string(self.alias.as_str()),
        ))
    }
}

impl PrettyPrint for CollectionSource {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "{}{}{}",
            self.database.as_ref().map_or("".to_string(), |x| format!(
                "{}.",
                identifier_to_string(x.as_str())
            )),
            identifier_to_string(self.collection.as_str()),
            self.alias.as_ref().map_or("".to_string(), |x| format!(
                " AS {}",
                identifier_to_string(x)
            )),
        ))
    }
}

impl PrettyPrint for DerivedSource {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "({}) AS {}",
            self.query.pretty_print()?,
            identifier_to_string(self.alias.as_str())
        ))
    }
}

impl PrettyPrint for JoinSource {
    fn pretty_print(&self) -> Result<String> {
        if matches!(self.right.as_ref(), Datasource::Join(_)) {
            return Err(Error::RightAssociativeJoin);
        }
        Ok(format!(
            "{} {} JOIN {}{}",
            self.left.pretty_print()?,
            self.join_type.pretty_print()?,
            self.right.pretty_print()?,
            self.condition
                .as_ref()
                .map_or(Ok("".to_string()), |x| Ok(format!(
                    " ON {}",
                    x.pretty_print()?
                )))?,
        ))
    }
}

impl PrettyPrint for JoinType {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            JoinType::Left => "LEFT",
            JoinType::Right => "RIGHT",
            JoinType::Cross => "CROSS",
            JoinType::Inner => "INNER",
        }
        .to_string())
    }
}

impl PrettyPrint for GroupByClause {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            " GROUP BY {}{}",
            self.keys
                .iter()
                .map(|x| x.pretty_print())
                .collect::<Result<Vec<_>>>()?
                .join(", "),
            if self.aggregations.is_empty() {
                "".to_string()
            } else {
                format!(
                    " AGGREGATE {}",
                    self.aggregations
                        .iter()
                        .map(|y| y.pretty_print())
                        .collect::<Result<Vec<_>>>()?
                        .join(", ")
                )
            }
        ))
    }
}

impl PrettyPrint for OrderByClause {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            " ORDER BY {}",
            self.sort_specs
                .iter()
                .map(|x| x.pretty_print())
                .collect::<Result<Vec<_>>>()?
                .join(", "),
        ))
    }
}

impl PrettyPrint for SortSpec {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "{}{}",
            self.key.pretty_print()?,
            self.direction.pretty_print()?
        ))
    }
}

impl PrettyPrint for SortKey {
    fn pretty_print(&self) -> Result<String> {
        match self {
            SortKey::Simple(s) => s.pretty_print(),
            SortKey::Positional(u) => Ok(u.to_string()),
        }
    }
}

impl PrettyPrint for SortDirection {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            SortDirection::Asc => " ASC",
            SortDirection::Desc => " DESC",
        }
        .to_string())
    }
}

// The tier numbers here match the Expr tiers in the parser.
#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone, Copy)]
enum ExpressionTier {
    Tier1,
    Tier2,
    Tier3,
    Tier4,
    Tier5,
    Tier6,
    Tier7,
    Tier8,
    Tier9,
    Tier10,
    Tier11,
    Tier12,
    Tier13,
    Bottom,
}

impl ExpressionTier {
    fn format_sub_expr(&self, sub_expr: &Expression) -> Result<String> {
        if *self >= sub_expr.get_tier() {
            Ok(format!("({})", sub_expr.pretty_print()?))
        } else {
            sub_expr.pretty_print()
        }
    }

    fn strict_format_sub_expr(&self, sub_expr: &Expression) -> Result<String> {
        if *self > sub_expr.get_tier() {
            Ok(format!("({})", sub_expr.pretty_print()?))
        } else {
            sub_expr.pretty_print()
        }
    }
}

impl BinaryOp {
    fn get_tier(&self) -> ExpressionTier {
        use BinaryOp::*;
        use ExpressionTier::*;
        match self {
            In | NotIn => Tier4,
            Or => Tier5,
            And => Tier6,
            Comparison(_) => Tier7,
            Concat => Tier8,
            Add | Sub => Tier9,
            Mul | Div => Tier10,
        }
    }
}

impl BinaryExpr {
    fn get_tier(&self) -> ExpressionTier {
        self.op.get_tier()
    }
}

impl UnaryExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::Tier11
    }
}

impl LikeExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::Tier1
    }
}

impl IsExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::Tier2
    }
}

impl BetweenExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::Tier3
    }
}

impl SubqueryComparisonExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::Tier7
    }
}

impl TypeAssertionExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::Tier12
    }
}

impl Expression {
    fn get_tier(&self) -> ExpressionTier {
        use Expression::*;
        use ExpressionTier::*;
        match self {
            Like(l) => l.get_tier(),
            Is(i) => i.get_tier(),
            Between(b) => b.get_tier(),
            SubqueryComparison(s) => s.get_tier(),
            Binary(b) => b.get_tier(),
            Unary(u) => u.get_tier(),
            TypeAssertion(t) => t.get_tier(),
            Subpath(s) => s.get_tier(),
            Access(a) => a.get_tier(),
            // formatting for all the following is handled specially and will never conditionally
            // wrap arguments in parentheses
            Array(_) | Case(_) | Cast(_) | Document(_) | Exists(_) | Function(_) | Trim(_)
            | Extract(_) | Identifier(_) | Literal(_) | Subquery(_) | Tuple(_) => Bottom,
        }
    }
}

impl SubpathExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::Tier13
    }
}

impl AccessExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::Bottom
    }
}

impl PrettyPrint for Expression {
    fn pretty_print(&self) -> Result<String> {
        use Expression::*;
        match self {
            Identifier(s) => Ok(identifier_to_string(s)),
            Is(i) => i.pretty_print(),
            Like(l) => l.pretty_print(),
            TypeAssertion(t) => t.pretty_print(),
            Cast(c) => c.pretty_print(),
            Literal(l) => l.pretty_print(),
            Unary(u) => u.pretty_print(),
            Binary(b) => b.pretty_print(),
            Extract(e) => e.pretty_print(),
            Trim(t) => t.pretty_print(),
            Function(fun) => fun.pretty_print(),
            Access(a) => a.pretty_print(),
            Subpath(sp) => sp.pretty_print(),
            Array(a) => Ok(format!(
                "[{}]",
                a.iter()
                    .map(|x| x.pretty_print())
                    .collect::<Result<Vec<_>>>()?
                    .join(", ")
            )),
            Tuple(t) => Ok(format!(
                "({})",
                t.iter()
                    .map(|x| x.pretty_print())
                    .collect::<Result<Vec<_>>>()?
                    .join(", ")
            )),
            Document(d) => Ok(format!(
                "{{{}}}",
                d.iter()
                    .map(|kv| Ok(format!(
                        "'{}': {}",
                        escape_string_literal(&kv.key),
                        kv.value.pretty_print()?
                    )))
                    .collect::<Result<Vec<_>>>()?
                    .join(", ")
            )),
            Between(b) => b.pretty_print(),
            Case(c) => c.pretty_print(),
            Subquery(q) => Ok(format!("({})", q.pretty_print()?)),
            Exists(q) => Ok(format!("EXISTS({})", q.pretty_print()?)),
            SubqueryComparison(sc) => sc.pretty_print(),
        }
    }
}

impl PrettyPrint for IsExpr {
    fn pretty_print(&self) -> Result<String> {
        let formatted_expr = self.get_tier().format_sub_expr(&self.expr)?;
        match self.target_type {
            TypeOrMissing::Missing => Ok(format!("{} IS MISSING", formatted_expr)),
            TypeOrMissing::Type(t) => Ok(format!("{} IS {}", formatted_expr, t.pretty_print()?)),
            TypeOrMissing::Number => Ok(format!("{} IS NUMBER", formatted_expr)),
        }
    }
}

impl PrettyPrint for LikeExpr {
    fn pretty_print(&self) -> Result<String> {
        let (formatted_expr, formatted_pattern) = (
            self.get_tier().strict_format_sub_expr(&self.expr)?,
            self.get_tier().format_sub_expr(&self.pattern)?,
        );
        Ok(format!(
            "{} LIKE {}{}",
            formatted_expr,
            formatted_pattern,
            self.escape.as_ref().map_or("".to_string(), |x| format!(
                " ESCAPE '{}'",
                escape_string_literal(x)
            ))
        ))
    }
}

impl PrettyPrint for TypeAssertionExpr {
    fn pretty_print(&self) -> Result<String> {
        let formatted_expr = self.get_tier().format_sub_expr(&self.expr)?;
        Ok(format!(
            "{}::!{}",
            formatted_expr,
            self.target_type.pretty_print()?
        ))
    }
}

impl PrettyPrint for CastExpr {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "CAST({} AS {}{}{})",
            self.expr.pretty_print()?,
            self.to.pretty_print()?,
            self.on_null
                .as_ref()
                .map_or(Ok("".to_string()), |x| Ok(format!(
                    ", {} ON NULL",
                    x.pretty_print()?
                )))?,
            self.on_error
                .as_ref()
                .map_or(Ok("".to_string()), |x| Ok(format!(
                    ", {} ON ERROR",
                    x.pretty_print()?
                )))?
        ))
    }
}

impl PrettyPrint for AccessExpr {
    fn pretty_print(&self) -> Result<String> {
        let formatted_expr = self.get_tier().strict_format_sub_expr(&*self.expr)?;
        Ok(format!(
            "{}[{}]",
            formatted_expr,
            self.subfield.pretty_print()?
        ))
    }
}

impl PrettyPrint for SubpathExpr {
    fn pretty_print(&self) -> Result<String> {
        let subpath = identifier_to_string(&self.subpath);
        let formatted_expr = self.get_tier().strict_format_sub_expr(&*self.expr)?;
        Ok(format!("{}.{}", formatted_expr, subpath))
    }
}

impl PrettyPrint for FunctionExpr {
    fn pretty_print(&self) -> Result<String> {
        if self.function == FunctionName::CurrentTimestamp && self.args.is_empty() {
            return Ok("current_timestamp".to_string());
        }
        match self.function {
            FunctionName::Position => pretty_print_position(&self.args),
            _ => match self.set_quantifier {
                Some(SetQuantifier::Distinct) => Ok(format!(
                    "{}(DISTINCT {})",
                    self.function.pretty_print()?,
                    self.args.pretty_print()?
                )),
                Some(SetQuantifier::All) => Ok(format!(
                    "{}(ALL {})",
                    self.function.pretty_print()?,
                    self.args.pretty_print()?
                )),
                None => Ok(format!(
                    "{}({})",
                    self.function.pretty_print()?,
                    self.args.pretty_print()?
                )),
            },
        }
    }
}

fn pretty_print_position(args: &FunctionArguments) -> Result<String> {
    match args {
        FunctionArguments::Star => unreachable!(),
        FunctionArguments::Args(args) => {
            assert!(args.len() == 2);
            let tier = BinaryOp::In.get_tier();
            // This assumes the In operator is left associative, which is currently true.  If it
            // were right associative we would need to use strict_format_sub_expr on the right
            // argument.
            let (formatted_left, formatted_right) = (
                tier.strict_format_sub_expr(&args[0])?,
                tier.format_sub_expr(&args[1])?,
            );
            Ok(format!(
                "POSITION({} IN {})",
                formatted_left, formatted_right
            ))
        }
    }
}

impl PrettyPrint for FunctionName {
    fn pretty_print(&self) -> Result<String> {
        Ok(self.as_str().to_string())
    }
}

impl PrettyPrint for FunctionArguments {
    fn pretty_print(&self) -> Result<String> {
        match self {
            FunctionArguments::Star => Ok("*".to_string()),
            FunctionArguments::Args(ve) => Ok(ve
                .iter()
                .map(|e| e.pretty_print())
                .collect::<Result<Vec<_>>>()?
                .join(", ")),
        }
    }
}

impl PrettyPrint for SubqueryComparisonExpr {
    fn pretty_print(&self) -> Result<String> {
        let formatted_expr = self.get_tier().format_sub_expr(&self.expr)?;
        Ok(format!(
            "{} {} {}({})",
            formatted_expr,
            self.op.pretty_print()?,
            self.quantifier.pretty_print()?,
            self.subquery.pretty_print()?
        ))
    }
}

impl PrettyPrint for SubqueryQuantifier {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            SubqueryQuantifier::All => "ALL",
            SubqueryQuantifier::Any => "ANY",
        }
        .to_string())
    }
}

impl PrettyPrint for Type {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            Type::Array => "ARRAY",
            Type::BinData => "BINDATA",
            Type::Boolean => "BOOLEAN",
            Type::Datetime => "TIMESTAMP",
            Type::DbPointer => "DBPOINTER",
            Type::Decimal128 => "DECIMAL",
            Type::Document => "DOCUMENT",
            Type::Double => "DOUBLE",
            Type::Int32 => "INT",
            Type::Int64 => "LONG",
            Type::Javascript => "JAVASCRIPT",
            Type::JavascriptWithScope => "JAVASCRIPTWITHSCOPE",
            Type::MaxKey => "MAXKEY",
            Type::MinKey => "MINKEY",
            Type::Null => "NULL",
            Type::ObjectId => "OBJECTID",
            Type::RegularExpression => "REGEX",
            Type::String => "STRING",
            Type::Symbol => "SYMBOL",
            Type::Timestamp => "BSON_TIMESTAMP",
            Type::Undefined => "UNDEFINED",
        }
        .to_string())
    }
}

impl PrettyPrint for ExtractSpec {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            ExtractSpec::Year => "YEAR",
            ExtractSpec::Month => "MONTH",
            ExtractSpec::Day => "DAY",
            ExtractSpec::Hour => "HOUR",
            ExtractSpec::Minute => "MINUTE",
            ExtractSpec::Second => "SECOND",
        }
        .to_string())
    }
}

impl PrettyPrint for TrimSpec {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            TrimSpec::Leading => "LEADING",
            TrimSpec::Trailing => "TRAILING",
            TrimSpec::Both => "BOTH",
        }
        .to_string())
    }
}

impl PrettyPrint for BetweenExpr {
    fn pretty_print(&self) -> Result<String> {
        let between_tier = self.get_tier();
        let sub_tier = BinaryOp::And.get_tier();
        let (formatted_expr, formatted_min, formatted_max) = (
            between_tier.format_sub_expr(&*self.expr)?,
            sub_tier.format_sub_expr(&*self.min)?,
            sub_tier.format_sub_expr(&*self.max)?,
        );
        Ok(format!(
            "{} BETWEEN {} AND {}",
            formatted_expr, formatted_min, formatted_max
        ))
    }
}

impl PrettyPrint for CaseExpr {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "CASE {}{}{} END",
            self.expr
                .as_ref()
                .map_or(Ok("".to_string()), |x| Ok(format!(
                    "{} ",
                    x.pretty_print()?
                )))?,
            self.when_branch
                .iter()
                .map(|x| x.pretty_print())
                .collect::<Result<Vec<_>>>()?
                .join(" "),
            self.else_branch
                .as_ref()
                .map_or(Ok("".to_string()), |x| Ok(format!(
                    " ELSE {}",
                    x.pretty_print()?
                )))?,
        ))
    }
}

impl PrettyPrint for WhenBranch {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "WHEN {} THEN {}",
            self.when.pretty_print()?,
            self.then.pretty_print()?
        ))
    }
}

fn ident_needs_delimiters(s: &str) -> bool {
    if s.is_empty() {
        return true;
    }
    let mut char_iter = s.chars();
    let first = char_iter.next().unwrap();
    if !(first.is_ascii_alphabetic() | (first == '_')) {
        return true;
    }
    for c in char_iter {
        if !(c.is_ascii_alphanumeric() | (c == '_')) {
            return true;
        }
    }
    if KEYWORDS.is_match(s) {
        return true;
    }
    false
}

impl PrettyPrint for Literal {
    fn pretty_print(&self) -> Result<String> {
        match self {
            Literal::Null => Ok("NULL".to_string()),
            Literal::Boolean(b) => Ok(b.to_string()),
            Literal::String(s) => Ok(format!("'{}'", escape_string_literal(s))),
            Literal::Integer(i) => Ok(i.to_string()),
            Literal::Long(l) => Ok(l.to_string()),
            Literal::Double(d) => {
                let d = d.to_string();
                Ok(if !d.contains('.') {
                    format!("{}.0", d)
                } else {
                    d
                })
            }
        }
    }
}

impl PrettyPrint for UnaryExpr {
    fn pretty_print(&self) -> Result<String> {
        let formatted_expr = self.get_tier().strict_format_sub_expr(&*self.expr)?;
        Ok(format!("{}{}", self.op.pretty_print()?, formatted_expr))
    }
}

impl PrettyPrint for UnaryOp {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            UnaryOp::Pos => "+ ",
            UnaryOp::Neg => "- ",
            UnaryOp::Not => "NOT ",
        }
        .to_string())
    }
}

impl PrettyPrint for ExtractExpr {
    fn pretty_print(&self) -> Result<String> {
        Ok(format!(
            "EXTRACT({} FROM {})",
            self.extract_spec.pretty_print()?,
            self.arg.pretty_print()?
        ))
    }
}

impl PrettyPrint for TrimExpr {
    fn pretty_print(&self) -> Result<String> {
        let trim_spec = match self.trim_spec {
            TrimSpec::Both => "BOTH",
            TrimSpec::Leading => "LEADING",
            TrimSpec::Trailing => "TRAILING",
        };
        match *self.trim_chars {
            Expression::Literal(Literal::String(ref s)) if s.as_str() == " " => Ok(format!(
                "TRIM({} FROM {})",
                trim_spec,
                self.arg.pretty_print()?
            )),
            _ => Ok(format!(
                "TRIM({} {} FROM {})",
                trim_spec,
                self.trim_chars.pretty_print()?,
                self.arg.pretty_print()?
            )),
        }
    }
}

impl PrettyPrint for BinaryExpr {
    fn pretty_print(&self) -> Result<String> {
        let tier = self.get_tier();
        // This assumes all binary operators are left associative, which is currently true.
        // Right associative operators would need to use strict_format_sub_expr on the right
        // argument.
        let (formatted_left, formatted_right) = (
            tier.strict_format_sub_expr(&*self.left)?,
            tier.format_sub_expr(&*self.right)?,
        );
        Ok(format!(
            "{} {} {}",
            formatted_left,
            self.op.pretty_print()?,
            formatted_right
        ))
    }
}

impl PrettyPrint for BinaryOp {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            BinaryOp::Or => "OR".to_string(),
            BinaryOp::And => "AND".to_string(),
            BinaryOp::Comparison(c) => c.pretty_print()?,
            BinaryOp::Concat => "||".to_string(),
            BinaryOp::Add => "+".to_string(),
            BinaryOp::Sub => "-".to_string(),
            BinaryOp::Mul => "*".to_string(),
            BinaryOp::Div => "/".to_string(),
            BinaryOp::In => "IN".to_string(),
            BinaryOp::NotIn => "NOT IN".to_string(),
        })
    }
}

impl PrettyPrint for ComparisonOp {
    fn pretty_print(&self) -> Result<String> {
        Ok(match self {
            ComparisonOp::Lt => "<",
            ComparisonOp::Lte => "<=",
            ComparisonOp::Gte => ">=",
            ComparisonOp::Gt => ">",
            ComparisonOp::Eq => "=",
            ComparisonOp::Neq => "<>",
        }
        .to_string())
    }
}
