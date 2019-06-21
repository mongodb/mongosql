# Changelog

This file contains a list of user-facing changes.
A line should be added to this file for each commit made to this repository.
If no user-facing changes were made, the comment should reflect that fact.

## 2.12

BI-2028: Using current_timestamp and trim functions with too many arguments now gives a parse error.
BI-1843: Fixed a bug that caused timestampAdd to fail with out of range months.
BI-2198: Fixed a bug that caused the BI Connector to experience a pushdown failure when receiving invalid SQL queries attempting to group by an aggregate function.
BI-2197: Fixed a bug that caused datediff to fail to push down with polymorphic arguments.
BI-1750: Improved truncate pushdown coverage so that SQL queries with a column of integers representing the number of places to truncate are also pushed down, and uses the new $trunc operator for versions running on MongoDB versions >= 4.1.9.
BI-2250: No user-facing changes.
BI-2216: Pipelines in DRDL files are parsed with most recent version of extjson spec, but still support legacy binary, regex, and dbPointer formats.
BI-2218: No user-facing changes.
BI-2231: No user-facing changes.
BI-2016: No user-facing changes.
BI-2099: No user-facing changes.
BI-2228: No user-facing changes.
BI-2214: No user-facing changes.
BI-2229: No user-facing changes.
BI-2227: No user-facing changes.
BI-2223: No user-facing changes.
BI-1813: No user-facing changes.
BI-2173: No user-facing changes.
