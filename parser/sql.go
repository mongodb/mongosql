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
	SHARE                    = []byte("share")
	MODE                     = []byte("mode")
	IF_BYTES                 = []byte("if")
	VALUES_BYTES             = []byte("values")
	RIGHT_BYTES              = []byte("right")
	LEFT_BYTES               = []byte("left")
	MOD_BYTES                = []byte("mod")
	YEAR_BYTES               = []byte("year")
	QUARTER_BYTES            = []byte("quarter")
	MONTH_BYTES              = []byte("month")
	WEEK_BYTES               = []byte("week")
	DAY_BYTES                = []byte("day")
	HOUR_BYTES               = []byte("hour")
	MINUTE_BYTES             = []byte("minute")
	SECOND_BYTES             = []byte("second")
	MICROSECOND_BYTES        = []byte("microsecond")
	SECOND_MICROSECOND_BYTES = []byte("second_microsecond")
	MINUTE_MICROSECOND_BYTES = []byte("minute_microsecond")
	MINUTE_SECOND_BYTES      = []byte("minute_second")
	HOUR_MICROSECOND_BYTES   = []byte("hour_microsecond")
	HOUR_SECOND_BYTES        = []byte("hour_second")
	HOUR_MINUTE_BYTES        = []byte("hour_minute")
	DAY_MICROSECOND_BYTES    = []byte("day_microsecond")
	DAY_SECOND_BYTES         = []byte("day_second")
	DAY_MINUTE_BYTES         = []byte("day_minute")
	DAY_HOUR_BYTES           = []byte("day_hour")
	YEAR_MONTH_BYTES         = []byte("year_month")
	CHAR_BYTES               = []byte("char")
	DATE_BYTES               = []byte("date")
	DATETIME_BYTES           = []byte("datetime")
	FLOAT_BYTES              = []byte("float")
	INTEGER_BYTES            = []byte("integer")
)

//line sql.y:59
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
const LBRACE = 57386
const RBRACE = 57387
const TILDE = 57388
const DATE = 57389
const DATETIME = 57390
const TIME = 57391
const TIMESTAMP = 57392
const CURRENT_TIMESTAMP = 57393
const TIMESTAMPADD = 57394
const TIMESTAMPDIFF = 57395
const YEAR = 57396
const QUARTER = 57397
const MONTH = 57398
const WEEK = 57399
const DAY = 57400
const HOUR = 57401
const MINUTE = 57402
const SECOND = 57403
const MICROSECOND = 57404
const EXTRACT = 57405
const DATE_ADD = 57406
const DATE_SUB = 57407
const INTERVAL = 57408
const STR_TO_DATE = 57409
const SQL_TSI_YEAR = 57410
const SQL_TSI_QUARTER = 57411
const SQL_TSI_MONTH = 57412
const SQL_TSI_WEEK = 57413
const SQL_TSI_DAY = 57414
const SQL_TSI_HOUR = 57415
const SQL_TSI_MINUTE = 57416
const SQL_TSI_SECOND = 57417
const CONVERT = 57418
const CAST = 57419
const CHAR = 57420
const SIGNED = 57421
const UNSIGNED = 57422
const SQL_BIGINT = 57423
const SQL_VARCHAR = 57424
const SQL_DATE = 57425
const SQL_TIMESTAMP = 57426
const SQL_DOUBLE = 57427
const INTEGER = 57428
const SECOND_MICROSECOND = 57429
const MINUTE_MICROSECOND = 57430
const MINUTE_SECOND = 57431
const HOUR_MICROSECOND = 57432
const HOUR_SECOND = 57433
const HOUR_MINUTE = 57434
const DAY_MICROSECOND = 57435
const DAY_SECOND = 57436
const DAY_MINUTE = 57437
const DAY_HOUR = 57438
const YEAR_MONTH = 57439
const FROM = 57440
const UNION = 57441
const MINUS = 57442
const EXCEPT = 57443
const INTERSECT = 57444
const COMMA = 57445
const JOIN = 57446
const STRAIGHT_JOIN = 57447
const LEFT = 57448
const RIGHT = 57449
const INNER = 57450
const OUTER = 57451
const CROSS = 57452
const NATURAL = 57453
const USE = 57454
const FORCE = 57455
const ON = 57456
const NOT = 57457
const OR = 57458
const XOR = 57459
const AND = 57460
const BETWEEN = 57461
const NE = 57462
const EQ = 57463
const NULL_SAFE_EQUAL = 57464
const IS = 57465
const LIKE = 57466
const REGEXP = 57467
const IN = 57468
const LT = 57469
const GT = 57470
const LE = 57471
const GE = 57472
const BIT_AND = 57473
const BIT_OR = 57474
const CARET = 57475
const PLUS = 57476
const SUB = 57477
const TIMES = 57478
const MOD = 57479
const DIV = 57480
const IDIV = 57481
const DOT = 57482
const UNARY = 57483
const CASE = 57484
const WHEN = 57485
const THEN = 57486
const ELSE = 57487
const END = 57488
const BEGIN = 57489
const COMMIT = 57490
const ROLLBACK = 57491
const NAMES = 57492
const REPLACE = 57493
const ADMIN = 57494
const SHOW = 57495
const DATABASES = 57496
const TABLES = 57497
const PROXY = 57498
const VARIABLES = 57499
const FULL = 57500
const SESSION = 57501
const GLOBAL = 57502
const COLUMNS = 57503
const CREATE = 57504
const ALTER = 57505
const DROP = 57506
const RENAME = 57507
const TABLE = 57508
const INDEX = 57509
const VIEW = 57510
const TO = 57511
const IGNORE = 57512
const IF = 57513
const UNIQUE = 57514
const USING = 57515

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
	"LBRACE",
	"RBRACE",
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
	"EXTRACT",
	"DATE_ADD",
	"DATE_SUB",
	"INTERVAL",
	"STR_TO_DATE",
	"SQL_TSI_YEAR",
	"SQL_TSI_QUARTER",
	"SQL_TSI_MONTH",
	"SQL_TSI_WEEK",
	"SQL_TSI_DAY",
	"SQL_TSI_HOUR",
	"SQL_TSI_MINUTE",
	"SQL_TSI_SECOND",
	"CONVERT",
	"CAST",
	"CHAR",
	"SIGNED",
	"UNSIGNED",
	"SQL_BIGINT",
	"SQL_VARCHAR",
	"SQL_DATE",
	"SQL_TIMESTAMP",
	"SQL_DOUBLE",
	"INTEGER",
	"SECOND_MICROSECOND",
	"MINUTE_MICROSECOND",
	"MINUTE_SECOND",
	"HOUR_MICROSECOND",
	"HOUR_SECOND",
	"HOUR_MINUTE",
	"DAY_MICROSECOND",
	"DAY_SECOND",
	"DAY_MINUTE",
	"DAY_HOUR",
	"YEAR_MONTH",
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
	"AND",
	"BETWEEN",
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
	161, 37,
	-2, 39,
}

const yyNprod = 318
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1441

var yyAct = [...]int{

	115, 455, 568, 513, 146, 189, 339, 103, 354, 383,
	110, 447, 487, 374, 246, 281, 93, 259, 89, 134,
	301, 109, 44, 251, 46, 82, 136, 102, 47, 376,
	472, 474, 49, 64, 50, 171, 265, 52, 53, 54,
	85, 390, 174, 524, 79, 523, 75, 84, 522, 83,
	86, 51, 77, 18, 90, 243, 3, 100, 263, 96,
	99, 266, 56, 58, 59, 78, 63, 61, 62, 80,
	504, 375, 284, 445, 375, 477, 410, 283, 221, 105,
	200, 203, 201, 202, 170, 284, 167, 162, 473, 66,
	283, 274, 178, 207, 208, 209, 210, 195, 196, 197,
	198, 199, 200, 203, 201, 202, 219, 198, 199, 200,
	203, 201, 202, 182, 183, 187, 587, 65, 94, 275,
	166, 169, 173, 195, 196, 197, 198, 199, 200, 203,
	201, 202, 323, 164, 519, 448, 386, 321, 322, 320,
	224, 242, 177, 188, 448, 521, 585, 147, 148, 149,
	399, 400, 401, 402, 403, 325, 404, 405, 466, 464,
	520, 470, 79, 467, 465, 79, 469, 468, 255, 34,
	35, 36, 37, 438, 180, 181, 164, 279, 184, 584,
	191, 190, 76, 78, 545, 192, 78, 262, 264, 261,
	287, 222, 223, 479, 270, 508, 286, 582, 529, 478,
	288, 440, 272, 287, 256, 372, 583, 439, 249, 286,
	432, 528, 252, 431, 269, 190, 331, 253, 285, 430,
	245, 330, 255, 252, 319, 429, 306, 308, 310, 312,
	314, 316, 437, 95, 338, 348, 327, 349, 436, 583,
	353, 352, 355, 326, 367, 368, 254, 527, 79, 79,
	97, 98, 72, 526, 333, 335, 336, 583, 192, 74,
	387, 34, 35, 36, 37, 482, 369, 276, 277, 78,
	381, 192, 79, 382, 290, 291, 292, 293, 294, 295,
	296, 297, 298, 299, 300, 305, 307, 309, 311, 313,
	315, 317, 318, 78, 379, 324, 388, 328, 329, 409,
	391, 452, 378, 532, 396, 531, 397, 192, 503, 502,
	285, 350, 351, 192, 442, 577, 576, 164, 392, 18,
	19, 20, 21, 575, 160, 192, 378, 163, 411, 441,
	428, 418, 385, 412, 427, 413, 425, 414, 371, 415,
	278, 416, 88, 417, 476, 179, 423, 574, 573, 22,
	185, 572, 556, 537, 393, 394, 536, 159, 535, 395,
	530, 453, 426, 194, 218, 205, 204, 206, 216, 215,
	217, 213, 207, 208, 209, 210, 195, 196, 197, 198,
	199, 200, 203, 201, 202, 497, 435, 444, 450, 192,
	192, 451, 454, 555, 192, 489, 192, 91, 279, 534,
	279, 419, 420, 421, 554, 399, 400, 401, 402, 403,
	553, 404, 405, 462, 463, 271, 247, 234, 533, 135,
	225, 233, 248, 248, 285, 285, 488, 490, 491, 492,
	493, 494, 495, 496, 238, 237, 236, 235, 232, 231,
	230, 229, 228, 227, 499, 500, 501, 498, 226, 101,
	370, 240, 239, 65, 79, 446, 510, 483, 484, 485,
	486, 27, 28, 29, 161, 30, 32, 31, 408, 80,
	580, 475, 459, 458, 268, 509, 23, 24, 26, 25,
	407, 150, 151, 152, 153, 154, 155, 156, 157, 158,
	581, 541, 267, 250, 515, 340, 341, 342, 343, 344,
	345, 346, 347, 241, 481, 73, 175, 172, 168, 165,
	538, 539, 87, 507, 18, 92, 71, 257, 176, 542,
	147, 148, 149, 289, 551, 505, 549, 69, 422, 67,
	456, 518, 457, 384, 517, 461, 511, 514, 252, 377,
	586, 563, 18, 557, 560, 559, 562, 558, 561, 566,
	38, 567, 39, 57, 569, 569, 569, 570, 571, 60,
	525, 273, 186, 17, 79, 16, 15, 14, 13, 12,
	40, 41, 42, 43, 258, 45, 389, 260, 48, 81,
	380, 55, 579, 546, 588, 78, 540, 512, 589, 516,
	590, 460, 443, 244, 373, 114, 112, 550, 190, 552,
	116, 449, 108, 471, 282, 398, 280, 107, 406, 193,
	578, 147, 148, 149, 68, 334, 33, 70, 113, 133,
	11, 10, 142, 564, 565, 514, 9, 8, 7, 106,
	127, 128, 129, 6, 135, 332, 132, 5, 139, 122,
	4, 130, 131, 117, 118, 119, 150, 151, 152, 153,
	154, 155, 156, 157, 158, 123, 124, 125, 2, 126,
	1, 0, 0, 0, 0, 0, 0, 0, 120, 121,
	0, 214, 211, 212, 194, 218, 205, 204, 206, 216,
	215, 217, 213, 207, 208, 209, 210, 195, 196, 197,
	198, 199, 200, 203, 201, 202, 0, 0, 144, 143,
	506, 0, 0, 0, 0, 0, 0, 111, 0, 0,
	303, 304, 147, 148, 149, 302, 0, 0, 0, 113,
	133, 0, 0, 142, 0, 0, 137, 138, 104, 145,
	80, 127, 128, 129, 140, 135, 0, 132, 0, 139,
	122, 0, 130, 131, 117, 118, 119, 150, 151, 152,
	153, 154, 155, 156, 157, 158, 123, 124, 125, 0,
	126, 0, 0, 141, 0, 0, 0, 0, 0, 120,
	121, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 547, 548, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 144,
	143, 0, 0, 0, 0, 0, 0, 0, 111, 0,
	0, 0, 0, 147, 148, 149, 0, 0, 0, 0,
	113, 133, 0, 0, 142, 0, 0, 137, 138, 0,
	145, 106, 127, 128, 129, 140, 135, 337, 132, 0,
	139, 122, 0, 130, 131, 117, 118, 119, 150, 151,
	152, 153, 154, 155, 156, 157, 158, 123, 124, 125,
	0, 126, 0, 0, 141, 0, 0, 0, 0, 0,
	120, 121, 214, 211, 212, 194, 218, 205, 204, 206,
	216, 215, 217, 213, 207, 208, 209, 210, 195, 196,
	197, 198, 199, 200, 203, 201, 202, 0, 0, 0,
	144, 143, 0, 0, 0, 0, 0, 0, 0, 111,
	0, 0, 0, 0, 147, 148, 149, 0, 0, 0,
	0, 113, 133, 0, 0, 142, 0, 0, 137, 138,
	104, 145, 106, 127, 128, 129, 140, 135, 0, 132,
	0, 139, 122, 0, 130, 131, 117, 118, 119, 150,
	151, 152, 153, 154, 155, 156, 157, 158, 123, 124,
	125, 433, 126, 0, 0, 141, 0, 0, 0, 0,
	0, 120, 121, 214, 211, 212, 194, 218, 205, 204,
	206, 216, 215, 217, 213, 207, 208, 209, 210, 195,
	196, 197, 198, 199, 200, 203, 201, 202, 0, 544,
	0, 144, 143, 0, 0, 0, 0, 0, 0, 0,
	111, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 18, 0, 0, 137,
	138, 104, 145, 0, 0, 0, 0, 140, 0, 0,
	147, 148, 149, 0, 0, 0, 0, 113, 133, 0,
	0, 142, 0, 0, 0, 0, 0, 0, 80, 127,
	128, 129, 0, 135, 0, 132, 141, 139, 122, 0,
	130, 131, 117, 118, 119, 150, 151, 152, 153, 154,
	155, 156, 157, 158, 123, 124, 125, 543, 126, 0,
	0, 0, 0, 0, 0, 0, 0, 120, 121, 214,
	211, 212, 194, 218, 205, 204, 206, 216, 215, 217,
	213, 207, 208, 209, 210, 195, 196, 197, 198, 199,
	200, 203, 201, 202, 0, 0, 0, 144, 143, 0,
	0, 0, 0, 0, 0, 0, 111, 0, 0, 0,
	0, 147, 148, 149, 0, 0, 0, 0, 113, 133,
	0, 0, 142, 0, 0, 137, 138, 0, 145, 80,
	127, 128, 129, 140, 135, 0, 132, 0, 139, 122,
	0, 130, 131, 117, 118, 119, 150, 151, 152, 153,
	154, 155, 156, 157, 158, 123, 124, 125, 0, 126,
	0, 0, 141, 0, 0, 0, 0, 0, 120, 121,
	0, 0, 0, 0, 220, 150, 151, 152, 153, 154,
	155, 156, 157, 158, 0, 0, 65, 0, 0, 340,
	341, 342, 343, 344, 345, 346, 347, 0, 144, 143,
	434, 0, 0, 0, 0, 0, 0, 111, 356, 357,
	358, 359, 360, 361, 362, 363, 364, 365, 366, 0,
	0, 0, 0, 0, 0, 0, 137, 138, 0, 145,
	0, 0, 0, 0, 140, 214, 211, 212, 194, 218,
	205, 204, 206, 216, 215, 217, 213, 207, 208, 209,
	210, 195, 196, 197, 198, 199, 200, 203, 201, 202,
	0, 0, 0, 141, 214, 211, 212, 194, 218, 205,
	204, 206, 216, 215, 217, 213, 207, 208, 209, 210,
	195, 196, 197, 198, 199, 200, 203, 201, 202, 0,
	214, 211, 212, 194, 218, 205, 204, 206, 216, 215,
	217, 213, 207, 208, 209, 210, 195, 196, 197, 198,
	199, 200, 203, 201, 202, 214, 211, 212, 480, 218,
	205, 204, 206, 216, 215, 217, 213, 207, 208, 209,
	210, 195, 196, 197, 198, 199, 200, 203, 201, 202,
	214, 211, 212, 424, 218, 205, 204, 206, 216, 215,
	217, 213, 207, 208, 209, 210, 195, 196, 197, 198,
	199, 200, 203, 201, 202, 214, 211, 212, 194, 218,
	205, 204, 206, 216, 215, 217, 0, 207, 208, 209,
	210, 195, 196, 197, 198, 199, 200, 203, 201, 202,
	218, 205, 204, 206, 216, 215, 217, 213, 207, 208,
	209, 210, 195, 196, 197, 198, 199, 200, 203, 201,
	202,
}
var yyPact = [...]int{

	314, -1000, -1000, 70, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -144, -136, -115, -129, -1000, -1000, -1000,
	-1000, -92, 416, 537, 507, -1000, -1000, -1000, 504, -1000,
	485, 468, 161, 32, -146, -118, 416, -1000, -126, 416,
	-1000, 475, -153, 416, -153, 484, 109, -98, 153, 416,
	-104, -1000, -1000, -1000, 407, -1000, -1000, -1000, 895, -1000,
	316, 468, 429, -53, 468, 73, 472, -1000, -1, -1000,
	-54, 471, 6, 416, -1000, 470, -1000, -127, 469, 492,
	28, 416, 468, -1000, 1122, 1122, 109, 109, 1122, 153,
	17, 1122, 82, -1000, -1000, 1179, -62, -1000, -1000, -1000,
	-1000, 1122, 1122, 378, -1000, 406, 401, 400, 399, 398,
	397, 396, 379, 395, 394, 393, 392, -1000, -1000, -1000,
	414, 413, 466, -1000, -1000, 1021, -1000, -1000, -1000, -1000,
	1122, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	381, 432, 456, 529, 432, -1000, 1122, 416, -1000, 491,
	-156, -1000, 24, -1000, 455, -1000, -1000, 437, -1000, 380,
	1150, 1150, -1000, -1000, 1150, 109, -7, 1122, 1122, 297,
	1150, 35, 895, 499, 1122, 1122, 1122, 1122, 1122, 1122,
	1122, 1122, 1122, 1122, 693, 693, 693, 693, 693, 693,
	693, 1122, 1122, 377, 13, 1122, 128, 1122, 1122, -1000,
	416, 80, 1150, -1000, -1000, 537, 592, 895, 794, 427,
	427, 1122, 1122, 895, -1000, 1151, 895, 895, 895, -1000,
	-1000, 412, 295, 162, -69, 1150, 509, 432, 432, 214,
	-1000, 521, 1122, -1000, 1150, -1000, -1000, -1000, 22, 416,
	-1000, -128, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	509, 432, -1000, -1000, 1122, 1122, 1150, 1280, -1000, 1122,
	203, 46, 443, 48, -64, -1000, -1000, -1000, -1000, -1000,
	1150, -27, -27, -27, -56, -56, -1000, -1000, -1000, -1000,
	-34, 378, -1000, -1000, -1000, -34, 378, -34, 378, -8,
	378, -8, 378, -8, 378, -8, 378, 245, 245, -1000,
	377, 1122, 1122, 1122, -34, -1000, 501, -1000, -34, 1255,
	-1000, -1000, -1000, 293, 895, 291, 287, -1000, 122, 116,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 110, 107,
	858, 1205, 343, 140, 134, 75, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 104, 98, 286,
	269, -1000, -1000, -72, -1000, 1122, 21, 377, 70, 30,
	258, -1000, 521, 516, 519, 1150, 436, -1000, -1000, 435,
	-1000, -1000, 73, 1150, 1150, 1150, 525, 35, 35, -1000,
	-1000, 55, 54, 63, 62, 57, -82, -1000, 434, 301,
	38, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -34,
	-34, 1230, -1000, -1000, 1122, -1000, 222, -1000, -1000, 895,
	895, 895, 895, 348, 348, -1000, 895, 895, 895, 243,
	242, -1000, -1000, -76, -1000, 1122, 556, -1000, 481, 92,
	-1000, -1000, -1000, 432, 516, -1000, 1122, 1122, -1000, -1000,
	523, 518, 46, 20, -1000, 56, -1000, 41, -1000, -1000,
	-1000, -1000, -119, -122, -124, -1000, -1000, -1000, -1000, -1000,
	1122, 1301, -1000, 210, 204, 168, 155, 317, -1000, -1000,
	219, 217, -1000, -1000, -1000, -1000, -1000, -1000, 375, 315,
	313, 310, 895, 895, -1000, 1150, 1122, 458, 377, -1000,
	-1000, 984, 81, -1000, 757, -1000, 521, 1122, 1122, 1122,
	-1000, -1000, 368, 362, 351, 1301, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 309, -1000, -1000, -1000, 1151, 1151,
	1150, 534, -1000, 1122, 1122, 1122, -1000, -1000, -1000, 516,
	1150, 74, 1150, 416, 416, 416, -1000, 308, 305, 304,
	280, 273, 272, 432, 1150, 1150, -1000, 454, 154, -1000,
	136, 103, -1000, -1000, -1000, -1000, -1000, -1000, 73, -1000,
	533, -10, -1000, 416, -1000, -1000, -1000, 416, -1000, 416,
	-1000,
}
var yyPgo = [...]int{

	0, 660, 658, 55, 640, 637, 633, 628, 627, 626,
	621, 620, 550, 617, 616, 20, 614, 27, 7, 609,
	608, 79, 607, 606, 15, 605, 604, 252, 603, 2,
	23, 29, 602, 10, 19, 5, 601, 600, 4, 6,
	8, 12, 26, 596, 21, 595, 594, 13, 593, 592,
	591, 589, 9, 587, 3, 583, 1, 582, 14, 580,
	11, 46, 52, 342, 579, 578, 577, 576, 575, 574,
	0, 35, 569, 568, 567, 566, 565, 563, 250, 16,
	562, 561, 559, 553, 552,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 75, 75, 5, 6, 7, 7, 72, 73,
	74, 77, 80, 80, 81, 81, 81, 82, 82, 83,
	83, 83, 76, 76, 76, 76, 76, 8, 8, 8,
	9, 9, 9, 10, 11, 11, 11, 84, 12, 13,
	13, 14, 14, 14, 14, 14, 16, 16, 17, 17,
	18, 18, 18, 18, 19, 19, 19, 23, 23, 24,
	24, 24, 24, 20, 20, 20, 25, 25, 25, 25,
	25, 25, 25, 25, 25, 26, 26, 26, 26, 26,
	26, 26, 27, 27, 28, 28, 28, 28, 29, 29,
	30, 30, 79, 79, 79, 78, 78, 15, 15, 15,
	31, 31, 36, 36, 33, 33, 42, 35, 35, 21,
	21, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	39, 39, 39, 39, 39, 39, 39, 39, 38, 38,
	38, 38, 38, 38, 38, 38, 38, 40, 40, 40,
	40, 40, 40, 40, 40, 40, 40, 40, 41, 41,
	41, 41, 41, 41, 41, 41, 41, 41, 41, 41,
	37, 37, 37, 37, 37, 37, 43, 43, 43, 45,
	48, 48, 46, 46, 47, 49, 49, 44, 44, 32,
	32, 32, 32, 32, 32, 32, 32, 32, 34, 34,
	34, 50, 50, 51, 51, 52, 52, 53, 53, 54,
	55, 55, 55, 56, 56, 56, 56, 57, 57, 57,
	58, 58, 59, 59, 60, 60, 61, 61, 62, 63,
	63, 64, 64, 65, 65, 66, 66, 66, 66, 66,
	67, 67, 68, 68, 69, 69, 70, 71,
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
	3, 1, 1, 1, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 4, 3, 4, 3, 4, 3,
	4, 3, 4, 3, 4, 3, 4, 3, 3, 2,
	2, 3, 4, 3, 4, 3, 4, 3, 4, 3,
	4, 2, 5, 6, 1, 3, 4, 5, 4, 1,
	4, 3, 6, 6, 6, 6, 6, 6, 7, 4,
	6, 6, 6, 8, 8, 8, 8, 8, 8, 4,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 2, 1, 2, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 5,
	0, 1, 1, 2, 4, 0, 2, 1, 3, 1,
	1, 1, 2, 2, 2, 4, 1, 1, 1, 1,
	1, 0, 3, 0, 2, 0, 3, 1, 3, 2,
	0, 1, 1, 0, 2, 4, 4, 0, 2, 4,
	0, 3, 1, 3, 0, 5, 1, 3, 3, 0,
	2, 0, 3, 0, 1, 1, 1, 1, 1, 1,
	0, 1, 0, 1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -72, -73, -74, -75, -76, -77, 5, 6,
	7, 8, 35, 162, 163, 165, 164, 147, 148, 149,
	151, 153, 152, -14, 99, 100, 101, 102, -12, -84,
	-12, -12, -12, -12, 166, -68, 168, 172, -65, 168,
	170, 166, 166, 167, 168, -12, 154, -83, 155, 156,
	-82, 159, 160, 158, -70, 37, -3, 22, -16, 23,
	-13, 31, -27, 37, 98, -61, 150, -62, -44, -70,
	37, -64, 171, 167, -70, 166, -70, 37, -63, 171,
	-70, -63, 31, -79, 9, 124, 157, -78, 98, -70,
	161, 42, -17, -18, 136, -21, 37, -22, -32, -44,
	-33, 115, -43, 26, -45, -70, -37, 51, 52, 53,
	76, 77, 47, 63, 64, 65, 67, 38, 39, 40,
	49, 50, 44, 27, -34, 42, -42, 134, 135, 46,
	142, 171, 30, 107, 106, 137, -38, 19, 20, 21,
	54, 55, 56, 57, 58, 59, 60, 61, 62, 41,
	-27, 35, 140, -27, 103, 37, 121, 140, 37, 115,
	-70, -71, 37, -71, 169, 37, 26, 114, -70, -27,
	-21, -21, -79, -79, -21, -78, -80, 98, 126, -35,
	-21, 98, 103, -19, 118, 131, 132, 133, 134, 135,
	136, 138, 139, 137, 121, 120, 122, 127, 128, 129,
	130, 116, 117, 126, 115, 124, 123, 125, 119, -70,
	25, 140, -21, -21, -42, 42, 42, 42, 42, 42,
	42, 42, 42, 42, 38, 42, 42, 42, 42, 38,
	38, 37, -35, -3, -48, -21, -58, 35, 42, -61,
	37, -30, 9, -62, -21, -70, -71, 26, -69, 173,
	-66, 165, 163, 34, 164, 12, 37, 37, 37, -71,
	-58, 35, -79, -81, 98, 126, -21, -21, 43, 103,
	-23, -24, -26, 42, 37, -42, 161, 155, -18, 24,
	-21, -21, -21, -21, -21, -21, -21, -21, -21, -21,
	-21, -15, 22, 17, 18, -21, -15, -21, -15, -21,
	-15, -21, -15, -21, -15, -21, -15, -21, -21, -33,
	126, 124, 125, 119, -21, 27, 115, -34, -21, -21,
	-70, 136, 43, -17, 23, -17, -17, 43, -38, -39,
	68, 69, 70, 71, 72, 73, 74, 75, -38, -39,
	-21, -21, -18, -38, -40, -39, 87, 88, 89, 90,
	91, 92, 93, 94, 95, 96, 97, -18, -18, -17,
	38, 43, 43, -46, -47, 143, -31, 30, -3, -61,
	-59, -44, -30, -52, 12, -21, 114, -70, -71, -67,
	169, -31, -61, -21, -21, -21, -30, 103, -25, 104,
	105, 106, 107, 108, 110, 111, -20, 37, 25, -24,
	140, -42, -42, -42, -42, -42, -42, -42, -33, -21,
	-21, -21, 27, -34, 118, 43, -17, 43, 43, 103,
	103, 103, 103, 103, 25, 43, 98, 98, 98, 103,
	103, 43, 45, -49, -47, 145, -21, -60, 114, -36,
	-33, -60, 43, 103, -52, -56, 14, 13, 37, 37,
	-50, 10, -24, -24, 104, 109, 104, 109, 104, 104,
	104, -28, 112, 170, 113, 37, 43, 37, 161, 155,
	118, -21, 43, -17, -17, -17, -17, -41, 78, 47,
	79, 80, 81, 82, 83, 84, 85, 37, -41, -18,
	-18, -18, 66, 66, 146, -21, 144, 32, 103, -44,
	-56, -21, -53, -54, -21, -71, -51, 11, 13, 114,
	104, 104, 167, 167, 167, -21, 43, 43, 43, 43,
	43, 86, 86, 43, 24, 43, 43, 43, -18, -18,
	-21, 33, -33, 103, 15, 103, -55, 28, 29, -52,
	-21, -35, -21, 42, 42, 42, 43, -38, -40, -39,
	-38, -40, -39, 7, -21, -21, -54, -56, -29, -70,
	-29, -29, 43, 43, 43, 43, 43, 43, -61, -57,
	16, 36, 43, 103, 43, 43, 7, 126, -70, -70,
	-70,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 57, 57,
	57, 57, 57, 312, 303, 0, 0, 28, 29, 30,
	57, -2, 0, 0, 61, 63, 64, 65, 66, 59,
	0, 0, 0, 0, 301, 0, 0, 313, 0, 0,
	304, 0, 299, 0, 299, 0, 112, 0, 115, 0,
	0, 40, 41, 38, 0, 316, 19, 62, 0, 67,
	58, 0, 0, 102, 0, 26, 0, 296, 0, 257,
	316, 0, 0, 0, 317, 0, 317, 0, 0, 0,
	0, 0, 0, 42, 0, 0, 112, 112, 0, 115,
	0, 0, 17, 68, 70, 74, 316, 129, 131, 132,
	133, 0, 0, 0, 174, 257, 0, 179, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 259, 260, 261,
	0, 0, 0, 266, 267, 0, 125, 246, 247, 248,
	250, 240, 241, 242, 243, 244, 245, 268, 269, 270,
	208, 209, 210, 211, 212, 213, 214, 215, 216, 60,
	290, 0, 0, 110, 0, 27, 0, 0, 317, 0,
	314, 49, 0, 52, 0, 54, 300, 0, 317, 290,
	113, 114, 43, 44, 116, 112, 34, 0, 0, 0,
	127, 0, 0, 71, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 75,
	0, 0, 159, 160, 171, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 262, 0, 0, 0, 0, 263,
	264, 0, 0, 0, 0, 251, 0, 0, 0, 110,
	103, 275, 0, 297, 298, 258, 47, 302, 0, 0,
	317, 310, 305, 306, 307, 308, 309, 53, 55, 56,
	0, 0, 45, 46, 0, 0, 32, 33, 31, 0,
	110, 77, 83, 0, 95, 97, 98, 99, 69, 72,
	130, 134, 135, 136, 137, 138, 139, 140, 141, 142,
	143, 0, 117, 118, 119, 145, 0, 147, 0, 149,
	0, 151, 0, 153, 0, 155, 0, 157, 158, 161,
	0, 0, 0, 0, 163, 165, 0, 167, 169, 0,
	76, 73, 175, 0, 0, 0, 0, 181, 0, 0,
	200, 201, 202, 203, 204, 205, 206, 207, 0, 0,
	0, 0, 0, 0, 0, 0, 217, 218, 219, 220,
	221, 222, 223, 224, 225, 226, 227, 0, 0, 0,
	0, 124, 126, 255, 252, 0, 294, 0, 121, 294,
	0, 292, 275, 283, 0, 111, 0, 315, 50, 0,
	311, 22, 23, 35, 36, 128, 271, 0, 0, 86,
	87, 0, 0, 0, 0, 0, 104, 84, 0, 0,
	0, 144, 146, 148, 150, 152, 154, 156, 162, 164,
	170, 0, 166, 168, 0, 176, 0, 178, 180, 0,
	0, 0, 0, 0, 0, 189, 0, 0, 0, 0,
	0, 199, 265, 0, 253, 0, 0, 20, 0, 120,
	122, 21, 291, 0, 283, 25, 0, 0, 317, 51,
	273, 0, 78, 81, 88, 0, 90, 0, 92, 93,
	94, 79, 0, 0, 0, 85, 80, 96, 100, 101,
	0, 172, 177, 0, 0, 0, 0, 0, 228, 229,
	230, 232, 234, 235, 236, 237, 238, 239, 0, 0,
	0, 0, 0, 0, 249, 256, 0, 0, 0, 293,
	24, 284, 276, 277, 280, 48, 275, 0, 0, 0,
	89, 91, 0, 0, 0, 173, 182, 183, 184, 185,
	186, 231, 233, 187, 0, 190, 191, 192, 0, 0,
	254, 0, 123, 0, 0, 0, 279, 281, 282, 283,
	274, 272, 82, 0, 0, 0, 188, 0, 0, 0,
	0, 0, 0, 0, 285, 286, 278, 287, 0, 108,
	0, 0, 193, 194, 195, 196, 197, 198, 295, 18,
	0, 0, 105, 0, 106, 107, 288, 0, 109, 0,
	289,
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
	152, 153, 154, 155, 156, 157, 158, 159, 160, 161,
	162, 163, 164, 165, 166, 167, 168, 169, 170, 171,
	172, 173,
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
		//line sql.y:213
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:219
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:239
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 18:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:243
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].exprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:247
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 20:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:254
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:258
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
		//line sql.y:270
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:274
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
		//line sql.y:287
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 25:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:293
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:299
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:303
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:309
		{
			yyVAL.statement = &Begin{}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:315
		{
			yyVAL.statement = &Commit{}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:321
		{
			yyVAL.statement = &Rollback{}
		}
	case 31:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:327
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].exprs}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:333
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:337
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 34:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:342
		{
			yyVAL.expr = nil
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:346
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:350
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 37:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:355
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:359
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 39:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:364
		{
			yyVAL.str = AST_SHOW_SESSION_VARIABLE
		}
	case 40:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:368
		{
			yyVAL.str = AST_SHOW_SESSION_VARIABLE
		}
	case 41:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:372
		{
			yyVAL.str = AST_SHOW_GLOBAL_VARIABLE
		}
	case 42:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:379
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 43:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:383
		{
			yyVAL.statement = &Show{Section: "variables", Modifier: yyDollar[2].str, LikeOrWhere: yyDollar[4].expr}
		}
	case 44:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:387
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].expr, LikeOrWhere: yyDollar[4].expr}
		}
	case 45:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:391
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].expr, LikeOrWhere: yyDollar[5].expr}
		}
	case 46:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:395
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].expr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].expr}
		}
	case 47:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:401
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 48:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:405
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 49:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:410
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 50:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:416
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 51:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:420
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 52:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:425
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 53:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:431
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 54:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:437
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 55:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:441
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 56:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:446
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 57:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:451
		{
			SetAllowComments(yylex, true)
		}
	case 58:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:455
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 59:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:461
		{
			yyVAL.bytes2 = nil
		}
	case 60:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:465
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 61:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:471
		{
			yyVAL.str = AST_UNION
		}
	case 62:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:475
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 63:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:479
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 64:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:483
		{
			yyVAL.str = AST_EXCEPT
		}
	case 65:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:487
		{
			yyVAL.str = AST_INTERSECT
		}
	case 66:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:492
		{
			yyVAL.str = ""
		}
	case 67:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:496
		{
			yyVAL.str = AST_DISTINCT
		}
	case 68:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:502
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:506
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 70:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:512
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 71:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:516
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 72:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:520
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 73:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:524
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 74:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:529
		{
			yyVAL.bytes = nil
		}
	case 75:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:533
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 76:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:537
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 77:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:543
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:547
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:553
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:557
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:561
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 82:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:565
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].expr}
		}
	case 83:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:570
		{
			yyVAL.bytes = nil
		}
	case 84:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:574
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 85:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:578
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:584
		{
			yyVAL.str = AST_JOIN
		}
	case 87:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:588
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 88:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:592
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:596
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 90:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:600
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 91:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:604
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 92:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:608
		{
			yyVAL.str = AST_JOIN
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:612
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 94:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:616
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 95:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:622
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:626
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:630
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:634
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 99:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:638
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 100:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:642
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:646
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 102:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:652
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 103:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:656
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 104:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:661
		{
			yyVAL.indexHints = nil
		}
	case 105:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:665
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 106:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:669
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 107:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:673
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 108:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:679
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 109:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:683
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 110:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:688
		{
			yyVAL.expr = nil
		}
	case 111:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:692
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 112:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:697
		{
			yyVAL.expr = nil
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:701
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 114:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:705
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 115:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:710
		{
			yyVAL.expr = nil
		}
	case 116:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:714
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 117:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:720
		{
			yyVAL.str = AST_ALL
		}
	case 118:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:724
		{
			yyVAL.str = AST_SOME
		}
	case 119:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:728
		{
			yyVAL.str = AST_ANY
		}
	case 120:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:734
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 121:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:738
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 122:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:744
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 123:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:748
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 124:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:754
		{
			yyVAL.tuple = ValTuple(yyDollar[2].exprs)
		}
	case 125:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:758
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 126:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:764
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 127:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:770
		{
			yyVAL.exprs = Exprs{yyDollar[1].expr}
		}
	case 128:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:774
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 129:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:780
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 130:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:784
		{
			yyVAL.expr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 131:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:790
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 132:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:794
		{
			yyVAL.expr = yyDollar[1].colName
		}
	case 133:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:798
		{
			yyVAL.expr = yyDollar[1].tuple
		}
	case 134:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:802
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITAND, Right: yyDollar[3].expr}
		}
	case 135:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:806
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITOR, Right: yyDollar[3].expr}
		}
	case 136:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:810
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITXOR, Right: yyDollar[3].expr}
		}
	case 137:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:814
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_PLUS, Right: yyDollar[3].expr}
		}
	case 138:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:818
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MINUS, Right: yyDollar[3].expr}
		}
	case 139:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:822
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MULT, Right: yyDollar[3].expr}
		}
	case 140:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:826
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_DIV, Right: yyDollar[3].expr}
		}
	case 141:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:830
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_IDIV, Right: yyDollar[3].expr}
		}
	case 142:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:834
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MOD, Right: yyDollar[3].expr}
		}
	case 143:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:838
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, Right: yyDollar[3].expr}
		}
	case 144:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:842
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:846
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, Right: yyDollar[3].expr}
		}
	case 146:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:850
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:854
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, Right: yyDollar[3].expr}
		}
	case 148:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:858
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 149:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:862
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, Right: yyDollar[3].expr}
		}
	case 150:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:866
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 151:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:870
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, Right: yyDollar[3].expr}
		}
	case 152:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:874
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:878
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, Right: yyDollar[3].expr}
		}
	case 154:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:882
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:886
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, Right: yyDollar[3].expr}
		}
	case 156:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:890
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:894
		{
			yyVAL.expr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:898
		{
			yyVAL.expr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 159:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:902
		{
			yyVAL.expr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 160:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:906
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
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:921
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 162:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:925
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 163:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:929
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LIKE, Right: yyDollar[3].expr}
		}
	case 164:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:933
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_LIKE, Right: yyDollar[4].expr}
		}
	case 165:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:937
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 166:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:941
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 167:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:945
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: yyDollar[3].expr}
		}
	case 168:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:949
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: yyDollar[4].expr}
		}
	case 169:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:953
		{
			yyVAL.expr = &RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[3].expr}
		}
	case 170:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:957
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[4].expr}}
		}
	case 171:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:961
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 172:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:965
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_BETWEEN, From: yyDollar[3].expr, To: yyDollar[5].expr}
		}
	case 173:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:969
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].expr, To: yyDollar[6].expr}
		}
	case 174:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:973
		{
			yyVAL.expr = yyDollar[1].caseExpr
		}
	case 175:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:977
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 176:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:981
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 177:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:985
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 178:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:989
		{
			yyVAL.expr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 179:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:993
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 180:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:997
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 181:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1001
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 182:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1005
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 183:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1009
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 184:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1013
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 185:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1017
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 186:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1021
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 187:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1025
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 188:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:1029
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 189:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1033
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date"), Exprs: SelectExprs{yyDollar[3].selectExpr}}
		}
	case 190:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1037
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 191:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1041
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 192:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1045
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 193:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1049
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 194:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1053
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 195:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1057
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 196:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1061
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 197:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1065
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 198:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1069
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 199:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1073
		{
			yyVAL.expr = &FuncExpr{Name: []byte("str_to_date"), Exprs: yyDollar[3].selectExprs}
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1079
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1083
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1087
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1091
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1095
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 205:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1099
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1103
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1107
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1113
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 209:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1117
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1121
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1125
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1129
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1133
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1137
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1141
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1145
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1151
		{
			yyVAL.bytes = SECOND_MICROSECOND_BYTES
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1155
		{
			yyVAL.bytes = MINUTE_MICROSECOND_BYTES
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1159
		{
			yyVAL.bytes = MINUTE_SECOND_BYTES
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1163
		{
			yyVAL.bytes = HOUR_MICROSECOND_BYTES
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1167
		{
			yyVAL.bytes = HOUR_SECOND_BYTES
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1171
		{
			yyVAL.bytes = HOUR_MINUTE_BYTES
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1175
		{
			yyVAL.bytes = DAY_MICROSECOND_BYTES
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1179
		{
			yyVAL.bytes = DAY_SECOND_BYTES
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1183
		{
			yyVAL.bytes = DAY_MINUTE_BYTES
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1187
		{
			yyVAL.bytes = DAY_HOUR_BYTES
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1191
		{
			yyVAL.bytes = YEAR_MONTH_BYTES
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1197
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1201
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1205
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 231:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1209
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1213
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 233:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1217
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1221
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1225
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1229
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1233
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1237
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1244
		{
			if bytes.Equal(bytes.ToLower(yyDollar[1].bytes), []byte("datetime")) {
				yyVAL.bytes = DATETIME_BYTES
			} else {
				yylex.Error("expecting datetime")
				return 1
			}
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1255
		{
			yyVAL.bytes = IF_BYTES
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1259
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1263
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1267
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1271
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1275
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 246:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1281
		{
			yyVAL.byt = AST_UPLUS
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1285
		{
			yyVAL.byt = AST_UMINUS
		}
	case 248:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1289
		{
			yyVAL.byt = AST_TILDA
		}
	case 249:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1295
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].expr, Whens: yyDollar[3].whens, Else: yyDollar[4].expr}
		}
	case 250:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1300
		{
			yyVAL.expr = nil
		}
	case 251:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1304
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1310
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 253:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1314
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 254:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1320
		{
			yyVAL.when = &When{Cond: yyDollar[2].expr, Val: yyDollar[4].expr}
		}
	case 255:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1325
		{
			yyVAL.expr = nil
		}
	case 256:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1329
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1335
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 258:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1339
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1345
		{
			yyVAL.expr = StrVal(yyDollar[1].bytes)
		}
	case 260:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1349
		{
			yyVAL.expr = NumVal(yyDollar[1].bytes)
		}
	case 261:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1353
		{
			yyVAL.expr = ValArg(yyDollar[1].bytes)
		}
	case 262:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1357
		{
			yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 263:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1361
		{
			yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 264:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1365
		{
			yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 265:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1369
		{
			if bytes.Equal(bytes.ToLower(yyDollar[2].bytes), []byte("d")) {
				yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[3].bytes}
			} else if bytes.Equal(bytes.ToLower(yyDollar[2].bytes), []byte("t")) {
				yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[3].bytes}
			} else if bytes.Equal(bytes.ToLower(yyDollar[2].bytes), []byte("ts")) {
				yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[3].bytes}
			} else {
				yylex.Error("expecting d, t, or ts")
				return 1
			}
		}
	case 266:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1382
		{
			yyVAL.expr = &NullVal{}
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1386
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 268:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1392
		{
			yyVAL.expr = &TrueVal{}
		}
	case 269:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1396
		{
			yyVAL.expr = &FalseVal{}
		}
	case 270:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1400
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 271:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1405
		{
			yyVAL.exprs = nil
		}
	case 272:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1409
		{
			yyVAL.exprs = yyDollar[3].exprs
		}
	case 273:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1414
		{
			yyVAL.expr = nil
		}
	case 274:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1418
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 275:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1423
		{
			yyVAL.orderBy = nil
		}
	case 276:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1427
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 277:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1433
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 278:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1437
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 279:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1443
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 280:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1448
		{
			yyVAL.str = AST_ASC
		}
	case 281:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1452
		{
			yyVAL.str = AST_ASC
		}
	case 282:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1456
		{
			yyVAL.str = AST_DESC
		}
	case 283:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1461
		{
			yyVAL.limit = nil
		}
	case 284:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1465
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].expr}
		}
	case 285:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1469
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].expr, Rowcount: yyDollar[4].expr}
		}
	case 286:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1473
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].expr, Rowcount: yyDollar[2].expr}
		}
	case 287:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1478
		{
			yyVAL.str = ""
		}
	case 288:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1482
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 289:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1486
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
	case 290:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1499
		{
			yyVAL.columns = nil
		}
	case 291:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1503
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 292:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1509
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 293:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1513
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 294:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1518
		{
			yyVAL.updateExprs = nil
		}
	case 295:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1522
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 296:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1528
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 297:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1532
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 298:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1538
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].expr}
		}
	case 299:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1543
		{
			yyVAL.empty = struct{}{}
		}
	case 300:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1545
		{
			yyVAL.empty = struct{}{}
		}
	case 301:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1548
		{
			yyVAL.empty = struct{}{}
		}
	case 302:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1550
		{
			yyVAL.empty = struct{}{}
		}
	case 303:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1553
		{
			yyVAL.empty = struct{}{}
		}
	case 304:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1555
		{
			yyVAL.empty = struct{}{}
		}
	case 305:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1559
		{
			yyVAL.empty = struct{}{}
		}
	case 306:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1561
		{
			yyVAL.empty = struct{}{}
		}
	case 307:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1563
		{
			yyVAL.empty = struct{}{}
		}
	case 308:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1565
		{
			yyVAL.empty = struct{}{}
		}
	case 309:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1567
		{
			yyVAL.empty = struct{}{}
		}
	case 310:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1570
		{
			yyVAL.empty = struct{}{}
		}
	case 311:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1572
		{
			yyVAL.empty = struct{}{}
		}
	case 312:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1575
		{
			yyVAL.empty = struct{}{}
		}
	case 313:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1577
		{
			yyVAL.empty = struct{}{}
		}
	case 314:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1580
		{
			yyVAL.empty = struct{}{}
		}
	case 315:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1582
		{
			yyVAL.empty = struct{}{}
		}
	case 316:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1586
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 317:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1591
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
