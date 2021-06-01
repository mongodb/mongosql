use std::{
    collections::{
        btree_map::{IntoIter, Iter, Keys},
        BTreeMap,
    },
    iter::FromIterator,
};

#[derive(PartialEq, Eq, Debug, Clone)]
pub struct BindingTuple<T>(pub BTreeMap<Key, T>);

#[allow(dead_code)]
#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone)]
pub struct Key {
    pub datasource: DatasourceName,
    pub scope: u16,
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

#[allow(dead_code)]
#[derive(PartialEq, Eq, PartialOrd, Ord, Debug, Clone)]
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

    pub fn get(&self, k: &Key) -> Option<&T> {
        self.0.get(k)
    }

    #[allow(dead_code)]
    pub fn contains_key(&self, k: &Key) -> bool {
        self.0.contains_key(k)
    }

    pub fn insert(&mut self, k: Key, v: T) -> Option<T> {
        self.0.insert(k, v)
    }

    pub fn is_empty(&self) -> bool {
        self.0.is_empty()
    }

    #[allow(dead_code)]
    pub fn keys(&self) -> Keys<Key, T> {
        self.0.keys()
    }

    #[allow(dead_code)]
    pub fn iter(&self) -> Iter<Key, T> {
        self.0.iter()
    }

    #[allow(dead_code)]
    pub fn into_iter(self) -> IntoIter<Key, T> {
        self.0.into_iter()
    }

    pub fn merge(&mut self, other: BindingTuple<T>) {
        self.0.extend(other.0.into_iter());
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
