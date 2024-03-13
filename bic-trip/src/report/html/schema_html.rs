use build_html::{Container, HtmlContainer};

use crate::schema::CollectionAnalysis;

pub fn add_schema_analysis_html(analysis: &Option<crate::schema::SchemaAnalysis>) -> Container {
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
    let arrays_of_documents = if collection_analysis.arrays_of_documents.is_empty() {
        None
    } else {
        Some(format!(
            r#"<li>Arrays of documents found: {}. <p>Arrays of documents can be expanded with the
            <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/query-with-asql-statements/#unwind" target="_blank">UNWIND</a> operator,
            and the documents can be flattened the <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/query-with-asql-statements/#flatten" target="_blank">FLATTEN</a> operator (i.e. FLATTEN(UNWIND(...))).</p></li>"#,
            collection_analysis.arrays_of_documents.len()
        ))
    };
    let documents = if collection_analysis.documents.is_empty() {
        None
    } else {
        Some(format!(
            r#"<li>Documents found: {}. Documents can be flattened with the <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/query-with-asql-statements/#flatten" target="_blank">FLATTEN</a> operator.</li>"#,
            collection_analysis.documents.len()
        ))
    };
    let anyof = if collection_analysis.anyof.is_empty() {
        None
    } else {
        Some(format!(
            r#"<li>Field with multiple types found: {}.<p>These fields have been identified
            as holding multiple types of values, for example integers and Strings (5 or "5"),
            or some fields may contain a <code>null</code> value or be missing from documents in the collection.
            For best results, ensure <a href="https://www.mongodb.com/docs/atlas/data-federation/query/sql/language-reference/#type-conversions" target="_blank">CAST</a> is used on these fields to force the same type for analysis in your
            preferred BI tool.</p><ul>{}</ul></li>"#,
            collection_analysis.anyof.len(),
            collection_analysis
                .anyof
                .iter()
                .map(|(k, v)| format!("<li>Field <code>{}</code> at depth {}</li>", k, v))
                .collect::<Vec<String>>()
                .join("")
        ))
    };
    let unstable = if collection_analysis.unstable.is_empty() {
        None
    } else {
        Some(format!(
            r#"<li>Fields found to be unstable: {}. <p>These fields exhibit significant schema and type variations across documents,
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
        arrays_of_documents,
        documents,
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
            results
        }
    )
}
