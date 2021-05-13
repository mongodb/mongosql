macro_rules! test_codegen_plan {
    (
		$func_name:ident,
		$current_db:expr,
		Ok({
			database: $expected_db:expr,
			collection: $expected_collection:expr,
			pipeline: $expected_pipeline:expr,
		}),
		$input: expr,
	) => {
        #[test]
        fn $func_name() {
            use crate::codegen::{generate_mql, MqlTranslation};

            let current_db = $current_db.to_string();
            let input = $input;
            let expected_db = $expected_db;
            let expected_collection = $expected_collection;
            let expected_pipeline = $expected_pipeline;

            let MqlTranslation {
                database: db,
                collection: col,
                mapping_registry: _,
                pipeline: pipeline,
            } = generate_mql(current_db, input).expect("codegen failed");

            assert_eq!(expected_db, db);
            assert_eq!(expected_collection, col);
            assert_eq!(expected_pipeline, pipeline);
        }
    };

    ($func_name:ident, $current_db:expr, Err($expected_err:expr), $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::codegen::generate_mql;

            let current_db = $current_db.to_string();
            let input = $input;
            let expected = Err($expected_err);

            assert_eq!(expected, generate_mql(current_db, input));
        }
    };
}

macro_rules! test_codegen_expr {
    ($func_name:ident, $current_db:expr, $mapping_registry:expr, $expected:expr, $input:expr,) => {
        #[test]
        fn $func_name() {
            use crate::codegen::mql::MqlCodeGenerator;
            let current_database = $current_db.to_string();
            let mapping_registry = $mapping_registry;
            let expected = $expected;
            let input = $input;

            let gen = MqlCodeGenerator {
                current_database,
                mapping_registry,
            };
            assert_eq!(expected, gen.codegen_expression(input));
        }
    };

    ($func_name:ident, $expected:expr, $input:expr,) => {
        test_codegen_expr!(
            $func_name,
            "dbnameunused",
            crate::codegen::mql::MappingRegistry::default(),
            $expected,
            $input,
        );
    };

    ($func_name:ident, $mapping_registry:expr, $expected:expr, $input:expr,) => {
        test_codegen_expr!(
            $func_name,
            "dbnameunused",
            $mapping_registry,
            $expected,
            $input,
        );
    };
}

mod collection {
    use crate::ir::*;

    test_codegen_plan!(
        simple,
        "mydb",
        Ok({
            database: "mydb".to_string(),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
            ],
        }),
        Stage::Collection(Collection {
            db: "mydb".to_string(),
            collection: "col".to_string(),
        }),
    );
}

mod limit_offset {
    use crate::ir::*;

    test_codegen_plan!(
        limit_simple,
        "mydb",
        Ok({
            database: "mydb".to_string(),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$limit": 1u64},
            ],
        }),
        Stage::Limit(Limit {
            limit: 1,
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );

    test_codegen_plan!(
        offset_simple,
        "mydb",
        Ok({
            database: "mydb".to_string(),
            collection: Some("col".to_string()),
            pipeline: vec![
                bson::doc!{"$project": {"_id": 0, "col": "$$ROOT"}},
                bson::doc!{"$skip": 1u64},
            ],
        }),
        Stage::Offset(Offset {
            offset: 1,
            source: Stage::Collection(Collection {
                db: "mydb".to_string(),
                collection: "col".to_string(),
            }).into(),
        }),
    );
}

mod literal {
    use crate::ir::{Expression::*, Literal::*};
    use bson::{bson, Bson};

    test_codegen_expr!(null, Ok(bson!({ "$literal": Bson::Null })), Literal(Null),);
    test_codegen_expr!(bool, Ok(bson!({"$literal": true})), Literal(Boolean(true)),);
    test_codegen_expr!(
        string,
        Ok(bson!({"$literal": "abc"})),
        Literal(String("abc".into())),
    );
    test_codegen_expr!(int, Ok(bson!({"$literal": 5_i32})), Literal(Integer(5)),);
    test_codegen_expr!(long, Ok(bson!({"$literal": 6_i64})), Literal(Long(6)),);
    test_codegen_expr!(double, Ok(bson!({"$literal": 7.0})), Literal(Double(7.0)),);
}

mod reference {
    use crate::{
        codegen::{mql::MappingRegistry, Error},
        ir::Expression::*,
    };
    use bson::Bson;

    test_codegen_expr!(
        not_found,
        MappingRegistry::default(),
        Err(Error::ReferenceNotFound(("f", 0u16).into())),
        Reference(("f", 0u16).into()),
    );

    test_codegen_expr!(
        found,
        {
            let mut mr = MappingRegistry::default();
            mr.insert(("f", 0u16), "f");
            mr
        },
        Ok(Bson::String("$f".into())),
        Reference(("f", 0u16).into()),
    );
}

mod array {
    use crate::ir::{Expression::*, Literal};
    use bson::bson;

    test_codegen_expr!(empty, Ok(bson!([])), Array(vec![]),);
    test_codegen_expr!(
        non_empty,
        Ok(bson!([{"$literal": "abc"}])),
        Array(vec![Literal(Literal::String("abc".into()))]),
    );
    test_codegen_expr!(
        nested,
        Ok(bson!([{ "$literal": null }, [{ "$literal": null }]])),
        Array(vec![
            Literal(Literal::Null),
            Array(vec![Literal(Literal::Null)])
        ]),
    );
}

mod document {
    use crate::{
        codegen::Error,
        ir::{Expression::*, Literal},
        map,
    };
    use bson::bson;

    test_codegen_expr!(empty, Ok(bson!({"$literal": {}})), Document(map! {}),);
    test_codegen_expr!(
        non_empty,
        Ok(bson!({"foo": {"$literal": 1}})),
        Document(map! {"foo".to_string() => Literal(Literal::Integer(1)),}),
    );
    test_codegen_expr!(
        nested,
        Ok(bson!({"foo": {"$literal": 1}, "bar": {"baz": {"$literal": 2}}})),
        Document(map! {
            "foo".to_string() => Literal(Literal::Integer(1)),
            "bar".to_string() => Document(map!{
                "baz".to_string() => Literal(Literal::Integer(2))
            }),
        }),
    );
    test_codegen_expr!(
        dollar_prefixed_key_disallowed,
        Err(Error::DollarPrefixedDocumentKey),
        Document(map! {"$foo".to_string() => Literal(Literal::Integer(1)),}),
    );
    test_codegen_expr!(
        key_containing_dot_allowed,
        Ok(bson!({"foo.bar": {"$literal": 1}})),
        Document(map! {"foo.bar".to_string() => Literal(Literal::Integer(1)),}),
    );
}
