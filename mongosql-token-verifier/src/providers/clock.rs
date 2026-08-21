use std::time::{SystemTime, UNIX_EPOCH};

/// A time provider
pub trait ClockProvider {
    /// Provide the current amount of seconds since unix epoch (1970-01-01)
    fn now_unix(&self) -> i64;
}

/// An implementation of the clock provider using system time
pub struct SystemClock;
impl ClockProvider for SystemClock {
    fn now_unix(&self) -> i64 {
        SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map(|d| d.as_secs() as i64)
            .unwrap_or(0)
    }
}
