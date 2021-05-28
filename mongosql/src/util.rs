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

#[macro_export]

macro_rules! set {
	($($val:expr),* $(,)?) => {
		std::iter::Iterator::collect(std::array::IntoIter::new([
			$({
				($val)
			},)*
		]))
	};
}
