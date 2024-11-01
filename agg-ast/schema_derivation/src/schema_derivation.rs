use crate::{get_schema_for_path_mut, Error, Result};
use agg_ast::definitions::{
    Expression, LiteralValue, Ref, Stage, TaggedOperator, UntaggedOperator,
};
use mongosql::schema::{Atomic, Document, Satisfaction, Schema};
use std::collections::{BTreeMap, BTreeSet};

#[allow(dead_code)]
pub(crate) trait DeriveSchema {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema>;
}

#[allow(dead_code)]
pub(crate) struct ResultSetState<'a> {
    pub catalog: &'a BTreeMap<String, Schema>,
    pub variables: &'a BTreeMap<String, Schema>,
    pub result_set_schema: Schema,
}

impl DeriveSchema for Stage {
    fn derive_schema(&self, _state: &mut ResultSetState) -> Result<Schema> {
        match *self {
            Stage::AddFields(_) => todo!(),
            Stage::AtlasSearchStage(_) => todo!(),
            Stage::Bucket(_) => todo!(),
            Stage::BucketAuto(_) => todo!(),
            Stage::Collection(_) => todo!(),
            Stage::Count(_) => todo!(),
            Stage::Densify(_) => todo!(),
            Stage::Documents(_) => todo!(),
            Stage::EquiJoin(_) => todo!(),
            Stage::Facet(_) => todo!(),
            Stage::Fill(_) => todo!(),
            Stage::GeoNear(_) => todo!(),
            Stage::GraphLookup(_) => todo!(),
            Stage::Group(_) => todo!(),
            Stage::Join(_) => todo!(),
            Stage::Limit(_) => todo!(),
            Stage::Lookup(_) => todo!(),
            Stage::Match(_) => todo!(),
            Stage::Project(_) => todo!(),
            Stage::Redact(_) => todo!(),
            Stage::ReplaceWith(_) => todo!(),
            Stage::Sample(_) => todo!(),
            Stage::SetWindowFields(_) => todo!(),
            Stage::Skip(_) => todo!(),
            Stage::Sort(_) => todo!(),
            Stage::SortByCount(_) => todo!(),
            Stage::UnionWith(_) => todo!(),
            Stage::Unset(_) => todo!(),
            Stage::Unwind(_) => todo!(),
        }
    }
}

fn derive_schema_for_literal(literal_value: &LiteralValue) -> Result<Schema> {
    match literal_value {
        LiteralValue::Binary(_) => Ok(Schema::Atomic(Atomic::BinData)),
        LiteralValue::Boolean(_) => Ok(Schema::Atomic(Atomic::Boolean)),
        LiteralValue::DateTime(_) => Ok(Schema::Atomic(Atomic::Date)),
        LiteralValue::DbPointer(_) => Ok(Schema::Atomic(Atomic::DbPointer)),
        LiteralValue::Decimal128(_) => Ok(Schema::Atomic(Atomic::Decimal)),
        LiteralValue::Double(_) => Ok(Schema::Atomic(Atomic::Double)),
        LiteralValue::Int32(_) => Ok(Schema::Atomic(Atomic::Integer)),
        LiteralValue::Int64(_) => Ok(Schema::Atomic(Atomic::Long)),
        LiteralValue::JavaScriptCode(_) => Ok(Schema::Atomic(Atomic::Javascript)),
        LiteralValue::JavaScriptCodeWithScope(_) => Ok(Schema::Atomic(Atomic::JavascriptWithScope)),
        LiteralValue::MaxKey => Ok(Schema::Atomic(Atomic::MaxKey)),
        LiteralValue::MinKey => Ok(Schema::Atomic(Atomic::MinKey)),
        LiteralValue::Null => Ok(Schema::Atomic(Atomic::Null)),
        LiteralValue::ObjectId(_) => Ok(Schema::Atomic(Atomic::ObjectId)),
        LiteralValue::RegularExpression(_) => Ok(Schema::Atomic(Atomic::Regex)),
        LiteralValue::String(_) => Ok(Schema::Atomic(Atomic::String)),
        LiteralValue::Symbol(_) => Ok(Schema::Atomic(Atomic::Symbol)),
        LiteralValue::Timestamp(_) => Ok(Schema::Atomic(Atomic::Timestamp)),
        LiteralValue::Undefined => Err(Error::InvalidLiteralType),
    }
}

impl DeriveSchema for Expression {
    fn derive_schema(&self, state: &mut ResultSetState) -> Result<Schema> {
        match self {
            Expression::Array(ref a) => Ok(Schema::Array(Box::new(Schema::AnyOf(
                a.iter()
                    .map(|e| {
                        e.derive_schema(state)
                            .map(|schema| schema.upconvert_missing_to_null())
                    })
                    .collect::<Result<BTreeSet<_>>>()?,
            )))),
            Expression::Document(d) => {
                let (mut keys, mut required) = (BTreeMap::new(), BTreeSet::new());
                for (key, e) in d.iter() {
                    let key_schema = e.derive_schema(state)?;
                    match key_schema.satisfies(&Schema::Missing) {
                        Satisfaction::Not => {
                            required.insert(key.clone());
                            keys.insert(key.clone(), key_schema);
                        }
                        Satisfaction::May => {
                            keys.insert(key.clone(), key_schema);
                        }
                        Satisfaction::Must => (),
                    }
                }
                Ok(Schema::Document(Document {
                    keys,
                    required,
                    ..Default::default()
                }))
            }
            Expression::Literal(ref l) => derive_schema_for_literal(l),
            Expression::Ref(Ref::FieldRef(f)) => {
                let path = f.split(".").map(|s| s.to_string()).collect::<Vec<String>>();
                let schema = get_schema_for_path_mut(&mut state.result_set_schema, path);
                match schema {
                    Some(schema) => Ok(schema.clone()),
                    None => Err(Error::UnknownReference(f.clone())),
                }
            }
            Expression::Ref(Ref::VariableRef(v)) => match state.variables.get(v) {
                Some(schema) => Ok(schema.clone()),
                None => Err(Error::UnknownReference(v.clone())),
            },
            Expression::TaggedOperator(op) => op.derive_schema(state),
            Expression::UntaggedOperator(op) => op.derive_schema(state),
        }
    }
}

impl DeriveSchema for TaggedOperator {
    fn derive_schema(&self, _state: &mut ResultSetState) -> Result<Schema> {
        match self {
            TaggedOperator::Accumulator(_) => todo!(),
            TaggedOperator::Bottom(_) => todo!(),
            TaggedOperator::BottomN(_) => todo!(),
            TaggedOperator::Convert(_) => todo!(),
            TaggedOperator::DenseRank(_) => todo!(),
            TaggedOperator::Derivative(_) => todo!(),
            TaggedOperator::DocumentNumber(_) => todo!(),
            TaggedOperator::ExpMovingAvg(_) => todo!(),
            TaggedOperator::Filter(_) => todo!(),
            TaggedOperator::FirstN(_) => todo!(),
            TaggedOperator::Function(_) => todo!(),
            TaggedOperator::GetField(_) => todo!(),
            TaggedOperator::Integral(_) => todo!(),
            TaggedOperator::LastN(_) => todo!(),
            TaggedOperator::Let(_) => todo!(),
            TaggedOperator::Like(_) => todo!(),
            TaggedOperator::Map(_) => todo!(),
            TaggedOperator::MaxNArrayElement(_) => todo!(),
            TaggedOperator::Median(_) => todo!(),
            TaggedOperator::MinNArrayElement(_) => todo!(),
            TaggedOperator::Percentile(_) => todo!(),
            TaggedOperator::Rank(_) => todo!(),
            TaggedOperator::Reduce(_) => todo!(),
            TaggedOperator::Regex(_) => todo!(),
            TaggedOperator::RegexFind(_) => todo!(),
            TaggedOperator::RegexFindAll(_) => todo!(),
            TaggedOperator::ReplaceAll(_) => todo!(),
            TaggedOperator::ReplaceOne(_) => todo!(),
            TaggedOperator::SetField(_) => todo!(),
            TaggedOperator::Shift(_) => todo!(),
            TaggedOperator::SortArray(_) => todo!(),
            TaggedOperator::Subquery(_) => todo!(),
            TaggedOperator::SubqueryComparison(_) => todo!(),
            TaggedOperator::SubqueryExists(_) => todo!(),
            TaggedOperator::Switch(_) => todo!(),
            TaggedOperator::Top(_) => todo!(),
            TaggedOperator::TopN(_) => todo!(),
            TaggedOperator::UnsetField(_) => todo!(),
            TaggedOperator::LTrim(_) => todo!(),
            TaggedOperator::RTrim(_) => todo!(),
            TaggedOperator::Trim(_) => todo!(),
            TaggedOperator::Hour(_) => todo!(),
            TaggedOperator::Minute(_) => todo!(),
            TaggedOperator::Second(_) => todo!(),
            TaggedOperator::Millisecond(_) => todo!(),
            TaggedOperator::DayOfWeek(_) => todo!(),
            TaggedOperator::DayOfMonth(_) => todo!(),
            TaggedOperator::DayOfYear(_) => todo!(),
            TaggedOperator::IsoDayOfWeek(_) => todo!(),
            TaggedOperator::IsoWeek(_) => todo!(),
            TaggedOperator::IsoWeekYear(_) => todo!(),
            TaggedOperator::Week(_) => todo!(),
            TaggedOperator::Month(_) => todo!(),
            TaggedOperator::Year(_) => todo!(),
            TaggedOperator::DateFromParts(_) => todo!(),
            TaggedOperator::DateFromString(_) => todo!(),
            TaggedOperator::DateToParts(_) => todo!(),
            TaggedOperator::DateToString(_) => todo!(),
            TaggedOperator::DateAdd(_) => todo!(),
            TaggedOperator::DateSubtract(_) => todo!(),
            TaggedOperator::DateDiff(_) => todo!(),
            TaggedOperator::DateTrunc(_) => todo!(),
            TaggedOperator::Zip(_) => todo!(),
            TaggedOperator::SqlConvert(_) | TaggedOperator::SqlDivide(_) => {
                Err(Error::InvalidTaggedOperator(self.clone()))
            }
        }
    }
}

impl DeriveSchema for UntaggedOperator {
    fn derive_schema(&self, _state: &mut ResultSetState) -> Result<Schema> {
        Err(Error::InvalidUntaggedOperator(self.op.clone()))
    }
}
