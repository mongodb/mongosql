/// A time provider
pub trait ClockProvider {
    /// Provide the current amount of seconds since unix epoch (1970-01-01)
    fn now_unix(&self) -> i64;
}
