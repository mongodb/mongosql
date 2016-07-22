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
	tuple       Tuple
	exprs       Exprs
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
const WHERE = 57351
const GROUP = 57352
const HAVING = 57353
const ORDER = 57354
const BY = 57355
const LIMIT = 57356
const OFFSET = 57357
const FOR = 57358
const SOME = 57359
const ANY = 57360
const TRUE = 57361
const FALSE = 57362
const UNKNOWN = 57363
const ALL = 57364
const DISTINCT = 57365
const PRECISION = 57366
const AS = 57367
const EXISTS = 57368
const NULL = 57369
const ASC = 57370
const DESC = 57371
const VALUES = 57372
const INTO = 57373
const DUPLICATE = 57374
const KEY = 57375
const DEFAULT = 57376
const SET = 57377
const LOCK = 57378
const ID = 57379
const STRING = 57380
const NUMBER = 57381
const VALUE_ARG = 57382
const COMMENT = 57383
const LPAREN = 57384
const RPAREN = 57385
const TILDE = 57386
const DATE = 57387
const DATETIME = 57388
const TIME = 57389
const TIMESTAMP = 57390
const CURRENT_TIMESTAMP = 57391
const TIMESTAMPADD = 57392
const TIMESTAMPDIFF = 57393
const YEAR = 57394
const QUARTER = 57395
const MONTH = 57396
const WEEK = 57397
const DAY = 57398
const HOUR = 57399
const MINUTE = 57400
const SECOND = 57401
const MICROSECOND = 57402
const SQL_TSI_YEAR = 57403
const SQL_TSI_QUARTER = 57404
const SQL_TSI_MONTH = 57405
const SQL_TSI_WEEK = 57406
const SQL_TSI_DAY = 57407
const SQL_TSI_HOUR = 57408
const SQL_TSI_MINUTE = 57409
const SQL_TSI_SECOND = 57410
const CONVERT = 57411
const CAST = 57412
const CHAR = 57413
const SIGNED = 57414
const UNSIGNED = 57415
const SQL_BIGINT = 57416
const SQL_VARCHAR = 57417
const SQL_DATE = 57418
const SQL_TIMESTAMP = 57419
const SQL_DOUBLE = 57420
const INTEGER = 57421
const FROM = 57422
const UNION = 57423
const MINUS = 57424
const EXCEPT = 57425
const INTERSECT = 57426
const COMMA = 57427
const JOIN = 57428
const STRAIGHT_JOIN = 57429
const LEFT = 57430
const RIGHT = 57431
const INNER = 57432
const OUTER = 57433
const CROSS = 57434
const NATURAL = 57435
const USE = 57436
const FORCE = 57437
const ON = 57438
const NOT = 57439
const OR = 57440
const XOR = 57441
const BETWEEN = 57442
const AND = 57443
const NE = 57444
const EQ = 57445
const NULL_SAFE_EQUAL = 57446
const IS = 57447
const LIKE = 57448
const REGEXP = 57449
const IN = 57450
const LT = 57451
const GT = 57452
const LE = 57453
const GE = 57454
const BIT_AND = 57455
const BIT_OR = 57456
const CARET = 57457
const PLUS = 57458
const SUB = 57459
const TIMES = 57460
const MOD = 57461
const DIV = 57462
const IDIV = 57463
const DOT = 57464
const UNARY = 57465
const CASE = 57466
const WHEN = 57467
const THEN = 57468
const ELSE = 57469
const END = 57470
const BEGIN = 57471
const COMMIT = 57472
const ROLLBACK = 57473
const NAMES = 57474
const REPLACE = 57475
const ADMIN = 57476
const SHOW = 57477
const DATABASES = 57478
const TABLES = 57479
const PROXY = 57480
const VARIABLES = 57481
const FULL = 57482
const SESSION = 57483
const GLOBAL = 57484
const COLUMNS = 57485
const CREATE = 57486
const ALTER = 57487
const DROP = 57488
const RENAME = 57489
const TABLE = 57490
const INDEX = 57491
const VIEW = 57492
const TO = 57493
const IGNORE = 57494
const IF = 57495
const UNIQUE = 57496
const USING = 57497

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"LEX_ERROR",
	"SELECT",
	"INSERT",
	"UPDATE",
	"DELETE",
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
	"LPAREN",
	"RPAREN",
	"TILDE",
	"DATE",
	"DATETIME",
	"TIME",
	"TIMESTAMP",
	"CURRENT_TIMESTAMP",
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
	"CAST",
	"CHAR",
	"SIGNED",
	"UNSIGNED",
	"SQL_BIGINT",
	"SQL_VARCHAR",
	"SQL_DATE",
	"SQL_TIMESTAMP",
	"SQL_DOUBLE",
	"INTEGER",
	"FROM",
	"UNION",
	"MINUS",
	"EXCEPT",
	"INTERSECT",
	"COMMA",
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
	"NOT",
	"OR",
	"XOR",
	"BETWEEN",
	"AND",
	"NE",
	"EQ",
	"NULL_SAFE_EQUAL",
	"IS",
	"LIKE",
	"REGEXP",
	"IN",
	"LT",
	"GT",
	"LE",
	"GE",
	"BIT_AND",
	"BIT_OR",
	"CARET",
	"PLUS",
	"SUB",
	"TIMES",
	"MOD",
	"DIV",
	"IDIV",
	"DOT",
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
	143, 37,
	-2, 39,
	-1, 319,
	100, 0,
	-2, 171,
	-1, 392,
	100, 0,
	-2, 172,
}

const yyNprod = 295
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1414

var yyAct = [...]int{

	114, 417, 468, 184, 511, 102, 354, 447, 409, 109,
	345, 129, 103, 93, 271, 141, 329, 236, 241, 131,
	249, 108, 77, 75, 89, 18, 19, 20, 21, 347,
	44, 166, 46, 64, 82, 361, 47, 434, 436, 49,
	85, 50, 169, 290, 79, 479, 478, 84, 477, 83,
	86, 51, 439, 255, 90, 22, 96, 233, 3, 100,
	99, 52, 53, 54, 18, 78, 56, 58, 59, 459,
	63, 61, 62, 105, 346, 253, 80, 274, 256, 346,
	381, 407, 273, 216, 165, 192, 193, 194, 197, 195,
	196, 66, 173, 162, 157, 435, 274, 194, 197, 195,
	196, 273, 524, 161, 94, 264, 214, 313, 65, 164,
	177, 178, 97, 311, 312, 310, 182, 474, 168, 201,
	202, 203, 204, 189, 190, 191, 192, 193, 194, 197,
	195, 196, 219, 265, 232, 189, 190, 191, 192, 193,
	194, 197, 195, 196, 183, 159, 410, 357, 172, 27,
	28, 29, 441, 30, 32, 31, 410, 79, 440, 476,
	79, 522, 475, 245, 23, 24, 26, 25, 175, 176,
	432, 76, 179, 431, 428, 185, 430, 277, 78, 429,
	239, 78, 243, 276, 217, 218, 252, 254, 251, 321,
	159, 269, 260, 426, 262, 246, 277, 494, 427, 463,
	278, 95, 276, 520, 185, 259, 275, 521, 402, 235,
	401, 400, 180, 142, 143, 144, 320, 245, 309, 186,
	72, 315, 399, 317, 187, 98, 74, 323, 325, 326,
	370, 371, 372, 373, 374, 244, 375, 376, 79, 79,
	328, 338, 339, 295, 297, 299, 301, 303, 305, 520,
	358, 34, 35, 36, 37, 519, 266, 267, 353, 78,
	352, 350, 79, 280, 281, 282, 283, 284, 285, 286,
	287, 288, 289, 294, 296, 298, 300, 302, 304, 306,
	307, 308, 359, 78, 314, 363, 318, 319, 380, 367,
	362, 316, 155, 275, 349, 158, 343, 520, 483, 482,
	340, 341, 438, 486, 485, 505, 484, 488, 481, 504,
	382, 480, 442, 174, 414, 383, 356, 384, 349, 385,
	389, 386, 242, 387, 242, 388, 487, 503, 394, 502,
	396, 130, 220, 227, 34, 35, 36, 37, 364, 365,
	187, 187, 88, 366, 226, 370, 371, 372, 373, 374,
	187, 375, 376, 187, 187, 406, 415, 398, 412, 413,
	416, 213, 205, 199, 198, 200, 211, 210, 212, 208,
	201, 202, 203, 204, 189, 190, 191, 192, 193, 194,
	197, 195, 196, 424, 425, 390, 391, 392, 275, 275,
	225, 224, 223, 154, 261, 237, 222, 91, 368, 187,
	159, 238, 238, 221, 101, 443, 444, 445, 446, 397,
	379, 231, 458, 230, 229, 228, 79, 65, 465, 80,
	408, 437, 378, 421, 420, 258, 209, 206, 207, 213,
	205, 199, 198, 200, 211, 210, 212, 464, 201, 202,
	203, 204, 189, 190, 191, 192, 193, 194, 197, 195,
	196, 187, 470, 199, 198, 200, 211, 210, 212, 208,
	201, 202, 203, 204, 189, 190, 191, 192, 193, 194,
	197, 195, 196, 491, 517, 257, 395, 500, 498, 240,
	342, 460, 73, 170, 449, 450, 167, 268, 163, 160,
	87, 156, 466, 469, 518, 490, 462, 509, 18, 92,
	510, 71, 247, 512, 512, 512, 171, 79, 513, 514,
	448, 451, 452, 453, 454, 455, 456, 457, 187, 38,
	279, 525, 269, 348, 69, 526, 67, 527, 78, 269,
	515, 355, 418, 473, 419, 489, 142, 143, 144, 40,
	41, 42, 43, 423, 393, 472, 499, 185, 501, 242,
	55, 142, 143, 144, 523, 324, 506, 18, 112, 128,
	39, 57, 137, 60, 263, 181, 507, 508, 469, 106,
	121, 122, 123, 17, 130, 322, 134, 124, 127, 125,
	126, 116, 117, 118, 145, 146, 147, 148, 149, 150,
	151, 152, 153, 16, 15, 14, 13, 12, 248, 45,
	360, 119, 120, 145, 146, 147, 148, 149, 150, 151,
	152, 153, 330, 331, 332, 333, 334, 335, 336, 337,
	139, 138, 250, 48, 81, 351, 516, 495, 467, 110,
	471, 422, 405, 234, 344, 113, 111, 115, 411, 107,
	433, 272, 369, 270, 377, 188, 68, 33, 132, 133,
	104, 140, 70, 11, 10, 9, 135, 8, 7, 292,
	293, 142, 143, 144, 291, 6, 5, 4, 112, 128,
	2, 1, 137, 0, 0, 0, 0, 0, 0, 80,
	121, 122, 123, 0, 130, 136, 134, 124, 127, 125,
	126, 116, 117, 118, 145, 146, 147, 148, 149, 150,
	151, 152, 153, 0, 0, 0, 0, 0, 0, 0,
	0, 119, 120, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	139, 138, 0, 0, 0, 0, 0, 0, 0, 110,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 142,
	143, 144, 0, 0, 0, 0, 112, 128, 132, 133,
	137, 140, 0, 0, 0, 0, 135, 106, 121, 122,
	123, 0, 130, 327, 134, 124, 127, 125, 126, 116,
	117, 118, 145, 146, 147, 148, 149, 150, 151, 152,
	153, 0, 0, 0, 0, 136, 0, 0, 0, 119,
	120, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 139, 138,
	0, 0, 0, 0, 0, 0, 0, 110, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 132, 133, 104, 140,
	0, 0, 0, 0, 135, 0, 0, 0, 0, 142,
	143, 144, 0, 0, 0, 0, 112, 128, 0, 0,
	137, 0, 0, 0, 0, 0, 0, 106, 121, 122,
	123, 0, 130, 136, 134, 124, 127, 125, 126, 116,
	117, 118, 145, 146, 147, 148, 149, 150, 151, 152,
	153, 0, 0, 0, 0, 0, 0, 0, 0, 119,
	120, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 139, 138,
	0, 0, 0, 0, 0, 0, 0, 110, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 18, 132, 133, 104, 140,
	0, 0, 0, 0, 135, 0, 0, 0, 0, 142,
	143, 144, 0, 0, 0, 0, 112, 128, 0, 0,
	137, 0, 0, 0, 0, 0, 0, 80, 121, 122,
	123, 0, 130, 136, 134, 124, 127, 125, 126, 116,
	117, 118, 145, 146, 147, 148, 149, 150, 151, 152,
	153, 0, 0, 0, 0, 0, 0, 0, 0, 119,
	120, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 139, 138,
	0, 0, 0, 0, 0, 0, 0, 110, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 142, 143, 144,
	0, 0, 0, 0, 112, 128, 132, 133, 137, 140,
	0, 0, 0, 0, 135, 80, 121, 122, 123, 0,
	130, 0, 134, 124, 127, 125, 126, 116, 117, 118,
	145, 146, 147, 148, 149, 150, 151, 152, 153, 0,
	0, 0, 0, 136, 0, 0, 0, 119, 120, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 496, 497, 139, 138, 0, 0,
	0, 0, 0, 0, 0, 110, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 493, 0,
	0, 0, 0, 0, 132, 133, 0, 140, 0, 0,
	0, 0, 135, 209, 206, 207, 213, 205, 199, 198,
	200, 211, 210, 212, 208, 201, 202, 203, 204, 189,
	190, 191, 192, 193, 194, 197, 195, 196, 0, 0,
	0, 136, 461, 209, 206, 207, 213, 205, 199, 198,
	200, 211, 210, 212, 208, 201, 202, 203, 204, 189,
	190, 191, 192, 193, 194, 197, 195, 196, 492, 404,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	209, 206, 207, 213, 205, 199, 198, 200, 211, 210,
	212, 208, 201, 202, 203, 204, 189, 190, 191, 192,
	193, 194, 197, 195, 196, 215, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 65, 205, 199,
	198, 200, 211, 210, 212, 208, 201, 202, 203, 204,
	189, 190, 191, 192, 193, 194, 197, 195, 196, 0,
	0, 209, 206, 207, 213, 205, 199, 198, 200, 211,
	210, 212, 208, 201, 202, 203, 204, 189, 190, 191,
	192, 193, 194, 197, 195, 196, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 209, 206, 207,
	213, 205, 199, 198, 200, 211, 210, 212, 208, 201,
	202, 203, 204, 189, 190, 191, 192, 193, 194, 197,
	195, 196, 403, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 209, 206, 207, 213, 205, 199,
	198, 200, 211, 210, 212, 208, 201, 202, 203, 204,
	189, 190, 191, 192, 193, 194, 197, 195, 196, 209,
	206, 207, 213, 205, 199, 198, 200, 211, 210, 212,
	208, 201, 202, 203, 204, 189, 190, 191, 192, 193,
	194, 197, 195, 196,
}
var yyPact = [...]int{

	20, -1000, -1000, 170, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -118, -111, -97, -87, -1000, -1000, -1000,
	-1000, -70, 380, 552, 504, -1000, -1000, -1000, 501, -1000,
	470, 445, 146, 39, -119, -100, 380, -1000, -108, 380,
	-1000, 453, -129, 380, -129, 468, 95, -83, 145, 380,
	-84, -1000, -1000, -1000, 362, -1000, -1000, -1000, 840, -1000,
	352, 445, 456, -28, 445, 105, 452, -1000, 0, -1000,
	-29, 451, 12, 380, -1000, 449, -1000, -109, 446, 480,
	52, 380, 445, -1000, 1038, 1038, 95, 95, 1038, 145,
	36, 1038, 139, -1000, -1000, 1230, -39, -1000, -1000, -1000,
	1038, 1038, 290, -1000, 361, 354, 350, 349, 348, 302,
	291, -1000, -1000, -1000, 377, 376, 375, 373, -1000, -1000,
	950, -1000, -1000, -1000, -1000, 1038, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, 360, 382, 442, 540, 382,
	-1000, 1038, 380, -1000, 476, -135, -1000, 41, -1000, 438,
	-1000, -1000, 388, -1000, 359, 1292, 1292, -1000, -1000, 1292,
	95, 25, 1038, 1038, 444, 1292, 40, 840, 496, 1038,
	1038, 1038, 1038, 1038, 1038, 1038, 1038, 1038, 642, 642,
	642, 642, 642, 642, 642, 1038, 1038, 1038, 289, 7,
	1038, 194, 1038, 1038, -1000, 380, 71, 1292, -1000, -1000,
	552, 532, 840, 730, 551, 551, 1038, 1038, -1000, -1000,
	-1000, -1000, 437, 253, -51, 1292, 493, 382, 382, 315,
	-1000, 519, 1038, -1000, 1292, -1000, -1000, -1000, 51, 380,
	-1000, -116, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	493, 382, -1000, -1000, 1038, 1038, 1292, 329, -1000, 1038,
	313, 144, 385, 59, -42, -1000, -1000, -1000, -1000, -1000,
	-31, -31, -31, -21, -21, -1000, -1000, -1000, -1000, 10,
	290, -1000, -1000, -1000, 10, 290, 10, 290, 22, 290,
	22, 290, 22, 290, 22, 290, 351, 261, 261, -1000,
	289, 1038, 1038, 1038, 10, -1000, 517, -1000, 10, 1167,
	-1000, -1000, -1000, 433, 840, 366, 314, -1000, 137, 126,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 125, 123,
	1267, 1194, -1000, -1000, -46, -1000, 1038, 50, 289, 170,
	60, 271, -1000, 519, 518, 521, 1292, 387, -1000, -1000,
	386, -1000, -1000, 105, 1292, 1292, 1292, 533, 40, 40,
	-1000, -1000, 107, 88, 90, 87, 84, -57, -1000, 384,
	259, 15, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	10, 10, 1167, -1000, -1000, -1000, 269, -1000, -1000, 840,
	840, 840, 840, 439, 439, -59, -1000, 1038, 1066, -1000,
	464, 114, -1000, -1000, -1000, 382, 518, -1000, 1038, 1038,
	-1000, -1000, 534, 520, 144, 21, -1000, 76, -1000, 73,
	-1000, -1000, -1000, -1000, -101, -103, -104, -1000, -1000, -1000,
	-1000, -1000, -1000, 268, 265, 256, 255, 263, -1000, -1000,
	-1000, 225, 224, -1000, -1000, -1000, -1000, -1000, 283, -1000,
	1292, 1038, 462, 289, -1000, -1000, 1133, 112, -1000, 1096,
	-1000, 519, 1038, 1038, 1038, -1000, -1000, 287, 285, 267,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 262, 1292,
	549, -1000, 1038, 1038, 1038, -1000, -1000, -1000, 518, 1292,
	106, 1292, 380, 380, 380, -1000, 382, 1292, 1292, -1000,
	458, 212, -1000, 164, 118, 105, -1000, 547, -6, -1000,
	380, -1000, -1000, -1000, 380, -1000, 380, -1000,
}
var yyPgo = [...]int{

	0, 671, 670, 57, 667, 666, 665, 658, 657, 655,
	654, 653, 519, 652, 647, 43, 646, 5, 12, 645,
	644, 73, 643, 14, 642, 641, 220, 640, 4, 18,
	29, 639, 9, 11, 3, 638, 637, 15, 16, 7,
	19, 636, 21, 635, 634, 10, 633, 632, 631, 630,
	6, 628, 2, 627, 1, 626, 17, 625, 8, 23,
	22, 342, 624, 623, 622, 600, 599, 598, 0, 31,
	597, 596, 595, 594, 593, 573, 112, 13, 565, 564,
	563, 561, 560,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 73, 73, 5, 6, 7, 7, 70, 71,
	72, 75, 78, 78, 79, 79, 79, 80, 80, 81,
	81, 81, 74, 74, 74, 74, 74, 8, 8, 8,
	9, 9, 9, 10, 11, 11, 11, 82, 12, 13,
	13, 14, 14, 14, 14, 14, 16, 16, 17, 17,
	18, 18, 18, 18, 19, 19, 19, 22, 22, 23,
	23, 23, 23, 20, 20, 20, 24, 24, 24, 24,
	24, 24, 24, 24, 24, 25, 25, 25, 25, 25,
	25, 25, 26, 26, 27, 27, 27, 27, 28, 28,
	29, 29, 77, 77, 77, 76, 76, 15, 15, 15,
	30, 30, 35, 35, 32, 32, 40, 34, 34, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 38, 38,
	38, 38, 38, 38, 38, 38, 37, 37, 37, 37,
	37, 37, 37, 37, 37, 39, 39, 39, 39, 39,
	39, 39, 39, 39, 39, 39, 39, 36, 36, 36,
	36, 36, 36, 41, 41, 41, 43, 46, 46, 44,
	44, 45, 47, 47, 42, 42, 31, 31, 31, 31,
	31, 31, 31, 31, 31, 33, 33, 33, 48, 48,
	49, 49, 50, 50, 51, 51, 52, 53, 53, 53,
	54, 54, 54, 54, 55, 55, 55, 56, 56, 57,
	57, 58, 58, 59, 59, 60, 61, 61, 62, 62,
	63, 63, 64, 64, 64, 64, 64, 65, 65, 66,
	66, 67, 67, 68, 69,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 4, 12, 3,
	7, 7, 6, 6, 8, 7, 3, 4, 1, 1,
	1, 5, 2, 2, 0, 2, 2, 0, 1, 0,
	1, 1, 3, 4, 4, 5, 5, 5, 8, 4,
	6, 7, 4, 5, 4, 5, 5, 0, 2, 0,
	2, 1, 2, 1, 1, 1, 0, 1, 1, 3,
	1, 2, 3, 3, 0, 1, 2, 1, 3, 3,
	3, 3, 5, 0, 1, 2, 1, 1, 2, 3,
	2, 3, 2, 2, 2, 1, 3, 1, 1, 1,
	3, 3, 1, 3, 0, 5, 5, 5, 1, 3,
	0, 2, 0, 2, 2, 0, 2, 1, 1, 1,
	2, 1, 1, 3, 3, 1, 3, 1, 3, 1,
	1, 1, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 4, 3, 4, 3, 4, 3, 4, 3,
	4, 3, 4, 3, 4, 3, 3, 3, 2, 2,
	3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
	2, 3, 4, 1, 3, 4, 5, 4, 1, 4,
	3, 6, 6, 6, 6, 6, 6, 7, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 2,
	1, 2, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 5, 0, 1, 1,
	2, 4, 0, 2, 1, 3, 1, 1, 1, 2,
	2, 2, 2, 1, 1, 1, 1, 1, 0, 3,
	0, 2, 0, 3, 1, 3, 2, 0, 1, 1,
	0, 2, 4, 4, 0, 2, 4, 0, 3, 1,
	3, 0, 5, 1, 3, 3, 0, 2, 0, 3,
	0, 1, 1, 1, 1, 1, 1, 0, 1, 0,
	1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -70, -71, -72, -73, -74, -75, 5, 6,
	7, 8, 35, 144, 145, 147, 146, 129, 130, 131,
	133, 135, 134, -14, 81, 82, 83, 84, -12, -82,
	-12, -12, -12, -12, 148, -66, 150, 154, -63, 150,
	152, 148, 148, 149, 150, -12, 136, -81, 137, 138,
	-80, 141, 142, 140, -68, 37, -3, 22, -16, 23,
	-13, 31, -26, 37, 80, -59, 132, -60, -42, -68,
	37, -62, 153, 149, -68, 148, -68, 37, -61, 153,
	-68, -61, 31, -77, 9, 106, 139, -76, 80, -68,
	143, 42, -17, -18, 118, -21, 37, -31, -42, -32,
	97, -41, 26, -43, -68, -36, 49, 50, 51, 69,
	70, 38, 39, 40, 45, 47, 48, 46, 27, -33,
	42, -40, 116, 117, 44, 124, 153, 30, 89, 88,
	119, -37, 19, 20, 21, 52, 53, 54, 55, 56,
	57, 58, 59, 60, 41, -26, 35, 122, -26, 85,
	37, 103, 122, 37, 97, -68, -69, 37, -69, 151,
	37, 26, 96, -68, -26, -21, -21, -77, -77, -21,
	-76, -78, 80, 108, -34, -21, 80, 85, -19, 113,
	114, 115, 116, 117, 118, 120, 121, 119, 103, 102,
	104, 109, 110, 111, 112, 101, 98, 99, 108, 97,
	106, 105, 107, 100, -68, 25, 122, -21, -21, -40,
	42, 42, 42, 42, 42, 42, 42, 42, 38, 38,
	38, 38, -34, -3, -46, -21, -56, 35, 42, -59,
	37, -29, 9, -60, -21, -68, -69, 26, -67, 155,
	-64, 147, 145, 34, 146, 12, 37, 37, 37, -69,
	-56, 35, -77, -79, 80, 108, -21, -21, 43, 85,
	-22, -23, -25, 42, 37, -40, 143, 137, -18, 24,
	-21, -21, -21, -21, -21, -21, -21, -21, -21, -21,
	-15, 22, 17, 18, -21, -15, -21, -15, -21, -15,
	-21, -15, -21, -15, -21, -15, -21, -21, -21, -32,
	108, 106, 107, 100, -21, 27, 97, -33, -21, -21,
	-68, 118, 43, -17, 23, -17, -17, 43, -37, -38,
	61, 62, 63, 64, 65, 66, 67, 68, -37, -38,
	-21, -21, 43, 43, -44, -45, 125, -30, 30, -3,
	-59, -57, -42, -29, -50, 12, -21, 96, -68, -69,
	-65, 151, -30, -59, -21, -21, -21, -29, 85, -24,
	86, 87, 88, 89, 90, 92, 93, -20, 37, 25,
	-23, 122, -40, -40, -40, -40, -40, -40, -40, -32,
	-21, -21, -21, 27, -33, 43, -17, 43, 43, 85,
	85, 85, 85, 85, 25, -47, -45, 127, -21, -58,
	96, -35, -32, -58, 43, 85, -50, -54, 14, 13,
	37, 37, -48, 10, -23, -23, 86, 91, 86, 91,
	86, 86, 86, -27, 94, 152, 95, 37, 43, 37,
	143, 137, 43, -17, -17, -17, -17, -39, 71, 45,
	46, 72, 73, 74, 75, 76, 77, 78, -39, 128,
	-21, 126, 32, 85, -42, -54, -21, -51, -52, -21,
	-69, -49, 11, 13, 96, 86, 86, 149, 149, 149,
	43, 43, 43, 43, 43, 79, 79, 43, 24, -21,
	33, -32, 85, 15, 85, -53, 28, 29, -50, -21,
	-34, -21, 42, 42, 42, 43, 7, -21, -21, -52,
	-54, -28, -68, -28, -28, -59, -55, 16, 36, 43,
	85, 43, 43, 7, 108, -68, -68, -68,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 57, 57,
	57, 57, 57, 289, 280, 0, 0, 28, 29, 30,
	57, -2, 0, 0, 61, 63, 64, 65, 66, 59,
	0, 0, 0, 0, 278, 0, 0, 290, 0, 0,
	281, 0, 276, 0, 276, 0, 112, 0, 115, 0,
	0, 40, 41, 38, 0, 293, 19, 62, 0, 67,
	58, 0, 0, 102, 0, 26, 0, 273, 0, 234,
	293, 0, 0, 0, 294, 0, 294, 0, 0, 0,
	0, 0, 0, 42, 0, 0, 112, 112, 0, 115,
	0, 0, 17, 68, 70, 74, 293, 129, 130, 131,
	0, 0, 0, 173, 234, 0, 178, 0, 0, 0,
	0, 236, 237, 238, 0, 0, 0, 0, 243, 244,
	0, 125, 223, 224, 225, 227, 217, 218, 219, 220,
	221, 222, 245, 246, 247, 196, 197, 198, 199, 200,
	201, 202, 203, 204, 60, 267, 0, 0, 110, 0,
	27, 0, 0, 294, 0, 291, 49, 0, 52, 0,
	54, 277, 0, 294, 267, 113, 114, 43, 44, 116,
	112, 34, 0, 0, 0, 127, 0, 0, 71, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 75, 0, 0, 158, 159, 170,
	0, 0, 0, 0, 0, 0, 0, 0, 239, 240,
	241, 242, 0, 0, 0, 228, 0, 0, 0, 110,
	103, 252, 0, 274, 275, 235, 47, 279, 0, 0,
	294, 287, 282, 283, 284, 285, 286, 53, 55, 56,
	0, 0, 45, 46, 0, 0, 32, 33, 31, 0,
	110, 77, 83, 0, 95, 97, 98, 99, 69, 72,
	132, 133, 134, 135, 136, 137, 138, 139, 140, 141,
	0, 117, 118, 119, 143, 0, 145, 0, 147, 0,
	149, 0, 151, 0, 153, 0, 155, 156, 157, 160,
	0, 0, 0, 0, 162, 164, 0, 166, 168, -2,
	76, 73, 174, 0, 0, 0, 0, 180, 0, 0,
	188, 189, 190, 191, 192, 193, 194, 195, 0, 0,
	0, 0, 124, 126, 232, 229, 0, 271, 0, 121,
	271, 0, 269, 252, 260, 0, 111, 0, 292, 50,
	0, 288, 22, 23, 35, 36, 128, 248, 0, 0,
	86, 87, 0, 0, 0, 0, 0, 104, 84, 0,
	0, 0, 142, 144, 146, 148, 150, 152, 154, 161,
	163, 169, -2, 165, 167, 175, 0, 177, 179, 0,
	0, 0, 0, 0, 0, 0, 230, 0, 0, 20,
	0, 120, 122, 21, 268, 0, 260, 25, 0, 0,
	294, 51, 250, 0, 78, 81, 88, 0, 90, 0,
	92, 93, 94, 79, 0, 0, 0, 85, 80, 96,
	100, 101, 176, 0, 0, 0, 0, 0, 205, 206,
	207, 208, 210, 212, 213, 214, 215, 216, 0, 226,
	233, 0, 0, 0, 270, 24, 261, 253, 254, 257,
	48, 252, 0, 0, 0, 89, 91, 0, 0, 0,
	181, 182, 183, 184, 185, 209, 211, 186, 0, 231,
	0, 123, 0, 0, 0, 256, 258, 259, 260, 251,
	249, 82, 0, 0, 0, 187, 0, 262, 263, 255,
	264, 0, 108, 0, 0, 272, 18, 0, 0, 105,
	0, 106, 107, 265, 0, 109, 0, 266,
}
var yyTok1 = [...]int{

	1,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
	52, 53, 54, 55, 56, 57, 58, 59, 60, 61,
	62, 63, 64, 65, 66, 67, 68, 69, 70, 71,
	72, 73, 74, 75, 76, 77, 78, 79, 80, 81,
	82, 83, 84, 85, 86, 87, 88, 89, 90, 91,
	92, 93, 94, 95, 96, 97, 98, 99, 100, 101,
	102, 103, 104, 105, 106, 107, 108, 109, 110, 111,
	112, 113, 114, 115, 116, 117, 118, 119, 120, 121,
	122, 123, 124, 125, 126, 127, 128, 129, 130, 131,
	132, 133, 134, 135, 136, 137, 138, 139, 140, 141,
	142, 143, 144, 145, 146, 147, 148, 149, 150, 151,
	152, 153, 154, 155,
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
		//line sql.y:198
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:204
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:224
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 18:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:228
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].exprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:232
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 20:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:239
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:243
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
		//line sql.y:255
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:259
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
		//line sql.y:272
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 25:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:278
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:284
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:288
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:294
		{
			yyVAL.statement = &Begin{}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:300
		{
			yyVAL.statement = &Commit{}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:306
		{
			yyVAL.statement = &Rollback{}
		}
	case 31:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:312
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].exprs}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:318
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:322
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 34:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:327
		{
			yyVAL.expr = nil
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:331
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:335
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 37:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:340
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:344
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 39:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:349
		{
			yyVAL.str = AST_SHOW_SESSION_VARIABLE
		}
	case 40:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:353
		{
			yyVAL.str = AST_SHOW_SESSION_VARIABLE
		}
	case 41:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:357
		{
			yyVAL.str = AST_SHOW_GLOBAL_VARIABLE
		}
	case 42:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:364
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 43:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:368
		{
			yyVAL.statement = &Show{Section: "variables", Modifier: yyDollar[2].str, LikeOrWhere: yyDollar[4].expr}
		}
	case 44:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:372
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].expr, LikeOrWhere: yyDollar[4].expr}
		}
	case 45:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:376
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].expr, LikeOrWhere: yyDollar[5].expr}
		}
	case 46:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:380
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].expr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].expr}
		}
	case 47:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:386
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 48:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:390
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 49:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:395
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 50:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:401
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 51:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:405
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 52:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:410
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 53:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:416
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 54:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:422
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 55:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:426
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 56:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:431
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 57:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:436
		{
			SetAllowComments(yylex, true)
		}
	case 58:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:440
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 59:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:446
		{
			yyVAL.bytes2 = nil
		}
	case 60:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:450
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:456
		{
			yyVAL.str = AST_UNION
		}
	case 62:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:460
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 63:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:464
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:468
		{
			yyVAL.str = AST_EXCEPT
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:472
		{
			yyVAL.str = AST_INTERSECT
		}
	case 66:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:477
		{
			yyVAL.str = ""
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:481
		{
			yyVAL.str = AST_DISTINCT
		}
	case 68:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:487
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:491
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 70:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:497
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 71:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:501
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 72:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:505
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 73:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:509
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 74:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:514
		{
			yyVAL.bytes = nil
		}
	case 75:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:518
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 76:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:522
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 77:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:528
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:532
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:538
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:542
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:546
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 82:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:550
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].expr}
		}
	case 83:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:555
		{
			yyVAL.bytes = nil
		}
	case 84:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:559
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 85:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:563
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:569
		{
			yyVAL.str = AST_JOIN
		}
	case 87:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:573
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 88:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:577
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:581
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 90:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:585
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 91:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:589
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:593
		{
			yyVAL.str = AST_JOIN
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:597
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 94:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:601
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 95:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:607
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:611
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:615
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:619
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 99:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:623
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 100:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:627
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:631
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 102:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:637
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 103:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:641
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 104:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:646
		{
			yyVAL.indexHints = nil
		}
	case 105:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:650
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 106:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:654
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 107:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:658
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 108:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:664
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 109:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:668
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 110:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:673
		{
			yyVAL.expr = nil
		}
	case 111:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:677
		{
			yyVAL.expr = yyDollar[2].expr
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
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:690
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 115:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:695
		{
			yyVAL.expr = nil
		}
	case 116:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:699
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 117:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:705
		{
			yyVAL.str = AST_ALL
		}
	case 118:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:709
		{
			yyVAL.str = AST_SOME
		}
	case 119:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:713
		{
			yyVAL.str = AST_ANY
		}
	case 120:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:719
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 121:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:723
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 122:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:729
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 123:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:733
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 124:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:739
		{
			yyVAL.tuple = ValTuple(yyDollar[2].exprs)
		}
	case 125:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:743
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 126:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:749
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 127:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:755
		{
			yyVAL.exprs = Exprs{yyDollar[1].expr}
		}
	case 128:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:759
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 129:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:765
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 130:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:769
		{
			yyVAL.expr = yyDollar[1].colName
		}
	case 131:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:773
		{
			yyVAL.expr = yyDollar[1].tuple
		}
	case 132:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:777
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITAND, Right: yyDollar[3].expr}
		}
	case 133:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:781
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITOR, Right: yyDollar[3].expr}
		}
	case 134:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:785
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITXOR, Right: yyDollar[3].expr}
		}
	case 135:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:789
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_PLUS, Right: yyDollar[3].expr}
		}
	case 136:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:793
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MINUS, Right: yyDollar[3].expr}
		}
	case 137:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:797
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MULT, Right: yyDollar[3].expr}
		}
	case 138:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:801
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_DIV, Right: yyDollar[3].expr}
		}
	case 139:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:805
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_IDIV, Right: yyDollar[3].expr}
		}
	case 140:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:809
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MOD, Right: yyDollar[3].expr}
		}
	case 141:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:813
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, Right: yyDollar[3].expr}
		}
	case 142:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:817
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 143:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:821
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, Right: yyDollar[3].expr}
		}
	case 144:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:825
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:829
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, Right: yyDollar[3].expr}
		}
	case 146:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:833
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:837
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, Right: yyDollar[3].expr}
		}
	case 148:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:841
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 149:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:845
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, Right: yyDollar[3].expr}
		}
	case 150:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:849
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 151:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:853
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, Right: yyDollar[3].expr}
		}
	case 152:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:857
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:861
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, Right: yyDollar[3].expr}
		}
	case 154:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:865
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:869
		{
			yyVAL.expr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:873
		{
			yyVAL.expr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:877
		{
			yyVAL.expr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 158:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:881
		{
			yyVAL.expr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 159:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:885
		{
			if num, ok := yyDollar[2].expr.(NumVal); ok {
				switch yyDollar[1].byt {
				case '-':
					yyVAL.expr = append(NumVal("-"), num...)
				case '+':
					yyVAL.expr = num
				default:
					yyVAL.expr = &UnaryExpr{Operator: yyDollar[1].byt, Expr: yyDollar[2].expr}
				}
			} else {
				yyVAL.expr = &UnaryExpr{Operator: yyDollar[1].byt, Expr: yyDollar[2].expr}
			}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:900
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 161:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:904
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:908
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LIKE, Right: yyDollar[3].expr}
		}
	case 163:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:912
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_LIKE, Right: yyDollar[4].expr}
		}
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:916
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 165:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:920
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 166:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:924
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: yyDollar[3].expr}
		}
	case 167:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:928
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: yyDollar[4].expr}
		}
	case 168:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:932
		{
			yyVAL.expr = &RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[3].expr}
		}
	case 169:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:936
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[4].expr}}
		}
	case 170:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:940
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 171:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:944
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_BETWEEN, Range: yyDollar[3].expr}
		}
	case 172:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:948
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_NOT_BETWEEN, Range: yyDollar[4].expr}
		}
	case 173:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:952
		{
			yyVAL.expr = yyDollar[1].caseExpr
		}
	case 174:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:956
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 175:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:960
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 176:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:964
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 177:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:968
		{
			yyVAL.expr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 178:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:972
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 179:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:976
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 180:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:980
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 181:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:984
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 182:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:988
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 183:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:992
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 184:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:996
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 185:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1000
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 186:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1004
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 187:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:1008
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1014
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1018
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1022
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1026
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1030
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1034
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1038
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 195:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1042
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 196:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1048
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1052
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 198:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1056
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 199:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1060
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1064
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1068
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1072
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1076
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1080
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 205:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1086
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1090
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1094
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1098
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 209:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1102
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1106
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 211:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1110
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1114
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1118
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1122
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1126
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1130
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1136
		{
			yyVAL.bytes = IF_BYTES
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1140
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1144
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1148
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1152
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1156
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1162
		{
			yyVAL.byt = AST_UPLUS
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1166
		{
			yyVAL.byt = AST_UMINUS
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1170
		{
			yyVAL.byt = AST_TILDA
		}
	case 226:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1176
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].expr, Whens: yyDollar[3].whens, Else: yyDollar[4].expr}
		}
	case 227:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1181
		{
			yyVAL.expr = nil
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1185
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1191
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 230:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1195
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 231:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1201
		{
			yyVAL.when = &When{Cond: yyDollar[2].expr, Val: yyDollar[4].expr}
		}
	case 232:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1206
		{
			yyVAL.expr = nil
		}
	case 233:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1210
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1216
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 235:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1220
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1226
		{
			yyVAL.expr = StrVal(yyDollar[1].bytes)
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1230
		{
			yyVAL.expr = NumVal(yyDollar[1].bytes)
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1234
		{
			yyVAL.expr = ValArg(yyDollar[1].bytes)
		}
	case 239:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1238
		{
			yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 240:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1242
		{
			yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 241:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1246
		{
			yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 242:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1250
		{
			yyVAL.expr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1254
		{
			yyVAL.expr = &NullVal{}
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1258
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1264
		{
			yyVAL.expr = &TrueVal{}
		}
	case 246:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1268
		{
			yyVAL.expr = &FalseVal{}
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1272
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 248:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1277
		{
			yyVAL.exprs = nil
		}
	case 249:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1281
		{
			yyVAL.exprs = yyDollar[3].exprs
		}
	case 250:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1286
		{
			yyVAL.expr = nil
		}
	case 251:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1290
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 252:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1295
		{
			yyVAL.orderBy = nil
		}
	case 253:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1299
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 254:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1305
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 255:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1309
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 256:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1315
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 257:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1320
		{
			yyVAL.str = AST_ASC
		}
	case 258:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1324
		{
			yyVAL.str = AST_ASC
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1328
		{
			yyVAL.str = AST_DESC
		}
	case 260:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1333
		{
			yyVAL.limit = nil
		}
	case 261:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1337
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].expr}
		}
	case 262:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1341
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].expr, Rowcount: yyDollar[4].expr}
		}
	case 263:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1345
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].expr, Rowcount: yyDollar[2].expr}
		}
	case 264:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1350
		{
			yyVAL.str = ""
		}
	case 265:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1354
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 266:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1358
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
	case 267:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1371
		{
			yyVAL.columns = nil
		}
	case 268:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1375
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 269:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1381
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 270:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1385
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 271:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1390
		{
			yyVAL.updateExprs = nil
		}
	case 272:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1394
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 273:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1400
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 274:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1404
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 275:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1410
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].expr}
		}
	case 276:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1415
		{
			yyVAL.empty = struct{}{}
		}
	case 277:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1417
		{
			yyVAL.empty = struct{}{}
		}
	case 278:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1420
		{
			yyVAL.empty = struct{}{}
		}
	case 279:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1422
		{
			yyVAL.empty = struct{}{}
		}
	case 280:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1425
		{
			yyVAL.empty = struct{}{}
		}
	case 281:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1427
		{
			yyVAL.empty = struct{}{}
		}
	case 282:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1431
		{
			yyVAL.empty = struct{}{}
		}
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1433
		{
			yyVAL.empty = struct{}{}
		}
	case 284:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1435
		{
			yyVAL.empty = struct{}{}
		}
	case 285:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1437
		{
			yyVAL.empty = struct{}{}
		}
	case 286:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1439
		{
			yyVAL.empty = struct{}{}
		}
	case 287:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1442
		{
			yyVAL.empty = struct{}{}
		}
	case 288:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1444
		{
			yyVAL.empty = struct{}{}
		}
	case 289:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1447
		{
			yyVAL.empty = struct{}{}
		}
	case 290:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1449
		{
			yyVAL.empty = struct{}{}
		}
	case 291:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1452
		{
			yyVAL.empty = struct{}{}
		}
	case 292:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1454
		{
			yyVAL.empty = struct{}{}
		}
	case 293:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1458
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 294:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1463
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
