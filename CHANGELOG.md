# Changelog

This file contains a list of user-facing changes.
A line should be added to this file for each commit made to this repository.
If no user-facing changes were made, the comment should reflect that fact.

## 2.14

- BI-2407: No user-facing changes.
- BI-2184: Fix bug that caused natural left joins to fail in some cases.
- BI-2411: Increase default max_num_global_tables to 4000
- BI-2103: No user-facing changes.
- BI-2382: No user-facing changes.
- BI-2248: No user-facing changes.

## 2.13

- BI-2402: Update an error message for FLUSH SAMPLE authorization that was unclear.
- BI-2363: No user-facing changes.
- BI-2363: Reduce DynamicSourceStage memory usage
- BI-2349: Update schema mapping to skip empty field names and field names containing "." characters.
- BI-2398: No user-facing changes.
- BI-2394: Allow self-join optimization to be applied to sharded collections
- BI-2397: Fix bug that caused unix_timestamp to return incorrect results when pushed down because it did not previously handle DST when necessary.
- BI-2401: No user-facing changes.
- BI-2361: Add variable to set max number global tables (default 1000). Add variable to set max num tables per collection (default 200). Change max table nesting depth default to 10.
- BI-2380: No user-facing changes.
- BI-2392: No user-facing changes.
- BI-2353: No user facing changes.
- BI-2389: Fix bug that caused DRDL Tables to incorrectly marshal to BSON if the pipeline was empty.
- BI-2362: Made max_num_columns_per_table system variable a no-op and added max_num_fields_per_collection. Enhanced performance while sampling collections with many fields by reducing the amount of memory used.
- BI-1899: Improved ambiguous column behavior in subqueries, group by and order by clauses.
- BI-2343: `mongotranslate` is now consumable as a library.
- BI-2280: Add reconciliation to aggregate functions where necessary and add pushdown for multiple type conversions. Added a `reconcile_arithmetic_agg_functions` variable for opting back into the old reconciliation behavior.
- BI-2366: Fix bug in mongoast optimizer that incorrectly swapped project and match stages; introduce new mongoast optimizations including constant propagation, match splitting, stage reordering, and match coalescing.
- BI-2202: Improved error message when issuing a group by 0 or order by 0.
- BI-2284: Add support for Ubuntu 1804, SLES 15, Rhel8.
- BI-2379: Add type reconciliation for regexp, now regexp works with non-strings.
- BI-2376: No user-facing changes.
- BI-2355: Fix bug where some network errors during connection setup could cause panics.
- BI-2195: No user-facing changes.
- BI-2360: Improved performance for mapping schemas that contain arrays.
- BI-2230: No user-facing changes
- BI-2328: No user-facing changes.
- BI-2356: No user-facing changes.
- BI-2306: Adds support for `show dbs`.
- BI-2272: Enables `select *,<expr>` functionality.

## 2.12

- BI-1885: Pushdown correlated subqueries on versions of MongoDB >= 3.6.
- BI-2271: No user facing changes.
- BI-2161: Fix bug that caused mongosqld to hang if an election occurred during sampling. Now, an error will be returned. Clients can continue to connect to mongosqld after the election.
- BI-2261: No user-facing changes.
- BI-2232: No user-facing changes.
- BI-2312: Revendor mongoast to get the newest optimizations.
- BI-2289: Fix failing tpch-queries.
- BI-2288: Fix failures during foreign key creation with drdl files that have children tables with no _id column
- BI-2311: Fix Windows install directory version label to allow minor versions > 9.
- BI-1617: Pushdown regexp in aggregation language.
- BI-2245: Ensure that writeMode columns are sorted in declaration order. Add support for ALTER TABLE ENABLE/DISABLE KEYS
- BI-2304: Update catalog information to correctly show index and nullability information from --writeMode.
- BI-2106: Remove some unnecessary rounding from pushdown translations.
- BI-2244: Support basic INSERT statements in write mode.
- BI-1844: Fixed a bug in `div` evaluation that caused incorrect results.
- BI-1546: Support version-conditionally-executed comment statements.
- BI-1578: Fixed a bug that failed to correctly kill long-running queries.
- BI-2183: Fixed a bug that was blocking pushdown for some queries on the objectID field.
- BI-2243: New reserved words: INSERT, INTO
- BI-2290: No user-facing changes.
- BI-2117: Fix a bug that caused mongodrdl to fail to parse repl set seedlists for the `--host` flag.
- BI-2242: No user-facing changes.
- BI-2105: No user-facing changes.
- BI-2192: Remove alteration support. Make `enable_table_alterations` a no-op.
- BI-2061: LIKE expressions with literal pattern strings can be pushed down outside of a WHERE clause.
- BI-2240: No user-facing changes.
- BI-2104: Improved pushdown performance of queries with an EXISTS subquery.
- BI-2260: Added pushdown for char and str_to_date functions. Fixed a bug where str_to_date returned a date instead of a datetime for some non-constant format-string arguments.
- BI-1547: New keywords: low_priority, unlock
- BI-1444: Add support for SRV style connection URI to `mongodrdl` and `mongosqld`.
- BI-1067: Add support to `mongodrdl` for `--uri` flag.
- BI-2237: Added docs links to cli help output.
- BI-2196: Fixed a bug that caused some queries with aggregate functions to fail to push down.
- BI-2094: No user-facing changes.
- BI-2287: No user-facing changes.
- BI-2241: No user-facing changes.
- BI-2262: Fixed a bug that caused a pipeline parser error for unwind paths with numeric field names.
- BI-2267: Fixed a bug that caused mongodrdl to ignore the `--gssapiHostName` and `gssapiServiceName` flags.
- BI-2092: No user-facing changes.
- BI-2239: No user-facing changes.
- BI-1510: If no compressors are specified in the mongodb connection string, zlib,snappy will be used by default. If compressors are specified, they will be honored.
- BI-2025: Queries using `ln`, `ascii`, `user`, `database`, `version`, `connection_id`, and constant-valued time/date functions can be fully pushed down.
- BI-1460: No user facing changes.
- BI-2219: No user facing changes.
- BI-2238: New reserved words: KEY, FULLTEXT, PRIMARY.
- BI-2259: Fixed a bug in our foreign key construction that caused some foreign keys to point to unrelated tables.
- BI-2258: Fixed a bug that caused incorrect pushdown and type conversions for dates/datetimes that are too large/small. Also fixes instances where pushdown was failing due to ast.Unknown when ast.Constant is expected.
- BI-2251: Fixed a bug that caused information_schema tables to not alias properly.
- BI-2028: Using current_timestamp and trim functions with too many arguments now gives a parse error.
- BI-1843: Fixed a bug that caused timestampAdd to fail with out of range months.
- BI-2198: Fixed a bug that caused the BI Connector to experience a pushdown failure when receiving invalid SQL queries attempting to group by an aggregate function.
- BI-2197: Fixed a bug that caused datediff to fail to push down with polymorphic arguments.
- BI-1750: Improved truncate pushdown coverage so that SQL queries with a column of integers representing the number of places to truncate are also pushed down, and uses the new $trunc operator for versions running on MongoDB versions >= 4.1.9.
- BI-2250: No user-facing changes.
- BI-2216: Pipelines in DRDL files are parsed with most recent version of extjson spec, but still support legacy binary, regex, and dbPointer formats.
- BI-2218: No user-facing changes.
- BI-2231: No user-facing changes.
- BI-2016: No user-facing changes.
- BI-2099: No user-facing changes.
- BI-2228: No user-facing changes.
- BI-2214: No user-facing changes.
- BI-2229: No user-facing changes.
- BI-2227: No user-facing changes.
- BI-2223: No user-facing changes.
- BI-1813: No user-facing changes.
- BI-2173: No user-facing changes.
