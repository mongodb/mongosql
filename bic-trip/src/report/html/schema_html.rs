use build_html::{Container, HtmlContainer};

use crate::schema::CollectionAnalysis;

pub fn add_schema_analysis_html(analysis: Option<&crate::schema::SchemaAnalysis>) -> Container {
    if analysis.is_none() {
        return Container::default();
    }

    let analysis = analysis.as_ref().unwrap();
    let mut contents_html = Container::default();

    for (db, db_analysis) in &analysis.database_analyses {
        let mut db_html = Container::default().with_header(3, format!("Database: {}", db));
        for collection_analysis in &db_analysis.collection_analyses {
            db_html.add_container(
                Container::default()
                    .with_header(
                        4,
                        format!("Collection: {}", &collection_analysis.collection_name),
                    )
                    .with_raw(collection_summarization(collection_analysis)),
            )
        }
        contents_html.add_container(db_html);
    }
    contents_html
}

fn collection_summarization(collection_analysis: &CollectionAnalysis) -> String {
    let arrays = if collection_analysis.arrays.is_empty() {
        None
    } else {
        Some(format!(
            r#"<li>Arrays found: {}. Arrays can be expanded with the <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/query-with-asql-statements/#unwind" target="_blank">UNWIND</a> operator.</li>"#,
            collection_analysis.arrays.len()
        ))
    };

    let arrays_of_arrays = if collection_analysis.arrays_of_arrays.is_empty() {
        None
    } else {
        Some(format!(
            r#"<li>Arrays of arrays found: {}. Arrays of arrays can be expanded with nested <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/query-with-asql-statements/#unwind" target="_blank">UNWIND</a> operators (i.e. UNWIND(UNWIND(...))).</li>"#,
            collection_analysis.arrays_of_arrays.len()
        ))
    };
    let arrays_of_objects = if collection_analysis.arrays_of_objects.is_empty() {
        None
    } else {
        Some(format!(
            r#"<li>Arrays of objects found: {}. <p>Arrays of objects can be expanded with the
            <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/query-with-asql-statements/#unwind" target="_blank">UNWIND</a> operator,
            and the objects can be flattened the <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/query-with-asql-statements/#flatten" target="_blank">FLATTEN</a> operator (i.e. FLATTEN(UNWIND(...))).</p></li>"#,
            collection_analysis.arrays_of_objects.len()
        ))
    };
    let objects = if collection_analysis.objects.is_empty() {
        None
    } else {
        Some(format!(
            r#"<li>Objects found: {}. Objects can be flattened with the <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/query-with-asql-statements/#flatten" target="_blank">FLATTEN</a> operator.</li>"#,
            collection_analysis.objects.len()
        ))
    };
    let anyof = if collection_analysis.anyof.is_empty() {
        None
    } else {
        Some(format!(
            r#"<li>Field with multiple types found: {}.
            For best results, ensure <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/language-reference/#type-conversions" target="_blank">CAST</a> is used on these fields to force the same type for analysis in your
            preferred BI tool.</p><ul>{}</ul></li>"#,
            collection_analysis.anyof.len(),
            collection_analysis
                .anyof
                .iter()
                .map(|(k, (_, types))| format!(
                    r#"<li class="p-l-10"=><code>{}</code> with types: {}</li>"#,
                    k,
                    simplify_polymorphic(types).join(", ")
                ))
                .collect::<Vec<String>>()
                .join("")
        ))
    };
    let unstable = if collection_analysis.unstable.is_empty() {
        None
    } else {
        Some(format!(
            r#"<li>Fields found to be unstable: {}. <p>These fields exhibit significant schema and type variations across objects,
            leading to potential performance degredation if processed.</p>
            <p>See <a href="https://www.mongodb.com/scale/designing-a-database-schema" target="_blank">Designing a Database Schema</a>
            for information and ideas how to correct this.<ul>{}</ul></p></li>"#,
            collection_analysis.unstable.len(),
            collection_analysis
                .unstable
                .iter()
                .map(|(k, v)| format!("<li>Field <code>{}</code> at depth {}</li>", k, v))
                .collect::<Vec<String>>()
                .join("")
        ))
    };
    let results = vec![
        arrays,
        arrays_of_arrays,
        arrays_of_objects,
        objects,
        anyof,
        unstable,
    ]
    .into_iter()
    .flatten()
    .collect::<Vec<String>>()
    .join("");
    format!(
        "<ul>{}</ul>",
        if results.is_empty() {
            "Schema analysis found no items to highlight.".to_string()
        } else {
            format!("{}{}", find_max_depth(collection_analysis), results)
        }
    )
}

/// simplify_polymorphic takes in a `Vec<String>` that represents a polymoprhic
/// fields found types and simplifies it for users so they don't see multiple
/// instances of Object or Array
fn simplify_polymorphic(types: &[String]) -> Vec<String> {
    let array_type = types.iter().filter(|elem| *elem == "Array").count();
    let object_type = types.iter().filter(|elem| *elem == "Object").count();
    let mut simplified_types: Vec<String> = types
        .iter()
        .map(|elem| {
            if elem == "Array" && array_type > 1 {
                "Arrays of multiple types".to_string()
            } else if elem == "Object" && object_type > 1 {
                "Objects of multiple types".to_string()
            } else {
                elem.to_string()
            }
        })
        .collect();
    simplified_types.dedup();
    simplified_types
}

fn find_max_depth(collection_analysis: &CollectionAnalysis) -> String {
    let mut max_depth = 0;
    for v in collection_analysis.arrays.values() {
        if *v > max_depth {
            max_depth = *v;
        }
    }
    for v in collection_analysis.arrays_of_arrays.values() {
        if *v > max_depth {
            max_depth = *v;
        }
    }
    for v in collection_analysis.arrays_of_objects.values() {
        if *v > max_depth {
            max_depth = *v;
        }
    }
    for v in collection_analysis.objects.values() {
        if *v > max_depth {
            max_depth = *v;
        }
    }
    for v in collection_analysis.unstable.values() {
        if *v > max_depth {
            max_depth = *v;
        }
    }
    format!(
        r#"<li>The maximum nesting depth found for this collection is: <code>{}</code>.</li>"#,
        max_depth
    )
}
