pub mod custom_serde;
pub mod definitions;
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
