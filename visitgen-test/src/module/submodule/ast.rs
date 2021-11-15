use linked_hash_map::LinkedHashMap;
use mongosql_datastructures::{
    binding_tuple::BindingTuple, unique_linked_hash_map::UniqueLinkedHashMap,
};
use std::collections::{BTreeMap, HashMap};
use std::fmt;

#[derive(Debug, Clone, PartialEq, Eq, PartialOrd, Ord, Hash)]
pub struct Atom {
    pub name: String,
}

impl fmt::Display for Atom {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{:?}", self)
    }
}

#[allow(clippy::large_enum_variant)]
#[derive(Debug, Clone)]
pub enum Expression {
    Atom(Atom),
    Atoms(Vec<Atom>),
    Tree(Tree),
    Plus(Plus),
    Literal(String),
    Null,
}

#[derive(Debug, Clone)]
pub struct Plus {
    pub left: Box<Expression>,
    pub right: Box<Expression>,
}

#[allow(clippy::box_vec, clippy::vec_box)]
#[derive(Debug, Clone)]
pub struct Tree {
    pub branch_b1: Box<String>,
    pub branch_b2: Box<Expression>,
    pub branch_b3: Box<Option<Expression>>,
    pub branch_b4: Box<Vec<Expression>>,
    pub branch_b5: Box<BTreeMap<Atom, Expression>>,
    pub branch_b6: Box<BTreeMap<Box<Atom>, Box<Expression>>>,

    pub branch_o1: Option<String>,
    pub branch_o2: Option<Box<Expression>>,
    pub branch_o3: Option<Vec<Expression>>,
    pub branch_o4: Option<BTreeMap<Atom, Expression>>,
    pub branch_o5: Option<BTreeMap<Box<Atom>, Box<Expression>>>,

    pub branch_v1: Vec<String>,
    pub branch_v2: Vec<Box<Expression>>,
    pub branch_v3: Vec<Vec<Expression>>,
    pub branch_v4: Vec<BTreeMap<Atom, Expression>>,
    pub branch_v5: Vec<BTreeMap<Box<Atom>, Box<Expression>>>,

    pub branch_m1: BTreeMap<Box<Vec<Atom>>, Box<Vec<Expression>>>,
}

#[allow(clippy::box_vec)]
#[derive(Debug, Clone)]
pub struct HashTree {
    pub branch_m1: HashMap<String, String>,
    pub branch_m2: HashMap<Box<Atom>, String>,
    pub branch_m3: HashMap<String, Box<Atom>>,
    pub branch_m4: HashMap<Box<Atom>, Box<Atom>>,

    pub branch_l1: LinkedHashMap<String, String>,
    pub branch_l2: LinkedHashMap<Box<Atom>, String>,
    pub branch_l3: LinkedHashMap<String, Box<Atom>>,
    pub branch_l4: LinkedHashMap<Box<Atom>, Box<Atom>>,

    pub branch_ul1: UniqueLinkedHashMap<String, String>,
    pub branch_ul2: UniqueLinkedHashMap<Box<Atom>, String>,
    pub branch_ul3: UniqueLinkedHashMap<String, Box<Atom>>,
    pub branch_ul4: UniqueLinkedHashMap<Box<Atom>, Box<Atom>>,

    pub branch_bt1: BindingTuple<String>,
    pub branch_bt2: BindingTuple<Box<Atom>>,
}
