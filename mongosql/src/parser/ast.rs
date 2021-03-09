#[derive(PartialEq, Debug, Clone)]
pub enum Query {
    Select(SelectQuery),
    Set(SetQuery),
}

#[derive(PartialEq, Debug, Clone)]
pub struct SelectQuery {
    pub select_clause: SelectClause,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SetQuery {
    pub left: Box<Query>,
    pub op: SetOperator,
    pub right: Box<Query>,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SetOperator {
    Union,
    UnionAll,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SelectClause {
    pub set_quantifier: SetQuantifier,
    pub body: SelectBody,
}

#[derive(PartialEq, Debug, Clone, Copy)]
pub enum SetQuantifier {
    All,
    Distinct,
}

#[derive(PartialEq, Debug, Clone)]
pub enum SelectBody {
    Standard(Vec<SelectExpression>),
    Values(Vec<SelectValuesExpression>),
}

#[derive(PartialEq, Debug, Clone)]
pub enum SelectValuesExpression {
    Expression(Expression),
    Substar(SubstarExpression),
}

#[derive(PartialEq, Debug, Clone)]
pub enum SelectExpression {
    Star,
    Substar(SubstarExpression),
    Aliased(AliasedExpression),
}

#[derive(PartialEq, Debug, Clone)]
pub struct AliasedExpression {
    pub expression: Expression,
    pub alias: Option<String>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct SubstarExpression {
    pub datasource: String,
}

#[derive(PartialEq, Debug, Clone)]
pub enum Expression {
    Identifier(Identifier),
}

#[derive(PartialEq, Debug, Clone)]
pub enum Identifier {
    Simple(String),
    Compound(Vec<String>),
}
