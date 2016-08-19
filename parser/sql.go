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
const KILL = 57512
const CONNECTION = 57513
const QUERY = 57514
const SESSION = 57515
const GLOBAL = 57516
const CREATE = 57517
const ALTER = 57518
const DROP = 57519
const RENAME = 57520
const TABLE = 57521
const INDEX = 57522
const VIEW = 57523
const TO = 57524
const IGNORE = 57525
const IF = 57526
const UNIQUE = 57527
const USING = 57528

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
const yyInitialStackSize = 16

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 23,
	150, 52,
	-2, 72,
	-1, 33,
	169, 48,
	-2, 52,
}

const yyNprod = 333
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1617

var yyAct = [...]int{

	72, 217, 66, 554, 289, 599, 199, 135, 504, 437,
	67, 104, 419, 304, 350, 496, 430, 228, 200, 3,
	326, 457, 205, 345, 209, 91, 94, 137, 49, 361,
	51, 149, 60, 449, 52, 124, 234, 540, 542, 142,
	215, 132, 54, 145, 55, 237, 139, 572, 138, 57,
	58, 59, 144, 571, 126, 146, 570, 143, 56, 150,
	47, 48, 212, 177, 208, 92, 214, 508, 509, 507,
	178, 179, 355, 422, 367, 118, 120, 121, 421, 123,
	443, 140, 545, 442, 47, 48, 444, 356, 357, 229,
	474, 230, 327, 327, 407, 201, 365, 180, 495, 368,
	203, 19, 20, 21, 22, 19, 344, 222, 541, 165,
	166, 167, 168, 153, 154, 155, 156, 157, 158, 161,
	159, 160, 211, 156, 157, 158, 161, 159, 160, 612,
	198, 23, 253, 158, 161, 159, 160, 422, 125, 206,
	275, 226, 421, 567, 233, 273, 274, 272, 232, 497,
	569, 413, 241, 336, 242, 243, 244, 245, 246, 247,
	248, 249, 250, 251, 252, 257, 259, 261, 263, 265,
	267, 269, 270, 220, 445, 276, 223, 280, 281, 414,
	240, 337, 271, 236, 224, 105, 106, 107, 136, 300,
	301, 299, 568, 277, 538, 497, 305, 288, 298, 321,
	279, 425, 534, 303, 610, 424, 532, 535, 329, 330,
	547, 533, 333, 537, 546, 201, 536, 224, 342, 609,
	338, 323, 139, 577, 138, 139, 607, 138, 353, 348,
	302, 331, 332, 317, 318, 549, 334, 429, 364, 366,
	363, 351, 516, 28, 29, 31, 328, 401, 400, 283,
	285, 286, 352, 544, 207, 339, 393, 32, 34, 33,
	340, 319, 515, 514, 608, 425, 30, 513, 358, 424,
	392, 24, 25, 27, 26, 380, 381, 382, 371, 608,
	372, 278, 351, 379, 324, 373, 608, 374, 391, 375,
	390, 376, 354, 377, 399, 378, 258, 260, 262, 264,
	266, 268, 340, 398, 384, 153, 154, 155, 156, 157,
	158, 161, 159, 160, 484, 485, 486, 487, 488, 501,
	489, 490, 340, 340, 452, 404, 417, 340, 402, 408,
	484, 485, 486, 487, 488, 482, 489, 490, 415, 416,
	36, 37, 38, 39, 428, 321, 406, 139, 139, 138,
	435, 387, 409, 439, 433, 397, 210, 411, 36, 37,
	38, 39, 446, 436, 432, 389, 423, 388, 134, 386,
	322, 473, 440, 519, 518, 472, 224, 148, 403, 502,
	586, 426, 521, 585, 340, 584, 323, 451, 340, 583,
	582, 581, 557, 524, 523, 410, 522, 517, 396, 447,
	346, 520, 347, 593, 190, 592, 591, 347, 189, 475,
	93, 139, 181, 138, 213, 479, 480, 468, 478, 194,
	193, 192, 191, 188, 187, 340, 477, 340, 432, 340,
	323, 186, 185, 481, 494, 184, 183, 151, 469, 470,
	471, 182, 499, 219, 493, 320, 503, 196, 423, 500,
	195, 125, 512, 92, 543, 511, 492, 453, 454, 455,
	456, 510, 370, 369, 349, 133, 238, 172, 169, 170,
	152, 176, 163, 162, 164, 174, 173, 175, 527, 165,
	166, 167, 168, 153, 154, 155, 156, 157, 158, 161,
	159, 160, 235, 231, 225, 530, 531, 605, 197, 147,
	221, 40, 46, 139, 573, 550, 204, 552, 555, 423,
	423, 467, 551, 525, 526, 548, 19, 606, 131, 359,
	239, 459, 42, 43, 44, 45, 427, 105, 106, 107,
	560, 563, 129, 505, 117, 383, 119, 558, 561, 559,
	562, 431, 127, 566, 506, 438, 565, 556, 529, 351,
	611, 594, 458, 460, 461, 462, 463, 464, 465, 466,
	574, 19, 41, 61, 122, 412, 335, 588, 201, 590,
	18, 14, 17, 589, 587, 16, 15, 595, 596, 555,
	13, 597, 12, 360, 50, 448, 362, 53, 141, 441,
	227, 434, 600, 600, 600, 139, 598, 138, 601, 602,
	604, 578, 603, 105, 106, 107, 553, 284, 564, 613,
	70, 90, 528, 614, 100, 615, 405, 202, 325, 71,
	69, 218, 84, 85, 86, 73, 93, 282, 89, 498,
	97, 79, 65, 87, 88, 74, 75, 76, 108, 109,
	110, 111, 112, 113, 114, 115, 116, 80, 81, 82,
	539, 83, 420, 483, 418, 62, 491, 341, 128, 35,
	77, 78, 130, 172, 169, 170, 152, 176, 163, 162,
	164, 174, 173, 175, 171, 165, 166, 167, 168, 153,
	154, 155, 156, 157, 158, 161, 159, 160, 11, 10,
	102, 101, 476, 9, 8, 7, 6, 5, 4, 68,
	2, 1, 255, 256, 105, 106, 107, 254, 0, 0,
	0, 70, 90, 0, 0, 100, 0, 0, 95, 96,
	216, 103, 92, 84, 85, 86, 98, 93, 0, 89,
	0, 97, 79, 0, 87, 88, 74, 75, 76, 108,
	109, 110, 111, 112, 113, 114, 115, 116, 80, 81,
	82, 0, 83, 0, 0, 0, 0, 0, 0, 0,
	0, 77, 78, 0, 0, 0, 0, 0, 99, 152,
	176, 163, 162, 164, 174, 173, 175, 171, 165, 166,
	167, 168, 153, 154, 155, 156, 157, 158, 161, 159,
	160, 102, 101, 0, 0, 0, 0, 0, 0, 0,
	68, 0, 0, 0, 0, 105, 106, 107, 0, 0,
	0, 0, 70, 90, 0, 0, 100, 0, 0, 95,
	96, 0, 103, 218, 84, 85, 86, 98, 93, 287,
	89, 0, 97, 79, 0, 87, 88, 74, 75, 76,
	108, 109, 110, 111, 112, 113, 114, 115, 116, 80,
	81, 82, 0, 83, 0, 0, 0, 0, 0, 0,
	0, 0, 77, 78, 0, 0, 0, 0, 0, 99,
	176, 163, 162, 164, 174, 173, 175, 171, 165, 166,
	167, 168, 153, 154, 155, 156, 157, 158, 161, 159,
	160, 0, 102, 101, 0, 0, 0, 0, 0, 0,
	0, 68, 0, 0, 0, 0, 105, 106, 107, 0,
	0, 0, 0, 70, 90, 0, 0, 100, 0, 0,
	95, 96, 216, 103, 92, 84, 85, 86, 98, 93,
	0, 89, 0, 97, 79, 0, 87, 88, 74, 75,
	76, 108, 109, 110, 111, 112, 113, 114, 115, 116,
	80, 81, 82, 0, 83, 0, 0, 0, 0, 0,
	0, 0, 0, 77, 78, 0, 0, 0, 0, 0,
	99, 108, 109, 110, 111, 112, 113, 114, 115, 116,
	0, 0, 0, 0, 0, 290, 291, 292, 293, 294,
	295, 296, 297, 102, 101, 0, 0, 0, 0, 0,
	0, 0, 68, 0, 0, 0, 0, 105, 106, 107,
	0, 0, 0, 0, 70, 90, 0, 0, 100, 0,
	0, 95, 96, 0, 103, 218, 84, 85, 86, 98,
	93, 0, 89, 0, 97, 79, 0, 87, 88, 74,
	75, 76, 108, 109, 110, 111, 112, 113, 114, 115,
	116, 80, 81, 82, 0, 83, 0, 0, 63, 64,
	0, 0, 0, 0, 77, 78, 0, 0, 0, 0,
	0, 99, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 102, 101, 0, 0, 0, 0,
	0, 0, 0, 68, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 579, 580, 0, 0, 0,
	0, 0, 95, 96, 216, 103, 0, 0, 0, 19,
	98, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 105, 106, 107, 0, 0, 0, 0,
	70, 90, 0, 0, 100, 0, 0, 0, 0, 0,
	0, 92, 84, 85, 86, 0, 93, 0, 89, 0,
	97, 79, 99, 87, 88, 74, 75, 76, 108, 109,
	110, 111, 112, 113, 114, 115, 116, 80, 81, 82,
	0, 83, 0, 0, 0, 0, 0, 0, 0, 0,
	77, 78, 172, 169, 170, 152, 176, 163, 162, 164,
	174, 173, 175, 171, 165, 166, 167, 168, 153, 154,
	155, 156, 157, 158, 161, 159, 160, 0, 0, 0,
	102, 101, 0, 0, 0, 0, 0, 0, 0, 68,
	0, 0, 0, 0, 105, 106, 107, 0, 0, 0,
	0, 70, 90, 0, 0, 100, 0, 0, 95, 96,
	0, 103, 92, 84, 85, 86, 98, 93, 576, 89,
	0, 97, 79, 0, 87, 88, 74, 75, 76, 108,
	109, 110, 111, 112, 113, 114, 115, 116, 80, 81,
	82, 0, 83, 0, 0, 0, 0, 0, 0, 0,
	0, 77, 78, 0, 0, 0, 0, 0, 99, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 343, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 102, 101, 125, 0, 0, 0, 0, 0, 0,
	68, 0, 0, 0, 0, 0, 395, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 575, 0, 0, 95,
	96, 0, 103, 0, 0, 0, 0, 98, 172, 169,
	170, 152, 176, 163, 162, 164, 174, 173, 175, 171,
	165, 166, 167, 168, 153, 154, 155, 156, 157, 158,
	161, 159, 160, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 99,
	0, 172, 169, 170, 152, 176, 163, 162, 164, 174,
	173, 175, 171, 165, 166, 167, 168, 153, 154, 155,
	156, 157, 158, 161, 159, 160, 172, 169, 170, 152,
	176, 163, 162, 164, 174, 173, 175, 171, 165, 166,
	167, 168, 153, 154, 155, 156, 157, 158, 161, 159,
	160, 394, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 172, 169, 170, 152, 176, 163, 162,
	164, 174, 173, 175, 171, 165, 166, 167, 168, 153,
	154, 155, 156, 157, 158, 161, 159, 160, 172, 169,
	170, 152, 176, 163, 162, 164, 174, 173, 175, 171,
	165, 166, 167, 168, 153, 154, 155, 156, 157, 158,
	161, 159, 160, 172, 169, 170, 450, 176, 163, 162,
	164, 174, 173, 175, 171, 165, 166, 167, 168, 153,
	154, 155, 156, 157, 158, 161, 159, 160, 172, 169,
	170, 385, 176, 163, 162, 164, 174, 173, 175, 171,
	165, 166, 167, 168, 153, 154, 155, 156, 157, 158,
	161, 159, 160, 108, 109, 110, 111, 112, 113, 114,
	115, 116, 0, 0, 0, 0, 0, 290, 291, 292,
	293, 294, 295, 296, 297, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 306, 307, 308, 309,
	310, 311, 312, 313, 314, 315, 316,
}
var yyPact = [...]int{

	96, -1000, -1000, 259, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -113, -151, -139, -121, -130, -1000, -1000,
	887, -1000, -1000, -89, 414, 556, 520, -1000, -1000, -1000,
	509, -1000, 487, 428, 270, 28, -69, -1000, -1000, -145,
	-123, 414, -1000, -136, 414, -1000, 462, -153, 414, -153,
	1383, 1225, -1000, -1000, -1000, -1000, -1000, -1000, 1225, 1225,
	370, -1000, 399, 394, 393, 390, 389, 382, 381, 366,
	380, 379, 378, 377, -1000, -1000, -1000, 412, 409, 461,
	-1000, -1000, -10, 1124, -1000, -1000, -1000, -1000, 1225, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 475, 130, -103,
	258, 414, -107, -1000, 372, -1000, -1000, -1000, 988, -1000,
	402, 428, 465, -33, 428, 114, 457, -1000, 20, -1000,
	-62, 456, 33, 414, -1000, 455, -1000, -137, 429, 494,
	66, 414, 1225, 1225, 1225, 1225, 1225, 1225, 1225, 1225,
	1225, 1225, 685, 685, 685, 685, 685, 685, 685, 1225,
	1225, 368, 21, 1225, 166, 1225, 1225, 1383, 1383, -1000,
	-1000, 556, 584, 988, 786, 917, 917, 1225, 1225, 988,
	-1000, 1519, 988, 988, 988, -1000, -1000, 407, 414, 327,
	241, 1383, -50, 1383, 428, -1000, 1225, 1225, 130, 130,
	1225, 258, 55, 1225, 157, -1000, -1000, 1296, -34, -1000,
	365, 416, 427, 540, 416, -1000, 1225, 189, -1000, -80,
	-67, -1000, 493, -157, -1000, 62, -1000, 426, -1000, -1000,
	425, -1000, 1383, -11, -11, -11, -3, -3, -1000, -1000,
	-1000, -1000, -18, 370, -1000, -1000, -1000, -18, 370, -18,
	370, 174, 370, 174, 370, 174, 370, 174, 370, 651,
	651, -1000, 368, 1225, 1225, 1225, -18, -1000, 508, -1000,
	-18, 1433, -1000, 326, 988, 324, 322, -1000, 187, 185,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 167, 153,
	1358, 1321, 355, 257, 205, 196, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 145, 144, 285,
	333, -1000, -1000, 1225, -1000, -51, -1000, 1225, 360, 1383,
	1383, -1000, -1000, 1383, 130, 53, 1225, 1225, 283, 36,
	988, 502, -1000, 414, 101, 511, 416, 416, 273, -1000,
	533, 1225, -1000, 1383, -62, -73, -1000, -1000, -1000, -1000,
	60, 414, -1000, -149, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-18, -18, 1408, -1000, -1000, 1225, -1000, 281, -1000, -1000,
	988, 988, 988, 988, 474, 474, -1000, 988, 988, 988,
	309, 305, -1000, -1000, 1383, -56, -1000, 1225, 548, 511,
	416, -1000, -1000, 1225, 1225, 1383, 352, -1000, 232, 226,
	419, 100, -42, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	35, 368, 259, 81, 276, -1000, 533, 519, 531, 1383,
	-1000, -1000, -84, -90, -1000, 424, -1000, -1000, 418, -1000,
	1225, 751, -1000, 224, 220, 219, 199, 354, -1000, -1000,
	288, 287, -1000, -1000, -1000, -1000, -1000, -1000, 358, 353,
	351, 350, 988, 988, -1000, 1383, 1225, -1000, 114, 1383,
	1383, 538, 36, 36, -1000, -1000, 102, 98, 112, 109,
	90, -75, -1000, 417, 210, 45, -1000, 483, 132, -1000,
	-1000, -1000, 416, 519, -1000, 1225, 1225, -1000, -1000, -1000,
	-1000, -1000, 751, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, 349, -1000, -1000, -1000, 1519, 1519, 1383, 535, 530,
	226, 29, -1000, 88, -1000, 46, -1000, -1000, -1000, -1000,
	-124, -127, -133, -1000, -1000, -1000, -1000, -1000, 471, 368,
	-1000, -1000, 1253, 120, -1000, 1087, -1000, -1000, 348, 347,
	346, 342, 340, 337, 533, 1225, 1225, 1225, -1000, -1000,
	364, 363, 361, 544, -1000, 1225, 1225, 1225, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 519, 1383, 118,
	1383, 414, 414, 414, 416, 1383, 1383, -1000, 481, 183,
	-1000, 176, 161, 114, -1000, 543, 3, -1000, 414, -1000,
	-1000, -1000, 414, -1000, 414, -1000,
}
var yyPgo = [...]int{

	0, 701, 700, 18, 698, 697, 696, 695, 694, 693,
	689, 688, 501, 662, 659, 132, 658, 66, 40, 657,
	656, 1, 655, 654, 12, 653, 652, 41, 650, 5,
	14, 16, 632, 10, 25, 6, 629, 625, 11, 4,
	13, 21, 26, 620, 2, 619, 618, 20, 617, 616,
	612, 608, 9, 606, 3, 601, 8, 600, 23, 591,
	15, 7, 27, 590, 17, 589, 377, 588, 587, 586,
	585, 584, 583, 0, 36, 582, 580, 576, 575, 572,
	571, 570, 24, 22, 566, 565, 564, 502, 563, 562,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 3, 3,
	3, 4, 4, 78, 78, 5, 6, 7, 7, 7,
	63, 63, 64, 64, 64, 65, 65, 65, 65, 75,
	76, 77, 81, 84, 84, 85, 85, 85, 86, 86,
	88, 88, 87, 87, 87, 79, 79, 79, 79, 79,
	80, 80, 8, 8, 8, 9, 9, 9, 10, 11,
	11, 11, 89, 12, 13, 13, 14, 14, 14, 14,
	14, 16, 16, 17, 17, 18, 18, 18, 18, 19,
	19, 19, 23, 23, 24, 24, 24, 24, 20, 20,
	20, 25, 25, 25, 25, 25, 25, 25, 25, 25,
	26, 26, 26, 26, 26, 26, 26, 27, 27, 28,
	28, 28, 28, 29, 29, 30, 30, 83, 83, 83,
	82, 82, 15, 15, 15, 31, 31, 36, 36, 33,
	33, 42, 35, 35, 21, 21, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 39, 39, 39, 39, 39,
	39, 39, 39, 38, 38, 38, 38, 38, 38, 38,
	38, 38, 40, 40, 40, 40, 40, 40, 40, 40,
	40, 40, 40, 41, 41, 41, 41, 41, 41, 41,
	41, 41, 41, 41, 41, 37, 37, 37, 37, 37,
	37, 43, 43, 43, 45, 48, 48, 46, 46, 47,
	49, 49, 44, 44, 32, 32, 32, 32, 32, 32,
	32, 32, 32, 34, 34, 34, 50, 50, 51, 51,
	52, 52, 53, 53, 54, 55, 55, 55, 56, 56,
	56, 56, 57, 57, 57, 58, 58, 59, 59, 60,
	60, 61, 61, 62, 66, 66, 67, 67, 68, 68,
	69, 69, 69, 69, 69, 70, 70, 71, 71, 72,
	72, 73, 74,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 4, 12,
	3, 7, 7, 6, 6, 8, 7, 3, 4, 4,
	1, 3, 3, 2, 2, 2, 2, 2, 1, 1,
	1, 1, 5, 2, 2, 0, 2, 2, 0, 1,
	1, 1, 0, 1, 1, 3, 4, 4, 5, 5,
	2, 3, 5, 8, 4, 6, 7, 4, 5, 4,
	5, 5, 0, 2, 0, 2, 1, 2, 1, 1,
	1, 0, 1, 1, 3, 1, 2, 3, 3, 0,
	1, 2, 1, 3, 3, 3, 3, 5, 0, 1,
	2, 1, 1, 2, 3, 2, 3, 2, 2, 2,
	1, 3, 1, 1, 1, 3, 3, 1, 3, 0,
	5, 5, 5, 1, 3, 0, 2, 0, 2, 2,
	0, 2, 1, 1, 1, 2, 1, 1, 3, 3,
	1, 3, 1, 3, 1, 3, 1, 1, 1, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 4,
	3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
	3, 4, 3, 3, 2, 2, 3, 4, 3, 4,
	3, 4, 3, 4, 3, 4, 2, 5, 6, 1,
	3, 4, 5, 4, 1, 4, 3, 6, 6, 6,
	6, 6, 6, 7, 4, 6, 6, 6, 8, 8,
	8, 8, 8, 8, 4, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 2, 1, 2, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 5, 0, 1, 1, 2, 4,
	0, 2, 1, 3, 1, 1, 1, 2, 2, 2,
	4, 1, 1, 1, 1, 1, 0, 3, 0, 2,
	0, 3, 1, 3, 2, 0, 1, 1, 0, 2,
	4, 4, 0, 2, 4, 0, 3, 1, 3, 0,
	5, 1, 3, 3, 0, 2, 0, 3, 0, 1,
	1, 1, 1, 1, 1, 0, 1, 0, 1, 0,
	2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -75, -76, -80, -77, -78, -79, -81, 5,
	6, 7, 8, 35, 175, 176, 178, 177, 147, 148,
	170, 149, 161, 163, 162, -14, 99, 100, 101, 102,
	-12, -89, -12, -12, -12, -12, -87, 173, 174, 179,
	-71, 181, 185, -68, 181, 183, 179, 179, 180, 181,
	-21, -88, -22, 171, 172, -32, -44, -33, 115, -43,
	26, -45, -73, -37, 51, 52, 53, 76, 77, 47,
	63, 64, 65, 67, 38, 39, 40, 49, 50, 44,
	27, -34, 37, 42, -42, 134, 135, 46, 142, 184,
	30, 107, 106, 137, -38, 19, 20, 21, 54, 55,
	56, 57, 58, 59, 60, 61, 62, -12, 164, -87,
	165, 166, -86, 168, -73, 37, -3, 22, -16, 23,
	-13, 31, -27, 37, 98, -61, 160, -62, -44, -73,
	150, -67, 184, 180, -73, 179, -73, 37, -66, 184,
	-73, -66, 118, 131, 132, 133, 134, 135, 136, 138,
	139, 137, 121, 120, 122, 127, 128, 129, 130, 116,
	117, 126, 115, 124, 123, 125, 119, -21, -21, -21,
	-42, 42, 42, 42, 42, 42, 42, 42, 42, 42,
	38, 42, 42, 42, 42, 38, 38, 37, 140, -35,
	-3, -21, -48, -21, 31, -83, 9, 124, 167, -82,
	98, -73, 169, 42, -17, -18, 136, -21, 37, 41,
	-27, 35, 140, -27, 103, 37, 121, -63, -64, 151,
	153, 37, 115, -73, -74, 37, -74, 182, 37, 26,
	114, -73, -21, -21, -21, -21, -21, -21, -21, -21,
	-21, -21, -21, -15, 22, 17, 18, -21, -15, -21,
	-15, -21, -15, -21, -15, -21, -15, -21, -15, -21,
	-21, -33, 126, 124, 125, 119, -21, 27, 115, -34,
	-21, -21, 43, -17, 23, -17, -17, 43, -38, -39,
	68, 69, 70, 71, 72, 73, 74, 75, -38, -39,
	-21, -21, -18, -38, -40, -39, 87, 88, 89, 90,
	91, 92, 93, 94, 95, 96, 97, -18, -18, -17,
	38, -73, 43, 103, 43, -46, -47, 143, -27, -21,
	-21, -83, -83, -21, -82, -84, 98, 126, -35, 98,
	103, -19, -73, 25, 140, -58, 35, 42, -61, 37,
	-30, 9, -62, -21, 103, 152, 154, 155, -74, 26,
	-72, 186, -69, 178, 176, 34, 177, 12, 37, 37,
	37, -74, -42, -42, -42, -42, -42, -42, -42, -33,
	-21, -21, -21, 27, -34, 118, 43, -17, 43, 43,
	103, 103, 103, 103, 103, 25, 43, 98, 98, 98,
	103, 103, 43, 45, -21, -49, -47, 145, -21, -58,
	35, -83, -85, 98, 126, -21, -21, 43, -23, -24,
	-26, 42, 37, -42, 169, 165, -18, 24, -73, 136,
	-31, 30, -3, -61, -59, -44, -30, -52, 12, -21,
	-64, -65, 156, 153, 159, 114, -73, -74, -70, 182,
	118, -21, 43, -17, -17, -17, -17, -41, 78, 47,
	79, 80, 81, 82, 83, 84, 85, 37, -41, -18,
	-18, -18, 66, 66, 146, -21, 144, -31, -61, -21,
	-21, -30, 103, -25, 104, 105, 106, 107, 108, 110,
	111, -20, 37, 25, -24, 140, -60, 114, -36, -33,
	-60, 43, 103, -52, -56, 14, 13, 153, 157, 158,
	37, 37, -21, 43, 43, 43, 43, 43, 86, 86,
	43, 24, 43, 43, 43, -18, -18, -21, -50, 10,
	-24, -24, 104, 109, 104, 109, 104, 104, 104, -28,
	112, 183, 113, 37, 43, 37, 169, 165, 32, 103,
	-44, -56, -21, -53, -54, -21, -74, 43, -38, -40,
	-39, -38, -40, -39, -51, 11, 13, 114, 104, 104,
	180, 180, 180, 33, -33, 103, 15, 103, -55, 28,
	29, 43, 43, 43, 43, 43, 43, -52, -21, -35,
	-21, 42, 42, 42, 7, -21, -21, -54, -56, -29,
	-73, -29, -29, -61, -57, 16, 36, 43, 103, 43,
	43, 7, 126, -73, -73, -73,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 17, 72,
	72, 72, 72, -2, 327, 318, 0, 0, 39, 40,
	0, 41, 72, -2, 0, 0, 76, 78, 79, 80,
	81, 74, 0, 0, 0, 0, 0, 53, 54, 316,
	0, 0, 328, 0, 0, 319, 0, 314, 0, 314,
	60, 0, 144, 50, 51, 146, 147, 148, 0, 0,
	0, 189, 272, 0, 194, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 274, 275, 276, 0, 0, 0,
	281, 282, 331, 0, 140, 261, 262, 263, 265, 255,
	256, 257, 258, 259, 260, 283, 284, 285, 223, 224,
	225, 226, 227, 228, 229, 230, 231, 0, 127, 0,
	130, 0, 0, 49, 0, 331, 20, 77, 0, 82,
	73, 0, 0, 117, 0, 27, 0, 311, 0, 272,
	0, 0, 0, 0, 332, 0, 332, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 61, 174, 175,
	186, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	277, 0, 0, 0, 0, 278, 279, 0, 0, 0,
	0, 142, 0, 266, 0, 55, 0, 0, 127, 127,
	0, 130, 0, 0, 18, 83, 85, 89, 331, 75,
	305, 0, 0, 125, 0, 28, 0, 29, 30, 0,
	0, 332, 0, 329, 64, 0, 67, 0, 69, 315,
	0, 332, 145, 149, 150, 151, 152, 153, 154, 155,
	156, 157, 158, 0, 132, 133, 134, 160, 0, 162,
	0, 164, 0, 166, 0, 168, 0, 170, 0, 172,
	173, 176, 0, 0, 0, 0, 178, 180, 0, 182,
	184, 0, 190, 0, 0, 0, 0, 196, 0, 0,
	215, 216, 217, 218, 219, 220, 221, 222, 0, 0,
	0, 0, 0, 0, 0, 0, 232, 233, 234, 235,
	236, 237, 238, 239, 240, 241, 242, 0, 0, 0,
	0, 273, 139, 0, 141, 270, 267, 0, 305, 128,
	129, 56, 57, 131, 127, 45, 0, 0, 0, 0,
	0, 86, 90, 0, 0, 0, 0, 0, 125, 118,
	290, 0, 312, 313, 0, 0, 33, 34, 62, 317,
	0, 0, 332, 325, 320, 321, 322, 323, 324, 68,
	70, 71, 159, 161, 163, 165, 167, 169, 171, 177,
	179, 185, 0, 181, 183, 0, 191, 0, 193, 195,
	0, 0, 0, 0, 0, 0, 204, 0, 0, 0,
	0, 0, 214, 280, 143, 0, 268, 0, 0, 0,
	0, 58, 59, 0, 0, 43, 44, 42, 125, 92,
	98, 0, 110, 112, 113, 114, 84, 87, 91, 88,
	309, 0, 136, 309, 0, 307, 290, 298, 0, 126,
	31, 32, 0, 0, 38, 0, 330, 65, 0, 326,
	0, 187, 192, 0, 0, 0, 0, 0, 243, 244,
	245, 247, 249, 250, 251, 252, 253, 254, 0, 0,
	0, 0, 0, 0, 264, 271, 0, 23, 24, 46,
	47, 286, 0, 0, 101, 102, 0, 0, 0, 0,
	0, 119, 99, 0, 0, 0, 21, 0, 135, 137,
	22, 306, 0, 298, 26, 0, 0, 35, 36, 37,
	332, 66, 188, 197, 198, 199, 200, 201, 246, 248,
	202, 0, 205, 206, 207, 0, 0, 269, 288, 0,
	93, 96, 103, 0, 105, 0, 107, 108, 109, 94,
	0, 0, 0, 100, 95, 111, 115, 116, 0, 0,
	308, 25, 299, 291, 292, 295, 63, 203, 0, 0,
	0, 0, 0, 0, 290, 0, 0, 0, 104, 106,
	0, 0, 0, 0, 138, 0, 0, 0, 294, 296,
	297, 208, 209, 210, 211, 212, 213, 298, 289, 287,
	97, 0, 0, 0, 0, 300, 301, 293, 302, 0,
	123, 0, 0, 310, 19, 0, 0, 120, 0, 121,
	122, 303, 0, 124, 0, 304,
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
	182, 183, 184, 185, 186,
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
		//line sql.y:228
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:234
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 18:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:255
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 19:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:259
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].exprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 20:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:263
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 21:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:270
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 22:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:274
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
		//line sql.y:286
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 24:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:290
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
		//line sql.y:303
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 26:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:309
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 27:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:315
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 28:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:319
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name: []byte("names")}, Expr: StrVal(yyDollar[4].bytes)}}}
		}
	case 29:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:323
		{
			yyVAL.statement = &Set{Comments: append([][]byte{}, []byte(yyDollar[2].str), []byte("transaction"), yyDollar[4].bytes)}
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:329
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:333
		{
			yyVAL.bytes = append(yyDollar[1].bytes, append([]byte(", "), yyDollar[3].bytes...)...)
		}
	case 32:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:339
		{
			yyVAL.bytes = append([]byte("isolation level "), yyDollar[3].bytes...)
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:343
		{
			yyVAL.bytes = []byte("read write")
		}
	case 34:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:347
		{
			yyVAL.bytes = []byte("read only")
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:353
		{
			yyVAL.bytes = []byte("repeatable read")
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:357
		{
			yyVAL.bytes = []byte("read committed")
		}
	case 37:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:361
		{
			yyVAL.bytes = []byte("read uncommitted")
		}
	case 38:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:365
		{
			yyVAL.bytes = []byte("serializable")
		}
	case 39:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:371
		{
			yyVAL.statement = &Begin{}
		}
	case 40:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:377
		{
			yyVAL.statement = &Commit{}
		}
	case 41:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:383
		{
			yyVAL.statement = &Rollback{}
		}
	case 42:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:389
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].exprs}
		}
	case 43:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:395
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 44:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:399
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 45:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:404
		{
			yyVAL.expr = nil
		}
	case 46:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:408
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 47:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:412
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 48:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:417
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 49:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:421
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 50:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:427
		{
			yyVAL.str = AST_KILL_CONNECTION
		}
	case 51:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:431
		{
			yyVAL.str = AST_KILL_QUERY
		}
	case 52:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:436
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 53:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:440
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 54:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:444
		{
			yyVAL.str = AST_GLOBAL_SCOPE
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:451
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 56:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:455
		{
			yyVAL.statement = &Show{Section: "variables", Modifier: yyDollar[2].str, LikeOrWhere: yyDollar[4].expr}
		}
	case 57:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:459
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].expr, LikeOrWhere: yyDollar[4].expr}
		}
	case 58:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:463
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].expr, LikeOrWhere: yyDollar[5].expr}
		}
	case 59:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:467
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].expr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].expr}
		}
	case 60:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:473
		{
			yyVAL.statement = &Kill{Scope: AST_KILL_CONNECTION, ID: yyDollar[2].expr}
		}
	case 61:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:477
		{
			yyVAL.statement = &Kill{Scope: yyDollar[2].str, ID: yyDollar[3].expr}
		}
	case 62:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:483
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 63:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:487
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 64:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:492
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 65:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:498
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 66:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:502
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 67:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:507
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 68:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:513
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 69:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:519
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 70:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:523
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 71:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:528
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 72:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:533
		{
			SetAllowComments(yylex, true)
		}
	case 73:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:537
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 74:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:543
		{
			yyVAL.bytes2 = nil
		}
	case 75:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:547
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 76:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:553
		{
			yyVAL.str = AST_UNION
		}
	case 77:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:557
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 78:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:561
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 79:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:565
		{
			yyVAL.str = AST_EXCEPT
		}
	case 80:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:569
		{
			yyVAL.str = AST_INTERSECT
		}
	case 81:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:574
		{
			yyVAL.str = ""
		}
	case 82:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:578
		{
			yyVAL.str = AST_DISTINCT
		}
	case 83:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:584
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 84:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:588
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 85:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:594
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 86:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:598
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 87:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:602
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:606
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 89:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:611
		{
			yyVAL.bytes = nil
		}
	case 90:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:615
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 91:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:619
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 92:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:625
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 93:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:629
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 94:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:635
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:639
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:643
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 97:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:647
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].expr}
		}
	case 98:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:652
		{
			yyVAL.bytes = nil
		}
	case 99:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:656
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 100:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:660
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 101:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:666
		{
			yyVAL.str = AST_JOIN
		}
	case 102:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:670
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 103:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:674
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 104:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:678
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 105:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:682
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 106:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:686
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 107:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:690
		{
			yyVAL.str = AST_JOIN
		}
	case 108:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:694
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 109:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:698
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 110:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:704
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 111:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:708
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 112:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:712
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 113:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:716
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 114:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:720
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 115:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:724
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 116:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:728
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 117:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:734
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 118:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:738
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 119:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:743
		{
			yyVAL.indexHints = nil
		}
	case 120:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:747
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 121:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:751
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 122:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:755
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 123:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:761
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 124:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:765
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 125:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:770
		{
			yyVAL.expr = nil
		}
	case 126:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:774
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 127:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:779
		{
			yyVAL.expr = nil
		}
	case 128:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:783
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 129:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:787
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 130:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:792
		{
			yyVAL.expr = nil
		}
	case 131:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:796
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 132:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:802
		{
			yyVAL.str = AST_ALL
		}
	case 133:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:806
		{
			yyVAL.str = AST_SOME
		}
	case 134:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:810
		{
			yyVAL.str = AST_ANY
		}
	case 135:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:816
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 136:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:820
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 137:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:826
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 138:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:830
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 139:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:836
		{
			yyVAL.tuple = ValTuple(yyDollar[2].exprs)
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:840
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 141:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:846
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:852
		{
			yyVAL.exprs = Exprs{yyDollar[1].expr}
		}
	case 143:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:856
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 144:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:862
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:866
		{
			yyVAL.expr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:872
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 147:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:876
		{
			yyVAL.expr = yyDollar[1].colName
		}
	case 148:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:880
		{
			yyVAL.expr = yyDollar[1].tuple
		}
	case 149:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:884
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITAND, Right: yyDollar[3].expr}
		}
	case 150:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:888
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITOR, Right: yyDollar[3].expr}
		}
	case 151:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:892
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITXOR, Right: yyDollar[3].expr}
		}
	case 152:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:896
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_PLUS, Right: yyDollar[3].expr}
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:900
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MINUS, Right: yyDollar[3].expr}
		}
	case 154:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:904
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MULT, Right: yyDollar[3].expr}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:908
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_DIV, Right: yyDollar[3].expr}
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:912
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_IDIV, Right: yyDollar[3].expr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:916
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MOD, Right: yyDollar[3].expr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:920
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, Right: yyDollar[3].expr}
		}
	case 159:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:924
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:928
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, Right: yyDollar[3].expr}
		}
	case 161:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:932
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:936
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, Right: yyDollar[3].expr}
		}
	case 163:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:940
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:944
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, Right: yyDollar[3].expr}
		}
	case 165:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:948
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 166:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:952
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, Right: yyDollar[3].expr}
		}
	case 167:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:956
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 168:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:960
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, Right: yyDollar[3].expr}
		}
	case 169:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:964
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 170:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:968
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, Right: yyDollar[3].expr}
		}
	case 171:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:972
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 172:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:976
		{
			yyVAL.expr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 173:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:980
		{
			yyVAL.expr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 174:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:984
		{
			yyVAL.expr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 175:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:988
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
	case 176:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1003
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 177:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1007
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 178:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1011
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LIKE, Right: yyDollar[3].expr}
		}
	case 179:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1015
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_LIKE, Right: yyDollar[4].expr}
		}
	case 180:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1019
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 181:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 182:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1027
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: yyDollar[3].expr}
		}
	case 183:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1031
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: yyDollar[4].expr}
		}
	case 184:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1035
		{
			yyVAL.expr = &RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[3].expr}
		}
	case 185:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1039
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[4].expr}}
		}
	case 186:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1043
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 187:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1047
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_BETWEEN, From: yyDollar[3].expr, To: yyDollar[5].expr}
		}
	case 188:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1051
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].expr, To: yyDollar[6].expr}
		}
	case 189:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1055
		{
			yyVAL.expr = yyDollar[1].caseExpr
		}
	case 190:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1059
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 191:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1063
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 192:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1067
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 193:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1071
		{
			yyVAL.expr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 194:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1075
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 195:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1079
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 196:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1083
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 197:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1087
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 198:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1091
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 199:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1095
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 200:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1099
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 201:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1103
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 202:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1107
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 203:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:1111
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 204:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1115
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date"), Exprs: SelectExprs{yyDollar[3].selectExpr}}
		}
	case 205:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1119
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 206:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1123
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 207:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1127
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 208:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1131
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 209:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1135
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 210:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1139
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 211:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1143
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 212:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1147
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 213:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1151
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 214:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1155
		{
			yyVAL.expr = &FuncExpr{Name: []byte("str_to_date"), Exprs: yyDollar[3].selectExprs}
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1161
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1165
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1169
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1173
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1177
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1181
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1185
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1189
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1195
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1199
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1203
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1207
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1211
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1215
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1219
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1223
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 231:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1227
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1233
		{
			yyVAL.bytes = SECOND_MICROSECOND_BYTES
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1237
		{
			yyVAL.bytes = MINUTE_MICROSECOND_BYTES
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1241
		{
			yyVAL.bytes = MINUTE_SECOND_BYTES
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1245
		{
			yyVAL.bytes = HOUR_MICROSECOND_BYTES
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1249
		{
			yyVAL.bytes = HOUR_SECOND_BYTES
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1253
		{
			yyVAL.bytes = HOUR_MINUTE_BYTES
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1257
		{
			yyVAL.bytes = DAY_MICROSECOND_BYTES
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1261
		{
			yyVAL.bytes = DAY_SECOND_BYTES
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1265
		{
			yyVAL.bytes = DAY_MINUTE_BYTES
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1269
		{
			yyVAL.bytes = DAY_HOUR_BYTES
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1273
		{
			yyVAL.bytes = YEAR_MONTH_BYTES
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1279
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1283
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1287
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 246:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1291
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1295
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 248:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1299
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 249:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1303
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 250:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1307
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 251:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1311
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1315
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 253:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1319
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 254:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1326
		{
			if bytes.Equal(bytes.ToLower(yyDollar[1].bytes), []byte("datetime")) {
				yyVAL.bytes = DATETIME_BYTES
			} else {
				yylex.Error("expecting datetime")
				return 1
			}
		}
	case 255:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1337
		{
			yyVAL.bytes = IF_BYTES
		}
	case 256:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1341
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1345
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 258:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1349
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1353
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 260:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1357
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 261:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1363
		{
			yyVAL.byt = AST_UPLUS
		}
	case 262:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1367
		{
			yyVAL.byt = AST_UMINUS
		}
	case 263:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1371
		{
			yyVAL.byt = AST_TILDA
		}
	case 264:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1377
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].expr, Whens: yyDollar[3].whens, Else: yyDollar[4].expr}
		}
	case 265:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1382
		{
			yyVAL.expr = nil
		}
	case 266:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1386
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1392
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 268:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1396
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 269:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1402
		{
			yyVAL.when = &When{Cond: yyDollar[2].expr, Val: yyDollar[4].expr}
		}
	case 270:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1407
		{
			yyVAL.expr = nil
		}
	case 271:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1411
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 272:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1417
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 273:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1421
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 274:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1427
		{
			yyVAL.expr = StrVal(yyDollar[1].bytes)
		}
	case 275:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1431
		{
			yyVAL.expr = NumVal(yyDollar[1].bytes)
		}
	case 276:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1435
		{
			yyVAL.expr = ValArg(yyDollar[1].bytes)
		}
	case 277:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1439
		{
			yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 278:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1443
		{
			yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 279:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1447
		{
			yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 280:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1451
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
	case 281:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1464
		{
			yyVAL.expr = &NullVal{}
		}
	case 282:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1468
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1474
		{
			yyVAL.expr = &TrueVal{}
		}
	case 284:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1478
		{
			yyVAL.expr = &FalseVal{}
		}
	case 285:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1482
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 286:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1487
		{
			yyVAL.exprs = nil
		}
	case 287:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1491
		{
			yyVAL.exprs = yyDollar[3].exprs
		}
	case 288:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1496
		{
			yyVAL.expr = nil
		}
	case 289:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1500
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 290:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1505
		{
			yyVAL.orderBy = nil
		}
	case 291:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1509
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 292:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1515
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 293:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1519
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 294:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1525
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 295:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1530
		{
			yyVAL.str = AST_ASC
		}
	case 296:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1534
		{
			yyVAL.str = AST_ASC
		}
	case 297:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1538
		{
			yyVAL.str = AST_DESC
		}
	case 298:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1543
		{
			yyVAL.limit = nil
		}
	case 299:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1547
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].expr}
		}
	case 300:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1551
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].expr, Rowcount: yyDollar[4].expr}
		}
	case 301:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1555
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].expr, Rowcount: yyDollar[2].expr}
		}
	case 302:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1560
		{
			yyVAL.str = ""
		}
	case 303:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1564
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 304:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1568
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
	case 305:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1581
		{
			yyVAL.columns = nil
		}
	case 306:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1585
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 307:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1591
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 308:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1595
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 309:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1600
		{
			yyVAL.updateExprs = nil
		}
	case 310:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1604
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 311:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1610
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 312:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1614
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 313:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1620
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].expr}
		}
	case 314:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1625
		{
			yyVAL.empty = struct{}{}
		}
	case 315:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1627
		{
			yyVAL.empty = struct{}{}
		}
	case 316:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1630
		{
			yyVAL.empty = struct{}{}
		}
	case 317:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1632
		{
			yyVAL.empty = struct{}{}
		}
	case 318:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1635
		{
			yyVAL.empty = struct{}{}
		}
	case 319:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1637
		{
			yyVAL.empty = struct{}{}
		}
	case 320:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1641
		{
			yyVAL.empty = struct{}{}
		}
	case 321:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1643
		{
			yyVAL.empty = struct{}{}
		}
	case 322:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1645
		{
			yyVAL.empty = struct{}{}
		}
	case 323:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1647
		{
			yyVAL.empty = struct{}{}
		}
	case 324:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1649
		{
			yyVAL.empty = struct{}{}
		}
	case 325:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1652
		{
			yyVAL.empty = struct{}{}
		}
	case 326:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1654
		{
			yyVAL.empty = struct{}{}
		}
	case 327:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1657
		{
			yyVAL.empty = struct{}{}
		}
	case 328:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1659
		{
			yyVAL.empty = struct{}{}
		}
	case 329:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1662
		{
			yyVAL.empty = struct{}{}
		}
	case 330:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1664
		{
			yyVAL.empty = struct{}{}
		}
	case 331:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1668
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 332:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1673
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
