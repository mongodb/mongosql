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
const CHARACTER = 57503
const COLLATE = 57504
const REPLACE = 57505
const ADMIN = 57506
const SHOW = 57507
const DATABASES = 57508
const TABLES = 57509
const PROXY = 57510
const VARIABLES = 57511
const FULL = 57512
const COLUMNS = 57513
const KILL = 57514
const CONNECTION = 57515
const QUERY = 57516
const SESSION = 57517
const GLOBAL = 57518
const CREATE = 57519
const ALTER = 57520
const DROP = 57521
const RENAME = 57522
const TABLE = 57523
const INDEX = 57524
const VIEW = 57525
const TO = 57526
const IGNORE = 57527
const IF = 57528
const UNIQUE = 57529
const USING = 57530

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
	"CHARACTER",
	"COLLATE",
	"REPLACE",
	"ADMIN",
	"SHOW",
	"DATABASES",
	"TABLES",
	"PROXY",
	"VARIABLES",
	"FULL",
	"COLUMNS",
	"KILL",
	"CONNECTION",
	"QUERY",
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
const yyMaxDepth = 200

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 23,
	150, 54,
	-2, 74,
	-1, 33,
	171, 50,
	-2, 54,
}

const yyNprod = 335
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1682

var yyAct = [...]int{

	72, 218, 66, 559, 306, 604, 200, 441, 135, 67,
	104, 423, 509, 291, 434, 501, 352, 201, 3, 230,
	328, 462, 138, 347, 206, 210, 94, 236, 91, 365,
	150, 49, 60, 51, 143, 124, 454, 52, 216, 545,
	547, 54, 146, 55, 239, 577, 140, 576, 139, 132,
	575, 144, 145, 126, 56, 147, 57, 58, 59, 151,
	47, 48, 426, 178, 213, 550, 215, 425, 209, 355,
	179, 180, 513, 514, 371, 118, 120, 121, 512, 123,
	231, 19, 232, 448, 47, 48, 447, 359, 92, 449,
	360, 361, 141, 479, 329, 202, 369, 181, 255, 372,
	204, 154, 155, 156, 157, 158, 159, 162, 160, 161,
	617, 228, 546, 426, 329, 500, 411, 346, 425, 223,
	199, 125, 212, 153, 177, 164, 163, 165, 175, 174,
	176, 172, 166, 167, 168, 169, 154, 155, 156, 157,
	158, 159, 162, 160, 161, 235, 157, 158, 159, 162,
	160, 161, 417, 243, 207, 244, 245, 246, 247, 248,
	249, 250, 251, 252, 253, 254, 259, 261, 263, 265,
	267, 269, 271, 272, 234, 238, 278, 338, 282, 283,
	418, 221, 273, 225, 224, 159, 162, 160, 161, 572,
	302, 303, 429, 574, 502, 552, 428, 290, 300, 551,
	323, 301, 502, 305, 281, 339, 307, 450, 242, 331,
	332, 136, 137, 335, 573, 543, 202, 539, 542, 344,
	433, 340, 540, 140, 225, 139, 140, 541, 139, 304,
	357, 350, 319, 320, 333, 334, 615, 549, 336, 277,
	368, 370, 367, 429, 275, 276, 274, 428, 354, 341,
	285, 287, 288, 537, 342, 330, 325, 582, 538, 554,
	405, 362, 321, 260, 262, 264, 266, 268, 270, 208,
	614, 375, 404, 105, 106, 107, 326, 384, 385, 386,
	353, 279, 376, 612, 383, 521, 353, 377, 397, 378,
	396, 379, 395, 380, 394, 381, 613, 382, 489, 490,
	491, 492, 493, 520, 494, 495, 358, 403, 402, 388,
	519, 401, 19, 20, 21, 22, 489, 490, 491, 492,
	493, 211, 494, 495, 518, 506, 457, 408, 421, 406,
	613, 412, 36, 37, 38, 39, 36, 37, 38, 39,
	419, 420, 23, 613, 134, 342, 432, 323, 410, 140,
	140, 139, 439, 391, 413, 443, 393, 437, 392, 524,
	523, 415, 390, 342, 324, 436, 451, 440, 427, 280,
	342, 478, 477, 149, 487, 407, 591, 590, 445, 589,
	225, 430, 588, 587, 342, 507, 342, 586, 325, 342,
	526, 456, 562, 529, 452, 166, 167, 168, 169, 154,
	155, 156, 157, 158, 159, 162, 160, 161, 528, 525,
	527, 522, 400, 480, 598, 140, 342, 139, 342, 484,
	485, 473, 342, 483, 325, 414, 191, 348, 482, 597,
	190, 436, 349, 152, 349, 596, 93, 499, 182, 486,
	474, 475, 476, 214, 195, 504, 194, 193, 508, 472,
	192, 189, 427, 505, 28, 29, 31, 517, 188, 464,
	187, 458, 459, 460, 461, 186, 185, 184, 183, 220,
	32, 34, 33, 322, 197, 196, 498, 125, 610, 30,
	92, 548, 516, 532, 24, 25, 27, 26, 497, 515,
	463, 465, 466, 467, 468, 469, 470, 471, 611, 535,
	536, 444, 374, 373, 356, 351, 133, 240, 140, 237,
	555, 233, 557, 560, 427, 427, 530, 531, 226, 198,
	148, 556, 227, 222, 578, 40, 46, 553, 19, 205,
	131, 105, 106, 107, 363, 564, 567, 241, 431, 387,
	129, 563, 566, 561, 565, 568, 42, 43, 44, 45,
	127, 510, 571, 435, 511, 442, 570, 534, 117, 353,
	119, 616, 599, 19, 579, 41, 61, 122, 416, 337,
	18, 14, 593, 202, 595, 17, 16, 592, 594, 15,
	13, 12, 600, 601, 560, 364, 602, 50, 453, 366,
	53, 142, 446, 229, 438, 609, 583, 605, 605, 605,
	140, 558, 139, 606, 607, 603, 569, 533, 608, 409,
	203, 327, 71, 69, 618, 105, 106, 107, 619, 286,
	620, 73, 70, 90, 503, 65, 100, 544, 424, 488,
	422, 62, 496, 219, 84, 85, 86, 343, 93, 284,
	89, 128, 97, 79, 35, 87, 88, 74, 75, 76,
	108, 109, 110, 111, 112, 113, 114, 115, 116, 80,
	81, 82, 130, 83, 11, 10, 9, 8, 7, 6,
	5, 4, 77, 78, 2, 173, 170, 171, 153, 177,
	164, 163, 165, 175, 174, 176, 172, 166, 167, 168,
	169, 154, 155, 156, 157, 158, 159, 162, 160, 161,
	1, 0, 102, 101, 481, 0, 0, 0, 0, 0,
	0, 68, 0, 0, 257, 258, 105, 106, 107, 256,
	0, 0, 0, 70, 90, 0, 0, 100, 0, 0,
	95, 96, 217, 103, 92, 84, 85, 86, 98, 93,
	0, 89, 0, 97, 79, 0, 87, 88, 74, 75,
	76, 108, 109, 110, 111, 112, 113, 114, 115, 116,
	80, 81, 82, 0, 83, 0, 0, 0, 0, 0,
	0, 0, 0, 77, 78, 0, 0, 0, 0, 0,
	0, 0, 99, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 102, 101, 0, 0, 0, 0, 0,
	0, 0, 68, 0, 0, 0, 0, 105, 106, 107,
	0, 0, 0, 0, 70, 90, 0, 0, 100, 0,
	0, 95, 96, 0, 103, 219, 84, 85, 86, 98,
	93, 289, 89, 0, 97, 79, 0, 87, 88, 74,
	75, 76, 108, 109, 110, 111, 112, 113, 114, 115,
	116, 80, 81, 82, 0, 83, 0, 0, 0, 0,
	0, 0, 0, 0, 77, 78, 0, 0, 0, 0,
	0, 0, 0, 99, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 102, 101, 0, 0, 0, 0,
	0, 0, 0, 68, 0, 0, 0, 0, 105, 106,
	107, 0, 0, 0, 0, 70, 90, 0, 0, 100,
	0, 0, 95, 96, 217, 103, 92, 84, 85, 86,
	98, 93, 0, 89, 0, 97, 79, 0, 87, 88,
	74, 75, 76, 108, 109, 110, 111, 112, 113, 114,
	115, 116, 80, 81, 82, 0, 83, 0, 0, 0,
	0, 0, 0, 0, 0, 77, 78, 0, 0, 0,
	0, 0, 0, 0, 99, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 102, 101, 0, 0, 0,
	0, 0, 0, 0, 68, 0, 0, 0, 0, 105,
	106, 107, 0, 0, 0, 0, 70, 90, 0, 0,
	100, 0, 0, 95, 96, 0, 103, 219, 84, 85,
	86, 98, 93, 0, 89, 0, 97, 79, 0, 87,
	88, 74, 75, 76, 108, 109, 110, 111, 112, 113,
	114, 115, 116, 80, 81, 82, 0, 83, 0, 0,
	0, 0, 63, 64, 0, 0, 77, 78, 0, 0,
	0, 0, 0, 0, 0, 99, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 102, 101, 0, 0,
	0, 584, 585, 0, 0, 68, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 19, 0, 0, 0, 0,
	0, 0, 0, 0, 95, 96, 217, 103, 0, 105,
	106, 107, 98, 0, 0, 0, 70, 90, 0, 0,
	100, 0, 0, 0, 0, 0, 0, 92, 84, 85,
	86, 0, 93, 0, 89, 0, 97, 79, 0, 87,
	88, 74, 75, 76, 108, 109, 110, 111, 112, 113,
	114, 115, 116, 80, 81, 82, 99, 83, 0, 0,
	0, 0, 0, 0, 0, 0, 77, 78, 173, 170,
	171, 153, 177, 164, 163, 165, 175, 174, 176, 172,
	166, 167, 168, 169, 154, 155, 156, 157, 158, 159,
	162, 160, 161, 0, 0, 0, 102, 101, 0, 0,
	0, 0, 0, 0, 0, 68, 0, 0, 0, 0,
	105, 106, 107, 0, 0, 0, 0, 70, 90, 0,
	0, 100, 0, 0, 95, 96, 0, 103, 92, 84,
	85, 86, 98, 93, 581, 89, 0, 97, 79, 0,
	87, 88, 74, 75, 76, 108, 109, 110, 111, 112,
	113, 114, 115, 116, 80, 81, 82, 0, 83, 0,
	0, 0, 0, 0, 0, 0, 0, 77, 78, 0,
	0, 0, 0, 0, 0, 0, 99, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 345, 0,
	0, 0, 0, 0, 0, 0, 0, 102, 101, 0,
	125, 0, 0, 0, 0, 0, 68, 0, 0, 0,
	0, 0, 0, 399, 0, 0, 0, 0, 0, 0,
	0, 0, 580, 0, 0, 95, 96, 0, 103, 0,
	0, 0, 0, 98, 173, 170, 171, 153, 177, 164,
	163, 165, 175, 174, 176, 172, 166, 167, 168, 169,
	154, 155, 156, 157, 158, 159, 162, 160, 161, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 99, 173, 170,
	171, 153, 177, 164, 163, 165, 175, 174, 176, 172,
	166, 167, 168, 169, 154, 155, 156, 157, 158, 159,
	162, 160, 161, 173, 170, 171, 153, 177, 164, 163,
	165, 175, 174, 176, 172, 166, 167, 168, 169, 154,
	155, 156, 157, 158, 159, 162, 160, 161, 398, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	173, 170, 171, 153, 177, 164, 163, 165, 175, 174,
	176, 172, 166, 167, 168, 169, 154, 155, 156, 157,
	158, 159, 162, 160, 161, 173, 170, 171, 153, 177,
	164, 163, 165, 175, 174, 176, 172, 166, 167, 168,
	169, 154, 155, 156, 157, 158, 159, 162, 160, 161,
	173, 170, 171, 455, 177, 164, 163, 165, 175, 174,
	176, 172, 166, 167, 168, 169, 154, 155, 156, 157,
	158, 159, 162, 160, 161, 173, 170, 171, 389, 177,
	164, 163, 165, 175, 174, 176, 172, 166, 167, 168,
	169, 154, 155, 156, 157, 158, 159, 162, 160, 161,
	173, 170, 171, 153, 177, 164, 163, 165, 175, 174,
	176, 0, 166, 167, 168, 169, 154, 155, 156, 157,
	158, 159, 162, 160, 161, 177, 164, 163, 165, 175,
	174, 176, 172, 166, 167, 168, 169, 154, 155, 156,
	157, 158, 159, 162, 160, 161, 108, 109, 110, 111,
	112, 113, 114, 115, 116, 0, 0, 0, 0, 0,
	292, 293, 294, 295, 296, 297, 298, 299, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 308,
	309, 310, 311, 312, 313, 314, 315, 316, 317, 318,
	108, 109, 110, 111, 112, 113, 114, 115, 116, 0,
	0, 0, 0, 0, 292, 293, 294, 295, 296, 297,
	298, 299,
}
var yyPact = [...]int{

	307, -1000, -1000, 237, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -115, -150, -142, -127, -125, -1000, -1000,
	899, -1000, -1000, -91, 440, 558, 528, -1000, -1000, -1000,
	517, -1000, 499, 469, 246, 51, -58, -1000, -1000, -152,
	-131, 440, -1000, -139, 440, -1000, 483, -156, 440, -156,
	1380, 1221, -1000, -1000, -1000, -1000, -1000, -1000, 1221, 1221,
	396, -1000, 426, 425, 424, 423, 418, 416, 409, 388,
	408, 405, 404, 402, -1000, -1000, -1000, 437, 436, 482,
	-1000, -1000, -20, 1120, -1000, -1000, -1000, -1000, 1221, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 498, 145, -101,
	223, 440, -107, -1000, 401, -1000, -1000, -1000, 1000, -1000,
	428, 469, 488, -21, 469, 121, 481, 487, -1000, -10,
	-1000, -71, 474, 59, 440, -1000, 472, -1000, -140, 470,
	511, 94, 440, 1221, 1221, 1221, 1221, 1221, 1221, 1221,
	1221, 1221, 1221, 697, 697, 697, 697, 697, 697, 697,
	1221, 1221, 394, 120, 1221, 254, 1221, 1221, 1380, 1380,
	-1000, -1000, 558, 596, 1000, 798, 1606, 1606, 1221, 1221,
	1000, -1000, 1562, 1000, 1000, 1000, -1000, -1000, 435, 440,
	321, 233, 1380, -49, 1380, 469, -1000, 1221, 1221, 145,
	145, 1221, 223, 79, 1221, 151, -1000, -1000, 1293, -23,
	-1000, 392, 443, 468, 550, 443, -93, 467, 1221, 203,
	-1000, -65, -64, -1000, 508, -159, -1000, 62, -1000, 466,
	-1000, -1000, 465, -1000, 1380, 12, 12, 12, 49, 49,
	-1000, -1000, -1000, -1000, 268, 396, -1000, -1000, -1000, 268,
	396, 268, 396, -30, 396, -30, 396, -30, 396, -30,
	396, 5, 5, -1000, 394, 1221, 1221, 1221, 268, -1000,
	512, -1000, 268, 1430, -1000, 319, 1000, 315, 313, -1000,
	191, 189, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	187, 185, 1355, 1318, 369, 213, 210, 209, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 169,
	157, 286, 330, -1000, -1000, 1221, -1000, -29, -1000, 1221,
	390, 1380, 1380, -1000, -1000, 1380, 145, 54, 1221, 1221,
	285, 25, 1000, 514, -1000, 440, 84, 523, 443, 443,
	277, -1000, 543, 1221, -1000, 464, -1000, 1380, -71, -70,
	-1000, -1000, -1000, -1000, 93, 440, -1000, -148, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 268, 268, 1405, -1000, -1000, 1221,
	-1000, 283, -1000, -1000, 1000, 1000, 1000, 1000, 412, 412,
	-1000, 1000, 1000, 1000, 306, 305, -1000, -1000, 1380, -53,
	-1000, 1221, 560, 523, 443, -1000, -1000, 1221, 1221, 1380,
	1455, -1000, 271, 212, 451, 76, -25, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 88, 394, 237, 80, 282, -1000,
	543, 537, 541, 1380, -1000, -1000, -1000, -75, -85, -1000,
	452, -1000, -1000, 445, -1000, 1221, 1476, -1000, 281, 267,
	260, 242, 368, -1000, -1000, 274, 273, -1000, -1000, -1000,
	-1000, -1000, -1000, 366, 367, 365, 350, 1000, 1000, -1000,
	1380, 1221, -1000, 121, 1380, 1380, 547, 25, 25, -1000,
	-1000, 149, 113, 123, 114, 111, -73, -1000, 444, 194,
	28, -1000, 495, 156, -1000, -1000, -1000, 443, 537, -1000,
	1221, 1221, -1000, -1000, -1000, -1000, -1000, 1476, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 349, -1000, -1000, -1000,
	1562, 1562, 1380, 545, 539, 212, 75, -1000, 110, -1000,
	89, -1000, -1000, -1000, -1000, -132, -135, -137, -1000, -1000,
	-1000, -1000, -1000, 491, 394, -1000, -1000, 1249, 154, -1000,
	1083, -1000, -1000, 344, 340, 339, 336, 334, 333, 543,
	1221, 1221, 1221, -1000, -1000, 393, 387, 372, 555, -1000,
	1221, 1221, 1221, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 537, 1380, 153, 1380, 440, 440, 440, 443,
	1380, 1380, -1000, 462, 240, -1000, 227, 193, 121, -1000,
	554, -16, -1000, 440, -1000, -1000, -1000, 440, -1000, 440,
	-1000,
}
var yyPgo = [...]int{

	0, 700, 674, 17, 671, 670, 669, 668, 667, 666,
	665, 664, 525, 662, 644, 98, 641, 66, 38, 637,
	632, 1, 631, 630, 11, 629, 628, 49, 627, 5,
	16, 14, 625, 9, 28, 6, 624, 621, 10, 13,
	4, 21, 26, 613, 2, 612, 611, 20, 610, 609,
	607, 606, 7, 601, 3, 596, 12, 595, 23, 594,
	15, 8, 22, 593, 19, 592, 373, 591, 590, 589,
	588, 587, 585, 0, 27, 581, 580, 579, 576, 575,
	571, 570, 25, 24, 569, 568, 567, 526, 566, 565,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 3, 3,
	3, 4, 4, 78, 78, 5, 6, 7, 7, 7,
	7, 7, 63, 63, 64, 64, 64, 65, 65, 65,
	65, 75, 76, 77, 81, 84, 84, 85, 85, 85,
	86, 86, 88, 88, 87, 87, 87, 79, 79, 79,
	79, 79, 80, 80, 8, 8, 8, 9, 9, 9,
	10, 11, 11, 11, 89, 12, 13, 13, 14, 14,
	14, 14, 14, 16, 16, 17, 17, 18, 18, 18,
	18, 19, 19, 19, 23, 23, 24, 24, 24, 24,
	20, 20, 20, 25, 25, 25, 25, 25, 25, 25,
	25, 25, 26, 26, 26, 26, 26, 26, 26, 27,
	27, 28, 28, 28, 28, 29, 29, 30, 30, 83,
	83, 83, 82, 82, 15, 15, 15, 31, 31, 36,
	36, 33, 33, 42, 35, 35, 21, 21, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 39, 39, 39,
	39, 39, 39, 39, 39, 38, 38, 38, 38, 38,
	38, 38, 38, 38, 40, 40, 40, 40, 40, 40,
	40, 40, 40, 40, 40, 41, 41, 41, 41, 41,
	41, 41, 41, 41, 41, 41, 41, 37, 37, 37,
	37, 37, 37, 43, 43, 43, 45, 48, 48, 46,
	46, 47, 49, 49, 44, 44, 32, 32, 32, 32,
	32, 32, 32, 32, 32, 34, 34, 34, 50, 50,
	51, 51, 52, 52, 53, 53, 54, 55, 55, 55,
	56, 56, 56, 56, 57, 57, 57, 58, 58, 59,
	59, 60, 60, 61, 61, 62, 66, 66, 67, 67,
	68, 68, 69, 69, 69, 69, 69, 70, 70, 71,
	71, 72, 72, 73, 74,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 4, 12,
	3, 7, 7, 6, 6, 8, 7, 3, 4, 6,
	5, 4, 1, 3, 3, 2, 2, 2, 2, 2,
	1, 1, 1, 1, 5, 2, 2, 0, 2, 2,
	0, 1, 1, 1, 0, 1, 1, 3, 4, 4,
	5, 5, 2, 3, 5, 8, 4, 6, 7, 4,
	5, 4, 5, 5, 0, 2, 0, 2, 1, 2,
	1, 1, 1, 0, 1, 1, 3, 1, 2, 3,
	3, 0, 1, 2, 1, 3, 3, 3, 3, 5,
	0, 1, 2, 1, 1, 2, 3, 2, 3, 2,
	2, 2, 1, 3, 1, 1, 1, 3, 3, 1,
	3, 0, 5, 5, 5, 1, 3, 0, 2, 0,
	2, 2, 0, 2, 1, 1, 1, 2, 1, 1,
	3, 3, 1, 3, 1, 3, 1, 3, 1, 1,
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
	3, 4, 3, 4, 3, 3, 2, 2, 3, 4,
	3, 4, 3, 4, 3, 4, 3, 4, 2, 5,
	6, 1, 3, 4, 5, 4, 1, 4, 3, 6,
	6, 6, 6, 6, 6, 7, 4, 6, 6, 6,
	8, 8, 8, 8, 8, 8, 4, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 2, 1,
	2, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 5, 0, 1, 1,
	2, 4, 0, 2, 1, 3, 1, 1, 1, 2,
	2, 2, 4, 1, 1, 1, 1, 1, 0, 3,
	0, 2, 0, 3, 1, 3, 2, 0, 1, 1,
	0, 2, 4, 4, 0, 2, 4, 0, 3, 1,
	3, 0, 5, 1, 3, 3, 0, 2, 0, 3,
	0, 1, 1, 1, 1, 1, 1, 0, 1, 0,
	1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -75, -76, -80, -77, -78, -79, -81, 5,
	6, 7, 8, 35, 177, 178, 180, 179, 147, 148,
	172, 149, 163, 165, 164, -14, 99, 100, 101, 102,
	-12, -89, -12, -12, -12, -12, -87, 175, 176, 181,
	-71, 183, 187, -68, 183, 185, 181, 181, 182, 183,
	-21, -88, -22, 173, 174, -32, -44, -33, 115, -43,
	26, -45, -73, -37, 51, 52, 53, 76, 77, 47,
	63, 64, 65, 67, 38, 39, 40, 49, 50, 44,
	27, -34, 37, 42, -42, 134, 135, 46, 142, 186,
	30, 107, 106, 137, -38, 19, 20, 21, 54, 55,
	56, 57, 58, 59, 60, 61, 62, -12, 166, -87,
	167, 168, -86, 170, -73, 37, -3, 22, -16, 23,
	-13, 31, -27, 37, 98, -61, 160, 161, -62, -44,
	-73, 150, -67, 186, 182, -73, 181, -73, 37, -66,
	186, -73, -66, 118, 131, 132, 133, 134, 135, 136,
	138, 139, 137, 121, 120, 122, 127, 128, 129, 130,
	116, 117, 126, 115, 124, 123, 125, 119, -21, -21,
	-21, -42, 42, 42, 42, 42, 42, 42, 42, 42,
	42, 38, 42, 42, 42, 42, 38, 38, 37, 140,
	-35, -3, -21, -48, -21, 31, -83, 9, 124, 169,
	-82, 98, -73, 171, 42, -17, -18, 136, -21, 37,
	41, -27, 35, 140, -27, 103, 37, 35, 121, -63,
	-64, 151, 153, 37, 115, -73, -74, 37, -74, 184,
	37, 26, 114, -73, -21, -21, -21, -21, -21, -21,
	-21, -21, -21, -21, -21, -15, 22, 17, 18, -21,
	-15, -21, -15, -21, -15, -21, -15, -21, -15, -21,
	-15, -21, -21, -33, 126, 124, 125, 119, -21, 27,
	115, -34, -21, -21, 43, -17, 23, -17, -17, 43,
	-38, -39, 68, 69, 70, 71, 72, 73, 74, 75,
	-38, -39, -21, -21, -18, -38, -40, -39, 87, 88,
	89, 90, 91, 92, 93, 94, 95, 96, 97, -18,
	-18, -17, 38, -73, 43, 103, 43, -46, -47, 143,
	-27, -21, -21, -83, -83, -21, -82, -84, 98, 126,
	-35, 98, 103, -19, -73, 25, 140, -58, 35, 42,
	-61, 37, -30, 9, -62, 162, 37, -21, 103, 152,
	154, 155, -74, 26, -72, 188, -69, 180, 178, 34,
	179, 12, 37, 37, 37, -74, -42, -42, -42, -42,
	-42, -42, -42, -33, -21, -21, -21, 27, -34, 118,
	43, -17, 43, 43, 103, 103, 103, 103, 103, 25,
	43, 98, 98, 98, 103, 103, 43, 45, -21, -49,
	-47, 145, -21, -58, 35, -83, -85, 98, 126, -21,
	-21, 43, -23, -24, -26, 42, 37, -42, 171, 167,
	-18, 24, -73, 136, -31, 30, -3, -61, -59, -44,
	-30, -52, 12, -21, 37, -64, -65, 156, 153, 159,
	114, -73, -74, -70, 184, 118, -21, 43, -17, -17,
	-17, -17, -41, 78, 47, 79, 80, 81, 82, 83,
	84, 85, 37, -41, -18, -18, -18, 66, 66, 146,
	-21, 144, -31, -61, -21, -21, -30, 103, -25, 104,
	105, 106, 107, 108, 110, 111, -20, 37, 25, -24,
	140, -60, 114, -36, -33, -60, 43, 103, -52, -56,
	14, 13, 153, 157, 158, 37, 37, -21, 43, 43,
	43, 43, 43, 86, 86, 43, 24, 43, 43, 43,
	-18, -18, -21, -50, 10, -24, -24, 104, 109, 104,
	109, 104, 104, 104, -28, 112, 185, 113, 37, 43,
	37, 171, 167, 32, 103, -44, -56, -21, -53, -54,
	-21, -74, 43, -38, -40, -39, -38, -40, -39, -51,
	11, 13, 114, 104, 104, 182, 182, 182, 33, -33,
	103, 15, 103, -55, 28, 29, 43, 43, 43, 43,
	43, 43, -52, -21, -35, -21, 42, 42, 42, 7,
	-21, -21, -54, -56, -29, -73, -29, -29, -61, -57,
	16, 36, 43, 103, 43, 43, 7, 126, -73, -73,
	-73,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 17, 74,
	74, 74, 74, -2, 329, 320, 0, 0, 41, 42,
	0, 43, 74, -2, 0, 0, 78, 80, 81, 82,
	83, 76, 0, 0, 0, 0, 0, 55, 56, 318,
	0, 0, 330, 0, 0, 321, 0, 316, 0, 316,
	62, 0, 146, 52, 53, 148, 149, 150, 0, 0,
	0, 191, 274, 0, 196, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 276, 277, 278, 0, 0, 0,
	283, 284, 333, 0, 142, 263, 264, 265, 267, 257,
	258, 259, 260, 261, 262, 285, 286, 287, 225, 226,
	227, 228, 229, 230, 231, 232, 233, 0, 129, 0,
	132, 0, 0, 51, 0, 333, 20, 79, 0, 84,
	75, 0, 0, 119, 0, 27, 0, 0, 313, 0,
	274, 0, 0, 0, 0, 334, 0, 334, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 63, 176,
	177, 188, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 279, 0, 0, 0, 0, 280, 281, 0, 0,
	0, 0, 144, 0, 268, 0, 57, 0, 0, 129,
	129, 0, 132, 0, 0, 18, 85, 87, 91, 333,
	77, 307, 0, 0, 127, 0, 28, 0, 0, 31,
	32, 0, 0, 334, 0, 331, 66, 0, 69, 0,
	71, 317, 0, 334, 147, 151, 152, 153, 154, 155,
	156, 157, 158, 159, 160, 0, 134, 135, 136, 162,
	0, 164, 0, 166, 0, 168, 0, 170, 0, 172,
	0, 174, 175, 178, 0, 0, 0, 0, 180, 182,
	0, 184, 186, 0, 192, 0, 0, 0, 0, 198,
	0, 0, 217, 218, 219, 220, 221, 222, 223, 224,
	0, 0, 0, 0, 0, 0, 0, 0, 234, 235,
	236, 237, 238, 239, 240, 241, 242, 243, 244, 0,
	0, 0, 0, 275, 141, 0, 143, 272, 269, 0,
	307, 130, 131, 58, 59, 133, 129, 47, 0, 0,
	0, 0, 0, 88, 92, 0, 0, 0, 0, 0,
	127, 120, 292, 0, 314, 0, 30, 315, 0, 0,
	35, 36, 64, 319, 0, 0, 334, 327, 322, 323,
	324, 325, 326, 70, 72, 73, 161, 163, 165, 167,
	169, 171, 173, 179, 181, 187, 0, 183, 185, 0,
	193, 0, 195, 197, 0, 0, 0, 0, 0, 0,
	206, 0, 0, 0, 0, 0, 216, 282, 145, 0,
	270, 0, 0, 0, 0, 60, 61, 0, 0, 45,
	46, 44, 127, 94, 100, 0, 112, 114, 115, 116,
	86, 89, 93, 90, 311, 0, 138, 311, 0, 309,
	292, 300, 0, 128, 29, 33, 34, 0, 0, 40,
	0, 332, 67, 0, 328, 0, 189, 194, 0, 0,
	0, 0, 0, 245, 246, 247, 249, 251, 252, 253,
	254, 255, 256, 0, 0, 0, 0, 0, 0, 266,
	273, 0, 23, 24, 48, 49, 288, 0, 0, 103,
	104, 0, 0, 0, 0, 0, 121, 101, 0, 0,
	0, 21, 0, 137, 139, 22, 308, 0, 300, 26,
	0, 0, 37, 38, 39, 334, 68, 190, 199, 200,
	201, 202, 203, 248, 250, 204, 0, 207, 208, 209,
	0, 0, 271, 290, 0, 95, 98, 105, 0, 107,
	0, 109, 110, 111, 96, 0, 0, 0, 102, 97,
	113, 117, 118, 0, 0, 310, 25, 301, 293, 294,
	297, 65, 205, 0, 0, 0, 0, 0, 0, 292,
	0, 0, 0, 106, 108, 0, 0, 0, 0, 140,
	0, 0, 0, 296, 298, 299, 210, 211, 212, 213,
	214, 215, 300, 291, 289, 99, 0, 0, 0, 0,
	302, 303, 295, 304, 0, 125, 0, 0, 312, 19,
	0, 0, 122, 0, 123, 124, 305, 0, 126, 0,
	306,
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
	182, 183, 184, 185, 186, 187, 188,
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
		//line sql.y:230
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:236
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 18:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:257
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 19:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:261
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].exprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 20:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:265
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:272
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 22:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:276
		{
			cols := make(Columns, 0, len(yyDollar[6].updateExprs))
			vals := make(ValTuple, 0, len(yyDollar[6].updateExprs))
			for _, col := range yyDollar[6].updateExprs {
				cols = append(cols, &NonStarExpr{Expr: col.Name})
				vals = append(vals, col.Expr)
			}
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: cols, Rows: Values{vals}, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 23:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:288
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 24:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:292
		{
			cols := make(Columns, 0, len(yyDollar[6].updateExprs))
			vals := make(ValTuple, 0, len(yyDollar[6].updateExprs))
			for _, col := range yyDollar[6].updateExprs {
				cols = append(cols, &NonStarExpr{Expr: col.Name})
				vals = append(vals, col.Expr)
			}
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: cols, Rows: Values{vals}}
		}
	case 25:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:305
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 26:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:311
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 27:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:317
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 28:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:321
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_client")}, Expr: StrVal(yyDollar[4].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_results")}, Expr: StrVal(yyDollar[4].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_connection")}, Expr: StrVal(yyDollar[4].bytes)},
			}}
		}
	case 29:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:329
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_client")}, Expr: StrVal(yyDollar[4].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_results")}, Expr: StrVal(yyDollar[4].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_connection")}, Expr: StrVal(yyDollar[4].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@collation_connection")}, Expr: StrVal(yyDollar[6].bytes)},
			}}
		}
	case 30:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:338
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_client")}, Expr: StrVal(yyDollar[5].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_results")}, Expr: StrVal(yyDollar[5].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@collation_connection")}, Expr: &ColName{Name: []byte("@@collation_database")}},
			}}
		}
	case 31:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:346
		{
			yyVAL.statement = &Set{Comments: append([][]byte{}, []byte(yyDollar[2].str), []byte("transaction"), yyDollar[4].bytes)}
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:352
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:356
		{
			yyVAL.bytes = append(yyDollar[1].bytes, append([]byte(", "), yyDollar[3].bytes...)...)
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:362
		{
			yyVAL.bytes = append([]byte("isolation level "), yyDollar[3].bytes...)
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:366
		{
			yyVAL.bytes = []byte("read write")
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:370
		{
			yyVAL.bytes = []byte("read only")
		}
	case 37:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:376
		{
			yyVAL.bytes = []byte("repeatable read")
		}
	case 38:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:380
		{
			yyVAL.bytes = []byte("read committed")
		}
	case 39:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:384
		{
			yyVAL.bytes = []byte("read uncommitted")
		}
	case 40:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:388
		{
			yyVAL.bytes = []byte("serializable")
		}
	case 41:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:394
		{
			yyVAL.statement = &Begin{}
		}
	case 42:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:400
		{
			yyVAL.statement = &Commit{}
		}
	case 43:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:406
		{
			yyVAL.statement = &Rollback{}
		}
	case 44:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:412
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].exprs}
		}
	case 45:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:418
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 46:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:422
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 47:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:427
		{
			yyVAL.expr = nil
		}
	case 48:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:431
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 49:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:435
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 50:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:440
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 51:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:444
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 52:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:450
		{
			yyVAL.str = AST_KILL_CONNECTION
		}
	case 53:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:454
		{
			yyVAL.str = AST_KILL_QUERY
		}
	case 54:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:459
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 55:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:463
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 56:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:467
		{
			yyVAL.str = AST_GLOBAL_SCOPE
		}
	case 57:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:474
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 58:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:478
		{
			yyVAL.statement = &Show{Section: "variables", Modifier: yyDollar[2].str, LikeOrWhere: yyDollar[4].expr}
		}
	case 59:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:482
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].expr, LikeOrWhere: yyDollar[4].expr}
		}
	case 60:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:486
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].expr, LikeOrWhere: yyDollar[5].expr}
		}
	case 61:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:490
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].expr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].expr}
		}
	case 62:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:496
		{
			yyVAL.statement = &Kill{Scope: AST_KILL_CONNECTION, ID: yyDollar[2].expr}
		}
	case 63:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:500
		{
			yyVAL.statement = &Kill{Scope: yyDollar[2].str, ID: yyDollar[3].expr}
		}
	case 64:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:506
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 65:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:510
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 66:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:515
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 67:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:521
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 68:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:525
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 69:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:530
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 70:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:536
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 71:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:542
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 72:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:546
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 73:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:551
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 74:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:556
		{
			SetAllowComments(yylex, true)
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:560
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 76:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:566
		{
			yyVAL.bytes2 = nil
		}
	case 77:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:570
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 78:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:576
		{
			yyVAL.str = AST_UNION
		}
	case 79:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:580
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 80:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:584
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 81:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:588
		{
			yyVAL.str = AST_EXCEPT
		}
	case 82:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:592
		{
			yyVAL.str = AST_INTERSECT
		}
	case 83:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:597
		{
			yyVAL.str = ""
		}
	case 84:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:601
		{
			yyVAL.str = AST_DISTINCT
		}
	case 85:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:607
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 86:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:611
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 87:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:617
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 88:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:621
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:625
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:629
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 91:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:634
		{
			yyVAL.bytes = nil
		}
	case 92:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:638
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 93:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:642
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 94:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:648
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:652
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:658
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 97:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:662
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 98:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:666
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 99:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:670
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].expr}
		}
	case 100:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:675
		{
			yyVAL.bytes = nil
		}
	case 101:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:679
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 102:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:683
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 103:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:689
		{
			yyVAL.str = AST_JOIN
		}
	case 104:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:693
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 105:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:697
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 106:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:701
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 107:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:705
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 108:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:709
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 109:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:713
		{
			yyVAL.str = AST_JOIN
		}
	case 110:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:717
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 111:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:721
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 112:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:727
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 113:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:731
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 114:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:735
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 115:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:739
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 116:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:743
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 117:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:747
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 118:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:751
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 119:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:757
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 120:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:761
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 121:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:766
		{
			yyVAL.indexHints = nil
		}
	case 122:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:770
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 123:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:774
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 124:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:778
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 125:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:784
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 126:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:788
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 127:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:793
		{
			yyVAL.expr = nil
		}
	case 128:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:797
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 129:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:802
		{
			yyVAL.expr = nil
		}
	case 130:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:806
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 131:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:810
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 132:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:815
		{
			yyVAL.expr = nil
		}
	case 133:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:819
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 134:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:825
		{
			yyVAL.str = AST_ALL
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:829
		{
			yyVAL.str = AST_SOME
		}
	case 136:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:833
		{
			yyVAL.str = AST_ANY
		}
	case 137:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:839
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 138:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:843
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 139:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:849
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 140:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:853
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 141:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:859
		{
			yyVAL.tuple = ValTuple(yyDollar[2].exprs)
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:863
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 143:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:869
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 144:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:875
		{
			yyVAL.exprs = Exprs{yyDollar[1].expr}
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:879
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:885
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:889
		{
			yyVAL.expr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 148:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:895
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 149:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:899
		{
			yyVAL.expr = yyDollar[1].colName
		}
	case 150:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:903
		{
			yyVAL.expr = yyDollar[1].tuple
		}
	case 151:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:907
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITAND, Right: yyDollar[3].expr}
		}
	case 152:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:911
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITOR, Right: yyDollar[3].expr}
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:915
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITXOR, Right: yyDollar[3].expr}
		}
	case 154:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:919
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_PLUS, Right: yyDollar[3].expr}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:923
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MINUS, Right: yyDollar[3].expr}
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:927
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MULT, Right: yyDollar[3].expr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:931
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_DIV, Right: yyDollar[3].expr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:935
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_IDIV, Right: yyDollar[3].expr}
		}
	case 159:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:939
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MOD, Right: yyDollar[3].expr}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:943
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, Right: yyDollar[3].expr}
		}
	case 161:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:947
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:951
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, Right: yyDollar[3].expr}
		}
	case 163:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:955
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:959
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, Right: yyDollar[3].expr}
		}
	case 165:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:963
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 166:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:967
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, Right: yyDollar[3].expr}
		}
	case 167:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:971
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 168:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:975
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, Right: yyDollar[3].expr}
		}
	case 169:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:979
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 170:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:983
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, Right: yyDollar[3].expr}
		}
	case 171:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:987
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 172:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:991
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, Right: yyDollar[3].expr}
		}
	case 173:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:995
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 174:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:999
		{
			yyVAL.expr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 175:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1003
		{
			yyVAL.expr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 176:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1007
		{
			yyVAL.expr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 177:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1011
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
	case 178:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1026
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 179:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1030
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 180:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1034
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LIKE, Right: yyDollar[3].expr}
		}
	case 181:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1038
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_LIKE, Right: yyDollar[4].expr}
		}
	case 182:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1042
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 183:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1046
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 184:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1050
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: yyDollar[3].expr}
		}
	case 185:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1054
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: yyDollar[4].expr}
		}
	case 186:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1058
		{
			yyVAL.expr = &RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[3].expr}
		}
	case 187:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1062
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[4].expr}}
		}
	case 188:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1066
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 189:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1070
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_BETWEEN, From: yyDollar[3].expr, To: yyDollar[5].expr}
		}
	case 190:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1074
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].expr, To: yyDollar[6].expr}
		}
	case 191:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1078
		{
			yyVAL.expr = yyDollar[1].caseExpr
		}
	case 192:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1082
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 193:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1086
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 194:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1090
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 195:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1094
		{
			yyVAL.expr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 196:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1098
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 197:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1102
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 198:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1106
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 199:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1110
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 200:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1114
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 201:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1118
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 202:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1122
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 203:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1126
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 204:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1130
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 205:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:1134
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 206:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1138
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date"), Exprs: SelectExprs{yyDollar[3].selectExpr}}
		}
	case 207:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1142
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 208:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1146
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 209:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1150
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 210:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1154
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 211:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1158
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 212:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1162
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 213:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1166
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 214:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1170
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 215:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1174
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 216:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1178
		{
			yyVAL.expr = &FuncExpr{Name: []byte("str_to_date"), Exprs: yyDollar[3].selectExprs}
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1184
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1188
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1192
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1196
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1200
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1204
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1208
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1212
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1218
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1222
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1226
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1230
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1234
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1238
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 231:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1242
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1246
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1250
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1256
		{
			yyVAL.bytes = SECOND_MICROSECOND_BYTES
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1260
		{
			yyVAL.bytes = MINUTE_MICROSECOND_BYTES
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1264
		{
			yyVAL.bytes = MINUTE_SECOND_BYTES
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1268
		{
			yyVAL.bytes = HOUR_MICROSECOND_BYTES
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1272
		{
			yyVAL.bytes = HOUR_SECOND_BYTES
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1276
		{
			yyVAL.bytes = HOUR_MINUTE_BYTES
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1280
		{
			yyVAL.bytes = DAY_MICROSECOND_BYTES
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1284
		{
			yyVAL.bytes = DAY_SECOND_BYTES
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1288
		{
			yyVAL.bytes = DAY_MINUTE_BYTES
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1292
		{
			yyVAL.bytes = DAY_HOUR_BYTES
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1296
		{
			yyVAL.bytes = YEAR_MONTH_BYTES
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1302
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 246:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1306
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1310
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 248:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1314
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 249:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1318
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 250:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1322
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 251:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1326
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1330
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 253:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1334
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 254:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1338
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 255:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1342
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 256:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1349
		{
			if bytes.Equal(bytes.ToLower(yyDollar[1].bytes), []byte("datetime")) {
				yyVAL.bytes = DATETIME_BYTES
			} else {
				yylex.Error("expecting datetime")
				return 1
			}
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1360
		{
			yyVAL.bytes = IF_BYTES
		}
	case 258:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1364
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1368
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 260:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1372
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 261:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1376
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 262:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1380
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 263:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1386
		{
			yyVAL.byt = AST_UPLUS
		}
	case 264:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1390
		{
			yyVAL.byt = AST_UMINUS
		}
	case 265:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1394
		{
			yyVAL.byt = AST_TILDA
		}
	case 266:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1400
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].expr, Whens: yyDollar[3].whens, Else: yyDollar[4].expr}
		}
	case 267:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1405
		{
			yyVAL.expr = nil
		}
	case 268:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1409
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 269:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1415
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 270:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1419
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 271:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1425
		{
			yyVAL.when = &When{Cond: yyDollar[2].expr, Val: yyDollar[4].expr}
		}
	case 272:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1430
		{
			yyVAL.expr = nil
		}
	case 273:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1434
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 274:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1440
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 275:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1444
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 276:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1450
		{
			yyVAL.expr = StrVal(yyDollar[1].bytes)
		}
	case 277:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1454
		{
			yyVAL.expr = NumVal(yyDollar[1].bytes)
		}
	case 278:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1458
		{
			yyVAL.expr = ValArg(yyDollar[1].bytes)
		}
	case 279:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1462
		{
			yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 280:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1466
		{
			yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 281:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1470
		{
			yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 282:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1474
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
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1487
		{
			yyVAL.expr = &NullVal{}
		}
	case 284:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1491
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 285:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1497
		{
			yyVAL.expr = &TrueVal{}
		}
	case 286:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1501
		{
			yyVAL.expr = &FalseVal{}
		}
	case 287:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1505
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 288:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1510
		{
			yyVAL.exprs = nil
		}
	case 289:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1514
		{
			yyVAL.exprs = yyDollar[3].exprs
		}
	case 290:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1519
		{
			yyVAL.expr = nil
		}
	case 291:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1523
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 292:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1528
		{
			yyVAL.orderBy = nil
		}
	case 293:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1532
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 294:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1538
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 295:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1542
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 296:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1548
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 297:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1553
		{
			yyVAL.str = AST_ASC
		}
	case 298:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1557
		{
			yyVAL.str = AST_ASC
		}
	case 299:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1561
		{
			yyVAL.str = AST_DESC
		}
	case 300:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1566
		{
			yyVAL.limit = nil
		}
	case 301:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1570
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].expr}
		}
	case 302:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1574
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].expr, Rowcount: yyDollar[4].expr}
		}
	case 303:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1578
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].expr, Rowcount: yyDollar[2].expr}
		}
	case 304:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1583
		{
			yyVAL.str = ""
		}
	case 305:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1587
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 306:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1591
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
	case 307:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1604
		{
			yyVAL.columns = nil
		}
	case 308:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1608
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 309:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1614
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 310:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1618
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 311:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1623
		{
			yyVAL.updateExprs = nil
		}
	case 312:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1627
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 313:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1633
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 314:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1637
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 315:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1643
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].expr}
		}
	case 316:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1648
		{
			yyVAL.empty = struct{}{}
		}
	case 317:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1650
		{
			yyVAL.empty = struct{}{}
		}
	case 318:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1653
		{
			yyVAL.empty = struct{}{}
		}
	case 319:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1655
		{
			yyVAL.empty = struct{}{}
		}
	case 320:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1658
		{
			yyVAL.empty = struct{}{}
		}
	case 321:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1660
		{
			yyVAL.empty = struct{}{}
		}
	case 322:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1664
		{
			yyVAL.empty = struct{}{}
		}
	case 323:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1666
		{
			yyVAL.empty = struct{}{}
		}
	case 324:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1668
		{
			yyVAL.empty = struct{}{}
		}
	case 325:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1670
		{
			yyVAL.empty = struct{}{}
		}
	case 326:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1672
		{
			yyVAL.empty = struct{}{}
		}
	case 327:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1675
		{
			yyVAL.empty = struct{}{}
		}
	case 328:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1677
		{
			yyVAL.empty = struct{}{}
		}
	case 329:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1680
		{
			yyVAL.empty = struct{}{}
		}
	case 330:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1682
		{
			yyVAL.empty = struct{}{}
		}
	case 331:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1685
		{
			yyVAL.empty = struct{}{}
		}
	case 332:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1687
		{
			yyVAL.empty = struct{}{}
		}
	case 333:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1691
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 334:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1696
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
