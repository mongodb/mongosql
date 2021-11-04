## Type Constraint Tests
This directory contains type constraint tests for MongoSQL clauses and expressions.
A type constraint test verifies that an expression meets static type requirements.
For example, the arithmetic operators (`+`, `-`, `*`, `/`) statically require that
their operands' types be numeric or `NULL`, and they may be missing.

There are many MongoSQL expressions and clauses that have static type constraints,
and each one has many valid type combinations. To avoid an explosion in the number
of query tests that check static type constraints, we introduced this directory
and test format. See below for details about how the tests in this directory are
structured and how they are meant to be run.

#### Test Structure
A type constraint test is specified as a three field yaml structure:
```text
- description: A description of the test case
  query:       A MongoSQL query that includes the expression or clause that
               has type constraints to test. The type-constrained arguments
               should be named such that the i-th argument is called arg{i},
               for 1 <= i <= num_args, i.e. "arg1", "arg2", and so on.

               All queries should use "foo" as the main datasource.
               If another datasource is required, use "bar". If a query needs
               more than two datasources, the test runner will need to be
               updated such that those datasources are added to the catalog
               schema.
  valid_types: A sequence of maps which each map from operand names to lists
               of valid static types for those operands. A "valid static type"
               is one for which compilation of the query should succeed.

               For a single-argument expression or clause, the sequence is
               straightforward; it contains one map which lists the valid
               static types for the one argument.

               For a multi-argument expression or clause, the sequence may
               contain multiple maps. Each one indicates lists of valid
               static types for each operand. Any combination from the cross
               product of the lists in a map is valid. For expressions with
               correlated operands (i.e. comparison operators), there must be
               multiple maps.

               The type lists are specified as arrays of strings. Each element
               must be either:
                 1. A MongoSQL type name (i.e. "INT", "BOOL", "NULL", etc.)
                 2. "MISSING" to indicate the operand may be missing 

               See the example at the end of this section for reference.
```
\
Here is an example of such a test, followed by a brief explanation:
```text
- description: Addition type constraints
  query: "SELECT arg1 + arg2"
  valid_types:
      - {
          "arg1": ["INT", "LONG", "DOUBLE", "DECIMAL", "NULL", "MISSING"],
          "arg2": ["INT", "LONG", "DOUBLE", "DECIMAL", "NULL", "MISSING"]
        }
```
In this example, the query contains the expression `arg1 + arg2`, an addition operation.
There is one map in the `valid_types` sequence. It indicates that `arg1` can be any
numeric type, `NULL`, or missing; it also indicates that `arg2` can be any numeric type,
`NULL`, or missing. Since these lists are included in the same map, that means any
combination from the cross product of the lists is valid (i.e. `INT, INT`, `LONG, DECIMAL`,
and `DOUBLE,  NULL` are all valid, among others).

#### How to Run
The test runner for these test cases will substitute each argument in the query with a value
of each type. Specifically, it should substitute each argument with an identifier that has a
schema constraining it to a single type. For multiple-argument expressions or clauses, every
combination of types will be substituted. For each substitution, the test runner will  attempt
to compile the query which will either succeed or fail with a static error. The test runner will
check if the combination of substituted types is valid or not (based on the `valid_types` data).
The test will pass if either: the combination is valid and the compilation succeeded, or the
combination is invalid and the compilation failed. It will fail if otherwise.

**Note**: The test runner specifies the constant `MAX_NUM_ARGS`, which represents the maximum number
of arguments substituted in any type constraint test. If a new test is added with more than
`MAX_NUM_ARGS` arguments, the author should update this constant appropriately.

Formally, consider `T` the set of all MongoSQL types. For a query with `n` substitutable arguments,
the set of all combinations of types which will be substituted is S = T<sup>n</sup>, that is the
cross product of `T` with itself `n` times. Each element of `S` is an `n`-tuple of types.

For a `valid_types` sequence of length `m`, each with `n` argument lists:
```text
valid_types:
  - { a_1_1: T_1_1, a2: T_1_2, ..., a_1_n: T_1_n }
  - { a_2_1: T_2_1, a2: T_2_2, ..., a_2_n: T_2_n }
    ...
  - { a_m_1: T_m_1, a2: T_m_2, ..., a_m_n: T_m_n }
```
The set of all valid type combinations `C` is:
```text
C = (T_1_1 X T_1_2 X ... X T_1_n) U (T_2_1 X T_2_2 X ... X T_2_n) U ... U (T_m_1 X T_m_2 X ... X T_m_n)
```
as in, the union of the cross products of the type lists in each map.

Each element of `C` is an `n`-tuple of types and `C` is a subset of `S`. For each element
`s = (t1, ..., tn)` of `S`, the test runner substitutes the arguments of the query `(arg1, ..., argn)`
with an identifier constrianed to the corresponding type via a schema and attempts to compile. If compilation succeeds and
`s` is an element of `C`, the test passes; if compilation fails and `s` is not an element of `C`,
the test passes; otherwise, the test fails.
