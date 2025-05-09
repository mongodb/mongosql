use crate::{
    array_element_schema_or_error, get_schema_for_path, get_schema_for_path_mut,
    insert_required_key_into_document, promote_missing, remove_field, schema_difference,
    schema_for_bson, schema_for_type_numeric, schema_for_type_str, Error, MatchConstrainSchema,
    Result,
};
use agg_ast::definitions::{
    AtlasSearchStage, Bucket, BucketAuto, ConciseSubqueryLookup, Densify, Documents,
    EqualityLookup, Expression, Fill, FillOutput, GraphLookup, Group, LiteralValue, Lookup,
    LookupFrom, Namespace, ProjectItem, ProjectStage, Ref, SetWindowFields, Stage, SubqueryLookup,
    TaggedOperator, UnionWith, Unset, UntaggedOperator, UntaggedOperatorName, Unwind,
};
use linked_hash_map::LinkedHashMap;
use mongosql::{
    json_schema, map, schema::{
        Atomic, Document, Satisfaction, Schema, ANY_DOCUMENT, DATE_OR_NULLISH, EMPTY_DOCUMENT,
        INTEGRAL, NULLISH, NULLISH_OR_UNDEFINED, NUMERIC, NUMERIC_OR_NULLISH,
    }, set
};
use std::{
    collections::{BTreeMap, BTreeSet},
    sync::LazyLock,
};

pub static SEARCH_META: LazyLock<Schema> = LazyLock::new(|| {
    Schema::Document(Document {
        keys: map! {
            "count".to_string() => Schema::Document(Document {
                keys: map! {
                    "total".to_string() => Schema::Atomic(Atomic::Long),
                    "lowerBound".to_string() => Schema::Atomic(Atomic::Long),
                },
                // one key is required, but the result will always include one or
                // the other, so we cannot say that either is required.
                required: set![],
                ..Default::default()
            }),
        },
        required: set!["count".to_string(),],
        ..Default::default()
    })
});

#[allow(dead_code)]
pub(crate) trait DeriveSchema {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema>;
}

#[allow(dead_code)]
#[derive(Debug, Clone, PartialEq, Eq, PartialOrd, Ord)]
pub struct ResultSetState<'a> {
    pub catalog: &'a BTreeMap<Namespace, Schema>,
    pub variables: BTreeMap<String, Schema>,
    pub result_set_schema: Schema,
    pub current_db: String,
    // the null_behavior field allows us to keep track of what behavior we are expecting to be exhibited
    // by the rows returned by this query. This comes up in both normal schema derivation, where something like
    // $eq: [null, {$op: ...}] can influence the values returned by the operator), as well as in match schema derivation
    // where more broadly things like null field references or a falsifiable return type (e.g. {$eq: [{$op: ...}, 0])
    // may influcence they types of values the underlying result_set_schema can contain.
    pub null_behavior: Satisfaction,
    pub accumulator_stage: bool,
}

impl<'a> ResultSetState<'a> {
    pub fn new(catalog: &'a BTreeMap<Namespace, Schema>, current_db: String) -> Self {
        Self {
            catalog,
            variables: BTreeMap::new(),
            result_set_schema: Schema::Any,
            current_db,
            null_behavior: Satisfaction::Not,
            accumulator_stage: false,
        }
    }
}

/// derive_schema_for_pipeline is the main entrypoint for schema derivation. It takes in a pipeline,
/// the collection to run that pipeline on (the db will be part of the result set state), and the
/// other relevant context, and produces a schema. It can also be used for subpipelines in stages
/// such as $lookup and $facet.
pub fn derive_schema_for_pipeline(
    pipeline: Vec<Stage>,
    current_collection: Option<String>,
    state: &mut ResultSetState,
) -> Result<Schema> {
    // when this function is first called, we'd like to seed the result set schema with the collection
    // we are starting with, if specified. Any subsquent calls will not have a current_collection, so this
    // can only happen during the entrypoint to schema derivation
    if state.result_set_schema == Schema::Any {
        if let Some(collection) = current_collection {
            if let Some(schema) = state.catalog.get(&Namespace::new(
                state.current_db.clone(),
                collection.clone(),
            )) {
                state.result_set_schema = schema.clone()
            }
        }
    }
    pipeline.iter().try_for_each(|stage| {
        state.result_set_schema = stage.derive_schema(state)?;
        Ok(())
    })?;
    Ok(Schema::simplify(&std::mem::take(
        &mut state.result_set_schema,
    )))
}

impl DeriveSchema for Stage {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        /// densify_derive_schema derives the schema for a $densify stage. Any field specified in the partition
        /// by becomes required, and any field originally in the schema that is not mentioned becomes non-required.
        fn densify_derive_schema(densify: &Densify, state: &mut ResultSetState) -> Result<Schema> {
            // create a list of all the fields that densify references explicitly -- that is the partition by fields and
            // the actual field being densified.
            let mut paths: Vec<Vec<String>> = densify
                .partition_by_fields
                .clone()
                .unwrap_or_default()
                .iter()
                .map(|field| {
                    field
                        .split(".")
                        .map(|s| s.to_string())
                        .collect::<Vec<String>>()
                })
                .collect();
            paths.push(
                densify
                    .field
                    .split(".")
                    .map(|s| s.to_string())
                    .collect::<Vec<String>>(),
            );
            // we create a doc that contains all the required fields with their schemas, marking the full path to each
            // field as required. unioning this document with the existing schema will function like a mask on the required
            // fields, removing any field not part of the $densify as required (but preserving the required fields  that are part
            // of the stage). Fields that are part of the stage but not required do not become required, because the original documents
            // still persist.
            let mut required_doc = Schema::Document(Document {
                additional_properties: false,
                ..Default::default()
            });
            paths.into_iter().for_each(|path| {
                if let Some(field_schema) =
                    get_schema_for_path_mut(&mut state.result_set_schema, path.clone())
                {
                    insert_required_key_into_document(
                        &mut required_doc,
                        field_schema.clone(),
                        path.clone(),
                        true,
                    );
                }
            });
            Ok(state
                .result_set_schema
                .to_owned()
                .document_union(required_doc))
        }

        /// documents_derive_schema derives the schema for a $documents stage. $documents can either be a list of document
        /// literals, in which case we union the schema of those literals, or an expression, in which case we derive the
        /// schema for that expression.
        fn documents_derive_schema(
            documents: &Documents,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            match documents {
                Documents::Literals(docs) => {
                    // we use folding to get the schema for each document, and union them together, to get the resulting schema
                    let schema = docs.iter().try_fold(
                        None,
                        |schema: Option<Schema>, document: &LinkedHashMap<String, Expression>| {
                            // here we convert the map of field namespace - expression to a resulting map of
                            // field namespace - field schema. We collect in such a way that we can get the error from any derivation.
                            let doc_fields = document
                                .into_iter()
                                .filter(|(field, expr)| {
                                    !(*field == "$literal"
                                        && *expr == &Expression::Document(map!()))
                                })
                                .map(|(field, expr)| {
                                    let field_schema = expr.derive_schema(state)?;
                                    Ok((field.clone(), field_schema))
                                })
                                .collect::<Result<BTreeMap<String, Schema>>>()?;
                            let doc_schema = Schema::Document(Document {
                                required: doc_fields.keys().cloned().collect(),
                                keys: doc_fields,
                                ..Default::default()
                            });
                            Ok(match schema {
                                None => Some(doc_schema),
                                Some(schema) => Some(schema.union(&doc_schema)),
                            })
                        },
                    );
                    Ok(schema?.unwrap_or(Schema::AnyOf(set!(
                        Schema::Missing,
                        Schema::Document(Document::any())
                    ))))
                }
                Documents::Expr(expr) => expr.derive_schema(state),
            }
        }

        /// facet_derive_schema derives the schema for a $facet stage. It adds fields to the existing
        /// result set schema based on an underlying pipeline that we recursively call derive_schema_for_pipeline on
        fn facet_derive_schema(
            facet: &LinkedHashMap<String, Vec<Stage>>,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            // facet contains key - value pairs where the key is the field of a field in the output document,
            // and the value is an aggregation pipeline. We can use the same generic helper for aggregating over a whole pipeline
            // to get the schema for each field, cloning the incoming state.
            let facet_schema = facet
                .into_iter()
                .map(|(field, pipeline)| {
                    let mut field_state = state.clone();
                    let field_schema = Schema::Array(Box::new(derive_schema_for_pipeline(
                        pipeline.clone(),
                        None,
                        &mut field_state,
                    )?));
                    Ok((field.clone(), field_schema))
                })
                .collect::<Result<BTreeMap<String, Schema>>>()?;
            Ok(Schema::Document(Document {
                required: facet_schema.keys().cloned().collect(),
                keys: facet_schema,
                ..Default::default()
            }))
        }

        /// fill_derive_schema derives the schema for a $fill stage. For each field in the output
        /// map, we generate the schema and add insert it into the result set schema if it does not
        /// exist, or union it with the existing schema if it does.
        fn fill_derive_schema(fill: &Fill, state: &mut ResultSetState) -> Result<Schema> {
            for (path, fill_output) in fill.output.iter() {
                // Every key that appears in the output can no longer be missing, and can only be
                // null if the fill value is null.
                let path_vec = path
                    .split('.')
                    .map(|s| s.to_string())
                    .collect::<Vec<String>>();
                match fill_output {
                    FillOutput::Value(e) => {
                        let fill_schema = e.derive_schema(state)?;
                        match get_schema_for_path_mut(
                            &mut state.result_set_schema,
                            path_vec.clone(),
                        ) {
                            Some(path_schema) => {
                                *path_schema = std::mem::take(path_schema)
                                    .subtract_nullish()
                                    .union(&fill_schema);
                            }
                            None => {
                                insert_required_key_into_document(
                                    &mut state.result_set_schema,
                                    fill_schema,
                                    path_vec.clone(),
                                    true,
                                );
                            }
                        }
                    }
                    // The method does not matter, either of the currently supported methods will
                    // remove null and missing from the schema, and cannot change the schema in
                    // any other meaningful way.
                    FillOutput::Method(_m) => {
                        let path_schema =
                            get_schema_for_path_mut(&mut state.result_set_schema, path_vec)
                                .ok_or_else(|| Error::UnknownReference(path.clone()))?;
                        *path_schema = std::mem::take(path_schema).subtract_nullish();
                    }
                }
            }
            Ok(state.result_set_schema.to_owned())
        }

        /// graph_lookup_derive_schema derives the schema for a $graphLookup stage. Ultimately, this reduces to a collection
        /// lookup that we insert under the "depth_field" field name.
        fn graph_lookup_derive_schema(
            graph_lookup: &GraphLookup,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            // it is a bit annoying that we need to clone these Strings just to do a lookup, but I
            // don't think there is a way around that.
            let from_field = Namespace::new(state.current_db.clone(), graph_lookup.from.clone());
            let mut from_schema = state
                .catalog
                .get(&from_field)
                .ok_or_else(|| Error::UnknownReference(from_field.to_string()))?
                .clone();
            if let Some(ref depth_field) = graph_lookup.depth_field {
                insert_required_key_into_document(
                    &mut from_schema,
                    Schema::Atomic(Atomic::Long),
                    depth_field
                        .as_str()
                        .split('.')
                        .map(|s| s.to_string())
                        .collect(),
                    true,
                );
            }
            insert_required_key_into_document(
                &mut state.result_set_schema,
                Schema::Array(Box::new(from_schema.clone())),
                graph_lookup
                    .as_var
                    .as_str()
                    .split('.')
                    .map(|s| s.to_string())
                    .collect(),
                true,
            );
            Ok(state.result_set_schema.to_owned())
        }

        /// group_derive_schema derives schema for $group stages. The output will be any fields specified
        /// by the group (with their respective schemas) as well as _id
        fn group_derive_schema(group: &Group, state: &mut ResultSetState) -> Result<Schema> {
            state.accumulator_stage = true;
            // group is a map of field namespace to expression. We can derive the schema for each expression
            // and then union them together to get the resulting schema.
            let mut keys = group
                .aggregations
                .iter()
                .map(|(k, e)| {
                    let field_schema = e.derive_schema(state)?.upconvert_missing_to_null();
                    Ok((k.to_string(), field_schema))
                })
                .collect::<Result<BTreeMap<String, Schema>>>()?;
            let id_schema = group.keys.derive_schema(state)?.upconvert_missing_to_null();
            keys.insert("_id".to_string(), id_schema);
            state.accumulator_stage = false;
            Ok(Schema::Document(Document {
                required: keys.keys().cloned().collect(),
                keys,
                ..Default::default()
            }))
        }

        // add_fields_derive_schema is near set_window_fields_derive_schema because they are necessarily
        // identical in semantics: they both add fields to the schema incoming from the previous
        // pipeline stage. The only difference is that add_fields is a map of expressions, while set_window_fields
        // is a map of window functions.
        fn add_fields_derive_schema(
            add_fields: &LinkedHashMap<String, Expression>,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            for (field, expression) in add_fields.iter() {
                let expr_schema = expression.derive_schema(state)?;
                let path = field
                    .split(".")
                    .map(|s| s.to_string())
                    .collect::<Vec<String>>();
                insert_required_key_into_document(
                    &mut state.result_set_schema,
                    expr_schema,
                    path,
                    true,
                );
            }
            Ok(Schema::simplify(&state.result_set_schema).to_owned())
        }

        /// bucket_output_derive_keys is a helper for the $bucket and $bucketAuto stages.
        /// It produces a schema based on the output field if specified. If not, "count"
        /// will be the only field.
        fn bucket_output_derive_keys(
            output: Option<&LinkedHashMap<String, Expression>>,
            state: &mut ResultSetState,
        ) -> Result<BTreeMap<String, Schema>> {
            if let Some(output) = output {
                let keys = output
                    .iter()
                    .map(|(k, e)| {
                        let field_schema = e.derive_schema(state)?;
                        Ok((k.clone(), field_schema))
                    })
                    .collect::<Result<BTreeMap<String, Schema>>>()?;
                Ok(keys)
            } else {
                Ok(map! {
                    "count".to_string() => Schema::AnyOf(set!(Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)))
                })
            }
        }

        /// bucket_derive_schema derives the schema for a $bucket stage. The schema is defined by the output field,
        /// and contains _id as well.
        fn bucket_derive_schema(bucket: &Bucket, state: &mut ResultSetState) -> Result<Schema> {
            let mut id_schema = bucket.group_by.derive_schema(state)?;
            if let Some(default) = bucket.default.as_ref() {
                let default_schema = schema_for_bson(default);
                id_schema = id_schema.union(&default_schema);
            }
            let mut keys = bucket_output_derive_keys(bucket.output.as_ref(), state)?;
            keys.insert("_id".to_string(), id_schema);
            Ok(Schema::Document(Document {
                required: keys.keys().cloned().collect(),
                keys,
                ..Default::default()
            }))
        }

        /// bucket_auto_derive_schema derives the schema for a $bucketAuto stage. The schema is defined by the output field,
        /// and contains _id as well.
        fn bucket_auto_derive_schema(
            bucket: &BucketAuto,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            // The actual _id type will be a Document with min and max keys where the types of
            // those keys are this value denoted id_type.
            let id_type = bucket.group_by.derive_schema(state)?;
            let id_schema = Schema::Document(Document {
                required: set!["min".to_string(), "max".to_string()],
                keys: map! {
                    "min".to_string() => id_type.clone(),
                    "max".to_string() => id_type
                },
                ..Default::default()
            });
            let mut keys = bucket_output_derive_keys(bucket.output.as_ref(), state)?;
            keys.insert("_id".to_string(), id_schema);
            Ok(Schema::Document(Document {
                required: keys.keys().cloned().collect(),
                keys,
                ..Default::default()
            }))
        }

        /// set_window_fields_derive_schema derives the schema for a $setWindowFields stage. The schema is
        /// defined by the output field
        fn set_window_fields_derive_schema(
            set_windows: &SetWindowFields,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            state.accumulator_stage = true;
            let mut result_doc = state.result_set_schema.clone();
            for (k, e) in set_windows.output.clone() {
                let field_schema = e
                    .window_func
                    .derive_schema(state)?
                    .upconvert_missing_to_null();
                let path = k.split('.').map(|s| s.to_string()).collect::<Vec<String>>();
                insert_required_key_into_document(&mut result_doc, field_schema, path, true);
            }
            state.accumulator_stage = false;
            Ok(result_doc)
        }

        #[derive(Debug)]
        struct ProjectPathNode {
            pub children_nodes: LinkedHashMap<String, ProjectPathNode>,
            pub project_item: Option<ProjectItem>,
        }

        /// process_project_paths takes in the set of ProjectItems from the $project stage and
        /// creates a tree from the tree paths. Each "node" of the tree knows about each of its
        /// subpaths, as well as any project item terminating at that node.
        fn process_project_paths(
            items: &LinkedHashMap<String, ProjectItem>,
        ) -> LinkedHashMap<String, ProjectPathNode> {
            let mut project_paths = ProjectPathNode {
                children_nodes: LinkedHashMap::new(),
                project_item: None,
            };
            for (k, v) in items
                .iter()
                .filter(|(_, p)| !matches!(p, ProjectItem::Exclusion))
            {
                // if the top level key for a path is not present in the root node, create an empty
                // node for it
                let path = k.split('.').map(|s| s.to_string()).collect::<Vec<String>>();
                if !project_paths.children_nodes.contains_key(&path[0]) {
                    project_paths.children_nodes.insert(
                        path[0].clone(),
                        ProjectPathNode {
                            children_nodes: LinkedHashMap::new(),
                            project_item: None,
                        },
                    );
                }
                // beginning at the root node, descend down the path, adding new nodes
                // for any unencountered paths
                let mut cur = &mut project_paths;
                for (index, field) in path.iter().enumerate() {
                    if !cur.children_nodes.contains_key(field) {
                        cur.children_nodes.insert(
                            field.clone(),
                            ProjectPathNode {
                                children_nodes: LinkedHashMap::new(),
                                project_item: None,
                            },
                        );
                    };
                    cur = cur.children_nodes.get_mut(field).unwrap();
                    // once we reach the terminal field for this path, set that node's
                    // project item; i.e., for {$project: {"a.b.c": 1}}, once we reach "c",
                    // that node's project_item should be Some(ProjectItem::Inclusion)
                    if index == path.len() - 1 {
                        cur.project_item = Some(v.clone());
                    }
                }
            }
            project_paths.children_nodes
        }

        /// get_schema_for_project_node takes a node and generates the schema for all
        /// fields in the project by traversing top down and evaluating any project_items.
        /// the parent schema keeps a reference to the schema for the current node from the
        /// result set schema -- that is, we traverse the result set schema in parallel in
        /// order to properly evaluate inclusions.
        fn get_schema_for_project_node(
            node: &ProjectPathNode,
            parent_schema: Option<&Schema>,
            state: &mut ResultSetState,
        ) -> Schema {
            if let Some(schema) = parent_schema {
                // If the parent schema is missing, the child schema is also missing
                if schema == &Schema::Missing {
                    return Schema::Missing;
                }
                // we keep track of all schemas, both from this node (if it is an inclusion or
                // assignment) and from any children nodes, then make an AnyOf and simplify at the end
                let mut schemas: BTreeSet<Schema> = set!();

                // handle the case where this node is a terminal node (i.e. the field is named in the $project)
                if let Some(project_item) = &node.project_item {
                    match project_item {
                        ProjectItem::Inclusion => {
                            schemas.insert(schema.clone());
                        }
                        ProjectItem::Exclusion => return Schema::Missing,
                        ProjectItem::Assignment(e) => {
                            let schema = e.derive_schema(state);
                            if let Ok(s) = schema {
                                schemas.insert(s);
                            }
                        }
                    }
                }
                match schema {
                    // if the current node is an Array, any child schemas will have to come from nested documents in this array.
                    // evaluate the inner type of the array for any nested fields, then wrap the resulting schema in an Array
                    Schema::Array(a) => {
                        let inner_schema =
                            get_schema_for_project_node(node, Some(a.as_ref()), state);
                        schemas.insert(Schema::Array(Box::new(inner_schema)));
                    }
                    // similarly, anyofs can be part of a nested path if any of schemas are documents (or arrays of documents).
                    // we evaulate each subtype to verify if it contains any of the children nodes
                    Schema::AnyOf(ao) => {
                        ao.iter().for_each(|ao_schema| {
                            let schema = get_schema_for_project_node(node, Some(ao_schema), state);
                            if schema != Schema::Unsat {
                                schemas.insert(schema);
                            }
                        });
                    }
                    // if we want to get a nested path from an Any, we will treat it as a document with
                    // additional properties true or an array of documents with additional properties true.
                    // We should also include missing, as this path won't be required.
                    Schema::Any => {
                        let document_schema = get_schema_for_project_node(
                            node,
                            Some(&Schema::Document(Document::any())),
                            state,
                        );
                        schemas.insert(Schema::Missing);
                        schemas.insert(document_schema.clone());
                        schemas.insert(Schema::Array(Box::new(document_schema)));
                    }
                    // documents are where we actually evaluate the children schemas to see if their paths are present in the document.
                    // we iterate through each child, and build a document from the resulting key-schema pairs.
                    d @ Schema::Document(_) => {
                        let d = promote_missing(d);
                        let mut path_document = Schema::Document(Document::empty());
                        for (child_field, child_node) in node.children_nodes.iter() {
                            let mut parent_schema = d.get_key(child_field);
                            if parent_schema.is_none()
                                && matches!(
                                    d,
                                    Schema::Document(Document {
                                        additional_properties: true,
                                        ..
                                    })
                                )
                            {
                                parent_schema = Some(&Schema::Any);
                            }
                            let child_schema =
                                get_schema_for_project_node(child_node, parent_schema, state);
                            insert_required_key_into_document(
                                &mut path_document,
                                child_schema,
                                vec![child_field.clone()],
                                true,
                            );
                        }
                        schemas.insert(path_document);
                    }
                    _ => {}
                }
                Schema::simplify(&Schema::AnyOf(schemas))
            // assignments are often new fields, so we don't require a parent schema to evaluate
            } else if let Some(ProjectItem::Assignment(e)) = &node.project_item.as_ref() {
                e.derive_schema(state).unwrap()
            } else {
                Schema::Missing
            }
        }

        /// project_derive_schema derives the schema for a $project stage. Described in more detail within the
        /// helpers, it builds a tree of field paths, and traverses the result set schema in parallel to that
        /// tree to evaluate what schemas each inclusion and assignment can take on, building up the output schema.
        fn project_derive_schema(
            project: &ProjectStage,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            if project.items.is_empty() {
                return Err(Error::InvalidProjectStage);
            }
            state.result_set_schema = promote_missing(&state.result_set_schema);
            let exclude_id = project.items.iter().any(|(field, item)| {
                field == &"_id".to_string() && item == &ProjectItem::Exclusion
            });
            // If this is an exclusion $project, we can remove the fields from the schema and
            // return
            if project.items.iter().all(|(k, p)| {
                matches!(p, ProjectItem::Exclusion) || (k == "_id" && project.items.len() > 1)
            }) {
                project.items.iter().for_each(|(k, _)| {
                    remove_field(
                        &mut state.result_set_schema,
                        k.split('.').map(|s| s.to_string()).collect(),
                    );
                });
                // undo the promotion of missing
                state.result_set_schema = Schema::simplify(&state.result_set_schema);
                return Ok(state.result_set_schema.to_owned());
            }

            // if the $project is inclusions, assignments, or both, build up the schemas
            let project_paths = process_project_paths(&project.items);
            let mut result_doc = Schema::Document(Document::empty());

            // for each top level field, generate a schema by traversing the path tree top-down
            for (field, node) in project_paths.iter() {
                let schema = get_schema_for_project_node(
                    node,
                    state.result_set_schema.get_key(field),
                    &mut state.clone(),
                );
                insert_required_key_into_document(
                    &mut result_doc,
                    schema,
                    vec![field.clone()],
                    true,
                );
            }

            // if _id has not been excluded and has not been redefined, include it from the original schema
            if !project_paths.contains_key("_id") && !exclude_id {
                // Only insert _id if it exists in the incoming schema
                if let Some(id_value) = state.result_set_schema.get_key("_id") {
                    insert_required_key_into_document(
                        &mut result_doc,
                        id_value.clone(),
                        vec!["_id".to_string()],
                        true,
                    );
                }
            }

            // undo the promotion of missing
            state.result_set_schema = Schema::simplify(&state.result_set_schema);
            Ok(Schema::simplify(&result_doc))
        }

        /// lookup_derive_schema derives the schema for a $lookup stage by calling the appropriate function based on
        /// the lookup semantics
        fn lookup_derive_schema(lookup: &Lookup, state: &mut ResultSetState) -> Result<Schema> {
            match lookup {
                Lookup::Equality(le) => derive_equality_lookup_schema(le, state),
                Lookup::ConciseSubquery(lc) => derive_concise_lookup_schema(lc, state),
                Lookup::Subquery(ls) => derive_subquery_lookup_schema(ls, state),
            }
        }

        /// from_to_ns is a helper converting the $lookup "from" field into our standard namespace struct
        fn from_to_ns(from: &LookupFrom, state: &ResultSetState) -> Namespace {
            match from {
                LookupFrom::Collection(ref c) => {
                    Namespace::new(state.current_db.clone(), c.clone())
                }
                LookupFrom::Namespace(ref n) => {
                    Namespace::new(n.database.clone(), n.collection.clone())
                }
            }
        }

        /// derive_equality_lookup_schema is a helper for deriving schema for $lookup stages joining another namespace
        /// based on field equality (local field / foreign field). The output schema includes a new field
        /// (the "as" field) which is an array of docs with the from field's schema
        fn derive_equality_lookup_schema(
            lookup: &EqualityLookup,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            let from_ns = from_to_ns(&lookup.from, state);
            let from_schema = state
                .catalog
                .get(&from_ns)
                .ok_or_else(|| Error::UnknownReference(from_ns.into()))?;
            insert_required_key_into_document(
                &mut state.result_set_schema,
                Schema::Array(Box::new(from_schema.clone())),
                lookup
                    .as_var
                    .as_str()
                    .split('.')
                    .map(|s| s.to_string())
                    .collect(),
                true,
            );
            Ok(state.result_set_schema.to_owned())
        }

        /// derive_equality_lookup_schema is a helper for deriving schema for $lookup stages joining another
        /// schema based on a pipeline. The output has a new field, based on "as", that is a vec of documents
        /// with the pipeline's schema
        fn derive_concise_lookup_schema(
            lookup: &ConciseSubqueryLookup,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            derive_subquery_lookup_schema_helper(
                &lookup.let_body,
                &lookup.from,
                &lookup.pipeline,
                &lookup.as_var,
                state,
            )
        }

        /// derive_subquery_lookup_schema is a helper for deriving schema for $lookup stages joining another
        /// schema based on a pipeline. The output has a new field, based on "as", that is a vec of documents
        /// with the pipeline's schema
        fn derive_subquery_lookup_schema(
            lookup: &SubqueryLookup,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            derive_subquery_lookup_schema_helper(
                &lookup.let_body,
                &lookup.from,
                &lookup.pipeline,
                &lookup.as_var,
                state,
            )
        }

        /// derive_subquery_lookup_schema_helper is a helper that determines and inserts the schema of a sub
        /// pipeline for subquery (normal and concise syntax) lookups
        fn derive_subquery_lookup_schema_helper(
            // generally, we do not pass &Option<Type> but usually Option<&Type>, but this is an
            // internal helper function.
            let_body: &Option<LinkedHashMap<String, Expression>>,
            from: &Option<LookupFrom>,
            pipeline: &[Stage],
            as_var: &str,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            // the output of the lookup stage is the current result with the result of the pipeline
            // nested under the as_var. We can derive the schema for the from collection, and then
            // apply the pipeline to that schema to get the result.
            let mut variables = let_body
                .iter()
                .flatten()
                .map(|(k, v)| Ok((k.clone(), v.derive_schema(state)?)))
                .collect::<Result<BTreeMap<String, Schema>>>()?;
            let from_schema = match from {
                None => &Schema::Document(Document::empty()),
                Some(ns) => {
                    let from_ns = from_to_ns(ns, state);
                    state
                        .catalog
                        .get(&from_ns)
                        .ok_or_else(|| Error::UnknownReference(from_ns.into()))?
                }
            };
            variables.extend(state.variables.clone());
            let mut lookup_state = ResultSetState {
                catalog: state.catalog,
                variables,
                result_set_schema: from_schema.clone(),
                current_db: state.current_db.clone(),
                null_behavior: state.null_behavior,
                accumulator_stage: state.accumulator_stage,
            };
            let lookup_schema =
                derive_schema_for_pipeline(pipeline.to_owned(), None, &mut lookup_state)?;
            insert_required_key_into_document(
                &mut state.result_set_schema,
                Schema::Array(Box::new(lookup_schema.clone())),
                as_var.split('.').map(|s| s.to_string()).collect(),
                true,
            );
            Ok(state.result_set_schema.to_owned())
        }

        /// sort_by_count_derive_schema derives the schema for $sortByCount stages. The output contains two fields:
        /// _id (determined by the sort expression) and count
        fn sort_by_count_derive_schema(
            sort_expr: &Expression,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            Ok(Schema::Document(Document {
                keys: map! {
                    "_id".to_string() => sort_expr.derive_schema(state)?,
                    "count".to_string() => Schema::AnyOf(set!(Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)))
                },
                required: set!("_id".to_string(), "count".to_string()),
                ..Default::default()
            }))
        }

        /// union_with_derive_schema derives the schema for $unionWith stages. It gets the new schema defined by a
        /// collection or pipeline, and unions it with the existing result set schema.
        fn union_with_derive_schema(
            union: &UnionWith,
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            match union {
                UnionWith::Collection(c) => {
                    let from_ns = Namespace::new(state.current_db.clone(), c.clone());
                    let from_schema = state
                        .catalog
                        .get(&from_ns)
                        .ok_or_else(|| Error::UnknownReference(from_ns.into()))?;
                    state.result_set_schema = state.result_set_schema.union(from_schema);
                    Ok(state.result_set_schema.to_owned())
                }
                UnionWith::Pipeline(p) => {
                    if p.coll.is_none() && p.pipeline.is_none() {
                        return Err(Error::NotEnoughArguments("$unionWith".to_string()));
                    }
                    let from_schema = match p.coll.clone() {
                        Some(collection) => {
                            let from_ns = Namespace::new(state.current_db.clone(), collection);
                            state
                                .catalog
                                .get(&from_ns)
                                .ok_or_else(|| Error::UnknownReference(from_ns.into()))?
                        }
                        // if we don't have a from collection, the unioned schema will be determined just by the pipeline
                        None => &Schema::default(),
                    };
                    // if we have a pipeline, union the schema it generates. This will use the schema from the previous
                    // step, which is the collection schema if one is specified, or empty if not.
                    if let Some(pipeline) = p.pipeline.clone() {
                        let pipeline_state = &mut ResultSetState {
                            catalog: state.catalog,
                            variables: state.variables.clone(),
                            result_set_schema: from_schema.clone(),
                            current_db: state.current_db.clone(),
                            null_behavior: state.null_behavior,
                            accumulator_stage: state.accumulator_stage,
                        };
                        let pipeline_schema =
                            derive_schema_for_pipeline(pipeline, None, pipeline_state)?;
                        state.result_set_schema = state.result_set_schema.union(&pipeline_schema);
                    // if no pipeline is specified, we are unioning the documents of the collection directly, so just union the from_schema,
                    // which should represent the collection schema
                    } else {
                        state.result_set_schema = state.result_set_schema.union(from_schema);
                    }
                    Ok(state.result_set_schema.to_owned())
                }
            }
        }

        /// unwind_derive_schema derives schema for $unwind stages. It will update the schema for
        /// and existing field to unnest the inner schema from an array, and it will insert a new
        /// field if include_array_index is specified.
        fn unwind_derive_schema(unwind: &Unwind, state: &mut ResultSetState) -> Result<Schema> {
            let (path, preserve_null_and_empty_arrays) = match unwind {
                Unwind::FieldPath(Expression::Ref(Ref::FieldRef(r))) => (
                    Some(r.split(".").map(|s| s.to_string()).collect::<Vec<String>>()),
                    false,
                ),
                Unwind::Document(d) => {
                    if let Expression::Ref(Ref::FieldRef(r)) = d.path.as_ref() {
                        (
                            Some(r.split(".").map(|s| s.to_string()).collect::<Vec<String>>()),
                            d.preserve_null_and_empty_arrays == Some(true),
                        )
                    } else {
                        (None, false)
                    }
                }
                _ => (None, false),
            };
            let mut nullish = preserve_null_and_empty_arrays;
            state.result_set_schema = promote_missing(&state.result_set_schema);
            if let Some(path) = path.clone() {
                if let Some(s) = get_schema_for_path_mut(&mut state.result_set_schema, path.clone())
                {
                    // the schema of the field being unwound goes from type Array[X] to type X
                    match s {
                        Schema::Array(a) => {
                            *s = *std::mem::take(a);
                            if nullish {
                                *s = s.union(&Schema::Missing);
                            }
                        }
                        Schema::AnyOf(ao) => {
                            *s = ao
                                .iter()
                                .fold(None, |acc, x| {
                                    let schema = match x {
                                        Schema::Array(a) => *a.clone(),
                                        schema => {
                                            nullish = true;
                                            schema.clone()
                                        }
                                    };
                                    match acc {
                                        None => Some(schema),
                                        Some(acc) => Some(acc.union(&schema)),
                                    }
                                })
                                .unwrap_or(Schema::Missing);
                        }
                        _ => {}
                    };
                    if !preserve_null_and_empty_arrays {
                        schema_difference(s, set!(Schema::Missing));
                    }
                    if let Unwind::Document(d) = unwind {
                        // include_array_index will specify an output field to put the index; it can be nullish if
                        // preserve_null_and_empty_arrays is included
                        if let Some(field) = d.include_array_index.clone() {
                            let path = field
                                .split(".")
                                .map(|s| s.to_string())
                                .collect::<Vec<String>>();
                            if nullish {
                                insert_required_key_into_document(
                                    &mut state.result_set_schema,
                                    Schema::AnyOf(set!(
                                        Schema::Atomic(Atomic::Long),
                                        Schema::Atomic(Atomic::Null)
                                    )),
                                    path,
                                    true,
                                );
                            } else {
                                insert_required_key_into_document(
                                    &mut state.result_set_schema,
                                    Schema::Atomic(Atomic::Long),
                                    path,
                                    true,
                                );
                            }
                        }
                    }
                }
            }
            Ok(Schema::simplify(&state.result_set_schema.to_owned()))
        }

        /// unset_derive_schema derives schema for $unset stages. It simply removes any fields
        /// specified from the result set schema.
        fn unset_derive_schema(u: &Unset, state: &mut ResultSetState) -> Result<Schema> {
            let fields = match u {
                Unset::Single(field) => &vec![field.clone()],
                Unset::Multiple(fields) => fields,
            };
            fields.iter().for_each(|field| {
                remove_field(
                    &mut state.result_set_schema,
                    field.split('.').map(|s| s.to_string()).collect(),
                );
            });
            Ok(state.result_set_schema.to_owned())
        }

        match self {
            Stage::AddFields(a) => add_fields_derive_schema(a, state),
            Stage::AtlasSearchStage(a) => match a {
                AtlasSearchStage::Search(_) | AtlasSearchStage::VectorSearch(_) => {
                    Ok(state.result_set_schema.to_owned())
                }
                AtlasSearchStage::SearchMeta(_) => Ok(SEARCH_META.clone()),
            },
            Stage::Bucket(b) => bucket_derive_schema(b, state),
            Stage::BucketAuto(b) => bucket_auto_derive_schema(b, state),
            Stage::Collection(c) => {
                let ns = Namespace::new(c.db.clone(), c.collection.clone());
                state.result_set_schema = state
                    .catalog
                    .get(&ns)
                    .ok_or(Error::UnknownReference(ns.to_string()))?
                    .clone();
                Ok(state.result_set_schema.to_owned())
            }
            Stage::Count(c) => {
                state.result_set_schema = Schema::Document(Document {
                    keys: map! {
                        c.clone() => Schema::AnyOf(set!{Schema::Atomic(Atomic::Integer), Schema::Atomic(Atomic::Long)})
                    },
                    required: set! {c.clone()},
                    ..Default::default()
                });
                Ok(state.result_set_schema.to_owned())
            }
            Stage::Densify(d) => densify_derive_schema(d, state),
            Stage::Documents(d) => documents_derive_schema(d, state),
            e @ Stage::EquiJoin(_) => Err(Error::InvalidStage(e.clone())),
            Stage::Facet(f) => facet_derive_schema(f, state),
            Stage::Fill(f) => fill_derive_schema(f, state),
            Stage::GraphLookup(g) => graph_lookup_derive_schema(g, state),
            Stage::Group(g) => group_derive_schema(g, state),
            j @ Stage::Join(_) => Err(Error::InvalidStage(j.clone())),
            Stage::Limit(_) => Ok(state.result_set_schema.to_owned()),
            Stage::Lookup(l) => lookup_derive_schema(l, state),
            Stage::Match(ref m) => m.derive_schema(state),
            Stage::Project(p) => project_derive_schema(p, state),
            Stage::Redact(_) => todo!(),
            Stage::ReplaceWith(r) => r
                .to_owned()
                .expression()
                .derive_schema(state)
                .map(|schema| Schema::simplify(&schema)),
            Stage::Sample(_) => Ok(state.result_set_schema.to_owned()),
            Stage::SetWindowFields(s) => set_window_fields_derive_schema(s, state),
            Stage::Skip(_) | Stage::Sort(_) => Ok(state.result_set_schema.to_owned()),
            Stage::SortByCount(s) => sort_by_count_derive_schema(s.as_ref(), state),
            Stage::UnionWith(u) => union_with_derive_schema(u, state),
            Stage::Unset(u) => unset_derive_schema(u, state),
            Stage::Unwind(u) => unwind_derive_schema(u, state),
            // the following stages are not derivable, because they rely on udnerlying index information, which we do not have by
            // default given the schemas and aggregation pipelines
            Stage::GeoNear(_) => Err(Error::InvalidStage(self.clone())),
        }
    }
}

fn derive_schema_for_literal(literal_value: &LiteralValue) -> Result<Schema> {
    match literal_value {
        LiteralValue::Binary(_) => Ok(Schema::Atomic(Atomic::BinData)),
        LiteralValue::Boolean(_) => Ok(Schema::Atomic(Atomic::Boolean)),
        LiteralValue::DateTime(_) => Ok(Schema::Atomic(Atomic::Date)),
        LiteralValue::DbPointer(_) => Ok(Schema::Atomic(Atomic::DbPointer)),
        LiteralValue::Decimal128(_) => Ok(Schema::Atomic(Atomic::Decimal)),
        LiteralValue::Double(_) => Ok(Schema::Atomic(Atomic::Double)),
        LiteralValue::Int32(_) => Ok(Schema::Atomic(Atomic::Integer)),
        LiteralValue::Int64(_) => Ok(Schema::Atomic(Atomic::Long)),
        LiteralValue::JavaScriptCode(_) => Ok(Schema::Atomic(Atomic::Javascript)),
        LiteralValue::JavaScriptCodeWithScope(_) => Ok(Schema::Atomic(Atomic::JavascriptWithScope)),
        LiteralValue::MaxKey => Ok(Schema::Atomic(Atomic::MaxKey)),
        LiteralValue::MinKey => Ok(Schema::Atomic(Atomic::MinKey)),
        LiteralValue::Null => Ok(Schema::Atomic(Atomic::Null)),
        LiteralValue::ObjectId(_) => Ok(Schema::Atomic(Atomic::ObjectId)),
        LiteralValue::RegularExpression(_) => Ok(Schema::Atomic(Atomic::Regex)),
        LiteralValue::String(_) => Ok(Schema::Atomic(Atomic::String)),
        LiteralValue::Symbol(_) => Ok(Schema::Atomic(Atomic::Symbol)),
        LiteralValue::Timestamp(_) => Ok(Schema::Atomic(Atomic::Timestamp)),
        LiteralValue::Undefined => Ok(Schema::Atomic(Atomic::Undefined)),
    }
}

impl DeriveSchema for Expression {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        state.result_set_schema = promote_missing(&state.result_set_schema);
        match self {
            Expression::Array(ref a) => {
                let array_schema = a
                    .iter()
                    .map(|e| {
                        e.derive_schema(state)
                            .map(|schema| schema.upconvert_missing_to_null())
                    })
                    .collect::<Result<BTreeSet<_>>>()?;
                let array_schema = match array_schema.len() {
                    0 => Schema::Unsat,
                    1 => array_schema.into_iter().next().unwrap(),
                    _ => Schema::AnyOf(array_schema),
                };
                Ok(Schema::Array(Box::new(array_schema)))
            }
            Expression::Document(d) => {
                let (mut keys, mut required) = (BTreeMap::new(), BTreeSet::new());
                for (key, e) in d.iter() {
                    let key_schema = e.derive_schema(state)?;
                    match key_schema.satisfies(&Schema::Missing) {
                        Satisfaction::Not => {
                            required.insert(key.clone());
                            keys.insert(key.clone(), key_schema);
                        }
                        Satisfaction::May => {
                            keys.insert(key.clone(), key_schema);
                        }
                        Satisfaction::Must => (),
                    }
                }
                Ok(Schema::Document(Document {
                    keys,
                    required,
                    ..Default::default()
                }))
            }
            Expression::Literal(ref l) => derive_schema_for_literal(l),
            Expression::Ref(Ref::FieldRef(f)) => {
                let path = f.split(".").map(|s| s.to_string()).collect::<Vec<String>>();
                // If the user has rebound the CURRENT variable, we should use that schema instead of the result set schema to find any
                // path.
                let current_schema = state
                    .variables
                    .get_mut("CURRENT")
                    .unwrap_or(&mut state.result_set_schema);
                // if we have an Any schema, just short circuit and return any
                if get_schema_for_path_mut(current_schema, vec![path[0].clone()])
                    == Some(&mut Schema::Any)
                {
                    return Ok(Schema::Any);
                }
                let schema = get_schema_for_path(current_schema.clone(), path);
                match schema {
                    Some(schema) => Ok(schema.clone()),
                    // Unknown fields actually have the Schema Missing, while unknown variables are
                    // an error.
                    None => Ok(Schema::Missing),
                }
            }
            Expression::Ref(Ref::VariableRef(v)) => {
                match v.as_str() {
                    "NOW" => Ok(Schema::Atomic(Atomic::Date)),
                    "CURRENT_TIME" => Ok(Schema::Atomic(Atomic::Timestamp)),
                    "REMOVE" => Ok(Schema::Missing),
                    "ROOT" => Ok(state.result_set_schema.clone()),
                    "USER_ROLES" => Ok(Schema::Array(Box::new(Schema::Document(Document {
                        keys: map! {
                            "_id".to_string() => Schema::Atomic(Atomic::String),
                            "role".to_string() => Schema::Atomic(Atomic::String),
                            "db".to_string() => Schema::Atomic(Atomic::String)
                        },
                        required: set!["_id".to_string(), "role".to_string(), "db".to_string()],
                        ..Default::default()
                    })))),
                    "SEARCH_META" => Ok(SEARCH_META.clone()),
                    v => {
                        let path: Vec<String> = v.split('.').map(|s| s.to_string()).collect();
                        let var_schema = if v.contains(".") {
                            state.variables.get(&path[0]).and_then(|doc| {
                                get_schema_for_path(doc.clone(), path[1..].to_vec())
                            })
                        } else {
                            state.variables.get(v).cloned()
                        };
                        match var_schema {
                            Some(schema) => Ok(schema.clone()),
                            None => {
                                if path[0] == "CURRENT" {
                                    // CURRENT is equivalent to ROOT, if it has not been rebound
                                    // The reason we do this is because a field reference
                                    // `$<field>` is equivalent to `$$CURRENT.<field>.
                                    Ok(get_schema_for_path_mut(
                                        &mut state.result_set_schema,
                                        path[1..].to_vec(),
                                    )
                                    .unwrap()
                                    .clone())
                                // if the top level key is present but the full path is not, agg
                                // treats it as an empty document
                                } else if state.variables.contains_key(&path[0]) {
                                    Ok(Schema::Document(Document::empty()))
                                } else {
                                    Err(Error::UnknownReference(v.into()))
                                }
                            }
                        }
                    }
                }
            }
            Expression::TaggedOperator(op) => op.derive_schema(state),
            Expression::UntaggedOperator(op) => op.derive_schema(state),
        }
    }
}

/// This helper gets the maximal satisfaction of a list of expressions for a given type. This is primarily useful
/// for determining if _any_ argument must be null, or may be null, which can determine the output of an operator.
/// This has similar implications for Decimal, which affects many math ops.
fn arguments_schema_satisfies(
    args: &[&Expression],
    state: &mut ResultSetState,
    schema: &Schema,
) -> Result<Satisfaction> {
    let mut satisfaction = Satisfaction::Not;
    for arg in args.iter() {
        let arg_schema = arg.derive_schema(state)?.upconvert_missing_to_null();
        match (arg_schema.satisfies(schema), satisfaction) {
            (Satisfaction::May, Satisfaction::Not) => {
                satisfaction = Satisfaction::May;
            }
            (Satisfaction::Must, _) => {
                satisfaction = Satisfaction::Must;
            }
            _ => {}
        };
    }
    Ok(satisfaction)
}

/// handle_null_satisfaction captures the behavior of operators that are nullable if any of the arguments are null
/// by checking the nullability of an input schema, and applying that to the default schema.
fn handle_null_satisfaction(
    args: Vec<&Expression>,
    state: &mut ResultSetState,
    non_null_type: Schema,
) -> Result<Schema> {
    match arguments_schema_satisfies(&args, state, &NULLISH)? {
        Satisfaction::Not => Ok(non_null_type),
        Satisfaction::May => Ok(Schema::simplify(&Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Null),
            non_null_type
        )))),
        Satisfaction::Must => Ok(Schema::Atomic(Atomic::Null)),
    }
}

impl DeriveSchema for TaggedOperator {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        /// derive_window_func is a macro helper for getting the schema for $derivative and $expMovingAverage operators
        /// within $setWindowFunc stages. They are numeric with certain constraints about what types are returned based
        /// on the input.
        macro_rules! derive_window_func {
            ($input:expr) => {{
                let input_schema = $input.input.derive_schema(state)?;
                let mut types: BTreeSet<Schema> = set!(Schema::Atomic(Atomic::Null));
                if input_schema.satisfies(&Schema::Atomic(Atomic::Decimal)) != Satisfaction::Not {
                    types.insert(Schema::Atomic(Atomic::Decimal));
                }
                if input_schema.satisfies(&Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                ))) != Satisfaction::Not
                {
                    types.insert(Schema::Atomic(Atomic::Double));
                }
                Ok(Schema::simplify(&Schema::AnyOf(types)))
            }};
        }
        /// derive_date_addition is a macro helper for deriving the schema for the $dateAdd and $dateSubtract operators.
        /// They return date (possibly null)
        macro_rules! derive_date_addition {
            ($input:expr) => {{
                let args = vec![
                    $input.amount.as_ref(),
                    $input.start_date.as_ref(),
                    $input
                        .timezone
                        .as_ref()
                        .map_or(&Expression::Literal(LiteralValue::Boolean(true)), |x| {
                            x.as_ref()
                        }),
                    $input.unit.as_ref(),
                ];
                handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Date))
            }};
        }
        /// optional arg or truish is a macro used to help with determining nullabililty around
        /// optional arguments. It returns the value if it exists, otherwise returns true, ie, a
        /// value that does not evaluate to false / null / 0
        macro_rules! optional_arg_or_truish {
            ($input:expr) => {{
                $input
                    .as_ref()
                    .map_or(&Expression::Literal(LiteralValue::Boolean(true)), |x| {
                        x.as_ref()
                    })
            }};
        }
        match self {
            TaggedOperator::Convert(c) => {
                let mut convert_type = match c.to.as_ref() {
                    Expression::Literal(LiteralValue::String(s)) => schema_for_type_str(s.as_str()),
                    Expression::Literal(LiteralValue::Double(d)) => {
                        schema_for_type_numeric(*d as i32)
                    }
                    Expression::Literal(LiteralValue::Int32(i)) => schema_for_type_numeric(*i),
                    Expression::Literal(LiteralValue::Int64(i)) => {
                        schema_for_type_numeric(*i as i32)
                    }
                    Expression::Literal(LiteralValue::Decimal128(d)) => {
                        let decimal_string = d.to_string();
                        let decimal_as_double = decimal_string
                            .parse::<f64>()
                            .map_err(|_| Error::InvalidConvertTypeValue(decimal_string))?;
                        schema_for_type_numeric(decimal_as_double as i32)
                    }
                    // unfortunately, convert can take any expression as a to type. We use
                    // the full set of to types when we cant statically determine the output
                    _ => Schema::AnyOf(set!(
                        Schema::Atomic(Atomic::Integer),
                        Schema::Atomic(Atomic::Double),
                        Schema::Atomic(Atomic::Long),
                        Schema::Atomic(Atomic::Decimal),
                        Schema::Atomic(Atomic::Boolean),
                        Schema::Atomic(Atomic::Date),
                        Schema::Atomic(Atomic::Timestamp),
                        Schema::Atomic(Atomic::BinData),
                        Schema::Atomic(Atomic::String),
                    )),
                };
                if let Some(on_null) = c.on_null.as_ref() {
                    convert_type = convert_type.subtract_nullish();
                    convert_type = convert_type.union(&on_null.derive_schema(state)?);
                }
                if let Some(on_error) = c.on_error.as_ref() {
                    convert_type = convert_type.union(&on_error.derive_schema(state)?);
                }
                Ok(convert_type)
            }
            TaggedOperator::DenseRank(_)
            | TaggedOperator::DocumentNumber(_)
            | TaggedOperator::Rank(_) => Ok(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long)
            ))),
            TaggedOperator::Derivative(d) => derive_window_func!(d),
            TaggedOperator::ExpMovingAvg(e) => derive_window_func!(e),
            TaggedOperator::Median(m) => handle_null_satisfaction(
                vec![m.input.as_ref()],
                state,
                Schema::Atomic(Atomic::Double),
            ),
            TaggedOperator::Percentile(p) => Ok(Schema::Array(Box::new(handle_null_satisfaction(
                vec![p.input.as_ref()],
                state,
                Schema::Atomic(Atomic::Double),
            )?))),
            TaggedOperator::RegexFind(_) => Ok(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Null),
                Schema::Document(Document {
                    keys: map! {
                        "match".to_string() => Schema::Atomic(Atomic::String),
                        "idx".to_string() => Schema::Atomic(Atomic::Integer),
                        "captures".to_string() => Schema::Array(Box::new(Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)))))
                    },
                    required: set!(),
                    ..Default::default()
                })
            ))),
            TaggedOperator::RegexFindAll(_) => {
                Ok(Schema::Array(Box::new(Schema::Document(Document {
                    keys: map! {
                        "match".to_string() => Schema::Atomic(Atomic::String),
                        "idx".to_string() => Schema::Atomic(Atomic::Integer),
                        "captures".to_string() => Schema::Array(Box::new(Schema::AnyOf(set!(Schema::Atomic(Atomic::String), Schema::Atomic(Atomic::Null)))))
                    },
                    required: set!(),
                    ..Default::default()
                }))))
            }
            TaggedOperator::LTrim(t) | TaggedOperator::RTrim(t) | TaggedOperator::Trim(t) => {
                handle_null_satisfaction(
                    vec![t.input.as_ref(), optional_arg_or_truish!(t.chars)],
                    state,
                    Schema::Atomic(Atomic::String),
                )
            }
            TaggedOperator::DayOfWeek(d)
            | TaggedOperator::DayOfMonth(d)
            | TaggedOperator::DayOfYear(d)
            | TaggedOperator::IsoDayOfWeek(d)
            | TaggedOperator::IsoWeek(d)
            | TaggedOperator::IsoWeekYear(d)
            | TaggedOperator::Week(d)
            | TaggedOperator::Month(d)
            | TaggedOperator::Year(d)
            | TaggedOperator::Hour(d)
            | TaggedOperator::Minute(d)
            | TaggedOperator::Second(d)
            | TaggedOperator::Millisecond(d) => {
                let date = match d.date.as_ref() {
                    Expression::Array(a) => &a[0],
                    expr => expr,
                };
                handle_null_satisfaction(
                    vec![date, optional_arg_or_truish!(d.timezone)],
                    state,
                    Schema::Atomic(Atomic::Integer),
                )
            }
            TaggedOperator::DateFromParts(d) => {
                let args = vec![
                    optional_arg_or_truish!(d.year),
                    optional_arg_or_truish!(d.month),
                    optional_arg_or_truish!(d.day),
                    optional_arg_or_truish!(d.hour),
                    optional_arg_or_truish!(d.minute),
                    optional_arg_or_truish!(d.second),
                    optional_arg_or_truish!(d.millisecond),
                    optional_arg_or_truish!(d.iso_day_of_week),
                    optional_arg_or_truish!(d.iso_week),
                    optional_arg_or_truish!(d.iso_week_year),
                    optional_arg_or_truish!(d.timezone),
                ];
                handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Date))
            }
            // $dateFromString returns a Date, however we also union on any types implied by onNull or onError
            TaggedOperator::DateFromString(d) => {
                let nullable_args = vec![
                    d.date_string.as_ref(),
                    optional_arg_or_truish!(d.format),
                    optional_arg_or_truish!(d.timezone),
                ];
                let on_null_schema = d
                    .on_null
                    .as_ref()
                    .map(|x| x.derive_schema(state))
                    .unwrap_or(Ok(Schema::Atomic(Atomic::Null)))?;
                let mut types: BTreeSet<Schema> = set!(Schema::Atomic(Atomic::Date));
                match arguments_schema_satisfies(
                    &nullable_args,
                    state,
                    &Schema::Atomic(Atomic::Null),
                )? {
                    Satisfaction::Must => {
                        return Ok(on_null_schema);
                    }
                    Satisfaction::May => {
                        types.insert(on_null_schema);
                    }
                    _ => {}
                };
                if let Some(error_schema) = d.on_error.as_ref().map(|x| x.derive_schema(state)) {
                    types.insert(error_schema?);
                }
                Ok(Schema::simplify(&Schema::AnyOf(types)))
            }
            // $dateToParts returns a document with a fixed schema, however, the fixed schema is different
            // depending on whether iso8601 is true or false
            TaggedOperator::DateToParts(d) => {
                let args = vec![d.date.as_ref(), optional_arg_or_truish!(d.timezone)];
                match d.iso8601 {
                    Some(true) => handle_null_satisfaction(
                        args,
                        state,
                        Schema::Document(Document {
                            keys: map! {
                                "isoWeekYear".to_string() => Schema::Atomic(Atomic::Integer),
                                "isoWeek".to_string() => Schema::Atomic(Atomic::Integer),
                                "isoDayOfWeek".to_string() => Schema::Atomic(Atomic::Integer),
                                "hour".to_string() => Schema::Atomic(Atomic::Integer),
                                "minute".to_string() => Schema::Atomic(Atomic::Integer),
                                "second".to_string() => Schema::Atomic(Atomic::Integer),
                                "millisecond".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            required: set!(
                                "isoWeekYear".to_string(),
                                "isoWeek".to_string(),
                                "isoDayOfWeek".to_string(),
                                "hour".to_string(),
                                "minute".to_string(),
                                "second".to_string(),
                                "millisecond".to_string()
                            ),
                            ..Default::default()
                        }),
                    ),
                    _ => handle_null_satisfaction(
                        args,
                        state,
                        Schema::Document(Document {
                            keys: map! {
                                "year".to_string() => Schema::Atomic(Atomic::Integer),
                                "month".to_string() => Schema::Atomic(Atomic::Integer),
                                "day".to_string() => Schema::Atomic(Atomic::Integer),
                                "hour".to_string() => Schema::Atomic(Atomic::Integer),
                                "minute".to_string() => Schema::Atomic(Atomic::Integer),
                                "second".to_string() => Schema::Atomic(Atomic::Integer),
                                "millisecond".to_string() => Schema::Atomic(Atomic::Integer),
                            },
                            required: set!(
                                "year".to_string(),
                                "month".to_string(),
                                "day".to_string(),
                                "hour".to_string(),
                                "minute".to_string(),
                                "second".to_string(),
                                "millisecond".to_string()
                            ),
                            ..Default::default()
                        }),
                    ),
                }
            }
            // $dateToString produces a string, however we must also add null if possible, as well
            // as unioning the onNull expression's schema if provided
            TaggedOperator::DateToString(d) => {
                let nullable_args = vec![
                    d.date.as_ref(),
                    optional_arg_or_truish!(d.format),
                    optional_arg_or_truish!(d.timezone),
                ];
                let on_null_schema = d
                    .on_null
                    .as_ref()
                    .map(|x| x.derive_schema(state))
                    .unwrap_or(Ok(Schema::Atomic(Atomic::Null)))?;
                let mut types: BTreeSet<Schema> = set!(Schema::Atomic(Atomic::String));
                match arguments_schema_satisfies(
                    &nullable_args,
                    state,
                    &Schema::Atomic(Atomic::Null),
                )? {
                    Satisfaction::Must => {
                        return Ok(on_null_schema);
                    }
                    Satisfaction::May => {
                        types.insert(on_null_schema);
                    }
                    _ => {}
                };
                Ok(Schema::simplify(&Schema::AnyOf(types)))
            }
            TaggedOperator::DateAdd(d) => derive_date_addition!(d),
            TaggedOperator::DateSubtract(d) => derive_date_addition!(d),
            TaggedOperator::DateDiff(d) => {
                let args = vec![
                    d.start_date.as_ref(),
                    d.end_date.as_ref(),
                    optional_arg_or_truish!(d.timezone),
                    d.unit.as_ref(),
                    optional_arg_or_truish!(d.start_of_week),
                ];
                handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Long))
            }
            TaggedOperator::DateTrunc(d) => {
                let args = vec![
                    d.date.as_ref(),
                    d.unit.as_ref(),
                    optional_arg_or_truish!(d.timezone),
                    optional_arg_or_truish!(d.bin_size),
                    optional_arg_or_truish!(d.start_of_week),
                ];
                handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Date))
            }
            TaggedOperator::SortArray(s) => s.input.derive_schema(state),
            TaggedOperator::Let(l) => {
                // we create a copy of the underlying result set state, then add the vars to that.
                // this allows us to temporarily overwrite any variables from the top level if they are defined in
                // both places, and result set state remains unchanged for future operations.
                let mut variables = state.variables.clone();
                let mut let_state_variables = l
                    .vars
                    .iter()
                    .map(|(key, value)| {
                        value
                            .derive_schema(state)
                            .map(|schema| (key.clone(), schema))
                    })
                    .collect::<Result<BTreeMap<String, Schema>>>()?;
                variables.append(&mut let_state_variables);
                let mut let_state = ResultSetState {
                    result_set_schema: state.result_set_schema.clone(),
                    catalog: state.catalog,
                    variables,
                    current_db: state.current_db.clone(),
                    null_behavior: state.null_behavior,
                    accumulator_stage: state.accumulator_stage,
                };
                l.inside.derive_schema(&mut let_state)
            }
            TaggedOperator::GetField(g) => {
                let input_schema = g.input.derive_schema(state)?;
                let field_schema = get_schema_for_path(input_schema, vec![g.field.clone()]);
                match field_schema {
                    None => Ok(Schema::Missing),
                    Some(schema) => Ok(schema.clone()),
                }
            }
            TaggedOperator::SetField(s) => {
                // set field does not update the underlying result set schema, but rather, gets the schema
                // of the input, modifies that, and returns it. Thus, we copy the input schema and modify that accordingly.
                let mut input_schema = s.input.derive_schema(state)?;
                let value_schema = s.value.derive_schema(state)?;
                let field_schema =
                    get_schema_for_path_mut(&mut input_schema, vec![s.field.clone()]);
                match field_schema {
                    // if we are setting a new field, add it in appropriately, unless its missing (no-op)
                    None => {
                        if value_schema != Schema::Missing {
                            if let Schema::Document(_) = input_schema {
                                insert_required_key_into_document(
                                    &mut input_schema,
                                    value_schema.clone(),
                                    vec![s.field.clone()],
                                    true,
                                );
                                Ok(Schema::simplify(&input_schema))
                            } else {
                                let new_field = Schema::Document(Document {
                                    keys: map! {
                                        s.field.clone() => value_schema,
                                    },
                                    required: set!(s.field.clone()),
                                    ..Default::default()
                                });
                                Ok(Schema::simplify(&input_schema.union(&new_field)))
                            }
                        } else {
                            Ok(Schema::simplify(&input_schema))
                        }
                    }
                    // if we are handling a new field, check first if the schema is missing (could either be
                    // cause by setting to missing, or setting ot $$REMOVE). Remove it or set it to the new type accordingly
                    Some(field_schema) => {
                        match value_schema {
                            Schema::Missing => {
                                remove_field(&mut input_schema, vec![s.field.clone()]);
                            }
                            _ => {
                                *field_schema = value_schema;
                            }
                        }
                        Ok(Schema::simplify(&input_schema))
                    }
                }
            }
            TaggedOperator::UnsetField(u) => {
                // note: this is functionally the same as $setField with Schema::Missing or $$REMOVE
                let mut input_schema = u.input.derive_schema(state)?;
                remove_field(&mut input_schema, vec![u.field.clone()]);
                Ok(input_schema)
            }
            TaggedOperator::Bottom(b) => b.output.derive_schema(state),
            TaggedOperator::BottomN(b) => {
                Ok(Schema::Array(Box::new(b.output.derive_schema(state)?)))
            }
            TaggedOperator::Cond(c) => {
                let then_schema = c.then.derive_schema(state)?;
                let else_schema = c.r#else.derive_schema(state)?;
                Ok(Schema::simplify(&Schema::AnyOf(
                    set! {then_schema, else_schema},
                )))
            }
            // $filter can actually _constrain_ the result set schema, so we add the relevant variables to a
            // copy of the state and call match derive schema to get a constrained schema based on the cond
            TaggedOperator::Filter(f) => {
                let input_schema = f.input.derive_schema(state)?;
                let is_nullable = input_schema.satisfies(&NULLISH.clone()) != Satisfaction::Not;
                let array_element_schema = array_element_schema_or_error!(input_schema, f.input);
                let var_name = f._as.clone().unwrap_or("this".to_string());
                let mut temp_state = state.clone();
                temp_state
                    .variables
                    .insert(var_name.clone(), array_element_schema);
                f.cond.match_derive_schema(&mut temp_state)?;
                let filter_schema = temp_state.variables.remove(&var_name);
                if let Some(schema) = filter_schema {
                    let mut schema = Schema::Array(Box::new(schema));
                    if is_nullable {
                        schema = schema.union(&Schema::Atomic(Atomic::Null))
                    }
                    Ok(schema)
                // this should be unreachable, since we manually add and remove the variable from the state
                } else {
                    unreachable!()
                }
            }
            TaggedOperator::FirstN(n) => Ok(Schema::Array(Box::new(n.input.derive_schema(state)?))),
            TaggedOperator::Integral(i) => {
                get_decimal_double_or_nullish(vec![i.input.as_ref()], state)
            }
            TaggedOperator::LastN(n) => Ok(Schema::Array(Box::new(n.input.derive_schema(state)?))),
            // $map sets up a copy of the state and evaluates the in expression based on the schema of the input array.
            TaggedOperator::Map(m) => {
                let var = m._as.clone();
                let var = var.unwrap_or("this".to_string());
                let input_schema = m.input.derive_schema(state)?;
                if input_schema.satisfies(&NULLISH.clone()) == Satisfaction::Must {
                    return Ok(Schema::Atomic(Atomic::Null));
                }
                let array_element_schema =
                    array_element_schema_or_error!(input_schema.clone(), m.input);
                let mut new_state = state.clone();
                let mut variables = state.variables.clone();
                variables.insert(var, array_element_schema);
                new_state.variables = variables;
                let not_null_schema = Schema::Array(Box::new(
                    m.inside
                        .derive_schema(&mut new_state)?
                        .upconvert_missing_to_null(),
                ));
                Ok(
                    if input_schema.satisfies(&NULLISH.clone()) != Satisfaction::Not {
                        not_null_schema.union(&Schema::Atomic(Atomic::Null))
                    } else {
                        not_null_schema
                    },
                )
            }
            // Unfortunately, unlike $max and $min, $maxN and $minN cannot
            // reduce the scope of the result Schema beyond Array(InputSchema)
            // because doing so would require knowing all the data.
            TaggedOperator::MaxNArrayElement(m) | TaggedOperator::MinNArrayElement(m) => {
                let input_schema = m.input.derive_schema(state)?;
                if state.accumulator_stage {
                    Ok(Schema::Array(Box::new(input_schema)))
                } else {
                    Ok(input_schema)
                }
            }
            // $reduce sets up a copy of the state and evaluates the in expression based on the schema of the input array.
            TaggedOperator::Reduce(r) => {
                let input_schema = r.input.derive_schema(state)?;
                if input_schema.satisfies(&NULLISH.clone()) == Satisfaction::Must {
                    return Ok(Schema::Atomic(Atomic::Null));
                }
                let array_element_schema =
                    array_element_schema_or_error!(input_schema.clone(), r.input);
                let initial_schema = r.initial_value.derive_schema(state)?;
                let mut new_state = state.clone();
                let mut variables = state.variables.clone();
                variables.insert("this".to_string(), array_element_schema);
                variables.insert("value".to_string(), initial_schema);
                new_state.variables = variables;
                if input_schema.satisfies(&NULLISH.clone()) == Satisfaction::Not {
                    r.inside.derive_schema(&mut new_state)
                } else {
                    Ok(r.inside
                        .derive_schema(&mut new_state)?
                        .union(&Schema::Atomic(Atomic::Null)))
                }
            }
            TaggedOperator::Regex(_) => Ok(Schema::Atomic(Atomic::Integer)),
            TaggedOperator::ReplaceAll(r) | TaggedOperator::ReplaceOne(r) => {
                handle_null_satisfaction(
                    vec![r.input.as_ref(), r.find.as_ref(), r.replacement.as_ref()],
                    state,
                    Schema::Atomic(Atomic::String),
                )
            }
            TaggedOperator::Shift(s) => {
                let mut output_schema = s.output.derive_schema(state)?;
                output_schema = match s.default.as_ref() {
                    Some(default) => output_schema.union(&default.derive_schema(state)?),
                    None => output_schema.union(&Schema::Atomic(Atomic::Null)),
                };
                Ok(output_schema)
            }
            TaggedOperator::Switch(s) => {
                let mut schema = s.default.derive_schema(state)?;
                for branch in s.branches.iter() {
                    schema = schema.union(&branch.then.derive_schema(state)?);
                }
                Ok(schema)
            }
            TaggedOperator::Top(t) => t.output.derive_schema(state),
            TaggedOperator::TopN(t) => Ok(Schema::Array(Box::new(t.output.derive_schema(state)?))),
            // $zip unions the schema for all of the inputs as well as the defaults, if provided
            TaggedOperator::Zip(z) => {
                let inputs = match z.inputs.as_ref() {
                    Expression::Array(a) => a,
                    exp => {
                        return Err(Error::InvalidExpressionForField(
                            format!("{:?}", exp),
                            "inputs",
                        ))
                    }
                };
                let mut array_schema = Schema::Unsat;
                for input in inputs.iter() {
                    let input_schema = input.derive_schema(state)?;
                    array_schema = array_schema.union(&input_schema);
                }
                if let Some(defaults) = z.defaults.as_ref() {
                    let defaults_schema = defaults.derive_schema(state)?;
                    if matches!(defaults_schema, Schema::Array(_)) {
                        array_schema = array_schema.union(&defaults_schema);
                    }
                }
                Ok(Schema::Array(Box::new(array_schema)))
            }
            TaggedOperator::SqlAvg(s) => UntaggedOperator {
                op: UntaggedOperatorName::Avg,
                args: vec![(*s.var).clone()],
            }
            .derive_schema(state),
            TaggedOperator::SqlCount(s) => UntaggedOperator {
                op: UntaggedOperatorName::Count,
                args: vec![(*s.var).clone()],
            }
            .derive_schema(state),
            TaggedOperator::SqlFirst(s) => UntaggedOperator {
                op: UntaggedOperatorName::First,
                args: vec![(*s.var).clone()],
            }
            .derive_schema(state),
            TaggedOperator::SqlLast(s) => UntaggedOperator {
                op: UntaggedOperatorName::Last,
                args: vec![(*s.var).clone()],
            }
            .derive_schema(state),
            TaggedOperator::SqlMax(s) => UntaggedOperator {
                op: UntaggedOperatorName::Max,
                args: vec![(*s.var).clone()],
            }
            .derive_schema(state),
            TaggedOperator::SqlMergeObjects(s) => UntaggedOperator {
                op: UntaggedOperatorName::MergeObjects,
                args: vec![(*s.var).clone()],
            }
            .derive_schema(state),
            TaggedOperator::SqlMin(s) => UntaggedOperator {
                op: UntaggedOperatorName::Min,
                args: vec![(*s.var).clone()],
            }
            .derive_schema(state),
            TaggedOperator::SqlStdDevPop(s) => UntaggedOperator {
                op: UntaggedOperatorName::StdDevPop,
                args: vec![(*s.var).clone()],
            }
            .derive_schema(state),
            TaggedOperator::SqlStdDevSamp(s) => UntaggedOperator {
                op: UntaggedOperatorName::StdDevSamp,
                args: vec![(*s.var).clone()],
            }
            .derive_schema(state),
            TaggedOperator::SqlSum(s) => UntaggedOperator {
                op: UntaggedOperatorName::Sum,
                args: vec![(*s.var).clone()],
            }
            .derive_schema(state),
            TaggedOperator::Accumulator(_)
            | TaggedOperator::Function(_)
            | TaggedOperator::Like(_)
            | TaggedOperator::SqlConvert(_)
            | TaggedOperator::SqlDivide(_)
            | TaggedOperator::Subquery(_)
            | TaggedOperator::SubqueryComparison(_)
            | TaggedOperator::SubqueryExists(_) => Err(Error::InvalidTaggedOperator(self.clone())),
        }
    }
}

/// get_input_schema will take in the arguments of an untagged operator and return all possible types
/// of any input. there are two primary uses for this -- first, it allows us to easily determine if any
/// argument is null or nullable without inspecting the whole list. It similarly allows us to work with
/// numerics more easily, where (amongst other things) Decimal is often handled differently.
fn get_input_schema(args: &[&Expression], state: &mut ResultSetState) -> Result<Schema> {
    let x = args
        .iter()
        .map(|e| {
            e.derive_schema(state)
                .map(|schema| schema.upconvert_missing_to_null())
        })
        .collect::<Result<BTreeSet<_>>>()?;
    Ok(Schema::simplify(&Schema::AnyOf(x)))
}

/// get_decimal_double_or_nullish_from_schema is a helper for some numeric operators that return
/// decimal if any arg is a decimal, or double otherwise.
fn get_decimal_double_or_nullish_from_schema(schema: Schema) -> Schema {
    use Satisfaction::*;
    let numeric_satisfaction = schema.satisfies(&NUMERIC);
    if numeric_satisfaction == Not {
        return Schema::Atomic(Atomic::Null);
    }
    let numeric_schema = match schema.satisfies(&Schema::Atomic(Atomic::Decimal)) {
        Must => Schema::Atomic(Atomic::Decimal),
        May => Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal)
        )),
        Not => Schema::Atomic(Atomic::Double),
    };
    if numeric_satisfaction == Must {
        numeric_schema
    } else {
        numeric_schema.union(&Schema::Atomic(Atomic::Null))
    }
}

/// get_decimal_double_or_nullish handles one of the most common cases for math untagged operators,
/// which is that if any of the inputs is decimal, the operator returns a decimal; if there are any
/// other numeric types, they will return a double; and the operator is nullable, so must handle Null satisfaction.
fn get_decimal_double_or_nullish(
    args: Vec<&Expression>,
    state: &mut ResultSetState,
) -> Result<Schema> {
    let decimal_satisfaction =
        arguments_schema_satisfies(&args, state, &Schema::Atomic(Atomic::Decimal))?;
    let numeric_satisfaction = arguments_schema_satisfies(
        &args,
        state,
        &Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Integer),
            Schema::Atomic(Atomic::Long),
        )),
    )?;
    let schema = match (decimal_satisfaction, numeric_satisfaction) {
        (Satisfaction::Must, _) | (Satisfaction::May, Satisfaction::Not) => {
            Schema::Atomic(Atomic::Decimal)
        }
        (_, Satisfaction::Must) | (Satisfaction::Not, Satisfaction::May) => {
            Schema::Atomic(Atomic::Double)
        }
        (Satisfaction::May, Satisfaction::May) => Schema::AnyOf(set!(
            Schema::Atomic(Atomic::Double),
            Schema::Atomic(Atomic::Decimal)
        )),
        _ => Schema::Atomic(Atomic::Null),
    };
    handle_null_satisfaction(args, state, schema)
}

impl DeriveSchema for UntaggedOperator {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        /// get_sum_type is a helper for getting the output schema of a $sum operator. It filters
        /// out non-numeric types (the operator ignores these also) and then calculates return type
        /// based on max numeric type (+ overflow type for that numeric type)
        fn get_sum_type(s: Schema) -> Schema {
            // get the maximum numeric type
            let s = if let Schema::AnyOf(a) = s {
                a.into_iter()
                    .filter(|s| {
                        matches!(
                            s,
                            Schema::Atomic(Atomic::Integer)
                                | Schema::Atomic(Atomic::Long)
                                | Schema::Atomic(Atomic::Double)
                                | Schema::Atomic(Atomic::Decimal)
                        )
                    })
                    .max()
                    .unwrap_or(Schema::Atomic(Atomic::Integer))
            } else {
                s
            };

            // now use the maximum numeric type to determine the return type based on possible
            // overflow (int -> long -> double -> decimal), that is every type generates an AnyOf
            // of itself plus the larger types
            match s {
                Schema::Atomic(Atomic::Integer) => NUMERIC.clone(),
                Schema::Atomic(Atomic::Long) => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Long),
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal)
                )),
                Schema::Atomic(Atomic::Double) => Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Decimal)
                )),
                Schema::Atomic(Atomic::Decimal) => Schema::Atomic(Atomic::Decimal),
                // Sum returns 0 as an integer, if there are no numeric values to sum
                _ => Schema::Atomic(Atomic::Integer),
            }
        }

        fn get_numeric_operator_schema(
            args: &[&Expression],
            state: &mut ResultSetState,
        ) -> Result<Schema> {
            let mut type_set: BTreeSet<Schema> = set!();
            let mut arg_schemas: BTreeSet<Schema> = set!();
            // here we derive schema for all args and filter for numeric types only.
            // we also use this time to capture if any args MAY be nullable, and to
            // return null if any arg MUST be null or missing
            for arg in args.iter() {
                let arg_schema = arg.derive_schema(state)?;
                match arg_schema.satisfies(&NULLISH.clone()) {
                    Satisfaction::Must => return Ok(Schema::Atomic(Atomic::Null)),
                    Satisfaction::May => {
                        type_set.insert(Schema::Atomic(Atomic::Null));
                    }
                    Satisfaction::Not => {}
                }
                arg_schemas.insert(arg_schema.intersection(&NUMERIC.clone()));
            }
            // we can set the lower bound of numeric args as the largest type that is the minimum
            // of one of the arguments. For example, if the lowest ordered type an argument can take
            // on is a Double, the set that as the lower bound, as the operator will never return int or long.
            let lower_bound = arg_schemas
                .iter()
                .map(|schema| match schema {
                    Schema::Atomic(_) => schema,
                    Schema::AnyOf(ao) => ao.iter().min().unwrap(),
                    _ => &Schema::Unsat,
                })
                .max()
                .unwrap()
                .to_owned();
            type_set.insert(lower_bound.clone());
            // we then add any possible types that are > the lower bound to the set of types
            // our operator can return, and return the anyof of all the possible types.
            for arg_schema in arg_schemas {
                match arg_schema {
                    Schema::Atomic(_) => {
                        if arg_schema > lower_bound {
                            type_set.insert(arg_schema);
                        }
                    }
                    Schema::AnyOf(ao) => {
                        ao.into_iter().for_each(|schema| {
                            if schema > lower_bound {
                                type_set.insert(schema);
                            }
                        });
                    }
                    _ => {}
                }
            }
            let mut schema = Schema::simplify(&Schema::AnyOf(type_set));
            // the operators calling this, like multiply and add, will return a long if the
            // inputs are integers and the value is sufficiently large, so we add that type
            if schema.satisfies(&Schema::Atomic(Atomic::Long)) != Satisfaction::Not
                && schema != Schema::Unsat
            {
                schema = schema.union(&Schema::Atomic(Atomic::Double));
            }
            if schema.satisfies(&Schema::Atomic(Atomic::Integer)) != Satisfaction::Not
                && schema != Schema::Unsat
            {
                schema = schema.union(&Schema::Atomic(Atomic::Long));
            }
            Ok(schema)
        }

        let mut args = self.args.iter().collect();
        match self.op {
            // no-ops
            UntaggedOperatorName::Abs | UntaggedOperatorName::Ceil | UntaggedOperatorName::Floor | UntaggedOperatorName::ReverseArray | UntaggedOperatorName::Round | UntaggedOperatorName::SampleRate | UntaggedOperatorName::Slice
            | UntaggedOperatorName::Trunc
            // We cannot know anything about the Schema change from the set difference, since it only
            // removes values not types. The best we can do is keep the lhs Schema, which may
            // be overly broad.
            | UntaggedOperatorName::SetDifference => self.args[0].derive_schema(state),
            // operators returning constants
            UntaggedOperatorName::AllElementsTrue | UntaggedOperatorName::AnyElementTrue | UntaggedOperatorName::And | UntaggedOperatorName::Eq | UntaggedOperatorName::Gt | UntaggedOperatorName::Gte | UntaggedOperatorName::In
            | UntaggedOperatorName::IsArray | UntaggedOperatorName::IsNumber | UntaggedOperatorName::Lt | UntaggedOperatorName::Lte | UntaggedOperatorName::Not | UntaggedOperatorName::Ne | UntaggedOperatorName::Or
            | UntaggedOperatorName::SetEquals | UntaggedOperatorName::SetIsSubset => Ok(Schema::Atomic(Atomic::Boolean)),
            | UntaggedOperatorName::Cmp | UntaggedOperatorName::Strcasecmp | UntaggedOperatorName::StrLenBytes | UntaggedOperatorName::StrLenCP => {
                Ok(Schema::Atomic(Atomic::Integer))
            }
            UntaggedOperatorName::Count => Ok(Schema::AnyOf(set!(
                Schema::Atomic(Atomic::Integer),
                Schema::Atomic(Atomic::Long),
            ))),
            UntaggedOperatorName::Range => Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Integer)))),
            UntaggedOperatorName::Rand => Ok(Schema::Atomic(Atomic::Double)),
            UntaggedOperatorName::Substr | UntaggedOperatorName::SubstrBytes | UntaggedOperatorName::SubstrCP | UntaggedOperatorName::ToLower | UntaggedOperatorName::ToUpper | UntaggedOperatorName::Type => {
                Ok(Schema::Atomic(Atomic::String))
            }
            UntaggedOperatorName::ToHashedIndexKey => Ok(Schema::Atomic(Atomic::Long)),
            // Ops that return a constant schema but must handle nullability
            UntaggedOperatorName::BinarySize |  UntaggedOperatorName::BsonSize | UntaggedOperatorName::IndexOfArray
            | UntaggedOperatorName::IndexOfBytes | UntaggedOperatorName::IndexOfCP | UntaggedOperatorName::Size | UntaggedOperatorName::ToInt => {
                handle_null_satisfaction(
                    args, state,
                    Schema::Atomic(Atomic::Integer),
                )
            }
            UntaggedOperatorName::Concat | UntaggedOperatorName::ToString => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::String),
            ),
            UntaggedOperatorName::TSIncrement | UntaggedOperatorName::TSSecond | UntaggedOperatorName::ToLong => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Long),
            ),
            UntaggedOperatorName::ToBool => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Boolean),
            ),
            UntaggedOperatorName::ToDate => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Date),
            ),
            UntaggedOperatorName::ToDecimal => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Decimal),
            ),
            UntaggedOperatorName::ToDouble => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::Double),
            ),
            UntaggedOperatorName::ToObjectId => handle_null_satisfaction(
                args, state,
                Schema::Atomic(Atomic::ObjectId),
            ),
            UntaggedOperatorName::Avg => {
                if args.len() == 1 {
                    let arg_schema = args[0].derive_schema(state)?;
                    if let Schema::Array(item_schema) = arg_schema {
                        Ok(get_decimal_double_or_nullish_from_schema(*item_schema))
                    } else {
                        Ok(get_decimal_double_or_nullish_from_schema(arg_schema))
                    }
                }
                else {
                    get_decimal_double_or_nullish(args, state)
                }
            }
            // these operators can only return a decimal (if the input is a decimal), double for any other numeric input, or nullish.
            UntaggedOperatorName::Acos | UntaggedOperatorName::Acosh | UntaggedOperatorName::Asin | UntaggedOperatorName::Asinh | UntaggedOperatorName::Atan | UntaggedOperatorName::Atan2 | UntaggedOperatorName::Atanh
            | UntaggedOperatorName::Cos | UntaggedOperatorName::Cosh | UntaggedOperatorName::DegreesToRadians | UntaggedOperatorName::Divide | UntaggedOperatorName::Exp | UntaggedOperatorName::Ln | UntaggedOperatorName::Log
            | UntaggedOperatorName::Log10 | UntaggedOperatorName::RadiansToDegrees | UntaggedOperatorName::Sin | UntaggedOperatorName::Sinh | UntaggedOperatorName::Sqrt | UntaggedOperatorName::Tan | UntaggedOperatorName::Tanh =>
                get_decimal_double_or_nullish(args, state)
            ,
            // if any of the args are long, long; otherwise int. Int, long only possible types
            UntaggedOperatorName::BitAnd | UntaggedOperatorName::BitNot | UntaggedOperatorName::BitOr | UntaggedOperatorName::BitXor => {
                let non_null_schema = match arguments_schema_satisfies(&args, state, &Schema::Atomic(Atomic::Long))? {
                    Satisfaction::Must => Schema::Atomic(Atomic::Long),
                    Satisfaction::May => Schema::AnyOf(set!(
                        Schema::Atomic(Atomic::Long),
                        Schema::Atomic(Atomic::Integer),
                    )),
                    _ => Schema::Atomic(Atomic::Integer),
                };
                Ok(match state.null_behavior {
                    Satisfaction::Not => non_null_schema,
                    Satisfaction::May => Schema::simplify(&Schema::AnyOf(set!(Schema::Atomic(Atomic::Null), non_null_schema))),
                    Satisfaction::Must => Schema::Atomic(Atomic::Null)
                })
            }
            UntaggedOperatorName::Add => {
                // Separate any possible Date arguments from non-Date arguments.
                let arg_schemas = args
                    .iter()
                    .map(|arg| arg
                        .derive_schema(state)
                        .map(|schema| schema.upconvert_missing_to_null())
                    )
                    .collect::<Result<BTreeSet<_>>>()?;
                let (dates, mut non_dates): (BTreeSet<_>, BTreeSet<_>) = arg_schemas
                    .into_iter()
                    .partition(|arg_schema| {
                        arg_schema.satisfies(&Schema::Atomic(Atomic::Date)) != Satisfaction::Not
                    });

                // If there are any Date arguments, if any MUST be (nullable) Date,
                // then we know the result is (nullable) Date. If any MAY be Date or
                // any numeric type, then the result is either Date or the appropriate
                // numeric type based on the rules below. We add the non-Date types to
                // the non_dates set and proceed. At the end, we include Date as a
                // possible result type.
                let mut may_be_date = false;
                for date_schema in dates {
                    if date_schema.satisfies(&DATE_OR_NULLISH.clone()) == Satisfaction::Must {
                        return handle_null_satisfaction(args, state, Schema::Atomic(Atomic::Date))
                    } else {
                        let with_date_removed = date_schema.intersection(&NUMERIC_OR_NULLISH.clone());
                        non_dates.insert(with_date_removed);
                        may_be_date = true;
                    }
                }

                // Here, we (safely) assume (nullable) numeric types for all non-Date arguments.
                let numeric_schema = get_numeric_operator_schema(&args, state)?;

                if may_be_date {
                    Ok(Schema::simplify(&Schema::AnyOf(set!{numeric_schema, Schema::Atomic(Atomic::Date)})))
                } else {
                    Ok(numeric_schema)
                }
            }
            UntaggedOperatorName::Subtract => {
                let left_schema = args.first().unwrap().derive_schema(state)?.upconvert_missing_to_null();
                let right_schema = args.get(1).unwrap().derive_schema(state)?.upconvert_missing_to_null();

                let cp = left_schema.cartesian_product(&right_schema);

                Ok(Schema::simplify(&Schema::AnyOf(cp.into_iter().filter_map(|(l, r)| match (l, r) {
                    // If either operand is Null, the result may be Null
                    (Schema::Atomic(Atomic::Null), _) | (_, Schema::Atomic(Atomic::Null)) => Some(Schema::Atomic(Atomic::Null)),
                    // If both are Date, the result may be Long
                    (Schema::Atomic(Atomic::Date), Schema::Atomic(Atomic::Date)) => Some(Schema::Atomic(Atomic::Long)),
                    // If the first is a Date and the second is numeric, the result may be Date
                    (Schema::Atomic(Atomic::Date), _) => Some(Schema::Atomic(Atomic::Date)),
                    // If the first is a numeric and the second is a Date, this is invalid agg so
                    // we do not compute schema in this case. We include this case here to avoid
                    // accidentally hitting any of the numeric patterns below.
                    (_, Schema::Atomic(Atomic::Date)) => None,
                    // At this point, we (safely) assume only numeric pairs. Anything else is
                    // invalid agg which will produce a runtime error and therefore not require a
                    // schema computation.
                    // If either is a Decimal, the result may be Decimal
                    (Schema::Atomic(Atomic::Decimal), _) | (_, Schema::Atomic(Atomic::Decimal)) => Some(Schema::Atomic(Atomic::Decimal)),
                    // If either is a Double, the result may be Double
                    (Schema::Atomic(Atomic::Double), _) | (_, Schema::Atomic(Atomic::Double)) => Some(Schema::Atomic(Atomic::Double)),
                    // If either is a Long, the result may be Long
                    (Schema::Atomic(Atomic::Long), _) | (_, Schema::Atomic(Atomic::Long)) => Some(Schema::Atomic(Atomic::Long)),
                    // If either is an Integer, the result may be Integer or Long
                    (Schema::Atomic(Atomic::Integer), _) | (_, Schema::Atomic(Atomic::Integer)) => Some(INTEGRAL.clone()),
                    // Again, anything else is invalid agg. Therefore, we do not compute schema
                    // for any other cases.
                    _ => None
                }).collect())))
            }
            // int + int -> int or long; int + long, long + long -> long,
            UntaggedOperatorName::Multiply
            | UntaggedOperatorName::Pow => get_numeric_operator_schema(&args, state),
            // window function operators
            UntaggedOperatorName::CovariancePop | UntaggedOperatorName::CovarianceSamp | UntaggedOperatorName::StdDevPop | UntaggedOperatorName::StdDevSamp => {
                let input_schema = get_input_schema(&args, state)?;
                // window function operators can return null, even if the data is not null, based on the window
                let mut types: BTreeSet<Schema> = set!(Schema::Atomic(Atomic::Null));
                if input_schema.satisfies(&Schema::Atomic(Atomic::Decimal)) != Satisfaction::Not {
                    types.insert(Schema::Atomic(Atomic::Decimal));
                }
                // double for any numeric other than Decimal
                if input_schema.satisfies(&Schema::AnyOf(set!(
                    Schema::Atomic(Atomic::Double),
                    Schema::Atomic(Atomic::Integer),
                    Schema::Atomic(Atomic::Long),
                ))) != Satisfaction::Not
                {
                    types.insert(Schema::Atomic(Atomic::Double));
                }
                Ok(Schema::simplify(&Schema::AnyOf(types)))
            }
            // mod returns the maximal numeric type of its inputs
            UntaggedOperatorName::Mod => {
                let divisor = args[0].derive_schema(state)?.intersection(&NUMERIC_OR_NULLISH.clone());
                let remainder = args[1].derive_schema(state)?.intersection(&NUMERIC_OR_NULLISH.clone());
                if divisor.satisfies(&NULLISH) == Satisfaction::Must || remainder.satisfies(&NULLISH) == Satisfaction::Must {
                    return Ok(Schema::Atomic(Atomic::Null))
                };
                let types = divisor.union(&remainder);
                let lower_bound = match (&divisor, &remainder) {
                    (Schema::Atomic(_), Schema::Atomic(_)) => std::cmp::max(&divisor, &remainder),
                    (a @ Schema::Atomic(_), Schema::AnyOf(ao))
                    | (Schema::AnyOf(ao), a @ Schema::Atomic(_)) => {
                        std::cmp::max(ao.iter().min().unwrap(), a)
                    }
                    (Schema::AnyOf(a), Schema::AnyOf(b)) => {
                        std::cmp::max(a.iter().min().unwrap(), b.iter().min().unwrap())
                    }
                    _ => &Schema::Unsat
                };
                match types {
                    Schema::AnyOf(ao) => Ok(Schema::simplify(&Schema::AnyOf(ao.into_iter().filter(|s| s >= lower_bound).collect()))),
                    schema => Ok(schema)
                }
            }
            // $arrayElemAt operators get the sc hema of an element of the array, handling nullability. If the input
            // is not an array, it can still be valid if it must be null or missing.
            UntaggedOperatorName::ArrayElemAt => {
                let input_schema = self.args[0].derive_schema(state)?;
                match input_schema.clone() {
                    Schema::Array(a) => Ok(Schema::simplify(&a.as_ref().union(&Schema::Atomic(Atomic::Null)))),
                    ao @ Schema::AnyOf(_) => {
                        let array_schema = match ao.intersection(&Schema::Array(Box::new(Schema::Any))) {
                            Schema::Array(a) => *a,
                            _ => Schema::Unsat
                        };
                        let null_schema = ao.intersection(&NULLISH.clone());
                        if array_schema == Schema::Unsat && null_schema == Schema::Unsat {
                            Err(Error::InvalidType(input_schema, 0usize))
                        } else {
                            Ok(Schema::simplify(&null_schema.union(&array_schema).union(&Schema::Atomic(Atomic::Null))))
                        }
                    }
                    _ => {
                        if input_schema.satisfies(&NULLISH_OR_UNDEFINED) == Satisfaction::Must {
                            Ok(Schema::Atomic(Atomic::Null))
                        } else if input_schema.satisfies(&NULLISH_OR_UNDEFINED.union(&Schema::Array(Box::new(Schema::Any)))) == Satisfaction::Must {
                            Ok(input_schema.intersection(&Schema::Array(Box::new(Schema::Any))).union(&Schema::Atomic(Atomic::Null)))
                        } else {
                            Err(Error::InvalidType(input_schema, 0usize))
                        }
                    }
                }
            }
            UntaggedOperatorName::ArrayToObject => {
                // We could only know the keys, if we have the entire array.
                // We may consider making this more precise for array literals.
                // For now, just return an Any Document while capturing nullability
                let schema = self.args[0].derive_schema(state)?;
                Ok(match schema.intersection(&NULLISH.clone()) {
                    Schema::Unsat => Schema::Document(Document::any()),
                    nullish_schema => Schema::simplify(&nullish_schema.union(&Schema::Document(Document::any()))),
                })
            }
            // $concatArrays and $setUnion operators both return array schemas where the inner objects
            // have a schema that is a union of the inner schemas of all of the inputs
            UntaggedOperatorName::ConcatArrays | UntaggedOperatorName::SetUnion => {
                let mut array_schema = Schema::Unsat;
                let mut null_schema: Option<Schema> = None;
                for (i, arg) in self.args.iter().enumerate() {
                    let schema = arg.derive_schema(state)?;
                    match schema {
                        Schema::Array(a) => array_schema = array_schema.union(a.as_ref()),
                        // anyofs can capture nullish behavior in these operators
                        Schema::AnyOf(ao) => {
                            ao.iter().for_each(|ao_schema| {
                                match ao_schema {
                                    s @ (Schema::Atomic(Atomic::Null) | Schema::Missing) => {
                                        if let Some(ns) = null_schema.as_ref() {
                                            null_schema = Some(ns.union(s));
                                        } else {
                                            null_schema = Some(s.clone())
                                        }
                                    }
                                    Schema::Array(a) => array_schema = array_schema.union(a.as_ref()),
                                    _ => {},
                                }
                            });
                        }
                        _ => return Err(Error::InvalidType(schema, i)),
                    };
                }
                if let Some(null_schema) = null_schema {
                    Ok(null_schema.union(&Schema::Array(Box::new(array_schema))))
                } else {
                    Ok(Schema::Array(Box::new(array_schema)))
                }
            }
            // $setIntersection gets the intersecting types from the input arguments, while
            // accounting for numeric intersection (even if the types aren't exact)
            UntaggedOperatorName::SetIntersection => {
                if args.is_empty() {
                    return Ok(Schema::Array(Box::new(Schema::Unsat)));
                }
                let mut array_schema = match args.remove(0_usize).derive_schema(state)?{
                    Schema::Array(a) => *a,
                    Schema::Missing => return Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Null)))),
                    schema => return Err(Error::InvalidType(schema, 0)),
                };
                for (i, arg) in self.args.iter().enumerate() {
                    let schema = arg.derive_schema(state)?;
                    match schema {
                        Schema::Array(a) => {
                            // If the array_schema MAY satisify numeric, we need to augment the rhs
                            // of the intersection with the entire numeric Schema set because 42.0
                            // as a double is considered equivalent to 42 as an integer in mongo,
                            // meaning that intersection of [42.0] and [42] is [42.0]. Note
                            // specifically that Mongo retains the lhs value when there is
                            // equivalent numeric values in the rhs. This is why we pull out the
                            // first Schema individually before the loop.
                            let a = if a.satisfies(&NUMERIC) >= Satisfaction::May {
                                a.union(&NUMERIC)
                            } else {
                                *a
                            };
                            array_schema = array_schema.intersection(&a)
                        }
                        Schema::Missing => return Ok(Schema::Array(Box::new(Schema::Atomic(Atomic::Null)))),
                        _ => return Err(Error::InvalidType(schema, i+1)),
                    };
                }
                Ok(Schema::Array(Box::new(array_schema)))
            }
            UntaggedOperatorName::Locf => {
                self.args[0].derive_schema(state)
            }
            // $max and $min simply union together the schemas of all of the arguments
            // NOTE: this could be improved in precision, if one argument's schema supercedes all the others
            // however, this is a bit tricky to implement given AnyOfs / Any.
            UntaggedOperatorName::Max
            | UntaggedOperatorName::Min => {
                let schema = self.args.iter().try_fold(Schema::Unsat, |acc, arg| {
                    let arg_schema = arg.derive_schema(state)?.upconvert_missing_to_null();
                    Ok(acc.union(&arg_schema))
                })?;
                Ok(schema)
            }
            UntaggedOperatorName::AddToSet | UntaggedOperatorName::Push => {
                let schema = self.args[0].derive_schema(state)?;
                Ok(Schema::Array(Box::new(schema)))
            }
            UntaggedOperatorName::IfNull => {
                // Note that $ifNull is variadic, not binary. It returns the first
                // non-null argument. If all arguments are nullish, it returns the
                // last argument unmodified (meaning, even if the last argument is
                // null or missing, that value is returned). Therefore, the schema
                // for this expression is the union of all argument schemas up to
                // the first non-nullish argument (minus the nullish types).
                //
                // If all arguments up to the last one are nullish, then it is the
                // union of all argument schemas. The schema retains any nullish
                // types that the last argument may satisfy.
                let mut schema = Schema::Unsat;
                let last_elem_idx = args.len() - 1;
                for (i, arg) in args.into_iter().enumerate() {
                    let arg_schema = arg.derive_schema(state)?;
                    if i == last_elem_idx {
                        // If we get to the last element, we do not want to remove
                        // nullish types from this argument's schema since $ifNull
                        // returns this value no matter what.
                        schema = schema.union(&arg_schema);
                        break;
                    }

                    match arg_schema.satisfies(&NULLISH) {
                        // If this argument is never nullish, include this schema
                        // and break.
                        Satisfaction::Not => {
                            schema = schema.union(&arg_schema);
                            break;
                        }
                        // If this argument may be nullish, retain the non-nullish
                        // types only.
                        Satisfaction::May => {
                            schema = schema.union(&arg_schema.subtract_nullish());
                        }
                        // If this argument must be nullish, ignore it
                        Satisfaction::Must => {}
                    }
                }

                Ok(schema)
            }
            UntaggedOperatorName::MergeObjects => {
                // $mergeObjects ignores nullish arguments. If all arguments are
                // nullish, then the result is the empty document schema. It is
                // tempting to simply use document schema union to represent the
                // result schema for this operator, however that is not exactly
                // correct. $mergeObjects uses the last value for a key if it
                // appears multiple times. See the tests for examples.
                let arg_schemas: Result<Vec<Document>> = args.iter().filter_map(|arg| {
                    let arg_schema = arg.derive_schema(state);
                    match arg_schema {
                        Err(e) => Some(Err(e)),
                        Ok(arg_schema) => {
                            fn retain_only_doc_schemas(sch: Schema) -> Option<Schema> {
                                match sch {
                                    Schema::Unsat => None,
                                    Schema::Missing => None,
                                    Schema::Atomic(_) => None,
                                    Schema::Array(_) => None,
                                    Schema::Document(_) => Some(sch),
                                    Schema::Any => Some(ANY_DOCUMENT.clone()),
                                    Schema::AnyOf(ao) => {
                                        // Retain only the Document schemas in this AnyOf and
                                        // union them all together. The presence of any types
                                        // other than Document implies that any "required"
                                        // fields in any Document schemas are not required in
                                        // the resulting schema. This is achieved by starting
                                        // the fold with EMPTY_DOCUMENT. If the AnyOf only
                                        // contains Document schemas, then some fields in the
                                        // result schema may actually be required. This is
                                        // achieved by starting the fold with the first schema
                                        // from the AnyOf.
                                        if ao.is_empty() {
                                            return None
                                        }
                                        let init_doc_schema = if ao.iter().all(|s| matches!(s, Schema::Document(_))) {
                                            // At this point, we know ao is non-empty and contains
                                            // only document schemas.
                                            ao.first().unwrap().clone()
                                        } else {
                                            EMPTY_DOCUMENT.clone()
                                        };

                                        Some(ao.into_iter()
                                            .filter_map(retain_only_doc_schemas)
                                            .fold(init_doc_schema, Schema::document_union))
                                    }
                                }
                            }

                            match retain_only_doc_schemas(arg_schema) {
                                Some(Schema::Document(d)) => Some(Ok(d)),
                                _ => None,
                            }
                        }
                    }
                }).collect();
                Ok(Schema::simplify(&Schema::Document(arg_schemas?
                    .into_iter()
                    .fold(Document::empty(), |acc, arg_schema| {
                        // Generally, mergeObjects retains the last value seen
                        // for a key. Therefore, we iterate through the keys
                        // of this argument and insert them and their schemas.
                        let mut keys = acc.keys;
                        for (arg_key, mut arg_key_schema) in arg_schema.keys {
                            let current_key_schema = keys.get(&arg_key);
                            let schema_to_insert = if let Some(current_key_schema) = current_key_schema {
                                if arg_key_schema.satisfies(&Schema::Missing) == Satisfaction::May {
                                    // If this key already appears in the accumulated schema _and_
                                    // this argument's schema for this key is possibly missing, then
                                    // we cannot simply overwrite the accumulated schema for this
                                    // key. This is because in the case the later document's value
                                    // for this key is missing, the earlier document's value will be
                                    // returned. Therefore, we must union the accumulated schema and
                                    // this argument's schema for this key. See the tests for an
                                    // example.
                                    schema_difference(&mut arg_key_schema, set!{Schema::Missing});
                                    arg_key_schema.union(current_key_schema)
                                } else {
                                    arg_key_schema
                                }
                            } else {
                                arg_key_schema
                            };

                            keys.insert(arg_key, schema_to_insert);
                        }

                        // All required keys must still be required.
                        let mut required = acc.required;
                        required.extend(arg_schema.required);

                        // If any Document allows additional properties, the result
                        // must also allow additional_properties.
                        let additional_properties = acc.additional_properties || arg_schema.additional_properties;

                        Document {
                            keys,
                            required,
                            additional_properties,
                            ..Default::default()
                        }
                    })
                )))
            }
            // $objectToArray turns an object into an array of documents with fields k (the field names) and v (the field types)
            // we produce an Array(Document {k: String, v: ...}) where v's schema is the union of all value types.
            UntaggedOperatorName::ObjectToArray => {
                let input_doc = match &self.args[0].derive_schema(state)? {
                    Schema::Document(d) => Some(d.clone()),
                    Schema::AnyOf(ao) => {
                        let mut doc_type = None;
                        for schema in ao.iter() {
                            if let Schema::Document(d) = schema {
                                doc_type = Some(d.clone());
                                break;
                            }
                        }
                        doc_type
                    }
                    Schema::Any => Some(Document::any()),
                    _ => None
                };
                if let Some(d) = input_doc {
                    let document_value_types = d.keys.into_iter().try_fold(Schema::Unsat, |schema , (_, arg_schema)| {
                        Ok(match schema {
                            Schema::Unsat => arg_schema,
                            schema => schema.union(&arg_schema)
                        })
                    })?;
                    let array_type = Schema::Array(Box::new(Schema::Document(Document { keys: map! {
                            "k".to_string() => Schema::Atomic(Atomic::String),
                            "v".to_string() => document_value_types
                        },
                        required: set!("k".to_string(), "v".to_string()),
                        ..Default::default()
                    })));
                    Ok(handle_null_satisfaction(vec![args[0]], state, array_type)?)
                } else {
                    Err(Error::InvalidExpressionForField(
                        format!("{:?}", self.args[0]),
                        "object",))
                }
            }
            // described above, $sum ignores all non-numeric arguments and handles calculation of the
            // maximum numeric type. It can also handle summing not only arrays of arguments but singular
            // arguments, including numerics themselves
            UntaggedOperatorName::Sum => {
                if args.len() == 1 {
                    let arg_schema = args[0].derive_schema(state)?;
                    if let Schema::Array(item_schema) = arg_schema {
                        return Ok(get_sum_type(*item_schema))
                    }
                }
                let schema = get_numeric_operator_schema(&args, state)?;
                match schema {
                    Schema::Unsat => Ok(Schema::Atomic(Atomic::Integer)),
                    schema => Ok(schema)
                }
            }
            UntaggedOperatorName::First | UntaggedOperatorName::Last => {
                let schema = self.args[0].derive_schema(state)?.upconvert_missing_to_null();
                Ok(match schema {
                    Schema::Array(a) => *a,
                    schema => schema
                })
            }
            UntaggedOperatorName::Literal => self.args[0].derive_schema(state),
            UntaggedOperatorName::Split => handle_null_satisfaction(
                args, state,
                Schema::Array(Box::new(Schema::Atomic(Atomic::String))),
            ),
            _ => Err(Error::InvalidUntaggedOperator(self.op.into())),
        }
    }
}
