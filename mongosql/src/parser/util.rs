// process_delimited_ident removes the outer delimiters from an identifier and
// processes any escaped delimiters.
pub fn process_delimited_ident(value: &str) -> String {
    let delimiter = value.chars().next().unwrap_or_default();
    let trimmed_value = value[1..value.len() - 1].to_string();
    if delimiter == '\"' {
        trimmed_value.replace("\"\"", "\"")
    } else if delimiter == '`' {
        trimmed_value.replace("``", "`")
    } else {
        unreachable!("delimiters other than double-quote and backtick are not supported")
    }
}
