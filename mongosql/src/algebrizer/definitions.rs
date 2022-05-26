use crate::{
    ast::{self, pretty_print::PrettyPrint},
    catalog::Catalog,
    ir::{
        self,
        binding_tuple::{BindingTuple, DatasourceName, Key},
        schema::{CachedSchema, SchemaCache, SchemaInferenceState},
    },
    map,
    schema::{self, Satisfaction, SchemaEnvironment},
    util::unique_linked_hash_map::UniqueLinkedHashMap,
    SchemaCheckingMode,
};
use std::collections::BTreeSet;
use thiserror::Error;

type Result<T> = std::result::Result<T, Error>;

macro_rules! schema_check_return {
    ($self:ident, $e:expr $(,)?) => {{
        let ret = $e;
        ret.schema(&$self.schema_inference_state())?;
        return Ok(ret);
    }};
}

#[derive(Debug, Error, PartialEq)]
pub enum Error {
    #[error("ADD_TO_SET should be removed before try_from")]
    AddToSetDoesNotExistInIr,
    #[error("all SELECT queries must have a FROM clause")]
    NoFromClause,
    #[error("standard SELECT expressions can only contain *")]
    NonStarStandardSelectBody,
    #[error("collection datasources must have aliases")]
    CollectionMustHaveAlias,
    #[error("array datasource must be constant")]
    ArrayDatasourceMustBeLiteral,
    #[error("SELECT DISTINCT not allowed")]
    DistinctSelect,
    #[error("UNION DISTINCT not allowed")]
    DistinctUnion,
    #[error("no such datasource: {0:?}")]
    NoSuchDatasource(DatasourceName),
    #[error("field `{0}` cannot be resolved to any datasource")]
    FieldNotFound(String),
    #[error("ambiguous field `{0}`")]
    AmbiguousField(String),
    #[error("* argument only valid in COUNT function")]
    StarInNonCount,
    #[error("aggregation function {0} used in scalar position")]
    AggregationInPlaceOfScalar(String),
    #[error("scalar function {0} used in aggregation position")]
    ScalarInPlaceOfAggregation(String),
    #[error(
        "non-aggregation expression found in GROUP BY aggregation function list at position {0}"
    )]
    NonAggregationInPlaceOfAggregation(usize),
    #[error("aggregation functions must have exactly one argument")]
    AggregationFunctionMustHaveOneArgument,
    #[error("scalar functions cannot be DISTINCT")]
    DistinctScalarFunction,
    #[error("derived table datasources {2:?} have overlapping keys, schemata: {0:?} and {1:?}")]
    DerivedDatasouceOverlappingKeys(crate::schema::Schema, crate::schema::Schema, Satisfaction),
    #[error("{0} cannot be algebrized")]
    CannotBeAlgebrized(&'static str),
    #[error(transparent)]
    SchemaChecking(#[from] ir::schema::Error),
    #[error("OUTER JOINs must specify a JOIN condition")]
    NoOuterJoinCondition,
    #[error("cannot create schema environment with duplicate key: {0:?}")]
    DuplicateKey(Key),
    #[error("positional sort keys should have been rewritten to references")]
    PositionalSortKey,
    #[error("subquery expressions must have a degree of 1")]
    InvalidSubqueryDegree,
    #[error("found duplicate document key {0:?}")]
    DuplicateDocumentKey(String),
    #[error("found duplicate FLATTEN option {0:?}")]
    DuplicateFlattenOption(ast::FlattenOption),
    #[error("cannot exhaustively enumerate all field paths in schema {0:?}")]
    CannotEnumerateAllFieldPaths(crate::schema::Schema),
    #[error("cannot flatten field {0:?} since it has a polymorphic object schema")]
    PolymorphicObjectSchema(String),
    #[error("found duplicate UNWIND option {0:?}")]
    DuplicateUnwindOption(ast::UnwindOption),
    #[error("UNWIND must specify a PATH option")]
    NoUnwindPath,
    #[error("UNWIND PATH option must be an identifier or compound identifier")]
    InvalidUnwindPath,
}

impl TryFrom<ast::BinaryOp> for ir::ScalarFunction {
    type Error = Error;

    fn try_from(op: crate::ast::BinaryOp) -> Result<Self> {
        Ok(match op {
            ast::BinaryOp::Add => ir::ScalarFunction::Add,
            ast::BinaryOp::And => ir::ScalarFunction::And,
            ast::BinaryOp::Concat => ir::ScalarFunction::Concat,
            ast::BinaryOp::Div => ir::ScalarFunction::Div,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Eq) => ir::ScalarFunction::Eq,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Gt) => ir::ScalarFunction::Gt,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Gte) => ir::ScalarFunction::Gte,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Lt) => ir::ScalarFunction::Lt,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Lte) => ir::ScalarFunction::Lte,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Neq) => ir::ScalarFunction::Neq,
            ast::BinaryOp::Mul => ir::ScalarFunction::Mul,
            ast::BinaryOp::Or => ir::ScalarFunction::Or,
            ast::BinaryOp::Sub => ir::ScalarFunction::Sub,
            ast::BinaryOp::In | ast::BinaryOp::NotIn => {
                return Err(Error::CannotBeAlgebrized(op.as_str()))
            }
        })
    }
}

impl TryFrom<ast::FunctionName> for ir::ScalarFunction {
    type Error = Error;

    fn try_from(f: crate::ast::FunctionName) -> Result<Self> {
        Ok(match f {
            ast::FunctionName::BitLength => ir::ScalarFunction::BitLength,
            ast::FunctionName::CharLength => ir::ScalarFunction::CharLength,
            ast::FunctionName::Coalesce => ir::ScalarFunction::Coalesce,
            ast::FunctionName::CurrentTimestamp => ir::ScalarFunction::CurrentTimestamp,
            ast::FunctionName::Log => ir::ScalarFunction::Log,
            ast::FunctionName::Lower => ir::ScalarFunction::Lower,
            ast::FunctionName::NullIf => ir::ScalarFunction::NullIf,
            ast::FunctionName::OctetLength => ir::ScalarFunction::OctetLength,
            ast::FunctionName::Position => ir::ScalarFunction::Position,
            ast::FunctionName::Round => ir::ScalarFunction::Round,
            ast::FunctionName::Size => ir::ScalarFunction::Size,
            ast::FunctionName::Slice => ir::ScalarFunction::Slice,
            ast::FunctionName::Split => ir::ScalarFunction::Split,
            ast::FunctionName::Substring => ir::ScalarFunction::Substring,
            ast::FunctionName::Upper => ir::ScalarFunction::Upper,

            ast::FunctionName::AddToArray
            | ast::FunctionName::AddToSet
            | ast::FunctionName::Avg
            | ast::FunctionName::Count
            | ast::FunctionName::First
            | ast::FunctionName::Last
            | ast::FunctionName::Max
            | ast::FunctionName::MergeDocuments
            | ast::FunctionName::Min
            | ast::FunctionName::StddevPop
            | ast::FunctionName::StddevSamp
            | ast::FunctionName::Sum => {
                return Err(Error::AggregationInPlaceOfScalar(f.pretty_print().unwrap()))
            }
        })
    }
}

impl TryFrom<ast::FunctionName> for ir::AggregationFunction {
    type Error = Error;

    fn try_from(f: crate::ast::FunctionName) -> Result<Self> {
        Ok(match f {
            ast::FunctionName::AddToArray => ir::AggregationFunction::AddToArray,
            ast::FunctionName::AddToSet => return Err(Error::AddToSetDoesNotExistInIr),
            ast::FunctionName::Avg => ir::AggregationFunction::Avg,
            ast::FunctionName::Count => ir::AggregationFunction::Count,
            ast::FunctionName::First => ir::AggregationFunction::First,
            ast::FunctionName::Last => ir::AggregationFunction::Last,
            ast::FunctionName::Max => ir::AggregationFunction::Max,
            ast::FunctionName::MergeDocuments => ir::AggregationFunction::MergeDocuments,
            ast::FunctionName::Min => ir::AggregationFunction::Min,
            ast::FunctionName::StddevPop => ir::AggregationFunction::StddevPop,
            ast::FunctionName::StddevSamp => ir::AggregationFunction::StddevSamp,
            ast::FunctionName::Sum => ir::AggregationFunction::Sum,

            ast::FunctionName::BitLength
            | ast::FunctionName::CharLength
            | ast::FunctionName::Coalesce
            | ast::FunctionName::CurrentTimestamp
            | ast::FunctionName::Log
            | ast::FunctionName::Lower
            | ast::FunctionName::NullIf
            | ast::FunctionName::OctetLength
            | ast::FunctionName::Position
            | ast::FunctionName::Round
            | ast::FunctionName::Size
            | ast::FunctionName::Slice
            | ast::FunctionName::Split
            | ast::FunctionName::Substring
            | ast::FunctionName::Upper => {
                return Err(Error::ScalarInPlaceOfAggregation(f.pretty_print().unwrap()))
            }
        })
    }
}

impl From<crate::ast::ComparisonOp> for ir::SubqueryComparisonOp {
    fn from(op: crate::ast::ComparisonOp) -> Self {
        match op {
            ast::ComparisonOp::Eq => ir::SubqueryComparisonOp::Eq,
            ast::ComparisonOp::Gt => ir::SubqueryComparisonOp::Gt,
            ast::ComparisonOp::Gte => ir::SubqueryComparisonOp::Gte,
            ast::ComparisonOp::Lt => ir::SubqueryComparisonOp::Lt,
            ast::ComparisonOp::Lte => ir::SubqueryComparisonOp::Lte,
            ast::ComparisonOp::Neq => ir::SubqueryComparisonOp::Neq,
        }
    }
}

#[derive(Debug, Clone)]
pub struct Algebrizer<'a> {
    current_db: &'a str,
    schema_env: SchemaEnvironment,
    catalog: &'a Catalog,
    scope_level: u16,
    schema_checking_mode: SchemaCheckingMode,
}

impl<'a> Algebrizer<'a> {
    pub fn new(
        current_db: &'a str,
        catalog: &'a Catalog,
        scope_level: u16,
        schema_checking_mode: SchemaCheckingMode,
    ) -> Self {
        Self::with_schema_env(
            current_db,
            SchemaEnvironment::default(),
            catalog,
            scope_level,
            schema_checking_mode,
        )
    }

    pub fn with_schema_env(
        current_db: &'a str,
        schema_env: SchemaEnvironment,
        catalog: &'a Catalog,
        scope_level: u16,
        schema_checking_mode: SchemaCheckingMode,
    ) -> Self {
        Self {
            current_db,
            schema_env,
            catalog,
            scope_level,
            schema_checking_mode,
        }
    }

    pub fn schema_inference_state(&self) -> SchemaInferenceState {
        SchemaInferenceState {
            env: self.schema_env.clone(),
            catalog: self.catalog,
            scope_level: self.scope_level,
            schema_checking_mode: self.schema_checking_mode,
        }
    }

    pub fn subquery_algebrizer(&self) -> Self {
        Self {
            current_db: self.current_db,
            schema_env: self.schema_env.clone(),
            catalog: self.catalog,
            scope_level: self.scope_level + 1,
            schema_checking_mode: self.schema_checking_mode,
        }
    }

    pub fn algebrize_query(&self, ast_node: ast::Query) -> Result<ir::Stage> {
        match ast_node {
            ast::Query::Select(q) => self.algebrize_select_query(q),
            ast::Query::Set(s) => self.algebrize_set_query(s),
        }
    }

    fn with_merged_mappings(mut self, mappings: SchemaEnvironment) -> Result<Self> {
        self.schema_env
            .merge(mappings)
            .map_err(|e| Error::DuplicateKey(e.key))?;
        Ok(self)
    }

    pub fn algebrize_select_query(&self, ast_node: ast::SelectQuery) -> Result<ir::Stage> {
        let plan = self.algebrize_from_clause(ast_node.from_clause)?;
        let plan = self.algebrize_filter_clause(ast_node.where_clause, plan)?;
        let plan = self.algebrize_group_by_clause(ast_node.group_by_clause, plan)?;
        let plan = self.algebrize_filter_clause(ast_node.having_clause, plan)?;
        let plan = self.algebrize_select_clause(ast_node.select_clause, plan)?;
        let plan = self.algebrize_order_by_clause(ast_node.order_by_clause, plan)?;
        let plan = self.algebrize_offset_clause(ast_node.offset, plan)?;
        let plan = self.algebrize_limit_clause(ast_node.limit, plan)?;
        Ok(plan)
    }

    pub fn algebrize_set_query(&self, ast_node: ast::SetQuery) -> Result<ir::Stage> {
        match ast_node.op {
            ast::SetOperator::Union => Err(Error::DistinctUnion),
            ast::SetOperator::UnionAll => schema_check_return!(
                self,
                ir::Stage::Set(ir::Set {
                    operation: ir::SetOperation::UnionAll,
                    left: Box::new(self.algebrize_query(*ast_node.left)?),
                    right: Box::new(self.algebrize_query(*ast_node.right)?),
                    cache: SchemaCache::new(),
                })
            ),
        }
    }

    fn algebrize_select_values_body(
        &self,
        exprs: Vec<ast::SelectValuesExpression>,
        source: ir::Stage,
    ) -> Result<ir::Stage> {
        let expression_algebrizer = self.clone();
        // Algebrization for every node that has a source should get the schema for the source.
        // The SchemaEnvironment from the source is merged into the SchemaEnvironment from the
        // current Algebrizer, correctly giving us the the correlated bindings with the bindings
        // available from the current query level.
        #[allow(unused_variables)]
        let expression_algebrizer = expression_algebrizer
            .with_merged_mappings(source.schema(&self.schema_inference_state())?.schema_env)?;

        // We must check for duplicate Datasource Keys, which is an error. The datasources
        // Set keeps track of which Keys have been seen.
        let mut datasources = BTreeSet::new();
        // Build the Project expression from the SelectBody::Values(exprs)
        let expression = exprs
            .into_iter()
            .map(|expr| {
                match expr {
                    // An Expression is mapped to DatasourceName::Bottom.
                    ast::SelectValuesExpression::Expression(e) => {
                        let e = expression_algebrizer.algebrize_expression(e)?;
                        let bot = Key::bot(expression_algebrizer.scope_level);
                        datasources
                            .insert(bot.clone())
                            .then(|| ())
                            .ok_or_else(|| Error::DuplicateKey(bot.clone()))?;
                        Ok((bot, e))
                    }
                    // For a Substar, a.*, we map the name of the Substar, 'a', to a Key
                    // containing 'a' and the proper scope level.
                    ast::SelectValuesExpression::Substar(s) => {
                        let datasource = DatasourceName::Named(s.datasource.clone());
                        let key = Key {
                            datasource: datasource.clone(),
                            scope: expression_algebrizer.scope_level,
                        };
                        datasources
                            .insert(key.clone())
                            .then(|| ())
                            .ok_or_else(|| Error::DuplicateKey(key.clone()))?;
                        let scope = expression_algebrizer
                            .schema_env
                            .nearest_scope_for_datasource(
                                &datasource,
                                expression_algebrizer.scope_level,
                            )
                            .ok_or_else(|| Error::NoSuchDatasource(datasource.clone()))?;
                        Ok((
                            key,
                            ir::Expression::Reference(
                                Key {
                                    datasource: DatasourceName::Named(s.datasource),
                                    scope,
                                }
                                .into(),
                            ),
                        ))
                    }
                }
            })
            .collect::<Result<_>>()?;
        // Build the Project Stage using the source and built expression.
        let stage = ir::Stage::Project(ir::Project {
            source: Box::new(source),
            expression,
            cache: SchemaCache::new(),
        });
        stage.schema(&self.schema_inference_state())?;
        Ok(stage)
    }

    pub fn algebrize_from_clause(&self, ast_node: Option<ast::Datasource>) -> Result<ir::Stage> {
        let ast_node = ast_node.ok_or(Error::NoFromClause)?;
        self.algebrize_datasource(ast_node)
    }

    pub fn algebrize_datasource(&self, ast_node: ast::Datasource) -> Result<ir::Stage> {
        match ast_node {
            ast::Datasource::Array(a) => self.algebrize_array_datasource(a),
            ast::Datasource::Collection(c) => self.algebrize_collection_datasource(c),
            ast::Datasource::Join(j) => self.algebrize_join_datasource(j),
            ast::Datasource::Derived(d) => self.algebrize_derived_datasource(d),
            ast::Datasource::Flatten(f) => self.algebrize_flatten_datasource(f),
            ast::Datasource::Unwind(u) => self.algebrize_unwind_datasource(u),
        }
    }

    fn algebrize_array_datasource(&self, a: ast::ArraySource) -> Result<ir::Stage> {
        let (ve, alias) = (a.array, a.alias);
        let (ve, array_is_literal) = ast::visitors::are_literal(ve);
        if !array_is_literal {
            return Err(Error::ArrayDatasourceMustBeLiteral);
        }
        let stage = ir::Stage::Array(ir::ArraySource {
            array: ve
                .into_iter()
                .map(|e| self.algebrize_expression(e))
                .collect::<Result<_>>()?,
            alias,
            cache: SchemaCache::new(),
        });
        stage.schema(&self.schema_inference_state())?;
        Ok(stage)
    }

    fn algebrize_collection_datasource(&self, c: ast::CollectionSource) -> Result<ir::Stage> {
        let src = ir::Stage::Collection(ir::Collection {
            db: c.database.unwrap_or_else(|| self.current_db.to_string()),
            collection: c.collection.clone(),
            cache: SchemaCache::new(),
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
                    cache: SchemaCache::new(),
                })
            }
            None => return Err(Error::CollectionMustHaveAlias),
        };
        stage.schema(&self.schema_inference_state())?;
        Ok(stage)
    }

    fn algebrize_join_datasource(&self, j: ast::JoinSource) -> Result<ir::Stage> {
        let left_src = self.algebrize_datasource(*j.left)?;
        let right_src = self.algebrize_datasource(*j.right)?;
        let left_src_result_set = left_src.schema(&self.schema_inference_state())?;
        let right_src_result_set = right_src.schema(&self.schema_inference_state())?;
        let join_algebrizer = self
            .clone()
            .with_merged_mappings(left_src_result_set.schema_env)?
            .with_merged_mappings(right_src_result_set.schema_env)?;
        let condition = j
            .condition
            .map(|e| join_algebrizer.algebrize_expression(e))
            .transpose()?;
        condition
            .clone()
            .map(|e| e.schema(&join_algebrizer.schema_inference_state()));
        let stage = match j.join_type {
            ast::JoinType::Left => {
                if condition.is_none() {
                    return Err(Error::NoOuterJoinCondition);
                }
                ir::Stage::Join(ir::Join {
                    join_type: ir::JoinType::Left,
                    left: Box::new(left_src),
                    right: Box::new(right_src),
                    condition,
                    cache: SchemaCache::new(),
                })
            }
            ast::JoinType::Right => {
                if condition.is_none() {
                    return Err(Error::NoOuterJoinCondition);
                }
                ir::Stage::Join(ir::Join {
                    join_type: ir::JoinType::Left,
                    left: Box::new(right_src),
                    right: Box::new(left_src),
                    condition,
                    cache: SchemaCache::new(),
                })
            }
            ast::JoinType::Cross | ast::JoinType::Inner => ir::Stage::Join(ir::Join {
                join_type: ir::JoinType::Inner,
                left: Box::new(left_src),
                right: Box::new(right_src),
                condition,
                cache: SchemaCache::new(),
            }),
        };
        Ok(stage)
    }

    fn algebrize_derived_datasource(&self, d: ast::DerivedSource) -> Result<ir::Stage> {
        let derived_algebrizer = Algebrizer::new(
            self.current_db,
            self.catalog,
            self.scope_level + 1,
            self.schema_checking_mode,
        );
        let src = derived_algebrizer.algebrize_query(*d.query)?;
        let src_resultset = src.schema(&derived_algebrizer.schema_inference_state())?;
        let expression = map! {
            (d.alias, self.scope_level).into() =>
            ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::MergeObjects,
                args: src_resultset
                    .schema_env
                    .into_iter()
                    .map(|(k, _)| ir::Expression::Reference(k.into()))
                    .collect::<Vec<_>>(),
                cache: SchemaCache::new(),
            }),
        };
        let stage = ir::Stage::Project(ir::Project {
            source: Box::new(src),
            expression,
            cache: SchemaCache::new(),
        });
        stage
            .schema(&derived_algebrizer.schema_inference_state())
            .map_err(|e| match e {
                ir::schema::Error::CannotMergeObjects(s1, s2, sat) => {
                    Error::DerivedDatasouceOverlappingKeys(s1, s2, sat)
                }
                _ => Error::SchemaChecking(e),
            })?;

        Ok(ir::Stage::Derived(ir::Derived {
            source: Box::new(stage),
            cache: SchemaCache::new(),
        }))
    }

    fn algebrize_flatten_datasource(&self, f: ast::FlattenSource) -> Result<ir::Stage> {
        let source = self.algebrize_datasource(*f.datasource.clone())?;
        let source_result_set = source.schema(&self.schema_inference_state())?;

        // Extract user-specified separator and depth. Separator defaults to "_".
        let (separator, depth) = f
            .options
            .iter()
            .fold(Ok((None, None)), |acc, opt| match opt {
                ast::FlattenOption::Separator(s) => match acc? {
                    (Some(_), _) => Err(Error::DuplicateFlattenOption(opt.clone())),
                    (None, depth) => Ok((Some(s.as_str()), depth)),
                },
                ast::FlattenOption::Depth(d) => match acc? {
                    (_, Some(_)) => Err(Error::DuplicateFlattenOption(opt.clone())),
                    (separator, None) => Ok((separator, Some(*d))),
                },
            })?;
        let separator = separator.unwrap_or("_");

        // Build the Project expression
        let expression = source_result_set
            .schema_env
            .into_iter()
            .map(|(key, schema)| {
                let field_paths =
                    schema
                        .enumerate_field_paths(depth.map(|d| d + 1))
                        .map_err(|e| match e {
                            schema::Error::CannotEnumerateAllFieldPaths(s) => {
                                Error::CannotEnumerateAllFieldPaths(s)
                            }
                            _ => unreachable!(),
                        })?;
                // Error if any field path is a prefix of another path
                field_paths
                    .iter()
                    .flat_map(|p1| {
                        field_paths.iter().map(|p2| {
                            if p1.clone() == p2.clone() || p1.is_empty() || p2.is_empty() {
                                Ok(())
                            } else if p1.starts_with(p2.as_slice()) {
                                Err(Error::PolymorphicObjectSchema(p2.join(separator)))
                            } else if p2.starts_with(p1.as_slice()) {
                                Err(Error::PolymorphicObjectSchema(p1.join(separator)))
                            } else {
                                Ok(())
                            }
                        })
                    })
                    .collect::<Result<_>>()?;
                let mut project_expression = UniqueLinkedHashMap::new();
                project_expression
                    .insert_many(
                        field_paths
                            .into_iter()
                            .map(|path| {
                                (
                                    path.join(separator),
                                    self.algebrize_flattened_field_path(key.clone(), path),
                                )
                            })
                            .into_iter(),
                    )
                    .map_err(|e| Error::DuplicateDocumentKey(e.get_key_name()))?;
                Ok((key, ir::Expression::Document(project_expression.into())))
            })
            .collect::<Result<BindingTuple<ir::Expression>>>()?;

        // Build the Project stage using the source and built expression
        let stage = ir::Stage::Project(ir::Project {
            source: Box::new(source),
            expression,
            cache: SchemaCache::new(),
        });
        stage.schema(&self.schema_inference_state())?;
        Ok(stage)
    }

    fn algebrize_unwind_datasource(&self, u: ast::UnwindSource) -> Result<ir::Stage> {
        let src = self.algebrize_datasource(*u.datasource)?;

        // Extract user-specified options. OUTER defaults to false.
        let (path, index, outer) =
            u.options
                .iter()
                .fold(Ok((None, None, None)), |acc, opt| match opt {
                    ast::UnwindOption::Path(p) => match acc? {
                        (Some(_), _, _) => Err(Error::DuplicateUnwindOption(opt.clone())),
                        (None, i, o) => Ok((Some(p.clone()), i, o)),
                    },
                    ast::UnwindOption::Index(i) => match acc? {
                        (_, Some(_), _) => Err(Error::DuplicateUnwindOption(opt.clone())),
                        (p, None, o) => Ok((p, Some(i.clone()), o)),
                    },
                    ast::UnwindOption::Outer(o) => match acc? {
                        (_, _, Some(_)) => Err(Error::DuplicateUnwindOption(opt.clone())),
                        (p, i, None) => Ok((p, i, Some(*o))),
                    },
                })?;

        let path_expression_algebrizer = Algebrizer::with_schema_env(
            self.current_db,
            src.schema(&self.schema_inference_state())?.schema_env,
            self.catalog,
            self.scope_level,
            self.schema_checking_mode,
        );

        let path = match path {
            None => return Err(Error::NoUnwindPath),
            Some(e) => path_expression_algebrizer.algebrize_unwind_path(e)?,
        };

        let stage = ir::Stage::Unwind(ir::Unwind {
            source: Box::new(src),
            path: Box::new(path),
            index,
            outer: outer.unwrap_or(false),
            cache: Default::default(),
        });

        stage.schema(&self.schema_inference_state())?;
        Ok(stage)
    }

    /// Ensure an unwind PATH is exactly a compound identifier.
    ///
    /// A compound identifier is defined in the MongoSQL grammar as
    ///
    ///   <compound identifer> ::= <identifier> ("." <compound identifier>)?
    ///
    /// so this includes the case when it is just a simple, single-part
    /// identifier. Here, this means the algebrized expression is a FieldAccess
    /// expression which consists of only other FieldAccess expressions up the
    /// chain of exprs until it hits a Reference expression.
    fn algebrize_unwind_path(&self, path: ast::Expression) -> Result<ir::Expression> {
        /// Auxiliary function that recursively walks up the FieldAccess
        /// tree, ensuring each parent expression is a FieldAccess until
        /// it hits a reference.
        fn is_valid_path(e: ir::Expression) -> bool {
            match e {
                ir::Expression::Reference(_) => true,
                ir::Expression::FieldAccess(f) => is_valid_path(*f.expr),
                _ => false,
            }
        }

        let path = self.algebrize_expression(path)?;
        if is_valid_path(path.clone()) {
            if let ir::Expression::FieldAccess(_) = path {
                return Ok(path);
            }
        };

        Err(Error::InvalidUnwindPath)
    }

    pub fn algebrize_flattened_field_path(&self, key: Key, path: Vec<String>) -> ir::Expression {
        match path.len() {
            0 => ir::Expression::Reference(ir::ReferenceExpr {
                key,
                cache: SchemaCache::new(),
            }),
            _ => ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(
                    self.algebrize_flattened_field_path(key, path.split_last().unwrap().1.to_vec()),
                ),
                field: path.last().unwrap().to_string(),
                cache: SchemaCache::new(),
            }),
        }
    }

    pub fn algebrize_filter_clause(
        &self,
        ast_node: Option<ast::Expression>,
        source: ir::Stage,
    ) -> Result<ir::Stage> {
        let filtered = match ast_node {
            None => source,
            Some(expr) => {
                let expression_algebrizer = self.clone().with_merged_mappings(
                    source.schema(&self.schema_inference_state())?.schema_env,
                )?;
                ir::Stage::Filter(ir::Filter {
                    source: Box::new(source),
                    condition: expression_algebrizer.algebrize_expression(expr)?,
                    cache: SchemaCache::new(),
                })
            }
        };
        filtered.schema(&self.schema_inference_state())?;
        Ok(filtered)
    }

    pub fn algebrize_select_clause(
        &self,
        ast_node: ast::SelectClause,
        source: ir::Stage,
    ) -> Result<ir::Stage> {
        match ast_node.set_quantifier {
            ast::SetQuantifier::All => (),
            ast::SetQuantifier::Distinct => return Err(Error::DistinctSelect),
        };

        match ast_node.body {
            // Standard Select bodies must be only *, otherwise this is an
            // error.
            ast::SelectBody::Standard(exprs) => match exprs.as_slice() {
                [ast::SelectExpression::Star] => {
                    source.schema(&self.schema_inference_state())?;
                    Ok(source)
                }
                _ => Err(Error::NonStarStandardSelectBody),
            },
            // SELECT VALUES expressions must be Substar expressions or normal Expressions that are
            // Documents, i.e., that have Schema that Must satisfy ANY_DOCUMENT.
            //
            // All normal Expressions will be mapped as Datasource Bottom, and all Substars will be mapped
            // as their name as a Datasource.
            ast::SelectBody::Values(exprs) => self.algebrize_select_values_body(exprs, source),
        }
    }

    pub fn algebrize_order_by_clause(
        &self,
        ast_node: Option<ast::OrderByClause>,
        source: ir::Stage,
    ) -> Result<ir::Stage> {
        let expression_algebrizer = self
            .clone()
            .with_merged_mappings(source.schema(&self.schema_inference_state())?.schema_env)?;
        let ordered = match ast_node {
            None => source,
            Some(o) => {
                let sort_specs = o
                    .sort_specs
                    .into_iter()
                    .map(|s| {
                        let key = match s.key {
                            ast::SortKey::Simple(expr) => {
                                expression_algebrizer.algebrize_expression(expr)
                            }
                            ast::SortKey::Positional(_) => Err(Error::PositionalSortKey),
                        }?;
                        match s.direction {
                            ast::SortDirection::Asc => {
                                Ok(ir::SortSpecification::Asc(Box::new(key)))
                            }
                            ast::SortDirection::Desc => {
                                Ok(ir::SortSpecification::Desc(Box::new(key)))
                            }
                        }
                    })
                    .collect::<Result<Vec<ir::SortSpecification>>>()?;
                ir::Stage::Sort(ir::Sort {
                    source: Box::new(source),
                    specs: sort_specs,
                    cache: SchemaCache::new(),
                })
            }
        };
        ordered.schema(&self.schema_inference_state())?;
        Ok(ordered)
    }

    pub fn algebrize_group_by_clause(
        &self,
        ast_node: Option<ast::GroupByClause>,
        source: ir::Stage,
    ) -> Result<ir::Stage> {
        let grouped = match ast_node {
            None => source,
            Some(ast_expr) => {
                let expression_algebrizer = self.clone().with_merged_mappings(
                    source.schema(&self.schema_inference_state())?.schema_env,
                )?;

                let mut group_clause_aliases = UniqueLinkedHashMap::new();
                let keys = ast_expr
                    .keys
                    .into_iter()
                    .map(|ast_key| match ast_key {
                        ast::OptionallyAliasedExpr::Aliased(ast_key) => {
                            group_clause_aliases
                                .insert(ast_key.alias.clone(), ())
                                .map_err(|e| Error::DuplicateDocumentKey(e.get_key_name()))?;
                            Ok(ir::OptionallyAliasedExpr::Aliased(ir::AliasedExpr {
                                alias: ast_key.alias,
                                expr: expression_algebrizer.algebrize_expression(ast_key.expr)?,
                            }))
                        }
                        ast::OptionallyAliasedExpr::Unaliased(expr) => expression_algebrizer
                            .algebrize_expression(expr)
                            .map(ir::OptionallyAliasedExpr::Unaliased),
                    })
                    .collect::<Result<_>>()?;

                let aggregations = ast_expr
                    .aggregations
                    .into_iter()
                    .enumerate()
                    .map(|(index, ast_agg)| {
                        group_clause_aliases
                            .insert(ast_agg.alias.clone(), ())
                            .map_err(|e| Error::DuplicateDocumentKey(e.get_key_name()))?;
                        Ok(ir::AliasedAggregation {
                            agg_expr: match ast_agg.expr {
                                ast::Expression::Function(f) => {
                                    expression_algebrizer.algebrize_aggregation(f)
                                }
                                _ => Err(Error::NonAggregationInPlaceOfAggregation(index)),
                            }?,
                            alias: ast_agg.alias,
                        })
                    })
                    .collect::<Result<_>>()?;

                ir::Stage::Group(ir::Group {
                    source: Box::new(source),
                    keys,
                    aggregations,
                    cache: SchemaCache::new(),
                })
            }
        };

        grouped.schema(&self.schema_inference_state())?;
        Ok(grouped)
    }

    pub fn algebrize_aggregation(&self, f: ast::FunctionExpr) -> Result<ir::AggregationExpr> {
        let (distinct, function) = if f.function == ast::FunctionName::AddToSet {
            (true, ast::FunctionName::AddToArray)
        } else {
            (
                f.set_quantifier == Some(ast::SetQuantifier::Distinct),
                f.function,
            )
        };
        let ir_node = match f.args {
            ast::FunctionArguments::Star => {
                if f.function == ast::FunctionName::Count {
                    schema_check_return!(self, ir::AggregationExpr::CountStar(distinct),)
                }
                return Err(Error::StarInNonCount);
            }
            ast::FunctionArguments::Args(ve) => {
                ir::AggregationExpr::Function(ir::AggregationFunctionApplication {
                    function: ir::AggregationFunction::try_from(function)?,
                    arg: Box::new({
                        if ve.len() != 1 {
                            return Err(Error::AggregationFunctionMustHaveOneArgument);
                        }
                        self.algebrize_expression(ve[0].clone())?
                    }),
                    distinct,
                })
            }
        };

        schema_check_return!(self, ir_node,);
    }

    pub fn algebrize_expression(&self, ast_node: ast::Expression) -> Result<ir::Expression> {
        match ast_node {
            ast::Expression::Literal(l) => {
                Ok(ir::Expression::Literal(self.algebrize_literal(l).into()))
            }
            ast::Expression::Array(a) => Ok(ir::Expression::Array(
                a.into_iter()
                    .map(|e| self.algebrize_expression(e))
                    .collect::<Result<Vec<ir::Expression>>>()?
                    .into(),
            )),
            ast::Expression::Document(d) => Ok(ir::Expression::Document({
                let algebrized = d
                    .into_iter()
                    .map(|kv| Ok((kv.key, self.algebrize_expression(kv.value)?)))
                    .collect::<Result<Vec<_>>>()?;
                let mut out = UniqueLinkedHashMap::new();
                out.insert_many(algebrized.into_iter())
                    .map_err(|e| Error::DuplicateDocumentKey(e.get_key_name()))?;
                out.into()
            })),
            // If we ever see Identifier in algebrize_expression it must be an unqualified
            // reference, because we do not recurse on the expr field of Subpath if it is an
            // Identifier
            ast::Expression::Identifier(i) => self.algebrize_unqualified_identifier(i),
            ast::Expression::Subpath(s) => self.algebrize_subpath(s),
            ast::Expression::Unary(u) => self.algebrize_unary_expr(u),
            ast::Expression::Binary(b) => self.algebrize_binary_expr(b),
            ast::Expression::Function(f) => self.algebrize_function(f),
            ast::Expression::Between(b) => self.algebrize_between(b),
            ast::Expression::Trim(t) => self.algebrize_trim(t),
            ast::Expression::Extract(e) => self.algebrize_extract(e),
            ast::Expression::Access(a) => self.algebrize_access(a),
            ast::Expression::Case(c) => self.algebrize_case(c),
            ast::Expression::Cast(c) => self.algebrize_cast(c),
            ast::Expression::TypeAssertion(t) => self.algebrize_type_assertion(t),
            ast::Expression::Is(i) => self.algebrize_is(i),
            ast::Expression::Like(l) => self.algebrize_like(l),
            // Tuples should all be rewritten away.
            ast::Expression::Tuple(_) => Err(Error::CannotBeAlgebrized("tuples")),
            ast::Expression::Subquery(s) => self.algebrize_subquery(*s),
            ast::Expression::SubqueryComparison(s) => self.algebrize_subquery_comparison(s),
            ast::Expression::Exists(e) => self.algebrize_exists(*e),
        }
    }

    pub fn algebrize_literal(&self, ast_node: ast::Literal) -> ir::LiteralValue {
        match ast_node {
            ast::Literal::Null => ir::LiteralValue::Null,
            ast::Literal::Boolean(b) => ir::LiteralValue::Boolean(b),
            ast::Literal::String(s) => ir::LiteralValue::String(s),
            ast::Literal::Integer(i) => ir::LiteralValue::Integer(i),
            ast::Literal::Long(l) => ir::LiteralValue::Long(l),
            ast::Literal::Double(d) => ir::LiteralValue::Double(d),
        }
    }

    pub fn algebrize_limit_clause(
        &self,
        ast_node: Option<u32>,
        source: ir::Stage,
    ) -> Result<ir::Stage> {
        match ast_node {
            None => Ok(source),
            Some(x) => {
                let stage = ir::Stage::Limit(ir::Limit {
                    source: Box::new(source),
                    limit: u64::from(x),
                    cache: SchemaCache::new(),
                });
                stage.schema(&self.schema_inference_state())?;
                Ok(stage)
            }
        }
    }

    pub fn algebrize_offset_clause(
        &self,
        ast_node: Option<u32>,
        source: ir::Stage,
    ) -> Result<ir::Stage> {
        match ast_node {
            None => Ok(source),
            Some(x) => {
                let stage = ir::Stage::Offset(ir::Offset {
                    source: Box::new(source),
                    offset: u64::from(x),
                    cache: SchemaCache::new(),
                });
                stage.schema(&self.schema_inference_state())?;
                Ok(stage)
            }
        }
    }

    fn algebrize_function(&self, f: ast::FunctionExpr) -> Result<ir::Expression> {
        if f.set_quantifier == Some(ast::SetQuantifier::Distinct) {
            return Err(Error::DistinctScalarFunction);
        }

        // get the arguments as a vec of ast::Expressions. If the arguments are
        // Star this must be a COUNT function, otherwise it is an error.
        let args = match f.args {
            ast::FunctionArguments::Star => return Err(Error::StarInNonCount),
            ast::FunctionArguments::Args(ve) => ve,
        };

        // if the function is CURRENT_TIMESTAMP with exactly one arg,
        // throw away the argument. we break the spec intentionally
        // here by ignoring the date-precision argument. this is
        // implemented during algebrization instead of rewriting
        // because all other rewrites are compliant with the spec, and
        // this would be the only non-spec-compliant rewrite.
        let args = match (f.function, args.len()) {
            (ast::FunctionName::CurrentTimestamp, 1) => Vec::new(),
            _ => args,
        };

        schema_check_return!(
            self,
            ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::try_from(f.function)?,
                args: args
                    .into_iter()
                    .map(|e| self.algebrize_expression(e))
                    .collect::<Result<_>>()?,
                cache: SchemaCache::new(),
            })
        )
    }

    fn algebrize_unary_expr(&self, u: ast::UnaryExpr) -> Result<ir::Expression> {
        schema_check_return!(
            self,
            ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::from(u.op),
                args: vec![self.algebrize_expression(*u.expr)?],
                cache: SchemaCache::new(),
            }),
        );
    }

    fn algebrize_binary_expr(&self, b: ast::BinaryExpr) -> Result<ir::Expression> {
        schema_check_return!(
            self,
            ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::try_from(b.op)?,
                args: vec![
                    self.algebrize_expression(*b.left)?,
                    self.algebrize_expression(*b.right)?,
                ],
                cache: SchemaCache::new(),
            })
        );
    }

    fn algebrize_is(&self, ast_node: ast::IsExpr) -> Result<ir::Expression> {
        schema_check_return!(
            self,
            ir::Expression::Is(ir::IsExpr {
                expr: Box::new(self.algebrize_expression(*ast_node.expr)?),
                target_type: ir::TypeOrMissing::from(ast_node.target_type),
                cache: SchemaCache::new(),
            }),
        )
    }

    fn algebrize_like(&self, ast_node: ast::LikeExpr) -> Result<ir::Expression> {
        schema_check_return!(
            self,
            ir::Expression::Like(ir::LikeExpr {
                expr: Box::new(self.algebrize_expression(*ast_node.expr)?),
                pattern: Box::new(self.algebrize_expression(*ast_node.pattern)?),
                escape: ast_node.escape,
                cache: SchemaCache::new(),
            }),
        )
    }

    fn algebrize_between(&self, b: ast::BetweenExpr) -> Result<ir::Expression> {
        let (arg, min, max) = (
            self.algebrize_expression(*b.expr)?,
            self.algebrize_expression(*b.min)?,
            self.algebrize_expression(*b.max)?,
        );
        schema_check_return!(
            self,
            ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                function: ir::ScalarFunction::Between,
                args: vec![arg, min, max,],
                cache: SchemaCache::new(),
            })
        );
    }

    fn algebrize_trim(&self, t: ast::TrimExpr) -> Result<ir::Expression> {
        let function = match t.trim_spec {
            ast::TrimSpec::Leading => ir::ScalarFunction::LTrim,
            ast::TrimSpec::Trailing => ir::ScalarFunction::RTrim,
            ast::TrimSpec::Both => ir::ScalarFunction::BTrim,
        };
        schema_check_return!(
            self,
            ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                function,
                args: vec![
                    self.algebrize_expression(*t.trim_chars)?,
                    self.algebrize_expression(*t.arg)?,
                ],
                cache: SchemaCache::new(),
            }),
        );
    }

    fn algebrize_extract(&self, e: ast::ExtractExpr) -> Result<ir::Expression> {
        use crate::ast::ExtractSpec::*;
        let function = match e.extract_spec {
            Year => ir::ScalarFunction::Year,
            Month => ir::ScalarFunction::Month,
            Day => ir::ScalarFunction::Day,
            Hour => ir::ScalarFunction::Hour,
            Minute => ir::ScalarFunction::Minute,
            Second => ir::ScalarFunction::Second,
        };
        schema_check_return!(
            self,
            ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                function,
                args: vec![self.algebrize_expression(*e.arg)?],
                cache: SchemaCache::new(),
            }),
        )
    }

    fn algebrize_access(&self, a: ast::AccessExpr) -> Result<ir::Expression> {
        let expr = self.algebrize_expression(*a.expr)?;
        schema_check_return!(
            self,
            match *a.subfield {
                ast::Expression::Literal(ast::Literal::String(s)) =>
                    ir::Expression::FieldAccess(ir::FieldAccess {
                        expr: Box::new(expr),
                        field: s,
                        cache: SchemaCache::new(),
                    }),
                sf => ir::Expression::ScalarFunction(ir::ScalarFunctionApplication {
                    function: ir::ScalarFunction::ComputedFieldAccess,
                    args: vec![expr, self.algebrize_expression(sf)?],
                    cache: SchemaCache::new(),
                }),
            }
        );
    }

    fn algebrize_type_assertion(&self, t: ast::TypeAssertionExpr) -> Result<ir::Expression> {
        schema_check_return!(
            self,
            ir::Expression::TypeAssertion(ir::TypeAssertionExpr {
                expr: Box::new(self.algebrize_expression(*t.expr)?),
                target_type: ir::Type::from(t.target_type),
                cache: SchemaCache::new(),
            }),
        );
    }

    fn algebrize_case(&self, c: ast::CaseExpr) -> Result<ir::Expression> {
        let else_branch = c
            .else_branch
            .map(|e| self.algebrize_expression(*e))
            .transpose()?
            .map(Box::new)
            .unwrap_or_else(|| Box::new(ir::Expression::Literal(ir::LiteralValue::Null.into())));
        let expr = c.expr.map(|e| self.algebrize_expression(*e)).transpose()?;
        let when_branch = c
            .when_branch
            .into_iter()
            .map(|wb| {
                Ok(ir::WhenBranch {
                    when: Box::new(self.algebrize_expression(*wb.when)?),
                    then: Box::new(self.algebrize_expression(*wb.then)?),
                })
            })
            .collect::<Result<_>>()?;
        match expr {
            Some(expr) => {
                let expr = Box::new(expr);
                schema_check_return!(
                    self,
                    ir::Expression::SimpleCase(ir::SimpleCaseExpr {
                        expr,
                        when_branch,
                        else_branch,
                        cache: SchemaCache::new(),
                    }),
                )
            }
            None => {
                schema_check_return!(
                    self,
                    ir::Expression::SearchedCase(ir::SearchedCaseExpr {
                        when_branch,
                        else_branch,
                        cache: SchemaCache::new(),
                    }),
                )
            }
        }
    }

    fn algebrize_cast(&self, c: ast::CastExpr) -> Result<ir::Expression> {
        macro_rules! null_expr {
            () => {{
                Box::new(ast::Expression::Literal(ast::Literal::Null))
            }};
        }
        schema_check_return!(
            self,
            ir::Expression::Cast(ir::CastExpr {
                expr: Box::new(self.algebrize_expression(*c.expr)?),
                to: ir::Type::from(c.to),
                on_null: Box::new(
                    self.algebrize_expression(*(c.on_null.unwrap_or_else(|| null_expr!())))?
                ),
                on_error: Box::new(
                    self.algebrize_expression(*(c.on_error.unwrap_or_else(|| null_expr!())))?
                ),
                cache: SchemaCache::new(),
            }),
        );
    }

    pub fn algebrize_subquery_expr(&self, ast_node: ast::Query) -> Result<ir::SubqueryExpr> {
        let subquery_algebrizer = self.subquery_algebrizer();
        let subquery = Box::new(subquery_algebrizer.algebrize_query(ast_node)?);
        let result_set = subquery.schema(&subquery_algebrizer.schema_inference_state())?;

        match result_set.schema_env.len() {
            1 => {
                let (key, schema) = result_set.schema_env.into_iter().next().unwrap();
                let output_expr = match &schema.get_single_field_name() {
                    Some(field) => Ok(Box::new(ir::Expression::FieldAccess(ir::FieldAccess {
                        expr: Box::new(ir::Expression::Reference(key.into())),
                        field: field.to_string(),
                        cache: SchemaCache::new(),
                    }))),
                    None => Err(Error::InvalidSubqueryDegree),
                }?;
                Ok(ir::SubqueryExpr {
                    output_expr,
                    subquery,
                    cache: SchemaCache::new(),
                })
            }
            _ => Err(Error::InvalidSubqueryDegree),
        }
    }

    pub fn algebrize_subquery(&self, ast_node: ast::Query) -> Result<ir::Expression> {
        schema_check_return!(
            self,
            ir::Expression::Subquery(self.algebrize_subquery_expr(ast_node)?)
        )
    }

    pub fn algebrize_subquery_comparison(
        &self,
        s: ast::SubqueryComparisonExpr,
    ) -> Result<ir::Expression> {
        let modifier = match s.quantifier {
            ast::SubqueryQuantifier::All => ir::SubqueryModifier::All,
            ast::SubqueryQuantifier::Any => ir::SubqueryModifier::Any,
        };
        schema_check_return!(
            self,
            ir::Expression::SubqueryComparison(ir::SubqueryComparison {
                operator: ir::SubqueryComparisonOp::from(s.op),
                modifier,
                argument: Box::new(self.algebrize_expression(*s.expr)?),
                subquery_expr: self.algebrize_subquery_expr(*s.subquery)?,
                cache: SchemaCache::new(),
            })
        )
    }

    pub fn algebrize_exists(&self, ast_node: ast::Query) -> Result<ir::Expression> {
        let exists = self.subquery_algebrizer().algebrize_query(ast_node)?;
        schema_check_return!(self, ir::Expression::Exists(Box::new(exists).into()));
    }

    fn algebrize_subpath(&self, p: ast::SubpathExpr) -> Result<ir::Expression> {
        if let ast::Expression::Identifier(s) = *p.expr {
            schema_check_return!(
                self,
                self.algebrize_possibly_qualified_field_access(s, p.subpath)?,
            );
        }
        schema_check_return!(
            self,
            ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(self.algebrize_expression(*p.expr)?),
                field: p.subpath,
                cache: SchemaCache::new(),
            }),
        );
    }

    fn algebrize_possibly_qualified_field_access(
        &self,
        q: String,
        field: String,
    ) -> Result<ir::Expression> {
        // clone the field here so that we only have to clone once.
        // The borrow checker still isn't perfect.
        let cloned_field = field.clone();
        // First we check if q is a qualifier
        let possible_datasource = DatasourceName::from(q.clone());
        // If there is a nearest_scope for `q`, then it must be a datasource, meaning this is a
        // qualified field access
        self.schema_env
            .nearest_scope_for_datasource(&possible_datasource, self.scope_level)
            .map_or_else(
                move || {
                    Ok(ir::Expression::FieldAccess(ir::FieldAccess {
                        expr: Box::new(self.algebrize_unqualified_identifier(q)?),
                        // combinators make this clone necessary, unfortunately
                        field: cloned_field,
                        cache: SchemaCache::new(),
                    }))
                },
                move |scope|
                // Since this is qualified, we return `q.field`
                Ok(ir::Expression::FieldAccess(ir::FieldAccess {
                    expr: Box::new(ir::Expression::Reference(Key {
                    datasource: possible_datasource,
                    scope,
                }.into())),
                field,
                    cache: SchemaCache::new(),
            })),
            )
    }

    fn algebrize_unqualified_identifier(&self, i: String) -> Result<ir::Expression> {
        // Attempt to find a datasource for this unqualified reference
        // at _any_ scope level.
        // If we find exactly one datasource that May or Must contain
        // the field `i`, we return `datasource.i`. If there is more
        // than one, it is an ambiguous error.
        let mut i_containing_datasources = self
            .schema_env
            .iter()
            .filter(|(_, schema)| {
                let sat = schema.contains_field(i.as_ref());
                sat == Satisfaction::May || sat == Satisfaction::Must
            })
            .collect::<Vec<_>>();
        // If there is no datasource containing the field, the field is not found.
        if i_containing_datasources.is_empty() {
            return Err(Error::FieldNotFound(i));
        }
        // If there is exactly one possible datasource that May or Must
        // contain our reference, we use it.
        if i_containing_datasources.len() == 1 {
            return Ok(ir::Expression::FieldAccess(ir::FieldAccess {
                expr: Box::new(ir::Expression::Reference(
                    i_containing_datasources.remove(0).0.clone().into(),
                )),
                field: i,
                cache: SchemaCache::new(),
            }));
        }

        // Otherwise, we check datasources per scope, starting at the current scope,
        // to find the best datasource from multiple possible datasources.
        self.algebrize_unqualified_identifier_by_scope(i, self.scope_level)
    }

    fn algebrize_unqualified_identifier_by_scope(
        &self,
        i: String,
        scope_level: u16,
    ) -> Result<ir::Expression> {
        // When checking variables by scope, if a variable may exist, we treat that as ambiguous,
        // and only accept a single Must exist reference.
        let mut current_scope = scope_level;
        loop {
            let current_bot = Key::bot(current_scope);
            // Attempt to find a datasource for this reference in the current_scope.
            // If we find exactly one datasource Must contain the field `i`, we return
            // `datasource.i`. If there is more than one, it is an ambiguous error. As mentioned,
            // if there is a May exists, it is also an ambiguous variable error.
            let (datasource, mays, musts) = self
                .schema_env
                .iter()
                .filter(
                    |(
                        &Key {
                            datasource: _,
                            scope: n,
                        },
                        _,
                    )| n == current_scope,
                )
                .fold(
                    (&current_bot, 0, 0),
                    |(found_datasource, mays, musts), (curr_datasource, schema)| {
                        let sat = schema.contains_field(i.as_ref());
                        match sat {
                            Satisfaction::Must => (curr_datasource, mays, musts + 1),
                            Satisfaction::May => (found_datasource, mays + 1, musts),
                            Satisfaction::Not => (found_datasource, mays, musts),
                        }
                    },
                );
            if musts > 1 || mays > 0 {
                return Err(Error::AmbiguousField(i));
            }
            if musts == 1 {
                return Ok(ir::Expression::FieldAccess(ir::FieldAccess {
                    expr: Box::new(ir::Expression::Reference(datasource.clone().into())),
                    field: i,
                    cache: SchemaCache::new(),
                }));
            }

            // Otherwise, the field does not exist in datasource of the current_scope.
            //
            // If the current_scope is 0, it must be that this field does not exist in the
            // SchemaEnv at all, which means the field cannot be found. This should not
            // be possible at this point, because this error is handled in `algebrize_qualified_identifier`.
            if current_scope == 0 {
                unreachable!();
            }
            // Otherwise, check the next highest scope.
            current_scope -= 1;
        }
    }
}
