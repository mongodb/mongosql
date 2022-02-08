pub use mongosql_datastructures::unique_linked_hash_map;
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
macro_rules! set {
	($($val:expr),* $(,)?) => {
		std::iter::Iterator::collect([
			$({
				($val)
			},)*
		].into_iter())
	};
}

// The unchecked version unwraps insertions. This should only be used for testing.
#[cfg(test)]
#[macro_export]
macro_rules! unchecked_unique_linked_hash_map {
	($($key:expr => $val:expr),* $(,)?) => {{
            #[allow(unused_mut)]
            let mut out = mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap::new();
            $(
                out.insert($key, $val).unwrap();
            )*
            out
	}};
}
