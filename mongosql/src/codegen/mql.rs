use crate::{
    codegen::{Error, Result},
    ir::{
        self,
        binding_tuple::{BindingTuple, DatasourceName, Key},
        AggregationExpr, ScalarFunction, Stage, Type, TypeOrMissing,
    },
    map,
};
use bson::Bson;
use std::{
    collections::{BTreeMap, BTreeSet},
    convert::From,
};

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

    /// generate_path_components takes an expression and returns a vector of
    /// its components by recursively tracing its path.
    pub fn generate_path_components(&self, expr: ir::Expression) -> Result<Vec<String>> {
        match expr {
            ir::Expression::Reference(key) => match self.mapping_registry.get(&key) {
                Some(name) => Ok(vec![name.clone()]),
                None => Err(Error::ReferenceNotFound(key)),
            },
            ir::Expression::FieldAccess(fa) => {
                let mut path = self.generate_path_components(*fa.expr)?;
                path.push(fa.field.clone());
                Ok(path)
            }
            _ => Err(Error::NoFieldPathForExpr),
        }
    }
}

#[derive(Clone)]
pub struct MqlCodeGenerator {
    pub mapping_registry: MappingRegistry,
    pub scope_level: u16,
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

            // MergeObject
            MergeObjects => "$mergeObjects",
        })
    }
}

impl TypeOrMissing {
    pub fn mql_type(self) -> Option<&'static str> {
        use ir::TypeOrMissing::*;
        Some(match self {
            Missing => "missing",
            Number => return None,
            Type(t) => t.mql_type(),
        })
    }
}

impl Type {
    pub fn mql_type(self) -> &'static str {
        use ir::Type::*;
        match self {
            Array => "array",
            BinData => "binData",
            Boolean => "bool",
            Datetime => "date",
            DbPointer => "dbPointer",
            Decimal128 => "decimal",
            Document => "object",
            Double => "double",
            Int32 => "int",
            Int64 => "long",
            Javascript => "javascript",
            JavascriptWithScope => "javascriptWithScope",
            MaxKey => "maxKey",
            MinKey => "minKey",
            Null => "null",
            ObjectId => "objectId",
            RegularExpression => "regex",
            String => "string",
            Symbol => "symbol",
            Timestamp => "timestamp",
            Undefined => "undefined",
        }
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

    fn get_datasource_name(datasource: &DatasourceName, unique_bot_name: &str) -> String {
        match datasource {
            DatasourceName::Bottom => unique_bot_name.to_string(),
            DatasourceName::Named(s) => s.clone(),
        }
    }

    fn get_unique_alias(existing_aliases: BTreeSet<String>, mut alias: String) -> String {
        while existing_aliases.contains(&alias) {
            alias.insert(0, '_')
        }
        alias
    }

    /// generate_let_bindings binds each datasource in the correlated mapping registry
    /// to a generated name of the form <datasource name>_<nesting depth>. It returns
    /// the resulting $let document along with a new mapping registry that contains the
    /// generated names. Naming conflicts are resolved by prepending underscores until the
    /// generated name is unique.
    fn generate_let_bindings(self) -> Result<(bson::Document, MappingRegistry)> {
        let mut let_bindings: bson::Document = map![];
        let new_mapping_registry = MappingRegistry(
            self.mapping_registry
                .0
                .into_iter()
                .map(|(key, value)| {
                    let mut generated_name = format!(
                        "{}_{}",
                        MqlCodeGenerator::get_datasource_name(&key.datasource, "__bot"),
                        key.scope
                    );
                    while let_bindings.contains_key(&generated_name) {
                        generated_name.insert(0, '_');
                    }
                    let_bindings.insert(generated_name.clone(), format!("${}", value));
                    generated_name.insert(0, '$');
                    (key, generated_name)
                })
                .collect::<BTreeMap<Key, String>>(),
        );
        Ok((let_bindings, new_mapping_registry))
    }

    /// Recursively generates a translation for this stage and its
    /// sources. When this function is called, `self.mapping_registry`
    /// should include mappings for any datasources from outer scopes.
    /// Mappings for the current scope will be obtained by calling
    /// `codegen_stage` on source stages.
    pub fn codegen_stage(&self, stage: ir::Stage) -> Result<MqlTranslation> {
        use bson::doc;
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
                // we need to project away _id unless the query maps _id some other way.
                // {_id: 0} will be overwritten if _id is defined in the project expression.
                let mut project_body = doc! {"_id": 0};
                for (k, e) in p.expression.into_iter() {
                    let mapped_k =
                        MqlCodeGenerator::get_datasource_name(&k.datasource, &unique_bot_name);
                    if mapped_k.starts_with('$') || mapped_k.contains('.') {
                        return Err(Error::DotsOrDollarsInProjectField);
                    }

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
            Group(g) => self.codegen_group_by(g),
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
                    map! {(&c.collection, self.scope_level).into() => c.collection.clone()},
                ),
                // It is not technically possible to have a collection named _id in a mongod
                // instance, but it may be possible in datalake or other future instantiations
                // of the mongo protocol, so we make sure that a collection with name _id is
                // supported.
                pipeline: vec![match c.collection.as_str() {
                    "_id" => doc! {"$project": {"_id": "$$ROOT"}},
                    _ => doc! {"$project": {"_id": 0, &c.collection: "$$ROOT"}},
                }],
            }),
            Array(arr) => {
                let mapping_registry = MappingRegistry(
                    map! {(&arr.alias, self.scope_level).into() => arr.alias.clone()},
                );
                let docs = arr
                    .array
                    .into_iter()
                    .map(|e| self.codegen_expression(e))
                    .collect::<Result<Vec<Bson>>>()?;
                Ok(MqlTranslation {
                    database: None,
                    collection: None,
                    mapping_registry,
                    pipeline: vec![
                        doc! {"$documents": Bson::Array(docs)},
                        match arr.alias.as_ref() {
                            "_id" => doc! {"$project": {"_id": "$$ROOT"}},
                            x => doc! {"$project": {"_id": 0, x: "$$ROOT"}},
                        },
                    ],
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
                let mut output_registry = self.mapping_registry.clone();
                output_registry
                    .merge(left.mapping_registry.clone())
                    .merge(right.mapping_registry.clone());
                let pipeline = right.pipeline;
                let (condition, let_body) = match join.condition {
                    None => (None, None),
                    Some(expr) => {
                        let join_generator = MqlCodeGenerator {
                            mapping_registry: left.mapping_registry.clone(),
                            scope_level: self.scope_level,
                        };
                        let (let_bindings, left_registry) =
                            join_generator.generate_let_bindings()?;
                        let expression_generator = self
                            .clone()
                            .with_merged_mappings(MappingRegistry(left_registry.0))
                            .with_merged_mappings(right.mapping_registry);
                        let cond = expression_generator.codegen_expression(expr)?;
                        let cond_doc = doc! {"$match" : {"$expr": cond}};
                        (Some(cond_doc), Some(let_bindings))
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
            Set(s) => {
                let left = self.codegen_stage(*s.left)?;
                let right = self.codegen_stage(*s.right)?;

                let mut output_registry = self.mapping_registry.clone();
                output_registry
                    .merge(left.mapping_registry.clone())
                    .merge(right.mapping_registry.clone());

                let union_body = match right.collection {
                    Some(c) => doc! {"coll": c, "pipeline": right.pipeline},
                    None => doc! {"pipeline": right.pipeline},
                };

                Ok(left
                    .with_mapping_registry(output_registry)
                    .with_additional_stage(doc! {"$unionWith": union_body}))
            }
        }
    }

    fn with_merged_mappings(mut self, mappings: MappingRegistry) -> Self {
        self.mapping_registry.merge(mappings);
        self
    }

    fn codegen_group_by(&self, group: ir::Group) -> Result<MqlTranslation> {
        use bson::doc;
        use ir::AggregationFunction::*;

        let source_translation = self.codegen_stage(*group.source)?;
        let expression_generator = self
            .clone()
            .with_merged_mappings(source_translation.mapping_registry.clone());

        let mut group_keys = doc! {};
        let mut bot_body = doc! {};
        let mut project_body = doc! {"_id": 0};
        let mut counter = 1;

        // There shouldn't be any duplicate group key aliases post-algebrization.
        let unique_aliases = group
            .keys
            .iter()
            .filter_map(|k| k.alias.clone())
            .collect::<BTreeSet<String>>();

        for key in group.keys {
            // Project each key under the __bot namespace, unless the key is an unaliased
            // compound identifier, in which case it should be nested under its original
            // namespace and field name.
            let key_alias = match key.alias {
                Some(ref alias) => {
                    bot_body.insert(alias, format!("$_id.{}", alias));
                    alias.to_string()
                }
                None => {
                    let alias = MqlCodeGenerator::get_unique_alias(
                        unique_aliases.clone(),
                        format!("__unaliasedKey{}", counter),
                    );
                    if let ir::Expression::FieldAccess(ref fa) = key.inner {
                        // Throw an error if the FieldAccess expr doesn't resolve to a
                        // compound identifier.
                        match expression_generator
                            .codegen_expression(*fa.expr.clone())?
                            .as_str()
                        {
                            Some(datasource) => {
                                let stripped_ds =
                                    datasource.strip_prefix('$').unwrap_or(datasource);
                                // _id will be 0 instead of a Document. We assume it is
                                // always a Document in the else.
                                if stripped_ds == "_id" {
                                    project_body.insert(
                                        stripped_ds,
                                        bson::bson! {{fa.field.clone(): format!("$_id.{}", alias)}},
                                    )
                                } else {
                                    match project_body.get_mut(stripped_ds) {
	                                    // We can safely unwrap here since `doc` will always be a
	                                    // Bson::Document.
	                                    Some(doc) => {
	                                        doc
	                                        .as_document_mut()
	                                        .unwrap()
	                                        .insert(fa.field.clone(), format!("$_id.{}", alias))}
	                                    None => project_body.insert(
	                                        stripped_ds,
	                                        bson::bson! {{fa.field.clone(): format!("$_id.{}", alias)}},
	                                    ),
	                                }
                                }
                            }
                            None => return Err(Error::InvalidGroupKey),
                        };
                    };
                    alias
                }
            };
            group_keys.insert(
                key_alias,
                expression_generator.codegen_expression(key.inner)?,
            );
            counter += 1;
        }
        let mut group_body = doc! {
            "_id": group_keys,
        };

        // Reset `unique_aliases` since there is no risk of conflict in the $group stage
        // between the aliases nested under _id and the aliases at the same level as it.
        let mut unique_aliases = BTreeSet::new();
        unique_aliases.insert("_id".into());

        for agg in group.aggregations {
            let agg_func_doc = match agg.inner {
                AggregationExpr::CountStar(distinct) => {
                    doc! {"$sqlCount": {
                        "var": "$$ROOT".to_string(),
                        "distinct": distinct
                    }}
                }
                AggregationExpr::Function(f) => {
                    let var = expression_generator.codegen_expression(*f.arg)?;
                    let distinct_body = doc! {
                        "var": var.clone(),
                        "distinct": f.distinct
                    };
                    match f.function {
                        AddToArray => doc! {"$push": var},
                        Avg => doc! {"$sqlAvg": distinct_body},
                        Count => doc! {"$sqlCount": distinct_body},
                        First => doc! {"$first": var},
                        Last => doc! {"$sqlLast": var},
                        Max => doc! {"$max": var},
                        MergeDocuments => doc! {"$sqlMergeObjects": distinct_body},
                        Min => doc! {"$min": var},
                        StddevPop => doc! {"$sqlStdDevPop": distinct_body},
                        StddevSamp => doc! {"$sqlStdDevSamp": distinct_body},
                        Sum => doc! {"$sqlSum": distinct_body},
                    }
                }
            };
            // If an aggregated expression's alias is "_id", prepend additional underscores
            // until there is no longer a conflict in the $group stage.
            let unique_agg_alias = match agg.alias.as_str() {
                "_id" => {
                    MqlCodeGenerator::get_unique_alias(unique_aliases.clone(), "__id".to_string())
                }
                a => a.to_string(),
            };
            bot_body.insert(agg.alias, format!("${}", unique_agg_alias));
            group_body.insert(unique_agg_alias, agg_func_doc);
        }
        if !bot_body.is_empty() {
            project_body.insert("__bot", bot_body);
        }
        Ok(source_translation
            .with_additional_stage(doc! {"$group": group_body})
            .with_additional_stage(doc! {"$project": project_body}))
    }

    fn codegen_sort(&self, sort: ir::Sort) -> Result<MqlTranslation> {
        use bson::doc;
        use ir::{
            Expression::{FieldAccess, Reference},
            SortSpecification::*,
        };

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
                    Desc(expr) => (*expr, Bson::Int32(-1)),
                };

                match expr {
                    Reference(_) | FieldAccess(_) => Ok(()),
                    _ => Err(Error::InvalidSortKey),
                }?;

                // We still need to ensure that the result is a string, since not all FieldAccess
                // expressions will translate to single MQL references
                let key = match expression_generator.codegen_expression(expr)? {
                    Bson::String(s) => Ok(s[1..].to_string()),
                    _ => Err(Error::InvalidSortKey),
                }?;

                Ok((key, direction))
            })
            .collect::<Result<bson::Document>>()?;

        Ok(source_translation.with_additional_stage(doc! {"$sort": sort_specs}))
    }

    pub fn codegen_subquery(
        &self,
        subquery: Stage,
        output_expr: Option<ir::Expression>,
    ) -> Result<Bson> {
        use serde::{Deserialize, Serialize};
        #[derive(Serialize, Deserialize)]
        struct SubqueryBody {
            #[serde(skip_serializing_if = "Option::is_none")]
            db: Option<String>,
            #[serde(skip_serializing_if = "Option::is_none")]
            collection: Option<String>,
            #[serde(rename = "let")]
            let_bindings: bson::Document,
            #[serde(skip_serializing_if = "Option::is_none", rename = "outputPath")]
            output_path: Option<Vec<String>>,
            pipeline: Vec<bson::Document>,
        }

        let (let_bindings, new_mapping_registry) = self.clone().generate_let_bindings()?;
        let subquery_generator = MqlCodeGenerator {
            mapping_registry: self
                .clone()
                .with_merged_mappings(new_mapping_registry)
                .mapping_registry,
            scope_level: self.scope_level + 1,
        };
        let subquery_translation = subquery_generator.codegen_stage(subquery)?;

        let output_path = output_expr
            .map(|e| subquery_translation.generate_path_components(e))
            .transpose()?;

        Ok(bson::to_bson(&SubqueryBody {
            db: subquery_translation.database,
            collection: subquery_translation.collection,
            let_bindings,
            output_path,
            pipeline: subquery_translation.pipeline,
        })
        .unwrap())
    }

    /// Recursively generates a translation for this expression. When
    /// this function is called, `self.mapping_registry` should
    /// include mappings for all datasources in scope.
    pub fn codegen_expression(&self, expr: ir::Expression) -> Result<bson::Bson> {
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
                            if k.starts_with('$') || k.contains('.') {
                                Err(Error::DotsOrDollarsInDocumentKey)
                            } else {
                                Ok((k, self.codegen_expression(v)?))
                            }
                        })
                        .collect::<Result<bson::Document>>()?
                }
            })),
            FieldAccess(fa) => {
                let expr = self.codegen_expression(*fa.expr)?;
                Ok(if fa.field.contains('.') || fa.field.starts_with('$') {
                    bson::bson!({"$getField": {"field": fa.field, "input": expr}})
                } else {
                    match expr {
                        Bson::String(e) => Bson::String(format!("{}.{}", e, fa.field)),
                        _ => bson::bson!({"$getField": {"field": fa.field, "input": expr}}),
                    }
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
                    | Between | And | Or | NullIf | Coalesce | Slice | Position | Substring
                    | MergeObjects => {
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
            Subquery(s) => Ok(
                bson::bson!({ "$subquery":  self.codegen_subquery(*s.subquery, Some(*s.output_expr))? }),
            ),
            SubqueryComparison(s) => {
                use ir::{SubqueryComparisonOp::*, SubqueryModifier::*};
                let modifier = match s.modifier {
                    Any => "any",
                    All => "all",
                };
                let op = match s.operator {
                    Lt => "lt",
                    Lte => "lte",
                    Neq => "neq",
                    Eq => "eq",
                    Gt => "gt",
                    Gte => "gte",
                };
                Ok(bson::bson!({"$subqueryComparison": {
                    "op": op,
                    "modifier": modifier,
                    "arg": self.codegen_expression(*s.argument)?,
                    "subquery": self.codegen_subquery(*s.subquery_expr.subquery, Some(*s.subquery_expr.output_expr))?,
                }}))
            }
            Exists(e) => Ok(bson::bson!({ "$subqueryExists":  self.codegen_subquery(*e, None)? })),
            Is(expr) => Ok(match expr.target_type {
                TypeOrMissing::Number => {
                    Bson::Document(bson::doc! {"$isNumber": self.codegen_expression(*expr.expr)?})
                }
                TypeOrMissing::Type(Type::Null) => {
                    let expr_code = self.codegen_expression(*expr.expr)?;
                    Bson::Document(
                        bson::doc! {"$or": [{"$eq": [{"$type": &expr_code}, "missing"]},
                        {"$eq": [{"$type": &expr_code}, "null"]}]},
                    )
                }
                t => Bson::Document(
                    bson::doc! {"$eq": [{"$type": self.codegen_expression(*expr.expr)?}, t.mql_type().unwrap()]},
                ),
            }),
        }
    }
}
