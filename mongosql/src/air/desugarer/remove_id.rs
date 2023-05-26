use crate::air::{
    self,
    desugarer::{Pass, Result},
    visitor::Visitor,
    Project, ProjectItem,
};

/// This removes _id using the last Project stage in the pipeline
#[derive(Default)]
pub struct RemoveIdDesugarerPass;

impl Pass for RemoveIdDesugarerPass {
    fn apply(&self, pipeline: air::Stage) -> Result<air::Stage> {
        let mut visitor = RemoveIdDesugarerVisitor::default();
        Ok(pipeline.walk(&mut visitor))
    }
}

#[derive(Default)]
struct RemoveIdDesugarerVisitor {}

impl Visitor for RemoveIdDesugarerVisitor {
    fn visit_project(&mut self, mut node: Project) -> Project {
        let id = "_id".to_string();
        if !node.specifications.contains_key(&id) {
            // A duplicate key is impossible here
            node.specifications
                .insert(id, ProjectItem::Exclusion)
                .unwrap();
        }
        node
    }
}
