use crate::{
    air::{self, LetVariable, SQLOperator},
    mapping_registry::MqlReferenceType,
    mir,
    translator::{utils::ScalarFunctionType, Error, MqlTranslator, Result},
};
use mongosql_datastructures::{
    binding_tuple::Key,
    unique_linked_hash_map::{DuplicateKeyError, UniqueLinkedHashMap, UniqueLinkedHashMapEntry},
};

impl MqlTranslator {
    pub fn translate_expression(&self, mir_expression: mir::Expression) -> Result<air::Expression> {
        match mir_expression {
            mir::Expression::Array(expr) => self.translate_array_expression(expr.array),
            mir::Expression::Cast(cast) => self.translate_cast(cast),
            mir::Expression::DateFunction(date_func_app) => {
                self.translate_date_function(date_func_app)
            }
            mir::Expression::Document(doc) => self.translate_document(doc.document),
            mir::Expression::Exists(exists) => self.translate_exists(exists),
            mir::Expression::FieldAccess(field_access) => self.translate_field_access(field_access),
            mir::Expression::Is(is) => self.translate_is(is),
            mir::Expression::Like(like_expr) => self.translate_like(like_expr),
            mir::Expression::Literal(lit) => self.translate_literal(lit.value),
            mir::Expression::Reference(reference) => self.translate_reference(reference.key),

            mir::Expression::ScalarFunction(scalar_func) => {
                self.translate_scalar_function(scalar_func)
            }
            mir::Expression::SearchedCase(searched_case) => {
                self.translate_searched_case(searched_case)
            }
            mir::Expression::SimpleCase(simple_case) => self.translate_simple_case(simple_case),
            mir::Expression::Subquery(subquery) => self
                .translate_subquery(subquery)
                .map(air::Expression::Subquery),
            mir::Expression::SubqueryComparison(subquery_comparison) => {
                self.translate_subquery_comparison(subquery_comparison)
            }

            mir::Expression::TypeAssertion(ta) => self.translate_expression(*ta.expr),
        }
    }

    fn translate_array_expression(&self, array: Vec<mir::Expression>) -> Result<air::Expression> {
        Ok(air::Expression::Array(
            array
                .into_iter()
                .map(|x| self.translate_expression(x))
                .collect::<Result<Vec<air::Expression>>>()?,
        ))
    }

    fn translate_cast(&self, cast: mir::CastExpr) -> Result<air::Expression> {
        let input = self.translate_expression(*cast.expr)?.into();
        let to = cast.to.into();
        let on_null = self.translate_expression(*cast.on_null)?.into();
        let on_error = self.translate_expression(*cast.on_error)?.into();
        Ok(match to {
            air::Type::Array | air::Type::Document => {
                let sql_convert_to = match to {
                    air::Type::Array => air::SqlConvertTargetType::Array,
                    air::Type::Document => air::SqlConvertTargetType::Document,
                    _ => return Err(Error::InvalidSqlConvertToType(to)),
                };
                air::Expression::SqlConvert(air::SqlConvert {
                    input,
                    to: sql_convert_to,
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

    fn translate_date_function(
        &self,
        date_func_app: mir::DateFunctionApplication,
    ) -> Result<air::Expression> {
        Ok(air::Expression::DateFunction(
            air::DateFunctionApplication {
                function: air::DateFunction::from(date_func_app.function),
                unit: air::DatePart::from(date_func_app.date_part),
                args: date_func_app
                    .args
                    .into_iter()
                    .map(|expr| self.translate_expression(expr))
                    .collect::<Result<Vec<air::Expression>>>()?,
            },
        ))
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

    fn translate_exists(&self, exists: mir::ExistsExpr) -> Result<air::Expression> {
        // Clone self so that we can translate the subquery pipeline
        // without modifying self's mapping registry or scope.
        let mut subquery_translator = self.clone();

        let (let_bindings, pipeline) =
            subquery_translator.translate_subquery_pipeline(*exists.stage)?;

        Ok(air::Expression::SubqueryExists(air::SubqueryExists {
            let_bindings,
            pipeline: Box::new(pipeline),
        }))
    }

    fn translate_field_access(&self, field_access: mir::FieldAccess) -> Result<air::Expression> {
        let expr = self.translate_expression(*field_access.expr)?;
        let field = field_access.field;
        if let air::Expression::FieldRef(r) = expr.clone() {
            if !(field.contains('.') || field.starts_with('$') || field.as_str() == "") {
                return Ok(air::Expression::FieldRef(air::FieldRef {
                    parent: Some(Box::new(r)),
                    name: field,
                }));
            }
        }
        if let air::Expression::Variable(v) = expr.clone() {
            return Ok(air::Expression::Variable(air::Variable {
                parent: Some(Box::new(v)),
                name: field,
            }));
        }
        Ok(air::Expression::GetField(air::GetField {
            field,
            input: Box::new(expr),
        }))
    }

    fn translate_is(&self, is_expr: mir::IsExpr) -> Result<air::Expression> {
        let expr = self.translate_expression(*is_expr.expr)?;
        let target_type = air::TypeOrMissing::from(is_expr.target_type);
        Ok(air::Expression::Is(air::Is {
            expr: Box::new(expr),
            target_type,
        }))
    }

    fn translate_like(&self, like_expr: mir::LikeExpr) -> Result<air::Expression> {
        let expr = self.translate_expression(*like_expr.expr)?.into();
        let pattern = self.translate_expression(*like_expr.pattern)?.into();
        Ok(air::Expression::Like(air::Like {
            expr,
            pattern,
            escape: like_expr.escape,
        }))
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

    fn translate_reference(&self, key: Key) -> Result<air::Expression> {
        self.mapping_registry
            .get(&key)
            .ok_or(Error::ReferenceNotFound(key))
            .map(|s| match s.ref_type {
                MqlReferenceType::FieldRef => air::Expression::FieldRef(s.name.clone().into()),
                MqlReferenceType::Variable => air::Expression::Variable(s.name.to_string().into()),
            })
    }

    fn translate_scalar_function(
        &self,
        scalar_func: mir::ScalarFunctionApplication,
    ) -> Result<air::Expression> {
        let args: Vec<air::Expression> = scalar_func
            .args
            .into_iter()
            .map(|x| self.translate_expression(x))
            .collect::<Result<Vec<air::Expression>>>()?;
        let op = ScalarFunctionType::from(scalar_func.function);
        match op {
            ScalarFunctionType::Divide => Ok(air::Expression::SqlDivide(air::SqlDivide {
                dividend: Box::new(args[0].clone()),
                divisor: Box::new(args[1].clone()),
                on_error: Box::new(air::Expression::Literal(air::LiteralValue::Null)),
            })),
            ScalarFunctionType::Trim(op) => Ok(air::Expression::Trim(air::Trim {
                op,
                input: Box::new(args[1].clone()),
                chars: Box::new(args[0].clone()),
            })),
            // SqlOperator::IndexOfCP has reversed arguments
            ScalarFunctionType::Sql(SQLOperator::IndexOfCP) => Ok(
                air::Expression::SQLSemanticOperator(air::SQLSemanticOperator {
                    op: SQLOperator::IndexOfCP,
                    args: args.into_iter().rev().collect(),
                }),
            ),
            ScalarFunctionType::Sql(op) => Ok(air::Expression::SQLSemanticOperator(
                air::SQLSemanticOperator { op, args },
            )),
            ScalarFunctionType::Mql(op) => Ok(air::Expression::MQLSemanticOperator(
                air::MQLSemanticOperator { op, args },
            )),
        }
    }

    fn translate_searched_case(
        &self,
        searched_case: mir::SearchedCaseExpr,
    ) -> Result<air::Expression> {
        let default = self
            .translate_expression(*searched_case.else_branch)?
            .into();
        let branches = searched_case
            .when_branch
            .iter()
            .map(|branch| {
                Ok(air::SwitchCase {
                    case: Box::new(self.translate_expression(*branch.when.clone())?),
                    then: Box::new(self.translate_expression(*branch.then.clone())?),
                })
            })
            .collect::<Result<Vec<air::SwitchCase>>>()?;
        Ok(air::Expression::Switch(air::Switch { branches, default }))
    }

    fn translate_simple_case(&self, simple_case: mir::SimpleCaseExpr) -> Result<air::Expression> {
        let expr = self.translate_expression(*simple_case.expr)?;
        let default = self.translate_expression(*simple_case.else_branch)?.into();
        let branches = simple_case
            .when_branch
            .iter()
            .map(|branch| {
                Ok(air::SwitchCase {
                    case: Box::new(air::Expression::SQLSemanticOperator(
                        air::SQLSemanticOperator {
                            op: SQLOperator::Eq,
                            args: vec![
                                air::Expression::Variable("target".to_string().into()),
                                self.translate_expression(*branch.when.clone())?,
                            ],
                        },
                    )),
                    then: Box::new(self.translate_expression(*branch.then.clone())?),
                })
            })
            .collect::<Result<Vec<air::SwitchCase>>>()?;
        let switch = air::Expression::Switch(air::Switch { branches, default });
        Ok(air::Expression::Let(air::Let {
            vars: vec![LetVariable {
                name: "target".to_string(),
                expr: Box::new(expr),
            }],
            inside: Box::new(switch),
        }))
    }

    fn translate_subquery(&self, subquery: mir::SubqueryExpr) -> Result<air::Subquery> {
        // Clone self so that we can translate the subquery pipeline and output
        // path without modifying self's mapping registry or scope.
        let mut subquery_translator = self.clone();

        let (let_bindings, pipeline) =
            subquery_translator.translate_subquery_pipeline(*subquery.subquery)?;

        // Translate the output_expr to a Vec<String>. Instead of using
        // translate_expression, which may produce GetFields instead of
        // FieldRefs, we manually walk the mir::Expression FieldAccess
        // tree to get the components.
        let output_path = subquery_translator.generate_path_components(*subquery.output_expr)?;

        Ok(air::Subquery {
            let_bindings,
            output_path,
            pipeline: Box::new(pipeline),
        })
    }

    fn translate_subquery_comparison(
        &self,
        subquery_comparison: mir::SubqueryComparison,
    ) -> Result<air::Expression> {
        use mir::{SubqueryComparisonOp, SubqueryModifier};

        let op = match subquery_comparison.operator {
            SubqueryComparisonOp::Lt => air::SubqueryComparisonOp::Lt,
            SubqueryComparisonOp::Lte => air::SubqueryComparisonOp::Lte,
            SubqueryComparisonOp::Neq => air::SubqueryComparisonOp::Neq,
            SubqueryComparisonOp::Eq => air::SubqueryComparisonOp::Eq,
            SubqueryComparisonOp::Gt => air::SubqueryComparisonOp::Gt,
            SubqueryComparisonOp::Gte => air::SubqueryComparisonOp::Gte,
        };

        let modifier = match subquery_comparison.modifier {
            SubqueryModifier::Any => air::SubqueryModifier::Any,
            SubqueryModifier::All => air::SubqueryModifier::All,
        };

        let arg = self.translate_expression(*subquery_comparison.argument)?;

        let subquery = self.translate_subquery(subquery_comparison.subquery_expr)?;

        Ok(air::Expression::SubqueryComparison(
            air::SubqueryComparison {
                op,
                modifier,
                arg: Box::new(arg),
                subquery: Box::new(subquery),
            },
        ))
    }

    // Utility functions for Subquery* translations
    fn translate_subquery_pipeline(
        &mut self,
        subquery: mir::Stage,
    ) -> Result<(Vec<LetVariable>, air::Stage)> {
        // Generate let bindings for the subquery and update the subquery
        // translator's mapping registry
        let let_bindings = self.generate_let_bindings(self.mapping_registry.clone());

        // Increase the scope level to translate the subquery pipeline
        self.scope_level += 1;

        // Translate the subquery pipeline
        let subquery_translation = self.translate_stage(subquery)?;

        Ok((let_bindings, subquery_translation))
    }

    /// generate_path_components takes an expression and returns a vector of
    /// its components by recursively tracing its path.
    fn generate_path_components(&self, expr: mir::Expression) -> Result<Vec<String>> {
        match expr {
            mir::Expression::Reference(mir::ReferenceExpr { key, .. }) => {
                match self.mapping_registry.get(&key) {
                    Some(registry_value) => Ok(vec![registry_value.name.clone()]),
                    None => Err(Error::ReferenceNotFound(key)),
                }
            }
            mir::Expression::FieldAccess(fa) => {
                let mut path = self.generate_path_components(*fa.expr)?;
                path.push(fa.field.clone());
                Ok(path)
            }
            _ => Err(Error::SubqueryOutputPathNotFieldRef),
        }
    }
}
