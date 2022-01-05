use crate::schema::Schema;
use std::collections::BTreeMap;

#[derive(Debug, Ord, PartialOrd, Eq, PartialEq, Clone)]
pub struct Namespace {
    pub db: String,
    pub collection: String,
}

impl<D, S> From<(D, S)> for Namespace
where
    D: Into<String>,
    S: Into<String>,
{
    fn from(tup: (D, S)) -> Self {
        let (db, collection) = tup;
        Self {
            db: db.into(),
            collection: collection.into(),
        }
    }
}

#[derive(Debug, PartialEq, Default)]
pub struct Catalog {
    schemas: BTreeMap<Namespace, Schema>,
}

impl Catalog {
    pub fn new(schemas: BTreeMap<Namespace, Schema>) -> Catalog {
        Catalog { schemas }
    }

    pub fn get_schema_for_namespace(&self, namespace: &Namespace) -> Option<&Schema> {
        self.schemas.get(namespace)
    }
}

impl FromIterator<(Namespace, Schema)> for Catalog {
    fn from_iter<I: IntoIterator<Item = (Namespace, Schema)>>(iter: I) -> Self {
        let mut c = Catalog {
            schemas: BTreeMap::new(),
        };
        for (k, v) in iter {
            c.schemas.insert(k, v);
        }
        c
    }
}
