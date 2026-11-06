use quickcheck::*;

#[test]
fn prop_union_simplified_returns_simplified_schema() {
    use crate::schema::*;

    fn union_simplified_schemas(s1: Schema, s2: Schema) -> TestResult {
        let s1_simplified = Schema::simplify(s1);
        let s2_simplified = Schema::simplify(s2);
        let (union_simplified, _) = s1_simplified.union_simplified(s2_simplified);
        TestResult::from_bool(Schema::simplify(union_simplified.clone()) == union_simplified)
    }

    QuickCheck::new()
        .tests(1000)
        .max_tests(10000)
        .rng(Gen::new(0))
        .quickcheck(union_simplified_schemas as fn(Schema, Schema) -> TestResult);
}

#[test]
fn prop_merge_simplified_returns_simplified_schema() {
    use crate::schema::*;

    fn merge_simplified_schemas(s1: Schema, s2: Schema) -> TestResult {
        let s1_simplified = Schema::simplify(s1);
        let s2_simplified = Schema::simplify(s2);
        let (merge_simplified, _) = s1_simplified.merge_simplified(s2_simplified);
        TestResult::from_bool(Schema::simplify(merge_simplified.clone()) == merge_simplified)
    }

    QuickCheck::new()
        .tests(1000)
        .max_tests(10000)
        .rng(Gen::new(0))
        .quickcheck(merge_simplified_schemas as fn(Schema, Schema) -> TestResult);
}

mod arbitrary {
    use crate::schema::*;
    use quickcheck::{Arbitrary, Gen};
    use rand::{RngExt as _, rng};
    use std::collections::{BTreeMap, BTreeSet};

    // Maximum nesting depth of an arbitrary schema. This bounds how deeply nested
    // generated schemas can be. The test starts with a Gen of size 0, and each
    // nested Schema or Document is generated with a Gen whose size is incremented
    // by one via nested_gen. Once the size reaches MAX_NEST, only terminal
    // (non-recursive) Schema variants are generated, which stops further nesting.
    static MAX_NEST: usize = 10;

    /// nested_gen returns a new Gen whose size is incremented by one, used when
    /// generating nested schemas so that the nesting depth can be bounded.
    fn nested_gen(g: &mut Gen) -> Gen {
        Gen::new(g.size() + 1)
    }

    /// arbitrary_schema_terminal generates an arbitrary non-recursive Schema
    /// variant. It is used to stop nesting once MAX_NEST has been reached.
    /// Variants are chosen by weighted probability: Unsat 0.1%, Missing 0.4%,
    /// Any 4.5%, and Atomic 95%.
    fn arbitrary_schema_terminal(g: &mut Gen) -> Schema {
        match rand::rng().random_range(0.0..1.0) {
            p if p < 0.001 => Schema::Unsat,
            p if p < 0.005 => Schema::Missing,
            p if p < 0.05 => Schema::Any,
            _ => Schema::Atomic(Atomic::arbitrary(g)),
        }
    }

    impl Arbitrary for Schema {
        fn arbitrary(g: &mut Gen) -> Self {
            if g.size() >= MAX_NEST {
                return arbitrary_schema_terminal(g);
            }

            let nested_g = &mut nested_gen(g);
            // Choose a variant by weighted probability rather than uniformly:
            // Unsat 0.1%, Missing 0.4%, Any 0.5%, and the remaining 99% split
            // equally (24.75% each) among Atomic, Array, Document, and AnyOf.
            match rand::rng().random_range(0.0..1.0) {
                p if p < 0.001 => Self::Unsat,
                p if p < 0.005 => Self::Missing,
                p if p < 0.01 => Self::Any,
                p if p < 0.2575 => Self::Atomic(Atomic::arbitrary(g)),
                p if p < 0.505 => Self::Array(Box::new(Self::arbitrary(nested_g))),
                p if p < 0.7525 => Self::Document(Document::arbitrary(nested_g)),
                _ => Self::AnyOf(
                    (0..rand::rng().random_range(0..=4u32))
                        .map(|_| Self::arbitrary(nested_g))
                        .collect(),
                ),
            }
        }

        fn shrink(&self) -> Box<dyn Iterator<Item = Self>> {
            match self {
                // Terminal variants cannot be shrunk any further.
                Self::Unsat | Self::Missing | Self::Any => quickcheck::empty_shrinker(),
                // Atomic shrinks toward the simpler Unsat variant.
                Self::Atomic(_) => Box::new(std::iter::once(Self::Unsat)),
                // An Array shrinks to its inner schema and to shrinks of that
                // inner schema wrapped back up in an Array.
                Self::Array(inner) => {
                    let inner = (**inner).clone();
                    Box::new(
                        std::iter::once(inner.clone())
                            .chain(inner.shrink().map(|s| Self::Array(Box::new(s)))),
                    )
                }
                // A Document shrinks to each of its field schemas and to shrinks
                // of the Document itself.
                Self::Document(d) => {
                    let field_schemas = d.keys.values().cloned().collect::<Vec<_>>().into_iter();
                    Box::new(field_schemas.chain(d.shrink().map(Self::Document)))
                }
                // An AnyOf shrinks to each of its branches and to shrinks of the
                // branch set itself.
                Self::AnyOf(branches) => {
                    let branch_schemas = branches.clone().into_iter();
                    Box::new(branch_schemas.chain(branches.shrink().map(Self::AnyOf)))
                }
            }
        }
    }

    impl Arbitrary for Atomic {
        fn arbitrary(g: &mut Gen) -> Self {
            let rng = &(0..Self::VARIANT_COUNT).collect::<Vec<_>>();
            match g.choose(rng).unwrap() {
                0 => Self::MinKey,
                1 => Self::Null,
                2 => Self::Integer,
                3 => Self::Long,
                4 => Self::Double,
                5 => Self::Decimal,
                6 => Self::Symbol,
                7 => Self::String,
                8 => Self::BinData,
                9 => Self::Undefined,
                10 => Self::ObjectId,
                11 => Self::Boolean,
                12 => Self::Date,
                13 => Self::Timestamp,
                14 => Self::Regex,
                15 => Self::DbPointer,
                16 => Self::Javascript,
                17 => Self::JavascriptWithScope,
                18 => Self::MaxKey,
                _ => panic!("missing Atomic variant(s)"),
            }
        }
    }

    impl Arbitrary for Document {
        fn arbitrary(g: &mut Gen) -> Self {
            let keys: BTreeMap<String, Schema> = (0..rng().random_range(0..=5u32))
                .map(|_| {
                    (
                        g.choose(&["a", "b", "c", "d", "e", "f", "g", "h", "i", "j"])
                            .unwrap()
                            .to_string(),
                        Schema::arbitrary(g),
                    )
                })
                .collect();
            let required: BTreeSet<String> = keys
                .keys()
                .filter(|_| bool::arbitrary(g))
                .cloned()
                .collect();
            let additional_properties = bool::arbitrary(g);
            Document {
                keys,
                required,
                additional_properties,
                ..Default::default()
            }
        }
    }
}
