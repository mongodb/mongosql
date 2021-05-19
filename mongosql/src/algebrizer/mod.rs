#[cfg(test)]
mod test;

use crate::{
    ast,
    ir::{
        self,
        binding_tuple::BindingTuple,
        schema::{self, SchemaInferenceState},
    },
    schema::SchemaEnvironment,
};
use thiserror::Error;

type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("all SELECT queries must have a FROM clause")]
    NoFromClause,
    #[error("standard SELECT bodies are invalid except for SELECT *")]
    NonStarStandardSelectBody,
    #[error("collection datasources must have aliases")]
    CollectionMustHaveAlias,
    #[error(transparent)]
    SchemaChecking(#[from] schema::Error),
}

pub struct Algebrizer {
    current_db: String,
    schema_env: SchemaEnvironment,
    scope_level: u16,
}

impl Algebrizer {
    pub fn new(current_db: String, scope_level: u16) -> Self {
        Self {
            current_db,
            schema_env: SchemaEnvironment::default(),
            scope_level,
        }
    }

    pub fn schema_inference_state(&self) -> SchemaInferenceState {
        SchemaInferenceState {
            env: self.schema_env.clone(),
            scope_level: self.scope_level,
        }
    }

    pub fn algebrize_query(&self, ast_node: ast::Query) -> Result<ir::Stage> {
        match ast_node {
            ast::Query::Select(q) => self.algebrize_select_query(q),
            ast::Query::Set(_) => unimplemented!(),
        }
    }

    pub(crate) fn algebrize_select_query(&self, ast_node: ast::SelectQuery) -> Result<ir::Stage> {
        let from_ast = ast_node.from_clause.ok_or(Error::NoFromClause)?;
        let plan = self.algebrize_from_clause(from_ast)?;
        let plan = self.algebrize_select_clause(ast_node.select_clause, plan)?;

        Ok(plan)
    }

    pub(crate) fn algebrize_from_clause(&self, ast_node: ast::Datasource) -> Result<ir::Stage> {
        match ast_node {
            ast::Datasource::Array(_) => unimplemented!(),
            ast::Datasource::Collection(c) => {
                let src = ir::Stage::Collection(ir::Collection {
                    db: c.database.unwrap_or_else(|| self.current_db.clone()),
                    collection: c.collection.clone(),
                });
                let stage = match c.alias {
                    Some(alias) => {
                        let mut expr_map: BindingTuple<ir::Expression> = BindingTuple::new();
                        expr_map.insert(
                            (alias, self.scope_level).into(),
                            ir::Expression::Reference((c.collection, self.scope_level).into()),
                        );
                        ir::Stage::Project(ir::Project {
                            source: Box::new(src),
                            expression: expr_map,
                        })
                    }
                    None => return Err(Error::CollectionMustHaveAlias),
                };
                stage.schema(&self.schema_inference_state())?;
                Ok(stage)
            }
            ast::Datasource::Derived(_) => unimplemented!(),
            ast::Datasource::Join(_) => unimplemented!(),
        }
    }

    pub(crate) fn algebrize_select_clause(
        &self,
        ast_node: ast::SelectClause,
        source: ir::Stage,
    ) -> Result<ir::Stage> {
        match ast_node.set_quantifier {
            ast::SetQuantifier::All => (),
            ast::SetQuantifier::Distinct => unimplemented!(),
        };

        match ast_node.body {
            ast::SelectBody::Values(_) => unimplemented!(),
            ast::SelectBody::Standard(exprs) => match exprs.as_slice() {
                [ast::SelectExpression::Star] => Ok(source),
                _ => Err(Error::NonStarStandardSelectBody),
            },
        }
    }
}
