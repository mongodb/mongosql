use crate::{
    algebrizer::errors::Error,
    ast::{self, pretty_print::PrettyPrint},
    catalog::Catalog,
    map,
    mir::{
        self,
        binding_tuple::{BindingTuple, DatasourceName, Key},
        schema::{CachedSchema, SchemaCache, SchemaInferenceState},
        FieldAccess, LiteralValue,
    },
    schema::{
        self, Satisfaction, SchemaEnvironment, BOOLEAN_OR_NULLISH, INTEGER_LONG_OR_NULLISH,
        INTEGER_OR_NULLISH, NULLISH,
    },
    util::unique_linked_hash_map::UniqueLinkedHashMap,
    SchemaCheckingMode,
};
use std::collections::BTreeSet;

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

#[derive(Debug, Clone)]
pub struct Algebrizer<'a> {
    current_db: &'a str,
    pub schema_env: SchemaEnvironment,
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
        let plan = self.algebrize_filter_clause(ast_node.where_clause, plan)?;
        let plan = self.algebrize_group_by_clause(ast_node.group_by_clause, plan)?;
        let plan = self.algebrize_filter_clause(ast_node.having_clause, plan)?;
        let plan = self.algebrize_select_clause(ast_node.select_clause, plan)?;
        let plan = self.algebrize_order_by_clause(ast_node.order_by_clause, plan)?;
        let plan = self.algebrize_offset_clause(ast_node.offset, plan)?;
        let plan = self.algebrize_limit_clause(ast_node.limit, plan)?;
        Ok(plan)
    }

    pub fn algebrize_set_query(&self, ast_node: ast::SetQuery) -> Result<mir::Stage> {
        match ast_node.op {
            ast::SetOperator::Union => Err(Error::DistinctUnion),
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
    ) -> Result<mir::Stage> {
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
                        let e = expression_algebrizer.algebrize_expression(e)?;
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
                .algebrize_expression(ast::Expression::Document(bottom.unwrap()))?;
            let bot = Key::bot(expression_algebrizer.scope_level);
            datasources
                .insert(bot.clone())
                .then_some(())
                .ok_or_else(|| Error::DuplicateKey(bot.clone()))?;
            expression.insert(bot, e);
        }

        // Build the Project Stage using the source and built expression.
        let stage = mir::Stage::Project(mir::Project {
            source: Box::new(source),
            expression,
            cache: SchemaCache::new(),
        });
        stage.schema(&self.schema_inference_state())?;
        Ok(stage)
    }

    pub fn algebrize_from_clause(&self, ast_node: Option<ast::Datasource>) -> Result<mir::Stage> {
        let ast_node = ast_node.expect("all SELECT queries must have a FROM clause");
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
                .map(|e| self.algebrize_expression(e))
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
            .map(|e| join_algebrizer.algebrize_expression(e))
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
            ast::JoinType::Cross | ast::JoinType::Inner => mir::Stage::Join(mir::Join {
                join_type: mir::JoinType::Inner,
                left: Box::new(left_src),
                right: Box::new(right_src),
                condition,
                cache: SchemaCache::new(),
            }),
        };
        Ok(stage)
    }

    fn algebrize_derived_datasource(&self, d: ast::DerivedSource) -> Result<mir::Stage> {
        let derived_algebrizer = Algebrizer::new(
            self.current_db,
            self.catalog,
            self.scope_level + 1,
            self.schema_checking_mode,
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
                        .collect::<Result<_>>()?;
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
        let path = self.algebrize_expression(path)?;
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
                        .algebrize_expression(expr)
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
    ) -> Result<mir::Stage> {
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
        source: mir::Stage,
    ) -> Result<mir::Stage> {
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
                                expression_algebrizer.algebrize_expression(expr)
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
                                expr: expression_algebrizer.algebrize_expression(ast_key.expr)?,
                            }))
                        }
                        ast::OptionallyAliasedExpr::Unaliased(expr) => expression_algebrizer
                            .algebrize_expression(expr)
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
                mir::AggregationExpr::Function(mir::AggregationFunctionApplication {
                    function: mir::AggregationFunction::try_from(function)?,
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

        Ok(mir_node)
    }

    pub fn algebrize_expression(&self, ast_node: ast::Expression) -> Result<mir::Expression> {
        match ast_node {
            ast::Expression::Literal(l) => Ok(mir::Expression::Literal(self.algebrize_literal(l))),
            ast::Expression::Array(a) => Ok(mir::Expression::Array(
                a.into_iter()
                    .map(|e| self.algebrize_expression(e))
                    .collect::<Result<Vec<mir::Expression>>>()?
                    .into(),
            )),
            ast::Expression::Document(d) => Ok(mir::Expression::Document({
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
            ast::Literal::String(s) => mir::LiteralValue::String(s),
            ast::Literal::Integer(i) => mir::LiteralValue::Integer(i),
            ast::Literal::Long(l) => mir::LiteralValue::Long(l),
            ast::Literal::Double(d) => mir::LiteralValue::Double(d),
        }
    }

    pub fn algebrize_limit_clause(
        &self,
        ast_node: Option<u32>,
        source: mir::Stage,
    ) -> Result<mir::Stage> {
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

    fn algebrize_function(&self, f: ast::FunctionExpr) -> Result<mir::Expression> {
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
            _ => args
                .into_iter()
                .map(|e| self.algebrize_expression(e))
                .collect::<Result<Vec<_>>>()?,
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
        let arg = self.algebrize_expression(*u.expr)?;
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
                0 => mir::Expression::Literal(LiteralValue::Boolean(false)),
                1 => mir::Expression::Literal(LiteralValue::Boolean(true)),
                _ => mir::Expression::Literal(LiteralValue::Integer(i)),
            },
            _ => expr,
        }
    }

    fn algebrize_binary_expr(&self, b: ast::BinaryExpr) -> Result<mir::Expression> {
        let (mut left, mut right) = (
            self.algebrize_expression(*b.left)?,
            self.algebrize_expression(*b.right)?,
        );

        let mut cast_div_result: Option<mir::Type> = None;

        // Since we want to avoid schema checking when at all possible, we will
        // only check for boolean conversion of literal 1/0 for comparison ops, AND, and
        // OR since those are the only place it is valid. Here, we also check the
        // schema of the arguments for Div to see if we need to cast the result to a whole
        // number in order to ensure integer division (rather than normal division) takes place.
        match b.op {
            ast::BinaryOp::Comparison(_) => {
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
            ast::BinaryOp::Or | ast::BinaryOp::And => {
                right = Self::convert_literal_to_bool(right);
                left = Self::convert_literal_to_bool(left);
            }
            // Check to see if both Div arguments MUST be whole numbers.
            ast::BinaryOp::Div => {
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
                on_null: Box::new(mir::Expression::Literal(LiteralValue::Null)),
                on_error: Box::new(mir::Expression::Literal(LiteralValue::Null)),
                is_nullable,
            }))
        } else {
            Ok(scalar_function_expr)
        }
    }

    fn algebrize_is(&self, ast_node: ast::IsExpr) -> Result<mir::Expression> {
        Ok(mir::Expression::Is(mir::IsExpr {
            expr: Box::new(self.algebrize_expression(*ast_node.expr)?),
            target_type: mir::TypeOrMissing::try_from(ast_node.target_type)?,
        }))
    }

    fn algebrize_like(&self, ast_node: ast::LikeExpr) -> Result<mir::Expression> {
        Ok(mir::Expression::Like(mir::LikeExpr {
            expr: Box::new(self.algebrize_expression(*ast_node.expr)?),
            pattern: Box::new(self.algebrize_expression(*ast_node.pattern)?),
            escape: ast_node.escape,
        }))
    }

    fn algebrize_between(&self, b: ast::BetweenExpr) -> Result<mir::Expression> {
        let (arg, min, max) = (
            self.algebrize_expression(*b.expr)?,
            self.algebrize_expression(*b.min)?,
            self.algebrize_expression(*b.max)?,
        );
        let args = vec![arg, min, max];
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
            self.algebrize_expression(*t.trim_chars)?,
            self.algebrize_expression(*t.arg)?,
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
        let args = vec![self.algebrize_expression(*e.arg)?];
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
            .map(|e| self.algebrize_expression(e))
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
        let expr = self.algebrize_expression(*a.expr)?;
        Ok(match *a.subfield {
            ast::Expression::Literal(ast::Literal::String(s)) => {
                self.construct_field_access_expr(expr, s)?
            }
            sf => mir::Expression::ScalarFunction(mir::ScalarFunctionApplication::new(
                mir::ScalarFunction::ComputedFieldAccess,
                vec![expr, self.algebrize_expression(sf)?],
            )),
        })
    }

    fn algebrize_type_assertion(&self, t: ast::TypeAssertionExpr) -> Result<mir::Expression> {
        Ok(mir::Expression::TypeAssertion(mir::TypeAssertionExpr {
            expr: Box::new(self.algebrize_expression(*t.expr)?),
            target_type: mir::Type::try_from(t.target_type)?,
        }))
    }

    fn algebrize_case(&self, c: ast::CaseExpr) -> Result<mir::Expression> {
        let else_branch = c
            .else_branch
            .map(|e| self.algebrize_expression(*e))
            .transpose()?
            .map(Box::new)
            .unwrap_or_else(|| Box::new(mir::Expression::Literal(mir::LiteralValue::Null)));
        let expr = c.expr.map(|e| self.algebrize_expression(*e)).transpose()?;
        let when_branch = c
            .when_branch
            .into_iter()
            .map(|wb| {
                let when = self.algebrize_expression(*wb.when)?;
                Ok(mir::WhenBranch {
                    is_nullable: NULLISH.satisfies(&when.schema(&self.schema_inference_state())?)
                        != Satisfaction::Not,
                    when: Box::new(when),
                    then: Box::new(self.algebrize_expression(*wb.then)?),
                })
            })
            .collect::<Result<_>>()?;
        match expr {
            Some(expr) => {
                let expr = Box::new(expr);
                let expr_schema = expr.schema(&self.schema_inference_state())?;
                Ok(mir::Expression::SimpleCase(mir::SimpleCaseExpr {
                    expr,
                    when_branch,
                    else_branch,
                    is_nullable: NULLISH.satisfies(&expr_schema) != Satisfaction::Not,
                }))
            }
            None => Ok(mir::Expression::SearchedCase(mir::SearchedCaseExpr::new(
                when_branch,
                else_branch,
            ))),
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
                let expr = self.algebrize_expression(*c.expr)?;
                let on_null =
                    self.algebrize_expression(*(c.on_null.unwrap_or_else(|| null_expr!())))?;
                let on_error =
                    self.algebrize_expression(*(c.on_error.unwrap_or_else(|| null_expr!())))?;
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

    pub fn algebrize_subquery_expr(&self, ast_node: ast::Query) -> Result<mir::SubqueryExpr> {
        let subquery_algebrizer = self.subquery_algebrizer();
        let subquery = Box::new(subquery_algebrizer.algebrize_query(ast_node)?);
        let result_set = subquery.schema(&subquery_algebrizer.schema_inference_state())?;

        match result_set.schema_env.len() {
            1 => {
                let (key, schema) = result_set.schema_env.into_iter().next().unwrap();
                let output_expr = match &schema.get_single_field_name() {
                    Some(field) => Ok(Box::new(mir::Expression::FieldAccess(mir::FieldAccess {
                        expr: Box::new(mir::Expression::Reference(key.into())),
                        field: field.to_string(),
                        is_nullable: NULLISH.satisfies(&schema) != Satisfaction::Not,
                    }))),
                    None => Err(Error::InvalidSubqueryDegree),
                }?;
                let is_nullable = output_expr.is_nullable();
                Ok(mir::SubqueryExpr {
                    output_expr,
                    subquery,
                    is_nullable,
                })
            }
            _ => Err(Error::InvalidSubqueryDegree),
        }
    }

    pub fn algebrize_subquery(&self, ast_node: ast::Query) -> Result<mir::Expression> {
        Ok(mir::Expression::Subquery(
            self.algebrize_subquery_expr(ast_node)?,
        ))
    }

    pub fn algebrize_subquery_comparison(
        &self,
        s: ast::SubqueryComparisonExpr,
    ) -> Result<mir::Expression> {
        let modifier = match s.quantifier {
            ast::SubqueryQuantifier::All => mir::SubqueryModifier::All,
            ast::SubqueryQuantifier::Any => mir::SubqueryModifier::Any,
        };
        let argument = self.algebrize_expression(*s.expr)?;
        let arg_schema = argument.schema(&self.schema_inference_state())?;
        let is_nullable = NULLISH.satisfies(&arg_schema) != Satisfaction::Not;
        Ok(mir::Expression::SubqueryComparison(
            mir::SubqueryComparison {
                operator: mir::SubqueryComparisonOp::from(s.op),
                modifier,
                is_nullable,
                argument: Box::new(argument),
                subquery_expr: self.algebrize_subquery_expr(*s.subquery)?,
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
        let expr = self.algebrize_expression(*p.expr)?;
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
                Error::FieldNotFound(i, None)
            } else {
                Error::FieldNotFound(i, Some(all_keys))
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
                return self.construct_field_access_expr(
                    mir::Expression::Reference(datasource.clone().into()),
                    i,
                );
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
