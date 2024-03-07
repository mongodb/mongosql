use crate::{
    catalog::{Catalog, Namespace},
    map,
    mir::{
        optimizer::{determine_join_semantics::JoinSemanticsOptimizer, Optimizer},
        schema::{CachedSchema, SchemaCache, SchemaInferenceState},
        *,
    },
    schema::{Atomic, Document, Schema, SchemaEnvironment, INTEGER_OR_NULLISH},
    set, unchecked_unique_linked_hash_map,
    util::{mir_field_access, mir_field_path, mir_project_collection},
    SchemaCheckingMode,
};
use lazy_static::lazy_static;

lazy_static! {
    static ref CATALOG: Catalog = Catalog::new(map! {
        Namespace {db: "test_db".to_string(), collection: "local".to_string()} => Schema::Document(Document {
            keys: map! {
                "not_null".to_string() => Schema::Atomic(Atomic::Integer),
                "may_be_null".to_string() => INTEGER_OR_NULLISH.clone(),
            },
            required: set! {"not_null".to_string()},
            additional_properties: false,
            ..Default::default()
            }),
        Namespace {db: "test_db".to_string(), collection: "foreign".to_string()} => Schema::Document(Document {
            keys: map! {
                "not_null".to_string() => Schema::Atomic(Atomic::Integer),
                "may_be_null".to_string() => INTEGER_OR_NULLISH.clone(),
            },
            required: set! {"not_null".to_string()},
            additional_properties: false,
            ..Default::default()
            }),
        Namespace {db: "test_db".to_string(), collection: "other".to_string()} => Schema::Document(Document {
            keys: map! {
                "x".to_string() => Schema::Atomic(Atomic::Integer),
            },
            required: set! {"x".to_string()},
            additional_properties: false,
            ..Default::default()
            }),
    });
}

macro_rules! test_determine_join_semantics {
    ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
        #[test]
        fn $func_name() {
            let input = $input;
            let expected = $expected;

            let state = SchemaInferenceState::new(
                0u16,
                SchemaEnvironment::default(),
                &*CATALOG,
                SchemaCheckingMode::Relaxed,
            );

            // Manually set the schema_env since this artificial test
            // environment doesn't actually set it in the State.
            let res = input.schema(&state).unwrap();
            let state = SchemaInferenceState::new(
                0u16,
                res.schema_env,
                &*CATALOG,
                SchemaCheckingMode::Relaxed,
            );

            let optimizer = &JoinSemanticsOptimizer;
            let (actual, _) = optimizer.optimize(input, SchemaCheckingMode::Relaxed, &state);
            assert_eq!(expected, actual);
        }
    };
}

macro_rules! test_determine_join_semantics_no_op {
    ($func_name:ident, $input:expr) => {
        test_determine_join_semantics! { $func_name, expected = $input, input = $input }
    };
}

fn make_standard_join(condition: Option<Expression>) -> Stage {
    Stage::Join(Join {
        join_type: JoinType::Inner,
        left: mir_project_collection(None, "local", None, None),
        right: mir_project_collection(None, "foreign", None, None),
        condition,
        cache: SchemaCache::new(),
    })
}

fn make_equality_condition(arg1: Expression, arg2: Expression) -> Expression {
    Expression::ScalarFunction(ScalarFunctionApplication::new(
        ScalarFunction::Eq,
        vec![arg1, arg2],
    ))
}

mod do_not_change {
    use super::*;

    test_determine_join_semantics_no_op!(
        right_not_collection,
        Stage::Join(Join {
            join_type: JoinType::Inner,
            left: mir_project_collection(None, "local", None, None),
            right: Box::new(Stage::Array(ArraySource {
                array: vec![Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::Literal(LiteralValue::Integer(5))
                    }
                    .into()
                )],
                alias: "arr".to_string(),
                cache: SchemaCache::new(),
            })),
            condition: Some(make_equality_condition(
                *mir_field_access("local", "may_be_null", true),
                *mir_field_access("arr", "a", false),
            )),
            cache: SchemaCache::new(),
        })
    );

    test_determine_join_semantics_no_op!(no_condition, make_standard_join(None));

    test_determine_join_semantics_no_op!(
        equality_condition_with_pure_local_arg_and_impure_other,
        make_standard_join(Some(make_equality_condition(
            *mir_field_access("local", "may_be_null", true),
            Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::Literal(LiteralValue::Integer(10)),
                    }
                    .into()
                )),
                field: "a".to_string(),
                is_nullable: false,
            })
        )))
    );

    test_determine_join_semantics_no_op!(
        equality_condition_with_pure_foreign_arg_and_impure_other,
        make_standard_join(Some(make_equality_condition(
            *mir_field_access("foreign", "may_be_null", true),
            Expression::FieldAccess(FieldAccess {
                expr: Box::new(Expression::Document(
                    unchecked_unique_linked_hash_map! {
                        "a".into() => Expression::Literal(LiteralValue::Integer(10)),
                    }
                    .into()
                )),
                field: "a".to_string(),
                is_nullable: true,
            })
        )))
    );

    test_determine_join_semantics_no_op!(
        equality_condition_with_non_field_access_args,
        make_standard_join(Some(make_equality_condition(
            Expression::Literal(LiteralValue::Integer(1)),
            Expression::Literal(LiteralValue::Integer(2)),
        )))
    );

    test_determine_join_semantics_no_op!(
        non_equality_condition,
        make_standard_join(Some(Expression::ScalarFunction(
            ScalarFunctionApplication {
                function: ScalarFunction::Lt,
                args: vec![
                    *mir_field_access("local", "may_be_null", true),
                    *mir_field_access("foreign", "may_be_null", true),
                ],
                is_nullable: true,
            }
        )))
    );

    test_determine_join_semantics_no_op!(
        equality_condition_with_both_fields_from_local,
        make_standard_join(Some(Expression::ScalarFunction(
            ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                args: vec![
                    *mir_field_access("local", "may_be_null", true),
                    *mir_field_access("local", "not_null", false),
                ],
                is_nullable: true,
            }
        )))
    );
}

mod change {
    use super::*;

    test_determine_join_semantics!(
        when_local_must_not_be_nullable_change_without_special_filter,
        expected = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: mir_project_collection(None, "local", None, None),
            from: mir_project_collection(None, "foreign", None, None),
            local_field: Box::new(mir_field_path("local", vec!["not_null"])),
            foreign_field: Box::new(mir_field_path("foreign", vec!["may_be_null"])),
            cache: SchemaCache::new(),
        })),
        input = make_standard_join(Some(make_equality_condition(
            *mir_field_access("local", "not_null", false),
            *mir_field_access("foreign", "may_be_null", true),
        )))
    );

    test_determine_join_semantics!(
        when_foreign_must_not_be_nullable_change_without_special_filter,
        expected = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: mir_project_collection(None, "local", None, None),
            from: mir_project_collection(None, "foreign", None, None),
            local_field: Box::new(mir_field_path("local", vec!["may_be_null"])),
            foreign_field: Box::new(mir_field_path("foreign", vec!["not_null"])),
            cache: SchemaCache::new(),
        })),
        input = make_standard_join(Some(make_equality_condition(
            *mir_field_access("local", "may_be_null", true),
            *mir_field_access("foreign", "not_null", false),
        )))
    );

    test_determine_join_semantics!(
        local_and_foreign_can_appear_on_either_side_of_condition,
        expected = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: mir_project_collection(None, "local", None, None),
            from: mir_project_collection(None, "foreign", None, None),
            local_field: Box::new(mir_field_path("local", vec!["not_null"])),
            foreign_field: Box::new(mir_field_path("foreign", vec!["not_null"])),
            cache: SchemaCache::new(),
        })),
        input = make_standard_join(Some(make_equality_condition(
            *mir_field_access("foreign", "not_null", false),
            *mir_field_access("local", "not_null", false),
        )))
    );

    test_determine_join_semantics!(
        when_both_may_be_nullable_change_with_special_filter,
        expected = Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
            join_type: JoinType::Inner,
            source: Box::new(Stage::Filter(Filter {
                source: mir_project_collection(None, "local", None, None),
                condition: Expression::MQLIntrinsicFieldExistence(FieldAccess::new(
                    Box::new(Expression::Reference(("local", 0u16).into())),
                    "may_be_null".to_string(),
                )),
                cache: SchemaCache::new(),
            })),
            from: mir_project_collection(None, "foreign", None, None),
            local_field: Box::new(mir_field_path("local", vec!["may_be_null"])),
            foreign_field: Box::new(mir_field_path("foreign", vec!["may_be_null"])),
            cache: SchemaCache::new(),
        })),
        input = make_standard_join(Some(make_equality_condition(
            *mir_field_access("local", "may_be_null", true),
            *mir_field_access("foreign", "may_be_null", true),
        )))
    );
}
