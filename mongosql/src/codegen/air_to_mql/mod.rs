use crate::air::{
    self, AggregationFunction, MQLOperator, SQLOperator, SqlConvertTargetType, Type, TypeOrMissing,
};

use bson::{bson, doc, Bson};
use thiserror::Error;

#[cfg(test)]
mod test;

pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug, Error, PartialEq, Eq)]
pub enum Error {
    #[error("air method is not implemented")]
    UnimplementedAIR,
    #[error("cannot generate MQL for {0:?} operator")]
    UnsupportedOperator(SQLOperator),
}

#[derive(PartialEq, Debug)]
pub struct MqlTranslation {
    pub database: Option<String>,
    pub collection: Option<String>,
    pub pipeline: Vec<bson::Document>,
}

#[derive(Clone, Debug)]
pub struct MqlCodeGenerator {}

impl MqlCodeGenerator {
    fn agg_func_to_mql_op(mqla: AggregationFunction) -> &'static str {
        use AggregationFunction::*;
        match mqla {
            AddToArray => "$push",
            Avg => "$avg",
            Count => unreachable!(),
            First => "$first",
            Last => "$last",
            Max => "$max",
            MergeDocuments => "$mergeObjects",
            Min => "$min",
            StddevPop => "$stdDevPop",
            StddevSamp => "$stdDevSamp",
            Sum => "$sum",
        }
    }

    fn agg_func_to_sql_op(mqla: AggregationFunction) -> &'static str {
        use AggregationFunction::*;
        match mqla {
            AddToArray => "$sqlPush",
            Avg => "$sqlAvg",
            Count => "$sqlCount",
            First => "$sqlFirst",
            Last => "$sqlLast",
            Max => "$sqlMax",
            MergeDocuments => "$sqlMergeObjects",
            Min => "$sqlMin",
            StddevPop => "$sqlStdDevPop",
            StddevSamp => "$sqlStdDevSamp",
            Sum => "$sqlSum",
        }
    }

    fn to_mql_op(mqlo: MQLOperator) -> &'static str {
        use MQLOperator::*;
        match mqlo {
            // String operators
            Concat => "$concat",

            // Arithmetic operators
            Add => "$add",
            Subtract => "$subtract",
            Multiply => "$multiply",
            Divide => "$divide",

            // Comparison operators
            Lt => "$lt",
            Lte => "$lte",
            Ne => "$ne",
            Eq => "$eq",
            Gt => "$gt",
            Gte => "$gte",

            // Boolean operators
            Not => "$not",
            And => "$and",
            Or => "$or",

            // Array scalar functions
            Slice => "$slice",
            Size => "$size",

            // Numeric value scalar functions
            IndexOfCP => "$indexOfCP",
            IndexOfBytes => "$indexOfBytes",
            StrLenCP => "$strLenCP",
            StrLenBytes => "$strLenBytes",
            Abs => "$abs",
            Ceil => "$ceil",
            Cos => "$cos",
            DegreesToRadians => "$degreesToRadians",
            Floor => "$floor",
            Log => "$log",
            Mod => "$mod",
            Pow => "$pow",
            RadiansToDegrees => "$radiansToDegrees",
            Round => "$round",
            Sin => "$sin",
            Tan => "$tan",
            Sqrt => "$sqrt",

            // String value scalar functions
            SubstrCP => "$substrCP",
            SubstrBytes => "$substrBytes",
            ToUpper => "$toUpper",
            ToLower => "$toLower",
            Trim => "$trim",
            LTrim => "$ltrim",
            RTrim => "$rtrim",
            Split => "$split",

            // Datetime value scalar function
            Year => "$year",
            Month => "$month",
            DayOfMonth => "$dayOfMonth",
            Hour => "$hour",
            Minute => "$minute",
            Second => "$second",
            Week => "$week",
            DayOfYear => "$dayOfYear",
            IsoWeek => "$isoWeek",
            IsoDayOfWeek => "$isoDayOfWeek",
            DateAdd => "$dateAdd",
            DateDiff => "$dateDiff",
            DateTrunc => "$dateTrunc",

            // MergeObjects merges an array of objects
            MergeObjects => "$mergeObjects",
        }
    }

    fn to_sql_op(sqlo: SQLOperator) -> Option<&'static str> {
        use SQLOperator::*;
        Some(match sqlo {
            // Arithmetic operators
            Divide => "$sqlDivide",
            Pos => "$sqlPos",
            Neg => "$sqlNeg",

            // Comparison operators
            Lt => "$sqlLt",
            Lte => "$sqlLte",
            Ne => "$sqlNe",
            Eq => "$sqlEq",
            Gt => "$sqlGt",
            Gte => "$sqlGte",
            Between => "$sqlBetween",

            // Boolean operators
            Not => "$sqlNot",
            And => "$sqlAnd",
            Or => "$sqlOr",

            // Conditional scalar functions
            NullIf => "$nullIf",
            Coalesce => "$coalesce",

            // Array scalar functions
            Slice => "$sqlSlice",
            Size => "$sqlSize",

            // Numeric value scalar functions
            IndexOfCP => "$sqlIndexOfCP",
            StrLenCP => "$sqlStrLenCP",
            StrLenBytes => "$sqlStrLenBytes",
            BitLength => "$sqlBitLength",
            Cos => "$sqlCos",
            Log => "$sqlLog",
            Mod => "$sqlMod",
            Round => "$sqlRound",
            Sin => "$sqlSin",
            Sqrt => "$sqlSqrt",
            Tan => "$sqlTan",

            // String value scalar functions
            SubstrCP => "$sqlSubstrCP",
            ToUpper => "$sqlToUpper",
            ToLower => "$sqlToLower",
            Split => "$sqlSplit",
            Trim => "$trim",
            LTrim => "$ltrim",
            RTrim => "$rtrim",

            // ComputedFieldAccess, CurrentTimestamp
            _ => return None,
        })
    }

    fn sql_convert_type(t: SqlConvertTargetType) -> &'static str {
        use SqlConvertTargetType::*;
        match t {
            Array => "array",
            Document => "object",
        }
    }

    fn type_as_str(t: Type) -> &'static str {
        use Type::*;
        match t {
            Array => "array",
            BinData => "binData",
            Boolean => "bool",
            Datetime => "date",
            DbPointer => "dbPointer",
            Decimal128 => "decimal",
            Document => "object",
            Double => "double",
            Int32 => "int",
            Int64 => "long",
            Javascript => "javascript",
            JavascriptWithScope => "javascriptWithScope",
            MaxKey => "maxKey",
            MinKey => "minKey",
            Null => "null",
            ObjectId => "objectId",
            RegularExpression => "regex",
            String => "string",
            Symbol => "symbol",
            Timestamp => "timestamp",
            Undefined => "undefined",
        }
    }

    fn type_or_missing_as_str(t: TypeOrMissing) -> &'static str {
        use TypeOrMissing::*;
        match t {
            Missing => "missing",
            Number => "number",
            Type(ty) => Self::type_as_str(ty),
        }
    }

    /// Wraps a string value, s, in $literal if the condition, f, is true for the string.
    fn wrap_in_literal_if<F>(s: String, f: F) -> Bson
    where
        F: Fn(String) -> bool,
    {
        if f(s.clone()) {
            bson!({ "$literal": s })
        } else {
            Bson::String(s)
        }
    }

    pub fn codegen_air_expression(&self, expr: air::Expression) -> Result<bson::Bson> {
        use air::{Expression::*, LiteralValue::*};
        match expr {
            Literal(lit) => Ok(bson::bson!({
                "$literal": match lit {
                    Null => Bson::Null,
                    Boolean(b) => Bson::Boolean(b),
                    String(s) => Bson::String(s),
                    Integer(i) => Bson::Int32(i),
                    Long(l) => Bson::Int64(l),
                    Double(d) => Bson::Double(d),
                },
            })),
            Document(document) => Ok(Bson::Document({
                if document.is_empty() {
                    bson::doc! {"$literal": {}}
                } else {
                    document
                        .into_iter()
                        .map(|(k, v)| Ok((k, self.codegen_air_expression(v)?)))
                        .collect::<Result<bson::Document>>()?
                }
            })),
            Array(array) => Ok(Bson::Array(
                array
                    .into_iter()
                    .map(|e| self.codegen_air_expression(e))
                    .collect::<Result<Vec<Bson>>>()?,
            )),
            Variable(var) => Ok(Bson::String(format!("$${var}"))),
            FieldRef(fr) => Ok(Bson::String(self.codegen_field_ref(fr))),
            MQLSemanticOperator(mqls) => {
                let ops = mqls
                    .args
                    .into_iter()
                    .map(|x| self.codegen_air_expression(x))
                    .collect::<Result<Vec<_>>>()?;
                let operator = Self::to_mql_op(mqls.op);
                Ok(bson::bson!({ operator: Bson::Array(ops) }))
            }
            SQLSemanticOperator(sqls) => {
                Ok(match sqls.op {
                    SQLOperator::Size
                    | SQLOperator::StrLenCP
                    | SQLOperator::StrLenBytes
                    | SQLOperator::ToUpper
                    | SQLOperator::ToLower => {
                        bson::bson!({ Self::to_sql_op(sqls.op).unwrap(): self.codegen_air_expression(sqls.args[0].clone())?})
                    }
                    SQLOperator::And
                    | SQLOperator::Between
                    | SQLOperator::BitLength
                    | SQLOperator::Coalesce
                    | SQLOperator::Cos
                    | SQLOperator::Eq
                    | SQLOperator::Gt
                    | SQLOperator::Gte
                    | SQLOperator::Log
                    | SQLOperator::Lt
                    | SQLOperator::Lte
                    | SQLOperator::Mod
                    | SQLOperator::Ne
                    | SQLOperator::Neg
                    | SQLOperator::Not
                    | SQLOperator::NullIf
                    | SQLOperator::Or
                    | SQLOperator::Pos
                    | SQLOperator::Round
                    | SQLOperator::Sin
                    | SQLOperator::Slice
                    | SQLOperator::Split
                    | SQLOperator::Sqrt
                    | SQLOperator::SubstrCP
                    | SQLOperator::Tan => {
                        let ops = sqls
                            .args
                            .into_iter()
                            .map(|x| self.codegen_air_expression(x))
                            .collect::<Result<Vec<_>>>()?;
                        bson::bson!({ Self::to_sql_op(sqls.op).unwrap(): Bson::Array(ops) })
                    }
                    SQLOperator::ComputedFieldAccess => {
                        // Adding this feature is tracked in SQL-673
                        return Err(Error::UnsupportedOperator(SQLOperator::ComputedFieldAccess));
                    }
                    SQLOperator::CurrentTimestamp => Bson::String("$$NOW".to_string()),
                    SQLOperator::Divide => Bson::Document(bson::doc! {
                        Self::to_sql_op(sqls.op).unwrap(): {
                            "dividend": self.codegen_air_expression(sqls.args[0].clone())?,
                            "divisor": self.codegen_air_expression(sqls.args[1].clone())?,
                            "onError": {"$literal": Bson::Null}
                        }
                    }),
                    // operators that reverse argument order
                    SQLOperator::IndexOfCP => {
                        let args = Bson::Array(
                            sqls.args
                                .into_iter()
                                .rev()
                                .map(|e| self.codegen_air_expression(e))
                                .collect::<Result<Vec<Bson>>>()?,
                        );
                        Bson::Document(bson::doc! { Self::to_sql_op(sqls.op).unwrap(): args})
                    }
                    SQLOperator::Trim | SQLOperator::LTrim | SQLOperator::RTrim => bson::bson!({
                        Self::to_sql_op(sqls.op).unwrap(): {"input": self.codegen_air_expression(sqls.args[1].clone())?,
                            "chars": self.codegen_air_expression(sqls.args[0].clone())?,
                    }}),
                })
            }
            GetField(gf) => Ok({
                let input = self.codegen_air_expression(*gf.input)?;
                bson!({
                    "$getField": {
                        "field": gf.field,
                        "input": input,
                    }
                })
            }),
            SqlConvert(sc) => {
                let input = self.codegen_air_expression(*sc.input)?;
                let convert_to = Self::sql_convert_type(sc.to);
                let on_null = self.codegen_air_expression(*sc.on_null)?;
                let on_error = self.codegen_air_expression(*sc.on_error)?;
                Ok({
                    bson!({
                        "$sqlConvert": {
                            "input": input,
                            "to": convert_to,
                            "onNull": on_null,
                            "onError": on_error,
                        }
                    })
                })
            }
            Is(is) => {
                let expr = self.codegen_air_expression(*is.expr).unwrap();
                let target_type = Self::type_or_missing_as_str(is.target_type);
                Ok(bson ! ({"$sqlIs": [expr, {"$literal": target_type}]}))
            }
            SetField(sf) => {
                let field = Self::wrap_in_literal_if(sf.field, |s| s.starts_with('$'));
                let input = self.codegen_air_expression(*sf.input)?;
                let value = self.codegen_air_expression(*sf.value)?;
                Ok(bson!({"$setField": {
                    "field": field,
                    "input": input,
                    "value": value
                }}))
            }
            UnsetField(uf) => {
                let field = Self::wrap_in_literal_if(uf.field, |s| s.starts_with('$'));
                let input = self.codegen_air_expression(*uf.input)?;
                Ok(bson!({"$unsetField": {"field": field, "input": input}}))
            }
            _ => Err(Error::UnimplementedAIR),
        }
    }

    #[allow(clippy::only_used_in_recursion)] // false positive
    fn codegen_field_ref(&self, field_ref: air::FieldRef) -> String {
        match field_ref.parent {
            None => format!("${}", field_ref.name),
            Some(parent) => format!("{}.{}", self.codegen_field_ref(*parent), field_ref.name),
        }
    }

    pub fn codegen_air_stage(&self, stage: air::Stage) -> Result<MqlTranslation> {
        match stage {
            air::Stage::Project(p) => self.codegen_project(p),
            air::Stage::Group(g) => self.codegen_group(g),
            air::Stage::Limit(_l) => Err(Error::UnimplementedAIR),
            air::Stage::Sort(_s) => Err(Error::UnimplementedAIR),
            air::Stage::Collection(c) => self.codegen_collection(c),
            air::Stage::Join(_j) => Err(Error::UnimplementedAIR),
            air::Stage::Unwind(_u) => Err(Error::UnimplementedAIR),
            air::Stage::Lookup(_l) => Err(Error::UnimplementedAIR),
            air::Stage::ReplaceWith(r) => self.codegen_replace_with(r),
            air::Stage::Match(_m) => Err(Error::UnimplementedAIR),
            air::Stage::UnionWith(_u) => Err(Error::UnimplementedAIR),
            air::Stage::Skip(_s) => Err(Error::UnimplementedAIR),
            air::Stage::Documents(d) => self.codegen_documents(d),
        }
    }

    fn codegen_replace_with(&self, air_replace_with: air::ReplaceWith) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_replace_with.source)?;
        let mut pipeline = source_translation.pipeline;
        let expr = self.codegen_air_expression(*air_replace_with.new_root)?;

        pipeline.push(doc! {"$replaceWith": expr});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_documents(&self, air_docs: air::Documents) -> Result<MqlTranslation> {
        let docs = air_docs
            .array
            .into_iter()
            .map(|e| self.codegen_air_expression(e))
            .collect::<Result<Vec<Bson>>>()?;
        Ok(MqlTranslation {
            database: None,
            collection: None,
            pipeline: vec![doc! {"$documents": Bson::Array(docs)}],
        })
    }

    fn codegen_collection(&self, air_coll: air::Collection) -> Result<MqlTranslation> {
        Ok(MqlTranslation {
            database: Some(air_coll.db),
            collection: Some(air_coll.collection),
            pipeline: vec![],
        })
    }

    fn codegen_project(&self, air_project: air::Project) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_project.source)?;
        let mut pipeline = source_translation.pipeline;
        let mut project_doc = air_project
            .specifications
            .into_iter()
            .map(|(k, v)| Ok((k, self.codegen_air_expression(v)?)))
            .collect::<Result<bson::Document>>()?;
        if !project_doc.contains_key("_id") {
            // we create a temporary so that _id: 0 will always be the first element
            // in the doc. This does add another linear factor to the code, but makes
            // testing easier.
            let mut tmp_project_doc = doc! {"_id": 0};
            tmp_project_doc.extend(project_doc);
            project_doc = tmp_project_doc;
        }
        pipeline.push(doc! {"$project": project_doc});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }

    fn codegen_group(&self, air_group: air::Group) -> Result<MqlTranslation> {
        let source_translation = self.codegen_air_stage(*air_group.source)?;
        let mut pipeline = source_translation.pipeline;
        let id_doc = air_group
            .keys
            .into_iter()
            .map(|air::NameExprPair { name: k, expr: v }| Ok((k, self.codegen_air_expression(v)?)))
            .collect::<Result<bson::Document>>()?;
        let mut group_doc = doc! {"_id": id_doc};
        let aggs = air_group
            .aggregations
            .into_iter()
            .map(
                |air::AccumulatorExpr {
                     alias,
                     function,
                     distinct,
                     arg,
                 }| {
                    Ok(if distinct || function == AggregationFunction::Count {
                        (alias, bson!({ Self::agg_func_to_sql_op(function): {"var": self.codegen_air_expression(*arg)?, "distinct": distinct }}))
                    } else {
                        (alias, bson!({ Self::agg_func_to_mql_op(function): self.codegen_air_expression(*arg)? }))
                    })
                },
            )
            .collect::<Result<bson::Document>>()?;
        group_doc.extend(aggs);
        pipeline.push(doc! {"$group": group_doc});
        Ok(MqlTranslation {
            database: source_translation.database,
            collection: source_translation.collection,
            pipeline,
        })
    }
}
