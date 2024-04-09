/// Optimizes Join stages by converting them to EquiJoin stages when possible.
/// If a Join stage's right source is a collection source and the condition is
/// a simple equality comparison of fields (one from the left and one from the
/// right), then the Join can be converted into an EquiJoin.
///
/// Note that the intent here is to enable the translator and codegen to produce
/// an equijoin-style $lookup stage in the final pipeline. This style $lookup
/// provides favorable runtime over expressive $lookup.
///
/// The equijoin syntax does not provide the SQL-style null semantics MongoSQL
/// requires, so we must ensure those semantics are preserved as part of this
/// optimization. This is why the optimization checks the schema of the local
/// and foreign fields and adjusts the conditions appropriately. In order:
///   - If the schema indicates either the local or the foreign field MUST be
///     null or missing, instead of replacing the Join with an EquiJoin, we just
///     replace the Join condition with a literal false value.
///   - If the schema indicates either the local or the foreign field must NOT
///     be null or missing, no filter is needed.
///   - Otherwise, the schema indicates both MAY be null or missing. Therefore,
///     a Filter stage that asserts the existence of the local field is placed
///     before the EquiJoin stage.
///
#[cfg(test)]
mod test;

use crate::{
    mir::{
        optimizer::Optimizer,
        schema::{CachedSchema, SchemaCache, SchemaInferenceState},
        visitor::Visitor,
        EquiJoin, Expression, FieldAccess, FieldPath, Filter, MQLStage, ScalarFunction,
        ScalarFunctionApplication, Stage,
    },
    schema::ResultSet,
    SchemaCheckingMode,
};

pub(crate) struct JoinSemanticsOptimizer;

impl Optimizer for JoinSemanticsOptimizer {
    fn optimize(
        &self,
        st: Stage,
        _sm: SchemaCheckingMode,
        schema_state: &SchemaInferenceState,
    ) -> (Stage, bool) {
        let mut v = JoinSemanticsOptimizerVisitor {
            schema_state,
            changed: false,
        };
        let new_stage = v.visit_stage(st);
        (new_stage, v.changed)
    }
}

struct JoinSemanticsOptimizerVisitor<'a> {
    schema_state: &'a SchemaInferenceState<'a>,
    changed: bool,
}

impl<'a> JoinSemanticsOptimizerVisitor<'a> {
    /// get_if_resembles_collection_source gets the 'right' field for the
    /// EquiJoin _if_ the Join's right stage "resembles a collection source".
    /// The argument, stage, must be the candidate Join stage's right source.
    ///
    /// "resembles a collection source" means that the stage is a Project with
    /// a Collection source. The Project stage should emit one key-value pair,
    /// where the value is a Reference expression referring to the name of the
    /// Collection stage.
    fn resembles_collection_source(&self, stage: &Stage) -> bool {
        match stage {
            Stage::Project(p) => match p.source.as_ref() {
                Stage::Collection(_) => {
                    // Note that we operate here on the assumption that the
                    // algebrizer is correct. If we produced a Project with
                    // len 1 expression, and with a Collection source, then
                    // it must be that the single binding is a Reference to
                    // the Collection.
                    p.expression.len() == 1
                }
                // If the Project's source is not a Collection, the Join is ineligible for
                // this rewrite.
                _ => false,
            },
            // If the stage is not a Project, the Join is ineligible for this rewrite.
            _ => false,
        }
    }

    /// get_fields_if_valid_condition gets the local and foreign fields for
    /// the EquiJoin if the Join's condition is valid. The argument, condition,
    /// must be the candidate Join stage's condition expression.
    ///
    /// A "valid" condition is an Equals ScalarFunction between two field
    /// paths. One path must come from the LHS of the Join and one from the
    /// RHS.
    ///
    /// If the argument matches that structure, the fields are returned
    /// as FieldPaths. The first is the local, the second the foreign.
    fn get_fields_if_valid_condition(
        &self,
        condition: &Expression,
        left_schema: ResultSet,
        right_schema: ResultSet,
    ) -> Option<(FieldPath, FieldPath)> {
        match condition {
            Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Eq,
                is_nullable: _,
                args,
            }) if args.len() == 2 => {
                let arg1: FieldPath = match args.first().unwrap().try_into() {
                    Ok(fp) => fp,
                    Err(_) => return None,
                };
                let arg2: FieldPath = match args.get(1).unwrap().try_into() {
                    Ok(fp) => fp,
                    Err(_) => return None,
                };

                let left_contains_arg1 = left_schema.contains_field(&arg1.key, &arg1.fields);
                let left_contains_arg2 = left_schema.contains_field(&arg2.key, &arg2.fields);
                let right_contains_arg1 = right_schema.contains_field(&arg1.key, &arg1.fields);
                let right_contains_arg2 = right_schema.contains_field(&arg2.key, &arg2.fields);

                // A valid result is when one argument is exclusive to the
                // left schema, and the other argument is exclusive to the
                // right schema. In that case the left argument is the local
                // field and the right is the foreign field.
                if left_contains_arg1
                    && !left_contains_arg2
                    && !right_contains_arg1
                    && right_contains_arg2
                {
                    Some((arg1, arg2))
                } else if !left_contains_arg1
                    && left_contains_arg2
                    && right_contains_arg1
                    && !right_contains_arg2
                {
                    Some((arg2, arg1))
                } else {
                    None
                }
            }
            // If the expression is not an Equality comparison of 2 arguments,
            // the Join is ineligible for this rewrite.
            _ => None,
        }
    }
}

impl<'a> Visitor for JoinSemanticsOptimizerVisitor<'a> {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);

        match &node {
            Stage::Join(j) => {
                let condition = match &j.condition {
                    // If there is no condition, we cannot rewrite.
                    None => return node,
                    Some(condition) => condition,
                };
                if self.resembles_collection_source(&j.right) {
                    let left_schema = j.left.schema(self.schema_state).unwrap();
                    let right_schema = j.right.schema(self.schema_state).unwrap();
                    match self.get_fields_if_valid_condition(
                        condition,
                        left_schema.clone(),
                        right_schema.clone(),
                    ) {
                        // If the condition is not an equality of pure FieldAccess expressions
                        // -- one from the left and one from the right -- we cannot rewrite.
                        None => node,
                        Some((local_field, foreign_field)) => {
                            self.changed = true;
                            if local_field.is_nullable && foreign_field.is_nullable {
                                Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
                                    join_type: j.join_type,
                                    source: Box::new(Stage::Filter(Filter {
                                        source: j.left.clone(),
                                        condition: Expression::MQLIntrinsicFieldExistence(
                                            local_field.clone().into(),
                                        ),
                                        cache: SchemaCache::new(),
                                    })),
                                    from: j.right.clone(),
                                    local_field: Box::new(local_field),
                                    foreign_field: Box::new(foreign_field),
                                    cache: SchemaCache::new(),
                                }))
                            } else {
                                Stage::MQLIntrinsic(MQLStage::EquiJoin(EquiJoin {
                                    join_type: j.join_type,
                                    source: j.left.clone(),
                                    from: j.right.clone(),
                                    local_field: Box::new(local_field),
                                    foreign_field: Box::new(foreign_field),
                                    cache: SchemaCache::new(),
                                }))
                            }
                        }
                    }
                } else {
                    node
                }
            }
            _ => node,
        }
    }
}

impl From<FieldPath> for FieldAccess {
    fn from(value: FieldPath) -> Self {
        let mut fields = value.fields.clone();
        let first = fields.pop();
        let mut ret = FieldAccess {
            expr: Box::new(Expression::Reference(value.key.into())),
            field: first.unwrap(),
            is_nullable: value.is_nullable,
        };

        for field in fields {
            ret = FieldAccess {
                expr: Box::new(Expression::FieldAccess(ret)),
                field,
                is_nullable: value.is_nullable,
            }
        }
        ret
    }
}
