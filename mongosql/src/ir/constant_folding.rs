use crate::{
    ir::{definitions::*, schema::SchemaInferenceState, visitor::Visitor},
    schema::{Atomic, Satisfaction, Schema, SchemaEnvironment},
    set,
};
use lazy_static::lazy_static;

#[derive(Default)]
struct ConstantFoldExprVisitor;

lazy_static! {
    static ref EMPTY_STATE: SchemaInferenceState = SchemaInferenceState {
        scope_level: 0u16,
        env: SchemaEnvironment::default(),
    };
}

impl ConstantFoldExprVisitor {
    // Checks if a vector of expressions contains a null or missing expression
    fn has_null_arg(args: &[Expression]) -> bool {
        for expr in args {
            match expr.schema(&EMPTY_STATE) {
                Err(_) => return false,
                Ok(sch) => {
                    if sch.satisfies(&Schema::AnyOf(set![
                        Schema::Missing,
                        Schema::Atomic(Atomic::Null),
                    ])) == Satisfaction::Must
                    {
                        return true;
                    }
                }
            }
        }
        false
    }

    // This is not a general purpose function and is not capable of checking equality of very
    // large longs. It is used to check special arithmetic edge cases like 0 and 1.
    fn numeric_eq(expr: &Expression, num: f64) -> bool {
        match expr {
            Expression::Literal(Literal::Integer(val)) => *val == num as i32,
            Expression::Literal(Literal::Long(val)) => *val == num as i64,
            Expression::Literal(Literal::Double(val)) => *val == num,
            _ => false,
        }
    }

    // Constant folds boolean functions
    fn fold_logical_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        let (nullish, non_nullish): (Vec<Expression>, Vec<Expression>) =
            sf.args.clone().into_iter().partition(|e| {
                e.schema(&EMPTY_STATE)
                    .unwrap_or(Schema::Any)
                    .satisfies(&Schema::AnyOf(set![
                        Schema::Missing,
                        Schema::Atomic(Atomic::Null),
                    ]))
                    == Satisfaction::Must
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
                Expression::Literal(Literal::Boolean(val)) => op(acc, val),
                expr => {
                    non_literals.push(expr);
                    acc
                }
            });
        let folded_expr = Expression::Literal(Literal::Boolean(folded_constant));
        if non_literals.is_empty() && !has_null {
            return folded_expr;
        }
        match sf.function {
            ScalarFunction::And => {
                if !folded_constant {
                    return Expression::Literal(Literal::Boolean(false));
                }
            }
            ScalarFunction::Or => {
                if folded_constant {
                    return Expression::Literal(Literal::Boolean(true));
                }
            }
            _ => unreachable!("fold logical functions is only called on And and Or"),
        };
        let args = if has_null {
            [vec![Expression::Literal(Literal::Null)], non_literals].concat()
        } else {
            [vec![folded_expr], non_literals].concat()
        };
        if args.len() == 1 {
            return args[0].clone();
        }
        Expression::ScalarFunction(ScalarFunctionApplication {
            function: sf.function,
            args,
        })
    }

    // Constant folds constants of the same type within an associative arithmetic function
    fn fold_associative_arithmetic_function(
        &mut self,
        sf: ScalarFunctionApplication,
    ) -> Expression {
        if Self::has_null_arg(&sf.args) {
            return Expression::Literal(Literal::Null);
        }
        if sf.args.is_empty() {
            match sf.function {
                ScalarFunction::Add => return Expression::Literal(Literal::Integer(0)),
                ScalarFunction::Mul => return Expression::Literal(Literal::Integer(1)),
                _ => unreachable!("fold associative function only called on Add and Mul"),
            }
        }
        let mut non_literals = Vec::<Expression>::new();
        let (int_fold, long_fold, float_fold) = match sf.function {
            ScalarFunction::Add => {
                sf.args
                    .into_iter()
                    .fold((None, None, None), |(i, l, f), expr| match expr {
                        Expression::Literal(Literal::Integer(val)) => match i {
                            None => (Some(val), l, f),
                            Some(num) => (Some(num + val), l, f),
                        },
                        Expression::Literal(Literal::Long(val)) => match l {
                            None => (i, Some(val), f),
                            Some(num) => (i, Some(num + val), f),
                        },
                        Expression::Literal(Literal::Double(val)) => match f {
                            None => (i, l, Some(val)),
                            Some(num) => (i, l, Some(num + val)),
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
                        Expression::Literal(Literal::Integer(val)) => match i {
                            None => (Some(val), l, f),
                            Some(num) => (Some(num * val), l, f),
                        },
                        Expression::Literal(Literal::Long(val)) => match l {
                            None => (i, Some(val), f),
                            Some(num) => (i, Some(num * val), f),
                        },
                        Expression::Literal(Literal::Double(val)) => match f {
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
            int_fold.map(|val| Expression::Literal(Literal::Integer(val))),
            long_fold.map(|val| Expression::Literal(Literal::Long(val))),
            float_fold.map(|val| Expression::Literal(Literal::Double(val))),
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
            args,
        })
    }

    // Constant folds binary arithmetic functions: subtract and divide
    fn fold_binary_arithmetic_function(
        &mut self,
        function: ScalarFunction,
        left: Expression,
        right: Expression,
    ) -> Expression {
        match function {
            ScalarFunction::Sub => {
                if Self::numeric_eq(&right, 0.0) {
                    left
                } else {
                    match (&left, &right) {
                        (
                            Expression::Literal(Literal::Integer(l)),
                            Expression::Literal(Literal::Integer(r)),
                        ) => Expression::Literal(Literal::Integer(l - r)),
                        (
                            Expression::Literal(Literal::Long(l)),
                            Expression::Literal(Literal::Long(r)),
                        ) => Expression::Literal(Literal::Long(l - r)),
                        (
                            Expression::Literal(Literal::Double(l)),
                            Expression::Literal(Literal::Double(r)),
                        ) => Expression::Literal(Literal::Double(l - r)),
                        _ => Expression::ScalarFunction(ScalarFunctionApplication {
                            function,
                            args: vec![left, right],
                        }),
                    }
                }
            }
            ScalarFunction::Div => {
                if Self::numeric_eq(&right, 0.0) {
                    Expression::Literal(Literal::Null)
                } else if Self::numeric_eq(&right, 1.0) {
                    left
                } else {
                    match (&left, &right) {
                        (
                            Expression::Literal(Literal::Integer(l)),
                            Expression::Literal(Literal::Integer(r)),
                        ) => Expression::Literal(Literal::Integer(l / r)),
                        (
                            Expression::Literal(Literal::Long(l)),
                            Expression::Literal(Literal::Long(r)),
                        ) => Expression::Literal(Literal::Long(l / r)),
                        (
                            Expression::Literal(Literal::Double(l)),
                            Expression::Literal(Literal::Double(r)),
                        ) => Expression::Literal(Literal::Double(l / r)),
                        _ => Expression::ScalarFunction(ScalarFunctionApplication {
                            function,
                            args: vec![left, right],
                        }),
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
    ) -> Expression {
        use std::cmp::Ordering;
        let ord = match (&left, &right) {
            (
                Expression::Literal(Literal::Boolean(l)),
                Expression::Literal(Literal::Boolean(r)),
            ) => l.partial_cmp(r),
            (
                Expression::Literal(Literal::Integer(l)),
                Expression::Literal(Literal::Integer(r)),
            ) => l.partial_cmp(r),
            (Expression::Literal(Literal::Long(l)), Expression::Literal(Literal::Long(r))) => {
                l.partial_cmp(r)
            }
            (Expression::Literal(Literal::Double(l)), Expression::Literal(Literal::Double(r))) => {
                l.partial_cmp(r)
            }
            (Expression::Literal(Literal::String(l)), Expression::Literal(Literal::String(r))) => {
                l.partial_cmp(r)
            }
            _ => None,
        };
        if ord.is_none() {
            return Expression::ScalarFunction(ScalarFunctionApplication {
                function,
                args: vec![left, right],
            });
        }
        let ord = ord.unwrap();
        let val = match function {
            ScalarFunction::Eq => ord == Ordering::Equal,
            ScalarFunction::Gt => ord == Ordering::Greater,
            ScalarFunction::Gte => ord != Ordering::Less,
            ScalarFunction::Lt => ord == Ordering::Less,
            ScalarFunction::Lte => ord != Ordering::Greater,
            ScalarFunction::Neq => ord != Ordering::Equal,
            _ => unreachable!("non-comparison function cannot be called"),
        };
        Expression::Literal(Literal::Boolean(val))
    }

    // Constant folds the between function
    fn fold_between(&mut self, sf: ScalarFunctionApplication) -> Expression {
        assert!(
            sf.args.len() == 3,
            "between scalar function must contain 3 args"
        );
        let (arg, bottom, top) = (sf.args[0].clone(), sf.args[1].clone(), sf.args[2].clone());
        let new_sf = Expression::ScalarFunction(ScalarFunctionApplication {
            function: ScalarFunction::And,
            args: vec![
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Lte,
                    args: vec![arg.clone(), top],
                }),
                Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::Gte,
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
        assert!(
            sf.args.len() == 1,
            "Unary function should only have one arg"
        );
        if Self::has_null_arg(&sf.args) {
            return Expression::Literal(Literal::Null);
        }
        let arg = sf.args[0].clone();
        let func = sf.function;
        let sf_expr = Expression::ScalarFunction(sf);
        if let Expression::Array(ref array) = arg {
            if func == ScalarFunction::Size {
                Expression::Literal(Literal::Integer(array.len() as i32))
            } else {
                sf_expr
            }
        } else if let Expression::Literal(lit) = arg {
            match func {
                ScalarFunction::Pos => match lit {
                    Literal::Integer(_) | Literal::Long(_) | Literal::Double(_) => {
                        Expression::Literal(lit)
                    }
                    _ => sf_expr,
                },
                ScalarFunction::Neg => match lit {
                    Literal::Integer(val) => Expression::Literal(Literal::Integer(-val)),
                    Literal::Long(val) => Expression::Literal(Literal::Long(-val)),
                    Literal::Double(val) => Expression::Literal(Literal::Double(-val)),
                    _ => sf_expr,
                },
                ScalarFunction::Not => {
                    if let Literal::Boolean(val) = lit {
                        Expression::Literal(Literal::Boolean(!val))
                    } else {
                        sf_expr
                    }
                }
                ScalarFunction::Upper => {
                    if let Literal::String(val) = lit {
                        Expression::Literal(Literal::String(val.to_ascii_uppercase()))
                    } else {
                        sf_expr
                    }
                }
                ScalarFunction::Lower => {
                    if let Literal::String(val) = lit {
                        Expression::Literal(Literal::String(val.to_ascii_lowercase()))
                    } else {
                        sf_expr
                    }
                }
                ScalarFunction::CharLength => {
                    if let Literal::String(val) = lit {
                        Expression::Literal(Literal::Integer(val.chars().count() as i32))
                    } else {
                        sf_expr
                    }
                }
                ScalarFunction::OctetLength => {
                    if let Literal::String(val) = lit {
                        Expression::Literal(Literal::Integer(val.bytes().count() as i32))
                    } else {
                        sf_expr
                    }
                }
                ScalarFunction::BitLength => {
                    if let Literal::String(val) = lit {
                        Expression::Literal(Literal::Integer(val.bytes().count() as i32 * 8))
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
    ) -> Expression {
        if let (
            Expression::Literal(Literal::String(sub)),
            Expression::Literal(Literal::String(st)),
        ) = (&substr, &string)
        {
            let pattern = &sub.chars().collect::<Vec<char>>()[..];
            let val = match function {
                ScalarFunction::BTrim => st.trim_matches(pattern).to_string(),
                ScalarFunction::RTrim => st.trim_end_matches(pattern).to_string(),
                ScalarFunction::LTrim => st.trim_start_matches(pattern).to_string(),
                _ => unreachable!("fold trim is only called on trim functions"),
            };
            Expression::Literal(Literal::String(val))
        } else {
            Expression::ScalarFunction(ScalarFunctionApplication {
                function,
                args: vec![substr, string],
            })
        }
    }

    // Constant folds the substring function
    fn fold_substring_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        use std::cmp;
        if Self::has_null_arg(&sf.args) {
            return Expression::Literal(Literal::Null);
        }
        let (string, start, len) = if sf.args.len() == 2 {
            (
                sf.args[0].clone(),
                sf.args[1].clone(),
                Expression::Literal(Literal::Integer(-1)),
            )
        } else if sf.args.len() == 3 {
            (sf.args[0].clone(), sf.args[1].clone(), sf.args[2].clone())
        } else {
            panic!("Substring must have two or three args")
        };
        if let (
            Expression::Literal(Literal::String(st)),
            Expression::Literal(Literal::Integer(start)),
            Expression::Literal(Literal::Integer(len)),
        ) = (string, start, len)
        {
            let string_len = st.len() as i32;
            let end = if len < 0 {
                cmp::max(start, string_len)
            } else {
                start + len
            };
            if start >= string_len || end < 0 {
                Expression::Literal(Literal::String("".to_string()))
            } else {
                let start = cmp::max(start, 0);
                let end = cmp::min(end, string_len);
                let len = end - start;
                let substr = st
                    .chars()
                    .skip(start as usize)
                    .take(len as usize)
                    .collect::<String>();
                Expression::Literal(Literal::String(substr))
            }
        } else {
            Expression::ScalarFunction(sf)
        }
    }

    // Constant folds the concat function
    fn fold_concat_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        if sf.args.is_empty() {
            return Expression::Literal(Literal::String("".to_string()));
        }
        if Self::has_null_arg(&sf.args) {
            return Expression::Literal(Literal::Null);
        }
        let mut result = Vec::<Expression>::new();
        for expr in sf.args {
            match &expr {
                Expression::Literal(Literal::String(val)) => {
                    if result.is_empty() {
                        result.push(expr);
                    } else if let Expression::Literal(Literal::String(prev_val)) =
                        result.last().unwrap().clone()
                    {
                        result.pop();
                        result.push(Expression::Literal(Literal::String(prev_val + val)));
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
                args: result,
            })
        }
    }

    // Constant folds the null if function
    fn fold_null_if_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        assert!(sf.args.len() == 2, "null if should only have two args");
        match (&sf.args[0], &sf.args[1]) {
            (Expression::Literal(l), Expression::Literal(r)) => {
                if l.eq(r) {
                    Expression::Literal(Literal::Null)
                } else {
                    sf.args[0].clone()
                }
            }
            _ => Expression::ScalarFunction(sf),
        }
    }

    // Constant folds the computed field function
    fn fold_computed_field_function(&mut self, left: Expression, right: Expression) -> Expression {
        if let (Expression::Document(doc), Expression::Literal(Literal::String(field))) =
            (&left.clone(), &right)
        {
            doc.get(field)
                .unwrap_or(&Expression::ScalarFunction(ScalarFunctionApplication {
                    function: ScalarFunction::ComputedFieldAccess,
                    args: vec![left, right],
                }))
                .clone()
        } else {
            Expression::ScalarFunction(ScalarFunctionApplication {
                function: ScalarFunction::ComputedFieldAccess,
                args: vec![left, right],
            })
        }
    }

    // Constant folds the coalesce function
    fn fold_coalesce_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        if sf.args.is_empty() {
            return Expression::Literal(Literal::Null);
        }
        let mut is_all_null = true;
        for expr in &sf.args {
            match expr.schema(&EMPTY_STATE) {
                Err(_) => break,
                Ok(sch) => {
                    let sat = sch.satisfies(&Schema::AnyOf(set![
                        Schema::Missing,
                        Schema::Atomic(Atomic::Null),
                    ]));
                    if sat == Satisfaction::Not {
                        return expr.clone();
                    }
                    is_all_null = is_all_null && sat == Satisfaction::Must;
                }
            }
        }
        if is_all_null {
            return Expression::Literal(Literal::Null);
        }
        Expression::ScalarFunction(sf)
    }

    // Constant folds the merge objects function
    fn fold_merge_objects_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        use crate::map;
        use linked_hash_map::LinkedHashMap;
        let mut result_doc: LinkedHashMap<String, Expression> = map! {};
        for expr in &sf.args {
            if let Expression::Document(map) = expr {
                for (key, value) in map {
                    result_doc.insert(key.clone(), value.clone());
                }
            } else {
                return Expression::ScalarFunction(sf);
            }
        }
        Expression::Document(result_doc)
    }

    // Constant folds the slice function
    fn fold_slice_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        use std::cmp;
        if Self::has_null_arg(&sf.args) {
            return Expression::Literal(Literal::Null);
        }
        if sf.args.len() == 2 {
            let (array, len) = (sf.args[0].clone(), sf.args[1].clone());
            if let (Expression::Array(array), Expression::Literal(Literal::Integer(len))) =
                (array, len)
            {
                let array_len = array.len() as i32;
                if len < 0 {
                    let len = cmp::max(0, array_len + len);
                    return Expression::Array(
                        array
                            .into_iter()
                            .skip(len as usize)
                            .collect::<Vec<Expression>>(),
                    );
                } else {
                    let len = cmp::min(len, array_len);
                    return Expression::Array(
                        array
                            .into_iter()
                            .take(len as usize)
                            .collect::<Vec<Expression>>(),
                    );
                }
            }
        } else if sf.args.len() == 3 {
            let (array, start, len) = (sf.args[0].clone(), sf.args[1].clone(), sf.args[2].clone());
            if let (
                Expression::Array(array),
                Expression::Literal(Literal::Integer(start)),
                Expression::Literal(Literal::Integer(len)),
            ) = (array, start, len)
            {
                let array_len = array.len() as i32;
                if len < 0 {
                    return Expression::Literal(Literal::Null);
                }
                if start >= array_len {
                    return Expression::Array(vec![]);
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
                        .collect::<Vec<Expression>>(),
                );
            }
        } else {
            panic!("Slice must have two or three args")
        };
        Expression::ScalarFunction(sf)
    }

    // Constant folds all binary function that evaluate to null if either arg is null
    fn fold_binary_null_checked_function(&mut self, sf: ScalarFunctionApplication) -> Expression {
        assert!(
            sf.args.len() == 2,
            "Binary functions must only have two args"
        );
        if Self::has_null_arg(&sf.args) {
            return Expression::Literal(Literal::Null);
        }
        let (left, right) = (sf.args[0].clone(), sf.args[1].clone());
        match sf.function {
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
        }
    }
}

impl Visitor for ConstantFoldExprVisitor {
    fn visit_expression(&mut self, e: Expression) -> Expression {
        let e = e.walk(self);
        match e {
            Expression::Array(_) => e,
            Expression::Cast(_) => e,
            Expression::Document(_) => e,
            Expression::Exists(_) => e,
            Expression::FieldAccess(_) => e,
            Expression::Is(_) => e,
            Expression::Like(_) => e,
            Expression::Literal(_) => e,
            Expression::Reference(_) => e,
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
                _ => Expression::ScalarFunction(f),
            },
            Expression::SearchedCase(_) => e,
            Expression::SimpleCase(_) => e,
            Expression::SubqueryComparison(_) => e,
            Expression::Subquery(_) => e,
            Expression::TypeAssertion(_) => e,
        }
    }

    fn visit_stage(&mut self, st: Stage) -> Stage {
        let st = st.walk(self);
        match st {
            Stage::Array(_) => st,
            Stage::Collection(_) => st,
            Stage::Filter(_) => st,
            Stage::Group(_) => st,
            Stage::Join(_) => st,
            Stage::Limit(_) => st,
            Stage::Offset(_) => st,
            Stage::Project(_) => st,
            Stage::Set(_) => st,
            Stage::Sort(_) => st,
        }
    }
}

pub fn fold_constants(st: Stage) -> Stage {
    let mut cf = ConstantFoldExprVisitor::default();
    cf.visit_stage(st)
}
