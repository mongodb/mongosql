// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

%{
package sqlparser

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
  SHARE =        []byte("share")
  MODE  =        []byte("mode")
  IF_BYTES =     []byte("if")
  VALUES_BYTES = []byte("values")
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
  boolExpr    BoolExpr
  valExpr     ValExpr
  tuple       Tuple
  valExprs    ValExprs
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
%token <empty> SELECT INSERT UPDATE DELETE FROM WHERE GROUP HAVING ORDER BY LIMIT FOR SOME ANY TRUE FALSE
%token <empty> ALL DISTINCT PRECISION AS EXISTS IN IS LIKE BETWEEN NULL ASC DESC VALUES INTO DUPLICATE KEY DEFAULT SET LOCK
%token <bytes> ID STRING NUMBER VALUE_ARG COMMENT
%token <empty> LE GE NE NULL_SAFE_EQUAL
%token <empty> '(' '=' '<' '>' '~'
%token <empty> DATE DATETIME TIME TIMESTAMP YEAR

%left <empty> UNION MINUS EXCEPT INTERSECT
%left <empty> ','
%left <empty> JOIN STRAIGHT_JOIN LEFT RIGHT INNER OUTER CROSS NATURAL USE FORCE
%left <empty> ON
%left <empty> AND OR
%right <empty> NOT
%left <empty> '&' '|' '^'
%left <empty> '+' '-'
%left <empty> '*' '/' '%'
%nonassoc <empty> '.'
%left <empty> UNARY
%right <empty> CASE, WHEN, THEN, ELSE
%left <empty> END

// Transaction Tokens
%token <empty> BEGIN COMMIT ROLLBACK

// Charset Tokens
%token <empty> NAMES 

// Replace
%token <empty> REPLACE

// Mixer admin
%token <empty> ADMIN

// Show
%token <empty> SHOW
%token <empty> DATABASES TABLES PROXY VARIABLES FULL COLUMNS

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
%type <str> distinct_opt
%type <selectExprs> select_expression_list
%type <selectExpr> select_expression
%type <bytes> as_lower_opt as_opt
%type <expr> expression
%type <tableExprs> table_expression_list
%type <tableExpr> table_expression
%type <str> join_type
%type <smTableExpr> simple_table_expression
%type <tableName> dml_table_expression
%type <indexHints> index_hint_list
%type <bytes2> index_list
%type <expr> where_expression_opt
%type <boolExpr> boolean_expression condition
%type <str> compare
%type <insRows> row_list
%type <valExpr> value value_expression
%type <tuple> tuple
%type <valExprs> value_expression_list
%type <values> tuple_list
%type <bytes> keyword_as_func
%type <subquery> subquery
%type <byt> unary_operator
%type <colName> column_name
%type <caseExpr> case_expression
%type <whens> when_expression_list
%type <when> when_expression
%type <valExpr> value_expression_opt else_expression_opt
%type <valExprs> group_by_opt
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
%type <empty> exists_opt not_exists_opt ignore_opt non_rename_operation to_opt constraint_opt using_opt
%type <bytes> sql_id
%type <str> ctor_type time_type
%type <empty> force_eof

%type <statement> begin_statement commit_statement rollback_statement
%type <statement> replace_statement
%type <statement> show_statement
%type <statement> admin_statement

%type <valExpr> from_opt
%type <expr> like_or_where_opt
%type <valExpr> show_from_in show_from_in_opt
%type <str> show_full
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
| rollback_statement
| replace_statement
| show_statement
| admin_statement

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
    $$ = &Set{Comments: Comments($2), Exprs: UpdateExprs{&UpdateExpr{Name: &ColName{Name:[]byte("names")}, Expr: StrVal($4)}}}
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
  ADMIN sql_id '(' value_expression_list ')'
  {
    $$ = &Admin{Name : $2, Values : $4}
  }

show_from_in:
  FROM value_expression
  {
    $$ = $2
  }
| IN value_expression
  {
    $$ = $2
  }

show_from_in_opt:
  {
    $$ = nil
  }
| FROM value_expression
  {
    $$ = $2
  }
| IN value_expression
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
  

show_statement:
  SHOW DATABASES like_or_where_opt 
  {
    $$ = &Show{Section: "databases", LikeOrWhere: $3}
  }
| SHOW VARIABLES like_or_where_opt 
  {
    $$ = &Show{Section: "variables", LikeOrWhere: $3}
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
| select_expression_list ',' select_expression
  {
    $$ = append($$, $3)
  }

select_expression:
  '*'
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
| ID '.' '*'
  {
    $$ = &StarExpr{TableName: $1}
  }

expression:
  boolean_expression
  {
    $$ = $1
  }
| value_expression
  {
    $$ = $1
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
| AS ctor_type
  {
    $$ = []byte($2)
  }


table_expression_list:
  table_expression
  {
    $$ = TableExprs{$1}
  }
| table_expression_list ',' table_expression
  {
    $$ = append($$, $3)
  }

table_expression:
  simple_table_expression as_opt index_hint_list
  {
    $$ = &AliasedTableExpr{Expr:$1, As: $2, Hints: $3}
  }
| '(' table_expression ')'
  {
    $$ = &ParenTableExpr{Expr: $2}
  }
| table_expression join_type table_expression %prec JOIN
  {
    $$ = &JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3}
  }
| table_expression join_type table_expression ON boolean_expression %prec JOIN
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
| ID '.' ID
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
| ID '.' COLUMNS // hack for tokenizer, maybe cleaner way
  {
    $$ = &TableName{Qualifier: $1, Name: []byte("columns")}
  }
| ID '.' TABLES // hack for tokenizer, maybe cleaner way
  {
    $$ = &TableName{Qualifier: $1, Name: []byte("tables")}
  }

dml_table_expression:
ID
  {
    $$ = &TableName{Name: $1}
  }
| ID '.' ID
  {
    $$ = &TableName{Qualifier: $1, Name: $3}
  }

index_hint_list:
  {
    $$ = nil
  }
| USE INDEX '(' index_list ')'
  {
    $$ = &IndexHints{Type: AST_USE, Indexes: $4}
  }
| IGNORE INDEX '(' index_list ')'
  {
    $$ = &IndexHints{Type: AST_IGNORE, Indexes: $4}
  }
| FORCE INDEX '(' index_list ')'
  {
    $$ = &IndexHints{Type: AST_FORCE, Indexes: $4}
  }

index_list:
  sql_id
  {
    $$ = [][]byte{$1}
  }
| index_list ',' sql_id
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
| LIKE value_expression
  {
    $$ = $2
  }

from_opt:
  {
    $$ = nil
  }
| FROM value_expression
  {
    $$ = $2
  }

boolean_expression:
  condition
| expression AND expression
  {
    $$ = &AndExpr{Left: $1, Right: $3}
  }
| expression OR expression
  {
    $$ = &OrExpr{Left: $1, Right: $3}
  }
| NOT expression
  {
    $$ = &NotExpr{Expr: $2}
  }
| '(' boolean_expression ')'
  {
    $$ = &ParenBoolExpr{Expr: $2}
  }

condition:
  value_expression compare value_expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: $2, Right: $3}
  }
| value_expression IN tuple
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_IN, Right: $3}
  }
| value_expression compare ALL subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: $2, SubqueryOperator: AST_ALL, Right: $4}
  }
| value_expression compare ANY subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: $2, SubqueryOperator: AST_ANY, Right: $4}
  }
| value_expression compare SOME subquery
  {
    $$ = &ComparisonExpr{Left: $1, Operator: $2, SubqueryOperator: AST_SOME, Right: $4}
  }
| value_expression NOT IN tuple
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NOT_IN, Right: $4}
  }
| value_expression LIKE value_expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_LIKE, Right: $3}
  }
| value_expression NOT LIKE value_expression
  {
    $$ = &ComparisonExpr{Left: $1, Operator: AST_NOT_LIKE, Right: $4}
  }
| value_expression BETWEEN value_expression AND value_expression
  {
    $$ = &RangeCond{Left: $1, Operator: AST_BETWEEN, From: $3, To: $5}
  }
| value_expression NOT BETWEEN value_expression AND value_expression
  {
    $$ = &RangeCond{Left: $1, Operator: AST_NOT_BETWEEN, From: $4, To: $6}
  }
| value_expression IS NULL
  {
    $$ = &NullCheck{Operator: AST_IS_NULL, Expr: $1}
  }
| value_expression IS NOT NULL
  {
    $$ = &NullCheck{Operator: AST_IS_NOT_NULL, Expr: $1}
  }
| EXISTS subquery
  {
    $$ = &ExistsExpr{Subquery: $2}
  }

compare:
  '='
  {
    $$ = AST_EQ
  }
| '<'
  {
    $$ = AST_LT
  }
| '>'
  {
    $$ = AST_GT
  }
| LE
  {
    $$ = AST_LE
  }
| GE
  {
    $$ = AST_GE
  }
| NE
  {
    $$ = AST_NE
  }
| NULL_SAFE_EQUAL
  {
    $$ = AST_NSE
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
| tuple_list ',' tuple
  {
    $$ = append($1, $3)
  }

tuple:
  '(' value_expression_list ')'
  {
    $$ = ValTuple($2)
  }
| subquery
  {
    $$ = $1
  }

subquery:
  '(' select_statement ')'
  {
    $$ = &Subquery{$2}
  }

value_expression_list:
  value_expression
  {
    $$ = ValExprs{$1}
  }
| value_expression_list ',' value_expression
  {
    $$ = append($1, $3)
  }

value_expression:
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
| value_expression '&' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITAND, Right: $3}
  }
| value_expression '|' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITOR, Right: $3}
  }
| value_expression '^' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_BITXOR, Right: $3}
  }
| value_expression '+' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_PLUS, Right: $3}
  }
| value_expression '-' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MINUS, Right: $3}
  }
| value_expression '*' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MULT, Right: $3}
  }
| value_expression '/' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_DIV, Right: $3}
  }
| value_expression '%' value_expression
  {
    $$ = &BinaryExpr{Left: $1, Operator: AST_MOD, Right: $3}
  }
| unary_operator value_expression %prec UNARY
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
| sql_id '(' ')'
  {
    $$ = &FuncExpr{Name: bytes.ToLower($1)}
  }
| sql_id '(' select_expression_list ')'
  {
    $$ = &FuncExpr{Name: bytes.ToLower($1), Exprs: $3}
  }
| sql_id '(' DISTINCT select_expression_list ')'
  {
    $$ = &FuncExpr{Name: bytes.ToLower($1), Distinct: true, Exprs: $4}
  }
| '(' ctor_type value_expression_list ')'
  {
    $$ = &CtorExpr{Name: $2, Exprs: $3}
  }
| keyword_as_func '(' select_expression_list ')'
  {
    $$ = &FuncExpr{Name: $1, Exprs: $3}
  }
| case_expression
  {
    $$ = $1
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

unary_operator:
  '+'
  {
    $$ = AST_UPLUS
  }
| '-'
  {
    $$ = AST_UMINUS
  }
| '~'
  {
    $$ = AST_TILDA
  }

case_expression:
  CASE value_expression_opt when_expression_list else_expression_opt END
  {
    $$ = &CaseExpr{Expr: $2, Whens: $3, Else: $4}
  }

value_expression_opt:
  {
    $$ = nil
  }
| value_expression
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
  WHEN boolean_expression THEN value_expression
  {
    $$ = &When{Cond: $2, Val: $4}
  }
| WHEN value_expression THEN value_expression
  {
    $$ = &When{Cond: $2, Val: $4}
  }

else_expression_opt:
  {
    $$ = nil
  }
| ELSE value_expression
  {
    $$ = $2
  }

column_name:
  sql_id
  {
    $$ = &ColName{Name: $1}
  }
| ID '.' sql_id
  {
    $$ = &ColName{Qualifier: $1, Name: $3}
  }

value:
TRUE
  {
    $$ = &TrueVal{}
  }
| FALSE
  {
    $$ = &FalseVal{}
  }
| STRING
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
| NULL
  {
    $$ = &NullVal{}
  }

group_by_opt:
  {
    $$ = nil
  }
| GROUP BY value_expression_list
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
| order_list ',' order
  {
    $$ = append($1, $3)
  }

order:
  value_expression asc_desc_opt
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
| LIMIT value_expression
  {
    $$ = &Limit{Rowcount: $2}
  }
| LIMIT value_expression ',' value_expression
  {
    $$ = &Limit{Offset: $2, Rowcount: $4}
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
| '(' column_list ')'
  {
    $$ = $2
  }

column_list:
  column_name
  {
    $$ = Columns{&NonStarExpr{Expr: $1}}
  }
| column_list ',' column_name
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
| update_list ',' update_expression
  {
    $$ = append($1, $3)
  }

update_expression:
  column_name '=' value_expression
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

ctor_type:
  time_type
  {
    $$ = $1
  }

time_type:
  DATE
  {
    $$ = AST_DATE
  }
| TIME
  {
    $$ = AST_TIME
  }
| TIMESTAMP
  {
    $$ = AST_TIMESTAMP
  }
| DATETIME
  {
    $$ = AST_DATETIME
  }
| YEAR
  {
    $$ = AST_YEAR
  }

force_eof:
{
  ForceEOF(yylex)
}
