#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub enum Stage {
    Project(Project),
    Group(Group),
    Limit(Limit),
    Sort(Sort),
    Collection(Collection),
    Join(Join),
    Unwind(Unwind),
    Lookup(Lookup),
    ReplaceRoot(ReplaceRoot),
    Match(Match),
    UnionWith(UnionWith),
    Skip(Skip),
    Documents(Documents),
}

#[derive(PartialEq, Debug, Clone)]
pub struct Project {}

#[derive(PartialEq, Debug, Clone)]
pub struct Group {
    pub source: Box<Stage>,
    pub keys: Vec<NameExprPair>,
    pub aggregations: Vec<AccumulatorExpr>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct NameExprPair {
    pub name: String,
    pub expr: Expression,
}

#[derive(PartialEq, Debug, Clone)]
pub struct AccumulatorExpr {
    pub alias: String,
    pub function: AggregationFunction,
    pub distinct: bool,
    pub arg: Box<Expression>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum AggregationFunction {
    AddToArray,
    Avg,
    Count,
    First,
    Last,
    Max,
    MergeDocuments,
    Min,
    StddevPop,
    StddevSamp,
    Sum,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Limit {
    pub source: Box<Stage>,
    pub limit: u64,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Sort {
    pub source: Box<Stage>,
    pub specs: Vec<SortSpecification>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub enum SortSpecification {
    Asc(String),
    Desc(String),
}

#[derive(PartialEq, Debug, Clone)]
pub struct Collection {
    pub db: String,
    pub collection: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Join {
    pub join_type: JoinType,
    pub left: Box<Stage>,
    pub right: Box<Stage>,
    pub condition: Option<Expression>,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone, Copy)]
pub enum JoinType {
    Left,
    Inner,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Unwind {
    pub source: Box<Stage>,
    pub path: Box<Expression>,
    pub index: Option<String>,
    pub outer: bool,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Lookup {
    pub source: Box<Stage>,
    pub let_vars: Option<Vec<LetVariable>>,
    pub pipeline: Box<Stage>,
    pub as_var: String,
}

#[derive(PartialEq, Debug, Clone)]
pub struct LetVariable {
    pub name: String,
    pub expr: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct ReplaceRoot {}

#[derive(PartialEq, Debug, Clone)]
pub struct Match {
    pub source: Box<Stage>,
    pub expr: Box<Expression>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct UnionWith {
    pub source: Box<Stage>,
    pub pipeline: Box<Stage>,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Skip {
    pub source: Box<Stage>,
    pub skip: u64,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Documents {
    pub array: Vec<Expression>,
    pub alias: String,
}

#[allow(dead_code)]
#[derive(PartialEq, Debug, Clone)]
pub enum Expression {}
