// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

%{
package parser

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

%}

%union {
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

%token LEX_ERROR
%token <empty> SELECT INSERT UPDATE DELETE WHERE GROUP HAVING ORDER BY LIMIT OFFSET FOR SOME ANY TRUE FALSE UNKNOWN
%token <empty> ALL DISTINCT PRECISION AS EXISTS NULL ASC DESC VALUES INTO DUPLICATE KEY DEFAULT SET LOCK
%token <bytes> ID STRING NUMBER VALUE_ARG COMMENT
%token <empty> LPAREN RPAREN LBRACE RBRACE TILDE
%token <empty> DATE DATETIME TIME TIMESTAMP CURRENT_TIMESTAMP
%token <empty> TIMESTAMPADD TIMESTAMPDIFF YEAR QUARTER MONTH WEEK DAY HOUR MINUTE SECOND MICROSECOND EXTRACT DATE_ADD
%token <empty> DATE_SUB INTERVAL STR_TO_DATE
%token <empty> SQL_TSI_YEAR SQL_TSI_QUARTER SQL_TSI_MONTH SQL_TSI_WEEK SQL_TSI_DAY SQL_TSI_HOUR SQL_TSI_MINUTE SQL_TSI_SECOND
%token <empty> CONVERT CAST CHAR SIGNED UNSIGNED SQL_BIGINT SQL_VARCHAR SQL_DATE SQL_TIMESTAMP SQL_DOUBLE INTEGER
%token <empty> SECOND_MICROSECOND MINUTE_MICROSECOND MINUTE_SECOND HOUR_MICROSECOND HOUR_SECOND HOUR_MINUTE DAY_MICROSECOND DAY_SECOND
%token <empty> DAY_MINUTE DAY_HOUR YEAR_MONTH

%nonassoc <empty> FROM
%left <empty> UNION MINUS EXCEPT INTERSECT
%left <empty> COMMA
%left <empty> JOIN STRAIGHT_JOIN LEFT RIGHT INNER OUTER CROSS NATURAL USE FORCE
%left <empty> ON
%right <empty> NOT
%left <empty> OR XOR
%left <empty> AND
%nonassoc <empty> BETWEEN
%left <empty> NE EQ NULL_SAFE_EQUAL IS LIKE REGEXP IN
%left <empty> LT GT LE GE
%left <empty> BIT_AND BIT_OR CARET
%left <empty> PLUS SUB
%left <empty> TIMES MOD DIV IDIV
%nonassoc <empty> DOT
%left <empty> UNARY
%right <empty> CASE, WHEN, THEN, ELSE
%left <empty> END

// Transaction Tokens
%token <bytes> BEGIN COMMIT ROLLBACK
%token <bytes> TRANSACTION ISOLATION LEVEL
%token <bytes> READ WRITE ONLY
%token <bytes> REPEATABLE COMMITTED UNCOMMITTED SERIALIZABLE

// Charset Tokens
%token <empty> NAMES
%token <empty> CHARACTER
%token <empty> COLLATE

// Replace
%token <empty> REPLACE

// Mixer admin
%token <empty> ADMIN

// Show
%token <empty> SHOW
%token <empty> DATABASES TABLES PROXY VARIABLES FULL COLUMNS COLLATION PROCESSLIST STATUS CHARSET

// Explain
%token <empty> EXPLAIN DESCRIBE
%token <empty> EXTENDED PARTITIONS FORMAT TRADITIONAL JSON

// Kill
%token <empty> KILL
%token <empty> CONNECTION QUERY

// Scope
%token <empty> SESSION GLOBAL 

// DDL Tokens
%token <empty> CREATE ALTER DROP RENAME
%token <empty> TABLE INDEX VIEW TO IGNORE IF UNIQUE USING

%start any_command

%type <statement> command
%type <selStmt> select_statement
%type <statement> insert_statement update_statement delete_statement set_statement
%type <statement> create_statement alter_statement rename_statement drop_statement
%type <bytes2> comment_opt comment_list
%type <str> union_op
%type <str> all_any_some
%type <str> distinct_opt
%type <selectExprs> select_expression_list
%type <selectExpr> select_expression
%type <bytes> as_lower_opt as_opt
%type <expr> expression expr
%type <tableExprs> table_expression_list
%type <tableExpr> table_expression
%type <str> join_type
%type <smTableExpr> simple_table_expression
%type <tableName> dml_table_expression table_name
%type <indexHints> index_hint_list
%type <bytes2> index_list
%type <expr> where_expression_opt
%type <insRows> row_list
%type <expr> value
%type <tuple> tuple
%type <expr> boolean_value
%type <exprs> expression_list
%type <values> tuple_list
%type <bytes> keyword_as_func
%type <bytes> time_interval
%type <bytes> sql_time_interval
%type <bytes> sql_time_unit
%type <bytes> sql_types
%type <subquery> subquery
%type <byt> unary_operator
%type <colName> column_name column_name_opt
%type <caseExpr> case_expression
%type <whens> when_expression_list
%type <when> when_expression
%type <expr> expression_opt else_expression_opt
%type <exprs> group_by_opt
%type <expr> having_opt
%type <orderBy> order_by_opt order_list
%type <order> order
%type <str> asc_desc_opt
%type <limit> limit_opt
%type <str> lock_opt
%type <columns> column_list_opt column_list
%type <updateExprs> on_dup_opt
%type <updateExprs> update_list
%type <updateExpr> update_expression
%type <bytes> transaction_characteristics
%type <bytes> transaction_characteristic
%type <bytes> transaction_level
%type <empty> exists_opt not_exists_opt ignore_opt non_rename_operation to_opt constraint_opt using_opt
%type <bytes> sql_id
%type <empty> force_eof
%type <empty> explain_alias

%type <statement> begin_statement commit_statement rollback_statement
%type <statement> replace_statement
%type <statement> show_statement
%type <statement> kill_statement
%type <statement> admin_statement
%type <statement> explain_statement
%type <statement> explainable_stmt

%type <expr> from_opt
%type <expr> like_or_where_opt
%type <expr> show_from_in show_from_in_opt
%type <str> show_full
%type <str> scope_modifier_opt
%type <str> explain_type
%type <str> format_name
%type <str> kill_modifier
%%

any_command:
  command
  {
    SetParseTree(yylex, $1)
  }

command:
  select_statement
  {
    $$ = $1
  }
| insert_statement
| update_statement
| delete_statement
| set_statement
| create_statement
| alter_statement
| rename_statement
| drop_statement
| begin_statement
| commit_statement
| kill_statement
| rollback_statement
| replace_statement
| show_statement
| admin_statement
| explain_statement

select_statement:
  SELECT comment_opt distinct_opt select_expression_list
  {
    $$ = &SimpleSelect{Comments: Comments($2), Distinct: $3, SelectExprs: $4}
  }
| SELECT comment_opt distinct_opt select_expression_list FROM table_expression_list where_expression_opt group_by_opt having_opt order_by_opt limit_opt lock_opt
  {
    $$ = &Select{Comments: Comments($2), Distinct: $3, SelectExprs: $4, From: $6, Where: NewWhere(AST_WHERE, $7), GroupBy: GroupBy($8), Having: NewWhere(AST_HAVING, $9), OrderBy: $10, Limit: $11, Lock: $12}
  }
| select_statement union_op select_statement %prec UNION
  {
    $$ = &Union{Type: $2, Left: $1, Right: $3}
  }


insert_statement:
  INSERT comment_opt INTO dml_table_expression column_list_opt row_list on_dup_opt
  {
    $$ = &Insert{Comments: Comments($2), Table: $4, Columns: $5, Rows: $6, OnDup: OnDup($7)}
  }
| INSERT comment_opt INTO dml_table_expression SET update_list on_dup_opt
  {
    cols := make(Columns, 0, len($6))
    vals := make(ValTuple, 0, len($6))
    for _, col := range $6 {
      cols = append(cols, &NonStarExpr{Expr: col.Name})
      vals = append(vals, col.Expr)
    }
    $$ = &Insert{Comments: Comments($2), Table: $4, Columns: cols, Rows: Values{vals}, OnDup: OnDup($7)}
  }

replace_statement:
  REPLACE comment_opt INTO dml_table_expression column_list_opt row_list
  {
    $$ = &Replace{Comments: Comments($2), Table: $4, Columns: $5, Rows: $6}
  }
| REPLACE comment_opt INTO dml_table_expression SET update_list
  {
    cols := make(Columns, 0, len($6))
    vals := make(ValTuple, 0, len($6))
    for _, col := range $6 {
      cols = append(cols, &NonStarExpr{Expr: col.Name})
      vals = append(vals, col.Expr)
    }
    $$ = &Replace{Comments: Comments($2), Table: $4, Columns: cols, Rows: Values{vals}}
  }


update_statement:
  UPDATE comment_opt dml_table_expression SET update_list where_expression_opt order_by_opt limit_opt
  {
    $$ = &Update{Comments: Comments($2), Table: $3, Exprs: $5, Where: NewWhere(AST_WHERE, $6), OrderBy: $7, Limit: $8}
  }

delete_statement:
  DELETE comment_opt FROM dml_table_expression where_expression_opt order_by_opt limit_opt
  {
    $$ = &Delete{Comments: Comments($2), Table: $4, Where: NewWhere(AST_WHERE, $5), OrderBy: $6, Limit: $7}
  }

set_statement:
  SET comment_opt update_list
  {
    $$ = &Set{Comments: Comments($2), Exprs: $3}
  }
| SET comment_opt NAMES ID
  {
    $$ = &Set{Comments: Comments($2), Exprs: UpdateExprs{
      &UpdateExpr{Name: &ColName{Name:[]byte("@@character_set_client")}, Expr: StrVal($4)},
      &UpdateExpr{Name: &ColName{Name:[]byte("@@character_set_results")}, Expr: StrVal($4)},
      &UpdateExpr{Name: &ColName{Name:[]byte("@@character_set_connection")}, Expr: StrVal($4)},
    }}
  }
| SET comment_opt NAMES ID COLLATE ID
  {
    $$ = &Set{Comments: Comments($2), Exprs: UpdateExprs{
      &UpdateExpr{Name: &ColName{Name:[]byte("@@character_set_client")}, Expr: StrVal($4)},
      &UpdateExpr{Name: &ColName{Name:[]byte("@@character_set_results")}, Expr: StrVal($4)},
      &UpdateExpr{Name: &ColName{Name:[]byte("@@character_set_connection")}, Expr: StrVal($4)},
      &UpdateExpr{Name: &ColName{Name:[]byte("@@collation_connection")}, Expr: StrVal($6)},
    }}
  }
| SET comment_opt CHARACTER SET ID
  {
    $$ = &Set{Comments: Comments($2), Exprs: UpdateExprs{
      &UpdateExpr{Name: &ColName{Name:[]byte("@@character_set_client")}, Expr: StrVal($5)},
      &UpdateExpr{Name: &ColName{Name:[]byte("@@character_set_results")}, Expr: StrVal($5)},
      &UpdateExpr{Name: &ColName{Name:[]byte("@@collation_connection")}, Expr: &ColName{Name:[]byte("@@collation_database")}},
    }}
  }
| SET scope_modifier_opt TRANSACTION transaction_characteristics
  {
    $$ = &Set{Comments: append([][]byte{}, []byte($2), []byte("transaction"), $4)}
  }

transaction_characteristics:
  transaction_characteristic
  {
    $$ = $1
  }
| transaction_characteristics COMMA transaction_characteristic
  {
    $$ = append($1, append([]byte(", "), $3...)...)
  }

transaction_characteristic:
  ISOLATION LEVEL transaction_level
  {
    $$ = append([]byte("isolation level "), $3...)
  }
| READ WRITE
  {
    $$ = []byte("read write")
  }
| READ ONLY
  {
    $$ = []byte("read only")
  }

transaction_level:
  REPEATABLE READ
  {
    $$ = []byte("repeatable read")
  }
| READ COMMITTED
  {
    $$ = []byte("read committed")
  }
| READ UNCOMMITTED
  {
    $$ = []byte("read uncommitted")
  }
| SERIALIZABLE
  {
    $$ = []byte("serializable")
  }

begin_statement:
  BEGIN
  {
    $$ = &Begin{}
  }

commit_statement:
  COMMIT
  {
    $$ = &Commit{}
  }

rollback_statement:
  ROLLBACK
  {
    $$ = &Rollback{}
  }

admin_statement:
  ADMIN sql_id LPAREN expression_list RPAREN
  {
    $$ = &Admin{Name : $2, Values : $4}
  }

show_from_in:
  FROM expression
  {
    $$ = $2
  }
| IN expression
  {
    $$ = $2
  }

show_from_in_opt:
  {
    $$ = nil
  }
| FROM expression
  {
    $$ = $2
  }
| IN expression
  {
    $$ = $2
  }

show_full:
  {
    $$ = AST_SHOW_NO_MOD
  }
| FULL
  {
    $$ = AST_SHOW_FULL
  }

kill_modifier:
  CONNECTION
  {
    $$ = AST_KILL_CONNECTION
  }
| QUERY
  {
    $$ = AST_KILL_QUERY
  }

scope_modifier_opt:
  {
    $$ = AST_SESSION_SCOPE
  }
| SESSION
  {
    $$ = AST_SESSION_SCOPE
  }
| GLOBAL
  {
    $$ = AST_GLOBAL_SCOPE
  }


show_statement:
  SHOW DATABASES like_or_where_opt
  {
    $$ = &Show{Section: "databases", LikeOrWhere: $3}
  }
| SHOW scope_modifier_opt VARIABLES like_or_where_opt
  {
    $$ = &Show{Section: "variables", Modifier: $2, LikeOrWhere: $4}
  }
| SHOW TABLES from_opt like_or_where_opt
  {
    $$ = &Show{Section: "tables", From: $3, LikeOrWhere: $4}
  }
| SHOW PROXY sql_id from_opt like_or_where_opt
  {
    $$ = &Show{Section: "proxy", Key: string($3), From: $4, LikeOrWhere: $5}
  }
| SHOW show_full COLUMNS show_from_in show_from_in_opt
  {
    $$ = &Show{Section: "columns", From: $4, Modifier: $2, DBFilter: $5}
  }
| SHOW show_full PROCESSLIST
  {
    $$ = &Show{Section: "processlist", Modifier: $2}
  }
| SHOW scope_modifier_opt STATUS like_or_where_opt
  {
    $$ = &Show{Section: "status", Modifier: $2, LikeOrWhere: $4}
  }
| SHOW CHARACTER SET like_or_where_opt
  {
    $$ = &Show{Section: "charset", LikeOrWhere: $4}
  }
| SHOW CHARSET like_or_where_opt
  {
    $$ = &Show{Section: "charset", LikeOrWhere: $3}
  }
| SHOW TABLE STATUS show_from_in_opt like_or_where_opt
  {
    $$ = &Show{Section: "tablestatus", From: $4, LikeOrWhere: $5}
  }
| SHOW COLLATION like_or_where_opt
  {
    $$ = &Show{Section: "collation", LikeOrWhere: $3}
  }

format_name:
  TRADITIONAL
  {
    $$ = AST_EXPLAIN_FORMAT_TRADITIONAL
  }
| JSON
  {
    $$ = AST_EXPLAIN_FORMAT_JSON
  }

explain_type:
  EXTENDED
  {
    $$ = AST_EXPLAIN_EXTENDED
  }
| PARTITIONS
  {
    $$ = AST_EXPLAIN_PARTITIONS
  }
| FORMAT EQ format_name
  {
    $$ = $3
  } 

explainable_stmt:
  select_statement
  {
    $$ = $1
  }
| LPAREN explainable_stmt RPAREN
  {
    $$ = $2
  }

explain_alias:
  EXPLAIN
  {
    $$ = $1
  }
| DESCRIBE
  {
    $$ = $1
  }
| DESC
  {
    $$ = $1
  }

table_name:
  sql_id
  {
    $$ = &TableName{Name: $1}
  }
  
explain_statement:
  explain_alias table_name column_name_opt
  {
    $$ = &Explain{Section: "table", Table: $2, Column: $3}
  }
| explain_alias explain_type explainable_stmt
  {
    $$ = &Explain{Section: "plan", ExplainType: $2, Statement: $3}
  }
| explain_alias explain_type FOR CONNECTION NUMBER
  {
    $$ = &Explain{Section: "plan", ExplainType: $2, Connection: $5}
  }

kill_statement:
  KILL expression
  {
    $$ = &Kill{Scope: AST_KILL_CONNECTION, ID: $2}
  }
| KILL kill_modifier expression
  {
    $$ = &Kill{Scope: $2, ID: $3}
  }

create_statement:
  CREATE TABLE not_exists_opt ID force_eof
  {
    $$ = &DDL{Action: AST_CREATE, NewName: $4}
  }
| CREATE constraint_opt INDEX sql_id using_opt ON ID force_eof
  {
    // Change this to an alter statement
    $$ = &DDL{Action: AST_ALTER, Table: $7, NewName: $7}
  }
| CREATE VIEW sql_id force_eof
  {
    $$ = &DDL{Action: AST_CREATE, NewName: $3}
  }

alter_statement:
  ALTER ignore_opt TABLE ID non_rename_operation force_eof
  {
    $$ = &DDL{Action: AST_ALTER, Table: $4, NewName: $4}
  }
| ALTER ignore_opt TABLE ID RENAME to_opt ID
  {
    // Change this to a rename statement
    $$ = &DDL{Action: AST_RENAME, Table: $4, NewName: $7}
  }
| ALTER VIEW sql_id force_eof
  {
    $$ = &DDL{Action: AST_ALTER, Table: $3, NewName: $3}
  }

rename_statement:
  RENAME TABLE ID TO ID
  {
    $$ = &DDL{Action: AST_RENAME, Table: $3, NewName: $5}
  }

drop_statement:
  DROP TABLE exists_opt ID
  {
    $$ = &DDL{Action: AST_DROP, Table: $4}
  }
| DROP INDEX sql_id ON ID
  {
    // Change this to an alter statement
    $$ = &DDL{Action: AST_ALTER, Table: $5, NewName: $5}
  }
| DROP VIEW exists_opt sql_id force_eof
  {
    $$ = &DDL{Action: AST_DROP, Table: $4}
  }

comment_opt:
  {
    SetAllowComments(yylex, true)
  }
  comment_list
  {
    $$ = $2
    SetAllowComments(yylex, false)
  }

comment_list:
  {
    $$ = nil
  }
| comment_list COMMENT
  {
    $$ = append($1, $2)
  }

union_op:
  UNION
  {
    $$ = AST_UNION
  }
| UNION ALL
  {
    $$ = AST_UNION_ALL
  }
| MINUS
  {
    $$ = AST_SET_MINUS
  }
| EXCEPT
  {
    $$ = AST_EXCEPT
  }
| INTERSECT
  {
    $$ = AST_INTERSECT
  }

distinct_opt:
  {
    $$ = ""
  }
| DISTINCT
  {
    $$ = AST_DISTINCT
  }

select_expression_list:
  select_expression
  {
    $$ = SelectExprs{$1}
  }
| select_expression_list COMMA select_expression
  {
    $$ = append($$, $3)
  }

select_expression:
  TIMES
  {
    $$ = &StarExpr{}
  }
| expression as_lower_opt
  {
    $$ = &NonStarExpr{Expr: $1, As: $2}
  }
| expression as_lower_opt PRECISION
  {
    $$ = &NonStarExpr{Expr: $1, As: $2}
  }
| ID DOT TIMES
  {
    $$ = &StarExpr{TableName: $1}
  }

as_lower_opt:
  {
    $$ = nil
  }
| sql_id
  {
    $$ = $1
  }
| AS sql_id
  {
    $$ = $2
  }

table_expression_list:
  table_expression
  {
    $$ = TableExprs{$1}
  }
| table_expression_list COMMA table_expression
  {
    $$ = append($$, $3)
  }

table_expression:
  simple_table_expression as_opt index_hint_list
  {
    $$ = &AliasedTableExpr{Expr:$1, As: $2, Hints: $3}
  }
| LPAREN table_expression RPAREN
  {
    $$ = &ParenTableExpr{Expr: $2}
  }
| table_expression join_type table_expression %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3}
  }
| table_expression join_type table_expression ON expression %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3, On: $5}
  }

as_opt:
  {
    $$ = nil
  }
| ID
  {
    $$ = $1
  }
| AS ID
  {
    $$ = $2
  }

join_type:
  JOIN
  {
    $$ = AST_JOIN
  }
| STRAIGHT_JOIN
  {
    $$ = AST_STRAIGHT_JOIN
  }
| LEFT JOIN
  {
    $$ = AST_LEFT_JOIN
  }
| LEFT OUTER JOIN
  {
    $$ = AST_LEFT_JOIN
  }
| RIGHT JOIN
  {
    $$ = AST_RIGHT_JOIN
  }
| RIGHT OUTER JOIN
  {
    $$ = AST_RIGHT_JOIN
  }
| INNER JOIN
  {
    $$ = AST_JOIN
  }
| CROSS JOIN
  {
    $$ = AST_CROSS_JOIN
  }
| NATURAL JOIN
  {
    $$ = AST_NATURAL_JOIN
  }

simple_table_expression:
ID
  {
    $$ = &TableName{Name: $1}
  }
| ID DOT ID
  {
    $$ = &TableName{Qualifier: $1, Name: $3}
  }
| subquery
  {
    $$ = $1
  }
| COLUMNS // hack for tokenizer, maybe cleaner way
  {
    $$ = &TableName{Name: []byte("columns")}
  }
| TABLES // hack for tokenizer, maybe cleaner way
  {
    $$ = &TableName{Name: []byte("tables")}
  }
| ID DOT COLUMNS // hack for tokenizer, maybe cleaner way
  {
    $$ = &TableName{Qualifier: $1, Name: []byte("columns")}
  }
| ID DOT TABLES // hack for tokenizer, maybe cleaner way
  {
    $$ = &TableName{Qualifier: $1, Name: []byte("tables")}
  }

dml_table_expression:
ID
  {
    $$ = &TableName{Name: $1}
  }
| ID DOT ID
  {
    $$ = &TableName{Qualifier: $1, Name: $3}
  }

index_hint_list:
  {
    $$ = nil
  }
| USE INDEX LPAREN index_list RPAREN
  {
    $$ = &IndexHints{Type: AST_USE, Indexes: $4}
  }
| IGNORE INDEX LPAREN index_list RPAREN
  {
    $$ = &IndexHints{Type: AST_IGNORE, Indexes: $4}
  }
| FORCE INDEX LPAREN index_list RPAREN
  {
    $$ = &IndexHints{Type: AST_FORCE, Indexes: $4}
  }

index_list:
  sql_id
  {
    $$ = [][]byte{$1}
  }
| index_list COMMA sql_id
  {
    $$ = append($1, $3)
  }

where_expression_opt:
  {
    $$ = nil
  }
| WHERE expression
  {
    $$ = $2
  }

like_or_where_opt:
  {
    $$ = nil
  }
| WHERE expression
  {
    $$ = $2
  }
| LIKE expression
  {
    $$ = $2
  }

from_opt:
  {
    $$ = nil
  }
| FROM expression
  {
    $$ = $2
  }

all_any_some:
ALL
  {
    $$ = AST_ALL
  }
| SOME
  {
    $$ = AST_SOME
  }
| ANY
  {
    $$ = AST_ANY
  }

row_list:
  VALUES tuple_list
  {
    $$ = $2
  }
| select_statement
  {
    $$ = $1
  }

tuple_list:
  tuple
  {
    $$ = Values{$1}
  }
| tuple_list COMMA tuple
  {
    $$ = append($1, $3)
  }

tuple:
  LPAREN expression_list RPAREN
  {
    $$ = ValTuple($2)
  }
| subquery
  {
    $$ = $1
  }

subquery:
  LPAREN select_statement RPAREN
  {
    $$ = &Subquery{$2}
  }

expression_list:
  expression
  {
    $$ = Exprs{$1}
  }
| expression_list COMMA expression
  {
    $$ = append($1, $3)
  }

expression:
  expr %prec BETWEEN
    {
      $$ = $1
    }
  | expression AND expression %prec NOT
    {
      $$ = &AndExpr{Left: $1, Right: $3}
    }

expr:
  value
  {
    $$ = $1
  }
| column_name
  {
    $$ = $1
  }
| tuple
  {
    $$ = $1
  }
| expression BIT_AND expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITAND, Right: $3}
  }
| expression BIT_OR expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITOR, Right: $3}
  }
| expression CARET expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITXOR, Right: $3}
  }
| expression PLUS expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_PLUS, Right: $3}
  }
| expression SUB expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MINUS, Right: $3}
  }
| expression TIMES expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MULT, Right: $3}
  }
| expression DIV expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_DIV, Right: $3}
  }
| expression IDIV expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_IDIV, Right: $3}
  }
| expression MOD expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MOD, Right: $3}
  }
| expression EQ expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_EQ, Right: $3}
  }
| expression EQ all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_EQ, SubqueryOperator: $3, Right: $4}
  }
| expression NE expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NE, Right: $3}
  }
| expression NE all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NE, SubqueryOperator: $3, Right: $4}
  }
| expression NULL_SAFE_EQUAL expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NSE, Right: $3}
  }
| expression NULL_SAFE_EQUAL all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NSE, SubqueryOperator: $3, Right: $4}
  }
| expression LT expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LT, Right: $3}
  }
| expression LT all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LT, SubqueryOperator: $3, Right: $4}
  }
| expression GT expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_GT, Right: $3}
  }
| expression GT all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_GT, SubqueryOperator: $3, Right: $4}
  }
| expression LE expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LE, Right: $3}
  }
| expression LE all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LE, SubqueryOperator: $3, Right: $4}
  }
| expression GE expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_GE, Right: $3}
  }
| expression GE all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_GE, SubqueryOperator: $3, Right: $4}
  }
| expression OR expression
  {
    $$ = &OrExpr{Left: $1, Right: $3}
  }
| expression XOR expression
  {
    $$ = &XorExpr{Left: $1, Right: $3}
  }
| NOT expression
  {
    $$ = &NotExpr{Expr: $2}
  }
| unary_operator expression %prec UNARY
  {
    if num, ok := $2.(NumVal); ok {
      switch $1 {
      case '-':
        $$ = append(NumVal("-"), num...)
      case '+':
        $$ = num
      default:
        $$ = &UnaryExpr{Operator: $1, Expr: $2}
      }
    } else {
      $$ = &UnaryExpr{Operator: $1, Expr: $2}
    }
  }
| expression IN tuple
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IN, Right: $3}
  }
| expression NOT IN tuple
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NOT_IN, Right: $4}
  }
| expression LIKE expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LIKE, Right: $3}
  }
| expression NOT LIKE expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NOT_LIKE, Right: $4}
  }
| expression IS NULL
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IS, Right: &NullVal{}}
  }
| expression IS NOT NULL
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IS_NOT, Right: &NullVal{}}
  }
| expression IS boolean_value
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IS, Right: $3}
  }
| expression IS NOT boolean_value
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IS_NOT, Right: $4}
  }
| expression REGEXP expression
  {
    $$ = &RegexExpr{Operand: $1, Pattern: $3}
  }
| expression NOT REGEXP expression
  {
    $$ = &NotExpr{&RegexExpr{Operand: $1, Pattern: $4}}
  }
| EXISTS subquery
  {
    $$ = &ExistsExpr{Subquery: $2}
  }
| expression BETWEEN expression AND expression
  {
    $$ = &RangeCond{Left: $1, Operator: AST_BETWEEN, From: $3, To: $5}
  }
| expression NOT BETWEEN expression AND expression
  {
    $$ = &RangeCond{Left: $1, Operator: AST_NOT_BETWEEN, From: $4, To: $6}
  }
| case_expression
  {
    $$ = $1
  }
| sql_id LPAREN RPAREN
  {
    $$ = &FuncExpr{Name: bytes.ToLower($1)}
  }
| sql_id LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: bytes.ToLower($1), Exprs: $3}
  }
| sql_id LPAREN DISTINCT select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: bytes.ToLower($1), Distinct: true, Exprs: $4}
  }
| keyword_as_func LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: $1, Exprs: $3}
  }
| CURRENT_TIMESTAMP
  {
    $$ = &FuncExpr{Name: []byte("current_timestamp")}
  }
| CURRENT_TIMESTAMP LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: []byte("current_timestamp")}
  }
| CURRENT_TIMESTAMP LPAREN RPAREN
  {
    $$ = &FuncExpr{Name: []byte("current_timestamp")}
  }
| TIMESTAMPADD LPAREN time_interval COMMA select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5...)}
  }
| TIMESTAMPADD LPAREN sql_time_interval COMMA select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: []byte("timestampadd"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5...)}
  }
| TIMESTAMPDIFF LPAREN time_interval COMMA select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5...)}
  }
| TIMESTAMPDIFF LPAREN sql_time_interval COMMA select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: []byte("timestampdiff"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5...)}
  }
| CONVERT LPAREN expression COMMA sql_types RPAREN
  {
    $$ = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr:$3}, &NonStarExpr{Expr:KeywordVal($5)}})}
  }
| CAST LPAREN expression AS sql_types RPAREN
  {
    $$ = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr:$3}, &NonStarExpr{Expr:KeywordVal($5)}})}
  }
| CAST LPAREN expression AS sql_types PRECISION RPAREN
  {
    $$ = &FuncExpr{Name: []byte("convert"), Exprs: append(SelectExprs{&NonStarExpr{Expr:$3}, &NonStarExpr{Expr:KeywordVal($5)}})}
  }
| DATE LPAREN select_expression RPAREN
  {
    $$ = &FuncExpr{Name: []byte("date"), Exprs: SelectExprs{$3}}
  }
| EXTRACT LPAREN time_interval FROM select_expression RPAREN
  {
    $$ = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5)}
  }
| EXTRACT LPAREN sql_time_unit FROM select_expression RPAREN
  {
    $$ = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5)}
  }
| EXTRACT LPAREN sql_time_interval FROM select_expression RPAREN
  {
    $$ = &FuncExpr{Name: []byte("extract"), Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5)}
  }
| DATE_ADD LPAREN select_expression COMMA INTERVAL select_expression time_interval RPAREN
  {
    $$ = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{$3, $6, &NonStarExpr{Expr: KeywordVal($7)}})}
  }
| DATE_ADD LPAREN select_expression COMMA INTERVAL select_expression sql_time_unit RPAREN
  {
    $$ = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{$3, $6, &NonStarExpr{Expr: KeywordVal($7)}})}
  }
| DATE_ADD LPAREN select_expression COMMA INTERVAL select_expression sql_time_interval RPAREN
  {
    $$ = &FuncExpr{Name: []byte("date_add"), Exprs: append(SelectExprs{$3, $6, &NonStarExpr{Expr: KeywordVal($7)}})}
  }
| DATE_SUB LPAREN select_expression COMMA INTERVAL select_expression time_interval RPAREN
  {
    $$ = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{$3, $6, &NonStarExpr{Expr: KeywordVal($7)}})}
  }
| DATE_SUB LPAREN select_expression COMMA INTERVAL select_expression sql_time_unit RPAREN
  {
    $$ = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{$3, $6, &NonStarExpr{Expr: KeywordVal($7)}})}
  }
| DATE_SUB LPAREN select_expression COMMA INTERVAL select_expression sql_time_interval RPAREN
  {
    $$ = &FuncExpr{Name: []byte("date_sub"), Exprs: append(SelectExprs{$3, $6, &NonStarExpr{Expr: KeywordVal($7)}})}
  }
| STR_TO_DATE LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: []byte("str_to_date"), Exprs: $3}
  }

sql_time_interval:
    SQL_TSI_YEAR
  {
    $$ = YEAR_BYTES
  }
| SQL_TSI_QUARTER
  {
    $$ = QUARTER_BYTES
  }
| SQL_TSI_MONTH
  {
    $$ = MONTH_BYTES
  }
| SQL_TSI_WEEK
  {
    $$ = WEEK_BYTES
  }
| SQL_TSI_DAY
  {
    $$ = DAY_BYTES
  }
| SQL_TSI_HOUR
  {
    $$ = HOUR_BYTES
  }
| SQL_TSI_MINUTE
  {
    $$ = MINUTE_BYTES
  }
| SQL_TSI_SECOND
  {
    $$ = SECOND_BYTES
  }

time_interval:
  YEAR
  {
    $$ = YEAR_BYTES
  }
| QUARTER
  {
    $$ = QUARTER_BYTES
  }
| MONTH
  {
    $$ = MONTH_BYTES
  }
| WEEK
  {
    $$ = WEEK_BYTES
  }
| DAY
  {
    $$ = DAY_BYTES
  }
| HOUR
  {
    $$ = HOUR_BYTES
  }
| MINUTE
  {
    $$ = MINUTE_BYTES
  }
| SECOND
  {
    $$ = SECOND_BYTES
  }
| MICROSECOND
  {
    $$ = MICROSECOND_BYTES
  }

sql_time_unit:
  SECOND_MICROSECOND
  {
    $$ = SECOND_MICROSECOND_BYTES
  }
| MINUTE_MICROSECOND
  {
    $$ = MINUTE_MICROSECOND_BYTES
  }
| MINUTE_SECOND
  {
    $$ = MINUTE_SECOND_BYTES
  }
| HOUR_MICROSECOND
  {
    $$ = HOUR_MICROSECOND_BYTES
  }
| HOUR_SECOND
  {
    $$ = HOUR_SECOND_BYTES
  }
| HOUR_MINUTE
  {
    $$ = HOUR_MINUTE_BYTES
  }
| DAY_MICROSECOND
  {
    $$ = DAY_MICROSECOND_BYTES
  }
| DAY_SECOND
  {
    $$ = DAY_SECOND_BYTES
  }
| DAY_MINUTE
  {
    $$ = DAY_MINUTE_BYTES
  }
| DAY_HOUR
  {
    $$ = DAY_HOUR_BYTES
  }
| YEAR_MONTH
  {
    $$ = YEAR_MONTH_BYTES
  }

sql_types:
  CHAR
    {
      $$ = CHAR_BYTES
    }
  | DATE
    {
      $$ = DATE_BYTES
    }
  | SIGNED
    {
      $$ = INTEGER_BYTES
    }
  | SIGNED INTEGER
    {
      $$ = INTEGER_BYTES
    }
  | UNSIGNED
    {
      $$ = INTEGER_BYTES
    }
  | UNSIGNED INTEGER
    {
      $$ = INTEGER_BYTES
    }
  | SQL_BIGINT
    {
      $$ = INTEGER_BYTES
    }
  | SQL_VARCHAR
    {
      $$ = CHAR_BYTES
    }
  | SQL_DATE
    {
      $$ = DATE_BYTES
    }
  | SQL_TIMESTAMP
    {
      $$ = DATETIME_BYTES
    }
  | SQL_DOUBLE
    {
      $$ = FLOAT_BYTES
    }
  // We don't want to parse datetime as a token since MySQL allows
  // it to be a table name. As a result we parse it as a standard
  // ID and check it to see if it matches datetime.
  | ID
    {
      if bytes.Equal(bytes.ToLower($1), []byte("datetime")) {
        $$ = DATETIME_BYTES
      } else {
        yylex.Error("expecting datetime")
        return 1
      }
    }

keyword_as_func:
  IF
  {
    $$ = IF_BYTES
  }
| VALUES
  {
    $$ = VALUES_BYTES
  }
| RIGHT
  {
    $$ = RIGHT_BYTES
  }
| LEFT
  {
    $$ = LEFT_BYTES
  }
| MOD
  {
    $$ = MOD_BYTES
  }
| time_interval
  {
    $$ = $1
  }

unary_operator:
  PLUS
  {
    $$ = AST_UPLUS
  }
| SUB
  {
    $$ = AST_UMINUS
  }
| TILDE
  {
    $$ = AST_TILDA
  }

case_expression:
  CASE expression_opt when_expression_list else_expression_opt END
  {
    $$ = &CaseExpr{Expr: $2, Whens: $3, Else: $4}
  }

expression_opt:
  {
    $$ = nil
  }
| expression
  {
    $$ = $1
  }

when_expression_list:
  when_expression
  {
    $$ = []*When{$1}
  }
| when_expression_list when_expression
  {
    $$ = append($1, $2)
  }

when_expression:
  WHEN expression THEN expression
  {
    $$ = &When{Cond: $2, Val: $4}
  }

else_expression_opt:
  {
    $$ = nil
  }
| ELSE expression
  {
    $$ = $2
  }

column_name:
  sql_id
  {
    $$ = &ColName{Name: $1}
  }
| ID DOT sql_id
  {
    $$ = &ColName{Qualifier: $1, Name: $3}
  }

column_name_opt:
  {
    $$ = nil
  }
| sql_id
  {
    $$ = &ColName{Name: $1}
  }
| STRING
  {
    $$ = &ColName{Name: $1}
  }

value:
STRING
  {
    $$ = StrVal($1)
  }
| NUMBER
  {
    $$ = NumVal($1)
  }
| VALUE_ARG
  {
    $$ = ValArg($1)
  }
| DATE STRING
  {
    $$ = DateVal{Name: AST_DATE, Val: $2}
  }
 | TIME STRING
  {
    $$ = DateVal{Name: AST_TIME, Val: $2}
  }
| TIMESTAMP STRING
  {
    $$ = DateVal{Name: AST_TIMESTAMP, Val: $2}
  }
| LBRACE ID STRING RBRACE
  {
    if bytes.Equal(bytes.ToLower($2), []byte("d")) {
      $$ = DateVal{Name: AST_DATE, Val: $3}
    } else if bytes.Equal(bytes.ToLower($2), []byte("t")) {
      $$ = DateVal{Name: AST_TIME, Val: $3}
    } else if bytes.Equal(bytes.ToLower($2), []byte("ts")) {
      $$ = DateVal{Name: AST_TIMESTAMP, Val: $3}
    } else {
      yylex.Error("expecting d, t, or ts")
      return 1
    }
  }
| NULL
  {
    $$ = &NullVal{}
  }
| boolean_value
  {
    $$ = $1
  }

boolean_value:
TRUE
  {
    $$ = &TrueVal{}
  }
| FALSE
  {
    $$ = &FalseVal{}
  }
| UNKNOWN
  {
    $$ = &UnknownVal{}
  }

group_by_opt:
  {
    $$ = nil
  }
| GROUP BY expression_list
  {
    $$ = $3
  }

having_opt:
  {
    $$ = nil
  }
| HAVING expression
  {
    $$ = $2
  }

order_by_opt:
  {
    $$ = nil
  }
| ORDER BY order_list
  {
    $$ = $3
  }

order_list:
  order
  {
    $$ = OrderBy{$1}
  }
| order_list COMMA order
  {
    $$ = append($1, $3)
  }

order:
  expression asc_desc_opt
  {
    $$ = &Order{Expr: $1, Direction: $2}
  }

asc_desc_opt:
  {
    $$ = AST_ASC
  }
| ASC
  {
    $$ = AST_ASC
  }
| DESC
  {
    $$ = AST_DESC
  }

limit_opt:
  {
    $$ = nil
  }
| LIMIT expression
  {
    $$ = &Limit{Rowcount: $2}
  }
| LIMIT expression COMMA expression
  {
    $$ = &Limit{Offset: $2, Rowcount: $4}
  }
| LIMIT expression OFFSET expression
  {
    $$ = &Limit{Offset: $4, Rowcount: $2}
  }

lock_opt:
  {
    $$ = ""
  }
| FOR UPDATE
  {
    $$ = AST_FOR_UPDATE
  }
| LOCK IN sql_id sql_id
  {
    if !bytes.Equal($3, SHARE) {
      yylex.Error("expecting share")
      return 1
    }
    if !bytes.Equal($4, MODE) {
      yylex.Error("expecting mode")
      return 1
    }
    $$ = AST_SHARE_MODE
  }

column_list_opt:
  {
    $$ = nil
  }
| LPAREN column_list RPAREN
  {
    $$ = $2
  }

column_list:
  column_name
  {
    $$ = Columns{&NonStarExpr{Expr: $1}}
  }
| column_list COMMA column_name
  {
    $$ = append($$, &NonStarExpr{Expr: $3})
  }

on_dup_opt:
  {
    $$ = nil
  }
| ON DUPLICATE KEY UPDATE update_list
  {
    $$ = $5
  }

update_list:
  update_expression
  {
    $$ = UpdateExprs{$1}
  }
| update_list COMMA update_expression
  {
    $$ = append($1, $3)
  }

update_expression:
  column_name EQ expression
  {
    $$ = &UpdateExpr{Name: $1, Expr: $3}
  }

exists_opt:
  { $$ = struct{}{} }
| IF EXISTS
  { $$ = struct{}{} }

not_exists_opt:
  { $$ = struct{}{} }
| IF NOT EXISTS
  { $$ = struct{}{} }

ignore_opt:
  { $$ = struct{}{} }
| IGNORE
  { $$ = struct{}{} }

non_rename_operation:
  ALTER
  { $$ = struct{}{} }
| DEFAULT
  { $$ = struct{}{} }
| DROP
  { $$ = struct{}{} }
| ORDER
  { $$ = struct{}{} }
| ID
  { $$ = struct{}{} }

to_opt:
  { $$ = struct{}{} }
| TO
  { $$ = struct{}{} }

constraint_opt:
  { $$ = struct{}{} }
| UNIQUE
  { $$ = struct{}{} }

using_opt:
  { $$ = struct{}{} }
| USING sql_id
  { $$ = struct{}{} }

sql_id:
  ID
  {
    $$ = $1
  }

force_eof:
{
  ForceEOF(yylex)
}
