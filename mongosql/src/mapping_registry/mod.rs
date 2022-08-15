pub use mongosql_datastructures::binding_tuple::Key;
use std::collections::BTreeMap;

#[derive(PartialEq, Eq, Debug, Clone)]
pub struct MqlMappingRegistry(BTreeMap<Key, String>);

impl MqlMappingRegistry {
    pub fn new() -> Self {
        MqlMappingRegistry(BTreeMap::new())
    }

    pub fn with_registry(tree: BTreeMap<Key, String>) -> Self {
        MqlMappingRegistry(tree)
    }

    pub fn get(&self, k: &Key) -> Option<&String> {
        self.0.get(k)
    }

    pub fn get_registry(&self) -> &BTreeMap<Key, String> {
        &self.0
    }

    pub fn remove(&mut self, k: &Key) -> Option<String> {
        self.0.remove(k)
    }

    pub fn insert<K: Into<Key>, V: Into<String>>(&mut self, k: K, v: V) -> Option<String> {
        self.0.insert(k.into(), v.into())
    }

    pub fn merge(&mut self, other: MqlMappingRegistry) -> &mut Self {
        self.0.extend(other.0.into_iter());
        self
    }
}

impl Default for MqlMappingRegistry {
    fn default() -> Self {
        Self::new()
    }
}
