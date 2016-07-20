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
const CHAR = 57412
const SIGNED = 57413
const UNSIGNED = 57414
const SQL_BIGINT = 57415
const SQL_VARCHAR = 57416
const SQL_DATE = 57417
const SQL_TIMESTAMP = 57418
const SQL_DOUBLE = 57419
const INTEGER = 57420
const FROM = 57421
const UNION = 57422
const MINUS = 57423
const EXCEPT = 57424
const INTERSECT = 57425
const COMMA = 57426
const JOIN = 57427
const STRAIGHT_JOIN = 57428
const LEFT = 57429
const RIGHT = 57430
const INNER = 57431
const OUTER = 57432
const CROSS = 57433
const NATURAL = 57434
const USE = 57435
const FORCE = 57436
const ON = 57437
const NOT = 57438
const OR = 57439
const XOR = 57440
const BETWEEN = 57441
const AND = 57442
const NE = 57443
const EQ = 57444
const NULL_SAFE_EQUAL = 57445
const IS = 57446
const LIKE = 57447
const REGEXP = 57448
const IN = 57449
const LT = 57450
const GT = 57451
const LE = 57452
const GE = 57453
const BIT_AND = 57454
const BIT_OR = 57455
const CARET = 57456
const PLUS = 57457
const SUB = 57458
const TIMES = 57459
const MOD = 57460
const DIV = 57461
const IDIV = 57462
const DOT = 57463
const UNARY = 57464
const CASE = 57465
const WHEN = 57466
const THEN = 57467
const ELSE = 57468
const END = 57469
const BEGIN = 57470
const COMMIT = 57471
const ROLLBACK = 57472
const NAMES = 57473
const REPLACE = 57474
const ADMIN = 57475
const SHOW = 57476
const DATABASES = 57477
const TABLES = 57478
const PROXY = 57479
const VARIABLES = 57480
const FULL = 57481
const SESSION = 57482
const GLOBAL = 57483
const COLUMNS = 57484
const CREATE = 57485
const ALTER = 57486
const DROP = 57487
const RENAME = 57488
const TABLE = 57489
const INDEX = 57490
const VIEW = 57491
const TO = 57492
const IGNORE = 57493
const IF = 57494
const UNIQUE = 57495
const USING = 57496

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
	142, 37,
	-2, 39,
	-1, 317,
	99, 0,
	-2, 171,
	-1, 389,
	99, 0,
	-2, 172,
}

const yyNprod = 293
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1226

var yyAct = [...]int{

	114, 413, 463, 503, 183, 351, 109, 102, 342, 405,
	128, 93, 103, 234, 130, 140, 327, 77, 239, 269,
	247, 108, 89, 75, 82, 18, 19, 20, 21, 344,
	44, 165, 46, 64, 358, 49, 47, 50, 430, 432,
	52, 53, 54, 288, 79, 168, 474, 84, 473, 472,
	86, 83, 85, 51, 90, 22, 253, 100, 96, 454,
	99, 231, 3, 80, 343, 78, 18, 378, 215, 56,
	58, 59, 105, 63, 61, 62, 435, 343, 251, 403,
	161, 254, 156, 262, 164, 191, 192, 193, 196, 194,
	195, 516, 172, 65, 181, 66, 431, 94, 272, 193,
	196, 194, 195, 271, 160, 97, 213, 272, 176, 177,
	514, 263, 271, 311, 158, 141, 142, 143, 167, 309,
	310, 308, 182, 313, 163, 406, 469, 218, 406, 354,
	171, 471, 470, 428, 230, 200, 201, 202, 203, 188,
	189, 190, 191, 192, 193, 196, 194, 195, 27, 28,
	29, 512, 30, 32, 31, 158, 79, 76, 427, 79,
	72, 426, 243, 23, 24, 26, 25, 174, 175, 513,
	511, 178, 98, 319, 184, 437, 241, 78, 267, 237,
	78, 436, 487, 216, 217, 340, 478, 258, 250, 252,
	249, 260, 314, 95, 244, 74, 424, 275, 458, 276,
	273, 425, 184, 274, 257, 179, 275, 233, 422, 477,
	512, 512, 274, 423, 307, 318, 243, 34, 35, 36,
	37, 315, 34, 35, 36, 37, 481, 186, 321, 323,
	324, 240, 154, 242, 240, 157, 79, 79, 338, 326,
	336, 337, 293, 295, 297, 299, 301, 303, 355, 476,
	186, 475, 438, 173, 264, 265, 350, 78, 349, 347,
	79, 278, 279, 280, 281, 282, 283, 284, 285, 286,
	287, 292, 294, 296, 298, 300, 302, 304, 305, 306,
	356, 78, 312, 360, 316, 317, 273, 364, 359, 185,
	186, 377, 186, 186, 186, 400, 346, 410, 367, 368,
	369, 370, 371, 379, 372, 373, 365, 399, 380, 158,
	381, 398, 382, 353, 383, 386, 384, 397, 385, 396,
	346, 480, 479, 497, 259, 391, 235, 496, 495, 129,
	393, 236, 88, 236, 219, 361, 362, 225, 411, 224,
	363, 188, 189, 190, 191, 192, 193, 196, 194, 195,
	402, 434, 408, 223, 222, 221, 412, 409, 212, 204,
	198, 197, 199, 210, 209, 211, 207, 200, 201, 202,
	203, 188, 189, 190, 191, 192, 193, 196, 194, 195,
	273, 273, 387, 388, 389, 420, 421, 91, 220, 101,
	153, 229, 228, 367, 368, 369, 370, 371, 227, 372,
	373, 226, 483, 155, 439, 440, 441, 442, 395, 394,
	392, 65, 79, 339, 460, 266, 404, 80, 433, 417,
	416, 256, 208, 205, 206, 212, 204, 198, 197, 199,
	210, 209, 211, 459, 200, 201, 202, 203, 188, 189,
	190, 191, 192, 193, 196, 194, 195, 376, 465, 186,
	186, 186, 445, 446, 267, 509, 267, 255, 238, 375,
	73, 169, 166, 162, 159, 484, 87, 457, 92, 18,
	245, 71, 491, 493, 170, 510, 455, 444, 447, 448,
	449, 450, 451, 452, 453, 38, 277, 461, 464, 69,
	501, 67, 414, 502, 345, 468, 504, 504, 504, 79,
	505, 506, 141, 142, 143, 40, 41, 42, 43, 415,
	390, 352, 467, 517, 419, 240, 55, 518, 515, 519,
	78, 498, 507, 18, 39, 57, 60, 261, 180, 482,
	17, 16, 15, 14, 13, 12, 246, 45, 357, 248,
	492, 184, 494, 141, 142, 143, 48, 322, 81, 348,
	112, 127, 508, 488, 136, 462, 466, 418, 499, 500,
	464, 106, 120, 121, 122, 401, 129, 320, 133, 123,
	126, 124, 125, 116, 117, 118, 144, 145, 146, 147,
	148, 149, 150, 151, 152, 232, 341, 113, 111, 443,
	115, 407, 107, 119, 290, 291, 141, 142, 143, 289,
	429, 270, 366, 112, 127, 268, 374, 136, 187, 68,
	33, 138, 137, 70, 80, 120, 121, 122, 11, 129,
	110, 133, 123, 126, 124, 125, 116, 117, 118, 144,
	145, 146, 147, 148, 149, 150, 151, 152, 10, 131,
	132, 104, 139, 9, 8, 7, 119, 134, 6, 5,
	4, 2, 1, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 138, 137, 0, 0, 0, 0,
	0, 0, 0, 110, 0, 0, 135, 0, 141, 142,
	143, 0, 0, 0, 0, 112, 127, 0, 0, 136,
	0, 0, 131, 132, 0, 139, 106, 120, 121, 122,
	134, 129, 325, 133, 123, 126, 124, 125, 116, 117,
	118, 144, 145, 146, 147, 148, 149, 150, 151, 152,
	0, 0, 0, 0, 0, 0, 0, 0, 119, 135,
	0, 141, 142, 143, 0, 0, 0, 0, 112, 127,
	0, 0, 136, 0, 0, 0, 138, 137, 0, 106,
	120, 121, 122, 0, 129, 110, 133, 123, 126, 124,
	125, 116, 117, 118, 144, 145, 146, 147, 148, 149,
	150, 151, 152, 0, 131, 132, 104, 139, 0, 0,
	0, 119, 134, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 138,
	137, 0, 0, 0, 0, 0, 0, 0, 110, 0,
	0, 135, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 18, 131, 132, 104,
	139, 0, 0, 0, 0, 134, 0, 0, 0, 0,
	141, 142, 143, 0, 0, 0, 0, 112, 127, 0,
	0, 136, 0, 0, 0, 0, 0, 0, 80, 120,
	121, 122, 0, 129, 135, 133, 123, 126, 124, 125,
	116, 117, 118, 144, 145, 146, 147, 148, 149, 150,
	151, 152, 0, 0, 0, 0, 0, 0, 0, 0,
	119, 0, 0, 141, 142, 143, 0, 0, 0, 0,
	112, 127, 0, 0, 136, 0, 0, 0, 138, 137,
	0, 80, 120, 121, 122, 0, 129, 110, 133, 123,
	126, 124, 125, 116, 117, 118, 144, 145, 146, 147,
	148, 149, 150, 151, 152, 0, 131, 132, 0, 139,
	0, 0, 0, 119, 134, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 138, 137, 489, 490, 0, 0, 0, 0, 0,
	110, 0, 0, 135, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 486, 0, 131,
	132, 0, 139, 0, 0, 0, 0, 134, 208, 205,
	206, 212, 204, 198, 197, 199, 210, 209, 211, 207,
	200, 201, 202, 203, 188, 189, 190, 191, 192, 193,
	196, 194, 195, 0, 0, 0, 135, 456, 0, 0,
	0, 208, 205, 206, 212, 204, 198, 197, 199, 210,
	209, 211, 207, 200, 201, 202, 203, 188, 189, 190,
	191, 192, 193, 196, 194, 195, 485, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 208, 205,
	206, 212, 204, 198, 197, 199, 210, 209, 211, 207,
	200, 201, 202, 203, 188, 189, 190, 191, 192, 193,
	196, 194, 195, 214, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 65, 208, 205, 206, 212,
	204, 198, 197, 199, 210, 209, 211, 207, 200, 201,
	202, 203, 188, 189, 190, 191, 192, 193, 196, 194,
	195, 204, 198, 197, 199, 210, 209, 211, 207, 200,
	201, 202, 203, 188, 189, 190, 191, 192, 193, 196,
	194, 195, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 208, 205, 206, 212, 204, 198,
	197, 199, 210, 209, 211, 207, 200, 201, 202, 203,
	188, 189, 190, 191, 192, 193, 196, 194, 195, 198,
	197, 199, 210, 209, 211, 207, 200, 201, 202, 203,
	188, 189, 190, 191, 192, 193, 196, 194, 195, 144,
	145, 146, 147, 148, 149, 150, 151, 152, 328, 329,
	330, 331, 332, 333, 334, 335,
}
var yyPact = [...]int{

	20, -1000, -1000, 137, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -117, -114, -94, -107, -1000, -1000, -1000,
	-1000, -66, 374, 518, 469, -1000, -1000, -1000, 466, -1000,
	440, 423, 116, 26, -128, -97, 374, -1000, -95, 374,
	-1000, 429, -130, 374, -130, 437, 88, -80, 93, 374,
	-85, -1000, -1000, -1000, 347, -1000, -1000, -1000, 712, -1000,
	349, 423, 368, -39, 423, 71, 427, -1000, 2, -1000,
	-41, 426, 28, 374, -1000, 425, -1000, -105, 424, 448,
	35, 374, 423, -1000, 874, 874, 88, 88, 874, 93,
	15, 874, 210, -1000, -1000, 1068, -53, -1000, -1000, -1000,
	874, 874, 292, -1000, 346, 313, 312, 311, 297, 295,
	-1000, -1000, -1000, 363, 360, 354, 353, -1000, -1000, 821,
	-1000, -1000, -1000, -1000, 874, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 291, 380, 421, 506, 380, -1000,
	874, 374, -1000, 444, -134, -1000, 44, -1000, 420, -1000,
	-1000, 384, -1000, 289, 1010, 1010, -1000, -1000, 1010, 88,
	4, 874, 874, 372, 1010, 70, 712, 462, 874, 874,
	874, 874, 874, 874, 874, 874, 874, 577, 577, 577,
	577, 577, 577, 577, 874, 874, 874, 287, 14, 874,
	96, 874, 874, -1000, 374, 56, 1010, -1000, -1000, 518,
	524, 712, 659, 1157, 1157, 712, -1000, -1000, -1000, -1000,
	370, 142, -60, 1010, 464, 380, 380, 225, -1000, 499,
	874, -1000, 1010, -1000, -1000, -1000, 34, 374, -1000, -116,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 464, 380,
	-1000, -1000, 874, 874, 1010, 326, -1000, 874, 222, 213,
	422, 61, -54, -1000, -1000, -1000, -1000, -1000, -30, -30,
	-30, -18, -18, -1000, -1000, -1000, -1000, 27, 292, -1000,
	-1000, -1000, 27, 292, 27, 292, 229, 292, 229, 292,
	229, 292, 229, 292, 1088, 259, 259, -1000, 287, 874,
	874, 874, 27, -1000, 483, -1000, 27, 1031, -1000, -1000,
	-1000, 367, 712, 366, 365, -1000, 235, 233, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 227, 223, 211, -1000,
	-1000, -47, -1000, 874, 33, 287, 137, 30, 254, -1000,
	499, 478, 496, 1010, 383, -1000, -1000, 382, -1000, -1000,
	71, 1010, 1010, 1010, 504, 70, 70, -1000, -1000, 123,
	111, 76, 73, 48, -55, -1000, 381, 308, 39, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 27, 27, 1031,
	-1000, -1000, -1000, 209, -1000, -1000, 712, 712, 712, 712,
	407, -68, -1000, 874, 902, -1000, 435, 114, -1000, -1000,
	-1000, 380, 478, -1000, 874, 874, -1000, -1000, 501, 482,
	213, 31, -1000, 47, -1000, 46, -1000, -1000, -1000, -1000,
	-99, -100, -102, -1000, -1000, -1000, -1000, -1000, -1000, 208,
	206, 166, 143, 279, -1000, -1000, -1000, 243, 148, -1000,
	-1000, -1000, -1000, -1000, -1000, 1010, 874, 369, 287, -1000,
	-1000, 972, 98, -1000, 935, -1000, 499, 874, 874, 874,
	-1000, -1000, 286, 285, 281, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 1010, 514, -1000, 874, 874, 874, -1000, -1000,
	-1000, 478, 1010, 94, 1010, 374, 374, 374, 380, 1010,
	1010, -1000, 439, 127, -1000, 126, 67, 71, -1000, 511,
	-16, -1000, 374, -1000, -1000, -1000, 374, -1000, 374, -1000,
}
var yyPgo = [...]int{

	0, 652, 651, 61, 650, 649, 648, 645, 644, 643,
	638, 618, 485, 613, 610, 43, 609, 7, 12, 608,
	606, 72, 605, 19, 602, 601, 160, 600, 3, 18,
	29, 592, 6, 10, 4, 591, 590, 15, 16, 589,
	14, 588, 21, 587, 586, 8, 585, 565, 557, 556,
	5, 555, 2, 553, 1, 552, 13, 549, 9, 23,
	17, 332, 548, 546, 539, 538, 537, 536, 0, 31,
	535, 534, 533, 532, 531, 530, 105, 11, 528, 527,
	526, 525, 524,
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
	21, 21, 21, 21, 21, 21, 38, 38, 38, 38,
	38, 38, 38, 38, 37, 37, 37, 37, 37, 37,
	37, 37, 37, 39, 39, 39, 39, 39, 39, 39,
	39, 39, 39, 39, 39, 36, 36, 36, 36, 36,
	36, 41, 41, 41, 43, 46, 46, 44, 44, 45,
	47, 47, 42, 42, 31, 31, 31, 31, 31, 31,
	31, 31, 31, 33, 33, 33, 48, 48, 49, 49,
	50, 50, 51, 51, 52, 53, 53, 53, 54, 54,
	54, 54, 55, 55, 55, 56, 56, 57, 57, 58,
	58, 59, 59, 60, 61, 61, 62, 62, 63, 63,
	64, 64, 64, 64, 64, 65, 65, 66, 66, 67,
	67, 68, 69,
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
	3, 6, 6, 6, 6, 6, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 2, 1, 2,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 5, 0, 1, 1, 2, 4,
	0, 2, 1, 3, 1, 1, 1, 2, 2, 2,
	2, 1, 1, 1, 1, 1, 0, 3, 0, 2,
	0, 3, 1, 3, 2, 0, 1, 1, 0, 2,
	4, 4, 0, 2, 4, 0, 3, 1, 3, 0,
	5, 1, 3, 3, 0, 2, 0, 3, 0, 1,
	1, 1, 1, 1, 1, 0, 1, 0, 1, 0,
	2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -70, -71, -72, -73, -74, -75, 5, 6,
	7, 8, 35, 143, 144, 146, 145, 128, 129, 130,
	132, 134, 133, -14, 80, 81, 82, 83, -12, -82,
	-12, -12, -12, -12, 147, -66, 149, 153, -63, 149,
	151, 147, 147, 148, 149, -12, 135, -81, 136, 137,
	-80, 140, 141, 139, -68, 37, -3, 22, -16, 23,
	-13, 31, -26, 37, 79, -59, 131, -60, -42, -68,
	37, -62, 152, 148, -68, 147, -68, 37, -61, 152,
	-68, -61, 31, -77, 9, 105, 138, -76, 79, -68,
	142, 42, -17, -18, 117, -21, 37, -31, -42, -32,
	96, -41, 26, -43, -68, -36, 49, 50, 51, 69,
	38, 39, 40, 45, 47, 48, 46, 27, -33, 42,
	-40, 115, 116, 44, 123, 152, 30, 88, 87, 118,
	-37, 19, 20, 21, 52, 53, 54, 55, 56, 57,
	58, 59, 60, 41, -26, 35, 121, -26, 84, 37,
	102, 121, 37, 96, -68, -69, 37, -69, 150, 37,
	26, 95, -68, -26, -21, -21, -77, -77, -21, -76,
	-78, 79, 107, -34, -21, 79, 84, -19, 112, 113,
	114, 115, 116, 117, 119, 120, 118, 102, 101, 103,
	108, 109, 110, 111, 100, 97, 98, 107, 96, 105,
	104, 106, 99, -68, 25, 121, -21, -21, -40, 42,
	42, 42, 42, 42, 42, 42, 38, 38, 38, 38,
	-34, -3, -46, -21, -56, 35, 42, -59, 37, -29,
	9, -60, -21, -68, -69, 26, -67, 154, -64, 146,
	144, 34, 145, 12, 37, 37, 37, -69, -56, 35,
	-77, -79, 79, 107, -21, -21, 43, 84, -22, -23,
	-25, 42, 37, -40, 142, 136, -18, 24, -21, -21,
	-21, -21, -21, -21, -21, -21, -21, -21, -15, 22,
	17, 18, -21, -15, -21, -15, -21, -15, -21, -15,
	-21, -15, -21, -15, -21, -21, -21, -32, 107, 105,
	106, 99, -21, 27, 96, -33, -21, -21, -68, 117,
	43, -17, 23, -17, -17, 43, -37, -38, 61, 62,
	63, 64, 65, 66, 67, 68, -37, -38, -18, 43,
	43, -44, -45, 124, -30, 30, -3, -59, -57, -42,
	-29, -50, 12, -21, 95, -68, -69, -65, 150, -30,
	-59, -21, -21, -21, -29, 84, -24, 85, 86, 87,
	88, 89, 91, 92, -20, 37, 25, -23, 121, -40,
	-40, -40, -40, -40, -40, -40, -32, -21, -21, -21,
	27, -33, 43, -17, 43, 43, 84, 84, 84, 84,
	84, -47, -45, 126, -21, -58, 95, -35, -32, -58,
	43, 84, -50, -54, 14, 13, 37, 37, -48, 10,
	-23, -23, 85, 90, 85, 90, 85, 85, 85, -27,
	93, 151, 94, 37, 43, 37, 142, 136, 43, -17,
	-17, -17, -17, -39, 70, 45, 46, 71, 72, 73,
	74, 75, 76, 77, 127, -21, 125, 32, 84, -42,
	-54, -21, -51, -52, -21, -69, -49, 11, 13, 95,
	85, 85, 148, 148, 148, 43, 43, 43, 43, 43,
	78, 78, -21, 33, -32, 84, 15, 84, -53, 28,
	29, -50, -21, -34, -21, 42, 42, 42, 7, -21,
	-21, -52, -54, -28, -68, -28, -28, -59, -55, 16,
	36, 43, 84, 43, 43, 7, 107, -68, -68, -68,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 57, 57,
	57, 57, 57, 287, 278, 0, 0, 28, 29, 30,
	57, -2, 0, 0, 61, 63, 64, 65, 66, 59,
	0, 0, 0, 0, 276, 0, 0, 288, 0, 0,
	279, 0, 274, 0, 274, 0, 112, 0, 115, 0,
	0, 40, 41, 38, 0, 291, 19, 62, 0, 67,
	58, 0, 0, 102, 0, 26, 0, 271, 0, 232,
	291, 0, 0, 0, 292, 0, 292, 0, 0, 0,
	0, 0, 0, 42, 0, 0, 112, 112, 0, 115,
	0, 0, 17, 68, 70, 74, 291, 129, 130, 131,
	0, 0, 0, 173, 232, 0, 178, 0, 0, 0,
	234, 235, 236, 0, 0, 0, 0, 241, 242, 0,
	125, 221, 222, 223, 225, 215, 216, 217, 218, 219,
	220, 243, 244, 245, 194, 195, 196, 197, 198, 199,
	200, 201, 202, 60, 265, 0, 0, 110, 0, 27,
	0, 0, 292, 0, 289, 49, 0, 52, 0, 54,
	275, 0, 292, 265, 113, 114, 43, 44, 116, 112,
	34, 0, 0, 0, 127, 0, 0, 71, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 75, 0, 0, 158, 159, 170, 0,
	0, 0, 0, 0, 0, 0, 237, 238, 239, 240,
	0, 0, 0, 226, 0, 0, 0, 110, 103, 250,
	0, 272, 273, 233, 47, 277, 0, 0, 292, 285,
	280, 281, 282, 283, 284, 53, 55, 56, 0, 0,
	45, 46, 0, 0, 32, 33, 31, 0, 110, 77,
	83, 0, 95, 97, 98, 99, 69, 72, 132, 133,
	134, 135, 136, 137, 138, 139, 140, 141, 0, 117,
	118, 119, 143, 0, 145, 0, 147, 0, 149, 0,
	151, 0, 153, 0, 155, 156, 157, 160, 0, 0,
	0, 0, 162, 164, 0, 166, 168, -2, 76, 73,
	174, 0, 0, 0, 0, 180, 0, 0, 186, 187,
	188, 189, 190, 191, 192, 193, 0, 0, 0, 124,
	126, 230, 227, 0, 269, 0, 121, 269, 0, 267,
	250, 258, 0, 111, 0, 290, 50, 0, 286, 22,
	23, 35, 36, 128, 246, 0, 0, 86, 87, 0,
	0, 0, 0, 0, 104, 84, 0, 0, 0, 142,
	144, 146, 148, 150, 152, 154, 161, 163, 169, -2,
	165, 167, 175, 0, 177, 179, 0, 0, 0, 0,
	0, 0, 228, 0, 0, 20, 0, 120, 122, 21,
	266, 0, 258, 25, 0, 0, 292, 51, 248, 0,
	78, 81, 88, 0, 90, 0, 92, 93, 94, 79,
	0, 0, 0, 85, 80, 96, 100, 101, 176, 0,
	0, 0, 0, 0, 203, 204, 205, 206, 208, 210,
	211, 212, 213, 214, 224, 231, 0, 0, 0, 268,
	24, 259, 251, 252, 255, 48, 250, 0, 0, 0,
	89, 91, 0, 0, 0, 181, 182, 183, 184, 185,
	207, 209, 229, 0, 123, 0, 0, 0, 254, 256,
	257, 258, 249, 247, 82, 0, 0, 0, 0, 260,
	261, 253, 262, 0, 108, 0, 0, 270, 18, 0,
	0, 105, 0, 106, 107, 263, 0, 109, 0, 264,
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
	152, 153, 154,
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
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 186:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1006
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 187:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1010
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1014
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1018
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1022
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1026
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1030
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1034
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1040
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 195:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1044
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 196:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1048
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1052
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 198:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1056
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 199:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1060
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1064
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1068
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1072
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1078
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1082
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 205:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1086
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1090
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1094
		{
			yyVAL.bytes = INTEGER_BYTES
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
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1110
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1114
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1118
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1122
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1128
		{
			yyVAL.bytes = IF_BYTES
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1132
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1136
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1140
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1144
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1148
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1154
		{
			yyVAL.byt = AST_UPLUS
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1158
		{
			yyVAL.byt = AST_UMINUS
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1162
		{
			yyVAL.byt = AST_TILDA
		}
	case 224:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1168
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].expr, Whens: yyDollar[3].whens, Else: yyDollar[4].expr}
		}
	case 225:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1173
		{
			yyVAL.expr = nil
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1177
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1183
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 228:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1187
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 229:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1193
		{
			yyVAL.when = &When{Cond: yyDollar[2].expr, Val: yyDollar[4].expr}
		}
	case 230:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1198
		{
			yyVAL.expr = nil
		}
	case 231:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1202
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1208
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 233:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1212
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1218
		{
			yyVAL.expr = StrVal(yyDollar[1].bytes)
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1222
		{
			yyVAL.expr = NumVal(yyDollar[1].bytes)
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1226
		{
			yyVAL.expr = ValArg(yyDollar[1].bytes)
		}
	case 237:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1230
		{
			yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 238:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1234
		{
			yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 239:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1238
		{
			yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 240:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1242
		{
			yyVAL.expr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1246
		{
			yyVAL.expr = &NullVal{}
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1250
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1256
		{
			yyVAL.expr = &TrueVal{}
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1260
		{
			yyVAL.expr = &FalseVal{}
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1264
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 246:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1269
		{
			yyVAL.exprs = nil
		}
	case 247:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1273
		{
			yyVAL.exprs = yyDollar[3].exprs
		}
	case 248:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1278
		{
			yyVAL.expr = nil
		}
	case 249:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1282
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 250:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1287
		{
			yyVAL.orderBy = nil
		}
	case 251:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1291
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1297
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 253:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1301
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 254:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1307
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 255:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1312
		{
			yyVAL.str = AST_ASC
		}
	case 256:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1316
		{
			yyVAL.str = AST_ASC
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1320
		{
			yyVAL.str = AST_DESC
		}
	case 258:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1325
		{
			yyVAL.limit = nil
		}
	case 259:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1329
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].expr}
		}
	case 260:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1333
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].expr, Rowcount: yyDollar[4].expr}
		}
	case 261:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1337
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].expr, Rowcount: yyDollar[2].expr}
		}
	case 262:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1342
		{
			yyVAL.str = ""
		}
	case 263:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1346
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 264:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1350
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
	case 265:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1363
		{
			yyVAL.columns = nil
		}
	case 266:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1367
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1373
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 268:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1377
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 269:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1382
		{
			yyVAL.updateExprs = nil
		}
	case 270:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1386
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 271:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1392
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 272:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1396
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 273:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1402
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].expr}
		}
	case 274:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1407
		{
			yyVAL.empty = struct{}{}
		}
	case 275:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1409
		{
			yyVAL.empty = struct{}{}
		}
	case 276:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1412
		{
			yyVAL.empty = struct{}{}
		}
	case 277:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1414
		{
			yyVAL.empty = struct{}{}
		}
	case 278:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1417
		{
			yyVAL.empty = struct{}{}
		}
	case 279:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1419
		{
			yyVAL.empty = struct{}{}
		}
	case 280:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1423
		{
			yyVAL.empty = struct{}{}
		}
	case 281:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1425
		{
			yyVAL.empty = struct{}{}
		}
	case 282:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1427
		{
			yyVAL.empty = struct{}{}
		}
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1429
		{
			yyVAL.empty = struct{}{}
		}
	case 284:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1431
		{
			yyVAL.empty = struct{}{}
		}
	case 285:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1434
		{
			yyVAL.empty = struct{}{}
		}
	case 286:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1436
		{
			yyVAL.empty = struct{}{}
		}
	case 287:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1439
		{
			yyVAL.empty = struct{}{}
		}
	case 288:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1441
		{
			yyVAL.empty = struct{}{}
		}
	case 289:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1444
		{
			yyVAL.empty = struct{}{}
		}
	case 290:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1446
		{
			yyVAL.empty = struct{}{}
		}
	case 291:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1450
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 292:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1455
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
