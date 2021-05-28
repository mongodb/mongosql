use std::collections::BTreeMap;

pub type BindingTuple<T> = BTreeMap<Key, T>;

#[allow(dead_code)]
#[derive(PartialEq, Eq, Hash, PartialOrd, Ord, Debug, Clone)]
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
#[derive(PartialEq, Eq, Hash, PartialOrd, Ord, Debug, Clone)]
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
