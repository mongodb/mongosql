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
    util::are_literal,
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
    #[error("array datasource must be constant")]
    ArrayDatasourceMustBeLiteral,
    #[error(transparent)]
    SchemaChecking(#[from] schema::Error),
}

#[derive(Debug, Clone)]
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

    fn with_merged_mappings(mut self, mappings: SchemaEnvironment) -> Self {
        self.schema_env.merge(mappings);
        self
    }

    pub(crate) fn algebrize_select_query(&self, ast_node: ast::SelectQuery) -> Result<ir::Stage> {
        let from_ast = ast_node.from_clause.ok_or(Error::NoFromClause)?;
        let plan = self.algebrize_from_clause(from_ast)?;
        let plan = self.algebrize_select_clause(ast_node.select_clause, plan)?;
        Ok(plan)
    }

    pub(crate) fn algebrize_from_clause(&self, ast_node: ast::Datasource) -> Result<ir::Stage> {
        match ast_node {
            ast::Datasource::Array(a) => {
                let (ve, alias) = (a.array, a.alias);
                let (ve, array_is_literal) = are_literal(ve);
                if !array_is_literal {
                    return Err(Error::ArrayDatasourceMustBeLiteral);
                }
                let stage = ir::Stage::Array(ir::Array {
                    array: ve
                        .into_iter()
                        .map(|e| self.algebrize_expression(e))
                        .collect::<Result<_>>()?,
                    alias,
                });
                stage.schema(&self.schema_inference_state())?;
                Ok(stage)
            }
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
        let expression_algebrizer = self.clone();
        // Aglebrization for every node that has a source should get the schema for the source.
        // The SchemaEnvironment from the source is merged into the SchemaEnvironment from the
        // current Algebrizer, correctly giving us the the correlated bindings with the bindings
        // available from the current query level.
        #[allow(unused_variables)]
        let expression_algebrizer = expression_algebrizer
            .with_merged_mappings(source.schema(&self.schema_inference_state())?.schema_env);
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

    pub(crate) fn algebrize_expression(&self, ast_node: ast::Expression) -> Result<ir::Expression> {
        match ast_node {
            ast::Expression::Literal(l) => Ok(ir::Expression::Literal(self.algebrize_literal(l)?)),
            ast::Expression::Array(a) => Ok(ir::Expression::Array(
                a.into_iter()
                    .map(|e| self.algebrize_expression(e))
                    .collect::<Result<_>>()?,
            )),
            ast::Expression::Document(d) => Ok(ir::Expression::Document(
                d.into_iter()
                    .map(|(k, e)| Ok((k, self.algebrize_expression(e)?)))
                    .collect::<Result<_>>()?,
            )),
            ast::Expression::Unary(_)
            | ast::Expression::Identifier(_)
            | ast::Expression::Binary(_)
            | ast::Expression::Between(_)
            | ast::Expression::Case(_)
            | ast::Expression::Function(_)
            | ast::Expression::Cast(_)
            | ast::Expression::Subquery(_)
            | ast::Expression::SubqueryComparison(_)
            | ast::Expression::Exists(_)
            | ast::Expression::Access(_)
            | ast::Expression::Subpath(_)
            | ast::Expression::Is(_)
            | ast::Expression::Like(_)
            | ast::Expression::Tuple(_)
            | ast::Expression::TypeAssertion(_) => unimplemented!(),
        }
    }

    pub(crate) fn algebrize_literal(&self, ast_node: ast::Literal) -> Result<ir::Literal> {
        match ast_node {
            ast::Literal::Null => Ok(ir::Literal::Null),
            ast::Literal::Boolean(b) => Ok(ir::Literal::Boolean(b)),
            ast::Literal::String(s) => Ok(ir::Literal::String(s)),
            ast::Literal::Integer(i) => Ok(ir::Literal::Integer(i)),
            ast::Literal::Long(l) => Ok(ir::Literal::Long(l)),
            ast::Literal::Double(d) => Ok(ir::Literal::Double(d)),
        }
    }
}
