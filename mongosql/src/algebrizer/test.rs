macro_rules! test_algebrize {
    ($func_name:ident, $method:ident, $expected:expr, $ast:expr $(,)?) => {
        #[test]
        fn $func_name() {
            use crate::{
                algebrizer::{Algebrizer, Error},
                ast, ir,
            };
            let algebrizer = Algebrizer::new("test".into(), 0u16);
            let expected: Result<ir::Stage, Error> = $expected;
            let res: Result<ir::Stage, Error> = algebrizer.$method($ast);
            assert_eq!(expected, res);
        }
    };
}

mod from_clause {
    use crate::map;

    test_algebrize!(
        collection_must_have_alias,
        algebrize_from_clause,
        Err(Error::CollectionMustHaveAlias),
        ast::Datasource::Collection(ast::CollectionSource {
            database: None,
            collection: "foo".into(),
            alias: None,
        }),
    );
    test_algebrize!(
        basic_collection,
        algebrize_from_clause,
        Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir::Stage::Collection(ir::Collection {
                db: "test".into(),
                collection: "foo".into(),
            })),
            expression: map! {
                ("bar", 0u16).into() =>
                    ir::Expression::Reference(("foo", 0u16).into())
            }
        })),
        ast::Datasource::Collection(ast::CollectionSource {
            database: None,
            collection: "foo".into(),
            alias: Some("bar".into()),
        }),
    );
    test_algebrize!(
        qualified_collection,
        algebrize_from_clause,
        Ok(ir::Stage::Project(ir::Project {
            source: Box::new(ir::Stage::Collection(ir::Collection {
                db: "test2".into(),
                collection: "foo".into(),
            })),
            expression: map! {
                ("bar", 0u16).into() =>
                    ir::Expression::Reference(("foo", 0u16).into())
            }
        })),
        ast::Datasource::Collection(ast::CollectionSource {
            database: Some("test2".into()),
            collection: "foo".into(),
            alias: Some("bar".into()),
        }),
    );
}
