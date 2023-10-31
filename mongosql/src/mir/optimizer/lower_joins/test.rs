use crate::{
    catalog::Catalog,
    mir::{
        optimizer::{lower_joins::LowerJoinsOptimizer, Optimizer},
        schema::{SchemaCache, SchemaCheckingMode, SchemaInferenceState},
        Expression, Filter, Join, JoinType, LateralJoin, MQLStage, ScalarFunction,
        ScalarFunctionApplication, Stage,
    },
    schema::SchemaEnvironment,
    util::{mir_field_access, mir_project_collection},
};

macro_rules! test_lower_joins {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let input = $input;
            let expected = $expected;

            let cat = Catalog::default();
            let state = SchemaInferenceState::new(
                0u16,
                SchemaEnvironment::default(),
                &cat,
                SchemaCheckingMode::Relaxed,
            );

            let optimizer = &LowerJoinsOptimizer;
            let actual = optimizer.optimize(input, SchemaCheckingMode::Relaxed, &state);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_lower_joins_no_op {
    ($func_name:ident, $input:expr) => {
        test_lower_joins! { $func_name, expected = $input, input = $input }
    };
}

test_lower_joins_no_op!(
    do_not_rewrite_if_no_condition,
    Stage::Join(Join {
        join_type: JoinType::Inner,
        left: mir_project_collection(None, "foo", None, None),
        right: mir_project_collection(None, "bar", None, None),
        condition: None,
        cache: SchemaCache::new(),
    })
);

test_lower_joins!(
    rewrite_if_condition,
    expected = Stage::MQLIntrinsic(MQLStage::LateralJoin(LateralJoin {
        join_type: JoinType::Inner,
        source: mir_project_collection(None, "foo", None, None),
        subquery: Box::new(Stage::Filter(Filter {
            source: mir_project_collection(None, "bar", None, None),
            condition: Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    *mir_field_access("foo", "a", false),
                    *mir_field_access("bar", "b", false)
                ],
                cache: SchemaCache::new(),
                is_nullable: false,
            }),
            cache: SchemaCache::new(),
        })),
        cache: SchemaCache::new(),
    })),
    input = Stage::Join(Join {
        join_type: JoinType::Inner,
        left: mir_project_collection(None, "foo", None, None),
        right: mir_project_collection(None, "bar", None, None),
        condition: Some(Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::Eq,
            args: vec![
                *mir_field_access("foo", "a", false),
                *mir_field_access("bar", "b", false)
            ],
            cache: SchemaCache::new(),
            is_nullable: false,
        })),
        cache: SchemaCache::new(),
    })
);
