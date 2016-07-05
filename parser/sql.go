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
)

//line sql.y:43
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
const UNION = 57415
const MINUS = 57416
const EXCEPT = 57417
const INTERSECT = 57418
const JOIN = 57419
const STRAIGHT_JOIN = 57420
const LEFT = 57421
const RIGHT = 57422
const INNER = 57423
const OUTER = 57424
const CROSS = 57425
const NATURAL = 57426
const USE = 57427
const FORCE = 57428
const ON = 57429
const AND = 57430
const OR = 57431
const NOT = 57432
const MOD = 57433
const DIV = 57434
const UNARY = 57435
const CASE = 57436
const WHEN = 57437
const THEN = 57438
const ELSE = 57439
const END = 57440
const BEGIN = 57441
const COMMIT = 57442
const ROLLBACK = 57443
const NAMES = 57444
const REPLACE = 57445
const ADMIN = 57446
const SHOW = 57447
const DATABASES = 57448
const TABLES = 57449
const PROXY = 57450
const VARIABLES = 57451
const FULL = 57452
const COLUMNS = 57453
const CREATE = 57454
const ALTER = 57455
const DROP = 57456
const RENAME = 57457
const TABLE = 57458
const INDEX = 57459
const VIEW = 57460
const TO = 57461
const IGNORE = 57462
const IF = 57463
const UNIQUE = 57464
const USING = 57465

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
	-1, 218,
	94, 72,
	95, 72,
	-2, 148,
	-1, 459,
	94, 71,
	95, 71,
	-2, 81,
}

const yyNprod = 267
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1188

var yyAct = [...]int{

	114, 215, 106, 468, 73, 103, 429, 105, 111, 331,
	387, 100, 130, 379, 267, 322, 237, 324, 162, 112,
	477, 216, 3, 232, 139, 310, 101, 75, 34, 35,
	36, 37, 302, 62, 477, 91, 347, 348, 349, 350,
	351, 245, 352, 353, 77, 87, 80, 82, 338, 49,
	84, 50, 76, 477, 88, 64, 183, 52, 53, 54,
	97, 183, 183, 183, 183, 385, 183, 183, 265, 265,
	44, 165, 46, 440, 439, 438, 47, 479, 81, 83,
	51, 98, 161, 404, 406, 270, 409, 78, 419, 304,
	169, 478, 408, 94, 269, 358, 172, 189, 171, 175,
	18, 180, 181, 164, 187, 18, 19, 20, 21, 251,
	476, 421, 218, 445, 213, 217, 219, 214, 444, 443,
	442, 414, 384, 369, 367, 303, 264, 405, 323, 323,
	376, 174, 224, 158, 249, 153, 270, 252, 160, 22,
	155, 435, 95, 231, 290, 269, 56, 58, 59, 57,
	61, 380, 63, 77, 185, 186, 77, 235, 380, 241,
	240, 76, 334, 74, 76, 168, 273, 411, 398, 437,
	272, 410, 396, 399, 454, 455, 181, 397, 242, 238,
	436, 262, 263, 239, 402, 401, 70, 400, 255, 278,
	241, 276, 277, 280, 256, 271, 288, 289, 155, 292,
	293, 294, 295, 296, 297, 298, 299, 300, 301, 291,
	274, 284, 258, 279, 27, 28, 29, 273, 30, 32,
	31, 272, 265, 188, 248, 250, 247, 23, 24, 26,
	25, 182, 306, 308, 77, 77, 185, 186, 327, 63,
	176, 452, 76, 329, 333, 238, 335, 309, 319, 320,
	424, 345, 330, 373, 326, 372, 151, 371, 77, 154,
	370, 86, 340, 341, 342, 336, 76, 451, 343, 201,
	202, 204, 205, 203, 339, 157, 462, 170, 326, 461,
	257, 460, 271, 344, 357, 34, 35, 36, 37, 363,
	364, 234, 185, 186, 359, 360, 361, 173, 225, 223,
	222, 150, 221, 412, 183, 362, 196, 197, 198, 199,
	200, 201, 202, 204, 205, 203, 89, 155, 220, 368,
	99, 347, 348, 349, 350, 351, 378, 352, 353, 217,
	356, 377, 233, 450, 229, 448, 228, 375, 227, 226,
	386, 383, 474, 234, 63, 382, 355, 78, 196, 197,
	198, 199, 200, 201, 202, 204, 205, 203, 271, 271,
	394, 395, 407, 391, 390, 475, 365, 254, 413, 196,
	197, 198, 199, 200, 201, 202, 204, 205, 203, 420,
	253, 236, 415, 416, 417, 418, 77, 71, 166, 163,
	159, 427, 156, 85, 425, 430, 152, 426, 196, 197,
	198, 199, 200, 201, 202, 204, 205, 203, 423, 431,
	90, 18, 69, 366, 92, 441, 199, 200, 201, 202,
	204, 205, 203, 481, 446, 447, 38, 285, 243, 286,
	287, 167, 260, 93, 178, 275, 458, 181, 67, 457,
	325, 217, 456, 459, 449, 388, 40, 41, 42, 43,
	261, 65, 179, 464, 465, 434, 389, 55, 430, 466,
	332, 469, 469, 469, 77, 470, 471, 467, 472, 433,
	393, 238, 76, 96, 120, 121, 72, 307, 482, 480,
	119, 463, 483, 18, 484, 129, 39, 60, 135, 259,
	177, 17, 16, 15, 14, 104, 122, 123, 124, 13,
	12, 244, 45, 337, 109, 246, 48, 79, 133, 125,
	128, 126, 127, 116, 117, 141, 142, 143, 144, 145,
	146, 147, 148, 149, 328, 473, 453, 428, 432, 392,
	374, 230, 321, 118, 113, 115, 381, 110, 190, 137,
	136, 107, 403, 268, 346, 266, 354, 184, 66, 33,
	108, 68, 11, 10, 131, 132, 102, 9, 8, 138,
	7, 120, 121, 140, 6, 5, 4, 119, 2, 1,
	0, 0, 129, 0, 0, 135, 0, 0, 0, 0,
	0, 0, 104, 122, 123, 124, 0, 0, 0, 0,
	134, 109, 0, 305, 0, 133, 125, 128, 126, 127,
	116, 117, 141, 142, 143, 144, 145, 146, 147, 148,
	149, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 137, 136, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 108, 0, 0,
	0, 131, 132, 102, 0, 0, 138, 0, 0, 0,
	140, 0, 0, 0, 0, 283, 282, 120, 121, 281,
	0, 0, 0, 0, 0, 0, 0, 0, 129, 0,
	0, 135, 0, 0, 0, 0, 0, 134, 78, 122,
	123, 124, 0, 0, 0, 0, 0, 173, 0, 0,
	0, 133, 125, 128, 126, 127, 116, 117, 141, 142,
	143, 144, 145, 146, 147, 148, 149, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 137, 136, 0, 0, 0, 0, 0, 18,
	0, 0, 0, 0, 0, 0, 0, 131, 132, 0,
	0, 0, 138, 0, 120, 121, 140, 0, 0, 0,
	119, 0, 0, 0, 0, 129, 0, 0, 135, 0,
	0, 0, 0, 0, 0, 78, 122, 123, 124, 0,
	0, 0, 0, 134, 109, 0, 0, 0, 133, 125,
	128, 126, 127, 116, 117, 141, 142, 143, 144, 145,
	146, 147, 148, 149, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 137,
	136, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	108, 0, 0, 0, 131, 132, 0, 0, 0, 138,
	0, 120, 121, 140, 0, 0, 0, 119, 0, 0,
	0, 0, 129, 0, 0, 135, 0, 0, 0, 0,
	0, 0, 78, 122, 123, 124, 0, 0, 0, 0,
	134, 109, 0, 0, 0, 133, 125, 128, 126, 127,
	116, 117, 141, 142, 143, 144, 145, 146, 147, 148,
	149, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 137, 136, 0, 18,
	0, 0, 0, 0, 0, 0, 0, 108, 0, 0,
	0, 131, 132, 0, 120, 121, 138, 0, 0, 0,
	140, 0, 0, 0, 0, 129, 0, 0, 135, 0,
	0, 0, 0, 0, 0, 78, 122, 123, 124, 0,
	0, 0, 0, 0, 173, 0, 0, 134, 133, 125,
	128, 126, 127, 116, 117, 141, 142, 143, 144, 145,
	146, 147, 148, 149, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 137,
	136, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 131, 132, 0, 120, 121, 138,
	0, 0, 0, 140, 0, 0, 0, 0, 129, 0,
	0, 135, 0, 0, 0, 0, 0, 0, 78, 122,
	123, 124, 0, 0, 0, 0, 0, 173, 0, 0,
	134, 133, 125, 128, 126, 127, 116, 117, 141, 142,
	143, 144, 145, 146, 147, 148, 149, 191, 195, 193,
	194, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 137, 136, 0, 0, 209, 210, 211, 212,
	0, 206, 207, 208, 0, 0, 0, 131, 132, 0,
	0, 0, 138, 0, 0, 0, 140, 141, 142, 143,
	144, 145, 146, 147, 148, 149, 311, 312, 313, 314,
	315, 316, 317, 318, 0, 0, 0, 0, 191, 195,
	193, 194, 0, 134, 0, 0, 192, 196, 197, 198,
	199, 200, 201, 202, 204, 205, 203, 209, 210, 211,
	212, 422, 206, 207, 208, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 192, 196, 197,
	198, 199, 200, 201, 202, 204, 205, 203,
}
var yyPact = [...]int{

	100, -1000, -1000, 207, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -61, -84, -51, -74, -1000, -1000, -1000,
	-1000, 25, 303, 478, 429, -1000, -1000, -1000, 415, -1000,
	377, 346, 467, 46, -90, -54, 303, -1000, -52, 303,
	-1000, 352, -91, 303, -91, 375, 404, 404, 464, 303,
	-45, -1000, 270, -1000, -1000, -1000, 541, -1000, 256, 346,
	357, 28, 346, 116, 351, -1000, 224, -1000, 26, 349,
	42, 303, -1000, 348, -1000, -63, 347, 405, 72, 303,
	346, -1000, 811, 977, -1000, 404, 977, 464, 425, 977,
	222, -1000, -1000, 198, -10, -1000, 1081, -1000, 811, 724,
	-1000, -1000, -1000, 977, 268, 252, 250, 249, -1000, 248,
	-1000, -1000, -1000, -1000, -1000, 297, 296, 294, 292, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	977, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, 293, 306, 340, 461, 306, -1000, 977, 303, -1000,
	402, -97, -1000, 96, -1000, 339, -1000, -1000, 326, -1000,
	241, 60, 301, 894, -1000, 301, 404, 423, 977, 977,
	-13, 301, 44, 541, 411, 811, 811, -1000, 303, 111,
	637, 247, 400, 977, 977, 113, 977, 977, 977, 977,
	977, 977, 977, 977, 977, 977, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -107, -14, -50, 60, 1081, -1000,
	454, 541, 1026, 1026, -1000, 478, -1000, -1000, -1000, -1000,
	19, 301, 406, 306, 306, 235, -1000, 447, 811, -1000,
	301, -1000, -1000, -1000, 69, 303, -1000, -86, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 406, 306, -1000, -1000,
	977, 977, 301, 301, -1000, 977, 169, 238, 305, 95,
	-12, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	301, 248, 248, 248, -1000, 247, 977, 977, 301, 272,
	-1000, 382, 316, 316, 316, 167, 167, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -15, 541, -16, 178,
	175, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 173,
	171, 18, -1000, 811, 65, 247, 207, 58, -17, -1000,
	447, 430, 442, 60, 323, -1000, -1000, 322, -1000, -1000,
	116, 301, 301, 301, 459, 44, 44, -1000, -1000, 89,
	85, 104, 102, 101, -8, -1000, 321, -47, 45, -1000,
	-1000, -1000, -1000, 301, 209, 977, -1000, -1000, -18, -1000,
	541, 541, 541, 541, -25, -1000, 977, 0, 1020, -1000,
	372, 168, -1000, -1000, -1000, 306, 430, -1000, 977, 811,
	-1000, -1000, 457, 441, 238, 48, -1000, 97, -1000, 86,
	-1000, -1000, -1000, -1000, -57, -58, -59, -1000, -1000, -1000,
	-1000, -1000, 977, 301, -1000, -19, -20, -21, -26, -1000,
	301, 977, 977, 298, 247, -1000, -1000, 251, 159, -1000,
	142, -1000, 447, 811, 977, 811, -1000, -1000, 231, 229,
	226, 301, -1000, -1000, -1000, -1000, 301, 301, 474, -1000,
	977, 977, 811, -1000, -1000, -1000, 430, 60, 140, -1000,
	303, 303, 303, 306, 301, 301, -1000, 325, -29, -1000,
	-48, -62, 116, -1000, 472, 396, -1000, 303, -1000, -1000,
	-1000, 303, -1000, 303, -1000,
}
var yyPgo = [...]int{

	0, 569, 568, 21, 566, 565, 564, 560, 558, 557,
	553, 552, 426, 551, 549, 548, 11, 26, 547, 546,
	5, 545, 14, 544, 543, 186, 542, 3, 16, 7,
	541, 538, 17, 537, 2, 19, 1, 536, 535, 24,
	25, 12, 534, 8, 533, 532, 15, 531, 530, 529,
	528, 9, 527, 6, 526, 10, 525, 23, 524, 13,
	4, 27, 261, 507, 506, 505, 503, 502, 501, 0,
	18, 500, 499, 494, 493, 492, 491, 142, 35, 490,
	489, 487, 486,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 74, 74, 5, 6, 7, 7, 71, 72,
	73, 76, 79, 79, 80, 80, 80, 81, 81, 75,
	75, 75, 75, 75, 8, 8, 8, 9, 9, 9,
	10, 11, 11, 11, 82, 12, 13, 13, 14, 14,
	14, 14, 14, 15, 15, 16, 16, 17, 17, 17,
	17, 20, 20, 18, 18, 18, 21, 21, 22, 22,
	22, 22, 19, 19, 19, 23, 23, 23, 23, 23,
	23, 23, 23, 23, 24, 24, 24, 24, 24, 24,
	24, 25, 25, 26, 26, 26, 26, 27, 27, 28,
	28, 78, 78, 78, 77, 77, 29, 29, 29, 29,
	29, 30, 30, 30, 30, 30, 30, 30, 30, 30,
	30, 30, 30, 30, 31, 31, 31, 31, 31, 31,
	31, 32, 32, 37, 37, 35, 35, 41, 36, 36,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 40, 40, 40, 40, 40, 40, 40,
	40, 39, 39, 39, 39, 39, 39, 39, 39, 39,
	38, 38, 38, 38, 38, 38, 42, 42, 42, 44,
	47, 47, 45, 45, 46, 46, 48, 48, 43, 43,
	33, 33, 33, 33, 33, 33, 33, 33, 33, 33,
	49, 49, 50, 50, 51, 51, 52, 52, 53, 54,
	54, 54, 55, 55, 55, 55, 56, 56, 56, 57,
	57, 58, 58, 59, 59, 60, 60, 61, 62, 62,
	63, 63, 64, 64, 65, 65, 65, 65, 65, 66,
	66, 67, 67, 68, 68, 69, 70,
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
	2, 0, 2, 2, 0, 2, 1, 3, 3, 2,
	3, 3, 3, 4, 4, 4, 4, 3, 4, 5,
	6, 3, 4, 2, 1, 1, 1, 1, 1, 1,
	1, 2, 1, 1, 3, 3, 1, 3, 1, 3,
	1, 1, 1, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 2, 3, 4, 5, 4, 6, 6,
	6, 6, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 5,
	0, 1, 1, 2, 4, 4, 0, 2, 1, 3,
	1, 1, 1, 1, 1, 2, 2, 2, 2, 1,
	0, 3, 0, 2, 0, 3, 1, 3, 2, 0,
	1, 1, 0, 2, 4, 4, 0, 2, 4, 0,
	3, 1, 3, 0, 5, 1, 3, 3, 0, 2,
	0, 3, 0, 1, 1, 1, 1, 1, 1, 0,
	1, 0, 1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -71, -72, -73, -74, -75, -76, 5, 6,
	7, 8, 39, 127, 128, 130, 129, 114, 115, 116,
	118, 120, 119, -14, 78, 79, 80, 81, -12, -82,
	-12, -12, -12, -12, 131, -67, 133, 137, -64, 133,
	135, 131, 131, 132, 133, -12, 121, 124, 122, 123,
	-81, 125, -69, 41, -3, 22, -15, 23, -13, 35,
	-25, 41, 9, -60, 117, -61, -43, -69, 41, -63,
	136, 132, -69, 131, -69, 41, -62, 136, -69, -62,
	35, -78, 10, 29, -78, -77, 9, -69, 126, 50,
	-16, -17, 102, -20, 41, -29, -34, -30, 96, 50,
	-33, -43, -35, -42, -69, -38, 59, 60, -44, 26,
	20, 21, 42, 43, 44, 55, 57, 58, 56, 31,
	-41, 100, 101, 54, 136, 34, 86, 85, 105, -39,
	109, 61, 62, 63, 64, 65, 66, 67, 68, 69,
	45, -25, 39, 107, -25, 82, 41, 51, 107, 41,
	96, -69, -70, 41, -70, 134, 41, 26, 93, -69,
	-25, -20, -34, 50, -78, -34, -77, -79, 9, 27,
	-36, -34, 9, 82, -18, 94, 95, -69, 25, 107,
	-31, 27, 96, 29, 30, 28, 97, 98, 99, 100,
	101, 102, 103, 106, 104, 105, 51, 52, 53, 46,
	47, 48, 49, -20, -29, -36, -3, -20, -34, -34,
	50, 50, 50, 50, -41, 50, 42, 42, 42, 42,
	-47, -34, -57, 39, 50, -60, 41, -28, 10, -61,
	-34, -69, -70, 26, -68, 138, -65, 130, 128, 38,
	129, 13, 41, 41, 41, -70, -57, 39, -78, -80,
	9, 27, -34, -34, 139, 82, -21, -22, -24, 50,
	41, -41, 126, 122, -17, 24, -20, -20, -69, 102,
	-34, 22, 19, 18, -35, 27, 29, 30, -34, -34,
	31, 96, -34, -34, -34, -34, -34, -34, -34, -34,
	-34, -34, 139, 139, 139, 139, -16, 23, -16, -39,
	-40, 70, 71, 72, 73, 74, 75, 76, 77, -39,
	-40, -45, -46, 110, -32, 34, -3, -60, -58, -43,
	-28, -51, 13, -20, 93, -69, -70, -66, 134, -32,
	-60, -34, -34, -34, -28, 82, -23, 83, 84, 85,
	86, 87, 89, 90, -19, 41, 25, -22, 107, -41,
	-41, -41, -35, -34, -34, 94, 31, 139, -16, 139,
	82, 82, 82, 82, -48, -46, 112, -29, -34, -59,
	93, -37, -35, -59, 139, 82, -51, -55, 15, 14,
	41, 41, -49, 11, -22, -22, 83, 88, 83, 88,
	83, 83, 83, -26, 91, 135, 92, 41, 139, 41,
	126, 122, 94, -34, 139, -16, -16, -16, -16, 113,
	-34, 111, 111, 36, 82, -43, -55, -34, -52, -53,
	-20, -70, -50, 12, 14, 93, 83, 83, 132, 132,
	132, -34, 139, 139, 139, 139, -34, -34, 37, -35,
	82, 16, 82, -54, 32, 33, -51, -20, -36, -29,
	50, 50, 50, 7, -34, -34, -53, -55, -27, -69,
	-27, -27, -60, -56, 17, 40, 139, 82, 139, 139,
	7, 27, -69, -69, -69,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 261, 252, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 250, 0, 0, 262, 0, 0,
	253, 0, 248, 0, 248, 0, 111, 111, 114, 0,
	0, 38, 0, 265, 19, 59, 0, 64, 55, 0,
	0, 101, 0, 26, 0, 245, 0, 208, 265, 0,
	0, 0, 266, 0, 266, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 111, 0, 114, 0, 0,
	17, 65, 67, 73, 265, 71, 72, 116, 0, 0,
	150, 151, 152, 0, 208, 0, 0, 0, 172, 0,
	210, 211, 212, 213, 214, 0, 0, 0, 0, 219,
	146, 196, 197, 198, 190, 191, 192, 193, 194, 195,
	200, 181, 182, 183, 184, 185, 186, 187, 188, 189,
	57, 239, 0, 0, 109, 0, 27, 0, 0, 266,
	0, 263, 46, 0, 49, 0, 51, 249, 0, 266,
	239, 112, 113, 0, 41, 115, 111, 34, 0, 0,
	0, 148, 0, 0, 68, 0, 0, 74, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 134, 135, 136, 137,
	138, 139, 140, 119, 71, 0, 0, 0, -2, 163,
	0, 0, 0, 0, 133, 0, 215, 216, 217, 218,
	0, 201, 0, 0, 0, 109, 102, 224, 0, 246,
	247, 209, 44, 251, 0, 0, 266, 259, 254, 255,
	256, 257, 258, 50, 52, 53, 0, 0, 42, 43,
	0, 0, 32, 33, 31, 0, 109, 76, 82, 0,
	94, 96, 97, 98, 66, 69, 117, 118, 75, 70,
	121, 0, 0, 0, 122, 0, 0, 0, 127, 0,
	131, 0, 153, 154, 155, 156, 157, 158, 159, 160,
	161, 162, 120, 145, 147, 164, 0, 0, 0, 0,
	0, 173, 174, 175, 176, 177, 178, 179, 180, 0,
	0, 206, 202, 0, 243, 0, 142, 243, 0, 241,
	224, 232, 0, 110, 0, 264, 47, 0, 260, 22,
	23, 35, 36, 149, 220, 0, 0, 85, 86, 0,
	0, 0, 0, 0, 103, 83, 0, 0, 0, 123,
	124, 125, 126, 128, 0, 0, 132, 165, 0, 167,
	0, 0, 0, 0, 0, 203, 0, 71, 72, 20,
	0, 141, 143, 21, 240, 0, 232, 25, 0, 0,
	266, 48, 222, 0, 77, 80, 87, 0, 89, 0,
	91, 92, 93, 78, 0, 0, 0, 84, 79, 95,
	99, 100, 0, 129, 166, 0, 0, 0, 0, 199,
	207, 0, 0, 0, 0, 242, 24, 233, 225, 226,
	229, 45, 224, 0, 0, 0, 88, 90, 0, 0,
	0, 130, 168, 169, 170, 171, 204, 205, 0, 144,
	0, 0, 0, 228, 230, 231, 232, 223, 221, -2,
	0, 0, 0, 0, 234, 235, 227, 236, 0, 107,
	0, 0, 244, 18, 0, 0, 104, 0, 105, 106,
	237, 0, 108, 0, 238,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 104, 97, 3,
	50, 139, 102, 100, 82, 101, 107, 103, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	52, 51, 53, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 99, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 98, 3, 54,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 47, 48, 49, 55, 56,
	57, 58, 59, 60, 61, 62, 63, 64, 65, 66,
	67, 68, 69, 70, 71, 72, 73, 74, 75, 76,
	77, 78, 79, 80, 81, 83, 84, 85, 86, 87,
	88, 89, 90, 91, 92, 93, 94, 95, 96, 105,
	106, 108, 109, 110, 111, 112, 113, 114, 115, 116,
	117, 118, 119, 120, 121, 122, 123, 124, 125, 126,
	127, 128, 129, 130, 131, 132, 133, 134, 135, 136,
	137, 138,
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
		//line sql.y:189
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:195
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:215
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 18:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:219
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].valExprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:223
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 20:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:230
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:234
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
		//line sql.y:246
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:250
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
		//line sql.y:263
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 25:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:269
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:275
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:279
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:285
		{
			yyVAL.statement = &Begin{}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:291
		{
			yyVAL.statement = &Commit{}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:297
		{
			yyVAL.statement = &Rollback{}
		}
	case 31:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:303
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].valExprs}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:309
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:313
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 34:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:318
		{
			yyVAL.valExpr = nil
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:322
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:326
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 37:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:331
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:335
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 39:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:342
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:346
		{
			yyVAL.statement = &Show{Section: "variables", LikeOrWhere: yyDollar[3].expr}
		}
	case 41:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:350
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].valExpr, LikeOrWhere: yyDollar[4].expr}
		}
	case 42:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:354
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].valExpr, LikeOrWhere: yyDollar[5].expr}
		}
	case 43:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:358
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].valExpr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].valExpr}
		}
	case 44:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:364
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 45:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:368
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:373
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 47:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:379
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 48:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:383
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 49:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:388
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 50:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:394
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 51:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:400
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 52:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:404
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 53:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:409
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 54:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:414
		{
			SetAllowComments(yylex, true)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:418
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 56:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:424
		{
			yyVAL.bytes2 = nil
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:428
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:434
		{
			yyVAL.str = AST_UNION
		}
	case 59:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:438
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:442
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:446
		{
			yyVAL.str = AST_EXCEPT
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:450
		{
			yyVAL.str = AST_INTERSECT
		}
	case 63:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:455
		{
			yyVAL.str = ""
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:459
		{
			yyVAL.str = AST_DISTINCT
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:465
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:469
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:475
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:479
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:483
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 70:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:487
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:493
		{
			yyVAL.expr = yyDollar[1].boolExpr
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:497
		{
			yyVAL.expr = yyDollar[1].valExpr
		}
	case 73:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:502
		{
			yyVAL.bytes = nil
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:506
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:510
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 76:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:516
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 77:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:520
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:526
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:530
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:534
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 81:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:538
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].boolExpr}
		}
	case 82:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:543
		{
			yyVAL.bytes = nil
		}
	case 83:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:547
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 84:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:551
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 85:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:557
		{
			yyVAL.str = AST_JOIN
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:561
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 87:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:565
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:569
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 89:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:573
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:577
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 91:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:581
		{
			yyVAL.str = AST_JOIN
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:585
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:589
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 94:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:595
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:599
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 96:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:603
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:607
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:611
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:615
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 100:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:619
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 101:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:625
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:629
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 103:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:634
		{
			yyVAL.indexHints = nil
		}
	case 104:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:638
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 105:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:642
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 106:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:646
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 107:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:652
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 108:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:656
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 109:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:661
		{
			yyVAL.expr = nil
		}
	case 110:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:665
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 111:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:670
		{
			yyVAL.expr = nil
		}
	case 112:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:674
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:678
		{
			yyVAL.expr = yyDollar[2].valExpr
		}
	case 114:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:683
		{
			yyVAL.valExpr = nil
		}
	case 115:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:687
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 117:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:694
		{
			yyVAL.boolExpr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 118:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:698
		{
			yyVAL.boolExpr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 119:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:702
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 120:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:706
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyDollar[2].boolExpr}
		}
	case 121:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:712
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, Right: yyDollar[3].valExpr}
		}
	case 122:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:716
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 123:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:720
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ALL, Right: yyDollar[4].subquery}
		}
	case 124:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:724
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ANY, Right: yyDollar[4].subquery}
		}
	case 125:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:728
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_SOME, Right: yyDollar[4].subquery}
		}
	case 126:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:732
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 127:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:736
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_LIKE, Right: yyDollar[3].valExpr}
		}
	case 128:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:740
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_LIKE, Right: yyDollar[4].valExpr}
		}
	case 129:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:744
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_BETWEEN, From: yyDollar[3].valExpr, To: yyDollar[5].valExpr}
		}
	case 130:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:748
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].valExpr, To: yyDollar[6].valExpr}
		}
	case 131:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:752
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NULL, Expr: yyDollar[1].valExpr}
		}
	case 132:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:756
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NOT_NULL, Expr: yyDollar[1].valExpr}
		}
	case 133:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:760
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 134:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:766
		{
			yyVAL.str = AST_EQ
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:770
		{
			yyVAL.str = AST_LT
		}
	case 136:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:774
		{
			yyVAL.str = AST_GT
		}
	case 137:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:778
		{
			yyVAL.str = AST_LE
		}
	case 138:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:782
		{
			yyVAL.str = AST_GE
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:786
		{
			yyVAL.str = AST_NE
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:790
		{
			yyVAL.str = AST_NSE
		}
	case 141:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:796
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:800
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:806
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 144:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:810
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:816
		{
			yyVAL.tuple = ValTuple(yyDollar[2].valExprs)
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:820
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:826
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 148:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:832
		{
			yyVAL.valExprs = ValExprs{yyDollar[1].valExpr}
		}
	case 149:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:836
		{
			yyVAL.valExprs = append(yyDollar[1].valExprs, yyDollar[3].valExpr)
		}
	case 150:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:842
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 151:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:846
		{
			yyVAL.valExpr = yyDollar[1].colName
		}
	case 152:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:850
		{
			yyVAL.valExpr = yyDollar[1].tuple
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:854
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITAND, Right: yyDollar[3].valExpr}
		}
	case 154:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:858
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITOR, Right: yyDollar[3].valExpr}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:862
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITXOR, Right: yyDollar[3].valExpr}
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:866
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_PLUS, Right: yyDollar[3].valExpr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:870
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MINUS, Right: yyDollar[3].valExpr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:874
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MULT, Right: yyDollar[3].valExpr}
		}
	case 159:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:878
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_DIV, Right: yyDollar[3].valExpr}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:882
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_IDIV, Right: yyDollar[3].valExpr}
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:886
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:890
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 163:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:894
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
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:909
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 165:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:913
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 166:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:917
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 167:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:921
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 168:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:925
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 169:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:929
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 170:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:933
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 171:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:937
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 172:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:941
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 173:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:947
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 174:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:951
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 175:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:955
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 176:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:959
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 177:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:963
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 178:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:967
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 179:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:971
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 180:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:975
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 181:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:981
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 182:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:985
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 183:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:989
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 184:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:993
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 185:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:997
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 186:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1001
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 187:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1005
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1009
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1013
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1019
		{
			yyVAL.bytes = IF_BYTES
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1027
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1031
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1035
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 195:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1039
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 196:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1045
		{
			yyVAL.byt = AST_UPLUS
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1049
		{
			yyVAL.byt = AST_UMINUS
		}
	case 198:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1053
		{
			yyVAL.byt = AST_TILDA
		}
	case 199:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1059
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 200:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1064
		{
			yyVAL.valExpr = nil
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1068
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1074
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 203:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1078
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 204:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1084
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 205:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1088
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 206:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1093
		{
			yyVAL.valExpr = nil
		}
	case 207:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1097
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1103
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 209:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1107
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1113
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1117
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1121
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1125
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1129
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 215:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1133
		{
			yyVAL.valExpr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 216:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1137
		{
			yyVAL.valExpr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 217:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1141
		{
			yyVAL.valExpr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 218:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1145
		{
			yyVAL.valExpr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1149
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 220:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1154
		{
			yyVAL.valExprs = nil
		}
	case 221:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1158
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 222:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1163
		{
			yyVAL.expr = nil
		}
	case 223:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1167
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 224:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1172
		{
			yyVAL.orderBy = nil
		}
	case 225:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1176
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1182
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 227:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1186
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 228:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1192
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 229:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1197
		{
			yyVAL.str = AST_ASC
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1201
		{
			yyVAL.str = AST_ASC
		}
	case 231:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1205
		{
			yyVAL.str = AST_DESC
		}
	case 232:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1210
		{
			yyVAL.limit = nil
		}
	case 233:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1214
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 234:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1218
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 235:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1222
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].valExpr, Rowcount: yyDollar[2].valExpr}
		}
	case 236:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1227
		{
			yyVAL.str = ""
		}
	case 237:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1231
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 238:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1235
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
	case 239:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1248
		{
			yyVAL.columns = nil
		}
	case 240:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1252
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1258
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 242:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1262
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 243:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1267
		{
			yyVAL.updateExprs = nil
		}
	case 244:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1271
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1277
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 246:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1281
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 247:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1287
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 248:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1292
		{
			yyVAL.empty = struct{}{}
		}
	case 249:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1294
		{
			yyVAL.empty = struct{}{}
		}
	case 250:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1297
		{
			yyVAL.empty = struct{}{}
		}
	case 251:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1299
		{
			yyVAL.empty = struct{}{}
		}
	case 252:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1302
		{
			yyVAL.empty = struct{}{}
		}
	case 253:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1304
		{
			yyVAL.empty = struct{}{}
		}
	case 254:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1308
		{
			yyVAL.empty = struct{}{}
		}
	case 255:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1310
		{
			yyVAL.empty = struct{}{}
		}
	case 256:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1312
		{
			yyVAL.empty = struct{}{}
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1314
		{
			yyVAL.empty = struct{}{}
		}
	case 258:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1316
		{
			yyVAL.empty = struct{}{}
		}
	case 259:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1319
		{
			yyVAL.empty = struct{}{}
		}
	case 260:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1321
		{
			yyVAL.empty = struct{}{}
		}
	case 261:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1324
		{
			yyVAL.empty = struct{}{}
		}
	case 262:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1326
		{
			yyVAL.empty = struct{}{}
		}
	case 263:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1329
		{
			yyVAL.empty = struct{}{}
		}
	case 264:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1331
		{
			yyVAL.empty = struct{}{}
		}
	case 265:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1335
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 266:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1340
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
