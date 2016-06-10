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
const SOME = 57359
const ANY = 57360
const TRUE = 57361
const FALSE = 57362
const ALL = 57363
const DISTINCT = 57364
const PRECISION = 57365
const AS = 57366
const EXISTS = 57367
const IN = 57368
const IS = 57369
const LIKE = 57370
const BETWEEN = 57371
const NULL = 57372
const ASC = 57373
const DESC = 57374
const VALUES = 57375
const INTO = 57376
const DUPLICATE = 57377
const KEY = 57378
const DEFAULT = 57379
const SET = 57380
const LOCK = 57381
const ID = 57382
const STRING = 57383
const NUMBER = 57384
const VALUE_ARG = 57385
const COMMENT = 57386
const LE = 57387
const GE = 57388
const NE = 57389
const NULL_SAFE_EQUAL = 57390
const DATE = 57391
const DATETIME = 57392
const TIME = 57393
const TIMESTAMP = 57394
const YEAR = 57395
const UNION = 57396
const MINUS = 57397
const EXCEPT = 57398
const INTERSECT = 57399
const JOIN = 57400
const STRAIGHT_JOIN = 57401
const LEFT = 57402
const RIGHT = 57403
const INNER = 57404
const OUTER = 57405
const CROSS = 57406
const NATURAL = 57407
const USE = 57408
const FORCE = 57409
const ON = 57410
const AND = 57411
const OR = 57412
const NOT = 57413
const UNARY = 57414
const CASE = 57415
const WHEN = 57416
const THEN = 57417
const ELSE = 57418
const END = 57419
const BEGIN = 57420
const COMMIT = 57421
const ROLLBACK = 57422
const NAMES = 57423
const REPLACE = 57424
const ADMIN = 57425
const SHOW = 57426
const DATABASES = 57427
const TABLES = 57428
const PROXY = 57429
const VARIABLES = 57430
const FULL = 57431
const COLUMNS = 57432
const CREATE = 57433
const ALTER = 57434
const DROP = 57435
const RENAME = 57436
const TABLE = 57437
const INDEX = 57438
const VIEW = 57439
const TO = 57440
const IGNORE = 57441
const IF = 57442
const UNIQUE = 57443
const USING = 57444

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
	-1, 200,
	75, 72,
	76, 72,
	-2, 149,
	-1, 424,
	75, 71,
	76, 71,
	-2, 82,
}

const yyNprod = 245
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 877

var yyAct = [...]int{

	114, 361, 106, 432, 73, 103, 399, 105, 111, 196,
	145, 126, 306, 251, 353, 112, 297, 221, 299, 197,
	3, 198, 100, 75, 101, 441, 441, 216, 34, 35,
	36, 37, 286, 62, 91, 322, 323, 324, 325, 326,
	229, 327, 328, 87, 77, 441, 166, 82, 359, 80,
	84, 166, 76, 64, 88, 166, 166, 49, 166, 50,
	97, 235, 313, 249, 249, 249, 44, 148, 46, 52,
	53, 54, 47, 383, 391, 410, 409, 408, 81, 98,
	443, 442, 144, 378, 380, 233, 83, 288, 236, 382,
	152, 51, 94, 78, 95, 147, 155, 389, 154, 158,
	440, 388, 164, 358, 170, 298, 347, 350, 298, 163,
	346, 345, 200, 343, 194, 199, 207, 195, 342, 287,
	248, 184, 185, 186, 63, 379, 333, 172, 141, 136,
	157, 212, 168, 169, 385, 215, 77, 143, 384, 77,
	219, 254, 225, 224, 76, 405, 354, 76, 276, 74,
	253, 309, 151, 226, 407, 232, 234, 231, 138, 164,
	406, 372, 223, 239, 246, 247, 373, 264, 376, 354,
	18, 375, 262, 225, 260, 261, 265, 255, 374, 274,
	275, 240, 278, 279, 280, 281, 282, 283, 284, 285,
	269, 258, 159, 263, 242, 277, 18, 19, 20, 21,
	138, 164, 257, 249, 70, 254, 256, 417, 289, 56,
	58, 59, 57, 61, 253, 370, 222, 165, 77, 77,
	371, 394, 302, 140, 241, 427, 76, 304, 308, 22,
	310, 291, 293, 294, 295, 218, 301, 305, 222, 86,
	414, 311, 77, 426, 425, 270, 315, 316, 317, 213,
	76, 133, 318, 182, 183, 184, 185, 186, 211, 314,
	301, 34, 35, 36, 37, 255, 257, 332, 319, 320,
	256, 166, 217, 164, 134, 338, 339, 137, 334, 335,
	336, 210, 209, 218, 27, 28, 29, 337, 30, 32,
	31, 138, 208, 63, 89, 153, 99, 23, 24, 26,
	25, 352, 63, 171, 199, 331, 351, 202, 205, 203,
	204, 206, 78, 349, 438, 344, 356, 357, 360, 63,
	381, 330, 365, 364, 238, 237, 174, 178, 176, 177,
	220, 71, 255, 255, 368, 369, 149, 439, 393, 146,
	90, 142, 139, 387, 85, 190, 191, 192, 193, 135,
	187, 188, 189, 390, 168, 169, 69, 341, 18, 244,
	77, 271, 396, 272, 273, 397, 400, 445, 395, 92,
	227, 161, 150, 65, 259, 401, 245, 175, 179, 180,
	181, 182, 183, 184, 185, 186, 300, 93, 162, 411,
	392, 67, 362, 386, 412, 413, 179, 180, 181, 182,
	183, 184, 185, 186, 404, 363, 307, 164, 403, 422,
	415, 199, 367, 424, 423, 421, 222, 96, 72, 429,
	400, 444, 428, 431, 430, 18, 433, 433, 433, 77,
	434, 435, 39, 436, 60, 243, 160, 76, 120, 121,
	17, 292, 446, 16, 119, 15, 447, 14, 448, 125,
	13, 12, 131, 38, 201, 228, 45, 312, 230, 104,
	122, 123, 124, 322, 323, 324, 325, 326, 109, 327,
	328, 48, 129, 40, 41, 42, 43, 79, 18, 303,
	437, 418, 398, 402, 55, 115, 116, 366, 348, 214,
	296, 118, 120, 121, 113, 117, 108, 355, 119, 110,
	127, 128, 102, 125, 173, 107, 131, 132, 377, 252,
	321, 250, 329, 78, 122, 123, 124, 167, 66, 33,
	68, 18, 109, 11, 10, 9, 129, 202, 205, 203,
	204, 206, 8, 7, 130, 120, 121, 290, 6, 115,
	116, 5, 4, 2, 1, 0, 125, 0, 0, 131,
	108, 0, 0, 0, 127, 128, 78, 122, 123, 124,
	0, 132, 0, 0, 0, 156, 419, 420, 0, 129,
	202, 205, 203, 204, 206, 0, 0, 0, 120, 121,
	0, 0, 115, 116, 119, 0, 0, 0, 130, 125,
	0, 0, 131, 0, 0, 0, 0, 127, 128, 104,
	122, 123, 124, 0, 132, 0, 0, 0, 109, 0,
	0, 0, 129, 179, 180, 181, 182, 183, 184, 185,
	186, 0, 0, 0, 0, 115, 116, 0, 0, 0,
	0, 130, 0, 0, 0, 0, 108, 0, 0, 0,
	127, 128, 102, 0, 0, 0, 0, 132, 0, 0,
	0, 0, 268, 267, 120, 121, 266, 179, 180, 181,
	182, 183, 184, 185, 186, 125, 0, 0, 131, 0,
	0, 0, 0, 0, 130, 78, 122, 123, 124, 0,
	0, 0, 0, 0, 156, 120, 121, 0, 129, 0,
	0, 119, 0, 0, 0, 0, 125, 0, 0, 131,
	0, 115, 116, 0, 0, 0, 78, 122, 123, 124,
	0, 0, 0, 0, 0, 109, 127, 128, 0, 129,
	0, 18, 0, 132, 0, 0, 0, 0, 0, 0,
	0, 0, 115, 116, 0, 120, 121, 0, 0, 0,
	0, 0, 0, 108, 0, 0, 125, 127, 128, 131,
	130, 0, 0, 0, 132, 0, 78, 122, 123, 124,
	0, 0, 0, 0, 0, 156, 120, 121, 0, 129,
	0, 0, 0, 0, 0, 0, 0, 125, 0, 0,
	131, 130, 115, 116, 0, 0, 0, 78, 122, 123,
	124, 0, 0, 0, 0, 0, 156, 127, 128, 0,
	129, 0, 340, 0, 132, 179, 180, 181, 182, 183,
	184, 185, 186, 115, 116, 0, 0, 174, 178, 176,
	177, 0, 0, 0, 0, 0, 0, 0, 127, 128,
	416, 130, 0, 0, 0, 132, 190, 191, 192, 193,
	0, 187, 188, 189, 0, 179, 180, 181, 182, 183,
	184, 185, 186, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 130, 0, 0, 0, 0, 0, 175, 179,
	180, 181, 182, 183, 184, 185, 186,
}
var yyPact = [...]int{

	191, -1000, -1000, 202, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -44, -55, -19, -41, -1000, -1000, -1000,
	-1000, 109, 262, 420, 352, -1000, -1000, -1000, 369, -1000,
	322, 291, 409, 53, -66, -33, 262, -1000, -24, 262,
	-1000, 304, -72, 262, -72, 306, 359, 359, 408, 262,
	-26, -1000, 247, -1000, -1000, -1000, 559, -1000, 207, 291,
	311, 43, 291, 137, 302, -1000, 173, -1000, 42, 301,
	60, 262, -1000, 299, -1000, -46, 296, 347, 78, 262,
	291, -1000, 666, 747, -1000, 359, 747, 408, 362, 747,
	208, -1000, -1000, 279, 41, -1000, 791, -1000, 666, 473,
	-1000, -1000, -1000, 747, 243, 233, 232, 209, -1000, 200,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 747, -1000, 234, 272, 290, 406, 272, -1000,
	747, 262, -1000, 345, -77, -1000, 48, -1000, 285, -1000,
	-1000, 284, -1000, 186, 57, 579, 516, -1000, 579, 359,
	350, 747, 747, 2, 579, 101, 559, 351, 666, 666,
	-1000, 253, 84, 635, 196, 335, 747, 747, 118, 747,
	747, 747, 747, 747, 747, 747, 747, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -86, 1, -31, 747, 57,
	791, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 419, 559,
	559, 559, -1000, 420, 19, 579, 353, 272, 272, 228,
	-1000, 393, 666, -1000, 579, -1000, -1000, -1000, 77, 262,
	-1000, -51, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	353, 272, -1000, -1000, 747, 747, 579, 579, -1000, 747,
	206, 399, 281, 165, 40, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, 579, 200, 200, 200, -1000,
	716, 196, 747, 747, 579, 727, -1000, 327, 172, 172,
	172, 38, 38, -1000, -1000, -1000, -1000, -1000, -1000, 0,
	-1000, -5, 559, -7, -8, -12, 16, -1000, 666, 72,
	196, 202, 95, -15, -1000, 393, 377, 391, 57, 283,
	-1000, -1000, 282, -1000, -1000, 137, 579, 579, 579, 401,
	101, 101, -1000, -1000, 151, 97, 114, 107, 104, 11,
	-1000, 280, -29, 33, -1000, -1000, -1000, -1000, 579, 318,
	747, -1000, -1000, -1000, -17, -1000, -1000, -1000, 5, -1000,
	747, -16, 300, -1000, 303, 158, -1000, -1000, -1000, 272,
	377, -1000, 747, 747, -1000, -1000, 396, 390, 399, 71,
	-1000, 96, -1000, 90, -1000, -1000, -1000, -1000, -34, -35,
	-36, -1000, -1000, -1000, -1000, -1000, 747, 579, -1000, -1000,
	579, 747, 747, 204, 196, -1000, -1000, 767, 144, -1000,
	535, -1000, 393, 666, 747, 666, -1000, -1000, 195, 194,
	176, 579, 579, 579, 415, -1000, 747, 747, -1000, -1000,
	-1000, 377, 57, 140, -1000, 262, 262, 262, 272, 579,
	-1000, 298, -18, -1000, -37, -38, 137, -1000, 414, 341,
	-1000, 262, -1000, -1000, -1000, 262, -1000, 262, -1000,
}
var yyPgo = [...]int{

	0, 544, 543, 19, 542, 541, 538, 533, 532, 525,
	524, 523, 453, 520, 519, 518, 22, 24, 517, 512,
	5, 511, 13, 510, 509, 204, 508, 3, 17, 7,
	505, 504, 18, 499, 2, 15, 9, 497, 495, 11,
	494, 8, 491, 490, 16, 489, 488, 487, 483, 12,
	482, 6, 481, 1, 480, 27, 479, 14, 4, 23,
	239, 477, 471, 458, 457, 456, 455, 0, 21, 454,
	10, 451, 450, 447, 445, 443, 440, 94, 34, 436,
	435, 434, 432,
}
var yyR1 = [...]int{

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
	30, 30, 30, 30, 30, 31, 31, 31, 31, 31,
	31, 31, 32, 32, 37, 37, 35, 35, 39, 36,
	36, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 38, 38, 40, 40, 40, 42, 45, 45, 43,
	43, 44, 44, 46, 46, 41, 41, 33, 33, 33,
	33, 33, 33, 47, 47, 48, 48, 49, 49, 50,
	50, 51, 52, 52, 52, 53, 53, 53, 54, 54,
	54, 55, 55, 56, 56, 57, 57, 58, 58, 59,
	60, 60, 61, 61, 62, 62, 63, 63, 63, 63,
	63, 64, 64, 65, 65, 66, 66, 67, 68, 69,
	69, 69, 69, 69, 70,
}
var yyR2 = [...]int{

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
	2, 3, 3, 3, 4, 4, 4, 4, 3, 4,
	5, 6, 3, 4, 2, 1, 1, 1, 1, 1,
	1, 1, 2, 1, 1, 3, 3, 1, 3, 1,
	3, 1, 1, 1, 3, 3, 3, 3, 3, 3,
	3, 3, 2, 3, 4, 4, 4, 5, 4, 4,
	1, 1, 1, 1, 1, 1, 5, 0, 1, 1,
	2, 4, 4, 0, 2, 1, 3, 1, 1, 1,
	1, 1, 1, 0, 3, 0, 2, 0, 3, 1,
	3, 2, 0, 1, 1, 0, 2, 4, 0, 2,
	4, 0, 3, 1, 3, 0, 5, 1, 3, 3,
	0, 2, 0, 3, 0, 1, 1, 1, 1, 1,
	1, 0, 1, 0, 1, 0, 2, 1, 1, 1,
	1, 1, 1, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -71, -72, -73, -74, -75, -76, 5, 6,
	7, 8, 38, 106, 107, 109, 108, 93, 94, 95,
	97, 99, 98, -14, 59, 60, 61, 62, -12, -82,
	-12, -12, -12, -12, 110, -65, 112, 116, -62, 112,
	114, 110, 110, 111, 112, -12, 100, 103, 101, 102,
	-81, 104, -67, 40, -3, 21, -15, 22, -13, 34,
	-25, 40, 9, -58, 96, -59, -41, -67, 40, -61,
	115, 111, -67, 110, -67, 40, -60, 115, -67, -60,
	34, -78, 10, 28, -78, -77, 9, -67, 105, 49,
	-16, -17, 83, -20, 40, -29, -34, -30, 77, 49,
	-33, -41, -35, -40, -67, 66, 67, -38, -42, 25,
	19, 20, 41, 42, 43, 30, -39, 81, 82, 53,
	115, 33, 88, 44, -25, 38, 86, -25, 63, 40,
	50, 86, 40, 77, -67, -70, 40, -70, 113, 40,
	25, 74, -67, -25, -20, -34, 49, -78, -34, -77,
	-79, 9, 26, -36, -34, 9, 63, -18, 75, 76,
	-67, 24, 86, -31, 26, 77, 28, 29, 27, 78,
	79, 80, 81, 82, 83, 84, 85, 50, 51, 52,
	45, 46, 47, 48, -20, -29, -36, -3, -68, -20,
	-34, -69, 54, 56, 57, 55, 58, -34, 49, 49,
	49, 49, -39, 49, -45, -34, -55, 38, 49, -58,
	40, -28, 10, -59, -34, -67, -70, 25, -66, 117,
	-63, 109, 107, 37, 108, 13, 40, 40, 40, -70,
	-55, 38, -78, -80, 9, 26, -34, -34, 118, 63,
	-21, -22, -24, 49, 40, -39, 105, 101, -17, 23,
	-20, -20, -67, -68, 83, -34, 21, 18, 17, -35,
	49, 26, 28, 29, -34, -34, 30, 77, -34, -34,
	-34, -34, -34, -34, -34, -34, 118, 118, 118, -36,
	118, -16, 22, -16, -16, -16, -43, -44, 89, -32,
	33, -3, -58, -56, -41, -28, -49, 13, -20, 74,
	-67, -70, -64, 113, -32, -58, -34, -34, -34, -28,
	63, -23, 64, 65, 66, 67, 68, 70, 71, -19,
	40, 24, -22, 86, -39, -39, -39, -35, -34, -34,
	75, 30, 118, 118, -16, 118, 118, 118, -46, -44,
	91, -29, -34, -57, 74, -37, -35, -57, 118, 63,
	-49, -53, 15, 14, 40, 40, -47, 11, -22, -22,
	64, 69, 64, 69, 64, 64, 64, -26, 72, 114,
	73, 40, 118, 40, 105, 101, 75, -34, 118, 92,
	-34, 90, 90, 35, 63, -41, -53, -34, -50, -51,
	-34, -70, -48, 12, 14, 74, 64, 64, 111, 111,
	111, -34, -34, -34, 36, -35, 63, 63, -52, 31,
	32, -49, -20, -36, -29, 49, 49, 49, 7, -34,
	-51, -53, -27, -67, -27, -27, -58, -54, 16, 39,
	118, 63, 118, 118, 7, 26, -67, -67, -67,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 233, 224, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 222, 0, 0, 234, 0, 0,
	225, 0, 220, 0, 220, 0, 112, 112, 115, 0,
	0, 38, 0, 237, 19, 59, 0, 64, 55, 0,
	0, 102, 0, 26, 0, 217, 0, 185, 237, 0,
	0, 0, 244, 0, 244, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 112, 0, 115, 0, 0,
	17, 65, 67, 73, 237, 71, 72, 117, 0, 0,
	151, 152, 153, 0, 185, 0, 0, 0, 170, 0,
	187, 188, 189, 190, 191, 192, 147, 173, 174, 175,
	171, 172, 177, 57, 211, 0, 0, 110, 0, 27,
	0, 0, 244, 0, 235, 46, 0, 49, 0, 51,
	221, 0, 244, 211, 113, 114, 0, 41, 116, 112,
	34, 0, 0, 0, 149, 0, 0, 68, 0, 0,
	74, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 135, 136, 137,
	138, 139, 140, 141, 120, 71, 0, 0, 0, 0,
	-2, 238, 239, 240, 241, 242, 243, 162, 0, 0,
	0, 0, 134, 0, 0, 178, 0, 0, 0, 110,
	103, 197, 0, 218, 219, 186, 44, 223, 0, 0,
	244, 231, 226, 227, 228, 229, 230, 50, 52, 53,
	0, 0, 42, 43, 0, 0, 32, 33, 31, 0,
	110, 77, 83, 0, 95, 97, 98, 99, 66, 69,
	118, 119, 75, 76, 70, 122, 0, 0, 0, 123,
	0, 0, 0, 0, 128, 0, 132, 0, 154, 155,
	156, 157, 158, 159, 160, 161, 121, 146, 148, 0,
	163, 0, 0, 0, 0, 0, 183, 179, 0, 215,
	0, 143, 215, 0, 213, 197, 205, 0, 111, 0,
	236, 47, 0, 232, 22, 23, 35, 36, 150, 193,
	0, 0, 86, 87, 0, 0, 0, 0, 0, 104,
	84, 0, 0, 0, 124, 125, 126, 127, 129, 0,
	0, 133, 168, 164, 0, 165, 166, 169, 0, 180,
	0, 71, 72, 20, 0, 142, 144, 21, 212, 0,
	205, 25, 0, 0, 244, 48, 195, 0, 78, 81,
	88, 0, 90, 0, 92, 93, 94, 79, 0, 0,
	0, 85, 80, 96, 100, 101, 0, 130, 167, 176,
	184, 0, 0, 0, 0, 214, 24, 206, 198, 199,
	202, 45, 197, 0, 0, 0, 89, 91, 0, 0,
	0, 131, 181, 182, 0, 145, 0, 0, 201, 203,
	204, 205, 196, 194, -2, 0, 0, 0, 0, 207,
	200, 208, 0, 108, 0, 0, 216, 18, 0, 0,
	105, 0, 106, 107, 209, 0, 109, 0, 210,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 85, 78, 3,
	49, 118, 83, 81, 63, 82, 86, 84, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	51, 50, 52, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 80, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 79, 3, 53,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 47, 48, 54, 55, 56,
	57, 58, 59, 60, 61, 62, 64, 65, 66, 67,
	68, 69, 70, 71, 72, 73, 74, 75, 76, 77,
	87, 88, 89, 90, 91, 92, 93, 94, 95, 96,
	97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
	107, 108, 109, 110, 111, 112, 113, 114, 115, 116,
	117,
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
		//line sql.y:173
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:179
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:199
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 18:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:203
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].valExprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:207
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 20:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:214
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:218
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
		//line sql.y:230
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:234
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
		//line sql.y:247
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 25:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:253
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:259
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:263
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:269
		{
			yyVAL.statement = &Begin{}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:275
		{
			yyVAL.statement = &Commit{}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:281
		{
			yyVAL.statement = &Rollback{}
		}
	case 31:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:287
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].valExprs}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:293
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:297
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 34:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:302
		{
			yyVAL.valExpr = nil
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:306
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:310
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 37:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:315
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:319
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 39:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:326
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:330
		{
			yyVAL.statement = &Show{Section: "variables", LikeOrWhere: yyDollar[3].expr}
		}
	case 41:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:334
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].valExpr, LikeOrWhere: yyDollar[4].expr}
		}
	case 42:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:338
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].valExpr, LikeOrWhere: yyDollar[5].expr}
		}
	case 43:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:342
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].valExpr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].valExpr}
		}
	case 44:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:348
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 45:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:352
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:357
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 47:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:363
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 48:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:367
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 49:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:372
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 50:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:378
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 51:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:384
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 52:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:388
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 53:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:393
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 54:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:398
		{
			SetAllowComments(yylex, true)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:402
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 56:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:408
		{
			yyVAL.bytes2 = nil
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:412
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:418
		{
			yyVAL.str = AST_UNION
		}
	case 59:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:422
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:426
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:430
		{
			yyVAL.str = AST_EXCEPT
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:434
		{
			yyVAL.str = AST_INTERSECT
		}
	case 63:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:439
		{
			yyVAL.str = ""
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:443
		{
			yyVAL.str = AST_DISTINCT
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:449
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:453
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:459
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:463
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:467
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 70:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:471
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:477
		{
			yyVAL.expr = yyDollar[1].boolExpr
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:481
		{
			yyVAL.expr = yyDollar[1].valExpr
		}
	case 73:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:486
		{
			yyVAL.bytes = nil
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:490
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:494
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 76:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:498
		{
			yyVAL.bytes = []byte(yyDollar[2].str)
		}
	case 77:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:505
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:509
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:515
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:519
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:523
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 82:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:527
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].boolExpr}
		}
	case 83:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:532
		{
			yyVAL.bytes = nil
		}
	case 84:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:536
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 85:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:540
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:546
		{
			yyVAL.str = AST_JOIN
		}
	case 87:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:550
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 88:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:554
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:558
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 90:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:562
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 91:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:566
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:570
		{
			yyVAL.str = AST_JOIN
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:574
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 94:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:578
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 95:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:584
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:588
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:592
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:596
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 99:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:600
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 100:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:604
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:608
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 102:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:614
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 103:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:618
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 104:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:623
		{
			yyVAL.indexHints = nil
		}
	case 105:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:627
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 106:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:631
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 107:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:635
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 108:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:641
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 109:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:645
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 110:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:650
		{
			yyVAL.expr = nil
		}
	case 111:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:654
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 112:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:659
		{
			yyVAL.expr = nil
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:663
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 114:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:667
		{
			yyVAL.expr = yyDollar[2].valExpr
		}
	case 115:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:672
		{
			yyVAL.valExpr = nil
		}
	case 116:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:676
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 118:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:683
		{
			yyVAL.boolExpr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 119:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:687
		{
			yyVAL.boolExpr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 120:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:691
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 121:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:695
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyDollar[2].boolExpr}
		}
	case 122:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:701
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, Right: yyDollar[3].valExpr}
		}
	case 123:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:705
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 124:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:709
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ALL, Right: yyDollar[4].subquery}
		}
	case 125:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:713
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ANY, Right: yyDollar[4].subquery}
		}
	case 126:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:717
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_SOME, Right: yyDollar[4].subquery}
		}
	case 127:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:721
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 128:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:725
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_LIKE, Right: yyDollar[3].valExpr}
		}
	case 129:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:729
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_LIKE, Right: yyDollar[4].valExpr}
		}
	case 130:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:733
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_BETWEEN, From: yyDollar[3].valExpr, To: yyDollar[5].valExpr}
		}
	case 131:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:737
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].valExpr, To: yyDollar[6].valExpr}
		}
	case 132:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:741
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NULL, Expr: yyDollar[1].valExpr}
		}
	case 133:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:745
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NOT_NULL, Expr: yyDollar[1].valExpr}
		}
	case 134:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:749
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:755
		{
			yyVAL.str = AST_EQ
		}
	case 136:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:759
		{
			yyVAL.str = AST_LT
		}
	case 137:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:763
		{
			yyVAL.str = AST_GT
		}
	case 138:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:767
		{
			yyVAL.str = AST_LE
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:771
		{
			yyVAL.str = AST_GE
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:775
		{
			yyVAL.str = AST_NE
		}
	case 141:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:779
		{
			yyVAL.str = AST_NSE
		}
	case 142:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:785
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:789
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 144:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:795
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:799
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 146:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:805
		{
			yyVAL.tuple = ValTuple(yyDollar[2].valExprs)
		}
	case 147:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:809
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 148:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:815
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 149:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:821
		{
			yyVAL.valExprs = ValExprs{yyDollar[1].valExpr}
		}
	case 150:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:825
		{
			yyVAL.valExprs = append(yyDollar[1].valExprs, yyDollar[3].valExpr)
		}
	case 151:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:831
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 152:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:835
		{
			yyVAL.valExpr = yyDollar[1].colName
		}
	case 153:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:839
		{
			yyVAL.valExpr = yyDollar[1].tuple
		}
	case 154:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:843
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITAND, Right: yyDollar[3].valExpr}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:847
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITOR, Right: yyDollar[3].valExpr}
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:851
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITXOR, Right: yyDollar[3].valExpr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:855
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_PLUS, Right: yyDollar[3].valExpr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:859
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MINUS, Right: yyDollar[3].valExpr}
		}
	case 159:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:863
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MULT, Right: yyDollar[3].valExpr}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:867
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_DIV, Right: yyDollar[3].valExpr}
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:871
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 162:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:875
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
	case 163:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:890
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 164:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:894
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 165:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:898
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("left"), Exprs: yyDollar[3].selectExprs}
		}
	case 166:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:902
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("right"), Exprs: yyDollar[3].selectExprs}
		}
	case 167:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:906
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 168:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:910
		{
			yyVAL.valExpr = &CtorExpr{Name: yyDollar[2].str, Exprs: yyDollar[3].valExprs}
		}
	case 169:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:914
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 170:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:918
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 171:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:924
		{
			yyVAL.bytes = IF_BYTES
		}
	case 172:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:928
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 173:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:934
		{
			yyVAL.byt = AST_UPLUS
		}
	case 174:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:938
		{
			yyVAL.byt = AST_UMINUS
		}
	case 175:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:942
		{
			yyVAL.byt = AST_TILDA
		}
	case 176:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:948
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 177:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:953
		{
			yyVAL.valExpr = nil
		}
	case 178:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:957
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 179:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:963
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 180:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:967
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 181:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:973
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 182:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:977
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 183:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:982
		{
			yyVAL.valExpr = nil
		}
	case 184:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:986
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 185:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:992
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 186:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:996
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 187:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1002
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1006
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1010
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1014
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1018
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1022
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 193:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1027
		{
			yyVAL.valExprs = nil
		}
	case 194:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1031
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 195:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1036
		{
			yyVAL.expr = nil
		}
	case 196:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1040
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 197:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1045
		{
			yyVAL.orderBy = nil
		}
	case 198:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1049
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 199:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1055
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 200:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1059
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 201:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1065
		{
			yyVAL.order = &Order{Expr: yyDollar[1].valExpr, Direction: yyDollar[2].str}
		}
	case 202:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1070
		{
			yyVAL.str = AST_ASC
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1074
		{
			yyVAL.str = AST_ASC
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1078
		{
			yyVAL.str = AST_DESC
		}
	case 205:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1083
		{
			yyVAL.limit = nil
		}
	case 206:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1087
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 207:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1091
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 208:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1096
		{
			yyVAL.str = ""
		}
	case 209:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1100
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 210:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1104
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
	case 211:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1117
		{
			yyVAL.columns = nil
		}
	case 212:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1121
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1127
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 214:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1131
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 215:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1136
		{
			yyVAL.updateExprs = nil
		}
	case 216:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1140
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1146
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 218:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1150
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 219:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1156
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 220:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1161
		{
			yyVAL.empty = struct{}{}
		}
	case 221:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1163
		{
			yyVAL.empty = struct{}{}
		}
	case 222:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1166
		{
			yyVAL.empty = struct{}{}
		}
	case 223:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1168
		{
			yyVAL.empty = struct{}{}
		}
	case 224:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1171
		{
			yyVAL.empty = struct{}{}
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1173
		{
			yyVAL.empty = struct{}{}
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1177
		{
			yyVAL.empty = struct{}{}
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1179
		{
			yyVAL.empty = struct{}{}
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1181
		{
			yyVAL.empty = struct{}{}
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1183
		{
			yyVAL.empty = struct{}{}
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1185
		{
			yyVAL.empty = struct{}{}
		}
	case 231:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1188
		{
			yyVAL.empty = struct{}{}
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1190
		{
			yyVAL.empty = struct{}{}
		}
	case 233:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1193
		{
			yyVAL.empty = struct{}{}
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1195
		{
			yyVAL.empty = struct{}{}
		}
	case 235:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1198
		{
			yyVAL.empty = struct{}{}
		}
	case 236:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1200
		{
			yyVAL.empty = struct{}{}
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1204
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1210
		{
			yyVAL.str = yyDollar[1].str
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1216
		{
			yyVAL.str = AST_DATE
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1220
		{
			yyVAL.str = AST_TIME
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1224
		{
			yyVAL.str = AST_TIMESTAMP
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1228
		{
			yyVAL.str = AST_DATETIME
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1232
		{
			yyVAL.str = AST_YEAR
		}
	case 244:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1237
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
