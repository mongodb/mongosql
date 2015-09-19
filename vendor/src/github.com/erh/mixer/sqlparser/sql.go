//line sql.y:6
package sqlparser

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
	SHARE        = []byte("share")
	MODE         = []byte("mode")
	IF_BYTES     = []byte("if")
	VALUES_BYTES = []byte("values")
)

//line sql.y:31
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
const FOR = 57358
const ALL = 57359
const DISTINCT = 57360
const AS = 57361
const EXISTS = 57362
const IN = 57363
const IS = 57364
const LIKE = 57365
const BETWEEN = 57366
const NULL = 57367
const ASC = 57368
const DESC = 57369
const VALUES = 57370
const INTO = 57371
const DUPLICATE = 57372
const KEY = 57373
const DEFAULT = 57374
const SET = 57375
const LOCK = 57376
const ID = 57377
const STRING = 57378
const NUMBER = 57379
const VALUE_ARG = 57380
const COMMENT = 57381
const LE = 57382
const GE = 57383
const NE = 57384
const NULL_SAFE_EQUAL = 57385
const UNION = 57386
const MINUS = 57387
const EXCEPT = 57388
const INTERSECT = 57389
const JOIN = 57390
const STRAIGHT_JOIN = 57391
const LEFT = 57392
const RIGHT = 57393
const INNER = 57394
const OUTER = 57395
const CROSS = 57396
const NATURAL = 57397
const USE = 57398
const FORCE = 57399
const ON = 57400
const AND = 57401
const OR = 57402
const NOT = 57403
const UNARY = 57404
const CASE = 57405
const WHEN = 57406
const THEN = 57407
const ELSE = 57408
const END = 57409
const BEGIN = 57410
const COMMIT = 57411
const ROLLBACK = 57412
const NAMES = 57413
const REPLACE = 57414
const ADMIN = 57415
const SHOW = 57416
const DATABASES = 57417
const TABLES = 57418
const PROXY = 57419
const VARIABLES = 57420
const FULL = 57421
const COLUMNS = 57422
const CREATE = 57423
const ALTER = 57424
const DROP = 57425
const RENAME = 57426
const TABLE = 57427
const INDEX = 57428
const VIEW = 57429
const TO = 57430
const IGNORE = 57431
const IF = 57432
const UNIQUE = 57433
const USING = 57434

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
	"FOR",
	"ALL",
	"DISTINCT",
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
const yyMaxDepth = 200

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 228
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 638

var yyAct = [...]int{

	114, 332, 151, 401, 73, 105, 369, 193, 111, 141,
	284, 238, 324, 275, 277, 208, 75, 216, 410, 112,
	203, 91, 100, 410, 122, 101, 410, 194, 3, 168,
	169, 87, 80, 62, 300, 301, 302, 303, 304, 291,
	305, 306, 144, 380, 77, 163, 330, 82, 163, 379,
	84, 163, 76, 236, 88, 236, 44, 49, 46, 50,
	97, 64, 47, 34, 35, 36, 37, 349, 351, 106,
	378, 83, 267, 412, 121, 81, 51, 127, 411, 94,
	98, 409, 140, 360, 78, 118, 119, 120, 353, 354,
	148, 276, 311, 153, 143, 95, 152, 125, 150, 155,
	359, 329, 161, 319, 165, 167, 317, 160, 268, 350,
	235, 222, 195, 78, 191, 192, 196, 154, 137, 132,
	123, 124, 269, 52, 53, 54, 276, 128, 322, 257,
	220, 202, 77, 223, 70, 77, 206, 139, 212, 211,
	76, 134, 199, 76, 375, 356, 168, 169, 213, 355,
	63, 210, 325, 325, 126, 287, 161, 147, 226, 377,
	18, 233, 234, 241, 74, 376, 106, 246, 212, 86,
	227, 258, 240, 250, 248, 249, 255, 256, 229, 259,
	260, 261, 262, 263, 264, 265, 266, 242, 247, 245,
	241, 251, 347, 156, 346, 219, 221, 218, 343, 240,
	106, 106, 345, 344, 130, 77, 77, 133, 209, 280,
	162, 168, 169, 76, 282, 286, 134, 288, 236, 244,
	271, 273, 283, 243, 89, 149, 362, 289, 209, 77,
	386, 279, 364, 293, 294, 295, 341, 76, 228, 296,
	136, 342, 292, 181, 182, 183, 244, 204, 396, 205,
	243, 298, 310, 297, 163, 279, 313, 314, 205, 272,
	395, 117, 63, 394, 153, 242, 121, 200, 198, 127,
	197, 134, 312, 99, 129, 106, 104, 118, 119, 120,
	309, 166, 323, 78, 352, 109, 336, 335, 321, 125,
	383, 407, 225, 328, 331, 318, 308, 63, 327, 56,
	58, 59, 57, 61, 34, 35, 36, 37, 108, 408,
	339, 340, 123, 124, 102, 224, 207, 131, 358, 128,
	71, 145, 142, 242, 242, 361, 388, 389, 138, 135,
	85, 77, 363, 366, 90, 385, 367, 370, 18, 365,
	179, 180, 181, 182, 183, 371, 126, 69, 316, 270,
	176, 177, 178, 179, 180, 181, 182, 183, 414, 214,
	381, 278, 146, 92, 252, 382, 253, 254, 176, 177,
	178, 179, 180, 181, 182, 183, 93, 161, 231, 391,
	67, 393, 392, 390, 384, 65, 333, 374, 398, 370,
	232, 334, 400, 399, 285, 402, 402, 402, 77, 403,
	404, 158, 405, 18, 117, 373, 76, 338, 96, 121,
	209, 415, 127, 159, 72, 416, 413, 417, 117, 104,
	118, 119, 120, 121, 397, 18, 127, 38, 109, 39,
	60, 230, 125, 78, 118, 119, 120, 300, 301, 302,
	303, 304, 109, 305, 306, 157, 125, 40, 41, 42,
	43, 108, 17, 16, 15, 123, 124, 102, 55, 14,
	13, 12, 128, 117, 215, 108, 45, 290, 121, 123,
	124, 127, 217, 48, 18, 79, 128, 281, 78, 118,
	119, 120, 406, 387, 368, 372, 337, 109, 320, 126,
	201, 125, 274, 116, 121, 113, 115, 127, 18, 19,
	20, 21, 326, 126, 78, 118, 119, 120, 110, 170,
	108, 107, 348, 153, 123, 124, 239, 125, 299, 237,
	103, 128, 307, 164, 357, 66, 22, 176, 177, 178,
	179, 180, 181, 182, 183, 33, 68, 11, 10, 9,
	123, 124, 8, 7, 6, 5, 315, 128, 126, 176,
	177, 178, 179, 180, 181, 182, 183, 176, 177, 178,
	179, 180, 181, 182, 183, 4, 2, 1, 0, 0,
	0, 0, 0, 0, 126, 0, 27, 28, 29, 0,
	30, 32, 31, 171, 175, 173, 174, 0, 0, 23,
	24, 26, 25, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 187, 188, 189, 190, 0, 184, 185, 186,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 172,
	176, 177, 178, 179, 180, 181, 182, 183,
}
var yyPact = [...]int{

	493, -1000, -1000, 255, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -44, -45, -24, 23, -1000, -1000, -1000,
	-1000, 209, 227, 420, 368, -1000, -1000, -1000, 362, -1000,
	318, 285, 405, 78, -73, -26, 227, -1000, -29, 227,
	-1000, 295, -74, 227, -74, 305, 353, 353, 399, 227,
	-15, -1000, 229, -1000, -1000, -1000, 384, -1000, 235, 285,
	284, 43, 285, 163, 294, -1000, 195, -1000, 42, 293,
	70, 227, -1000, 287, -1000, -61, 286, 342, 93, 227,
	285, -1000, 443, 49, -1000, 353, 49, 399, 392, 49,
	201, -1000, -1000, 262, 29, 81, 562, -1000, 443, 398,
	-1000, -1000, -1000, 49, 226, 224, -1000, 223, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 49, -1000,
	214, 248, 281, 400, 248, -1000, 49, 227, -1000, 339,
	-90, -1000, 98, -1000, 280, -1000, -1000, 257, -1000, 205,
	81, 562, 489, 469, -1000, 489, 353, 369, 49, 49,
	2, 489, 128, 384, -1000, -1000, 227, 115, 443, 443,
	49, 220, 343, 49, 49, 104, 49, 49, 49, 49,
	49, 49, 49, 49, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -36, 0, 14, 562, -1000, 241, 384, -1000,
	420, 12, 489, 333, 248, 248, 218, -1000, 381, 443,
	-1000, 489, -1000, -1000, -1000, 91, 227, -1000, -64, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 333, 248, -1000,
	-1000, 49, 49, 489, 489, -1000, 49, 198, 383, 261,
	155, 16, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	489, -1000, 220, 49, 49, 489, 481, -1000, 323, 269,
	269, 269, 170, 170, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -2, 384, -5, 47, -1000, 443, 89, 220, 255,
	88, -7, -1000, 381, 371, 377, 81, 252, -1000, -1000,
	251, -1000, -1000, 163, 489, 489, 489, 396, 128, 128,
	-1000, -1000, 182, 144, 148, 140, 138, 5, -1000, 249,
	-20, 54, -1000, 489, 459, 49, -1000, -1000, -8, -1000,
	1, -1000, 49, 146, -1000, 302, 179, -1000, -1000, -1000,
	248, 371, -1000, 49, 49, -1000, -1000, 393, 373, 383,
	80, -1000, 111, -1000, 105, -1000, -1000, -1000, -1000, -31,
	-52, -58, -1000, -1000, -1000, -1000, -1000, 49, 489, -1000,
	-1000, 489, 49, 259, 220, -1000, -1000, 282, 177, -1000,
	300, -1000, 381, 443, 49, 443, -1000, -1000, 219, 216,
	204, 489, 489, 417, -1000, 49, 49, -1000, -1000, -1000,
	371, 81, 165, 81, 227, 227, 227, 248, 489, -1000,
	275, -27, -1000, -30, -35, 163, -1000, 409, 337, -1000,
	227, -1000, -1000, -1000, 227, -1000, 227, -1000,
}
var yyPgo = [...]int{

	0, 567, 566, 27, 565, 545, 544, 543, 542, 539,
	538, 537, 427, 536, 535, 525, 22, 25, 523, 522,
	520, 519, 11, 518, 516, 134, 512, 3, 15, 5,
	511, 509, 14, 508, 2, 19, 7, 502, 496, 24,
	495, 8, 493, 492, 13, 490, 488, 486, 485, 10,
	484, 6, 483, 1, 482, 20, 477, 12, 4, 16,
	169, 475, 473, 472, 467, 466, 464, 0, 9, 461,
	460, 459, 454, 453, 452, 95, 21, 445, 431, 430,
	429,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 72, 72, 5, 6, 7, 7, 69, 70,
	71, 74, 77, 77, 78, 78, 78, 79, 79, 73,
	73, 73, 73, 73, 8, 8, 8, 9, 9, 9,
	10, 11, 11, 11, 80, 12, 13, 13, 14, 14,
	14, 14, 14, 15, 15, 16, 16, 17, 17, 17,
	20, 20, 18, 18, 18, 21, 21, 22, 22, 22,
	22, 19, 19, 19, 23, 23, 23, 23, 23, 23,
	23, 23, 23, 24, 24, 24, 24, 24, 24, 24,
	25, 25, 26, 26, 26, 26, 27, 27, 28, 28,
	76, 76, 76, 75, 75, 29, 29, 29, 29, 29,
	30, 30, 30, 30, 30, 30, 30, 30, 30, 30,
	31, 31, 31, 31, 31, 31, 31, 32, 32, 37,
	37, 35, 35, 39, 36, 36, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 38, 38, 40, 40, 40, 42, 45,
	45, 43, 43, 44, 46, 46, 41, 41, 33, 33,
	33, 33, 47, 47, 48, 48, 49, 49, 50, 50,
	51, 52, 52, 52, 53, 53, 53, 54, 54, 54,
	55, 55, 56, 56, 57, 57, 58, 58, 59, 60,
	60, 61, 61, 62, 62, 63, 63, 63, 63, 63,
	64, 64, 65, 65, 66, 66, 67, 68,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 4, 12, 3,
	7, 7, 6, 6, 8, 7, 3, 4, 1, 1,
	1, 5, 2, 2, 0, 2, 2, 0, 1, 3,
	3, 4, 5, 5, 5, 8, 4, 6, 7, 4,
	5, 4, 5, 5, 0, 2, 0, 2, 1, 2,
	1, 1, 1, 0, 1, 1, 3, 1, 2, 3,
	1, 1, 0, 1, 2, 1, 3, 3, 3, 3,
	5, 0, 1, 2, 1, 1, 2, 3, 2, 3,
	2, 2, 2, 1, 3, 1, 1, 1, 3, 3,
	1, 3, 0, 5, 5, 5, 1, 3, 0, 2,
	0, 2, 2, 0, 2, 1, 3, 3, 2, 3,
	3, 3, 4, 3, 4, 5, 6, 3, 4, 2,
	1, 1, 1, 1, 1, 1, 1, 2, 1, 1,
	3, 3, 1, 3, 1, 3, 1, 1, 1, 3,
	3, 3, 3, 3, 3, 3, 3, 2, 3, 4,
	5, 4, 1, 1, 1, 1, 1, 1, 5, 0,
	1, 1, 2, 4, 0, 2, 1, 3, 1, 1,
	1, 1, 0, 3, 0, 2, 0, 3, 1, 3,
	2, 0, 1, 1, 0, 2, 4, 0, 2, 4,
	0, 3, 1, 3, 0, 5, 1, 3, 3, 0,
	2, 0, 3, 0, 1, 1, 1, 1, 1, 1,
	0, 1, 0, 1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -69, -70, -71, -72, -73, -74, 5, 6,
	7, 8, 33, 96, 97, 99, 98, 83, 84, 85,
	87, 89, 88, -14, 49, 50, 51, 52, -12, -80,
	-12, -12, -12, -12, 100, -65, 102, 106, -62, 102,
	104, 100, 100, 101, 102, -12, 90, 93, 91, 92,
	-79, 94, -67, 35, -3, 17, -15, 18, -13, 29,
	-25, 35, 9, -58, 86, -59, -41, -67, 35, -61,
	105, 101, -67, 100, -67, 35, -60, 105, -67, -60,
	29, -76, 10, 23, -76, -75, 9, -67, 95, 44,
	-16, -17, 73, -20, 35, -29, -34, -30, 67, 44,
	-33, -41, -35, -40, -67, -38, -42, 20, 36, 37,
	38, 25, -39, 71, 72, 48, 105, 28, 78, 39,
	-25, 33, 76, -25, 53, 35, 45, 76, 35, 67,
	-67, -68, 35, -68, 103, 35, 20, 64, -67, -25,
	-29, -34, -34, 44, -76, -34, -75, -77, 9, 21,
	-36, -34, 9, 53, -18, -67, 19, 76, 65, 66,
	-31, 21, 67, 23, 24, 22, 68, 69, 70, 71,
	72, 73, 74, 75, 45, 46, 47, 40, 41, 42,
	43, -29, -29, -36, -3, -34, -34, 44, 44, -39,
	44, -45, -34, -55, 33, 44, -58, 35, -28, 10,
	-59, -34, -67, -68, 20, -66, 107, -63, 99, 97,
	32, 98, 13, 35, 35, 35, -68, -55, 33, -76,
	-78, 9, 21, -34, -34, 108, 53, -21, -22, -24,
	44, 35, -39, 95, 91, -17, -67, 73, -29, -29,
	-34, -35, 21, 23, 24, -34, -34, 25, 67, -34,
	-34, -34, -34, -34, -34, -34, -34, 108, 108, 108,
	108, -16, 18, -16, -43, -44, 79, -32, 28, -3,
	-58, -56, -41, -28, -49, 13, -29, 64, -67, -68,
	-64, 103, -32, -58, -34, -34, -34, -28, 53, -23,
	54, 55, 56, 57, 58, 60, 61, -19, 35, 19,
	-22, 76, -35, -34, -34, 65, 25, 108, -16, 108,
	-46, -44, 81, -29, -57, 64, -37, -35, -57, 108,
	53, -49, -53, 15, 14, 35, 35, -47, 11, -22,
	-22, 54, 59, 54, 59, 54, 54, 54, -26, 62,
	104, 63, 35, 108, 35, 95, 91, 65, -34, 108,
	82, -34, 80, 30, 53, -41, -53, -34, -50, -51,
	-34, -68, -48, 12, 14, 64, 54, 54, 101, 101,
	101, -34, -34, 31, -35, 53, 53, -52, 26, 27,
	-49, -29, -36, -29, 44, 44, 44, 7, -34, -51,
	-53, -27, -67, -27, -27, -58, -54, 16, 34, 108,
	53, 108, 108, 7, 21, -67, -67, -67,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 222, 213, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 211, 0, 0, 223, 0, 0,
	214, 0, 209, 0, 209, 0, 110, 110, 113, 0,
	0, 38, 0, 226, 19, 59, 0, 64, 55, 0,
	0, 100, 0, 26, 0, 206, 0, 176, 226, 0,
	0, 0, 227, 0, 227, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 110, 0, 113, 0, 0,
	17, 65, 67, 72, 226, 70, 71, 115, 0, 0,
	146, 147, 148, 0, 176, 0, 162, 0, 178, 179,
	180, 181, 142, 165, 166, 167, 163, 164, 169, 57,
	200, 0, 0, 108, 0, 27, 0, 0, 227, 0,
	224, 46, 0, 49, 0, 51, 210, 0, 227, 200,
	111, 0, 112, 0, 41, 114, 110, 34, 0, 0,
	0, 144, 0, 0, 68, 73, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 130, 131, 132, 133, 134, 135,
	136, 118, 0, 0, 0, 144, 157, 0, 0, 129,
	0, 0, 170, 0, 0, 0, 108, 101, 186, 0,
	207, 208, 177, 44, 212, 0, 0, 227, 220, 215,
	216, 217, 218, 219, 50, 52, 53, 0, 0, 42,
	43, 0, 0, 32, 33, 31, 0, 108, 75, 81,
	0, 93, 95, 96, 97, 66, 74, 69, 116, 117,
	120, 121, 0, 0, 0, 123, 0, 127, 0, 149,
	150, 151, 152, 153, 154, 155, 156, 119, 141, 143,
	158, 0, 0, 0, 174, 171, 0, 204, 0, 138,
	204, 0, 202, 186, 194, 0, 109, 0, 225, 47,
	0, 221, 22, 23, 35, 36, 145, 182, 0, 0,
	84, 85, 0, 0, 0, 0, 0, 102, 82, 0,
	0, 0, 122, 124, 0, 0, 128, 159, 0, 161,
	0, 172, 0, 0, 20, 0, 137, 139, 21, 201,
	0, 194, 25, 0, 0, 227, 48, 184, 0, 76,
	79, 86, 0, 88, 0, 90, 91, 92, 77, 0,
	0, 0, 83, 78, 94, 98, 99, 0, 125, 160,
	168, 175, 0, 0, 0, 203, 24, 195, 187, 188,
	191, 45, 186, 0, 0, 0, 87, 89, 0, 0,
	0, 126, 173, 0, 140, 0, 0, 190, 192, 193,
	194, 185, 183, 80, 0, 0, 0, 0, 196, 189,
	197, 0, 106, 0, 0, 205, 18, 0, 0, 103,
	0, 104, 105, 198, 0, 107, 0, 199,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 75, 68, 3,
	44, 108, 73, 71, 53, 72, 76, 74, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	46, 45, 47, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 70, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 69, 3, 48,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 49, 50, 51, 52, 54, 55, 56, 57,
	58, 59, 60, 61, 62, 63, 64, 65, 66, 67,
	77, 78, 79, 80, 81, 82, 83, 84, 85, 86,
	87, 88, 89, 90, 91, 92, 93, 94, 95, 96,
	97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
	107,
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
		//line sql.y:171
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:177
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:197
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 18:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:201
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].boolExpr), GroupBy: GroupBy(yyDollar[8].valExprs), Having: NewWhere(AST_HAVING, yyDollar[9].boolExpr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:205
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 20:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:212
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:216
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
		//line sql.y:228
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:232
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
		//line sql.y:245
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].boolExpr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 25:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:251
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].boolExpr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:257
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:261
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:267
		{
			yyVAL.statement = &Begin{}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:273
		{
			yyVAL.statement = &Commit{}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:279
		{
			yyVAL.statement = &Rollback{}
		}
	case 31:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:285
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].valExprs}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:291
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:295
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 34:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:300
		{
			yyVAL.valExpr = nil
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:304
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:308
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 37:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:313
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:317
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 39:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:324
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:328
		{
			yyVAL.statement = &Show{Section: "variables", LikeOrWhere: yyDollar[3].expr}
		}
	case 41:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:332
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].valExpr, LikeOrWhere: yyDollar[4].expr}
		}
	case 42:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:336
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].valExpr, LikeOrWhere: yyDollar[5].expr}
		}
	case 43:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:340
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].valExpr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].valExpr}
		}
	case 44:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:346
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 45:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:350
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:355
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 47:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:361
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 48:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:365
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 49:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:370
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 50:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:376
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 51:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:382
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 52:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:386
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 53:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:391
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 54:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:396
		{
			SetAllowComments(yylex, true)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:400
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 56:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:406
		{
			yyVAL.bytes2 = nil
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:410
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:416
		{
			yyVAL.str = AST_UNION
		}
	case 59:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:420
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:424
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:428
		{
			yyVAL.str = AST_EXCEPT
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:432
		{
			yyVAL.str = AST_INTERSECT
		}
	case 63:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:437
		{
			yyVAL.str = ""
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:441
		{
			yyVAL.str = AST_DISTINCT
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:447
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:451
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:457
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:461
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:465
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 70:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:471
		{
			yyVAL.expr = yyDollar[1].boolExpr
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:475
		{
			yyVAL.expr = yyDollar[1].valExpr
		}
	case 72:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:480
		{
			yyVAL.bytes = nil
		}
	case 73:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:484
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 74:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:488
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 75:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:494
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 76:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:498
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 77:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:504
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:508
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:512
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 80:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:516
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].boolExpr}
		}
	case 81:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:521
		{
			yyVAL.bytes = nil
		}
	case 82:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:525
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 83:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:529
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 84:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:535
		{
			yyVAL.str = AST_JOIN
		}
	case 85:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:539
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 86:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:543
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 87:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:547
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 88:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:551
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:555
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 90:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:559
		{
			yyVAL.str = AST_JOIN
		}
	case 91:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:563
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:567
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 93:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:573
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 94:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:577
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 95:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:581
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 96:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:585
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:589
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 98:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:593
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:597
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 100:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:603
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:607
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 102:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:612
		{
			yyVAL.indexHints = nil
		}
	case 103:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:616
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 104:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:620
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 105:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:624
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 106:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:630
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 107:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:634
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 108:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:639
		{
			yyVAL.boolExpr = nil
		}
	case 109:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:643
		{
			yyVAL.boolExpr = yyDollar[2].boolExpr
		}
	case 110:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:648
		{
			yyVAL.expr = nil
		}
	case 111:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:652
		{
			yyVAL.expr = yyDollar[2].boolExpr
		}
	case 112:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:656
		{
			yyVAL.expr = yyDollar[2].valExpr
		}
	case 113:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:661
		{
			yyVAL.valExpr = nil
		}
	case 114:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:665
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 116:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:672
		{
			yyVAL.boolExpr = &AndExpr{Left: yyDollar[1].boolExpr, Right: yyDollar[3].boolExpr}
		}
	case 117:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:676
		{
			yyVAL.boolExpr = &OrExpr{Left: yyDollar[1].boolExpr, Right: yyDollar[3].boolExpr}
		}
	case 118:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:680
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyDollar[2].boolExpr}
		}
	case 119:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:684
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyDollar[2].boolExpr}
		}
	case 120:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:690
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, Right: yyDollar[3].valExpr}
		}
	case 121:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:694
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 122:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:698
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 123:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:702
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_LIKE, Right: yyDollar[3].valExpr}
		}
	case 124:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:706
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_LIKE, Right: yyDollar[4].valExpr}
		}
	case 125:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:710
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_BETWEEN, From: yyDollar[3].valExpr, To: yyDollar[5].valExpr}
		}
	case 126:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:714
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].valExpr, To: yyDollar[6].valExpr}
		}
	case 127:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:718
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NULL, Expr: yyDollar[1].valExpr}
		}
	case 128:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:722
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NOT_NULL, Expr: yyDollar[1].valExpr}
		}
	case 129:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:726
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 130:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:732
		{
			yyVAL.str = AST_EQ
		}
	case 131:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:736
		{
			yyVAL.str = AST_LT
		}
	case 132:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:740
		{
			yyVAL.str = AST_GT
		}
	case 133:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:744
		{
			yyVAL.str = AST_LE
		}
	case 134:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:748
		{
			yyVAL.str = AST_GE
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:752
		{
			yyVAL.str = AST_NE
		}
	case 136:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:756
		{
			yyVAL.str = AST_NSE
		}
	case 137:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:762
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 138:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:766
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:772
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 140:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:776
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 141:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:782
		{
			yyVAL.tuple = ValTuple(yyDollar[2].valExprs)
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:786
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 143:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:792
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 144:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:798
		{
			yyVAL.valExprs = ValExprs{yyDollar[1].valExpr}
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:802
		{
			yyVAL.valExprs = append(yyDollar[1].valExprs, yyDollar[3].valExpr)
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:808
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 147:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:812
		{
			yyVAL.valExpr = yyDollar[1].colName
		}
	case 148:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:816
		{
			yyVAL.valExpr = yyDollar[1].tuple
		}
	case 149:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:820
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITAND, Right: yyDollar[3].valExpr}
		}
	case 150:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:824
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITOR, Right: yyDollar[3].valExpr}
		}
	case 151:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:828
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITXOR, Right: yyDollar[3].valExpr}
		}
	case 152:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:832
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_PLUS, Right: yyDollar[3].valExpr}
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:836
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MINUS, Right: yyDollar[3].valExpr}
		}
	case 154:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:840
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MULT, Right: yyDollar[3].valExpr}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:844
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_DIV, Right: yyDollar[3].valExpr}
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:848
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 157:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:852
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
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:867
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes}
		}
	case 159:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:871
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 160:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:875
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 161:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:879
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 162:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:883
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 163:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:889
		{
			yyVAL.bytes = IF_BYTES
		}
	case 164:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:893
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 165:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:899
		{
			yyVAL.byt = AST_UPLUS
		}
	case 166:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:903
		{
			yyVAL.byt = AST_UMINUS
		}
	case 167:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:907
		{
			yyVAL.byt = AST_TILDA
		}
	case 168:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:913
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 169:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:918
		{
			yyVAL.valExpr = nil
		}
	case 170:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:922
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 171:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:928
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 172:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:932
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 173:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:938
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 174:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:943
		{
			yyVAL.valExpr = nil
		}
	case 175:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:947
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 176:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:953
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 177:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:957
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 178:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:963
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 179:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:967
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 180:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:971
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 181:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:975
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 182:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:980
		{
			yyVAL.valExprs = nil
		}
	case 183:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:984
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 184:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:989
		{
			yyVAL.boolExpr = nil
		}
	case 185:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:993
		{
			yyVAL.boolExpr = yyDollar[2].boolExpr
		}
	case 186:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:998
		{
			yyVAL.orderBy = nil
		}
	case 187:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1002
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1008
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 189:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1012
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 190:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1018
		{
			yyVAL.order = &Order{Expr: yyDollar[1].valExpr, Direction: yyDollar[2].str}
		}
	case 191:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.str = AST_ASC
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1027
		{
			yyVAL.str = AST_ASC
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1031
		{
			yyVAL.str = AST_DESC
		}
	case 194:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1036
		{
			yyVAL.limit = nil
		}
	case 195:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1040
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 196:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1044
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 197:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1049
		{
			yyVAL.str = ""
		}
	case 198:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1053
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 199:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1057
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
	case 200:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1070
		{
			yyVAL.columns = nil
		}
	case 201:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1074
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1080
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 203:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1084
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 204:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1089
		{
			yyVAL.updateExprs = nil
		}
	case 205:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1093
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1099
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 207:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1103
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 208:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1109
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 209:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1114
		{
			yyVAL.empty = struct{}{}
		}
	case 210:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1116
		{
			yyVAL.empty = struct{}{}
		}
	case 211:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1119
		{
			yyVAL.empty = struct{}{}
		}
	case 212:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1121
		{
			yyVAL.empty = struct{}{}
		}
	case 213:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1124
		{
			yyVAL.empty = struct{}{}
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1126
		{
			yyVAL.empty = struct{}{}
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1130
		{
			yyVAL.empty = struct{}{}
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1132
		{
			yyVAL.empty = struct{}{}
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1134
		{
			yyVAL.empty = struct{}{}
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1136
		{
			yyVAL.empty = struct{}{}
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1138
		{
			yyVAL.empty = struct{}{}
		}
	case 220:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1141
		{
			yyVAL.empty = struct{}{}
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1143
		{
			yyVAL.empty = struct{}{}
		}
	case 222:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1146
		{
			yyVAL.empty = struct{}{}
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1148
		{
			yyVAL.empty = struct{}{}
		}
	case 224:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1151
		{
			yyVAL.empty = struct{}{}
		}
	case 225:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1153
		{
			yyVAL.empty = struct{}{}
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1157
		{
			yyVAL.bytes = bytes.ToLower(yyDollar[1].bytes)
		}
	case 227:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1162
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
