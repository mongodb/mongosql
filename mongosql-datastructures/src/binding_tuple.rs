use std::collections::{btree_map, BTreeMap};

#[derive(Debug)]
pub struct DuplicateKeyError {
    pub key: Key,
}

#[derive(PartialEq, Eq, Debug, Clone)]
pub struct BindingTuple<T>(pub BTreeMap<Key, T>);

#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone, Hash)]
pub struct Key {
    pub datasource: DatasourceName,
    pub scope: u16,
}

impl Key {
    pub fn bot(scope: u16) -> Self {
        Self {
            datasource: DatasourceName::Bottom,
            scope,
        }
    }

    pub fn named(name: &str, scope: u16) -> Self {
        Self {
            datasource: DatasourceName::Named(name.to_string()),
            scope,
        }
    }
}

impl<D, S> From<(D, S)> for Key
where
    D: Into<DatasourceName>,
    S: Into<u16>,
{
    fn from(tup: (D, S)) -> Self {
        let (datasource_name, scope) = tup;
        Self {
            datasource: datasource_name.into(),
            scope: scope.into(),
        }
    }
}

#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone, Hash)]
pub enum DatasourceName {
    Bottom,
    Named(String),
}

impl<S> From<S> for DatasourceName
where
    S: Into<String>,
{
    fn from(name: S) -> Self {
        Self::Named(name.into())
    }
}

impl<T> BindingTuple<T> {
    pub fn new() -> BindingTuple<T> {
        BindingTuple(BTreeMap::new())
    }

    pub fn nearest_scope_for_datasource(&self, d: &DatasourceName, scope: u16) -> Option<u16> {
        for current_scope in (0..=scope).rev() {
            let possible_key = Key {
                datasource: d.clone(),
                scope: current_scope,
            };

            if self.contains_key(&possible_key) {
                return Some(current_scope);
            }
        }
        None
    }

    pub fn len(&self) -> usize {
        self.0.len()
    }

    pub fn get(&self, k: &Key) -> Option<&T> {
        self.0.get(k)
    }

    pub fn remove(&mut self, k: &Key) -> Option<T> {
        self.0.remove(k)
    }

    pub fn contains_key(&self, k: &Key) -> bool {
        self.0.contains_key(k)
    }

    pub fn insert(&mut self, k: Key, v: T) -> Option<T> {
        self.0.insert(k, v)
    }

    pub fn is_empty(&self) -> bool {
        self.0.is_empty()
    }

    pub fn keys(&self) -> impl Iterator<Item = &Key> {
        self.0.keys()
    }

    pub fn iter(&self) -> impl Iterator<Item = (&Key, &T)> {
        self.0.iter()
    }

    pub fn merge(&mut self, other: BindingTuple<T>) -> Result<(), DuplicateKeyError> {
        for (k, v) in other.0.into_iter() {
            if self.0.insert(k.clone(), v).is_some() {
                return Err(DuplicateKeyError { key: k });
            }
        }
        Ok(())
    }

    pub fn with_merged_mappings(
        mut self,
        mappings: BindingTuple<T>,
    ) -> Result<Self, DuplicateKeyError> {
        self.merge(mappings)?;
        Ok(self)
    }
}

impl<T> IntoIterator for BindingTuple<T> {
    type Item = (Key, T);
    type IntoIter = btree_map::IntoIter<Key, T>;

    fn into_iter(self) -> Self::IntoIter {
        self.0.into_iter()
    }
}

impl<T> Default for BindingTuple<T> {
    fn default() -> Self {
        BindingTuple(BTreeMap::default())
    }
}

impl<T> FromIterator<(Key, T)> for BindingTuple<T> {
    fn from_iter<I: IntoIterator<Item = (Key, T)>>(iter: I) -> Self {
        let mut bt = BindingTuple(BTreeMap::new());
        for (k, v) in iter {
            bt.insert(k, v);
        }
        bt
    }
}
