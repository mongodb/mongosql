// nolint: golint
package parser

var (
	ADD_BYTES                = []byte("add")
	ADDDATE_BYTES            = []byte("adddate")
	ADMIN_BYTES              = []byte("admin")
	ALL_BYTES                = []byte("all")
	ALTER_BYTES              = []byte("alter")
	AND_BYTES                = []byte("and")
	ANY_BYTES                = []byte("any")
	AS_BYTES                 = []byte("as")
	ASC_BYTES                = []byte("asc")
	AUTO_INCREMENT_BYTES     = []byte("auto_increment")
	BEGIN_BYTES              = []byte("begin")
	BETWEEN_BYTES            = []byte("between")
	BIGINT_BYTES             = []byte("bigint")
	BINARY_BYTES             = []byte("binary")
	BINLOG_BYTES             = []byte("binlog")
	BIT_BYTES                = []byte("bit")
	BLOB_BYTES               = []byte("blob")
	BOOL_BYTES               = []byte("bool")
	BOOLEAN_BYTES            = []byte("boolean")
	BOTH_BYTES               = []byte("both")
	BTREE_BYTES              = []byte("btree")
	BY_BYTES                 = []byte("by")
	CASCADE_BYTES            = []byte("cascade")
	CASE_BYTES               = []byte("case")
	CAST_BYTES               = []byte("cast")
	CHANGE_BYTES             = []byte("change")
	CHANNEL_BYTES            = []byte("channel")
	CHAR_BYTES               = []byte("char")
	CHARACTER_BYTES          = []byte("character")
	CHARSET_BYTES            = []byte("charset")
	CODE_BYTES               = []byte("code")
	COLLATE_BYTES            = []byte("collate")
	COLLATION_BYTES          = []byte("collation")
	COLUMN_BYTES             = []byte("column")
	COLUMNS_BYTES            = []byte("columns")
	COMMIT_BYTES             = []byte("commit")
	COMMITTED_BYTES          = []byte("committed")
	COMMENT_BYTES            = []byte("comment")
	CONNECTION_BYTES         = []byte("connection")
	CONVERT_BYTES            = []byte("convert")
	COUNT_BYTES              = []byte("count")
	CREATE_BYTES             = []byte("create")
	CROSS_BYTES              = []byte("cross")
	CURRENT_DATE_BYTES       = []byte("current_date")
	CURRENT_TIMESTAMP_BYTES  = []byte("current_timestamp")
	D_BYTES                  = []byte("d")
	DATABASE_BYTES           = []byte("database")
	DATABASES_BYTES          = []byte("databases")
	DATE_ADD_BYTES           = []byte("date_add")
	DATE_SUB_BYTES           = []byte("date_sub")
	DATE_BYTES               = []byte("date")
	DATETIME_BYTES           = []byte("datetime")
	DAY_HOUR_BYTES           = []byte("day_hour")
	DAY_MICROSECOND_BYTES    = []byte("day_microsecond")
	DAY_MINUTE_BYTES         = []byte("day_minute")
	DAY_SECOND_BYTES         = []byte("day_second")
	DAY_BYTES                = []byte("day")
	DECIMAL_BYTES            = []byte("decimal")
	DEFAULT_BYTES            = []byte("default")
	DELETE_BYTES             = []byte("delete")
	DESC_BYTES               = []byte("desc")
	DESCRIBE_BYTES           = []byte("describe")
	DISABLE_BYTES            = []byte("disable")
	DISTINCT_BYTES           = []byte("distinct")
	DIV_BYTES                = []byte("div")
	DOUBLE_BYTES             = []byte("double")
	DROP_BYTES               = []byte("drop")
	DUPLICATE_BYTES          = []byte("duplicate")
	ELSE_BYTES               = []byte("else")
	ENABLE_BYTES             = []byte("enable")
	END_BYTES                = []byte("end")
	ENGINE_BYTES             = []byte("engine")
	ENGINES_BYTES            = []byte("engines")
	ENUM_BYTES               = []byte("enum")
	ERRORS_BYTES             = []byte("errors")
	ESCAPE_BYTES             = []byte("escape")
	EVENT_BYTES              = []byte("event")
	EVENTS_BYTES             = []byte("events")
	EXCEPT_BYTES             = []byte("except")
	EXISTS_BYTES             = []byte("exists")
	EXPLAIN_BYTES            = []byte("explain")
	EXTENDED_BYTES           = []byte("extended")
	EXTRACT_BYTES            = []byte("extract")
	FALSE_BYTES              = []byte("false")
	FLOAT_BYTES              = []byte("float")
	FN_BYTES                 = []byte("fn")
	FOR_BYTES                = []byte("for")
	FORCE_BYTES              = []byte("force")
	FORMAT_BYTES             = []byte("format")
	FROM_BYTES               = []byte("from")
	FULL_BYTES               = []byte("full")
	FULLTEXT_BYTES           = []byte("fulltext")
	FUNCTION_BYTES           = []byte("function")
	GLOBAL_BYTES             = []byte("global")
	GRANTS_BYTES             = []byte("grants")
	GROUP_BYTES              = []byte("group")
	GROUP_CONCAT_BYTES       = []byte("group_concat")
	HASH_BYTES               = []byte("hash")
	HAVING_BYTES             = []byte("having")
	HOSTS_BYTES              = []byte("hosts")
	HOUR_MICROSECOND_BYTES   = []byte("hour_microsecond")
	HOUR_MINUTE_BYTES        = []byte("hour_minute")
	HOUR_SECOND_BYTES        = []byte("hour_second")
	HOUR_BYTES               = []byte("hour")
	IF_BYTES                 = []byte("if")
	IGNORE_BYTES             = []byte("ignore")
	IN_BYTES                 = []byte("in")
	INDEX_BYTES              = []byte("index")
	INDEXES_BYTES            = []byte("indexes")
	INNER_BYTES              = []byte("inner")
	INSERT_BYTES             = []byte("insert")
	INT_BYTES                = []byte("int")
	INTEGER_BYTES            = []byte("integer")
	INTERSECT_BYTES          = []byte("intersect")
	INTERVAL_BYTES           = []byte("interval")
	INTO_BYTES               = []byte("into")
	IS_BYTES                 = []byte("is")
	ISOLATION_BYTES          = []byte("isolation")
	JOIN_BYTES               = []byte("join")
	JSON_BYTES               = []byte("json")
	KEY_BYTES                = []byte("key")
	KEYS_BYTES               = []byte("keys")
	KILL_BYTES               = []byte("kill")
	LEADING_BYTES            = []byte("leading")
	LEFT_BYTES               = []byte("left")
	LEVEL_BYTES              = []byte("level")
	LIKE_BYTES               = []byte("like")
	LIMIT_BYTES              = []byte("limit")
	LOCAL_BYTES              = []byte("local")
	LOCK_BYTES               = []byte("lock")
	LOGS_BYTES               = []byte("logs")
	LONGTEXT_BYTES           = []byte("longtext")
	LOW_PRIORITY_BYTES       = []byte("low_priority")
	MASTER_BYTES             = []byte("master")
	MEDIUMBLOB_BYTES         = []byte("mediumblob")
	MEDIUMTEXT_BYTES         = []byte("mediumtext")
	MICROSECOND_BYTES        = []byte("microsecond")
	MINUS_BYTES              = []byte("minus")
	MINUTE_MICROSECOND_BYTES = []byte("minute_microsecond")
	MINUTE_SECOND_BYTES      = []byte("minute_second")
	MINUTE_BYTES             = []byte("minute")
	MOD_BYTES                = []byte("mod")
	MODE_BYTES               = []byte("mode")
	MODIFY_BYTES             = []byte("modify")
	MONTH_BYTES              = []byte("month")
	MUTEX_BYTES              = []byte("mutex")
	NAMES_BYTES              = []byte("names")
	NATURAL_BYTES            = []byte("natural")
	NCHAR_BYTES              = []byte("nchar")
	NOT_BYTES                = []byte("not")
	NULL_BYTES               = []byte("null")
	NUMBER_BYTES             = []byte("number")
	NUMERIC_BYTES            = []byte("numeric")
	OBJECT_ID_BYTES          = []byte("objectid")
	OFF_BYTES                = []byte("off")
	OFFSET_BYTES             = []byte("offset")
	OJ_BYTES                 = []byte("oj")
	ON_BYTES                 = []byte("on")
	ONLY_BYTES               = []byte("only")
	OPEN_BYTES               = []byte("open")
	OR_BYTES                 = []byte("or")
	ORDER_BYTES              = []byte("order")
	OUTER_BYTES              = []byte("outer")
	PARTITIONS_BYTES         = []byte("partitions")
	PLUGINS_BYTES            = []byte("plugins")
	PRECISION_BYTES          = []byte("precision")
	PRIMARY_BYTES            = []byte("primary")
	PRIVILEGES_BYTES         = []byte("privileges")
	PROCEDURE_BYTES          = []byte("procedure")
	PROCESSLIST_BYTES        = []byte("processlist")
	PROFILE_BYTES            = []byte("profile")
	PROFILES_BYTES           = []byte("profiles")
	PROXY_BYTES              = []byte("proxy")
	QUARTER_BYTES            = []byte("quarter")
	QUERY_BYTES              = []byte("query")
	READ_BYTES               = []byte("read")
	RECURSIVE_BYTES          = []byte("recursive")
	REGEXP_BYTES             = []byte("regexp")
	RELAYLOG_BYTES           = []byte("relaylog")
	RENAME_BYTES             = []byte("rename")
	REPEATABLE_BYTES         = []byte("repeatable")
	REPLACE_BYTES            = []byte("replace")
	RESTRICT_BYTES           = []byte("restrict")
	RIGHT_BYTES              = []byte("right")
	ROLLBACK_BYTES           = []byte("rollback")
	ROW_BYTES                = []byte("row")
	SAMPLE_BYTES             = []byte("sample")
	SCHEMA_BYTES             = []byte("schema")
	SCHEMAS_BYTES            = []byte("schemas")
	SECOND_MICROSECOND_BYTES = []byte("second_microsecond")
	SECOND_BYTES             = []byte("second")
	SELECT_BYTES             = []byte("select")
	SERIAL_BYTES             = []byte("serial")
	SERIALIZABLE_BYTES       = []byte("serializable")
	SESSION_BYTES            = []byte("session")
	SET_BYTES                = []byte("set")
	SHARE_BYTES              = []byte("share")
	SHOW_BYTES               = []byte("show")
	SIGNED_BYTES             = []byte("signed")
	SLAVE_BYTES              = []byte("slave")
	SMALLINT_BYTES           = []byte("smallint")
	SOME_BYTES               = []byte("some")
	SQL_BIGINT_BYTES         = []byte("sql_bigint")
	SQL_DATE_BYTES           = []byte("sql_date")
	SQL_DOUBLE_BYTES         = []byte("sql_double")
	SQL_TIMESTAMP_BYTES      = []byte("sql_timestamp")
	SQL_TSI_DAY_BYTES        = []byte("sql_tsi_day")
	SQL_TSI_HOUR_BYTES       = []byte("sql_tsi_hour")
	SQL_TSI_MINUTE_BYTES     = []byte("sql_tsi_minute")
	SQL_TSI_MONTH_BYTES      = []byte("sql_tsi_month")
	SQL_TSI_QUARTER_BYTES    = []byte("sql_tsi_quarter")
	SQL_TSI_SECOND_BYTES     = []byte("sql_tsi_second")
	SQL_TSI_WEEK_BYTES       = []byte("sql_tsi_week")
	SQL_TSI_YEAR_BYTES       = []byte("sql_tsi_year")
	SQL_VARCHAR_BYTES        = []byte("sql_varchar")
	STATUS_BYTES             = []byte("status")
	STORAGE_BYTES            = []byte("storage")
	STRAIGHT_JOIN_BYTES      = []byte("straight_join")
	SUBDATE_BYTES            = []byte("subdate")
	SUBSTR_BYTES             = []byte("substr")
	SUBSTRING_BYTES          = []byte("substring")
	T_BYTES                  = []byte("t")
	TABLE_BYTES              = []byte("table")
	TABLES_BYTES             = []byte("tables")
	TEMPORARY_BYTES          = []byte("temporary")
	TEXT_BYTES               = []byte("text")
	THEN_BYTES               = []byte("then")
	TIME_BYTES               = []byte("time")
	TIMESTAMP_BYTES          = []byte("timestamp")
	TIMESTAMPADD_BYTES       = []byte("timestampadd")
	TIMESTAMPDIFF_BYTES      = []byte("timestampdiff")
	TINYINT_BYTES            = []byte("tinyint")
	TINYTEXT_BYTES           = []byte("tinytext")
	TO_BYTES                 = []byte("to")
	TRADITIONAL_BYTES        = []byte("traditional")
	TRAILING_BYTES           = []byte("trailing")
	TRANSACTION_BYTES        = []byte("transaction")
	TRIGGER_BYTES            = []byte("trigger")
	TRIGGERS_BYTES           = []byte("triggers")
	TRIM_BYTES               = []byte("trim")
	TRUE_BYTES               = []byte("true")
	TS_BYTES                 = []byte("ts")
	UNCOMMITTED_BYTES        = []byte("uncommitted")
	UNION_BYTES              = []byte("union")
	UNIQUE_BYTES             = []byte("unique")
	UNKNOWN_BYTES            = []byte("unknown")
	UNLOCK_BYTES             = []byte("unlock")
	UNSIGNED_BYTES           = []byte("unsigned")
	UPDATE_BYTES             = []byte("update")
	USE_BYTES                = []byte("use")
	USER_BYTES               = []byte("user")
	USING_BYTES              = []byte("using")
	UTC_DATE_BYTES           = []byte("utc_date")
	UTC_TIMESTAMP_BYTES      = []byte("utc_timestamp")
	VALUE_BYTES              = []byte("value")
	VALUES_BYTES             = []byte("values")
	VARIABLES_BYTES          = []byte("variables")
	VARCHAR_BYTES            = []byte("varchar")
	VIEW_BYTES               = []byte("view")
	WARNINGS_BYTES           = []byte("warnings")
	WEEK_BYTES               = []byte("week")
	WHEN_BYTES               = []byte("when")
	WHERE_BYTES              = []byte("where")
	WITH_BYTES               = []byte("with")
	WRITE_BYTES              = []byte("write")
	XOR_BYTES                = []byte("xor")
	YEAR_MONTH_BYTES         = []byte("year_month")
	YEAR_BYTES               = []byte("year")
)

var keywords = map[string]int{
	"add":                ADD,
	"adddate":            ADDDATE,
	"all":                ALL,
	"alter":              ALTER,
	"and":                AND,
	"any":                ANY,
	"as":                 AS,
	"asc":                ASC,
	"auto_increment":     AUTO_INCREMENT,
	"between":            BETWEEN,
	"bigint":             BIGINT,
	"binary":             BINARY,
	"binlog":             BINLOG,
	"bit":                BIT,
	"blob":               BLOB,
	"bool":               BOOL,
	"boolean":            BOOLEAN,
	"both":               BOTH,
	"btree":              BTREE,
	"by":                 BY,
	"cascade":            CASCADE,
	"case":               CASE,
	"cast":               CAST,
	"change":             CHANGE,
	"channel":            CHANNEL,
	"char":               CHAR,
	"character":          CHARACTER,
	"charset":            CHARSET,
	"code":               CODE,
	"collate":            COLLATE,
	"collation":          COLLATION,
	"column":             COLUMN,
	"columns":            COLUMNS,
	"committed":          COMMITTED,
	"comment":            COMMENT_KWD,
	"connection":         CONNECTION,
	"convert":            CONVERT,
	"count":              COUNT,
	"create":             CREATE,
	"cross":              CROSS,
	"current_date":       CURRENT_DATE,
	"current_timestamp":  CURRENT_TIMESTAMP,
	"database":           DATABASE,
	"databases":          DATABASES,
	"date_add":           DATE_ADD,
	"date_sub":           DATE_SUB,
	"date":               DATE,
	"datetime":           DATETIME,
	"day_hour":           DAY_HOUR,
	"day_microsecond":    DAY_MICROSECOND,
	"day_minute":         DAY_MINUTE,
	"day_second":         DAY_SECOND,
	"day":                DAY,
	"decimal":            DECIMAL,
	"default":            DEFAULT,
	"desc":               DESC,
	"describe":           DESCRIBE,
	"disable":            DISABLE,
	"distinct":           DISTINCT,
	"div":                IDIV,
	"double":             DOUBLE,
	"drop":               DROP,
	"dual":               DUAL,
	"else":               ELSE,
	"enable":             ENABLE,
	"end":                END,
	"engine":             ENGINE,
	"engines":            ENGINES,
	"enum":               ENUM,
	"errors":             ERRORS,
	"escape":             ESCAPE,
	"event":              EVENT,
	"events":             EVENTS,
	"except":             EXCEPT,
	"exists":             EXISTS,
	"explain":            EXPLAIN,
	"extended":           EXTENDED,
	"extract":            EXTRACT,
	"false":              FALSE,
	"fields":             COLUMNS, /* treat as SHOW COLUMNS */
	"float":              FLOAT,
	"flush":              FLUSH,
	"fn":                 FN,
	"for":                FOR,
	"force":              FORCE,
	"format":             FORMAT,
	"from":               FROM,
	"full":               FULL,
	"fulltext":           FULLTEXT,
	"function":           FUNCTION,
	"global":             GLOBAL,
	"grants":             GRANTS,
	"group":              GROUP,
	"group_concat":       GROUP_CONCAT,
	"hash":               HASH,
	"having":             HAVING,
	"hosts":              HOSTS,
	"hour_microsecond":   HOUR_MICROSECOND,
	"hour_minute":        HOUR_MINUTE,
	"hour_second":        HOUR_SECOND,
	"hour":               HOUR,
	"if":                 IF,
	"ignore":             IGNORE,
	"in":                 IN,
	"index":              INDEX,
	"indexes":            INDEXES,
	"inner":              INNER,
	"insert":             INSERT,
	"int":                INT,
	"integer":            INTEGER,
	"intersect":          INTERSECT,
	"interval":           INTERVAL,
	"into":               INTO,
	"is":                 IS,
	"isolation":          ISOLATION,
	"join":               JOIN,
	"json":               JSON,
	"key":                KEY,
	"keys":               KEYS,
	"kill":               KILL,
	"leading":            LEADING,
	"left":               LEFT,
	"level":              LEVEL,
	"like":               LIKE,
	"limit":              LIMIT,
	"local":              LOCAL,
	"lock":               LOCK,
	"logs":               LOGS,
	"longtext":           LONGTEXT,
	"low_priority":       LOW_PRIORITY,
	"master":             MASTER,
	"mediumblob":         MEDIUMBLOB,
	"mediumtext":         MEDIUMTEXT,
	"microsecond":        MICROSECOND,
	"minus":              MINUS,
	"minute_microsecond": MINUTE_MICROSECOND,
	"minute_second":      MINUTE_SECOND,
	"minute":             MINUTE,
	"mod":                MOD,
	"modify":             MODIFY,
	"month":              MONTH,
	"mutex":              MUTEX,
	"names":              NAMES,
	"natural":            NATURAL,
	"nchar":              NCHAR,
	"not":                NOT,
	"null":               NULL,
	"numeric":            NUMERIC,
	"objectid":           OBJECT_ID,
	"off":                OFF,
	"offset":             OFFSET,
	"oj":                 OJ,
	"on":                 ON,
	"only":               ONLY,
	"open":               OPEN,
	"or":                 OR,
	"order":              ORDER,
	"outer":              OUTER,
	"partitions":         PARTITIONS,
	"plugins":            PLUGINS,
	"precision":          PRECISION,
	"primary":            PRIMARY,
	"privileges":         PRIVILEGES,
	"procedure":          PROCEDURE,
	"processlist":        PROCESSLIST,
	"profile":            PROFILE,
	"profiles":           PROFILES,
	"proxy":              PROXY,
	"quarter":            QUARTER,
	"query":              QUERY,
	"read":               READ,
	"recursive":          RECURSIVE,
	"regexp":             REGEXP,
	"relaylog":           RELAYLOG,
	"rename":             RENAME,
	"repeatable":         REPEATABLE,
	"restrict":           RESTRICT,
	"right":              RIGHT,
	"rlike":              RLIKE,
	"row":                ROW,
	"sample":             SAMPLE,
	"schema":             SCHEMA,
	"schemas":            SCHEMAS,
	"second_microsecond": SECOND_MICROSECOND,
	"second":             SECOND,
	"select":             SELECT,
	"serial":             SERIAL,
	"serializable":       SERIALIZABLE,
	"session":            SESSION,
	"set":                SET,
	"separator":          SEPARATOR,
	"show":               SHOW,
	"signed":             SIGNED,
	"slave":              SLAVE,
	"smallint":           SMALLINT,
	"some":               SOME,
	"sql_bigint":         SQL_BIGINT,
	"sql_date":           SQL_DATE,
	"sql_double":         SQL_DOUBLE,
	"sql_timestamp":      SQL_TIMESTAMP,
	"sql_tsi_day":        SQL_TSI_DAY,
	"sql_tsi_hour":       SQL_TSI_HOUR,
	"sql_tsi_minute":     SQL_TSI_MINUTE,
	"sql_tsi_month":      SQL_TSI_MONTH,
	"sql_tsi_quarter":    SQL_TSI_QUARTER,
	"sql_tsi_second":     SQL_TSI_SECOND,
	"sql_tsi_week":       SQL_TSI_WEEK,
	"sql_tsi_year":       SQL_TSI_YEAR,
	"sql_varchar":        SQL_VARCHAR,
	"status":             STATUS,
	"storage":            STORAGE,
	"straight_join":      STRAIGHT_JOIN,
	"subdate":            SUBDATE,
	"substr":             SUBSTR,
	"substring":          SUBSTRING,
	"table":              TABLE,
	"tables":             TABLES,
	"temporary":          TEMPORARY,
	"text":               TEXT,
	"then":               THEN,
	"time":               TIME,
	"timestamp":          TIMESTAMP,
	"timestampadd":       TIMESTAMPADD,
	"timestampdiff":      TIMESTAMPDIFF,
	"tinyint":            TINYINT,
	"tinytext":           TINYTEXT,
	"to":                 TO,
	"traditional":        TRADITIONAL,
	"trailing":           TRAILING,
	"transaction":        TRANSACTION,
	"trigger":            TRIGGER,
	"triggers":           TRIGGERS,
	"trim":               TRIM,
	"true":               TRUE,
	"uncommitted":        UNCOMMITTED,
	"union":              UNION,
	"unlock":             UNLOCK,
	"unique":             UNIQUE,
	"unknown":            UNKNOWN,
	"unsigned":           UNSIGNED,
	"update":             UPDATE,
	"use":                USE,
	"user":               USER,
	"using":              USING,
	"utc_timestamp":      UTC_TIMESTAMP,
	"utc_date":           UTC_DATE,
	"value":              VALUE,
	"values":             VALUES,
	"varchar":            VARCHAR,
	"variables":          VARIABLES,
	"view":               VIEW,
	"warnings":           WARNINGS,
	"week":               WEEK,
	"when":               WHEN,
	"where":              WHERE,
	"with":               WITH,
	"write":              WRITE,
	"xor":                XOR,
	"year_month":         YEAR_MONTH,
	"year":               YEAR,
}

var keywordAsID = map[string]struct{}{
	string(ANY_BYTES):             {},
	string(BINLOG_BYTES):          {},
	string(AUTO_INCREMENT_BYTES):  {},
	string(BIT_BYTES):             {},
	string(BLOB_BYTES):            {},
	string(BOOL_BYTES):            {},
	string(BTREE_BYTES):           {},
	string(CHANNEL_BYTES):         {},
	string(CHARSET_BYTES):         {},
	string(CODE_BYTES):            {},
	string(COLLATION_BYTES):       {},
	string(COLUMNS_BYTES):         {},
	string(COMMENT_BYTES):         {},
	string(COMMITTED_BYTES):       {},
	string(CONNECTION_BYTES):      {},
	string(COUNT_BYTES):           {},
	string(DATE_BYTES):            {},
	string(DATETIME_BYTES):        {},
	string(DAY_BYTES):             {},
	string(DECIMAL_BYTES):         {},
	string(DISABLE_BYTES):         {},
	string(ENABLE_BYTES):          {},
	string(ENGINE_BYTES):          {},
	string(ENGINES_BYTES):         {},
	string(ENUM_BYTES):            {},
	string(ERRORS_BYTES):          {},
	string(EVENT_BYTES):           {},
	string(EVENTS_BYTES):          {},
	string(EXTENDED_BYTES):        {},
	string(FLOAT_BYTES):           {},
	string(FORMAT_BYTES):          {},
	string(FULL_BYTES):            {},
	string(FUNCTION_BYTES):        {},
	string(GRANTS_BYTES):          {},
	string(HASH_BYTES):            {},
	string(HOSTS_BYTES):           {},
	string(HOUR_BYTES):            {},
	string(INDEXES_BYTES):         {},
	string(ISOLATION_BYTES):       {},
	string(JSON_BYTES):            {},
	string(LEVEL_BYTES):           {},
	string(LOCAL_BYTES):           {},
	string(LOGS_BYTES):            {},
	string(LONGTEXT_BYTES):        {},
	string(MASTER_BYTES):          {},
	string(MEDIUMBLOB_BYTES):      {},
	string(MEDIUMTEXT_BYTES):      {},
	string(MICROSECOND_BYTES):     {},
	string(MINUTE_BYTES):          {},
	string(MONTH_BYTES):           {},
	string(MUTEX_BYTES):           {},
	string(NAMES_BYTES):           {},
	string(NCHAR_BYTES):           {},
	string(NUMBER_BYTES):          {},
	string(OFFSET_BYTES):          {},
	string(OBJECT_ID_BYTES):       {},
	string(ONLY_BYTES):            {},
	string(OPEN_BYTES):            {},
	string(PARTITIONS_BYTES):      {},
	string(PLUGINS_BYTES):         {},
	string(PRIVILEGES_BYTES):      {},
	string(PROCESSLIST_BYTES):     {},
	string(PROFILE_BYTES):         {},
	string(PROFILES_BYTES):        {},
	string(PROXY_BYTES):           {},
	string(QUARTER_BYTES):         {},
	string(QUERY_BYTES):           {},
	string(RELAYLOG_BYTES):        {},
	string(REPEATABLE_BYTES):      {},
	string(ROW_BYTES):             {},
	string(SECOND_BYTES):          {},
	string(SERIAL_BYTES):          {},
	string(SERIALIZABLE_BYTES):    {},
	string(SIGNED_BYTES):          {},
	string(SLAVE_BYTES):           {},
	string(SMALLINT_BYTES):        {},
	string(SOME_BYTES):            {},
	string(SQL_TSI_DAY_BYTES):     {},
	string(SQL_TSI_HOUR_BYTES):    {},
	string(SQL_TSI_MINUTE_BYTES):  {},
	string(SQL_TSI_MONTH_BYTES):   {},
	string(SQL_TSI_QUARTER_BYTES): {},
	string(SQL_TSI_SECOND_BYTES):  {},
	string(SQL_TSI_WEEK_BYTES):    {},
	string(SQL_TSI_YEAR_BYTES):    {},
	string(STATUS_BYTES):          {},
	string(STORAGE_BYTES):         {},
	string(TABLES_BYTES):          {},
	string(TEMPORARY_BYTES):       {},
	string(TIME_BYTES):            {},
	string(TIMESTAMP_BYTES):       {},
	string(TIMESTAMPADD_BYTES):    {},
	string(TIMESTAMPDIFF_BYTES):   {},
	string(TINYINT_BYTES):         {},
	string(TRANSACTION_BYTES):     {},
	string(TRIGGERS_BYTES):        {},
	string(UNCOMMITTED_BYTES):     {},
	string(UNKNOWN_BYTES):         {},
	string(USER_BYTES):            {},
	string(VALUE_BYTES):           {},
	string(VARIABLES_BYTES):       {},
	string(VIEW_BYTES):            {},
	string(WARNINGS_BYTES):        {},
	string(WEEK_BYTES):            {},
	string(YEAR_BYTES):            {},
}
