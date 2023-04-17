use super::{Error, MqlCodeGenerator, MqlTranslation, Result};
use crate::air::{self, AggregationFunction};
use bson::{bson, doc, Bson};

impl MqlCodeGenerator {
    pub fn codegen_air_stage(&self, stage: air::Stage) -> Result<MqlTranslation> {
        match stage {
            air::Stage::Project(p) => self.codegen_project(p),
            air::Stage::Group(g) => self.codegen_group(g),
            air::Stage::Limit(l) => self.codegen_limit(l),
            air::Stage::Sort(s) => self.codegen_sort(s),
            air::Stage::Collection(c) => self.codegen_collection(c),
            air::Stage::Join(j) => self.codegen_join(j),
            air::Stage::Unwind(u) => self.codegen_unwind(u),
            air::Stage::Lookup(l) => self.codegen_lookup(l),
            air::Stage::ReplaceWith(r) => self.codegen_replace_with(r),
            air::Stage::Match(m) => self.codegen_match(m),
            air::Stage::UnionWith(_u) => Err(Error::UnimplementedAIR),
            air::Stage::Skip(s) => self.codegen_skip(s),
            air::Stage::Documents(d) => self.codegen_documents(d),
        }
    }

    fn codegen_sort(&self, air_sort: air::Sort) -> Result<MqlTranslation> {
        use air::SortSpecification::*;

        let source_translation = self.codegen_air_stage(*air_sort.source)?;
        let mut pipeline = source_translation.pipeline;
        let sort_specs = air_sort
            .specs
            .into_iter()
            .map(|spec| {
                let (key, direction) = match spec {
                    Asc(key) => (key, Bson::Int32(1)),
                    Desc(key) => (key, Bson::Int32(-1)),
                };
                Ok((key, direction))
            })
            .collect::<Result<bson::Document>>()?;

        pipeline.push(doc! {"$sort": sort_specs});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_match(&self, air_match: air::Match) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_match.source)?;
        let mut pipeline = source_translation.pipeline;
        let expr = bson!({ "$expr": self.codegen_air_expression(*air_match.expr)?});

        pipeline.push(doc! {"$match": expr});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_replace_with(&self, air_replace_with: air::ReplaceWith) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_replace_with.source)?;
        let mut pipeline = source_translation.pipeline;
        let expr = self.codegen_air_expression(*air_replace_with.new_root)?;

        pipeline.push(doc! {"$replaceWith": expr});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_documents(&self, air_docs: air::Documents) -> Result<MqlTranslation> {
        let docs = air_docs
            .array
            .into_iter()
            .map(|e| self.codegen_air_expression(e))
            .collect::<Result<Vec<Bson>>>()?;
        Ok(MqlTranslation {
            database: None,
            collection: None,
            pipeline: vec![doc! {"$documents": Bson::Array(docs)}],
        })
    }

    fn codegen_collection(&self, air_coll: air::Collection) -> Result<MqlTranslation> {
        Ok(MqlTranslation {
            database: Some(air_coll.db),
            collection: Some(air_coll.collection),
            pipeline: vec![],
        })
    }

    fn codegen_lookup(&self, air_lookup: air::Lookup) -> Result<MqlTranslation> {
        let lookup_pipeline_translation = self.codegen_air_stage(*air_lookup.pipeline)?.pipeline;
        let source_translation = self.codegen_air_stage(*air_lookup.source)?;
        let mut pipeline = source_translation.pipeline;
        let mut lookup_doc = doc! {
            "as": air_lookup.as_var,
            "pipeline": lookup_pipeline_translation,
        };
        match (air_lookup.from_db, air_lookup.from_coll) {
            (None, Some(from_coll)) => lookup_doc.extend(doc! {"from": from_coll}),
            (Some(from_db), Some(from_coll)) => {
                lookup_doc.extend(doc! {"from": {"db": from_db, "coll": from_coll}})
            }
            _ => {}
        }
        if let Some(let_vars) = air_lookup.let_vars {
            lookup_doc.extend(doc! { "let":
                let_vars
                    .into_iter()
                    .map(|v| Ok((v.name, self.codegen_air_expression(*v.expr)?)))
                    .collect::<Result<bson::Document>>()?
            });
        }
        pipeline.push(doc! {"$lookup": lookup_doc});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_project(&self, air_project: air::Project) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_project.source)?;
        let mut pipeline = source_translation.pipeline;
        let mut project_doc = air_project
            .specifications
            .into_iter()
            .map(|(k, v)| Ok((k, self.codegen_air_expression(v)?)))
            .collect::<Result<bson::Document>>()?;
        if !project_doc.contains_key("_id") {
            // we create a temporary so that _id: 0 will always be the first element
            // in the doc. This does add another linear factor to the code, but makes
            // testing easier.
            let mut tmp_project_doc = doc! {"_id": 0};
            tmp_project_doc.extend(project_doc);
            project_doc = tmp_project_doc;
        }
        pipeline.push(doc! {"$project": project_doc});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_group(&self, air_group: air::Group) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_group.source)?;
        let mut pipeline = source_translation.pipeline;
        let id_doc = air_group
            .keys
            .into_iter()
            .map(|air::NameExprPair { name: k, expr: v }| Ok((k, self.codegen_air_expression(v)?)))
            .collect::<Result<bson::Document>>()?;
        let mut group_doc = doc! {"_id": id_doc};
        let aggs = air_group
            .aggregations
            .into_iter()
            .map(
                |air::AccumulatorExpr {
                     alias,
                     function,
                     distinct,
                     arg,
                 }| {
                    Ok(if function == AggregationFunction::AddToArray {
                        if distinct {
                            (alias, bson!({ "$addToSet": self.codegen_air_expression(*arg)? }))
                        }
                        else{
                            (alias, bson!({ "$push": self.codegen_air_expression(*arg)? }))
                        }
                    }
                    else if distinct || function == AggregationFunction::Count {
                        (alias, bson!({ Self::agg_func_to_sql_op(function): {"var": self.codegen_air_expression(*arg)?, "distinct": distinct }}))
                    } else {
                        (alias, bson!({ Self::agg_func_to_mql_op(function): self.codegen_air_expression(*arg)? }))
                    })
                },
            )
            .collect::<Result<bson::Document>>()?;
        group_doc.extend(aggs);
        pipeline.push(doc! {"$group": group_doc});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_unwind(&self, air_unwind: air::Unwind) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_unwind.source)?;
        let mut pipeline = source_translation.pipeline;
        let mut include_array_index = air_unwind.index;
        // If the `path` field_ref has a parent, we need to crawl up the parent chain
        // and find the root parent and append the argument to `includeArrayIndex` to that.
        // If the path does not have a parent, then we're already at the root level of the document
        // and can just use the argument as is.
        if include_array_index.is_some() {
            if let Some(field_ref) = match &*air_unwind.path {
                air::Expression::FieldRef(f) => Some(f),
                _ => None,
            } {
                if field_ref.parent.is_some() {
                    include_array_index = Some(format!(
                        "{}.{}",
                        field_ref.root_parent(),
                        include_array_index.unwrap()
                    ));
                }
            }
        }
        let path = self.codegen_air_expression(*air_unwind.path)?;
        let preserve_null_and_empty_arrays = air_unwind.outer;

        let unwind_body = match (
            include_array_index.is_some(),
            preserve_null_and_empty_arrays,
        ) {
            (true, true) => {
                doc! {"path": path, "includeArrayIndex": include_array_index.unwrap(), "preserveNullAndEmptyArrays": preserve_null_and_empty_arrays}
            }
            (true, false) => doc! {"path": path, "includeArrayIndex": include_array_index.unwrap()},
            (false, true) => {
                doc! {"path": path, "preserveNullAndEmptyArrays": preserve_null_and_empty_arrays}
            }
            (false, false) => doc! {"path": path},
        };

        pipeline.push(doc! {"$unwind": unwind_body});

        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_skip(&self, air_skip: air::Skip) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_skip.source)?;
        let mut pipeline = source_translation.pipeline;
        pipeline.push(doc! {"$skip": Bson::Int64(air_skip.skip)});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_limit(&self, air_limit: air::Limit) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_limit.source)?;
        let mut pipeline = source_translation.pipeline;
        pipeline.push(doc! {"$limit": Bson::Int64(air_limit.limit)});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_join(&self, air_join: air::Join) -> Result<MqlTranslation> {
        let mut left_translation = self.codegen_air_stage(*air_join.left)?;
        let right_translation = self.codegen_air_stage(*air_join.right)?;
        let join_type = match air_join.join_type {
            air::JoinType::Inner => "inner",
            air::JoinType::Left => "left",
        };
        let mut join_doc = doc! {};
        if let Some(right_db) = right_translation.database.clone() {
            if left_translation.database != right_translation.database {
                join_doc.insert("database", right_db);
            }
        }

        if right_translation.collection.is_some() {
            join_doc.insert("collection", right_translation.collection);
        }

        join_doc.insert("joinType", join_type);

        if air_join.let_vars.is_some() {
            join_doc.insert(
                "let",
                air_join
                    .let_vars
                    .unwrap()
                    .iter()
                    .map(|v| {
                        Ok((
                            v.name.clone(),
                            self.codegen_air_expression(*v.expr.clone())?,
                        ))
                    })
                    .collect::<Result<bson::Document>>()?,
            );
        }

        join_doc.insert("pipeline", right_translation.pipeline);
        if air_join.condition.is_some() {
            let cond = self.codegen_air_expression(air_join.condition.unwrap())?;
            join_doc.insert("condition", doc! {"$match": {"$expr": cond}});
        }

        left_translation.pipeline.push(doc! {"$join": join_doc});
        Ok(left_translation)
    }
}
