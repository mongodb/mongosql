use std::fmt::{Display, Formatter, Result};

use crate::ast::*;

impl Display for Query {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self {
            Query::Select(q) => write!(f, "{}", q),
            Query::Set(q) => write!(f, "{}", q),
        }
    }
}

impl Display for SetQuery {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(f, "{} {} {}", self.left, self.op, self.right)
    }
}

impl Display for SetOperator {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{}",
            match self {
                SetOperator::Union => "UNION",
                SetOperator::UnionAll => "UNION ALL",
            }
        )
    }
}

impl Display for SelectQuery {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{}{}{}{}{}{}{}{}",
            self.select_clause,
            self.from_clause
                .as_ref()
                .map_or("".to_string(), |x| format!(" FROM {}", x)),
            self.where_clause
                .as_ref()
                .map_or("".to_string(), |x| format!(" WHERE {}", x)),
            self.group_by_clause
                .as_ref()
                .map_or("".to_string(), |x| x.to_string()),
            self.having_clause
                .as_ref()
                .map_or("".to_string(), |x| format!(" HAVING {}", x)),
            self.order_by_clause
                .as_ref()
                .map_or("".to_string(), |x| x.to_string()),
            self.limit
                .map_or("".to_string(), |x| format!(" LIMIT {}", x)),
            self.offset
                .map_or("".to_string(), |x| format!(" OFFSET {}", x)),
        )
    }
}

impl Display for SelectClause {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(f, "SELECT{}{}", self.set_quantifier, self.body)
    }
}

impl Display for SetQuantifier {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{}",
            match self {
                SetQuantifier::All => "",
                SetQuantifier::Distinct => " DISTINCT",
            }
        )
    }
}

impl Display for SelectBody {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self {
            SelectBody::Standard(v) => write!(
                f,
                " {}",
                v.iter()
                    .map(|x| format!("{}", x))
                    .collect::<Vec<_>>()
                    .join(", ")
            ),
            SelectBody::Values(v) => write!(
                f,
                " VALUES {}",
                v.iter()
                    .map(|x| format!("{}", x))
                    .collect::<Vec<_>>()
                    .join(", ")
            ),
        }
    }
}

impl Display for SelectExpression {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self {
            SelectExpression::Star => write!(f, "*"),
            SelectExpression::Substar(s) => write!(f, "{}", s),
            SelectExpression::Aliased(ae) => write!(f, "{}", ae),
        }
    }
}

impl Display for SubstarExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(f, "{}.*", self.datasource)
    }
}

impl Display for AliasedExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{}{}",
            self.expr,
            self.alias
                .as_ref()
                .map_or("".to_string(), |x| format!(" AS {}", x))
        )
    }
}

impl Display for SelectValuesExpression {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self {
            SelectValuesExpression::Substar(s) => write!(f, "{}", s),
            SelectValuesExpression::Expression(e) => write!(f, "{}", e),
        }
    }
}

impl Display for Datasource {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self {
            Datasource::Array(a) => write!(f, "{}", a),
            Datasource::Collection(c) => write!(f, "{}", c),
            Datasource::Derived(d) => write!(f, "{}", d),
            Datasource::Join(j) => write!(f, "{}", j),
        }
    }
}

impl Display for ArraySource {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "[{}] AS {}",
            self.array
                .iter()
                .map(|x| format!("{}", x))
                .collect::<Vec<_>>()
                .join(", "),
            self.alias,
        )
    }
}

impl Display for CollectionSource {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{}{}{}",
            self.database
                .as_ref()
                .map_or("".to_string(), |x| format!("{}.", x)),
            self.collection,
            self.alias
                .as_ref()
                .map_or("".to_string(), |x| format!(" AS {}", x)),
        )
    }
}

impl Display for DerivedSource {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(f, "({}) AS {}", self.query, self.alias)
    }
}

impl Display for JoinSource {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{} {} JOIN {}{}",
            self.left,
            self.join_type,
            self.right,
            self.condition
                .as_ref()
                .map_or("".to_string(), |x| format!(" ON {}", x)),
        )
    }
}

impl Display for JoinType {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{}",
            match self {
                JoinType::Left => "LEFT",
                JoinType::Right => "RIGHT",
                JoinType::Cross => "CROSS",
                JoinType::Inner => "INNER",
            }
        )
    }
}

impl Display for GroupByClause {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            " GROUP BY {}{}",
            self.keys
                .iter()
                .map(|x| format!("{}", x))
                .collect::<Vec<_>>()
                .join(", "),
            if self.aggregations.is_empty() {
                "".to_string()
            } else {
                format!(
                    " AGGREGATE {}",
                    self.aggregations
                        .iter()
                        .map(|y| format!("{}", y))
                        .collect::<Vec<_>>()
                        .join(", ")
                )
            }
        )
    }
}

impl Display for OrderByClause {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            " ORDER BY {}",
            self.sort_specs
                .iter()
                .map(|x| format!("{}", x))
                .collect::<Vec<_>>()
                .join(", "),
        )
    }
}

impl Display for SortSpec {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(f, "{}{}", self.key, self.direction)
    }
}

impl Display for SortKey {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self {
            SortKey::Simple(s) => write!(f, "{}", s),
            SortKey::Positional(u) => write!(f, "{}", u),
        }
    }
}

impl Display for SortDirection {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{}",
            match self {
                SortDirection::Asc => "",
                SortDirection::Desc => " DESC",
            }
        )
    }
}

// The tier numbers here match the Expr tiers in the parser.
#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone, Copy)]
enum ExpressionTier {
    Tier1Expr,
    Tier2Expr,
    Tier3Expr,
    Tier4Expr,
    Tier5Expr,
    Tier6Expr,
    Tier7Expr,
    Tier8Expr,
    Tier9Expr,
    Tier10Expr,
    Tier11Expr,
    Tier12Expr,
    Tier13Expr,
    BottomExpr,
}

impl BinaryExpr {
    fn get_tier(&self) -> ExpressionTier {
        use BinaryOp::*;
        use ExpressionTier::*;
        match self.op {
            In | NotIn => Tier4Expr,
            Or => Tier5Expr,
            And => Tier6Expr,
            Lt | Lte | Gte | Gt | Eq | Neq => Tier7Expr,
            Concat => Tier8Expr,
            Add | Sub => Tier9Expr,
            Mul | Div => Tier10Expr,
        }
    }
}

impl UnaryExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::Tier11Expr
    }
}

impl Expression {
    fn get_tier(&self) -> ExpressionTier {
        use Expression::*;
        use ExpressionTier::*;
        match self {
            Like(_) => Tier1Expr,
            Is(_) => Tier2Expr,
            Between(_) => Tier3Expr,
            SubqueryComparison(_) => Tier7Expr,
            Binary(b) => b.get_tier(),
            Unary(u) => u.get_tier(),
            TypeAssertion(_) => Tier12Expr,
            Subpath(s) => s.get_tier(),
            Access(a) => a.get_tier(),
            Array(_) => BottomExpr,
            Case(_) => BottomExpr,
            Cast(_) => BottomExpr,
            Document(_) => BottomExpr,
            Exists(_) => BottomExpr,
            Function(_) => BottomExpr,
            Identifier(_) => BottomExpr,
            Literal(_) => BottomExpr,
            Subquery(_) => BottomExpr,
            Tuple(_) => BottomExpr,
        }
    }
}

impl SubpathExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::Tier13Expr
    }
}

impl AccessExpr {
    fn get_tier(&self) -> ExpressionTier {
        ExpressionTier::BottomExpr
    }
}

impl Display for Expression {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        use Expression::*;
        match self {
            Identifier(s) => {
                if has_special_chars(s) {
                    write!(f, "`{}`", s)
                } else {
                    write!(f, "{}", s)
                }
            }
            Is(i) => write!(f, "{}", i),
            Like(l) => write!(f, "{}", l),
            TypeAssertion(t) => write!(f, "{}", t),
            Cast(c) => write!(f, "{}", c),
            Literal(l) => write!(f, "{}", l),
            Unary(u) => write!(f, "{}", u),
            Binary(b) => write!(f, "{}", b),
            Function(fun) => write!(f, "{}", fun),
            Access(a) => write!(f, "{}", a),
            Subpath(sp) => write!(f, "{}", sp),
            Array(a) => write!(
                f,
                "[{}]",
                a.iter()
                    .map(|x| format!("{}", x))
                    .collect::<Vec<_>>()
                    .join(", ")
            ),
            Tuple(t) => write!(
                f,
                "({})",
                t.iter()
                    .map(|x| format!("{}", x))
                    .collect::<Vec<_>>()
                    .join(", ")
            ),
            Document(d) => write!(
                f,
                "{{{}}}",
                d.iter()
                    .map(|(k, v)| format!("'{}': {}", k, v))
                    .collect::<Vec<_>>()
                    .join(", ")
            ),
            Between(b) => write!(f, "{}", b),
            Case(c) => write!(f, "{}", c),
            Subquery(q) => write!(f, "({})", q),
            Exists(q) => write!(f, "EXISTS({})", q),
            SubqueryComparison(sc) => write!(f, "{}", sc),
        }
    }
}

impl Display for IsExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self.target_type {
            TypeOrMissing::Missing => write!(f, "{} IS MISSING", self.expr),
            TypeOrMissing::Type(t) => write!(f, "{} IS {}", self.expr, t),
        }
    }
}

impl Display for LikeExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{} LIKE {}{}",
            self.expr,
            self.pattern,
            self.escape
                .as_ref()
                .map_or("".to_string(), |x| format!(" ESCAPE '{}'", x))
        )
    }
}

impl Display for TypeAssertionExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(f, "{}::!{}", self.expr, self.target_type)
    }
}

impl Display for CastExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "CAST({} AS {}{}{})",
            self.expr,
            self.to,
            self.on_null
                .as_ref()
                .map_or("".to_string(), |x| format!(", {} ON NULL", x)),
            self.on_error
                .as_ref()
                .map_or("".to_string(), |x| format!(", {} ON ERROR", x))
        )
    }
}

impl Display for AccessExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        if self.expr.get_tier() < self.get_tier() {
            write!(f, "({})[{}]", self.expr, self.subfield)
        } else {
            write!(f, "{}[{}]", self.expr, self.subfield)
        }
    }
}

impl Display for SubpathExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        if self.expr.get_tier() < self.get_tier() {
            write!(f, "({}).{}", self.expr, self.subpath)
        } else {
            write!(f, "{}.{}", self.expr, self.subpath)
        }
    }
}

impl Display for FunctionExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self.function.0.as_ref() {
            "EXTRACT" => fmt_extract(f, &self.args),
            "POSITION" => fmt_position(f, &self.args),
            "TRIM" => fmt_trim(f, &self.args),
            _ => {
                let args = self
                    .args
                    .iter()
                    .map(|x| format!("{}", x))
                    .collect::<Vec<_>>()
                    .join(", ");
                write!(f, "{}({})", self.function.0, args)
            }
        }
    }
}

fn fmt_extract(f: &mut Formatter<'_>, args: &[FunctionArg]) -> Result {
    assert!(args.len() == 2);
    write!(f, "EXTRACT({} FROM {})", args[0], args[1])
}

fn fmt_position(f: &mut Formatter<'_>, args: &[FunctionArg]) -> Result {
    assert!(args.len() == 2);
    write!(f, "POSITION({} IN {})", args[0], args[1])
}

fn fmt_trim(f: &mut Formatter<'_>, args: &[FunctionArg]) -> Result {
    assert!(args.len() == 3);
    let trim_spec = match args[0] {
        FunctionArg::Trim(TrimSpec::Both) => "BOTH",
        FunctionArg::Trim(TrimSpec::Leading) => "LEADING",
        FunctionArg::Trim(TrimSpec::Trailing) => "TRAILING",
        _ => unreachable!(),
    };
    if let FunctionArg::Expr(Expression::Identifier(s)) = &args[1] {
        if s == " " {
            write!(f, "TRIM({} FROM {})", trim_spec, args[2])
        } else {
            write!(f, "TRIM({} {} FROM {})", trim_spec, s, args[2])
        }
    } else {
        write!(f, "TRIM({} {} FROM {})", trim_spec, args[1], args[2])
    }
}

impl Display for FunctionArg {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self {
            FunctionArg::Star => write!(f, "*"),
            FunctionArg::Expr(e) => write!(f, "{}", e),
            FunctionArg::Extract(ext) => write!(f, "{}", ext),
            FunctionArg::Trim(t) => write!(f, "{}", t),
        }
    }
}

impl Display for SubqueryComparisonExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{} {} {}({})",
            self.expr, self.op, self.quantifier, self.subquery
        )
    }
}

impl Display for SubqueryQuantifier {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self {
            SubqueryQuantifier::All => write!(f, "ALL"),
            SubqueryQuantifier::Any => write!(f, "ANY"),
        }
    }
}

impl Display for Type {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{}",
            match self {
                Type::Array => "ARRAY",
                Type::BinData => "BINDATA",
                Type::Boolean => "BOOLEAN",
                Type::Datetime => "BSON_DATE",
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
                Type::Timestamp => "TIMESTAMP",
                Type::Undefined => "UNDEFINED",
            }
        )
    }
}

impl Display for ExtractSpec {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        let out = match self {
            ExtractSpec::TimezoneHour => "TIMEZONE_HOUR",
            ExtractSpec::TimezoneMinute => "TIMEZONE_MINUTE",
            ExtractSpec::Year => "YEAR",
            ExtractSpec::Month => "MONTH",
            ExtractSpec::Day => "DAY",
            ExtractSpec::Hour => "HOUR",
            ExtractSpec::Minute => "MINUTE",
            ExtractSpec::Second => "SECOND",
        };
        write!(f, "{}", out)
    }
}

impl Display for TrimSpec {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self {
            TrimSpec::Leading => write!(f, "LEADING"),
            TrimSpec::Trailing => write!(f, "TRAILING"),
            TrimSpec::Both => write!(f, "BOTH"),
        }
    }
}

impl Display for BetweenExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(f, "{} BETWEEN {} AND {}", self.expr, self.min, self.max)
    }
}

impl Display for CaseExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "CASE {}{}{} END",
            self.expr
                .as_ref()
                .map_or("".to_string(), |x| format!("{} ", x)),
            self.when_branch
                .iter()
                .map(|x| format!("{}", x))
                .collect::<Vec<_>>()
                .join(" "),
            self.else_branch
                .as_ref()
                .map_or("".to_string(), |x| format!(" ELSE {}", x)),
        )
    }
}

impl Display for WhenBranch {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(f, "WHEN {} THEN {}", self.when, self.then)
    }
}

fn has_special_chars(s: &str) -> bool {
    let search = "$.";
    for c in s.chars() {
        if search.contains(c) {
            return true;
        }
    }
    false
}

impl Display for Literal {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        match self {
            Literal::Null => write!(f, "NULL"),
            Literal::Boolean(b) => write!(f, "{}", b),
            Literal::String(s) => write!(f, "'{}'", s),
            Literal::Integer(i) => write!(f, "{}", i),
            Literal::Long(l) => write!(f, "{}", l),
            Literal::Double(d) => {
                let d = d.to_string();
                if !d.contains('.') {
                    write!(f, "{}.0", d)
                } else {
                    write!(f, "{}", d)
                }
            }
        }
    }
}

impl Display for UnaryExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        if self.op == UnaryOp::Pos {
            return write!(f, "{}", self.expr);
        }
        if self.expr.get_tier() < self.get_tier() {
            write!(f, "{}({})", self.op, self.expr)
        } else {
            write!(f, "{}{}", self.op, self.expr)
        }
    }
}

impl Display for UnaryOp {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{}",
            match self {
                UnaryOp::Pos => "",
                UnaryOp::Neg => "-",
                UnaryOp::Not => "NOT ",
            }
        )
    }
}

impl Display for BinaryExpr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        let paren_format = |x: &Expression| {
            if self.get_tier() > x.get_tier() {
                format!("({})", x)
            } else {
                format!("{}", x)
            }
        };
        write!(
            f,
            "{} {} {}",
            paren_format(self.left.as_ref()),
            self.op,
            paren_format(self.right.as_ref()),
        )
    }
}

impl Display for BinaryOp {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result {
        write!(
            f,
            "{}",
            match self {
                BinaryOp::Or => "OR",
                BinaryOp::And => "AND",
                BinaryOp::Lt => "<",
                BinaryOp::Lte => "<=",
                BinaryOp::Gte => ">=",
                BinaryOp::Gt => ">",
                BinaryOp::Eq => "=",
                BinaryOp::Neq => "<>",
                BinaryOp::Concat => "||",
                BinaryOp::Add => "+",
                BinaryOp::Sub => "-",
                BinaryOp::Mul => "*",
                BinaryOp::Div => "/",
                BinaryOp::In => "IN",
                BinaryOp::NotIn => "NOT IN",
            }
        )
    }
}
