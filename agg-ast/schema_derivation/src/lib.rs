#[cfg(test)]
mod test;

use mongosql::schema::Schema;

#[allow(dead_code)]
pub(crate) fn get_schema_for_path_mut(
    schema: &mut Schema,
    path: Vec<String>,
) -> Option<&mut Schema> {
    let mut schema = Some(schema);
    for field in path {
        schema = match schema {
            Some(Schema::Document(d)) => d.keys.get_mut(&field),
            _ => {
                return None;
            }
        };
    }
    schema
}
