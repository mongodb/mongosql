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
const PRECISION = 57361
const AS = 57362
const EXISTS = 57363
const IN = 57364
const IS = 57365
const LIKE = 57366
const BETWEEN = 57367
const NULL = 57368
const ASC = 57369
const DESC = 57370
const VALUES = 57371
const INTO = 57372
const DUPLICATE = 57373
const KEY = 57374
const DEFAULT = 57375
const SET = 57376
const LOCK = 57377
const ID = 57378
const STRING = 57379
const NUMBER = 57380
const VALUE_ARG = 57381
const COMMENT = 57382
const LE = 57383
const GE = 57384
const NE = 57385
const NULL_SAFE_EQUAL = 57386
const DATE = 57387
const DATETIME = 57388
const TIME = 57389
const TIMESTAMP = 57390
const YEAR = 57391
const UNION = 57392
const MINUS = 57393
const EXCEPT = 57394
const INTERSECT = 57395
const JOIN = 57396
const STRAIGHT_JOIN = 57397
const LEFT = 57398
const RIGHT = 57399
const INNER = 57400
const OUTER = 57401
const CROSS = 57402
const NATURAL = 57403
const USE = 57404
const FORCE = 57405
const ON = 57406
const AND = 57407
const OR = 57408
const NOT = 57409
const UNARY = 57410
const CASE = 57411
const WHEN = 57412
const THEN = 57413
const ELSE = 57414
const END = 57415
const BEGIN = 57416
const COMMIT = 57417
const ROLLBACK = 57418
const NAMES = 57419
const REPLACE = 57420
const ADMIN = 57421
const SHOW = 57422
const DATABASES = 57423
const TABLES = 57424
const PROXY = 57425
const VARIABLES = 57426
const FULL = 57427
const COLUMNS = 57428
const CREATE = 57429
const ALTER = 57430
const DROP = 57431
const RENAME = 57432
const TABLE = 57433
const INDEX = 57434
const VIEW = 57435
const TO = 57436
const IGNORE = 57437
const IF = 57438
const UNIQUE = 57439
const USING = 57440

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
	71, 72,
	72, 72,
	-2, 146,
	-1, 405,
	71, 71,
	72, 71,
	-2, 82,
}

const yyNprod = 237
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 651

var yyAct = []int{

	114, 344, 106, 413, 73, 103, 381, 105, 111, 192,
	141, 122, 295, 245, 336, 112, 286, 215, 288, 194,
	100, 193, 3, 75, 210, 101, 121, 277, 422, 127,
	223, 87, 80, 62, 91, 302, 78, 118, 119, 120,
	144, 49, 422, 50, 77, 152, 422, 82, 162, 125,
	84, 342, 76, 162, 88, 64, 392, 44, 162, 46,
	97, 391, 390, 47, 311, 312, 313, 314, 315, 243,
	316, 317, 243, 243, 52, 53, 54, 123, 124, 361,
	363, 81, 140, 424, 128, 34, 35, 36, 37, 366,
	148, 83, 94, 372, 51, 143, 151, 423, 150, 154,
	78, 421, 160, 371, 166, 374, 341, 98, 331, 159,
	287, 126, 196, 329, 190, 195, 203, 191, 365, 95,
	139, 362, 18, 287, 328, 334, 322, 278, 242, 206,
	153, 209, 77, 248, 168, 77, 213, 137, 219, 218,
	76, 132, 247, 76, 279, 180, 181, 182, 387, 220,
	368, 337, 63, 248, 367, 160, 74, 298, 217, 233,
	240, 241, 247, 56, 58, 59, 57, 61, 256, 219,
	254, 255, 259, 249, 234, 265, 266, 147, 269, 270,
	271, 272, 273, 274, 275, 276, 260, 257, 252, 389,
	236, 117, 164, 165, 251, 258, 121, 160, 250, 127,
	18, 19, 20, 21, 280, 267, 104, 118, 119, 120,
	70, 388, 77, 77, 251, 109, 291, 155, 250, 125,
	76, 293, 297, 167, 299, 282, 284, 134, 216, 22,
	216, 294, 290, 359, 358, 300, 77, 357, 337, 63,
	304, 305, 306, 108, 76, 134, 307, 123, 124, 102,
	229, 243, 268, 303, 128, 398, 290, 136, 355, 249,
	353, 321, 308, 356, 160, 354, 324, 325, 376, 408,
	227, 407, 86, 230, 164, 165, 161, 309, 323, 134,
	130, 126, 131, 133, 27, 28, 29, 406, 30, 32,
	31, 261, 235, 195, 207, 335, 205, 23, 24, 26,
	25, 149, 333, 212, 330, 339, 340, 343, 34, 35,
	36, 37, 400, 401, 211, 178, 179, 180, 181, 182,
	204, 249, 249, 351, 352, 212, 162, 89, 99, 370,
	129, 320, 311, 312, 313, 314, 315, 373, 316, 317,
	226, 228, 225, 77, 63, 378, 78, 319, 379, 382,
	364, 377, 348, 63, 419, 347, 232, 231, 383, 175,
	176, 177, 178, 179, 180, 181, 182, 198, 201, 199,
	200, 202, 393, 420, 395, 369, 214, 394, 175, 176,
	177, 178, 179, 180, 181, 182, 71, 145, 142, 160,
	138, 403, 396, 195, 135, 405, 404, 402, 85, 375,
	410, 382, 18, 90, 412, 411, 69, 414, 414, 414,
	77, 415, 416, 283, 417, 327, 117, 18, 76, 253,
	426, 121, 238, 427, 127, 92, 289, 428, 221, 429,
	146, 104, 118, 119, 120, 239, 67, 65, 121, 93,
	109, 127, 157, 262, 125, 263, 264, 345, 78, 118,
	119, 120, 386, 346, 18, 158, 296, 152, 385, 350,
	38, 125, 198, 201, 199, 200, 202, 216, 108, 96,
	117, 72, 123, 124, 102, 121, 425, 409, 127, 128,
	40, 41, 42, 43, 18, 78, 118, 119, 120, 123,
	124, 55, 39, 60, 109, 237, 128, 156, 125, 198,
	201, 199, 200, 202, 17, 16, 126, 15, 14, 281,
	13, 18, 12, 117, 197, 222, 45, 301, 121, 224,
	48, 127, 108, 126, 79, 292, 123, 124, 78, 118,
	119, 120, 121, 128, 418, 127, 399, 109, 380, 384,
	349, 125, 78, 118, 119, 120, 332, 208, 285, 116,
	113, 152, 115, 338, 110, 125, 169, 107, 360, 246,
	126, 310, 244, 318, 163, 108, 66, 33, 68, 123,
	124, 11, 170, 174, 172, 173, 128, 10, 9, 8,
	7, 6, 5, 123, 124, 4, 2, 1, 0, 397,
	128, 186, 187, 188, 189, 0, 183, 184, 185, 0,
	0, 0, 0, 126, 175, 176, 177, 178, 179, 180,
	181, 182, 0, 0, 0, 0, 0, 126, 0, 0,
	0, 0, 0, 171, 175, 176, 177, 178, 179, 180,
	181, 182, 326, 0, 0, 175, 176, 177, 178, 179,
	180, 181, 182, 175, 176, 177, 178, 179, 180, 181,
	182,
}
var yyPact = []int{

	195, -1000, -1000, 253, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -49, -67, -12, -32, -1000, -1000, -1000,
	-1000, 67, 308, 479, 420, -1000, -1000, -1000, 418, -1000,
	376, 350, 462, 64, -79, -26, 308, -1000, -15, 308,
	-1000, 362, -80, 308, -80, 373, 415, 415, 460, 308,
	6, -1000, 283, -1000, -1000, -1000, 170, -1000, 290, 350,
	248, 59, 350, 186, 358, -1000, 211, -1000, 55, 354,
	47, 308, -1000, 352, -1000, -69, 351, 409, 107, 308,
	350, -1000, 492, 0, -1000, 415, 0, 460, 433, 0,
	267, -1000, -1000, 203, 52, -1000, 550, -1000, 492, 449,
	-1000, -1000, -1000, 0, 275, 251, -1000, 249, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 0, -1000,
	280, 310, 340, 457, 310, -1000, 0, 308, -1000, 407,
	-83, -1000, 237, -1000, 321, -1000, -1000, 320, -1000, 258,
	121, 569, 412, -1000, 569, 415, 413, 0, 0, 14,
	569, 97, 170, 400, 492, 492, -1000, 317, 116, 0,
	246, 421, 0, 0, 179, 0, 0, 0, 0, 0,
	0, 0, 0, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -87, 13, 30, 0, 121, 550, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 395, 170, -1000, 479, 25, 569,
	397, 310, 310, 220, -1000, 443, 492, -1000, 569, -1000,
	-1000, -1000, 87, 308, -1000, -74, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 397, 310, -1000, -1000, 0, 0,
	569, 569, -1000, 0, 218, 272, 311, 117, 44, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 569,
	-1000, 506, 246, 0, 0, 569, 561, -1000, 389, 238,
	238, 238, 66, 66, -1000, -1000, -1000, -1000, -1000, -1000,
	10, -1000, -1, 170, -6, 38, -1000, 492, 81, 246,
	253, 168, -8, -1000, 443, 432, 439, 121, 319, -1000,
	-1000, 316, -1000, -1000, 186, 569, 569, 569, 448, 97,
	97, -1000, -1000, 200, 198, 177, 174, 173, 11, -1000,
	314, 4, 53, -1000, 569, 304, 0, -1000, -1000, -1000,
	-11, -1000, 5, -1000, 0, 19, -1000, 368, 209, -1000,
	-1000, -1000, 310, 432, -1000, 0, 0, -1000, -1000, 446,
	438, 272, 78, -1000, 151, -1000, 129, -1000, -1000, -1000,
	-1000, -45, -46, -51, -1000, -1000, -1000, -1000, -1000, 0,
	569, -1000, -1000, 569, 0, 342, 246, -1000, -1000, 530,
	196, -1000, 285, -1000, 443, 492, 0, 492, -1000, -1000,
	242, 226, 224, 569, 569, 470, -1000, 0, 0, -1000,
	-1000, -1000, 432, 121, 192, -1000, 308, 308, 308, 310,
	569, -1000, 338, -13, -1000, -17, -31, 186, -1000, 469,
	398, -1000, 308, -1000, -1000, -1000, 308, -1000, 308, -1000,
}
var yyPgo = []int{

	0, 587, 586, 21, 585, 582, 581, 580, 579, 578,
	577, 571, 460, 568, 567, 566, 20, 25, 564, 563,
	5, 562, 13, 561, 559, 210, 558, 3, 17, 7,
	557, 556, 18, 554, 2, 15, 9, 553, 552, 11,
	550, 8, 549, 548, 16, 547, 546, 540, 539, 12,
	538, 6, 536, 1, 534, 24, 525, 14, 4, 23,
	272, 524, 520, 519, 517, 516, 515, 0, 19, 514,
	10, 512, 510, 508, 507, 505, 504, 119, 34, 497,
	495, 493, 492,
}
var yyR1 = []int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 74, 74, 5, 6, 7, 7, 71, 72,
	73, 76, 79, 79, 80, 80, 80, 81, 81, 75,
	75, 75, 75, 75, 8, 8, 8, 9, 9, 9,
	10, 11, 11, 11, 82, 12, 13, 13, 14, 14,
	14, 14, 14, 15, 15, 16, 16, 17, 17, 17,
	17, 20, 20, 18, 18, 18, 18, 21, 21, 22,
	22, 22, 22, 19, 19, 19, 23, 23, 23, 23,
	23, 23, 23, 23, 23, 24, 24, 24, 24, 24,
	24, 24, 25, 25, 26, 26, 26, 26, 27, 27,
	28, 28, 78, 78, 78, 77, 77, 29, 29, 29,
	29, 29, 30, 30, 30, 30, 30, 30, 30, 30,
	30, 30, 31, 31, 31, 31, 31, 31, 31, 32,
	32, 37, 37, 35, 35, 39, 36, 36, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 38, 38, 40, 40,
	40, 42, 45, 45, 43, 43, 44, 46, 46, 41,
	41, 33, 33, 33, 33, 47, 47, 48, 48, 49,
	49, 50, 50, 51, 52, 52, 52, 53, 53, 53,
	54, 54, 54, 55, 55, 56, 56, 57, 57, 58,
	58, 59, 60, 60, 61, 61, 62, 62, 63, 63,
	63, 63, 63, 64, 64, 65, 65, 66, 66, 67,
	68, 69, 69, 69, 69, 69, 70,
}
var yyR2 = []int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 4, 12, 3,
	7, 7, 6, 6, 8, 7, 3, 4, 1, 1,
	1, 5, 2, 2, 0, 2, 2, 0, 1, 3,
	3, 4, 5, 5, 5, 8, 4, 6, 7, 4,
	5, 4, 5, 5, 0, 2, 0, 2, 1, 2,
	1, 1, 1, 0, 1, 1, 3, 1, 2, 3,
	3, 1, 1, 0, 1, 2, 2, 1, 3, 3,
	3, 3, 5, 0, 1, 2, 1, 1, 2, 3,
	2, 3, 2, 2, 2, 1, 3, 1, 1, 1,
	3, 3, 1, 3, 0, 5, 5, 5, 1, 3,
	0, 2, 0, 2, 2, 0, 2, 1, 3, 3,
	2, 3, 3, 3, 4, 3, 4, 5, 6, 3,
	4, 2, 1, 1, 1, 1, 1, 1, 1, 2,
	1, 1, 3, 3, 1, 3, 1, 3, 1, 1,
	1, 3, 3, 3, 3, 3, 3, 3, 3, 2,
	3, 4, 5, 4, 4, 1, 1, 1, 1, 1,
	1, 5, 0, 1, 1, 2, 4, 0, 2, 1,
	3, 1, 1, 1, 1, 0, 3, 0, 2, 0,
	3, 1, 3, 2, 0, 1, 1, 0, 2, 4,
	0, 2, 4, 0, 3, 1, 3, 0, 5, 1,
	3, 3, 0, 2, 0, 3, 0, 1, 1, 1,
	1, 1, 1, 0, 1, 0, 1, 0, 2, 1,
	1, 1, 1, 1, 1, 1, 0,
}
var yyChk = []int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -71, -72, -73, -74, -75, -76, 5, 6,
	7, 8, 34, 102, 103, 105, 104, 89, 90, 91,
	93, 95, 94, -14, 55, 56, 57, 58, -12, -82,
	-12, -12, -12, -12, 106, -65, 108, 112, -62, 108,
	110, 106, 106, 107, 108, -12, 96, 99, 97, 98,
	-81, 100, -67, 36, -3, 17, -15, 18, -13, 30,
	-25, 36, 9, -58, 92, -59, -41, -67, 36, -61,
	111, 107, -67, 106, -67, 36, -60, 111, -67, -60,
	30, -78, 10, 24, -78, -77, 9, -67, 101, 45,
	-16, -17, 79, -20, 36, -29, -34, -30, 73, 45,
	-33, -41, -35, -40, -67, -38, -42, 21, 37, 38,
	39, 26, -39, 77, 78, 49, 111, 29, 84, 40,
	-25, 34, 82, -25, 59, 36, 46, 82, 36, 73,
	-67, -70, 36, -70, 109, 36, 21, 70, -67, -25,
	-20, -34, 45, -78, -34, -77, -79, 9, 22, -36,
	-34, 9, 59, -18, 71, 72, -67, 20, 82, -31,
	22, 73, 24, 25, 23, 74, 75, 76, 77, 78,
	79, 80, 81, 46, 47, 48, 41, 42, 43, 44,
	-20, -29, -36, -3, -68, -20, -34, -69, 50, 52,
	53, 51, 54, -34, 45, 45, -39, 45, -45, -34,
	-55, 34, 45, -58, 36, -28, 10, -59, -34, -67,
	-70, 21, -66, 113, -63, 105, 103, 33, 104, 13,
	36, 36, 36, -70, -55, 34, -78, -80, 9, 22,
	-34, -34, 114, 59, -21, -22, -24, 45, 36, -39,
	101, 97, -17, 19, -20, -20, -67, -68, 79, -34,
	-35, 45, 22, 24, 25, -34, -34, 26, 73, -34,
	-34, -34, -34, -34, -34, -34, -34, 114, 114, 114,
	-36, 114, -16, 18, -16, -43, -44, 85, -32, 29,
	-3, -58, -56, -41, -28, -49, 13, -20, 70, -67,
	-70, -64, 109, -32, -58, -34, -34, -34, -28, 59,
	-23, 60, 61, 62, 63, 64, 66, 67, -19, 36,
	20, -22, 82, -35, -34, -34, 71, 26, 114, 114,
	-16, 114, -46, -44, 87, -29, -57, 70, -37, -35,
	-57, 114, 59, -49, -53, 15, 14, 36, 36, -47,
	11, -22, -22, 60, 65, 60, 65, 60, 60, 60,
	-26, 68, 110, 69, 36, 114, 36, 101, 97, 71,
	-34, 114, 88, -34, 86, 31, 59, -41, -53, -34,
	-50, -51, -34, -70, -48, 12, 14, 70, 60, 60,
	107, 107, 107, -34, -34, 32, -35, 59, 59, -52,
	27, 28, -49, -20, -36, -29, 45, 45, 45, 7,
	-34, -51, -53, -27, -67, -27, -27, -58, -54, 16,
	35, 114, 59, 114, 114, 7, 22, -67, -67, -67,
}
var yyDef = []int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 225, 216, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 214, 0, 0, 226, 0, 0,
	217, 0, 212, 0, 212, 0, 112, 112, 115, 0,
	0, 38, 0, 229, 19, 59, 0, 64, 55, 0,
	0, 102, 0, 26, 0, 209, 0, 179, 229, 0,
	0, 0, 236, 0, 236, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 112, 0, 115, 0, 0,
	17, 65, 67, 73, 229, 71, 72, 117, 0, 0,
	148, 149, 150, 0, 179, 0, 165, 0, 181, 182,
	183, 184, 144, 168, 169, 170, 166, 167, 172, 57,
	203, 0, 0, 110, 0, 27, 0, 0, 236, 0,
	227, 46, 0, 49, 0, 51, 213, 0, 236, 203,
	113, 114, 0, 41, 116, 112, 34, 0, 0, 0,
	146, 0, 0, 68, 0, 0, 74, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 132, 133, 134, 135, 136, 137, 138,
	120, 71, 0, 0, 0, 0, -2, 230, 231, 232,
	233, 234, 235, 159, 0, 0, 131, 0, 0, 173,
	0, 0, 0, 110, 103, 189, 0, 210, 211, 180,
	44, 215, 0, 0, 236, 223, 218, 219, 220, 221,
	222, 50, 52, 53, 0, 0, 42, 43, 0, 0,
	32, 33, 31, 0, 110, 77, 83, 0, 95, 97,
	98, 99, 66, 69, 118, 119, 75, 76, 70, 122,
	123, 0, 0, 0, 0, 125, 0, 129, 0, 151,
	152, 153, 154, 155, 156, 157, 158, 121, 143, 145,
	0, 160, 0, 0, 0, 177, 174, 0, 207, 0,
	140, 207, 0, 205, 189, 197, 0, 111, 0, 228,
	47, 0, 224, 22, 23, 35, 36, 147, 185, 0,
	0, 86, 87, 0, 0, 0, 0, 0, 104, 84,
	0, 0, 0, 124, 126, 0, 0, 130, 163, 161,
	0, 164, 0, 175, 0, 71, 20, 0, 139, 141,
	21, 204, 0, 197, 25, 0, 0, 236, 48, 187,
	0, 78, 81, 88, 0, 90, 0, 92, 93, 94,
	79, 0, 0, 0, 85, 80, 96, 100, 101, 0,
	127, 162, 171, 178, 0, 0, 0, 206, 24, 198,
	190, 191, 194, 45, 189, 0, 0, 0, 89, 91,
	0, 0, 0, 128, 176, 0, 142, 0, 0, 193,
	195, 196, 197, 188, 186, -2, 0, 0, 0, 0,
	199, 192, 200, 0, 108, 0, 0, 208, 18, 0,
	0, 105, 0, 106, 107, 201, 0, 109, 0, 202,
}
var yyTok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 81, 74, 3,
	45, 114, 79, 77, 59, 78, 82, 80, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	47, 46, 48, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 76, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 75, 3, 49,
}
var yyTok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 50, 51, 52, 53, 54, 55, 56,
	57, 58, 60, 61, 62, 63, 64, 65, 66, 67,
	68, 69, 70, 71, 72, 73, 83, 84, 85, 86,
	87, 88, 89, 90, 91, 92, 93, 94, 95, 96,
	97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
	107, 108, 109, 110, 111, 112, 113,
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
			yyVAL.selectExpr = &NonStarExpr{Expr: yyS[yypt-2].expr, As: yyS[yypt-1].bytes}
		}
	case 70:
		//line sql.y:471
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyS[yypt-2].bytes}
		}
	case 71:
		//line sql.y:477
		{
			yyVAL.expr = yyS[yypt-0].boolExpr
		}
	case 72:
		//line sql.y:481
		{
			yyVAL.expr = yyS[yypt-0].valExpr
		}
	case 73:
		//line sql.y:486
		{
			yyVAL.bytes = nil
		}
	case 74:
		//line sql.y:490
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 75:
		//line sql.y:494
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 76:
		//line sql.y:498
		{
			yyVAL.bytes = []byte(yyS[yypt-0].str)
		}
	case 77:
		//line sql.y:505
		{
			yyVAL.tableExprs = TableExprs{yyS[yypt-0].tableExpr}
		}
	case 78:
		//line sql.y:509
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyS[yypt-0].tableExpr)
		}
	case 79:
		//line sql.y:515
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyS[yypt-2].smTableExpr, As: yyS[yypt-1].bytes, Hints: yyS[yypt-0].indexHints}
		}
	case 80:
		//line sql.y:519
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyS[yypt-1].tableExpr}
		}
	case 81:
		//line sql.y:523
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-2].tableExpr, Join: yyS[yypt-1].str, RightExpr: yyS[yypt-0].tableExpr}
		}
	case 82:
		//line sql.y:527
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-4].tableExpr, Join: yyS[yypt-3].str, RightExpr: yyS[yypt-2].tableExpr, On: yyS[yypt-0].boolExpr}
		}
	case 83:
		//line sql.y:532
		{
			yyVAL.bytes = nil
		}
	case 84:
		//line sql.y:536
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 85:
		//line sql.y:540
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 86:
		//line sql.y:546
		{
			yyVAL.str = AST_JOIN
		}
	case 87:
		//line sql.y:550
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 88:
		//line sql.y:554
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 89:
		//line sql.y:558
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 90:
		//line sql.y:562
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 91:
		//line sql.y:566
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 92:
		//line sql.y:570
		{
			yyVAL.str = AST_JOIN
		}
	case 93:
		//line sql.y:574
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 94:
		//line sql.y:578
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 95:
		//line sql.y:584
		{
			yyVAL.smTableExpr = &TableName{Name: yyS[yypt-0].bytes}
		}
	case 96:
		//line sql.y:588
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyS[yypt-2].bytes, Name: yyS[yypt-0].bytes}
		}
	case 97:
		//line sql.y:592
		{
			yyVAL.smTableExpr = yyS[yypt-0].subquery
		}
	case 98:
		//line sql.y:596
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 99:
		//line sql.y:600
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 100:
		//line sql.y:604
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyS[yypt-2].bytes, Name: []byte("columns")}
		}
	case 101:
		//line sql.y:608
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyS[yypt-2].bytes, Name: []byte("tables")}
		}
	case 102:
		//line sql.y:614
		{
			yyVAL.tableName = &TableName{Name: yyS[yypt-0].bytes}
		}
	case 103:
		//line sql.y:618
		{
			yyVAL.tableName = &TableName{Qualifier: yyS[yypt-2].bytes, Name: yyS[yypt-0].bytes}
		}
	case 104:
		//line sql.y:623
		{
			yyVAL.indexHints = nil
		}
	case 105:
		//line sql.y:627
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyS[yypt-1].bytes2}
		}
	case 106:
		//line sql.y:631
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyS[yypt-1].bytes2}
		}
	case 107:
		//line sql.y:635
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyS[yypt-1].bytes2}
		}
	case 108:
		//line sql.y:641
		{
			yyVAL.bytes2 = [][]byte{yyS[yypt-0].bytes}
		}
	case 109:
		//line sql.y:645
		{
			yyVAL.bytes2 = append(yyS[yypt-2].bytes2, yyS[yypt-0].bytes)
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
		//line sql.y:659
		{
			yyVAL.expr = nil
		}
	case 113:
		//line sql.y:663
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 114:
		//line sql.y:667
		{
			yyVAL.expr = yyS[yypt-0].valExpr
		}
	case 115:
		//line sql.y:672
		{
			yyVAL.valExpr = nil
		}
	case 116:
		//line sql.y:676
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 117:
		yyVAL.boolExpr = yyS[yypt-0].boolExpr
	case 118:
		//line sql.y:683
		{
			yyVAL.boolExpr = &AndExpr{Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 119:
		//line sql.y:687
		{
			yyVAL.boolExpr = &OrExpr{Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 120:
		//line sql.y:691
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyS[yypt-0].expr}
		}
	case 121:
		//line sql.y:695
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyS[yypt-1].boolExpr}
		}
	case 122:
		//line sql.y:701
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyS[yypt-2].valExpr, Operator: yyS[yypt-1].str, Right: yyS[yypt-0].valExpr}
		}
	case 123:
		//line sql.y:705
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyS[yypt-2].valExpr, Operator: AST_IN, Right: yyS[yypt-0].tuple}
		}
	case 124:
		//line sql.y:709
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyS[yypt-3].valExpr, Operator: AST_NOT_IN, Right: yyS[yypt-0].tuple}
		}
	case 125:
		//line sql.y:713
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyS[yypt-2].valExpr, Operator: AST_LIKE, Right: yyS[yypt-0].valExpr}
		}
	case 126:
		//line sql.y:717
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyS[yypt-3].valExpr, Operator: AST_NOT_LIKE, Right: yyS[yypt-0].valExpr}
		}
	case 127:
		//line sql.y:721
		{
			yyVAL.boolExpr = &RangeCond{Left: yyS[yypt-4].valExpr, Operator: AST_BETWEEN, From: yyS[yypt-2].valExpr, To: yyS[yypt-0].valExpr}
		}
	case 128:
		//line sql.y:725
		{
			yyVAL.boolExpr = &RangeCond{Left: yyS[yypt-5].valExpr, Operator: AST_NOT_BETWEEN, From: yyS[yypt-2].valExpr, To: yyS[yypt-0].valExpr}
		}
	case 129:
		//line sql.y:729
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NULL, Expr: yyS[yypt-2].valExpr}
		}
	case 130:
		//line sql.y:733
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NOT_NULL, Expr: yyS[yypt-3].valExpr}
		}
	case 131:
		//line sql.y:737
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyS[yypt-0].subquery}
		}
	case 132:
		//line sql.y:743
		{
			yyVAL.str = AST_EQ
		}
	case 133:
		//line sql.y:747
		{
			yyVAL.str = AST_LT
		}
	case 134:
		//line sql.y:751
		{
			yyVAL.str = AST_GT
		}
	case 135:
		//line sql.y:755
		{
			yyVAL.str = AST_LE
		}
	case 136:
		//line sql.y:759
		{
			yyVAL.str = AST_GE
		}
	case 137:
		//line sql.y:763
		{
			yyVAL.str = AST_NE
		}
	case 138:
		//line sql.y:767
		{
			yyVAL.str = AST_NSE
		}
	case 139:
		//line sql.y:773
		{
			yyVAL.insRows = yyS[yypt-0].values
		}
	case 140:
		//line sql.y:777
		{
			yyVAL.insRows = yyS[yypt-0].selStmt
		}
	case 141:
		//line sql.y:783
		{
			yyVAL.values = Values{yyS[yypt-0].tuple}
		}
	case 142:
		//line sql.y:787
		{
			yyVAL.values = append(yyS[yypt-2].values, yyS[yypt-0].tuple)
		}
	case 143:
		//line sql.y:793
		{
			yyVAL.tuple = ValTuple(yyS[yypt-1].valExprs)
		}
	case 144:
		//line sql.y:797
		{
			yyVAL.tuple = yyS[yypt-0].subquery
		}
	case 145:
		//line sql.y:803
		{
			yyVAL.subquery = &Subquery{yyS[yypt-1].selStmt}
		}
	case 146:
		//line sql.y:809
		{
			yyVAL.valExprs = ValExprs{yyS[yypt-0].valExpr}
		}
	case 147:
		//line sql.y:813
		{
			yyVAL.valExprs = append(yyS[yypt-2].valExprs, yyS[yypt-0].valExpr)
		}
	case 148:
		//line sql.y:819
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 149:
		//line sql.y:823
		{
			yyVAL.valExpr = yyS[yypt-0].colName
		}
	case 150:
		//line sql.y:827
		{
			yyVAL.valExpr = yyS[yypt-0].tuple
		}
	case 151:
		//line sql.y:831
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_BITAND, Right: yyS[yypt-0].valExpr}
		}
	case 152:
		//line sql.y:835
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_BITOR, Right: yyS[yypt-0].valExpr}
		}
	case 153:
		//line sql.y:839
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_BITXOR, Right: yyS[yypt-0].valExpr}
		}
	case 154:
		//line sql.y:843
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_PLUS, Right: yyS[yypt-0].valExpr}
		}
	case 155:
		//line sql.y:847
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_MINUS, Right: yyS[yypt-0].valExpr}
		}
	case 156:
		//line sql.y:851
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_MULT, Right: yyS[yypt-0].valExpr}
		}
	case 157:
		//line sql.y:855
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_DIV, Right: yyS[yypt-0].valExpr}
		}
	case 158:
		//line sql.y:859
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyS[yypt-2].valExpr, Operator: AST_MOD, Right: yyS[yypt-0].valExpr}
		}
	case 159:
		//line sql.y:863
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
	case 160:
		//line sql.y:878
		{
			yyVAL.valExpr = &FuncExpr{Name: yyS[yypt-2].bytes}
		}
	case 161:
		//line sql.y:882
		{
			yyVAL.valExpr = &FuncExpr{Name: yyS[yypt-3].bytes, Exprs: yyS[yypt-1].selectExprs}
		}
	case 162:
		//line sql.y:886
		{
			yyVAL.valExpr = &FuncExpr{Name: yyS[yypt-4].bytes, Distinct: true, Exprs: yyS[yypt-1].selectExprs}
		}
	case 163:
		//line sql.y:890
		{
			yyVAL.valExpr = &CtorExpr{Name: yyS[yypt-2].str, Exprs: yyS[yypt-1].valExprs}
		}
	case 164:
		//line sql.y:894
		{
			yyVAL.valExpr = &FuncExpr{Name: yyS[yypt-3].bytes, Exprs: yyS[yypt-1].selectExprs}
		}
	case 165:
		//line sql.y:898
		{
			yyVAL.valExpr = yyS[yypt-0].caseExpr
		}
	case 166:
		//line sql.y:904
		{
			yyVAL.bytes = IF_BYTES
		}
	case 167:
		//line sql.y:908
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 168:
		//line sql.y:914
		{
			yyVAL.byt = AST_UPLUS
		}
	case 169:
		//line sql.y:918
		{
			yyVAL.byt = AST_UMINUS
		}
	case 170:
		//line sql.y:922
		{
			yyVAL.byt = AST_TILDA
		}
	case 171:
		//line sql.y:928
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyS[yypt-3].valExpr, Whens: yyS[yypt-2].whens, Else: yyS[yypt-1].valExpr}
		}
	case 172:
		//line sql.y:933
		{
			yyVAL.valExpr = nil
		}
	case 173:
		//line sql.y:937
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 174:
		//line sql.y:943
		{
			yyVAL.whens = []*When{yyS[yypt-0].when}
		}
	case 175:
		//line sql.y:947
		{
			yyVAL.whens = append(yyS[yypt-1].whens, yyS[yypt-0].when)
		}
	case 176:
		//line sql.y:953
		{
			yyVAL.when = &When{Cond: yyS[yypt-2].boolExpr, Val: yyS[yypt-0].valExpr}
		}
	case 177:
		//line sql.y:958
		{
			yyVAL.valExpr = nil
		}
	case 178:
		//line sql.y:962
		{
			yyVAL.valExpr = yyS[yypt-0].valExpr
		}
	case 179:
		//line sql.y:968
		{
			yyVAL.colName = &ColName{Name: yyS[yypt-0].bytes}
		}
	case 180:
		//line sql.y:972
		{
			yyVAL.colName = &ColName{Qualifier: yyS[yypt-2].bytes, Name: yyS[yypt-0].bytes}
		}
	case 181:
		//line sql.y:978
		{
			yyVAL.valExpr = StrVal(yyS[yypt-0].bytes)
		}
	case 182:
		//line sql.y:982
		{
			yyVAL.valExpr = NumVal(yyS[yypt-0].bytes)
		}
	case 183:
		//line sql.y:986
		{
			yyVAL.valExpr = ValArg(yyS[yypt-0].bytes)
		}
	case 184:
		//line sql.y:990
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 185:
		//line sql.y:995
		{
			yyVAL.valExprs = nil
		}
	case 186:
		//line sql.y:999
		{
			yyVAL.valExprs = yyS[yypt-0].valExprs
		}
	case 187:
		//line sql.y:1004
		{
			yyVAL.expr = nil
		}
	case 188:
		//line sql.y:1008
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 189:
		//line sql.y:1013
		{
			yyVAL.orderBy = nil
		}
	case 190:
		//line sql.y:1017
		{
			yyVAL.orderBy = yyS[yypt-0].orderBy
		}
	case 191:
		//line sql.y:1023
		{
			yyVAL.orderBy = OrderBy{yyS[yypt-0].order}
		}
	case 192:
		//line sql.y:1027
		{
			yyVAL.orderBy = append(yyS[yypt-2].orderBy, yyS[yypt-0].order)
		}
	case 193:
		//line sql.y:1033
		{
			yyVAL.order = &Order{Expr: yyS[yypt-1].valExpr, Direction: yyS[yypt-0].str}
		}
	case 194:
		//line sql.y:1038
		{
			yyVAL.str = AST_ASC
		}
	case 195:
		//line sql.y:1042
		{
			yyVAL.str = AST_ASC
		}
	case 196:
		//line sql.y:1046
		{
			yyVAL.str = AST_DESC
		}
	case 197:
		//line sql.y:1051
		{
			yyVAL.limit = nil
		}
	case 198:
		//line sql.y:1055
		{
			yyVAL.limit = &Limit{Rowcount: yyS[yypt-0].valExpr}
		}
	case 199:
		//line sql.y:1059
		{
			yyVAL.limit = &Limit{Offset: yyS[yypt-2].valExpr, Rowcount: yyS[yypt-0].valExpr}
		}
	case 200:
		//line sql.y:1064
		{
			yyVAL.str = ""
		}
	case 201:
		//line sql.y:1068
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 202:
		//line sql.y:1072
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
	case 203:
		//line sql.y:1085
		{
			yyVAL.columns = nil
		}
	case 204:
		//line sql.y:1089
		{
			yyVAL.columns = yyS[yypt-1].columns
		}
	case 205:
		//line sql.y:1095
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyS[yypt-0].colName}}
		}
	case 206:
		//line sql.y:1099
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyS[yypt-0].colName})
		}
	case 207:
		//line sql.y:1104
		{
			yyVAL.updateExprs = nil
		}
	case 208:
		//line sql.y:1108
		{
			yyVAL.updateExprs = yyS[yypt-0].updateExprs
		}
	case 209:
		//line sql.y:1114
		{
			yyVAL.updateExprs = UpdateExprs{yyS[yypt-0].updateExpr}
		}
	case 210:
		//line sql.y:1118
		{
			yyVAL.updateExprs = append(yyS[yypt-2].updateExprs, yyS[yypt-0].updateExpr)
		}
	case 211:
		//line sql.y:1124
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyS[yypt-2].colName, Expr: yyS[yypt-0].valExpr}
		}
	case 212:
		//line sql.y:1129
		{
			yyVAL.empty = struct{}{}
		}
	case 213:
		//line sql.y:1131
		{
			yyVAL.empty = struct{}{}
		}
	case 214:
		//line sql.y:1134
		{
			yyVAL.empty = struct{}{}
		}
	case 215:
		//line sql.y:1136
		{
			yyVAL.empty = struct{}{}
		}
	case 216:
		//line sql.y:1139
		{
			yyVAL.empty = struct{}{}
		}
	case 217:
		//line sql.y:1141
		{
			yyVAL.empty = struct{}{}
		}
	case 218:
		//line sql.y:1145
		{
			yyVAL.empty = struct{}{}
		}
	case 219:
		//line sql.y:1147
		{
			yyVAL.empty = struct{}{}
		}
	case 220:
		//line sql.y:1149
		{
			yyVAL.empty = struct{}{}
		}
	case 221:
		//line sql.y:1151
		{
			yyVAL.empty = struct{}{}
		}
	case 222:
		//line sql.y:1153
		{
			yyVAL.empty = struct{}{}
		}
	case 223:
		//line sql.y:1156
		{
			yyVAL.empty = struct{}{}
		}
	case 224:
		//line sql.y:1158
		{
			yyVAL.empty = struct{}{}
		}
	case 225:
		//line sql.y:1161
		{
			yyVAL.empty = struct{}{}
		}
	case 226:
		//line sql.y:1163
		{
			yyVAL.empty = struct{}{}
		}
	case 227:
		//line sql.y:1166
		{
			yyVAL.empty = struct{}{}
		}
	case 228:
		//line sql.y:1168
		{
			yyVAL.empty = struct{}{}
		}
	case 229:
		//line sql.y:1172
		{
			yyVAL.bytes = bytes.ToLower(yyS[yypt-0].bytes)
		}
	case 230:
		//line sql.y:1179
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 231:
		//line sql.y:1185
		{
			yyVAL.str = AST_DATE
		}
	case 232:
		//line sql.y:1189
		{
			yyVAL.str = AST_TIME
		}
	case 233:
		//line sql.y:1193
		{
			yyVAL.str = AST_TIMESTAMP
		}
	case 234:
		//line sql.y:1197
		{
			yyVAL.str = AST_DATETIME
		}
	case 235:
		//line sql.y:1201
		{
			yyVAL.str = AST_YEAR
		}
	case 236:
		//line sql.y:1206
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
