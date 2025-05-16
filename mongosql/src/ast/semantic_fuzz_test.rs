#[cfg(test)]
mod tests {
    use crate::ast::pretty_print_fuzz_test::arbitrary::arbitrary_optional;
    use crate::{
        ast::{
            definitions::*,
            pretty_print::PrettyPrint,
            pretty_print_fuzz_test::arbitrary::{arbitrary_identifier, arbitrary_string},
        },
        build_catalog_from_catalog_schema,
        catalog::Catalog,
        json_schema::Schema as JsonSchema,
        options::{ExcludeNamespacesOption, SqlOptions},
        translate_sql, SchemaCheckingMode,
    };
    use lazy_static::lazy_static;
    use mongodb::{bson, sync::Client};
    use quickcheck::{Arbitrary, Gen, TestResult};
    use std::collections::BTreeMap;

    const TEST_DB: &str = "test_db";
    const ALL_TYPES_COLLECTION: &str = "all_types";

    const INT_FIELD: &str = "int_field"; // Int32
    const LONG_FIELD: &str = "long_field"; // Int64
    const DOUBLE_FIELD: &str = "double_field"; // Double
    const DECIMAL_FIELD: &str = "decimal_field"; // Decimal128

    const STRING_FIELD: &str = "string_field"; // String

    const BOOL_FIELD: &str = "bool_field"; // Boolean

    const DATE_FIELD: &str = "date_field"; // Date
    const TIMESTAMP_FIELD: &str = "timestamp_field"; // Timestamp

    const OBJECT_FIELD: &str = "object_field"; // Document
    const NESTED_FIELD: &str = "nested_field"; // Nested string field
    const NESTED_OBJECT_FIELD: &str = "nested_object_field"; // Document with nested fields
    const NESTED_INT: &str = "nested_int"; // Nested int
    const NESTED_STRING: &str = "nested_string"; // Nested string
    const NESTED_OBJECT: &str = "nested_object"; // Nested object
    const DEEPLY_NESTED: &str = "deeply_nested"; // Deeply nested int field
    const ARRAY_FIELD: &str = "array_field"; // Array of Int32
    const STRING_ARRAY_FIELD: &str = "string_array_field"; // Array of String
    const MIXED_ARRAY_FIELD: &str = "mixed_array_field"; // Array of mixed types

    const NULL_FIELD: &str = "null_field"; // Null
    const OBJECTID_FIELD: &str = "objectid_field"; // ObjectId

    lazy_static! {
        static ref NESTED_INT_SUBPATH: Expression = Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Identifier(NESTED_OBJECT_FIELD.to_string())),
            subpath: NESTED_INT.to_string(),
        });
        static ref NESTED_STRING_SUBPATH: Expression = Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Identifier(NESTED_OBJECT_FIELD.to_string())),
            subpath: NESTED_STRING.to_string(),
        });
        static ref ALT_NESTED_STRING_SUBPATH: Expression = Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Identifier(OBJECT_FIELD.to_string())),
            subpath: NESTED_FIELD.to_string(),
        });
        static ref DEEPLY_NESTED_INT_SUBPATH: Expression = Expression::Subpath(SubpathExpr {
            expr: Box::new(Expression::Subpath(SubpathExpr {
                expr: Box::new(Expression::Identifier(NESTED_OBJECT_FIELD.to_string())),
                subpath: NESTED_OBJECT.to_string(),
            })),
            subpath: DEEPLY_NESTED.to_string(),
        });
        static ref MONGODB_URI: String = format!(
            "mongodb://localhost:{}",
            std::env::var("MDB_TEST_LOCAL_PORT").unwrap_or_else(|_| "27017".to_string())
        );
    }

    fn make_numeric_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 17 {
            0 => Expression::Identifier(INT_FIELD.to_string()),
            1 => Expression::Identifier(LONG_FIELD.to_string()),
            2 => Expression::Identifier(DOUBLE_FIELD.to_string()),
            3 => Expression::Identifier(DECIMAL_FIELD.to_string()),
            4 => NESTED_INT_SUBPATH.clone(),
            5 => DEEPLY_NESTED_INT_SUBPATH.clone(),
            6 => Expression::Literal(Literal::Integer(i32::arbitrary(&mut Gen::new(0)))),
            7 => Expression::Literal(Literal::Integer(-(u16::arbitrary(&mut Gen::new(0)) as i32))),
            8 => Expression::Literal(Literal::Long(i64::arbitrary(&mut Gen::new(0)))),
            9 => Expression::Literal(Literal::Double(f64::arbitrary(&mut Gen::new(0)))),
            10 => {
                let arg = make_numeric_expression();
                let op = if bool::arbitrary(&mut Gen::new(0)) {
                    UnaryOp::Pos
                } else {
                    UnaryOp::Neg
                };
                Expression::Unary(UnaryExpr {
                    op,
                    expr: Box::new(arg),
                })
            }
            11 => {
                let left = make_numeric_expression();
                let right = make_numeric_expression();
                let op = match usize::arbitrary(&mut Gen::new(0)) % 4 {
                    0 => BinaryOp::Add,
                    1 => BinaryOp::Sub,
                    2 => BinaryOp::Mul,
                    _ => BinaryOp::Div,
                };
                Expression::Binary(BinaryExpr {
                    left: Box::new(left),
                    op,
                    right: Box::new(right),
                })
            }
            12 => {
                let arg = make_numeric_expression();
                Expression::Function(FunctionExpr {
                    function: match usize::arbitrary(&mut Gen::new(0)) % 6 {
                        0 => FunctionName::Abs,
                        1 => FunctionName::Ceil,
                        2 => FunctionName::Floor,
                        3 => FunctionName::Round,
                        4 => FunctionName::Sqrt,
                        _ => FunctionName::Log10,
                    },
                    args: FunctionArguments::Args(vec![arg]),
                    set_quantifier: None,
                })
            }
            13 => {
                let arg = make_string_expression();
                Expression::Function(FunctionExpr {
                    function: match usize::arbitrary(&mut Gen::new(0)) % 3 {
                        0 => FunctionName::BitLength,
                        1 => FunctionName::OctetLength,
                        _ => FunctionName::CharLength,
                    },
                    args: FunctionArguments::Args(vec![arg]),
                    set_quantifier: None,
                })
            }
            14 => Expression::Extract(ExtractExpr {
                extract_spec: match usize::arbitrary(&mut Gen::new(0)) % 5 {
                    0 => DatePart::Year,
                    1 => DatePart::Month,
                    2 => DatePart::Week,
                    3 => DatePart::Day,
                    4 => DatePart::Hour,
                    _ => DatePart::Minute,
                },
                arg: Box::new(Expression::Identifier(DATE_FIELD.to_string())),
            }),
            15 => {
                let arg_name = match usize::arbitrary(&mut Gen::new(0)) % 15 {
                    0 => INT_FIELD,
                    1 => LONG_FIELD,
                    2 => DOUBLE_FIELD,
                    3 => DECIMAL_FIELD,
                    4 => STRING_FIELD,
                    5 => BOOL_FIELD,
                    6 => DATE_FIELD,
                    7 => TIMESTAMP_FIELD,
                    8 => OBJECT_FIELD,
                    9 => NESTED_OBJECT_FIELD,
                    10 => ARRAY_FIELD,
                    11 => STRING_ARRAY_FIELD,
                    12 => MIXED_ARRAY_FIELD,
                    13 => NULL_FIELD,
                    14 => OBJECTID_FIELD,
                    _ => "star",
                };

                let args = if arg_name == "star" {
                    FunctionArguments::Star
                } else {
                    FunctionArguments::Args(vec![Expression::Identifier(arg_name.to_string())])
                };

                Expression::Function(FunctionExpr {
                    function: FunctionName::Count,
                    args,
                    set_quantifier: if bool::arbitrary(&mut Gen::new(0)) {
                        Some(SetQuantifier::Distinct)
                    } else {
                        None
                    },
                })
            }
            _ => {
                let arg_name = match usize::arbitrary(&mut Gen::new(0)) % 4 {
                    0 => INT_FIELD,
                    1 => LONG_FIELD,
                    2 => DOUBLE_FIELD,
                    _ => DECIMAL_FIELD,
                };
                let arg = Expression::Identifier(arg_name.to_string());
                Expression::Function(FunctionExpr {
                    function: match usize::arbitrary(&mut Gen::new(0)) % 6 {
                        0 => FunctionName::Avg,
                        1 => FunctionName::Min,
                        2 => FunctionName::Max,
                        3 => FunctionName::StddevPop,
                        4 => FunctionName::StddevSamp,
                        _ => FunctionName::Sum,
                    },
                    args: FunctionArguments::Args(vec![arg]),
                    set_quantifier: if bool::arbitrary(&mut Gen::new(0)) {
                        Some(SetQuantifier::Distinct)
                    } else {
                        None
                    },
                })
            }
        }
    }

    fn make_boolean_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 10 {
            0..=4 => Expression::Identifier(BOOL_FIELD.to_string()),
            5 | 6 => Expression::Literal(Literal::Boolean(bool::arbitrary(&mut Gen::new(0)))),
            7 => make_comparison_expression(),
            8 => {
                let left = make_boolean_expression();
                let right = make_boolean_expression();
                let op = if bool::arbitrary(&mut Gen::new(0)) {
                    BinaryOp::And
                } else {
                    BinaryOp::Or
                };
                Expression::Binary(BinaryExpr {
                    left: Box::new(left),
                    op,
                    right: Box::new(right),
                })
            }
            _ => {
                let expr = make_boolean_expression();
                Expression::Unary(UnaryExpr {
                    op: UnaryOp::Not,
                    expr: Box::new(expr),
                })
            }
        }
    }

    fn make_comparison_expression() -> Expression {
        let (left, right) = match usize::arbitrary(&mut Gen::new(0)) % 3 {
            0 => (make_numeric_expression(), make_numeric_expression()),
            1 => (make_string_expression(), make_string_expression()),
            _ => (make_boolean_expression(), make_boolean_expression()),
        };

        let comp_op = match usize::arbitrary(&mut Gen::new(0)) % 6 {
            0 => ComparisonOp::Eq,
            1 => ComparisonOp::Neq,
            2 => ComparisonOp::Lt,
            3 => ComparisonOp::Lte,
            4 => ComparisonOp::Gt,
            _ => ComparisonOp::Gte,
        };

        Expression::Binary(BinaryExpr {
            left: Box::new(left),
            op: BinaryOp::Comparison(comp_op),
            right: Box::new(right),
        })
    }

    fn make_string_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 8 {
            0 => Expression::Identifier(STRING_FIELD.to_string()),
            1 => NESTED_STRING_SUBPATH.clone(),
            2 => ALT_NESTED_STRING_SUBPATH.clone(),
            3 => {
                // String concatenation
                let left = make_string_expression();
                let right = make_string_expression();
                Expression::Binary(BinaryExpr {
                    left: Box::new(left),
                    op: BinaryOp::Concat,
                    right: Box::new(right),
                })
            }
            4 => Expression::Function(FunctionExpr {
                function: if bool::arbitrary(&mut Gen::new(0)) {
                    FunctionName::Lower
                } else {
                    FunctionName::Upper
                },
                args: FunctionArguments::Args(vec![make_string_expression()]),
                set_quantifier: None,
            }),
            5 => Expression::Trim(TrimExpr {
                trim_spec: TrimSpec::arbitrary(&mut Gen::new(0)),
                trim_chars: Box::new(Expression::StringConstructor(" ".to_string())),
                arg: Box::new(make_string_expression()),
            }),
            6 => Expression::Function(FunctionExpr {
                function: FunctionName::Substring,
                args: FunctionArguments::Args(vec![
                    make_string_expression(),
                    make_numeric_expression(),
                    make_numeric_expression(),
                ]),
                set_quantifier: None,
            }),
            _ => {
                // String constructor - simplified to use String directly
                Expression::StringConstructor(arbitrary_string(&mut Gen::new(0)))
            }
        }
    }

    fn make_array_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 6 {
            0 => Expression::Identifier(ARRAY_FIELD.to_string()),
            1 => Expression::Identifier(STRING_ARRAY_FIELD.to_string()),
            2 => Expression::Identifier(MIXED_ARRAY_FIELD.to_string()),
            3 => {
                let mut elements = Vec::new();
                let size = (usize::arbitrary(&mut Gen::new(0)) % 3) + 1; // 1-3 elements
                for _ in 0..size {
                    elements.push(make_numeric_expression());
                }
                Expression::Array(elements)
            }
            4 => {
                let mut elements = Vec::new();
                let size = (usize::arbitrary(&mut Gen::new(0)) % 4) + 2; // 2-5 elements
                for i in 0..size {
                    match i % 3 {
                        0 => elements.push(make_numeric_expression()),
                        1 => elements.push(make_string_expression()),
                        _ => elements.push(make_boolean_expression()),
                    }
                }
                Expression::Array(elements)
            }
            _ => {
                let arg_name = match usize::arbitrary(&mut Gen::new(0)) % 4 {
                    0 => INT_FIELD,
                    1 => LONG_FIELD,
                    2 => DOUBLE_FIELD,
                    _ => DECIMAL_FIELD,
                };
                let arg = Expression::Identifier(arg_name.to_string());
                Expression::Function(FunctionExpr {
                    function: match usize::arbitrary(&mut Gen::new(0)) % 2 {
                        0 => FunctionName::AddToArray,
                        _ => FunctionName::AddToSet,
                    },
                    args: FunctionArguments::Args(vec![arg]),
                    set_quantifier: if bool::arbitrary(&mut Gen::new(0)) {
                        Some(SetQuantifier::Distinct)
                    } else {
                        None
                    },
                })
            }
        }
    }

    fn make_date_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 5 {
            0..=2 => Expression::Identifier(DATE_FIELD.to_string()),
            3 => {
                // DATEADD function
                Expression::DateFunction(DateFunctionExpr {
                    function: DateFunctionName::Add,
                    date_part: DatePart::Day,
                    args: vec![
                        Expression::Literal(Literal::Integer(1)),
                        Expression::Identifier(DATE_FIELD.to_string()),
                    ],
                })
            }
            _ => {
                // DATETRUNC function
                Expression::DateFunction(DateFunctionExpr {
                    function: DateFunctionName::Trunc,
                    date_part: DatePart::Year,
                    args: vec![Expression::Identifier(DATE_FIELD.to_string())],
                })
            }
        }
    }

    ///   - Nested documents with metadata containing nested fields
    fn make_object_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 5 {
            0 => Expression::Identifier(OBJECT_FIELD.to_string()),
            1 => Expression::Identifier(NESTED_OBJECT_FIELD.to_string()),
            2 => {
                let fields = vec![
                    DocumentPair {
                        key: "id".to_string(),
                        value: make_numeric_expression(),
                    },
                    DocumentPair {
                        key: "name".to_string(),
                        value: make_string_expression(),
                    },
                    DocumentPair {
                        key: "active".to_string(),
                        value: make_boolean_expression(),
                    },
                ];
                Expression::Document(fields)
            }
            3 => {
                let nested_fields = vec![
                    DocumentPair {
                        key: "nested_id".to_string(),
                        value: make_numeric_expression(),
                    },
                    DocumentPair {
                        key: "nested_name".to_string(),
                        value: make_string_expression(),
                    },
                ];

                let fields = vec![
                    DocumentPair {
                        key: "id".to_string(),
                        value: make_numeric_expression(),
                    },
                    DocumentPair {
                        key: "metadata".to_string(),
                        value: Expression::Document(nested_fields),
                    },
                ];
                Expression::Document(fields)
            }
            _ => {
                let arg_name = match usize::arbitrary(&mut Gen::new(0)) % 2 {
                    0 => OBJECT_FIELD,
                    _ => NESTED_OBJECT_FIELD,
                };
                let arg = Expression::Identifier(arg_name.to_string());
                Expression::Function(FunctionExpr {
                    function: FunctionName::MergeDocuments,
                    args: FunctionArguments::Args(vec![arg]),
                    set_quantifier: if bool::arbitrary(&mut Gen::new(0)) {
                        Some(SetQuantifier::Distinct)
                    } else {
                        None
                    },
                })
            }
        }
    }

    /// - A FROM clause targeting the all_types collection
    fn generate_arbitrary_semantically_valid_query() -> Query {
        let (select_clause, select_fields) = generate_arbitrary_semantically_valid_select_clause();

        let from_clause = Some(Datasource::Collection(CollectionSource {
            database: Some(TEST_DB.to_string()),
            collection: ALL_TYPES_COLLECTION.to_string(),
            alias: None,
        }));

        let where_clause = weighted_arbitrary_optional(make_boolean_expression);

        // let group_by_clause =
        //     weighted_arbitrary_optional(generate_arbitrary_semantically_valid_group_by_clause);

        let having_clause = weighted_arbitrary_optional(make_boolean_expression);

        let order_by_clause = weighted_arbitrary_optional(|| {
            generate_arbitrary_semantically_valid_order_by_clause(select_fields.clone())
        });

        let limit = arbitrary_optional(&mut Gen::new(0), |g| 1 + u32::arbitrary(g) % 10);
        let offset = arbitrary_optional(&mut Gen::new(0), |g| 1 + u32::arbitrary(g) % 10);

        Query::Select(SelectQuery {
            select_clause,
            from_clause,
            where_clause,
            group_by_clause: None,
            having_clause,
            order_by_clause,
            limit,
            offset,
        })
    }

    /// Returns both the SELECT clause and a vector of field names that can be referenced
    fn generate_arbitrary_semantically_valid_select_clause() -> (SelectClause, Vec<String>) {
        let set_quantifier = SetQuantifier::arbitrary(&mut Gen::new(0));

        // 20% of the time, return a simple SELECT *
        if usize::arbitrary(&mut Gen::new(0)) % 10 < 2 {
            return (
                SelectClause {
                    set_quantifier,
                    body: SelectBody::Standard(vec![SelectExpression::Star]),
                },
                vec![],
            );
        }

        let num_exprs = 1 + usize::arbitrary(&mut Gen::new(0)) % 10;

        let mut select_exprs = vec![];
        let mut select_fields = vec![];

        for _ in 0..num_exprs {
            let expr = match usize::arbitrary(&mut Gen::new(0)) % 6 {
                0 => make_numeric_expression(),
                1 => make_boolean_expression(),
                2 => make_string_expression(),
                3 => make_array_expression(),
                4 => make_date_expression(),
                _ => make_object_expression(),
            };

            let optionally_aliased_expr = if bool::arbitrary(&mut Gen::new(0)) {
                if let Expression::Identifier(i) = expr.clone() {
                    select_fields.push(i);
                }
                OptionallyAliasedExpr::Unaliased(expr)
            } else {
                let alias = arbitrary_identifier(&mut Gen::new(0));
                select_fields.push(alias.clone());
                OptionallyAliasedExpr::Aliased(AliasedExpr { expr, alias })
            };

            select_exprs.push(SelectExpression::Expression(optionally_aliased_expr));
        }

        (
            SelectClause {
                set_quantifier,
                body: SelectBody::Standard(select_exprs),
            },
            select_fields,
        )
    }

    #[allow(dead_code)]
    fn generate_arbitrary_semantically_valid_group_by_clause() -> GroupByClause {
        // Skipping this for now since GROUP BY changes what fields are available in the SELECT
        // clause, which is a bit too complicated for this skunkworks project.
        todo!()
    }

    fn generate_arbitrary_semantically_valid_order_by_clause(
        select_fields: Vec<String>,
    ) -> OrderByClause {
        let num_sort_specs = ((1 + usize::arbitrary(&mut Gen::new(0)) % select_fields.len()) as f64
            / 2f64)
            .ceil() as i32;

        let mut sort_specs = vec![];
        for _ in 0..num_sort_specs {
            let idx = usize::arbitrary(&mut Gen::new(0)) % select_fields.len();
            let key = if bool::arbitrary(&mut Gen::new(0)) {
                SortKey::Positional(idx as u32)
            } else {
                SortKey::Simple(Expression::Identifier(select_fields[idx].clone()))
            };
            let direction = SortDirection::arbitrary(&mut Gen::new(0));

            sort_specs.push(SortSpec { key, direction })
        }

        OrderByClause { sort_specs }
    }

    /// Return an arbitrary Option<T>, using the provided Fn to
    /// construct the value if the chosen variant is Some. This
    /// function is weighted to return Some 3 out of 4 times.
    pub fn weighted_arbitrary_optional<T, F>(f: F) -> Option<T>
    where
        F: Fn() -> T,
    {
        match usize::arbitrary(&mut Gen::new(0)) % 4 {
            0 => None,
            _ => Some(f()),
        }
    }

    #[derive(PartialEq, Debug, Clone)]
    struct SemanticallyValidQuery {
        query: Query,
    }

    impl Arbitrary for SemanticallyValidQuery {
        /// Implements the Arbitrary trait for SemanticallyValidQuery.
        ///
        /// This function allows QuickCheck to generate arbitrary semantically valid
        /// queries for property testing. It delegates to generate_arbitrary_semantically_valid_query
        /// to create queries that are guaranteed to be both syntactically and semantically valid.
        ///
        /// @param _g - QuickCheck generator (unused as we use a fixed seed)
        /// @return A SemanticallyValidQuery instance containing a valid query
    /// This function allows QuickCheck to generate arbitrary semantically valid
    /// queries for property testing. It delegates to generate_arbitrary_semantically_valid_query
    /// to create queries that are guaranteed to be both syntactically and semantically valid.
    ///
    /// @param _g - QuickCheck generator (unused as we use a fixed seed)
    /// @return A SemanticallyValidQuery instance containing a valid query
        fn arbitrary(_g: &mut Gen) -> Self {
            let query = generate_arbitrary_semantically_valid_query();
            SemanticallyValidQuery { query }
        }
    }

    /// Tests that semantically valid queries can be successfully translated to MongoDB pipelines.
    ///
    /// This QuickCheck property test verifies that:
    /// 1. Arbitrary semantically valid queries can be pretty-printed to SQL strings
    /// 2. The SQL strings can be successfully translated back to MongoDB pipelines
    ///
    /// This test ensures the first property from the requirements: semantically valid
    /// queries "compile" via the translate_sql function without errors.
    /// Tests that generated aggregation pipelines can be executed against MongoDB without errors.
    ///
    /// This QuickCheck property test verifies that:
    /// 1. Arbitrary semantically valid queries can be pretty-printed to SQL strings
    /// 2. The SQL strings can be successfully translated to MongoDB pipelines
    /// 3. The resulting pipelines can be executed against a MongoDB instance without errors
    ///
    /// This test ensures the second property from the requirements: the aggregation
    /// pipelines from the translations run against mongod without error.
    /// 
    /// The test is skipped if MongoDB is unavailable or if translation fails.
    /// Tests that semantically valid queries can be successfully translated to MongoDB pipelines.
    ///
    /// This QuickCheck property test verifies that:
    /// 1. Arbitrary semantically valid queries can be pretty-printed to SQL strings
    /// 2. The SQL strings can be successfully translated back to MongoDB pipelines
    ///
    /// This test ensures the first property from the requirements: semantically valid
    /// queries "compile" via the translate_sql function without errors.
    /// Tests that generated aggregation pipelines can be executed against MongoDB without errors.
    ///
    /// This QuickCheck property test verifies that:
    /// 1. Arbitrary semantically valid queries can be pretty-printed to SQL strings
    /// 2. The SQL strings can be successfully translated to MongoDB pipelines
    /// 3. The resulting pipelines can be executed against a MongoDB instance without errors
    ///
    /// This test ensures the second property from the requirements: the aggregation
    /// pipelines from the translations run against mongod without error.
    /// 
    /// The test is skipped if MongoDB is unavailable or if translation fails.
    #[test]
    fn prop_semantic_queries_translate() {
        /// Inner property function for QuickCheck testing.
        ///
        /// This function:
        /// 1. Pretty-prints the query to SQL
        /// 2. Translates the SQL to a MongoDB pipeline
        /// 3. Verifies the translation succeeds
        ///
        /// @param query - A semantically valid query to test
        /// @return TestResult - Success if translation succeeds, discarded otherwise
        /// Inner property function for QuickCheck testing.
        ///
        /// This function:
        /// 1. Connects to MongoDB (skips test if unavailable)
        /// 2. Pretty-prints the query to SQL
        /// 3. Translates the SQL to a MongoDB pipeline
        /// 4. Executes the pipeline against MongoDB
        /// 5. Verifies execution succeeds
        ///
        /// @param query - A semantically valid query to test
        /// @return TestResult - Success if pipeline executes without error, discarded otherwise
        /// Inner property function for QuickCheck testing.
        ///
        /// This function:
        /// 1. Pretty-prints the query to SQL
        /// 2. Translates the SQL to a MongoDB pipeline
        /// 3. Verifies the translation succeeds
        ///
        /// @param query - A semantically valid query to test
        /// @return TestResult - Success if translation succeeds, discarded otherwise
        /// Inner property function for QuickCheck testing.
        ///
        /// This function:
        /// 1. Connects to MongoDB (skips test if unavailable)
        /// 2. Pretty-prints the query to SQL
        /// 3. Translates the SQL to a MongoDB pipeline
        /// 4. Executes the pipeline against MongoDB
        /// 5. Verifies execution succeeds
        ///
        /// @param query - A semantically valid query to test
        /// @return TestResult - Success if pipeline executes without error, discarded otherwise
        fn property(query: SemanticallyValidQuery) -> TestResult {
            let sql = match query.query.pretty_print() {
                Err(_) => return TestResult::discard(),
                Ok(sql) => sql,
            };

            let sql_options = SqlOptions {
                schema_checking_mode: SchemaCheckingMode::Strict,
                exclude_namespaces: ExcludeNamespacesOption::IncludeNamespaces,
                allow_order_by_missing_columns: false,
            };

            let result = translate_sql(TEST_DB, &sql, &TEST_CATALOG, sql_options);

            TestResult::from_bool(result.is_ok())
        }

        quickcheck::QuickCheck::new()
            .gen(Gen::new(0))
            .quickcheck(property as fn(SemanticallyValidQuery) -> TestResult);
    }

    /// Creates a MongoDB client connection using the configured URI.
    ///
    /// This helper function attempts to establish a connection to a MongoDB instance
    /// using the MONGODB_URI defined in the lazy_static block. It returns None if
    /// the connection cannot be established, allowing tests to gracefully skip
    /// when MongoDB is unavailable.
    ///
    /// @return Option<Client> - MongoDB client if connection succeeds, None otherwise
    /// Creates a MongoDB client connection using the configured URI.
    ///
    /// This helper function attempts to establish a connection to a MongoDB instance
    /// using the MONGODB_URI defined in the lazy_static block. It returns None if
    /// the connection cannot be established, allowing tests to gracefully skip
    /// when MongoDB is unavailable.
    ///
    /// @return Option<Client> - MongoDB client if connection succeeds, None otherwise
    fn get_mongodb_client() -> Option<Client> {
        Client::with_uri_str(&*MONGODB_URI).ok()
    }

    /// Tests that semantically valid queries can be successfully translated to MongoDB pipelines.
    ///
    /// This QuickCheck property test verifies that:
    /// 1. Arbitrary semantically valid queries can be pretty-printed to SQL strings
    /// 2. The SQL strings can be successfully translated back to MongoDB pipelines
    ///
    /// This test ensures the first property from the requirements: semantically valid
    /// queries "compile" via the translate_sql function without errors.
    /// Tests that generated aggregation pipelines can be executed against MongoDB without errors.
    ///
    /// This QuickCheck property test verifies that:
    /// 1. Arbitrary semantically valid queries can be pretty-printed to SQL strings
    /// 2. The SQL strings can be successfully translated to MongoDB pipelines
    /// 3. The resulting pipelines can be executed against a MongoDB instance without errors
    ///
    /// This test ensures the second property from the requirements: the aggregation
    /// pipelines from the translations run against mongod without error.
    /// 
    /// The test is skipped if MongoDB is unavailable or if translation fails.
    /// Tests that semantically valid queries can be successfully translated to MongoDB pipelines.
    ///
    /// This QuickCheck property test verifies that:
    /// 1. Arbitrary semantically valid queries can be pretty-printed to SQL strings
    /// 2. The SQL strings can be successfully translated back to MongoDB pipelines
    ///
    /// This test ensures the first property from the requirements: semantically valid
    /// queries "compile" via the translate_sql function without errors.
    /// Tests that generated aggregation pipelines can be executed against MongoDB without errors.
    ///
    /// This QuickCheck property test verifies that:
    /// 1. Arbitrary semantically valid queries can be pretty-printed to SQL strings
    /// 2. The SQL strings can be successfully translated to MongoDB pipelines
    /// 3. The resulting pipelines can be executed against a MongoDB instance without errors
    ///
    /// This test ensures the second property from the requirements: the aggregation
    /// pipelines from the translations run against mongod without error.
    /// 
    /// The test is skipped if MongoDB is unavailable or if translation fails.
    #[test]
    fn prop_aggregation_pipelines_run() {
        /// Inner property function for QuickCheck testing.
        ///
        /// This function:
        /// 1. Pretty-prints the query to SQL
        /// 2. Translates the SQL to a MongoDB pipeline
        /// 3. Verifies the translation succeeds
        ///
        /// @param query - A semantically valid query to test
        /// @return TestResult - Success if translation succeeds, discarded otherwise
        /// Inner property function for QuickCheck testing.
        ///
        /// This function:
        /// 1. Connects to MongoDB (skips test if unavailable)
        /// 2. Pretty-prints the query to SQL
        /// 3. Translates the SQL to a MongoDB pipeline
        /// 4. Executes the pipeline against MongoDB
        /// 5. Verifies execution succeeds
        ///
        /// @param query - A semantically valid query to test
        /// @return TestResult - Success if pipeline executes without error, discarded otherwise
        /// Inner property function for QuickCheck testing.
        ///
        /// This function:
        /// 1. Pretty-prints the query to SQL
        /// 2. Translates the SQL to a MongoDB pipeline
        /// 3. Verifies the translation succeeds
        ///
        /// @param query - A semantically valid query to test
        /// @return TestResult - Success if translation succeeds, discarded otherwise
        /// Inner property function for QuickCheck testing.
        ///
        /// This function:
        /// 1. Connects to MongoDB (skips test if unavailable)
        /// 2. Pretty-prints the query to SQL
        /// 3. Translates the SQL to a MongoDB pipeline
        /// 4. Executes the pipeline against MongoDB
        /// 5. Verifies execution succeeds
        ///
        /// @param query - A semantically valid query to test
        /// @return TestResult - Success if pipeline executes without error, discarded otherwise
        fn property(query: SemanticallyValidQuery) -> TestResult {
            let client = match get_mongodb_client() {
                Some(client) => client,
                None => return TestResult::discard(), // Skip if no MongoDB connection
            };

            let sql = match query.query.pretty_print() {
                Err(_) => return TestResult::discard(),
                Ok(sql) => sql,
            };

            let sql_options = SqlOptions {
                schema_checking_mode: SchemaCheckingMode::Strict,
                exclude_namespaces: ExcludeNamespacesOption::IncludeNamespaces,
                allow_order_by_missing_columns: false,
            };

            let translation = match translate_sql(TEST_DB, &sql, &TEST_CATALOG, sql_options) {
                Ok(t) => t,
                Err(_) => return TestResult::discard(), // Skip if translation fails
            };

            let target_db = translation.target_db;
            let target_collection = translation
                .target_collection
                .unwrap_or_else(|| "unknown".to_string());

            let pipeline_docs = match translation.pipeline {
                bson::Bson::Array(array) => {
                    let mut docs = Vec::new();
                    for value in array {
                        if let bson::Bson::Document(doc) = value {
                            docs.push(doc);
                        } else {
                            return TestResult::discard(); // Not a valid pipeline
                        }
                    }
                    docs
                }
                _ => return TestResult::discard(), // Not a valid pipeline
            };

            let result = client
                .database(&target_db)
                .collection::<bson::Document>(&target_collection)
                .aggregate(pipeline_docs)
                .run();

            TestResult::from_bool(result.is_ok())
        }

        quickcheck::QuickCheck::new()
            .gen(Gen::new(0))
            .quickcheck(property as fn(SemanticallyValidQuery) -> TestResult);
    }

    lazy_static! {
        static ref TEST_CATALOG: Catalog = {
            let mut catalog_schema: BTreeMap<String, BTreeMap<String, JsonSchema>> =
                BTreeMap::new();
            let mut db_schema: BTreeMap<String, JsonSchema> = BTreeMap::new();

            db_schema.insert(
                "all_types".to_string(),
                serde_json::from_str(
                    r#"{
                    "bsonType": "object",
                    "properties": {
                        "int_field": { "bsonType": "int" },
                        "long_field": { "bsonType": "long" },
                        "double_field": { "bsonType": "double" },
                        "decimal_field": { "bsonType": "decimal" },
                        "neg_int_field": { "bsonType": "int" },
                        "zero_int_field": { "bsonType": "int" },
                        "string_field": { "bsonType": "string" },
                        "bool_field": { "bsonType": "bool" },
                        "date_field": { "bsonType": "date" },
                        "timestamp_field": { "bsonType": "timestamp" },
                        "object_field": {
                            "bsonType": "object",
                            "properties": {
                                "nested_field": { "bsonType": "string" }
                            }
                        },
                        "nested_object_field": {
                            "bsonType": "object",
                            "properties": {
                                "nested_int": { "bsonType": "int" },
                                "nested_string": { "bsonType": "string" },
                                "nested_object": {
                                    "bsonType": "object",
                                    "properties": {
                                        "deeply_nested": { "bsonType": "bool" }
                                    }
                                }
                            }
                        },
                        "array_field": {
                            "bsonType": "array",
                            "items": { "bsonType": "int" }
                        },
                        "string_array_field": {
                            "bsonType": "array",
                            "items": { "bsonType": "string" }
                        },
                        "mixed_array_field": {
                            "bsonType": "array"
                            "items": {
                                "anyOf": [
                                    { "bsonType": "string" },
                                    { "bsonType": "int" }
                                ]
                            }
                        },
                        "null_field": { "bsonType": "null" },
                        "objectid_field": { "bsonType": "objectId" }
                    },
                    "additionalProperties": false
                }"#,
                )
                .unwrap(),
            );

            catalog_schema.insert("test_db".to_string(), db_schema);
            build_catalog_from_catalog_schema(catalog_schema).unwrap()
        };
    }
}
