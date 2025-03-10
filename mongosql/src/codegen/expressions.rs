use crate::{
    air::{self, SQLOperator, TrimOperator},
    codegen::{Error, MqlCodeGenerator, Result},
};
use bson::{bson, doc, Bson};
use mongosql_datastructures::unique_linked_hash_map::UniqueLinkedHashMap;

impl MqlCodeGenerator {
    pub fn codegen_expression(&self, expr: air::Expression) -> Result<Bson> {
        use air::Expression::*;
        match expr {
            MQLSemanticOperator(air::MQLSemanticOperator {
                op: air::MQLOperator::ReplaceAll,
                args,
            }) => self.codegen_replace_all(args),
            MQLSemanticOperator(mql_op) => self.codegen_mql_semantic_operator(mql_op),
            SQLSemanticOperator(sql_op) => self.codegen_sql_semantic_operator(sql_op),
            Literal(lit) => self.codegen_literal(lit),
            FieldRef(fr) => Ok(Bson::String(self.codegen_field_ref(fr))),
            Variable(var) => Ok(Bson::String(self.codegen_variable(var))),
            GetField(gf) => self.codegen_get_field(gf),
            SetField(sf) => self.codegen_set_field(sf),
            UnsetField(uf) => self.codegen_unset_field(uf),
            Switch(switch) => self.codegen_switch(switch),
            Let(l) => self.codegen_let(l),
            SqlConvert(sc) => self.codegen_sql_convert(sc),
            Convert(c) => self.codegen_convert(c),
            Like(like) => self.codegen_like(like),
            Is(is) => self.codegen_is(is),
            DateFunction(df) => self.codegen_date_function(df),
            RegexMatch(r) => self.codegen_regex_match(r),
            SqlDivide(sd) => self.codegen_sql_divide(sd),
            Trim(trim) => self.codegen_trim(trim),
            Map(m) => self.codegen_map(m),
            Reduce(r) => self.codegen_reduce(r),
            Subquery(s) => self.codegen_subquery_expr(s),
            SubqueryComparison(sc) => self.codegen_subquery_comparison(sc),
            SubqueryExists(se) => self.codegen_subquery_exists(se),
            Array(array) => self.codegen_array(array),
            Document(document) => self.codegen_document(document),
        }
    }

    // If we ever support another function that uses this document-style arguments, we should
    // abstract this to something like:
    // fn code_gen_mql_semantic_operator_with_document_args(
    //     &self,
    //     arg_names: &[&str],
    //     mql_op: air::MQLSemanticOperator) -> Result<Bson>
    // At this point, it is overkill.
    fn codegen_replace_all(&self, args: Vec<air::Expression>) -> Result<Bson> {
        // $replaceAll uses a document format for args that goes against
        // most other scalar functions in mongodb, unfortunately.
        let ops = ["input", "find", "replacement"]
            .into_iter()
            .zip(args)
            .map(|(arg_name, arg)| Ok((arg_name.to_string(), self.codegen_expression(arg)?)))
            .collect::<Result<bson::Document>>()?;
        // We still use to_mql_op so that all MQL operator names can be found in one place.
        let operator = Self::to_mql_op(air::MQLOperator::ReplaceAll);
        Ok(bson::bson!({ operator: Bson::Document(ops)}))
    }

    fn codegen_mql_semantic_operator(&self, mql_op: air::MQLSemanticOperator) -> Result<Bson> {
        let ops = mql_op
            .args
            .into_iter()
            .map(|x| self.codegen_expression(x))
            .collect::<Result<Vec<_>>>()?;
        let operator = Self::to_mql_op(mql_op.op);
        Ok(bson::bson!({ operator: Bson::Array(ops) }))
    }

    fn codegen_sql_semantic_operator(&self, sql_op: air::SQLSemanticOperator) -> Result<Bson> {
        Ok(match sql_op.op {
            SQLOperator::Size
            | SQLOperator::StrLenCP
            | SQLOperator::StrLenBytes
            | SQLOperator::ToUpper
            | SQLOperator::ToLower => {
                bson::bson!({ Self::to_sql_op(sql_op.op).unwrap(): self.codegen_expression(sql_op.args[0].clone())?})
            }
            SQLOperator::And
            | SQLOperator::Between
            | SQLOperator::BitLength
            | SQLOperator::Coalesce
            | SQLOperator::Cos
            | SQLOperator::Eq
            | SQLOperator::Gt
            | SQLOperator::Gte
            | SQLOperator::IndexOfCP
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
                let ops = sql_op
                    .args
                    .into_iter()
                    .map(|x| self.codegen_expression(x))
                    .collect::<Result<Vec<_>>>()?;
                bson::bson!({ Self::to_sql_op(sql_op.op).unwrap(): Bson::Array(ops) })
            }
            SQLOperator::ComputedFieldAccess => {
                // Adding this feature is tracked in SQL-673
                return Err(Error::UnsupportedOperator(SQLOperator::ComputedFieldAccess));
            }
            SQLOperator::CurrentTimestamp => Bson::String("$$NOW".to_string()),
        })
    }

    fn codegen_literal(&self, lit: air::LiteralValue) -> Result<Bson> {
        use air::LiteralValue::*;
        Ok(bson::bson!({
            "$literal": match lit {
                Null => Bson::Null,
                Boolean(b) => Bson::Boolean(b),
                String(s) => Bson::String(s),
                Integer(i) => Bson::Int32(i),
                Long(l) => Bson::Int64(l),
                Double(d) => Bson::Double(d),
                Decimal128(d) => Bson::Decimal128(d),
                ObjectId(o) => Bson::ObjectId(o),
                DateTime(d) => Bson::DateTime(d),
                DbPointer(d) => Bson::DbPointer(d),
                Undefined => Bson::Undefined,
                Timestamp(t) => Bson::Timestamp(t),
                RegularExpression(r) => Bson::RegularExpression(r),
                MinKey => Bson::MinKey,
                MaxKey => Bson::MaxKey,
                Symbol(s) => Bson::Symbol(s),
                JavaScriptCode(j) => Bson::JavaScriptCode(j),
                JavaScriptCodeWithScope(j) => Bson::JavaScriptCodeWithScope(j),
                Binary(b) => Bson::Binary(b),
            },
        }))
    }

    fn codegen_field_ref(&self, field_ref: air::FieldRef) -> String {
        format!("${}", self.codegen_field_ref_path_only(field_ref))
    }

    #[allow(clippy::only_used_in_recursion)] // false positive
    fn codegen_variable(&self, var: air::Variable) -> String {
        match var.parent {
            None => format!("$${}", var.name),
            Some(parent) => format!("{}.{}", self.codegen_variable(*parent), var.name),
        }
    }

    fn codegen_get_field(&self, gf: air::GetField) -> Result<Bson> {
        Ok({
            let input = self.codegen_expression(*gf.input)?;
            let field = Self::wrap_in_literal_if(gf.field, |s| s.starts_with('$'));
            bson!({
                "$getField": {
                    "field": field,
                    "input": input,
                }
            })
        })
    }

    fn codegen_set_field(&self, sf: air::SetField) -> Result<Bson> {
        let field = Self::wrap_in_literal_if(sf.field, |s| s.starts_with('$'));
        let input = self.codegen_expression(*sf.input)?;
        let value = self.codegen_expression(*sf.value)?;
        Ok(bson!({"$setField": {
            "field": field,
            "input": input,
            "value": value
        }}))
    }

    fn codegen_unset_field(&self, uf: air::UnsetField) -> Result<Bson> {
        let field = Self::wrap_in_literal_if(uf.field, |s| s.starts_with('$'));
        let input = self.codegen_expression(*uf.input)?;
        Ok(bson!({"$unsetField": {"field": field, "input": input}}))
    }

    fn codegen_switch(&self, switch: air::Switch) -> Result<Bson> {
        let branches = switch
            .branches
            .into_iter()
            .map(|sw| {
                Ok(doc! {"case": self.codegen_expression(*sw.case)?,
                "then": self.codegen_expression(*sw.then)?})
            })
            .collect::<Result<Vec<bson::Document>>>()?;
        let default = self.codegen_expression(*switch.default)?;

        Ok(bson!({
            "$switch": {
                "branches": branches,
                "default": default,
            }
        }))
    }

    fn codegen_let(&self, let_expr: air::Let) -> Result<Bson> {
        let vars = let_expr
            .vars
            .into_iter()
            .map(|v| Ok((v.name, self.codegen_expression(*v.expr)?)))
            .collect::<Result<bson::Document>>()?;

        let inside = self.codegen_expression(*let_expr.inside)?;

        Ok(bson!({"$let": {"vars": vars, "in": inside}}))
    }

    fn codegen_sql_convert(&self, sql_convert: air::SqlConvert) -> Result<Bson> {
        Ok({
            let input = self.codegen_expression(*sql_convert.input)?;
            let on_error = self.codegen_expression(*sql_convert.on_error)?;
            let on_null = self.codegen_expression(*sql_convert.on_null)?;
            bson!({
                "$sqlConvert": {
                    "input": input,
                    "to": sql_convert.to.to_str(),
                    "onNull": on_null,
                    "onError": on_error
                }
            })
        })
    }

    fn codegen_convert(&self, convert: air::Convert) -> Result<Bson> {
        Ok({
            let input = self.codegen_expression(*convert.input)?;
            let on_error = self.codegen_expression(*convert.on_error)?;
            let on_null = self.codegen_expression(*convert.on_null)?;
            // Until we support extra CastExpr options, this will ensure we maintain
            // the same output for all versions of MongoDB. Because format *must* be specified
            // in 8.0 when converting binData to string, and it is a hard parse error in versions
            // < 8.0, we can't include it anywhere until users can specify it in their query themselves.
            let to_type = Self::convert_mql_type(convert.to)?;
            let to = if convert.to == air::Type::String {
                bson!({
                    "$cond": [
                        { "$eq": [ { "$type": input.clone() }, "binData" ] },
                        "null",
                        to_type,
                    ],
                })
            } else {
                Bson::String(to_type.to_string())
            };

            bson!({
                "$convert": {
                    "input": input,
                    "to": to,
                    "onNull": on_null,
                    "onError": on_error
                }
            })
        })
    }

    fn codegen_like(&self, like: air::Like) -> Result<Bson> {
        let mut like_doc = doc! {
            "input": self.codegen_expression(*like.expr)?,
            "pattern": self.codegen_expression(*like.pattern)?,
        };
        if like.escape.is_some() {
            like_doc.insert("escape", like.escape.unwrap().to_string());
        }
        Ok(Bson::Document(doc! {"$like": like_doc}))
    }

    fn codegen_is(&self, is: air::Is) -> Result<Bson> {
        let expr = self.codegen_expression(*is.expr).unwrap();
        let target_type = is.target_type.to_str();
        Ok(bson ! ({"$sqlIs": [expr, {"$literal": target_type}]}))
    }

    fn codegen_date_function(&self, date_func_app: air::DateFunctionApplication) -> Result<Bson> {
        use air::DateFunction::*;

        Ok(match date_func_app.function {
            Add => {
                bson::bson!({"$dateAdd" : {
                    "startDate": self.codegen_expression(date_func_app.args[1].clone())?,
                    "unit": Self::date_part_to_mql_unit(date_func_app.unit),
                    "amount": self.codegen_expression(date_func_app.args[0].clone())?,
                }})
            }
            Diff => {
                bson::bson!({"$dateDiff" : {
                    "startDate": self.codegen_expression(date_func_app.args[0].clone())?,
                    "endDate": self.codegen_expression(date_func_app.args[1].clone())?,
                    "unit": Self::date_part_to_mql_unit(date_func_app.unit),
                    "startOfWeek": self.codegen_expression(date_func_app.args[2].clone())?,
                }})
            }
            Trunc => {
                bson::bson!({"$dateTrunc" : {
                    "date": self.codegen_expression(date_func_app.args[0].clone())?,
                    "unit": Self::date_part_to_mql_unit(date_func_app.unit),
                    "startOfWeek": self.codegen_expression(date_func_app.args[1].clone())?,
                }})
            }
        })
    }

    fn codegen_regex_match(&self, regex_match: air::RegexMatch) -> Result<Bson> {
        let input = self.codegen_expression(*regex_match.input)?;
        let regex = self.codegen_expression(*regex_match.regex)?;
        Ok(match regex_match.options {
            Some(opts) => {
                bson!({"$regexMatch": {"input": input, "regex": regex, "options": self.codegen_expression(*opts)?}})
            }
            None => bson!({"$regexMatch": {"input": input, "regex": regex}}),
        })
    }

    fn codegen_sql_divide(&self, sql_divide: air::SqlDivide) -> Result<Bson> {
        let dividend = self.codegen_expression(*sql_divide.dividend)?;
        let divisor = self.codegen_expression(*sql_divide.divisor)?;
        let on_error = self.codegen_expression(*sql_divide.on_error)?;
        Ok(bson!({"$sqlDivide": {"dividend": dividend, "divisor": divisor, "onError": on_error}}))
    }

    fn codegen_trim(&self, trim: air::Trim) -> Result<Bson> {
        let op = match trim.op {
            TrimOperator::Trim => "$trim",
            TrimOperator::LTrim => "$ltrim",
            TrimOperator::RTrim => "$rtrim",
        };
        Ok(Bson::Document(doc! {
            op: {"input": self.codegen_expression(*trim.input)?,
                "chars": self.codegen_expression(*trim.chars)?}
        }))
    }

    fn codegen_map(&self, m: air::Map) -> Result<Bson> {
        let input = self.codegen_expression(*m.input)?;
        let inside = self.codegen_expression(*m.inside)?;

        let mut doc = doc! {"input": input, "in": inside};

        if let Some(as_name) = m.as_name {
            doc.insert("as", as_name);
        }

        Ok(bson!({"$map": doc}))
    }

    fn codegen_reduce(&self, reduce: air::Reduce) -> Result<Bson> {
        let input = self.codegen_expression(*reduce.input)?;
        let init_value = self.codegen_expression(*reduce.init_value)?;
        let inside = self.codegen_expression(*reduce.inside)?;
        Ok(bson!({"$reduce": {"input": input, "initialValue": init_value, "in": inside}}))
    }

    fn codegen_subquery(
        &self,
        pipeline: air::Stage,
        let_bindings: Vec<air::LetVariable>,
        output_path: Option<Vec<String>>,
    ) -> Result<Bson> {
        let mut subquery_body = doc! {};

        let pipeline_translation = self.codegen_stage(pipeline)?;

        if let Some(db) = pipeline_translation.database {
            subquery_body.insert("db", db);
        }

        if let Some(collection) = pipeline_translation.collection {
            subquery_body.insert("collection", collection);
        }

        let let_bindings = let_bindings
            .into_iter()
            .map(|v| Ok((v.name.clone(), self.codegen_expression(*v.expr)?)))
            .collect::<Result<bson::Document>>()?;

        subquery_body.insert("let", let_bindings);

        if let Some(output_path) = output_path {
            subquery_body.insert("outputPath", output_path);
        }

        subquery_body.insert("pipeline", pipeline_translation.pipeline);

        Ok(bson!(subquery_body))
    }

    fn codegen_subquery_expr(&self, subquery: air::Subquery) -> Result<Bson> {
        Ok(
            bson!({ "$subquery": self.codegen_subquery(*subquery.pipeline, subquery.let_bindings, Some(subquery.output_path))? }),
        )
    }

    fn codegen_subquery_comparison(&self, sc: air::SubqueryComparison) -> Result<Bson> {
        use air::{SubqueryComparisonOp::*, SubqueryComparisonOpType::*, SubqueryModifier::*};

        let op = match (sc.op_type, sc.op) {
            (Mql, Lt) => "lt",
            (Mql, Lte) => "lte",
            (Mql, Neq) => "ne",
            (Mql, Eq) => "eq",
            (Mql, Gt) => "gt",
            (Mql, Gte) => "gte",
            (Sql, Lt) => "sqlLt",
            (Sql, Lte) => "sqlLte",
            (Sql, Neq) => "sqlNe",
            (Sql, Eq) => "sqlEq",
            (Sql, Gt) => "sqlGt",
            (Sql, Gte) => "sqlGte",
        };

        let modifier = match sc.modifier {
            Any => "any",
            All => "all",
        };

        let arg = self.codegen_expression(*sc.arg)?;

        let subquery = self.codegen_subquery(
            *sc.subquery.pipeline,
            sc.subquery.let_bindings,
            Some(sc.subquery.output_path),
        )?;

        Ok(bson!({"$subqueryComparison": {
            "op": op,
            "modifier": modifier,
            "arg": arg,
            "subquery": subquery
        }}))
    }

    fn codegen_subquery_exists(&self, subquery_exists: air::SubqueryExists) -> Result<Bson> {
        Ok(
            bson!({ "$subqueryExists": self.codegen_subquery(*subquery_exists.pipeline, subquery_exists.let_bindings, None)? }),
        )
    }

    fn codegen_array(&self, array: Vec<air::Expression>) -> Result<Bson> {
        Ok(Bson::Array(
            array
                .into_iter()
                .map(|e| self.codegen_expression(e))
                .collect::<Result<Vec<Bson>>>()?,
        ))
    }

    fn codegen_document(
        &self,
        document: UniqueLinkedHashMap<String, air::Expression>,
    ) -> Result<Bson> {
        Ok(Bson::Document({
            if document.is_empty() {
                bson::doc! {"$literal": {}}
            } else {
                document
                    .into_iter()
                    .map(|(k, v)| Ok((k, self.codegen_expression(v)?)))
                    .collect::<Result<bson::Document>>()?
            }
        }))
    }
}
