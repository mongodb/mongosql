use std::{
    collections::HashMap,
    ops::{Deref, DerefMut},
};

use crate::NamespaceInfoWithSchema;

/// A Schema Catalog
///
/// The catalog contains all schemas in a database by their namespace
#[derive(Debug, Default)]
pub struct Catalog(HashMap<String, NamespaceInfoWithSchema>);

impl IntoIterator for Catalog {
    type Item = (String, NamespaceInfoWithSchema);
    type IntoIter = std::collections::hash_map::IntoIter<String, NamespaceInfoWithSchema>;

    fn into_iter(self) -> Self::IntoIter {
        self.0.into_iter()
    }
}

impl Deref for Catalog {
    type Target = HashMap<String, NamespaceInfoWithSchema>;

    fn deref(&self) -> &Self::Target {
        &self.0
    }
}

impl DerefMut for Catalog {
    fn deref_mut(&mut self) -> &mut Self::Target {
        &mut self.0
    }
}

impl FromIterator<(String, NamespaceInfoWithSchema)> for Catalog {
    fn from_iter<T: IntoIterator<Item = (String, NamespaceInfoWithSchema)>>(iter: T) -> Self {
        Catalog(iter.into_iter().collect())
    }
}
