#[macro_export]
macro_rules! map {
	($($key:expr => $val:expr),* $(,)?) => {
		std::iter::Iterator::collect(std::array::IntoIter::new([
			$({
				($key, $val)
			},)*
		]))
	};
}
