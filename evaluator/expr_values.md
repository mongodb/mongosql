
# SQLValues Overview

This document gives an overview of all the types that pertain to SQLValues, how they relate to each other, and what they are for.
The types discussed here can be found in `expr_values.go`, `expr_values_base.go`, `expr_values_mongosql.go`, and `expr_value_mysql.go`.

## SQLValue Interface

The `SQLValue` interface is the primary way in which the rest of the `evaluator` package will interact with the `SQLValue` type system.
A `SQLValue` has the following methods/embedded interfaces.

### Embeds `SQLExpr`

A `SQLValue` must be a `SQLExpr`. This should be self-explanatory.

### Embeds `SQLProtocolEncoder`

Every `SQLValue` must have a `WireProtocolEncode` method, which defines how this `SQLValue` is serialized into bytes to be sent across the wire.

### Embeds `SQLValueConverter`

Each `SQLValue` must be able to be converted to each different kind of `SQLValue`.

### Implements `IsNull()`

The `IsNull` function returns `true` if a SQLValue is null, and `false` otherwise.
A `SQLValue` of any type is nullable, and `SQLNull` is no longer its own `SQLValue`, as it was prior to BI-1718.
In practice, the type of a null value rarely (if ever) matters, but we attempt to accurately type nulls wherever possible nonetheless.

### Implements `Value()`

The `Value` function returns a go value (of type `interface{}`) that is equivalent to the SQLValue receiver.

### Implements `Kind()`

The `Kind` function returns a `SQLValueKind` (see below for more details), that indicates the kind of the SQLValue receiver.
During the evaluation of a single query, only a single kind of value should be used.

### Implements `Size()`

The `Size` function returns a uint64, which represents the size in bytes of the SQLValue.
This is used by the memory monitor to track and limit our memory consumption while evaluating queries.

## SQLValueKind Enum

`SQLValueKind` is an enum type that has three possible values:

  - `NoSQLValueKind`
  - `MongoSQLValueKind`
  - `MySQLValueKind`

Each `SQLValue` must have a `SQLValue` kind of either `MongoSQLValueKind` or `MySQLValueKind`.
`NoSQLValueKind` is the zero value for `SQLValueKind`, and is not a legal `SQLValueKind` for a `SQLValue`.

## SQLBool,SQLDate,... Interfaces

In addition to the `SQLValue` interface, we also have more strongly typed interfaces, roughly one for each `EvalType` (though not a strict one-to-one mapping).
Those interfaces are as follows:

  - `SQLBool`
  - `SQLDate`
  - `SQLDecimal128`
  - `SQLFloat`
  - `SQLInt32`
  - `SQLInt64`
  - `SQLUint32`
  - `SQLUint64`
  - `SQLTimestamp`
  - `SQLVarchar`

These are simple interfaces, each embedding `SQLValue` and adding one additional no-op method.
This method's name is the interface name prefixed with the character `i`.
The `SQLValue` interface still encompasses all the functionality for these types,
but the additional differentiation that these interfaces enable allow for more explicit, type-safe code.
For example, these interfaces allow us to enforce that the `SQLBool` function from the `SQLValueConverter`
interface returns a `SQLBool` (i.e. a `MySQLBool` or a `MongoSQLBool`), as opposed to some other `SQLValue`.

## NewSQL* Constructors

Each of the types of `SQLValues` listed above has a constructor (`NewSQLBool`, for example) that takes a `SQLValueKind` and a go value.
First, the constructor will create a `BaseSQLBool` instance from the value and kind parameters.
Then, depending on the `SQLValueKind` provided, it will wrap the `BaseSQLBool` in either a `MongoSQLBool` or a `MySQLBool` and return it.

Thus, the invocation `NewSQLBool(MySQLValueKind, false)` will return a `MySQLBool` representing the `false` value,
while `NewSQLBool(MongoSQLValueKind, true)` will return a `MongoSQLBool` representing the `true` value.
Because we only use these constructors to create new `SQLValue`s, this helps us ensure that we are never using a `BaseSQL*` type
in the evaluator, and instead are always using one of the `MongoSQL*` or `MySQL*` types that wrap the `BaseSQL` types.

There are also two constructors for creating null values, `NewSQLNull` and `NewSQLNullUntyped`.
`NewSQLNull` takes a `SQLValueKind` and an `EvalType`, while `NewSQLNullUntyped` only takes a `SQLValueKind`,
and will return a null value of Evaltype `EvalString`.
Semantically, it doesn't matter what type a null value has, but we try to type them correctly
whenever possible anyways.

## MongoSQLBool,MongoSQLDate,... Structs

The `MongoSQL*` types embed the corresponding `BaseSQL*` structs.
Since the `BaseSQL` methods implement the MongoSQL conversion behavior, the `MongoSQL` types do not need to override any of the methods.

## MySQLBool,MySQLDate,... Structs

The `MySQL*` types embed the corresponding `BaseSQL*` structs.
Wherever the MySQL type conversion behavior differs from the MongoSQL behavior, the `BaseSQL` method corresponding
to that conversion is overridden on the `MySQL` type.

## BaseSQLBool,BaseSQLDate,... Structs

The `BaseSQL*` types implement most of the behavior for all `SQLValue`s.
Each `BaseSQL*` type implements every method needed to satisfy the `SQLValue` interface.
This way, types that embed `BaseSQL*` only need to implement methods when they want to override this default behavior.

The type conversion behavior implemented by the `BaseSQL` methods is the MongoSQL conversion behavior.
This was chosen as the default behavior because MongoSQL type conversions are more sane,
and are more likely to be reusable in other sets of type conversion semantics than the MySQL conversions.

The `BaseSQL` types are intended only for embedding in other types (in this way, they are like OOP's abstract classes).
When we return a `SQLValue` from a `BaseSQL*` method, we always want the `SQLValue` returned to be either a `MongoSQL` type
or a `MySQL` type, never a `BaseSQL` type.

To accomplish this goal, we store the `SQLValueKind` on each `BaseSQL*` struct, and return `SQLValues` created by `NewSQL*`
constructors instead of returning `BaseSQL` literals.
Because of the constructor implementation described above, this ensures that the `BaseSQL` methods always return values
of the correct types.

Without this implementation strategy, we would end up having to duplicate method declarations.
For example, if we got rid of `BaseSQL*` types and simply embedded `MongoSQL*` types in the `MySQL*` types,
we wouldn't be able to dispatch function calls to the embedded structs, because they would return
values of type `MongoSQL*` instead of `MySQL*`.
The shared set of `BaseSQL` implementations solves this problem without requiring us to reimplement or
wrap all of the `MongoSQL*` method implementations in another method on the `MySQL*` structs.