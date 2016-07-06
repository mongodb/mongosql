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
const NOT = 57442
const MOD = 57443
const DIV = 57444
const UNARY = 57445
const CASE = 57446
const WHEN = 57447
const THEN = 57448
const ELSE = 57449
const END = 57450
const BEGIN = 57451
const COMMIT = 57452
const ROLLBACK = 57453
const NAMES = 57454
const REPLACE = 57455
const ADMIN = 57456
const SHOW = 57457
const DATABASES = 57458
const TABLES = 57459
const PROXY = 57460
const VARIABLES = 57461
const FULL = 57462
const COLUMNS = 57463
const CREATE = 57464
const ALTER = 57465
const DROP = 57466
const RENAME = 57467
const TABLE = 57468
const INDEX = 57469
const VIEW = 57470
const TO = 57471
const IGNORE = 57472
const IF = 57473
const UNIQUE = 57474
const USING = 57475

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
	-1, 219,
	104, 72,
	105, 72,
	-2, 148,
	-1, 477,
	104, 71,
	105, 71,
	-2, 81,
}

const yyNprod = 280
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1268

var yyAct = [...]int{

	114, 216, 106, 486, 73, 103, 444, 105, 111, 334,
	391, 100, 131, 383, 269, 325, 239, 327, 163, 112,
	495, 217, 3, 234, 312, 101, 495, 75, 140, 495,
	184, 184, 184, 62, 184, 91, 350, 351, 352, 353,
	354, 461, 355, 356, 77, 304, 247, 82, 184, 389,
	84, 184, 76, 184, 88, 64, 267, 87, 80, 341,
	97, 34, 35, 36, 37, 44, 267, 46, 49, 455,
	50, 47, 52, 53, 54, 454, 166, 497, 453, 413,
	408, 410, 162, 496, 18, 83, 494, 460, 459, 458,
	170, 457, 412, 94, 81, 51, 173, 436, 172, 176,
	253, 181, 182, 165, 188, 418, 388, 98, 372, 434,
	370, 326, 219, 305, 214, 218, 220, 215, 361, 326,
	272, 380, 306, 266, 409, 251, 272, 190, 254, 271,
	78, 175, 159, 226, 154, 271, 56, 58, 59, 57,
	61, 186, 187, 63, 233, 200, 201, 202, 203, 205,
	206, 204, 161, 156, 77, 450, 384, 77, 237, 337,
	243, 242, 76, 169, 384, 76, 472, 473, 292, 452,
	415, 402, 183, 400, 414, 451, 403, 182, 401, 244,
	240, 95, 264, 265, 241, 18, 19, 20, 21, 257,
	280, 243, 278, 279, 282, 258, 273, 290, 291, 406,
	294, 295, 296, 297, 298, 299, 300, 301, 302, 303,
	276, 275, 286, 260, 281, 274, 74, 275, 405, 22,
	404, 274, 156, 267, 470, 250, 252, 249, 202, 203,
	205, 206, 204, 308, 310, 439, 77, 77, 186, 187,
	330, 240, 70, 293, 76, 332, 336, 189, 338, 322,
	377, 323, 311, 321, 333, 184, 329, 376, 375, 469,
	77, 374, 348, 63, 343, 344, 345, 339, 76, 373,
	346, 350, 351, 352, 353, 354, 342, 355, 356, 177,
	329, 463, 462, 86, 273, 347, 360, 34, 35, 36,
	37, 366, 367, 158, 480, 479, 362, 363, 364, 478,
	259, 174, 235, 227, 27, 28, 29, 365, 30, 32,
	31, 236, 152, 236, 225, 155, 224, 23, 24, 26,
	25, 371, 223, 156, 222, 221, 186, 187, 99, 382,
	151, 359, 218, 171, 381, 468, 231, 230, 89, 229,
	379, 228, 63, 390, 387, 492, 78, 358, 386, 411,
	197, 198, 199, 200, 201, 202, 203, 205, 206, 204,
	395, 273, 273, 398, 399, 394, 256, 255, 493, 416,
	238, 417, 197, 198, 199, 200, 201, 202, 203, 205,
	206, 204, 71, 435, 167, 419, 420, 421, 422, 164,
	77, 160, 157, 85, 153, 442, 466, 438, 440, 445,
	368, 441, 18, 197, 198, 199, 200, 201, 202, 203,
	205, 206, 204, 446, 90, 69, 369, 425, 426, 456,
	197, 198, 199, 200, 201, 202, 203, 205, 206, 204,
	287, 328, 288, 289, 92, 499, 245, 262, 168, 464,
	465, 424, 427, 428, 429, 430, 431, 432, 433, 179,
	277, 476, 182, 93, 475, 263, 218, 474, 477, 467,
	67, 65, 392, 449, 393, 335, 448, 180, 397, 240,
	96, 482, 483, 72, 498, 481, 445, 484, 18, 487,
	487, 487, 77, 488, 489, 485, 490, 38, 39, 60,
	76, 261, 121, 122, 178, 309, 500, 17, 120, 16,
	501, 15, 502, 130, 14, 13, 136, 40, 41, 42,
	43, 12, 246, 104, 123, 124, 125, 45, 55, 340,
	248, 48, 109, 79, 331, 491, 134, 126, 129, 127,
	128, 116, 117, 142, 143, 144, 145, 146, 147, 148,
	149, 150, 471, 443, 447, 396, 378, 232, 324, 119,
	118, 113, 423, 115, 385, 110, 191, 107, 407, 270,
	349, 268, 357, 185, 66, 33, 68, 138, 137, 11,
	10, 9, 8, 7, 6, 5, 4, 2, 108, 1,
	0, 0, 132, 133, 102, 0, 0, 139, 0, 121,
	122, 141, 0, 0, 0, 120, 0, 0, 0, 0,
	130, 0, 0, 136, 0, 0, 0, 0, 0, 0,
	104, 123, 124, 125, 0, 0, 0, 0, 135, 109,
	0, 307, 0, 134, 126, 129, 127, 128, 116, 117,
	142, 143, 144, 145, 146, 147, 148, 149, 150, 0,
	0, 0, 0, 0, 0, 0, 0, 118, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 138, 137, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 108, 0, 0, 0, 132,
	133, 102, 0, 0, 139, 0, 0, 0, 141, 0,
	0, 0, 0, 285, 284, 121, 122, 283, 0, 0,
	0, 0, 0, 0, 0, 0, 130, 0, 0, 136,
	0, 0, 0, 0, 0, 135, 78, 123, 124, 125,
	0, 0, 0, 0, 0, 174, 0, 0, 0, 134,
	126, 129, 127, 128, 116, 117, 142, 143, 144, 145,
	146, 147, 148, 149, 150, 0, 0, 0, 0, 0,
	0, 0, 0, 118, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	138, 137, 0, 0, 0, 0, 0, 18, 0, 0,
	0, 0, 0, 0, 0, 132, 133, 0, 0, 0,
	139, 0, 121, 122, 141, 0, 0, 0, 120, 0,
	0, 0, 0, 130, 0, 0, 136, 0, 0, 0,
	0, 0, 0, 78, 123, 124, 125, 0, 0, 0,
	0, 135, 109, 0, 0, 0, 134, 126, 129, 127,
	128, 116, 117, 142, 143, 144, 145, 146, 147, 148,
	149, 150, 0, 0, 0, 0, 0, 0, 0, 0,
	118, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 138, 137, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 108, 0,
	0, 0, 132, 133, 0, 0, 0, 139, 0, 121,
	122, 141, 0, 0, 0, 120, 0, 0, 0, 0,
	130, 0, 0, 136, 0, 0, 0, 0, 0, 0,
	78, 123, 124, 125, 0, 0, 0, 0, 135, 109,
	0, 0, 0, 134, 126, 129, 127, 128, 116, 117,
	142, 143, 144, 145, 146, 147, 148, 149, 150, 0,
	0, 0, 0, 0, 0, 0, 0, 118, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 138, 137, 0, 18, 0, 0,
	0, 0, 0, 0, 0, 108, 0, 0, 0, 132,
	133, 0, 121, 122, 139, 0, 0, 0, 141, 0,
	0, 0, 0, 130, 0, 0, 136, 0, 0, 0,
	0, 0, 0, 78, 123, 124, 125, 0, 0, 0,
	0, 0, 174, 0, 0, 135, 134, 126, 129, 127,
	128, 116, 117, 142, 143, 144, 145, 146, 147, 148,
	149, 150, 0, 0, 0, 0, 0, 0, 0, 0,
	118, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 138, 137, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 132, 133, 0, 121, 122, 139, 0, 0,
	0, 141, 0, 0, 0, 0, 130, 0, 0, 136,
	0, 0, 0, 0, 0, 0, 78, 123, 124, 125,
	0, 0, 0, 0, 0, 174, 0, 0, 135, 134,
	126, 129, 127, 128, 116, 117, 142, 143, 144, 145,
	146, 147, 148, 149, 150, 0, 0, 0, 0, 0,
	0, 0, 0, 118, 192, 196, 194, 195, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	138, 137, 0, 210, 211, 212, 213, 0, 207, 208,
	209, 192, 196, 194, 195, 132, 133, 0, 0, 0,
	139, 0, 0, 0, 141, 0, 0, 0, 0, 0,
	210, 211, 212, 213, 0, 207, 208, 209, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 135, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 193, 197, 198, 199, 200, 201, 202,
	203, 205, 206, 204, 0, 0, 0, 0, 437, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	193, 197, 198, 199, 200, 201, 202, 203, 205, 206,
	204, 142, 143, 144, 145, 146, 147, 148, 149, 150,
	313, 314, 315, 316, 317, 318, 319, 320,
}
var yyPact = [...]int{

	180, -1000, -1000, 199, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -76, -75, -46, -69, -1000, -1000, -1000,
	-1000, 5, 301, 473, 439, -1000, -1000, -1000, 437, -1000,
	380, 341, 464, 89, -88, -48, 301, -1000, -56, 301,
	-1000, 352, -89, 301, -89, 379, 424, 424, 461, 301,
	-29, -1000, 278, -1000, -1000, -1000, 569, -1000, 285, 341,
	355, 17, 341, 130, 351, -1000, 242, -1000, 15, 350,
	46, 301, -1000, 348, -1000, -68, 343, 412, 60, 301,
	341, -1000, 869, 1055, -1000, 424, 1055, 461, 440, 1055,
	163, -1000, -1000, 222, 10, -1000, 1134, -1000, 869, 772,
	-1000, -1000, -1000, 1055, 275, 274, 272, 266, 264, -1000,
	253, -1000, -1000, -1000, -1000, -1000, 299, 297, 295, 294,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, 1055, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 263, 305, 329, 459, 305, -1000, 1055, 301,
	-1000, 410, -102, -1000, 87, -1000, 326, -1000, -1000, 325,
	-1000, 261, 37, 313, 962, -1000, 313, 424, 428, 1055,
	1055, -26, 313, 85, 569, 426, 869, 869, -1000, 301,
	102, 675, 251, 403, 1055, 1055, 137, 1055, 1055, 1055,
	1055, 1055, 1055, 1055, 1055, 1055, 1055, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -104, -36, -27, 37, 1134,
	-1000, 472, 569, 1190, 1190, 569, -1000, 473, -1000, -1000,
	-1000, -1000, -9, 313, 397, 305, 305, 231, -1000, 452,
	869, -1000, 313, -1000, -1000, -1000, 56, 301, -1000, -85,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 397, 305,
	-1000, -1000, 1055, 1055, 313, 313, -1000, 1055, 170, 178,
	306, 79, 1, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 313, 253, 253, 253, -1000, 251, 1055, 1055,
	313, 296, -1000, 385, 35, 35, 35, 116, 116, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -39, 569,
	-41, 177, 169, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, 166, 165, 158, -1, -1000, 869, 53, 251, 199,
	61, -43, -1000, 452, 447, 450, 37, 324, -1000, -1000,
	319, -1000, -1000, 130, 313, 313, 313, 457, 85, 85,
	-1000, -1000, 80, 78, 127, 125, 106, -21, -1000, 308,
	-57, 38, -1000, -1000, -1000, -1000, 313, 265, 1055, -1000,
	-1000, -44, -1000, 569, 569, 569, 569, 362, -14, -1000,
	1055, -24, 1107, -1000, 361, 143, -1000, -1000, -1000, 305,
	447, -1000, 1055, 869, -1000, -1000, 454, 449, 178, 52,
	-1000, 82, -1000, 76, -1000, -1000, -1000, -1000, -64, -67,
	-73, -1000, -1000, -1000, -1000, -1000, 1055, 313, -1000, -58,
	-60, -61, -62, -108, -1000, -1000, -1000, 195, 194, -1000,
	-1000, -1000, -1000, -1000, -1000, 313, 1055, 1055, 359, 251,
	-1000, -1000, 243, 132, -1000, 134, -1000, 452, 869, 1055,
	869, -1000, -1000, 249, 245, 244, 313, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 313, 313, 468, -1000, 1055, 1055,
	869, -1000, -1000, -1000, 447, 37, 131, -1000, 301, 301,
	301, 305, 313, 313, -1000, 328, -63, -1000, -66, -72,
	130, -1000, 467, 408, -1000, 301, -1000, -1000, -1000, 301,
	-1000, 301, -1000,
}
var yyPgo = [...]int{

	0, 579, 577, 21, 576, 575, 574, 573, 572, 571,
	570, 569, 487, 566, 565, 564, 11, 25, 563, 562,
	5, 561, 14, 560, 559, 242, 558, 3, 16, 7,
	557, 556, 17, 555, 2, 19, 1, 554, 553, 28,
	24, 552, 12, 551, 8, 549, 548, 15, 547, 546,
	545, 544, 9, 543, 6, 542, 10, 525, 23, 524,
	13, 4, 27, 283, 523, 521, 520, 519, 517, 512,
	0, 18, 511, 505, 504, 501, 499, 497, 181, 35,
	494, 491, 489, 488,
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
	29, 30, 30, 30, 30, 30, 30, 30, 30, 30,
	30, 30, 30, 30, 31, 31, 31, 31, 31, 31,
	31, 32, 32, 37, 37, 35, 35, 42, 36, 36,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 40, 40, 40, 40, 40, 40,
	40, 40, 39, 39, 39, 39, 39, 39, 39, 39,
	39, 41, 41, 41, 41, 41, 41, 41, 41, 41,
	41, 41, 41, 38, 38, 38, 38, 38, 38, 43,
	43, 43, 45, 48, 48, 46, 46, 47, 47, 49,
	49, 44, 44, 33, 33, 33, 33, 33, 33, 33,
	33, 33, 33, 50, 50, 51, 51, 52, 52, 53,
	53, 54, 55, 55, 55, 56, 56, 56, 56, 57,
	57, 57, 58, 58, 59, 59, 60, 60, 61, 61,
	62, 63, 63, 64, 64, 65, 65, 66, 66, 66,
	66, 66, 67, 67, 68, 68, 69, 69, 70, 71,
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
	6, 6, 6, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 2, 1, 2, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 5, 0, 1, 1, 2, 4, 4, 0,
	2, 1, 3, 1, 1, 1, 1, 1, 2, 2,
	2, 2, 1, 0, 3, 0, 2, 0, 3, 1,
	3, 2, 0, 1, 1, 0, 2, 4, 4, 0,
	2, 4, 0, 3, 1, 3, 0, 5, 1, 3,
	3, 0, 2, 0, 3, 0, 1, 1, 1, 1,
	1, 1, 0, 1, 0, 1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -72, -73, -74, -75, -76, -77, 5, 6,
	7, 8, 39, 137, 138, 140, 139, 124, 125, 126,
	128, 130, 129, -14, 88, 89, 90, 91, -12, -83,
	-12, -12, -12, -12, 141, -68, 143, 147, -65, 143,
	145, 141, 141, 142, 143, -12, 131, 134, 132, 133,
	-82, 135, -70, 41, -3, 22, -15, 23, -13, 35,
	-25, 41, 9, -61, 127, -62, -44, -70, 41, -64,
	146, 142, -70, 141, -70, 41, -63, 146, -70, -63,
	35, -79, 10, 29, -79, -78, 9, -70, 136, 50,
	-16, -17, 112, -20, 41, -29, -34, -30, 106, 50,
	-33, -44, -35, -43, -70, -38, 59, 60, 78, -45,
	26, 20, 21, 42, 43, 44, 55, 57, 58, 56,
	31, -42, 110, 111, 54, 146, 34, 96, 95, 115,
	-39, 119, 61, 62, 63, 64, 65, 66, 67, 68,
	69, 45, -25, 39, 117, -25, 92, 41, 51, 117,
	41, 106, -70, -71, 41, -71, 144, 41, 26, 103,
	-70, -25, -20, -34, 50, -79, -34, -78, -80, 9,
	27, -36, -34, 9, 92, -18, 104, 105, -70, 25,
	117, -31, 27, 106, 29, 30, 28, 107, 108, 109,
	110, 111, 112, 113, 116, 114, 115, 51, 52, 53,
	46, 47, 48, 49, -20, -29, -36, -3, -20, -34,
	-34, 50, 50, 50, 50, 50, -42, 50, 42, 42,
	42, 42, -48, -34, -58, 39, 50, -61, 41, -28,
	10, -62, -34, -70, -71, 26, -69, 148, -66, 140,
	138, 38, 139, 13, 41, 41, 41, -71, -58, 39,
	-79, -81, 9, 27, -34, -34, 149, 92, -21, -22,
	-24, 50, 41, -42, 136, 132, -17, 24, -20, -20,
	-70, 112, -34, 22, 19, 18, -35, 27, 29, 30,
	-34, -34, 31, 106, -34, -34, -34, -34, -34, -34,
	-34, -34, -34, -34, 149, 149, 149, 149, -16, 23,
	-16, -39, -40, 70, 71, 72, 73, 74, 75, 76,
	77, -39, -40, -17, -46, -47, 120, -32, 34, -3,
	-61, -59, -44, -28, -52, 13, -20, 103, -70, -71,
	-67, 144, -32, -61, -34, -34, -34, -28, 92, -23,
	93, 94, 95, 96, 97, 99, 100, -19, 41, 25,
	-22, 117, -42, -42, -42, -35, -34, -34, 104, 31,
	149, -16, 149, 92, 92, 92, 92, 92, -49, -47,
	122, -29, -34, -60, 103, -37, -35, -60, 149, 92,
	-52, -56, 15, 14, 41, 41, -50, 11, -22, -22,
	93, 98, 93, 98, 93, 93, 93, -26, 101, 145,
	102, 41, 149, 41, 136, 132, 104, -34, 149, -16,
	-16, -16, -16, -41, 79, 55, 56, 80, 81, 82,
	83, 84, 85, 86, 123, -34, 121, 121, 36, 92,
	-44, -56, -34, -53, -54, -20, -71, -51, 12, 14,
	103, 93, 93, 142, 142, 142, -34, 149, 149, 149,
	149, 149, 87, 87, -34, -34, 37, -35, 92, 16,
	92, -55, 32, 33, -52, -20, -36, -29, 50, 50,
	50, 7, -34, -34, -54, -56, -27, -70, -27, -27,
	-61, -57, 17, 40, 149, 92, 149, 149, 7, 27,
	-70, -70, -70,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 274, 265, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 263, 0, 0, 275, 0, 0,
	266, 0, 261, 0, 261, 0, 111, 111, 114, 0,
	0, 38, 0, 278, 19, 59, 0, 64, 55, 0,
	0, 101, 0, 26, 0, 258, 0, 221, 278, 0,
	0, 0, 279, 0, 279, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 111, 0, 114, 0, 0,
	17, 65, 67, 73, 278, 71, 72, 116, 0, 0,
	150, 151, 152, 0, 221, 0, 0, 0, 0, 173,
	0, 223, 224, 225, 226, 227, 0, 0, 0, 0,
	232, 146, 209, 210, 211, 203, 204, 205, 206, 207,
	208, 213, 182, 183, 184, 185, 186, 187, 188, 189,
	190, 57, 252, 0, 0, 109, 0, 27, 0, 0,
	279, 0, 276, 46, 0, 49, 0, 51, 262, 0,
	279, 252, 112, 113, 0, 41, 115, 111, 34, 0,
	0, 0, 148, 0, 0, 68, 0, 0, 74, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 134, 135, 136,
	137, 138, 139, 140, 119, 71, 0, 0, 0, -2,
	163, 0, 0, 0, 0, 0, 133, 0, 228, 229,
	230, 231, 0, 214, 0, 0, 0, 109, 102, 237,
	0, 259, 260, 222, 44, 264, 0, 0, 279, 272,
	267, 268, 269, 270, 271, 50, 52, 53, 0, 0,
	42, 43, 0, 0, 32, 33, 31, 0, 109, 76,
	82, 0, 94, 96, 97, 98, 66, 69, 117, 118,
	75, 70, 121, 0, 0, 0, 122, 0, 0, 0,
	127, 0, 131, 0, 153, 154, 155, 156, 157, 158,
	159, 160, 161, 162, 120, 145, 147, 164, 0, 0,
	0, 0, 0, 174, 175, 176, 177, 178, 179, 180,
	181, 0, 0, 0, 219, 215, 0, 256, 0, 142,
	256, 0, 254, 237, 245, 0, 110, 0, 277, 47,
	0, 273, 22, 23, 35, 36, 149, 233, 0, 0,
	85, 86, 0, 0, 0, 0, 0, 103, 83, 0,
	0, 0, 123, 124, 125, 126, 128, 0, 0, 132,
	165, 0, 167, 0, 0, 0, 0, 0, 0, 216,
	0, 71, 72, 20, 0, 141, 143, 21, 253, 0,
	245, 25, 0, 0, 279, 48, 235, 0, 77, 80,
	87, 0, 89, 0, 91, 92, 93, 78, 0, 0,
	0, 84, 79, 95, 99, 100, 0, 129, 166, 0,
	0, 0, 0, 0, 191, 192, 193, 194, 196, 198,
	199, 200, 201, 202, 212, 220, 0, 0, 0, 0,
	255, 24, 246, 238, 239, 242, 45, 237, 0, 0,
	0, 88, 90, 0, 0, 0, 130, 168, 169, 170,
	171, 172, 195, 197, 217, 218, 0, 144, 0, 0,
	0, 241, 243, 244, 245, 236, 234, -2, 0, 0,
	0, 0, 247, 248, 240, 249, 0, 107, 0, 0,
	257, 18, 0, 0, 104, 0, 105, 106, 250, 0,
	108, 0, 251,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 114, 107, 3,
	50, 149, 112, 110, 92, 111, 117, 113, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	52, 51, 53, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 109, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 108, 3, 54,
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
	98, 99, 100, 101, 102, 103, 104, 105, 106, 115,
	116, 118, 119, 120, 121, 122, 123, 124, 125, 126,
	127, 128, 129, 130, 131, 132, 133, 134, 135, 136,
	137, 138, 139, 140, 141, 142, 143, 144, 145, 146,
	147, 148,
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
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:708
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 120:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:712
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyDollar[2].boolExpr}
		}
	case 121:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:718
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, Right: yyDollar[3].valExpr}
		}
	case 122:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:722
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 123:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:726
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ALL, Right: yyDollar[4].subquery}
		}
	case 124:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:730
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ANY, Right: yyDollar[4].subquery}
		}
	case 125:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:734
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_SOME, Right: yyDollar[4].subquery}
		}
	case 126:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:738
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 127:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:742
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_LIKE, Right: yyDollar[3].valExpr}
		}
	case 128:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:746
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_LIKE, Right: yyDollar[4].valExpr}
		}
	case 129:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:750
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_BETWEEN, From: yyDollar[3].valExpr, To: yyDollar[5].valExpr}
		}
	case 130:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:754
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].valExpr, To: yyDollar[6].valExpr}
		}
	case 131:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:758
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NULL, Expr: yyDollar[1].valExpr}
		}
	case 132:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:762
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NOT_NULL, Expr: yyDollar[1].valExpr}
		}
	case 133:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:766
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 134:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:772
		{
			yyVAL.str = AST_EQ
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:776
		{
			yyVAL.str = AST_LT
		}
	case 136:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:780
		{
			yyVAL.str = AST_GT
		}
	case 137:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:784
		{
			yyVAL.str = AST_LE
		}
	case 138:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:788
		{
			yyVAL.str = AST_GE
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:792
		{
			yyVAL.str = AST_NE
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:796
		{
			yyVAL.str = AST_NSE
		}
	case 141:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:802
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:806
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:812
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 144:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:816
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:822
		{
			yyVAL.tuple = ValTuple(yyDollar[2].valExprs)
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:826
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:832
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 148:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:838
		{
			yyVAL.valExprs = ValExprs{yyDollar[1].valExpr}
		}
	case 149:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:842
		{
			yyVAL.valExprs = append(yyDollar[1].valExprs, yyDollar[3].valExpr)
		}
	case 150:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:848
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 151:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:852
		{
			yyVAL.valExpr = yyDollar[1].colName
		}
	case 152:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:856
		{
			yyVAL.valExpr = yyDollar[1].tuple
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:860
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITAND, Right: yyDollar[3].valExpr}
		}
	case 154:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:864
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITOR, Right: yyDollar[3].valExpr}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:868
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITXOR, Right: yyDollar[3].valExpr}
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:872
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_PLUS, Right: yyDollar[3].valExpr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:876
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MINUS, Right: yyDollar[3].valExpr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:880
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MULT, Right: yyDollar[3].valExpr}
		}
	case 159:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:884
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_DIV, Right: yyDollar[3].valExpr}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:888
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_IDIV, Right: yyDollar[3].valExpr}
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:892
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:896
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 163:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:900
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
		//line sql.y:915
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 165:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:919
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 166:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:923
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 167:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:927
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 168:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:931
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
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
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
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
			yyVAL.valExpr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 173:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:951
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 174:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:957
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 175:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:961
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 176:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:965
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 177:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:969
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 178:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:973
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 179:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:977
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 180:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:981
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 181:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:985
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 182:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:991
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 183:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:995
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 184:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:999
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 185:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1003
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 186:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1007
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 187:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1011
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1015
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1019
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1029
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1033
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1037
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1041
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 195:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1045
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 196:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1049
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 197:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1053
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 198:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1057
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 199:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1061
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1065
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1069
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1073
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1079
		{
			yyVAL.bytes = IF_BYTES
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1083
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 205:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1087
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1091
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1095
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1099
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 209:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1105
		{
			yyVAL.byt = AST_UPLUS
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1109
		{
			yyVAL.byt = AST_UMINUS
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1113
		{
			yyVAL.byt = AST_TILDA
		}
	case 212:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1119
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 213:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1124
		{
			yyVAL.valExpr = nil
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1128
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1134
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 216:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1138
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 217:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1144
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 218:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1148
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 219:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1153
		{
			yyVAL.valExpr = nil
		}
	case 220:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1157
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1163
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 222:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1167
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1173
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1177
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1181
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1185
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1189
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 228:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1193
		{
			yyVAL.valExpr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 229:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1197
		{
			yyVAL.valExpr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 230:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1201
		{
			yyVAL.valExpr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 231:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1205
		{
			yyVAL.valExpr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1209
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 233:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1214
		{
			yyVAL.valExprs = nil
		}
	case 234:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1218
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 235:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1223
		{
			yyVAL.expr = nil
		}
	case 236:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1227
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 237:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1232
		{
			yyVAL.orderBy = nil
		}
	case 238:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1236
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1242
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 240:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1246
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 241:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1252
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 242:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1257
		{
			yyVAL.str = AST_ASC
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1261
		{
			yyVAL.str = AST_ASC
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1265
		{
			yyVAL.str = AST_DESC
		}
	case 245:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1270
		{
			yyVAL.limit = nil
		}
	case 246:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1274
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 247:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1278
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 248:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1282
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].valExpr, Rowcount: yyDollar[2].valExpr}
		}
	case 249:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1287
		{
			yyVAL.str = ""
		}
	case 250:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1291
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 251:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1295
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
	case 252:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1308
		{
			yyVAL.columns = nil
		}
	case 253:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1312
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 254:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1318
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 255:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1322
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 256:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1327
		{
			yyVAL.updateExprs = nil
		}
	case 257:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1331
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 258:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1337
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 259:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1341
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 260:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1347
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 261:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1352
		{
			yyVAL.empty = struct{}{}
		}
	case 262:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1354
		{
			yyVAL.empty = struct{}{}
		}
	case 263:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1357
		{
			yyVAL.empty = struct{}{}
		}
	case 264:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1359
		{
			yyVAL.empty = struct{}{}
		}
	case 265:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1362
		{
			yyVAL.empty = struct{}{}
		}
	case 266:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1364
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
		//line sql.y:1370
		{
			yyVAL.empty = struct{}{}
		}
	case 269:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1372
		{
			yyVAL.empty = struct{}{}
		}
	case 270:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1374
		{
			yyVAL.empty = struct{}{}
		}
	case 271:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1376
		{
			yyVAL.empty = struct{}{}
		}
	case 272:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1379
		{
			yyVAL.empty = struct{}{}
		}
	case 273:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1381
		{
			yyVAL.empty = struct{}{}
		}
	case 274:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1384
		{
			yyVAL.empty = struct{}{}
		}
	case 275:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1386
		{
			yyVAL.empty = struct{}{}
		}
	case 276:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1389
		{
			yyVAL.empty = struct{}{}
		}
	case 277:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1391
		{
			yyVAL.empty = struct{}{}
		}
	case 278:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1395
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 279:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1400
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
