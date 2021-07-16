use crate::{
    codegen::{Error, Result},
    ir::{
        self,
        binding_tuple::{BindingTuple, DatasourceName, Key},
        ScalarFunction,
    },
    map,
};
use std::collections::BTreeMap;

#[derive(PartialEq, Debug, Clone)]
pub struct MappingRegistry(BTreeMap<Key, String>);

impl MappingRegistry {
    pub fn new() -> Self {
        MappingRegistry(BTreeMap::new())
    }

    pub fn get(&self, k: &Key) -> Option<&String> {
        self.0.get(k)
    }

    pub fn insert<K: Into<Key>, V: Into<String>>(&mut self, k: K, v: V) -> Option<String> {
        self.0.insert(k.into(), v.into())
    }

    pub fn merge(&mut self, other: MappingRegistry) -> &mut Self {
        self.0.extend(other.0.into_iter());
        self
    }
}

impl Default for MappingRegistry {
    fn default() -> Self {
        Self::new()
    }
}

#[derive(PartialEq, Debug)]
pub struct MqlTranslation {
    pub database: Option<String>,
    pub collection: Option<String>,
    pub mapping_registry: MappingRegistry,
    pub pipeline: Vec<bson::Document>,
}

impl MqlTranslation {
    fn with_additional_stage(mut self, stage: bson::Document) -> Self {
        self.pipeline.push(stage);
        self
    }

    fn with_mapping_registry(self, mapping_registry: MappingRegistry) -> Self {
        MqlTranslation {
            mapping_registry,
            ..self
        }
    }
}

#[derive(Clone)]
pub struct MqlCodeGenerator {
    pub mapping_registry: MappingRegistry,
}

impl ScalarFunction {
    pub fn mql_op(self) -> Option<&'static str> {
        use ir::ScalarFunction::*;

        Some(match self {
            Concat => "$concat",

            // Arithmetic operators
            Add => "$add",
            Sub => "$subtract",
            Mul => "$multiply",
            Div => "$sqlDivide",

            // Comparison operators
            Lt => "$sqlLt",
            Lte => "$sqlLte",
            Neq => "$sqlNe",
            Eq => "$sqlEq",
            Gt => "$sqlGt",
            Gte => "$sqlGte",
            Between => "$sqlBetween",

            // Boolean operators
            Not => "$sqlNot",
            And => "$sqlAnd",
            Or => "$sqlOr",

            // Conditional scalar functions
            NullIf => "$nullIf",
            Coalesce => "$coalesce",

            // Array scalar functions
            Slice => "$sqlSlice",
            Size => "$sqlSize",

            // Numeric value scalar functions
            Position => "$sqlIndexOfCP",
            CharLength => "$sqlStrLenCP",
            OctetLength => "$sqlStrLenBytes",
            BitLength => "$sqlStrLenBytes", // with $mul

            // String value scalar functions
            Substring => "$sqlSubstrCP",
            Upper => "$sqlToUpper",
            Lower => "$sqlToLower",
            Pos => return None,
            Neg => return None,
            ComputedFieldAccess => return None,
            BTrim => "$trim",
            LTrim => "$ltrim",
            RTrim => "$rtrim",
            CurrentTimestamp => return None,
            Year => "$year",
            Month => "$month",
            Day => "$day",
            Hour => "$hour",
            Minute => "$minute",
            Second => "$second",
        })
    }
}

impl MqlCodeGenerator {
    fn get_unique_bot_name(project_names: &BindingTuple<ir::Expression>) -> String {
        let mut ret = "__bot".to_string();
        if project_names.is_empty() {
            return ret;
        }
        // find the current scope, if it is not the same in all keys in this
        // Project expression, it is because of an Algebrization error.
        let current_scope = project_names.keys().next().unwrap().scope;
        while project_names.contains_key(&(ret.clone(), current_scope).into()) {
            ret.insert(0, '_');
        }
        ret
    }

    fn map_project_datasource(datasource: &DatasourceName, unique_bot_name: &str) -> String {
        match datasource {
            DatasourceName::Bottom => unique_bot_name.to_string(),
            DatasourceName::Named(s) => s.clone(),
        }
    }

    /// Recursively generates a translation for this stage and its
    /// sources. When this function is called, `self.mapping_registry`
    /// should include mappings for any datasources from outer scopes.
    /// Mappings for the current scope will be obtained by calling
    /// `codegen_stage` on source stages.
    pub fn codegen_stage(&self, stage: ir::Stage) -> Result<MqlTranslation> {
        use bson::{doc, Bson};
        use ir::Stage::*;
        match stage {
            Filter(f) => Ok(self.codegen_stage(*f.source)?.with_additional_stage(
                doc! {"$match": {"$expr": self.codegen_expression(f.condition)?}},
            )),
            Project(p) => {
                let source_translation = self.codegen_stage(*p.source)?;
                // expression_generator contains all the possible correlated mappings
                // from the self MqlCodeGenerator merged with the mappings from
                // the source Stage.
                let expression_generator = self
                    .clone()
                    .with_merged_mappings(source_translation.mapping_registry.clone());
                // output_registry will be the registry we output from this
                // stage. It should only contain mappings for the names
                // defined by the $project, because $project kills all
                // other values, anyway.
                let mut output_registry = MappingRegistry::new();
                let unique_bot_name = MqlCodeGenerator::get_unique_bot_name(&p.expression);
                // we need to porject away _id unless the query maps _id some other way.
                // {_id: 0} will be overwritten if _id is defined in the project expression.
                let mut project_body = doc! {"_id": 0};
                for (k, e) in p.expression.into_iter() {
                    let mapped_k =
                        MqlCodeGenerator::map_project_datasource(&k.datasource, &unique_bot_name);
                    project_body.insert(
                        mapped_k.clone(),
                        expression_generator.codegen_expression(e)?,
                    );
                    output_registry.insert(k, mapped_k);
                }
                Ok(source_translation
                    .with_mapping_registry(output_registry)
                    .with_additional_stage(doc! {"$project": project_body}))
            }
            Group(_) => unimplemented!(),
            Limit(l) => Ok(self
                .codegen_stage(*l.source)?
                .with_additional_stage(doc! {"$limit": l.limit})),
            Offset(o) => Ok(self
                .codegen_stage(*o.source)?
                .with_additional_stage(doc! {"$skip": o.offset})),
            Sort(s) => self.codegen_sort(s),
            Collection(c) => Ok(MqlTranslation {
                database: Some(c.db),
                collection: Some(c.collection.clone()),
                mapping_registry: MappingRegistry(
                    map! {(&c.collection, 0u16).into() => c.collection.clone()},
                ),
                pipeline: vec![doc! {"$project": {"_id": 0, &c.collection: "$$ROOT"}}],
            }),
            Array(arr) => {
                let mapping_registry =
                    MappingRegistry(map! {(&arr.alias, 0u16).into() => arr.alias.clone()});
                let docs = arr
                    .array
                    .into_iter()
                    .map(|e| self.codegen_expression(e))
                    .collect::<Result<Vec<Bson>>>()?;
                Ok(MqlTranslation {
                    database: None,
                    collection: None,
                    mapping_registry,
                    pipeline: vec![doc! {"$array": {arr.alias: Bson::Array(docs)}}],
                })
            }
            Join(join) => {
                use serde::{Deserialize, Serialize};
                #[derive(Serialize, Deserialize)]
                struct JoinBody {
                    #[serde(skip_serializing_if = "Option::is_none")]
                    database: Option<String>,
                    #[serde(skip_serializing_if = "Option::is_none")]
                    collection: Option<String>,
                    #[serde(rename = "joinType")]
                    join_type: &'static str,
                    #[serde(skip_serializing_if = "Option::is_none", rename = "let")]
                    let_body: Option<bson::Document>,
                    pipeline: Vec<bson::Document>,
                    #[serde(skip_serializing_if = "Option::is_none")]
                    condition: Option<bson::Document>,
                }
                let left = self.codegen_stage(*join.left)?;
                let right = self.codegen_stage(*join.right)?;
                let database = match (left.database.clone(), right.database) {
                    (Some(ref left_db), Some(ref right_db)) => {
                        if left_db != right_db {
                            Some(right_db.clone())
                        } else {
                            None
                        }
                    }
                    (None, Some(ref right_db)) => Some(right_db.clone()),
                    (_, None) => None,
                };
                let collection = right.collection;
                let join_type = match join.join_type {
                    ir::JoinType::Inner => "inner",
                    ir::JoinType::Left => "left",
                };
                let output_registry = self
                    .mapping_registry
                    .clone()
                    .merge(left.mapping_registry.clone())
                    .merge(right.mapping_registry.clone())
                    .clone();
                let pipeline = right.pipeline;
                let (condition, let_body) = match join.condition {
                    None => (None, None),
                    Some(expr) => {
                        use bson::bson;
                        let left_registry = left
                            .mapping_registry
                            .clone()
                            .0
                            .into_iter()
                            .map(|(k, v)| (k.clone(), format!("$__{}_{}", v, k.scope)))
                            .collect();
                        let let_doc: bson::Document = left
                            .mapping_registry
                            .clone()
                            .0
                            .into_iter()
                            .map(|(k, v)| {
                                (format!("__{}_{}", v, k.scope), bson! {format!("${}", v)})
                            })
                            .collect();
                        let expression_generator = self
                            .clone()
                            .with_merged_mappings(MappingRegistry(left_registry))
                            .with_merged_mappings(right.mapping_registry);
                        let cond = expression_generator.codegen_expression(expr)?;
                        let cond_doc = doc! {"$match" : {"$expr": cond}};
                        (Some(cond_doc), Some(let_doc))
                    }
                };
                let join_body = bson::to_bson(&JoinBody {
                    database,
                    collection,
                    join_type,
                    let_body,
                    pipeline,
                    condition,
                })
                .unwrap();
                Ok(left
                    .with_mapping_registry(output_registry)
                    .with_additional_stage(doc! {"$join": join_body}))
            }
            Set(_) => unimplemented!(),
        }
    }

    fn with_merged_mappings(mut self, mappings: MappingRegistry) -> Self {
        self.mapping_registry.merge(mappings);
        self
    }

    fn codegen_sort(&self, sort: ir::Sort) -> Result<MqlTranslation> {
        use bson::{doc, Bson};
        use ir::{Expression::*, SortSpecification::*};

        let source_translation = self.codegen_stage(*sort.source)?;
        let expression_generator = self
            .clone()
            .with_merged_mappings(source_translation.mapping_registry.clone());

        let sort_specs = sort
            .specs
            .into_iter()
            .map(|spec| {
                let (expr, direction) = match spec {
                    Asc(expr) => (*expr, Bson::Int32(1)),
                    Dsc(expr) => (*expr, Bson::Int32(-1)),
                };

                // anything that's not a reference or a static field
                // access cannot be used as a sort key
                match expr {
                    Reference(_) | FieldAccess(_) => Ok(()),
                    _ => Err(Error::InvalidSortKey),
                }?;

                // we still need to ensure that the result is a
                // string, since not all FieldAccess expressions will
                // translate to single MQL references
                let expr = expression_generator.codegen_expression(expr)?;
                let key = match expr {
                    Bson::String(s) => Ok(s[1..].to_string()),
                    _ => Err(Error::InvalidSortKey),
                }?;

                Ok((key, direction))
            })
            .collect::<Result<bson::Document>>()?;

        Ok(source_translation.with_additional_stage(doc! {"$sort": sort_specs}))
    }

    /// Recursively generates a translation for this expression. When
    /// this function is called, `self.mapping_registry` should
    /// include mappings for all datasources in scope.
    #[allow(dead_code)]
    pub fn codegen_expression(&self, expr: ir::Expression) -> Result<bson::Bson> {
        use bson::Bson;
        use ir::{Expression::*, Literal::*};
        match expr {
            Literal(lit) => Ok(bson::bson!({
                "$literal": match lit {
                    Null => Bson::Null,
                    Boolean(b) => Bson::Boolean(b),
                    String(s) => Bson::String(s),
                    Integer(i) => Bson::Int32(i),
                    Long(l) => Bson::Int64(l),
                    Double(d) => Bson::Double(d),
                },
            })),
            Reference(key) => self
                .mapping_registry
                .0
                .get(&key)
                .ok_or(Error::ReferenceNotFound(key))
                .map(|s| Bson::String(format!("${}", s))),
            Array(exprs) => Ok(Bson::Array(
                exprs
                    .into_iter()
                    .map(|e| self.codegen_expression(e))
                    .collect::<Result<Vec<Bson>>>()?,
            )),
            Document(map) => Ok(Bson::Document({
                if map.is_empty() {
                    bson::doc! {"$literal": {}}
                } else {
                    map.into_iter()
                        .map(|(k, v)| {
                            if k.starts_with('$') {
                                Err(Error::DotsOrDollarsInFieldName)
                            } else {
                                Ok((k, self.codegen_expression(v)?))
                            }
                        })
                        .collect::<Result<bson::Document>>()?
                }
            })),
            FieldAccess(fa) => {
                if fa.field.contains(&['.', '$'] as &[_]) {
                    return Err(Error::DotsOrDollarsInFieldName);
                };
                Ok(match self.codegen_expression(*fa.expr)? {
                    Bson::String(e) => Bson::String(format!("{}.{}", e, fa.field)),
                    e => bson::bson!({"$getField": {"field": fa.field, "input": e}}),
                })
            }
            SearchedCase(ce) => {
                let br = ce
                    .when_branch
                    .into_iter()
                    .map(|wb| {
                        Ok(bson::doc! {"case": self.codegen_expression(*wb.when)?,
                        "then": self.codegen_expression(*wb.then)?})
                    })
                    .collect::<Result<Vec<bson::Document>>>()?;

                Ok(bson::bson!({"$switch": {
                    "branches": br,
                    "default": self.codegen_expression(*ce.else_branch)?
                    }
                }))
            }
            SimpleCase(ce) => {
                let br = ce
                    .when_branch
                    .into_iter()
                    .map(|wb| {
                        Ok(bson::doc! {"case": {"$sqlEq": ["$$target", self.codegen_expression(*wb.when)?]},
                        "then": self.codegen_expression(*wb.then)?})
                    })
                    .collect::<Result<Vec<bson::Document>>>()?;

                Ok(
                    bson::bson!({"$let": {"vars": {"target": self.codegen_expression(*ce.expr)?}},
                        "in": {"$switch": {
                        "branches": br,
                        "default": self.codegen_expression(*ce.else_branch)?
                        }
                    }}),
                )
            }
            Cast(_) => unimplemented!(),
            TypeAssertion(_) => unimplemented!(),
            ScalarFunction(sa) => {
                use crate::ir::ScalarFunction::*;
                Ok(match sa.function {
                    Pos => self.codegen_expression(sa.args[0].clone())?,
                    Neg => bson::bson!({
                        "$multiply" : [
                             self.codegen_expression(sa.args[0].clone())?,
                             Bson::Int32(-1),
                        ]
                    }),
                    ComputedFieldAccess => {
                        return Err(Error::UnsupportedFunction(ComputedFieldAccess))
                    } // depend on new $getField operator in SQL-282
                    CurrentTimestamp => Bson::String("$$NOW".to_string()),
                    CharLength | OctetLength | Size | Upper | Lower | Year | Month | Day | Hour
                    | Minute | Second => {
                        bson::bson!({ sa.function.mql_op().unwrap(): self.codegen_expression(sa.args[0].clone())?})
                    }
                    BitLength => bson::bson!({
                        "$multiply" : [
                            { sa.function.mql_op().unwrap(): self.codegen_expression(sa.args[0].clone())?},
                             Bson::Int32(8),
                        ]
                    }),
                    BTrim | LTrim | RTrim => bson::bson!({
                        sa.function.mql_op().unwrap(): {"input": self.codegen_expression(sa.args[1].clone())?,
                            "chars": self.codegen_expression(sa.args[0].clone())?,
                    }}),
                    Not | Concat | Add | Sub | Mul | Div | Lt | Lte | Neq | Eq | Gt | Gte
                    | Between | And | Or | NullIf | Coalesce | Slice | Position | Substring => {
                        let args = Bson::Array(
                            sa.args
                                .into_iter()
                                .map(|e| self.codegen_expression(e))
                                .collect::<Result<Vec<Bson>>>()?,
                        );
                        Bson::Document(bson::doc! { sa.function.mql_op().unwrap(): args})
                    }
                })
            }
            Like(expr) => Ok(match expr.escape {
                Some(escape) => Bson::Document(bson::doc! {
                "$like": {"input": self.codegen_expression(*expr.expr)?,
                          "pattern": self.codegen_expression(*expr.pattern)?,
                          "escape": escape}}),
                None => Bson::Document(bson::doc! {
                "$like": {"input": self.codegen_expression(*expr.expr)?,
                          "pattern": self.codegen_expression(*expr.pattern)?}}),
            }),
            SubqueryExpression(_) => unimplemented!(),
            SubqueryComparison(_) => unimplemented!(),
            Exists(_) => unimplemented!(),
            Is(expr) => {
                use crate::ir::TypeOrMissing;
                Ok(match expr.target_type {
                    TypeOrMissing::Number => Bson::Document(
                        bson::doc! {"$isNumber": self.codegen_expression(*expr.expr)?},
                    ),
                    _ => unimplemented!(),
                })
            }
        }
    }
}
