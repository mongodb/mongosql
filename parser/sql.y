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

%}

%union {
  empty       struct{}
  statement   Statement
  selStmt     SelectStatement
  bool        bool
  byt         byte
  bytes       []byte
  bytes2      [][]byte
  str         string
  selectExprs SelectExprs
  selectExpr  SelectExpr
  columns     Columns
  columnExprs ColumnExprs
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
  updateExprs UpdateExprs
  updateExpr  *UpdateExpr
  alterSpec   *AlterSpec
  alterSpecs  []*AlterSpec
  renameSpec  *RenameSpec
  renameSpecs []*RenameSpec
}

%token LEX_ERROR
%token <bytes> ID STRING NUMBER VALUE_ARG COMMENT
%token <empty> LPAREN RPAREN LBRACE RBRACE TILDE

%token <empty> SELECT DROP CREATE SET SHOW UPDATE WHERE GROUP HAVING ORDER BY LIMIT OFFSET FOR SOME ANY TRUE FALSE UNKNOWN
%token <empty> ALTER ADD CHANGE RENAME COLUMN TO
%token <empty> ALL DISTINCT PRECISION AS EXISTS NULL ASC DESC VALUES DEFAULT LOCK
%token <empty> DATE DATETIME TIME TIMESTAMP CURRENT_TIMESTAMP CURRENT_DATE UTC_TIMESTAMP UTC_DATE DECIMAL NCHAR
%token <empty> TIMESTAMPADD TIMESTAMPDIFF EXTRACT DATE_ADD ADDDATE
%token <empty> DATE_SUB SUBDATE ROW
%token <empty> CONVERT CAST CHAR SIGNED UNSIGNED SQL_BIGINT SQL_VARCHAR SQL_DATE SQL_TIMESTAMP SQL_DOUBLE INTEGER
%token <empty> BOTH LEADING TRAILING TRIM SUBSTRING SUBSTR
%token <empty> BINARY MASTER LOGS DATABASE SCHEMA EVENT FUNCTION PROCEDURE BINLOG EVENTS TRIGGER USER
%token <empty> ENGINE MUTEX ENGINES STORAGE ERRORS COUNT CODE GRANTS OPEN PLUGINS PRIVILEGES
%token <empty> PROFILE PROFILES RELAYLOG SLAVE HOSTS TRIGGERS WARNINGS CHANNEL INDEXES KEYS SCHEMAS
%token <empty> FN OJ ESCAPE
%token <empty> TABLE INDEX VIEW IGNORE IF
%token <bytes> TRANSACTION ISOLATION LEVEL
%token <bytes> READ WRITE ONLY
%token <bytes> REPEATABLE COMMITTED UNCOMMITTED SERIALIZABLE
%token <empty> NAMES CHARACTER COLLATE
%token <empty> DATABASES TABLES PROXY VARIABLES FULL COLUMNS COLLATION PROCESSLIST STATUS CHARSET
%token <empty> EXPLAIN DESCRIBE
%token <empty> EXTENDED PARTITIONS FORMAT TRADITIONAL JSON
%token <empty> KILL FLUSH SAMPLE
%token <empty> CONNECTION QUERY
%token <empty> SESSION GLOBAL 
%token <empty> TEMPORARY RESTRICT CASCADE
%token <empty> USING

%nonassoc <empty> YEAR QUARTER MONTH WEEK DAY HOUR MINUTE SECOND MICROSECOND
%nonassoc <empty> SECOND_MICROSECOND MINUTE_MICROSECOND MINUTE_SECOND HOUR_MICROSECOND HOUR_SECOND HOUR_MINUTE DAY_MICROSECOND DAY_SECOND DAY_MINUTE DAY_HOUR YEAR_MONTH
%nonassoc <empty> SQL_TSI_YEAR SQL_TSI_QUARTER SQL_TSI_MONTH SQL_TSI_WEEK SQL_TSI_DAY SQL_TSI_HOUR SQL_TSI_MINUTE SQL_TSI_SECOND
%nonassoc <empty> FROM
%left <empty> UNION MINUS EXCEPT INTERSECT
%left <empty> COMMA
%left <empty> JOIN STRAIGHT_JOIN LEFT RIGHT INNER OUTER CROSS USE FORCE
%left <empty> NATURAL
%left <empty> ON
%left <empty> OR 
%left <empty> XOR
%left <empty> AND
%right <empty> NOT
%left <empty> BETWEEN CASE WHEN THEN ELSE
%left <empty> EQ NULL_SAFE_EQUAL GE GT LE LT NE IS LIKE REGEXP IN
%left <empty> BIT_AND BIT_OR CARET
%left <empty> PLUS SUB
%left <empty> TIMES MOD DIV IDIV
%nonassoc <empty> DOT
%left <empty> UNARY
%left <empty> END
%left <empty> INTERVAL

%start any_command

%type <statement> command
%type <selStmt> select_statement select_statement_with_paren_order_limit
%type <statement> set_statement use_statement show_statement explain_statement explainable_stmt
%type <statement> kill_statement drop_statement alter_statement rename_statement
%type <statement> flush_statement
%type <alterSpecs> alter_spec_list
%type <alterSpec> alter_spec
%type <renameSpec> table_rename
%type <renameSpecs> table_rename_list
%type <bytes> column_opt to_as_opt
%type <bytes2> comment_opt comment_list
%type <str> union_op
%type <str> all_any_some
%type <str> distinct_opt
%type <selectExprs> select_expression_list
%type <selectExpr> select_expression
%type <bytes> as_opt
%type <expr> expression bool_pri predicate bit_expr simple_expr func_expr func_expr_reserved_keyword func_expr_unconventional func_expr_generic func_expr_conflict
%type <tableExprs> table_expression_list
%type <columnExprs> column_expression_list
%type <tableExpr> table_expression join_expression
%type <str> join_type natural_join_type
%type <smTableExpr> simple_table_expression
%type <tableName> table_name
%type <indexHints> index_hint_list
%type <bytes2> index_list
%type <expr> where_expression_opt like_escape_opt
%type <expr> value
%type <tuple> tuple
%type <expr> boolean_value
%type <exprs> expression_list
%type <bytes> interval_unit
%type <bytes> time_interval
%type <bytes> sql_time_interval
%type <bytes> sql_time_unit
%type <bytes> sql_types
%type <bytes> substr 
%type <subquery> subquery non_derived_subquery 
%type <byt> unary_operator
%type <colName> column_name explain_column_name
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
%type <updateExprs> update_list
%type <updateExpr> update_expression
%type <bytes> transaction_characteristics
%type <bytes> transaction_characteristic
%type <bytes> transaction_level
%type <empty> explain_alias
%type <empty> in_or_from optional_parens
%type <bool> exists_opt temporary_opt
%type <bytes> cascade_or_restrict_opt

%type <empty> if_not_exists_opt storage_opt
%type <expr> in_opt from_opt
%type <expr> like_or_where_opt
%type <expr> show_from_in show_from_in_opt
%type <str> show_full
%type <str> scope_modifier_opt
%type <str> explain_type
%type <str> format_name
%type <str> kill_modifier
%type <bytes> for_user_opt for_channel_opt both_leading_trailing_opt
%type <bytes> sql_id keyword_as_id
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
| select_statement_with_paren_order_limit
  {
    $$ = $1
  }
| set_statement
| kill_statement
| show_statement
| explain_statement
| use_statement
| drop_statement
| flush_statement
| alter_statement
| rename_statement

select_statement_with_paren_order_limit:
  non_derived_subquery order_by_opt limit_opt
  {
    $$ = &Select{SelectExprs: []SelectExpr{&StarExpr{}}, From: []TableExpr{&AliasedTableExpr{Expr:$1}}, OrderBy: $2, Limit: $3}
  }

select_statement:
  non_derived_subquery
  {
    $$ = &Select{SelectExprs: []SelectExpr{&StarExpr{}}, From: []TableExpr{&AliasedTableExpr{Expr:$1}}}
  }
| SELECT comment_opt distinct_opt select_expression_list limit_opt
  {
    $$ = &SimpleSelect{Comments: Comments($2), Distinct: $3, SelectExprs: $4, Limit: $5}
  }
| SELECT comment_opt distinct_opt select_expression_list FROM table_expression_list where_expression_opt group_by_opt having_opt order_by_opt limit_opt lock_opt
  {
    $$ = &Select{Comments: Comments($2), Distinct: $3, SelectExprs: $4, From: $6, Where: NewWhere(AST_WHERE, $7), GroupBy: GroupBy($8), Having: NewWhere(AST_HAVING, $9), OrderBy: $10, Limit: $11, Lock: $12}
  }
| select_statement union_op select_statement %prec UNION
  {
    $$ = &Union{Type: $2, Left: $1, Right: $3}
  }

non_derived_subquery:
 LPAREN select_statement RPAREN
 {
    $$ = &Subquery{Select: $2, IsDerived: false}
 }

use_statement:
  USE ID
  {
    $$ = &Use{DBName: $2}
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
    $$ = &Set{Comments: append([][]byte{}, []byte($2), TRANSACTION_BYTES, $4)}
  }

drop_statement:
  DROP temporary_opt TABLE exists_opt table_name cascade_or_restrict_opt
  {
    $$ = &DropTable{Name: $5, Temporary: $2, Exists: $4, Opt: $6 }
  }

flush_statement:
  FLUSH LOGS
  {
    $$ = &Flush{Kind: FlushLogs}
  }
| FLUSH SAMPLE
  {
    $$ = &Flush{Kind: FlushSample}
  }

rename_statement:
  RENAME TABLE table_rename_list
  {
    $$ = &RenameTable{ Renames: $3 }
  }

table_rename_list:
  table_rename
  {
    $$ = []*RenameSpec{$1}
  }
| table_rename_list COMMA table_rename
  {
    $$ = append($$, $3)
  }

table_rename:
  table_name TO table_name
  {
    $$ = &RenameSpec{ Table: $1, NewTable: $3 }
  }

alter_statement:
  ALTER TABLE table_name alter_spec_list
  {
    $$ = &AlterTable{ Table: $3, Specs: $4 }
  }

alter_spec_list:
  alter_spec
  {
    $$ = []*AlterSpec{$1}
  }
| alter_spec_list COMMA alter_spec
  {
    $$ = append($$, $3)
  }

alter_spec:
  CHANGE column_opt column_name column_name
  {
    $$ = &AlterSpec{
        Type: "rename_column",
        Column: $3,
        NewColumn: $4,
    }
  }
| DROP column_opt column_name
  {
    $$ = &AlterSpec{
        Type: "drop_column",
        Column: $3,
    }
  }
| RENAME to_as_opt table_name
  {
    $$ = &AlterSpec{
        Type: "rename_table",
        NewTable: $3,
    }
  }

column_opt:
  { $$ = nil }
| COLUMN
  { $$ = COLUMN_BYTES }

to_as_opt:
  { $$ = nil }
| TO
  { $$ = TO_BYTES }
| AS
  { $$ = AS_BYTES }

temporary_opt:
  { $$ = false }
| TEMPORARY
  { $$ = true }

exists_opt:
  { $$ = false }
| IF EXISTS
  { $$ = true }

cascade_or_restrict_opt:
  { $$ = nil }
| CASCADE
  { $$ = CASCADE_BYTES }
| RESTRICT
  { $$ = RESTRICT_BYTES }

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
    $$ = SERIALIZABLE_BYTES
  }

substr:
  SUBSTR
  {
    $$ = SUBSTR_BYTES
  }
| SUBSTRING
  {
    $$ = SUBSTRING_BYTES
  }

in_or_from:
  IN 
| FROM

show_from_in:
  in_or_from DOT ID 
  {
    $$ = StrVal($3)
  }
| in_or_from ID DOT ID 
  {
    $$ = &ColName{Qualifier: $2, Name: $4}
  }
| in_or_from ID in_or_from ID
  {
    $$ = &ColName{Qualifier: $4, Name: $2}
  }
| in_or_from ID 
  {
    $$ = StrVal($2)
  }

show_from_in_opt:
  {
    $$ = nil
  }
| show_from_in
  {
    $$ = $1
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
  SHOW BINARY LOGS
  {
    $$ = &Show{Section: "binary logs"}
  }
| SHOW MASTER LOGS
  {
    $$ = &Show{Section: "master logs"}
  }
| SHOW BINLOG EVENTS in_opt from_opt limit_opt
  {
    $$ = &Show{Section: "binlog events"}
  }
| SHOW CREATE DATABASE if_not_exists_opt ID
  {
    $$ = &Show{Section: "create database", Modifier: string($5)}
  }
| SHOW CREATE SCHEMA if_not_exists_opt ID
  {
    $$ = &Show{Section: "create schema", Modifier: string($5)}
  }
| SHOW CREATE EVENT ID
  {
    $$ = &Show{Section: "create event", Modifier: string($4)}
  }
| SHOW CREATE FUNCTION ID
  {
    $$ = &Show{Section: "create function", Modifier: string($4)}
  }
| SHOW CREATE PROCEDURE ID
  {
    $$ = &Show{Section: "create procedure", Modifier: string($4)}
  }
| SHOW CREATE TABLE ID DOT ID
  {
    $$ = &Show{Section: "create table", From: &ColName{$6, $4}}
  }
| SHOW CREATE TABLE DOT ID
  {
    $$ = &Show{Section: "create table", From: StrVal($5)}
  }
| SHOW CREATE TABLE ID
  {
    $$ = &Show{Section: "create table", From: StrVal($4)}
  }
| SHOW CREATE TRIGGER ID
  {
    $$ = &Show{Section: "create trigger", Modifier: string($4)}
  }
| SHOW CREATE USER ID
  {
    $$ = &Show{Section: "create user", Modifier: string($4)}
  }
| SHOW CREATE VIEW ID
  {
    $$ = &Show{Section: "create view", Modifier: string($4)}
  }
| SHOW ENGINE ID STATUS
  {
    $$ = &Show{Section: "engine", Modifier: string($3)}
  }
| SHOW ENGINE ID MUTEX
  {
    $$ = &Show{Section: "engine", Modifier: string($3)}
  }
| SHOW storage_opt ENGINES 
  {
    $$ = &Show{Section: "engines"}
  }
| SHOW ERRORS limit_opt
  {
    $$ = &Show{Section: "errors"}
  }
| SHOW COUNT LPAREN TIMES RPAREN ERRORS
  {
    $$ = &Show{Section: "count(*) errors"}
  }
| SHOW EVENTS show_from_in_opt like_or_where_opt 
  {
    $$ = &Show{Section: "events"}
  }
| SHOW FUNCTION CODE ID
  {
    $$ = &Show{Section: "function code", Modifier: string($4)}
  }
| SHOW FUNCTION STATUS like_or_where_opt
  {
    $$ = &Show{Section: "function status"}
  }
| SHOW GRANTS for_user_opt
  {
    $$ = &Show{Section: "grants", Modifier: string($3)}
  }
| SHOW INDEX show_from_in where_expression_opt
  {
    $$ = &Show{Section: "index"}
  }
| SHOW INDEXES show_from_in where_expression_opt
  {
    $$ = &Show{Section: "indexes"}
  }
| SHOW KEYS show_from_in where_expression_opt
  {
    $$ = &Show{Section: "keys"}
  }
| SHOW MASTER STATUS
  {
    $$ = &Show{Section: "master status"}
  }
| SHOW OPEN TABLES show_from_in_opt like_or_where_opt
  {
    $$ = &Show{Section: "open tables"}
  }
| SHOW PLUGINS
  {
    $$ = &Show{Section: "plugins"}
  }
| SHOW PRIVILEGES
  {
    $$ = &Show{Section: "privileges"}
  }
| SHOW PROCEDURE CODE ID
  {
    $$ = &Show{Section: "procedure code", Modifier: string($4)}
  }
| SHOW PROCEDURE STATUS like_or_where_opt
  {
    $$ = &Show{Section: "procedure status"}
  }
| SHOW PROFILE
  {
    $$ = &Show{Section: "profile"}
  }
| SHOW PROFILES
  {
    $$ = &Show{Section: "profiles"}
  }
| SHOW RELAYLOG EVENTS in_opt from_opt limit_opt
  {
    $$ = &Show{Section: "relaylog events"}
  }
| SHOW SLAVE HOSTS
  {
    $$ = &Show{Section: "slave hosts"}
  }
| SHOW SLAVE STATUS for_channel_opt
  {
    $$ = &Show{Section: "slave status", Modifier: string($4)}
  }
| SHOW TABLE STATUS show_from_in_opt like_or_where_opt
  {
    $$ = &Show{Section: "table status"}
  }
| SHOW TRIGGERS show_from_in_opt like_or_where_opt
  {
    $$ = &Show{Section: "table status"}
  }
| SHOW WARNINGS limit_opt
  {
    $$ = &Show{Section: "warnings"}
  }
| SHOW COUNT LPAREN TIMES RPAREN WARNINGS
  {
    $$ = &Show{Section: "count(*) errors"}
  }
| SHOW DATABASES like_or_where_opt
  {
    $$ = &Show{Section: "databases", LikeOrWhere: $3}
  }
| SHOW SCHEMAS like_or_where_opt
  {
    $$ = &Show{Section: "schemas", LikeOrWhere: $3}
  }
| SHOW scope_modifier_opt VARIABLES like_or_where_opt
  {
    $$ = &Show{Section: "variables", Modifier: $2, LikeOrWhere: $4}
  }
| SHOW show_full TABLES show_from_in_opt like_or_where_opt
  {
    $$ = &Show{Section: "tables", Modifier: $2, From: $4, LikeOrWhere: $5}
  }
| SHOW PROXY ID from_opt like_or_where_opt
  {
    $$ = &Show{Section: "proxy", Key: string($3), From: $4, LikeOrWhere: $5}
  }
| SHOW show_full COLUMNS show_from_in like_or_where_opt
  {
    $$ = &Show{Section: "columns", From: $4, Modifier: $2, LikeOrWhere: $5}
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
| DOT sql_id
  {
    $$ = &TableName{Name: $2}
  }
| sql_id DOT sql_id
  {
    $$ = &TableName{Qualifier: $1, Name: $3}
  }
  
explain_statement:
  explain_alias table_name explain_column_name
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

explain_column_name:
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

kill_statement:
  KILL expression
  {
    $$ = &Kill{Scope: AST_KILL_CONNECTION, ID: $2}
  }
| KILL kill_modifier expression
  {
    $$ = &Kill{Scope: $2, ID: $3}
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
| expression as_opt
  {
    $$ = &NonStarExpr{Expr: $1, As: $2}
  }
| expression as_opt PRECISION
  {
    $$ = &NonStarExpr{Expr: $1, As: $2}
  }
| sql_id DOT TIMES
  {
    $$ = &StarExpr{TableName: $1}
  }

column_expression_list:
  ID
  {
    $$ = ColumnExprs{ &ColName{Name: $1} }
  }
| column_expression_list COMMA ID
  {
    $$ = append($$, &ColName{Name: $3} )
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
| join_expression
  {
    $$ = $1
  }
| LBRACE OJ join_expression RBRACE
  {
    $$ = $3
  }

join_expression:
  table_expression join_type table_expression %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3}
  }
|
  table_expression STRAIGHT_JOIN table_expression %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: AST_STRAIGHT_JOIN, RightExpr: $3}
  }
| table_expression join_type table_expression ON expression %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3, On: $5}
  }
| table_expression STRAIGHT_JOIN table_expression ON expression %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: AST_STRAIGHT_JOIN, RightExpr: $3, On: $5}
  }
| table_expression join_type table_expression USING LPAREN column_expression_list RPAREN %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3, Using: $6}
  }
| table_expression natural_join_type table_expression %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3}
  }

as_opt: %prec INTERVAL
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
| STRING
  {
    $$ = $1
  }
| AS STRING 
  {
    $$ = $2
  }

join_type:
  JOIN
  {
    $$ = AST_JOIN
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

natural_join_type:
  NATURAL JOIN
  {
    $$ = AST_NATURAL_JOIN
  }
| NATURAL RIGHT JOIN
  {
    $$ = AST_NATURAL_RIGHT_JOIN
  }
| NATURAL RIGHT OUTER JOIN
  {
    $$ = AST_NATURAL_RIGHT_JOIN
  }
| NATURAL LEFT JOIN
  {
    $$ = AST_NATURAL_LEFT_JOIN
  }
| NATURAL LEFT OUTER JOIN
  {
    $$ = AST_NATURAL_LEFT_JOIN
  }

simple_table_expression:
  table_name
  {
    $$ = $1
  }
| subquery
  {
    $$ = $1
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
  ID
  {
    $$ = [][]byte{$1}
  }
| index_list COMMA ID
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

in_opt:
  {
    $$ = nil
  }
| IN expression
  {
    $$ = $2
  }

if_not_exists_opt:
  {
    $$ = struct{}{}
  }
| IF NOT EXISTS
  {
    $$ = struct{}{}
  }

from_opt:
  {
    $$ = nil
  }
| FROM expression
  {
    $$ = $2
  }

for_channel_opt:
  {
    $$ = nil
  }
| FOR CHANNEL ID 
  {
    $$ = $3
  }

for_user_opt:
  {
    $$ = nil
  }
| FOR ID
  {
    $$ = $2
  }

storage_opt:
  {
    $$ = struct{}{}
  }
| STORAGE
  {
    $$ = struct{}{}
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

tuple:
  ROW LPAREN expression_list COMMA expression RPAREN
  {
    $$ = ValTuple(append($3, $5))
  }
| LPAREN expression_list RPAREN
  {
    $$ = ValTuple($2)
  }

subquery:
  LPAREN select_statement RPAREN
  {
    $$ = &Subquery{Select: $2, IsDerived: true}
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
  expression OR expression
  {
    $$ = &OrExpr{Left: $1, Right: $3}
  }
| expression XOR expression
  {
    $$ = &XorExpr{Left: $1, Right: $3}
  }
| expression AND expression
  {
    $$ = &AndExpr{Left: $1, Right: $3}
  }
| NOT expression
  {
    $$ = &NotExpr{Expr: $2}
  }
| bool_pri IS boolean_value
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IS, Right: $3}
  }
| bool_pri IS NOT boolean_value
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IS_NOT, Right: $4}
  }
| bool_pri


bool_pri:
  bool_pri IS NULL
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IS, Right: &NullVal{}}
  }
| bool_pri IS NOT NULL
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IS_NOT, Right: &NullVal{}}
  }
| bool_pri EQ predicate
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_EQ, Right: $3}
  }
| bool_pri EQ all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_EQ, SubqueryOperator: $3, Right: $4}
  }
| bool_pri NE predicate
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NE, Right: $3}
  }
| bool_pri NE all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NE, SubqueryOperator: $3, Right: $4}
  }
| bool_pri NULL_SAFE_EQUAL predicate
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NSE, Right: $3}
  }
| bool_pri NULL_SAFE_EQUAL all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NSE, SubqueryOperator: $3, Right: $4}
  }
| bool_pri LT predicate
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LT, Right: $3}
  }
| bool_pri LT all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LT, SubqueryOperator: $3, Right: $4}
  }
| bool_pri GT predicate
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_GT, Right: $3}
  }
| bool_pri GT all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_GT, SubqueryOperator: $3, Right: $4}
  }
| bool_pri LE predicate
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LE, Right: $3}
  }
| bool_pri LE all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LE, SubqueryOperator: $3, Right: $4}
  }
| bool_pri GE predicate
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_GE, Right: $3}
  }
| bool_pri GE all_any_some subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_GE, SubqueryOperator: $3, Right: $4}
  }
| predicate


predicate:
  bit_expr IN subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IN, SubqueryOperator: AST_IN, Right: $3}
  }
| bit_expr IN tuple
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IN, Right: $3}
  }
| bit_expr NOT IN subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NOT_IN, SubqueryOperator: AST_NOT_IN, Right: $4}
  }
| bit_expr NOT IN tuple
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NOT_IN, Right: $4}
  }
| bit_expr BETWEEN bit_expr AND predicate
  {
    $$ = &RangeCond{Left: $1, Operator: AST_BETWEEN, From: $3, To: $5}
  }
| bit_expr NOT BETWEEN bit_expr AND predicate
  {
    $$ = &RangeCond{Left: $1, Operator: AST_NOT_BETWEEN, From: $4, To: $6}
  }
| bit_expr LIKE simple_expr like_escape_opt
  {
    $$ = &LikeExpr{Left: $1, Operator: AST_LIKE, Right: $3, Escape: $4}
  }
| bit_expr NOT LIKE simple_expr like_escape_opt
  {
    $$ = &LikeExpr{Left: $1, Operator: AST_NOT_LIKE, Right: $4, Escape: $5}
  }
| bit_expr REGEXP bit_expr
  {
    $$ = &RegexExpr{Operand: $1, Pattern: $3}
  }
| bit_expr NOT REGEXP bit_expr
  {
    $$ = &NotExpr{&RegexExpr{Operand: $1, Pattern: $4}}
  }
| bit_expr


bit_expr:
  bit_expr BIT_AND bit_expr
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITAND, Right: $3}
  }
| bit_expr BIT_OR bit_expr
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITOR, Right: $3}
  }
| bit_expr PLUS bit_expr
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_PLUS, Right: $3}
  }
| bit_expr PLUS INTERVAL expression interval_unit %prec PLUS
  {
    $$ = &FuncExpr{
      Name: DATE_ADD_BYTES,
      Exprs: append(SelectExprs{
        &NonStarExpr{Expr: $1},
        &NonStarExpr{Expr: $4},
        &NonStarExpr{Expr: KeywordVal($5)},
      }),
    }
  }
| INTERVAL expression interval_unit PLUS bit_expr %prec INTERVAL
  {
    $$ = &FuncExpr{
      Name: DATE_ADD_BYTES,
      Exprs: append(SelectExprs{
        &NonStarExpr{Expr: $5},
        &NonStarExpr{Expr: $2},
        &NonStarExpr{Expr: KeywordVal($3)},
      }),
    }
  }
| bit_expr SUB bit_expr
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MINUS, Right: $3}
  }
| bit_expr SUB INTERVAL expression interval_unit %prec SUB
  {
    $$ = &FuncExpr{
      Name: SUBDATE_BYTES,
      Exprs: append(SelectExprs{
        &NonStarExpr{Expr: $1},
        &NonStarExpr{Expr: $4},
        &NonStarExpr{Expr: KeywordVal($5)},
      }),
    }
  }
| bit_expr TIMES bit_expr
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MULT, Right: $3}
  }
| bit_expr DIV bit_expr
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_DIV, Right: $3}
  }
| bit_expr IDIV bit_expr
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_IDIV, Right: $3}
  }
| bit_expr MOD bit_expr
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MOD, Right: $3}
  }
| bit_expr CARET bit_expr
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITXOR, Right: $3}
  }
| simple_expr


simple_expr:
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
| subquery
  {
    $$ = $1
  }
| unary_operator simple_expr
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
| EXISTS subquery
  {
    $$ = &ExistsExpr{Subquery: $2}
  }
| case_expression
  {
    $$ = $1
  }
| func_expr
  {
    $$ = $1
  }
| VALUES LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: VALUES_BYTES, Exprs: $3}
  }
| LBRACE FN func_expr RBRACE
  {
    $$ = $3
  }

func_expr:
  func_expr_reserved_keyword
  {
    $$ = $1
  }
| func_expr_unconventional
  {
    $$ = $1
  }
| func_expr_generic
  {
    $$ = $1
  }
| func_expr_conflict
  {
    $$ = $1
  }

/*
  function calls using reserved keywords with either conventional
  or unconventional syntax.
*/
func_expr_reserved_keyword:
  CHAR LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: CHAR_BYTES, Exprs: $3}
  }
| CONVERT LPAREN expression COMMA sql_types RPAREN
  {
    $$ = &FuncExpr{Name: CONVERT_BYTES, Exprs: append(SelectExprs{&NonStarExpr{Expr:$3}, &NonStarExpr{Expr:KeywordVal($5)}})}
  }
| LEFT LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: LEFT_BYTES, Exprs: $3}
  }
| RIGHT LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: RIGHT_BYTES, Exprs: $3}
  }

/*
  function calls using unconventional call syntax. Most functions are called as (arg,arg,arg). The ones
  in this production are called with a different syntax.
*/ 
func_expr_unconventional:
  ADDDATE LPAREN select_expression COMMA INTERVAL select_expression interval_unit RPAREN
  {
    $$ = &FuncExpr{Name: ADDDATE_BYTES, Exprs: append(SelectExprs{$3, $6, &NonStarExpr{Expr: KeywordVal($7)}})}
  }
| ADDDATE LPAREN select_expression COMMA select_expression RPAREN
  {
    $$ = &FuncExpr{Name: ADDDATE_BYTES, Exprs: append(SelectExprs{$3, $5, &NonStarExpr{Expr: KeywordVal(DAY_BYTES)}})}
  }
| CAST LPAREN expression AS sql_types RPAREN
  {
    $$ = &FuncExpr{Name: CONVERT_BYTES, Exprs: append(SelectExprs{&NonStarExpr{Expr:$3}, &NonStarExpr{Expr:KeywordVal($5)}})}
  }
| CAST LPAREN expression AS sql_types PRECISION RPAREN
  {
    $$ = &FuncExpr{Name: CONVERT_BYTES, Exprs: append(SelectExprs{&NonStarExpr{Expr:$3}, &NonStarExpr{Expr:KeywordVal($5)}})}
  }
| CURRENT_DATE optional_parens
  {
    $$ = &FuncExpr{Name: CURRENT_DATE_BYTES}
  }
| CURRENT_TIMESTAMP optional_parens
  {
    $$ = &FuncExpr{Name: CURRENT_TIMESTAMP_BYTES}
  }
| CURRENT_TIMESTAMP LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: CURRENT_TIMESTAMP_BYTES}
  }
| DATE_ADD LPAREN select_expression COMMA INTERVAL select_expression interval_unit RPAREN
  {
    $$ = &FuncExpr{Name: DATE_ADD_BYTES, Exprs: append(SelectExprs{$3, $6, &NonStarExpr{Expr: KeywordVal($7)}})}
  }
| DATE_SUB LPAREN select_expression COMMA INTERVAL select_expression interval_unit RPAREN
  {
    $$ = &FuncExpr{Name: DATE_SUB_BYTES, Exprs: append(SelectExprs{$3, $6, &NonStarExpr{Expr: KeywordVal($7)}})}
  }
| EXTRACT LPAREN interval_unit FROM select_expression RPAREN
  {
    $$ = &FuncExpr{Name: EXTRACT_BYTES, Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5)}
  }
| SUBDATE LPAREN select_expression COMMA select_expression RPAREN
  {
    $$ = &FuncExpr{Name: SUBDATE_BYTES, Exprs: append(SelectExprs{$3, $5, &NonStarExpr{Expr: KeywordVal(DAY_BYTES)}})}
  }
| SUBDATE LPAREN select_expression COMMA INTERVAL select_expression interval_unit RPAREN
  {
    $$ = &FuncExpr{Name: SUBDATE_BYTES, Exprs: append(SelectExprs{$3, $6, &NonStarExpr{Expr: KeywordVal($7)}})}
  }
| TIMESTAMPADD LPAREN time_interval COMMA select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: TIMESTAMPADD_BYTES, Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5...)}
  }
| TIMESTAMPADD LPAREN sql_time_interval COMMA select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: TIMESTAMPADD_BYTES, Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5...)}
  }
| TIMESTAMPDIFF LPAREN time_interval COMMA select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: TIMESTAMPDIFF_BYTES, Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5...)}
  }
| TIMESTAMPDIFF LPAREN sql_time_interval COMMA select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: TIMESTAMPDIFF_BYTES, Exprs: append(SelectExprs{&NonStarExpr{Expr: KeywordVal($3)}}, $5...)}
  }
| TRIM LPAREN select_expression_list RPAREN 
  {
    $$ = &FuncExpr{Name: TRIM_BYTES, Exprs: $3}
  }
| TRIM LPAREN both_leading_trailing_opt select_expression FROM select_expression RPAREN 
  {
    $$ = &FuncExpr{Name: TRIM_BYTES, Exprs: []SelectExpr{$6, &NonStarExpr{Expr: StrVal($3)}, $4}}
  }
| TRIM LPAREN select_expression FROM select_expression RPAREN 
  {
    $$ = &FuncExpr{Name: TRIM_BYTES, Exprs: []SelectExpr{$5, &NonStarExpr{Expr: StrVal(BOTH_BYTES)}, $3}}
  }

| substr LPAREN select_expression FROM select_expression FOR select_expression RPAREN
  {
    $$ = &FuncExpr{Name: $1, Exprs: []SelectExpr{$3,$5,$7}}
  }

| substr LPAREN select_expression FROM select_expression RPAREN
  {
    $$ = &FuncExpr{Name: $1, Exprs: []SelectExpr{$3,$5}}
  }

| substr LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: $1, Exprs: $3} 
  }
| UTC_TIMESTAMP optional_parens
  {
    $$ = &FuncExpr{Name: UTC_TIMESTAMP_BYTES}
  }
| UTC_DATE optional_parens
  {
    $$ = &FuncExpr{Name: UTC_DATE_BYTES}
  }

/*
  function calls using a non reserved keyword, and using a regular syntax.
  Because the non reserved keyword is used in another part of the grammar,
  a dedicated rule is needed here.
*/
func_expr_conflict:
  COUNT LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: COUNT_BYTES, Exprs: $3}
  }
| COUNT LPAREN DISTINCT select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: COUNT_BYTES, Distinct: true, Exprs: $4}
  }
| DATABASE LPAREN RPAREN
  {
    $$ = &FuncExpr{Name: DATABASE_BYTES}
  }
| DATE LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: DATE_BYTES, Exprs: $3}
  }
| DAY LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: DAY_BYTES, Exprs: $3}
  }
| HOUR LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: HOUR_BYTES, Exprs: $3}
  }
| IF LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: IF_BYTES, Exprs: $3}
  }
| INTERVAL LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: INTERVAL_BYTES, Exprs: $3}
  }
| MICROSECOND LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: MICROSECOND_BYTES, Exprs: $3}
  }
| MINUTE LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: MINUTE_BYTES, Exprs: $3}
  }
| MOD LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: MOD_BYTES, Exprs: $3}
  }
| MONTH LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: MONTH_BYTES, Exprs: $3}
  }
| QUARTER LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: QUARTER_BYTES, Exprs: $3}
  }
| SCHEMA LPAREN RPAREN
  {
    $$ = &FuncExpr{Name: SCHEMA_BYTES}
  }
| SECOND LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: SECOND_BYTES, Exprs: $3}
  }
| TIMESTAMP LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: TIMESTAMP_BYTES, Exprs: $3}
  }
| WEEK LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: WEEK_BYTES, Exprs: $3}
  }
| USER LPAREN RPAREN
  {
    $$ = &FuncExpr{Name: USER_BYTES}
  }
| YEAR LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: YEAR_BYTES, Exprs: $3}
  }

/*
  regular function call where the function name is NOT a token.
*/
func_expr_generic:
  ID LPAREN RPAREN
  {
    $$ = &FuncExpr{Name: bytes.ToLower($1)}
  }
| ID LPAREN select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: bytes.ToLower($1), Exprs: $3}
  }
| ID LPAREN DISTINCT select_expression_list RPAREN
  {
    $$ = &FuncExpr{Name: bytes.ToLower($1), Distinct: true, Exprs: $4}
  }

optional_parens:
  {}
| LPAREN RPAREN {}

like_escape_opt:
  {
    $$ = StrVal("\\")
  }
| ESCAPE simple_expr
  {
    $$ = $2
  }
| LBRACE ESCAPE simple_expr RBRACE
  {
    $$ = $3
  }

both_leading_trailing_opt:
  BOTH 
  {
    $$ = BOTH_BYTES
  }
| LEADING
  {
    $$ = LEADING_BYTES
  }
| TRAILING
  {
    $$ = TRAILING_BYTES
  }

interval_unit:
  time_interval
  {
    $$ = $1
  }
| sql_time_unit
  {
    $$ = $1
  }
| sql_time_interval
  {
    $$ = $1
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
  BINARY
    {
      $$ = BINARY_BYTES
    }
  | BINARY LPAREN NUMBER RPAREN
    {
      $$ = BINARY_BYTES
    }
  | CHAR
    {
      $$ = CHAR_BYTES
    }
  | CHAR LPAREN NUMBER RPAREN
    {
      $$ = CHAR_BYTES
    }
  | DATE
    {
      $$ = DATE_BYTES
    }
  | DATETIME
    {
      $$ = DATETIME_BYTES
    }
  | DECIMAL
    {
      $$ = DECIMAL_BYTES
    }
  | DECIMAL LPAREN NUMBER RPAREN
    {
      $$ = DECIMAL_BYTES
    }
  | DECIMAL LPAREN NUMBER COMMA NUMBER RPAREN
    {
      $$ = DECIMAL_BYTES
    }
  | NCHAR
    {
      $$ = CHAR_BYTES
    }
  | NCHAR LPAREN NUMBER RPAREN
    {
      $$ = CHAR_BYTES
    }
  | SIGNED
    {
      $$ = SIGNED_BYTES
    }
  | SIGNED INTEGER
    {
      $$ = SIGNED_BYTES
    }
  | TIME
    {
      $$ = TIME_BYTES
    }
  | UNSIGNED
    {
      $$ = UNSIGNED_BYTES
    }
  | UNSIGNED INTEGER
    {
      $$ = UNSIGNED_BYTES
    }
  | SQL_BIGINT
    {
      $$ = SIGNED_BYTES
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
| sql_id DOT sql_id
  {
    $$ = &ColName{Qualifier: $1, Name: $3}
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
    if bytes.Equal(bytes.ToLower($2), D_BYTES) {
      $$ = DateVal{Name: AST_DATE, Val: $3}
    } else if bytes.Equal(bytes.ToLower($2), T_BYTES) {
      $$ = DateVal{Name: AST_TIME, Val: $3}
    } else if bytes.Equal(bytes.ToLower($2), TS_BYTES) {
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
| LOCK IN ID ID
  {
    if !bytes.Equal($3, SHARE_BYTES) {
      yylex.Error("expecting share")
      return 1
    }
    if !bytes.Equal($4, MODE_BYTES) {
      yylex.Error("expecting mode")
      return 1
    }
    $$ = AST_SHARE_MODE
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
| column_name EQ DEFAULT
  {
    $$ = &UpdateExpr{Name: $1, Expr: StrVal("default")}
  }

sql_id:
  ID
  {
    $$ = $1
  }
| keyword_as_id
  {
    $$ = $1
  }

/*
  keywords that require special treatment for use as identifiers such as table and column names.
*/
keyword_as_id:
  ANY
  {
    $$ = ANY_BYTES
  }
| BINLOG
  {
    $$ = BINLOG_BYTES
  }
| CHANNEL
  {
    $$ = CHANNEL_BYTES
  }
| CHARSET
  {
    $$ = CHARSET_BYTES
  }
| CODE
  {
    $$ = CODE_BYTES
  }
| COLLATION
  {
    $$ = COLLATION_BYTES
  }
| COLUMNS
  {
    $$ = COLUMNS_BYTES
  }
| COMMITTED
  {
    $$ = COMMITTED_BYTES
  }
| CONNECTION
  {
    $$ = CONNECTION_BYTES
  }
| COUNT
  {
    $$ = COUNT_BYTES
  }
| DATE
  {
    $$ = DATE_BYTES
  }
| DATETIME
  {
    $$ = DATETIME_BYTES
  }
| DAY
  {
    $$ = DAY_BYTES
  }
| ENGINE
  {
    $$ = ENGINE_BYTES
  }
| ENGINES
  {
    $$ = ENGINES_BYTES
  }
| ERRORS
  {
    $$ = ERRORS_BYTES
  }
| EVENT
  {
    $$ = EVENT_BYTES
  }
| EVENTS
  {
    $$ = EVENTS_BYTES
  }
| EXTENDED
  {
    $$ = EXTENDED_BYTES
  }
| FORMAT
  {
    $$ = FORMAT_BYTES
  }
| FULL
  {
    $$ = FULL_BYTES
  }
| FUNCTION
  {
    $$ = FUNCTION_BYTES
  }
| GRANTS
  {
    $$ = GRANTS_BYTES
  }
| HOSTS
  {
    $$ = HOSTS_BYTES
  }
| HOUR
  {
    $$ = HOUR_BYTES
  }
| INDEXES
  {
    $$ = INDEXES_BYTES
  }
| ISOLATION
  {
    $$ = ISOLATION_BYTES
  }
| JSON
  {
    $$ = JSON_BYTES
  }
| LEVEL
  {
    $$ = LEVEL_BYTES
  }
| LOGS
  {
    $$ = LOGS_BYTES
  }
| MASTER
  {
    $$ = MASTER_BYTES
  }
| MICROSECOND
  {
    $$ = MICROSECOND_BYTES
  }
| MINUTE
  {
    $$ = MINUTE_BYTES
  }
| MONTH
  {
    $$ = MONTH_BYTES
  }
| MUTEX
  {
    $$ = MUTEX_BYTES
  }
| NAMES
  {
    $$ = NAMES_BYTES
  }
| NCHAR
  {
    $$ = NCHAR_BYTES
  }
| NUMBER
  {
    $$ = NUMBER_BYTES
  }
| OFFSET
  {
    $$ = OFFSET_BYTES
  }
| ONLY
  {
    $$ = ONLY_BYTES
  }
| OPEN
  {
    $$ = OPEN_BYTES
  }
| PARTITIONS
  {
    $$ = PARTITIONS_BYTES
  }
| PLUGINS
  {
    $$ = PLUGINS_BYTES
  }
| PRIVILEGES
  {
    $$ = PRIVILEGES_BYTES
  }
| PROCESSLIST
  {
    $$ = PROCESSLIST_BYTES
  }
| PROFILE
  {
    $$ = PROFILE_BYTES
  }
| PROFILES
  {
    $$ = PROFILES_BYTES
  }
| PROXY
  {
    $$ = PROXY_BYTES
  }
| QUARTER
  {
    $$ = QUARTER_BYTES
  }
| QUERY
  {
    $$ = QUERY_BYTES
  }
| RELAYLOG
  {
    $$ = RELAYLOG_BYTES
  }
| REPEATABLE
  {
    $$ = REPEATABLE_BYTES
  }
| ROW
  {
    $$ = ROW_BYTES
  }
| SECOND
  {
    $$ = SECOND_BYTES
  }
| SERIALIZABLE
  {
    $$ = SERIALIZABLE_BYTES
  }
| SIGNED
  {
    $$ = SIGNED_BYTES
  }
| SLAVE
  {
    $$ = SLAVE_BYTES
  }
| SOME
  {
    $$ = SOME_BYTES
  }
| SQL_TSI_DAY
  {
    $$ = SQL_TSI_DAY_BYTES
  }
| SQL_TSI_HOUR
  {
    $$ = SQL_TSI_HOUR_BYTES
  }
| SQL_TSI_MINUTE
  {
    $$ = SQL_TSI_MINUTE_BYTES
  }
| SQL_TSI_MONTH
  {
    $$ = SQL_TSI_MONTH_BYTES
  }
| SQL_TSI_QUARTER
  {
    $$ = SQL_TSI_QUARTER_BYTES
  }
| SQL_TSI_SECOND
  {
    $$ = SQL_TSI_SECOND_BYTES
  }
| SQL_TSI_WEEK
  {
    $$ = SQL_TSI_WEEK_BYTES
  }
| SQL_TSI_YEAR
  {
    $$ = SQL_TSI_YEAR_BYTES
  }
| STATUS
  {
    $$ = STATUS_BYTES
  }
| STORAGE
  {
    $$ = STORAGE_BYTES
  }
| TABLES
  {
    $$ = TABLES_BYTES
  }
| TEMPORARY
  {
    $$ = TEMPORARY_BYTES
  }
| TIME
  {
    $$ = TIME_BYTES
  }
| TIMESTAMP
  {
    $$ = TIMESTAMP_BYTES
  }
| TIMESTAMPADD
  {
    $$ = TIMESTAMPADD_BYTES
  }
| TIMESTAMPDIFF
  {
    $$ = TIMESTAMPDIFF_BYTES
  }
| TRANSACTION
  {
    $$ = TRANSACTION_BYTES
  }
| TRIGGERS
  {
    $$ = TRIGGERS_BYTES
  }
| UNCOMMITTED
  {
    $$ = UNCOMMITTED_BYTES
  }
| UNKNOWN
  {
    $$ = UNKNOWN_BYTES
  }
| USER
  {
    $$ = USER_BYTES
  }
| VARIABLES
  {
    $$ = VARIABLES_BYTES
  }
| VIEW
  {
    $$ = VIEW_BYTES
  }
| WARNINGS
  {
    $$ = WARNINGS_BYTES
  }
| WEEK
  {
    $$ = WEEK_BYTES
  }
| YEAR
  {
    $$ = YEAR_BYTES
  }
