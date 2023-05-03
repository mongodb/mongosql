use crate::{
    air::{
        desugarer::{Pass, Result},
        visitor::Visitor,
        Expression,
        Expression::*,
        Join, JoinType, Lookup, MQLOperator, MQLSemanticOperator, Match, Project, ReplaceWith,
        Stage,
        Stage::*,
        Unwind,
    },
    map,
};
use linked_hash_map::LinkedHashMap;

/// Desugars any Join stages in the pipeline into sequences of equivalent,
/// existing MQL stages. Specifically, a Join is desugared into a sequence
/// of Lookup, Unwind, ReplaceWith, and Project.
pub struct JoinDesugarerPass;

impl Pass for JoinDesugarerPass {
    fn apply(&self, pipeline: Stage) -> Result<Stage> {
        let mut visitor = JoinDesugarerPassVisitor::default();
        Ok(visitor.visit_stage(pipeline))
    }
}

#[derive(Default)]
struct JoinDesugarerPassVisitor;

impl JoinDesugarerPassVisitor {
    /// desugar_join_stage desugars a Join stage into a sequence of Lookup,
    /// Unwind, ReplaceWith, and Project stages. Specifically, in MQL, it
    /// turns:
    ///
    ///    { $join: {
    ///        database: <db name>,
    ///        collection: <coll name>,
    ///        joinType: <join type>,
    ///        let: <let doc>,
    ///        pipeline: <pipeline>,
    ///        condition: <condition>
    ///    }}
    ///
    /// into:
    ///
    ///    { $lookup: {
    ///        from: {
    ///            db: <db name>,
    ///            coll: <coll name>
    ///        },
    ///        let: <let doc>,
    ///        pipeline: [<pipeline>..., <condition>],
    ///        as: <as name>
    ///    }},
    ///    { $unwind: {
    ///        path: "$<as name>",
    ///        preserveNullAndEmptyArrays: <join type> == "left"
    ///    }},
    ///    { $replaceWith: {
    ///        $mergeObject: ["$$ROOT", "$<as name>"]
    ///    }},
    ///    { $project: { _id: 0, <as name>: 0 } }
    ///
    /// Note that in the $unwind stage, preserveNullAndEmptyArrays is
    /// true when the joinType is "left" and false when it is "inner".
    ///
    /// Also note that the database, collection, let, and condition fields
    /// are all optional.
    fn desugar_join_stage(&self, join: Join) -> Stage {
        let lookup_pipeline = match join.condition {
            None => join.right,
            Some(expr) => Box::new(Match(Match {
                source: join.right,
                expr: Box::new(expr),
            })),
        };

        // Arbitrary unique field name. This is highly unlikely to conflict
        // with user data.
        let as_var_name = "eca58228-b657-498a-b76e-f48a9161a404".to_string();

        let lookup = Lookup(Lookup {
            source: join.left,
            let_vars: join.let_vars,
            pipeline: lookup_pipeline,
            as_var: as_var_name.clone(),
        });

        let unwind = Unwind(Unwind {
            source: Box::new(lookup),
            path: Box::new(FieldRef(as_var_name.clone().into())),
            index: None,
            outer: join.join_type == JoinType::Left,
        });

        let replace_with = ReplaceWith(ReplaceWith {
            source: Box::new(unwind),
            new_root: Box::new(MQLSemanticOperator(MQLSemanticOperator {
                op: MQLOperator::MergeObjects,
                args: vec![
                    Variable("ROOT".to_string().into()),
                    FieldRef(as_var_name.clone().into()),
                ],
            })),
        });

        let remove_var = Variable("REMOVE".to_string().into());

        let specs: LinkedHashMap<String, Expression> = map! {
            "_id".to_string() => remove_var.clone(),
            as_var_name => remove_var,
        };

        Project(Project {
            source: Box::new(replace_with),
            specifications: specs.into(),
        })
    }
}

impl Visitor for JoinDesugarerPassVisitor {
    fn visit_stage(&mut self, node: Stage) -> Stage {
        let node = node.walk(self);
        match node {
            Join(j) => self.desugar_join_stage(j),
            _ => node,
        }
    }
}
