use crate::{
    algebrizer::errors::Error,
    ast::{self, pretty_print::PrettyPrint},
    catalog::Catalog,
    map,
    mir::{
        self,
        binding_tuple::{BindingTuple, DatasourceName, Key},
        schema::{CachedSchema, SchemaCache, SchemaInferenceState},
        AliasedExpr, Expression, FieldAccess, OptionallyAliasedExpr, ReferenceExpr,
    },
    schema::{
        self, Satisfaction, SchemaEnvironment, ANY_DOCUMENT, BOOLEAN_OR_NULLISH,
        INTEGER_LONG_OR_NULLISH, INTEGER_OR_NULLISH, NULLISH, STRING_OR_NULLISH,
    },
    util::unique_linked_hash_map::UniqueLinkedHashMap,
    SchemaCheckingMode,
};
use std::{cell::RefCell, collections::BTreeSet};

type Result<T> = std::result::Result<T, Error>;

macro_rules! schema_check_return {
    ($self:ident, $e:expr $(,)?) => {{
        let ret = $e;
        ret.schema(&$self.schema_inference_state())?;
        return Ok(ret);
    }};
}

impl TryFrom<ast::BinaryOp> for mir::ScalarFunction {
    type Error = Error;

    fn try_from(op: crate::ast::BinaryOp) -> Result<Self> {
        Ok(match op {
            ast::BinaryOp::Add => mir::ScalarFunction::Add,
            ast::BinaryOp::And => mir::ScalarFunction::And,
            ast::BinaryOp::Concat => mir::ScalarFunction::Concat,
            ast::BinaryOp::Div => mir::ScalarFunction::Div,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Eq) => mir::ScalarFunction::Eq,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Gt) => mir::ScalarFunction::Gt,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Gte) => mir::ScalarFunction::Gte,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Lt) => mir::ScalarFunction::Lt,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Lte) => mir::ScalarFunction::Lte,
            ast::BinaryOp::Comparison(ast::ComparisonOp::Neq) => mir::ScalarFunction::Neq,
            ast::BinaryOp::Mul => mir::ScalarFunction::Mul,
            ast::BinaryOp::Or => mir::ScalarFunction::Or,
            ast::BinaryOp::Sub => mir::ScalarFunction::Sub,
            ast::BinaryOp::In | ast::BinaryOp::NotIn => {
                panic!("{0} cannot be algebrized", op.as_str())
            }
        })
    }
}

impl TryFrom<ast::FunctionName> for mir::ScalarFunction {
    type Error = Error;

    fn try_from(f: crate::ast::FunctionName) -> Result<Self> {
        Ok(match f {
            ast::FunctionName::Abs => mir::ScalarFunction::Abs,
            ast::FunctionName::BitLength => mir::ScalarFunction::BitLength,
            ast::FunctionName::Ceil => mir::ScalarFunction::Ceil,
            ast::FunctionName::CharLength => mir::ScalarFunction::CharLength,
            ast::FunctionName::Coalesce => mir::ScalarFunction::Coalesce,
            ast::FunctionName::Cos => mir::ScalarFunction::Cos,
            ast::FunctionName::CurrentTimestamp => mir::ScalarFunction::CurrentTimestamp,
            ast::FunctionName::Degrees => mir::ScalarFunction::Degrees,
            ast::FunctionName::Floor => mir::ScalarFunction::Floor,
            ast::FunctionName::Log => mir::ScalarFunction::Log,
            ast::FunctionName::Lower => mir::ScalarFunction::Lower,
            ast::FunctionName::Mod => mir::ScalarFunction::Mod,
            ast::FunctionName::NullIf => mir::ScalarFunction::NullIf,
            ast::FunctionName::OctetLength => mir::ScalarFunction::OctetLength,
            ast::FunctionName::Position => mir::ScalarFunction::Position,
            ast::FunctionName::Pow => mir::ScalarFunction::Pow,
            ast::FunctionName::Radians => mir::ScalarFunction::Radians,
            ast::FunctionName::Replace => mir::ScalarFunction::Replace,
            ast::FunctionName::Sin => mir::ScalarFunction::Sin,
            ast::FunctionName::Size => mir::ScalarFunction::Size,
            ast::FunctionName::Slice => mir::ScalarFunction::Slice,
            ast::FunctionName::Split => mir::ScalarFunction::Split,
            ast::FunctionName::Sqrt => mir::ScalarFunction::Sqrt,
            ast::FunctionName::Substring => mir::ScalarFunction::Substring,
            ast::FunctionName::Tan => mir::ScalarFunction::Tan,
            ast::FunctionName::Upper => mir::ScalarFunction::Upper,
            ast::FunctionName::Round => mir::ScalarFunction::Round,
            ast::FunctionName::DayOfWeek => mir::ScalarFunction::DayOfWeek,
            ast::FunctionName::LTrim
            | ast::FunctionName::RTrim
            | ast::FunctionName::Log10
            | ast::FunctionName::DateAdd
            | ast::FunctionName::DateDiff
            | ast::FunctionName::DateTrunc
            | ast::FunctionName::Year
            | ast::FunctionName::Month
            | ast::FunctionName::Week
            | ast::FunctionName::DayOfMonth
            | ast::FunctionName::DayOfYear
            | ast::FunctionName::Hour
            | ast::FunctionName::Minute
            | ast::FunctionName::Second
            | ast::FunctionName::Millisecond => unreachable! {},
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

impl TryFrom<ast::FunctionName> for mir::AggregationFunction {
    type Error = Error;

    fn try_from(f: crate::ast::FunctionName) -> Result<Self> {
        Ok(match f {
            ast::FunctionName::AddToArray => mir::AggregationFunction::AddToArray,
            ast::FunctionName::AddToSet => panic!("ADD_TO_SET should be removed before try_from"),
            ast::FunctionName::Avg => mir::AggregationFunction::Avg,
            ast::FunctionName::Count => mir::AggregationFunction::Count,
            ast::FunctionName::First => mir::AggregationFunction::First,
            ast::FunctionName::Last => mir::AggregationFunction::Last,
            ast::FunctionName::Max => mir::AggregationFunction::Max,
            ast::FunctionName::MergeDocuments => mir::AggregationFunction::MergeDocuments,
            ast::FunctionName::Min => mir::AggregationFunction::Min,
            ast::FunctionName::StddevPop => mir::AggregationFunction::StddevPop,
            ast::FunctionName::StddevSamp => mir::AggregationFunction::StddevSamp,
            ast::FunctionName::Sum => mir::AggregationFunction::Sum,

            ast::FunctionName::Abs
            | ast::FunctionName::BitLength
            | ast::FunctionName::Ceil
            | ast::FunctionName::CharLength
            | ast::FunctionName::Coalesce
            | ast::FunctionName::Cos
            | ast::FunctionName::CurrentTimestamp
            | ast::FunctionName::Degrees
            | ast::FunctionName::Floor
            | ast::FunctionName::Log
            | ast::FunctionName::Log10
            | ast::FunctionName::Lower
            | ast::FunctionName::LTrim
            | ast::FunctionName::Mod
            | ast::FunctionName::NullIf
            | ast::FunctionName::OctetLength
            | ast::FunctionName::Position
            | ast::FunctionName::Pow
            | ast::FunctionName::Round
            | ast::FunctionName::Radians
            | ast::FunctionName::Replace
            | ast::FunctionName::RTrim
            | ast::FunctionName::Sin
            | ast::FunctionName::Size
            | ast::FunctionName::Slice
            | ast::FunctionName::Split
            | ast::FunctionName::Sqrt
            | ast::FunctionName::Substring
            | ast::FunctionName::Tan
            | ast::FunctionName::Upper
            | ast::FunctionName::DateAdd
            | ast::FunctionName::DateDiff
            | ast::FunctionName::DateTrunc
            | ast::FunctionName::Year
            | ast::FunctionName::Month
            | ast::FunctionName::Week
            | ast::FunctionName::DayOfWeek
            | ast::FunctionName::DayOfMonth
            | ast::FunctionName::DayOfYear
            | ast::FunctionName::Hour
            | ast::FunctionName::Minute
            | ast::FunctionName::Second
            | ast::FunctionName::Millisecond => {
                return Err(Error::ScalarInPlaceOfAggregation(f.pretty_print().unwrap()))
            }
        })
    }
}

impl From<crate::ast::ComparisonOp> for mir::SubqueryComparisonOp {
    fn from(op: crate::ast::ComparisonOp) -> Self {
        match op {
            ast::ComparisonOp::Eq => mir::SubqueryComparisonOp::Eq,
            ast::ComparisonOp::Gt => mir::SubqueryComparisonOp::Gt,
            ast::ComparisonOp::Gte => mir::SubqueryComparisonOp::Gte,
            ast::ComparisonOp::Lt => mir::SubqueryComparisonOp::Lt,
            ast::ComparisonOp::Lte => mir::SubqueryComparisonOp::Lte,
            ast::ComparisonOp::Neq => mir::SubqueryComparisonOp::Neq,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub enum ClauseType {
    From,
    GroupBy,
    Having,
    Limit,
    Offset,
    OrderBy,
    Select,
    Unintialized,
    Where,
}

impl std::fmt::Display for ClauseType {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        match self {
            ClauseType::From => write!(f, "FROM"),
            ClauseType::GroupBy => write!(f, "GROUP BY"),
            ClauseType::Having => write!(f, "HAVING"),
            ClauseType::Limit => write!(f, "LIMIT"),
            ClauseType::Offset => write!(f, "OFFSET"),
            ClauseType::OrderBy => write!(f, "ORDER BY"),
            ClauseType::Select => write!(f, "SELECT"),
            ClauseType::Unintialized => write!(f, "UNINITIALIZED"),
            ClauseType::Where => write!(f, "WHERE"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct Algebrizer<'a> {
    current_db: &'a str,
    pub schema_env: SchemaEnvironment,
    catalog: &'a Catalog,
    scope_level: u16,
    schema_checking_mode: SchemaCheckingMode,
    allow_order_by_missing_columns: bool,
    clause_type: RefCell<ClauseType>,
}

impl<'a> Algebrizer<'a> {
    pub fn new(
        current_db: &'a str,
        catalog: &'a Catalog,
        scope_level: u16,
        schema_checking_mode: SchemaCheckingMode,
        allow_order_by_missing_columns: bool,
        clause_type: ClauseType,
    ) -> Self {
        Self::with_schema_env(
            current_db,
            SchemaEnvironment::default(),
            catalog,
            scope_level,
            schema_checking_mode,
            allow_order_by_missing_columns,
            clause_type,
        )
    }

    pub fn with_schema_env(
        current_db: &'a str,
        schema_env: SchemaEnvironment,
        catalog: &'a Catalog,
        scope_level: u16,
        schema_checking_mode: SchemaCheckingMode,
        allow_order_by_missing_columns: bool,
        clause_type: ClauseType,
    ) -> Self {
        Self {
            current_db,
            schema_env,
            catalog,
            scope_level,
            schema_checking_mode,
            allow_order_by_missing_columns,
            clause_type: RefCell::new(clause_type),
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
            // subquery should use a copy of the current ClauseType, not share a RefCell with the
            // parent. It probably does not matter, but we do not want subqueries to modify parent
            // state.
            allow_order_by_missing_columns: self.allow_order_by_missing_columns,
            clause_type: RefCell::new(*self.clause_type.borrow()),
        }
    }

    fn args_are_nullable(args: &[mir::Expression]) -> bool {
        args.iter().any(|e| e.is_nullable())
    }

    pub fn determine_scalar_function_nullability(
        func: mir::ScalarFunction,
        args: &[mir::Expression],
    ) -> bool {
        // some functions can always be nullable regardless of argument nullablity,
        // we check those first. If this function is not one of those, we set nullablity
        // based off the arguments.
        func.is_always_nullable() || Self::args_are_nullable(args)
    }

    pub fn algebrize_query(&self, ast_node: ast::Query) -> Result<mir::Stage> {
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

    pub fn construct_field_access_expr(
        &self,
        expr: mir::Expression,
        field: String,
    ) -> Result<mir::Expression> {
        let mut fa = FieldAccess::new(Box::new(expr), field);

        fa.schema(&self.schema_inference_state()).map(|schema| {
            fa.is_nullable = NULLISH.satisfies(&schema) != Satisfaction::Not;
            Ok(mir::Expression::FieldAccess(fa))
        })?
    }

    pub fn algebrize_select_query(&self, ast_node: ast::SelectQuery) -> Result<mir::Stage> {
        let plan = self.algebrize_from_clause(ast_node.from_clause)?;
        let plan = self.algebrize_where_clause(ast_node.where_clause, plan)?;
        let plan = self.algebrize_group_by_clause(ast_node.group_by_clause, plan)?;
        let plan = self.algebrize_having_clause(ast_node.having_clause, plan)?;
        let plan = if self.allow_order_by_missing_columns {
            self.algebrize_select_and_order_by_clause(
                ast_node.select_clause,
                ast_node.order_by_clause,
                plan,
            )?
        } else {
            let plan = self.algebrize_select_clause(ast_node.select_clause, plan, false)?;
            self.algebrize_order_by_clause(ast_node.order_by_clause, plan)?
        };
        let plan = self.algebrize_offset_clause(ast_node.offset, plan)?;
        let plan = self.algebrize_limit_clause(ast_node.limit, plan)?;
        Ok(plan)
    }

    pub fn algebrize_set_query(&self, ast_node: ast::SetQuery) -> Result<mir::Stage> {
        match ast_node.op {
            ast::SetOperator::Union => {
                let union_all_stage = mir::Stage::Set(mir::Set {
                    operation: mir::SetOperation::UnionAll,
                    left: Box::new(self.algebrize_query(*ast_node.left)?),
                    right: Box::new(self.algebrize_query(*ast_node.right)?),
                    cache: SchemaCache::new(),
                });

                let union_result_set = union_all_stage.schema(&self.schema_inference_state())?;
                let datasources: BTreeSet<Key> = union_result_set
                    .schema_env
                    .keys()
                    .filter(|key| key.scope == self.scope_level)
                    .cloned()
                    .collect();

                // Create group keys for each datasource
                let group_keys: Vec<OptionallyAliasedExpr> = datasources
                    .iter()
                    .enumerate()
                    .map(|(i, key)| {
                        OptionallyAliasedExpr::Aliased(AliasedExpr {
                            alias: format!("__groupKey{}", i),
                            expr: Expression::Reference(ReferenceExpr { key: key.clone() }),
                        })
                    })
                    .collect();

                // Adding group stage to union all to remove duplicates
                let group_stage = mir::Stage::Group(mir::Group {
                    source: Box::new(union_all_stage),
                    keys: group_keys,
                    aggregations: vec![],
                    cache: SchemaCache::new(),
                    scope: self.scope_level,
                });

                let mut project_expression = BindingTuple::new();
                for (i, key) in datasources.iter().enumerate() {
                    project_expression.insert(
                        key.clone(),
                        Expression::FieldAccess(FieldAccess {
                            expr: Box::new(Expression::Reference(ReferenceExpr {
                                key: Key::bot(self.scope_level),
                            })),
                            field: format!("__groupKey{}", i),
                            // Setting is_nullable to true because the result set coming into the
                            // UNION clause could be empty.
                            is_nullable: true,
                        }),
                    );
                }

                let project_stage = mir::Stage::Project(mir::Project {
                    source: Box::new(group_stage),
                    expression: project_expression,
                    is_add_fields: false,
                    cache: SchemaCache::new(),
                });

                schema_check_return!(self, project_stage)
            }
            ast::SetOperator::UnionAll => schema_check_return!(
                self,
                mir::Stage::Set(mir::Set {
                    operation: mir::SetOperation::UnionAll,
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
        source: mir::Stage,
        is_add_fields: bool,
        is_distinct: bool,
    ) -> Result<mir::Stage> {
        let expression_algebrizer = self.clone();
        // Algebrization for every node that has a source should get the schema for the source.
        // The SchemaEnvironment from the source is merged into the SchemaEnvironment from the
        // current Algebrizer, correctly giving us the correlated bindings with the bindings
        // available from the current query level.
        #[allow(unused_variables)]
        let expression_algebrizer = expression_algebrizer
            .with_merged_mappings(source.schema(&self.schema_inference_state())?.schema_env)?;

        // We must check for duplicate Datasource Keys, which is an error. The datasources
        // Set keeps track of which Keys have been seen.
        let mut datasources = BTreeSet::new();
        let mut bottom: Option<Vec<ast::DocumentPair>> = None;

        // Build the Project expression from the SelectBody::Values(exprs)
        let mut expression = BindingTuple::new();
        for expr in exprs.into_iter() {
            match expr {
                // An Expression is mapped to DatasourceName::Bottom. Bottom should be a document,
                // but can be in multiple parts based on the query. Unify them using a single vec
                // and algebrize after all expressions have been evaluated.
                ast::SelectValuesExpression::Expression(e) => match e {
                    ast::Expression::Document(d) => {
                        bottom = if let Some(mut bottom) = bottom {
                            bottom.extend(d);
                            Some(bottom)
                        } else {
                            Some(d)
                        }
                    }
                    // If select values are not a document, an error will ultimately be thrown. Algebrize
                    // for now, and depending on the rest of the select query, a DuplicateKey or SchemaChecking
                    // error will occur downstream.
                    _ => {
                        let e = expression_algebrizer.algebrize_expression(e, false)?;
                        let bot = Key::bot(expression_algebrizer.scope_level);
                        datasources
                            .insert(bot.clone())
                            .then_some(())
                            .ok_or_else(|| Error::DuplicateKey(bot.clone()))?;
                        expression.insert(bot, e);
                    }
                },
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
                        .then_some(())
                        .ok_or_else(|| Error::DuplicateKey(key.clone()))?;
                    let scope = expression_algebrizer
                        .schema_env
                        .nearest_scope_for_datasource(
                            &datasource,
                            expression_algebrizer.scope_level,
                        )
                        .ok_or_else(|| Error::NoSuchDatasource(datasource.clone()))?;
                    expression.insert(
                        key,
                        mir::Expression::Reference(
                            Key {
                                datasource: DatasourceName::Named(s.datasource),
                                scope,
                            }
                            .into(),
                        ),
                    );
                }
            }
        }

        // if we found Expressions's, algebrize them as a single document, and add it to the expression
        // under the Bottom namespace.
        if bottom.is_some() {
            let e = expression_algebrizer
                .algebrize_expression(ast::Expression::Document(bottom.unwrap()), false)?;
            let bot = Key::bot(expression_algebrizer.scope_level);
            datasources
                .insert(bot.clone())
                .then_some(())
                .ok_or_else(|| Error::DuplicateKey(bot.clone()))?;
            expression.insert(bot, e);
        }

        // Build the Project Stage using the source and built expression.
        let project_stage = mir::Stage::Project(mir::Project {
            source: Box::new(source),
            expression: expression.clone(),
            is_add_fields,
            cache: SchemaCache::new(),
        });
        project_stage.schema(&self.schema_inference_state())?;
        Ok(if is_distinct {
            let distinct_stages = Self::build_distinct_values_group_and_project_stage(
                project_stage,
                self.scope_level,
                datasources,
            );
            distinct_stages.schema(&self.schema_inference_state())?;
            distinct_stages
        } else {
            project_stage
        })
    }

    pub fn build_distinct_values_group_and_project_stage(
        source_stage: mir::Stage,
        scope_level: u16,
        datasources: BTreeSet<Key>,
    ) -> mir::Stage {
        let mut group_keys = Vec::new();
        let mut project_expressions = BindingTuple::new();

        for (counter, key) in datasources.iter().enumerate() {
            let group_alias = format!("__groupKey{}", counter);

            group_keys.push(OptionallyAliasedExpr::Aliased(AliasedExpr {
                alias: group_alias.clone(),
                expr: Expression::Reference(ReferenceExpr { key: key.clone() }),
            }));
            project_expressions.insert(
                key.clone(),
                Expression::FieldAccess(FieldAccess {
                    expr: Box::new(Expression::Reference(ReferenceExpr {
                        key: Key::bot(scope_level),
                    })),
                    field: group_alias,
                    // Setting is_nullable to true because the result set coming into the
                    // SELECT DISTINCT clause could be empty.
                    is_nullable: true,
                }),
            );
        }

        let group_stage = mir::Stage::Group(mir::Group {
            source: Box::new(source_stage),
            keys: group_keys,
            aggregations: vec![],
            cache: Default::default(),
            scope: scope_level,
        });

        mir::Stage::Project(mir::Project {
            source: Box::new(group_stage),
            expression: project_expressions,
            is_add_fields: false,
            cache: Default::default(),
        })
    }

    pub fn algebrize_from_clause(&self, ast_node: Option<ast::Datasource>) -> Result<mir::Stage> {
        *self.clause_type.borrow_mut() = ClauseType::From;
        let ast_node = ast_node.expect("all SELECT queries must have a FROM clause");
        // Datasource aliases are tracked in the algebrizer's SchemaEnvironment.
        // Duplicate aliases are invalid.  Under some circumstances, the
        // SchemaEnvironment with_merged_mappings method allows duplicates
        // to be inserted, therefore we must catch them explicitly here.
        ast_node
            .check_duplicate_aliases()
            .map_err(|e| Error::DuplicateKey((e, self.scope_level).into()))?;
        self.algebrize_datasource(ast_node)
    }

    pub fn algebrize_datasource(&self, ast_node: ast::Datasource) -> Result<mir::Stage> {
        match ast_node {
            ast::Datasource::Array(a) => self.algebrize_array_datasource(a),
            ast::Datasource::Collection(c) => self.algebrize_collection_datasource(c),
            ast::Datasource::Join(j) => self.algebrize_join_datasource(j),
            ast::Datasource::Derived(d) => self.algebrize_derived_datasource(d),
            ast::Datasource::Flatten(f) => self.algebrize_flatten_datasource(f),
            ast::Datasource::Unwind(u) => self.algebrize_unwind_datasource(u),
        }
    }

    fn algebrize_array_datasource(&self, a: ast::ArraySource) -> Result<mir::Stage> {
        let (ve, alias) = (a.array, a.alias.clone());
        let (ve, array_is_literal) = ast::visitors::are_literal(ve);
        if !array_is_literal {
            return Err(Error::ArrayDatasourceMustBeLiteral);
        }
        let src = mir::Stage::Array(mir::ArraySource {
            array: ve
                .into_iter()
                .map(|e| self.algebrize_expression(e, false))
                .collect::<Result<_>>()?,
            alias,
            cache: SchemaCache::new(),
        });
        let expr_map = map! {
            (a.alias.clone(), self.scope_level).into() =>
                mir::Expression::Reference((a.alias, self.scope_level).into())
        };
        let stage = mir::Stage::Project(mir::Project {
            source: Box::new(src),
            expression: expr_map,
            is_add_fields: false,
            cache: SchemaCache::new(),
        });
        stage.schema(&self.schema_inference_state())?;
        Ok(stage)
    }

    fn algebrize_collection_datasource(&self, c: ast::CollectionSource) -> Result<mir::Stage> {
        let src = mir::Stage::Collection(mir::Collection {
            db: c.database.unwrap_or_else(|| self.current_db.to_string()),
            collection: c.collection.clone(),
            cache: SchemaCache::new(),
        });
        let stage = match c.alias {
            Some(alias) => {
                let mut expr_map: BindingTuple<mir::Expression> = BindingTuple::new();
                expr_map.insert(
                    (alias, self.scope_level).into(),
                    mir::Expression::Reference((c.collection, self.scope_level).into()),
                );
                mir::Stage::Project(mir::Project {
                    source: Box::new(src),
                    expression: expr_map,
                    is_add_fields: false,
                    cache: SchemaCache::new(),
                })
            }
            None => panic!("collection datasources must have aliases"),
        };
        stage.schema(&self.schema_inference_state())?;
        Ok(stage)
    }

    fn algebrize_join_datasource(&self, j: ast::JoinSource) -> Result<mir::Stage> {
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
            .map(|e| join_algebrizer.algebrize_expression(e, false))
            .transpose()?
            .map(Self::convert_literal_to_bool);
        condition
            .clone()
            .map(|e| e.schema(&join_algebrizer.schema_inference_state()));
        let stage = match j.join_type {
            ast::JoinType::Left => {
                if condition.is_none() {
                    return Err(Error::NoOuterJoinCondition);
                }
                mir::Stage::Join(mir::Join {
                    join_type: mir::JoinType::Left,
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
                mir::Stage::Join(mir::Join {
                    join_type: mir::JoinType::Left,
                    left: Box::new(right_src),
                    right: Box::new(left_src),
                    condition,
                    cache: SchemaCache::new(),
                })
            }
            ast::JoinType::Cross | ast::JoinType::Inner => {
                let join = mir::Stage::Join(mir::Join {
                    join_type: mir::JoinType::Inner,
                    left: Box::new(left_src),
                    right: Box::new(right_src),
                    condition: None,
                    cache: SchemaCache::new(),
                });
                // The stage_movement optimization will place this condition in the Join if it makes sense. Otherwise,
                // it will move it as early in the pipeline as possible.
                if let Some(condition) = condition {
                    mir::Stage::Filter(mir::Filter {
                        source: Box::new(join),
                        condition,
                        cache: SchemaCache::new(),
                    })
                } else {
                    join
                }
            }
        };
        stage.schema(&self.schema_inference_state())?;
        Ok(stage)
    }

    fn algebrize_derived_datasource(&self, d: ast::DerivedSource) -> Result<mir::Stage> {
        let derived_algebrizer = Algebrizer::new(
            self.current_db,
            self.catalog,
            self.scope_level + 1,
            self.schema_checking_mode,
            self.allow_order_by_missing_columns,
            *self.clause_type.borrow(),
        );
        let src = derived_algebrizer.algebrize_query(*d.query)?;
        let src_resultset = src.schema(&derived_algebrizer.schema_inference_state())?;
        let mut datasource_refs = src_resultset
            .schema_env
            .into_iter()
            .map(|(k, _)| mir::Expression::Reference(k.into()))
            .collect::<Vec<mir::Expression>>();
        let expression = map! {
            (d.alias.clone(), self.scope_level).into() =>
            if datasource_refs.len() == 1 {
                datasource_refs.pop().unwrap()
            } else {
                mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                    function: mir::ScalarFunction::MergeObjects,
                    is_nullable: false,
                    args: datasource_refs
                })
            },
        };
        let stage = mir::Stage::Project(mir::Project {
            source: Box::new(src),
            expression,
            is_add_fields: false,
            cache: SchemaCache::new(),
        });
        stage
            .schema(&derived_algebrizer.schema_inference_state())
            .map_err(|e| match e {
                mir::schema::Error::CannotMergeObjects(s1, s2, sat) => {
                    Error::DerivedDatasourceOverlappingKeys(s1, s2, d.alias, sat)
                }
                _ => Error::SchemaChecking(e),
            })?;

        Ok(mir::Stage::Derived(mir::Derived {
            source: Box::new(stage),
            cache: SchemaCache::new(),
        }))
    }

    fn algebrize_flatten_datasource(&self, f: ast::FlattenSource) -> Result<mir::Stage> {
        let source = self.algebrize_datasource(*f.datasource.clone())?;
        let source_result_set = source.schema(&self.schema_inference_state())?;

        // Extract user-specified separator and depth. Separator defaults to "_".
        #[allow(clippy::manual_try_fold)]
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
                let (field_paths, has_only_nullable_polymorphism) = schema
                    .enumerate_field_paths(depth.map(|d| d + 1))
                    .map_err(|e| match e {
                        schema::Error::CannotEnumerateAllFieldPaths(s) => {
                            Error::CannotEnumerateAllFieldPaths(s)
                        }
                        _ => unreachable!(),
                    })?;

                // Throw an error if the schema has a polymorphic object other than just null or missing polymorphism. To figure out which field is the problem,
                // we check to see which field path is a prefix of another field path.
                if !has_only_nullable_polymorphism {
                    field_paths
                        .iter()
                        .flat_map(|p1| {
                            field_paths.iter().map(|p2| {
                                if p1.clone() != p2.clone() && !(p1.is_empty() || p2.is_empty()) {
                                    // Find which field has a polymorphic object schema.
                                    if p1.starts_with(p2.as_slice()) {
                                        return Err(Error::PolymorphicObjectSchema(
                                            p2.join(separator),
                                        ));
                                    } else if p2.starts_with(p1.as_slice()) {
                                        return Err(Error::PolymorphicObjectSchema(
                                            p1.join(separator),
                                        ));
                                    }
                                };
                                Ok(())
                            })
                        })
                        .collect::<Result<Vec<_>>>()?;
                }

                // Check to see if any field path is a prefix of another field path. If there is a prefix, remove it.
                // Note: Any prefixes found are the result of objects that can be null or missing.
                let mut field_paths_copy = field_paths.clone();
                for p1 in field_paths.iter() {
                    for p2 in field_paths.iter() {
                        if p1.clone() != p2.clone() && !p1.is_empty() && !p2.is_empty() {
                            if p1.starts_with(p2.as_slice()) {
                                field_paths_copy.remove(p2);
                            } else if p2.starts_with(p1.as_slice()) {
                                field_paths_copy.remove(p1);
                            }
                        }
                    }
                }

                let mut project_expression = UniqueLinkedHashMap::new();
                let mut sub_schema_env = SchemaEnvironment::new();
                sub_schema_env.insert(key.clone(), schema);
                let path_algebrizer = Algebrizer::with_schema_env(
                    self.current_db,
                    sub_schema_env,
                    self.catalog,
                    self.scope_level,
                    self.schema_checking_mode,
                    self.allow_order_by_missing_columns,
                    *self.clause_type.borrow(),
                );
                project_expression
                    .insert_many(field_paths_copy.into_iter().map(|path| {
                        (
                            path.join(separator),
                            path_algebrizer.algebrize_flattened_field_path(key.clone(), path),
                        )
                    }))
                    .map_err(|e| Error::DuplicateDocumentKey(e.get_key_name()))?;
                Ok((key, mir::Expression::Document(project_expression.into())))
            })
            .collect::<Result<BindingTuple<mir::Expression>>>()?;

        // Build the Project stage using the source and built expression
        let stage = mir::Stage::Project(mir::Project {
            source: Box::new(source),
            expression,
            is_add_fields: false,
            cache: SchemaCache::new(),
        });
        stage.schema(&self.schema_inference_state())?;
        Ok(stage)
    }

    fn algebrize_unwind_datasource(&self, u: ast::UnwindSource) -> Result<mir::Stage> {
        let src = self.algebrize_datasource(*u.datasource)?;

        // Extract user-specified options. OUTER defaults to false.
        #[allow(clippy::manual_try_fold)]
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
            self.allow_order_by_missing_columns,
            *self.clause_type.borrow(),
        );

        let path = match path {
            None => return Err(Error::NoUnwindPath),
            Some(e) => path_expression_algebrizer.algebrize_unwind_path(e)?,
        };

        let stage = mir::Stage::Unwind(mir::Unwind {
            source: Box::new(src),
            path,
            index,
            outer: outer.unwrap_or(false),
            cache: Default::default(),
            is_prefiltered: false,
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
    fn algebrize_unwind_path(&self, path: ast::Expression) -> Result<mir::FieldPath> {
        let path = self.algebrize_expression(path, false)?;
        (&path).try_into().map_err(|_| Error::InvalidUnwindPath)
    }

    #[allow(clippy::only_used_in_recursion)] // false positive
    pub fn algebrize_flattened_field_path(&self, key: Key, path: Vec<String>) -> mir::Expression {
        match path.len() {
            0 => mir::Expression::Reference(mir::ReferenceExpr { key }),
            _ => self
                .construct_field_access_expr(
                    self.algebrize_flattened_field_path(key, path.split_last().unwrap().1.to_vec()),
                    path.last().unwrap().to_string(),
                )
                .unwrap(),
        }
    }

    pub fn algebrize_having_clause(
        &self,
        ast_node: Option<ast::Expression>,
        source: mir::Stage,
    ) -> Result<mir::Stage> {
        *self.clause_type.borrow_mut() = ClauseType::Having;
        self.algebrize_filter_clause(ast_node, source)
    }

    pub fn algebrize_where_clause(
        &self,
        ast_node: Option<ast::Expression>,
        source: mir::Stage,
    ) -> Result<mir::Stage> {
        *self.clause_type.borrow_mut() = ClauseType::Where;
        self.algebrize_filter_clause(ast_node, source)
    }

    pub fn algebrize_filter_clause(
        &self,
        ast_node: Option<ast::Expression>,
        source: mir::Stage,
    ) -> Result<mir::Stage> {
        let filtered = match ast_node {
            None => source,
            Some(expr) => {
                let expression_algebrizer = self.clone().with_merged_mappings(
                    source.schema(&self.schema_inference_state())?.schema_env,
                )?;
                mir::Stage::Filter(mir::Filter {
                    source: Box::new(source),
                    condition: expression_algebrizer
                        .algebrize_expression(expr, false)
                        .map(Self::convert_literal_to_bool)?,
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
        source: mir::Stage,
        is_add_fields: bool,
    ) -> Result<mir::Stage> {
        *self.clause_type.borrow_mut() = ClauseType::Select;

        match ast_node.set_quantifier {
            ast::SetQuantifier::All => match ast_node.body {
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
                ast::SelectBody::Values(exprs) => {
                    self.algebrize_select_values_body(exprs, source, is_add_fields, false)
                }
            },
            ast::SetQuantifier::Distinct => match ast_node.body {
                ast::SelectBody::Standard(exprs) => match exprs.as_slice() {
                    [ast::SelectExpression::Star] => {
                        self.build_distinct_star_group_and_project_stage(source)
                    }
                    _ => Err(Error::NonStarStandardSelectBody),
                },
                ast::SelectBody::Values(exprs) => {
                    self.algebrize_select_values_body(exprs, source, is_add_fields, true)
                }
            },
        }
    }

    fn build_distinct_star_group_and_project_stage(
        &self,
        source: mir::Stage,
    ) -> Result<mir::Stage> {
        let source_result_set = source.schema(&self.schema_inference_state())?;

        let mut datasource_counter = 0;
        let mut datasources = Vec::new();

        // Identify data sources from the result set schema
        for key in source_result_set.schema_env.keys() {
            if key.scope == self.scope_level {
                let group_key_name = format!("__groupKey{datasource_counter}");
                datasources.push((group_key_name, key.clone()));
                datasource_counter += 1;
            }
        }
        let group_keys: Vec<OptionallyAliasedExpr> = datasources
            .iter()
            .map(|(group_key_name, key)| {
                OptionallyAliasedExpr::Aliased(AliasedExpr {
                    alias: group_key_name.clone(),
                    expr: Expression::Reference(ReferenceExpr { key: key.clone() }),
                })
            })
            .collect();

        // Adding the group stage to deduplicate rows
        let group_stage = mir::Stage::Group(mir::Group {
            source: Box::new(source),
            keys: group_keys,
            aggregations: vec![],
            cache: SchemaCache::new(),
            scope: self.scope_level,
        });

        let mut expression = BindingTuple::new();
        for (group_key_name, key) in &datasources {
            expression.insert(
                key.clone(),
                Expression::FieldAccess(FieldAccess {
                    // Group stages always output fields under the Bottom datasource
                    expr: Box::new(Expression::Reference(ReferenceExpr {
                        key: Key::bot(self.scope_level),
                    })),
                    field: group_key_name.clone().to_string(),
                    // Setting is_nullable to true because the result set coming into the
                    // SELECT DISTINCT clause could be empty.
                    is_nullable: true,
                }),
            );
        }
        // Create the final project stage to output the original data sources
        let project = mir::Stage::Project(mir::Project {
            source: Box::new(group_stage),
            expression,
            is_add_fields: false,
            cache: SchemaCache::new(),
        });
        Ok(project)
    }

    // This function is used to algebrize a Select and Order By clause together. This is necessary
    // because the Order By clause can reference keys that are defined in the Select clause, but
    // also columns from before the Select clause. Because of this, we algebrized the Select using
    // a mir Project stage with is_add_fields set to true, so that we can still access the fields
    // defined before the Select in the Order By.
    pub fn algebrize_select_and_order_by_clause(
        &self,
        select_node: ast::SelectClause,
        order_by_node: Option<ast::OrderByClause>,
        source: mir::Stage,
    ) -> Result<mir::Stage> {
        // If there is no Order By, we can just treat this as a normal Select Clause when
        // allow_order_by_missing_columns is not set.
        if order_by_node.is_none() {
            return self.algebrize_select_clause(select_node, source, false);
        }
        let select = self.algebrize_select_clause(select_node, source, true)?;
        // The project_body must maintain the expressions defined by the $addFields or else
        // the output may include extraneous fields
        let project_body = match select {
            mir::Stage::Project(ref p) => p
                .expression
                .iter()
                .map(|(k, e)| (k.clone(), e.clone()))
                .collect(),
            // We only reach this case when we have SELECT *, so we can just return
            // algebrize_order_by
            _ => {
                return self.algebrize_order_by_clause(order_by_node, select);
            }
        };
        let ordered = self.algebrize_order_by_clause(order_by_node, select)?;
        Ok(mir::Stage::Project(mir::Project {
            source: Box::new(ordered),
            expression: project_body,
            is_add_fields: false,
            cache: SchemaCache::new(),
        }))
    }

    pub fn algebrize_order_by_clause(
        &self,
        ast_node: Option<ast::OrderByClause>,
        source: mir::Stage,
    ) -> Result<mir::Stage> {
        *self.clause_type.borrow_mut() = ClauseType::OrderBy;
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
                        let sort_key = match s.key {
                            ast::SortKey::Simple(expr) => {
                                expression_algebrizer.algebrize_expression(expr, false)
                            }
                            ast::SortKey::Positional(_) => panic!(
                                "positional sort keys should have been rewritten to references"
                            ),
                        }?;
                        match s.direction {
                            ast::SortDirection::Asc => Ok(mir::SortSpecification::Asc(
                                sort_key
                                    .clone()
                                    .try_into()
                                    .map_err(|_| Error::InvalidSortKey(sort_key))?,
                            )),
                            ast::SortDirection::Desc => Ok(mir::SortSpecification::Desc(
                                sort_key
                                    .clone()
                                    .try_into()
                                    .map_err(|_| Error::InvalidSortKey(sort_key))?,
                            )),
                        }
                    })
                    .collect::<Result<Vec<mir::SortSpecification>>>()?;
                mir::Stage::Sort(mir::Sort {
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
        source: mir::Stage,
    ) -> Result<mir::Stage> {
        *self.clause_type.borrow_mut() = ClauseType::GroupBy;
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
                            Ok(mir::OptionallyAliasedExpr::Aliased(mir::AliasedExpr {
                                alias: ast_key.alias,
                                expr: expression_algebrizer
                                    .algebrize_expression(ast_key.expr, false)?,
                            }))
                        }
                        ast::OptionallyAliasedExpr::Unaliased(expr) => expression_algebrizer
                            .algebrize_expression(expr, false)
                            .map(mir::OptionallyAliasedExpr::Unaliased),
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
                        Ok(mir::AliasedAggregation {
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

                mir::Stage::Group(mir::Group {
                    source: Box::new(source),
                    keys,
                    aggregations,
                    cache: SchemaCache::new(),
                    scope: self.scope_level,
                })
            }
        };

        grouped.schema(&self.schema_inference_state())?;
        Ok(grouped)
    }

    pub fn algebrize_aggregation(&self, f: ast::FunctionExpr) -> Result<mir::AggregationExpr> {
        let (distinct, function) = if f.function == ast::FunctionName::AddToSet {
            (true, ast::FunctionName::AddToArray)
        } else {
            (
                f.set_quantifier == Some(ast::SetQuantifier::Distinct),
                f.function,
            )
        };
        let mir_node = match f.args {
            ast::FunctionArguments::Star => {
                if f.function == ast::FunctionName::Count {
                    return Ok(mir::AggregationExpr::CountStar(distinct));
                }
                return Err(Error::StarInNonCount);
            }
            ast::FunctionArguments::Args(ve) => {
                let arg = if ve.len() != 1 {
                    return Err(Error::AggregationFunctionMustHaveOneArgument);
                } else {
                    self.algebrize_expression(ve[0].clone(), false)?
                };
                mir::AggregationExpr::Function(mir::AggregationFunctionApplication {
                    function: mir::AggregationFunction::try_from(function)?,
                    arg: Box::new(arg.clone()),
                    distinct,
                    arg_is_possibly_doc: arg
                        .schema(&self.schema_inference_state())?
                        .satisfies(&ANY_DOCUMENT.clone()),
                })
            }
        };

        Ok(mir_node)
    }

    pub fn algebrize_expression(
        &self,
        ast_node: ast::Expression,
        in_implicit_type_conversion_context: bool,
    ) -> Result<mir::Expression> {
        match ast_node {
            ast::Expression::Literal(l) => Ok(mir::Expression::Literal(self.algebrize_literal(l))),
            ast::Expression::StringConstructor(s) => {
                Ok(self.algebrize_string_constructor(s, in_implicit_type_conversion_context))
            }
            ast::Expression::Array(a) => Ok(mir::Expression::Array(
                a.into_iter()
                    .map(|e| self.algebrize_expression(e, false))
                    .collect::<Result<Vec<mir::Expression>>>()?
                    .into(),
            )),
            ast::Expression::Document(d) => Ok(mir::Expression::Document({
                let algebrized = d
                    .into_iter()
                    .map(|kv| Ok((kv.key, self.algebrize_expression(kv.value, false)?)))
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
            ast::Expression::DateFunction(d) => self.algebrize_date_function(d),
            ast::Expression::Extract(e) => self.algebrize_extract(e),
            ast::Expression::Access(a) => self.algebrize_access(a),
            ast::Expression::Case(c) => self.algebrize_case(c),
            ast::Expression::Cast(c) => self.algebrize_cast(c),
            ast::Expression::TypeAssertion(t) => self.algebrize_type_assertion(t),
            ast::Expression::Is(i) => self.algebrize_is(i),
            ast::Expression::Like(l) => self.algebrize_like(l),
            // Tuples should all be rewritten away.
            ast::Expression::Tuple(_) => panic!("tuples cannot be algebrized"),
            ast::Expression::Subquery(s) => self.algebrize_subquery(*s),
            ast::Expression::SubqueryComparison(s) => self.algebrize_subquery_comparison(s),
            ast::Expression::Exists(e) => self.algebrize_exists(*e),
        }
    }

    pub fn algebrize_literal(&self, ast_node: ast::Literal) -> mir::LiteralValue {
        match ast_node {
            ast::Literal::Null => mir::LiteralValue::Null,
            ast::Literal::Boolean(b) => mir::LiteralValue::Boolean(b),
            ast::Literal::Integer(i) => mir::LiteralValue::Integer(i),
            ast::Literal::Long(l) => mir::LiteralValue::Long(l),
            ast::Literal::Double(d) => mir::LiteralValue::Double(d),
        }
    }

    pub fn algebrize_string_constructor(
        &self,
        s: String,
        in_implicit_type_conversion_context: bool,
    ) -> mir::Expression {
        // treat the string as a literal if we are not in a context where we want to implicitly cast
        if !in_implicit_type_conversion_context {
            return mir::Expression::Literal(mir::LiteralValue::String(s));
        }

        // if we are in a context where we want to implicit cast, use serde_json to
        // try to interpret the string as extended json. If the deserialization is successful
        // (i.e., the string contained extended json), convert the bson to a mir expression; otherwise,
        // return the original string as a string literal
        match serde_json::from_str::<bson::Bson>(s.as_str()) {
            Ok(bson) => bson.into(),
            Err(_) => mir::Expression::Literal(mir::LiteralValue::String(s)),
        }
    }

    pub fn algebrize_limit_clause(
        &self,
        ast_node: Option<u32>,
        source: mir::Stage,
    ) -> Result<mir::Stage> {
        *self.clause_type.borrow_mut() = ClauseType::Limit;
        match ast_node {
            None => Ok(source),
            Some(x) => {
                let stage = mir::Stage::Limit(mir::Limit {
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
        source: mir::Stage,
    ) -> Result<mir::Stage> {
        *self.clause_type.borrow_mut() = ClauseType::Offset;
        match ast_node {
            None => Ok(source),
            Some(x) => {
                let stage = mir::Stage::Offset(mir::Offset {
                    source: Box::new(source),
                    offset: i64::from(x),
                    cache: SchemaCache::new(),
                });
                stage.schema(&self.schema_inference_state())?;
                Ok(stage)
            }
        }
    }

    /// This is a helper function for algebrizing non-literal comparison operands so that
    /// we can determine if literal operands in those comparisons need to be algebrized in
    /// implicit type conversion contexts. This function returns the algebrized expression
    /// as well as a boolean flag indicating if the expression MUST satisfy STRING_OR_NULLISH.
    fn algebrize_non_literal_itc_operand(
        &self,
        non_literal: ast::Expression,
    ) -> Result<(mir::Expression, bool)> {
        let non_literal = self.algebrize_expression(non_literal, false)?;
        let non_literal_schema = non_literal.schema(&self.schema_inference_state())?;
        let is_nullable_string =
            non_literal_schema.satisfies(&STRING_OR_NULLISH) == Satisfaction::Must;

        Ok((non_literal, is_nullable_string))
    }

    /// This is a helper function for algebrizing binary comparison operands when one is a
    /// StringConstructor (literal) and one is not (non_literal). The non_literal is algebrized
    /// first. If its schema MUST satisfy STRING_OR_NULLISH, then the literal is algebrized with
    /// in_implicit_type_conversion_context set to false because that means a String is expected;
    /// otherwise, it is algebrized with that value set to true.
    ///
    /// - "itc" stands for "implicit type conversion".
    /// - The tuple returned is always (literal, non_literal).
    fn algebrize_itc_eligible_binary_comparison_operands(
        &self,
        literal: ast::Expression,
        non_literal: ast::Expression,
    ) -> Result<(mir::Expression, mir::Expression)> {
        let (non_literal, is_nullable_string) =
            self.algebrize_non_literal_itc_operand(non_literal)?;

        // Again, if is_nullable_string is true, that means we _do_ expect the StringConstructor
        // to be a String value, so we set in_implicit_type_conversion_context to false.
        let literal = self.algebrize_expression(literal, !is_nullable_string)?;

        Ok((literal, non_literal))
    }

    /// This is a helper function for algebrizing binary comparison operands.
    /// This applies to not only BinaryExprs but also the NullIf Function and
    /// other expressions that perform a binary comparison. This function
    /// determines if exactly one of the operands is a StringConstructor and
    /// dispatches to algebrize_itc_eligible_binary_comparison_operands if
    /// that is the case. Otherwise, this function algebrizes both operands
    /// with in_implicit_type_conversion_context set to false. This is because
    /// if both are StringConstructors, we can leave them both as String values,
    /// and if neither are StringConstructors, there is nothing to convert.
    fn algebrize_binary_comparison_operands(
        &self,
        left: ast::Expression,
        right: ast::Expression,
    ) -> Result<(mir::Expression, mir::Expression)> {
        match (left, right) {
            (
                l @ ast::Expression::StringConstructor(_),
                r @ ast::Expression::StringConstructor(_),
            ) => Ok((
                self.algebrize_expression(l, false)?,
                self.algebrize_expression(r, false)?,
            )),
            (literal @ ast::Expression::StringConstructor(_), non_literal) => {
                let (literal, non_literal) =
                    self.algebrize_itc_eligible_binary_comparison_operands(literal, non_literal)?;
                Ok((literal, non_literal))
            }
            (non_literal, literal @ ast::Expression::StringConstructor(_)) => {
                let (literal, non_literal) =
                    self.algebrize_itc_eligible_binary_comparison_operands(literal, non_literal)?;
                Ok((non_literal, literal))
            }
            (l, r) => Ok((
                self.algebrize_expression(l, false)?,
                self.algebrize_expression(r, false)?,
            )),
        }
    }

    fn algebrize_function(&self, f: ast::FunctionExpr) -> Result<mir::Expression> {
        if f.set_quantifier == Some(ast::SetQuantifier::Distinct) {
            return Err(Error::DistinctScalarFunction);
        }

        // get the arguments as a vec of ast::Expressions. If the arguments are
        // Star this must be a COUNT function, otherwise it is an error.
        let args = match f.clone().args {
            ast::FunctionArguments::Star => return Err(Error::StarInNonCount),
            ast::FunctionArguments::Args(ve) => ve,
        };

        // When algebrizing function arguments below, we set in_implicit_type_conversion_context to true whenever a String argument is unexpected,
        // and false wherever a String argument is expected.
        let args = match (f.function, args.len()) {
            // if the function is CURRENT_TIMESTAMP with exactly one arg,
            // throw away the argument. we break the spec intentionally
            // here by ignoring the date-precision argument. this is
            // implemented during algebrization instead of rewriting
            // because all other rewrites are compliant with the spec, and
            // this would be the only non-spec-compliant rewrite.
            (ast::FunctionName::CurrentTimestamp, 1) => Vec::new(),
            (ast::FunctionName::Abs, _)
            | (ast::FunctionName::Ceil, _)
            | (ast::FunctionName::Coalesce, _)
            | (ast::FunctionName::Cos, _)
            | (ast::FunctionName::CurrentTimestamp, _)
            | (ast::FunctionName::Degrees, _)
            | (ast::FunctionName::Floor, _)
            | (ast::FunctionName::Log, _)
            | (ast::FunctionName::Mod, _)
            | (ast::FunctionName::Pow, _)
            | (ast::FunctionName::Radians, _)
            | (ast::FunctionName::Sin, _)
            | (ast::FunctionName::Size, _)
            | (ast::FunctionName::Slice, _)
            | (ast::FunctionName::Sqrt, _)
            | (ast::FunctionName::Tan, _)
            | (ast::FunctionName::Round, _)
            | (ast::FunctionName::DayOfWeek, _) => args
                .into_iter()
                .map(|e| self.algebrize_expression(e, true))
                .collect::<Result<Vec<_>>>()?,
            (ast::FunctionName::Split, 3) => {
                let [string, delimiter, token_number]: [ast::Expression; 3] = args
                    .try_into()
                    .expect("Could not unpack args for ast Split function");
                vec![
                    self.algebrize_expression(string, false)?,
                    self.algebrize_expression(delimiter, false)?,
                    self.algebrize_expression(token_number, true)?,
                ]
            }
            (ast::FunctionName::Substring, 3) => {
                let [string, start, length]: [ast::Expression; 3] = args
                    .try_into()
                    .expect("Could not unpack args for ast Substring function");
                vec![
                    self.algebrize_expression(string, false)?,
                    self.algebrize_expression(start, true)?,
                    self.algebrize_expression(length, true)?,
                ]
            }
            (ast::FunctionName::NullIf, 2) => {
                let [v1, v2]: [ast::Expression; 2] = args
                    .try_into()
                    .expect("Could not unpack args for ast NullIf function");
                let (expr1, expr2) = self.algebrize_binary_comparison_operands(v1, v2)?;
                vec![expr1, expr2]
            }
            (ast::FunctionName::Split, _)
            | (ast::FunctionName::Substring, _)
            | (ast::FunctionName::NullIf, _)
            | (ast::FunctionName::BitLength, _)
            | (ast::FunctionName::CharLength, _)
            | (ast::FunctionName::Lower, _)
            | (ast::FunctionName::OctetLength, _)
            | (ast::FunctionName::Position, _)
            | (ast::FunctionName::Replace, _)
            | (ast::FunctionName::Upper, _) => args
                .into_iter()
                .map(|e| self.algebrize_expression(e, false))
                .collect::<Result<Vec<_>>>()?,
            // the following patterns should not be hit due to rewrites; throw an error for aggregation functions
            // to indicate programmer error, otherwise panic
            (ast::FunctionName::AddToArray, _)
            | (ast::FunctionName::AddToSet, _)
            | (ast::FunctionName::Avg, _)
            | (ast::FunctionName::Count, _)
            | (ast::FunctionName::First, _)
            | (ast::FunctionName::Last, _)
            | (ast::FunctionName::Max, _)
            | (ast::FunctionName::MergeDocuments, _)
            | (ast::FunctionName::Min, _)
            | (ast::FunctionName::StddevPop, _)
            | (ast::FunctionName::StddevSamp, _)
            | (ast::FunctionName::Sum, _) => {
                return Err(Error::AggregationInPlaceOfScalar(f.pretty_print().unwrap()))
            }
            (ast::FunctionName::LTrim, _)
            | (ast::FunctionName::RTrim, _)
            | (ast::FunctionName::Log10, _)
            | (ast::FunctionName::DateAdd, _)
            | (ast::FunctionName::DateDiff, _)
            | (ast::FunctionName::DateTrunc, _)
            | (ast::FunctionName::Year, _)
            | (ast::FunctionName::Month, _)
            | (ast::FunctionName::Week, _)
            | (ast::FunctionName::DayOfMonth, _)
            | (ast::FunctionName::DayOfYear, _)
            | (ast::FunctionName::Hour, _)
            | (ast::FunctionName::Minute, _)
            | (ast::FunctionName::Second, _)
            | (ast::FunctionName::Millisecond, _) => {
                unreachable!("{:?} should have been rewritten", f.function)
            }
        };

        let function = mir::ScalarFunction::try_from(f.function)?;
        let is_nullable = Self::determine_scalar_function_nullability(function, &args);

        // here we don't use the new constructor because we're setting the
        // proper nullability
        Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function,
                is_nullable,
                args,
            },
        ))
    }

    fn algebrize_unary_expr(&self, u: ast::UnaryExpr) -> Result<mir::Expression> {
        let arg = self.algebrize_expression(*u.expr, true)?;
        let args = vec![arg];
        Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function: mir::ScalarFunction::from(u.op),
                is_nullable: Self::args_are_nullable(&args),
                args,
            },
        ))
    }

    fn convert_literal_to_bool(expr: mir::Expression) -> mir::Expression {
        match expr {
            mir::Expression::Literal(mir::LiteralValue::Integer(i)) => match i {
                0 => mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
                1 => mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
                _ => mir::Expression::Literal(mir::LiteralValue::Integer(i)),
            },
            _ => expr,
        }
    }

    fn algebrize_binary_expr(&self, b: ast::BinaryExpr) -> Result<mir::Expression> {
        use crate::ast::BinaryOp::*;

        // First, we must determine if the left and right operands each need to
        // be algebrized in an implicit type conversion context (this is done
        // by toggling the bool argument to aglebrize_expression). The different
        // cases are detailed below.
        let (mut left, mut right) = match b.op {
            // Add, And, Div, Mul, Or, and Sub do not expect String operands,
            // therefore we algebrize their left and right operands with true.
            // This means we _should_ attempt to implicitly convert any
            // StringConstructors into different literal types.
            Add | And | Div | Mul | Or | Sub => (
                self.algebrize_expression(*b.left, true)?,
                self.algebrize_expression(*b.right, true)?,
            ),

            // Concat expects String operands, therefore we algebrize its left
            // and right operands with false. This means we should not attempt
            // to convert any StringConstructors into different literal types.
            Concat => (
                self.algebrize_expression(*b.left, false)?,
                self.algebrize_expression(*b.right, false)?,
            ),

            Comparison(_) => self.algebrize_binary_comparison_operands(*b.left, *b.right)?,

            // In and NotIn should have been rewritten during ast rewriting.
            In | NotIn => panic!("'{}' cannot be algebrized", b.op.as_str()),
        };

        let mut cast_div_result: Option<mir::Type> = None;

        // Since we want to avoid schema checking when at all possible, we will
        // only check for boolean conversion of literal 1/0 for comparison ops, AND, and
        // OR since those are the only place it is valid. Here, we also check the
        // schema of the arguments for Div to see if we need to cast the result to a whole
        // number in order to ensure integer division (rather than normal division) takes place.
        match b.op {
            Comparison(_) => {
                let (left_schema, right_schema) = (
                    left.schema(&self.schema_inference_state())?,
                    right.schema(&self.schema_inference_state())?,
                );

                // Here we convert integer literals of 1 or 0 to the proper True or False,
                // as necessary.
                if left_schema.satisfies(&BOOLEAN_OR_NULLISH) == Satisfaction::Must {
                    right = Self::convert_literal_to_bool(right);
                }
                if right_schema.satisfies(&BOOLEAN_OR_NULLISH) == Satisfaction::Must {
                    left = Self::convert_literal_to_bool(left);
                }
            }
            // And and Or only work with boolean types, so converting 1 to true and 0 to false
            // is always correct, unlike in comparisons.
            Or | And => {
                right = Self::convert_literal_to_bool(right);
                left = Self::convert_literal_to_bool(left);
            }
            // Check to see if both Div arguments MUST be whole numbers.
            Div => {
                let (left_schema, right_schema) = (
                    left.schema(&self.schema_inference_state())?,
                    right.schema(&self.schema_inference_state())?,
                );

                if left_schema.satisfies(&INTEGER_LONG_OR_NULLISH) == Satisfaction::Must
                    && right_schema.satisfies(&INTEGER_LONG_OR_NULLISH) == Satisfaction::Must
                {
                    if left_schema.satisfies(&INTEGER_OR_NULLISH) == Satisfaction::Must
                        && right_schema.satisfies(&INTEGER_OR_NULLISH) == Satisfaction::Must
                    {
                        cast_div_result = Some(mir::Type::Int32);
                    } else {
                        cast_div_result = Some(mir::Type::Int64);
                    }
                }
            }
            _ => (),
        }

        let args = vec![left, right];
        let function = mir::ScalarFunction::try_from(b.op)?;
        let is_nullable = Self::determine_scalar_function_nullability(function, &args);

        // here we don't use the new constructor because we're setting the
        // calculated nullability
        let scalar_function_expr =
            mir::Expression::ScalarFunction(mir::ScalarFunctionApplication {
                function,
                is_nullable,
                args,
            });

        if let Some(div_result_target_type) = cast_div_result {
            Ok(mir::Expression::Cast(mir::CastExpr {
                expr: Box::new(scalar_function_expr),
                to: div_result_target_type,
                on_null: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                on_error: Box::new(mir::Expression::Literal(mir::LiteralValue::Null)),
                is_nullable,
            }))
        } else {
            Ok(scalar_function_expr)
        }
    }

    fn algebrize_is(&self, ast_node: ast::IsExpr) -> Result<mir::Expression> {
        Ok(mir::Expression::Is(mir::IsExpr {
            expr: Box::new(self.algebrize_expression(*ast_node.expr, true)?),
            target_type: mir::TypeOrMissing::try_from(ast_node.target_type)?,
        }))
    }

    fn algebrize_like(&self, ast_node: ast::LikeExpr) -> Result<mir::Expression> {
        Ok(mir::Expression::Like(mir::LikeExpr {
            expr: Box::new(self.algebrize_expression(*ast_node.expr, false)?),
            pattern: Box::new(self.algebrize_expression(*ast_node.pattern, false)?),
            escape: ast_node.escape,
        }))
    }

    fn algebrize_between(&self, b: ast::BetweenExpr) -> Result<mir::Expression> {
        // First, we must determine if the arguments need to be algebrized in an
        // implicit type conversion context (this is done by toggling the bool
        // argument to algebrize_expression). We do this in two steps. First, we
        // algebrize the non-StringConstructor arguments and determine if all of
        // their schema MUST satisfy STRING_OR_NULLISH. Then, we algebrize any
        // StringConstructor operands using that information. If all non-literal
        // operand schema MUST be nullable Strings, we are not in an implicit
        // type conversion context (because Strings are expected); otherwise, we
        // are.
        let mut non_literals_are_nullable_strings = true;
        let mut arg_lit = "".to_string();
        let arg = match *b.arg {
            ast::Expression::StringConstructor(s) => {
                arg_lit = s;
                None
            }
            non_literal_arg => {
                let (non_literal_arg, arg_is_nullable_string) =
                    self.algebrize_non_literal_itc_operand(non_literal_arg)?;
                non_literals_are_nullable_strings =
                    non_literals_are_nullable_strings && arg_is_nullable_string;
                Some(non_literal_arg)
            }
        };
        let mut min_lit = "".to_string();
        let min = match *b.min {
            ast::Expression::StringConstructor(s) => {
                min_lit = s;
                None
            }
            non_literal_min => {
                let (non_literal_min, min_is_nullable_string) =
                    self.algebrize_non_literal_itc_operand(non_literal_min)?;
                non_literals_are_nullable_strings =
                    non_literals_are_nullable_strings && min_is_nullable_string;
                Some(non_literal_min)
            }
        };
        let mut max_lit = "".to_string();
        let max = match *b.max {
            ast::Expression::StringConstructor(s) => {
                max_lit = s;
                None
            }
            non_literal_max => {
                let (non_literal_max, max_is_nullable_string) =
                    self.algebrize_non_literal_itc_operand(non_literal_max)?;
                non_literals_are_nullable_strings =
                    non_literals_are_nullable_strings && max_is_nullable_string;
                Some(non_literal_max)
            }
        };

        // Again, if non_literals_are_nullable_strings is true, that means we
        // _do_ expect any StringConstructors to be String values, so we set
        // in_implicit_type_conversion_context to false.
        let in_implicit_type_conversion_context = !non_literals_are_nullable_strings;
        let args = vec![
            arg.unwrap_or(
                self.algebrize_string_constructor(arg_lit, in_implicit_type_conversion_context),
            ),
            min.unwrap_or(
                self.algebrize_string_constructor(min_lit, in_implicit_type_conversion_context),
            ),
            max.unwrap_or(
                self.algebrize_string_constructor(max_lit, in_implicit_type_conversion_context),
            ),
        ];

        let function = mir::ScalarFunction::Between;
        let is_nullable = Self::args_are_nullable(&args);
        Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function,
                is_nullable,
                args,
            },
        ))
    }

    fn algebrize_trim(&self, t: ast::TrimExpr) -> Result<mir::Expression> {
        let function = match t.trim_spec {
            ast::TrimSpec::Leading => mir::ScalarFunction::LTrim,
            ast::TrimSpec::Trailing => mir::ScalarFunction::RTrim,
            ast::TrimSpec::Both => mir::ScalarFunction::BTrim,
        };
        let args = vec![
            self.algebrize_expression(*t.trim_chars, false)?,
            self.algebrize_expression(*t.arg, false)?,
        ];
        Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function,
                is_nullable: Self::args_are_nullable(&args),
                args,
            },
        ))
    }

    fn algebrize_extract(&self, e: ast::ExtractExpr) -> Result<mir::Expression> {
        use crate::ast::DatePart::*;
        let function = match e.extract_spec {
            Year => mir::ScalarFunction::Year,
            Month => mir::ScalarFunction::Month,
            Day => mir::ScalarFunction::Day,
            Hour => mir::ScalarFunction::Hour,
            Minute => mir::ScalarFunction::Minute,
            Second => mir::ScalarFunction::Second,
            Millisecond => mir::ScalarFunction::Millisecond,
            Week => mir::ScalarFunction::Week,
            DayOfYear => mir::ScalarFunction::DayOfYear,
            DayOfWeek => mir::ScalarFunction::DayOfWeek,
            IsoWeek => mir::ScalarFunction::IsoWeek,
            IsoWeekday => mir::ScalarFunction::IsoWeekday,
            Quarter => panic!("'Quarter' is not a supported date part for EXTRACT"),
        };
        let args = vec![self.algebrize_expression(*e.arg, true)?];
        let is_nullable = Self::args_are_nullable(&args);
        Ok(mir::Expression::ScalarFunction(
            mir::ScalarFunctionApplication {
                function,
                is_nullable,
                args,
            },
        ))
    }

    fn algebrize_date_function(&self, d: ast::DateFunctionExpr) -> Result<mir::Expression> {
        use crate::ast::{DateFunctionName::*, DatePart::*};
        let function = match d.function {
            Add => mir::DateFunction::Add,
            Diff => mir::DateFunction::Diff,
            Trunc => mir::DateFunction::Trunc,
        };
        let date_part = match d.date_part {
            Year => mir::DatePart::Year,
            Month => mir::DatePart::Month,
            Day => mir::DatePart::Day,
            Hour => mir::DatePart::Hour,
            Minute => mir::DatePart::Minute,
            Second => mir::DatePart::Second,
            Millisecond => mir::DatePart::Millisecond,
            Week => mir::DatePart::Week,
            Quarter => mir::DatePart::Quarter,
            IsoWeek | IsoWeekday | DayOfYear | DayOfWeek => {
                panic!(
                    "'{0:?}' is not a supported date part for DATEADD, DATEDIFF, and DATETRUNC",
                    d.date_part
                )
            }
        };

        let args = d
            .args
            .into_iter()
            .map(|e| self.algebrize_expression(e, true))
            .collect::<Result<Vec<_>>>()?;

        // All date functions are nullable
        let is_nullable = Self::args_are_nullable(&args);

        Ok(mir::Expression::DateFunction(
            mir::DateFunctionApplication {
                function,
                is_nullable,
                date_part,
                args,
            },
        ))
    }

    fn algebrize_access(&self, a: ast::AccessExpr) -> Result<mir::Expression> {
        let expr = self.algebrize_expression(*a.expr, true)?;
        Ok(match *a.subfield {
            ast::Expression::StringConstructor(s) => self.construct_field_access_expr(expr, s)?,
            sf => mir::Expression::ScalarFunction(mir::ScalarFunctionApplication::new(
                mir::ScalarFunction::ComputedFieldAccess,
                vec![expr, self.algebrize_expression(sf, false)?],
            )),
        })
    }

    fn algebrize_type_assertion(&self, t: ast::TypeAssertionExpr) -> Result<mir::Expression> {
        // If the target_type is String, we do not implicitly convert the
        // expr since it is being asserted as a String. Otherwise, we do
        // attempt to convert.
        Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
            expr: Box::new(self.algebrize_expression(*t.expr, t.target_type != ast::Type::String)?),
            target_type: mir::Type::try_from(t.target_type)?,
        }))
    }

    fn algebrize_case(&self, c: ast::CaseExpr) -> Result<mir::Expression> {
        // if we don't have an else branch, the resulting case expression _is_ nullable; otherwise, we will
        // check the then and else expressions to determine nullability
        let mut is_nullable = c.else_branch.is_none();
        let else_branch = c
            .else_branch
            .map(|e| self.algebrize_expression(*e, false))
            .transpose()?
            .inspect(|expr| {
                is_nullable = is_nullable
                    || (NULLISH.satisfies(&expr.schema(&self.schema_inference_state()).unwrap())
                        != Satisfaction::Not);
            })
            .map(Box::new)
            .unwrap_or_else(|| Box::new(mir::Expression::Literal(mir::LiteralValue::Null)));

        // algebrize the when branches without implicit casting, keeping track of the schema of all of the when expressions
        // to inform how we algebrize the expr
        let mut when_branches_all_strings = true;
        let mut when_branch: Vec<mir::WhenBranch> = c
            .when_branch
            .clone()
            .into_iter()
            .map(|wb| {
                let when = self.algebrize_expression(*wb.when, false)?;
                let then = self.algebrize_expression(*wb.then, false)?;
                let then_is_nullable = NULLISH
                    .satisfies(&then.schema(&self.schema_inference_state()).unwrap())
                    != Satisfaction::Not;
                is_nullable = is_nullable || then_is_nullable;
                when_branches_all_strings = when_branches_all_strings
                    && (when
                        .schema(&self.schema_inference_state())?
                        .satisfies(&STRING_OR_NULLISH)
                        == Satisfaction::Must);
                Ok(mir::WhenBranch {
                    is_nullable: then_is_nullable,
                    when: Box::new(when),
                    then: Box::new(then),
                })
            })
            .collect::<Result<_>>()?;

        // if all of the when branches have string schemas, we should not implicitly convert our expr if we have one, so that they may
        // be compared directly. Otherwise, we algebrize the expr with implicit type conversion
        let expr = c
            .expr
            .map(|e| self.algebrize_expression(*e, !when_branches_all_strings))
            .transpose()?;

        // if the expr exists and MUST satisfy string, we can use the previously algebrized when_branch without implicit type conversion;
        // otherwise, we should re-algebrize with implicit type conversion for comparison with non-string expressions (SimpleCase), or for SearchedCase
        let expr_schema = expr.as_ref().map(|e| {
            e.schema(&self.schema_inference_state())
                .unwrap_or(schema::Schema::Unsat)
        });
        if expr_schema
            .as_ref()
            .is_none_or(|schema| schema.satisfies(&STRING_OR_NULLISH) != Satisfaction::Must)
        {
            when_branch = c
                .when_branch
                .into_iter()
                .map(|wb| {
                    let when = self.algebrize_expression(*wb.when, true)?;
                    let then = self.algebrize_expression(*wb.then, false)?;
                    Ok(mir::WhenBranch {
                        is_nullable: NULLISH
                            .satisfies(&then.schema(&self.schema_inference_state())?)
                            != Satisfaction::Not,
                        when: Box::new(when),
                        then: Box::new(then),
                    })
                })
                .collect::<Result<_>>()?;
        }

        match expr {
            Some(expr) => {
                let expr = Box::new(expr);
                Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
                    expr,
                    when_branch,
                    else_branch,
                    is_nullable,
                }))
            }
            None => Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr {
                when_branch,
                else_branch,
                is_nullable,
            })),
        }
    }

    fn algebrize_cast(&self, c: ast::CastExpr) -> Result<mir::Expression> {
        use crate::ast::Type::*;
        macro_rules! null_expr {
            () => {{
                Box::new(ast::Expression::Literal(ast::Literal::Null))
            }};
        }

        match c.to {
            BinData | DbPointer | Javascript | JavascriptWithScope | MaxKey | MinKey
            | RegularExpression | Symbol | Timestamp | Undefined | Date | Time => {
                Err(Error::InvalidCast(c.to))
            }
            Array | Boolean | Datetime | Decimal128 | Document | Double | Int32 | Int64 | Null
            | ObjectId | String => {
                let expr = self.algebrize_expression(*c.expr, true)?;
                let on_null =
                    self.algebrize_expression(*(c.on_null.unwrap_or_else(|| null_expr!())), false)?;
                let on_error = self
                    .algebrize_expression(*(c.on_error.unwrap_or_else(|| null_expr!())), false)?;
                let is_nullable =
                    expr.is_nullable() || on_error.is_nullable() || on_null.is_nullable();
                Ok(mir::Expression::Cast(mir::CastExpr {
                    expr: Box::new(expr),
                    to: mir::Type::try_from(c.to)?,
                    on_null: Box::new(on_null),
                    on_error: Box::new(on_error),
                    is_nullable,
                }))
            }
        }
    }

    pub fn algebrize_subquery_expr(
        &self,
        ast_node: ast::Query,
    ) -> Result<(mir::SubqueryExpr, schema::Schema)> {
        let subquery_algebrizer = self.subquery_algebrizer();
        let subquery = Box::new(subquery_algebrizer.algebrize_query(ast_node)?);
        let result_set = subquery.schema(&subquery_algebrizer.schema_inference_state())?;

        match result_set.schema_env.len() {
            1 => {
                let (key, schema) = result_set.schema_env.into_iter().next().unwrap();
                let (output_expr, output_expr_schema) =
                    match &schema.get_single_field_name_and_schema() {
                        Some((field_name, field_schema)) => Ok((
                            Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                                expr: Box::new(mir::Expression::Reference(key.into())),
                                field: field_name.to_string(),
                                is_nullable: NULLISH.satisfies(field_schema) != Satisfaction::Not,
                            })),
                            field_schema.clone(),
                        )),
                        None => Err(Error::InvalidSubqueryDegree),
                    }?;
                let is_nullable = output_expr.is_nullable();
                Ok((
                    mir::SubqueryExpr {
                        output_expr,
                        subquery,
                        is_nullable,
                    },
                    output_expr_schema,
                ))
            }
            _ => Err(Error::InvalidSubqueryDegree),
        }
    }

    pub fn algebrize_subquery(&self, ast_node: ast::Query) -> Result<mir::Expression> {
        let (subquery_expr, _) = self.algebrize_subquery_expr(ast_node)?;
        Ok(mir::Expression::Subquery(subquery_expr))
    }

    pub fn algebrize_subquery_comparison(
        &self,
        s: ast::SubqueryComparisonExpr,
    ) -> Result<mir::Expression> {
        let modifier = match s.quantifier {
            ast::SubqueryQuantifier::All => mir::SubqueryModifier::All,
            ast::SubqueryQuantifier::Any => mir::SubqueryModifier::Any,
        };
        // Algebrize the subquery expr and get the schema of the output field.
        let (subquery_expr, subquery_schema) = self.algebrize_subquery_expr(*s.subquery)?;

        // If the schema of the subquery output field MUST satisfy STRING_OR_NULLISH,
        // then we algebrize s.expr with in_implicit_type_conversion_context set to
        // false because that means a String is expected. Otherwise, it is algebrized
        // with that value set to true.
        let is_nullable_string =
            subquery_schema.satisfies(&STRING_OR_NULLISH) == Satisfaction::Must;
        let argument = self.algebrize_expression(*s.expr, !is_nullable_string)?;

        // Determine the overall nullability of the subquery comparison expr.
        let arg_schema = argument.schema(&self.schema_inference_state())?;
        let is_nullable =
            subquery_expr.is_nullable || NULLISH.satisfies(&arg_schema) != Satisfaction::Not;
        Ok(mir::Expression::SubqueryComparison(
            mir::SubqueryComparison {
                operator: mir::SubqueryComparisonOp::from(s.op),
                modifier,
                is_nullable,
                argument: Box::new(argument),
                subquery_expr,
            },
        ))
    }

    pub fn algebrize_exists(&self, ast_node: ast::Query) -> Result<mir::Expression> {
        let exists = self.subquery_algebrizer().algebrize_query(ast_node)?;
        Ok(mir::Expression::Exists(Box::new(exists).into()))
    }

    fn algebrize_subpath(&self, p: ast::SubpathExpr) -> Result<mir::Expression> {
        if let ast::Expression::Identifier(s) = *p.expr {
            return self.algebrize_possibly_qualified_field_access(s, p.subpath);
        }
        let expr = self.algebrize_expression(*p.expr, true)?;
        let expr_schema = expr.schema(&self.schema_inference_state())?;
        let is_nullable = NULLISH.satisfies(&expr_schema) != Satisfaction::Not
            || expr_schema.contains_field(&p.subpath) != Satisfaction::Must;
        Ok(mir::Expression::FieldAccess(mir::FieldAccess {
            expr: Box::new(expr),
            field: p.subpath,
            is_nullable,
        }))
    }

    fn algebrize_possibly_qualified_field_access(
        &self,
        q: String,
        field: String,
    ) -> Result<mir::Expression> {
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
                    let expr = self.algebrize_unqualified_identifier(q)?;
                    self.construct_field_access_expr(
                        expr,
                        // combinators make this clone necessary, unfortunately
                        cloned_field,
                    )
                },
                move |scope| {
                    // Since this is qualified, we return `q.field`
                    self.construct_field_access_expr(
                        mir::Expression::Reference(
                            Key {
                                datasource: possible_datasource,
                                scope,
                            }
                            .into(),
                        ),
                        field,
                    )
                },
            )
    }

    fn algebrize_unqualified_identifier(&self, i: String) -> Result<mir::Expression> {
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
            let all_keys = self
                .schema_env
                .clone()
                .into_iter()
                .flat_map(|(_, s)| s.keys())
                .collect::<Vec<_>>();

            let err = if all_keys.is_empty() {
                Error::FieldNotFound(i, None, *self.clause_type.borrow(), self.scope_level)
            } else {
                Error::FieldNotFound(
                    i,
                    Some(all_keys),
                    *self.clause_type.borrow(),
                    self.scope_level,
                )
            };

            return Err(err);
        }
        // If there is exactly one possible datasource that May or Must
        // contain our reference, we use it.
        if i_containing_datasources.len() == 1 {
            return self.construct_field_access_expr(
                mir::Expression::Reference(i_containing_datasources.remove(0).0.clone().into()),
                i,
            );
        }

        // Otherwise, we check datasources per scope, starting at the current scope,
        // to find the best datasource from multiple possible datasources.
        self.algebrize_unqualified_identifier_by_scope(i, self.scope_level)
    }

    fn algebrize_unqualified_identifier_by_scope(
        &self,
        i: String,
        scope_level: u16,
    ) -> Result<mir::Expression> {
        // When checking variables by scope, if a variable may exist, we treat that as ambiguous,
        // and only accept a single Must exist reference.
        let mut current_scope = scope_level;
        let mut found_bot = false;
        loop {
            let current_bot = Key::bot(current_scope);
            // Attempt to find a datasource for this reference in the current_scope.
            // If we find exactly one datasource Must contain the field `i`, we return
            // `datasource.i`. If there is more than one, it is an ambiguous error. As mentioned,
            // if there is a May exists, it is also an ambiguous variable error.
            // There is one caveat, if exactly two datasources Must satisfy, and one is bot, we return
            // bot as the datasource. For optimization we would prefer to return the non-bot source
            // since it would allow for more Sort movement, but it's possible the bot datasource
            // has updated the value or is otherwise completely indepenedent, and it is more recent
            // and shadows the non-bot datasource. This currently only happens when we allow
            // for ordering by columns not in the select list, but is safe in all contexts.
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
                            Satisfaction::Must => {
                                if curr_datasource == &current_bot {
                                    found_bot = true;
                                }
                                (curr_datasource, mays, musts + 1)
                            }
                            Satisfaction::May => {
                                if curr_datasource == &current_bot {
                                    found_bot = true;
                                }
                                (found_datasource, mays + 1, musts)
                            }
                            Satisfaction::Not => (found_datasource, mays, musts),
                        }
                    },
                );
            if musts == 1 && mays == 0 {
                return self.construct_field_access_expr(
                    mir::Expression::Reference(datasource.clone().into()),
                    i,
                );
            }
            if musts + mays == 2 {
                // It's actually impossible for bot to be the may here, since the SELECT
                // or GROUP BY always define the value. This could change if we ever allow
                // something like $$REMOVE in our SQL dialect.
                if found_bot {
                    return self.construct_field_access_expr(
                        mir::Expression::Reference(current_bot.into()),
                        i,
                    );
                }
                // if we have two May/Must datasources, and neither is bot, we return an ambiguous error
                return Err(Error::AmbiguousField(
                    i,
                    *self.clause_type.borrow(),
                    current_scope,
                ));
            }
            if mays > 0 || musts > 1 {
                return Err(Error::AmbiguousField(
                    i,
                    *self.clause_type.borrow(),
                    current_scope,
                ));
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

mod convert_to_bool {
    macro_rules! test_convert_literal_to_bool {
        ($func_name:ident, expected = $expected:expr, input = $input:expr) => {
            #[test]
            fn $func_name() {
                use super::*;
                use crate::algebrizer::Algebrizer;

                let actual = Algebrizer::convert_literal_to_bool($input);
                assert_eq!($expected, actual);
            }
        };
    }

    test_convert_literal_to_bool!(
        convert_one_to_true,
        expected = mir::Expression::Literal(mir::LiteralValue::Boolean(true)),
        input = mir::Expression::Literal(mir::LiteralValue::Integer(1))
    );

    test_convert_literal_to_bool!(
        convert_zero_to_false,
        expected = mir::Expression::Literal(mir::LiteralValue::Boolean(false)),
        input = mir::Expression::Literal(mir::LiteralValue::Integer(0))
    );

    test_convert_literal_to_bool!(
        non_bool_integer_not_converted,
        expected = mir::Expression::Literal(mir::LiteralValue::Integer(3)),
        input = mir::Expression::Literal(mir::LiteralValue::Integer(3))
    );

    test_convert_literal_to_bool!(
        non_integer_not_converted,
        expected = mir::Expression::Literal(mir::LiteralValue::String("should not change".into())),
        input = mir::Expression::Literal(mir::LiteralValue::String("should not change".into()))
    );
}
