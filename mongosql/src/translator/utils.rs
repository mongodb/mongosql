use crate::{
    air::{self, LetVariable, MQLOperator, SQLOperator, TrimOperator},
    mapping_registry::{MqlMappingRegistry, MqlMappingRegistryValue, MqlReferenceType},
    mir,
    translator::{Error, MqlTranslator, Result},
};
use lazy_static::lazy_static;
use mongosql_datastructures::binding_tuple::{BindingTuple, DatasourceName, Key};
use std::collections::BTreeMap;

lazy_static! {
    pub static ref ROOT: air::Expression = air::Expression::Variable(air::Variable {
        parent: None,
        name: "ROOT".into()
    });
}

#[derive(Debug)]
pub(crate) enum ScalarFunctionType {
    Divide,
    Mql(MQLOperator),
    Sql(SQLOperator),
    Trim(TrimOperator),
}

impl MqlTranslator {
    /// Generate a unique bottom name given a predicate closure. Keeps pre-pending
    /// `_` until the predicate returns false, indicating that that name is not in use.
    pub(crate) fn generate_unique_bot_name<F>(name_exists: F) -> String
    where
        F: Fn(&String) -> bool,
    {
        let mut ret = "__bot".to_string();
        while name_exists(&ret) {
            ret.insert(0, '_');
        }
        ret
    }

    pub(crate) fn get_unique_bot_name(project_names: &BindingTuple<mir::Expression>) -> String {
        if project_names.is_empty() {
            return "__bot".to_string();
        }
        let current_scope = project_names.keys().next().unwrap().scope;
        MqlTranslator::generate_unique_bot_name(|s| {
            project_names.contains_key(&(s.clone(), current_scope).into())
        })
    }

    pub(crate) fn get_datasource_name(
        datasource: &DatasourceName,
        unique_bot_name: &str,
    ) -> String {
        match datasource {
            DatasourceName::Bottom => unique_bot_name.to_string(),
            DatasourceName::Named(s) => s.clone(),
        }
    }

    pub(crate) fn get_reference_key_name(&self, expr: mir::Expression) -> Result<String> {
        match expr {
            mir::Expression::Reference(reference) => self
                .mapping_registry
                .get(&reference.key)
                .ok_or(Error::ReferenceNotFound(reference.key))
                .map(|s| s.name.clone()),
            mir::Expression::FieldAccess(reference) => Ok(format!(
                "{}.{}",
                self.get_reference_key_name(*reference.expr)?,
                reference.field
            )),
            _ => Err(Error::ExprNotReferenceOrFieldAccess),
        }
    }

    pub(crate) fn generate_let_bindings(
        &mut self,
        registry: MqlMappingRegistry,
    ) -> Vec<LetVariable> {
        let mut let_bindings: Vec<LetVariable> = vec![];
        let new_mapping_registry = MqlMappingRegistry::with_registry(
            registry
                .get_registry()
                .clone()
                .into_iter()
                .map(|(key, value)| {
                    let mut generated_name = format!(
                        "v{}_{}",
                        Self::get_datasource_name(&key.datasource, "__bot"),
                        key.scope
                    );

                    // Here, we replace any invalid characters with an underscore
                    // and then lowercase the whole name. The [[:word:]] character
                    // class matches word characters ([0-9A-Za-z_]).
                    generated_name = regex::Regex::new(r"[[:ascii:]&&[:^word:]]")
                        .unwrap()
                        .replace_all(generated_name.as_str(), "_")
                        .to_string()
                        .to_ascii_lowercase();
                    while let_bindings.iter().any(|x| x.name == generated_name) {
                        generated_name.push('_');
                    }
                    let expr = match value.ref_type {
                        MqlReferenceType::FieldRef => air::Expression::FieldRef(value.name.into()),
                        MqlReferenceType::Variable => air::Expression::Variable(value.name.into()),
                    };
                    let_bindings.push(LetVariable {
                        name: generated_name.clone(),
                        expr: Box::new(expr),
                    });
                    (
                        key,
                        MqlMappingRegistryValue::new(generated_name, MqlReferenceType::Variable),
                    )
                })
                .collect::<BTreeMap<Key, MqlMappingRegistryValue>>(),
        );
        // update the mapping registry with the new values for existing keys
        self.mapping_registry.merge(new_mapping_registry);
        let_bindings
    }
}

impl From<mir::Type> for air::Type {
    fn from(t: mir::Type) -> Self {
        match t {
            mir::Type::Array => air::Type::Array,
            mir::Type::BinData => air::Type::BinData,
            mir::Type::Boolean => air::Type::Boolean,
            mir::Type::Datetime => air::Type::Datetime,
            mir::Type::DbPointer => air::Type::DbPointer,
            mir::Type::Decimal128 => air::Type::Decimal128,
            mir::Type::Document => air::Type::Document,
            mir::Type::Double => air::Type::Double,
            mir::Type::Int32 => air::Type::Int32,
            mir::Type::Int64 => air::Type::Int64,
            mir::Type::Javascript => air::Type::Javascript,
            mir::Type::JavascriptWithScope => air::Type::JavascriptWithScope,
            mir::Type::MaxKey => air::Type::MaxKey,
            mir::Type::MinKey => air::Type::MinKey,
            mir::Type::Null => air::Type::Null,
            mir::Type::ObjectId => air::Type::ObjectId,
            mir::Type::RegularExpression => air::Type::RegularExpression,
            mir::Type::String => air::Type::String,
            mir::Type::Symbol => air::Type::Symbol,
            mir::Type::Timestamp => air::Type::Timestamp,
            mir::Type::Undefined => air::Type::Undefined,
        }
    }
}

impl From<mir::TypeOrMissing> for air::TypeOrMissing {
    fn from(item: mir::TypeOrMissing) -> Self {
        match item {
            mir::TypeOrMissing::Missing => air::TypeOrMissing::Missing,
            mir::TypeOrMissing::Type(t) => air::TypeOrMissing::Type(t.into()),
            mir::TypeOrMissing::Number => air::TypeOrMissing::Number,
        }
    }
}

impl From<mir::DatePart> for air::DatePart {
    fn from(dp: mir::DatePart) -> Self {
        match dp {
            mir::DatePart::Year => air::DatePart::Year,
            mir::DatePart::Quarter => air::DatePart::Quarter,
            mir::DatePart::Month => air::DatePart::Month,
            mir::DatePart::Week => air::DatePart::Week,
            mir::DatePart::Day => air::DatePart::Day,
            mir::DatePart::Hour => air::DatePart::Hour,
            mir::DatePart::Minute => air::DatePart::Minute,
            mir::DatePart::Second => air::DatePart::Second,
        }
    }
}

impl From<mir::DateFunction> for air::DateFunction {
    fn from(df: mir::DateFunction) -> Self {
        match df {
            mir::DateFunction::Add => air::DateFunction::Add,
            mir::DateFunction::Diff => air::DateFunction::Diff,
            mir::DateFunction::Trunc => air::DateFunction::Trunc,
        }
    }
}

impl From<mir::ScalarFunction> for ScalarFunctionType {
    fn from(func: mir::ScalarFunction) -> Self {
        use mir::ScalarFunction::*;
        match func {
            // String operators
            Concat => ScalarFunctionType::Mql(MQLOperator::Concat),

            // Unary arithmetic operators
            Pos => ScalarFunctionType::Sql(SQLOperator::Pos),
            Neg => ScalarFunctionType::Sql(SQLOperator::Neg),

            // Arithmetic operators
            Add => ScalarFunctionType::Mql(MQLOperator::Add),
            Sub => ScalarFunctionType::Mql(MQLOperator::Subtract),
            Mul => ScalarFunctionType::Mql(MQLOperator::Multiply),
            Div => ScalarFunctionType::Divide,

            // Comparison operators
            Lt => ScalarFunctionType::Sql(SQLOperator::Lt),
            Lte => ScalarFunctionType::Sql(SQLOperator::Lte),
            Neq => ScalarFunctionType::Sql(SQLOperator::Ne),
            Eq => ScalarFunctionType::Sql(SQLOperator::Eq),
            Gt => ScalarFunctionType::Sql(SQLOperator::Gt),
            Gte => ScalarFunctionType::Sql(SQLOperator::Gte),
            Between => ScalarFunctionType::Sql(SQLOperator::Between),

            // Boolean operators
            Not => ScalarFunctionType::Sql(SQLOperator::Not),
            And => ScalarFunctionType::Sql(SQLOperator::And),
            Or => ScalarFunctionType::Sql(SQLOperator::Or),

            // Computed Field Access operator
            // when the field is not known until runtime.
            ComputedFieldAccess => ScalarFunctionType::Sql(SQLOperator::ComputedFieldAccess),

            // Conditional scalar functions
            NullIf => ScalarFunctionType::Sql(SQLOperator::NullIf),
            Coalesce => ScalarFunctionType::Sql(SQLOperator::Coalesce),

            // Array scalar functions
            Slice => ScalarFunctionType::Sql(SQLOperator::Slice),
            Size => ScalarFunctionType::Sql(SQLOperator::Size),

            // Numeric value scalar functions
            Position => ScalarFunctionType::Sql(SQLOperator::IndexOfCP),
            CharLength => ScalarFunctionType::Sql(SQLOperator::StrLenCP),
            OctetLength => ScalarFunctionType::Sql(SQLOperator::StrLenBytes),
            BitLength => ScalarFunctionType::Sql(SQLOperator::BitLength),
            Abs => ScalarFunctionType::Mql(MQLOperator::Abs),
            Ceil => ScalarFunctionType::Mql(MQLOperator::Ceil),
            Cos => ScalarFunctionType::Sql(SQLOperator::Cos),
            Degrees => ScalarFunctionType::Mql(MQLOperator::RadiansToDegrees),
            Floor => ScalarFunctionType::Mql(MQLOperator::Floor),
            Log => ScalarFunctionType::Sql(SQLOperator::Log),
            Mod => ScalarFunctionType::Sql(SQLOperator::Mod),
            Pow => ScalarFunctionType::Mql(MQLOperator::Pow),
            Radians => ScalarFunctionType::Mql(MQLOperator::DegreesToRadians),
            Round => ScalarFunctionType::Sql(SQLOperator::Round),
            Sin => ScalarFunctionType::Sql(SQLOperator::Sin),
            Sqrt => ScalarFunctionType::Sql(SQLOperator::Sqrt),
            Tan => ScalarFunctionType::Sql(SQLOperator::Tan),

            // String value scalar functions
            Substring => ScalarFunctionType::Sql(SQLOperator::SubstrCP),
            Upper => ScalarFunctionType::Sql(SQLOperator::ToUpper),
            Lower => ScalarFunctionType::Sql(SQLOperator::ToLower),
            BTrim => ScalarFunctionType::Trim(TrimOperator::Trim),
            LTrim => ScalarFunctionType::Trim(TrimOperator::LTrim),
            RTrim => ScalarFunctionType::Trim(TrimOperator::RTrim),
            Split => ScalarFunctionType::Sql(SQLOperator::Split),

            // Datetime value scalar function
            CurrentTimestamp => ScalarFunctionType::Sql(SQLOperator::CurrentTimestamp),
            Year => ScalarFunctionType::Mql(MQLOperator::Year),
            Month => ScalarFunctionType::Mql(MQLOperator::Month),
            Day => ScalarFunctionType::Mql(MQLOperator::DayOfMonth),
            Hour => ScalarFunctionType::Mql(MQLOperator::Hour),
            Minute => ScalarFunctionType::Mql(MQLOperator::Minute),
            Second => ScalarFunctionType::Mql(MQLOperator::Second),
            Week => ScalarFunctionType::Mql(MQLOperator::Week),
            DayOfYear => ScalarFunctionType::Mql(MQLOperator::DayOfYear),
            IsoWeek => ScalarFunctionType::Mql(MQLOperator::IsoWeek),
            IsoWeekday => ScalarFunctionType::Mql(MQLOperator::IsoDayOfWeek),

            // MergeObjects merges an array of objects
            MergeObjects => ScalarFunctionType::Mql(MQLOperator::MergeObjects),
        }
    }
}
