use crate::air::optimizer::Pass;

/// Accepts a string representing the current database and returns a Pass that
/// will remove the from_db name from any Lookup when the from_db matches the
/// current database name. This is because the information is redundant and
/// some implementions of mongodb aggregation language, e.g. mongod itself, do
/// not support the from_db field.
pub fn get_simplify_lookups(_current_db: &str) -> Box<dyn Pass> {
    todo!()
}
