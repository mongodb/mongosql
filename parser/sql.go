package parser

import __yyfmt__ "fmt"

import (
	"bytes"

	"github.com/10gen/sqlproxy/internal/option"
	"strconv"
)

func SetParseTree(yylex interface{}, stmt Statement) {
	yylex.(*Tokenizer).ParseTree = stmt
}

func SetAllowComments(yylex interface{}, allow bool) {
	yylex.(*Tokenizer).AllowComments = allow
}

func ForceEOF(yylex interface{}) {
	yylex.(*Tokenizer).ForceEOF = true
}

type yySymType struct {
	yys               int
	empty             struct{}
	statement         Statement
	selStmt           SelectStatement
	bool              bool
	intopt            option.Int
	byt               byte
	bytes             []byte
	cte               *CTE
	cte_list          []*CTE
	keyPart           KeyPart
	keyPartList       []KeyPart
	with              *With
	str               string
	strs              []string
	stropt            OptString
	selectExprs       SelectExprs
	selectExpr        SelectExpr
	columnOrIndexDefs []ColumnOrIndexDefinition
	columnOrIndexDef  ColumnOrIndexDefinition
	columns           Columns
	columnList        []*ColName
	columnExprs       ColumnExprs
	colName           *ColName
	colTy             ColumnType
	tableExprs        TableExprs
	tableExpr         TableExpr
	tableOption       TableOption
	tableOptions      []TableOption
	smTableExpr       SimpleTableExpr
	tableName         *TableName
	indexHints        *IndexHints
	expr              Expr
	tuple             Tuple
	exprs             Exprs
	subquery          *Subquery
	caseExpr          *CaseExpr
	whens             []*When
	when              *When
	orderBy           OrderBy
	order             *Order
	limit             *Limit
	updateExprs       UpdateExprs
	updateExpr        *UpdateExpr
	alterSpec         *AlterSpec
	alterSpecs        []*AlterSpec
	tableLock         TableLock
	tableLocks        []TableLock
	lockType          LockType
	renameSpec        *RenameSpec
	renameSpecs       []*RenameSpec
	queryGlobals      *QueryGlobals
	val               Value
	valueList         ValueList
	valueListList     ValueListList
}

type yyXError struct {
	state, xsym int
}

const (
	yyDefault          = 57635
	yyEofCode          = 57344
	ADD                = 57380
	ADDDATE            = 57420
	ALL                = 57390
	ALTER              = 57379
	AND                = 57601
	ANY                = 57372
	AS                 = 57393
	ASC                = 57396
	AUTO_INCREMENT     = 57546
	BETWEEN            = 57604
	BIGINT             = 57437
	BINARY             = 57459
	BINLOG             = 57467
	BIT                = 57443
	BIT_AND            = 57621
	BIT_OR             = 57622
	BLOB               = 57444
	BOOL               = 57445
	BOOLEAN            = 57442
	BOTH               = 57453
	BTREE              = 57550
	BY                 = 57367
	CARET              = 57623
	CASCADE            = 57544
	CASE               = 57605
	CAST               = 57425
	CHANGE             = 57381
	CHANNEL            = 57489
	CHAR               = 57426
	CHARACTER          = 57515
	CHARSET            = 57526
	CODE               = 57477
	COLLATE            = 57516
	COLLATION          = 57523
	COLUMN             = 57384
	COLUMNS            = 57522
	COMMA              = 57587
	COMMENT            = 57351
	COMMENT_KWD        = 57386
	COMMITTED          = 57511
	CONNECTION         = 57538
	CONVERT            = 57424
	COUNT              = 57476
	CREATE             = 57359
	CROSS              = 57594
	CURRENT_DATE       = 57408
	CURRENT_TIMESTAMP  = 57407
	DATABASE           = 57462
	DATABASES          = 57517
	DATE               = 57403
	DATETIME           = 57404
	DATE_ADD           = 57419
	DATE_SUB           = 57421
	DAY                = 57558
	DAY_HOUR           = 57572
	DAY_MICROSECOND    = 57569
	DAY_MINUTE         = 57571
	DAY_SECOND         = 57570
	DBS                = 57527
	DECIMAL            = 57411
	DEFAULT            = 57400
	DESC               = 57397
	DESCRIBE           = 57529
	DISABLE            = 57553
	DISTINCT           = 57391
	DIV                = 57628
	DOT                = 57630
	DOUBLE             = 57438
	DROP               = 57358
	DUAL               = 57497
	ELSE               = 57608
	ENABLE             = 57552
	END                = 57632
	ENGINE             = 57471
	ENGINES            = 57473
	ENUM               = 57446
	EQ                 = 57609
	ERRORS             = 57475
	ESCAPE             = 57495
	EVENT              = 57464
	EVENTS             = 57468
	EXCEPT             = 57585
	EXISTS             = 57394
	EXPLAIN            = 57528
	EXTENDED           = 57530
	EXTRACT            = 57418
	FALSE              = 57374
	FLOAT              = 57412
	FLUSH              = 57536
	FN                 = 57493
	FOR                = 57370
	FORCE              = 57596
	FORMAT             = 57532
	FROM               = 57582
	FULL               = 57521
	FULLTEXT           = 57387
	FUNCTION           = 57465
	GE                 = 57611
	GLOBAL             = 57541
	GRANTS             = 57478
	GROUP              = 57364
	GROUP_CONCAT       = 57414
	GT                 = 57612
	HASH               = 57551
	HAVING             = 57365
	HOSTS              = 57486
	HOUR               = 57559
	HOUR_MICROSECOND   = 57566
	HOUR_MINUTE        = 57568
	HOUR_SECOND        = 57567
	ID                 = 57347
	IDIV               = 57629
	IF                 = 57501
	IGNORE             = 57500
	IN                 = 57620
	INDEX              = 57498
	INDEXES            = 57490
	INNER              = 57592
	INSERT             = 57388
	INT                = 57436
	INTEGER            = 57434
	INTERSECT          = 57586
	INTERVAL           = 57633
	INTO               = 57389
	IS                 = 57616
	ISOLATION          = 57503
	JOIN               = 57588
	JSON               = 57534
	KEY                = 57602
	KEYS               = 57491
	KILL               = 57535
	LBRACE             = 57354
	LE                 = 57613
	LEADING            = 57454
	LEFT               = 57590
	LEVEL              = 57504
	LEX_ERROR          = 57346
	LIKE               = 57617
	LIMIT              = 57368
	LOCAL              = 57509
	LOCK               = 57401
	LOGS               = 57461
	LONGTEXT           = 57447
	LOW_PRIORITY       = 57508
	LPAREN             = 57352
	LT                 = 57614
	MASTER             = 57460
	MEDIUMBLOB         = 57448
	MEDIUMTEXT         = 57449
	MICROSECOND        = 57562
	MINUS              = 57584
	MINUTE             = 57560
	MINUTE_MICROSECOND = 57564
	MINUTE_SECOND      = 57565
	MOD                = 57627
	MODIFY             = 57382
	MONTH              = 57556
	MUTEX              = 57472
	NAMES              = 57514
	NATURAL            = 57597
	NCHAR              = 57413
	NE                 = 57615
	NOT                = 57603
	NULL               = 57395
	NULL_SAFE_EQUAL    = 57610
	NUMBER             = 57349
	NUMERIC            = 57439
	OBJECT_ID          = 57415
	OFF                = 57547
	OFFSET             = 57369
	OJ                 = 57494
	ON                 = 57598
	ONLY               = 57507
	OPEN               = 57479
	OR                 = 57599
	ORDER              = 57366
	OUTER              = 57593
	PARTITIONS         = 57531
	PLUGINS            = 57480
	PLUS               = 57624
	PRECISION          = 57392
	PRIMARY            = 57549
	PRIVILEGES         = 57481
	PROCEDURE          = 57466
	PROCESSLIST        = 57524
	PROFILE            = 57482
	PROFILES           = 57483
	PROXY              = 57519
	QUARTER            = 57555
	QUERY              = 57539
	RBRACE             = 57355
	READ               = 57505
	RECURSIVE          = 57377
	REGEXP             = 57618
	RELAYLOG           = 57484
	RENAME             = 57383
	REPEATABLE         = 57510
	RESTRICT           = 57543
	RIGHT              = 57591
	RLIKE              = 57619
	ROW                = 57423
	RPAREN             = 57353
	SAMPLE             = 57537
	SCHEMA             = 57463
	SCHEMAS            = 57492
	SECOND             = 57561
	SECOND_MICROSECOND = 57563
	SELECT             = 57357
	SEPARATOR          = 57378
	SERIAL             = 57451
	SERIALIZABLE       = 57513
	SESSION            = 57540
	SET                = 57360
	SHOW               = 57361
	SIGNED             = 57427
	SLAVE              = 57485
	SMALLINT           = 57452
	SOME               = 57371
	SQL_BIGINT         = 57429
	SQL_DATE           = 57431
	SQL_DOUBLE         = 57433
	SQL_TIMESTAMP      = 57432
	SQL_TSI_DAY        = 57578
	SQL_TSI_HOUR       = 57579
	SQL_TSI_MINUTE     = 57580
	SQL_TSI_MONTH      = 57576
	SQL_TSI_QUARTER    = 57575
	SQL_TSI_SECOND     = 57581
	SQL_TSI_WEEK       = 57577
	SQL_TSI_YEAR       = 57574
	SQL_VARCHAR        = 57430
	STATUS             = 57525
	STORAGE            = 57474
	STRAIGHT_JOIN      = 57589
	STRING             = 57348
	SUB                = 57625
	SUBDATE            = 57422
	SUBSTR             = 57458
	SUBSTRING          = 57457
	TABLE              = 57496
	TABLES             = 57518
	TEMPORARY          = 57542
	TEXT               = 57440
	THEN               = 57607
	TILDE              = 57356
	TIME               = 57405
	TIMES              = 57626
	TIMESTAMP          = 57406
	TIMESTAMPADD       = 57416
	TIMESTAMPDIFF      = 57417
	TINYINT            = 57435
	TINYTEXT           = 57450
	TO                 = 57385
	TRADITIONAL        = 57533
	TRAILING           = 57455
	TRANSACTION        = 57502
	TRIGGER            = 57469
	TRIGGERS           = 57487
	TRIM               = 57456
	TRUE               = 57373
	UNARY              = 57631
	UNCOMMITTED        = 57512
	UNION              = 57583
	UNIQUE             = 57548
	UNKNOWN            = 57375
	UNLOCK             = 57402
	UNSIGNED           = 57428
	UPDATE             = 57362
	USE                = 57595
	USER               = 57470
	USING              = 57545
	UTC_DATE           = 57410
	UTC_TIMESTAMP      = 57409
	VALUE              = 57398
	VALUES             = 57399
	VALUE_ARG          = 57350
	VARCHAR            = 57441
	VARIABLES          = 57520
	VIEW               = 57499
	WARNINGS           = 57488
	WEEK               = 57557
	WHEN               = 57606
	WHERE              = 57363
	WITH               = 57376
	WRITE              = 57506
	XOR                = 57600
	YEAR               = 57554
	YEAR_MONTH         = 57573
	yyErrCode          = 57345

	yyMaxDepth = 200
	yyTabOfs   = -742
)

var (
	yyPrec = map[int]int{
		YEAR:               0,
		QUARTER:            0,
		MONTH:              0,
		WEEK:               0,
		DAY:                0,
		HOUR:               0,
		MINUTE:             0,
		SECOND:             0,
		MICROSECOND:        0,
		SECOND_MICROSECOND: 1,
		MINUTE_MICROSECOND: 1,
		MINUTE_SECOND:      1,
		HOUR_MICROSECOND:   1,
		HOUR_SECOND:        1,
		HOUR_MINUTE:        1,
		DAY_MICROSECOND:    1,
		DAY_SECOND:         1,
		DAY_MINUTE:         1,
		DAY_HOUR:           1,
		YEAR_MONTH:         1,
		SQL_TSI_YEAR:       2,
		SQL_TSI_QUARTER:    2,
		SQL_TSI_MONTH:      2,
		SQL_TSI_WEEK:       2,
		SQL_TSI_DAY:        2,
		SQL_TSI_HOUR:       2,
		SQL_TSI_MINUTE:     2,
		SQL_TSI_SECOND:     2,
		FROM:               3,
		UNION:              4,
		MINUS:              4,
		EXCEPT:             4,
		INTERSECT:          4,
		COMMA:              5,
		JOIN:               6,
		STRAIGHT_JOIN:      6,
		LEFT:               6,
		RIGHT:              6,
		INNER:              6,
		OUTER:              6,
		CROSS:              6,
		USE:                6,
		FORCE:              6,
		NATURAL:            7,
		ON:                 8,
		OR:                 9,
		XOR:                10,
		AND:                11,
		KEY:                12,
		NOT:                13,
		BETWEEN:            14,
		CASE:               14,
		WHEN:               14,
		THEN:               14,
		ELSE:               14,
		EQ:                 15,
		NULL_SAFE_EQUAL:    15,
		GE:                 15,
		GT:                 15,
		LE:                 15,
		LT:                 15,
		NE:                 15,
		IS:                 15,
		LIKE:               15,
		REGEXP:             15,
		RLIKE:              15,
		IN:                 15,
		BIT_AND:            16,
		BIT_OR:             16,
		CARET:              16,
		PLUS:               17,
		SUB:                17,
		TIMES:              18,
		MOD:                18,
		DIV:                18,
		IDIV:               18,
		DOT:                19,
		UNARY:              20,
		END:                21,
		INTERVAL:           22,
	}

	yyXLAT = map[int]int{
		57386: 0,   // COMMENT_KWD (619x)
		57344: 1,   // $end (604x)
		57546: 2,   // AUTO_INCREMENT (601x)
		57347: 3,   // ID (596x)
		57554: 4,   // YEAR (596x)
		57558: 5,   // DAY (594x)
		57559: 6,   // HOUR (594x)
		57562: 7,   // MICROSECOND (594x)
		57560: 8,   // MINUTE (594x)
		57556: 9,   // MONTH (594x)
		57555: 10,  // QUARTER (594x)
		57561: 11,  // SECOND (594x)
		57557: 12,  // WEEK (594x)
		57578: 13,  // SQL_TSI_DAY (593x)
		57579: 14,  // SQL_TSI_HOUR (593x)
		57580: 15,  // SQL_TSI_MINUTE (593x)
		57576: 16,  // SQL_TSI_MONTH (593x)
		57575: 17,  // SQL_TSI_QUARTER (593x)
		57581: 18,  // SQL_TSI_SECOND (593x)
		57577: 19,  // SQL_TSI_WEEK (593x)
		57574: 20,  // SQL_TSI_YEAR (593x)
		57526: 21,  // CHARSET (569x)
		57349: 22,  // NUMBER (569x)
		57471: 23,  // ENGINE (568x)
		57353: 24,  // RPAREN (568x)
		57525: 25,  // STATUS (567x)
		57403: 26,  // DATE (565x)
		57518: 27,  // TABLES (563x)
		57405: 28,  // TIME (563x)
		57406: 29,  // TIMESTAMP (563x)
		57375: 30,  // UNKNOWN (562x)
		57587: 31,  // COMMA (561x)
		57404: 32,  // DATETIME (561x)
		57411: 33,  // DECIMAL (561x)
		57412: 34,  // FLOAT (561x)
		57398: 35,  // VALUE (561x)
		57520: 36,  // VARIABLES (561x)
		57522: 37,  // COLUMNS (560x)
		57473: 38,  // ENGINES (560x)
		57468: 39,  // EVENTS (560x)
		57461: 40,  // LOGS (560x)
		57524: 41,  // PROCESSLIST (560x)
		57477: 42,  // CODE (559x)
		57476: 43,  // COUNT (559x)
		57553: 44,  // DISABLE (559x)
		57552: 45,  // ENABLE (559x)
		57475: 46,  // ERRORS (559x)
		57465: 47,  // FUNCTION (559x)
		57503: 48,  // ISOLATION (559x)
		57413: 49,  // NCHAR (559x)
		57415: 50,  // OBJECT_ID (559x)
		57423: 51,  // ROW (559x)
		57427: 52,  // SIGNED (559x)
		57542: 53,  // TEMPORARY (559x)
		57435: 54,  // TINYINT (559x)
		57470: 55,  // USER (559x)
		57488: 56,  // WARNINGS (559x)
		57467: 57,  // BINLOG (558x)
		57443: 58,  // BIT (558x)
		57444: 59,  // BLOB (558x)
		57445: 60,  // BOOL (558x)
		57550: 61,  // BTREE (558x)
		57489: 62,  // CHANNEL (558x)
		57523: 63,  // COLLATION (558x)
		57511: 64,  // COMMITTED (558x)
		57538: 65,  // CONNECTION (558x)
		57527: 66,  // DBS (558x)
		57446: 67,  // ENUM (558x)
		57464: 68,  // EVENT (558x)
		57521: 69,  // FULL (558x)
		57478: 70,  // GRANTS (558x)
		57551: 71,  // HASH (558x)
		57486: 72,  // HOSTS (558x)
		57490: 73,  // INDEXES (558x)
		57534: 74,  // JSON (558x)
		57504: 75,  // LEVEL (558x)
		57509: 76,  // LOCAL (558x)
		57447: 77,  // LONGTEXT (558x)
		57460: 78,  // MASTER (558x)
		57448: 79,  // MEDIUMBLOB (558x)
		57449: 80,  // MEDIUMTEXT (558x)
		57472: 81,  // MUTEX (558x)
		57369: 82,  // OFFSET (558x)
		57507: 83,  // ONLY (558x)
		57479: 84,  // OPEN (558x)
		57480: 85,  // PLUGINS (558x)
		57481: 86,  // PRIVILEGES (558x)
		57482: 87,  // PROFILE (558x)
		57483: 88,  // PROFILES (558x)
		57519: 89,  // PROXY (558x)
		57484: 90,  // RELAYLOG (558x)
		57510: 91,  // REPEATABLE (558x)
		57451: 92,  // SERIAL (558x)
		57513: 93,  // SERIALIZABLE (558x)
		57485: 94,  // SLAVE (558x)
		57452: 95,  // SMALLINT (558x)
		57474: 96,  // STORAGE (558x)
		57416: 97,  // TIMESTAMPADD (558x)
		57417: 98,  // TIMESTAMPDIFF (558x)
		57502: 99,  // TRANSACTION (558x)
		57487: 100, // TRIGGERS (558x)
		57512: 101, // UNCOMMITTED (558x)
		57499: 102, // VIEW (558x)
		57372: 103, // ANY (557x)
		57530: 104, // EXTENDED (557x)
		57532: 105, // FORMAT (557x)
		57514: 106, // NAMES (557x)
		57531: 107, // PARTITIONS (557x)
		57539: 108, // QUERY (557x)
		57371: 109, // SOME (557x)
		57348: 110, // STRING (481x)
		57590: 111, // LEFT (479x)
		57591: 112, // RIGHT (479x)
		57624: 113, // PLUS (431x)
		57352: 114, // LPAREN (424x)
		57627: 115, // MOD (414x)
		57603: 116, // NOT (413x)
		57625: 117, // SUB (413x)
		57585: 118, // EXCEPT (394x)
		57586: 119, // INTERSECT (394x)
		57584: 120, // MINUS (394x)
		57583: 121, // UNION (394x)
		57354: 122, // LBRACE (383x)
		57370: 123, // FOR (381x)
		57368: 124, // LIMIT (380x)
		57363: 125, // WHERE (367x)
		57602: 126, // KEY (364x)
		57392: 127, // PRECISION (362x)
		57401: 128, // LOCK (361x)
		57366: 129, // ORDER (359x)
		57548: 130, // UNIQUE (358x)
		57549: 131, // PRIMARY (357x)
		57588: 132, // JOIN (348x)
		57601: 133, // AND (347x)
		57572: 134, // DAY_HOUR (346x)
		57569: 135, // DAY_MICROSECOND (346x)
		57571: 136, // DAY_MINUTE (346x)
		57570: 137, // DAY_SECOND (346x)
		57566: 138, // HOUR_MICROSECOND (346x)
		57568: 139, // HOUR_MINUTE (346x)
		57567: 140, // HOUR_SECOND (346x)
		57564: 141, // MINUTE_MICROSECOND (346x)
		57565: 142, // MINUTE_SECOND (346x)
		57563: 143, // SECOND_MICROSECOND (346x)
		57573: 144, // YEAR_MONTH (346x)
		57599: 145, // OR (345x)
		57589: 146, // STRAIGHT_JOIN (345x)
		57626: 147, // TIMES (345x)
		57600: 148, // XOR (345x)
		57582: 149, // FROM (344x)
		57617: 150, // LIKE (344x)
		57365: 151, // HAVING (343x)
		57364: 152, // GROUP (338x)
		57594: 153, // CROSS (337x)
		57592: 154, // INNER (337x)
		57597: 155, // NATURAL (337x)
		57545: 156, // USING (337x)
		57355: 157, // RBRACE (335x)
		57598: 158, // ON (332x)
		57378: 159, // SEPARATOR (331x)
		57393: 160, // AS (325x)
		57609: 161, // EQ (315x)
		57606: 162, // WHEN (315x)
		57632: 163, // END (314x)
		57608: 164, // ELSE (312x)
		57397: 165, // DESC (311x)
		57396: 166, // ASC (310x)
		57607: 167, // THEN (309x)
		57611: 168, // GE (307x)
		57612: 169, // GT (307x)
		57616: 170, // IS (307x)
		57613: 171, // LE (307x)
		57614: 172, // LT (307x)
		57615: 173, // NE (307x)
		57610: 174, // NULL_SAFE_EQUAL (307x)
		57620: 175, // IN (283x)
		57621: 176, // BIT_AND (275x)
		57622: 177, // BIT_OR (275x)
		57623: 178, // CARET (275x)
		57628: 179, // DIV (275x)
		57629: 180, // IDIV (275x)
		57604: 181, // BETWEEN (270x)
		57618: 182, // REGEXP (270x)
		57619: 183, // RLIKE (270x)
		57426: 184, // CHAR (255x)
		57399: 185, // VALUES (255x)
		57495: 186, // ESCAPE (221x)
		57630: 187, // DOT (189x)
		57395: 188, // NULL (184x)
		57721: 189, // keyword_as_id (176x)
		57759: 190, // sql_id (176x)
		57400: 191, // DEFAULT (163x)
		57501: 192, // IF (147x)
		57374: 193, // FALSE (145x)
		57373: 194, // TRUE (145x)
		57462: 195, // DATABASE (144x)
		57394: 196, // EXISTS (143x)
		57633: 197, // INTERVAL (143x)
		57350: 198, // VALUE_ARG (143x)
		57388: 199, // INSERT (142x)
		57463: 200, // SCHEMA (142x)
		57420: 201, // ADDDATE (141x)
		57425: 202, // CAST (141x)
		57424: 203, // CONVERT (141x)
		57408: 204, // CURRENT_DATE (141x)
		57407: 205, // CURRENT_TIMESTAMP (141x)
		57419: 206, // DATE_ADD (141x)
		57421: 207, // DATE_SUB (141x)
		57418: 208, // EXTRACT (141x)
		57414: 209, // GROUP_CONCAT (141x)
		57422: 210, // SUBDATE (141x)
		57458: 211, // SUBSTR (141x)
		57457: 212, // SUBSTRING (141x)
		57456: 213, // TRIM (141x)
		57410: 214, // UTC_DATE (141x)
		57409: 215, // UTC_TIMESTAMP (141x)
		57605: 216, // CASE (140x)
		57356: 217, // TILDE (140x)
		57765: 218, // subquery (137x)
		57656: 219, // column_name (132x)
		57646: 220, // boolean_value (124x)
		57595: 221, // USE (122x)
		57788: 222, // value (122x)
		57515: 223, // CHARACTER (121x)
		57596: 224, // FORCE (121x)
		57500: 225, // IGNORE (121x)
		57505: 226, // READ (121x)
		57782: 227, // tuple (121x)
		57695: 228, // func_expr (120x)
		57696: 229, // func_expr_conflict (120x)
		57697: 230, // func_expr_generic (120x)
		57698: 231, // func_expr_reserved_keyword (120x)
		57699: 232, // func_expr_unconventional (120x)
		57766: 233, // substr (120x)
		57506: 234, // WRITE (120x)
		57649: 235, // case_expression (119x)
		57757: 236, // simple_expr (119x)
		57783: 237, // unary_operator (119x)
		57459: 238, // BINARY (118x)
		57434: 239, // INTEGER (117x)
		57508: 240, // LOW_PRIORITY (117x)
		57360: 241, // SET (116x)
		57358: 242, // DROP (114x)
		57383: 243, // RENAME (114x)
		57437: 244, // BIGINT (113x)
		57442: 245, // BOOLEAN (113x)
		57381: 246, // CHANGE (113x)
		57438: 247, // DOUBLE (113x)
		57436: 248, // INT (113x)
		57382: 249, // MODIFY (113x)
		57439: 250, // NUMERIC (113x)
		57440: 251, // TEXT (113x)
		57385: 252, // TO (113x)
		57441: 253, // VARCHAR (113x)
		57644: 254, // bit_expr (112x)
		57544: 255, // CASCADE (112x)
		57543: 256, // RESTRICT (112x)
		57450: 257, // TINYTEXT (112x)
		57516: 258, // COLLATE (111x)
		57738: 259, // predicate (96x)
		57645: 260, // bool_pri (87x)
		57684: 261, // expression (87x)
		57743: 262, // select_expression (51x)
		57357: 263, // SELECT (35x)
		57376: 264, // WITH (35x)
		57744: 265, // select_expression_list (32x)
		57771: 266, // table_name (19x)
		57725: 267, // like_or_where_opt (17x)
		57731: 268, // non_derived_subquery (15x)
		57745: 269, // select_statement (15x)
		57797: 270, // with_statement (15x)
		57391: 271, // DISTINCT (11x)
		57761: 272, // sql_time_interval (11x)
		57777: 273, // time_interval (11x)
		57784: 274, // union_op (11x)
		57707: 275, // in_or_from (10x)
		57498: 276, // INDEX (10x)
		57715: 277, // interval_unit (9x)
		57726: 278, // limit_opt (9x)
		57752: 279, // show_from_in (9x)
		57762: 280, // sql_time_unit (9x)
		57496: 281, // TABLE (9x)
		57390: 282, // ALL (8x)
		57717: 283, // join_expression (8x)
		57758: 284, // simple_table_expression (8x)
		57767: 285, // table_expression (8x)
		57636: 286, // all_any_some (7x)
		57351: 287, // COMMENT (6x)
		57669: 288, // database_name (6x)
		57718: 289, // join_type (6x)
		57641: 290, // as_opt (5x)
		57685: 291, // expression_list (5x)
		57491: 292, // KEYS (5x)
		57753: 293, // show_from_in_opt (5x)
		57795: 294, // where_expression_opt (5x)
		57677: 295, // equal_opt (4x)
		57724: 296, // like_escape_opt (4x)
		57734: 297, // optional_parens (4x)
		57736: 298, // order_by_opt (4x)
		57593: 299, // OUTER (4x)
		57634: 300, // $@1 (3x)
		57384: 301, // COLUMN (3x)
		57657: 302, // column_opt (3x)
		57661: 303, // comment_opt (3x)
		57666: 304, // cte (3x)
		57693: 305, // from_opt (3x)
		57709: 306, // index_list (3x)
		57749: 307, // set_spec (3x)
		57760: 308, // sql_id_or_string (3x)
		57791: 309, // value_or_default (3x)
		57637: 310, // alter_spec (2x)
		57367: 311, // BY (2x)
		57650: 312, // character_set (2x)
		57654: 313, // column_expression_list (2x)
		57359: 314, // CREATE (2x)
		57663: 315, // create_definition (2x)
		57667: 316, // cte_list (2x)
		57497: 317, // DUAL (2x)
		57674: 318, // dual_table (2x)
		57682: 319, // explainable_stmt (2x)
		57387: 320, // FULLTEXT (2x)
		57694: 321, // fulltext_opt (2x)
		57541: 322, // GLOBAL (2x)
		57700: 323, // group_by_opt (2x)
		57701: 324, // having_opt (2x)
		57702: 325, // if_exists_opt (2x)
		57703: 326, // if_not_exists_opt (2x)
		57704: 327, // if_not_exists_opt_string (2x)
		57706: 328, // in_opt (2x)
		57719: 329, // key_part (2x)
		57728: 330, // lock_opt (2x)
		57735: 331, // order (2x)
		57466: 332, // PROCEDURE (2x)
		57740: 333, // query_globals_opt (2x)
		57742: 334, // scope_modifier_opt (2x)
		57540: 335, // SESSION (2x)
		57429: 336, // SQL_BIGINT (2x)
		57431: 337, // SQL_DATE (2x)
		57433: 338, // SQL_DOUBLE (2x)
		57432: 339, // SQL_TIMESTAMP (2x)
		57763: 340, // sql_types (2x)
		57430: 341, // SQL_VARCHAR (2x)
		57768: 342, // table_expression_list (2x)
		57769: 343, // table_lock (2x)
		57774: 344, // table_rename (2x)
		57776: 345, // temporary_opt (2x)
		57779: 346, // transaction_characteristic (2x)
		57428: 347, // UNSIGNED (2x)
		57789: 348, // value_list (2x)
		57793: 349, // when_expression (2x)
		57379: 350, // ALTER (1x)
		57638: 351, // alter_spec_list (1x)
		57639: 352, // alter_statement (1x)
		57640: 353, // any_command (1x)
		57642: 354, // asc_desc_opt (1x)
		57643: 355, // auto_increment_opt (1x)
		57453: 356, // BOTH (1x)
		57647: 357, // both_leading_trailing_opt (1x)
		57648: 358, // cascade_or_restrict_opt (1x)
		57651: 359, // column_comment_opt (1x)
		57652: 360, // column_data_type (1x)
		57653: 361, // column_definition (1x)
		57655: 362, // column_list (1x)
		57658: 363, // comma_opt (1x)
		57659: 364, // command (1x)
		57660: 365, // comment_list (1x)
		57662: 366, // create_database_statement (1x)
		57664: 367, // create_definition_list (1x)
		57665: 368, // create_table_statement (1x)
		57668: 369, // data_type (1x)
		57517: 370, // DATABASES (1x)
		57670: 371, // default_opt (1x)
		57529: 372, // DESCRIBE (1x)
		57671: 373, // distinct_opt (1x)
		57672: 374, // drop_database_statement (1x)
		57673: 375, // drop_table_statement (1x)
		57675: 376, // else_expression_opt (1x)
		57676: 377, // enum_column_data_type (1x)
		57528: 378, // EXPLAIN (1x)
		57678: 379, // explain_alias (1x)
		57679: 380, // explain_column_name (1x)
		57680: 381, // explain_statement (1x)
		57681: 382, // explain_type (1x)
		57683: 383, // explicit_scope_modifier_opt (1x)
		57686: 384, // expression_opt (1x)
		57687: 385, // float_column_data_type (1x)
		57688: 386, // float_width_opt (1x)
		57536: 387, // FLUSH (1x)
		57689: 388, // flush_statement (1x)
		57493: 389, // FN (1x)
		57690: 390, // for_channel_opt (1x)
		57691: 391, // for_user_opt (1x)
		57692: 392, // format_name (1x)
		57705: 393, // ignored_statement (1x)
		57708: 394, // index_hint_list (1x)
		57710: 395, // index_name_opt (1x)
		57711: 396, // index_or_key (1x)
		57712: 397, // index_type_opt (1x)
		57713: 398, // insert_columns_opt (1x)
		57714: 399, // insert_statement (1x)
		57389: 400, // INTO (1x)
		57716: 401, // into_opt (1x)
		57720: 402, // key_part_list (1x)
		57535: 403, // KILL (1x)
		57722: 404, // kill_modifier (1x)
		57723: 405, // kill_statement (1x)
		57454: 406, // LEADING (1x)
		57727: 407, // local_opt (1x)
		57729: 408, // lock_type (1x)
		57730: 409, // low_priority_opt (1x)
		57732: 410, // normal_column_data_type (1x)
		57733: 411, // null_opt (1x)
		57547: 412, // OFF (1x)
		57494: 413, // OJ (1x)
		57737: 414, // order_list (1x)
		57739: 415, // primary_key_opt (1x)
		57377: 416, // RECURSIVE (1x)
		57741: 417, // rename_statement (1x)
		57537: 418, // SAMPLE (1x)
		57492: 419, // SCHEMAS (1x)
		57746: 420, // select_statement_with_paren_order_limit (1x)
		57747: 421, // separator_opt (1x)
		57748: 422, // set_expr (1x)
		57750: 423, // set_spec_list (1x)
		57751: 424, // set_statement (1x)
		57361: 425, // SHOW (1x)
		57754: 426, // show_full (1x)
		57755: 427, // show_statement (1x)
		57756: 428, // simple_column_data_type (1x)
		57764: 429, // storage_opt (1x)
		57770: 430, // table_lock_list (1x)
		57772: 431, // table_option (1x)
		57773: 432, // table_options_opt (1x)
		57775: 433, // table_rename_list (1x)
		57778: 434, // to_as_opt (1x)
		57533: 435, // TRADITIONAL (1x)
		57455: 436, // TRAILING (1x)
		57780: 437, // transaction_characteristics (1x)
		57781: 438, // transaction_level (1x)
		57469: 439, // TRIGGER (1x)
		57785: 440, // unique_key_opt (1x)
		57786: 441, // unique_opt (1x)
		57402: 442, // UNLOCK (1x)
		57362: 443, // UPDATE (1x)
		57787: 444, // use_statement (1x)
		57790: 445, // value_list_list (1x)
		57792: 446, // value_or_values (1x)
		57794: 447, // when_expression_list (1x)
		57796: 448, // width_opt (1x)
		57635: 449, // $default (0x)
		57380: 450, // ADD (0x)
		57345: 451, // error (0x)
		57346: 452, // LEX_ERROR (0x)
		57631: 453, // UNARY (0x)
	}

	yySymNames = []string{
		"COMMENT_KWD",
		"$end",
		"AUTO_INCREMENT",
		"ID",
		"YEAR",
		"DAY",
		"HOUR",
		"MICROSECOND",
		"MINUTE",
		"MONTH",
		"QUARTER",
		"SECOND",
		"WEEK",
		"SQL_TSI_DAY",
		"SQL_TSI_HOUR",
		"SQL_TSI_MINUTE",
		"SQL_TSI_MONTH",
		"SQL_TSI_QUARTER",
		"SQL_TSI_SECOND",
		"SQL_TSI_WEEK",
		"SQL_TSI_YEAR",
		"CHARSET",
		"NUMBER",
		"ENGINE",
		"RPAREN",
		"STATUS",
		"DATE",
		"TABLES",
		"TIME",
		"TIMESTAMP",
		"UNKNOWN",
		"COMMA",
		"DATETIME",
		"DECIMAL",
		"FLOAT",
		"VALUE",
		"VARIABLES",
		"COLUMNS",
		"ENGINES",
		"EVENTS",
		"LOGS",
		"PROCESSLIST",
		"CODE",
		"COUNT",
		"DISABLE",
		"ENABLE",
		"ERRORS",
		"FUNCTION",
		"ISOLATION",
		"NCHAR",
		"OBJECT_ID",
		"ROW",
		"SIGNED",
		"TEMPORARY",
		"TINYINT",
		"USER",
		"WARNINGS",
		"BINLOG",
		"BIT",
		"BLOB",
		"BOOL",
		"BTREE",
		"CHANNEL",
		"COLLATION",
		"COMMITTED",
		"CONNECTION",
		"DBS",
		"ENUM",
		"EVENT",
		"FULL",
		"GRANTS",
		"HASH",
		"HOSTS",
		"INDEXES",
		"JSON",
		"LEVEL",
		"LOCAL",
		"LONGTEXT",
		"MASTER",
		"MEDIUMBLOB",
		"MEDIUMTEXT",
		"MUTEX",
		"OFFSET",
		"ONLY",
		"OPEN",
		"PLUGINS",
		"PRIVILEGES",
		"PROFILE",
		"PROFILES",
		"PROXY",
		"RELAYLOG",
		"REPEATABLE",
		"SERIAL",
		"SERIALIZABLE",
		"SLAVE",
		"SMALLINT",
		"STORAGE",
		"TIMESTAMPADD",
		"TIMESTAMPDIFF",
		"TRANSACTION",
		"TRIGGERS",
		"UNCOMMITTED",
		"VIEW",
		"ANY",
		"EXTENDED",
		"FORMAT",
		"NAMES",
		"PARTITIONS",
		"QUERY",
		"SOME",
		"STRING",
		"LEFT",
		"RIGHT",
		"PLUS",
		"LPAREN",
		"MOD",
		"NOT",
		"SUB",
		"EXCEPT",
		"INTERSECT",
		"MINUS",
		"UNION",
		"LBRACE",
		"FOR",
		"LIMIT",
		"WHERE",
		"KEY",
		"PRECISION",
		"LOCK",
		"ORDER",
		"UNIQUE",
		"PRIMARY",
		"JOIN",
		"AND",
		"DAY_HOUR",
		"DAY_MICROSECOND",
		"DAY_MINUTE",
		"DAY_SECOND",
		"HOUR_MICROSECOND",
		"HOUR_MINUTE",
		"HOUR_SECOND",
		"MINUTE_MICROSECOND",
		"MINUTE_SECOND",
		"SECOND_MICROSECOND",
		"YEAR_MONTH",
		"OR",
		"STRAIGHT_JOIN",
		"TIMES",
		"XOR",
		"FROM",
		"LIKE",
		"HAVING",
		"GROUP",
		"CROSS",
		"INNER",
		"NATURAL",
		"USING",
		"RBRACE",
		"ON",
		"SEPARATOR",
		"AS",
		"EQ",
		"WHEN",
		"END",
		"ELSE",
		"DESC",
		"ASC",
		"THEN",
		"GE",
		"GT",
		"IS",
		"LE",
		"LT",
		"NE",
		"NULL_SAFE_EQUAL",
		"IN",
		"BIT_AND",
		"BIT_OR",
		"CARET",
		"DIV",
		"IDIV",
		"BETWEEN",
		"REGEXP",
		"RLIKE",
		"CHAR",
		"VALUES",
		"ESCAPE",
		"DOT",
		"NULL",
		"keyword_as_id",
		"sql_id",
		"DEFAULT",
		"IF",
		"FALSE",
		"TRUE",
		"DATABASE",
		"EXISTS",
		"INTERVAL",
		"VALUE_ARG",
		"INSERT",
		"SCHEMA",
		"ADDDATE",
		"CAST",
		"CONVERT",
		"CURRENT_DATE",
		"CURRENT_TIMESTAMP",
		"DATE_ADD",
		"DATE_SUB",
		"EXTRACT",
		"GROUP_CONCAT",
		"SUBDATE",
		"SUBSTR",
		"SUBSTRING",
		"TRIM",
		"UTC_DATE",
		"UTC_TIMESTAMP",
		"CASE",
		"TILDE",
		"subquery",
		"column_name",
		"boolean_value",
		"USE",
		"value",
		"CHARACTER",
		"FORCE",
		"IGNORE",
		"READ",
		"tuple",
		"func_expr",
		"func_expr_conflict",
		"func_expr_generic",
		"func_expr_reserved_keyword",
		"func_expr_unconventional",
		"substr",
		"WRITE",
		"case_expression",
		"simple_expr",
		"unary_operator",
		"BINARY",
		"INTEGER",
		"LOW_PRIORITY",
		"SET",
		"DROP",
		"RENAME",
		"BIGINT",
		"BOOLEAN",
		"CHANGE",
		"DOUBLE",
		"INT",
		"MODIFY",
		"NUMERIC",
		"TEXT",
		"TO",
		"VARCHAR",
		"bit_expr",
		"CASCADE",
		"RESTRICT",
		"TINYTEXT",
		"COLLATE",
		"predicate",
		"bool_pri",
		"expression",
		"select_expression",
		"SELECT",
		"WITH",
		"select_expression_list",
		"table_name",
		"like_or_where_opt",
		"non_derived_subquery",
		"select_statement",
		"with_statement",
		"DISTINCT",
		"sql_time_interval",
		"time_interval",
		"union_op",
		"in_or_from",
		"INDEX",
		"interval_unit",
		"limit_opt",
		"show_from_in",
		"sql_time_unit",
		"TABLE",
		"ALL",
		"join_expression",
		"simple_table_expression",
		"table_expression",
		"all_any_some",
		"COMMENT",
		"database_name",
		"join_type",
		"as_opt",
		"expression_list",
		"KEYS",
		"show_from_in_opt",
		"where_expression_opt",
		"equal_opt",
		"like_escape_opt",
		"optional_parens",
		"order_by_opt",
		"OUTER",
		"$@1",
		"COLUMN",
		"column_opt",
		"comment_opt",
		"cte",
		"from_opt",
		"index_list",
		"set_spec",
		"sql_id_or_string",
		"value_or_default",
		"alter_spec",
		"BY",
		"character_set",
		"column_expression_list",
		"CREATE",
		"create_definition",
		"cte_list",
		"DUAL",
		"dual_table",
		"explainable_stmt",
		"FULLTEXT",
		"fulltext_opt",
		"GLOBAL",
		"group_by_opt",
		"having_opt",
		"if_exists_opt",
		"if_not_exists_opt",
		"if_not_exists_opt_string",
		"in_opt",
		"key_part",
		"lock_opt",
		"order",
		"PROCEDURE",
		"query_globals_opt",
		"scope_modifier_opt",
		"SESSION",
		"SQL_BIGINT",
		"SQL_DATE",
		"SQL_DOUBLE",
		"SQL_TIMESTAMP",
		"sql_types",
		"SQL_VARCHAR",
		"table_expression_list",
		"table_lock",
		"table_rename",
		"temporary_opt",
		"transaction_characteristic",
		"UNSIGNED",
		"value_list",
		"when_expression",
		"ALTER",
		"alter_spec_list",
		"alter_statement",
		"any_command",
		"asc_desc_opt",
		"auto_increment_opt",
		"BOTH",
		"both_leading_trailing_opt",
		"cascade_or_restrict_opt",
		"column_comment_opt",
		"column_data_type",
		"column_definition",
		"column_list",
		"comma_opt",
		"command",
		"comment_list",
		"create_database_statement",
		"create_definition_list",
		"create_table_statement",
		"data_type",
		"DATABASES",
		"default_opt",
		"DESCRIBE",
		"distinct_opt",
		"drop_database_statement",
		"drop_table_statement",
		"else_expression_opt",
		"enum_column_data_type",
		"EXPLAIN",
		"explain_alias",
		"explain_column_name",
		"explain_statement",
		"explain_type",
		"explicit_scope_modifier_opt",
		"expression_opt",
		"float_column_data_type",
		"float_width_opt",
		"FLUSH",
		"flush_statement",
		"FN",
		"for_channel_opt",
		"for_user_opt",
		"format_name",
		"ignored_statement",
		"index_hint_list",
		"index_name_opt",
		"index_or_key",
		"index_type_opt",
		"insert_columns_opt",
		"insert_statement",
		"INTO",
		"into_opt",
		"key_part_list",
		"KILL",
		"kill_modifier",
		"kill_statement",
		"LEADING",
		"local_opt",
		"lock_type",
		"low_priority_opt",
		"normal_column_data_type",
		"null_opt",
		"OFF",
		"OJ",
		"order_list",
		"primary_key_opt",
		"RECURSIVE",
		"rename_statement",
		"SAMPLE",
		"SCHEMAS",
		"select_statement_with_paren_order_limit",
		"separator_opt",
		"set_expr",
		"set_spec_list",
		"set_statement",
		"SHOW",
		"show_full",
		"show_statement",
		"simple_column_data_type",
		"storage_opt",
		"table_lock_list",
		"table_option",
		"table_options_opt",
		"table_rename_list",
		"to_as_opt",
		"TRADITIONAL",
		"TRAILING",
		"transaction_characteristics",
		"transaction_level",
		"TRIGGER",
		"unique_key_opt",
		"unique_opt",
		"UNLOCK",
		"UPDATE",
		"use_statement",
		"value_list_list",
		"value_or_values",
		"when_expression_list",
		"width_opt",
		"$default",
		"ADD",
		"error",
		"LEX_ERROR",
		"UNARY",
	}

	yyTokenLiteralStrings = map[int]string{}

	yyReductions = map[int]struct{ xsym, components int }{
		0:   {0, 1},
		1:   {353, 1},
		2:   {364, 1},
		3:   {364, 1},
		4:   {364, 1},
		5:   {364, 1},
		6:   {364, 1},
		7:   {364, 1},
		8:   {364, 1},
		9:   {364, 1},
		10:  {364, 1},
		11:  {364, 1},
		12:  {364, 1},
		13:  {364, 1},
		14:  {364, 1},
		15:  {364, 1},
		16:  {364, 1},
		17:  {364, 1},
		18:  {420, 3},
		19:  {269, 1},
		20:  {269, 5},
		21:  {269, 6},
		22:  {269, 12},
		23:  {269, 4},
		24:  {269, 13},
		25:  {269, 3},
		26:  {304, 5},
		27:  {304, 8},
		28:  {316, 1},
		29:  {316, 3},
		30:  {270, 2},
		31:  {270, 3},
		32:  {268, 3},
		33:  {444, 2},
		34:  {424, 3},
		35:  {424, 3},
		36:  {424, 3},
		37:  {424, 5},
		38:  {424, 4},
		39:  {424, 4},
		40:  {423, 1},
		41:  {423, 3},
		42:  {307, 3},
		43:  {307, 3},
		44:  {422, 1},
		45:  {422, 1},
		46:  {422, 1},
		47:  {366, 4},
		48:  {368, 9},
		49:  {432, 0},
		50:  {432, 3},
		51:  {363, 0},
		52:  {363, 1},
		53:  {431, 3},
		54:  {431, 4},
		55:  {431, 3},
		56:  {431, 3},
		57:  {312, 2},
		58:  {312, 1},
		59:  {295, 0},
		60:  {295, 1},
		61:  {367, 1},
		62:  {367, 3},
		63:  {315, 8},
		64:  {315, 8},
		65:  {360, 1},
		66:  {360, 2},
		67:  {360, 2},
		68:  {360, 1},
		69:  {428, 1},
		70:  {428, 1},
		71:  {428, 1},
		72:  {428, 1},
		73:  {428, 1},
		74:  {428, 1},
		75:  {410, 1},
		76:  {410, 1},
		77:  {410, 1},
		78:  {410, 1},
		79:  {410, 1},
		80:  {410, 1},
		81:  {410, 1},
		82:  {410, 1},
		83:  {410, 1},
		84:  {410, 1},
		85:  {410, 1},
		86:  {410, 1},
		87:  {410, 1},
		88:  {410, 1},
		89:  {410, 1},
		90:  {410, 1},
		91:  {385, 1},
		92:  {385, 1},
		93:  {385, 1},
		94:  {385, 1},
		95:  {385, 2},
		96:  {377, 1},
		97:  {377, 1},
		98:  {448, 0},
		99:  {448, 3},
		100: {386, 0},
		101: {386, 5},
		102: {411, 0},
		103: {411, 1},
		104: {411, 2},
		105: {371, 0},
		106: {371, 2},
		107: {371, 2},
		108: {355, 0},
		109: {355, 1},
		110: {440, 0},
		111: {440, 1},
		112: {440, 2},
		113: {415, 0},
		114: {415, 1},
		115: {415, 2},
		116: {359, 0},
		117: {359, 2},
		118: {321, 0},
		119: {321, 1},
		120: {441, 0},
		121: {441, 1},
		122: {396, 1},
		123: {396, 1},
		124: {395, 0},
		125: {395, 1},
		126: {397, 0},
		127: {397, 2},
		128: {397, 2},
		129: {402, 1},
		130: {402, 3},
		131: {329, 1},
		132: {329, 2},
		133: {329, 2},
		134: {374, 4},
		135: {375, 6},
		136: {388, 2},
		137: {388, 2},
		138: {417, 3},
		139: {433, 1},
		140: {433, 3},
		141: {344, 3},
		142: {393, 3},
		143: {393, 2},
		144: {393, 2},
		145: {393, 2},
		146: {430, 1},
		147: {430, 3},
		148: {343, 3},
		149: {408, 2},
		150: {408, 2},
		151: {407, 0},
		152: {407, 1},
		153: {409, 0},
		154: {409, 1},
		155: {399, 6},
		156: {401, 0},
		157: {401, 1},
		158: {398, 0},
		159: {398, 2},
		160: {398, 3},
		161: {362, 3},
		162: {362, 1},
		163: {446, 1},
		164: {446, 1},
		165: {445, 5},
		166: {445, 3},
		167: {348, 3},
		168: {348, 1},
		169: {348, 0},
		170: {309, 1},
		171: {309, 1},
		172: {352, 4},
		173: {352, 5},
		174: {352, 5},
		175: {351, 1},
		176: {351, 3},
		177: {310, 4},
		178: {310, 3},
		179: {310, 4},
		180: {310, 3},
		181: {302, 0},
		182: {302, 1},
		183: {361, 1},
		184: {369, 1},
		185: {369, 1},
		186: {369, 1},
		187: {369, 1},
		188: {369, 1},
		189: {369, 1},
		190: {369, 1},
		191: {369, 1},
		192: {369, 1},
		193: {369, 1},
		194: {369, 1},
		195: {369, 1},
		196: {369, 1},
		197: {369, 1},
		198: {369, 1},
		199: {369, 1},
		200: {369, 1},
		201: {434, 0},
		202: {434, 1},
		203: {434, 1},
		204: {345, 0},
		205: {345, 1},
		206: {325, 0},
		207: {325, 2},
		208: {326, 0},
		209: {326, 3},
		210: {358, 0},
		211: {358, 1},
		212: {358, 1},
		213: {437, 1},
		214: {437, 3},
		215: {346, 3},
		216: {346, 2},
		217: {346, 2},
		218: {438, 2},
		219: {438, 2},
		220: {438, 2},
		221: {438, 1},
		222: {233, 1},
		223: {233, 1},
		224: {275, 1},
		225: {275, 1},
		226: {279, 3},
		227: {279, 4},
		228: {279, 4},
		229: {279, 2},
		230: {293, 0},
		231: {293, 1},
		232: {426, 0},
		233: {426, 1},
		234: {404, 1},
		235: {404, 1},
		236: {334, 0},
		237: {334, 1},
		238: {334, 1},
		239: {383, 1},
		240: {383, 1},
		241: {427, 3},
		242: {427, 3},
		243: {427, 6},
		244: {427, 5},
		245: {427, 5},
		246: {427, 4},
		247: {427, 4},
		248: {427, 4},
		249: {427, 6},
		250: {427, 5},
		251: {427, 4},
		252: {427, 4},
		253: {427, 4},
		254: {427, 4},
		255: {427, 4},
		256: {427, 4},
		257: {427, 3},
		258: {427, 3},
		259: {427, 6},
		260: {427, 4},
		261: {427, 4},
		262: {427, 4},
		263: {427, 3},
		264: {427, 4},
		265: {427, 4},
		266: {427, 4},
		267: {427, 3},
		268: {427, 5},
		269: {427, 2},
		270: {427, 2},
		271: {427, 4},
		272: {427, 4},
		273: {427, 2},
		274: {427, 2},
		275: {427, 6},
		276: {427, 3},
		277: {427, 4},
		278: {427, 5},
		279: {427, 4},
		280: {427, 3},
		281: {427, 6},
		282: {427, 3},
		283: {427, 3},
		284: {427, 3},
		285: {427, 4},
		286: {427, 5},
		287: {427, 5},
		288: {427, 5},
		289: {427, 3},
		290: {427, 4},
		291: {427, 4},
		292: {427, 3},
		293: {427, 3},
		294: {392, 1},
		295: {392, 1},
		296: {382, 1},
		297: {382, 1},
		298: {382, 3},
		299: {382, 0},
		300: {319, 1},
		301: {319, 3},
		302: {379, 1},
		303: {379, 1},
		304: {379, 1},
		305: {288, 1},
		306: {266, 1},
		307: {266, 2},
		308: {266, 3},
		309: {318, 1},
		310: {381, 3},
		311: {381, 3},
		312: {381, 5},
		313: {380, 0},
		314: {380, 1},
		315: {380, 1},
		316: {405, 2},
		317: {405, 3},
		318: {300, 0},
		319: {303, 2},
		320: {365, 0},
		321: {365, 2},
		322: {333, 0},
		323: {333, 1},
		324: {333, 1},
		325: {333, 2},
		326: {333, 2},
		327: {274, 1},
		328: {274, 2},
		329: {274, 1},
		330: {274, 1},
		331: {274, 1},
		332: {373, 0},
		333: {373, 1},
		334: {421, 0},
		335: {421, 2},
		336: {265, 1},
		337: {265, 3},
		338: {262, 1},
		339: {262, 2},
		340: {262, 3},
		341: {262, 3},
		342: {262, 5},
		343: {313, 1},
		344: {313, 3},
		345: {342, 1},
		346: {342, 3},
		347: {285, 3},
		348: {285, 3},
		349: {285, 1},
		350: {285, 4},
		351: {283, 3},
		352: {283, 3},
		353: {283, 5},
		354: {283, 5},
		355: {283, 7},
		356: {290, 0},
		357: {290, 1},
		358: {290, 2},
		359: {290, 1},
		360: {290, 2},
		361: {289, 1},
		362: {289, 2},
		363: {289, 3},
		364: {289, 2},
		365: {289, 3},
		366: {289, 2},
		367: {289, 2},
		368: {289, 2},
		369: {289, 3},
		370: {289, 4},
		371: {289, 3},
		372: {289, 4},
		373: {284, 1},
		374: {284, 1},
		375: {394, 0},
		376: {394, 5},
		377: {394, 5},
		378: {394, 5},
		379: {306, 1},
		380: {306, 3},
		381: {294, 0},
		382: {294, 2},
		383: {267, 0},
		384: {267, 2},
		385: {267, 2},
		386: {328, 0},
		387: {328, 2},
		388: {327, 0},
		389: {327, 3},
		390: {305, 0},
		391: {305, 2},
		392: {390, 0},
		393: {390, 3},
		394: {391, 0},
		395: {391, 2},
		396: {429, 0},
		397: {429, 1},
		398: {286, 1},
		399: {286, 1},
		400: {286, 1},
		401: {227, 6},
		402: {227, 3},
		403: {218, 3},
		404: {291, 1},
		405: {291, 3},
		406: {261, 3},
		407: {261, 3},
		408: {261, 3},
		409: {261, 2},
		410: {261, 3},
		411: {261, 4},
		412: {261, 1},
		413: {260, 3},
		414: {260, 4},
		415: {260, 3},
		416: {260, 4},
		417: {260, 3},
		418: {260, 4},
		419: {260, 3},
		420: {260, 4},
		421: {260, 3},
		422: {260, 4},
		423: {260, 3},
		424: {260, 4},
		425: {260, 3},
		426: {260, 4},
		427: {260, 3},
		428: {260, 4},
		429: {260, 1},
		430: {259, 3},
		431: {259, 3},
		432: {259, 4},
		433: {259, 4},
		434: {259, 5},
		435: {259, 6},
		436: {259, 4},
		437: {259, 5},
		438: {259, 5},
		439: {259, 6},
		440: {259, 3},
		441: {259, 4},
		442: {259, 3},
		443: {259, 4},
		444: {259, 1},
		445: {254, 3},
		446: {254, 3},
		447: {254, 3},
		448: {254, 5},
		449: {254, 5},
		450: {254, 3},
		451: {254, 5},
		452: {254, 3},
		453: {254, 3},
		454: {254, 3},
		455: {254, 3},
		456: {254, 3},
		457: {254, 1},
		458: {236, 1},
		459: {236, 1},
		460: {236, 1},
		461: {236, 1},
		462: {236, 2},
		463: {236, 2},
		464: {236, 1},
		465: {236, 1},
		466: {236, 4},
		467: {236, 4},
		468: {228, 1},
		469: {228, 1},
		470: {228, 1},
		471: {228, 1},
		472: {231, 4},
		473: {231, 6},
		474: {231, 4},
		475: {231, 4},
		476: {231, 4},
		477: {232, 8},
		478: {232, 6},
		479: {232, 6},
		480: {232, 7},
		481: {232, 2},
		482: {232, 2},
		483: {232, 4},
		484: {232, 8},
		485: {232, 8},
		486: {232, 6},
		487: {232, 7},
		488: {232, 6},
		489: {232, 8},
		490: {232, 6},
		491: {232, 6},
		492: {232, 6},
		493: {232, 6},
		494: {232, 4},
		495: {232, 7},
		496: {232, 6},
		497: {232, 8},
		498: {232, 6},
		499: {232, 4},
		500: {232, 2},
		501: {232, 2},
		502: {229, 4},
		503: {229, 5},
		504: {229, 3},
		505: {229, 4},
		506: {229, 4},
		507: {229, 4},
		508: {229, 4},
		509: {229, 4},
		510: {229, 4},
		511: {229, 4},
		512: {229, 4},
		513: {229, 4},
		514: {229, 4},
		515: {229, 3},
		516: {229, 4},
		517: {229, 4},
		518: {229, 4},
		519: {229, 3},
		520: {229, 4},
		521: {230, 3},
		522: {230, 4},
		523: {230, 5},
		524: {297, 0},
		525: {297, 2},
		526: {296, 0},
		527: {296, 2},
		528: {296, 4},
		529: {357, 1},
		530: {357, 1},
		531: {357, 1},
		532: {277, 1},
		533: {277, 1},
		534: {277, 1},
		535: {272, 1},
		536: {272, 1},
		537: {272, 1},
		538: {272, 1},
		539: {272, 1},
		540: {272, 1},
		541: {272, 1},
		542: {272, 1},
		543: {273, 1},
		544: {273, 1},
		545: {273, 1},
		546: {273, 1},
		547: {273, 1},
		548: {273, 1},
		549: {273, 1},
		550: {273, 1},
		551: {273, 1},
		552: {280, 1},
		553: {280, 1},
		554: {280, 1},
		555: {280, 1},
		556: {280, 1},
		557: {280, 1},
		558: {280, 1},
		559: {280, 1},
		560: {280, 1},
		561: {280, 1},
		562: {280, 1},
		563: {340, 1},
		564: {340, 4},
		565: {340, 1},
		566: {340, 4},
		567: {340, 1},
		568: {340, 1},
		569: {340, 1},
		570: {340, 4},
		571: {340, 6},
		572: {340, 1},
		573: {340, 4},
		574: {340, 1},
		575: {340, 1},
		576: {340, 1},
		577: {340, 1},
		578: {340, 2},
		579: {340, 1},
		580: {340, 1},
		581: {340, 2},
		582: {340, 1},
		583: {340, 1},
		584: {340, 1},
		585: {340, 1},
		586: {340, 1},
		587: {237, 1},
		588: {237, 1},
		589: {237, 1},
		590: {235, 5},
		591: {384, 0},
		592: {384, 1},
		593: {447, 1},
		594: {447, 2},
		595: {349, 4},
		596: {376, 0},
		597: {376, 2},
		598: {219, 1},
		599: {219, 3},
		600: {219, 5},
		601: {222, 1},
		602: {222, 1},
		603: {222, 1},
		604: {222, 2},
		605: {222, 2},
		606: {222, 2},
		607: {222, 4},
		608: {222, 1},
		609: {222, 1},
		610: {220, 1},
		611: {220, 1},
		612: {220, 1},
		613: {323, 0},
		614: {323, 3},
		615: {324, 0},
		616: {324, 2},
		617: {298, 0},
		618: {298, 3},
		619: {414, 1},
		620: {414, 3},
		621: {331, 2},
		622: {354, 0},
		623: {354, 1},
		624: {354, 1},
		625: {278, 0},
		626: {278, 2},
		627: {278, 4},
		628: {278, 4},
		629: {330, 0},
		630: {330, 2},
		631: {330, 4},
		632: {190, 1},
		633: {190, 1},
		634: {308, 1},
		635: {308, 1},
		636: {189, 1},
		637: {189, 1},
		638: {189, 1},
		639: {189, 1},
		640: {189, 1},
		641: {189, 1},
		642: {189, 1},
		643: {189, 1},
		644: {189, 1},
		645: {189, 1},
		646: {189, 1},
		647: {189, 1},
		648: {189, 1},
		649: {189, 1},
		650: {189, 1},
		651: {189, 1},
		652: {189, 1},
		653: {189, 1},
		654: {189, 1},
		655: {189, 1},
		656: {189, 1},
		657: {189, 1},
		658: {189, 1},
		659: {189, 1},
		660: {189, 1},
		661: {189, 1},
		662: {189, 1},
		663: {189, 1},
		664: {189, 1},
		665: {189, 1},
		666: {189, 1},
		667: {189, 1},
		668: {189, 1},
		669: {189, 1},
		670: {189, 1},
		671: {189, 1},
		672: {189, 1},
		673: {189, 1},
		674: {189, 1},
		675: {189, 1},
		676: {189, 1},
		677: {189, 1},
		678: {189, 1},
		679: {189, 1},
		680: {189, 1},
		681: {189, 1},
		682: {189, 1},
		683: {189, 1},
		684: {189, 1},
		685: {189, 1},
		686: {189, 1},
		687: {189, 1},
		688: {189, 1},
		689: {189, 1},
		690: {189, 1},
		691: {189, 1},
		692: {189, 1},
		693: {189, 1},
		694: {189, 1},
		695: {189, 1},
		696: {189, 1},
		697: {189, 1},
		698: {189, 1},
		699: {189, 1},
		700: {189, 1},
		701: {189, 1},
		702: {189, 1},
		703: {189, 1},
		704: {189, 1},
		705: {189, 1},
		706: {189, 1},
		707: {189, 1},
		708: {189, 1},
		709: {189, 1},
		710: {189, 1},
		711: {189, 1},
		712: {189, 1},
		713: {189, 1},
		714: {189, 1},
		715: {189, 1},
		716: {189, 1},
		717: {189, 1},
		718: {189, 1},
		719: {189, 1},
		720: {189, 1},
		721: {189, 1},
		722: {189, 1},
		723: {189, 1},
		724: {189, 1},
		725: {189, 1},
		726: {189, 1},
		727: {189, 1},
		728: {189, 1},
		729: {189, 1},
		730: {189, 1},
		731: {189, 1},
		732: {189, 1},
		733: {189, 1},
		734: {189, 1},
		735: {189, 1},
		736: {189, 1},
		737: {189, 1},
		738: {189, 1},
		739: {189, 1},
		740: {189, 1},
		741: {189, 1},
	}

	yyXErrors = map[yyXError]string{}

	yyParseTab = [1336][]uint16{
		// 0
		{44: 775, 774, 114: 765, 128: 772, 165: 781, 199: 776, 221: 766, 241: 767, 769, 771, 263: 762, 764, 268: 761, 745, 763, 314: 768, 350: 777, 352: 757, 743, 364: 744, 366: 753, 368: 752, 372: 780, 374: 754, 755, 378: 779, 782, 381: 750, 387: 770, 756, 393: 759, 399: 760, 403: 783, 405: 748, 417: 758, 420: 746, 424: 747, 778, 427: 749, 442: 773, 444: 751},
		{1: 742},
		{1: 741},
		{1: 740, 118: 1178, 1179, 1177, 1176, 274: 1174},
		{1: 739},
		// 5
		{1: 738},
		{1: 737},
		{1: 736},
		{1: 735},
		{1: 734},
		// 10
		{1: 733},
		{1: 732},
		{1: 731},
		{1: 730},
		{1: 729},
		// 15
		{1: 728},
		{1: 727},
		{1: 726},
		{1: 725},
		{1: 723, 118: 723, 723, 723, 723, 124: 125, 129: 1232, 298: 2076},
		// 20
		{424, 2: 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 25: 424, 424, 424, 424, 424, 424, 32: 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 122: 424, 146: 424, 424, 184: 424, 424, 188: 424, 192: 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 271: 424, 287: 424, 300: 1899, 303: 2065},
		{114: 765, 263: 1961, 764, 268: 1162, 1962, 763},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 1940, 304: 1941, 316: 1942, 416: 1943},
		{114: 765, 263: 762, 764, 268: 1162, 1939, 763},
		{3: 1938},
		// 25
		{424, 2: 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 25: 424, 424, 424, 424, 424, 424, 32: 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 506, 424, 424, 424, 424, 424, 424, 1894, 424, 424, 424, 223: 1895, 287: 424, 300: 1899, 303: 1892, 322: 1898, 334: 1896, 1897, 383: 1893},
		{53: 1756, 195: 1768, 281: 538, 345: 1769},
		{53: 1756, 195: 1754, 281: 538, 345: 1755},
		{40: 1752, 418: 1753},
		{281: 1744},
		// 30
		{27: 1730},
		{27: 1729},
		{292: 1728},
		{292: 1727},
		{586, 2: 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 25: 586, 586, 586, 586, 586, 586, 32: 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 586, 187: 586, 400: 1696, 1695},
		// 35
		{281: 1650},
		{21: 1517, 23: 1489, 25: 506, 27: 510, 36: 506, 510, 346, 1493, 41: 510, 43: 1492, 46: 1491, 1494, 56: 1509, 1487, 63: 1518, 66: 1511, 69: 1482, 1495, 73: 1497, 78: 1486, 84: 1499, 1500, 1501, 1503, 1504, 1515, 1505, 94: 1506, 96: 1519, 100: 1508, 223: 1516, 238: 1485, 276: 1496, 281: 1507, 292: 1498, 314: 1488, 322: 1484, 332: 1502, 334: 1513, 1483, 370: 1510, 419: 1512, 426: 1514, 429: 1490},
		{440, 2: 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 25: 440, 440, 440, 440, 440, 440, 32: 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 440, 114: 440, 123: 440, 187: 440, 263: 440, 440},
		{439, 2: 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 25: 439, 439, 439, 439, 439, 439, 32: 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 439, 114: 439, 123: 439, 187: 439, 263: 439, 439},
		{438, 2: 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 25: 438, 438, 438, 438, 438, 438, 32: 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 438, 114: 438, 123: 438, 187: 438, 263: 438, 438},
		// 40
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 1455, 1457, 909, 1456, 964, 929, 114: 443, 123: 443, 187: 1459, 189: 864, 1458, 263: 443, 443, 266: 1460, 382: 1461},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 786, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 787, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 788, 404: 789},
		{114: 520},
		{114: 519},
		{508, 92, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 25: 508, 508, 508, 508, 508, 508, 32: 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 122: 508, 133: 92, 145: 92, 147: 92, 92, 150: 92, 161: 92, 168: 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 508, 508, 187: 92, 508, 192: 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508, 508},
		// 45
		{507, 39, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 25: 507, 507, 507, 507, 507, 507, 32: 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 122: 507, 133: 39, 145: 39, 147: 39, 39, 150: 39, 161: 39, 168: 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 507, 507, 187: 39, 507, 192: 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507, 507},
		{1: 426, 133: 985, 145: 983, 148: 984},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1454},
		{36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 1408, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 186: 36, 36},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 1164, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1173, 263: 762, 764, 268: 1162, 1163, 763, 291: 1167},
		// 50
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1453},
		{330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 118: 330, 330, 330, 330, 123: 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 148: 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 330, 1416, 330, 330, 330, 330, 330, 330, 1422, 1420, 1415, 1421, 1419, 1417, 1418},
		{313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 118: 313, 313, 313, 313, 123: 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 148: 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313},
		{298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 1205, 115: 1210, 1369, 1206, 298, 298, 298, 298, 123: 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 1207, 298, 298, 1371, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 1368, 1203, 1204, 1211, 1208, 1209, 1370, 1372, 1373},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 1144, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1367},
		// 55
		{285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 115: 285, 285, 285, 285, 285, 285, 285, 123: 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285},
		{284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 115: 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 186: 284},
		{283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 115: 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 186: 283},
		{282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 115: 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 186: 282},
		{281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 115: 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 186: 281},
		// 60
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 1348, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 1366, 802},
		{114: 1364, 218: 1365},
		{278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 115: 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 186: 278},
		{277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 115: 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 186: 277},
		{114: 1361},
		// 65
		{3: 1338, 389: 1337},
		{274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 115: 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 186: 274},
		{273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 115: 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 186: 273},
		{272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 115: 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 186: 272},
		{271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 115: 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 186: 271},
		// 70
		{114: 1334},
		{114: 1329},
		{114: 1326},
		{114: 1323},
		{114: 1320},
		// 75
		{114: 1311},
		{114: 1270},
		{218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 1075, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 186: 218, 297: 1269},
		{218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 1266, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 186: 218, 297: 1265},
		{114: 1258},
		// 80
		{114: 1251},
		{114: 1246},
		{114: 1227},
		{114: 1137},
		{14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 1128, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 186: 14, 14},
		// 85
		{13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 1102, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 186: 13, 13},
		{114: 1088},
		{114: 1078},
		{218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 1075, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 186: 218, 297: 1077},
		{218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 1075, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 186: 218, 297: 1074},
		// 90
		{91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 1068, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 186: 91, 91},
		{114: 1066},
		{90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 1063, 90, 90, 90, 1062, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 186: 90, 90},
		{88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 1059, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 186: 88, 88},
		{69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 1056, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 186: 69, 69},
		// 95
		{114: 1053},
		{58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 1050, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 186: 58, 58},
		{57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 1047, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 186: 57, 57},
		{114: 1044},
		{56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 1041, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 186: 56, 56},
		// 100
		{40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 1038, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 186: 40, 40},
		{114: 1036},
		{35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 1033, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 186: 35, 35},
		{15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 1030, 15, 15, 15, 1029, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 186: 15, 15},
		{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1026, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 186: 2, 2},
		// 105
		{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 1024, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 186: 7, 7},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1021, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 186: 1, 1},
		{110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 996, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 186: 110, 110},
		{155, 2: 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 25: 155, 155, 155, 155, 155, 155, 32: 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 117: 155, 122: 155, 184: 155, 155, 188: 155, 192: 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155, 155},
		{154, 2: 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 25: 154, 154, 154, 154, 154, 154, 32: 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 117: 154, 122: 154, 184: 154, 154, 188: 154, 192: 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154, 154},
		// 110
		{153, 2: 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 25: 153, 153, 153, 153, 153, 153, 32: 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 117: 153, 122: 153, 184: 153, 153, 188: 153, 192: 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153, 153},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 162: 151, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 977, 384: 978},
		{144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 115: 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 186: 144, 951, 238: 144, 144, 241: 144, 244: 144, 144, 247: 144, 144, 250: 144, 144, 253: 144, 257: 144},
		{141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 115: 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 141, 186: 141},
		{140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 115: 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 140, 186: 140, 52},
		// 115
		{139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 115: 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 139, 186: 139},
		{16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 950, 16, 16, 16, 115: 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 186: 16, 16},
		{134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 115: 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 186: 134},
		{133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 115: 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 186: 133},
		{132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 115: 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 132, 186: 132},
		// 120
		{131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 115: 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 131, 186: 131},
		{130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 115: 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 186: 130, 8},
		{109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 191: 109, 221: 109, 223: 109, 109, 109, 109, 234: 109, 238: 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 109, 255: 109, 109, 109, 109},
		{106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 191: 106, 221: 106, 223: 106, 106, 106, 106, 234: 106, 238: 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 255: 106, 106, 106, 106},
		{105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 191: 105, 221: 105, 223: 105, 105, 105, 105, 234: 105, 238: 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 255: 105, 105, 105, 105},
		// 125
		{104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 191: 104, 221: 104, 223: 104, 104, 104, 104, 234: 104, 238: 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 104, 255: 104, 104, 104, 104},
		{103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 191: 103, 221: 103, 223: 103, 103, 103, 103, 234: 103, 238: 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 103, 255: 103, 103, 103, 103},
		{102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 191: 102, 221: 102, 223: 102, 102, 102, 102, 234: 102, 238: 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 255: 102, 102, 102, 102},
		{101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 191: 101, 221: 101, 223: 101, 101, 101, 101, 234: 101, 238: 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 101, 255: 101, 101, 101, 101},
		{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 191: 100, 221: 100, 223: 100, 100, 100, 100, 234: 100, 238: 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 255: 100, 100, 100, 100},
		// 130
		{99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 191: 99, 221: 99, 223: 99, 99, 99, 99, 234: 99, 238: 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 255: 99, 99, 99, 99},
		{98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 191: 98, 221: 98, 223: 98, 98, 98, 98, 234: 98, 238: 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 98, 255: 98, 98, 98, 98},
		{97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 191: 97, 221: 97, 223: 97, 97, 97, 97, 234: 97, 238: 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 97, 255: 97, 97, 97, 97},
		{96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 191: 96, 221: 96, 223: 96, 96, 96, 96, 234: 96, 238: 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 96, 255: 96, 96, 96, 96},
		{95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 191: 95, 221: 95, 223: 95, 95, 95, 95, 234: 95, 238: 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 95, 255: 95, 95, 95, 95},
		// 135
		{94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 191: 94, 221: 94, 223: 94, 94, 94, 94, 234: 94, 238: 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 94, 255: 94, 94, 94, 94},
		{93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 191: 93, 221: 93, 223: 93, 93, 93, 93, 234: 93, 238: 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 93, 255: 93, 93, 93, 93},
		{89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 191: 89, 221: 89, 223: 89, 89, 89, 89, 234: 89, 238: 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 89, 255: 89, 89, 89, 89},
		{87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 191: 87, 221: 87, 223: 87, 87, 87, 87, 234: 87, 238: 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 255: 87, 87, 87, 87},
		{86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 191: 86, 221: 86, 223: 86, 86, 86, 86, 234: 86, 238: 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 255: 86, 86, 86, 86},
		// 140
		{85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 191: 85, 221: 85, 223: 85, 85, 85, 85, 234: 85, 238: 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 255: 85, 85, 85, 85},
		{84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 191: 84, 221: 84, 223: 84, 84, 84, 84, 234: 84, 238: 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 255: 84, 84, 84, 84},
		{83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 191: 83, 221: 83, 223: 83, 83, 83, 83, 234: 83, 238: 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 255: 83, 83, 83, 83},
		{82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 191: 82, 221: 82, 223: 82, 82, 82, 82, 234: 82, 238: 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 255: 82, 82, 82, 82},
		{81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 191: 81, 221: 81, 223: 81, 81, 81, 81, 234: 81, 238: 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 255: 81, 81, 81, 81},
		// 145
		{80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 191: 80, 221: 80, 223: 80, 80, 80, 80, 234: 80, 238: 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 255: 80, 80, 80, 80},
		{79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 191: 79, 221: 79, 223: 79, 79, 79, 79, 234: 79, 238: 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 255: 79, 79, 79, 79},
		{78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 191: 78, 221: 78, 223: 78, 78, 78, 78, 234: 78, 238: 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 255: 78, 78, 78, 78},
		{77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 191: 77, 221: 77, 223: 77, 77, 77, 77, 234: 77, 238: 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 255: 77, 77, 77, 77},
		{76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 191: 76, 221: 76, 223: 76, 76, 76, 76, 234: 76, 238: 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 255: 76, 76, 76, 76},
		// 150
		{75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 191: 75, 221: 75, 223: 75, 75, 75, 75, 234: 75, 238: 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 255: 75, 75, 75, 75},
		{74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 191: 74, 221: 74, 223: 74, 74, 74, 74, 234: 74, 238: 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 255: 74, 74, 74, 74},
		{73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 191: 73, 221: 73, 223: 73, 73, 73, 73, 234: 73, 238: 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 255: 73, 73, 73, 73},
		{72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 191: 72, 221: 72, 223: 72, 72, 72, 72, 234: 72, 238: 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 255: 72, 72, 72, 72},
		{71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 191: 71, 221: 71, 223: 71, 71, 71, 71, 234: 71, 238: 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 255: 71, 71, 71, 71},
		// 155
		{70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 191: 70, 221: 70, 223: 70, 70, 70, 70, 234: 70, 238: 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 255: 70, 70, 70, 70},
		{68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 191: 68, 221: 68, 223: 68, 68, 68, 68, 234: 68, 238: 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 255: 68, 68, 68, 68},
		{67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 191: 67, 221: 67, 223: 67, 67, 67, 67, 234: 67, 238: 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 255: 67, 67, 67, 67},
		{66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 191: 66, 221: 66, 223: 66, 66, 66, 66, 234: 66, 238: 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 255: 66, 66, 66, 66},
		{65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 191: 65, 221: 65, 223: 65, 65, 65, 65, 234: 65, 238: 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 255: 65, 65, 65, 65},
		// 160
		{64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 191: 64, 221: 64, 223: 64, 64, 64, 64, 234: 64, 238: 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 255: 64, 64, 64, 64},
		{63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 191: 63, 221: 63, 223: 63, 63, 63, 63, 234: 63, 238: 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 255: 63, 63, 63, 63},
		{62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 191: 62, 221: 62, 223: 62, 62, 62, 62, 234: 62, 238: 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 255: 62, 62, 62, 62},
		{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 191: 61, 221: 61, 223: 61, 61, 61, 61, 234: 61, 238: 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 255: 61, 61, 61, 61},
		{60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 191: 60, 221: 60, 223: 60, 60, 60, 60, 234: 60, 238: 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 255: 60, 60, 60, 60},
		// 165
		{59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 191: 59, 221: 59, 223: 59, 59, 59, 59, 234: 59, 238: 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 255: 59, 59, 59, 59},
		{55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 191: 55, 221: 55, 223: 55, 55, 55, 55, 234: 55, 238: 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 255: 55, 55, 55, 55},
		{54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 191: 54, 221: 54, 223: 54, 54, 54, 54, 234: 54, 238: 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 255: 54, 54, 54, 54},
		{53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 191: 53, 221: 53, 223: 53, 53, 53, 53, 234: 53, 238: 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 255: 53, 53, 53, 53},
		{51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 191: 51, 221: 51, 223: 51, 51, 51, 51, 234: 51, 238: 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 255: 51, 51, 51, 51},
		// 170
		{50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 191: 50, 221: 50, 223: 50, 50, 50, 50, 234: 50, 238: 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 255: 50, 50, 50, 50},
		{49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 191: 49, 221: 49, 223: 49, 49, 49, 49, 234: 49, 238: 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 255: 49, 49, 49, 49},
		{48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 191: 48, 221: 48, 223: 48, 48, 48, 48, 234: 48, 238: 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 255: 48, 48, 48, 48},
		{47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 191: 47, 221: 47, 223: 47, 47, 47, 47, 234: 47, 238: 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 255: 47, 47, 47, 47},
		{46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 191: 46, 221: 46, 223: 46, 46, 46, 46, 234: 46, 238: 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 255: 46, 46, 46, 46},
		// 175
		{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 191: 45, 221: 45, 223: 45, 45, 45, 45, 234: 45, 238: 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 255: 45, 45, 45, 45},
		{44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 191: 44, 221: 44, 223: 44, 44, 44, 44, 234: 44, 238: 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 255: 44, 44, 44, 44},
		{43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 191: 43, 221: 43, 223: 43, 43, 43, 43, 234: 43, 238: 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 255: 43, 43, 43, 43},
		{42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 191: 42, 221: 42, 223: 42, 42, 42, 42, 234: 42, 238: 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 255: 42, 42, 42, 42},
		{41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 191: 41, 221: 41, 223: 41, 41, 41, 41, 234: 41, 238: 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 255: 41, 41, 41, 41},
		// 180
		{38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 191: 38, 221: 38, 223: 38, 38, 38, 38, 234: 38, 238: 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 255: 38, 38, 38, 38},
		{37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 191: 37, 221: 37, 223: 37, 37, 37, 37, 234: 37, 238: 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 255: 37, 37, 37, 37},
		{34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 191: 34, 221: 34, 223: 34, 34, 34, 34, 234: 34, 238: 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 255: 34, 34, 34, 34},
		{33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 191: 33, 221: 33, 223: 33, 33, 33, 33, 234: 33, 238: 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 255: 33, 33, 33, 33},
		{32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 191: 32, 221: 32, 223: 32, 32, 32, 32, 234: 32, 238: 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 255: 32, 32, 32, 32},
		// 185
		{31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 191: 31, 221: 31, 223: 31, 31, 31, 31, 234: 31, 238: 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 255: 31, 31, 31, 31},
		{30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 191: 30, 221: 30, 223: 30, 30, 30, 30, 234: 30, 238: 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 255: 30, 30, 30, 30},
		{29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 191: 29, 221: 29, 223: 29, 29, 29, 29, 234: 29, 238: 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 255: 29, 29, 29, 29},
		{28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 191: 28, 221: 28, 223: 28, 28, 28, 28, 234: 28, 238: 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 255: 28, 28, 28, 28},
		{27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 191: 27, 221: 27, 223: 27, 27, 27, 27, 234: 27, 238: 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 255: 27, 27, 27, 27},
		// 190
		{26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 191: 26, 221: 26, 223: 26, 26, 26, 26, 234: 26, 238: 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 255: 26, 26, 26, 26},
		{25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 191: 25, 221: 25, 223: 25, 25, 25, 25, 234: 25, 238: 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 255: 25, 25, 25, 25},
		{24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 191: 24, 221: 24, 223: 24, 24, 24, 24, 234: 24, 238: 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 255: 24, 24, 24, 24},
		{23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 191: 23, 221: 23, 223: 23, 23, 23, 23, 234: 23, 238: 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 255: 23, 23, 23, 23},
		{22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 191: 22, 221: 22, 223: 22, 22, 22, 22, 234: 22, 238: 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 255: 22, 22, 22, 22},
		// 195
		{21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 191: 21, 221: 21, 223: 21, 21, 21, 21, 234: 21, 238: 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 255: 21, 21, 21, 21},
		{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 191: 20, 221: 20, 223: 20, 20, 20, 20, 234: 20, 238: 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 255: 20, 20, 20, 20},
		{19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 191: 19, 221: 19, 223: 19, 19, 19, 19, 234: 19, 238: 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 255: 19, 19, 19, 19},
		{18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 191: 18, 221: 18, 223: 18, 18, 18, 18, 234: 18, 238: 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 255: 18, 18, 18, 18},
		{17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 191: 17, 221: 17, 223: 17, 17, 17, 17, 234: 17, 238: 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 255: 17, 17, 17, 17},
		// 200
		{12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 191: 12, 221: 12, 223: 12, 12, 12, 12, 234: 12, 238: 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 255: 12, 12, 12, 12},
		{11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 191: 11, 221: 11, 223: 11, 11, 11, 11, 234: 11, 238: 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 255: 11, 11, 11, 11},
		{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 191: 10, 221: 10, 223: 10, 10, 10, 10, 234: 10, 238: 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 255: 10, 10, 10, 10},
		{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 191: 9, 221: 9, 223: 9, 9, 9, 9, 234: 9, 238: 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 255: 9, 9, 9, 9},
		{6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 191: 6, 221: 6, 223: 6, 6, 6, 6, 234: 6, 238: 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 255: 6, 6, 6, 6},
		// 205
		{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 191: 5, 221: 5, 223: 5, 5, 5, 5, 234: 5, 238: 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 255: 5, 5, 5, 5},
		{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 191: 4, 221: 4, 223: 4, 4, 4, 4, 234: 4, 238: 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 255: 4, 4, 4, 4},
		{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 191: 3, 221: 3, 223: 3, 3, 3, 3, 234: 3, 238: 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 255: 3, 3, 3, 3},
		{137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 115: 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 137, 186: 137},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 952},
		// 210
		{143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 115: 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 186: 143, 975, 238: 143, 143, 241: 143, 244: 143, 143, 247: 143, 143, 250: 143, 143, 253: 143, 257: 143},
		{110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 191: 110, 221: 110, 223: 110, 110, 110, 110, 234: 110, 238: 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 255: 110, 110, 110, 110},
		{92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 191: 92, 221: 92, 223: 92, 92, 92, 92, 234: 92, 238: 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 92, 255: 92, 92, 92, 92},
		{91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 191: 91, 221: 91, 223: 91, 91, 91, 91, 234: 91, 238: 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 255: 91, 91, 91, 91},
		{90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 191: 90, 221: 90, 223: 90, 90, 90, 90, 234: 90, 238: 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 255: 90, 90, 90, 90},
		// 215
		{88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 191: 88, 221: 88, 223: 88, 88, 88, 88, 234: 88, 238: 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 255: 88, 88, 88, 88},
		{69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 191: 69, 221: 69, 223: 69, 69, 69, 69, 234: 69, 238: 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 255: 69, 69, 69, 69},
		{58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 191: 58, 221: 58, 223: 58, 58, 58, 58, 234: 58, 238: 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 255: 58, 58, 58, 58},
		{57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 191: 57, 221: 57, 223: 57, 57, 57, 57, 234: 57, 238: 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 255: 57, 57, 57, 57},
		{56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 191: 56, 221: 56, 223: 56, 56, 56, 56, 234: 56, 238: 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 255: 56, 56, 56, 56},
		// 220
		{52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 191: 52, 221: 52, 223: 52, 52, 52, 52, 234: 52, 238: 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 255: 52, 52, 52, 52},
		{40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 191: 40, 221: 40, 223: 40, 40, 40, 40, 234: 40, 238: 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 255: 40, 40, 40, 40},
		{39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 191: 39, 221: 39, 223: 39, 39, 39, 39, 234: 39, 238: 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 255: 39, 39, 39, 39},
		{36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 191: 36, 221: 36, 223: 36, 36, 36, 36, 234: 36, 238: 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 255: 36, 36, 36, 36},
		{35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 191: 35, 221: 35, 223: 35, 35, 35, 35, 234: 35, 238: 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 255: 35, 35, 35, 35},
		// 225
		{16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 191: 16, 221: 16, 223: 16, 16, 16, 16, 234: 16, 238: 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 255: 16, 16, 16, 16},
		{15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 191: 15, 221: 15, 223: 15, 15, 15, 15, 234: 15, 238: 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 255: 15, 15, 15, 15},
		{14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 191: 14, 221: 14, 223: 14, 14, 14, 14, 234: 14, 238: 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 255: 14, 14, 14, 14},
		{13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 191: 13, 221: 13, 223: 13, 13, 13, 13, 234: 13, 238: 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 255: 13, 13, 13, 13},
		{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 191: 8, 221: 8, 223: 8, 8, 8, 8, 234: 8, 238: 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 255: 8, 8, 8, 8},
		// 230
		{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 191: 7, 221: 7, 223: 7, 7, 7, 7, 234: 7, 238: 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 255: 7, 7, 7, 7},
		{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 191: 2, 221: 2, 223: 2, 2, 2, 2, 234: 2, 238: 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 255: 2, 2, 2, 2},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 191: 1, 221: 1, 223: 1, 1, 1, 1, 234: 1, 238: 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 255: 1, 1, 1, 1},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 976},
		{142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 115: 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 142, 186: 142, 238: 142, 142, 241: 142, 244: 142, 142, 247: 142, 142, 250: 142, 142, 253: 142, 257: 142},
		// 235
		{133: 985, 145: 983, 148: 984, 162: 150},
		{162: 981, 349: 980, 447: 979},
		{162: 981, 146, 993, 349: 992, 376: 991},
		{162: 149, 149, 149},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 982},
		// 240
		{133: 985, 145: 983, 148: 984, 167: 986},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 990},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 989},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 988},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 987},
		// 245
		{133: 985, 145: 983, 148: 984, 162: 147, 147, 147},
		{334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 118: 334, 334, 334, 334, 123: 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 148: 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 334, 162: 334, 334, 334, 334, 334, 334},
		{335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 118: 335, 335, 335, 335, 123: 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 985, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 148: 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 335, 162: 335, 335, 335, 335, 335, 335},
		{336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 118: 336, 336, 336, 336, 123: 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 985, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 148: 984, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 336, 162: 336, 336, 336, 336, 336, 336},
		{163: 995},
		// 250
		{162: 148, 148, 148},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 994},
		{133: 985, 145: 983, 148: 984, 163: 145},
		{152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 115: 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 152, 186: 152},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 1002, 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 998, 271: 1003},
		// 255
		{1: 406, 24: 406, 31: 406, 118: 406, 406, 406, 406, 124: 406, 129: 406, 149: 406, 159: 406},
		{24: 1020, 31: 1005},
		{1: 404, 4: 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 24: 404, 31: 404, 118: 404, 404, 404, 404, 123: 404, 404, 129: 404, 134: 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 149: 404, 159: 404},
		{877, 386, 867, 953, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 873, 962, 884, 386, 938, 956, 940, 967, 968, 971, 386, 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 1016, 118: 386, 386, 386, 386, 123: 386, 386, 127: 386, 129: 386, 133: 985, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 983, 148: 984, 386, 159: 386, 1015, 189: 864, 1014, 290: 1013},
		{144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 113: 144, 115: 144, 144, 144, 144, 144, 144, 144, 123: 144, 144, 127: 144, 129: 144, 133: 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 147: 144, 144, 144, 144, 159: 144, 144, 144, 168: 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 144, 187: 1008},
		// 260
		{221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 115: 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 186: 221},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1004},
		{24: 1006, 31: 1005},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1007},
		{219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 115: 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 186: 219},
		// 265
		{1: 405, 24: 405, 31: 405, 118: 405, 405, 405, 405, 124: 405, 129: 405, 149: 405, 159: 405},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 147: 1009, 189: 864, 1010},
		{1: 401, 4: 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 24: 401, 31: 401, 118: 401, 401, 401, 401, 123: 401, 401, 129: 401, 134: 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 149: 401, 159: 401},
		{143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 113: 143, 115: 143, 143, 143, 143, 143, 143, 143, 123: 143, 143, 127: 143, 129: 143, 133: 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 147: 143, 143, 143, 143, 159: 143, 143, 143, 168: 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 143, 187: 1011},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 147: 1012, 189: 864, 976},
		// 270
		{1: 400, 4: 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 24: 400, 31: 400, 118: 400, 400, 400, 400, 123: 400, 400, 129: 400, 134: 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 149: 400, 159: 400},
		{1: 403, 4: 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 24: 403, 31: 403, 118: 403, 403, 403, 403, 123: 403, 403, 127: 1019, 129: 403, 134: 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 149: 403, 159: 403},
		{1: 385, 4: 385, 385, 385, 385, 385, 385, 385, 385, 385, 385, 385, 385, 385, 385, 385, 385, 385, 24: 385, 31: 385, 111: 385, 385, 118: 385, 385, 385, 385, 123: 385, 385, 385, 127: 385, 385, 385, 132: 385, 134: 385, 385, 385, 385, 385, 385, 385, 385, 385, 385, 385, 146: 385, 149: 385, 151: 385, 385, 385, 385, 385, 385, 385, 385, 385, 221: 385, 224: 385, 385, 385, 234: 385, 240: 385},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 1018, 189: 864, 1017},
		{1: 383, 4: 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 24: 383, 31: 383, 111: 383, 383, 118: 383, 383, 383, 383, 123: 383, 383, 383, 127: 383, 383, 383, 132: 383, 134: 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 146: 383, 149: 383, 151: 383, 383, 383, 383, 383, 383, 383, 383, 383, 221: 383, 224: 383, 383, 383, 234: 383, 240: 383},
		// 275
		{1: 384, 4: 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 24: 384, 31: 384, 111: 384, 384, 118: 384, 384, 384, 384, 123: 384, 384, 384, 127: 384, 384, 384, 132: 384, 134: 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 146: 384, 149: 384, 151: 384, 384, 384, 384, 384, 384, 384, 384, 384, 221: 384, 224: 384, 384, 384, 234: 384, 240: 384},
		{1: 382, 4: 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 24: 382, 31: 382, 111: 382, 382, 118: 382, 382, 382, 382, 123: 382, 382, 382, 127: 382, 382, 382, 132: 382, 134: 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 146: 382, 149: 382, 151: 382, 382, 382, 382, 382, 382, 382, 382, 382, 221: 382, 224: 382, 382, 382, 234: 382, 240: 382},
		{1: 402, 4: 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 24: 402, 31: 402, 118: 402, 402, 402, 402, 123: 402, 402, 129: 402, 134: 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 149: 402, 159: 402},
		{220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 115: 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 186: 220},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1022},
		// 280
		{24: 1023, 31: 1005},
		{222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 115: 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 186: 222},
		{24: 1025},
		{223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 115: 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 186: 223},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1027},
		// 285
		{24: 1028, 31: 1005},
		{224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 115: 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 186: 224},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1031},
		{136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 115: 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 186: 136},
		{24: 1032, 31: 1005},
		// 290
		{225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 115: 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 186: 225},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1034},
		{24: 1035, 31: 1005},
		{226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 115: 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 186: 226},
		{24: 1037},
		// 295
		{227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 115: 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 186: 227},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1039},
		{24: 1040, 31: 1005},
		{228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 115: 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 186: 228},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1042},
		// 300
		{24: 1043, 31: 1005},
		{229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 115: 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 186: 229},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1045},
		{24: 1046, 31: 1005},
		{230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 115: 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 186: 230},
		// 305
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1048},
		{24: 1049, 31: 1005},
		{231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 115: 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 186: 231},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1051},
		{24: 1052, 31: 1005},
		// 310
		{232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 115: 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 186: 232},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1054},
		{24: 1055, 31: 1005},
		{234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 115: 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 186: 234},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1057},
		// 315
		{24: 1058, 31: 1005},
		{235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 115: 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 186: 235},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1060},
		{24: 1061, 31: 1005},
		{236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 115: 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 186: 236},
		// 320
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1064},
		{138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 115: 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 138, 186: 138},
		{24: 1065, 31: 1005},
		{237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 115: 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 186: 237},
		{24: 1067},
		// 325
		{238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 115: 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 186: 238},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1069, 271: 1070},
		{24: 1073, 31: 1005},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1071},
		{24: 1072, 31: 1005},
		// 330
		{239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 115: 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 186: 239},
		{240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 115: 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 186: 240},
		{241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 115: 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 186: 241},
		{24: 1076},
		{217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 115: 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 186: 217},
		// 335
		{242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 115: 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 186: 242},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1079, 265: 1080},
		{24: 406, 31: 406, 149: 1082},
		{24: 1081, 31: 1005},
		{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 115: 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 186: 243},
		// 340
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1083},
		{24: 1085, 123: 1084},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1086},
		{244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 115: 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 186: 244},
		{24: 1087},
		// 345
		{245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 115: 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 186: 245},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1089, 356: 1091, 1090, 406: 1092, 436: 1093},
		{24: 1098, 149: 1099},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1094},
		{213, 2: 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 25: 213, 213, 213, 213, 213, 213, 32: 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 122: 213, 147: 213, 184: 213, 213, 188: 213, 192: 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213},
		// 350
		{212, 2: 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 25: 212, 212, 212, 212, 212, 212, 32: 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 122: 212, 147: 212, 184: 212, 212, 188: 212, 192: 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212},
		{211, 2: 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 25: 211, 211, 211, 211, 211, 211, 32: 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 122: 211, 147: 211, 184: 211, 211, 188: 211, 192: 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211},
		{149: 1095},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1096},
		{24: 1097},
		// 355
		{247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 115: 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 186: 247},
		{248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 115: 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 186: 248},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1100},
		{24: 1101},
		{246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 115: 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 186: 246},
		// 360
		{4: 1113, 1117, 1118, 1121, 1119, 1115, 1114, 1120, 1116, 1109, 1110, 1111, 1107, 1106, 1112, 1108, 1105, 272: 1104, 1103},
		{31: 1125},
		{31: 1122},
		{207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 115: 207, 207, 207, 207, 207, 207, 207, 123: 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207},
		{206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 115: 206, 206, 206, 206, 206, 206, 206, 123: 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206},
		// 365
		{205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 115: 205, 205, 205, 205, 205, 205, 205, 123: 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205},
		{204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 115: 204, 204, 204, 204, 204, 204, 204, 123: 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204},
		{203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 115: 203, 203, 203, 203, 203, 203, 203, 123: 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203},
		{202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 115: 202, 202, 202, 202, 202, 202, 202, 123: 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202},
		{201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 115: 201, 201, 201, 201, 201, 201, 201, 123: 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201},
		// 370
		{200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 115: 200, 200, 200, 200, 200, 200, 200, 123: 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200},
		{199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 115: 199, 199, 199, 199, 199, 199, 199, 123: 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199},
		{198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 115: 198, 198, 198, 198, 198, 198, 198, 123: 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198},
		{197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 115: 197, 197, 197, 197, 197, 197, 197, 123: 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197},
		{196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 115: 196, 196, 196, 196, 196, 196, 196, 123: 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196},
		// 375
		{195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 115: 195, 195, 195, 195, 195, 195, 195, 123: 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195},
		{194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 115: 194, 194, 194, 194, 194, 194, 194, 123: 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194},
		{193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 115: 193, 193, 193, 193, 193, 193, 193, 123: 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193},
		{192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 115: 192, 192, 192, 192, 192, 192, 192, 123: 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192},
		{191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 115: 191, 191, 191, 191, 191, 191, 191, 123: 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191},
		// 380
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1123},
		{24: 1124, 31: 1005},
		{249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 115: 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 186: 249},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1126},
		{24: 1127, 31: 1005},
		// 385
		{250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 115: 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 186: 250},
		{4: 1113, 1117, 1118, 1121, 1119, 1115, 1114, 1120, 1116, 1109, 1110, 1111, 1107, 1106, 1112, 1108, 1105, 272: 1130, 1129},
		{31: 1134},
		{31: 1131},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1132},
		// 390
		{24: 1133, 31: 1005},
		{251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 115: 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 186: 251},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1135},
		{24: 1136, 31: 1005},
		{252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 115: 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 186: 252},
		// 395
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1138},
		{31: 1139},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 1140, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1141},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 1144, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1143, 1145},
		{24: 1142},
		// 400
		{254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 115: 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 186: 254},
		{877, 2: 867, 953, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 1016, 127: 386, 133: 985, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 386, 983, 148: 984, 160: 1015, 189: 864, 1014, 272: 1149, 1147, 277: 1183, 280: 1148, 290: 1013},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 1164, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1166, 997, 762, 764, 1165, 268: 1162, 1163, 763, 291: 1167},
		{4: 1113, 1117, 1118, 1121, 1119, 1115, 1114, 1120, 1116, 1109, 1110, 1111, 1107, 1106, 1112, 1108, 1105, 134: 1159, 1156, 1158, 1157, 1153, 1155, 1154, 1151, 1152, 1150, 1160, 272: 1149, 1147, 277: 1146, 280: 1148},
		{24: 1161},
		// 405
		{210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 115: 210, 210, 210, 210, 210, 210, 210, 123: 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210},
		{209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 115: 209, 209, 209, 209, 209, 209, 209, 123: 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209},
		{208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 115: 208, 208, 208, 208, 208, 208, 208, 123: 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208},
		{190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 115: 190, 190, 190, 190, 190, 190, 190, 123: 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190},
		{189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 115: 189, 189, 189, 189, 189, 189, 189, 123: 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189},
		// 410
		{188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 115: 188, 188, 188, 188, 188, 188, 188, 123: 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188},
		{187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 115: 187, 187, 187, 187, 187, 187, 187, 123: 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187},
		{186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 115: 186, 186, 186, 186, 186, 186, 186, 123: 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186},
		{185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 115: 185, 185, 185, 185, 185, 185, 185, 123: 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185},
		{184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 115: 184, 184, 184, 184, 184, 184, 184, 123: 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184},
		// 415
		{183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 115: 183, 183, 183, 183, 183, 183, 183, 123: 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183},
		{182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 115: 182, 182, 182, 182, 182, 182, 182, 123: 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182},
		{181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 115: 181, 181, 181, 181, 181, 181, 181, 123: 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181},
		{180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 115: 180, 180, 180, 180, 180, 180, 180, 123: 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180},
		{253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 115: 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 186: 253},
		// 420
		{1: 723, 24: 723, 118: 723, 723, 723, 723},
		{24: 1182, 118: 1178, 1179, 1177, 1176, 274: 1174},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 1164, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1173, 263: 762, 764, 268: 1162, 1172, 763, 291: 1167},
		{24: 1171, 31: 1005},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 386, 938, 956, 940, 967, 968, 971, 386, 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 1016, 127: 386, 133: 985, 145: 983, 148: 984, 160: 1015, 189: 864, 1014, 290: 1013},
		// 425
		{24: 1168, 31: 1169},
		{340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 115: 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 340, 186: 340},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1170},
		{1: 337, 24: 337, 31: 337, 118: 337, 337, 337, 337, 123: 337, 337, 128: 337, 337, 133: 985, 145: 983, 148: 984, 151: 337},
		{233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 115: 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 186: 233},
		// 430
		{24: 1175, 118: 1178, 1179, 1177, 1176, 274: 1174},
		{1: 338, 24: 338, 31: 338, 118: 338, 338, 338, 338, 123: 338, 338, 128: 338, 338, 133: 985, 145: 983, 148: 984, 151: 338},
		{114: 765, 263: 762, 764, 268: 1162, 1181, 763},
		{339, 2: 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 710, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 115: 339, 339, 339, 710, 710, 710, 710, 127: 339, 132: 339, 339, 145: 339, 339, 339, 339, 150: 339, 153: 339, 339, 339, 160: 339, 339, 168: 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 221: 339, 224: 339, 339},
		{114: 415, 263: 415, 415, 282: 1180},
		// 435
		{114: 413, 263: 413, 413},
		{114: 412, 263: 412, 412},
		{114: 411, 263: 411, 411},
		{114: 414, 263: 414, 414},
		{1: 717, 24: 717, 118: 717, 717, 717, 717, 274: 1174},
		// 440
		{339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 115: 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 339, 186: 339, 221: 339, 224: 339, 339},
		{113: 1201},
		{4: 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 113: 207, 127: 21, 134: 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21},
		{4: 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 113: 206, 127: 24, 134: 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24},
		{4: 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 113: 205, 127: 25, 134: 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25},
		// 445
		{4: 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 113: 204, 127: 22, 134: 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22},
		{4: 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 113: 203, 127: 28, 134: 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28},
		{4: 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 113: 202, 127: 27, 134: 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27},
		{4: 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 113: 201, 127: 26, 134: 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26},
		{4: 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 113: 200, 127: 23, 134: 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23},
		// 450
		{4: 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 113: 199, 127: 1, 134: 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{4: 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 113: 198, 127: 40, 134: 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40},
		{4: 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 113: 197, 127: 56, 134: 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56},
		{4: 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 113: 196, 127: 2, 134: 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		{4: 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 113: 195, 127: 88, 134: 88, 88, 88, 88, 88, 88, 88, 88, 88, 88, 88},
		// 455
		{4: 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 113: 194, 127: 69, 134: 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69},
		{4: 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 113: 193, 127: 57, 134: 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57},
		{4: 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 113: 192, 127: 35, 134: 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35},
		{4: 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 113: 191, 127: 58, 134: 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1202},
		// 460
		{293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 115: 293, 293, 293, 293, 293, 293, 293, 123: 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1226},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1225},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 1222, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1221},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 1218, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1217},
		// 465
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1216},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1215},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1214},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1213},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1212},
		// 470
		{286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 1205, 115: 1210, 286, 1206, 286, 286, 286, 286, 123: 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 1207, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 1208, 1209, 286, 286, 286},
		{287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 115: 287, 287, 287, 287, 287, 287, 287, 123: 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287},
		{288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 115: 288, 288, 288, 288, 288, 288, 288, 123: 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288},
		{289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 115: 289, 289, 289, 289, 289, 289, 289, 123: 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289},
		{290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 115: 290, 290, 290, 290, 290, 290, 290, 123: 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290},
		// 475
		{292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 115: 1210, 292, 292, 292, 292, 292, 292, 123: 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 1207, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 1208, 1209, 292, 292, 292},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 1144, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1219},
		{4: 1113, 1117, 1118, 1121, 1119, 1115, 1114, 1120, 1116, 1109, 1110, 1111, 1107, 1106, 1112, 1108, 1105, 133: 985, 1159, 1156, 1158, 1157, 1153, 1155, 1154, 1151, 1152, 1150, 1160, 983, 148: 984, 272: 1149, 1147, 277: 1220, 280: 1148},
		{291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 115: 291, 291, 291, 291, 291, 291, 291, 123: 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291},
		{295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 115: 1210, 295, 295, 295, 295, 295, 295, 123: 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 1207, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 1208, 1209, 295, 295, 295},
		// 480
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 1144, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1223},
		{4: 1113, 1117, 1118, 1121, 1119, 1115, 1114, 1120, 1116, 1109, 1110, 1111, 1107, 1106, 1112, 1108, 1105, 133: 985, 1159, 1156, 1158, 1157, 1153, 1155, 1154, 1151, 1152, 1150, 1160, 983, 148: 984, 272: 1149, 1147, 277: 1224, 280: 1148},
		{294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 115: 294, 294, 294, 294, 294, 294, 294, 123: 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294},
		{296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 1205, 115: 1210, 296, 1206, 296, 296, 296, 296, 123: 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 1207, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 1208, 1209, 296, 296, 296},
		{297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 1205, 115: 1210, 297, 1206, 297, 297, 297, 297, 123: 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 1207, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 1208, 1209, 297, 297, 297},
		// 485
		{410, 2: 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 25: 410, 410, 410, 410, 410, 410, 32: 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 122: 410, 147: 410, 184: 410, 410, 188: 410, 192: 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 410, 271: 1228, 373: 1229},
		{409, 2: 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 25: 409, 409, 409, 409, 409, 409, 32: 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 122: 409, 147: 409, 184: 409, 409, 188: 409, 192: 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409, 409},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1230},
		{24: 125, 31: 1005, 129: 1232, 159: 125, 298: 1231},
		{24: 408, 159: 1242, 421: 1243},
		// 490
		{311: 1233},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1234, 331: 1236, 414: 1235},
		{1: 120, 24: 120, 31: 120, 118: 120, 120, 120, 120, 123: 120, 120, 128: 120, 133: 985, 145: 983, 148: 984, 159: 120, 165: 1241, 1240, 354: 1239},
		{1: 124, 24: 124, 31: 1237, 118: 124, 124, 124, 124, 123: 124, 124, 128: 124, 159: 124},
		{1: 123, 24: 123, 31: 123, 118: 123, 123, 123, 123, 123: 123, 123, 128: 123, 159: 123},
		// 495
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1234, 331: 1238},
		{1: 122, 24: 122, 31: 122, 118: 122, 122, 122, 122, 123: 122, 122, 128: 122, 159: 122},
		{1: 121, 24: 121, 31: 121, 118: 121, 121, 121, 121, 123: 121, 121, 128: 121, 159: 121},
		{1: 119, 24: 119, 31: 119, 118: 119, 119, 119, 119, 123: 119, 119, 128: 119, 159: 119},
		{1: 118, 24: 118, 31: 118, 118: 118, 118, 118, 118, 123: 118, 118, 128: 118, 159: 118},
		// 500
		{110: 1245},
		{24: 1244},
		{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 115: 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 186: 255},
		{24: 407},
		{4: 1113, 1117, 1118, 1121, 1119, 1115, 1114, 1120, 1116, 1109, 1110, 1111, 1107, 1106, 1112, 1108, 1105, 134: 1159, 1156, 1158, 1157, 1153, 1155, 1154, 1151, 1152, 1150, 1160, 272: 1149, 1147, 277: 1247, 280: 1148},
		// 505
		{149: 1248},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1249},
		{24: 1250},
		{256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 115: 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 186: 256},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1252},
		// 510
		{31: 1253},
		{197: 1254},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1255},
		{4: 1113, 1117, 1118, 1121, 1119, 1115, 1114, 1120, 1116, 1109, 1110, 1111, 1107, 1106, 1112, 1108, 1105, 134: 1159, 1156, 1158, 1157, 1153, 1155, 1154, 1151, 1152, 1150, 1160, 272: 1149, 1147, 277: 1256, 280: 1148},
		{24: 1257},
		// 515
		{257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 115: 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 186: 257},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1259},
		{31: 1260},
		{197: 1261},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1262},
		// 520
		{4: 1113, 1117, 1118, 1121, 1119, 1115, 1114, 1120, 1116, 1109, 1110, 1111, 1107, 1106, 1112, 1108, 1105, 134: 1159, 1156, 1158, 1157, 1153, 1155, 1154, 1151, 1152, 1150, 1160, 272: 1149, 1147, 277: 1263, 280: 1148},
		{24: 1264},
		{258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 115: 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 186: 258},
		{260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 115: 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 186: 260},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 1076, 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1267},
		// 525
		{24: 1268},
		{259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 115: 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 186: 259},
		{261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 115: 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 186: 261},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1271},
		{133: 985, 145: 983, 148: 984, 160: 1272},
		// 530
		{26: 1276, 28: 1284, 32: 1277, 1278, 1280, 49: 1279, 1282, 52: 1283, 184: 1275, 238: 1274, 1281, 336: 1286, 1288, 1290, 1289, 1273, 1287, 347: 1285},
		{24: 1308, 127: 1309},
		{24: 179, 114: 1305, 127: 179},
		{24: 177, 114: 1302, 127: 177},
		{24: 175, 127: 175},
		// 535
		{24: 174, 127: 174},
		{24: 173, 114: 1296, 127: 173},
		{24: 170, 114: 1293, 127: 170},
		{24: 168, 127: 168},
		{24: 167, 127: 167},
		// 540
		{24: 166, 127: 166},
		{24: 165, 127: 165, 239: 1292},
		{24: 163, 127: 163},
		{24: 162, 127: 162, 239: 1291},
		{24: 160, 127: 160},
		// 545
		{24: 159, 127: 159},
		{24: 158, 127: 158},
		{24: 157, 127: 157},
		{24: 156, 127: 156},
		{24: 161, 127: 161},
		// 550
		{24: 164, 127: 164},
		{22: 1294},
		{24: 1295},
		{24: 169, 127: 169},
		{22: 1297},
		// 555
		{24: 1298, 31: 1299},
		{24: 172, 127: 172},
		{22: 1300},
		{24: 1301},
		{24: 171, 127: 171},
		// 560
		{22: 1303},
		{24: 1304},
		{24: 176, 127: 176},
		{22: 1306},
		{24: 1307},
		// 565
		{24: 178, 127: 178},
		{263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 115: 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 186: 263},
		{24: 1310},
		{262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 115: 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 186: 262},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1312},
		// 570
		{31: 1313},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 1314, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 1315},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 1144, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1143, 1317},
		{24: 1316},
		{264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 115: 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 186: 264},
		// 575
		{4: 1113, 1117, 1118, 1121, 1119, 1115, 1114, 1120, 1116, 1109, 1110, 1111, 1107, 1106, 1112, 1108, 1105, 134: 1159, 1156, 1158, 1157, 1153, 1155, 1154, 1151, 1152, 1150, 1160, 272: 1149, 1147, 277: 1318, 280: 1148},
		{24: 1319},
		{265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 115: 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 186: 265},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1321},
		{24: 1322, 31: 1005},
		// 580
		{266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 115: 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 186: 266},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1324},
		{24: 1325, 31: 1005},
		{267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 115: 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 186: 267},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1327},
		// 585
		{24: 1328, 31: 1005},
		{268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 115: 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 186: 268},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1330},
		{31: 1331, 133: 985, 145: 983, 148: 984},
		{26: 1276, 28: 1284, 32: 1277, 1278, 1280, 49: 1279, 1282, 52: 1283, 184: 1275, 238: 1274, 1281, 336: 1286, 1288, 1290, 1289, 1332, 1287, 347: 1285},
		// 590
		{24: 1333},
		{269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 115: 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 186: 269},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1335},
		{24: 1336, 31: 1005},
		{270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 115: 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 186: 270},
		// 595
		{3: 1358, 1357, 1346, 1347, 1349, 1350, 1351, 1352, 1353, 1355, 26: 1345, 29: 1354, 43: 1344, 55: 1356, 97: 1342, 1343, 111: 815, 816, 115: 840, 184: 812, 192: 837, 195: 833, 197: 1348, 199: 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 228: 1341, 811, 810, 808, 809, 829},
		{110: 1339},
		{157: 1340},
		{135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 115: 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 186: 135},
		{157: 1360},
		// 600
		{114: 1128},
		{114: 1102},
		{114: 1068},
		{114: 1062},
		{114: 1059},
		// 605
		{114: 1056},
		{114: 1359},
		{114: 1050},
		{114: 1047},
		{114: 1041},
		// 610
		{114: 1038},
		{114: 1033},
		{114: 1029},
		{114: 1026},
		{114: 1024},
		// 615
		{114: 1021},
		{114: 996},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1165},
		{275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 115: 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 186: 275},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1362},
		// 620
		{24: 1363, 31: 1005},
		{276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 115: 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 186: 276},
		{114: 765, 263: 762, 764, 268: 1162, 1163, 763},
		{279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 115: 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 186: 279},
		{280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 115: 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 186: 280},
		// 625
		{4: 1113, 1117, 1118, 1121, 1119, 1115, 1114, 1120, 1116, 1109, 1110, 1111, 1107, 1106, 1112, 1108, 1105, 133: 985, 1159, 1156, 1158, 1157, 1153, 1155, 1154, 1151, 1152, 1150, 1160, 983, 148: 984, 272: 1149, 1147, 277: 1183, 280: 1148},
		{51: 1405, 114: 791, 218: 1413, 227: 1414},
		{150: 1392, 175: 1390, 181: 1391, 1393, 1394},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1387},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 1348, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 1376, 802, 1377},
		// 630
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1375},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1374},
		{300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 1205, 115: 1210, 117: 1206, 300, 300, 300, 300, 123: 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 1207, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 176: 1203, 1204, 1211, 1208, 1209},
		{302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 1205, 115: 1210, 117: 1206, 302, 302, 302, 302, 123: 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 1207, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 176: 1203, 1204, 1211, 1208, 1209},
		{216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 118: 216, 216, 216, 216, 1381, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 148: 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 186: 1380, 296: 1386},
		// 635
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 1348, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 1378, 802},
		{216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 118: 216, 216, 216, 216, 1381, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 148: 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 186: 1380, 296: 1379},
		{304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 118: 304, 304, 304, 304, 123: 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 148: 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 1348, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 1385, 802},
		{186: 1382},
		// 640
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 1348, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 1383, 802},
		{157: 1384},
		{214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 118: 214, 214, 214, 214, 123: 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 148: 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214},
		{215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 118: 215, 215, 215, 215, 123: 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 148: 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215},
		{306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 118: 306, 306, 306, 306, 123: 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 148: 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306},
		// 645
		{113: 1205, 115: 1210, 117: 1206, 133: 1388, 147: 1207, 176: 1203, 1204, 1211, 1208, 1209},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 1389},
		{308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 118: 308, 308, 308, 308, 123: 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 148: 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308},
		{51: 1405, 114: 791, 218: 1406, 227: 1407},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1402},
		// 650
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 1348, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 1397, 802, 1398},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1396},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 1395},
		{299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 1205, 115: 1210, 117: 1206, 299, 299, 299, 299, 123: 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 1207, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 176: 1203, 1204, 1211, 1208, 1209},
		{301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 1205, 115: 1210, 117: 1206, 301, 301, 301, 301, 123: 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 1207, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 176: 1203, 1204, 1211, 1208, 1209},
		// 655
		{216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 118: 216, 216, 216, 216, 1381, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 148: 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 186: 1380, 296: 1401},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 1348, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 1399, 802},
		{216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 118: 216, 216, 216, 216, 1381, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 148: 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 186: 1380, 296: 1400},
		{303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 118: 303, 303, 303, 303, 123: 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 148: 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303},
		{305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 118: 305, 305, 305, 305, 123: 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 148: 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305},
		// 660
		{113: 1205, 115: 1210, 117: 1206, 133: 1403, 147: 1207, 176: 1203, 1204, 1211, 1208, 1209},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 1404},
		{307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 118: 307, 307, 307, 307, 123: 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 148: 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307},
		{114: 1408},
		{310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 118: 310, 310, 310, 310, 123: 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 148: 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310},
		// 665
		{309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 118: 309, 309, 309, 309, 123: 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 148: 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1173, 291: 1409},
		{31: 1410},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1411},
		{24: 1412, 31: 337, 133: 985, 145: 983, 148: 984},
		// 670
		{341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 115: 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 341, 186: 341},
		{312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 118: 312, 312, 312, 312, 123: 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 148: 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312},
		{311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 118: 311, 311, 311, 311, 123: 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 148: 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311},
		{30: 1450, 116: 1448, 188: 1449, 193: 862, 861, 220: 1447},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 1425, 890, 892, 909, 915, 964, 1424, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 1444, 282: 1423, 286: 1445},
		// 675
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 1425, 890, 892, 909, 915, 964, 1424, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 1441, 282: 1423, 286: 1442},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 1425, 890, 892, 909, 915, 964, 1424, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 1438, 282: 1423, 286: 1439},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 1425, 890, 892, 909, 915, 964, 1424, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 1435, 282: 1423, 286: 1436},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 1425, 890, 892, 909, 915, 964, 1424, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 1432, 282: 1423, 286: 1433},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 1425, 890, 892, 909, 915, 964, 1424, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 1429, 282: 1423, 286: 1430},
		// 680
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 1425, 890, 892, 909, 915, 964, 1424, 855, 815, 816, 850, 791, 840, 117: 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 1426, 282: 1423, 286: 1427},
		{114: 344},
		{29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 343, 29, 29, 29, 29, 29, 29, 29, 123: 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 187: 29},
		{106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 342, 106, 106, 106, 106, 106, 106, 106, 123: 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 106, 187: 106},
		{315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 118: 315, 315, 315, 315, 123: 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 148: 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315},
		// 685
		{114: 1364, 218: 1428},
		{314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 118: 314, 314, 314, 314, 123: 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 148: 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314},
		{317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 118: 317, 317, 317, 317, 123: 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 148: 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317, 317},
		{114: 1364, 218: 1431},
		{316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 118: 316, 316, 316, 316, 123: 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 148: 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316},
		// 690
		{319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 118: 319, 319, 319, 319, 123: 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 148: 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319},
		{114: 1364, 218: 1434},
		{318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 118: 318, 318, 318, 318, 123: 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 148: 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318, 318},
		{321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 118: 321, 321, 321, 321, 123: 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 148: 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321},
		{114: 1364, 218: 1437},
		// 695
		{320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 118: 320, 320, 320, 320, 123: 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 148: 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320},
		{323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 118: 323, 323, 323, 323, 123: 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 148: 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323, 323},
		{114: 1364, 218: 1440},
		{322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 118: 322, 322, 322, 322, 123: 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 148: 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322, 322},
		{325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 118: 325, 325, 325, 325, 123: 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 148: 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325, 325},
		// 700
		{114: 1364, 218: 1443},
		{324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 118: 324, 324, 324, 324, 123: 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 148: 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324, 324},
		{327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 118: 327, 327, 327, 327, 123: 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 148: 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327, 327},
		{114: 1364, 218: 1446},
		{326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 118: 326, 326, 326, 326, 123: 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 148: 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326, 326},
		// 705
		{332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 118: 332, 332, 332, 332, 123: 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 148: 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 332, 162: 332, 332, 332, 332, 332, 332},
		{30: 1450, 188: 1452, 193: 862, 861, 220: 1451},
		{329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 118: 329, 329, 329, 329, 123: 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 148: 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329, 329},
		{130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 118: 130, 130, 130, 130, 123: 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 148: 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 130, 162: 130, 130, 130, 130, 130, 130},
		{331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 118: 331, 331, 331, 331, 123: 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 148: 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 331, 162: 331, 331, 331, 331, 331, 331},
		// 710
		{328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 118: 328, 328, 328, 328, 123: 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 148: 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328, 328},
		{333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 118: 333, 333, 333, 333, 123: 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 148: 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 333, 162: 333, 333, 333, 333, 333, 333},
		{1: 425, 133: 985, 145: 983, 148: 984},
		{77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 25: 77, 77, 77, 77, 77, 77, 32: 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 114: 446, 123: 446, 187: 77, 263: 446, 446},
		{47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 25: 47, 47, 47, 47, 47, 47, 32: 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 114: 445, 123: 445, 187: 47, 263: 445, 445},
		// 715
		{75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 25: 75, 75, 75, 75, 75, 75, 32: 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 161: 1478, 187: 75},
		{436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 436, 114: 436, 118: 436, 436, 436, 436, 123: 436, 436, 436, 128: 436, 436, 132: 436, 146: 436, 151: 436, 436, 436, 436, 436, 436, 436, 436, 160: 436, 185: 436, 187: 1476, 221: 436, 224: 436, 436, 436, 234: 436, 240: 436, 242: 436, 436, 246: 436, 249: 436, 252: 436, 255: 436, 436},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 1475},
		{877, 429, 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 1474, 189: 864, 1473, 380: 1472},
		{114: 1463, 123: 1465, 263: 762, 764, 268: 1162, 1462, 763, 319: 1464},
		// 720
		{1: 442, 118: 1178, 1179, 1177, 1176, 274: 1174},
		{114: 1463, 263: 762, 764, 268: 1162, 1468, 763, 319: 1469},
		{1: 431},
		{65: 1466},
		{22: 1467},
		// 725
		{1: 430},
		{24: 1471, 118: 1178, 1179, 1177, 1176, 274: 1174},
		{24: 1470},
		{1: 441, 24: 441},
		{1: 710, 24: 710, 118: 710, 710, 710, 710, 124: 710, 129: 710},
		// 730
		{1: 432},
		{1: 428},
		{1: 427},
		{435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 435, 114: 435, 118: 435, 435, 435, 435, 123: 435, 435, 435, 128: 435, 435, 132: 435, 146: 435, 151: 435, 435, 435, 435, 435, 435, 435, 435, 160: 435, 185: 435, 221: 435, 224: 435, 435, 435, 234: 435, 240: 435, 242: 435, 435, 246: 435, 249: 435, 252: 435, 255: 435, 435},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 1477},
		// 735
		{434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 434, 114: 434, 118: 434, 434, 434, 434, 123: 434, 434, 434, 128: 434, 434, 132: 434, 146: 434, 151: 434, 434, 434, 434, 434, 434, 434, 434, 160: 434, 185: 434, 221: 434, 224: 434, 434, 434, 234: 434, 240: 434, 242: 434, 434, 246: 434, 249: 434, 252: 434, 255: 434, 434},
		{74: 1480, 392: 1481, 435: 1479},
		{114: 448, 123: 448, 263: 448, 448},
		{114: 447, 123: 447, 263: 447, 447},
		{114: 444, 123: 444, 263: 444, 444},
		// 740
		{27: 509, 37: 509, 41: 509},
		{25: 505, 36: 505},
		{25: 504, 36: 504},
		{40: 1649},
		{25: 1648, 40: 1647},
		// 745
		{39: 1643},
		{47: 1619, 55: 1623, 68: 1618, 102: 1624, 195: 1616, 200: 1617, 281: 1621, 332: 1620, 439: 1622},
		{3: 1613},
		{38: 1612},
		{1: 117, 124: 1559, 278: 1611},
		// 750
		{114: 1606},
		{1: 512, 125: 512, 149: 1537, 512, 175: 1536, 275: 1538, 279: 1548, 293: 1604},
		{25: 1601, 42: 1600},
		{1: 348, 123: 1598, 391: 1597},
		{149: 1537, 175: 1536, 275: 1538, 279: 1595},
		// 755
		{149: 1537, 175: 1536, 275: 1538, 279: 1593},
		{149: 1537, 175: 1536, 275: 1538, 279: 1589},
		{27: 1586},
		{1: 473},
		{1: 472},
		// 760
		{25: 1583, 42: 1582},
		{1: 469},
		{1: 468},
		{39: 1576},
		{25: 1571, 72: 1570},
		// 765
		{25: 1567},
		{1: 512, 125: 512, 149: 1537, 512, 175: 1536, 275: 1538, 279: 1548, 293: 1565},
		{1: 117, 124: 1559, 278: 1558},
		{1: 359, 125: 1521, 150: 1522, 267: 1557},
		{1: 359, 125: 1521, 150: 1522, 267: 1556},
		// 770
		{1: 359, 125: 1521, 150: 1522, 267: 1555},
		{25: 1552, 36: 1551},
		{27: 1533, 37: 1534, 41: 1535},
		{3: 1528},
		{241: 1526},
		// 775
		{1: 359, 125: 1521, 150: 1522, 267: 1525},
		{1: 359, 125: 1521, 150: 1522, 267: 1520},
		{38: 345},
		{1: 449},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1524},
		// 780
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1523},
		{1: 357, 133: 985, 145: 983, 148: 984},
		{1: 358, 133: 985, 145: 983, 148: 984},
		{1: 450},
		{1: 359, 125: 1521, 150: 1522, 267: 1527},
		// 785
		{1: 451},
		{1: 352, 125: 352, 149: 1530, 352, 305: 1529},
		{1: 359, 125: 1521, 150: 1522, 267: 1532},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1531},
		{1: 351, 124: 351, 351, 133: 985, 145: 983, 148: 984, 150: 351},
		// 790
		{1: 455},
		{1: 512, 125: 512, 149: 1537, 512, 175: 1536, 275: 1538, 279: 1548, 293: 1549},
		{149: 1537, 175: 1536, 275: 1538, 279: 1539},
		{1: 453},
		{3: 518, 187: 518},
		// 795
		{3: 517, 187: 517},
		{3: 1542, 187: 1541},
		{1: 359, 125: 1521, 150: 1522, 267: 1540},
		{1: 454},
		{3: 1547},
		// 800
		{1: 513, 125: 513, 149: 1537, 513, 175: 1536, 187: 1543, 275: 1544},
		{3: 1546},
		{3: 1545},
		{1: 514, 125: 514, 150: 514},
		{1: 515, 125: 515, 150: 515},
		// 805
		{1: 516, 125: 516, 150: 516},
		{1: 511, 125: 511, 150: 511},
		{1: 359, 125: 1521, 150: 1522, 267: 1550},
		{1: 456},
		{1: 359, 125: 1521, 150: 1522, 267: 1554},
		// 810
		{1: 359, 125: 1521, 150: 1522, 267: 1553},
		{1: 452},
		{1: 457},
		{1: 458},
		{1: 459},
		// 815
		{1: 460},
		{1: 462},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1560},
		{1: 116, 24: 116, 31: 1561, 82: 1562, 118: 116, 116, 116, 116, 123: 116, 128: 116, 133: 985, 145: 983, 148: 984},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1564},
		// 820
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1563},
		{1: 114, 24: 114, 118: 114, 114, 114, 114, 123: 114, 128: 114, 133: 985, 145: 983, 148: 984},
		{1: 115, 24: 115, 118: 115, 115, 115, 115, 123: 115, 128: 115, 133: 985, 145: 983, 148: 984},
		{1: 359, 125: 1521, 150: 1522, 267: 1566},
		{1: 463},
		// 825
		{1: 512, 125: 512, 149: 1537, 512, 175: 1536, 275: 1538, 279: 1548, 293: 1568},
		{1: 359, 125: 1521, 150: 1522, 267: 1569},
		{1: 464},
		{1: 466},
		{1: 350, 123: 1573, 390: 1572},
		// 830
		{1: 465},
		{62: 1574},
		{3: 1575},
		{1: 349},
		{1: 356, 124: 356, 149: 356, 175: 1578, 328: 1577},
		// 835
		{1: 352, 124: 352, 149: 1530, 305: 1580},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1579},
		{1: 355, 124: 355, 133: 985, 145: 983, 148: 984, 355},
		{1: 117, 124: 1559, 278: 1581},
		{1: 467},
		// 840
		{3: 1585},
		{1: 359, 125: 1521, 150: 1522, 267: 1584},
		{1: 470},
		{1: 471},
		{1: 512, 125: 512, 149: 1537, 512, 175: 1536, 275: 1538, 279: 1548, 293: 1587},
		// 845
		{1: 359, 125: 1521, 150: 1522, 267: 1588},
		{1: 474},
		{1: 361, 125: 1591, 294: 1590},
		{1: 476},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1592},
		// 850
		{1: 360, 24: 360, 118: 360, 360, 360, 360, 123: 360, 360, 128: 360, 360, 133: 985, 145: 983, 148: 984, 151: 360, 360},
		{1: 361, 125: 1591, 294: 1594},
		{1: 477},
		{1: 361, 125: 1591, 294: 1596},
		{1: 478},
		// 855
		{1: 479},
		{3: 1599},
		{1: 347},
		{3: 1603},
		{1: 359, 125: 1521, 150: 1522, 267: 1602},
		// 860
		{1: 480},
		{1: 481},
		{1: 359, 125: 1521, 150: 1522, 267: 1605},
		{1: 482},
		{147: 1607},
		// 865
		{24: 1608},
		{46: 1609, 56: 1610},
		{1: 483},
		{1: 461},
		{1: 484},
		// 870
		{1: 485},
		{25: 1614, 81: 1615},
		{1: 487},
		{1: 486},
		{3: 354, 192: 1637, 327: 1641},
		// 875
		{3: 354, 192: 1637, 327: 1636},
		{3: 1635},
		{3: 1634},
		{3: 1633},
		{3: 1628, 187: 1629},
		// 880
		{3: 1627},
		{3: 1626},
		{3: 1625},
		{1: 488},
		{1: 489},
		// 885
		{1: 490},
		{1: 491, 187: 1631},
		{3: 1630},
		{1: 492},
		{3: 1632},
		// 890
		{1: 493},
		{1: 494},
		{1: 495},
		{1: 496},
		{3: 1640},
		// 895
		{116: 1638},
		{196: 1639},
		{3: 353},
		{1: 497},
		{3: 1642},
		// 900
		{1: 498},
		{1: 356, 124: 356, 149: 356, 175: 1578, 328: 1644},
		{1: 352, 124: 352, 149: 1530, 305: 1645},
		{1: 117, 124: 1559, 278: 1646},
		{1: 499},
		// 905
		{1: 500},
		{1: 475},
		{1: 501},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 187: 1459, 189: 864, 1458, 266: 1651},
		{44: 1654, 1653, 242: 1657, 1659, 246: 1656, 249: 1658, 310: 1655, 351: 1652},
		// 910
		{1: 570, 31: 1693},
		{292: 1692},
		{292: 1691},
		{1: 567, 31: 567},
		{561, 2: 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 25: 561, 561, 561, 561, 561, 561, 32: 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 301: 1665, 1688},
		// 915
		{561, 2: 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 25: 561, 561, 561, 561, 561, 561, 32: 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 301: 1665, 1686},
		{561, 2: 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 25: 561, 561, 561, 561, 561, 561, 32: 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 561, 301: 1665, 1664},
		{541, 2: 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 25: 541, 541, 541, 541, 541, 541, 32: 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 541, 160: 1662, 187: 541, 252: 1661, 434: 1660},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 187: 1459, 189: 864, 1458, 266: 1663},
		{540, 2: 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 25: 540, 540, 540, 540, 540, 540, 32: 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 540, 187: 540},
		// 920
		{539, 2: 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 25: 539, 539, 539, 539, 539, 539, 32: 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 539, 187: 539},
		{1: 562, 31: 562},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1666},
		{560, 2: 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 25: 560, 560, 560, 560, 560, 560, 32: 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560, 560},
		{4: 1681, 26: 1677, 28: 1678, 1679, 32: 1680, 1675, 1674, 54: 1669, 238: 1683, 1671, 244: 1672, 1685, 247: 1673, 1670, 250: 1676, 1684, 253: 1682, 361: 1667, 369: 1668},
		// 925
		{1: 563, 31: 563},
		{1: 559, 31: 559},
		{1: 558, 31: 558},
		{1: 557, 31: 557},
		{1: 556, 31: 556},
		// 930
		{1: 555, 31: 555},
		{1: 554, 31: 554},
		{1: 553, 31: 553},
		{1: 552, 31: 552},
		{1: 551, 31: 551},
		// 935
		{1: 550, 31: 550},
		{1: 549, 31: 549},
		{1: 548, 31: 548},
		{1: 547, 31: 547},
		{1: 546, 31: 546},
		// 940
		{1: 545, 31: 545},
		{1: 544, 31: 544},
		{1: 543, 31: 543},
		{1: 542, 31: 542},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1687},
		// 945
		{1: 564, 31: 564},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1689},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1690},
		{1: 565, 31: 565},
		{1: 568},
		// 950
		{1: 569},
		{242: 1657, 1659, 246: 1656, 249: 1658, 310: 1694},
		{1: 566, 31: 566},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 187: 1459, 189: 864, 1458, 266: 1697},
		{585, 2: 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 25: 585, 585, 585, 585, 585, 585, 32: 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 585, 187: 585},
		// 955
		{35: 584, 114: 1699, 185: 584, 398: 1698},
		{35: 1707, 185: 1708, 446: 1706},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 1700, 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1702, 362: 1701},
		{35: 583, 185: 583},
		{24: 1703, 31: 1704},
		// 960
		{24: 580, 31: 580},
		{35: 582, 185: 582},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1705},
		{24: 581, 31: 581},
		{114: 1710, 445: 1709},
		// 965
		{114: 579},
		{114: 578},
		{1: 587, 31: 1723},
		{22: 1715, 24: 573, 26: 1716, 28: 1717, 1718, 1450, 573, 110: 855, 122: 1719, 188: 859, 191: 1714, 193: 862, 861, 198: 857, 220: 860, 222: 1713, 309: 1712, 348: 1711},
		{24: 1720, 31: 1721},
		// 970
		{24: 574, 31: 574},
		{24: 572, 31: 572},
		{24: 571, 31: 571},
		{24: 140, 31: 140},
		{110: 1063},
		// 975
		{110: 950},
		{110: 1030},
		{3: 1338},
		{1: 576, 31: 576},
		{22: 1715, 26: 1716, 28: 1717, 1718, 1450, 110: 855, 122: 1719, 188: 859, 191: 1714, 193: 862, 861, 198: 857, 220: 860, 222: 1713, 309: 1722},
		// 980
		{24: 575, 31: 575},
		{114: 1724},
		{22: 1715, 24: 573, 26: 1716, 28: 1717, 1718, 1450, 573, 110: 855, 122: 1719, 188: 859, 191: 1714, 193: 862, 861, 198: 857, 220: 860, 222: 1713, 309: 1712, 348: 1725},
		{24: 1726, 31: 1721},
		{1: 577, 31: 577},
		// 985
		{1: 597},
		{1: 598},
		{1: 599},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 187: 1459, 189: 864, 1458, 266: 1733, 343: 1732, 430: 1731},
		{1: 600, 31: 1742},
		// 990
		{1: 596, 31: 596},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 1016, 160: 1015, 189: 864, 1014, 226: 386, 234: 386, 240: 386, 290: 1734},
		{226: 1736, 234: 589, 240: 1738, 408: 1735, 1737},
		{1: 594, 31: 594},
		{1: 591, 31: 591, 76: 1741, 407: 1740},
		// 995
		{234: 1739},
		{234: 588},
		{1: 592, 31: 592},
		{1: 593, 31: 593},
		{1: 590, 31: 590},
		// 1000
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 187: 1459, 189: 864, 1458, 266: 1733, 343: 1743},
		{1: 595, 31: 595},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 187: 1459, 189: 864, 1458, 266: 1747, 344: 1746, 433: 1745},
		{1: 604, 31: 1750},
		{1: 603, 31: 603},
		// 1005
		{252: 1748},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 187: 1459, 189: 864, 1458, 266: 1749},
		{1: 601, 31: 601},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 187: 1459, 189: 864, 1458, 266: 1747, 344: 1751},
		{1: 602, 31: 602},
		// 1010
		{1: 606},
		{1: 605},
		{536, 2: 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 25: 536, 536, 536, 536, 536, 536, 32: 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 192: 1759, 325: 1765},
		{281: 1757},
		{281: 537},
		// 1015
		{536, 2: 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 25: 536, 536, 536, 536, 536, 536, 32: 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 187: 536, 192: 1759, 325: 1758},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 187: 1459, 189: 864, 1458, 266: 1761},
		{196: 1760},
		{535, 2: 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 25: 535, 535, 535, 535, 535, 535, 32: 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 535, 187: 535},
		{1: 532, 255: 1763, 1764, 358: 1762},
		// 1020
		{1: 607},
		{1: 531},
		{1: 530},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 1767, 288: 1766},
		{1: 608},
		// 1025
		{437, 437, 21: 437, 23: 437, 31: 437, 114: 437, 156: 437, 191: 437, 223: 437},
		{534, 2: 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 25: 534, 534, 534, 534, 534, 534, 32: 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 192: 1772, 326: 1890},
		{281: 1770},
		{534, 2: 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 25: 534, 534, 534, 534, 534, 534, 32: 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 534, 187: 534, 192: 1772, 326: 1771},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 187: 1459, 189: 864, 1458, 266: 1775},
		// 1030
		{116: 1773},
		{196: 1774},
		{533, 2: 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 25: 533, 533, 533, 533, 533, 533, 32: 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 533, 187: 533},
		{114: 1776},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 126: 624, 130: 624, 189: 864, 854, 219: 1779, 276: 624, 315: 1778, 320: 1781, 1780, 367: 1777},
		// 1035
		{24: 1866, 31: 1867},
		{24: 681, 31: 681},
		{4: 1811, 26: 1809, 29: 1820, 32: 1819, 1829, 1831, 54: 1814, 58: 1813, 1827, 1807, 67: 1833, 77: 1828, 79: 1812, 1826, 92: 1810, 95: 1815, 184: 1821, 238: 1824, 1817, 241: 1834, 244: 1818, 1808, 247: 1832, 1816, 250: 1830, 1823, 253: 1822, 257: 1825, 360: 1802, 377: 1806, 385: 1805, 410: 1804, 428: 1803},
		{126: 622, 130: 1783, 276: 622, 441: 1782},
		{126: 623, 130: 623, 276: 623},
		// 1040
		{126: 1786, 276: 1785, 396: 1784},
		{126: 621, 276: 621},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 114: 618, 156: 618, 189: 864, 1767, 288: 1788, 395: 1787},
		{620, 2: 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 25: 620, 620, 620, 620, 620, 620, 32: 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 620, 114: 620, 156: 620},
		{619, 2: 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 25: 619, 619, 619, 619, 619, 619, 32: 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 619, 114: 619, 156: 619},
		// 1045
		{114: 616, 156: 1790, 397: 1789},
		{114: 617, 156: 617},
		{114: 1793},
		{61: 1791, 71: 1792},
		{114: 615},
		// 1050
		{114: 614},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1796, 329: 1795, 402: 1794},
		{24: 1799, 31: 1800},
		{24: 613, 31: 613},
		{24: 611, 31: 611, 165: 1798, 1797},
		// 1055
		{24: 610, 31: 610},
		{24: 609, 31: 609},
		{24: 678, 31: 678},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1796, 329: 1801},
		{24: 612, 31: 612},
		// 1060
		{640, 2: 640, 24: 640, 31: 640, 116: 1848, 126: 640, 130: 640, 640, 188: 1847, 191: 640, 411: 1846},
		{677, 2: 677, 24: 677, 31: 677, 116: 677, 126: 677, 130: 677, 677, 188: 677, 191: 677},
		{644, 2: 644, 24: 644, 31: 644, 114: 1843, 116: 644, 126: 644, 130: 644, 644, 188: 644, 191: 644, 448: 1842},
		{642, 2: 642, 24: 642, 31: 642, 114: 1837, 116: 642, 126: 642, 130: 642, 642, 188: 642, 191: 642, 386: 1836},
		{674, 2: 674, 24: 674, 31: 674, 116: 674, 126: 674, 130: 674, 674, 188: 674, 191: 674},
		// 1065
		{673, 2: 673, 24: 673, 31: 673, 116: 673, 126: 673, 130: 673, 673, 188: 673, 191: 673},
		{672, 2: 672, 24: 672, 31: 672, 116: 672, 126: 672, 130: 672, 672, 188: 672, 191: 672},
		{671, 2: 671, 24: 671, 31: 671, 116: 671, 126: 671, 130: 671, 671, 188: 671, 191: 671},
		{670, 2: 670, 24: 670, 31: 670, 116: 670, 126: 670, 130: 670, 670, 188: 670, 191: 670},
		{669, 2: 669, 24: 669, 31: 669, 116: 669, 126: 669, 130: 669, 669, 188: 669, 191: 669},
		// 1070
		{668, 2: 668, 24: 668, 31: 668, 116: 668, 126: 668, 130: 668, 668, 188: 668, 191: 668},
		{667, 2: 667, 24: 667, 31: 667, 114: 667, 116: 667, 126: 667, 130: 667, 667, 188: 667, 191: 667},
		{666, 2: 666, 24: 666, 31: 666, 114: 666, 116: 666, 126: 666, 130: 666, 666, 188: 666, 191: 666},
		{665, 2: 665, 24: 665, 31: 665, 114: 665, 116: 665, 126: 665, 130: 665, 665, 188: 665, 191: 665},
		{664, 2: 664, 24: 664, 31: 664, 114: 664, 116: 664, 126: 664, 130: 664, 664, 188: 664, 191: 664},
		// 1075
		{663, 2: 663, 24: 663, 31: 663, 114: 663, 116: 663, 126: 663, 130: 663, 663, 188: 663, 191: 663},
		{662, 2: 662, 24: 662, 31: 662, 114: 662, 116: 662, 126: 662, 130: 662, 662, 188: 662, 191: 662},
		{661, 2: 661, 24: 661, 31: 661, 114: 661, 116: 661, 126: 661, 130: 661, 661, 188: 661, 191: 661},
		{660, 2: 660, 24: 660, 31: 660, 114: 660, 116: 660, 126: 660, 130: 660, 660, 188: 660, 191: 660},
		{659, 2: 659, 24: 659, 31: 659, 114: 659, 116: 659, 126: 659, 130: 659, 659, 188: 659, 191: 659},
		// 1080
		{658, 2: 658, 24: 658, 31: 658, 114: 658, 116: 658, 126: 658, 130: 658, 658, 188: 658, 191: 658},
		{657, 2: 657, 24: 657, 31: 657, 114: 657, 116: 657, 126: 657, 130: 657, 657, 188: 657, 191: 657},
		{656, 2: 656, 24: 656, 31: 656, 114: 656, 116: 656, 126: 656, 130: 656, 656, 188: 656, 191: 656},
		{655, 2: 655, 24: 655, 31: 655, 114: 655, 116: 655, 126: 655, 130: 655, 655, 188: 655, 191: 655},
		{654, 2: 654, 24: 654, 31: 654, 114: 654, 116: 654, 126: 654, 130: 654, 654, 188: 654, 191: 654},
		// 1085
		{653, 2: 653, 24: 653, 31: 653, 114: 653, 116: 653, 126: 653, 130: 653, 653, 188: 653, 191: 653},
		{652, 2: 652, 24: 652, 31: 652, 114: 652, 116: 652, 126: 652, 130: 652, 652, 188: 652, 191: 652},
		{651, 2: 651, 24: 651, 31: 651, 114: 651, 116: 651, 126: 651, 130: 651, 651, 188: 651, 191: 651},
		{650, 2: 650, 24: 650, 31: 650, 114: 650, 116: 650, 126: 650, 130: 650, 650, 188: 650, 191: 650},
		{649, 2: 649, 24: 649, 31: 649, 114: 649, 116: 649, 126: 649, 130: 649, 649, 188: 649, 191: 649},
		// 1090
		{648, 2: 648, 24: 648, 31: 648, 114: 648, 116: 648, 126: 648, 1835, 130: 648, 648, 188: 648, 191: 648},
		{646, 2: 646, 24: 646, 31: 646, 116: 646, 126: 646, 130: 646, 646, 188: 646, 191: 646},
		{645, 2: 645, 24: 645, 31: 645, 116: 645, 126: 645, 130: 645, 645, 188: 645, 191: 645},
		{647, 2: 647, 24: 647, 31: 647, 114: 647, 116: 647, 126: 647, 130: 647, 647, 188: 647, 191: 647},
		{675, 2: 675, 24: 675, 31: 675, 116: 675, 126: 675, 130: 675, 675, 188: 675, 191: 675},
		// 1095
		{22: 1838},
		{31: 1839},
		{22: 1840},
		{24: 1841},
		{641, 2: 641, 24: 641, 31: 641, 116: 641, 126: 641, 130: 641, 641, 188: 641, 191: 641},
		// 1100
		{676, 2: 676, 24: 676, 31: 676, 116: 676, 126: 676, 130: 676, 676, 188: 676, 191: 676},
		{22: 1844},
		{24: 1845},
		{643, 2: 643, 24: 643, 31: 643, 116: 643, 126: 643, 130: 643, 643, 188: 643, 191: 643},
		{637, 2: 637, 24: 637, 31: 637, 126: 637, 130: 637, 637, 191: 1851, 371: 1850},
		// 1105
		{639, 2: 639, 24: 639, 31: 639, 126: 639, 130: 639, 639, 191: 639},
		{188: 1849},
		{638, 2: 638, 24: 638, 31: 638, 126: 638, 130: 638, 638, 191: 638},
		{634, 2: 1855, 24: 634, 31: 634, 126: 634, 130: 634, 634, 355: 1854},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 1852, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1853},
		// 1110
		{636, 2: 636, 24: 636, 31: 636, 113: 134, 115: 134, 134, 134, 126: 636, 130: 636, 636, 133: 134, 145: 134, 147: 134, 134, 150: 134, 161: 134, 168: 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134},
		{635, 2: 635, 24: 635, 31: 635, 126: 635, 130: 635, 635, 133: 985, 145: 983, 148: 984},
		{632, 24: 632, 31: 632, 126: 632, 130: 1857, 632, 440: 1856},
		{633, 24: 633, 31: 633, 126: 633, 130: 633, 633},
		{629, 24: 629, 31: 629, 126: 1860, 131: 1861, 415: 1859},
		// 1115
		{631, 24: 631, 31: 631, 126: 1858, 131: 631},
		{630, 24: 630, 31: 630, 126: 630, 131: 630},
		{1864, 24: 626, 31: 626, 359: 1863},
		{628, 24: 628, 31: 628},
		{126: 1862},
		// 1120
		{627, 24: 627, 31: 627},
		{24: 679, 31: 679},
		{110: 1865},
		{24: 625, 31: 625},
		{693, 693, 21: 693, 23: 693, 31: 693, 191: 693, 223: 693, 432: 1869},
		// 1125
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 126: 624, 130: 624, 189: 864, 854, 219: 1779, 276: 624, 315: 1868, 320: 1781, 1780},
		{24: 680, 31: 680},
		{691, 694, 21: 691, 23: 691, 31: 1871, 191: 691, 223: 691, 363: 1870},
		{1873, 21: 1878, 23: 1876, 191: 1874, 223: 1877, 312: 1875, 431: 1872},
		{690, 21: 690, 23: 690, 191: 690, 223: 690},
		// 1130
		{692, 692, 21: 692, 23: 692, 31: 692, 191: 692, 223: 692},
		{110: 683, 161: 1881, 295: 1888},
		{21: 1878, 223: 1877, 312: 1885},
		{683, 2: 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 25: 683, 683, 683, 683, 683, 683, 32: 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 161: 1881, 295: 1883},
		{683, 2: 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 25: 683, 683, 683, 683, 683, 683, 32: 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 161: 1881, 295: 1880},
		// 1135
		{241: 1879},
		{684, 2: 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 25: 684, 684, 684, 684, 684, 684, 32: 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 684, 161: 684},
		{685, 2: 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 25: 685, 685, 685, 685, 685, 685, 32: 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 685, 161: 685},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 1767, 288: 1882},
		{682, 2: 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 25: 682, 682, 682, 682, 682, 682, 32: 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682, 682},
		// 1140
		{686, 686, 21: 686, 23: 686, 31: 686, 191: 686, 223: 686},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 1767, 288: 1884},
		{687, 687, 21: 687, 23: 687, 31: 687, 191: 687, 223: 687},
		{683, 2: 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 25: 683, 683, 683, 683, 683, 683, 32: 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 683, 161: 1881, 295: 1886},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 1767, 288: 1887},
		// 1145
		{688, 688, 21: 688, 23: 688, 31: 688, 191: 688, 223: 688},
		{110: 1889},
		{689, 689, 21: 689, 23: 689, 31: 689, 191: 689, 223: 689},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 1767, 288: 1891},
		{1: 695},
		// 1150
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1927, 307: 1935, 423: 1934},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1927, 307: 1926},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 1922, 189: 864, 1921, 308: 1923},
		{241: 1919},
		{99: 1902},
		// 1155
		{503, 2: 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 25: 503, 503, 503, 503, 503, 503, 32: 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503, 505, 503, 503, 503, 503, 503, 503, 503, 503, 503, 503},
		{502, 2: 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 25: 502, 502, 502, 502, 502, 502, 32: 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502, 504, 502, 502, 502, 502, 502, 502, 502, 502, 502, 502},
		{422, 2: 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 25: 422, 422, 422, 422, 422, 422, 32: 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 122: 422, 146: 422, 422, 184: 422, 422, 188: 422, 192: 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 422, 271: 422, 287: 422, 365: 1900},
		{423, 2: 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 25: 423, 423, 423, 423, 423, 423, 32: 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 122: 423, 146: 423, 423, 184: 423, 423, 188: 423, 192: 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 423, 271: 423, 287: 1901},
		{421, 2: 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 25: 421, 421, 421, 421, 421, 421, 32: 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 122: 421, 146: 421, 421, 184: 421, 421, 188: 421, 192: 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 421, 271: 421, 287: 421},
		// 1160
		{48: 1905, 226: 1906, 346: 1904, 437: 1903},
		{1: 703, 31: 1917},
		{1: 529, 31: 529},
		{75: 1909},
		{83: 1908, 234: 1907},
		// 1165
		{1: 526, 31: 526},
		{1: 525, 31: 525},
		{91: 1911, 93: 1913, 226: 1912, 438: 1910},
		{1: 527, 31: 527},
		{226: 1916},
		// 1170
		{64: 1914, 101: 1915},
		{1: 521, 31: 521},
		{1: 523, 31: 523},
		{1: 522, 31: 522},
		{1: 524, 31: 524},
		// 1175
		{48: 1905, 226: 1906, 346: 1918},
		{1: 528, 31: 528},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 1922, 189: 864, 1921, 308: 1920},
		{1: 704},
		{1: 108, 258: 108},
		// 1180
		{1: 107, 258: 107},
		{1: 706, 258: 1924},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 1922, 189: 864, 1921, 308: 1925},
		{1: 705},
		{1: 707},
		// 1185
		{161: 1928},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 158: 1931, 184: 812, 806, 188: 859, 864, 854, 1930, 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1933, 412: 1932, 422: 1929},
		{1: 700, 31: 700},
		{1: 699, 31: 699},
		{1: 698, 31: 698},
		// 1190
		{1: 697, 31: 697},
		{1: 696, 31: 696, 133: 985, 145: 983, 148: 984},
		{1: 708, 31: 1936},
		{1: 702, 31: 702},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 854, 219: 1927, 307: 1937},
		// 1195
		{1: 701, 31: 701},
		{1: 709},
		{24: 1471, 118: 1178, 1179, 1177, 1176, 274: 1174},
		{114: 1948, 160: 1947},
		{31: 714, 114: 714, 263: 714, 714},
		// 1200
		{31: 1945, 114: 712, 263: 712, 712},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 1940, 304: 1941, 316: 1944},
		{31: 1945, 114: 711, 263: 711, 711},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 189: 864, 1940, 304: 1946},
		{31: 713, 114: 713, 263: 713, 713},
		// 1205
		{114: 1958},
		{3: 1950, 313: 1949},
		{24: 1951, 31: 1952},
		{24: 399, 31: 399},
		{160: 1954},
		// 1210
		{3: 1953},
		{24: 398, 31: 398},
		{114: 1955},
		{114: 765, 263: 762, 764, 268: 1162, 1956, 763},
		{24: 1957, 118: 1178, 1179, 1177, 1176, 274: 1174},
		// 1215
		{31: 715, 114: 715, 263: 715, 715},
		{114: 765, 263: 762, 764, 268: 1162, 1959, 763},
		{24: 1960, 118: 1178, 1179, 1177, 1176, 274: 1174},
		{31: 716, 114: 716, 263: 716, 716},
		{424, 2: 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 25: 424, 424, 424, 424, 424, 424, 32: 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 122: 424, 146: 424, 424, 184: 424, 424, 188: 424, 192: 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 424, 271: 424, 287: 424, 300: 1899, 303: 1965},
		// 1220
		{118: 1178, 1179, 1177, 1176, 274: 1963},
		{114: 765, 263: 762, 764, 268: 1162, 1964, 763},
		{1: 719, 24: 719, 118: 719, 719, 719, 719, 274: 1174},
		{420, 2: 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 25: 420, 420, 420, 420, 420, 420, 32: 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 122: 420, 146: 1968, 420, 184: 420, 420, 188: 420, 192: 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 271: 1967, 333: 1966},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 1971},
		// 1225
		{419, 2: 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 25: 419, 419, 419, 419, 419, 419, 32: 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 122: 419, 146: 1970, 419, 184: 419, 419, 188: 419, 192: 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419},
		{418, 2: 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 25: 418, 418, 418, 418, 418, 418, 32: 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 122: 418, 147: 418, 184: 418, 418, 188: 418, 192: 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 271: 1969},
		{416, 2: 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 25: 416, 416, 416, 416, 416, 416, 32: 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 122: 416, 147: 416, 184: 416, 416, 188: 416, 192: 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416},
		{417, 2: 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 25: 417, 417, 417, 417, 417, 417, 32: 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 122: 417, 147: 417, 184: 417, 417, 188: 417, 192: 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417},
		{31: 1005, 118: 117, 117, 117, 117, 124: 1559, 149: 1973, 278: 1972},
		// 1230
		{1: 722, 24: 722, 118: 722, 722, 722, 722},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 114: 1979, 122: 1981, 187: 1459, 189: 864, 1458, 218: 1983, 266: 1982, 283: 1980, 1978, 1977, 317: 1976, 1974, 342: 1975},
		{1: 721, 24: 721, 118: 721, 721, 721, 721},
		{1: 361, 24: 361, 31: 2047, 118: 361, 361, 361, 361, 123: 361, 361, 1591, 128: 361, 361, 151: 361, 361, 294: 2046},
		{1: 433, 24: 433, 118: 433, 433, 433, 433},
		// 1235
		{1: 397, 24: 397, 31: 397, 111: 1990, 1991, 118: 397, 397, 397, 397, 123: 397, 397, 397, 128: 397, 397, 132: 1989, 146: 1988, 151: 397, 397, 1993, 1992, 1994, 289: 1987},
		{877, 386, 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 386, 938, 956, 940, 967, 968, 971, 386, 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 1016, 386, 386, 118: 386, 386, 386, 386, 123: 386, 386, 386, 128: 386, 386, 132: 386, 146: 386, 151: 386, 386, 386, 386, 386, 386, 386, 386, 160: 1015, 189: 864, 1014, 221: 386, 224: 386, 386, 290: 2026},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 114: 2023, 122: 1981, 187: 1459, 189: 864, 1458, 218: 1983, 263: 762, 764, 266: 1982, 268: 1162, 1163, 763, 283: 1980, 1978, 2024},
		{1: 393, 24: 393, 31: 393, 111: 393, 393, 118: 393, 393, 393, 393, 123: 393, 393, 393, 128: 393, 393, 132: 393, 146: 393, 151: 393, 393, 393, 393, 393, 393, 393, 393},
		{413: 1984},
		// 1240
		{369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 369, 118: 369, 369, 369, 369, 123: 369, 369, 369, 128: 369, 369, 132: 369, 146: 369, 151: 369, 369, 369, 369, 369, 369, 369, 369, 160: 369, 221: 369, 224: 369, 369},
		{368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 368, 118: 368, 368, 368, 368, 123: 368, 368, 368, 128: 368, 368, 132: 368, 146: 368, 151: 368, 368, 368, 368, 368, 368, 368, 368, 160: 368, 221: 368, 224: 368, 368},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 114: 1979, 122: 1981, 187: 1459, 189: 864, 1458, 218: 1983, 266: 1982, 283: 1985, 1978, 1986},
		{111: 393, 393, 132: 393, 146: 393, 153: 393, 393, 393, 157: 2022},
		{111: 1990, 1991, 132: 1989, 146: 1988, 153: 1993, 1992, 1994, 289: 1987},
		// 1245
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 114: 1979, 122: 1981, 187: 1459, 189: 864, 1458, 218: 1983, 266: 1982, 283: 1980, 1978, 2015},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 114: 1979, 122: 1981, 187: 1459, 189: 864, 1458, 218: 1983, 266: 1982, 283: 1980, 1978, 2012},
		{381, 2: 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 25: 381, 381, 381, 381, 381, 381, 32: 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 114: 381, 122: 381, 187: 381},
		{132: 2009, 299: 2010},
		{132: 2006, 299: 2007},
		// 1250
		{132: 2005},
		{132: 2004},
		{111: 1997, 1996, 132: 1995},
		{374, 2: 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 25: 374, 374, 374, 374, 374, 374, 32: 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 374, 114: 374, 122: 374, 187: 374},
		{132: 2001, 299: 2002},
		// 1255
		{132: 1998, 299: 1999},
		{371, 2: 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 25: 371, 371, 371, 371, 371, 371, 32: 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 371, 114: 371, 122: 371, 187: 371},
		{132: 2000},
		{370, 2: 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 25: 370, 370, 370, 370, 370, 370, 32: 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 370, 114: 370, 122: 370, 187: 370},
		{373, 2: 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 25: 373, 373, 373, 373, 373, 373, 32: 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 373, 114: 373, 122: 373, 187: 373},
		// 1260
		{132: 2003},
		{372, 2: 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 25: 372, 372, 372, 372, 372, 372, 32: 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 372, 114: 372, 122: 372, 187: 372},
		{375, 2: 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 25: 375, 375, 375, 375, 375, 375, 32: 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 375, 114: 375, 122: 375, 187: 375},
		{376, 2: 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 25: 376, 376, 376, 376, 376, 376, 32: 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 376, 114: 376, 122: 376, 187: 376},
		{378, 2: 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 25: 378, 378, 378, 378, 378, 378, 32: 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 378, 114: 378, 122: 378, 187: 378},
		// 1265
		{132: 2008},
		{377, 2: 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 25: 377, 377, 377, 377, 377, 377, 32: 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 377, 114: 377, 122: 377, 187: 377},
		{380, 2: 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 25: 380, 380, 380, 380, 380, 380, 32: 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 114: 380, 122: 380, 187: 380},
		{132: 2011},
		{379, 2: 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 25: 379, 379, 379, 379, 379, 379, 32: 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 379, 114: 379, 122: 379, 187: 379},
		// 1270
		{1: 390, 24: 390, 31: 390, 111: 390, 390, 118: 390, 390, 390, 390, 123: 390, 390, 390, 128: 390, 390, 132: 390, 146: 390, 151: 390, 390, 390, 390, 1994, 390, 390, 2013, 289: 1987},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 2014},
		{1: 388, 24: 388, 31: 388, 111: 388, 388, 118: 388, 388, 388, 388, 123: 388, 388, 388, 128: 388, 388, 132: 388, 985, 145: 983, 388, 148: 984, 151: 388, 388, 388, 388, 388, 388, 388, 388},
		{1: 391, 24: 391, 31: 391, 111: 391, 391, 118: 391, 391, 391, 391, 123: 391, 391, 391, 128: 391, 391, 132: 391, 146: 391, 151: 391, 391, 391, 391, 1994, 2017, 391, 2016, 289: 1987},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 2021},
		// 1275
		{114: 2018},
		{3: 1950, 313: 2019},
		{24: 2020, 31: 1952},
		{1: 387, 24: 387, 31: 387, 111: 387, 387, 118: 387, 387, 387, 387, 123: 387, 387, 387, 128: 387, 387, 132: 387, 146: 387, 151: 387, 387, 387, 387, 387, 387, 387, 387},
		{1: 389, 24: 389, 31: 389, 111: 389, 389, 118: 389, 389, 389, 389, 123: 389, 389, 389, 128: 389, 389, 132: 389, 985, 145: 983, 389, 148: 984, 151: 389, 389, 389, 389, 389, 389, 389, 389},
		// 1280
		{1: 392, 24: 392, 31: 392, 111: 392, 392, 118: 392, 392, 392, 392, 123: 392, 392, 392, 128: 392, 392, 132: 392, 146: 392, 151: 392, 392, 392, 392, 392, 392, 392, 392},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 114: 2023, 122: 1981, 187: 1459, 189: 864, 1458, 218: 1983, 263: 762, 764, 266: 1982, 268: 1162, 1172, 763, 283: 1980, 1978, 2024},
		{24: 2025, 111: 1990, 1991, 132: 1989, 146: 1988, 153: 1993, 1992, 1994, 289: 1987},
		{1: 394, 24: 394, 31: 394, 111: 394, 394, 118: 394, 394, 394, 394, 123: 394, 394, 394, 128: 394, 394, 132: 394, 146: 394, 151: 394, 394, 394, 394, 394, 394, 394, 394},
		{1: 367, 24: 367, 31: 367, 111: 367, 367, 118: 367, 367, 367, 367, 123: 367, 367, 367, 128: 367, 367, 132: 367, 146: 367, 151: 367, 367, 367, 367, 367, 367, 367, 367, 221: 2028, 224: 2030, 2029, 394: 2027},
		// 1285
		{1: 395, 24: 395, 31: 395, 111: 395, 395, 118: 395, 395, 395, 395, 123: 395, 395, 395, 128: 395, 395, 132: 395, 146: 395, 151: 395, 395, 395, 395, 395, 395, 395, 395},
		{276: 2042},
		{276: 2038},
		{276: 2031},
		{114: 2032},
		// 1290
		{3: 2034, 306: 2033},
		{24: 2035, 31: 2036},
		{24: 363, 31: 363},
		{1: 364, 24: 364, 31: 364, 111: 364, 364, 118: 364, 364, 364, 364, 123: 364, 364, 364, 128: 364, 364, 132: 364, 146: 364, 151: 364, 364, 364, 364, 364, 364, 364, 364},
		{3: 2037},
		// 1295
		{24: 362, 31: 362},
		{114: 2039},
		{3: 2034, 306: 2040},
		{24: 2041, 31: 2036},
		{1: 365, 24: 365, 31: 365, 111: 365, 365, 118: 365, 365, 365, 365, 123: 365, 365, 365, 128: 365, 365, 132: 365, 146: 365, 151: 365, 365, 365, 365, 365, 365, 365, 365},
		// 1300
		{114: 2043},
		{3: 2034, 306: 2044},
		{24: 2045, 31: 2036},
		{1: 366, 24: 366, 31: 366, 111: 366, 366, 118: 366, 366, 366, 366, 123: 366, 366, 366, 128: 366, 366, 132: 366, 146: 366, 151: 366, 366, 366, 366, 366, 366, 366, 366},
		{1: 129, 24: 129, 118: 129, 129, 129, 129, 123: 129, 129, 128: 129, 129, 151: 129, 2050, 323: 2049},
		// 1305
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 114: 1979, 122: 1981, 187: 1459, 189: 864, 1458, 218: 1983, 266: 1982, 283: 1980, 1978, 2048},
		{1: 396, 24: 396, 31: 396, 111: 1990, 1991, 118: 396, 396, 396, 396, 123: 396, 396, 396, 128: 396, 396, 132: 1989, 146: 1988, 151: 396, 396, 1993, 1992, 1994, 289: 1987},
		{1: 127, 24: 127, 118: 127, 127, 127, 127, 123: 127, 127, 128: 127, 127, 151: 2054, 324: 2053},
		{311: 2051},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1173, 291: 2052},
		// 1310
		{1: 128, 24: 128, 31: 1169, 118: 128, 128, 128, 128, 123: 128, 128, 128: 128, 128, 151: 128},
		{1: 125, 24: 125, 118: 125, 125, 125, 125, 123: 125, 125, 128: 125, 1232, 298: 2056},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 184: 812, 806, 188: 859, 864, 854, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 2055},
		{1: 126, 24: 126, 118: 126, 126, 126, 126, 123: 126, 126, 128: 126, 126, 133: 985, 145: 983, 148: 984},
		{1: 117, 24: 117, 118: 117, 117, 117, 117, 123: 117, 1559, 128: 117, 278: 2057},
		// 1315
		{1: 113, 24: 113, 118: 113, 113, 113, 113, 123: 2059, 128: 2060, 330: 2058},
		{1: 718, 24: 718, 118: 720, 720, 720, 720},
		{443: 2064},
		{175: 2061},
		{3: 2062},
		// 1320
		{3: 2063},
		{1: 111, 24: 111, 118: 111, 111, 111, 111},
		{1: 112, 24: 112, 118: 112, 112, 112, 112},
		{420, 2: 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 25: 420, 420, 420, 420, 420, 420, 32: 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 122: 420, 146: 1968, 420, 184: 420, 420, 188: 420, 192: 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 420, 271: 1967, 333: 2066},
		{877, 2: 867, 849, 848, 835, 836, 838, 839, 841, 842, 844, 846, 930, 931, 932, 933, 934, 935, 936, 937, 873, 856, 884, 25: 938, 834, 940, 858, 845, 863, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 832, 882, 883, 887, 894, 899, 910, 912, 790, 926, 941, 942, 847, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 826, 827, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 855, 815, 816, 850, 791, 840, 792, 851, 122: 807, 147: 999, 184: 812, 806, 188: 859, 864, 1001, 192: 837, 862, 861, 833, 803, 796, 857, 814, 843, 817, 818, 813, 819, 820, 821, 822, 823, 824, 825, 784, 785, 828, 831, 830, 853, 852, 801, 799, 860, 222: 798, 227: 800, 805, 811, 810, 808, 809, 829, 235: 804, 797, 802, 254: 795, 259: 794, 793, 1000, 997, 265: 2067},
		// 1325
		{1: 117, 24: 117, 31: 1005, 118: 117, 117, 117, 117, 124: 1559, 149: 2068, 278: 1972},
		{877, 2: 867, 953, 974, 957, 958, 959, 960, 961, 963, 966, 973, 930, 931, 932, 933, 934, 935, 936, 937, 873, 962, 884, 25: 938, 956, 940, 967, 968, 971, 32: 879, 881, 891, 946, 947, 876, 885, 889, 903, 918, 874, 955, 882, 883, 887, 894, 899, 910, 912, 965, 926, 941, 942, 972, 949, 866, 868, 869, 870, 871, 872, 875, 878, 954, 880, 886, 888, 893, 895, 896, 897, 898, 900, 901, 902, 904, 905, 906, 907, 908, 911, 913, 914, 916, 917, 919, 920, 921, 922, 923, 924, 925, 927, 928, 939, 969, 970, 943, 944, 945, 948, 865, 890, 892, 909, 915, 964, 929, 114: 1979, 122: 1981, 187: 1459, 189: 864, 1458, 218: 1983, 266: 1982, 283: 1980, 1978, 1977, 317: 1976, 1974, 342: 2069},
		{1: 361, 24: 361, 31: 2047, 118: 361, 361, 361, 361, 123: 361, 361, 1591, 128: 361, 361, 151: 361, 361, 294: 2070},
		{1: 129, 24: 129, 118: 129, 129, 129, 129, 123: 129, 129, 128: 129, 129, 151: 129, 2050, 323: 2071},
		{1: 127, 24: 127, 118: 127, 127, 127, 127, 123: 127, 127, 128: 127, 127, 151: 2054, 324: 2072},
		// 1330
		{1: 125, 24: 125, 118: 125, 125, 125, 125, 123: 125, 125, 128: 125, 1232, 298: 2073},
		{1: 117, 24: 117, 118: 117, 117, 117, 117, 123: 117, 1559, 128: 117, 278: 2074},
		{1: 113, 24: 113, 118: 113, 113, 113, 113, 123: 2059, 128: 2060, 330: 2075},
		{1: 720, 24: 720, 118: 720, 720, 720, 720},
		{1: 117, 124: 1559, 278: 2077},
		// 1335
		{1: 724},
	}
)

var yyDebug = 0

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyLexerEx interface {
	yyLexer
	Reduced(rule, state int, lval *yySymType) bool
}

func yySymName(c int) (s string) {
	x, ok := yyXLAT[c]
	if ok {
		return yySymNames[x]
	}

	if c < 0x7f {
		return __yyfmt__.Sprintf("%q", c)
	}

	return __yyfmt__.Sprintf("%d", c)
}

func yylex1(yylex yyLexer, lval *yySymType) (n int) {
	n = yylex.Lex(lval)
	if n <= 0 {
		n = yyEofCode
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("\nlex %s(%#x %d), lval: %+v\n", yySymName(n), n, n, lval)
	}
	return n
}

func yyParse(yylex yyLexer) int {
	const yyError = 451

	yyEx, _ := yylex.(yyLexerEx)
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	yyS := make([]yySymType, 200)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yyerrok := func() {
		if yyDebug >= 2 {
			__yyfmt__.Printf("yyerrok()\n")
		}
		Errflag = 0
	}
	_ = yyerrok
	yystate := 0
	yychar := -1
	var yyxchar int
	var yyshift int
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	if yychar < 0 {
		yylval.yys = yystate
		yychar = yylex1(yylex, &yylval)
		var ok bool
		if yyxchar, ok = yyXLAT[yychar]; !ok {
			yyxchar = len(yySymNames) // > tab width
		}
	}
	if yyDebug >= 4 {
		var a []int
		for _, v := range yyS[:yyp+1] {
			a = append(a, v.yys)
		}
		__yyfmt__.Printf("state stack %v\n", a)
	}
	row := yyParseTab[yystate]
	yyn = 0
	if yyxchar < len(row) {
		if yyn = int(row[yyxchar]); yyn != 0 {
			yyn += yyTabOfs
		}
	}
	switch {
	case yyn > 0: // shift
		yychar = -1
		yyVAL = yylval
		yystate = yyn
		yyshift = yyn
		if yyDebug >= 2 {
			__yyfmt__.Printf("shift, and goto state %d\n", yystate)
		}
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	case yyn < 0: // reduce
	case yystate == 1: // accept
		if yyDebug >= 2 {
			__yyfmt__.Println("accept")
		}
		goto ret0
	}

	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			if yyDebug >= 1 {
				__yyfmt__.Printf("no action for %s in state %d\n", yySymName(yychar), yystate)
			}
			msg, ok := yyXErrors[yyXError{yystate, yyxchar}]
			if !ok {
				msg, ok = yyXErrors[yyXError{yystate, -1}]
			}
			if !ok && yyshift != 0 {
				msg, ok = yyXErrors[yyXError{yyshift, yyxchar}]
			}
			if !ok {
				msg, ok = yyXErrors[yyXError{yyshift, -1}]
			}
			if yychar > 0 {
				ls := yyTokenLiteralStrings[yychar]
				if ls == "" {
					ls = yySymName(yychar)
				}
				if ls != "" {
					switch {
					case msg == "":
						msg = __yyfmt__.Sprintf("unexpected %s", ls)
					default:
						msg = __yyfmt__.Sprintf("unexpected %s, %s", ls, msg)
					}
				}
			}
			if msg == "" {
				msg = "syntax error"
			}
			yylex.Error(msg)
			Nerrs++
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				row := yyParseTab[yyS[yyp].yys]
				if yyError < len(row) {
					yyn = int(row[yyError]) + yyTabOfs
					if yyn > 0 { // hit
						if yyDebug >= 2 {
							__yyfmt__.Printf("error recovery found error shift in state %d\n", yyS[yyp].yys)
						}
						yystate = yyn /* simulate a shift of "error" */
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
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery failed\n")
			}
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yySymName(yychar))
			}
			if yychar == yyEofCode {
				goto ret1
			}

			yychar = -1
			goto yynewstate /* try again in the same state */
		}
	}

	r := -yyn
	x0 := yyReductions[r]
	x, n := x0.xsym, x0.components
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= n
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	exState := yystate
	yystate = int(yyParseTab[yyS[yyp].yys][x]) + yyTabOfs
	/* reduction by production r */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce using rule %v (%s), and goto state %d\n", r, yySymNames[x], yystate)
	}

	switch r {
	case 1:
		{
			SetParseTree(yylex, yyS[yypt-0].statement)
		}
	case 2:
		{
			yyVAL.statement = yyS[yypt-0].selStmt
		}
	case 3:
		{
			yyVAL.statement = yyS[yypt-0].selStmt
		}
	case 18:
		{
			yyVAL.selStmt = &Select{QueryGlobals: &QueryGlobals{}, SelectExprs: []SelectExpr{&StarExpr{}}, From: []TableExpr{&AliasedTableExpr{Expr: yyS[yypt-2].subquery}}, OrderBy: yyS[yypt-1].orderBy, Limit: yyS[yypt-0].limit}
		}
	case 19:
		{
			yyVAL.selStmt = &Select{QueryGlobals: &QueryGlobals{}, SelectExprs: []SelectExpr{&StarExpr{}}, From: []TableExpr{&AliasedTableExpr{Expr: yyS[yypt-0].subquery}}}
		}
	case 20:
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyS[yypt-3].strs), QueryGlobals: yyS[yypt-2].queryGlobals, SelectExprs: yyS[yypt-1].selectExprs, Limit: yyS[yypt-0].limit}
		}
	case 21:
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyS[yypt-4].strs), QueryGlobals: yyS[yypt-3].queryGlobals, SelectExprs: yyS[yypt-2].selectExprs, From: yyS[yypt-0].tableExprs}
		}
	case 22:
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyS[yypt-10].strs), QueryGlobals: yyS[yypt-9].queryGlobals, SelectExprs: yyS[yypt-8].selectExprs, From: yyS[yypt-6].tableExprs, Where: NewWhere(AST_WHERE, yyS[yypt-5].expr), GroupBy: GroupBy(yyS[yypt-4].exprs), Having: NewWhere(AST_HAVING, yyS[yypt-3].expr), OrderBy: yyS[yypt-2].orderBy, Limit: yyS[yypt-1].limit, Lock: yyS[yypt-0].str}
		}
	case 23:
		{
			yyVAL.selStmt = &Union{With: yyS[yypt-3].with, Left: yyS[yypt-2].selStmt, Right: yyS[yypt-0].selStmt, Type: yyS[yypt-1].str}
		}
	case 24:
		{
			yyVAL.selStmt = &Select{With: yyS[yypt-12].with, Comments: Comments(yyS[yypt-10].strs), QueryGlobals: yyS[yypt-9].queryGlobals, SelectExprs: yyS[yypt-8].selectExprs, From: yyS[yypt-6].tableExprs, Where: NewWhere(AST_WHERE, yyS[yypt-5].expr), GroupBy: GroupBy(yyS[yypt-4].exprs), Having: NewWhere(AST_HAVING, yyS[yypt-3].expr), OrderBy: yyS[yypt-2].orderBy, Limit: yyS[yypt-1].limit, Lock: yyS[yypt-0].str}
		}
	case 25:
		{
			yyVAL.selStmt = &Union{Type: yyS[yypt-1].str, Left: yyS[yypt-2].selStmt, Right: yyS[yypt-0].selStmt}
		}
	case 26:
		{
			yyVAL.cte = &CTE{TableName: &TableName{Name: yyS[yypt-4].str}, ColumnExprs: nil, Query: yyS[yypt-1].selStmt}
		}
	case 27:
		{
			yyVAL.cte = &CTE{TableName: &TableName{Name: yyS[yypt-7].str}, ColumnExprs: yyS[yypt-5].columnExprs, Query: yyS[yypt-1].selStmt}
		}
	case 28:
		{
			yyVAL.cte_list = []*CTE{yyS[yypt-0].cte}
		}
	case 29:
		{
			yyVAL.cte_list = append(yyS[yypt-2].cte_list, yyS[yypt-0].cte)
		}
	case 30:
		{
			yyVAL.with = &With{CTEs: yyS[yypt-0].cte_list, Recursive: false}
		}
	case 31:
		{
			yyVAL.with = &With{CTEs: yyS[yypt-0].cte_list, Recursive: true}
		}
	case 32:
		{
			yyVAL.subquery = &Subquery{Select: yyS[yypt-1].selStmt, IsDerived: false}
		}
	case 33:
		{
			yyVAL.statement = &Use{DBName: string(yyS[yypt-0].bytes)}
		}
	case 34:
		{
			yyVAL.statement = &Set{Comments: Comments(yyS[yypt-1].strs), Exprs: yyS[yypt-0].updateExprs}
		}
	case 35:
		{
			yyVAL.statement = &Set{Scope: yyS[yypt-1].str, Exprs: UpdateExprs(append([]*UpdateExpr{}, yyS[yypt-0].updateExpr))}
		}
	case 36:
		{
			yyVAL.statement = &Set{Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: "@@character_set_client"}, Expr: StrVal(yyS[yypt-0].str)},
				&UpdateExpr{Name: &ColName{Name: "@@character_set_results"}, Expr: StrVal(yyS[yypt-0].str)},
				&UpdateExpr{Name: &ColName{Name: "@@character_set_connection"}, Expr: StrVal(yyS[yypt-0].str)},
			}}
		}
	case 37:
		{
			yyVAL.statement = &Set{Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: "@@character_set_client"}, Expr: StrVal(yyS[yypt-2].str)},
				&UpdateExpr{Name: &ColName{Name: "@@character_set_results"}, Expr: StrVal(yyS[yypt-2].str)},
				&UpdateExpr{Name: &ColName{Name: "@@character_set_connection"}, Expr: StrVal(yyS[yypt-2].str)},
				&UpdateExpr{Name: &ColName{Name: "@@collation_connection"}, Expr: StrVal(yyS[yypt-0].str)},
			}}
		}
	case 38:
		{
			yyVAL.statement = &Set{Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: "@@character_set_client"}, Expr: StrVal(yyS[yypt-0].str)},
				&UpdateExpr{Name: &ColName{Name: "@@character_set_results"}, Expr: StrVal(yyS[yypt-0].str)},
				&UpdateExpr{Name: &ColName{Name: "@@collation_connection"}, Expr: &ColName{Name: "@@collation_database"}},
			}}
		}
	case 39:
		{
			yyVAL.statement = &Set{Comments: append([]string{}, yyS[yypt-2].str, string(TRANSACTION_BYTES), yyS[yypt-0].str)}
		}
	case 40:
		{
			yyVAL.updateExprs = UpdateExprs{yyS[yypt-0].updateExpr}
		}
	case 41:
		{
			yyVAL.updateExprs = append(yyS[yypt-2].updateExprs, yyS[yypt-0].updateExpr)
		}
	case 42:
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyS[yypt-2].colName, Expr: yyS[yypt-0].expr}
		}
	case 43:
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyS[yypt-2].colName, Expr: StrVal("default")}
		}
	case 44:
		{
			yyVAL.expr = NumVal([]byte("1"))
		}
	case 45:
		{
			yyVAL.expr = NumVal([]byte("0"))
		}
	case 46:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 47:
		{
			yyVAL.statement = &CreateDatabase{Name: yyS[yypt-0].str, IfNotExists: yyS[yypt-1].bool}
		}
	case 48:
		{
			if yyS[yypt-7].bool {
				yylex.Error("temporary tables are not supported")
				return 1
			}
			yyVAL.statement = &CreateTable{Name: yyS[yypt-4].tableName, IfNotExists: yyS[yypt-5].bool, Definitions: yyS[yypt-2].columnOrIndexDefs, TableOptions: yyS[yypt-0].tableOptions}
		}
	case 49:
		{
			yyVAL.tableOptions = []TableOption{}
		}
	case 50:
		{
			yyVAL.tableOptions = append(yyS[yypt-2].tableOptions, yyS[yypt-0].tableOption)
		}
	case 51:
		{
			yyVAL.empty = struct{}{}
		}
	case 52:
		{
			yyVAL.empty = struct{}{}
		}
	case 53:
		{
			yyVAL.tableOption = TableComment(string(yyS[yypt-0].bytes))
		}
	case 54:
		{
			yyVAL.tableOption = IgnoredTableOption{}
		}
	case 55:
		{
			yyVAL.tableOption = IgnoredTableOption{}
		}
	case 56:
		{
			yyVAL.tableOption = IgnoredTableOption{}
		}
	case 57:
		{
			yyVAL.empty = struct{}{}
		}
	case 58:
		{
			yyVAL.empty = struct{}{}
		}
	case 59:
		{
			yyVAL.empty = struct{}{}
		}
	case 60:
		{
			yyVAL.empty = struct{}{}
		}
	case 61:
		{
			yyVAL.columnOrIndexDefs = []ColumnOrIndexDefinition{yyS[yypt-0].columnOrIndexDef}
		}
	case 62:
		{
			yyVAL.columnOrIndexDefs = append(yyS[yypt-2].columnOrIndexDefs, yyS[yypt-0].columnOrIndexDef)
		}
	case 63:
		{
			if yyS[yypt-1].bool {
				yylex.Error("PRIMARY KEYS are not supported at this time")
				return 1
			}
			if yyS[yypt-3].bool {
				yylex.Error("AUTO_INCREMENT is not supported at this time")
				return 1
			}
			yyVAL.columnOrIndexDef = &ColumnDefinition{
				Name:    yyS[yypt-7].colName,
				Type:    yyS[yypt-6].colTy,
				Null:    yyS[yypt-5].bool,
				Unique:  yyS[yypt-2].bool,
				Comment: yyS[yypt-0].stropt,
			}
		}
	case 64:
		{
			if yyS[yypt-7].bool && yyS[yypt-6].bool {
				yylex.Error("indexes cannot be both UNIQUE and FULLTEXT")
				return 1
			}
			yyVAL.columnOrIndexDef = &IndexDefinition{Name: yyS[yypt-4].stropt, Unique: yyS[yypt-6].bool, FullText: yyS[yypt-7].bool, KeyParts: yyS[yypt-1].keyPartList}
		}
	case 65:
		{
			yyVAL.colTy = ColumnType{BaseType: string(yyS[yypt-0].bytes), Width: option.NoneInt()}
		}
	case 66:
		{
			yyVAL.colTy = ColumnType{BaseType: string(yyS[yypt-1].bytes), Width: yyS[yypt-0].intopt}
		}
	case 67:
		{
			yyVAL.colTy = ColumnType{BaseType: string(yyS[yypt-1].bytes), Width: option.NoneInt()}
		}
	case 68:
		{
			yyVAL.colTy = ColumnType{BaseType: string(yyS[yypt-0].bytes), Width: option.NoneInt()}
		}
	case 69:
		{
			yyVAL.bytes = BOOL_BYTES
		}
	case 70:
		{
			yyVAL.bytes = BOOLEAN_BYTES
		}
	case 71:
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 72:
		{
			yylex.Error("SERIAL is not supported")
			return 1
		}
	case 73:
		{
			yylex.Error("YEAR is not supported")
			return 1
		}
	case 74:
		{
			yylex.Error("MEDIUMBLOB is not supported")
			return 1
		}
	case 75:
		{
			yyVAL.bytes = BIT_BYTES
		}
	case 76:
		{
			yyVAL.bytes = TINYINT_BYTES
		}
	case 77:
		{
			yyVAL.bytes = SMALLINT_BYTES
		}
	case 78:
		{
			yyVAL.bytes = INT_BYTES
		}
	case 79:
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 80:
		{
			yyVAL.bytes = BIGINT_BYTES
		}
	case 81:
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 82:
		{
			yyVAL.bytes = TIMESTAMP_BYTES
		}
	case 83:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 84:
		{
			yyVAL.bytes = VARCHAR_BYTES
		}
	case 85:
		{
			yyVAL.bytes = TEXT_BYTES
		}
	case 86:
		{
			yylex.Error("BINARY is not supported")
			return 1
		}
	case 87:
		{
			yyVAL.bytes = TINYTEXT_BYTES
		}
	case 88:
		{
			yyVAL.bytes = MEDIUMTEXT_BYTES
		}
	case 89:
		{
			yylex.Error("BLOB is not supported")
			return 1
		}
	case 90:
		{
			yyVAL.bytes = LONGTEXT_BYTES
		}
	case 91:
		{
			yyVAL.bytes = DECIMAL_BYTES
		}
	case 92:
		{
			yyVAL.bytes = NUMERIC_BYTES
		}
	case 93:
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 94:
		{
			yyVAL.bytes = DOUBLE_BYTES
		}
	case 95:
		{
			yyVAL.bytes = DOUBLE_BYTES
		}
	case 96:
		{
			yylex.Error("ENUM is not supported")
			return 1
		}
	case 97:
		{
			yylex.Error("SET is not supported")
			return 1
		}
	case 98:
		{
			yyVAL.intopt = option.NoneInt()
		}
	case 99:
		{
			i, err := strconv.Atoi(string(yyS[yypt-1].bytes))
			if err != nil {
				yylex.Error("width for datatype must be an integer, not a float")
				return 1
			}
			yyVAL.intopt = option.SomeInt(i)
		}
	case 100:
		{
			yyVAL.empty = struct{}{}
		}
	case 101:
		{
			yyVAL.empty = struct{}{}
		}
	case 102:
		{
			yyVAL.bool = true
		}
	case 103:
		{
			yyVAL.bool = true
		}
	case 104:
		{
			yyVAL.bool = false
		}
	case 105:
		{
			yyVAL.empty = struct{}{}
		}
	case 106:
		{
			yyVAL.empty = struct{}{}
		}
	case 107:
		{
			yylex.Error("only NULL defaults are supported")
			return 1
		}
	case 108:
		{
			yyVAL.bool = false
		}
	case 109:
		{
			yyVAL.bool = true
		}
	case 110:
		{
			yyVAL.bool = false
		}
	case 111:
		{
			yyVAL.bool = true
		}
	case 112:
		{
			yyVAL.bool = true
		}
	case 113:
		{
			yyVAL.bool = false
		}
	case 114:
		{
			yyVAL.bool = true
		}
	case 115:
		{
			yyVAL.bool = true
		}
	case 116:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 117:
		{
			yyVAL.stropt = option.SomeString(string(yyS[yypt-0].bytes))
		}
	case 118:
		{
			yyVAL.bool = false
		}
	case 119:
		{
			yyVAL.bool = true
		}
	case 120:
		{
			yyVAL.bool = false
		}
	case 121:
		{
			yyVAL.bool = true
		}
	case 122:
		{
			yyVAL.empty = struct{}{}
		}
	case 123:
		{
			yyVAL.empty = struct{}{}
		}
	case 124:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 125:
		{
			yyVAL.stropt = option.SomeString(string(yyS[yypt-0].str))
		}
	case 126:
		{
			yyVAL.empty = struct{}{}
		}
	case 127:
		{
			yylex.Error("index types are not supported")
		}
	case 128:
		{
			yylex.Error("index types are not supported")
		}
	case 129:
		{
			yyVAL.keyPartList = []KeyPart{yyS[yypt-0].keyPart}
		}
	case 130:
		{
			yyVAL.keyPartList = append(yyS[yypt-2].keyPartList, yyS[yypt-0].keyPart)
		}
	case 131:
		{
			yyVAL.keyPart = KeyPart{Column: yyS[yypt-0].colName, Direction: 1}
		}
	case 132:
		{
			yyVAL.keyPart = KeyPart{Column: yyS[yypt-1].colName, Direction: 1}
		}
	case 133:
		{
			yyVAL.keyPart = KeyPart{Column: yyS[yypt-1].colName, Direction: -1}
		}
	case 134:
		{
			yyVAL.statement = &DropDatabase{Name: yyS[yypt-0].str, IfExists: yyS[yypt-1].bool}
		}
	case 135:
		{
			yyVAL.statement = &DropTable{Name: yyS[yypt-1].tableName, IfExists: yyS[yypt-2].bool, Opt: yyS[yypt-0].stropt}
		}
	case 136:
		{
			yyVAL.statement = &Flush{Kind: FlushLogs}
		}
	case 137:
		{
			yyVAL.statement = &Flush{Kind: FlushSample}
		}
	case 138:
		{
			yyVAL.statement = &RenameTable{Renames: yyS[yypt-0].renameSpecs}
		}
	case 139:
		{
			yyVAL.renameSpecs = []*RenameSpec{yyS[yypt-0].renameSpec}
		}
	case 140:
		{
			yyVAL.renameSpecs = append(yyS[yypt-2].renameSpecs, yyS[yypt-0].renameSpec)
		}
	case 141:
		{
			yyVAL.renameSpec = &RenameSpec{Table: yyS[yypt-2].tableName, NewTable: yyS[yypt-0].tableName}
		}
	case 142:
		{
			yyVAL.statement = &IgnoredStatement{Statement: LockTables{LockList: yyS[yypt-0].tableLocks}}
		}
	case 143:
		{
			yyVAL.statement = &IgnoredStatement{Statement: UnlockTables{}}
		}
	case 144:
		{
			yyVAL.statement = &IgnoredStatement{Statement: EnableKeys{}}
		}
	case 145:
		{
			yyVAL.statement = &IgnoredStatement{Statement: DisableKeys{}}
		}
	case 146:
		{
			yyVAL.tableLocks = []TableLock{yyS[yypt-0].tableLock}
		}
	case 147:
		{
			yyVAL.tableLocks = append(yyS[yypt-2].tableLocks, yyS[yypt-0].tableLock)
		}
	case 148:
		{
			yyVAL.tableLock = TableLock{TableName: yyS[yypt-2].tableName, Alias: yyS[yypt-1].stropt, LockType: yyS[yypt-0].lockType}
		}
	case 149:
		{
			yyVAL.lockType = GetLockType(string(READ_BYTES))
		}
	case 150:
		{
			yyVAL.lockType = GetLockType(string(WRITE_BYTES))
		}
	case 155:
		{
			yyVAL.statement = &Insert{Table: yyS[yypt-3].tableName, Columns: yyS[yypt-2].columnList, Values: yyS[yypt-0].valueListList}
		}
	case 158:
		{
			yyVAL.columnList = []*ColName{}
		}
	case 159:
		{
			yyVAL.columnList = []*ColName{}
		}
	case 160:
		{
			yyVAL.columnList = yyS[yypt-1].columnList
		}
	case 161:
		{
			yyVAL.columnList = append(yyS[yypt-2].columnList, yyS[yypt-0].colName)
		}
	case 162:
		{
			yyVAL.columnList = []*ColName{yyS[yypt-0].colName}
		}
	case 165:
		{
			yyVAL.valueListList = append(yyS[yypt-4].valueListList, yyS[yypt-1].valueList)
		}
	case 166:
		{
			yyVAL.valueListList = ValueListList{yyS[yypt-1].valueList}
		}
	case 167:
		{
			yyVAL.valueList = append(yyS[yypt-2].valueList, yyS[yypt-0].val)
		}
	case 168:
		{
			yyVAL.valueList = ValueList{yyS[yypt-0].val}
		}
	case 169:
		{
			yyVAL.valueList = ValueList{}
		}
	case 170:
		{
			yyVAL.val = yyS[yypt-0].val
		}
	case 171:
		{
			yyVAL.val = Default{}
		}
	case 172:
		{
			yyVAL.statement = &AlterTable{Table: yyS[yypt-1].tableName, Specs: yyS[yypt-0].alterSpecs}
		}
	case 173:
		{
			yyVAL.statement = &IgnoredStatement{Statement: EnableKeys{}}
		}
	case 174:
		{
			yyVAL.statement = &IgnoredStatement{Statement: DisableKeys{}}
		}
	case 175:
		{
			yyVAL.alterSpecs = []*AlterSpec{yyS[yypt-0].alterSpec}
		}
	case 176:
		{
			yyVAL.alterSpecs = append(yyS[yypt-2].alterSpecs, yyS[yypt-0].alterSpec)
		}
	case 177:
		{
			yyVAL.alterSpec = &AlterSpec{
				Type:      "rename_column",
				Column:    yyS[yypt-1].colName,
				NewColumn: yyS[yypt-0].colName,
			}
		}
	case 178:
		{
			yyVAL.alterSpec = &AlterSpec{
				Type:   "drop_column",
				Column: yyS[yypt-0].colName,
			}
		}
	case 179:
		{
			yyVAL.alterSpec = &AlterSpec{
				Type:          "modify_column",
				Column:        yyS[yypt-1].colName,
				NewColumnType: string(yyS[yypt-0].bytes),
			}
		}
	case 180:
		{
			yyVAL.alterSpec = &AlterSpec{
				Type:     "rename_table",
				NewTable: yyS[yypt-0].tableName,
			}
		}
	case 181:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 182:
		{
			yyVAL.stropt = option.SomeString(string(COLUMN_BYTES))
		}
	case 183:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 184:
		{
			yyVAL.bytes = TINYINT_BYTES
		}
	case 185:
		{
			yyVAL.bytes = INT_BYTES
		}
	case 186:
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 187:
		{
			yyVAL.bytes = BIGINT_BYTES
		}
	case 188:
		{
			yyVAL.bytes = DOUBLE_BYTES
		}
	case 189:
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 190:
		{
			yyVAL.bytes = DECIMAL_BYTES
		}
	case 191:
		{
			yyVAL.bytes = NUMERIC_BYTES
		}
	case 192:
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 193:
		{
			yyVAL.bytes = TIME_BYTES
		}
	case 194:
		{
			yyVAL.bytes = TIMESTAMP_BYTES
		}
	case 195:
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 196:
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 197:
		{
			yyVAL.bytes = VARCHAR_BYTES
		}
	case 198:
		{
			yyVAL.bytes = BINARY_BYTES
		}
	case 199:
		{
			yyVAL.bytes = TEXT_BYTES
		}
	case 200:
		{
			yyVAL.bytes = BOOLEAN_BYTES
		}
	case 201:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 202:
		{
			yyVAL.stropt = option.SomeString(string(TO_BYTES))
		}
	case 203:
		{
			yyVAL.stropt = option.SomeString(string(AS_BYTES))
		}
	case 204:
		{
			yyVAL.bool = false
		}
	case 205:
		{
			yyVAL.bool = true
		}
	case 206:
		{
			yyVAL.bool = false
		}
	case 207:
		{
			yyVAL.bool = true
		}
	case 208:
		{
			yyVAL.bool = false
		}
	case 209:
		{
			yyVAL.bool = true
		}
	case 210:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 211:
		{
			yyVAL.stropt = option.SomeString(string(CASCADE_BYTES))
		}
	case 212:
		{
			yyVAL.stropt = option.SomeString(string(RESTRICT_BYTES))
		}
	case 213:
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 214:
		{
			yyVAL.str = yyS[yypt-2].str + ", " + yyS[yypt-0].str
		}
	case 215:
		{
			yyVAL.str = "isolation level " + yyS[yypt-0].str
		}
	case 216:
		{
			yyVAL.str = "read write"
		}
	case 217:
		{
			yyVAL.str = "read only"
		}
	case 218:
		{
			yyVAL.str = "repeatable read"
		}
	case 219:
		{
			yyVAL.str = "read committed"
		}
	case 220:
		{
			yyVAL.str = "read uncommitted"
		}
	case 221:
		{
			yyVAL.str = string(SERIALIZABLE_BYTES)
		}
	case 222:
		{
			yyVAL.bytes = SUBSTR_BYTES
		}
	case 223:
		{
			yyVAL.bytes = SUBSTRING_BYTES
		}
	case 226:
		{
			yyVAL.expr = StrVal(yyS[yypt-0].bytes)
		}
	case 227:
		{
			yyVAL.expr = &ColName{Qualifier: option.SomeString(string(yyS[yypt-2].bytes)), Name: string(yyS[yypt-0].bytes)}
		}
	case 228:
		{
			yyVAL.expr = &ColName{Qualifier: option.SomeString(string(yyS[yypt-0].bytes)), Name: string(yyS[yypt-2].bytes)}
		}
	case 229:
		{
			yyVAL.expr = StrVal(yyS[yypt-0].bytes)
		}
	case 230:
		{
			yyVAL.expr = nil
		}
	case 231:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 232:
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 233:
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 234:
		{
			yyVAL.str = AST_KILL_CONNECTION
		}
	case 235:
		{
			yyVAL.str = AST_KILL_QUERY
		}
	case 236:
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 237:
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 238:
		{
			yyVAL.str = AST_GLOBAL_SCOPE
		}
	case 239:
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 240:
		{
			yyVAL.str = AST_GLOBAL_SCOPE
		}
	case 241:
		{
			yyVAL.statement = &Show{Section: "binary logs"}
		}
	case 242:
		{
			yyVAL.statement = &Show{Section: "master logs"}
		}
	case 243:
		{
			yyVAL.statement = &Show{Section: "binlog events"}
		}
	case 244:
		{
			yyVAL.statement = &Show{Section: "create database", Modifier: yyS[yypt-1].str, From: StrVal(yyS[yypt-0].bytes)}
		}
	case 245:
		{
			yyVAL.statement = &Show{Section: "create schema", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 246:
		{
			yyVAL.statement = &Show{Section: "create event", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 247:
		{
			yyVAL.statement = &Show{Section: "create function", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 248:
		{
			yyVAL.statement = &Show{Section: "create procedure", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 249:
		{
			yyVAL.statement = &Show{Section: "create table", From: &ColName{option.NoneString(), option.SomeString(string(yyS[yypt-2].bytes)), string(yyS[yypt-0].bytes)}}
		}
	case 250:
		{
			yyVAL.statement = &Show{Section: "create table", From: StrVal(yyS[yypt-0].bytes)}
		}
	case 251:
		{
			yyVAL.statement = &Show{Section: "create table", From: StrVal(yyS[yypt-0].bytes)}
		}
	case 252:
		{
			yyVAL.statement = &Show{Section: "create trigger", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 253:
		{
			yyVAL.statement = &Show{Section: "create user", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 254:
		{
			yyVAL.statement = &Show{Section: "create view", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 255:
		{
			yyVAL.statement = &Show{Section: "engine", Modifier: string(yyS[yypt-1].bytes)}
		}
	case 256:
		{
			yyVAL.statement = &Show{Section: "engine", Modifier: string(yyS[yypt-1].bytes)}
		}
	case 257:
		{
			yyVAL.statement = &Show{Section: "engines"}
		}
	case 258:
		{
			yyVAL.statement = &Show{Section: "errors"}
		}
	case 259:
		{
			yyVAL.statement = &Show{Section: "count(*) errors"}
		}
	case 260:
		{
			yyVAL.statement = &Show{Section: "events"}
		}
	case 261:
		{
			yyVAL.statement = &Show{Section: "function code", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 262:
		{
			yyVAL.statement = &Show{Section: "function status"}
		}
	case 263:
		{
			yyVAL.statement = &Show{Section: "grants", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 264:
		{
			yyVAL.statement = &Show{Section: "indexes", From: yyS[yypt-1].expr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 265:
		{
			yyVAL.statement = &Show{Section: "indexes", From: yyS[yypt-1].expr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 266:
		{
			yyVAL.statement = &Show{Section: "keys", From: yyS[yypt-1].expr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 267:
		{
			yyVAL.statement = &Show{Section: "master status"}
		}
	case 268:
		{
			yyVAL.statement = &Show{Section: "open tables"}
		}
	case 269:
		{
			yyVAL.statement = &Show{Section: "plugins"}
		}
	case 270:
		{
			yyVAL.statement = &Show{Section: "privileges"}
		}
	case 271:
		{
			yyVAL.statement = &Show{Section: "procedure code", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 272:
		{
			yyVAL.statement = &Show{Section: "procedure status"}
		}
	case 273:
		{
			yyVAL.statement = &Show{Section: "profile"}
		}
	case 274:
		{
			yyVAL.statement = &Show{Section: "profiles"}
		}
	case 275:
		{
			yyVAL.statement = &Show{Section: "relaylog events"}
		}
	case 276:
		{
			yyVAL.statement = &Show{Section: "slave hosts"}
		}
	case 277:
		{
			yyVAL.statement = &Show{Section: "slave status", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 278:
		{
			yyVAL.statement = &Show{Section: "table status"}
		}
	case 279:
		{
			yyVAL.statement = &Show{Section: "table status"}
		}
	case 280:
		{
			yyVAL.statement = &Show{Section: "warnings"}
		}
	case 281:
		{
			yyVAL.statement = &Show{Section: "count(*) errors"}
		}
	case 282:
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 283:
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 284:
		{
			yyVAL.statement = &Show{Section: "schemas", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 285:
		{
			yyVAL.statement = &Show{Section: "variables", Modifier: yyS[yypt-2].str, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 286:
		{
			yyVAL.statement = &Show{Section: "tables", Modifier: yyS[yypt-3].str, From: yyS[yypt-1].expr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 287:
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyS[yypt-2].bytes), From: yyS[yypt-1].expr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 288:
		{
			yyVAL.statement = &Show{Section: "columns", From: yyS[yypt-1].expr, Modifier: yyS[yypt-3].str, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 289:
		{
			yyVAL.statement = &Show{Section: "processlist", Modifier: yyS[yypt-1].str}
		}
	case 290:
		{
			yyVAL.statement = &Show{Section: "status", Modifier: yyS[yypt-2].str, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 291:
		{
			yyVAL.statement = &Show{Section: "charset", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 292:
		{
			yyVAL.statement = &Show{Section: "charset", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 293:
		{
			yyVAL.statement = &Show{Section: "collation", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 294:
		{
			yyVAL.str = AST_EXPLAIN_FORMAT_TRADITIONAL
		}
	case 295:
		{
			yyVAL.str = AST_EXPLAIN_FORMAT_JSON
		}
	case 296:
		{
			yyVAL.str = AST_EXPLAIN_EXTENDED
		}
	case 297:
		{
			yyVAL.str = AST_EXPLAIN_PARTITIONS
		}
	case 298:
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 299:
		{
			yyVAL.str = ""
		}
	case 300:
		{
			yyVAL.statement = yyS[yypt-0].selStmt
		}
	case 301:
		{
			yyVAL.statement = yyS[yypt-1].statement
		}
	case 302:
		{
			yyVAL.empty = yyS[yypt-0].empty
		}
	case 303:
		{
			yyVAL.empty = yyS[yypt-0].empty
		}
	case 304:
		{
			yyVAL.empty = yyS[yypt-0].empty
		}
	case 305:
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 306:
		{
			yyVAL.tableName = &TableName{Name: yyS[yypt-0].str}
		}
	case 307:
		{
			yyVAL.tableName = &TableName{Name: yyS[yypt-0].str}
		}
	case 308:
		{
			yyVAL.tableName = &TableName{Qualifier: option.SomeString(yyS[yypt-2].str), Name: yyS[yypt-0].str}
		}
	case 309:
		{
			yyVAL.tableExprs = TableExprs{&DualTableExpr{}}
		}
	case 310:
		{
			yyVAL.statement = &Explain{Section: "table", Table: yyS[yypt-1].tableName, Column: yyS[yypt-0].colName}
		}
	case 311:
		{
			yyVAL.statement = &Explain{Section: "plan", ExplainType: yyS[yypt-1].str, Statement: yyS[yypt-0].statement}
		}
	case 312:
		{
			yyVAL.statement = &Explain{Section: "plan", ExplainType: yyS[yypt-3].str, Connection: option.SomeString(string(yyS[yypt-0].bytes))}
		}
	case 313:
		{
			yyVAL.colName = nil
		}
	case 314:
		{
			yyVAL.colName = &ColName{Name: yyS[yypt-0].str}
		}
	case 315:
		{
			yyVAL.colName = &ColName{Name: string(yyS[yypt-0].bytes)}
		}
	case 316:
		{
			yyVAL.statement = &Kill{Scope: AST_KILL_CONNECTION, ID: yyS[yypt-0].expr}
		}
	case 317:
		{
			yyVAL.statement = &Kill{Scope: yyS[yypt-1].str, ID: yyS[yypt-0].expr}
		}
	case 318:
		{
			SetAllowComments(yylex, true)
		}
	case 319:
		{
			yyVAL.strs = yyS[yypt-0].strs
			SetAllowComments(yylex, false)
		}
	case 320:
		{
			yyVAL.strs = nil
		}
	case 321:
		{
			yyVAL.strs = append(yyS[yypt-1].strs, string(yyS[yypt-0].bytes))
		}
	case 322:
		{
			yyVAL.queryGlobals = &QueryGlobals{Distinct: false, StraightJoin: false}
		}
	case 323:
		{
			yyVAL.queryGlobals = &QueryGlobals{Distinct: true, StraightJoin: false}
		}
	case 324:
		{
			yyVAL.queryGlobals = &QueryGlobals{Distinct: false, StraightJoin: true}
		}
	case 325:
		{
			yyVAL.queryGlobals = &QueryGlobals{Distinct: true, StraightJoin: true}
		}
	case 326:
		{
			yyVAL.queryGlobals = &QueryGlobals{Distinct: true, StraightJoin: true}
		}
	case 327:
		{
			yyVAL.str = AST_UNION
		}
	case 328:
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 329:
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 330:
		{
			yyVAL.str = AST_EXCEPT
		}
	case 331:
		{
			yyVAL.str = AST_INTERSECT
		}
	case 332:
		{
			yyVAL.bool = false
		}
	case 333:
		{
			yyVAL.bool = true
		}
	case 334:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 335:
		{
			yyVAL.stropt = option.SomeString(string(yyS[yypt-0].bytes))
		}
	case 336:
		{
			yyVAL.selectExprs = SelectExprs{yyS[yypt-0].selectExpr}
		}
	case 337:
		{
			yyVAL.selectExprs = append(yyS[yypt-2].selectExprs, yyS[yypt-0].selectExpr)
		}
	case 338:
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 339:
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyS[yypt-1].expr, As: yyS[yypt-0].stropt}
		}
	case 340:
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyS[yypt-2].expr, As: yyS[yypt-1].stropt}
		}
	case 341:
		{
			yyVAL.selectExpr = &StarExpr{TableName: option.SomeString(yyS[yypt-2].str)}
		}
	case 342:
		{
			yyVAL.selectExpr = &StarExpr{DatabaseName: option.SomeString(yyS[yypt-4].str), TableName: option.SomeString(yyS[yypt-2].str)}
		}
	case 343:
		{
			yyVAL.columnExprs = ColumnExprs{&ColName{Name: string(yyS[yypt-0].bytes)}}
		}
	case 344:
		{
			yyVAL.columnExprs = append(yyS[yypt-2].columnExprs, &ColName{Name: string(yyS[yypt-0].bytes)})
		}
	case 345:
		{
			yyVAL.tableExprs = TableExprs{yyS[yypt-0].tableExpr}
		}
	case 346:
		{
			yyVAL.tableExprs = append(yyS[yypt-2].tableExprs, yyS[yypt-0].tableExpr)
		}
	case 347:
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyS[yypt-2].smTableExpr, As: yyS[yypt-1].stropt, Hints: yyS[yypt-0].indexHints}
		}
	case 348:
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyS[yypt-1].tableExpr}
		}
	case 349:
		{
			yyVAL.tableExpr = yyS[yypt-0].tableExpr
		}
	case 350:
		{
			yyVAL.tableExpr = yyS[yypt-1].tableExpr
		}
	case 351:
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-2].tableExpr, Join: yyS[yypt-1].str, RightExpr: yyS[yypt-0].tableExpr}
		}
	case 352:
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-2].tableExpr, Join: AST_STRAIGHT_JOIN, RightExpr: yyS[yypt-0].tableExpr}
		}
	case 353:
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-4].tableExpr, Join: yyS[yypt-3].str, RightExpr: yyS[yypt-2].tableExpr, On: yyS[yypt-0].expr}
		}
	case 354:
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-4].tableExpr, Join: AST_STRAIGHT_JOIN, RightExpr: yyS[yypt-2].tableExpr, On: yyS[yypt-0].expr}
		}
	case 355:
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-6].tableExpr, Join: yyS[yypt-5].str, RightExpr: yyS[yypt-4].tableExpr, Using: yyS[yypt-1].columnExprs}
		}
	case 356:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 357:
		{
			yyVAL.stropt = option.SomeString(yyS[yypt-0].str)
		}
	case 358:
		{
			yyVAL.stropt = option.SomeString(yyS[yypt-0].str)
		}
	case 359:
		{
			yyVAL.stropt = option.SomeString(string(yyS[yypt-0].bytes))
		}
	case 360:
		{
			yyVAL.stropt = option.SomeString(string(yyS[yypt-0].bytes))
		}
	case 361:
		{
			yyVAL.str = AST_JOIN
		}
	case 362:
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 363:
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 364:
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 365:
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 366:
		{
			yyVAL.str = AST_JOIN
		}
	case 367:
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 368:
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 369:
		{
			yyVAL.str = AST_NATURAL_RIGHT_JOIN
		}
	case 370:
		{
			yyVAL.str = AST_NATURAL_RIGHT_JOIN
		}
	case 371:
		{
			yyVAL.str = AST_NATURAL_LEFT_JOIN
		}
	case 372:
		{
			yyVAL.str = AST_NATURAL_LEFT_JOIN
		}
	case 373:
		{
			yyVAL.smTableExpr = yyS[yypt-0].tableName
		}
	case 374:
		{
			yyVAL.smTableExpr = yyS[yypt-0].subquery
		}
	case 375:
		{
			yyVAL.indexHints = nil
		}
	case 376:
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyS[yypt-1].strs}
		}
	case 377:
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyS[yypt-1].strs}
		}
	case 378:
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyS[yypt-1].strs}
		}
	case 379:
		{
			yyVAL.strs = []string{string(yyS[yypt-0].bytes)}
		}
	case 380:
		{
			yyVAL.strs = append(yyS[yypt-2].strs, string(yyS[yypt-0].bytes))
		}
	case 381:
		{
			yyVAL.expr = nil
		}
	case 382:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 383:
		{
			yyVAL.expr = nil
		}
	case 384:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 385:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 386:
		{
			yyVAL.expr = nil
		}
	case 387:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 388:
		{
			yyVAL.str = string("")
		}
	case 389:
		{
			yyVAL.str = "IF NOT EXISTS"
		}
	case 390:
		{
			yyVAL.expr = nil
		}
	case 391:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 392:
		{
			yyVAL.bytes = nil
		}
	case 393:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 394:
		{
			yyVAL.bytes = nil
		}
	case 395:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 396:
		{
			yyVAL.empty = struct{}{}
		}
	case 397:
		{
			yyVAL.empty = struct{}{}
		}
	case 398:
		{
			yyVAL.str = AST_ALL
		}
	case 399:
		{
			yyVAL.str = AST_SOME
		}
	case 400:
		{
			yyVAL.str = AST_ANY
		}
	case 401:
		{
			yyVAL.tuple = ValTuple(append(yyS[yypt-3].exprs, yyS[yypt-1].expr))
		}
	case 402:
		{
			yyVAL.tuple = ValTuple(yyS[yypt-1].exprs)
		}
	case 403:
		{
			yyVAL.subquery = &Subquery{Select: yyS[yypt-1].selStmt, IsDerived: true}
		}
	case 404:
		{
			yyVAL.exprs = Exprs{yyS[yypt-0].expr}
		}
	case 405:
		{
			yyVAL.exprs = append(yyS[yypt-2].exprs, yyS[yypt-0].expr)
		}
	case 406:
		{
			yyVAL.expr = &OrExpr{Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 407:
		{
			yyVAL.expr = &XorExpr{Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 408:
		{
			yyVAL.expr = &AndExpr{Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 409:
		{
			yyVAL.expr = &NotExpr{Expr: yyS[yypt-0].expr}
		}
	case 410:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_IS, Right: yyS[yypt-0].val}
		}
	case 411:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_IS_NOT, Right: yyS[yypt-0].val}
		}
	case 413:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 414:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 415:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_EQ, Right: yyS[yypt-0].expr}
		}
	case 416:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_EQ, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 417:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_NE, Right: yyS[yypt-0].expr}
		}
	case 418:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_NE, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 419:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_NSE, Right: yyS[yypt-0].expr}
		}
	case 420:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_NSE, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 421:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_LT, Right: yyS[yypt-0].expr}
		}
	case 422:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_LT, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 423:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_GT, Right: yyS[yypt-0].expr}
		}
	case 424:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_GT, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 425:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_LE, Right: yyS[yypt-0].expr}
		}
	case 426:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_LE, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 427:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_GE, Right: yyS[yypt-0].expr}
		}
	case 428:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_GE, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 430:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_IN, SubqueryOperator: AST_IN, Right: yyS[yypt-0].subquery}
		}
	case 431:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_IN, Right: yyS[yypt-0].tuple}
		}
	case 432:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_NOT_IN, SubqueryOperator: AST_NOT_IN, Right: yyS[yypt-0].subquery}
		}
	case 433:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_NOT_IN, Right: yyS[yypt-0].tuple}
		}
	case 434:
		{
			yyVAL.expr = &RangeCond{Left: yyS[yypt-4].expr, Operator: AST_BETWEEN, From: yyS[yypt-2].expr, To: yyS[yypt-0].expr}
		}
	case 435:
		{
			yyVAL.expr = &RangeCond{Left: yyS[yypt-5].expr, Operator: AST_NOT_BETWEEN, From: yyS[yypt-2].expr, To: yyS[yypt-0].expr}
		}
	case 436:
		{
			yyVAL.expr = &LikeExpr{Left: yyS[yypt-3].expr, Operator: AST_LIKE, Right: yyS[yypt-1].expr, Escape: yyS[yypt-0].expr}
		}
	case 437:
		{
			yyVAL.expr = &LikeExpr{Left: yyS[yypt-4].expr, Operator: AST_NOT_LIKE, Right: yyS[yypt-1].expr, Escape: yyS[yypt-0].expr}
		}
	case 438:
		{
			yyVAL.expr = &LikeExpr{Left: yyS[yypt-4].expr, Operator: AST_LIKE_BINARY, Right: yyS[yypt-1].expr, Escape: yyS[yypt-0].expr}
		}
	case 439:
		{
			yyVAL.expr = &LikeExpr{Left: yyS[yypt-5].expr, Operator: AST_NOT_LIKE_BINARY, Right: yyS[yypt-1].expr, Escape: yyS[yypt-0].expr}
		}
	case 440:
		{
			yyVAL.expr = &RegexExpr{Operand: yyS[yypt-2].expr, Pattern: yyS[yypt-0].expr}
		}
	case 441:
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyS[yypt-3].expr, Pattern: yyS[yypt-0].expr}}
		}
	case 442:
		{
			yyVAL.expr = &RLikeExpr{Operand: yyS[yypt-2].expr, Pattern: yyS[yypt-0].expr}
		}
	case 443:
		{
			yyVAL.expr = &NotExpr{&RLikeExpr{Operand: yyS[yypt-3].expr, Pattern: yyS[yypt-0].expr}}
		}
	case 445:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_BITAND, Right: yyS[yypt-0].expr}
		}
	case 446:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_BITOR, Right: yyS[yypt-0].expr}
		}
	case 447:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_PLUS, Right: yyS[yypt-0].expr}
		}
	case 448:
		{
			yyVAL.expr = &FuncExpr{
				Name: string(DATE_ADD_BYTES),
				Exprs: append(SelectExprs{
					&NonStarExpr{Expr: yyS[yypt-4].expr},
					&NonStarExpr{Expr: yyS[yypt-1].expr},
					&NonStarExpr{Expr: KeywordVal(yyS[yypt-0].bytes)},
				}),
			}
		}
	case 449:
		{
			yyVAL.expr = &FuncExpr{
				Name: string(DATE_ADD_BYTES),
				Exprs: append(SelectExprs{
					&NonStarExpr{Expr: yyS[yypt-0].expr},
					&NonStarExpr{Expr: yyS[yypt-3].expr},
					&NonStarExpr{Expr: KeywordVal(yyS[yypt-2].bytes)},
				}),
			}
		}
	case 450:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_MINUS, Right: yyS[yypt-0].expr}
		}
	case 451:
		{
			yyVAL.expr = &FuncExpr{
				Name: string(SUBDATE_BYTES),
				Exprs: append(SelectExprs{
					&NonStarExpr{Expr: yyS[yypt-4].expr},
					&NonStarExpr{Expr: yyS[yypt-1].expr},
					&NonStarExpr{Expr: KeywordVal(yyS[yypt-0].bytes)},
				}),
			}
		}
	case 452:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_MULT, Right: yyS[yypt-0].expr}
		}
	case 453:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_DIV, Right: yyS[yypt-0].expr}
		}
	case 454:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_IDIV, Right: yyS[yypt-0].expr}
		}
	case 455:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_MOD, Right: yyS[yypt-0].expr}
		}
	case 456:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_BITXOR, Right: yyS[yypt-0].expr}
		}
	case 458:
		{
			yyVAL.expr = yyS[yypt-0].val
		}
	case 459:
		{
			yyVAL.expr = yyS[yypt-0].colName
		}
	case 460:
		{
			yyVAL.expr = yyS[yypt-0].tuple
		}
	case 461:
		{
			yyVAL.expr = yyS[yypt-0].subquery
		}
	case 462:
		{
			if num, ok := yyS[yypt-0].expr.(NumVal); ok {
				switch yyS[yypt-1].byt {
				case '-':
					yyVAL.expr = "-" + num
				case '+':
					yyVAL.expr = num
				default:
					yyVAL.expr = &UnaryExpr{Operator: yyS[yypt-1].byt, Expr: yyS[yypt-0].expr}
				}
			} else {
				yyVAL.expr = &UnaryExpr{Operator: yyS[yypt-1].byt, Expr: yyS[yypt-0].expr}
			}
		}
	case 463:
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyS[yypt-0].subquery}
		}
	case 464:
		{
			yyVAL.expr = yyS[yypt-0].caseExpr
		}
	case 465:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 466:
		{
			yyVAL.expr = &FuncExpr{Name: string(VALUES_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 467:
		{
			yyVAL.expr = yyS[yypt-1].expr
		}
	case 468:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 469:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 470:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 471:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 472:
		{
			yyVAL.expr = &FuncExpr{Name: string(CHAR_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 473:
		{
			yyVAL.expr = &FuncExpr{Name: string(CONVERT_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyS[yypt-3].expr}, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 474:
		{
			yyVAL.expr = &FuncExpr{Name: string(INSERT_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 475:
		{
			yyVAL.expr = &FuncExpr{Name: string(LEFT_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 476:
		{
			yyVAL.expr = &FuncExpr{Name: string(RIGHT_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 477:
		{
			yyVAL.expr = &FuncExpr{Name: string(ADDDATE_BYTES), Exprs: append(SelectExprs{yyS[yypt-5].selectExpr, yyS[yypt-2].selectExpr, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 478:
		{
			yyVAL.expr = &FuncExpr{Name: string(ADDDATE_BYTES), Exprs: append(SelectExprs{yyS[yypt-3].selectExpr, yyS[yypt-1].selectExpr, &NonStarExpr{Expr: KeywordVal(DAY_BYTES)}})}
		}
	case 479:
		{
			yyVAL.expr = &FuncExpr{Name: string(CONVERT_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyS[yypt-3].expr}, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 480:
		{
			yyVAL.expr = &FuncExpr{Name: string(CONVERT_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyS[yypt-4].expr}, &NonStarExpr{Expr: KeywordVal(yyS[yypt-2].bytes)}})}
		}
	case 481:
		{
			yyVAL.expr = &FuncExpr{Name: string(CURRENT_DATE_BYTES)}
		}
	case 482:
		{
			yyVAL.expr = &FuncExpr{Name: string(CURRENT_TIMESTAMP_BYTES)}
		}
	case 483:
		{
			yyVAL.expr = &FuncExpr{Name: string(CURRENT_TIMESTAMP_BYTES)}
		}
	case 484:
		{
			yyVAL.expr = &FuncExpr{Name: string(DATE_ADD_BYTES), Exprs: append(SelectExprs{yyS[yypt-5].selectExpr, yyS[yypt-2].selectExpr, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 485:
		{
			yyVAL.expr = &FuncExpr{Name: string(DATE_SUB_BYTES), Exprs: append(SelectExprs{yyS[yypt-5].selectExpr, yyS[yypt-2].selectExpr, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 486:
		{
			yyVAL.expr = &FuncExpr{Name: string(EXTRACT_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyS[yypt-3].bytes)}}, yyS[yypt-1].selectExpr)}
		}
	case 487:
		{
			yyVAL.expr = &FuncExpr{Name: string(GROUP_CONCAT_BYTES), Distinct: yyS[yypt-4].bool, Exprs: yyS[yypt-3].selectExprs, OrderBy: yyS[yypt-2].orderBy, Separator: yyS[yypt-1].stropt}
		}
	case 488:
		{
			yyVAL.expr = &FuncExpr{Name: string(SUBDATE_BYTES), Exprs: append(SelectExprs{yyS[yypt-3].selectExpr, yyS[yypt-1].selectExpr, &NonStarExpr{Expr: KeywordVal(DAY_BYTES)}})}
		}
	case 489:
		{
			yyVAL.expr = &FuncExpr{Name: string(SUBDATE_BYTES), Exprs: append(SelectExprs{yyS[yypt-5].selectExpr, yyS[yypt-2].selectExpr, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 490:
		{
			yyVAL.expr = &FuncExpr{Name: string(TIMESTAMPADD_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyS[yypt-3].bytes)}}, yyS[yypt-1].selectExprs...)}
		}
	case 491:
		{
			yyVAL.expr = &FuncExpr{Name: string(TIMESTAMPADD_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyS[yypt-3].bytes)}}, yyS[yypt-1].selectExprs...)}
		}
	case 492:
		{
			yyVAL.expr = &FuncExpr{Name: string(TIMESTAMPDIFF_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyS[yypt-3].bytes)}}, yyS[yypt-1].selectExprs...)}
		}
	case 493:
		{
			yyVAL.expr = &FuncExpr{Name: string(TIMESTAMPDIFF_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyS[yypt-3].bytes)}}, yyS[yypt-1].selectExprs...)}
		}
	case 494:
		{
			yyVAL.expr = &FuncExpr{Name: string(TRIM_BYTES), Exprs: []SelectExpr{yyS[yypt-1].selectExpr}}
		}
	case 495:
		{
			yyVAL.expr = &FuncExpr{Name: string(TRIM_BYTES), Exprs: []SelectExpr{yyS[yypt-1].selectExpr, &NonStarExpr{Expr: StrVal(yyS[yypt-4].bytes)}, yyS[yypt-3].selectExpr}}
		}
	case 496:
		{
			yyVAL.expr = &FuncExpr{Name: string(TRIM_BYTES), Exprs: []SelectExpr{yyS[yypt-1].selectExpr, &NonStarExpr{Expr: StrVal(BOTH_BYTES)}, yyS[yypt-3].selectExpr}}
		}
	case 497:
		{
			yyVAL.expr = &FuncExpr{Name: string(yyS[yypt-7].bytes), Exprs: []SelectExpr{yyS[yypt-5].selectExpr, yyS[yypt-3].selectExpr, yyS[yypt-1].selectExpr}}
		}
	case 498:
		{
			yyVAL.expr = &FuncExpr{Name: string(yyS[yypt-5].bytes), Exprs: []SelectExpr{yyS[yypt-3].selectExpr, yyS[yypt-1].selectExpr}}
		}
	case 499:
		{
			yyVAL.expr = &FuncExpr{Name: string(yyS[yypt-3].bytes), Exprs: yyS[yypt-1].selectExprs}
		}
	case 500:
		{
			yyVAL.expr = &FuncExpr{Name: string(UTC_TIMESTAMP_BYTES)}
		}
	case 501:
		{
			yyVAL.expr = &FuncExpr{Name: string(UTC_DATE_BYTES)}
		}
	case 502:
		{
			yyVAL.expr = &FuncExpr{Name: string(COUNT_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 503:
		{
			yyVAL.expr = &FuncExpr{Name: string(COUNT_BYTES), Distinct: true, Exprs: yyS[yypt-1].selectExprs}
		}
	case 504:
		{
			yyVAL.expr = &FuncExpr{Name: string(DATABASE_BYTES)}
		}
	case 505:
		{
			yyVAL.expr = &FuncExpr{Name: string(DATE_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 506:
		{
			yyVAL.expr = &FuncExpr{Name: string(DAY_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 507:
		{
			yyVAL.expr = &FuncExpr{Name: string(HOUR_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 508:
		{
			yyVAL.expr = &FuncExpr{Name: string(IF_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 509:
		{
			yyVAL.expr = &FuncExpr{Name: string(INTERVAL_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 510:
		{
			yyVAL.expr = &FuncExpr{Name: string(MICROSECOND_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 511:
		{
			yyVAL.expr = &FuncExpr{Name: string(MINUTE_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 512:
		{
			yyVAL.expr = &FuncExpr{Name: string(MOD_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 513:
		{
			yyVAL.expr = &FuncExpr{Name: string(MONTH_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 514:
		{
			yyVAL.expr = &FuncExpr{Name: string(QUARTER_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 515:
		{
			yyVAL.expr = &FuncExpr{Name: string(SCHEMA_BYTES)}
		}
	case 516:
		{
			yyVAL.expr = &FuncExpr{Name: string(SECOND_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 517:
		{
			yyVAL.expr = &FuncExpr{Name: string(TIMESTAMP_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 518:
		{
			yyVAL.expr = &FuncExpr{Name: string(WEEK_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 519:
		{
			yyVAL.expr = &FuncExpr{Name: string(USER_BYTES)}
		}
	case 520:
		{
			yyVAL.expr = &FuncExpr{Name: string(YEAR_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 521:
		{
			yyVAL.expr = &FuncExpr{Name: string(bytes.ToLower(yyS[yypt-2].bytes))}
		}
	case 522:
		{
			yyVAL.expr = &FuncExpr{Name: string(bytes.ToLower(yyS[yypt-3].bytes)), Exprs: yyS[yypt-1].selectExprs}
		}
	case 523:
		{
			yyVAL.expr = &FuncExpr{Name: string(bytes.ToLower(yyS[yypt-4].bytes)), Distinct: true, Exprs: yyS[yypt-1].selectExprs}
		}
	case 526:
		{
			yyVAL.expr = StrVal("\\")
		}
	case 527:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 528:
		{
			yyVAL.expr = yyS[yypt-1].expr
		}
	case 529:
		{
			yyVAL.bytes = BOTH_BYTES
		}
	case 530:
		{
			yyVAL.bytes = LEADING_BYTES
		}
	case 531:
		{
			yyVAL.bytes = TRAILING_BYTES
		}
	case 532:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 533:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 534:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 535:
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 536:
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 537:
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 538:
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 539:
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 540:
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 541:
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 542:
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 543:
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 544:
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 545:
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 546:
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 547:
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 548:
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 549:
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 550:
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 551:
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 552:
		{
			yyVAL.bytes = SECOND_MICROSECOND_BYTES
		}
	case 553:
		{
			yyVAL.bytes = MINUTE_MICROSECOND_BYTES
		}
	case 554:
		{
			yyVAL.bytes = MINUTE_SECOND_BYTES
		}
	case 555:
		{
			yyVAL.bytes = HOUR_MICROSECOND_BYTES
		}
	case 556:
		{
			yyVAL.bytes = HOUR_SECOND_BYTES
		}
	case 557:
		{
			yyVAL.bytes = HOUR_MINUTE_BYTES
		}
	case 558:
		{
			yyVAL.bytes = DAY_MICROSECOND_BYTES
		}
	case 559:
		{
			yyVAL.bytes = DAY_SECOND_BYTES
		}
	case 560:
		{
			yyVAL.bytes = DAY_MINUTE_BYTES
		}
	case 561:
		{
			yyVAL.bytes = DAY_HOUR_BYTES
		}
	case 562:
		{
			yyVAL.bytes = YEAR_MONTH_BYTES
		}
	case 563:
		{
			yyVAL.bytes = BINARY_BYTES
		}
	case 564:
		{
			yyVAL.bytes = BINARY_BYTES
		}
	case 565:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 566:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 567:
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 568:
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 569:
		{
			yyVAL.bytes = DECIMAL_BYTES
		}
	case 570:
		{
			yyVAL.bytes = DECIMAL_BYTES
		}
	case 571:
		{
			yyVAL.bytes = DECIMAL_BYTES
		}
	case 572:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 573:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 574:
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 575:
		{
			yyVAL.bytes = SIGNED_BYTES
		}
	case 576:
		{
			yyVAL.bytes = OBJECT_ID_BYTES
		}
	case 577:
		{
			yyVAL.bytes = SIGNED_BYTES
		}
	case 578:
		{
			yyVAL.bytes = SIGNED_BYTES
		}
	case 579:
		{
			yyVAL.bytes = TIME_BYTES
		}
	case 580:
		{
			yyVAL.bytes = UNSIGNED_BYTES
		}
	case 581:
		{
			yyVAL.bytes = UNSIGNED_BYTES
		}
	case 582:
		{
			yyVAL.bytes = SIGNED_BYTES
		}
	case 583:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 584:
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 585:
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 586:
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 587:
		{
			yyVAL.byt = AST_UPLUS
		}
	case 588:
		{
			yyVAL.byt = AST_UMINUS
		}
	case 589:
		{
			yyVAL.byt = AST_TILDA
		}
	case 590:
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyS[yypt-3].expr, Whens: yyS[yypt-2].whens, Else: yyS[yypt-1].expr}
		}
	case 591:
		{
			yyVAL.expr = nil
		}
	case 592:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 593:
		{
			yyVAL.whens = []*When{yyS[yypt-0].when}
		}
	case 594:
		{
			yyVAL.whens = append(yyS[yypt-1].whens, yyS[yypt-0].when)
		}
	case 595:
		{
			yyVAL.when = &When{Cond: yyS[yypt-2].expr, Val: yyS[yypt-0].expr}
		}
	case 596:
		{
			yyVAL.expr = nil
		}
	case 597:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 598:
		{
			yyVAL.colName = &ColName{Name: yyS[yypt-0].str}
		}
	case 599:
		{
			yyVAL.colName = &ColName{Qualifier: option.SomeString(yyS[yypt-2].str), Name: yyS[yypt-0].str}
		}
	case 600:
		{
			yyVAL.colName = &ColName{Database: option.SomeString(yyS[yypt-4].str), Qualifier: option.SomeString(yyS[yypt-2].str), Name: yyS[yypt-0].str}
		}
	case 601:
		{
			yyVAL.val = StrVal(yyS[yypt-0].bytes)
		}
	case 602:
		{
			yyVAL.val = NumVal(yyS[yypt-0].bytes)
		}
	case 603:
		{
			yyVAL.val = ValArg(yyS[yypt-0].bytes)
		}
	case 604:
		{
			yyVAL.val = &DateVal{Name: AST_DATE, Val: string(yyS[yypt-0].bytes)}
		}
	case 605:
		{
			yyVAL.val = &DateVal{Name: AST_TIME, Val: string(yyS[yypt-0].bytes)}
		}
	case 606:
		{
			yyVAL.val = &DateVal{Name: AST_TIMESTAMP, Val: string(yyS[yypt-0].bytes)}
		}
	case 607:
		{
			if bytes.Equal(bytes.ToLower(yyS[yypt-2].bytes), D_BYTES) {
				yyVAL.val = &DateVal{Name: AST_DATE, Val: string(yyS[yypt-1].bytes)}
			} else if bytes.Equal(bytes.ToLower(yyS[yypt-2].bytes), T_BYTES) {
				yyVAL.val = &DateVal{Name: AST_TIME, Val: string(yyS[yypt-1].bytes)}
			} else if bytes.Equal(bytes.ToLower(yyS[yypt-2].bytes), TS_BYTES) {
				yyVAL.val = &DateVal{Name: AST_TIMESTAMP, Val: string(yyS[yypt-1].bytes)}
			} else {
				yylex.Error("expecting d, t, or ts")
				return 1
			}
		}
	case 608:
		{
			yyVAL.val = &NullVal{}
		}
	case 609:
		{
			yyVAL.val = yyS[yypt-0].val
		}
	case 610:
		{
			yyVAL.val = &TrueVal{}
		}
	case 611:
		{
			yyVAL.val = &FalseVal{}
		}
	case 612:
		{
			yyVAL.val = &UnknownVal{}
		}
	case 613:
		{
			yyVAL.exprs = nil
		}
	case 614:
		{
			yyVAL.exprs = yyS[yypt-0].exprs
		}
	case 615:
		{
			yyVAL.expr = nil
		}
	case 616:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 617:
		{
			yyVAL.orderBy = nil
		}
	case 618:
		{
			yyVAL.orderBy = yyS[yypt-0].orderBy
		}
	case 619:
		{
			yyVAL.orderBy = OrderBy{yyS[yypt-0].order}
		}
	case 620:
		{
			yyVAL.orderBy = append(yyS[yypt-2].orderBy, yyS[yypt-0].order)
		}
	case 621:
		{
			yyVAL.order = &Order{Expr: yyS[yypt-1].expr, Direction: yyS[yypt-0].str}
		}
	case 622:
		{
			yyVAL.str = AST_ASC
		}
	case 623:
		{
			yyVAL.str = AST_ASC
		}
	case 624:
		{
			yyVAL.str = AST_DESC
		}
	case 625:
		{
			yyVAL.limit = nil
		}
	case 626:
		{
			yyVAL.limit = &Limit{Rowcount: yyS[yypt-0].expr}
		}
	case 627:
		{
			yyVAL.limit = &Limit{Offset: yyS[yypt-2].expr, Rowcount: yyS[yypt-0].expr}
		}
	case 628:
		{
			yyVAL.limit = &Limit{Offset: yyS[yypt-0].expr, Rowcount: yyS[yypt-2].expr}
		}
	case 629:
		{
			yyVAL.str = ""
		}
	case 630:
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 631:
		{
			if !bytes.Equal(yyS[yypt-1].bytes, SHARE_BYTES) {
				yylex.Error("expecting share")
				return 1
			}
			if !bytes.Equal(yyS[yypt-0].bytes, MODE_BYTES) {
				yylex.Error("expecting mode")
				return 1
			}
			yyVAL.str = AST_SHARE_MODE
		}
	case 632:
		{
			yyVAL.str = string(yyS[yypt-0].bytes)
		}
	case 633:
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 634:
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 635:
		{
			yyVAL.str = string(yyS[yypt-0].bytes)
		}
	case 636:
		{
			yyVAL.str = string(ANY_BYTES)
		}
	case 637:
		{
			yyVAL.str = string(BINLOG_BYTES)
		}
	case 638:
		{
			yyVAL.str = string(AUTO_INCREMENT_BYTES)
		}
	case 639:
		{
			yyVAL.str = string(BIT_BYTES)
		}
	case 640:
		{
			yyVAL.str = string(BLOB_BYTES)
		}
	case 641:
		{
			yyVAL.str = string(BOOL_BYTES)
		}
	case 642:
		{
			yyVAL.str = string(BTREE_BYTES)
		}
	case 643:
		{
			yyVAL.str = string(CHANNEL_BYTES)
		}
	case 644:
		{
			yyVAL.str = string(CHARSET_BYTES)
		}
	case 645:
		{
			yyVAL.str = string(CODE_BYTES)
		}
	case 646:
		{
			yyVAL.str = string(COLLATION_BYTES)
		}
	case 647:
		{
			yyVAL.str = string(COLUMNS_BYTES)
		}
	case 648:
		{
			yyVAL.str = string(COMMENT_BYTES)
		}
	case 649:
		{
			yyVAL.str = string(COMMITTED_BYTES)
		}
	case 650:
		{
			yyVAL.str = string(CONNECTION_BYTES)
		}
	case 651:
		{
			yyVAL.str = string(COUNT_BYTES)
		}
	case 652:
		{
			yyVAL.str = string(DATE_BYTES)
		}
	case 653:
		{
			yyVAL.str = string(DATETIME_BYTES)
		}
	case 654:
		{
			yyVAL.str = string(DAY_BYTES)
		}
	case 655:
		{
			yyVAL.str = string(DBS_BYTES)
		}
	case 656:
		{
			yyVAL.str = string(DECIMAL_BYTES)
		}
	case 657:
		{
			yyVAL.str = string(DISABLE_BYTES)
		}
	case 658:
		{
			yyVAL.str = string(ENABLE_BYTES)
		}
	case 659:
		{
			yyVAL.str = string(ENGINE_BYTES)
		}
	case 660:
		{
			yyVAL.str = string(ENGINES_BYTES)
		}
	case 661:
		{
			yyVAL.str = string(ENUM_BYTES)
		}
	case 662:
		{
			yyVAL.str = string(ERRORS_BYTES)
		}
	case 663:
		{
			yyVAL.str = string(EVENT_BYTES)
		}
	case 664:
		{
			yyVAL.str = string(EVENTS_BYTES)
		}
	case 665:
		{
			yyVAL.str = string(EXTENDED_BYTES)
		}
	case 666:
		{
			yyVAL.str = string(FLOAT_BYTES)
		}
	case 667:
		{
			yyVAL.str = string(FORMAT_BYTES)
		}
	case 668:
		{
			yyVAL.str = string(FULL_BYTES)
		}
	case 669:
		{
			yyVAL.str = string(FUNCTION_BYTES)
		}
	case 670:
		{
			yyVAL.str = string(GRANTS_BYTES)
		}
	case 671:
		{
			yyVAL.str = string(HASH)
		}
	case 672:
		{
			yyVAL.str = string(HOSTS_BYTES)
		}
	case 673:
		{
			yyVAL.str = string(HOUR_BYTES)
		}
	case 674:
		{
			yyVAL.str = string(INDEXES_BYTES)
		}
	case 675:
		{
			yyVAL.str = string(ISOLATION_BYTES)
		}
	case 676:
		{
			yyVAL.str = string(JSON_BYTES)
		}
	case 677:
		{
			yyVAL.str = string(LEVEL_BYTES)
		}
	case 678:
		{
			yyVAL.str = string(LOCAL_BYTES)
		}
	case 679:
		{
			yyVAL.str = string(LOGS_BYTES)
		}
	case 680:
		{
			yyVAL.str = string(LONGTEXT_BYTES)
		}
	case 681:
		{
			yyVAL.str = string(MASTER_BYTES)
		}
	case 682:
		{
			yyVAL.str = string(MEDIUMBLOB_BYTES)
		}
	case 683:
		{
			yyVAL.str = string(MEDIUMTEXT_BYTES)
		}
	case 684:
		{
			yyVAL.str = string(MICROSECOND_BYTES)
		}
	case 685:
		{
			yyVAL.str = string(MINUTE_BYTES)
		}
	case 686:
		{
			yyVAL.str = string(MONTH_BYTES)
		}
	case 687:
		{
			yyVAL.str = string(MUTEX_BYTES)
		}
	case 688:
		{
			yyVAL.str = string(NAMES_BYTES)
		}
	case 689:
		{
			yyVAL.str = string(NCHAR_BYTES)
		}
	case 690:
		{
			yyVAL.str = string(NUMBER_BYTES)
		}
	case 691:
		{
			yyVAL.str = string(OFFSET_BYTES)
		}
	case 692:
		{
			yyVAL.str = string(OBJECT_ID_BYTES)
		}
	case 693:
		{
			yyVAL.str = string(ONLY_BYTES)
		}
	case 694:
		{
			yyVAL.str = string(OPEN_BYTES)
		}
	case 695:
		{
			yyVAL.str = string(PARTITIONS_BYTES)
		}
	case 696:
		{
			yyVAL.str = string(PLUGINS_BYTES)
		}
	case 697:
		{
			yyVAL.str = string(PRIVILEGES_BYTES)
		}
	case 698:
		{
			yyVAL.str = string(PROCESSLIST_BYTES)
		}
	case 699:
		{
			yyVAL.str = string(PROFILE_BYTES)
		}
	case 700:
		{
			yyVAL.str = string(PROFILES_BYTES)
		}
	case 701:
		{
			yyVAL.str = string(PROXY_BYTES)
		}
	case 702:
		{
			yyVAL.str = string(QUARTER_BYTES)
		}
	case 703:
		{
			yyVAL.str = string(QUERY_BYTES)
		}
	case 704:
		{
			yyVAL.str = string(RELAYLOG_BYTES)
		}
	case 705:
		{
			yyVAL.str = string(REPEATABLE_BYTES)
		}
	case 706:
		{
			yyVAL.str = string(ROW_BYTES)
		}
	case 707:
		{
			yyVAL.str = string(SECOND_BYTES)
		}
	case 708:
		{
			yyVAL.str = string(SERIAL_BYTES)
		}
	case 709:
		{
			yyVAL.str = string(SERIALIZABLE_BYTES)
		}
	case 710:
		{
			yyVAL.str = string(SIGNED_BYTES)
		}
	case 711:
		{
			yyVAL.str = string(SLAVE_BYTES)
		}
	case 712:
		{
			yyVAL.str = string(SMALLINT_BYTES)
		}
	case 713:
		{
			yyVAL.str = string(SOME_BYTES)
		}
	case 714:
		{
			yyVAL.str = string(SQL_TSI_DAY_BYTES)
		}
	case 715:
		{
			yyVAL.str = string(SQL_TSI_HOUR_BYTES)
		}
	case 716:
		{
			yyVAL.str = string(SQL_TSI_MINUTE_BYTES)
		}
	case 717:
		{
			yyVAL.str = string(SQL_TSI_MONTH_BYTES)
		}
	case 718:
		{
			yyVAL.str = string(SQL_TSI_QUARTER_BYTES)
		}
	case 719:
		{
			yyVAL.str = string(SQL_TSI_SECOND_BYTES)
		}
	case 720:
		{
			yyVAL.str = string(SQL_TSI_WEEK_BYTES)
		}
	case 721:
		{
			yyVAL.str = string(SQL_TSI_YEAR_BYTES)
		}
	case 722:
		{
			yyVAL.str = string(STATUS_BYTES)
		}
	case 723:
		{
			yyVAL.str = string(STORAGE_BYTES)
		}
	case 724:
		{
			yyVAL.str = string(TABLES_BYTES)
		}
	case 725:
		{
			yyVAL.str = string(TEMPORARY_BYTES)
		}
	case 726:
		{
			yyVAL.str = string(TIME_BYTES)
		}
	case 727:
		{
			yyVAL.str = string(TIMESTAMP_BYTES)
		}
	case 728:
		{
			yyVAL.str = string(TIMESTAMPADD_BYTES)
		}
	case 729:
		{
			yyVAL.str = string(TIMESTAMPDIFF_BYTES)
		}
	case 730:
		{
			yyVAL.str = string(TINYINT_BYTES)
		}
	case 731:
		{
			yyVAL.str = string(TRANSACTION_BYTES)
		}
	case 732:
		{
			yyVAL.str = string(TRIGGERS_BYTES)
		}
	case 733:
		{
			yyVAL.str = string(UNCOMMITTED_BYTES)
		}
	case 734:
		{
			yyVAL.str = string(UNKNOWN_BYTES)
		}
	case 735:
		{
			yyVAL.str = string(USER_BYTES)
		}
	case 736:
		{
			yyVAL.str = string(VALUE_BYTES)
		}
	case 737:
		{
			yyVAL.str = string(VARIABLES_BYTES)
		}
	case 738:
		{
			yyVAL.str = string(VIEW_BYTES)
		}
	case 739:
		{
			yyVAL.str = string(WARNINGS_BYTES)
		}
	case 740:
		{
			yyVAL.str = string(WEEK_BYTES)
		}
	case 741:
		{
			yyVAL.str = string(YEAR_BYTES)
		}

	}

	if yyEx != nil && yyEx.Reduced(r, exState, &yyVAL) {
		return -1
	}
	goto yystack /* stack new state and value */
}
