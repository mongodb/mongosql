pub mod definitions;
#[cfg(test)]
mod deserialize_test;

pub const ROOT_NAME: &str = "ROOT";

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
