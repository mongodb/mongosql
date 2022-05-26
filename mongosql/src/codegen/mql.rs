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
use std::collections::{BTreeMap, BTreeSet};

#[derive(PartialEq, Debug, Clone)]
pub struct MqlMappingRegistry(BTreeMap<Key, String>);

impl MqlMappingRegistry {
    pub fn new() -> Self {
        MqlMappingRegistry(BTreeMap::new())
    }

    pub fn get(&self, k: &Key) -> Option<&String> {
        self.0.get(k)
    }

    pub fn remove(&mut self, k: &Key) -> Option<String> {
        self.0.remove(k)
    }

    pub fn insert<K: Into<Key>, V: Into<String>>(&mut self, k: K, v: V) -> Option<String> {
        self.0.insert(k.into(), v.into())
    }

    pub fn merge(&mut self, other: MqlMappingRegistry) -> &mut Self {
        self.0.extend(other.0.into_iter());
        self
    }
}

impl Default for MqlMappingRegistry {
    fn default() -> Self {
        Self::new()
    }
}

#[derive(PartialEq, Debug)]
pub struct MqlTranslation {
    pub database: Option<String>,
    pub collection: Option<String>,
    pub mapping_registry: MqlMappingRegistry,
    pub pipeline: Vec<bson::Document>,
}

impl MqlTranslation {
    fn with_additional_stage(mut self, stage: bson::Document) -> Self {
        self.pipeline.push(stage);
        self
    }

    fn with_mapping_registry(self, mapping_registry: MqlMappingRegistry) -> Self {
        MqlTranslation {
            mapping_registry,
            ..self
        }
    }

    /// generate_path_components takes an expression and returns a vector of
    /// its components by recursively tracing its path.
    pub fn generate_path_components(&self, expr: ir::Expression) -> Result<Vec<String>> {
        match expr {
            ir::Expression::Reference(ir::ReferenceExpr { key, .. }) => {
                match self.mapping_registry.get(&key) {
                    Some(name) => Ok(vec![name.clone()]),
                    None => Err(Error::ReferenceNotFound(key)),
                }
            }
            ir::Expression::FieldAccess(fa) => {
                let mut path = self.generate_path_components(*fa.expr)?;
                path.push(fa.field.clone());
                Ok(path)
            }
            _ => Err(Error::NoFieldPathForExpr),
        }
    }

    /// replace_bot will be called after the codegen pipeline to add a stage
    /// that replaces the unique bottom name __bot with an empty string "".
    pub fn replace_bot(mut self) -> Self {
        use bson::doc;

        let key = Key {
            datasource: DatasourceName::Bottom,
            scope: 0u16,
        };
        let mongo_bot_name = self.mapping_registry.remove(&key);
        return match mongo_bot_name {
            Some(name) => {
                self.mapping_registry.insert(key, "");
                self.with_additional_stage(doc! {"$replaceWith":
                    {"$unsetField":
                        {"field": name.clone(),
                         "input":
                            {"$setField":
                                {"field": "",
                                 "input": "$$ROOT",
                                 "value": format! ("${}", name)
                                }
                            }
                        }
                    }
                })
            }
            None => self,
        };
    }
}

#[derive(Clone)]
pub struct MqlCodeGenerator {
    pub mapping_registry: MqlMappingRegistry,
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
            Log => "$sqlLog",
            Round => "$sqlRound",

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
            Day => "$dayOfMonth",
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
    /// Generate a unique bottom name given a predicate closure. Keeps pre-pending
    /// `_` until the predicate returns false, indicating that that name is not in use.
    fn generate_unique_bot_name<F>(name_exists: F) -> String
    where
        F: Fn(&String) -> bool,
    {
        let mut ret = "__bot".to_string();
        while name_exists(&ret) {
            ret.insert(0, '_');
        }
        ret
    }
    fn get_unique_bot_name(project_names: &BindingTuple<ir::Expression>) -> String {
        if project_names.is_empty() {
            return "__bot".to_string();
        }
        let current_scope = project_names.keys().next().unwrap().scope;
        MqlCodeGenerator::generate_unique_bot_name(|s| {
            project_names.contains_key(&(s.clone(), current_scope).into())
        })
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
    /// to a generated name of the form lower(<datasource name>)_<nesting depth>. It returns
    /// the resulting $let document along with a new mapping registry that contains the
    /// generated names. Naming conflicts are resolved by appending underscores until the
    /// generated name is unique. We must lowercase the datasource name and append underscores
    /// (as opposed to prepend) because aggregation variables are only allowed to start with
    /// lowercase ASCII letters or non-ASCII characters.
    fn generate_let_bindings(self) -> Result<(bson::Document, MqlMappingRegistry)> {
        let mut let_bindings: bson::Document = map![];
        let new_mapping_registry = MqlMappingRegistry(
            self.mapping_registry
                .0
                .into_iter()
                .map(|(key, value)| {
                    let mut generated_name = format!(
                        "{}_{}",
                        MqlCodeGenerator::get_datasource_name(&key.datasource, "__bot"),
                        key.scope
                    )
                    .to_ascii_lowercase();
                    while let_bindings.contains_key(&generated_name) {
                        generated_name.push('_');
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
            Filter(f) => {
                let source_translation = self.codegen_stage(*f.source)?;
                let expression_generator = self
                    .clone()
                    .with_merged_mappings(source_translation.mapping_registry.clone());
                let expr = expression_generator.codegen_expression(f.condition)?;
                Ok(source_translation.with_additional_stage(doc! {"$match": {"$expr": expr}}))
            }
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
                let mut output_registry = MqlMappingRegistry::new();
                let unique_bot_name = MqlCodeGenerator::get_unique_bot_name(&p.expression);
                // we need to project away _id unless the query maps _id some other way.
                // {_id: 0} will be overwritten if _id is defined in the project expression.
                let mut project_body = doc! {"_id": 0};
                for (k, e) in p.expression.into_iter() {
                    let mapped_k =
                        MqlCodeGenerator::get_datasource_name(&k.datasource, &unique_bot_name);
                    if mapped_k.starts_with('$')
                        || mapped_k.contains('.')
                        || mapped_k.as_str() == ""
                    {
                        return Err(Error::InvalidProjectField);
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
            Limit(l) => {
                let limit = i64::try_from(l.limit).or(Err(Error::LimitOutOfI64Range(l.limit)))?;
                Ok(self
                    .codegen_stage(*l.source)?
                    .with_additional_stage(doc! {"$limit": limit}))
            }
            Offset(o) => {
                let offset =
                    i64::try_from(o.offset).or(Err(Error::OffsetOutOfI64Range(o.offset)))?;
                Ok(self
                    .codegen_stage(*o.source)?
                    .with_additional_stage(doc! {"$skip": offset}))
            }
            Sort(s) => self.codegen_sort(s),
            Collection(c) => Ok(MqlTranslation {
                database: Some(c.db),
                collection: Some(c.collection.clone()),
                mapping_registry: MqlMappingRegistry(
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
                let mapping_registry = MqlMappingRegistry(
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
                            .with_merged_mappings(MqlMappingRegistry(left_registry.0))
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
            Derived(s) => MqlCodeGenerator {
                mapping_registry: self.mapping_registry.clone(),
                scope_level: self.scope_level + 1,
            }
            .codegen_stage(*s.source),
        }
    }

    fn with_merged_mappings(mut self, mappings: MqlMappingRegistry) -> Self {
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
        let mut output_registry = MqlMappingRegistry::new();

        // There shouldn't be any duplicate group key aliases post-algebrization.
        let unique_aliases = group
            .keys
            .iter()
            .filter_map(|k| k.get_alias().map(String::from))
            .collect::<BTreeSet<String>>();

        for key in group.keys {
            // Project each key under the __bot namespace, unless the key is an unaliased
            // compound identifier, in which case it should be nested under its original
            // namespace and field name.
            let (expr, key_alias) = match key {
                ir::OptionallyAliasedExpr::Aliased(ir::AliasedExpr {
                    ref expr,
                    ref alias,
                }) => {
                    bot_body.insert(alias, format!("$_id.{}", alias));
                    (expr.clone(), alias.to_string())
                }
                ir::OptionallyAliasedExpr::Unaliased(expr) => {
                    let alias = MqlCodeGenerator::get_unique_alias(
                        unique_aliases.clone(),
                        format!("__unaliasedKey{}", counter),
                    );
                    if let ir::Expression::FieldAccess(ref fa) = expr {
                        // We know that after the aliasing rewrite pass, any unaliased group key can
                        // only have one layer of field access. Nested field accesses would have been
                        // rewritten to aliased expressions, so at this point the left hand side expression
                        // of the FieldAccess inside an OptionallyAliasedExpression must be a Reference.
                        let key = match *fa.expr {
                            ir::Expression::Reference(ir::ReferenceExpr { ref key, .. }) => key,
                            _ => return Err(Error::InvalidGroupKey),
                        };
                        let datasource = source_translation
                            .mapping_registry
                            .0
                            .get(key)
                            .ok_or_else(|| Error::ReferenceNotFound(key.clone()))?;
                        // _id will be 0 instead of a Document. We assume it is
                        // always a Document in the else.
                        if datasource == "_id" {
                            output_registry.insert(key.clone(), datasource);
                            project_body.insert(
                                datasource,
                                bson::bson! {{fa.field.clone(): format!("$_id.{}", alias)}},
                            )
                        } else {
                            match project_body.get_mut(datasource) {
                                // We can safely unwrap here since `doc` will always be a
                                // Bson::Document.
                                Some(doc) => doc
                                    .as_document_mut()
                                    .unwrap()
                                    .insert(fa.field.clone(), format!("$_id.{}", alias)),
                                None => {
                                    output_registry.insert(key.clone(), datasource);
                                    project_body.insert(
                                        datasource,
                                        bson::bson! {{fa.field.clone(): format!("$_id.{}", alias)}},
                                    )
                                }
                            }
                        };
                    };
                    (expr, alias)
                }
            };
            group_keys.insert(key_alias, expression_generator.codegen_expression(expr)?);
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
            let agg_func_doc = match agg.agg_expr {
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
                        AddToArray => {
                            if f.distinct {
                                doc! {"$addToSet": var}
                            } else {
                                doc! {"$push": var}
                            }
                        }
                        Avg => doc! {"$sqlAvg": distinct_body},
                        Count => doc! {"$sqlCount": distinct_body},
                        First => doc! {"$first": var},
                        Last => doc! {"$sqlLast": distinct_body},
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
            let bot_name =
                MqlCodeGenerator::generate_unique_bot_name(|s| project_body.contains_key(s));
            project_body.insert(bot_name.clone(), bot_body);
            output_registry.insert(Key::bot(self.scope_level), bot_name);
        }
        Ok(source_translation
            .with_additional_stage(doc! {"$group": group_body})
            .with_additional_stage(doc! {"$project": project_body})
            .with_mapping_registry(output_registry))
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
        use ir::{Expression::*, LiteralValue::*};
        match expr {
            Literal(lit) => Ok(bson::bson!({
                "$literal": match lit.value {
                    Null => Bson::Null,
                    Boolean(b) => Bson::Boolean(b),
                    String(s) => Bson::String(s),
                    Integer(i) => Bson::Int32(i),
                    Long(l) => Bson::Int64(l),
                    Double(d) => Bson::Double(d),
                },
            })),
            Reference(ir::ReferenceExpr { key, .. }) => self
                .mapping_registry
                .0
                .get(&key)
                .ok_or(Error::ReferenceNotFound(key))
                .map(|s| Bson::String(format!("${}", s))),
            Array(exprs) => Ok(Bson::Array(
                exprs
                    .array
                    .into_iter()
                    .map(|e| self.codegen_expression(e))
                    .collect::<Result<Vec<Bson>>>()?,
            )),
            Document(ir::DocumentExpr { document, .. }) => Ok(Bson::Document({
                if document.is_empty() {
                    bson::doc! {"$literal": {}}
                } else {
                    document
                        .into_iter()
                        .map(|(k, v)| {
                            if k.starts_with('$') || k.contains('.') || k.as_str() == "" {
                                Err(Error::InvalidDocumentKey)
                            } else {
                                Ok((k, self.codegen_expression(v)?))
                            }
                        })
                        .collect::<Result<bson::Document>>()?
                }
            })),
            FieldAccess(fa) => {
                let expr = self.codegen_expression(*fa.expr)?;
                Ok(
                    if fa.field.contains('.')
                        || fa.field.starts_with('$')
                        || fa.field.as_str() == ""
                    {
                        bson::bson!({"$getField": {"field": fa.field, "input": expr}})
                    } else {
                        match expr {
                            Bson::String(e) => Bson::String(format!("{}.{}", e, fa.field)),
                            _ => bson::bson!({"$getField": {"field": fa.field, "input": expr}}),
                        }
                    },
                )
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

                Ok(bson::bson!({
                    "$let": {
                        "vars": {"target": self.codegen_expression(*ce.expr)?},
                        "in": {
                            "$switch": {
                                "branches": br,
                                "default": self.codegen_expression(*ce.else_branch)?
                            }
                        }
                    }
                }))
            }
            Cast(ce) => {
                let convert_op = match ce.to {
                    ir::Type::Document | ir::Type::Array => "$sqlConvert",
                    _ => "$convert",
                };

                Ok(bson::bson!({
                    convert_op: {
                        "input": self.codegen_expression(*ce.expr)?,
                        "to": ce.to.mql_type(),
                        "onNull":  self.codegen_expression(*ce.on_null)?,
                        "onError":self.codegen_expression(*ce.on_error)?
                    }
                }))
            }
            TypeAssertion(ta) => self.codegen_expression(*ta.expr),
            ScalarFunction(sa) => {
                use crate::ir::ScalarFunction::*;
                Ok(match sa.function {
                    Pos => self.codegen_expression(sa.args[0].clone())?,
                    Neg => bson::bson!({
                        "$multiply" : [
                             self.codegen_expression(sa.args[0].clone())?,
                             {"$literal":
                                 Bson::Int32(-1),
                             },
                        ]
                    }),
                    ComputedFieldAccess => {
                        // Adding this feature is tracked in SQL-673
                        return Err(Error::UnsupportedFunction(ComputedFieldAccess));
                    }
                    CurrentTimestamp => Bson::String("$$NOW".to_string()),
                    CharLength | OctetLength | Size | Upper | Lower | Year | Month | Day | Hour
                    | Minute | Second => {
                        bson::bson!({ sa.function.mql_op().unwrap(): self.codegen_expression(sa.args[0].clone())?})
                    }
                    BitLength => bson::bson!({
                        "$multiply" : [
                            { sa.function.mql_op().unwrap(): self.codegen_expression(sa.args[0].clone())?},
                            {"$literal":
                                Bson::Int32(8),
                            },
                        ]
                    }),
                    BTrim | LTrim | RTrim => bson::bson!({
                        sa.function.mql_op().unwrap(): {"input": self.codegen_expression(sa.args[1].clone())?,
                            "chars": self.codegen_expression(sa.args[0].clone())?,
                    }}),
                    Div => {
                        // Div will always have exactly two arguments because it's parsed as a BinaryOp
                        Bson::Document(bson::doc! {
                            sa.function.mql_op().unwrap(): {
                                "dividend": self.codegen_expression(sa.args[0].clone())?,
                                "divisor": self.codegen_expression(sa.args[1].clone())?,
                                "onError": {"$literal": Bson::Null}
                            }
                        })
                    }
                    Not | Concat | Add | Sub | Mul | Lt | Lte | Neq | Eq | Gt | Gte | Between
                    | And | Or | NullIf | Coalesce | Slice | Substring | MergeObjects | Log
                    | Round => {
                        let args = Bson::Array(
                            sa.args
                                .into_iter()
                                .map(|e| self.codegen_expression(e))
                                .collect::<Result<Vec<Bson>>>()?,
                        );
                        Bson::Document(bson::doc! { sa.function.mql_op().unwrap(): args})
                    }
                    // scalar functions with reversed argument order.
                    Position => {
                        let args = Bson::Array(
                            sa.args
                                .into_iter()
                                .rev()
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
                    Neq => "ne",
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
            Exists(ir::ExistsExpr { stage: e, .. }) => {
                Ok(bson::bson!({ "$subqueryExists":  self.codegen_subquery(*e, None)? }))
            }
            Is(expr) => Ok(match expr.target_type {
                TypeOrMissing::Number => {
                    Bson::Document(bson::doc! {"$isNumber": self.codegen_expression(*expr.expr)?})
                }
                TypeOrMissing::Type(Type::Null) => {
                    let expr_code = self.codegen_expression(*expr.expr)?;
                    Bson::Document(
                        bson::doc! {"$or": [{"$eq": [{"$type": &expr_code}, "null"]},
                        {"$eq": [{"$type": &expr_code}, "missing"]}]},
                    )
                }
                t => Bson::Document(
                    bson::doc! {"$eq": [{"$type": self.codegen_expression(*expr.expr)?}, t.mql_type().unwrap()]},
                ),
            }),
        }
    }
}
