use crate::{
    air::{self, LetVariable, MQLOperator, SQLOperator, TrimOperator},
    mapping_registry::{MqlMappingRegistry, MqlMappingRegistryValue, MqlReferenceType},
    mir::{self, ScalarFunction},
    translator::{Error, MqlTranslator, Result},
};
use mongosql_datastructures::binding_tuple::{BindingTuple, DatasourceName, Key};
use std::collections::BTreeMap;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub(crate) enum ScalarFunctionType {
    Divide,
    Mql(MQLOperator),
    Sql(SQLOperator),
    Trim(TrimOperator),
}

impl MqlTranslator {
    /// Generate a unique name given a predicate closure. Keeps pre-pending '_'
    /// until the predicate returns false, indicating that the name is not in
    /// use.
    pub(crate) fn generate_unique_datasource_name<F>(base_name: String, name_exists: F) -> String
    where
        F: Fn(&String) -> bool,
    {
        let mut ret = base_name;
        while name_exists(&ret) {
            ret.insert(0, '_')
        }
        ret
    }

    /// Ensures the base_name does not conflict with any datasource names at
    /// this scope. It does this by prepending '_' until there is no conflict.
    /// This function checks the project_names and self.mapping_registry for
    /// conflicts. This is because we need to handle conflicts within Projects,
    /// such as
    ///    SELECT `$foo`.*, `_foo`.* FROM ...
    /// as well as conflicts from datasources external to a Project, such as
    ///    SELECT * FROM `$foo` JOIN `_foo`
    ///
    /// In the first example, we will map `$foo` to `__foo` since we find
    /// `_foo` in the Project expression. We will map `_foo` to `_foo`.
    ///
    /// In the second example, we will map `$foo` to `_foo` since it has
    /// no conflicts in its individual (implicit) Project or in the mapping
    /// registry. We will map `_foo` to `__foo` since it will conflict with
    /// the `$foo` mapping from the mapping registry.
    pub(crate) fn ensure_unique_datasource_name(
        &mut self,
        base_name: String,
        project_names: &BindingTuple<mir::Expression>,
    ) -> String {
        MqlTranslator::generate_unique_datasource_name(base_name, |s| {
            let k = (s.clone(), self.scope_level).into();
            project_names.contains_key(&k)
                || (self.is_join && self.mapping_registry.contains_mapping(self.scope_level, s))
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

    /// Map a DatasourceName to a valid MQL project key name. Use the provided
    /// unique_bot_name for Bottom datasource. For datasources that contain '.'
    /// or start with '$', replace those characters with '_' and ensure the
    /// result is unique. This allows MongoSQL to support datasource aliases
    /// that contain '.'s or start with '$'s. Project keys with those attributes
    /// have different semantics or are invalid in MQL, so we must replace them
    /// with valid characters.
    pub(crate) fn get_mapped_project_name(
        &mut self,
        datasource: &DatasourceName,
        unique_bot_name: &str,
        project_names: &BindingTuple<mir::Expression>,
    ) -> Result<String> {
        let mut mapped_k = Self::get_datasource_name(datasource, unique_bot_name);
        if mapped_k.as_str() == "" {
            return Err(Error::InvalidProjectField);
        }
        let mut needs_unique_check = false;
        if mapped_k.starts_with('$') {
            mapped_k = mapped_k.replacen('$', "_", 1);
            needs_unique_check = true;
        }
        if mapped_k.contains('.') {
            mapped_k = mapped_k.replace('.', "_");
            needs_unique_check = true;
        }
        needs_unique_check = needs_unique_check
            || (self.is_join
                && self
                    .mapping_registry
                    .contains_mapping(self.scope_level, &mapped_k));
        if needs_unique_check {
            mapped_k = self.ensure_unique_datasource_name(mapped_k, project_names);
        }
        Ok(mapped_k)
    }

    pub(crate) fn get_field_path_name(&self, fp: mir::FieldPath) -> Result<String> {
        let datasource_name = self
            .mapping_registry
            .get(&fp.key)
            .ok_or(Error::ReferenceNotFound(fp.key))
            .map(|s| s.name.clone())?;
        Ok(format!("{datasource_name}.{0}", fp.fields.join(".")))
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

    pub(crate) fn translate_literal_value(&self, lit: mir::LiteralValue) -> air::LiteralValue {
        match lit {
            mir::LiteralValue::Null => air::LiteralValue::Null,
            mir::LiteralValue::Boolean(b) => air::LiteralValue::Boolean(b),
            mir::LiteralValue::String(s) => air::LiteralValue::String(s),
            mir::LiteralValue::Integer(i) => air::LiteralValue::Integer(i),
            mir::LiteralValue::Long(l) => air::LiteralValue::Long(l),
            mir::LiteralValue::Double(d) => air::LiteralValue::Double(d),
            mir::LiteralValue::DbPointer(d) => air::LiteralValue::DbPointer(d),
            mir::LiteralValue::Undefined => air::LiteralValue::Undefined,
            mir::LiteralValue::DateTime(d) => air::LiteralValue::DateTime(d),
            mir::LiteralValue::Decimal128(d) => air::LiteralValue::Decimal128(d),
            mir::LiteralValue::MinKey => air::LiteralValue::MinKey,
            mir::LiteralValue::MaxKey => air::LiteralValue::MaxKey,
            mir::LiteralValue::Timestamp(t) => air::LiteralValue::Timestamp(t),
            mir::LiteralValue::RegularExpression(r) => air::LiteralValue::RegularExpression(r),
            mir::LiteralValue::ObjectId(o) => air::LiteralValue::ObjectId(o),
            mir::LiteralValue::JavaScriptCode(j) => air::LiteralValue::JavaScriptCode(j),
            mir::LiteralValue::JavaScriptCodeWithScope(j) => {
                air::LiteralValue::JavaScriptCodeWithScope(j)
            }
            mir::LiteralValue::Symbol(s) => air::LiteralValue::Symbol(s),
            mir::LiteralValue::Binary(b) => air::LiteralValue::Binary(b),
        }
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
            mir::DatePart::Millisecond => air::DatePart::Millisecond,
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

pub(crate) fn scalar_function_to_scalar_function_type(
    is_nullable: bool,
    function: mir::ScalarFunction,
) -> ScalarFunctionType {
    if is_nullable {
        ScalarFunctionType::from(function)
    } else {
        match function {
            ScalarFunction::Or => ScalarFunctionType::Mql(MQLOperator::Or),
            ScalarFunction::And => ScalarFunctionType::Mql(MQLOperator::And),
            ScalarFunction::Between => ScalarFunctionType::Mql(MQLOperator::Between),
            ScalarFunction::Eq => ScalarFunctionType::Mql(MQLOperator::Eq),
            ScalarFunction::Position => ScalarFunctionType::Mql(MQLOperator::IndexOfCP),
            ScalarFunction::Lt => ScalarFunctionType::Mql(MQLOperator::Lt),
            ScalarFunction::Lte => ScalarFunctionType::Mql(MQLOperator::Lte),
            ScalarFunction::Gt => ScalarFunctionType::Mql(MQLOperator::Gt),
            ScalarFunction::Gte => ScalarFunctionType::Mql(MQLOperator::Gte),
            ScalarFunction::Neq => ScalarFunctionType::Mql(MQLOperator::Ne),
            ScalarFunction::Not => ScalarFunctionType::Mql(MQLOperator::Not),
            ScalarFunction::Size => ScalarFunctionType::Mql(MQLOperator::Size),
            ScalarFunction::OctetLength => ScalarFunctionType::Mql(MQLOperator::StrLenBytes),
            ScalarFunction::CharLength => ScalarFunctionType::Mql(MQLOperator::StrLenCP),
            ScalarFunction::Substring => ScalarFunctionType::Mql(MQLOperator::SubstrCP),
            ScalarFunction::Lower => ScalarFunctionType::Mql(MQLOperator::ToLower),
            ScalarFunction::Upper => ScalarFunctionType::Mql(MQLOperator::ToUpper),
            _ => ScalarFunctionType::from(function),
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
            Replace => ScalarFunctionType::Mql(MQLOperator::ReplaceAll),
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
            Millisecond => ScalarFunctionType::Mql(MQLOperator::Millisecond),
            Week => ScalarFunctionType::Mql(MQLOperator::Week),
            DayOfWeek => ScalarFunctionType::Mql(MQLOperator::DayOfWeek),
            DayOfYear => ScalarFunctionType::Mql(MQLOperator::DayOfYear),
            IsoWeek => ScalarFunctionType::Mql(MQLOperator::IsoWeek),
            IsoWeekday => ScalarFunctionType::Mql(MQLOperator::IsoDayOfWeek),

            // MergeObjects merges an array of objects
            MergeObjects => ScalarFunctionType::Mql(MQLOperator::MergeObjects),
        }
    }
}
