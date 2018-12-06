# System Variables #

This file contains a description of the variables present in the MongoDB Connector for Business Intelligence (BIC).

## BIC System Variables ##

This is a description of the [BIC](https://www.mongodb.com)'s exposed system variables.

|Name|Default Value|Possible Values|Description|
|:-|:-|:-|:-|
|enable_table_alterations|false|boolean|If enabled, the server will process any ALTER TABLE statements issued.|
|log_level|0|-1, 0, 1, 2|This variable sets the logging level for the MongoDB Connector for Business Intelligence. The following table shows the permitted values.<br>-1	No logging<br>0	Log only messages that notify the user of basic mongosqld events and state changes.<br>1	Log only messages that would be useful/understandable for mongosqld admins.<br>2	Log only messagges that target primarily at MongoDB developers, TSEs, etc.|
|metrics_backend|off|log, stitch, off|Determines where the BIC will store metrics related to queries issued.|
|mongodb_max_server_size|0|integer|The maximum size in bytes of memory that will be allocated for evaluating any query on the BIC. If the value is 0 (default) no limit is imposed.|
|mongodb_max_connection_size|0|integer|The maximum size in bytes of memory that will be allocated for evaluating any query on any given client connection. If the value is 0 (default) no limit is imposed.|
|mongodb_max_stage_size|0|integer|The maximum size in bytes of memory that will be allocated for evaluating any query in any given query evaluation stage. If the value is 0 (default) no limit is imposed.|
|mongodb_max_varchar_length|0|integer|The maximum size of data returned in columns with a varchar data type. If the value is 0 (default) no limit is imposed.|
|mongodb_version_compatibility|N/A|varchar|For mixed cluster MongoDB installations, the minimum version of any process within the cluster.|
|mongodb_git_version|N/A|varchar|The git version of MongoDB the BIC is connected to for a given client connection.|
|mongodb_version|N/A|varchar|The version of MongoDB the BIC is connected to for a given client connection.|
|mongosqld_version|N/A|varchar|The BIC version.|
|full_pushdown_exec_mode|false|boolean|If enabled, a query error will be returned for any query that isn't fully pushed down to MongoDB.|
|max_nested_table_depth|50|integer|The maximum number of unique MongoDB nested array field paths (when any non-json mapping mode is used) that mongosqld will map to a relational table for any given collection.|
|max_num_columns_per_table|1000|integer|The maximum number of unique MongoDB fields that mongosqld will map to relational columns for any given collection.|
|optimize_cross_joins|true|boolean|If enabled, cross joins are optimized to inner joins when possible.|
|optimize_evaluations|true|boolean|If enabled, constant-folding is performed.|
|optimize_filtering|true|boolean|If enabled, predicates in WHERE clauses are moved as close as possible to the MongoDB data source they operate on.|
|optimize_inner_joins|true|boolean|If enabled, inner joins are reordered for more optimal query execution.|
|optimize_self_joins|true|boolean|If enabled, when any non-json mapping mode (schema_mapping_mode) is used, it will cause parent/progeny join queries to be evluated more optimally.|
|optimize_view_sampling|true|boolean|If enabled, during sampling of a MongoDB view, $sample will be moved ahead of non-cardinality altering pipeline stages for the view.|
|polymorphic_type_conversion_mode|off|fast, safe, off|Determines how fields with multiple types in MongoDB (e.g. a field "count" might be present as a string in one document, and as an integer in another document) are translated for query evaluation. When set to "off", some queries - e.g. "select count + 2" - will fail if not explicitly cast.<br>When set to "fast", the BIC will appropriately cast any such fields ("count" in this case, to integer) if during sampling, it sampled both documents.<br>When set to "safe", the BIC will unconditionally cast such fields as appropriate.|
|pushdown|true|boolean|If enabled, queries are translated to MongoDB's native aggregation language.|
|sample_refresh_interval_secs|0|integer|This variable defines the global policy for controlling how frequently the BIC schema is updated. If the value is 0 (the default), there is no re-sampling after the BIC is started.|
|sample_size|1000|integer|This variable defines the global policy for controlling how many documents the BIC will sample in generating its schema. If the value is 0, the BIC will perform a collection scan across all namespaces.|
|schema_mapping_mode|lattice|lattice, majority|This variable determines how the MongoDB schema is transformed into a relational schema.|
|type_conversion_mode|mongosql|mysql, mongosql|This variable determines the semantics with which the BIC performs type conversions - for functions like CAST.|

## MySQL Status Variables ##

These are variables that indicate the current status of various parameters in the BIC. For descriptions, please see https://dev.mysql.com/doc/refman/8.0/en/server-status-variable-reference.html

## MySQL System Variables (supported-only) ##

The additional variables below are MySQL system variables that are currently supported by the BIC. For descriptions, please see https://dev.mysql.com/doc/refman/8.0/en/server-system-variable-reference.html

|Name|Default Value|Possible Values|Description|
|:-|:-|:-|:-|
|autocommit|x|x|x|
|character_set_client|x|x|x|
|character_set_connection|x|x|x|
|character_set_database|x|x|x|
|character_set_results|x|x|x|
|collation_connection|x|x|x|
|collation_database|x|x|x|
|collation_server|x|x|x|
|group_concat_max_len|x|x|x|
|interactive_timeout|x|x|x|
|max_allowed_packet|x|x|x|
|max_connections|x|x|x|
|max_execution_time|x|x|x|
|socket|x|x|x|
|sql_auto_is_null|x|x|x|
|sql_select_limit|x|x|x|
|version|x|x|x|
|version_comment|x|x|x|
|wait_timeout|x|x|x|

## MySQL System Variables (unsupported-only) ##

These are variables from MySQL that aren't yet currently implemented in the BIC. They are contained in the [stub_variables.go](https://github.com/deafgoat/sqlproxy/blob/1868/internal/variable/stub_variables.go) file.
