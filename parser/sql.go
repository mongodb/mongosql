//line sql.y:5
package parser

import __yyfmt__ "fmt"

//line sql.y:7
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
const YEAR = 57396
const UNION = 57397
const MINUS = 57398
const EXCEPT = 57399
const INTERSECT = 57400
const JOIN = 57401
const STRAIGHT_JOIN = 57402
const LEFT = 57403
const RIGHT = 57404
const INNER = 57405
const OUTER = 57406
const CROSS = 57407
const NATURAL = 57408
const USE = 57409
const FORCE = 57410
const ON = 57411
const AND = 57412
const OR = 57413
const NOT = 57414
const MOD = 57415
const DIV = 57416
const UNARY = 57417
const CASE = 57418
const WHEN = 57419
const THEN = 57420
const ELSE = 57421
const END = 57422
const BEGIN = 57423
const COMMIT = 57424
const ROLLBACK = 57425
const NAMES = 57426
const REPLACE = 57427
const ADMIN = 57428
const SHOW = 57429
const DATABASES = 57430
const TABLES = 57431
const PROXY = 57432
const VARIABLES = 57433
const FULL = 57434
const COLUMNS = 57435
const CREATE = 57436
const ALTER = 57437
const DROP = 57438
const RENAME = 57439
const TABLE = 57440
const INDEX = 57441
const VIEW = 57442
const TO = 57443
const IGNORE = 57444
const IF = 57445
const UNIQUE = 57446
const USING = 57447

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
const yyMaxDepth = 200

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 206,
	76, 72,
	77, 72,
	-2, 148,
	-1, 421,
	76, 71,
	77, 71,
	-2, 81,
}

const yyNprod = 245
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 938

var yyAct = [...]int{

	114, 203, 106, 430, 73, 103, 395, 105, 111, 305,
	357, 128, 296, 253, 349, 223, 100, 298, 150, 112,
	439, 204, 3, 75, 218, 101, 439, 439, 18, 19,
	20, 21, 171, 62, 355, 91, 321, 322, 323, 324,
	325, 288, 326, 327, 77, 231, 87, 82, 171, 171,
	84, 251, 76, 251, 88, 64, 80, 44, 312, 46,
	97, 153, 22, 47, 406, 34, 35, 36, 37, 374,
	376, 83, 49, 405, 50, 404, 95, 441, 52, 53,
	54, 81, 149, 440, 438, 51, 98, 379, 256, 384,
	157, 354, 378, 94, 385, 78, 160, 255, 159, 163,
	387, 168, 169, 152, 175, 343, 341, 297, 289, 346,
	250, 237, 206, 375, 201, 205, 207, 202, 297, 27,
	28, 29, 332, 30, 32, 31, 290, 177, 146, 210,
	141, 162, 23, 24, 26, 25, 235, 173, 174, 238,
	217, 77, 148, 63, 77, 221, 401, 227, 226, 76,
	381, 259, 76, 74, 380, 258, 56, 58, 59, 57,
	61, 143, 18, 350, 169, 70, 228, 225, 308, 248,
	249, 156, 350, 403, 164, 276, 241, 264, 227, 262,
	263, 266, 257, 242, 274, 275, 265, 278, 279, 280,
	281, 282, 283, 284, 285, 286, 287, 260, 256, 270,
	244, 189, 190, 192, 193, 191, 402, 255, 234, 236,
	233, 187, 188, 189, 190, 192, 193, 191, 416, 417,
	77, 77, 277, 372, 301, 292, 294, 371, 76, 303,
	307, 176, 309, 370, 368, 139, 366, 304, 142, 369,
	300, 367, 224, 224, 77, 143, 170, 63, 314, 315,
	316, 310, 76, 251, 317, 414, 158, 390, 413, 145,
	313, 259, 173, 174, 300, 258, 243, 257, 318, 331,
	34, 35, 36, 37, 219, 337, 338, 220, 424, 333,
	334, 335, 173, 174, 86, 220, 321, 322, 323, 324,
	325, 336, 326, 327, 423, 422, 319, 143, 161, 211,
	348, 171, 209, 205, 208, 347, 412, 99, 345, 138,
	342, 215, 214, 213, 356, 212, 353, 63, 436, 352,
	330, 184, 185, 186, 187, 188, 189, 190, 192, 193,
	191, 257, 257, 364, 365, 78, 329, 377, 361, 89,
	360, 437, 383, 240, 239, 222, 71, 154, 151, 386,
	147, 144, 85, 140, 410, 389, 77, 18, 90, 69,
	271, 393, 272, 273, 391, 396, 382, 392, 340, 184,
	185, 186, 187, 188, 189, 190, 192, 193, 191, 397,
	443, 229, 155, 261, 67, 407, 299, 92, 65, 339,
	408, 409, 184, 185, 186, 187, 188, 189, 190, 192,
	193, 191, 420, 169, 246, 419, 93, 205, 418, 421,
	411, 358, 400, 359, 166, 426, 427, 306, 399, 363,
	396, 428, 247, 431, 431, 431, 77, 432, 433, 429,
	434, 38, 167, 224, 76, 96, 118, 119, 72, 293,
	444, 442, 117, 425, 445, 18, 446, 127, 39, 60,
	133, 40, 41, 42, 43, 245, 165, 104, 120, 121,
	122, 17, 55, 16, 15, 14, 109, 13, 12, 230,
	131, 123, 126, 124, 125, 45, 311, 232, 48, 79,
	302, 435, 415, 135, 134, 394, 398, 362, 344, 216,
	295, 118, 119, 116, 108, 113, 115, 117, 129, 130,
	102, 351, 127, 136, 110, 133, 178, 137, 107, 373,
	254, 320, 104, 120, 121, 122, 252, 328, 172, 66,
	33, 109, 68, 11, 10, 131, 123, 126, 124, 125,
	9, 8, 7, 6, 132, 5, 4, 291, 135, 134,
	2, 1, 0, 269, 268, 118, 119, 267, 0, 108,
	0, 0, 0, 129, 130, 102, 127, 0, 136, 133,
	0, 0, 137, 0, 0, 0, 78, 120, 121, 122,
	0, 0, 0, 0, 0, 161, 0, 0, 0, 131,
	123, 126, 124, 125, 18, 0, 0, 0, 0, 132,
	0, 0, 135, 134, 0, 0, 0, 0, 0, 118,
	119, 0, 0, 0, 0, 117, 0, 129, 130, 0,
	127, 0, 136, 133, 0, 0, 137, 0, 0, 0,
	78, 120, 121, 122, 0, 0, 0, 0, 0, 109,
	0, 0, 0, 131, 123, 126, 124, 125, 0, 0,
	0, 0, 0, 132, 0, 0, 135, 134, 0, 0,
	0, 0, 0, 118, 119, 0, 0, 108, 0, 117,
	0, 129, 130, 0, 127, 0, 136, 133, 0, 0,
	137, 0, 0, 0, 78, 120, 121, 122, 0, 0,
	0, 0, 0, 109, 0, 0, 0, 131, 123, 126,
	124, 125, 18, 0, 0, 0, 0, 132, 0, 0,
	135, 134, 0, 0, 0, 0, 0, 118, 119, 0,
	0, 108, 0, 0, 0, 129, 130, 0, 127, 0,
	136, 133, 0, 0, 137, 0, 0, 0, 78, 120,
	121, 122, 0, 0, 0, 0, 0, 161, 0, 0,
	0, 131, 123, 126, 124, 125, 0, 0, 0, 0,
	0, 132, 0, 0, 135, 134, 0, 0, 0, 0,
	0, 118, 119, 0, 0, 0, 0, 0, 0, 129,
	130, 0, 127, 0, 136, 133, 0, 0, 137, 0,
	0, 0, 78, 120, 121, 122, 0, 0, 0, 0,
	0, 161, 0, 0, 0, 131, 123, 126, 124, 125,
	0, 0, 0, 0, 0, 132, 0, 0, 135, 134,
	0, 0, 0, 0, 179, 183, 181, 182, 0, 0,
	0, 0, 0, 129, 130, 0, 0, 0, 136, 0,
	0, 0, 137, 197, 198, 199, 200, 0, 194, 195,
	196, 184, 185, 186, 187, 188, 189, 190, 192, 193,
	191, 0, 0, 0, 0, 0, 0, 0, 0, 132,
	0, 0, 0, 0, 0, 180, 184, 185, 186, 187,
	188, 189, 190, 192, 193, 191, 179, 183, 181, 182,
	388, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 197, 198, 199, 200, 0,
	194, 195, 196, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 180, 184, 185,
	186, 187, 188, 189, 190, 192, 193, 191,
}
var yyPact = [...]int{

	23, -1000, -1000, 210, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -56, -43, -28, -35, -1000, -1000, -1000,
	-1000, 53, 276, 440, 366, -1000, -1000, -1000, 361, -1000,
	324, 305, 429, 54, -62, -33, 276, -1000, -42, 276,
	-1000, 311, -72, 276, -72, 323, 377, 377, 426, 276,
	-22, -1000, 257, -1000, -1000, -1000, 471, -1000, 264, 305,
	314, 41, 305, 181, 310, -1000, 208, -1000, 39, 309,
	64, 276, -1000, 307, -1000, -55, 306, 356, 96, 276,
	305, -1000, 633, 741, -1000, 377, 741, 426, 405, 741,
	237, -1000, -1000, 206, 38, -1000, 849, -1000, 633, 579,
	-1000, -1000, -1000, 741, 254, 252, -1000, 249, -1000, -1000,
	-1000, -1000, -1000, 273, 271, 270, 269, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 741, -1000, 235,
	294, 304, 423, 294, -1000, 741, 276, -1000, 355, -75,
	-1000, 98, -1000, 303, -1000, -1000, 302, -1000, 227, 61,
	762, 687, -1000, 762, 377, 395, 741, 741, -11, 762,
	47, 471, 359, 633, 633, -1000, 276, 102, 525, 248,
	333, 741, 741, 144, 741, 741, 741, 741, 741, 741,
	741, 741, 741, 741, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -80, -13, 5, 61, 849, -1000, 416, 471,
	-1000, 440, -1000, -1000, -1000, -1000, 26, 762, 352, 294,
	294, 233, -1000, 404, 633, -1000, 762, -1000, -1000, -1000,
	93, 276, -1000, -58, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 352, 294, -1000, -1000, 741, 741, 762, 762,
	-1000, 741, 232, 221, 295, 157, 33, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 762, 249, 249, 249,
	-1000, 248, 741, 741, 762, 313, -1000, 337, 129, 129,
	129, 117, 117, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -15, 471, -16, 15, -1000, 633, 88, 248,
	210, 97, -30, -1000, 404, 396, 399, 61, 299, -1000,
	-1000, 297, -1000, -1000, 181, 762, 762, 762, 408, 47,
	47, -1000, -1000, 171, 169, 168, 162, 158, -4, -1000,
	296, -29, 46, -1000, -1000, -1000, -1000, 762, 290, 741,
	-1000, -1000, -32, -1000, -1, -1000, 741, 7, 787, -1000,
	319, 193, -1000, -1000, -1000, 294, 396, -1000, 741, 633,
	-1000, -1000, 406, 398, 221, 71, -1000, 141, -1000, 108,
	-1000, -1000, -1000, -1000, -39, -41, -50, -1000, -1000, -1000,
	-1000, -1000, 741, 762, -1000, -1000, 762, 741, 741, 317,
	248, -1000, -1000, 242, 191, -1000, 186, -1000, 404, 633,
	741, 633, -1000, -1000, 245, 244, 228, 762, 762, 762,
	436, -1000, 741, 741, 633, -1000, -1000, -1000, 396, 61,
	189, -1000, 276, 276, 276, 294, 762, 762, -1000, 301,
	-37, -1000, -38, -44, 181, -1000, 434, 353, -1000, 276,
	-1000, -1000, -1000, 276, -1000, 276, -1000,
}
var yyPgo = [...]int{

	0, 541, 540, 21, 536, 535, 533, 532, 531, 530,
	524, 523, 431, 522, 520, 519, 16, 25, 518, 517,
	5, 516, 13, 511, 510, 165, 509, 3, 15, 7,
	508, 506, 17, 504, 2, 19, 1, 501, 496, 11,
	495, 8, 493, 490, 12, 489, 488, 487, 486, 9,
	485, 6, 482, 10, 481, 24, 480, 14, 4, 23,
	284, 479, 478, 477, 476, 475, 469, 0, 18, 468,
	467, 465, 464, 463, 461, 76, 35, 456, 455, 449,
	448,
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
	53, 53, 53, 53, 54, 54, 54, 55, 55, 56,
	56, 57, 57, 58, 58, 59, 60, 60, 61, 61,
	62, 62, 63, 63, 63, 63, 63, 64, 64, 65,
	65, 66, 66, 67, 68,
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
	0, 2, 4, 4, 0, 2, 4, 0, 3, 1,
	3, 0, 5, 1, 3, 3, 0, 2, 0, 3,
	0, 1, 1, 1, 1, 1, 1, 0, 1, 0,
	1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -69, -70, -71, -72, -73, -74, 5, 6,
	7, 8, 39, 109, 110, 112, 111, 96, 97, 98,
	100, 102, 101, -14, 60, 61, 62, 63, -12, -80,
	-12, -12, -12, -12, 113, -65, 115, 119, -62, 115,
	117, 113, 113, 114, 115, -12, 103, 106, 104, 105,
	-79, 107, -67, 41, -3, 22, -15, 23, -13, 35,
	-25, 41, 9, -58, 99, -59, -41, -67, 41, -61,
	118, 114, -67, 113, -67, 41, -60, 118, -67, -60,
	35, -76, 10, 29, -76, -75, 9, -67, 108, 50,
	-16, -17, 84, -20, 41, -29, -34, -30, 78, 50,
	-33, -41, -35, -40, -67, -38, -42, 26, 20, 21,
	42, 43, 44, 55, 57, 58, 56, 31, -39, 82,
	83, 54, 118, 34, 68, 67, 87, 91, 45, -25,
	39, 89, -25, 64, 41, 51, 89, 41, 78, -67,
	-68, 41, -68, 116, 41, 26, 75, -67, -25, -20,
	-34, 50, -76, -34, -75, -77, 9, 27, -36, -34,
	9, 64, -18, 76, 77, -67, 25, 89, -31, 27,
	78, 29, 30, 28, 79, 80, 81, 82, 83, 84,
	85, 88, 86, 87, 51, 52, 53, 46, 47, 48,
	49, -20, -29, -36, -3, -20, -34, -34, 50, 50,
	-39, 50, 42, 42, 42, 42, -45, -34, -55, 39,
	50, -58, 41, -28, 10, -59, -34, -67, -68, 26,
	-66, 120, -63, 112, 110, 38, 111, 13, 41, 41,
	41, -68, -55, 39, -76, -78, 9, 27, -34, -34,
	121, 64, -21, -22, -24, 50, 41, -39, 108, 104,
	-17, 24, -20, -20, -67, 84, -34, 22, 19, 18,
	-35, 27, 29, 30, -34, -34, 31, 78, -34, -34,
	-34, -34, -34, -34, -34, -34, -34, -34, 121, 121,
	121, 121, -16, 23, -16, -43, -44, 92, -32, 34,
	-3, -58, -56, -41, -28, -49, 13, -20, 75, -67,
	-68, -64, 116, -32, -58, -34, -34, -34, -28, 64,
	-23, 65, 66, 67, 68, 69, 71, 72, -19, 41,
	25, -22, 89, -39, -39, -39, -35, -34, -34, 76,
	31, 121, -16, 121, -46, -44, 94, -29, -34, -57,
	75, -37, -35, -57, 121, 64, -49, -53, 15, 14,
	41, 41, -47, 11, -22, -22, 65, 70, 65, 70,
	65, 65, 65, -26, 73, 117, 74, 41, 121, 41,
	108, 104, 76, -34, 121, 95, -34, 93, 93, 36,
	64, -41, -53, -34, -50, -51, -20, -68, -48, 12,
	14, 75, 65, 65, 114, 114, 114, -34, -34, -34,
	37, -35, 64, 16, 64, -52, 32, 33, -49, -20,
	-36, -29, 50, 50, 50, 7, -34, -34, -51, -53,
	-27, -67, -27, -27, -58, -54, 17, 40, 121, 64,
	121, 121, 7, 27, -67, -67, -67,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 239, 230, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 228, 0, 0, 240, 0, 0,
	231, 0, 226, 0, 226, 0, 111, 111, 114, 0,
	0, 38, 0, 243, 19, 59, 0, 64, 55, 0,
	0, 101, 0, 26, 0, 223, 0, 186, 243, 0,
	0, 0, 244, 0, 244, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 111, 0, 114, 0, 0,
	17, 65, 67, 73, 243, 71, 72, 116, 0, 0,
	150, 151, 152, 0, 186, 0, 168, 0, 188, 189,
	190, 191, 192, 0, 0, 0, 0, 197, 146, 174,
	175, 176, 169, 170, 171, 172, 173, 178, 57, 217,
	0, 0, 109, 0, 27, 0, 0, 244, 0, 241,
	46, 0, 49, 0, 51, 227, 0, 244, 217, 112,
	113, 0, 41, 115, 111, 34, 0, 0, 0, 148,
	0, 0, 68, 0, 0, 74, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 134, 135, 136, 137, 138, 139,
	140, 119, 71, 0, 0, 0, -2, 163, 0, 0,
	133, 0, 193, 194, 195, 196, 0, 179, 0, 0,
	0, 109, 102, 202, 0, 224, 225, 187, 44, 229,
	0, 0, 244, 237, 232, 233, 234, 235, 236, 50,
	52, 53, 0, 0, 42, 43, 0, 0, 32, 33,
	31, 0, 109, 76, 82, 0, 94, 96, 97, 98,
	66, 69, 117, 118, 75, 70, 121, 0, 0, 0,
	122, 0, 0, 0, 127, 0, 131, 0, 153, 154,
	155, 156, 157, 158, 159, 160, 161, 162, 120, 145,
	147, 164, 0, 0, 0, 184, 180, 0, 221, 0,
	142, 221, 0, 219, 202, 210, 0, 110, 0, 242,
	47, 0, 238, 22, 23, 35, 36, 149, 198, 0,
	0, 85, 86, 0, 0, 0, 0, 0, 103, 83,
	0, 0, 0, 123, 124, 125, 126, 128, 0, 0,
	132, 165, 0, 167, 0, 181, 0, 71, 72, 20,
	0, 141, 143, 21, 218, 0, 210, 25, 0, 0,
	244, 48, 200, 0, 77, 80, 87, 0, 89, 0,
	91, 92, 93, 78, 0, 0, 0, 84, 79, 95,
	99, 100, 0, 129, 166, 177, 185, 0, 0, 0,
	0, 220, 24, 211, 203, 204, 207, 45, 202, 0,
	0, 0, 88, 90, 0, 0, 0, 130, 182, 183,
	0, 144, 0, 0, 0, 206, 208, 209, 210, 201,
	199, -2, 0, 0, 0, 0, 212, 213, 205, 214,
	0, 107, 0, 0, 222, 18, 0, 0, 104, 0,
	105, 106, 215, 0, 108, 0, 216,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 86, 79, 3,
	50, 121, 84, 82, 64, 83, 89, 85, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	52, 51, 53, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 81, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 80, 3, 54,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 47, 48, 49, 55, 56,
	57, 58, 59, 60, 61, 62, 63, 65, 66, 67,
	68, 69, 70, 71, 72, 73, 74, 75, 76, 77,
	78, 87, 88, 90, 91, 92, 93, 94, 95, 96,
	97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
	107, 108, 109, 110, 111, 112, 113, 114, 115, 116,
	117, 118, 119, 120,
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
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1116
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].valExpr, Rowcount: yyDollar[2].valExpr}
		}
	case 214:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1121
		{
			yyVAL.str = ""
		}
	case 215:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1125
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 216:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1129
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
	case 217:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1142
		{
			yyVAL.columns = nil
		}
	case 218:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1146
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1152
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 220:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1156
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 221:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1161
		{
			yyVAL.updateExprs = nil
		}
	case 222:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1165
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1171
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 224:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1175
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 225:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1181
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 226:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1186
		{
			yyVAL.empty = struct{}{}
		}
	case 227:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1188
		{
			yyVAL.empty = struct{}{}
		}
	case 228:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1191
		{
			yyVAL.empty = struct{}{}
		}
	case 229:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1193
		{
			yyVAL.empty = struct{}{}
		}
	case 230:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1196
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
		//line sql.y:1202
		{
			yyVAL.empty = struct{}{}
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1204
		{
			yyVAL.empty = struct{}{}
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1206
		{
			yyVAL.empty = struct{}{}
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1208
		{
			yyVAL.empty = struct{}{}
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1210
		{
			yyVAL.empty = struct{}{}
		}
	case 237:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1213
		{
			yyVAL.empty = struct{}{}
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1215
		{
			yyVAL.empty = struct{}{}
		}
	case 239:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1218
		{
			yyVAL.empty = struct{}{}
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1220
		{
			yyVAL.empty = struct{}{}
		}
	case 241:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1223
		{
			yyVAL.empty = struct{}{}
		}
	case 242:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1225
		{
			yyVAL.empty = struct{}{}
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1229
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 244:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1234
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
