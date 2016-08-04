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
const TILDE = 57386
const DATE = 57387
const DATETIME = 57388
const TIME = 57389
const TIMESTAMP = 57390
const CURRENT_TIMESTAMP = 57391
const TIMESTAMPADD = 57392
const TIMESTAMPDIFF = 57393
const YEAR = 57394
const QUARTER = 57395
const MONTH = 57396
const WEEK = 57397
const DAY = 57398
const HOUR = 57399
const MINUTE = 57400
const SECOND = 57401
const MICROSECOND = 57402
const EXTRACT = 57403
const DATE_ADD = 57404
const DATE_SUB = 57405
const INTERVAL = 57406
const STR_TO_DATE = 57407
const SQL_TSI_YEAR = 57408
const SQL_TSI_QUARTER = 57409
const SQL_TSI_MONTH = 57410
const SQL_TSI_WEEK = 57411
const SQL_TSI_DAY = 57412
const SQL_TSI_HOUR = 57413
const SQL_TSI_MINUTE = 57414
const SQL_TSI_SECOND = 57415
const CONVERT = 57416
const CAST = 57417
const CHAR = 57418
const SIGNED = 57419
const UNSIGNED = 57420
const SQL_BIGINT = 57421
const SQL_VARCHAR = 57422
const SQL_DATE = 57423
const SQL_TIMESTAMP = 57424
const SQL_DOUBLE = 57425
const INTEGER = 57426
const SECOND_MICROSECOND = 57427
const MINUTE_MICROSECOND = 57428
const MINUTE_SECOND = 57429
const HOUR_MICROSECOND = 57430
const HOUR_SECOND = 57431
const HOUR_MINUTE = 57432
const DAY_MICROSECOND = 57433
const DAY_SECOND = 57434
const DAY_MINUTE = 57435
const DAY_HOUR = 57436
const YEAR_MONTH = 57437
const FROM = 57438
const UNION = 57439
const MINUS = 57440
const EXCEPT = 57441
const INTERSECT = 57442
const COMMA = 57443
const JOIN = 57444
const STRAIGHT_JOIN = 57445
const LEFT = 57446
const RIGHT = 57447
const INNER = 57448
const OUTER = 57449
const CROSS = 57450
const NATURAL = 57451
const USE = 57452
const FORCE = 57453
const ON = 57454
const NOT = 57455
const OR = 57456
const XOR = 57457
const AND = 57458
const BETWEEN = 57459
const NE = 57460
const EQ = 57461
const NULL_SAFE_EQUAL = 57462
const IS = 57463
const LIKE = 57464
const REGEXP = 57465
const IN = 57466
const LT = 57467
const GT = 57468
const LE = 57469
const GE = 57470
const BIT_AND = 57471
const BIT_OR = 57472
const CARET = 57473
const PLUS = 57474
const SUB = 57475
const TIMES = 57476
const MOD = 57477
const DIV = 57478
const IDIV = 57479
const DOT = 57480
const UNARY = 57481
const CASE = 57482
const WHEN = 57483
const THEN = 57484
const ELSE = 57485
const END = 57486
const BEGIN = 57487
const COMMIT = 57488
const ROLLBACK = 57489
const NAMES = 57490
const REPLACE = 57491
const ADMIN = 57492
const SHOW = 57493
const DATABASES = 57494
const TABLES = 57495
const PROXY = 57496
const VARIABLES = 57497
const FULL = 57498
const SESSION = 57499
const GLOBAL = 57500
const COLUMNS = 57501
const CREATE = 57502
const ALTER = 57503
const DROP = 57504
const RENAME = 57505
const TABLE = 57506
const INDEX = 57507
const VIEW = 57508
const TO = 57509
const IGNORE = 57510
const IF = 57511
const UNIQUE = 57512
const USING = 57513

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
	159, 37,
	-2, 39,
}

const yyNprod = 318
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1470

var yyAct = [...]int{

	115, 453, 566, 511, 146, 382, 339, 189, 354, 110,
	485, 445, 373, 103, 281, 93, 134, 246, 301, 259,
	251, 89, 375, 75, 82, 136, 102, 49, 44, 50,
	46, 171, 389, 64, 47, 470, 472, 265, 52, 53,
	54, 522, 174, 521, 79, 520, 83, 84, 85, 51,
	86, 109, 100, 77, 90, 243, 3, 96, 80, 263,
	99, 502, 266, 56, 58, 59, 409, 63, 61, 62,
	374, 374, 443, 475, 221, 167, 284, 162, 105, 65,
	274, 283, 18, 187, 170, 200, 203, 201, 202, 66,
	585, 94, 178, 471, 166, 78, 195, 196, 197, 198,
	199, 200, 203, 201, 202, 169, 219, 519, 275, 517,
	323, 188, 182, 183, 284, 321, 322, 320, 173, 283,
	207, 208, 209, 210, 195, 196, 197, 198, 199, 200,
	203, 201, 202, 198, 199, 200, 203, 201, 202, 224,
	446, 385, 177, 242, 147, 148, 149, 398, 399, 400,
	401, 402, 325, 403, 404, 164, 164, 583, 464, 462,
	518, 468, 79, 465, 463, 79, 446, 467, 255, 76,
	466, 279, 252, 180, 181, 543, 331, 184, 530, 506,
	190, 34, 35, 36, 37, 249, 262, 264, 261, 477,
	222, 223, 287, 439, 582, 476, 438, 270, 286, 431,
	256, 272, 371, 580, 95, 527, 288, 191, 430, 429,
	269, 526, 192, 78, 190, 581, 78, 285, 253, 245,
	252, 330, 255, 319, 306, 308, 310, 312, 314, 316,
	287, 428, 437, 327, 338, 348, 286, 349, 326, 436,
	353, 97, 355, 435, 98, 254, 525, 352, 79, 79,
	367, 368, 581, 333, 335, 336, 34, 35, 36, 37,
	386, 581, 74, 192, 396, 369, 276, 277, 529, 192,
	381, 378, 79, 290, 291, 292, 293, 294, 295, 296,
	297, 298, 299, 300, 305, 307, 309, 311, 313, 315,
	317, 318, 387, 390, 324, 391, 328, 329, 408, 78,
	380, 395, 377, 501, 192, 500, 532, 575, 574, 285,
	350, 351, 164, 573, 572, 571, 570, 554, 535, 18,
	19, 20, 21, 78, 534, 531, 377, 410, 533, 88,
	417, 384, 411, 474, 412, 528, 413, 434, 414, 524,
	415, 185, 416, 422, 72, 480, 450, 440, 427, 22,
	426, 424, 370, 392, 393, 278, 234, 553, 394, 552,
	233, 425, 194, 218, 205, 204, 206, 216, 215, 217,
	213, 207, 208, 209, 210, 195, 196, 197, 198, 199,
	200, 203, 201, 202, 91, 442, 448, 452, 551, 135,
	449, 225, 398, 399, 400, 401, 402, 192, 403, 404,
	418, 419, 420, 192, 451, 192, 192, 271, 192, 192,
	279, 460, 461, 279, 248, 238, 160, 247, 237, 163,
	236, 235, 285, 285, 248, 232, 231, 230, 229, 228,
	227, 226, 241, 101, 159, 487, 488, 179, 240, 239,
	65, 80, 473, 457, 496, 456, 268, 267, 250, 497,
	498, 499, 79, 444, 508, 481, 482, 483, 484, 27,
	28, 29, 73, 30, 32, 31, 486, 489, 490, 491,
	492, 493, 494, 495, 23, 24, 26, 25, 150, 151,
	152, 153, 154, 155, 156, 157, 158, 578, 513, 407,
	175, 172, 340, 341, 342, 343, 344, 345, 346, 347,
	168, 406, 479, 507, 165, 87, 38, 579, 161, 539,
	505, 18, 92, 71, 536, 537, 540, 147, 148, 149,
	547, 257, 503, 176, 549, 421, 40, 41, 42, 43,
	67, 289, 69, 509, 512, 454, 376, 55, 516, 455,
	383, 555, 558, 557, 560, 556, 559, 564, 515, 565,
	459, 252, 567, 567, 567, 568, 569, 523, 584, 18,
	561, 39, 79, 57, 60, 273, 186, 17, 16, 15,
	14, 13, 12, 258, 45, 388, 260, 48, 81, 379,
	577, 544, 586, 538, 510, 576, 587, 514, 588, 458,
	441, 244, 372, 114, 548, 190, 550, 112, 116, 447,
	108, 469, 282, 397, 280, 107, 405, 147, 148, 149,
	193, 334, 68, 78, 113, 133, 33, 70, 142, 11,
	562, 563, 512, 10, 9, 106, 127, 128, 129, 8,
	135, 332, 139, 122, 132, 130, 131, 117, 118, 119,
	150, 151, 152, 153, 154, 155, 156, 157, 158, 123,
	124, 125, 7, 126, 6, 5, 4, 2, 1, 0,
	0, 0, 120, 121, 0, 214, 211, 212, 194, 218,
	205, 204, 206, 216, 215, 217, 213, 207, 208, 209,
	210, 195, 196, 197, 198, 199, 200, 203, 201, 202,
	0, 0, 144, 143, 504, 0, 0, 0, 0, 0,
	0, 111, 0, 0, 303, 304, 147, 148, 149, 302,
	0, 0, 0, 113, 133, 0, 0, 142, 0, 0,
	137, 138, 104, 145, 80, 127, 128, 129, 140, 135,
	0, 139, 122, 132, 130, 131, 117, 118, 119, 150,
	151, 152, 153, 154, 155, 156, 157, 158, 123, 124,
	125, 0, 126, 0, 0, 0, 0, 141, 0, 0,
	0, 120, 121, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 545, 546, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 144, 143, 0, 0, 0, 0, 0, 0, 0,
	111, 0, 0, 0, 0, 147, 148, 149, 0, 0,
	0, 0, 113, 133, 0, 0, 142, 0, 0, 137,
	138, 0, 145, 106, 127, 128, 129, 140, 135, 337,
	139, 122, 132, 130, 131, 117, 118, 119, 150, 151,
	152, 153, 154, 155, 156, 157, 158, 123, 124, 125,
	0, 126, 0, 0, 0, 0, 141, 0, 0, 0,
	120, 121, 214, 211, 212, 194, 218, 205, 204, 206,
	216, 215, 217, 213, 207, 208, 209, 210, 195, 196,
	197, 198, 199, 200, 203, 201, 202, 0, 0, 0,
	144, 143, 0, 0, 0, 0, 0, 0, 0, 111,
	0, 0, 0, 0, 147, 148, 149, 0, 0, 0,
	0, 113, 133, 0, 0, 142, 0, 0, 137, 138,
	104, 145, 106, 127, 128, 129, 140, 135, 0, 139,
	122, 132, 130, 131, 117, 118, 119, 150, 151, 152,
	153, 154, 155, 156, 157, 158, 123, 124, 125, 432,
	126, 0, 0, 0, 0, 141, 0, 0, 0, 120,
	121, 214, 211, 212, 194, 218, 205, 204, 206, 216,
	215, 217, 213, 207, 208, 209, 210, 195, 196, 197,
	198, 199, 200, 203, 201, 202, 0, 0, 0, 144,
	143, 542, 0, 0, 0, 0, 0, 0, 111, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 18, 137, 138, 104,
	145, 0, 0, 0, 0, 140, 0, 0, 0, 0,
	147, 148, 149, 0, 0, 0, 0, 113, 133, 0,
	0, 142, 0, 0, 0, 0, 0, 0, 80, 127,
	128, 129, 0, 135, 141, 139, 122, 132, 130, 131,
	117, 118, 119, 150, 151, 152, 153, 154, 155, 156,
	157, 158, 123, 124, 125, 0, 126, 541, 0, 0,
	0, 0, 0, 0, 0, 120, 121, 0, 0, 214,
	211, 212, 194, 218, 205, 204, 206, 216, 215, 217,
	213, 207, 208, 209, 210, 195, 196, 197, 198, 199,
	200, 203, 201, 202, 0, 144, 143, 0, 0, 0,
	0, 0, 0, 0, 111, 0, 0, 0, 0, 147,
	148, 149, 0, 0, 0, 0, 113, 133, 0, 0,
	142, 0, 0, 137, 138, 0, 145, 80, 127, 128,
	129, 140, 135, 0, 139, 122, 132, 130, 131, 117,
	118, 119, 150, 151, 152, 153, 154, 155, 156, 157,
	158, 123, 124, 125, 0, 126, 0, 0, 0, 0,
	141, 0, 0, 0, 120, 121, 0, 0, 0, 0,
	0, 0, 220, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 65, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 144, 143, 0, 433, 0, 0,
	0, 0, 0, 111, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 137, 138, 0, 145, 0, 0, 0, 0,
	140, 214, 211, 212, 194, 218, 205, 204, 206, 216,
	215, 217, 213, 207, 208, 209, 210, 195, 196, 197,
	198, 199, 200, 203, 201, 202, 0, 0, 0, 141,
	214, 211, 212, 194, 218, 205, 204, 206, 216, 215,
	217, 213, 207, 208, 209, 210, 195, 196, 197, 198,
	199, 200, 203, 201, 202, 214, 211, 212, 194, 218,
	205, 204, 206, 216, 215, 217, 213, 207, 208, 209,
	210, 195, 196, 197, 198, 199, 200, 203, 201, 202,
	214, 211, 212, 478, 218, 205, 204, 206, 216, 215,
	217, 213, 207, 208, 209, 210, 195, 196, 197, 198,
	199, 200, 203, 201, 202, 214, 211, 212, 423, 218,
	205, 204, 206, 216, 215, 217, 213, 207, 208, 209,
	210, 195, 196, 197, 198, 199, 200, 203, 201, 202,
	214, 211, 212, 194, 218, 205, 204, 206, 216, 215,
	217, 0, 207, 208, 209, 210, 195, 196, 197, 198,
	199, 200, 203, 201, 202, 218, 205, 204, 206, 216,
	215, 217, 213, 207, 208, 209, 210, 195, 196, 197,
	198, 199, 200, 203, 201, 202, 150, 151, 152, 153,
	154, 155, 156, 157, 158, 0, 0, 0, 0, 0,
	340, 341, 342, 343, 344, 345, 346, 347, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 356,
	357, 358, 359, 360, 361, 362, 363, 364, 365, 366,
}
var yyPact = [...]int{

	314, -1000, -1000, 84, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -136, -139, -115, -126, -1000, -1000, -1000,
	-1000, -89, 403, 554, 508, -1000, -1000, -1000, 509, -1000,
	482, 425, 166, 21, -145, -119, 403, -1000, -116, 403,
	-1000, 468, -148, 403, -148, 481, 82, -98, 148, 403,
	-107, -1000, -1000, -1000, 391, -1000, -1000, -1000, 885, -1000,
	393, 425, 473, -61, 425, 55, 467, -1000, -25, -1000,
	-63, 463, -8, 403, -1000, 454, -1000, -125, 453, 497,
	30, 403, 425, -1000, 1110, 1110, 82, 82, 1110, 148,
	-13, 1110, 111, -1000, -1000, 1167, -64, -1000, -1000, -1000,
	-1000, 1110, 1110, 349, -1000, 389, 388, 387, 386, 385,
	384, 383, 318, 379, 378, 376, 373, -1000, -1000, -1000,
	401, 400, 394, -1000, -1000, 1011, -1000, -1000, -1000, -1000,
	1110, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	382, 404, 411, 542, 404, -1000, 1110, 403, -1000, 495,
	-152, -1000, 25, -1000, 410, -1000, -1000, 409, -1000, 372,
	1138, 1138, -1000, -1000, 1138, 82, -16, 1110, 1110, 312,
	1138, 39, 885, 507, 1110, 1110, 1110, 1110, 1110, 1110,
	1110, 1110, 1110, 1110, 687, 687, 687, 687, 687, 687,
	687, 1110, 1110, 347, -7, 1110, 125, 1110, 1110, -1000,
	403, 42, 1138, -1000, -1000, 554, 588, 885, 786, 426,
	426, 1110, 1110, 885, -1000, 1374, 885, 885, 885, -1000,
	-1000, -1000, 309, 159, -70, 1138, 506, 404, 404, 211,
	-1000, 528, 1110, -1000, 1138, -1000, -1000, -1000, 29, 403,
	-1000, -135, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	506, 404, -1000, -1000, 1110, 1110, 1138, 1267, -1000, 1110,
	163, 45, 464, 77, -72, -1000, -1000, -1000, -1000, -1000,
	1138, 1, 1, 1, -49, -49, -1000, -1000, -1000, -1000,
	-5, 349, -1000, -1000, -1000, -5, 349, -5, 349, -33,
	349, -33, 349, -33, 349, -33, 349, 246, 246, -1000,
	347, 1110, 1110, 1110, -5, -1000, 498, -1000, -5, 1242,
	-1000, -1000, -1000, 308, 885, 307, 305, -1000, 130, 108,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 107, 98,
	848, 1192, 294, 147, 143, 136, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 95, 92, 304,
	-1000, -1000, -71, -1000, 1110, 28, 347, 84, 54, 303,
	-1000, 528, 521, 526, 1138, 408, -1000, -1000, 406, -1000,
	-1000, 55, 1138, 1138, 1138, 540, 39, 39, -1000, -1000,
	57, 56, 68, 65, 59, -75, -1000, 405, 290, 36,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -5, -5,
	1217, -1000, -1000, 1110, -1000, 302, -1000, -1000, 885, 885,
	885, 885, 390, 390, -1000, 885, 885, 885, 241, 239,
	-1000, -83, -1000, 1110, 552, -1000, 478, 78, -1000, -1000,
	-1000, 404, 521, -1000, 1110, 1110, -1000, -1000, 537, 525,
	45, -3, -1000, 58, -1000, 5, -1000, -1000, -1000, -1000,
	-120, -122, -124, -1000, -1000, -1000, -1000, -1000, 1110, 1288,
	-1000, 296, 203, 168, 162, 292, -1000, -1000, -1000, 184,
	94, -1000, -1000, -1000, -1000, -1000, 282, 285, 281, 275,
	885, 885, -1000, 1138, 1110, 476, 347, -1000, -1000, 976,
	74, -1000, 749, -1000, 528, 1110, 1110, 1110, -1000, -1000,
	346, 317, 315, 1288, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 274, -1000, -1000, -1000, 1374, 1374, 1138, 553,
	-1000, 1110, 1110, 1110, -1000, -1000, -1000, 521, 1138, 70,
	1138, 403, 403, 403, -1000, 273, 272, 271, 270, 265,
	264, 404, 1138, 1138, -1000, 471, 160, -1000, 151, 114,
	-1000, -1000, -1000, -1000, -1000, -1000, 55, -1000, 551, -34,
	-1000, 403, -1000, -1000, -1000, 403, -1000, 403, -1000,
}
var yyPgo = [...]int{

	0, 658, 657, 55, 656, 655, 654, 652, 629, 624,
	623, 619, 506, 617, 616, 18, 612, 26, 13, 610,
	606, 78, 605, 604, 14, 603, 602, 344, 601, 2,
	20, 22, 600, 9, 16, 7, 599, 598, 4, 6,
	8, 10, 25, 597, 51, 593, 592, 12, 591, 590,
	589, 587, 5, 584, 3, 581, 1, 580, 17, 579,
	11, 23, 53, 329, 578, 577, 576, 575, 574, 573,
	0, 31, 572, 571, 570, 569, 568, 567, 241, 15,
	566, 565, 564, 563, 561,
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
	1, 1, 2, 1, 2, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 5,
	0, 1, 1, 2, 4, 0, 2, 1, 3, 1,
	1, 1, 2, 2, 2, 2, 1, 1, 1, 1,
	1, 0, 3, 0, 2, 0, 3, 1, 3, 2,
	0, 1, 1, 0, 2, 4, 4, 0, 2, 4,
	0, 3, 1, 3, 0, 5, 1, 3, 3, 0,
	2, 0, 3, 0, 1, 1, 1, 1, 1, 1,
	0, 1, 0, 1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -72, -73, -74, -75, -76, -77, 5, 6,
	7, 8, 35, 160, 161, 163, 162, 145, 146, 147,
	149, 151, 150, -14, 97, 98, 99, 100, -12, -84,
	-12, -12, -12, -12, 164, -68, 166, 170, -65, 166,
	168, 164, 164, 165, 166, -12, 152, -83, 153, 154,
	-82, 157, 158, 156, -70, 37, -3, 22, -16, 23,
	-13, 31, -27, 37, 96, -61, 148, -62, -44, -70,
	37, -64, 169, 165, -70, 164, -70, 37, -63, 169,
	-70, -63, 31, -79, 9, 122, 155, -78, 96, -70,
	159, 42, -17, -18, 134, -21, 37, -22, -32, -44,
	-33, 113, -43, 26, -45, -70, -37, 49, 50, 51,
	74, 75, 45, 61, 62, 63, 65, 38, 39, 40,
	47, 48, 46, 27, -34, 42, -42, 132, 133, 44,
	140, 169, 30, 105, 104, 135, -38, 19, 20, 21,
	52, 53, 54, 55, 56, 57, 58, 59, 60, 41,
	-27, 35, 138, -27, 101, 37, 119, 138, 37, 113,
	-70, -71, 37, -71, 167, 37, 26, 112, -70, -27,
	-21, -21, -79, -79, -21, -78, -80, 96, 124, -35,
	-21, 96, 101, -19, 116, 129, 130, 131, 132, 133,
	134, 136, 137, 135, 119, 118, 120, 125, 126, 127,
	128, 114, 115, 124, 113, 122, 121, 123, 117, -70,
	25, 138, -21, -21, -42, 42, 42, 42, 42, 42,
	42, 42, 42, 42, 38, 42, 42, 42, 42, 38,
	38, 38, -35, -3, -48, -21, -58, 35, 42, -61,
	37, -30, 9, -62, -21, -70, -71, 26, -69, 171,
	-66, 163, 161, 34, 162, 12, 37, 37, 37, -71,
	-58, 35, -79, -81, 96, 124, -21, -21, 43, 101,
	-23, -24, -26, 42, 37, -42, 159, 153, -18, 24,
	-21, -21, -21, -21, -21, -21, -21, -21, -21, -21,
	-21, -15, 22, 17, 18, -21, -15, -21, -15, -21,
	-15, -21, -15, -21, -15, -21, -15, -21, -21, -33,
	124, 122, 123, 117, -21, 27, 113, -34, -21, -21,
	-70, 134, 43, -17, 23, -17, -17, 43, -38, -39,
	66, 67, 68, 69, 70, 71, 72, 73, -38, -39,
	-21, -21, -18, -38, -40, -39, 85, 86, 87, 88,
	89, 90, 91, 92, 93, 94, 95, -18, -18, -17,
	43, 43, -46, -47, 141, -31, 30, -3, -61, -59,
	-44, -30, -52, 12, -21, 112, -70, -71, -67, 167,
	-31, -61, -21, -21, -21, -30, 101, -25, 102, 103,
	104, 105, 106, 108, 109, -20, 37, 25, -24, 138,
	-42, -42, -42, -42, -42, -42, -42, -33, -21, -21,
	-21, 27, -34, 116, 43, -17, 43, 43, 101, 101,
	101, 101, 101, 25, 43, 96, 96, 96, 101, 101,
	43, -49, -47, 143, -21, -60, 112, -36, -33, -60,
	43, 101, -52, -56, 14, 13, 37, 37, -50, 10,
	-24, -24, 102, 107, 102, 107, 102, 102, 102, -28,
	110, 168, 111, 37, 43, 37, 159, 153, 116, -21,
	43, -17, -17, -17, -17, -41, 76, 45, 46, 77,
	78, 79, 80, 81, 82, 83, -41, -18, -18, -18,
	64, 64, 144, -21, 142, 32, 101, -44, -56, -21,
	-53, -54, -21, -71, -51, 11, 13, 112, 102, 102,
	165, 165, 165, -21, 43, 43, 43, 43, 43, 84,
	84, 43, 24, 43, 43, 43, -18, -18, -21, 33,
	-33, 101, 15, 101, -55, 28, 29, -52, -21, -35,
	-21, 42, 42, 42, 43, -38, -40, -39, -38, -40,
	-39, 7, -21, -21, -54, -56, -29, -70, -29, -29,
	43, 43, 43, 43, 43, 43, -61, -57, 16, 36,
	43, 101, 43, 43, 7, 124, -70, -70, -70,
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
	264, 265, 0, 0, 0, 251, 0, 0, 0, 110,
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
	124, 126, 255, 252, 0, 294, 0, 121, 294, 0,
	292, 275, 283, 0, 111, 0, 315, 50, 0, 311,
	22, 23, 35, 36, 128, 271, 0, 0, 86, 87,
	0, 0, 0, 0, 0, 104, 84, 0, 0, 0,
	144, 146, 148, 150, 152, 154, 156, 162, 164, 170,
	0, 166, 168, 0, 176, 0, 178, 180, 0, 0,
	0, 0, 0, 0, 189, 0, 0, 0, 0, 0,
	199, 0, 253, 0, 0, 20, 0, 120, 122, 21,
	291, 0, 283, 25, 0, 0, 317, 51, 273, 0,
	78, 81, 88, 0, 90, 0, 92, 93, 94, 79,
	0, 0, 0, 85, 80, 96, 100, 101, 0, 172,
	177, 0, 0, 0, 0, 0, 228, 229, 230, 231,
	233, 235, 236, 237, 238, 239, 0, 0, 0, 0,
	0, 0, 249, 256, 0, 0, 0, 293, 24, 284,
	276, 277, 280, 48, 275, 0, 0, 0, 89, 91,
	0, 0, 0, 173, 182, 183, 184, 185, 186, 232,
	234, 187, 0, 190, 191, 192, 0, 0, 254, 0,
	123, 0, 0, 0, 279, 281, 282, 283, 274, 272,
	82, 0, 0, 0, 188, 0, 0, 0, 0, 0,
	0, 0, 285, 286, 278, 287, 0, 108, 0, 0,
	193, 194, 195, 196, 197, 198, 295, 18, 0, 0,
	105, 0, 106, 107, 288, 0, 109, 0, 289,
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
		//line sql.y:898
		{
			yyVAL.expr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:902
		{
			yyVAL.expr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 159:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:906
		{
			yyVAL.expr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 160:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:910
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
		//line sql.y:925
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 162:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:929
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 163:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:933
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LIKE, Right: yyDollar[3].expr}
		}
	case 164:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:937
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_LIKE, Right: yyDollar[4].expr}
		}
	case 165:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:941
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 166:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:945
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 167:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:949
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: yyDollar[3].expr}
		}
	case 168:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:953
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: yyDollar[4].expr}
		}
	case 169:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:957
		{
			yyVAL.expr = &RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[3].expr}
		}
	case 170:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:961
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[4].expr}}
		}
	case 171:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:965
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 172:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:969
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_BETWEEN, From: yyDollar[3].expr, To: yyDollar[5].expr}
		}
	case 173:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:973
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].expr, To: yyDollar[6].expr}
		}
	case 174:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:977
		{
			yyVAL.expr = yyDollar[1].caseExpr
		}
	case 175:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:981
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 176:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:985
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 177:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:989
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 178:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:993
		{
			yyVAL.expr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 179:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:997
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 180:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1001
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 181:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1005
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 182:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1009
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 183:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1013
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 184:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1017
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 185:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1021
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 186:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1025
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 187:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1029
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 188:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:1033
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 189:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1037
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date"), Exprs: SelectExprs{yyDollar[3].selectExpr}}
		}
	case 190:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1041
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 191:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1045
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 192:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1049
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 193:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1053
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 194:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1057
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 195:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1061
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 196:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1065
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 197:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1069
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 198:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1073
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 199:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1077
		{
			yyVAL.expr = &FuncExpr{Name: []byte("str_to_date"), Exprs: yyDollar[3].selectExprs}
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1083
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1087
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1091
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1095
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1099
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 205:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1103
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1107
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1111
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1117
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 209:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1121
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1125
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1129
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1133
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1137
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1141
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1145
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1149
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1155
		{
			yyVAL.bytes = SECOND_MICROSECOND_BYTES
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1159
		{
			yyVAL.bytes = MINUTE_MICROSECOND_BYTES
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1163
		{
			yyVAL.bytes = MINUTE_SECOND_BYTES
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1167
		{
			yyVAL.bytes = HOUR_MICROSECOND_BYTES
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1171
		{
			yyVAL.bytes = HOUR_SECOND_BYTES
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1175
		{
			yyVAL.bytes = HOUR_MINUTE_BYTES
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1179
		{
			yyVAL.bytes = DAY_MICROSECOND_BYTES
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1183
		{
			yyVAL.bytes = DAY_SECOND_BYTES
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1187
		{
			yyVAL.bytes = DAY_MINUTE_BYTES
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1191
		{
			yyVAL.bytes = DAY_HOUR_BYTES
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1195
		{
			yyVAL.bytes = YEAR_MONTH_BYTES
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1201
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1205
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1209
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 231:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1213
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 232:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1217
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1221
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 234:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1225
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1229
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1233
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1237
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1241
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1245
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1251
		{
			yyVAL.bytes = IF_BYTES
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1255
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1259
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1263
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1267
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1271
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 246:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1277
		{
			yyVAL.byt = AST_UPLUS
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1281
		{
			yyVAL.byt = AST_UMINUS
		}
	case 248:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1285
		{
			yyVAL.byt = AST_TILDA
		}
	case 249:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1291
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].expr, Whens: yyDollar[3].whens, Else: yyDollar[4].expr}
		}
	case 250:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1296
		{
			yyVAL.expr = nil
		}
	case 251:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1300
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1306
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 253:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1310
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 254:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1316
		{
			yyVAL.when = &When{Cond: yyDollar[2].expr, Val: yyDollar[4].expr}
		}
	case 255:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1321
		{
			yyVAL.expr = nil
		}
	case 256:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1325
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1331
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 258:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1335
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1341
		{
			yyVAL.expr = StrVal(yyDollar[1].bytes)
		}
	case 260:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1345
		{
			yyVAL.expr = NumVal(yyDollar[1].bytes)
		}
	case 261:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1349
		{
			yyVAL.expr = ValArg(yyDollar[1].bytes)
		}
	case 262:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1353
		{
			yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 263:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1357
		{
			yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 264:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1361
		{
			yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 265:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1365
		{
			yyVAL.expr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 266:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1369
		{
			yyVAL.expr = &NullVal{}
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1373
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 268:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1379
		{
			yyVAL.expr = &TrueVal{}
		}
	case 269:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1383
		{
			yyVAL.expr = &FalseVal{}
		}
	case 270:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1387
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 271:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1392
		{
			yyVAL.exprs = nil
		}
	case 272:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1396
		{
			yyVAL.exprs = yyDollar[3].exprs
		}
	case 273:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1401
		{
			yyVAL.expr = nil
		}
	case 274:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1405
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 275:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1410
		{
			yyVAL.orderBy = nil
		}
	case 276:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1414
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 277:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1420
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 278:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1424
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 279:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1430
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 280:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1435
		{
			yyVAL.str = AST_ASC
		}
	case 281:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1439
		{
			yyVAL.str = AST_ASC
		}
	case 282:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1443
		{
			yyVAL.str = AST_DESC
		}
	case 283:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1448
		{
			yyVAL.limit = nil
		}
	case 284:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1452
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].expr}
		}
	case 285:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1456
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].expr, Rowcount: yyDollar[4].expr}
		}
	case 286:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1460
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].expr, Rowcount: yyDollar[2].expr}
		}
	case 287:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1465
		{
			yyVAL.str = ""
		}
	case 288:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1469
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 289:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1473
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
		//line sql.y:1486
		{
			yyVAL.columns = nil
		}
	case 291:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1490
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 292:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1496
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 293:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1500
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 294:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1505
		{
			yyVAL.updateExprs = nil
		}
	case 295:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1509
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 296:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1515
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 297:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1519
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 298:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1525
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].expr}
		}
	case 299:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1530
		{
			yyVAL.empty = struct{}{}
		}
	case 300:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1532
		{
			yyVAL.empty = struct{}{}
		}
	case 301:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1535
		{
			yyVAL.empty = struct{}{}
		}
	case 302:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1537
		{
			yyVAL.empty = struct{}{}
		}
	case 303:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1540
		{
			yyVAL.empty = struct{}{}
		}
	case 304:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1542
		{
			yyVAL.empty = struct{}{}
		}
	case 305:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1546
		{
			yyVAL.empty = struct{}{}
		}
	case 306:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1548
		{
			yyVAL.empty = struct{}{}
		}
	case 307:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1550
		{
			yyVAL.empty = struct{}{}
		}
	case 308:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1552
		{
			yyVAL.empty = struct{}{}
		}
	case 309:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1554
		{
			yyVAL.empty = struct{}{}
		}
	case 310:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1557
		{
			yyVAL.empty = struct{}{}
		}
	case 311:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1559
		{
			yyVAL.empty = struct{}{}
		}
	case 312:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1562
		{
			yyVAL.empty = struct{}{}
		}
	case 313:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1564
		{
			yyVAL.empty = struct{}{}
		}
	case 314:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1567
		{
			yyVAL.empty = struct{}{}
		}
	case 315:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1569
		{
			yyVAL.empty = struct{}{}
		}
	case 316:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1573
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 317:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1578
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
