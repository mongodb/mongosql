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
	RIGHT_BYTES  = []byte("right")
	LEFT_BYTES   = []byte("left")
	MOD_BYTES    = []byte("mod")
)

//line sql.y:34
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
const MOD = 57414
const DIV = 57415
const UNARY = 57416
const CASE = 57417
const WHEN = 57418
const THEN = 57419
const ELSE = 57420
const END = 57421
const BEGIN = 57422
const COMMIT = 57423
const ROLLBACK = 57424
const NAMES = 57425
const REPLACE = 57426
const ADMIN = 57427
const SHOW = 57428
const DATABASES = 57429
const TABLES = 57430
const PROXY = 57431
const VARIABLES = 57432
const FULL = 57433
const COLUMNS = 57434
const CREATE = 57435
const ALTER = 57436
const DROP = 57437
const RENAME = 57438
const TABLE = 57439
const INDEX = 57440
const VIEW = 57441
const TO = 57442
const IGNORE = 57443
const IF = 57444
const UNIQUE = 57445
const USING = 57446

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
	-1, 206,
	75, 72,
	76, 72,
	-2, 148,
	-1, 420,
	75, 71,
	76, 71,
	-2, 81,
}

const yyNprod = 244
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 936

var yyAct = [...]int{

	114, 203, 106, 428, 73, 103, 395, 105, 111, 305,
	357, 128, 296, 253, 349, 223, 100, 298, 150, 112,
	437, 204, 3, 75, 218, 101, 437, 437, 171, 18,
	19, 20, 21, 62, 355, 91, 321, 322, 323, 324,
	325, 288, 326, 327, 77, 231, 87, 82, 171, 171,
	84, 251, 76, 251, 88, 64, 80, 44, 312, 46,
	97, 153, 22, 47, 406, 34, 35, 36, 37, 374,
	376, 83, 49, 405, 50, 404, 379, 439, 52, 53,
	54, 81, 149, 438, 436, 384, 51, 98, 78, 385,
	157, 354, 378, 94, 387, 297, 160, 346, 159, 163,
	95, 168, 169, 152, 175, 343, 341, 18, 289, 297,
	250, 332, 206, 375, 201, 205, 207, 202, 177, 27,
	28, 29, 146, 30, 32, 31, 290, 141, 401, 210,
	148, 162, 23, 24, 26, 25, 173, 174, 237, 381,
	217, 77, 256, 380, 77, 221, 74, 227, 226, 76,
	256, 255, 76, 56, 58, 59, 57, 61, 63, 255,
	276, 350, 235, 308, 169, 238, 228, 225, 70, 248,
	249, 189, 190, 192, 193, 191, 241, 264, 227, 262,
	263, 266, 257, 242, 274, 275, 156, 278, 279, 280,
	281, 282, 283, 284, 285, 286, 287, 260, 164, 270,
	244, 265, 415, 416, 143, 259, 176, 277, 403, 258,
	402, 368, 372, 259, 366, 350, 369, 258, 371, 367,
	77, 77, 63, 370, 301, 292, 294, 143, 76, 303,
	307, 251, 309, 413, 234, 236, 233, 304, 139, 390,
	300, 142, 243, 170, 77, 145, 173, 174, 314, 315,
	316, 310, 76, 220, 317, 219, 224, 173, 174, 158,
	313, 423, 224, 86, 300, 422, 220, 257, 318, 331,
	34, 35, 36, 37, 421, 337, 338, 138, 161, 333,
	334, 335, 187, 188, 189, 190, 192, 193, 191, 211,
	209, 336, 321, 322, 323, 324, 325, 171, 326, 327,
	348, 330, 208, 205, 99, 347, 412, 215, 345, 319,
	342, 214, 213, 212, 356, 143, 353, 329, 89, 352,
	63, 184, 185, 186, 187, 188, 189, 190, 192, 193,
	191, 257, 257, 364, 365, 434, 78, 377, 361, 360,
	240, 239, 383, 222, 71, 154, 151, 147, 144, 386,
	85, 140, 410, 389, 18, 90, 77, 69, 435, 340,
	271, 393, 272, 273, 391, 396, 382, 392, 92, 184,
	185, 186, 187, 188, 189, 190, 192, 193, 191, 397,
	441, 229, 299, 261, 155, 407, 93, 246, 67, 339,
	408, 409, 184, 185, 186, 187, 188, 189, 190, 192,
	193, 191, 419, 169, 247, 418, 65, 205, 417, 420,
	411, 358, 400, 359, 166, 425, 306, 224, 399, 396,
	426, 363, 429, 429, 429, 77, 430, 431, 427, 432,
	38, 167, 96, 76, 118, 119, 72, 293, 442, 440,
	117, 424, 443, 18, 444, 127, 39, 60, 133, 245,
	40, 41, 42, 43, 165, 104, 120, 121, 122, 17,
	16, 55, 15, 14, 109, 13, 12, 230, 131, 123,
	126, 124, 125, 45, 311, 232, 48, 79, 302, 433,
	414, 135, 134, 394, 398, 362, 344, 216, 295, 118,
	119, 116, 108, 113, 115, 117, 129, 130, 102, 351,
	127, 136, 110, 133, 178, 137, 107, 373, 254, 320,
	104, 120, 121, 122, 252, 328, 172, 66, 33, 109,
	68, 11, 10, 131, 123, 126, 124, 125, 9, 8,
	7, 6, 132, 5, 4, 291, 135, 134, 2, 1,
	0, 269, 268, 118, 119, 267, 0, 108, 0, 0,
	0, 129, 130, 102, 127, 0, 136, 133, 0, 0,
	137, 0, 0, 0, 78, 120, 121, 122, 0, 0,
	0, 0, 0, 161, 0, 0, 0, 131, 123, 126,
	124, 125, 0, 18, 0, 0, 0, 132, 0, 0,
	135, 134, 0, 0, 0, 0, 0, 118, 119, 0,
	0, 0, 0, 117, 0, 129, 130, 0, 127, 0,
	136, 133, 0, 0, 137, 0, 0, 0, 78, 120,
	121, 122, 0, 0, 0, 0, 0, 109, 0, 0,
	0, 131, 123, 126, 124, 125, 0, 0, 0, 0,
	0, 132, 0, 0, 135, 134, 0, 0, 0, 0,
	0, 118, 119, 0, 0, 108, 0, 117, 0, 129,
	130, 0, 127, 0, 136, 133, 0, 0, 137, 0,
	0, 0, 78, 120, 121, 122, 0, 0, 0, 0,
	0, 109, 0, 0, 0, 131, 123, 126, 124, 125,
	0, 18, 0, 0, 0, 132, 0, 0, 135, 134,
	0, 0, 0, 0, 0, 118, 119, 0, 0, 108,
	0, 0, 0, 129, 130, 0, 127, 0, 136, 133,
	0, 0, 137, 0, 0, 0, 78, 120, 121, 122,
	0, 0, 0, 0, 0, 161, 0, 0, 0, 131,
	123, 126, 124, 125, 0, 0, 0, 0, 0, 132,
	0, 0, 135, 134, 0, 0, 0, 0, 0, 118,
	119, 0, 0, 0, 0, 0, 0, 129, 130, 0,
	127, 0, 136, 133, 0, 0, 137, 0, 0, 0,
	78, 120, 121, 122, 0, 0, 0, 0, 0, 161,
	0, 0, 0, 131, 123, 126, 124, 125, 0, 0,
	0, 0, 0, 132, 0, 0, 135, 134, 0, 0,
	0, 0, 179, 183, 181, 182, 0, 0, 0, 0,
	0, 129, 130, 0, 0, 0, 136, 0, 0, 0,
	137, 197, 198, 199, 200, 0, 194, 195, 196, 184,
	185, 186, 187, 188, 189, 190, 192, 193, 191, 0,
	0, 0, 0, 0, 0, 0, 0, 132, 0, 0,
	0, 0, 0, 180, 184, 185, 186, 187, 188, 189,
	190, 192, 193, 191, 179, 183, 181, 182, 388, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 197, 198, 199, 200, 0, 194, 195,
	196, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 180, 184, 185, 186, 187,
	188, 189, 190, 192, 193, 191,
}
var yyPact = [...]int{

	24, -1000, -1000, 211, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -55, -42, -26, -34, -1000, -1000, -1000,
	-1000, 51, 280, 438, 385, -1000, -1000, -1000, 366, -1000,
	323, 304, 427, 48, -61, -32, 280, -1000, -41, 280,
	-1000, 310, -71, 280, -71, 321, 358, 358, 423, 280,
	-20, -1000, 255, -1000, -1000, -1000, 470, -1000, 233, 304,
	313, 39, 304, 164, 308, -1000, 195, -1000, 34, 307,
	53, 280, -1000, 306, -1000, -54, 305, 359, 112, 280,
	304, -1000, 632, 740, -1000, 358, 740, 423, 405, 740,
	234, -1000, -1000, 182, 30, -1000, 848, -1000, 632, 578,
	-1000, -1000, -1000, 740, 253, 241, -1000, 240, -1000, -1000,
	-1000, -1000, -1000, 272, 271, 270, 266, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 740, -1000, 217,
	296, 303, 407, 296, -1000, 740, 280, -1000, 356, -74,
	-1000, 125, -1000, 301, -1000, -1000, 300, -1000, 204, 61,
	761, 686, -1000, 761, 358, 378, 740, 740, -10, 761,
	110, 470, 360, 632, 632, -1000, 280, 118, 524, 229,
	334, 740, 740, 130, 740, 740, 740, 740, 740, 740,
	740, 740, 740, 740, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -79, -12, 6, 61, 848, -1000, 415, 470,
	-1000, 438, -1000, -1000, -1000, -1000, 18, 761, 349, 296,
	296, 252, -1000, 403, 632, -1000, 761, -1000, -1000, -1000,
	89, 280, -1000, -57, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 349, 296, -1000, -1000, 740, 740, 761, 761,
	-1000, 740, 246, 228, 277, 102, 23, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 761, 240, 240, 240,
	-1000, 229, 740, 740, 761, 314, -1000, 329, 201, 201,
	201, 88, 88, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -14, 470, -15, 4, -1000, 632, 87, 229,
	211, 141, -29, -1000, 403, 396, 399, 61, 299, -1000,
	-1000, 298, -1000, -1000, 164, 761, 761, 761, 410, 110,
	110, -1000, -1000, 150, 147, 159, 154, 148, -3, -1000,
	297, -28, 36, -1000, -1000, -1000, -1000, 761, 291, 740,
	-1000, -1000, -35, -1000, -5, -1000, 740, 2, 786, -1000,
	318, 176, -1000, -1000, -1000, 296, 396, -1000, 740, 632,
	-1000, -1000, 406, 398, 228, 54, -1000, 146, -1000, 144,
	-1000, -1000, -1000, -1000, -38, -40, -49, -1000, -1000, -1000,
	-1000, -1000, 740, 761, -1000, -1000, 761, 740, 740, 316,
	229, -1000, -1000, 243, 170, -1000, 171, -1000, 403, 632,
	740, 632, -1000, -1000, 225, 216, 212, 761, 761, 761,
	434, -1000, 740, 632, -1000, -1000, -1000, 396, 61, 168,
	-1000, 280, 280, 280, 296, 761, -1000, 319, -36, -1000,
	-37, -43, 164, -1000, 432, 354, -1000, 280, -1000, -1000,
	-1000, 280, -1000, 280, -1000,
}
var yyPgo = [...]int{

	0, 539, 538, 21, 534, 533, 531, 530, 529, 528,
	522, 521, 430, 520, 518, 517, 16, 25, 516, 515,
	5, 514, 13, 509, 508, 168, 507, 3, 15, 7,
	506, 504, 17, 502, 2, 19, 1, 499, 494, 11,
	493, 8, 491, 488, 12, 487, 486, 485, 484, 9,
	483, 6, 480, 10, 479, 24, 478, 14, 4, 23,
	263, 477, 476, 475, 474, 473, 467, 0, 18, 466,
	465, 463, 462, 460, 459, 100, 35, 454, 449, 447,
	446,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 72, 72, 5, 6, 7, 7, 69, 70,
	71, 74, 77, 77, 78, 78, 78, 79, 79, 73,
	73, 73, 73, 73, 8, 8, 8, 9, 9, 9,
	10, 11, 11, 11, 80, 12, 13, 13, 14, 14,
	14, 14, 14, 15, 15, 16, 16, 17, 17, 17,
	17, 20, 20, 18, 18, 18, 21, 21, 22, 22,
	22, 22, 19, 19, 19, 23, 23, 23, 23, 23,
	23, 23, 23, 23, 24, 24, 24, 24, 24, 24,
	24, 25, 25, 26, 26, 26, 26, 27, 27, 28,
	28, 76, 76, 76, 75, 75, 29, 29, 29, 29,
	29, 30, 30, 30, 30, 30, 30, 30, 30, 30,
	30, 30, 30, 30, 31, 31, 31, 31, 31, 31,
	31, 32, 32, 37, 37, 35, 35, 39, 36, 36,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 34,
	34, 34, 34, 34, 34, 34, 34, 34, 34, 38,
	38, 38, 38, 38, 40, 40, 40, 42, 45, 45,
	43, 43, 44, 44, 46, 46, 41, 41, 33, 33,
	33, 33, 33, 33, 33, 33, 33, 33, 47, 47,
	48, 48, 49, 49, 50, 50, 51, 52, 52, 52,
	53, 53, 53, 54, 54, 54, 55, 55, 56, 56,
	57, 57, 58, 58, 59, 60, 60, 61, 61, 62,
	62, 63, 63, 63, 63, 63, 64, 64, 65, 65,
	66, 66, 67, 68,
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
	3, 3, 3, 2, 3, 4, 5, 4, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 5, 0, 1,
	1, 2, 4, 4, 0, 2, 1, 3, 1, 1,
	1, 1, 1, 2, 2, 2, 2, 1, 0, 3,
	0, 2, 0, 3, 1, 3, 2, 0, 1, 1,
	0, 2, 4, 0, 2, 4, 0, 3, 1, 3,
	0, 5, 1, 3, 3, 0, 2, 0, 3, 0,
	1, 1, 1, 1, 1, 1, 0, 1, 0, 1,
	0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -69, -70, -71, -72, -73, -74, 5, 6,
	7, 8, 38, 108, 109, 111, 110, 95, 96, 97,
	99, 101, 100, -14, 59, 60, 61, 62, -12, -80,
	-12, -12, -12, -12, 112, -65, 114, 118, -62, 114,
	116, 112, 112, 113, 114, -12, 102, 105, 103, 104,
	-79, 106, -67, 40, -3, 21, -15, 22, -13, 34,
	-25, 40, 9, -58, 98, -59, -41, -67, 40, -61,
	117, 113, -67, 112, -67, 40, -60, 117, -67, -60,
	34, -76, 10, 28, -76, -75, 9, -67, 107, 49,
	-16, -17, 83, -20, 40, -29, -34, -30, 77, 49,
	-33, -41, -35, -40, -67, -38, -42, 25, 19, 20,
	41, 42, 43, 54, 56, 57, 55, 30, -39, 81,
	82, 53, 117, 33, 67, 66, 86, 90, 44, -25,
	38, 88, -25, 63, 40, 50, 88, 40, 77, -67,
	-68, 40, -68, 115, 40, 25, 74, -67, -25, -20,
	-34, 49, -76, -34, -75, -77, 9, 26, -36, -34,
	9, 63, -18, 75, 76, -67, 24, 88, -31, 26,
	77, 28, 29, 27, 78, 79, 80, 81, 82, 83,
	84, 87, 85, 86, 50, 51, 52, 45, 46, 47,
	48, -20, -29, -36, -3, -20, -34, -34, 49, 49,
	-39, 49, 41, 41, 41, 41, -45, -34, -55, 38,
	49, -58, 40, -28, 10, -59, -34, -67, -68, 25,
	-66, 119, -63, 111, 109, 37, 110, 13, 40, 40,
	40, -68, -55, 38, -76, -78, 9, 26, -34, -34,
	120, 63, -21, -22, -24, 49, 40, -39, 107, 103,
	-17, 23, -20, -20, -67, 83, -34, 21, 18, 17,
	-35, 26, 28, 29, -34, -34, 30, 77, -34, -34,
	-34, -34, -34, -34, -34, -34, -34, -34, 120, 120,
	120, 120, -16, 22, -16, -43, -44, 91, -32, 33,
	-3, -58, -56, -41, -28, -49, 13, -20, 74, -67,
	-68, -64, 115, -32, -58, -34, -34, -34, -28, 63,
	-23, 64, 65, 66, 67, 68, 70, 71, -19, 40,
	24, -22, 88, -39, -39, -39, -35, -34, -34, 75,
	30, 120, -16, 120, -46, -44, 93, -29, -34, -57,
	74, -37, -35, -57, 120, 63, -49, -53, 15, 14,
	40, 40, -47, 11, -22, -22, 64, 69, 64, 69,
	64, 64, 64, -26, 72, 116, 73, 40, 120, 40,
	107, 103, 75, -34, 120, 94, -34, 92, 92, 35,
	63, -41, -53, -34, -50, -51, -20, -68, -48, 12,
	14, 74, 64, 64, 113, 113, 113, -34, -34, -34,
	36, -35, 63, 63, -52, 31, 32, -49, -20, -36,
	-29, 49, 49, 49, 7, -34, -51, -53, -27, -67,
	-27, -27, -58, -54, 16, 39, 120, 63, 120, 120,
	7, 26, -67, -67, -67,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 238, 229, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 227, 0, 0, 239, 0, 0,
	230, 0, 225, 0, 225, 0, 111, 111, 114, 0,
	0, 38, 0, 242, 19, 59, 0, 64, 55, 0,
	0, 101, 0, 26, 0, 222, 0, 186, 242, 0,
	0, 0, 243, 0, 243, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 111, 0, 114, 0, 0,
	17, 65, 67, 73, 242, 71, 72, 116, 0, 0,
	150, 151, 152, 0, 186, 0, 168, 0, 188, 189,
	190, 191, 192, 0, 0, 0, 0, 197, 146, 174,
	175, 176, 169, 170, 171, 172, 173, 178, 57, 216,
	0, 0, 109, 0, 27, 0, 0, 243, 0, 240,
	46, 0, 49, 0, 51, 226, 0, 243, 216, 112,
	113, 0, 41, 115, 111, 34, 0, 0, 0, 148,
	0, 0, 68, 0, 0, 74, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 134, 135, 136, 137, 138, 139,
	140, 119, 71, 0, 0, 0, -2, 163, 0, 0,
	133, 0, 193, 194, 195, 196, 0, 179, 0, 0,
	0, 109, 102, 202, 0, 223, 224, 187, 44, 228,
	0, 0, 243, 236, 231, 232, 233, 234, 235, 50,
	52, 53, 0, 0, 42, 43, 0, 0, 32, 33,
	31, 0, 109, 76, 82, 0, 94, 96, 97, 98,
	66, 69, 117, 118, 75, 70, 121, 0, 0, 0,
	122, 0, 0, 0, 127, 0, 131, 0, 153, 154,
	155, 156, 157, 158, 159, 160, 161, 162, 120, 145,
	147, 164, 0, 0, 0, 184, 180, 0, 220, 0,
	142, 220, 0, 218, 202, 210, 0, 110, 0, 241,
	47, 0, 237, 22, 23, 35, 36, 149, 198, 0,
	0, 85, 86, 0, 0, 0, 0, 0, 103, 83,
	0, 0, 0, 123, 124, 125, 126, 128, 0, 0,
	132, 165, 0, 167, 0, 181, 0, 71, 72, 20,
	0, 141, 143, 21, 217, 0, 210, 25, 0, 0,
	243, 48, 200, 0, 77, 80, 87, 0, 89, 0,
	91, 92, 93, 78, 0, 0, 0, 84, 79, 95,
	99, 100, 0, 129, 166, 177, 185, 0, 0, 0,
	0, 219, 24, 211, 203, 204, 207, 45, 202, 0,
	0, 0, 88, 90, 0, 0, 0, 130, 182, 183,
	0, 144, 0, 0, 206, 208, 209, 210, 201, 199,
	-2, 0, 0, 0, 0, 212, 205, 213, 0, 107,
	0, 0, 221, 18, 0, 0, 104, 0, 105, 106,
	214, 0, 108, 0, 215,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 85, 78, 3,
	49, 120, 83, 81, 63, 82, 88, 84, 3, 3,
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
	86, 87, 89, 90, 91, 92, 93, 94, 95, 96,
	97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
	107, 108, 109, 110, 111, 112, 113, 114, 115, 116,
	117, 118, 119,
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
		//line sql.y:175
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:181
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:201
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 18:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:205
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].valExprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:209
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 20:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:216
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:220
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
		//line sql.y:232
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:236
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
		//line sql.y:249
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 25:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:255
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:261
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:265
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:271
		{
			yyVAL.statement = &Begin{}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:277
		{
			yyVAL.statement = &Commit{}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:283
		{
			yyVAL.statement = &Rollback{}
		}
	case 31:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:289
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].valExprs}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:295
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:299
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 34:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:304
		{
			yyVAL.valExpr = nil
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:308
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:312
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 37:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:317
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:321
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 39:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:328
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:332
		{
			yyVAL.statement = &Show{Section: "variables", LikeOrWhere: yyDollar[3].expr}
		}
	case 41:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:336
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].valExpr, LikeOrWhere: yyDollar[4].expr}
		}
	case 42:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:340
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].valExpr, LikeOrWhere: yyDollar[5].expr}
		}
	case 43:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:344
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].valExpr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].valExpr}
		}
	case 44:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:350
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 45:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:354
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 46:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:359
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 47:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:365
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 48:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:369
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 49:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:374
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 50:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:380
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 51:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:386
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 52:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:390
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 53:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:395
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 54:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:400
		{
			SetAllowComments(yylex, true)
		}
	case 55:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:404
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 56:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:410
		{
			yyVAL.bytes2 = nil
		}
	case 57:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:414
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 58:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:420
		{
			yyVAL.str = AST_UNION
		}
	case 59:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:424
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:428
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:432
		{
			yyVAL.str = AST_EXCEPT
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:436
		{
			yyVAL.str = AST_INTERSECT
		}
	case 63:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:441
		{
			yyVAL.str = ""
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:445
		{
			yyVAL.str = AST_DISTINCT
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:451
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:455
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:461
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:465
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:469
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 70:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:473
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:479
		{
			yyVAL.expr = yyDollar[1].boolExpr
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:483
		{
			yyVAL.expr = yyDollar[1].valExpr
		}
	case 73:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:488
		{
			yyVAL.bytes = nil
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:492
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:496
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 76:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:502
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 77:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:506
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:512
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:516
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:520
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 81:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:524
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].boolExpr}
		}
	case 82:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:529
		{
			yyVAL.bytes = nil
		}
	case 83:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:533
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 84:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:537
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 85:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:543
		{
			yyVAL.str = AST_JOIN
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:547
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 87:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:551
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:555
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 89:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:559
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:563
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 91:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:567
		{
			yyVAL.str = AST_JOIN
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:571
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:575
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 94:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:581
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:585
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 96:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:589
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:593
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:597
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:601
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 100:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:605
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 101:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:611
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:615
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 103:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:620
		{
			yyVAL.indexHints = nil
		}
	case 104:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:624
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 105:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:628
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 106:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:632
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 107:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:638
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 108:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:642
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 109:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:647
		{
			yyVAL.expr = nil
		}
	case 110:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:651
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 111:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:656
		{
			yyVAL.expr = nil
		}
	case 112:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:660
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:664
		{
			yyVAL.expr = yyDollar[2].valExpr
		}
	case 114:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:669
		{
			yyVAL.valExpr = nil
		}
	case 115:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:673
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 117:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:680
		{
			yyVAL.boolExpr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 118:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:684
		{
			yyVAL.boolExpr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 119:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:688
		{
			yyVAL.boolExpr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 120:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:692
		{
			yyVAL.boolExpr = &ParenBoolExpr{Expr: yyDollar[2].boolExpr}
		}
	case 121:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:698
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, Right: yyDollar[3].valExpr}
		}
	case 122:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:702
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 123:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:706
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ALL, Right: yyDollar[4].subquery}
		}
	case 124:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:710
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_ANY, Right: yyDollar[4].subquery}
		}
	case 125:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:714
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: yyDollar[2].str, SubqueryOperator: AST_SOME, Right: yyDollar[4].subquery}
		}
	case 126:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:718
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 127:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:722
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_LIKE, Right: yyDollar[3].valExpr}
		}
	case 128:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:726
		{
			yyVAL.boolExpr = &ComparisonExpr{Left: yyDollar[1].valExpr, Operator: AST_NOT_LIKE, Right: yyDollar[4].valExpr}
		}
	case 129:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:730
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_BETWEEN, From: yyDollar[3].valExpr, To: yyDollar[5].valExpr}
		}
	case 130:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:734
		{
			yyVAL.boolExpr = &RangeCond{Left: yyDollar[1].valExpr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].valExpr, To: yyDollar[6].valExpr}
		}
	case 131:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:738
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NULL, Expr: yyDollar[1].valExpr}
		}
	case 132:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:742
		{
			yyVAL.boolExpr = &NullCheck{Operator: AST_IS_NOT_NULL, Expr: yyDollar[1].valExpr}
		}
	case 133:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:746
		{
			yyVAL.boolExpr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 134:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:752
		{
			yyVAL.str = AST_EQ
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:756
		{
			yyVAL.str = AST_LT
		}
	case 136:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:760
		{
			yyVAL.str = AST_GT
		}
	case 137:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:764
		{
			yyVAL.str = AST_LE
		}
	case 138:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:768
		{
			yyVAL.str = AST_GE
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:772
		{
			yyVAL.str = AST_NE
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:776
		{
			yyVAL.str = AST_NSE
		}
	case 141:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:782
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:786
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:792
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 144:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:796
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:802
		{
			yyVAL.tuple = ValTuple(yyDollar[2].valExprs)
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:806
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:812
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 148:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:818
		{
			yyVAL.valExprs = ValExprs{yyDollar[1].valExpr}
		}
	case 149:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:822
		{
			yyVAL.valExprs = append(yyDollar[1].valExprs, yyDollar[3].valExpr)
		}
	case 150:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:828
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 151:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:832
		{
			yyVAL.valExpr = yyDollar[1].colName
		}
	case 152:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:836
		{
			yyVAL.valExpr = yyDollar[1].tuple
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:840
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITAND, Right: yyDollar[3].valExpr}
		}
	case 154:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:844
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITOR, Right: yyDollar[3].valExpr}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:848
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_BITXOR, Right: yyDollar[3].valExpr}
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:852
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_PLUS, Right: yyDollar[3].valExpr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:856
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MINUS, Right: yyDollar[3].valExpr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:860
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MULT, Right: yyDollar[3].valExpr}
		}
	case 159:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:864
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_DIV, Right: yyDollar[3].valExpr}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:868
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_IDIV, Right: yyDollar[3].valExpr}
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:872
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:876
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 163:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:880
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
		//line sql.y:895
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 165:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:899
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 166:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:903
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 167:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:907
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 168:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:911
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 169:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:917
		{
			yyVAL.bytes = IF_BYTES
		}
	case 170:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:921
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 171:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:925
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 172:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:929
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 173:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:933
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 174:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:939
		{
			yyVAL.byt = AST_UPLUS
		}
	case 175:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:943
		{
			yyVAL.byt = AST_UMINUS
		}
	case 176:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:947
		{
			yyVAL.byt = AST_TILDA
		}
	case 177:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:953
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 178:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:958
		{
			yyVAL.valExpr = nil
		}
	case 179:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:962
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 180:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:968
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 181:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:972
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 182:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:978
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 183:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:982
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 184:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:987
		{
			yyVAL.valExpr = nil
		}
	case 185:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:991
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 186:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:997
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 187:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1001
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1007
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1011
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1015
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1019
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 193:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1027
		{
			yyVAL.valExpr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 194:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1031
		{
			yyVAL.valExpr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 195:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1035
		{
			yyVAL.valExpr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 196:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1039
		{
			yyVAL.valExpr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1043
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 198:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1048
		{
			yyVAL.valExprs = nil
		}
	case 199:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1052
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 200:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1057
		{
			yyVAL.expr = nil
		}
	case 201:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1061
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 202:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1066
		{
			yyVAL.orderBy = nil
		}
	case 203:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1070
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1076
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 205:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1080
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 206:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1086
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 207:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1091
		{
			yyVAL.str = AST_ASC
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1095
		{
			yyVAL.str = AST_ASC
		}
	case 209:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1099
		{
			yyVAL.str = AST_DESC
		}
	case 210:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1104
		{
			yyVAL.limit = nil
		}
	case 211:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1108
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 212:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1112
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 213:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1117
		{
			yyVAL.str = ""
		}
	case 214:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1121
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 215:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1125
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
	case 216:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1138
		{
			yyVAL.columns = nil
		}
	case 217:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1142
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1148
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 219:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1152
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 220:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1157
		{
			yyVAL.updateExprs = nil
		}
	case 221:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1161
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1167
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 223:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1171
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 224:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1177
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 225:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1182
		{
			yyVAL.empty = struct{}{}
		}
	case 226:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1184
		{
			yyVAL.empty = struct{}{}
		}
	case 227:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1187
		{
			yyVAL.empty = struct{}{}
		}
	case 228:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1189
		{
			yyVAL.empty = struct{}{}
		}
	case 229:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1192
		{
			yyVAL.empty = struct{}{}
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1194
		{
			yyVAL.empty = struct{}{}
		}
	case 231:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1198
		{
			yyVAL.empty = struct{}{}
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1200
		{
			yyVAL.empty = struct{}{}
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1202
		{
			yyVAL.empty = struct{}{}
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1204
		{
			yyVAL.empty = struct{}{}
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1206
		{
			yyVAL.empty = struct{}{}
		}
	case 236:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1209
		{
			yyVAL.empty = struct{}{}
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1211
		{
			yyVAL.empty = struct{}{}
		}
	case 238:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1214
		{
			yyVAL.empty = struct{}{}
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1216
		{
			yyVAL.empty = struct{}{}
		}
	case 240:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1219
		{
			yyVAL.empty = struct{}{}
		}
	case 241:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1221
		{
			yyVAL.empty = struct{}{}
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1225
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 243:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1230
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
