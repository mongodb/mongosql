pub mod custom_serde;
pub mod definitions;
pub use definitions::Namespace;
#[cfg(test)]
mod serde_test;

pub const ROOT_NAME: &str = "ROOT";
pub const PRUNE_NAME: &str = "PRUNE";

#[allow(dead_code)]
pub const KEEP_NAME: &str = "KEEP";
#[allow(dead_code)]
pub const DESCEND_NAME: &str = "DESCEND";

#[macro_export]
macro_rules! map {
	($($key:expr => $val:expr),* $(,)?) => {
		std::iter::Iterator::collect([
			$({
				($key, $val)
			},)*
		].into_iter())
	};
}

#[macro_export]
macro_rules! vector_pipeline {
    () => {
           vec![Stage::AtlasSearchStage(VectorSearch(Box::new(
                Expression::Document(map! {
                    "index".to_string() => Literal(LiteralValue::String("movie_collection_index".to_string())),
                    "path".to_string() => Literal(LiteralValue::String("title".to_string())),
                    "queryVector".to_string() => Expression::Array(vec![Literal(LiteralValue::Double(10.6)), Expression::Literal(LiteralValue::Double(60.5))]),
                    "numCandidates".to_string() => Literal(LiteralValue::Int32(500)),
                }),
            )))]
    };
}

#[macro_export]
macro_rules! text_search_pipeline {
    () => {

        vec![Stage::AtlasSearchStage(
            Search(Box::new(Expression::Document(
                map! {
                    "index".to_string() => Literal(LiteralValue::String("hybrid-full-text-search".to_string())),
                    "phrase".to_string() => Expression::Document(map! {
                        "query".to_string() => Literal(LiteralValue::String("star wars".to_string())),
                        "path".to_string() => Literal(LiteralValue::String("title".to_string())),
                    })
                },
            )))
        ), Limit(20)]
    };
}
