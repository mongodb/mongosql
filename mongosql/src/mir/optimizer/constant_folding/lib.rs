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
use lazy_static::lazy_static;

#[derive(Clone)]
pub(crate) struct ConstantFoldExprVisitor<'a> {
    pub(crate) state: SchemaInferenceState<'a>,
}

lazy_static! {
    static ref DEFAULT_CATALOG: Catalog = Catalog::default();
}

impl<'a> ConstantFoldExprVisitor<'a> {
    // Checks if a vector of expressions contains a null or missing expression
    fn has_null_arg(&self, args: &[Expression]) -> bool {
        for expr in args {
            match expr.schema(&self.state) {
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
    fn fold_logical_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        let (nullish, non_nullish): (Vec<Expression>, Vec<Expression>) =
            sf.args.clone().into_iter().partition(|e| {
                let schema = e.schema(&self.state).unwrap_or(Schema::Any);
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
            return folded_expr;
        }
        if has_null {
            non_literals.push(Expression::Literal(LiteralValue::Null))
        }
        if non_literals.len() == 1 {
            return non_literals[0].clone();
        }
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: sf.function,
            is_nullable: sf.is_nullable,
            args: non_literals,
        })
    }

    // Constant folds constants of the same type within an associative arithmetic function
    fn fold_associative_arithmetic_function(
        &mut self,
        sf: ScalarFunctionApplication,
    ) -> Expression {
        if self.has_null_arg(&sf.args) {
            return Expression::Literal(LiteralValue::Null);
        }
        if sf.args.is_empty() {
            match sf.function {
                ScalarFunction::Add => return Expression::Literal(LiteralValue::Integer(0)),
                ScalarFunction::Mul => return Expression::Literal(LiteralValue::Integer(1)),
                _ => unreachable!("fold associative function only called on Add and Mul"),
            }
        }
        let mut non_literals = Vec::<Expression>::new();
        let (int_fold, long_fold, float_fold) = match sf.function {
            ScalarFunction::Add => {
                sf.args
                    .into_iter()
                    .fold((None, None, None), |(i, l, f), expr| match expr {
                        Expression::Literal(LiteralValue::Integer(val)) => match i {
                            Some(num) => (Some(val + num), l, f),
                            None => (Some(val), l, f),
                        },
                        Expression::Literal(LiteralValue::Long(val)) => match l {
                            Some(num) => (i, Some(val + num), f),
                            None => (i, Some(val), f),
                        },
                        Expression::Literal(LiteralValue::Double(val)) => match f {
                            Some(num) => (i, l, Some(num + val)),
                            None => (i, l, Some(val)),
                        },
                        _ => {
                            non_literals.push(expr);
                            (i, l, f)
                        }
                    })
            }
            ScalarFunction::Mul => {
                sf.args
                    .into_iter()
                    .fold((None, None, None), |(i, l, f), expr| match expr {
                        Expression::Literal(LiteralValue::Integer(val)) => match i {
                            None => (Some(val), l, f),
                            Some(num) => (Some(num * val), l, f),
                        },
                        Expression::Literal(LiteralValue::Long(val)) => match l {
                            None => (i, Some(val), f),
                            Some(num) => (i, Some(num * val), f),
                        },
                        Expression::Literal(LiteralValue::Double(val)) => match f {
                            None => (i, l, Some(val)),
                            Some(num) => (i, l, Some(num * val)),
                        },
                        _ => {
                            non_literals.push(expr);
                            (i, l, f)
                        }
                    })
            }
            _ => unreachable!("fold associative function is only called on And and Mul"),
        };
        let literals: Vec<Expression> = vec![
            int_fold.map(|val| Expression::Literal(LiteralValue::Integer(val))),
            long_fold.map(|val| Expression::Literal(LiteralValue::Long(val))),
            float_fold.map(|val| Expression::Literal(LiteralValue::Double(val))),
        ]
        .into_iter()
        .flatten()
        .collect();
        let filtered_literals: Vec<Expression> = match sf.function {
            ScalarFunction::Add => literals
                .clone()
                .into_iter()
                .filter(|expr| !Self::numeric_eq(expr, 0.0))
                .collect(),
            ScalarFunction::Mul => literals
                .clone()
                .into_iter()
                .filter(|expr| !Self::numeric_eq(expr, 1.0))
                .collect(),
            _ => unreachable!("fold associative function is only called on And and Mul"),
        };
        let args = [filtered_literals, non_literals].concat();
        if args.is_empty() {
            return literals.last().unwrap().clone();
        }
        if args.len() == 1 {
            return args[0].clone();
        }
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: sf.function,
            is_nullable: sf.is_nullable,
            args,
        })
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
    fn fold_between(&mut self, sf: ScalarFunctionApplication) -> Expression {
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
            folded_expr
        } else {
            Expression::ScalarFunction(sf)
        }
    }

    // Constant folds unary functions
    fn fold_unary_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        assert_eq!(sf.args.len(), 1, "Unary function should only have one arg");
        if self.has_null_arg(&sf.args) {
            return Expression::Literal(LiteralValue::Null);
        }
        let arg = sf.args[0].clone();
        let func = sf.function;
        let sf_expr = Expression::ScalarFunction(sf);
        if let Expression::Array(ArrayExpr { ref array, .. }) = arg {
            if func == ScalarFunction::Size {
                Expression::Literal(LiteralValue::Integer(array.len() as i32))
            } else {
                sf_expr
            }
        } else if let Expression::Literal(lit) = arg {
            match func {
                ScalarFunction::Pos => match lit {
                    LiteralValue::Integer(_) | LiteralValue::Long(_) | LiteralValue::Double(_) => {
                        Expression::Literal(lit)
                    }
                    _ => sf_expr,
                },
                ScalarFunction::Neg => match lit {
                    LiteralValue::Integer(val) => Expression::Literal(LiteralValue::Integer(-val)),
                    LiteralValue::Long(val) => Expression::Literal(LiteralValue::Long(-val)),
                    LiteralValue::Double(val) => Expression::Literal(LiteralValue::Double(-val)),
                    _ => sf_expr,
                },
                ScalarFunction::Not => {
                    if let LiteralValue::Boolean(val) = lit {
                        Expression::Literal(LiteralValue::Boolean(!val))
                    } else {
                        sf_expr
                    }
                }
                ScalarFunction::Upper => {
                    if let LiteralValue::String(val) = lit {
                        Expression::Literal(LiteralValue::String(val.to_ascii_uppercase()))
                    } else {
                        sf_expr
                    }
                }
                ScalarFunction::Lower => {
                    if let LiteralValue::String(val) = lit {
                        Expression::Literal(LiteralValue::String(val.to_ascii_lowercase()))
                    } else {
                        sf_expr
                    }
                }
                ScalarFunction::CharLength => {
                    if let LiteralValue::String(val) = lit {
                        Expression::Literal(LiteralValue::Integer(val.chars().count() as i32))
                    } else {
                        sf_expr
                    }
                }
                ScalarFunction::OctetLength => {
                    if let LiteralValue::String(val) = lit {
                        Expression::Literal(LiteralValue::Integer(val.bytes().len() as i32))
                    } else {
                        sf_expr
                    }
                }
                ScalarFunction::BitLength => {
                    if let LiteralValue::String(val) = lit {
                        Expression::Literal(LiteralValue::Integer(val.bytes().len() as i32 * 8))
                    } else {
                        sf_expr
                    }
                }
                _ => unreachable!("fold unary function is only called on Pos, Neg, Not"),
            }
        } else {
            sf_expr
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
    fn fold_substring_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        use std::cmp;
        if self.has_null_arg(&sf.args) {
            return Expression::Literal(LiteralValue::Null);
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
                Expression::Literal(LiteralValue::String("".to_string()))
            } else {
                let start = cmp::max(start, 0);
                let end = cmp::min(end, string_len);
                let len = end - start;
                let substr = st
                    .chars()
                    .skip(start as usize)
                    .take(len as usize)
                    .collect::<String>();
                Expression::Literal(LiteralValue::String(substr))
            }
        } else {
            Expression::ScalarFunction(sf)
        }
    }

    // Constant folds the concat function
    fn fold_concat_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        if sf.args.is_empty() {
            return Expression::Literal(LiteralValue::String("".to_string()));
        }
        if self.has_null_arg(&sf.args) {
            return Expression::Literal(LiteralValue::Null);
        }
        let mut result = Vec::<Expression>::new();
        for expr in sf.args {
            match &expr {
                Expression::Literal(LiteralValue::String(val)) => {
                    if result.is_empty() {
                        result.push(expr);
                    } else if let Expression::Literal(LiteralValue::String(prev_val)) =
                        result.last().unwrap().clone()
                    {
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
            result[0].clone()
        } else {
            Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::Concat,
                is_nullable: sf.is_nullable,
                args: result,
            })
        }
    }

    // Constant folds the null if function
    fn fold_null_if_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        assert_eq!(sf.args.len(), 2, "null if should only have two args");
        match (&sf.args[0], &sf.args[1]) {
            (Expression::Literal(l), Expression::Literal(r)) => {
                if l.eq(r) {
                    Expression::Literal(LiteralValue::Null)
                } else {
                    sf.args[0].clone()
                }
            }
            _ => Expression::ScalarFunction(sf),
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
    fn fold_coalesce_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        if sf.args.is_empty() {
            return Expression::Literal(LiteralValue::Null);
        }
        let mut is_all_null = true;
        for expr in &sf.args {
            match expr.schema(&self.state) {
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
                        return expr.clone();
                    }
                    is_all_null = is_all_null && sat == Satisfaction::Must;
                }
            }
        }
        if is_all_null {
            return Expression::Literal(LiteralValue::Null);
        }
        Expression::ScalarFunction(sf)
    }

    // Constant folds the merge objects function
    fn fold_merge_objects_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        use crate::util::unique_linked_hash_map::UniqueLinkedHashMap;
        // This is one case where it is actually not correct to Error if we have duplicate keys,
        // as that is allowed in the is_nullable of merge_objects.
        let mut result_doc = linked_hash_map::LinkedHashMap::new();
        for (i, expr) in sf.args.clone().into_iter().enumerate() {
            if let Expression::Document(DocumentExpr { document, .. }) = expr {
                for (key, value) in document.iter() {
                    result_doc.insert(key.clone(), value.clone());
                }
            } else if result_doc.is_empty() {
                // If there is a non-constant argument and the result_doc is empty, just return the
                // input function.
                return Expression::ScalarFunction(sf);
            } else {
                return Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::MergeObjects,
                    is_nullable: sf.is_nullable,
                    args: [
                        vec![Expression::Document(result_doc.into())],
                        sf.args.into_iter().skip(i).collect(),
                    ]
                    .concat(),
                });
            }
        }
        Expression::Document(UniqueLinkedHashMap::from(result_doc).into())
    }

    // Constant folds the position function
    fn fold_position_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        if self.has_null_arg(&sf.args) {
            Expression::Literal(LiteralValue::Null)
        } else {
            Expression::ScalarFunction(sf)
        }
    }

    // Constant folds the slice function
    fn fold_slice_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        use std::cmp;
        if self.has_null_arg(&sf.args) {
            return Expression::Literal(LiteralValue::Null);
        }
        if sf.args.len() == 2 {
            let (array, len) = (sf.args[0].clone(), sf.args[1].clone());
            if let (
                Expression::Array(ArrayExpr { array, .. }),
                Expression::Literal(LiteralValue::Integer(len)),
            ) = (array, len)
            {
                let array_len = array.len() as i32;
                if len < 0 {
                    let len = cmp::max(0, array_len + len);
                    return Expression::Array(
                        array
                            .into_iter()
                            .skip(len as usize)
                            .collect::<Vec<Expression>>()
                            .into(),
                    );
                } else {
                    let len = cmp::min(len, array_len);
                    return Expression::Array(
                        array
                            .into_iter()
                            .take(len as usize)
                            .collect::<Vec<Expression>>()
                            .into(),
                    );
                }
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
                    return Expression::Literal(LiteralValue::Null);
                }
                if start >= array_len {
                    return Expression::Array(vec![].into());
                }
                let start = if start.abs() >= array_len {
                    0
                } else {
                    (start + array_len) % array_len
                };
                let len = cmp::min(len, array_len - start);
                return Expression::Array(
                    array
                        .into_iter()
                        .skip(start as usize)
                        .take(len as usize)
                        .collect::<Vec<Expression>>()
                        .into(),
                );
            }
        } else {
            panic!("Slice must have two or three args")
        };
        Expression::ScalarFunction(sf)
    }

    // Constant folds all binary function that evaluate to null if either arg is null
    fn fold_binary_null_checked_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        assert_eq!(sf.args.len(), 2, "Binary functions must only have two args");
        if self.has_null_arg(&sf.args) {
            return Expression::Literal(LiteralValue::Null);
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

        folded.unwrap_or(Expression::ScalarFunction(sf))
    }

    // Constant folds the is expression
    fn fold_is_expr(&mut self, is_expr: IsExpr) -> Expression {
        let schema = is_expr.expr.schema(&self.state);
        match schema {
            Err(_) => Expression::Is(is_expr),
            Ok(schema) => {
                let target_schema = Schema::from(is_expr.target_type);
                match target_schema {
                    Schema::Atomic(Atomic::Null) => Expression::Is(is_expr),
                    _ => match schema.satisfies(&target_schema) {
                        Satisfaction::Must => Expression::Literal(LiteralValue::Boolean(true)),
                        Satisfaction::Not => Expression::Literal(LiteralValue::Boolean(false)),
                        Satisfaction::May => Expression::Is(is_expr),
                    },
                }
            }
        }
    }

    // Constant folds the cast expression
    fn fold_cast_expr(&mut self, cast_expr: CastExpr) -> Expression {
        use crate::schema::{ANY_ARRAY, ANY_DOCUMENT};
        let schema = cast_expr.expr.schema(&self.state);
        if schema.is_err() {
            return Expression::Cast(cast_expr);
        }
        let target_schema = Schema::from(cast_expr.to);
        let schema = schema.unwrap();
        let sat = schema.satisfies(&target_schema);
        if self.schema_is_exactly_nullish(schema) {
            *cast_expr.on_null
        } else if sat == Satisfaction::Must {
            *cast_expr.expr
        } else if sat == Satisfaction::Not
            && (target_schema == ANY_ARRAY.clone() || target_schema == ANY_DOCUMENT.clone())
        {
            *cast_expr.on_error
        } else {
            Expression::Cast(cast_expr)
        }
    }

    // Folds the simple case expression
    fn fold_simple_case_expr(&mut self, case_expr: SimpleCaseExpr) -> Expression {
        let mut new_case_branches: Vec<WhenBranch> = vec![];
        for when_branch in case_expr.when_branch.clone() {
            if case_expr.expr.eq(&when_branch.when) && new_case_branches.is_empty() {
                return *when_branch.then;
            }
            match (&*case_expr.expr, &*when_branch.when) {
                (Expression::Literal(_), Expression::Literal(_)) => {
                    if case_expr.expr.eq(&when_branch.when) {
                        new_case_branches.push(when_branch)
                    }
                }
                _ => new_case_branches.push(when_branch),
            }
        }
        if new_case_branches.is_empty() {
            *case_expr.else_branch
        } else {
            Expression::SimpleCase(SimpleCaseExpr {
                expr: case_expr.expr,
                when_branch: new_case_branches,
                else_branch: case_expr.else_branch,
                is_nullable: case_expr.is_nullable,
            })
        }
    }

    // Folds the searched case expression
    fn fold_searched_case_expr(&mut self, case_expr: SearchedCaseExpr) -> Expression {
        let mut new_case_branches: Vec<WhenBranch> = vec![];
        for when_branch in case_expr.when_branch.clone() {
            match &*when_branch.when {
                Expression::Literal(lit) => {
                    if lit.eq(&LiteralValue::Boolean(true)) {
                        if new_case_branches.is_empty() {
                            return *when_branch.then;
                        } else {
                            new_case_branches.push(when_branch)
                        }
                    }
                }
                _ => new_case_branches.push(when_branch),
            }
        }
        if new_case_branches.is_empty() {
            *case_expr.else_branch
        } else {
            Expression::SearchedCase(SearchedCaseExpr {
                when_branch: new_case_branches,
                else_branch: case_expr.else_branch,
                is_nullable: case_expr.is_nullable,
            })
        }
    }

    // Folds the field access expression
    fn fold_field_access_expr(&mut self, field_expr: FieldAccess) -> Expression {
        let expr = field_expr.clone();
        if let Expression::Document(DocumentExpr { ref document, .. }) = *field_expr.expr {
            let res = document.get(&field_expr.field);
            return res.unwrap_or(&Expression::FieldAccess(expr)).clone();
        }
        Expression::FieldAccess(field_expr)
    }

    // Folds the filter stage
    fn fold_filter_stage(&mut self, filter_stage: Filter) -> Stage {
        if let Expression::Literal(LiteralValue::Boolean(val)) = filter_stage.condition {
            if val {
                return *filter_stage.source;
            }
        }
        Stage::Filter(filter_stage)
    }

    // Folds the offset stage
    fn fold_offset_stage(&mut self, offset_stage: Offset) -> Stage {
        if offset_stage.offset == 0 {
            *offset_stage.source
        } else {
            Stage::Offset(offset_stage)
        }
    }
}

impl<'a> Visitor for ConstantFoldExprVisitor<'a> {
    fn visit_expression(&mut self, e: Expression) -> Expression {
        let e = e.walk(self);
        match e {
            Expression::Array(_) => e,
            Expression::Cast(cast_expr) => self.fold_cast_expr(cast_expr),
            Expression::Document(_) => e,
            Expression::Exists(_) => e,
            Expression::FieldAccess(field_expr) => self.fold_field_access_expr(field_expr),
            Expression::Is(is_expr) => self.fold_is_expr(is_expr),
            Expression::Like(_) => e,
            Expression::Literal(_) => e,
            Expression::Reference(_) => e,
            Expression::DateFunction(_) => e,
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
                _ => Expression::ScalarFunction(f),
            },
            Expression::SearchedCase(case_expr) => self.fold_searched_case_expr(case_expr),
            Expression::SimpleCase(case_expr) => self.fold_simple_case_expr(case_expr),
            Expression::SubqueryComparison(_) => e,
            Expression::Subquery(_) => e,
            Expression::TypeAssertion(_) => e,
            Expression::MQLIntrinsicFieldExistence(_) => e,
        }
    }

    fn visit_stage(&mut self, st: Stage) -> Stage {
        let st = st.walk(self);
        match st {
            Stage::Array(_) => st,
            Stage::Collection(_) => st,
            Stage::Filter(filter) => self.fold_filter_stage(filter),
            Stage::Group(_) => st,
            Stage::Join(_) => st,
            Stage::Limit(_) => st,
            Stage::Offset(offset) => self.fold_offset_stage(offset),
            Stage::Project(_) => st,
            Stage::Set(_) => st,
            Stage::Sort(_) => st,
            Stage::Derived(_) => st,
            Stage::Unwind(_) => st,
            Stage::MQLIntrinsic(_) => st,
            Stage::Sentinel => unreachable!(),
        }
    }
}
