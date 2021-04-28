use std::collections::HashMap;

pub type BindingTuple<T> = HashMap<Key, T>;

#[allow(dead_code)]
#[derive(PartialEq, Eq, Hash, Debug, Clone)]
pub struct Key {
    datasource: DatasourceName,
    scope: u16,
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
#[derive(PartialEq, Eq, Hash, Debug, Clone)]
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
