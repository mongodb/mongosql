#[cfg(test)]

mod fuzz_test {
    use crate::{
        ast::{
            definitions::*,
            rewrites::{Pass, SingleTupleRewritePass},
        },
        parser::Parser,
    };
    use quickcheck::*;
    use std::io::Write;

    #[test]
    // For arbitrary Query q, this test asserts the property:
    //
    //   q == Parse(PrettyPrint(q))
    //
    // As in, pretty printing a query and then reparsing that
    // pretty string results in the original query.
    fn prop_pretty_print_parse_is_idempotent() {
        fn pretty_print_query(q: Query) -> TestResult {
            // Display is fallible, but the format! macro does not propagate the error -- it
            // simply panics.  So we first use the write! macro to see if the pretty printer
            // fails. If it does, we discard this input.  Otherwise we proceed to assert the
            // property.
            let mut w = Vec::new();
            let res = write!(&mut w, "{}", q);
            if res.is_err() {
                return TestResult::discard();
            }

            // When we reparse after pretty printing,  some parenthesized  expressions are
            // interpreted as single-element tuples.  This is an intentional choice in the
            // parser on our part. The SingleTupleRewritePass unwraps these single-element
            // tuples into their single expressions.  We discard any  tests that encounter
            // rewrite errors,  though the  SingleTupleRewritePass  should never return an
            // error anyway.
            let p = format!("{}", q);
            let reparsed = SingleTupleRewritePass.apply(Parser::new().parse_query(&p).unwrap());
            match reparsed {
                Err(_) => TestResult::discard(),
                Ok(r) => TestResult::from_bool(q == r),
            }
        }

        QuickCheck::new()
            .gen(Gen::new(0))
            .quickcheck(pretty_print_query as fn(Query) -> TestResult);
    }
}

mod arbitrary {
    use crate::ast::definitions::*;
    use quickcheck::{Arbitrary, Gen};
    use rand::{thread_rng, Rng};
    use std::convert::TryFrom;

    // For SELECT, GROUP BY, and ORDER BY clauses
    static MIN_CLAUSE_EXPRS: u32 = 1; // minimum number of expressions in an arbitrary clause
    static MAX_CLAUSE_EXPRS: u32 = 4; // maximum number of expressions in an arbitrary clause

    static MIN_COMPOSITE_DATA_LEN: u32 = 0; // minimum length of an arbitrary array or document
    static MAX_COMPOSITE_DATA_LEN: u32 = 4; // maximum length of an arbitrary array or document

    // Maximum nesting level of an arbitrary query or expression. This does not indicate
    // the actual depth of the query, but instead is used as a bound to stop generating
    // deeper arbitrary subqueries and subexpressions. The test initially starts with a
    // Gen of size 0. Each time a subquery or subexpression needs to be generated, the
    // Arbitrary implementations below create a new Gen with arbitrary size between
    // NEST_LOWER_BOUND and NEST_UPPER_BOUND. Eventually, this will reach or exceed the
    // MAX_NEST and prevent further nesting. When MAX_NEST is met or exceeded, nesting
    // ceases and a terminal expression (or subquery) is returned.
    static MAX_NEST: usize = 50;
    static NEST_LOWER_BOUND: usize = 10; // At most 5 levels of nesting
    static NEST_UPPER_BOUND: usize = 25; // At least 2 levels of nesting

    /// Return an arbitrary String without null characters.
    ///
    /// These Strings can be used for aliases, identifiers, or literals.
    fn arbitrary_string() -> String {
        let g = &mut Gen::new(rand_len(1, 20) as usize);
        String::arbitrary(g).replace("\u{0}", "")
    }

    /// Return an arbitrary Option<String> without null characters.
    fn arbitrary_optional_string(g: &mut Gen) -> Option<String> {
        if bool::arbitrary(g) {
            None
        } else {
            Some(arbitrary_string())
        }
    }

    fn arbitrary_expr_terminal(g: &mut Gen) -> Expression {
        if bool::arbitrary(g) {
            Expression::Literal(Literal::arbitrary(g))
        } else {
            Expression::Identifier(arbitrary_string())
        }
    }

    fn rand_len(low: u32, high: u32) -> u32 {
        let mut rng = thread_rng();
        rng.gen_range(low..high)
    }

    fn nested_gen(g: &mut Gen) -> Gen {
        let nested_size = g.size() + thread_rng().gen_range(NEST_LOWER_BOUND..NEST_UPPER_BOUND);
        Gen::new(nested_size)
    }

    impl Arbitrary for Query {
        fn arbitrary(g: &mut Gen) -> Self {
            if g.size() >= MAX_NEST {
                return Self::Select(SelectQuery::arbitrary(g));
            }

            let nested_g = &mut nested_gen(g);
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Select(SelectQuery::arbitrary(nested_g)),
                1 => Self::Set(SetQuery::arbitrary(nested_g)),
                _ => panic!("missing Query variant(s)"),
            }
        }
    }

    impl Arbitrary for SelectQuery {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                select_clause: SelectClause::arbitrary(g),
                from_clause: Option::arbitrary(g),
                where_clause: Option::arbitrary(g),
                group_by_clause: Option::arbitrary(g),
                having_clause: Option::arbitrary(g),
                order_by_clause: Option::arbitrary(g),
                limit: Option::arbitrary(g),
                offset: Option::arbitrary(g),
            }
        }
    }

    impl Arbitrary for SetQuery {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                left: Box::new(Query::arbitrary(g)),
                op: SetOperator::arbitrary(g),
                right: Box::new(Query::arbitrary(g)),
            }
        }
    }

    impl Arbitrary for SetOperator {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Union,
                1 => Self::UnionAll,
                _ => panic!("missing SetOperator variant(s)"),
            }
        }
    }

    impl Arbitrary for SelectClause {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                set_quantifier: SetQuantifier::arbitrary(g),
                body: SelectBody::arbitrary(g),
            }
        }
    }

    impl Arbitrary for SetQuantifier {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::All,
                1 => Self::Distinct,
                _ => panic!("missing SetQuantifier variant(s)"),
            }
        }
    }

    impl Arbitrary for SelectBody {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Standard(
                    (0..rand_len(MIN_CLAUSE_EXPRS, MAX_CLAUSE_EXPRS))
                        .map(|_| SelectExpression::arbitrary(g))
                        .collect(),
                ),
                1 => Self::Values(
                    (0..rand_len(MIN_CLAUSE_EXPRS, MAX_CLAUSE_EXPRS))
                        .map(|_| SelectValuesExpression::arbitrary(g))
                        .collect(),
                ),
                _ => panic!("missing SelectBody variant(s)"),
            }
        }
    }

    impl Arbitrary for SelectValuesExpression {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Expression(Expression::arbitrary(g)),
                1 => Self::Substar(SubstarExpr::arbitrary(g)),
                _ => panic!("missing SelectValuesExpression variant(s)"),
            }
        }
    }

    impl Arbitrary for SelectExpression {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Star,
                1 => Self::Substar(SubstarExpr::arbitrary(g)),
                2 => Self::Expression(OptionallyAliasedExpr::arbitrary(g)),
                _ => panic!("missing SelectExpression variant(s)"),
            }
        }
    }

    impl Arbitrary for SubstarExpr {
        fn arbitrary(_g: &mut Gen) -> Self {
            Self {
                datasource: arbitrary_string(),
            }
        }
    }

    impl Arbitrary for Datasource {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Array(ArraySource::arbitrary(g)),
                1 => Self::Collection(CollectionSource::arbitrary(g)),
                2 => Self::Derived(DerivedSource::arbitrary(g)),
                3 => Self::Join(JoinSource::arbitrary(g)),
                _ => panic!("missing Datasource variant(s)"),
            }
        }
    }

    impl Arbitrary for ArraySource {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                array: (0..rand_len(MIN_COMPOSITE_DATA_LEN, MAX_COMPOSITE_DATA_LEN))
                    .map(|_| Expression::arbitrary(g))
                    .collect(),
                alias: arbitrary_string(),
            }
        }
    }

    impl Arbitrary for CollectionSource {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                database: arbitrary_optional_string(g),
                collection: arbitrary_string(),
                alias: arbitrary_optional_string(g),
            }
        }
    }

    impl Arbitrary for DerivedSource {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                query: Box::new(Query::arbitrary(g)),
                alias: arbitrary_string(),
            }
        }
    }

    impl Arbitrary for OptionallyAliasedExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Aliased(AliasedExpr::arbitrary(g)),
                1 => Self::Unaliased(Expression::arbitrary(g)),
                _ => panic!("missing OptionallyAliasedExpr variant(s)"),
            }
        }
    }

    impl Arbitrary for AliasedExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                expr: Expression::arbitrary(g),
                alias: arbitrary_string(),
            }
        }
    }

    impl Arbitrary for JoinSource {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                join_type: JoinType::arbitrary(g),
                left: Box::new(Datasource::arbitrary(g)),
                right: Box::new(Datasource::arbitrary(g)),
                condition: Option::arbitrary(g),
            }
        }
    }

    impl Arbitrary for JoinType {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Left,
                1 => Self::Right,
                2 => Self::Cross,
                3 => Self::Inner,
                _ => panic!("missing JoinType variant(s)"),
            }
        }
    }

    impl Arbitrary for Expression {
        fn arbitrary(g: &mut Gen) -> Self {
            if g.size() >= MAX_NEST {
                return arbitrary_expr_terminal(g);
            }

            let nested_g = &mut nested_gen(g);
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Binary(BinaryExpr::arbitrary(nested_g)),
                1 => Self::Unary(UnaryExpr::arbitrary(nested_g)),
                2 => Self::Between(BetweenExpr::arbitrary(nested_g)),
                3 => Self::Case(CaseExpr::arbitrary(nested_g)),
                4 => Self::Function(FunctionExpr::arbitrary(nested_g)),
                5 => Self::Trim(TrimExpr::arbitrary(nested_g)),
                6 => Self::Extract(ExtractExpr::arbitrary(nested_g)),
                7 => Self::Cast(CastExpr::arbitrary(nested_g)),
                8 => Self::Array(
                    (0..rand_len(MIN_COMPOSITE_DATA_LEN, MAX_COMPOSITE_DATA_LEN))
                        .map(|_| Self::arbitrary(nested_g))
                        .collect(),
                ),
                9 => Self::Subquery(Box::new(Query::arbitrary(nested_g))),
                10 => Self::Exists(Box::new(Query::arbitrary(nested_g))),
                11 => Self::SubqueryComparison(SubqueryComparisonExpr::arbitrary(nested_g)),
                12 => Self::Document(
                    (0..rand_len(MIN_COMPOSITE_DATA_LEN, MAX_COMPOSITE_DATA_LEN))
                        .map(|_| DocumentPair::arbitrary(nested_g))
                        .collect(),
                ),
                13 => Self::Access(AccessExpr::arbitrary(nested_g)),
                14 => Self::Subpath(SubpathExpr::arbitrary(nested_g)),
                15 => Self::Identifier(arbitrary_string()),
                16 => Self::Is(IsExpr::arbitrary(nested_g)),
                17 => Self::Like(LikeExpr::arbitrary(nested_g)),
                18 => Self::Literal(Literal::arbitrary(nested_g)),
                19 => Self::Tuple((1..4).map(|_| Self::arbitrary(nested_g)).collect()),
                20 => Self::TypeAssertion(TypeAssertionExpr::arbitrary(nested_g)),
                _ => panic!("missing Expression variant(s)"),
            }
        }
    }

    impl Arbitrary for DocumentPair {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                key: arbitrary_string(),
                value: Expression::arbitrary(g),
            }
        }
    }

    impl Arbitrary for CastExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                expr: Box::new(Expression::arbitrary(g)),
                to: Type::arbitrary(g),
                on_null: Option::arbitrary(g),
                on_error: Option::arbitrary(g),
            }
        }
    }

    impl Arbitrary for BinaryExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                left: Box::new(Expression::arbitrary(g)),
                op: BinaryOp::arbitrary(g),
                right: Box::new(Expression::arbitrary(g)),
            }
        }
    }

    impl Arbitrary for UnaryExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                op: UnaryOp::arbitrary(g),
                expr: Box::new(Expression::arbitrary(g)),
            }
        }
    }

    impl Arbitrary for BetweenExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                expr: Box::new(Expression::arbitrary(g)),
                min: Box::new(Expression::arbitrary(g)),
                max: Box::new(Expression::arbitrary(g)),
            }
        }
    }

    impl Arbitrary for CaseExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                expr: Option::arbitrary(g),
                when_branch: (1..rand_len(2, 4))
                    .map(|_| WhenBranch::arbitrary(g))
                    .collect(),
                else_branch: Option::arbitrary(g),
            }
        }
    }

    impl Arbitrary for WhenBranch {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                when: Box::new(Expression::arbitrary(g)),
                then: Box::new(Expression::arbitrary(g)),
            }
        }
    }

    impl Arbitrary for SubqueryQuantifier {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::All,
                1 => Self::Any,
                _ => panic!("missing SubqueryQuantifier variant(s)"),
            }
        }
    }

    impl Arbitrary for SubqueryComparisonExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                expr: Box::new(Expression::arbitrary(g)),
                op: BinaryOp::arbitrary(g),
                quantifier: SubqueryQuantifier::arbitrary(g),
                subquery: Box::new(Query::arbitrary(g)),
            }
        }
    }

    impl Arbitrary for FunctionExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            // CurrentTimestamp, Position, and Substring need to be special-cased
            // because they each have special syntactic constraints on their
            // argument lists.
            let rng = &(0..FunctionName::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                // CurrentTimestamp can syntactically accept 0 or 1 argument(s).
                0 => Self {
                    function: FunctionName::CurrentTimestamp,
                    args: FunctionArguments::Args(
                        (0..rand_len(0, 1))
                            .map(|_| Expression::arbitrary(g))
                            .collect(),
                    ),
                    set_quantifier: None,
                },

                // Position can syntactically accept exactly 2 arguments.
                1 => Self {
                    function: FunctionName::Position,
                    args: FunctionArguments::Args(vec![
                        Expression::arbitrary(g),
                        Expression::arbitrary(g),
                    ]),
                    set_quantifier: None,
                },

                // Substring can syntactically accept 2 or 3 arguments.
                2 => Self {
                    function: FunctionName::Substring,
                    args: FunctionArguments::Args(
                        (0..rand_len(2, 3))
                            .map(|_| Expression::arbitrary(g))
                            .collect(),
                    ),
                    set_quantifier: None,
                },

                // Anything else can syntactically accept any number of arguments
                // and have a set quantifier.
                _ => Self {
                    function: FunctionName::arbitrary(g),
                    args: FunctionArguments::arbitrary(g),
                    set_quantifier: Option::arbitrary(g),
                },
            }
        }
    }

    impl Arbitrary for ExtractExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                extract_spec: ExtractSpec::arbitrary(g),
                arg: Box::new(Expression::arbitrary(g)),
            }
        }
    }

    impl Arbitrary for TrimExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                trim_spec: TrimSpec::arbitrary(g),
                trim_chars: Box::new(Expression::arbitrary(g)),
                arg: Box::new(Expression::arbitrary(g)),
            }
        }
    }

    impl Arbitrary for FunctionName {
        fn arbitrary(g: &mut Gen) -> Self {
            // Intentionally omitting CurrentTimestamp, Position, and Substring
            // because they cannot be built as arbitrarily as the other functions.
            // They each have special constraints on their arguments, as noted in
            // FunctionExpr::arbitrary.
            let rng = &(0..(Self::VARIANT_COUNT - 3)).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::AddToArray,
                1 => Self::AddToSet,
                2 => Self::Avg,
                3 => Self::Count,
                4 => Self::First,
                5 => Self::Last,
                6 => Self::Max,
                7 => Self::MergeDocuments,
                8 => Self::Min,
                9 => Self::StddevPop,
                10 => Self::StddevSamp,
                11 => Self::Sum,
                12 => Self::BitLength,
                13 => Self::CharLength,
                14 => Self::Coalesce,
                15 => Self::Lower,
                16 => Self::NullIf,
                17 => Self::OctetLength,
                18 => Self::Size,
                19 => Self::Slice,
                20 => Self::Upper,
                _ => panic!("missing FunctionName variant(s)"),
            }
        }
    }

    impl Arbitrary for FunctionArguments {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..10).collect::<Vec<i32>>();
            // artificially prefer Args to Star 10-to-1
            match g.choose(rng).unwrap() {
                0 => Self::Star,
                _ => Self::Args(
                    (0..rand_len(MIN_COMPOSITE_DATA_LEN, MAX_COMPOSITE_DATA_LEN))
                        .map(|_| Expression::arbitrary(g))
                        .collect(),
                ),
            }
        }
    }

    impl Arbitrary for ExtractSpec {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Year,
                1 => Self::Month,
                2 => Self::Day,
                3 => Self::Hour,
                4 => Self::Minute,
                5 => Self::Second,
                _ => panic!("missing ExtractSpec variant(s)"),
            }
        }
    }

    impl Arbitrary for TrimSpec {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Leading,
                1 => Self::Trailing,
                2 => Self::Both,
                _ => panic!("missing TrimSpec variant(s)"),
            }
        }
    }

    impl Arbitrary for AccessExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                expr: Box::new(Expression::arbitrary(g)),
                subfield: Box::new(Expression::arbitrary(g)),
            }
        }
    }

    impl Arbitrary for SubpathExpr {
        // TODO: SQL-467: Do not special-case SubpathExpr expr
        //   - For now, it is special-cased to avoid the problem of
        //     the parser rejecting expressions like 1.a, for example
        fn arbitrary(_g: &mut Gen) -> Self {
            Self {
                expr: Box::new(Expression::Identifier(arbitrary_string())),
                subpath: arbitrary_string(),
            }
        }
    }

    impl Arbitrary for TypeOrMissing {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Type(Type::arbitrary(g)),
                1 => Self::Number,
                2 => Self::Missing,
                _ => panic!("missing TypeOrMissing variant(s)"),
            }
        }
    }

    impl Arbitrary for IsExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                expr: Box::new(Expression::arbitrary(g)),
                target_type: TypeOrMissing::arbitrary(g),
            }
        }
    }

    impl Arbitrary for LikeExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                expr: Box::new(Expression::arbitrary(g)),
                pattern: Box::new(Expression::arbitrary(g)),
                escape: arbitrary_optional_string(g),
            }
        }
    }

    impl Arbitrary for TypeAssertionExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                expr: Box::new(Expression::arbitrary(g)),
                target_type: Type::arbitrary(g),
            }
        }
    }

    impl Arbitrary for UnaryOp {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Pos,
                1 => Self::Neg,
                2 => Self::Not,
                _ => panic!("missing UnaryOp variant(s)"),
            }
        }
    }

    impl Arbitrary for BinaryOp {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Add,
                1 => Self::And,
                2 => Self::Concat,
                3 => Self::Div,
                4 => Self::Eq,
                5 => Self::Gt,
                6 => Self::Gte,
                7 => Self::In,
                8 => Self::Lt,
                9 => Self::Lte,
                10 => Self::Mul,
                11 => Self::Neq,
                12 => Self::NotIn,
                13 => Self::Or,
                14 => Self::Sub,
                _ => panic!("missing BinaryOp variant(s)"),
            }
        }
    }

    impl Arbitrary for GroupByClause {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                keys: (0..rand_len(MIN_CLAUSE_EXPRS, MAX_CLAUSE_EXPRS))
                    .map(|_| OptionallyAliasedExpr::arbitrary(g))
                    .collect(),
                aggregations: (0..rand_len(MIN_CLAUSE_EXPRS, MAX_CLAUSE_EXPRS))
                    .map(|_| AliasedExpr::arbitrary(g))
                    .collect(),
            }
        }
    }

    impl Arbitrary for OrderByClause {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                sort_specs: (0..rand_len(MIN_CLAUSE_EXPRS, MAX_CLAUSE_EXPRS))
                    .map(|_| SortSpec::arbitrary(g))
                    .collect(),
            }
        }
    }

    impl Arbitrary for SortSpec {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                key: SortKey::arbitrary(g),
                direction: SortDirection::arbitrary(g),
            }
        }
    }

    impl Arbitrary for SortKey {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => {
                    // The domain of u32s the parser accepts consists
                    // only of u32s between 0 and MAX_INT_32.
                    let mut p = u32::arbitrary(g);
                    while p > i32::MAX as u32 {
                        p = u32::arbitrary(g);
                    }
                    Self::Positional(p)
                }
                1 => {
                    let rng = &(0..2).collect::<Vec<i32>>();
                    Self::Simple(match g.choose(rng).unwrap() {
                        0 => Expression::Identifier(arbitrary_string()),
                        1 => Expression::Subpath(SubpathExpr::arbitrary(g)),
                        _ => panic!(),
                    })
                }
                _ => panic!("missing SortKey variant(s)"),
            }
        }
    }

    impl Arbitrary for SortDirection {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Asc,
                1 => Self::Desc,
                _ => panic!("missing SortDirection variant(s)"),
            }
        }
    }

    impl Arbitrary for Literal {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Null,
                1 => Self::Boolean(bool::arbitrary(g)),
                2 => Self::String(arbitrary_string()),
                3 => Self::Integer(i32::arbitrary(g).saturating_abs()),
                4 => {
                    let long = i64::arbitrary(g).saturating_abs();
                    if let Ok(int) = i32::try_from(long) {
                        Self::Integer(int)
                    } else {
                        Self::Long(long)
                    }
                }
                5 => {
                    let double = f64::arbitrary(g).abs();
                    if double.is_finite() {
                        Self::Double(double)
                    } else {
                        Self::Double(f64::from(0))
                    }
                }
                _ => panic!("missing Literal variant(s)"),
            }
        }
    }

    impl Arbitrary for Type {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Array,
                1 => Self::BinData,
                2 => Self::Boolean,
                3 => Self::Datetime,
                4 => Self::DbPointer,
                5 => Self::Decimal128,
                6 => Self::Document,
                7 => Self::Double,
                8 => Self::Int32,
                9 => Self::Int64,
                10 => Self::Javascript,
                11 => Self::JavascriptWithScope,
                12 => Self::MaxKey,
                13 => Self::MinKey,
                14 => Self::Null,
                15 => Self::ObjectId,
                16 => Self::RegularExpression,
                17 => Self::String,
                18 => Self::Symbol,
                19 => Self::Timestamp,
                20 => Self::Undefined,
                _ => panic!("missing Type variant(s)"),
            }
        }
    }
}
