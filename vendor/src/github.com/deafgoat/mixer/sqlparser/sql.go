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
const DATE = 57386
const DATETIME = 57387
const TIME = 57388
const TIMESTAMP = 57389
const YEAR = 57390
const UNION = 57391
const MINUS = 57392
const EXCEPT = 57393
const INTERSECT = 57394
const JOIN = 57395
const STRAIGHT_JOIN = 57396
const LEFT = 57397
const RIGHT = 57398
const INNER = 57399
const OUTER = 57400
const CROSS = 57401
const NATURAL = 57402
const USE = 57403
const FORCE = 57404
const ON = 57405
const AND = 57406
const OR = 57407
const NOT = 57408
const UNARY = 57409
const CASE = 57410
const WHEN = 57411
const THEN = 57412
const ELSE = 57413
const END = 57414
const BEGIN = 57415
const COMMIT = 57416
const ROLLBACK = 57417
const NAMES = 57418
const REPLACE = 57419
const ADMIN = 57420
const SHOW = 57421
const DATABASES = 57422
const TABLES = 57423
const PROXY = 57424
const VARIABLES = 57425
const FULL = 57426
const COLUMNS = 57427
const CREATE = 57428
const ALTER = 57429
const DROP = 57430
const RENAME = 57431
const TABLE = 57432
const INDEX = 57433
const VIEW = 57434
const TO = 57435
const IGNORE = 57436
const IF = 57437
const UNIQUE = 57438
const USING = 57439

var yyToknames = []string{
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
	"DATE",
	"DATETIME",
	"TIME",
	"TIMESTAMP",
	"YEAR",
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
}
var yyStatenames = []string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line yacctab:1
var yyExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 196,
	70, 71,
	71, 71,
	-2, 144,
	-1, 403,
	70, 70,
	71, 70,
	-2, 80,
}

const yyNprod = 235
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 642

var yyAct = []int{

	114, 342, 106, 411, 73, 103, 379, 105, 111, 192,
	141, 122, 293, 245, 334, 112, 284, 286, 215, 75,
	100, 117, 193, 3, 101, 420, 121, 210, 420, 127,
	275, 223, 87, 62, 91, 80, 104, 118, 119, 120,
	300, 49, 420, 50, 77, 109, 162, 82, 340, 125,
	84, 162, 76, 162, 88, 243, 64, 44, 243, 46,
	97, 144, 390, 47, 309, 310, 311, 312, 313, 243,
	314, 315, 389, 108, 52, 53, 54, 123, 124, 102,
	422, 388, 140, 421, 128, 34, 35, 36, 37, 81,
	148, 83, 94, 370, 51, 143, 151, 419, 150, 154,
	78, 369, 160, 339, 166, 95, 329, 18, 327, 159,
	326, 126, 196, 276, 190, 195, 203, 191, 363, 98,
	359, 361, 372, 285, 242, 332, 285, 364, 320, 206,
	153, 209, 77, 168, 139, 77, 213, 248, 219, 218,
	76, 137, 132, 76, 277, 385, 247, 134, 265, 220,
	180, 181, 182, 335, 217, 160, 74, 63, 335, 233,
	240, 241, 360, 56, 58, 59, 57, 61, 255, 219,
	253, 254, 257, 249, 296, 263, 264, 234, 267, 268,
	269, 270, 271, 272, 273, 274, 258, 252, 366, 281,
	236, 117, 365, 147, 387, 266, 121, 160, 251, 127,
	256, 70, 250, 155, 278, 248, 104, 118, 119, 120,
	164, 165, 77, 77, 247, 109, 289, 386, 353, 125,
	76, 291, 295, 354, 297, 280, 282, 18, 19, 20,
	21, 357, 292, 288, 167, 298, 77, 356, 355, 161,
	302, 303, 304, 108, 76, 134, 305, 123, 124, 102,
	63, 351, 301, 216, 128, 22, 352, 288, 216, 249,
	243, 319, 160, 306, 322, 323, 251, 396, 229, 374,
	250, 130, 235, 86, 133, 406, 321, 34, 35, 36,
	37, 126, 136, 212, 279, 164, 165, 227, 162, 405,
	230, 195, 149, 333, 178, 179, 180, 181, 182, 129,
	331, 307, 328, 337, 338, 341, 134, 211, 404, 259,
	27, 28, 29, 207, 30, 32, 31, 205, 212, 249,
	249, 349, 350, 23, 24, 26, 25, 368, 89, 398,
	399, 63, 204, 99, 318, 371, 417, 78, 373, 362,
	346, 77, 345, 376, 232, 395, 377, 380, 231, 375,
	317, 214, 71, 145, 418, 131, 381, 226, 228, 225,
	175, 176, 177, 178, 179, 180, 181, 182, 142, 138,
	391, 135, 85, 393, 18, 392, 175, 176, 177, 178,
	179, 180, 181, 182, 90, 69, 325, 160, 92, 401,
	394, 195, 238, 403, 402, 400, 221, 287, 408, 380,
	18, 93, 410, 409, 239, 412, 412, 412, 77, 413,
	414, 260, 415, 261, 262, 117, 76, 157, 424, 146,
	121, 425, 67, 127, 65, 426, 343, 427, 384, 158,
	78, 118, 119, 120, 309, 310, 311, 312, 313, 109,
	314, 315, 344, 125, 198, 201, 199, 200, 202, 367,
	18, 294, 175, 176, 177, 178, 179, 180, 181, 182,
	383, 348, 216, 96, 72, 423, 407, 108, 18, 39,
	121, 123, 124, 127, 60, 237, 156, 17, 128, 16,
	78, 118, 119, 120, 15, 14, 13, 12, 197, 152,
	194, 222, 45, 125, 198, 201, 199, 200, 202, 299,
	18, 117, 224, 48, 79, 126, 121, 290, 416, 127,
	397, 378, 382, 347, 330, 208, 78, 118, 119, 120,
	121, 123, 124, 127, 283, 109, 116, 113, 128, 125,
	78, 118, 119, 120, 115, 336, 110, 121, 169, 152,
	127, 107, 358, 125, 246, 38, 308, 78, 118, 119,
	120, 244, 316, 108, 163, 126, 152, 123, 124, 66,
	125, 33, 68, 11, 128, 40, 41, 42, 43, 10,
	9, 123, 124, 8, 7, 6, 55, 5, 128, 4,
	2, 1, 170, 174, 172, 173, 0, 0, 123, 124,
	0, 126, 0, 0, 0, 128, 0, 0, 0, 0,
	0, 186, 187, 188, 189, 126, 183, 184, 185, 324,
	0, 0, 175, 176, 177, 178, 179, 180, 181, 182,
	0, 0, 126, 175, 176, 177, 178, 179, 180, 181,
	182, 0, 0, 171, 175, 176, 177, 178, 179, 180,
	181, 182,
}
var yyPact = []int{

	222, -1000, -1000, 223, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -48, -66, -11, -31, -1000, -1000, -1000,
	-1000, 68, 296, 463, 407, -1000, -1000, -1000, 404, -1000,
	356, 317, 455, 65, -75, -17, 296, -1000, -14, 296,
	-1000, 337, -78, 296, -78, 355, 378, 378, 454, 296,
	19, -1000, 289, -1000, -1000, -1000, 1, -1000, 260, 317,
	322, 61, 317, 187, 336, -1000, 237, -1000, 60, 334,
	62, 296, -1000, 333, -1000, -47, 318, 399, 124, 296,
	317, -1000, 481, 512, -1000, 378, 512, 454, 408, 512,
	230, -1000, -1000, 215, 52, -1000, 561, -1000, 481, 395,
	-1000, -1000, -1000, 512, 288, 273, -1000, 269, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 512, -1000,
	274, 302, 316, 452, 302, -1000, 512, 296, -1000, 376,
	-81, -1000, 255, -1000, 313, -1000, -1000, 309, -1000, 239,
	140, 550, 445, -1000, 550, 378, 383, 512, 512, 11,
	550, 170, 1, -1000, 481, 481, -1000, 296, 122, 512,
	265, 390, 512, 512, 123, 512, 512, 512, 512, 512,
	512, 512, 512, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -83, 0, 31, 512, 140, 561, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 171, 1, -1000, 463, 42, 550,
	369, 302, 302, 248, -1000, 438, 481, -1000, 550, -1000,
	-1000, -1000, 105, 296, -1000, -68, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 369, 302, -1000, -1000, 512, 512,
	550, 550, -1000, 512, 243, 375, 315, 102, 47, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 550, -1000, 495,
	265, 512, 512, 550, 539, -1000, 361, 218, 218, 218,
	72, 72, -1000, -1000, -1000, -1000, -1000, -1000, -3, -1000,
	-5, 1, -7, 39, -1000, 481, 84, 265, 223, 89,
	-10, -1000, 438, 411, 428, 140, 307, -1000, -1000, 305,
	-1000, -1000, 187, 550, 550, 550, 450, 170, 170, -1000,
	-1000, 192, 159, 179, 178, 172, 53, -1000, 304, 5,
	92, -1000, 550, 379, 512, -1000, -1000, -1000, -12, -1000,
	6, -1000, 512, 37, -1000, 308, 211, -1000, -1000, -1000,
	302, 411, -1000, 512, 512, -1000, -1000, 448, 414, 375,
	76, -1000, 158, -1000, 135, -1000, -1000, -1000, -1000, -25,
	-34, -44, -1000, -1000, -1000, -1000, -1000, 512, 550, -1000,
	-1000, 550, 512, 342, 265, -1000, -1000, 287, 209, -1000,
	303, -1000, 438, 481, 512, 481, -1000, -1000, 264, 245,
	231, 550, 550, 459, -1000, 512, 512, -1000, -1000, -1000,
	411, 140, 202, -1000, 296, 296, 296, 302, 550, -1000,
	320, -16, -1000, -30, -33, 187, -1000, 458, 397, -1000,
	296, -1000, -1000, -1000, 296, -1000, 296, -1000,
}
var yyPgo = []int{

	0, 581, 580, 22, 579, 577, 575, 574, 573, 570,
	569, 563, 545, 562, 561, 559, 20, 24, 554, 552,
	5, 551, 13, 546, 544, 201, 542, 3, 18, 7,
	541, 538, 17, 536, 2, 15, 9, 535, 534, 11,
	527, 8, 526, 524, 16, 515, 514, 513, 512, 12,
	511, 6, 510, 1, 508, 27, 507, 14, 4, 19,
	273, 504, 503, 502, 499, 492, 491, 0, 490, 488,
	10, 487, 486, 485, 484, 479, 477, 105, 34, 476,
	475, 474, 469,
}
var yyR1 = []int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 74, 74, 5, 6, 7, 7, 71, 72,
	73, 76, 79, 79, 80, 80, 80, 81, 81, 75,
	75, 75, 75, 75, 8, 8, 8, 9, 9, 9,
	10, 11, 11, 11, 82, 12, 13, 13, 14, 14,
	14, 14, 14, 15, 15, 16, 16, 17, 17, 17,
	20, 20, 18, 18, 18, 21, 21, 22, 22, 22,
	22, 19, 19, 19, 23, 23, 23, 23, 23, 23,
	23, 23, 23, 24, 24, 24, 24, 24, 24, 24,
	25, 25, 26, 26, 26, 26, 27, 27, 28, 28,
	78, 78, 78, 77, 77, 29, 29, 29, 29, 29,
	30, 30, 30, 30, 30, 30, 30, 30, 30, 30,
	31, 31, 31, 31, 31, 31, 31, 32, 32, 37,
	37, 35, 35, 39, 36, 36, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 38, 38, 40, 40, 40, 42,
	45, 45, 43, 43, 44, 46, 46, 41, 41, 33,
	33, 33, 33, 47, 47, 48, 48, 49, 49, 50,
	50, 51, 52, 52, 52, 53, 53, 53, 54, 54,
	54, 55, 55, 56, 56, 57, 57, 58, 58, 59,
	60, 60, 61, 61, 62, 62, 63, 63, 63, 63,
	63, 64, 64, 65, 65, 66, 66, 67, 68, 69,
	69, 69, 69, 69, 70,
}
var yyR2 = []int{

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
	5, 4, 4, 1, 1, 1, 1, 1, 1, 5,
	0, 1, 1, 2, 4, 0, 2, 1, 3, 1,
	1, 1, 1, 0, 3, 0, 2, 0, 3, 1,
	3, 2, 0, 1, 1, 0, 2, 4, 0, 2,
	4, 0, 3, 1, 3, 0, 5, 1, 3, 3,
	0, 2, 0, 3, 0, 1, 1, 1, 1, 1,
	1, 0, 1, 0, 1, 0, 2, 1, 1, 1,
	1, 1, 1, 1, 0,
}
var yyChk = []int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -71, -72, -73, -74, -75, -76, 5, 6,
	7, 8, 33, 101, 102, 104, 103, 88, 89, 90,
	92, 94, 93, -14, 54, 55, 56, 57, -12, -82,
	-12, -12, -12, -12, 105, -65, 107, 111, -62, 107,
	109, 105, 105, 106, 107, -12, 95, 98, 96, 97,
	-81, 99, -67, 35, -3, 17, -15, 18, -13, 29,
	-25, 35, 9, -58, 91, -59, -41, -67, 35, -61,
	110, 106, -67, 105, -67, 35, -60, 110, -67, -60,
	29, -78, 10, 23, -78, -77, 9, -67, 100, 44,
	-16, -17, 78, -20, 35, -29, -34, -30, 72, 44,
	-33, -41, -35, -40, -67, -38, -42, 20, 36, 37,
	38, 25, -39, 76, 77, 48, 110, 28, 83, 39,
	-25, 33, 81, -25, 58, 35, 45, 81, 35, 72,
	-67, -70, 35, -70, 108, 35, 20, 69, -67, -25,
	-20, -34, 44, -78, -34, -77, -79, 9, 21, -36,
	-34, 9, 58, -18, 70, 71, -67, 19, 81, -31,
	21, 72, 23, 24, 22, 73, 74, 75, 76, 77,
	78, 79, 80, 45, 46, 47, 40, 41, 42, 43,
	-20, -29, -36, -3, -68, -20, -34, -69, 49, 51,
	52, 50, 53, -34, 44, 44, -39, 44, -45, -34,
	-55, 33, 44, -58, 35, -28, 10, -59, -34, -67,
	-70, 20, -66, 112, -63, 104, 102, 32, 103, 13,
	35, 35, 35, -70, -55, 33, -78, -80, 9, 21,
	-34, -34, 113, 58, -21, -22, -24, 44, 35, -39,
	100, 96, -17, -20, -20, -67, 78, -34, -35, 44,
	21, 23, 24, -34, -34, 25, 72, -34, -34, -34,
	-34, -34, -34, -34, -34, 113, 113, 113, -36, 113,
	-16, 18, -16, -43, -44, 84, -32, 28, -3, -58,
	-56, -41, -28, -49, 13, -20, 69, -67, -70, -64,
	108, -32, -58, -34, -34, -34, -28, 58, -23, 59,
	60, 61, 62, 63, 65, 66, -19, 35, 19, -22,
	81, -35, -34, -34, 70, 25, 113, 113, -16, 113,
	-46, -44, 86, -29, -57, 69, -37, -35, -57, 113,
	58, -49, -53, 15, 14, 35, 35, -47, 11, -22,
	-22, 59, 64, 59, 64, 59, 59, 59, -26, 67,
	109, 68, 35, 113, 35, 100, 96, 70, -34, 113,
	87, -34, 85, 30, 58, -41, -53, -34, -50, -51,
	-34, -70, -48, 12, 14, 69, 59, 59, 106, 106,
	106, -34, -34, 31, -35, 58, 58, -52, 26, 27,
	-49, -20, -36, -29, 44, 44, 44, 7, -34, -51,
	-53, -27, -67, -27, -27, -58, -54, 16, 34, 113,
	58, 113, 113, 7, 21, -67, -67, -67,
}
var yyDef = []int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 223, 214, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 212, 0, 0, 224, 0, 0,
	215, 0, 210, 0, 210, 0, 110, 110, 113, 0,
	0, 38, 0, 227, 19, 59, 0, 64, 55, 0,
	0, 100, 0, 26, 0, 207, 0, 177, 227, 0,
	0, 0, 234, 0, 234, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 110, 0, 113, 0, 0,
	17, 65, 67, 72, 227, 70, 71, 115, 0, 0,
	146, 147, 148, 0, 177, 0, 163, 0, 179, 180,
	181, 182, 142, 166, 167, 168, 164, 165, 170, 57,
	201, 0, 0, 108, 0, 27, 0, 0, 234, 0,
	225, 46, 0, 49, 0, 51, 211, 0, 234, 201,
	111, 112, 0, 41, 114, 110, 34, 0, 0, 0,
	144, 0, 0, 68, 0, 0, 73, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 130, 131, 132, 133, 134, 135, 136,
	118, 70, 0, 0, 0, 0, -2, 228, 229, 230,
	231, 232, 233, 157, 0, 0, 129, 0, 0, 171,
	0, 0, 0, 108, 101, 187, 0, 208, 209, 178,
	44, 213, 0, 0, 234, 221, 216, 217, 218, 219,
	220, 50, 52, 53, 0, 0, 42, 43, 0, 0,
	32, 33, 31, 0, 108, 75, 81, 0, 93, 95,
	96, 97, 66, 116, 117, 74, 69, 120, 121, 0,
	0, 0, 0, 123, 0, 127, 0, 149, 150, 151,
	152, 153, 154, 155, 156, 119, 141, 143, 0, 158,
	0, 0, 0, 175, 172, 0, 205, 0, 138, 205,
	0, 203, 187, 195, 0, 109, 0, 226, 47, 0,
	222, 22, 23, 35, 36, 145, 183, 0, 0, 84,
	85, 0, 0, 0, 0, 0, 102, 82, 0, 0,
	0, 122, 124, 0, 0, 128, 161, 159, 0, 162,
	0, 173, 0, 70, 20, 0, 137, 139, 21, 202,
	0, 195, 25, 0, 0, 234, 48, 185, 0, 76,
	79, 86, 0, 88, 0, 90, 91, 92, 77, 0,
	0, 0, 83, 78, 94, 98, 99, 0, 125, 160,
	169, 176, 0, 0, 0, 204, 24, 196, 188, 189,
	192, 45, 187, 0, 0, 0, 87, 89, 0, 0,
	0, 126, 174, 0, 140, 0, 0, 191, 193, 194,
	195, 186, 184, -2, 0, 0, 0, 0, 197, 190,
	198, 0, 106, 0, 0, 206, 18, 0, 0, 103,
	0, 104, 105, 199, 0, 107, 0, 200,
}
var yyTok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 80, 73, 3,
	44, 113, 78, 76, 58, 77, 81, 79, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	46, 45, 47, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 75, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 74, 3, 48,
}
var yyTok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 49, 50, 51, 52, 53, 54, 55, 56,
	57, 59, 60, 61, 62, 63, 64, 65, 66, 67,
	68, 69, 70, 71, 72, 82, 83, 84, 85, 86,
	87, 88, 89, 90, 91, 92, 93, 94, 95, 96,
	97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
	107, 108, 109, 110, 111, 112,
}
var yyTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var yyDebug = 0

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

const yyFlag = -1000

func yyTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(yyToknames) {
		if yyToknames[c-4] != "" {
			return yyToknames[c-4]
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

func yylex1(lex yyLexer, lval *yySymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		c = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			c = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		c = yyTok3[i+0]
		if c == char {
			c = yyTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(c), uint(char))
	}
	return c
}

func yyParse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yychar), yyStatname(yystate))
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
		yychar = yylex1(yylex, &yylval)
	}
	yyn += yychar
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yychar { /* valid shift */
		yychar = -1
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
			yychar = yylex1(yylex, &yylval)
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
			if yyn < 0 || yyn == yychar {
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
			yylex.Error("syntax error")
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yychar))
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
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yychar))
			}
			if yychar == yyEofCode {
				goto ret1
			}
			yychar = -1
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
		//line sql.y:173
		{
			SetParseTree(yylex, yyS[yypt-0].statement)
		}
	case 2:
		//line sql.y:179
		{
			yyVAL.statement = yyS[yypt-0].selStmt
		}
	case 3:
		yyVAL.statement = yyS[yypt-0].statement
	case 4:
		yyVAL.statement = yyS[yypt-0].statement
	case 5:
		yyVAL.statement = yyS[yypt-0].statement
	case 6:
		yyVAL.statement = yyS[yypt-0].statement
	case 7:
		yyVAL.statement = yyS[yypt-0].statement
	case 8:
		yyVAL.statement = yyS[yypt-0].statement
	case 9:
		yyVAL.statement = yyS[yypt-0].statement
	case 10:
		yyVAL.statement = yyS[yypt-0].statement
	case 11:
		yyVAL.statement = yyS[yypt-0].statement
	case 12:
		yyVAL.statement = yyS[yypt-0].statement
	case 13:
		yyVAL.statement = yyS[yypt-0].statement
	case 14:
		yyVAL.statement = yyS[yypt-0].statement
	case 15:
		yyVAL.statement = yyS[yypt-0].statement
	case 16:
		yyVAL.statement = yyS[yypt-0].statement
	case 17:
		//line sql.y:199
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyS[yypt-2].bytes2), Distinct: yyS[yypt-1].str, SelectExprs: yyS[yypt-0].selectExprs}
		}
	case 18:
		//line sql.y:203
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyS[yypt-10].bytes2), Distinct: yyS[yypt-9].str, SelectExprs: yyS[yypt-8].selectExprs, From: yyS[yypt-6].tableExprs, Where: NewWhere(AST_WHERE, yyS[yypt-5].expr), GroupBy: GroupBy(yyS[yypt-4].valExprs), Having: NewWhere(AST_HAVING, yyS[yypt-3].expr), OrderBy: yyS[yypt-2].orderBy, Limit: yyS[yypt-1].limit, Lock: yyS[yypt-0].str}
		}
	case 19:
		//line sql.y:207
		{
			yyVAL.selStmt = &Union{Type: yyS[yypt-1].str, Left: yyS[yypt-2].selStmt, Right: yyS[yypt-0].selStmt}
		}
	case 20:
		//line sql.y:214
		{
			yyVAL.statement = &Insert{Comments: Comments(yyS[yypt-5].bytes2), Table: yyS[yypt-3].tableName, Columns: yyS[yypt-2].columns, Rows: yyS[yypt-1].insRows, OnDup: OnDup(yyS[yypt-0].updateExprs)}
		}
	case 21:
		//line sql.y:218
		{
			cols := make(Columns, 0, len(yyS[yypt-1].updateExprs))
			vals := make(ValTuple, 0, len(yyS[yypt-1].updateExprs))
			for _, col := range yyS[yypt-1].updateExprs {
				cols = append(cols, &NonStarExpr{Expr: col.Name})
				vals = append(vals, col.Expr)
			}
			yyVAL.statement = &Insert{Comments: Comments(yyS[yypt-5].bytes2), Table: yyS[yypt-3].tableName, Columns: cols, Rows: Values{vals}, OnDup: OnDup(yyS[yypt-0].updateExprs)}
		}
	case 22:
		//line sql.y:230
		{
			yyVAL.statement = &Replace{Comments: Comments(yyS[yypt-4].bytes2), Table: yyS[yypt-2].tableName, Columns: yyS[yypt-1].columns, Rows: yyS[yypt-0].insRows}
		}
	case 23:
		//line sql.y:234
		{
			cols := make(Columns, 0, len(yyS[yypt-0].updateExprs))
			vals := make(ValTuple, 0, len(yyS[yypt-0].updateExprs))
			for _, col := range yyS[yypt-0].updateExprs {
				cols = append(cols, &NonStarExpr{Expr: col.Name})
				vals = append(vals, col.Expr)
			}
			yyVAL.statement = &Replace{Comments: Comments(yyS[yypt-4].bytes2), Table: yyS[yypt-2].tableName, Columns: cols, Rows: Values{vals}}
		}
	case 24:
		//line sql.y:247
		{
			yyVAL.statement = &Update{Comments: Comments(yyS[yypt-6].bytes2), Table: yyS[yypt-5].tableName, Exprs: yyS[yypt-3].updateExprs, Where: NewWhere(AST_WHERE, yyS[yypt-2].expr), OrderBy: yyS[yypt-1].orderBy, Limit: yyS[yypt-0].limit}
		}
	case 25:
		//line sql.y:253
		{
			yyVAL.statement = &Delete{Comments: Comments(yyS[yypt-5].bytes2), Table: yyS[yypt-3].tableName, Where: NewWhere(AST_WHERE, yyS[yypt-2].expr), OrderBy: yyS[yypt-1].orderBy, Limit: yyS[yypt-0].limit}
		}
	case 26:
		//line sql.y:259
		{
			yyVAL.statement = &Set{Comments: Comments(yyS[yypt-1].bytes2), Exprs: yyS[yypt-0].updateExprs}
		}
	case 27:
		//line sql.y:263
		{
			yyVAL.statement = &Set{Comments: Comments(yyS[yypt-2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyS[yypt-0].bytes)}}}
		}
	case 28:
		//line sql.y:269
		{
			yyVAL.statement = &Begin{}
		}
	case 29:
		//line sql.y:275
		{
			yyVAL.statement = &Commit{}
		}
	case 30:
		//line sql.y:281
		{
			yyVAL.statement = &Rollback{}
		}
	case 31:
		//line sql.y:287
		{
			yyVAL.statement = &Admin{Name: yyS[yypt-3].bytes, Values: yyS[yypt-1].valExprs}
		}
	case 32:
		//line sql.y:293
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 33:
		//line sql.y:297
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 34:
		//line sql.y:302
		{
			yyVAL.valExpr = nil
		}
	case 35:
		//line sql.y:306
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 36:
		//line sql.y:310
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 37:
		//line sql.y:315
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 38:
		//line sql.y:319
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 39:
		//line sql.y:326
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 40:
		//line sql.y:330
		{
			yyVAL.statement = &Show{Section: "variables", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 41:
		//line sql.y:334
		{
			yyVAL.statement = &Show{Section: "tables", From: yyS[yypt-1].valExpr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 42:
		//line sql.y:338
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyS[yypt-2].bytes), From: yyS[yypt-1].valExpr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 43:
		//line sql.y:342
		{
			yyVAL.statement = &Show{Section: "columns", From: yyS[yypt-1].valExpr, Modifier: yyS[yypt-3].str, DBFilter: yyS[yypt-0].valExpr}
		}
	case 44:
		//line sql.y:348
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyS[yypt-1].bytes}
		}
	case 45:
		//line sql.y:352
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyS[yypt-1].bytes, NewName: yyS[yypt-1].bytes}
		}
	case 46:
		//line sql.y:357
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyS[yypt-1].bytes}
		}
	case 47:
		//line sql.y:363
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyS[yypt-2].bytes, NewName: yyS[yypt-2].bytes}
		}
	case 48:
		//line sql.y:367
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyS[yypt-3].bytes, NewName: yyS[yypt-0].bytes}
		}
	case 49:
		//line sql.y:372
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyS[yypt-1].bytes, NewName: yyS[yypt-1].bytes}
		}
	case 50:
		//line sql.y:378
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyS[yypt-2].bytes, NewName: yyS[yypt-0].bytes}
		}
	case 51:
		//line sql.y:384
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyS[yypt-0].bytes}
		}
	case 52:
		//line sql.y:388
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyS[yypt-0].bytes, NewName: yyS[yypt-0].bytes}
		}
	case 53:
		//line sql.y:393
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyS[yypt-1].bytes}
		}
	case 54:
		//line sql.y:398
		{
			SetAllowComments(yylex, true)
		}
	case 55:
		//line sql.y:402
		{
			yyVAL.bytes2 = yyS[yypt-0].bytes2
			SetAllowComments(yylex, false)
		}
	case 56:
		//line sql.y:408
		{
			yyVAL.bytes2 = nil
		}
	case 57:
		//line sql.y:412
		{
			yyVAL.bytes2 = append(yyS[yypt-1].bytes2, yyS[yypt-0].bytes)
		}
	case 58:
		//line sql.y:418
		{
			yyVAL.str = AST_UNION
		}
	case 59:
		//line sql.y:422
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 60:
		//line sql.y:426
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 61:
		//line sql.y:430
		{
			yyVAL.str = AST_EXCEPT
		}
	case 62:
		//line sql.y:434
		{
			yyVAL.str = AST_INTERSECT
		}
	case 63:
		//line sql.y:439
		{
			yyVAL.str = ""
		}
	case 64:
		//line sql.y:443
		{
			yyVAL.str = AST_DISTINCT
		}
	case 65:
		//line sql.y:449
		{
			yyVAL.selectExprs = SelectExprs{yyS[yypt-0].selectExpr}
		}
	case 66:
		//line sql.y:453
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyS[yypt-0].selectExpr)
		}
	case 67:
		//line sql.y:459
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 68:
		//line sql.y:463
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyS[yypt-1].expr, As: yyS[yypt-0].bytes}
		}
	case 69:
		//line sql.y:467
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyS[yypt-2].bytes}
		}
	case 70:
		//line sql.y:473
		{
			yyVAL.expr = yyS[yypt-0].boolExpr
		}
	case 71:
		//line sql.y:477
		{
			yyVAL.expr = yyS[yypt-0].valExpr
		}
	case 72:
		//line sql.y:482
		{
			yyVAL.bytes = nil
		}
	case 73:
		//line sql.y:486
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 74:
		//line sql.y:490
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 75:
		//line sql.y:496
		{
			yyVAL.tableExprs = TableExprs{yyS[yypt-0].tableExpr}
		}
	case 76:
		//line sql.y:500
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyS[yypt-0].tableExpr)
		}
	case 77:
		//line sql.y:506
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyS[yypt-2].smTableExpr, As: yyS[yypt-1].bytes, Hints: yyS[yypt-0].indexHints}
		}
	case 78:
		//line sql.y:510
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyS[yypt-1].tableExpr}
		}
	case 79:
		//line sql.y:514
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-2].tableExpr, Join: yyS[yypt-1].str, RightExpr: yyS[yypt-0].tableExpr}
		}
	case 80:
		//line sql.y:518
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-4].tableExpr, Join: yyS[yypt-3].str, RightExpr: yyS[yypt-2].tableExpr, On: yyS[yypt-0].boolExpr}
		}
	case 81:
		//line sql.y:523
		{
			yyVAL.bytes = nil
		}
	case 82:
		//line sql.y:527
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 83:
		//line sql.y:531
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 84:
		//line sql.y:537
		{
			yyVAL.str = AST_JOIN
		}
	case 85:
		//line sql.y:541
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 86:
		//line sql.y:545
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 87:
		//line sql.y:549
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 88:
		//line sql.y:553
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 89:
		//line sql.y:557
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 90:
		//line sql.y:561
		{
			yyVAL.str = AST_JOIN
		}
	case 91:
		//line sql.y:565
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 92:
		//line sql.y:569
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 93:
		//line sql.y:575
		{
			yyVAL.smTableExpr = &TableName{Name: yyS[yypt-0].bytes}
		}
	case 94:
		//line sql.y:579
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyS[yypt-2].bytes, Name: yyS[yypt-0].bytes}
		}
	case 95:
		//line sql.y:583
		{
			yyVAL.smTableExpr = yyS[yypt-0].subquery
		}
	case 96:
		//line sql.y:587
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 97:
		//line sql.y:591
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 98:
		//line sql.y:595
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyS[yypt-2].bytes, Name: []byte("columns")}
		}
	case 99:
		//line sql.y:599
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyS[yypt-2].bytes, Name: []byte("tables")}
		}
	case 100:
		//line sql.y:605
		{
			yyVAL.tableName = &TableName{Name: yyS[yypt-0].bytes}
		}
	case 101:
		//line sql.y:609
		{
			yyVAL.tableName = &TableName{Qualifier: yyS[yypt-2].bytes, Name: yyS[yypt-0].bytes}
		}
	case 102:
		//line sql.y:614
		{
			yyVAL.indexHints = nil
		}
	case 103:
		//line sql.y:618
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyS[yypt-1].bytes2}
		}
	case 104:
		//line sql.y:622
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyS[yypt-1].bytes2}
		}
	case 105:
		//line sql.y:626
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyS[yypt-1].bytes2}
		}
	case 106:
		//line sql.y:632
		{
			yyVAL.bytes2 = [][]byte{yyS[yypt-0].bytes}
		}
	case 107:
		//line sql.y:636
		{
			yyVAL.bytes2 = append(yyS[yypt-2].bytes2, yyS[yypt-0].bytes)
		}
	case 108:
		//line sql.y:641
		{
			yyVAL.expr = nil
		}
	case 109:
		//line sql.y:645
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 110:
		//line sql.y:650
		{
			yyVAL.expr = nil
		}
	case 111:
		//line sql.y:654
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 112:
		//line sql.y:658
		{
			yyVAL.expr = yyS[yypt-0].valExpr
		}
	case 113:
		//line sql.y:663
		{
			yyVAL.valExpr = nil
		}
	case 114:
		//line sql.y:667
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 115:
		yyVAL.boolExpr = yyS[yypt-0].boolExpr
	case 116:
		//line sql.y:674
		{
			yyVAL.boolExpr = &AndExpr{Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 117:
		//line sql.y:678
		{
			yyVAL.boolExpr = &OrExpr{Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 118:
		//line sql.y:682
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyS[yypt-0].expr}
		}
	case 119:
		//line sql.y:686
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyS[yypt-1].boolExpr}
		}
	case 120:
		//line sql.y:692
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyS[yypt-2].valExpr, Operator: yyS[yypt-1].str, Right: yyS[yypt-0].valExpr}
		}
	case 121:
		//line sql.y:696
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyS[yypt-2].valExpr, Operator: AST_IN, Right: yyS[yypt-0].tuple}
		}
	case 122:
		//line sql.y:700
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyS[yypt-3].valExpr, Operator: AST_NOT_IN, Right: yyS[yypt-0].tuple}
		}
	case 123:
		//line sql.y:704
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyS[yypt-2].valExpr, Operator: AST_LIKE, Right: yyS[yypt-0].valExpr}
		}
	case 124:
		//line sql.y:708
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyS[yypt-3].valExpr, Operator: AST_NOT_LIKE, Right: yyS[yypt-0].valExpr}
		}
	case 125:
		//line sql.y:712
		{
			yyVAL.boolExpr = &RangeCond{Left: yyS[yypt-4].valExpr, Operator: AST_BETWEEN, From: yyS[yypt-2].valExpr, To: yyS[yypt-0].valExpr}
		}
	case 126:
		//line sql.y:716
		{
			yyVAL.boolExpr = &RangeCond{Left: yyS[yypt-5].valExpr, Operator: AST_NOT_BETWEEN, From: yyS[yypt-2].valExpr, To: yyS[yypt-0].valExpr}
		}
	case 127:
		//line sql.y:720
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NULL, Expr: yyS[yypt-2].valExpr}
		}
	case 128:
		//line sql.y:724
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NOT_NULL, Expr: yyS[yypt-3].valExpr}
		}
	case 129:
		//line sql.y:728
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyS[yypt-0].subquery}
		}
	case 130:
		//line sql.y:734
		{
			yyVAL.str = AST_EQ
		}
	case 131:
		//line sql.y:738
		{
			yyVAL.str = AST_LT
		}
	case 132:
		//line sql.y:742
		{
			yyVAL.str = AST_GT
		}
	case 133:
		//line sql.y:746
		{
			yyVAL.str = AST_LE
		}
	case 134:
		//line sql.y:750
		{
			yyVAL.str = AST_GE
		}
	case 135:
		//line sql.y:754
		{
			yyVAL.str = AST_NE
		}
	case 136:
		//line sql.y:758
		{
			yyVAL.str = AST_NSE
		}
	case 137:
		//line sql.y:764
		{
			yyVAL.insRows = yyS[yypt-0].values
		}
	case 138:
		//line sql.y:768
		{
			yyVAL.insRows = yyS[yypt-0].selStmt
		}
	case 139:
		//line sql.y:774
		{
			yyVAL.values = Values{yyS[yypt-0].tuple}
		}
	case 140:
		//line sql.y:778
		{
			yyVAL.values = append(yyS[yypt-2].values, yyS[yypt-0].tuple)
		}
	case 141:
		//line sql.y:784
		{
			yyVAL.tuple = ValTuple(yyS[yypt-1].valExprs)
		}
	case 142:
		//line sql.y:788
		{
			yyVAL.tuple = yyS[yypt-0].subquery
		}
	case 143:
		//line sql.y:794
		{
			yyVAL.subquery = &Subquery{yyS[yypt-1].selStmt}
		}
	case 144:
		//line sql.y:800
		{
			yyVAL.valExprs = ValExprs{yyS[yypt-0].valExpr}
		}
	case 145:
		//line sql.y:804
		{
			yyVAL.valExprs = append(yyS[yypt-2].valExprs, yyS[yypt-0].valExpr)
		}
	case 146:
		//line sql.y:810
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 147:
		//line sql.y:814
		{
			yyVAL.valExpr = yyS[yypt-0].colName
		}
	case 148:
		//line sql.y:818
		{
			yyVAL.valExpr = yyS[yypt-0].tuple
		}
	case 149:
		//line sql.y:822
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_BITAND, Right: yyS[yypt-0].valExpr}
		}
	case 150:
		//line sql.y:826
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_BITOR, Right: yyS[yypt-0].valExpr}
		}
	case 151:
		//line sql.y:830
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_BITXOR, Right: yyS[yypt-0].valExpr}
		}
	case 152:
		//line sql.y:834
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_PLUS, Right: yyS[yypt-0].valExpr}
		}
	case 153:
		//line sql.y:838
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_MINUS, Right: yyS[yypt-0].valExpr}
		}
	case 154:
		//line sql.y:842
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_MULT, Right: yyS[yypt-0].valExpr}
		}
	case 155:
		//line sql.y:846
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_DIV, Right: yyS[yypt-0].valExpr}
		}
	case 156:
		//line sql.y:850
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_MOD, Right: yyS[yypt-0].valExpr}
		}
	case 157:
		//line sql.y:854
		{
			if num, ok := yyS[yypt-0].valExpr.(NumVal); ok {
				switch yyS[yypt-1].byt {
				case '-':
					yyVAL.valExpr = append(NumVal("-"), num...)
				case '+':
					yyVAL.valExpr = num
				default:
					yyVAL.valExpr = &UnaryExpr{Operator: yyS[yypt-1].byt, Expr: yyS[yypt-0].valExpr}
				}
			} else {
				yyVAL.valExpr = &UnaryExpr{Operator: yyS[yypt-1].byt, Expr: yyS[yypt-0].valExpr}
			}
		}
	case 158:
		//line sql.y:869
		{
			yyVAL.valExpr = &FuncExpr{Name: yyS[yypt-2].bytes}
		}
	case 159:
		//line sql.y:873
		{
			yyVAL.valExpr = &FuncExpr{Name: yyS[yypt-3].bytes, Exprs: yyS[yypt-1].selectExprs}
		}
	case 160:
		//line sql.y:877
		{
			yyVAL.valExpr = &FuncExpr{Name: yyS[yypt-4].bytes, Distinct: true, Exprs: yyS[yypt-1].selectExprs}
		}
	case 161:
		//line sql.y:881
		{
			yyVAL.valExpr = &CtorExpr{Name: yyS[yypt-2].str, Exprs: yyS[yypt-1].valExprs}
		}
	case 162:
		//line sql.y:885
		{
			yyVAL.valExpr = &FuncExpr{Name: yyS[yypt-3].bytes, Exprs: yyS[yypt-1].selectExprs}
		}
	case 163:
		//line sql.y:889
		{
			yyVAL.valExpr = yyS[yypt-0].caseExpr
		}
	case 164:
		//line sql.y:895
		{
			yyVAL.bytes = IF_BYTES
		}
	case 165:
		//line sql.y:899
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 166:
		//line sql.y:905
		{
			yyVAL.byt = AST_UPLUS
		}
	case 167:
		//line sql.y:909
		{
			yyVAL.byt = AST_UMINUS
		}
	case 168:
		//line sql.y:913
		{
			yyVAL.byt = AST_TILDA
		}
	case 169:
		//line sql.y:919
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyS[yypt-3].valExpr, Whens: yyS[yypt-2].whens, Else: yyS[yypt-1].valExpr}
		}
	case 170:
		//line sql.y:924
		{
			yyVAL.valExpr = nil
		}
	case 171:
		//line sql.y:928
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 172:
		//line sql.y:934
		{
			yyVAL.whens = []*When{yyS[yypt-0].when}
		}
	case 173:
		//line sql.y:938
		{
			yyVAL.whens = append(yyS[yypt-1].whens, yyS[yypt-0].when)
		}
	case 174:
		//line sql.y:944
		{
			yyVAL.when = &When{Cond: yyS[yypt-2].boolExpr, Val: yyS[yypt-0].valExpr}
		}
	case 175:
		//line sql.y:949
		{
			yyVAL.valExpr = nil
		}
	case 176:
		//line sql.y:953
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 177:
		//line sql.y:959
		{
			yyVAL.colName = &ColName{Name: yyS[yypt-0].bytes}
		}
	case 178:
		//line sql.y:963
		{
			yyVAL.colName = &ColName{Qualifier: yyS[yypt-2].bytes, Name: yyS[yypt-0].bytes}
		}
	case 179:
		//line sql.y:969
		{
			yyVAL.valExpr = StrVal(yyS[yypt-0].bytes)
		}
	case 180:
		//line sql.y:973
		{
			yyVAL.valExpr = NumVal(yyS[yypt-0].bytes)
		}
	case 181:
		//line sql.y:977
		{
			yyVAL.valExpr = ValArg(yyS[yypt-0].bytes)
		}
	case 182:
		//line sql.y:981
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 183:
		//line sql.y:986
		{
			yyVAL.valExprs = nil
		}
	case 184:
		//line sql.y:990
		{
			yyVAL.valExprs = yyS[yypt-0].valExprs
		}
	case 185:
		//line sql.y:995
		{
			yyVAL.expr = nil
		}
	case 186:
		//line sql.y:999
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 187:
		//line sql.y:1004
		{
			yyVAL.orderBy = nil
		}
	case 188:
		//line sql.y:1008
		{
			yyVAL.orderBy = yyS[yypt-0].orderBy
		}
	case 189:
		//line sql.y:1014
		{
			yyVAL.orderBy = OrderBy{yyS[yypt-0].order}
		}
	case 190:
		//line sql.y:1018
		{
			yyVAL.orderBy = append(yyS[yypt-2].orderBy, yyS[yypt-0].order)
		}
	case 191:
		//line sql.y:1024
		{
			yyVAL.order = &Order{Expr: yyS[yypt-1].valExpr, Direction: yyS[yypt-0].str}
		}
	case 192:
		//line sql.y:1029
		{
			yyVAL.str = AST_ASC
		}
	case 193:
		//line sql.y:1033
		{
			yyVAL.str = AST_ASC
		}
	case 194:
		//line sql.y:1037
		{
			yyVAL.str = AST_DESC
		}
	case 195:
		//line sql.y:1042
		{
			yyVAL.limit = nil
		}
	case 196:
		//line sql.y:1046
		{
			yyVAL.limit = &Limit{Rowcount: yyS[yypt-0].valExpr}
		}
	case 197:
		//line sql.y:1050
		{
			yyVAL.limit = &Limit{Offset: yyS[yypt-2].valExpr, Rowcount: yyS[yypt-0].valExpr}
		}
	case 198:
		//line sql.y:1055
		{
			yyVAL.str = ""
		}
	case 199:
		//line sql.y:1059
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 200:
		//line sql.y:1063
		{
			if !bytes.Equal(yyS[yypt-1].bytes, SHARE) {
				yylex.Error("expecting share")
				return 1
			}
			if !bytes.Equal(yyS[yypt-0].bytes, MODE) {
				yylex.Error("expecting mode")
				return 1
			}
			yyVAL.str = AST_SHARE_MODE
		}
	case 201:
		//line sql.y:1076
		{
			yyVAL.columns = nil
		}
	case 202:
		//line sql.y:1080
		{
			yyVAL.columns = yyS[yypt-1].columns
		}
	case 203:
		//line sql.y:1086
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyS[yypt-0].colName}}
		}
	case 204:
		//line sql.y:1090
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyS[yypt-0].colName})
		}
	case 205:
		//line sql.y:1095
		{
			yyVAL.updateExprs = nil
		}
	case 206:
		//line sql.y:1099
		{
			yyVAL.updateExprs = yyS[yypt-0].updateExprs
		}
	case 207:
		//line sql.y:1105
		{
			yyVAL.updateExprs = UpdateExprs{yyS[yypt-0].updateExpr}
		}
	case 208:
		//line sql.y:1109
		{
			yyVAL.updateExprs = append(yyS[yypt-2].updateExprs, yyS[yypt-0].updateExpr)
		}
	case 209:
		//line sql.y:1115
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyS[yypt-2].colName, Expr: yyS[yypt-0].valExpr}
		}
	case 210:
		//line sql.y:1120
		{
			yyVAL.empty = struct{}{}
		}
	case 211:
		//line sql.y:1122
		{
			yyVAL.empty = struct{}{}
		}
	case 212:
		//line sql.y:1125
		{
			yyVAL.empty = struct{}{}
		}
	case 213:
		//line sql.y:1127
		{
			yyVAL.empty = struct{}{}
		}
	case 214:
		//line sql.y:1130
		{
			yyVAL.empty = struct{}{}
		}
	case 215:
		//line sql.y:1132
		{
			yyVAL.empty = struct{}{}
		}
	case 216:
		//line sql.y:1136
		{
			yyVAL.empty = struct{}{}
		}
	case 217:
		//line sql.y:1138
		{
			yyVAL.empty = struct{}{}
		}
	case 218:
		//line sql.y:1140
		{
			yyVAL.empty = struct{}{}
		}
	case 219:
		//line sql.y:1142
		{
			yyVAL.empty = struct{}{}
		}
	case 220:
		//line sql.y:1144
		{
			yyVAL.empty = struct{}{}
		}
	case 221:
		//line sql.y:1147
		{
			yyVAL.empty = struct{}{}
		}
	case 222:
		//line sql.y:1149
		{
			yyVAL.empty = struct{}{}
		}
	case 223:
		//line sql.y:1152
		{
			yyVAL.empty = struct{}{}
		}
	case 224:
		//line sql.y:1154
		{
			yyVAL.empty = struct{}{}
		}
	case 225:
		//line sql.y:1157
		{
			yyVAL.empty = struct{}{}
		}
	case 226:
		//line sql.y:1159
		{
			yyVAL.empty = struct{}{}
		}
	case 227:
		//line sql.y:1163
		{
			yyVAL.bytes = bytes.ToLower(yyS[yypt-0].bytes)
		}
	case 228:
		//line sql.y:1170
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 229:
		//line sql.y:1176
		{
			yyVAL.str = AST_DATE
		}
	case 230:
		//line sql.y:1180
		{
			yyVAL.str = AST_TIME
		}
	case 231:
		//line sql.y:1184
		{
			yyVAL.str = AST_TIMESTAMP
		}
	case 232:
		//line sql.y:1188
		{
			yyVAL.str = AST_DATETIME
		}
	case 233:
		//line sql.y:1192
		{
			yyVAL.str = AST_YEAR
		}
	case 234:
		//line sql.y:1197
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
