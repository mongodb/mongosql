use crate::{
    air,
    mapping_registry::{Key, MqlMappingRegistry, MqlMappingRegistryValue, MqlReferenceType},
    mir,
    translator::{Error, MqlTranslator, Result},
    util::ROOT_NAME,
};
use mongosql_datastructures::binding_tuple::DatasourceName;
use mongosql_datastructures::{
    unique_linked_hash_map, unique_linked_hash_map::UniqueLinkedHashMap,
};
use std::collections::BTreeSet;

impl MqlTranslator {
    pub(crate) fn translate_stage(&mut self, mir_stage: mir::Stage) -> Result<air::Stage> {
        match mir_stage {
            mir::Stage::Filter(f) => self.translate_filter(f),
            mir::Stage::Project(p) => self.translate_project(p),
            mir::Stage::Group(g) => self.translate_group(g),
            mir::Stage::Limit(l) => self.translate_limit(l),
            mir::Stage::Offset(o) => self.translate_offset(o),
            mir::Stage::Sort(s) => self.translate_sort(s),
            mir::Stage::Collection(c) => self.translate_collection(c),
            mir::Stage::Array(arr) => self.translate_array_stage(arr),
            mir::Stage::Join(j) => self.translate_join(j),
            mir::Stage::Set(s) => self.translate_set(s),
            mir::Stage::Derived(d) => self.translate_derived(d),
            mir::Stage::Unwind(u) => self.translate_unwind(u),
            mir::Stage::MQLIntrinsic(i) => self.translate_mql_intrinsic(i),
            mir::Stage::Sentinel => unreachable!(),
        }
    }

    fn translate_filter(&mut self, mir_filter: mir::Filter) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_filter.source)?;
        let expr_translation = self.translate_expression(mir_filter.condition)?;

        Ok(air::Stage::Match(air::Match::ExprLanguage(
            air::ExprLanguage {
                source: Box::new(source_translation),
                expr: Box::new(expr_translation),
            },
        )))
    }

    fn translate_project(&mut self, mir_project: mir::Project) -> Result<air::Stage> {
        // Store the incoming mapping registry so we can maintain references
        // to outer scopes in the output registry.
        let mut outer_scope_mr = self.mapping_registry.clone();

        // When we translate expressions in this project, we'll also want to
        // maintain references to outer scopes, so we clone the translator.
        //
        // The reason we need expr_translator AND outer_scope_mr is that the
        // expr_translator must be extended with the mapping registry from the
        // source of this Project, which are only in scope for the translation
        // of this stage, whereas the outer_scope_mr is not extended with those
        // mappings. Instead, outer_scope_mr is extended with the new mappings
        // created by this Project's translation.
        let mut expr_translator = self.clone();

        // Translate the source stage.
        let source_translation = self.translate_stage(*mir_project.source)?;

        // Extend the expr_translator with mappings from the source. This means
        // we can use expr_translator to translate expressions in this Project
        // because expr_translator contains mappings from outer scopes as well
        // as from this Project's source.
        expr_translator
            .mapping_registry
            .merge(self.mapping_registry.clone());

        // We will add mappings to the mapping registry introduced by this Project Stage, which
        // is all of the keys. Previous bindings are removed after the subexpressions are
        // translated because Project kills all its inputs.
        let unique_bot_name =
            self.ensure_unique_datasource_name("__bot".to_string(), &mir_project.expression);
        let mut project_body = UniqueLinkedHashMap::new();
        let mut output_registry = MqlMappingRegistry::new();
        for (k, e) in mir_project.expression.clone().into_iter() {
            let mapped_k = self.get_mapped_project_name(
                &k.datasource,
                &unique_bot_name,
                &mir_project.expression,
            )?;

            project_body.insert(
                mapped_k.clone(),
                air::ProjectItem::Assignment(expr_translator.translate_expression(e)?),
            )?;
            output_registry.insert(
                k,
                MqlMappingRegistryValue::new(mapped_k, MqlReferenceType::FieldRef),
            );
        }

        // Extend the outer_scope_mr with the output mappings for this stage's
        // translation.
        outer_scope_mr.merge(output_registry);

        self.mapping_registry = outer_scope_mr;
        Ok(air::Stage::Project(air::Project {
            source: Box::new(source_translation),
            specifications: project_body,
        }))
    }

    fn translate_group(&mut self, mir_group: mir::Group) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_group.source)?;

        // specifications are the top level projection fields. Every key will
        // be a Datasource. Unaliased group keys will need to be inserted
        // directly into the specifications with the proper Datasource name.
        let mut specifications = UniqueLinkedHashMap::new();
        // bot_body is a Document containing all the field mappings under the Bot Datasource.
        // These will be all aliased group keys and all the aggregations.
        let mut bot_body = UniqueLinkedHashMap::new();

        // map the group key aliases and translate the expressions. Unliased keys will be Projected
        // straight into the specifications
        let keys = self.translate_group_keys(mir_group.keys, &mut specifications, &mut bot_body)?;

        // map the group aggregation aliases and translate the expresisons.
        let aggregations =
            self.translate_group_aggregations(mir_group.aggregations, &mut bot_body)?;

        // Fixup mapping_registry to contain only the output of the final Project Stage for this
        // GROUP BY.
        self.mapping_registry = MqlMappingRegistry::new();

        for key in specifications.keys() {
            self.mapping_registry.insert(
                Key::named(key, self.scope_level),
                MqlMappingRegistryValue::new(key.clone(), MqlReferenceType::FieldRef),
            );
        }
        // bot_body will only be empty if all Group keys are references and there are not
        // aggregation functions. If it is not empty, we need to add it to the output Project
        // Stage under the unique bot name.
        if !bot_body.is_empty() {
            let unique_bot_name = Self::generate_unique_datasource_name("__bot".to_string(), |s| {
                specifications.contains_key(s)
            });
            specifications.insert(
                unique_bot_name.clone(),
                air::ProjectItem::Assignment(air::Expression::Document(bot_body)),
            )?;
            self.mapping_registry.insert(
                Key::bot(self.scope_level),
                MqlMappingRegistryValue::new(unique_bot_name, MqlReferenceType::FieldRef),
            );
        }

        // Return the proper Stages, which is a Project with specifications set as above and the
        // source being the generated Group Stage.
        Ok(air::Stage::Project(air::Project {
            source: Box::new(air::Stage::Group(air::Group {
                source: Box::new(source_translation),
                keys,
                aggregations,
            })),
            specifications,
        }))
    }

    fn translate_limit(&mut self, mir_limit: mir::Limit) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_limit.source)?;
        let limit =
            i64::try_from(mir_limit.limit).or(Err(Error::LimitOutOfI64Range(mir_limit.limit)))?;

        Ok(air::Stage::Limit(air::Limit {
            source: Box::new(source_translation),
            limit,
        }))
    }

    fn translate_offset(&mut self, mir_offset: mir::Offset) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_offset.source)?;

        Ok(air::Stage::Skip(air::Skip {
            source: Box::new(source_translation),
            skip: mir_offset.offset,
        }))
    }

    fn translate_sort(&mut self, mir_sort: mir::Sort) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_sort.source)?;
        let specifications = mir_sort
            .specs
            .into_iter()
            .map(|spec| {
                Ok(match spec {
                    mir::SortSpecification::Asc(mir_field_path) => {
                        air::SortSpecification::Asc(self.get_field_path_name(mir_field_path)?)
                    }
                    mir::SortSpecification::Desc(mir_field_path) => {
                        air::SortSpecification::Desc(self.get_field_path_name(mir_field_path)?)
                    }
                })
            })
            .collect::<Result<Vec<air::SortSpecification>>>()?;
        Ok(air::Stage::Sort(air::Sort {
            source: source_translation.into(),
            specs: specifications,
        }))
    }

    fn translate_collection(&mut self, mir_collection: mir::Collection) -> Result<air::Stage> {
        self.mapping_registry.insert(
            Key::named(&mir_collection.collection, self.scope_level),
            MqlMappingRegistryValue::new(ROOT_NAME.to_string(), MqlReferenceType::Variable),
        );

        Ok(air::Stage::Collection(air::Collection {
            db: mir_collection.db,
            collection: mir_collection.collection,
        }))
    }

    fn translate_array_stage(&mut self, mir_arr: mir::ArraySource) -> Result<air::Stage> {
        self.mapping_registry.insert(
            Key::named(&mir_arr.alias, self.scope_level),
            MqlMappingRegistryValue::new(ROOT_NAME.to_string(), MqlReferenceType::Variable),
        );

        Ok(air::Stage::Documents(air::Documents {
            array: mir_arr
                .array
                .iter()
                .map(|mir_expr| self.translate_expression(mir_expr.clone()))
                .collect::<Result<Vec<air::Expression>>>()?,
        }))
    }

    fn translate_join(&mut self, mir_join: mir::Join) -> Result<air::Stage> {
        let join_type = match mir_join.join_type {
            mir::JoinType::Inner => air::JoinType::Inner,
            mir::JoinType::Left => air::JoinType::Left,
        };

        // We need to store the left_registry to generate the let bindings
        // since only datasources from the left will be bound to variables.
        let left = self.translate_stage(*mir_join.left)?;
        let left_registry = self.mapping_registry.clone();

        // When translating the right side, we need the Translator to indicate
        // this context because naming conflicts can occur between the left and
        // right datasources if the left datasources need to be renamed.
        // Therefore, we store the current is_join value, set is_join to true,
        // translate the right side, and then restore the old is_join value.
        // Project translation considers the is_join information to determine
        // if name conflicts need to be resolved or not.
        let previous_is_join = self.is_join;
        self.is_join = true;
        let right = self.translate_stage(*mir_join.right)?;
        self.is_join = previous_is_join;

        let mut let_vars = None;
        let condition = mir_join
            .condition
            .map(|x| {
                let_vars = Some(self.generate_let_bindings(left_registry.clone()));
                self.translate_expression(x)
            })
            .transpose()?;

        // Restore the original mappings for the left datasource since they may have been
        // overwritten to map to Variable references when translating the condition.
        self.mapping_registry.merge(left_registry);

        Ok(air::Stage::Join(air::Join {
            join_type,
            left: Box::new(left),
            right: Box::new(right),
            let_vars,
            condition,
        }))
    }

    fn translate_set(&mut self, mir_set: mir::Set) -> Result<air::Stage> {
        let source = self.translate_stage(*mir_set.left)?;
        let left_registry = self.mapping_registry.clone();
        let pipeline = self.translate_stage(*mir_set.right)?;

        // Ensure we retain mappings from the right side _and_ the left.
        self.mapping_registry.merge(left_registry);

        Ok(air::Stage::UnionWith(air::UnionWith {
            source: Box::new(source),
            pipeline: Box::new(pipeline),
        }))
    }

    fn translate_derived(&mut self, mir_derived: mir::Derived) -> Result<air::Stage> {
        self.scope_level += 1;
        let derived_translation = self.translate_stage(*mir_derived.source)?;
        self.scope_level -= 1;
        Ok(derived_translation)
    }

    fn translate_unwind(&mut self, mir_unwind: mir::Unwind) -> Result<air::Stage> {
        Ok(air::Stage::Unwind(air::Unwind {
            source: Box::new(self.translate_stage(*mir_unwind.source)?),
            path: self.translate_field_path_to_expresssion(mir_unwind.path)?,
            index: mir_unwind.index,
            outer: mir_unwind.outer,
        }))
    }

    fn translate_mql_intrinsic(&mut self, mir_mql_intrinsic: mir::MQLStage) -> Result<air::Stage> {
        match mir_mql_intrinsic {
            mir::MQLStage::EquiJoin(ej) => self.translate_equijoin(ej),
            mir::MQLStage::LateralJoin(lj) => self.translate_lateral_join(lj),
            mir::MQLStage::MatchFilter(mf) => self.translate_match_filter(mf),
        }
    }

    fn translate_equijoin(&mut self, mir_join: mir::EquiJoin) -> Result<air::Stage> {
        let join_type = match mir_join.join_type {
            mir::JoinType::Inner => air::JoinType::Inner,
            mir::JoinType::Left => air::JoinType::Left,
        };

        let source = self.translate_stage(*mir_join.source)?;
        let from = self.translate_stage(*mir_join.from)?;

        // The `from` must be of the form Project{source: Collection, ...}, we will discard
        // the Project since EquiJoins do not have actual pipelines.
        let collection = if let air::Stage::Project(p) = from {
            if let air::Stage::Collection(c) = *p.source {
                c
            } else {
                return Err(Error::ExpectedCollection);
            }
        } else {
            return Err(Error::ExpectedCollection);
        };

        let translate_to_field_ref = |e: mir::Expression| -> Result<air::FieldRef> {
            let e = self.translate_expression(e)?;
            if let air::Expression::FieldRef(f) = e {
                Ok(f)
            } else {
                Err(Error::ExpectedFieldRef)
            }
        };
        let local_field = translate_to_field_ref(*mir_join.local_field)?;
        let foreign_field = translate_to_field_ref(*mir_join.foreign_field)?;

        Ok(air::Stage::EquiJoin(air::EquiJoin {
            join_type,
            source: Box::new(source),
            from: collection,
            local_field,
            foreign_field,
        }))
    }

    fn translate_lateral_join(&mut self, _mir_join: mir::LateralJoin) -> Result<air::Stage> {
        // SQL-1576
        todo!()
    }

    fn translate_match_filter(&mut self, mir_match_filter: mir::MatchFilter) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_match_filter.source)?;
        let condition_translation = self.translate_match_query(mir_match_filter.condition)?;

        Ok(air::Stage::Match(air::Match::MatchLanguage(
            air::MatchLanguage {
                source: Box::new(source_translation),
                expr: Box::new(condition_translation),
            },
        )))
    }

    // Utility functions for Group translation
    fn get_unique_alias(existing_aliases: &BTreeSet<String>, mut alias: String) -> String {
        while existing_aliases.contains(&alias) {
            alias.insert(0, '_')
        }
        alias
    }

    fn translate_agg_function(afa: mir::AggregationFunction) -> air::AggregationFunction {
        match afa {
            mir::AggregationFunction::AddToArray => air::AggregationFunction::AddToArray,
            mir::AggregationFunction::Avg => air::AggregationFunction::Avg,
            mir::AggregationFunction::Count => air::AggregationFunction::Count,
            mir::AggregationFunction::First => air::AggregationFunction::First,
            mir::AggregationFunction::Last => air::AggregationFunction::Last,
            mir::AggregationFunction::Max => air::AggregationFunction::Max,
            mir::AggregationFunction::MergeDocuments => air::AggregationFunction::MergeDocuments,
            mir::AggregationFunction::Min => air::AggregationFunction::Min,
            mir::AggregationFunction::StddevPop => air::AggregationFunction::StddevPop,
            mir::AggregationFunction::StddevSamp => air::AggregationFunction::StddevSamp,
            mir::AggregationFunction::Sum => air::AggregationFunction::Sum,
        }
    }

    fn get_datasource_and_field_for_unaliased_group_key<'a>(
        &'a self,
        e: &'a mir::Expression,
    ) -> Result<(&'a Key, &'a MqlMappingRegistryValue, &'a String)> {
        let (key, field) = match e {
            mir::Expression::FieldAccess(mir::FieldAccess {
                ref expr,
                ref field,
                ..
            }) => match **expr {
                mir::Expression::Reference(mir::ReferenceExpr { ref key, .. }) => (key, field),
                _ => return Err(Error::InvalidGroupKey),
            },
            _ => return Err(Error::InvalidGroupKey),
        };
        Ok((
            key,
            self.mapping_registry
                .get(key)
                .ok_or_else(|| Error::ReferenceNotFound(key.clone()))?,
            field,
        ))
    }

    fn translate_group_keys(
        &self,
        keys: Vec<mir::OptionallyAliasedExpr>,
        specifications: &mut UniqueLinkedHashMap<String, air::ProjectItem>,
        bot_body: &mut UniqueLinkedHashMap<String, air::Expression>,
    ) -> Result<Vec<air::NameExprPair>> {
        let unique_aliases = keys
            .iter()
            .filter_map(|k| k.get_alias().map(String::from))
            .collect::<BTreeSet<_>>();

        let mut translated_keys = Vec::new();

        let make_key_ref = |name| air::Expression::FieldRef(format!("_id.{name}").into());
        for (i, k) in keys.into_iter().enumerate() {
            match k {
                mir::OptionallyAliasedExpr::Aliased(ae) => {
                    // an aliased key will be projected under Bot
                    translated_keys.push(air::NameExprPair {
                        name: ae.alias.clone(),
                        expr: self.translate_expression(ae.expr)?,
                    });
                    bot_body.insert(ae.alias.clone(), make_key_ref(ae.alias))?;
                }
                // An unaliased key will be projected under its originating Datasource.
                // After the Aliasing rewrite pass, an unaliased key can only exist when it is a
                // FieldAccess, and the parent of the FieldAccess *must* be a Datasource.
                mir::OptionallyAliasedExpr::Unaliased(e) => {
                    let position_counter = i + 1;
                    let unique_name = Self::get_unique_alias(
                        &unique_aliases,
                        format!("__unaliasedKey{position_counter}"),
                    );
                    let (og_key, mapped_datasource, field) =
                        self.get_datasource_and_field_for_unaliased_group_key(&e)?;
                    let datasource_name = match mapped_datasource.ref_type {
                        MqlReferenceType::FieldRef => mapped_datasource.name.clone(),
                        // If the MappingRegistryValue is a Variable, we should use the
                        // original Key's name since that is what we actually want to
                        // project the group values under, not the internal Variable name.
                        MqlReferenceType::Variable => match &og_key.datasource {
                            DatasourceName::Named(n) => n.clone(),
                            // Cannot have Bottom mapped to a Variable in this context.
                            DatasourceName::Bottom => return Err(Error::InvalidGroupKey),
                        },
                    };
                    match specifications.get_mut(&datasource_name) {
                        // If we have already put something under this Datasource, we just update
                        // the document.
                        Some(air::ProjectItem::Assignment(air::Expression::Document(ref mut d))) => {
                            d.insert(field.clone(), make_key_ref(unique_name.clone()))?
                        }
                        // We have nothing under this Datasource, so we need to create a new
                        // Document with one key/value pair.
                        None => specifications.insert(
                            datasource_name.clone(),
                            air::ProjectItem::Assignment(air::Expression::Document(unique_linked_hash_map! {field.clone() => make_key_ref(unique_name.clone())})),
                        )?,
                        // We only generate Project fields where the values are Documents because the keys of
                        // the Project are Datasources. If the following case is hit, it is a bug in
                        // the code.
                        _ => unreachable!(),
                    }
                    translated_keys.push(air::NameExprPair {
                        name: unique_name,
                        expr: self.translate_expression(e)?,
                    });
                }
            }
        }

        Ok(translated_keys)
    }

    fn translate_group_aggregations(
        &self,
        aggregations: Vec<mir::AliasedAggregation>,
        bot_body: &mut UniqueLinkedHashMap<String, air::Expression>,
    ) -> Result<Vec<air::AccumulatorExpr>> {
        let mut unique_aliases = aggregations
            .iter()
            .map(|k| k.alias.clone())
            .collect::<BTreeSet<_>>();
        // the Group keys will be under _id in MQL, so we need to rename any alias that is _id.
        unique_aliases.insert("_id".to_string());

        let mut translated_aggregations = Vec::new();

        for a in aggregations.into_iter() {
            let alias = a.alias;
            let unique_alias = if alias.as_str() == "_id" {
                Self::get_unique_alias(&unique_aliases, "_id".to_string())
            } else {
                alias.clone()
            };
            let (function, distinct, arg) = match a.agg_expr {
                mir::AggregationExpr::CountStar(b) => (
                    air::AggregationFunction::Count,
                    b,
                    Box::new(air::Expression::Literal(air::LiteralValue::Integer(1))),
                ),
                mir::AggregationExpr::Function(afa) => (
                    Self::translate_agg_function(afa.function),
                    afa.distinct,
                    Box::new(self.translate_expression(*afa.arg)?),
                ),
            };
            bot_body.insert(
                alias.clone(),
                air::Expression::FieldRef(unique_alias.clone().into()),
            )?;
            translated_aggregations.push(air::AccumulatorExpr {
                alias: unique_alias,
                function,
                distinct,
                arg,
            });
        }

        Ok(translated_aggregations)
    }
}
