use crate::{
    air,
    air::MatchLanguageIn,
    codegen::{MqlCodeGenerator, Result},
};
use bson::{bson, Bson};

/// When a match operator is nested in a $elemMatch, it does not contain
/// a field ref "input". This macro is utilized for codegenning match ops
/// that may have not input.
macro_rules! possibly_nest_under_field {
    ($self:ident, $input:expr, $op:expr) => {
        match $input {
            None => Ok($op),
            Some(fr) => {
                let field = &$self.codegen_field_ref_path_only(fr);
                Ok(bson!({ field: $op }))
            }
        }
    };
}

impl MqlCodeGenerator {
    pub fn codegen_match_query(&self, q: air::MatchQuery) -> Result<Bson> {
        use air::MatchQuery::*;
        match q {
            Or(v) => self.codegen_match_logical_operator("$or", v),
            And(v) => self.codegen_match_logical_operator("$and", v),
            Not(inner) => self.codegen_match_not(*inner),
            Type(t) => self.codegen_match_type(t),
            Regex(r) => self.codegen_match_regex(r),
            ElemMatch(em) => self.codegen_match_elem_match(em),
            Comparison(c) => self.codegen_match_comparison(c),
            // Some versions of MongoDB do not properly optimize {$match: { $expr: false } } so we codegen it
            // this way to ensure efficient performance. Otherwise, such a query could result in a needless collection
            // scan.
            False => Ok(bson!({"_id": Bson::MinKey, "$expr": false})),
            In(c) => self.codegen_match_elem_in(c),
        }
    }

    fn codegen_match_elem_in(&self, args: MatchLanguageIn) -> Result<Bson> {
        let field = self.codegen_field_ref_path_only(args.expression);

        let values: Vec<Bson> = args
            .array_expression
            .into_iter()
            .map(|lit| self.codegen_match_literal_value(lit))
            .collect();

        let op = bson!({ "$in": Bson::Array(values) });

        match args.op {
            air::MatchLanguageInOp::In => Ok(bson!({ field: op })),
            air::MatchLanguageInOp::NotIn => Ok(bson!({
                field: {
                    "$not": op
                }
            })),
        }
    }

    /// Codegens `$not` scoped to a field's operator expression (mirroring how `NotIn` wraps
    /// `$in`), since `$not` — unlike `$and`/`$or`/`$nor` — is not a valid top-level `$match`
    /// boolean operator.
    /// For
    fn codegen_match_not(&self, inner: air::MatchQuery) -> Result<Bson> {
        match inner {
            // $not is not allowed in top level expressions, so we use the $nor operator instead for $and, and $or.
            // Output will look like
            //
            // $nor: [
            //     { $and: [ ... ] }
            // ]
            //
            // We can translate OR directly to $nor because it's the logical negation of OR.
            air::MatchQuery::Or(v) => self.codegen_match_logical_operator("$nor", v),
            air::MatchQuery::And(v) => {
                self.codegen_match_logical_operator("$nor", vec![air::MatchQuery::And(v)])
            }
            _ => {
                let inner_codegen = self.codegen_match_query(inner)?;
                Ok(match inner_codegen {
                    // Field-scoped operator, e.g. {"age": {"$gt": 10}} -> {"age": {"$not": {"$gt": 10}}}
                    Bson::Document(doc) if doc.len() == 1 => {
                        let (key, value) = doc.into_iter().next().expect("checked len == 1");
                        match value {
                            Bson::Document(_) => bson!({ key: { "$not": value } }),
                            // Fieldless operator (e.g. a comparison nested inside $elemMatch, which has
                            // no field ref of its own), e.g. {"$gt": 10} -> {"$not": {"$gt": 10}}
                            _ => bson!({ "$not": { key: value } }),
                        }
                    }
                    // Fieldless operator, e.g. {"$gt": 10} -> {"$not": {"$gt": 10}}
                    bson => bson!({ "$not": bson }),
                })
            }
        }
    }

    fn codegen_match_logical_operator(
        &self,
        op_name: &str,
        args: Vec<air::MatchQuery>,
    ) -> Result<Bson> {
        let args = args
            .into_iter()
            .map(|arg| self.codegen_match_query(arg))
            .collect::<Result<Vec<_>>>()?;
        Ok(bson::bson!({ op_name: Bson::Array(args) }))
    }

    fn codegen_match_type(&self, t: air::MatchLanguageType) -> Result<Bson> {
        let op = match t.target_type {
            air::TypeOrMissing::Missing => bson!({ "$exists": false }),
            air::TypeOrMissing::Number | air::TypeOrMissing::Type(_) => {
                let target_type = t.target_type.to_str();
                bson!({ "$type": target_type })
            }
        };
        possibly_nest_under_field!(self, t.input, op)
    }

    fn codegen_match_regex(&self, r: air::MatchLanguageRegex) -> Result<Bson> {
        let op = bson!({"$regex": Bson::String(r.regex), "$options": Bson::String(r.options)});
        possibly_nest_under_field!(self, r.input, op)
    }

    fn codegen_match_elem_match(&self, em: air::ElemMatch) -> Result<Bson> {
        let field = self.codegen_field_ref_path_only(em.input);
        let condition = self.codegen_match_query(*em.condition)?;

        Ok(bson!({ field: { "$elemMatch": condition } }))
    }

    fn codegen_match_comparison(&self, c: air::MatchLanguageComparison) -> Result<Bson> {
        use air::MatchLanguageComparisonOp::*;
        let arg = self.codegen_match_literal_value(c.arg);
        let comp_op = match c.function {
            Lt => "$lt",
            Lte => "$lte",
            Ne => "$ne",
            Eq => "$eq",
            Gt => "$gt",
            Gte => "$gte",
        };
        let op = bson!({ comp_op: arg });
        possibly_nest_under_field!(self, c.input, op)
    }

    fn codegen_match_literal_value(&self, lit: air::LiteralValue) -> Bson {
        use air::LiteralValue::*;
        match lit {
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
        }
    }
}
