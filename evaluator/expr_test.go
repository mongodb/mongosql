package evaluator_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/collation"
	. "github.com/10gen/sqlproxy/evaluator"
	. "github.com/10gen/sqlproxy/evaluator/results"
	. "github.com/10gen/sqlproxy/evaluator/types"
	. "github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/shopspring/decimal"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	valKind = MySQLValueKind
	knd     = valKind
)

func TestEvaluates(t *testing.T) {
	req := require.New(t)

	type test struct {
		name   string
		sql    string
		result SQLValue
	}

	runTests := func(t *testing.T, cfg *ExecutionConfig, st *ExecutionState, tests []test) {
		schema := MustLoadSchema(testSchema3)
		oCfg := CreateTestOptimizerCfg(collation.Default, cfg)
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req = require.New(t)
				subject, err := GetSQLExpr(schema, dbOne, tableTwoName, test.sql, true, oCfg)
				req.Nil(err, "unable to get SQLExpr for sql statement")

				result, err := subject.Evaluate(context.Background(), cfg, st)
				req.Nil(err, "unable to evaluate SQLExpr")

				if test.result.IsNull() {
					actualVal, ok := result.(SQLValue)
					req.True(ok, "expected result to be a SQLValue")
					req.True(actualVal.IsNull(), fmt.Sprintf("expected result to be null, but got %#v", actualVal))
				} else {
					req.Equal(
						test.result, result,
						"expected SQLValue does not match evaluated SQLValue",
					)
				}
			})
		}
	}

	type typeTest struct {
		name   string
		sql    string
		result EvalType
	}

	runTypeTests := func(t *testing.T, tests []typeTest) {
		sc := MustLoadSchema(testSchema3)
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req = require.New(t)
				subject, err := GetSQLExpr(sc, dbOne, tableTwoName, test.sql, false, nil)
				req.Nil(err, "unable to get SQLExpr for sql statement")
				result := subject.EvalType()
				req.Equal(
					test.result,
					result,
					"type of evaluated SQLExpr does not match expected type",
				)
			})
		}
	}

	type errTest struct {
		name  string
		expr  SQLExpr
		valid bool
	}

	runEvaluateErrTests := func(t *testing.T, cfg *ExecutionConfig, st *ExecutionState, tests []errTest) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req = require.New(t)
				_, err := test.expr.Evaluate(context.Background(), cfg, st)
				if test.valid {
					req.Nil(err, "evaluate should not return an error for these types")
				} else {
					req.NotNil(err, "evaluate should return an error for these types")
				}
			})
		}
	}

	runFoldConstantsErrTests := func(t *testing.T, eCfg *ExecutionConfig, tests []errTest) {
		oCfg := CreateTestOptimizerCfg(collation.Default, eCfg)
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req = require.New(t)
				_, err := test.expr.FoldConstants(oCfg)
				if test.valid {
					req.Nil(err, "evaluate should not return an error for these types")
				} else {
					req.NotNil(err, "evaluate should return an error for these types")
				}
			})
		}
	}

	execCfg := createTestExecutionCfg(knd)

	t.Run("evaluates", func(t *testing.T) {
		row := &Row{
			Data: RowValues{
				{
					SelectID: 1,
					Database: "test",
					Table:    "bar",
					Name:     "a",
					Data:     NewSQLInt64(knd, 123),
				},
				{
					SelectID: 1,
					Database: "test",
					Table:    "bar",
					Name:     "b",
					Data:     NewSQLInt64(knd, 456),
				},
				{
					SelectID: 1,
					Database: "test",
					Table:    "bar",
					Name:     "c",
					Data:     NewSQLNull(knd),
				},
			},
		}

		bgCtx := context.Background()
		execState := NewExecutionState().WithRows(row)

		// defines the scalar functions expressions to evaluates, along with
		// the name for the test and the expected result
		tests := []test{
			{"sql_add_expr_int_0", "0 + 0", NewSQLInt64(knd, 0)},
			{"sql_add_expr_int_1", "-1 + 1", NewSQLInt64(knd, 0)},
			{"sql_add_expr_int_2", "10 + 32", NewSQLInt64(knd, 42)},
			{"sql_add_expr_int_3", "-10 + -32", NewSQLInt64(knd, -42)},
			{"sql_add_expr_bool_0", "true + true", NewSQLInt64(knd, 2)},
			{"sql_add_expr_bool_1", "true + true + false", NewSQLInt64(knd, 2)},
			{"sql_add_expr_bool_2", "false + true + true", NewSQLInt64(knd, 2)},
			{"sql_add_expr_mixed_0", "true - '-1'", NewSQLFloat(knd, 2)},
			{"sql_add_expr_mixed_1", "true + '0'", NewSQLFloat(knd, 1)},
			{"sql_and_expr_0", "1 AND 1", NewSQLBool(knd, true)},
			{"sql_and_expr_1", "1 AND 0", NewSQLBool(knd, false)},
			{"sql_and_expr_2", "0 AND 1", NewSQLBool(knd, false)},
			{"sql_and_expr_3", "0 AND 0", NewSQLBool(knd, false)},
			{"sql_and_expr_4", "1 && 1", NewSQLBool(knd, true)},
			{"sql_and_expr_5", "1 && 0", NewSQLBool(knd, false)},
			{"sql_and_expr_6", "0 && 1", NewSQLBool(knd, false)},
			{"sql_and_expr_7", "0 && 0", NewSQLBool(knd, false)},
			{"sql_and_expr_8", "1 && 'bar'", NewSQLBool(knd, false)},
			{"sql_and_expr_9", "true && 'bar'", NewSQLBool(knd, false)},
			{"sql_and_expr_with_null_0", "NULL && 0", NewSQLBool(knd, false)},
			{"sql_and_expr_with_null_1", "NULL && 1", NewSQLNull(knd)},
			{"sql_and_expr_with_null_2", "NULL && NULL", NewSQLNull(knd)},
			{"sql_and_expr_with_bool_0", "true AND true", NewSQLBool(knd, true)},
			{"sql_and_expr_with_bool_1", "true AND false", NewSQLBool(knd, false)},
			{"sql_and_expr_with_bool_2", "false AND true", NewSQLBool(knd, false)},
			{"sql_and_expr_with_bool_3", "false AND false", NewSQLBool(knd, false)},
			{"sql_benchmark_expr_0", "BENCHMARK(10, 1)", NewSQLInt64(knd, 0)},
			{"sql_benchmark_expr_1", "BENCHMARK(0, 10)", NewSQLInt64(knd, 0)},
			{"sql_benchmark_expr_2", "BENCHMARK(NULL, 0)", NewSQLInt64(knd, 0)},
			{"sql_date_expr_with_add_0", "DATE '2014-04-13' + 0", NewSQLInt64(knd, 20140413)},
			{"sql_date_expr_with_add_1", "DATE '2014-04-13' + 2", NewSQLInt64(knd, 20140415)},
			// Need to support Time type
			// {
			// 	"sql_time_expr_with_add_0",
			// 	"TIME '11:04:13' + 0",
			// 	NewSQLDecimal128(knd, decimal.NewFromFloat(110413)),
			// },
			// {
			// 	"sql_time_expr_with_add_1",
			// 	"TIME '11:04:13' + 2",
			// 	NewSQLDecimal128(knd, decimal.NewFromFloat(110415)),
			// },
			// {
			// 	"sql_time_expr_with_add_2",
			// 	"TIME '11:04:13' + '2'",
			// 	NewSQLDecimal128(knd, decimal.NewFromFloat(110415)),
			// },
			// {
			// 	"sql_time_expr_with_add_3",
			// 	"'2' + TIME '11:04:13'",
			// 	NewSQLDecimal128(knd, decimal.NewFromFloat(110415)),
			// },
			// {
			// 	"sql_time_expr_with_subtract_0",
			// 	"TIME '11:04:13' - 0",
			// 	NewSQLDecimal128(knd, decimal.NewFromFloat(110413)),
			// },
			// {
			// 	"sql_time_expr_with_subtract_1",
			// 	"TIME '11:04:13' - 2",
			// 	NewSQLDecimal128(knd, decimal.NewFromFloat(110411)),
			// },
			// {
			// 	"sql_time_expr_with_subtract_2",
			// 	"TIME '11:04:13' - '2'",
			// 	NewSQLDecimal128(knd, decimal.NewFromFloat(110411)),
			// },
			// {
			// 	"sql_time_expr_with_multiply_0",
			// 	"TIME '11:04:13' * 0",
			// 	NewSQLDecimal128(knd, decimal.NewFromFloat(0)),
			// },
			// {
			// 	"sql_time_expr_with_multiply_1",
			// 	"TIME '11:04:13' * 2",
			// 	NewSQLDecimal128(knd, decimal.NewFromFloat(220826)),
			// },
			// {
			// 	"sql_time_expr_with_multiply_2",
			// 	"TIME '11:04:13' * '2'",
			// 	NewSQLDecimal128(knd, decimal.NewFromFloat(220826)),
			// },
			// {"sql_time_division_0", "TIME '11:04:13' / 0", NewSQLNull(knd)},
			// {
			// 	"sql_time_division_1",
			// 	"TIME '11:04:13' / 2",
			// 	NewSQLDecimal128(knd, decimal.New(552065000, -4)),
			// },
			// {
			// 	"sql_time_division_2",
			// 	"TIME '11:04:13' / '2'",
			// 	NewSQLDecimal128(knd, decimal.New(552065000, -4)),
			// },
			{
				"sql_timestamp_expr_with_add_0",
				"TIMESTAMP '2014-04-13 11:04:13' + 0",
				NewSQLDecimal128(knd, decimal.NewFromFloat(20140413110413)),
			},
			{
				"sql_timestamp_expr_with_add_1",
				"TIMESTAMP '2014-04-13 11:04:13' + 2",
				NewSQLDecimal128(knd, decimal.NewFromFloat(20140413110415)),
			},
			{"sql_date_expr_with_subtract_0", "DATE '2014-04-13' - 0",
				NewSQLInt64(knd, 20140413)},
			{"sql_date_expr_with_subtract_1", "DATE '2014-04-13' - 2",
				NewSQLInt64(knd, 20140411)},
			{
				"sql_timestamp_expr_with_subtract_0",
				"TIMESTAMP '2014-04-13 11:04:13' - 0",
				NewSQLDecimal128(knd, decimal.NewFromFloat(20140413110413)),
			},
			{
				"sql_timestamp_expr_with_subtract_1",
				"TIMESTAMP '2014-04-13 11:04:13' - 2",
				NewSQLDecimal128(knd, decimal.NewFromFloat(20140413110411)),
			},
			{"sql_date_expr_with_multiply_0", "DATE '2014-04-13' * 0", NewSQLInt64(knd, 0)},
			{"sql_date_expr_with_multiply_0", "DATE '2014-04-13' * 2",
				NewSQLInt64(knd, 40280826)},
			{
				"sql_timestamp_expr_with_multiply_0",
				"TIMESTAMP '2014-04-13 11:04:13' * 0",
				NewSQLDecimal128(knd, decimal.NewFromFloat(0)),
			},
			{
				"sql_timestamp_expr_with_multiply_1",
				"TIMESTAMP '2014-04-13 11:04:13' * 2",
				NewSQLDecimal128(knd, decimal.NewFromFloat(40280826220826)),
			},
			{
				"sql_float_division_expr_0",
				"1.2 / 0.2",
				NewSQLDecimal128(knd, decimal.New(600000, -5)),
			},
			{
				"sql_float_division_expr_1",
				"1.2 / 0.23",
				NewSQLDecimal128(knd, decimal.New(521739, -5)),
			},
			{"sql_date_division_0", "DATE '2014-04-13' / 0", NewSQLNull(knd)},
			{"sql_date_division_1", "DATE '2014-04-13' / 2", NewSQLFloat(knd, 10070206.5)},
			{
				"sql_timestamp_division_0",
				"TIMESTAMP '2014-04-13 11:04:13' / 0",
				NewSQLNull(knd),
			},
			{
				"sql_timestamp_division_1",
				"TIMESTAMP '2014-04-13 11:04:13' / 2",
				NewSQLDecimal128(knd, decimal.New(100702065552065000, -4)),
			},
			{"sql_date_less_than_0", "DATE '2014-04-13' < 0", NewSQLBool(knd, false)},
			{
				"sql_date_less_than_1",
				"DATE '2014-04-13' < DATE '2014-04-14'",
				NewSQLBool(knd, true),
			},
			{"sql_date_greater_than_0", "DATE '2014-04-13' > 0", NewSQLBool(knd, true)},
			{
				"sql_date_greater_than_1",
				"DATE '2014-04-13' > DATE '2014-04-14'",
				NewSQLBool(knd, false),
			},
			{"sql_date_equality_0", "DATE '2014-04-13' = '0'", NewSQLNull(knd)},
			{"sql_date_equality_1", "DATE '2014-04-13' = DATE '2014-04-13'", NewSQLBool(knd, true)},
			{"sql_date_equality_2", "DATE '2014-04-13' = 0", NewSQLBool(knd, false)},
			{
				"sql_case_expr_0",
				"CASE 3 WHEN 3 THEN 'three' WHEN 1 THEN 'one' ELSE 'else' END",
				NewSQLVarchar(knd, "three"),
			},
			{
				"sql_case_expr_1",
				"CASE WHEN 5 > 3 THEN 'true' else 'false' END",
				NewSQLVarchar(knd, "true"),
			},
			{
				"sql_case_expr_2",
				"CASE WHEN a = 123 THEN 'yes' else 'no' END",
				NewSQLVarchar(knd, "yes"),
			},
			{"sql_conv_0", "conv('12A', 13, 7)", NewSQLVarchar(knd, "412")},
			{"sql_conv_1", "conv('345', 8, 10)", NewSQLVarchar(knd, "229")},
			{"sql_conv_2", "conv('123.-/', 10, 6)", NewSQLVarchar(knd, "323")},
			{"sql_conv_3", "conv('4A', 10, 2)", NewSQLVarchar(knd, "0")},
			{"sql_conv_4", "conv('23-', 5, 3)", NewSQLVarchar(knd, "0")},
			{"sql_conv_5", "conv('74', 48, 3)", NewSQLNull(knd)},
			{"sql_conv_6", "conv('13', 1, 7)", NewSQLNull(knd)},
			{"sql_conv_7", "conv('-34', 5, 11)", NewSQLVarchar(knd, "-18")},
			{"sql_conv_8", "conv('-0', 5, 7)", NewSQLVarchar(knd, "0")},
			{"sql_case_expr_3", "CASE WHEN a = 245 THEN 'yes' END", NewSQLNull(knd)},
			{"sql_int_division_expr_0", "-1 / 1", NewSQLFloat(knd, -1)},
			{"sql_int_division_expr_1", "100 / 10", NewSQLFloat(knd, 10)},
			{"sql_int_division_expr_2", "-10 / 10", NewSQLFloat(knd, -1)},
			{"sql_int_equality_expr_0", "0 = 0", NewSQLBool(knd, true)},
			{"sql_int_equality_expr_1", "-1 = 1", NewSQLBool(knd, false)},
			{"sql_int_equality_expr_2", "10 = 10", NewSQLBool(knd, true)},
			{"sql_int_equality_expr_3", "-10 = -10", NewSQLBool(knd, true)},
			{"sql_mixed_equality_expr", "false = '0'", NewSQLBool(knd, true)},
			{"sql_int_comparison_0", "0 > 0", NewSQLBool(knd, false)},
			{"sql_int_comparison_1", "-1 > 1", NewSQLBool(knd, false)},
			{"sql_int_comparison_2", "1 > -1", NewSQLBool(knd, true)},
			{"sql_int_comparison_3", "11 > 10", NewSQLBool(knd, true)},
			{"sql_mixed_comparison_0", "true > '-1'", NewSQLBool(knd, true)},
			{"sql_int_greater_than_0", "0 >= 0", NewSQLBool(knd, true)},
			{"sql_int_greater_than_1", "-1 >= 1", NewSQLBool(knd, false)},
			{"sql_int_greater_than_2", "1 >= -1", NewSQLBool(knd, true)},
			{"sql_int_greater_than_3", "11 >= 10", NewSQLBool(knd, true)},
			{"sql_is_bool_expr_0", "1 is true", NewSQLBool(knd, true)},
			{"sql_is_bool_expr_1", "null is true", NewSQLBool(knd, false)},
			{"sql_is_unknown_expr_0", "null is unknown", NewSQLBool(knd, true)},
			{"sql_is_unknown_expr_1", "1 is unknown", NewSQLBool(knd, false)},
			{"sql_is_expr_0", "true is true", NewSQLBool(knd, true)},
			{"sql_is_expr_1", "0 is false", NewSQLBool(knd, true)},
			{"sql_is_expr_2", "1 is false", NewSQLBool(knd, false)},
			{"sql_is_expr_3", "'1' is true", NewSQLBool(knd, true)},
			{"sql_is_expr_4", "'0.0' is true", NewSQLBool(knd, false)},
			{"sql_is_expr_5", "'cats' is false", NewSQLBool(knd, true)},
			{"sql_date_is_bool_expr_0", "DATE '2006-05-04' is false", NewSQLBool(knd, false)},
			{
				"sql_date_is_bool_expr_1",
				"TIMESTAMP '2008-04-06 15:32:23' is true",
				NewSQLBool(knd, true),
			},
			{"sql_is_null_expr_0", "1 is null", NewSQLBool(knd, false)},
			{"sql_is_null_expr_1", "null is null", NewSQLBool(knd, true)},
			{"sql_is_not_expr_1", "1 is not true", NewSQLBool(knd, false)},
			{"sql_is_not_expr_2", "null is not true", NewSQLBool(knd, true)},
			{"sql_is_not_expr_3", "null is not unknown", NewSQLBool(knd, false)},
			{"sql_is_not_expr_4", "1 is not unknown", NewSQLBool(knd, true)},
			{"sql_is_not_expr_5", "false is not true", NewSQLBool(knd, true)},
			{"sql_is_not_expr_6", "0 is not false", NewSQLBool(knd, false)},
			{"sql_is_not_expr_7", "1 is not false", NewSQLBool(knd, true)},
			{"sql_is_not_expr_8", "'1' is not true", NewSQLBool(knd, false)},
			{"sql_is_not_expr_9", "'0.0' is not true", NewSQLBool(knd, true)},
			{"sql_is_not_expr_10", "'cats' is not false", NewSQLBool(knd, false)},
			{"sql_is_not_expr_11", "DATE '2006-05-04' is not false", NewSQLBool(knd, true)},
			{
				"sql_is_not_expr_12",
				"TIMESTAMP '2008-04-06 15:32:23' is not true",
				NewSQLBool(knd, false),
			},
			{"sql_is_not_expr_13", "1 is not null", NewSQLBool(knd, true)},
			{"sql_is_not_expr_14", "null is not null", NewSQLBool(knd, false)},
			{"sql_divide_expr_0", "0 DIV 0", NewSQLNull(knd)},
			{"sql_divide_expr_1", "0 DIV 5", NewSQLInt64(knd, 0)},
			{"sql_divide_expr_2", "5.5 DIV 2", NewSQLInt64(knd, 2)},
			{"sql_divide_expr_3", "-5 DIV 2", NewSQLInt64(knd, -2)},
			{"sql_divide_expr_4", "NULL DIV 1", NewSQLNull(knd)},
			{"sql_divide_expr_5", "1 DIV NULL", NewSQLNull(knd)},
			{"sql_divide_expr_6", "2 DIV 1.2", NewSQLInt64(knd, 1)},
			{"sql_in_expr_0", "0 IN(0)", NewSQLBool(knd, true)},
			{"sql_in_expr_1", "-1 IN(1)", NewSQLBool(knd, false)},
			{"sql_in_expr_2", "0 IN(10, 0)", NewSQLBool(knd, true)},
			{"sql_in_expr_3", "-1 IN(1, 10)", NewSQLBool(knd, false)},
			{"sql_in_expr_4", "NULL IN(0, 1)", NewSQLNull(knd)},
			{"sql_in_expr_5", "NULL IN(0, NULL)", NewSQLNull(knd)},
			{"sql_less_than_expr_0", "0 < 0", NewSQLBool(knd, false)},
			{"sql_less_than_expr_1", "-1 < 1", NewSQLBool(knd, true)},
			{"sql_less_than_expr_2", "1 < -1", NewSQLBool(knd, false)},
			{"sql_less_than_expr_3", "10 < 11", NewSQLBool(knd, true)},
			{"sql_less_than_expr_4", "true < '5'", NewSQLBool(knd, true)},
			{"sql_less_than_or_equal_expr_0", "0 <= 0", NewSQLBool(knd, true)},
			{"sql_less_than_or_equal_expr_1", "-1 <= 1", NewSQLBool(knd, true)},
			{"sql_less_than_or_equal_expr_2", "1 <= -1", NewSQLBool(knd, false)},
			{"sql_less_than_or_equal_expr_3", "10 <= 11", NewSQLBool(knd, true)},
			{"sql_like_expr_0", "'Á Â Ã Ä' LIKE '%'", NewSQLBool(knd, true)},
			{"sql_like_expr_1", "'Á Â Ã Ä' LIKE 'Á Â Ã Ä'", NewSQLBool(knd, true)},
			{"sql_like_expr_2", "'Á Â Ã Ä' LIKE 'Á%'", NewSQLBool(knd, true)},
			{"sql_like_expr_3", "'a' LIKE 'a'", NewSQLBool(knd, true)},
			{"sql_like_expr_4", "'Adam' LIKE 'am'", NewSQLBool(knd, false)},
			{"sql_like_expr_5", "'Adam' LIKE 'adaM'", NewSQLBool(knd, true)}, // mixed case test
			{"sql_like_expr_6", "'Adam' LIKE '%am%'", NewSQLBool(knd, true)},
			{"sql_like_expr_7", "'Adam' LIKE 'Ada_'", NewSQLBool(knd, true)},
			{"sql_like_expr_8", "'Adam' LIKE '__am'", NewSQLBool(knd, true)},
			{"sql_like_expr_9", "'Clever' LIKE '%is'", NewSQLBool(knd, false)},
			{"sql_like_expr_10", "'Adam is nice' LIKE '%xs '", NewSQLBool(knd, false)},
			{"sql_like_expr_11", "'Adam is nice' LIKE '%is nice'", NewSQLBool(knd, true)},
			{"sql_like_expr_12", "'abc' LIKE 'ABC'", NewSQLBool(knd, true)},    //case insensitive
			{"sql_like_expr_13", "'abc' LIKE 'ABC '", NewSQLBool(knd, false)},  // trailing space
			{"sql_like_expr_14", "'abc' LIKE ' ABC'", NewSQLBool(knd, false)},  // leading space
			{"sql_like_expr_15", "'abc' LIKE ' ABC '", NewSQLBool(knd, false)}, // padded space
			{"sql_like_expr_16", "'abc' LIKE 'ABC	'", NewSQLBool(knd, false)}, // trailing tab
			{"sql_like_expr_17", "'10' LIKE '1%'", NewSQLBool(knd, true)},
			{"sql_like_expr_18", "'a   ' LIKE 'A   '", NewSQLBool(knd, true)},
			{"sql_like_expr_19", "CURRENT_DATE() LIKE '2015-05-31%'", NewSQLBool(knd, false)},
			{"sql_like_expr_20", "CURDATE() LIKE '2015-05-31%'", NewSQLBool(knd, false)},
			{"sql_like_expr_21", "(DATE '2008-01-02') LIKE '2008-01%'", NewSQLBool(knd, true)},
			{
				"sql_like_expr_22",
				"NOW() LIKE '" + strconv.Itoa(time.Now().Year()) + "%' ",
				NewSQLBool(knd, true),
			},
			{"sql_like_expr_23", "10 LIKE '1%'", NewSQLBool(knd, true)},
			{"sql_like_expr_24", "1.20 LIKE '1.2%'", NewSQLBool(knd, true)},
			{"sql_like_expr_25", "NULL LIKE '1%'", NewSQLNull(knd)},
			{"sql_like_expr_26", "10 LIKE NULL", NewSQLNull(knd)},
			{"sql_like_expr_27", "NULL LIKE NULL", NewSQLNull(knd)},
			{"sql_like_expr_28", "'David_' LIKE 'David\\_'", NewSQLBool(knd, true)},
			{"sql_like_expr_29", "'David%' LIKE 'David\\%'", NewSQLBool(knd, true)},
			{"sql_like_expr_30", "'David_' LIKE 'David|_' ESCAPE '|'", NewSQLBool(knd, true)},
			{"sql_like_expr_31", "'David\\_' LIKE 'David\\_' ESCAPE ''", NewSQLBool(knd, true)},
			{"sql_like_expr_32", "'David_' LIKE 'David\\_' ESCAPE char(92)", NewSQLBool(knd, true)},
			{"sql_like_expr_33", "'David_' LIKE 'David|_' {escape '|'}", NewSQLBool(knd, true)},
			{"sql_like_expr_34", "'hello' LIKE concat('h_','llo')", NewSQLBool(knd, true)}, // Constant with function to match on
			{"sql_like_binary_expr_0", "'Á Â Ã Ä' LIKE BINARY '%'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_1", "'Á Â Ã Ä' LIKE BINARY 'Á Â Ã Ä'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_2", "'Á Â Ã Ä' LIKE BINARY 'Á%'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_3", "'a' LIKE BINARY 'a'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_4", "'Adam' LIKE BINARY 'am'", NewSQLBool(knd, false)},
			{"sql_like_binary_expr_5", "'Adam' LIKE BINARY 'adaM'", NewSQLBool(knd, false)},
			{"sql_like_binary_expr_6", "'Adam' LIKE BINARY '%am%'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_7", "'Adam' LIKE BINARY 'Ada_'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_8", "'Adam' LIKE BINARY '__am'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_9", "'Clever' LIKE BINARY '%is'", NewSQLBool(knd, false)},
			{"sql_like_binary_expr_10", "'Adam is nice' LIKE BINARY '%xs '",
				NewSQLBool(knd, false)},
			{"sql_like_binary_expr_11", "'Adam is nice' LIKE BINARY '%is nice'",
				NewSQLBool(knd, true)},
			{"sql_like_binary_expr_12", "'abc' LIKE BINARY 'ABC'", NewSQLBool(knd, false)},
			{"sql_like_binary_expr_13", "'abc' LIKE BINARY 'ABC '", NewSQLBool(knd, false)},
			{"sql_like_binary_expr_14", "'abc' LIKE BINARY ' ABC'", NewSQLBool(knd, false)},
			{"sql_like_binary_expr_15", "'abc' LIKE BINARY ' ABC '", NewSQLBool(knd, false)},
			{"sql_like_binary_expr_16", "'abc' LIKE BINARY 'ABC	'", NewSQLBool(knd, false)},
			{"sql_like_binary_expr_17", "'10' LIKE BINARY '1%'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_18", "'a   ' LIKE BINARY 'A   '", NewSQLBool(knd, false)},
			{"sql_like_binary_expr_19", "CURRENT_DATE() LIKE BINARY '2015-05-31%'",
				NewSQLBool(knd, false)},
			{"sql_like_binary_expr_20", "CURDATE() LIKE BINARY '2015-05-31%'",
				NewSQLBool(knd, false)},
			{"sql_like_binary_expr_21", "(DATE '2008-01-02') LIKE BINARY '2008-01%'",
				NewSQLBool(knd, true)},
			{"sql_like_binary_expr_22", "NOW() LIKE BINARY '" + strconv.Itoa(time.Now().Year()) +
				"%' ", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_23", "10 LIKE BINARY '1%'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_24", "1.20 LIKE BINARY '1.2%'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_25", "NULL LIKE BINARY '1%'", NewSQLNull(knd)},
			{"sql_like_binary_expr_26", "10 LIKE BINARY NULL", NewSQLNull(knd)},
			{"sql_like_binary_expr_27", "NULL LIKE BINARY NULL", NewSQLNull(knd)},
			{"sql_like_binary_expr_28", "'David_' LIKE BINARY 'David\\_'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_29", "'David%' LIKE BINARY 'David\\%'", NewSQLBool(knd, true)},
			{"sql_like_binary_expr_30", "'David_' LIKE BINARY 'David|_' ESCAPE '|'",
				NewSQLBool(knd, true)},
			{"sql_like_binary_expr_31", "'David\\_' LIKE BINARY 'David\\_' ESCAPE ''",
				NewSQLBool(knd, true)},
			{"sql_like_binary_expr_32", "'David_' LIKE BINARY 'David\\_' ESCAPE char(92)",
				NewSQLBool(knd, true)},
			{"sql_like_binary_expr_33", "'David_' LIKE BINARY 'David|_' {escape '|'}",
				NewSQLBool(knd, true)},
			{"sql_mixed_arithmetic_and_bool_0", "(5<6) + 1", NewSQLInt64(knd, 2)},
			{"sql_mixed_arithmetic_and_bool_1", "(5<6) && (6>4)", NewSQLBool(knd, true)},
			{"sql_mixed_arithmetic_and_bool_2", "(5<6) || (6>4)", NewSQLBool(knd, true)},
			{"sql_mixed_arithmetic_and_bool_3", "(5<6) XOR (6>4)", NewSQLBool(knd, false)},
			{"sql_mixed_arithmetic_and_bool_4", "(5<6)<7", NewSQLBool(knd, true)},
			{"sql_mixed_arithmetic_and_bool_5", "1+(5<6)", NewSQLInt64(knd, 2)},
			{"sql_mixed_arithmetic_and_bool_6", "1+(5>6)", NewSQLInt64(knd, 1)},
			{"sql_mixed_arithmetic_and_bool_7", "1+(NULL>6)", NewSQLNull(knd)},
			{"sql_mixed_arithmetic_and_bool_8", "NULL+(5>6)", NewSQLNull(knd)},
			{"sql_mixed_arithmetic_and_bool_9", "20/(5<6)", NewSQLFloat(knd, 20)},
			{"sql_mixed_arithmetic_and_bool_10", "20*(5<6)", NewSQLInt64(knd, 20)},
			{"sql_mixed_arithmetic_and_bool_11", "20/5<6", NewSQLBool(knd, true)},
			{"sql_mixed_arithmetic_and_bool_12", "20*5<6", NewSQLBool(knd, false)},
			{"sql_mixed_arithmetic_and_bool_13", "20+5<6", NewSQLBool(knd, false)},
			{"sql_mixed_arithmetic_and_bool_14", "20-5<6", NewSQLBool(knd, false)},
			{"sql_mixed_arithmetic_and_bool_15", "20+true", NewSQLInt64(knd, 21)},
			{"sql_mixed_arithmetic_and_bool_16", "20+false", NewSQLInt64(knd, 20)},
			{"sql_mod_expr_0", "0 % 0", NewSQLNull(knd)},
			{"sql_mod_expr_1", "5 % 2", NewSQLFloat(knd, 1)},
			{"sql_mod_expr_2", "5.5 % 2", NewSQLFloat(knd, 1.5)},
			{"sql_mod_expr_3", "-5 % -3", NewSQLFloat(knd, -2)},
			{"sql_mod_expr_4", "5 MOD 2", NewSQLFloat(knd, 1)},
			{"sql_mod_expr_5", "5.5 MOD 2", NewSQLFloat(knd, 1.5)},
			{"sql_mod_expr_6", "-5 MOD -3", NewSQLFloat(knd, -2)},
			{"sql_mult_expr_0", "0 * 0", NewSQLInt64(knd, 0)},
			{"sql_mult_expr_1", "-1 * 1", NewSQLInt64(knd, -1)},
			{"sql_mult_expr_2", "10 * 32", NewSQLInt64(knd, 320)},
			{"sql_mult_expr_3", "-10 * -32", NewSQLInt64(knd, 320)},
			{"sql_mult_expr_4", "2.5 * 3", NewSQLDecimal128(knd, decimal.New(75, -1))},
			{"sql_not_equal_expr_0", "0 <> 0", NewSQLBool(knd, false)},
			{"sql_not_equal_expr_1", "-1 <> 1", NewSQLBool(knd, true)},
			{"sql_not_equal_expr_2", "10 <> 10", NewSQLBool(knd, false)},
			{"sql_not_equal_expr_3", "-10 <> -10", NewSQLBool(knd, false)},
			{"sql_not_expr_0", "NOT 1", NewSQLBool(knd, false)},
			{"sql_not_expr_1", "NOT 0", NewSQLBool(knd, true)},
			{"sql_not_expr_2", "NOT true", NewSQLBool(knd, false)},
			{"sql_not_expr_3", "NOT false", NewSQLBool(knd, true)},
			{"sql_not_expr_4", "NOT NULL", NewSQLNull(knd)},
			{"sql_not_expr_5", "! 1", NewSQLBool(knd, false)},
			{"sql_not_expr_6", "! 0", NewSQLBool(knd, true)},
			{"sql_null_safe_equal_0", "0 <=> 0", NewSQLBool(knd, true)},
			{"sql_null_safe_equal_1", "-1 <=> 1", NewSQLBool(knd, false)},
			{"sql_null_safe_equal_2", "10 <=> 10", NewSQLBool(knd, true)},
			{"sql_null_safe_equal_3", "-10 <=> -10", NewSQLBool(knd, true)},
			{"sql_null_safe_equal_4", "1 <=> 1", NewSQLBool(knd, true)},
			{"sql_null_safe_equal_5", "NULL <=> NULL", NewSQLBool(knd, true)},
			{"sql_null_safe_equal_6", "1 <=> NULL", NewSQLBool(knd, false)},
			{"sql_null_safe_equal_7", "NULL <=> 1", NewSQLBool(knd, false)},
			{"sql_or_expr_0", "1 OR 1", NewSQLBool(knd, true)},
			{"sql_or_expr_1", "1 OR 0", NewSQLBool(knd, true)},
			{"sql_or_expr_2", "0 OR 1", NewSQLBool(knd, true)},
			{"sql_or_expr_3", "NULL OR 1", NewSQLBool(knd, true)},
			{"sql_or_expr_4", "NULL OR 0", NewSQLNull(knd)},
			{"sql_or_expr_5", "NULL OR NULL", NewSQLNull(knd)},
			{"sql_or_expr_6", "0 OR 0", NewSQLBool(knd, false)},
			{"sql_or_expr_7", "true OR true", NewSQLBool(knd, true)},
			{"sql_or_expr_8", "true OR false", NewSQLBool(knd, true)},
			{"sql_or_expr_9", "false OR true", NewSQLBool(knd, true)},
			{"sql_or_expr_10", "false OR false", NewSQLBool(knd, false)},
			{"sql_or_expr_11", "1 || 1", NewSQLBool(knd, true)},
			{"sql_or_expr_12", "1 || 0", NewSQLBool(knd, true)},
			{"sql_or_expr_13", "0 || 1", NewSQLBool(knd, true)},
			{"sql_or_expr_14", "0 || 0", NewSQLBool(knd, false)},
			{"sql_or_expr_15", "'foo' || 'bar'", NewSQLBool(knd, false)},
			{"sql_or_expr_16", "'foo' || true", NewSQLBool(knd, true)},
			{"sql_or_expr_17", "true  || 'bar'", NewSQLBool(knd, true)},
			{"sql_or_expr_18", "false || 'bar'", NewSQLBool(knd, false)},
			{"sql_or_expr_19", "1 || 'bar'", NewSQLBool(knd, true)},
			{"sql_or_expr_20", "0 || 'bar'", NewSQLBool(knd, false)},
			{"sql_or_expr_21", "'foo' || 1", NewSQLBool(knd, true)},
			{"sql_or_expr_22", "'foo' || 0", NewSQLBool(knd, false)},
			{"sql_xor_expr_0", "1 XOR 1", NewSQLBool(knd, false)},
			{"sql_xor_expr_1", "1 XOR 0", NewSQLBool(knd, true)},
			{"sql_xor_expr_2", "0 XOR 1", NewSQLBool(knd, true)},
			{"sql_xor_expr_3", "0 XOR 0", NewSQLBool(knd, false)},
			{"sql_not_regex_expr_0", "'ABC123' NOT REGEXP 'AB'", NewSQLBool(knd, false)},
			{"sql_not_regex_expr_1", "'ABC123' NOT REGEXP 'ABD'", NewSQLBool(knd, true)},
			{"sql_not_regex_expr_2", "'ABC123' NOT REGEXP '[[:alpha:]]'", NewSQLBool(knd, false)},
			{"sql_not_regex_expr_3", "'fofo' NOT REGEXP '^fo'", NewSQLBool(knd, false)},
			{"sql_not_regex_expr_4", "'fofo' NOT REGEXP '^f.*$'", NewSQLBool(knd, false)},
			{"sql_not_regex_expr_5", "'pi' NOT REGEXP 'pi|apa'", NewSQLBool(knd, false)},
			{"sql_not_regex_expr_6", "'abcde' NOT REGEXP 'a[bcd]{2}e'", NewSQLBool(knd, true)},
			{"sql_not_regex_expr_7", "'abcde' NOT REGEXP 'a[bcd]{1,10}e'", NewSQLBool(knd, false)},
			{"sql_not_regex_expr_8", "null REGEXP 'abc'", NewSQLNull(knd)},
			{"sql_not_regex_expr_9", "'a' REGEXP null", NewSQLNull(knd)},
			{"sql_not_regex_expr_10", "2-1 NOT REGEXP '1'", NewSQLBool(knd, false)},
			{"sql_regex_expr_0", "'ABC123' REGEXP 'AB'", NewSQLBool(knd, true)},
			{"sql_regex_expr_1", "'ABC123' REGEXP 'ABD'", NewSQLBool(knd, false)},
			{"sql_regex_expr_2", "'ABC123' REGEXP '[[:alpha:]]'", NewSQLBool(knd, true)},
			{"sql_regex_expr_3", "'fofo' REGEXP '^fo'", NewSQLBool(knd, true)},
			{"sql_regex_expr_4", "'fofo' REGEXP '^f.*$'", NewSQLBool(knd, true)},
			{"sql_regex_expr_5", "'pi' REGEXP 'pi|apa'", NewSQLBool(knd, true)},
			{"sql_regex_expr_6", "'abcde' REGEXP 'a[bcd]{2}e'", NewSQLBool(knd, false)},
			{"sql_regex_expr_7", "'abcde' REGEXP 'a[bcd]{1,10}e'", NewSQLBool(knd, true)},
			{"sql_regex_expr_8", "null REGEXP 'abc'", NewSQLNull(knd)},
			{"sql_regex_expr_9", "'a' REGEXP null", NewSQLNull(knd)},
			{"sql_regex_expr_10", "2-1 REGEXP '1'", NewSQLBool(knd, true)},
			{"sql_scalar_abs_expr_0", "ABS(NULL)", NewSQLNull(knd)},
			{"sql_scalar_abs_expr_1", "ABS('C')", NewSQLFloat(knd, 0)},
			{"sql_scalar_abs_expr_2", "ABS(-20)", NewSQLFloat(knd, 20)},
			{"sql_scalar_abs_expr_3", "ABS(20)", NewSQLFloat(knd, 20)},
			{"sql_scalar_acos_expr_0", "ACOS(NULL)", NewSQLNull(knd)},
			{"sql_scalar_acos_expr_1", "ACOS(20)", NewSQLNull(knd)},
			{"sql_scalar_acos_expr_2", "ACOS(-20)", NewSQLNull(knd)},
			{"sql_scalar_acos_expr_3", "ACOS('C')", NewSQLFloat(knd, 1.5707963267948966)},
			{"sql_scalar_acos_expr_4", "ACOS(0)", NewSQLFloat(knd, 1.5707963267948966)},
			{"sql_scalar_asin_expr_0", "ASIN(NULL)", NewSQLNull(knd)},
			{"sql_scalar_asin_expr_1", "ASIN(20)", NewSQLNull(knd)},
			{"sql_scalar_asin_expr_2", "ASIN(-20)", NewSQLNull(knd)},
			{"sql_scalar_asin_expr_3", "ASIN('C')", NewSQLFloat(knd, 0)},
			{"sql_scalar_asin_expr_4", "ASIN(0)", NewSQLFloat(knd, 0)},
			{"sql_scalar_atan_expr_0", "ATAN(NULL)", NewSQLNull(knd)},
			{"sql_scalar_atan_expr_1", "ATAN(20)", NewSQLFloat(knd, 1.5208379310729538)},
			{"sql_scalar_atan_expr_2", "ATAN(-20)", NewSQLFloat(knd, -1.5208379310729538)},
			{"sql_scalar_atan_expr_3", "ATAN('C')", NewSQLFloat(knd, 0)},
			{"sql_scalar_atan_expr_4", "ATAN(0)", NewSQLFloat(knd, 0)},
			{"sql_scalar_atan_expr_5", "ATAN(NULL, NULL)", NewSQLNull(knd)},
			{"sql_scalar_atan_expr_6", "ATAN('C', 2)", NewSQLFloat(knd, 0)},
			{"sql_scalar_atan_expr_7", "ATAN(0, 2)", NewSQLFloat(knd, 0)},
			{"sql_scalar_atan2_expr_0", "ATAN2(NULL, NULL)", NewSQLNull(knd)},
			{"sql_scalar_atan2_expr_1", "ATAN2('C', 2)", NewSQLFloat(knd, 0)},
			{"sql_scalar_atan2_expr_2", "ATAN2(0, 2)", NewSQLFloat(knd, 0)},
			{"sql_ascii_0", "ASCII(NULL)", NewSQLNull(knd)},
			{"sql_ascii_1", "ASCII('')", NewSQLInt64(knd, 0)},
			{"sql_ascii_2", "ASCII('A')", NewSQLInt64(knd, 65)},
			{"sql_ascii_3", "ASCII('AWESOME')", NewSQLInt64(knd, 65)},
			{"sql_ascii_4", "ASCII('¢')", NewSQLInt64(knd, 194)},
			{
				"sql_ascii_5",
				"ASCII('Č')",
				NewSQLInt64(knd, 196), // This is actually 268, but the first byte is 196
			},
			{"sql_ceil_0", "CEIL(NULL)", NewSQLNull(knd)},
			{"sql_ceil_1", "CEIL(20)", NewSQLInt64(knd, 20)},
			{"sql_ceil_2", "CEIL(-20)", NewSQLInt64(knd, -20)},
			{"sql_ceil_3", "CEIL('C')", NewSQLInt64(knd, 0)},
			{"sql_ceil_4", "CEIL(0.56)", NewSQLInt64(knd, 1)},
			{"sql_ceil_5", "CEIL(-0.56)", NewSQLInt64(knd, 0)},
			{"sql_ceiling_expr_0", "CEIL(NULL)", NewSQLNull(knd)},
			{"sql_ceiling_expr_1", "CEIL(20)", NewSQLInt64(knd, 20)},
			{"sql_ceiling_expr_2", "CEIL(-20)", NewSQLInt64(knd, -20)},
			{"sql_ceiling_expr_3", "CEIL('C')", NewSQLInt64(knd, 0)},
			{"sql_ceiling_expr_4", "CEIL(0.56)", NewSQLInt64(knd, 1)},
			{"sql_ceiling_expr_5", "CEIL(-0.56)", NewSQLInt64(knd, 0)},
			{"sql_char_expr_0", "CHAR(NULL)", NewSQLVarchar(knd, "")},
			{"sql_char_expr_1", "CHAR(77,121,83,81,'76')", NewSQLVarchar(knd, "MySQL")},
			{
				"sql_char_expr_2",
				"CHAR(77,121,NULL, 83, NULL, 81,'76')",
				NewSQLVarchar(knd, "MySQL"),
			},
			{"sql_char_expr_3", "CHAR(256)", NewSQLVarchar(knd, string([]byte{1, 0}))},
			{"sql_char_expr_4", "CHAR(512)", NewSQLVarchar(knd, string([]byte{2, 0}))},
			{"sql_char_expr_5", "CHAR(513)", NewSQLVarchar(knd, string([]byte{2, 1}))},
			{"sql_char_expr_6", "CHAR(256, 512)", NewSQLVarchar(knd, string([]byte{1, 0, 2, 0}))},
			{"sql_char_expr_7", "CHAR(65537)", NewSQLVarchar(knd, string([]byte{1, 0, 1}))},
			{"sql_char_length_0", "CHAR_LENGTH(NULL)", NewSQLNull(knd)},
			{"sql_char_length_1", "CHAR_LENGTH('sDg')", NewSQLInt64(knd, 3)},
			{"sql_char_length_2", "CHAR_LENGTH('世界')", NewSQLInt64(knd, 2)},
			{"sql_char_length_3", "CHAR_LENGTH('')", NewSQLInt64(knd, 0)},
			{"sql_char_length_4", "CHARACTER_LENGTH(NULL)", NewSQLNull(knd)},
			{"sql_char_length_5", "CHARACTER_LENGTH('sDg')", NewSQLInt64(knd, 3)},
			{"sql_char_length_6", "CHARACTER_LENGTH('世界')", NewSQLInt64(knd, 2)},
			{"sql_char_length_7", "CHARACTER_LENGTH('')", NewSQLInt64(knd, 0)},
			{"sql_coalesce_expr_0", "COALESCE(NULL)", NewSQLNull(knd)},
			{"sql_coalesce_expr_1", "COALESCE('A')", NewSQLVarchar(knd, "A")},
			{"sql_coalesce_expr_2", "COALESCE('A', NULL)", NewSQLVarchar(knd, "A")},
			{"sql_coalesce_expr_3", "COALESCE('A', 'B')", NewSQLVarchar(knd, "A")},
			{"sql_coalesce_expr_4", "COALESCE(NULL, 'A', NULL, 'B')", NewSQLVarchar(knd, "A")},
			{"sql_coalesce_expr_5", "COALESCE(NULL, NULL, NULL)", NewSQLNull(knd)},
			{"sql_concat_expr_0", "CONCAT(NULL)", NewSQLNull(knd)},
			{"sql_concat_expr_1", "CONCAT('A')", NewSQLVarchar(knd, "A")},
			{"sql_concat_expr_2", "CONCAT('A', 'B')", NewSQLVarchar(knd, "AB")},
			{"sql_concat_expr_3", "CONCAT('A', NULL, 'B')", NewSQLNull(knd)},
			{"sql_concat_expr_4", "CONCAT('A', 123, 'B')", NewSQLVarchar(knd, "A123B")},
			{"sql_concat_ws_expr_0", "CONCAT_WS(NULL, NULL)", NewSQLNull(knd)},
			{"sql_concat_ws_expr_1", "CONCAT_WS(',','A')", NewSQLVarchar(knd, "A")},
			{"sql_concat_ws_expr_2", "CONCAT_WS(',','A', 'B')", NewSQLVarchar(knd, "A,B")},
			{"sql_concat_ws_expr_3", "CONCAT_WS(',','A', NULL, 'B')", NewSQLVarchar(knd, "A,B")},
			{
				"sql_concat_ws_expr_4",
				"CONCAT_WS(',','A', 123, 'B')",
				NewSQLVarchar(knd, "A,123,B"),
			},
			{"sql_connection_id_expr", "CONNECTION_ID()", NewSQLUint64(knd, 42)},
			{"sql_cos_expr_0", "COS(NULL)", NewSQLNull(knd)},
			{"sql_cos_expr_1", "COS(20)", NewSQLFloat(knd, 0.40808206181339196)},
			{"sql_cos_expr_2", "COS(-20)", NewSQLFloat(knd, 0.40808206181339196)},
			{"sql_cos_expr_3", "COS('C')", NewSQLFloat(knd, 1)},
			{"sql_cos_expr_4", "COS(0)", NewSQLFloat(knd, 1)},
			{"sql_cot_expr_0", "COT(NULL)", NewSQLNull(knd)},
			{"sql_cot_expr_1", "COT(19)", NewSQLFloat(knd, 6.596764247280111)},
			{"sql_cot_expr_2", "COT(-19)", NewSQLFloat(knd, -6.596764247280111)},
			// current time tests do not work
			/*
				{
					"sql_current_date_expr",
					"CURRENT_DATE()",
					NewSQLDate(knd, time.Now().UTC()),
				},
				{
					"sql_current_ts_0",
					"CURRENT_TIMESTAMP()",
					NewSQLTimestamp(knd, time.Now().UTC()),
				},
				{
					"sql_current_ts_1",
					"CURRENT_TIMESTAMP",
					NewSQLTimestamp(knd, time.Now().UTC()),
				},
				{
					"sql_curtime_0",
					"CURRENT_TIMESTAMP()",
					NewSQLTimestamp(knd, time.Now().UTC()),
				},
				{
					"sql_curtime_1",
					"CURRENT_TIMESTAMP",
					NewSQLTimestamp(knd, time.Now().UTC()),
				},
				{
					"sql_utc_ts_0",
					"UTC_TIMESTAMP()",
					NewSQLTimestamp(knd, time.Now().UTC()),
				},
				{
					"sql_utc_ts_1",
					"UTC_TIMESTAMP",
					NewSQLTimestamp(knd, time.Now().UTC()),
				},
				{
					"sql_now_0",
					"NOW()",
					NewSQLTimestamp(knd, time.Now().UTC()),
				},
			*/
			{"sql_user_expr_0", "CURRENT_USER()", NewSQLVarchar(knd, "evaluator_unit_test_user@evaluator_unit_test_remoteHost")},
			{"sql_user_expr_1", "SESSION_USER()", NewSQLVarchar(knd, "evaluator_unit_test_user@evaluator_unit_test_remoteHost")},
			{"sql_user_expr_2", "SYSTEM_USER()", NewSQLVarchar(knd, "evaluator_unit_test_user@evaluator_unit_test_remoteHost")},
			{"sql_user_expr_3", "USER()", NewSQLVarchar(knd, "evaluator_unit_test_user@evaluator_unit_test_remoteHost")},
			{"sql_db_0", "DATABASE()", NewSQLVarchar(knd, "evaluator_unit_test_dbname")},
			{"sql_schema_0", "SCHEMA()", NewSQLVarchar(knd, "evaluator_unit_test_dbname")},
			{
				"sql_date_diff_0",
				"DATEDIFF('2017-01-01', '2016-01-01 23:08:56')",
				NewSQLInt64(knd, 366),
			},
			{"sql_date_diff_1", "DATEDIFF('2017-01-01', '2017-01-01')", NewSQLInt64(knd, 0)},
			{
				"sql_date_diff_2",
				"DATEDIFF('2017-08-23 10:40:43', '2017-09-30 12:19:50')",
				NewSQLInt64(knd, -38),
			},
			{"sql_date_diff_3", "DATEDIFF(NULL, '2017-09-30 12:19:50')", NewSQLNull(knd)},
			{"sql_date_diff_4", "DATEDIFF('2002-09-07', '1700-08-02')", NewSQLInt64(knd, 106751)},
			{"sql_date_diff_5", "DATEDIFF('1657-08-02', '2002-09-07')",
				NewSQLInt64(knd, -106751)},
			{
				"sql_date_diff_6",
				"DATEDIFF(20170823104043, '2017-09-30 12:19:50')",
				NewSQLInt64(knd, -38),
			},
			{
				"sql_date_diff_7",
				"DATEDIFF(20170823.09809, '2017-09-30 12:19:50')",
				NewSQLInt64(knd, -38),
			},
			{
				"sql_date_diff_8",
				"DATEDIFF('biconnectorisfun', '2017-09-30 12:19:50')",
				NewSQLNull(knd),
			},
			{"sql_date_diff_9", "DATEDIFF('2000-9-1', '2012-6-7')",
				NewSQLInt64(knd, -4297)},
			{"sql_date_diff_10", "DATEDIFF('00-09-1', '12-06-07')",
				NewSQLInt64(knd, -4297)},
			{"sql_date_format_0", "DATE_FORMAT('2009-10-04', NULL)", NewSQLNull(knd)},
			{"sql_date_format_1", "DATE_FORMAT(NULL, '2009-10-04')", NewSQLNull(knd)},
			{
				"sql_date_format_2",
				"DATE_FORMAT('2009-10-04 22:23:00', '%W %M 01 %Y')",
				NewSQLVarchar(knd, "Sunday October 01 2009"),
			},
			{
				"sql_date_format_3",
				"DATE_FORMAT('2009-10-04 22:23:00', '%W %M %Y')",
				NewSQLVarchar(knd, "Sunday October 2009"),
			},
			{
				"sql_date_format_4",
				"DATE_FORMAT('2007-10-04 22:23:00', '%H:01:%i:%s')",
				NewSQLVarchar(knd, "22:01:23:00"),
			},
			{
				"sql_date_format_5",
				"DATE_FORMAT('2007-10-04 22:23:00', '%H:%g:01%%:%i:%s%')",
				NewSQLVarchar(knd, "22:%g:01%:23:00%"),
			},
			{
				"sql_date_format_6",
				"DATE_FORMAT('2007-10-04 22:23:00', '%H:%i:%s')",
				NewSQLVarchar(knd, "22:23:00"),
			},
			{
				"sql_date_format_7",
				"DATE_FORMAT('1900-10-04 22:23:00', '%D %y %a %d %m %b %j')",
				NewSQLVarchar(knd, "4th 00 Thu 04 10 Oct 277"),
			},
			{
				"sql_date_format_8",
				"DATE_FORMAT('1997-10-04 22:23:00', '%H %k %I %r %T %S %w')",
				NewSQLVarchar(knd, "22 22 10 10:23:00 PM 22:23:00 00 6"),
			},
			{
				"sql_date_format_9",
				"DATE_FORMAT('1999-01-01', '%X %V')",
				NewSQLVarchar(knd, "1998 52"),
			},
			{
				"sql_date_format_10",
				"DATE_FORMAT('1989-05-14 01:03:01.232335','%a|%b|%c|%D|%d|%e|%f|%H|%h|%I|%i|%j|%k" +
					"|%l|%M|%m|%p|%r|%S|%s|%T|%U|%u|%V|%v|%W|%w|%X|%x|%Y|%y|%%|%x')",
				NewSQLVarchar(knd, "Sun|May|5|14th|14|14|232335|01|01|01|03|134|1|1|May|05|AM"+
					"|01:03:01 AM|01|01|01:03:01|20|19|20|19|Sunday|0|1989|1989|1989|89|%|1989"),
			},
			{
				"sql_date_format_11",
				"DATE_FORMAT('1900-10-04 22:23:00', '%a|%b|%c|%D|%d|%e|%f|%H|%h|%I|%i|%j|%k|%l|%M" +
					"|%m|%p|%r|%S|%s|%T|%U|%u|%V|%v|%W|%w|%X|%x|%Y|%y|%%|%x')",
				NewSQLVarchar(knd, "Thu|Oct|10|4th|04|4|000000|22|10|10|23|277|22|10|October|10"+
					"|PM|10:23:00 PM|00|00|22:23:00|39|40|39|40|Thursday|4|1900"+
					"|1900|1900|00|%|1900"),
			},
			{
				"sql_date_format_12",
				"DATE_FORMAT('1983-07-05 23:22', '%a|%b|%c|%D|%d|%e|%f|%H|%h|%I|%i|%j|%k|%l|%M|%m" +
					"|%p|%r|%S|%s|%T|%U|%u|%V|%v|%W|%w|%X|%x|%Y|%y|%%|%x')",
				NewSQLVarchar(knd, "Tue|Jul|7|5th|05|5|000000|23|11|11|22|186|23|11|July|07|PM"+
					"|11:22:00 PM|00|00|23:22:00|27|27|27|27|Tuesday|2|1983|1983|1983|83|%|1983"),
			},
			{"sql_date_name_0", "DAYNAME(NULL)", NewSQLNull(knd)},
			{"sql_date_name_1", "DAYNAME(14)", NewSQLNull(knd)},
			{"sql_date_name_2", "DAYNAME('2016-01-01 00:00:00')", NewSQLVarchar(knd, "Friday")},
			{"sql_date_name_3", "DAYNAME('2016-1-1')", NewSQLVarchar(knd, "Friday")},
			{"sql_date_of_month_0", "DAYOFMONTH(NULL)", NewSQLNull(knd)},
			{"sql_date_of_month_1", "DAYOFMONTH(14)", NewSQLNull(knd)},
			{"sql_date_of_month_2", "DAYOFMONTH('2016-01-01')", NewSQLInt64(knd, 1)},
			{"sql_date_of_month_3", "DAYOFMONTH('2016-1-1')", NewSQLInt64(knd, 1)},
			{"sql_date_of_week_0", "DAYOFWEEK(NULL)", NewSQLNull(knd)},
			{"sql_date_of_week_1", "DAYOFWEEK(14)", NewSQLNull(knd)},
			{"sql_date_of_week_2", "DAYOFWEEK('2016-01-01')", NewSQLInt64(knd, 6)},
			{"sql_date_of_week_3", "DAYOFWEEK('2016-1-1')", NewSQLInt64(knd, 6)},
			{"sql_date_of_year_0", "DAYOFYEAR(NULL)", NewSQLNull(knd)},
			{"sql_date_of_year_1", "DAYOFYEAR(14)", NewSQLNull(knd)},
			{"sql_date_of_year_2", "DAYOFYEAR('2016-1-1')", NewSQLInt64(knd, 1)},
			{"sql_date_of_year_3", "DAYOFYEAR('2016-01-01')", NewSQLInt64(knd, 1)},
			{"sql_degrees_0", "DEGREES(NULL)", NewSQLNull(knd)},
			{"sql_degrees_1", "DEGREES(20)", NewSQLFloat(knd, 1145.9155902616465)},
			{"sql_degrees_2", "DEGREES(-20)", NewSQLFloat(knd, -1145.9155902616465)},
			{"sql_elt_expr_0", "ELT(NULL, 'a', 'b')", NewSQLNull(knd)},
			{"sql_elt_expr_1", "ELT(0, 'a', 'b')", NewSQLNull(knd)},
			{"sql_elt_expr_2", "ELT(1, 'a', 'b')", NewSQLVarchar(knd, "a")},
			{"sql_elt_expr_3", "ELT(2, 'a', 'b')", NewSQLVarchar(knd, "b")},
			{"sql_elt_expr_4", "ELT(3, 'a', 'b', NULL)", NewSQLNull(knd)},
			{"sql_elt_expr_5", "ELT(4, 'a', 'b', NULL)", NewSQLNull(knd)},
			{"sql_exp_expr_0", "EXP(NULL)", NewSQLNull(knd)},
			{"sql_exp_expr_1", "EXP('sdg')", NewSQLFloat(knd, 1)},
			{"sql_exp_expr_2", "EXP(0)", NewSQLFloat(knd, 1)},
			{"sql_exp_expr_3", "EXP(2)", NewSQLFloat(knd, 7.38905609893065)},
			{"sql_extract_expr_0", "EXTRACT(YEAR FROM NULL)", NewSQLNull(knd)},
			{
				"sql_extract_expr_1",
				"EXTRACT(YEAR FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 2006),
			},
			{
				"sql_extract_expr_2",
				"EXTRACT(QUARTER FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 2),
			},
			{
				"sql_extract_expr_3",
				"EXTRACT(WEEK FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 14),
			},
			{
				"sql_extract_expr_4",
				"EXTRACT(DAY FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 7),
			},
			{
				"sql_extract_expr_5",
				"EXTRACT(HOUR FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 7),
			},
			{
				"sql_extract_expr_6",
				"EXTRACT(MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 14),
			},
			{
				"sql_extract_expr_7",
				"EXTRACT(SECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 23),
			},
			{
				"sql_extract_expr_8",
				"EXTRACT(MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 0),
			},
			{
				"sql_extract_expr_9",
				"EXTRACT(YEAR_MONTH FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 200604),
			},
			{
				"sql_extract_expr_10",
				"EXTRACT(DAY_HOUR FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 707),
			},
			{
				"sql_extract_expr_11",
				"EXTRACT(DAY_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 70714),
			},
			{
				"sql_extract_expr_12",
				"EXTRACT(DAY_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 7071423),
			},
			{
				"sql_extract_expr_13",
				"EXTRACT(DAY_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 7071423000000),
			},
			{
				"sql_extract_expr_14",
				"EXTRACT(HOUR_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 714),
			},
			{
				"sql_extract_expr_15",
				"EXTRACT(HOUR_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 71423),
			},
			{
				"sql_extract_expr_16",
				"EXTRACT(HOUR_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 71423000000),
			},
			{
				"sql_extract_expr_17",
				"EXTRACT(MINUTE_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 1423),
			},
			{
				"sql_extract_expr_18",
				"EXTRACT(MINUTE_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 1423000000),
			},
			{
				"sql_extract_expr_19",
				"EXTRACT(SECOND_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 23000000),
			},
			{
				"sql_extract_expr_20",
				"EXTRACT(SQL_TSI_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')",
				NewSQLInt64(knd, 14),
			},
			{"sql_floor_expr_0", "FLOOR(NULL)", NewSQLNull(knd)},
			{"sql_floor_expr_1", "FLOOR('sdg')", NewSQLInt64(knd, 0)},
			{"sql_floor_expr_2", "FLOOR(1.23)", NewSQLInt64(knd, 1)},
			{"sql_floor_expr_3", "FLOOR(-1.23)", NewSQLInt64(knd, -2)},
			{"sql_from_unixtime_0", "FROM_UNIXTIME(NULL)", NewSQLNull(knd)},
			{"sql_from_unixtime_1", "FROM_UNIXTIME(-1)", NewSQLNull(knd)},
			{
				"sql_from_unixtime_2",
				"FROM_UNIXTIME(1447430881) + 0",
				NewSQLDecimal128(knd, decimal.New(20151113160801, 0)),
			},
			{
				"sql_from_unixtime_3",
				"FROM_UNIXTIME(1447430881.5) + 0",
				NewSQLDecimal128(knd, decimal.New(20151113160802, 0)),
			},
			{
				"sql_from_unixtime_4",
				"CONCAT(FROM_UNIXTIME(1447430881), '')",
				NewSQLVarchar(knd, "2015-11-13 16:08:01.000000"),
			},
			{
				"sql_from_unixtime_5",
				"CONCAT(FROM_UNIXTIME(1447430881.5), '')",
				NewSQLVarchar(knd, "2015-11-13 16:08:02.000000"),
			},
			{"sql_hour_0", "HOUR(NULL)", NewSQLNull(knd)},
			{"sql_hour_1", "HOUR('sdg')", NewSQLInt64(knd, 0)},
			{"sql_hour_2", "HOUR('10:23:52')", NewSQLInt64(knd, 10)},
			{"sql_hour_3", "HOUR('10:61:52')", NewSQLNull(knd)},
			{"sql_hour_4", "HOUR('10:23:52.23.25.26')", NewSQLInt64(knd, 10)},
			{"sql_if_expr_0", "IF(1<2, 4, 5)", NewSQLInt64(knd, 4)},
			{"sql_if_expr_1", "IF(1>2, 4, 5)", NewSQLInt64(knd, 5)},
			{"sql_if_expr_2", "IF(1, 4, 5)", NewSQLInt64(knd, 4)},
			{"sql_if_expr_3", "IF(-0, 4, 5)", NewSQLInt64(knd, 5)},
			{"sql_if_expr_4", "IF(1-1, 4, 5)", NewSQLInt64(knd, 5)},
			{"sql_if_expr_5", "IF('cat', 4, 5)", NewSQLInt64(knd, 5)},
			{"sql_if_expr_6", "IF('3', 4, 5)", NewSQLInt64(knd, 4)},
			{"sql_if_expr_7", "IF('0', 4, 5)", NewSQLInt64(knd, 5)},
			{"sql_if_expr_8", "IF('-0.0', 4, 5)", NewSQLInt64(knd, 5)},
			{"sql_if_expr_9", "IF('2.2', 4, 5)", NewSQLInt64(knd, 4)},
			{"sql_if_expr_10", "IF('true', 4, 5)", NewSQLInt64(knd, 5)},
			{"sql_if_expr_11", "IF(null, 4, 'cat')", NewSQLVarchar(knd, "cat")},
			{"sql_if_expr_12", "IF(true, 'dog', 'cat')", NewSQLVarchar(knd, "dog")},
			{"sql_if_expr_13", "IF(false, 'dog', 'cat')", NewSQLVarchar(knd, "cat")},
			{"sql_if_expr_14", "IF('ca.gh', 4, 5)", NewSQLInt64(knd, 5)},
			{
				"sql_if_expr_15",
				"IF(current_timestamp(), 4, 5)",
				NewSQLInt64(knd, 4), // not being parsed as dates, being parsed as string
			},
			{"sql_if_expr_16", "IF(current_timestamp, 4, 5)", NewSQLInt64(knd, 4)},
			{"sql_if_null_0", "IFNULL(1,0)", NewSQLInt64(knd, 1)},
			{"sql_if_null_1", "IFNULL(NULL,3)", NewSQLInt64(knd, 3)},
			{"sql_if_null_2", "IFNULL(NULL,NULL)", NewSQLNull(knd)},
			{"sql_if_null_3", "IFNULL('cat', null)", NewSQLVarchar(knd, "cat")},
			{"sql_if_null_4", "IFNULL(null, 'dog')", NewSQLVarchar(knd, "dog")},
			{"sql_if_null_5", "IFNULL(1/0, 4)", NewSQLInt64(knd, 4)},
			{"sql_interval_expr_0", "INTERVAL(1,0)", NewSQLInt64(knd, 1)},
			{"sql_interval_expr_1", "INTERVAL(NULL, 3)", NewSQLInt64(knd, -1)},
			{"sql_interval_expr_2", "INTERVAL(NULL, NULL)", NewSQLInt64(knd, -1)},
			{"sql_interval_expr_3", "INTERVAL(2, 1, 2, 3, 4)", NewSQLInt64(knd, 2)},
			{"sql_interval_expr_4", "INTERVAL('1.1', 0, 1.1, 2)", NewSQLInt64(knd, 2)},
			{"sql_interval_expr_5", "INTERVAL(-1, NULL, 4)", NewSQLInt64(knd, 1)},
			{"sql_interval_expr_6", "INTERVAL(4, 1, 2, 4)", NewSQLInt64(knd, 3)},
			{"sql_is_null_0", "ISNULL(a)", NewSQLBool(knd, false)},
			{"sql_is_null_1", "ISNULL(c)", NewSQLBool(knd, true)},
			{"sql_is_null_2", `ISNULL("")`, NewSQLBool(knd, false)},
			{"sql_is_null_3", `ISNULL(NULL)`, NewSQLBool(knd, true)},
			{"sql_insert_expr_0", "INSERT('Quadratic', NULL, 4, 'What')", NewSQLNull(knd)},
			{
				"sql_insert_expr_1",
				"INSERT('Quadratic', 3, 4, 'What')",
				NewSQLVarchar(knd, "QuWhattic"),
			},
			{
				"sql_insert_expr_2",
				"INSERT('Quadratic', -1, 4, 'What')",
				NewSQLVarchar(knd, "Quadratic"),
			},
			{
				"sql_insert_expr_3",
				"INSERT('Quadratic', 3, 100, 'What')",
				NewSQLVarchar(knd, "QuWhat"),
			},
			{
				"sql_insert_expr_4",
				"INSERT('Quadratic', 9, 4, 'What')",
				NewSQLVarchar(knd, "QuadratiWhat"),
			},
			{
				"sql_insert_expr_5",
				"INSERT('Quadratic', 8.5, 3.5, 'What')",
				NewSQLVarchar(knd, "QuadratiWhat"),
			},
			{
				"sql_insert_expr_6",
				"INSERT('Quadratic', 8.4, 3.4, 'What')",
				NewSQLVarchar(knd, "QuadratWhat"),
			},
			{"sql_instr_expr_0", "INSTR(NULL, NULL)", NewSQLNull(knd)},
			{"sql_instr_expr_1", "INSTR('sDg', 'D')", NewSQLInt64(knd, 2)},
			{"sql_instr_expr_2", "INSTR(124, 124)", NewSQLInt64(knd, 1)},
			{"sql_instr_expr_3", "INSTR('awesome','so')", NewSQLInt64(knd, 4)},
			{"sql_lcase_0", "LCASE(NULL)", NewSQLNull(knd)},
			{"sql_lcase_1", "LCASE('sDg')", NewSQLVarchar(knd, "sdg")},
			{"sql_lcase_2", "LCASE(124)", NewSQLVarchar(knd, "124")},
			{"sql_lowercase_0", "LOWER(NULL)", NewSQLNull(knd)},
			{"sql_lowercase_1", "LOWER('')", NewSQLVarchar(knd, "")},
			{"sql_lowercase_2", "LOWER('A')", NewSQLVarchar(knd, "a")},
			{"sql_lowercase_3", "LOWER('awesome')", NewSQLVarchar(knd, "awesome")},
			{"sql_lowercase_4", "LOWER('AwEsOmE')", NewSQLVarchar(knd, "awesome")},
			// if any argument null, should return null
			{"sql_left_null_arg_0", "LEFT(NULL, NULL)", NewSQLNull(knd)},
			{"sql_left_null_arg_1", "LEFT('hi', NULL)", NewSQLNull(knd)},
			{"sql_left_null_arg_2", "LEFT(NULL, 5)", NewSQLNull(knd)},
			// basic cases w/ string, int inputs and positive int length
			{"sql_left_base_0", "LEFT('sDgcdcdc', 4)", NewSQLVarchar(knd, "sDgc")},
			{"sql_left_base_1", "LEFT(124, 2)", NewSQLVarchar(knd, "12")},
			// negative lengths and 0 give empty string
			{"sql_left_negative_0", "LEFT('hi', -1)", NewSQLVarchar(knd, "")},
			{"sql_left_negative_1", "LEFT('hi', 0)", NewSQLVarchar(knd, "")},
			{"sql_left_negative_2", "LEFT('hi', -2.5)", NewSQLVarchar(knd, "")},
			// float lengths should be rounded to closest int
			{"sql_left_float_0", "LEFT('hello', 2.4)", NewSQLVarchar(knd, "he")},
			{"sql_left_float_1", "LEFT('hello', 2.5)", NewSQLVarchar(knd, "hel")},
			{"sql_left_float_2", "LEFT(1234, 2.3)", NewSQLVarchar(knd, "12")},
			{"sql_left_float_3", "LEFT(1234, 2.5)", NewSQLVarchar(knd, "123")},
			{"sql_left_float_4", "LEFT('yo', 2.5)", NewSQLVarchar(knd, "yo")},
			// strings with spaces and symbols
			{"sql_left_symbols_0", "LEFT('  ', 1)", NewSQLVarchar(knd, " ")},
			{"sql_left_symbols_1", "LEFT('@!%', 2)", NewSQLVarchar(knd, "@!")},
			// boolean for string
			{"sql_left_bool_0", "LEFT(true, 3)", NewSQLVarchar(knd, "1")},
			{"sql_left_bool_1", "LEFT(false, 3)", NewSQLVarchar(knd, "0")},
			// boolean for length
			{"sql_left_bool_2", "LEFT('hello', true)", NewSQLVarchar(knd, "h")},
			{"sql_left_bool_3", "LEFT('hello', false)", NewSQLVarchar(knd, "")},
			// string for length
			{"sql_left_string", "LEFT('hello', 'hi')", NewSQLVarchar(knd, "")},
			// len > length of string
			{"sql_left_edge_case", "LEFT('hi', 5)", NewSQLVarchar(knd, "hi")},
			// string number as length
			{"sql_left_string_int_0", "LEFT('hello', '2')", NewSQLVarchar(knd, "he")},
			{"sql_left_string_int_1", "LEFT('hello', '-3')", NewSQLVarchar(knd, "")},
			// unlike with floats, string #s always round down
			{"sql_left_string_float_0", "LEFT('hello', '2.4')", NewSQLVarchar(knd, "he")},
			{"sql_left_string_float_1", "LEFT('hello', '2.6')", NewSQLVarchar(knd, "he")},
			{"sql_length_0", "LENGTH(NULL)", NewSQLNull(knd)},
			{"sql_length_1", "LENGTH('sDg')", NewSQLInt64(knd, 3)},
			{"sql_length_2", "LENGTH('世界')", NewSQLInt64(knd, 6)},
			{"sql_ln_expr_0", "LN(NULL)", NewSQLNull(knd)},
			{"sql_ln_expr_1", "LN(1)", NewSQLFloat(knd, 0)},
			{"sql_ln_expr_2", "LN(16.5)", NewSQLFloat(knd, 2.803360380906535)},
			{"sql_ln_expr_3", "LN(-16.5)", NewSQLNull(knd)},
			{"sql_log_expr_0", "LOG(NULL)", NewSQLNull(knd)},
			{"sql_log_expr_1", "LOG(1)", NewSQLFloat(knd, 0)},
			{"sql_log_expr_2", "LOG(16.5)", NewSQLFloat(knd, 2.803360380906535)},
			{"sql_log_expr_3", "LOG(-16.5)", NewSQLNull(knd)},
			{"sql_log_expr_4", "LOG10(100)", NewSQLFloat(knd, 2)},
			{"sql_log_expr_5", "LOG(10,100)", NewSQLFloat(knd, 2)},
			{"sql_locate_0", "LOCATE(NULL, 'foobarbar')", NewSQLNull(knd)},
			{"sql_locate_1", "LOCATE('bar', NULL)", NewSQLNull(knd)},
			{"sql_locate_2", "LOCATE('bar', 'foobarbar')", NewSQLInt64(knd, 4)},
			{"sql_locate_3", "LOCATE('xbar', 'foobarbar')", NewSQLInt64(knd, 0)},
			{"sql_locate_4", "LOCATE('bar', 'foobarbar', 5)", NewSQLInt64(knd, 7)},
			{"sql_locate_5", "LOCATE('bar', 'foobarbar', 4)", NewSQLInt64(knd, 4)},
			{"sql_locate_6", "LOCATE('e', 'dvd', 6)", NewSQLInt64(knd, 0)},
			{"sql_locate_7", "LOCATE('f', 'asdf', 4)", NewSQLInt64(knd, 4)},
			{"sql_locate_8", "LOCATE('語', '日本語')", NewSQLInt64(knd, 3)},
			{"sql_log2_0", "LOG2(NULL)", NewSQLNull(knd)},
			{"sql_log2_1", "LOG2(4)", NewSQLFloat(knd, 2)},
			{"sql_log2_2", "LOG2(-100)", NewSQLNull(knd)},
			{"sql_log10_0", "LOG10(NULL)", NewSQLNull(knd)},
			{"sql_log10_1", "LOG10('sdg')", NewSQLNull(knd)},
			{"sql_log10_2", "LOG10(2)", NewSQLFloat(knd, 0.3010299956639812)},
			{"sql_log10_3", "LOG10(100)", NewSQLFloat(knd, 2)},
			{"sql_log10_4", "LOG10(0)", NewSQLNull(knd)},
			{"sql_log10_5", "LOG10(-100)", NewSQLNull(knd)},
			{"sql_ltrim_0", "LTRIM(NULL)", NewSQLNull(knd)},
			{"sql_ltrim_1", "LTRIM('   barbar')", NewSQLVarchar(knd, "barbar")},
			{"sql_md5_0", "MD5(NULL)", NewSQLNull(knd)},
			{"sql_md5_1", "MD5(NULL + NULL)", NewSQLNull(knd)},
			{
				"sql_md5_2",
				"MD5('testing')",
				NewSQLVarchar(knd, "ae2b1fca515949e5d54fb22b8ed95575"),
			},
			{"sql_md5_3", "MD5('hello')", NewSQLVarchar(knd, "5d41402abc4b2a76b9719d911017c592")},
			{"sql_md5_4", "MD5(12)", NewSQLVarchar(knd, "c20ad4d76fe97759aa27a0c99bff6710")},
			{"sql_md5_5", "MD5(6.23)", NewSQLVarchar(knd, "fec8db978f6b7844b09d9bd54fb8335c")},
			{
				"sql_md5_6",
				"MD5('12:STR.002234')",
				NewSQLVarchar(knd, "81d56d5aeb92a55298af2f091e86ab61"),
			},
			{
				"sql_md5_7",
				"MD5(REPEAT('a', 30))",
				NewSQLVarchar(knd, "59e794d45697b360e18ba972bada0123"),
			},
			{"sql_microsecond_0", "MICROSECOND(NULL)", NewSQLNull(knd)},
			{"sql_microsecond_1", "MICROSECOND('')", NewSQLNull(knd)},
			{"sql_microsecond_2", "MICROSECOND('NULL')", NewSQLInt64(knd, 0)},
			{"sql_microsecond_3", "MICROSECOND('hello')", NewSQLInt64(knd, 0)},
			{"sql_microsecond_4", "MICROSECOND(TRUE)", NewSQLInt64(knd, 0)},
			{"sql_microsecond_5", "MICROSECOND('true')", NewSQLInt64(knd, 0)},
			{"sql_microsecond_6", "MICROSECOND('FALSE')", NewSQLInt64(knd, 0)},
			{"sql_microsecond_7", "MICROSECOND('11:38:24')", NewSQLInt64(knd, 0)},
			{"sql_microsecond_8", "MICROSECOND('11:38')", NewSQLInt64(knd, 0)},
			{"sql_microsecond_9", "MICROSECOND('11 38 24')", NewSQLInt64(knd, 0)},
			{"sql_microsecond_10", "MICROSECOND('11:38:24.000000')", NewSQLInt64(knd, 0)},
			{"sql_microsecond_11", "MICROSECOND('11:38:24.000001')", NewSQLInt64(knd, 1)},
			{"sql_microsecond_12", "MICROSECOND('11:38:24.123456')", NewSQLInt64(knd, 123456)},
			{"sql_microsecond_13", "MICROSECOND('1978-9-22 1:58:59')", NewSQLInt64(knd, 0)},
			{"sql_microsecond_14", "MICROSECOND('1978-9-22 1:58:59.00001')",
				NewSQLInt64(knd, 10)},
			{
				"sql_microsecond_15",
				"MICROSECOND('1978-9-22 1:58:59.0000104')",
				NewSQLInt64(knd, 10),
			},
			{"sql_microsecond_16", "MICROSECOND('12:STUFF.002234')", NewSQLInt64(knd, 0)},
			{"sql_mid_0", "MID('foobarbar', 4, NULL)", NewSQLNull(knd)},
			{"sql_mid_1", "MID('Quadratically', 5, 6)", NewSQLVarchar(knd, "ratica")},
			{"sql_mid_2", "MID('Quadratically', 12, 2)", NewSQLVarchar(knd, "ly")},
			{"sql_mid_3", "MID('Sakila', -5, 3)", NewSQLVarchar(knd, "aki")},
			{"sql_mid_4", "MID('日本語', 2, 1)", NewSQLVarchar(knd, "本")},
			{"sql_minute_0", "MINUTE(NULL)", NewSQLNull(knd)},
			{"sql_minute_1", "MINUTE('sdg')", NewSQLInt64(knd, 0)},
			{"sql_minute_2", "MINUTE('10:23:52')", NewSQLInt64(knd, 23)},
			{"sql_minute_3", "MINUTE('10:61:52')", NewSQLNull(knd)},
			{"sql_minute_4", "MINUTE('10:23:52.25.26.27.28')", NewSQLInt64(knd, 23)},
			{"sql_mod_0", "MOD(NULL, NULL)", NewSQLNull(knd)},
			{"sql_mod_1", "MOD(234, NULL)", NewSQLNull(knd)},
			{"sql_mod_2", "MOD(NULL, 10)", NewSQLNull(knd)},
			{"sql_mod_3", "MOD(234, 0)", NewSQLNull(knd)},
			{"sql_mod_4", "MOD(234, 10)", NewSQLFloat(knd, 4)},
			{"sql_mod_5", "MOD(253, 7)", NewSQLFloat(knd, 1)},
			{"sql_mod_6", "MOD(34.5, 3)", NewSQLFloat(knd, 1.5)},
			{"sql_month_0", "MONTH(NULL)", NewSQLNull(knd)},
			{"sql_month_1", "MONTH('sdg')", NewSQLNull(knd)},
			{"sql_month_2", "MONTH('2016-1-01 10:23:52')", NewSQLInt64(knd, 1)},
			{"sql_month_name_expr_0", "MONTHNAME(NULL)", NewSQLNull(knd)},
			{"sql_month_name_expr_1", "MONTHNAME('sdg')", NewSQLNull(knd)},
			{
				"sql_month_name_expr_2",
				"MONTHNAME('2016-1-01 10:23:52')",
				NewSQLVarchar(knd, "January"),
			},
			{"sql_null_if_0", "NULLIF(1,1)", NewSQLNull(knd)},
			{"sql_null_if_1", "NULLIF(1,3)", NewSQLInt64(knd, 1)},
			{"sql_null_if_2", "NULLIF(null, null)", NewSQLNull(knd)},
			{"sql_null_if_3", "NULLIF(null, 4)", NewSQLNull(knd)},
			{"sql_null_if_4", "NULLIF(3, null)", NewSQLInt64(knd, 3)},
			//test{"sql_null_if_5", "NULLIF(3, '3')", NewSQLNull(knd)},
			{"sql_null_if_6", "NULLIF('abc', 'abc')", NewSQLNull(knd)},
			//test{"sql_null_if_7", "NULLIF('abc', 3)", NewSQLVarchar(knd, "abc")},
			//test{"sql_null_if_8", "NULLIF('1', true)", NewSQLNull(knd)},
			//test{"sql_null_if_9", "NULLIF('1', false)", NewSQLVarchar(knd, "1")},
			{"sql_pi_expr", "PI()", NewSQLFloat(knd, 3.141592653589793116)},
			{"sql_quarter_0", "QUARTER(NULL)", NewSQLNull(knd)},
			{"sql_quarter_1", "QUARTER('sdg')", NewSQLNull(knd)},
			{"sql_quarter_2", "QUARTER('2016-1-01 10:23:52')", NewSQLInt64(knd, 1)},
			{"sql_quarter_3", "QUARTER('2016-4-01 10:23:52')", NewSQLInt64(knd, 2)},
			{"sql_quarter_4", "QUARTER('2016-8-01 10:23:52')", NewSQLInt64(knd, 3)},
			{"sql_quarter_5", "QUARTER('2016-11-01 10:23:52')", NewSQLInt64(knd, 4)},
			{"sql_radians_0", "RADIANS(NULL)", NewSQLNull(knd)},
			{"sql_radians_1", "RADIANS(1145.9155902616465)", NewSQLFloat(knd, 20)},
			{"sql_radians_2", "RADIANS(-1145.9155902616465)", NewSQLFloat(knd, -20)},
			{"sql_rand_0", "RAND(NULL)", NewSQLFloat(knd, 0.9451961492941164)},
			{"sql_rand_1", "RAND('hello')", NewSQLFloat(knd, 0.9451961492941164)},
			{"sql_rand_2", "RAND(0)", NewSQLFloat(knd, 0.9451961492941164)},
			{"sql_rand_3", "RAND(1145.9155902616465)", NewSQLFloat(knd, 0.16758646518059656)},
			{"sql_rand_4", "RAND(-1145.9155902616465)", NewSQLFloat(knd, 0.8321372077808122)},
			{"sql_repeat_0", "REPEAT(NULL, NULL)", NewSQLNull(knd)},
			{"sql_repeat_1", "REPEAT(NULL, 3)", NewSQLNull(knd)},
			{"sql_repeat_2", "REPEAT('apples', NULL)", NewSQLNull(knd)},
			{"sql_repeat_3", "REPEAT('apples', -1)", NewSQLVarchar(knd, "")},
			{"sql_repeat_4", "REPEAT('apples', 0)", NewSQLVarchar(knd, "")},
			{"sql_repeat_5", "REPEAT('apples', 1)", NewSQLVarchar(knd, "apples")},
			{"sql_repeat_6", "REPEAT('a', 5)", NewSQLVarchar(knd, "aaaaa")},
			{"sql_repeat_7", "REPEAT(3, 5)", NewSQLVarchar(knd, "33333")},
			{"sql_repeat_8", "REPEAT(FALSE, 5)", NewSQLVarchar(knd, "00000")},
			{"sql_repeat_9", "REPEAT(FALSE, TRUE)", NewSQLVarchar(knd, "0")},
			{"sql_repeat_10", "REPEAT('', 10)", NewSQLVarchar(knd, "")},
			{"sql_repeat_11", "REPEAT(0, '4')", NewSQLVarchar(knd, "0000")},
			{"sql_repeat_12", "REPEAT(NULL, 4)", NewSQLNull(knd)},
			{"sql_repeat_13", "REPEAT(1.4, 3)", NewSQLVarchar(knd, "1.41.41.4")},
			{"sql_repeat_14", "REPEAT('a', .3)", NewSQLVarchar(knd, "")},
			{"sql_repeat_15", "REPEAT('a', 3.2)", NewSQLVarchar(knd, "aaa")},
			{"sql_repeat_16", "REPEAT('a', 3.6)", NewSQLVarchar(knd, "aaaa")},
			{"sql_replace_0", "REPLACE(NULL, NULL, NULL)", NewSQLNull(knd)},
			{"sql_replace_1", "REPLACE('sDgcdcdc', 'D', 'd')", NewSQLVarchar(knd, "sdgcdcdc")},
			{
				"sql_replace_2",
				"REPLACE('www.mysql.com', 'w', 'Ww')",
				NewSQLVarchar(knd, "WwWwWw.mysql.com"),
			},
			{"sql_reverse_0", "REVERSE(NULL)", NewSQLNull(knd)},
			{"sql_reverse_1", "REVERSE(3.14159265)", NewSQLVarchar(knd, "56295141.3")},
			{"sql_reverse_2", "REVERSE(655)", NewSQLVarchar(knd, "556")},
			{"sql_reverse_3", "REVERSE('www.mysql.com')", NewSQLVarchar(knd, "moc.lqsym.www")},
			// if any argument null, should return null
			{"sql_right_null_0", "RIGHT(NULL, NULL)", NewSQLNull(knd)},
			{"sql_right_null_1", "RIGHT('hi', NULL)", NewSQLNull(knd)},
			{"sql_right_null_2", "RIGHT(NULL, 5)", NewSQLNull(knd)},
			// basic cases w/ string, int inputs and positive int length
			{"sql_right_base_case_0", "RIGHT('sDgcdcdc', 4)", NewSQLVarchar(knd, "dcdc")},
			{"sql_right_base_case_1", "RIGHT(124, 2)", NewSQLVarchar(knd, "24")},
			// negative lengths and 0 give empty string
			{"sql_right_negative_0", "RIGHT('hi', -1)", NewSQLVarchar(knd, "")},
			{"sql_right_negative_1", "RIGHT('hi', 0)", NewSQLVarchar(knd, "")},
			{"sql_right_negative_2", "RIGHT('hi', -2.5)", NewSQLVarchar(knd, "")},
			// float lengths should be rounded to closest int
			{"sql_right_float_0", "RIGHT('hello', 2.4)", NewSQLVarchar(knd, "lo")},
			{"sql_right_float_1", "RIGHT('hello', 2.5)", NewSQLVarchar(knd, "llo")},
			{"sql_right_float_2", "RIGHT(1234, 2.3)", NewSQLVarchar(knd, "34")},
			{"sql_right_float_3", "RIGHT(1234, 2.5)", NewSQLVarchar(knd, "234")},
			{"sql_right_float_4", "RIGHT('yo', 2.5)", NewSQLVarchar(knd, "yo")},
			// strings with spaces and symbols
			{"sql_right_symbols_0", "RIGHT('  ', 1)", NewSQLVarchar(knd, " ")},
			{"sql_right_symbols_1", "RIGHT('@!%', 2)", NewSQLVarchar(knd, "!%")},
			// boolean for string
			{"sql_right_bool_0", "RIGHT(true, 3)", NewSQLVarchar(knd, "1")},
			{"sql_right_bool_1", "RIGHT(false, 3)", NewSQLVarchar(knd, "0")},
			// boolean for length
			{"sql_right_bool_length_0", "RIGHT('hello', true)", NewSQLVarchar(knd, "o")},
			{"sql_right_bool_length_1", "RIGHT('hello', false)", NewSQLVarchar(knd, "")},
			// string for length
			{"sql_right_string_length", "RIGHT('hello', 'hi')", NewSQLVarchar(knd, "")},
			// len > length of string
			{"sql_right_edge", "RIGHT('hi', 5)", NewSQLVarchar(knd, "hi")},
			// string number as length
			{"sql_right_num_as_length_0", "RIGHT('hello', '2')", NewSQLVarchar(knd, "lo")},
			{"sql_right_num_as_length_1", "RIGHT('hello', '-3')", NewSQLVarchar(knd, "")},
			// unlike with floats, string #s always round down
			{"sql_right_float_as_length_0", "RIGHT('hello', '2.4')", NewSQLVarchar(knd, "lo")},
			{"sql_right_float_as_length_1", "RIGHT('hello', '2.6')", NewSQLVarchar(knd, "lo")},
			{"sql_round_0", "ROUND(NULL, NULL)", NewSQLNull(knd)},
			{"sql_round_1", "ROUND(NULL, 4)", NewSQLNull(knd)},
			{"sql_round_2", "ROUND(-16.55555, 4)", NewSQLFloat(knd, -16.5556)},
			{"sql_round_3", "ROUND(4.56, 1)", NewSQLFloat(knd, 4.6)},
			{"sql_round_4", "ROUND(-16.5, -1)", NewSQLFloat(knd, 0)},
			{"sql_round_5", "ROUND(-16.5)", NewSQLFloat(knd, -17)},
			{"sql_rtrim_0", "RTRIM(NULL)", NewSQLNull(knd)},
			{"sql_rtrim_1", "RTRIM('barbar   ')", NewSQLVarchar(knd, "barbar")},
			// LPAD(str, len, padStr)
			// basic case
			{"sql_lpad_0", "LPAD('hello', 7, 'x')", NewSQLVarchar(knd, "xxhello")},
			// nulls in various positions
			{"sql_lpad_null_0", "LPAD(NULL, 5, 'a')", NewSQLNull(knd)},
			{"sql_lpad_null_1", "LPAD('hi', NULL, 'a')", NewSQLNull(knd)},
			{"sql_lpad_null_2", "LPAD('hi', 5, NULL)", NewSQLNull(knd)},
			{"sql_lpad_null_3", "LPAD(NULL, NULL, NULL)", NewSQLNull(knd)},
			// str: empty
			{"sql_lpad_empty_0", "LPAD('', 0, 'a')", NewSQLVarchar(knd, "")},
			{"sql_lpad_empty_1", "LPAD('', 1, 'a')", NewSQLVarchar(knd, "a")},
			{"sql_lpad_empty_2", "LPAD('', 7, 'ab')", NewSQLVarchar(knd, "abababa")},
			// str: spaces and symbols
			{"sql_lpad_symbols_0", "LPAD(' hi', 4, 'x')", NewSQLVarchar(knd, "x hi")},
			{"sql_lpad_symbols_1", "LPAD('  ', 5, ' ')", NewSQLVarchar(knd, "     ")},
			{"sql_lpad_symbols_2", "LPAD('@!#_', 10, '.')", NewSQLVarchar(knd, "......@!#_")},
			{"sql_lpad_symbols_3", "LPAD('I♥NY', 8, 'x')", NewSQLVarchar(knd, "xxxxI♥NY")},
			{"sql_lpad_symbols_4", "LPAD('ƏŨ Ó€', 8, 'x')", NewSQLVarchar(knd, "xxxƏŨ Ó€")},
			{
				"sql_lpad_symbols_5",
				"LPAD('⅓ ⅔ † ‡ µ ¢ £', 8, 'x')",
				NewSQLVarchar(knd, "⅓ ⅔ † ‡ "),
			},
			{"sql_lpad_symbols_6", "LPAD('∞π∅≤≥≠≈', 8, 'x')", NewSQLVarchar(knd, "x∞π∅≤≥≠≈")},
			{"sql_lpad_symbols_7", "LPAD('hello', 8, '♥')", NewSQLVarchar(knd, "♥♥♥hello")},
			{"sql_lpad_symbols_8", "LPAD('hello', 8, 'ƏŨ')", NewSQLVarchar(knd, "ƏŨƏhello")},
			// str type: numbers
			{"sql_lpad_numbers_0", "LPAD(5, 4, 'a')", NewSQLVarchar(knd, "aaa5")},
			{"sql_lpad_numbers_1", "LPAD(10, 4, 'a')", NewSQLVarchar(knd, "aa10")},
			{"sql_lpad_numbers_2", "LPAD(10.2, 4, 'a')", NewSQLVarchar(knd, "10.2")},
			// str type: boolean
			{"sql_lpad_bool_0", "LPAD(true, 4, 'a')", NewSQLVarchar(knd, "aaa1")},
			{"sql_lpad_bool_1", "LPAD(false, 4, 'a')", NewSQLVarchar(knd, "aaa0")},
			// len < 0
			{"sql_lpad_neg_length", "LPAD('hi', -1, 'a')", NewSQLNull(knd)},
			// len = 0
			{"sql_lpad_zero", "LPAD('hi', 0, 'a')", NewSQLVarchar(knd, "")},
			// len <= len(str)
			{"sql_lpad_edge_0", "LPAD('hello', 2, 'x')", NewSQLVarchar(knd, "he")},
			{"sql_lpad_edge_1", "LPAD('hello', 5, 'x')", NewSQLVarchar(knd, "hello")},
			// len type: str
			{"sql_lpad_edge_2", "LPAD('hello', '5', 'x')", NewSQLVarchar(knd, "hello")},
			{"sql_lpad_edge_3", "LPAD('hello', '5.6', 'x')", NewSQLVarchar(knd, "hello")},
			{"sql_lpad_edge_3", "LPAD('hello', '6', 'x')", NewSQLVarchar(knd, "xhello")},
			{"sql_lpad_edge_4", "LPAD('hello', '6.2', 'x')", NewSQLVarchar(knd, "xhello")},
			// if can't be cast to #, then use length 0
			{"sql_lpad_edge_5", "LPAD('hello', 'a', 'x')", NewSQLVarchar(knd, "")},
			// len: floating point
			{"sql_lpad_edge_6", "LPAD('hello', 5.4, 'x')", NewSQLVarchar(knd, "hello")},
			{"sql_lpad_edge_7", "LPAD('hello', 5.5, 'x')", NewSQLVarchar(knd, "xhello")},
			// len float values close to 0 - round to closest int
			{"sql_lpad_edge_8", "LPAD('hello', 0.4, 'x')", NewSQLVarchar(knd, "")},
			{"sql_lpad_edge_9", "LPAD('hello', 0.5, 'x')", NewSQLVarchar(knd, "h")},
			{"sql_lpad_edge_10", "LPAD('hello', -0.4, 'x')", NewSQLVarchar(knd, "")},
			{"sql_lpad_edge_11", "LPAD('hello', -0.5, 'x')", NewSQLNull(knd)},
			// len string values close to 0 - always round toward 0
			{"sql_lpad_edge_12", "LPAD('hello', '0.4', 'x')", NewSQLVarchar(knd, "")},
			{"sql_lpad_edge_13", "LPAD('hello', '0.5', 'x')", NewSQLVarchar(knd, "")},
			{"sql_lpad_edge_14", "LPAD('hello', '-0.4', 'x')", NewSQLVarchar(knd, "")},
			{"sql_lpad_edge_15", "LPAD('hello', '-0.5', 'x')", NewSQLVarchar(knd, "")},
			// len: bool
			{"sql_lpad_edge_16", "LPAD('hello', true, 'x')", NewSQLVarchar(knd, "h")},
			{"sql_lpad_edge_17", "LPAD('hello', false, 'x')", NewSQLVarchar(knd, "")},
			// len(padStr) > 1
			{"sql_lpad_edge_18", "LPAD('hello', 7, 'xy')", NewSQLVarchar(knd, "xyhello")},
			{"sql_lpad_edge_19", "LPAD('hello', 8, 'xy')", NewSQLVarchar(knd, "xyxhello")},
			// padStr type: number
			{"sql_lpad_edge_20", "LPAD('hello', 7, 1)", NewSQLVarchar(knd, "11hello")},
			{"sql_lpad_edge_21", "LPAD('hello', 10, 1.1)", NewSQLVarchar(knd, "1.11.hello")},
			{"sql_lpad_edge_22", "LPAD('hello', 10, -1)", NewSQLVarchar(knd, "-1-1-hello")},

			// padStr type: boolean
			{"sql_lpad_edge_23", "LPAD('hello', 7, true)", NewSQLVarchar(knd, "11hello")},
			{"sql_lpad_edge_24", "LPAD('hello', 10, false)", NewSQLVarchar(knd, "00000hello")},
			// RPAD(str, len, padStr)
			// basic case
			{"sql_rpad_0", "RPAD('hello', 7, 'x')", NewSQLVarchar(knd, "helloxx")},
			// nulls in various positions
			{"sql_rpad_null_0", "RPAD(NULL, 5, 'a')", NewSQLNull(knd)},
			{"sql_rpad_null_1", "RPAD('hi', NULL, 'a')", NewSQLNull(knd)},
			{"sql_rpad_null_2", "RPAD('hi', 5, NULL)", NewSQLNull(knd)},
			{"sql_rpad_null_3", "RPAD(NULL, NULL, NULL)", NewSQLNull(knd)},
			// str: empty
			{"sql_rpad_str_empty_0", "RPAD('', 0, 'a')", NewSQLVarchar(knd, "")},
			{"sql_rpad_str_empty_1", "RPAD('', 1, 'a')", NewSQLVarchar(knd, "a")},
			{"sql_rpad_str_empty_2", "RPAD('', 7, 'ab')", NewSQLVarchar(knd, "abababa")},
			// str: spaces and symbols
			{"sql_rpad_symbols_0", "RPAD(' hi', 4, 'x')", NewSQLVarchar(knd, " hix")},
			{"sql_rpad_symbols_1", "RPAD('  ', 5, ' ')", NewSQLVarchar(knd, "     ")},
			{"sql_rpad_symbols_2", "RPAD('@!#_', 10, '.')", NewSQLVarchar(knd, "@!#_......")},
			{"sql_rpad_symbols_3", "RPAD('I♥NY', 8, 'x')", NewSQLVarchar(knd, "I♥NYxxxx")},
			{"sql_rpad_symbols_4", "RPAD('ƏŨ Ó€', 8, 'x')", NewSQLVarchar(knd, "ƏŨ Ó€xxx")},
			{
				"sql_rpad_symbols_5",
				"RPAD('⅓ ⅔ † ‡ µ ¢ £', 8, 'x')",
				NewSQLVarchar(knd, "⅓ ⅔ † ‡ "),
			},
			{"sql_rpad_symbols_6", "RPAD('∞π∅≤≥≠≈', 8, 'x')", NewSQLVarchar(knd, "∞π∅≤≥≠≈x")},
			{"sql_rpad_symbols_7", "RPAD('hello', 8, '♥')", NewSQLVarchar(knd, "hello♥♥♥")},
			{"sql_rpad_symbols_8", "RPAD('hello', 8, 'ƏŨ')", NewSQLVarchar(knd, "helloƏŨƏ")},
			// str type: numbers
			{"sql_rpad_numbers_0", "RPAD(5, 4, 'a')", NewSQLVarchar(knd, "5aaa")},
			{"sql_rpad_numbers_1", "RPAD(10, 4, 'a')", NewSQLVarchar(knd, "10aa")},
			{"sql_rpad_numbers_2", "RPAD(10.2, 4, 'a')", NewSQLVarchar(knd, "10.2")},
			// str type: boolean
			{"sql_rpad_bool_0", "RPAD(true, 4, 'a')", NewSQLVarchar(knd, "1aaa")},
			{"sql_rpad_bool_1", "RPAD(false, 4, 'a')", NewSQLVarchar(knd, "0aaa")},
			// len < 0
			{"sql_rpad_len", "RPAD('hi', -1, 'a')", NewSQLNull(knd)},
			// len = 0
			{"sql_rpad_len_1", "RPAD('hi', 0, 'a')", NewSQLVarchar(knd, "")},
			// len <= len(str)
			{"sql_rpad_len_2", "RPAD('hello', 2, 'x')", NewSQLVarchar(knd, "he")},
			{"sql_rpad_len_3", "RPAD('hello', 5, 'x')", NewSQLVarchar(knd, "hello")},
			// len type: str
			{"sql_rpad_len_4", "RPAD('hello', '5', 'x')", NewSQLVarchar(knd, "hello")},
			{"sql_rpad_len_5", "RPAD('hello', '5.6', 'x')", NewSQLVarchar(knd, "hello")},
			{"sql_rpad_len_6", "RPAD('hello', '6', 'x')", NewSQLVarchar(knd, "hellox")},
			{"sql_rpad_len_7", "RPAD('hello', '6.2', 'x')", NewSQLVarchar(knd, "hellox")},
			// if can't be cast to #, then use length 0
			{"sql_rpad_len_8", "RPAD('hello', 'a', 'x')", NewSQLVarchar(knd, "")},
			// len: floating point
			{"sql_rpad_len_9", "RPAD('hello', 5.4, 'x')", NewSQLVarchar(knd, "hello")},
			{"sql_rpad_len_10", "RPAD('hello', 5.5, 'x')", NewSQLVarchar(knd, "hellox")},
			// len float values close to 0 - round to closest int
			{"sql_rpad_len_11", "RPAD('hello', 0.4, 'x')", NewSQLVarchar(knd, "")},
			{"sql_rpad_len_12", "RPAD('hello', 0.5, 'x')", NewSQLVarchar(knd, "h")},
			{"sql_rpad_len_13", "RPAD('hello', -0.4, 'x')", NewSQLVarchar(knd, "")},
			{"sql_rpad_len_14", "RPAD('hello', -0.5, 'x')", NewSQLNull(knd)},
			// len string values close to 0 - always round toward 0
			{"sql_rpad_len_15", "RPAD('hello', '0.4', 'x')", NewSQLVarchar(knd, "")},
			{"sql_rpad_len_16", "RPAD('hello', '0.5', 'x')", NewSQLVarchar(knd, "")},
			{"sql_rpad_len_17", "RPAD('hello', '-0.4', 'x')", NewSQLVarchar(knd, "")},
			{"sql_rpad_len_18", "RPAD('hello', '-0.5', 'x')", NewSQLVarchar(knd, "")},
			// len: bool
			{"sql_rpad_len_19", "RPAD('hello', true, 'x')", NewSQLVarchar(knd, "h")},
			{"sql_rpad_len_20", "RPAD('hello', false, 'x')", NewSQLVarchar(knd, "")},
			// len(padStr) > 1
			{"sql_rpad_len_21", "RPAD('hello', 7, 'xy')", NewSQLVarchar(knd, "helloxy")},
			{"sql_rpad_len_22", "RPAD('hello', 8, 'xy')", NewSQLVarchar(knd, "helloxyx")},
			// padStr type: number
			{"sql_rpad_len_23", "RPAD('hello', 7, 1)", NewSQLVarchar(knd, "hello11")},
			{"sql_rpad_len_24", "RPAD('hello', 10, 1.1)", NewSQLVarchar(knd, "hello1.11.")},
			{"sql_rpad_len_25", "RPAD('hello', 10, -1)", NewSQLVarchar(knd, "hello-1-1-")},
			// padStr type: boolean
			{"sql_rpad_len_26", "RPAD('hello', 7, true)", NewSQLVarchar(knd, "hello11")},
			{"sql_rpad_len_27", "RPAD('hello', 10, false)", NewSQLVarchar(knd, "hello00000")},
			{"sql_second_0", "SECOND(NULL)", NewSQLNull(knd)},
			{"sql_second_1", "SECOND('sdg')", NewSQLInt64(knd, 0)},
			{"sql_second_2", "SECOND('10:23:52')", NewSQLInt64(knd, 52)},
			{"sql_second_3", "SECOND('10:61:52.24')", NewSQLNull(knd)},
			{"sql_second_4", "SECOND('10:23:52.24.25.26.27')", NewSQLInt64(knd, 52)},
			{"sql_sign_0", "SIGN(NULL)", NewSQLNull(knd)},
			{"sql_sign_1", "SIGN(-42)", NewSQLInt64(knd, -1)},
			{"sql_sign_2", "SIGN(0)", NewSQLInt64(knd, 0)},
			{"sql_sign_3", "SIGN(42)", NewSQLInt64(knd, 1)},
			{"sql_sign_4", "SIGN(42.0)", NewSQLInt64(knd, 1)},
			{"sql_sign_5", "SIGN(-42.0)", NewSQLInt64(knd, -1)},
			{"sql_sign_6", "SIGN('hello world')", NewSQLInt64(knd, 0)},
			{"sql_sin_0", "SIN(NULL)", NewSQLNull(knd)},
			{"sql_sin_1", "SIN(19)", NewSQLFloat(knd, 0.14987720966295234)},
			{"sql_sin_2", "SIN(-19)", NewSQLFloat(knd, -0.14987720966295234)},
			{"sql_sin_3", "SIN('C')", NewSQLFloat(knd, 0)},
			{"sql_sin_4", "SIN(0)", NewSQLFloat(knd, 0)},
			{"sql_space_0", "SPACE(NULL)", NewSQLNull(knd)},
			{"sql_space_1", "SPACE(5)", NewSQLVarchar(knd, "     ")},
			{"sql_space_2", "SPACE(-3)", NewSQLVarchar(knd, "")},
			{"sql_sort_0", "SQRT(NULL)", NewSQLNull(knd)},
			{"sql_sort_1", "SQRT('sdg')", NewSQLFloat(knd, 0)},
			{"sql_sort_2", "SQRT(-16)", NewSQLNull(knd)},
			{"sql_sort_3", "SQRT(4)", NewSQLFloat(knd, 2)},
			{"sql_sort_4", "SQRT(20)", NewSQLFloat(knd, 4.47213595499958)},
			{"sql_substring_0", "SUBSTRING(NULL, 4)", NewSQLNull(knd)},
			{"sql_substring_1", "SUBSTRING('foobarbar', NULL)", NewSQLNull(knd)},
			{"sql_substring_2", "SUBSTRING('foobarbar', 4, NULL)", NewSQLNull(knd)},
			{"sql_substring_3", "SUBSTRING('Quadratically', 5)", NewSQLVarchar(knd, "ratically")},
			{"sql_substring_4", "SUBSTRING('Quadratically', 5, 6)", NewSQLVarchar(knd, "ratica")},
			{"sql_substring_5", "SUBSTRING('Quadratically', 12, 2)", NewSQLVarchar(knd, "ly")},
			{"sql_substring_6", "SUBSTRING('Sakila', -3)", NewSQLVarchar(knd, "ila")},
			{"sql_substring_7", "SUBSTRING('Sakila', -5, 3)", NewSQLVarchar(knd, "aki")},
			{"sql_substring_8", "SUBSTRING('日本語', 2)", NewSQLVarchar(knd, "本語")},
			{"sql_substring_9", "SUBSTR(NULL, 4)", NewSQLNull(knd)},
			{"sql_substring_10", "SUBSTR('foobarbar', NULL)", NewSQLNull(knd)},
			{"sql_substring_11", "SUBSTR('foobarbar', 4, NULL)", NewSQLNull(knd)},
			{"sql_substring_12", "SUBSTR('Quadratically', 5)", NewSQLVarchar(knd, "ratically")},
			{"sql_substring_13", "SUBSTR('Quadratically', 5, 6)", NewSQLVarchar(knd, "ratica")},
			{"sql_substring_14", "SUBSTR('Sakila', -3)", NewSQLVarchar(knd, "ila")},
			{"sql_substring_15", "SUBSTR('Sakila', -5, 3)", NewSQLVarchar(knd, "aki")},
			{"sql_substring_16", "SUBSTR('日本語', 2)", NewSQLVarchar(knd, "本語")},
			{"sql_substring_17", "SUBSTR('five', 2, 2)", NewSQLVarchar(knd, "iv")},
			{"sql_substring_18", "SUBSTR('nine', 4, 9)", NewSQLVarchar(knd, "e")},
			{"sql_substring_19", "SUBSTR('five', 4, 3)", NewSQLVarchar(knd, "e")},
			{"sql_substring_20", "SUBSTR('five', -1, 1)", NewSQLVarchar(knd, "e")},
			{"sql_substring_21", "SUBSTR('five', 4, 0)", NewSQLVarchar(knd, "")},
			{"sql_substring_22", "SUBSTR('ZBA', 0)", NewSQLVarchar(knd, "")},
			{"sql_substring_23", "SUBSTR('ZBA', 0, 1)", NewSQLVarchar(knd, "")},
			{"sql_substring_24", "SUBSTR('ZBA', 0, -1)", NewSQLVarchar(knd, "")},
			{"sql_substring_25", "SUBSTR('ZBA', -1, 0)", NewSQLVarchar(knd, "")},
			{"sql_substring_26", "SUBSTR('ZBA', 1, 0)", NewSQLVarchar(knd, "")},
			{"sql_substring_27", "SUBSTR('ZBA', 0, 0)", NewSQLVarchar(knd, "")},
			{"sql_substring_28", "SUBSTRING(NULL from 4)", NewSQLNull(knd)},
			{"sql_substring_29", "SUBSTRING('foobarbar' from NULL)", NewSQLNull(knd)},
			{"sql_substring_30", "SUBSTRING('foobarbar' from 4 for NULL)", NewSQLNull(knd)},
			{
				"sql_substring_31",
				"SUBSTRING('Quadratically' FROM 5)",
				NewSQLVarchar(knd, "ratically"),
			},
			{
				"sql_substring_32",
				"SUBSTRING('Quadratically' FROM  5 for 6)",
				NewSQLVarchar(knd, "ratica"),
			},
			{
				"sql_substring_33",
				"SUBSTRING('Quadratically' from 12 FOR 2)",
				NewSQLVarchar(knd, "ly"),
			},
			{"sql_substring_34", "SUBSTRING('Sakila' FROM -3)", NewSQLVarchar(knd, "ila")},
			{"sql_substring_35", "SUBSTRING('Sakila' from -5 for 3)", NewSQLVarchar(knd, "aki")},
			{"sql_substring_36", "SUBSTRING('日本語' FROM  2)", NewSQLVarchar(knd, "本語")},
			{"sql_substring_37", "SUBSTR(NULL FROM 4)", NewSQLNull(knd)},
			{"sql_substring_38", "SUBSTR('foobarbar' FROM NULL)", NewSQLNull(knd)},
			{"sql_substring_39", "SUBSTR('foobarbar' FROM 4 FOR NULL)", NewSQLNull(knd)},
			{
				"sql_substring_40",
				"SUBSTR('Quadratically' FROM  5)",
				NewSQLVarchar(knd, "ratically"),
			},
			{
				"sql_substring_41",
				"SUBSTR('Quadratically' FROM  5 for 6)",
				NewSQLVarchar(knd, "ratica"),
			},
			{"sql_substring_42", "SUBSTR('Sakila' from -3)", NewSQLVarchar(knd, "ila")},
			{"sql_substring_43", "SUBSTR('Sakila' from -5 for 3)", NewSQLVarchar(knd, "aki")},
			{"sql_substring_44", "SUBSTR('日本語' from 2)", NewSQLVarchar(knd, "本語")},
			{"sql_substring_45", "SUBSTR('five' from 2 for 2)", NewSQLVarchar(knd, "iv")},
			{"sql_substring_46", "SUBSTR('nine' from 4 for  9)", NewSQLVarchar(knd, "e")},
			{"sql_substring_47", "SUBSTR('five' FROM 4 FOR 3)", NewSQLVarchar(knd, "e")},
			{"sql_substring_48", "SUBSTR('five' FROM -1 FOR  1)", NewSQLVarchar(knd, "e")},
			{"sql_substring_49", "SUBSTR('five' FROM 4 FOR  0)", NewSQLVarchar(knd, "")},
			{"sql_substring_50", "SUBSTR('ZBA' FROM 0)", NewSQLVarchar(knd, "")},
			{"sql_substring_51", "SUBSTR('ZBA' FROM 0 FOR  1)", NewSQLVarchar(knd, "")},
			{"sql_substring_52", "SUBSTR('ZBA' FROM 0 for  -1)", NewSQLVarchar(knd, "")},
			{"sql_substring_53", "SUBSTR('ZBA' from -1 for  0)", NewSQLVarchar(knd, "")},
			{"sql_substring_54", "SUBSTR('ZBA' from 1 FOR 0)", NewSQLVarchar(knd, "")},
			{"sql_substring_55", "SUBSTR('ZBA' from 0 for 0)", NewSQLVarchar(knd, "")},
			{"sql_substring_56", "SUBSTR('this', -5.2)", NewSQLVarchar(knd, "")},
			{"sql_substring_57", "SUBSTR('this' from -5.2)", NewSQLVarchar(knd, "")},
			{"sql_substring_58", "SUBSTR('this', 2.632)", NewSQLVarchar(knd, "is")},
			{"sql_substring_59", "SUBSTR('this', '2.632')", NewSQLVarchar(knd, "his")},
			{"sql_substring_60", "SUBSTR('this', '2.1')", NewSQLVarchar(knd, "his")},
			{"sql_substring_61", "SUBSTR('this' from -2.632)", NewSQLVarchar(knd, "his")},
			{"sql_substring_62", "SUBSTR('this', 2.4, 1.4)", NewSQLVarchar(knd, "h")},
			{"sql_substring_63", "SUBSTR('this' from 2.4 for -1.4 )", NewSQLVarchar(knd, "")},
			{"sql_substring_64", "SUBSTR('this', 1.6, 2.6)", NewSQLVarchar(knd, "his")},
			{"sql_substring_65", "SUBSTR('this', 1.6, '2.6')", NewSQLVarchar(knd, "hi")},
			{"sql_substring_66", "SUBSTR('this', 1.6, '2.1')", NewSQLVarchar(knd, "hi")},
			{"sql_substring_67", "SUBSTR('this', -11.6)", NewSQLVarchar(knd, "")},
			{"sql_substring_68", "SUBSTR(NULL, -4)", NewSQLNull(knd)},
			{"sql_substring_69", "SUBSTR(NULL, -4, 2)", NewSQLNull(knd)},
			{"sql_substring_70", "SUBSTR('this' FROM NULL FOR 2)", NewSQLNull(knd)},
			{"sql_substring_71", "SUBSTR('this', 2, NULL )", NewSQLNull(knd)},
			{"sql_substring_72", "SUBSTR('this' FROM 3 FOR NULL)", NewSQLNull(knd)},
			{
				"sql_substring_index_0",
				"SUBSTRING_INDEX('www.cmysql.com', '.', NULL)",
				NewSQLNull(knd),
			},
			{
				"sql_substring_index_1",
				"SUBSTRING_INDEX('www.cmysql.com', '.', 0)",
				NewSQLVarchar(knd, ""),
			},
			{
				"sql_substring_index_2",
				"SUBSTRING_INDEX('www.cmysql.com', '.', 1)",
				NewSQLVarchar(knd, "www"),
			},
			{
				"sql_substring_index_3",
				"SUBSTRING_INDEX('www.cmysql.com', '.c', 1)",
				NewSQLVarchar(knd, "www"),
			},
			{
				"sql_substring_index_4",
				"SUBSTRING_INDEX('www.cmysql.com', '.', 2)",
				NewSQLVarchar(knd, "www.cmysql"),
			},
			{
				"sql_substring_index_5",
				"SUBSTRING_INDEX('www.cmysql.com', '.', 1000)",
				NewSQLVarchar(knd, "www.cmysql.com"),
			},
			{
				"sql_substring_index_6",
				"SUBSTRING_INDEX('www.cmysql.com', '.c', 2)",
				NewSQLVarchar(knd, "www.cmysql"),
			},
			{
				"sql_substring_index_7",
				"SUBSTRING_INDEX('www.cmysql.com', '.', -2)",
				NewSQLVarchar(knd, "cmysql.com"),
			},
			{
				"sql_substring_index_8",
				"SUBSTRING_INDEX('www.cmysql.com', '.', -1)",
				NewSQLVarchar(knd, "com"),
			},
			{"sql_tan_0", "TAN(NULL)", NewSQLNull(knd)},
			{"sql_tan_1", "TAN(19)", NewSQLFloat(knd, 0.15158947061240008)},
			{"sql_tan_2", "TAN(-19)", NewSQLFloat(knd, -0.15158947061240008)},
			{"sql_tan_3", "TAN('C')", NewSQLFloat(knd, 0)},
			{"sql_tan_4", "TAN(0)", NewSQLFloat(knd, 0)},
			{"sql_time_to_sec_0", "TIME_TO_SEC(NULL)", NewSQLNull(knd)},
			{"sql_time_to_sec_1", "TIME_TO_SEC('22:23:00')", NewSQLFloat(knd, 80580)},
			{"sql_time_to_sec_2", "TIME_TO_SEC('12:34')", NewSQLFloat(knd, 45240)},
			{"sql_time_to_sec_3", "TIME_TO_SEC('00:39:38')", NewSQLFloat(knd, 2378)},
			{"sql_time_to_sec_4", "TIME_TO_SEC(1010103)", NewSQLFloat(knd, 363663)},
			{"sql_time_to_sec_5", "TIME_TO_SEC('2222')", NewSQLFloat(knd, 1342)},
			{"sql_time_to_sec_6", "TIME_TO_SEC(101010)", NewSQLFloat(knd, 36610)},
			{"sql_time_to_sec_7", "TIME_TO_SEC(-222)", NewSQLFloat(knd, -142)},
			{"sql_time_to_sec_8", "TIME_TO_SEC('-22:33:32')", NewSQLFloat(knd, -81212)},
			{"sql_time_to_sec_9", "TIME_TO_SEC(535911)", NewSQLFloat(knd, 194351)},
			{"sql_time_to_sec_10", "TIME_TO_SEC('-850:00:00')", NewSQLFloat(knd, -3020399)},
			{"sql_time_to_sec_11", "TIME_TO_SEC('-838:59:59')", NewSQLFloat(knd, -3020399)},
			{
				"sql_time_to_sec_12",
				"TIME_TO_SEC(CONCAT('48:2','4:59'))",
				NewSQLFloat(knd, 174299),
			},
			{"sql_time_to_sec_13", "TIME_TO_SEC(535959.9)", NewSQLFloat(knd, 194399)},
			{"sql_time_to_sec_14", "TIME_TO_SEC(534422333)", NewSQLNull(knd)},
			{"sql_time_to_sec_15", "TIME_TO_SEC(539911)", NewSQLNull(knd)},
			{"sql_time_to_sec_16", "TIME_TO_SEC(8991111)", NewSQLNull(knd)},
			{"sql_time_to_sec_17", "TIME_TO_SEC('-5359:11')", NewSQLFloat(knd, -3020399)},
			{"sql_time_to_sec_18", "TIME_TO_SEC('2004-07-09 10:17:35')", NewSQLFloat(knd, 37055)},
			{
				"sql_time_to_sec_19",
				"TIME_TO_SEC('2004-07-09 10:17:35.238238')",
				NewSQLFloat(knd, 37055),
			},
			{"sql_timediff_0", "TIMEDIFF('2000:11:11 00:00:00', NULL)", NewSQLNull(knd)},
			{"sql_timediff_1", "TIMEDIFF(NULL, '2000:11:11 00:00:00')", NewSQLNull(knd)},
			{
				"sql_timediff_2",
				"TIMEDIFF('2000:09:11 00:00:00', '2000:09:31 00:00:01:323211')",
				NewSQLNull(knd),
			},
			{
				"sql_timediff_3",
				"TIMEDIFF('2008-12-31 23:59:59.000001','2008-12-31 23:59:58.000001')",
				NewSQLVarchar(knd, "00:00:01"),
			},
			{
				"sql_timediff_4",
				"TIMEDIFF('2000:11:11 00:00:00', '2000:11:11 10:00:00.000231')",
				NewSQLVarchar(knd, "-10:00:00.000231"),
			},
			{
				"sql_timediff_5",
				"TIMEDIFF('2000:01:01 00:00:00','2000:01:01 00:00:00.000001')",
				NewSQLVarchar(knd, "-00:00:00.000001"),
			},
			{
				"sql_timediff_6",
				"TIMEDIFF('2008-12-31 23:59:59.000001','2008-12-30 01:01:01.000002')",
				NewSQLVarchar(knd, "46:58:57.999999"),
			},
			{
				"sql_timestampdiff_0",
				"TIMESTAMPDIFF(YEAR, DATE '2001-01-02', DATE '2002-01-02')",
				NewSQLInt64(knd, 1),
			},
			{
				"sql_timestampdiff_1",
				"TIMESTAMPDIFF(YEAR, DATE '2002-01-02', DATE '2001-01-02')",
				NewSQLInt64(knd, -1),
			},
			{
				"sql_timestampdiff_2",
				"TIMESTAMPDIFF(YEAR, DATE '2001-01-03', DATE '2002-01-02')",
				NewSQLInt64(knd, 0),
			},
			{
				"sql_timestampdiff_3",
				"TIMESTAMPDIFF(YEAR, DATE '2001-01-02', DATE '2002-01-03')",
				NewSQLInt64(knd, 1),
			},
			{
				"sql_timestampdiff_4",
				"TIMESTAMPDIFF(QUARTER, DATE '2002-04-02', DATE '2002-01-02')",
				NewSQLInt64(knd, -1),
			},
			{
				"sql_timestampdiff_5",
				"TIMESTAMPDIFF(QUARTER, DATE '2002-01-02', DATE '2002-06-02')",
				NewSQLInt64(knd, 1),
			},
			{
				"sql_timestampdiff_6",
				"TIMESTAMPDIFF(QUARTER, DATE '2002-01-02', DATE '2002-07-02')",
				NewSQLInt64(knd, 2),
			},
			{
				"sql_timestampdiff_7",
				"TIMESTAMPDIFF(QUARTER, DATE '2002-07-02', DATE '2002-01-02')",
				NewSQLInt64(knd, -2),
			},
			{
				"sql_timestampdiff_8",
				"TIMESTAMPDIFF(MONTH, DATE '2002-01-02', DATE '2002-02-01')",
				NewSQLInt64(knd, 0),
			},
			{
				"sql_timestampdiff_9",
				"TIMESTAMPDIFF(MONTH, DATE '2002-02-01', DATE '2001-01-02')",
				NewSQLInt64(knd, -12),
			},
			{
				"sql_timestampdiff_10",
				"TIMESTAMPDIFF(MONTH, DATE '2002-01-02', DATE '2002-02-02')",
				NewSQLInt64(knd, 1),
			},
			{
				"sql_timestampdiff_11",
				"TIMESTAMPDIFF(MONTH, DATE '2002-02-03', DATE '2002-01-02')",
				NewSQLInt64(knd, -1),
			},
			{
				"sql_timestampdiff_12",
				"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-16')",
				NewSQLInt64(knd, 2),
			},
			{
				"sql_timestampdiff_13",
				"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-15')",
				NewSQLInt64(knd, 1),
			},
			{
				"sql_timestampdiff_14",
				"TIMESTAMPDIFF(WEEK, DATE '2001-01-15', DATE '2001-01-02')",
				NewSQLInt64(knd, -1),
			},
			{
				"sql_timestampdiff_15",
				"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-17')",
				NewSQLInt64(knd, 2),
			},
			{
				"sql_timestampdiff_16",
				"TIMESTAMPDIFF(DAY, DATE '2003-01-04', DATE '2003-01-16')",
				NewSQLInt64(knd, 12),
			},
			{
				"sql_timestampdiff_17",
				"TIMESTAMPDIFF(DAY, DATE '2003-01-16', DATE '2003-01-04')",
				NewSQLInt64(knd, -12),
			},
			{
				"sql_timestampdiff_18",
				"TIMESTAMPDIFF(HOUR, DATE '2003-01-04', DATE '2003-01-06')",
				NewSQLInt64(knd, 48),
			},
			{
				"sql_timestampdiff_19",
				"TIMESTAMPDIFF(MINUTE, DATE '2003-01-04', DATE '2003-01-06')",
				NewSQLInt64(knd, 2880),
			},
			{
				"sql_timestampdiff_20",
				"TIMESTAMPDIFF(SECOND, DATE '2003-01-04', DATE '2003-01-05')",
				NewSQLInt64(knd, 86400),
			},
			{
				"sql_timestampdiff_21",
				"TIMESTAMPDIFF(MICROSECOND, DATE '2003-01-04', DATE '2003-01-05')",
				NewSQLInt64(knd, 86400000000),
			},
			{
				"sql_timestampdiff_22",
				"TIMESTAMPDIFF(MICROSECOND, TIMESTAMP '2002-01-01 12:30:09', " +
					"TIMESTAMP '2002-01-02 13:40:33')",
				NewSQLInt64(knd, 90624000000),
			},
			{
				"sql_timestampdiff_23",
				"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-01-02 12:30:09', " +
					"TIMESTAMP '2003-03-04 12:45:30')",
				NewSQLInt64(knd, 1),
			},
			{
				"sql_timestampdiff_24",
				"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-01-02 12:30:09', " +
					"TIMESTAMP '2002-03-04 12:45:30')",
				NewSQLInt64(knd, 0),
			},
			{
				"sql_timestampdiff_25",
				"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-03-04 12:45:30', " +
					"TIMESTAMP '2002-01-02 12:30:09')",
				NewSQLInt64(knd, 0),
			},
			{
				"sql_timestampdiff_26",
				"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2003-03-04 12:30:06', DATE '2002-03-04')",
				NewSQLInt64(knd, -1),
			},
			{
				"sql_timestampdiff_27",
				"TIMESTAMPDIFF(SQL_TSI_YEAR, DATE '2004-03-04', TIMESTAMP '2003-03-04 12:30:06')",
				NewSQLInt64(knd, 0),
			},
			{
				"sql_timestampdiff_28",
				"TIMESTAMPDIFF(SQL_TSI_QUARTER, DATE '2002-01-01', " +
					"TIMESTAMP '2002-04-01 12:30:06')",
				NewSQLInt64(knd, 1),
			},
			{
				"sql_timestampdiff_29",
				"TIMESTAMPDIFF(SQL_TSI_QUARTER, TIMESTAMP '2002-04-01 12:30:06', " +
					"DATE '2002-01-01')",
				NewSQLInt64(knd, -1),
			},
			{
				"sql_timestampdiff_30",
				"TIMESTAMPDIFF(SQL_TSI_QUARTER, TIMESTAMP '2002-01-01 12:30:06', " +
					"DATE '2002-04-01')",
				NewSQLInt64(knd, 0),
			},
			{
				"sql_timestampdiff_31",
				"TIMESTAMPDIFF(SQL_TSI_QUARTER, DATE '2002-04-01', " +
					"TIMESTAMP '2002-01-01 12:30:06')",
				NewSQLInt64(knd, 0),
			},
			{
				"sql_timestampdiff_32",
				"TIMESTAMPDIFF(SQL_TSI_MONTH, DATE '2002-01-01', TIMESTAMP '2002-03-01 12:30:09')",
				NewSQLInt64(knd, 2),
			},
			{
				"sql_timestampdiff_33",
				"TIMESTAMPDIFF(SQL_TSI_MONTH, TIMESTAMP '2002-03-01 12:30:09', DATE '2002-01-01')",
				NewSQLInt64(knd, -2),
			},
			{
				"sql_timestampdiff_34",
				"TIMESTAMPDIFF(SQL_TSI_MONTH, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-03-01')",
				NewSQLInt64(knd, 1),
			},
			{
				"sql_timestampdiff_35",
				"TIMESTAMPDIFF(SQL_TSI_MONTH, DATE '2002-03-01', TIMESTAMP '2002-01-01 12:30:09')",
				NewSQLInt64(knd, -1),
			},
			{
				"sql_timestampdiff_36",
				"TIMESTAMPDIFF(SQL_TSI_WEEK, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-08')",
				NewSQLInt64(knd, 0),
			},
			{
				"sql_timestampdiff_37",
				"TIMESTAMPDIFF(SQL_TSI_WEEK, DATE '2002-01-01', TIMESTAMP '2002-01-08 12:30:09')",
				NewSQLInt64(knd, 1),
			},
			{
				"sql_timestampdiff_38",
				"TIMESTAMPDIFF(SQL_TSI_WEEK, TIMESTAMP '2002-01-08 12:30:09', DATE '2002-01-01')",
				NewSQLInt64(knd, -1),
			},
			{
				"sql_timestampdiff_39",
				"TIMESTAMPDIFF(SQL_TSI_DAY, DATE '2002-01-01', TIMESTAMP '2002-01-02 12:30:09')",
				NewSQLInt64(knd, 1),
			},
			{
				"sql_timestampdiff_40",
				"TIMESTAMPDIFF(SQL_TSI_DAY, TIMESTAMP '2002-01-02 12:30:09', DATE '2002-01-01')",
				NewSQLInt64(knd, -1),
			},
			{
				"sql_timestampdiff_41",
				"TIMESTAMPDIFF(SQL_TSI_DAY, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')",
				NewSQLInt64(knd, 0),
			},
			{
				"sql_timestampdiff_42",
				"TIMESTAMPDIFF(SQL_TSI_HOUR, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')",
				NewSQLInt64(knd, 11),
			},
			{
				"sql_timestampdiff_43",
				"TIMESTAMPDIFF(SQL_TSI_HOUR, TIMESTAMP '2002-01-01 12:30:09', " +
					"TIMESTAMP '2002-01-02 11:02:33')",
				NewSQLInt64(knd, 22),
			},
			{
				"sql_timestampdiff_44",
				"TIMESTAMPDIFF(SQL_TSI_MINUTE, TIMESTAMP '2002-01-01 12:30:09', " +
					"TIMESTAMP '2002-01-01 13:02:33')",
				NewSQLInt64(knd, 32),
			},
			{
				"sql_timestampdiff_45",
				"TIMESTAMPDIFF(SQL_TSI_MINUTE, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')",
				NewSQLInt64(knd, 689),
			},
			{
				"sql_timestampdiff_46",
				"TIMESTAMPDIFF(SQL_TSI_SECOND, TIMESTAMP '2002-01-01 12:30:09', " +
					"TIMESTAMP '2002-01-02 14:40:33')",
				NewSQLInt64(knd, 94224),
			},
			{"sql_to_days_0", "TO_DAYS(NULL)", NewSQLNull(knd)},
			{"sql_to_days_1", "TO_DAYS('')", NewSQLNull(knd)},
			{"sql_to_days_2", "TO_DAYS('0000-00-00')", NewSQLNull(knd)},
			{"sql_to_days_3", "TO_DAYS('0000-01-01')", NewSQLInt64(knd, 1)},
			{"sql_to_days_4", "TO_DAYS('0000-11-11')", NewSQLInt64(knd, 315)},
			{"sql_to_days_5", "TO_DAYS('00-11-11')", NewSQLInt64(knd, 730800)},
			{"sql_to_days_6", "TO_DAYS('950501')", NewSQLInt64(knd, 728779)},
			{"sql_to_days_7", "TO_DAYS(950501)", NewSQLInt64(knd, 728779)},
			{"sql_to_days_8", "TO_DAYS('1995-05-01')", NewSQLInt64(knd, 728779)},
			{"sql_to_days_9", "TO_DAYS('2007-10-07')", NewSQLInt64(knd, 733321)},
			{"sql_to_days_10", "TO_DAYS(881111)", NewSQLInt64(knd, 726417)},
			{"sql_to_days_11", "TO_DAYS('2006-01-02')", NewSQLInt64(knd, 732678)},
			{"sql_to_days_12", "TO_DAYS('1452-04-15')", NewSQLInt64(knd, 530437)},
			{"sql_to_days_13", "TO_DAYS('4222-12-12')", NewSQLInt64(knd, 1542399)},
			{"sql_to_days_14", "TO_DAYS('2000-09-23 13:45:00')", NewSQLInt64(knd, 730751)},
			{"sql_to_days_15", "TO_DAYS('2000-09-24 13:45:00')", NewSQLInt64(knd, 730752)},
			{"sql_to_days_16", "TO_DAYS('2000-10-24 13:45:00')", NewSQLInt64(knd, 730782)},
			{"sql_to_seconds_0", "TO_SECONDS(NULL)", NewSQLNull(knd)},
			{"sql_to_seconds_1", "TO_SECONDS('')", NewSQLNull(knd)},
			{"sql_to_seconds_2", "TO_SECONDS('0000-00-00')", NewSQLNull(knd)},
			{"sql_to_seconds_3", "TO_SECONDS('0000-01-01')", NewSQLInt64(knd, 86400)},
			{"sql_to_seconds_4", "TO_SECONDS('0000-11-11')", NewSQLInt64(knd, 27216000)},
			{"sql_to_seconds_5", "TO_SECONDS('00-11-11')", NewSQLInt64(knd, 63141120000)},
			{"sql_to_seconds_6", "TO_SECONDS('950501')", NewSQLInt64(knd, 62966505600)},
			{"sql_to_seconds_7", "TO_SECONDS(950501)", NewSQLInt64(knd, 62966505600)},
			{"sql_to_seconds_8", "TO_SECONDS('1995-05-01')", NewSQLInt64(knd, 62966505600)},
			{"sql_to_seconds_9", "TO_SECONDS('2007-10-07')", NewSQLInt64(knd, 63358934400)},
			{"sql_to_seconds_10", "TO_SECONDS(881111)", NewSQLInt64(knd, 62762428800)},
			{"sql_to_seconds_11", "TO_SECONDS('2006-01-02')", NewSQLInt64(knd, 63303379200)},
			{"sql_to_seconds_12", "TO_SECONDS('1452-04-15')", NewSQLInt64(knd, 45829756800)},
			{"sql_to_seconds_13", "TO_SECONDS('4222-12-12')", NewSQLInt64(knd, 133263273600)},
			{
				"sql_to_seconds_14",
				"TO_SECONDS('2000-09-23 13:45:00')",
				NewSQLInt64(knd, 63136935900),
			},
			{
				"sql_to_seconds_15",
				"TO_SECONDS('2000-09-24 13:45:00')",
				NewSQLInt64(knd, 63137022300),
			},
			{
				"sql_to_seconds_16",
				"TO_SECONDS('2000-10-24 13:45:00')",
				NewSQLInt64(knd, 63139614300),
			},
			{
				"sql_to_seconds_17",
				"TO_SECONDS('2000-10-24 15:45:00')",
				NewSQLInt64(knd, 63139621500),
			},
			{
				"sql_to_seconds_18",
				"TO_SECONDS('2000-10-24 13:47:00')",
				NewSQLInt64(knd, 63139614420),
			},
			{
				"sql_to_seconds_19",
				"TO_SECONDS('2000-10-24 13:45:59')",
				NewSQLInt64(knd, 63139614359),
			},
			{"sql_trim_0", "TRIM(NULL)", NewSQLNull(knd)},
			{"sql_trim_1", "TRIM('   bar   ')", NewSQLVarchar(knd, "bar")},
			{"sql_trim_2", "TRIM(BOTH 'xyz' FROM 'xyzbarxyzxyz')", NewSQLVarchar(knd, "bar")},
			{
				"sql_trim_3",
				"TRIM(LEADING 'xyz' FROM 'xyzbarxyzxyz')",
				NewSQLVarchar(knd, "barxyzxyz"),
			},
			{
				"sql_trim_4",
				"TRIM(TRAILING 'xyz' FROM 'xyzbarxyzxyz')",
				NewSQLVarchar(knd, "xyzbar"),
			},
			{"sql_trim_5", "TRIM('xyz' FROM 'xyzbarxyzxyz')", NewSQLVarchar(knd, "bar")},
			{"sql_truncate_0", "TRUNCATE(NULL, 2)", NewSQLNull(knd)},
			{"sql_truncate_1", "TRUNCATE(1234.1234, NULL)", NewSQLNull(knd)},
			{"sql_truncate_2", "TRUNCATE(1 / 0, 2)", NewSQLNull(knd)},
			{"sql_truncate_3", "TRUNCATE(1234.1234, 1 / 0)", NewSQLNull(knd)},
			{"sql_truncate_4", "TRUNCATE(1234.1234, 3)", NewSQLFloat(knd, 1234.123)},
			{"sql_truncate_5", "TRUNCATE(1234.1234, 5)", NewSQLFloat(knd, 1234.1234)},
			{"sql_truncate_6", "TRUNCATE(1234.1234, 0)", NewSQLFloat(knd, 1234)},
			{"sql_truncate_7", "TRUNCATE(1234.1234, -3)", NewSQLFloat(knd, 1000)},
			{"sql_truncate_8", "TRUNCATE(1234.1234, -5)", NewSQLFloat(knd, 0)},
			{"sql_truncate_9", "TRUNCATE(-1234.1234, 3)", NewSQLFloat(knd, -1234.123)},
			{"sql_truncate_10", "TRUNCATE(-1234.1234, -3)", NewSQLFloat(knd, -1000)},
			{"sql_ucase_0", "UCASE(NULL)", NewSQLNull(knd)},
			{"sql_ucase_1", "UCASE('sdg')", NewSQLVarchar(knd, "SDG")},
			{"sql_ucase_2", "UCASE(124)", NewSQLVarchar(knd, "124")},
			{"sql_ucase_3", "UPPER(NULL)", NewSQLNull(knd)},
			{"sql_ucase_4", "UPPER('')", NewSQLVarchar(knd, "")},
			{"sql_ucase_5", "UPPER('a')", NewSQLVarchar(knd, "A")},
			{"sql_ucase_6", "UPPER('AWESOME')", NewSQLVarchar(knd, "AWESOME")},
			{"sql_ucase_7", "UPPER('AwEsOmE')", NewSQLVarchar(knd, "AWESOME")},
			{"sql_unix_timestamp_0", "UNIX_TIMESTAMP(NULL)", NewSQLNull(knd)},
			{"sql_unix_timestamp_1", "UNIX_TIMESTAMP('1923-12-12')", NewSQLFloat(knd, 0)},
			/*
				These tests will fail if run on a server in a timezone
				different from EST (-05:00) - thus are flaky and commented out.
				test{
					"sql_unix_timestamp_2",
					"UNIX_TIMESTAMP('2015-11-13 10:20:19')",
					SQLUint(1447428019),
				},
				test{
					"sql_unix_timestamp_3",
					"UNIX_TIMESTAMP('2017-03-27 03:00:00')",
					SQLUint(1490598000),
				},
				test{
					"sql_unix_timestamp_4",
					"UNIX_TIMESTAMP('2012-11-17 12:00:00')",
					SQLUint(1353171600),
				},
				test{"sql_unix_timestamp_5", "UNIX_TIMESTAMP('1985-03-21')", SQLUint(480229200)},
				test{"sql_unix_timestamp_6", "UNIX_TIMESTAMP('1985')", SQLFloat(0)},
				test{"sql_unix_timestamp_7", "UNIX_TIMESTAMP('1985-12')", SQLFloat(0)},
				test{"sql_unix_timestamp_8", "UNIX_TIMESTAMP('1985-12-aa')", SQLFloat(0)},
				test{"sql_unix_timestamp_9", "UNIX_TIMESTAMP('1985-12-')", SQLFloat(0)},
				test{"sql_unix_timestamp_10", "UNIX_TIMESTAMP('1985-12-1')", SQLUint(502261200)},
				test{"sql_unix_timestamp_11", "UNIX_TIMESTAMP('1985-12-01')", SQLUint(502261200)},
			*/
			{"sql_week_0", "WEEK(NULL)", NewSQLNull(knd)},
			{"sql_week_1", "WEEK('sdg')", NewSQLNull(knd)},
			{"sql_week_2", "WEEK('2016-1-01 10:23:52')", NewSQLInt64(knd, 0)},
			{"sql_week_3", "WEEK(DATE '2009-1-01')", NewSQLInt64(knd, 0)},
			{"sql_week_4", "WEEK(DATE '2009-1-01',0)", NewSQLInt64(knd, 0)},
			{"sql_week_5", "WEEK(DATE '2009-1-01','str')", NewSQLInt64(knd, 0)},
			{"sql_week_6", "WEEK(DATE '2009-1-01',1)", NewSQLInt64(knd, 1)},
			{"sql_week_7", "WEEK(DATE '2009-1-01',2)", NewSQLInt64(knd, 52)},
			{"sql_week_8", "WEEK(DATE '2009-1-01',3)", NewSQLInt64(knd, 1)},
			{"sql_week_9", "WEEK(DATE '2009-1-01',4)", NewSQLInt64(knd, 0)},
			{"sql_week_10", "WEEK(DATE '2009-1-01',5)", NewSQLInt64(knd, 0)},
			{"sql_week_11", "WEEK(DATE '2009-1-01',6)", NewSQLInt64(knd, 53)},
			{"sql_week_12", "WEEK(DATE '2009-1-01',7)", NewSQLInt64(knd, 52)},
			{"sql_week_13", "WEEK(DATE '2009-1-05')", NewSQLInt64(knd, 1)},
			{"sql_week_14", "WEEK(DATE '2009-1-05',1)", NewSQLInt64(knd, 2)},
			{"sql_week_15", "WEEK(DATE '2009-1-05',2)", NewSQLInt64(knd, 1)},
			{"sql_week_16", "WEEK(DATE '2009-1-05',3)", NewSQLInt64(knd, 2)},
			{"sql_week_17", "WEEK(DATE '2009-1-05',4)", NewSQLInt64(knd, 1)},
			{"sql_week_18", "WEEK(DATE '2009-1-05',5)", NewSQLInt64(knd, 1)},
			{"sql_week_19", "WEEK(DATE '2009-1-05',6)", NewSQLInt64(knd, 1)},
			{"sql_week_20", "WEEK(DATE '2009-1-05',7)", NewSQLInt64(knd, 1)},
			{"sql_week_21", "WEEK(DATE '2009-12-31')", NewSQLInt64(knd, 52)},
			{"sql_week_22", "WEEK(DATE '2009-12-31',1)", NewSQLInt64(knd, 53)},
			{"sql_week_23", "WEEK(DATE '2009-12-31',2)", NewSQLInt64(knd, 52)},
			{"sql_week_24", "WEEK(DATE '2009-12-31',3)", NewSQLInt64(knd, 53)},
			{"sql_week_25", "WEEK(DATE '2009-12-31',4)", NewSQLInt64(knd, 52)},
			{"sql_week_26", "WEEK(DATE '2009-12-31',5)", NewSQLInt64(knd, 52)},
			{"sql_week_27", "WEEK(DATE '2009-12-31',6)", NewSQLInt64(knd, 52)},
			{"sql_week_28", "WEEK(DATE '2009-12-31',7)", NewSQLInt64(knd, 52)},
			{"sql_week_29", "WEEK(DATE '2007-12-31')", NewSQLInt64(knd, 52)},
			{"sql_week_30", "WEEK(DATE '2007-12-31',1)", NewSQLInt64(knd, 53)},
			{"sql_week_31", "WEEK(DATE '2007-12-31',2)", NewSQLInt64(knd, 52)},
			{"sql_week_32", "WEEK(DATE '2007-12-31',3)", NewSQLInt64(knd, 1)},
			{"sql_week_33", "WEEK(DATE '2007-12-31',4)", NewSQLInt64(knd, 53)},
			{"sql_week_34", "WEEK(DATE '2007-12-31',5)", NewSQLInt64(knd, 53)},
			{"sql_week_35", "WEEK(DATE '2007-12-31',6)", NewSQLInt64(knd, 1)},
			{"sql_week_36", "WEEK(DATE '2007-12-31',7)", NewSQLInt64(knd, 53)},
			{"sql_weekday_0", "WEEKDAY(NULL)", NewSQLNull(knd)},
			{"sql_weekday_1", "WEEKDAY('sdg')", NewSQLNull(knd)},
			{"sql_weekday_2", "WEEKDAY('2016-1-01 10:23:52')", NewSQLInt64(knd, 4)},
			{"sql_weekday_3", "WEEKDAY('2005-05-11')", NewSQLInt64(knd, 2)},
			{"sql_weekday_4", "WEEKDAY(DATE '2016-7-10')", NewSQLInt64(knd, 6)},
			{"sql_weekday_5", "WEEKDAY(DATE '2016-7-11')", NewSQLInt64(knd, 0)},
			{"sql_weekday_6", "WEEKDAY(TIMESTAMP '2016-7-13 21:22:23')", NewSQLInt64(knd, 2)},
			{"sql_weekofyear_0", "WEEKOFYEAR(NULL)", NewSQLNull(knd)},
			{"sql_weekofyear_1", "WEEKOFYEAR('sdg')", NewSQLNull(knd)},
			{"sql_weekofyear_2", "WEEKOFYEAR('2008-02-20')", NewSQLInt64(knd, 8)},
			{"sql_weekofyear_3", "WEEKOFYEAR('2009-01-01')", NewSQLInt64(knd, 1)},
			{"sql_weekofyear_4", "WEEKOFYEAR(DATE '2009-01-05')", NewSQLInt64(knd, 2)},
			{"sql_subtract_expr_0", "0 - 0", NewSQLInt64(knd, 0)},
			{"sql_subtract_expr_1", "-1 - 1", NewSQLInt64(knd, -2)},
			{"sql_subtract_expr_2", "10 - 32", NewSQLInt64(knd, -22)},
			{"sql_subtract_expr_3", "-10 - -32", NewSQLInt64(knd, 22)},
			{"sql_unary_minus_0", "- 10", NewSQLInt64(knd, -10)},
			{"sql_unary_minus_1", "- a", NewSQLInt64(knd, -123)},
			{"sql_unary_minus_2", "- b", NewSQLInt64(knd, -456)},
			{"sql_unary_minus_3", "- null", NewSQLNull(knd)},
			{"sql_unary_minus_4", "- true", NewSQLInt64(knd, -1)},
			{"sql_unary_minus_5", "- false", NewSQLInt64(knd, 0)},
			{"sql_unary_minus_6", "- date '2005-05-11'", NewSQLInt64(knd, -20050511)},
			{
				"sql_unary_minus_7",
				"- timestamp '2005-05-11 12:22:04'",
				NewSQLInt64(knd, -20050511122204),
			},
			{"sql_unary_minus_8", "- '4' ", NewSQLFloat(knd, -4)},
			{"sql_unary_minus_9", "- 6.7", NewSQLDecimal128(knd, decimal.New(-67, -1))},
			{"sql_unary_minus_10", "- '3.3'", NewSQLFloat(knd, -3.3)},
			{"sql_variable_expr_0", "@@autocommit", NewSQLBool(knd, true)},
			{"sql_variable_expr_1", "@@global.autocommit", NewSQLBool(knd, true)},
			{"sql_unary_plus_expr_0", "+1", NewSQLInt64(knd, 1)},
			{"sql_unary_plus_expr_1", "+'string'", NewSQLVarchar(knd, "string")},
			{"sql_unary_plus_expr_2", "+a", NewSQLInt64(knd, 123)},
			{"sql_early_eval_0", "(1, 3) > (2, 4)", NewSQLBool(knd, false)},
			{"sql_early_eval_1", "(1, 3) > ROW(2, 4)", NewSQLBool(knd, false)},
		}

		runTests(t, execCfg, execState, tests)

		// aggregation tests
		var t1, t2 time.Time
		t1 = time.Now()
		t2 = t1.Add(time.Hour)

		aggRows := []*Row{
			{Data: RowValues{
				{SelectID: 1, Database: "test", Table: "bar", Name: "a",
					Data: NewSQLNull(knd)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "b",
					Data: NewSQLInt64(knd, 3)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "c",
					Data: NewSQLNull(knd)},
				{
					SelectID: 1,
					Database: "test",
					Table:    "bar",
					Name:     "g",
					Data:     NewSQLDate(knd, t1),
				},
			}},
			{Data: RowValues{
				{SelectID: 1, Database: "test", Table: "bar", Name: "a",
					Data: NewSQLInt64(knd, 3)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "b",
					Data: NewSQLNull(knd)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "c",
					Data: NewSQLNull(knd)},
				{
					SelectID: 1,
					Database: "test",
					Table:    "bar",
					Name:     "g",
					Data:     NewSQLDate(knd, t2),
				},
			}},
			{Data: RowValues{
				{SelectID: 1, Database: "test", Table: "bar", Name: "a",
					Data: NewSQLInt64(knd, 5)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "b",
					Data: NewSQLInt64(knd, 6)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "c",
					Data: NewSQLNull(knd)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "g",
					Data: NewSQLNull(knd)},
			}},
		}
		aggState := NewExecutionState().WithRows(aggRows...)

		aggTests := []test{
			{"sql_agg_expr_avg_0", "AVG(NULL)", NewSQLNull(knd)},
			{"sql_agg_expr_avg_1", "AVG(a)", NewSQLFloat(knd, 4)},
			{"sql_agg_expr_avg_2", "AVG(b)", NewSQLFloat(knd, 4.5)},
			{"sql_agg_expr_avg_3", "AVG(c)", NewSQLNull(knd)},
			{"sql_agg_expr_avg_4", "AVG('a')", NewSQLFloat(knd, 0)},
			{"sql_agg_expr_avg_5", "AVG(-20)", NewSQLFloat(knd, -20)},
			{"sql_agg_expr_avg_6", "AVG(20)", NewSQLFloat(knd, 20)},
			{"sql_count_expr_0", "COUNT(NULL)", NewSQLUint64(knd, 0)},
			{"sql_count_expr_1", "COUNT(a)", NewSQLUint64(knd, 2)},
			{"sql_count_expr_2", "COUNT(b)", NewSQLUint64(knd, 2)},
			{"sql_count_expr_3", "COUNT(c)", NewSQLUint64(knd, 0)},
			{"sql_count_expr_4", "COUNT(g)", NewSQLUint64(knd, 2)},
			{"sql_count_expr_5", "COUNT('a')", NewSQLUint64(knd, 3)},
			{"sql_count_expr_6", "COUNT(-20)", NewSQLUint64(knd, 3)},
			{"sql_count_expr_7", "COUNT(20)", NewSQLUint64(knd, 3)},
			{"sql_group_concat_expr_1", "GROUP_CONCAT(a)", NewSQLVarchar(knd, "3,5")},
			{"sql_group_concat_expr_2", "GROUP_CONCAT(a,b)", NewSQLVarchar(knd, "56")},
			{"sql_group_concat_expr_3", "GROUP_CONCAT(DISTINCT a)", NewSQLVarchar(knd, "3,5")},
			{"sql_group_concat_expr_4", "GROUP_CONCAT(a SEPARATOR \"hi\")",
				NewSQLVarchar(knd, "3hi5")},
			{"sql_group_concat_expr_5", "GROUP_CONCAT(a,c)", NewSQLNull(knd)},
			{"sql_group_concat_expr_6", "GROUP_CONCAT(null)", NewSQLNull(knd)},
			{"sql_min_expr_0", "MIN(NULL)", NewSQLNull(knd)},
			{"sql_min_expr_1", "MIN(a)", NewSQLInt64(knd, 3)},
			{"sql_min_expr_2", "MIN(b)", NewSQLInt64(knd, 3)},
			{"sql_min_expr_3", "MIN(c)", NewSQLNull(knd)},
			{"sql_min_expr_4", "MIN('a')", NewSQLVarchar(knd, "a")},
			{"sql_min_expr_5", "MIN(-20)", NewSQLInt64(knd, -20)},
			{"sql_min_expr_6", "MIN(20)", NewSQLInt64(knd, 20)},
			{"sql_max_expr_0", "MAX(NULL)", NewSQLNull(knd)},
			{"sql_max_expr_1", "MAX(a)", NewSQLInt64(knd, 5)},
			{"sql_max_expr_2", "MAX(b)", NewSQLInt64(knd, 6)},
			{"sql_max_expr_3", "MAX(c)", NewSQLNull(knd)},
			{"sql_max_expr_4", "MAX('a')", NewSQLVarchar(knd, "a")},
			{"sql_max_expr_5", "MAX(-20)", NewSQLInt64(knd, -20)},
			{"sql_max_expr_6", "MAX(20)", NewSQLInt64(knd, 20)},
			{"sql_sleep_expr_0", "SLEEP(1)", NewSQLInt64(knd, 0)},
			{"sql_sleep_expr_1", "SLEEP(1.5)", NewSQLInt64(knd, 0)},
			{"sql_sleep_expr_2", "SLEEP(0)", NewSQLInt64(knd, 0)},
			{"sql_sum_expr_0", "SUM(NULL)", NewSQLNull(knd)},
			{"sql_sum_expr_1", "SUM(a)", NewSQLFloat(knd, 8)},
			{"sql_sum_expr_2", "SUM(b)", NewSQLFloat(knd, 9)},
			{"sql_sum_expr_3", "SUM(c)", NewSQLNull(knd)},
			{"sql_sum_expr_4", "SUM('a')", NewSQLFloat(knd, 0)},
			{"sql_sum_expr_5", "SUM(-20)", NewSQLFloat(knd, -60)},
			{"sql_sum_expr_6", "SUM(20)", NewSQLFloat(knd, 60)},
			{"sql_std_expr_0", "STD(NULL)", NewSQLNull(knd)},
			{"sql_std_dev_expr", "STDDEV(a)", NewSQLFloat(knd, 1)},
			{"sql_std_dev_pop_expr", "STDDEV_POP(b)", NewSQLFloat(knd, 1.5)},
			{"sql_std_expr_1", "STD(c)", NewSQLNull(knd)},
			{"sql_std_dev_samp_expr_0", "STDDEV_SAMP(NULL)", NewSQLNull(knd)},
			{"sql_std_dev_samp_expr_1", "STDDEV_SAMP(a)", NewSQLFloat(knd, 1.4142135623730951)},
			{"sql_std_dev_samp_expr_2", "STDDEV_SAMP(b)", NewSQLFloat(knd, 2.1213203435596424)},
			{"sql_std_dev_samp_expr_3", "STDDEV_SAMP(c)", NewSQLNull(knd)},
		}
		runTests(t, execCfg, aggState, aggTests)

		// type tests
		typeTests := []typeTest{
			{"sql_coalesce_type_0", "COALESCE(NULL, 1, 'A')", EvalString},
			{"sql_coalesce_type_1", "COALESCE(NULL, 1, 23)", EvalInt64},
			{"sql_convert_type_0", "CONVERT(DATE '2006-05-11', SIGNED)", EvalInt64},
			{"sql_convert_type_1", "CONVERT(true, SQL_DOUBLE)", EvalDouble},
			{"sql_convert_type_2", "CONVERT('16a', CHAR)", EvalString},
			{"sql_convert_type_3", "CONVERT('2006-05-11', DATE)", EvalDate},
			{
				"sql_convert_type_4",
				"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATETIME)",
				EvalDatetime,
			},
			{
				"sql_convert_type_5",
				"CONVERT(DATE '2006-05-11', SQL_TIMESTAMP)",
				EvalDatetime,
			},
			{"sql_date_add_type_0", "DATE_ADD('2002-01-02', INTERVAL 1 YEAR)",
				EvalDatetime},
			{
				"sql_date_add_type_1",
				"DATE_ADD(DATE '2002-01-02', INTERVAL 1 HOUR)",
				EvalDatetime,
			},
			{
				"sql_date_add_type_2",
				"DATE_ADD(TIMESTAMP '2003-01-02 10:28:06', INTERVAL '2 2:3' DAY_SECOND)",
				EvalDatetime,
			},
			{"sql_date_sub_type_0", "DATE_SUB('2002-01-02', INTERVAL 1 YEAR)",
				EvalDatetime},
			{
				"sql_date_sub_type_1",
				"DATE_SUB(DATE '2002-01-02', INTERVAL 1 HOUR)",
				EvalDatetime,
			},
			{
				"sql_date_sub_type_2",
				"DATE_SUB(TIMESTAMP '2003-01-02 10:28:06', INTERVAL '2 2:3' DAY_SECOND)",
				EvalDatetime,
			},
			{
				"sql_greatest_type_0",
				"GREATEST(DATE '2005-05-11', DATE '2006-05-11', DATE '2000-05-11')",
				EvalDate,
			},
			{"sql_greatest_type_1", "GREATEST(1, 123.52, 'something')",
				EvalDecimal128},
			{"sql_if_type_0", "IF('ca.gh', 4, 5)", EvalInt64},
			{"sql_if_type_1", "IF('ca.gh', 4, 5.3)", EvalDecimal128},
			{"sql_if_type_2", "IF('ca.gh', 'sdf', 5.2)", EvalString},
			{"sql_if_type_3", "IF('ca.gh', 'sdf', NULL)", EvalString},
			{"sql_if_null_type_0", "IFNULL(4, 5)", EvalInt64},
			{"sql_if_null_type_1", "IFNULL(4, 5.3)", EvalDecimal128},
			{"sql_if_null_type_2", "IFNULL('sdf', NULL)", EvalString},
			{"sql_interval_type_0", "INTERVAL(4, 5)", EvalInt64},
			{"sql_interval_type_1", "INTERVAL(4, 5.3)", EvalInt64},
			{"sql_interval_type_2", "INTERVAL(NULL, 4)", EvalInt64},
			{"sql_null_if_type_0", "NULLIF(3, null)", EvalInt64},
			{"sql_null_if_type_1", "NULLIF('abc', 'abc')", EvalString},
			{
				"sql_least_type_0",
				"LEAST(DATE '2005-05-11', DATE '2006-05-11', DATE '2000-05-11')",
				EvalDate,
			},
			{"sql_least_type_1", "LEAST(1, 123.52, 'something')", EvalDecimal128},
			{"sql_str_to_date_type_0", "STR_TO_DATE('2000-01-01', '%Y-%m-%d')", EvalDate},
			{"sql_str_to_date_type_1", "STR_TO_DATE('2000-01-01', 'hello%iworld')", EvalDatetime},
			{"sql_str_to_date_type_2", "STR_TO_DATE('2000-01-01', 'hello%%Hworld')", EvalDatetime},
			{"sql_str_to_date_type_3", "STR_TO_DATE('2000-01-01', 'hi%%wd')", EvalDate},
			{"sql_str_to_date_type_4", "STR_TO_DATE('2000-01-01', '%S%Y')", EvalDatetime},
			{
				"sql_timestampadd_type_0",
				"TIMESTAMPADD(SQL_TSI_QUARTER, 2, DATE '2002-07-02')",
				EvalDatetime,
			},
			{
				"sql_timestampadd_type_1",
				"TIMESTAMPADD(SQL_TSI_SECOND, 1, TIMESTAMP '2003-01-02 12:30:08')",
				EvalDatetime,
			},
		}
		runTypeTests(t, typeTests)

		// error tests
		t.Run("argument validation", func(t *testing.T) {
			strVal := NewSQLValueExpr(NewSQLVarchar(knd, "bar"))
			intVal := NewSQLValueExpr(NewSQLInt64(knd, 1))
			floatVal := NewSQLValueExpr(NewSQLFloat(knd, 1))
			boolVal := NewSQLValueExpr(NewSQLBool(knd, true))

			// errTests is a list of test cases for the binary and unary operator and agg function SQLExprs. They
			// have custom reconcile implementations, so they are all tested to check for errors when the argument
			// types are unexpected.
			// This list does not contain tests for the scalar functions since those all have their reconcile, Evaluate,
			// and FoldConstants implementations generated the same way. They (mostly) all convert their arguments to
			// the expected types and all call validateArgs in their Evaluate and FoldConstants implementations, so
			// the unit tests for validateArgs is sufficient for them.
			errTests := []errTest{
				// arithmetic expressions: can evaluate if both arguments are numeric, cannot evaluate if at least one is not.
				{"add(int,int)", NewSQLAddExpr(intVal, intVal), true},
				{"add(float,float)", NewSQLAddExpr(floatVal, floatVal), true},
				{"add(int,float)", NewSQLAddExpr(intVal, floatVal), true},
				{"add(bool,bool)", NewSQLAddExpr(boolVal, boolVal), false},
				{"add(int,bool)", NewSQLAddExpr(intVal, boolVal), false},
				{"add(float,bool)", NewSQLAddExpr(floatVal, boolVal), false},
				{"add(string,string)", NewSQLAddExpr(strVal, strVal), false},
				{"add(int,string)", NewSQLAddExpr(intVal, strVal), false},
				{"add(float,string)", NewSQLAddExpr(floatVal, strVal), false},

				{"div(int,int)", NewSQLDivideExpr(intVal, intVal), true},
				{"div(float,float)", NewSQLDivideExpr(floatVal, floatVal), true},
				{"div(int,float)", NewSQLDivideExpr(intVal, floatVal), true},
				{"div(bool,bool)", NewSQLDivideExpr(boolVal, boolVal), false},
				{"div(int,bool)", NewSQLDivideExpr(intVal, boolVal), false},
				{"div(float,bool)", NewSQLDivideExpr(floatVal, boolVal), false},
				{"div(string,string)", NewSQLDivideExpr(strVal, strVal), false},
				{"div(int,string)", NewSQLDivideExpr(intVal, strVal), false},
				{"div(float,string)", NewSQLDivideExpr(floatVal, strVal), false},

				{"idiv(int,int)", NewSQLIDivideExpr(intVal, intVal), true},
				{"idiv(float,float)", NewSQLIDivideExpr(floatVal, floatVal), true},
				{"idiv(int,float)", NewSQLIDivideExpr(intVal, floatVal), true},
				{"idiv(bool,bool)", NewSQLIDivideExpr(boolVal, boolVal), false},
				{"idiv(int,bool)", NewSQLIDivideExpr(intVal, boolVal), false},
				{"idiv(float,bool)", NewSQLIDivideExpr(floatVal, boolVal), false},
				{"idiv(string,string)", NewSQLIDivideExpr(strVal, strVal), false},
				{"idiv(int,string)", NewSQLIDivideExpr(intVal, strVal), false},
				{"idiv(float,string)", NewSQLIDivideExpr(floatVal, strVal), false},

				{"mod(int,int)", NewSQLModExpr(intVal, intVal), true},
				{"mod(float,float)", NewSQLModExpr(floatVal, floatVal), true},
				{"mod(int,float)", NewSQLModExpr(intVal, floatVal), true},
				{"mod(bool,bool)", NewSQLModExpr(boolVal, boolVal), false},
				{"mod(int,bool)", NewSQLModExpr(intVal, boolVal), false},
				{"mod(float,bool)", NewSQLModExpr(floatVal, boolVal), false},
				{"mod(string,string)", NewSQLModExpr(strVal, strVal), false},
				{"mod(int,string)", NewSQLModExpr(intVal, strVal), false},
				{"mod(float,string)", NewSQLModExpr(floatVal, strVal), false},

				{"mult(int,int)", NewSQLMultiplyExpr(intVal, intVal), true},
				{"mult(float,float)", NewSQLMultiplyExpr(floatVal, floatVal), true},
				{"mult(int,float)", NewSQLMultiplyExpr(intVal, floatVal), true},
				{"mult(bool,bool)", NewSQLMultiplyExpr(boolVal, boolVal), false},
				{"mult(int,bool)", NewSQLMultiplyExpr(intVal, boolVal), false},
				{"mult(float,bool)", NewSQLMultiplyExpr(floatVal, boolVal), false},
				{"mult(string,string)", NewSQLMultiplyExpr(strVal, strVal), false},
				{"mult(int,string)", NewSQLMultiplyExpr(intVal, strVal), false},
				{"mult(float,string)", NewSQLMultiplyExpr(floatVal, strVal), false},

				{"sub(int,int)", NewSQLSubtractExpr(intVal, intVal), true},
				{"sub(float,float)", NewSQLSubtractExpr(floatVal, floatVal), true},
				{"sub(int,float)", NewSQLSubtractExpr(intVal, floatVal), true},
				{"sub(bool,bool)", NewSQLSubtractExpr(boolVal, boolVal), false},
				{"sub(int,bool)", NewSQLSubtractExpr(intVal, boolVal), false},
				{"sub(float,bool)", NewSQLSubtractExpr(floatVal, boolVal), false},
				{"sub(string,string)", NewSQLSubtractExpr(strVal, strVal), false},
				{"sub(int,string)", NewSQLSubtractExpr(intVal, strVal), false},
				{"sub(float,string)", NewSQLSubtractExpr(floatVal, strVal), false},

				// logical expressions: can evaluate if both arguments are boolean comparable (int, uint, bool), cannot evaluate if at least one is not.
				{"and(bool,bool)", NewSQLAndExpr(boolVal, boolVal), true},
				{"and(int,bool)", NewSQLAndExpr(intVal, boolVal), true},
				{"and(int,int)", NewSQLAndExpr(intVal, intVal), true},
				{"and(int,string)", NewSQLAndExpr(intVal, strVal), false},
				{"and(string,int)", NewSQLAndExpr(strVal, intVal), false},
				{"and(string,string)", NewSQLAndExpr(strVal, strVal), false},

				{"or(bool,bool)", NewSQLOrExpr(boolVal, boolVal), true},
				{"or(int,bool)", NewSQLOrExpr(intVal, boolVal), true},
				{"or(int,int)", NewSQLOrExpr(intVal, intVal), true},
				{"or(int,string)", NewSQLOrExpr(intVal, strVal), false},
				{"or(string,int)", NewSQLOrExpr(strVal, intVal), false},
				{"or(string,string)", NewSQLOrExpr(strVal, strVal), false},

				{"xor(bool,bool)", NewSQLXorExpr(boolVal, boolVal), true},
				{"xor(int,bool)", NewSQLXorExpr(intVal, boolVal), true},
				{"xor(int,int)", NewSQLXorExpr(intVal, intVal), true},
				{"xor(int,string)", NewSQLXorExpr(intVal, strVal), false},
				{"xor(string,int)", NewSQLXorExpr(strVal, intVal), false},
				{"xor(string,string)", NewSQLXorExpr(strVal, strVal), false},

				// comparison expressions: can evaluate if both arguments are similar (numeric types are similar), cannot evaluate if both types are different.
				{"eq(string,string)", NewSQLEqualsExpr(strVal, strVal), true},
				{"eq(int,int)", NewSQLEqualsExpr(intVal, intVal), true},
				{"eq(bool,bool)", NewSQLEqualsExpr(boolVal, boolVal), true},
				{"eq(int,float)", NewSQLEqualsExpr(intVal, floatVal), true},
				{"eq(int,bool)", NewSQLEqualsExpr(intVal, boolVal), false},
				{"eq(int,string)", NewSQLEqualsExpr(intVal, strVal), false},
				{"eq(string,bool)", NewSQLEqualsExpr(strVal, boolVal), false},

				{"gt(string,string)", NewSQLGreaterThanExpr(strVal, strVal), true},
				{"gt(int,int)", NewSQLGreaterThanExpr(intVal, intVal), true},
				{"gt(bool,bool)", NewSQLGreaterThanExpr(boolVal, boolVal), true},
				{"gt(int,float)", NewSQLGreaterThanExpr(intVal, floatVal), true},
				{"gt(int,bool)", NewSQLGreaterThanExpr(intVal, boolVal), false},
				{"gt(int,string)", NewSQLGreaterThanExpr(intVal, strVal), false},
				{"gt(string,bool)", NewSQLGreaterThanExpr(strVal, boolVal), false},

				{"gte(string,string)", NewSQLGreaterThanOrEqualExpr(strVal, strVal), true},
				{"gte(int,int)", NewSQLGreaterThanOrEqualExpr(intVal, intVal), true},
				{"gte(bool,bool)", NewSQLGreaterThanOrEqualExpr(boolVal, boolVal), true},
				{"gte(int,float)", NewSQLGreaterThanOrEqualExpr(intVal, floatVal), true},
				{"gte(int,bool)", NewSQLGreaterThanOrEqualExpr(intVal, boolVal), false},
				{"gte(int,string)", NewSQLGreaterThanOrEqualExpr(intVal, strVal), false},
				{"gte(string,bool)", NewSQLGreaterThanOrEqualExpr(strVal, boolVal), false},

				{"lt(string,string)", NewSQLLessThanExpr(strVal, strVal), true},
				{"lt(int,int)", NewSQLLessThanExpr(intVal, intVal), true},
				{"lt(bool,bool)", NewSQLLessThanExpr(boolVal, boolVal), true},
				{"lt(int,float)", NewSQLLessThanExpr(intVal, floatVal), true},
				{"lt(int,bool)", NewSQLLessThanExpr(intVal, boolVal), false},
				{"lt(int,string)", NewSQLLessThanExpr(intVal, strVal), false},
				{"lt(string,bool)", NewSQLLessThanExpr(strVal, boolVal), false},

				{"lte(string,string)", NewSQLLessThanOrEqualExpr(strVal, strVal), true},
				{"lte(int,int)", NewSQLLessThanOrEqualExpr(intVal, intVal), true},
				{"lte(bool,bool)", NewSQLLessThanOrEqualExpr(boolVal, boolVal), true},
				{"lte(int,float)", NewSQLLessThanOrEqualExpr(intVal, floatVal), true},
				{"lte(int,bool)", NewSQLLessThanOrEqualExpr(intVal, boolVal), false},
				{"lte(int,string)", NewSQLLessThanOrEqualExpr(intVal, strVal), false},
				{"lte(string,bool)", NewSQLLessThanOrEqualExpr(strVal, boolVal), false},

				{"neq(string,string)", NewSQLNotEqualsExpr(strVal, strVal), true},
				{"neq(int,int)", NewSQLNotEqualsExpr(intVal, intVal), true},
				{"neq(bool,bool)", NewSQLNotEqualsExpr(boolVal, boolVal), true},
				{"neq(int,float)", NewSQLNotEqualsExpr(intVal, floatVal), true},
				{"neq(int,bool)", NewSQLNotEqualsExpr(intVal, boolVal), false},
				{"neq(int,string)", NewSQLNotEqualsExpr(intVal, strVal), false},
				{"neq(string,bool)", NewSQLNotEqualsExpr(strVal, boolVal), false},

				{"nse(string,string)", NewSQLNullSafeEqualsExpr(strVal, strVal), true},
				{"nse(int,int)", NewSQLNullSafeEqualsExpr(intVal, intVal), true},
				{"nse(bool,bool)", NewSQLNullSafeEqualsExpr(boolVal, boolVal), true},
				{"nse(int,float)", NewSQLNullSafeEqualsExpr(intVal, floatVal), true},
				{"nse(int,bool)", NewSQLNullSafeEqualsExpr(intVal, boolVal), false},
				{"nse(int,string)", NewSQLNullSafeEqualsExpr(intVal, strVal), false},
				{"nse(string,bool)", NewSQLNullSafeEqualsExpr(strVal, boolVal), false},

				// is expression: right must always be boolean; can evaluate if left is numeric or boolean, cannot evaluate otherwise.
				{"is(bool,bool)", NewSQLIsExpr(boolVal, boolVal), true},
				{"is(int,bool)", NewSQLIsExpr(intVal, boolVal), true},
				{"is(float,bool)", NewSQLIsExpr(floatVal, boolVal), true},
				{"is(string,bool)", NewSQLIsExpr(strVal, boolVal), false},

				// unary ops.
				{"not(int)", NewSQLNotExpr(intVal), true},
				{"not(bool)", NewSQLNotExpr(boolVal), true},
				{"not(float)", NewSQLNotExpr(floatVal), false},
				{"not(string)", NewSQLNotExpr(strVal), false},

				{"unary_minus(int)", NewSQLUnaryMinusExpr(intVal), true},
				{"unary_minus(bool)", NewSQLUnaryMinusExpr(boolVal), false},
				{"unary_minus(float)", NewSQLUnaryMinusExpr(floatVal), true},
				{"unary_minus(string)", NewSQLUnaryMinusExpr(strVal), false},

				{"tilde(int)", NewSQLTildeExpr(intVal), true},
				{"tilde(bool)", NewSQLTildeExpr(boolVal), false},
				{"tilde(float)", NewSQLTildeExpr(floatVal), true},
				{"tilde(string)", NewSQLTildeExpr(strVal), false},

				// agg functions: all of the aggregation functions' reconcile methods are no-ops, so all invocations are valid.
				{"avg", NewSQLAggregationFunctionExpr(parser.AvgAggregateName, false, []SQLExpr{intVal, intVal}), true},
				{"count", NewSQLAggregationFunctionExpr(parser.CountAggregateName, false, []SQLExpr{intVal, intVal}), true},
				{"groupConcat", NewSQLAggregationFunctionExpr(parser.GroupConcatAggregateName, false, []SQLExpr{strVal, strVal}), true},
				{"max", NewSQLAggregationFunctionExpr(parser.MaxAggregateName, false, []SQLExpr{intVal, intVal}), true},
				{"min", NewSQLAggregationFunctionExpr(parser.MinAggregateName, false, []SQLExpr{intVal, intVal}), true},
				{"sum", NewSQLAggregationFunctionExpr(parser.SumAggregateName, false, []SQLExpr{intVal, intVal}), true},
				{"stdDev", NewSQLAggregationFunctionExpr(parser.StdDevAggregateName, false, []SQLExpr{intVal, intVal}), true},
				{"stdDevSample", NewSQLAggregationFunctionExpr(parser.StdDevSampleAggregateName, false, []SQLExpr{intVal, intVal}), true},
			}

			t.Run("Evaluate", func(t *testing.T) {
				runEvaluateErrTests(t, execCfg, execState, errTests)
			})
			t.Run("FoldConstants", func(t *testing.T) {
				runFoldConstantsErrTests(t, execCfg, errTests)
			})
		})

		t.Run("sql_sleep_with_neg_value", func(t *testing.T) {
			req := require.New(t)
			subject, err := NewSQLScalarFunctionExpr(
				"sleep", []SQLExpr{NewSQLValueExpr(NewSQLInt64(knd, -1))})
			req.Nil(err, "unable to create scalar sleep expression")
			_, err = subject.Evaluate(bgCtx, execCfg, execState)
			req.NotNil(err, "did not return error on negative sleep value")
		})

		t.Run("sql_sleep_with_null_value", func(t *testing.T) {
			req := require.New(t)
			subject, err := NewSQLScalarFunctionExpr(
				"sleep",
				[]SQLExpr{NewSQLValueExpr(NewSQLNull(knd))},
			)
			req.Nil(err, "unable to create scalar sleep expression")
			_, err = subject.Evaluate(bgCtx, execCfg, execState)
			req.NotNil(err, "did not return error on null sleep value")
		})

		t.Run("sql_assignment_expr", func(t *testing.T) {
			req := require.New(t)
			e := NewSQLAssignmentExpr(
				NewSQLVariableExpr(
					"test",
					variable.UserKind,
					variable.SessionScope,
					NewSQLNull(MongoSQLValueKind),
				),
				NewSQLAddExpr(
					NewSQLValueExpr(NewSQLFloat(knd, 1)),
					NewSQLValueExpr(NewSQLFloat(knd, 3)),
				),
			)

			result, err := e.Evaluate(bgCtx, execCfg, execState)
			req.Nil(err, "unable to evaluate sql assignment expression")
			req.Equal(
				result,
				NewSQLFloat(knd, 4),
				"expected value of assignment does not match actual value",
			)
		})

		t.Run("sql_divide_by_zero", func(t *testing.T) {
			req := require.New(t)
			subject := NewSQLDivideExpr(
				NewSQLValueExpr(NewSQLInt64(knd, 10)),
				NewSQLValueExpr(NewSQLInt64(knd, 0)),
			)
			result, err := subject.Evaluate(bgCtx, execCfg, execState)
			req.Nil(err, "unable to evaluate sql expression")
			req.True(result.IsNull(), "SQLValue should be null")
		})

		t.Run("sqlcolumnexpr", func(t *testing.T) {
			t.Run("should return the value of the field when it exists", func(t *testing.T) {
				req := require.New(t)
				subject := testSQLColumnExpr(1,
					"test",
					"bar",
					"a",
					EvalInt64,
					schema.MongoInt,
					false,
				)
				result, err := subject.Evaluate(bgCtx, execCfg, execState)
				req.Nil(err, "unable to evalute sql expression")
				req.Equal(
					result,
					NewSQLInt64(knd, 123),
					"actual value of evaluated expression does not match expected value",
				)
			})

			t.Run("should return nil when the field is null", func(t *testing.T) {
				req := require.New(t)
				subject := testSQLColumnExpr(1,
					"test",
					"bar",
					"c",
					EvalInt64,
					schema.MongoInt,
					false,
				)
				result, err := subject.Evaluate(bgCtx, execCfg, execState)
				req.Nil(err, "unable to evalute sql expression")
				req.True(result.IsNull(), "SQLValue should be null")
			})

			t.Run("should panic when the field doesn't exists", func(t *testing.T) {
				req := require.New(t)
				subject := testSQLColumnExpr(1,
					"test",
					"bar",
					"no_existy",
					EvalInt64,
					schema.MongoInt,
					false,
				)
				didPanic := assert.PanicsWithValue(t,
					"cannot find column \"test.bar.no_existy\"",
					func() {
						_, _ = subject.Evaluate(bgCtx, execCfg, execState)
					},
					bgCtx, execCfg, execState)
				req.True(didPanic, "evaluating a field that doesn't exist should panic")
			})
		})

		t.Run("date", func(t *testing.T) {
			dateTime0, _ := time.Parse("2006-01-02", "2014-04-13")
			dateTime1, _ := time.Parse("15:04:05", "11:49:36")
			dateTime2, _ := time.Parse("2006-01-02 15:04:05.999999999", "1997-01-31 09:26:50.124")
			tests = []test{
				{
					"sql_date_parse_comparison_0",
					"DATE '2014-04-13'",
					NewSQLDate(knd, dateTime0),
				},
				{
					"sql_date_parse_comparison_1",
					"{d '2014-04-13'}",
					NewSQLDate(knd, dateTime0),
				},
				{
					"sql_date_parse_comparison_2",
					"TIME '11:49:36'",
					NewSQLTimestamp(knd, dateTime1),
				},
				{
					"sql_date_parse_comparison_3",
					"{t '11:49:36'}",
					NewSQLTimestamp(knd, dateTime1),
				},
				{
					"sql_timestamp_parse_0",
					"TIMESTAMP '1997-01-31 09:26:50.124'",
					NewSQLTimestamp(knd, dateTime2),
				},
				{
					"sql_timestamp_parse_1",
					"{ts '1997-01-31 09:26:50.124'}",
					NewSQLTimestamp(knd, dateTime2),
				},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("adddate", func(t *testing.T) {
			req := require.New(t)

			d, err := time.Parse("2006-01-02", "2003-01-02")
			req.Nil(err, "unable to parse time from string")
			t1, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
			req.Nil(err, "unable to parse time from string")
			t2, err := time.Parse("2006-01-02 15:04:05", "2003-02-03 10:28:06")
			req.Nil(err, "unable to parse time from string")
			t3, err := time.Parse("2006-01-02 15:04:05", "2003-02-14 10:28:06")
			req.Nil(err, "unable to parse time from string")
			d2, err := time.Parse("2006-01-02", "2003-11-30")
			req.Nil(err, "unable to parse time from string")
			d3, err := time.Parse("2006-01-02", "2008-02-02")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_add_date_0", "ADDDATE(NULL, INTERVAL 1 YEAR)", NewSQLNull(knd)},
				{
					"sql_add_date_1",
					"ADDDATE('2002-01-02', INTERVAL 1 YEAR)",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_add_date_2",
					"ADDDATE('2003-08-31', INTERVAL 1 QUARTER)",
					NewSQLTimestamp(knd, d2),
				},
				{
					"sql_add_date_3",
					"ADDDATE('2003-10-31', INTERVAL 1 MONTH)",
					NewSQLTimestamp(knd, d2),
				},
				{
					"sql_add_date_4",
					"ADDDATE('2003-01-01', INTERVAL 1 DAY)",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_add_date_5",
					"ADDDATE('2003-01-02 14:30:09', INTERVAL -2 HOUR)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_6",
					"ADDDATE('2003-01-02 12:23:09', INTERVAL 7 MINUTE)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_7",
					"ADDDATE('2003-01-02 12:30:12', INTERVAL -3 SECOND)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_8",
					"ADDDATE('2003-01-02 12:32:10', INTERVAL '-2:1' MINUTE_SECOND)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_9",
					"ADDDATE('2003-01-02 05:27:06', INTERVAL '7:3:3' HOUR_SECOND)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_10",
					"ADDDATE('2003-01-02 15:32:09', INTERVAL '-3:2' HOUR_MINUTE)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_11",
					"ADDDATE('2002-12-31 10:27:05', INTERVAL '2 2:3:4' DAY_SECOND)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_12",
					"ADDDATE('2002-12-31 10:27:09', INTERVAL '2 2:3' DAY_MINUTE)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_13",
					"ADDDATE('2003-01-01 08:30:09', INTERVAL '1 4' DAY_HOUR)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_14",
					"ADDDATE('2000-09-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_15",
					"ADDDATE('2003-01-02 12:33:09', INTERVAL '-3' HOUR_MINUTE)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_16",
					"ADDDATE('2003-01-02 10:28:06', INTERVAL '2 2:3' DAY_SECOND)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_add_date_17",
					"ADDDATE('2003-01-02 10:28:06', 32)",
					NewSQLTimestamp(knd, t2),
				},
				{
					"sql_add_date_18",
					"ADDDATE('2003-01-02 10:28:06', 43)",
					NewSQLTimestamp(knd, t3),
				},
				{
					"sql_add_date_19",
					"ADDDATE('2003-01-02 10:28:06.000', 43)",
					NewSQLTimestamp(knd, t3),
				},
				{
					"sql_add_date_20",
					"ADDDATE('2003-01-02 10:28:06.000000', 43)",
					NewSQLTimestamp(knd, t3),
				},
				{"sql_add_date_21", "ADDDATE('2008-01-02', 31)", NewSQLTimestamp(knd, d3)},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("convert", func(t *testing.T) {
			req := require.New(t)

			d, err := time.Parse("2006-01-02", "2006-05-11")
			req.Nil(err, "unable to parse time from string")
			t1, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 12:32:12")
			req.Nil(err, "unable to parse time from string")
			dt, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 00:00:00")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_convert_expr_66", "CONVERT('2006-05-11', DATE)", NewSQLDate(knd, d)},
				{"sql_convert_expr_67", "CONVERT(true, DATE)", NewSQLNull(knd)},
				{
					"sql_convert_expr_68",
					"CONVERT(DATE '2006-05-11', DATE)",
					NewSQLDate(knd, d),
				},
				{
					"sql_convert_expr_69",
					"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATE)",
					NewSQLDate(knd, d),
				},
				{"sql_convert_expr_70", "CONVERT(NULL, DATETIME)", NewSQLNull(knd)},
				{"sql_convert_expr_71", "CONVERT(-3.4, DATETIME)", NewSQLNull(knd)},
				{"sql_convert_expr_72", "CONVERT('janna', DATETIME)",
					NewSQLNull(knd)},
				{
					"sql_convert_expr_73",
					"CONVERT('2006-05-11', DATETIME)",
					NewSQLTimestamp(knd, dt),
				},
				{"sql_convert_expr_74", "CONVERT(true, DATETIME)", NewSQLNull(knd)},
				{"sql_convert_expr_75", "CONVERT(3, SQL_TIMESTAMP)", NewSQLNull(knd)},
				{
					"sql_convert_expr_76",
					"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATETIME)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_convert_expr_77",
					"CONVERT(DATE '2006-05-11', SQL_TIMESTAMP)",
					NewSQLTimestamp(knd, dt),
				},
				//{
				//	"sql_convert_expr_78",
				//	"CONVERT('12:32:12', TIME)",
				//	SQLTimestamp{Time: time.Date(0, 1, 1, 12, 32, 12, 0, time.UTC)},
				//},
				//{
				//	"sql_convert_expr_79",
				//	"CONVERT('2006-04-11 12:32:12', TIME)",
				//	SQLTimestamp{Time: time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)},
				//},
				{"sql_convert_expr_80", "CONVERT('0', DATE)",
					NewSQLNull(knd)},
				{"sql_convert_expr_81", "CONVERT(0, DATE)",
					NewSQLDate(knd, NullDate)},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("cot should error when out of range", func(t *testing.T) {
			req := require.New(t)
			subject, err := NewSQLScalarFunctionExpr(
				"cot",
				[]SQLExpr{NewSQLValueExpr(NewSQLFloat(knd, 0))},
			)
			req.Nil(err, "unable to create sql scalar expression")
			_, err = subject.Evaluate(bgCtx, execCfg, execState)
			req.NotNil(err, "did not return nil for out of range cot expression")
		})

		t.Run("utc_date", func(t *testing.T) {
			now := time.Now().In(time.UTC)
			t0 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			tests := []test{
				{"sql_utc_date_0", "UTC_DATE()", NewSQLDate(knd, t0)},
				{"sql_utc_date_1", "UTC_DATE", NewSQLDate(knd, t0)},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("date", func(t *testing.T) {
			req := require.New(t)
			fmtString := "2006-01-02"

			d, err := time.Parse(fmtString, "2016-03-01")
			req.Nil(err, "unable to parse time from string")

			dExpected := NewSQLDate(knd, d)

			preCutoff, err := time.Parse(fmtString, "2069-12-31")
			req.Nil(err, "unable to parse time from string")

			postCutoff, err := time.Parse(fmtString, "1970-01-01")
			req.Nil(err, "unable to parse time from string")

			jan112000, err := time.Parse(fmtString, "2000-01-11")
			req.Nil(err, "unable to parse time from string")

			nov102000, err := time.Parse(fmtString, "2000-11-10")
			req.Nil(err, "unable to parse time from string")

			nov102006, err := time.Parse(fmtString, "2006-11-10")
			req.Nil(err, "unable to parse time from string")

			nov100116, err := time.Parse(fmtString, "0116-11-10")
			req.Nil(err, "unable to parse time from string")

			may042000, err := time.Parse(fmtString, "2000-05-04")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				// invalid inputs
				{"sql_date_invalid_0", "DATE(NULL)", NewSQLNull(knd)},
				{"sql_date_invalid_1", "DATE(23)", NewSQLNull(knd)},
				{"sql_date_invalid_2", "DATE('cat')", NewSQLNull(knd)},
				{"sql_date_invalid_3", "DATE(6911)", NewSQLNull(knd)},
				{"sql_date_invalid_4", "DATE(2017110722040)", NewSQLNull(knd)},
				{"sql_date_invalid_5", "DATE(-50)", NewSQLNull(knd)},
				{"sql_date_invalid_6", "DATE('')", NewSQLNull(knd)},

				// explicitly labeling input as date/timestamp
				{"sql_date_explicit_label_0", "DATE(TIMESTAMP '2016-03-01 12:32:23')", dExpected},
				{"sql_date_explicit_label_1", "DATE(DATE '2016-03-01')", dExpected},

				// unlabeled string inputs
				{"sql_date_unlabeled_0", "DATE('2016-03-01 12:32:23')", dExpected},
				{"sql_date_unlabeled_1", "DATE('2016-03-01')", dExpected},
				{"sql_date_unlabeled_2", "DATE('20160301')", dExpected},

				// number inputs
				{"sql_date_number_inputs_0", "DATE(20160301)", dExpected},
				{"sql_date_number_inputs_1", "DATE(20160301123456)", dExpected},
				{"sql_date_number_inputs_2", "DATE(160301123456)", dExpected},
				{"sql_date_number_inputs_3", "DATE(160301)", dExpected},

				// numbers that are too short to pad
				{"sql_date_non_paddable_0", "DATE(1)", NewSQLNull(knd)},
				{"sql_date_non_paddable_1", "DATE(11)", NewSQLNull(knd)},

				// number inputs requiring padding
				{"sql_date_padded_nums_0", "DATE(111)", NewSQLDate(knd, jan112000)},
				{"sql_date_padded_nums_1", "DATE(1110)", NewSQLDate(knd, nov102000)},
				{"sql_date_padded_nums_2", "DATE(61110)", NewSQLDate(knd, nov102006)},
				{"sql_date_padded_nums_3", "DATE(1161110)", NewSQLDate(knd, nov100116)},
				{"sql_date_padded_nums_4", "DATE(504123025)", NewSQLDate(knd, may042000)},
				{"sql_date_padded_nums_5", "DATE(1110123025)", NewSQLDate(knd, nov102000)},
				{"sql_date_padded_nums_6", "DATE(61110123025)", NewSQLDate(knd, nov102006)},
				{
					"sql_date_padded_nums_7",
					"DATE(61110123025.22)",
					NewSQLDate(knd, nov102006),
				},
				{
					"sql_date_padded_nums_8",
					"DATE(1161110123025)",
					NewSQLDate(knd, nov100116),
				},

				// alternate delimiters
				{"sql_date_alternate_delimiters_0", "DATE('16-03-01')", dExpected},
				{"sql_date_alternate_delimiters_1", "DATE('2016.03.01')", dExpected},

				// mixed delimiters
				{"sql_date_mixed_delimiters_0", "DATE('2016@03.01')", dExpected},
				{"sql_date_mixed_delimiters_1", "DATE('2016-03-01 12.32.23')", dExpected},

				// shortened form of single-digit values
				{"sql_date_shortened_0", "DATE('16-03-1')", dExpected},
				{"sql_date_shortened_1", "DATE('2016.3.1')", dExpected},
				{"sql_date_shortened_2", "DATE('16.3.1')", dExpected},

				// timestamp w/ fractional seconds
				{"sql_date_fraction", "DATE('2016-03-01 12.32.23.3333')", dExpected},

				// use T instead of space to separate
				{"sql_date_char_replace", "DATE('2016-03-01T12.32.23.3333')", dExpected},

				// make sure behavior around year cutoff is correct -
				// 0-69 are intepreted as 2000-2069, while 70-99 are
				// interpreted as 1970-1999.
				{"sql_date_cutoff_0", "DATE('69-12-31')", NewSQLDate(knd, preCutoff)},
				{"sql_date_cutoff_1", "DATE('70-01-01')", NewSQLDate(knd, postCutoff)},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("date_add", func(t *testing.T) {
			req := require.New(t)
			d, err := time.Parse("2006-01-02", "2003-01-02")
			req.Nil(err, "unable to parse time from string")
			t0, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
			req.Nil(err, "unable to parse time from string")
			d2, err := time.Parse("2006-01-02", "2003-11-30")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{
					"sql_date_add_0",
					"DATE_ADD('2002-12-31 10:27:04.500000', INTERVAL '2 2:3:4.5' DAY_MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},

				{
					"sql_date_add_1",
					"DATE_ADD('2003-01-02 10:28:05.500000', INTERVAL '2:2:3.5' DAY_MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_2",
					"DATE_ADD('2003-01-02 10:28:05.500000', INTERVAL '2:2:3.5' HOUR_MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_3",
					"DATE_ADD('2002-12-31 10:27:05', INTERVAL '2 2:3:4' DAY_SECOND)",
					NewSQLTimestamp(knd, t0),
				},

				{
					"sql_date_add_4",
					"DATE_ADD('2003-01-02 12:27:04.500000', INTERVAL '3:4.5' DAY_MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_5",
					"DATE_ADD('2003-01-02 12:27:04.500000', INTERVAL '3:4.5' HOUR_MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_6",
					"DATE_ADD('2003-01-02 12:27:04.500000', INTERVAL '3:4.5' MINUTE_MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_7",
					"DATE_ADD('2003-01-02 10:27:05', INTERVAL '2:3:4' DAY_SECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_8",
					"DATE_ADD('2003-01-02 10:27:05', INTERVAL '2:3:4' HOUR_SECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_9",
					"DATE_ADD('2002-12-31 10:27:09', INTERVAL '2 2:3' DAY_MINUTE)",
					NewSQLTimestamp(knd, t0),
				},

				{
					"sql_date_add_10",
					"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' DAY_MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_11",
					"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' HOUR_MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_12",
					"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' MINUTE_MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_13",
					"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' SECOND_MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_14",
					"DATE_ADD('2003-01-02 12:32:10', INTERVAL '-2:1' DAY_SECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_15",
					"DATE_ADD('2003-01-02 12:32:10', INTERVAL '-2:1' HOUR_SECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_16",
					"DATE_ADD('2003-01-02 12:32:10', INTERVAL '-2:1' MINUTE_SECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_17",
					"DATE_ADD('2003-01-02 15:32:09', INTERVAL '-3:2' DAY_MINUTE)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_18",
					"DATE_ADD('2003-01-02 15:32:09', INTERVAL '-3:2' HOUR_MINUTE)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_19",
					"DATE_ADD('2002-12-31 10:30:09', INTERVAL '2 2' DAY_HOUR)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_20",
					"DATE_ADD('2000-09-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)",
					NewSQLTimestamp(knd, t0),
				},

				{
					"sql_date_add_21",
					"DATE_ADD('2002-01-02', INTERVAL NULL YEAR)",
					NewSQLNull(knd),
				},
				{
					"sql_date_add_22",
					"DATE_ADD(NULL, INTERVAL 1 YEAR)",
					NewSQLNull(knd),
				},
				{
					"sql_date_add_23",
					"DATE_ADD('2002-01-02', INTERVAL 1 YEAR)",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_date_add_24",
					"DATE_ADD('2003-08-31', INTERVAL 1 QUARTER)",
					NewSQLTimestamp(knd, d2),
				},
				{
					"sql_date_add_25",
					"DATE_ADD('2003-10-31', INTERVAL 1 MONTH)",
					NewSQLTimestamp(knd, d2),
				},
				{
					"sql_date_add_26",
					"DATE_ADD('2003-01-01', INTERVAL 1 DAY)",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_date_add_27",
					"DATE_ADD('2003-01-02 14:30:09', INTERVAL -2 HOUR)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_28",
					"DATE_ADD('2003-01-02 12:23:09', INTERVAL 7 MINUTE)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_29",
					"DATE_ADD('2003-01-02 12:30:12', INTERVAL -3 SECOND)",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_date_add_30",
					"DATE_ADD('2003-01-02 12:30:08.999999', INTERVAL 1 MICROSECOND)",
					NewSQLTimestamp(knd, t0),
				},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("date_sub, subdate", func(t *testing.T) {
			req := require.New(t)
			d, err := time.Parse("2006-01-02", "2003-01-02")
			req.Nil(err, "unable to parse time from string")
			t1, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
			req.Nil(err, "unable to parse time from string")
			t2, err := time.Parse("2006-01-02 15:04:05", "2007-12-02 12:00:00")
			req.Nil(err, "unable to parse time from string")
			d2, err := time.Parse("2006-01-02", "2003-11-30")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{
					"sql_date_sub_0",
					"DATE_SUB('2004-01-02', INTERVAL NULL YEAR)",
					NewSQLNull(knd),
				},
				{"sql_date_sub_1", "DATE_SUB(NULL, INTERVAL 1 YEAR)", NewSQLNull(knd)},
				{
					"sql_date_sub_2",
					"DATE_SUB('2004-01-02', INTERVAL 1 YEAR)",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_date_sub_3",
					"DATE_SUB('2003-04-02', INTERVAL 1 QUARTER)",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_date_sub_4",
					"DATE_SUB('2003-12-31', INTERVAL 1 MONTH)",
					NewSQLTimestamp(knd, d2),
				},
				{
					"sql_date_sub_5",
					"DATE_SUB('2003-01-03', INTERVAL 1 DAY)",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_date_sub_6",
					"SUBDATE('2004-01-02', INTERVAL 1 YEAR)",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_date_sub_7",
					"SUBDATE('2003-04-02', INTERVAL 1 QUARTER)",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_date_sub_8",
					"SUBDATE('2003-12-31', INTERVAL 1 MONTH)",
					NewSQLTimestamp(knd, d2),
				},
				{
					"sql_date_sub_9",
					"SUBDATE('2008-01-02 12:00:00', 31)",
					NewSQLTimestamp(knd, t2),
				},
				{
					"sql_date_sub_10",
					"SUBDATE('2016-01-02 12:00:00', 2953)",
					NewSQLTimestamp(knd, t2),
				},
				{
					"sql_date_sub_11",
					"DATE_SUB('2003-01-02 10:30:09', INTERVAL -2 HOUR)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_12",
					"DATE_SUB('2003-01-02 12:37:09', INTERVAL 7 MINUTE)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_13",
					"DATE_SUB('2003-01-02 12:30:12', INTERVAL 3 SECOND)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_14",
					"DATE_SUB('2003-01-02 12:32:10', INTERVAL '2:1' MINUTE_SECOND)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_15",
					"DATE_SUB('2003-01-02 19:33:12', INTERVAL '7:3:3' HOUR_SECOND)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_16",
					"DATE_SUB('2003-01-02 15:32:09', INTERVAL '3:2' HOUR_MINUTE)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_17",
					"DATE_SUB('2003-01-04 14:33:13', INTERVAL '2 2:3:4' DAY_SECOND)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_18",
					"DATE_SUB('2003-01-04 14:33:09', INTERVAL '2 2:3' DAY_MINUTE)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_19",
					"DATE_SUB('2003-01-03 16:30:09', INTERVAL '1 4' DAY_HOUR)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_20",
					"DATE_SUB('2005-05-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_21",
					"DATE_SUB('2003-01-02 12:33:09', INTERVAL '3' HOUR_MINUTE)",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_date_sub_22",
					"DATE_SUB('2003-01-02 14:32:12', INTERVAL '2 2:3' DAY_SECOND)",
					NewSQLTimestamp(knd, t1),
				},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("from_days", func(t *testing.T) {

			t1 := time.Date(0001, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
			t2 := time.Date(2000, 7, 3, 0, 0, 0, 0, schema.DefaultLocale)
			t3 := time.Date(10000, 3, 15, 0, 0, 0, 0, schema.DefaultLocale)
			t4 := time.Date(0005, 6, 29, 0, 0, 0, 0, schema.DefaultLocale)
			t5 := time.Date(2112, 1, 8, 0, 0, 0, 0, schema.DefaultLocale)

			tests := []test{
				{"sql_from_days_0", "FROM_DAYS(NULL)", NewSQLNull(knd)},
				{"sql_from_days_1", "FROM_DAYS('sdg')", NewSQLNull(knd)},
				{"sql_from_days_2", "FROM_DAYS(1.23)", NewSQLNull(knd)},
				{"sql_from_days_3", "FROM_DAYS(-1.23)", NewSQLNull(knd)},
				{"sql_from_days_4", "FROM_DAYS(-223.33)", NewSQLNull(knd)},
				{"sql_from_days_5", "FROM_DAYS(223.33)", NewSQLNull(knd)},
				{"sql_from_days_6", "FROM_DAYS(365.33)", NewSQLNull(knd)},
				{"sql_from_days_7", "FROM_DAYS(3652499.5)", NewSQLNull(knd)},
				{"sql_from_days_8", "FROM_DAYS(-771399.216)", NewSQLNull(knd)},
				{"sql_from_days_9", "FROM_DAYS(365.93)", NewSQLDate(knd, t1)},
				{"sql_from_days_10", "FROM_DAYS(343+23)", NewSQLDate(knd, t1)},
				{"sql_from_days_11", "FROM_DAYS(730669)", NewSQLDate(knd, t2)},
				{"sql_from_days_12", "FROM_DAYS(3652499.3)", NewSQLDate(knd, t3)},
				{"sql_from_days_13", "FROM_DAYS('2006-05-11')", NewSQLDate(knd, t4)},
				{"sql_from_days_14", "FROM_DAYS(771399.216)", NewSQLDate(knd, t5)},
			}

			runTests(t, execCfg, execState, tests)
		})

		t.Run("greatest", func(t *testing.T) {
			req := require.New(t)
			d, err := time.Parse("2006-01-02", "2006-05-11")
			req.Nil(err, "unable to parse time from string")
			t0, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 12:32:23")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_greatest_expr_0", "GREATEST(NULL, 1, 2)", NewSQLNull(knd)},
				{"sql_greatest_expr_1", "GREATEST(1,3,2)", NewSQLInt64(knd, 3)},
				{
					"sql_greatest_expr_2",
					"GREATEST(2,2.3)",
					NewSQLDecimal128(knd, decimal.New(23, -1)),
				},
				{
					"sql_greatest_expr_3",
					"GREATEST('cats', '4', '2')",
					NewSQLVarchar(knd, "cats"),
				},
				{
					"sql_greatest_expr_4",
					"GREATEST('dog', 'cats', 'bird')",
					NewSQLVarchar(knd, "dog"),
				},
				{
					"sql_greatest_expr_5",
					"GREATEST('cat', 'bird', 2)",
					NewSQLInt64(knd, 2),
				},
				{
					"sql_greatest_expr_6",
					"GREATEST('cat', 2.2)",
					NewSQLDecimal128(knd, decimal.New(22, -1)),
				},
				{
					"sql_greatest_expr_7",
					"GREATEST(false, true)",
					NewSQLBool(knd, true),
				},
				{
					"sql_greatest_expr_8",
					"GREATEST(DATE '2005-05-11', DATE '2006-05-11', DATE '2000-05-11')",
					NewSQLDate(knd, d),
				},
				{
					"sql_greatest_expr_9",
					"GREATEST(DATE '2006-05-11', 14, 4235)",
					NewSQLInt64(knd, 20060511),
				},
				{
					"sql_greatest_expr_10",
					"GREATEST(DATE '2006-05-11', 14, 20080622)",
					NewSQLInt64(knd, 20080622),
				},
				{
					"sql_greatest_expr_11",
					"GREATEST(DATE '2006-05-11', 14, 20080622.1)",
					NewSQLDecimal128(knd, decimal.New(200806221, -1)),
				},
				{
					"sql_greatest_expr_12",
					"GREATEST(DATE '2006-05-11', 14, 4235.2)",
					NewSQLDecimal128(knd, decimal.New(20060511, 0)),
				},
				{
					"sql_greatest_expr_13",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', DATE '2006-05-11')",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_greatest_expr_14",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 12, 345)",
					NewSQLInt64(knd, 20060511123223),
				},
				{
					"sql_greatest_expr_15",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 20080923124345.3)",
					NewSQLDecimal128(knd, decimal.New(200809231243453, -1)),
				},
				{
					"sql_greatest_expr_16",
					"GREATEST(DATE '2006-05-11', 'cat', '2007-04-11')",
					NewSQLVarchar(knd, "2007-04-11"),
				},
				{
					"sql_greatest_expr_17",
					"GREATEST(DATE '2006-05-11', 20080912, '2007-04-11')",
					NewSQLInt64(knd, 20080912),
				},
				{
					"sql_greatest_expr_18",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', TIMESTAMP '2006-05-11 10:32:45')",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_greatest_expr_19",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 20080913, DATE '2007-08-23')",
					NewSQLInt64(knd, 20060511123223),
				},
				{
					"sql_greatest_expr_20",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', '2008-09-13')",
					NewSQLVarchar(knd, "2008-09-13"),
				},
				{
					"sql_greatest_expr_21",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', '2005-09-13')",
					NewSQLTimestamp(knd, t0),
				},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("last_day", func(t *testing.T) {
			req := require.New(t)
			d1, err := time.Parse("2006-01-02", "2003-02-28")
			req.Nil(err, "unable to parse time from string")
			d2, err := time.Parse("2006-01-02", "2004-02-29")
			req.Nil(err, "unable to parse time from string")
			d3, err := time.Parse("2006-01-02", "2004-01-31")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_last_day_0", "LAST_DAY('')", NewSQLNull(knd)},
				{"sql_last_day_1", "LAST_DAY(NULL)", NewSQLNull(knd)},
				{"sql_last_day_2", "LAST_DAY('2003-03-32')", NewSQLNull(knd)},
				{"sql_last_day_3", "LAST_DAY('2003-02-05')", NewSQLDate(knd, d1)},
				{"sql_last_day_4", "LAST_DAY('2004-02-05')", NewSQLDate(knd, d2)},
				{"sql_last_day_5", "LAST_DAY('2004-01-01 01:01:01')", NewSQLDate(knd, d3)},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("least", func(t *testing.T) {
			req := require.New(t)
			d, err := time.Parse("2006-01-02", "2005-05-11")
			req.Nil(err, "unable to parse time from string")
			t0, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 00:00:00")
			req.Nil(err, "unable to parse time from string")
			t1, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 10:32:23")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_least_expr_0", "LEAST(NULL, 1, 2)", NewSQLNull(knd)},
				{"sql_least_expr_1", "LEAST(1,3,2)", NewSQLInt64(knd, 1)},
				{"sql_least_expr_2", "LEAST(2,2.3)", NewSQLDecimal128(knd, decimal.New(2, 0))},
				{"sql_least_expr_3", "LEAST('cats', '4', '2')", NewSQLVarchar(knd, "2")},
				{"sql_least_expr_4", "LEAST('dog', 'cats', 'bird')", NewSQLVarchar(knd, "bird")},
				{"sql_least_expr_5", "LEAST(false, true)", NewSQLBool(knd, false)},
				{
					"sql_least_expr_6",
					"LEAST(DATE '2005-05-11', DATE '2006-05-11', DATE '2007-05-11')",
					NewSQLDate(knd, d),
				},
				{
					"sql_least_expr_7",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', DATE '2006-05-11')",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_least_expr_8",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', TIMESTAMP '2006-05-11 10:32:23')",
					NewSQLTimestamp(knd, t1),
				},
				{"sql_least_expr_9", "LEAST('cat', 'bird', 2)", NewSQLInt64(knd, 0)},
				{"sql_least_expr_10", "LEAST('cat', 2.2)",
					NewSQLDecimal128(knd, decimal.New(0, 0))},
				{"sql_least_expr_11", "LEAST(DATE '2006-05-11', 14, 4235)", NewSQLInt64(knd, 14)},
				{
					"sql_least_expr_12",
					"LEAST(DATE '2006-05-11', 14, 20080622.1)",
					NewSQLDecimal128(knd, decimal.New(14, 0)),
				},
				{
					"sql_least_expr_13",
					"LEAST(DATE '2006-05-11', 14, 4235.2)",
					NewSQLDecimal128(knd, decimal.New(14, 0)),
				},
				{
					"sql_least_expr_14",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', 12, 345)",
					NewSQLInt64(knd, 12),
				},
				{
					"sql_least_expr_15",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', 20080923124345.3)",
					NewSQLDecimal128(knd, decimal.New(20060511123223, 0)),
				},
				{
					"sql_least_expr_16",
					"LEAST(DATE '2006-05-11', 'cat', '2007-04-11')",
					NewSQLVarchar(knd, "cat"),
				},
				{
					"sql_least_expr_17",
					"LEAST(DATE '2006-05-11', 20080912, '2007-04-11')",
					NewSQLInt64(knd, 2007),
				},
				{
					"sql_least_expr_18",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', 20080913, DATE '2007-08-23')",
					NewSQLInt64(knd, 20070823),
				},
				{
					"sql_least_expr_19",
					"LEAST(TIMESTAMP '2006-05-11 10:32:23', '2008-09-13')",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_least_expr_20",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', '2005-09-13')",
					NewSQLVarchar(knd, "2005-09-13"),
				},
			}
			runTests(t, execCfg, execState, tests)

		})

		t.Run("makedate", func(t *testing.T) {
			req := require.New(t)
			d, err := time.Parse("2006-01-02", "2000-02-01")
			req.Nil(err, "unable to parse time from string")
			d1, err := time.Parse("2006-01-02", "2012-02-01")
			req.Nil(err, "unable to parse time from string")
			d2, err := time.Parse("2006-01-02", "1977-03-07")
			req.Nil(err, "unable to parse time from string")
			d3, err := time.Parse("2006-01-02", "0100-02-01")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_makedate_0", "MAKEDATE(NULL, 4)", NewSQLNull(knd)},
				{"sql_makedate_1", "MAKEDATE(2004, 0)", NewSQLNull(knd)},
				{"sql_makedate_2", "MAKEDATE(9999, 370)", NewSQLNull(knd)},
				{"sql_makedate_3", "MAKEDATE('sdg', 32)", NewSQLDate(knd, d)},
				{"sql_makedate_4", "MAKEDATE('2000.9', 32)", NewSQLDate(knd, d)},
				{"sql_makedate_5", "MAKEDATE(1999.5, 32)", NewSQLDate(knd, d)},
				{"sql_makedate_6", "MAKEDATE('2000.9', '32.9')", NewSQLDate(knd, d)},
				{"sql_makedate_7", "MAKEDATE(1999.5, 31.5)", NewSQLDate(knd, d)},
				{"sql_makedate_8", "MAKEDATE(2000, 32)", NewSQLDate(knd, d)},
				{"sql_makedate_9", "MAKEDATE(12, 32)", NewSQLDate(knd, d1)},
				{"sql_makedate_10", "MAKEDATE(77, 66)", NewSQLDate(knd, d2)},
				{"sql_makedate_11", "MAKEDATE(99.5, 31.5)", NewSQLDate(knd, d3)},
				{"sql_makedate_12", "MAKEDATE('100.9', '32.5')", NewSQLDate(knd, d3)},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("str_to_date", func(t *testing.T) {
			req := require.New(t)
			d, err := time.Parse("2006-01-02", "2016-04-03")
			req.Nil(err, "unable to parse time from string")
			t0, err := time.Parse("2006-01-02 15:04:05", "2016-04-03 12:22:22")
			req.Nil(err, "unable to parse time from string")
			t1, err := time.Parse("2006-01-02 15:04:05", "2005-04-02 00:12:00")
			req.Nil(err, "unable to parse time from string")
			t2, err := time.Parse("2006-01-02 15:04:05", "2016-04-03 12:22:00")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_str_to_date_0", "STR_TO_DATE(NULL, 4)", NewSQLNull(knd)},
				{"sql_str_to_date_1", "STR_TO_DATE('foobarbar', NULL)", NewSQLNull(knd)},
				{
					"sql_str_to_date_2",
					"STR_TO_DATE('2016-04-03','%Y-%m-%d')",
					NewSQLDate(knd, d),
				},
				{
					"sql_str_to_date_3",
					"STR_TO_DATE('04,03,2016', '%m,%d,%Y')",
					NewSQLDate(knd, d),
				},
				{
					"sql_str_to_date_4",
					"STR_TO_DATE('04,03,a16', '%m,%d,a%y')",
					NewSQLDate(knd, d),
				},
				{
					"sql_str_to_date_5",
					"STR_TO_DATE('2016-04-03 12:22:22', '%Y-%m-%d %H:%i:%s')",
					NewSQLTimestamp(knd, t0),
				},
				{
					"sql_str_to_date_6",
					"STR_TO_DATE('2016-04-03 12:22', '%Y-%m-%d %H:%i')",
					NewSQLTimestamp(knd, t2),
				},
				{
					"sql_str_to_date_7",
					"STR_TO_DATE('2005-04-02 12', '%Y-%m-%d %i')",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_str_to_date_8",
					"STR_TO_DATE('Apr 03, 2016', '%b %d, %Y')",
					NewSQLDate(knd, d),
				},
				{
					"sql_str_to_date_9",
					"STR_TO_DATE('Tue 2016-04-03', '%a %Y-%m-%d')",
					NewSQLDate(knd, d),
				},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("timestamp", func(t *testing.T) {
			req := require.New(t)
			t1, err := time.Parse("2006-01-02 15:04:05.000000", "2010-01-01 22:35:10.523236")
			req.Nil(err, "unable to parse time from stringstamp")
			t2, err := time.Parse("2006-01-02 15:04:05.000000", "2010-01-01 23:33:11.400000")
			req.Nil(err, "unable to parse time from stringstamp")
			t3, err := time.Parse("2006-01-02 15:04:05", "2004-01-01 00:00:00")
			req.Nil(err, "unable to parse time from stringstamp")
			t4, err := time.Parse("2006-01-02 15:04:05", "2003-12-31 00:00:00")
			req.Nil(err, "unable to parse time from stringstamp")
			t5, err := time.Parse("2006-01-02 15:04:05.000000", "2003-12-31 12:00:12.300000")
			req.Nil(err, "unable to parse time from stringstamp")
			t6, err := time.Parse("2006-01-02 15:04:05", "2003-12-31 12:23:23")
			req.Nil(err, "unable to parse time from stringstamp")
			t7, err := time.Parse("2006-01-02 15:04:05", "2010-01-01 12:33:23")
			req.Nil(err, "unable to parse time from stringstamp")

			tests := []test{
				{"sql_timestamp_0", "TIMESTAMP(NULL)", NewSQLNull(knd)},
				{"sql_timestamp_1", "TIMESTAMP(NULL, NULL)", NewSQLNull(knd)},
				{"sql_timestamp_2", "TIMESTAMP(NULL, '12:22:22')", NewSQLNull(knd)},
				{"sql_timestamp_3", "TIMESTAMP('2002-01-02', NULL)", NewSQLNull(knd)},
				{
					"sql_timestamp_4",
					"TIMESTAMP('2010-01-01 11:11:11', '11:71:11')",
					NewSQLNull(knd),
				},
				{
					"sql_timestamp_5",
					"TIMESTAMP('2010-01-01 11:11:11', '11:23:59.5232355')",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_timestamp_6",
					"TIMESTAMP('2010-01-01 11:11:11', '12:22.4:12')",
					NewSQLTimestamp(knd, t2),
				},
				{
					"sql_timestamp_7",
					"TIMESTAMP('2003-12-31 12:00:00', '12:00:00')",
					NewSQLTimestamp(knd, t3),
				},
				{"sql_timestamp_8", "TIMESTAMP(20031231)", NewSQLTimestamp(knd, t4)},
				{"sql_timestamp_9", "TIMESTAMP('2003-12-31')", NewSQLTimestamp(knd, t4)},
				{
					"sql_timestamp_10",
					"TIMESTAMP('2003-12-31 12:00:00', '12.3:10:30')",
					NewSQLTimestamp(knd, t5),
				},
				{
					"sql_timestamp_11",
					"TIMESTAMP('2003-12-31 12:23:23')",
					NewSQLTimestamp(knd, t6),
				},
				{
					"sql_timestamp_12",
					"TIMESTAMP('2010-01-01 11:11:11', '12212')",
					NewSQLTimestamp(knd, t7),
				},
				{
					"sql_timestamp_13",
					"TIMESTAMP('2010-01-01 11:11:11', 12212)",
					NewSQLTimestamp(knd, t7),
				},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("timestampadd", func(t *testing.T) {
			req := require.New(t)
			d, err := time.Parse("2006-01-02", "2003-01-02")
			req.Nil(err, "unable to parse time from stringstamp")
			t1, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
			req.Nil(err, "unable to parse time from stringstamp")
			dt, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 01:00:00")
			req.Nil(err, "unable to parse time from stringstamp")
			t2 := t1.Add(time.Duration(15000) * time.Microsecond)

			tests := []test{
				{
					"sql_timestampadd_0",
					"TIMESTAMPADD(YEAR, 1, DATE '2002-01-02')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_1",
					"TIMESTAMPADD(YEAR, 0.5, DATE '2002-01-02')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_2",
					"TIMESTAMPADD(QUARTER, 1, DATE '2002-10-02')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_3",
					"TIMESTAMPADD(QUARTER, 0.5, DATE '2002-10-02')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_4",
					"TIMESTAMPADD(MONTH, 1, DATE '2002-12-02')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_5",
					"TIMESTAMPADD(MONTH, 0.5, DATE '2002-12-02')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_6",
					"TIMESTAMPADD(WEEK, 1, DATE '2002-12-26')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_7",
					"TIMESTAMPADD(WEEK, 0.5, DATE '2002-12-26')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_8",
					"TIMESTAMPADD(DAY, 1, DATE '2003-01-01')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_9",
					"TIMESTAMPADD(DAY, 0.5, DATE '2003-01-01')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_10",
					"TIMESTAMPADD(HOUR, 1, DATE '2003-01-02')",
					NewSQLTimestamp(knd, dt),
				},
				{
					"sql_timestampadd_11",
					"TIMESTAMPADD(HOUR, 0.5, DATE '2003-01-02')",
					NewSQLTimestamp(knd, dt),
				},
				{
					"sql_timestampadd_12",
					"TIMESTAMPADD(MINUTE, 60, DATE '2003-01-02')",
					NewSQLTimestamp(knd, dt),
				},
				{
					"sql_timestampadd_13",
					"TIMESTAMPADD(MINUTE, 59.5, DATE '2003-01-02')",
					NewSQLTimestamp(knd, dt),
				},
				{
					"sql_timestampadd_14",
					"TIMESTAMPADD(SECOND, 3600, DATE '2003-01-02')",
					NewSQLTimestamp(knd, dt),
				},
				// No round test for SECOND, SECOND is not rounded.
				{
					"sql_timestampadd_15",
					"TIMESTAMPADD(MICROSECOND, 15000, TIMESTAMP '2003-01-02 12:30:09')",
					NewSQLTimestamp(knd, t2),
				},
				{
					"sql_timestampadd_16",
					"TIMESTAMPADD(DAY, 1, TIMESTAMP '2003-01-01 12:30:09')",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_timestampadd_17",
					"TIMESTAMPADD(WEEK, 2, TIMESTAMP '2002-12-19 12:30:09')",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_timestampadd_18",
					"TIMESTAMPADD(SQL_TSI_YEAR, 2, TIMESTAMP '2001-01-02 12:30:09')",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_timestampadd_19",
					"TIMESTAMPADD(SQL_TSI_QUARTER, 2, DATE '2002-07-02')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_20",
					"TIMESTAMPADD(SQL_TSI_MONTH, 1, TIMESTAMP '2002-12-02 12:30:09')",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_timestampadd_21",
					"TIMESTAMPADD(SQL_TSI_WEEK, 1, DATE '2002-12-26')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_22",
					"TIMESTAMPADD(SQL_TSI_DAY, 1, DATE '2003-01-01')",
					NewSQLTimestamp(knd, d),
				},
				{
					"sql_timestampadd_23",
					"TIMESTAMPADD(SQL_TSI_HOUR, 1, TIMESTAMP '2003-01-02 11:30:09')",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_timestampadd_24",
					"TIMESTAMPADD(SQL_TSI_MINUTE, 1, TIMESTAMP '2003-01-02 12:29:09')",
					NewSQLTimestamp(knd, t1),
				},
				{
					"sql_timestampadd_25",
					"TIMESTAMPADD(SQL_TSI_SECOND, 1, TIMESTAMP '2003-01-02 12:30:08')",
					NewSQLTimestamp(knd, t1),
				},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("year", func(t *testing.T) {
			t.Skip()
			tests := []test{
				{"sql_year_0", "YEAR(NULL)", NewSQLNull(knd)},
				{"sql_year_1", "YEAR('sdg')", NewSQLNull(knd)},
				{"sql_year_2", "YEAR('2016-1-01 10:23:52')", NewSQLInt64(knd, 53)},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("yearweek", func(t *testing.T) {
			t.Skip()
			tests := []test{
				{"sql_yearweek_0", "YEARWEEK(NULL)", NewSQLNull(knd)},
				{"sql_yearweek_1", "YEARWEEK('sdg')", NewSQLNull(knd)},
				{"sql_yearweek_2", "YEARWEEK('2000-01-01')", NewSQLInt64(knd, 199252)},
				{"sql_yearweek_3", "YEARWEEK('2001-01-01')", NewSQLInt64(knd, 200053)},
				{"sql_yearweek_4", "YEARWEEK('2002-01-01')", NewSQLInt64(knd, 200152)},
				{"sql_yearweek_5", "YEARWEEK('2003-01-01')", NewSQLInt64(knd, 200252)},
				{"sql_yearweek_6", "YEARWEEK('2004-01-01')", NewSQLInt64(knd, 200352)},
				{"sql_yearweek_7", "YEARWEEK('2005-01-01')", NewSQLInt64(knd, 200452)},
				{"sql_yearweek_8", "YEARWEEK('2006-01-01')", NewSQLInt64(knd, 200601)},
				{"sql_yearweek_9", "YEARWEEK('2000-01-06')", NewSQLInt64(knd, 200001)},
				{"sql_yearweek_10", "YEARWEEK('2001-01-06')", NewSQLInt64(knd, 200053)},
				{"sql_yearweek_11", "YEARWEEK('2002-01-06')", NewSQLInt64(knd, 200201)},
				{"sql_yearweek_12", "YEARWEEK('2003-01-06')", NewSQLInt64(knd, 200301)},
				{"sql_yearweek_13", "YEARWEEK('2004-01-06')", NewSQLInt64(knd, 200401)},
				{"sql_yearweek_14", "YEARWEEK('2005-01-06')", NewSQLInt64(knd, 200501)},
				{"sql_yearweek_15", "YEARWEEK('2006-01-06')", NewSQLInt64(knd, 200601)},
				{"sql_yearweek_16", "YEARWEEK('2000-01-01',1)", NewSQLInt64(knd, 199252)},
				{"sql_yearweek_17", "YEARWEEK('2001-01-01',1)", NewSQLInt64(knd, 200101)},
				{"sql_yearweek_18", "YEARWEEK('2002-01-01',1)", NewSQLInt64(knd, 200201)},
				{"sql_yearweek_19", "YEARWEEK('2003-01-01',1)", NewSQLInt64(knd, 200301)},
				{"sql_yearweek_20", "YEARWEEK('2004-01-01',1)", NewSQLInt64(knd, 200401)},
				{"sql_yearweek_21", "YEARWEEK('2005-01-01',1)", NewSQLInt64(knd, 200453)},
				{"sql_yearweek_22", "YEARWEEK('2006-01-01',1)", NewSQLInt64(knd, 200552)},
				{"sql_yearweek_23", "YEARWEEK('2000-01-06',1)", NewSQLInt64(knd, 200001)},
				{"sql_yearweek_24", "YEARWEEK('2001-01-06',1)", NewSQLInt64(knd, 200101)},
				{"sql_yearweek_25", "YEARWEEK('2002-01-06',1)", NewSQLInt64(knd, 200201)},
				{"sql_yearweek_26", "YEARWEEK('2003-01-06',1)", NewSQLInt64(knd, 200301)},
				{"sql_yearweek_27", "YEARWEEK('2004-01-06',1)", NewSQLInt64(knd, 200402)},
				{"sql_yearweek_28", "YEARWEEK('2005-01-06',1)", NewSQLInt64(knd, 200501)},
				{"sql_yearweek_29", "YEARWEEK('2006-01-06',1)", NewSQLInt64(knd, 200601)},
			}
			runTests(t, execCfg, execState, tests)
		})

		t.Run("sqlunarytildeexpr", func(t *testing.T) {
			t.Skip()
			//TODO: I'm not convinced we have this correct.
		})
	})
}

func TestSQLLikeExprConvertToPattern(t *testing.T) {
	test := func(syntax, expected string) {
		name := syntax
		t.Run(name, func(t *testing.T) {
			req := require.New(t)
			pattern := ConvertSQLValueToPattern(NewSQLVarchar(knd, syntax), '\\')
			req.Equal(pattern, expected)
		})
	}

	test("David", "^David$")
	test("Da\\vid", "^David$")
	test("Da\\\\vid", "^Da\\\\vid$")
	test("Da_id", "^Da.id$")
	test("Da\\_id", "^Da_id$")
	test("Da%d", "^Da.*d$")
	test("Da\\%d", "^Da%d$")
	test("Sto_. %ow", "^Sto.\\. .*ow$")
}

func TestCompareTo(t *testing.T) {

	var (
		diff        = 969 * time.Hour
		sameDayDiff = time.Duration(1)
		now         = time.Now()
	)

	type test struct {
		left     SQLValue
		right    SQLValue
		expected int
	}

	runTests := func(t *testing.T, tests []test) {
		for idx, tst := range tests {
			name := fmt.Sprintf("%d", idx)
			t.Run(name, func(t *testing.T) {
				req := require.New(t)
				compareTo, err := CompareTo(tst.left, tst.right, collation.Default)
				req.NoError(err)
				req.Equal(tst.expected, compareTo)
			})
		}
	}

	t.Run("SQLInt", func(t *testing.T) {
		tests := []test{
			{NewSQLInt64(knd, 1), NewSQLInt64(knd, 0), 1},
			{NewSQLInt64(knd, 1), NewSQLInt64(knd, 1), 0},
			{NewSQLInt64(knd, 1), NewSQLInt64(knd, 2), -1},
			{NewSQLInt64(knd, 1), NewSQLUint64(knd, 1), 0},
			{NewSQLInt64(knd, 1), NewSQLFloat(knd, 1), 0},
			{NewSQLInt64(knd, 1), NewSQLBool(knd, false), 1},
			{NewSQLInt64(knd, 1), NewSQLBool(knd, true), 0},
			{NewSQLInt64(knd, 1), NewSQLNull(knd), 1},
			{NewSQLInt64(knd, 1), NewSQLVarchar(knd, "56e0750e1d857aea925a4ba1"), -1},
			{NewSQLInt64(knd, 1), NewSQLVarchar(knd, "bac"), 1},
			{NewSQLInt64(knd, 1), NewSQLDate(knd, now), -1},
			{NewSQLInt64(knd, 1), NewSQLTimestamp(knd, now), -1},
		}
		runTests(t, tests)
	})

	t.Run("SQLFloat", func(t *testing.T) {
		tests := []test{
			{NewSQLFloat(knd, 0.1), NewSQLInt64(knd, 0), 1},
			{NewSQLFloat(knd, 1.1), NewSQLInt64(knd, 1), 1},
			{NewSQLFloat(knd, 0.1), NewSQLInt64(knd, 2), -1},
			{NewSQLFloat(knd, 1.1), NewSQLUint64(knd, 1), 1},
			{NewSQLFloat(knd, 1.1), NewSQLFloat(knd, 1), 1},
			{NewSQLFloat(knd, 0.1), NewSQLBool(knd, false), 1},
			{NewSQLFloat(knd, 0.1), NewSQLBool(knd, true), -1},
			{NewSQLFloat(knd, 0.1), NewSQLNull(knd), 1},
			{NewSQLFloat(knd, 0.1), NewSQLVarchar(knd, "56e0750e1d857aea925a4ba1"), -1},
			{NewSQLFloat(knd, 0.1), NewSQLVarchar(knd, "bac"), 1},
			{NewSQLFloat(knd, 0.0), NewSQLInt64(knd, 1), -1},
			{NewSQLFloat(knd, 0.1), NewSQLDate(knd, now), -1},
			{NewSQLFloat(knd, 0.1), NewSQLTimestamp(knd, now), -1},
		}
		runTests(t, tests)
	})

	t.Run("SQLBool", func(t *testing.T) {
		tests := []test{
			{NewSQLBool(knd, true), NewSQLInt64(knd, 0), 1},
			{NewSQLBool(knd, true), NewSQLInt64(knd, 1), 0},
			{NewSQLBool(knd, true), NewSQLInt64(knd, 2), -1},
			{NewSQLBool(knd, true), NewSQLUint64(knd, 1), 0},
			{NewSQLBool(knd, true), NewSQLFloat(knd, 1), 0},
			{NewSQLBool(knd, true), NewSQLBool(knd, false), 1},
			{NewSQLBool(knd, true), NewSQLBool(knd, true), 0},
			{NewSQLBool(knd, true), NewSQLNull(knd), 1},
			{NewSQLBool(knd, true), NewSQLVarchar(knd, "56e0750e1d857aea925a4ba1"), -1},
			{NewSQLBool(knd, true), NewSQLVarchar(knd, "bac"), 1},
			{NewSQLBool(knd, true), NewSQLDate(knd, now), -1},
			{NewSQLBool(knd, true), NewSQLTimestamp(knd, now), -1},
			{NewSQLBool(knd, false), NewSQLInt64(knd, 0), 0},
			{NewSQLBool(knd, false), NewSQLInt64(knd, 1), -1},
			{NewSQLBool(knd, false), NewSQLInt64(knd, 2), -1},
			{NewSQLBool(knd, false), NewSQLUint64(knd, 1), -1},
			{NewSQLBool(knd, false), NewSQLFloat(knd, 1), -1},
			{NewSQLBool(knd, false), NewSQLBool(knd, false), 0},
			{NewSQLBool(knd, false), NewSQLBool(knd, true), -1},
			{NewSQLBool(knd, false), NewSQLNull(knd), 1},
			{NewSQLBool(knd, false), NewSQLVarchar(knd, "56e0750e1d857aea925a4ba1"), -1},
			{NewSQLBool(knd, false), NewSQLVarchar(knd, "bac"), 0},
			{NewSQLBool(knd, false), NewSQLDate(knd, now), -1},
			{NewSQLBool(knd, false), NewSQLTimestamp(knd, now), -1},
		}
		runTests(t, tests)
	})

	t.Run("SQLDate", func(t *testing.T) {
		tests := []test{
			{NewSQLDate(knd, now), NewSQLInt64(knd, 0), 1},
			{NewSQLDate(knd, now), NewSQLInt64(knd, 1), 1},
			{NewSQLDate(knd, now), NewSQLInt64(knd, 2), 1},
			{NewSQLDate(knd, now), NewSQLUint64(knd, 1), 1},
			{NewSQLDate(knd, now), NewSQLFloat(knd, 1), 1},
			{NewSQLDate(knd, now), NewSQLBool(knd, false), 1},
			{NewSQLDate(knd, now), NewSQLDate(knd, now.Add(diff)), -1},
			{NewSQLDate(knd, now), NewSQLNull(knd), 1},
			{NewSQLDate(knd, now), NewSQLVarchar(knd, "56e0750e1d857aea925a4ba1"), 1},
			{NewSQLDate(knd, now), NewSQLVarchar(knd, "bac"), 1},
			{NewSQLDate(knd, now), NewSQLDate(knd, now.Add(-diff)), 1},
			{NewSQLDate(knd, now), NewSQLTimestamp(knd, now.Add(diff)), -1},
			{NewSQLDate(knd, now), NewSQLTimestamp(knd, now.Add(-diff)), 1},
			{NewSQLDate(knd, now), NewSQLDate(knd, now), 0},
		}
		runTests(t, tests)
	})

	t.Run("SQLTimestamp", func(t *testing.T) {
		tests := []test{
			{NewSQLTimestamp(knd, now), NewSQLInt64(knd, 0), 1},
			{NewSQLTimestamp(knd, now), NewSQLInt64(knd, 1), 1},
			{NewSQLTimestamp(knd, now), NewSQLInt64(knd, 2), 1},
			{NewSQLTimestamp(knd, now), NewSQLUint64(knd, 1), 1},
			{NewSQLTimestamp(knd, now), NewSQLFloat(knd, 1), 1},
			{NewSQLTimestamp(knd, now), NewSQLBool(knd, false), 1},
			{NewSQLTimestamp(knd, now), NewSQLNull(knd), 1},
			{NewSQLTimestamp(knd, now), NewSQLVarchar(knd, "56e0750e1d857aea925a4ba1"), 1},
			{NewSQLTimestamp(knd, now), NewSQLVarchar(knd, "bac"), 1},
			{NewSQLTimestamp(knd, now), NewSQLTimestamp(knd, now.Add(diff)), -1},
			{NewSQLTimestamp(knd, now), NewSQLTimestamp(knd, now.Add(-diff)), 1},
			{NewSQLTimestamp(knd, now), NewSQLTimestamp(knd, now), 0},
			{NewSQLTimestamp(knd, now), NewSQLDate(knd, now), 1},
			{NewSQLTimestamp(knd, now), NewSQLDate(knd, now.Add(diff)), -1},
			{NewSQLTimestamp(knd, now), NewSQLDate(knd, now.Add(-diff)), 1},
			{NewSQLTimestamp(knd, now.Add(sameDayDiff)), NewSQLDate(knd, now), 1},
		}
		runTests(t, tests)
	})

	t.Run("SQLNullValue", func(t *testing.T) {
		tests := []test{
			{NewSQLNull(knd), NewSQLInt64(knd, 0), -1},
			{NewSQLNull(knd), NewSQLInt64(knd, 1), -1},
			{NewSQLNull(knd), NewSQLInt64(knd, 2), -1},
			{NewSQLNull(knd), NewSQLUint64(knd, 1), -1},
			{NewSQLNull(knd), NewSQLFloat(knd, 1), -1},
			{NewSQLNull(knd), NewSQLBool(knd, false), -1},
			{NewSQLNull(knd), NewSQLVarchar(knd, "56e0750e1d857aea925a4ba1"), -1},
			{NewSQLNull(knd), NewSQLVarchar(knd, "bac"), -1},
			{NewSQLNull(knd), NewSQLDate(knd, now), -1},
			{NewSQLNull(knd), NewSQLTimestamp(knd, now), -1},
			{NewSQLNull(knd), NewSQLNull(knd), 0},
		}
		runTests(t, tests)
	})

	t.Run("SQLVarchar", func(t *testing.T) {
		tests := []test{
			{NewSQLVarchar(knd, "bac"), NewSQLInt64(knd, 0), 0},
			{NewSQLVarchar(knd, "bac"), NewSQLInt64(knd, 1), -1},
			{NewSQLVarchar(knd, "bac"), NewSQLInt64(knd, 2), -1},
			{NewSQLVarchar(knd, "bac"), NewSQLUint64(knd, 1), -1},
			{NewSQLVarchar(knd, "bac"), NewSQLFloat(knd, 1), -1},
			{NewSQLVarchar(knd, "bac"), NewSQLBool(knd, false), 0},
			{NewSQLVarchar(knd, "bac"), NewSQLVarchar(knd, "56e0750e1d857aea925a4ba1"), 1},
			{NewSQLVarchar(knd, "bac"), NewSQLVarchar(knd, "cba"), -1},
			{NewSQLVarchar(knd, "bac"), NewSQLVarchar(knd, "bac"), 0},
			{NewSQLVarchar(knd, "bac"), NewSQLVarchar(knd, "abc"), 1},
			{NewSQLVarchar(knd, "bac"), NewSQLNull(knd), 1},
		}
		runTests(t, tests)
	})
}

func TestCompareToPairwise(t *testing.T) {
	// different lengths => error
	t.Run("different length slices returns an error", func(t *testing.T) {
		_, err := CompareToPairwise([]SQLValue{}, []SQLValue{NewSQLInt64(knd, 1)}, nil)
		require.NotNil(t, err, "expected error")
	})

	type test struct {
		name     string
		left     []SQLValue
		right    []SQLValue
		expected int
	}

	tests := []test{
		{"empty = empty", []SQLValue{}, []SQLValue{}, 0},
		{
			"single value = single value",
			[]SQLValue{NewSQLInt64(knd, 0)},
			[]SQLValue{NewSQLInt64(knd, 0)},
			0,
		},
		{
			"single value < single value",
			[]SQLValue{NewSQLInt64(knd, 0)},
			[]SQLValue{NewSQLInt64(knd, 1)},
			-1,
		},
		{
			"single value > single value",
			[]SQLValue{NewSQLInt64(knd, 1)},
			[]SQLValue{NewSQLInt64(knd, 0)},
			1,
		},
		{
			"multiple values first <",
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 1), NewSQLInt64(knd, 2)},
			[]SQLValue{NewSQLInt64(knd, 1), NewSQLInt64(knd, 1), NewSQLInt64(knd, 2)},
			-1,
		},
		{
			"multiple values first >",
			[]SQLValue{NewSQLInt64(knd, 1), NewSQLInt64(knd, 1), NewSQLInt64(knd, 2)},
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 1), NewSQLInt64(knd, 2)},
			1,
		},
		{
			"multiple values first =, mid <",
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 0), NewSQLInt64(knd, 2)},
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 1), NewSQLInt64(knd, 2)},
			-1,
		},
		{
			"multiple values first =, mid >",
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 1), NewSQLInt64(knd, 2)},
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 0), NewSQLInt64(knd, 2)},
			1,
		},
		{
			"multiple values first =, mid =, last <",
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 1), NewSQLInt64(knd, 1)},
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 1), NewSQLInt64(knd, 2)},
			-1,
		},
		{
			"multiple values first =, mid =, last >",
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 1), NewSQLInt64(knd, 2)},
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 1), NewSQLInt64(knd, 1)},
			1,
		},
		{
			"multiple values all equal",
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 1), NewSQLInt64(knd, 2)},
			[]SQLValue{NewSQLInt64(knd, 0), NewSQLInt64(knd, 1), NewSQLInt64(knd, 2)},
			0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := CompareToPairwise(tc.left, tc.right, nil)
			require.Nil(t, err, "unexpected error")
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestBoolIsFalsy(t *testing.T) {
	req := require.New(t)

	d, err := time.Parse("2006-01-02", "2003-01-02")
	req.NoError(err)

	ts, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
	req.NoError(err)

	t.Run("Bool", func(t *testing.T) {
		req := require.New(t)

		truthy := Bool(NewSQLTimestamp(knd, ts))
		req.True(truthy)

		truthy = Bool(NewSQLDate(knd, d))
		req.True(truthy)

		truthy = Bool(NewSQLInt64(knd, 0))
		req.False(truthy)

		truthy = Bool(NewSQLInt64(knd, 1))
		req.True(truthy)

		truthy = Bool(NewSQLVarchar(knd, "dsf"))
		req.False(truthy)

		truthy = Bool(NewSQLVarchar(knd, "16"))
		req.True(truthy)
	})

	t.Run("IsFalsy", func(t *testing.T) {
		req := require.New(t)

		truthy := IsFalsy(NewSQLTimestamp(knd, ts))
		req.False(truthy)

		truthy = IsFalsy(NewSQLDate(knd, d))
		req.False(truthy)

		truthy = IsFalsy(NewSQLInt64(knd, 0))
		req.True(truthy)

		truthy = IsFalsy(NewSQLInt64(knd, 1))
		req.False(truthy)

		truthy = IsFalsy(NewSQLVarchar(knd, "dsf"))
		req.True(truthy)

		truthy = IsFalsy(NewSQLVarchar(knd, "16"))
		req.False(truthy)
	})
}

func TestIsUUID(t *testing.T) {
	req := require.New(t)

	req.True(IsUUID(schema.MongoUUID))
	req.True(IsUUID(schema.MongoUUIDCSharp))
	req.True(IsUUID(schema.MongoUUIDJava))
	req.True(IsUUID(schema.MongoUUIDOld))
	req.False(IsUUID(schema.MongoString))
	req.False(IsUUID(schema.MongoGeo2D))
	req.False(IsUUID(schema.MongoObjectID))
	req.False(IsUUID(schema.MongoBool))
	req.False(IsUUID(schema.MongoInt))
	req.False(IsUUID(schema.MongoInt64))
}

func TestGetBinaryFromExpr(t *testing.T) {

	expected := []byte{
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c,
		0x0d, 0x0e, 0x0f, 0x10,
	}

	t.Run("invalid sqlexpr", func(t *testing.T) {
		req := require.New(t)
		_, ok := GetBinaryFromExpr(schema.MongoUUID, NewSQLValueExpr(NewSQLVarchar(knd, "3")))
		req.False(ok)
	})

	t.Run("with dashes", func(t *testing.T) {
		req := require.New(t)
		b, ok := GetBinaryFromExpr(schema.MongoUUID,
			NewSQLValueExpr(NewSQLVarchar(knd, "01020304-0506-0708-090a-0b0c0d0e0f10")))
		req.True(ok)
		req.Equal(byte(0x04), b.Subtype)
		req.Zero(convey.ShouldResemble(b.Data, expected))

		b, ok = GetBinaryFromExpr(schema.MongoUUIDOld,
			NewSQLValueExpr(NewSQLVarchar(knd, "01020304-0506-0708-090a-0b0c0d0e0f10")))
		req.True(ok)
		req.Equal(byte(0x03), b.Subtype)
		req.Zero(convey.ShouldResemble(b.Data, expected))
	})

	t.Run("without dashes", func(t *testing.T) {
		req := require.New(t)
		b, ok := GetBinaryFromExpr(schema.MongoUUIDJava,
			NewSQLValueExpr(NewSQLVarchar(knd, "0807060504030201100f0e0d0c0b0a09")))
		req.True(ok)
		req.Equal(byte(0x03), b.Subtype)
		req.Zero(convey.ShouldResemble(b.Data, expected))

		b, ok = GetBinaryFromExpr(schema.MongoUUIDCSharp,
			NewSQLValueExpr(NewSQLVarchar(knd, "0403020106050807090a0b0c0d0e0f10")))
		req.True(ok)
		req.Equal(byte(0x03), b.Subtype)
		req.Zero(convey.ShouldResemble(b.Data, expected))
	})
}

func TestNormalizeUUID(t *testing.T) {

	expected := []byte{
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c,
		0x0d, 0x0e, 0x0f, 0x10,
	}

	t.Run("standard", func(t *testing.T) {
		req := require.New(t)
		bytes := []byte{
			0x01, 0x02, 0x03, 0x04,
			0x05, 0x06, 0x07, 0x08,
			0x09, 0x0a, 0x0b, 0x0c,
			0x0d, 0x0e, 0x0f, 0x10,
		}
		req.NoError(NormalizeUUID(schema.MongoUUID, bytes))
		req.Zero(convey.ShouldResemble(bytes, expected))
	})

	t.Run("old", func(t *testing.T) {
		req := require.New(t)
		bytes := []byte{
			0x01, 0x02, 0x03, 0x04,
			0x05, 0x06, 0x07, 0x08,
			0x09, 0x0a, 0x0b, 0x0c,
			0x0d, 0x0e, 0x0f, 0x10,
		}
		req.NoError(NormalizeUUID(schema.MongoUUIDOld, bytes))
		req.Zero(convey.ShouldResemble(bytes, expected))
	})

	t.Run("csharp", func(t *testing.T) {
		req := require.New(t)
		bytes := []byte{
			0x04, 0x03, 0x02, 0x01,
			0x06, 0x05, 0x08, 0x07,
			0x09, 0x0a, 0x0b, 0x0c,
			0x0d, 0x0e, 0x0f, 0x10,
		}
		req.NoError(NormalizeUUID(schema.MongoUUIDCSharp, bytes))
		req.Zero(convey.ShouldResemble(bytes, expected))
	})

	t.Run("java", func(t *testing.T) {
		req := require.New(t)
		bytes := []byte{
			0x08, 0x07, 0x06, 0x05,
			0x04, 0x03, 0x02, 0x01,
			0x10, 0x0f, 0x0e, 0x0d,
			0x0c, 0x0b, 0x0a, 0x09,
		}
		req.NoError(NormalizeUUID(schema.MongoUUIDJava, bytes))
		req.Zero(convey.ShouldResemble(bytes, expected))
	})

}
