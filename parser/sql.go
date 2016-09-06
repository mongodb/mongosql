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
const KILL = 57518
const CONNECTION = 57519
const QUERY = 57520
const SESSION = 57521
const GLOBAL = 57522
const CREATE = 57523
const ALTER = 57524
const DROP = 57525
const RENAME = 57526
const TABLE = 57527
const INDEX = 57528
const VIEW = 57529
const TO = 57530
const IGNORE = 57531
const IF = 57532
const UNIQUE = 57533
const USING = 57534

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
	150, 54,
	-2, 80,
	-1, 33,
	171, 50,
	173, 50,
	-2, 54,
}

const yyNprod = 341
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1629

var yyAct = [...]int{

	72, 228, 66, 573, 139, 204, 457, 618, 523, 67,
	104, 439, 515, 301, 450, 316, 367, 205, 3, 240,
	478, 226, 352, 91, 142, 215, 94, 246, 362, 380,
	154, 49, 60, 51, 338, 128, 470, 52, 559, 561,
	147, 54, 136, 55, 249, 591, 144, 590, 143, 57,
	58, 59, 149, 130, 589, 151, 148, 150, 56, 155,
	47, 48, 210, 182, 213, 222, 225, 123, 92, 214,
	183, 184, 118, 120, 121, 218, 127, 219, 126, 386,
	370, 124, 527, 528, 526, 47, 48, 442, 375, 376,
	374, 125, 441, 564, 145, 206, 19, 185, 495, 464,
	208, 384, 463, 339, 387, 465, 241, 339, 242, 426,
	163, 166, 164, 165, 631, 560, 161, 162, 163, 166,
	164, 165, 217, 514, 361, 233, 203, 353, 442, 129,
	287, 238, 244, 441, 586, 285, 286, 284, 211, 265,
	158, 159, 160, 161, 162, 163, 166, 164, 165, 245,
	516, 466, 252, 553, 235, 354, 349, 253, 554, 254,
	255, 256, 257, 258, 259, 260, 261, 262, 263, 264,
	269, 271, 273, 275, 277, 279, 281, 282, 231, 248,
	288, 234, 292, 293, 350, 235, 283, 221, 551, 223,
	588, 140, 141, 552, 312, 313, 516, 36, 37, 38,
	39, 300, 310, 291, 333, 311, 587, 315, 557, 556,
	317, 336, 629, 341, 342, 555, 314, 445, 346, 329,
	330, 444, 628, 566, 626, 535, 206, 565, 449, 359,
	355, 534, 418, 144, 356, 143, 144, 365, 143, 357,
	372, 335, 533, 347, 532, 19, 20, 21, 22, 383,
	385, 382, 340, 212, 295, 297, 298, 596, 445, 568,
	369, 420, 444, 368, 368, 419, 331, 36, 37, 38,
	39, 377, 627, 412, 411, 23, 343, 344, 345, 520,
	473, 390, 627, 351, 627, 357, 410, 399, 400, 401,
	437, 357, 391, 409, 398, 373, 417, 392, 416, 393,
	421, 394, 357, 395, 357, 396, 216, 397, 270, 272,
	274, 276, 278, 280, 403, 157, 181, 168, 167, 169,
	179, 178, 180, 176, 170, 171, 172, 173, 158, 159,
	160, 161, 162, 163, 166, 164, 165, 423, 538, 521,
	357, 427, 408, 138, 537, 105, 106, 107, 407, 494,
	335, 432, 433, 289, 405, 435, 436, 501, 235, 493,
	357, 448, 333, 406, 144, 144, 143, 455, 453, 428,
	459, 431, 425, 334, 422, 605, 153, 604, 603, 446,
	452, 467, 456, 443, 602, 601, 600, 28, 29, 31,
	576, 543, 540, 461, 503, 504, 505, 506, 507, 542,
	508, 509, 357, 32, 34, 33, 472, 541, 357, 468,
	430, 539, 536, 429, 357, 434, 30, 415, 363, 563,
	364, 24, 25, 27, 26, 364, 612, 195, 496, 611,
	144, 194, 143, 335, 499, 489, 156, 610, 490, 491,
	492, 290, 93, 498, 186, 224, 452, 199, 198, 197,
	196, 193, 192, 513, 191, 500, 190, 189, 188, 187,
	230, 518, 512, 522, 332, 201, 519, 200, 443, 129,
	92, 562, 592, 531, 511, 530, 474, 475, 476, 477,
	503, 504, 505, 506, 507, 529, 508, 509, 177, 174,
	175, 157, 181, 168, 167, 169, 179, 178, 180, 546,
	170, 171, 172, 173, 158, 159, 160, 161, 162, 163,
	166, 164, 165, 549, 550, 544, 545, 460, 389, 388,
	371, 366, 144, 137, 569, 250, 571, 574, 443, 443,
	247, 570, 181, 168, 167, 169, 179, 178, 180, 176,
	170, 171, 172, 173, 158, 159, 160, 161, 162, 163,
	166, 164, 165, 243, 236, 577, 580, 575, 579, 582,
	578, 581, 170, 171, 172, 173, 158, 159, 160, 161,
	162, 163, 166, 164, 165, 624, 202, 152, 593, 237,
	232, 220, 567, 209, 40, 19, 607, 206, 609, 135,
	606, 608, 378, 46, 251, 625, 614, 615, 574, 447,
	616, 105, 106, 107, 133, 42, 43, 44, 45, 402,
	451, 619, 619, 619, 144, 617, 143, 117, 622, 620,
	621, 131, 105, 106, 107, 524, 296, 119, 632, 70,
	90, 585, 633, 100, 634, 525, 458, 584, 548, 368,
	229, 84, 85, 86, 630, 93, 294, 89, 613, 97,
	79, 19, 87, 88, 74, 75, 76, 108, 109, 110,
	111, 112, 113, 114, 115, 116, 80, 81, 82, 41,
	83, 61, 122, 348, 18, 14, 17, 16, 15, 77,
	78, 13, 177, 174, 175, 157, 181, 168, 167, 169,
	179, 178, 180, 176, 170, 171, 172, 173, 158, 159,
	160, 161, 162, 163, 166, 164, 165, 12, 379, 102,
	101, 497, 50, 469, 381, 53, 146, 462, 68, 239,
	454, 267, 268, 105, 106, 107, 266, 623, 597, 572,
	70, 90, 583, 547, 100, 424, 207, 95, 96, 227,
	103, 92, 84, 85, 86, 98, 93, 337, 89, 71,
	97, 79, 69, 87, 88, 74, 75, 76, 108, 109,
	110, 111, 112, 113, 114, 115, 116, 80, 81, 82,
	73, 83, 517, 65, 558, 440, 502, 438, 62, 510,
	77, 78, 358, 132, 177, 174, 175, 157, 181, 168,
	167, 169, 179, 99, 180, 176, 170, 171, 172, 173,
	158, 159, 160, 161, 162, 163, 166, 164, 165, 35,
	102, 101, 134, 11, 10, 9, 8, 7, 6, 68,
	5, 4, 2, 1, 105, 106, 107, 0, 0, 0,
	0, 70, 90, 0, 0, 100, 0, 0, 95, 96,
	0, 103, 229, 84, 85, 86, 98, 93, 299, 89,
	0, 97, 79, 0, 87, 88, 74, 75, 76, 108,
	109, 110, 111, 112, 113, 114, 115, 116, 80, 81,
	82, 0, 83, 0, 0, 0, 0, 0, 0, 0,
	0, 77, 78, 108, 109, 110, 111, 112, 113, 114,
	115, 116, 0, 0, 99, 0, 0, 302, 303, 304,
	305, 306, 307, 308, 309, 0, 0, 0, 0, 0,
	0, 102, 101, 0, 0, 0, 0, 0, 0, 0,
	68, 0, 0, 0, 0, 105, 106, 107, 0, 0,
	0, 0, 70, 90, 0, 0, 100, 0, 0, 95,
	96, 227, 103, 92, 84, 85, 86, 98, 93, 0,
	89, 0, 97, 79, 0, 87, 88, 74, 75, 76,
	108, 109, 110, 111, 112, 113, 114, 115, 116, 80,
	81, 82, 0, 83, 0, 0, 0, 0, 0, 0,
	0, 488, 77, 78, 0, 0, 0, 0, 0, 0,
	0, 480, 0, 0, 0, 99, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 102, 101, 0, 0, 0, 0, 0, 0,
	0, 68, 479, 481, 482, 483, 484, 485, 486, 487,
	0, 105, 106, 107, 0, 0, 0, 0, 70, 90,
	95, 96, 100, 103, 0, 0, 0, 0, 98, 229,
	84, 85, 86, 0, 93, 0, 89, 0, 97, 79,
	0, 87, 88, 74, 75, 76, 108, 109, 110, 111,
	112, 113, 114, 115, 116, 80, 81, 82, 0, 83,
	0, 0, 0, 63, 64, 0, 0, 0, 77, 78,
	0, 0, 0, 0, 0, 0, 99, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 102, 101,
	0, 0, 598, 599, 0, 0, 0, 68, 0, 0,
	0, 0, 0, 0, 0, 0, 19, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 95, 96, 227, 103,
	105, 106, 107, 0, 98, 0, 0, 70, 90, 0,
	0, 100, 0, 0, 0, 0, 0, 0, 92, 84,
	85, 86, 0, 93, 0, 89, 0, 97, 79, 0,
	87, 88, 74, 75, 76, 108, 109, 110, 111, 112,
	113, 114, 115, 116, 80, 81, 82, 0, 83, 0,
	0, 0, 99, 0, 0, 0, 0, 77, 78, 177,
	174, 175, 157, 181, 168, 167, 169, 179, 178, 180,
	176, 170, 171, 172, 173, 158, 159, 160, 161, 162,
	163, 166, 164, 165, 0, 0, 0, 102, 101, 0,
	0, 0, 0, 0, 0, 0, 68, 0, 0, 0,
	0, 105, 106, 107, 0, 0, 0, 0, 70, 90,
	0, 0, 100, 0, 0, 95, 96, 0, 103, 92,
	84, 85, 86, 98, 93, 595, 89, 0, 97, 79,
	0, 87, 88, 74, 75, 76, 108, 109, 110, 111,
	112, 113, 114, 115, 116, 80, 81, 82, 0, 83,
	0, 0, 0, 0, 0, 0, 0, 0, 77, 78,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 99, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 360, 0, 0, 0, 0, 102, 101,
	0, 0, 0, 0, 0, 129, 0, 68, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 414, 0,
	0, 0, 0, 594, 0, 0, 95, 96, 0, 103,
	0, 0, 0, 0, 98, 177, 174, 175, 157, 181,
	168, 167, 169, 179, 178, 180, 176, 170, 171, 172,
	173, 158, 159, 160, 161, 162, 163, 166, 164, 165,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 99, 177, 174, 175, 157, 181, 168, 167,
	169, 179, 178, 180, 176, 170, 171, 172, 173, 158,
	159, 160, 161, 162, 163, 166, 164, 165, 177, 174,
	175, 157, 181, 168, 167, 169, 179, 178, 180, 176,
	170, 171, 172, 173, 158, 159, 160, 161, 162, 163,
	166, 164, 165, 413, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 177, 174, 175, 157, 181,
	168, 167, 169, 179, 178, 180, 176, 170, 171, 172,
	173, 158, 159, 160, 161, 162, 163, 166, 164, 165,
	177, 174, 175, 157, 181, 168, 167, 169, 179, 178,
	180, 176, 170, 171, 172, 173, 158, 159, 160, 161,
	162, 163, 166, 164, 165, 177, 174, 175, 471, 181,
	168, 167, 169, 179, 178, 180, 176, 170, 171, 172,
	173, 158, 159, 160, 161, 162, 163, 166, 164, 165,
	177, 174, 175, 404, 181, 168, 167, 169, 179, 178,
	180, 176, 170, 171, 172, 173, 158, 159, 160, 161,
	162, 163, 166, 164, 165, 108, 109, 110, 111, 112,
	113, 114, 115, 116, 0, 0, 0, 0, 0, 302,
	303, 304, 305, 306, 307, 308, 309, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 318, 319,
	320, 321, 322, 323, 324, 325, 326, 327, 328,
}
var yyPact = [...]int{

	240, -1000, -1000, 98, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -119, -154, -146, -127, -136, -1000, -1000,
	906, -1000, -1000, -94, 432, 646, 599, -1000, -1000, -1000,
	581, -1000, 558, 486, 245, 31, -56, -1000, -1000, -150,
	-130, 432, -1000, -128, 432, -1000, 540, -160, 432, -160,
	1395, 1232, -1000, -1000, -1000, -1000, -1000, -1000, 1232, 1232,
	402, -1000, 417, 416, 415, 414, 412, 410, 409, 389,
	408, 407, 406, 405, -1000, -1000, -1000, 429, 427, 539,
	-1000, -1000, -14, 1131, -1000, -1000, -1000, -1000, 1232, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 552, 129, -105,
	208, 432, -96, 546, 129, -109, 129, -1000, 403, -1000,
	-1000, -1000, 1012, -1000, 419, 486, 545, -15, 486, 51,
	517, 544, -1000, 10, -1000, -45, 516, 17, 432, -1000,
	493, -1000, -144, 488, 568, 38, 432, 1232, 1232, 1232,
	1232, 1232, 1232, 1232, 1232, 1232, 1232, 704, 704, 704,
	704, 704, 704, 704, 1232, 1232, 400, 11, 1232, 326,
	1232, 1232, 1395, 1395, -1000, -1000, 646, 603, 1012, 805,
	829, 829, 1232, 1232, 1012, -1000, 1531, 1012, 1012, 1012,
	-1000, -1000, 426, 432, 330, 168, 1395, -40, 1395, 486,
	-1000, 1232, 1232, 129, 129, 129, 1232, 208, 58, -1000,
	129, -1000, 29, -1000, 1232, 136, -1000, -1000, 1308, -16,
	-1000, 383, 433, 484, 630, 433, -82, 483, 1232, 192,
	-1000, -62, -66, -1000, 566, -163, -1000, 67, -1000, 482,
	-1000, -1000, 481, -1000, 1395, -18, -18, -18, -26, -26,
	-1000, -1000, -1000, -1000, 435, 402, -1000, -1000, -1000, 435,
	402, 435, 402, 9, 402, 9, 402, 9, 402, 9,
	402, 197, 197, -1000, 400, 1232, 1232, 1232, 435, -1000,
	582, -1000, 435, 1445, -1000, 311, 1012, 305, 299, -1000,
	190, 183, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	171, 170, 1370, 1333, 374, 200, 198, 134, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 162,
	158, 257, 329, -1000, -1000, 1232, -1000, -36, -1000, 1232,
	378, 1395, 1395, -1000, -1000, -1000, 1395, 129, 29, 1232,
	1232, -1000, 129, 1232, 1232, 247, 50, 1012, 575, -1000,
	432, 92, 580, 433, 433, 255, -1000, 624, 1232, -1000,
	480, -1000, 1395, -45, -54, -1000, -1000, -1000, -1000, 37,
	432, -1000, -152, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 435,
	435, 1420, -1000, -1000, 1232, -1000, 237, -1000, -1000, 1012,
	1012, 1012, 1012, 944, 944, -1000, 1012, 1012, 1012, 293,
	283, -1000, -1000, 1395, -48, -1000, 1232, 567, 580, 433,
	-1000, -1000, 1395, 373, -1000, 1395, 669, -1000, 254, 290,
	437, 91, -17, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	36, 400, 98, 82, 236, -1000, 624, 611, 622, 1395,
	-1000, -1000, -1000, -69, -75, -1000, 448, -1000, -1000, 438,
	-1000, 1232, 413, -1000, 201, 199, 188, 182, 369, -1000,
	-1000, 258, 252, -1000, -1000, -1000, -1000, -1000, -1000, 368,
	364, 356, 348, 1012, 1012, -1000, 1395, 1232, -1000, 51,
	628, 50, 50, -1000, -1000, 84, 49, 111, 105, 104,
	-74, -1000, 434, 376, 56, -1000, 550, 156, -1000, -1000,
	-1000, 433, 611, -1000, 1232, 1232, -1000, -1000, -1000, -1000,
	-1000, 413, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	347, -1000, -1000, -1000, 1531, 1531, 1395, 626, 618, 290,
	20, -1000, 102, -1000, 86, -1000, -1000, -1000, -1000, -132,
	-139, -141, -1000, -1000, -1000, -1000, -1000, 439, 400, -1000,
	-1000, 1260, 154, -1000, 1094, -1000, -1000, 343, 342, 341,
	335, 334, 332, 624, 1232, 1232, 1232, -1000, -1000, 395,
	387, 384, 641, -1000, 1232, 1232, 1232, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 611, 1395, 138, 1395,
	432, 432, 432, 433, 1395, 1395, -1000, 559, 181, -1000,
	179, 169, 51, -1000, 637, -12, -1000, 432, -1000, -1000,
	-1000, 432, -1000, 432, -1000,
}
var yyPgo = [...]int{

	0, 823, 822, 17, 821, 820, 818, 817, 816, 815,
	814, 813, 584, 812, 809, 139, 783, 66, 21, 782,
	779, 1, 778, 777, 11, 776, 775, 42, 774, 7,
	16, 14, 773, 9, 23, 5, 772, 770, 10, 13,
	15, 20, 26, 752, 2, 749, 747, 34, 736, 735,
	733, 732, 6, 729, 3, 728, 8, 727, 28, 720,
	12, 4, 24, 719, 19, 717, 376, 716, 715, 714,
	713, 712, 708, 0, 27, 707, 681, 678, 677, 676,
	675, 674, 25, 62, 673, 22, 672, 593, 671, 669,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 3, 3,
	3, 4, 4, 78, 78, 5, 6, 7, 7, 7,
	7, 7, 63, 63, 64, 64, 64, 65, 65, 65,
	65, 75, 76, 77, 81, 84, 84, 85, 85, 85,
	86, 86, 88, 88, 87, 87, 87, 79, 79, 79,
	79, 79, 79, 79, 79, 79, 79, 79, 80, 80,
	8, 8, 8, 9, 9, 9, 10, 11, 11, 11,
	89, 12, 13, 13, 14, 14, 14, 14, 14, 16,
	16, 17, 17, 18, 18, 18, 18, 19, 19, 19,
	23, 23, 24, 24, 24, 24, 20, 20, 20, 25,
	25, 25, 25, 25, 25, 25, 25, 25, 26, 26,
	26, 26, 26, 26, 26, 27, 27, 28, 28, 28,
	28, 29, 29, 30, 30, 83, 83, 83, 82, 82,
	15, 15, 15, 31, 31, 36, 36, 33, 33, 42,
	35, 35, 21, 21, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 39, 39, 39, 39, 39, 39, 39,
	39, 38, 38, 38, 38, 38, 38, 38, 38, 38,
	40, 40, 40, 40, 40, 40, 40, 40, 40, 40,
	40, 41, 41, 41, 41, 41, 41, 41, 41, 41,
	41, 41, 41, 37, 37, 37, 37, 37, 37, 43,
	43, 43, 45, 48, 48, 46, 46, 47, 49, 49,
	44, 44, 32, 32, 32, 32, 32, 32, 32, 32,
	32, 34, 34, 34, 50, 50, 51, 51, 52, 52,
	53, 53, 54, 55, 55, 55, 56, 56, 56, 56,
	57, 57, 57, 58, 58, 59, 59, 60, 60, 61,
	61, 62, 66, 66, 67, 67, 68, 68, 69, 69,
	69, 69, 69, 70, 70, 71, 71, 72, 72, 73,
	74,
}
var yyR2 = [...]int{

	0, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 4, 12,
	3, 7, 7, 6, 6, 8, 7, 3, 4, 6,
	5, 4, 1, 3, 3, 2, 2, 2, 2, 2,
	1, 1, 1, 1, 5, 2, 2, 0, 2, 2,
	0, 1, 1, 1, 0, 1, 1, 3, 4, 4,
	5, 5, 3, 4, 4, 3, 5, 3, 2, 3,
	5, 8, 4, 6, 7, 4, 5, 4, 5, 5,
	0, 2, 0, 2, 1, 2, 1, 1, 1, 0,
	1, 1, 3, 1, 2, 3, 3, 0, 1, 2,
	1, 3, 3, 3, 3, 5, 0, 1, 2, 1,
	1, 2, 3, 2, 3, 2, 2, 2, 1, 3,
	1, 1, 1, 3, 3, 1, 3, 0, 5, 5,
	5, 1, 3, 0, 2, 0, 2, 2, 0, 2,
	1, 1, 1, 2, 1, 1, 3, 3, 1, 3,
	1, 3, 1, 3, 1, 1, 1, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 4, 3, 4,
	3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
	3, 3, 2, 2, 3, 4, 3, 4, 3, 4,
	3, 4, 3, 4, 2, 5, 6, 1, 3, 4,
	5, 4, 1, 4, 3, 6, 6, 6, 6, 6,
	6, 7, 4, 6, 6, 6, 8, 8, 8, 8,
	8, 8, 4, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 2, 1, 2, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 5, 0, 1, 1, 2, 4, 0, 2,
	1, 3, 1, 1, 1, 2, 2, 2, 4, 1,
	1, 1, 1, 1, 0, 3, 0, 2, 0, 3,
	1, 3, 2, 0, 1, 1, 0, 2, 4, 4,
	0, 2, 4, 0, 3, 1, 3, 0, 5, 1,
	3, 3, 0, 2, 0, 3, 0, 1, 1, 1,
	1, 1, 1, 0, 1, 0, 1, 0, 2, 1,
	0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -75, -76, -80, -77, -78, -79, -81, 5,
	6, 7, 8, 35, 181, 182, 184, 183, 147, 148,
	176, 149, 163, 165, 164, -14, 99, 100, 101, 102,
	-12, -89, -12, -12, -12, -12, -87, 179, 180, 185,
	-71, 187, 191, -68, 187, 189, 185, 185, 186, 187,
	-21, -88, -22, 177, 178, -32, -44, -33, 115, -43,
	26, -45, -73, -37, 51, 52, 53, 76, 77, 47,
	63, 64, 65, 67, 38, 39, 40, 49, 50, 44,
	27, -34, 37, 42, -42, 134, 135, 46, 142, 190,
	30, 107, 106, 137, -38, 19, 20, 21, 54, 55,
	56, 57, 58, 59, 60, 61, 62, -12, 166, -87,
	167, 168, -86, 161, 175, 185, 172, 170, -73, 37,
	-3, 22, -16, 23, -13, 31, -27, 37, 98, -61,
	160, 161, -62, -44, -73, 150, -67, 190, 186, -73,
	185, -73, 37, -66, 190, -73, -66, 118, 131, 132,
	133, 134, 135, 136, 138, 139, 137, 121, 120, 122,
	127, 128, 129, 130, 116, 117, 126, 115, 124, 123,
	125, 119, -21, -21, -21, -42, 42, 42, 42, 42,
	42, 42, 42, 42, 42, 38, 42, 42, 42, 42,
	38, 38, 37, 140, -35, -3, -21, -48, -21, 31,
	-83, 9, 124, 169, 174, -82, 98, -73, 171, 173,
	35, -83, 174, -83, 42, -17, -18, 136, -21, 37,
	41, -27, 35, 140, -27, 103, 37, 35, 121, -63,
	-64, 151, 153, 37, 115, -73, -74, 37, -74, 188,
	37, 26, 114, -73, -21, -21, -21, -21, -21, -21,
	-21, -21, -21, -21, -21, -15, 22, 17, 18, -21,
	-15, -21, -15, -21, -15, -21, -15, -21, -15, -21,
	-15, -21, -21, -33, 126, 124, 125, 119, -21, 27,
	115, -34, -21, -21, 43, -17, 23, -17, -17, 43,
	-38, -39, 68, 69, 70, 71, 72, 73, 74, 75,
	-38, -39, -21, -21, -18, -38, -40, -39, 87, 88,
	89, 90, 91, 92, 93, 94, 95, 96, 97, -18,
	-18, -17, 38, -73, 43, 103, 43, -46, -47, 143,
	-27, -21, -21, -83, -83, -83, -21, -82, -84, 98,
	126, -83, -85, 98, 126, -35, 98, 103, -19, -73,
	25, 140, -58, 35, 42, -61, 37, -30, 9, -62,
	162, 37, -21, 103, 152, 154, 155, -74, 26, -72,
	192, -69, 184, 182, 34, 183, 12, 37, 37, 37,
	-74, -42, -42, -42, -42, -42, -42, -42, -33, -21,
	-21, -21, 27, -34, 118, 43, -17, 43, 43, 103,
	103, 103, 103, 103, 25, 43, 98, 98, 98, 103,
	103, 43, 45, -21, -49, -47, 145, -21, -58, 35,
	-83, -85, -21, -21, -83, -21, -21, 43, -23, -24,
	-26, 42, 37, -42, 171, 167, -18, 24, -73, 136,
	-31, 30, -3, -61, -59, -44, -30, -52, 12, -21,
	37, -64, -65, 156, 153, 159, 114, -73, -74, -70,
	188, 118, -21, 43, -17, -17, -17, -17, -41, 78,
	47, 79, 80, 81, 82, 83, 84, 85, 37, -41,
	-18, -18, -18, 66, 66, 146, -21, 144, -31, -61,
	-30, 103, -25, 104, 105, 106, 107, 108, 110, 111,
	-20, 37, 25, -24, 140, -60, 114, -36, -33, -60,
	43, 103, -52, -56, 14, 13, 153, 157, 158, 37,
	37, -21, 43, 43, 43, 43, 43, 86, 86, 43,
	24, 43, 43, 43, -18, -18, -21, -50, 10, -24,
	-24, 104, 109, 104, 109, 104, 104, 104, -28, 112,
	189, 113, 37, 43, 37, 171, 167, 32, 103, -44,
	-56, -21, -53, -54, -21, -74, 43, -38, -40, -39,
	-38, -40, -39, -51, 11, 13, 114, 104, 104, 186,
	186, 186, 33, -33, 103, 15, 103, -55, 28, 29,
	43, 43, 43, 43, 43, 43, -52, -21, -35, -21,
	42, 42, 42, 7, -21, -21, -54, -56, -29, -73,
	-29, -29, -61, -57, 16, 36, 43, 103, 43, 43,
	7, 126, -73, -73, -73,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 17, 80,
	80, 80, 80, -2, 335, 326, 0, 0, 41, 42,
	0, 43, 80, -2, 0, 0, 84, 86, 87, 88,
	89, 82, 0, 0, 0, 0, 0, 55, 56, 324,
	0, 0, 336, 0, 0, 327, 0, 322, 0, 322,
	68, 0, 152, 52, 53, 154, 155, 156, 0, 0,
	0, 197, 280, 0, 202, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 282, 283, 284, 0, 0, 0,
	289, 290, 339, 0, 148, 269, 270, 271, 273, 263,
	264, 265, 266, 267, 268, 291, 292, 293, 231, 232,
	233, 234, 235, 236, 237, 238, 239, 0, 135, 0,
	138, 0, 0, 0, 135, 0, 135, 51, 0, 339,
	20, 85, 0, 90, 81, 0, 0, 125, 0, 27,
	0, 0, 319, 0, 280, 0, 0, 0, 0, 340,
	0, 340, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 69, 182, 183, 194, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 285, 0, 0, 0, 0,
	286, 287, 0, 0, 0, 0, 150, 0, 274, 0,
	57, 0, 0, 135, 135, 135, 0, 138, 0, 62,
	135, 65, 47, 67, 0, 18, 91, 93, 97, 339,
	83, 313, 0, 0, 133, 0, 28, 0, 0, 31,
	32, 0, 0, 340, 0, 337, 72, 0, 75, 0,
	77, 323, 0, 340, 153, 157, 158, 159, 160, 161,
	162, 163, 164, 165, 166, 0, 140, 141, 142, 168,
	0, 170, 0, 172, 0, 174, 0, 176, 0, 178,
	0, 180, 181, 184, 0, 0, 0, 0, 186, 188,
	0, 190, 192, 0, 198, 0, 0, 0, 0, 204,
	0, 0, 223, 224, 225, 226, 227, 228, 229, 230,
	0, 0, 0, 0, 0, 0, 0, 0, 240, 241,
	242, 243, 244, 245, 246, 247, 248, 249, 250, 0,
	0, 0, 0, 281, 147, 0, 149, 278, 275, 0,
	313, 136, 137, 58, 63, 59, 139, 135, 47, 0,
	0, 64, 135, 0, 0, 0, 0, 0, 94, 98,
	0, 0, 0, 0, 0, 133, 126, 298, 0, 320,
	0, 30, 321, 0, 0, 35, 36, 70, 325, 0,
	0, 340, 333, 328, 329, 330, 331, 332, 76, 78,
	79, 167, 169, 171, 173, 175, 177, 179, 185, 187,
	193, 0, 189, 191, 0, 199, 0, 201, 203, 0,
	0, 0, 0, 0, 0, 212, 0, 0, 0, 0,
	0, 222, 288, 151, 0, 276, 0, 0, 0, 0,
	60, 61, 45, 46, 66, 48, 49, 44, 133, 100,
	106, 0, 118, 120, 121, 122, 92, 95, 99, 96,
	317, 0, 144, 317, 0, 315, 298, 306, 0, 134,
	29, 33, 34, 0, 0, 40, 0, 338, 73, 0,
	334, 0, 195, 200, 0, 0, 0, 0, 0, 251,
	252, 253, 255, 257, 258, 259, 260, 261, 262, 0,
	0, 0, 0, 0, 0, 272, 279, 0, 23, 24,
	294, 0, 0, 109, 110, 0, 0, 0, 0, 0,
	127, 107, 0, 0, 0, 21, 0, 143, 145, 22,
	314, 0, 306, 26, 0, 0, 37, 38, 39, 340,
	74, 196, 205, 206, 207, 208, 209, 254, 256, 210,
	0, 213, 214, 215, 0, 0, 277, 296, 0, 101,
	104, 111, 0, 113, 0, 115, 116, 117, 102, 0,
	0, 0, 108, 103, 119, 123, 124, 0, 0, 316,
	25, 307, 299, 300, 303, 71, 211, 0, 0, 0,
	0, 0, 0, 298, 0, 0, 0, 112, 114, 0,
	0, 0, 0, 146, 0, 0, 0, 302, 304, 305,
	216, 217, 218, 219, 220, 221, 306, 297, 295, 105,
	0, 0, 0, 0, 308, 309, 301, 310, 0, 131,
	0, 0, 318, 19, 0, 0, 128, 0, 129, 130,
	311, 0, 132, 0, 312,
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
	192,
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
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:494
		{
			yyVAL.statement = &Show{Section: "processlist", Modifier: yyDollar[2].str}
		}
	case 63:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:498
		{
			yyVAL.statement = &Show{Section: "status", Modifier: yyDollar[2].str, LikeOrWhere: yyDollar[4].expr}
		}
	case 64:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:502
		{
			yyVAL.statement = &Show{Section: "charset", LikeOrWhere: yyDollar[4].expr}
		}
	case 65:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:506
		{
			yyVAL.statement = &Show{Section: "charset", LikeOrWhere: yyDollar[3].expr}
		}
	case 66:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:510
		{
			yyVAL.statement = &Show{Section: "tablestatus", From: yyDollar[4].expr, LikeOrWhere: yyDollar[5].expr}
		}
	case 67:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:514
		{
			yyVAL.statement = &Show{Section: "collation", LikeOrWhere: yyDollar[3].expr}
		}
	case 68:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:520
		{
			yyVAL.statement = &Kill{Scope: AST_KILL_CONNECTION, ID: yyDollar[2].expr}
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:524
		{
			yyVAL.statement = &Kill{Scope: yyDollar[2].str, ID: yyDollar[3].expr}
		}
	case 70:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:530
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[4].bytes}
		}
	case 71:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:534
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[7].bytes, NewName: yyDollar[7].bytes}
		}
	case 72:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:539
		{
			yyVAL.statement = &DDL{Action: AST_CREATE, NewName: yyDollar[3].bytes}
		}
	case 73:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:545
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[4].bytes, NewName: yyDollar[4].bytes}
		}
	case 74:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:549
		{
			// Change this to a rename statement
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[4].bytes, NewName: yyDollar[7].bytes}
		}
	case 75:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:554
		{
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[3].bytes, NewName: yyDollar[3].bytes}
		}
	case 76:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:560
		{
			yyVAL.statement = &DDL{Action: AST_RENAME, Table: yyDollar[3].bytes, NewName: yyDollar[5].bytes}
		}
	case 77:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:566
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 78:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:570
		{
			// Change this to an alter statement
			yyVAL.statement = &DDL{Action: AST_ALTER, Table: yyDollar[5].bytes, NewName: yyDollar[5].bytes}
		}
	case 79:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:575
		{
			yyVAL.statement = &DDL{Action: AST_DROP, Table: yyDollar[4].bytes}
		}
	case 80:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:580
		{
			SetAllowComments(yylex, true)
		}
	case 81:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:584
		{
			yyVAL.bytes2 = yyDollar[2].bytes2
			SetAllowComments(yylex, false)
		}
	case 82:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:590
		{
			yyVAL.bytes2 = nil
		}
	case 83:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:594
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[2].bytes)
		}
	case 84:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:600
		{
			yyVAL.str = AST_UNION
		}
	case 85:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:604
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 86:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:608
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 87:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:612
		{
			yyVAL.str = AST_EXCEPT
		}
	case 88:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:616
		{
			yyVAL.str = AST_INTERSECT
		}
	case 89:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:621
		{
			yyVAL.str = ""
		}
	case 90:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:625
		{
			yyVAL.str = AST_DISTINCT
		}
	case 91:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:631
		{
			yyVAL.selectExprs = SelectExprs{yyDollar[1].selectExpr}
		}
	case 92:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:635
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyDollar[3].selectExpr)
		}
	case 93:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:641
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 94:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:645
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:649
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyDollar[1].expr, As: yyDollar[2].bytes}
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:653
		{
			yyVAL.selectExpr = &StarExpr{TableName: yyDollar[1].bytes}
		}
	case 97:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:658
		{
			yyVAL.bytes = nil
		}
	case 98:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:662
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 99:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:666
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 100:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:672
		{
			yyVAL.tableExprs = TableExprs{yyDollar[1].tableExpr}
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:676
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyDollar[3].tableExpr)
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:682
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyDollar[1].smTableExpr, As: yyDollar[2].bytes, Hints: yyDollar[3].indexHints}
		}
	case 103:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:686
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyDollar[2].tableExpr}
		}
	case 104:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:690
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr}
		}
	case 105:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:694
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyDollar[1].tableExpr, Join: yyDollar[2].str, RightExpr: yyDollar[3].tableExpr, On: yyDollar[5].expr}
		}
	case 106:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:699
		{
			yyVAL.bytes = nil
		}
	case 107:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:703
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 108:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:707
		{
			yyVAL.bytes = yyDollar[2].bytes
		}
	case 109:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:713
		{
			yyVAL.str = AST_JOIN
		}
	case 110:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:717
		{
			yyVAL.str = AST_STRAIGHT_JOIN
		}
	case 111:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:721
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 112:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:725
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 113:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:729
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 114:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:733
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 115:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:737
		{
			yyVAL.str = AST_JOIN
		}
	case 116:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:741
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 117:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:745
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 118:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:751
		{
			yyVAL.smTableExpr = &TableName{Name: yyDollar[1].bytes}
		}
	case 119:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:755
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 120:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:759
		{
			yyVAL.smTableExpr = yyDollar[1].subquery
		}
	case 121:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:763
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("columns")}
		}
	case 122:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:767
		{
			yyVAL.smTableExpr = &TableName{Name: []byte("tables")}
		}
	case 123:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:771
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("columns")}
		}
	case 124:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:775
		{
			yyVAL.smTableExpr = &TableName{Qualifier: yyDollar[1].bytes, Name: []byte("tables")}
		}
	case 125:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:781
		{
			yyVAL.tableName = &TableName{Name: yyDollar[1].bytes}
		}
	case 126:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:785
		{
			yyVAL.tableName = &TableName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 127:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:790
		{
			yyVAL.indexHints = nil
		}
	case 128:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:794
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyDollar[4].bytes2}
		}
	case 129:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:798
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyDollar[4].bytes2}
		}
	case 130:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:802
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyDollar[4].bytes2}
		}
	case 131:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:808
		{
			yyVAL.bytes2 = [][]byte{yyDollar[1].bytes}
		}
	case 132:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:812
		{
			yyVAL.bytes2 = append(yyDollar[1].bytes2, yyDollar[3].bytes)
		}
	case 133:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:817
		{
			yyVAL.expr = nil
		}
	case 134:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:821
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 135:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:826
		{
			yyVAL.expr = nil
		}
	case 136:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:830
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 137:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:834
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 138:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:839
		{
			yyVAL.expr = nil
		}
	case 139:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:843
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 140:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:849
		{
			yyVAL.str = AST_ALL
		}
	case 141:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:853
		{
			yyVAL.str = AST_SOME
		}
	case 142:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:857
		{
			yyVAL.str = AST_ANY
		}
	case 143:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:863
		{
			yyVAL.insRows = yyDollar[2].values
		}
	case 144:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:867
		{
			yyVAL.insRows = yyDollar[1].selStmt
		}
	case 145:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:873
		{
			yyVAL.values = Values{yyDollar[1].tuple}
		}
	case 146:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:877
		{
			yyVAL.values = append(yyDollar[1].values, yyDollar[3].tuple)
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:883
		{
			yyVAL.tuple = ValTuple(yyDollar[2].exprs)
		}
	case 148:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:887
		{
			yyVAL.tuple = yyDollar[1].subquery
		}
	case 149:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:893
		{
			yyVAL.subquery = &Subquery{yyDollar[2].selStmt}
		}
	case 150:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:899
		{
			yyVAL.exprs = Exprs{yyDollar[1].expr}
		}
	case 151:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:903
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 152:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:909
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:913
		{
			yyVAL.expr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 154:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:919
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 155:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:923
		{
			yyVAL.expr = yyDollar[1].colName
		}
	case 156:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:927
		{
			yyVAL.expr = yyDollar[1].tuple
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:931
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITAND, Right: yyDollar[3].expr}
		}
	case 158:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:935
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITOR, Right: yyDollar[3].expr}
		}
	case 159:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:939
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITXOR, Right: yyDollar[3].expr}
		}
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:943
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_PLUS, Right: yyDollar[3].expr}
		}
	case 161:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:947
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MINUS, Right: yyDollar[3].expr}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:951
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MULT, Right: yyDollar[3].expr}
		}
	case 163:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:955
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_DIV, Right: yyDollar[3].expr}
		}
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:959
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_IDIV, Right: yyDollar[3].expr}
		}
	case 165:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:963
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MOD, Right: yyDollar[3].expr}
		}
	case 166:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:967
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, Right: yyDollar[3].expr}
		}
	case 167:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:971
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 168:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:975
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, Right: yyDollar[3].expr}
		}
	case 169:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:979
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 170:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:983
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, Right: yyDollar[3].expr}
		}
	case 171:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:987
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 172:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:991
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, Right: yyDollar[3].expr}
		}
	case 173:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:995
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 174:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:999
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, Right: yyDollar[3].expr}
		}
	case 175:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1003
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 176:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1007
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, Right: yyDollar[3].expr}
		}
	case 177:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1011
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 178:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1015
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, Right: yyDollar[3].expr}
		}
	case 179:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1019
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 180:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.expr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 181:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1027
		{
			yyVAL.expr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 182:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1031
		{
			yyVAL.expr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 183:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1035
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
	case 184:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1050
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 185:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1054
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 186:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1058
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LIKE, Right: yyDollar[3].expr}
		}
	case 187:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1062
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_LIKE, Right: yyDollar[4].expr}
		}
	case 188:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1066
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 189:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1070
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 190:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1074
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: yyDollar[3].expr}
		}
	case 191:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1078
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: yyDollar[4].expr}
		}
	case 192:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1082
		{
			yyVAL.expr = &RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[3].expr}
		}
	case 193:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1086
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[4].expr}}
		}
	case 194:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1090
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 195:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1094
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_BETWEEN, From: yyDollar[3].expr, To: yyDollar[5].expr}
		}
	case 196:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1098
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_NOT_BETWEEN, From: yyDollar[4].expr, To: yyDollar[6].expr}
		}
	case 197:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1102
		{
			yyVAL.expr = yyDollar[1].caseExpr
		}
	case 198:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1106
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 199:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1110
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 200:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1114
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 201:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1118
		{
			yyVAL.expr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1122
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 203:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1126
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 204:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1130
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 205:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1134
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 206:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1138
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 207:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1142
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 208:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1146
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 209:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1150
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 210:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1154
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 211:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:1158
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 212:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1162
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date"), Exprs: SelectExprs{yyDollar[3].selectExpr}}
		}
	case 213:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1166
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 214:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1170
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 215:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1174
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 216:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1178
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 217:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1182
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 218:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1186
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 219:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1190
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 220:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1194
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 221:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1198
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 222:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1202
		{
			yyVAL.expr = &FuncExpr{Name: []byte("str_to_date"), Exprs: yyDollar[3].selectExprs}
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1208
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1212
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1216
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1220
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1224
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1228
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1232
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1236
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 231:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1242
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1246
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 233:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1250
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1254
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1258
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1262
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1266
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1270
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1274
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1280
		{
			yyVAL.bytes = SECOND_MICROSECOND_BYTES
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1284
		{
			yyVAL.bytes = MINUTE_MICROSECOND_BYTES
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1288
		{
			yyVAL.bytes = MINUTE_SECOND_BYTES
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1292
		{
			yyVAL.bytes = HOUR_MICROSECOND_BYTES
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1296
		{
			yyVAL.bytes = HOUR_SECOND_BYTES
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1300
		{
			yyVAL.bytes = HOUR_MINUTE_BYTES
		}
	case 246:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1304
		{
			yyVAL.bytes = DAY_MICROSECOND_BYTES
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1308
		{
			yyVAL.bytes = DAY_SECOND_BYTES
		}
	case 248:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1312
		{
			yyVAL.bytes = DAY_MINUTE_BYTES
		}
	case 249:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1316
		{
			yyVAL.bytes = DAY_HOUR_BYTES
		}
	case 250:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1320
		{
			yyVAL.bytes = YEAR_MONTH_BYTES
		}
	case 251:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1326
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 252:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1330
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 253:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1334
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 254:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1338
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 255:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1342
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 256:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1346
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 257:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1350
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 258:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1354
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1358
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 260:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1362
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 261:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1366
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 262:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1373
		{
			if bytes.Equal(bytes.ToLower(yyDollar[1].bytes), []byte("datetime")) {
				yyVAL.bytes = DATETIME_BYTES
			} else {
				yylex.Error("expecting datetime")
				return 1
			}
		}
	case 263:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1384
		{
			yyVAL.bytes = IF_BYTES
		}
	case 264:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1388
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 265:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1392
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 266:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1396
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1400
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 268:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1404
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 269:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1410
		{
			yyVAL.byt = AST_UPLUS
		}
	case 270:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1414
		{
			yyVAL.byt = AST_UMINUS
		}
	case 271:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1418
		{
			yyVAL.byt = AST_TILDA
		}
	case 272:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1424
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].expr, Whens: yyDollar[3].whens, Else: yyDollar[4].expr}
		}
	case 273:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1429
		{
			yyVAL.expr = nil
		}
	case 274:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1433
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 275:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1439
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 276:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1443
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 277:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1449
		{
			yyVAL.when = &When{Cond: yyDollar[2].expr, Val: yyDollar[4].expr}
		}
	case 278:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1454
		{
			yyVAL.expr = nil
		}
	case 279:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1458
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 280:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1464
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 281:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1468
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 282:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1474
		{
			yyVAL.expr = StrVal(yyDollar[1].bytes)
		}
	case 283:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1478
		{
			yyVAL.expr = NumVal(yyDollar[1].bytes)
		}
	case 284:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1482
		{
			yyVAL.expr = ValArg(yyDollar[1].bytes)
		}
	case 285:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1486
		{
			yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 286:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1490
		{
			yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 287:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1494
		{
			yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 288:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1498
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
	case 289:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1511
		{
			yyVAL.expr = &NullVal{}
		}
	case 290:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1515
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 291:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1521
		{
			yyVAL.expr = &TrueVal{}
		}
	case 292:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1525
		{
			yyVAL.expr = &FalseVal{}
		}
	case 293:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1529
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 294:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1534
		{
			yyVAL.exprs = nil
		}
	case 295:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1538
		{
			yyVAL.exprs = yyDollar[3].exprs
		}
	case 296:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1543
		{
			yyVAL.expr = nil
		}
	case 297:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1547
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 298:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1552
		{
			yyVAL.orderBy = nil
		}
	case 299:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1556
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 300:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1562
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 301:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1566
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 302:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1572
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 303:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1577
		{
			yyVAL.str = AST_ASC
		}
	case 304:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1581
		{
			yyVAL.str = AST_ASC
		}
	case 305:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1585
		{
			yyVAL.str = AST_DESC
		}
	case 306:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1590
		{
			yyVAL.limit = nil
		}
	case 307:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1594
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].expr}
		}
	case 308:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1598
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].expr, Rowcount: yyDollar[4].expr}
		}
	case 309:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1602
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].expr, Rowcount: yyDollar[2].expr}
		}
	case 310:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1607
		{
			yyVAL.str = ""
		}
	case 311:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1611
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 312:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1615
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
	case 313:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1628
		{
			yyVAL.columns = nil
		}
	case 314:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1632
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 315:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1638
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 316:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1642
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 317:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1647
		{
			yyVAL.updateExprs = nil
		}
	case 318:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1651
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 319:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1657
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 320:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1661
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 321:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1667
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].expr}
		}
	case 322:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1672
		{
			yyVAL.empty = struct{}{}
		}
	case 323:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1674
		{
			yyVAL.empty = struct{}{}
		}
	case 324:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1677
		{
			yyVAL.empty = struct{}{}
		}
	case 325:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1679
		{
			yyVAL.empty = struct{}{}
		}
	case 326:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1682
		{
			yyVAL.empty = struct{}{}
		}
	case 327:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1684
		{
			yyVAL.empty = struct{}{}
		}
	case 328:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1688
		{
			yyVAL.empty = struct{}{}
		}
	case 329:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1690
		{
			yyVAL.empty = struct{}{}
		}
	case 330:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1692
		{
			yyVAL.empty = struct{}{}
		}
	case 331:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1694
		{
			yyVAL.empty = struct{}{}
		}
	case 332:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1696
		{
			yyVAL.empty = struct{}{}
		}
	case 333:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1699
		{
			yyVAL.empty = struct{}{}
		}
	case 334:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1701
		{
			yyVAL.empty = struct{}{}
		}
	case 335:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1704
		{
			yyVAL.empty = struct{}{}
		}
	case 336:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1706
		{
			yyVAL.empty = struct{}{}
		}
	case 337:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1709
		{
			yyVAL.empty = struct{}{}
		}
	case 338:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1711
		{
			yyVAL.empty = struct{}{}
		}
	case 339:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1715
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 340:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1720
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
