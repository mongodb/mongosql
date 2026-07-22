///
/// Constant Folding
///
/// This optimization replaces constant expressions with the values they will evaluate to at
/// query time. The goal of this is to reduce the amount of work done during query execution.
///
use crate::{
    catalog::Catalog,
    mir::{definitions::*, schema::SchemaInferenceState, visitor::Visitor},
    schema::{Atomic, Satisfaction, Schema, NULLISH},
};
use bson::{oid::ObjectId, Decimal128};
use chrono::Utc;
use lazy_static::lazy_static;
use std::str::FromStr;

#[derive(Clone)]
pub(crate) struct ConstantFoldExprVisitor<'a> {
    pub(crate) state: &'a SchemaInferenceState<'a>,

    // changed is a flag for tracking whether this optimization changes the
    // input. It must be manually set to true by the implementor when a Stage
    // or Expression is constant-folded in some way. A different way to track
    // changes would be to compare the before and after input to the Visitor.
    // We do not do this since it would require cloning the entire plan tree
    // to make the comparison.
    pub(crate) changed: bool,
}

lazy_static! {
    static ref DEFAULT_CATALOG: Catalog = Catalog::default();
}

impl ConstantFoldExprVisitor<'_> {
    // Checks if a vector of expressions contains a null or missing expression
    fn has_null_arg(&self, args: &[Expression]) -> bool {
        for expr in args {
            match expr.schema(self.state) {
                Err(_) => return false,
                Ok(sch) => {
                    if self.schema_is_exactly_nullish(sch) {
                        return true;
                    }
                }
            }
        }
        false
    }

    fn schema_is_exactly_nullish(&self, schema: Schema) -> bool {
        schema.satisfies(&NULLISH) == Satisfaction::Must
    }

    // This is not a general purpose function and is not capable of checking equality of very
    // large longs. It is used to check special arithmetic edge cases like 0 and 1.
    fn numeric_eq(expr: &Expression, num: f64) -> bool {
        match expr {
            Expression::Literal(l) => match l {
                LiteralValue::Integer(val) => *val == num as i32,
                LiteralValue::Long(val) => *val == num as i64,
                LiteralValue::Double(val) => *val == num,
                _ => false,
            },
            _ => false,
        }
    }

    // Constant folds boolean functions
    fn fold_logical_function(&mut self, sf: ScalarFunctionApplication) -> (Expression, bool) {
        let (nullish, non_nullish): (Vec<Expression>, Vec<Expression>) =
            sf.args.clone().into_iter().partition(|e| {
                let schema = e.schema(self.state).unwrap_or(Schema::Any);
                self.schema_is_exactly_nullish(schema)
            });
        let has_null = !nullish.is_empty();
        let (fold_init, op): (bool, Box<dyn Fn(bool, bool) -> bool>) = match sf.function {
            ScalarFunction::And => (true, Box::new(|acc, x| x && acc)),
            ScalarFunction::Or => (false, Box::new(|acc, x| x || acc)),
            _ => unreachable!("fold logical functions is only called on And and Or"),
        };
        let mut non_literals = Vec::<Expression>::new();
        let folded_constant = non_nullish
            .into_iter()
            .fold(fold_init, |acc, expr| match expr {
                Expression::Literal(LiteralValue::Boolean(val)) => op(acc, val),
                expr => {
                    non_literals.push(expr);
                    acc
                }
            });
        let folded_expr = Expression::Literal(LiteralValue::Boolean(folded_constant));
        if non_literals.is_empty() && !has_null
            || (sf.function == ScalarFunction::And && !folded_constant)
            || (sf.function == ScalarFunction::Or && folded_constant)
        {
            return (folded_expr, true);
        }
        if has_null {
            non_literals.push(Expression::Literal(LiteralValue::Null))
        }
        if non_literals.len() == 1 {
            return (non_literals[0].clone(), true);
        }
        // At this point, we may or may not have simplified the Expression by
        // reducing the number of arguments. If the number of arguments has been
        // reduced, we must set changed to true. This catches cases such as
        //   null OR a OR null OR b
        // which will be simplified to
        //   a OR b OR null
        // This simplification would not trigger any of the other "changed"
        // cases above.
        let changed = non_literals.len() < sf.args.len();
        (
            Expression::ScalarFunction(ScalarFunctionApplication {
                function: sf.function,
                is_nullable: sf.is_nullable,
                args: non_literals,
            }),
            changed,
        )
    }

    // Constant folds constants of the same type within an associative arithmetic function
    fn fold_associative_arithmetic_function(
        &mut self,
        sf: ScalarFunctionApplication,
    ) -> (Expression, bool) {
        if self.has_null_arg(&sf.args) {
            return (Expression::Literal(LiteralValue::Null), true);
        }
        if sf.args.is_empty() {
            match sf.function {
                ScalarFunction::Add => {
                    return (Expression::Literal(LiteralValue::Integer(0)), true);
                }
                ScalarFunction::Mul => {
                    return (Expression::Literal(LiteralValue::Integer(1)), true);
                }
                _ => unreachable!("fold associative function only called on Add and Mul"),
            }
        }
        let mut non_literals = Vec::<Expression>::new();
        let (int_fold, long_fold, float_fold, arg_count) = match sf.function {
            ScalarFunction::Add => {
                sf.args
                    .into_iter()
                    .fold((None, None, None, 0), |(i, l, f, count), expr| match expr {
                        Expression::Literal(LiteralValue::Integer(val)) => match i {
                            Some(num) => (Some(val + num), l, f, count + 1),
                            None => (Some(val), l, f, count + 1),
                        },
                        Expression::Literal(LiteralValue::Long(val)) => match l {
                            Some(num) => (i, Some(val + num), f, count + 1),
                            None => (i, Some(val), f, count + 1),
                        },
                        Expression::Literal(LiteralValue::Double(val)) => match f {
                            Some(num) => (i, l, Some(num + val), count + 1),
                            None => (i, l, Some(val), count + 1),
                        },
                        _ => {
                            non_literals.push(expr);
                            (i, l, f, count + 1)
                        }
                    })
            }
            ScalarFunction::Mul => {
                sf.args
                    .into_iter()
                    .fold((None, None, None, 0), |(i, l, f, count), expr| match expr {
                        Expression::Literal(LiteralValue::Integer(val)) => match i {
                            None => (Some(val), l, f, count + 1),
                            Some(num) => (Some(num * val), l, f, count + 1),
                        },
                        Expression::Literal(LiteralValue::Long(val)) => match l {
                            None => (i, Some(val), f, count + 1),
                            Some(num) => (i, Some(num * val), f, count + 1),
                        },
                        Expression::Literal(LiteralValue::Double(val)) => match f {
                            None => (i, l, Some(val), count + 1),
                            Some(num) => (i, l, Some(num * val), count + 1),
                        },
                        _ => {
                            non_literals.push(expr);
                            (i, l, f, count + 1)
                        }
                    })
            }
            _ => unreachable!("fold associative function is only called on Add and Mul"),
        };
        let literals: Vec<Expression> = vec![
            int_fold.map(|val| Expression::Literal(LiteralValue::Integer(val))),
            long_fold.map(|val| Expression::Literal(LiteralValue::Long(val))),
            float_fold.map(|val| Expression::Literal(LiteralValue::Double(val))),
        ]
        .into_iter()
        .flatten()
        .collect();
        let filter_value = match sf.function {
            ScalarFunction::Add => 0.0,
            ScalarFunction::Mul => 1.0,
            _ => unreachable!("fold associative function is only called on Add and Mul"),
        };
        let filtered_literals = literals
            .clone()
            .into_iter()
            .filter(|expr| !Self::numeric_eq(expr, filter_value))
            .collect();
        let args = [filtered_literals, non_literals].concat();
        if args.is_empty() {
            return (literals.last().unwrap().clone(), true);
        }
        if args.len() == 1 {
            return (args[0].clone(), true);
        }
        let changed = args.len() < arg_count;
        (
            Expression::ScalarFunction(ScalarFunctionApplication {
                function: sf.function,
                is_nullable: sf.is_nullable,
                args,
            }),
            changed,
        )
    }

    // Constant folds binary arithmetic functions: subtract and divide
    fn fold_binary_arithmetic_function(
        &mut self,
        function: ScalarFunction,
        left: Expression,
        right: Expression,
    ) -> Option<Expression> {
        match function {
            ScalarFunction::Sub => {
                if Self::numeric_eq(&right, 0.0) {
                    Some(left)
                } else {
                    match (&left, &right) {
                        (
                            Expression::Literal(LiteralValue::Integer(l)),
                            Expression::Literal(LiteralValue::Integer(r)),
                        ) => Some(Expression::Literal(LiteralValue::Integer(l - r))),
                        (
                            Expression::Literal(LiteralValue::Long(l)),
                            Expression::Literal(LiteralValue::Long(r)),
                        ) => Some(Expression::Literal(LiteralValue::Long(l - r))),
                        (
                            Expression::Literal(LiteralValue::Double(l)),
                            Expression::Literal(LiteralValue::Double(r)),
                        ) => Some(Expression::Literal(LiteralValue::Double(l - r))),
                        _ => None,
                    }
                }
            }
            ScalarFunction::Div => {
                if Self::numeric_eq(&right, 0.0) {
                    Some(Expression::Literal(LiteralValue::Null))
                } else if Self::numeric_eq(&right, 1.0) {
                    Some(left)
                } else {
                    match (&left, &right) {
                        (
                            Expression::Literal(LiteralValue::Integer(l)),
                            Expression::Literal(LiteralValue::Integer(r)),
                        ) => Some(Expression::Literal(LiteralValue::Integer(l / r))),
                        (
                            Expression::Literal(LiteralValue::Long(l)),
                            Expression::Literal(LiteralValue::Long(r)),
                        ) => Some(Expression::Literal(LiteralValue::Long(l / r))),
                        (
                            Expression::Literal(LiteralValue::Double(l)),
                            Expression::Literal(LiteralValue::Double(r)),
                        ) => Some(Expression::Literal(LiteralValue::Double(l / r))),
                        _ => None,
                    }
                }
            }
            _ => unreachable!("fold binary arithmetic only called on sub and div"),
        }
    }

    // Constant folds binary comparison functions
    fn fold_comparison_function(
        &mut self,
        function: ScalarFunction,
        left: Expression,
        right: Expression,
    ) -> Option<Expression> {
        use std::cmp::Ordering;
        let ord = match (&left, &right) {
            (
                Expression::Literal(LiteralValue::Boolean(l)),
                Expression::Literal(LiteralValue::Boolean(r)),
            ) => l.partial_cmp(r),
            (
                Expression::Literal(LiteralValue::Integer(l)),
                Expression::Literal(LiteralValue::Integer(r)),
            ) => l.partial_cmp(r),
            (
                Expression::Literal(LiteralValue::Long(l)),
                Expression::Literal(LiteralValue::Long(r)),
            ) => l.partial_cmp(r),
            (
                Expression::Literal(LiteralValue::Double(l)),
                Expression::Literal(LiteralValue::Double(r)),
            ) => l.partial_cmp(r),
            (
                Expression::Literal(LiteralValue::String(l)),
                Expression::Literal(LiteralValue::String(r)),
            ) => l.partial_cmp(r),
            _ => None,
        };
        ord.map(|ord_val| {
            let val = match function {
                ScalarFunction::Eq => ord_val == Ordering::Equal,
                ScalarFunction::Gt => ord_val == Ordering::Greater,
                ScalarFunction::Gte => ord_val != Ordering::Less,
                ScalarFunction::Lt => ord_val == Ordering::Less,
                ScalarFunction::Lte => ord_val != Ordering::Greater,
                ScalarFunction::Neq => ord_val != Ordering::Equal,
                _ => unreachable!("non-comparison function cannot be called"),
            };
            Expression::Literal(LiteralValue::Boolean(val))
        })
    }

    // Constant folds the between function
    fn fold_between(&mut self, sf: ScalarFunctionApplication) -> (Expression, bool) {
        assert_eq!(
            sf.args.len(),
            3,
            "between scalar function must contain 3 args"
        );
        let (arg, bottom, top) = (sf.args[0].clone(), sf.args[1].clone(), sf.args[2].clone());
        let new_sf = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            is_nullable: sf.is_nullable,
            args: vec![
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Lte,
                    is_nullable: sf.is_nullable,
                    args: vec![arg.clone(), top],
                }),
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Gte,
                    is_nullable: sf.is_nullable,
                    args: vec![arg, bottom],
                }),
            ],
        });
        let folded_expr = self.visit_expression(new_sf);
        if let Expression::Literal(_) = folded_expr {
            (folded_expr, true)
        } else {
            (Expression::ScalarFunction(sf), false)
        }
    }

    // Constant folds unary functions
    // clippy erroneously thinks these as_bytes calls are unneeded
    #[allow(clippy::needless_as_bytes)]
    fn fold_unary_function(&mut self, sf: ScalarFunctionApplication) -> (Expression, bool) {
        assert_eq!(sf.args.len(), 1, "Unary function should only have one arg");
        if self.has_null_arg(&sf.args) {
            return (Expression::Literal(LiteralValue::Null), true);
        }
        let arg = sf.args[0].clone();
        let func = sf.function;
        let sf_expr = Expression::ScalarFunction(sf);
        let folded = if let Expression::Array(ArrayExpr { ref array, .. }) = arg {
            if func == ScalarFunction::Size {
                Some(Expression::Literal(LiteralValue::Integer(
                    array.len() as i32
                )))
            } else {
                None
            }
        } else if let Expression::Literal(lit) = arg {
            match func {
                ScalarFunction::Pos => match lit {
                    LiteralValue::Integer(_) | LiteralValue::Long(_) | LiteralValue::Double(_) => {
                        Some(Expression::Literal(lit))
                    }
                    _ => None,
                },
                ScalarFunction::Neg => match lit {
                    LiteralValue::Integer(val) => {
                        Some(Expression::Literal(LiteralValue::Integer(-val)))
                    }
                    LiteralValue::Long(val) => Some(Expression::Literal(LiteralValue::Long(-val))),
                    LiteralValue::Double(val) => {
                        Some(Expression::Literal(LiteralValue::Double(-val)))
                    }
                    _ => None,
                },
                ScalarFunction::Not => {
                    if let LiteralValue::Boolean(val) = lit {
                        Some(Expression::Literal(LiteralValue::Boolean(!val)))
                    } else {
                        None
                    }
                }
                ScalarFunction::Upper => {
                    if let LiteralValue::String(val) = lit {
                        Some(Expression::Literal(LiteralValue::String(
                            val.to_ascii_uppercase(),
                        )))
                    } else {
                        None
                    }
                }
                ScalarFunction::Lower => {
                    if let LiteralValue::String(val) = lit {
                        Some(Expression::Literal(LiteralValue::String(
                            val.to_ascii_lowercase(),
                        )))
                    } else {
                        None
                    }
                }
                ScalarFunction::CharLength => {
                    if let LiteralValue::String(val) = lit {
                        Some(Expression::Literal(LiteralValue::Integer(
                            val.chars().count() as i32,
                        )))
                    } else {
                        None
                    }
                }
                ScalarFunction::OctetLength => {
                    if let LiteralValue::String(val) = lit {
                        Some(Expression::Literal(LiteralValue::Integer(val.len() as i32)))
                    } else {
                        None
                    }
                }
                ScalarFunction::BitLength => {
                    if let LiteralValue::String(val) = lit {
                        Some(Expression::Literal(LiteralValue::Integer(
                            val.len() as i32 * 8,
                        )))
                    } else {
                        None
                    }
                }
                _ => unreachable!("fold unary function is only called on Pos, Neg, Not"),
            }
        } else {
            None
        };

        match folded {
            Some(expr) => (expr, true),
            None => (sf_expr, false),
        }
    }

    // Constant folds string trim functions
    fn fold_trim_function(
        &mut self,
        function: ScalarFunction,
        substr: Expression,
        string: Expression,
    ) -> Option<Expression> {
        if let (
            Expression::Literal(LiteralValue::String(sub)),
            Expression::Literal(LiteralValue::String(st)),
        ) = (&substr, &string)
        {
            let pattern = &sub.chars().collect::<Vec<char>>()[..];
            let val = match function {
                ScalarFunction::BTrim => st.trim_matches(pattern).to_string(),
                ScalarFunction::RTrim => st.trim_end_matches(pattern).to_string(),
                ScalarFunction::LTrim => st.trim_start_matches(pattern).to_string(),
                _ => unreachable!("fold trim is only called on trim functions"),
            };
            Some(Expression::Literal(LiteralValue::String(val)))
        } else {
            None
        }
    }

    // Constant folds the substring function
    fn fold_substring_function(&mut self, sf: ScalarFunctionApplication) -> (Expression, bool) {
        use std::cmp;
        if self.has_null_arg(&sf.args) {
            return (Expression::Literal(LiteralValue::Null), true);
        }
        let (string, start, len) = if sf.args.len() == 2 {
            (
                sf.args[0].clone(),
                sf.args[1].clone(),
                Expression::Literal(LiteralValue::Integer(-1)),
            )
        } else if sf.args.len() == 3 {
            (sf.args[0].clone(), sf.args[1].clone(), sf.args[2].clone())
        } else {
            panic!("Substring must have two or three args")
        };
        if let (
            Expression::Literal(LiteralValue::String(st)),
            Expression::Literal(LiteralValue::Integer(start)),
            Expression::Literal(LiteralValue::Integer(len)),
        ) = (string, start, len)
        {
            let string_len = st.len() as i32;
            let end = if len < 0 {
                cmp::max(start, string_len)
            } else {
                start + len
            };
            if start >= string_len || end < 0 {
                (
                    Expression::Literal(LiteralValue::String("".to_string())),
                    true,
                )
            } else {
                let start = cmp::max(start, 0);
                let end = cmp::min(end, string_len);
                let len = end - start;
                let substr = st
                    .chars()
                    .skip(start as usize)
                    .take(len as usize)
                    .collect::<String>();
                (Expression::Literal(LiteralValue::String(substr)), true)
            }
        } else {
            (Expression::ScalarFunction(sf), false)
        }
    }

    // Constant folds the concat function
    fn fold_concat_function(&mut self, sf: ScalarFunctionApplication) -> (Expression, bool) {
        if sf.args.is_empty() {
            return (
                Expression::Literal(LiteralValue::String("".to_string())),
                true,
            );
        }
        if self.has_null_arg(&sf.args) {
            return (Expression::Literal(LiteralValue::Null), true);
        }
        let mut result = Vec::<Expression>::new();
        let mut changed = false;
        for expr in sf.args {
            match &expr {
                Expression::Literal(LiteralValue::String(val)) => {
                    if result.is_empty() {
                        result.push(expr);
                    } else if let Expression::Literal(LiteralValue::String(prev_val)) =
                        result.last().unwrap().clone()
                    {
                        changed = true;
                        result.pop();
                        result.push(Expression::Literal(LiteralValue::String(prev_val + val)));
                    } else {
                        result.push(expr)
                    }
                }
                _ => result.push(expr),
            }
        }
        if result.len() == 1 {
            (result[0].clone(), true)
        } else {
            (
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Concat,
                    is_nullable: sf.is_nullable,
                    args: result,
                }),
                changed,
            )
        }
    }

    // Constant folds the null if function
    fn fold_null_if_function(&mut self, sf: ScalarFunctionApplication) -> (Expression, bool) {
        assert_eq!(sf.args.len(), 2, "null if should only have two args");
        match (&sf.args[0], &sf.args[1]) {
            (Expression::Literal(l), Expression::Literal(r)) => {
                if l.eq(r) {
                    (Expression::Literal(LiteralValue::Null), true)
                } else {
                    (sf.args[0].clone(), true)
                }
            }
            _ => (Expression::ScalarFunction(sf), false),
        }
    }

    // Constant folds the computed field function
    fn fold_computed_field_function(
        &mut self,
        left: Expression,
        right: Expression,
    ) -> Option<Expression> {
        if let (
            Expression::Document(DocumentExpr { document, .. }),
            Expression::Literal(LiteralValue::String(field)),
        ) = (&left, &right)
        {
            document.get(field).cloned()
        } else {
            None
        }
    }

    // Constant folds the coalesce function
    fn fold_coalesce_function(&mut self, sf: ScalarFunctionApplication) -> (Expression, bool) {
        if sf.args.is_empty() {
            return (Expression::Literal(LiteralValue::Null), true);
        }
        let mut is_all_null = true;
        for expr in &sf.args {
            match expr.schema(self.state) {
                Err(_) => {
                    // If an Err occurs, it means there is a reference in this `expr`, so we cannot
                    // possibly know if this expression satisfies NULLISH, thus is_all_null must be
                    // set to false.
                    is_all_null = false;
                    break;
                }
                Ok(sch) => {
                    let sat = sch.satisfies(&NULLISH);
                    if sat == Satisfaction::Not {
                        return (expr.clone(), true);
                    }
                    is_all_null = is_all_null && sat == Satisfaction::Must;
                }
            }
        }
        if is_all_null {
            return (Expression::Literal(LiteralValue::Null), true);
        }
        (Expression::ScalarFunction(sf), false)
    }

    // Constant folds the merge objects function
    fn fold_merge_objects_function(&mut self, sf: ScalarFunctionApplication) -> (Expression, bool) {
        use crate::util::unique_linked_hash_map::UniqueLinkedHashMap;
        // This is one case where it is actually not correct to Error if we have duplicate keys,
        // as that is allowed in the semantics of merge_objects.
        let mut result_doc = linked_hash_map::LinkedHashMap::new();
        for (i, expr) in sf.args.clone().into_iter().enumerate() {
            if let Expression::Document(DocumentExpr { document, .. }) = expr {
                for (key, value) in document.iter() {
                    result_doc.insert(key.clone(), value.clone());
                }
            } else if result_doc.is_empty() {
                // If there is a non-constant argument and the result_doc is empty, just return the
                // input function.
                return (Expression::ScalarFunction(sf), false);
            } else {
                // If `i` is greater than 1, that means at least 2 literal documents
                // have been merged, and therefore the Expression has changed.
                return (
                    Expression::ScalarFunction(ScalarFunctionApplication {
                        function: ScalarFunction::MergeObjects,
                        is_nullable: sf.is_nullable,
                        args: [
                            vec![Expression::Document(result_doc.into())],
                            sf.args.into_iter().skip(i).collect(),
                        ]
                        .concat(),
                    }),
                    i > 1,
                );
            }
        }
        (
            Expression::Document(UniqueLinkedHashMap::from(result_doc).into()),
            true,
        )
    }

    // Constant folds the position function
    fn fold_position_function(&mut self, sf: ScalarFunctionApplication) -> (Expression, bool) {
        if self.has_null_arg(&sf.args) {
            (Expression::Literal(LiteralValue::Null), true)
        } else {
            (Expression::ScalarFunction(sf), false)
        }
    }

    // Constant folds the slice function
    fn fold_slice_function(&mut self, sf: ScalarFunctionApplication) -> (Expression, bool) {
        use std::cmp;
        if self.has_null_arg(&sf.args) {
            return (Expression::Literal(LiteralValue::Null), true);
        }
        if sf.args.len() == 2 {
            let (array, len) = (sf.args[0].clone(), sf.args[1].clone());
            if let (
                Expression::Array(ArrayExpr { array, .. }),
                Expression::Literal(LiteralValue::Integer(len)),
            ) = (array, len)
            {
                let array_len = array.len() as i32;
                return if len < 0 {
                    let len = cmp::max(0, array_len + len);
                    (
                        Expression::Array(
                            array
                                .into_iter()
                                .skip(len as usize)
                                .collect::<Vec<Expression>>()
                                .into(),
                        ),
                        true,
                    )
                } else {
                    let len = cmp::min(len, array_len);
                    (
                        Expression::Array(
                            array
                                .into_iter()
                                .take(len as usize)
                                .collect::<Vec<Expression>>()
                                .into(),
                        ),
                        true,
                    )
                };
            }
        } else if sf.args.len() == 3 {
            let (array, start, len) = (sf.args[0].clone(), sf.args[1].clone(), sf.args[2].clone());
            if let (
                Expression::Array(ArrayExpr { array, .. }),
                Expression::Literal(LiteralValue::Integer(start)),
                Expression::Literal(LiteralValue::Integer(len)),
            ) = (array, start, len)
            {
                let array_len = array.len() as i32;
                if len < 0 {
                    return (Expression::Literal(LiteralValue::Null), true);
                }
                if start >= array_len {
                    return (Expression::Array(vec![].into()), true);
                }
                let start = if start.abs() >= array_len {
                    0
                } else {
                    (start + array_len) % array_len
                };
                let len = cmp::min(len, array_len - start);
                return (
                    Expression::Array(
                        array
                            .into_iter()
                            .skip(start as usize)
                            .take(len as usize)
                            .collect::<Vec<Expression>>()
                            .into(),
                    ),
                    true,
                );
            }
        } else {
            panic!("Slice must have two or three args")
        };
        (Expression::ScalarFunction(sf), false)
    }

    // Constant folds all binary function that evaluate to null if either arg is null
    fn fold_binary_null_checked_function(
        &mut self,
        sf: ScalarFunctionApplication,
    ) -> (Expression, bool) {
        assert_eq!(sf.args.len(), 2, "Binary functions must only have two args");
        if self.has_null_arg(&sf.args) {
            return (Expression::Literal(LiteralValue::Null), true);
        }
        let (left, right) = (sf.args[0].clone(), sf.args[1].clone());
        let folded = match sf.function {
            ScalarFunction::Sub | ScalarFunction::Div => {
                self.fold_binary_arithmetic_function(sf.function, left, right)
            }
            ScalarFunction::Eq
            | ScalarFunction::Gt
            | ScalarFunction::Gte
            | ScalarFunction::Lt
            | ScalarFunction::Lte
            | ScalarFunction::Neq => self.fold_comparison_function(sf.function, left, right),
            ScalarFunction::ComputedFieldAccess => self.fold_computed_field_function(left, right),
            ScalarFunction::BTrim | ScalarFunction::LTrim | ScalarFunction::RTrim => {
                self.fold_trim_function(sf.function, left, right)
            }
            _ => unreachable!("fold binary functions is only called on binary functions"),
        };

        match folded {
            Some(expr) => (expr, true),
            None => (Expression::ScalarFunction(sf), false),
        }
    }

    // Constant folds the is expression
    fn fold_is_expr(&mut self, is_expr: IsExpr) -> (Expression, bool) {
        let schema = is_expr.expr.schema(self.state);
        match schema {
            Err(_) => (Expression::Is(is_expr), false),
            Ok(schema) => {
                let target_schema = Schema::from(is_expr.target_type);
                match target_schema {
                    Schema::Atomic(Atomic::Null) => (Expression::Is(is_expr), false),
                    _ => match schema.satisfies(&target_schema) {
                        Satisfaction::Must => {
                            (Expression::Literal(LiteralValue::Boolean(true)), true)
                        }
                        Satisfaction::Not => {
                            (Expression::Literal(LiteralValue::Boolean(false)), true)
                        }
                        Satisfaction::May => (Expression::Is(is_expr), false),
                    },
                }
            }
        }
    }

    /// Attempts to fold a Cast whose input is a constant literal into a literal
    /// of the target type.
    ///
    /// None means no conversion could occur (fall through to schema-based
    /// folding).
    ///
    /// Some(Err(())) means there was an opportunity for conversion, but it
    /// statically fails; since it would also fail at runtime, we fold to the
    /// Cast's on_error.
    ///
    /// Some(Ok(_)) means we successfully converted a literal.
    fn convert_literal(l: &LiteralValue, to: Type) -> Option<Result<Expression, ()>> {
        match to {
            Type::Boolean => Self::convert_to_boolean(l),
            Type::Int32 => Self::convert_to_int(l),
            Type::Decimal128 => Self::convert_to_decimal(l),
            Type::Double => Self::convert_to_double(l),
            Type::Int64 => Self::convert_to_long(l),
            Type::Datetime => Self::convert_to_datetime(l),
            Type::ObjectId => Self::convert_to_object_id(l),
            Type::String => Self::convert_to_string(l),
            Type::Array
            | Type::Document
            | Type::BinData
            | Type::DbPointer
            | Type::Javascript
            | Type::JavascriptWithScope
            | Type::MaxKey
            | Type::MinKey
            | Type::Null
            | Type::RegularExpression
            | Type::Symbol
            | Type::Timestamp
            | Type::Undefined => None,
        }
    }

    /// Attempts to fold a Cast whose input is a literal Array into a literal of
    /// the target type. Pre-8.3 server versions only support converting arrays
    /// to boolean (unconditionally `true`). In the future, we may expand this
    /// method to support post-8.3 conversions.
    ///
    /// This function does not do array-to-array conversion as that is done more
    /// efficiently directly in `fold_cast_expr`.
    fn convert_literal_array(_: &ArrayExpr, to: Type) -> Option<Result<Expression, ()>> {
        match to {
            Type::Boolean => Some(Ok(Expression::Literal(LiteralValue::Boolean(true)))),
            _ => None,
        }
    }

    /// Attempts to fold a Cast whose input is a literal Document into a
    /// literal of the target type. Pre-8.3 server versions only support
    /// converting documents to boolean (unconditionally `true`). In the
    /// future, we may expand this method to support post-8.3 conversions.
    ///
    /// This function does not do document-to-document conversion as that
    /// is done more efficiently directly in `fold_cast_expr`.
    fn convert_literal_document(_: &DocumentExpr, to: Type) -> Option<Result<Expression, ()>> {
        match to {
            Type::Boolean => Some(Ok(Expression::Literal(LiteralValue::Boolean(true)))),
            _ => None,
        }
    }

    fn convert_to_boolean(l: &LiteralValue) -> Option<Result<Expression, ()>> {
        match l {
            // Null literal values are handled directly by the `fold_cast_expr` method. Here, we
            // return None to indicate folding did not happen, but still could in `fold_cast_expr`.
            LiteralValue::Null => None,

            // Booleans are trivially converted to themselves.
            LiteralValue::Boolean(v) => Some(Ok(Expression::Literal(LiteralValue::Boolean(*v)))),

            // Numeric types are converted to true if non-zero, false otherwise.
            LiteralValue::Integer(v) => {
                Some(Ok(Expression::Literal(LiteralValue::Boolean(*v != 0))))
            }
            LiteralValue::Long(v) => Some(Ok(Expression::Literal(LiteralValue::Boolean(*v != 0)))),
            LiteralValue::Double(v) => {
                Some(Ok(Expression::Literal(LiteralValue::Boolean(*v != 0.0))))
            }
            LiteralValue::Decimal128(v) => Some(Ok(Expression::Literal(LiteralValue::Boolean(
                v.ne(&DECIMAL_ZERO),
            )))),

            // These types are explicitly supported by MongoDB's `$convert` expression. They all
            // unconditionally convert to true.
            LiteralValue::Binary(_)
            | LiteralValue::DateTime(_)
            | LiteralValue::JavaScriptCode(_)
            | LiteralValue::JavaScriptCodeWithScope(_)
            | LiteralValue::MaxKey
            | LiteralValue::MinKey
            | LiteralValue::ObjectId(_)
            | LiteralValue::RegularExpression(_)
            | LiteralValue::String(_)
            | LiteralValue::Timestamp(_) => {
                Some(Ok(Expression::Literal(LiteralValue::Boolean(true))))
            }

            // These types are deprecated and their convert-to-bool behavior is no longer documented
            // by MongoDB. Therefore, we do not attempt to fold them.
            LiteralValue::DbPointer(_) | LiteralValue::Symbol(_) | LiteralValue::Undefined => None,
        }
    }

    fn convert_to_int(l: &LiteralValue) -> Option<Result<Expression, ()>> {
        match l {
            // Null literal values are handled directly by the `fold_cast_expr` method. Here, we
            // return None to indicate folding did not happen, but still could in `fold_cast_expr`.
            LiteralValue::Null => None,

            // False converts to 0, true converts to 1.
            LiteralValue::Boolean(false) => Some(Ok(Expression::Literal(LiteralValue::Integer(0)))),
            LiteralValue::Boolean(true) => Some(Ok(Expression::Literal(LiteralValue::Integer(1)))),

            // Integers are trivially converted to themselves.
            LiteralValue::Integer(v) => Some(Ok(Expression::Literal(LiteralValue::Integer(*v)))),

            // Other numeric types may be converted to int if they are in range.
            LiteralValue::Long(v) => Some(
                i32::try_from(*v)
                    .map(|i| Expression::Literal(LiteralValue::Integer(i)))
                    .map_err(|_| ()),
            ),
            LiteralValue::Double(v) => {
                // MongoDB truncates when converting a double to an int.
                let truncated = v.trunc();
                if truncated.is_finite()
                    && truncated >= i32::MIN as f64
                    && truncated <= i32::MAX as f64
                {
                    Some(Ok(Expression::Literal(LiteralValue::Integer(
                        truncated as i32,
                    ))))
                } else {
                    Some(Err(()))
                }
            }
            LiteralValue::Decimal128(v) => {
                // MongoDB truncates when converting a decimal to an int.
                // Since we cannot operate on Decimal128 values directly, we convert to string and
                // truncate by taking only the part before the decimal point.
                let s = v.to_string();
                let truncated = s.split('.').next().unwrap_or(&s);
                Some(
                    i32::from_str(&truncated)
                        .map(|i| Expression::Literal(LiteralValue::Integer(i)))
                        .map_err(|_| ()),
                )
            }

            // Strings may be converted to int if they represent integral numeric values in range of
            // int32. If they are non-numeric, contain floating point values, or are out of range,
            // conversion fails.
            LiteralValue::String(v) => Some(
                i32::from_str(v)
                    .map(|i| Expression::Literal(LiteralValue::Integer(i)))
                    .map_err(|_| ()),
            ),

            // These types are not supported for conversion to int.
            LiteralValue::Binary(_)
            | LiteralValue::DateTime(_)
            | LiteralValue::JavaScriptCode(_)
            | LiteralValue::JavaScriptCodeWithScope(_)
            | LiteralValue::MaxKey
            | LiteralValue::MinKey
            | LiteralValue::ObjectId(_)
            | LiteralValue::RegularExpression(_)
            | LiteralValue::Timestamp(_)
            | LiteralValue::DbPointer(_)
            | LiteralValue::Symbol(_)
            | LiteralValue::Undefined => None,
        }
    }

    fn convert_to_decimal(l: &LiteralValue) -> Option<Result<Expression, ()>> {
        let str_to_convert = match l {
            // Null literal values are handled directly by the `fold_cast_expr` method. Here, we
            // return None to indicate folding did not happen, but still could in `fold_cast_expr`.
            LiteralValue::Null => return None,

            // False converts to 0, true converts to 1.
            LiteralValue::Boolean(false) => "0",
            LiteralValue::Boolean(true) => "1",

            // Numeric types convert to decimal without loss of precision.
            LiteralValue::Integer(v) => &format!("{v}"),
            LiteralValue::Long(v) => &format!("{v}"),
            LiteralValue::Double(v) => &format!("{v}"),

            // Decimal128s are trivially converted to themselves.
            LiteralValue::Decimal128(v) => {
                return Some(Ok(Expression::Literal(LiteralValue::Decimal128(*v))))
            }

            // Strings may be converted to decimal128 if they represent numeric values in range of
            // decimal128. If they are non-numeric or are out of range, conversion fails.
            LiteralValue::String(v) => v,

            // Dates may be converted to decimal128 by converting the number of milliseconds since
            // the epoch that corresponds to the date value to decimal128.
            LiteralValue::DateTime(v) => &format!("{}", v.timestamp_millis()),

            // These types are not supported for conversion to decimal128.
            LiteralValue::Binary(_)
            | LiteralValue::JavaScriptCode(_)
            | LiteralValue::JavaScriptCodeWithScope(_)
            | LiteralValue::MaxKey
            | LiteralValue::MinKey
            | LiteralValue::ObjectId(_)
            | LiteralValue::RegularExpression(_)
            | LiteralValue::Timestamp(_)
            | LiteralValue::DbPointer(_)
            | LiteralValue::Symbol(_)
            | LiteralValue::Undefined => return None,
        };

        Some(
            Decimal128::from_str(str_to_convert)
                .map(|dec| Expression::Literal(LiteralValue::Decimal128(dec)))
                .map_err(|_| ()),
        )
    }

    fn convert_to_double(l: &LiteralValue) -> Option<Result<Expression, ()>> {
        match l {
            // Null literal values are handled directly by the `fold_cast_expr` method. Here, we
            // return None to indicate folding did not happen, but still could in `fold_cast_expr`.
            LiteralValue::Null => None,

            // False converts to 0, true converts to 1.
            LiteralValue::Boolean(false) => {
                Some(Ok(Expression::Literal(LiteralValue::Double(0.0))))
            }
            LiteralValue::Boolean(true) => Some(Ok(Expression::Literal(LiteralValue::Double(1.0)))),

            // Doubles are trivially converted to themselves.
            LiteralValue::Double(v) => Some(Ok(Expression::Literal(LiteralValue::Double(*v)))),

            // Other numeric types may be converted to double if they are in range.
            LiteralValue::Integer(v) => {
                Some(Ok(Expression::Literal(LiteralValue::Double(*v as f64))))
            }
            LiteralValue::Long(v) => {
                // MongoDB's conversion of long to double can result in loss of precision, so we do
                // the same by using Rust's `as f64` operation which also loses precision.
                Some(Ok(Expression::Literal(LiteralValue::Double(*v as f64))))
            }
            LiteralValue::Decimal128(v) => {
                // Since we cannot operate on Decimal128 values directly, we convert to string and
                // parse as f64. When parsing an f64 value from a string, if the value exceeds the
                // range of f64, the parse will result in f64::Inf or f64::NegInf. We check if the
                // parsed value is finite and return an Err otherwise.
                let s = v.to_string();
                match f64::from_str(&s) {
                    Ok(f) if f.is_finite() => {
                        Some(Ok(Expression::Literal(LiteralValue::Double(f))))
                    }
                    _ => Some(Err(())),
                }
            }

            // Strings may be converted to double if they represent numeric values in range. If they
            // are non-numeric or are out of range conversion fails.
            LiteralValue::String(v) => {
                match v.to_ascii_lowercase().as_str() {
                    // We manually check for "nan", "inf", and "-inf" since numbers that exceed the
                    // bounds of f64 will be parsed into f64::Inf or f64::NegInf, which is not how
                    // MongoDB behaves. By manually checking these 3 special values, we can still
                    // leverage f64::from_str/f64::is_finite for values that are in-range.
                    "nan" => Some(Ok(Expression::Literal(LiteralValue::Double(f64::NAN)))),
                    "inf" => Some(Ok(Expression::Literal(LiteralValue::Double(f64::INFINITY)))),
                    "-inf" => Some(Ok(Expression::Literal(LiteralValue::Double(
                        f64::NEG_INFINITY,
                    )))),
                    v => match f64::from_str(v) {
                        Ok(f) if f.is_finite() => {
                            Some(Ok(Expression::Literal(LiteralValue::Double(f))))
                        }
                        _ => Some(Err(())),
                    },
                }
            }

            // Dates may be converted to double by converting the number of milliseconds since the
            // epoch that corresponds to the date value to double.
            LiteralValue::DateTime(v) => Some(Ok(Expression::Literal(LiteralValue::Double(
                v.timestamp_millis() as f64,
            )))),

            // These types are not supported for conversion to double.
            LiteralValue::Binary(_)
            | LiteralValue::JavaScriptCode(_)
            | LiteralValue::JavaScriptCodeWithScope(_)
            | LiteralValue::MaxKey
            | LiteralValue::MinKey
            | LiteralValue::ObjectId(_)
            | LiteralValue::RegularExpression(_)
            | LiteralValue::Timestamp(_)
            | LiteralValue::DbPointer(_)
            | LiteralValue::Symbol(_)
            | LiteralValue::Undefined => None,
        }
    }

    fn convert_to_long(l: &LiteralValue) -> Option<Result<Expression, ()>> {
        match l {
            // Null literal values are handled directly by the `fold_cast_expr` method. Here, we
            // return None to indicate folding did not happen, but still could in `fold_cast_expr`.
            LiteralValue::Null => None,

            // False converts to 0, true converts to 1.
            LiteralValue::Boolean(false) => Some(Ok(Expression::Literal(LiteralValue::Long(0)))),
            LiteralValue::Boolean(true) => Some(Ok(Expression::Literal(LiteralValue::Long(1)))),

            // Longs are trivially converted to themselves.
            LiteralValue::Long(v) => Some(Ok(Expression::Literal(LiteralValue::Long(*v)))),

            // Other numeric types may be converted to long if they are in range.
            LiteralValue::Integer(v) => {
                Some(Ok(Expression::Literal(LiteralValue::Long(*v as i64))))
            }
            LiteralValue::Double(v) => {
                // MongoDB truncates when converting a double to a long.
                let truncated = v.trunc();
                if truncated.is_finite()
                    && truncated >= i64::MIN as f64
                    && truncated <= i64::MAX as f64
                {
                    Some(Ok(Expression::Literal(LiteralValue::Long(
                        truncated as i64,
                    ))))
                } else {
                    Some(Err(()))
                }
            }
            LiteralValue::Decimal128(v) => {
                // MongoDB truncates when converting a decimal to a long.
                // Since we cannot operate on Decimal128 values directly, we convert to string and
                // truncate by taking only the part before the decimal point.
                let s = v.to_string();
                let truncated = s.split('.').next().unwrap_or(&s);
                Some(
                    i64::from_str(&truncated)
                        .map(|i| Expression::Literal(LiteralValue::Long(i)))
                        .map_err(|_| ()),
                )
            }

            // Strings may be converted to long if they represent integral numeric values in range
            // of i64. If they are non-numeric, contain floating point values, or are out of range,
            // conversion fails.
            LiteralValue::String(v) => Some(
                i64::from_str(v)
                    .map(|i| Expression::Literal(LiteralValue::Long(i)))
                    .map_err(|_| ()),
            ),

            // Dates may be converted to long by getting the number of milliseconds since the
            // epoch that corresponds to the date, which is already a long.
            LiteralValue::DateTime(v) => Some(Ok(Expression::Literal(LiteralValue::Long(
                v.timestamp_millis(),
            )))),

            // These types are not supported for conversion to long.
            LiteralValue::Binary(_)
            | LiteralValue::JavaScriptCode(_)
            | LiteralValue::JavaScriptCodeWithScope(_)
            | LiteralValue::MaxKey
            | LiteralValue::MinKey
            | LiteralValue::ObjectId(_)
            | LiteralValue::RegularExpression(_)
            | LiteralValue::Timestamp(_)
            | LiteralValue::DbPointer(_)
            | LiteralValue::Symbol(_)
            | LiteralValue::Undefined => None,
        }
    }

    fn convert_to_datetime(l: &LiteralValue) -> Option<Result<Expression, ()>> {
        match l {
            // Null literal values are handled directly by the `fold_cast_expr` method. Here, we
            // return None to indicate folding did not happen, but still could in `fold_cast_expr`.
            LiteralValue::Null => None,

            // Datetimes are trivially converted to themselves.
            LiteralValue::DateTime(v) => Some(Ok(Expression::Literal(LiteralValue::DateTime(*v)))),

            // Numbers may be converted to datetime by converting to long and then to datetime.
            // Numbers that are out of range of i64 result in an error.
            LiteralValue::Long(v) => Some(Ok(Expression::Literal(LiteralValue::DateTime(
                bson::DateTime::from_millis(*v),
            )))),
            LiteralValue::Double(v) => {
                // MongoDB truncates when converting a double to a date.
                let truncated = v.trunc();
                if truncated.is_finite()
                    && truncated >= i64::MIN as f64
                    && truncated <= i64::MAX as f64
                {
                    Some(Ok(Expression::Literal(LiteralValue::DateTime(
                        bson::DateTime::from_millis(truncated as i64),
                    ))))
                } else {
                    Some(Err(()))
                }
            }
            LiteralValue::Decimal128(v) => {
                // MongoDB truncates when converting a decimal to a date.
                // Since we cannot operate on Decimal128 values directly, we convert to string and
                // truncate by taking only the part before the decimal point.
                let s = v.to_string();
                let truncated = s.split('.').next().unwrap_or(&s);
                Some(
                    i64::from_str(&truncated)
                        .map(|i| {
                            Expression::Literal(LiteralValue::DateTime(
                                bson::DateTime::from_millis(i),
                            ))
                        })
                        .map_err(|_| ()),
                )
            }

            // ObjectIds may be converted to datetime by getting the timestamp from the ObjectId.
            LiteralValue::ObjectId(v) => Some(Ok(Expression::Literal(LiteralValue::DateTime(
                v.timestamp(),
            )))),

            // Strings can be converted into the dates that correspond to the date string.
            // Valid formats that MongoDB supports are:
            //   - "2018-03-03"
            //   - "2018-03-03T12:00:00Z"
            //   - "2018-03-03T12:00:00+0500"
            LiteralValue::String(v) => {
                // Check if string ends with timezone offset: +HHMM or -HHMM
                let has_timezone_offset = v.len() >= 5 && {
                    let last_5 = &v[v.len() - 5..];
                    let bytes = last_5.as_bytes();
                    (bytes[0] == b'+' || bytes[0] == b'-')
                        && bytes[1..].iter().all(|b| b.is_ascii_digit())
                };

                // If it ends with 'Z' or '+/-HHMM', we can attempt to parse it directly. That is,
                // if it matches the format YYYY-MM-DD(T| )HH:mm:SS((.ms)?Z|(+|-)HHMM)?, where `?`
                // denotes optional.
                let chrono_dt: Option<chrono::DateTime<Utc>> =
                    if v.ends_with('Z') || has_timezone_offset {
                        v.parse().ok()
                    } else if v.len() == 10 {
                        // If it does not end with 'Z' or '+/-HHMM' and is length 10, we assume it
                        // is in YYYY-MM-DD format. We append the T00:00:00Z time to the string so
                        // that we can attempt to parse it into a chrono::DateTime<Utc>.
                        let mut s = v.to_string();
                        s.push_str("T00:00:00Z");
                        s.parse().ok()
                    } else {
                        // If it does not end with 'Z' or '+/-HHMM' and is not length 10, we assume
                        // it is in YYYY-MM-DD(T| )HH:mm:SS(.ms)? format. We append the 'Z' to the
                        // string so we can attempt to parse it into a chrono::DateTime<Utc>.
                        let mut s = v.to_string();
                        s.push('Z');
                        s.parse().ok()
                    };

                // We cannot guarantee our format here supports everything that MongoDB supports,
                // so we instead fall back to the original cast expression when our attempt to
                // create a Date fails, and allow MongoDB to succeed or fail as it will.
                chrono_dt.map(|chrono_dt| {
                    Ok(Expression::Literal(LiteralValue::DateTime(
                        chrono_dt.into(),
                    )))
                })
            }

            // These types are not supported for conversion to datetime.
            LiteralValue::Boolean(_)
            | LiteralValue::Binary(_)
            | LiteralValue::Integer(_)
            | LiteralValue::JavaScriptCode(_)
            | LiteralValue::JavaScriptCodeWithScope(_)
            | LiteralValue::MaxKey
            | LiteralValue::MinKey
            | LiteralValue::RegularExpression(_)
            | LiteralValue::Timestamp(_)
            | LiteralValue::DbPointer(_)
            | LiteralValue::Symbol(_)
            | LiteralValue::Undefined => None,
        }
    }

    fn convert_to_object_id(l: &LiteralValue) -> Option<Result<Expression, ()>> {
        match l {
            // Null literal values are handled directly by the `fold_cast_expr` method. Here, we
            // return None to indicate folding did not happen, but still could in `fold_cast_expr`.
            LiteralValue::Null => None,

            // ObjectIds are trivially converted to themselves.
            LiteralValue::ObjectId(v) => Some(Ok(Expression::Literal(LiteralValue::ObjectId(*v)))),

            // Strings that are encoded as valid ObjectIds may be converted to ObjectIds.
            LiteralValue::String(v) => Some(
                ObjectId::parse_str(v)
                    .map(|oid| Expression::Literal(LiteralValue::ObjectId(oid)))
                    .map_err(|_| ()),
            ),

            // These types are not supported for conversion to ObjectId.
            LiteralValue::Boolean(_)
            | LiteralValue::Integer(_)
            | LiteralValue::Long(_)
            | LiteralValue::Double(_)
            | LiteralValue::RegularExpression(_)
            | LiteralValue::JavaScriptCode(_)
            | LiteralValue::JavaScriptCodeWithScope(_)
            | LiteralValue::Timestamp(_)
            | LiteralValue::Binary(_)
            | LiteralValue::DateTime(_)
            | LiteralValue::Symbol(_)
            | LiteralValue::Decimal128(_)
            | LiteralValue::Undefined
            | LiteralValue::MaxKey
            | LiteralValue::MinKey
            | LiteralValue::DbPointer(_) => None,
        }
    }

    fn convert_to_string(l: &LiteralValue) -> Option<Result<Expression, ()>> {
        match l {
            // Null literal values are handled directly by the `fold_cast_expr` method. Here, we
            // return None to indicate folding did not happen, but still could in `fold_cast_expr`.
            LiteralValue::Null => None,

            // Strings are trivially converted to themselves.
            LiteralValue::String(v) => {
                Some(Ok(Expression::Literal(LiteralValue::String(v.clone()))))
            }

            // A subset of types is supported by MongoDB pre-8.3, so we constrain ourselves to those
            // types here.
            LiteralValue::Boolean(v) => {
                Some(Ok(Expression::Literal(LiteralValue::String(v.to_string()))))
            }
            LiteralValue::Integer(v) => {
                Some(Ok(Expression::Literal(LiteralValue::String(v.to_string()))))
            }
            LiteralValue::Long(v) => {
                Some(Ok(Expression::Literal(LiteralValue::String(v.to_string()))))
            }
            LiteralValue::Double(v) => {
                Some(Ok(Expression::Literal(LiteralValue::String(v.to_string()))))
            }
            LiteralValue::Decimal128(v) => {
                Some(Ok(Expression::Literal(LiteralValue::String(v.to_string()))))
            }
            LiteralValue::ObjectId(v) => {
                Some(Ok(Expression::Literal(LiteralValue::String(v.to_string()))))
            }
            LiteralValue::DateTime(v) => {
                Some(Ok(Expression::Literal(LiteralValue::String(v.to_string()))))
            }

            // These types are not supported for conversion to string.
            LiteralValue::RegularExpression(_)
            | LiteralValue::JavaScriptCode(_)
            | LiteralValue::JavaScriptCodeWithScope(_)
            | LiteralValue::Timestamp(_)
            | LiteralValue::Binary(_)
            | LiteralValue::Symbol(_)
            | LiteralValue::Undefined
            | LiteralValue::MaxKey
            | LiteralValue::MinKey
            | LiteralValue::DbPointer(_) => None,
        }
    }

    // Constant folds the cast expression
    fn fold_cast_expr(&mut self, cast_expr: CastExpr) -> (Expression, bool) {
        use crate::schema::{ANY_ARRAY, ANY_DOCUMENT};

        // If the input is a constant literal, attempt to fold the cast into a
        // literal of the target type. A static conversion failure folds to the
        // Cast's on_error, matching runtime CAST semantics.
        let conversion_result = match cast_expr.expr.as_ref() {
            Expression::Literal(ref l) => Self::convert_literal(l, cast_expr.to),
            Expression::Array(ref a) => Self::convert_literal_array(a, cast_expr.to),
            Expression::Document(ref d) => Self::convert_literal_document(d, cast_expr.to),
            _ => None,
        };

        match conversion_result {
            Some(Ok(folded)) => return (folded, true),
            Some(Err(())) => return (*cast_expr.on_error, true),
            None => {}
        }

        let schema = cast_expr.expr.schema(self.state);
        let Ok(schema) = schema else {
            return (Expression::Cast(cast_expr), false);
        };
        let target_schema = Schema::from(cast_expr.to);
        let sat = schema.satisfies(&target_schema);
        if self.schema_is_exactly_nullish(schema) {
            (*cast_expr.on_null, true)
        } else if sat == Satisfaction::Must {
            (*cast_expr.expr, true)
        } else if sat == Satisfaction::Not
            && (target_schema == ANY_ARRAY.clone() || target_schema == ANY_DOCUMENT.clone())
        {
            (*cast_expr.on_error, true)
        } else {
            (Expression::Cast(cast_expr), false)
        }
    }

    // Folds the simple case expression
    fn fold_simple_case_expr(&mut self, case_expr: SimpleCaseExpr) -> (Expression, bool) {
        let mut new_case_branches: Vec<WhenBranch> = vec![];
        let mut changed = false;
        for when_branch in case_expr.when_branch.clone() {
            if case_expr.expr.eq(&when_branch.when) && new_case_branches.is_empty() {
                return (*when_branch.then, true);
            }
            match (&*case_expr.expr, &*when_branch.when) {
                (Expression::Literal(_), Expression::Literal(_)) => {
                    // Only retain Literal-Literal comparison branches if we
                    // know they are equal. We cannot simplify to this branch's
                    // then value since it is possible an earlier, non-literal
                    // branch will match the case_expr at runtime.
                    if case_expr.expr.eq(&when_branch.when) {
                        new_case_branches.push(when_branch)
                    } else {
                        changed = true;
                    }
                }
                _ => new_case_branches.push(when_branch),
            }
        }
        if new_case_branches.is_empty() {
            (*case_expr.else_branch, true)
        } else {
            (
                Expression::SimpleCase(SimpleCaseExpr {
                    expr: case_expr.expr,
                    when_branch: new_case_branches,
                    else_branch: case_expr.else_branch,
                    is_nullable: case_expr.is_nullable,
                }),
                changed,
            )
        }
    }

    // Folds the searched case expression
    fn fold_searched_case_expr(&mut self, case_expr: SearchedCaseExpr) -> (Expression, bool) {
        let mut new_case_branches: Vec<WhenBranch> = vec![];
        let mut changed = false;
        for when_branch in case_expr.when_branch.clone() {
            match &*when_branch.when {
                Expression::Literal(lit) => {
                    if lit.eq(&LiteralValue::Boolean(true)) {
                        if new_case_branches.is_empty() {
                            return (*when_branch.then, true);
                        } else {
                            new_case_branches.push(when_branch)
                        }
                    } else {
                        changed = true;
                    }
                }
                _ => new_case_branches.push(when_branch),
            }
        }
        if new_case_branches.is_empty() {
            (*case_expr.else_branch, true)
        } else {
            (
                Expression::SearchedCase(SearchedCaseExpr {
                    when_branch: new_case_branches,
                    else_branch: case_expr.else_branch,
                    is_nullable: case_expr.is_nullable,
                }),
                changed,
            )
        }
    }

    // Folds the field access expression
    fn fold_field_access_expr(&mut self, field_expr: FieldAccess) -> (Expression, bool) {
        match *field_expr.expr {
            Expression::Document(DocumentExpr { ref document, .. }) => {
                if !document.contains_key(&field_expr.field) {
                    return (Expression::FieldAccess(field_expr), false);
                }
                if let Expression::Document(DocumentExpr { mut document, .. }) = *field_expr.expr {
                    (document.remove(&field_expr.field).unwrap(), true)
                } else {
                    unreachable!()
                }
            }
            _ => (Expression::FieldAccess(field_expr), false),
        }
    }

    // Folds the filter stage
    fn fold_filter_stage(&mut self, filter_stage: Filter) -> (Stage, bool) {
        if let Expression::Literal(LiteralValue::Boolean(val)) = filter_stage.condition {
            if val {
                return (*filter_stage.source, true);
            }
        }
        (Stage::Filter(filter_stage), false)
    }

    // Folds the offset stage
    fn fold_offset_stage(&mut self, offset_stage: Offset) -> (Stage, bool) {
        if offset_stage.offset == 0 {
            (*offset_stage.source, true)
        } else {
            (Stage::Offset(offset_stage), false)
        }
    }
}

impl Visitor for ConstantFoldExprVisitor<'_> {
    fn visit_expression(&mut self, e: Expression) -> Expression {
        let e = e.walk(self);
        let (folded, changed) = match e {
            Expression::Array(_) => (e, false),
            Expression::Cast(cast_expr) => self.fold_cast_expr(cast_expr),
            Expression::Document(_) => (e, false),
            Expression::Exists(_) => (e, false),
            Expression::FieldAccess(field_expr) => self.fold_field_access_expr(field_expr),
            Expression::Is(is_expr) => self.fold_is_expr(is_expr),
            Expression::Like(_) => (e, false),
            Expression::Literal(_) => (e, false),
            Expression::Reference(_) => (e, false),
            Expression::DateFunction(_) => (e, false),
            Expression::ScalarFunction(f) => match f.function {
                ScalarFunction::Add | ScalarFunction::Mul => {
                    self.fold_associative_arithmetic_function(f)
                }
                ScalarFunction::Sub
                | ScalarFunction::Div
                | ScalarFunction::Eq
                | ScalarFunction::Gt
                | ScalarFunction::Gte
                | ScalarFunction::Lt
                | ScalarFunction::Lte
                | ScalarFunction::Neq
                | ScalarFunction::BTrim
                | ScalarFunction::LTrim
                | ScalarFunction::RTrim
                | ScalarFunction::ComputedFieldAccess => self.fold_binary_null_checked_function(f),
                ScalarFunction::And | ScalarFunction::Or => self.fold_logical_function(f),
                ScalarFunction::Between => self.fold_between(f),
                ScalarFunction::Neg
                | ScalarFunction::Not
                | ScalarFunction::Pos
                | ScalarFunction::Upper
                | ScalarFunction::Lower
                | ScalarFunction::BitLength
                | ScalarFunction::CharLength
                | ScalarFunction::OctetLength
                | ScalarFunction::Size => self.fold_unary_function(f),
                ScalarFunction::Substring => self.fold_substring_function(f),
                ScalarFunction::Concat => self.fold_concat_function(f),
                ScalarFunction::Coalesce => self.fold_coalesce_function(f),
                ScalarFunction::MergeObjects => self.fold_merge_objects_function(f),
                ScalarFunction::NullIf => self.fold_null_if_function(f),
                ScalarFunction::Slice => self.fold_slice_function(f),
                ScalarFunction::Position => self.fold_position_function(f),
                _ => (Expression::ScalarFunction(f), false),
            },
            Expression::SearchedCase(case_expr) => self.fold_searched_case_expr(case_expr),
            Expression::SimpleCase(case_expr) => self.fold_simple_case_expr(case_expr),
            Expression::SubqueryComparison(_) => (e, false),
            Expression::Subquery(_) => (e, false),
            Expression::TypeAssertion(_) => (e, false),
            Expression::HigherOrderFunction(_) => (e, false),
            Expression::Variable(_) => (e, false),
            Expression::MqlIntrinsicFieldExistence(f) => {
                // Patrick: we clone in case somehow the field access is mutated into a non-field access by
                // fold_field_access_expr. This would imply a bug in our code generation, *I
                // think*, but I am doing this rather than panicking, just to be safe, since I do
                // not have a proof of this. The clone should be relatively cheap.
                // Initially, I had a panic, and all of our query tests still pass, but I am not
                // condifident on removing this clone and possibly breaking some customer query.
                if let (Expression::FieldAccess(fa), changed) =
                    self.fold_field_access_expr(f.clone())
                {
                    (Expression::MqlIntrinsicFieldExistence(fa), changed)
                } else {
                    (Expression::MqlIntrinsicFieldExistence(f), false)
                }
            }
        };

        self.changed |= changed;
        folded
    }

    fn visit_stage(&mut self, st: Stage) -> Stage {
        let st = st.walk(self);
        let (folded, changed) = match st {
            Stage::Array(_) => (st, false),
            Stage::Collection(_) => (st, false),
            Stage::Filter(filter) => self.fold_filter_stage(filter),
            Stage::Group(_) => (st, false),
            Stage::Join(_) => (st, false),
            Stage::Limit(_) => (st, false),
            Stage::Offset(offset) => self.fold_offset_stage(offset),
            Stage::Project(_) => (st, false),
            Stage::Set(_) => (st, false),
            Stage::Sort(_) => (st, false),
            Stage::Derived(_) => (st, false),
            Stage::Unwind(_) => (st, false),
            Stage::MqlIntrinsic(_) => (st, false),
            Stage::Sentinel => unreachable!(),
        };

        self.changed |= changed;
        folded
    }
}
