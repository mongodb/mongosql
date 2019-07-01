package parser

import __yyfmt__ "fmt"

import (
	"bytes"

	"github.com/10gen/sqlproxy/internal/option"
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
	yys          int
	empty        struct{}
	statement    Statement
	selStmt      SelectStatement
	bool         bool
	byt          byte
	bytes        []byte
	cte          *CTE
	cte_list     []*CTE
	with         *With
	str          string
	strs         []string
	stropt       OptString
	selectExprs  SelectExprs
	selectExpr   SelectExpr
	columns      Columns
	columnExprs  ColumnExprs
	colName      *ColName
	tableExprs   TableExprs
	tableExpr    TableExpr
	smTableExpr  SimpleTableExpr
	tableName    *TableName
	indexHints   *IndexHints
	expr         Expr
	tuple        Tuple
	exprs        Exprs
	subquery     *Subquery
	caseExpr     *CaseExpr
	whens        []*When
	when         *When
	orderBy      OrderBy
	order        *Order
	limit        *Limit
	updateExprs  UpdateExprs
	updateExpr   *UpdateExpr
	alterSpec    *AlterSpec
	alterSpecs   []*AlterSpec
	renameSpec   *RenameSpec
	renameSpecs  []*RenameSpec
	queryGlobals *QueryGlobals
}

type yyXError struct {
	state, xsym int
}

const (
	yyDefault          = 57608
	yyEofCode          = 57344
	ADD                = 57380
	ADDDATE            = 57414
	ALL                = 57386
	ALTER              = 57379
	AND                = 57575
	ANY                = 57372
	AS                 = 57389
	ASC                = 57392
	BETWEEN            = 57577
	BIGINT             = 57431
	BINARY             = 57443
	BINLOG             = 57451
	BIT_AND            = 57594
	BIT_OR             = 57595
	BOOLEAN            = 57436
	BOTH               = 57437
	BY                 = 57367
	CARET              = 57596
	CASCADE            = 57525
	CASE               = 57578
	CAST               = 57419
	CHANGE             = 57381
	CHANNEL            = 57473
	CHAR               = 57420
	CHARACTER          = 57497
	CHARSET            = 57508
	CODE               = 57461
	COLLATE            = 57498
	COLLATION          = 57505
	COLUMN             = 57384
	COLUMNS            = 57504
	COMMA              = 57561
	COMMENT            = 57351
	COMMITTED          = 57493
	CONNECTION         = 57519
	CONVERT            = 57418
	COUNT              = 57460
	CREATE             = 57359
	CROSS              = 57568
	CURRENT_DATE       = 57402
	CURRENT_TIMESTAMP  = 57401
	DATABASE           = 57446
	DATABASES          = 57499
	DATE               = 57397
	DATETIME           = 57398
	DATE_ADD           = 57413
	DATE_SUB           = 57415
	DAY                = 57532
	DAY_HOUR           = 57546
	DAY_MICROSECOND    = 57543
	DAY_MINUTE         = 57545
	DAY_SECOND         = 57544
	DECIMAL            = 57405
	DEFAULT            = 57395
	DESC               = 57393
	DESCRIBE           = 57510
	DISTINCT           = 57387
	DIV                = 57601
	DOT                = 57603
	DOUBLE             = 57432
	DROP               = 57358
	DUAL               = 57481
	ELSE               = 57581
	END                = 57605
	ENGINE             = 57455
	ENGINES            = 57457
	EQ                 = 57582
	ERRORS             = 57459
	ESCAPE             = 57479
	EVENT              = 57448
	EVENTS             = 57452
	EXCEPT             = 57559
	EXISTS             = 57390
	EXPLAIN            = 57509
	EXTENDED           = 57511
	EXTRACT            = 57412
	FALSE              = 57374
	FLOAT              = 57406
	FLUSH              = 57517
	FN                 = 57477
	FOR                = 57370
	FORCE              = 57570
	FORMAT             = 57513
	FROM               = 57556
	FULL               = 57503
	FUNCTION           = 57449
	GE                 = 57584
	GLOBAL             = 57522
	GRANTS             = 57462
	GROUP              = 57364
	GROUP_CONCAT       = 57408
	GT                 = 57585
	HAVING             = 57365
	HOSTS              = 57470
	HOUR               = 57533
	HOUR_MICROSECOND   = 57540
	HOUR_MINUTE        = 57542
	HOUR_SECOND        = 57541
	ID                 = 57347
	IDIV               = 57602
	IF                 = 57485
	IGNORE             = 57484
	IN                 = 57593
	INDEX              = 57482
	INDEXES            = 57474
	INNER              = 57566
	INT                = 57430
	INTEGER            = 57428
	INTERSECT          = 57560
	INTERVAL           = 57606
	IS                 = 57589
	ISOLATION          = 57487
	JOIN               = 57562
	JSON               = 57515
	KEYS               = 57475
	KILL               = 57516
	LBRACE             = 57354
	LE                 = 57586
	LEADING            = 57438
	LEFT               = 57564
	LEVEL              = 57488
	LEX_ERROR          = 57346
	LIKE               = 57590
	LIMIT              = 57368
	LOCK               = 57396
	LOGS               = 57445
	LPAREN             = 57352
	LT                 = 57587
	MASTER             = 57444
	MICROSECOND        = 57536
	MINUS              = 57558
	MINUTE             = 57534
	MINUTE_MICROSECOND = 57538
	MINUTE_SECOND      = 57539
	MOD                = 57600
	MODIFY             = 57382
	MONTH              = 57530
	MUTEX              = 57456
	NAMES              = 57496
	NATURAL            = 57571
	NCHAR              = 57407
	NE                 = 57588
	NOT                = 57576
	NULL               = 57391
	NULL_SAFE_EQUAL    = 57583
	NUMBER             = 57349
	NUMERIC            = 57433
	OBJECT_ID          = 57409
	OFF                = 57527
	OFFSET             = 57369
	OJ                 = 57478
	ON                 = 57572
	ONLY               = 57491
	OPEN               = 57463
	OR                 = 57573
	ORDER              = 57366
	OUTER              = 57567
	PARTITIONS         = 57512
	PLUGINS            = 57464
	PLUS               = 57597
	PRECISION          = 57388
	PRIVILEGES         = 57465
	PROCEDURE          = 57450
	PROCESSLIST        = 57506
	PROFILE            = 57466
	PROFILES           = 57467
	PROXY              = 57501
	QUARTER            = 57529
	QUERY              = 57520
	RBRACE             = 57355
	READ               = 57489
	RECURSIVE          = 57377
	REGEXP             = 57591
	RELAYLOG           = 57468
	RENAME             = 57383
	REPEATABLE         = 57492
	RESTRICT           = 57524
	RIGHT              = 57565
	RLIKE              = 57592
	ROW                = 57417
	RPAREN             = 57353
	SAMPLE             = 57518
	SCHEMA             = 57447
	SCHEMAS            = 57476
	SECOND             = 57535
	SECOND_MICROSECOND = 57537
	SELECT             = 57357
	SEPARATOR          = 57378
	SERIALIZABLE       = 57495
	SESSION            = 57521
	SET                = 57360
	SHOW               = 57361
	SIGNED             = 57421
	SLAVE              = 57469
	SOME               = 57371
	SQL_BIGINT         = 57423
	SQL_DATE           = 57425
	SQL_DOUBLE         = 57427
	SQL_TIMESTAMP      = 57426
	SQL_TSI_DAY        = 57552
	SQL_TSI_HOUR       = 57553
	SQL_TSI_MINUTE     = 57554
	SQL_TSI_MONTH      = 57550
	SQL_TSI_QUARTER    = 57549
	SQL_TSI_SECOND     = 57555
	SQL_TSI_WEEK       = 57551
	SQL_TSI_YEAR       = 57548
	SQL_VARCHAR        = 57424
	STATUS             = 57507
	STORAGE            = 57458
	STRAIGHT_JOIN      = 57563
	STRING             = 57348
	SUB                = 57598
	SUBDATE            = 57416
	SUBSTR             = 57442
	SUBSTRING          = 57441
	TABLE              = 57480
	TABLES             = 57500
	TEMPORARY          = 57523
	TEXT               = 57434
	THEN               = 57580
	TILDE              = 57356
	TIME               = 57399
	TIMES              = 57599
	TIMESTAMP          = 57400
	TIMESTAMPADD       = 57410
	TIMESTAMPDIFF      = 57411
	TINYINT            = 57429
	TO                 = 57385
	TRADITIONAL        = 57514
	TRAILING           = 57439
	TRANSACTION        = 57486
	TRIGGER            = 57453
	TRIGGERS           = 57471
	TRIM               = 57440
	TRUE               = 57373
	UNARY              = 57604
	UNCOMMITTED        = 57494
	UNION              = 57557
	UNKNOWN            = 57375
	UNSIGNED           = 57422
	UPDATE             = 57362
	USE                = 57569
	USER               = 57454
	USING              = 57526
	UTC_DATE           = 57404
	UTC_TIMESTAMP      = 57403
	VALUES             = 57394
	VALUE_ARG          = 57350
	VARCHAR            = 57435
	VARIABLES          = 57502
	VIEW               = 57483
	WARNINGS           = 57472
	WEEK               = 57531
	WHEN               = 57579
	WHERE              = 57363
	WITH               = 57376
	WRITE              = 57490
	XOR                = 57574
	YEAR               = 57528
	YEAR_MONTH         = 57547
	yyErrCode          = 57345

	yyMaxDepth = 200
	yyTabOfs   = -593
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
		NOT:                12,
		BETWEEN:            13,
		CASE:               13,
		WHEN:               13,
		THEN:               13,
		ELSE:               13,
		EQ:                 14,
		NULL_SAFE_EQUAL:    14,
		GE:                 14,
		GT:                 14,
		LE:                 14,
		LT:                 14,
		NE:                 14,
		IS:                 14,
		LIKE:               14,
		REGEXP:             14,
		RLIKE:              14,
		IN:                 14,
		BIT_AND:            15,
		BIT_OR:             15,
		CARET:              15,
		PLUS:               16,
		SUB:                16,
		TIMES:              17,
		MOD:                17,
		DIV:                17,
		IDIV:               17,
		DOT:                18,
		UNARY:              19,
		END:                20,
		INTERVAL:           21,
	}

	yyXLAT = map[int]int{
		57344: 0,   // $end (551x)
		57347: 1,   // ID (542x)
		57528: 2,   // YEAR (542x)
		57532: 3,   // DAY (541x)
		57533: 4,   // HOUR (541x)
		57536: 5,   // MICROSECOND (541x)
		57534: 6,   // MINUTE (541x)
		57530: 7,   // MONTH (541x)
		57529: 8,   // QUARTER (541x)
		57535: 9,   // SECOND (541x)
		57531: 10,  // WEEK (541x)
		57552: 11,  // SQL_TSI_DAY (540x)
		57553: 12,  // SQL_TSI_HOUR (540x)
		57554: 13,  // SQL_TSI_MINUTE (540x)
		57550: 14,  // SQL_TSI_MONTH (540x)
		57549: 15,  // SQL_TSI_QUARTER (540x)
		57555: 16,  // SQL_TSI_SECOND (540x)
		57551: 17,  // SQL_TSI_WEEK (540x)
		57548: 18,  // SQL_TSI_YEAR (540x)
		57507: 19,  // STATUS (514x)
		57349: 20,  // NUMBER (510x)
		57397: 21,  // DATE (508x)
		57500: 22,  // TABLES (508x)
		57502: 23,  // VARIABLES (508x)
		57504: 24,  // COLUMNS (507x)
		57398: 25,  // DATETIME (507x)
		57405: 26,  // DECIMAL (507x)
		57457: 27,  // ENGINES (507x)
		57452: 28,  // EVENTS (507x)
		57406: 29,  // FLOAT (507x)
		57445: 30,  // LOGS (507x)
		57506: 31,  // PROCESSLIST (507x)
		57399: 32,  // TIME (507x)
		57461: 33,  // CODE (506x)
		57460: 34,  // COUNT (506x)
		57459: 35,  // ERRORS (506x)
		57449: 36,  // FUNCTION (506x)
		57487: 37,  // ISOLATION (506x)
		57407: 38,  // NCHAR (506x)
		57409: 39,  // OBJECT_ID (506x)
		57417: 40,  // ROW (506x)
		57421: 41,  // SIGNED (506x)
		57400: 42,  // TIMESTAMP (506x)
		57375: 43,  // UNKNOWN (506x)
		57454: 44,  // USER (506x)
		57472: 45,  // WARNINGS (506x)
		57451: 46,  // BINLOG (505x)
		57473: 47,  // CHANNEL (505x)
		57508: 48,  // CHARSET (505x)
		57505: 49,  // COLLATION (505x)
		57493: 50,  // COMMITTED (505x)
		57519: 51,  // CONNECTION (505x)
		57455: 52,  // ENGINE (505x)
		57448: 53,  // EVENT (505x)
		57503: 54,  // FULL (505x)
		57462: 55,  // GRANTS (505x)
		57470: 56,  // HOSTS (505x)
		57474: 57,  // INDEXES (505x)
		57515: 58,  // JSON (505x)
		57488: 59,  // LEVEL (505x)
		57444: 60,  // MASTER (505x)
		57456: 61,  // MUTEX (505x)
		57369: 62,  // OFFSET (505x)
		57491: 63,  // ONLY (505x)
		57463: 64,  // OPEN (505x)
		57464: 65,  // PLUGINS (505x)
		57465: 66,  // PRIVILEGES (505x)
		57466: 67,  // PROFILE (505x)
		57467: 68,  // PROFILES (505x)
		57501: 69,  // PROXY (505x)
		57468: 70,  // RELAYLOG (505x)
		57492: 71,  // REPEATABLE (505x)
		57495: 72,  // SERIALIZABLE (505x)
		57469: 73,  // SLAVE (505x)
		57458: 74,  // STORAGE (505x)
		57523: 75,  // TEMPORARY (505x)
		57410: 76,  // TIMESTAMPADD (505x)
		57411: 77,  // TIMESTAMPDIFF (505x)
		57486: 78,  // TRANSACTION (505x)
		57471: 79,  // TRIGGERS (505x)
		57494: 80,  // UNCOMMITTED (505x)
		57483: 81,  // VIEW (505x)
		57372: 82,  // ANY (504x)
		57511: 83,  // EXTENDED (504x)
		57513: 84,  // FORMAT (504x)
		57496: 85,  // NAMES (504x)
		57512: 86,  // PARTITIONS (504x)
		57520: 87,  // QUERY (504x)
		57371: 88,  // SOME (504x)
		57353: 89,  // RPAREN (468x)
		57564: 90,  // LEFT (457x)
		57565: 91,  // RIGHT (457x)
		57348: 92,  // STRING (448x)
		57561: 93,  // COMMA (444x)
		57597: 94,  // PLUS (408x)
		57600: 95,  // MOD (391x)
		57598: 96,  // SUB (390x)
		57559: 97,  // EXCEPT (374x)
		57560: 98,  // INTERSECT (374x)
		57558: 99,  // MINUS (374x)
		57557: 100, // UNION (374x)
		57370: 101, // FOR (361x)
		57352: 102, // LPAREN (361x)
		57368: 103, // LIMIT (360x)
		57354: 104, // LBRACE (358x)
		57576: 105, // NOT (351x)
		57363: 106, // WHERE (346x)
		57388: 107, // PRECISION (341x)
		57396: 108, // LOCK (340x)
		57366: 109, // ORDER (339x)
		57562: 110, // JOIN (328x)
		57546: 111, // DAY_HOUR (326x)
		57543: 112, // DAY_MICROSECOND (326x)
		57545: 113, // DAY_MINUTE (326x)
		57544: 114, // DAY_SECOND (326x)
		57540: 115, // HOUR_MICROSECOND (326x)
		57542: 116, // HOUR_MINUTE (326x)
		57541: 117, // HOUR_SECOND (326x)
		57538: 118, // MINUTE_MICROSECOND (326x)
		57539: 119, // MINUTE_SECOND (326x)
		57537: 120, // SECOND_MICROSECOND (326x)
		57547: 121, // YEAR_MONTH (326x)
		57575: 122, // AND (325x)
		57563: 123, // STRAIGHT_JOIN (325x)
		57556: 124, // FROM (324x)
		57365: 125, // HAVING (323x)
		57573: 126, // OR (323x)
		57599: 127, // TIMES (323x)
		57574: 128, // XOR (323x)
		57590: 129, // LIKE (322x)
		57364: 130, // GROUP (318x)
		57568: 131, // CROSS (317x)
		57566: 132, // INNER (317x)
		57571: 133, // NATURAL (317x)
		57355: 134, // RBRACE (315x)
		57572: 135, // ON (312x)
		57378: 136, // SEPARATOR (311x)
		57526: 137, // USING (311x)
		57389: 138, // AS (304x)
		57579: 139, // WHEN (295x)
		57605: 140, // END (294x)
		57581: 141, // ELSE (292x)
		57393: 142, // DESC (290x)
		57392: 143, // ASC (289x)
		57580: 144, // THEN (289x)
		57582: 145, // EQ (288x)
		57584: 146, // GE (286x)
		57585: 147, // GT (286x)
		57589: 148, // IS (286x)
		57586: 149, // LE (286x)
		57587: 150, // LT (286x)
		57588: 151, // NE (286x)
		57583: 152, // NULL_SAFE_EQUAL (286x)
		57593: 153, // IN (262x)
		57594: 154, // BIT_AND (254x)
		57595: 155, // BIT_OR (254x)
		57596: 156, // CARET (254x)
		57601: 157, // DIV (254x)
		57602: 158, // IDIV (254x)
		57577: 159, // BETWEEN (249x)
		57591: 160, // REGEXP (249x)
		57592: 161, // RLIKE (249x)
		57479: 162, // ESCAPE (201x)
		57603: 163, // DOT (162x)
		57666: 164, // keyword_as_id (157x)
		57697: 165, // sql_id (157x)
		57485: 166, // IF (142x)
		57420: 167, // CHAR (141x)
		57606: 168, // INTERVAL (141x)
		57446: 169, // DATABASE (140x)
		57390: 170, // EXISTS (140x)
		57374: 171, // FALSE (140x)
		57391: 172, // NULL (140x)
		57447: 173, // SCHEMA (140x)
		57373: 174, // TRUE (140x)
		57414: 175, // ADDDATE (139x)
		57419: 176, // CAST (139x)
		57418: 177, // CONVERT (139x)
		57402: 178, // CURRENT_DATE (139x)
		57401: 179, // CURRENT_TIMESTAMP (139x)
		57413: 180, // DATE_ADD (139x)
		57415: 181, // DATE_SUB (139x)
		57412: 182, // EXTRACT (139x)
		57408: 183, // GROUP_CONCAT (139x)
		57416: 184, // SUBDATE (139x)
		57442: 185, // SUBSTR (139x)
		57441: 186, // SUBSTRING (139x)
		57440: 187, // TRIM (139x)
		57404: 188, // UTC_DATE (139x)
		57403: 189, // UTC_TIMESTAMP (139x)
		57578: 190, // CASE (138x)
		57356: 191, // TILDE (138x)
		57350: 192, // VALUE_ARG (138x)
		57394: 193, // VALUES (138x)
		57703: 194, // subquery (135x)
		57624: 195, // column_name (124x)
		57618: 196, // boolean_value (119x)
		57716: 197, // tuple (119x)
		57651: 198, // func_expr (118x)
		57652: 199, // func_expr_conflict (118x)
		57653: 200, // func_expr_generic (118x)
		57654: 201, // func_expr_reserved_keyword (118x)
		57655: 202, // func_expr_unconventional (118x)
		57704: 203, // substr (118x)
		57621: 204, // case_expression (117x)
		57695: 205, // simple_expr (117x)
		57717: 206, // unary_operator (117x)
		57720: 207, // value (117x)
		57616: 208, // bit_expr (110x)
		57569: 209, // USE (103x)
		57570: 210, // FORCE (102x)
		57484: 211, // IGNORE (102x)
		57443: 212, // BINARY (98x)
		57428: 213, // INTEGER (97x)
		57358: 214, // DROP (95x)
		57383: 215, // RENAME (95x)
		57381: 216, // CHANGE (94x)
		57382: 217, // MODIFY (94x)
		57678: 218, // predicate (94x)
		57385: 219, // TO (94x)
		57431: 220, // BIGINT (93x)
		57436: 221, // BOOLEAN (93x)
		57525: 222, // CASCADE (93x)
		57432: 223, // DOUBLE (93x)
		57430: 224, // INT (93x)
		57433: 225, // NUMERIC (93x)
		57524: 226, // RESTRICT (93x)
		57434: 227, // TEXT (93x)
		57429: 228, // TINYINT (93x)
		57435: 229, // VARCHAR (93x)
		57498: 230, // COLLATE (92x)
		57617: 231, // bool_pri (85x)
		57643: 232, // expression (85x)
		57682: 233, // select_expression (50x)
		57357: 234, // SELECT (35x)
		57376: 235, // WITH (35x)
		57683: 236, // select_expression_list (31x)
		57670: 237, // like_or_where_opt (16x)
		57673: 238, // non_derived_subquery (15x)
		57684: 239, // select_statement (15x)
		57707: 240, // table_name (15x)
		57724: 241, // with_statement (15x)
		57387: 242, // DISTINCT (11x)
		57699: 243, // sql_time_interval (11x)
		57711: 244, // time_interval (11x)
		57718: 245, // union_op (11x)
		57660: 246, // in_or_from (10x)
		57663: 247, // interval_unit (9x)
		57671: 248, // limit_opt (9x)
		57691: 249, // show_from_in (9x)
		57700: 250, // sql_time_unit (9x)
		57386: 251, // ALL (8x)
		57664: 252, // join_expression (8x)
		57696: 253, // simple_table_expression (8x)
		57705: 254, // table_expression (8x)
		57609: 255, // all_any_some (7x)
		57480: 256, // TABLE (7x)
		57351: 257, // COMMENT (6x)
		57665: 258, // join_type (6x)
		57644: 259, // expression_list (5x)
		57692: 260, // show_from_in_opt (5x)
		57723: 261, // where_expression_opt (5x)
		57614: 262, // as_opt (4x)
		57482: 263, // INDEX (4x)
		57669: 264, // like_escape_opt (4x)
		57674: 265, // optional_parens (4x)
		57676: 266, // order_by_opt (4x)
		57567: 267, // OUTER (4x)
		57489: 268, // READ (4x)
		57607: 269, // $@1 (3x)
		57384: 270, // COLUMN (3x)
		57625: 271, // column_opt (3x)
		57628: 272, // comment_opt (3x)
		57629: 273, // cte (3x)
		57650: 274, // from_opt (3x)
		57662: 275, // index_list (3x)
		57360: 276, // SET (3x)
		57688: 277, // set_spec (3x)
		57698: 278, // sql_id_or_string (3x)
		57610: 279, // alter_spec (2x)
		57367: 280, // BY (2x)
		57497: 281, // CHARACTER (2x)
		57623: 282, // column_expression_list (2x)
		57630: 283, // cte_list (2x)
		57481: 284, // DUAL (2x)
		57634: 285, // dual_table (2x)
		57641: 286, // explainable_stmt (2x)
		57522: 287, // GLOBAL (2x)
		57656: 288, // group_by_opt (2x)
		57657: 289, // having_opt (2x)
		57658: 290, // if_not_exists_opt (2x)
		57659: 291, // in_opt (2x)
		57672: 292, // lock_opt (2x)
		57675: 293, // order (2x)
		57450: 294, // PROCEDURE (2x)
		57679: 295, // query_globals_opt (2x)
		57681: 296, // scope_modifier_opt (2x)
		57521: 297, // SESSION (2x)
		57423: 298, // SQL_BIGINT (2x)
		57425: 299, // SQL_DATE (2x)
		57427: 300, // SQL_DOUBLE (2x)
		57426: 301, // SQL_TIMESTAMP (2x)
		57701: 302, // sql_types (2x)
		57424: 303, // SQL_VARCHAR (2x)
		57706: 304, // table_expression_list (2x)
		57708: 305, // table_rename (2x)
		57713: 306, // transaction_characteristic (2x)
		57422: 307, // UNSIGNED (2x)
		57721: 308, // when_expression (2x)
		57379: 309, // ALTER (1x)
		57611: 310, // alter_spec_list (1x)
		57612: 311, // alter_statement (1x)
		57613: 312, // any_command (1x)
		57615: 313, // asc_desc_opt (1x)
		57437: 314, // BOTH (1x)
		57619: 315, // both_leading_trailing_opt (1x)
		57620: 316, // cascade_or_restrict_opt (1x)
		57622: 317, // column_definition (1x)
		57626: 318, // command (1x)
		57627: 319, // comment_list (1x)
		57359: 320, // CREATE (1x)
		57631: 321, // data_type (1x)
		57499: 322, // DATABASES (1x)
		57395: 323, // DEFAULT (1x)
		57510: 324, // DESCRIBE (1x)
		57632: 325, // distinct_opt (1x)
		57633: 326, // drop_statement (1x)
		57635: 327, // else_expression_opt (1x)
		57636: 328, // exists_opt (1x)
		57509: 329, // EXPLAIN (1x)
		57637: 330, // explain_alias (1x)
		57638: 331, // explain_column_name (1x)
		57639: 332, // explain_statement (1x)
		57640: 333, // explain_type (1x)
		57642: 334, // explicit_scope_modifier_opt (1x)
		57645: 335, // expression_opt (1x)
		57517: 336, // FLUSH (1x)
		57646: 337, // flush_statement (1x)
		57477: 338, // FN (1x)
		57647: 339, // for_channel_opt (1x)
		57648: 340, // for_user_opt (1x)
		57649: 341, // format_name (1x)
		57661: 342, // index_hint_list (1x)
		57475: 343, // KEYS (1x)
		57516: 344, // KILL (1x)
		57667: 345, // kill_modifier (1x)
		57668: 346, // kill_statement (1x)
		57438: 347, // LEADING (1x)
		57527: 348, // OFF (1x)
		57478: 349, // OJ (1x)
		57677: 350, // order_list (1x)
		57377: 351, // RECURSIVE (1x)
		57680: 352, // rename_statement (1x)
		57518: 353, // SAMPLE (1x)
		57476: 354, // SCHEMAS (1x)
		57685: 355, // select_statement_with_paren_order_limit (1x)
		57686: 356, // separator_opt (1x)
		57687: 357, // set_expr (1x)
		57689: 358, // set_spec_list (1x)
		57690: 359, // set_statement (1x)
		57361: 360, // SHOW (1x)
		57693: 361, // show_full (1x)
		57694: 362, // show_statement (1x)
		57702: 363, // storage_opt (1x)
		57709: 364, // table_rename_list (1x)
		57710: 365, // temporary_opt (1x)
		57712: 366, // to_as_opt (1x)
		57514: 367, // TRADITIONAL (1x)
		57439: 368, // TRAILING (1x)
		57714: 369, // transaction_characteristics (1x)
		57715: 370, // transaction_level (1x)
		57453: 371, // TRIGGER (1x)
		57362: 372, // UPDATE (1x)
		57719: 373, // use_statement (1x)
		57722: 374, // when_expression_list (1x)
		57490: 375, // WRITE (1x)
		57608: 376, // $default (0x)
		57380: 377, // ADD (0x)
		57345: 378, // error (0x)
		57346: 379, // LEX_ERROR (0x)
		57604: 380, // UNARY (0x)
	}

	yySymNames = []string{
		"$end",
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
		"STATUS",
		"NUMBER",
		"DATE",
		"TABLES",
		"VARIABLES",
		"COLUMNS",
		"DATETIME",
		"DECIMAL",
		"ENGINES",
		"EVENTS",
		"FLOAT",
		"LOGS",
		"PROCESSLIST",
		"TIME",
		"CODE",
		"COUNT",
		"ERRORS",
		"FUNCTION",
		"ISOLATION",
		"NCHAR",
		"OBJECT_ID",
		"ROW",
		"SIGNED",
		"TIMESTAMP",
		"UNKNOWN",
		"USER",
		"WARNINGS",
		"BINLOG",
		"CHANNEL",
		"CHARSET",
		"COLLATION",
		"COMMITTED",
		"CONNECTION",
		"ENGINE",
		"EVENT",
		"FULL",
		"GRANTS",
		"HOSTS",
		"INDEXES",
		"JSON",
		"LEVEL",
		"MASTER",
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
		"SERIALIZABLE",
		"SLAVE",
		"STORAGE",
		"TEMPORARY",
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
		"RPAREN",
		"LEFT",
		"RIGHT",
		"STRING",
		"COMMA",
		"PLUS",
		"MOD",
		"SUB",
		"EXCEPT",
		"INTERSECT",
		"MINUS",
		"UNION",
		"FOR",
		"LPAREN",
		"LIMIT",
		"LBRACE",
		"NOT",
		"WHERE",
		"PRECISION",
		"LOCK",
		"ORDER",
		"JOIN",
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
		"AND",
		"STRAIGHT_JOIN",
		"FROM",
		"HAVING",
		"OR",
		"TIMES",
		"XOR",
		"LIKE",
		"GROUP",
		"CROSS",
		"INNER",
		"NATURAL",
		"RBRACE",
		"ON",
		"SEPARATOR",
		"USING",
		"AS",
		"WHEN",
		"END",
		"ELSE",
		"DESC",
		"ASC",
		"THEN",
		"EQ",
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
		"ESCAPE",
		"DOT",
		"keyword_as_id",
		"sql_id",
		"IF",
		"CHAR",
		"INTERVAL",
		"DATABASE",
		"EXISTS",
		"FALSE",
		"NULL",
		"SCHEMA",
		"TRUE",
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
		"VALUE_ARG",
		"VALUES",
		"subquery",
		"column_name",
		"boolean_value",
		"tuple",
		"func_expr",
		"func_expr_conflict",
		"func_expr_generic",
		"func_expr_reserved_keyword",
		"func_expr_unconventional",
		"substr",
		"case_expression",
		"simple_expr",
		"unary_operator",
		"value",
		"bit_expr",
		"USE",
		"FORCE",
		"IGNORE",
		"BINARY",
		"INTEGER",
		"DROP",
		"RENAME",
		"CHANGE",
		"MODIFY",
		"predicate",
		"TO",
		"BIGINT",
		"BOOLEAN",
		"CASCADE",
		"DOUBLE",
		"INT",
		"NUMERIC",
		"RESTRICT",
		"TEXT",
		"TINYINT",
		"VARCHAR",
		"COLLATE",
		"bool_pri",
		"expression",
		"select_expression",
		"SELECT",
		"WITH",
		"select_expression_list",
		"like_or_where_opt",
		"non_derived_subquery",
		"select_statement",
		"table_name",
		"with_statement",
		"DISTINCT",
		"sql_time_interval",
		"time_interval",
		"union_op",
		"in_or_from",
		"interval_unit",
		"limit_opt",
		"show_from_in",
		"sql_time_unit",
		"ALL",
		"join_expression",
		"simple_table_expression",
		"table_expression",
		"all_any_some",
		"TABLE",
		"COMMENT",
		"join_type",
		"expression_list",
		"show_from_in_opt",
		"where_expression_opt",
		"as_opt",
		"INDEX",
		"like_escape_opt",
		"optional_parens",
		"order_by_opt",
		"OUTER",
		"READ",
		"$@1",
		"COLUMN",
		"column_opt",
		"comment_opt",
		"cte",
		"from_opt",
		"index_list",
		"SET",
		"set_spec",
		"sql_id_or_string",
		"alter_spec",
		"BY",
		"CHARACTER",
		"column_expression_list",
		"cte_list",
		"DUAL",
		"dual_table",
		"explainable_stmt",
		"GLOBAL",
		"group_by_opt",
		"having_opt",
		"if_not_exists_opt",
		"in_opt",
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
		"table_rename",
		"transaction_characteristic",
		"UNSIGNED",
		"when_expression",
		"ALTER",
		"alter_spec_list",
		"alter_statement",
		"any_command",
		"asc_desc_opt",
		"BOTH",
		"both_leading_trailing_opt",
		"cascade_or_restrict_opt",
		"column_definition",
		"command",
		"comment_list",
		"CREATE",
		"data_type",
		"DATABASES",
		"DEFAULT",
		"DESCRIBE",
		"distinct_opt",
		"drop_statement",
		"else_expression_opt",
		"exists_opt",
		"EXPLAIN",
		"explain_alias",
		"explain_column_name",
		"explain_statement",
		"explain_type",
		"explicit_scope_modifier_opt",
		"expression_opt",
		"FLUSH",
		"flush_statement",
		"FN",
		"for_channel_opt",
		"for_user_opt",
		"format_name",
		"index_hint_list",
		"KEYS",
		"KILL",
		"kill_modifier",
		"kill_statement",
		"LEADING",
		"OFF",
		"OJ",
		"order_list",
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
		"storage_opt",
		"table_rename_list",
		"temporary_opt",
		"to_as_opt",
		"TRADITIONAL",
		"TRAILING",
		"transaction_characteristics",
		"transaction_level",
		"TRIGGER",
		"UPDATE",
		"use_statement",
		"when_expression_list",
		"WRITE",
		"$default",
		"ADD",
		"error",
		"LEX_ERROR",
		"UNARY",
	}

	yyTokenLiteralStrings = map[int]string{}

	yyReductions = map[int]struct{ xsym, components int }{
		0:   {0, 1},
		1:   {312, 1},
		2:   {318, 1},
		3:   {318, 1},
		4:   {318, 1},
		5:   {318, 1},
		6:   {318, 1},
		7:   {318, 1},
		8:   {318, 1},
		9:   {318, 1},
		10:  {318, 1},
		11:  {318, 1},
		12:  {318, 1},
		13:  {355, 3},
		14:  {239, 1},
		15:  {239, 5},
		16:  {239, 6},
		17:  {239, 12},
		18:  {239, 4},
		19:  {239, 13},
		20:  {239, 3},
		21:  {273, 5},
		22:  {273, 8},
		23:  {283, 1},
		24:  {283, 3},
		25:  {241, 2},
		26:  {241, 3},
		27:  {238, 3},
		28:  {373, 2},
		29:  {359, 3},
		30:  {359, 3},
		31:  {359, 3},
		32:  {359, 5},
		33:  {359, 4},
		34:  {359, 4},
		35:  {358, 1},
		36:  {358, 3},
		37:  {277, 3},
		38:  {277, 3},
		39:  {357, 1},
		40:  {357, 1},
		41:  {357, 1},
		42:  {326, 6},
		43:  {337, 2},
		44:  {337, 2},
		45:  {352, 3},
		46:  {364, 1},
		47:  {364, 3},
		48:  {305, 3},
		49:  {311, 4},
		50:  {310, 1},
		51:  {310, 3},
		52:  {279, 4},
		53:  {279, 3},
		54:  {279, 4},
		55:  {279, 3},
		56:  {271, 0},
		57:  {271, 1},
		58:  {317, 1},
		59:  {321, 1},
		60:  {321, 1},
		61:  {321, 1},
		62:  {321, 1},
		63:  {321, 1},
		64:  {321, 1},
		65:  {321, 1},
		66:  {321, 1},
		67:  {321, 1},
		68:  {321, 1},
		69:  {321, 1},
		70:  {321, 1},
		71:  {321, 1},
		72:  {321, 1},
		73:  {321, 1},
		74:  {321, 1},
		75:  {321, 1},
		76:  {366, 0},
		77:  {366, 1},
		78:  {366, 1},
		79:  {365, 0},
		80:  {365, 1},
		81:  {328, 0},
		82:  {328, 2},
		83:  {316, 0},
		84:  {316, 1},
		85:  {316, 1},
		86:  {369, 1},
		87:  {369, 3},
		88:  {306, 3},
		89:  {306, 2},
		90:  {306, 2},
		91:  {370, 2},
		92:  {370, 2},
		93:  {370, 2},
		94:  {370, 1},
		95:  {203, 1},
		96:  {203, 1},
		97:  {246, 1},
		98:  {246, 1},
		99:  {249, 3},
		100: {249, 4},
		101: {249, 4},
		102: {249, 2},
		103: {260, 0},
		104: {260, 1},
		105: {361, 0},
		106: {361, 1},
		107: {345, 1},
		108: {345, 1},
		109: {296, 0},
		110: {296, 1},
		111: {296, 1},
		112: {334, 1},
		113: {334, 1},
		114: {362, 3},
		115: {362, 3},
		116: {362, 6},
		117: {362, 5},
		118: {362, 5},
		119: {362, 4},
		120: {362, 4},
		121: {362, 4},
		122: {362, 6},
		123: {362, 5},
		124: {362, 4},
		125: {362, 4},
		126: {362, 4},
		127: {362, 4},
		128: {362, 4},
		129: {362, 4},
		130: {362, 3},
		131: {362, 3},
		132: {362, 6},
		133: {362, 4},
		134: {362, 4},
		135: {362, 4},
		136: {362, 3},
		137: {362, 4},
		138: {362, 4},
		139: {362, 4},
		140: {362, 3},
		141: {362, 5},
		142: {362, 2},
		143: {362, 2},
		144: {362, 4},
		145: {362, 4},
		146: {362, 2},
		147: {362, 2},
		148: {362, 6},
		149: {362, 3},
		150: {362, 4},
		151: {362, 5},
		152: {362, 4},
		153: {362, 3},
		154: {362, 6},
		155: {362, 3},
		156: {362, 3},
		157: {362, 4},
		158: {362, 5},
		159: {362, 5},
		160: {362, 5},
		161: {362, 3},
		162: {362, 4},
		163: {362, 4},
		164: {362, 3},
		165: {362, 3},
		166: {341, 1},
		167: {341, 1},
		168: {333, 1},
		169: {333, 1},
		170: {333, 3},
		171: {333, 0},
		172: {286, 1},
		173: {286, 3},
		174: {330, 1},
		175: {330, 1},
		176: {330, 1},
		177: {240, 1},
		178: {240, 2},
		179: {240, 3},
		180: {285, 1},
		181: {332, 3},
		182: {332, 3},
		183: {332, 5},
		184: {331, 0},
		185: {331, 1},
		186: {331, 1},
		187: {346, 2},
		188: {346, 3},
		189: {269, 0},
		190: {272, 2},
		191: {319, 0},
		192: {319, 2},
		193: {295, 0},
		194: {295, 1},
		195: {295, 1},
		196: {295, 2},
		197: {295, 2},
		198: {245, 1},
		199: {245, 2},
		200: {245, 1},
		201: {245, 1},
		202: {245, 1},
		203: {325, 0},
		204: {325, 1},
		205: {356, 0},
		206: {356, 2},
		207: {236, 1},
		208: {236, 3},
		209: {233, 1},
		210: {233, 2},
		211: {233, 3},
		212: {233, 3},
		213: {233, 5},
		214: {282, 1},
		215: {282, 3},
		216: {304, 1},
		217: {304, 3},
		218: {254, 3},
		219: {254, 3},
		220: {254, 1},
		221: {254, 4},
		222: {252, 3},
		223: {252, 3},
		224: {252, 5},
		225: {252, 5},
		226: {252, 7},
		227: {262, 0},
		228: {262, 1},
		229: {262, 2},
		230: {262, 1},
		231: {262, 2},
		232: {258, 1},
		233: {258, 2},
		234: {258, 3},
		235: {258, 2},
		236: {258, 3},
		237: {258, 2},
		238: {258, 2},
		239: {258, 2},
		240: {258, 3},
		241: {258, 4},
		242: {258, 3},
		243: {258, 4},
		244: {253, 1},
		245: {253, 1},
		246: {342, 0},
		247: {342, 5},
		248: {342, 5},
		249: {342, 5},
		250: {275, 1},
		251: {275, 3},
		252: {261, 0},
		253: {261, 2},
		254: {237, 0},
		255: {237, 2},
		256: {237, 2},
		257: {291, 0},
		258: {291, 2},
		259: {290, 0},
		260: {290, 3},
		261: {274, 0},
		262: {274, 2},
		263: {339, 0},
		264: {339, 3},
		265: {340, 0},
		266: {340, 2},
		267: {363, 0},
		268: {363, 1},
		269: {255, 1},
		270: {255, 1},
		271: {255, 1},
		272: {197, 6},
		273: {197, 3},
		274: {194, 3},
		275: {259, 1},
		276: {259, 3},
		277: {232, 3},
		278: {232, 3},
		279: {232, 3},
		280: {232, 2},
		281: {232, 3},
		282: {232, 4},
		283: {232, 1},
		284: {231, 3},
		285: {231, 4},
		286: {231, 3},
		287: {231, 4},
		288: {231, 3},
		289: {231, 4},
		290: {231, 3},
		291: {231, 4},
		292: {231, 3},
		293: {231, 4},
		294: {231, 3},
		295: {231, 4},
		296: {231, 3},
		297: {231, 4},
		298: {231, 3},
		299: {231, 4},
		300: {231, 1},
		301: {218, 3},
		302: {218, 3},
		303: {218, 4},
		304: {218, 4},
		305: {218, 5},
		306: {218, 6},
		307: {218, 4},
		308: {218, 5},
		309: {218, 5},
		310: {218, 6},
		311: {218, 3},
		312: {218, 4},
		313: {218, 3},
		314: {218, 4},
		315: {218, 1},
		316: {208, 3},
		317: {208, 3},
		318: {208, 3},
		319: {208, 5},
		320: {208, 5},
		321: {208, 3},
		322: {208, 5},
		323: {208, 3},
		324: {208, 3},
		325: {208, 3},
		326: {208, 3},
		327: {208, 3},
		328: {208, 1},
		329: {205, 1},
		330: {205, 1},
		331: {205, 1},
		332: {205, 1},
		333: {205, 2},
		334: {205, 2},
		335: {205, 1},
		336: {205, 1},
		337: {205, 4},
		338: {205, 4},
		339: {198, 1},
		340: {198, 1},
		341: {198, 1},
		342: {198, 1},
		343: {201, 4},
		344: {201, 6},
		345: {201, 4},
		346: {201, 4},
		347: {202, 8},
		348: {202, 6},
		349: {202, 6},
		350: {202, 7},
		351: {202, 2},
		352: {202, 2},
		353: {202, 4},
		354: {202, 8},
		355: {202, 8},
		356: {202, 6},
		357: {202, 7},
		358: {202, 6},
		359: {202, 8},
		360: {202, 6},
		361: {202, 6},
		362: {202, 6},
		363: {202, 6},
		364: {202, 4},
		365: {202, 7},
		366: {202, 6},
		367: {202, 8},
		368: {202, 6},
		369: {202, 4},
		370: {202, 2},
		371: {202, 2},
		372: {199, 4},
		373: {199, 5},
		374: {199, 3},
		375: {199, 4},
		376: {199, 4},
		377: {199, 4},
		378: {199, 4},
		379: {199, 4},
		380: {199, 4},
		381: {199, 4},
		382: {199, 4},
		383: {199, 4},
		384: {199, 4},
		385: {199, 3},
		386: {199, 4},
		387: {199, 4},
		388: {199, 4},
		389: {199, 3},
		390: {199, 4},
		391: {200, 3},
		392: {200, 4},
		393: {200, 5},
		394: {265, 0},
		395: {265, 2},
		396: {264, 0},
		397: {264, 2},
		398: {264, 4},
		399: {315, 1},
		400: {315, 1},
		401: {315, 1},
		402: {247, 1},
		403: {247, 1},
		404: {247, 1},
		405: {243, 1},
		406: {243, 1},
		407: {243, 1},
		408: {243, 1},
		409: {243, 1},
		410: {243, 1},
		411: {243, 1},
		412: {243, 1},
		413: {244, 1},
		414: {244, 1},
		415: {244, 1},
		416: {244, 1},
		417: {244, 1},
		418: {244, 1},
		419: {244, 1},
		420: {244, 1},
		421: {244, 1},
		422: {250, 1},
		423: {250, 1},
		424: {250, 1},
		425: {250, 1},
		426: {250, 1},
		427: {250, 1},
		428: {250, 1},
		429: {250, 1},
		430: {250, 1},
		431: {250, 1},
		432: {250, 1},
		433: {302, 1},
		434: {302, 4},
		435: {302, 1},
		436: {302, 4},
		437: {302, 1},
		438: {302, 1},
		439: {302, 1},
		440: {302, 4},
		441: {302, 6},
		442: {302, 1},
		443: {302, 4},
		444: {302, 1},
		445: {302, 1},
		446: {302, 1},
		447: {302, 1},
		448: {302, 2},
		449: {302, 1},
		450: {302, 1},
		451: {302, 2},
		452: {302, 1},
		453: {302, 1},
		454: {302, 1},
		455: {302, 1},
		456: {302, 1},
		457: {206, 1},
		458: {206, 1},
		459: {206, 1},
		460: {204, 5},
		461: {335, 0},
		462: {335, 1},
		463: {374, 1},
		464: {374, 2},
		465: {308, 4},
		466: {327, 0},
		467: {327, 2},
		468: {195, 1},
		469: {195, 3},
		470: {195, 5},
		471: {207, 1},
		472: {207, 1},
		473: {207, 1},
		474: {207, 2},
		475: {207, 2},
		476: {207, 2},
		477: {207, 4},
		478: {207, 1},
		479: {207, 1},
		480: {196, 1},
		481: {196, 1},
		482: {196, 1},
		483: {288, 0},
		484: {288, 3},
		485: {289, 0},
		486: {289, 2},
		487: {266, 0},
		488: {266, 3},
		489: {350, 1},
		490: {350, 3},
		491: {293, 2},
		492: {313, 0},
		493: {313, 1},
		494: {313, 1},
		495: {248, 0},
		496: {248, 2},
		497: {248, 4},
		498: {248, 4},
		499: {292, 0},
		500: {292, 2},
		501: {292, 4},
		502: {165, 1},
		503: {165, 1},
		504: {278, 1},
		505: {278, 1},
		506: {164, 1},
		507: {164, 1},
		508: {164, 1},
		509: {164, 1},
		510: {164, 1},
		511: {164, 1},
		512: {164, 1},
		513: {164, 1},
		514: {164, 1},
		515: {164, 1},
		516: {164, 1},
		517: {164, 1},
		518: {164, 1},
		519: {164, 1},
		520: {164, 1},
		521: {164, 1},
		522: {164, 1},
		523: {164, 1},
		524: {164, 1},
		525: {164, 1},
		526: {164, 1},
		527: {164, 1},
		528: {164, 1},
		529: {164, 1},
		530: {164, 1},
		531: {164, 1},
		532: {164, 1},
		533: {164, 1},
		534: {164, 1},
		535: {164, 1},
		536: {164, 1},
		537: {164, 1},
		538: {164, 1},
		539: {164, 1},
		540: {164, 1},
		541: {164, 1},
		542: {164, 1},
		543: {164, 1},
		544: {164, 1},
		545: {164, 1},
		546: {164, 1},
		547: {164, 1},
		548: {164, 1},
		549: {164, 1},
		550: {164, 1},
		551: {164, 1},
		552: {164, 1},
		553: {164, 1},
		554: {164, 1},
		555: {164, 1},
		556: {164, 1},
		557: {164, 1},
		558: {164, 1},
		559: {164, 1},
		560: {164, 1},
		561: {164, 1},
		562: {164, 1},
		563: {164, 1},
		564: {164, 1},
		565: {164, 1},
		566: {164, 1},
		567: {164, 1},
		568: {164, 1},
		569: {164, 1},
		570: {164, 1},
		571: {164, 1},
		572: {164, 1},
		573: {164, 1},
		574: {164, 1},
		575: {164, 1},
		576: {164, 1},
		577: {164, 1},
		578: {164, 1},
		579: {164, 1},
		580: {164, 1},
		581: {164, 1},
		582: {164, 1},
		583: {164, 1},
		584: {164, 1},
		585: {164, 1},
		586: {164, 1},
		587: {164, 1},
		588: {164, 1},
		589: {164, 1},
		590: {164, 1},
		591: {164, 1},
		592: {164, 1},
	}

	yyXErrors = map[yyXError]string{}

	yyParseTab = [1119][]uint16{
		// 0
		{102: 611, 142: 621, 209: 612, 214: 614, 616, 234: 608, 610, 238: 607, 596, 241: 609, 276: 613, 309: 617, 311: 605, 594, 318: 595, 324: 620, 326: 603, 329: 619, 622, 332: 601, 336: 615, 604, 344: 623, 346: 599, 352: 606, 355: 597, 359: 598, 618, 362: 600, 373: 602},
		{593},
		{592},
		{591, 97: 998, 999, 997, 996, 245: 994},
		{590},
		// 5
		{589},
		{588},
		{587},
		{586},
		{585},
		// 10
		{584},
		{583},
		{582},
		{581},
		{579, 97: 579, 579, 579, 579, 103: 106, 109: 1052, 266: 1710},
		// 15
		{1: 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 90: 404, 404, 404, 94: 404, 404, 404, 102: 404, 104: 404, 404, 123: 404, 127: 404, 166: 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 242: 404, 257: 404, 269: 1533, 272: 1699},
		{102: 611, 234: 1595, 610, 238: 982, 1596, 241: 609},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 1574, 273: 1575, 283: 1576, 351: 1577},
		{102: 611, 234: 608, 610, 238: 982, 1573, 241: 609},
		{1: 1572},
		// 20
		{1: 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 484, 404, 404, 404, 404, 404, 404, 1528, 404, 404, 404, 257: 404, 269: 1533, 272: 1526, 281: 1529, 287: 1532, 296: 1530, 1531, 334: 1527},
		{75: 1517, 256: 514, 365: 1516},
		{30: 1514, 353: 1515},
		{256: 1506},
		{256: 1465},
		// 25
		{19: 484, 22: 488, 484, 488, 27: 326, 1310, 31: 488, 34: 1309, 1308, 1311, 45: 1326, 1304, 48: 1333, 1334, 52: 1306, 54: 1299, 1312, 57: 1314, 60: 1303, 64: 1316, 1317, 1318, 1320, 1321, 1331, 1322, 73: 1323, 1335, 79: 1325, 212: 1302, 256: 1324, 263: 1313, 281: 1332, 287: 1301, 294: 1319, 296: 1329, 1300, 320: 1305, 322: 1327, 343: 1315, 354: 1328, 361: 1330, 363: 1307},
		{1: 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 419, 101: 419, 419, 163: 419, 234: 419, 419},
		{1: 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 418, 101: 418, 418, 163: 418, 234: 418, 418},
		{1: 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 417, 101: 417, 417, 163: 417, 234: 417, 417},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 1272, 1274, 733, 1273, 784, 751, 101: 422, 422, 163: 1276, 703, 1275, 234: 422, 422, 240: 1277, 333: 1278},
		// 30
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 626, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 627, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 628, 345: 629},
		{102: 498},
		{102: 497},
		{79, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 90: 486, 486, 486, 94: 486, 486, 486, 102: 486, 104: 486, 486, 122: 79, 126: 79, 79, 79, 79, 145: 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 163: 79, 166: 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486, 486},
		{35, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 90: 485, 485, 485, 94: 485, 485, 485, 102: 485, 104: 485, 485, 122: 35, 126: 35, 35, 35, 35, 145: 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 163: 35, 166: 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485, 485},
		// 35
		{406, 122: 805, 126: 803, 128: 804},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1271},
		{32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 1225, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 984, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 993, 234: 608, 610, 238: 982, 983, 241: 609, 259: 987},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1270},
		// 40
		{310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 97: 310, 310, 310, 310, 310, 103: 310, 106: 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 128: 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 310, 1233, 1239, 1237, 1232, 1238, 1236, 1234, 1235},
		{293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 97: 293, 293, 293, 293, 293, 103: 293, 106: 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 128: 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293, 293},
		{278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 1025, 1030, 1026, 278, 278, 278, 278, 278, 103: 278, 105: 1186, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 1027, 278, 1188, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 1185, 1023, 1024, 1031, 1028, 1029, 1187, 1189, 1190},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 964, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1184},
		{265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 103: 265, 105: 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265, 265},
		// 45
		{264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 103: 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264, 264},
		{263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 103: 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263, 263},
		{262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 103: 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262, 262},
		{261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 103: 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261, 261},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 1165, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 1183, 642, 638},
		// 50
		{102: 1181, 194: 1182},
		{258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 103: 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258, 258},
		{257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 103: 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257, 257},
		{102: 1178},
		{1: 1155, 338: 1154},
		// 55
		{254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 103: 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254, 254},
		{253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 103: 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253, 253},
		{252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 103: 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252, 252},
		{251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 103: 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251, 251},
		{102: 1151},
		// 60
		{102: 1146},
		{102: 1143},
		{102: 1140},
		{102: 1131},
		{102: 1090},
		// 65
		{199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 895, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 265: 1089},
		{199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 1086, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 265: 1085},
		{102: 1078},
		{102: 1071},
		{102: 1066},
		// 70
		{102: 1047},
		{102: 957},
		{12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 948, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12},
		{11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 922, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11},
		{102: 908},
		// 75
		{102: 898},
		{199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 895, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 265: 897},
		{199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 895, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 199, 265: 894},
		{78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 888, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78},
		{102: 886},
		// 80
		{77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 883, 77, 77, 77, 77, 77, 77, 77, 77, 77, 882, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77},
		{75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 879, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75},
		{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 876, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61},
		{102: 873},
		{54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 870, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54},
		// 85
		{53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 867, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53},
		{102: 864},
		{52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 861, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52},
		{36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 858, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36},
		{102: 856},
		// 90
		{31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 853, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31},
		{13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 850, 13, 13, 13, 13, 13, 13, 13, 13, 13, 849, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13},
		{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 846, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		{6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 844, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 841, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		// 95
		{91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 816, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91},
		{1: 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 90: 136, 136, 136, 94: 136, 136, 136, 102: 136, 104: 136, 166: 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136, 136},
		{1: 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 90: 135, 135, 135, 94: 135, 135, 135, 102: 135, 104: 135, 166: 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135, 135},
		{1: 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 90: 134, 134, 134, 94: 134, 134, 134, 102: 134, 104: 134, 166: 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134, 134},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 139: 132, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 797, 335: 798},
		// 100
		{125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 103: 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 771, 212: 125, 125, 220: 125, 125, 223: 125, 125, 125, 227: 125, 125, 125},
		{122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 103: 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122, 122},
		{121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 103: 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 121, 48},
		{120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 103: 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120, 120},
		{14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 770, 14, 14, 14, 14, 14, 14, 14, 14, 14, 103: 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14},
		// 105
		{115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 103: 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115, 115},
		{114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 103: 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114, 114},
		{113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 103: 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113, 113},
		{112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 103: 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112, 112},
		{111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 103: 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 7},
		// 110
		{90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 209: 90, 90, 90, 90, 90, 90, 90, 90, 90, 219: 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90, 90},
		{87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 209: 87, 87, 87, 87, 87, 87, 87, 87, 87, 219: 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87},
		{86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 209: 86, 86, 86, 86, 86, 86, 86, 86, 86, 219: 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86, 86},
		{85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 209: 85, 85, 85, 85, 85, 85, 85, 85, 85, 219: 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85, 85},
		{84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 209: 84, 84, 84, 84, 84, 84, 84, 84, 84, 219: 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84, 84},
		// 115
		{83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 209: 83, 83, 83, 83, 83, 83, 83, 83, 83, 219: 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83, 83},
		{82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 209: 82, 82, 82, 82, 82, 82, 82, 82, 82, 219: 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82, 82},
		{81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 209: 81, 81, 81, 81, 81, 81, 81, 81, 81, 219: 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81, 81},
		{80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 209: 80, 80, 80, 80, 80, 80, 80, 80, 80, 219: 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80, 80},
		{76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 209: 76, 76, 76, 76, 76, 76, 76, 76, 76, 219: 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76, 76},
		// 120
		{74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 209: 74, 74, 74, 74, 74, 74, 74, 74, 74, 219: 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74, 74},
		{73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 209: 73, 73, 73, 73, 73, 73, 73, 73, 73, 219: 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73, 73},
		{72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 209: 72, 72, 72, 72, 72, 72, 72, 72, 72, 219: 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72, 72},
		{71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 209: 71, 71, 71, 71, 71, 71, 71, 71, 71, 219: 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71, 71},
		{70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 209: 70, 70, 70, 70, 70, 70, 70, 70, 70, 219: 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70},
		// 125
		{69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 209: 69, 69, 69, 69, 69, 69, 69, 69, 69, 219: 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69, 69},
		{68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 209: 68, 68, 68, 68, 68, 68, 68, 68, 68, 219: 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68},
		{67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 209: 67, 67, 67, 67, 67, 67, 67, 67, 67, 219: 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67},
		{66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 209: 66, 66, 66, 66, 66, 66, 66, 66, 66, 219: 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66},
		{65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 209: 65, 65, 65, 65, 65, 65, 65, 65, 65, 219: 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65},
		// 130
		{64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 209: 64, 64, 64, 64, 64, 64, 64, 64, 64, 219: 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64},
		{63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 209: 63, 63, 63, 63, 63, 63, 63, 63, 63, 219: 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63, 63},
		{62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 209: 62, 62, 62, 62, 62, 62, 62, 62, 62, 219: 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62},
		{60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 209: 60, 60, 60, 60, 60, 60, 60, 60, 60, 219: 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60},
		{59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 209: 59, 59, 59, 59, 59, 59, 59, 59, 59, 219: 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59, 59},
		// 135
		{58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 209: 58, 58, 58, 58, 58, 58, 58, 58, 58, 219: 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58, 58},
		{57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 209: 57, 57, 57, 57, 57, 57, 57, 57, 57, 219: 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57, 57},
		{56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 209: 56, 56, 56, 56, 56, 56, 56, 56, 56, 219: 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56, 56},
		{55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 209: 55, 55, 55, 55, 55, 55, 55, 55, 55, 219: 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55, 55},
		{51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 209: 51, 51, 51, 51, 51, 51, 51, 51, 51, 219: 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51},
		// 140
		{50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 209: 50, 50, 50, 50, 50, 50, 50, 50, 50, 219: 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50},
		{49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 209: 49, 49, 49, 49, 49, 49, 49, 49, 49, 219: 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49, 49},
		{47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 209: 47, 47, 47, 47, 47, 47, 47, 47, 47, 219: 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47, 47},
		{46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 209: 46, 46, 46, 46, 46, 46, 46, 46, 46, 219: 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46, 46},
		{45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 209: 45, 45, 45, 45, 45, 45, 45, 45, 45, 219: 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45, 45},
		// 145
		{44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 209: 44, 44, 44, 44, 44, 44, 44, 44, 44, 219: 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44, 44},
		{43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 209: 43, 43, 43, 43, 43, 43, 43, 43, 43, 219: 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43},
		{42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 209: 42, 42, 42, 42, 42, 42, 42, 42, 42, 219: 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42},
		{41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 209: 41, 41, 41, 41, 41, 41, 41, 41, 41, 219: 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41, 41},
		{40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 209: 40, 40, 40, 40, 40, 40, 40, 40, 40, 219: 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40, 40},
		// 150
		{39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 209: 39, 39, 39, 39, 39, 39, 39, 39, 39, 219: 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39},
		{38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 209: 38, 38, 38, 38, 38, 38, 38, 38, 38, 219: 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38, 38},
		{37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 209: 37, 37, 37, 37, 37, 37, 37, 37, 37, 219: 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37, 37},
		{34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 209: 34, 34, 34, 34, 34, 34, 34, 34, 34, 219: 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34},
		{33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 209: 33, 33, 33, 33, 33, 33, 33, 33, 33, 219: 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33, 33},
		// 155
		{30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 209: 30, 30, 30, 30, 30, 30, 30, 30, 30, 219: 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30},
		{29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 209: 29, 29, 29, 29, 29, 29, 29, 29, 29, 219: 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29, 29},
		{28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 209: 28, 28, 28, 28, 28, 28, 28, 28, 28, 219: 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28, 28},
		{27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 209: 27, 27, 27, 27, 27, 27, 27, 27, 27, 219: 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27},
		{26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 209: 26, 26, 26, 26, 26, 26, 26, 26, 26, 219: 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26},
		// 160
		{25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 209: 25, 25, 25, 25, 25, 25, 25, 25, 25, 219: 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25},
		{24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 209: 24, 24, 24, 24, 24, 24, 24, 24, 24, 219: 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24},
		{23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 209: 23, 23, 23, 23, 23, 23, 23, 23, 23, 219: 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23},
		{22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 209: 22, 22, 22, 22, 22, 22, 22, 22, 22, 219: 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22},
		{21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 209: 21, 21, 21, 21, 21, 21, 21, 21, 21, 219: 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21},
		// 165
		{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 209: 20, 20, 20, 20, 20, 20, 20, 20, 20, 219: 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
		{19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 209: 19, 19, 19, 19, 19, 19, 19, 19, 19, 219: 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19},
		{18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 209: 18, 18, 18, 18, 18, 18, 18, 18, 18, 219: 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18},
		{17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 209: 17, 17, 17, 17, 17, 17, 17, 17, 17, 219: 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17},
		{16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 209: 16, 16, 16, 16, 16, 16, 16, 16, 16, 219: 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16},
		// 170
		{15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 209: 15, 15, 15, 15, 15, 15, 15, 15, 15, 219: 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15},
		{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 209: 10, 10, 10, 10, 10, 10, 10, 10, 10, 219: 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
		{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 209: 9, 9, 9, 9, 9, 9, 9, 9, 9, 219: 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
		{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 209: 8, 8, 8, 8, 8, 8, 8, 8, 8, 219: 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8},
		{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 209: 5, 5, 5, 5, 5, 5, 5, 5, 5, 219: 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
		// 175
		{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 209: 4, 4, 4, 4, 4, 4, 4, 4, 4, 219: 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4},
		{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 209: 3, 3, 3, 3, 3, 3, 3, 3, 3, 219: 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
		{118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 103: 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118, 118},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 772},
		{124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 103: 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 795, 212: 124, 124, 220: 124, 124, 223: 124, 124, 124, 227: 124, 124, 124},
		// 180
		{91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 209: 91, 91, 91, 91, 91, 91, 91, 91, 91, 219: 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91, 91},
		{79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 209: 79, 79, 79, 79, 79, 79, 79, 79, 79, 219: 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79, 79},
		{78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 209: 78, 78, 78, 78, 78, 78, 78, 78, 78, 219: 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78, 78},
		{77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 209: 77, 77, 77, 77, 77, 77, 77, 77, 77, 219: 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77, 77},
		{75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 209: 75, 75, 75, 75, 75, 75, 75, 75, 75, 219: 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75},
		// 185
		{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 209: 61, 61, 61, 61, 61, 61, 61, 61, 61, 219: 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61},
		{54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 209: 54, 54, 54, 54, 54, 54, 54, 54, 54, 219: 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54},
		{53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 209: 53, 53, 53, 53, 53, 53, 53, 53, 53, 219: 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53},
		{52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 209: 52, 52, 52, 52, 52, 52, 52, 52, 52, 219: 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52},
		{48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 209: 48, 48, 48, 48, 48, 48, 48, 48, 48, 219: 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48},
		// 190
		{36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 209: 36, 36, 36, 36, 36, 36, 36, 36, 36, 219: 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36},
		{35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 209: 35, 35, 35, 35, 35, 35, 35, 35, 35, 219: 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35, 35},
		{32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 209: 32, 32, 32, 32, 32, 32, 32, 32, 32, 219: 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32, 32},
		{31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 209: 31, 31, 31, 31, 31, 31, 31, 31, 31, 219: 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31},
		{14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 209: 14, 14, 14, 14, 14, 14, 14, 14, 14, 219: 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14, 14},
		// 195
		{13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 209: 13, 13, 13, 13, 13, 13, 13, 13, 13, 219: 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13},
		{12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 209: 12, 12, 12, 12, 12, 12, 12, 12, 12, 219: 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12, 12},
		{11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 209: 11, 11, 11, 11, 11, 11, 11, 11, 11, 219: 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11},
		{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 209: 7, 7, 7, 7, 7, 7, 7, 7, 7, 219: 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7},
		{6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 209: 6, 6, 6, 6, 6, 6, 6, 6, 6, 219: 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6},
		// 200
		{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 209: 2, 2, 2, 2, 2, 2, 2, 2, 2, 219: 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 209: 1, 1, 1, 1, 1, 1, 1, 1, 1, 219: 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 796},
		{123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 103: 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 123, 212: 123, 123, 220: 123, 123, 223: 123, 123, 123, 227: 123, 123, 123},
		{122: 805, 126: 803, 128: 804, 139: 131},
		// 205
		{139: 801, 308: 800, 374: 799},
		{139: 801, 127, 813, 308: 812, 327: 811},
		{139: 130, 130, 130},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 802},
		{122: 805, 126: 803, 128: 804, 144: 806},
		// 210
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 810},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 809},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 808},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 807},
		{122: 805, 126: 803, 128: 804, 139: 128, 128, 128},
		// 215
		{314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 97: 314, 314, 314, 314, 314, 103: 314, 106: 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 128: 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314, 314},
		{315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 97: 315, 315, 315, 315, 315, 103: 315, 106: 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 805, 315, 315, 315, 315, 128: 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315, 315},
		{316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 97: 316, 316, 316, 316, 316, 103: 316, 106: 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 805, 316, 316, 316, 316, 128: 804, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316, 316},
		{140: 815},
		{139: 129, 129, 129},
		// 220
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 814},
		{122: 805, 126: 803, 128: 804, 140: 126},
		{133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 103: 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133, 133},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 822, 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 818, 242: 823},
		{386, 89: 386, 93: 386, 97: 386, 386, 386, 386, 103: 386, 109: 386, 124: 386, 136: 386},
		// 225
		{89: 840, 93: 825},
		{384, 2: 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 89: 384, 93: 384, 97: 384, 384, 384, 384, 384, 103: 384, 109: 384, 111: 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 384, 124: 384, 136: 384},
		{366, 773, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 366, 92: 836, 366, 97: 366, 366, 366, 366, 366, 103: 366, 107: 366, 109: 366, 111: 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 805, 124: 366, 126: 803, 128: 804, 136: 366, 138: 835, 164: 703, 834, 262: 833},
		{125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 92: 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 103: 125, 105: 125, 107: 125, 109: 125, 111: 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 124: 125, 126: 125, 125, 125, 125, 136: 125, 138: 125, 145: 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 125, 163: 828},
		{202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 103: 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202, 202},
		// 230
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 824},
		{89: 826, 93: 825},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 827},
		{200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 103: 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 200},
		{385, 89: 385, 93: 385, 97: 385, 385, 385, 385, 103: 385, 109: 385, 124: 385, 136: 385},
		// 235
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 127: 829, 164: 703, 830},
		{381, 2: 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 89: 381, 93: 381, 97: 381, 381, 381, 381, 381, 103: 381, 109: 381, 111: 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 381, 124: 381, 136: 381},
		{124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 92: 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 103: 124, 105: 124, 107: 124, 109: 124, 111: 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124: 124, 126: 124, 124, 124, 124, 136: 124, 138: 124, 145: 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 124, 163: 831},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 127: 832, 164: 703, 796},
		{380, 2: 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 89: 380, 93: 380, 97: 380, 380, 380, 380, 380, 103: 380, 109: 380, 111: 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 380, 124: 380, 136: 380},
		// 240
		{383, 2: 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 89: 383, 93: 383, 97: 383, 383, 383, 383, 383, 103: 383, 107: 839, 109: 383, 111: 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 383, 124: 383, 136: 383},
		{365, 2: 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 89: 365, 365, 365, 93: 365, 97: 365, 365, 365, 365, 365, 103: 365, 106: 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 365, 123: 365, 365, 365, 130: 365, 365, 365, 365, 365, 365, 365, 365, 209: 365, 365, 365},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 92: 838, 164: 703, 837},
		{363, 2: 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 89: 363, 363, 363, 93: 363, 97: 363, 363, 363, 363, 363, 103: 363, 106: 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 363, 123: 363, 363, 363, 130: 363, 363, 363, 363, 363, 363, 363, 363, 209: 363, 363, 363},
		{364, 2: 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 89: 364, 364, 364, 93: 364, 97: 364, 364, 364, 364, 364, 103: 364, 106: 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 364, 123: 364, 364, 364, 130: 364, 364, 364, 364, 364, 364, 364, 364, 209: 364, 364, 364},
		// 245
		{362, 2: 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 89: 362, 362, 362, 93: 362, 97: 362, 362, 362, 362, 362, 103: 362, 106: 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 362, 123: 362, 362, 362, 130: 362, 362, 362, 362, 362, 362, 362, 362, 209: 362, 362, 362},
		{382, 2: 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 89: 382, 93: 382, 97: 382, 382, 382, 382, 382, 103: 382, 109: 382, 111: 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 382, 124: 382, 136: 382},
		{201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 103: 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201, 201},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 842},
		{89: 843, 93: 825},
		// 250
		{203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 103: 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203, 203},
		{89: 845},
		{204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 103: 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204, 204},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 847},
		{89: 848, 93: 825},
		// 255
		{205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 103: 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205, 205},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 851},
		{117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 103: 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117, 117},
		{89: 852, 93: 825},
		{206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 103: 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206, 206},
		// 260
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 854},
		{89: 855, 93: 825},
		{207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 103: 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207, 207},
		{89: 857},
		{208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 103: 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208, 208},
		// 265
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 859},
		{89: 860, 93: 825},
		{209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 103: 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209, 209},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 862},
		{89: 863, 93: 825},
		// 270
		{210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 103: 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210, 210},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 865},
		{89: 866, 93: 825},
		{211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 103: 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211, 211},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 868},
		// 275
		{89: 869, 93: 825},
		{212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 103: 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212, 212},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 871},
		{89: 872, 93: 825},
		{213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 103: 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213, 213},
		// 280
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 874},
		{89: 875, 93: 825},
		{215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 103: 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215, 215},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 877},
		{89: 878, 93: 825},
		// 285
		{216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 103: 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216, 216},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 880},
		{89: 881, 93: 825},
		{217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 103: 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217, 217},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 884},
		// 290
		{119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 103: 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119, 119},
		{89: 885, 93: 825},
		{218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 103: 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218, 218},
		{89: 887},
		{219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 103: 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219, 219},
		// 295
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 889, 242: 890},
		{89: 893, 93: 825},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 891},
		{89: 892, 93: 825},
		{220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 103: 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220, 220},
		// 300
		{221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 103: 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221, 221},
		{222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 103: 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222, 222},
		{89: 896},
		{198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 103: 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198, 198},
		{223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 103: 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223, 223},
		// 305
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 899, 236: 900},
		{89: 386, 93: 386, 124: 902},
		{89: 901, 93: 825},
		{224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 103: 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224, 224},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 903},
		// 310
		{89: 905, 101: 904},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 906},
		{225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 103: 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225, 225},
		{89: 907},
		{226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 103: 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226, 226},
		// 315
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 909, 314: 911, 910, 347: 912, 368: 913},
		{89: 918, 124: 919},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 914},
		{1: 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 90: 194, 194, 194, 94: 194, 194, 194, 102: 194, 104: 194, 194, 127: 194, 166: 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194, 194},
		{1: 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 90: 193, 193, 193, 94: 193, 193, 193, 102: 193, 104: 193, 193, 127: 193, 166: 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193, 193},
		// 320
		{1: 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 90: 192, 192, 192, 94: 192, 192, 192, 102: 192, 104: 192, 192, 127: 192, 166: 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192, 192},
		{124: 915},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 916},
		{89: 917},
		{228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 103: 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228, 228},
		// 325
		{229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 103: 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229, 229},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 920},
		{89: 921},
		{227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 103: 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227, 227},
		{2: 933, 937, 938, 941, 939, 935, 934, 940, 936, 929, 930, 931, 927, 926, 932, 928, 925, 243: 924, 923},
		// 330
		{93: 945},
		{93: 942},
		{188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 103: 188, 105: 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188, 188},
		{187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 103: 187, 105: 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187, 187},
		{186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 103: 186, 105: 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186, 186},
		// 335
		{185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 103: 185, 105: 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185, 185},
		{184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 103: 184, 105: 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184, 184},
		{183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 103: 183, 105: 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183, 183},
		{182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 103: 182, 105: 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182, 182},
		{181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 103: 181, 105: 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181, 181},
		// 340
		{180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 103: 180, 105: 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180, 180},
		{179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 103: 179, 105: 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179, 179},
		{178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 103: 178, 105: 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178, 178},
		{177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 103: 177, 105: 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177, 177},
		{176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 103: 176, 105: 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176, 176},
		// 345
		{175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 103: 175, 105: 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175, 175},
		{174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 103: 174, 105: 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174, 174},
		{173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 103: 173, 105: 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173, 173},
		{172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 103: 172, 105: 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172, 172},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 943},
		// 350
		{89: 944, 93: 825},
		{230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 103: 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230, 230},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 946},
		{89: 947, 93: 825},
		{231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 103: 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231, 231},
		// 355
		{2: 933, 937, 938, 941, 939, 935, 934, 940, 936, 929, 930, 931, 927, 926, 932, 928, 925, 243: 950, 949},
		{93: 954},
		{93: 951},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 952},
		{89: 953, 93: 825},
		// 360
		{232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 103: 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232, 232},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 955},
		{89: 956, 93: 825},
		{233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 103: 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233, 233},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 958},
		// 365
		{93: 959},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 960, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 961},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 964, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 963, 965},
		{89: 962},
		{235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 103: 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235, 235},
		// 370
		{1: 773, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 92: 836, 107: 366, 111: 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 366, 805, 126: 803, 128: 804, 138: 835, 164: 703, 834, 243: 969, 967, 247: 1003, 250: 968, 262: 833},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 984, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 986, 817, 608, 610, 985, 238: 982, 983, 241: 609, 259: 987},
		{2: 933, 937, 938, 941, 939, 935, 934, 940, 936, 929, 930, 931, 927, 926, 932, 928, 925, 111: 979, 976, 978, 977, 973, 975, 974, 971, 972, 970, 980, 243: 969, 967, 247: 966, 250: 968},
		{89: 981},
		{191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 103: 191, 105: 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191, 191},
		// 375
		{190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 103: 190, 105: 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190, 190},
		{189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 103: 189, 105: 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189, 189},
		{171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 103: 171, 105: 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171, 171},
		{170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 103: 170, 105: 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170, 170},
		{169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 103: 169, 105: 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169, 169},
		// 380
		{168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 103: 168, 105: 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168, 168},
		{167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 103: 167, 105: 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167, 167},
		{166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 103: 166, 105: 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166, 166},
		{165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 103: 165, 105: 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165, 165},
		{164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 103: 164, 105: 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164, 164},
		// 385
		{163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 103: 163, 105: 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163, 163},
		{162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 103: 162, 105: 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162, 162},
		{161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 103: 161, 105: 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161, 161},
		{234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 103: 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234, 234},
		{579, 89: 579, 97: 579, 579, 579, 579},
		// 390
		{89: 1002, 97: 998, 999, 997, 996, 245: 994},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 984, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 993, 234: 608, 610, 238: 982, 992, 241: 609, 259: 987},
		{89: 991, 93: 825},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 366, 92: 836, 366, 107: 366, 122: 805, 126: 803, 128: 804, 138: 835, 164: 703, 834, 262: 833},
		{89: 988, 93: 989},
		// 395
		{320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 103: 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320, 320},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 990},
		{317, 89: 317, 93: 317, 97: 317, 317, 317, 317, 317, 103: 317, 108: 317, 317, 122: 805, 125: 317, 803, 128: 804},
		{214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 103: 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214, 214},
		{89: 995, 97: 998, 999, 997, 996, 245: 994},
		// 400
		{318, 89: 318, 93: 318, 97: 318, 318, 318, 318, 318, 103: 318, 108: 318, 318, 122: 805, 125: 318, 803, 128: 804},
		{102: 611, 234: 608, 610, 238: 982, 1001, 241: 609},
		{1: 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 566, 319, 319, 319, 319, 319, 319, 319, 566, 566, 566, 566, 105: 319, 107: 319, 110: 319, 122: 319, 319, 126: 319, 319, 319, 319, 131: 319, 319, 319, 138: 319, 145: 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 209: 319, 319, 319},
		{102: 395, 234: 395, 395, 251: 1000},
		{102: 393, 234: 393, 393},
		// 405
		{102: 392, 234: 392, 392},
		{102: 391, 234: 391, 391},
		{102: 394, 234: 394, 394},
		{573, 89: 573, 97: 573, 573, 573, 573, 245: 994},
		{319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 103: 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 319, 209: 319, 319, 319},
		// 410
		{94: 1021},
		{2: 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 94: 188, 107: 19, 111: 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19},
		{2: 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 94: 187, 107: 22, 111: 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22},
		{2: 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 94: 186, 107: 23, 111: 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23},
		{2: 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 94: 185, 107: 20, 111: 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
		// 415
		{2: 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 94: 184, 107: 26, 111: 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26},
		{2: 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 94: 183, 107: 25, 111: 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25},
		{2: 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 94: 182, 107: 24, 111: 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24},
		{2: 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 94: 181, 107: 21, 111: 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21},
		{2: 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 94: 180, 107: 1, 111: 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		// 420
		{2: 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 94: 179, 107: 36, 111: 36, 36, 36, 36, 36, 36, 36, 36, 36, 36, 36},
		{2: 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 94: 178, 107: 52, 111: 52, 52, 52, 52, 52, 52, 52, 52, 52, 52, 52},
		{2: 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 94: 177, 107: 2, 111: 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
		{2: 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 94: 176, 107: 75, 111: 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75},
		{2: 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 94: 175, 107: 61, 111: 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61},
		// 425
		{2: 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 94: 174, 107: 53, 111: 53, 53, 53, 53, 53, 53, 53, 53, 53, 53, 53},
		{2: 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 94: 173, 107: 31, 111: 31, 31, 31, 31, 31, 31, 31, 31, 31, 31, 31},
		{2: 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 94: 172, 107: 54, 111: 54, 54, 54, 54, 54, 54, 54, 54, 54, 54, 54},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1022},
		{273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 103: 273, 105: 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273, 273},
		// 430
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1046},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1045},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 1042, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1041},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 1038, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1037},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1036},
		// 435
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1035},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1034},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1033},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1032},
		{266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 1025, 1030, 1026, 266, 266, 266, 266, 266, 103: 266, 105: 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 1027, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 266, 1028, 1029, 266, 266, 266},
		// 440
		{267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 103: 267, 105: 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267, 267},
		{268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 103: 268, 105: 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268, 268},
		{269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 103: 269, 105: 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269, 269},
		{270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 103: 270, 105: 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270, 270},
		{272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 1030, 272, 272, 272, 272, 272, 272, 103: 272, 105: 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 1027, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 272, 1028, 1029, 272, 272, 272},
		// 445
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 964, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1039},
		{2: 933, 937, 938, 941, 939, 935, 934, 940, 936, 929, 930, 931, 927, 926, 932, 928, 925, 111: 979, 976, 978, 977, 973, 975, 974, 971, 972, 970, 980, 805, 126: 803, 128: 804, 243: 969, 967, 247: 1040, 250: 968},
		{271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 103: 271, 105: 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271, 271},
		{275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 1030, 275, 275, 275, 275, 275, 275, 103: 275, 105: 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 1027, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 275, 1028, 1029, 275, 275, 275},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 964, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1043},
		// 450
		{2: 933, 937, 938, 941, 939, 935, 934, 940, 936, 929, 930, 931, 927, 926, 932, 928, 925, 111: 979, 976, 978, 977, 973, 975, 974, 971, 972, 970, 980, 805, 126: 803, 128: 804, 243: 969, 967, 247: 1044, 250: 968},
		{274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 103: 274, 105: 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274, 274},
		{276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 1025, 1030, 1026, 276, 276, 276, 276, 276, 103: 276, 105: 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 1027, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 276, 1028, 1029, 276, 276, 276},
		{277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 1025, 1030, 1026, 277, 277, 277, 277, 277, 103: 277, 105: 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 1027, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 277, 1028, 1029, 277, 277, 277},
		{1: 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 90: 390, 390, 390, 94: 390, 390, 390, 102: 390, 104: 390, 390, 127: 390, 166: 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 390, 242: 1048, 325: 1049},
		// 455
		{1: 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 90: 389, 389, 389, 94: 389, 389, 389, 102: 389, 104: 389, 389, 127: 389, 166: 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389, 389},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 1050},
		{89: 106, 93: 825, 109: 1052, 136: 106, 266: 1051},
		{89: 388, 136: 1062, 356: 1063},
		{280: 1053},
		// 460
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1054, 293: 1056, 350: 1055},
		{101, 89: 101, 93: 101, 97: 101, 101, 101, 101, 101, 103: 101, 108: 101, 122: 805, 126: 803, 128: 804, 136: 101, 142: 1061, 1060, 313: 1059},
		{105, 89: 105, 93: 1057, 97: 105, 105, 105, 105, 105, 103: 105, 108: 105, 136: 105},
		{104, 89: 104, 93: 104, 97: 104, 104, 104, 104, 104, 103: 104, 108: 104, 136: 104},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1054, 293: 1058},
		// 465
		{103, 89: 103, 93: 103, 97: 103, 103, 103, 103, 103, 103: 103, 108: 103, 136: 103},
		{102, 89: 102, 93: 102, 97: 102, 102, 102, 102, 102, 103: 102, 108: 102, 136: 102},
		{100, 89: 100, 93: 100, 97: 100, 100, 100, 100, 100, 103: 100, 108: 100, 136: 100},
		{99, 89: 99, 93: 99, 97: 99, 99, 99, 99, 99, 103: 99, 108: 99, 136: 99},
		{92: 1065},
		// 470
		{89: 1064},
		{236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 103: 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236, 236},
		{89: 387},
		{2: 933, 937, 938, 941, 939, 935, 934, 940, 936, 929, 930, 931, 927, 926, 932, 928, 925, 111: 979, 976, 978, 977, 973, 975, 974, 971, 972, 970, 980, 243: 969, 967, 247: 1067, 250: 968},
		{124: 1068},
		// 475
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 1069},
		{89: 1070},
		{237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 103: 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237, 237},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 1072},
		{93: 1073},
		// 480
		{168: 1074},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 1075},
		{2: 933, 937, 938, 941, 939, 935, 934, 940, 936, 929, 930, 931, 927, 926, 932, 928, 925, 111: 979, 976, 978, 977, 973, 975, 974, 971, 972, 970, 980, 243: 969, 967, 247: 1076, 250: 968},
		{89: 1077},
		{238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 103: 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238, 238},
		// 485
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 1079},
		{93: 1080},
		{168: 1081},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 1082},
		{2: 933, 937, 938, 941, 939, 935, 934, 940, 936, 929, 930, 931, 927, 926, 932, 928, 925, 111: 979, 976, 978, 977, 973, 975, 974, 971, 972, 970, 980, 243: 969, 967, 247: 1083, 250: 968},
		// 490
		{89: 1084},
		{239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 103: 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239, 239},
		{241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 103: 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241, 241},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 896, 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 1087},
		{89: 1088},
		// 495
		{240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 103: 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240, 240},
		{242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 103: 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242, 242},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1091},
		{122: 805, 126: 803, 128: 804, 138: 1092},
		{21: 1096, 25: 1097, 1098, 29: 1100, 32: 1104, 38: 1099, 1102, 41: 1103, 167: 1095, 212: 1094, 1101, 298: 1106, 1108, 1110, 1109, 1093, 1107, 307: 1105},
		// 500
		{89: 1128, 107: 1129},
		{89: 160, 102: 1125, 107: 160},
		{89: 158, 102: 1122, 107: 158},
		{89: 156, 107: 156},
		{89: 155, 107: 155},
		// 505
		{89: 154, 102: 1116, 107: 154},
		{89: 151, 102: 1113, 107: 151},
		{89: 149, 107: 149},
		{89: 148, 107: 148},
		{89: 147, 107: 147},
		// 510
		{89: 146, 107: 146, 213: 1112},
		{89: 144, 107: 144},
		{89: 143, 107: 143, 213: 1111},
		{89: 141, 107: 141},
		{89: 140, 107: 140},
		// 515
		{89: 139, 107: 139},
		{89: 138, 107: 138},
		{89: 137, 107: 137},
		{89: 142, 107: 142},
		{89: 145, 107: 145},
		// 520
		{20: 1114},
		{89: 1115},
		{89: 150, 107: 150},
		{20: 1117},
		{89: 1118, 93: 1119},
		// 525
		{89: 153, 107: 153},
		{20: 1120},
		{89: 1121},
		{89: 152, 107: 152},
		{20: 1123},
		// 530
		{89: 1124},
		{89: 157, 107: 157},
		{20: 1126},
		{89: 1127},
		{89: 159, 107: 159},
		// 535
		{244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 103: 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244, 244},
		{89: 1130},
		{243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 103: 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243, 243},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 1132},
		{93: 1133},
		// 540
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 1134, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 1135},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 964, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 963, 1137},
		{89: 1136},
		{245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 103: 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245, 245},
		{2: 933, 937, 938, 941, 939, 935, 934, 940, 936, 929, 930, 931, 927, 926, 932, 928, 925, 111: 979, 976, 978, 977, 973, 975, 974, 971, 972, 970, 980, 243: 969, 967, 247: 1138, 250: 968},
		// 545
		{89: 1139},
		{246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 103: 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246, 246},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 1141},
		{89: 1142, 93: 825},
		{247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 103: 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247},
		// 550
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 1144},
		{89: 1145, 93: 825},
		{248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 103: 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248, 248},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1147},
		{93: 1148, 122: 805, 126: 803, 128: 804},
		// 555
		{21: 1096, 25: 1097, 1098, 29: 1100, 32: 1104, 38: 1099, 1102, 41: 1103, 167: 1095, 212: 1094, 1101, 298: 1106, 1108, 1110, 1109, 1149, 1107, 307: 1105},
		{89: 1150},
		{249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 103: 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249, 249},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 1152},
		{89: 1153, 93: 825},
		// 560
		{250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 103: 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250, 250},
		{1: 1175, 1174, 1163, 1164, 1166, 1167, 1168, 1169, 1170, 1172, 21: 1162, 34: 1161, 42: 1171, 44: 1173, 76: 1159, 1160, 90: 654, 655, 95: 679, 166: 676, 652, 1165, 672, 173: 682, 175: 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 198: 1158, 651, 650, 648, 649, 668},
		{92: 1156},
		{134: 1157},
		{116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 103: 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116, 116},
		// 565
		{134: 1177},
		{102: 948},
		{102: 922},
		{102: 888},
		{102: 882},
		// 570
		{102: 879},
		{102: 876},
		{102: 1176},
		{102: 870},
		{102: 867},
		// 575
		{102: 861},
		{102: 858},
		{102: 853},
		{102: 849},
		{102: 846},
		// 580
		{102: 844},
		{102: 841},
		{102: 816},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 985},
		{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 103: 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		// 585
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 1179},
		{89: 1180, 93: 825},
		{256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 103: 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256, 256},
		{102: 611, 234: 608, 610, 238: 982, 983, 241: 609},
		{259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 103: 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259, 259},
		// 590
		{260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 103: 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260, 260},
		{2: 933, 937, 938, 941, 939, 935, 934, 940, 936, 929, 930, 931, 927, 926, 932, 928, 925, 111: 979, 976, 978, 977, 973, 975, 974, 971, 972, 970, 980, 805, 126: 803, 128: 804, 243: 969, 967, 247: 1003, 250: 968},
		{40: 1222, 102: 631, 194: 1230, 197: 1231},
		{129: 1209, 153: 1207, 159: 1208, 1210, 1211},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1204},
		// 595
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 1165, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 1193, 642, 638, 212: 1194},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1192},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1191},
		{280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 1025, 1030, 1026, 280, 280, 280, 280, 280, 103: 280, 106: 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 1027, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 280, 154: 1023, 1024, 1031, 1028, 1029},
		{282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 1025, 1030, 1026, 282, 282, 282, 282, 282, 103: 282, 106: 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 1027, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 282, 154: 1023, 1024, 1031, 1028, 1029},
		// 600
		{197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 97: 197, 197, 197, 197, 197, 103: 197, 1198, 106: 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 128: 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 162: 1197, 264: 1203},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 1165, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 1195, 642, 638},
		{197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 97: 197, 197, 197, 197, 197, 103: 197, 1198, 106: 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 128: 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 162: 1197, 264: 1196},
		{284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 97: 284, 284, 284, 284, 284, 103: 284, 106: 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 128: 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284, 284},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 1165, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 1202, 642, 638},
		// 605
		{162: 1199},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 1165, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 1200, 642, 638},
		{134: 1201},
		{195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 97: 195, 195, 195, 195, 195, 103: 195, 106: 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 128: 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195, 195},
		{196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 97: 196, 196, 196, 196, 196, 103: 196, 106: 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 128: 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196, 196},
		// 610
		{286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 97: 286, 286, 286, 286, 286, 103: 286, 106: 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 128: 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286, 286},
		{94: 1025, 1030, 1026, 122: 1205, 127: 1027, 154: 1023, 1024, 1031, 1028, 1029},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 1206},
		{288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 97: 288, 288, 288, 288, 288, 103: 288, 106: 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 128: 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288, 288},
		{40: 1222, 102: 631, 194: 1223, 197: 1224},
		// 615
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1219},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 1165, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 1214, 642, 638, 212: 1215},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1213},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 1212},
		{279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 1025, 1030, 1026, 279, 279, 279, 279, 279, 103: 279, 106: 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 1027, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 279, 154: 1023, 1024, 1031, 1028, 1029},
		// 620
		{281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 1025, 1030, 1026, 281, 281, 281, 281, 281, 103: 281, 106: 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 1027, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 281, 154: 1023, 1024, 1031, 1028, 1029},
		{197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 97: 197, 197, 197, 197, 197, 103: 197, 1198, 106: 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 128: 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 162: 1197, 264: 1218},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 1165, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 1216, 642, 638},
		{197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 97: 197, 197, 197, 197, 197, 103: 197, 1198, 106: 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 128: 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 197, 162: 1197, 264: 1217},
		{283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 97: 283, 283, 283, 283, 283, 103: 283, 106: 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 128: 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283, 283},
		// 625
		{285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 97: 285, 285, 285, 285, 285, 103: 285, 106: 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 128: 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285, 285},
		{94: 1025, 1030, 1026, 122: 1220, 127: 1027, 154: 1023, 1024, 1031, 1028, 1029},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 1221},
		{287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 97: 287, 287, 287, 287, 287, 103: 287, 106: 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 128: 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287, 287},
		{102: 1225},
		// 630
		{290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 97: 290, 290, 290, 290, 290, 103: 290, 106: 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 128: 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290, 290},
		{289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 97: 289, 289, 289, 289, 289, 103: 289, 106: 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 128: 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289, 289},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 993, 259: 1226},
		{93: 1227},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1228},
		// 635
		{89: 1229, 93: 317, 122: 805, 126: 803, 128: 804},
		{321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 103: 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321, 321},
		{292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 97: 292, 292, 292, 292, 292, 103: 292, 106: 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 128: 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292, 292},
		{291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 97: 291, 291, 291, 291, 291, 103: 291, 106: 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 128: 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291, 291},
		{43: 1267, 105: 1265, 171: 701, 1266, 174: 700, 196: 1264},
		// 640
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 1242, 719, 721, 733, 739, 784, 1241, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 1261, 251: 1240, 255: 1262},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 1242, 719, 721, 733, 739, 784, 1241, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 1258, 251: 1240, 255: 1259},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 1242, 719, 721, 733, 739, 784, 1241, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 1255, 251: 1240, 255: 1256},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 1242, 719, 721, 733, 739, 784, 1241, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 1252, 251: 1240, 255: 1253},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 1242, 719, 721, 733, 739, 784, 1241, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 1249, 251: 1240, 255: 1250},
		// 645
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 1242, 719, 721, 733, 739, 784, 1241, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 1246, 251: 1240, 255: 1247},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 1242, 719, 721, 733, 739, 784, 1241, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 1243, 251: 1240, 255: 1244},
		{102: 324},
		{27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 323, 27, 105: 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 27, 163: 27},
		{87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 322, 87, 105: 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 87, 163: 87},
		// 650
		{295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 97: 295, 295, 295, 295, 295, 103: 295, 106: 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 128: 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295, 295},
		{102: 1181, 194: 1245},
		{294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 97: 294, 294, 294, 294, 294, 103: 294, 106: 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 128: 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294, 294},
		{297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 97: 297, 297, 297, 297, 297, 103: 297, 106: 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 128: 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297, 297},
		{102: 1181, 194: 1248},
		// 655
		{296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 97: 296, 296, 296, 296, 296, 103: 296, 106: 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 128: 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296, 296},
		{299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 97: 299, 299, 299, 299, 299, 103: 299, 106: 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 128: 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299, 299},
		{102: 1181, 194: 1251},
		{298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 97: 298, 298, 298, 298, 298, 103: 298, 106: 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 128: 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298, 298},
		{301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 97: 301, 301, 301, 301, 301, 103: 301, 106: 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 128: 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301, 301},
		// 660
		{102: 1181, 194: 1254},
		{300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 97: 300, 300, 300, 300, 300, 103: 300, 106: 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 128: 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300, 300},
		{303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 97: 303, 303, 303, 303, 303, 103: 303, 106: 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 128: 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303, 303},
		{102: 1181, 194: 1257},
		{302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 97: 302, 302, 302, 302, 302, 103: 302, 106: 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 128: 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302, 302},
		// 665
		{305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 97: 305, 305, 305, 305, 305, 103: 305, 106: 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 128: 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305, 305},
		{102: 1181, 194: 1260},
		{304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 97: 304, 304, 304, 304, 304, 103: 304, 106: 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 128: 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304, 304},
		{307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 97: 307, 307, 307, 307, 307, 103: 307, 106: 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 128: 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307, 307},
		{102: 1181, 194: 1263},
		// 670
		{306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 97: 306, 306, 306, 306, 306, 103: 306, 106: 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 128: 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306, 306},
		{312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 97: 312, 312, 312, 312, 312, 103: 312, 106: 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 128: 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312, 312},
		{43: 1267, 171: 701, 1269, 174: 700, 196: 1268},
		{309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 97: 309, 309, 309, 309, 309, 103: 309, 106: 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 128: 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309, 309},
		{111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 97: 111, 111, 111, 111, 111, 103: 111, 106: 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 128: 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111, 111},
		// 675
		{311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 97: 311, 311, 311, 311, 311, 103: 311, 106: 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 128: 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311, 311},
		{308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 97: 308, 308, 308, 308, 308, 103: 308, 106: 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 128: 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308, 308},
		{313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 97: 313, 313, 313, 313, 313, 103: 313, 106: 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 128: 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313, 313},
		{405, 122: 805, 126: 803, 128: 804},
		{68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 92: 68, 101: 425, 425, 163: 68, 234: 425, 425},
		// 680
		{43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 43, 92: 43, 101: 424, 424, 163: 43, 234: 424, 424},
		{66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 66, 92: 66, 145: 1295, 163: 66},
		{416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 416, 97: 416, 416, 416, 416, 416, 103: 416, 106: 416, 108: 416, 416, 416, 123: 416, 125: 416, 130: 416, 416, 416, 416, 416, 416, 137: 416, 416, 163: 1293, 209: 416, 416, 416, 214: 416, 416, 416, 416, 219: 416, 222: 416, 226: 416},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 1292},
		{409, 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 92: 1291, 164: 703, 1290, 331: 1289},
		// 685
		{101: 1282, 1280, 234: 608, 610, 238: 982, 1279, 241: 609, 286: 1281},
		{421, 97: 998, 999, 997, 996, 245: 994},
		{102: 1280, 234: 608, 610, 238: 982, 1285, 241: 609, 286: 1286},
		{411},
		{51: 1283},
		// 690
		{20: 1284},
		{410},
		{89: 1288, 97: 998, 999, 997, 996, 245: 994},
		{89: 1287},
		{420, 89: 420},
		// 695
		{566, 89: 566, 97: 566, 566, 566, 566, 103: 566, 109: 566},
		{412},
		{408},
		{407},
		{415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 415, 97: 415, 415, 415, 415, 415, 103: 415, 106: 415, 108: 415, 415, 415, 123: 415, 125: 415, 130: 415, 415, 415, 415, 415, 415, 137: 415, 415, 209: 415, 415, 415, 214: 415, 415, 415, 415, 219: 415, 222: 415, 226: 415},
		// 700
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 1294},
		{414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 414, 97: 414, 414, 414, 414, 414, 103: 414, 106: 414, 108: 414, 414, 414, 123: 414, 125: 414, 130: 414, 414, 414, 414, 414, 414, 137: 414, 414, 209: 414, 414, 414, 214: 414, 414, 414, 414, 219: 414, 222: 414, 226: 414},
		{58: 1297, 341: 1298, 367: 1296},
		{101: 427, 427, 234: 427, 427},
		{101: 426, 426, 234: 426, 426},
		// 705
		{101: 423, 423, 234: 423, 423},
		{22: 487, 24: 487, 31: 487},
		{19: 483, 23: 483},
		{19: 482, 23: 482},
		{30: 1464},
		// 710
		{19: 1463, 30: 1462},
		{28: 1458},
		{36: 1434, 44: 1438, 53: 1433, 81: 1439, 169: 1431, 173: 1432, 256: 1436, 294: 1435, 371: 1437},
		{1: 1428},
		{27: 1427},
		// 715
		{98, 103: 1374, 248: 1426},
		{102: 1421},
		{490, 106: 490, 124: 1353, 129: 490, 153: 1352, 246: 1354, 249: 1364, 260: 1419},
		{19: 1416, 33: 1415},
		{328, 101: 1413, 340: 1412},
		// 720
		{124: 1353, 153: 1352, 246: 1354, 249: 1410},
		{124: 1353, 153: 1352, 246: 1354, 249: 1408},
		{124: 1353, 153: 1352, 246: 1354, 249: 1404},
		{22: 1401},
		{451},
		// 725
		{450},
		{19: 1398, 33: 1397},
		{447},
		{446},
		{28: 1391},
		// 730
		{19: 1386, 56: 1385},
		{19: 1382},
		{490, 106: 490, 124: 1353, 129: 490, 153: 1352, 246: 1354, 249: 1364, 260: 1380},
		{98, 103: 1374, 248: 1373},
		{339, 106: 1337, 129: 1338, 237: 1372},
		// 735
		{339, 106: 1337, 129: 1338, 237: 1371},
		{19: 1368, 23: 1367},
		{22: 1349, 24: 1350, 31: 1351},
		{1: 1344},
		{276: 1342},
		// 740
		{339, 106: 1337, 129: 1338, 237: 1341},
		{339, 106: 1337, 129: 1338, 237: 1336},
		{27: 325},
		{428},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1340},
		// 745
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1339},
		{337, 122: 805, 126: 803, 128: 804},
		{338, 122: 805, 126: 803, 128: 804},
		{429},
		{339, 106: 1337, 129: 1338, 237: 1343},
		// 750
		{430},
		{332, 106: 332, 124: 1346, 129: 332, 274: 1345},
		{339, 106: 1337, 129: 1338, 237: 1348},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1347},
		{331, 103: 331, 106: 331, 122: 805, 126: 803, 128: 804, 331},
		// 755
		{434},
		{490, 106: 490, 124: 1353, 129: 490, 153: 1352, 246: 1354, 249: 1364, 260: 1365},
		{124: 1353, 153: 1352, 246: 1354, 249: 1355},
		{432},
		{1: 496, 163: 496},
		// 760
		{1: 495, 163: 495},
		{1: 1358, 163: 1357},
		{339, 106: 1337, 129: 1338, 237: 1356},
		{433},
		{1: 1363},
		// 765
		{491, 106: 491, 124: 1353, 129: 491, 153: 1352, 163: 1359, 246: 1360},
		{1: 1362},
		{1: 1361},
		{492, 106: 492, 129: 492},
		{493, 106: 493, 129: 493},
		// 770
		{494, 106: 494, 129: 494},
		{489, 106: 489, 129: 489},
		{339, 106: 1337, 129: 1338, 237: 1366},
		{435},
		{339, 106: 1337, 129: 1338, 237: 1370},
		// 775
		{339, 106: 1337, 129: 1338, 237: 1369},
		{431},
		{436},
		{437},
		{438},
		// 780
		{440},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1375},
		{97, 62: 1377, 89: 97, 93: 1376, 97: 97, 97, 97, 97, 97, 108: 97, 122: 805, 126: 803, 128: 804},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1379},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1378},
		// 785
		{95, 89: 95, 97: 95, 95, 95, 95, 95, 108: 95, 122: 805, 126: 803, 128: 804},
		{96, 89: 96, 97: 96, 96, 96, 96, 96, 108: 96, 122: 805, 126: 803, 128: 804},
		{339, 106: 1337, 129: 1338, 237: 1381},
		{441},
		{490, 106: 490, 124: 1353, 129: 490, 153: 1352, 246: 1354, 249: 1364, 260: 1383},
		// 790
		{339, 106: 1337, 129: 1338, 237: 1384},
		{442},
		{444},
		{330, 101: 1388, 339: 1387},
		{443},
		// 795
		{47: 1389},
		{1: 1390},
		{329},
		{336, 103: 336, 124: 336, 153: 1393, 291: 1392},
		{332, 103: 332, 124: 1346, 274: 1395},
		// 800
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1394},
		{335, 103: 335, 122: 805, 124: 335, 126: 803, 128: 804},
		{98, 103: 1374, 248: 1396},
		{445},
		{1: 1400},
		// 805
		{339, 106: 1337, 129: 1338, 237: 1399},
		{448},
		{449},
		{490, 106: 490, 124: 1353, 129: 490, 153: 1352, 246: 1354, 249: 1364, 260: 1402},
		{339, 106: 1337, 129: 1338, 237: 1403},
		// 810
		{452},
		{341, 106: 1406, 261: 1405},
		{454},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1407},
		{340, 89: 340, 97: 340, 340, 340, 340, 340, 103: 340, 108: 340, 340, 122: 805, 125: 340, 803, 128: 804, 130: 340},
		// 815
		{341, 106: 1406, 261: 1409},
		{455},
		{341, 106: 1406, 261: 1411},
		{456},
		{457},
		// 820
		{1: 1414},
		{327},
		{1: 1418},
		{339, 106: 1337, 129: 1338, 237: 1417},
		{458},
		// 825
		{459},
		{339, 106: 1337, 129: 1338, 237: 1420},
		{460},
		{127: 1422},
		{89: 1423},
		// 830
		{35: 1424, 45: 1425},
		{461},
		{439},
		{462},
		{463},
		// 835
		{19: 1429, 61: 1430},
		{465},
		{464},
		{1: 334, 166: 1452, 290: 1456},
		{1: 334, 166: 1452, 290: 1451},
		// 840
		{1: 1450},
		{1: 1449},
		{1: 1448},
		{1: 1443, 163: 1444},
		{1: 1442},
		// 845
		{1: 1441},
		{1: 1440},
		{466},
		{467},
		{468},
		// 850
		{469, 163: 1446},
		{1: 1445},
		{470},
		{1: 1447},
		{471},
		// 855
		{472},
		{473},
		{474},
		{1: 1455},
		{105: 1453},
		// 860
		{170: 1454},
		{1: 333},
		{475},
		{1: 1457},
		{476},
		// 865
		{336, 103: 336, 124: 336, 153: 1393, 291: 1459},
		{332, 103: 332, 124: 1346, 274: 1460},
		{98, 103: 1374, 248: 1461},
		{477},
		{478},
		// 870
		{453},
		{479},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 163: 1276, 703, 1275, 240: 1466},
		{214: 1470, 1472, 1469, 1471, 279: 1468, 310: 1467},
		{544, 93: 1504},
		// 875
		{543, 93: 543},
		{1: 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 270: 1478, 1501},
		{1: 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 270: 1478, 1499},
		{1: 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 537, 270: 1478, 1477},
		{1: 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 517, 138: 1475, 163: 517, 219: 1474, 366: 1473},
		// 880
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 163: 1276, 703, 1275, 240: 1476},
		{1: 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 516, 163: 516},
		{1: 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 515, 163: 515},
		{538, 93: 538},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 693, 195: 1479},
		// 885
		{1: 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536, 536},
		{2: 1494, 21: 1490, 25: 1493, 1488, 29: 1487, 32: 1491, 42: 1492, 212: 1496, 1484, 220: 1485, 1498, 223: 1486, 1483, 1489, 227: 1497, 1482, 1495, 317: 1480, 321: 1481},
		{539, 93: 539},
		{535, 93: 535},
		{534, 93: 534},
		// 890
		{533, 93: 533},
		{532, 93: 532},
		{531, 93: 531},
		{530, 93: 530},
		{529, 93: 529},
		// 895
		{528, 93: 528},
		{527, 93: 527},
		{526, 93: 526},
		{525, 93: 525},
		{524, 93: 524},
		// 900
		{523, 93: 523},
		{522, 93: 522},
		{521, 93: 521},
		{520, 93: 520},
		{519, 93: 519},
		// 905
		{518, 93: 518},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 693, 195: 1500},
		{540, 93: 540},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 693, 195: 1502},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 693, 195: 1503},
		// 910
		{541, 93: 541},
		{214: 1470, 1472, 1469, 1471, 279: 1505},
		{542, 93: 542},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 163: 1276, 703, 1275, 240: 1509, 305: 1508, 364: 1507},
		{548, 93: 1512},
		// 915
		{547, 93: 547},
		{219: 1510},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 163: 1276, 703, 1275, 240: 1511},
		{545, 93: 545},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 163: 1276, 703, 1275, 240: 1509, 305: 1513},
		// 920
		{546, 93: 546},
		{550},
		{549},
		{256: 1518},
		{256: 513},
		// 925
		{1: 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 512, 163: 512, 166: 1520, 328: 1519},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 163: 1276, 703, 1275, 240: 1522},
		{170: 1521},
		{1: 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 511, 163: 511},
		{510, 222: 1524, 226: 1525, 316: 1523},
		// 930
		{551},
		{509},
		{508},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 693, 195: 1561, 277: 1569, 358: 1568},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 693, 195: 1561, 277: 1560},
		// 935
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 92: 1556, 164: 703, 1555, 278: 1557},
		{276: 1553},
		{78: 1536},
		{1: 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481, 483, 481, 481, 481, 481, 481, 481, 481, 481, 481, 481},
		{1: 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480, 482, 480, 480, 480, 480, 480, 480, 480, 480, 480, 480},
		// 940
		{1: 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 90: 402, 402, 402, 94: 402, 402, 402, 102: 402, 104: 402, 402, 123: 402, 127: 402, 166: 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 402, 242: 402, 257: 402, 319: 1534},
		{1: 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 90: 403, 403, 403, 94: 403, 403, 403, 102: 403, 104: 403, 403, 123: 403, 127: 403, 166: 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 403, 242: 403, 257: 1535},
		{1: 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 90: 401, 401, 401, 94: 401, 401, 401, 102: 401, 104: 401, 401, 123: 401, 127: 401, 166: 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 401, 242: 401, 257: 401},
		{37: 1539, 268: 1540, 306: 1538, 369: 1537},
		{559, 93: 1551},
		// 945
		{507, 93: 507},
		{59: 1543},
		{63: 1542, 375: 1541},
		{504, 93: 504},
		{503, 93: 503},
		// 950
		{71: 1545, 1547, 268: 1546, 370: 1544},
		{505, 93: 505},
		{268: 1550},
		{50: 1548, 80: 1549},
		{499, 93: 499},
		// 955
		{501, 93: 501},
		{500, 93: 500},
		{502, 93: 502},
		{37: 1539, 268: 1540, 306: 1552},
		{506, 93: 506},
		// 960
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 92: 1556, 164: 703, 1555, 278: 1554},
		{560},
		{89, 230: 89},
		{88, 230: 88},
		{562, 230: 1558},
		// 965
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 92: 1556, 164: 703, 1555, 278: 1559},
		{561},
		{563},
		{145: 1562},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 135: 1565, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1567, 323: 1564, 348: 1566, 357: 1563},
		// 970
		{556, 93: 556},
		{555, 93: 555},
		{554, 93: 554},
		{553, 93: 553},
		{552, 93: 552, 122: 805, 126: 803, 128: 804},
		// 975
		{564, 93: 1570},
		{558, 93: 558},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 693, 195: 1561, 277: 1571},
		{557, 93: 557},
		{565},
		// 980
		{89: 1288, 97: 998, 999, 997, 996, 245: 994},
		{102: 1582, 138: 1581},
		{93: 570, 102: 570, 234: 570, 570},
		{93: 1579, 102: 568, 234: 568, 568},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 1574, 273: 1575, 283: 1578},
		// 985
		{93: 1579, 102: 567, 234: 567, 567},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 164: 703, 1574, 273: 1580},
		{93: 569, 102: 569, 234: 569, 569},
		{102: 1592},
		{1: 1584, 282: 1583},
		// 990
		{89: 1585, 93: 1586},
		{89: 379, 93: 379},
		{138: 1588},
		{1: 1587},
		{89: 378, 93: 378},
		// 995
		{102: 1589},
		{102: 611, 234: 608, 610, 238: 982, 1590, 241: 609},
		{89: 1591, 97: 998, 999, 997, 996, 245: 994},
		{93: 571, 102: 571, 234: 571, 571},
		{102: 611, 234: 608, 610, 238: 982, 1593, 241: 609},
		// 1000
		{89: 1594, 97: 998, 999, 997, 996, 245: 994},
		{93: 572, 102: 572, 234: 572, 572},
		{1: 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 90: 404, 404, 404, 94: 404, 404, 404, 102: 404, 104: 404, 404, 123: 404, 127: 404, 166: 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 404, 242: 404, 257: 404, 269: 1533, 272: 1599},
		{97: 998, 999, 997, 996, 245: 1597},
		{102: 611, 234: 608, 610, 238: 982, 1598, 241: 609},
		// 1005
		{575, 89: 575, 97: 575, 575, 575, 575, 245: 994},
		{1: 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 90: 400, 400, 400, 94: 400, 400, 400, 102: 400, 104: 400, 400, 123: 1602, 127: 400, 166: 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 242: 1601, 295: 1600},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 1605},
		{1: 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 90: 399, 399, 399, 94: 399, 399, 399, 102: 399, 104: 399, 399, 123: 1604, 127: 399, 166: 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399, 399},
		{1: 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 90: 398, 398, 398, 94: 398, 398, 398, 102: 398, 104: 398, 398, 127: 398, 166: 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 398, 242: 1603},
		// 1010
		{1: 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 90: 396, 396, 396, 94: 396, 396, 396, 102: 396, 104: 396, 396, 127: 396, 166: 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396, 396},
		{1: 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 90: 397, 397, 397, 94: 397, 397, 397, 102: 397, 104: 397, 397, 127: 397, 166: 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397, 397},
		{93: 825, 97: 98, 98, 98, 98, 103: 1374, 124: 1607, 248: 1606},
		{578, 89: 578, 97: 578, 578, 578, 578},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 102: 1613, 104: 1615, 163: 1276, 703, 1275, 194: 1617, 240: 1616, 252: 1614, 1612, 1611, 284: 1610, 1608, 304: 1609},
		// 1015
		{577, 89: 577, 97: 577, 577, 577, 577},
		{341, 89: 341, 93: 1681, 97: 341, 341, 341, 341, 341, 103: 341, 106: 1406, 108: 341, 341, 125: 341, 130: 341, 261: 1680},
		{413, 89: 413, 97: 413, 413, 413, 413},
		{377, 89: 377, 1624, 1625, 93: 377, 97: 377, 377, 377, 377, 377, 103: 377, 106: 377, 108: 377, 377, 1623, 123: 1622, 125: 377, 130: 377, 1627, 1626, 1628, 258: 1621},
		{366, 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 366, 366, 366, 836, 366, 97: 366, 366, 366, 366, 366, 103: 366, 106: 366, 108: 366, 366, 366, 123: 366, 125: 366, 130: 366, 366, 366, 366, 366, 366, 137: 366, 835, 164: 703, 834, 209: 366, 366, 366, 262: 1660},
		// 1020
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 102: 1657, 104: 1615, 163: 1276, 703, 1275, 194: 1617, 234: 608, 610, 238: 982, 983, 1616, 609, 252: 1614, 1612, 1658},
		{373, 89: 373, 373, 373, 93: 373, 97: 373, 373, 373, 373, 373, 103: 373, 106: 373, 108: 373, 373, 373, 123: 373, 125: 373, 130: 373, 373, 373, 373, 373, 373, 137: 373},
		{349: 1618},
		{349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 349, 97: 349, 349, 349, 349, 349, 103: 349, 106: 349, 108: 349, 349, 349, 123: 349, 125: 349, 130: 349, 349, 349, 349, 349, 349, 137: 349, 349, 209: 349, 349, 349},
		{348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 348, 97: 348, 348, 348, 348, 348, 103: 348, 106: 348, 108: 348, 348, 348, 123: 348, 125: 348, 130: 348, 348, 348, 348, 348, 348, 137: 348, 348, 209: 348, 348, 348},
		// 1025
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 102: 1613, 104: 1615, 163: 1276, 703, 1275, 194: 1617, 240: 1616, 252: 1619, 1612, 1620},
		{90: 373, 373, 110: 373, 123: 373, 131: 373, 373, 373, 1656},
		{90: 1624, 1625, 110: 1623, 123: 1622, 131: 1627, 1626, 1628, 258: 1621},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 102: 1613, 104: 1615, 163: 1276, 703, 1275, 194: 1617, 240: 1616, 252: 1614, 1612, 1649},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 102: 1613, 104: 1615, 163: 1276, 703, 1275, 194: 1617, 240: 1616, 252: 1614, 1612, 1646},
		// 1030
		{1: 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 361, 102: 361, 104: 361, 163: 361},
		{110: 1643, 267: 1644},
		{110: 1640, 267: 1641},
		{110: 1639},
		{110: 1638},
		// 1035
		{90: 1631, 1630, 110: 1629},
		{1: 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 354, 102: 354, 104: 354, 163: 354},
		{110: 1635, 267: 1636},
		{110: 1632, 267: 1633},
		{1: 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 351, 102: 351, 104: 351, 163: 351},
		// 1040
		{110: 1634},
		{1: 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 350, 102: 350, 104: 350, 163: 350},
		{1: 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 353, 102: 353, 104: 353, 163: 353},
		{110: 1637},
		{1: 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 352, 102: 352, 104: 352, 163: 352},
		// 1045
		{1: 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 355, 102: 355, 104: 355, 163: 355},
		{1: 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 356, 102: 356, 104: 356, 163: 356},
		{1: 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 358, 102: 358, 104: 358, 163: 358},
		{110: 1642},
		{1: 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 357, 102: 357, 104: 357, 163: 357},
		// 1050
		{1: 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 360, 102: 360, 104: 360, 163: 360},
		{110: 1645},
		{1: 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 359, 102: 359, 104: 359, 163: 359},
		{370, 89: 370, 370, 370, 93: 370, 97: 370, 370, 370, 370, 370, 103: 370, 106: 370, 108: 370, 370, 370, 123: 370, 125: 370, 130: 370, 370, 370, 1628, 370, 1647, 137: 370, 258: 1621},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1648},
		// 1055
		{368, 89: 368, 368, 368, 93: 368, 97: 368, 368, 368, 368, 368, 103: 368, 106: 368, 108: 368, 368, 368, 122: 805, 368, 125: 368, 803, 128: 804, 130: 368, 368, 368, 368, 368, 368, 137: 368},
		{371, 89: 371, 371, 371, 93: 371, 97: 371, 371, 371, 371, 371, 103: 371, 106: 371, 108: 371, 371, 371, 123: 371, 125: 371, 130: 371, 371, 371, 1628, 371, 1650, 137: 1651, 258: 1621},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1655},
		{102: 1652},
		{1: 1584, 282: 1653},
		// 1060
		{89: 1654, 93: 1586},
		{367, 89: 367, 367, 367, 93: 367, 97: 367, 367, 367, 367, 367, 103: 367, 106: 367, 108: 367, 367, 367, 123: 367, 125: 367, 130: 367, 367, 367, 367, 367, 367, 137: 367},
		{369, 89: 369, 369, 369, 93: 369, 97: 369, 369, 369, 369, 369, 103: 369, 106: 369, 108: 369, 369, 369, 122: 805, 369, 125: 369, 803, 128: 804, 130: 369, 369, 369, 369, 369, 369, 137: 369},
		{372, 89: 372, 372, 372, 93: 372, 97: 372, 372, 372, 372, 372, 103: 372, 106: 372, 108: 372, 372, 372, 123: 372, 125: 372, 130: 372, 372, 372, 372, 372, 372, 137: 372},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 102: 1657, 104: 1615, 163: 1276, 703, 1275, 194: 1617, 234: 608, 610, 238: 982, 992, 1616, 609, 252: 1614, 1612, 1658},
		// 1065
		{89: 1659, 1624, 1625, 110: 1623, 123: 1622, 131: 1627, 1626, 1628, 258: 1621},
		{374, 89: 374, 374, 374, 93: 374, 97: 374, 374, 374, 374, 374, 103: 374, 106: 374, 108: 374, 374, 374, 123: 374, 125: 374, 130: 374, 374, 374, 374, 374, 374, 137: 374},
		{347, 89: 347, 347, 347, 93: 347, 97: 347, 347, 347, 347, 347, 103: 347, 106: 347, 108: 347, 347, 347, 123: 347, 125: 347, 130: 347, 347, 347, 347, 347, 347, 137: 347, 209: 1662, 1664, 1663, 342: 1661},
		{375, 89: 375, 375, 375, 93: 375, 97: 375, 375, 375, 375, 375, 103: 375, 106: 375, 108: 375, 375, 375, 123: 375, 125: 375, 130: 375, 375, 375, 375, 375, 375, 137: 375},
		{263: 1676},
		// 1070
		{263: 1672},
		{263: 1665},
		{102: 1666},
		{1: 1668, 275: 1667},
		{89: 1669, 93: 1670},
		// 1075
		{89: 343, 93: 343},
		{344, 89: 344, 344, 344, 93: 344, 97: 344, 344, 344, 344, 344, 103: 344, 106: 344, 108: 344, 344, 344, 123: 344, 125: 344, 130: 344, 344, 344, 344, 344, 344, 137: 344},
		{1: 1671},
		{89: 342, 93: 342},
		{102: 1673},
		// 1080
		{1: 1668, 275: 1674},
		{89: 1675, 93: 1670},
		{345, 89: 345, 345, 345, 93: 345, 97: 345, 345, 345, 345, 345, 103: 345, 106: 345, 108: 345, 345, 345, 123: 345, 125: 345, 130: 345, 345, 345, 345, 345, 345, 137: 345},
		{102: 1677},
		{1: 1668, 275: 1678},
		// 1085
		{89: 1679, 93: 1670},
		{346, 89: 346, 346, 346, 93: 346, 97: 346, 346, 346, 346, 346, 103: 346, 106: 346, 108: 346, 346, 346, 123: 346, 125: 346, 130: 346, 346, 346, 346, 346, 346, 137: 346},
		{110, 89: 110, 97: 110, 110, 110, 110, 110, 103: 110, 108: 110, 110, 125: 110, 130: 1684, 288: 1683},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 102: 1613, 104: 1615, 163: 1276, 703, 1275, 194: 1617, 240: 1616, 252: 1614, 1612, 1682},
		{376, 89: 376, 1624, 1625, 93: 376, 97: 376, 376, 376, 376, 376, 103: 376, 106: 376, 108: 376, 376, 1623, 123: 1622, 125: 376, 130: 376, 1627, 1626, 1628, 258: 1621},
		// 1090
		{108, 89: 108, 97: 108, 108, 108, 108, 108, 103: 108, 108: 108, 108, 125: 1688, 289: 1687},
		{280: 1685},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 993, 259: 1686},
		{109, 89: 109, 93: 989, 97: 109, 109, 109, 109, 109, 103: 109, 108: 109, 109, 125: 109},
		{106, 89: 106, 97: 106, 106, 106, 106, 106, 103: 106, 108: 106, 1052, 266: 1690},
		// 1095
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 164: 703, 693, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 1689},
		{107, 89: 107, 97: 107, 107, 107, 107, 107, 103: 107, 108: 107, 107, 122: 805, 126: 803, 128: 804},
		{98, 89: 98, 97: 98, 98, 98, 98, 98, 103: 1374, 108: 98, 248: 1691},
		{94, 89: 94, 97: 94, 94, 94, 94, 1693, 108: 1694, 292: 1692},
		{574, 89: 574, 97: 576, 576, 576, 576},
		// 1100
		{372: 1698},
		{153: 1695},
		{1: 1696},
		{1: 1697},
		{92, 89: 92, 97: 92, 92, 92, 92},
		// 1105
		{93, 89: 93, 97: 93, 93, 93, 93},
		{1: 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 90: 400, 400, 400, 94: 400, 400, 400, 102: 400, 104: 400, 400, 123: 1602, 127: 400, 166: 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 400, 242: 1601, 295: 1700},
		{1: 688, 687, 674, 675, 677, 678, 680, 681, 683, 685, 752, 753, 754, 755, 756, 757, 758, 759, 760, 695, 673, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 697, 708, 671, 716, 723, 727, 734, 736, 630, 749, 684, 702, 686, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 665, 666, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 90: 654, 655, 694, 94: 689, 679, 690, 102: 631, 104: 647, 632, 127: 819, 164: 703, 821, 676, 652, 636, 672, 643, 701, 698, 682, 700, 656, 657, 653, 658, 659, 660, 661, 662, 663, 664, 624, 625, 667, 670, 669, 692, 691, 696, 646, 641, 639, 699, 640, 645, 651, 650, 648, 649, 668, 644, 637, 642, 638, 635, 218: 634, 231: 633, 820, 817, 236: 1701},
		{98, 89: 98, 93: 825, 97: 98, 98, 98, 98, 103: 1374, 124: 1702, 248: 1606},
		{1: 773, 794, 777, 778, 779, 780, 781, 783, 786, 793, 752, 753, 754, 755, 756, 757, 758, 759, 760, 782, 776, 762, 767, 710, 712, 713, 715, 718, 720, 730, 742, 787, 708, 775, 716, 723, 727, 734, 736, 785, 749, 788, 791, 792, 769, 705, 706, 707, 709, 711, 774, 714, 717, 722, 724, 725, 726, 728, 729, 731, 732, 735, 737, 738, 740, 741, 743, 744, 745, 746, 747, 748, 750, 761, 763, 789, 790, 764, 765, 766, 768, 704, 719, 721, 733, 739, 784, 751, 102: 1613, 104: 1615, 163: 1276, 703, 1275, 194: 1617, 240: 1616, 252: 1614, 1612, 1611, 284: 1610, 1608, 304: 1703},
		// 1110
		{341, 89: 341, 93: 1681, 97: 341, 341, 341, 341, 341, 103: 341, 106: 1406, 108: 341, 341, 125: 341, 130: 341, 261: 1704},
		{110, 89: 110, 97: 110, 110, 110, 110, 110, 103: 110, 108: 110, 110, 125: 110, 130: 1684, 288: 1705},
		{108, 89: 108, 97: 108, 108, 108, 108, 108, 103: 108, 108: 108, 108, 125: 1688, 289: 1706},
		{106, 89: 106, 97: 106, 106, 106, 106, 106, 103: 106, 108: 106, 1052, 266: 1707},
		{98, 89: 98, 97: 98, 98, 98, 98, 98, 103: 1374, 108: 98, 248: 1708},
		// 1115
		{94, 89: 94, 97: 94, 94, 94, 94, 1693, 108: 1694, 292: 1709},
		{576, 89: 576, 97: 576, 576, 576, 576},
		{98, 103: 1374, 248: 1711},
		{580},
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
		return __yyfmt__.Sprintf("'%c'", c)
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
	const yyError = 378

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
	case 13:
		{
			yyVAL.selStmt = &Select{QueryGlobals: &QueryGlobals{}, SelectExprs: []SelectExpr{&StarExpr{}}, From: []TableExpr{&AliasedTableExpr{Expr: yyS[yypt-2].subquery}}, OrderBy: yyS[yypt-1].orderBy, Limit: yyS[yypt-0].limit}
		}
	case 14:
		{
			yyVAL.selStmt = &Select{QueryGlobals: &QueryGlobals{}, SelectExprs: []SelectExpr{&StarExpr{}}, From: []TableExpr{&AliasedTableExpr{Expr: yyS[yypt-0].subquery}}}
		}
	case 15:
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyS[yypt-3].strs), QueryGlobals: yyS[yypt-2].queryGlobals, SelectExprs: yyS[yypt-1].selectExprs, Limit: yyS[yypt-0].limit}
		}
	case 16:
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyS[yypt-4].strs), QueryGlobals: yyS[yypt-3].queryGlobals, SelectExprs: yyS[yypt-2].selectExprs, From: yyS[yypt-0].tableExprs}
		}
	case 17:
		{
			yyVAL.selStmt = &Select{Comments: Comments(yyS[yypt-10].strs), QueryGlobals: yyS[yypt-9].queryGlobals, SelectExprs: yyS[yypt-8].selectExprs, From: yyS[yypt-6].tableExprs, Where: NewWhere(AST_WHERE, yyS[yypt-5].expr), GroupBy: GroupBy(yyS[yypt-4].exprs), Having: NewWhere(AST_HAVING, yyS[yypt-3].expr), OrderBy: yyS[yypt-2].orderBy, Limit: yyS[yypt-1].limit, Lock: yyS[yypt-0].str}
		}
	case 18:
		{
			yyVAL.selStmt = &Union{With: yyS[yypt-3].with, Left: yyS[yypt-2].selStmt, Right: yyS[yypt-0].selStmt, Type: yyS[yypt-1].str}
		}
	case 19:
		{
			yyVAL.selStmt = &Select{With: yyS[yypt-12].with, Comments: Comments(yyS[yypt-10].strs), QueryGlobals: yyS[yypt-9].queryGlobals, SelectExprs: yyS[yypt-8].selectExprs, From: yyS[yypt-6].tableExprs, Where: NewWhere(AST_WHERE, yyS[yypt-5].expr), GroupBy: GroupBy(yyS[yypt-4].exprs), Having: NewWhere(AST_HAVING, yyS[yypt-3].expr), OrderBy: yyS[yypt-2].orderBy, Limit: yyS[yypt-1].limit, Lock: yyS[yypt-0].str}
		}
	case 20:
		{
			yyVAL.selStmt = &Union{Type: yyS[yypt-1].str, Left: yyS[yypt-2].selStmt, Right: yyS[yypt-0].selStmt}
		}
	case 21:
		{
			yyVAL.cte = &CTE{TableName: &TableName{Name: yyS[yypt-4].str}, ColumnExprs: nil, Query: yyS[yypt-1].selStmt}
		}
	case 22:
		{
			yyVAL.cte = &CTE{TableName: &TableName{Name: yyS[yypt-7].str}, ColumnExprs: yyS[yypt-5].columnExprs, Query: yyS[yypt-1].selStmt}
		}
	case 23:
		{
			yyVAL.cte_list = []*CTE{yyS[yypt-0].cte}
		}
	case 24:
		{
			yyVAL.cte_list = append(yyS[yypt-2].cte_list, yyS[yypt-0].cte)
		}
	case 25:
		{
			yyVAL.with = &With{CTEs: yyS[yypt-0].cte_list, Recursive: false}
		}
	case 26:
		{
			yyVAL.with = &With{CTEs: yyS[yypt-0].cte_list, Recursive: true}
		}
	case 27:
		{
			yyVAL.subquery = &Subquery{Select: yyS[yypt-1].selStmt, IsDerived: false}
		}
	case 28:
		{
			yyVAL.statement = &Use{DBName: string(yyS[yypt-0].bytes)}
		}
	case 29:
		{
			yyVAL.statement = &Set{Comments: Comments(yyS[yypt-1].strs), Exprs: yyS[yypt-0].updateExprs}
		}
	case 30:
		{
			yyVAL.statement = &Set{Scope: yyS[yypt-1].str, Exprs: UpdateExprs(append([]*UpdateExpr{}, yyS[yypt-0].updateExpr))}
		}
	case 31:
		{
			yyVAL.statement = &Set{Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: "@@character_set_client"}, Expr: StrVal(yyS[yypt-0].str)},
				&UpdateExpr{Name: &ColName{Name: "@@character_set_results"}, Expr: StrVal(yyS[yypt-0].str)},
				&UpdateExpr{Name: &ColName{Name: "@@character_set_connection"}, Expr: StrVal(yyS[yypt-0].str)},
			}}
		}
	case 32:
		{
			yyVAL.statement = &Set{Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: "@@character_set_client"}, Expr: StrVal(yyS[yypt-2].str)},
				&UpdateExpr{Name: &ColName{Name: "@@character_set_results"}, Expr: StrVal(yyS[yypt-2].str)},
				&UpdateExpr{Name: &ColName{Name: "@@character_set_connection"}, Expr: StrVal(yyS[yypt-2].str)},
				&UpdateExpr{Name: &ColName{Name: "@@collation_connection"}, Expr: StrVal(yyS[yypt-0].str)},
			}}
		}
	case 33:
		{
			yyVAL.statement = &Set{Exprs: UpdateExprs{
				&UpdateExpr{Name: &ColName{Name: "@@character_set_client"}, Expr: StrVal(yyS[yypt-0].str)},
				&UpdateExpr{Name: &ColName{Name: "@@character_set_results"}, Expr: StrVal(yyS[yypt-0].str)},
				&UpdateExpr{Name: &ColName{Name: "@@collation_connection"}, Expr: &ColName{Name: "@@collation_database"}},
			}}
		}
	case 34:
		{
			yyVAL.statement = &Set{Comments: append([]string{}, yyS[yypt-2].str, string(TRANSACTION_BYTES), yyS[yypt-0].str)}
		}
	case 35:
		{
			yyVAL.updateExprs = UpdateExprs{yyS[yypt-0].updateExpr}
		}
	case 36:
		{
			yyVAL.updateExprs = append(yyS[yypt-2].updateExprs, yyS[yypt-0].updateExpr)
		}
	case 37:
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyS[yypt-2].colName, Expr: yyS[yypt-0].expr}
		}
	case 38:
		{
			yyVAL.updateExpr = &UpdateExpr{Name: yyS[yypt-2].colName, Expr: StrVal("default")}
		}
	case 39:
		{
			yyVAL.expr = NumVal([]byte("1"))
		}
	case 40:
		{
			yyVAL.expr = NumVal([]byte("0"))
		}
	case 41:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 42:
		{
			yyVAL.statement = &DropTable{Name: yyS[yypt-1].tableName, Temporary: yyS[yypt-4].bool, Exists: yyS[yypt-2].bool, Opt: yyS[yypt-0].stropt}
		}
	case 43:
		{
			yyVAL.statement = &Flush{Kind: FlushLogs}
		}
	case 44:
		{
			yyVAL.statement = &Flush{Kind: FlushSample}
		}
	case 45:
		{
			yyVAL.statement = &RenameTable{Renames: yyS[yypt-0].renameSpecs}
		}
	case 46:
		{
			yyVAL.renameSpecs = []*RenameSpec{yyS[yypt-0].renameSpec}
		}
	case 47:
		{
			yyVAL.renameSpecs = append(yyVAL.renameSpecs, yyS[yypt-0].renameSpec)
		}
	case 48:
		{
			yyVAL.renameSpec = &RenameSpec{Table: yyS[yypt-2].tableName, NewTable: yyS[yypt-0].tableName}
		}
	case 49:
		{
			yyVAL.statement = &AlterTable{Table: yyS[yypt-1].tableName, Specs: yyS[yypt-0].alterSpecs}
		}
	case 50:
		{
			yyVAL.alterSpecs = []*AlterSpec{yyS[yypt-0].alterSpec}
		}
	case 51:
		{
			yyVAL.alterSpecs = append(yyVAL.alterSpecs, yyS[yypt-0].alterSpec)
		}
	case 52:
		{
			yyVAL.alterSpec = &AlterSpec{
				Type:      "rename_column",
				Column:    yyS[yypt-1].colName,
				NewColumn: yyS[yypt-0].colName,
			}
		}
	case 53:
		{
			yyVAL.alterSpec = &AlterSpec{
				Type:   "drop_column",
				Column: yyS[yypt-0].colName,
			}
		}
	case 54:
		{
			yyVAL.alterSpec = &AlterSpec{
				Type:          "modify_column",
				Column:        yyS[yypt-1].colName,
				NewColumnType: string(yyS[yypt-0].bytes),
			}
		}
	case 55:
		{
			yyVAL.alterSpec = &AlterSpec{
				Type:     "rename_table",
				NewTable: yyS[yypt-0].tableName,
			}
		}
	case 56:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 57:
		{
			yyVAL.stropt = option.SomeString(string(COLUMN_BYTES))
		}
	case 58:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 59:
		{
			yyVAL.bytes = TINYINT_BYTES
		}
	case 60:
		{
			yyVAL.bytes = INT_BYTES
		}
	case 61:
		{
			yyVAL.bytes = INTEGER_BYTES
		}
	case 62:
		{
			yyVAL.bytes = BIGINT_BYTES
		}
	case 63:
		{
			yyVAL.bytes = DOUBLE_BYTES
		}
	case 64:
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 65:
		{
			yyVAL.bytes = DECIMAL_BYTES
		}
	case 66:
		{
			yyVAL.bytes = NUMERIC_BYTES
		}
	case 67:
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 68:
		{
			yyVAL.bytes = TIME_BYTES
		}
	case 69:
		{
			yyVAL.bytes = TIMESTAMP_BYTES
		}
	case 70:
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 71:
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 72:
		{
			yyVAL.bytes = VARCHAR_BYTES
		}
	case 73:
		{
			yyVAL.bytes = BINARY_BYTES
		}
	case 74:
		{
			yyVAL.bytes = TEXT_BYTES
		}
	case 75:
		{
			yyVAL.bytes = BOOLEAN_BYTES
		}
	case 76:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 77:
		{
			yyVAL.stropt = option.SomeString(string(TO_BYTES))
		}
	case 78:
		{
			yyVAL.stropt = option.SomeString(string(AS_BYTES))
		}
	case 79:
		{
			yyVAL.bool = false
		}
	case 80:
		{
			yyVAL.bool = true
		}
	case 81:
		{
			yyVAL.bool = false
		}
	case 82:
		{
			yyVAL.bool = true
		}
	case 83:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 84:
		{
			yyVAL.stropt = option.SomeString(string(CASCADE_BYTES))
		}
	case 85:
		{
			yyVAL.stropt = option.SomeString(string(RESTRICT_BYTES))
		}
	case 86:
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 87:
		{
			yyVAL.str = yyS[yypt-2].str + ", " + yyS[yypt-0].str
		}
	case 88:
		{
			yyVAL.str = "isolation level " + yyS[yypt-0].str
		}
	case 89:
		{
			yyVAL.str = "read write"
		}
	case 90:
		{
			yyVAL.str = "read only"
		}
	case 91:
		{
			yyVAL.str = "repeatable read"
		}
	case 92:
		{
			yyVAL.str = "read committed"
		}
	case 93:
		{
			yyVAL.str = "read uncommitted"
		}
	case 94:
		{
			yyVAL.str = string(SERIALIZABLE_BYTES)
		}
	case 95:
		{
			yyVAL.bytes = SUBSTR_BYTES
		}
	case 96:
		{
			yyVAL.bytes = SUBSTRING_BYTES
		}
	case 99:
		{
			yyVAL.expr = StrVal(yyS[yypt-0].bytes)
		}
	case 100:
		{
			yyVAL.expr = &ColName{Qualifier: option.SomeString(string(yyS[yypt-2].bytes)), Name: string(yyS[yypt-0].bytes)}
		}
	case 101:
		{
			yyVAL.expr = &ColName{Qualifier: option.SomeString(string(yyS[yypt-0].bytes)), Name: string(yyS[yypt-2].bytes)}
		}
	case 102:
		{
			yyVAL.expr = StrVal(yyS[yypt-0].bytes)
		}
	case 103:
		{
			yyVAL.expr = nil
		}
	case 104:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 105:
		{
			yyVAL.str = AST_SHOW_NO_MOD
		}
	case 106:
		{
			yyVAL.str = AST_SHOW_FULL
		}
	case 107:
		{
			yyVAL.str = AST_KILL_CONNECTION
		}
	case 108:
		{
			yyVAL.str = AST_KILL_QUERY
		}
	case 109:
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 110:
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 111:
		{
			yyVAL.str = AST_GLOBAL_SCOPE
		}
	case 112:
		{
			yyVAL.str = AST_SESSION_SCOPE
		}
	case 113:
		{
			yyVAL.str = AST_GLOBAL_SCOPE
		}
	case 114:
		{
			yyVAL.statement = &Show{Section: "binary logs"}
		}
	case 115:
		{
			yyVAL.statement = &Show{Section: "master logs"}
		}
	case 116:
		{
			yyVAL.statement = &Show{Section: "binlog events"}
		}
	case 117:
		{
			yyVAL.statement = &Show{Section: "create database", Modifier: yyS[yypt-1].str, From: StrVal(yyS[yypt-0].bytes)}
		}
	case 118:
		{
			yyVAL.statement = &Show{Section: "create schema", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 119:
		{
			yyVAL.statement = &Show{Section: "create event", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 120:
		{
			yyVAL.statement = &Show{Section: "create function", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 121:
		{
			yyVAL.statement = &Show{Section: "create procedure", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 122:
		{
			yyVAL.statement = &Show{Section: "create table", From: &ColName{option.NoneString(), option.SomeString(string(yyS[yypt-2].bytes)), string(yyS[yypt-0].bytes)}}
		}
	case 123:
		{
			yyVAL.statement = &Show{Section: "create table", From: StrVal(yyS[yypt-0].bytes)}
		}
	case 124:
		{
			yyVAL.statement = &Show{Section: "create table", From: StrVal(yyS[yypt-0].bytes)}
		}
	case 125:
		{
			yyVAL.statement = &Show{Section: "create trigger", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 126:
		{
			yyVAL.statement = &Show{Section: "create user", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 127:
		{
			yyVAL.statement = &Show{Section: "create view", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 128:
		{
			yyVAL.statement = &Show{Section: "engine", Modifier: string(yyS[yypt-1].bytes)}
		}
	case 129:
		{
			yyVAL.statement = &Show{Section: "engine", Modifier: string(yyS[yypt-1].bytes)}
		}
	case 130:
		{
			yyVAL.statement = &Show{Section: "engines"}
		}
	case 131:
		{
			yyVAL.statement = &Show{Section: "errors"}
		}
	case 132:
		{
			yyVAL.statement = &Show{Section: "count(*) errors"}
		}
	case 133:
		{
			yyVAL.statement = &Show{Section: "events"}
		}
	case 134:
		{
			yyVAL.statement = &Show{Section: "function code", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 135:
		{
			yyVAL.statement = &Show{Section: "function status"}
		}
	case 136:
		{
			yyVAL.statement = &Show{Section: "grants", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 137:
		{
			yyVAL.statement = &Show{Section: "indexes", From: yyS[yypt-1].expr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 138:
		{
			yyVAL.statement = &Show{Section: "indexes", From: yyS[yypt-1].expr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 139:
		{
			yyVAL.statement = &Show{Section: "keys", From: yyS[yypt-1].expr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 140:
		{
			yyVAL.statement = &Show{Section: "master status"}
		}
	case 141:
		{
			yyVAL.statement = &Show{Section: "open tables"}
		}
	case 142:
		{
			yyVAL.statement = &Show{Section: "plugins"}
		}
	case 143:
		{
			yyVAL.statement = &Show{Section: "privileges"}
		}
	case 144:
		{
			yyVAL.statement = &Show{Section: "procedure code", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 145:
		{
			yyVAL.statement = &Show{Section: "procedure status"}
		}
	case 146:
		{
			yyVAL.statement = &Show{Section: "profile"}
		}
	case 147:
		{
			yyVAL.statement = &Show{Section: "profiles"}
		}
	case 148:
		{
			yyVAL.statement = &Show{Section: "relaylog events"}
		}
	case 149:
		{
			yyVAL.statement = &Show{Section: "slave hosts"}
		}
	case 150:
		{
			yyVAL.statement = &Show{Section: "slave status", Modifier: string(yyS[yypt-0].bytes)}
		}
	case 151:
		{
			yyVAL.statement = &Show{Section: "table status"}
		}
	case 152:
		{
			yyVAL.statement = &Show{Section: "table status"}
		}
	case 153:
		{
			yyVAL.statement = &Show{Section: "warnings"}
		}
	case 154:
		{
			yyVAL.statement = &Show{Section: "count(*) errors"}
		}
	case 155:
		{
			yyVAL.statement = &Show{Section: "databases", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 156:
		{
			yyVAL.statement = &Show{Section: "schemas", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 157:
		{
			yyVAL.statement = &Show{Section: "variables", Modifier: yyS[yypt-2].str, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 158:
		{
			yyVAL.statement = &Show{Section: "tables", Modifier: yyS[yypt-3].str, From: yyS[yypt-1].expr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 159:
		{
			yyVAL.statement = &Show{Section: "proxy", Key: string(yyS[yypt-2].bytes), From: yyS[yypt-1].expr, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 160:
		{
			yyVAL.statement = &Show{Section: "columns", From: yyS[yypt-1].expr, Modifier: yyS[yypt-3].str, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 161:
		{
			yyVAL.statement = &Show{Section: "processlist", Modifier: yyS[yypt-1].str}
		}
	case 162:
		{
			yyVAL.statement = &Show{Section: "status", Modifier: yyS[yypt-2].str, LikeOrWhere: yyS[yypt-0].expr}
		}
	case 163:
		{
			yyVAL.statement = &Show{Section: "charset", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 164:
		{
			yyVAL.statement = &Show{Section: "charset", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 165:
		{
			yyVAL.statement = &Show{Section: "collation", LikeOrWhere: yyS[yypt-0].expr}
		}
	case 166:
		{
			yyVAL.str = AST_EXPLAIN_FORMAT_TRADITIONAL
		}
	case 167:
		{
			yyVAL.str = AST_EXPLAIN_FORMAT_JSON
		}
	case 168:
		{
			yyVAL.str = AST_EXPLAIN_EXTENDED
		}
	case 169:
		{
			yyVAL.str = AST_EXPLAIN_PARTITIONS
		}
	case 170:
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 171:
		{
			yyVAL.str = ""
		}
	case 172:
		{
			yyVAL.statement = yyS[yypt-0].selStmt
		}
	case 173:
		{
			yyVAL.statement = yyS[yypt-1].statement
		}
	case 174:
		{
			yyVAL.empty = yyS[yypt-0].empty
		}
	case 175:
		{
			yyVAL.empty = yyS[yypt-0].empty
		}
	case 176:
		{
			yyVAL.empty = yyS[yypt-0].empty
		}
	case 177:
		{
			yyVAL.tableName = &TableName{Name: yyS[yypt-0].str}
		}
	case 178:
		{
			yyVAL.tableName = &TableName{Name: yyS[yypt-0].str}
		}
	case 179:
		{
			yyVAL.tableName = &TableName{Qualifier: option.SomeString(yyS[yypt-2].str), Name: yyS[yypt-0].str}
		}
	case 180:
		{
			yyVAL.tableExprs = TableExprs{&DualTableExpr{}}
		}
	case 181:
		{
			yyVAL.statement = &Explain{Section: "table", Table: yyS[yypt-1].tableName, Column: yyS[yypt-0].colName}
		}
	case 182:
		{
			yyVAL.statement = &Explain{Section: "plan", ExplainType: yyS[yypt-1].str, Statement: yyS[yypt-0].statement}
		}
	case 183:
		{
			yyVAL.statement = &Explain{Section: "plan", ExplainType: yyS[yypt-3].str, Connection: option.SomeString(string(yyS[yypt-0].bytes))}
		}
	case 184:
		{
			yyVAL.colName = nil
		}
	case 185:
		{
			yyVAL.colName = &ColName{Name: yyS[yypt-0].str}
		}
	case 186:
		{
			yyVAL.colName = &ColName{Name: string(yyS[yypt-0].bytes)}
		}
	case 187:
		{
			yyVAL.statement = &Kill{Scope: AST_KILL_CONNECTION, ID: yyS[yypt-0].expr}
		}
	case 188:
		{
			yyVAL.statement = &Kill{Scope: yyS[yypt-1].str, ID: yyS[yypt-0].expr}
		}
	case 189:
		{
			SetAllowComments(yylex, true)
		}
	case 190:
		{
			yyVAL.strs = yyS[yypt-0].strs
			SetAllowComments(yylex, false)
		}
	case 191:
		{
			yyVAL.strs = nil
		}
	case 192:
		{
			yyVAL.strs = append(yyS[yypt-1].strs, string(yyS[yypt-0].bytes))
		}
	case 193:
		{
			yyVAL.queryGlobals = &QueryGlobals{Distinct: false, StraightJoin: false}
		}
	case 194:
		{
			yyVAL.queryGlobals = &QueryGlobals{Distinct: true, StraightJoin: false}
		}
	case 195:
		{
			yyVAL.queryGlobals = &QueryGlobals{Distinct: false, StraightJoin: true}
		}
	case 196:
		{
			yyVAL.queryGlobals = &QueryGlobals{Distinct: true, StraightJoin: true}
		}
	case 197:
		{
			yyVAL.queryGlobals = &QueryGlobals{Distinct: true, StraightJoin: true}
		}
	case 198:
		{
			yyVAL.str = AST_UNION
		}
	case 199:
		{
			yyVAL.str = AST_UNION_ALL
		}
	case 200:
		{
			yyVAL.str = AST_SET_MINUS
		}
	case 201:
		{
			yyVAL.str = AST_EXCEPT
		}
	case 202:
		{
			yyVAL.str = AST_INTERSECT
		}
	case 203:
		{
			yyVAL.bool = false
		}
	case 204:
		{
			yyVAL.bool = true
		}
	case 205:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 206:
		{
			yyVAL.stropt = option.SomeString(string(yyS[yypt-0].bytes))
		}
	case 207:
		{
			yyVAL.selectExprs = SelectExprs{yyS[yypt-0].selectExpr}
		}
	case 208:
		{
			yyVAL.selectExprs = append(yyVAL.selectExprs, yyS[yypt-0].selectExpr)
		}
	case 209:
		{
			yyVAL.selectExpr = &StarExpr{}
		}
	case 210:
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyS[yypt-1].expr, As: yyS[yypt-0].stropt}
		}
	case 211:
		{
			yyVAL.selectExpr = &NonStarExpr{Expr: yyS[yypt-2].expr, As: yyS[yypt-1].stropt}
		}
	case 212:
		{
			yyVAL.selectExpr = &StarExpr{TableName: option.SomeString(yyS[yypt-2].str)}
		}
	case 213:
		{
			yyVAL.selectExpr = &StarExpr{DatabaseName: option.SomeString(yyS[yypt-4].str), TableName: option.SomeString(yyS[yypt-2].str)}
		}
	case 214:
		{
			yyVAL.columnExprs = ColumnExprs{&ColName{Name: string(yyS[yypt-0].bytes)}}
		}
	case 215:
		{
			yyVAL.columnExprs = append(yyVAL.columnExprs, &ColName{Name: string(yyS[yypt-0].bytes)})
		}
	case 216:
		{
			yyVAL.tableExprs = TableExprs{yyS[yypt-0].tableExpr}
		}
	case 217:
		{
			yyVAL.tableExprs = append(yyVAL.tableExprs, yyS[yypt-0].tableExpr)
		}
	case 218:
		{
			yyVAL.tableExpr = &AliasedTableExpr{Expr: yyS[yypt-2].smTableExpr, As: yyS[yypt-1].stropt, Hints: yyS[yypt-0].indexHints}
		}
	case 219:
		{
			yyVAL.tableExpr = &ParenTableExpr{Expr: yyS[yypt-1].tableExpr}
		}
	case 220:
		{
			yyVAL.tableExpr = yyS[yypt-0].tableExpr
		}
	case 221:
		{
			yyVAL.tableExpr = yyS[yypt-1].tableExpr
		}
	case 222:
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-2].tableExpr, Join: yyS[yypt-1].str, RightExpr: yyS[yypt-0].tableExpr}
		}
	case 223:
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-2].tableExpr, Join: AST_STRAIGHT_JOIN, RightExpr: yyS[yypt-0].tableExpr}
		}
	case 224:
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-4].tableExpr, Join: yyS[yypt-3].str, RightExpr: yyS[yypt-2].tableExpr, On: yyS[yypt-0].expr}
		}
	case 225:
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-4].tableExpr, Join: AST_STRAIGHT_JOIN, RightExpr: yyS[yypt-2].tableExpr, On: yyS[yypt-0].expr}
		}
	case 226:
		{
			yyVAL.tableExpr = &JoinTableExpr{LeftExpr: yyS[yypt-6].tableExpr, Join: yyS[yypt-5].str, RightExpr: yyS[yypt-4].tableExpr, Using: yyS[yypt-1].columnExprs}
		}
	case 227:
		{
			yyVAL.stropt = option.NoneString()
		}
	case 228:
		{
			yyVAL.stropt = option.SomeString(yyS[yypt-0].str)
		}
	case 229:
		{
			yyVAL.stropt = option.SomeString(yyS[yypt-0].str)
		}
	case 230:
		{
			yyVAL.stropt = option.SomeString(string(yyS[yypt-0].bytes))
		}
	case 231:
		{
			yyVAL.stropt = option.SomeString(string(yyS[yypt-0].bytes))
		}
	case 232:
		{
			yyVAL.str = AST_JOIN
		}
	case 233:
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 234:
		{
			yyVAL.str = AST_LEFT_JOIN
		}
	case 235:
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 236:
		{
			yyVAL.str = AST_RIGHT_JOIN
		}
	case 237:
		{
			yyVAL.str = AST_JOIN
		}
	case 238:
		{
			yyVAL.str = AST_CROSS_JOIN
		}
	case 239:
		{
			yyVAL.str = AST_NATURAL_JOIN
		}
	case 240:
		{
			yyVAL.str = AST_NATURAL_RIGHT_JOIN
		}
	case 241:
		{
			yyVAL.str = AST_NATURAL_RIGHT_JOIN
		}
	case 242:
		{
			yyVAL.str = AST_NATURAL_LEFT_JOIN
		}
	case 243:
		{
			yyVAL.str = AST_NATURAL_LEFT_JOIN
		}
	case 244:
		{
			yyVAL.smTableExpr = yyS[yypt-0].tableName
		}
	case 245:
		{
			yyVAL.smTableExpr = yyS[yypt-0].subquery
		}
	case 246:
		{
			yyVAL.indexHints = nil
		}
	case 247:
		{
			yyVAL.indexHints = &IndexHints{Type: AST_USE, Indexes: yyS[yypt-1].strs}
		}
	case 248:
		{
			yyVAL.indexHints = &IndexHints{Type: AST_IGNORE, Indexes: yyS[yypt-1].strs}
		}
	case 249:
		{
			yyVAL.indexHints = &IndexHints{Type: AST_FORCE, Indexes: yyS[yypt-1].strs}
		}
	case 250:
		{
			yyVAL.strs = []string{string(yyS[yypt-0].bytes)}
		}
	case 251:
		{
			yyVAL.strs = append(yyS[yypt-2].strs, string(yyS[yypt-0].bytes))
		}
	case 252:
		{
			yyVAL.expr = nil
		}
	case 253:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 254:
		{
			yyVAL.expr = nil
		}
	case 255:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 256:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 257:
		{
			yyVAL.expr = nil
		}
	case 258:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 259:
		{
			yyVAL.str = string("")
		}
	case 260:
		{
			yyVAL.str = "IF NOT EXISTS"
		}
	case 261:
		{
			yyVAL.expr = nil
		}
	case 262:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 263:
		{
			yyVAL.bytes = nil
		}
	case 264:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 265:
		{
			yyVAL.bytes = nil
		}
	case 266:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 267:
		{
			yyVAL.empty = struct{}{}
		}
	case 268:
		{
			yyVAL.empty = struct{}{}
		}
	case 269:
		{
			yyVAL.str = AST_ALL
		}
	case 270:
		{
			yyVAL.str = AST_SOME
		}
	case 271:
		{
			yyVAL.str = AST_ANY
		}
	case 272:
		{
			yyVAL.tuple = ValTuple(append(yyS[yypt-3].exprs, yyS[yypt-1].expr))
		}
	case 273:
		{
			yyVAL.tuple = ValTuple(yyS[yypt-1].exprs)
		}
	case 274:
		{
			yyVAL.subquery = &Subquery{Select: yyS[yypt-1].selStmt, IsDerived: true}
		}
	case 275:
		{
			yyVAL.exprs = Exprs{yyS[yypt-0].expr}
		}
	case 276:
		{
			yyVAL.exprs = append(yyS[yypt-2].exprs, yyS[yypt-0].expr)
		}
	case 277:
		{
			yyVAL.expr = &OrExpr{Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 278:
		{
			yyVAL.expr = &XorExpr{Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 279:
		{
			yyVAL.expr = &AndExpr{Left: yyS[yypt-2].expr, Right: yyS[yypt-0].expr}
		}
	case 280:
		{
			yyVAL.expr = &NotExpr{Expr: yyS[yypt-0].expr}
		}
	case 281:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_IS, Right: yyS[yypt-0].expr}
		}
	case 282:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_IS_NOT, Right: yyS[yypt-0].expr}
		}
	case 284:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_IS, Right: &NullVal{}}
		}
	case 285:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_IS_NOT, Right: &NullVal{}}
		}
	case 286:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_EQ, Right: yyS[yypt-0].expr}
		}
	case 287:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_EQ, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 288:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_NE, Right: yyS[yypt-0].expr}
		}
	case 289:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_NE, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 290:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_NSE, Right: yyS[yypt-0].expr}
		}
	case 291:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_NSE, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 292:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_LT, Right: yyS[yypt-0].expr}
		}
	case 293:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_LT, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 294:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_GT, Right: yyS[yypt-0].expr}
		}
	case 295:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_GT, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 296:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_LE, Right: yyS[yypt-0].expr}
		}
	case 297:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_LE, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 298:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_GE, Right: yyS[yypt-0].expr}
		}
	case 299:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_GE, SubqueryOperator: yyS[yypt-1].str, Right: yyS[yypt-0].subquery}
		}
	case 301:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_IN, SubqueryOperator: AST_IN, Right: yyS[yypt-0].subquery}
		}
	case 302:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-2].expr, Operator: AST_IN, Right: yyS[yypt-0].tuple}
		}
	case 303:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_NOT_IN, SubqueryOperator: AST_NOT_IN, Right: yyS[yypt-0].subquery}
		}
	case 304:
		{
			yyVAL.expr = &ComparisonExpr{Left: yyS[yypt-3].expr, Operator: AST_NOT_IN, Right: yyS[yypt-0].tuple}
		}
	case 305:
		{
			yyVAL.expr = &RangeCond{Left: yyS[yypt-4].expr, Operator: AST_BETWEEN, From: yyS[yypt-2].expr, To: yyS[yypt-0].expr}
		}
	case 306:
		{
			yyVAL.expr = &RangeCond{Left: yyS[yypt-5].expr, Operator: AST_NOT_BETWEEN, From: yyS[yypt-2].expr, To: yyS[yypt-0].expr}
		}
	case 307:
		{
			yyVAL.expr = &LikeExpr{Left: yyS[yypt-3].expr, Operator: AST_LIKE, Right: yyS[yypt-1].expr, Escape: yyS[yypt-0].expr}
		}
	case 308:
		{
			yyVAL.expr = &LikeExpr{Left: yyS[yypt-4].expr, Operator: AST_NOT_LIKE, Right: yyS[yypt-1].expr, Escape: yyS[yypt-0].expr}
		}
	case 309:
		{
			yyVAL.expr = &LikeExpr{Left: yyS[yypt-4].expr, Operator: AST_LIKE_BINARY, Right: yyS[yypt-1].expr, Escape: yyS[yypt-0].expr}
		}
	case 310:
		{
			yyVAL.expr = &LikeExpr{Left: yyS[yypt-5].expr, Operator: AST_NOT_LIKE_BINARY, Right: yyS[yypt-1].expr, Escape: yyS[yypt-0].expr}
		}
	case 311:
		{
			yyVAL.expr = &RegexExpr{Operand: yyS[yypt-2].expr, Pattern: yyS[yypt-0].expr}
		}
	case 312:
		{
			yyVAL.expr = &NotExpr{&RegexExpr{Operand: yyS[yypt-3].expr, Pattern: yyS[yypt-0].expr}}
		}
	case 313:
		{
			yyVAL.expr = &RLikeExpr{Operand: yyS[yypt-2].expr, Pattern: yyS[yypt-0].expr}
		}
	case 314:
		{
			yyVAL.expr = &NotExpr{&RLikeExpr{Operand: yyS[yypt-3].expr, Pattern: yyS[yypt-0].expr}}
		}
	case 316:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_BITAND, Right: yyS[yypt-0].expr}
		}
	case 317:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_BITOR, Right: yyS[yypt-0].expr}
		}
	case 318:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_PLUS, Right: yyS[yypt-0].expr}
		}
	case 319:
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
	case 320:
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
	case 321:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_MINUS, Right: yyS[yypt-0].expr}
		}
	case 322:
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
	case 323:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_MULT, Right: yyS[yypt-0].expr}
		}
	case 324:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_DIV, Right: yyS[yypt-0].expr}
		}
	case 325:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_IDIV, Right: yyS[yypt-0].expr}
		}
	case 326:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_MOD, Right: yyS[yypt-0].expr}
		}
	case 327:
		{
			yyVAL.expr = &BinaryExpr{Left: yyS[yypt-2].expr, Operator: AST_BITXOR, Right: yyS[yypt-0].expr}
		}
	case 329:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 330:
		{
			yyVAL.expr = yyS[yypt-0].colName
		}
	case 331:
		{
			yyVAL.expr = yyS[yypt-0].tuple
		}
	case 332:
		{
			yyVAL.expr = yyS[yypt-0].subquery
		}
	case 333:
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
	case 334:
		{
			yyVAL.expr = &ExistsExpr{Subquery: yyS[yypt-0].subquery}
		}
	case 335:
		{
			yyVAL.expr = yyS[yypt-0].caseExpr
		}
	case 336:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 337:
		{
			yyVAL.expr = &FuncExpr{Name: string(VALUES_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 338:
		{
			yyVAL.expr = yyS[yypt-1].expr
		}
	case 339:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 340:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 341:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 342:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 343:
		{
			yyVAL.expr = &FuncExpr{Name: string(CHAR_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 344:
		{
			yyVAL.expr = &FuncExpr{Name: string(CONVERT_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyS[yypt-3].expr}, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 345:
		{
			yyVAL.expr = &FuncExpr{Name: string(LEFT_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 346:
		{
			yyVAL.expr = &FuncExpr{Name: string(RIGHT_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 347:
		{
			yyVAL.expr = &FuncExpr{Name: string(ADDDATE_BYTES), Exprs: append(SelectExprs{yyS[yypt-5].selectExpr, yyS[yypt-2].selectExpr, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 348:
		{
			yyVAL.expr = &FuncExpr{Name: string(ADDDATE_BYTES), Exprs: append(SelectExprs{yyS[yypt-3].selectExpr, yyS[yypt-1].selectExpr, &NonStarExpr{Expr: KeywordVal(DAY_BYTES)}})}
		}
	case 349:
		{
			yyVAL.expr = &FuncExpr{Name: string(CONVERT_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyS[yypt-3].expr}, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 350:
		{
			yyVAL.expr = &FuncExpr{Name: string(CONVERT_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: yyS[yypt-4].expr}, &NonStarExpr{Expr: KeywordVal(yyS[yypt-2].bytes)}})}
		}
	case 351:
		{
			yyVAL.expr = &FuncExpr{Name: string(CURRENT_DATE_BYTES)}
		}
	case 352:
		{
			yyVAL.expr = &FuncExpr{Name: string(CURRENT_TIMESTAMP_BYTES)}
		}
	case 353:
		{
			yyVAL.expr = &FuncExpr{Name: string(CURRENT_TIMESTAMP_BYTES)}
		}
	case 354:
		{
			yyVAL.expr = &FuncExpr{Name: string(DATE_ADD_BYTES), Exprs: append(SelectExprs{yyS[yypt-5].selectExpr, yyS[yypt-2].selectExpr, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 355:
		{
			yyVAL.expr = &FuncExpr{Name: string(DATE_SUB_BYTES), Exprs: append(SelectExprs{yyS[yypt-5].selectExpr, yyS[yypt-2].selectExpr, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 356:
		{
			yyVAL.expr = &FuncExpr{Name: string(EXTRACT_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyS[yypt-3].bytes)}}, yyS[yypt-1].selectExpr)}
		}
	case 357:
		{
			yyVAL.expr = &FuncExpr{Name: string(GROUP_CONCAT_BYTES), Distinct: yyS[yypt-4].bool, Exprs: yyS[yypt-3].selectExprs, OrderBy: yyS[yypt-2].orderBy, Separator: yyS[yypt-1].stropt}
		}
	case 358:
		{
			yyVAL.expr = &FuncExpr{Name: string(SUBDATE_BYTES), Exprs: append(SelectExprs{yyS[yypt-3].selectExpr, yyS[yypt-1].selectExpr, &NonStarExpr{Expr: KeywordVal(DAY_BYTES)}})}
		}
	case 359:
		{
			yyVAL.expr = &FuncExpr{Name: string(SUBDATE_BYTES), Exprs: append(SelectExprs{yyS[yypt-5].selectExpr, yyS[yypt-2].selectExpr, &NonStarExpr{Expr: KeywordVal(yyS[yypt-1].bytes)}})}
		}
	case 360:
		{
			yyVAL.expr = &FuncExpr{Name: string(TIMESTAMPADD_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyS[yypt-3].bytes)}}, yyS[yypt-1].selectExprs...)}
		}
	case 361:
		{
			yyVAL.expr = &FuncExpr{Name: string(TIMESTAMPADD_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyS[yypt-3].bytes)}}, yyS[yypt-1].selectExprs...)}
		}
	case 362:
		{
			yyVAL.expr = &FuncExpr{Name: string(TIMESTAMPDIFF_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyS[yypt-3].bytes)}}, yyS[yypt-1].selectExprs...)}
		}
	case 363:
		{
			yyVAL.expr = &FuncExpr{Name: string(TIMESTAMPDIFF_BYTES), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal(yyS[yypt-3].bytes)}}, yyS[yypt-1].selectExprs...)}
		}
	case 364:
		{
			yyVAL.expr = &FuncExpr{Name: string(TRIM_BYTES), Exprs: []SelectExpr{yyS[yypt-1].selectExpr}}
		}
	case 365:
		{
			yyVAL.expr = &FuncExpr{Name: string(TRIM_BYTES), Exprs: []SelectExpr{yyS[yypt-1].selectExpr, &NonStarExpr{Expr: StrVal(yyS[yypt-4].bytes)}, yyS[yypt-3].selectExpr}}
		}
	case 366:
		{
			yyVAL.expr = &FuncExpr{Name: string(TRIM_BYTES), Exprs: []SelectExpr{yyS[yypt-1].selectExpr, &NonStarExpr{Expr: StrVal(BOTH_BYTES)}, yyS[yypt-3].selectExpr}}
		}
	case 367:
		{
			yyVAL.expr = &FuncExpr{Name: string(yyS[yypt-7].bytes), Exprs: []SelectExpr{yyS[yypt-5].selectExpr, yyS[yypt-3].selectExpr, yyS[yypt-1].selectExpr}}
		}
	case 368:
		{
			yyVAL.expr = &FuncExpr{Name: string(yyS[yypt-5].bytes), Exprs: []SelectExpr{yyS[yypt-3].selectExpr, yyS[yypt-1].selectExpr}}
		}
	case 369:
		{
			yyVAL.expr = &FuncExpr{Name: string(yyS[yypt-3].bytes), Exprs: yyS[yypt-1].selectExprs}
		}
	case 370:
		{
			yyVAL.expr = &FuncExpr{Name: string(UTC_TIMESTAMP_BYTES)}
		}
	case 371:
		{
			yyVAL.expr = &FuncExpr{Name: string(UTC_DATE_BYTES)}
		}
	case 372:
		{
			yyVAL.expr = &FuncExpr{Name: string(COUNT_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 373:
		{
			yyVAL.expr = &FuncExpr{Name: string(COUNT_BYTES), Distinct: true, Exprs: yyS[yypt-1].selectExprs}
		}
	case 374:
		{
			yyVAL.expr = &FuncExpr{Name: string(DATABASE_BYTES)}
		}
	case 375:
		{
			yyVAL.expr = &FuncExpr{Name: string(DATE_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 376:
		{
			yyVAL.expr = &FuncExpr{Name: string(DAY_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 377:
		{
			yyVAL.expr = &FuncExpr{Name: string(HOUR_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 378:
		{
			yyVAL.expr = &FuncExpr{Name: string(IF_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 379:
		{
			yyVAL.expr = &FuncExpr{Name: string(INTERVAL_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 380:
		{
			yyVAL.expr = &FuncExpr{Name: string(MICROSECOND_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 381:
		{
			yyVAL.expr = &FuncExpr{Name: string(MINUTE_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 382:
		{
			yyVAL.expr = &FuncExpr{Name: string(MOD_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 383:
		{
			yyVAL.expr = &FuncExpr{Name: string(MONTH_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 384:
		{
			yyVAL.expr = &FuncExpr{Name: string(QUARTER_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 385:
		{
			yyVAL.expr = &FuncExpr{Name: string(SCHEMA_BYTES)}
		}
	case 386:
		{
			yyVAL.expr = &FuncExpr{Name: string(SECOND_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 387:
		{
			yyVAL.expr = &FuncExpr{Name: string(TIMESTAMP_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 388:
		{
			yyVAL.expr = &FuncExpr{Name: string(WEEK_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 389:
		{
			yyVAL.expr = &FuncExpr{Name: string(USER_BYTES)}
		}
	case 390:
		{
			yyVAL.expr = &FuncExpr{Name: string(YEAR_BYTES), Exprs: yyS[yypt-1].selectExprs}
		}
	case 391:
		{
			yyVAL.expr = &FuncExpr{Name: string(bytes.ToLower(yyS[yypt-2].bytes))}
		}
	case 392:
		{
			yyVAL.expr = &FuncExpr{Name: string(bytes.ToLower(yyS[yypt-3].bytes)), Exprs: yyS[yypt-1].selectExprs}
		}
	case 393:
		{
			yyVAL.expr = &FuncExpr{Name: string(bytes.ToLower(yyS[yypt-4].bytes)), Distinct: true, Exprs: yyS[yypt-1].selectExprs}
		}
	case 396:
		{
			yyVAL.expr = StrVal("\\")
		}
	case 397:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 398:
		{
			yyVAL.expr = yyS[yypt-1].expr
		}
	case 399:
		{
			yyVAL.bytes = BOTH_BYTES
		}
	case 400:
		{
			yyVAL.bytes = LEADING_BYTES
		}
	case 401:
		{
			yyVAL.bytes = TRAILING_BYTES
		}
	case 402:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 403:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 404:
		{
			yyVAL.bytes = yyS[yypt-0].bytes
		}
	case 405:
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 406:
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 407:
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 408:
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 409:
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 410:
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 411:
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 412:
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 413:
		{
			yyVAL.bytes = YEAR_BYTES
		}
	case 414:
		{
			yyVAL.bytes = QUARTER_BYTES
		}
	case 415:
		{
			yyVAL.bytes = MONTH_BYTES
		}
	case 416:
		{
			yyVAL.bytes = WEEK_BYTES
		}
	case 417:
		{
			yyVAL.bytes = DAY_BYTES
		}
	case 418:
		{
			yyVAL.bytes = HOUR_BYTES
		}
	case 419:
		{
			yyVAL.bytes = MINUTE_BYTES
		}
	case 420:
		{
			yyVAL.bytes = SECOND_BYTES
		}
	case 421:
		{
			yyVAL.bytes = MICROSECOND_BYTES
		}
	case 422:
		{
			yyVAL.bytes = SECOND_MICROSECOND_BYTES
		}
	case 423:
		{
			yyVAL.bytes = MINUTE_MICROSECOND_BYTES
		}
	case 424:
		{
			yyVAL.bytes = MINUTE_SECOND_BYTES
		}
	case 425:
		{
			yyVAL.bytes = HOUR_MICROSECOND_BYTES
		}
	case 426:
		{
			yyVAL.bytes = HOUR_SECOND_BYTES
		}
	case 427:
		{
			yyVAL.bytes = HOUR_MINUTE_BYTES
		}
	case 428:
		{
			yyVAL.bytes = DAY_MICROSECOND_BYTES
		}
	case 429:
		{
			yyVAL.bytes = DAY_SECOND_BYTES
		}
	case 430:
		{
			yyVAL.bytes = DAY_MINUTE_BYTES
		}
	case 431:
		{
			yyVAL.bytes = DAY_HOUR_BYTES
		}
	case 432:
		{
			yyVAL.bytes = YEAR_MONTH_BYTES
		}
	case 433:
		{
			yyVAL.bytes = BINARY_BYTES
		}
	case 434:
		{
			yyVAL.bytes = BINARY_BYTES
		}
	case 435:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 436:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 437:
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 438:
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 439:
		{
			yyVAL.bytes = DECIMAL_BYTES
		}
	case 440:
		{
			yyVAL.bytes = DECIMAL_BYTES
		}
	case 441:
		{
			yyVAL.bytes = DECIMAL_BYTES
		}
	case 442:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 443:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 444:
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 445:
		{
			yyVAL.bytes = SIGNED_BYTES
		}
	case 446:
		{
			yyVAL.bytes = OBJECT_ID_BYTES
		}
	case 447:
		{
			yyVAL.bytes = SIGNED_BYTES
		}
	case 448:
		{
			yyVAL.bytes = SIGNED_BYTES
		}
	case 449:
		{
			yyVAL.bytes = TIME_BYTES
		}
	case 450:
		{
			yyVAL.bytes = UNSIGNED_BYTES
		}
	case 451:
		{
			yyVAL.bytes = UNSIGNED_BYTES
		}
	case 452:
		{
			yyVAL.bytes = SIGNED_BYTES
		}
	case 453:
		{
			yyVAL.bytes = CHAR_BYTES
		}
	case 454:
		{
			yyVAL.bytes = DATE_BYTES
		}
	case 455:
		{
			yyVAL.bytes = DATETIME_BYTES
		}
	case 456:
		{
			yyVAL.bytes = FLOAT_BYTES
		}
	case 457:
		{
			yyVAL.byt = AST_UPLUS
		}
	case 458:
		{
			yyVAL.byt = AST_UMINUS
		}
	case 459:
		{
			yyVAL.byt = AST_TILDA
		}
	case 460:
		{
			yyVAL.caseExpr = &CaseExpr{Expr: yyS[yypt-3].expr, Whens: yyS[yypt-2].whens, Else: yyS[yypt-1].expr}
		}
	case 461:
		{
			yyVAL.expr = nil
		}
	case 462:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 463:
		{
			yyVAL.whens = []*When{yyS[yypt-0].when}
		}
	case 464:
		{
			yyVAL.whens = append(yyS[yypt-1].whens, yyS[yypt-0].when)
		}
	case 465:
		{
			yyVAL.when = &When{Cond: yyS[yypt-2].expr, Val: yyS[yypt-0].expr}
		}
	case 466:
		{
			yyVAL.expr = nil
		}
	case 467:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 468:
		{
			yyVAL.colName = &ColName{Name: yyS[yypt-0].str}
		}
	case 469:
		{
			yyVAL.colName = &ColName{Qualifier: option.SomeString(yyS[yypt-2].str), Name: yyS[yypt-0].str}
		}
	case 470:
		{
			yyVAL.colName = &ColName{Database: option.SomeString(yyS[yypt-4].str), Qualifier: option.SomeString(yyS[yypt-2].str), Name: yyS[yypt-0].str}
		}
	case 471:
		{
			yyVAL.expr = StrVal(yyS[yypt-0].bytes)
		}
	case 472:
		{
			yyVAL.expr = NumVal(yyS[yypt-0].bytes)
		}
	case 473:
		{
			yyVAL.expr = ValArg(yyS[yypt-0].bytes)
		}
	case 474:
		{
			yyVAL.expr = &DateVal{Name: AST_DATE, Val: string(yyS[yypt-0].bytes)}
		}
	case 475:
		{
			yyVAL.expr = &DateVal{Name: AST_TIME, Val: string(yyS[yypt-0].bytes)}
		}
	case 476:
		{
			yyVAL.expr = &DateVal{Name: AST_TIMESTAMP, Val: string(yyS[yypt-0].bytes)}
		}
	case 477:
		{
			if bytes.Equal(bytes.ToLower(yyS[yypt-2].bytes), D_BYTES) {
				yyVAL.expr = &DateVal{Name: AST_DATE, Val: string(yyS[yypt-1].bytes)}
			} else if bytes.Equal(bytes.ToLower(yyS[yypt-2].bytes), T_BYTES) {
				yyVAL.expr = &DateVal{Name: AST_TIME, Val: string(yyS[yypt-1].bytes)}
			} else if bytes.Equal(bytes.ToLower(yyS[yypt-2].bytes), TS_BYTES) {
				yyVAL.expr = &DateVal{Name: AST_TIMESTAMP, Val: string(yyS[yypt-1].bytes)}
			} else {
				yylex.Error("expecting d, t, or ts")
				return 1
			}
		}
	case 478:
		{
			yyVAL.expr = &NullVal{}
		}
	case 479:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 480:
		{
			yyVAL.expr = &TrueVal{}
		}
	case 481:
		{
			yyVAL.expr = &FalseVal{}
		}
	case 482:
		{
			yyVAL.expr = &UnknownVal{}
		}
	case 483:
		{
			yyVAL.exprs = nil
		}
	case 484:
		{
			yyVAL.exprs = yyS[yypt-0].exprs
		}
	case 485:
		{
			yyVAL.expr = nil
		}
	case 486:
		{
			yyVAL.expr = yyS[yypt-0].expr
		}
	case 487:
		{
			yyVAL.orderBy = nil
		}
	case 488:
		{
			yyVAL.orderBy = yyS[yypt-0].orderBy
		}
	case 489:
		{
			yyVAL.orderBy = OrderBy{yyS[yypt-0].order}
		}
	case 490:
		{
			yyVAL.orderBy = append(yyS[yypt-2].orderBy, yyS[yypt-0].order)
		}
	case 491:
		{
			yyVAL.order = &Order{Expr: yyS[yypt-1].expr, Direction: yyS[yypt-0].str}
		}
	case 492:
		{
			yyVAL.str = AST_ASC
		}
	case 493:
		{
			yyVAL.str = AST_ASC
		}
	case 494:
		{
			yyVAL.str = AST_DESC
		}
	case 495:
		{
			yyVAL.limit = nil
		}
	case 496:
		{
			yyVAL.limit = &Limit{Rowcount: yyS[yypt-0].expr}
		}
	case 497:
		{
			yyVAL.limit = &Limit{Offset: yyS[yypt-2].expr, Rowcount: yyS[yypt-0].expr}
		}
	case 498:
		{
			yyVAL.limit = &Limit{Offset: yyS[yypt-0].expr, Rowcount: yyS[yypt-2].expr}
		}
	case 499:
		{
			yyVAL.str = ""
		}
	case 500:
		{
			yyVAL.str = AST_FOR_UPDATE
		}
	case 501:
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
	case 502:
		{
			yyVAL.str = string(yyS[yypt-0].bytes)
		}
	case 503:
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 504:
		{
			yyVAL.str = yyS[yypt-0].str
		}
	case 505:
		{
			yyVAL.str = string(yyS[yypt-0].bytes)
		}
	case 506:
		{
			yyVAL.str = string(ANY_BYTES)
		}
	case 507:
		{
			yyVAL.str = string(BINLOG_BYTES)
		}
	case 508:
		{
			yyVAL.str = string(CHANNEL_BYTES)
		}
	case 509:
		{
			yyVAL.str = string(CHARSET_BYTES)
		}
	case 510:
		{
			yyVAL.str = string(CODE_BYTES)
		}
	case 511:
		{
			yyVAL.str = string(COLLATION_BYTES)
		}
	case 512:
		{
			yyVAL.str = string(COLUMNS_BYTES)
		}
	case 513:
		{
			yyVAL.str = string(COMMITTED_BYTES)
		}
	case 514:
		{
			yyVAL.str = string(CONNECTION_BYTES)
		}
	case 515:
		{
			yyVAL.str = string(COUNT_BYTES)
		}
	case 516:
		{
			yyVAL.str = string(DATE_BYTES)
		}
	case 517:
		{
			yyVAL.str = string(DATETIME_BYTES)
		}
	case 518:
		{
			yyVAL.str = string(DAY_BYTES)
		}
	case 519:
		{
			yyVAL.str = string(DECIMAL_BYTES)
		}
	case 520:
		{
			yyVAL.str = string(ENGINE_BYTES)
		}
	case 521:
		{
			yyVAL.str = string(ENGINES_BYTES)
		}
	case 522:
		{
			yyVAL.str = string(ERRORS_BYTES)
		}
	case 523:
		{
			yyVAL.str = string(EVENT_BYTES)
		}
	case 524:
		{
			yyVAL.str = string(EVENTS_BYTES)
		}
	case 525:
		{
			yyVAL.str = string(EXTENDED_BYTES)
		}
	case 526:
		{
			yyVAL.str = string(FLOAT_BYTES)
		}
	case 527:
		{
			yyVAL.str = string(FORMAT_BYTES)
		}
	case 528:
		{
			yyVAL.str = string(FULL_BYTES)
		}
	case 529:
		{
			yyVAL.str = string(FUNCTION_BYTES)
		}
	case 530:
		{
			yyVAL.str = string(GRANTS_BYTES)
		}
	case 531:
		{
			yyVAL.str = string(HOSTS_BYTES)
		}
	case 532:
		{
			yyVAL.str = string(HOUR_BYTES)
		}
	case 533:
		{
			yyVAL.str = string(INDEXES_BYTES)
		}
	case 534:
		{
			yyVAL.str = string(ISOLATION_BYTES)
		}
	case 535:
		{
			yyVAL.str = string(JSON_BYTES)
		}
	case 536:
		{
			yyVAL.str = string(LEVEL_BYTES)
		}
	case 537:
		{
			yyVAL.str = string(LOGS_BYTES)
		}
	case 538:
		{
			yyVAL.str = string(MASTER_BYTES)
		}
	case 539:
		{
			yyVAL.str = string(MICROSECOND_BYTES)
		}
	case 540:
		{
			yyVAL.str = string(MINUTE_BYTES)
		}
	case 541:
		{
			yyVAL.str = string(MONTH_BYTES)
		}
	case 542:
		{
			yyVAL.str = string(MUTEX_BYTES)
		}
	case 543:
		{
			yyVAL.str = string(NAMES_BYTES)
		}
	case 544:
		{
			yyVAL.str = string(NCHAR_BYTES)
		}
	case 545:
		{
			yyVAL.str = string(NUMBER_BYTES)
		}
	case 546:
		{
			yyVAL.str = string(OFFSET_BYTES)
		}
	case 547:
		{
			yyVAL.str = string(OBJECT_ID_BYTES)
		}
	case 548:
		{
			yyVAL.str = string(ONLY_BYTES)
		}
	case 549:
		{
			yyVAL.str = string(OPEN_BYTES)
		}
	case 550:
		{
			yyVAL.str = string(PARTITIONS_BYTES)
		}
	case 551:
		{
			yyVAL.str = string(PLUGINS_BYTES)
		}
	case 552:
		{
			yyVAL.str = string(PRIVILEGES_BYTES)
		}
	case 553:
		{
			yyVAL.str = string(PROCESSLIST_BYTES)
		}
	case 554:
		{
			yyVAL.str = string(PROFILE_BYTES)
		}
	case 555:
		{
			yyVAL.str = string(PROFILES_BYTES)
		}
	case 556:
		{
			yyVAL.str = string(PROXY_BYTES)
		}
	case 557:
		{
			yyVAL.str = string(QUARTER_BYTES)
		}
	case 558:
		{
			yyVAL.str = string(QUERY_BYTES)
		}
	case 559:
		{
			yyVAL.str = string(RELAYLOG_BYTES)
		}
	case 560:
		{
			yyVAL.str = string(REPEATABLE_BYTES)
		}
	case 561:
		{
			yyVAL.str = string(ROW_BYTES)
		}
	case 562:
		{
			yyVAL.str = string(SECOND_BYTES)
		}
	case 563:
		{
			yyVAL.str = string(SERIALIZABLE_BYTES)
		}
	case 564:
		{
			yyVAL.str = string(SIGNED_BYTES)
		}
	case 565:
		{
			yyVAL.str = string(SLAVE_BYTES)
		}
	case 566:
		{
			yyVAL.str = string(SOME_BYTES)
		}
	case 567:
		{
			yyVAL.str = string(SQL_TSI_DAY_BYTES)
		}
	case 568:
		{
			yyVAL.str = string(SQL_TSI_HOUR_BYTES)
		}
	case 569:
		{
			yyVAL.str = string(SQL_TSI_MINUTE_BYTES)
		}
	case 570:
		{
			yyVAL.str = string(SQL_TSI_MONTH_BYTES)
		}
	case 571:
		{
			yyVAL.str = string(SQL_TSI_QUARTER_BYTES)
		}
	case 572:
		{
			yyVAL.str = string(SQL_TSI_SECOND_BYTES)
		}
	case 573:
		{
			yyVAL.str = string(SQL_TSI_WEEK_BYTES)
		}
	case 574:
		{
			yyVAL.str = string(SQL_TSI_YEAR_BYTES)
		}
	case 575:
		{
			yyVAL.str = string(STATUS_BYTES)
		}
	case 576:
		{
			yyVAL.str = string(STORAGE_BYTES)
		}
	case 577:
		{
			yyVAL.str = string(TABLES_BYTES)
		}
	case 578:
		{
			yyVAL.str = string(TEMPORARY_BYTES)
		}
	case 579:
		{
			yyVAL.str = string(TIME_BYTES)
		}
	case 580:
		{
			yyVAL.str = string(TIMESTAMP_BYTES)
		}
	case 581:
		{
			yyVAL.str = string(TIMESTAMPADD_BYTES)
		}
	case 582:
		{
			yyVAL.str = string(TIMESTAMPDIFF_BYTES)
		}
	case 583:
		{
			yyVAL.str = string(TRANSACTION_BYTES)
		}
	case 584:
		{
			yyVAL.str = string(TRIGGERS_BYTES)
		}
	case 585:
		{
			yyVAL.str = string(UNCOMMITTED_BYTES)
		}
	case 586:
		{
			yyVAL.str = string(UNKNOWN_BYTES)
		}
	case 587:
		{
			yyVAL.str = string(USER_BYTES)
		}
	case 588:
		{
			yyVAL.str = string(VARIABLES_BYTES)
		}
	case 589:
		{
			yyVAL.str = string(VIEW_BYTES)
		}
	case 590:
		{
			yyVAL.str = string(WARNINGS_BYTES)
		}
	case 591:
		{
			yyVAL.str = string(WEEK_BYTES)
		}
	case 592:
		{
			yyVAL.str = string(YEAR_BYTES)
		}

	}

	if yyEx != nil && yyEx.Reduced(r, exState, &yyVAL) {
		return -1
	}
	goto yystack /* stack new state and value */
}
