use crate::air;
use agg_ast::definitions::{
    Expression, GroupAccumulatorExpr, JoinType, LiteralValue, LookupFrom, ProjectItem, Stage,
    StringOrRef, Subquery, SubqueryExists, TaggedOperator, UntaggedOperator, Unwind,
};
use itertools::Itertools;
use linked_hash_map::LinkedHashMap;
use std::collections::HashMap;

fn translate_pipeline(root: Option<air::Stage>, pipeline: Vec<Stage>) -> Option<air::Stage> {
    if pipeline.is_empty() {
        return root;
    }
    pipeline
        .into_iter()
        .fold(root, |acc, curr| Some((acc, curr).into()))
}

impl From<(Option<air::Stage>, Stage)> for air::Stage {
    fn from((source, ast_stage): (Option<air::Stage>, Stage)) -> Self {
        match ast_stage {
            Stage::Collection(c) => air::Stage::Collection(air::Collection {
                db: c.db,
                collection: c.collection,
            }),
            Stage::Documents(d) => {
                let array: Vec<air::Expression> = d
                    .into_iter()
                    .map(|m| Expression::Document(m).into())
                    .collect();

                air::Stage::Documents(air::Documents { array })
            }
            Stage::Project(p) => {
                let specs: LinkedHashMap<String, air::ProjectItem> = p
                    .into_iter()
                    .sorted_by(|a, b| Ord::cmp(&a.0, &b.0))
                    .map(|(k, v)| (k, air::ProjectItem::from(v)))
                    // sort alphabetically for testing purposes
                    .collect::<_>();

                air::Stage::Project(air::Project {
                    source: Box::new(source.expect("$project without valid source stage")),
                    specifications: specs.into(),
                })
            }
            Stage::ReplaceWith(r) => air::Stage::ReplaceWith(air::ReplaceWith {
                source: Box::new(source.expect("$replaceWith without valid source stage")),
                new_root: Box::new(r.into()),
            }),
            Stage::Match(m) => air::Stage::Match(air::Match::ExprLanguage(air::ExprLanguage {
                source: Box::new(source.expect("$match without valid source stage")),
                expr: Box::new((*m.expr).into()),
            })),
            Stage::Limit(l) => air::Stage::Limit(air::Limit {
                source: Box::new(source.expect("$limit without valid source stage")),
                limit: l,
            }),
            Stage::Skip(s) => air::Stage::Skip(air::Skip {
                source: Box::new(source.expect("$skip without valid source stage")),
                skip: s,
            }),
            Stage::Sort(s) => {
                let specs = s
                    .into_iter()
                    // sort alphabetically for testing purposes -- this obviously changes sort semantics, however that doesn't matter for desugarer tests
                    .sorted_by(|a, b| Ord::cmp(&a.0, &b.0))
                    .map(|(k, v)| {
                        if v >= 1 {
                            air::SortSpecification::Asc(k)
                        } else if v <= -1 {
                            air::SortSpecification::Desc(k)
                        } else {
                            panic!("sort spec cannot be 0 but was for '{k}'")
                        }
                    })
                    .collect::<Vec<air::SortSpecification>>();

                air::Stage::Sort(air::Sort {
                    source: Box::new(source.expect("$sort without valid source stage")),
                    specs,
                })
            }
            Stage::Group(g) => {
                let keys = match g.keys {
                    // keys of form: _id: { name_1: expr_1, ... name_n, expr_n}
                    Expression::Document(d) => d
                        .into_iter()
                        .map(|(k, v)| air::NameExprPair {
                            name: k,
                            expr: v.into(),
                        })
                        .collect(),
                    // keys of form: _id: null
                    Expression::Literal(LiteralValue::Null) => vec![],
                    _ => panic!(),
                };
                let aggregations = g
                    .aggregations
                    .into_iter()
                    .sorted_by(|a, b| Ord::cmp(&a.0, &b.0))
                    .map(|(key, accumulator_expr)| match accumulator_expr.expr {
                        // accumulators of form: $<acc>: {"var": <expr>, "distinct": <bool>}
                        GroupAccumulatorExpr::SqlAccumulator { distinct, var } => {
                            air::AccumulatorExpr {
                                alias: key,
                                function: accumulator_expr.function.into(),
                                distinct,
                                arg: Box::new(air::Expression::from(*var)),
                            }
                        }
                        // accumulators of form: $<acc>: <expr>
                        GroupAccumulatorExpr::NonSqlAccumulator(expr) => air::AccumulatorExpr {
                            alias: key,
                            function: accumulator_expr.function.into(),
                            distinct: false,
                            arg: Box::new(expr.into()),
                        },
                    })
                    .collect();
                air::Stage::Group(air::Group {
                    source: Box::new(source.expect("$group without valid source stage")),
                    keys,
                    aggregations,
                })
            }
            Stage::Join(j) => {
                let join_type = match j.join_type {
                    JoinType::Inner => air::JoinType::Inner,
                    JoinType::Left => air::JoinType::Left,
                };
                let let_vars = j.let_body.map(|body| {
                    body.into_iter()
                        .map(|(k, v)| air::LetVariable {
                            name: k,
                            expr: Box::new(v.into()),
                        })
                        .collect()
                });

                // seed the right pipeline with the collection. If the db and/or collection aren't specified, use defaults
                let source_collection = make_optional_collection_stage(j.database, j.collection);

                // right source is the join pipeline
                let right = j
                    .pipeline
                    .into_iter()
                    .fold(source_collection, |air_stage, agg_ast_stage| {
                        Some(air::Stage::from((air_stage, agg_ast_stage)))
                    });

                air::Stage::Join(air::Join {
                    join_type,
                    left: Box::new(source.expect("$join without valid source stage")),
                    right: Box::new(right.unwrap()),
                    let_vars,
                    condition: j.condition.map(|expr| expr.into()),
                })
            }
            Stage::EquiJoin(eqj) => {
                let join_type = match eqj.join_type {
                    JoinType::Inner => air::JoinType::Inner,
                    JoinType::Left => air::JoinType::Left,
                };

                air::Stage::EquiJoin(air::EquiJoin {
                    join_type,
                    source: Box::new(source.expect("$equijoin without valid source stage")),
                    from: air::Collection {
                        db: eqj.database.unwrap_or_else(|| "test".to_string()),
                        collection: eqj.collection.unwrap_or_else(|| "default".to_string()),
                    },
                    local_field: eqj.local_field.into(),
                    foreign_field: eqj.foreign_field.into(),
                    as_name: eqj.as_var,
                })
            }
            Stage::Unwind(u) => match u {
                Unwind::FieldPath(path) => air::Stage::Unwind(air::Unwind {
                    source: Box::new(source.expect("$unwind without valid source stage")),
                    path: path.into(),
                    index: None,
                    outer: false,
                }),
                Unwind::Document(d) => air::Stage::Unwind(air::Unwind {
                    source: Box::new(source.expect("$unwind without valid source stage")),
                    path: (*d.path).into(),
                    index: d.include_array_index,
                    outer: d.preserve_null_and_empty_arrays.unwrap_or(false),
                }),
            },
            Stage::Lookup(l) => {
                let (from_db, from_coll) = match l.from {
                    Some(LookupFrom::Collection(c)) => (None, Some(c)),
                    Some(LookupFrom::Namespace(n)) => (Some(n.db), Some(n.coll)),
                    None => (None, None),
                };
                let let_vars = l.let_body.map(|let_body| {
                    let_body
                        .into_iter()
                        // sorted for testing purposes
                        .sorted_by(|a, b| Ord::cmp(&a.0, &b.0))
                        .map(|(k, v)| air::LetVariable {
                            name: k,
                            expr: Box::new(v.into()),
                        })
                        .collect::<Vec<air::LetVariable>>()
                });

                let source_collection = make_optional_collection_stage(from_db, from_coll);
                let pipeline = l
                    .pipeline
                    .into_iter()
                    .fold(source_collection, |acc, cur| {
                        Some(air::Stage::from((acc, cur)))
                    })
                    .unwrap();

                air::Stage::Lookup(air::Lookup {
                    source: Box::new(source.expect("$lookup without valid source stage")),
                    let_vars,
                    pipeline: Box::new(pipeline),
                    as_var: l.as_var,
                })
            }
            Stage::EquiLookup(eql) => {
                let (from_db, from_coll) = match eql.from {
                    LookupFrom::Collection(c) => (None, c),
                    LookupFrom::Namespace(n) => (Some(n.db), n.coll),
                };
                air::Stage::EquiLookup(air::EquiLookup {
                    source: Box::new(source.expect("$equilookup without valid source stage")),
                    from: air::Collection {
                        db: from_db.unwrap_or_else(|| "test".to_string()),
                        collection: from_coll,
                    },
                    local_field: eql.local_field.into(),
                    foreign_field: eql.foreign_field.into(),
                    as_var: eql.as_var,
                })
            }
        }
    }
}

impl From<ProjectItem> for air::ProjectItem {
    fn from(ast_project_item: ProjectItem) -> Self {
        match ast_project_item {
            ProjectItem::Exclusion => air::ProjectItem::Exclusion,
            ProjectItem::Inclusion => air::ProjectItem::Inclusion,
            ProjectItem::Assignment(ast_expr) => air::ProjectItem::Assignment(ast_expr.into()),
        }
    }
}

impl From<Expression> for air::Expression {
    fn from(ast_expr: Expression) -> Self {
        match ast_expr {
            Expression::Literal(lv) => air::Expression::Literal(air::LiteralValue::from(lv)),
            Expression::StringOrRef(v) => match v {
                StringOrRef::String(s) => air::Expression::Literal(air::LiteralValue::String(s)),
                StringOrRef::FieldRef(s) => air::Expression::FieldRef(s.into()),
                StringOrRef::Variable(s) => air::Expression::Variable(s.into()),
            },
            Expression::TaggedOperator(to) => to.into(),
            Expression::UntaggedOperator(uo) => uo.into(),
            Expression::Array(a) => {
                air::Expression::Array(a.into_iter().map(air::Expression::from).collect::<Vec<_>>())
            }
            Expression::Document(m) => {
                let lhm: LinkedHashMap<String, air::Expression> = m
                    .into_iter()
                    // sort alphabetically for testing purposes
                    .sorted_by(|a, b| Ord::cmp(&a.0, &b.0))
                    .map(|(k, v)| (k, air::Expression::from(v)))
                    .collect::<_>();

                air::Expression::Document(lhm.into())
            }
        }
    }
}

impl From<LiteralValue> for air::LiteralValue {
    fn from(ast_lv: LiteralValue) -> Self {
        match ast_lv {
            LiteralValue::Null => air::LiteralValue::Null,
            LiteralValue::Boolean(v) => air::LiteralValue::Boolean(v),
            LiteralValue::Integer(v) => air::LiteralValue::Integer(v),
            LiteralValue::Long(v) => air::LiteralValue::Long(v),
            LiteralValue::Double(v) => air::LiteralValue::Double(v),
        }
    }
}

impl From<Box<Expression>> for Box<air::Expression> {
    fn from(e: Box<Expression>) -> Self {
        Box::new((*e).into())
    }
}

fn translate_let_bindings(let_bindings: HashMap<String, Expression>) -> Vec<air::LetVariable> {
    let_bindings
        .into_iter()
        .map(|(k, e)| air::LetVariable {
            name: k,
            expr: Box::new(e.into()),
        })
        .collect()
}

fn make_optional_collection_stage(
    db: Option<String>,
    collection: Option<String>,
) -> Option<air::Stage> {
    let db = db.unwrap_or_else(|| "test".to_string());
    let collection = collection.unwrap_or_else(|| "default".to_string());
    Some(air::Stage::Collection(air::Collection { db, collection }))
}

impl From<Subquery> for air::Subquery {
    fn from(s: Subquery) -> Self {
        air::Subquery {
            let_bindings: match s.let_bindings {
                None => vec![],
                Some(let_bindings) => translate_let_bindings(let_bindings),
            },
            output_path: s.output_path.unwrap_or_default(),
            pipeline: translate_pipeline(
                make_optional_collection_stage(s.db, s.collection),
                s.pipeline,
            )
            .expect("$subquery with empty pipeline not supported")
            .into(),
        }
    }
}

impl From<SubqueryExists> for air::SubqueryExists {
    fn from(s: SubqueryExists) -> Self {
        air::SubqueryExists {
            let_bindings: match s.let_bindings {
                None => vec![],
                Some(let_bindings) => translate_let_bindings(let_bindings),
            },
            pipeline: translate_pipeline(
                make_optional_collection_stage(s.db, s.collection),
                s.pipeline,
            )
            .expect("$subquery with empty pipeline not supported")
            .into(),
        }
    }
}

impl From<TaggedOperator> for air::Expression {
    fn from(ast_op: TaggedOperator) -> Self {
        match ast_op {
            TaggedOperator::GetField(gf) => air::Expression::GetField(air::GetField {
                field: gf.field,
                input: gf.input.into(),
            }),
            TaggedOperator::SetField(sf) => air::Expression::SetField(air::SetField {
                field: sf.field,
                input: sf.input.into(),
                value: sf.value.into(),
            }),
            TaggedOperator::UnsetField(uf) => air::Expression::UnsetField(air::UnsetField {
                field: uf.field,
                input: uf.input.into(),
            }),
            TaggedOperator::Switch(s) => air::Expression::Switch(air::Switch {
                branches: s
                    .branches
                    .into_iter()
                    .map(|b| air::SwitchCase {
                        case: b.case.into(),
                        then: b.then.into(),
                    })
                    .collect(),
                default: s.default.into(),
            }),
            TaggedOperator::Let(l) => air::Expression::Let(air::Let {
                vars: l
                    .vars
                    .into_iter()
                    // sort alphabetically for testing purposes
                    .sorted_by(|a, b| Ord::cmp(&a.0, &b.0))
                    .map(|(k, v)| air::LetVariable {
                        name: k,
                        expr: Box::new(v.into()),
                    })
                    .collect(),
                inside: l.inside.into(),
            }),
            TaggedOperator::RegexMatch(r) => air::Expression::RegexMatch(air::RegexMatch {
                input: r.input.into(),
                regex: r.regex.into(),
                options: r.options.map(|o| o.into()),
            }),
            TaggedOperator::SqlConvert(c) => {
                let target_type = match c.to.as_str() {
                    "array" => air::SqlConvertTargetType::Array,
                    "object" => air::SqlConvertTargetType::Document,
                    _ => panic!("invalid '$sqlConvert' target type '{}'", c.to),
                };

                air::Expression::SqlConvert(air::SqlConvert {
                    input: c.input.into(),
                    to: target_type,
                    on_null: c.on_null.into(),
                    on_error: c.on_error.into(),
                })
            }
            TaggedOperator::Convert(c) => air::Expression::Convert(air::Convert {
                input: c.input.into(),
                to: str_to_air_type(c.to),
                on_null: c.on_null.into(),
                on_error: c.on_error.into(),
            }),
            TaggedOperator::Like(l) => air::Expression::Like(air::Like {
                expr: l.input.into(),
                pattern: l.pattern.into(),
                escape: l.escape,
            }),
            TaggedOperator::SqlDivide(d) => air::Expression::SqlDivide(air::SqlDivide {
                dividend: d.dividend.into(),
                divisor: d.divisor.into(),
                on_error: d.on_error.into(),
            }),
            TaggedOperator::Subquery(s) => air::Expression::Subquery(s.into()),
            TaggedOperator::SubqueryComparison(sc) => {
                let (op, op_type) = str_to_air_sqcop(&sc.op)
                    .unwrap_or_else(|| panic!("found bad subquery comparison op: {}", sc.op));
                air::Expression::SubqueryComparison(air::SubqueryComparison {
                    op,
                    op_type,
                    modifier: str_to_air_sqcmod(&sc.modifier)
                        .unwrap_or_else(|| panic!("found bad subquery modifier: {}", sc.modifier)),
                    arg: sc.arg.into(),
                    subquery: Box::new((*sc.subquery).into()),
                })
            }
            TaggedOperator::SubqueryExists(se) => air::Expression::SubqueryExists(se.into()),
            TaggedOperator::Trim(t) => air::Expression::Trim(air::Trim {
                op: air::TrimOperator::Trim,
                input: t.input.into(),
                chars: t.chars.into(),
            }),
            TaggedOperator::LTrim(lt) => air::Expression::Trim(air::Trim {
                op: air::TrimOperator::LTrim,
                input: lt.input.into(),
                chars: lt.chars.into(),
            }),
            TaggedOperator::RTrim(rt) => air::Expression::Trim(air::Trim {
                op: air::TrimOperator::RTrim,
                input: rt.input.into(),
                chars: rt.chars.into(),
            }),
            TaggedOperator::Reduce(r) => air::Expression::Reduce(air::Reduce {
                input: r.input.into(),
                init_value: r.initial_value.into(),
                inside: r.inside.into(),
            }),
        }
    }
}

fn str_to_air_sqcop(
    op: &str,
) -> Option<(air::SubqueryComparisonOp, air::SubqueryComparisonOpType)> {
    let op_type = if op.starts_with("sql") {
        air::SubqueryComparisonOpType::Sql
    } else {
        air::SubqueryComparisonOpType::Mql
    };
    let op = op.trim_start_matches("sql");
    match op {
        "lt" | "Lt" => Some((air::SubqueryComparisonOp::Lt, op_type)),
        "lte" | "Lte" => Some((air::SubqueryComparisonOp::Lte, op_type)),
        // might as well support ne and neq both
        "ne" | "neq" | "Ne" | "Neq" => Some((air::SubqueryComparisonOp::Neq, op_type)),
        "eq" | "Eq" => Some((air::SubqueryComparisonOp::Eq, op_type)),
        "gt" | "Gt" => Some((air::SubqueryComparisonOp::Gt, op_type)),
        "gte" | "Gte" => Some((air::SubqueryComparisonOp::Gte, op_type)),
        _ => None,
    }
}

fn str_to_air_sqcmod(m: &str) -> Option<air::SubqueryModifier> {
    match m {
        "all" => Some(air::SubqueryModifier::All),
        "any" => Some(air::SubqueryModifier::Any),
        _ => None,
    }
}

fn str_to_air_type(t: String) -> air::Type {
    match t.as_str() {
        "array" => air::Type::Array,
        "binData" => air::Type::BinData,
        "bool" => air::Type::Boolean,
        "date" => air::Type::Datetime,
        "dbPointer" => air::Type::DbPointer,
        "decimal" => air::Type::Decimal128,
        "object" => air::Type::Document,
        "double" => air::Type::Double,
        "int" => air::Type::Int32,
        "long" => air::Type::Int64,
        "javascript" => air::Type::Javascript,
        "javascriptWithScope" => air::Type::JavascriptWithScope,
        "maxKey" => air::Type::MaxKey,
        "minKey" => air::Type::MinKey,
        "null" => air::Type::Null,
        "objectId" => air::Type::ObjectId,
        "regex" => air::Type::RegularExpression,
        "string" => air::Type::String,
        "symbol" => air::Type::Symbol,
        "timestamp" => air::Type::Timestamp,
        "undefined" => air::Type::Undefined,
        _ => panic!("invalid $convert or $is target type '{t}'"),
    }
}

impl From<UntaggedOperator> for air::Expression {
    fn from(ast_op: UntaggedOperator) -> Self {
        let args: Vec<air::Expression> =
            ast_op.args.into_iter().map(air::Expression::from).collect();

        // Special cases:
        //   - $literal becomes a Literal
        //   - $is/$sqlIs become an Is
        //   - $nullIf and $coalesce are SQL operators but don't start with $sql
        match ast_op.op.as_str() {
            "$literal" => {
                let arg = args.first().unwrap();
                match arg {
                    air::Expression::Literal(_) => return arg.clone(),
                    _ => panic!("invalid $literal"),
                }
            }
            "$is" | "$sqlIs" => {
                let expr = Box::new(args.first().unwrap().clone());

                let arg2 = match args.get(1).unwrap() {
                    air::Expression::Literal(air::LiteralValue::String(s)) => s,
                    _ => panic!("invalid second arg to $is/$sqlIs"),
                };
                let target_type = to_type_or_missing(arg2.clone());

                return air::Expression::Is(air::Is { expr, target_type });
            }
            "$nullIf" => {
                return air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                    op: air::SQLOperator::NullIf,
                    args,
                })
            }
            "$coalesce" => {
                return air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                    op: air::SQLOperator::Coalesce,
                    args,
                })
            }
            "$numberDouble" => {
                if let air::Expression::Literal(air::LiteralValue::String(s)) = args[0].clone() {
                    if s == "Infinity" {
                        return air::Expression::Literal(air::LiteralValue::Double(f64::INFINITY));
                    } else if s == "-Infinity" {
                        return air::Expression::Literal(air::LiteralValue::Double(
                            f64::NEG_INFINITY,
                        ));
                    } else if s == "NaN" {
                        return air::Expression::Literal(air::LiteralValue::Double(f64::NAN));
                    }
                }
            }
            _ => (),
        }

        if ast_op.op.starts_with("$sql") {
            let op = match ast_op.op.as_str() {
                "$sqlPos" => air::SQLOperator::Pos,
                "$sqlNeg" => air::SQLOperator::Neg,
                "$sqlLt" => air::SQLOperator::Lt,
                "$sqlLte" => air::SQLOperator::Lte,
                "$sqlNe" => air::SQLOperator::Ne,
                "$sqlEq" => air::SQLOperator::Eq,
                "$sqlGt" => air::SQLOperator::Gt,
                "$sqlGte" => air::SQLOperator::Gte,
                "$sqlBetween" => air::SQLOperator::Between,
                "$sqlNot" => air::SQLOperator::Not,
                "$sqlAnd" => air::SQLOperator::And,
                "$sqlOr" => air::SQLOperator::Or,
                "$sqlSlice" => air::SQLOperator::Slice,
                "$sqlSize" => air::SQLOperator::Size,
                "$sqlIndexOfCP" => air::SQLOperator::IndexOfCP,
                "$sqlStrLenCP" => air::SQLOperator::StrLenCP,
                "$sqlStrLenBytes" => air::SQLOperator::StrLenBytes,
                "$sqlBitLength" => air::SQLOperator::BitLength,
                "$sqlCos" => air::SQLOperator::Cos,
                "$sqlLog" => air::SQLOperator::Log,
                "$sqlMod" => air::SQLOperator::Mod,
                "$sqlRound" => air::SQLOperator::Round,
                "$sqlSin" => air::SQLOperator::Sin,
                "$sqlSqrt" => air::SQLOperator::Sqrt,
                "$sqlTan" => air::SQLOperator::Tan,
                "$sqlSubstrCP" => air::SQLOperator::SubstrCP,
                "$sqlToUpper" => air::SQLOperator::ToUpper,
                "$sqlToLower" => air::SQLOperator::ToLower,
                "$sqlSplit" => air::SQLOperator::Split,
                _ => panic!("invalid sql operator '{}'", ast_op.op),
            };

            air::Expression::SQLSemanticOperator(air::SQLSemanticOperator { op, args })
        } else {
            let op = match ast_op.op.as_str() {
                "$concat" => air::MQLOperator::Concat,
                "$cond" => air::MQLOperator::Cond,
                "$ifNull" => air::MQLOperator::IfNull,
                "$add" => air::MQLOperator::Add,
                "$subtract" => air::MQLOperator::Subtract,
                "$multiply" => air::MQLOperator::Multiply,
                "$divide" => air::MQLOperator::Divide,
                "$lt" => air::MQLOperator::Lt,
                "$lte" => air::MQLOperator::Lte,
                "$ne" => air::MQLOperator::Ne,
                "$eq" => air::MQLOperator::Eq,
                "$gt" => air::MQLOperator::Gt,
                "$gte" => air::MQLOperator::Gte,
                "$mqlBetween" => air::MQLOperator::Between,
                "$not" => air::MQLOperator::Not,
                "$and" => air::MQLOperator::And,
                "$or" => air::MQLOperator::Or,
                "$slice" => air::MQLOperator::Slice,
                "$size" => air::MQLOperator::Size,
                "$arrayElemAt" => air::MQLOperator::ElemAt,
                "$in" => air::MQLOperator::In,
                "$first" => air::MQLOperator::First,
                "$last" => air::MQLOperator::Last,
                "$indexOfCP" => air::MQLOperator::IndexOfCP,
                "$indexOfBytes" => air::MQLOperator::IndexOfBytes,
                "$strLenCP" => air::MQLOperator::StrLenCP,
                "$strLenBytes" => air::MQLOperator::StrLenBytes,
                "$abs" => air::MQLOperator::Abs,
                "$ceil" => air::MQLOperator::Ceil,
                "$cos" => air::MQLOperator::Cos,
                "$degreesToRadians" => air::MQLOperator::DegreesToRadians,
                "$floor" => air::MQLOperator::Floor,
                "$log" => air::MQLOperator::Log,
                "$mod" => air::MQLOperator::Mod,
                "$pow" => air::MQLOperator::Pow,
                "$radiansToDegrees" => air::MQLOperator::RadiansToDegrees,
                "$round" => air::MQLOperator::Round,
                "$sin" => air::MQLOperator::Sin,
                "$tan" => air::MQLOperator::Tan,
                "$sqrt" => air::MQLOperator::Sqrt,
                "$avg" => air::MQLOperator::Avg,
                "$max" => air::MQLOperator::Max,
                "$min" => air::MQLOperator::Min,
                "$sum" => air::MQLOperator::Sum,
                "$stdDevPop" => air::MQLOperator::StddevPop,
                "$stdDevSamp" => air::MQLOperator::StddevSamp,
                "$substrCP" => air::MQLOperator::SubstrCP,
                "$substrBytes" => air::MQLOperator::SubstrBytes,
                "$toUpper" => air::MQLOperator::ToUpper,
                "$toLower" => air::MQLOperator::ToLower,
                "$split" => air::MQLOperator::Split,
                "$year" => air::MQLOperator::Year,
                "$month" => air::MQLOperator::Month,
                "$dayOfMonth" => air::MQLOperator::DayOfMonth,
                "$hour" => air::MQLOperator::Hour,
                "$minute" => air::MQLOperator::Minute,
                "$second" => air::MQLOperator::Second,
                "$week" => air::MQLOperator::Week,
                "$dayOfYear" => air::MQLOperator::DayOfYear,
                "$isoWeek" => air::MQLOperator::IsoWeek,
                "$isoDayOfWeek" => air::MQLOperator::IsoDayOfWeek,
                "$dateAdd" => air::MQLOperator::DateAdd,
                "$dateDiff" => air::MQLOperator::DateDiff,
                "$dateTrunc" => air::MQLOperator::DateTrunc,
                "$mergeObjects" => air::MQLOperator::MergeObjects,
                "$type" => air::MQLOperator::Type,
                "$isArray" => air::MQLOperator::IsArray,
                "$isNumber" => air::MQLOperator::IsNumber,
                _ => panic!("invalid mql operator '{}'", ast_op.op),
            };

            air::Expression::MQLSemanticOperator(air::MQLSemanticOperator { op, args })
        }
    }
}

fn to_type_or_missing(s: String) -> air::TypeOrMissing {
    match s.as_str() {
        "missing" => air::TypeOrMissing::Missing,
        "number" => air::TypeOrMissing::Number,
        _ => air::TypeOrMissing::Type(str_to_air_type(s)),
    }
}

impl From<String> for air::AggregationFunction {
    fn from(s: String) -> Self {
        match s.as_str() {
            "$addToSet" => air::AggregationFunction::AddToSet,
            "$push" => air::AggregationFunction::AddToArray,
            "$avg" | "$sqlAvg" => air::AggregationFunction::Avg,
            "$sqlCount" => air::AggregationFunction::Count,
            "$first" | "$sqlFirst" => air::AggregationFunction::First,
            "$last" | "$sqlLast" => air::AggregationFunction::Last,
            "$max" | "$sqlMax" => air::AggregationFunction::Max,
            "$mergeObjects" | "$sqlMergeObjects" => air::AggregationFunction::MergeDocuments,
            "$min" | "$sqlMin" => air::AggregationFunction::Min,
            "$stdDevPop" | "$sqlStdDevPop" => air::AggregationFunction::StddevPop,
            "$stdDevSamp" | "$sqlStdDevSamp" => air::AggregationFunction::StddevSamp,
            "$sum" | "$sqlSum" => air::AggregationFunction::Sum,
            _ => panic!("Recieved invalid Group aggregation function: {s}"),
        }
    }
}
