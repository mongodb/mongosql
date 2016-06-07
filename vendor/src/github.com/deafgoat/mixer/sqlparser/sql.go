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
	-1, 198,
	75, 72,
	76, 72,
	-2, 149,
	-1, 416,
	75, 71,
	76, 71,
	-2, 82,
}

const yyNprod = 243
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 809

var yyAct = [...]int{

	114, 353, 106, 424, 73, 103, 391, 105, 111, 194,
	143, 124, 300, 247, 345, 112, 291, 217, 293, 196,
	100, 195, 3, 75, 212, 101, 433, 282, 433, 34,
	35, 36, 37, 62, 91, 316, 317, 318, 319, 320,
	225, 321, 322, 87, 77, 433, 164, 82, 351, 80,
	84, 164, 76, 164, 88, 64, 245, 49, 245, 50,
	97, 307, 146, 118, 119, 245, 44, 83, 46, 117,
	402, 401, 47, 400, 123, 81, 51, 129, 52, 53,
	54, 435, 142, 434, 104, 120, 121, 122, 284, 374,
	150, 98, 94, 109, 383, 145, 153, 127, 152, 156,
	432, 380, 162, 350, 168, 95, 339, 381, 337, 161,
	78, 336, 198, 283, 192, 197, 205, 193, 292, 327,
	244, 108, 370, 372, 170, 125, 126, 102, 139, 208,
	155, 134, 130, 211, 77, 141, 375, 77, 215, 397,
	221, 220, 76, 250, 399, 76, 56, 58, 59, 57,
	61, 222, 249, 292, 136, 342, 18, 162, 398, 128,
	219, 235, 242, 243, 371, 346, 74, 182, 183, 184,
	258, 221, 256, 257, 261, 251, 236, 270, 271, 346,
	274, 275, 276, 277, 278, 279, 280, 281, 265, 259,
	254, 250, 238, 231, 303, 70, 63, 377, 272, 162,
	249, 376, 364, 157, 253, 149, 285, 365, 252, 180,
	181, 182, 183, 184, 77, 77, 362, 229, 296, 368,
	232, 363, 76, 298, 302, 367, 304, 287, 289, 18,
	19, 20, 21, 299, 295, 166, 167, 305, 77, 260,
	366, 169, 309, 310, 311, 273, 76, 136, 312, 245,
	409, 386, 253, 218, 163, 308, 252, 63, 295, 86,
	218, 251, 22, 326, 313, 132, 237, 419, 135, 162,
	138, 332, 333, 131, 328, 329, 330, 214, 316, 317,
	318, 319, 320, 331, 321, 322, 151, 228, 230, 227,
	418, 63, 166, 167, 213, 344, 417, 266, 197, 78,
	343, 34, 35, 36, 37, 214, 314, 341, 164, 338,
	348, 349, 352, 136, 89, 209, 207, 27, 28, 29,
	206, 30, 32, 31, 411, 412, 251, 251, 360, 361,
	23, 24, 26, 25, 99, 325, 430, 379, 373, 357,
	356, 234, 233, 216, 71, 382, 147, 63, 144, 140,
	133, 324, 77, 137, 388, 85, 385, 389, 392, 431,
	387, 200, 203, 201, 202, 204, 406, 393, 18, 90,
	69, 177, 178, 179, 180, 181, 182, 183, 184, 335,
	267, 403, 268, 269, 437, 378, 404, 405, 177, 178,
	179, 180, 181, 182, 183, 184, 294, 240, 223, 162,
	159, 414, 407, 197, 148, 416, 415, 413, 255, 92,
	65, 421, 392, 354, 241, 423, 422, 160, 425, 425,
	425, 77, 426, 427, 67, 428, 38, 93, 396, 76,
	118, 119, 355, 288, 438, 301, 117, 395, 439, 359,
	440, 123, 218, 96, 129, 72, 40, 41, 42, 43,
	436, 104, 120, 121, 122, 420, 408, 55, 18, 39,
	109, 60, 239, 158, 127, 17, 16, 15, 14, 13,
	18, 177, 178, 179, 180, 181, 182, 183, 184, 12,
	199, 224, 45, 306, 118, 119, 226, 48, 108, 79,
	117, 297, 125, 126, 102, 123, 429, 410, 129, 130,
	390, 394, 358, 340, 210, 78, 120, 121, 122, 290,
	116, 113, 115, 18, 109, 347, 110, 171, 127, 200,
	203, 201, 202, 204, 107, 369, 128, 118, 119, 286,
	177, 178, 179, 180, 181, 182, 183, 184, 123, 248,
	315, 129, 108, 246, 323, 165, 125, 126, 78, 120,
	121, 122, 66, 130, 33, 68, 11, 154, 10, 9,
	8, 127, 200, 203, 201, 202, 204, 7, 264, 263,
	118, 119, 262, 6, 5, 4, 2, 1, 0, 0,
	128, 123, 0, 0, 129, 0, 0, 0, 0, 125,
	126, 78, 120, 121, 122, 0, 130, 0, 0, 0,
	154, 0, 0, 0, 127, 0, 118, 119, 0, 0,
	18, 0, 117, 0, 0, 0, 0, 123, 0, 0,
	129, 0, 0, 128, 118, 119, 0, 78, 120, 121,
	122, 0, 125, 126, 0, 123, 109, 0, 129, 130,
	127, 0, 0, 0, 0, 78, 120, 121, 122, 118,
	119, 0, 0, 0, 154, 0, 0, 0, 127, 0,
	123, 0, 0, 129, 108, 0, 128, 0, 125, 126,
	78, 120, 121, 122, 0, 130, 0, 0, 0, 154,
	0, 0, 0, 127, 0, 0, 125, 126, 0, 0,
	0, 334, 0, 130, 177, 178, 179, 180, 181, 182,
	183, 184, 128, 172, 176, 174, 175, 0, 0, 0,
	0, 125, 126, 0, 0, 0, 0, 0, 130, 0,
	128, 0, 188, 189, 190, 191, 0, 185, 186, 187,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 128, 0, 0, 0, 172,
	176, 174, 175, 0, 173, 177, 178, 179, 180, 181,
	182, 183, 184, 0, 0, 0, 0, 384, 188, 189,
	190, 191, 0, 185, 186, 187, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	173, 177, 178, 179, 180, 181, 182, 183, 184,
}
var yyPact = [...]int{

	224, -1000, -1000, 242, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -44, -55, -34, -32, -1000, -1000, -1000,
	-1000, 46, 251, 453, 389, -1000, -1000, -1000, 402, -1000,
	336, 304, 436, 70, -66, -36, 251, -1000, -43, 251,
	-1000, 315, -72, 251, -72, 335, 399, 399, 434, 251,
	-14, -1000, 285, -1000, -1000, -1000, 44, -1000, 229, 304,
	312, 45, 304, 184, 313, -1000, 220, -1000, 42, 309,
	58, 251, -1000, 308, -1000, -51, 306, 379, 131, 251,
	304, -1000, 587, 630, -1000, 399, 630, 434, 391, 630,
	245, -1000, -1000, 217, 38, -1000, 723, -1000, 587, 465,
	-1000, -1000, -1000, 630, 271, 267, -1000, 266, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	630, -1000, 256, 259, 303, 432, 259, -1000, 630, 251,
	-1000, 373, -77, -1000, 180, -1000, 302, -1000, -1000, 301,
	-1000, 228, 160, 452, 508, -1000, 452, 399, 388, 630,
	630, 2, 452, 103, 44, 385, 587, 587, -1000, 307,
	156, 551, 248, 354, 630, 630, 168, 630, 630, 630,
	630, 630, 630, 630, 630, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -91, -5, -30, 630, 160, 723, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 411, 44, -1000, 453,
	29, 452, 363, 259, 259, 250, -1000, 422, 587, -1000,
	452, -1000, -1000, -1000, 120, 251, -1000, -52, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 363, 259, -1000, -1000,
	630, 630, 452, 452, -1000, 630, 243, 214, 311, 151,
	33, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, 452, 266, 266, 266, -1000, 605, 248, 630, 630,
	452, 616, -1000, 349, 128, 128, 128, 84, 84, -1000,
	-1000, -1000, -1000, -1000, -1000, -7, -1000, -10, 44, -12,
	64, -1000, 587, 105, 248, 242, 91, -15, -1000, 422,
	398, 418, 160, 300, -1000, -1000, 299, -1000, -1000, 184,
	452, 452, 452, 428, 103, 103, -1000, -1000, 152, 138,
	176, 161, 155, 50, -1000, 298, -29, 96, -1000, -1000,
	-1000, -1000, 452, 310, 630, -1000, -1000, -1000, -17, -1000,
	15, -1000, 630, 4, 677, -1000, 321, 188, -1000, -1000,
	-1000, 259, 398, -1000, 630, 630, -1000, -1000, 425, 414,
	214, 65, -1000, 94, -1000, 80, -1000, -1000, -1000, -1000,
	-38, -40, -41, -1000, -1000, -1000, -1000, -1000, 630, 452,
	-1000, -1000, 452, 630, 630, 330, 248, -1000, -1000, 393,
	187, -1000, 293, -1000, 422, 587, 630, 587, -1000, -1000,
	247, 241, 218, 452, 452, 452, 448, -1000, 630, 630,
	-1000, -1000, -1000, 398, 160, 186, -1000, 251, 251, 251,
	259, 452, -1000, 320, -18, -1000, -35, -37, 184, -1000,
	443, 358, -1000, 251, -1000, -1000, -1000, 251, -1000, 251,
	-1000,
}
var yyPgo = [...]int{

	0, 577, 576, 21, 575, 574, 573, 567, 560, 559,
	558, 556, 426, 555, 554, 552, 20, 25, 545, 544,
	5, 543, 13, 540, 539, 195, 525, 3, 17, 7,
	524, 517, 18, 516, 2, 15, 9, 515, 512, 11,
	511, 8, 510, 509, 16, 504, 503, 502, 501, 12,
	500, 6, 497, 1, 496, 24, 491, 14, 4, 23,
	259, 489, 487, 486, 483, 482, 481, 0, 19, 480,
	10, 479, 469, 468, 467, 466, 465, 105, 34, 463,
	462, 461, 459,
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
	34, 34, 34, 34, 34, 34, 34, 34, 34, 38,
	38, 40, 40, 40, 42, 45, 45, 43, 43, 44,
	44, 46, 46, 41, 41, 33, 33, 33, 33, 33,
	33, 47, 47, 48, 48, 49, 49, 50, 50, 51,
	52, 52, 52, 53, 53, 53, 54, 54, 54, 55,
	55, 56, 56, 57, 57, 58, 58, 59, 60, 60,
	61, 61, 62, 62, 63, 63, 63, 63, 63, 64,
	64, 65, 65, 66, 66, 67, 68, 69, 69, 69,
	69, 69, 70,
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
	3, 3, 2, 3, 4, 5, 4, 4, 1, 1,
	1, 1, 1, 1, 5, 0, 1, 1, 2, 4,
	4, 0, 2, 1, 3, 1, 1, 1, 1, 1,
	1, 0, 3, 0, 2, 0, 3, 1, 3, 2,
	0, 1, 1, 0, 2, 4, 0, 2, 4, 0,
	3, 1, 3, 0, 5, 1, 3, 3, 0, 2,
	0, 3, 0, 1, 1, 1, 1, 1, 1, 0,
	1, 0, 1, 0, 2, 1, 1, 1, 1, 1,
	1, 1, 0,
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
	-33, -41, -35, -40, -67, -38, -42, 25, 19, 20,
	41, 42, 43, 30, -39, 81, 82, 53, 115, 33,
	88, 44, -25, 38, 86, -25, 63, 40, 50, 86,
	40, 77, -67, -70, 40, -70, 113, 40, 25, 74,
	-67, -25, -20, -34, 49, -78, -34, -77, -79, 9,
	26, -36, -34, 9, 63, -18, 75, 76, -67, 24,
	86, -31, 26, 77, 28, 29, 27, 78, 79, 80,
	81, 82, 83, 84, 85, 50, 51, 52, 45, 46,
	47, 48, -20, -29, -36, -3, -68, -20, -34, -69,
	54, 56, 57, 55, 58, -34, 49, 49, -39, 49,
	-45, -34, -55, 38, 49, -58, 40, -28, 10, -59,
	-34, -67, -70, 25, -66, 117, -63, 109, 107, 37,
	108, 13, 40, 40, 40, -70, -55, 38, -78, -80,
	9, 26, -34, -34, 118, 63, -21, -22, -24, 49,
	40, -39, 105, 101, -17, 23, -20, -20, -67, -68,
	83, -34, 21, 18, 17, -35, 49, 26, 28, 29,
	-34, -34, 30, 77, -34, -34, -34, -34, -34, -34,
	-34, -34, 118, 118, 118, -36, 118, -16, 22, -16,
	-43, -44, 89, -32, 33, -3, -58, -56, -41, -28,
	-49, 13, -20, 74, -67, -70, -64, 113, -32, -58,
	-34, -34, -34, -28, 63, -23, 64, 65, 66, 67,
	68, 70, 71, -19, 40, 24, -22, 86, -39, -39,
	-39, -35, -34, -34, 75, 30, 118, 118, -16, 118,
	-46, -44, 91, -29, -34, -57, 74, -37, -35, -57,
	118, 63, -49, -53, 15, 14, 40, 40, -47, 11,
	-22, -22, 64, 69, 64, 69, 64, 64, 64, -26,
	72, 114, 73, 40, 118, 40, 105, 101, 75, -34,
	118, 92, -34, 90, 90, 35, 63, -41, -53, -34,
	-50, -51, -34, -70, -48, 12, 14, 74, 64, 64,
	111, 111, 111, -34, -34, -34, 36, -35, 63, 63,
	-52, 31, 32, -49, -20, -36, -29, 49, 49, 49,
	7, -34, -51, -53, -27, -67, -27, -27, -58, -54,
	16, 39, 118, 63, 118, 118, 7, 26, -67, -67,
	-67,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 231, 222, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 220, 0, 0, 232, 0, 0,
	223, 0, 218, 0, 218, 0, 112, 112, 115, 0,
	0, 38, 0, 235, 19, 59, 0, 64, 55, 0,
	0, 102, 0, 26, 0, 215, 0, 183, 235, 0,
	0, 0, 242, 0, 242, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 112, 0, 115, 0, 0,
	17, 65, 67, 73, 235, 71, 72, 117, 0, 0,
	151, 152, 153, 0, 183, 0, 168, 0, 185, 186,
	187, 188, 189, 190, 147, 171, 172, 173, 169, 170,
	175, 57, 209, 0, 0, 110, 0, 27, 0, 0,
	242, 0, 233, 46, 0, 49, 0, 51, 219, 0,
	242, 209, 113, 114, 0, 41, 116, 112, 34, 0,
	0, 0, 149, 0, 0, 68, 0, 0, 74, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 135, 136, 137, 138, 139,
	140, 141, 120, 71, 0, 0, 0, 0, -2, 236,
	237, 238, 239, 240, 241, 162, 0, 0, 134, 0,
	0, 176, 0, 0, 0, 110, 103, 195, 0, 216,
	217, 184, 44, 221, 0, 0, 242, 229, 224, 225,
	226, 227, 228, 50, 52, 53, 0, 0, 42, 43,
	0, 0, 32, 33, 31, 0, 110, 77, 83, 0,
	95, 97, 98, 99, 66, 69, 118, 119, 75, 76,
	70, 122, 0, 0, 0, 123, 0, 0, 0, 0,
	128, 0, 132, 0, 154, 155, 156, 157, 158, 159,
	160, 161, 121, 146, 148, 0, 163, 0, 0, 0,
	181, 177, 0, 213, 0, 143, 213, 0, 211, 195,
	203, 0, 111, 0, 234, 47, 0, 230, 22, 23,
	35, 36, 150, 191, 0, 0, 86, 87, 0, 0,
	0, 0, 0, 104, 84, 0, 0, 0, 124, 125,
	126, 127, 129, 0, 0, 133, 166, 164, 0, 167,
	0, 178, 0, 71, 72, 20, 0, 142, 144, 21,
	210, 0, 203, 25, 0, 0, 242, 48, 193, 0,
	78, 81, 88, 0, 90, 0, 92, 93, 94, 79,
	0, 0, 0, 85, 80, 96, 100, 101, 0, 130,
	165, 174, 182, 0, 0, 0, 0, 212, 24, 204,
	196, 197, 200, 45, 195, 0, 0, 0, 89, 91,
	0, 0, 0, 131, 179, 180, 0, 145, 0, 0,
	199, 201, 202, 203, 194, 192, -2, 0, 0, 0,
	0, 205, 198, 206, 0, 108, 0, 0, 214, 18,
	0, 0, 105, 0, 106, 107, 207, 0, 109, 0,
	208,
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
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:898
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 166:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:902
		{
			yyVAL.valExpr = &CtorExpr{Name: yyDollar[2].str, Exprs: yyDollar[3].valExprs}
		}
	case 167:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:906
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 168:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:910
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 169:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:916
		{
			yyVAL.bytes = IF_BYTES
		}
	case 170:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:920
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 171:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:926
		{
			yyVAL.byt = AST_UPLUS
		}
	case 172:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:930
		{
			yyVAL.byt = AST_UMINUS
		}
	case 173:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:934
		{
			yyVAL.byt = AST_TILDA
		}
	case 174:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:940
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 175:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:945
		{
			yyVAL.valExpr = nil
		}
	case 176:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:949
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 177:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:955
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 178:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:959
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 179:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:965
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 180:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:969
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 181:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:974
		{
			yyVAL.valExpr = nil
		}
	case 182:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:978
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 183:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:984
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 184:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:988
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 185:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:994
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 186:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:998
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 187:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1002
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1006
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1010
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1014
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 191:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1019
		{
			yyVAL.valExprs = nil
		}
	case 192:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 193:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1028
		{
			yyVAL.expr = nil
		}
	case 194:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1032
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 195:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1037
		{
			yyVAL.orderBy = nil
		}
	case 196:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1041
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1047
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 198:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1051
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 199:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1057
		{
			yyVAL.order = &Order{Expr: yyDollar[1].valExpr, Direction: yyDollar[2].str}
		}
	case 200:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1062
		{
			yyVAL.str = AST_ASC
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1066
		{
			yyVAL.str = AST_ASC
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1070
		{
			yyVAL.str = AST_DESC
		}
	case 203:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1075
		{
			yyVAL.limit = nil
		}
	case 204:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1079
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 205:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1083
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 206:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1088
		{
			yyVAL.str = ""
		}
	case 207:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1092
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 208:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1096
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
	case 209:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1109
		{
			yyVAL.columns = nil
		}
	case 210:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1113
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1119
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 212:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1123
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 213:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1128
		{
			yyVAL.updateExprs = nil
		}
	case 214:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1132
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1138
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 216:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1142
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 217:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1148
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 218:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1153
		{
			yyVAL.empty = struct{}{}
		}
	case 219:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1155
		{
			yyVAL.empty = struct{}{}
		}
	case 220:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1158
		{
			yyVAL.empty = struct{}{}
		}
	case 221:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1160
		{
			yyVAL.empty = struct{}{}
		}
	case 222:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1163
		{
			yyVAL.empty = struct{}{}
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1165
		{
			yyVAL.empty = struct{}{}
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1169
		{
			yyVAL.empty = struct{}{}
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1171
		{
			yyVAL.empty = struct{}{}
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1173
		{
			yyVAL.empty = struct{}{}
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1175
		{
			yyVAL.empty = struct{}{}
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1177
		{
			yyVAL.empty = struct{}{}
		}
	case 229:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1180
		{
			yyVAL.empty = struct{}{}
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1182
		{
			yyVAL.empty = struct{}{}
		}
	case 231:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1185
		{
			yyVAL.empty = struct{}{}
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1187
		{
			yyVAL.empty = struct{}{}
		}
	case 233:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1190
		{
			yyVAL.empty = struct{}{}
		}
	case 234:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1192
		{
			yyVAL.empty = struct{}{}
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1196
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1202
		{
			yyVAL.str = yyDollar[1].str
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1208
		{
			yyVAL.str = AST_DATE
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1212
		{
			yyVAL.str = AST_TIME
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1216
		{
			yyVAL.str = AST_TIMESTAMP
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1220
		{
			yyVAL.str = AST_DATETIME
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1224
		{
			yyVAL.str = AST_YEAR
		}
	case 242:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1229
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
