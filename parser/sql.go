//line sql.y:5
package parser

import __yyfmt__ "fmt"

//line sql.y:7
import "bytes"

func SetParseTree(yylex interface{}, stmt Statement) {
	yylex.(*Tokenizer).ParseTree = stmt
}

func SetAllowComments(yylex interface{}, allow bool) {
	yylex.(*Tokenizer).AllowComments = allow
}

func ForceEOF(yylex interface{}) {
	yylex.(*Tokenizer).ForceEOF = true
}

var (
	SHARE             = []byte("share")
	MODE              = []byte("mode")
	IF_BYTES          = []byte("if")
	VALUES_BYTES      = []byte("values")
	RIGHT_BYTES       = []byte("right")
	LEFT_BYTES        = []byte("left")
	MOD_BYTES         = []byte("mod")
	YEAR_BYTES        = []byte("year")
	QUARTER_BYTES     = []byte("quarter")
	MONTH_BYTES       = []byte("month")
	WEEK_BYTES        = []byte("week")
	DAY_BYTES         = []byte("day")
	HOUR_BYTES        = []byte("hour")
	MINUTE_BYTES      = []byte("minute")
	SECOND_BYTES      = []byte("second")
	MICROSECOND_BYTES = []byte("microsecond")
	CHAR_BYTES        = []byte("char")
	DATE_BYTES        = []byte("date")
	DATETIME_BYTES    = []byte("datetime")
	FLOAT_BYTES       = []byte("float")
	INTEGER_BYTES     = []byte("integer")
)

//line sql.y:48
type yySymType struct {
	yys         int
	empty       struct{}
	statement   Statement
	selStmt     SelectStatement
	byt         byte
	bytes       []byte
	bytes2      [][]byte
	str         string
	selectExprs SelectExprs
	selectExpr  SelectExpr
	columns     Columns
	colName     *ColName
	tableExprs  TableExprs
	tableExpr   TableExpr
	smTableExpr SimpleTableExpr
	tableName   *TableName
	indexHints  *IndexHints
	expr        Expr
	boolExpr    BoolExpr
	valExpr     ValExpr
	tuple       Tuple
	valExprs    ValExprs
	values      Values
	subquery    *Subquery
	caseExpr    *CaseExpr
	whens       []*When
	when        *When
	orderBy     OrderBy
	order       *Order
	limit       *Limit
	insRows     InsertRows
	updateExprs UpdateExprs
	updateExpr  *UpdateExpr
}

const LEX_ERROR = 57346
const SELECT = 57347
const INSERT = 57348
const UPDATE = 57349
const DELETE = 57350
const FROM = 57351
const WHERE = 57352
const GROUP = 57353
const HAVING = 57354
const ORDER = 57355
const BY = 57356
const LIMIT = 57357
const OFFSET = 57358
const FOR = 57359
const SOME = 57360
const ANY = 57361
const TRUE = 57362
const FALSE = 57363
const UNKNOWN = 57364
const ALL = 57365
const DISTINCT = 57366
const PRECISION = 57367
const AS = 57368
const EXISTS = 57369
const IN = 57370
const IS = 57371
const LIKE = 57372
const BETWEEN = 57373
const NULL = 57374
const ASC = 57375
const DESC = 57376
const VALUES = 57377
const INTO = 57378
const DUPLICATE = 57379
const KEY = 57380
const DEFAULT = 57381
const SET = 57382
const LOCK = 57383
const ID = 57384
const STRING = 57385
const NUMBER = 57386
const VALUE_ARG = 57387
const COMMENT = 57388
const LE = 57389
const GE = 57390
const NE = 57391
const NULL_SAFE_EQUAL = 57392
const DATE = 57393
const DATETIME = 57394
const TIME = 57395
const TIMESTAMP = 57396
const TIMESTAMPADD = 57397
const TIMESTAMPDIFF = 57398
const YEAR = 57399
const QUARTER = 57400
const MONTH = 57401
const WEEK = 57402
const DAY = 57403
const HOUR = 57404
const MINUTE = 57405
const SECOND = 57406
const MICROSECOND = 57407
const SQL_TSI_YEAR = 57408
const SQL_TSI_QUARTER = 57409
const SQL_TSI_MONTH = 57410
const SQL_TSI_WEEK = 57411
const SQL_TSI_DAY = 57412
const SQL_TSI_HOUR = 57413
const SQL_TSI_MINUTE = 57414
const SQL_TSI_SECOND = 57415
const CONVERT = 57416
const CHAR = 57417
const SIGNED = 57418
const UNSIGNED = 57419
const SQL_BIGINT = 57420
const SQL_VARCHAR = 57421
const SQL_DATE = 57422
const SQL_TIMESTAMP = 57423
const SQL_DOUBLE = 57424
const INTEGER = 57425
const UNION = 57426
const MINUS = 57427
const EXCEPT = 57428
const INTERSECT = 57429
const JOIN = 57430
const STRAIGHT_JOIN = 57431
const LEFT = 57432
const RIGHT = 57433
const INNER = 57434
const OUTER = 57435
const CROSS = 57436
const NATURAL = 57437
const USE = 57438
const FORCE = 57439
const ON = 57440
const AND = 57441
const OR = 57442
const XOR = 57443
const NOT = 57444
const MOD = 57445
const DIV = 57446
const UNARY = 57447
const CASE = 57448
const WHEN = 57449
const THEN = 57450
const ELSE = 57451
const END = 57452
const BEGIN = 57453
const COMMIT = 57454
const ROLLBACK = 57455
const NAMES = 57456
const REPLACE = 57457
const ADMIN = 57458
const SHOW = 57459
const DATABASES = 57460
const TABLES = 57461
const PROXY = 57462
const VARIABLES = 57463
const FULL = 57464
const SESSION = 57465
const GLOBAL = 57466
const COLUMNS = 57467
const CREATE = 57468
const ALTER = 57469
const DROP = 57470
const RENAME = 57471
const TABLE = 57472
const INDEX = 57473
const VIEW = 57474
const TO = 57475
const IGNORE = 57476
const IF = 57477
const UNIQUE = 57478
const USING = 57479

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"LEX_ERROR",
	"SELECT",
	"INSERT",
	"UPDATE",
	"DELETE",
	"FROM",
	"WHERE",
	"GROUP",
	"HAVING",
	"ORDER",
	"BY",
	"LIMIT",
	"OFFSET",
	"FOR",
	"SOME",
	"ANY",
	"TRUE",
	"FALSE",
	"UNKNOWN",
	"ALL",
	"DISTINCT",
	"PRECISION",
	"AS",
	"EXISTS",
	"IN",
	"IS",
	"LIKE",
	"BETWEEN",
	"NULL",
	"ASC",
	"DESC",
	"VALUES",
	"INTO",
	"DUPLICATE",
	"KEY",
	"DEFAULT",
	"SET",
	"LOCK",
	"ID",
	"STRING",
	"NUMBER",
	"VALUE_ARG",
	"COMMENT",
	"LE",
	"GE",
	"NE",
	"NULL_SAFE_EQUAL",
	"'('",
	"'='",
	"'<'",
	"'>'",
	"'~'",
	"DATE",
	"DATETIME",
	"TIME",
	"TIMESTAMP",
	"TIMESTAMPADD",
	"TIMESTAMPDIFF",
	"YEAR",
	"QUARTER",
	"MONTH",
	"WEEK",
	"DAY",
	"HOUR",
	"MINUTE",
	"SECOND",
	"MICROSECOND",
	"SQL_TSI_YEAR",
	"SQL_TSI_QUARTER",
	"SQL_TSI_MONTH",
	"SQL_TSI_WEEK",
	"SQL_TSI_DAY",
	"SQL_TSI_HOUR",
	"SQL_TSI_MINUTE",
	"SQL_TSI_SECOND",
	"CONVERT",
	"CHAR",
	"SIGNED",
	"UNSIGNED",
	"SQL_BIGINT",
	"SQL_VARCHAR",
	"SQL_DATE",
	"SQL_TIMESTAMP",
	"SQL_DOUBLE",
	"INTEGER",
	"UNION",
	"MINUS",
	"EXCEPT",
	"INTERSECT",
	"','",
	"JOIN",
	"STRAIGHT_JOIN",
	"LEFT",
	"RIGHT",
	"INNER",
	"OUTER",
	"CROSS",
	"NATURAL",
	"USE",
	"FORCE",
	"ON",
	"AND",
	"OR",
	"XOR",
	"NOT",
	"'&'",
	"'|'",
	"'^'",
	"'+'",
	"'-'",
	"'*'",
	"'/'",
	"'%'",
	"MOD",
	"DIV",
	"'.'",
	"UNARY",
	"CASE",
	"WHEN",
	"THEN",
	"ELSE",
	"END",
	"BEGIN",
	"COMMIT",
	"ROLLBACK",
	"NAMES",
	"REPLACE",
	"ADMIN",
	"SHOW",
	"DATABASES",
	"TABLES",
	"PROXY",
	"VARIABLES",
	"FULL",
	"SESSION",
	"GLOBAL",
	"COLUMNS",
	"CREATE",
	"ALTER",
	"DROP",
	"RENAME",
	"TABLE",
	"INDEX",
	"VIEW",
	"TO",
	"IGNORE",
	"IF",
	"UNIQUE",
	"USING",
	"')'",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 31,
	140, 37,
	-2, 39,
	-1, 225,
	93, 154,
	153, 154,
	-2, 75,
	-1, 486,
	105, 74,
	106, 74,
	107, 74,
	-2, 84,
}

const yyNprod = 288
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1273

var yyAct = [...]int{

	116, 400, 108, 495, 75, 107, 342, 453, 113, 222,
	167, 102, 132, 392, 275, 335, 333, 114, 131, 245,
	504, 223, 3, 240, 470, 141, 105, 103, 77, 320,
	504, 504, 312, 64, 358, 359, 360, 361, 362, 189,
	363, 364, 93, 189, 79, 189, 253, 84, 189, 189,
	86, 89, 78, 398, 90, 66, 34, 35, 36, 37,
	99, 82, 189, 189, 273, 273, 44, 49, 46, 50,
	349, 170, 47, 52, 53, 54, 464, 463, 462, 83,
	506, 417, 419, 100, 166, 96, 85, 51, 422, 80,
	505, 503, 174, 421, 369, 196, 445, 169, 177, 469,
	259, 181, 443, 468, 187, 467, 194, 334, 466, 427,
	334, 186, 389, 397, 225, 163, 158, 221, 226, 65,
	314, 176, 381, 379, 313, 272, 257, 165, 418, 260,
	191, 192, 193, 160, 278, 232, 459, 220, 224, 179,
	180, 393, 345, 277, 393, 239, 18, 56, 58, 59,
	246, 63, 61, 62, 173, 411, 461, 97, 79, 460,
	412, 79, 243, 415, 249, 248, 78, 414, 195, 78,
	208, 209, 211, 212, 210, 250, 76, 409, 413, 160,
	424, 187, 410, 278, 65, 263, 423, 270, 271, 247,
	273, 288, 277, 479, 246, 188, 287, 249, 448, 264,
	289, 279, 386, 297, 298, 385, 302, 303, 304, 305,
	306, 307, 308, 309, 310, 311, 293, 282, 284, 285,
	286, 301, 143, 144, 145, 266, 281, 384, 383, 256,
	258, 255, 280, 356, 299, 34, 35, 36, 37, 316,
	318, 382, 79, 79, 472, 471, 338, 191, 192, 193,
	78, 340, 88, 265, 346, 319, 329, 182, 162, 331,
	330, 489, 337, 341, 242, 347, 79, 241, 488, 72,
	351, 352, 353, 344, 78, 281, 354, 160, 242, 189,
	350, 280, 487, 481, 482, 178, 337, 478, 233, 155,
	279, 231, 368, 230, 355, 229, 228, 227, 374, 375,
	101, 237, 236, 370, 371, 372, 235, 91, 234, 65,
	300, 367, 373, 157, 80, 18, 19, 20, 21, 378,
	206, 207, 208, 209, 211, 212, 210, 366, 420, 380,
	404, 403, 358, 359, 360, 361, 362, 391, 363, 364,
	390, 156, 501, 262, 159, 261, 244, 73, 399, 388,
	22, 171, 396, 168, 395, 191, 192, 193, 164, 161,
	87, 224, 175, 447, 477, 475, 502, 92, 18, 279,
	279, 407, 408, 71, 294, 508, 295, 296, 251, 426,
	203, 204, 205, 206, 207, 208, 209, 211, 212, 210,
	94, 172, 444, 268, 428, 429, 430, 431, 336, 79,
	283, 450, 69, 401, 451, 67, 184, 449, 458, 402,
	95, 425, 269, 343, 455, 203, 204, 205, 206, 207,
	208, 209, 211, 212, 210, 185, 457, 406, 465, 454,
	246, 143, 144, 145, 98, 74, 27, 28, 29, 507,
	30, 32, 31, 377, 490, 18, 39, 57, 473, 474,
	60, 23, 24, 26, 25, 267, 183, 17, 16, 15,
	14, 187, 13, 483, 12, 486, 476, 252, 485, 203,
	204, 205, 206, 207, 208, 209, 211, 212, 210, 45,
	491, 492, 348, 254, 484, 494, 224, 493, 496, 496,
	496, 79, 497, 498, 48, 499, 38, 81, 339, 78,
	143, 144, 145, 500, 317, 509, 454, 122, 480, 510,
	452, 511, 130, 456, 405, 137, 40, 41, 42, 43,
	387, 238, 106, 123, 124, 125, 332, 55, 121, 115,
	432, 111, 117, 394, 112, 135, 126, 129, 127, 128,
	118, 119, 146, 147, 148, 149, 150, 151, 152, 153,
	154, 197, 109, 416, 434, 435, 376, 276, 357, 120,
	203, 204, 205, 206, 207, 208, 209, 211, 212, 210,
	274, 365, 190, 68, 33, 70, 139, 138, 433, 436,
	437, 438, 439, 440, 441, 442, 11, 10, 110, 9,
	8, 7, 133, 134, 104, 6, 5, 140, 4, 2,
	1, 142, 0, 143, 144, 145, 0, 0, 0, 0,
	122, 0, 0, 0, 0, 130, 0, 0, 137, 0,
	0, 0, 0, 0, 0, 106, 123, 124, 125, 0,
	136, 0, 0, 315, 111, 0, 0, 0, 135, 126,
	129, 127, 128, 118, 119, 146, 147, 148, 149, 150,
	151, 152, 153, 154, 0, 0, 0, 0, 0, 0,
	0, 0, 120, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 139,
	138, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 110, 0, 0, 0, 133, 134, 104, 0, 0,
	140, 0, 0, 0, 142, 292, 291, 143, 144, 145,
	290, 0, 0, 0, 0, 0, 0, 0, 0, 130,
	0, 0, 137, 0, 0, 0, 0, 0, 0, 80,
	123, 124, 125, 136, 0, 0, 0, 0, 178, 0,
	0, 0, 135, 126, 129, 127, 128, 118, 119, 146,
	147, 148, 149, 150, 151, 152, 153, 154, 0, 0,
	0, 0, 0, 0, 0, 0, 120, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 139, 138, 0, 0, 0, 0, 0,
	18, 0, 0, 0, 0, 0, 0, 0, 0, 133,
	134, 0, 0, 0, 140, 143, 144, 145, 142, 0,
	0, 0, 122, 0, 0, 0, 0, 130, 0, 0,
	137, 0, 0, 0, 0, 0, 0, 80, 123, 124,
	125, 0, 0, 0, 0, 0, 111, 136, 0, 0,
	135, 126, 129, 127, 128, 118, 119, 146, 147, 148,
	149, 150, 151, 152, 153, 154, 0, 0, 0, 0,
	0, 0, 0, 0, 120, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 139, 138, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 110, 0, 0, 0, 133, 134, 0,
	0, 0, 140, 143, 144, 145, 142, 0, 0, 0,
	122, 0, 0, 0, 0, 130, 0, 0, 137, 0,
	0, 0, 0, 0, 0, 80, 123, 124, 125, 0,
	0, 0, 0, 0, 111, 136, 0, 0, 135, 126,
	129, 127, 128, 118, 119, 146, 147, 148, 149, 150,
	151, 152, 153, 154, 0, 0, 0, 0, 0, 0,
	0, 0, 120, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 139,
	138, 0, 0, 0, 0, 0, 18, 0, 0, 0,
	0, 110, 0, 0, 0, 133, 134, 0, 0, 0,
	140, 143, 144, 145, 142, 0, 0, 0, 0, 0,
	0, 0, 0, 130, 0, 0, 137, 0, 0, 0,
	0, 0, 0, 80, 123, 124, 125, 0, 0, 0,
	0, 0, 178, 136, 0, 0, 135, 126, 129, 127,
	128, 118, 119, 146, 147, 148, 149, 150, 151, 152,
	153, 154, 0, 0, 0, 0, 0, 0, 0, 0,
	120, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 139, 138, 0,
	0, 0, 0, 0, 198, 202, 200, 201, 0, 0,
	0, 0, 0, 133, 134, 0, 0, 0, 140, 143,
	144, 145, 142, 216, 217, 218, 219, 0, 213, 214,
	215, 130, 0, 0, 137, 0, 0, 0, 0, 0,
	0, 80, 123, 124, 125, 0, 0, 0, 0, 0,
	178, 136, 0, 0, 135, 126, 129, 127, 128, 118,
	119, 146, 147, 148, 149, 150, 151, 152, 153, 154,
	0, 0, 0, 0, 0, 0, 0, 0, 120, 0,
	0, 0, 0, 0, 199, 203, 204, 205, 206, 207,
	208, 209, 211, 212, 210, 139, 138, 0, 0, 446,
	0, 0, 198, 202, 200, 201, 0, 0, 0, 0,
	0, 133, 134, 0, 0, 0, 140, 0, 0, 0,
	142, 216, 217, 218, 219, 0, 213, 214, 215, 146,
	147, 148, 149, 150, 151, 152, 153, 154, 321, 322,
	323, 324, 325, 326, 327, 328, 0, 0, 0, 136,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 199, 203, 204, 205, 206, 207, 208, 209,
	211, 212, 210,
}
var yyPact = [...]int{

	310, -1000, -1000, 146, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -79, -80, -58, -72, -1000, -1000, -1000,
	-1000, 14, 267, 440, 382, -1000, -1000, -1000, 378, -1000,
	337, 305, 426, 47, -89, -67, 267, -1000, -59, 267,
	-1000, 318, -99, 267, -99, 331, 380, -51, 425, 267,
	-57, -1000, -1000, -1000, 249, -1000, -1000, -1000, 583, -1000,
	243, 305, 273, -3, 305, 86, 317, -1000, 206, -1000,
	-4, 316, 19, 267, -1000, 311, -1000, -77, 309, 364,
	50, 267, 305, -1000, 883, 1079, 380, 380, 1079, 425,
	397, 1079, 186, -1000, -1000, 142, -24, -1000, 1154, -1000,
	883, 785, -1000, -1000, -1000, 1079, 246, 245, 244, 242,
	240, -1000, 237, -1000, -1000, -1000, 265, 263, 259, 258,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 1079, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 227, 272, 304, 420,
	272, -1000, 1079, 267, -1000, 351, -106, -1000, 87, -1000,
	303, -1000, -1000, 301, -1000, 213, 25, 360, 981, -1000,
	-1000, 360, 380, 384, 1079, 1079, -28, 360, 92, 583,
	375, 883, 883, 883, -1000, 267, 77, 687, 234, 346,
	1079, 1079, 202, 1079, 1079, 1079, 1079, 1079, 1079, 1079,
	1079, 1079, 1079, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -121, -29, -33, 25, 1154, -1000, 480, 583, 1147,
	1147, 583, -1000, 440, -1000, -1000, -1000, -1000, -15, 360,
	363, 272, 272, 184, -1000, 400, 883, -1000, 360, -1000,
	-1000, -1000, 38, 267, -1000, -78, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 363, 272, -1000, -1000, 1079, 1079,
	360, 360, -1000, 1079, 140, 238, 285, 141, -25, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 360,
	237, 237, 237, -1000, 234, 1079, 1079, 360, 451, -1000,
	411, -1000, 208, 208, 208, 56, 56, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -30, 583, -31, 148,
	135, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 134,
	112, 109, -12, -1000, 883, 37, 234, 146, 40, -40,
	-1000, 400, 388, 395, 25, 289, -1000, -1000, 288, -1000,
	-1000, 86, 360, 360, 360, 416, 92, 92, -1000, -1000,
	83, 61, 84, 73, 69, -21, -1000, 286, -60, 46,
	-1000, -1000, -1000, -1000, 360, 306, 1079, -1000, -1000, -1000,
	-44, -1000, 583, 583, 583, 583, 498, -23, -1000, 1079,
	-27, 1056, -1000, 326, 105, -1000, -1000, -1000, 272, 388,
	-1000, 1079, 883, -1000, -1000, 414, 394, 238, 32, -1000,
	65, -1000, 62, -1000, -1000, -1000, -1000, -68, -69, -70,
	-1000, -1000, -1000, -1000, -1000, 1079, 360, -1000, -45, -48,
	-50, -54, -129, -1000, -1000, -1000, 157, 156, -1000, -1000,
	-1000, -1000, -1000, -1000, 360, 1079, 1079, 327, 234, -1000,
	-1000, 271, 100, -1000, 250, -1000, 400, 883, 1079, 883,
	-1000, -1000, 231, 217, 210, 360, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, 360, 360, 437, -1000, 1079, 1079, 883,
	-1000, -1000, -1000, 388, 25, 97, -1000, 267, 267, 267,
	272, 360, 360, -1000, 325, -62, -1000, -63, -73, 86,
	-1000, 432, 347, -1000, 267, -1000, -1000, -1000, 267, -1000,
	267, -1000,
}
var yyPgo = [...]int{

	0, 600, 599, 21, 598, 596, 595, 591, 590, 589,
	587, 586, 496, 575, 574, 573, 11, 27, 572, 571,
	26, 570, 14, 558, 557, 269, 553, 3, 19, 5,
	552, 551, 15, 534, 2, 17, 9, 18, 533, 532,
	25, 29, 530, 12, 529, 8, 528, 526, 16, 521,
	520, 514, 513, 6, 510, 7, 508, 1, 503, 23,
	498, 13, 4, 28, 252, 497, 494, 483, 482, 479,
	467, 0, 10, 464, 462, 460, 459, 458, 457, 157,
	42, 456, 455, 450, 447, 446,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 76, 76, 5, 6, 7, 7, 73, 74,
	75, 78, 81, 81, 82, 82, 82, 83, 83, 84,
	84, 84, 77, 77, 77, 77, 77, 8, 8, 8,
	9, 9, 9, 10, 11, 11, 11, 85, 12, 13,
	13, 14, 14, 14, 14, 14, 15, 15, 16, 16,
	17, 17, 17, 17, 20, 20, 18, 18, 18, 21,
	21, 22, 22, 22, 22, 19, 19, 19, 23, 23,
	23, 23, 23, 23, 23, 23, 23, 24, 24, 24,
	24, 24, 24, 24, 25, 25, 26, 26, 26, 26,
	27, 27, 28, 28, 80, 80, 80, 79, 79, 29,
	29, 29, 29, 29, 29, 30, 30, 30, 30, 30,
	30, 30, 30, 30, 30, 30, 30, 30, 30, 30,
	31, 31, 31, 31, 31, 31, 31, 32, 32, 38,
	38, 35, 35, 43, 36, 36, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	41, 41, 41, 41, 41, 41, 41, 41, 40, 40,
	40, 40, 40, 40, 40, 40, 40, 42, 42, 42,
	42, 42, 42, 42, 42, 42, 42, 42, 42, 39,
	39, 39, 39, 39, 39, 44, 44, 44, 46, 49,
	49, 47, 47, 48, 48, 50, 50, 45, 45, 33,
	33, 33, 33, 33, 33, 33, 33, 33, 37, 37,
	37, 51, 51, 52, 52, 53, 53, 54, 54, 55,
	56, 56, 56, 57, 57, 57, 57, 58, 58, 58,
	59, 59, 60, 60, 61, 61, 62, 62, 63, 64,
	64, 65, 65, 66, 66, 67, 67, 67, 67, 67,
	68, 68, 69, 69, 70, 70, 71, 72,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 4, 12, 3,
	7, 7, 6, 6, 8, 7, 3, 4, 1, 1,
	1, 5, 2, 2, 0, 2, 2, 0, 1, 0,
	1, 1, 3, 4, 4, 5, 5, 5, 8, 4,
	6, 7, 4, 5, 4, 5, 5, 0, 2, 0,
	2, 1, 2, 1, 1, 1, 0, 1, 1, 3,
	1, 2, 3, 3, 1, 1, 0, 1, 2, 1,
	3, 3, 3, 3, 5, 0, 1, 2, 1, 1,
	2, 3, 2, 3, 2, 2, 2, 1, 3, 1,
	1, 1, 3, 3, 1, 3, 0, 5, 5, 5,
	1, 3, 0, 2, 0, 2, 2, 0, 2, 1,
	3, 3, 3, 2, 3, 3, 3, 4, 4, 4,
	4, 3, 4, 5, 6, 3, 4, 2, 3, 4,
	1, 1, 1, 1, 1, 1, 1, 2, 1, 1,
	3, 3, 1, 3, 1, 3, 1, 1, 1, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 2,
	3, 4, 5, 4, 6, 6, 6, 6, 6, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 2, 1, 2, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 5, 0,
	1, 1, 2, 4, 4, 0, 2, 1, 3, 1,
	1, 1, 2, 2, 2, 2, 1, 1, 1, 1,
	1, 0, 3, 0, 2, 0, 3, 1, 3, 2,
	0, 1, 1, 0, 2, 4, 4, 0, 2, 4,
	0, 3, 1, 3, 0, 5, 1, 3, 3, 0,
	2, 0, 3, 0, 1, 1, 1, 1, 1, 1,
	0, 1, 0, 1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -73, -74, -75, -76, -77, -78, 5, 6,
	7, 8, 40, 141, 142, 144, 143, 126, 127, 128,
	130, 132, 131, -14, 89, 90, 91, 92, -12, -85,
	-12, -12, -12, -12, 145, -69, 147, 151, -66, 147,
	149, 145, 145, 146, 147, -12, 133, -84, 134, 135,
	-83, 138, 139, 137, -71, 42, -3, 23, -15, 24,
	-13, 36, -25, 42, 9, -62, 129, -63, -45, -71,
	42, -65, 150, 146, -71, 145, -71, 42, -64, 150,
	-71, -64, 36, -80, 10, 30, 136, -79, 9, -71,
	140, 51, -16, -17, 114, -20, 42, -29, -34, -30,
	108, 51, -33, -45, -35, -44, -71, -39, 60, 61,
	79, -46, 27, 43, 44, 45, 56, 58, 59, 57,
	32, -37, -43, 112, 113, 55, 150, 35, 97, 96,
	117, -40, 121, 20, 21, 22, 62, 63, 64, 65,
	66, 67, 68, 69, 70, 46, -25, 40, 119, -25,
	93, 42, 52, 119, 42, 108, -71, -72, 42, -72,
	148, 42, 27, 104, -71, -25, -20, -34, 51, -80,
	-80, -34, -79, -81, 9, 28, -36, -34, 9, 93,
	-18, 105, 106, 107, -71, 26, 119, -31, 28, 108,
	30, 31, 29, 109, 110, 111, 112, 113, 114, 115,
	118, 116, 117, 52, 53, 54, 47, 48, 49, 50,
	-20, -29, -36, -3, -20, -34, -34, 51, 51, 51,
	51, 51, -43, 51, 43, 43, 43, 43, -49, -34,
	-59, 40, 51, -62, 42, -28, 10, -63, -34, -71,
	-72, 27, -70, 152, -67, 144, 142, 39, 143, 13,
	42, 42, 42, -72, -59, 40, -80, -82, 9, 28,
	-34, -34, 153, 93, -21, -22, -24, 51, 42, -43,
	140, 134, -17, 25, -20, -20, -20, -71, 114, -34,
	23, 19, 18, -35, 28, 30, 31, -34, -34, 32,
	108, -37, -34, -34, -34, -34, -34, -34, -34, -34,
	-34, -34, 153, 153, 153, 153, -16, 24, -16, -40,
	-41, 71, 72, 73, 74, 75, 76, 77, 78, -40,
	-41, -17, -47, -48, 122, -32, 35, -3, -62, -60,
	-45, -28, -53, 13, -20, 104, -71, -72, -68, 148,
	-32, -62, -34, -34, -34, -28, 93, -23, 94, 95,
	96, 97, 98, 100, 101, -19, 42, 26, -22, 119,
	-43, -43, -43, -35, -34, -34, 105, 32, -37, 153,
	-16, 153, 93, 93, 93, 93, 93, -50, -48, 124,
	-29, -34, -61, 104, -38, -35, -61, 153, 93, -53,
	-57, 15, 14, 42, 42, -51, 11, -22, -22, 94,
	99, 94, 99, 94, 94, 94, -26, 102, 149, 103,
	42, 153, 42, 140, 134, 105, -34, 153, -16, -16,
	-16, -16, -42, 80, 56, 57, 81, 82, 83, 84,
	85, 86, 87, 125, -34, 123, 123, 37, 93, -45,
	-57, -34, -54, -55, -20, -72, -52, 12, 14, 104,
	94, 94, 146, 146, 146, -34, 153, 153, 153, 153,
	153, 88, 88, -34, -34, 38, -35, 93, 16, 93,
	-56, 33, 34, -53, -20, -36, -29, 51, 51, 51,
	7, -34, -34, -55, -57, -27, -71, -27, -27, -62,
	-58, 17, 41, 153, 93, 153, 153, 7, 28, -71,
	-71, -71,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 57, 57,
	57, 57, 57, 282, 273, 0, 0, 28, 29, 30,
	57, -2, 0, 0, 61, 63, 64, 65, 66, 59,
	0, 0, 0, 0, 271, 0, 0, 283, 0, 0,
	274, 0, 269, 0, 269, 0, 114, 0, 117, 0,
	0, 40, 41, 38, 0, 286, 19, 62, 0, 67,
	58, 0, 0, 104, 0, 26, 0, 266, 0, 227,
	286, 0, 0, 0, 287, 0, 287, 0, 0, 0,
	0, 0, 0, 42, 0, 0, 114, 114, 0, 117,
	0, 0, 17, 68, 70, 76, 286, 74, 75, 119,
	0, 0, 156, 157, 158, 0, 227, 0, 0, 0,
	0, 179, 0, 229, 230, 231, 0, 0, 0, 0,
	236, 237, 152, 215, 216, 217, 209, 210, 211, 212,
	213, 214, 219, 238, 239, 240, 188, 189, 190, 191,
	192, 193, 194, 195, 196, 60, 260, 0, 0, 112,
	0, 27, 0, 0, 287, 0, 284, 49, 0, 52,
	0, 54, 270, 0, 287, 260, 115, 116, 0, 43,
	44, 118, 114, 34, 0, 0, 0, 154, 0, 0,
	71, 0, 0, 0, 77, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 140, 141, 142, 143, 144, 145, 146,
	123, 74, 0, 0, 0, -2, 169, 0, 0, 0,
	0, 0, 137, 0, 232, 233, 234, 235, 0, 220,
	0, 0, 0, 112, 105, 245, 0, 267, 268, 228,
	47, 272, 0, 0, 287, 280, 275, 276, 277, 278,
	279, 53, 55, 56, 0, 0, 45, 46, 0, 0,
	32, 33, 31, 0, 112, 79, 85, 0, 97, 99,
	100, 101, 69, 72, 120, 121, 122, 78, 73, 125,
	0, 0, 0, 126, 0, 0, 0, 131, 0, 135,
	0, 138, 159, 160, 161, 162, 163, 164, 165, 166,
	167, 168, 124, 151, 153, 170, 0, 0, 0, 0,
	0, 180, 181, 182, 183, 184, 185, 186, 187, 0,
	0, 0, 225, 221, 0, 264, 0, 148, 264, 0,
	262, 245, 253, 0, 113, 0, 285, 50, 0, 281,
	22, 23, 35, 36, 155, 241, 0, 0, 88, 89,
	0, 0, 0, 0, 0, 106, 86, 0, 0, 0,
	127, 128, 129, 130, 132, 0, 0, 136, 139, 171,
	0, 173, 0, 0, 0, 0, 0, 0, 222, 0,
	74, 75, 20, 0, 147, 149, 21, 261, 0, 253,
	25, 0, 0, 287, 51, 243, 0, 80, 83, 90,
	0, 92, 0, 94, 95, 96, 81, 0, 0, 0,
	87, 82, 98, 102, 103, 0, 133, 172, 0, 0,
	0, 0, 0, 197, 198, 199, 200, 202, 204, 205,
	206, 207, 208, 218, 226, 0, 0, 0, 0, 263,
	24, 254, 246, 247, 250, 48, 245, 0, 0, 0,
	91, 93, 0, 0, 0, 134, 174, 175, 176, 177,
	178, 201, 203, 223, 224, 0, 150, 0, 0, 0,
	249, 251, 252, 253, 244, 242, -2, 0, 0, 0,
	0, 255, 256, 248, 257, 0, 110, 0, 0, 265,
	18, 0, 0, 107, 0, 108, 109, 258, 0, 111,
	0, 259,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 116, 109, 3,
	51, 153, 114, 112, 93, 113, 119, 115, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	53, 52, 54, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 111, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 110, 3, 55,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 47, 48, 49, 50, 56,
	57, 58, 59, 60, 61, 62, 63, 64, 65, 66,
	67, 68, 69, 70, 71, 72, 73, 74, 75, 76,
	77, 78, 79, 80, 81, 82, 83, 84, 85, 86,
	87, 88, 89, 90, 91, 92, 94, 95, 96, 97,
	98, 99, 100, 101, 102, 103, 104, 105, 106, 107,
	108, 117, 118, 120, 121, 122, 123, 124, 125, 126,
	127, 128, 129, 130, 131, 132, 133, 134, 135, 136,
	137, 138, 139, 140, 141, 142, 143, 144, 145, 146,
	147, 148, 149, 150, 151, 152,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lookahead func() int
}

func (p *yyParserImpl) Lookahead() int {
	return p.lookahead()
}

func yyNewParser() yyParser {
	p := &yyParserImpl{
		lookahead: func() int { return -1 },
	}
	return p
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yytoken := -1 // yychar translated into internal numbering
	yyrcvr.lookahead = func() int { return yychar }
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yychar = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar, yytoken = yylex1(yylex, &yylval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yychar = -1
		yytoken = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar, yytoken = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yychar = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:197
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:203
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:223
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 18:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:227
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].valExprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:231
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 20:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:238
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:242
		{
			cols := make(Columns, 0, len(yyDollar[6].updateExprs))
			vals := make(ValTuple, 0, len(yyDollar[6].updateExprs))
			for _, col := range yyDollar[6].updateExprs {
				cols = append(cols, &NonStarExpr{Expr: col.Name})
				vals = append(vals, col.Expr)
			}
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: cols, Rows: Values{vals}, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 22:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:254
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:258
		{
			cols := make(Columns, 0, len(yyDollar[6].updateExprs))
			vals := make(ValTuple, 0, len(yyDollar[6].updateExprs))
			for _, col := range yyDollar[6].updateExprs {
				cols = append(cols, &NonStarExpr{Expr: col.Name})
				vals = append(vals, col.Expr)
			}
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: cols, Rows: Values{vals}}
		}
	case 24:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:271
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 25:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:277
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:283
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:287
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:293
		{
			yyVAL.statement = &Begin{}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:299
		{
			yyVAL.statement = &Commit{}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:305
		{
			yyVAL.statement = &Rollback{}
		}
	case 31:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:311
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].valExprs}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:317
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:321
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 34:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:326
		{
			yyVAL.valExpr = nil
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:330
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:334
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 37:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:339
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:343
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 39:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:348
		{
			yyVAL.str = AST_SHOW_SESSION_VARIABLE
		}
	case 40:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:352
		{
			yyVAL.str = AST_SHOW_SESSION_VARIABLE
		}
	case 41:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:356
		{
			yyVAL.str = AST_SHOW_GLOBAL_VARIABLE
		}
	case 42:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:363
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 43:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:367
		{
			yyVAL.statement = &Show{Section: "variables", Modifier: yyDollar[2].str, LikeOrWhere: yyDollar[4].expr}
		}
	case 44:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:371
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].valExpr, LikeOrWhere: yyDollar[4].expr}
		}
	case 45:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:375
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].valExpr, LikeOrWhere: yyDollar[5].expr}
		}
	case 46:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:379
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].valExpr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].valExpr}
		}
	case 47:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:385
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 48:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:389
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 49:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:394
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 50:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:400
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 51:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:404
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 52:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:409
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 53:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:415
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 54:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:421
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 55:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:425
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 56:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:430
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 57:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:435
		{
			SetAllowComments(yylex, true)
		}
	case 58:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:439
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 59:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:445
		{
			yyVAL.bytes2 = nil
		}
	case 60:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:449
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:455
		{
			yyVAL.str = AST_UNION
		}
	case 62:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:459
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 63:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:463
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:467
		{
			yyVAL.str = AST_EXCEPT
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:471
		{
			yyVAL.str = AST_INTERSECT
		}
	case 66:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:476
		{
			yyVAL.str = ""
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:480
		{
			yyVAL.str = AST_DISTINCT
		}
	case 68:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:486
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:490
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 70:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:496
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 71:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:500
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 72:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:504
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 73:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:508
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:514
		{
			yyVAL.expr = yyDollar[1].boolExpr
		}
	case 75:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:518
		{
			yyVAL.expr = yyDollar[1].valExpr
		}
	case 76:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:523
		{
			yyVAL.bytes = nil
		}
	case 77:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:527
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 78:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:531
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 79:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:537
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:541
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:547
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 82:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:551
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 83:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:555
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 84:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:559
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].boolExpr}
		}
	case 85:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:564
		{
			yyVAL.bytes = nil
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:568
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 87:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:572
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 88:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:578
		{
			yyVAL.str = AST_JOIN
		}
	case 89:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:582
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 90:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:586
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 91:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:590
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:594
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 93:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:598
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 94:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:602
		{
			yyVAL.str = AST_JOIN
		}
	case 95:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:606
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 96:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:610
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:616
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 98:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:620
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 99:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:624
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 100:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:628
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 101:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:632
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:636
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 103:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:640
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 104:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:646
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 105:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:650
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 106:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:655
		{
			yyVAL.indexHints = nil
		}
	case 107:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:659
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 108:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:663
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 109:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:667
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 110:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:673
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 111:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:677
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 112:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:682
		{
			yyVAL.expr = nil
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:686
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 114:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:691
		{
			yyVAL.expr = nil
		}
	case 115:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:695
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 116:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:699
		{
			yyVAL.expr = yyDollar[2].valExpr
		}
	case 117:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:704
		{
			yyVAL.valExpr = nil
		}
	case 118:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:708
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 120:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:715
		{
			yyVAL.boolExpr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 121:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:719
		{
			yyVAL.boolExpr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 122:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:723
		{
			yyVAL.boolExpr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 123:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:727
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 124:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:731
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyDollar[2].boolExpr}
		}
	case 125:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:737
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, Right: yyDollar[3].valExpr}
		}
	case 126:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:741
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 127:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:745
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ALL, Right: yyDollar[4].subquery}
		}
	case 128:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:749
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ANY, Right: yyDollar[4].subquery}
		}
	case 129:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:753
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_SOME, Right: yyDollar[4].subquery}
		}
	case 130:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:757
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 131:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:761
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_LIKE, Right: yyDollar[3].valExpr}
		}
	case 132:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:765
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_LIKE, Right: yyDollar[4].valExpr}
		}
	case 133:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:769
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_BETWEEN, From: yyDollar[3].valExpr, To: yyDollar[5].valExpr}
		}
	case 134:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:773
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].valExpr, To: yyDollar[6].valExpr}
		}
	case 135:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:777
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 136:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:781
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 137:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:785
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 138:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:789
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS, Right: yyDollar[3].valExpr}
		}
	case 139:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:793
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS_NOT, Right: yyDollar[4].valExpr}
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:799
		{
			yyVAL.str = AST_EQ
		}
	case 141:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:803
		{
			yyVAL.str = AST_LT
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:807
		{
			yyVAL.str = AST_GT
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:811
		{
			yyVAL.str = AST_LE
		}
	case 144:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:815
		{
			yyVAL.str = AST_GE
		}
	case 145:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:819
		{
			yyVAL.str = AST_NE
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:823
		{
			yyVAL.str = AST_NSE
		}
	case 147:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:829
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 148:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:833
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 149:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:839
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 150:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:843
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 151:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:849
		{
			yyVAL.tuple = ValTuple(yyDollar[2].valExprs)
		}
	case 152:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:853
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:859
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 154:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:865
		{
			yyVAL.valExprs = ValExprs{yyDollar[1].valExpr}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:869
		{
			yyVAL.valExprs = append(yyDollar[1].valExprs, yyDollar[3].valExpr)
		}
	case 156:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:875
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 157:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:879
		{
			yyVAL.valExpr = yyDollar[1].colName
		}
	case 158:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:883
		{
			yyVAL.valExpr = yyDollar[1].tuple
		}
	case 159:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:887
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITAND, Right: yyDollar[3].valExpr}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:891
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITOR, Right: yyDollar[3].valExpr}
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:895
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITXOR, Right: yyDollar[3].valExpr}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:899
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_PLUS, Right: yyDollar[3].valExpr}
		}
	case 163:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:903
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MINUS, Right: yyDollar[3].valExpr}
		}
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:907
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MULT, Right: yyDollar[3].valExpr}
		}
	case 165:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:911
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_DIV, Right: yyDollar[3].valExpr}
		}
	case 166:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:915
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_IDIV, Right: yyDollar[3].valExpr}
		}
	case 167:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:919
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 168:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:923
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 169:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:927
		{
			if num, ok := yyDollar[2].valExpr.(NumVal); ok {
				switch yyDollar[1].byt {
				case '-':
					yyVAL.valExpr = append(NumVal("-"), num...)
				case '+':
					yyVAL.valExpr = num
				default:
					yyVAL.valExpr = &UnaryExpr{Operator: yyDollar[1].byt, Expr: yyDollar[2].valExpr}
				}
			} else {
				yyVAL.valExpr = &UnaryExpr{Operator: yyDollar[1].byt, Expr: yyDollar[2].valExpr}
			}
		}
	case 170:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:942
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 171:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:946
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 172:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:950
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 173:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:954
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 174:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:958
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 175:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:962
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 176:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:966
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 177:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:970
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 178:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:974
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 179:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:978
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 180:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:984
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 181:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:988
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 182:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:992
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 183:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:996
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 184:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1000
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 185:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1004
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 186:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1008
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 187:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1012
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1018
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1022
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1026
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1030
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1034
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1038
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1042
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 195:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1046
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 196:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1050
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1056
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 198:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1060
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 199:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1064
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1068
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1072
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1076
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 203:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1080
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1084
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 205:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1088
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1092
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1096
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1100
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 209:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1106
		{
			yyVAL.bytes = IF_BYTES
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1110
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1114
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1118
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1122
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1126
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1132
		{
			yyVAL.byt = AST_UPLUS
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1136
		{
			yyVAL.byt = AST_UMINUS
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1140
		{
			yyVAL.byt = AST_TILDA
		}
	case 218:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1146
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 219:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1151
		{
			yyVAL.valExpr = nil
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1155
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1161
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 222:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1165
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 223:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1171
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 224:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1175
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 225:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1180
		{
			yyVAL.valExpr = nil
		}
	case 226:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1184
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1190
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 228:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1194
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1200
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1204
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 231:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1208
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 232:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1212
		{
			yyVAL.valExpr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 233:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1216
		{
			yyVAL.valExpr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 234:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1220
		{
			yyVAL.valExpr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 235:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1224
		{
			yyVAL.valExpr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1228
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1232
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1238
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1242
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1246
		{
			yyVAL.valExpr = &UnknownVal{}
		}
	case 241:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1251
		{
			yyVAL.valExprs = nil
		}
	case 242:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1255
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 243:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1260
		{
			yyVAL.expr = nil
		}
	case 244:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1264
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 245:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1269
		{
			yyVAL.orderBy = nil
		}
	case 246:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1273
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1279
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 248:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1283
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 249:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1289
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 250:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1294
		{
			yyVAL.str = AST_ASC
		}
	case 251:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1298
		{
			yyVAL.str = AST_ASC
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1302
		{
			yyVAL.str = AST_DESC
		}
	case 253:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1307
		{
			yyVAL.limit = nil
		}
	case 254:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1311
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 255:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1315
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 256:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1319
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].valExpr, Rowcount: yyDollar[2].valExpr}
		}
	case 257:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1324
		{
			yyVAL.str = ""
		}
	case 258:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1328
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 259:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1332
		{
			if !bytes.Equal(yyDollar[3].bytes, SHARE) {
				yylex.Error("expecting share")
				return 1
			}
			if !bytes.Equal(yyDollar[4].bytes, MODE) {
				yylex.Error("expecting mode")
				return 1
			}
			yyVAL.str = AST_SHARE_MODE
		}
	case 260:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1345
		{
			yyVAL.columns = nil
		}
	case 261:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1349
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 262:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1355
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 263:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1359
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 264:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1364
		{
			yyVAL.updateExprs = nil
		}
	case 265:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1368
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 266:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1374
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 267:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1378
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 268:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1384
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 269:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1389
		{
			yyVAL.empty = struct{}{}
		}
	case 270:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1391
		{
			yyVAL.empty = struct{}{}
		}
	case 271:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1394
		{
			yyVAL.empty = struct{}{}
		}
	case 272:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1396
		{
			yyVAL.empty = struct{}{}
		}
	case 273:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1399
		{
			yyVAL.empty = struct{}{}
		}
	case 274:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1401
		{
			yyVAL.empty = struct{}{}
		}
	case 275:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1405
		{
			yyVAL.empty = struct{}{}
		}
	case 276:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1407
		{
			yyVAL.empty = struct{}{}
		}
	case 277:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1409
		{
			yyVAL.empty = struct{}{}
		}
	case 278:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1411
		{
			yyVAL.empty = struct{}{}
		}
	case 279:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1413
		{
			yyVAL.empty = struct{}{}
		}
	case 280:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1416
		{
			yyVAL.empty = struct{}{}
		}
	case 281:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1418
		{
			yyVAL.empty = struct{}{}
		}
	case 282:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1421
		{
			yyVAL.empty = struct{}{}
		}
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1423
		{
			yyVAL.empty = struct{}{}
		}
	case 284:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1426
		{
			yyVAL.empty = struct{}{}
		}
	case 285:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1428
		{
			yyVAL.empty = struct{}{}
		}
	case 286:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1432
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 287:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1437
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
