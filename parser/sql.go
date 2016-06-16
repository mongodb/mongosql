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
	-1, 203,
	75, 72,
	76, 72,
	-2, 149,
	-1, 432,
	75, 71,
	76, 71,
	-2, 82,
}

const yyNprod = 248
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 935

var yyAct = [...]int{

	114, 199, 106, 440, 73, 103, 407, 105, 111, 313,
	369, 127, 304, 255, 361, 306, 220, 225, 146, 112,
	201, 200, 3, 75, 100, 292, 101, 34, 35, 36,
	37, 233, 449, 62, 449, 91, 329, 330, 331, 332,
	333, 87, 334, 335, 77, 80, 320, 82, 449, 167,
	84, 367, 76, 167, 88, 64, 167, 49, 418, 50,
	97, 167, 167, 167, 253, 253, 253, 44, 149, 46,
	52, 53, 54, 47, 417, 416, 81, 83, 51, 391,
	386, 388, 145, 98, 95, 397, 258, 399, 294, 451,
	153, 450, 390, 94, 78, 257, 156, 305, 155, 159,
	340, 164, 165, 148, 171, 448, 396, 305, 366, 358,
	355, 239, 203, 354, 197, 202, 210, 198, 353, 352,
	350, 349, 293, 252, 387, 173, 18, 144, 142, 137,
	413, 158, 216, 169, 170, 237, 219, 77, 240, 139,
	77, 223, 393, 229, 228, 76, 392, 362, 76, 261,
	362, 316, 74, 260, 56, 58, 59, 57, 61, 280,
	165, 258, 230, 227, 152, 250, 251, 63, 415, 414,
	257, 244, 243, 266, 229, 264, 265, 269, 259, 139,
	278, 279, 160, 282, 283, 284, 285, 286, 287, 288,
	289, 290, 291, 267, 262, 273, 246, 424, 185, 186,
	188, 189, 187, 295, 165, 384, 281, 236, 238, 235,
	268, 172, 180, 181, 182, 183, 184, 185, 186, 188,
	189, 187, 77, 77, 261, 383, 309, 63, 260, 70,
	76, 311, 315, 382, 317, 253, 297, 299, 300, 301,
	302, 312, 308, 226, 425, 226, 77, 402, 427, 428,
	322, 323, 324, 318, 76, 166, 325, 86, 141, 380,
	321, 435, 169, 170, 381, 378, 308, 121, 122, 259,
	379, 339, 326, 120, 245, 221, 434, 165, 126, 345,
	346, 132, 341, 342, 343, 222, 222, 433, 104, 123,
	124, 125, 169, 170, 274, 344, 327, 109, 139, 135,
	217, 130, 138, 215, 18, 19, 20, 21, 360, 167,
	214, 202, 89, 359, 116, 117, 357, 213, 212, 211,
	154, 99, 368, 351, 365, 108, 134, 364, 63, 128,
	129, 102, 78, 338, 115, 389, 373, 22, 133, 259,
	259, 376, 377, 34, 35, 36, 37, 446, 394, 337,
	395, 180, 181, 182, 183, 184, 185, 186, 188, 189,
	187, 398, 372, 242, 241, 131, 224, 71, 77, 150,
	447, 136, 147, 405, 143, 140, 403, 408, 347, 404,
	85, 180, 181, 182, 183, 184, 185, 186, 188, 189,
	187, 409, 422, 401, 27, 28, 29, 419, 30, 32,
	31, 18, 420, 421, 90, 69, 348, 23, 24, 26,
	25, 453, 92, 231, 431, 165, 151, 430, 263, 202,
	429, 432, 423, 275, 65, 276, 277, 437, 67, 307,
	93, 408, 438, 370, 441, 441, 441, 77, 442, 443,
	439, 444, 248, 412, 371, 76, 121, 122, 63, 298,
	454, 314, 120, 411, 455, 375, 456, 126, 226, 249,
	132, 38, 205, 208, 206, 207, 209, 104, 123, 124,
	125, 329, 330, 331, 332, 333, 109, 334, 335, 162,
	130, 40, 41, 42, 43, 96, 72, 18, 452, 18,
	436, 39, 55, 116, 117, 60, 163, 247, 161, 17,
	16, 121, 122, 15, 108, 14, 13, 120, 128, 129,
	102, 12, 126, 115, 204, 132, 232, 133, 45, 319,
	234, 48, 78, 123, 124, 125, 79, 310, 445, 426,
	406, 109, 410, 374, 356, 130, 205, 208, 206, 207,
	209, 18, 218, 303, 131, 119, 113, 296, 116, 117,
	118, 363, 110, 174, 107, 121, 122, 385, 256, 108,
	328, 254, 336, 128, 129, 168, 126, 66, 115, 132,
	33, 68, 133, 11, 10, 9, 78, 123, 124, 125,
	8, 7, 6, 5, 4, 157, 2, 1, 0, 130,
	205, 208, 206, 207, 209, 0, 0, 0, 0, 131,
	0, 0, 116, 117, 0, 0, 0, 272, 271, 121,
	122, 270, 0, 0, 0, 0, 0, 128, 129, 0,
	126, 0, 115, 132, 0, 0, 133, 0, 0, 0,
	78, 123, 124, 125, 0, 0, 0, 0, 0, 157,
	0, 0, 0, 130, 183, 184, 185, 186, 188, 189,
	187, 0, 0, 131, 0, 0, 116, 117, 0, 0,
	0, 0, 0, 121, 122, 0, 0, 0, 0, 120,
	0, 128, 129, 0, 126, 0, 115, 132, 0, 0,
	133, 0, 0, 0, 78, 123, 124, 125, 0, 0,
	0, 0, 0, 109, 0, 0, 0, 130, 0, 0,
	0, 0, 0, 18, 0, 0, 0, 131, 0, 0,
	116, 117, 0, 0, 0, 0, 0, 121, 122, 0,
	0, 108, 0, 0, 0, 128, 129, 0, 126, 0,
	115, 132, 0, 0, 133, 0, 0, 0, 78, 123,
	124, 125, 0, 0, 0, 0, 0, 157, 0, 0,
	0, 130, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 131, 0, 0, 116, 117, 0, 0, 0, 0,
	0, 121, 122, 0, 0, 0, 0, 0, 0, 128,
	129, 0, 126, 0, 115, 132, 0, 0, 133, 0,
	0, 0, 78, 123, 124, 125, 0, 0, 0, 0,
	0, 157, 0, 0, 0, 130, 175, 179, 177, 178,
	0, 0, 0, 0, 0, 131, 0, 0, 116, 117,
	0, 0, 0, 0, 0, 193, 194, 195, 196, 0,
	190, 191, 192, 128, 129, 0, 0, 0, 115, 0,
	0, 0, 133, 180, 181, 182, 183, 184, 185, 186,
	188, 189, 187, 0, 0, 0, 0, 176, 180, 181,
	182, 183, 184, 185, 186, 188, 189, 187, 0, 131,
	0, 0, 400, 175, 179, 177, 178, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 193, 194, 195, 196, 0, 190, 191, 192,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 176, 180, 181, 182, 183, 184,
	185, 186, 188, 189, 187,
}
var yyPact = [...]int{

	299, -1000, -1000, 284, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -45, -57, -34, -42, -1000, -1000, -1000,
	-1000, 52, 288, 484, 403, -1000, -1000, -1000, 406, -1000,
	371, 327, 477, 54, -72, -37, 288, -1000, -35, 288,
	-1000, 340, -76, 288, -76, 370, 402, 402, 476, 288,
	-24, -1000, 272, -1000, -1000, -1000, 248, -1000, 282, 327,
	333, 41, 327, 116, 335, -1000, 208, -1000, 40, 334,
	50, 288, -1000, 332, -1000, -47, 329, 391, 90, 288,
	327, -1000, 644, 752, -1000, 402, 752, 476, 470, 752,
	246, -1000, -1000, 187, 37, -1000, 847, -1000, 644, 482,
	-1000, -1000, -1000, 752, 270, 269, 268, 261, 254, -1000,
	251, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, 752, -1000, 237, 292, 326, 448, 292,
	-1000, 752, 288, -1000, 388, -88, -1000, 98, -1000, 324,
	-1000, -1000, 323, -1000, 236, 58, 765, 536, -1000, 765,
	402, 433, 752, 752, 3, 765, 46, 248, 395, 644,
	644, -1000, 408, 127, 590, 245, 397, 752, 752, 129,
	752, 752, 752, 752, 752, 752, 752, 752, 752, 752,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -95, 2,
	-32, 752, 58, 847, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, 427, 248, 248, 248, 248, -1000, 484, 6, 765,
	396, 292, 292, 235, -1000, 438, 644, -1000, 765, -1000,
	-1000, -1000, 77, 288, -1000, -69, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 396, 292, -1000, -1000, 752, 752,
	765, 765, -1000, 752, 233, 407, 309, 121, 12, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 765,
	251, 251, 251, -1000, 698, 245, 752, 752, 765, 303,
	-1000, 376, 563, 563, 563, 115, 115, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, 1, -1000, 0, 248, -1,
	-2, -7, -10, 16, -1000, 644, 73, 245, 284, 76,
	-12, -1000, 438, 418, 430, 58, 322, -1000, -1000, 296,
	-1000, -1000, 116, 765, 765, 765, 444, 46, 46, -1000,
	-1000, 201, 195, 169, 161, 141, 8, -1000, 295, -28,
	39, -1000, -1000, -1000, -1000, 765, 273, 752, -1000, -1000,
	-1000, -14, -1000, -1000, -1000, -1000, -9, -1000, 752, -5,
	780, -1000, 358, 184, -1000, -1000, -1000, 292, 418, -1000,
	752, 644, -1000, -1000, 441, 429, 407, 56, -1000, 105,
	-1000, 104, -1000, -1000, -1000, -1000, -38, -39, -55, -1000,
	-1000, -1000, -1000, -1000, 752, 765, -1000, -1000, 765, 752,
	752, 356, 245, -1000, -1000, 134, 181, -1000, 217, -1000,
	438, 644, 752, 644, -1000, -1000, 238, 227, 212, 765,
	765, 765, 483, -1000, 752, 644, -1000, -1000, -1000, 418,
	58, 172, -1000, 288, 288, 288, 292, 765, -1000, 331,
	-15, -1000, -29, -31, 116, -1000, 481, 385, -1000, 288,
	-1000, -1000, -1000, 288, -1000, 288, -1000,
}
var yyPgo = [...]int{

	0, 587, 586, 21, 584, 583, 582, 581, 580, 575,
	574, 573, 461, 571, 570, 567, 24, 26, 565, 562,
	5, 561, 13, 560, 558, 229, 557, 3, 17, 7,
	554, 553, 15, 552, 2, 19, 1, 551, 550, 11,
	546, 8, 545, 543, 12, 542, 534, 533, 532, 9,
	530, 6, 529, 10, 528, 16, 527, 14, 4, 23,
	257, 526, 521, 520, 519, 518, 516, 0, 20, 514,
	18, 511, 506, 505, 503, 500, 499, 84, 35, 498,
	497, 495, 491,
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
	34, 34, 34, 34, 38, 38, 40, 40, 40, 42,
	45, 45, 43, 43, 44, 44, 46, 46, 41, 41,
	33, 33, 33, 33, 33, 33, 47, 47, 48, 48,
	49, 49, 50, 50, 51, 52, 52, 52, 53, 53,
	53, 54, 54, 54, 55, 55, 56, 56, 57, 57,
	58, 58, 59, 60, 60, 61, 61, 62, 62, 63,
	63, 63, 63, 63, 64, 64, 65, 65, 66, 66,
	67, 68, 69, 69, 69, 69, 69, 70,
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
	3, 3, 3, 3, 2, 3, 4, 4, 4, 4,
	5, 4, 4, 1, 1, 1, 1, 1, 1, 5,
	0, 1, 1, 2, 4, 4, 0, 2, 1, 3,
	1, 1, 1, 1, 1, 1, 0, 3, 0, 2,
	0, 3, 1, 3, 2, 0, 1, 1, 0, 2,
	4, 0, 2, 4, 0, 3, 1, 3, 0, 5,
	1, 3, 3, 0, 2, 0, 3, 0, 1, 1,
	1, 1, 1, 1, 0, 1, 0, 1, 0, 2,
	1, 1, 1, 1, 1, 1, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -71, -72, -73, -74, -75, -76, 5, 6,
	7, 8, 38, 108, 109, 111, 110, 95, 96, 97,
	99, 101, 100, -14, 59, 60, 61, 62, -12, -82,
	-12, -12, -12, -12, 112, -65, 114, 118, -62, 114,
	116, 112, 112, 113, 114, -12, 102, 105, 103, 104,
	-81, 106, -67, 40, -3, 21, -15, 22, -13, 34,
	-25, 40, 9, -58, 98, -59, -41, -67, 40, -61,
	117, 113, -67, 112, -67, 40, -60, 117, -67, -60,
	34, -78, 10, 28, -78, -77, 9, -67, 107, 49,
	-16, -17, 83, -20, 40, -29, -34, -30, 77, 49,
	-33, -41, -35, -40, -67, 86, 66, 67, -38, -42,
	25, 19, 20, 41, 42, 43, 30, -39, 81, 82,
	53, 117, 33, 90, 44, -25, 38, 88, -25, 63,
	40, 50, 88, 40, 77, -67, -70, 40, -70, 115,
	40, 25, 74, -67, -25, -20, -34, 49, -78, -34,
	-77, -79, 9, 26, -36, -34, 9, 63, -18, 75,
	76, -67, 24, 88, -31, 26, 77, 28, 29, 27,
	78, 79, 80, 81, 82, 83, 84, 87, 85, 86,
	50, 51, 52, 45, 46, 47, 48, -20, -29, -36,
	-3, -68, -20, -34, -69, 54, 56, 57, 55, 58,
	-34, 49, 49, 49, 49, 49, -39, 49, -45, -34,
	-55, 38, 49, -58, 40, -28, 10, -59, -34, -67,
	-70, 25, -66, 119, -63, 111, 109, 37, 110, 13,
	40, 40, 40, -70, -55, 38, -78, -80, 9, 26,
	-34, -34, 120, 63, -21, -22, -24, 49, 40, -39,
	107, 103, -17, 23, -20, -20, -67, -68, 83, -34,
	21, 18, 17, -35, 49, 26, 28, 29, -34, -34,
	30, 77, -34, -34, -34, -34, -34, -34, -34, -34,
	-34, -34, 120, 120, 120, -36, 120, -16, 22, -16,
	-16, -16, -16, -43, -44, 91, -32, 33, -3, -58,
	-56, -41, -28, -49, 13, -20, 74, -67, -70, -64,
	115, -32, -58, -34, -34, -34, -28, 63, -23, 64,
	65, 66, 67, 68, 70, 71, -19, 40, 24, -22,
	88, -39, -39, -39, -35, -34, -34, 75, 30, 120,
	120, -16, 120, 120, 120, 120, -46, -44, 93, -29,
	-34, -57, 74, -37, -35, -57, 120, 63, -49, -53,
	15, 14, 40, 40, -47, 11, -22, -22, 64, 69,
	64, 69, 64, 64, 64, -26, 72, 116, 73, 40,
	120, 40, 107, 103, 75, -34, 120, 94, -34, 92,
	92, 35, 63, -41, -53, -34, -50, -51, -20, -70,
	-48, 12, 14, 74, 64, 64, 113, 113, 113, -34,
	-34, -34, 36, -35, 63, 63, -52, 31, 32, -49,
	-20, -36, -29, 49, 49, 49, 7, -34, -51, -53,
	-27, -67, -27, -27, -58, -54, 16, 39, 120, 63,
	120, 120, 7, 26, -67, -67, -67,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 54, 54,
	54, 54, 54, 236, 227, 0, 0, 28, 29, 30,
	54, 37, 0, 0, 58, 60, 61, 62, 63, 56,
	0, 0, 0, 0, 225, 0, 0, 237, 0, 0,
	228, 0, 223, 0, 223, 0, 112, 112, 115, 0,
	0, 38, 0, 240, 19, 59, 0, 64, 55, 0,
	0, 102, 0, 26, 0, 220, 0, 188, 240, 0,
	0, 0, 247, 0, 247, 0, 0, 0, 0, 0,
	0, 39, 0, 0, 40, 112, 0, 115, 0, 0,
	17, 65, 67, 73, 240, 71, 72, 117, 0, 0,
	151, 152, 153, 0, 188, 0, 0, 0, 0, 173,
	0, 190, 191, 192, 193, 194, 195, 147, 176, 177,
	178, 174, 175, 180, 57, 214, 0, 0, 110, 0,
	27, 0, 0, 247, 0, 238, 46, 0, 49, 0,
	51, 224, 0, 247, 214, 113, 114, 0, 41, 116,
	112, 34, 0, 0, 0, 149, 0, 0, 68, 0,
	0, 74, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	135, 136, 137, 138, 139, 140, 141, 120, 71, 0,
	0, 0, 0, -2, 241, 242, 243, 244, 245, 246,
	164, 0, 0, 0, 0, 0, 134, 0, 0, 181,
	0, 0, 0, 110, 103, 200, 0, 221, 222, 189,
	44, 226, 0, 0, 247, 234, 229, 230, 231, 232,
	233, 50, 52, 53, 0, 0, 42, 43, 0, 0,
	32, 33, 31, 0, 110, 77, 83, 0, 95, 97,
	98, 99, 66, 69, 118, 119, 75, 76, 70, 122,
	0, 0, 0, 123, 0, 0, 0, 0, 128, 0,
	132, 0, 154, 155, 156, 157, 158, 159, 160, 161,
	162, 163, 121, 146, 148, 0, 165, 0, 0, 0,
	0, 0, 0, 186, 182, 0, 218, 0, 143, 218,
	0, 216, 200, 208, 0, 111, 0, 239, 47, 0,
	235, 22, 23, 35, 36, 150, 196, 0, 0, 86,
	87, 0, 0, 0, 0, 0, 104, 84, 0, 0,
	0, 124, 125, 126, 127, 129, 0, 0, 133, 171,
	166, 0, 167, 168, 169, 172, 0, 183, 0, 71,
	72, 20, 0, 142, 144, 21, 215, 0, 208, 25,
	0, 0, 247, 48, 198, 0, 78, 81, 88, 0,
	90, 0, 92, 93, 94, 79, 0, 0, 0, 85,
	80, 96, 100, 101, 0, 130, 170, 179, 187, 0,
	0, 0, 0, 217, 24, 209, 201, 202, 205, 45,
	200, 0, 0, 0, 89, 91, 0, 0, 0, 131,
	184, 185, 0, 145, 0, 0, 204, 206, 207, 208,
	199, 197, -2, 0, 0, 0, 0, 210, 203, 211,
	0, 108, 0, 0, 219, 18, 0, 0, 105, 0,
	106, 107, 212, 0, 109, 0, 213,
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
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_IDIV, Right: yyDollar[3].valExpr}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:875
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 163:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:879
		{
			yyVAL.valExpr = &BinaryExpr{Left: yyDollar[1].valExpr, Operator: AST_MOD, Right: yyDollar[3].valExpr}
		}
	case 164:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:883
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
	case 165:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:898
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 166:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:902
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 167:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:907
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("mod"), Exprs: yyDollar[3].selectExprs}
		}
	case 168:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:911
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("left"), Exprs: yyDollar[3].selectExprs}
		}
	case 169:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:915
		{
			yyVAL.valExpr = &FuncExpr{Name: []byte("right"), Exprs: yyDollar[3].selectExprs}
		}
	case 170:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:919
		{
			yyVAL.valExpr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 171:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:923
		{
			yyVAL.valExpr = &CtorExpr{Name: yyDollar[2].str, Exprs: yyDollar[3].valExprs}
		}
	case 172:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:927
		{
			yyVAL.valExpr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 173:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:931
		{
			yyVAL.valExpr = yyDollar[1].caseExpr
		}
	case 174:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:937
		{
			yyVAL.bytes = IF_BYTES
		}
	case 175:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:941
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 176:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:947
		{
			yyVAL.byt = AST_UPLUS
		}
	case 177:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:951
		{
			yyVAL.byt = AST_UMINUS
		}
	case 178:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:955
		{
			yyVAL.byt = AST_TILDA
		}
	case 179:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:961
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].valExpr, Whens: yyDollar[3].whens, Else: yyDollar[4].valExpr}
		}
	case 180:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:966
		{
			yyVAL.valExpr = nil
		}
	case 181:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:970
		{
			yyVAL.valExpr = yyDollar[1].valExpr
		}
	case 182:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:976
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 183:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:980
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 184:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:986
		{
			yyVAL.when = &When{Cond: yyDollar[2].boolExpr, Val: yyDollar[4].valExpr}
		}
	case 185:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:990
		{
			yyVAL.when = &When{Cond: yyDollar[2].valExpr, Val: yyDollar[4].valExpr}
		}
	case 186:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:995
		{
			yyVAL.valExpr = nil
		}
	case 187:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:999
		{
			yyVAL.valExpr = yyDollar[2].valExpr
		}
	case 188:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1005
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 189:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1009
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 190:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1015
		{
			yyVAL.valExpr = &TrueVal{}
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1019
		{
			yyVAL.valExpr = &FalseVal{}
		}
	case 192:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.valExpr = StrVal(yyDollar[1].bytes)
		}
	case 193:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1027
		{
			yyVAL.valExpr = NumVal(yyDollar[1].bytes)
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1031
		{
			yyVAL.valExpr = ValArg(yyDollar[1].bytes)
		}
	case 195:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1035
		{
			yyVAL.valExpr = &NullVal{}
		}
	case 196:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1040
		{
			yyVAL.valExprs = nil
		}
	case 197:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1044
		{
			yyVAL.valExprs = yyDollar[3].valExprs
		}
	case 198:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1049
		{
			yyVAL.expr = nil
		}
	case 199:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1053
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 200:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1058
		{
			yyVAL.orderBy = nil
		}
	case 201:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1062
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1068
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 203:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1072
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 204:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1078
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 205:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1083
		{
			yyVAL.str = AST_ASC
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1087
		{
			yyVAL.str = AST_ASC
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1091
		{
			yyVAL.str = AST_DESC
		}
	case 208:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1096
		{
			yyVAL.limit = nil
		}
	case 209:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1100
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].valExpr}
		}
	case 210:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1104
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].valExpr, Rowcount: yyDollar[4].valExpr}
		}
	case 211:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1109
		{
			yyVAL.str = ""
		}
	case 212:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1113
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 213:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1117
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
	case 214:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1130
		{
			yyVAL.columns = nil
		}
	case 215:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1134
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1140
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 217:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1144
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 218:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1149
		{
			yyVAL.updateExprs = nil
		}
	case 219:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1153
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1159
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 221:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1163
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 222:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1169
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].valExpr}
		}
	case 223:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1174
		{
			yyVAL.empty = struct{}{}
		}
	case 224:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1176
		{
			yyVAL.empty = struct{}{}
		}
	case 225:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1179
		{
			yyVAL.empty = struct{}{}
		}
	case 226:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1181
		{
			yyVAL.empty = struct{}{}
		}
	case 227:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1184
		{
			yyVAL.empty = struct{}{}
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1186
		{
			yyVAL.empty = struct{}{}
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1190
		{
			yyVAL.empty = struct{}{}
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1192
		{
			yyVAL.empty = struct{}{}
		}
	case 231:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1194
		{
			yyVAL.empty = struct{}{}
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1196
		{
			yyVAL.empty = struct{}{}
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1198
		{
			yyVAL.empty = struct{}{}
		}
	case 234:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1201
		{
			yyVAL.empty = struct{}{}
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1203
		{
			yyVAL.empty = struct{}{}
		}
	case 236:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1206
		{
			yyVAL.empty = struct{}{}
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1208
		{
			yyVAL.empty = struct{}{}
		}
	case 238:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1211
		{
			yyVAL.empty = struct{}{}
		}
	case 239:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1213
		{
			yyVAL.empty = struct{}{}
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1217
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1223
		{
			yyVAL.str = yyDollar[1].str
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1229
		{
			yyVAL.str = AST_DATE
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1233
		{
			yyVAL.str = AST_TIME
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1237
		{
			yyVAL.str = AST_TIMESTAMP
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1241
		{
			yyVAL.str = AST_DATETIME
		}
	case 246:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1245
		{
			yyVAL.str = AST_YEAR
		}
	case 247:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1250
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
