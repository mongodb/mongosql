use crate::{
    air::{self, SQLOperator},
    mapping_registry::{Key, MqlMappingRegistry},
    mir,
};

use lazy_static::lazy_static;
use mongosql_datastructures::{
    binding_tuple::{BindingTuple, DatasourceName},
    unique_linked_hash_map,
    unique_linked_hash_map::{DuplicateKeyError, UniqueLinkedHashMap, UniqueLinkedHashMapEntry},
};
use std::collections::BTreeSet;

use thiserror::Error;

type Result<T> = std::result::Result<T, Error>;

lazy_static! {
    pub static ref ROOT: air::Expression = air::Expression::Variable("ROOT".into());
}

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("Struct is not implemented")]
    UnimplementedStruct,
    #[error("invalid document key '{0}': document keys may not be empty, contain dots, or start with dollars")]
    InvalidDocumentKey(String),
    #[error("binding tuple key {0:?} not found in mapping registry")]
    ReferenceNotFound(Key),
    #[error("duplicate key found: {0}")]
    DuplicateKey(#[from] DuplicateKeyError),
    #[error("project fields may not be empty, contain dots, or start with dollars")]
    InvalidProjectField,
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

#[derive(Debug)]
enum ScalarFunctionType {
    Sql(air::SQLOperator),
    Mql(air::MQLOperator),
}

#[derive(Clone)]
pub struct MqlTranslator {
    pub mapping_registry: MqlMappingRegistry,
    pub scope_level: u16,
}

impl MqlTranslator {
    pub fn new() -> Self {
        Self {
            mapping_registry: Default::default(),
            scope_level: 0u16,
        }
    }

    /// Generate a unique bottom name given a predicate closure. Keeps pre-pending
    /// `_` until the predicate returns false, indicating that that name is not in use.
    fn generate_unique_bot_name<F>(name_exists: F) -> String
    where
        F: Fn(&String) -> bool,
    {
        let mut ret = "__bot".to_string();
        while name_exists(&ret) {
            ret.insert(0, '_');
        }
        ret
    }

    fn get_unique_bot_name(project_names: &BindingTuple<mir::Expression>) -> String {
        if project_names.is_empty() {
            return "__bot".to_string();
        }
        let current_scope = project_names.keys().next().unwrap().scope;
        MqlTranslator::generate_unique_bot_name(|s| {
            project_names.contains_key(&(s.clone(), current_scope).into())
        })
    }

    fn get_datasource_name(datasource: &DatasourceName, unique_bot_name: &str) -> String {
        match datasource {
            DatasourceName::Bottom => unique_bot_name.to_string(),
            DatasourceName::Named(s) => s.clone(),
        }
    }

    // translate_plan is the entry point, it mostly just calls translate_stage, but also
    // sets up a project to replace __bot with the empty key: ''.
    pub fn translate_plan(&mut self, mir_stage: mir::Stage) -> Result<air::Stage> {
        let source = self.translate_stage(mir_stage)?;
        self.get_replace_bot_stage(source)
    }

    /// replace_bot will be called at the end of translate_plan to add a stage
    /// that replaces the unique bottom name with an empty string "" as the last stage.
    fn get_replace_bot_stage(&mut self, source: air::Stage) -> Result<air::Stage> {
        let key = Key {
            datasource: DatasourceName::Bottom,
            scope: 0u16,
        };
        let mongo_bot_name = self.mapping_registry.remove(&key);
        Ok(match mongo_bot_name {
            Some(name) => {
                self.mapping_registry.insert(key, "");
                air::Stage::ReplaceRoot(air::ReplaceRoot {
                    source: Box::new(source),
                    new_root: Box::new(air::Expression::Unset(air::Unset {
                        field: name.clone(),
                        input: Box::new(air::Expression::SetField(air::SetField {
                            field: "".to_string(),
                            input: Box::new(air::Expression::Variable("ROOT".to_string())),
                            value: Box::new(air::Expression::FieldRef(air::FieldRef {
                                parent: None,
                                name,
                            })),
                        })),
                    })),
                })
            }
            None => source,
        })
    }

    pub(crate) fn translate_stage(&mut self, mir_stage: mir::Stage) -> Result<air::Stage> {
        match mir_stage {
            mir::Stage::Array(arr) => self.translate_array_stage(arr),
            mir::Stage::Collection(c) => self.translate_collection(c),
            mir::Stage::Project(p) => self.translate_project(p),
            mir::Stage::Filter(f) => self.translate_filter(f),
            mir::Stage::Group(g) => self.translate_group(g),
            mir::Stage::Limit(_l) => Err(Error::UnimplementedStruct),
            mir::Stage::Offset(_o) => Err(Error::UnimplementedStruct),
            mir::Stage::Sort(_s) => Err(Error::UnimplementedStruct),
            mir::Stage::Join(_j) => Err(Error::UnimplementedStruct),
            mir::Stage::Set(_s) => Err(Error::UnimplementedStruct),
            mir::Stage::Derived(_d) => Err(Error::UnimplementedStruct),
            mir::Stage::Unwind(_u) => Err(Error::UnimplementedStruct),
        }
    }

    fn translate_array_stage(&mut self, mir_arr: mir::ArraySource) -> Result<air::Stage> {
        let doc_stage = air::Stage::Documents(air::Documents {
            array: mir_arr
                .array
                .iter()
                .map(|mir_expr| self.translate_expression(mir_expr.clone()))
                .collect::<Result<Vec<air::Expression>>>()?,
        });

        self.mapping_registry.insert(
            Key::named(&mir_arr.alias, self.scope_level),
            mir_arr.alias.clone(),
        );

        Ok(air::Stage::Project(air::Project {
            source: Box::new(doc_stage),
            specifications: unique_linked_hash_map! {
                mir_arr.alias => ROOT.clone(),
            },
        }))
    }

    fn translate_collection(&mut self, mir_collection: mir::Collection) -> Result<air::Stage> {
        let coll_stage = air::Stage::Collection(air::Collection {
            db: mir_collection.db,
            collection: mir_collection.collection.clone(),
        });

        self.mapping_registry.insert(
            Key::named(&mir_collection.collection, self.scope_level),
            mir_collection.collection.clone(),
        );
        Ok(air::Stage::Project(air::Project {
            source: Box::new(coll_stage),
            specifications: unique_linked_hash_map! {
                mir_collection.collection => ROOT.clone(),
            },
        }))
    }

    fn translate_project(&mut self, mir_project: mir::Project) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_project.source)?;

        // We will add mappings to the mapping registry introduced by this Project Stage, which
        // is all of the keys. Previous bindings are removed after the subexpressions are
        // translated because Project kills all its inputs.
        let unique_bot_name = Self::get_unique_bot_name(&mir_project.expression);
        let mut project_body = UniqueLinkedHashMap::new();
        let mut output_registry = MqlMappingRegistry::new();
        for (k, e) in mir_project.expression.into_iter() {
            let mapped_k = Self::get_datasource_name(&k.datasource, &unique_bot_name);
            if mapped_k.starts_with('$') || mapped_k.contains('.') || mapped_k.as_str() == "" {
                return Err(Error::InvalidProjectField);
            }

            project_body.insert(mapped_k.clone(), self.translate_expression(e)?)?;
            output_registry.insert(k, mapped_k);
        }
        self.mapping_registry = output_registry;
        Ok(air::Stage::Project(air::Project {
            source: Box::new(source_translation),
            specifications: project_body,
        }))
    }

    fn translate_filter(&mut self, mir_filter: mir::Filter) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_filter.source)?;
        let expr_translation = self.translate_expression(mir_filter.condition)?;

        Ok(air::Stage::Match(air::Match {
            source: Box::new(source_translation),
            expr: Box::new(expr_translation),
        }))
    }

    fn get_unique_alias(existing_aliases: &BTreeSet<String>, mut alias: String) -> String {
        while existing_aliases.contains(&alias) {
            alias.insert(0, '_')
        }
        alias
    }

    fn translate_agg_function(afa: mir::AggregationFunction) -> air::AggregationFunction {
        match afa {
            mir::AggregationFunction::AddToArray => air::AggregationFunction::AddToArray,
            mir::AggregationFunction::Avg => air::AggregationFunction::Avg,
            mir::AggregationFunction::Count => air::AggregationFunction::Count,
            mir::AggregationFunction::First => air::AggregationFunction::First,
            mir::AggregationFunction::Last => air::AggregationFunction::Last,
            mir::AggregationFunction::Max => air::AggregationFunction::Max,
            mir::AggregationFunction::MergeDocuments => air::AggregationFunction::MergeDocuments,
            mir::AggregationFunction::Min => air::AggregationFunction::Min,
            mir::AggregationFunction::StddevPop => air::AggregationFunction::StddevPop,
            mir::AggregationFunction::StddevSamp => air::AggregationFunction::StddevSamp,
            mir::AggregationFunction::Sum => air::AggregationFunction::Sum,
        }
    }

    fn translate_group(&mut self, mir_group: mir::Group) -> Result<air::Stage> {
        let source_translation = self.translate_stage(*mir_group.source)?;
        let unique_aliases = mir_group
            .keys
            .iter()
            .filter_map(|k| k.get_alias().map(String::from))
            .collect::<BTreeSet<_>>();

        // all the keys defined by this stage must flow out to the next stage. All the incoming
        // keys are killed by the Group stage.
        let mut out_registry = MqlMappingRegistry::new();

        // map the group keys and translate the expressions.
        let keys = mir_group
            .keys
            .into_iter()
            .enumerate()
            .map(|(i, k)| {
                Ok(match k {
                    mir::OptionallyAliasedExpr::Aliased(ae) => {
                        let registry_key = Key::named(&ae.alias, self.scope_level);
                        out_registry.insert(registry_key, ae.alias.clone());
                        air::NameExprPair {
                            name: ae.alias,
                            expr: self.translate_expression(ae.expr)?,
                        }
                    }
                    mir::OptionallyAliasedExpr::Unaliased(e) => {
                        let position_counter = i + 1;
                        let unique_name = Self::get_unique_alias(
                            &unique_aliases,
                            format!("__unaliasedKey{position_counter}"),
                        );
                        let registry_key = Key::named(&unique_name, self.scope_level);
                        out_registry.insert(registry_key, unique_name.clone());
                        air::NameExprPair {
                            name: unique_name,
                            expr: self.translate_expression(e)?,
                        }
                    }
                })
            })
            .collect::<Result<Vec<_>>>()?;

        let aggregations = mir_group
            .aggregations
            .into_iter()
            .map(|a| {
                let alias = a.alias;
                let registry_key = Key::named(&alias, self.scope_level);
                out_registry.insert(registry_key, alias.clone());
                let (function, distinct, arg) = match a.agg_expr {
                    mir::AggregationExpr::CountStar(b) => (
                        air::AggregationFunction::Count,
                        b,
                        Box::new(air::Expression::Literal(air::LiteralValue::Integer(1))),
                    ),
                    mir::AggregationExpr::Function(afa) => (
                        Self::translate_agg_function(afa.function),
                        afa.distinct,
                        Box::new(self.translate_expression(*afa.arg)?),
                    ),
                };
                Ok(air::AccumulatorExpr {
                    alias,
                    function,
                    distinct,
                    arg,
                })
            })
            .collect::<Result<Vec<_>>>()?;

        // set the new mapping registry for following stages.
        self.mapping_registry = out_registry;
        // map the group aggregations and translate the expressions.
        Ok(air::Stage::Group(air::Group {
            source: Box::new(source_translation),
            keys,
            aggregations,
        }))
    }

    pub fn translate_expression(&self, mir_expression: mir::Expression) -> Result<air::Expression> {
        match mir_expression {
            mir::Expression::Literal(lit) => self.translate_literal(lit.value),
            mir::Expression::Document(doc) => self.translate_document(doc.document),
            mir::Expression::Array(expr) => self.translate_array(expr.array),
            mir::Expression::Reference(reference) => self.translate_reference(reference.key),
            mir::Expression::Cast(cast) => self.translate_cast(cast),
            mir::Expression::ScalarFunction(scalar_func) => {
                self.translate_scalar_function(scalar_func)
            }
            _ => Err(Error::UnimplementedStruct),
        }
    }

    fn translate_literal(&self, lit: mir::LiteralValue) -> Result<air::Expression> {
        Ok(air::Expression::Literal(match lit {
            mir::LiteralValue::Null => air::LiteralValue::Null,
            mir::LiteralValue::Boolean(b) => air::LiteralValue::Boolean(b),
            mir::LiteralValue::String(s) => air::LiteralValue::String(s),
            mir::LiteralValue::Integer(i) => air::LiteralValue::Integer(i),
            mir::LiteralValue::Long(l) => air::LiteralValue::Long(l),
            mir::LiteralValue::Double(d) => air::LiteralValue::Double(d),
        }))
    }

    fn translate_document(
        &self,
        mir_document: UniqueLinkedHashMap<String, mir::Expression>,
    ) -> Result<air::Expression> {
        Ok(air::Expression::Document(
            mir_document
                .into_iter()
                .map(|(k, v)| {
                    if k.starts_with('$') || k.contains('.') || k.is_empty() {
                        Err(Error::InvalidDocumentKey(k))
                    } else {
                        Ok(UniqueLinkedHashMapEntry::new(
                            k,
                            self.translate_expression(v)?,
                        ))
                    }
                })
                .collect::<Result<
                    std::result::Result<
                        UniqueLinkedHashMap<String, air::Expression>,
                        DuplicateKeyError,
                    >,
                >>()??,
        ))
    }

    fn translate_array(&self, array: Vec<mir::Expression>) -> Result<air::Expression> {
        Ok(air::Expression::Array(
            array
                .into_iter()
                .map(|x| self.translate_expression(x))
                .collect::<Result<Vec<air::Expression>>>()?,
        ))
    }

    fn translate_reference(&self, key: Key) -> Result<air::Expression> {
        self.mapping_registry
            .get(&key)
            .ok_or(Error::ReferenceNotFound(key))
            .map(|s| {
                air::Expression::FieldRef(air::FieldRef {
                    parent: None,
                    name: s.clone(),
                })
            })
    }

    fn translate_cast(&self, cast: mir::CastExpr) -> Result<air::Expression> {
        let input = self.translate_expression(*cast.expr)?.into();
        let to = cast.to.into();
        let on_null = self.translate_expression(*cast.on_null)?.into();
        let on_error = self.translate_expression(*cast.on_error)?.into();
        Ok(match to {
            air::Type::Array | air::Type::Document => {
                air::Expression::SqlConvert(air::SqlConvert {
                    input,
                    to,
                    on_null,
                    on_error,
                })
            }
            _ => air::Expression::Convert(air::Convert {
                input,
                to,
                on_null,
                on_error,
            }),
        })
    }

    fn to_air_op(func: mir::ScalarFunction) -> ScalarFunctionType {
        use mir::ScalarFunction::*;
        match func {
            // String operators
            Concat => ScalarFunctionType::Mql(air::MQLOperator::Concat),

            // Unary arithmetic operators
            Pos => ScalarFunctionType::Sql(air::SQLOperator::Pos),
            Neg => ScalarFunctionType::Sql(air::SQLOperator::Neg),

            // Arithmetic operators
            Add => ScalarFunctionType::Mql(air::MQLOperator::Add),
            Sub => ScalarFunctionType::Mql(air::MQLOperator::Subtract),
            Mul => ScalarFunctionType::Mql(air::MQLOperator::Multiply),
            Div => ScalarFunctionType::Sql(air::SQLOperator::Divide),

            // Comparison operators
            Lt => ScalarFunctionType::Sql(air::SQLOperator::Lt),
            Lte => ScalarFunctionType::Sql(air::SQLOperator::Lte),
            Neq => ScalarFunctionType::Sql(air::SQLOperator::Ne),
            Eq => ScalarFunctionType::Sql(air::SQLOperator::Eq),
            Gt => ScalarFunctionType::Sql(air::SQLOperator::Gt),
            Gte => ScalarFunctionType::Sql(air::SQLOperator::Gte),
            Between => ScalarFunctionType::Sql(air::SQLOperator::Between),

            // Boolean operators
            Not => ScalarFunctionType::Sql(air::SQLOperator::Not),
            And => ScalarFunctionType::Sql(air::SQLOperator::And),
            Or => ScalarFunctionType::Sql(air::SQLOperator::Or),

            // Computed Field Access operator
            // when the field is not known until runtime.
            ComputedFieldAccess => ScalarFunctionType::Sql(SQLOperator::ComputedFieldAccess),

            // Conditional scalar functions
            NullIf => ScalarFunctionType::Sql(air::SQLOperator::NullIf),
            Coalesce => ScalarFunctionType::Sql(air::SQLOperator::Coalesce),

            // Array scalar functions
            Slice => ScalarFunctionType::Sql(air::SQLOperator::Slice),
            Size => ScalarFunctionType::Sql(air::SQLOperator::Size),

            // Numeric value scalar functions
            Position => ScalarFunctionType::Sql(air::SQLOperator::IndexOfCP),
            CharLength => ScalarFunctionType::Sql(air::SQLOperator::StrLenCP),
            OctetLength => ScalarFunctionType::Sql(air::SQLOperator::StrLenBytes),
            BitLength => ScalarFunctionType::Sql(air::SQLOperator::BitLength),
            Abs => ScalarFunctionType::Mql(air::MQLOperator::Abs),
            Ceil => ScalarFunctionType::Mql(air::MQLOperator::Ceil),
            Cos => ScalarFunctionType::Mql(air::MQLOperator::Cos),
            Degrees => ScalarFunctionType::Mql(air::MQLOperator::RadiansToDegrees),
            Floor => ScalarFunctionType::Mql(air::MQLOperator::Floor),
            Log => ScalarFunctionType::Sql(air::SQLOperator::Log),
            Mod => ScalarFunctionType::Sql(air::SQLOperator::Mod),
            Pow => ScalarFunctionType::Mql(air::MQLOperator::Pow),
            Radians => ScalarFunctionType::Mql(air::MQLOperator::DegreesToRadians),
            Round => ScalarFunctionType::Sql(air::SQLOperator::Round),
            Sin => ScalarFunctionType::Mql(air::MQLOperator::Sin),
            Sqrt => ScalarFunctionType::Mql(air::MQLOperator::Sqrt),
            Tan => ScalarFunctionType::Mql(air::MQLOperator::Tan),

            // String value scalar functions
            Substring => ScalarFunctionType::Sql(air::SQLOperator::SubstrCP),
            Upper => ScalarFunctionType::Sql(air::SQLOperator::ToUpper),
            Lower => ScalarFunctionType::Sql(air::SQLOperator::ToLower),
            BTrim => ScalarFunctionType::Sql(air::SQLOperator::Trim),
            LTrim => ScalarFunctionType::Sql(air::SQLOperator::LTrim),
            RTrim => ScalarFunctionType::Sql(air::SQLOperator::RTrim),
            Split => ScalarFunctionType::Sql(air::SQLOperator::Split),

            // Datetime value scalar function
            CurrentTimestamp => ScalarFunctionType::Sql(air::SQLOperator::CurrentTimestamp),
            Year => ScalarFunctionType::Mql(air::MQLOperator::Year),
            Month => ScalarFunctionType::Mql(air::MQLOperator::Month),
            Day => ScalarFunctionType::Mql(air::MQLOperator::DayOfMonth),
            Hour => ScalarFunctionType::Mql(air::MQLOperator::Hour),
            Minute => ScalarFunctionType::Mql(air::MQLOperator::Minute),
            Second => ScalarFunctionType::Mql(air::MQLOperator::Second),
            Week => ScalarFunctionType::Mql(air::MQLOperator::Week),
            DayOfYear => ScalarFunctionType::Mql(air::MQLOperator::DayOfYear),
            IsoWeek => ScalarFunctionType::Mql(air::MQLOperator::IsoWeek),
            IsoWeekday => ScalarFunctionType::Mql(air::MQLOperator::IsoDayOfWeek),

            // MergeObjects merges an array of objects
            MergeObjects => ScalarFunctionType::Mql(air::MQLOperator::MergeObjects),
        }
    }

    fn translate_scalar_function(
        &self,
        scalar_func: mir::ScalarFunctionApplication,
    ) -> Result<air::Expression> {
        let args = scalar_func
            .args
            .into_iter()
            .map(|x| self.translate_expression(x))
            .collect::<Result<_>>()?;
        let op = Self::to_air_op(scalar_func.function);
        match op {
            ScalarFunctionType::Sql(op) => Ok(air::Expression::SQLSemanticOperator(
                air::SQLSemanticOperator { op, args },
            )),
            ScalarFunctionType::Mql(op) => Ok(air::Expression::MQLSemanticOperator(
                air::MQLSemanticOperator { op, args },
            )),
        }
    }
}
