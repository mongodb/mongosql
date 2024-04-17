use crate::{
    air,
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
            Type(t) => self.codegen_match_type(t),
            Regex(r) => self.codegen_match_regex(r),
            ElemMatch(em) => self.codegen_match_elem_match(em),
            Comparison(c) => self.codegen_match_comparison(c),
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
