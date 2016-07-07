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
const COLUMNS = 57465
const CREATE = 57466
const ALTER = 57467
const DROP = 57468
const RENAME = 57469
const TABLE = 57470
const INDEX = 57471
const VIEW = 57472
const TO = 57473
const IGNORE = 57474
const IF = 57475
const UNIQUE = 57476
const USING = 57477

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
	-1, 222,
	93, 151,
	151, 151,
	-2, 72,
	-1, 483,
	105, 71,
	106, 71,
	107, 71,
	-2, 81,
}

const yyNprod = 285
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1281

var yyAct = [...]int{

	114, 397, 106, 492, 73, 105, 339, 450, 111, 219,
	165, 100, 130, 389, 272, 332, 330, 112, 129, 242,
	501, 220, 3, 75, 467, 139, 103, 101, 237, 317,
	501, 501, 309, 62, 355, 356, 357, 358, 359, 186,
	360, 361, 91, 186, 77, 186, 250, 82, 186, 186,
	84, 87, 76, 395, 88, 64, 34, 35, 36, 37,
	97, 80, 186, 186, 270, 270, 44, 49, 46, 50,
	346, 168, 47, 52, 53, 54, 461, 460, 503, 414,
	416, 78, 164, 459, 275, 81, 83, 51, 502, 500,
	172, 418, 98, 274, 419, 167, 175, 466, 440, 178,
	94, 465, 184, 464, 191, 256, 463, 424, 331, 183,
	386, 394, 222, 442, 163, 218, 223, 331, 311, 174,
	378, 376, 310, 269, 415, 56, 58, 59, 57, 61,
	366, 254, 18, 229, 257, 217, 221, 193, 177, 161,
	422, 156, 456, 236, 200, 201, 202, 203, 204, 205,
	206, 208, 209, 207, 390, 158, 77, 342, 63, 77,
	240, 171, 246, 245, 76, 192, 390, 76, 74, 275,
	188, 189, 190, 247, 95, 458, 278, 457, 274, 184,
	277, 63, 244, 260, 267, 268, 421, 141, 142, 143,
	420, 408, 412, 284, 246, 411, 409, 286, 276, 296,
	294, 295, 261, 299, 300, 301, 302, 303, 304, 305,
	306, 307, 308, 290, 279, 281, 282, 283, 298, 410,
	478, 479, 263, 203, 204, 205, 206, 208, 209, 207,
	285, 158, 253, 255, 252, 270, 313, 315, 243, 77,
	77, 243, 476, 335, 188, 189, 190, 76, 337, 406,
	445, 343, 316, 326, 407, 185, 328, 327, 383, 334,
	338, 278, 344, 77, 382, 277, 381, 348, 349, 350,
	341, 76, 179, 351, 380, 297, 379, 347, 205, 206,
	208, 209, 207, 334, 475, 469, 468, 276, 86, 365,
	160, 352, 188, 189, 190, 371, 372, 153, 486, 485,
	367, 368, 369, 484, 355, 356, 357, 358, 359, 370,
	360, 361, 34, 35, 36, 37, 375, 18, 19, 20,
	21, 353, 262, 176, 158, 238, 377, 230, 228, 227,
	226, 225, 70, 239, 388, 224, 239, 387, 99, 186,
	364, 234, 233, 89, 232, 396, 385, 231, 63, 393,
	498, 392, 22, 78, 472, 417, 363, 401, 221, 400,
	18, 474, 259, 258, 241, 71, 276, 276, 404, 405,
	169, 166, 162, 159, 499, 155, 423, 200, 201, 202,
	203, 204, 205, 206, 208, 209, 207, 85, 444, 441,
	333, 425, 426, 427, 428, 90, 77, 69, 447, 505,
	248, 448, 154, 265, 446, 157, 280, 291, 373, 292,
	293, 452, 200, 201, 202, 203, 204, 205, 206, 208,
	209, 207, 266, 173, 170, 462, 451, 200, 201, 202,
	203, 204, 205, 206, 208, 209, 207, 181, 27, 28,
	29, 92, 30, 32, 31, 470, 471, 141, 142, 143,
	67, 23, 24, 26, 25, 65, 182, 398, 184, 374,
	480, 93, 483, 473, 455, 482, 399, 340, 454, 403,
	243, 96, 72, 504, 487, 18, 39, 488, 489, 60,
	264, 481, 491, 221, 490, 493, 493, 493, 77, 494,
	495, 180, 496, 17, 16, 38, 76, 141, 142, 143,
	15, 314, 506, 451, 120, 14, 507, 13, 508, 128,
	12, 249, 135, 45, 345, 40, 41, 42, 43, 104,
	121, 122, 123, 251, 431, 432, 55, 48, 109, 79,
	336, 497, 133, 124, 127, 125, 126, 116, 117, 144,
	145, 146, 147, 148, 149, 150, 151, 152, 430, 433,
	434, 435, 436, 437, 438, 439, 118, 477, 449, 453,
	402, 384, 235, 329, 119, 113, 429, 115, 391, 110,
	194, 107, 413, 137, 136, 273, 354, 271, 362, 187,
	66, 33, 68, 11, 10, 108, 9, 8, 7, 131,
	132, 102, 6, 5, 138, 141, 142, 143, 140, 4,
	2, 1, 120, 0, 0, 0, 0, 128, 0, 0,
	135, 0, 0, 0, 0, 0, 0, 104, 121, 122,
	123, 0, 0, 0, 0, 134, 109, 0, 312, 0,
	133, 124, 127, 125, 126, 116, 117, 144, 145, 146,
	147, 148, 149, 150, 151, 152, 0, 0, 0, 0,
	0, 0, 0, 0, 118, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 137, 136, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 108, 0, 0, 0, 131, 132, 102,
	0, 0, 138, 0, 0, 0, 140, 0, 0, 0,
	289, 288, 141, 142, 143, 287, 0, 0, 0, 0,
	0, 0, 0, 0, 128, 0, 0, 135, 0, 0,
	0, 0, 0, 134, 78, 121, 122, 123, 0, 0,
	0, 0, 0, 176, 0, 0, 0, 133, 124, 127,
	125, 126, 116, 117, 144, 145, 146, 147, 148, 149,
	150, 151, 152, 0, 0, 0, 0, 0, 0, 0,
	0, 118, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 137, 136,
	0, 0, 0, 0, 0, 18, 0, 0, 0, 0,
	0, 0, 0, 0, 131, 132, 0, 0, 0, 138,
	141, 142, 143, 140, 0, 0, 0, 120, 0, 0,
	0, 0, 128, 0, 0, 135, 0, 0, 0, 0,
	0, 0, 78, 121, 122, 123, 0, 0, 0, 0,
	134, 109, 0, 0, 0, 133, 124, 127, 125, 126,
	116, 117, 144, 145, 146, 147, 148, 149, 150, 151,
	152, 0, 0, 0, 0, 0, 0, 0, 0, 118,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 137, 136, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 108, 0,
	0, 0, 131, 132, 0, 0, 0, 138, 141, 142,
	143, 140, 0, 0, 0, 120, 0, 0, 0, 0,
	128, 0, 0, 135, 0, 0, 0, 0, 0, 0,
	78, 121, 122, 123, 0, 0, 0, 0, 134, 109,
	0, 0, 0, 133, 124, 127, 125, 126, 116, 117,
	144, 145, 146, 147, 148, 149, 150, 151, 152, 0,
	0, 0, 0, 0, 0, 0, 0, 118, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 137, 136, 0, 18, 0, 0,
	0, 0, 0, 0, 0, 0, 108, 0, 0, 0,
	131, 132, 141, 142, 143, 138, 0, 0, 0, 140,
	0, 0, 0, 0, 128, 0, 0, 135, 0, 0,
	0, 0, 0, 0, 78, 121, 122, 123, 0, 0,
	0, 0, 0, 176, 0, 0, 134, 133, 124, 127,
	125, 126, 116, 117, 144, 145, 146, 147, 148, 149,
	150, 151, 152, 0, 0, 0, 0, 0, 0, 0,
	0, 118, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 137, 136,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 131, 132, 141, 142, 143, 138,
	0, 0, 0, 140, 0, 0, 0, 0, 128, 0,
	0, 135, 0, 0, 0, 0, 0, 0, 78, 121,
	122, 123, 0, 0, 0, 0, 0, 176, 0, 0,
	134, 133, 124, 127, 125, 126, 116, 117, 144, 145,
	146, 147, 148, 149, 150, 151, 152, 0, 0, 0,
	0, 0, 0, 0, 0, 118, 195, 199, 197, 198,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 137, 136, 0, 213, 214, 215, 216, 0,
	210, 211, 212, 195, 199, 197, 198, 0, 131, 132,
	0, 0, 0, 138, 0, 0, 0, 140, 0, 0,
	0, 0, 213, 214, 215, 216, 0, 210, 211, 212,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 134, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 196, 200, 201, 202,
	203, 204, 205, 206, 208, 209, 207, 0, 0, 0,
	0, 443, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 196, 200, 201, 202, 203, 204, 205,
	206, 208, 209, 207, 144, 145, 146, 147, 148, 149,
	150, 151, 152, 318, 319, 320, 321, 322, 323, 324,
	325,
}
var yyPact = [...]int{

	312, -1000, -1000, 223, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -77, -78, -56, -70, -1000, -1000, -1000,
	-1000, -8, 306, 470, 432, -1000, -1000, -1000, 426, -1000,
	361, 323, 463, 39, -87, -59, 306, -1000, -57, 306,
	-1000, 345, -97, 306, -97, 359, 431, 431, 462, 306,
	-46, -1000, 287, -1000, -1000, -1000, 575, -1000, 251, 323,
	335, 22, 323, 138, 331, -1000, 238, -1000, 20, 330,
	6, 306, -1000, 329, -1000, -75, 328, 397, 57, 306,
	323, -1000, 878, 1066, -1000, 431, 1066, 462, 428, 1066,
	246, -1000, -1000, 139, 18, -1000, 1145, -1000, 878, 780,
	-1000, -1000, -1000, 1066, 284, 280, 279, 278, 277, -1000,
	276, -1000, -1000, -1000, 304, 301, 299, 298, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	1066, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 285, 311, 322, 460, 311, -1000,
	1066, 306, -1000, 373, -104, -1000, 92, -1000, 321, -1000,
	-1000, 320, -1000, 282, 65, 318, 972, -1000, 318, 431,
	394, 1066, 1066, -28, 318, 42, 575, 381, 878, 878,
	878, -1000, 306, 116, 682, 272, 379, 1066, 1066, 167,
	1066, 1066, 1066, 1066, 1066, 1066, 1066, 1066, 1066, 1066,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -119, -29,
	-33, 65, 1145, -1000, 477, 575, 1202, 1202, 575, -1000,
	470, -1000, -1000, -1000, -1000, -5, 318, 355, 311, 311,
	231, -1000, 454, 878, -1000, 318, -1000, -1000, -1000, 53,
	306, -1000, -76, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, 355, 311, -1000, -1000, 1066, 1066, 318, 318, -1000,
	1066, 228, 210, 314, 127, 11, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 318, 276, 276, 276,
	-1000, 272, 1066, 1066, 318, 303, -1000, 427, -1000, 111,
	111, 111, 164, 164, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -30, 575, -31, 183, 181, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 173, 171, 165, -14,
	-1000, 878, 50, 272, 223, 62, -40, -1000, 454, 442,
	452, 65, 317, -1000, -1000, 315, -1000, -1000, 138, 318,
	318, 318, 458, 42, 42, -1000, -1000, 155, 97, 125,
	101, 98, -23, -1000, 313, -60, 52, -1000, -1000, -1000,
	-1000, 318, 35, 1066, -1000, -1000, -1000, -44, -1000, 575,
	575, 575, 575, 468, -27, -1000, 1066, -10, 1118, -1000,
	351, 157, -1000, -1000, -1000, 311, 442, -1000, 1066, 878,
	-1000, -1000, 456, 450, 210, 38, -1000, 83, -1000, 81,
	-1000, -1000, -1000, -1000, -61, -67, -68, -1000, -1000, -1000,
	-1000, -1000, 1066, 318, -1000, -45, -48, -50, -54, -127,
	-1000, -1000, -1000, 198, 197, -1000, -1000, -1000, -1000, -1000,
	-1000, 318, 1066, 1066, 316, 272, -1000, -1000, 268, 149,
	-1000, 187, -1000, 454, 878, 1066, 878, -1000, -1000, 252,
	248, 247, 318, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	318, 318, 467, -1000, 1066, 1066, 878, -1000, -1000, -1000,
	442, 65, 142, -1000, 306, 306, 306, 311, 318, 318,
	-1000, 333, -62, -1000, -63, -73, 138, -1000, 466, 371,
	-1000, 306, -1000, -1000, -1000, 306, -1000, 306, -1000,
}
var yyPgo = [...]int{

	0, 601, 600, 21, 599, 593, 592, 588, 587, 586,
	584, 583, 495, 582, 581, 580, 11, 27, 579, 578,
	26, 577, 14, 576, 575, 332, 572, 3, 19, 5,
	571, 570, 15, 569, 2, 17, 9, 18, 568, 567,
	25, 29, 566, 12, 565, 8, 564, 563, 16, 562,
	561, 560, 559, 6, 558, 7, 557, 1, 531, 28,
	530, 13, 4, 23, 288, 529, 527, 523, 514, 513,
	511, 0, 10, 510, 507, 505, 500, 494, 493, 174,
	42, 491, 480, 479, 476,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 76, 76, 5, 6, 7, 7, 73, 74,
	75, 78, 81, 81, 82, 82, 82, 83, 83, 77,
	77, 77, 77, 77, 8, 8, 8, 9, 9, 9,
	10, 11, 11, 11, 84, 12, 13, 13, 14, 14,
	14, 14, 14, 15, 15, 16, 16, 17, 17, 17,
	17, 20, 20, 18, 18, 18, 21, 21, 22, 22,
	22, 22, 19, 19, 19, 23, 23, 23, 23, 23,
	23, 23, 23, 23, 24, 24, 24, 24, 24, 24,
	24, 25, 25, 26, 26, 26, 26, 27, 27, 28,
	28, 80, 80, 80, 79, 79, 29, 29, 29, 29,
	29, 29, 30, 30, 30, 30, 30, 30, 30, 30,
	30, 30, 30, 30, 30, 30, 30, 31, 31, 31,
	31, 31, 31, 31, 32, 32, 38, 38, 35, 35,
	43, 36, 36, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 41, 41, 41,
	41, 41, 41, 41, 41, 40, 40, 40, 40, 40,
	40, 40, 40, 40, 42, 42, 42, 42, 42, 42,
	42, 42, 42, 42, 42, 42, 39, 39, 39, 39,
	39, 39, 44, 44, 44, 46, 49, 49, 47, 47,
	48, 48, 50, 50, 45, 45, 33, 33, 33, 33,
	33, 33, 33, 33, 33, 37, 37, 37, 51, 51,
	52, 52, 53, 53, 54, 54, 55, 56, 56, 56,
	57, 57, 57, 57, 58, 58, 58, 59, 59, 60,
	60, 61, 61, 62, 62, 63, 64, 64, 65, 65,
	66, 66, 67, 67, 67, 67, 67, 68, 68, 69,
	69, 70, 70, 71, 72,
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
	5, 6, 3, 4, 2, 3, 4, 1, 1, 1,
	1, 1, 1, 1, 2, 1, 1, 3, 3, 1,
	3, 1, 3, 1, 1, 1, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 2, 3, 4, 5,
	4, 6, 6, 6, 6, 6, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 2, 1,
	2, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 5, 0, 1, 1, 2,
	4, 4, 0, 2, 1, 3, 1, 1, 1, 2,
	2, 2, 2, 1, 1, 1, 1, 1, 0, 3,
	0, 2, 0, 3, 1, 3, 2, 0, 1, 1,
	0, 2, 4, 4, 0, 2, 4, 0, 3, 1,
	3, 0, 5, 1, 3, 3, 0, 2, 0, 3,
	0, 1, 1, 1, 1, 1, 1, 0, 1, 0,
	1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -73, -74, -75, -76, -77, -78, 5, 6,
	7, 8, 40, 139, 140, 142, 141, 126, 127, 128,
	130, 132, 131, -14, 89, 90, 91, 92, -12, -84,
	-12, -12, -12, -12, 143, -69, 145, 149, -66, 145,
	147, 143, 143, 144, 145, -12, 133, 136, 134, 135,
	-83, 137, -71, 42, -3, 23, -15, 24, -13, 36,
	-25, 42, 9, -62, 129, -63, -45, -71, 42, -65,
	148, 144, -71, 143, -71, 42, -64, 148, -71, -64,
	36, -80, 10, 30, -80, -79, 9, -71, 138, 51,
	-16, -17, 114, -20, 42, -29, -34, -30, 108, 51,
	-33, -45, -35, -44, -71, -39, 60, 61, 79, -46,
	27, 43, 44, 45, 56, 58, 59, 57, 32, -37,
	-43, 112, 113, 55, 148, 35, 97, 96, 117, -40,
	121, 20, 21, 22, 62, 63, 64, 65, 66, 67,
	68, 69, 70, 46, -25, 40, 119, -25, 93, 42,
	52, 119, 42, 108, -71, -72, 42, -72, 146, 42,
	27, 104, -71, -25, -20, -34, 51, -80, -34, -79,
	-81, 9, 28, -36, -34, 9, 93, -18, 105, 106,
	107, -71, 26, 119, -31, 28, 108, 30, 31, 29,
	109, 110, 111, 112, 113, 114, 115, 118, 116, 117,
	52, 53, 54, 47, 48, 49, 50, -20, -29, -36,
	-3, -20, -34, -34, 51, 51, 51, 51, 51, -43,
	51, 43, 43, 43, 43, -49, -34, -59, 40, 51,
	-62, 42, -28, 10, -63, -34, -71, -72, 27, -70,
	150, -67, 142, 140, 39, 141, 13, 42, 42, 42,
	-72, -59, 40, -80, -82, 9, 28, -34, -34, 151,
	93, -21, -22, -24, 51, 42, -43, 138, 134, -17,
	25, -20, -20, -20, -71, 114, -34, 23, 19, 18,
	-35, 28, 30, 31, -34, -34, 32, 108, -37, -34,
	-34, -34, -34, -34, -34, -34, -34, -34, -34, 151,
	151, 151, 151, -16, 24, -16, -40, -41, 71, 72,
	73, 74, 75, 76, 77, 78, -40, -41, -17, -47,
	-48, 122, -32, 35, -3, -62, -60, -45, -28, -53,
	13, -20, 104, -71, -72, -68, 146, -32, -62, -34,
	-34, -34, -28, 93, -23, 94, 95, 96, 97, 98,
	100, 101, -19, 42, 26, -22, 119, -43, -43, -43,
	-35, -34, -34, 105, 32, -37, 151, -16, 151, 93,
	93, 93, 93, 93, -50, -48, 124, -29, -34, -61,
	104, -38, -35, -61, 151, 93, -53, -57, 15, 14,
	42, 42, -51, 11, -22, -22, 94, 99, 94, 99,
	94, 94, 94, -26, 102, 147, 103, 42, 151, 42,
	138, 134, 105, -34, 151, -16, -16, -16, -16, -42,
	80, 56, 57, 81, 82, 83, 84, 85, 86, 87,
	125, -34, 123, 123, 37, 93, -45, -57, -34, -54,
	-55, -20, -72, -52, 12, 14, 104, 94, 94, 144,
	144, 144, -34, 151, 151, 151, 151, 151, 88, 88,
	-34, -34, 38, -35, 93, 16, 93, -56, 33, 34,
	-53, -20, -36, -29, 51, 51, 51, 7, -34, -34,
	-55, -57, -27, -71, -27, -27, -62, -58, 17, 41,
	151, 93, 151, 151, 7, 28, -71, -71, -71,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 279, 270, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 268, 0, 0, 280, 0, 0,
	271, 0, 266, 0, 266, 0, 111, 111, 114, 0,
	0, 38, 0, 283, 19, 59, 0, 64, 55, 0,
	0, 101, 0, 26, 0, 263, 0, 224, 283, 0,
	0, 0, 284, 0, 284, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 111, 0, 114, 0, 0,
	17, 65, 67, 73, 283, 71, 72, 116, 0, 0,
	153, 154, 155, 0, 224, 0, 0, 0, 0, 176,
	0, 226, 227, 228, 0, 0, 0, 0, 233, 234,
	149, 212, 213, 214, 206, 207, 208, 209, 210, 211,
	216, 235, 236, 237, 185, 186, 187, 188, 189, 190,
	191, 192, 193, 57, 257, 0, 0, 109, 0, 27,
	0, 0, 284, 0, 281, 46, 0, 49, 0, 51,
	267, 0, 284, 257, 112, 113, 0, 41, 115, 111,
	34, 0, 0, 0, 151, 0, 0, 68, 0, 0,
	0, 74, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	137, 138, 139, 140, 141, 142, 143, 120, 71, 0,
	0, 0, -2, 166, 0, 0, 0, 0, 0, 134,
	0, 229, 230, 231, 232, 0, 217, 0, 0, 0,
	109, 102, 242, 0, 264, 265, 225, 44, 269, 0,
	0, 284, 277, 272, 273, 274, 275, 276, 50, 52,
	53, 0, 0, 42, 43, 0, 0, 32, 33, 31,
	0, 109, 76, 82, 0, 94, 96, 97, 98, 66,
	69, 117, 118, 119, 75, 70, 122, 0, 0, 0,
	123, 0, 0, 0, 128, 0, 132, 0, 135, 156,
	157, 158, 159, 160, 161, 162, 163, 164, 165, 121,
	148, 150, 167, 0, 0, 0, 0, 0, 177, 178,
	179, 180, 181, 182, 183, 184, 0, 0, 0, 222,
	218, 0, 261, 0, 145, 261, 0, 259, 242, 250,
	0, 110, 0, 282, 47, 0, 278, 22, 23, 35,
	36, 152, 238, 0, 0, 85, 86, 0, 0, 0,
	0, 0, 103, 83, 0, 0, 0, 124, 125, 126,
	127, 129, 0, 0, 133, 136, 168, 0, 170, 0,
	0, 0, 0, 0, 0, 219, 0, 71, 72, 20,
	0, 144, 146, 21, 258, 0, 250, 25, 0, 0,
	284, 48, 240, 0, 77, 80, 87, 0, 89, 0,
	91, 92, 93, 78, 0, 0, 0, 84, 79, 95,
	99, 100, 0, 130, 169, 0, 0, 0, 0, 0,
	194, 195, 196, 197, 199, 201, 202, 203, 204, 205,
	215, 223, 0, 0, 0, 0, 260, 24, 251, 243,
	244, 247, 45, 242, 0, 0, 0, 88, 90, 0,
	0, 0, 131, 171, 172, 173, 174, 175, 198, 200,
	220, 221, 0, 147, 0, 0, 0, 246, 248, 249,
	250, 241, 239, -2, 0, 0, 0, 0, 252, 253,
	245, 254, 0, 107, 0, 0, 262, 18, 0, 0,
	104, 0, 105, 106, 255, 0, 108, 0, 256,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 116, 109, 3,
	51, 151, 114, 112, 93, 113, 119, 115, 3, 3,
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
	147, 148, 149, 150,
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
		//line sql.y:196
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:202
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:222
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 18:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:226
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].valExprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:230
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 20:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:237
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:241
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
		//line sql.y:253
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:257
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
		//line sql.y:270
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 25:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:276
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:282
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:286
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:292
		{
			yyVAL.statement = &Begin{}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:298
		{
			yyVAL.statement = &Commit{}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:304
		{
			yyVAL.statement = &Rollback{}
		}
	case 31:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:310
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].valExprs}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:316
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:320
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 34:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:325
		{
			yyVAL.valExpr = nil
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:329
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:333
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 37:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:338
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:342
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 39:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:349
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:353
		{
			yyVAL.statement = &Show{Section: "variables", LikeOrWhere: yyDollar[3].expr}
		}
	case 41:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:357
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].valExpr, LikeOrWhere: yyDollar[4].expr}
		}
	case 42:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:361
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].valExpr, LikeOrWhere: yyDollar[5].expr}
		}
	case 43:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:365
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].valExpr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].valExpr}
		}
	case 44:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:371
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 45:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:375
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:380
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 47:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:386
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 48:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:390
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 49:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:395
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 50:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:401
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 51:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:407
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 52:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:411
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 53:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:416
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 54:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:421
		{
			SetAllowComments(yylex, true)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:425
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 56:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:431
		{
			yyVAL.bytes2 = nil
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:435
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:441
		{
			yyVAL.str = AST_UNION
		}
	case 59:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:445
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:449
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:453
		{
			yyVAL.str = AST_EXCEPT
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:457
		{
			yyVAL.str = AST_INTERSECT
		}
	case 63:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:462
		{
			yyVAL.str = ""
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:466
		{
			yyVAL.str = AST_DISTINCT
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:472
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:476
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:482
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:486
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:490
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 70:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:494
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:500
		{
			yyVAL.expr = yyDollar[1].boolExpr
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:504
		{
			yyVAL.expr = yyDollar[1].valExpr
		}
	case 73:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:509
		{
			yyVAL.bytes = nil
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:513
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:517
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 76:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:523
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 77:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:527
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:533
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:537
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:541
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 81:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:545
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].boolExpr}
		}
	case 82:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:550
		{
			yyVAL.bytes = nil
		}
	case 83:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:554
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 84:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:558
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 85:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:564
		{
			yyVAL.str = AST_JOIN
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:568
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 87:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:572
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:576
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 89:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:580
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:584
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 91:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:588
		{
			yyVAL.str = AST_JOIN
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:592
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:596
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 94:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:602
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:606
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 96:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:610
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:614
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:618
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:622
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 100:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:626
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 101:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:632
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:636
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 103:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:641
		{
			yyVAL.indexHints = nil
		}
	case 104:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:645
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 105:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:649
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 106:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:653
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 107:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:659
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 108:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:663
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 109:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:668
		{
			yyVAL.expr = nil
		}
	case 110:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:672
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 111:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:677
		{
			yyVAL.expr = nil
		}
	case 112:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:681
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:685
		{
			yyVAL.expr = yyDollar[2].valExpr
		}
	case 114:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:690
		{
			yyVAL.valExpr = nil
		}
	case 115:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:694
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 117:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:701
		{
			yyVAL.boolExpr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 118:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:705
		{
			yyVAL.boolExpr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 119:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:709
		{
			yyVAL.boolExpr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 120:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:713
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 121:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:717
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyDollar[2].boolExpr}
		}
	case 122:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:723
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, Right: yyDollar[3].valExpr}
		}
	case 123:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:727
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 124:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:731
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ALL, Right: yyDollar[4].subquery}
		}
	case 125:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:735
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ANY, Right: yyDollar[4].subquery}
		}
	case 126:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:739
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_SOME, Right: yyDollar[4].subquery}
		}
	case 127:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:743
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 128:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:747
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_LIKE, Right: yyDollar[3].valExpr}
		}
	case 129:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:751
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_LIKE, Right: yyDollar[4].valExpr}
		}
	case 130:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:755
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_BETWEEN, From: yyDollar[3].valExpr, To: yyDollar[5].valExpr}
		}
	case 131:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:759
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].valExpr, To: yyDollar[6].valExpr}
		}
	case 132:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:763
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 133:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:767
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 134:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:771
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 135:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:775
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS, Right: yyDollar[3].valExpr}
		}
	case 136:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:779
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS_NOT, Right: yyDollar[4].valExpr}
		}
	case 137:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:785
		{
			yyVAL.str = AST_EQ
		}
	case 138:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:789
		{
			yyVAL.str = AST_LT
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:793
		{
			yyVAL.str = AST_GT
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:797
		{
			yyVAL.str = AST_LE
		}
	case 141:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:801
		{
			yyVAL.str = AST_GE
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:805
		{
			yyVAL.str = AST_NE
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:809
		{
			yyVAL.str = AST_NSE
		}
	case 144:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:815
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 145:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:819
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:825
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:829
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 148:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:835
		{
			yyVAL.tuple = ValTuple(yyDollar[2].valExprs)
		}
	case 149:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:839
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 150:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:845
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 151:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:851
		{
			yyVAL.valExprs = ValExprs{yyDollar[1].valExpr}
		}
	case 152:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:855
		{
			yyVAL.valExprs = append(yyDollar[1].valExprs, yyDollar[3].valExpr)
		}
	case 153:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:861
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 154:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:865
		{
			yyVAL.valExpr = yyDollar[1].colName
		}
	case 155:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:869
		{
			yyVAL.valExpr = yyDollar[1].tuple
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:873
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITAND, Right: yyDollar[3].valExpr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:877
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITOR, Right: yyDollar[3].valExpr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:881
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITXOR, Right: yyDollar[3].valExpr}
		}
	case 159:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:885
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_PLUS, Right: yyDollar[3].valExpr}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:889
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MINUS, Right: yyDollar[3].valExpr}
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:893
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MULT, Right: yyDollar[3].valExpr}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:897
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_DIV, Right: yyDollar[3].valExpr}
		}
	case 163:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:901
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_IDIV, Right: yyDollar[3].valExpr}
		}
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:905
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 165:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:909
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 166:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:913
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
	case 167:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:928
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 168:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:932
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 169:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:936
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 170:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:940
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 171:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:944
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 172:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:948
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 173:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:952
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 174:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:956
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 175:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:960
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 176:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:964
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 177:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:970
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 178:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:974
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 179:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:978
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 180:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:982
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 181:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:986
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 182:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:990
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 183:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:994
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 184:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:998
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 185:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1004
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 186:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1008
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 187:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1012
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1016
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1020
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1024
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1028
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1032
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1036
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1042
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 195:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1046
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 196:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1050
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1054
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 198:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1058
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 199:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1062
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 200:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1066
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1070
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1074
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1078
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1082
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 205:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1086
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1092
		{
			yyVAL.bytes = IF_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1096
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1100
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 209:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1104
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1108
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1112
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1118
		{
			yyVAL.byt = AST_UPLUS
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1122
		{
			yyVAL.byt = AST_UMINUS
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1126
		{
			yyVAL.byt = AST_TILDA
		}
	case 215:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1132
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 216:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1137
		{
			yyVAL.valExpr = nil
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1141
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1147
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 219:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1151
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 220:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1157
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 221:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1161
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 222:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1166
		{
			yyVAL.valExpr = nil
		}
	case 223:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1170
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1176
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 225:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1180
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1186
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1190
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1194
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 229:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1198
		{
			yyVAL.valExpr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 230:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1202
		{
			yyVAL.valExpr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 231:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1206
		{
			yyVAL.valExpr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 232:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1210
		{
			yyVAL.valExpr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1214
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1218
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1224
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1228
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1232
		{
			yyVAL.valExpr = &UnknownVal{}
		}
	case 238:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1237
		{
			yyVAL.valExprs = nil
		}
	case 239:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1241
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 240:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1246
		{
			yyVAL.expr = nil
		}
	case 241:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1250
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 242:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1255
		{
			yyVAL.orderBy = nil
		}
	case 243:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1259
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1265
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 245:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1269
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 246:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1275
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 247:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1280
		{
			yyVAL.str = AST_ASC
		}
	case 248:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1284
		{
			yyVAL.str = AST_ASC
		}
	case 249:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1288
		{
			yyVAL.str = AST_DESC
		}
	case 250:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1293
		{
			yyVAL.limit = nil
		}
	case 251:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1297
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 252:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1301
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 253:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1305
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].valExpr, Rowcount: yyDollar[2].valExpr}
		}
	case 254:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1310
		{
			yyVAL.str = ""
		}
	case 255:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1314
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 256:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1318
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
	case 257:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1331
		{
			yyVAL.columns = nil
		}
	case 258:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1335
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1341
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 260:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1345
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 261:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1350
		{
			yyVAL.updateExprs = nil
		}
	case 262:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1354
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 263:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1360
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 264:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1364
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 265:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1370
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 266:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1375
		{
			yyVAL.empty = struct{}{}
		}
	case 267:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1377
		{
			yyVAL.empty = struct{}{}
		}
	case 268:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1380
		{
			yyVAL.empty = struct{}{}
		}
	case 269:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1382
		{
			yyVAL.empty = struct{}{}
		}
	case 270:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1385
		{
			yyVAL.empty = struct{}{}
		}
	case 271:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1387
		{
			yyVAL.empty = struct{}{}
		}
	case 272:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1391
		{
			yyVAL.empty = struct{}{}
		}
	case 273:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1393
		{
			yyVAL.empty = struct{}{}
		}
	case 274:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1395
		{
			yyVAL.empty = struct{}{}
		}
	case 275:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1397
		{
			yyVAL.empty = struct{}{}
		}
	case 276:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1399
		{
			yyVAL.empty = struct{}{}
		}
	case 277:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1402
		{
			yyVAL.empty = struct{}{}
		}
	case 278:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1404
		{
			yyVAL.empty = struct{}{}
		}
	case 279:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1407
		{
			yyVAL.empty = struct{}{}
		}
	case 280:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1409
		{
			yyVAL.empty = struct{}{}
		}
	case 281:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1412
		{
			yyVAL.empty = struct{}{}
		}
	case 282:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1414
		{
			yyVAL.empty = struct{}{}
		}
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1418
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 284:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1423
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
