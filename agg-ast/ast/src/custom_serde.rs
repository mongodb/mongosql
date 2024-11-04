use crate::{
    definitions::{
        DateExpression, Expression, GroupAccumulator, GroupAccumulatorExpr, LiteralValue,
        MatchArrayExpression, MatchArrayQuery, MatchBinaryOp, MatchElement, MatchExpression,
        MatchField, MatchNot, MatchNotExpression, MatchRegex, MatchStage, ProjectItem,
        ProjectStage, Ref, SetWindowFieldsOutput, UntaggedOperator, VecOrSingleExpr, Window,
    },
    map,
};
use bson::{self, doc, Bson, Document};
use linked_hash_map::LinkedHashMap;
use serde::{
    de::{self, Deserialize, Deserializer, Error as serde_err, MapAccess, Visitor},
    ser::{self, SerializeMap},
};
use std::{fmt, sync::LazyLock};

static DECIMAL_ZERO: LazyLock<bson::Decimal128> = LazyLock::new(|| "0.0".parse().unwrap());

#[derive(Default)]
struct MatchStageVisitor {
    marker: std::marker::PhantomData<MatchStage>,
}

#[derive(Default)]
struct MatchArrayQueryVisitor {
    marker: std::marker::PhantomData<MatchArrayQuery>,
}

#[derive(Default)]
struct ProjectStageVisitor {
    marker: std::marker::PhantomData<ProjectStage>,
}

#[derive(Default)]
struct MatchNotVisitor {
    marker: std::marker::PhantomData<MatchNot>,
}

#[derive(Default)]
struct MatchElementVisitor {
    marker: std::marker::PhantomData<MatchElement>,
}

#[derive(Default)]
struct MatchRegexVisitor {
    marker: std::marker::PhantomData<MatchRegex>,
}

#[derive(Default)]
struct MatchFieldVisitor {
    marker: std::marker::PhantomData<MatchField>,
}

fn get_single_entry(doc: &bson::Document) -> Option<(String, bson::Bson)> {
    if let Some((key, value)) = doc.into_iter().next() {
        return Some((key.clone(), value.clone()));
    }
    None
}

impl TryFrom<&str> for MatchBinaryOp {
    type Error = String;
    fn try_from(s: &str) -> Result<Self, String> {
        let deserialized: Self =
            bson::from_bson(bson::Bson::String(s.to_string())).map_err(|e| e.to_string())?;
        Ok(deserialized)
    }
}

impl From<MatchBinaryOp> for String {
    fn from(o: MatchBinaryOp) -> Self {
        let serialized: Bson =
            bson::to_bson(&o).expect("failed to serialize, this is a code error.");
        match serialized {
            bson::Bson::String(s) => s,
            _ => unreachable!(),
        }
    }
}

impl MatchArrayQueryVisitor {
    fn new() -> Self {
        MatchArrayQueryVisitor {
            marker: std::marker::PhantomData,
        }
    }
}

impl<'de> de::Visitor<'de> for MatchArrayQueryVisitor {
    // The type that our Visitor is going to produce.
    type Value = MatchArrayQuery;
    // Deserialize MatchArrayQuery from an abstract "map" provided by the
    // Deserializer. The MapVisitor input is a callback provided by the Deserializer to let us see
    // each entry in the map.
    fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
    where
        M: de::MapAccess<'de>,
    {
        // If the size_hint is None, we assume that there will be at least one key pair.
        let mut values = MatchArrayQuery::with_capacity(access.size_hint().unwrap_or(1));

        while let Some((key, value)) = access.next_entry()? {
            let key: String = key;
            let value: bson::Bson = value;
            let mut sub_expr: Document = Document::new();
            sub_expr.insert(key, value);
            let deserialized: MatchExpression = bson::from_bson(bson::Bson::Document(sub_expr))
                .expect("failed to deserialize, this is a code error.");
            values.push(deserialized)
        }
        Ok(values)
    }

    // This is called, if an unexpected type is hit.
    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        write!(formatter, "a map or `null`")
    }
}

// This is the trait that informs Serde how to deserialize MatchArrayQuery.
impl<'de> de::Deserialize<'de> for MatchArrayQuery {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: de::Deserializer<'de>,
    {
        // Instantiate our Visitor and ask the Deserializer to drive
        // it over the input data, resulting in an instance of MatchArrayQuery.
        deserializer.deserialize_map(MatchArrayQueryVisitor::new())
    }
}

// This is the trait that informs Serde how to serialize MatchArrayQuery.
impl ser::Serialize for MatchArrayQuery {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
        S::Error: ser::Error,
    {
        let mut map = serializer.serialize_map(Some(self.len()))?;
        for v in self.query.iter() {
            let serialized: Bson =
                bson::to_bson(&v).expect("failed to serialize, this is a code error.");
            match serialized {
                bson::Bson::Document(doc) => {
                    let (name, expr) = get_single_entry(&doc).unwrap();
                    map.serialize_entry(name.as_str(), &expr)?
                }
                _ => {
                    return Err(ser::Error::custom(format_args!(
                        "expected Bson::Document during serialization of MatchArrayQuery items, found {:?}",
                        serialized
                    )));
                }
            }
        }
        map.end()
    }
}

impl<'de> de::Visitor<'de> for MatchNotVisitor {
    // The type that our Visitor is going to produce.
    type Value = MatchNot;

    // Deserialize MatchNot from an abstract "map" provided by the
    // Deserializer. The MapVisitor input is a callback provided by the Deserializer to let us see
    // each entry in the map.
    fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
    where
        M: de::MapAccess<'de>,
        M::Error: de::Error,
    {
        if let Some(len) = access.size_hint() {
            if len != 1 {
                return Err(de::Error::custom(format_args!(
                    "invalid map length: {}, expected exactly 1 key-value pair in $not context",
                    len
                )));
            }
        } else {
            return Err(de::Error::custom(format_args!(
                "match expression is not a MatchNot"
            )));
        }
        // unwrap is safe because we know there is exactly one key.
        let (field, value): (String, bson::Bson) = access.next_entry()?.unwrap();
        let field = Ref::FieldRef(field);
        match value {
            Bson::Document(doc) => {
                if doc.len() != 1 {
                    return Err(de::Error::custom(format_args!(
                        "invalid map length: {}, expected exactly 1 key-value pair in $not expression context",
                        doc.len()
                    )));
                }
                let (op, value) = get_single_entry(&doc).unwrap();
                if op != "$not" {
                    return Err(de::Error::custom(format_args!(
                        "invalid key: {}, expected $not in $not expression context",
                        op
                    )));
                }
                let expr: MatchNotExpression = bson::from_bson(value).map_err(de::Error::custom)?;
                Ok(MatchNot { field, expr })
            }
            _ => Err(de::Error::custom(format_args!(
                "expected a map in MatchNot context, found {:?}",
                value
            ))),
        }
    }

    // This is called if an unexpected type is hit.
    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        write!(formatter, "a map")
    }
}

impl<'de> de::Deserialize<'de> for MatchNot {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: de::Deserializer<'de>,
    {
        // Instantiate our Visitor and ask the Deserializer to drive
        // it over the input data, resulting in an instance of MatchNot.
        deserializer.deserialize_map(MatchNotVisitor::default())
    }
}

// This is the trait that informs Serde how to serialize MatchNot.
impl ser::Serialize for MatchNot {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
    {
        let mut map = serializer.serialize_map(Some(1))?;
        let expr: LinkedHashMap<_, _> = map! {"$not" => &self.expr};
        map.serialize_entry(self.field.as_str(), &expr)?;
        map.end()
    }
}

impl<'de> de::Visitor<'de> for MatchElementVisitor {
    // The type that our Visitor is going to produce.
    type Value = MatchElement;

    // Deserialize MatchElement from an abstract "map" provided by the
    // Deserializer. The MapVisitor input is a callback provided by the Deserializer to let us see
    // each entry in the map.
    fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
    where
        M: de::MapAccess<'de>,
        M::Error: de::Error,
    {
        if let Some(len) = access.size_hint() {
            if len != 1 {
                return Err(de::Error::custom(format_args!(
                    "invalid map length: {}, expected exactly 1 key-value pair in MatchElement expression context",
                    len
                )));
            }
        } else {
            return Err(de::Error::custom(format_args!(
                "match expression is not a MatchElement"
            )));
        }
        // unwrap is safe because we know there is exactly one key.
        let (field, value): (String, bson::Bson) = access.next_entry()?.unwrap();
        let field = Ref::FieldRef(field);
        match value {
            Bson::Document(doc) => {
                if doc.len() != 1 {
                    return Err(de::Error::custom(format_args!(
                        "invalid map length: {}, expected exactly 1 key-value pair in $elemMatch expression context",
                        doc.len()
                    )));
                }
                let (op, value) = get_single_entry(&doc).unwrap();
                if op != "$elemMatch" {
                    return Err(de::Error::custom(format_args!(
                        "invalid key: {}, expected $elemMatch in $elemMatch expression context",
                        op
                    )));
                }
                let query: MatchArrayExpression =
                    bson::from_bson(value).map_err(de::Error::custom)?;
                Ok(MatchElement { field, query })
            }
            _ => Err(de::Error::custom(format_args!(
                "expected a map in MatchElement context, found {:?}",
                value
            ))),
        }
    }

    // This is called if an unexpected type is hit.
    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        write!(formatter, "a map")
    }
}

// This is the trait that informs Serde how to deserialize MatchElement.
impl<'de> de::Deserialize<'de> for MatchElement {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: de::Deserializer<'de>,
    {
        // Instantiate our Visitor and ask the Deserializer to drive
        // it over the input data, resulting in an instance of MatchElement.
        deserializer.deserialize_map(MatchElementVisitor::default())
    }
}

// This is the trait that informs Serde how to serialize MatchElement.
impl ser::Serialize for MatchElement {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
    {
        let mut map = serializer.serialize_map(Some(1))?;
        let expr: LinkedHashMap<_, _> = map! {"$elemMatch" => &self.query};
        map.serialize_entry(self.field.as_str(), &expr)?;
        map.end()
    }
}

impl<'de> de::Visitor<'de> for MatchRegexVisitor {
    // The type that our Visitor is going to produce.
    type Value = MatchRegex;

    // Deserialize MatchRegex from an abstract "map" provided by the
    // Deserializer. The MapVisitor input is a callback provided by the Deserializer to let us see
    // each entry in the map.
    fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
    where
        M: de::MapAccess<'de>,
    {
        if let Some(len) = access.size_hint() {
            if len != 1 {
                return Err(de::Error::custom(format_args!(
                    "invalid map length: {}, expected exactly 1 key-value pair in MatchRegex expression context",
                    len
                )));
            }
        } else {
            return Err(de::Error::custom(format_args!(
                "match expression is not a MatchRegex"
            )));
        }
        // unwrap is safe because we know there is exactly one key.
        let (field, value): (String, bson::Bson) = access.next_entry()?.unwrap();
        let field = Ref::FieldRef(field);
        match value {
            Bson::Document(mut doc) => {
                if let Some(pattern) = doc.remove("$regex") {
                    let mut ret = MatchRegex {
                        field,
                        pattern,
                        options: None,
                    };
                    ret.options = doc.remove("$options");
                    Ok(ret)
                } else {
                    Err(de::Error::custom(format_args!(
                        "expected $regex key in MatchRegex context, found {:?}",
                        doc
                    )))
                }
            }
            _ => Err(de::Error::custom(format_args!(
                "expected a map in MatchRegex context, found {:?}",
                value
            ))),
        }
    }

    // This is called if an unexpected type is hit.
    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        write!(formatter, "a map")
    }
}

// This is the trait that informs Serde how to deserialize MatchRegex.
impl<'de> de::Deserialize<'de> for MatchRegex {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: de::Deserializer<'de>,
    {
        // Instantiate our Visitor and ask the Deserializer to drive
        // it over the input data, resulting in an instance of MatchRegex.
        deserializer.deserialize_map(MatchRegexVisitor::default())
    }
}

// This is the trait that informs Serde how to serialize MatchRegex.
impl ser::Serialize for MatchRegex {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
        S::Error: ser::Error,
    {
        let mut map = serializer.serialize_map(Some(1))?;
        let expr: LinkedHashMap<_, _> = match self.options {
            Some(ref options) => {
                map! {"$regex" => self.pattern.clone(), "$options" =>  options.clone()}
            }
            None => map! {"$regex" => self.pattern.clone()},
        };
        map.serialize_entry(self.field.as_str(), &expr)?;
        map.end()
    }
}

impl<'de> de::Visitor<'de> for MatchFieldVisitor {
    // The type that our Visitor is going to produce.
    type Value = MatchField;

    // Deserialize MatchField from an abstract "map" provided by the
    // Deserializer. The MapVisitor input is a callback provided by the Deserializer to let us see
    // each entry in the map.
    fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
    where
        M: de::MapAccess<'de>,
        M::Error: de::Error,
    {
        if let Some(len) = access.size_hint() {
            if len != 1 {
                return Err(de::Error::custom(format_args!(
                    "invalid map length: {}, expected exactly 1 key-value pair in MatchField expression context",
                    len
                )));
            }
        } else {
            return Err(de::Error::custom(format_args!(
                "match expression is not a MatchField"
            )));
        }
        // unwrap is safe because we know there is exactly one key.
        let (field, value): (String, bson::Bson) = access.next_entry()?.unwrap();
        let field = Ref::FieldRef(field);
        match value {
            Bson::Document(doc) => {
                let mut ops: LinkedHashMap<MatchBinaryOp, bson::Bson> = LinkedHashMap::new();
                for (key, value) in doc.into_iter() {
                    let key: String = key;
                    let value: bson::Bson = value;
                    ops.insert(key.as_str().try_into().map_err(de::Error::custom)?, value);
                }
                Ok(MatchField { field, ops })
            }
            // If this isn't a map, it must be a direct single equality
            _ => Ok(MatchField {
                field,
                ops: map! {MatchBinaryOp::Eq => value},
            }),
        }
    }

    // This is called if an unexpected type is hit.
    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        write!(formatter, "a map")
    }
}

// This is the trait that informs Serde how to deserialize MatchField.
impl<'de> de::Deserialize<'de> for MatchField {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: de::Deserializer<'de>,
    {
        // Instantiate our Visitor and ask the Deserializer to drive
        // it over the input data, resulting in an instance of MatchField.
        deserializer.deserialize_map(MatchFieldVisitor::default())
    }
}

// This is the trait that informs Serde how to serialize MatchField.
impl ser::Serialize for MatchField {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
    {
        let mut map = serializer.serialize_map(Some(1))?;
        map.serialize_entry(self.field.as_str(), &self.ops)?;
        map.end()
    }
}

impl MatchStageVisitor {
    fn new() -> Self {
        MatchStageVisitor {
            marker: std::marker::PhantomData,
        }
    }
}

impl<'de> de::Visitor<'de> for MatchStageVisitor {
    // The type that our Visitor is going to produce.
    type Value = MatchStage;

    // Deserialize MatchStage from an abstract "map" provided by the
    // Deserializer. The MapVisitor input is a callback provided by the Deserializer to let us see
    // each entry in the map.
    fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
    where
        M: de::MapAccess<'de>,
    {
        // If the size_hint is None, we assume that there will be at least one key pair.
        let mut values = MatchStage::with_capacity(access.size_hint().unwrap_or(1));

        while let Some((key, value)) = access.next_entry()? {
            let key: String = key;
            let value: bson::Bson = value;
            let mut sub_expr: Document = Document::new();
            sub_expr.insert(key, value);
            let deserialized: MatchExpression = bson::from_bson(bson::Bson::Document(sub_expr))
                .expect("failed to deserialize, this is a code error.");
            values.push(deserialized)
        }

        Ok(values)
    }

    // This is called if an unexpected type is hit.
    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        write!(formatter, "a map or `null`")
    }
}

// This is the trait that informs Serde how to deserialize MatchStage.
impl<'de> de::Deserialize<'de> for MatchStage {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: de::Deserializer<'de>,
    {
        // Instantiate our Visitor and ask the Deserializer to drive
        // it over the input data, resulting in an instance of MatchStage.
        deserializer.deserialize_map(MatchStageVisitor::new())
    }
}

// This is the trait that informs Serde how to serialize MatchStage.
impl ser::Serialize for MatchStage {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
        S::Error: ser::Error,
    {
        let mut map = serializer.serialize_map(Some(self.len()))?;
        for v in self.expr.iter() {
            let serialized: Bson =
                bson::to_bson(&v).expect("failed to serialize, this is a code error.");
            match serialized {
                bson::Bson::Document(doc) => {
                    let (name, expr) = get_single_entry(&doc).unwrap();
                    map.serialize_entry(name.as_str(), &expr)?
                }
                _ => {
                    return Err(ser::Error::custom(format_args!(
                        "expected Bson::Document during serialization of MatchStage items, found {:?}",
                        serialized
                    )));
                }
            }
        }
        map.end()
    }
}

impl From<Expression> for ProjectItem {
    fn from(e: Expression) -> Self {
        match e {
            Expression::Literal(lit) => match lit {
                LiteralValue::Boolean(b) => {
                    if !b {
                        ProjectItem::Exclusion
                    } else {
                        ProjectItem::Inclusion
                    }
                }
                LiteralValue::Int32(i) => {
                    if i == 0 {
                        ProjectItem::Exclusion
                    } else {
                        ProjectItem::Inclusion
                    }
                }
                LiteralValue::Int64(i) => {
                    if i == 0 {
                        ProjectItem::Exclusion
                    } else {
                        ProjectItem::Inclusion
                    }
                }
                LiteralValue::Double(d) => {
                    if d == 0.0 {
                        ProjectItem::Exclusion
                    } else {
                        ProjectItem::Inclusion
                    }
                }
                LiteralValue::Decimal128(d) => {
                    if d == *DECIMAL_ZERO {
                        ProjectItem::Exclusion
                    } else {
                        ProjectItem::Inclusion
                    }
                }
                lit => ProjectItem::Assignment(Expression::Literal(lit)),
            },
            x => ProjectItem::Assignment(x),
        }
    }
}

impl<'de> de::Visitor<'de> for ProjectStageVisitor {
    // The type that our Visitor is going to produce.
    type Value = ProjectStage;

    // Deserialize ProjectStage from an abstract "map" provided by the
    // Deserializer. The MapVisitor input is a callback provided by the Deserializer to let us see
    // each entry in the map.
    fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
    where
        M: de::MapAccess<'de>,
    {
        // If the size_hint is None, we assume that there will be at least one key pair.
        let mut values = ProjectStage::with_capacity(access.size_hint().unwrap_or(1));

        while let Some((key, value)) = access.next_entry()? {
            let key: String = key;
            let value: bson::Bson = value;
            let deserialize: Expression =
                bson::from_bson(value).expect("failed to deserialize, this is a code error.");
            let value: ProjectItem = deserialize.into();
            values.push((key, value))
        }

        Ok(values)
    }

    // This is called if an unexpected type is hit.
    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        write!(formatter, "a map or `null`")
    }
}

// This is the trait that informs Serde how to deserialize ProjectStage.
impl<'de> de::Deserialize<'de> for ProjectStage {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: de::Deserializer<'de>,
    {
        // Instantiate our Visitor and ask the Deserializer to drive
        // it over the input data, resulting in an instance of ProjectStage.
        deserializer.deserialize_map(ProjectStageVisitor::default())
    }
}

// This is the trait that informs Serde how to serialize ProjectStage.
impl ser::Serialize for ProjectStage {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
        S::Error: ser::Error,
    {
        let mut map = serializer.serialize_map(Some(self.len()))?;
        for v in self.items.iter() {
            match v {
                (key, ProjectItem::Assignment(expr)) => {
                    map.serialize_entry(key, &expr)?;
                }
                (key, ProjectItem::Inclusion) => {
                    map.serialize_entry(key, &true)?;
                }
                (key, ProjectItem::Exclusion) => {
                    map.serialize_entry(key, &false)?;
                }
            }
        }
        map.end()
    }
}

struct RefVisitor {
    marker: std::marker::PhantomData<Ref>,
}

impl RefVisitor {
    fn new() -> Self {
        RefVisitor {
            marker: std::marker::PhantomData,
        }
    }
}

// Deserialization for Ref adapted from an out-of-date tutorial:
// https://riptutorial.com/rust/example/16175/implement-deserialize-for-a-custom-map-type
// This is the trait that Deserializers are going to be driving.
impl<'de> de::Visitor<'de> for RefVisitor {
    // The type that our Visitor is going to produce.
    type Value = Ref;

    fn visit_str<E>(self, s: &str) -> Result<Self::Value, E>
    where
        E: de::Error,
    {
        let bytes = s.as_bytes();
        if !bytes.is_empty() && bytes[0] == b'$' {
            if bytes.len() > 1 && bytes[1] == b'$' {
                Ok(Ref::VariableRef(
                    std::str::from_utf8(&bytes[2..]).unwrap().into(),
                ))
            } else {
                Ok(Ref::FieldRef(
                    std::str::from_utf8(&bytes[1..]).unwrap().into(),
                ))
            }
        } else {
            Err(de::Error::invalid_value(de::Unexpected::Str(s), &self))
        }
    }

    // This is called if an unexepected type is hit.
    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        write!(formatter, "a string beginning with '$'")
    }
}

// This is the trait that informs Serde how to deserialize Ref.
impl<'de> de::Deserialize<'de> for Ref {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: de::Deserializer<'de>,
    {
        // Instantiate our Visitor and ask the Deserializer to drive
        // it over the input data, resulting in an instance of Ref.
        deserializer.deserialize_str(RefVisitor::new())
    }
}

impl ser::Serialize for Ref {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
    {
        let (prefix, s): (String, &str) = match self {
            Ref::FieldRef(s) => ("$".into(), s),
            Ref::VariableRef(s) => ("$$".into(), s),
        };
        serializer.serialize_str(&(prefix + s))
    }
}

impl<'de> Deserialize<'de> for SetWindowFieldsOutput {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        /// Custom map visitor for identifying and deserializing Window Operators.
        struct SetWindowFieldsOutputVisitor;

        impl<'de> Visitor<'de> for SetWindowFieldsOutputVisitor {
            type Value = SetWindowFieldsOutput;

            fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
                formatter.write_str(r#"{"$op": <expression>, "window": {"documents": [<lower boundary>, <upper boundary>], "range": [<lower boundary>, <upper boundary>], "unit": <time unit>}}"#)
            }

            fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
            where
                M: MapAccess<'de>,
            {
                // The $setWindowFields "output" field looks like:
                // {
                //    <output field 1>: {
                //        <"$"-prefixed window operator>: <window operator parameters>,
                //        window: {
                //            documents: [ <lower boundary>, <upper boundary> ],
                //            range: [ <lower boundary>, <upper boundary> ],
                //            unit: <time unit>,
                //        },
                //    },
                //    <output field 2>: { ... },
                //    ...
                //    <output field n>: { ... }
                // }
                // This function parses a single output field value, meaning a document with the
                // variable '$'-prefixed window operator and "window" fields. The "window" field is
                // optional so a second kv pair may not be present.
                let kv1 = access.next_entry::<String, Bson>()?;
                let kv2 = access.next_entry::<String, Bson>()?;

                // Figure out which is the window operator and which is the "window"
                let (window_func_key, window_func_val, window_val) = match (kv1, kv2) {
                    (Some((key1, value1)), Some((key2, value2))) => {
                        let key1_is_window_func = key1.starts_with('$');
                        let key2_is_window_func = key2.starts_with('$');

                        if key1_is_window_func && key2_is_window_func {
                            let err_msg =
                                format!("multiple window operators found: '{key1}' and '{key2}'");
                            return Err(serde_err::custom(err_msg));
                        } else if !key1_is_window_func && !key2_is_window_func {
                            return Err(serde_err::custom("no window operator found"));
                        } else if key1_is_window_func {
                            if key2 != "window" {
                                let err_msg = format!("expected key 'window' but found '{key2}'");
                                return Err(serde_err::custom(err_msg));
                            }
                            (key1, value1, Some(value2))
                        } else {
                            if key1 != "window" {
                                let err_msg = format!("expected key 'window' but found '{key1}'");
                                return Err(serde_err::custom(err_msg));
                            }
                            (key2, value2, Some(value1))
                        }
                    }
                    (Some((key, value)), None) => {
                        if !key.starts_with('$') {
                            let err_msg = format!(
                                "expected window operator key to start with '$' but found '{key}'"
                            );
                            return Err(serde_err::custom(err_msg));
                        }
                        (key, value, None)
                    }
                    _ => {
                        return Err(serde_err::custom(
                            "setWindowFields output could not be parsed",
                        ))
                    }
                };

                // Parse the window operator into an Expression.
                let window_func_as_doc = doc! {window_func_key: window_func_val};
                let window_func_as_expr: Expression = bson::from_document(window_func_as_doc)
                    .expect("failed to deserialize window function, this is a code error.");

                // Parse the window into a Window.
                let window: Option<Window> = window_val.map(|val| {
                    bson::from_bson(val)
                        .expect("failed to deserialize window, this is a code error")
                });

                Ok(SetWindowFieldsOutput {
                    window_func: Box::new(window_func_as_expr),
                    window,
                })
            }
        }

        const FIELDS: &[&str] = &["window_func", "window"];
        deserializer.deserialize_struct(
            "SetWindowFieldsOutput",
            FIELDS,
            SetWindowFieldsOutputVisitor,
        )
    }
}

/// Custom deserialization function for untagged aggregation operators.
///

/// Custom map visitor for identifying and deserializing UntaggedOperators.
struct UntaggedOperatorVisitor {}

impl UntaggedOperatorVisitor {
    fn new() -> Self {
        Self {}
    }
}

pub fn deserialize_mql_operator<'de, D>(deserializer: D) -> Result<UntaggedOperator, D::Error>
where
    D: Deserializer<'de>,
{
    deserializer.deserialize_map(UntaggedOperatorVisitor::new())
}

pub fn serialize_mql_operator<S>(value: &UntaggedOperator, serializer: S) -> Result<S::Ok, S::Error>
where
    S: ser::Serializer,
{
    let mut map = serializer.serialize_map(Some(1))?;
    map.serialize_entry(&value.op, &value.args)?;
    map.end()
}

impl<'de> Visitor<'de> for UntaggedOperatorVisitor {
    type Value = UntaggedOperator;

    fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
        formatter.write_str("{\"$op\": [args]}")
    }

    fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
    where
        M: MapAccess<'de>,
    {
        let kv = access.next_entry::<String, VecOrSingleExpr>()?;
        if let Some((key, value)) = kv {
            // If the key does not start with a "$", then it is not an agg operator.
            // Ignore this map and stop attempting to deserialize with this function.
            if !key.starts_with('$') {
                return Err(serde_err::custom("ignoring key that does not start with $"));
            }

            // Immediately return when we see one key that starts with a "$".
            // In a general environment, this would be very brittle, however in this
            // controlled test environment, we safely make the assumption that
            // a single key that starts with a "$" is present and indicates an operator.
            return Ok(UntaggedOperator {
                op: key,
                args: value.get_as_vec(),
            });
        }

        Err(serde_err::custom(
            "fail when there are no keys; this lets empty doc be parsed as Document",
        ))
    }
}

impl<'de> Deserialize<'de> for GroupAccumulator {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        /// Custom map visitor for identifying and deserializing Accumulators.
        struct AccumulatorVisitor;

        impl<'de> Visitor<'de> for AccumulatorVisitor {
            type Value = GroupAccumulator;

            fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
                formatter.write_str("{\"$op\": <expression or struct>}")
            }

            fn visit_map<M>(self, mut access: M) -> Result<Self::Value, M::Error>
            where
                M: MapAccess<'de>,
            {
                let kv = access.next_entry::<String, GroupAccumulatorExpr>()?;
                if let Some((key, value)) = kv {
                    // If the key does not start with a "$", then it is not an accumulator function.
                    // Ignore this map and stop attempting to deserialize with this function.
                    if !key.starts_with('$') {
                        return Err(serde_err::custom("ignoring key that does not start with $"));
                    }

                    // Immediately return when we see one key that starts with a "$".
                    // In a general environment, this would be very brittle, however in this
                    // controlled test environment, we safely make the assumption that
                    // a single key that starts with a "$" is present and indicates an operator.
                    // let value = value.get_as_vec();
                    return Ok(GroupAccumulator {
                        function: key,
                        expr: value,
                    });
                }

                Err(serde_err::custom("no accumulator could be parsed"))
            }
        }

        const FIELDS: &[&str] = &["function", "expr"];
        deserializer.deserialize_struct("GroupAccumulator", FIELDS, AccumulatorVisitor)
    }
}

impl ser::Serialize for GroupAccumulator {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
    {
        let mut map = serializer.serialize_map(Some(1))?;
        map.serialize_entry(&self.function, &self.expr)?;
        map.end()
    }
}

impl<'de> Deserialize<'de> for DateExpression {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        // date expressions can take one of two forms;
        // 1. $op: <dateExpression>
        // 2. $op: { date: <dateExpression>, timezone: <timezoneExpression>}
        // deserialize into expression, and try to convert any documents into the latter; otherwise, return the former
        let expression = Expression::deserialize(deserializer)?;
        match expression {
            Expression::Document(mut d) => match d.remove("date") {
                Some(date_expr) => {
                    let date = Box::new(date_expr);
                    let timezone = d.remove("timezone").map(Box::new);
                    if !d.is_empty() {
                        return Err(serde_err::custom("too many arguments for date Document"));
                    }
                    Ok(DateExpression { date, timezone })
                }
                None => Err(serde_err::custom(
                    "document to date operator does not contain field \"date\"",
                )),
            },
            expr => Ok(DateExpression {
                date: Box::new(expr),
                timezone: None,
            }),
        }
    }
}

impl ser::Serialize for DateExpression {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
    {
        if let Some(ref timezone) = self.timezone {
            let mut map = serializer.serialize_map(Some(2))?;
            map.serialize_entry("date", &self.date)?;
            map.serialize_entry("timezone", timezone)?;
            map.end()
        } else {
            self.date.serialize(serializer)
        }
    }
}

// This is the trait that informs Serde how to deserialize LiteralValue.
impl<'de> de::Deserialize<'de> for LiteralValue {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: de::Deserializer<'de>,
    {
        let b: Bson = de::Deserialize::deserialize(deserializer)?;
        Ok(match b {
            Bson::Double(d) => LiteralValue::Double(d),
            Bson::String(s) => LiteralValue::String(s),
            Bson::Array(_) | Bson::Document(_) => {
                return Err(de::Error::custom(format_args!(
                    "expected a literal value, found a document or array"
                )))
            }
            Bson::Boolean(b) => LiteralValue::Boolean(b),
            Bson::Null => LiteralValue::Null,
            Bson::RegularExpression(r) => LiteralValue::RegularExpression(r),
            Bson::JavaScriptCode(c) => LiteralValue::JavaScriptCode(c),
            Bson::JavaScriptCodeWithScope(js) => LiteralValue::JavaScriptCodeWithScope(js),
            Bson::Int32(i) => LiteralValue::Int32(i),
            Bson::Int64(i) => LiteralValue::Int64(i),
            Bson::Timestamp(ts) => LiteralValue::Timestamp(ts),
            Bson::Binary(b) => LiteralValue::Binary(b),
            Bson::ObjectId(oid) => LiteralValue::ObjectId(oid),
            Bson::DateTime(dt) => LiteralValue::DateTime(dt),
            Bson::Symbol(s) => LiteralValue::Symbol(s),
            Bson::Decimal128(d) => LiteralValue::Decimal128(d),
            Bson::Undefined => LiteralValue::Undefined,
            Bson::MaxKey => LiteralValue::MaxKey,
            Bson::MinKey => LiteralValue::MinKey,
            Bson::DbPointer(dp) => LiteralValue::DbPointer(dp),
        })
    }
}

// This is the trait that informs Serde how to serialize LiteralValue.
impl ser::Serialize for LiteralValue {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: ser::Serializer,
        S::Error: ser::Error,
    {
        let b = match self {
            LiteralValue::Double(d) => Bson::Double(*d),
            LiteralValue::String(s) => Bson::String(s.clone()),
            LiteralValue::Boolean(b) => Bson::Boolean(*b),
            LiteralValue::Null => Bson::Null,
            LiteralValue::RegularExpression(r) => Bson::RegularExpression(r.clone()),
            LiteralValue::JavaScriptCode(c) => Bson::JavaScriptCode(c.clone()),
            LiteralValue::JavaScriptCodeWithScope(js) => Bson::JavaScriptCodeWithScope(js.clone()),
            LiteralValue::Int32(i) => Bson::Int32(*i),
            LiteralValue::Int64(i) => Bson::Int64(*i),
            LiteralValue::Timestamp(ts) => Bson::Timestamp(*ts),
            LiteralValue::Binary(b) => Bson::Binary(b.clone()),
            LiteralValue::ObjectId(oid) => Bson::ObjectId(*oid),
            LiteralValue::DateTime(dt) => Bson::DateTime(*dt),
            LiteralValue::Symbol(s) => Bson::Symbol(s.clone()),
            LiteralValue::Decimal128(d) => Bson::Decimal128(*d),
            LiteralValue::Undefined => Bson::Undefined,
            LiteralValue::MaxKey => Bson::MaxKey,
            LiteralValue::MinKey => Bson::MinKey,
            LiteralValue::DbPointer(dp) => Bson::DbPointer(dp.clone()),
        };
        b.serialize(serializer)
    }
}
