use crate::ast::{
    self,
    rewrites::{Pass, Result},
    visitor::Visitor,
};

pub struct HigherOrderFunctionsRewritePass;

impl Pass for HigherOrderFunctionsRewritePass {
    fn apply(&self, query: ast::Query) -> Result<ast::Query> {
        let mut func_alias_visitor = HigherOrderFunctionsAliasVisitor;
        let query = query.walk(&mut func_alias_visitor);

        let mut func_arg_visitor = FunctionArgumentVisitor;
        let query = query.walk(&mut func_arg_visitor);

        Ok(query)
    }
}

struct HigherOrderFunctionsAliasVisitor;

// SQL-3297: Implement HigherOrderFunctionsAliasVisitor
impl Visitor for HigherOrderFunctionsAliasVisitor {}

struct FunctionArgumentVisitor;

// SQL-3296: Implement FunctionArgumentVisitor
impl Visitor for FunctionArgumentVisitor {}
