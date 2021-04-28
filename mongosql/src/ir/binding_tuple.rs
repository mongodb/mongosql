use std::collections::HashMap;

pub type BindingTuple<T> = HashMap<Key, T>;

#[allow(dead_code)]
#[derive(PartialEq, Eq, Hash, Debug, Clone)]
pub struct Key {
    datasource: DatasourceName,
    scope: u16,
}

#[allow(dead_code)]
#[derive(PartialEq, Eq, Hash, Debug, Clone)]
pub enum DatasourceName {
    Bottom,
    Named(String),
}
