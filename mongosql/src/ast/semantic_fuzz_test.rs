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
    
    #[allow(dead_code)]
    fn is_numeric_field(field_name: &str) -> bool {
        matches!(field_type(field_name), 
            Type::Int32 | Type::Int64 | Type::Double | Type::Decimal128)
    }
    
    #[allow(dead_code)]
    fn is_boolean_field(field_name: &str) -> bool {
        field_type(field_name) == Type::Boolean
    }
    
    #[allow(dead_code)]
    fn is_string_field(field_name: &str) -> bool {
        field_type(field_name) == Type::String
    }

    #[allow(dead_code)]
    fn make_query_semantic(query: &mut Query) {
        match query {
            Query::Select(select) => {
                if select.from_clause.is_none() {
                    let collection = if bool::arbitrary(&mut Gen::new(0)) {
                        ALL_TYPES_COLLECTION
                    } else {
                        RELATED_DATA_COLLECTION
                    };
                    
                    select.from_clause = Some(Datasource::Collection(CollectionSource {
                        database: Some(TEST_DB.to_string()),
                        collection: collection.to_string(),
                        alias: None,
                    }));
                }
                make_select_query_semantic(select);
            },
            Query::Set(set) => {
                make_query_semantic(set.left.as_mut());
                make_query_semantic(set.right.as_mut());
                
                match set.op {
                    SetOperator::Union | SetOperator::UnionAll => {
                        set.op = SetOperator::Union;
                    },
                }
            },
            Query::With(with) => {
                if with.queries.is_empty() {
                    with.queries.push(NamedQuery {
                        name: format!("cte_{}", usize::arbitrary(&mut Gen::new(0)) % 100),
                        query: Query::Select(SelectQuery {
                            select_clause: SelectClause {
                                set_quantifier: SetQuantifier::All,
                                body: SelectBody::Standard(vec![
                                    SelectExpression::Expression(OptionallyAliasedExpr::Unaliased(
                                        make_numeric_expression()
                                    ))
                                ]),
                            },
                            from_clause: Some(Datasource::Collection(CollectionSource {
                                database: Some(TEST_DB.to_string()),
                                collection: ALL_TYPES_COLLECTION.to_string(),
                                alias: None,
                            })),
                            where_clause: None,
                            group_by_clause: None,
                            having_clause: None,
                            order_by_clause: None,
                            limit: None,
                            offset: None,
                        }),
                    });
                }
                
                if let Query::Select(select) = &mut *with.body {
                    if select.from_clause.is_none() {
                        let collection = if bool::arbitrary(&mut Gen::new(0)) {
                            ALL_TYPES_COLLECTION
                        } else {
                            RELATED_DATA_COLLECTION
                        };
                        
                        select.from_clause = Some(Datasource::Collection(CollectionSource {
                            database: Some(TEST_DB.to_string()),
                            collection: collection.to_string(),
                            alias: None,
                        }));
                    }
                }
                
                make_query_semantic(&mut with.body);
                
                for query in &mut with.queries {
                    make_query_semantic(&mut query.query);
                    
                    if query.name.is_empty() {
                        query.name = format!("cte_{}", usize::arbitrary(&mut Gen::new(0)) % 100);
                    }
                }
            },
        }
    }

    #[allow(dead_code)]
    fn make_select_query_semantic(query: &mut SelectQuery) {
        let collection = if bool::arbitrary(&mut Gen::new(0)) {
            ALL_TYPES_COLLECTION
        } else {
            RELATED_DATA_COLLECTION
        };
        
        query.from_clause = Some(Datasource::Collection(CollectionSource {
            database: Some(TEST_DB.to_string()),
            collection: collection.to_string(),
            alias: None,
        }));

        match &mut query.select_clause.body {
            SelectBody::Standard(exprs) => {
                if exprs.is_empty() {
                    exprs.push(SelectExpression::Expression(OptionallyAliasedExpr::Unaliased(
                        make_numeric_expression()
                    )));
                }
                
                for expr in exprs {
                    match expr {
                        SelectExpression::Star => {},
                        SelectExpression::Substar(substar) => {
                            if substar.datasource.is_empty() || 
                               !substar.datasource.chars().all(|c| c.is_ascii_alphanumeric() || c == '_') {
                                substar.datasource = match usize::arbitrary(&mut Gen::new(0)) % 4 {
                                    0 => ALL_TYPES_COLLECTION.to_string(),
                                    1 => RELATED_DATA_COLLECTION.to_string(),
                                    2 => NUMERIC_COLLECTION.to_string(),
                                    _ => ARRAY_COLLECTION.to_string(),
                                };
                            }
                        },
                        SelectExpression::Expression(opt_aliased) => {
                            match opt_aliased {
                                OptionallyAliasedExpr::Aliased(aliased) => {
                                    make_expression_semantic(&mut aliased.expr);
                                },
                                OptionallyAliasedExpr::Unaliased(expr) => {
                                    make_expression_semantic(expr);
                                },
                            }
                        },
                    }
                }
            },
            SelectBody::Values(values) => {
                if values.is_empty() {
                    values.push(SelectValuesExpression::Expression(make_numeric_expression()));
                }
                
                for value in values {
                    match value {
                        SelectValuesExpression::Expression(expr) => {
                            make_expression_semantic(expr);
                        },
                        SelectValuesExpression::Substar(substar) => {
                            if substar.datasource.is_empty() || 
                               !substar.datasource.chars().all(|c| c.is_ascii_alphanumeric() || c == '_') {
                                substar.datasource = match usize::arbitrary(&mut Gen::new(0)) % 4 {
                                    0 => ALL_TYPES_COLLECTION.to_string(),
                                    1 => RELATED_DATA_COLLECTION.to_string(),
                                    2 => NUMERIC_COLLECTION.to_string(),
                                    _ => ARRAY_COLLECTION.to_string(),
                                };
                            }
                        }
                    }
                }
            }
        }

        if let Some(expr) = &mut query.where_clause {
            make_expression_semantic(expr);
        }

        if let Some(group_by) = &mut query.group_by_clause {
            for key in &mut group_by.keys {
                match key {
                    OptionallyAliasedExpr::Aliased(aliased) => {
                        make_expression_semantic(&mut aliased.expr);
                    },
                    OptionallyAliasedExpr::Unaliased(expr) => {
                        make_expression_semantic(expr);
                    },
                }
            }
            
            for agg in &mut group_by.aggregations {
                make_expression_semantic(&mut agg.expr);
            }
        }

        if let Some(expr) = &mut query.having_clause {
            make_expression_semantic(expr);
        }

        if let Some(order_by) = &mut query.order_by_clause {
            let mut valid_identifiers = Vec::new();
            match &query.select_clause.body {
                SelectBody::Standard(exprs) => {
                    for expr in exprs {
                        match expr {
                            SelectExpression::Expression(opt_aliased) => {
                                match opt_aliased {
                                    OptionallyAliasedExpr::Aliased(aliased) => {
                                        valid_identifiers.push(Expression::Identifier(aliased.alias.clone()));
                                    },
                                    OptionallyAliasedExpr::Unaliased(expr) => {
                                        if let Expression::Identifier(id) = expr {
                                            valid_identifiers.push(Expression::Identifier(id.clone()));
                                        }
                                    },
                                }
                            },
                            _ => {}, // Skip Star and Substar expressions
                        }
                    }
                },
                SelectBody::Values(values) => {
                    for value in values {
                        if let SelectValuesExpression::Expression(Expression::Identifier(id)) = value {
                            valid_identifiers.push(Expression::Identifier(id.clone()));
                        }
                    }
                }
            }
            
            if valid_identifiers.is_empty() {
                valid_identifiers.push(Expression::Identifier("_id".to_string()));
            }
            
            for sort_spec in &mut order_by.sort_specs {
                match &mut sort_spec.key {
                    SortKey::Simple(expr) => {
                        make_expression_semantic(expr);
                    },
                    SortKey::Positional(pos) => {
                        let idx = (*pos as usize) % valid_identifiers.len();
                        let identifier = valid_identifiers[idx].clone();
                        sort_spec.key = SortKey::Simple(identifier);
                    }
                }
            }
        }

        if query.limit.is_some() {
            query.limit = Some(10); // Use a reasonable limit
        }
        
        if query.offset.is_some() {
            query.offset = Some(0); // Use a reasonable offset
        }
    }

    // Generate a numeric expression (Int32, Int64, Double, Decimal128)
    fn make_numeric_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 8 {
            0 => Expression::Identifier(INT_FIELD.to_string()),
            1 => Expression::Identifier(LONG_FIELD.to_string()),
            2 => Expression::Identifier(DOUBLE_FIELD.to_string()),
            3 => Expression::Identifier(DECIMAL_FIELD.to_string()),
            4 => Expression::Literal(Literal::Integer(42)),
            5 => Expression::Literal(Literal::Integer(-10)),
            6 => Expression::Literal(Literal::Long(1000000)),
            _ => Expression::Literal(Literal::Double(std::f64::consts::PI)),
        }
    }
    
    fn make_boolean_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 4 {
            0 => Expression::Identifier(BOOL_FIELD.to_string()),
            1 => Expression::Identifier(TRUE_FIELD.to_string()),
            2 => Expression::Identifier(FALSE_FIELD.to_string()),
            _ => Expression::Literal(Literal::Boolean(bool::arbitrary(&mut Gen::new(0)))),
        }
    }
    
    fn make_string_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 3 {
            0 => Expression::Identifier(STRING_FIELD.to_string()),
            1 => Expression::Identifier(EMPTY_STRING_FIELD.to_string()),
            _ => Expression::Identifier(DESCRIPTION_FIELD.to_string()),
        }
    }
    
    fn make_array_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 3 {
            0 => Expression::Identifier(ARRAY_FIELD.to_string()),
            1 => Expression::Identifier(STRING_ARRAY_FIELD.to_string()),
            _ => Expression::Identifier(MIXED_ARRAY_FIELD.to_string()),
        }
    }
    
    fn make_date_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 3 {
            0 => Expression::Identifier(DATE_FIELD.to_string()),
            1 => Expression::Identifier(TIMESTAMP_FIELD.to_string()),
            _ => Expression::Identifier(TIME_FIELD.to_string()),
        }
    }
    
    #[allow(dead_code)]
    fn make_object_expression() -> Expression {
        match usize::arbitrary(&mut Gen::new(0)) % 2 {
            0 => Expression::Identifier(OBJECT_FIELD.to_string()),
            _ => Expression::Identifier(NESTED_OBJECT_FIELD.to_string()),
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
    
    #[allow(dead_code)]
    fn make_expression_semantic(expr: &mut Expression) {
        match expr {
            Expression::Identifier(_) => {
                let collection = match usize::arbitrary(&mut Gen::new(0)) % 4 {
                    0 => ALL_TYPES_COLLECTION,
                    1 => RELATED_DATA_COLLECTION,
                    2 => NUMERIC_COLLECTION,
                    _ => ARRAY_COLLECTION,
                };
                
                let field = match collection {
                    ALL_TYPES_COLLECTION => {
                        let fields = [
                            INT_FIELD, LONG_FIELD, DOUBLE_FIELD, DECIMAL_FIELD,
                            NEGATIVE_INT_FIELD, ZERO_INT_FIELD, STRING_FIELD, 
                            EMPTY_STRING_FIELD, BOOL_FIELD, TRUE_FIELD, FALSE_FIELD,
                            DATE_FIELD, TIMESTAMP_FIELD, TIME_FIELD, OBJECT_FIELD,
                            NESTED_OBJECT_FIELD, ARRAY_FIELD, STRING_ARRAY_FIELD,
                            MIXED_ARRAY_FIELD, NULL_FIELD, OBJECTID_FIELD
                        ];
                        fields[usize::arbitrary(&mut Gen::new(0)) % fields.len()]
                    },
                    RELATED_DATA_COLLECTION => {
                        let fields = [ID_FIELD, ALL_TYPES_ID_FIELD, DESCRIPTION_FIELD];
                        fields[usize::arbitrary(&mut Gen::new(0)) % fields.len()]
                    },
                    NUMERIC_COLLECTION => {
                        let fields = ["id", "int_value", "long_value", "double_value", "decimal_value", "calculated_field"];
                        fields[usize::arbitrary(&mut Gen::new(0)) % fields.len()]
                    },
                    _ => { // ARRAY_COLLECTION
                        let fields = ["id", "int_array", "string_array", "object_array", "nested_array"];
                        fields[usize::arbitrary(&mut Gen::new(0)) % fields.len()]
                    }
                };
                
                *expr = Expression::Identifier(field.to_string());
            },
            Expression::Binary(binary) => {
                make_expression_semantic(&mut binary.left);
                make_expression_semantic(&mut binary.right);
                
                // Generate a more diverse set of binary operations
                let op = match usize::arbitrary(&mut Gen::new(0)) % 8 {
                    0 => BinaryOp::Add,
                    1 => BinaryOp::Sub,
                    2 => BinaryOp::Mul,
                    3 => BinaryOp::Div,
                    4 => BinaryOp::And,
                    5 => BinaryOp::Or,
                    6 => BinaryOp::Concat,
                    _ => BinaryOp::Comparison(ComparisonOp::Eq),
                };
                
                binary.op = op;
                
                match op {
                    BinaryOp::Add | BinaryOp::Sub | BinaryOp::Mul | BinaryOp::Div => {
                        // Ensure numeric operands for arithmetic operations
                        *binary.left = make_numeric_expression();
                        *binary.right = make_numeric_expression();
                    },
                    BinaryOp::And | BinaryOp::Or => {
                        // Ensure boolean operands for logical operations
                        *binary.left = make_boolean_expression();
                        *binary.right = make_boolean_expression();
                    },
                    BinaryOp::Concat => {
                        *binary.left = make_string_expression();
                        *binary.right = make_string_expression();
                    },
                    BinaryOp::In | BinaryOp::NotIn => {
                        *binary.right = make_array_expression();
                        *binary.left = make_numeric_expression();
                    },
                    BinaryOp::Comparison(comp_op) => {
                        let left_type = expression_type(&binary.left);
                        let right_type = expression_type(&binary.right);
                        
                        if !are_types_compatible(left_type, right_type) {
                            match comp_op {
                                ComparisonOp::Eq | ComparisonOp::Neq => {
                                    *binary.left = make_numeric_expression();
                                    *binary.right = make_numeric_expression();
                                },
                                ComparisonOp::Lt | ComparisonOp::Lte | 
                                ComparisonOp::Gt | ComparisonOp::Gte => {
                                    *binary.left = make_numeric_expression();
                                    *binary.right = make_numeric_expression();
                                }
                            }
                        }
                    }
                }
            },
            Expression::Unary(unary) => {
                make_expression_semantic(&mut unary.expr);
                
                let op = match usize::arbitrary(&mut Gen::new(0)) % 3 {
                    0 => UnaryOp::Not,
                    1 => UnaryOp::Neg,
                    _ => UnaryOp::Pos,
                };
                
                unary.op = op;
                
                match op {
                    UnaryOp::Not => {
                        *unary.expr = make_boolean_expression();
                    },
                    UnaryOp::Neg | UnaryOp::Pos => {
                        *unary.expr = make_numeric_expression();
                    },
                }
            },
            Expression::Cast(cast) => {
                make_expression_semantic(&mut cast.expr);
                
                let source_type = expression_type(&cast.expr);
                
                cast.to = match source_type {
                    Type::Int32 | Type::Int64 | Type::Double | Type::Decimal128 => {
                        match usize::arbitrary(&mut Gen::new(0)) % 4 {
                            0 => Type::Int32,
                            1 => Type::Int64,
                            2 => Type::Double,
                            _ => Type::Decimal128,
                        }
                    },
                    Type::String => {
                        Type::Int32
                    },
                    Type::Boolean => {
                        Type::Int32
                    },
                    _ => {
                        Type::Int32
                    }
                };
            },
            Expression::Case(case) => {
                if let Some(expr) = &mut case.expr {
                    make_expression_semantic(expr);
                }
                
                if case.when_branch.is_empty() {
                    case.when_branch.push(WhenBranch {
                        when: Box::new(make_boolean_expression()),
                        then: Box::new(make_numeric_expression()),
                    });
                }
                
                for branch in &mut case.when_branch {
                    *branch.when = make_boolean_expression();
                    
                    make_expression_semantic(&mut branch.then);
                }
                
                if let Some(expr) = &mut case.else_branch {
                    make_expression_semantic(expr);
                    
                    if !case.when_branch.is_empty() {
                        let then_type = expression_type(&case.when_branch[0].then);
                        let else_type = expression_type(expr);
                        
                        if !are_types_compatible(then_type, else_type) {
                            match then_type {
                                Type::Int32 | Type::Int64 | Type::Double | Type::Decimal128 => {
                                    *expr = Box::new(make_numeric_expression());
                                },
                                Type::Boolean => {
                                    *expr = Box::new(make_boolean_expression());
                                },
                                Type::String => {
                                    *expr = Box::new(make_string_expression());
                                },
                                _ => {
                                    *expr = Box::new(make_numeric_expression());
                                }
                            }
                        }
                    }
                }
            },
            Expression::Literal(lit) => {
                *lit = match usize::arbitrary(&mut Gen::new(0)) % 6 {
                    0 => Literal::Integer(42),
                    1 => Literal::Integer(-10),
                    2 => Literal::Long(1000000),
                    3 => Literal::Double(std::f64::consts::PI),
                    4 => Literal::Boolean(bool::arbitrary(&mut Gen::new(0))),
                    _ => Literal::Null,
                };
            },
            Expression::Array(array) => {
                if array.is_empty() {
                    array.push(make_numeric_expression());
                }
                
                for elem in &mut *array {
                    make_expression_semantic(elem);
                }
                
                if !array.is_empty() {
                    let first_type = expression_type(&array[0]);
                    for elem in array.iter_mut().skip(1) {
                        let elem_type = expression_type(elem);
                        if !are_types_compatible(first_type, elem_type) {
                            match first_type {
                                Type::Int32 | Type::Int64 | Type::Double | Type::Decimal128 => {
                                    *elem = make_numeric_expression();
                                },
                                Type::Boolean => {
                                    *elem = make_boolean_expression();
                                },
                                Type::String => {
                                    *elem = make_string_expression();
                                },
                                _ => {
                                    *elem = make_numeric_expression();
                                }
                            }
                        }
                    }
                }
            },
            Expression::StringConstructor(_str_constructor) => {
                *expr = make_string_expression();
            },
            Expression::Function(func) => {
                if let FunctionArguments::Args(args) = &mut func.args {
                    for arg in &mut *args {
                        make_expression_semantic(arg);
                    }
                    
                    if !args.is_empty() {
                        match func.function {
                            FunctionName::Split | FunctionName::LTrim | FunctionName::RTrim => {
                                args[0] = make_string_expression();
                            },
                            FunctionName::Sum | FunctionName::Avg | FunctionName::Min | FunctionName::Max => {
                                args[0] = make_numeric_expression();
                            },
                            _ => {
                                args[0] = make_numeric_expression();
                            }
                        }
                    }
                }
            },
            Expression::TypeAssertion(type_assertion) => {
                make_expression_semantic(&mut type_assertion.expr);
            },
            Expression::Between(between) => {
                make_expression_semantic(&mut between.arg);
                make_expression_semantic(&mut between.min);
                make_expression_semantic(&mut between.max);
                
                *between.arg = make_numeric_expression();
                *between.min = make_numeric_expression();
                *between.max = make_numeric_expression();
            },
            Expression::Tuple(_) => {
                *expr = make_numeric_expression();
            },
            Expression::Trim(trim) => {
                make_expression_semantic(&mut trim.arg);
                *trim.arg = make_string_expression();
            },
            Expression::Is(is_expr) => {
                make_expression_semantic(&mut is_expr.expr);
                
                match is_expr.target_type {
                    TypeOrMissing::Missing => {
                    },
                    TypeOrMissing::Number => {
                        is_expr.expr = Box::new(make_numeric_expression());
                    },
                    TypeOrMissing::Type(typ) => {
                        match typ {
                            Type::Int32 | Type::Int64 | Type::Double | Type::Decimal128 => {
                                is_expr.expr = Box::new(make_numeric_expression());
                            },
                            Type::String => {
                                is_expr.expr = Box::new(make_string_expression());
                            },
                            Type::Boolean => {
                                is_expr.expr = Box::new(make_boolean_expression());
                            },
                            Type::Date | Type::Timestamp | Type::Time => {
                                is_expr.expr = Box::new(make_date_expression());
                            },
                            Type::Array => {
                                is_expr.expr = Box::new(make_array_expression());
                            },
                            _ => {
                            }
                        }
                    }
                }
            },
            Expression::Extract(extract) => {
                make_expression_semantic(&mut extract.arg);
                *extract.arg = make_date_expression();
                
                extract.extract_spec = match usize::arbitrary(&mut Gen::new(0)) % 7 {
                    0 => DatePart::Year,
                    1 => DatePart::Month,
                    2 => DatePart::Day,
                    3 => DatePart::Hour,
                    4 => DatePart::Minute,
                    5 => DatePart::Second,
                    _ => DatePart::Millisecond,
                };
            },
            Expression::Subpath(subpath) => {
                make_expression_semantic(&mut subpath.expr);
                
                if !matches!(*subpath.expr, Expression::Identifier(_) | Expression::Document(_)) {
                    *subpath.expr = Expression::Identifier(INT_FIELD.to_string());
                }
                
                if subpath.subpath.is_empty() || !subpath.subpath.chars().all(|c| c.is_ascii_alphanumeric() || c == '_') {
                    subpath.subpath = INT_FIELD.to_string();
                }
            },
            _ => {
                *expr = Expression::Identifier(INT_FIELD.to_string());
            }
        }
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
                _ => None, // Default for other expression types
            }
        }
        
        #[allow(dead_code)]
        fn visit_expression(&mut self, node: Expression) -> Expression {
            if self.target_type.is_none() {
                return node.walk(self);
            }
            
            let node_type = expression_type(&node);
            let target_type = self.target_type.unwrap();
            
            let node = if node_type != target_type && !are_types_compatible(node_type, target_type) {
                match target_type {
                    Type::Boolean => make_boolean_expression(),
                    Type::Int32 | Type::Int64 | Type::Double | Type::Decimal128 => make_numeric_expression(),
                    Type::String => make_string_expression(),
                    Type::Array => make_array_expression(),
                    Type::Date | Type::Datetime | Type::Timestamp => make_date_expression(),
                    Type::Document => make_object_expression(),
                    _ => node, // Keep the original node for other types
                }
            } else {
                node
            };
            
            let child_target_type = self.determine_child_target_type(&node);
            
            let old_target_type = self.target_type;
            self.target_type = child_target_type;
            let new_node = node.walk(self);
            self.target_type = old_target_type;
            
            new_node
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
        
        #[allow(dead_code)]
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
            if self.target_type.is_none() {
                return;
            }
            
            let node_type = expression_type(node);
            let target_type = self.target_type.unwrap();
            
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
                _ => {}
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
