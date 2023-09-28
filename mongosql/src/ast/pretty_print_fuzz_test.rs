#[cfg(test)]

mod fuzz_test {
    use crate::{
        ast::{
            definitions::*,
            pretty_print::PrettyPrint,
            rewrites::{Pass, SingleTupleRewritePass},
        },
        parser,
    };
    use quickcheck::*;

    #[test]
    // For arbitrary Query q, this test asserts the property:
    //
    //   q == Parse(PrettyPrint(q))
    //
    // As in, pretty printing a query and then reparsing that
    // pretty string results in the original query.
    fn prop_pretty_print_parse_is_idempotent() {
        fn pretty_print_query(q: Query) -> TestResult {
            // First pretty print the Query.
            let p = match q.pretty_print() {
                Err(_) => return TestResult::discard(),
                Ok(p) => p,
            };

            let reparsed = parser::parse_query(&p);
            let reparsed = match reparsed {
                Ok(parsed) => parsed,
                Err(ref e) => {
                    // If we failed to parse, we want to show the Error, the pretty printed
                    // query, and the query AST so we can more easily see the issue.
                    // The panic will print the query AST.
                    panic!("\nError:\n{e:?}\n=========\n{p}\n\n_________\n\n");
                }
            };

            // When we reparse after pretty printing,  some parenthesized  expressions are
            // interpreted as single-element tuples.  This is an intentional choice in the
            // parser on our part. The SingleTupleRewritePass unwraps these single-element
            // tuples into their single expressions.  We discard any  tests that encounter
            // rewrite errors,  though the  SingleTupleRewritePass  should never return an
            // error anyway.
            let reparsed = SingleTupleRewritePass.apply(reparsed);
            match reparsed {
                Err(_) => TestResult::discard(),
                Ok(r) if q != r => {
                    panic!(
                        r#"Reparsed query AST does not equal original AST

Original AST:
{q:?}

Pretty-printed query:
{p:?}

Reparsed AST:
{r:?}
"#
                    )
                }
                _ => TestResult::from_bool(true),
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
    static NEST_LOWER_BOUND: usize = 10;
    static NEST_UPPER_BOUND: usize = 25;

    /// Return an arbitrary String without null characters and with
    /// the specified maximum length.
    ///
    /// These Strings can be used for aliases, identifiers, or literals.
    fn arbitrary_string_with_max_len(max_len: usize) -> String {
        let g = &mut Gen::new(max_len);
        String::arbitrary(g).replace('\u{0}', "")
    }

    /// Return an arbitrary String without null characters.
    ///
    /// These Strings can be used for aliases, identifiers, or literals.
    fn arbitrary_string(_: &mut Gen) -> String {
        arbitrary_string_with_max_len(rand_len(1, 20) as usize)
    }

    /// Return an arbitrary Option<T>, using the provided Fn to
    /// construct the value if the chosen variant is Some.
    fn arbitrary_optional<T, F>(g: &mut Gen, f: F) -> Option<T>
    where
        F: Fn(&mut Gen) -> T,
    {
        if bool::arbitrary(g) {
            None
        } else {
            Some(f(g))
        }
    }

    /// arbitrary_identifier generates an arbitraty identifier, currently
    /// it just uses arbitrary_string, but this allows us to fine tune
    /// easily if we decide to use different rules for identifiers from
    /// strings.
    fn arbitrary_identifier(g: &mut Gen) -> String {
        arbitrary_string(g)
    }

    /// Return an arbitrary Option<String> conforming to identifier charset.
    fn arbitrary_optional_identifier(g: &mut Gen) -> Option<String> {
        arbitrary_optional(g, arbitrary_identifier)
    }

    fn arbitrary_expr_terminal(g: &mut Gen) -> Expression {
        if bool::arbitrary(g) {
            Expression::Literal(Literal::arbitrary(g))
        } else {
            Expression::Identifier(arbitrary_identifier(g))
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

        fn shrink(&self) -> Box<dyn Iterator<Item = Self>> {
            match self {
                Self::Select(q) => Box::new(q.shrink().map(Self::Select)),
                Self::Set(q) => {
                    // Note that shrink returns an iterator of values that
                    // are smaller than self. For a Set query, this includes
                    // not only the SetQuery itself being shrunk, but also
                    // just the left query and just the right query.
                    let chain = q
                        .shrink()
                        .map(Self::Set)
                        .chain(q.left.shrink().map(|l| *l))
                        .chain(q.right.shrink().map(|r| *r));
                    Box::new(chain)
                }
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

        fn shrink(&self) -> Box<dyn Iterator<Item = Self>> {
            // TODO:
            //   - shrink select_clause
            //   - better shrinking overall
            let mut s = vec![];

            if self.from_clause.is_some() {
                let mut q = self.clone();
                q.from_clause = None;
                s.push(q)
            };

            if self.where_clause.is_some() {
                let mut q = self.clone();
                q.where_clause = None;
                s.push(q)
            };

            if self.group_by_clause.is_some() {
                let mut q = self.clone();
                q.group_by_clause = None;
                s.push(q)
            };

            if self.having_clause.is_some() {
                let mut q = self.clone();
                q.having_clause = None;
                s.push(q)
            };

            if self.order_by_clause.is_some() {
                let mut q = self.clone();
                q.order_by_clause = None;
                s.push(q)
            };

            if self.limit.is_some() {
                let mut q = self.clone();
                q.limit = None;
                s.push(q)
            };

            if self.offset.is_some() {
                let mut q = self.clone();
                q.offset = None;
                s.push(q)
            };

            Box::new(s.into_iter())
        }
    }

    impl Arbitrary for SetQuery {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                left: Box::new(Query::arbitrary(g)),
                op: SetOperator::arbitrary(g),
                // Only generate Select queries for the RHS
                // because the parser always produces left-deep Set Queries.
                right: Box::new(Query::Select(SelectQuery::arbitrary(g))),
            }
        }

        fn shrink(&self) -> Box<dyn Iterator<Item = Self>> {
            let left = self.left.clone();
            let right = self.right.clone();
            let op = self.op;
            Box::new((left, right).shrink().map(move |(l, r)| SetQuery {
                left: l,
                right: r,
                op,
            }))
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
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                datasource: arbitrary_identifier(g),
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
                4 => Self::Flatten(FlattenSource::arbitrary(g)),
                5 => Self::Unwind(UnwindSource::arbitrary(g)),
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
                alias: arbitrary_identifier(g),
            }
        }
    }

    impl Arbitrary for CollectionSource {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                database: arbitrary_optional_identifier(g),
                collection: arbitrary_identifier(g),
                alias: arbitrary_optional_identifier(g),
            }
        }
    }

    impl Arbitrary for DerivedSource {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                query: Box::new(Query::arbitrary(g)),
                alias: arbitrary_identifier(g),
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
                alias: arbitrary_identifier(g),
            }
        }
    }

    impl Arbitrary for JoinSource {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Datasource::VARIANT_COUNT - 1).collect::<Vec<_>>();
            // The RHS should not be a Join because the parser always produces left-deep Joins.  To
            // avoid right-deep joins, we subtract 1 from the Datasource VARIANT_COUNT to remove
            // the possiblity of generating a Join and directly generate an arbitrary non-Join RHS
            // datasource here.
            let rhs = Box::new(match g.choose(rng).unwrap() {
                0 => Datasource::Array(ArraySource::arbitrary(g)),
                1 => Datasource::Collection(CollectionSource::arbitrary(g)),
                2 => Datasource::Derived(DerivedSource::arbitrary(g)),
                3 => Datasource::Flatten(FlattenSource::arbitrary(g)),
                4 => Datasource::Unwind(UnwindSource::arbitrary(g)),
                _ => panic!("missing Datasource variant(s)"),
            });
            Self {
                join_type: JoinType::arbitrary(g),
                left: Box::new(Datasource::arbitrary(g)),
                right: rhs,
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

    impl Arbitrary for FlattenSource {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                datasource: Box::new(Datasource::arbitrary(g)),
                options: (0..rand_len(MIN_COMPOSITE_DATA_LEN, MAX_COMPOSITE_DATA_LEN))
                    .map(|_| FlattenOption::arbitrary(g))
                    .collect(),
            }
        }
    }

    impl Arbitrary for FlattenOption {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Separator(arbitrary_string(g)),
                1 => Self::Depth(u32::arbitrary(g)),
                _ => panic!("missing Flatten variant(s)"),
            }
        }
    }

    impl Arbitrary for UnwindSource {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                datasource: Box::new(Datasource::arbitrary(g)),
                options: (0..rand_len(MIN_COMPOSITE_DATA_LEN, MAX_COMPOSITE_DATA_LEN))
                    .map(|_| UnwindOption::arbitrary(g))
                    .collect(),
            }
        }
    }

    impl Arbitrary for UnwindOption {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => {
                    // The parser only supports Identifiers or Subpath expressions for PATH.
                    match bool::arbitrary(g) {
                        true => Self::Path(Expression::Identifier(arbitrary_identifier(g))),
                        false => Self::Path(Expression::Subpath(SubpathExpr::arbitrary(g))),
                    }
                }
                1 => Self::Index(arbitrary_identifier(g)),
                2 => Self::Outer(bool::arbitrary(g)),
                _ => panic!("missing UnwindOption variant(s)"),
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
                7 => Self::DateFunction(DateFunctionExpr::arbitrary(nested_g)),
                8 => Self::Cast(CastExpr::arbitrary(nested_g)),
                9 => Self::Array(
                    (0..rand_len(MIN_COMPOSITE_DATA_LEN, MAX_COMPOSITE_DATA_LEN))
                        .map(|_| Self::arbitrary(nested_g))
                        .collect(),
                ),
                10 => Self::Subquery(Box::new(Query::arbitrary(nested_g))),
                11 => Self::Exists(Box::new(Query::arbitrary(nested_g))),
                12 => Self::SubqueryComparison(SubqueryComparisonExpr::arbitrary(nested_g)),
                13 => Self::Document(
                    (0..rand_len(MIN_COMPOSITE_DATA_LEN, MAX_COMPOSITE_DATA_LEN))
                        .map(|_| DocumentPair::arbitrary(nested_g))
                        .collect(),
                ),
                14 => Self::Access(AccessExpr::arbitrary(nested_g)),
                15 => Self::Subpath(SubpathExpr::arbitrary(nested_g)),
                16 => Self::Identifier(arbitrary_identifier(g)),
                17 => Self::Is(IsExpr::arbitrary(nested_g)),
                18 => Self::Like(LikeExpr::arbitrary(nested_g)),
                19 => Self::Literal(Literal::arbitrary(nested_g)),
                20 => Self::Tuple((1..4).map(|_| Self::arbitrary(nested_g)).collect()),
                21 => Self::TypeAssertion(TypeAssertionExpr::arbitrary(nested_g)),
                _ => panic!("missing Expression variant(s)"),
            }
        }
    }

    impl Arbitrary for DocumentPair {
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                key: arbitrary_string(g),
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
                op: ComparisonOp::arbitrary(g),
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
                extract_spec: DatePart::arbitrary(g),
                arg: Box::new(Expression::arbitrary(g)),
            }
        }
    }
    impl Arbitrary for DateFunctionName {
        fn arbitrary(g: &mut Gen) -> Self {
            // Intentionally omitting Diff because it cannot be built
            // as arbitrarily as the other functions, as noted in
            // DateFunctionExpr::arbitrary.
            let rng = &(0..(Self::VARIANT_COUNT - 1)).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Add,
                1 => Self::Trunc,
                _ => panic!("missing DateFunctionName variant(s)"),
            }
        }
    }

    impl Arbitrary for DateFunctionExpr {
        fn arbitrary(g: &mut Gen) -> Self {
            // DateDiff needs to be special-cased since it requires more arguments
            // than the other two date functions.
            let rng = &(0..DateFunctionName::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                // DateDiff can syntactically accept exactly 3 non-date part arguments.
                0 => Self {
                    function: DateFunctionName::Diff,
                    date_part: DatePart::arbitrary(g),
                    args: vec![
                        Expression::arbitrary(g),
                        Expression::arbitrary(g),
                        Expression::arbitrary(g),
                    ],
                },
                // DateAdd and DateTrunc can syntactically accept 1 or 2 non-date part arguments.
                _ => Self {
                    function: DateFunctionName::arbitrary(g),
                    date_part: DatePart::arbitrary(g),
                    args: vec![Expression::arbitrary(g), Expression::arbitrary(g)],
                },
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
                // Aggregation Functions
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

                // Scalar functions.
                12 => Self::Abs,
                13 => Self::BitLength,
                14 => Self::Ceil,
                15 => Self::CharLength,
                16 => Self::Coalesce,
                17 => Self::Cos,
                // Self::CurrentTimestamp intentionally omitted
                18 => Self::Degrees,
                19 => Self::Floor,
                20 => Self::Log,
                21 => Self::Lower,
                22 => Self::Mod,
                23 => Self::NullIf,
                24 => Self::OctetLength,
                // Self::Position intentionally omitted
                25 => Self::Pow,
                26 => Self::Radians,
                27 => Self::Round,
                28 => Self::Sin,
                29 => Self::Size,
                30 => Self::Slice,
                31 => Self::Split,
                32 => Self::Sqrt,
                // Self::Substring intentionally omitted
                33 => Self::Tan,
                34 => Self::Upper,
                35 => Self::LTrim,
                36 => Self::RTrim,
                37 => Self::Log10,
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

    impl Arbitrary for DatePart {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT - 4).collect::<Vec<_>>();
            // Intentionally omitting Quarter, DayOfYear, IsoWeek, and IsoWeekday so
            // both Extract and DateFunction can utilize this method.
            match g.choose(rng).unwrap() {
                0 => Self::Year,
                1 => Self::Month,
                2 => Self::Day,
                3 => Self::Hour,
                4 => Self::Minute,
                5 => Self::Second,
                6 => Self::Week,
                _ => panic!("missing DatePart variant(s)"),
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
        fn arbitrary(g: &mut Gen) -> Self {
            Self {
                expr: Box::new(Expression::Identifier(arbitrary_identifier(g))),
                subpath: arbitrary_identifier(g),
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
            let escape = match char::arbitrary(g) {
                '\0' => 'x',
                c => c,
            };
            Self {
                expr: Box::new(Expression::arbitrary(g)),
                pattern: Box::new(Expression::arbitrary(g)),
                escape: arbitrary_optional(g, |_| escape),
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
                4 => Self::In,
                5 => Self::Mul,
                6 => Self::NotIn,
                7 => Self::Or,
                8 => Self::Sub,
                9 => Self::Comparison(ComparisonOp::arbitrary(g)),
                _ => panic!("missing BinaryOp variant(s)"),
            }
        }
    }

    impl Arbitrary for ComparisonOp {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::Eq,
                1 => Self::Gt,
                2 => Self::Gte,
                3 => Self::Lt,
                4 => Self::Lte,
                5 => Self::Neq,
                _ => panic!("missing ComparisonOp variant(s)"),
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
                        0 => Expression::Identifier(arbitrary_identifier(g)),
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
                2 => Self::String(arbitrary_string(g)),
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
                3 => Self::Date,
                4 => Self::Datetime,
                5 => Self::DbPointer,
                6 => Self::Decimal128,
                7 => Self::Document,
                8 => Self::Double,
                9 => Self::Int32,
                10 => Self::Int64,
                11 => Self::Javascript,
                12 => Self::JavascriptWithScope,
                13 => Self::MaxKey,
                14 => Self::MinKey,
                15 => Self::Null,
                16 => Self::ObjectId,
                17 => Self::RegularExpression,
                18 => Self::String,
                19 => Self::Symbol,
                20 => Self::Time,
                21 => Self::Timestamp,
                22 => Self::Undefined,
                _ => panic!("missing Type variant(s)"),
            }
        }
    }
}
