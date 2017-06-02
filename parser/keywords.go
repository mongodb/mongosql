package parser

var (
	ADDDATE_BYTES            = []byte("adddate")
	ADMIN_BYTES              = []byte("admin")
	ALL_BYTES                = []byte("all")
	ALTER_BYTES              = []byte("alter")
	AND_BYTES                = []byte("and")
	ANY_BYTES                = []byte("any")
	AS_BYTES                 = []byte("as")
	ASC_BYTES                = []byte("asc")
	BEGIN_BYTES              = []byte("begin")
	BETWEEN_BYTES            = []byte("between")
	BINARY_BYTES             = []byte("binary")
	BINLOG_BYTES             = []byte("binlog")
	BOTH_BYTES               = []byte("both")
	BY_BYTES                 = []byte("by")
	CASE_BYTES               = []byte("case")
	CAST_BYTES               = []byte("cast")
	CHANNEL_BYTES            = []byte("channel")
	CHAR_BYTES               = []byte("char")
	CHARACTER_BYTES          = []byte("character")
	CHARSET_BYTES            = []byte("charset")
	CODE_BYTES               = []byte("code")
	COLLATE_BYTES            = []byte("collate")
	COLLATION_BYTES          = []byte("collation")
	COLUMNS_BYTES            = []byte("columns")
	COMMIT_BYTES             = []byte("commit")
	COMMITTED_BYTES          = []byte("committed")
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
	DISTINCT_BYTES           = []byte("distinct")
	DIV_BYTES                = []byte("div")
	DROP_BYTES               = []byte("drop")
	DUPLICATE_BYTES          = []byte("duplicate")
	ELSE_BYTES               = []byte("else")
	END_BYTES                = []byte("end")
	ENGINE_BYTES             = []byte("engine")
	ENGINES_BYTES            = []byte("engines")
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
	FUNCTION_BYTES           = []byte("function")
	GLOBAL_BYTES             = []byte("global")
	GRANTS_BYTES             = []byte("grants")
	GROUP_BYTES              = []byte("group")
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
	LOCK_BYTES               = []byte("lock")
	LOGS_BYTES               = []byte("logs")
	MASTER_BYTES             = []byte("master")
	MICROSECOND_BYTES        = []byte("microsecond")
	MINUS_BYTES              = []byte("minus")
	MINUTE_MICROSECOND_BYTES = []byte("minute_microsecond")
	MINUTE_SECOND_BYTES      = []byte("minute_second")
	MINUTE_BYTES             = []byte("minute")
	MOD_BYTES                = []byte("mod")
	MODE_BYTES               = []byte("mode")
	MONTH_BYTES              = []byte("month")
	MUTEX_BYTES              = []byte("mutex")
	NAMES_BYTES              = []byte("names")
	NATURAL_BYTES            = []byte("natural")
	NCHAR_BYTES              = []byte("nchar")
	NOT_BYTES                = []byte("not")
	NULL_BYTES               = []byte("null")
	NUMBER_BYTES             = []byte("number")
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
	PRIVILEGES_BYTES         = []byte("privileges")
	PROCEDURE_BYTES          = []byte("procedure")
	PROCESSLIST_BYTES        = []byte("processlist")
	PROFILE_BYTES            = []byte("profile")
	PROFILES_BYTES           = []byte("profiles")
	PROXY_BYTES              = []byte("proxy")
	QUARTER_BYTES            = []byte("quarter")
	QUERY_BYTES              = []byte("query")
	READ_BYTES               = []byte("read")
	REGEXP_BYTES             = []byte("regexp")
	RELAYLOG_BYTES           = []byte("relaylog")
	RENAME_BYTES             = []byte("rename")
	REPEATABLE_BYTES         = []byte("repeatable")
	REPLACE_BYTES            = []byte("replace")
	RIGHT_BYTES              = []byte("right")
	ROLLBACK_BYTES           = []byte("rollback")
	ROW_BYTES                = []byte("row")
	SCHEMA_BYTES             = []byte("schema")
	SCHEMAS_BYTES            = []byte("schemas")
	SECOND_MICROSECOND_BYTES = []byte("second_microsecond")
	SECOND_BYTES             = []byte("second")
	SELECT_BYTES             = []byte("select")
	SERIALIZABLE_BYTES       = []byte("serializable")
	SESSION_BYTES            = []byte("session")
	SET_BYTES                = []byte("set")
	SHARE_BYTES              = []byte("share")
	SHOW_BYTES               = []byte("show")
	SIGNED_BYTES             = []byte("signed")
	SLAVE_BYTES              = []byte("slave")
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
	T_BYTES                  = []byte("t")
	TABLE_BYTES              = []byte("table")
	TABLES_BYTES             = []byte("tables")
	THEN_BYTES               = []byte("then")
	TIME_BYTES               = []byte("time")
	TIMESTAMP_BYTES          = []byte("timestamp")
	TIMESTAMPADD_BYTES       = []byte("timestampadd")
	TIMESTAMPDIFF_BYTES      = []byte("timestampdiff")
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
	UNSIGNED_BYTES           = []byte("unsigned")
	UPDATE_BYTES             = []byte("update")
	USE_BYTES                = []byte("use")
	USER_BYTES               = []byte("user")
	USING_BYTES              = []byte("using")
	UTC_TIMESTAMP_BYTES      = []byte("utc_timestamp")
	VALUES_BYTES             = []byte("values")
	VARIABLES_BYTES          = []byte("variables")
	VIEW_BYTES               = []byte("view")
	WARNINGS_BYTES           = []byte("warnings")
	WEEK_BYTES               = []byte("week")
	WHEN_BYTES               = []byte("when")
	WHERE_BYTES              = []byte("where")
	WRITE_BYTES              = []byte("write")
	XOR_BYTES                = []byte("xor")
	YEAR_MONTH_BYTES         = []byte("year_month")
	YEAR_BYTES               = []byte("year")
)

var keywords = map[string]int{
	"adddate":            ADDDATE,
	"all":                ALL,
	"and":                AND,
	"any":                ANY,
	"as":                 AS,
	"asc":                ASC,
	"between":            BETWEEN,
	"binary":             BINARY,
	"binlog":             BINLOG,
	"both":               BOTH,
	"by":                 BY,
	"case":               CASE,
	"cast":               CAST,
	"channel":            CHANNEL,
	"char":               CHAR,
	"character":          CHARACTER,
	"charset":            CHARSET,
	"code":               CODE,
	"collate":            COLLATE,
	"collation":          COLLATION,
	"columns":            COLUMNS,
	"committed":          COMMITTED,
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
	"distinct":           DISTINCT,
	"div":                IDIV,
	"else":               ELSE,
	"end":                END,
	"engine":             ENGINE,
	"engines":            ENGINES,
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
	"fn":                 FN,
	"for":                FOR,
	"force":              FORCE,
	"format":             FORMAT,
	"from":               FROM,
	"full":               FULL,
	"function":           FUNCTION,
	"global":             GLOBAL,
	"grants":             GRANTS,
	"group":              GROUP,
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
	"integer":            INTEGER,
	"intersect":          INTERSECT,
	"interval":           INTERVAL,
	"is":                 IS,
	"isolation":          ISOLATION,
	"join":               JOIN,
	"json":               JSON,
	"keys":               KEYS,
	"kill":               KILL,
	"leading":            LEADING,
	"left":               LEFT,
	"level":              LEVEL,
	"like":               LIKE,
	"limit":              LIMIT,
	"lock":               LOCK,
	"logs":               LOGS,
	"master":             MASTER,
	"microsecond":        MICROSECOND,
	"minus":              MINUS,
	"minute_microsecond": MINUTE_MICROSECOND,
	"minute_second":      MINUTE_SECOND,
	"minute":             MINUTE,
	"mod":                MOD,
	"month":              MONTH,
	"mutex":              MUTEX,
	"names":              NAMES,
	"natural":            NATURAL,
	"nchar":              NCHAR,
	"not":                NOT,
	"null":               NULL,
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
	"privileges":         PRIVILEGES,
	"procedure":          PROCEDURE,
	"processlist":        PROCESSLIST,
	"profile":            PROFILE,
	"profiles":           PROFILES,
	"proxy":              PROXY,
	"quarter":            QUARTER,
	"query":              QUERY,
	"read":               READ,
	"regexp":             REGEXP,
	"relaylog":           RELAYLOG,
	"repeatable":         REPEATABLE,
	"right":              RIGHT,
	"row":                ROW,
	"schema":             SCHEMA,
	"schemas":            SCHEMAS,
	"second_microsecond": SECOND_MICROSECOND,
	"second":             SECOND,
	"select":             SELECT,
	"serializable":       SERIALIZABLE,
	"session":            SESSION,
	"set":                SET,
	"show":               SHOW,
	"signed":             SIGNED,
	"slave":              SLAVE,
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
	"table":              TABLE,
	"tables":             TABLES,
	"then":               THEN,
	"time":               TIME,
	"timestamp":          TIMESTAMP,
	"timestampadd":       TIMESTAMPADD,
	"timestampdiff":      TIMESTAMPDIFF,
	"traditional":        TRADITIONAL,
	"trailing":           TRAILING,
	"transaction":        TRANSACTION,
	"trigger":            TRIGGER,
	"triggers":           TRIGGERS,
	"trim":               TRIM,
	"true":               TRUE,
	"uncommitted":        UNCOMMITTED,
	"union":              UNION,
	"unknown":            UNKNOWN,
	"unsigned":           UNSIGNED,
	"update":             UPDATE,
	"use":                USE,
	"user":               USER,
	"utc_timestamp":      UTC_TIMESTAMP,
	"values":             VALUES,
	"variables":          VARIABLES,
	"view":               VIEW,
	"warnings":           WARNINGS,
	"week":               WEEK,
	"when":               WHEN,
	"where":              WHERE,
	"write":              WRITE,
	"xor":                XOR,
	"year_month":         YEAR_MONTH,
	"year":               YEAR,
}
