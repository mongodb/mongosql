#[cfg(test)]
mod tests {
    use crate::{
        ast::{
            definitions::*,
            definitions::visitor::Visitor,
            pretty_print::PrettyPrint,
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
    const RELATED_DATA_COLLECTION: &str = "related_data";
    const NUMERIC_COLLECTION: &str = "numeric_data";
    const ARRAY_COLLECTION: &str = "array_data";

    const INT_FIELD: &str = "int_field";           // Int32
    const LONG_FIELD: &str = "long_field";         // Int64
    const DOUBLE_FIELD: &str = "double_field";     // Double
    const DECIMAL_FIELD: &str = "decimal_field";   // Decimal128
    const NEGATIVE_INT_FIELD: &str = "neg_int_field";     // Int32 (negative)
    const ZERO_INT_FIELD: &str = "zero_int_field";        // Int32 (zero)
    
    const STRING_FIELD: &str = "string_field";     // String
    const EMPTY_STRING_FIELD: &str = "empty_string_field"; // String (empty)
    const DESCRIPTION_FIELD: &str = "description"; // String
    
    const BOOL_FIELD: &str = "bool_field";         // Boolean
    const TRUE_FIELD: &str = "true_field";         // Boolean (true)
    const FALSE_FIELD: &str = "false_field";       // Boolean (false)
    
    const DATE_FIELD: &str = "date_field";         // Date
    const TIMESTAMP_FIELD: &str = "timestamp_field"; // Timestamp
    const TIME_FIELD: &str = "time_field";         // Time
    
    const OBJECT_FIELD: &str = "object_field";     // Document
    const NESTED_OBJECT_FIELD: &str = "nested_object_field"; // Document with nested fields
    const ARRAY_FIELD: &str = "array_field";       // Array of Int32
    const STRING_ARRAY_FIELD: &str = "string_array_field"; // Array of String
    const MIXED_ARRAY_FIELD: &str = "mixed_array_field";   // Array of mixed types
    
    const NULL_FIELD: &str = "null_field";         // Null
    const OBJECTID_FIELD: &str = "objectid_field"; // ObjectId
    const ID_FIELD: &str = "id";                   // Int32 (for related_data)
    const ALL_TYPES_ID_FIELD: &str = "all_types_id"; // Int32 (foreign key)
    
    fn field_type(field_name: &str) -> Type {
        match field_name {
            INT_FIELD | NEGATIVE_INT_FIELD | ZERO_INT_FIELD => Type::Int32,
            LONG_FIELD => Type::Int64,
            DOUBLE_FIELD => Type::Double,
            DECIMAL_FIELD => Type::Decimal128,
            
            STRING_FIELD | EMPTY_STRING_FIELD | DESCRIPTION_FIELD => Type::String,
            
            BOOL_FIELD | TRUE_FIELD | FALSE_FIELD => Type::Boolean,
            
            DATE_FIELD => Type::Date,
            TIMESTAMP_FIELD => Type::Timestamp,
            TIME_FIELD => Type::Time,
            
            OBJECT_FIELD | NESTED_OBJECT_FIELD => Type::Document,
            ARRAY_FIELD | STRING_ARRAY_FIELD | MIXED_ARRAY_FIELD => Type::Array,
            
            NULL_FIELD => Type::Null,
            OBJECTID_FIELD => Type::ObjectId,
            ID_FIELD | ALL_TYPES_ID_FIELD => Type::Int32,
            
            _ => Type::String,
        }
    }
    






    // Generate a numeric expression (Int32, Int64, Double, Decimal128)
    fn make_numeric_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 9 {
            0 => Expression::Identifier(INT_FIELD.to_string()),
            1 => Expression::Identifier(LONG_FIELD.to_string()),
            2 => Expression::Identifier(DOUBLE_FIELD.to_string()),
            3 => Expression::Identifier(DECIMAL_FIELD.to_string()),
            4 => Expression::Literal(Literal::Integer(42)),
            5 => Expression::Literal(Literal::Integer(-10)),
            6 => Expression::Literal(Literal::Long(1000000)),
            7 => Expression::Literal(Literal::Double(std::f64::consts::PI)),
            _ => {
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
        }
    }
    
    fn make_boolean_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 7 {
            0 => Expression::Identifier(BOOL_FIELD.to_string()),
            1 => Expression::Identifier(TRUE_FIELD.to_string()),
            2 => Expression::Identifier(FALSE_FIELD.to_string()),
            3 => Expression::Literal(Literal::Boolean(bool::arbitrary(&mut Gen::new(0)))),
            4 => {
                let left = make_numeric_expression();
                let right = make_numeric_expression();
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
            },
            5 => {
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
            },
            _ => {
                let expr = make_boolean_expression();
                Expression::Unary(UnaryExpr {
                    op: UnaryOp::Not,
                    expr: Box::new(expr),
                })
            }
        }
    }
    
    fn make_string_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 5 {
            0 => Expression::Identifier(STRING_FIELD.to_string()),
            1 => Expression::Identifier(EMPTY_STRING_FIELD.to_string()),
            2 => Expression::Identifier(DESCRIPTION_FIELD.to_string()),
            3 => {
                // String concatenation
                let left = make_string_expression();
                let right = make_string_expression();
                Expression::Binary(BinaryExpr {
                    left: Box::new(left),
                    op: BinaryOp::Concat,
                    right: Box::new(right),
                })
            },
            _ => {
                // String constructor - simplified to use String directly
                Expression::StringConstructor(format!("Hello {}!", STRING_FIELD))
            }
        }
    }
    
    fn make_array_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 5 {
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
            },
            _ => {
                let mut elements = Vec::new();
                let size = (usize::arbitrary(&mut Gen::new(0)) % 3) + 1; // 1-3 elements
                for _ in 0..size {
                    elements.push(make_string_expression());
                }
                Expression::Array(elements)
            }
        }
    }
    
    fn make_date_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 5 {
            0 => Expression::Identifier(DATE_FIELD.to_string()),
            1 => Expression::Identifier(TIMESTAMP_FIELD.to_string()),
            2 => Expression::Identifier(TIME_FIELD.to_string()),
            3 => {
                // Date function
                Expression::DateFunction(DateFunctionExpr {
                    function: match usize::arbitrary(&mut Gen::new(0)) % 3 {
                        0 => DateFunctionName::Add,
                        1 => DateFunctionName::Diff,
                        _ => DateFunctionName::Trunc,
                    },
                    date_part: DatePart::Day,
                    args: vec![
                        Expression::Identifier(DATE_FIELD.to_string()),
                        Expression::Literal(Literal::Integer(1)),
                    ]
                })
            },
            _ => {
                Expression::Extract(ExtractExpr {
                    extract_spec: match usize::arbitrary(&mut Gen::new(0)) % 6 {
                        0 => DatePart::Year,
                        1 => DatePart::Quarter,
                        2 => DatePart::Month,
                        3 => DatePart::Week,
                        4 => DatePart::Day,
                        5 => DatePart::Hour,
                        _ => DatePart::Minute,
                    },
                    arg: Box::new(Expression::Identifier(DATE_FIELD.to_string())),
                })
            }
        }
    }
    
    fn make_object_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 4 {
            0 => Expression::Identifier(OBJECT_FIELD.to_string()),
            1 => Expression::Identifier(NESTED_OBJECT_FIELD.to_string()),
            2 => {
                let mut fields = Vec::new();
                fields.push(DocumentPair { key: "id".to_string(), value: make_numeric_expression() });
                fields.push(DocumentPair { key: "name".to_string(), value: make_string_expression() });
                fields.push(DocumentPair { key: "active".to_string(), value: make_boolean_expression() });
                Expression::Document(fields)
            },
            _ => {
                let mut fields = Vec::new();
                fields.push(DocumentPair { key: "id".to_string(), value: make_numeric_expression() });
                
                let mut nested_fields = Vec::new();
                nested_fields.push(DocumentPair { key: "nested_id".to_string(), value: make_numeric_expression() });
                nested_fields.push(DocumentPair { key: "nested_name".to_string(), value: make_string_expression() });
                
                fields.push(DocumentPair { key: "metadata".to_string(), value: Expression::Document(nested_fields) });
                Expression::Document(fields)
            }
        }
    }
    
    #[allow(dead_code)]
    fn make_comparison_expression() -> Expression {
        let left = make_numeric_expression();
        let right = make_numeric_expression();
        
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
    

    
    
    fn expression_type(expr: &Expression) -> Type {
        match expr {
            Expression::Identifier(name) => field_type(name),
            Expression::Literal(lit) => match lit {
                Literal::Integer(_) => Type::Int32,
                Literal::Long(_) => Type::Int64,
                Literal::Double(_) => Type::Double,
                Literal::Boolean(_) => Type::Boolean,
                Literal::Null => Type::Null,
            },
            Expression::Binary(binary) => {
                match binary.op {
                    BinaryOp::Add | BinaryOp::Sub | BinaryOp::Mul | BinaryOp::Div => {
                        let left_type = expression_type(&binary.left);
                        let right_type = expression_type(&binary.right);
                        
                        if left_type == Type::Decimal128 || right_type == Type::Decimal128 {
                            Type::Decimal128
                        } else if left_type == Type::Double || right_type == Type::Double {
                            Type::Double
                        } else if left_type == Type::Int64 || right_type == Type::Int64 {
                            Type::Int64
                        } else {
                            Type::Int32
                        }
                    },
                    BinaryOp::And | BinaryOp::Or => Type::Boolean,
                    BinaryOp::Comparison(_) => Type::Boolean,
                    BinaryOp::In | BinaryOp::NotIn => Type::Boolean,
                    BinaryOp::Concat => Type::String,
                }
            },
            Expression::Unary(unary) => {
                match unary.op {
                    UnaryOp::Not => Type::Boolean,
                    UnaryOp::Neg | UnaryOp::Pos => expression_type(&unary.expr),
                }
            },
            Expression::Cast(cast) => cast.to,
            Expression::Between(_) => Type::Boolean,
            Expression::Case(case) => case.else_branch.as_ref()
                .map_or_else(
                    || case.when_branch.first().map_or(Type::Null, |wb| expression_type(&wb.then)),
                    |else_expr| expression_type(else_expr)
                ),
            Expression::Function(func) => match func.function {
                // Aggregation functions
                FunctionName::Sum | FunctionName::Avg | FunctionName::Min | FunctionName::Max => Type::Double,
                FunctionName::Count => Type::Int64,
                FunctionName::AddToSet | FunctionName::AddToArray => Type::Array,
                FunctionName::First | FunctionName::Last => Type::String, // Depends on the argument type
                
                // String functions
                FunctionName::Substring => Type::String,
                FunctionName::Lower | FunctionName::Upper => Type::String,
                FunctionName::LTrim | FunctionName::RTrim => Type::String,
                FunctionName::Replace => Type::String,
                
                // Date functions
                FunctionName::DateAdd | FunctionName::DateDiff | FunctionName::DateTrunc => Type::Date,
                FunctionName::CurrentTimestamp => Type::Date,
                FunctionName::Year | FunctionName::Month | FunctionName::Week => Type::Int32,
                FunctionName::DayOfWeek | FunctionName::DayOfMonth | FunctionName::DayOfYear => Type::Int32,
                FunctionName::Hour | FunctionName::Minute | FunctionName::Second | FunctionName::Millisecond => Type::Int32,
                
                // Numeric functions
                FunctionName::Abs | FunctionName::Ceil | FunctionName::Floor | FunctionName::Round => Type::Double,
                FunctionName::Log | FunctionName::Log10 | FunctionName::Sqrt => Type::Double,
                FunctionName::Pow => Type::Double,
                FunctionName::Mod => Type::Int32,
                
                // Other functions
                FunctionName::Coalesce => Type::String, // Depends on arguments
                FunctionName::NullIf => Type::String,   // Depends on arguments
                FunctionName::Size => Type::Int32,
                
                _ => Type::String, // Default for other functions
            },
            Expression::Array(_) => Type::Array,
            Expression::Document(_) => Type::Document,
            Expression::Access(access) => {
                let parent_type = expression_type(&access.expr);
                if parent_type == Type::Document {
                    Type::String // Field access from a document, assuming String for simplicity
                } else if parent_type == Type::Array {
                    Type::Int32 // Array access assumes numeric index
                } else {
                    Type::String // Default case
                }
            },
            Expression::Subquery(_) => Type::Array,
            Expression::Exists(_) => Type::Boolean,
            Expression::SubqueryComparison(_) => Type::Boolean,
            Expression::Subpath(_) => Type::String,
            Expression::Is(_) => Type::Boolean,
            Expression::Like(_) => Type::Boolean,
            Expression::StringConstructor(_) => Type::String,
            Expression::Tuple(_) => Type::Array,
            Expression::TypeAssertion(type_assertion) => type_assertion.target_type,
            Expression::Trim(_) => Type::String,
            Expression::DateFunction(_) => Type::Date,
            Expression::Extract(_) => Type::Int32,
        }
    }
    
    fn are_types_compatible(type1: Type, type2: Type) -> bool {
        if type1 == type2 {
            return true;
        }
        
        let is_type1_numeric = matches!(type1, Type::Int32 | Type::Int64 | Type::Double | Type::Decimal128);
        let is_type2_numeric = matches!(type2, Type::Int32 | Type::Int64 | Type::Double | Type::Decimal128);
        
        if is_type1_numeric && is_type2_numeric {
            return true;
        }
        
        
        false
    }
    
    struct SemanticVisitor {
        target_type: Option<Type>,
    }
    
    impl SemanticVisitor {
        fn visit_select_query(&mut self, node: SelectQuery) -> SelectQuery {
            let select_clause = node.select_clause.walk(self);

            let from_clause = Some(Datasource::Collection(CollectionSource {
                database: Some(TEST_DB.to_string()),
                collection: ALL_TYPES_COLLECTION.to_string(),
                alias: None,
            }));

            let old_target_type = self.target_type;
            self.target_type = Some(Type::Boolean);
            let where_clause = node.where_clause.map(|wc| wc.walk(self));
            self.target_type = old_target_type;

            let group_by_clause = node.group_by_clause.map(|gbc| gbc.walk(self));

            let old_target_type = self.target_type;
            self.target_type = Some(Type::Boolean);
            let having_clause = node.having_clause.map(|hc| hc.walk(self));
            self.target_type = old_target_type;

            let order_by_clause = node.order_by_clause.map(|obc| obc.walk(self));

            let limit = node.limit.map(|_| 10);
            let offset = node.offset.map(|_| 0);

            SelectQuery {
                select_clause,
                from_clause,
                where_clause,
                group_by_clause,
                having_clause,
                order_by_clause,
                limit,
                offset,
            }
        }
        
        fn determine_child_target_type(&self, node: &Expression) -> Option<Type> {
            match node {
                Expression::Binary(binary) => {
                    match binary.op {
                        BinaryOp::Add | BinaryOp::Sub | BinaryOp::Mul | BinaryOp::Div => {
                            Some(Type::Double)
                        },
                        BinaryOp::And | BinaryOp::Or => {
                            Some(Type::Boolean)
                        },
                        BinaryOp::Comparison(_) => {
                            None
                        },
                        BinaryOp::In | BinaryOp::NotIn => {
                            None
                        },
                        BinaryOp::Concat => {
                            Some(Type::String)
                        },
                    }
                },
                Expression::Unary(unary) => {
                    match unary.op {
                        UnaryOp::Not => Some(Type::Boolean),
                        UnaryOp::Neg | UnaryOp::Pos => Some(Type::Double),
                    }
                },
                Expression::Function(func) => {
                    match func.function {
                        // Aggregation functions
                        FunctionName::Sum | FunctionName::Avg | FunctionName::Min | FunctionName::Max => Some(Type::Double),
                        FunctionName::Count => None, // Count can take any type
                        FunctionName::AddToSet | FunctionName::AddToArray => None, // Can add any type to arrays
                        
                        // String functions
                        FunctionName::Substring | FunctionName::Lower | FunctionName::Upper => Some(Type::String),
                        FunctionName::LTrim | FunctionName::RTrim => Some(Type::String),
                        FunctionName::Replace => Some(Type::String),
                        
                        // Date functions
                        FunctionName::DateAdd | FunctionName::DateDiff | FunctionName::DateTrunc => Some(Type::Date),
                        FunctionName::CurrentTimestamp => Some(Type::Date),
                        
                        // Numeric functions
                        FunctionName::Abs | FunctionName::Ceil | FunctionName::Floor | FunctionName::Round => Some(Type::Double),
                        FunctionName::Log | FunctionName::Log10 | FunctionName::Sqrt => Some(Type::Double),
                        FunctionName::Pow => Some(Type::Double),
                        
                        // Other functions
                        FunctionName::Coalesce | FunctionName::NullIf => None,
                        FunctionName::Size => None,
                        
                        _ => None, // Default for other functions
                    }
                },
                Expression::Case(_case) => {
                    Some(Type::Boolean)
                },
                Expression::Between(_) => {
                    None
                },
                Expression::Is(_) | Expression::Like(_) | Expression::Exists(_) => {
                    None
                },
                Expression::Array(_) => None,
                Expression::Document(_) => None,
                Expression::Access(_) => None,
                Expression::Subquery(_) => None,
                Expression::SubqueryComparison(_) => None,
                Expression::Subpath(_) => None,
                Expression::StringConstructor(_) => None,
                Expression::TypeAssertion(_) => None,
                Expression::Trim(_) => None,
                Expression::DateFunction(_) => None,
                Expression::Extract(_) => None,
                Expression::Identifier(_) => None,
                Expression::Literal(_) => None,
                Expression::Tuple(_) => None,
                Expression::Cast(_) => None,
            }
        }
        

    }
    
    impl visitor::Visitor for SemanticVisitor {
        fn visit_query(&mut self, node: Query) -> Query {
            match node {
                Query::Select(select_query) => {
                    Query::Select(self.visit_select_query(select_query))
                },
                Query::Set(set_query) => {
                    let old_target_type = self.target_type;
                    self.target_type = None; // Clear target_type when walking set operations
                    let walked = Query::Set(set_query.walk(self));
                    self.target_type = old_target_type;
                    walked
                },
                Query::With(with_query) => {
                    let old_target_type = self.target_type;
                    self.target_type = None; // Clear target_type when walking with queries
                    let walked = Query::With(with_query.walk(self));
                    self.target_type = old_target_type;
                    walked
                },
            }
        }
        
        fn visit_expression(&mut self, node: Expression) -> Expression {
            let mut expr = node.clone();
            self.visit_expression_custom(&mut expr);
            expr
        }
        
        fn visit_sort_key(&mut self, node: SortKey) -> SortKey {
            match node {
                SortKey::Positional(_) => {
                    SortKey::Simple(Expression::Identifier(INT_FIELD.to_string()))
                },
                _ => node.walk(self),
            }
        }
    }
    
    impl SemanticVisitor {
        fn visit_expression_custom(&mut self, node: &mut Expression) {
            if let Expression::Tuple(_) = node {
                *node = make_numeric_expression();
                return;
            }
            
            if let Some(target_type) = self.target_type {
                let node_type = expression_type(node);
                
                if node_type != target_type && !are_types_compatible(node_type, target_type) {
                    *node = match target_type {
                        Type::Boolean => make_boolean_expression(),
                        Type::Int32 | Type::Int64 | Type::Double | Type::Decimal128 => make_numeric_expression(),
                        Type::String => make_string_expression(),
                        Type::Array => make_array_expression(),
                        Type::Date | Type::Datetime | Type::Timestamp => make_date_expression(),
                        Type::Document => make_object_expression(),
                        _ => node.clone(), // Keep the original node for other types
                    };
                }
            }
            
            let child_target_type = self.determine_child_target_type(node);
            
            let old_target_type = self.target_type;
            self.target_type = child_target_type;
            
            match node {
                Expression::Binary(bin) => {
                    self.visit_expression_custom(&mut bin.left);
                    self.visit_expression_custom(&mut bin.right);
                },
                Expression::Unary(un) => {
                    self.visit_expression_custom(&mut un.expr);
                },
                Expression::Function(func) => {
                    if let FunctionArguments::Args(args) = &mut func.args {
                        for arg in args {
                            self.visit_expression_custom(arg);
                        }
                    }
                },
                Expression::Case(case) => {
                    for branch in &mut case.when_branch {
                        self.visit_expression_custom(&mut branch.when);
                        self.visit_expression_custom(&mut branch.then);
                    }
                    if let Some(else_branch) = &mut case.else_branch {
                        self.visit_expression_custom(else_branch);
                    }
                },
                Expression::Array(array) => {
                    for elem in array {
                        self.visit_expression_custom(elem);
                    }
                },
                Expression::Document(doc) => {
                    for pair in doc {
                        self.visit_expression_custom(&mut pair.value);
                    }
                },
                Expression::Access(access) => {
                    self.visit_expression_custom(&mut access.expr);
                },
                Expression::Subquery(subquery) => {
                },
                Expression::Exists(exists) => {
                },
                Expression::SubqueryComparison(comp) => {
                },
                Expression::Subpath(subpath) => {
                    self.visit_expression_custom(&mut subpath.expr);
                },
                Expression::Is(is_expr) => {
                    self.visit_expression_custom(&mut is_expr.expr);
                },
                Expression::Like(like) => {
                    self.visit_expression_custom(&mut like.expr);
                    self.visit_expression_custom(&mut like.pattern);
                },
                Expression::StringConstructor(_) => {
                },
                Expression::TypeAssertion(type_assertion) => {
                    self.visit_expression_custom(&mut type_assertion.expr);
                },
                Expression::Between(between) => {
                    self.visit_expression_custom(&mut between.arg);
                    self.visit_expression_custom(&mut between.min);
                    self.visit_expression_custom(&mut between.max);
                },
                Expression::Trim(trim) => {
                    self.visit_expression_custom(&mut trim.arg);
                },
                Expression::DateFunction(date_func) => {
                },
                Expression::Extract(extract) => {
                    self.visit_expression_custom(&mut extract.arg);
                },
                Expression::Identifier(_) | Expression::Literal(_) => {
                },
                Expression::Cast(cast) => {
                    self.visit_expression_custom(&mut cast.expr);
                },
                Expression::Tuple(tuple) => {
                    for expr in tuple {
                        self.visit_expression_custom(expr);
                    }
                },
            }
            
            self.target_type = old_target_type;
        }
    }
    
    #[allow(dead_code)]
    fn ensure_numeric_expression(expr: &mut Expression) {
        if !matches!(expression_type(expr), Type::Int32 | Type::Int64 | Type::Double | Type::Decimal128) {
            *expr = match usize::arbitrary(&mut Gen::new(0)) % 4 {
                0 => Expression::Identifier(INT_FIELD.to_string()),
                1 => Expression::Identifier(LONG_FIELD.to_string()),
                2 => Expression::Identifier(DOUBLE_FIELD.to_string()),
                _ => Expression::Literal(Literal::Integer(42)),
            };
        }
    }
    
    #[allow(dead_code)]
    fn ensure_boolean_expression(expr: &mut Expression) {
        if expression_type(expr) != Type::Boolean {
            *expr = match usize::arbitrary(&mut Gen::new(0)) % 3 {
                0 => Expression::Identifier(BOOL_FIELD.to_string()),
                1 => Expression::Literal(Literal::Boolean(bool::arbitrary(&mut Gen::new(0)))),
                _ => {
                    Expression::Binary(BinaryExpr {
                        left: Box::new(Expression::Identifier(INT_FIELD.to_string())),
                        op: BinaryOp::Comparison(ComparisonOp::Eq),
                        right: Box::new(Expression::Literal(Literal::Integer(42))),
                    })
                }
            };
        }
    }

    fn contains_invalid_select_query(query: &Query) -> bool {
        match query {
            Query::Select(select) => {
                select.from_clause.is_none() && matches!(select.select_clause.body, SelectBody::Values(_))
            },
            Query::Set(set) => {
                contains_invalid_select_query(&set.left) || contains_invalid_select_query(&set.right)
            },
            Query::With(with) => {
                if contains_invalid_select_query(&with.body) {
                    return true;
                }
                
                for named_query in &with.queries {
                    if contains_invalid_select_query(&named_query.query) {
                        return true;
                    }
                }
                false
            }
        }
    }

    #[test]
    fn prop_semantic_queries_translate() {
        fn property(mut query: Query) -> TestResult {
            if contains_invalid_select_query(&query) {
                return TestResult::discard();
            }
            
            let mut v = SemanticVisitor { target_type: None };
            query = v.visit_query(query);
            
            let sql = match query.pretty_print() {
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
            .quickcheck(property as fn(Query) -> TestResult);
    }

    lazy_static! {
        static ref MONGODB_URI: String = format!(
            "mongodb://localhost:{}",
            std::env::var("MDB_TEST_LOCAL_PORT").unwrap_or_else(|_| "27017".to_string())
        );
    }

    fn get_mongodb_client() -> Option<Client> {
        Client::with_uri_str(&*MONGODB_URI).ok()
    }

    #[test]
    fn prop_aggregation_pipelines_run() {
        // Skip test if MongoDB connection fails
        let _client = match get_mongodb_client() {
            Some(client) => client,
            None => {
                println!("Skipping test: MongoDB connection failed");
                return;
            }
        };
        
        fn property(mut query: Query) -> TestResult {
            if contains_invalid_select_query(&query) {
                return TestResult::discard();
            }
            
            let mut v = SemanticVisitor { target_type: None };
            query = v.visit_query(query);
            
            let client = match get_mongodb_client() {
                Some(client) => client,
                None => return TestResult::discard(), // Skip if no MongoDB connection
            };
            
            let sql = match query.pretty_print() {
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
            let target_collection = translation.target_collection.unwrap_or_else(|| "unknown".to_string());
            
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
                },
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
            .quickcheck(property as fn(Query) -> TestResult);
    }

    lazy_static! {
        static ref TEST_CATALOG: Catalog = {
            let mut catalog_schema: BTreeMap<String, BTreeMap<String, JsonSchema>> = BTreeMap::new();
            let mut db_schema: BTreeMap<String, JsonSchema> = BTreeMap::new();
            
            db_schema.insert(
                "all_types".to_string(),
                serde_json::from_str(r#"{
                    "bsonType": "object",
                    "properties": {
                        "int_field": { "bsonType": "int" },
                        "long_field": { "bsonType": "long" },
                        "double_field": { "bsonType": "double" },
                        "decimal_field": { "bsonType": "decimal" },
                        "neg_int_field": { "bsonType": "int" },
                        "zero_int_field": { "bsonType": "int" },
                        "string_field": { "bsonType": "string" },
                        "empty_string_field": { "bsonType": "string" },
                        "bool_field": { "bsonType": "bool" },
                        "true_field": { "bsonType": "bool" },
                        "false_field": { "bsonType": "bool" },
                        "date_field": { "bsonType": "date" },
                        "timestamp_field": { "bsonType": "timestamp" },
                        "time_field": { "bsonType": "timestamp" },
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
                        },
                        "null_field": { "bsonType": "null" },
                        "objectid_field": { "bsonType": "objectId" }
                    },
                    "additionalProperties": false
                }"#).unwrap(),
            );
            
            db_schema.insert(
                "related_data".to_string(),
                serde_json::from_str(r#"{
                    "bsonType": "object",
                    "properties": {
                        "id": { "bsonType": "int" },
                        "all_types_id": { "bsonType": "int" },
                        "description": { "bsonType": "string" }
                    },
                    "additionalProperties": false
                }"#).unwrap(),
            );
            
            db_schema.insert(
                "numeric_data".to_string(),
                serde_json::from_str(r#"{
                    "bsonType": "object",
                    "properties": {
                        "id": { "bsonType": "int" },
                        "int_value": { "bsonType": "int" },
                        "long_value": { "bsonType": "long" },
                        "double_value": { "bsonType": "double" },
                        "decimal_value": { "bsonType": "decimal" },
                        "calculated_field": { "bsonType": "double" }
                    },
                    "additionalProperties": false
                }"#).unwrap(),
            );
            
            db_schema.insert(
                "array_data".to_string(),
                serde_json::from_str(r#"{
                    "bsonType": "object",
                    "properties": {
                        "id": { "bsonType": "int" },
                        "int_array": { 
                            "bsonType": "array",
                            "items": { "bsonType": "int" }
                        },
                        "string_array": { 
                            "bsonType": "array",
                            "items": { "bsonType": "string" }
                        },
                        "object_array": { 
                            "bsonType": "array",
                            "items": { 
                                "bsonType": "object",
                                "properties": {
                                    "key": { "bsonType": "string" },
                                    "value": { "bsonType": "int" }
                                }
                            }
                        },
                        "nested_array": { 
                            "bsonType": "array",
                            "items": { 
                                "bsonType": "array",
                                "items": { "bsonType": "int" }
                            }
                        }
                    },
                    "additionalProperties": false
                }"#).unwrap(),
            );
            
            catalog_schema.insert("test_db".to_string(), db_schema);
            build_catalog_from_catalog_schema(catalog_schema).unwrap()
        };
    }
}
