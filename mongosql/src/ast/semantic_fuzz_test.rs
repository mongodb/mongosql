#[cfg(test)]
mod tests {
    use crate::{
        ast::{
            definitions::*,
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

    const INT_FIELD: &str = "int_field";
    const LONG_FIELD: &str = "long_field";
    const DOUBLE_FIELD: &str = "double_field";
    const DECIMAL_FIELD: &str = "decimal_field";
    const STRING_FIELD: &str = "string_field";
    const BOOL_FIELD: &str = "bool_field";
    const DATE_FIELD: &str = "date_field";
    const OBJECT_FIELD: &str = "object_field";
    const ARRAY_FIELD: &str = "array_field";
    const NULL_FIELD: &str = "null_field";
    const ID_FIELD: &str = "id";
    const ALL_TYPES_ID_FIELD: &str = "all_types_id";
    const DESCRIPTION_FIELD: &str = "description";

    fn make_query_semantic(query: &mut Query) {
        match query {
            Query::Select(select) => make_select_query_semantic(select),
            Query::Set(set) => {
                make_query_semantic(set.left.as_mut());
                make_query_semantic(set.right.as_mut());
            },
            Query::With(with) => {
                make_query_semantic(&mut with.body);
                for query in &mut with.queries {
                    make_query_semantic(&mut query.query);
                }
            },
        }
    }

    fn make_select_query_semantic(query: &mut SelectQuery) {
        if query.from_clause.is_some() {
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
        }

        if let SelectBody::Standard(exprs) = &mut query.select_clause.body {
            for expr in exprs {
                match expr {
                    SelectExpression::Star => {},
                    SelectExpression::Substar(substar) => {
                        substar.datasource = if bool::arbitrary(&mut Gen::new(0)) {
                            ALL_TYPES_COLLECTION.to_string()
                        } else {
                            RELATED_DATA_COLLECTION.to_string()
                        };
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
            for sort_spec in &mut order_by.sort_specs {
                if let SortKey::Simple(expr) = &mut sort_spec.key {
                    make_expression_semantic(expr);
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

    fn make_expression_semantic(expr: &mut Expression) {
        match expr {
            Expression::Identifier(_) => {
                let collection = if bool::arbitrary(&mut Gen::new(0)) {
                    ALL_TYPES_COLLECTION
                } else {
                    RELATED_DATA_COLLECTION
                };
                
                let field = match collection {
                    ALL_TYPES_COLLECTION => {
                        let fields = [
                            INT_FIELD, LONG_FIELD, DOUBLE_FIELD, DECIMAL_FIELD,
                            STRING_FIELD, BOOL_FIELD, DATE_FIELD, OBJECT_FIELD,
                            ARRAY_FIELD, NULL_FIELD
                        ];
                        fields[usize::arbitrary(&mut Gen::new(0)) % fields.len()]
                    },
                    _ => {
                        let fields = [ID_FIELD, ALL_TYPES_ID_FIELD, DESCRIPTION_FIELD];
                        fields[usize::arbitrary(&mut Gen::new(0)) % fields.len()]
                    }
                };
                
                *expr = Expression::Identifier(field.to_string());
            },
            Expression::Binary(binary) => {
                make_expression_semantic(&mut binary.left);
                make_expression_semantic(&mut binary.right);
                
                binary.op = match usize::arbitrary(&mut Gen::new(0)) % 3 {
                    0 => BinaryOp::Add,
                    1 => BinaryOp::And,
                    _ => BinaryOp::Or,
                };
            },
            Expression::Unary(unary) => {
                make_expression_semantic(&mut unary.expr);
                unary.op = UnaryOp::Not; // Only use Not as it's definitely supported
            },
            Expression::Function(func) => {
                if let FunctionArguments::Args(args) = &mut func.args {
                    for arg in args {
                        make_expression_semantic(arg);
                    }
                }
                
                func.function = FunctionName::Count;
            },
            Expression::Cast(cast) => {
                make_expression_semantic(&mut cast.expr);
                
                cast.to = Type::Int32;
            },
            Expression::Case(case) => {
                if let Some(expr) = &mut case.expr {
                    make_expression_semantic(expr);
                }
                
                for branch in &mut case.when_branch {
                    make_expression_semantic(&mut branch.when);
                    make_expression_semantic(&mut branch.then);
                }
                
                if let Some(expr) = &mut case.else_branch {
                    make_expression_semantic(expr);
                }
            },
            Expression::Literal(lit) => {
                *lit = match usize::arbitrary(&mut Gen::new(0)) % 4 {
                    0 => Literal::Integer(42),
                    1 => Literal::Double(std::f64::consts::PI),
                    2 => Literal::Boolean(true),
                    _ => Literal::Null,
                };
            },
            _ => {
                *expr = Expression::Identifier(INT_FIELD.to_string());
            }
        }
    }

    #[test]
    fn prop_semantic_queries_translate() {
        fn property(mut query: Query) -> TestResult {
            make_query_semantic(&mut query);
            
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
            make_query_semantic(&mut query);
            
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
                        "string_field": { "bsonType": "string" },
                        "bool_field": { "bsonType": "bool" },
                        "date_field": { "bsonType": "date" },
                        "object_field": { 
                            "bsonType": "object",
                            "properties": {
                                "nested_field": { "bsonType": "string" }
                            }
                        },
                        "array_field": { 
                            "bsonType": "array",
                            "items": { "bsonType": "int" }
                        },
                        "null_field": { "bsonType": "null" }
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
            
            catalog_schema.insert("test_db".to_string(), db_schema);
            build_catalog_from_catalog_schema(catalog_schema).unwrap()
        };
    }
}
