//line sql.y:6
package parser

import __yyfmt__ "fmt"

//line sql.y:6
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
const ALL = 57364
const DISTINCT = 57365
const PRECISION = 57366
const AS = 57367
const EXISTS = 57368
const IN = 57369
const IS = 57370
const LIKE = 57371
const BETWEEN = 57372
const NULL = 57373
const ASC = 57374
const DESC = 57375
const VALUES = 57376
const INTO = 57377
const DUPLICATE = 57378
const KEY = 57379
const DEFAULT = 57380
const SET = 57381
const LOCK = 57382
const ID = 57383
const STRING = 57384
const NUMBER = 57385
const VALUE_ARG = 57386
const COMMENT = 57387
const LE = 57388
const GE = 57389
const NE = 57390
const NULL_SAFE_EQUAL = 57391
const DATE = 57392
const DATETIME = 57393
const TIME = 57394
const TIMESTAMP = 57395
const TIMESTAMPADD = 57396
const TIMESTAMPDIFF = 57397
const YEAR = 57398
const QUARTER = 57399
const MONTH = 57400
const WEEK = 57401
const DAY = 57402
const HOUR = 57403
const MINUTE = 57404
const SECOND = 57405
const MICROSECOND = 57406
const SQL_TSI_YEAR = 57407
const SQL_TSI_QUARTER = 57408
const SQL_TSI_MONTH = 57409
const SQL_TSI_WEEK = 57410
const SQL_TSI_DAY = 57411
const SQL_TSI_HOUR = 57412
const SQL_TSI_MINUTE = 57413
const SQL_TSI_SECOND = 57414
const CONVERT = 57415
const CHAR = 57416
const SIGNED = 57417
const UNSIGNED = 57418
const SQL_BIGINT = 57419
const SQL_VARCHAR = 57420
const SQL_DATE = 57421
const SQL_TIMESTAMP = 57422
const SQL_DOUBLE = 57423
const INTEGER = 57424
const UNION = 57425
const MINUS = 57426
const EXCEPT = 57427
const INTERSECT = 57428
const JOIN = 57429
const STRAIGHT_JOIN = 57430
const LEFT = 57431
const RIGHT = 57432
const INNER = 57433
const OUTER = 57434
const CROSS = 57435
const NATURAL = 57436
const USE = 57437
const FORCE = 57438
const ON = 57439
const AND = 57440
const OR = 57441
const XOR = 57442
const NOT = 57443
const MOD = 57444
const DIV = 57445
const UNARY = 57446
const CASE = 57447
const WHEN = 57448
const THEN = 57449
const ELSE = 57450
const END = 57451
const BEGIN = 57452
const COMMIT = 57453
const ROLLBACK = 57454
const NAMES = 57455
const REPLACE = 57456
const ADMIN = 57457
const SHOW = 57458
const DATABASES = 57459
const TABLES = 57460
const PROXY = 57461
const VARIABLES = 57462
const FULL = 57463
const COLUMNS = 57464
const CREATE = 57465
const ALTER = 57466
const DROP = 57467
const RENAME = 57468
const TABLE = 57469
const INDEX = 57470
const VIEW = 57471
const TO = 57472
const IGNORE = 57473
const IF = 57474
const UNIQUE = 57475
const USING = 57476

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
const yyInitialStackSize = 16

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 220,
	92, 149,
	150, 149,
	-2, 72,
	-1, 479,
	104, 71,
	105, 71,
	106, 71,
	-2, 81,
}

const yyNprod = 281
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1233

var yyAct = [...]int{

	114, 393, 106, 488, 73, 105, 336, 446, 111, 217,
	163, 100, 131, 385, 270, 329, 327, 112, 240, 218,
	3, 235, 140, 75, 314, 497, 103, 101, 352, 353,
	354, 355, 356, 62, 357, 358, 34, 35, 36, 37,
	463, 306, 91, 497, 77, 497, 248, 82, 184, 184,
	84, 87, 76, 64, 88, 184, 184, 80, 343, 184,
	97, 166, 391, 184, 184, 268, 268, 44, 49, 46,
	50, 457, 456, 47, 52, 53, 54, 455, 273, 81,
	410, 412, 162, 499, 83, 414, 51, 272, 438, 98,
	170, 63, 78, 436, 18, 165, 173, 328, 308, 176,
	94, 498, 182, 496, 189, 161, 462, 461, 328, 181,
	382, 363, 220, 460, 459, 216, 221, 420, 191, 172,
	390, 374, 372, 307, 267, 411, 159, 415, 154, 452,
	273, 386, 156, 227, 339, 215, 219, 169, 175, 272,
	186, 187, 188, 386, 234, 56, 58, 59, 57, 61,
	254, 404, 454, 453, 77, 408, 405, 77, 238, 402,
	244, 243, 76, 283, 403, 76, 294, 407, 406, 156,
	276, 245, 190, 183, 275, 252, 95, 182, 255, 74,
	242, 258, 265, 266, 18, 19, 20, 21, 63, 268,
	241, 282, 244, 259, 472, 284, 274, 441, 292, 293,
	379, 296, 297, 298, 299, 300, 301, 302, 303, 304,
	305, 288, 277, 279, 280, 281, 378, 377, 22, 417,
	261, 241, 276, 416, 376, 375, 275, 201, 202, 203,
	204, 206, 207, 205, 310, 312, 465, 77, 77, 474,
	475, 332, 295, 464, 158, 76, 334, 313, 323, 340,
	324, 186, 187, 188, 325, 331, 184, 335, 482, 86,
	341, 77, 260, 481, 480, 345, 346, 347, 338, 76,
	174, 348, 350, 237, 177, 344, 251, 253, 250, 331,
	203, 204, 206, 207, 205, 274, 70, 362, 349, 34,
	35, 36, 37, 368, 369, 228, 226, 225, 364, 365,
	366, 224, 223, 156, 27, 28, 29, 367, 30, 32,
	31, 186, 187, 188, 89, 222, 99, 23, 24, 26,
	25, 151, 361, 373, 236, 232, 352, 353, 354, 355,
	356, 384, 357, 358, 383, 237, 231, 230, 360, 494,
	229, 153, 392, 381, 63, 78, 389, 413, 388, 397,
	193, 197, 195, 196, 396, 219, 152, 257, 256, 155,
	239, 71, 495, 274, 274, 400, 401, 167, 164, 211,
	212, 213, 214, 419, 208, 209, 210, 171, 160, 157,
	85, 468, 440, 90, 69, 437, 371, 421, 422, 423,
	424, 18, 77, 289, 443, 290, 291, 444, 501, 92,
	442, 263, 246, 168, 418, 179, 278, 448, 198, 199,
	200, 201, 202, 203, 204, 206, 207, 205, 93, 264,
	330, 458, 447, 180, 67, 65, 38, 394, 451, 395,
	194, 198, 199, 200, 201, 202, 203, 204, 206, 207,
	205, 466, 467, 337, 450, 439, 40, 41, 42, 43,
	399, 241, 96, 72, 182, 500, 476, 55, 479, 469,
	483, 478, 198, 199, 200, 201, 202, 203, 204, 206,
	207, 205, 18, 484, 485, 39, 60, 477, 487, 219,
	486, 489, 489, 489, 77, 490, 491, 262, 492, 178,
	17, 16, 76, 15, 121, 122, 14, 311, 502, 447,
	120, 13, 503, 12, 504, 130, 247, 45, 136, 342,
	249, 48, 79, 333, 493, 104, 123, 124, 125, 473,
	445, 449, 398, 380, 109, 233, 326, 119, 134, 126,
	129, 127, 128, 116, 117, 142, 143, 144, 145, 146,
	147, 148, 149, 150, 113, 425, 115, 387, 110, 370,
	192, 107, 118, 198, 199, 200, 201, 202, 203, 204,
	206, 207, 205, 409, 271, 351, 269, 359, 185, 138,
	137, 66, 33, 68, 11, 10, 9, 471, 8, 7,
	6, 108, 5, 4, 2, 132, 133, 102, 1, 0,
	139, 0, 121, 122, 141, 0, 0, 0, 120, 0,
	0, 0, 0, 130, 0, 0, 136, 0, 0, 0,
	0, 0, 0, 104, 123, 124, 125, 0, 427, 428,
	0, 135, 109, 0, 309, 0, 134, 126, 129, 127,
	128, 116, 117, 142, 143, 144, 145, 146, 147, 148,
	149, 150, 426, 429, 430, 431, 432, 433, 434, 435,
	118, 0, 0, 470, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 138, 137, 198,
	199, 200, 201, 202, 203, 204, 206, 207, 205, 108,
	0, 0, 0, 132, 133, 102, 0, 0, 139, 0,
	0, 0, 141, 0, 0, 0, 0, 287, 286, 121,
	122, 285, 0, 0, 0, 0, 0, 0, 0, 0,
	130, 0, 0, 136, 0, 0, 0, 0, 0, 135,
	78, 123, 124, 125, 0, 0, 0, 0, 0, 174,
	0, 0, 0, 134, 126, 129, 127, 128, 116, 117,
	142, 143, 144, 145, 146, 147, 148, 149, 150, 0,
	0, 0, 0, 0, 0, 0, 0, 118, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 138, 137, 0, 0, 0, 0,
	0, 0, 18, 0, 0, 0, 0, 0, 0, 0,
	132, 133, 0, 0, 0, 139, 0, 121, 122, 141,
	0, 0, 0, 120, 0, 0, 0, 0, 130, 0,
	0, 136, 0, 0, 0, 0, 0, 0, 78, 123,
	124, 125, 0, 0, 0, 0, 135, 109, 0, 0,
	0, 134, 126, 129, 127, 128, 116, 117, 142, 143,
	144, 145, 146, 147, 148, 149, 150, 0, 0, 0,
	0, 0, 0, 0, 0, 118, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 138, 137, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 108, 0, 0, 0, 132, 133,
	0, 0, 0, 139, 0, 121, 122, 141, 0, 0,
	0, 120, 0, 0, 0, 0, 130, 0, 0, 136,
	0, 0, 0, 0, 0, 0, 78, 123, 124, 125,
	0, 0, 0, 0, 135, 109, 0, 0, 0, 134,
	126, 129, 127, 128, 116, 117, 142, 143, 144, 145,
	146, 147, 148, 149, 150, 0, 0, 0, 0, 0,
	0, 0, 0, 118, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	138, 137, 0, 0, 18, 0, 0, 0, 0, 0,
	0, 0, 108, 0, 0, 0, 132, 133, 0, 121,
	122, 139, 0, 0, 0, 141, 0, 0, 0, 0,
	130, 0, 0, 136, 0, 0, 0, 0, 0, 0,
	78, 123, 124, 125, 0, 0, 0, 0, 0, 174,
	0, 0, 135, 134, 126, 129, 127, 128, 116, 117,
	142, 143, 144, 145, 146, 147, 148, 149, 150, 0,
	0, 0, 0, 0, 0, 0, 0, 118, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 138, 137, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	132, 133, 0, 121, 122, 139, 0, 0, 0, 141,
	0, 0, 0, 0, 130, 0, 0, 136, 0, 0,
	0, 0, 0, 0, 78, 123, 124, 125, 0, 0,
	0, 0, 0, 174, 0, 0, 135, 134, 126, 129,
	127, 128, 116, 117, 142, 143, 144, 145, 146, 147,
	148, 149, 150, 0, 0, 0, 0, 0, 0, 0,
	0, 118, 193, 197, 195, 196, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 138, 137,
	0, 211, 212, 213, 214, 0, 208, 209, 210, 0,
	0, 0, 0, 0, 132, 133, 0, 0, 0, 139,
	0, 0, 0, 141, 142, 143, 144, 145, 146, 147,
	148, 149, 150, 315, 316, 317, 318, 319, 320, 321,
	322, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	135, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 194, 198, 199, 200, 201, 202, 203, 204,
	206, 207, 205,
}
var yyPact = [...]int{

	179, -1000, -1000, 201, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -75, -76, -56, -68, -1000, -1000, -1000,
	-1000, 13, 303, 467, 403, -1000, -1000, -1000, 401, -1000,
	349, 320, 444, 51, -90, -64, 303, -1000, -58, 303,
	-1000, 339, -96, 303, -96, 348, 389, 389, 443, 303,
	-48, -1000, 266, -1000, -1000, -1000, 572, -1000, 276, 320,
	302, 10, 320, 77, 338, -1000, 193, -1000, 8, 337,
	-2, 303, -1000, 327, -1000, -84, 326, 377, 34, 303,
	320, -1000, 875, 1063, -1000, 389, 1063, 443, 396, 1063,
	164, -1000, -1000, 147, 0, -1000, 1115, -1000, 875, 777,
	-1000, -1000, -1000, 1063, 265, 252, 251, 247, 246, -1000,
	245, -1000, -1000, -1000, -1000, -1000, 298, 295, 294, 283,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, 1063, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 285, 304, 319, 441, 304, -1000, 1063, 303,
	-1000, 376, -103, -1000, 137, -1000, 317, -1000, -1000, 316,
	-1000, 223, 36, 354, 969, -1000, 354, 389, 392, 1063,
	1063, -26, 354, 37, 572, 382, 875, 875, 875, -1000,
	303, 50, 679, 220, 366, 1063, 1063, 135, 1063, 1063,
	1063, 1063, 1063, 1063, 1063, 1063, 1063, 1063, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -109, -27, -52, 36,
	1115, -1000, 474, 572, 1123, 1123, 572, -1000, 467, -1000,
	-1000, -1000, -1000, -24, 354, 386, 304, 304, 211, -1000,
	430, 875, -1000, 354, -1000, -1000, -1000, 31, 303, -1000,
	-87, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 386,
	304, -1000, -1000, 1063, 1063, 354, 354, -1000, 1063, 180,
	233, 297, 89, -7, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 354, 245, 245, 245, -1000, 220,
	1063, 1063, 354, 445, -1000, 355, 116, 116, 116, 167,
	167, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-28, 572, -29, 133, 132, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, 125, 124, 108, -13, -1000, 875, 28,
	220, 201, 40, -30, -1000, 430, 412, 415, 36, 313,
	-1000, -1000, 308, -1000, -1000, 77, 354, 354, 354, 439,
	37, 37, -1000, -1000, 66, 58, 75, 74, 62, -21,
	-1000, 306, -65, 86, -1000, -1000, -1000, -1000, 354, 300,
	1063, -1000, -1000, -33, -1000, 572, 572, 572, 572, 563,
	-31, -1000, 1063, -34, 323, -1000, 346, 105, -1000, -1000,
	-1000, 304, 412, -1000, 1063, 875, -1000, -1000, 432, 414,
	233, 26, -1000, 60, -1000, 59, -1000, -1000, -1000, -1000,
	-66, -71, -72, -1000, -1000, -1000, -1000, -1000, 1063, 354,
	-1000, -36, -37, -43, -44, -110, -1000, -1000, -1000, 156,
	149, -1000, -1000, -1000, -1000, -1000, -1000, 354, 1063, 1063,
	344, 220, -1000, -1000, 561, 102, -1000, 207, -1000, 430,
	875, 1063, 875, -1000, -1000, 214, 213, 208, 354, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 354, 354, 453, -1000,
	1063, 1063, 875, -1000, -1000, -1000, 412, 36, 97, -1000,
	303, 303, 303, 304, 354, 354, -1000, 322, -47, -1000,
	-49, -67, 77, -1000, 448, 371, -1000, 303, -1000, -1000,
	-1000, 303, -1000, 303, -1000,
}
var yyPgo = [...]int{

	0, 588, 584, 19, 583, 582, 580, 579, 578, 576,
	575, 574, 426, 573, 572, 571, 11, 27, 568, 567,
	26, 566, 14, 565, 564, 286, 563, 3, 18, 5,
	551, 550, 15, 548, 2, 17, 9, 547, 546, 22,
	24, 545, 12, 544, 8, 527, 526, 16, 525, 523,
	522, 521, 6, 520, 7, 519, 1, 514, 21, 513,
	13, 4, 23, 259, 512, 511, 510, 509, 507, 506,
	0, 10, 503, 501, 496, 493, 491, 490, 176, 42,
	489, 487, 476, 475,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 75, 75, 5, 6, 7, 7, 72, 73,
	74, 77, 80, 80, 81, 81, 81, 82, 82, 76,
	76, 76, 76, 76, 8, 8, 8, 9, 9, 9,
	10, 11, 11, 11, 83, 12, 13, 13, 14, 14,
	14, 14, 14, 15, 15, 16, 16, 17, 17, 17,
	17, 20, 20, 18, 18, 18, 21, 21, 22, 22,
	22, 22, 19, 19, 19, 23, 23, 23, 23, 23,
	23, 23, 23, 23, 24, 24, 24, 24, 24, 24,
	24, 25, 25, 26, 26, 26, 26, 27, 27, 28,
	28, 79, 79, 79, 78, 78, 29, 29, 29, 29,
	29, 29, 30, 30, 30, 30, 30, 30, 30, 30,
	30, 30, 30, 30, 30, 31, 31, 31, 31, 31,
	31, 31, 32, 32, 37, 37, 35, 35, 42, 36,
	36, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 40, 40, 40, 40, 40,
	40, 40, 40, 39, 39, 39, 39, 39, 39, 39,
	39, 39, 41, 41, 41, 41, 41, 41, 41, 41,
	41, 41, 41, 41, 38, 38, 38, 38, 38, 38,
	43, 43, 43, 45, 48, 48, 46, 46, 47, 47,
	49, 49, 44, 44, 33, 33, 33, 33, 33, 33,
	33, 33, 33, 33, 50, 50, 51, 51, 52, 52,
	53, 53, 54, 55, 55, 55, 56, 56, 56, 56,
	57, 57, 57, 58, 58, 59, 59, 60, 60, 61,
	61, 62, 63, 63, 64, 64, 65, 65, 66, 66,
	66, 66, 66, 67, 67, 68, 68, 69, 69, 70,
	71,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 4, 12, 3,
	7, 7, 6, 6, 8, 7, 3, 4, 1, 1,
	1, 5, 2, 2, 0, 2, 2, 0, 1, 3,
	3, 4, 5, 5, 5, 8, 4, 6, 7, 4,
	5, 4, 5, 5, 0, 2, 0, 2, 1, 2,
	1, 1, 1, 0, 1, 1, 3, 1, 2, 3,
	3, 1, 1, 0, 1, 2, 1, 3, 3, 3,
	3, 5, 0, 1, 2, 1, 1, 2, 3, 2,
	3, 2, 2, 2, 1, 3, 1, 1, 1, 3,
	3, 1, 3, 0, 5, 5, 5, 1, 3, 0,
	2, 0, 2, 2, 0, 2, 1, 3, 3, 3,
	2, 3, 3, 3, 4, 4, 4, 4, 3, 4,
	5, 6, 3, 4, 2, 1, 1, 1, 1, 1,
	1, 1, 2, 1, 1, 3, 3, 1, 3, 1,
	3, 1, 1, 1, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 2, 3, 4, 5, 4, 6,
	6, 6, 6, 6, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 2, 1, 2, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 5, 0, 1, 1, 2, 4, 4,
	0, 2, 1, 3, 1, 1, 1, 1, 1, 2,
	2, 2, 2, 1, 0, 3, 0, 2, 0, 3,
	1, 3, 2, 0, 1, 1, 0, 2, 4, 4,
	0, 2, 4, 0, 3, 1, 3, 0, 5, 1,
	3, 3, 0, 2, 0, 3, 0, 1, 1, 1,
	1, 1, 1, 0, 1, 0, 1, 0, 2, 1,
	0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -72, -73, -74, -75, -76, -77, 5, 6,
	7, 8, 39, 138, 139, 141, 140, 125, 126, 127,
	129, 131, 130, -14, 88, 89, 90, 91, -12, -83,
	-12, -12, -12, -12, 142, -68, 144, 148, -65, 144,
	146, 142, 142, 143, 144, -12, 132, 135, 133, 134,
	-82, 136, -70, 41, -3, 22, -15, 23, -13, 35,
	-25, 41, 9, -61, 128, -62, -44, -70, 41, -64,
	147, 143, -70, 142, -70, 41, -63, 147, -70, -63,
	35, -79, 10, 29, -79, -78, 9, -70, 137, 50,
	-16, -17, 113, -20, 41, -29, -34, -30, 107, 50,
	-33, -44, -35, -43, -70, -38, 59, 60, 78, -45,
	26, 20, 21, 42, 43, 44, 55, 57, 58, 56,
	31, -42, 111, 112, 54, 147, 34, 96, 95, 116,
	-39, 120, 61, 62, 63, 64, 65, 66, 67, 68,
	69, 45, -25, 39, 118, -25, 92, 41, 51, 118,
	41, 107, -70, -71, 41, -71, 145, 41, 26, 103,
	-70, -25, -20, -34, 50, -79, -34, -78, -80, 9,
	27, -36, -34, 9, 92, -18, 104, 105, 106, -70,
	25, 118, -31, 27, 107, 29, 30, 28, 108, 109,
	110, 111, 112, 113, 114, 117, 115, 116, 51, 52,
	53, 46, 47, 48, 49, -20, -29, -36, -3, -20,
	-34, -34, 50, 50, 50, 50, 50, -42, 50, 42,
	42, 42, 42, -48, -34, -58, 39, 50, -61, 41,
	-28, 10, -62, -34, -70, -71, 26, -69, 149, -66,
	141, 139, 38, 140, 13, 41, 41, 41, -71, -58,
	39, -79, -81, 9, 27, -34, -34, 150, 92, -21,
	-22, -24, 50, 41, -42, 137, 133, -17, 24, -20,
	-20, -20, -70, 113, -34, 22, 19, 18, -35, 27,
	29, 30, -34, -34, 31, 107, -34, -34, -34, -34,
	-34, -34, -34, -34, -34, -34, 150, 150, 150, 150,
	-16, 23, -16, -39, -40, 70, 71, 72, 73, 74,
	75, 76, 77, -39, -40, -17, -46, -47, 121, -32,
	34, -3, -61, -59, -44, -28, -52, 13, -20, 103,
	-70, -71, -67, 145, -32, -61, -34, -34, -34, -28,
	92, -23, 93, 94, 95, 96, 97, 99, 100, -19,
	41, 25, -22, 118, -42, -42, -42, -35, -34, -34,
	104, 31, 150, -16, 150, 92, 92, 92, 92, 92,
	-49, -47, 123, -29, -34, -60, 103, -37, -35, -60,
	150, 92, -52, -56, 15, 14, 41, 41, -50, 11,
	-22, -22, 93, 98, 93, 98, 93, 93, 93, -26,
	101, 146, 102, 41, 150, 41, 137, 133, 104, -34,
	150, -16, -16, -16, -16, -41, 79, 55, 56, 80,
	81, 82, 83, 84, 85, 86, 124, -34, 122, 122,
	36, 92, -44, -56, -34, -53, -54, -20, -71, -51,
	12, 14, 103, 93, 93, 143, 143, 143, -34, 150,
	150, 150, 150, 150, 87, 87, -34, -34, 37, -35,
	92, 16, 92, -55, 32, 33, -52, -20, -36, -29,
	50, 50, 50, 7, -34, -34, -54, -56, -27, -70,
	-27, -27, -61, -57, 17, 40, 150, 92, 150, 150,
	7, 27, -70, -70, -70,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 275, 266, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 264, 0, 0, 276, 0, 0,
	267, 0, 262, 0, 262, 0, 111, 111, 114, 0,
	0, 38, 0, 279, 19, 59, 0, 64, 55, 0,
	0, 101, 0, 26, 0, 259, 0, 222, 279, 0,
	0, 0, 280, 0, 280, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 111, 0, 114, 0, 0,
	17, 65, 67, 73, 279, 71, 72, 116, 0, 0,
	151, 152, 153, 0, 222, 0, 0, 0, 0, 174,
	0, 224, 225, 226, 227, 228, 0, 0, 0, 0,
	233, 147, 210, 211, 212, 204, 205, 206, 207, 208,
	209, 214, 183, 184, 185, 186, 187, 188, 189, 190,
	191, 57, 253, 0, 0, 109, 0, 27, 0, 0,
	280, 0, 277, 46, 0, 49, 0, 51, 263, 0,
	280, 253, 112, 113, 0, 41, 115, 111, 34, 0,
	0, 0, 149, 0, 0, 68, 0, 0, 0, 74,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 135, 136,
	137, 138, 139, 140, 141, 120, 71, 0, 0, 0,
	-2, 164, 0, 0, 0, 0, 0, 134, 0, 229,
	230, 231, 232, 0, 215, 0, 0, 0, 109, 102,
	238, 0, 260, 261, 223, 44, 265, 0, 0, 280,
	273, 268, 269, 270, 271, 272, 50, 52, 53, 0,
	0, 42, 43, 0, 0, 32, 33, 31, 0, 109,
	76, 82, 0, 94, 96, 97, 98, 66, 69, 117,
	118, 119, 75, 70, 122, 0, 0, 0, 123, 0,
	0, 0, 128, 0, 132, 0, 154, 155, 156, 157,
	158, 159, 160, 161, 162, 163, 121, 146, 148, 165,
	0, 0, 0, 0, 0, 175, 176, 177, 178, 179,
	180, 181, 182, 0, 0, 0, 220, 216, 0, 257,
	0, 143, 257, 0, 255, 238, 246, 0, 110, 0,
	278, 47, 0, 274, 22, 23, 35, 36, 150, 234,
	0, 0, 85, 86, 0, 0, 0, 0, 0, 103,
	83, 0, 0, 0, 124, 125, 126, 127, 129, 0,
	0, 133, 166, 0, 168, 0, 0, 0, 0, 0,
	0, 217, 0, 71, 72, 20, 0, 142, 144, 21,
	254, 0, 246, 25, 0, 0, 280, 48, 236, 0,
	77, 80, 87, 0, 89, 0, 91, 92, 93, 78,
	0, 0, 0, 84, 79, 95, 99, 100, 0, 130,
	167, 0, 0, 0, 0, 0, 192, 193, 194, 195,
	197, 199, 200, 201, 202, 203, 213, 221, 0, 0,
	0, 0, 256, 24, 247, 239, 240, 243, 45, 238,
	0, 0, 0, 88, 90, 0, 0, 0, 131, 169,
	170, 171, 172, 173, 196, 198, 218, 219, 0, 145,
	0, 0, 0, 242, 244, 245, 246, 237, 235, -2,
	0, 0, 0, 0, 248, 249, 241, 250, 0, 107,
	0, 0, 258, 18, 0, 0, 104, 0, 105, 106,
	251, 0, 108, 0, 252,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 115, 108, 3,
	50, 150, 113, 111, 92, 112, 118, 114, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	52, 51, 53, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 110, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 109, 3, 54,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 47, 48, 49, 55, 56,
	57, 58, 59, 60, 61, 62, 63, 64, 65, 66,
	67, 68, 69, 70, 71, 72, 73, 74, 75, 76,
	77, 78, 79, 80, 81, 82, 83, 84, 85, 86,
	87, 88, 89, 90, 91, 93, 94, 95, 96, 97,
	98, 99, 100, 101, 102, 103, 104, 105, 106, 107,
	116, 117, 119, 120, 121, 122, 123, 124, 125, 126,
	127, 128, 129, 130, 131, 132, 133, 134, 135, 136,
	137, 138, 139, 140, 141, 142, 143, 144, 145, 146,
	147, 148, 149,
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
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
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
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
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
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
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
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
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
			yyrcvr.char = -1
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
		//line sql.y:195
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:201
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:221
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 18:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:225
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].valExprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:229
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 20:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:236
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:240
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
		//line sql.y:252
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:256
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
		//line sql.y:269
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 25:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:275
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:281
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:285
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:291
		{
			yyVAL.statement = &Begin{}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:297
		{
			yyVAL.statement = &Commit{}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:303
		{
			yyVAL.statement = &Rollback{}
		}
	case 31:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:309
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].valExprs}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:315
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:319
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 34:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:324
		{
			yyVAL.valExpr = nil
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:328
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:332
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 37:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:337
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:341
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 39:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:348
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:352
		{
			yyVAL.statement = &Show{Section: "variables", LikeOrWhere: yyDollar[3].expr}
		}
	case 41:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:356
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].valExpr, LikeOrWhere: yyDollar[4].expr}
		}
	case 42:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:360
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].valExpr, LikeOrWhere: yyDollar[5].expr}
		}
	case 43:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:364
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].valExpr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].valExpr}
		}
	case 44:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:370
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 45:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:374
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:379
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 47:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:385
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 48:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:389
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 49:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:394
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 50:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:400
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 51:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:406
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 52:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:410
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 53:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:415
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 54:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:420
		{
			SetAllowComments(yylex, true)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:424
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 56:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:430
		{
			yyVAL.bytes2 = nil
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:434
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:440
		{
			yyVAL.str = AST_UNION
		}
	case 59:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:444
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:448
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:452
		{
			yyVAL.str = AST_EXCEPT
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:456
		{
			yyVAL.str = AST_INTERSECT
		}
	case 63:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:461
		{
			yyVAL.str = ""
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:465
		{
			yyVAL.str = AST_DISTINCT
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:471
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:475
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:481
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:485
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:489
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 70:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:493
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:499
		{
			yyVAL.expr = yyDollar[1].boolExpr
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:503
		{
			yyVAL.expr = yyDollar[1].valExpr
		}
	case 73:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:508
		{
			yyVAL.bytes = nil
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:512
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:516
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 76:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:522
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 77:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:526
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:532
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:536
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:540
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 81:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:544
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].boolExpr}
		}
	case 82:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:549
		{
			yyVAL.bytes = nil
		}
	case 83:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:553
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 84:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:557
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 85:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:563
		{
			yyVAL.str = AST_JOIN
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:567
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 87:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:571
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:575
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 89:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:579
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:583
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 91:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:587
		{
			yyVAL.str = AST_JOIN
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:591
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:595
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 94:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:601
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:605
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 96:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:609
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:613
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:617
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:621
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 100:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:625
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 101:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:631
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:635
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 103:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:640
		{
			yyVAL.indexHints = nil
		}
	case 104:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:644
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 105:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:648
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 106:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:652
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 107:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:658
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 108:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:662
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 109:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:667
		{
			yyVAL.expr = nil
		}
	case 110:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:671
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 111:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:676
		{
			yyVAL.expr = nil
		}
	case 112:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:680
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:684
		{
			yyVAL.expr = yyDollar[2].valExpr
		}
	case 114:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:689
		{
			yyVAL.valExpr = nil
		}
	case 115:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:693
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 117:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:700
		{
			yyVAL.boolExpr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 118:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:704
		{
			yyVAL.boolExpr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 119:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:708
		{
			yyVAL.boolExpr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 120:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:712
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 121:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:716
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyDollar[2].boolExpr}
		}
	case 122:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:722
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, Right: yyDollar[3].valExpr}
		}
	case 123:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:726
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 124:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:730
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ALL, Right: yyDollar[4].subquery}
		}
	case 125:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:734
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ANY, Right: yyDollar[4].subquery}
		}
	case 126:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:738
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_SOME, Right: yyDollar[4].subquery}
		}
	case 127:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:742
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 128:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:746
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_LIKE, Right: yyDollar[3].valExpr}
		}
	case 129:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:750
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_LIKE, Right: yyDollar[4].valExpr}
		}
	case 130:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:754
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_BETWEEN, From: yyDollar[3].valExpr, To: yyDollar[5].valExpr}
		}
	case 131:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:758
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].valExpr, To: yyDollar[6].valExpr}
		}
	case 132:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:762
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NULL, Expr: yyDollar[1].valExpr}
		}
	case 133:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:766
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NOT_NULL, Expr: yyDollar[1].valExpr}
		}
	case 134:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:770
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:776
		{
			yyVAL.str = AST_EQ
		}
	case 136:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:780
		{
			yyVAL.str = AST_LT
		}
	case 137:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:784
		{
			yyVAL.str = AST_GT
		}
	case 138:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:788
		{
			yyVAL.str = AST_LE
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:792
		{
			yyVAL.str = AST_GE
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:796
		{
			yyVAL.str = AST_NE
		}
	case 141:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:800
		{
			yyVAL.str = AST_NSE
		}
	case 142:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:806
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:810
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 144:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:816
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:820
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 146:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:826
		{
			yyVAL.tuple = ValTuple(yyDollar[2].valExprs)
		}
	case 147:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:830
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 148:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:836
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 149:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:842
		{
			yyVAL.valExprs = ValExprs{yyDollar[1].valExpr}
		}
	case 150:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:846
		{
			yyVAL.valExprs = append(yyDollar[1].valExprs, yyDollar[3].valExpr)
		}
	case 151:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:852
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 152:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:856
		{
			yyVAL.valExpr = yyDollar[1].colName
		}
	case 153:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:860
		{
			yyVAL.valExpr = yyDollar[1].tuple
		}
	case 154:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:864
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITAND, Right: yyDollar[3].valExpr}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:868
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITOR, Right: yyDollar[3].valExpr}
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:872
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITXOR, Right: yyDollar[3].valExpr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:876
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_PLUS, Right: yyDollar[3].valExpr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:880
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MINUS, Right: yyDollar[3].valExpr}
		}
	case 159:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:884
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MULT, Right: yyDollar[3].valExpr}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:888
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_DIV, Right: yyDollar[3].valExpr}
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:892
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_IDIV, Right: yyDollar[3].valExpr}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:896
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 163:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:900
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 164:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:904
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
	case 165:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:919
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 166:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:923
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 167:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:927
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 168:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:931
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 169:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:935
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 170:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:939
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 171:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:943
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 172:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:947
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 173:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:951
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 174:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:955
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 175:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:961
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 176:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:965
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 177:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:969
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 178:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:973
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 179:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:977
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 180:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:981
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 181:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:985
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 182:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:989
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 183:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:995
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 184:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:999
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 185:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1003
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 186:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1007
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 187:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1011
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1015
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1019
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1027
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1033
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1037
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1041
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 195:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1045
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 196:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1049
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1053
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 198:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1057
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 199:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1061
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1065
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1069
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1073
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1077
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1083
		{
			yyVAL.bytes = IF_BYTES
		}
	case 205:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1087
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1091
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1095
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1099
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 209:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1103
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1109
		{
			yyVAL.byt = AST_UPLUS
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1113
		{
			yyVAL.byt = AST_UMINUS
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1117
		{
			yyVAL.byt = AST_TILDA
		}
	case 213:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1123
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 214:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1128
		{
			yyVAL.valExpr = nil
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1132
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1138
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 217:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1142
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 218:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1148
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 219:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1152
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 220:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1157
		{
			yyVAL.valExpr = nil
		}
	case 221:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1161
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1167
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 223:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1171
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1177
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1181
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1185
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1189
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1193
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 229:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1197
		{
			yyVAL.valExpr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 230:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1201
		{
			yyVAL.valExpr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 231:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1205
		{
			yyVAL.valExpr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 232:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1209
		{
			yyVAL.valExpr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1213
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 234:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1218
		{
			yyVAL.valExprs = nil
		}
	case 235:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1222
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 236:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1227
		{
			yyVAL.expr = nil
		}
	case 237:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1231
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 238:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1236
		{
			yyVAL.orderBy = nil
		}
	case 239:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1240
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1246
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 241:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1250
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 242:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1256
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 243:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1261
		{
			yyVAL.str = AST_ASC
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1265
		{
			yyVAL.str = AST_ASC
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1269
		{
			yyVAL.str = AST_DESC
		}
	case 246:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1274
		{
			yyVAL.limit = nil
		}
	case 247:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1278
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 248:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1282
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 249:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1286
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].valExpr, Rowcount: yyDollar[2].valExpr}
		}
	case 250:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1291
		{
			yyVAL.str = ""
		}
	case 251:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1295
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 252:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1299
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
	case 253:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1312
		{
			yyVAL.columns = nil
		}
	case 254:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1316
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 255:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1322
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 256:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1326
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 257:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1331
		{
			yyVAL.updateExprs = nil
		}
	case 258:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1335
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1341
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 260:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1345
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 261:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1351
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 262:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1356
		{
			yyVAL.empty = struct{}{}
		}
	case 263:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1358
		{
			yyVAL.empty = struct{}{}
		}
	case 264:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1361
		{
			yyVAL.empty = struct{}{}
		}
	case 265:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1363
		{
			yyVAL.empty = struct{}{}
		}
	case 266:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1366
		{
			yyVAL.empty = struct{}{}
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1368
		{
			yyVAL.empty = struct{}{}
		}
	case 268:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1372
		{
			yyVAL.empty = struct{}{}
		}
	case 269:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1374
		{
			yyVAL.empty = struct{}{}
		}
	case 270:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1376
		{
			yyVAL.empty = struct{}{}
		}
	case 271:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1378
		{
			yyVAL.empty = struct{}{}
		}
	case 272:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1380
		{
			yyVAL.empty = struct{}{}
		}
	case 273:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1383
		{
			yyVAL.empty = struct{}{}
		}
	case 274:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1385
		{
			yyVAL.empty = struct{}{}
		}
	case 275:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1388
		{
			yyVAL.empty = struct{}{}
		}
	case 276:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1390
		{
			yyVAL.empty = struct{}{}
		}
	case 277:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1393
		{
			yyVAL.empty = struct{}{}
		}
	case 278:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1395
		{
			yyVAL.empty = struct{}{}
		}
	case 279:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1399
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 280:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1404
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
