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
const CURRENT_TIMESTAMP = 57398
const TIMESTAMPADD = 57399
const TIMESTAMPDIFF = 57400
const YEAR = 57401
const QUARTER = 57402
const MONTH = 57403
const WEEK = 57404
const DAY = 57405
const HOUR = 57406
const MINUTE = 57407
const SECOND = 57408
const MICROSECOND = 57409
const SQL_TSI_YEAR = 57410
const SQL_TSI_QUARTER = 57411
const SQL_TSI_MONTH = 57412
const SQL_TSI_WEEK = 57413
const SQL_TSI_DAY = 57414
const SQL_TSI_HOUR = 57415
const SQL_TSI_MINUTE = 57416
const SQL_TSI_SECOND = 57417
const CONVERT = 57418
const CHAR = 57419
const SIGNED = 57420
const UNSIGNED = 57421
const SQL_BIGINT = 57422
const SQL_VARCHAR = 57423
const SQL_DATE = 57424
const SQL_TIMESTAMP = 57425
const SQL_DOUBLE = 57426
const INTEGER = 57427
const UNION = 57428
const MINUS = 57429
const EXCEPT = 57430
const INTERSECT = 57431
const JOIN = 57432
const STRAIGHT_JOIN = 57433
const LEFT = 57434
const RIGHT = 57435
const INNER = 57436
const OUTER = 57437
const CROSS = 57438
const NATURAL = 57439
const USE = 57440
const FORCE = 57441
const ON = 57442
const AND = 57443
const OR = 57444
const XOR = 57445
const NOT = 57446
const MOD = 57447
const DIV = 57448
const UNARY = 57449
const CASE = 57450
const WHEN = 57451
const THEN = 57452
const ELSE = 57453
const END = 57454
const BEGIN = 57455
const COMMIT = 57456
const ROLLBACK = 57457
const NAMES = 57458
const REPLACE = 57459
const ADMIN = 57460
const SHOW = 57461
const DATABASES = 57462
const TABLES = 57463
const PROXY = 57464
const VARIABLES = 57465
const FULL = 57466
const SESSION = 57467
const GLOBAL = 57468
const COLUMNS = 57469
const CREATE = 57470
const ALTER = 57471
const DROP = 57472
const RENAME = 57473
const TABLE = 57474
const INDEX = 57475
const VIEW = 57476
const TO = 57477
const IGNORE = 57478
const IF = 57479
const UNIQUE = 57480
const USING = 57481

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
	142, 37,
	-2, 39,
	-1, 227,
	95, 156,
	155, 156,
	-2, 75,
	-1, 495,
	107, 74,
	108, 74,
	109, 74,
	-2, 84,
}

const yyNprod = 293
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1325

var yyAct = [...]int{

	116, 409, 108, 504, 75, 107, 349, 462, 113, 224,
	168, 102, 133, 401, 278, 342, 340, 142, 114, 132,
	248, 225, 3, 243, 327, 77, 105, 34, 35, 36,
	37, 103, 479, 64, 365, 366, 367, 368, 369, 513,
	370, 371, 93, 513, 79, 513, 317, 84, 190, 190,
	86, 256, 78, 190, 90, 66, 190, 89, 82, 190,
	99, 356, 407, 190, 190, 190, 276, 276, 44, 49,
	46, 50, 171, 473, 47, 52, 53, 54, 472, 471,
	83, 426, 428, 431, 167, 281, 85, 51, 100, 96,
	452, 319, 175, 430, 376, 280, 18, 170, 178, 515,
	341, 182, 398, 514, 188, 512, 195, 80, 478, 477,
	454, 187, 341, 476, 227, 197, 475, 223, 228, 436,
	164, 177, 406, 390, 389, 387, 318, 275, 427, 262,
	159, 166, 65, 281, 161, 468, 235, 222, 226, 180,
	181, 402, 352, 280, 174, 402, 242, 56, 58, 59,
	470, 63, 61, 62, 469, 260, 490, 491, 263, 79,
	424, 161, 79, 246, 420, 252, 251, 78, 423, 421,
	78, 210, 211, 213, 214, 212, 253, 433, 97, 284,
	422, 276, 188, 432, 488, 283, 266, 250, 273, 274,
	18, 19, 20, 21, 418, 189, 76, 290, 252, 419,
	267, 292, 282, 249, 301, 302, 291, 306, 307, 308,
	309, 310, 311, 312, 313, 314, 315, 316, 296, 287,
	288, 289, 285, 305, 457, 22, 269, 284, 196, 395,
	192, 193, 194, 283, 208, 209, 210, 211, 213, 214,
	212, 321, 323, 324, 65, 79, 79, 249, 394, 345,
	326, 336, 393, 78, 347, 392, 391, 353, 337, 481,
	259, 261, 258, 480, 163, 344, 338, 348, 354, 79,
	144, 145, 146, 358, 359, 360, 351, 78, 183, 361,
	268, 190, 303, 357, 192, 193, 194, 244, 363, 344,
	88, 65, 245, 282, 498, 375, 487, 497, 362, 245,
	496, 381, 382, 383, 179, 236, 377, 378, 379, 192,
	193, 194, 234, 27, 28, 29, 380, 30, 32, 31,
	34, 35, 36, 37, 386, 233, 232, 231, 23, 24,
	26, 25, 161, 230, 388, 365, 366, 367, 368, 369,
	229, 370, 371, 72, 400, 91, 297, 399, 298, 299,
	101, 156, 240, 239, 374, 408, 397, 238, 237, 405,
	304, 80, 404, 199, 203, 201, 202, 429, 226, 300,
	373, 413, 412, 265, 264, 486, 282, 282, 416, 417,
	247, 73, 218, 219, 220, 221, 204, 435, 215, 216,
	217, 205, 206, 207, 208, 209, 210, 211, 213, 214,
	212, 453, 172, 437, 438, 439, 440, 169, 79, 165,
	459, 510, 162, 460, 87, 157, 458, 158, 160, 484,
	434, 456, 92, 464, 205, 206, 207, 208, 209, 210,
	211, 213, 214, 212, 18, 511, 176, 474, 463, 71,
	271, 517, 254, 173, 286, 200, 205, 206, 207, 208,
	209, 210, 211, 213, 214, 212, 69, 482, 483, 272,
	455, 144, 145, 146, 343, 67, 410, 467, 411, 350,
	188, 94, 492, 385, 495, 384, 485, 494, 466, 205,
	206, 207, 208, 209, 210, 211, 213, 214, 212, 500,
	501, 95, 415, 493, 503, 226, 502, 505, 505, 505,
	79, 506, 507, 185, 508, 38, 249, 98, 78, 144,
	145, 146, 74, 322, 518, 463, 123, 516, 519, 499,
	520, 131, 186, 18, 138, 40, 41, 42, 43, 39,
	57, 106, 124, 125, 126, 60, 55, 443, 444, 270,
	184, 111, 17, 16, 15, 136, 127, 130, 128, 129,
	118, 119, 120, 147, 148, 149, 150, 151, 152, 153,
	154, 155, 442, 445, 446, 447, 448, 449, 450, 451,
	121, 205, 206, 207, 208, 209, 210, 211, 213, 214,
	212, 14, 13, 12, 255, 45, 355, 140, 139, 257,
	48, 81, 346, 509, 489, 461, 465, 414, 396, 110,
	241, 339, 122, 134, 135, 104, 115, 441, 141, 117,
	403, 112, 143, 144, 145, 146, 198, 109, 425, 279,
	123, 364, 277, 372, 191, 131, 68, 33, 138, 70,
	11, 10, 9, 8, 7, 106, 124, 125, 126, 6,
	5, 137, 4, 2, 320, 111, 1, 0, 0, 136,
	127, 130, 128, 129, 118, 119, 120, 147, 148, 149,
	150, 151, 152, 153, 154, 155, 0, 0, 0, 0,
	0, 0, 0, 0, 121, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 140, 139, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 110, 0, 0, 0, 134, 135, 104,
	0, 0, 141, 0, 0, 0, 143, 144, 145, 146,
	0, 0, 0, 0, 123, 0, 0, 0, 0, 131,
	0, 0, 138, 0, 0, 0, 0, 0, 0, 106,
	124, 125, 126, 0, 0, 137, 0, 0, 325, 111,
	0, 0, 0, 136, 127, 130, 128, 129, 118, 119,
	120, 147, 148, 149, 150, 151, 152, 153, 154, 155,
	0, 0, 0, 0, 0, 0, 0, 0, 121, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 140, 139, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 110, 0, 0,
	0, 134, 135, 104, 0, 0, 141, 0, 0, 0,
	143, 295, 294, 144, 145, 146, 293, 0, 0, 0,
	0, 0, 0, 0, 0, 131, 0, 0, 138, 0,
	0, 0, 0, 0, 0, 80, 124, 125, 126, 137,
	0, 0, 0, 0, 0, 179, 0, 0, 0, 136,
	127, 130, 128, 129, 118, 119, 120, 147, 148, 149,
	150, 151, 152, 153, 154, 155, 18, 0, 0, 0,
	0, 0, 0, 0, 121, 0, 0, 0, 0, 0,
	0, 144, 145, 146, 0, 0, 0, 0, 123, 0,
	0, 140, 139, 131, 0, 0, 138, 0, 0, 0,
	0, 0, 0, 80, 124, 125, 126, 134, 135, 0,
	0, 0, 141, 111, 0, 0, 143, 136, 127, 130,
	128, 129, 118, 119, 120, 147, 148, 149, 150, 151,
	152, 153, 154, 155, 0, 0, 0, 0, 0, 0,
	0, 0, 121, 0, 0, 137, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 140,
	139, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 110, 0, 0, 0, 134, 135, 0, 0, 0,
	141, 0, 0, 0, 143, 144, 145, 146, 0, 0,
	0, 0, 123, 0, 0, 0, 0, 131, 0, 0,
	138, 0, 0, 0, 0, 0, 0, 80, 124, 125,
	126, 0, 0, 137, 0, 0, 0, 111, 0, 0,
	0, 136, 127, 130, 128, 129, 118, 119, 120, 147,
	148, 149, 150, 151, 152, 153, 154, 155, 0, 0,
	0, 0, 0, 0, 0, 0, 121, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 140, 139, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 18, 110, 0, 0, 0, 134,
	135, 0, 0, 0, 141, 0, 0, 0, 143, 144,
	145, 146, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 131, 0, 0, 138, 0, 0, 0, 0, 0,
	0, 80, 124, 125, 126, 0, 0, 137, 0, 0,
	0, 179, 0, 0, 0, 136, 127, 130, 128, 129,
	118, 119, 120, 147, 148, 149, 150, 151, 152, 153,
	154, 155, 0, 0, 0, 0, 0, 0, 0, 0,
	121, 0, 0, 0, 0, 0, 0, 144, 145, 146,
	0, 0, 0, 0, 0, 0, 0, 140, 139, 131,
	0, 0, 138, 0, 0, 0, 0, 0, 0, 80,
	124, 125, 126, 134, 135, 0, 0, 0, 141, 179,
	0, 0, 143, 136, 127, 130, 128, 129, 118, 119,
	120, 147, 148, 149, 150, 151, 152, 153, 154, 155,
	0, 0, 0, 0, 0, 0, 0, 0, 121, 0,
	0, 137, 199, 203, 201, 202, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 140, 139, 0, 0, 0,
	0, 218, 219, 220, 221, 204, 0, 215, 216, 217,
	0, 134, 135, 0, 0, 0, 141, 0, 0, 0,
	143, 147, 148, 149, 150, 151, 152, 153, 154, 155,
	328, 329, 330, 331, 332, 333, 334, 335, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 137,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 200, 205, 206, 207, 208, 209,
	210, 211, 213, 214, 212,
}
var yyPact = [...]int{

	185, -1000, -1000, 229, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -79, -80, -60, -72, -1000, -1000, -1000,
	-1000, 12, 249, 518, 442, -1000, -1000, -1000, 432, -1000,
	403, 339, 503, 65, -94, -68, 249, -1000, -61, 249,
	-1000, 372, -95, 249, -95, 386, 461, -49, 498, 249,
	-54, -1000, -1000, -1000, 298, -1000, -1000, -1000, 697, -1000,
	305, 339, 377, 9, 339, 66, 370, -1000, 211, -1000,
	-1, 367, 21, 249, -1000, 365, -1000, -78, 360, 416,
	38, 249, 339, -1000, 975, 1147, 461, 461, 1147, 498,
	494, 1147, 186, -1000, -1000, 202, -6, -1000, 1204, -1000,
	975, 871, -1000, -1000, -1000, 1147, 288, 281, 275, 274,
	273, 260, -1000, 253, -1000, -1000, -1000, 315, 314, 310,
	309, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, 1147, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 247, 319, 338,
	496, 319, -1000, 1147, 249, -1000, 415, -103, -1000, 116,
	-1000, 332, -1000, -1000, 331, -1000, 240, 177, 460, 1079,
	-1000, -1000, 460, 461, 431, 1147, 1147, -28, 460, 43,
	697, 419, 975, 975, 975, -1000, 249, 90, 803, 252,
	318, 1147, 1147, 250, 1147, 1147, 1147, 1147, 1147, 1147,
	1147, 1147, 1147, 1147, 1147, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -109, -29, -64, 177, 1204, -1000, 489,
	697, 593, 1207, 1207, 697, -1000, 518, -1000, -1000, -1000,
	-1000, -12, 460, 429, 319, 319, 237, -1000, 456, 975,
	-1000, 460, -1000, -1000, -1000, 36, 249, -1000, -89, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 429, 319, -1000,
	-1000, 1147, 1147, 460, 460, -1000, 1147, 193, 239, 328,
	91, -27, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 460, 253, 253, 253, -1000, 252, 1147, 1147,
	1147, 460, 368, -1000, 441, -1000, 460, 120, 120, 120,
	55, 55, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -30, 697, -31, -32, -1000, 161, 160, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 157, 153, 134, -24,
	-1000, 975, 35, 252, 229, 39, -33, -1000, 456, 451,
	454, 177, 330, -1000, -1000, 329, -1000, -1000, 66, 460,
	460, 460, 481, 43, 43, -1000, -1000, 98, 68, 84,
	72, 64, -23, -1000, 325, -62, 41, -1000, -1000, -1000,
	-1000, 460, 313, 460, 1147, -1000, -1000, -1000, -36, -1000,
	-1000, 697, 697, 697, 697, 480, -37, -1000, 1147, -15,
	335, -1000, 384, 129, -1000, -1000, -1000, 319, 451, -1000,
	1147, 975, -1000, -1000, 466, 453, 239, 29, -1000, 58,
	-1000, 54, -1000, -1000, -1000, -1000, -69, -70, -75, -1000,
	-1000, -1000, -1000, -1000, 1147, 460, -1000, -39, -42, -46,
	-47, -123, -1000, -1000, -1000, 173, 169, -1000, -1000, -1000,
	-1000, -1000, -1000, 460, 1147, 1147, 381, 252, -1000, -1000,
	280, 89, -1000, 123, -1000, 456, 975, 1147, 975, -1000,
	-1000, 248, 245, 242, 460, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 460, 460, 512, -1000, 1147, 1147, 975, -1000,
	-1000, -1000, 451, 177, 86, -1000, 249, 249, 249, 319,
	460, 460, -1000, 394, -50, -1000, -52, -56, 66, -1000,
	510, 413, -1000, 249, -1000, -1000, -1000, 249, -1000, 249,
	-1000,
}
var yyPgo = [...]int{

	0, 646, 643, 21, 642, 640, 639, 634, 633, 632,
	631, 630, 505, 629, 627, 626, 11, 31, 624, 623,
	26, 622, 14, 621, 619, 343, 618, 3, 20, 5,
	617, 616, 15, 611, 2, 18, 9, 19, 610, 609,
	17, 24, 607, 12, 606, 8, 602, 601, 16, 600,
	598, 597, 596, 6, 595, 7, 594, 1, 593, 23,
	592, 13, 4, 25, 290, 591, 590, 589, 586, 585,
	584, 0, 10, 583, 582, 581, 544, 543, 542, 178,
	42, 540, 539, 535, 530, 529,
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
	34, 34, 34, 34, 34, 41, 41, 41, 41, 41,
	41, 41, 41, 40, 40, 40, 40, 40, 40, 40,
	40, 40, 42, 42, 42, 42, 42, 42, 42, 42,
	42, 42, 42, 42, 39, 39, 39, 39, 39, 39,
	44, 44, 44, 46, 49, 49, 47, 47, 48, 48,
	50, 50, 45, 45, 33, 33, 33, 33, 33, 33,
	33, 33, 33, 37, 37, 37, 51, 51, 52, 52,
	53, 53, 54, 54, 55, 56, 56, 56, 57, 57,
	57, 57, 58, 58, 58, 59, 59, 60, 60, 61,
	61, 62, 62, 63, 64, 64, 65, 65, 66, 66,
	67, 67, 67, 67, 67, 68, 68, 69, 69, 70,
	70, 71, 72,
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
	3, 2, 3, 4, 5, 4, 1, 4, 3, 6,
	6, 6, 6, 6, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 2, 1, 2, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 5, 0, 1, 1, 2, 4, 4,
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
	-10, -11, -73, -74, -75, -76, -77, -78, 5, 6,
	7, 8, 40, 143, 144, 146, 145, 128, 129, 130,
	132, 134, 133, -14, 91, 92, 93, 94, -12, -85,
	-12, -12, -12, -12, 147, -69, 149, 153, -66, 149,
	151, 147, 147, 148, 149, -12, 135, -84, 136, 137,
	-83, 140, 141, 139, -71, 42, -3, 23, -15, 24,
	-13, 36, -25, 42, 9, -62, 131, -63, -45, -71,
	42, -65, 152, 148, -71, 147, -71, 42, -64, 152,
	-71, -64, 36, -80, 10, 30, 138, -79, 9, -71,
	142, 52, -16, -17, 116, -20, 42, -29, -34, -30,
	110, 52, -33, -45, -35, -44, -71, -39, 61, 62,
	63, 81, -46, 27, 43, 44, 45, 57, 59, 60,
	58, 32, -37, -43, 114, 115, 56, 152, 35, 99,
	98, 119, -40, 123, 20, 21, 22, 64, 65, 66,
	67, 68, 69, 70, 71, 72, 46, -25, 40, 121,
	-25, 95, 42, 53, 121, 42, 110, -71, -72, 42,
	-72, 150, 42, 27, 106, -71, -25, -20, -34, 52,
	-80, -80, -34, -79, -81, 9, 28, -36, -34, 9,
	95, -18, 107, 108, 109, -71, 26, 121, -31, 28,
	110, 30, 31, 29, 51, 111, 112, 113, 114, 115,
	116, 117, 120, 118, 119, 53, 54, 55, 47, 48,
	49, 50, -20, -29, -36, -3, -20, -34, -34, 52,
	52, 52, 52, 52, 52, -43, 52, 43, 43, 43,
	43, -49, -34, -59, 40, 52, -62, 42, -28, 10,
	-63, -34, -71, -72, 27, -70, 154, -67, 146, 144,
	39, 145, 13, 42, 42, 42, -72, -59, 40, -80,
	-82, 9, 28, -34, -34, 155, 95, -21, -22, -24,
	52, 42, -43, 142, 136, -17, 25, -20, -20, -20,
	-71, 116, -34, 23, 19, 18, -35, 28, 30, 31,
	51, -34, -34, 32, 110, -37, -34, -34, -34, -34,
	-34, -34, -34, -34, -34, -34, -34, 155, 155, 155,
	155, -16, 24, -16, -16, 155, -40, -41, 73, 74,
	75, 76, 77, 78, 79, 80, -40, -41, -17, -47,
	-48, 124, -32, 35, -3, -62, -60, -45, -28, -53,
	13, -20, 106, -71, -72, -68, 150, -32, -62, -34,
	-34, -34, -28, 95, -23, 96, 97, 98, 99, 100,
	102, 103, -19, 42, 26, -22, 121, -43, -43, -43,
	-35, -34, -34, -34, 107, 32, -37, 155, -16, 155,
	155, 95, 95, 95, 95, 95, -50, -48, 126, -29,
	-34, -61, 106, -38, -35, -61, 155, 95, -53, -57,
	15, 14, 42, 42, -51, 11, -22, -22, 96, 101,
	96, 101, 96, 96, 96, -26, 104, 151, 105, 42,
	155, 42, 142, 136, 107, -34, 155, -16, -16, -16,
	-16, -42, 82, 57, 58, 83, 84, 85, 86, 87,
	88, 89, 127, -34, 125, 125, 37, 95, -45, -57,
	-34, -54, -55, -20, -72, -52, 12, 14, 106, 96,
	96, 148, 148, 148, -34, 155, 155, 155, 155, 155,
	90, 90, -34, -34, 38, -35, 95, 16, 95, -56,
	33, 34, -53, -20, -36, -29, 52, 52, 52, 7,
	-34, -34, -55, -57, -27, -71, -27, -27, -62, -58,
	17, 41, 155, 95, 155, 155, 7, 28, -71, -71,
	-71,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 57, 57,
	57, 57, 57, 287, 278, 0, 0, 28, 29, 30,
	57, -2, 0, 0, 61, 63, 64, 65, 66, 59,
	0, 0, 0, 0, 276, 0, 0, 288, 0, 0,
	279, 0, 274, 0, 274, 0, 114, 0, 117, 0,
	0, 40, 41, 38, 0, 291, 19, 62, 0, 67,
	58, 0, 0, 104, 0, 26, 0, 271, 0, 232,
	291, 0, 0, 0, 292, 0, 292, 0, 0, 0,
	0, 0, 0, 42, 0, 0, 114, 114, 0, 117,
	0, 0, 17, 68, 70, 76, 291, 74, 75, 119,
	0, 0, 158, 159, 160, 0, 232, 0, 176, 0,
	0, 0, 184, 0, 234, 235, 236, 0, 0, 0,
	0, 241, 242, 154, 220, 221, 222, 214, 215, 216,
	217, 218, 219, 224, 243, 244, 245, 193, 194, 195,
	196, 197, 198, 199, 200, 201, 60, 265, 0, 0,
	112, 0, 27, 0, 0, 292, 0, 289, 49, 0,
	52, 0, 54, 275, 0, 292, 265, 115, 116, 0,
	43, 44, 118, 114, 34, 0, 0, 0, 156, 0,
	0, 71, 0, 0, 0, 77, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 142, 143, 144, 145, 146,
	147, 148, 123, 74, 0, 0, 0, -2, 171, 0,
	0, 0, 0, 0, 0, 139, 0, 237, 238, 239,
	240, 0, 225, 0, 0, 0, 112, 105, 250, 0,
	272, 273, 233, 47, 277, 0, 0, 292, 285, 280,
	281, 282, 283, 284, 53, 55, 56, 0, 0, 45,
	46, 0, 0, 32, 33, 31, 0, 112, 79, 85,
	0, 97, 99, 100, 101, 69, 72, 120, 121, 122,
	78, 73, 125, 0, 0, 0, 126, 0, 0, 0,
	0, 131, 0, 135, 0, 140, 137, 161, 162, 163,
	164, 165, 166, 167, 168, 169, 170, 124, 153, 155,
	172, 0, 0, 0, 0, 178, 0, 0, 185, 186,
	187, 188, 189, 190, 191, 192, 0, 0, 0, 230,
	226, 0, 269, 0, 150, 269, 0, 267, 250, 258,
	0, 113, 0, 290, 50, 0, 286, 22, 23, 35,
	36, 157, 246, 0, 0, 88, 89, 0, 0, 0,
	0, 0, 106, 86, 0, 0, 0, 127, 128, 129,
	130, 132, 0, 138, 0, 136, 141, 173, 0, 175,
	177, 0, 0, 0, 0, 0, 0, 227, 0, 74,
	75, 20, 0, 149, 151, 21, 266, 0, 258, 25,
	0, 0, 292, 51, 248, 0, 80, 83, 90, 0,
	92, 0, 94, 95, 96, 81, 0, 0, 0, 87,
	82, 98, 102, 103, 0, 133, 174, 0, 0, 0,
	0, 0, 202, 203, 204, 205, 207, 209, 210, 211,
	212, 213, 223, 231, 0, 0, 0, 0, 268, 24,
	259, 251, 252, 255, 48, 250, 0, 0, 0, 91,
	93, 0, 0, 0, 134, 179, 180, 181, 182, 183,
	206, 208, 228, 229, 0, 152, 0, 0, 0, 254,
	256, 257, 258, 249, 247, -2, 0, 0, 0, 0,
	260, 261, 253, 262, 0, 110, 0, 0, 270, 18,
	0, 0, 107, 0, 108, 109, 263, 0, 111, 0,
	264,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 118, 111, 3,
	52, 155, 116, 114, 95, 115, 121, 117, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	54, 53, 55, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 113, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 112, 3, 56,
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
	87, 88, 89, 90, 91, 92, 93, 94, 96, 97,
	98, 99, 100, 101, 102, 103, 104, 105, 106, 107,
	108, 109, 110, 119, 120, 122, 123, 124, 125, 126,
	127, 128, 129, 130, 131, 132, 133, 134, 135, 136,
	137, 138, 139, 140, 141, 142, 143, 144, 145, 146,
	147, 148, 149, 150, 151, 152, 153, 154,
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
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:966
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 177:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:970
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 178:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:974
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 179:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:978
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 180:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:982
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 181:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:986
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 182:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:990
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: StrVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 183:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:994
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 184:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:998
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
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
		//line sql.y:1038
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1042
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 195:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1046
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 196:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1050
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1054
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 198:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1058
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 199:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1062
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1066
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1070
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1076
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1080
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1084
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 205:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1088
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1092
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1096
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 208:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1100
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 209:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1104
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1108
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1112
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1116
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1120
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1126
		{
			yyVAL.bytes = IF_BYTES
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1130
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1134
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1138
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1142
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1146
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1152
		{
			yyVAL.byt = AST_UPLUS
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1156
		{
			yyVAL.byt = AST_UMINUS
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1160
		{
			yyVAL.byt = AST_TILDA
		}
	case 223:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1166
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 224:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1171
		{
			yyVAL.valExpr = nil
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1175
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1181
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 227:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1185
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 228:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1191
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 229:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1195
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 230:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1200
		{
			yyVAL.valExpr = nil
		}
	case 231:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1204
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1210
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 233:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1214
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1220
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1224
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1228
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 237:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1232
		{
			yyVAL.valExpr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 238:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1236
		{
			yyVAL.valExpr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 239:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1240
		{
			yyVAL.valExpr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 240:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1244
		{
			yyVAL.valExpr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1248
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1252
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1258
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1262
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1266
		{
			yyVAL.valExpr = &UnknownVal{}
		}
	case 246:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1271
		{
			yyVAL.valExprs = nil
		}
	case 247:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1275
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 248:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1280
		{
			yyVAL.expr = nil
		}
	case 249:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1284
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 250:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1289
		{
			yyVAL.orderBy = nil
		}
	case 251:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1293
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1299
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 253:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1303
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 254:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1309
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 255:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1314
		{
			yyVAL.str = AST_ASC
		}
	case 256:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1318
		{
			yyVAL.str = AST_ASC
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1322
		{
			yyVAL.str = AST_DESC
		}
	case 258:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1327
		{
			yyVAL.limit = nil
		}
	case 259:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1331
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 260:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1335
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 261:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1339
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].valExpr, Rowcount: yyDollar[2].valExpr}
		}
	case 262:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1344
		{
			yyVAL.str = ""
		}
	case 263:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1348
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 264:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1352
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
		//line sql.y:1365
		{
			yyVAL.columns = nil
		}
	case 266:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1369
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1375
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 268:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1379
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 269:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1384
		{
			yyVAL.updateExprs = nil
		}
	case 270:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1388
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 271:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1394
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 272:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1398
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 273:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1404
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 274:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1409
		{
			yyVAL.empty = struct{}{}
		}
	case 275:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1411
		{
			yyVAL.empty = struct{}{}
		}
	case 276:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1414
		{
			yyVAL.empty = struct{}{}
		}
	case 277:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1416
		{
			yyVAL.empty = struct{}{}
		}
	case 278:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1419
		{
			yyVAL.empty = struct{}{}
		}
	case 279:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1421
		{
			yyVAL.empty = struct{}{}
		}
	case 280:
		yyDollar = yyS[yypt-1 : yypt+1]
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
		//line sql.y:1429
		{
			yyVAL.empty = struct{}{}
		}
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1431
		{
			yyVAL.empty = struct{}{}
		}
	case 284:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1433
		{
			yyVAL.empty = struct{}{}
		}
	case 285:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1436
		{
			yyVAL.empty = struct{}{}
		}
	case 286:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1438
		{
			yyVAL.empty = struct{}{}
		}
	case 287:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1441
		{
			yyVAL.empty = struct{}{}
		}
	case 288:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1443
		{
			yyVAL.empty = struct{}{}
		}
	case 289:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1446
		{
			yyVAL.empty = struct{}{}
		}
	case 290:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1448
		{
			yyVAL.empty = struct{}{}
		}
	case 291:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1452
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 292:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1457
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
