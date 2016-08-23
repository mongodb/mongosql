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
const TRANSACTION = 57492
const ISOLATION = 57493
const LEVEL = 57494
const READ = 57495
const WRITE = 57496
const ONLY = 57497
const REPEATABLE = 57498
const COMMITTED = 57499
const UNCOMMITTED = 57500
const SERIALIZABLE = 57501
const NAMES = 57502
const REPLACE = 57503
const ADMIN = 57504
const SHOW = 57505
const DATABASES = 57506
const TABLES = 57507
const PROXY = 57508
const VARIABLES = 57509
const FULL = 57510
const COLUMNS = 57511
const SESSION = 57512
const GLOBAL = 57513
const CREATE = 57514
const ALTER = 57515
const DROP = 57516
const RENAME = 57517
const TABLE = 57518
const INDEX = 57519
const VIEW = 57520
const TO = 57521
const IGNORE = 57522
const IF = 57523
const UNIQUE = 57524
const USING = 57525

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
	"TRANSACTION",
	"ISOLATION",
	"LEVEL",
	"READ",
	"WRITE",
	"ONLY",
	"REPEATABLE",
	"COMMITTED",
	"UNCOMMITTED",
	"SERIALIZABLE",
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
	"SESSION",
	"GLOBAL",
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
	-1, 22,
	150, 49,
	-2, 67,
	-1, 31,
	169, 47,
	-2, 49,
}

const yyNprod = 328
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1437

var yyAct = [...]int{

	117, 470, 586, 531, 76, 364, 195, 393, 177, 505,
	384, 138, 462, 291, 111, 171, 386, 257, 249, 3,
	148, 95, 349, 47, 252, 49, 136, 269, 107, 50,
	490, 492, 91, 65, 84, 52, 87, 53, 405, 275,
	180, 112, 542, 541, 80, 55, 56, 57, 540, 85,
	86, 54, 67, 88, 495, 18, 102, 92, 79, 104,
	294, 273, 98, 101, 276, 293, 105, 59, 61, 62,
	473, 64, 81, 45, 46, 45, 46, 399, 474, 475,
	398, 78, 82, 400, 264, 265, 176, 294, 172, 522,
	173, 263, 293, 385, 184, 460, 385, 179, 491, 201,
	202, 203, 204, 205, 206, 209, 207, 208, 225, 204,
	205, 206, 209, 207, 208, 206, 209, 207, 208, 66,
	188, 189, 425, 227, 169, 186, 187, 230, 311, 190,
	164, 284, 196, 18, 19, 20, 21, 193, 605, 175,
	168, 539, 228, 229, 248, 213, 214, 215, 216, 201,
	202, 203, 204, 205, 206, 209, 207, 208, 537, 285,
	463, 166, 401, 22, 80, 194, 196, 80, 255, 96,
	261, 251, 463, 414, 415, 416, 417, 418, 79, 419,
	420, 79, 497, 266, 333, 183, 496, 538, 297, 331,
	332, 330, 296, 279, 484, 77, 488, 260, 482, 485,
	272, 274, 271, 483, 34, 35, 36, 37, 487, 295,
	280, 486, 603, 282, 602, 297, 166, 289, 341, 296,
	600, 547, 286, 287, 563, 526, 455, 340, 261, 300,
	301, 302, 303, 304, 305, 306, 307, 308, 309, 310,
	315, 317, 319, 321, 323, 325, 327, 328, 259, 337,
	334, 546, 338, 339, 80, 80, 348, 358, 389, 359,
	454, 329, 363, 99, 365, 298, 360, 361, 79, 391,
	402, 388, 601, 392, 601, 27, 28, 29, 396, 403,
	601, 198, 80, 197, 97, 494, 407, 395, 198, 30,
	32, 31, 343, 345, 346, 447, 79, 406, 382, 388,
	23, 24, 26, 25, 379, 295, 362, 424, 411, 377,
	378, 198, 258, 408, 409, 258, 446, 445, 410, 149,
	150, 151, 444, 426, 262, 453, 452, 335, 427, 451,
	428, 100, 429, 75, 430, 545, 431, 544, 432, 500,
	316, 318, 320, 322, 324, 326, 414, 415, 416, 417,
	418, 467, 419, 420, 34, 35, 36, 37, 456, 550,
	434, 435, 436, 438, 549, 191, 457, 521, 443, 595,
	442, 520, 433, 224, 211, 210, 212, 222, 221, 223,
	219, 213, 214, 215, 216, 201, 202, 203, 204, 205,
	206, 209, 207, 208, 459, 198, 440, 198, 381, 198,
	469, 288, 466, 73, 441, 90, 412, 594, 593, 166,
	592, 468, 591, 590, 461, 336, 574, 555, 198, 554,
	553, 548, 450, 573, 295, 295, 480, 481, 198, 465,
	198, 220, 217, 218, 200, 224, 211, 210, 212, 222,
	221, 223, 219, 213, 214, 215, 216, 201, 202, 203,
	204, 205, 206, 209, 207, 208, 198, 281, 289, 516,
	524, 289, 552, 93, 254, 572, 253, 515, 499, 80,
	240, 528, 380, 254, 239, 571, 162, 507, 137, 165,
	161, 551, 231, 527, 244, 533, 243, 242, 241, 523,
	238, 237, 236, 235, 234, 233, 232, 103, 185, 246,
	529, 532, 245, 66, 501, 502, 503, 504, 506, 508,
	509, 510, 511, 512, 513, 514, 423, 81, 517, 518,
	519, 598, 493, 477, 476, 278, 277, 543, 422, 256,
	152, 153, 154, 155, 156, 157, 158, 159, 160, 247,
	74, 599, 567, 569, 350, 351, 352, 353, 354, 355,
	356, 357, 181, 558, 178, 174, 167, 89, 163, 559,
	525, 94, 576, 579, 568, 196, 570, 584, 560, 585,
	72, 44, 587, 587, 587, 588, 589, 575, 578, 577,
	580, 267, 80, 182, 18, 299, 596, 556, 557, 70,
	582, 583, 532, 149, 150, 151, 79, 344, 68, 471,
	115, 135, 606, 60, 144, 536, 607, 472, 608, 387,
	394, 108, 129, 130, 131, 535, 137, 342, 134, 479,
	141, 124, 38, 132, 133, 119, 120, 121, 152, 153,
	154, 155, 156, 157, 158, 159, 160, 125, 126, 127,
	258, 128, 40, 41, 42, 43, 149, 150, 151, 604,
	122, 123, 581, 58, 437, 18, 39, 63, 283, 192,
	17, 16, 15, 14, 13, 12, 268, 48, 404, 565,
	566, 270, 51, 83, 397, 170, 390, 597, 564, 530,
	146, 145, 534, 478, 458, 250, 383, 116, 114, 113,
	118, 464, 313, 314, 149, 150, 151, 312, 110, 489,
	292, 115, 135, 413, 290, 144, 109, 421, 139, 140,
	106, 147, 81, 129, 130, 131, 142, 137, 199, 134,
	69, 141, 124, 33, 132, 133, 119, 120, 121, 152,
	153, 154, 155, 156, 157, 158, 159, 160, 125, 126,
	127, 71, 128, 11, 10, 9, 8, 7, 6, 5,
	4, 122, 123, 2, 1, 143, 220, 217, 218, 200,
	224, 211, 210, 212, 222, 221, 223, 219, 213, 214,
	215, 216, 201, 202, 203, 204, 205, 206, 209, 207,
	208, 146, 145, 0, 0, 0, 0, 0, 0, 0,
	113, 0, 0, 0, 0, 149, 150, 151, 0, 0,
	0, 0, 115, 135, 0, 0, 144, 0, 0, 139,
	140, 0, 147, 108, 129, 130, 131, 142, 137, 347,
	134, 0, 141, 124, 0, 132, 133, 119, 120, 121,
	152, 153, 154, 155, 156, 157, 158, 159, 160, 125,
	126, 127, 0, 128, 0, 0, 0, 0, 0, 0,
	0, 0, 122, 123, 0, 0, 143, 0, 562, 200,
	224, 211, 210, 212, 222, 221, 223, 219, 213, 214,
	215, 216, 201, 202, 203, 204, 205, 206, 209, 207,
	208, 0, 146, 145, 0, 0, 0, 0, 0, 0,
	0, 113, 0, 0, 0, 0, 149, 150, 151, 0,
	0, 0, 0, 115, 135, 0, 0, 144, 0, 0,
	139, 140, 106, 147, 108, 129, 130, 131, 142, 137,
	0, 134, 0, 141, 124, 0, 132, 133, 119, 120,
	121, 152, 153, 154, 155, 156, 157, 158, 159, 160,
	125, 126, 127, 0, 128, 0, 561, 0, 0, 0,
	0, 0, 0, 122, 123, 0, 0, 143, 220, 217,
	218, 200, 224, 211, 210, 212, 222, 221, 223, 219,
	213, 214, 215, 216, 201, 202, 203, 204, 205, 206,
	209, 207, 208, 146, 145, 0, 0, 0, 0, 0,
	0, 0, 113, 0, 0, 0, 0, 0, 226, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	66, 139, 140, 106, 147, 18, 0, 0, 0, 142,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 149,
	150, 151, 0, 0, 0, 0, 115, 135, 0, 0,
	144, 0, 0, 0, 0, 0, 0, 81, 129, 130,
	131, 0, 137, 0, 134, 0, 141, 124, 143, 132,
	133, 119, 120, 121, 152, 153, 154, 155, 156, 157,
	158, 159, 160, 125, 126, 127, 0, 128, 0, 0,
	0, 0, 0, 0, 0, 0, 122, 123, 220, 217,
	218, 200, 224, 211, 210, 212, 222, 221, 223, 219,
	213, 214, 215, 216, 201, 202, 203, 204, 205, 206,
	209, 207, 208, 0, 0, 0, 146, 145, 0, 0,
	0, 0, 0, 0, 0, 113, 0, 0, 0, 0,
	149, 150, 151, 0, 0, 0, 0, 115, 135, 0,
	0, 144, 0, 0, 139, 140, 0, 147, 81, 129,
	130, 131, 142, 137, 0, 134, 0, 141, 124, 0,
	132, 133, 119, 120, 121, 152, 153, 154, 155, 156,
	157, 158, 159, 160, 125, 126, 127, 449, 128, 0,
	448, 0, 0, 0, 0, 0, 0, 122, 123, 0,
	0, 143, 220, 217, 218, 200, 224, 211, 210, 212,
	222, 221, 223, 219, 213, 214, 215, 216, 201, 202,
	203, 204, 205, 206, 209, 207, 208, 146, 145, 0,
	0, 0, 0, 0, 0, 0, 113, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 139, 140, 0, 147, 0,
	0, 0, 0, 142, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 220, 217, 218,
	200, 224, 211, 210, 212, 222, 221, 223, 219, 213,
	214, 215, 216, 201, 202, 203, 204, 205, 206, 209,
	207, 208, 143, 220, 217, 218, 200, 224, 211, 210,
	212, 222, 221, 223, 219, 213, 214, 215, 216, 201,
	202, 203, 204, 205, 206, 209, 207, 208, 220, 217,
	218, 498, 224, 211, 210, 212, 222, 221, 223, 219,
	213, 214, 215, 216, 201, 202, 203, 204, 205, 206,
	209, 207, 208, 220, 217, 218, 439, 224, 211, 210,
	212, 222, 221, 223, 219, 213, 214, 215, 216, 201,
	202, 203, 204, 205, 206, 209, 207, 208, 220, 217,
	218, 200, 224, 211, 210, 212, 222, 221, 223, 0,
	213, 214, 215, 216, 201, 202, 203, 204, 205, 206,
	209, 207, 208, 152, 153, 154, 155, 156, 157, 158,
	159, 160, 0, 0, 0, 0, 0, 350, 351, 352,
	353, 354, 355, 356, 357, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 366, 367, 368, 369,
	370, 371, 372, 373, 374, 375, 376,
}
var yyPact = [...]int{

	128, -1000, -1000, 105, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -95, -153, -143, -125, -131, -1000, -1000, -1000,
	-1000, -97, 466, 650, 576, -1000, -1000, -1000, 566, -1000,
	539, 503, 235, 35, -68, -1000, -1000, -147, -128, 466,
	-1000, -140, 466, -1000, 520, -149, 466, -149, 530, 160,
	-105, 233, 466, -113, -1000, 455, -1000, -1000, -1000, 877,
	-1000, 439, 503, 523, -10, 503, 113, 519, -1000, 19,
	-1000, -16, -63, 518, 24, 466, -1000, 517, -1000, -139,
	515, 557, 71, 466, 503, -1000, 1111, 1111, 160, 160,
	1111, 233, 39, 1111, 185, -1000, -1000, 973, -17, -1000,
	-1000, -1000, -1000, 1111, 1111, 440, -1000, 454, 453, 452,
	451, 450, 449, 448, 432, 446, 445, 444, 442, -1000,
	-1000, -1000, 464, 461, 502, -1000, -1000, 1010, -1000, -1000,
	-1000, -1000, 1111, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 431, 480, 492, 631, 480, -1000, 1111, 466,
	221, -1000, -61, -70, -1000, 555, -156, -1000, 27, -1000,
	489, -1000, -1000, 488, -1000, 422, 1178, 1178, -1000, -1000,
	1178, 160, 33, 1111, 1111, 358, 1178, 23, 877, 561,
	1111, 1111, 1111, 1111, 1111, 1111, 1111, 1111, 1111, 1111,
	675, 675, 675, 675, 675, 675, 675, 1111, 1111, 436,
	65, 1111, 300, 1111, 1111, -1000, 466, 82, 1178, -1000,
	-1000, 650, 574, 877, 776, 476, 476, 1111, 1111, 877,
	-1000, 1339, 877, 877, 877, -1000, -1000, 434, 355, 255,
	-47, 1178, 579, 480, 480, 306, -1000, 598, 1111, -1000,
	1178, -1000, -63, -76, -1000, -1000, -1000, -1000, 48, 466,
	-1000, -141, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	579, 480, -1000, -1000, 1111, 1111, 1178, 1253, -1000, 1111,
	303, 69, 491, 50, -18, -1000, -1000, -1000, -1000, -1000,
	1178, -25, -25, -25, -21, -21, -1000, -1000, -1000, -1000,
	18, 440, -1000, -1000, -1000, 18, 440, 18, 440, -32,
	440, -32, 440, -32, 440, -32, 440, 741, 741, -1000,
	436, 1111, 1111, 1111, 18, -1000, 627, -1000, 18, 1228,
	-1000, -1000, -1000, 353, 877, 327, 325, -1000, 219, 214,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 213, 192,
	1077, 1152, 379, 231, 228, 227, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 157, 123, 315,
	321, -1000, -1000, -50, -1000, 1111, 46, 436, 105, 58,
	308, -1000, 598, 585, 594, 1178, -1000, -1000, -83, -79,
	-1000, 487, -1000, -1000, 486, -1000, -1000, 113, 1178, 1178,
	1178, 609, 23, 23, -1000, -1000, 94, 90, 107, 104,
	92, -82, -1000, 485, 242, 17, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 18, 18, 1203, -1000, -1000, 1111,
	-1000, 296, -1000, -1000, 877, 877, 877, 877, 430, 430,
	-1000, 877, 877, 877, 305, 301, -1000, -1000, -57, -1000,
	1111, 316, -1000, 528, 122, -1000, -1000, -1000, 480, 585,
	-1000, 1111, 1111, -1000, -1000, -1000, -1000, -1000, 604, 592,
	69, 44, -1000, 83, -1000, 37, -1000, -1000, -1000, -1000,
	-129, -134, -135, -1000, -1000, -1000, -1000, -1000, 1111, 254,
	-1000, 294, 292, 208, 178, 378, -1000, -1000, 278, 273,
	-1000, -1000, -1000, -1000, -1000, -1000, 438, 377, 376, 374,
	877, 877, -1000, 1178, 1111, 526, 436, -1000, -1000, 843,
	121, -1000, 641, -1000, 598, 1111, 1111, 1111, -1000, -1000,
	433, 423, 381, 254, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 373, -1000, -1000, -1000, 1339, 1339, 1178, 645,
	-1000, 1111, 1111, 1111, -1000, -1000, -1000, 585, 1178, 114,
	1178, 466, 466, 466, -1000, 370, 369, 367, 365, 364,
	326, 480, 1178, 1178, -1000, 505, 177, -1000, 171, 169,
	-1000, -1000, -1000, -1000, -1000, -1000, 113, -1000, 642, 12,
	-1000, 466, -1000, -1000, -1000, 466, -1000, 466, -1000,
}
var yyPgo = [...]int{

	0, 754, 753, 18, 750, 749, 748, 747, 746, 745,
	744, 743, 622, 741, 723, 128, 720, 59, 66, 718,
	707, 28, 706, 704, 13, 703, 700, 403, 699, 2,
	17, 16, 698, 41, 26, 6, 691, 690, 20, 22,
	5, 9, 11, 688, 14, 687, 686, 10, 685, 684,
	683, 682, 7, 679, 3, 678, 1, 677, 24, 676,
	12, 4, 81, 675, 15, 674, 405, 673, 672, 671,
	668, 667, 666, 0, 8, 665, 664, 663, 662, 661,
	660, 263, 21, 659, 658, 657, 571, 656,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 78, 78, 5, 6, 7, 7, 7, 63,
	63, 64, 64, 64, 65, 65, 65, 65, 75, 76,
	77, 80, 83, 83, 84, 84, 84, 85, 85, 86,
	86, 86, 79, 79, 79, 79, 79, 8, 8, 8,
	9, 9, 9, 10, 11, 11, 11, 87, 12, 13,
	13, 14, 14, 14, 14, 14, 16, 16, 17, 17,
	18, 18, 18, 18, 19, 19, 19, 23, 23, 24,
	24, 24, 24, 20, 20, 20, 25, 25, 25, 25,
	25, 25, 25, 25, 25, 26, 26, 26, 26, 26,
	26, 26, 27, 27, 28, 28, 28, 28, 29, 29,
	30, 30, 82, 82, 82, 81, 81, 15, 15, 15,
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
	58, 58, 59, 59, 60, 60, 61, 61, 62, 66,
	66, 67, 67, 68, 68, 69, 69, 69, 69, 69,
	70, 70, 71, 71, 72, 72, 73, 74,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 4, 12, 3,
	7, 7, 6, 6, 8, 7, 3, 4, 4, 1,
	3, 3, 2, 2, 2, 2, 2, 1, 1, 1,
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
	-10, -11, -75, -76, -77, -78, -79, -80, 5, 6,
	7, 8, 35, 172, 173, 175, 174, 147, 148, 149,
	161, 163, 162, -14, 99, 100, 101, 102, -12, -87,
	-12, -12, -12, -12, -86, 170, 171, 176, -71, 178,
	182, -68, 178, 180, 176, 176, 177, 178, -12, 164,
	-86, 165, 166, -85, 168, -73, 37, -3, 22, -16,
	23, -13, 31, -27, 37, 98, -61, 160, -62, -44,
	-73, 37, 150, -67, 181, 177, -73, 176, -73, 37,
	-66, 181, -73, -66, 31, -82, 9, 124, 167, -81,
	98, -73, 169, 42, -17, -18, 136, -21, 37, -22,
	-32, -44, -33, 115, -43, 26, -45, -73, -37, 51,
	52, 53, 76, 77, 47, 63, 64, 65, 67, 38,
	39, 40, 49, 50, 44, 27, -34, 42, -42, 134,
	135, 46, 142, 181, 30, 107, 106, 137, -38, 19,
	20, 21, 54, 55, 56, 57, 58, 59, 60, 61,
	62, 41, -27, 35, 140, -27, 103, 37, 121, 140,
	-63, -64, 151, 153, 37, 115, -73, -74, 37, -74,
	179, 37, 26, 114, -73, -27, -21, -21, -82, -82,
	-21, -81, -83, 98, 126, -35, -21, 98, 103, -19,
	118, 131, 132, 133, 134, 135, 136, 138, 139, 137,
	121, 120, 122, 127, 128, 129, 130, 116, 117, 126,
	115, 124, 123, 125, 119, -73, 25, 140, -21, -21,
	-42, 42, 42, 42, 42, 42, 42, 42, 42, 42,
	38, 42, 42, 42, 42, 38, 38, 37, -35, -3,
	-48, -21, -58, 35, 42, -61, 37, -30, 9, -62,
	-21, -73, 103, 152, 154, 155, -74, 26, -72, 183,
	-69, 175, 173, 34, 174, 12, 37, 37, 37, -74,
	-58, 35, -82, -84, 98, 126, -21, -21, 43, 103,
	-23, -24, -26, 42, 37, -42, 169, 165, -18, 24,
	-21, -21, -21, -21, -21, -21, -21, -21, -21, -21,
	-21, -15, 22, 17, 18, -21, -15, -21, -15, -21,
	-15, -21, -15, -21, -15, -21, -15, -21, -21, -33,
	126, 124, 125, 119, -21, 27, 115, -34, -21, -21,
	-73, 136, 43, -17, 23, -17, -17, 43, -38, -39,
	68, 69, 70, 71, 72, 73, 74, 75, -38, -39,
	-21, -21, -18, -38, -40, -39, 87, 88, 89, 90,
	91, 92, 93, 94, 95, 96, 97, -18, -18, -17,
	38, 43, 43, -46, -47, 143, -31, 30, -3, -61,
	-59, -44, -30, -52, 12, -21, -64, -65, 156, 153,
	159, 114, -73, -74, -70, 179, -31, -61, -21, -21,
	-21, -30, 103, -25, 104, 105, 106, 107, 108, 110,
	111, -20, 37, 25, -24, 140, -42, -42, -42, -42,
	-42, -42, -42, -33, -21, -21, -21, 27, -34, 118,
	43, -17, 43, 43, 103, 103, 103, 103, 103, 25,
	43, 98, 98, 98, 103, 103, 43, 45, -49, -47,
	145, -21, -60, 114, -36, -33, -60, 43, 103, -52,
	-56, 14, 13, 153, 157, 158, 37, 37, -50, 10,
	-24, -24, 104, 109, 104, 109, 104, 104, 104, -28,
	112, 180, 113, 37, 43, 37, 169, 165, 118, -21,
	43, -17, -17, -17, -17, -41, 78, 47, 79, 80,
	81, 82, 83, 84, 85, 37, -41, -18, -18, -18,
	66, 66, 146, -21, 144, 32, 103, -44, -56, -21,
	-53, -54, -21, -74, -51, 11, 13, 114, 104, 104,
	177, 177, 177, -21, 43, 43, 43, 43, 43, 86,
	86, 43, 24, 43, 43, 43, -18, -18, -21, 33,
	-33, 103, 15, 103, -55, 28, 29, -52, -21, -35,
	-21, 42, 42, 42, 43, -38, -40, -39, -38, -40,
	-39, 7, -21, -21, -54, -56, -29, -73, -29, -29,
	43, 43, 43, 43, 43, 43, -61, -57, 16, 36,
	43, 103, 43, 43, 7, 126, -73, -73, -73,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 67, 67,
	67, 67, -2, 322, 313, 0, 0, 38, 39, 40,
	67, -2, 0, 0, 71, 73, 74, 75, 76, 69,
	0, 0, 0, 0, 0, 50, 51, 311, 0, 0,
	323, 0, 0, 314, 0, 309, 0, 309, 0, 122,
	0, 125, 0, 0, 48, 0, 326, 19, 72, 0,
	77, 68, 0, 0, 112, 0, 26, 0, 306, 0,
	267, 326, 0, 0, 0, 0, 327, 0, 327, 0,
	0, 0, 0, 0, 0, 52, 0, 0, 122, 122,
	0, 125, 0, 0, 17, 78, 80, 84, 326, 139,
	141, 142, 143, 0, 0, 0, 184, 267, 0, 189,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 269,
	270, 271, 0, 0, 0, 276, 277, 0, 135, 256,
	257, 258, 260, 250, 251, 252, 253, 254, 255, 278,
	279, 280, 218, 219, 220, 221, 222, 223, 224, 225,
	226, 70, 300, 0, 0, 120, 0, 27, 0, 0,
	28, 29, 0, 0, 327, 0, 324, 59, 0, 62,
	0, 64, 310, 0, 327, 300, 123, 124, 53, 54,
	126, 122, 44, 0, 0, 0, 137, 0, 0, 81,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 85, 0, 0, 169, 170,
	181, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	272, 0, 0, 0, 0, 273, 274, 0, 0, 0,
	0, 261, 0, 0, 0, 120, 113, 285, 0, 307,
	308, 268, 0, 0, 32, 33, 57, 312, 0, 0,
	327, 320, 315, 316, 317, 318, 319, 63, 65, 66,
	0, 0, 55, 56, 0, 0, 42, 43, 41, 0,
	120, 87, 93, 0, 105, 107, 108, 109, 79, 82,
	140, 144, 145, 146, 147, 148, 149, 150, 151, 152,
	153, 0, 127, 128, 129, 155, 0, 157, 0, 159,
	0, 161, 0, 163, 0, 165, 0, 167, 168, 171,
	0, 0, 0, 0, 173, 175, 0, 177, 179, 0,
	86, 83, 185, 0, 0, 0, 0, 191, 0, 0,
	210, 211, 212, 213, 214, 215, 216, 217, 0, 0,
	0, 0, 0, 0, 0, 0, 227, 228, 229, 230,
	231, 232, 233, 234, 235, 236, 237, 0, 0, 0,
	0, 134, 136, 265, 262, 0, 304, 0, 131, 304,
	0, 302, 285, 293, 0, 121, 30, 31, 0, 0,
	37, 0, 325, 60, 0, 321, 22, 23, 45, 46,
	138, 281, 0, 0, 96, 97, 0, 0, 0, 0,
	0, 114, 94, 0, 0, 0, 154, 156, 158, 160,
	162, 164, 166, 172, 174, 180, 0, 176, 178, 0,
	186, 0, 188, 190, 0, 0, 0, 0, 0, 0,
	199, 0, 0, 0, 0, 0, 209, 275, 0, 263,
	0, 0, 20, 0, 130, 132, 21, 301, 0, 293,
	25, 0, 0, 34, 35, 36, 327, 61, 283, 0,
	88, 91, 98, 0, 100, 0, 102, 103, 104, 89,
	0, 0, 0, 95, 90, 106, 110, 111, 0, 182,
	187, 0, 0, 0, 0, 0, 238, 239, 240, 242,
	244, 245, 246, 247, 248, 249, 0, 0, 0, 0,
	0, 0, 259, 266, 0, 0, 0, 303, 24, 294,
	286, 287, 290, 58, 285, 0, 0, 0, 99, 101,
	0, 0, 0, 183, 192, 193, 194, 195, 196, 241,
	243, 197, 0, 200, 201, 202, 0, 0, 264, 0,
	133, 0, 0, 0, 289, 291, 292, 293, 284, 282,
	92, 0, 0, 0, 198, 0, 0, 0, 0, 0,
	0, 0, 295, 296, 288, 297, 0, 118, 0, 0,
	203, 204, 205, 206, 207, 208, 305, 18, 0, 0,
	115, 0, 116, 117, 298, 0, 119, 0, 299,
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
	172, 173, 174, 175, 176, 177, 178, 179, 180, 181,
	182, 183,
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
		//line sql.y:222
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:228
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:248
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 18:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:252
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].exprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:256
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 20:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:263
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:267
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
		//line sql.y:279
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:283
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
		//line sql.y:296
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 25:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:302
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:308
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 27:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:312
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 28:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:316
		{
			yyVAL.statement = &Set{Comments: append([][]byte{}, []byte(yyDollar[2].str), []byte("transaction"), yyDollar[4].bytes)}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:322
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:326
		{
			yyVAL.bytes = append(yyDollar[1].bytes, append([]byte(", "), yyDollar[3].bytes...)...)
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:332
		{
			yyVAL.bytes = append([]byte("isolation level "), yyDollar[3].bytes...)
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:336
		{
			yyVAL.bytes = []byte("read write")
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:340
		{
			yyVAL.bytes = []byte("read only")
		}
	case 34:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:346
		{
			yyVAL.bytes = []byte("repeatable read")
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:350
		{
			yyVAL.bytes = []byte("read committed")
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:354
		{
			yyVAL.bytes = []byte("read uncommitted")
		}
	case 37:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:358
		{
			yyVAL.bytes = []byte("serializable")
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:364
		{
			yyVAL.statement = &Begin{}
		}
	case 39:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:370
		{
			yyVAL.statement = &Commit{}
		}
	case 40:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:376
		{
			yyVAL.statement = &Rollback{}
		}
	case 41:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:382
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].exprs}
		}
	case 42:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:388
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 43:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:392
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 44:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:397
		{
			yyVAL.expr = nil
		}
	case 45:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:401
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 46:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:405
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 47:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:410
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 48:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:414
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 49:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:419
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 50:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:423
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 51:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:427
		{
			yyVAL.str = AST_GLOBAL_SCOPE
		}
	case 52:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:434
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 53:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:438
		{
			yyVAL.statement = &Show{Section: "variables", Modifier: yyDollar[2].str, LikeOrWhere: yyDollar[4].expr}
		}
	case 54:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:442
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].expr, LikeOrWhere: yyDollar[4].expr}
		}
	case 55:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:446
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].expr, LikeOrWhere: yyDollar[5].expr}
		}
	case 56:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:450
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].expr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].expr}
		}
	case 57:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:456
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 58:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:460
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 59:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:465
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 60:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:471
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 61:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:475
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 62:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:480
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 63:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:486
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 64:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:492
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 65:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:496
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 66:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:501
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 67:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:506
		{
			SetAllowComments(yylex, true)
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:510
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 69:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:516
		{
			yyVAL.bytes2 = nil
		}
	case 70:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:520
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:526
		{
			yyVAL.str = AST_UNION
		}
	case 72:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:530
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 73:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:534
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:538
		{
			yyVAL.str = AST_EXCEPT
		}
	case 75:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:542
		{
			yyVAL.str = AST_INTERSECT
		}
	case 76:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:547
		{
			yyVAL.str = ""
		}
	case 77:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:551
		{
			yyVAL.str = AST_DISTINCT
		}
	case 78:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:557
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 79:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:561
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 80:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:567
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 81:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:571
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 82:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:575
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 83:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:579
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 84:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:584
		{
			yyVAL.bytes = nil
		}
	case 85:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:588
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 86:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:592
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 87:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:598
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:602
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:608
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:612
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 91:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:616
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 92:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:620
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].expr}
		}
	case 93:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:625
		{
			yyVAL.bytes = nil
		}
	case 94:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:629
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 95:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:633
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 96:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:639
		{
			yyVAL.str = AST_JOIN
		}
	case 97:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:643
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 98:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:647
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:651
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 100:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:655
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:659
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 102:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:663
		{
			yyVAL.str = AST_JOIN
		}
	case 103:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:667
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 104:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:671
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 105:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:677
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 106:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:681
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 107:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:685
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 108:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:689
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 109:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:693
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 110:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:697
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 111:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:701
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 112:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:707
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 113:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:711
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 114:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:716
		{
			yyVAL.indexHints = nil
		}
	case 115:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:720
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 116:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:724
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 117:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:728
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 118:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:734
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 119:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:738
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 120:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:743
		{
			yyVAL.expr = nil
		}
	case 121:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:747
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 122:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:752
		{
			yyVAL.expr = nil
		}
	case 123:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:756
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 124:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:760
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 125:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:765
		{
			yyVAL.expr = nil
		}
	case 126:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:769
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 127:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:775
		{
			yyVAL.str = AST_ALL
		}
	case 128:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:779
		{
			yyVAL.str = AST_SOME
		}
	case 129:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:783
		{
			yyVAL.str = AST_ANY
		}
	case 130:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:789
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 131:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:793
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 132:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:799
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 133:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:803
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 134:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:809
		{
			yyVAL.tuple = ValTuple(yyDollar[2].exprs)
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:813
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 136:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:819
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 137:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:825
		{
			yyVAL.exprs = Exprs{yyDollar[1].expr}
		}
	case 138:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:829
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:835
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 140:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:839
		{
			yyVAL.expr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 141:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:845
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:849
		{
			yyVAL.expr = yyDollar[1].colName
		}
	case 143:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:853
		{
			yyVAL.expr = yyDollar[1].tuple
		}
	case 144:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:857
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITAND, Right: yyDollar[3].expr}
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:861
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITOR, Right: yyDollar[3].expr}
		}
	case 146:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:865
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITXOR, Right: yyDollar[3].expr}
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:869
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_PLUS, Right: yyDollar[3].expr}
		}
	case 148:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:873
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MINUS, Right: yyDollar[3].expr}
		}
	case 149:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:877
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MULT, Right: yyDollar[3].expr}
		}
	case 150:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:881
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_DIV, Right: yyDollar[3].expr}
		}
	case 151:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:885
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_IDIV, Right: yyDollar[3].expr}
		}
	case 152:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:889
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MOD, Right: yyDollar[3].expr}
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:893
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, Right: yyDollar[3].expr}
		}
	case 154:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:897
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:901
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, Right: yyDollar[3].expr}
		}
	case 156:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:905
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:909
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, Right: yyDollar[3].expr}
		}
	case 158:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:913
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 159:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:917
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, Right: yyDollar[3].expr}
		}
	case 160:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:921
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:925
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, Right: yyDollar[3].expr}
		}
	case 162:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:929
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 163:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:933
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, Right: yyDollar[3].expr}
		}
	case 164:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:937
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 165:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:941
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, Right: yyDollar[3].expr}
		}
	case 166:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:945
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 167:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:949
		{
			yyVAL.expr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 168:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:953
		{
			yyVAL.expr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 169:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:957
		{
			yyVAL.expr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 170:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:961
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
	case 171:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:976
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 172:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:980
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 173:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:984
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LIKE, Right: yyDollar[3].expr}
		}
	case 174:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:988
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_LIKE, Right: yyDollar[4].expr}
		}
	case 175:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:992
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 176:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:996
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 177:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1000
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: yyDollar[3].expr}
		}
	case 178:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1004
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: yyDollar[4].expr}
		}
	case 179:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1008
		{
			yyVAL.expr = &RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[3].expr}
		}
	case 180:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1012
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[4].expr}}
		}
	case 181:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1016
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 182:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1020
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_BETWEEN, From: yyDollar[3].expr, To: yyDollar[5].expr}
		}
	case 183:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1024
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].expr, To: yyDollar[6].expr}
		}
	case 184:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1028
		{
			yyVAL.expr = yyDollar[1].caseExpr
		}
	case 185:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1032
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 186:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1036
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 187:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1040
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 188:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1044
		{
			yyVAL.expr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1048
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 190:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1052
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 191:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1056
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 192:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1060
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 193:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1064
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 194:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1068
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 195:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1072
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 196:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1076
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 197:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1080
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 198:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:1084
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 199:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1088
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date"), Exprs: SelectExprs{yyDollar[3].selectExpr}}
		}
	case 200:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1092
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 201:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1096
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 202:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1100
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 203:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1104
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 204:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1108
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 205:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1112
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 206:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1116
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 207:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1120
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 208:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1124
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 209:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1128
		{
			yyVAL.expr = &FuncExpr{Name: []byte("str_to_date"), Exprs: yyDollar[3].selectExprs}
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1134
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1138
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1142
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1146
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1150
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1154
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1158
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1162
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1168
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1172
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1176
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1180
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1184
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1188
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1192
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1196
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1200
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1206
		{
			yyVAL.bytes = SECOND_MICROSECOND_BYTES
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1210
		{
			yyVAL.bytes = MINUTE_MICROSECOND_BYTES
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1214
		{
			yyVAL.bytes = MINUTE_SECOND_BYTES
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1218
		{
			yyVAL.bytes = HOUR_MICROSECOND_BYTES
		}
	case 231:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1222
		{
			yyVAL.bytes = HOUR_SECOND_BYTES
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1226
		{
			yyVAL.bytes = HOUR_MINUTE_BYTES
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1230
		{
			yyVAL.bytes = DAY_MICROSECOND_BYTES
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1234
		{
			yyVAL.bytes = DAY_SECOND_BYTES
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1238
		{
			yyVAL.bytes = DAY_MINUTE_BYTES
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1242
		{
			yyVAL.bytes = DAY_HOUR_BYTES
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1246
		{
			yyVAL.bytes = YEAR_MONTH_BYTES
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1252
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1256
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1260
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 241:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1264
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1268
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 243:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1272
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1276
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1280
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 246:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1284
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1288
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 248:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1292
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 249:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1299
		{
			if bytes.Equal(bytes.ToLower(yyDollar[1].bytes), []byte("datetime")) {
				yyVAL.bytes = DATETIME_BYTES
			} else {
				yylex.Error("expecting datetime")
				return 1
			}
		}
	case 250:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1310
		{
			yyVAL.bytes = IF_BYTES
		}
	case 251:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1314
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1318
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 253:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1322
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 254:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1326
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 255:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1330
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 256:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1336
		{
			yyVAL.byt = AST_UPLUS
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1340
		{
			yyVAL.byt = AST_UMINUS
		}
	case 258:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1344
		{
			yyVAL.byt = AST_TILDA
		}
	case 259:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1350
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].expr, Whens: yyDollar[3].whens, Else: yyDollar[4].expr}
		}
	case 260:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1355
		{
			yyVAL.expr = nil
		}
	case 261:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1359
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 262:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1365
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 263:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1369
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 264:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1375
		{
			yyVAL.when = &When{Cond: yyDollar[2].expr, Val: yyDollar[4].expr}
		}
	case 265:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1380
		{
			yyVAL.expr = nil
		}
	case 266:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1384
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1390
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 268:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1394
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 269:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1400
		{
			yyVAL.expr = StrVal(yyDollar[1].bytes)
		}
	case 270:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1404
		{
			yyVAL.expr = NumVal(yyDollar[1].bytes)
		}
	case 271:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1408
		{
			yyVAL.expr = ValArg(yyDollar[1].bytes)
		}
	case 272:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1412
		{
			yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 273:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1416
		{
			yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 274:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1420
		{
			yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 275:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1424
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
	case 276:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1437
		{
			yyVAL.expr = &NullVal{}
		}
	case 277:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1441
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 278:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1447
		{
			yyVAL.expr = &TrueVal{}
		}
	case 279:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1451
		{
			yyVAL.expr = &FalseVal{}
		}
	case 280:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1455
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 281:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1460
		{
			yyVAL.exprs = nil
		}
	case 282:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1464
		{
			yyVAL.exprs = yyDollar[3].exprs
		}
	case 283:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1469
		{
			yyVAL.expr = nil
		}
	case 284:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1473
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 285:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1478
		{
			yyVAL.orderBy = nil
		}
	case 286:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1482
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 287:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1488
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 288:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1492
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 289:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1498
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 290:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1503
		{
			yyVAL.str = AST_ASC
		}
	case 291:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1507
		{
			yyVAL.str = AST_ASC
		}
	case 292:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1511
		{
			yyVAL.str = AST_DESC
		}
	case 293:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1516
		{
			yyVAL.limit = nil
		}
	case 294:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1520
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].expr}
		}
	case 295:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1524
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].expr, Rowcount: yyDollar[4].expr}
		}
	case 296:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1528
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].expr, Rowcount: yyDollar[2].expr}
		}
	case 297:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1533
		{
			yyVAL.str = ""
		}
	case 298:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1537
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 299:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1541
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
	case 300:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1554
		{
			yyVAL.columns = nil
		}
	case 301:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1558
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 302:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1564
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 303:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1568
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 304:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1573
		{
			yyVAL.updateExprs = nil
		}
	case 305:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1577
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 306:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1583
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 307:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1587
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 308:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1593
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].expr}
		}
	case 309:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1598
		{
			yyVAL.empty = struct{}{}
		}
	case 310:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1600
		{
			yyVAL.empty = struct{}{}
		}
	case 311:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1603
		{
			yyVAL.empty = struct{}{}
		}
	case 312:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1605
		{
			yyVAL.empty = struct{}{}
		}
	case 313:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1608
		{
			yyVAL.empty = struct{}{}
		}
	case 314:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1610
		{
			yyVAL.empty = struct{}{}
		}
	case 315:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1614
		{
			yyVAL.empty = struct{}{}
		}
	case 316:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1616
		{
			yyVAL.empty = struct{}{}
		}
	case 317:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1618
		{
			yyVAL.empty = struct{}{}
		}
	case 318:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1620
		{
			yyVAL.empty = struct{}{}
		}
	case 319:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1622
		{
			yyVAL.empty = struct{}{}
		}
	case 320:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1625
		{
			yyVAL.empty = struct{}{}
		}
	case 321:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1627
		{
			yyVAL.empty = struct{}{}
		}
	case 322:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1630
		{
			yyVAL.empty = struct{}{}
		}
	case 323:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1632
		{
			yyVAL.empty = struct{}{}
		}
	case 324:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1635
		{
			yyVAL.empty = struct{}{}
		}
	case 325:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1637
		{
			yyVAL.empty = struct{}{}
		}
	case 326:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1641
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 327:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1646
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
