use linked_hash_map::LinkedHashMap;
use std::{fmt::Display, hash::Hash, iter::IntoIterator};
use thiserror::Error;

#[derive(Debug, Hash, Default, Clone, PartialEq, Eq)]
pub struct UniqueLinkedHashMap<K, V>(LinkedHashMap<K, V>)
where
    K: Hash + Eq + PartialEq + Display;

#[derive(Debug, Error, PartialEq, Eq)]
#[error("duplicate key found: {0}")]
pub struct DuplicateKeyError(pub String);

impl DuplicateKeyError {
    pub fn get_key_name(self) -> String {
        self.0
    }
}

impl<K, V> UniqueLinkedHashMap<K, V>
where
    K: Hash + PartialEq + Eq + Display,
{
    pub fn new() -> Self {
        Self(LinkedHashMap::new())
    }

    pub fn insert_many(
        &mut self,
        other: impl Iterator<Item = (K, V)>,
    ) -> Result<(), DuplicateKeyError> {
        for (k, v) in other {
            self.insert(k, v)?;
        }
        Ok(())
    }

    pub fn insert(&mut self, k: K, v: V) -> Result<(), DuplicateKeyError> {
        // We check if the key already exists to avoid the clone
        // necessary to check _after_ inserting, since we want
        // to return the key in the error, not the value.
        if self.0.contains_key(&k) {
            return Err(DuplicateKeyError(format!("{}", k)));
        }
        self.0.insert(k, v);
        Ok(())
    }

    pub fn get(&self, k: &K) -> Option<&V> {
        self.0.get(k)
    }

    pub fn remove(&mut self, k: &K) -> Option<V> {
        self.0.remove(k)
    }

    pub fn contains_key(&self, k: &K) -> bool {
        self.0.contains_key(k)
    }

    pub fn is_empty(&self) -> bool {
        self.0.is_empty()
    }

    pub fn keys(&self) -> impl Iterator<Item = &K> {
        self.0.keys()
    }

    pub fn iter(&self) -> impl Iterator<Item = (&K, &V)> {
        self.0.iter()
    }
}

impl<K, V> IntoIterator for UniqueLinkedHashMap<K, V>
where
    K: Hash + PartialEq + Eq + Display,
{
    type Item = (K, V);
    type IntoIter = linked_hash_map::IntoIter<K, V>;

    fn into_iter(self) -> Self::IntoIter {
        self.0.into_iter()
    }
}

pub struct UniqueLinkedHashMapEntry<K, V>(K, V);

impl<K, V> UniqueLinkedHashMapEntry<K, V>
where
    K: Hash + PartialEq + Eq + Display,
{
    pub fn new(k: K, v: V) -> Self {
        UniqueLinkedHashMapEntry(k, v)
    }
}

impl<K, V> FromIterator<UniqueLinkedHashMapEntry<K, V>>
    for Result<UniqueLinkedHashMap<K, V>, DuplicateKeyError>
where
    K: Hash + PartialEq + Eq + Display,
{
    fn from_iter<I: IntoIterator<Item = UniqueLinkedHashMapEntry<K, V>>>(iter: I) -> Self {
        let mut hm = UniqueLinkedHashMap::new();
        for entry in iter {
            hm.insert(entry.0, entry.1)?;
        }
        Ok(hm)
    }
}

impl<K, V> From<UniqueLinkedHashMap<K, V>> for LinkedHashMap<K, V>
where
    K: Hash + Eq + PartialEq + Display,
{
    fn from(ulhm: UniqueLinkedHashMap<K, V>) -> Self {
        ulhm.0
    }
}

impl<'a, K, V> From<&'a UniqueLinkedHashMap<K, V>> for &'a LinkedHashMap<K, V>
where
    K: Hash + Eq + PartialEq + Display,
{
    fn from(ulhm: &'a UniqueLinkedHashMap<K, V>) -> Self {
        &ulhm.0
    }
}

impl<K, V> From<LinkedHashMap<K, V>> for UniqueLinkedHashMap<K, V>
where
    K: Hash + Eq + PartialEq + Display,
{
    fn from(lhm: LinkedHashMap<K, V>) -> Self {
        Self(lhm)
    }
}

#[macro_export]
macro_rules! unique_linked_hash_map {
	($($key:expr => $val:expr),* $(,)?) => {{
            #[allow(unused_mut)]
            let mut out = mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap::new();
            $(
                out.insert($key, $val)?;
            )*
            out
	}};
}
