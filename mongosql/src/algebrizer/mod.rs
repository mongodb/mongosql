use crate::{
    ast,
    ir::{
        self,
        schema::{self, SchemaInferenceState},
    },
    schema::SchemaEnvironment,
};
use thiserror::Error;

type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error)]
pub enum Error {
    #[error("all SELECT queries must have a FROM clause")]
    NoFromClause,
    #[error("standard SELECT bodies are invalid except for SELECT *")]
    NonStarStandardSelectBody,
    #[error(transparent)]
    SchemaChecking(#[from] schema::Error),
}

pub struct Algebrizer {
    current_db: String,
    schema_env: SchemaEnvironment,
}

impl Algebrizer {
    pub fn new(current_db: String) -> Self {
        Self {
            current_db,
            schema_env: SchemaEnvironment::default(),
        }
    }

    fn schema_inference_state(&self) -> SchemaInferenceState<'_> {
        SchemaInferenceState::from(&self.schema_env)
    }

    pub fn algebrize_query(&self, ast_node: ast::Query) -> Result<ir::Stage> {
        match ast_node {
            ast::Query::Select(q) => self.algebrize_select_query(q),
            ast::Query::Set(_) => unimplemented!(),
        }
    }

    fn algebrize_select_query(&self, ast_node: ast::SelectQuery) -> Result<ir::Stage> {
        let from_ast = ast_node.from_clause.ok_or(Error::NoFromClause)?;
        let plan = self.algebrize_from_clause(from_ast)?;
        let plan = self.algebrize_select_clause(ast_node.select_clause, plan)?;

        Ok(plan)
    }

    fn algebrize_from_clause(&self, ast_node: ast::Datasource) -> Result<ir::Stage> {
        match ast_node {
            ast::Datasource::Array(_) => unimplemented!(),
            ast::Datasource::Collection(c) => {
                let src = ir::Stage::Collection(ir::Collection {
                    db: c.database.unwrap_or_else(|| self.current_db.clone()),
                    collection: c.collection,
                });
                let stage = match c.alias {
                    Some(_alias) => unimplemented!(),
                    None => src,
                };
                stage.schema(self.schema_inference_state())?;
                Ok(stage)
            }
            ast::Datasource::Derived(_) => unimplemented!(),
            ast::Datasource::Join(_) => unimplemented!(),
        }
    }

    fn algebrize_select_clause(
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
