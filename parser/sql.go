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
const COLLATION = 57514
const PROCESSLIST = 57515
const STATUS = 57516
const CHARSET = 57517
const EXPLAIN = 57518
const DESCRIBE = 57519
const EXTENDED = 57520
const PARTITIONS = 57521
const FORMAT = 57522
const TRADITIONAL = 57523
const JSON = 57524
const KILL = 57525
const CONNECTION = 57526
const QUERY = 57527
const SESSION = 57528
const GLOBAL = 57529
const CREATE = 57530
const ALTER = 57531
const DROP = 57532
const RENAME = 57533
const TABLE = 57534
const INDEX = 57535
const VIEW = 57536
const TO = 57537
const IGNORE = 57538
const IF = 57539
const UNIQUE = 57540
const USING = 57541

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
	"COLLATION",
	"PROCESSLIST",
	"STATUS",
	"CHARSET",
	"EXPLAIN",
	"DESCRIBE",
	"EXTENDED",
	"PARTITIONS",
	"FORMAT",
	"TRADITIONAL",
	"JSON",
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
	-1, 24,
	150, 55,
	-2, 95,
	-1, 34,
	171, 51,
	173, 51,
	-2, 55,
}

const yyNprod = 359
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1634

var yyAct = [...]int{

	77, 247, 71, 599, 150, 215, 483, 644, 549, 72,
	109, 465, 541, 320, 476, 335, 391, 216, 3, 259,
	504, 245, 386, 357, 153, 371, 99, 265, 96, 239,
	404, 165, 226, 65, 585, 587, 133, 137, 54, 244,
	56, 128, 147, 617, 57, 158, 123, 125, 126, 496,
	132, 155, 131, 154, 59, 129, 60, 160, 141, 268,
	162, 62, 63, 64, 166, 410, 52, 53, 193, 221,
	616, 615, 130, 375, 159, 194, 195, 161, 61, 52,
	53, 378, 379, 20, 284, 233, 97, 408, 224, 394,
	411, 134, 468, 225, 229, 490, 230, 467, 489, 398,
	217, 491, 196, 553, 554, 219, 399, 400, 260, 552,
	261, 156, 521, 358, 358, 468, 450, 590, 586, 540,
	467, 172, 173, 174, 177, 175, 176, 228, 174, 177,
	175, 176, 657, 385, 252, 214, 237, 181, 182, 183,
	184, 169, 170, 171, 172, 173, 174, 177, 175, 176,
	306, 257, 263, 243, 241, 304, 305, 303, 612, 134,
	264, 222, 542, 655, 492, 271, 614, 372, 272, 613,
	273, 274, 275, 276, 277, 278, 279, 280, 281, 282,
	283, 288, 290, 292, 294, 296, 298, 300, 301, 250,
	267, 307, 253, 311, 312, 373, 254, 302, 579, 232,
	368, 234, 654, 580, 583, 331, 332, 542, 582, 151,
	152, 652, 319, 329, 581, 352, 330, 561, 334, 310,
	392, 336, 471, 653, 360, 361, 470, 333, 369, 365,
	348, 349, 138, 139, 140, 254, 354, 217, 314, 316,
	317, 374, 407, 409, 406, 471, 355, 592, 383, 470,
	350, 591, 155, 622, 154, 155, 389, 154, 475, 396,
	241, 366, 653, 359, 289, 291, 293, 295, 297, 299,
	577, 653, 376, 380, 594, 578, 223, 381, 381, 393,
	169, 170, 171, 172, 173, 174, 177, 175, 176, 444,
	401, 110, 111, 112, 362, 363, 364, 443, 560, 308,
	414, 370, 41, 42, 43, 44, 423, 424, 425, 559,
	436, 415, 558, 422, 527, 392, 416, 442, 417, 435,
	418, 546, 419, 499, 420, 434, 421, 433, 397, 188,
	185, 186, 168, 192, 179, 178, 180, 190, 427, 191,
	187, 181, 182, 183, 184, 169, 170, 171, 172, 173,
	174, 177, 175, 176, 461, 430, 447, 589, 381, 441,
	451, 529, 530, 531, 532, 533, 445, 534, 535, 381,
	456, 457, 381, 440, 459, 460, 41, 42, 43, 44,
	449, 547, 452, 381, 432, 474, 352, 309, 155, 155,
	154, 481, 479, 455, 485, 431, 227, 429, 149, 564,
	353, 563, 520, 472, 478, 493, 482, 469, 519, 254,
	446, 631, 566, 164, 354, 630, 20, 487, 529, 530,
	531, 532, 533, 629, 534, 535, 381, 240, 249, 628,
	498, 565, 627, 494, 626, 602, 454, 569, 568, 567,
	562, 458, 463, 439, 381, 638, 20, 20, 21, 22,
	23, 637, 522, 242, 155, 381, 154, 381, 525, 515,
	354, 453, 516, 517, 518, 387, 636, 524, 388, 98,
	478, 39, 388, 500, 501, 502, 503, 24, 167, 539,
	197, 526, 206, 242, 462, 235, 205, 544, 210, 548,
	209, 208, 545, 207, 469, 204, 203, 202, 201, 557,
	188, 185, 186, 168, 192, 179, 178, 180, 190, 189,
	191, 187, 181, 182, 183, 184, 169, 170, 171, 172,
	173, 174, 177, 175, 176, 572, 200, 199, 198, 523,
	134, 238, 538, 351, 212, 211, 134, 97, 588, 575,
	576, 570, 571, 556, 537, 555, 486, 413, 155, 412,
	595, 395, 597, 600, 469, 469, 390, 596, 168, 192,
	179, 178, 180, 190, 189, 191, 187, 181, 182, 183,
	184, 169, 170, 171, 172, 173, 174, 177, 175, 176,
	650, 603, 606, 601, 605, 608, 604, 607, 148, 29,
	30, 32, 269, 266, 262, 255, 213, 163, 256, 51,
	651, 251, 231, 618, 619, 33, 35, 34, 593, 220,
	20, 146, 633, 217, 635, 402, 632, 634, 37, 38,
	624, 625, 640, 641, 600, 31, 642, 270, 473, 144,
	25, 26, 28, 27, 124, 477, 142, 645, 645, 645,
	155, 643, 154, 611, 648, 646, 647, 550, 110, 111,
	112, 551, 315, 484, 658, 75, 95, 610, 659, 105,
	660, 574, 110, 111, 112, 392, 248, 89, 90, 91,
	426, 98, 313, 94, 656, 102, 84, 639, 92, 93,
	79, 80, 81, 113, 114, 115, 116, 117, 118, 119,
	120, 121, 85, 86, 87, 20, 88, 46, 66, 377,
	136, 127, 367, 19, 18, 82, 83, 188, 185, 186,
	168, 192, 179, 178, 180, 190, 189, 191, 187, 181,
	182, 183, 184, 169, 170, 171, 172, 173, 174, 177,
	175, 176, 14, 17, 16, 107, 106, 15, 13, 12,
	36, 403, 55, 495, 73, 405, 58, 286, 287, 110,
	111, 112, 285, 157, 488, 258, 75, 95, 480, 649,
	105, 623, 598, 100, 101, 246, 108, 97, 89, 90,
	91, 103, 98, 609, 94, 573, 102, 84, 45, 92,
	93, 79, 80, 81, 113, 114, 115, 116, 117, 118,
	119, 120, 121, 85, 86, 87, 448, 88, 218, 356,
	47, 48, 49, 50, 76, 236, 82, 83, 74, 78,
	543, 70, 122, 113, 114, 115, 116, 117, 118, 119,
	120, 121, 584, 135, 466, 528, 104, 321, 322, 323,
	324, 325, 326, 327, 328, 464, 107, 106, 67, 536,
	382, 143, 40, 145, 11, 73, 10, 9, 8, 7,
	110, 111, 112, 6, 5, 4, 2, 75, 95, 1,
	0, 105, 0, 0, 100, 101, 0, 108, 248, 89,
	90, 91, 103, 98, 318, 94, 0, 102, 84, 0,
	92, 93, 79, 80, 81, 113, 114, 115, 116, 117,
	118, 119, 120, 121, 85, 86, 87, 0, 88, 0,
	0, 0, 0, 0, 0, 0, 0, 82, 83, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 104, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 107, 106, 0,
	0, 0, 0, 0, 0, 0, 73, 0, 0, 0,
	0, 110, 111, 112, 0, 0, 0, 0, 75, 95,
	0, 0, 105, 0, 0, 100, 101, 246, 108, 97,
	89, 90, 91, 103, 98, 0, 94, 0, 102, 84,
	0, 92, 93, 79, 80, 81, 113, 114, 115, 116,
	117, 118, 119, 120, 121, 85, 86, 87, 0, 88,
	0, 0, 0, 0, 0, 0, 0, 0, 82, 83,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 104, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 107, 106,
	0, 0, 0, 0, 0, 0, 0, 73, 0, 0,
	0, 0, 110, 111, 112, 0, 0, 0, 0, 75,
	95, 0, 0, 105, 0, 0, 100, 101, 0, 108,
	248, 89, 90, 91, 103, 98, 0, 94, 0, 102,
	84, 0, 92, 93, 79, 80, 81, 113, 114, 115,
	116, 117, 118, 119, 120, 121, 85, 86, 87, 0,
	88, 0, 0, 0, 0, 0, 0, 0, 514, 82,
	83, 0, 0, 0, 0, 0, 68, 69, 506, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 104,
	0, 0, 621, 0, 0, 0, 0, 0, 0, 107,
	106, 0, 0, 0, 0, 0, 0, 0, 73, 505,
	507, 508, 509, 510, 511, 512, 513, 20, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 100, 101, 246,
	108, 110, 111, 112, 0, 103, 0, 0, 75, 95,
	0, 0, 105, 0, 0, 0, 0, 0, 0, 97,
	89, 90, 91, 0, 98, 0, 94, 0, 102, 84,
	0, 92, 93, 79, 80, 81, 113, 114, 115, 116,
	117, 118, 119, 120, 121, 85, 86, 87, 0, 88,
	620, 0, 0, 0, 0, 0, 0, 0, 82, 83,
	104, 0, 188, 185, 186, 168, 192, 179, 178, 180,
	190, 189, 191, 187, 181, 182, 183, 184, 169, 170,
	171, 172, 173, 174, 177, 175, 176, 0, 107, 106,
	0, 0, 0, 0, 0, 0, 0, 73, 0, 0,
	0, 0, 110, 111, 112, 0, 0, 0, 0, 75,
	95, 0, 0, 105, 0, 0, 100, 101, 0, 108,
	97, 89, 90, 91, 103, 98, 0, 94, 0, 102,
	84, 0, 92, 93, 79, 80, 81, 113, 114, 115,
	116, 117, 118, 119, 120, 121, 85, 86, 87, 384,
	88, 0, 0, 0, 0, 0, 0, 0, 0, 82,
	83, 134, 0, 0, 0, 0, 113, 114, 115, 116,
	117, 118, 119, 120, 121, 0, 0, 0, 0, 104,
	321, 322, 323, 324, 325, 326, 327, 328, 0, 107,
	106, 438, 0, 0, 0, 0, 0, 0, 73, 337,
	338, 339, 340, 341, 342, 343, 344, 345, 346, 347,
	0, 0, 0, 0, 0, 0, 0, 100, 101, 0,
	108, 0, 0, 0, 0, 103, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 188,
	185, 186, 168, 192, 179, 178, 180, 190, 189, 191,
	187, 181, 182, 183, 184, 169, 170, 171, 172, 173,
	174, 177, 175, 176, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	104, 188, 185, 186, 168, 192, 179, 178, 180, 190,
	189, 191, 187, 181, 182, 183, 184, 169, 170, 171,
	172, 173, 174, 177, 175, 176, 437, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 188, 185,
	186, 168, 192, 179, 178, 180, 190, 189, 191, 187,
	181, 182, 183, 184, 169, 170, 171, 172, 173, 174,
	177, 175, 176, 188, 185, 186, 168, 192, 179, 178,
	180, 190, 189, 191, 187, 181, 182, 183, 184, 169,
	170, 171, 172, 173, 174, 177, 175, 176, 188, 185,
	186, 497, 192, 179, 178, 180, 190, 189, 191, 187,
	181, 182, 183, 184, 169, 170, 171, 172, 173, 174,
	177, 175, 176, 188, 185, 186, 428, 192, 179, 178,
	180, 190, 189, 191, 187, 181, 182, 183, 184, 169,
	170, 171, 172, 173, 174, 177, 175, 176, 188, 185,
	186, 168, 192, 179, 178, 180, 190, 189, 191, 0,
	181, 182, 183, 184, 169, 170, 171, 172, 173, 174,
	177, 175, 176, 192, 179, 178, 180, 190, 189, 191,
	187, 181, 182, 183, 184, 169, 170, 171, 172, 173,
	174, 177, 175, 176,
}
var yyPact = [...]int{

	442, -1000, -1000, 277, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -107, -154, -140, -114, -131, -1000,
	-1000, 932, -1000, -1000, -120, 499, 54, -1000, -1000, -1000,
	690, 614, -1000, -1000, -1000, 606, -1000, 580, 551, 300,
	49, -39, -1000, -1000, -152, -119, 499, -1000, -115, 499,
	-1000, 560, -166, 499, -166, 1398, 1253, -1000, -1000, -1000,
	-1000, -1000, -1000, 1253, 1253, 438, -1000, 486, 485, 484,
	456, 455, 454, 453, 444, 451, 449, 448, 446, -1000,
	-1000, -1000, 497, 496, 559, -1000, -1000, -5, 1152, -1000,
	-1000, -1000, -1000, 1253, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 578, 152, -81, 298, 499, -77, 567, 152,
	-89, 152, -1000, 443, -1000, 493, 411, -1000, -1000, -1000,
	32, -1000, -1000, 1033, -1000, 387, 551, 566, -6, 551,
	132, 558, 563, -1000, 30, -1000, -43, 557, 37, 499,
	-1000, 556, -1000, -136, 555, 601, 51, 499, 1253, 1253,
	1253, 1253, 1253, 1253, 1253, 1253, 1253, 1253, 730, 730,
	730, 730, 730, 730, 730, 1253, 1253, 427, 31, 1253,
	272, 1253, 1253, 1398, 1398, -1000, -1000, 690, 629, 1033,
	831, 759, 759, 1253, 1253, 1033, -1000, 1282, 1033, 1033,
	1033, -1000, -1000, 495, 499, 357, 203, 1398, -30, 1398,
	551, -1000, 1253, 1253, 152, 152, 152, 1253, 298, 102,
	-1000, 152, -1000, 69, -1000, 1253, -1000, -1000, -1000, -1000,
	-111, 277, 441, -100, 175, -1000, -1000, 1294, -7, -1000,
	430, 500, 519, 656, 500, -73, 514, 1253, 225, -1000,
	-53, -48, -1000, 589, -169, -1000, 53, -1000, 512, -1000,
	-1000, 510, -1000, 1398, -13, -13, -13, -8, -8, -1000,
	-1000, -1000, -1000, 10, 438, -1000, -1000, -1000, 10, 438,
	10, 438, 149, 438, 149, 438, 149, 438, 149, 438,
	440, 440, -1000, 427, 1253, 1253, 1253, 10, -1000, 643,
	-1000, 10, 1448, -1000, 354, 1033, 352, 341, -1000, 224,
	222, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 216,
	207, 1373, 1336, 400, 275, 261, 219, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 194, 186,
	323, 365, -1000, -1000, 1253, -1000, -29, -1000, 1253, 426,
	1398, 1398, -1000, -1000, -1000, 1398, 152, 69, 1253, 1253,
	-1000, 152, 1253, 1253, 311, 445, 399, -1000, -1000, -1000,
	55, 1033, 604, -1000, 499, 122, 605, 500, 500, 306,
	-1000, 641, 1253, -1000, 509, -1000, 1398, -43, -58, -1000,
	-1000, -1000, -1000, 50, 499, -1000, -146, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, 10, 10, 1423, -1000, -1000, 1253, -1000,
	280, -1000, -1000, 1033, 1033, 1033, 1033, 1071, 1071, -1000,
	1033, 1033, 1033, 342, 336, -1000, -1000, 1398, -34, -1000,
	1253, 385, 605, 500, -1000, -1000, 1398, 1473, -1000, 1398,
	214, -1000, -1000, -1000, 211, 257, 507, 78, -21, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 48, 427, 277, 93,
	278, -1000, 641, 633, 638, 1398, -1000, -1000, -1000, -44,
	-54, -1000, 508, -1000, -1000, 506, -1000, 1253, 1494, -1000,
	269, 266, 255, 174, 397, -1000, -1000, 315, 313, -1000,
	-1000, -1000, -1000, -1000, -1000, 388, 396, 395, 394, 1033,
	1033, -1000, 1398, 1253, -1000, 132, 651, 55, 55, -1000,
	-1000, 166, 94, 110, 104, 100, -78, -1000, 501, 314,
	80, -1000, 576, 171, -1000, -1000, -1000, 500, 633, -1000,
	1253, 1253, -1000, -1000, -1000, -1000, -1000, 1494, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 392, -1000, -1000, -1000,
	1282, 1282, 1398, 646, 630, 257, 44, -1000, 65, -1000,
	62, -1000, -1000, -1000, -1000, -122, -123, -150, -1000, -1000,
	-1000, -1000, -1000, 570, 427, -1000, -1000, 1117, 150, -1000,
	592, -1000, -1000, 391, 389, 386, 380, 372, 368, 641,
	1253, 1253, 1253, -1000, -1000, 424, 409, 403, 670, -1000,
	1253, 1253, 1253, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 633, 1398, 133, 1398, 499, 499, 499, 500,
	1398, 1398, -1000, 564, 168, -1000, 159, 120, 132, -1000,
	667, 6, -1000, 499, -1000, -1000, -1000, 499, -1000, 499,
	-1000,
}
var yyPgo = [...]int{

	0, 859, 856, 17, 855, 854, 853, 849, 848, 847,
	846, 844, 778, 843, 842, 84, 841, 39, 21, 840,
	839, 1, 838, 835, 11, 825, 824, 42, 823, 822,
	7, 16, 14, 811, 9, 28, 5, 810, 809, 10,
	13, 15, 20, 26, 808, 2, 805, 804, 799, 23,
	798, 796, 775, 773, 6, 762, 3, 761, 8, 759,
	22, 758, 12, 4, 24, 755, 19, 754, 413, 753,
	746, 745, 743, 742, 741, 0, 27, 740, 739, 738,
	737, 734, 733, 732, 704, 703, 29, 32, 69, 702,
	25, 701, 599, 700, 699, 698, 697,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 3,
	3, 3, 4, 4, 81, 81, 5, 6, 7, 7,
	7, 7, 7, 65, 65, 66, 66, 66, 67, 67,
	67, 67, 78, 79, 80, 84, 89, 89, 90, 90,
	90, 91, 91, 95, 95, 92, 92, 92, 82, 82,
	82, 82, 82, 82, 82, 82, 82, 82, 82, 94,
	94, 93, 93, 93, 86, 86, 77, 77, 77, 28,
	85, 85, 85, 83, 83, 8, 8, 8, 9, 9,
	9, 10, 11, 11, 11, 96, 12, 13, 13, 14,
	14, 14, 14, 14, 16, 16, 17, 17, 18, 18,
	18, 18, 19, 19, 19, 23, 23, 24, 24, 24,
	24, 20, 20, 20, 25, 25, 25, 25, 25, 25,
	25, 25, 25, 26, 26, 26, 26, 26, 26, 26,
	27, 27, 29, 29, 29, 29, 30, 30, 31, 31,
	88, 88, 88, 87, 87, 15, 15, 15, 32, 32,
	37, 37, 34, 34, 43, 36, 36, 21, 21, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 40, 40,
	40, 40, 40, 40, 40, 40, 39, 39, 39, 39,
	39, 39, 39, 39, 39, 41, 41, 41, 41, 41,
	41, 41, 41, 41, 41, 41, 42, 42, 42, 42,
	42, 42, 42, 42, 42, 42, 42, 42, 38, 38,
	38, 38, 38, 38, 44, 44, 44, 47, 50, 50,
	48, 48, 49, 51, 51, 45, 45, 46, 46, 46,
	33, 33, 33, 33, 33, 33, 33, 33, 33, 35,
	35, 35, 52, 52, 53, 53, 54, 54, 55, 55,
	56, 57, 57, 57, 58, 58, 58, 58, 59, 59,
	59, 60, 60, 61, 61, 62, 62, 63, 63, 64,
	68, 68, 69, 69, 70, 70, 71, 71, 71, 71,
	71, 72, 72, 73, 73, 74, 74, 75, 76,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 4,
	12, 3, 7, 7, 6, 6, 8, 7, 3, 4,
	6, 5, 4, 1, 3, 3, 2, 2, 2, 2,
	2, 1, 1, 1, 1, 5, 2, 2, 0, 2,
	2, 0, 1, 1, 1, 0, 1, 1, 3, 4,
	4, 5, 5, 3, 4, 4, 3, 5, 3, 1,
	1, 1, 1, 3, 1, 3, 1, 1, 1, 1,
	3, 3, 5, 2, 3, 5, 8, 4, 6, 7,
	4, 5, 4, 5, 5, 0, 2, 0, 2, 1,
	2, 1, 1, 1, 0, 1, 1, 3, 1, 2,
	3, 3, 0, 1, 2, 1, 3, 3, 3, 3,
	5, 0, 1, 2, 1, 1, 2, 3, 2, 3,
	2, 2, 2, 1, 3, 1, 1, 1, 3, 3,
	1, 3, 0, 5, 5, 5, 1, 3, 0, 2,
	0, 2, 2, 0, 2, 1, 1, 1, 2, 1,
	1, 3, 3, 1, 3, 1, 3, 1, 3, 1,
	1, 1, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 4, 3, 4, 3, 4, 3, 4, 3,
	4, 3, 4, 3, 4, 3, 3, 2, 2, 3,
	4, 3, 4, 3, 4, 3, 4, 3, 4, 2,
	5, 6, 1, 3, 4, 5, 4, 1, 4, 3,
	6, 6, 6, 6, 6, 6, 7, 4, 6, 6,
	6, 8, 8, 8, 8, 8, 8, 4, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 2,
	1, 2, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 5, 0, 1,
	1, 2, 4, 0, 2, 1, 3, 0, 1, 1,
	1, 1, 1, 2, 2, 2, 4, 1, 1, 1,
	1, 1, 0, 3, 0, 2, 0, 3, 1, 3,
	2, 0, 1, 1, 0, 2, 4, 4, 0, 2,
	4, 0, 3, 1, 3, 0, 5, 1, 3, 3,
	0, 2, 0, 3, 0, 1, 1, 1, 1, 1,
	1, 0, 1, 0, 1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -78, -79, -83, -80, -81, -82, -84, -85,
	5, 6, 7, 8, 35, 188, 189, 191, 190, 147,
	148, 183, 149, 163, 165, 164, -77, 176, 177, 29,
	-14, 99, 100, 101, 102, -12, -96, -12, -12, -12,
	-12, -92, 186, 187, 192, -73, 194, 198, -70, 194,
	196, 192, 192, 193, 194, -21, -95, -22, 184, 185,
	-33, -45, -34, 115, -44, 26, -47, -75, -38, 51,
	52, 53, 76, 77, 47, 63, 64, 65, 67, 38,
	39, 40, 49, 50, 44, 27, -35, 37, 42, -43,
	134, 135, 46, 142, 197, 30, 107, 106, 137, -39,
	19, 20, 21, 54, 55, 56, 57, 58, 59, 60,
	61, 62, -12, 166, -92, 167, 168, -91, 161, 175,
	192, 172, 170, -75, 37, -28, -93, -75, 178, 179,
	180, -3, 22, -16, 23, -13, 31, -27, 37, 98,
	-63, 160, 161, -64, -45, -75, 150, -69, 197, 193,
	-75, 192, -75, 37, -68, 197, -75, -68, 118, 131,
	132, 133, 134, 135, 136, 138, 139, 137, 121, 120,
	122, 127, 128, 129, 130, 116, 117, 126, 115, 124,
	123, 125, 119, -21, -21, -21, -43, 42, 42, 42,
	42, 42, 42, 42, 42, 42, 38, 42, 42, 42,
	42, 38, 38, 37, 140, -36, -3, -21, -50, -21,
	31, -88, 9, 124, 169, 174, -87, 98, -75, 171,
	173, 35, -88, 174, -88, 42, -46, -75, 38, -86,
	16, -3, 42, 121, -17, -18, 136, -21, 37, 41,
	-27, 35, 140, -27, 103, 37, 35, 121, -65, -66,
	151, 153, 37, 115, -75, -76, 37, -76, 195, 37,
	26, 114, -75, -21, -21, -21, -21, -21, -21, -21,
	-21, -21, -21, -21, -15, 22, 17, 18, -21, -15,
	-21, -15, -21, -15, -21, -15, -21, -15, -21, -15,
	-21, -21, -34, 126, 124, 125, 119, -21, 27, 115,
	-35, -21, -21, 43, -17, 23, -17, -17, 43, -39,
	-40, 68, 69, 70, 71, 72, 73, 74, 75, -39,
	-40, -21, -21, -18, -39, -41, -40, 87, 88, 89,
	90, 91, 92, 93, 94, 95, 96, 97, -18, -18,
	-17, 38, -75, 43, 103, 43, -48, -49, 143, -27,
	-21, -21, -88, -88, -88, -21, -87, -89, 98, 126,
	-88, -90, 98, 126, -36, 184, -86, -94, 181, 182,
	98, 103, -19, -75, 25, 140, -60, 35, 42, -63,
	37, -31, 9, -64, 162, 37, -21, 103, 152, 154,
	155, -76, 26, -74, 199, -71, 191, 189, 34, 190,
	12, 37, 37, 37, -76, -43, -43, -43, -43, -43,
	-43, -43, -34, -21, -21, -21, 27, -35, 118, 43,
	-17, 43, 43, 103, 103, 103, 103, 103, 25, 43,
	98, 98, 98, 103, 103, 43, 45, -21, -51, -49,
	145, -21, -60, 35, -88, -90, -21, -21, -88, -21,
	-21, 43, 39, 43, -23, -24, -26, 42, 37, -43,
	171, 167, -18, 24, -75, 136, -32, 30, -3, -63,
	-61, -45, -31, -54, 12, -21, 37, -66, -67, 156,
	153, 159, 114, -75, -76, -72, 195, 118, -21, 43,
	-17, -17, -17, -17, -42, 78, 47, 79, 80, 81,
	82, 83, 84, 85, 37, -42, -18, -18, -18, 66,
	66, 146, -21, 144, -32, -63, -31, 103, -25, 104,
	105, 106, 107, 108, 110, 111, -20, 37, 25, -24,
	140, -62, 114, -37, -34, -62, 43, 103, -54, -58,
	14, 13, 153, 157, 158, 37, 37, -21, 43, 43,
	43, 43, 43, 86, 86, 43, 24, 43, 43, 43,
	-18, -18, -21, -52, 10, -24, -24, 104, 109, 104,
	109, 104, 104, 104, -29, 112, 196, 113, 37, 43,
	37, 171, 167, 32, 103, -45, -58, -21, -55, -56,
	-21, -76, 43, -39, -41, -40, -39, -41, -40, -53,
	11, 13, 114, 104, 104, 193, 193, 193, 33, -34,
	103, 15, 103, -57, 28, 29, 43, 43, 43, 43,
	43, 43, -54, -21, -36, -21, 42, 42, 42, 7,
	-21, -21, -56, -58, -30, -75, -30, -30, -63, -59,
	16, 36, 43, 103, 43, 43, 7, 126, -75, -75,
	-75,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 17, 18,
	95, 95, 95, 95, -2, 353, 344, 0, 0, 42,
	43, 0, 44, 95, -2, 0, 0, 76, 77, 78,
	0, 99, 101, 102, 103, 104, 97, 0, 0, 0,
	0, 0, 56, 57, 342, 0, 0, 354, 0, 0,
	345, 0, 340, 0, 340, 83, 0, 167, 53, 54,
	169, 170, 171, 0, 0, 0, 212, 295, 0, 217,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 300,
	301, 302, 0, 0, 0, 307, 308, 357, 0, 163,
	284, 285, 286, 288, 278, 279, 280, 281, 282, 283,
	309, 310, 311, 246, 247, 248, 249, 250, 251, 252,
	253, 254, 0, 150, 0, 153, 0, 0, 0, 150,
	0, 150, 52, 0, 357, 297, 0, 79, 71, 72,
	0, 21, 100, 0, 105, 96, 0, 0, 140, 0,
	28, 0, 0, 337, 0, 295, 0, 0, 0, 0,
	358, 0, 358, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 84, 197, 198, 209, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 303, 0, 0, 0,
	0, 304, 305, 0, 0, 0, 0, 165, 0, 289,
	0, 58, 0, 0, 150, 150, 150, 0, 153, 0,
	63, 150, 66, 48, 68, 0, 80, 298, 299, 81,
	0, 74, 0, 0, 19, 106, 108, 112, 357, 98,
	331, 0, 0, 148, 0, 29, 0, 0, 32, 33,
	0, 0, 358, 0, 355, 87, 0, 90, 0, 92,
	341, 0, 358, 168, 172, 173, 174, 175, 176, 177,
	178, 179, 180, 181, 0, 155, 156, 157, 183, 0,
	185, 0, 187, 0, 189, 0, 191, 0, 193, 0,
	195, 196, 199, 0, 0, 0, 0, 201, 203, 0,
	205, 207, 0, 213, 0, 0, 0, 0, 219, 0,
	0, 238, 239, 240, 241, 242, 243, 244, 245, 0,
	0, 0, 0, 0, 0, 0, 0, 255, 256, 257,
	258, 259, 260, 261, 262, 263, 264, 265, 0, 0,
	0, 0, 296, 162, 0, 164, 293, 290, 0, 331,
	151, 152, 59, 64, 60, 154, 150, 48, 0, 0,
	65, 150, 0, 0, 0, 0, 0, 73, 69, 70,
	0, 0, 109, 113, 0, 0, 0, 0, 0, 148,
	141, 316, 0, 338, 0, 31, 339, 0, 0, 36,
	37, 85, 343, 0, 0, 358, 351, 346, 347, 348,
	349, 350, 91, 93, 94, 182, 184, 186, 188, 190,
	192, 194, 200, 202, 208, 0, 204, 206, 0, 214,
	0, 216, 218, 0, 0, 0, 0, 0, 0, 227,
	0, 0, 0, 0, 0, 237, 306, 166, 0, 291,
	0, 0, 0, 0, 61, 62, 46, 47, 67, 49,
	50, 45, 82, 75, 148, 115, 121, 0, 133, 135,
	136, 137, 107, 110, 114, 111, 335, 0, 159, 335,
	0, 333, 316, 324, 0, 149, 30, 34, 35, 0,
	0, 41, 0, 356, 88, 0, 352, 0, 210, 215,
	0, 0, 0, 0, 0, 266, 267, 268, 270, 272,
	273, 274, 275, 276, 277, 0, 0, 0, 0, 0,
	0, 287, 294, 0, 24, 25, 312, 0, 0, 124,
	125, 0, 0, 0, 0, 0, 142, 122, 0, 0,
	0, 22, 0, 158, 160, 23, 332, 0, 324, 27,
	0, 0, 38, 39, 40, 358, 89, 211, 220, 221,
	222, 223, 224, 269, 271, 225, 0, 228, 229, 230,
	0, 0, 292, 314, 0, 116, 119, 126, 0, 128,
	0, 130, 131, 132, 117, 0, 0, 0, 123, 118,
	134, 138, 139, 0, 0, 334, 26, 325, 317, 318,
	321, 86, 226, 0, 0, 0, 0, 0, 0, 316,
	0, 0, 0, 127, 129, 0, 0, 0, 0, 161,
	0, 0, 0, 320, 322, 323, 231, 232, 233, 234,
	235, 236, 324, 315, 313, 120, 0, 0, 0, 0,
	326, 327, 319, 328, 0, 146, 0, 0, 336, 20,
	0, 0, 143, 0, 144, 145, 329, 0, 147, 0,
	330,
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
	182, 183, 184, 185, 186, 187, 188, 189, 190, 191,
	192, 193, 194, 195, 196, 197, 198, 199,
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
		//line sql.y:239
		{
			SetParseTree(yylex, yyDollar[1].statement)
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:245
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 19:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:267
		{
			yyVAL.selStmt = &SimpleSelect{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs}
		}
	case 20:
		yyDollar = yyS[yypt-12 : yypt+1]
		//line sql.y:271
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyDollar[2].bytes2), Distinct: yyDollar[3].str, SelectExprs: yyDollar[4].selectExprs, From: yyDollar[6].tableExprs, Where: NewWhere(AST_WHERE, yyDollar[7].expr), GroupBy: GroupBy(yyDollar[8].exprs), Having: NewWhere(AST_HAVING, yyDollar[9].expr), OrderBy: yyDollar[10].orderBy, Limit: yyDollar[11].limit, Lock: yyDollar[12].str}
		}
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:275
		{
			yyVAL.selStmt = &Union{Type: yyDollar[2].str, Left: yyDollar[1].selStmt, Right: yyDollar[3].selStmt}
		}
	case 22:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:282
		{
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 23:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:286
		{
			cols := make(Columns, 0, len(yyDollar[6].updateExprs))
			vals := make(ValTuple, 0, len(yyDollar[6].updateExprs))
			for _, col := range yyDollar[6].updateExprs {
				cols = append(cols, &NonStarExpr{Expr: col.Name})
				vals = append(vals, col.Expr)
			}
			yyVAL.statement = &Insert{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: cols, Rows: Values{vals}, OnDup: OnDup(yyDollar[7].updateExprs)}
		}
	case 24:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:298
		{
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: yyDollar[5].columns, Rows: yyDollar[6].insRows}
		}
	case 25:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:302
		{
			cols := make(Columns, 0, len(yyDollar[6].updateExprs))
			vals := make(ValTuple, 0, len(yyDollar[6].updateExprs))
			for _, col := range yyDollar[6].updateExprs {
				cols = append(cols, &NonStarExpr{Expr: col.Name})
				vals = append(vals, col.Expr)
			}
			yyVAL.statement = &Replace{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Columns: cols, Rows: Values{vals}}
		}
	case 26:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:315
		{
			yyVAL.statement = &Update{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[3].tableName, Exprs: yyDollar[5].updateExprs, Where: NewWhere(AST_WHERE, yyDollar[6].expr), OrderBy: yyDollar[7].orderBy, Limit: yyDollar[8].limit}
		}
	case 27:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:321
		{
			yyVAL.statement = &Delete{Comments: Comments(yyDollar[2].bytes2), Table: yyDollar[4].tableName, Where: NewWhere(AST_WHERE, yyDollar[5].expr), OrderBy: yyDollar[6].orderBy, Limit: yyDollar[7].limit}
		}
	case 28:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:327
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: yyDollar[3].updateExprs}
		}
	case 29:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:331
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_client")}, Expr: StrVal(yyDollar[4].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_results")}, Expr: StrVal(yyDollar[4].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_connection")}, Expr: StrVal(yyDollar[4].bytes)},
			}}
		}
	case 30:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:339
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_client")}, Expr: StrVal(yyDollar[4].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_results")}, Expr: StrVal(yyDollar[4].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_connection")}, Expr: StrVal(yyDollar[4].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@collation_connection")}, Expr: StrVal(yyDollar[6].bytes)},
			}}
		}
	case 31:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:348
		{
			yyVAL.statement = &Set{Comments: Comments(yyDollar[2].bytes2), Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_client")}, Expr: StrVal(yyDollar[5].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@character_set_results")}, Expr: StrVal(yyDollar[5].bytes)},
				&UpdateExpr{Name: &ColName{Name: []byte("@@collation_connection")}, Expr: &ColName{Name: []byte("@@collation_database")}},
			}}
		}
	case 32:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:356
		{
			yyVAL.statement = &Set{Comments: append([][]byte{}, []byte(yyDollar[2].str), []byte("transaction"), yyDollar[4].bytes)}
		}
	case 33:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:362
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:366
		{
			yyVAL.bytes = append(yyDollar[1].bytes, append([]byte(", "), yyDollar[3].bytes...)...)
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:372
		{
			yyVAL.bytes = append([]byte("isolation level "), yyDollar[3].bytes...)
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:376
		{
			yyVAL.bytes = []byte("read write")
		}
	case 37:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:380
		{
			yyVAL.bytes = []byte("read only")
		}
	case 38:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:386
		{
			yyVAL.bytes = []byte("repeatable read")
		}
	case 39:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:390
		{
			yyVAL.bytes = []byte("read committed")
		}
	case 40:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:394
		{
			yyVAL.bytes = []byte("read uncommitted")
		}
	case 41:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:398
		{
			yyVAL.bytes = []byte("serializable")
		}
	case 42:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:404
		{
			yyVAL.statement = &Begin{}
		}
	case 43:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:410
		{
			yyVAL.statement = &Commit{}
		}
	case 44:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:416
		{
			yyVAL.statement = &Rollback{}
		}
	case 45:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:422
		{
			yyVAL.statement = &Admin{Name: yyDollar[2].bytes, Values: yyDollar[4].exprs}
		}
	case 46:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:428
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 47:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:432
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 48:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:437
		{
			yyVAL.expr = nil
		}
	case 49:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:441
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 50:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:445
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 51:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:450
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 52:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:454
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 53:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:460
		{
			yyVAL.str = AST_KILL_CONNECTION
		}
	case 54:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:464
		{
			yyVAL.str = AST_KILL_QUERY
		}
	case 55:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:469
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 56:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:473
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 57:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:477
		{
			yyVAL.str = AST_GLOBAL_SCOPE
		}
	case 58:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:484
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyDollar[3].expr}
		}
	case 59:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:488
		{
			yyVAL.statement = &Show{Section: "variables", Modifier: yyDollar[2].str, LikeOrWhere: yyDollar[4].expr}
		}
	case 60:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:492
		{
			yyVAL.statement = &Show{Section: "tables", From: yyDollar[3].expr, LikeOrWhere: yyDollar[4].expr}
		}
	case 61:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:496
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyDollar[3].bytes), From: yyDollar[4].expr, LikeOrWhere: yyDollar[5].expr}
		}
	case 62:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:500
		{
			yyVAL.statement = &Show{Section: "columns", From: yyDollar[4].expr, Modifier: yyDollar[2].str, DBFilter: yyDollar[5].expr}
		}
	case 63:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:504
		{
			yyVAL.statement = &Show{Section: "processlist", Modifier: yyDollar[2].str}
		}
	case 64:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:508
		{
			yyVAL.statement = &Show{Section: "status", Modifier: yyDollar[2].str, LikeOrWhere: yyDollar[4].expr}
		}
	case 65:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:512
		{
			yyVAL.statement = &Show{Section: "charset", LikeOrWhere: yyDollar[4].expr}
		}
	case 66:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:516
		{
			yyVAL.statement = &Show{Section: "charset", LikeOrWhere: yyDollar[3].expr}
		}
	case 67:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:520
		{
			yyVAL.statement = &Show{Section: "tablestatus", From: yyDollar[4].expr, LikeOrWhere: yyDollar[5].expr}
		}
	case 68:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:524
		{
			yyVAL.statement = &Show{Section: "collation", LikeOrWhere: yyDollar[3].expr}
		}
	case 69:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:530
		{
			yyVAL.str = AST_EXPLAIN_FORMAT_TRADITIONAL
		}
	case 70:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:534
		{
			yyVAL.str = AST_EXPLAIN_FORMAT_JSON
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:540
		{
			yyVAL.str = AST_EXPLAIN_EXTENDED
		}
	case 72:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:544
		{
			yyVAL.str = AST_EXPLAIN_PARTITIONS
		}
	case 73:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:548
		{
			yyVAL.str = yyDollar[3].str
		}
	case 74:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:554
		{
			yyVAL.statement = yyDollar[1].selStmt
		}
	case 75:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:558
		{
			yyVAL.statement = yyDollar[2].statement
		}
	case 76:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:564
		{
			yyVAL.empty = yyDollar[1].empty
		}
	case 77:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:568
		{
			yyVAL.empty = yyDollar[1].empty
		}
	case 78:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:572
		{
			yyVAL.empty = yyDollar[1].empty
		}
	case 79:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:578
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 80:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:584
		{
			yyVAL.statement = &Explain{Section: "table", Table: yyDollar[2].tableName, Column: yyDollar[3].colName}
		}
	case 81:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:588
		{
			yyVAL.statement = &Explain{Section: "plan", ExplainType: yyDollar[2].str, Statement: yyDollar[3].statement}
		}
	case 82:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:592
		{
			yyVAL.statement = &Explain{Section: "plan", ExplainType: yyDollar[2].str, Connection: yyDollar[5].bytes}
		}
	case 83:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:598
		{
			yyVAL.statement = &Kill{Scope: AST_KILL_CONNECTION, ID: yyDollar[2].expr}
		}
	case 84:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:602
		{
			yyVAL.statement = &Kill{Scope: yyDollar[2].str, ID: yyDollar[3].expr}
		}
	case 85:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:608
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 86:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:612
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 87:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:617
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 88:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:623
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 89:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:627
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 90:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:632
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 91:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:638
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 92:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:644
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 93:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:648
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 94:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:653
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 95:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:658
		{
			SetAllowComments(yylex, true)
		}
	case 96:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:662
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 97:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:668
		{
			yyVAL.bytes2 = nil
		}
	case 98:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:672
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 99:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:678
		{
			yyVAL.str = AST_UNION
		}
	case 100:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:682
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 101:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:686
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 102:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:690
		{
			yyVAL.str = AST_EXCEPT
		}
	case 103:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:694
		{
			yyVAL.str = AST_INTERSECT
		}
	case 104:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:699
		{
			yyVAL.str = ""
		}
	case 105:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:703
		{
			yyVAL.str = AST_DISTINCT
		}
	case 106:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:709
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 107:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:713
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 108:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:719
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 109:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:723
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 110:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:727
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 111:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:731
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 112:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:736
		{
			yyVAL.bytes = nil
		}
	case 113:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:740
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 114:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:744
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 115:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:750
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 116:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:754
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 117:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:760
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 118:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:764
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 119:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:768
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 120:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:772
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].expr}
		}
	case 121:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:777
		{
			yyVAL.bytes = nil
		}
	case 122:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:781
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 123:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:785
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 124:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:791
		{
			yyVAL.str = AST_JOIN
		}
	case 125:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:795
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 126:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:799
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 127:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:803
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 128:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:807
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 129:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:811
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 130:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:815
		{
			yyVAL.str = AST_JOIN
		}
	case 131:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:819
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 132:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:823
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 133:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:829
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 134:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:833
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:837
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 136:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:841
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 137:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:845
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 138:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:849
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 139:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:853
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:859
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 141:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:863
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 142:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:868
		{
			yyVAL.indexHints = nil
		}
	case 143:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:872
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 144:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:876
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 145:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:880
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 146:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:886
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:890
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 148:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:895
		{
			yyVAL.expr = nil
		}
	case 149:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:899
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 150:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:904
		{
			yyVAL.expr = nil
		}
	case 151:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:908
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 152:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:912
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 153:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:917
		{
			yyVAL.expr = nil
		}
	case 154:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:921
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 155:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:927
		{
			yyVAL.str = AST_ALL
		}
	case 156:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:931
		{
			yyVAL.str = AST_SOME
		}
	case 157:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:935
		{
			yyVAL.str = AST_ANY
		}
	case 158:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:941
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 159:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:945
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 160:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:951
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:955
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:961
		{
			yyVAL.tuple = ValTuple(yyDollar[2].exprs)
		}
	case 163:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:965
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:971
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 165:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:977
		{
			yyVAL.exprs = Exprs{yyDollar[1].expr}
		}
	case 166:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:981
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 167:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:987
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 168:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:991
		{
			yyVAL.expr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 169:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:997
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 170:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1001
		{
			yyVAL.expr = yyDollar[1].colName
		}
	case 171:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1005
		{
			yyVAL.expr = yyDollar[1].tuple
		}
	case 172:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1009
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITAND, Right: yyDollar[3].expr}
		}
	case 173:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1013
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITOR, Right: yyDollar[3].expr}
		}
	case 174:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1017
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITXOR, Right: yyDollar[3].expr}
		}
	case 175:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1021
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_PLUS, Right: yyDollar[3].expr}
		}
	case 176:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1025
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MINUS, Right: yyDollar[3].expr}
		}
	case 177:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1029
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MULT, Right: yyDollar[3].expr}
		}
	case 178:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1033
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_DIV, Right: yyDollar[3].expr}
		}
	case 179:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1037
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_IDIV, Right: yyDollar[3].expr}
		}
	case 180:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1041
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MOD, Right: yyDollar[3].expr}
		}
	case 181:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1045
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, Right: yyDollar[3].expr}
		}
	case 182:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1049
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 183:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1053
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, Right: yyDollar[3].expr}
		}
	case 184:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1057
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 185:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1061
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, Right: yyDollar[3].expr}
		}
	case 186:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1065
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 187:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1069
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, Right: yyDollar[3].expr}
		}
	case 188:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1073
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 189:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1077
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, Right: yyDollar[3].expr}
		}
	case 190:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1081
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 191:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1085
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, Right: yyDollar[3].expr}
		}
	case 192:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1089
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 193:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1093
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, Right: yyDollar[3].expr}
		}
	case 194:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1097
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 195:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1101
		{
			yyVAL.expr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 196:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1105
		{
			yyVAL.expr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 197:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1109
		{
			yyVAL.expr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 198:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1113
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
	case 199:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1128
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 200:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1132
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 201:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1136
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LIKE, Right: yyDollar[3].expr}
		}
	case 202:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1140
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_LIKE, Right: yyDollar[4].expr}
		}
	case 203:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1144
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 204:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1148
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 205:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1152
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: yyDollar[3].expr}
		}
	case 206:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1156
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: yyDollar[4].expr}
		}
	case 207:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1160
		{
			yyVAL.expr = &RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[3].expr}
		}
	case 208:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1164
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[4].expr}}
		}
	case 209:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1168
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 210:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1172
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_BETWEEN, From: yyDollar[3].expr, To: yyDollar[5].expr}
		}
	case 211:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1176
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].expr, To: yyDollar[6].expr}
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1180
		{
			yyVAL.expr = yyDollar[1].caseExpr
		}
	case 213:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1184
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 214:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1188
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 215:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1192
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 216:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1196
		{
			yyVAL.expr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1200
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 218:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1204
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 219:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1208
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 220:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1212
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 221:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1216
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 222:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1220
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 223:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1224
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 224:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1228
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 225:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1232
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 226:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:1236
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 227:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1240
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date"), Exprs: SelectExprs{yyDollar[3].selectExpr}}
		}
	case 228:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1244
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 229:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1248
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 230:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1252
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 231:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1256
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 232:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1260
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 233:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1264
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 234:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1268
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 235:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1272
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 236:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1276
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 237:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1280
		{
			yyVAL.expr = &FuncExpr{Name: []byte("str_to_date"), Exprs: yyDollar[3].selectExprs}
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1286
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1290
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1294
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1298
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1302
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1306
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1310
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1314
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 246:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1320
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1324
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 248:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1328
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 249:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1332
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 250:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1336
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 251:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1340
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1344
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 253:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1348
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 254:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1352
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 255:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1358
		{
			yyVAL.bytes = SECOND_MICROSECOND_BYTES
		}
	case 256:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1362
		{
			yyVAL.bytes = MINUTE_MICROSECOND_BYTES
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1366
		{
			yyVAL.bytes = MINUTE_SECOND_BYTES
		}
	case 258:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1370
		{
			yyVAL.bytes = HOUR_MICROSECOND_BYTES
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1374
		{
			yyVAL.bytes = HOUR_SECOND_BYTES
		}
	case 260:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1378
		{
			yyVAL.bytes = HOUR_MINUTE_BYTES
		}
	case 261:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1382
		{
			yyVAL.bytes = DAY_MICROSECOND_BYTES
		}
	case 262:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1386
		{
			yyVAL.bytes = DAY_SECOND_BYTES
		}
	case 263:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1390
		{
			yyVAL.bytes = DAY_MINUTE_BYTES
		}
	case 264:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1394
		{
			yyVAL.bytes = DAY_HOUR_BYTES
		}
	case 265:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1398
		{
			yyVAL.bytes = YEAR_MONTH_BYTES
		}
	case 266:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1404
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1408
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 268:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1412
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 269:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1416
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 270:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1420
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 271:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1424
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 272:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1428
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 273:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1432
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 274:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1436
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 275:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1440
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 276:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1444
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 277:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1451
		{
			if bytes.Equal(bytes.ToLower(yyDollar[1].bytes), []byte("datetime")) {
				yyVAL.bytes = DATETIME_BYTES
			} else {
				yylex.Error("expecting datetime")
				return 1
			}
		}
	case 278:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1462
		{
			yyVAL.bytes = IF_BYTES
		}
	case 279:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1466
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 280:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1470
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 281:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1474
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 282:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1478
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1482
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 284:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1488
		{
			yyVAL.byt = AST_UPLUS
		}
	case 285:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1492
		{
			yyVAL.byt = AST_UMINUS
		}
	case 286:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1496
		{
			yyVAL.byt = AST_TILDA
		}
	case 287:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1502
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].expr, Whens: yyDollar[3].whens, Else: yyDollar[4].expr}
		}
	case 288:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1507
		{
			yyVAL.expr = nil
		}
	case 289:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1511
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 290:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1517
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 291:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1521
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 292:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1527
		{
			yyVAL.when = &When{Cond: yyDollar[2].expr, Val: yyDollar[4].expr}
		}
	case 293:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1532
		{
			yyVAL.expr = nil
		}
	case 294:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1536
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 295:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1542
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 296:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1546
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 297:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1551
		{
			yyVAL.colName = nil
		}
	case 298:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1555
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 299:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1559
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 300:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1565
		{
			yyVAL.expr = StrVal(yyDollar[1].bytes)
		}
	case 301:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1569
		{
			yyVAL.expr = NumVal(yyDollar[1].bytes)
		}
	case 302:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1573
		{
			yyVAL.expr = ValArg(yyDollar[1].bytes)
		}
	case 303:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1577
		{
			yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 304:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1581
		{
			yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 305:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1585
		{
			yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 306:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1589
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
	case 307:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1602
		{
			yyVAL.expr = &NullVal{}
		}
	case 308:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1606
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 309:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1612
		{
			yyVAL.expr = &TrueVal{}
		}
	case 310:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1616
		{
			yyVAL.expr = &FalseVal{}
		}
	case 311:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1620
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 312:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1625
		{
			yyVAL.exprs = nil
		}
	case 313:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1629
		{
			yyVAL.exprs = yyDollar[3].exprs
		}
	case 314:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1634
		{
			yyVAL.expr = nil
		}
	case 315:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1638
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 316:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1643
		{
			yyVAL.orderBy = nil
		}
	case 317:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1647
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 318:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1653
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 319:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1657
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 320:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1663
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 321:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1668
		{
			yyVAL.str = AST_ASC
		}
	case 322:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1672
		{
			yyVAL.str = AST_ASC
		}
	case 323:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1676
		{
			yyVAL.str = AST_DESC
		}
	case 324:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1681
		{
			yyVAL.limit = nil
		}
	case 325:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1685
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].expr}
		}
	case 326:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1689
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].expr, Rowcount: yyDollar[4].expr}
		}
	case 327:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1693
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].expr, Rowcount: yyDollar[2].expr}
		}
	case 328:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1698
		{
			yyVAL.str = ""
		}
	case 329:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1702
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 330:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1706
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
	case 331:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1719
		{
			yyVAL.columns = nil
		}
	case 332:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1723
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 333:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1729
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 334:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1733
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 335:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1738
		{
			yyVAL.updateExprs = nil
		}
	case 336:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1742
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 337:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1748
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 338:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1752
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 339:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1758
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].expr}
		}
	case 340:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1763
		{
			yyVAL.empty = struct{}{}
		}
	case 341:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1765
		{
			yyVAL.empty = struct{}{}
		}
	case 342:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1768
		{
			yyVAL.empty = struct{}{}
		}
	case 343:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1770
		{
			yyVAL.empty = struct{}{}
		}
	case 344:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1773
		{
			yyVAL.empty = struct{}{}
		}
	case 345:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1775
		{
			yyVAL.empty = struct{}{}
		}
	case 346:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1779
		{
			yyVAL.empty = struct{}{}
		}
	case 347:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1781
		{
			yyVAL.empty = struct{}{}
		}
	case 348:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1783
		{
			yyVAL.empty = struct{}{}
		}
	case 349:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1785
		{
			yyVAL.empty = struct{}{}
		}
	case 350:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1787
		{
			yyVAL.empty = struct{}{}
		}
	case 351:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1790
		{
			yyVAL.empty = struct{}{}
		}
	case 352:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1792
		{
			yyVAL.empty = struct{}{}
		}
	case 353:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1795
		{
			yyVAL.empty = struct{}{}
		}
	case 354:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1797
		{
			yyVAL.empty = struct{}{}
		}
	case 355:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1800
		{
			yyVAL.empty = struct{}{}
		}
	case 356:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1802
		{
			yyVAL.empty = struct{}{}
		}
	case 357:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1806
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 358:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1811
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
