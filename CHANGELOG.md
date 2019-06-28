# Changelog

This file contains a list of user-facing changes.
A line should be added to this file for each commit made to this repository.
If no user-facing changes were made, the comment should reflect that fact.

## 2.11.1

BI-2259: Fixed a bug in our foreign key construction that caused some foreign keys to point to unrelated tables.
BI-2258: Fixed a bug that caused incorrect pushdown and type conversions for dates/datetimes that are too large/small. Also fixes instances where pushdown was failing due to ast.Unknown when ast.Constant is expected.
BI-2251: Fixed a bug that caused information_schema tables to not alias properly.
