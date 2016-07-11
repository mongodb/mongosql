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
const BETWEEN = 57458
const AND = 57459
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
	"BETWEEN",
	"AND",
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
	-1, 328,
	116, 0,
	-2, 171,
	-1, 419,
	116, 0,
	-2, 172,
}

const yyNprod = 317
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 1392

var yyAct = [...]int{

	114, 507, 561, 451, 145, 188, 338, 481, 381, 353,
	443, 109, 372, 103, 280, 93, 133, 245, 44, 258,
	46, 108, 299, 75, 47, 468, 470, 89, 135, 102,
	374, 170, 82, 64, 49, 518, 50, 264, 52, 53,
	54, 85, 388, 173, 79, 517, 516, 84, 83, 250,
	86, 51, 77, 18, 90, 242, 3, 100, 96, 262,
	99, 80, 265, 498, 373, 78, 56, 58, 59, 408,
	63, 61, 62, 220, 65, 283, 473, 373, 105, 441,
	282, 166, 161, 469, 169, 283, 273, 186, 580, 66,
	282, 94, 177, 205, 206, 207, 208, 193, 194, 195,
	196, 197, 198, 201, 199, 200, 218, 198, 201, 199,
	200, 165, 181, 182, 274, 187, 168, 513, 172, 209,
	203, 202, 204, 215, 214, 216, 212, 205, 206, 207,
	208, 193, 194, 195, 196, 197, 198, 201, 199, 200,
	241, 223, 193, 194, 195, 196, 197, 198, 201, 199,
	200, 196, 197, 198, 201, 199, 200, 146, 147, 148,
	515, 79, 322, 444, 79, 324, 384, 254, 320, 321,
	319, 330, 76, 179, 180, 176, 462, 183, 460, 578,
	189, 463, 78, 461, 248, 78, 261, 263, 260, 221,
	222, 286, 475, 514, 577, 163, 269, 285, 474, 255,
	271, 286, 466, 465, 95, 287, 444, 285, 464, 268,
	251, 163, 190, 189, 251, 278, 252, 191, 244, 284,
	329, 254, 538, 502, 318, 437, 304, 306, 308, 310,
	312, 314, 326, 337, 347, 436, 348, 576, 429, 352,
	428, 354, 427, 426, 253, 575, 351, 79, 79, 366,
	367, 325, 576, 472, 435, 332, 334, 335, 434, 385,
	34, 35, 36, 37, 433, 275, 276, 368, 78, 379,
	377, 79, 289, 290, 291, 292, 293, 294, 295, 296,
	297, 298, 303, 305, 307, 309, 311, 313, 315, 316,
	317, 386, 78, 323, 390, 327, 328, 407, 380, 98,
	389, 376, 395, 576, 370, 74, 163, 525, 97, 349,
	350, 284, 397, 398, 399, 400, 401, 524, 402, 403,
	18, 19, 20, 21, 497, 376, 496, 522, 409, 394,
	383, 416, 570, 410, 569, 411, 568, 412, 567, 413,
	521, 414, 421, 415, 520, 519, 476, 448, 438, 425,
	22, 424, 391, 392, 422, 369, 277, 393, 34, 35,
	36, 37, 72, 423, 203, 202, 204, 215, 214, 216,
	212, 205, 206, 207, 208, 193, 194, 195, 196, 197,
	198, 201, 199, 200, 440, 191, 88, 446, 447, 450,
	397, 398, 399, 400, 401, 566, 402, 403, 191, 417,
	418, 419, 191, 191, 191, 449, 191, 191, 184, 191,
	458, 459, 191, 278, 278, 565, 527, 549, 530, 529,
	528, 523, 432, 270, 284, 284, 246, 233, 548, 547,
	247, 232, 546, 247, 159, 526, 134, 162, 224, 492,
	237, 91, 236, 235, 234, 231, 230, 493, 494, 495,
	79, 229, 442, 228, 504, 178, 477, 478, 479, 480,
	27, 28, 29, 227, 30, 32, 31, 226, 225, 101,
	158, 503, 406, 240, 239, 23, 24, 26, 25, 238,
	65, 80, 471, 455, 405, 454, 509, 217, 209, 203,
	202, 204, 215, 214, 216, 212, 205, 206, 207, 208,
	193, 194, 195, 196, 197, 198, 201, 199, 200, 573,
	531, 532, 267, 266, 535, 249, 73, 174, 544, 542,
	499, 171, 167, 164, 87, 160, 501, 483, 484, 574,
	534, 505, 508, 92, 71, 256, 550, 553, 552, 555,
	559, 551, 554, 18, 175, 288, 560, 562, 562, 562,
	563, 564, 69, 38, 67, 452, 512, 79, 482, 485,
	486, 487, 488, 489, 490, 491, 453, 382, 375, 511,
	457, 251, 579, 40, 41, 42, 43, 581, 78, 533,
	571, 582, 556, 583, 55, 146, 147, 148, 18, 39,
	543, 189, 545, 420, 57, 60, 272, 185, 17, 16,
	146, 147, 148, 15, 333, 14, 13, 112, 132, 12,
	257, 141, 45, 387, 259, 557, 558, 508, 106, 126,
	127, 128, 48, 134, 331, 138, 121, 131, 129, 130,
	116, 117, 118, 149, 150, 151, 152, 153, 154, 155,
	156, 157, 122, 123, 124, 81, 125, 378, 572, 539,
	506, 510, 456, 439, 243, 119, 120, 371, 213, 210,
	211, 217, 209, 203, 202, 204, 215, 214, 216, 212,
	205, 206, 207, 208, 193, 194, 195, 196, 197, 198,
	201, 199, 200, 113, 111, 143, 142, 500, 115, 445,
	107, 467, 281, 396, 110, 279, 404, 301, 302, 146,
	147, 148, 300, 192, 68, 33, 112, 132, 70, 11,
	141, 10, 9, 136, 137, 104, 144, 80, 126, 127,
	128, 139, 134, 8, 138, 121, 131, 129, 130, 116,
	117, 118, 149, 150, 151, 152, 153, 154, 155, 156,
	157, 122, 123, 124, 7, 125, 6, 5, 4, 2,
	140, 1, 0, 0, 119, 120, 0, 0, 149, 150,
	151, 152, 153, 154, 155, 156, 157, 0, 0, 0,
	540, 541, 339, 340, 341, 342, 343, 344, 345, 346,
	0, 0, 0, 0, 143, 142, 0, 0, 0, 0,
	0, 0, 0, 110, 0, 0, 0, 0, 146, 147,
	148, 0, 0, 0, 0, 112, 132, 0, 0, 141,
	0, 0, 136, 137, 0, 144, 106, 126, 127, 128,
	139, 134, 336, 138, 121, 131, 129, 130, 116, 117,
	118, 149, 150, 151, 152, 153, 154, 155, 156, 157,
	122, 123, 124, 0, 125, 0, 0, 0, 0, 140,
	0, 0, 0, 119, 120, 213, 210, 211, 217, 209,
	203, 202, 204, 215, 214, 216, 212, 205, 206, 207,
	208, 193, 194, 195, 196, 197, 198, 201, 199, 200,
	0, 0, 0, 143, 142, 0, 0, 0, 0, 0,
	0, 0, 110, 0, 0, 0, 0, 146, 147, 148,
	0, 0, 0, 0, 112, 132, 0, 0, 141, 0,
	0, 136, 137, 104, 144, 106, 126, 127, 128, 139,
	134, 0, 138, 121, 131, 129, 130, 116, 117, 118,
	149, 150, 151, 152, 153, 154, 155, 156, 157, 122,
	123, 124, 430, 125, 0, 0, 0, 0, 140, 0,
	0, 0, 119, 120, 213, 210, 211, 217, 209, 203,
	202, 204, 215, 214, 216, 212, 205, 206, 207, 208,
	193, 194, 195, 196, 197, 198, 201, 199, 200, 0,
	0, 0, 143, 142, 537, 0, 0, 0, 0, 0,
	0, 110, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 18,
	136, 137, 104, 144, 0, 0, 0, 0, 139, 0,
	0, 0, 0, 146, 147, 148, 0, 0, 0, 0,
	112, 132, 0, 0, 141, 0, 0, 0, 0, 0,
	0, 80, 126, 127, 128, 0, 134, 140, 138, 121,
	131, 129, 130, 116, 117, 118, 149, 150, 151, 152,
	153, 154, 155, 156, 157, 122, 123, 124, 0, 125,
	536, 0, 0, 0, 0, 0, 0, 0, 119, 120,
	0, 0, 213, 210, 211, 217, 209, 203, 202, 204,
	215, 214, 216, 212, 205, 206, 207, 208, 193, 194,
	195, 196, 197, 198, 201, 199, 200, 0, 143, 142,
	0, 0, 0, 0, 0, 0, 0, 110, 0, 0,
	0, 0, 146, 147, 148, 0, 0, 0, 0, 112,
	132, 0, 0, 141, 0, 0, 136, 137, 0, 144,
	80, 126, 127, 128, 139, 134, 0, 138, 121, 131,
	129, 130, 116, 117, 118, 149, 150, 151, 152, 153,
	154, 155, 156, 157, 122, 123, 124, 0, 125, 0,
	0, 0, 0, 140, 0, 0, 0, 119, 120, 0,
	0, 0, 0, 0, 0, 219, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 65, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 143, 142, 0,
	431, 0, 0, 0, 0, 0, 110, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 136, 137, 0, 144, 0,
	0, 0, 0, 139, 213, 210, 211, 217, 209, 203,
	202, 204, 215, 214, 216, 212, 205, 206, 207, 208,
	193, 194, 195, 196, 197, 198, 201, 199, 200, 0,
	0, 0, 140, 213, 210, 211, 217, 209, 203, 202,
	204, 215, 214, 216, 212, 205, 206, 207, 208, 193,
	194, 195, 196, 197, 198, 201, 199, 200, 213, 210,
	211, 217, 209, 203, 202, 204, 215, 214, 216, 212,
	205, 206, 207, 208, 193, 194, 195, 196, 197, 198,
	201, 199, 200, 213, 210, 211, 217, 209, 203, 202,
	204, 215, 214, 216, 0, 205, 206, 207, 208, 193,
	194, 195, 196, 197, 198, 201, 199, 200, 149, 150,
	151, 152, 153, 154, 155, 156, 157, 0, 0, 0,
	0, 0, 339, 340, 341, 342, 343, 344, 345, 346,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 355, 356, 357, 358, 359, 360, 361, 362, 363,
	364, 365,
}
var yyPact = [...]int{

	315, -1000, -1000, 163, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -146, -132, -113, -126, -1000, -1000, -1000,
	-1000, -86, 443, 583, 532, -1000, -1000, -1000, 529, -1000,
	503, 479, 209, 24, -137, -117, 443, -1000, -123, 443,
	-1000, 487, -142, 443, -142, 502, 82, -97, 203, 443,
	-102, -1000, -1000, -1000, 427, -1000, -1000, -1000, 878, -1000,
	429, 479, 490, -56, 479, 110, 486, -1000, -8, -1000,
	-57, 485, 3, 443, -1000, 484, -1000, -124, 480, 518,
	63, 443, 479, -1000, 1103, 1103, 82, 82, 1103, 203,
	-9, 1103, 116, -1000, -1000, 1160, -65, -1000, -1000, -1000,
	1103, 1103, 396, -1000, 426, 425, 421, 411, 409, 404,
	403, 389, 402, 401, 400, 398, -1000, -1000, -1000, 441,
	436, 435, -1000, -1000, 1004, -1000, -1000, -1000, -1000, 1103,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 391,
	444, 478, 562, 444, -1000, 1103, 443, -1000, 509, -152,
	-1000, 25, -1000, 476, -1000, -1000, 475, -1000, 388, 1131,
	1131, -1000, -1000, 1131, 82, -10, 1103, 1103, 313, 1131,
	38, 878, 521, 1103, 1103, 1103, 1103, 1103, 1103, 1103,
	1103, 1103, 680, 680, 680, 680, 680, 680, 680, 1103,
	1103, 1103, 394, 46, 1103, 138, 1103, 1103, -1000, 443,
	37, 1131, -1000, -1000, 583, 581, 878, 779, 706, 706,
	1103, 1103, 878, -1000, 1296, 878, 878, 878, -1000, -1000,
	-1000, 312, 261, -77, 1131, 538, 444, 444, 205, -1000,
	555, 1103, -1000, 1131, -1000, -1000, -1000, 54, 443, -1000,
	-125, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, 538,
	444, -1000, -1000, 1103, 1103, 1131, 1210, -1000, 1103, 201,
	288, 447, 48, -69, -1000, -1000, -1000, -1000, -1000, 19,
	19, 19, -27, -27, -1000, -1000, -1000, -1000, -32, 396,
	-1000, -1000, -1000, -32, 396, -32, 396, 13, 396, 13,
	396, 13, 396, 13, 396, 246, 371, 371, -1000, 394,
	1103, 1103, 1103, -32, -1000, 566, -1000, -32, 2, -1000,
	-1000, -1000, 311, 878, 308, 306, -1000, 142, 141, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 139, 137, 841,
	1185, 379, 168, 162, 158, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, 134, 124, 305, -1000,
	-1000, -64, -1000, 1103, 51, 394, 163, 94, 304, -1000,
	555, 541, 553, 1131, 448, -1000, -1000, 446, -1000, -1000,
	110, 1131, 1131, 1131, 560, 38, 38, -1000, -1000, 76,
	74, 106, 101, 100, -85, -1000, 445, 210, 39, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, -32, -32, 2,
	-1000, -1000, -1000, 303, -1000, -1000, 878, 878, 878, 878,
	482, 482, -1000, 878, 878, 878, 262, 260, -1000, -81,
	-1000, 1103, 545, -1000, 494, 122, -1000, -1000, -1000, 444,
	541, -1000, 1103, 1103, -1000, -1000, 558, 543, 288, 5,
	-1000, 91, -1000, 58, -1000, -1000, -1000, -1000, -119, -120,
	-130, -1000, -1000, -1000, -1000, -1000, -1000, 302, 301, 297,
	284, 378, -1000, -1000, -1000, 233, 223, -1000, -1000, -1000,
	-1000, -1000, 392, 377, 376, 375, 878, 878, -1000, 1131,
	1103, 497, 394, -1000, -1000, 969, 121, -1000, 742, -1000,
	555, 1103, 1103, 1103, -1000, -1000, 390, 387, 386, -1000,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000, 374, -1000, -1000,
	-1000, 1296, 1296, 1131, 575, -1000, 1103, 1103, 1103, -1000,
	-1000, -1000, 541, 1131, 114, 1131, 443, 443, 443, -1000,
	372, 352, 295, 293, 291, 289, 444, 1131, 1131, -1000,
	493, 202, -1000, 151, 136, -1000, -1000, -1000, -1000, -1000,
	-1000, 110, -1000, 565, -36, -1000, 443, -1000, -1000, -1000,
	443, -1000, 443, -1000,
}
var yyPgo = [...]int{

	0, 751, 749, 55, 748, 747, 746, 744, 723, 712,
	711, 709, 553, 708, 705, 22, 704, 29, 13, 703,
	696, 78, 695, 14, 693, 692, 362, 691, 2, 49,
	30, 690, 11, 16, 5, 689, 688, 4, 6, 9,
	7, 28, 684, 21, 683, 657, 12, 654, 653, 652,
	651, 8, 650, 1, 649, 3, 648, 17, 647, 10,
	23, 52, 386, 645, 622, 614, 613, 612, 610, 0,
	31, 609, 606, 605, 603, 599, 598, 308, 15, 597,
	596, 595, 594, 589,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 3, 3, 3,
	4, 4, 74, 74, 5, 6, 7, 7, 71, 72,
	73, 76, 79, 79, 80, 80, 80, 81, 81, 82,
	82, 82, 75, 75, 75, 75, 75, 8, 8, 8,
	9, 9, 9, 10, 11, 11, 11, 83, 12, 13,
	13, 14, 14, 14, 14, 14, 16, 16, 17, 17,
	18, 18, 18, 18, 19, 19, 19, 22, 22, 23,
	23, 23, 23, 20, 20, 20, 24, 24, 24, 24,
	24, 24, 24, 24, 24, 25, 25, 25, 25, 25,
	25, 25, 26, 26, 27, 27, 27, 27, 28, 28,
	29, 29, 78, 78, 78, 77, 77, 15, 15, 15,
	30, 30, 35, 35, 32, 32, 41, 34, 34, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 38,
	38, 38, 38, 38, 38, 38, 38, 37, 37, 37,
	37, 37, 37, 37, 37, 37, 39, 39, 39, 39,
	39, 39, 39, 39, 39, 39, 39, 40, 40, 40,
	40, 40, 40, 40, 40, 40, 40, 40, 40, 36,
	36, 36, 36, 36, 36, 42, 42, 42, 44, 47,
	47, 45, 45, 46, 48, 48, 43, 43, 31, 31,
	31, 31, 31, 31, 31, 31, 31, 33, 33, 33,
	49, 49, 50, 50, 51, 51, 52, 52, 53, 54,
	54, 54, 55, 55, 55, 55, 56, 56, 56, 57,
	57, 58, 58, 59, 59, 60, 60, 61, 62, 62,
	63, 63, 64, 64, 65, 65, 65, 65, 65, 66,
	66, 67, 67, 68, 68, 69, 70,
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
	1, 1, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 4, 3, 4, 3, 4, 3, 4, 3,
	4, 3, 4, 3, 4, 3, 3, 3, 2, 2,
	3, 4, 3, 4, 3, 4, 3, 4, 3, 4,
	2, 3, 4, 1, 3, 4, 5, 4, 1, 4,
	3, 6, 6, 6, 6, 6, 6, 7, 4, 6,
	6, 6, 8, 8, 8, 8, 8, 8, 4, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 2, 1, 2, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 5, 0,
	1, 1, 2, 4, 0, 2, 1, 3, 1, 1,
	1, 2, 2, 2, 2, 1, 1, 1, 1, 1,
	0, 3, 0, 2, 0, 3, 1, 3, 2, 0,
	1, 1, 0, 2, 4, 4, 0, 2, 4, 0,
	3, 1, 3, 0, 5, 1, 3, 3, 0, 2,
	0, 3, 0, 1, 1, 1, 1, 1, 1, 0,
	1, 0, 1, 0, 2, 1, 0,
}
var yyChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -11, -71, -72, -73, -74, -75, -76, 5, 6,
	7, 8, 35, 160, 161, 163, 162, 145, 146, 147,
	149, 151, 150, -14, 97, 98, 99, 100, -12, -83,
	-12, -12, -12, -12, 164, -67, 166, 170, -64, 166,
	168, 164, 164, 165, 166, -12, 152, -82, 153, 154,
	-81, 157, 158, 156, -69, 37, -3, 22, -16, 23,
	-13, 31, -26, 37, 96, -60, 148, -61, -43, -69,
	37, -63, 169, 165, -69, 164, -69, 37, -62, 169,
	-69, -62, 31, -78, 9, 122, 155, -77, 96, -69,
	159, 42, -17, -18, 134, -21, 37, -31, -43, -32,
	113, -42, 26, -44, -69, -36, 49, 50, 51, 74,
	75, 45, 61, 62, 63, 65, 38, 39, 40, 47,
	48, 46, 27, -33, 42, -41, 132, 133, 44, 140,
	169, 30, 105, 104, 135, -37, 19, 20, 21, 52,
	53, 54, 55, 56, 57, 58, 59, 60, 41, -26,
	35, 138, -26, 101, 37, 119, 138, 37, 113, -69,
	-70, 37, -70, 167, 37, 26, 112, -69, -26, -21,
	-21, -78, -78, -21, -77, -79, 96, 124, -34, -21,
	96, 101, -19, 129, 130, 131, 132, 133, 134, 136,
	137, 135, 119, 118, 120, 125, 126, 127, 128, 117,
	114, 115, 124, 113, 122, 121, 123, 116, -69, 25,
	138, -21, -21, -41, 42, 42, 42, 42, 42, 42,
	42, 42, 42, 38, 42, 42, 42, 42, 38, 38,
	38, -34, -3, -47, -21, -57, 35, 42, -60, 37,
	-29, 9, -61, -21, -69, -70, 26, -68, 171, -65,
	163, 161, 34, 162, 12, 37, 37, 37, -70, -57,
	35, -78, -80, 96, 124, -21, -21, 43, 101, -22,
	-23, -25, 42, 37, -41, 159, 153, -18, 24, -21,
	-21, -21, -21, -21, -21, -21, -21, -21, -21, -15,
	22, 17, 18, -21, -15, -21, -15, -21, -15, -21,
	-15, -21, -15, -21, -15, -21, -21, -21, -32, 124,
	122, 123, 116, -21, 27, 113, -33, -21, -21, -69,
	134, 43, -17, 23, -17, -17, 43, -37, -38, 66,
	67, 68, 69, 70, 71, 72, 73, -37, -38, -21,
	-21, -18, -37, -39, -38, 85, 86, 87, 88, 89,
	90, 91, 92, 93, 94, 95, -18, -18, -17, 43,
	43, -45, -46, 141, -30, 30, -3, -60, -58, -43,
	-29, -51, 12, -21, 112, -69, -70, -66, 167, -30,
	-60, -21, -21, -21, -29, 101, -24, 102, 103, 104,
	105, 106, 108, 109, -20, 37, 25, -23, 138, -41,
	-41, -41, -41, -41, -41, -41, -32, -21, -21, -21,
	27, -33, 43, -17, 43, 43, 101, 101, 101, 101,
	101, 25, 43, 96, 96, 96, 101, 101, 43, -48,
	-46, 143, -21, -59, 112, -35, -32, -59, 43, 101,
	-51, -55, 14, 13, 37, 37, -49, 10, -23, -23,
	102, 107, 102, 107, 102, 102, 102, -27, 110, 168,
	111, 37, 43, 37, 159, 153, 43, -17, -17, -17,
	-17, -40, 76, 45, 46, 77, 78, 79, 80, 81,
	82, 83, -40, -18, -18, -18, 64, 64, 144, -21,
	142, 32, 101, -43, -55, -21, -52, -53, -21, -70,
	-50, 11, 13, 112, 102, 102, 165, 165, 165, 43,
	43, 43, 43, 43, 84, 84, 43, 24, 43, 43,
	43, -18, -18, -21, 33, -32, 101, 15, 101, -54,
	28, 29, -51, -21, -34, -21, 42, 42, 42, 43,
	-37, -39, -38, -37, -39, -38, 7, -21, -21, -53,
	-55, -28, -69, -28, -28, 43, 43, 43, 43, 43,
	43, -60, -56, 16, 36, 43, 101, 43, 43, 7,
	124, -69, -69, -69,
}
var yyDef = [...]int{

	0, -2, 1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16, 57, 57,
	57, 57, 57, 311, 302, 0, 0, 28, 29, 30,
	57, -2, 0, 0, 61, 63, 64, 65, 66, 59,
	0, 0, 0, 0, 300, 0, 0, 312, 0, 0,
	303, 0, 298, 0, 298, 0, 112, 0, 115, 0,
	0, 40, 41, 38, 0, 315, 19, 62, 0, 67,
	58, 0, 0, 102, 0, 26, 0, 295, 0, 256,
	315, 0, 0, 0, 316, 0, 316, 0, 0, 0,
	0, 0, 0, 42, 0, 0, 112, 112, 0, 115,
	0, 0, 17, 68, 70, 74, 315, 129, 130, 131,
	0, 0, 0, 173, 256, 0, 178, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 258, 259, 260, 0,
	0, 0, 265, 266, 0, 125, 245, 246, 247, 249,
	239, 240, 241, 242, 243, 244, 267, 268, 269, 207,
	208, 209, 210, 211, 212, 213, 214, 215, 60, 289,
	0, 0, 110, 0, 27, 0, 0, 316, 0, 313,
	49, 0, 52, 0, 54, 299, 0, 316, 289, 113,
	114, 43, 44, 116, 112, 34, 0, 0, 0, 127,
	0, 0, 71, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 75, 0,
	0, 158, 159, 170, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 261, 0, 0, 0, 0, 262, 263,
	264, 0, 0, 0, 250, 0, 0, 0, 110, 103,
	274, 0, 296, 297, 257, 47, 301, 0, 0, 316,
	309, 304, 305, 306, 307, 308, 53, 55, 56, 0,
	0, 45, 46, 0, 0, 32, 33, 31, 0, 110,
	77, 83, 0, 95, 97, 98, 99, 69, 72, 132,
	133, 134, 135, 136, 137, 138, 139, 140, 141, 0,
	117, 118, 119, 143, 0, 145, 0, 147, 0, 149,
	0, 151, 0, 153, 0, 155, 156, 157, 160, 0,
	0, 0, 0, 162, 164, 0, 166, 168, -2, 76,
	73, 174, 0, 0, 0, 0, 180, 0, 0, 199,
	200, 201, 202, 203, 204, 205, 206, 0, 0, 0,
	0, 0, 0, 0, 0, 216, 217, 218, 219, 220,
	221, 222, 223, 224, 225, 226, 0, 0, 0, 124,
	126, 254, 251, 0, 293, 0, 121, 293, 0, 291,
	274, 282, 0, 111, 0, 314, 50, 0, 310, 22,
	23, 35, 36, 128, 270, 0, 0, 86, 87, 0,
	0, 0, 0, 0, 104, 84, 0, 0, 0, 142,
	144, 146, 148, 150, 152, 154, 161, 163, 169, -2,
	165, 167, 175, 0, 177, 179, 0, 0, 0, 0,
	0, 0, 188, 0, 0, 0, 0, 0, 198, 0,
	252, 0, 0, 20, 0, 120, 122, 21, 290, 0,
	282, 25, 0, 0, 316, 51, 272, 0, 78, 81,
	88, 0, 90, 0, 92, 93, 94, 79, 0, 0,
	0, 85, 80, 96, 100, 101, 176, 0, 0, 0,
	0, 0, 227, 228, 229, 230, 232, 234, 235, 236,
	237, 238, 0, 0, 0, 0, 0, 0, 248, 255,
	0, 0, 0, 292, 24, 283, 275, 276, 279, 48,
	274, 0, 0, 0, 89, 91, 0, 0, 0, 181,
	182, 183, 184, 185, 231, 233, 186, 0, 189, 190,
	191, 0, 0, 253, 0, 123, 0, 0, 0, 278,
	280, 281, 282, 273, 271, 82, 0, 0, 0, 187,
	0, 0, 0, 0, 0, 0, 0, 284, 285, 277,
	286, 0, 108, 0, 0, 192, 193, 194, 195, 196,
	197, 294, 18, 0, 0, 105, 0, 106, 107, 287,
	0, 109, 0, 288,
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
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:784
		{
			yyVAL.expr = yyDollar[1].colName
		}
	case 131:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:788
		{
			yyVAL.expr = yyDollar[1].tuple
		}
	case 132:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:792
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITAND, Right: yyDollar[3].expr}
		}
	case 133:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:796
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITOR, Right: yyDollar[3].expr}
		}
	case 134:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:800
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_BITXOR, Right: yyDollar[3].expr}
		}
	case 135:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:804
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_PLUS, Right: yyDollar[3].expr}
		}
	case 136:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:808
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MINUS, Right: yyDollar[3].expr}
		}
	case 137:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:812
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MULT, Right: yyDollar[3].expr}
		}
	case 138:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:816
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_DIV, Right: yyDollar[3].expr}
		}
	case 139:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:820
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_IDIV, Right: yyDollar[3].expr}
		}
	case 140:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:824
		{
			yyVAL.expr = &BinaryExpr{Left: yyDollar[1].expr, Operator: AST_MOD, Right: yyDollar[3].expr}
		}
	case 141:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:828
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, Right: yyDollar[3].expr}
		}
	case 142:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:832
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_EQ, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 143:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:836
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, Right: yyDollar[3].expr}
		}
	case 144:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:840
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 145:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:844
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, Right: yyDollar[3].expr}
		}
	case 146:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:848
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NSE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 147:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:852
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, Right: yyDollar[3].expr}
		}
	case 148:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:856
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 149:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:860
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, Right: yyDollar[3].expr}
		}
	case 150:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:864
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GT, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 151:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:868
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, Right: yyDollar[3].expr}
		}
	case 152:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:872
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 153:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:876
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, Right: yyDollar[3].expr}
		}
	case 154:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:880
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_GE, SubqueryOperator: yyDollar[3].str, Right: yyDollar[4].subquery}
		}
	case 155:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:884
		{
			yyVAL.expr = &AndExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 156:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:888
		{
			yyVAL.expr = &OrExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 157:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:892
		{
			yyVAL.expr = &XorExpr{Left: yyDollar[1].expr, Right: yyDollar[3].expr}
		}
	case 158:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:896
		{
			yyVAL.expr = &NotExpr{Expr: yyDollar[2].expr}
		}
	case 159:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:900
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
	case 160:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:915
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IN, Right: yyDollar[3].tuple}
		}
	case 161:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:919
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_IN, Right: yyDollar[4].tuple}
		}
	case 162:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:923
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_LIKE, Right: yyDollar[3].expr}
		}
	case 163:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:927
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_NOT_LIKE, Right: yyDollar[4].expr}
		}
	case 164:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:931
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 165:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:935
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 166:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:939
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS, Right: yyDollar[3].expr}
		}
	case 167:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:943
		{
			yyVAL.expr = &ComparisonExpr{Left: yyDollar[1].expr, Operator: AST_IS_NOT, Right: yyDollar[4].expr}
		}
	case 168:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:947
		{
			yyVAL.expr = &RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[3].expr}
		}
	case 169:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:951
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyDollar[1].expr, Pattern: yyDollar[4].expr}}
		}
	case 170:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:955
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyDollar[2].subquery}
		}
	case 171:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:959
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_BETWEEN, Range: yyDollar[3].expr}
		}
	case 172:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:963
		{
			yyVAL.expr = &RangeCond{Left: yyDollar[1].expr, Operator: AST_NOT_BETWEEN, Range: yyDollar[4].expr}
		}
	case 173:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:967
		{
			yyVAL.expr = yyDollar[1].caseExpr
		}
	case 174:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:971
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes)}
		}
	case 175:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:975
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Exprs: yyDollar[3].selectExprs}
		}
	case 176:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:979
		{
			yyVAL.expr = &FuncExpr{Name: bytes.ToLower(yyDollar[1].bytes), Distinct: true, Exprs: yyDollar[4].selectExprs}
		}
	case 177:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:983
		{
			yyVAL.expr = &FuncExpr{Name: yyDollar[1].bytes, Exprs: yyDollar[3].selectExprs}
		}
	case 178:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:987
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 179:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:991
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 180:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:995
		{
			yyVAL.expr = &FuncExpr{Name: []byte("current_timestamp")}
		}
	case 181:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:999
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 182:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1003
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 183:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1007
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 184:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1011
		{
			yyVAL.expr = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExprs...)}
		}
	case 185:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1015
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 186:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1019
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 187:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line sql.y:1023
		{
			yyVAL.expr = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyDollar[3].expr}, &NonStarExpr{Expr: KeywordVal(yyDollar[5].bytes)}})}
		}
	case 188:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1027
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date"), Exprs: SelectExprs{yyDollar[3].selectExpr}}
		}
	case 189:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1031
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 190:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1035
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 191:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line sql.y:1039
		{
			yyVAL.expr = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyDollar[3].bytes)}}, yyDollar[5].selectExpr)}
		}
	case 192:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1043
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 193:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1047
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 194:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1051
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 195:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1055
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 196:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1059
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 197:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line sql.y:1063
		{
			yyVAL.expr = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{yyDollar[3].selectExpr, yyDollar[6].selectExpr, &NonStarExpr{Expr: KeywordVal(yyDollar[7].bytes)}})}
		}
	case 198:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1067
		{
			yyVAL.expr = &FuncExpr{Name: []byte("str_to_date"), Exprs: yyDollar[3].selectExprs}
		}
	case 199:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1073
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 200:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1077
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 201:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1081
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 202:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1085
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 203:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1089
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 204:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1093
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 205:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1097
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 206:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1101
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 207:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1107
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 208:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1111
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 209:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1115
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 210:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1119
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 211:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1123
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 212:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1127
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 213:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1131
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 214:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1135
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 215:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1139
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 216:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1145
		{
			yyVAL.bytes = SECOND_MICROSECOND_BYTES
		}
	case 217:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1149
		{
			yyVAL.bytes = MINUTE_MICROSECOND_BYTES
		}
	case 218:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1153
		{
			yyVAL.bytes = MINUTE_SECOND_BYTES
		}
	case 219:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1157
		{
			yyVAL.bytes = HOUR_MICROSECOND_BYTES
		}
	case 220:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1161
		{
			yyVAL.bytes = HOUR_SECOND_BYTES
		}
	case 221:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1165
		{
			yyVAL.bytes = HOUR_MINUTE_BYTES
		}
	case 222:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1169
		{
			yyVAL.bytes = DAY_MICROSECOND_BYTES
		}
	case 223:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1173
		{
			yyVAL.bytes = DAY_SECOND_BYTES
		}
	case 224:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1177
		{
			yyVAL.bytes = DAY_MINUTE_BYTES
		}
	case 225:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1181
		{
			yyVAL.bytes = DAY_HOUR_BYTES
		}
	case 226:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1185
		{
			yyVAL.bytes = YEAR_MONTH_BYTES
		}
	case 227:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1191
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 228:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1195
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 229:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1199
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 230:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1203
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 231:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1207
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 232:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1211
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 233:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1215
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 234:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1219
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 235:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1223
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 236:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1227
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 237:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1231
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 238:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1235
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 239:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1241
		{
			yyVAL.bytes = IF_BYTES
		}
	case 240:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1245
		{
			yyVAL.bytes = VALUES_BYTES
		}
	case 241:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1249
		{
			yyVAL.bytes = RIGHT_BYTES
		}
	case 242:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1253
		{
			yyVAL.bytes = LEFT_BYTES
		}
	case 243:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1257
		{
			yyVAL.bytes = MOD_BYTES
		}
	case 244:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1261
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 245:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1267
		{
			yyVAL.byt = AST_UPLUS
		}
	case 246:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1271
		{
			yyVAL.byt = AST_UMINUS
		}
	case 247:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1275
		{
			yyVAL.byt = AST_TILDA
		}
	case 248:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1281
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyDollar[2].expr, Whens: yyDollar[3].whens, Else: yyDollar[4].expr}
		}
	case 249:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1286
		{
			yyVAL.expr = nil
		}
	case 250:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1290
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 251:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1296
		{
			yyVAL.whens = []*When{yyDollar[1].when}
		}
	case 252:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1300
		{
			yyVAL.whens = append(yyDollar[1].whens, yyDollar[2].when)
		}
	case 253:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1306
		{
			yyVAL.when = &When{Cond: yyDollar[2].expr, Val: yyDollar[4].expr}
		}
	case 254:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1311
		{
			yyVAL.expr = nil
		}
	case 255:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1315
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 256:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1321
		{
			yyVAL.colName = &ColName{Name: yyDollar[1].bytes}
		}
	case 257:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1325
		{
			yyVAL.colName = &ColName{Qualifier: yyDollar[1].bytes, Name: yyDollar[3].bytes}
		}
	case 258:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1331
		{
			yyVAL.expr = StrVal(yyDollar[1].bytes)
		}
	case 259:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1335
		{
			yyVAL.expr = NumVal(yyDollar[1].bytes)
		}
	case 260:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1339
		{
			yyVAL.expr = ValArg(yyDollar[1].bytes)
		}
	case 261:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1343
		{
			yyVAL.expr = DateVal{Name: AST_DATE, Val: yyDollar[2].bytes}
		}
	case 262:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1347
		{
			yyVAL.expr = DateVal{Name: AST_TIME, Val: yyDollar[2].bytes}
		}
	case 263:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1351
		{
			yyVAL.expr = DateVal{Name: AST_TIMESTAMP, Val: yyDollar[2].bytes}
		}
	case 264:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1355
		{
			yyVAL.expr = DateVal{Name: AST_DATETIME, Val: yyDollar[2].bytes}
		}
	case 265:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1359
		{
			yyVAL.expr = &NullVal{}
		}
	case 266:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1363
		{
			yyVAL.expr = yyDollar[1].expr
		}
	case 267:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1369
		{
			yyVAL.expr = &TrueVal{}
		}
	case 268:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1373
		{
			yyVAL.expr = &FalseVal{}
		}
	case 269:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1377
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 270:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1382
		{
			yyVAL.exprs = nil
		}
	case 271:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1386
		{
			yyVAL.exprs = yyDollar[3].exprs
		}
	case 272:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1391
		{
			yyVAL.expr = nil
		}
	case 273:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1395
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 274:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1400
		{
			yyVAL.orderBy = nil
		}
	case 275:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1404
		{
			yyVAL.orderBy = yyDollar[3].orderBy
		}
	case 276:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1410
		{
			yyVAL.orderBy = OrderBy{yyDollar[1].order}
		}
	case 277:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1414
		{
			yyVAL.orderBy = append(yyDollar[1].orderBy, yyDollar[3].order)
		}
	case 278:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1420
		{
			yyVAL.order = &Order{Expr: yyDollar[1].expr, Direction: yyDollar[2].str}
		}
	case 279:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1425
		{
			yyVAL.str = AST_ASC
		}
	case 280:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1429
		{
			yyVAL.str = AST_ASC
		}
	case 281:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1433
		{
			yyVAL.str = AST_DESC
		}
	case 282:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1438
		{
			yyVAL.limit = nil
		}
	case 283:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1442
		{
			yyVAL.limit = &Limit{Rowcount: yyDollar[2].expr}
		}
	case 284:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1446
		{
			yyVAL.limit = &Limit{Offset: yyDollar[2].expr, Rowcount: yyDollar[4].expr}
		}
	case 285:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1450
		{
			yyVAL.limit = &Limit{Offset: yyDollar[4].expr, Rowcount: yyDollar[2].expr}
		}
	case 286:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1455
		{
			yyVAL.str = ""
		}
	case 287:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1459
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 288:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sql.y:1463
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
	case 289:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1476
		{
			yyVAL.columns = nil
		}
	case 290:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1480
		{
			yyVAL.columns = yyDollar[2].columns
		}
	case 291:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1486
		{
			yyVAL.columns = Columns{&NonStarExpr{Expr: yyDollar[1].colName}}
		}
	case 292:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1490
		{
			yyVAL.columns = append(yyVAL.columns, &NonStarExpr{Expr: yyDollar[3].colName})
		}
	case 293:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1495
		{
			yyVAL.updateExprs = nil
		}
	case 294:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sql.y:1499
		{
			yyVAL.updateExprs = yyDollar[5].updateExprs
		}
	case 295:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1505
		{
			yyVAL.updateExprs = UpdateExprs{yyDollar[1].updateExpr}
		}
	case 296:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1509
		{
			yyVAL.updateExprs = append(yyDollar[1].updateExprs, yyDollar[3].updateExpr)
		}
	case 297:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1515
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyDollar[1].colName, Expr: yyDollar[3].expr}
		}
	case 298:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1520
		{
			yyVAL.empty = struct{}{}
		}
	case 299:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1522
		{
			yyVAL.empty = struct{}{}
		}
	case 300:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1525
		{
			yyVAL.empty = struct{}{}
		}
	case 301:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sql.y:1527
		{
			yyVAL.empty = struct{}{}
		}
	case 302:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1530
		{
			yyVAL.empty = struct{}{}
		}
	case 303:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1532
		{
			yyVAL.empty = struct{}{}
		}
	case 304:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1536
		{
			yyVAL.empty = struct{}{}
		}
	case 305:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1538
		{
			yyVAL.empty = struct{}{}
		}
	case 306:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1540
		{
			yyVAL.empty = struct{}{}
		}
	case 307:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1542
		{
			yyVAL.empty = struct{}{}
		}
	case 308:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1544
		{
			yyVAL.empty = struct{}{}
		}
	case 309:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1547
		{
			yyVAL.empty = struct{}{}
		}
	case 310:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1549
		{
			yyVAL.empty = struct{}{}
		}
	case 311:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1552
		{
			yyVAL.empty = struct{}{}
		}
	case 312:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1554
		{
			yyVAL.empty = struct{}{}
		}
	case 313:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1557
		{
			yyVAL.empty = struct{}{}
		}
	case 314:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sql.y:1559
		{
			yyVAL.empty = struct{}{}
		}
	case 315:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sql.y:1563
		{
			yyVAL.bytes = yyDollar[1].bytes
		}
	case 316:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line sql.y:1568
		{
			ForceEOF(yylex)
		}
	}
	goto yystack /* stack new state and value */
}
