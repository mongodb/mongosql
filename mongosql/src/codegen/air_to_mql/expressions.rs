use super::{Error, MqlCodeGenerator, Result};
use crate::air::{self, SQLOperator};
use bson::{bson, doc, Bson};

impl MqlCodeGenerator {
    pub fn codegen_air_expression(&self, expr: air::Expression) -> Result<bson::Bson> {
        use air::{Expression::*, LiteralValue::*};
        match expr {
            Switch(switch) => {
                let branches = switch
                    .branches
                    .into_iter()
                    .map(|sw| {
                        Ok(doc! {"case": self.codegen_air_expression(*sw.case)?,
                        "then": self.codegen_air_expression(*sw.then)?})
                    })
                    .collect::<Result<Vec<bson::Document>>>()?;
                let default = self.codegen_air_expression(*switch.default)?;

                Ok(bson!({
                    "$switch": {
                        "branches": branches,
                        "default": default,
                    }
                }))
            }
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
                })
            }
            GetField(gf) => Ok({
                let input = self.codegen_air_expression(*gf.input)?;
                let field = Self::wrap_in_literal_if(gf.field, |s| s.starts_with('$'));
                bson!({
                    "$getField": {
                        "field": field,
                        "input": input,
                    }
                })
            }),
            Is(is) => {
                let expr = self.codegen_air_expression(*is.expr).unwrap();
                let target_type = is.target_type.to_str();
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
            Like(like_expr) => {
                let mut like = doc! {
                    "input": self.codegen_air_expression(*like_expr.expr)?,
                    "pattern": self.codegen_air_expression(*like_expr.pattern)?,
                };
                if like_expr.escape.is_some() {
                    like.insert("escape", like_expr.escape.unwrap());
                }
                Ok(Bson::Document(doc! {"$like": like}))
            }
            UnsetField(uf) => {
                let field = Self::wrap_in_literal_if(uf.field, |s| s.starts_with('$'));
                let input = self.codegen_air_expression(*uf.input)?;
                Ok(bson!({"$unsetField": {"field": field, "input": input}}))
            }
            Convert(c) => Ok({
                let input = self.codegen_air_expression(*c.input)?;
                let on_error = self.codegen_air_expression(*c.on_error)?;
                let on_null = self.codegen_air_expression(*c.on_null)?;
                bson!({
                    "$convert": {
                        "input": input,
                        "to": Self::convert_mql_type(c.to)?,
                        "onNull": on_null,
                        "onError": on_error
                    }
                })
            }),
            SqlConvert(sc) => Ok({
                let input = self.codegen_air_expression(*sc.input)?;
                let on_error = self.codegen_air_expression(*sc.on_error)?;
                let on_null = self.codegen_air_expression(*sc.on_null)?;
                bson!({
                    "$sqlConvert": {
                        "input": input,
                        "to": sc.to.to_str(),
                        "onNull": on_null,
                        "onError": on_error
                    }
                })
            }),
            Let(l) => {
                let vars = l
                    .vars
                    .into_iter()
                    .map(|v| Ok((v.name, self.codegen_air_expression(*v.expr)?)))
                    .collect::<Result<bson::Document>>()?;

                let inside = self.codegen_air_expression(*l.inside)?;

                Ok(bson!({"$let": {"vars": vars, "in": inside}}))
            }
            _ => Err(Error::UnimplementedAIR),
        }
    }
}
