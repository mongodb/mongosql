# MongoSQL Syntax & Semantics

- [MongoSQL Syntax & Semantics](#mongosql-syntax--semantics)
  - [Overview](#overview)
  - [Abstract Model](#abstract-model)
  - [MongoSQL Grammar](#mongosql-grammar)
  - [Type System](#type-system)
    - [Data Types](#data-types)
    - [Type Conversions](#type-conversions)
    - [Schema/Type Constraints](#schematype-constraints)
  - [Clauses](#clauses)
    - [SELECT Clause](#select-clause)
    - [FROM Clause](#from-clause)
      - [Datasources](#datasources)
    - [WHERE Clause](#where-clause)
    - [GROUP BY Clause](#group-by-clause)
    - [HAVING clause](#having-clause)
    - [ORDER BY Clause](#order-by-clause)
    - [LIMIT, FETCH FIRST, and OFFSET Clauses](#limit-fetch-first-and-offset-clauses)
  - [Set Operations](#set-operations)
  - [Expressions](#expressions)
    - [Identifiers](#identifiers)
    - [Literals](#literals)
    - [Parenthesized Expressions](#parenthesized-expressions)
    - [Operators](#operators)
    - [Scalar Functions](#scalar-functions)
      - [Numeric Value Scalar Functions](#numeric-value-scalar-functions)
      - [Addititional Numeric Scalar Functions](#addititional-numeric-scalar-functions)
      - [String Value Scalar Functions](#string-value-scalar-functions)
      - [Datetime Value Scalar Functions](#datetime-value-scalar-functions)
    - [Subquery Expressions](#subquery-expressions)
    - [Document and Field-Access Expressions](#document-and-field-access-expressions)
    - [Array, Indexing, and Slicing Expressions](#array-indexing-and-slicing-expressions)
    - [Null & Missing](#null--missing)
  - [Scoping Rules](#scoping-rules)
  - [Miscellaneous](#miscellaneous)
    - [Document Key Ordering Semantics](#document-key-ordering-semantics)
    - [Not A Number (NaN) Equality](#not-a-number-nan-equality)
    - [MongoSQL Exceptional Behavior](#mongosql-exceptional-behavior)
    - [Comments](#comments)
  - [Future Work](#future-work)
    - [Date and Time Types](#date-and-time-types)
    - [Exposing MQL Functionality](#exposing-mql-functionality)
    - [Collations](#collations)
    - [ORDER BY arbitrary expressions](#order-by-arbitrary-expressions)
    - [Supporting Non-Document BSON Values In Query Results](#supporting-non-document-bson-values-in-query-results)
    - [SELECT DISTINCT](#select-distinct)
    - [UNION DISTINCT](#union-distinct)
    - [USING CLAUSE](#using-clause)
    - [Unify arrays and subqueries](#unify-arrays-and-subqueries)
  - [Appendix](#appendix)
    - [Formal Definitions](#formal-definitions)
    - [Implementation Considerations for Translation to MQL](#implementation-considerations-for-translation-to-mql)

## Overview

This document describes the syntax and semantics of MongoSQL, a SQL
dialect that is designed to provide first-class support for working with
structured and schemaless data in MongoDB, while remaining as compatible
with the SQL-92 standard as possible.

This document is intended to be descriptive, not a formal specification,
though we do provide some formal definitions as part of our behavioral
descriptions where it is helpful to do so. At the moment, the only
formal specification that we provide for MongoSQL is a set of [spec
tests](https://github.com/10gen/mongosql-rs/tree/master/tests/spec_tests).
We expect the language to evolve as needed over the course of the
MongoSQL initiative
[INIT-63](https://jira.mongodb.org/browse/INIT-63)) and
beyond.

## Abstract Model

### Summary

This section provides a model for describing the behavior of a MongoSQL
query, and standardizes definitions and notation that are used
throughout this document for precisely specifying the behavior of
various language features. This model is an "Abstract" model because it
only exists for the purposes of this document, and does not imply or
prescribe any particular implementation decisions.

### Environments

Each MongoSQL (sub-)query and MongoSQL (sub-)expression `q` is evaluated
within a catalog environment ρ<sub>c′</sub>, and values environment
ρ<sub>n′</sub>, where `n` is a subscript used to denote different
versions of `ρ` during evaluation. The values environment, ρ<sub>n′</sub>,
contains values which are in scope at a given point (temporally and syntactically) in a query
execution, `n`. ρ<sub>c</sub> can be thought of as the catalog from most SQL
databases. The initial ρ<sub>n</sub> at the start of execution is noted as
ρ<sub>0</sub>. In general, `n` need not be an integer; the subscript serves
only as a convenient way to distinguish distinct environments during
query execution.

In all cases, an environment is a binding tuple, which is a set
⟨(x<sub>1</sub>,i<sub>1</sub>): v<sub>1</sub>, ..., (x<sub>n</sub>,i<sub>n</sub>): v<sub>n</sub>⟩
where each x<sub>i</sub> is a valid BSON key or `⟘`, each i<sub>i</sub> is a subquery nesting depth, starting
at `0` for the top-level query, and each v<sub>i</sub> is a BSON value. The
shorthand ⟨X:v, ...⟩ stands in for ⟨(X, 0):v, ...⟩.

In a top-level query, ρ<sub>0</sub> is always the empty binding tuple because
there are no local value bindings prior to query execution.

We also define the common set operators over arrays where the results
are ordered in terms of the left argument.

### Basic Query Clause Semantics

In MongoSQL, each clause can be defined in isolation from the others. A
clause is a function that inputs and outputs collections of binding
tuples and the environments defined above. Formally, for _Ρ,_ the set of
all possible environments (binding tuples), and _C_, the set of all
possible arrays/collections of binding tuples, each clause is a
function:

Clause: _Ρ × C → Ρ × C_

That is, each clause is a function accepting a tuple of (ρ<sub>i</sub>, c)
where each `ρ` is a values environment as defined before, and each `c`
is an array/collection of binding tuples. The global catalog
environment, ρ<sub>c</sub>, is always accessible, and never modified, so we
elide it from function inputs and outputs. No clauses actually make
changes to `ρ`, but instead create modified versions of `ρ` for use in
[subqueries](#subquery-expressions) or later clauses.

This is very similar to MQL's stages, which are functions over
collections of BSON documents with one global environment containing the
catalog and the state of certain global variables (e.g. `$$NOW`). The
MongoSQL values environment ρ<sub>i</sub> is analogous to the set of fields
defined by the current document or by let bindings in stages like
\$lookup.

As in standard SQL, the clauses of a MongoSQL query are evaluated in the
following order:

- FROM
- WHERE
- GROUP BY
- HAVING
- SELECT
- ORDER BY
- OFFSET
- LIMIT

Consider the following example:

ρ<sub>c</sub>: ⟨ test: {</br>
&nbsp;&nbsp;foo: [</br>
&nbsp;&nbsp;&nbsp;&nbsp;{'a': 24.5},</br>
&nbsp;&nbsp;&nbsp;&nbsp;{'a': 999}</br>
&nbsp;&nbsp;],</br>
&nbsp;&nbsp;bar: [</br>
&nbsp;&nbsp;&nbsp;&nbsp;{'a':41, 'b':42},</br>
&nbsp;&nbsp;&nbsp;&nbsp;{'a':21, 'c':23}</br>
&nbsp;&nbsp;]</br>
}⟩

ρ<sub>0</sub>: ⟨⟩</br>
SELECT *</br>
FROM test.foo AS x CROSS JOIN test.bar AS y</br>
WHERE x.a > y.a</br>

The query will be executed as follows, producing the specified inputs
and outputs at each step. This example provides a basic introduction to
the semantics of [SELECT](#select-clause),
[FROM](#from-clause), and
[WHERE](#where-clause) clauses. See the specific sections
for a more in depth treatment.

1.

FROM test.foo as x CROSS JOIN test.bar as y</br>
FROM<sub>in</sub> = (ρ<sub>0</sub>, c<sup>FROM</sup><sub>in</sub> = [ ])</br>
FROM<sub>out</sub> = WHERE<sub>in</sub> = (ρ<sub>0</sub>, c<sup>WHERE</sup><sub>in</sub> = [</br>
&nbsp;&nbsp; ⟨(x, 0): {a: 24.5}, (y, 0): {a: 41, b: 42}⟩,</br>
&nbsp;&nbsp; ⟨(x, 0): {a: 24.5}, (y, 0): {a: 21, c: 23}⟩,</br>
&nbsp;&nbsp; ⟨(x, 0): {a: 999}, (y, 0): {a: 41, b: 42}⟩,</br>
&nbsp;&nbsp; ⟨(x, 0): {a: 999}, (y, 0): {a: 21, c: 23}⟩</br>
])</br>

2.

WHERE x.a > y.a</br>
WHERE<sub>out</sub> = SELECT<sub>in</sub> = (ρ<sub>0</sub>, c<sup>SELECT</sup><sub>in</sub> = [</br>
&nbsp;&nbsp;⟨(x, 0): {a: 24.5}, (y, 0): {a: 21, c: 23}⟩,</br>
&nbsp;&nbsp;⟨(x, 0): {a: 999}, (y, 0): {a: 41, b: 42}⟩,</br>
&nbsp;&nbsp;⟨(x, 0): {a: 999}, (y, 0): {a: 21, c: 23}⟩</br>
])</br>

3.

SELECT *</br>
SELECT<sub>out</sub> = (ρ<sub>0</sub>, RESULT = [</br>
&nbsp;&nbsp;⟨(x, 0): {a: 24.5}, (y, 0): {a: 21, c: 23}⟩,</br>
&nbsp;&nbsp;⟨(x, 0): {a: 999}, (y, 0): {a: 41, b: 42}⟩,</br>
&nbsp;&nbsp;⟨(x, 0): {a: 999}, (y, 0): {a: 21, c: 23}⟩</br>
])</br>

First, the FROM clause is evaluated within the environments ρ<sub>c</sub> and
ρ<sub>0</sub> shown at the top of the example. After that, the FROM clause
outputs the shown array of binding tuples, leaving the environments
unchanged. In each binding tuple of FROM<sub>out</sub>,the field names correspond
to the aliases defined in the FROM clause, namely `x` and `y`. Because this
is a cross join, the resulting binding tuples are formed from the cross
product of test.foo and test.bar.

There are no restrictions that the types of values bound to `x` and `y` be
homogeneous, and fields can be missing, as the example shows. The
binding tuples themselves, however, _are_ homogeneous: the same set of
binding-tuple keys must appear in every tuple from the same result set.

Each clause following the FROM clause takes as its input an environment
and an array of binding tuples and outputs an environment and an array
of binding tuples (which then become the input to the next clause.

In the above example, the WHERE clause filters out those binding tuples
where `x.a` is not greater than `y.a` by evaluating its condition for each
binding tuple in the input array, and producing a new output array of
binding tuples containing only those tuples that meet the condition.

Each conditional evaluation in the WHERE clause is performed within the
environments ρ<sub>c</sub> and ρ<sub>i</sub> = _tupleConcat_(ρ<sub>0</sub>,t<sub>j</sub>).
This means that the values from ρ<sub>0</sub> will be consistent while the values in
ρ<sub>1</sub> that come from the current tuple t<sub>j</sub> will vary from tuple to
tuple.

The value environment when the first input tuple is evaluated by the
WHERE clause is:

ρ<sub>WHERE 1</sub> = _tupleConcat_(ρ<sub>0</sub>, d<sub>0</sub>)</br>
&nbsp;&nbsp;= _tupleConcat_(⟨⟩, ⟨(x, 0): {a: 24.5}, y: {a: 41, b: 42}⟩)</br>
&nbsp;&nbsp;= ⟨(x, 0): {a: 24.5}, y: {a: 41, b: 42}⟩</br>

As we can see, ρ<sub>WHERE 1</sub> is equivalent to the first input binding
tuple, because ρ<sub>0</sub> for this query is empty.

The WHERE condition evaluates to false for the first binding tuple in
WHERE<sub>in</sub> because:

(ρ<sub>c</sub>, ρ<sub>WHERE 1</sub>) ⊢ x.a > y.a</br>
&nbsp;&nbsp;→ {a: 24.5}.a > {a: 41, b: 42}.a</br>
&nbsp;&nbsp;→ 24.5 > 41</br>
&nbsp;&nbsp;→ false

The last three binding tuples match the condition, and thus three
binding tuples are sent to the SELECT clause:

(ρ<sub>c</sub>, ρ<sub>WHERE 2</sub> = _tupleConcat_(⟨⟩, ⟨(x, 0): {a: 24.5}, y: {a: 21, c: 23}⟩))</br>
&nbsp;&nbsp;⊢ x.a > y.b</br>
&nbsp;&nbsp;→ {a: 24.5}.a > {a: 21, b: 23}.a</br>
&nbsp;&nbsp;→ 24.5 > 23</br>
&nbsp;&nbsp;→ true

(ρ<sub>c</sub>, ρ<sub>WHERE 3</sub> = _tupleConcat_(⟨⟩, ⟨(x, 0): {a: 999, y: {a: 41, b: 42}⟩))</br>
&nbsp;&nbsp;⊢ x.hello > y.b</br>
&nbsp;&nbsp;→ {a: 999}.a > {a: 41, b: 42}.a</br>
&nbsp;&nbsp;→ 999 > 42</br>
&nbsp;&nbsp;→ true</br>

(ρ<sub>c</sub>, ρ<sub>WHERE 4</sub> _= tupleConcat_(⟨⟩, ⟨(x, 0): {a: 999, y: {a: 21, c: 23}⟩))</br>
&nbsp;&nbsp;⊢ x.hello > y.b</br>
&nbsp;&nbsp;→ {a: 999}.a > {a: 21, b: 23}.a</br>
&nbsp;&nbsp;→ 999 > 23</br>
&nbsp;&nbsp;→ true

This pattern of "input an array (stream) of binding tuples, evaluate
the constituent expressions, output an array (stream) of binding
tuples" has a couple exceptions: LIMIT and OFFSET do not need to
actually evaluate any expressions for each binding tuple. Additionally,
the FROM clause takes no input. For example, a LIMIT 10 clause that
inputs an array or stream with 100 binding tuples need not access
binding tuples 11-100.

Overall, each clause of MongoSQL is an operator that inputs/outputs
arrays (or streams) of binding tuples. As such, we can define the
semantics of each clause separately from the semantics of the other
clauses.

#### Materializing Binding Tuples as Documents

MongoDB's wire protocol allows only for result sets of documents, not
binding tuples. Therefore, we will need a way of transforming binding
tuples in our result sets into BSON documents so that they can be
returned in MongoDB result sets.

There are many possible algorithms for rendering a binding tuple to a
BSON document, and MongoSQL does not require any particular one as part
of the specification. The choice of rendering algorithm(s) is left to
the database hosting the MongoSQL implementation. For the sake of
consistency, we describe a reasonable default algorithm below:

- Check each binding in the binding tuple

  - If keys exist with nesting depths other than 0, it is an error[<sup>1</sup>](#1).

  - If the key's datasource name is ⟂:

    - The value for that key must be a document, and all the keys
      of that document will be materialized in the root of the
      target document

  - If the key's datasource name is a valid BSON key name:

    - If the value is a document:

      - If all keys of that document do not conflict with any

        keys of any other subdocuments or the root, we
        materialize that document's bindings in the root of
        the target document

      - If there are key conflicts, we materialize the document
        as a nested subdocument of the root materialized
        document with the binding tuple key as a BSON key

    - If the value is not a document, materialize as a BSON key,
      value pair using the binding tuple key

Examples:

⟨(a, 0): 1, (b, 0): 2⟩

materializes as

{a: 1, b: 2}

⟨(⟂, 0): {a: 1}, (b, 0): {a: 1, b: 2}⟩

materializes as

{a: 1, b: {a: 1, b: 2}}

Because b.a would conflict with a if rerooted.

⟨(⟂, 0): {a: 1}, (b, 0): {c: 1, b: 2}⟩

materializes as

{a: 1, c: 1, b: 2}

Because there are no conflicts introduced by rerooting b.

If the keys of documents in the generated binding tuples cannot be
statically enumerated at query planning time, we assume that conflicts
exist, meaning that the result will be materialized as a nested
subdocument under the binding tuple key.

> <sup id="1">1</sup> This is an error because the only place where we should ever be transforming
  a MongoSQL binding-tuple result set to a BSON-document result set is at the root
  of a query. This condition should never arise unless there is an implementation error.

## MongoSQL Grammar

This is a linked grammar of the entire MongoSQL syntax. Each production
links to the section of the document that further elucidates the syntax
and semantics of the given construct.

\<MongoSQL statement\> ::= \<select query\>

\<select query\> ::= [\<set query\>](#set-operations)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| [\<select clause\>](#select-grammar) [\<from clause\>](#from-grammar)?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[\<where clause\>](#where-grammar)? [\<group by clause\>](#group-by-clause)?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[\<having clause\>](#having-clause)? [\<order by clause\>](#order-by-grammar)?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\<limit/offset clause\>?

\<limit/offset clause\> ::= [\<limit clause\>](#limit-fetch-first-and-offset-clauses) [\<offset clause\>](#limit-fetch-first-and-offset-clauses)?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| [\<offset clause\>](#limit-fetch-first-and-offset-clauses) [\<limit clause\>](#limit-fetch-first-and-offset-clauses)?

## Type System

### Data Types

#### Behavioral Description (Data Types)

MongoSQL types are the set of [BSON types](http://bsonspec.org/spec.html). All of these types
can be queried in MongoSQL, though not all of them have literal syntax
(deprecated BSON types, for example). See below for the list of all BSON
types:

- Double
- String
- Document
- Array
- BinData
- ObjectId
- Boolean
- Date
- Null
- Regex
- DBPointer
- Javascript
- Symbol
- JavascriptWithScope
- Int
- Timestamp
- Long
- Decimal
- MinKey
- MaxKey

#### Type Names and Aliases

Each type in MongoSQL has a name, which is a keyword that can be used to
reference the type when necessary (e.g. in an expression like CAST).
Some types have one or more aliases in addition to their primary name.

Generally, MongoSQL types use their BSON type names and, for SQL-92
compatibility, their corresponding SQL-92 type names as aliases. For
example, the MongoSQL Integer type uses the name INT from BSON and the
aliases INTEGER and SMALLINT from SQL-92.

There are two exceptions to this: the BSON Date type and the BSON
Timestamp type. The BSON Date type name conflicts with the SQL-92 Date
type name. A BSON Date is a datetime, whereas a SQL-92 Date is just a
date. The BSON Timestamp type name conflicts with the SQL-92 Timestamp
type name. A BSON Timestamp is a special type for MongoDB, whereas a
SQL-92 Timestamp is a datetime. Since MongoSQL must be SQL-92 compliant,
SQL-92 names take precedence. Therefore, to address the naming
conflicts, the MongoSQL Datetime type, which corresponds to the BSON
Date type, uses the SQL-92 Timestamp type name, TIMESTAMP, and the
MongoSQL Timestamp type, which corresponds to the BSON Timestamp type,
uses the name BSON_TIMESTAMP. The MongoSQL Datetime type also uses the
alias BSON_DATE. See the [Grammar](#type-alias-grammar) for all type
names and aliases.

Type aliases are rewritten to their core names in MongoSQL queries. See
the [Examples](#type-alias-examples) for details.

<div id="type-alias-grammar"/>

#### Grammar

\<type\> ::= \<double type\> \| \<string type\> \| \<document type\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<array type\> \| \<binary type\> \| \< undefined type\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<objectId type\> \| \<boolean type\> \| \<date type\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<null type\> \| \<regular expression type\> \| \<dbPointer type\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<javascript type\> \| \<symbol type\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<javascriptWithScope type\> \| \<32-bit integer type\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<timestamp type\> \| \<64-bit integer type\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<decimal128 type\> \| \<min key type\> \| \<max key type\>

\<double type\> ::= DOUBLE PRECISION?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| REAL</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| FLOAT (\"(\" [\<integer literal\>](#literals) \")\")?</br>

\<string type\> ::= STRING</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| VARCHAR (\"(\" [\<integer literal\>](#literals) \")\")?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| CHAR (\"(\" [\<integer literal\>](#literals) \")\")?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| CHARACTER (\"(\" [\<integer literal\>](#literals) \")\")?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| CHAR VARYING ( \"(\" [\<integer literal\>](#literals) \")\")?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| CHARACTER VARYING ( \"(\" [\<integer literal\>](#literals) \")\")?

\<document type\> ::= DOCUMENT</br>
\<array type\> ::= ARRAY</br>
\<binary data type\> ::= BINDATA</br>
\<undefined type\> ::= UNDEFINED</br>
\<objectId type\> ::= OBJECTID</br>
\<boolean type\> ::= BOOL \| BIT \| BOOLEAN</br>
\<datetime type\> ::= BSON_DATE \| TIMESTAMP</br>
\<null type\> ::= NULL</br>
\<regular expression type\> ::= REGEX</br>
\<dbPointer type\> ::= DBPOINTER</br>
\<javascript type\> ::= JAVASCRIPT</br>
\<symbol type\> ::= SYMBOL</br>
\<javascriptWithScope type\> ::= JAVASCRIPTWITHSCOPE</br>
\<32-bit integer type\> ::= INT \| INTEGER \| SMALLINT</br>
\<timestamp type\> ::= BSON_TIMESTAMP</br>
\<64-bit integer type\> ::= LONG</br>
\<decimal128 type\> ::= DECIMAL (\"(\" [\<integer literal\>](#literals) (\",\" [\<integer literal\>](#literals))? ")")?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| DEC (\"(\" [\<integer literal\>](#literals) (\",\" [\<integer literal\>](#literals))? ")")?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| NUMERIC (\"(\" [\<integer literal\>](#literals) (\",\" [\<integer literal\>](#literals))? ")")?</br>
\<min key type\> ::= MINKEY</br>
\<max key type\> ::= MAXKEY

<div id="type-alias-examples"/>

#### Examples

See [spec tests](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/rewrite_tests/type_aliases.yml).

#### Rejected Alternatives/Competitive Analysis

The INTERVAL type is not supported by MongoSQL. There are different
levels of SQL-92 support and INTERVAL is not part of SQL-92 ENTRY
support, so we chose not to include it. Additionally, there is no
corresponding BSON type for INTERVAL.

Given the name conflict between the BSON Timestamp and SQL-92 Timestamp
types, and the BSON Date and SQL-92 Date types, we chose to prefix the
MongoSQL names for the BSON types with "BSON". For consistency
across all BSON types, we considered supporting aliases for all BSON
types that used the "BSON\_" prefix. We ultimately decided against
this because (1) it avoids clutter, and (2) consistency across alias
names for all BSON types may actually lead to more user confusion as
opposed to reducing it.

### Type Conversions

#### Behavioral Description (Type Conversions)

Type conversions fall into two categories: explicit and implicit.
Explicit type conversions are directly invoked by users. Implicit type
conversions are performed without users specifically requesting them. In
MongoSQL, all type conversions are explicit.

Explicit type conversions are expressed via the CAST [scalar function](#scalar-functions)
or the `::` [operator](#operators). The CAST function accepts up to
four arguments. As defined by SQL-92, the base-level invocation of CAST
accepts an operand to cast and a target data type. In addition to that
form, MongoSQL accepts two additional, optional arguments. One
additional argument is an expression to return if the operand is NULL or
MISSING; if omitted, casting NULL or MISSING to anything results in
NULL. The other additional argument is an expression to return if the
conversion produces a runtime error; if omitted, MongoSQL will return
NULL if it encounters a casting error. See the
[Grammar](#type-conversion-grammar) for details on how to provide these
extra arguments.

The `::` operator is a shorthand alias for the two-argument form of the
CAST function.

\<expr\>`::`\<type\></br>
will be rewritten as:</br>
CAST(\<expr\> AS \<type\>)

In addition to CAST and `::`, implementations may also define
"constructor" scalar functions that alias CAST invocations to certain
types. For example, an implementation may define an
OBJECTID(\<expression\>) function as shorthand for CAST(\<expression\>
AS OBJECTID).

Although all type conversions must be explicit, that does not mean type
conversions are required in all circumstances. Numeric types are all
mutually comparable. Therefore, MongoSQL allows operations between the
various numeric types without casting the operands to be the same
numeric type. For example, an int can be added to a double, or a long
could be compared to a decimal, etc.

#### Conversion Behavior

The semantics of type conversions in MongoSQL are similar to those of
the aggregation pipeline operator
[\$convert](https://docs.mongodb.com/manual/reference/operator/aggregation/convert/#definition).
The range of target types is a subset of all MongoSQL types. Valid
target types are ARRAY, DOCUMENT, DOUBLE, STRING, OBJECTID, BOOL,
BSON_DATE, INT, LONG, and DECIMAL, or any of their corresponding SQL-92
type aliases as shown in the [Data Types](#data-types)
section. If the provided target type is invalid, a static error is
returned.

For all target types except ARRAY and DOCUMENT, MongoSQL type conversion
behaves the same as
[\$convert](https://docs.mongodb.com/manual/reference/operator/aggregation/convert/#definition).
The exception to this is that MongoSQL CAST returns NULL instead of
throwing a runtime error if a conversion error is encountered and no ON
ERROR argument is provided. Attempting to CAST to a target type from an
incompatible source type, for example BOOL to BSON_DATE, is considered a
conversion error and evaluates to ON ERROR or NULL if that is not
provided.

Casting to ARRAY behaves as follows:

| **Input** | **Behavior** |
|-----------|--------------|
| ARRAY     | No-op |
| NULL</br>MISSING | ON NULL expression, if provided</br> NULL otherwise |
| Any other type | ON ERROR expression, if provided </br> NULL otherwise |

Casting to DOCUMENT behaves as follows:

| **Input** | **Behavior** |
|-----------|--------------|
| DOCUMENT | No-op |
| NULL </br> MISSING | ON NULL expression, if provided </br> NULL otherwise |
| Any other type | ON ERROR expression, if provided </br> NULL otherwise |

<div id="type-conversion-grammar" />

#### Grammar

\<cast expression\> ::= [\<expression\>](#expressions) \"::\" [\<type\>](#data-types)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| CAST \"(\" [\<expression\>](#expressions) AS [\<type\>](#data-types)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(\",\" [\<expression\>](#expressions) ON NULL)?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(\",\" [\<expression\>](#expressions) ON ERROR)?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\")\"

<div id="type-conversion-examples" />

#### Examples

See query spec tests
[here](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/type_conversions.yml)
and rewrite spec tests
[here](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/rewrite_tests/type_conversions.yml).

<div id="type-conversion-rejected" />

#### Rejected Alternatives/Competitive Analysis

SQL-92 specifies that implicit casting is only required within groups of
similar types, i.e. numeric types, character types, bit types, datetime
types, and interval types, not between them ([SQL-92 4.6](https://www.contrib.andrew.cmu.edu/~shadow/sql/sql1992.txt)).

MongoSQL does not have implicit type conversions between different
groups of types. For example, \<int\> + \<string\> requires that the
\<string\> be explicitly cast to a numeric type. MongoSQL does support
some operations between different numeric types, but this is transparent
to users. Implicit conversions lead to confusing semantics and
potentially unexpected and surprising behavior. Competitors do not
support implicit type conversions:

- [SQL++](https://asterixdb.apache.org/docs/0.9.1/sqlpp/manual.html#Type_errors) clearly specifies it does not.
- N1QL specifies it does not by each \"Functions\"/\"Operators\" docs page.
- Rocket specifies it does not by each \"Functions\" docs page.
- PartiQL does not specify in the docs, but we observed that it does not.
- Presto does not specify in the docs, but we observed that it does not.

Popular SQL dialects MySQL and SQLite do seem to have implicit casting
such that \<int\> + \<string\> works without an explicit cast.

### Schema/Type Constraints

<div id="schematype-behavioral-description" />

#### Behavioral Description

A _schema_ in MongoSQL is a collection of facts about an expression or
collection that are known to be true at compile time.

MongoSQL schemas are similar to a [structural type
system](https://en.wikipedia.org/wiki/Structural_type_system),
though we use the word "schema" instead of "type" to avoid confusion
with MongoSQL data types. For example, a MongoSQL schema might tell us
that "this expression is either a boolean or a document with subfields a
and b", or "this expression is either an array of length 1 or a positive
integer".

#### Schema Inference

MongoSQL gets schema information from a number of different places:

- The database that is hosting a MongoSQL implementation may provide schema information
  about some or all collections.
- The types of all other MongoSQL expressions (literals, operators, scalar functions, etc)
  can be determined at compile time. For example, we can know at compile time that \"abc\"
  is a string, and that CASE WHEN x THEN 5 ELSE false END is either an integer or a boolean.

If schema information is not available from the host database, MongoSQL
still assigns static types to all expressions (including column
references); these types will just be much less constrained than they
would be if schema data were available. For example, when no schema data
is available for the collection foo, the field a in the query SELECT a
FROM foo could be a value of any MongoSQL type, or could be missing.

#### Type Checking

There are numerous situations where an expression's type needs to be
verified at compile time. Every function and operator has constraints on
the types of its arguments, and some clauses also have type constraints
for expressions in certain positions. Some expressions that would
typically check types at runtime (x IS INTEGER, for example) may also be
evaluated at compile time if sufficient type information is available.

If a static type constraint is not satisfied, then the query will fail
to compile.

#### Type assertion

In order to match the type constraint mentioned above, users will need
to add CAST on expressions to declare its type in the schema-less mode.
To simplify this process and reduce the unnecessary conversion stage
created during translation, we introduced the type assertion operator
`::!`, with which the expression will be statically treated as the type
appended to the operator. During evaluation, there won't be any
conversion applied on the result of this expression.

When the type assertion operator is applied on the expression with
determined types, a static check will be performed to make sure that the
target type is among those types. If not, a static error will be thrown.
For example, given schema

\'foo\': {</br>
&nbsp;&nbsp;\'bsonType\': \"int\"</br>
},

foo::!STRING will fail the static check because STRING is not part of
the possible types foo contains.

Given the schema:

\'foo\': {</br>
&nbsp;&nbsp;\'anyOf\': \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;{ \'bsonType\': \"int\" },</br>
&nbsp;&nbsp;&nbsp;&nbsp;{ \'bsonType\': \"string\" }</br>
&nbsp;&nbsp;\]</br>
}

foo::!STRING will pass the static check and be treated as string type in
the query.

Be aware if the column reference contains a type that doesn\'t work
under the expression, even if it can pass the static check, a runtime
error may be returned when the query is executed. For example,
substr(foo::!STRING, 1, 2) will throw a runtime error if foo is not
actually a STRING type because substr only accepts a STRING or NULL
value.

<div id="schematype-grammar" />

#### Grammar

\<type assertion\> ::= [\<expression\>](#expressions)::\![\<type\>](#type-alias-grammar)

There is no additional syntax associated with schema/type constraints. A
MongoSQL implementation may want to provide an API for getting type
information about things such as collections, result sets, or
expressions, but it is intentional that this document specifies neither
syntax for accessing that information in a query nor a standard textual
representation of a MongoSQL type.

<div id="schematype-examples" />

#### Examples

Since there is no new syntax associated with this section, most of the
type-related testing will reside with the spec tests for various clauses
and expressions. When it is relevant to do so, spec tests for clauses
and expressions include:

- Tests validating expression return types
- Tests validating type constraints for arguments
- Tests validating that a specific type inference can be made

<div id="schematype-rejected" />

#### Rejected Alternatives/Competitive Analysis

- Schemas in PartiQL

  - Though PartiQL is the SQL dialect most similar to MongoSQL, the
    section of its spec that discusses types is marked as "WIP",
    and has no useful information. For all intents and purposes,
    the behavior of types/schemas in PartiQL is unspecified.

- Type metadata in relational databases

  - Most SQL databases provide type information in a queryable
    format via the INFORMATION_SCHEMA database or DESCRIBE
    command.

  - There are a few reasons why we choose not to provide this
    information via the MongoSQL query language:

    - We expect that MongoSQL will be "hosted" by a
      MongoDB-compatible database server. Unlike many other SQL
      databases, the SQL interface is not the only way of
      interacting with the database. Metadata can be made
      available via a separate command

    - The data model of MongoSQL is fundamentally different from
      that of standard SQL, and so we would not be able to
      expose MongoSQL's richer type information in a
      standards-compliant manner

    - Consumers of MongoSQL who need a standards-compatible way of
      accessing metadata (notably BI Tools) will be able to do
      so via ODBC or JDBC, as opposed to by writing DESCRIBE or
      INFORMATION_SCHEMA queries

## Clauses

### SELECT Clause

<div id="select-behavior" />

#### Behavioral Description

The main form of the \<select clause\> is SELECT VALUES. Other formats
of \<select clause\> are syntactic sugar that can be syntactically
rewritten to an equivalent SELECT VALUES query.

#### VALUE vs VALUES

MongoSQL allows SELECT VALUE and SELECT VALUES to be used
interchangeably. Both are supported to allow for a more natural-reading
query in the presence of single or multiple expressions. MongoSQL will
rewrite the keyword to match the number of arguments, as shown in the
following examples:

SELECT VALUE a.\*, b.\* FROM a JOIN b</br>
will be rewritten to</br>
SELECT VALUES a.\*, b.\* FROM a JOIN b

SELECT VALUES {'a': 1} FROM foo</br>
will be rewritten to</br>
SELECT VALUE {'a': 1} FROM foo

#### SELECT VALUE Semantics

SELECT VALUE(S) accepts a list of \<select value expr\>s, which can be
one of two things: an ordinary expression that resolves to a document or
a sub-star expression. If an expression provided to SELECT VALUE is not
a sub-star expression and cannot be statically determined to be a
document, then a static error will be returned.

SELECT VALUE constructs one output binding tuple for each tuple in the
SELECT<sub>in</sub> stream. Each output binding tuple has one binding per
\<select value expr\>. If the \<select value expr\> is a document
expression, then the binding's key is ⟘ and its value is the result of
evaluating the document in the local values environment _ρ_. If the
\<select value expr\> is a sub-star expression, then the binding's key
is the identifier preceding the star, and the value is the root value
from the datasource referenced by that identifier.

Consider the following query that demonstrates how document expressions
are handled:

SELECT VALUE {a: y.a, b: y.c} FROM \[{a: 1, c: 2}, {a: 3}\] AS y

Since this SELECT query is a top-level query (not a correlated
subquery), the initial values environment is empty (i.e. ρ<sub>0</sub> = ⟨⟩).
For each input tuple, we create the local values environment by
concatenating ρ<sub>0</sub> with the tuple. In the example above, the document
expression will be evaluated twice; first in the environment:

ρ<sub>row1</sub> = _tupleConcat(ρ<sub>0</sub>, ⟨y: {a: 1, c: 2}⟩)_</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;= _tupleConcat(⟨⟩, ⟨y: {a: 1, c: 2}⟩)_</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;= ⟨y: {a: 1, c: 2}⟩</br>

And then (for the second input tuple) in the environment

ρ<sub>row2</sub> = _tupleConcat(ρ<sub>0</sub>, ⟨y: {a: 3}⟩)_</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;= _tupleConcat(⟨⟩, ⟨y: {a: 3}⟩)_</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;= ⟨y: {a: 3}⟩

We create one binding tuple per input binding tuple, each mapping the ⟘
key to the value of the document expression. For the example query
above, the output stream is:

SELECT<sub>out</sub> = \[⟨⟘ : ρ<sub>row1</sub> ⊢ \<document expression\>⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨⟘ : ρ<sub>row2</sub> ⊢ \<document expression\>⟩\]

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;= \[⟨⟘ : ρ<sub>row1</sub> ⊢ {a: y.a, b: y.c}⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨⟘ : ρ<sub>row2</sub> ⊢ {a: y.a, b: y.c}⟩\]

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;= \[⟨⟘ : ⟨y: {a: 1, c: 2}⟩ ⊢ {a: y.a, b: y.c}⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨⟘ : ⟨y: {a: 3}⟩ ⊢ {a: y.a, b: y.c}⟩\]

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;= \[⟨⟘ : {a: 1, b: 2}⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨⟘ : {a: 3}⟩\]

Next, consider a query that demonstrates how sub-star expressions are
handled:

SELECT VALUE y.\* FROM \[{a: 1, c: 2}, {a: 3}\] AS y

As in the previous example, the sub-star expression will be evaluated
twice; first in the environment:

ρ<sub>row1</sub> = ⟨y: {a: 1, c: 2}⟩

And then (for the second input tuple) in the environment

ρ<sub>row2</sub> = ⟨y: {a: 3}⟩

Which gives us the output stream

SELECT<sub>out</sub> = \[⟨y : _getReference_(ρ<sub>row1</sub>, y)⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨y : _getReference_(ρ<sub>row2</sub>, y)⟩\]

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;= \[⟨y : _getReference_(⟨y: {a: 1, c: 2}⟩, y)⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨y : _getReference_(⟨y: {a: 3}⟩, y)⟩\]

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;= \[⟨y : {a: 1, c: 2}⟩⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨y : {a: 3}⟩\]

##### Rewriting SELECT ... with no FROM clause

SELECT ...

will be rewritten as:

SELECT ... FROM \[{}\] AS \_dual

##### Rewriting SELECT x.y.a, b, ... where x.y.a and b are field references

SELECT x.y.a, b, ...

will be rewritten as:

SELECT x.y.a AS a, b AS b, ...

<div id="rewriting-select-expr-expr"/>

##### Rewriting SELECT \<expression1\>, \<expression2\>, ...

SELECT \<expression1\>, \<expression2\>, ...

will be rewritten as:

SELECT \<expression1\> AS \_1, \<expression2\> AS \_2

Where \_1 and \_2 are names generated by the implementation
corresponding to position in the resulting documents, starting at 1
because SQL traditionally numbers the first column as 1.

<div id="rewriting-select-expr-sub-expr-sub" />

##### Rewriting SELECT \<expression<sub>1</sub>\> AS x<sub>1</sub>, \<expression<sub>2</sub>\> as x<sub>2</sub>, ..., t.\*, ...

SELECT \<expression<sub>1</sub>\> AS x<sub>1</sub>, \<expression2\> as x<sub>2</sub>, ..., t.\*,
...

will be rewritten as:

SELECT VALUES {x<sub>1</sub>: \<expression<sub>1</sub>\>, x<sub>2</sub>: \<expression<sub>2</sub>\>}, ...,
t.\*, ...

#### Semantics of SELECT \*

SELECT \* will return all binding tuples from the input stream
unmodified. Other expressions may not be present in the select list
alongside \*.

<div id="select-grammar" />

#### Grammar

\<select clause\> ::= SELECT \<set quantifier\>? \<value keyword\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\<select value list\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| SELECT \<set quantifier\>? \<select list\>

\<value keyword\> ::= VALUE \| VALUES

\<set quantifier\> ::= ALL

\<select value list\> ::= \<select value expr\> (, \<select value expr\>)\*

\<select value expr\> ::= [\<expression\>](#expressions) \| \<substar expr\>

\<select list\> ::= \<select expr\> (, \<select expr\>)\*

\<select expr\> ::= \"\*\"</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<substar expr\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<aliased expr\>

\<alias\> ::= AS? [\<identifier\>](#identifiers)

\<aliased expr\> ::= [\<expression\>](#expressions) \<alias\>?

\<substar expr\> ::= [\<identifier\>](#identifiers) \".\" \"\*\"

<div id="select-examples" />

#### Examples

[Rewrite Examples](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/rewrite_tests/select.yml)

[Query Examples](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/select.yml)

<div id="select-rejected" />

#### Rejected Alternatives/Competitive Analysis

- We chose not to support a DUAL table, which deviates from existing
  BI Connector and SQL in ADL behavior.

  - The DUAL table was originally added in Oracle, and several other
    database systems support it, e.g., MySQL.

  - However, many database systems, e.g. SQLite and Postgres, do not
    have a DUAL table, see
    [https://en.wikipedia.org/wiki/DUAL_table](https://en.wikipedia.org/wiki/DUAL_table)
    for more information on which database systems have DUAL
    tables and which just allow SELECT without a table to achieve
    the same effect.

### FROM Clause

<div id="from-behavior" />

#### Behavioral Description

FROM is the first clause of every MongoSQL query to be evaluated. This
makes it a special case, because it does not take a binding-tuple stream
as its input. Instead, it generates its output tuple stream from various
datasources.

MongoSQL has four categories of datasources:

- Simple datasources
  - [Collection](#collection-datasource)
  - [Array](#array-datasource)
  - [Derived Table](#derived-table-datasource)
- Compound (Join) datasources
  - [Cross Join](#comma-cross-join-datasource)
  - [Inner Join](#inner-join-datasource)
  - [Left Outer Join](#left-outer-join-datasource)
  - [Right Outer Join](#right-outer-join-datasource)
- [Unwind datasources](#unwind-datasource)
- [Flatten datasources](#flatten-datasource)

Datasources provide various ways of creating streams of binding tuples.
Simple datasources create streams of tuples with a single key, while
compound datasources create streams of tuples with multiple keys.

The top-level datasource in a FROM clause forms the clause's output
stream.

#### Datasources

##### Simple Datasources

###### Collection Datasource

A collection datasource is composed of a collection reference (qualified
or unqualified) and an optional alias. Formally, the collection
reference is resolved in the catalog environment using the
_getReference_ function defined in the [Abstract
Model](#environments) section. Informally, qualified
references are treated as \<db\>.\<collection\> pairs, while unqualified
references are treated as collections in the current database.

Collection datasources without an explicit alias are syntactically
rewritten to have an alias. For unqualified identifiers, the whole
identifier is used as the alias. For qualified identifiers, the
collection part is used as the alias.

A collection datasource creates a stream of binding tuples with a single
key-value pair. One binding tuple is created per document in the
referenced collection. The key of each tuple is the alias name, and the
value is the root document.

For example, consider the output of the collection datasource in the
following query:

SELECT \* FROM collection AS alias

SELECT<sub>out</sub> = FROM<sub>out</sub> = \[⟨alias: d⟩ for d ∊ collection\]

###### Array Datasource

An array datasource is composed of an array literal and an alias. The
array's elements must statically evaluate to document values;
syntactically, expressions are permitted inside the array, as long as
they can be evaluated at compile time.

An array datasource creates a stream of binding tuples with a single
key-value pair. One binding tuple is created per value in the array. The
key of each tuple is the alias name, and the value is the array element.

For example, consider the output of the array datasource in the
following query:

SELECT \* FROM \[{'a': 1}, {'a': 2}\] AS alias

SELECT<sub>out</sub> = FROM<sub>out</sub> = \[⟨alias: {'a': 1}⟩, ⟨alias: {'a': 2}⟩\]

###### Derived Table Datasource

A derived table datasource is made up of a parenthesized MongoSQL query
and an alias. Note, unlike a [subquery
expression](#subquery-expressions), a derived table
datasource cannot have correlated fields from outer queries.

A derived table datasource creates a stream of binding tuples as defined
by the semantics for the SELECT clause of the derived table datasource
query. One new binding tuple is created per binding tuple returned by
the SELECT clause of the derived table query. The key of each tuple is
the alias name, and the value is the result of merging the values of
_all_ bindings in the corresponding input tuple. For example, consider
the output of the derived table datasource in the following query:

SELECT \* FROM (</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;SELECT \* FROM \[{'a': 1}\] AS arr1</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;CROSS JOIN \[{\'b\': 2}, {\'b\': 3}\] AS arr2</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;) AS derived

SELECT<sub>out</sub> = FROM<sub>out</sub> = \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨derived: {'a': 1, \'b\': 2}⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨derived: {'a': 1, \'b\': 3}⟩</br>
\]

The semantics for derived tables for FROM (_q_) AS _x_ are thus:

FROM<sub>out</sub> = \[ ⟨_x_: \$mergeObjects(_v<sub>0</sub>, ...,v<sub>n</sub>_)⟩</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;where _d_ = ⟨y<sub>0</sub>: _v<sub>0</sub>, ..., y<sub>n</sub> : v<sub>n</sub>_⟩</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;for _d_ ∊ _q_\]

where \$mergeObjects is the MQL function, which has semantics similar to
_tupleConcat_, but applied to documents rather than binding tuples.

##### Compound (Join) Datasources

A join datasource is a compound datasource that combines two other
datasources. The binding tuples created by the join contain the keys
from the two combined datasources. This requires that the sets of
datasource names created by each side of the join must be disjoint. If
the same datasource name appears on both sides of a join, the query will
fail to compile.

The number and contents of the tuples output by a join datasource
depends on the type of join and the join criteria. Behavior for each
join type is described below. MongoSQL supports INNER JOIN, (CROSS)
JOIN, LEFT OUTER JOIN, and RIGHT OUTER JOIN.

###### Rewrites

There are two types of JOIN that may be rewritten syntactically. We also
rewrite to have aliases as specified in the [Collection Datasource](#collection-datasource) section.

###### Comma (CROSS) JOIN Datasource

CROSS JOIN performs a mathematical cross product of two datasources. For
example, consider the output of the join datasource in the following
query:

SELECT \* FROM A AS a1 CROSS JOIN B AS b1

SELECT<sub>out</sub> = FROM<sub>out</sub> = \[⟨a1: a, b1: b⟩ for (a, b) ∊ A ⨯ B\]

Comma join:

\<datasource\>, \<datasource\>

will be rewritten as:

\<datasource\> CROSS JOIN \<datasource\>

###### INNER JOIN Datasource

Semantically, an INNER JOIN is equivalent to a CROSS JOIN filtered by a
WHERE clause. For example:

SELECT \* FROM X INNER JOIN Y ON \<condition\>

is equivalent to

SELECT \* FROM X CROSS JOIN Y WHERE \<condition\>

The same predicate typing restrictions apply when using ON as when using
WHERE. The only difference between an inner join's ON predicate and a
WHERE clause is that the only values in scope for the ON predicate are
those in the two datasources being joined.

For example, consider the output of the join datasource in the following
query. For the purpose of the formal definition, we consider the join
criteria \<condition\> a function that takes a binding tuple and returns
a boolean.

SELECT \* FROM A as a1 INNER JOIN B as b1 ON \<condition\>

SELECT<sub>out</sub> = FROM<sub>out</sub> = \[tup if \<condition\>(tup)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;where tup = ⟨a1: a, b1: b⟩</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;for (a, b) ∊ A ⨯ B</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\]

MongoSQL does not support the USING Clause at the initial version(see
[Future Work](#future-work)).

###### LEFT OUTER JOIN Datasource

Like in standard SQL, left outer joins in MongoSQL guarantee that every
tuple from the left side of the join appears at least once in the
result. The main way in which MongoSQL differs from standard SQL is that
we cannot necessarily enumerate all field names in a datasource. So, in
the cases where SQL would return null values for all fields on the right
side of a join, we instead set the value for all right-side datasource
names to the empty document.

SELECT \* FROM A AS a1 LEFT OUTER JOIN B AS b1 ON \<condition\>

SELECT<sub>out</sub> = FROM<sub>out</sub> = \[\..., ⟨a1: {\...}, b1: {\...}⟩, ⟨a1: {\...}, b1: {}⟩, \...\]

###### RIGHT OUTER JOIN Datasource

A right outer join is the inverse of a left outer join. Since MongoSQL
does not provide any guarantees about field order, the following queries
are semantically equivalent:

SELECT \* FROM A AS a1 LEFT OUTER JOIN B AS b1 ON \<condition\>

SELECT \* FROM B AS b1 RIGHT OUTER JOIN A AS a1 ON \<condition\>

<div id="join-examples" />

###### Examples

[Rewrite Tests](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/rewrite_tests/join.yml)\
[Query Tests](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/join.yml)

##### Unwind Datasource

The syntax for unwinding array fields is an `UNWIND` keyword that can be used in the `FROM` clause in conjunction with a datasource and options.

UNWIND(\<datasource\> WITH PATH => \<path\>,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;INDEX => \<name\>,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;OUTER => \<bool\>)

where

- `datasource` is the source of the array field
- `PATH` is a field path that references the field in the datasource to unwind
- `INDEX` is the name to assign the index column
- `OUTER` indicates whether documents with null, missing, or empty array values are preserved

<div id="unwind-semantics" />

###### Semantics

The `UNWIND` datasource "unwinds" an array field into multiple rows–each one corresponding
to a value from the array field. It is a datasource that operates on any valid
MongoSQL [datasource](#datasources) (collection, array, join, another unwind, etc.). The `UNWIND` datasource specifies a datasource and a sequence of named arguments. The only required
argument is a "`PATH`", which is a compound identifier that references an array field from the datasource. There are optional "`INDEX`" and "`OUTER`" arguments that behave similarly to their [$unwind counterparts](https://docs.mongodb.com/manual/reference/operator/aggregation/unwind/#document-operand-with-options), "`includeArrayIndex`" and "`preserveNullAndEmtpyArrays`", respectively. By default, if those arguments are not included, the output does not contain an index field and does not preserve documents for which the unwind field is null, missing, or an empty array.

Note that `UNWIND` operates on arbitrary datasources, meaning it could operate on simple
_or_ compound datasources. For the latter case, the `UNWIND` datasource outputs
multiple namespaces. Therefore, the `UNWIND` datasource cannot be aliased; instead,
it uses the alias(es) of the provided datasource. The field that is unwound is nested
along its field path under the alias of its datasource. See the [Examples](#unwind-examples)
for a more concrete view of this.

Generally, an `UNWIND` datasource can be thought of as a two-step process. First, the inner datasource is evaluated to produce a stream of binding tuples. Then, each binding tuple is sent through the "unwinding" process to produce 0 or more binding tuples, depending on the value of the field specified in the `UNWIND`. In the abstract model, this is actually modeled in one step as follows:

SELECT * FROM UNWIND(ds WITH PATH => array_path,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;INDEX => i)

SELECT<sub>out</sub> = FROM<sub>out</sub> = [⟨bk: {...d, array_path: val, i: idx}, ...bt⟩</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;for (idx, val) ∊ enumerate(d.array_path)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;where (bk: d) is the binding in bt that</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;contains array_path</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;for bt ∊ ds</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;]

For each binding tuple `bt` in datasource `ds`, determine the binding (`bk`, `d`) that
contains the unwind path according to the MongoSQL [Scoping Rules](#scoping-rules)
and then "`enumerate`" `d.array_path`. Here, `enumerate` means to range over the values
in the array along with the index for that value. For each index-value pair (`idx`, `val`)
in the array, output a binding tuple that binds "`bk`" to the document containing everything in `d`
(denoted as `...d`) with "`array_path`" replaced by the unwind value `val` and an "`i`"
field added with the index value `idx`, as well as all other bindings in `bt` (denoted as `...bt`).

There are several important notes to point out here. The [Examples](#unwind-examples) highlight these points in finer detail.

- If the value at the `PATH` for a certain document `d` is _not_ an array and is not `NULL` or missing, the abstract model `enumerate` function simply outputs the value at the path and the index value `NULL`.
- If the value at the `PATH` for a certain document `d` is an empty array, `NULL`, or missing, `bt` is omitted from the output by default.
  - If the optional argument `OUTER` is provided and set to `TRUE`, then `bt` is included in the output similar to how MongoDB [$unwind](https://docs.mongodb.com/manual/reference/operator/aggregation/unwind/#preservenullandemptyarrays) works.
  - For brevity, the abstract model equations in this document omit the steps that exclude such values by default.
  - Note that several competitors use `JOIN UNNEST` syntax to unwind array values from datasources.
    - A `CROSS JOIN` (the default) omits documents when the unwind path references a null, empty, or missing array.
    - A `LEFT JOIN` includes such documents.

- If the datasource is a compound datasource that contains multiple datasources,
`UNWIND` follows the MongoSQL [Scoping Rules](#scoping-rules) to determine which field
from which datasource is being referenced. The unwound data is nested under the appropriate datasource in the output.

  - To make this clear, here is how this looks according to the abstract model. For simplicity of notation in this example, we can assume `a` and `b` are both collection datasources:</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;SELECT * FROM UNWIND(a JOIN b WITH PATH => a.array_path)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;SELECT<sub>out</sub> = FROM<sub>out<sup> </sup></sub>= [</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨a: {...d<sub>a</sub>, array_path: val}, b: d<sub>b</sub>⟩</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;for (idx, val) in enumerate(d<sub>a</sub>.array_path)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;for (d<sub>a</sub>, d<sub>b</sub>)`∊`a ⨯ b</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;]

  - See the [Examples](#unwind-examples) for more details.

- If the `INDEX` name MAY<sup>[2](#unwind-definitions)</sup> or MUST conflict with a field name in the datasource, then the `UNWIND` source is semantically invalid and an error is produced.

- [MongoSQL uses 0-indexing](#array-indexing-and-slicing-expressions).

<div id="unwind-examples" />

###### Examples

[Spec tests](https://github.com/10gen/mongosql-rs/pull/282)

<div id="unwind-definitions" />

###### Definitions

There are three schema satisfaction levels:

- `MUST` contain `f` => we have a schema that proves a datasource is guaranteed to have the top-level field `f`
- `CANNOT` contain `f` => we have a schema that proves a datasource cannot contain the top-level field `f`
- `MAY` contain `f` => we cannot prove or disprove the existence of a top-level field `f` in a datasource

##### Flatten Datasource

_FLATTEN(\<datasource\> WITH depth => n, separator => \<some_string\>)_, where

- _depth => n_
  - `n` is an nonnegative integer
  - Optional and defaults to flattening all subdocuments
- separator => \<some_string\>
  - Use \<some_string\> as the delimiter when concatenating field names
  - Optional and defaults to `'_'`

<div id="flatten-semantics" />

###### Semantics

A `FLATTEN` datasource operates on any valid MongoSQL datasource (collection, array, join, another flatten, etc.),
some of which are simple datasources, while others are compound. If `FLATTEN` is operating
on a compound datasource, it will output multiple namespaces. Therefore, the `FLATTEN`
datasource cannot be aliased; instead, it uses the alias(es) of the provided datasource.
The flattened fields are nested under the alias of their datasource.

Refer to the [spec tests](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/from_flatten.yml) for a more concrete view of this.

`FLATTEN` creates a stream of binding tuples with the same keys as the input tuples; one binding tuple is created per document in the datasource. For example, consider the following query

SELECT * FROM FLATTEN(\<input datasource\> WITH DEPTH => n,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;SEPARATOR => s)

its output is defined as:

SELECT<sub>out</sub> = FROM<sub>out</sub> = [flatten_tuple(t, n, s) for t ∊ datasource]

where _flatten_tuple(t, n, s)_ is defined as follows:

for (datasource_name, root_value) in t:</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;emit_binding(datasource_name, flatten_doc(root_value, n, s))

and _flatten_doc(d, n, s)_ is the following recursive algorithm that produces a document:

```sh
for (k, v) in d:
    if v is not a document or n <= 0:
        emit(k, v)
    else:
        for (sub_k, sub_v) in flatten_doc(v, n-1, s):
            emit(k + s + sub_k, sub_v)
```

Note: `n` defaults to `∞` and `s` defaults to `'_'`.

In other words, the flattening process will recursively flatten each top-level field until it reaches a non-document subfield or the user-specified document depth. Here, `emit_binding(name, doc)` means return a binding of `name` to `doc`. `emit(k, v)` means include the key-value pair `(k, v)` in the result document. If the user sets the value of depth to 0, `flatten_doc` will be a no-op.

###### Schema Information

Flattening will return an error if

- the datasource contains polymorphic object fields
- we cannot enumerate all of the keys for a field given its document schema
- a naming collision `MAY` or `MUST`<sup>[1](#unwind-definitions)</sup> occur.

  - For example, suppose we have the following data:</br>

    ```sh
    { 'foo': { 'a_b': 1, 'a': { 'b': 2 } } }
    SELECT * FROM FLATTEN(foo)
    ```

     would produce `'a_b'` the original field, and `'a_b'` the flattened field.
     Those names conflict so MongoSQL will return an error.

<div id="flatten-examples" />

###### Examples

[Spec Tests](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/from_flatten.yml)

<div id="from-grammar" />

#### Grammar

\<from clause\> ::= FROM \<datasource\>

\<datasource\> ::= \<simple datasource\> \| \<compound datasource\> \| \<unwind datasource\> \| \<flatten datasource\>

\<simple datasource\> ::= [\<collection datasource\>](#collection-datasource)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| [\<array datasource\>](#array-datasource)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| [\<derived table datasource\>](#derived-table-datasource)

\<collection datasource\> ::= \<collection reference\> [\<alias\>](#aliases)?

\<collection reference\> ::= [\<compound identifier\>](#identifiers)

\<array datasource\> ::= [\<array expression\>](#array-indexing-and-slicing-expressions) [\<alias\>](#aliases)

\<derived table datasource\> ::= \"(\" [\<select query\>](#select-clause) \")\" [\<alias\>](#aliases)

\<compound datasource\> ::= \<join type\>

\<join type\> ::= [INNER](#inner-join-datasource)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| [CROSS](#comma-cross-join-datasource)?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| [LEFT OUTER](#left-outer-join-datasource)?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| [RIGHT OUTER](#right-outer-join-datasource)?

\<cross join\> ::= \<datasource\> CROSS? JOIN \<datasource\>

\<qualified join\> ::= \<datasource\> \<join type\> JOIN \<datasource\> \<join spec\>?

\<join spec\> ::= ON [\<expression\>](#expressions)

\<join column list\> ::= [\<compound identifier\>](#identifiers) (\",\" [\<compound identifier\>](#identifiers))\*

\<unwind datasource\> ::= UNWIND "(" \<datasource\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;(WITH \<unwind options\>)? ")"

\<unwind options\> ::= \<unwind option\> ("," \<unwind option\>)*

\<unwind option\> ::= \<path option\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \<index option\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \<outer option\>

\<path option\> ::= PATH "=>" [\<compound identifier\>](#identifiers) (\",\" [\<compound identifier\>](#identifiers))\*</br>
\<index option\> ::= INDEX "=>" \<identifier\></br>
\<outer option\> ::= OUTER "=>" \<boolean literal\>

\<flatten datasource\> ::= FLATTEN "(" \<datasource\> (WITH \<flatten option\> ("," \<flatten option\>)*)? ")"

\<flatten option\> ::= SEPARATOR "=>" \<string literal\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| DEPTH "=>" \<integer literal\>

## WHERE Clause

<div id="where-behavior" />

### Behavioral Description

A WHERE clause is a filter on the incoming binding tuples. Its output
binding tuples are those binding tuples for which the condition is met.

Formally, for WHERE e</br>
error if e(x) IS NOT {BOOLEAN, NULL, MISSING} (statically)</br>
otherwise</br>
WHERE<sub>out</sub> = \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;x if e(x)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;for x ∊ WHERE<sub>IN</sub></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\]

Of note here is that we report an error if e(x) is not statically
guaranteed to be a BOOLEAN. The expression, e, must statically return a
boolean, either by using a CAST to BOOLEAN, or by using only expressions
that are guaranteed to return BOOLEAN, such as comparison operators,
AND, and OR, or a CASE statically determined to always return BOOLEAN.
Note that NULL (and MISSING) being distinct from TRUE, causes the
current document, x, to not be included in the result set. This is
consistent with every major implementation of SQL.

<div id="where-grammar" />

### Grammar

\<where clause\> ::= WHERE [\<expression\>](#expressions)

<div id="where-examples" />

### Examples

[Query Tests](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/where.yml)

<div id="where-rejected" />

### Rejected Alternatives/Competitive Analysis

Many different implementations of SQL automatically coerce WHERE
arguments to BOOLEANs. Some, like MySQL, do not even have a BOOLEAN
type. The behavior of requiring the expression to be a BOOLEAN is
consistent with Postgres and PartiQL, however, except that PartiQL makes
this determination dynamically, while we prefer the route of statically
ensuring the argument will be a BOOLEAN (or NULL/MISSING). Having
queries fail during execution due to polymorphic data is something we
wish to avoid because analytic queries can often take hours to run.

## GROUP BY Clause

<div id="group-by-behavior" />

### Behavioral Description

GROUP BY provides a means for grouping data into equivalence classes.
Aggregations can then be done on these equivalence classes. The main
form of GROUP BY is as follows:

GROUP BY e<sub>1</sub> AS x<sub>1</sub> , ..., e<sub>m</sub> AS x<sub>m</sub> AGGREGATE agg_function(e) AS
y<sub>1</sub>, ... agg_function\'(e\') AS y<sub>n</sub>

The expressions e<sub>1</sub> ... e<sub>m</sub>, when taken as an array, provide the
equivalence class key. For each array value achieved by evaluating these
expressions on a document, a separate group is formed. Unless it is a
column reference, each element of the equivalence class key must have an
alias for reuse purposes after the GROUP stage. Aliases are
automatically generated where needed using the alias rules in Section
[Rewrites for Generation of GROUP Aliases](#rewrites-for-generation-of-group-aliases).

The output of the clause contains one binding tuple per unique value of
the equivalence class key with the values of the group equivalence class
key and the results of the aggregation functions named as mentioned.
Top-level MISSING values are converted to NULL in GROUP BY expressions,
as in all arrays. GROUP keys use the semantics of MongoSQL =, except
that all NULLs are considered equivalent. Specifically this means that
the double 3.0 and the integer 3 are grouped into the same group.
Because of this, GROUP keys must be statically proved to be comparable
to each other via equality. When multiple numeric types are grouped into
one group, the type in the output group chosen is undefined behavior and
implementation specific<sup>[2](#2)</sup>. The arguments to agg_function can be any
expression.

### ALL and DISTINCT agg_functions

ALL does nothing in agg_functions as that is the default behavior. It is
removed during syntactic rewrite. Thus:

agg_function(ALL e)

will be rewritten as:

agg_function(e)

DISTINCT agg_functions only consider distinct elements of the groups
over which they are aggregating. Distinctness is defined using the
MongoSQL equality operator, except that NULL is considered equivalent to
NULL, and MISSING is converted to NULL. Thus, arguments to DISTINCT
agg_functions must be statically proved to be comparable to each other
via equality.

### GROUP BY Clause Output

The GROUP BY clause, like all clauses, outputs a stream of binding
tuples. As described above, the output stream contains one tuple for
each group key equivalence class. The binding tuples contain only the
group keys and aggregates defined in the GROUP BY clause.

Consider the following example query and the output tuples generated by
its GROUP clause. In this example, all group keys are aliased, so every
field is nested under the ⟘ datasource name in the output tuples:

SELECT \* FROM foo GROUP BY foo.a AS key AGGREGATE sum(a) AS sum;</br>
SELECT<sub>in</sub> = GROUP<sub>out</sub> = \[⟨⟘: {key: \<value\>, sum: \<value\>}⟩,
\...\]

If there is one or more unaliased group keys, those keys are nested
under their original namespaces in the binding tuple:

SELECT \* FROM foo GROUP BY a AGGREGATE sum(a) AS sum;</br>
SELECT<sub>out</sub> = GROUP<sub>out</sub> = \[⟨⟘: {sum: \<value\>}, foo: {a: \<value\>}⟩,
\...\]

> <sup id="2">2</sup> mongod 4.2 chooses the first type seen during grouping.

### Rewrite implicit GROUP BY to explicit GROUP BY NULL

Using an aggregate function in the select clause creates an implicit
GROUP BY where only one group is created. The explicit GROUP BY we
rewrite this to uses NULL as the group key, though grouping by any
constant value is equivalent.

SELECT ..., ... agg_function(x) ... AS y, ... FROM \<from item\>

will be rewritten as:

SELECT ..., ... agg_function(x) ... AS y, ... FROM \<from item\> GROUP
BY NULL

### Rewrite SELECT clause aggregate functions into AGGREGATE clause

SELECT ..., ... agg_function(e) ..., ... FROM \<from item\> GROUP BY ...
AGGREGATE ...

will be rewritten as:

SELECT ..., ... \_aggN ..., ... FROM \<from item\> GROUP BY ...
AGGREGATE agg_function(e) AS \_aggN, ...

Duplicated aggregate function expressions, e.g., SUM(x + 1) and
SUM(x + 1) should only have one computation in the AGGREGATE section for
efficiency reasons. The same alias can be used in each position.

### Rewrite HAVING clause aggregate functions into AGGREGATE clause

GROUP BY ... HAVING ... agg_function(e) ...

will be rewritten as:

GROUP BY ... AGGREGATE agg_function(e) AS \_aggn HAVING ... \_aggn ...

As above, duplicated aggregation function expressions will be rewritten
to occur only once in the AGGREGATE phrase of the clause.

### Rewrites for Generation of GROUP Aliases

When the expressions e<sub>1</sub> ... e<sub>n+m</sub> are non-reference expressions

GROUP BY e<sub>1</sub>, ..., e<sub>n</sub> AGGREGATE e<sub>n+1</sub>, ..., e<sub>n+m</sub>

will be rewritten as:

GROUP BY e<sub>1</sub> AS _groupKey1, ..., e<sub>n</sub> AS \_groupKeyn

AGGREGATE e<sub>n+1</sub> AS _agg1, ..., e<sub>n+m</sub> AS \_aggm

When an expression in e<sub>1</sub> ... e<sub>n</sub> is a reference to a top-level field,
we do not generate an alias for the field, so that it can be referenced
by its fully qualified name in subsequent clauses. Since we can't
distinguish between top-level field references and subpaths during
syntactic rewrites, we will skip generating an alias for any GROUP BY
key that might resemble a top-level field.

When both SELECT and HAVING contain agg_function expressions, the SELECT
agg_function expressions are added to AGGREGATE first, followed by
HAVING. If there is duplication of agg_function expressions, the
position is determined by the first instance of the duplicated
agg_function expression. This necessarily has an impact on alias
generation. For example:

SELECT SUM(x), COUNT(y)</br>
FROM foo</br>
GROUP BY z</br>
HAVING COUNT(y) + SUM(x) + SUM(y) > 20</br>

will be rewritten as:

SELECT VALUE {\'\_1\': \', \_agg1, \'\_2\': \_agg2}</br>
FROM foo AS foo</br>
GROUP BY z AS z</br>
AGGREGATE SUM(x) AS \_agg1, COUNT(y) AS \_agg2, SUM(y) AS \_agg3</br>
HAVING \_agg2 + \_agg1 + \_agg3 > 20

### GROUP BY Alias

MongoSQL supports grouping by aliases. For example, MongoSQL can translate the query `SELECT calculation AS calc FROM foo GROUP BY calc`.

The GROUP BY clause, like all clauses, outputs a stream of binding tuples: one tuple for each group key equivalence class. The binding tuples contain only the group keys and aggregates defined in the GROUP BY clause.

Consider the following example query and the output tuples generated by its GROUP clause. Aliased group keys are nested under the ⟘ datasource name in the output tuples, while unaliased group keys are nested under their original namespaces:

SELECT * FROM foo GROUP BY a AGGREGATE sum(a) AS sum;
SELECT<sub>out</sub> = GROUP<sub>out</sub> = [⟨⟘: {sum: <value>}, foo: {a: <value>}⟩, ...]

MongoSQL rewrites GROUP BY keys that are aliases in the SELECT clause to the corresponding aliased expression in the SELECT clause.

SELECT e1 AS x1, e2 AS x2, … GROUP BY x1, x2, …

will get rewritten to:

SELECT x1, x2 … GROUP BY e1 AS x1, e2 AS x2, …

### Aggregation Functions

The following are the aggregation functions supported by MongoSQL. Each
one is evaluated on all of the specified elements from each value in a
group as determined by the group key value.

- ADD_TO_ARRAY - Pushes the argument to the end of an array, the total
  output of this function will be an array.

  - The type of the argument to ADD_TO_ARRAY does not matter.

- ADD_TO_SET - Pushes the argument to the end of an array removing
  duplicates, the total output of this function will be an array
  with all duplicate items removed. Duplicates are determined using
  MongoSQL\'s = operator. Note that ADD_TO_SET(x) is equivalent to
  ADD_TO_ARRAY(DISTINCT x), and maintained for compatibility with
  MQL.

  - The type of the argument to ADD_TO_SET does not matter.

- AVG - Takes the average of all the arguments.

  - The argument must be statically typed to a numeric type.

- COUNT - Counts the number of elements. COUNT(\*) counts all values
  unconditionally,
  COUNT([\<expression\>](#expressions)) counts all
  values for which the expression does not result in NULL or
  MISSING.

  - The type of the argument to COUNT does not matter.

- FIRST - Returns the first element in the group. Deterministic only
  when the input has deterministic order, otherwise undefined.

  - The type of the argument to FIRST does not matter.

- LAST - Returns the first element in the group. Deterministic only
  when the input has deterministic order, otherwise undefined.

  - The type of the argument to LAST does not matter.

- MAX - Returns the max element as ordered by the MongoSQL > operator.

  - The argument must be statically typed to be comparable via the > operator.

- MERGE_DOCUMENTS - Returns a document formed by successively merging
  documents, with the previous element used as the left hand side.
  In the case of duplicate keys, the value of the key in the new
  element is kept.

  - The argument must be statically typed as DOCUMENT, and thus
    MERGE_DOCUMENTS(DISTINCT x) is not allowed at this time. As
    with FIRST and LAST, the output is only deterministic when the
    input has deterministic ordering.

- MIN - Returns the min element as ordered by the MongoSQL < operator.

  - The argument must be statically typed to be comparable via the < operator.

- STDDEV_POP - Returns the standard deviation of all elements over the
  entire group population.

  - The argument must be statically typed to a numeric type. See
    more [Here](https://docs.mongodb.com/manual/reference/operator/aggregation/stdDevPop).

- STDDEV_SAMP - Returns the standard deviation of a sample of all
  elements in the group.

  - The argument must be statically typed to a numeric type. See
    more [Here](https://docs.mongodb.com/manual/reference/operator/aggregation/stdDevSamp).

- SUM - Takes the sum of all the arguments.
  - The argument must be statically typed to a numeric type

## HAVING clause

The HAVING clause operates the same as a WHERE clause, but after the
GROUP BY clause, meaning it can reference aliases defined in the GROUP
BY and can contain expressions with agg_functions; only aliases defined
in the GROUP BY are available to the HAVING clause. This is necessary
for filters that need the values computed by the GROUP BY, as the WHERE
clause is applied before GROUP BY. Just as a WHERE clause, the HAVING
clause takes an expression that must statically have type BOOL or NULL
and may evaluate to MISSING. HAVING agg_function applications will be
rewritten as follows:

GROUP BY ... AGGREGATE ... HAVING ... agg_function(e) ...

will be rewritten as:

GROUP BY ... AGGREGATE ... agg_function(e) AS \_n ... HAVING ... \_n ...

The alias \_n is derived numbering left to right as in Section [SELECT Clause](#select-clause).

<div id="having-grammar" />

### Grammar

\<group by clause\> ::= GROUP BY \<group key list\> \<aggregations\>?

\<having clause\> ::= HAVING [\<expression\>](#expressions)

\<group key list\> ::= \<group key\> (\",\" \<group key\>)\*

\<group key\> ::= [\<expression\>](#expressions) (AS?  [\<identifier\>](#identifiers))?

\<aggregations\> ::= AGGREGATE \<aggregation function application\> AS?
[\<identifier\>](#identifiers) (\",\" \<aggregation
function application\> AS?
[\<identifier\>](#identifiers))\*

\<aggregation function application\> ::= \<aggregation function\> \"(\" (DISTINCT \| ALL)?
[\<expression\>](#expressions) (\",\"
[\<expression\>](#expressions))\* \")\"

\<aggregation function\> ::= ADD_TO_ARRAY \| ADD_TO_SET \| AVG \| COUNT</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| FIRST \| LAST \| MAX \| MERGE_OBJECTS \| MIN \| PUSH \| STDDEV_POP</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| STDDEV_SAMP \| SUM

<div id="having-examples" />

### Examples

[Query Test](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/group_by.yml)</br>
[Rewrite Test](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/rewrite_tests/group_by.yml)

## ORDER BY Clause

<div id="order-by-behavior" />

### Behavioral Description

SQL's ORDER BY clause provides a way to order a result set by one or
more sort keys. Each sort key can be a column reference, or can be an
integer literal referring to a SELECT expression by its position in the
select expr list. Sort keys that are column references can be compound
identifiers. These compound identifiers can be qualified with datasource
names or refer to document subfields. Name resolution follows the
[Scoping Rules](#scoping-rules).

The semantics for MongoSQL's ordering are consistent with the behavior
described by section 13.1 of the SQL-92 Specification, which we will
quote here instead of attempting to rephrase:

> 3) If an \<order by clause\> is specified, then the ordering of rows
  of the result is effectively determined by the \<order by clause\>
  as follows:
>> a) Each \<sort specification\> specifies the sort direction for
the corresponding sort key Ki. If DESC is not specified in
the i-th \<sort specification\>, then the sort direction for Ki
is ascending and the applicable \<comp op\> is the \<less than
operator\>. Otherwise, the sort direction for Ki is descending
and the applicable \<comp op\> is the \<greater than operator\>.</br></br>
>>b) Let P be any row of the result table and let Q be any other
row of that table, and let PVi and QVi be the values of Ki
in these rows, respectively. The relative position of rows
P and Q in the result is determined by comparing PVi and
QVi according to the rules of Subclause 8.2, \"\<comparison
predicate\>\", where the \<comp op\> is the applicable \<comp op\>
for Ki, with the following special treatment of null values.
Whether a sort key value that is null is considered greater
or less than a non-null value is implementation-defined, but
all sort key values that are null shall either be considered
greater than all non-null values or be considered less than
all non-null values. PVi is said to precede QVi if the value
of the \<comparison predicate\> \"PVi \<comp op\> QVi\" is true for
the applicable \<comp op\>.</br></br>
>>c) In the result table, the relative position of row P is before
row Q if and only if PVn precedes QVn for some n greater than
0 and less than the number of \<sort specification\>s and PVi
= QVi for all i < n. The relative order of two rows that are
not distinct is implementation-dependent.

The only addition we make to the ordering behavior described in the SQL
92 Specification is to clarify an implementation-defined behavior:

- MongoSQL sorts MISSING before NULL, and NULL before all other values
  (this is consistent with the behavior of MongoSQL's less-than
  operator)

#### Rewrite: Positional sort keys to references

All positional select-expression references in the ORDER BY clauses are
rewritten to be expressions. We check each sort key to determine whether
it needs to be rewritten; if a sort key is any expression other than an
integer literal, no transformation is needed.

If a sort key _is_ an integer literal, then it is treated as a
one-indexed reference to the list of select expressions. We perform a
syntactic rewrite, substituting the alias of the select expression from
the indicated position as the new sort key expression.

For example,

SELECT e1 AS a, e2 AS b FROM foo ORDER BY 1, 2

is rewritten to

SELECT e1 AS a, e2 AS b FROM foo ORDER BY a, b

There are a few circumstances under which this rewrite will fail.
Queries that lead to these circumstances are not allowed in MongoSQL:

- A positional sort key is used with SELECT VALUE
- A positional sort key is used with a select list containing a star
  expression

#### Rewrite: Implicit to Explicit ASC

The default sort key direction is ASC if not specified by the user. This
is made explicit by a rewrite, where

... ORDER BY e, ...

is rewritten to:

... ORDER BY e ASC, ...

#### Type Constraints

The ORDER BY clause requires that all possible values in a sort key
expression can be statically verified to be comparable via the > and \<
operators.

<div id="order-by-grammar" />

### Grammar

\<order by clause\> ::= ORDER BY \<sort specification\> (, \<sort
specification\> )\*

\<sort specification\> ::= \<sort key\> \<sort direction\>?

\<sort key\> ::= [\<compound identifier\>](#identifiers)
\| [\<integer literal\>](#literals)

\<sort direction\> ::= ASC \| DESC

### Open Questions

- Is MongoDB's \$sort behavior consistent with its \$lt and \$gt
  behavior?

  - **\[Answer\]** Maybe. We\'ll have to discover during
    implementation

## LIMIT, FETCH FIRST, and OFFSET Clauses

FETCH FIRST is an alias for LIMIT. MongoSQL supports LIMIT and FETCH FIRST
interchangeably. Throughout this document, all specifications for LIMIT apply
to FETCH FIRST.

<div id="limit-behavior" />

### Behavioral Description

LIMIT and OFFSET allow users to retrieve only part of the rows generated
by a query. Both LIMIT and OFFSET numbers have to be positive integers.
Using LIMIT/OFFSET without ORDER BY does not guarantee the same result.

If a LIMIT number is provided, no more than that number of rows will be
returned. If an OFFSET number is provided, that number of rows is
skipped before returning rows.

When LIMIT and OFFSET are both set, the first OFFSET rows will be
skipped before returning the rest of the results which should contain no
more than the LIMIT number rows. _LIMIT i, j is a shorter form of LIMIT
i OFFSET j._

The LIMIT and OFFSET can be used in subqueries, unlike in some SQL
dialects.

Formally, for LIMIT i

LIMIT<sub>out</sub> = \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;x</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;for x ∊ \[x<sub>1</sub>, ..., x<sub>i</sub>\]</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;where \[x<sub>1</sub>, ..., x<sub>i</sub>, ..., x<sub>j</sub>\] = LIMIT<sub>IN</sub><sup>i</br>
\]

Formally, for OFFSET i

OFFSET<sub>out</sub><sup>i</sup> = \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;x</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;for x ∊ \[x<sub>i+1</sub>, ..., x<sub>j</sub>\]</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;where \[x<sub>1</sub>, ..., x<sub>i</sub>, ..., x<sub>j</sub>\] = OFFSET<sub>IN</sub><sup>i</sup></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\]

<div id="limit-grammar"/>

### Grammar

\<limit clause\> ::= LIMIT [\<integer literal\>](#literals) (\",\" [\<integer literal\>](#literals) )?</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| FETCH (FIRST | NEXT) [\<integer literal\>](#literals) ROW(S?) ONLY`

\<offset clause\> ::= OFFSET [\<integer
literal\>](#literals)

<div id="limit-rewrites" />

### Rewrites

LIMIT i, j

will be rewritten as

LIMIT i OFFSET j

<div id="limit-examples" />

### Examples

[Query Test](https://github.com/10gen/mongosql-rs/blob/e31c8dc2ccf88b0d648877e3d0312f92d969c228/tests/spec_tests/query_tests/limit_offset.yml)</br>
[Rewrite Test](https://github.com/10gen/mongosql-rs/blob/e31c8dc2ccf88b0d648877e3d0312f92d969c228/tests/spec_tests/rewrite_tests/limit_offset.yml)

<div id="limit-rejected" />

### Rejected Alternatives/Competitive Analysis

The LIMIT and OFFSET clauses are not part of SQL-92 standard, and
different database providers have slightly different implementations.
Since SQL-2008, the FETCH FIRST clause has become part of the SQL
standard. However, it's not the most widely supported syntax so far.

From this [table](<https://en.wikipedia.org/wiki/Select_(SQL)#FETCH_FIRST_clause>),
you can see LIMIT/OFFSET is best supported among DB providers. In
addition to the DB providers listed in that table, PartiQL and Presto
also support LIMIT and OFFSET. Also because we are targeting SQL-92
compatibility, SQL 2008 is not a requirement, so we choose to support
LIMIT/OFFSET syntax over FETCH FIRST.

## Set Operations

<div id="set-behavior" />

### Behavioral Description

MongoSQL provides a UNION ALL set operator for taking the union over the
result sets of two select queries. The UNION ALL operator does not
remove duplicate rows from the result set. The result set returned by
the UNION ALL operator does not have a defined order.

MongoSQL does not support distinct UNION (see [Future
Work](#future-work)). MongoSQL does not support the
INTERSECT or EXCEPT set operations.

UNION ALL outputs all the documents from each side of the UNION ALL. For
example, consider the output of the UNION ALL in the following query:

SELECT \* FROM X UNION ALL SELECT \* FROM Y

UNION<sub>out</sub> = \[x for x ∊ X, y for y ∊ Y\]

<div id="set-grammar"/>

### Grammar

\<set query\> ::= [\<select query\>](#select-clause) \<set
operator\> [\<select query\>](#select-clause)

\<set operator\> ::= UNION ALL?

### Open Questions

- Should we find a way to unify the union set operator and union join
  in the abstract model? It's a little strange to have two different
  places in the semantics that we're implementing what is
  essentially the same operation

  - **Proposed Answer:** Yes, that would be a good thing in the long
    run, but not a priority at the moment. The semantics currently
    outlined in this section would all be forward-compatible with
    such a unification.

  - The work for this unification is tracked in
    [SQL-118](https://jira.mongodb.org/browse/SQL-118)

<div id="set-rejected" />

### Rejected Alternatives/Competitive Analysis

- INTERSECT and EXCEPT

  - Though these are part of the SQL-92 spec, they are not supported
    by MongoSQL

  - Since not all SQL databases support them (notably MySQL), it
    seems unlikely that BI Tools will require them

  - If needed, they can be added in the future without breaking
    changes

  - MQL does not support any set operations other than UNION, so
    adding support for INTERSECT and EXCEPT would require
    non-trivial additions to the underlying database engine

#### Ambiguous Bindings

Unless we can prove statically that there are no duplicate top-level
keys in the documents to be merged, we disallow the query with a static
error. If the keys of documents in the subquery's output binding tuples
cannot be statically enumerated at query planning time (in the case of a
SELECT \*, for example), we assume that conflicts exist and raise a
static error:

SELECT \* FROM (</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;SELECT \* FROM foo AS foo</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;CROSS JOIN bar AS bar</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;) AS derived

Will result in an error:

The keys of datasources \'foo\', \'bar\' are not enumerable, and may be
ambiguous. Try replacing \'SELECT \*\' with direct references, or
providing schemata for \'foo\', \'bar\'.

<div id="ambiguous-rejected" />

##### Competitive Analysis And Rejected Alternatives

Different implementations of SQL handle ambiguous references in
different ways:

- MySQL reports duplicate key errors in the same cases we will, in so
  far as all MySQL tables have schemata

- Postgres allows duplicate keys to be displayed from a SELECT \* at
  the top, that is SELECT \* FROM (SELECT \* FROM foo AS f, foo AS
  f2) AS derived is allowed. However, there is no way to reference a
  key from the derived table in the SELECT clause, only SELECT \*
  works. Additionally, top level queries may reference a column
  multiple times.

- SQLite also allows duplicate keys as Postgres, but the SELECT \*
  results slightly mangle the names of duplicate columns such that
  no name is strictly duplicated; they are still not accessible.

- PartiQL allows duplicate field names in results, so does nothing
  special to handle ambiguous fields in derived tables. If a field
  is selected by name, the first field with that name from left to
  right is chosen.

- Presto operates similar to Postgres, allowing duplicate keys to be
  displayed from a SELECT \* at the top, that is SELECT \* FROM
  (SELECT \* FROM foo AS f, foo AS f2) AS derived is allowed.
  However, there is no way to reference a key from the derived table
  in the SELECT clause, only SELECT \* works.

While any of these approaches are feasible for MongoSQL, we choose to
report ambiguity as an error as early as possible (at the point of the
derived table query rather than at the uses of those derived values).
This results in an easier implementation and better immediate feedback
to users who may be writing queries that take hours or days.
Additionally, duplicate keys in MongoDB are undefined behavior.

## Expressions

<div id="expressions-grammar" />

### Grammar

\<expression\> ::= [\<unary operator expression\>](#operators)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| [\<binary operator expression\>](#operators)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<is operator expression\>](#semantics-of-type-operators)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<like operator expression\>](#semantics-of-string-operators)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<between operator expression\>](#semantics-of-comparison-operators)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<case operator expression\>](#conditional-scalar-functions)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<scalar function expression\>](#scalar-functions)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<subquery expression\>](#subquery-expressions)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<document expression\>](#document-and-field-access-expressions)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<document field access expression\>](#document-and-field-access-expressions)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<array expression\>](#array-indexing-and-slicing-expressions)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<array index expression\>](#array-indexing-and-slicing-expressions)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<array slice expression\>](#array-indexing-and-slicing-expressions)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<identifier\>](#identifiers)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<parenthesized expression\>](#parenthesized-expressions)

### Identifiers

#### Behavioral Description

Identifiers in MongoSQL refer to databases, tables, and columns. The
SQL-92 spec says that by default identifiers should only contain simple
latin letters, digits, and underscores, but it also explains that in
different contexts the character set could be expanded. MongoSQL
identifiers support _most_ utf-8 characters. The only exception to this
is the null character, \'\\x00\', which is disallowed in BSON keys
according to the [BSON
spec](http://bsonspec.org/spec.html).

There may be some semantic meaning associated with certain characters.
For example, [the grammar](#identifiers-grammar) proposes using the
"." character to create compound identifiers. The rest of this document
includes other examples of such characters. SQL-92 defines _delimited_
(quoted) and _regular_ (unquoted) identifiers. In MongoSQL, regular
identifiers are restricted to a limited subset of identifiers to avoid
conflicts with characters that have other semantic meaning; for an
identifier to include such a character, it must be delimited. For
example, among other things, an identifier must be delimited if it
begins with a digit or if it conflicts with a reserved keyword.

There should exist a bijection (a one-to-one mapping) from MongoSQL
identifiers to valid BSON keys. Given the above restrictions on regular
identifiers, this bijection exists specifically from delimited MongoSQL
identifiers to BSON keys.

Since a valid BSON key could contain the identifier delimiter characters
(\" and \`), MongoSQL identifiers must be able to contain these
characters. Of course, to include only double quotes or only backticks,
the other delimiter can be used (i.e. \`\"quoted\"\` or
\"\`backticked\`\"). To include a delimiter character in an identifier
delimited by that same character, double it<sup>[3](#3)</sup> (i.e.
\"contains_one\"\"\" or \`contains_one\`\`\`, which correspond to
contains_one\" and contains_one\`, respectively).

Identifiers should always be case-sensitive, whether delimited or not.
This is consistent with [MongoDB
identifiers](https://docs.mongodb.com/manual/reference/limits/#naming-restrictions)
for database, collection, and field names.

> <sup id="3">3</sup> This "double it" practice is used by
    [PostgreSQL](https://www.postgresql.org/docs/13/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS)
    and
    [MySQL](https://dev.mysql.com/doc/refman/5.7/en/identifiers.html).

#### Aliases

Identifiers are used for all aliases in MongoSQL. In most cases, it is a
static semantic error for aliases to be used more than once in the same
MongoSQL clause. The exception to this is that aliases can be repeated
on both sides of a UNION. This also applies to automatically generated
aliases (see Sections [SELECT Clause](#select-clause) and
[FROM Clause](#from-clause)) conflicting with user aliases
or user field names.

#### Keywords

MongoSQL keywords (such as SELECT, FROM, JOIN, etc.) are generally not
allowed to be used as undelimited identifiers, to simplify
lexing/parsing.

We reserve the right to add new keywords to MongoSQL in the future,
though we will never introduce a keyword that begins with an underscore.
This means that a user who wants an ironclad guarantee that a query will
not start failing with a newer version of MongoSQL because it uses an
identifier that has become a keyword should delimit all identifiers that
do not begin with an underscore.

In most cases, however, such a drastic approach is likely unnecessary.
Some mitigating factors to consider:

- If a new keyword is introduced, it is very likely that it will be a
  keyword in another existing SQL dialect, or a stage name in MQL
- When possible, we will allow newly introduced keywords to continue
  to be used as identifiers

<div id="identifiers-grammar" />

#### Grammar

\<compound identifier\> ::= [\<identifier\>](#identifiers) (\".\" [\<compound identifier\>](#identifiers))?

[\<identifier\>](#identifiers) ::= \<regular identifier\> \| \<delimited identifier\>

\<regular identifier\> ::= (\[A-Za-z\] \| \"\_\")\[A-Za-z0-9\_\]\*

\<delimited identifier\> ::= \" \<identifier character\>\* \" \| \` \<identifier character\>\* \`

\<identifier character\> ::= \[\^\\x00\]

#### Examples

See Identifiers spec tests
[here](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/identifier.yml)
and
[here](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/identifier_literal_null_char.yml).

See Aliases spec tests
[here](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/alias.yml).

#### Open Questions

- Should we require that identifiers are valid UTF-8?

  - **Proposed Answer:** Require UTF-8 identifiers to start, relax
    the requirement later on if it becomes an actual problem for
    customers.

  - Though the BSON spec requires key names to be valid UTF-8,
    MongoDB does not actually validate this, and so may contain
    documents with non-UTF-8 keys.

  - If we require our identifiers to be valid UTF-8, it would mean
    that we could not construct an identifier to reference
    non-UTF-8 fields that, despite being invalid BSON, may still
    exist in a MongoDB collection.

#### Rejected Alternatives/Competitive Analysis

##### Special characters in identifiers

The proposed solution is to use delimited identifiers when semantically
significant characters, i.e. ".", are part of an identifier. An
alternative to this would be to introduce a mechanism for escaping
characters in identifiers, such as an escape character. This is not
desirable since there is no precedent for it in other SQL dialects and
is not required or even considered by the SQL-92 spec. Along with that,
there is no equivalent mechanism in MongoDB, so introducing one in
MongoSQL seems unnecessary and unreasonable.

##### Case sensitivity

SQL-92 specifies how comparisons between identifiers should behave.
Comparisons between regular identifiers and regular identifiers are
case-insensitive; comparisons between regular identifiers and delimited
identifiers are case-insensitive; comparisons between delimited
identifiers and delimited identifiers are case-sensitive ([SQL-92
5.2.10-14](https://www.contrib.andrew.cmu.edu/~shadow/sql/sql1992.txt)).
It seems some other dialects deviate from this behavior:

- Partiql seems to be [case-sensitive](https://docs.aws.amazon.com/qldb/latest/developerguide/ql-reference.query.html)

- N1QL is [case-sensitive](https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/identifiers.html)

- SQL++ does not specify

- Presto is [case-insensitive](https://github.com/prestosql/presto/wiki/Delimited-Identifiers)
  but it looks like they built (or are building?) an API to
  configure case-sensitivity

- Drill is [case-insensitive](https://drill.apache.org/docs/lexical-structure/#case-sensitivity);
  however, data sources may be case-sensitive.

Given these deviations, in addition to MongoDB's native
case-sensitivity, it is reasonable for MongoSQL to always consider all
identifiers case-sensitive.

### Literals

#### Behavioral Description

MongoSQL supports literals for booleans, null, numbers, and strings.
Booleans, null, and strings are represented as expected, as seen in the
[Grammar](#literals-grammar). Importantly, strings are enclosed
in single quotes. To include a single quote character in a string
literal, double it.<sup>[4](#4)</sup> Numbers are slightly more nuanced: literal
integers are typed as INT when possible (i.e. values within the int32
range) and LONG otherwise, and literal floating point numbers or
scientific notation numbers are always considered to have type DOUBLE.
Note that the grammar does not specify signs for numeric literals. To
write a literal negative number, users can use the unary minus operator
before the numeric literal (this is effectively the same as supporting
literal negative numbers).

MongoSQL does not support literals for every type, for example OBJECTID,
BSON_DATE, and DECIMAL have no literal syntax. For such types,
pseudo-literal values can be obtained by using the CAST operator to go
from a string or numeric representation of those types to their
respective type. Some types may also have \"constructor\" functions
which alias the relevant CAST invocations. See the [Type
Conversions](#type-conversions) section for more details.

> <sup id="4">4</sup> The "double it" practice is used by [PostgreSQL](https://www.postgresql.org/docs/13/sql-syntax-lexical.html#SQL-SYNTAX-CONSTANTS) and [MySQL](https://dev.mysql.com/doc/refman/5.7/en/string-literals.html).

<div id = "literals-grammar" />

#### Grammar

\<literal\> ::= \<null literal\> \| \<boolean literal\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<string literal\> \| \<numeric literal\>

\<null literal\> ::= NULL

\<boolean literal\> ::= TRUE \| FALSE

\<string literal\> ::= \' \<any utf-8 character\>\* \'

\<numeric literal\> ::= \<integer literal\> \| \<double literal\>

\<integer literal\> ::= 0 \| (\[1-9\] \[0-9\]\*)

\<double literal\> ::= (\<integer literal\> \".\" \<exp component\>?)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| (\<integer literal\>? \".\" \[0-9\]+ \<exp component\>?)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| (\<integer literal\> \<exp component\>)<sup>[5](#5)</sup>

\<exp component\> ::= (e \| E) (\"+\" \| \"-\")? \[0-9\]+<sup>6(#6)</sup>

> <sup id="5">5</sup>  Note: MongoDB treats all numbers with exponent components as doubles, so MongoSQL does the same, even if the number is an integer (i.e. 1.2e2).</br>
> <sup id="6">6</sup>  Intentionally allowing multiple 0s after the “e”, as do many SQL dialects.

#### Examples

See spec tests
[here](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/literal.yml).

#### Open Questions

- Should we require that string literals are valid UTF-8?

  - **Proposed Answer:** Require UTF-8 string literals to start,
    relax the requirement later on if it becomes an actual problem
    for customers.

  - The BSON spec technically requires this, but MongoDB does not
    check it during validation.

  - We don't have any data on how common it is for customers to have
    invalid UTF-8 strings in practice.

  - If we do require string literals to be valid UTF-8, it could
    make it difficult for customers to write certain queries over
    data containing non-UTF-8 strings

#### Rejected Alternatives/Competitive Analysis

Extended JSON was considered as an option for representing all
BSON/MongoSQL types as literals, however this was ultimately rejected.
First, it is not clear that users would benefit from the ability to
represent all BSON types as literals. Also, while extended JSON is
technically human-readable and writable, it is not an ergonomic
experience to type out literals in that form. Lastly, the object
notation used in MongoSQL is JSON-like but not guaranteed to be
completely JSON-compliant, so interspersing it with extended JSON values
could prove to be challenging from a parsing perspective.

MongoSQL does not have a literal syntax for every data type. For
example, regular expression literals of the form /pattern/options could
be supported instead of or in addition to the REGEX() function.

### Parenthesized Expressions

A parenthesized expression is an expression grouped by parentheses,
similar to the majority of programming languages and other
implementations of SQL92 compliant SQL dialects. Any time infix
operators are present, the need for parentheses (or a similar mechanism)
to distinguish order of operations may be necessary. MongoSQL has
several infix operators, such as \'+\' and \'::\'. For example, the
value of 1 + 2 \* 3 is 7, while the value of (1 + 2) \* 3 is 9.

<div id="parenthisized-grammar" />

#### Grammar

\<parenthesized expression\> ::= \"(\" \<expression\> \")\"

### Operators

#### Behavioral Description

MongoSQL provides several operators for the built-in data types.
Specifically, it supports all operators defined in the SQL-92 spec with
only one exception. There is no support for the INTERVAL data type in
MongoSQL, so arithmetic operators involving BSON_DATE and INTERVAL
values are unsupported. See the [Data Types](#data-types)
section for more details about supported types.

In addition to the SQL-92 operators, MongoSQL also has array and
document operators. See the [Document and Field-Access
Expressions](#document-and-field-access-expressions)
section for details about the document field access operators, and see
the [Array, Indexing, and Slicing Expressions](#array-indexing-and-slicing-expressions)
section for details about the array index access operator.

Note, the SQL-92 spec describes operators as monadic, dyadic, or n-adic,
corresponding to operators having one operand, two operands, or a
variable number of operands, respectively ([SQL-92
3.1](https://www.contrib.andrew.cmu.edu/~shadow/sql/sql1992.txt)).
This document uses the adjectives unary, binary, and n-ary to describe
operators.

Operators can be divided into several different groups. There are string
operators, arithmetic operators, comparison operators, boolean
operators, control-flow operators, and type operators. The following
sections explain the behavior of each type of operator.

#### Semantics of String Operators

String operators are those which operate on strings. The operands of
string operators must statically have type NULL or STRING, and may
evaluate to MISSING. If an operand evaluates to NULL or MISSING, the
result of the operation is NULL.

The binary string operator \|\| specifies string concatenation. It
returns the string made by joining its string operands in the order
given. The result of string concatenation either has type NULL or
STRING.

The 3-ary string operator LIKE determines whether a string matches a
pattern. The third argument is optional. When provided, it specifies an
escape character used in the pattern. If the third argument is not
specified, the default escape character, \'\\\', is used. In the pattern, an unescaped
underscore character \'_\' represents any single character and an
unescaped percent character \'%\' represents any number of characters,
even zero characters.

To paraphrase SQL-92: \"If there is not a partitioning of the pattern
into substrings such that each substring has length 1 or 2, no substring
of length 1 is the escape character, and each substring of length 2 is
the escape character followed by either the escape character, an \'_\',
or a \'%\',\" then the result is NULL ([SQL-92
8.5](https://www.contrib.andrew.cmu.edu/~shadow/sql/sql1992.txt)).

The first two operands do not need to be string literals, they can be
any expressions that statically have type NULL or STRING. If provided,
the optional third argument must be a literal string consisting of
exactly one character, any other value produces a static error. The
result of LIKE either has type NULL or BOOL.

The inverse, NOT LIKE, is syntactic sugar for the negation of the result
of LIKE. As in,

e1 NOT LIKE e2

is rewritten as:

NOT (e1 LIKE e2)

#### Semantics of Arithmetic Operators

Arithmetic operators are those which operate on numeric data types. The
operands of arithmetic operations must statically have type NULL or a
numeric type --- INT, LONG, DOUBLE, or DECIMAL --- and may evaluate to
MISSING. If an operand evaluates to NULL or MISSING, the result of the
operation is NULL. There are unary and binary arithmetic operators,
described below.

The unary arithmetic operators + and - specify unary addition and unary
subtraction, respectively. These unary operators are used to specify the
sign of their operands. Unary addition does not change its operand;
unary subtraction reverses the sign of its operand. The result of a
unary arithmetic operation has the same type as its operand.

The binary arithmetic operators +, -, \*, and / specify addition,
subtraction, multiplication, and division, respectively. If the value of
a divisor is zero, then the result of division is NULL. The operands of
binary arithmetic operations do not need to have the same type; any
combination of valid operand types (as declared at the beginning of this
section) is allowed. When both operand types are numeric, the result of
a binary arithmetic operation has a type according to the following
table:

|**Type of Operand 1** | **Type of Operand 2** | **Result**
|-----------------------|-----------------------|------------
|INT                    |INT                    |INT
|INT or LONG            |LONG                   |LONG
|Any numeric non-DECIMAL|DOUBLE                 |DOUBLE
|Any numeric            |DECIMAL                |DECIMAL

For division operations including only INTs and/or LONGs, the result may
not be exactly an INT or LONG and will need to be truncated or rounded.
For example, 5 / 2 is truly 2.5, however the result must be an INT since
the operands are both INTs. The choice of whether to round or truncate
is implementation-defined.

Note that arithmetic operations that result in overflow or underflow
have undefined behavior. For example, if 1 (an INT) is added to the
maximum INT value, the result exceeds the bounds of the INT type and
therefore cannot be represented as an INT. MongoSQL does not specify
behavior for such cases.

<div id="comparison-operator-semantics"/>

#### Semantics of Comparison Operators

Comparison operators are those which compare values. The operands of
comparison operations must statically have comparable types. In most
cases, that means the operands must have the same type (or NULL). The
exceptions to this are numeric types. Any combination of numeric types
can appear as operands for comparison operators. If an operand evaluates
to NULL or MISSING, the result of the operation is NULL. Note, this
deviates from MongoDB\'s comparison semantics.

The binary comparison operators \<, \<=, \<\>, =, \>, and \>= specify
less than, less than or equals, not equals, equals, greater than, and
greater than or equals, respectively. MongoSQL does not support
comparison operations on structured data (documents and arrays).
Booleans are compared such that FALSE is less than TRUE. Numbers are
compared with respect to their algebraic values. Strings are compared
lexicographically. Datetimes are compared as expected. The result of a
binary comparison operation has either type NULL or BOOL.

The != operator is a non-standard operator that specifies not equals.
Despite being non-standard, it is supported by many competitors and is
likely expected by many users. MongoSQL supports it as well, though it
is rewritten to the standard \<\> operator:\
a != b</br>
is rewritten as:</br>
a \<\> b

The 3-ary comparison operator BETWEEN specifies a range comparison.
Logically, the expression x BETWEEN y AND z is equivalent to x \>= y AND
x \<= z, though this specification [does not require](#rejected-alternativescompetitive-analysis-10)
that it be rewritten as such. The same type constraints and comparison
behaviors described above for the binary comparison operators apply to
BETWEEN. The result of a BETWEEN operation has either type NULL or BOOL.

The inverse, NOT BETWEEN, is syntactic sugar for the negation of the
result of BETWEEN. As in,

e1 NOT BETWEEN e2 AND e3

is rewritten as:

NOT (e1 BETWEEN e2 AND e3)

#### Semantics of Boolean Operators

Boolean operators are those which operate on boolean data types. The
operands of boolean operations must statically have type NULL or
BOOLEAN, and may evaluate to MISSING. If an operand evaluates to NULL or
MISSING, the result of the operation is NULL. There are unary and binary
boolean operators, described below.

The semantics of the unary boolean operator NOT are described by the
truth table below.

|**a**| **NOT a**|
|-|-|
|TRUE| FALSE|
|FALSE |TRUE|
|NULL or MISSING |NULL|

The semantics of the binary boolean operators OR and AND are described
by the truth table below.

|**a** |**b** |**a AND b** |**a OR b**|
|-|-|-|-|
|TRUE |TRUE |TRUE |TRUE|
|TRUE |FALSE |FALSE |TRUE|
|TRUE |NULL or MISSING |NULL |TRUE|
|FALSE |TRUE |FALSE |TRUE|
|FALSE |FALSE |FALSE |FALSE|
|FALSE |NULL or MISSING |FALSE |NULL|
|NULL or MISSING |TRUE |NULL |TRUE|
|NULL or MISSING |FALSE |FALSE |NULL|
|NULL or MISSING |NULL or MISSING |NULL |NULL|

---

#### Semantics of Control-Flow Operators

MongoSQL supports the CASE operator for control-flow. In this context,
\"control-flow\" refers to conditionally producing values based on
criteria argued to the operator. Concretely, a CASE expression consists
of one or more \"WHEN clauses,\" which each specify a boolean condition
and a result, and an \"ELSE clause,\" which specifies a default result
if none of the conditions are TRUE. The result of a CASE expression is
the result of the first (leftmost) WHEN clause whose condition is TRUE,
or the default result specified by the ELSE clause if none of the
conditions are TRUE. The type of the result of a CASE expression is the
union of types from the WHEN and ELSE clauses.

SQL-92 specifies two forms of the CASE expression, _simple_ and
_searched_.

A simple CASE expression has the following form:

CASE co WHEN wo<sub>1</sub> THEN r<sub>1</sub> WHEN wo<sub>2</sub> THEN r<sub>2</sub> \... ELSE r<sub>d</sub> END

In this form, the first expression, co, is the \"case operand,\" the
WHEN expressions, wo<sub>i</sub>, are the \"when operands,\" and the THEN and
ELSE expressions, r<sub>i</sub>, are the potential result values (\"d\" stands
for \"default\"). The result of a simple CASE expression is the r<sub>i</sub>
corresponding to the first (leftmost) wo<sub>i</sub> for which co = wo<sub>i</sub>
evaluates to TRUE, or r<sub>d</sub> if none of the comparisons evaluate to TRUE.
This is equivalent to a searched CASE expression where each condition is
co = wo<sub>i</sub>, though this specification [does not
require](#rejected-alternativescompetitive-analysis-10)
that simple CASE expressions be rewritten as searched CASE expressions.
Since the case operand and when operands are compared using the equals
operator, they must follow the type constraint rules described in the
[Comparison Operators](#semantics-of-comparison-operators)
section.

A searched CASE expression has the following form:\
CASE WHEN c<sub>1</sub> THEN r<sub>1</sub> WHEN c<sub>2</sub> THEN r<sub>2</sub> \... ELSE r<sub>d</sub> END

In this form, the WHEN expressions, c<sub>i</sub>, are the \"search conditions,\"
and the THEN and ELSE expressions, r<sub>i</sub>, are the potential result values
(\"d\" stands for \"default\"). The result of a searched CASE expression
is the r<sub>i</sub> corresponding to the first (leftmost) c<sub>i</sub> that evaluates to
TRUE, or r<sub>d</sub> if none of the conditions evaluate to TRUE. The search
conditions must statically have type NULL or BOOLEAN, and may evaluate
to MISSING. Again, note that if a condition evaluates to TRUE, its
corresponding result is the result of the expression. Therefore, if a
condition evaluates to NULL or MISSING, nothing special happens; the
result of the expression is either the next result corresponding to a
TRUE condition or the default result if none of the following conditions
evaluate to TRUE.

For either form of the CASE expression, if an ELSE clause is not
provided, then ELSE NULL is implicit. As in,

CASE o WHEN e THEN r END</br>
is rewritten as:</br>
CASE o WHEN e THEN r ELSE NULL END</br>

and

CASE WHEN e THEN r END</br>
is rewritten as:</br>
CASE WHEN e THEN r ELSE NULL END

#### Semantics of Type Operators

MongoSQL provides the binary IS operator to check the type of an
expression. The left operand can be any expression and the right operand
can be any MongoSQL type name or the keyword MISSING. The operator
returns TRUE if the argued expression evaluates to a value of the argued
type, and FALSE if it evaluates to a value of a different type. If the
left operand evaluates to MISSING, the operator returns TRUE if the
right operand is the keyword MISSING or NULL, and FALSE otherwise. The
result of the IS operator therefore always has type BOOL.

The SQL-92 \"null predicate\", \<expr\> IS NOT? NULL, returns TRUE if
the expression is a NULL _value_ and FALSE otherwise. In MongoSQL, NULL
is a type with a single value (also called NULL), and the value MISSING
is more analogous to SQL\'s NULL. To support the SQL-92 \"null
predicate\", MongoSQL has the following semantics for IS NULL and IS
MISSING.

|**A** |**A IS NULL** |**A IS MISSING**|
|-|-|-|
|NULL |TRUE |FALSE|
|MISSING |TRUE |TRUE|

Therefore, in MongoSQL, the operation \<expr\> IS NOT? NULL is not
strictly the type-checking operation described at the beginning of this
section; it is more like the value-checking operation defined by SQL-92.

The inverse, IS NOT, is syntactic sugar for the negation of the result
of IS. As in,

e IS NOT t

is rewritten as:

NOT (e IS t)

<div id="operators-grammar" />

#### Grammar

\<unary operator expression\> ::= \<unary operator\> [\<expression\>](#expressions)

\<unary operator\> ::= + \| - \| NOT

\<binary operator expression\> ::= [\<expression\>](#expressions) \<binary operator\> [\<expression\>](#expressions)

\<binary operator\> ::= -</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \*</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| /</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| +</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \|\|</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| AND</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| OR</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \<comparison operator\></br>

\<comparison operator\> ::= \<</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<=</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \<\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| !=</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| =</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \>=</br>

\<is operator expression\> ::= [\<expression\>](#expressions) IS NOT?  ([\<type\>](#data-types) \| MISSING)

\<like operator expression\> ::= [\<expression\>](#expressions) NOT? LIKE [\<expression\>](#expressions) (ESCAPE [\<string literal\>](#literals))?

\<between operator expression\> ::= [\<expression\>](#expressions) NOT? BETWEEN [\<expression\>](#expressions) AND [\<expression\>](#expressions)

\<case operator expression\> ::= CASE \<when clause\>+ \<else clause\>? END</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| CASE [\<expression\>](#expressions) \<when clause\>+ \<else clause\>? END

\<when clause\> ::= WHEN [\<expression\>](#expressions) THEN [\<expression\>](#expressions)

\<else clause\> ::= ELSE [\<expression\>](#expressions)

#### Examples

See query tests, rewrite tests, and type constraint tests
[here](https://github.com/10gen/mongosql-rs/tree/master/tests/spec_tests).

#### Open Questions

- What are the semantics for comparing different numeric types? Is
  NumberDouble(20) the same as NumberDecimal(20)?

  - **Proposed Answer:** Numeric values will be compared according
    to their true mathematical value.

- Should we support the SQL-92 \<null predicate\> operator \"\<expr\>
  IS \[NOT\] NULL\"?

  - **Proposed Answer:** we can simply support IS NULL as a
    type-checking operator, which would work since NULL is a type
    with a single value (also called NULL).

#### Rejected Alternatives/Competitive Analysis

MongoSQL does not support the SQL-92 \<boolean test\> operator IS.
Specifically, MongoSQL does not support \<expr\> IS NOT? \<bool value\>.
This operator is not strictly required for SQL compatibility. One
benefit of not supporting it is that IS can unambiguously be used as the
type-check operator.

For now, we rejected comparisons for structured data (documents and
arrays). There are still several questions to work through for
structured data comparisons (how to handle NULL elements, for example).
Structured data types are not part of SQL-92 and therefore we do not
need to address this at this time. We are not excluded from supporting
this in the future.

We also rejected polymorphic comparisons. Comparisons are not required
to be polymorphic for SQL compatibility and being restrictive now does
not preclude us from allowing this in the future. In addition to those
reasons, there are other considerations that may impact this decision.
For example, if MongoSQL needs to support TIME and DATE types, then
cross-type comparisons will not easily be able to follow BSON type
ordering. Also, if MongoDB ever introduces type collations then
prescribing a type ordering in the spec may prevent us from supporting
such a feature in the future. Most competitor documentation does not
elaborate on type constraints on comparison operators:

- PrestoDB, N1QL, SQL++, and PartiQL do not specify type constraints

- MySQL allows polymorphic comparisons via implicit conversions

- PostgreSQL does not allow polymorphic comparisons

- [Rockset](https://docs.rockset.com/conditional-expressions/#comparison-operators)
  allows polymorphic comparisons but returns null if the two
  operands have different types.

This specification does not require that implementations rewrite BETWEEN
operations as ANDs of less than and greater than operations. This is
left up to implementations since prescribing the rewrite would
necessitate that the first operand be copied multiple times (one for
each comparison) which could have performance implications which this
spec aims to avoid.

Similarly, it does not require that implementations rewrite simple CASE
expressions to searched CASE expressions. This is also left up to
implementations since prescribing the rewrite would necessitate that the
first operand be copied multiple times (one for each when clause).

### Scalar Functions

#### Behavioral Description

MongoSQL requires most scalar functions defined in the SQL-92 spec, with
a few exceptions [highlighted
below](#rejected-alternativescompetitive-analysis-11). In
addition to the SQL-92 functions, MongoSQL also specifies a general
syntactic construct that implementations can use to introduce additional
scalar functions. For example, not all scalar functions from MongoDB
aggregation are required by this spec, but any could be included by an
implementation.

The required scalar functions are described below. They are separated
into groups based on their behavior, input types, or output types. The
descriptions cover the basics of what the functions do as well as any
static type constraints on their arguments (where relevant); see the
[spec tests](#examples-11) for full behavior details.

##### Conditional Scalar Functions

Conditional scalar functions are those which specify a conditional
value. SQL-92 requires the NULLIF and COALESCE conditional functions.
These functions are equivalent to CASE operations, though this
specification [does not
require](#rejected-alternativescompetitive-analysis-11)
that they be rewritten as such.

The NULLIF(v1, v2) scalar function returns NULL if v1 = v2 is TRUE, and
otherwise returns v1. Since the arguments are compared for equality,
they must be statically comparable according to the rules described in
the [Comparison
Operators](#semantics-of-comparison-operators) section.
NULLIF(v1, v2) is equivalent to the following searched CASE operation:

CASE WHEN v1 = v2 THEN NULL ELSE v1 END

The COALESCE(v1, v2, \..., vn) scalar function returns the first
non-NULL argument, or NULL if there are no non-NULL arguments.

COALESCE(v1, v2) is equivalent to the following searched CASE operation:

CASE WHEN v1 IS NOT NULL THEN v1 ELSE v2 END

and COALESCE(v1, v2, \..., vn) is equivalent to:

CASE WHEN v1 IS NOT NULL THEN v1 ELSE COALESCE(v2, \..., vn) END

##### Type Conversion Scalar Function

The type conversion scalar function CAST converts an expression to a
specified type. See the [Type Conversions](#type-conversions) section for details about
the CAST scalar function.

##### Array Scalar Functions

Array scalar functions are those which operate on arrays. MongoSQL
specifies two such scalar functions, SLICE and SIZE. See the
[Array, Indexing, and Slicing Expressions](#slicing) section for
details about the SLICE scalar function.

The SIZE(array) scalar function counts and returns the total number of
items in an array. The semantics of the SIZE function mostly match those
of the aggregation operator
[\$size](https://docs.mongodb.com/manual/reference/operator/aggregation/size/).
Notably, the type constraints are more relaxed than \$size. The argument
must statically have type ARRAY or NULL, and may be missing. If the
argument is NULL or MISSING, the result is NULL.

##### Numeric Value Scalar Functions

Numeric value scalar functions are those which return numeric values.
SQL-92 requires the POSITION, CHAR_LENGTH, OCTET_LENGTH, BIT_LENGTH, and
EXTRACT numeric value functions.

The POSITION(substring IN string) scalar function returns the position
of the first occurrence of substring in the string, or -1 if it does not
occur. MongoSQL uses 0-indexing, so the first character is at position
0, the second at position 1, and so on. The arguments must statically
have type STRING or NULL, and may be missing. If either argument is NULL
or MISSING, the result is NULL.

The function CHAR_LENGTH may also be written as CHARACTER_LENGTH.
MongoSQL will rewrite the latter into the former, so that CHAR_LENGTH is
the canonical form of the function name.

The length functions CHAR_LENGTH(string), OCTET_LENGTH(string), and
BIT_LENGTH(string) return the length of string arguments in terms of
characters, bytes, and bits, respectively. The semantics of
CHARACTER_LENGTH match those of the aggregation operator
[\$strLenCP](https://docs.mongodb.com/manual/reference/operator/aggregation/strLenCP/);
the semantics of OCTET_LENGTH match those of the aggregation operator
[\$strLenBytes](https://docs.mongodb.com/manual/reference/operator/aggregation/strLenBytes/);
the semantics of BIT_LENGTH match those of
[\$strLenBytes](https://docs.mongodb.com/manual/reference/operator/aggregation/strLenBytes/),
as well, except the result is multiplied by 8. For each, the argument
must statically have type STRING or NULL, and may be missing. If the
argument is NULL or MISSING, the result is NULL.

The EXTRACT(field FROM source) scalar function extracts a component of a
datetime value. The source argument must statically have type TIMESTAMP
or NULL, and may be missing. If it is NULL or MISSING, the result is
NULL.

##### Addititional Numeric Scalar Functions

The following list of functions operate on numeric values and return numeric values.
All arguments to these functions must statically have type `NULL` or a numeric type
— `INT`, `LONG`, `DOUBLE`, or `DECIMAL` — and may evaluate to `MISSING`.  If an
argument evaluates to `NULL` or `MISSING`, the result of the operation is `NULL`.
If an argument resolves to `NaN`, the result is `NaN`. Some functions have
additional special behaviors detailed below.

The ABS ( `number` ) scalar function returns the absolute value of the given `number`.
The result has the same type as its operand.  If the argument is `Infinity` or
-`Infinity`, the result is `Infinity`.

The CEIL ( `number` ) scalar function returns a number to the nearest whole number of equal
or greater value.  The result has the same type as its operand. If the argument evaluates
to negative or positive `Infinity`, the result is negative or positive `Infinity` respectively.

The COS ( `number` ) scalar function returns the cosine of an angle. The argument is specified in
radians. If the argument is of type `DECIMAL` the return type is `DECIMAL`. All other numeric argument types
return `DOUBLE`. If the argument evaluates to negative or positive `Infinity`, the result of the operation
is `NULL`.

The DEGREES ( `number` ) scalar function converts a given `number` in radians to degrees.
If the argument is of type `DECIMAL` the return type is `DECIMAL`. All other numeric argument types
return `DOUBLE`. If the argument evaluates to negative or positive `Infinity`, the result
is negative or positive `Infinity` respectively.

The FLOOR ( `number` ) scalar function returns a number to the nearest whole number of equal
or lesser value.  The result has the same type as its operand. If the argument evaluates to
negative or positive `Infinity`, the result is negative or positive `Infinity` respectively.

The LOG ( `number`, `base` ) scalar function returns the logarithm of a number for the given base.
`Number` must be a non-negative number, `base` must be a positive number greater than 1.
Arguments outside of the supported ranges will return NULL.  If the `number` or `base` is of
type `DECIMAL` the return type is `DECIMAL`. All other numeric types return `DOUBLE`. If
`number` evaluates to Infinity, the result is `Infinity`.  If base evaluates to Infinity,
the result is 0.

The MOD ( `number`, `divisor` ) scalar function divides number by divisor and returns the
remainder.  The result has the same [type as its operands](#semantics-of-arithmetic-operators)
except when it cannot be represented accurately in that type.  In these cases:

- INT will be converted to a LONG if the result is representable as a LONG.
- INT will be converted to a DOUBLE if the result is not representable as a LONG.
- LONG will be converted to DOUBLE if the result is not representable as a LONG.
- A DECIMAL divisor will have a result of type DECIMAL

If `number` evaluates to negative or positive Infinity, the result is NaN.  If `divisor`
evaluates to negative or positive `Infinity`, the result is `number`.  A `divisor` of 0 will result in NULL.

The POW ( `number`, `exponent` ) scalar function returns the number raised to the specified
exponent.  The result has the same [type as its operands](#semantics-of-arithmetic-operators)
except when it cannot be represented accurately in that type.  In these cases:

- INT will be converted to a LONG if the result is representable as a LONG.
- INT will be converted to a DOUBLE if the result is not representable as a LONG.
- LONG will be converted to DOUBLE if the result is not representable as a LONG.

If either number or power arguments evaluate to Infinity, the result is Infinity.  If power evaluates to -Infinity, the result is 0.  If `number` is -Infinity and `power` is even, the result is Infinity otherwise if power is odd the result is -Infinity.

The RADIANS ( `number` ) scalar function returns the given number converted from degrees to radians.
If the argument is of type `DECIMAL` the return type is `DECIMAL`. All other numeric argument types
return `DOUBLE`.
If the argument evaluates to negative or positive `Infinity`, the result is negative or
positive `Infinity` respectively.

The ROUND ( `number`, `decimals` ) scalar function rounds `number` to a specified number of digits.
The `decimals` argument specifies how many decimal points of precision to include in the final result.
The `decimals` argument must be an integer between -20 and 100.  Arguments outside
of the supported ranges will return NULL.  The result has the same type as the first operand.
If the first argument resolves to negative or positive infinity, the result is negative or positive infinity respectively.

The SIN( `expression` ) scalar function returns the sine of “expression”, where “expression” is an
angle expressed in radians. The argument must statically have type `DOUBLE, INT, LONG`,
`DECIMAL` or `NULL`, and may evaluate to MISSING. If the argument is `NULL` or `MISSING`,
the result is `NULL`. If the argument evaluates to negative or positive `Infinity`, the result
of the operation is `NULL`.

The SQRT( `number` ) scalar function returns the square root of a positive number.
If the argument is negative, the operation will return NULL.  If the argument evaluates to
`Infinity`, the result is `Infinity`.

The TAN( `number` ) scalar function returns the tangent of an angle, specified in radians.
If the argument is of type `DECIMAL` the return type is `DECIMAL`. All other numeric argument types
return `DOUBLE`. If the argument evaluates to negative or positive `Infinity`, the result of the operation is `NULL`.

##### String Value Scalar Functions

String value scalar functions are those which return string values.
SQL-92 requires the SUBSTRING, UPPER, LOWER, CONVERT, TRANSLATE, and
TRIM string value functions; however, MongoSQL [does not
require](#rejected-alternativescompetitive-analysis-11)
CONVERT or TRANSLATE.

The SUBSTRING(string FROM start FOR length) scalar function returns a
substring of the first argument starting from start and extending either
the specified length, if provided, or to the end of the string, if
length is not provided or if it is negative. SQL-92 specifies a runtime
error when the length is negative; MongoSQL deviates from this to be
consistent with [MongoDB aggregation](https://docs.mongodb.com/manual/reference/operator/aggregation/substr/#behavior)
by returning the substring extending to the end of the string. Also,
note that MongoSQL uses 0-indexing, so start should be 0 to start at the
first character, 1 to start at the second, and so on. The unit of the
start and length arguments is codepoint, not byte. This allows SUBSTRING
to work in a reasonable way for strings containing multibyte characters.
The string argument must statically have type STRING or NULL, and may be
missing. The start and length arguments must statically have type INT or
NULL, and may be missing. If any argument is NULL or MISSING, the result
is NULL. Note that the characters are 0-indexed, for consistency with
MongoDB.

The UPPER(string) and LOWER(string) scalar functions, collectively known
as \"fold\" functions, change the casing of their arguments to be all
uppercase or all lowercase, respectively. Similar to the aggregation
operators
[\$toLower](https://docs.mongodb.com/manual/reference/operator/aggregation/toLower/)
and
[\$toUpper](https://docs.mongodb.com/manual/reference/operator/aggregation/toUpper/),
these functions only have well-defined behavior for strings of ASCII
characters. For each, the argument must statically have type STRING or
NULL, and may be MISSING. If the argument is NULL or MISSING, the result
is NULL.

The TRIM(spec charset FROM string) scalar function removes leading
and/or trailing spaces, or an optionally specified charset, from the
argued string. The first two arguments are optional; by default, they
are BOTH and the space character, respectively. The latter two arguments
must statically have type STRING or NULL, and may be missing. If either
is NULL or MISSING, the result is NULL.

A charset is a string of unordered characters that will all be trimmed from 
the beginning and/or end of `string` as defined by `spec`.

As noted, the first two arguments for TRIM are optional and have the
default values BOTH and the space character, respectively. Invocations
with one or both arguments omitted will be rewritten to include the
defaults. Specifically:

TRIM(str)</br>
will be rewritten as:</br>
TRIM(BOTH \' \' FROM str)

and

TRIM(charset FROM str)</br>
will be rewritten as:</br>
TRIM(BOTH charset FROM str)

and

TRIM(spec FROM str)</br>
will be rewritten as:</br>
TRIM(spec \' \' FROM str)

##### Additional String Functions

- SPLIT(`string`, `delimiter`, `token number`) => `string`
  - Returns a substring from a `string`, using a `delimiter` character to divide the string into a sequence of tokens.

  The string is interpreted as an alternating sequence of delimiters and tokens.
  So for the string "abc-defgh-i-jkl", where the delimiter character is ‘-‘, the tokens
  are abc, defgh, i, and jlk. Think of these as tokens 1 through 4. SPLIT returns the token
  corresponding to the `token number`. When the token number is positive, tokens are counted
  starting from the left end of the string; when the token number is negative, tokens are
  counted starting from the right.

  If `token number` is greater than the number of tokens, the result is an empty string.

  The `string` argument must statically have type STRING or NULL, and
  may evaluate to MISSING. `token number` must statically have type INT or NULL,
  and may evaluate to MISSING. If any argument is NULL or MISSING, or a `delimiter` evaluates
  to an empty string, the result is NULL.

- REPLACE(`string`, `redex`, `replacement`) => `string`
  - Searches `string` for occurrences of `redex`, and replaces those occurrences with `replacement`.
    This is part of the [ODBC standard](https://learn.microsoft.com/en-us/sql/odbc/reference/appendixes/string-functions)

  As with all String functions, the `redex` and `replacement` are case sensitive.

##### Datetime Value Scalar Functions

Datetime value scalar functions are those which return datetime values.
SQL-92 requires the CURRENT_DATE, CURRENT_TIME, and CURRENT_TIMESTAMP
datetime value functions; however, MongoSQL [does not require](#rejected-alternativescompetitive-analysis-11)
CURRENT_DATE or CURRENT_TIME.

The CURRENT_TIMESTAMP scalar function returns the current timestamp,
\"with the time zone displacement equal to the current time zone
displacement of the SQL-session\" ([SQL-92 6.8](https://www.contrib.andrew.cmu.edu/~shadow/sql/sql1992.txt)).

The remaining semantics for MongoSQL's CURRENT_TIMESTAMP are consistent
with the behavior described by section 6.8 of the SQL-92 Specification,
which we will quote here instead of attempting to rephrase:

>_2) If specified, \<timestamp precision\> \... determines the
precision of the \... timestamp value returned._ (Default value is 6
if none is provided).
>_3) If an SQL-statement generally contains more than one reference to
one or more \<datetime value function\>s, then all such references are
effectively evaluated simultaneously. The time of evaluation of the
\<datetime value function\> during the execution of the SQL-statement
is implementation-dependent._

##### Additional Date Functions

- DATEADD(`date_part`, `interval`, `date`) => `datetime`
  - Returns the specified `date` with the specified number of `interval` added to the specified `date_part` of that `date`.
- DATEDIFF(`date_part`, `date1`, `date2`, \[`start_of_week`\]) => `int`
  - Returns the difference between `date1` and `date2` expressed in units of `date_part`.
- DATETRUNC(`date_part`, `date`, \[`start_of_week`]) => `datetime`
  - Truncates the specified `date` to the accuracy specified by `date_part`, and returns a new `date`.

The `date` argument must be a MongoSQL `datetime` value and `date_part`
must be a valid token. The default `start_of_week` is _"sunday"_ (case insensitive),
for DATEDIFF and DATETRUNC. If an argument evaluates to `NULL` or `MISSING`,
the result of the operation is `NULL`.

#### Examples

See query tests, rewrite tests, and type constraint tests
[here](https://github.com/10gen/mongosql-rs/tree/master/tests/spec_tests).

<div id="scalar-functions-grammar">

#### Grammar

\<scalar function expression\> ::= \<nullif function\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \<coalesce function\> \| \<size function\> \| \<position function\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \<character length function\> \| \<octet length function\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \<bit length function\> \| \<extract function\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \<substring function\> \| \<fold function\> \| \<trim function\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \<current date function\> \| \<current time function\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| \<current timestamp function\></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| [\<regular identifier\>](#identifiers) \"(\" [\<expression\>](#expressions)\* \")\"
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| < abs function ></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| < ceiling function > | < cos function > | < degrees function ></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| < floor function > | < log function > | < mod function ></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| < power function > | < radians function > | < round function ></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| < sin function > | < sqrt function > | < tan function ></br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| < regular identifier> "(" \<[expression](#expressions)\>* ")"

\<nullif function\> ::= NULLIF \"(\" [\<expression\>](#expressions) \",\" [\<expression\>](#expressions) \")\"

\<coalesce function\> ::= COALESCE \"(\" [\<expression\>](#expressions) (\",\" [\<expression\>](#expressions))\* \")\"

\<size function\> ::= SIZE \"(\" [\<expression\>](#expressions) \")\"

\<position function\> ::= POSITION \"(\" [\<expression\>](#expressions) IN [\<expression\>](#expressions) \")\"

\<character length function\> ::= (CHAR_LENGTH \| CHARACTER_LENGTH) \"(\" [\<expression\>](#expressions) \")\"

\<octet length function\> ::= OCTET_LENGTH \"(\" [\<expression\>](#expressions) \")\"

\<bit length function\> ::= BIT_LENGTH \"(\" [\<expression\>](#expressions) \")\"

\<extract function\> ::= EXTRACT \"(\" \<extract field\> FROM [\<expression\>](#expressions) \")\"

\<extract field\> ::= TIMEZONE_HOUR \| TIMEZONE_MINUTE</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| YEAR \| MONTH \| DAY \| HOUR \| MINUTE \| SECOND

\<substring function\> ::= SUBSTRING \"(\" [\<expression\>](#expressions) FROM [\<expression\>](#expressions) (FOR [\<expression\>](#expressions))? \")\"</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| SUBSTRING \"(\" [\<expression\>](#expressions) \",\" [\<expression\>](#expressions) (\",\" [\<expression\>](#expressions))? \")\"

\<fold function\> ::= (UPPER \| LOWER) \"(\" [\<expression\>](#expressions) \")\"

\<trim function\> ::= TRIM \"(\" \<trim options\>?  [\<expression\>](#expressions) \")\"

\<trim options\> ::= \<trim specification\>?  [\<expression\>](#expressions) FROM<sup>[7](#7)</sup>

\<trim specification\> ::= LEADING \| TRAILING \| BOTH

\<current timestamp function\> ::= CURRENT_TIMESTAMP (\"(\" [\<expression\>](#expressions) \")\")?

\<split function\> ::= SPLIT "(" \<[expression](#expressions)\> "," \<[expression](#expressions)\> "," \<[expression](#expressions)\> ")"

< abs function > ::= ABS "(" \<[expression](#expressions)\> ")"

< ceiling function > ::= CEIL "(" \<[expression](#expressions)\> ")"

< cos function > ::= COS "(" \<[expression](#expressions)\> ")"

< degrees function > ::= DEGREES "(" \<[expression](#expressions)\> ")"

< floor function > ::= FLOOR "(" \<[expression](#expressions)\> ")"

< log function > ::= LOG "(" \<[expression](#expressions)\>, \<[expression](#expressions)\> ")"

< mod function > ::= MOD "(" \<[expression](#expressions)\>, \<[expression](#expressions)\> ")"

< power function > ::= POW "(" \<[expression](#expressions)\>, \<[expression](#expressions)\> ")"

< radians function > ::= RADIANS "(" \<[expression](#expressions)\> ")"

< round function > ::= ROUND "(" \<[expression](#expressions)\>, \<[expression](#expressions)\> ")"

< sin function > ::= SIN "(" \<[expression](#expressions)\> ")"

< sqrt function > ::= SQRT "(" \<[expression](#expressions)\> ")"

< tan function > ::= TAN "(" \<[expression](#expressions)\> ")"

\<date function> ::= \<dateadd\> | \<datediff\> | \<datetrunc\>

\<standard date part> ::= YEAR | MONTH | DAY | HOUR | MINUTE | SECOND | WEEK

\<date_part\> ::= \<standard date part\> | QUARTER

\<dateadd\> ::= DATEADD "(" \<date_part\>"," \<[expression](#expressions)\> "," \<[expression](#expressions)\> ")"

\<datediff\> ::= DATEDIFF "(" \<date_part\> "," \<[expression](#expressions)\> "," \<[expression](#expressions)\> ("," \<[expression](#expressions)\>)? ")"

\<datetrunc\> ::= DATETRUNC "(" \<date_part\> "," \<[expression](#expressions)\> ("," \<[expression](#expressions)\>)? ")"

> <sup id="7">7</sup> Note, the SQL-92 grammar actually specifies that TRIM(FROM
    \<string\>) is a syntactically valid invocation of the TRIM
    function, so we preserve that in MongoSQL grammar ([SQL-92
    6.7](https://www.contrib.andrew.cmu.edu/~shadow/sql/sql1992.txt)).

#### Open Questions

- Do we want to support the aggregation functions (AVG, MIN, MAX,
  COUNT, SUM) as scalar functions that operate on arrays?

  - **Proposed Answer:** This could be a nice feature to support
    eventually, but does not seem necessary at this time since it
    is outside the scope of SQL-92.

#### Rejected Alternatives/Competitive Analysis

This specification does not require that implementations rewrite NULLIF
or COALESCE functions as CASE operations. This is left up to
implementations since prescribing the rewrite would necessitate that one
or more operands be copied multiple times which could have performance
implications which this spec aims to avoid.

We rejected the SQL-92 specified string value functions CONVERT and
TRANSLATE for several reasons, including their conformance-level
restrictions in SQL-92 and competitors\' lack of support for them.

We also rejected the SQL-92 specified datetime value functions
CURRENT_DATE and CURRENT_TIME because MongoSQL does not support the DATE
or TIME types at this time.

### Subquery Expressions

#### Behavioral Description

A subquery is a SQL query within a query.

- A subquery may occur anywhere an expression can be used.

- When executing a subquery, we set ρ<sub>0</sub> for the subquery
  to the current _ρ_ for the outer scope in which the subquery is
  executed. This means values from the current executing row are
  available to the subquery expression.

There are three types of subqueries defined in SQL-92:

- Scalar subquery, which returns a result set with zero or one row and
  one column. This is the simplest form of a subquery, and can be
  used in most places a literal or single column value is valid

- Row subquery, which returns a result set with zero or one row and
  more than 1 column

- Table subquery, which returns 0 or more rows and one or more columns

MongoSQL only supports scalar subquery and table subquery. The reason
for not supporting the row subquery is described in the [below
section](#rejected-alternativescompetitive-analysis-12).
Besides, only traditional SQL style SELECT is allowed in subquery. Users
cannot write (SELECT VALUE ...) in a subquery.

A table subquery will only be used at the right hand side(rhs) of a
[multiple-row operator such as IN, ANY, SOME, ALL, EXISTS. Their
behavior is described in the
[section](#behavior-and-grammar-for-in-any-some-all-exists)
below.

The subquery is required to statically prove its degree is equal to the
expression on the other side of the operator. The degree of the subquery
should either be derived from the SELECT clause expression list or in
the subquery (SELECT \* FROM foo), such information needs to be derived
from foo's schema. Otherwise, the compilation will fail.

[All other subqueries where table subquery is not applicable are
required to be scalar subquery. For scalar subquery, static check will
be done to make sure the degree of the subquery is 1 and the cardinality
is no larger than 1. Otherwise, the compilation will fail.

A scalar subquery expression evaluates to the single value returned by
the contained query. If the result of the query expression is empty, the
subquery will return MISSING.

The current environment from the outer query is inherited by the
subquery. This is described in the [Scoping
section](#scoping-for-value-names). Thus we will be able to
support correlated queries (a subquery using values from an outer query)
such as:

SELECT employee_number, department</br>
FROM employees AS emp</br>
WHERE salary > (</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;SELECT AVG(salary)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;FROM employees</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;WHERE department = emp.department</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;);

SQL-92 clearly defines the resolution strategy for unqualified
identifiers in subquery. In MongoSQL, we need to handle queries without
a schema. We decided not to allow subquery reference an unqualified
identifier in the schema-less setting with the following two exceptions.
The reason is elaborated in the below
[section](#unqualified-identifier).

1. If an unqualified name is referring to an identifier in subquery's
local scope, it has to be statically proven the identifier exists
in subquery's data source.

For example:

SELECT \* FROM foo WHERE EXISTS(SELECT \* FROM \[{a: 1, b: 1}\] AS bar
WHERE bar.b = a)

is allowed because we can statically prove field "a" exists in
datasource bar and according to SQL standard, the unqualified identifier
should be resolved to the most local scope which contains it.

SELECT \* FROM foo where EXISTS(SELECT \* FROM bar WHERE bar.b = a)

will get a compiling error because we cannot statically prove field "a"
exists in bar.

2. If an unqualified name is referring to an identifier in subquery's
outer scope, it has to be statically proven the identifier doesn't
exist in subquery's local scope and it exists in subquery's outer
scope.

For example:

SELECT \* FROM \[{a: 1}\] AS foo WHERE EXISTS(SELECT \* FROM \[{b: 1}\]
AS bar WHERE bar.b = a)

is allowed because we can statically prove field "a" only exists in
datasource foo but not in bar.

SELECT \* FROM foo where EXISTS(SELECT \* FROM \[{b: 1}\] AS bar WHERE
bar.b = a)

or

SELECT \* FROM \[{a: 1}\] AS foo WHERE EXISTS(SELECT \* FROM bar AS bar
WHERE bar.b = a)

will get compiling errors because in the first query, we cannot
statically prove field "a" exists in foo. And in the second query, we
cannot statically prove field "a" doesn't exist in bar.

<div id="subquery-grammar" />

#### Grammar

\<subquery expression\> ::= \"(\" [\<select query\>](#select-clause) \")\"

#### Rejected Alternatives/Competitive Analysis

##### Row constructors

We chose not to support row constructors for now. Some databases
including SQL server don't support row constructor, so it's very likely
this feature is not required by BI tools. An additional consequence of
this decision is that we do not need to distinguish table subqueries and
row subqueries, since row subqueries are only different when compared to
row constructors.. It would be easy to add support for row constructors
and row subqueries in the future if necessary.

##### Unqualified identifier

We choose to only support unqualified identifiers in subquery in the
cases listed above, due to the ambiguity generated by the missing
qualifiers. For example, in following query and catalog:

SELECT \* FROM foo where EXISTS(SELECT \* FROM bar WHERE bar.b = a)

_ρ_<sub>c</sub> = {</br>
&nbsp;&nbsp;test: {</br>
&nbsp;&nbsp;&nbsp;&nbsp;foo: \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{ a: 1}</br>
&nbsp;&nbsp;&nbsp;&nbsp;\],</br>
&nbsp;&nbsp;&nbsp;&nbsp;bar: \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{ b: 1}</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{ a: 2, b: 2}</br>
&nbsp;&nbsp;&nbsp;&nbsp;\]</br>
&nbsp;&nbsp;}</br>
}

We can interpret this "a" in the subquery as either foo.a or bar.a by
evaluating either one of the records in bar. If we read record { b: 1}
first, according to SQL standard, the unqualified identifier "a" should
be resolved to foo.a. Otherwise, it should be resolved to bar.a. Such
undefined behavior cannot be accepted in MongoSQL design.

#### Behavior and Grammar for IN, ANY, SOME, ALL, EXISTS

An ANY predicate determines whether a target value matches any value in
the subquery for a provided comparison operator. The result is
equivalent to the logical disjunction of the comparisons between the
target value and each value in the subquery result set. If one of the
comparison predicates returns TRUE, the result of the ANY predicate is
TRUE. Else, if any of the comparison predicates returns NULL, the result
is NULL. In all other cases, the result is FALSE.

Worth mentioning, each comparison follows the same rule from the
[Comparison Operators section](#comparison-operator-semantics), which means the types on both sides of the quantifier should be statically proven to
be comparable. This applies to the other predicates in this section
which also involves comparison operators.

For example:\
1 = ANY(SELECT a FROM \[{a: 1}, {a: NULL}\] AS arr)

The predicate will be evaluated to

(1 == 1) OR (1 == NULL) -\> TRUE OR NULL -\> TRUE

1 = ANY(SELECT a FROM \[{a: 0}, {a: NULL}\] AS arr)

The predicate will be evaluated to

(1 == 0) OR (1 == NULL) -\> FALSE OR NULL -\> NULL

1 = ANY(SELECT a FROM \[{a: 0}, {a: 2}\] AS arr)

The predicate will be evaluated to

(1 == 0) OR (1 == 2) -\> FALSE OR FALSE -\> FALSE

The SOME quantifier is equivalent to ANY. They generate the same result
and can be used interchangeably.

An ALL predicate determines whether a target value matches all the
values in the subquery for a provided comparison operator. The result is
equivalent to logical conjunction of the comparisons between the target
value and each value in the subquery result set. Only if all the
comparison predicates return TRUE, the result of the ALL predicate is
TRUE. If any comparison predicate returns FALSE, the result is FALSE. In
all other cases, the result is NULL.

For example:\
1 = ALL(SELECT a FROM \[{a: 1}, {a: NULL}\] AS arr)

The predicate will be evaluated to

(1 == 1) AND (1 == NULL) -\> TRUE AND NULL -\> NULL

1 = ALL(SELECT a FROM \[{a: 0}, {a: NULL}\] AS arr)

The predicate will be evaluated to

(1 == 0) AND (1 == NULL) -\> FALSE AND NULL -\> FALSE

1 = ANY(SELECT a FROM \[{a: 1}, {a: 1}\] AS arr)

The predicate will be evaluated to

(1 == 1) AND (1 == 1) -\> TRUE AND TRUE -\> TRUE

An IN predicate determines whether a target value equals any value in a
subquery or list of expressions. When it\'s a table subquery on the
right hand side of the predicate, this statement will be rewritten with
ANY predicate and has the same behavior..

X IN Y

will be rewritten to

X = ANY(Y)

In the case there is a comma separated expression list on the right hand
side of the predicate, the expression list will be first
[rewritten](#rewrites-2) with subquery and the same
rewrite as above will be used.

An EXISTS predicate evaluates to TRUE providing the subquery result set
is not empty, otherwise it evaluates to FALSE. This predicate does not
evaluate to NULL.

<div id="in-any-all-exists-grammar"/>

##### Grammar

\<subquery\> ::= \"(\" [\<select query\>](#select-clause)<sup>[8](#8)</sup> \")\"

\<in predicate\> ::= [\<expression\>](#expressions) NOT?  IN (\<subquery\> \| \"(\"\<in value list\>\"）\")

\<in value list\> ::= [\<expression\>](#expressions) (, [\<expression\>](#expressions))?

\<any predicate\> ::= [\<expression\>](#expressions) [\<comparison operator\>](#semantics-of-comparison-operators) ANY \<subquery\>

\<all predicate\> ::= [\<expression\>](#expressions) [\<comparison operator\>](#semantics-of-comparison-operators) ALL \<subquery\>

\<exists predicate\> ::= EXISTS \<subquery\>

> <sup id="8">8</sup> Subqueries using SELECT VALUE will parse, but will not pass
semantic validation at this time.

##### Rewrites

X IN \<subquery\></br>
will be rewritten to</br>
X = ANY \<subquery\>

X IN (A, B, C)</br>
will be rewritten to</br>
x = ANY(SELECT \_1 FROM \[{\_1: A}, {\_1: B}, {\_1: C}\])

X NOT IN \<subquery\></br>
will be rewritten to</br>
X \<\> ALL(\<subquery\>)

X NOT IN (A, B, C)</br>
will be rewritten to</br>
x \<\> ALL(SELECT \_1 FROM \[{\_1: A}, {\_1: B}, {\_1: C}\])

X = SOME(Y)</br>
will be rewritten to</br>
X = ANY(Y)

### Document and Field-Access Expressions

#### Behavioral Description

Documents can be represented with a syntax similar to JSON objects. Keys
must be strings and values can have any of the [supported types](#data-types).
To access document fields, MongoSQL supports two options: \"dot\" notation and \"bracket\" notation.

Dot notation is similar to field access in MongoDB aggregation. For
example, if a document doc contains a field f, then the expression doc.f
is used to access the value of that field. This type of expression is a
compound identifier, discussed in the [Identifiers](#identifiers) section.

Bracket notation uses square brackets, \[ and \], around a field name to
access the field with that name. For example, consider the same document
described before: doc\[\"f\"\] is used to access the value of that
field.

Document field access \<expr\>\[\<subfield\>\] has the following
behavior:

- \<expr\> must statically have type DOCUMENT or NULL, and may be
  missing.

  - If it is NULL or missing, the field access expression evaluates
    to NULL.

  - If it is DOCUMENT, the field access expression evaluates to the
    value of the subfield or MISSING if the subfield does not
    exist.

- \<subfield\> must statically have type STRING. Users can wrap the
  argument in CAST or use a type assertion to ensure this. See the
  [Type Conversions](#type-conversions) section for
  more details on casting.

<div id="doc-field-access-grammar" />

#### Grammar

\<document expression\> ::= {} \| { \<key-value pair\> (, \<key-value pair)\* }

\<key-value pair\><sup>[9](#9)</sup> ::= [\<string literal\>](#literals) \":\" [\<expression\>](#expressions)

\<document field access expression\> ::= [\<compound identifier\>](#identifiers) \| [\<expression\>](#expressions) \".\" [\<compound identifier\>](#identifiers)</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| [\<expression\>](#expressions) \"\[\" [\<expression\>](#expressions) \"\]\"

> <sup id="9">9</sup> Note: \<key-value pair\> is not actually independently part of the
grammar; it is displayed as a separate entity here for ease of
reading.

#### Examples

See spec tests [here](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/document.yml).

#### Rejected Alternatives/Competitive Analysis

MongoDB aggregation does not natively support bracket notation for field
access. We considered not supporting this syntax in MongoSQL but
ultimately decided to include it because 1) it is necessary to support
computed field names, and 2) there are viable options that will allow it
to be supported at implementation time. The competitor space is mixed on
this syntax:

- PartiQL and Presto support it

- N1QL and SQL++ do not

### Array, Indexing, and Slicing Expressions

#### Behavioral Description

Arrays are ordered lists of values. Elements can have any of the
[supported types](#data-types) and do not need to all be
the same type. To access array elements, MongoSQL supports the common
\"bracket\" notation \<array\>\[\<index\>\] and uses zero-indexing. For
example, if an array arr contains the elements 1, 2, and 3 in that
order, then the expression arr\[0\] is used to access the first
element, 1. MongoSQL does not use \"dot\" notation for array element
access.

For array element access, the \<index\> must statically have type to
NULL or INT, and may evaluate to MISSING. The semantics of array element
access match those of the aggregation operator
[\$arrayElemAt](https://docs.mongodb.com/manual/reference/operator/aggregation/arrayElemAt/#exp._S_arrayElemAt).
That is:

- If the index is zero or positive, it returns the element at that
  index, counting up from zero at the start of the array.

- If the index is negative, it returns the element at that index
  counting down from the end of the array, where -1 is the last
  element, -2 the second to last, and so on.

- If the index exceeds the array bounds, it returns MISSING.

- If the index is NULL or MISSING, it returns NULL.

#### Slicing

MongoSQL supports slicing arrays to access subsets of arrays via the
SLICE(\<array\>, \<start\>, \<length\>) [scalar
function](#scalar-functions). The semantics of array
slicing match those of the aggregation operator
[\$slice](https://docs.mongodb.com/manual/reference/operator/aggregation/slice/#exp._S_slice).
Specifically:

- If any of the arguments are NULL or MISSING, the result of the slice
  operation is NULL.

- The \<array\> argument is required and must statically have type
  ARRAY or NULL, and may be missing.

- A \<start\> argument is optional. If provided, it must statically
  have type INT or NULL, and may be missing.

  - If positive, the slice starts at that position counting up from
    the start of the array.

    - If it exceeds the bounds of the array, the resulting slice
      is empty.

  - If negative, the slice starts at that position counting down
    from the end of the array.

    - If it exceeds the bounds of the array, the starting position
      is the start of the array.

- The \<length\> argument is required and must statically have type
  INT or NULL, and may be missing. If \<start\> is provided,
  \<length\> must be positive; this will be enforced at runtime and
  the result will be NULL if it is not positive.

  - If positive, the slice includes up to the first \<length\>
    elements from the \<start\> position, or from the start of the
    array if \<start\> is omitted.

  - If negative, the slice includes up to the last \<length\>
    elements of the array.

<div id="array-grammar" />

#### Grammar

\<array expression\> ::= \"\[\" \"\]\"</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;\| \"\[\" [\<expression\>](#expressions) (,[\<expression\>](#expressions))\* \"\]\"

\<array index expression\> ::= [\<expression\>](#expressions) \"\[\" [\<expression\>](#expressions) \"\]\"

\<array slice expression\> ::= SLICE \"(\" [\<expression\>](#expressions) , ([\<expression\>](#expressions) ,)?  [\<expression\>](#expressions) \")\"

#### Examples

[Array indexing tests](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/array.yml)

[Slice tests](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/slice.yml)

#### Rejected Alternatives/Competitive Analysis

##### Indexing

We chose to zero-index arrays in MongoSQL because that is what the BSON
spec/MongoDB do. Competitors are split on this decision:

- PartiQL [zero-indexes](https://partiql.org/tutorial.html#array-number)

- N1QL [zero-indexes](https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/nestedops.html#element-selection)

- Presto [one-indexes](https://prestodb.io/docs/current/functions/array.html)

- Rockset [one-indexes](https://docs.rockset.com/data-types/#array)

##### Slicing

We considered three options to support slicing:

1. **Function**: A SLICE scalar function which maps onto \$slice.

2. **Operator**:

    a. A slice operator that uses square brackets and a colon.</br> \<array\>\[\<start\>:\<end\>\]

    b. A slice operator that uses square brackets and an ellipsis.</br> \<array\>\[\<start\>\...\<end\>\]

3. **Both**: The SLICE function _and_ the slice operator.

Typically, the two approaches have different semantics. The function
approach usually has \<start\> and \<length\> parameters, as the
[Slicing](#slicing) section describes. The operator
approach usually has \<start\> and \<end\> parameters, which takes the
slice from the \<start\> index to the \<end\> - 1 index. We chose not to
support the operator because 1) it does not map as cleanly onto \$slice, 2)
it would be unconventional to change it so that it did map cleanly
onto \$slice, and 3) it is not necessary to have two ways of doing the
same thing. Many competitors support array slicing, and each of them
only supports it one way:

- [Presto](https://prestodb.io/docs/current/functions/array.html)
  and [Rockset](https://docs.rockset.com/array-functions/#slice)
  support it via a function

- [N1QL](https://docs.couchbase.com/server/current/n1ql/n1ql-language-reference/nestedops.html#array-slicing)
  supports it via an operator

- PartiQL does not support it

##### Implicit Array Traversal

MongoSQL does not support any implicit array traversal. Implicit array
traversal refers to the application of an operation over an array
without explicit user-provided syntax to do so. There are many examples
of this. One such example is field access on arrays in MongoDB
aggregation. For the agg expression \"\$arr.b\", if the field ref \$arr
resolves to an array, then the subfield access for b is mapped over the
array, resulting in an array.

Another is conditional evaluation on arrays like the following:

ρ<sub>c</sub>: { </br>
&nbsp;&nbsp;test: {</br>
&nbsp;&nbsp;&nbsp;&nbsp;foo: \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{\'a\': \[4, 5, 6\]}</br>
&nbsp;&nbsp;&nbsp;&nbsp;\]</br>
&nbsp;&nbsp;}</br>
}

SELECT a FROM foo WHERE a > 5

Implicit array traversal may apply the > 5 operation to each element of
the array a, resulting in true if any element meets the criteria.
Alternatively, it could result in true if all elements meet the
criteria. Regardless, such implicit traversal is not supported by
MongoSQL. If users wish to achieve this behavior, they can use explicit
syntax to operate over arrays.

### Null & Missing

#### Behavioral Description

MQL (as well as BSON itself) has a distinction between NULL and MISSING.
In the case of NULL there is a field with the literal value NULL,
whereas in the case of MISSING, the field is gone. We will maintain this
distinction for several reasons:

- Maintaining the distinction simplifies translations, which should
  result in faster query performance (no conditionals to upconvert
  missing to NULL like in the BI Connector).

- Upconverting MISSING to NULL for base fields in the collection would
  require collecting the list of every field in every document for
  SELECT \* cases when no static schema is provided for the
  collections used in the query.

#### Specific Expression Semantics

The specific NULL/MISSING expression semantics will be consistent with
SQL92, for SQL92 operators.

#### String Operators

For all string operators, a MISSING as any argument will be treated as
NULL, and the entire result shall be NULL. See Section [Semantics of
String Operators](#semantics-of-string-operators).

#### Arithmetic Operators

MISSING as either argument will be treated as NULL, and the result shall
be NULL. See Section [Semantics of Arithmetic
Operators](#semantics-of-arithmetic-operators).

#### Comparison Operators

For all SQL92 comparison operators, a MISSING as either argument will be
treated as NULL, and the entire result shall be NULL.

#### Boolean Operators

MISSING as either argument will be treated as NULL, and the truth tables
from Section [Semantics of Boolean
Operators](#semantics-of-boolean-operators) will be
followed.

#### Control Flow Operators

There is no special handling for NULL or MISSING in conditions; cases
with conditions that evaluate to NULL or MISSING are skipped. See
section [Control Flow Operators](#control-flow-operators).

#### GROUP BY keys and DISTINCT

For the purposes of GROUP BY keys, and the DISTINCT keyword, MISSING is
converted to NULL, and NULL is considered distinct from all non NULL
values. See [this SERVER bug](https://jira.mongodb.org/browse/SERVER-21992?focusedCommentId=1370113&page=com.atlassian.jira.plugin.system.issuetabpanels%3Acomment-tabpanel#comment-1370113)
for a case where MQL _does not_ follow these rules at the moment.

#### Document Constructors

Any expression that results in a missing value in a document, will
remove the associated key from the document, for example:

ρ<sub>c</sub>: ⟨</br>
&nbsp;&nbsp;test: {</br>
&nbsp;&nbsp;&nbsp;&nbsp;bar: \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{'a':41, 'b':42},</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{'c':23}</br>
&nbsp;&nbsp;&nbsp;&nbsp;\]</br>
&nbsp;&nbsp;}</br>
⟩

SELECT VALUE {\'a\': a} FROM test.bar AS bar

Will result in:

{\'a\': 41}</br>
{}

Because the \'a\' key is MISSING in the second document.

#### Array Constructors

In array constructors, MISSING is converted to NULL. Consider again the
catalog from the [Document Constructors](#document-constructors) example, and the
following query:

SELECT VALUE {\'a\': \[a\]} FROM test.bar AS bar

Will result in:

{\'a\': \[41\]}</br>
{\'a\': \[null\]}

Because the MISSING \'a\' key in the second document is converted to
NULL.

#### MQL Scalar Functions

Any MQL scalar functions exposed through MongoSQL will have the same
NULL/MISSING semantics as the underlying scalar functions. By
translating directly to the MQL scalar functions, we improve query
performance.

##### Differences in handling of NULL and MISSING in Aggregation Expressions vs MQL Find Language

NULL/MISSING are treated differently in aggregation expressions vs MQL
find. A deep dive into this is available in
[Implementation Considerations for Translation to MQL: NULL/MISSING in Aggregation vs Find](#nullmissing-in-aggregation-vs-find).

#### Examples

[Query Tests](https://github.com/10gen/mongosql-rs/blob/master/tests/spec_tests/query_tests/null_and_missing.yml)

#### Rejected Alternatives/Competitive Analysis

We could upconvert MISSING into NULL, but it would result in larger data
sent across the wire (potentially exponential) and slower queries.

- Partiql/N1QL/SQL++ - Preserves distinction between NULL and MISSING.

- Presto does not have MISSING values.

- Drill returns NULL when MISSING values are accessed.

Next we will show where MongoSQL differs from PartiQL, but is consistent
with MQL:

- In PartiQL, a MISSING in an array context results in a literal value
  MISSING, which is then given to the particular serializer for the
  implementation (this part is left up to the particular data
  structures for which PartiQL is defined) to decide how to
  translate that MISSING.

  - SELECT VALUE \[foo.a, 43\] FROM bar

  - There is no literal MISSING value in BSON, so we will convert to
    NULL when foo.a is MISSING.

  - In some sense this is actually consistent with PartiQL in that
    our particular serializer (to BSON) upconverts MISSING to
    NULL, but it makes sense to explicitly reference this here.

<div id="scoping-rules" />

## Scoping Rules

### Scoping for Value Names

As in any programming language, the MongoSQL semantics have to deal with
issues of name scope. For example, how are references to foo resolved in
the following query with the following catalog:

SELECT foo.hello AS hello</br>
FROM test.bar AS foo</br>
WHERE foo.hello = ANY (SELECT foo.goodbye FROM test.foo AS foo)

_ρ_<sub>c</sub> = {</br>
&nbsp;&nbsp;test: {</br>
&nbsp;&nbsp;&nbsp;&nbsp;foo: \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{ foo: { bar: 1 }, bar: 2, goodbye: \"friend\"}</br>
&nbsp;&nbsp;&nbsp;&nbsp;\],</br>
&nbsp;&nbsp;&nbsp;&nbsp;bar: \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{ hello: \"world\" },</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{ hello: \"friend\" }</br>
&nbsp;&nbsp;&nbsp;&nbsp;\]</br>
&nbsp;&nbsp;}</br>
}

Generally, for each evaluation of a clause, the values environment tuple
is concatenated with the current input tuple. Names that exist only in
the values environment are therefore drawn from the values environment.
Names that exist only in the input tuple are therefore drawn from the
input tuple. Names that exist in both are drawn from the input tuple
since those have the closer scope.

Back to the example: Since this is a SQL query and MongoSQL is
compatible with SQL, the foo in foo.goodbye resolves to the variable foo
defined by the inner query's FROM clause. Technically, this scoping rule
is captured by the following handling of documents:

The inner query (and, therefore, its FROM clause) is evaluated with a
values environment ρ<sub>0</sub> = ⟨foo: v⟩; this foo is the one defined by the
outer FROM, which is test.bar. Then the inner FROM clause outputs a
stream of binding tuples of the form t = ⟨foo: v⟩; this foo is defined
by the inner FROM, which is test.foo. Then the expression foo.goodbye is
evaluated once for each tuple _t_ in the stream, in the environment
_ρ<sub>select</sub> =_ tupleConcat(ρ<sub>0</sub>, _t_).

Because foo appears in both ρ<sub>0</sub> and _t_, the concatenation keeps only
the foo value from its right argument (_t_) as given in the rules in
Section [Abstract Model](#abstract-model). Essentially, by
putting _t_ as the right argument of the concatenation, the variables of
_t_ have precedence over synonymous variables in the left argument
(ρ<sub>0</sub>).

### Determining Qualification For References

Qualified references are references formed as q.x, where q is the
qualifier. However, as seen in Section [Document and Field-Access
Expressions](#document-and-field-access-expressions), d.x
is an access to field x in document d. This makes it very important that
we have clear rules to determine whether foo.bar in a given position is
a qualified reference or a document field access.

We disambiguate by looking at the data source names in the values
environment (at any nesting depth). If foo is one of the data source
names, we assume that \"foo.\" is a qualification instead of a field
reference in a document named foo.

This begs the question, how does one access the bar field under the foo
field in the test.foo collection from _ρ_<sub>c</sub> in our above
example, if foo is a current qualified name? There are three
options: alias foo to something else in the FROM clause, use
foo.foo.bar, or use foo\[\'bar\'\], because \[\] field access is not
ambiguous.

Example:

SELECT foo.bar FROM foo AS foo

Here, foo.bar is the qualified field bar from foo. So the result is:

{\'bar\': 2}

If we meant to access the foo field in foo, we could, instead, write any
of the following:

SELECT foo.bar FROM foo AS f

or

SELECT foo.foo.bar FROM foo AS foo

or

SELECT foo\[\'bar\'\] FROM foo AS foo

each of which would result in:

{\'bar\': 1}

Note that a given FROM clause can introduce multiple data source
quantifiers:

FROM foo AS foo JOIN bar AS bar introduces both foo and bar as name
qualifiers. Additionally, qualifiers can come from smaller scope numbers
in the case of correlated subqueries: SELECT (SELECT \* FROM foo WHERE
foo.x = bar.x) FROM bar, here, bar.x is a qualified reference because
bar is a data source quantifier with a smaller scope number.

### Determining Datasources for References

In all cases, the following algorithm is applied _statically_ at _query
planning time_ to determine which MongoDB datasource to associate. No
dynamic datasource decisions are made during query execution. If a query
passes query planning stage, and a value is missing at _runtime_, it
becomes a MISSING value. If it is missing _statically_, it is an error
during query planning.

- Definitions:

  - MUST contain f =\> we have a schema that proves a datasource is
    guaranteed to have the top-level field f

  - CANNOT contain f =\> we have a schema that proves a datasource
    cannot contain the top-level field f

  - MAY contain f =\> cannot prove or disprove the existence of a
    top-level field f in a datasource

- Qualified Reference Resolution

  - Qualified Reference: q.f1(.f2\...fn)

  - We first need to resolve the qualifier q to a datasource

  - Once we have resolved the datasource, we resolve the field by
    accessing the subpath f1.(f2\...fn) in that datasource's
    top-level document

  - By the definition of a qualified reference, there is guaranteed
    to be at least one key (n, d) in the environment such that n =
    q

  - To resolve the qualifier q, we consider all the keys (n, d) with
    n = q, and choose the one with the largest value for d

- Unqualified Reference Resolution

  - Unqualified reference: f1(.f2\...fn)

  - Once we have resolved the datasource, we resolve the field by
    accessing the subpath f1.(f2\...fn) in that datasource's
    top-level document

  - We first need to resolve the implicit qualifier q to a
    datasource

  - If there is only a single datasource C such that C MUST or MAY
    contain f1, we resolve q to that datasource

  - Else, apply the following algorithm, starting with the largest
    nesting depth d.

    - Consider all bindings at the current nesting depth

    - If all such bindings CANNOT contain f1, continue at the
      next-smallest nesting depth d-1

    - If there is more than one such binding that MUST or MAY
      contain f1, error

      - Note: This is the case that covers non-subquery-related
        ambiguities like SELECT a FROM foo JOIN bar

    - If there is exactly one such binding that MUST contain f1,
      resolve q to that datasource

    - Else (i.e. if there is exactly one binding that MAY contain
      f1), error

      - Note: this is the case that covers correlation-related
        ambiguities. We error here because it means that the
        field could possibly resolve to multiple different
        nesting depths.

  - If the above algorithm does not resolve q to any datasource,
    return a "field not found" error

### Resolving Collection References

Identifiers that reference collections are distinct from other
references, and can be statically disambiguated based on their position
in the query: collection references occur only in FROM clauses, and
regular references to fields cannot occur in FROM clauses. A collection
reference may be a qualified reference or not, with the qualification
denoting a database rather than a datasource. Unlike expression
positions, this is easy to determine, because there are no field
references allowed in FROM clauses: any FROM clause reference containing
\`.\` is qualified. All unqualified references in FROM clauses are
turned into qualified references by prefixing the current database.

## Miscellaneous

### Document Key Ordering Semantics

#### Behavioral Description

MongoSQL does not provide any guarantees about the ordering of keys in
documents. This means that any document value may be returned from a
query with its keys in an arbitrary order, whether that document was
created by a literal, selected unmodified from a collection, or
otherwise. In practice, document keys will often be returned in the
order they are specified in a literal or the order they are stored in
the database, but users should not rely on this behavior.

Currently, MongoSQL itself does not implement any functionality that
depends on the orderedness (or unorderedness) of document keys. This
limitation is intentional; we wish to have the flexibility to decide in
the future that documents are ordered or unordered. This is why
comparison operators (and some uses of clauses that implicitly perform
comparisons, like ORDER BY, GROUP BY, and SELECT DISTINCT) are
disallowed over document fields.

#### Rejected Alternatives/Competitive Analysis

PostgreSQL does not allow comparison between JSON values. Comparison
between JSONB values and JSONB equality is not affected by key order.
JSON and JSONB documents in result sets do not necessarily have the same
key order as specified in the query.

MySQL allows comparison between JSON values, and JSON equality is not
affected by key order. JSON documents in result sets do not necessarily
have the same key order as specified in the query.

PartiQL allows comparisons between tuples, and equality is not affected
by key order. PartiQL explicitly defines its tuples as unordered sets of
key-value pairs.

Presto allows comparisons between JSON objects, and key order is not
considered for equality.

### Not A Number (NaN) Equality

MongoDB differs from IEEE Std 754-2008, Section 5.11, Para. 2 and
supports equality comparisons between NaN values (\<=, =, !=, \>=).

Example:

ρ<sub>c</sub>: ⟨ </br>
&nbsp;&nbsp;test: {</br>
&nbsp;&nbsp;&nbsp;&nbsp;nan: \[</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{\'a\': NaN, \'b: NaN},</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{\'a\': 0, \'b\': 42}</br>
&nbsp;&nbsp;\]</br>
&nbsp;&nbsp;}</br>
⟩

SELECT \* FROM test.nan AS foo WHERE nan.a = nan.b</br>
SELECT \* FROM test.nan AS foo WHERE nan.a = \'NaN\'::DOUBLE

Note: \'NaN\'::DOUBLE will result in a NaN value

Will both result in:

{\'a\': NaN, \'b\': NaN}

SELECT \* FROM test.nan AS foo WHERE nan.a != nan.b</br>
SELECT \* FROM test.nan AS foo WHERE nan.a != \'NaN\'::DOUBLE

Will both result in:

{\'a\': 0, \'b\': 42}

SELECT {\'c\': a = b } from test.nan AS nan

Will result in:

{\'c\': true}

This may be surprising to some users, but is required by MongoDB to
support indexing.

### MongoSQL Exceptional Behavior

MongoSQL has two categories of exceptional behavior:

- Static errors

- Undefined Behavior

A MongoSQL query should never result in a runtime error. Any runtime
error that causes a termination of a query shall be considered a bug. It
is common in the analytics world for queries to run for hours, or even
days, so runtime errors are a particularly undesirable failure mode.

#### Static Errors

We attempt to detect as many exceptional situations as possible at
compile time so that we can return a static error. Static errors
immediately halt execution and allow the user to correct the problem
with their query. Some examples of static errors:

- Ambiguous references or duplicate binding tuple names
- Scalar subqueries not guaranteed to return 0 or 1 binding tuples
- Arguments to operators and scalar functions that cannot be
  statically determined to be the expected types
- References known to be MISSING, statically
  - We view MISSING as something that is not intended behavior

#### Undefined Behavior

Not every observable behavior of a MongoSQL implementation is guaranteed
by this specification. Behaviors that are not guaranteed will be
referred to as "Undefined Behavior" (UB). UB should not be relied upon,
as it may differ between implementations or be changed at any time.

The reason for having UB is to prevent certain kinds of implementation
details from leaking into the specification without incurring an
unreasonable runtime cost. For example, MQL will upconvert ints to
doubles on arithmetic overflow, which is an undesirable behavior that
may cause problems for BI Tools. Instead of wrapping every integer
addition translation in an expensive and unidiomatic conditional that
detects when overflow will occur, we instead declare the integer
overflow behavior to be UB.

### Comments

Comments are sequences of characters within queries that do not impact
query execution. MongoSQL supports standard SQL comments and C-style
block comments.

Standard SQL comments begin with double dashes and end with a new line:

\-- This is a standard SQL comment

Block comments begin with /\* and end at the matching occurrence of \*/.
They may span multiple lines and can be nested:

/\* This is a multiline comment</br>
\* with nesting: /\* nested comment \*/</br>
\*/

## Future Work

### Exposing MongoDB-specific Types in ODBC/JDBC

- How do we expose MongoDB-specific types (like ObjectId) via
  ODBC/JDBC?

  - One option is custom types in JDBC/ODBC. It will require some
    experimentation to see whether BI Tools can actually consume
    custom types.
  - If custom types don't work, a second option is to expose the
    whole BSON value (including type tag) as binary data.
    - This would require us to distinguish between whole BSON
      values represented as binary data and actual BSON BinData
      values
  - If neither of the above options work, we can fall back to a hack
    similar to the one we used in the BI Connector: expose
    ObjectIds to BI Tools as strings, and try to infer when string
    literals in queries are actually supposed to be ObjectId
    literals
    - If we did this, we would want to hide it behind a feature
     flag that would be opt-in, and would likely only be used
     by JDBC and ODBC drivers.

### Date and Time Types

BSON includes a Date type and a Timestamp type. The BSON Date type
represents a datetime; i.e., it consists of a date part and a time part.
The BSON Timestamp type is a special type intended for internal MongoDB
use.

SQL-92 includes a Date type, a Time type, and a Timestamp type. The
SQL-92 Date type represents a date without a time; i.e., it only
consists of a date part. The SQL-92 Time type represents a time without
a date; i.e., it only consists of a time part. The SQL-92 Timestamp type
represents a datetime; i.e., it consists of a date part and a time part,
meaning it is the same as the BSON Date type.

Since we want our MongoSQL types to be the same as the set of BSON types
if possible, we are not currently supporting the SQL DATE or TIME type.
We will do so in the future if and only if it becomes necessary to do so
in order to make BI Tools work.

_Note_: the same approach will apply for other places where our type
system deviates from the SQL standard. For example, we would consider
adding semantic support for parameterized types like VARCHAR(3) if and
only if it becomes necessary to do so in order to get BI Tools working.

### Exposing MQL Functionality

We may want to consider adding a way to access arbitrary MQL expressions
via MongoSQL, so that users can leverage MQL expressions without the
need for adding explicit support for each one in MongoSQL itself. This
work is tracked in the
[SQL-70](https://jira.mongodb.org/browse/SQL-70) ticket,
and previous design discussions on this topic can be found in this
google doc's named version history.

### Collations

A number of clauses in standard SQL allow users to explicitly specify
the collation that should be used for comparisons (ORDER BY and GROUP
BY, for example). MongoSQL does not allow this at the moment, though we
may want to consider it in the future.

Beyond intra-type comparisons for strings, one could also imagine a
future where collations control structured-data comparisons (for
example, how key order and duplicate keys are treated in document
comparisons) and inter-type comparisons (for example, how a document and
a boolean compare).

### ORDER BY arbitrary expressions

For now, we only allow ORDER BY sort keys to be column names or integer
literals. In the future, we may wish to expand this to arbitrary
expressions.

### Supporting Non-Document BSON Values In Query Results

For now, only documents are returned as the values in binding tuples. We
can relax this in the future to allow for something such as SELECT VALUE
\[a, b\] FROM foo. This will require a slight rework on how the binding
names are modified in derived tables because that currently uses
\$mergeObjects on the underlying values of the original binding tuples
from the SELECT clause of the derived table query.

### SELECT DISTINCT

Support for SELECT DISTINCT has been deferred pending a decision re:
ordered vs. unordered document comparison support in MongoSQL & MQL

### UNION DISTINCT

Support for distinct UNION has been deferred pending a decision re:
ordered vs. unordered document comparison support in MongoSQL & MQL

### USING CLAUSE

Support for USING Clause in equality join conditions has been deferred
until we can find a solution on resolving the namespaces.

### Unify arrays and subqueries

In the future, we should be able to support operations currently
reserved for subqueries over arrays as well (and vice versa). This could
include some or all of the following:

- Support IN, ANY, and ALL expressions with array values on the RHS

- Support a syntax for using a subquery's entire result set as an
  array value

## Appendix

### Formal Definitions

#### tupleConcat

_tupleConcat_ : _Ρ × Ρ → P_

_tupleConcat_ is formally defined using the following rules. In the
rules, variables _x_ are keys, _v_ are BSON values, _x: v_ is a binding
from a binding tuple key _x_ to BSON value _v_, and variables _X_ are
arbitrary sequences of key, value bindings. Note that variables _X_ are
allowed to match the empty binding.

1. _tupleConcat_(⟨(_x_, _n_): v<sub>0</sub>, X<sub>1</sub>⟩, ⟨( _x_, _n_): v<sub>1</sub>, X<sub>2</sub>⟩) ⇒</br>
_tupleConcat_(⟨(_x_, _n_): v<sub>1</sub>, X<sub>1</sub>⟩, ⟨X<sub>2</sub>⟩)
2. _tupleConcat_(⟨X<sub>0</sub>⟩, ⟨X<sub>1</sub>⟩) _⇒_ ⟨X<sub>0</sub>, X<sub>1</sub>⟩ _otherwise_

These rules describe how we merge binding tuples: if two keys have the
same name between the two binding tuples, we choose the one from the
right binding tuple. Duplicate keys within one binding tuple are
undefined behavior, and the implementation is free to leave them or
remove them, and choose which duplicate to remove, if removal is chosen.
The following example shows how these rules are applied:

_tupleConcat_(⟨(x, 0): 1, (y, 0): 13, (z, 0): 4⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨(x, 0): 3, (a, 0): 1, (z, 0): 14⟩)

Apply rule 1:

_tupleConcat_(⟨(x, 0): 3, (y, 0): 13, (z, 0): 4⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨(a, 0): 1, (z, 0): 14⟩ )

Apply rule 1:

_tupleConcat_(⟨(x, 0): 3, (y, 0): 13, (z, 0): 14⟩,</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;⟨(a, 0): 1⟩ )

Apply rule 2:

⟨(x, 0): 3, (y, 0): 13, (z, 0): 14, (a, 0): 1⟩

And no more rules can be applied, this is our result.

We also extend _tupleConcat_ over arrays of binding tuples as follows,
given an array arr:

_tupleConcat_(d<sub>0</sub>, _arr_) = \[_tupleConcat_(d<sub>0</sub>, _x_) for _x_ ∊
_arr_\]

Similarly:

_tupleConcat_(_arr_, d<sub>0</sub>) = \[_tupleConcat_(_x_, d<sub>0</sub>) for _x_ ∊
_arr_\]

And:

_tupleConcat_(_arr<sub>0</sub>_, _arr<sub>1</sub>_) = \[_tupleConcat_(_x_,_y_) for _x_ ∊
_arr<sub>0</sub>_, for _y_ ∊ _arr<sub>1</sub>_\]

### Implementation Considerations for Translation to MQL

#### MQL Runtime Errors

We begin by looking at a non-exhaustive list of runtime errors thrown by
MQL as determined from the MQL documentation, any translation to MQL
must handle these errors. The \$convert operator can be used to catch
errors.

- \$acos: throws an error for input values outside of \[-1, 1\] or
  non-numeric input
- \$acosh: throws an error for input values outside of \[-1, ∞) or
  non-numeric input
- \$asin: throws an error for input values outside of \[-1, 1\] or
  non-numeric input
- \$atanh: throws an error for input values of exactly 1 and -1 or
  non-numeric input
- \$binarySize: throws an error for input that is not a string,
  binary, or BSON null
- \$bsonSize: throws an error for input that is not a document or BSON
  null
- \$convert: can throw an error if no onError attribute is specified
  and the expression to be converted throws an error or if an error
  occurs dugin conversion
  - \$convert may be used as a error handler because of this
    behavior
- \$cos: throws an error for input values of -∞ and ∞ or non-numeric
  input

- \$dateFromParts: throws errors for out of range parts, for MongoDB
  4.4 this means:
  - year: 1-9999
  - isoWeakYear: 1-9999
  - month: 1-12
  - isoWeek: 1-53
  - day: 1-31
  - isoDayOfWeek: 1-7
  - hour: 0-23
  - minute: 0-59
  - second: 0-59
  - millisecond: 0-999
  - any non-conformant timezone
- \$dateFromString: a convenience function for \$convert, works as
  \$convert and errors as \$convert
- \$dayOfMonth: throws an error if argument is not a Mongo data object
- \$dayOfWeek: throws an error if argument is not a Mongo data object
- \$dayOfYear: throws an error if argument is not a Mongo data object
- \$first: throws an error if argument is not an array, MISSING, or
  BSON null
- \$geoNear: throws an error if there is more than one 2d index or
  more than one 2dsphere index and a key is not provided
- \$hour: throws an error if argument is not a Mongo data object
- \$in: throws an error in either of the following cases: if the \$in
  expression is not given exactly two arguments, or if the second
  argument is not an array
- \$indexOfArray: throws an error if the first argument is not an
  array, BSON null, or MISSING.
- \$indexOfCP: throws an error if the first argument is not a string,
  BSON null, or MISSING.
- \$isoDayOfWeek: throws an error if argument is not a Mongo data
  object
- \$isoWeek: throws an error if argument is not a Mongo data object
- \$isoWeekYear: throws an error if argument is not a Mongo data
  object
- \$last: throws an error if argument is not an array, MISSING, or
  BSON null
- \$millisecond: throws an error if argument is not a Mongo data
  object
- \$minute: throws an error if argument is not a Mongo data object
- \$month: throws an error if argument is not a Mongo data object
- \$replaceAll: throws an error if any argument is not a string or
  BSON null (it throws an error if any argument is MISSING).
- \$replaceOne: throws an error if any argument is not a string or
  BSON null (it throws an error if any argument is MISSING).
- \$round: throws an error if any argument is not numeric, MISSING, or
  BSON null
- \$second: throws an error if argument is not a Mongo data object
- \$sin: throws an error for input values of -∞ and ∞ or non-numeric
  input
- \$sinh: throws an error for non-numeric input
- \$size: throws an error if the input is not an array (it errors for
  MISSING or BSON null)
- \$split: throws an error on non-string/BSON null/MISSING input
- \$sqrt: throws an error on non-numeric/BSON null/MISSING input or on
  negative numeric input
- \$substrBytes: throws an error if there are not exactly three
  arguments, or if either index results in a byte in the middle of a
  UTF-8 rune. It does not error on inputs that are not strings
- \$substrCP: throws an error if there are not exactly three
  arguments. It does not error on inputs that are not strings
- \$switch: throws an error if no default is specified and no case
  matches
- \$tan: throws an error for input values of -∞ and ∞ or non-numeric
  input
- \$tanh: throws an error for non-numeric input
- \$toBool, \$toDate, \$toDecimal, \$toDouble, \$toInt, \$toLong,
  \$toObjectId, \$toString: a convenience function for \$convert,
  works as \$convert and errors as \$convert
- \$trunc: throws an error for non-numeric input
- \$week: throws an error if argument is not a Mongo data object
- \$year: throws an error if argument is not a Mongo data object
- \$zip: if any argument is not an array, or useLongestLength is set
  to true and no defaults are specified or are empty.

Additionally, we have a list of known errors that are not in the
documentation:

- \$add: throws an error on non-numeric/BSON null/MISSING input

- \$cosh: throws an error for non-numeric input

- \$divide: throws an error on non-numeric/BSON null/MISSING input

- \$multiply: throws an error on non-numeric/BSON null/MISSING input

- \$subtract: throws an error on non-numeric/BSON null/MISSING input

- \$sinh: throws an error for non-numeric input

- \$tanh: throws an error for non-numeric input

#### NULL/MISSING in Aggregation vs Find

For the purposes of the MQL find language, which is also the default
language of the \$match aggregation stage, NULL and MISSING are treated
the same. In other parts of MQL aggregation, NULL and MISSING are
treated as distinct for the purposes of the \$eq function.

Example:

```sh

> db.nm.find()

{ "_id" : ObjectId("5fd50cdebe80dc7690b03783"), "a" : 1 }
{ "_id" : ObjectId("5fd50ce0be80dc7690b03784"), "a" : 2 }
{ "_id" : ObjectId("5fd50ce3be80dc7690b03785"), "a" : null }
{ "_id" : ObjectId("5fd50ce5be80dc7690b03786") }

> db.nm.find({"a": {"$eq": null}})

{ "_id" : ObjectId("5fd50ce3be80dc7690b03785"), "a" : null }
{ "_id" : ObjectId("5fd50ce5be80dc7690b03786") }

> db.nm.aggregate({"$addFields": {"out": {"$eq": ["$a", null]}}})

{ "_id" : ObjectId("5fd50cdebe80dc7690b03783"), "a" : 1, "out" : false }
{ "_id" : ObjectId("5fd50ce0be80dc7690b03784"), "a" : 2, "out" : false }
{ "_id" : ObjectId("5fd50ce3be80dc7690b03785"), "a" : null, "out" : true }
{ "_id" : ObjectId("5fd50ce5be80dc7690b03786"), "out" : false }
```

MISSING and NULL, however, are also treated the same in the \$group
stage for the purposes of grouping:

```sh
> db.nm.aggregate({"$group": {"_id": "$a"}})
{ "_id" : 2 }
{ "_id" : 1 }
{ "_id" : null }
```

As mentioned in Section [GROUP BY keys and DISTINCT](#group-by-keys-and-distinct),
see [this SERVER bug](https://jira.mongodb.org/browse/SERVER-21992?focusedCommentId=1370113&page=com.atlassian.jira.plugin.system.issuetabpanels%3Acomment-tabpanel#comment-1370113)
for a case where MQL _does not_ currently treat NULL and MISSING the
same. Our translations will need to take this bug into account.
