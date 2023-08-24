use crate::air::{
    desugarer::{Pass, Result},
    visitor::Visitor,
    Expression,
    Expression::*,
    Let, LetVariable, Limit, LiteralValue, Lookup, MQLOperator, MQLSemanticOperator, Project,
    ProjectItem, Reduce, SQLOperator, SQLSemanticOperator, Stage,
    Stage::*,
    Subquery, SubqueryComparison, SubqueryComparisonOp, SubqueryComparisonOpType, SubqueryExists,
    SubqueryModifier,
};
use linked_hash_map::LinkedHashMap;

/// Desugars any top-level subquery expressions (Subquery, SubqueryComparison,
/// and SubqueryExists). A subquery expression is desugared into three parts:
///   1. A Lookup stage that is placed before the containing stage. This
///      Lookup stage performs the subquery when the pipeline is executed.
///   2. An expression that replaces the subquery expression in the containing
///      stage. This replacement is dependent on the type of subquery expr.
///   3. A Project stage that is placed after the containing stage. This
///      Project removes the field introduced by the Lookup; if there are
///      multiple subquery expressions in a stage, there will be multiple
///      Lookup stages but only one Project stage at the end. If the subquery
///      is in a Group stage, we don't want to exclude the _id because it is
///      the group key.
pub struct SubqueryExprDesugarerPass;

impl Pass for SubqueryExprDesugarerPass {
    fn apply(&self, pipeline: Stage) -> Result<Stage> {
        // Subquery and SubqueryExists benefit from appending a limit of 1 since
        // for the former only the first result is used and for the latter the
        // existence of at least one result is all that is necessary.
        let subquery_limit_adder = &mut SubqueryLimitAdder::default();
        let pipeline = subquery_limit_adder.visit_stage(pipeline);

        let mut visitor = SubqueryExprDesugarerPassVisitor {
            subquery_counter: 0,
            as_names: vec![],
            subquery_lookups: pipeline.get_source(),
        };
        Ok(visitor.visit_stage(pipeline))
    }
}

/// A SubqueryComparison expr can desugar using SQL or MQL semantic operators, depending on
/// the value of its op_type field. This macro allows us to parametrize the actual desugarer
/// code by air types: SQLOperator vs MQLOperator, Expression::SQLSemanticOperator vs
/// Expression::MQLSemanticOperator, and SQLSemanticOperator vs MQLSemanticOperator.
macro_rules! subquery_comparison_semantic_desugarer {
    ($subquery_comp:ident, $as_name:ident, op_enum = $op_enum:ident, expr_variant = $expr_variant:ident, expr_struct = $expr_struct:ident) => {{
        let (initial_value, combinator_op) = match $subquery_comp.modifier {
            SubqueryModifier::Any => (Literal(LiteralValue::Boolean(false)), $op_enum::Or),
            SubqueryModifier::All => (Literal(LiteralValue::Boolean(true)), $op_enum::And),
        };

        let comp_op = match $subquery_comp.op {
            SubqueryComparisonOp::Lt => $op_enum::Lt,
            SubqueryComparisonOp::Lte => $op_enum::Lte,
            SubqueryComparisonOp::Neq => $op_enum::Ne,
            SubqueryComparisonOp::Eq => $op_enum::Eq,
            SubqueryComparisonOp::Gt => $op_enum::Gt,
            SubqueryComparisonOp::Gte => $op_enum::Gte,
        };

        let output_path = format!("this.{}", $subquery_comp.subquery.output_path.join(".")).into();

        // Return the replacement expression.
        Reduce(Reduce {
            input: Box::new(FieldRef($as_name.into())),
            init_value: Box::new(initial_value),
            inside: Box::new($expr_variant($expr_struct {
                op: combinator_op,
                args: vec![
                    Variable("value".to_string().into()),
                    $expr_variant($expr_struct {
                        op: comp_op,
                        args: vec![*$subquery_comp.arg, Variable(output_path)],
                    }),
                ],
            })),
        })
    }};
}
#[derive(PartialEq, Debug, Clone)]
pub struct SentinelReplacer {
    pub new_source: Box<Stage>,
}

impl Visitor for SentinelReplacer {
    fn visit_stage(
        &mut self,
        node: crate::air::definitions::Stage,
    ) -> crate::air::definitions::Stage {
        match node {
            Sentinel => *self.new_source.clone(),
            _ => node.walk(self),
        }
    }
}

struct SubqueryExprDesugarerPassVisitor {
    // subquery_counter tracks how many subquery
    // expressions are in a stage
    subquery_counter: u8,

    // as_names tracks the Lookup "as" names used
    // by the subquery lookup stages for the subquery
    // expressions in a stage
    as_names: Vec<String>,

    // subquery_lookups stores the new source for a stage
    // that contains subquery expressions. Since a subquery
    // expression is partly desugared into a Lookup stage,
    // this field will always start as the source of the
    // visited stage. As subquery expressions are encountered,
    // this field will update to store a new Lookup with the
    // previous value set as the Lookup source.
    subquery_lookups: Box<Stage>,
}

impl SubqueryExprDesugarerPassVisitor {
    /// desugar_subquery desugars a Subquery expression into a Lookup and
    /// a replacement expression that accesses the data specified by the
    /// output_path. Specifically, in MQL, it turns
    ///
    ///    { $subquery: {
    ///        db: <db name>,
    ///        collection: <collection name>,
    ///        let: <let doc>,
    ///        outputPath: [<component_1>, ...],
    ///        pipeline: <pipeline>,
    ///    }}
    ///
    /// into the $lookup stage:
    ///
    ///    { $lookup: {
    ///        from: {
    ///            db: <db name>,
    ///            coll: <coll name>
    ///        },
    ///        let: <let doc>,
    ///        pipeline: <pipeline>,
    ///        as: "<asName>"
    ///    }}
    ///
    /// and the replacement expression:
    ///
    ///    { $let: {
    ///        vars: {
    ///            docExpr: {
    ///                $arrayElemAt: ["$<asName>", 0]
    ///            }
    ///        },
    ///        in: "$$docExpr.<component_1>..."
    ///    }}
    ///
    /// The $lookup stage is placed before the stage that contains the
    /// $subquery expression. The replacement expression replaces the
    /// $subquery expression in the stage that contains it.
    fn desugar_subquery(&mut self, subquery: Subquery) -> Expression {
        // Update the visitor with the desugared info.
        let as_name = self.process_subquery_expr(subquery.let_bindings, subquery.pipeline);

        let var_name = "docExpr".to_string();
        let output_path = format!("{}.{}", var_name, subquery.output_path.join(".")).into();

        // Return the replacement expression.
        Let(Let {
            vars: vec![LetVariable {
                name: var_name,
                expr: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::ElemAt,
                    args: vec![FieldRef(as_name.into()), Literal(LiteralValue::Integer(0))],
                })),
            }],
            inside: Box::new(Variable(output_path)),
        })
    }

    /// desugar_subquery_comparison desugars a SubqueryComparison expression
    /// into a Lookup and a replacement expression that performs the comparison
    /// on the subquery output. Specifically, in MQL, it turns
    ///
    ///    { $subqueryComparison: {
    ///        op: <comp op>,
    ///        modifier: <modifier>,
    ///        argument: <arg>,
    ///        subquery: <subquery>
    ///    }}
    ///
    /// into the same $lookup stage as for $subquery (above) using the <subquery>
    /// parameter (which has the same fields as $subquery), and the replacement
    /// expression:
    ///
    ///    { $reduce: {
    ///        "input": "$<asName>",
    ///        "initialValue": <initial value>,
    ///        "in": {
    ///            <combinator func>: [
    ///                "$$value",
    ///                {
    ///                    "$<comp op>": [
    ///                        <arg>,
    ///                        "$$this.<component_1>..."
    ///                    ]
    ///                }
    ///            ]
    ///        }
    ///    }}
    ///
    /// When <modifier> is "any", <initial value> is false and <combinator func>
    /// is "$sqlOr". When <modifier> is "all", <initial value> is true and
    /// <combinator func> is "$sqlAnd".
    fn desugar_subquery_comparison(&mut self, subquery_comp: SubqueryComparison) -> Expression {
        // Update the visitor with the desugared info.
        let as_name = self.process_subquery_expr(
            subquery_comp.subquery.let_bindings,
            subquery_comp.subquery.pipeline,
        );

        match subquery_comp.op_type {
            SubqueryComparisonOpType::Sql => subquery_comparison_semantic_desugarer!(
                subquery_comp,
                as_name,
                op_enum = SQLOperator,
                expr_variant = SQLSemanticOperator,
                expr_struct = SQLSemanticOperator
            ),
            SubqueryComparisonOpType::Mql => subquery_comparison_semantic_desugarer!(
                subquery_comp,
                as_name,
                op_enum = MQLOperator,
                expr_variant = MQLSemanticOperator,
                expr_struct = MQLSemanticOperator
            ),
        }
    }

    /// desugar_subquery_exists desugars a SubqueryExists expression into a
    /// Lookup and a replacement expression that checks if the output is
    /// non-empty. Specifically, in MQL, it turns
    ///
    ///    { $subqueryExists: {
    ///        db: <db name>,
    ///        collection: <collection name>,
    ///        let: <let doc>,
    ///        pipeline: <pipeline>,
    ///    }}
    ///
    /// into the same $lookup stage as for $subquery (above), and
    /// the replacement expression:
    ///
    ///    { $gt: [{ $size: "$<asName>" }, 0] }
    fn desugar_subquery_exists(&mut self, subquery_exists: SubqueryExists) -> Expression {
        // Update the visitor with the desugared info.
        let as_name =
            self.process_subquery_expr(subquery_exists.let_bindings, subquery_exists.pipeline);

        // Return the replacement expression.
        MQLSemanticOperator(MQLSemanticOperator {
            op: MQLOperator::Gt,
            args: vec![
                MQLSemanticOperator(MQLSemanticOperator {
                    op: MQLOperator::Size,
                    args: vec![FieldRef(as_name.into())],
                }),
                Literal(LiteralValue::Integer(0)),
            ],
        })
    }

    /// process_subquery_expr updates the visitor with desugared info for
    /// this subquery expression. It returns the as_name used by the
    /// desugared Lookup stage.
    fn process_subquery_expr(
        &mut self,
        let_bindings: Vec<LetVariable>,
        pipeline: Box<Stage>,
    ) -> String {
        let as_var = format!("__subquery_result_{}", self.subquery_counter);

        let let_vars = if let_bindings.is_empty() {
            None
        } else {
            Some(let_bindings)
        };

        self.subquery_counter += 1;
        self.as_names.push(as_var.clone());
        self.subquery_lookups = Box::new(Lookup(Lookup {
            source: self.subquery_lookups.clone(),
            let_vars,
            pipeline,
            as_var: as_var.clone(),
        }));

        as_var
    }

    fn update_source(&self, new_source: Box<Stage>) -> Box<Stage> {
        let mut v = SentinelReplacer { new_source };

        Box::new(v.visit_stage(*self.subquery_lookups.clone()))
    }
}

impl Visitor for SubqueryExprDesugarerPassVisitor {
    /// Although this desugarer targets Subquery, SubqueryComparison, and
    /// SubqueryExists expressions, it must overwrite visit_stage. Any
    /// subquery expression will desugar into 3 parts, as described above.
    /// The Lookup stage that performs the subquery must be prepended to
    /// the Stage that contains the subquery expression, and the Project
    /// stage must be "appended" to the Stage that contains the subquery
    /// expression. This work must be done in visit_stage.
    fn visit_stage(&mut self, stage: Stage) -> Stage {
        // Store initial values so they can be restored after this stage is
        // processed.
        let initial_subquery_counter = self.subquery_counter;
        let initial_as_names = self.as_names.clone();
        let initial_subquery_lookups = self.subquery_lookups.clone();

        // Reset values to default.
        self.subquery_counter = 0;
        self.as_names = vec![];
        self.subquery_lookups = Box::new(Sentinel);

        // Walk this stage, desugaring subquery expressions and populating the
        // visitor's fields with desugared subquery data.
        let mut stage = stage.walk(self);
        self.subquery_lookups = self.update_source(stage.get_source());

        let new_stage = if self.subquery_counter == 0 {
            // If there are no subquery expressions in this Stage, there is
            // nothing more to do.
            stage
        } else {
            // Prepend this stage with the Lookup stages that perform the
            // subqueries.
            stage.set_source(self.subquery_lookups.clone());

            // Create the Project specifications to exclude the Lookup "as"
            // names after this stage.
            let specs = self
                .as_names
                .clone()
                .into_iter()
                .map(|name| (name, ProjectItem::Exclusion))
                .collect::<LinkedHashMap<String, ProjectItem>>();

            Stage::Project(Project {
                source: Box::new(stage),
                specifications: specs.into(),
            })
        };

        // Restore initial visitor state.
        self.subquery_counter = initial_subquery_counter;
        self.as_names = initial_as_names;
        self.subquery_lookups = initial_subquery_lookups;

        new_stage
    }

    fn visit_expression(&mut self, node: Expression) -> Expression {
        let node = node.walk(self);
        match node {
            Subquery(e) => self.desugar_subquery(e),
            SubqueryComparison(e) => self.desugar_subquery_comparison(e),
            SubqueryExists(e) => self.desugar_subquery_exists(e),
            _ => node,
        }
    }
}

#[derive(Default)]
struct SubqueryLimitAdder {}

impl Visitor for SubqueryLimitAdder {
    fn visit_subquery_exists(&mut self, node: SubqueryExists) -> SubqueryExists {
        let node = node.walk(self);

        // Update the SubqueryExists pipeline to end with Limit(1) if it does
        // not already.
        if matches!(
            *node.pipeline,
            Limit(Limit {
                source: _,
                limit: 1
            })
        ) {
            node
        } else {
            SubqueryExists {
                let_bindings: node.let_bindings,
                pipeline: Box::new(Limit(Limit {
                    source: node.pipeline,
                    limit: 1,
                })),
            }
        }
    }
}
