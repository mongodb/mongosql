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
const REGEXP = 57393
const DATE = 57394
const DATETIME = 57395
const TIME = 57396
const TIMESTAMP = 57397
const TIMESTAMPADD = 57398
const TIMESTAMPDIFF = 57399
const YEAR = 57400
const QUARTER = 57401
const MONTH = 57402
const WEEK = 57403
const DAY = 57404
const HOUR = 57405
const MINUTE = 57406
const SECOND = 57407
const MICROSECOND = 57408
const SQL_TSI_YEAR = 57409
const SQL_TSI_QUARTER = 57410
const SQL_TSI_MONTH = 57411
const SQL_TSI_WEEK = 57412
const SQL_TSI_DAY = 57413
const SQL_TSI_HOUR = 57414
const SQL_TSI_MINUTE = 57415
const SQL_TSI_SECOND = 57416
const CONVERT = 57417
const CHAR = 57418
const SIGNED = 57419
const UNSIGNED = 57420
const SQL_BIGINT = 57421
const SQL_VARCHAR = 57422
const SQL_DATE = 57423
const SQL_TIMESTAMP = 57424
const SQL_DOUBLE = 57425
const INTEGER = 57426
const UNION = 57427
const MINUS = 57428
const EXCEPT = 57429
const INTERSECT = 57430
const JOIN = 57431
const STRAIGHT_JOIN = 57432
const LEFT = 57433
const RIGHT = 57434
const INNER = 57435
const OUTER = 57436
const CROSS = 57437
const NATURAL = 57438
const USE = 57439
const FORCE = 57440
const ON = 57441
const AND = 57442
const OR = 57443
const XOR = 57444
const NOT = 57445
const MOD = 57446
const DIV = 57447
const UNARY = 57448
const CASE = 57449
const WHEN = 57450
const THEN = 57451
const ELSE = 57452
const END = 57453
const BEGIN = 57454
const COMMIT = 57455
const ROLLBACK = 57456
const NAMES = 57457
const REPLACE = 57458
const ADMIN = 57459
const SHOW = 57460
const DATABASES = 57461
const TABLES = 57462
const PROXY = 57463
const VARIABLES = 57464
const FULL = 57465
const SESSION = 57466
const GLOBAL = 57467
const COLUMNS = 57468
const CREATE = 57469
const ALTER = 57470
const DROP = 57471
const RENAME = 57472
const TABLE = 57473
const INDEX = 57474
const VIEW = 57475
const TO = 57476
const IGNORE = 57477
const IF = 57478
const UNIQUE = 57479
const USING = 57480

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
	"REGEXP",
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
const yyInitialStackSize = 16

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 31,
	141, 37,
	-2, 39,
	-1, 226,
	94, 156,
	154, 156,
	-2, 75,
	-1, 490,
	106, 74,
	107, 74,
	108, 74,
	-2, 84,
}

const yyNprod = 290
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1209

var yyAct = [...]int{

	116, 404, 108, 499, 75, 107, 345, 457, 113, 223,
	167, 102, 132, 396, 276, 338, 336, 141, 114, 131,
	246, 224, 3, 241, 323, 77, 105, 34, 35, 36,
	37, 103, 474, 64, 361, 362, 363, 364, 365, 508,
	366, 367, 93, 508, 79, 508, 315, 84, 189, 189,
	86, 254, 78, 189, 90, 66, 189, 89, 82, 189,
	99, 352, 402, 189, 189, 274, 274, 44, 49, 46,
	50, 170, 468, 47, 52, 53, 54, 467, 466, 83,
	421, 423, 85, 426, 166, 279, 51, 100, 165, 96,
	447, 317, 174, 425, 372, 278, 337, 169, 177, 510,
	449, 181, 260, 509, 187, 507, 194, 80, 473, 472,
	337, 186, 393, 471, 226, 196, 470, 222, 227, 431,
	163, 176, 401, 385, 383, 316, 273, 422, 258, 158,
	160, 261, 65, 463, 397, 233, 348, 221, 225, 179,
	180, 397, 143, 144, 145, 240, 18, 56, 58, 59,
	173, 63, 61, 62, 301, 191, 192, 193, 79, 97,
	188, 79, 244, 415, 250, 249, 78, 465, 416, 78,
	209, 210, 212, 213, 211, 251, 428, 464, 282, 419,
	418, 187, 427, 279, 281, 264, 248, 271, 272, 18,
	19, 20, 21, 278, 417, 76, 288, 250, 160, 265,
	290, 280, 274, 299, 300, 289, 304, 305, 306, 307,
	308, 309, 310, 311, 312, 313, 314, 294, 285, 286,
	287, 283, 303, 413, 22, 267, 483, 452, 414, 390,
	389, 302, 257, 259, 256, 485, 486, 247, 195, 247,
	319, 321, 388, 79, 79, 189, 387, 341, 322, 332,
	386, 78, 343, 476, 65, 349, 333, 475, 162, 182,
	493, 492, 491, 340, 334, 344, 350, 79, 266, 88,
	72, 354, 355, 356, 347, 78, 282, 357, 242, 155,
	243, 353, 281, 34, 35, 36, 37, 340, 178, 234,
	243, 280, 232, 371, 231, 230, 358, 229, 228, 377,
	378, 379, 101, 238, 373, 374, 375, 237, 191, 192,
	193, 27, 28, 29, 376, 30, 32, 31, 191, 192,
	193, 359, 382, 160, 91, 370, 23, 24, 26, 25,
	236, 235, 384, 207, 208, 209, 210, 212, 213, 211,
	395, 369, 156, 394, 295, 159, 296, 297, 65, 80,
	424, 403, 392, 505, 408, 400, 407, 263, 399, 198,
	202, 200, 201, 175, 225, 262, 245, 298, 73, 171,
	168, 164, 280, 280, 411, 412, 161, 506, 217, 218,
	219, 220, 203, 430, 214, 215, 216, 361, 362, 363,
	364, 365, 87, 366, 367, 157, 448, 479, 432, 433,
	434, 435, 451, 79, 18, 454, 92, 71, 455, 512,
	94, 453, 269, 252, 172, 429, 184, 284, 459, 204,
	205, 206, 207, 208, 209, 210, 212, 213, 211, 69,
	95, 270, 469, 458, 339, 185, 67, 405, 462, 406,
	199, 204, 205, 206, 207, 208, 209, 210, 212, 213,
	211, 346, 477, 478, 461, 450, 143, 144, 145, 410,
	247, 98, 74, 511, 494, 187, 18, 487, 381, 490,
	380, 480, 489, 39, 204, 205, 206, 207, 208, 209,
	210, 212, 213, 211, 495, 496, 57, 60, 488, 498,
	225, 497, 500, 500, 500, 79, 501, 502, 268, 503,
	38, 183, 17, 78, 143, 144, 145, 16, 320, 513,
	458, 122, 15, 514, 14, 515, 130, 13, 12, 137,
	40, 41, 42, 43, 253, 45, 106, 123, 124, 125,
	351, 55, 438, 439, 255, 48, 111, 81, 342, 504,
	135, 126, 129, 127, 128, 118, 119, 146, 147, 148,
	149, 150, 151, 152, 153, 154, 437, 440, 441, 442,
	443, 444, 445, 446, 120, 204, 205, 206, 207, 208,
	209, 210, 212, 213, 211, 484, 456, 460, 409, 391,
	239, 139, 138, 335, 121, 115, 436, 117, 398, 112,
	197, 109, 482, 110, 420, 277, 360, 133, 134, 104,
	275, 368, 140, 190, 68, 33, 142, 143, 144, 145,
	70, 11, 10, 9, 122, 8, 7, 6, 5, 130,
	4, 2, 137, 1, 0, 0, 0, 0, 0, 106,
	123, 124, 125, 0, 0, 136, 0, 0, 318, 111,
	0, 0, 0, 135, 126, 129, 127, 128, 118, 119,
	146, 147, 148, 149, 150, 151, 152, 153, 154, 0,
	0, 0, 0, 0, 0, 0, 0, 120, 0, 0,
	481, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 139, 138, 204, 205, 206, 207,
	208, 209, 210, 212, 213, 211, 110, 0, 0, 0,
	133, 134, 104, 0, 0, 140, 0, 0, 0, 142,
	293, 292, 143, 144, 145, 291, 0, 0, 0, 0,
	0, 0, 0, 0, 130, 0, 0, 137, 0, 0,
	0, 0, 0, 0, 80, 123, 124, 125, 136, 0,
	0, 0, 0, 0, 178, 0, 0, 0, 135, 126,
	129, 127, 128, 118, 119, 146, 147, 148, 149, 150,
	151, 152, 153, 154, 18, 0, 0, 0, 0, 0,
	0, 0, 120, 0, 0, 0, 0, 0, 0, 143,
	144, 145, 0, 0, 0, 0, 122, 0, 0, 139,
	138, 130, 0, 0, 137, 0, 0, 0, 0, 0,
	0, 80, 123, 124, 125, 133, 134, 0, 0, 0,
	140, 111, 0, 0, 142, 135, 126, 129, 127, 128,
	118, 119, 146, 147, 148, 149, 150, 151, 152, 153,
	154, 0, 0, 0, 0, 0, 0, 0, 0, 120,
	0, 0, 0, 136, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 139, 138, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 110, 0,
	0, 0, 133, 134, 0, 0, 0, 140, 0, 0,
	0, 142, 143, 144, 145, 0, 0, 0, 0, 122,
	0, 0, 0, 0, 130, 0, 0, 137, 0, 0,
	0, 0, 0, 0, 80, 123, 124, 125, 0, 0,
	136, 0, 0, 0, 111, 0, 0, 0, 135, 126,
	129, 127, 128, 118, 119, 146, 147, 148, 149, 150,
	151, 152, 153, 154, 0, 0, 0, 0, 0, 0,
	0, 0, 120, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 139,
	138, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	18, 110, 0, 0, 0, 133, 134, 0, 0, 0,
	140, 0, 0, 0, 142, 143, 144, 145, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 130, 0, 0,
	137, 0, 0, 0, 0, 0, 0, 80, 123, 124,
	125, 0, 0, 136, 0, 0, 0, 178, 0, 0,
	0, 135, 126, 129, 127, 128, 118, 119, 146, 147,
	148, 149, 150, 151, 152, 153, 154, 0, 0, 0,
	0, 0, 0, 0, 0, 120, 0, 0, 0, 0,
	0, 0, 143, 144, 145, 0, 0, 0, 0, 0,
	0, 0, 139, 138, 130, 0, 0, 137, 0, 0,
	0, 0, 0, 0, 80, 123, 124, 125, 133, 134,
	0, 0, 0, 140, 178, 0, 0, 142, 135, 126,
	129, 127, 128, 118, 119, 146, 147, 148, 149, 150,
	151, 152, 153, 154, 0, 0, 0, 0, 0, 0,
	0, 0, 120, 0, 0, 0, 136, 198, 202, 200,
	201, 0, 0, 0, 0, 0, 0, 0, 0, 139,
	138, 0, 0, 0, 0, 0, 217, 218, 219, 220,
	203, 0, 214, 215, 216, 133, 134, 0, 0, 0,
	140, 0, 0, 0, 142, 146, 147, 148, 149, 150,
	151, 152, 153, 154, 324, 325, 326, 327, 328, 329,
	330, 331, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 136, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 199, 204,
	205, 206, 207, 208, 209, 210, 212, 213, 211,
}
var yyPact = [...]int{

	184, -1000, -1000, 193, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -79, -80, -60, -72, -1000, -1000, -1000,
	-1000, 13, 306, 461, 413, -1000, -1000, -1000, 405, -1000,
	371, 326, 453, 65, -93, -68, 306, -1000, -64, 306,
	-1000, 350, -94, 306, -94, 370, 400, -48, 452, 306,
	-54, -1000, -1000, -1000, 250, -1000, -1000, -1000, 587, -1000,
	233, 326, 355, 9, 326, 104, 334, -1000, 205, -1000,
	0, 329, -21, 306, -1000, 328, -1000, -78, 327, 387,
	45, 306, 326, -1000, 862, 1032, 400, 400, 1032, 452,
	407, 1032, 151, -1000, -1000, 212, -5, -1000, 1089, -1000,
	862, 759, -1000, -1000, -1000, 1032, 246, 245, 243, 242,
	240, -1000, 237, -1000, -1000, -1000, 288, 287, 264, 260,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 1032, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 238, 307, 324, 450,
	307, -1000, 1032, 306, -1000, 386, -102, -1000, 89, -1000,
	323, -1000, -1000, 315, -1000, 228, 49, 455, 965, -1000,
	-1000, 455, 400, 403, 1032, 1032, -28, 455, 43, 587,
	392, 862, 862, 862, -1000, 306, 90, 692, 236, 316,
	1032, 1032, 122, 1032, 1032, 1032, 1032, 1032, 1032, 1032,
	1032, 1032, 1032, 1032, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -108, -29, -63, 49, 1089, -1000, 484, 587,
	1092, 1092, 587, -1000, 461, -1000, -1000, -1000, -1000, -27,
	455, 399, 307, 307, 229, -1000, 438, 862, -1000, 455,
	-1000, -1000, -1000, 31, 306, -1000, -88, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, 399, 307, -1000, -1000, 1032,
	1032, 455, 455, -1000, 1032, 227, 292, 299, 141, -26,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	455, 237, 237, 237, -1000, 236, 1032, 1032, 1032, 455,
	364, -1000, 436, -1000, 455, 220, 220, 220, 55, 55,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -30,
	587, -31, 156, 152, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 148, 136, 135, -13, -1000, 862, 29, 236,
	193, 36, -32, -1000, 438, 422, 425, 49, 314, -1000,
	-1000, 312, -1000, -1000, 104, 455, 455, 455, 448, 43,
	43, -1000, -1000, 128, 68, 99, 85, 84, -23, -1000,
	308, -61, 41, -1000, -1000, -1000, -1000, 455, 309, 455,
	1032, -1000, -1000, -1000, -35, -1000, 587, 587, 587, 587,
	475, -36, -1000, 1032, -24, 331, -1000, 365, 133, -1000,
	-1000, -1000, 307, 422, -1000, 1032, 862, -1000, -1000, 442,
	424, 292, 28, -1000, 82, -1000, 72, -1000, -1000, -1000,
	-1000, -69, -70, -75, -1000, -1000, -1000, -1000, -1000, 1032,
	455, -1000, -38, -41, -45, -46, -122, -1000, -1000, -1000,
	168, 164, -1000, -1000, -1000, -1000, -1000, -1000, 455, 1032,
	1032, 359, 236, -1000, -1000, 576, 132, -1000, 202, -1000,
	438, 862, 1032, 862, -1000, -1000, 210, 209, 208, 455,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 455, 455, 457,
	-1000, 1032, 1032, 862, -1000, -1000, -1000, 422, 49, 108,
	-1000, 306, 306, 306, 307, 455, 455, -1000, 336, -49,
	-1000, -51, -55, 104, -1000, 456, 381, -1000, 306, -1000,
	-1000, -1000, 306, -1000, 306, -1000,
}
var yyPgo = [...]int{

	0, 623, 621, 21, 620, 618, 617, 616, 615, 613,
	612, 611, 500, 610, 605, 604, 11, 31, 603, 601,
	26, 600, 14, 596, 595, 270, 594, 3, 20, 5,
	591, 590, 15, 589, 2, 18, 9, 19, 588, 587,
	17, 24, 586, 12, 585, 8, 584, 583, 16, 580,
	579, 578, 577, 6, 576, 7, 575, 1, 539, 23,
	538, 13, 4, 25, 269, 537, 535, 534, 530, 525,
	524, 0, 10, 518, 517, 514, 512, 507, 502, 159,
	42, 501, 498, 487, 486, 473,
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
	30, 30, 31, 31, 31, 31, 31, 31, 31, 32,
	32, 38, 38, 35, 35, 43, 36, 36, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 41, 41, 41, 41, 41, 41, 41, 41,
	40, 40, 40, 40, 40, 40, 40, 40, 40, 42,
	42, 42, 42, 42, 42, 42, 42, 42, 42, 42,
	42, 39, 39, 39, 39, 39, 39, 44, 44, 44,
	46, 49, 49, 47, 47, 48, 48, 50, 50, 45,
	45, 33, 33, 33, 33, 33, 33, 33, 33, 33,
	37, 37, 37, 51, 51, 52, 52, 53, 53, 54,
	54, 55, 56, 56, 56, 57, 57, 57, 57, 58,
	58, 58, 59, 59, 60, 60, 61, 61, 62, 62,
	63, 64, 64, 65, 65, 66, 66, 67, 67, 67,
	67, 67, 68, 68, 69, 69, 70, 70, 71, 72,
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
	4, 3, 4, 5, 6, 3, 4, 3, 4, 2,
	3, 4, 1, 1, 1, 1, 1, 1, 1, 2,
	1, 1, 3, 3, 1, 3, 1, 3, 1, 1,
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 2, 3, 4, 5, 4, 6, 6, 6, 6,
	6, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 2, 1, 2, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	5, 0, 1, 1, 2, 4, 4, 0, 2, 1,
	3, 1, 1, 1, 2, 2, 2, 2, 1, 1,
	1, 1, 1, 0, 3, 0, 2, 0, 3, 1,
	3, 2, 0, 1, 1, 0, 2, 4, 4, 0,
	2, 4, 0, 3, 1, 3, 0, 5, 1, 3,
	3, 0, 2, 0, 3, 0, 1, 1, 1, 1,
	1, 1, 0, 1, 0, 1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -73, -74, -75, -76, -77, -78, 5, 6,
	7, 8, 40, 142, 143, 145, 144, 127, 128, 129,
	131, 133, 132, -14, 90, 91, 92, 93, -12, -85,
	-12, -12, -12, -12, 146, -69, 148, 152, -66, 148,
	150, 146, 146, 147, 148, -12, 134, -84, 135, 136,
	-83, 139, 140, 138, -71, 42, -3, 23, -15, 24,
	-13, 36, -25, 42, 9, -62, 130, -63, -45, -71,
	42, -65, 151, 147, -71, 146, -71, 42, -64, 151,
	-71, -64, 36, -80, 10, 30, 137, -79, 9, -71,
	141, 52, -16, -17, 115, -20, 42, -29, -34, -30,
	109, 52, -33, -45, -35, -44, -71, -39, 61, 62,
	80, -46, 27, 43, 44, 45, 57, 59, 60, 58,
	32, -37, -43, 113, 114, 56, 151, 35, 98, 97,
	118, -40, 122, 20, 21, 22, 63, 64, 65, 66,
	67, 68, 69, 70, 71, 46, -25, 40, 120, -25,
	94, 42, 53, 120, 42, 109, -71, -72, 42, -72,
	149, 42, 27, 105, -71, -25, -20, -34, 52, -80,
	-80, -34, -79, -81, 9, 28, -36, -34, 9, 94,
	-18, 106, 107, 108, -71, 26, 120, -31, 28, 109,
	30, 31, 29, 51, 110, 111, 112, 113, 114, 115,
	116, 119, 117, 118, 53, 54, 55, 47, 48, 49,
	50, -20, -29, -36, -3, -20, -34, -34, 52, 52,
	52, 52, 52, -43, 52, 43, 43, 43, 43, -49,
	-34, -59, 40, 52, -62, 42, -28, 10, -63, -34,
	-71, -72, 27, -70, 153, -67, 145, 143, 39, 144,
	13, 42, 42, 42, -72, -59, 40, -80, -82, 9,
	28, -34, -34, 154, 94, -21, -22, -24, 52, 42,
	-43, 141, 135, -17, 25, -20, -20, -20, -71, 115,
	-34, 23, 19, 18, -35, 28, 30, 31, 51, -34,
	-34, 32, 109, -37, -34, -34, -34, -34, -34, -34,
	-34, -34, -34, -34, -34, 154, 154, 154, 154, -16,
	24, -16, -40, -41, 72, 73, 74, 75, 76, 77,
	78, 79, -40, -41, -17, -47, -48, 123, -32, 35,
	-3, -62, -60, -45, -28, -53, 13, -20, 105, -71,
	-72, -68, 149, -32, -62, -34, -34, -34, -28, 94,
	-23, 95, 96, 97, 98, 99, 101, 102, -19, 42,
	26, -22, 120, -43, -43, -43, -35, -34, -34, -34,
	106, 32, -37, 154, -16, 154, 94, 94, 94, 94,
	94, -50, -48, 125, -29, -34, -61, 105, -38, -35,
	-61, 154, 94, -53, -57, 15, 14, 42, 42, -51,
	11, -22, -22, 95, 100, 95, 100, 95, 95, 95,
	-26, 103, 150, 104, 42, 154, 42, 141, 135, 106,
	-34, 154, -16, -16, -16, -16, -42, 81, 57, 58,
	82, 83, 84, 85, 86, 87, 88, 126, -34, 124,
	124, 37, 94, -45, -57, -34, -54, -55, -20, -72,
	-52, 12, 14, 105, 95, 95, 147, 147, 147, -34,
	154, 154, 154, 154, 154, 89, 89, -34, -34, 38,
	-35, 94, 16, 94, -56, 33, 34, -53, -20, -36,
	-29, 52, 52, 52, 7, -34, -34, -55, -57, -27,
	-71, -27, -27, -62, -58, 17, 41, 154, 94, 154,
	154, 7, 28, -71, -71, -71,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 57, 57,
	57, 57, 57, 284, 275, 0, 0, 28, 29, 30,
	57, -2, 0, 0, 61, 63, 64, 65, 66, 59,
	0, 0, 0, 0, 273, 0, 0, 285, 0, 0,
	276, 0, 271, 0, 271, 0, 114, 0, 117, 0,
	0, 40, 41, 38, 0, 288, 19, 62, 0, 67,
	58, 0, 0, 104, 0, 26, 0, 268, 0, 229,
	288, 0, 0, 0, 289, 0, 289, 0, 0, 0,
	0, 0, 0, 42, 0, 0, 114, 114, 0, 117,
	0, 0, 17, 68, 70, 76, 288, 74, 75, 119,
	0, 0, 158, 159, 160, 0, 229, 0, 0, 0,
	0, 181, 0, 231, 232, 233, 0, 0, 0, 0,
	238, 239, 154, 217, 218, 219, 211, 212, 213, 214,
	215, 216, 221, 240, 241, 242, 190, 191, 192, 193,
	194, 195, 196, 197, 198, 60, 262, 0, 0, 112,
	0, 27, 0, 0, 289, 0, 286, 49, 0, 52,
	0, 54, 272, 0, 289, 262, 115, 116, 0, 43,
	44, 118, 114, 34, 0, 0, 0, 156, 0, 0,
	71, 0, 0, 0, 77, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 142, 143, 144, 145, 146, 147,
	148, 123, 74, 0, 0, 0, -2, 171, 0, 0,
	0, 0, 0, 139, 0, 234, 235, 236, 237, 0,
	222, 0, 0, 0, 112, 105, 247, 0, 269, 270,
	230, 47, 274, 0, 0, 289, 282, 277, 278, 279,
	280, 281, 53, 55, 56, 0, 0, 45, 46, 0,
	0, 32, 33, 31, 0, 112, 79, 85, 0, 97,
	99, 100, 101, 69, 72, 120, 121, 122, 78, 73,
	125, 0, 0, 0, 126, 0, 0, 0, 0, 131,
	0, 135, 0, 140, 137, 161, 162, 163, 164, 165,
	166, 167, 168, 169, 170, 124, 153, 155, 172, 0,
	0, 0, 0, 0, 182, 183, 184, 185, 186, 187,
	188, 189, 0, 0, 0, 227, 223, 0, 266, 0,
	150, 266, 0, 264, 247, 255, 0, 113, 0, 287,
	50, 0, 283, 22, 23, 35, 36, 157, 243, 0,
	0, 88, 89, 0, 0, 0, 0, 0, 106, 86,
	0, 0, 0, 127, 128, 129, 130, 132, 0, 138,
	0, 136, 141, 173, 0, 175, 0, 0, 0, 0,
	0, 0, 224, 0, 74, 75, 20, 0, 149, 151,
	21, 263, 0, 255, 25, 0, 0, 289, 51, 245,
	0, 80, 83, 90, 0, 92, 0, 94, 95, 96,
	81, 0, 0, 0, 87, 82, 98, 102, 103, 0,
	133, 174, 0, 0, 0, 0, 0, 199, 200, 201,
	202, 204, 206, 207, 208, 209, 210, 220, 228, 0,
	0, 0, 0, 265, 24, 256, 248, 249, 252, 48,
	247, 0, 0, 0, 91, 93, 0, 0, 0, 134,
	176, 177, 178, 179, 180, 203, 205, 225, 226, 0,
	152, 0, 0, 0, 251, 253, 254, 255, 246, 244,
	-2, 0, 0, 0, 0, 257, 258, 250, 259, 0,
	110, 0, 0, 267, 18, 0, 0, 107, 0, 108,
	109, 260, 0, 111, 0, 261,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 117, 110, 3,
	52, 154, 115, 113, 94, 114, 120, 116, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	54, 53, 55, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 112, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 111, 3, 56,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
	57, 58, 59, 60, 61, 62, 63, 64, 65, 66,
	67, 68, 69, 70, 71, 72, 73, 74, 75, 76,
	77, 78, 79, 80, 81, 82, 83, 84, 85, 86,
	87, 88, 89, 90, 91, 92, 93, 95, 96, 97,
	98, 99, 100, 101, 102, 103, 104, 105, 106, 107,
	108, 109, 118, 119, 121, 122, 123, 124, 125, 126,
	127, 128, 129, 130, 131, 132, 133, 134, 135, 136,
	137, 138, 139, 140, 141, 142, 143, 144, 145, 146,
	147, 148, 149, 150, 151, 152, 153,
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
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:785
		{
			yyVAL.boolExpr = &RegexExpr{Operand: yyDollar[1].valExpr, Pattern: yyDollar[3].valExpr}
		}
	case 138:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:789
		{
			yyVAL.boolExpr = &NotExpr{&RegexExpr{Operand: yyDollar[1].valExpr, Pattern: yyDollar[4].valExpr}}
		}
	case 139:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:793
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 140:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:797
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS, Right: yyDollar[3].valExpr}
		}
	case 141:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:801
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IS_NOT, Right: yyDollar[4].valExpr}
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:807
		{
			yyVAL.str = AST_EQ
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:811
		{
			yyVAL.str = AST_LT
		}
	case 144:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:815
		{
			yyVAL.str = AST_GT
		}
	case 145:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:819
		{
			yyVAL.str = AST_LE
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:823
		{
			yyVAL.str = AST_GE
		}
	case 147:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:827
		{
			yyVAL.str = AST_NE
		}
	case 148:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:831
		{
			yyVAL.str = AST_NSE
		}
	case 149:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:837
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 150:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:841
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 151:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:847
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 152:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:851
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:857
		{
			yyVAL.tuple = ValTuple(yyDollar[2].valExprs)
		}
	case 154:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:861
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:867
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 156:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:873
		{
			yyVAL.valExprs = ValExprs{yyDollar[1].valExpr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:877
		{
			yyVAL.valExprs = append(yyDollar[1].valExprs, yyDollar[3].valExpr)
		}
	case 158:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:883
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 159:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:887
		{
			yyVAL.valExpr = yyDollar[1].colName
		}
	case 160:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:891
		{
			yyVAL.valExpr = yyDollar[1].tuple
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:895
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITAND, Right: yyDollar[3].valExpr}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:899
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITOR, Right: yyDollar[3].valExpr}
		}
	case 163:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:903
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITXOR, Right: yyDollar[3].valExpr}
		}
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:907
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_PLUS, Right: yyDollar[3].valExpr}
		}
	case 165:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:911
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MINUS, Right: yyDollar[3].valExpr}
		}
	case 166:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:915
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MULT, Right: yyDollar[3].valExpr}
		}
	case 167:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:919
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_DIV, Right: yyDollar[3].valExpr}
		}
	case 168:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:923
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_IDIV, Right: yyDollar[3].valExpr}
		}
	case 169:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:927
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 170:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:931
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 171:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:935
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
	case 172:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:950
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 173:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:954
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 174:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:958
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 175:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:962
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 176:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:966
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 177:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:970
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 178:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:974
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 179:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:978
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 180:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:982
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 181:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:986
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 182:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:992
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 183:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:996
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 184:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1000
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 185:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1004
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 186:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1008
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 187:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1012
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1016
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1020
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1026
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1030
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1034
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1038
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1042
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 195:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1046
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 196:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1050
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1054
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 198:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1058
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 199:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1064
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1068
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1072
		{
			yyVAL.bytes = DATETIME_BYTES
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
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1088
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1092
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1096
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1100
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 209:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1104
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1108
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1114
		{
			yyVAL.bytes = IF_BYTES
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1118
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1122
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1126
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1130
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1134
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1140
		{
			yyVAL.byt = AST_UPLUS
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1144
		{
			yyVAL.byt = AST_UMINUS
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1148
		{
			yyVAL.byt = AST_TILDA
		}
	case 220:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1154
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 221:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1159
		{
			yyVAL.valExpr = nil
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1163
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1169
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 224:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1173
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 225:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1179
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 226:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1183
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 227:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1188
		{
			yyVAL.valExpr = nil
		}
	case 228:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1192
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1198
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 230:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1202
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 231:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1208
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1212
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1216
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 234:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1220
		{
			yyVAL.valExpr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 235:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1224
		{
			yyVAL.valExpr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 236:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1228
		{
			yyVAL.valExpr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 237:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1232
		{
			yyVAL.valExpr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1236
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1240
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1246
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1250
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1254
		{
			yyVAL.valExpr = &UnknownVal{}
		}
	case 243:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1259
		{
			yyVAL.valExprs = nil
		}
	case 244:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1263
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 245:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1268
		{
			yyVAL.expr = nil
		}
	case 246:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1272
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 247:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1277
		{
			yyVAL.orderBy = nil
		}
	case 248:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1281
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 249:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1287
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 250:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1291
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 251:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1297
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 252:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1302
		{
			yyVAL.str = AST_ASC
		}
	case 253:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1306
		{
			yyVAL.str = AST_ASC
		}
	case 254:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1310
		{
			yyVAL.str = AST_DESC
		}
	case 255:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1315
		{
			yyVAL.limit = nil
		}
	case 256:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1319
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 257:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1323
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 258:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1327
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].valExpr, Rowcount: yyDollar[2].valExpr}
		}
	case 259:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1332
		{
			yyVAL.str = ""
		}
	case 260:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1336
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 261:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1340
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
	case 262:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1353
		{
			yyVAL.columns = nil
		}
	case 263:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1357
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 264:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1363
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 265:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1367
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 266:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1372
		{
			yyVAL.updateExprs = nil
		}
	case 267:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1376
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 268:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1382
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 269:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1386
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 270:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1392
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 271:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1397
		{
			yyVAL.empty = struct{}{}
		}
	case 272:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1399
		{
			yyVAL.empty = struct{}{}
		}
	case 273:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1402
		{
			yyVAL.empty = struct{}{}
		}
	case 274:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1404
		{
			yyVAL.empty = struct{}{}
		}
	case 275:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1407
		{
			yyVAL.empty = struct{}{}
		}
	case 276:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1409
		{
			yyVAL.empty = struct{}{}
		}
	case 277:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1413
		{
			yyVAL.empty = struct{}{}
		}
	case 278:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1415
		{
			yyVAL.empty = struct{}{}
		}
	case 279:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1417
		{
			yyVAL.empty = struct{}{}
		}
	case 280:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1419
		{
			yyVAL.empty = struct{}{}
		}
	case 281:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1421
		{
			yyVAL.empty = struct{}{}
		}
	case 282:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1424
		{
			yyVAL.empty = struct{}{}
		}
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1426
		{
			yyVAL.empty = struct{}{}
		}
	case 284:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1429
		{
			yyVAL.empty = struct{}{}
		}
	case 285:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1431
		{
			yyVAL.empty = struct{}{}
		}
	case 286:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1434
		{
			yyVAL.empty = struct{}{}
		}
	case 287:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1436
		{
			yyVAL.empty = struct{}{}
		}
	case 288:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1440
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 289:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1445
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
