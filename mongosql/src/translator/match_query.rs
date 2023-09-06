use crate::{
    air, mir,
    translator::{Error, MqlTranslator, Result},
};

macro_rules! possibly_translate_input {
    ($self:ident, $input:expr) => {
        match $input {
            None => None,
            Some(input) => $self.translate_field_path(input)?,
        }
    };
}

impl MqlTranslator {
    pub fn translate_match_query(
        &self,
        mir_match_query: mir::MatchQuery,
    ) -> Result<air::MatchQuery> {
        use mir::MatchQuery::*;
        match mir_match_query {
            Logical(l) => self.translate_match_logical(l),
            Type(t) => self.translate_match_type(t),
            Regex(r) => self.translate_match_regex(r),
            ElemMatch(em) => self.translate_elem_match(em),
            Comparison(c) => self.translate_match_comparison(c),
        }
    }

    fn translate_match_logical(&self, l: mir::MatchLanguageLogical) -> Result<air::MatchQuery> {
        let args = l
            .args
            .into_iter()
            .map(|arg| self.translate_match_query(arg))
            .collect::<Result<Vec<_>>>()?;

        Ok(match l.op {
            mir::MatchLanguageLogicalOp::Or => air::MatchQuery::Or(args),
            mir::MatchLanguageLogicalOp::And => air::MatchQuery::And(args),
        })
    }

    fn translate_match_type(&self, t: mir::MatchLanguageType) -> Result<air::MatchQuery> {
        let input = possibly_translate_input!(self, t.input);
        Ok(air::MatchQuery::Type(air::MatchLanguageType {
            input,
            target_type: t.target_type.into(),
        }))
    }

    fn translate_match_regex(&self, r: mir::MatchLanguageRegex) -> Result<air::MatchQuery> {
        let input = possibly_translate_input!(self, r.input);
        Ok(air::MatchQuery::Regex(air::MatchLanguageRegex {
            input,
            regex: r.regex,
            options: r.options,
        }))
    }

    fn translate_elem_match(&self, em: mir::ElemMatch) -> Result<air::MatchQuery> {
        let input = self.translate_field_path(em.input)?;
        let input = match input {
            None => return Err(Error::InvalidMatchLanguageInputRef),
            Some(input) => input,
        };

        let condition = self.translate_match_query(*em.condition)?;
        Ok(air::MatchQuery::ElemMatch(air::ElemMatch {
            input,
            condition: Box::new(condition),
        }))
    }

    fn translate_match_comparison(
        &self,
        c: mir::MatchLanguageComparison,
    ) -> Result<air::MatchQuery> {
        use mir::MatchLanguageComparisonOp::*;
        let input = possibly_translate_input!(self, c.input);
        let arg = self.translate_literal_value(c.arg);
        let function = match c.function {
            Lt => air::MatchLanguageComparisonOp::Lt,
            Lte => air::MatchLanguageComparisonOp::Lte,
            Ne => air::MatchLanguageComparisonOp::Ne,
            Eq => air::MatchLanguageComparisonOp::Eq,
            Gt => air::MatchLanguageComparisonOp::Gt,
            Gte => air::MatchLanguageComparisonOp::Gte,
        };
        Ok(air::MatchQuery::Comparison(air::MatchLanguageComparison {
            function,
            input,
            arg,
        }))
    }
}
