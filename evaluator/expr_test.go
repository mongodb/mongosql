package evaluator_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
)

func TestEvaluates(t *testing.T) {
	req := require.New(t)

	type test struct {
		name   string
		sql    string
		result evaluator.SQLExpr
	}

	runTests := func(t *testing.T, ctx *evaluator.EvalCtx, tests []test) {
		schema := evaluator.MustLoadSchema(testSchema3)
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req = require.New(t)
				subject, err := evaluator.GetSQLExpr(schema, dbOne, tableTwoName, test.sql)
				req.Nil(err, "unable to get SQLExpr for sql statement")
				result, err := subject.Evaluate(ctx)
				req.Nil(err, "unable to evaluate SQLExpr")
				req.Equal(result, test.result, "expected SQLExpr does not match evaluated SQLExpr")
			})
		}
	}

	type typeTest struct {
		name   string
		sql    string
		result schema.SQLType
	}

	runTypeTests := func(t *testing.T, ctx *evaluator.EvalCtx, tests []typeTest) {
		sc := evaluator.MustLoadSchema(testSchema3)
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req = require.New(t)
				subject, err := evaluator.GetSQLExpr(sc, dbOne, tableTwoName, test.sql)
				req.Nil(err, "unable to get SQLExpr for sql statement")
				result := subject.Type()
				req.Equal(
					result,
					test.result,
					"type of evaluated SQLExpr does not match expected type",
				)
			})
		}
	}

	execCtx := createTestExecutionCtx(nil)

	t.Run("evaluates", func(t *testing.T) {
		row := &evaluator.Row{
			Data: evaluator.Values{
				{
					SelectID: 1,
					Database: "test",
					Table:    "bar",
					Name:     "a",
					Data:     evaluator.SQLInt(123),
				},
				{
					SelectID: 1,
					Database: "test",
					Table:    "bar",
					Name:     "b",
					Data:     evaluator.SQLInt(456),
				},
				{
					SelectID: 1,
					Database: "test",
					Table:    "bar",
					Name:     "c",
					Data:     evaluator.SQLNull,
				},
			},
		}
		evalCtx := evaluator.NewEvalCtx(execCtx, collation.Default, row)

		// defines the scalar functions expressions to evaluates, along with
		// the name for the test and the expected result
		tests := []test{
			{"sql_add_expr_int_0", "0 + 0", evaluator.SQLInt(0)},
			{"sql_add_expr_int_1", "-1 + 1", evaluator.SQLInt(0)},
			{"sql_add_expr_int_2", "10 + 32", evaluator.SQLInt(42)},
			{"sql_add_expr_int_3", "-10 + -32", evaluator.SQLInt(-42)},
			{"sql_add_expr_bool_0", "true + true", evaluator.SQLFloat(2)},
			{"sql_add_expr_bool_1", "true + true + false", evaluator.SQLFloat(2)},
			{"sql_add_expr_bool_2", "false + true + true", evaluator.SQLFloat(2)},
			{"sql_add_expr_mixed_0", "true - '-1'", evaluator.SQLFloat(2)},
			{"sql_add_expr_mixed_1", "true + '0'", evaluator.SQLFloat(1)},
			{"sql_and_expr_0", "1 AND 1", evaluator.SQLTrue},
			{"sql_and_expr_1", "1 AND 0", evaluator.SQLFalse},
			{"sql_and_expr_2", "0 AND 1", evaluator.SQLFalse},
			{"sql_and_expr_3", "0 AND 0", evaluator.SQLFalse},
			{"sql_and_expr_4", "1 && 1", evaluator.SQLTrue},
			{"sql_and_expr_5", "1 && 0", evaluator.SQLFalse},
			{"sql_and_expr_6", "0 && 1", evaluator.SQLFalse},
			{"sql_and_expr_7", "0 && 0", evaluator.SQLFalse},
			{"sql_and_expr_with_null_0", "NULL && 0", evaluator.SQLFalse},
			{"sql_and_expr_with_null_1", "NULL && 1", evaluator.SQLNull},
			{"sql_and_expr_with_null_2", "NULL && NULL", evaluator.SQLNull},
			{"sql_and_expr_with_bool_0", "true AND true", evaluator.SQLTrue},
			{"sql_and_expr_with_bool_1", "true AND false", evaluator.SQLFalse},
			{"sql_and_expr_with_bool_2", "false AND true", evaluator.SQLFalse},
			{"sql_and_expr_with_bool_3", "false AND false", evaluator.SQLFalse},
			{"sql_benchmark_expr_0", "BENCHMARK(10, 1)", evaluator.SQLInt(0)},
			{"sql_benchmark_expr_1", "BENCHMARK(0, 10)", evaluator.SQLInt(0)},
			{"sql_benchmark_expr_2", "BENCHMARK(NULL, 0)", evaluator.SQLInt(0)},
			{"sql_date_expr_with_add_0", "DATE '2014-04-13' + 0", evaluator.SQLInt(20140413)},
			{"sql_date_expr_with_add_1", "DATE '2014-04-13' + 2", evaluator.SQLInt(20140415)},
			{
				"sql_time_expr_with_add_0",
				"TIME '11:04:13' + 0",
				evaluator.SQLDecimal128(decimal.NewFromFloat(110413)),
			},
			{
				"sql_time_expr_with_add_1",
				"TIME '11:04:13' + 2",
				evaluator.SQLDecimal128(decimal.NewFromFloat(110415)),
			},
			{
				"sql_time_expr_with_add_2",
				"TIME '11:04:13' + '2'",
				evaluator.SQLDecimal128(decimal.NewFromFloat(110415)),
			},
			{
				"sql_time_expr_with_add_3",
				"'2' + TIME '11:04:13'",
				evaluator.SQLDecimal128(decimal.NewFromFloat(110415)),
			},
			{
				"sql_timestamp_expr_with_add_0",
				"TIMESTAMP '2014-04-13 11:04:13' + 0",
				evaluator.SQLDecimal128(decimal.NewFromFloat(20140413110413)),
			},
			{
				"sql_timestamp_expr_with_add_1",
				"TIMESTAMP '2014-04-13 11:04:13' + 2",
				evaluator.SQLDecimal128(decimal.NewFromFloat(20140413110415)),
			},
			{"sql_date_expr_with_subtract_0", "DATE '2014-04-13' - 0", evaluator.SQLInt(20140413)},
			{"sql_date_expr_with_subtract_1", "DATE '2014-04-13' - 2", evaluator.SQLInt(20140411)},
			{
				"sql_time_expr_with_subtract_0",
				"TIME '11:04:13' - 0",
				evaluator.SQLDecimal128(decimal.NewFromFloat(110413)),
			},
			{
				"sql_time_expr_with_subtract_1",
				"TIME '11:04:13' - 2",
				evaluator.SQLDecimal128(decimal.NewFromFloat(110411)),
			},
			{
				"sql_time_expr_with_subtract_2",
				"TIME '11:04:13' - '2'",
				evaluator.SQLDecimal128(decimal.NewFromFloat(110411)),
			},
			{
				"sql_timestamp_expr_with_subtract_0",
				"TIMESTAMP '2014-04-13 11:04:13' - 0",
				evaluator.SQLDecimal128(decimal.NewFromFloat(20140413110413)),
			},
			{
				"sql_timestamp_expr_with_subtract_1",
				"TIMESTAMP '2014-04-13 11:04:13' - 2",
				evaluator.SQLDecimal128(decimal.NewFromFloat(20140413110411)),
			},
			{"sql_date_expr_with_multiply_0", "DATE '2014-04-13' * 0", evaluator.SQLInt(0)},
			{"sql_date_expr_with_multiply_0", "DATE '2014-04-13' * 2", evaluator.SQLInt(40280826)},
			{
				"sql_time_expr_with_multiply_0",
				"TIME '11:04:13' * 0",
				evaluator.SQLDecimal128(decimal.NewFromFloat(0)),
			},
			{
				"sql_time_expr_with_multiply_1",
				"TIME '11:04:13' * 2",
				evaluator.SQLDecimal128(decimal.NewFromFloat(220826)),
			},
			{
				"sql_time_expr_with_multiply_2",
				"TIME '11:04:13' * '2'",
				evaluator.SQLDecimal128(decimal.NewFromFloat(220826)),
			},
			{
				"sql_timestamp_expr_with_multiply_0",
				"TIMESTAMP '2014-04-13 11:04:13' * 0",
				evaluator.SQLDecimal128(decimal.NewFromFloat(0)),
			},
			{
				"sql_timestamp_expr_with_multiply_1",
				"TIMESTAMP '2014-04-13 11:04:13' * 2",
				evaluator.SQLDecimal128(decimal.NewFromFloat(40280826220826)),
			},
			{
				"sql_float_division_expr_0",
				"1.2 / 0.2",
				evaluator.SQLDecimal128(decimal.New(600000, -5)),
			},
			{
				"sql_float_division_expr_1",
				"1.2 / 0.23",
				evaluator.SQLDecimal128(decimal.New(521739, -5)),
			},
			{"sql_date_division_0", "DATE '2014-04-13' / 0", evaluator.SQLNull},
			{"sql_date_division_1", "DATE '2014-04-13' / 2", evaluator.SQLFloat(10070206.5)},
			{"sql_time_division_0", "TIME '11:04:13' / 0", evaluator.SQLNull},
			{
				"sql_time_division_1",
				"TIME '11:04:13' / 2",
				evaluator.SQLDecimal128(decimal.New(552065000, -4)),
			},
			{
				"sql_time_division_2",
				"TIME '11:04:13' / '2'",
				evaluator.SQLDecimal128(decimal.New(552065000, -4)),
			},
			{"sql_timestamp_division_0", "TIMESTAMP '2014-04-13 11:04:13' / 0", evaluator.SQLNull},
			{
				"sql_timestamp_division_1",
				"TIMESTAMP '2014-04-13 11:04:13' / 2",
				evaluator.SQLDecimal128(decimal.New(100702065552065000, -4)),
			},
			{"sql_date_less_than_0", "DATE '2014-04-13' < 0", evaluator.SQLFalse},
			{"sql_date_less_than_1", "DATE '2014-04-13' < DATE '2014-04-14'", evaluator.SQLTrue},
			{"sql_date_greater_than_0", "DATE '2014-04-13' > 0", evaluator.SQLTrue},
			{
				"sql_date_greater_than_1",
				"DATE '2014-04-13' > DATE '2014-04-14'",
				evaluator.SQLFalse,
			},
			{"sql_date_equality_0", "DATE '2014-04-13' = '0'", evaluator.SQLFalse},
			{"sql_date_equality_1", "DATE '2014-04-13' = DATE '2014-04-13'", evaluator.SQLTrue},
			{
				"sql_case_expr_0",
				"CASE 3 WHEN 3 THEN 'three' WHEN 1 THEN 'one' ELSE 'else' END",
				evaluator.SQLVarchar("three"),
			},
			{
				"sql_case_expr_1",
				"CASE WHEN 5 > 3 THEN 'true' else 'false' END",
				evaluator.SQLVarchar("true"),
			},
			{
				"sql_case_expr_2",
				"CASE WHEN a = 123 THEN 'yes' else 'no' END",
				evaluator.SQLVarchar("yes"),
			},
			{"sql_case_expr_3", "CASE WHEN a = 245 THEN 'yes' END", evaluator.SQLNull},
			{"sql_int_division_expr_0", "-1 / 1", evaluator.SQLFloat(-1)},
			{"sql_int_division_expr_1", "100 / 10", evaluator.SQLFloat(10)},
			{"sql_int_division_expr_2", "-10 / 10", evaluator.SQLFloat(-1)},
			{"sql_int_equality_expr_0", "0 = 0", evaluator.SQLTrue},
			{"sql_int_equality_expr_1", "-1 = 1", evaluator.SQLFalse},
			{"sql_int_equality_expr_2", "10 = 10", evaluator.SQLTrue},
			{"sql_int_equality_expr_3", "-10 = -10", evaluator.SQLTrue},
			{"sql_mixed_equality_expr", "false = '0'", evaluator.SQLTrue},
			{"sql_int_comparison_0", "0 > 0", evaluator.SQLFalse},
			{"sql_int_comparison_1", "-1 > 1", evaluator.SQLFalse},
			{"sql_int_comparison_2", "1 > -1", evaluator.SQLTrue},
			{"sql_int_comparison_3", "11 > 10", evaluator.SQLTrue},
			{"sql_mixed_comparison_0", "true > '-1'", evaluator.SQLTrue},
			{"sql_int_greater_than_0", "0 >= 0", evaluator.SQLTrue},
			{"sql_int_greater_than_1", "-1 >= 1", evaluator.SQLFalse},
			{"sql_int_greater_than_2", "1 >= -1", evaluator.SQLTrue},
			{"sql_int_greater_than_3", "11 >= 10", evaluator.SQLTrue},
			{"sql_is_bool_expr_0", "1 is true", evaluator.SQLTrue},
			{"sql_is_bool_expr_1", "null is true", evaluator.SQLFalse},
			{"sql_is_unknown_expr_0", "null is unknown", evaluator.SQLTrue},
			{"sql_is_unknown_expr_1", "1 is unknown", evaluator.SQLFalse},
			{"sql_is_expr_0", "true is true", evaluator.SQLTrue},
			{"sql_is_expr_1", "0 is false", evaluator.SQLTrue},
			{"sql_is_expr_2", "1 is false", evaluator.SQLFalse},
			{"sql_is_expr_3", "'1' is true", evaluator.SQLTrue},
			{"sql_is_expr_4", "'0.0' is true", evaluator.SQLFalse},
			{"sql_is_expr_5", "'cats' is false", evaluator.SQLTrue},
			{"sql_date_is_bool_expr_0", "DATE '2006-05-04' is false", evaluator.SQLFalse},
			{
				"sql_date_is_bool_expr_1",
				"TIMESTAMP '2008-04-06 15:32:23' is true",
				evaluator.SQLTrue,
			},
			{"sql_is_null_expr_0", "1 is null", evaluator.SQLFalse},
			{"sql_is_null_expr_1", "null is null", evaluator.SQLTrue},
			{"sql_is_not_expr_1", "1 is not true", evaluator.SQLFalse},
			{"sql_is_not_expr_2", "null is not true", evaluator.SQLTrue},
			{"sql_is_not_expr_3", "null is not unknown", evaluator.SQLFalse},
			{"sql_is_not_expr_4", "1 is not unknown", evaluator.SQLTrue},
			{"sql_is_not_expr_5", "false is not true", evaluator.SQLTrue},
			{"sql_is_not_expr_6", "0 is not false", evaluator.SQLFalse},
			{"sql_is_not_expr_7", "1 is not false", evaluator.SQLTrue},
			{"sql_is_not_expr_8", "'1' is not true", evaluator.SQLFalse},
			{"sql_is_not_expr_9", "'0.0' is not true", evaluator.SQLTrue},
			{"sql_is_not_expr_10", "'cats' is not false", evaluator.SQLFalse},
			{"sql_is_not_expr_11", "DATE '2006-05-04' is not false", evaluator.SQLTrue},
			{
				"sql_is_not_expr_12",
				"TIMESTAMP '2008-04-06 15:32:23' is not true",
				evaluator.SQLFalse,
			},
			{"sql_is_not_expr_13", "1 is not null", evaluator.SQLTrue},
			{"sql_is_not_expr_14", "null is not null", evaluator.SQLFalse},
			{"sql_divide_expr_0", "0 DIV 0", evaluator.SQLNull},
			{"sql_divide_expr_1", "0 DIV 5", evaluator.SQLInt(0)},
			{"sql_divide_expr_2", "5.5 DIV 2", evaluator.SQLInt(2)},
			{"sql_divide_expr_3", "-5 DIV 2", evaluator.SQLInt(-2)},
			{"sql_divide_expr_4", "NULL DIV 1", evaluator.SQLNull},
			{"sql_divide_expr_5", "1 DIV NULL", evaluator.SQLNull},
			{"sql_in_expr_0", "0 IN(0)", evaluator.SQLTrue},
			{"sql_in_expr_1", "-1 IN(1)", evaluator.SQLFalse},
			{"sql_in_expr_2", "0 IN(10, 0)", evaluator.SQLTrue},
			{"sql_in_expr_3", "-1 IN(1, 10)", evaluator.SQLFalse},
			{"sql_in_expr_4", "NULL IN(0, 1)", evaluator.SQLNull},
			{"sql_in_expr_5", "NULL IN(0, NULL)", evaluator.SQLNull},
			{"sql_less_than_expr_0", "0 < 0", evaluator.SQLFalse},
			{"sql_less_than_expr_1", "-1 < 1", evaluator.SQLTrue},
			{"sql_less_than_expr_2", "1 < -1", evaluator.SQLFalse},
			{"sql_less_than_expr_3", "10 < 11", evaluator.SQLTrue},
			{"sql_less_than_expr_4", "true < '5'", evaluator.SQLTrue},
			{"sql_less_than_or_equal_expr_0", "0 <= 0", evaluator.SQLTrue},
			{"sql_less_than_or_equal_expr_1", "-1 <= 1", evaluator.SQLTrue},
			{"sql_less_than_or_equal_expr_2", "1 <= -1", evaluator.SQLFalse},
			{"sql_less_than_or_equal_expr_3", "10 <= 11", evaluator.SQLTrue},
			{"sql_like_expr_0", "'Á Â Ã Ä' LIKE '%'", evaluator.SQLTrue},
			{"sql_like_expr_1", "'Á Â Ã Ä' LIKE 'Á Â Ã Ä'", evaluator.SQLTrue},
			{"sql_like_expr_2", "'Á Â Ã Ä' LIKE 'Á%'", evaluator.SQLTrue},
			{"sql_like_expr_3", "'a' LIKE 'a'", evaluator.SQLTrue},
			{"sql_like_expr_4", "'Adam' LIKE 'am'", evaluator.SQLFalse},
			{"sql_like_expr_5", "'Adam' LIKE 'adaM'", evaluator.SQLTrue}, // mixed case test
			{"sql_like_expr_6", "'Adam' LIKE '%am%'", evaluator.SQLTrue},
			{"sql_like_expr_7", "'Adam' LIKE 'Ada_'", evaluator.SQLTrue},
			{"sql_like_expr_8", "'Adam' LIKE '__am'", evaluator.SQLTrue},
			{"sql_like_expr_9", "'Clever' LIKE '%is'", evaluator.SQLFalse},
			{"sql_like_expr_10", "'Adam is nice' LIKE '%xs '", evaluator.SQLFalse},
			{"sql_like_expr_11", "'Adam is nice' LIKE '%is nice'", evaluator.SQLTrue},
			{"sql_like_expr_12", "'abc' LIKE 'ABC'", evaluator.SQLTrue},    //case sensitive test
			{"sql_like_expr_13", "'abc' LIKE 'ABC '", evaluator.SQLFalse},  // trailing space test
			{"sql_like_expr_14", "'abc' LIKE ' ABC'", evaluator.SQLFalse},  // leading space test
			{"sql_like_expr_15", "'abc' LIKE ' ABC '", evaluator.SQLFalse}, // padded space test
			{"sql_like_expr_16", "'abc' LIKE 'ABC	'", evaluator.SQLFalse}, // trailing tab test
			{"sql_like_expr_17", "'10' LIKE '1%'", evaluator.SQLTrue},
			{"sql_like_expr_18", "'a   ' LIKE 'A   '", evaluator.SQLTrue},
			{"sql_like_expr_19", "CURRENT_DATE() LIKE '2015-05-31%'", evaluator.SQLFalse},
			{"sql_like_expr_20", "CURDATE() LIKE '2015-05-31%'", evaluator.SQLFalse},
			{"sql_like_expr_21", "(DATE '2008-01-02') LIKE '2008-01%'", evaluator.SQLTrue},
			{
				"sql_like_expr_22",
				"NOW() LIKE '" + strconv.Itoa(time.Now().Year()) + "%' ",
				evaluator.SQLTrue,
			},
			{"sql_like_expr_23", "10 LIKE '1%'", evaluator.SQLTrue},
			{"sql_like_expr_24", "1.20 LIKE '1.2%'", evaluator.SQLTrue},
			{"sql_like_expr_25", "NULL LIKE '1%'", evaluator.SQLNull},
			{"sql_like_expr_26", "10 LIKE NULL", evaluator.SQLNull},
			{"sql_like_expr_27", "NULL LIKE NULL", evaluator.SQLNull},
			{"sql_like_expr_28", "'David_' LIKE 'David\\_'", evaluator.SQLTrue},
			{"sql_like_expr_29", "'David%' LIKE 'David\\%'", evaluator.SQLTrue},
			{"sql_like_expr_30", "'David_' LIKE 'David|_' ESCAPE '|'", evaluator.SQLTrue},
			{"sql_like_expr_31", "'David\\_' LIKE 'David\\_' ESCAPE ''", evaluator.SQLTrue},
			{"sql_like_expr_32", "'David_' LIKE 'David\\_' ESCAPE char(92)", evaluator.SQLTrue},
			{"sql_like_expr_33", "'David_' LIKE 'David|_' {escape '|'}", evaluator.SQLTrue},
			{"sql_mixed_arithmetic_and_bool_0", "(5<6) + 1", evaluator.SQLInt(2)},
			{"sql_mixed_arithmetic_and_bool_1", "(5<6) && (6>4)", evaluator.SQLTrue},
			{"sql_mixed_arithmetic_and_bool_2", "(5<6) || (6>4)", evaluator.SQLTrue},
			{"sql_mixed_arithmetic_and_bool_3", "(5<6) XOR (6>4)", evaluator.SQLFalse},
			{"sql_mixed_arithmetic_and_bool_4", "(5<6)<7", evaluator.SQLTrue},
			{"sql_mixed_arithmetic_and_bool_5", "1+(5<6)", evaluator.SQLInt(2)},
			{"sql_mixed_arithmetic_and_bool_6", "1+(5>6)", evaluator.SQLInt(1)},
			{"sql_mixed_arithmetic_and_bool_7", "1+(NULL>6)", evaluator.SQLNull},
			{"sql_mixed_arithmetic_and_bool_8", "NULL+(5>6)", evaluator.SQLNull},
			{"sql_mixed_arithmetic_and_bool_9", "20/(5<6)", evaluator.SQLFloat(20)},
			{"sql_mixed_arithmetic_and_bool_10", "20*(5<6)", evaluator.SQLInt(20)},
			{"sql_mixed_arithmetic_and_bool_11", "20/5<6", evaluator.SQLTrue},
			{"sql_mixed_arithmetic_and_bool_12", "20*5<6", evaluator.SQLFalse},
			{"sql_mixed_arithmetic_and_bool_13", "20+5<6", evaluator.SQLFalse},
			{"sql_mixed_arithmetic_and_bool_14", "20-5<6", evaluator.SQLFalse},
			{"sql_mixed_arithmetic_and_bool_15", "20+true", evaluator.SQLInt(21)},
			{"sql_mixed_arithmetic_and_bool_16", "20+false", evaluator.SQLInt(20)},
			{"sql_mod_expr_0", "0 % 0", evaluator.SQLNull},
			{"sql_mod_expr_1", "5 % 2", evaluator.SQLFloat(1)},
			{"sql_mod_expr_2", "5.5 % 2", evaluator.SQLFloat(1.5)},
			{"sql_mod_expr_3", "-5 % -3", evaluator.SQLFloat(-2)},
			{"sql_mod_expr_4", "5 MOD 2", evaluator.SQLFloat(1)},
			{"sql_mod_expr_5", "5.5 MOD 2", evaluator.SQLFloat(1.5)},
			{"sql_mod_expr_6", "-5 MOD -3", evaluator.SQLFloat(-2)},
			{"sql_mult_expr_0", "0 * 0", evaluator.SQLInt(0)},
			{"sql_mult_expr_1", "-1 * 1", evaluator.SQLInt(-1)},
			{"sql_mult_expr_2", "10 * 32", evaluator.SQLInt(320)},
			{"sql_mult_expr_3", "-10 * -32", evaluator.SQLInt(320)},
			{"sql_mult_expr_4", "2.5 * 3", evaluator.SQLDecimal128(decimal.New(75, -1))},
			{"sql_not_equal_expr_0", "0 <> 0", evaluator.SQLFalse},
			{"sql_not_equal_expr_1", "-1 <> 1", evaluator.SQLTrue},
			{"sql_not_equal_expr_2", "10 <> 10", evaluator.SQLFalse},
			{"sql_not_equal_expr_3", "-10 <> -10", evaluator.SQLFalse},
			{"sql_not_expr_0", "NOT 1", evaluator.SQLFalse},
			{"sql_not_expr_1", "NOT 0", evaluator.SQLTrue},
			{"sql_not_expr_2", "NOT true", evaluator.SQLFalse},
			{"sql_not_expr_3", "NOT false", evaluator.SQLTrue},
			{"sql_not_expr_4", "NOT NULL", evaluator.SQLNull},
			{"sql_not_expr_5", "! 1", evaluator.SQLFalse},
			{"sql_not_expr_6", "! 0", evaluator.SQLTrue},
			{"sql_null_safe_equal_0", "0 <=> 0", evaluator.SQLTrue},
			{"sql_null_safe_equal_1", "-1 <=> 1", evaluator.SQLFalse},
			{"sql_null_safe_equal_2", "10 <=> 10", evaluator.SQLTrue},
			{"sql_null_safe_equal_3", "-10 <=> -10", evaluator.SQLTrue},
			{"sql_null_safe_equal_4", "1 <=> 1", evaluator.SQLTrue},
			{"sql_null_safe_equal_5", "NULL <=> NULL", evaluator.SQLTrue},
			{"sql_null_safe_equal_6", "1 <=> NULL", evaluator.SQLFalse},
			{"sql_null_safe_equal_7", "NULL <=> 1", evaluator.SQLFalse},
			{"sql_or_expr_0", "1 OR 1", evaluator.SQLTrue},
			{"sql_or_expr_1", "1 OR 0", evaluator.SQLTrue},
			{"sql_or_expr_2", "0 OR 1", evaluator.SQLTrue},
			{"sql_or_expr_3", "NULL OR 1", evaluator.SQLTrue},
			{"sql_or_expr_4", "NULL OR 0", evaluator.SQLNull},
			{"sql_or_expr_5", "NULL OR NULL", evaluator.SQLNull},
			{"sql_or_expr_6", "0 OR 0", evaluator.SQLFalse},
			{"sql_or_expr_7", "true OR true", evaluator.SQLTrue},
			{"sql_or_expr_8", "true OR false", evaluator.SQLTrue},
			{"sql_or_expr_9", "false OR true", evaluator.SQLTrue},
			{"sql_or_expr_10", "false OR false", evaluator.SQLFalse},
			{"sql_or_expr_11", "1 || 1", evaluator.SQLTrue},
			{"sql_or_expr_12", "1 || 0", evaluator.SQLTrue},
			{"sql_or_expr_13", "0 || 1", evaluator.SQLTrue},
			{"sql_or_expr_14", "0 || 0", evaluator.SQLFalse},
			{"sql_xor_expr_0", "1 XOR 1", evaluator.SQLFalse},
			{"sql_xor_expr_1", "1 XOR 0", evaluator.SQLTrue},
			{"sql_xor_expr_2", "0 XOR 1", evaluator.SQLTrue},
			{"sql_xor_expr_3", "0 XOR 0", evaluator.SQLFalse},
			{"sql_not_regex_expr_0", "'ABC123' NOT REGEXP 'AB'", evaluator.SQLFalse},
			{"sql_not_regex_expr_1", "'ABC123' NOT REGEXP 'ABD'", evaluator.SQLTrue},
			{"sql_not_regex_expr_2", "'ABC123' NOT REGEXP '[[:alpha:]]'", evaluator.SQLFalse},
			{"sql_not_regex_expr_3", "'fofo' NOT REGEXP '^fo'", evaluator.SQLFalse},
			{"sql_not_regex_expr_4", "'fofo' NOT REGEXP '^f.*$'", evaluator.SQLFalse},
			{"sql_not_regex_expr_5", "'pi' NOT REGEXP 'pi|apa'", evaluator.SQLFalse},
			{"sql_not_regex_expr_6", "'abcde' NOT REGEXP 'a[bcd]{2}e'", evaluator.SQLTrue},
			{"sql_not_regex_expr_7", "'abcde' NOT REGEXP 'a[bcd]{1,10}e'", evaluator.SQLFalse},
			{"sql_not_regex_expr_8", "null REGEXP 'abc'", evaluator.SQLNull},
			{"sql_not_regex_expr_9", "'a' REGEXP null", evaluator.SQLNull},
			{"sql_not_regex_expr_10", "2-1 NOT REGEXP '1'", evaluator.SQLFalse},
			{"sql_regex_expr_0", "'ABC123' REGEXP 'AB'", evaluator.SQLTrue},
			{"sql_regex_expr_1", "'ABC123' REGEXP 'ABD'", evaluator.SQLFalse},
			{"sql_regex_expr_2", "'ABC123' REGEXP '[[:alpha:]]'", evaluator.SQLTrue},
			{"sql_regex_expr_3", "'fofo' REGEXP '^fo'", evaluator.SQLTrue},
			{"sql_regex_expr_4", "'fofo' REGEXP '^f.*$'", evaluator.SQLTrue},
			{"sql_regex_expr_5", "'pi' REGEXP 'pi|apa'", evaluator.SQLTrue},
			{"sql_regex_expr_6", "'abcde' REGEXP 'a[bcd]{2}e'", evaluator.SQLFalse},
			{"sql_regex_expr_7", "'abcde' REGEXP 'a[bcd]{1,10}e'", evaluator.SQLTrue},
			{"sql_regex_expr_8", "null REGEXP 'abc'", evaluator.SQLNull},
			{"sql_regex_expr_9", "'a' REGEXP null", evaluator.SQLNull},
			{"sql_regex_expr_10", "2-1 REGEXP '1'", evaluator.SQLTrue},
			{"sql_scalar_abs_expr_0", "ABS(NULL)", evaluator.SQLNull},
			{"sql_scalar_abs_expr_1", "ABS('C')", evaluator.SQLFloat(0)},
			{"sql_scalar_abs_expr_2", "ABS(-20)", evaluator.SQLFloat(20)},
			{"sql_scalar_abs_expr_3", "ABS(20)", evaluator.SQLFloat(20)},
			{"sql_scalar_acos_expr_0", "ACOS(NULL)", evaluator.SQLNull},
			{"sql_scalar_acos_expr_1", "ACOS(20)", evaluator.SQLNull},
			{"sql_scalar_acos_expr_2", "ACOS(-20)", evaluator.SQLNull},
			{"sql_scalar_acos_expr_3", "ACOS('C')", evaluator.SQLFloat(1.5707963267948966)},
			{"sql_scalar_acos_expr_4", "ACOS(0)", evaluator.SQLFloat(1.5707963267948966)},
			{"sql_scalar_asin_expr_0", "ASIN(NULL)", evaluator.SQLNull},
			{"sql_scalar_asin_expr_1", "ASIN(20)", evaluator.SQLNull},
			{"sql_scalar_asin_expr_2", "ASIN(-20)", evaluator.SQLNull},
			{"sql_scalar_asin_expr_3", "ASIN('C')", evaluator.SQLFloat(0)},
			{"sql_scalar_asin_expr_4", "ASIN(0)", evaluator.SQLFloat(0)},
			{"sql_scalar_atan_expr_0", "ATAN(NULL)", evaluator.SQLNull},
			{"sql_scalar_atan_expr_1", "ATAN(20)", evaluator.SQLFloat(1.5208379310729538)},
			{"sql_scalar_atan_expr_2", "ATAN(-20)", evaluator.SQLFloat(-1.5208379310729538)},
			{"sql_scalar_atan_expr_3", "ATAN('C')", evaluator.SQLFloat(0)},
			{"sql_scalar_atan_expr_4", "ATAN(0)", evaluator.SQLFloat(0)},
			{"sql_scalar_atan_expr_5", "ATAN(NULL, NULL)", evaluator.SQLNull},
			{"sql_scalar_atan_expr_6", "ATAN(-2, 2)", evaluator.SQLFloat(-0.7853981633974483)},
			{"sql_scalar_atan_expr_7", "ATAN('C', 2)", evaluator.SQLFloat(0)},
			{"sql_scalar_atan_expr_8", "ATAN(0, 2)", evaluator.SQLFloat(0)},
			{"sql_scalar_atan2_expr_0", "ATAN2(NULL, NULL)", evaluator.SQLNull},
			{"sql_scalar_atan2_expr_1", "ATAN2(-2, 2)", evaluator.SQLFloat(-0.7853981633974483)},
			{"sql_scalar_atan2_expr_2", "ATAN2('C', 2)", evaluator.SQLFloat(0)},
			{"sql_scalar_atan2_expr_3", "ATAN2(0, 2)", evaluator.SQLFloat(0)},
			{"sql_ascii_0", "ASCII(NULL)", evaluator.SQLNull},
			{"sql_ascii_1", "ASCII('')", evaluator.SQLInt(0)},
			{"sql_ascii_2", "ASCII('A')", evaluator.SQLInt(65)},
			{"sql_ascii_3", "ASCII('AWESOME')", evaluator.SQLInt(65)},
			{"sql_ascii_4", "ASCII('¢')", evaluator.SQLInt(194)},
			{
				"sql_ascii_5",
				"ASCII('Č')",
				evaluator.SQLInt(196), // This is actually 268, but the first byte is 196
			},
			{"sql_ceil_0", "CEIL(NULL)", evaluator.SQLNull},
			{"sql_ceil_1", "CEIL(20)", evaluator.SQLFloat(20)},
			{"sql_ceil_2", "CEIL(-20)", evaluator.SQLFloat(-20)},
			{"sql_ceil_3", "CEIL('C')", evaluator.SQLFloat(0)},
			{"sql_ceil_4", "CEIL(0.56)", evaluator.SQLFloat(1)},
			{"sql_ceil_5", "CEIL(-0.56)", evaluator.SQLFloat(0)},
			{"sql_ceiling_expr_0", "CEIL(NULL)", evaluator.SQLNull},
			{"sql_ceiling_expr_1", "CEIL(20)", evaluator.SQLFloat(20)},
			{"sql_ceiling_expr_2", "CEIL(-20)", evaluator.SQLFloat(-20)},
			{"sql_ceiling_expr_3", "CEIL('C')", evaluator.SQLFloat(0)},
			{"sql_ceiling_expr_4", "CEIL(0.56)", evaluator.SQLFloat(1)},
			{"sql_ceiling_expr_5", "CEIL(-0.56)", evaluator.SQLFloat(0)},
			{"sql_char_expr_0", "CHAR(NULL)", evaluator.SQLVarchar("")},
			{"sql_char_expr_1", "CHAR(77,121,83,81,'76')", evaluator.SQLVarchar("MySQL")},
			{
				"sql_char_expr_2",
				"CHAR(77,121,NULL, 83, NULL, 81,'76')",
				evaluator.SQLVarchar("MySQL"),
			},
			{"sql_char_expr_3", "CHAR(256)", evaluator.SQLVarchar(string([]byte{1, 0}))},
			{"sql_char_expr_4", "CHAR(512)", evaluator.SQLVarchar(string([]byte{2, 0}))},
			{"sql_char_expr_5", "CHAR(513)", evaluator.SQLVarchar(string([]byte{2, 1}))},
			{"sql_char_expr_6", "CHAR(256, 512)", evaluator.SQLVarchar(string([]byte{1, 0, 2, 0}))},
			{"sql_char_expr_7", "CHAR(65537)", evaluator.SQLVarchar(string([]byte{1, 0, 1}))},
			{"sql_char_length_0", "CHAR_LENGTH(NULL)", evaluator.SQLNull},
			{"sql_char_length_1", "CHAR_LENGTH('sDg')", evaluator.SQLInt(3)},
			{"sql_char_length_2", "CHAR_LENGTH('世界')", evaluator.SQLInt(2)},
			{"sql_char_length_3", "CHAR_LENGTH('')", evaluator.SQLInt(0)},
			{"sql_char_length_4", "CHARACTER_LENGTH(NULL)", evaluator.SQLNull},
			{"sql_char_length_5", "CHARACTER_LENGTH('sDg')", evaluator.SQLInt(3)},
			{"sql_char_length_6", "CHARACTER_LENGTH('世界')", evaluator.SQLInt(2)},
			{"sql_char_length_7", "CHARACTER_LENGTH('')", evaluator.SQLInt(0)},
			{"sql_coalesce_expr_0", "COALESCE(NULL)", evaluator.SQLNull},
			{"sql_coalesce_expr_1", "COALESCE('A')", evaluator.SQLVarchar("A")},
			{"sql_coalesce_expr_2", "COALESCE('A', NULL)", evaluator.SQLVarchar("A")},
			{"sql_coalesce_expr_3", "COALESCE('A', 'B')", evaluator.SQLVarchar("A")},
			{"sql_coalesce_expr_4", "COALESCE(NULL, 'A', NULL, 'B')", evaluator.SQLVarchar("A")},
			{"sql_coalesce_expr_5", "COALESCE(NULL, NULL, NULL)", evaluator.SQLNull},
			{"sql_concat_expr_0", "CONCAT(NULL)", evaluator.SQLNull},
			{"sql_concat_expr_1", "CONCAT('A')", evaluator.SQLVarchar("A")},
			{"sql_concat_expr_2", "CONCAT('A', 'B')", evaluator.SQLVarchar("AB")},
			{"sql_concat_expr_3", "CONCAT('A', NULL, 'B')", evaluator.SQLNull},
			{"sql_concat_expr_4", "CONCAT('A', 123, 'B')", evaluator.SQLVarchar("A123B")},
			{"sql_concat_ws_expr_0", "CONCAT_WS(NULL, NULL)", evaluator.SQLNull},
			{"sql_concat_ws_expr_1", "CONCAT_WS(',','A')", evaluator.SQLVarchar("A")},
			{"sql_concat_ws_expr_2", "CONCAT_WS(',','A', 'B')", evaluator.SQLVarchar("A,B")},
			{"sql_concat_ws_expr_3", "CONCAT_WS(',','A', NULL, 'B')", evaluator.SQLVarchar("A,B")},
			{
				"sql_concat_ws_expr_4",
				"CONCAT_WS(',','A', 123, 'B')",
				evaluator.SQLVarchar("A,123,B"),
			},
			{"sql_connection_id_expr", "CONNECTION_ID()", evaluator.SQLUint32(42)},
			{"sql_cos_expr_0", "COS(NULL)", evaluator.SQLNull},
			{"sql_cos_expr_1", "COS(20)", evaluator.SQLFloat(0.40808206181339196)},
			{"sql_cos_expr_2", "COS(-20)", evaluator.SQLFloat(0.40808206181339196)},
			{"sql_cos_expr_3", "COS('C')", evaluator.SQLFloat(1)},
			{"sql_cos_expr_4", "COS(0)", evaluator.SQLFloat(1)},
			{"sql_cot_expr_0", "COT(NULL)", evaluator.SQLNull},
			{"sql_cot_expr_1", "COT(19)", evaluator.SQLFloat(6.596764247280111)},
			{"sql_cot_expr_2", "COT(-19)", evaluator.SQLFloat(-6.596764247280111)},
			// current time tests do not work
			/*
				{
					"sql_current_date_expr",
					"CURRENT_DATE()",
					evaluator.SQLDate{Time: time.Now().UTC()},
				},
				{
					"sql_current_ts_0",
					"CURRENT_TIMESTAMP()",
					evaluator.SQLTimestamp{Time: time.Now().UTC()},
				},
				{
					"sql_current_ts_1",
					"CURRENT_TIMESTAMP",
					evaluator.SQLTimestamp{Time: time.Now().UTC()},
				},
				{
					"sql_curtime_0",
					"CURRENT_TIMESTAMP()",
					evaluator.SQLTimestamp{Time: time.Now().UTC()},
				},
				{
					"sql_curtime_1",
					"CURRENT_TIMESTAMP",
					evaluator.SQLTimestamp{Time: time.Now().UTC()},
				},
				{
					"sql_utc_ts_0",
					"UTC_TIMESTAMP()",
					evaluator.SQLTimestamp{Time: time.Now().UTC()},
				},
				{
					"sql_utc_ts_1",
					"UTC_TIMESTAMP",
					evaluator.SQLTimestamp{Time: time.Now().UTC()},
				},
				{
					"sql_now_0",
					"NOW()",
					evaluator.SQLTimestamp{Time: time.Now().UTC()},
				},
			*/
			{"sql_user_expr_0", "CURRENT_USER()", evaluator.SQLVarchar("test user")},
			{"sql_user_expr_1", "SESSION_USER()", evaluator.SQLVarchar("test user")},
			{"sql_user_expr_2", "SYSTEM_USER()", evaluator.SQLVarchar("test user")},
			{"sql_user_expr_3", "USER()", evaluator.SQLVarchar("test user")},
			{"sql_db_0", "DATABASE()", evaluator.SQLVarchar("test")},
			{"sql_schema_0", "SCHEMA()", evaluator.SQLVarchar("test")},
			{
				"sql_date_diff_0",
				"DATEDIFF('2017-01-01', '2016-01-01 23:08:56')",
				evaluator.SQLInt(366),
			},
			{"sql_date_diff_1", "DATEDIFF('2017-01-01', '2017-01-01')", evaluator.SQLInt(0)},
			{
				"sql_date_diff_2",
				"DATEDIFF('2017-08-23 10:40:43', '2017-09-30 12:19:50')",
				evaluator.SQLInt(-38),
			},
			{"sql_date_diff_3", "DATEDIFF(NULL, '2017-09-30 12:19:50')", evaluator.SQLNull},
			{"sql_date_diff_4", "DATEDIFF('2002-09-07', '1700-08-02')", evaluator.SQLInt(106751)},
			{"sql_date_diff_5", "DATEDIFF('1657-08-02', '2002-09-07')", evaluator.SQLInt(-106751)},
			{
				"sql_date_diff_6",
				"DATEDIFF(20170823104043, '2017-09-30 12:19:50')",
				evaluator.SQLInt(-38),
			},
			{
				"sql_date_diff_7",
				"DATEDIFF(20170823.09809, '2017-09-30 12:19:50')",
				evaluator.SQLInt(-38),
			},
			{
				"sql_date_diff_8",
				"DATEDIFF('biconnectorisfun', '2017-09-30 12:19:50')",
				evaluator.SQLNull,
			},
			{"sql_date_diff_9", "DATEDIFF('2000-9-1', '2012-6-7')", evaluator.SQLInt(-4297)},
			{"sql_date_diff_10", "DATEDIFF('00-09-1', '12-06-07')", evaluator.SQLInt(-4297)},
			{"sql_date_format_0", "DATE_FORMAT('2009-10-04', NULL)", evaluator.SQLNull},
			{"sql_date_format_1", "DATE_FORMAT(NULL, '2009-10-04')", evaluator.SQLNull},
			{
				"sql_date_format_2",
				"DATE_FORMAT('2009-10-04 22:23:00', '%W %M 01 %Y')",
				evaluator.SQLVarchar("Sunday October 01 2009"),
			},
			{
				"sql_date_format_3",
				"DATE_FORMAT('2009-10-04 22:23:00', '%W %M %Y')",
				evaluator.SQLVarchar("Sunday October 2009"),
			},
			{
				"sql_date_format_4",
				"DATE_FORMAT('2007-10-04 22:23:00', '%H:01:%i:%s')",
				evaluator.SQLVarchar("22:01:23:00"),
			},
			{
				"sql_date_format_5",
				"DATE_FORMAT('2007-10-04 22:23:00', '%H:%g:01%%:%i:%s%')",
				evaluator.SQLVarchar("22:%g:01%:23:00%"),
			},
			{
				"sql_date_format_6",
				"DATE_FORMAT('2007-10-04 22:23:00', '%H:%i:%s')",
				evaluator.SQLVarchar("22:23:00"),
			},
			{
				"sql_date_format_7",
				"DATE_FORMAT('1900-10-04 22:23:00', '%D %y %a %d %m %b %j')",
				evaluator.SQLVarchar("4th 00 Thu 04 10 Oct 277"),
			},
			{
				"sql_date_format_8",
				"DATE_FORMAT('1997-10-04 22:23:00', '%H %k %I %r %T %S %w')",
				evaluator.SQLVarchar("22 22 10 10:23:00 PM 22:23:00 00 6"),
			},
			{
				"sql_date_format_9",
				"DATE_FORMAT('1999-01-01', '%X %V')",
				evaluator.SQLVarchar("1998 52"),
			},
			{
				"sql_date_format_10",
				"DATE_FORMAT('1989-05-14 01:03:01.232335','%a|%b|%c|%D|%d|%e|%f|%H|%h|%I|%i|%j|%k" +
					"|%l|%M|%m|%p|%r|%S|%s|%T|%U|%u|%V|%v|%W|%w|%X|%x|%Y|%y|%%|%x')",
				evaluator.SQLVarchar("Sun|May|5|14th|14|14|232335|01|01|01|03|134|1|1|May|05|AM" +
					"|01:03:01 AM|01|01|01:03:01|20|19|20|19|Sunday|0|1989|1989|1989|89|%|1989"),
			},
			{
				"sql_date_format_11",
				"DATE_FORMAT('1900-10-04 22:23:00', '%a|%b|%c|%D|%d|%e|%f|%H|%h|%I|%i|%j|%k|%l|%M" +
					"|%m|%p|%r|%S|%s|%T|%U|%u|%V|%v|%W|%w|%X|%x|%Y|%y|%%|%x')",
				evaluator.SQLVarchar("Thu|Oct|10|4th|04|4|000000|22|10|10|23|277|22|10|October|10" +
					"|PM|10:23:00 PM|00|00|22:23:00|39|40|39|40|Thursday|4|1900" +
					"|1900|1900|00|%|1900"),
			},
			{
				"sql_date_format_12",
				"DATE_FORMAT('1983-07-05 23:22', '%a|%b|%c|%D|%d|%e|%f|%H|%h|%I|%i|%j|%k|%l|%M|%m" +
					"|%p|%r|%S|%s|%T|%U|%u|%V|%v|%W|%w|%X|%x|%Y|%y|%%|%x')",
				evaluator.SQLVarchar("Tue|Jul|7|5th|05|5|000000|23|11|11|22|186|23|11|July|07|PM" +
					"|11:22:00 PM|00|00|23:22:00|27|27|27|27|Tuesday|2|1983|1983|1983|83|%|1983"),
			},
			{"sql_date_name_0", "DAYNAME(NULL)", evaluator.SQLNull},
			{"sql_date_name_1", "DAYNAME(14)", evaluator.SQLNull},
			{"sql_date_name_2", "DAYNAME('2016-01-01 00:00:00')", evaluator.SQLVarchar("Friday")},
			{"sql_date_name_3", "DAYNAME('2016-1-1')", evaluator.SQLVarchar("Friday")},
			{"sql_date_of_month_0", "DAYOFMONTH(NULL)", evaluator.SQLNull},
			{"sql_date_of_month_1", "DAYOFMONTH(14)", evaluator.SQLNull},
			{"sql_date_of_month_2", "DAYOFMONTH('2016-01-01')", evaluator.SQLInt(1)},
			{"sql_date_of_month_3", "DAYOFMONTH('2016-1-1')", evaluator.SQLInt(1)},
			{"sql_date_of_week_0", "DAYOFWEEK(NULL)", evaluator.SQLNull},
			{"sql_date_of_week_1", "DAYOFWEEK(14)", evaluator.SQLNull},
			{"sql_date_of_week_2", "DAYOFWEEK('2016-01-01')", evaluator.SQLInt(6)},
			{"sql_date_of_week_3", "DAYOFWEEK('2016-1-1')", evaluator.SQLInt(6)},
			{"sql_date_of_year_0", "DAYOFYEAR(NULL)", evaluator.SQLNull},
			{"sql_date_of_year_1", "DAYOFYEAR(14)", evaluator.SQLNull},
			{"sql_date_of_year_2", "DAYOFYEAR('2016-1-1')", evaluator.SQLInt(1)},
			{"sql_date_of_year_3", "DAYOFYEAR('2016-01-01')", evaluator.SQLInt(1)},
			{"sql_degrees_0", "DEGREES(NULL)", evaluator.SQLNull},
			{"sql_degrees_1", "DEGREES(20)", evaluator.SQLFloat(1145.9155902616465)},
			{"sql_degrees_2", "DEGREES(-20)", evaluator.SQLFloat(-1145.9155902616465)},
			{"sql_elt_expr_0", "ELT(NULL, 'a', 'b')", evaluator.SQLNull},
			{"sql_elt_expr_1", "ELT(0, 'a', 'b')", evaluator.SQLNull},
			{"sql_elt_expr_2", "ELT(1, 'a', 'b')", evaluator.SQLVarchar("a")},
			{"sql_elt_expr_3", "ELT(2, 'a', 'b')", evaluator.SQLVarchar("b")},
			{"sql_elt_expr_4", "ELT(3, 'a', 'b', NULL)", evaluator.SQLNull},
			{"sql_elt_expr_5", "ELT(4, 'a', 'b', NULL)", evaluator.SQLNull},
			{"sql_exp_expr_0", "EXP(NULL)", evaluator.SQLNull},
			{"sql_exp_expr_1", "EXP('sdg')", evaluator.SQLFloat(1)},
			{"sql_exp_expr_2", "EXP(0)", evaluator.SQLFloat(1)},
			{"sql_exp_expr_3", "EXP(2)", evaluator.SQLFloat(7.38905609893065)},
			{"sql_extract_expr_0", "EXTRACT(YEAR FROM NULL)", evaluator.SQLNull},
			{
				"sql_extract_expr_1",
				"EXTRACT(YEAR FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(2006),
			},
			{
				"sql_extract_expr_2",
				"EXTRACT(QUARTER FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(2),
			},
			{
				"sql_extract_expr_3",
				"EXTRACT(WEEK FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(14),
			},
			{
				"sql_extract_expr_4",
				"EXTRACT(DAY FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(7),
			},
			{
				"sql_extract_expr_5",
				"EXTRACT(HOUR FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(7),
			},
			{
				"sql_extract_expr_6",
				"EXTRACT(MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(14),
			},
			{
				"sql_extract_expr_7",
				"EXTRACT(SECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(23),
			},
			{
				"sql_extract_expr_8",
				"EXTRACT(MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(0),
			},
			{
				"sql_extract_expr_9",
				"EXTRACT(YEAR_MONTH FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(200604),
			},
			{
				"sql_extract_expr_10",
				"EXTRACT(DAY_HOUR FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(707),
			},
			{
				"sql_extract_expr_11",
				"EXTRACT(DAY_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(70714),
			},
			{
				"sql_extract_expr_12",
				"EXTRACT(DAY_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(7071423),
			},
			{
				"sql_extract_expr_13",
				"EXTRACT(DAY_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(7071423000000),
			},
			{
				"sql_extract_expr_14",
				"EXTRACT(HOUR_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(714),
			},
			{
				"sql_extract_expr_15",
				"EXTRACT(HOUR_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(71423),
			},
			{
				"sql_extract_expr_16",
				"EXTRACT(HOUR_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(71423000000),
			},
			{
				"sql_extract_expr_17",
				"EXTRACT(MINUTE_SECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(1423),
			},
			{
				"sql_extract_expr_18",
				"EXTRACT(MINUTE_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(1423000000),
			},
			{
				"sql_extract_expr_19",
				"EXTRACT(SECOND_MICROSECOND FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(23000000),
			},
			{
				"sql_extract_expr_20",
				"EXTRACT(SQL_TSI_MINUTE FROM TIMESTAMP '2006-04-07 07:14:23')",
				evaluator.SQLInt(14),
			},
			{"sql_floor_expr_0", "FLOOR(NULL)", evaluator.SQLNull},
			{"sql_floor_expr_1", "FLOOR('sdg')", evaluator.SQLFloat(0)},
			{"sql_floor_expr_2", "FLOOR(1.23)", evaluator.SQLFloat(1)},
			{"sql_floor_expr_3", "FLOOR(-1.23)", evaluator.SQLFloat(-2)},
			{"sql_from_unixtime_0", "FROM_UNIXTIME(NULL)", evaluator.SQLNull},
			{"sql_from_unixtime_1", "FROM_UNIXTIME(-1)", evaluator.SQLNull},
			{
				"sql_from_unixtime_2",
				"FROM_UNIXTIME(1447430881) + 0",
				evaluator.SQLDecimal128(decimal.New(20151113160801, 0)),
			},
			{
				"sql_from_unixtime_3",
				"FROM_UNIXTIME(1447430881.5) + 0",
				evaluator.SQLDecimal128(decimal.New(20151113160802, 0)),
			},
			{
				"sql_from_unixtime_4",
				"CONCAT(FROM_UNIXTIME(1447430881), '')",
				evaluator.SQLVarchar("2015-11-13 16:08:01.000000"),
			},
			{
				"sql_from_unixtime_5",
				"CONCAT(FROM_UNIXTIME(1447430881.5), '')",
				evaluator.SQLVarchar("2015-11-13 16:08:02.000000"),
			},
			{"sql_hour_0", "HOUR(NULL)", evaluator.SQLNull},
			{"sql_hour_1", "HOUR('sdg')", evaluator.SQLInt(0)},
			{"sql_hour_2", "HOUR('10:23:52')", evaluator.SQLInt(10)},
			{"sql_hour_3", "HOUR('10:61:52')", evaluator.SQLNull},
			{"sql_hour_4", "HOUR('10:23:52.23.25.26')", evaluator.SQLInt(10)},
			{"sql_if_expr_0", "IF(1<2, 4, 5)", evaluator.SQLInt(4)},
			{"sql_if_expr_1", "IF(1>2, 4, 5)", evaluator.SQLInt(5)},
			{"sql_if_expr_2", "IF(1, 4, 5)", evaluator.SQLInt(4)},
			{"sql_if_expr_3", "IF(-0, 4, 5)", evaluator.SQLInt(5)},
			{"sql_if_expr_4", "IF(1-1, 4, 5)", evaluator.SQLInt(5)},
			{"sql_if_expr_5", "IF('cat', 4, 5)", evaluator.SQLInt(5)},
			{"sql_if_expr_6", "IF('3', 4, 5)", evaluator.SQLInt(4)},
			{"sql_if_expr_7", "IF('0', 4, 5)", evaluator.SQLInt(5)},
			{"sql_if_expr_8", "IF('-0.0', 4, 5)", evaluator.SQLInt(5)},
			{"sql_if_expr_9", "IF('2.2', 4, 5)", evaluator.SQLInt(4)},
			{"sql_if_expr_10", "IF('true', 4, 5)", evaluator.SQLInt(5)},
			{"sql_if_expr_11", "IF(null, 4, 'cat')", evaluator.SQLVarchar("cat")},
			{"sql_if_expr_12", "IF(true, 'dog', 'cat')", evaluator.SQLVarchar("dog")},
			{"sql_if_expr_13", "IF(false, 'dog', 'cat')", evaluator.SQLVarchar("cat")},
			{"sql_if_expr_14", "IF('ca.gh', 4, 5)", evaluator.SQLInt(5)},
			{
				"sql_if_expr_15",
				"IF(current_timestamp(), 4, 5)",
				evaluator.SQLInt(4), // not being parsed as dates, being parsed as string
			},
			{"sql_if_expr_16", "IF(current_timestamp, 4, 5)", evaluator.SQLInt(4)},
			{"sql_if_null_0", "IFNULL(1,0)", evaluator.SQLInt(1)},
			{"sql_if_null_1", "IFNULL(NULL,3)", evaluator.SQLInt(3)},
			{"sql_if_null_2", "IFNULL(NULL,NULL)", evaluator.SQLNull},
			{"sql_if_null_3", "IFNULL('cat', null)", evaluator.SQLVarchar("cat")},
			{"sql_if_null_4", "IFNULL(null, 'dog')", evaluator.SQLVarchar("dog")},
			{"sql_if_null_5", "IFNULL(1/0, 4)", evaluator.SQLInt(4)},
			{"sql_interval_expr_0", "INTERVAL(1,0)", evaluator.SQLInt(1)},
			{"sql_interval_expr_1", "INTERVAL(NULL, 3)", evaluator.SQLInt(-1)},
			{"sql_interval_expr_2", "INTERVAL(NULL, NULL)", evaluator.SQLInt(-1)},
			{"sql_interval_expr_3", "INTERVAL(2, 1, 2, 3, 4)", evaluator.SQLInt(2)},
			{"sql_interval_expr_4", "INTERVAL('1.1', 0, 1.1, 2)", evaluator.SQLInt(2)},
			{"sql_interval_expr_5", "INTERVAL(-1, NULL, 4)", evaluator.SQLInt(1)},
			{"sql_interval_expr_6", "INTERVAL(4, 1, 2, 4)", evaluator.SQLInt(3)},
			{"sql_is_null_0", "ISNULL(a)", evaluator.SQLBool(0)},
			{"sql_is_null_1", "ISNULL(c)", evaluator.SQLBool(1)},
			{"sql_is_null_2", `ISNULL("")`, evaluator.SQLBool(0)},
			{"sql_is_null_3", `ISNULL(NULL)`, evaluator.SQLBool(1)},
			{"sql_insert_expr_0", "INSERT('Quadratic', NULL, 4, 'What')", evaluator.SQLNull},
			{
				"sql_insert_expr_1",
				"INSERT('Quadratic', 3, 4, 'What')",
				evaluator.SQLVarchar("QuWhattic"),
			},
			{
				"sql_insert_expr_2",
				"INSERT('Quadratic', -1, 4, 'What')",
				evaluator.SQLVarchar("Quadratic"),
			},
			{
				"sql_insert_expr_3",
				"INSERT('Quadratic', 3, 100, 'What')",
				evaluator.SQLVarchar("QuWhat"),
			},
			{
				"sql_insert_expr_4",
				"INSERT('Quadratic', 9, 4, 'What')",
				evaluator.SQLVarchar("QuadratiWhat"),
			},
			{
				"sql_insert_expr_5",
				"INSERT('Quadratic', 8.5, 3.5, 'What')",
				evaluator.SQLVarchar("QuadratiWhat"),
			},
			{
				"sql_insert_expr_6",
				"INSERT('Quadratic', 8.4, 3.4, 'What')",
				evaluator.SQLVarchar("QuadratWhat"),
			},
			{"sql_instr_expr_0", "INSTR(NULL, NULL)", evaluator.SQLNull},
			{"sql_instr_expr_1", "INSTR('sDg', 'D')", evaluator.SQLInt(2)},
			{"sql_instr_expr_2", "INSTR(124, 124)", evaluator.SQLInt(1)},
			{"sql_instr_expr_3", "INSTR('awesome','so')", evaluator.SQLInt(4)},
			{"sql_lcase_0", "LCASE(NULL)", evaluator.SQLNull},
			{"sql_lcase_1", "LCASE('sDg')", evaluator.SQLVarchar("sdg")},
			{"sql_lcase_2", "LCASE(124)", evaluator.SQLVarchar("124")},
			{"sql_lowercase_0", "LOWER(NULL)", evaluator.SQLNull},
			{"sql_lowercase_1", "LOWER('')", evaluator.SQLVarchar("")},
			{"sql_lowercase_2", "LOWER('A')", evaluator.SQLVarchar("a")},
			{"sql_lowercase_3", "LOWER('awesome')", evaluator.SQLVarchar("awesome")},
			{"sql_lowercase_4", "LOWER('AwEsOmE')", evaluator.SQLVarchar("awesome")},
			// if any argument null, should return null
			{"sql_left_null_arg_0", "LEFT(NULL, NULL)", evaluator.SQLNull},
			{"sql_left_null_arg_1", "LEFT('hi', NULL)", evaluator.SQLNull},
			{"sql_left_null_arg_2", "LEFT(NULL, 5)", evaluator.SQLNull},
			// basic cases w/ string, int inputs and positive int length
			{"sql_left_base_0", "LEFT('sDgcdcdc', 4)", evaluator.SQLVarchar("sDgc")},
			{"sql_left_base_1", "LEFT(124, 2)", evaluator.SQLVarchar("12")},
			// negative lengths and 0 give empty string
			{"sql_left_negative_0", "LEFT('hi', -1)", evaluator.SQLVarchar("")},
			{"sql_left_negative_1", "LEFT('hi', 0)", evaluator.SQLVarchar("")},
			{"sql_left_negative_2", "LEFT('hi', -2.5)", evaluator.SQLVarchar("")},
			// float lengths should be rounded to closest int
			{"sql_left_float_0", "LEFT('hello', 2.4)", evaluator.SQLVarchar("he")},
			{"sql_left_float_1", "LEFT('hello', 2.5)", evaluator.SQLVarchar("hel")},
			{"sql_left_float_2", "LEFT(1234, 2.3)", evaluator.SQLVarchar("12")},
			{"sql_left_float_3", "LEFT(1234, 2.5)", evaluator.SQLVarchar("123")},
			{"sql_left_float_4", "LEFT('yo', 2.5)", evaluator.SQLVarchar("yo")},
			// strings with spaces and symbols
			{"sql_left_symbols_0", "LEFT('  ', 1)", evaluator.SQLVarchar(" ")},
			{"sql_left_symbols_1", "LEFT('@!%', 2)", evaluator.SQLVarchar("@!")},
			// boolean for string
			{"sql_left_bool_0", "LEFT(true, 3)", evaluator.SQLVarchar("1")},
			{"sql_left_bool_1", "LEFT(false, 3)", evaluator.SQLVarchar("0")},
			// boolean for length
			{"sql_left_bool_2", "LEFT('hello', true)", evaluator.SQLVarchar("h")},
			{"sql_left_bool_3", "LEFT('hello', false)", evaluator.SQLVarchar("")},
			// string for length
			{"sql_left_string", "LEFT('hello', 'hi')", evaluator.SQLVarchar("")},
			// len > length of string
			{"sql_left_edge_case", "LEFT('hi', 5)", evaluator.SQLVarchar("hi")},
			// string number as length
			{"sql_left_string_int_0", "LEFT('hello', '2')", evaluator.SQLVarchar("he")},
			{"sql_left_string_int_1", "LEFT('hello', '-3')", evaluator.SQLVarchar("")},
			// unlike with floats, string #s always round down
			{"sql_left_string_float_0", "LEFT('hello', '2.4')", evaluator.SQLVarchar("he")},
			{"sql_left_string_float_1", "LEFT('hello', '2.6')", evaluator.SQLVarchar("he")},
			{"sql_length_0", "LENGTH(NULL)", evaluator.SQLNull},
			{"sql_length_1", "LENGTH('sDg')", evaluator.SQLInt(3)},
			{"sql_length_2", "LENGTH('世界')", evaluator.SQLInt(6)},
			{"sql_ln_expr_0", "LN(NULL)", evaluator.SQLNull},
			{"sql_ln_expr_1", "LN(1)", evaluator.SQLFloat(0)},
			{"sql_ln_expr_2", "LN(16.5)", evaluator.SQLFloat(2.803360380906535)},
			{"sql_ln_expr_3", "LN(-16.5)", evaluator.SQLNull},
			{"sql_log_expr_0", "LOG(NULL)", evaluator.SQLNull},
			{"sql_log_expr_1", "LOG(1)", evaluator.SQLFloat(0)},
			{"sql_log_expr_2", "LOG(16.5)", evaluator.SQLFloat(2.803360380906535)},
			{"sql_log_expr_3", "LOG(-16.5)", evaluator.SQLNull},
			{"sql_log_expr_4", "LOG10(100)", evaluator.SQLFloat(2)},
			{"sql_log_expr_5", "LOG(10,100)", evaluator.SQLFloat(2)},
			{"sql_locate_0", "LOCATE(NULL, 'foobarbar')", evaluator.SQLNull},
			{"sql_locate_1", "LOCATE('bar', NULL)", evaluator.SQLNull},
			{"sql_locate_2", "LOCATE('bar', 'foobarbar')", evaluator.SQLInt(4)},
			{"sql_locate_3", "LOCATE('xbar', 'foobarbar')", evaluator.SQLInt(0)},
			{"sql_locate_4", "LOCATE('bar', 'foobarbar', 5)", evaluator.SQLInt(7)},
			{"sql_locate_5", "LOCATE('bar', 'foobarbar', 4)", evaluator.SQLInt(4)},
			{"sql_locate_6", "LOCATE('e', 'dvd', 6)", evaluator.SQLInt(0)},
			{"sql_locate_7", "LOCATE('f', 'asdf', 4)", evaluator.SQLInt(4)},
			{"sql_locate_8", "LOCATE('語', '日本語')", evaluator.SQLInt(3)},
			{"sql_log2_0", "LOG2(NULL)", evaluator.SQLNull},
			{"sql_log2_1", "LOG2(4)", evaluator.SQLFloat(2)},
			{"sql_log2_2", "LOG2(-100)", evaluator.SQLNull},
			{"sql_log10_0", "LOG10(NULL)", evaluator.SQLNull},
			{"sql_log10_1", "LOG10('sdg')", evaluator.SQLNull},
			{"sql_log10_2", "LOG10(2)", evaluator.SQLFloat(0.3010299956639812)},
			{"sql_log10_3", "LOG10(100)", evaluator.SQLFloat(2)},
			{"sql_log10_4", "LOG10(0)", evaluator.SQLNull},
			{"sql_log10_5", "LOG10(-100)", evaluator.SQLNull},
			{"sql_ltrim_0", "LTRIM(NULL)", evaluator.SQLNull},
			{"sql_ltrim_1", "LTRIM('   barbar')", evaluator.SQLVarchar("barbar")},
			{"sql_md5_0", "MD5(NULL)", evaluator.SQLNull},
			{"sql_md5_1", "MD5(NULL + NULL)", evaluator.SQLNull},
			{
				"sql_md5_2",
				"MD5('testing')",
				evaluator.SQLVarchar("ae2b1fca515949e5d54fb22b8ed95575"),
			},
			{"sql_md5_3", "MD5('hello')", evaluator.SQLVarchar("5d41402abc4b2a76b9719d911017c592")},
			{"sql_md5_4", "MD5(12)", evaluator.SQLVarchar("c20ad4d76fe97759aa27a0c99bff6710")},
			{"sql_md5_5", "MD5(6.23)", evaluator.SQLVarchar("fec8db978f6b7844b09d9bd54fb8335c")},
			{
				"sql_md5_6",
				"MD5('12:STR.002234')",
				evaluator.SQLVarchar("81d56d5aeb92a55298af2f091e86ab61"),
			},
			{
				"sql_md5_7",
				"MD5(REPEAT('a', 30))",
				evaluator.SQLVarchar("59e794d45697b360e18ba972bada0123"),
			},
			{"sql_microsecond_0", "MICROSECOND(NULL)", evaluator.SQLNull},
			{"sql_microsecond_1", "MICROSECOND('')", evaluator.SQLNull},
			{"sql_microsecond_2", "MICROSECOND('NULL')", evaluator.SQLInt(0)},
			{"sql_microsecond_3", "MICROSECOND('hello')", evaluator.SQLInt(0)},
			{"sql_microsecond_4", "MICROSECOND(TRUE)", evaluator.SQLInt(0)},
			{"sql_microsecond_5", "MICROSECOND('true')", evaluator.SQLInt(0)},
			{"sql_microsecond_6", "MICROSECOND('FALSE')", evaluator.SQLInt(0)},
			{"sql_microsecond_7", "MICROSECOND('11:38:24')", evaluator.SQLInt(0)},
			{"sql_microsecond_8", "MICROSECOND('11:38')", evaluator.SQLInt(0)},
			{"sql_microsecond_9", "MICROSECOND('11 38 24')", evaluator.SQLInt(0)},
			{"sql_microsecond_10", "MICROSECOND('11:38:24.000000')", evaluator.SQLInt(0)},
			{"sql_microsecond_11", "MICROSECOND('11:38:24.000001')", evaluator.SQLInt(1)},
			{"sql_microsecond_12", "MICROSECOND('11:38:24.123456')", evaluator.SQLInt(123456)},
			{"sql_microsecond_13", "MICROSECOND('1978-9-22 1:58:59')", evaluator.SQLInt(0)},
			{"sql_microsecond_14", "MICROSECOND('1978-9-22 1:58:59.00001')", evaluator.SQLInt(10)},
			{
				"sql_microsecond_15",
				"MICROSECOND('1978-9-22 1:58:59.0000104')",
				evaluator.SQLInt(10),
			},
			{"sql_microsecond_16", "MICROSECOND('12:STUFF.002234')", evaluator.SQLInt(0)},
			{"sql_mid_0", "MID('foobarbar', 4, NULL)", evaluator.SQLNull},
			{"sql_mid_1", "MID('Quadratically', 5, 6)", evaluator.SQLVarchar("ratica")},
			{"sql_mid_2", "MID('Quadratically', 12, 2)", evaluator.SQLVarchar("ly")},
			{"sql_mid_3", "MID('Sakila', -5, 3)", evaluator.SQLVarchar("aki")},
			{"sql_mid_4", "MID('日本語', 2, 1)", evaluator.SQLVarchar("本")},
			{"sql_minute_0", "MINUTE(NULL)", evaluator.SQLNull},
			{"sql_minute_1", "MINUTE('sdg')", evaluator.SQLInt(0)},
			{"sql_minute_2", "MINUTE('10:23:52')", evaluator.SQLInt(23)},
			{"sql_minute_3", "MINUTE('10:61:52')", evaluator.SQLNull},
			{"sql_minute_4", "MINUTE('10:23:52.25.26.27.28')", evaluator.SQLInt(23)},
			{"sql_mod_0", "MOD(NULL, NULL)", evaluator.SQLNull},
			{"sql_mod_1", "MOD(234, NULL)", evaluator.SQLNull},
			{"sql_mod_2", "MOD(NULL, 10)", evaluator.SQLNull},
			{"sql_mod_3", "MOD(234, 0)", evaluator.SQLNull},
			{"sql_mod_4", "MOD(234, 10)", evaluator.SQLFloat(4)},
			{"sql_mod_5", "MOD(253, 7)", evaluator.SQLFloat(1)},
			{"sql_mod_6", "MOD(34.5, 3)", evaluator.SQLFloat(1.5)},
			{"sql_month_0", "MONTH(NULL)", evaluator.SQLNull},
			{"sql_month_1", "MONTH('sdg')", evaluator.SQLNull},
			{"sql_month_2", "MONTH('2016-1-01 10:23:52')", evaluator.SQLInt(1)},
			{"sql_month_name_expr_0", "MONTHNAME(NULL)", evaluator.SQLNull},
			{"sql_month_name_expr_1", "MONTHNAME('sdg')", evaluator.SQLNull},
			{
				"sql_month_name_expr_2",
				"MONTHNAME('2016-1-01 10:23:52')",
				evaluator.SQLVarchar("January"),
			},
			{"sql_null_if_0", "NULLIF(1,1)", evaluator.SQLNull},
			{"sql_null_if_1", "NULLIF(1,3)", evaluator.SQLInt(1)},
			{"sql_null_if_2", "NULLIF(null, null)", evaluator.SQLNull},
			{"sql_null_if_3", "NULLIF(null, 4)", evaluator.SQLNull},
			{"sql_null_if_4", "NULLIF(3, null)", evaluator.SQLInt(3)},
			//test{"sql_null_if_5", "NULLIF(3, '3')", evaluator.SQLNull},
			{"sql_null_if_6", "NULLIF('abc', 'abc')", evaluator.SQLNull},
			//test{"sql_null_if_7", "NULLIF('abc', 3)", evaluator.SQLVarchar("abc")},
			//test{"sql_null_if_8", "NULLIF('1', true)", evaluator.SQLNull},
			//test{"sql_null_if_9", "NULLIF('1', false)", evaluator.SQLVarchar("1")},
			{"sql_pi_expr", "PI()", evaluator.SQLFloat(3.141592653589793116)},
			{"sql_quarter_0", "QUARTER(NULL)", evaluator.SQLNull},
			{"sql_quarter_1", "QUARTER('sdg')", evaluator.SQLNull},
			{"sql_quarter_2", "QUARTER('2016-1-01 10:23:52')", evaluator.SQLInt(1)},
			{"sql_quarter_3", "QUARTER('2016-4-01 10:23:52')", evaluator.SQLInt(2)},
			{"sql_quarter_4", "QUARTER('2016-8-01 10:23:52')", evaluator.SQLInt(3)},
			{"sql_quarter_5", "QUARTER('2016-11-01 10:23:52')", evaluator.SQLInt(4)},
			{"sql_radians_0", "RADIANS(NULL)", evaluator.SQLNull},
			{"sql_radians_1", "RADIANS(1145.9155902616465)", evaluator.SQLFloat(20)},
			{"sql_radians_2", "RADIANS(-1145.9155902616465)", evaluator.SQLFloat(-20)},
			{"sql_rand_0", "RAND(NULL)", evaluator.SQLFloat(0.9451961492941164)},
			{"sql_rand_1", "RAND('hello')", evaluator.SQLFloat(0.9451961492941164)},
			{"sql_rand_2", "RAND(0)", evaluator.SQLFloat(0.9451961492941164)},
			{"sql_rand_3", "RAND(1145.9155902616465)", evaluator.SQLFloat(0.16758646518059656)},
			{"sql_rand_4", "RAND(-1145.9155902616465)", evaluator.SQLFloat(0.8321372077808122)},
			{"sql_repeat_0", "REPEAT(NULL, NULL)", evaluator.SQLNull},
			{"sql_repeat_1", "REPEAT(NULL, 3)", evaluator.SQLNull},
			{"sql_repeat_2", "REPEAT('apples', NULL)", evaluator.SQLNull},
			{"sql_repeat_3", "REPEAT('apples', -1)", evaluator.SQLVarchar("")},
			{"sql_repeat_4", "REPEAT('apples', 0)", evaluator.SQLVarchar("")},
			{"sql_repeat_5", "REPEAT('apples', 1)", evaluator.SQLVarchar("apples")},
			{"sql_repeat_6", "REPEAT('a', 5)", evaluator.SQLVarchar("aaaaa")},
			{"sql_repeat_7", "REPEAT(3, 5)", evaluator.SQLVarchar("33333")},
			{"sql_repeat_8", "REPEAT(FALSE, 5)", evaluator.SQLVarchar("00000")},
			{"sql_repeat_9", "REPEAT(FALSE, TRUE)", evaluator.SQLVarchar("0")},
			{"sql_repeat_10", "REPEAT('', 10)", evaluator.SQLVarchar("")},
			{"sql_repeat_11", "REPEAT(0, '4')", evaluator.SQLVarchar("0000")},
			{"sql_repeat_12", "REPEAT(NULL, 4)", evaluator.SQLNull},
			{"sql_repeat_13", "REPEAT(1.4, 3)", evaluator.SQLVarchar("1.41.41.4")},
			{"sql_repeat_14", "REPEAT('a', .3)", evaluator.SQLVarchar("")},
			{"sql_repeat_15", "REPEAT('a', 3.2)", evaluator.SQLVarchar("aaa")},
			{"sql_repeat_16", "REPEAT('a', 3.6)", evaluator.SQLVarchar("aaaa")},
			{"sql_replace_0", "REPLACE(NULL, NULL, NULL)", evaluator.SQLNull},
			{"sql_replace_1", "REPLACE('sDgcdcdc', 'D', 'd')", evaluator.SQLVarchar("sdgcdcdc")},
			{
				"sql_replace_2",
				"REPLACE('www.mysql.com', 'w', 'Ww')",
				evaluator.SQLVarchar("WwWwWw.mysql.com"),
			},
			{"sql_reverse_0", "REVERSE(NULL)", evaluator.SQLNull},
			{"sql_reverse_1", "REVERSE(3.14159265)", evaluator.SQLVarchar("56295141.3")},
			{"sql_reverse_2", "REVERSE(655)", evaluator.SQLVarchar("556")},
			{"sql_reverse_3", "REVERSE('www.mysql.com')", evaluator.SQLVarchar("moc.lqsym.www")},
			// if any argument null, should return null
			{"sql_right_null_0", "RIGHT(NULL, NULL)", evaluator.SQLNull},
			{"sql_right_null_1", "RIGHT('hi', NULL)", evaluator.SQLNull},
			{"sql_right_null_2", "RIGHT(NULL, 5)", evaluator.SQLNull},
			// basic cases w/ string, int inputs and positive int length
			{"sql_right_base_case_0", "RIGHT('sDgcdcdc', 4)", evaluator.SQLVarchar("dcdc")},
			{"sql_right_base_case_1", "RIGHT(124, 2)", evaluator.SQLVarchar("24")},
			// negative lengths and 0 give empty string
			{"sql_right_negative_0", "RIGHT('hi', -1)", evaluator.SQLVarchar("")},
			{"sql_right_negative_1", "RIGHT('hi', 0)", evaluator.SQLVarchar("")},
			{"sql_right_negative_2", "RIGHT('hi', -2.5)", evaluator.SQLVarchar("")},
			// float lengths should be rounded to closest int
			{"sql_right_float_0", "RIGHT('hello', 2.4)", evaluator.SQLVarchar("lo")},
			{"sql_right_float_1", "RIGHT('hello', 2.5)", evaluator.SQLVarchar("llo")},
			{"sql_right_float_2", "RIGHT(1234, 2.3)", evaluator.SQLVarchar("34")},
			{"sql_right_float_3", "RIGHT(1234, 2.5)", evaluator.SQLVarchar("234")},
			{"sql_right_float_4", "RIGHT('yo', 2.5)", evaluator.SQLVarchar("yo")},
			// strings with spaces and symbols
			{"sql_right_symbols_0", "RIGHT('  ', 1)", evaluator.SQLVarchar(" ")},
			{"sql_right_symbols_1", "RIGHT('@!%', 2)", evaluator.SQLVarchar("!%")},
			// boolean for string
			{"sql_right_bool_0", "RIGHT(true, 3)", evaluator.SQLVarchar("1")},
			{"sql_right_bool_1", "RIGHT(false, 3)", evaluator.SQLVarchar("0")},
			// boolean for length
			{"sql_right_bool_length_0", "RIGHT('hello', true)", evaluator.SQLVarchar("o")},
			{"sql_right_bool_length_1", "RIGHT('hello', false)", evaluator.SQLVarchar("")},
			// string for length
			{"sql_right_string_length", "RIGHT('hello', 'hi')", evaluator.SQLVarchar("")},
			// len > length of string
			{"sql_right_edge", "RIGHT('hi', 5)", evaluator.SQLVarchar("hi")},
			// string number as length
			{"sql_right_num_as_length_0", "RIGHT('hello', '2')", evaluator.SQLVarchar("lo")},
			{"sql_right_num_as_length_1", "RIGHT('hello', '-3')", evaluator.SQLVarchar("")},
			// unlike with floats, string #s always round down
			{"sql_right_float_as_length_0", "RIGHT('hello', '2.4')", evaluator.SQLVarchar("lo")},
			{"sql_right_float_as_length_1", "RIGHT('hello', '2.6')", evaluator.SQLVarchar("lo")},
			{"sql_round_0", "ROUND(NULL, NULL)", evaluator.SQLNull},
			{"sql_round_1", "ROUND(NULL, 4)", evaluator.SQLNull},
			{"sql_round_2", "ROUND(-16.55555, 4)", evaluator.SQLFloat(-16.5556)},
			{"sql_round_3", "ROUND(4.56, 1)", evaluator.SQLFloat(4.6)},
			{"sql_round_4", "ROUND(-16.5, -1)", evaluator.SQLFloat(0)},
			{"sql_round_5", "ROUND(-16.5)", evaluator.SQLFloat(-17)},
			{"sql_rtrim_0", "RTRIM(NULL)", evaluator.SQLNull},
			{"sql_rtrim_1", "RTRIM('barbar   ')", evaluator.SQLVarchar("barbar")},
			// LPAD(str, len, padStr)
			// basic case
			{"sql_lpad_0", "LPAD('hello', 7, 'x')", evaluator.SQLVarchar("xxhello")},
			// nulls in various positions
			{"sql_lpad_null_0", "LPAD(NULL, 5, 'a')", evaluator.SQLNull},
			{"sql_lpad_null_1", "LPAD('hi', NULL, 'a')", evaluator.SQLNull},
			{"sql_lpad_null_2", "LPAD('hi', 5, NULL)", evaluator.SQLNull},
			{"sql_lpad_null_3", "LPAD(NULL, NULL, NULL)", evaluator.SQLNull},
			// str: empty
			{"sql_lpad_empty_0", "LPAD('', 0, 'a')", evaluator.SQLVarchar("")},
			{"sql_lpad_empty_1", "LPAD('', 1, 'a')", evaluator.SQLVarchar("a")},
			{"sql_lpad_empty_2", "LPAD('', 7, 'ab')", evaluator.SQLVarchar("abababa")},
			// str: spaces and symbols
			{"sql_lpad_symbols_0", "LPAD(' hi', 4, 'x')", evaluator.SQLVarchar("x hi")},
			{"sql_lpad_symbols_1", "LPAD('  ', 5, ' ')", evaluator.SQLVarchar("     ")},
			{"sql_lpad_symbols_2", "LPAD('@!#_', 10, '.')", evaluator.SQLVarchar("......@!#_")},
			{"sql_lpad_symbols_3", "LPAD('I♥NY', 8, 'x')", evaluator.SQLVarchar("xxxxI♥NY")},
			{"sql_lpad_symbols_4", "LPAD('ƏŨ Ó€', 8, 'x')", evaluator.SQLVarchar("xxxƏŨ Ó€")},
			{
				"sql_lpad_symbols_5",
				"LPAD('⅓ ⅔ † ‡ µ ¢ £', 8, 'x')",
				evaluator.SQLVarchar("⅓ ⅔ † ‡ "),
			},
			{"sql_lpad_symbols_6", "LPAD('∞π∅≤≥≠≈', 8, 'x')", evaluator.SQLVarchar("x∞π∅≤≥≠≈")},
			{"sql_lpad_symbols_7", "LPAD('hello', 8, '♥')", evaluator.SQLVarchar("♥♥♥hello")},
			{"sql_lpad_symbols_8", "LPAD('hello', 8, 'ƏŨ')", evaluator.SQLVarchar("ƏŨƏhello")},
			// str type: numbers
			{"sql_lpad_numbers_0", "LPAD(5, 4, 'a')", evaluator.SQLVarchar("aaa5")},
			{"sql_lpad_numbers_1", "LPAD(10, 4, 'a')", evaluator.SQLVarchar("aa10")},
			{"sql_lpad_numbers_2", "LPAD(10.2, 4, 'a')", evaluator.SQLVarchar("10.2")},
			// str type: boolean
			{"sql_lpad_bool_0", "LPAD(true, 4, 'a')", evaluator.SQLVarchar("aaa1")},
			{"sql_lpad_bool_1", "LPAD(false, 4, 'a')", evaluator.SQLVarchar("aaa0")},
			// len < 0
			{"sql_lpad_neg_length", "LPAD('hi', -1, 'a')", evaluator.SQLNull},
			// len = 0
			{"sql_lpad_zero", "LPAD('hi', 0, 'a')", evaluator.SQLVarchar("")},
			// len <= len(str)
			{"sql_lpad_edge_0", "LPAD('hello', 2, 'x')", evaluator.SQLVarchar("he")},
			{"sql_lpad_edge_1", "LPAD('hello', 5, 'x')", evaluator.SQLVarchar("hello")},
			// len type: str
			{"sql_lpad_edge_2", "LPAD('hello', '5', 'x')", evaluator.SQLVarchar("hello")},
			{"sql_lpad_edge_3", "LPAD('hello', '5.6', 'x')", evaluator.SQLVarchar("hello")},
			{"sql_lpad_edge_3", "LPAD('hello', '6', 'x')", evaluator.SQLVarchar("xhello")},
			{"sql_lpad_edge_4", "LPAD('hello', '6.2', 'x')", evaluator.SQLVarchar("xhello")},
			// if can't be cast to #, then use length 0
			{"sql_lpad_edge_5", "LPAD('hello', 'a', 'x')", evaluator.SQLVarchar("")},
			// len: floating point
			{"sql_lpad_edge_6", "LPAD('hello', 5.4, 'x')", evaluator.SQLVarchar("hello")},
			{"sql_lpad_edge_7", "LPAD('hello', 5.5, 'x')", evaluator.SQLVarchar("xhello")},
			// len float values close to 0 - round to closest int
			{"sql_lpad_edge_8", "LPAD('hello', 0.4, 'x')", evaluator.SQLVarchar("")},
			{"sql_lpad_edge_9", "LPAD('hello', 0.5, 'x')", evaluator.SQLVarchar("h")},
			{"sql_lpad_edge_10", "LPAD('hello', -0.4, 'x')", evaluator.SQLVarchar("")},
			{"sql_lpad_edge_11", "LPAD('hello', -0.5, 'x')", evaluator.SQLNull},
			// len string values close to 0 - always round toward 0
			{"sql_lpad_edge_12", "LPAD('hello', '0.4', 'x')", evaluator.SQLVarchar("")},
			{"sql_lpad_edge_13", "LPAD('hello', '0.5', 'x')", evaluator.SQLVarchar("")},
			{"sql_lpad_edge_14", "LPAD('hello', '-0.4', 'x')", evaluator.SQLVarchar("")},
			{"sql_lpad_edge_15", "LPAD('hello', '-0.5', 'x')", evaluator.SQLVarchar("")},
			// len: bool
			{"sql_lpad_edge_16", "LPAD('hello', true, 'x')", evaluator.SQLVarchar("h")},
			{"sql_lpad_edge_17", "LPAD('hello', false, 'x')", evaluator.SQLVarchar("")},
			// len(padStr) > 1
			{"sql_lpad_edge_18", "LPAD('hello', 7, 'xy')", evaluator.SQLVarchar("xyhello")},
			{"sql_lpad_edge_19", "LPAD('hello', 8, 'xy')", evaluator.SQLVarchar("xyxhello")},
			// padStr type: number
			{"sql_lpad_edge_20", "LPAD('hello', 7, 1)", evaluator.SQLVarchar("11hello")},
			{"sql_lpad_edge_21", "LPAD('hello', 10, 1.1)", evaluator.SQLVarchar("1.11.hello")},
			{"sql_lpad_edge_22", "LPAD('hello', 10, -1)", evaluator.SQLVarchar("-1-1-hello")},

			// padStr type: boolean
			{"sql_lpad_edge_23", "LPAD('hello', 7, true)", evaluator.SQLVarchar("11hello")},
			{"sql_lpad_edge_24", "LPAD('hello', 10, false)", evaluator.SQLVarchar("00000hello")},
			// RPAD(str, len, padStr)
			// basic case
			{"sql_rpad_0", "RPAD('hello', 7, 'x')", evaluator.SQLVarchar("helloxx")},
			// nulls in various positions
			{"sql_rpad_null_0", "RPAD(NULL, 5, 'a')", evaluator.SQLNull},
			{"sql_rpad_null_1", "RPAD('hi', NULL, 'a')", evaluator.SQLNull},
			{"sql_rpad_null_2", "RPAD('hi', 5, NULL)", evaluator.SQLNull},
			{"sql_rpad_null_3", "RPAD(NULL, NULL, NULL)", evaluator.SQLNull},
			// str: empty
			{"sql_rpad_str_empty_0", "RPAD('', 0, 'a')", evaluator.SQLVarchar("")},
			{"sql_rpad_str_empty_1", "RPAD('', 1, 'a')", evaluator.SQLVarchar("a")},
			{"sql_rpad_str_empty_2", "RPAD('', 7, 'ab')", evaluator.SQLVarchar("abababa")},
			// str: spaces and symbols
			{"sql_rpad_symbols_0", "RPAD(' hi', 4, 'x')", evaluator.SQLVarchar(" hix")},
			{"sql_rpad_symbols_1", "RPAD('  ', 5, ' ')", evaluator.SQLVarchar("     ")},
			{"sql_rpad_symbols_2", "RPAD('@!#_', 10, '.')", evaluator.SQLVarchar("@!#_......")},
			{"sql_rpad_symbols_3", "RPAD('I♥NY', 8, 'x')", evaluator.SQLVarchar("I♥NYxxxx")},
			{"sql_rpad_symbols_4", "RPAD('ƏŨ Ó€', 8, 'x')", evaluator.SQLVarchar("ƏŨ Ó€xxx")},
			{
				"sql_rpad_symbols_5",
				"RPAD('⅓ ⅔ † ‡ µ ¢ £', 8, 'x')",
				evaluator.SQLVarchar("⅓ ⅔ † ‡ "),
			},
			{"sql_rpad_symbols_6", "RPAD('∞π∅≤≥≠≈', 8, 'x')", evaluator.SQLVarchar("∞π∅≤≥≠≈x")},
			{"sql_rpad_symbols_7", "RPAD('hello', 8, '♥')", evaluator.SQLVarchar("hello♥♥♥")},
			{"sql_rpad_symbols_8", "RPAD('hello', 8, 'ƏŨ')", evaluator.SQLVarchar("helloƏŨƏ")},
			// str type: numbers
			{"sql_rpad_numbers_0", "RPAD(5, 4, 'a')", evaluator.SQLVarchar("5aaa")},
			{"sql_rpad_numbers_1", "RPAD(10, 4, 'a')", evaluator.SQLVarchar("10aa")},
			{"sql_rpad_numbers_2", "RPAD(10.2, 4, 'a')", evaluator.SQLVarchar("10.2")},
			// str type: boolean
			{"sql_rpad_bool_0", "RPAD(true, 4, 'a')", evaluator.SQLVarchar("1aaa")},
			{"sql_rpad_bool_1", "RPAD(false, 4, 'a')", evaluator.SQLVarchar("0aaa")},
			// len < 0
			{"sql_rpad_len", "RPAD('hi', -1, 'a')", evaluator.SQLNull},
			// len = 0
			{"sql_rpad_len_1", "RPAD('hi', 0, 'a')", evaluator.SQLVarchar("")},
			// len <= len(str)
			{"sql_rpad_len_2", "RPAD('hello', 2, 'x')", evaluator.SQLVarchar("he")},
			{"sql_rpad_len_3", "RPAD('hello', 5, 'x')", evaluator.SQLVarchar("hello")},
			// len type: str
			{"sql_rpad_len_4", "RPAD('hello', '5', 'x')", evaluator.SQLVarchar("hello")},
			{"sql_rpad_len_5", "RPAD('hello', '5.6', 'x')", evaluator.SQLVarchar("hello")},
			{"sql_rpad_len_6", "RPAD('hello', '6', 'x')", evaluator.SQLVarchar("hellox")},
			{"sql_rpad_len_7", "RPAD('hello', '6.2', 'x')", evaluator.SQLVarchar("hellox")},
			// if can't be cast to #, then use length 0
			{"sql_rpad_len_8", "RPAD('hello', 'a', 'x')", evaluator.SQLVarchar("")},
			// len: floating point
			{"sql_rpad_len_9", "RPAD('hello', 5.4, 'x')", evaluator.SQLVarchar("hello")},
			{"sql_rpad_len_10", "RPAD('hello', 5.5, 'x')", evaluator.SQLVarchar("hellox")},
			// len float values close to 0 - round to closest int
			{"sql_rpad_len_11", "RPAD('hello', 0.4, 'x')", evaluator.SQLVarchar("")},
			{"sql_rpad_len_12", "RPAD('hello', 0.5, 'x')", evaluator.SQLVarchar("h")},
			{"sql_rpad_len_13", "RPAD('hello', -0.4, 'x')", evaluator.SQLVarchar("")},
			{"sql_rpad_len_14", "RPAD('hello', -0.5, 'x')", evaluator.SQLNull},
			// len string values close to 0 - always round toward 0
			{"sql_rpad_len_15", "RPAD('hello', '0.4', 'x')", evaluator.SQLVarchar("")},
			{"sql_rpad_len_16", "RPAD('hello', '0.5', 'x')", evaluator.SQLVarchar("")},
			{"sql_rpad_len_17", "RPAD('hello', '-0.4', 'x')", evaluator.SQLVarchar("")},
			{"sql_rpad_len_18", "RPAD('hello', '-0.5', 'x')", evaluator.SQLVarchar("")},
			// len: bool
			{"sql_rpad_len_19", "RPAD('hello', true, 'x')", evaluator.SQLVarchar("h")},
			{"sql_rpad_len_20", "RPAD('hello', false, 'x')", evaluator.SQLVarchar("")},
			// len(padStr) > 1
			{"sql_rpad_len_21", "RPAD('hello', 7, 'xy')", evaluator.SQLVarchar("helloxy")},
			{"sql_rpad_len_22", "RPAD('hello', 8, 'xy')", evaluator.SQLVarchar("helloxyx")},
			// padStr type: number
			{"sql_rpad_len_23", "RPAD('hello', 7, 1)", evaluator.SQLVarchar("hello11")},
			{"sql_rpad_len_24", "RPAD('hello', 10, 1.1)", evaluator.SQLVarchar("hello1.11.")},
			{"sql_rpad_len_25", "RPAD('hello', 10, -1)", evaluator.SQLVarchar("hello-1-1-")},
			// padStr type: boolean
			{"sql_rpad_len_26", "RPAD('hello', 7, true)", evaluator.SQLVarchar("hello11")},
			{"sql_rpad_len_27", "RPAD('hello', 10, false)", evaluator.SQLVarchar("hello00000")},
			{"sql_second_0", "SECOND(NULL)", evaluator.SQLNull},
			{"sql_second_1", "SECOND('sdg')", evaluator.SQLInt(0)},
			{"sql_second_2", "SECOND('10:23:52')", evaluator.SQLInt(52)},
			{"sql_second_3", "SECOND('10:61:52.24')", evaluator.SQLNull},
			{"sql_second_4", "SECOND('10:23:52.24.25.26.27')", evaluator.SQLInt(52)},
			{"sql_sign_0", "SIGN(NULL)", evaluator.SQLNull},
			{"sql_sign_1", "SIGN(-42)", evaluator.SQLInt(-1)},
			{"sql_sign_2", "SIGN(0)", evaluator.SQLInt(0)},
			{"sql_sign_3", "SIGN(42)", evaluator.SQLInt(1)},
			{"sql_sign_4", "SIGN(42.0)", evaluator.SQLInt(1)},
			{"sql_sign_5", "SIGN(-42.0)", evaluator.SQLInt(-1)},
			{"sql_sign_6", "SIGN('hello world')", evaluator.SQLInt(0)},
			{"sql_sin_0", "SIN(NULL)", evaluator.SQLNull},
			{"sql_sin_1", "SIN(19)", evaluator.SQLFloat(0.14987720966295234)},
			{"sql_sin_2", "SIN(-19)", evaluator.SQLFloat(-0.14987720966295234)},
			{"sql_sin_3", "SIN('C')", evaluator.SQLFloat(0)},
			{"sql_sin_4", "SIN(0)", evaluator.SQLFloat(0)},
			{"sql_space_0", "SPACE(NULL)", evaluator.SQLNull},
			{"sql_space_1", "SPACE(5)", evaluator.SQLVarchar("     ")},
			{"sql_space_2", "SPACE(-3)", evaluator.SQLVarchar("")},
			{"sql_sort_0", "SQRT(NULL)", evaluator.SQLNull},
			{"sql_sort_1", "SQRT('sdg')", evaluator.SQLFloat(0)},
			{"sql_sort_2", "SQRT(-16)", evaluator.SQLNull},
			{"sql_sort_3", "SQRT(4)", evaluator.SQLFloat(2)},
			{"sql_sort_4", "SQRT(20)", evaluator.SQLFloat(4.47213595499958)},
			{"sql_substring_0", "SUBSTRING(NULL, 4)", evaluator.SQLNull},
			{"sql_substring_1", "SUBSTRING('foobarbar', NULL)", evaluator.SQLNull},
			{"sql_substring_2", "SUBSTRING('foobarbar', 4, NULL)", evaluator.SQLNull},
			{"sql_substring_3", "SUBSTRING('Quadratically', 5)", evaluator.SQLVarchar("ratically")},
			{"sql_substring_4", "SUBSTRING('Quadratically', 5, 6)", evaluator.SQLVarchar("ratica")},
			{"sql_substring_5", "SUBSTRING('Quadratically', 12, 2)", evaluator.SQLVarchar("ly")},
			{"sql_substring_6", "SUBSTRING('Sakila', -3)", evaluator.SQLVarchar("ila")},
			{"sql_substring_7", "SUBSTRING('Sakila', -5, 3)", evaluator.SQLVarchar("aki")},
			{"sql_substring_8", "SUBSTRING('日本語', 2)", evaluator.SQLVarchar("本語")},
			{"sql_substring_9", "SUBSTR(NULL, 4)", evaluator.SQLNull},
			{"sql_substring_10", "SUBSTR('foobarbar', NULL)", evaluator.SQLNull},
			{"sql_substring_11", "SUBSTR('foobarbar', 4, NULL)", evaluator.SQLNull},
			{"sql_substring_12", "SUBSTR('Quadratically', 5)", evaluator.SQLVarchar("ratically")},
			{"sql_substring_13", "SUBSTR('Quadratically', 5, 6)", evaluator.SQLVarchar("ratica")},
			{"sql_substring_14", "SUBSTR('Sakila', -3)", evaluator.SQLVarchar("ila")},
			{"sql_substring_15", "SUBSTR('Sakila', -5, 3)", evaluator.SQLVarchar("aki")},
			{"sql_substring_16", "SUBSTR('日本語', 2)", evaluator.SQLVarchar("本語")},
			{"sql_substring_17", "SUBSTR('five', 2, 2)", evaluator.SQLVarchar("iv")},
			{"sql_substring_18", "SUBSTR('nine', 4, 9)", evaluator.SQLVarchar("e")},
			{"sql_substring_19", "SUBSTR('five', 4, 3)", evaluator.SQLVarchar("e")},
			{"sql_substring_20", "SUBSTR('five', -1, 1)", evaluator.SQLVarchar("e")},
			{"sql_substring_21", "SUBSTR('five', 4, 0)", evaluator.SQLVarchar("")},
			{"sql_substring_22", "SUBSTR('ZBA', 0)", evaluator.SQLVarchar("")},
			{"sql_substring_23", "SUBSTR('ZBA', 0, 1)", evaluator.SQLVarchar("")},
			{"sql_substring_24", "SUBSTR('ZBA', 0, -1)", evaluator.SQLVarchar("")},
			{"sql_substring_25", "SUBSTR('ZBA', -1, 0)", evaluator.SQLVarchar("")},
			{"sql_substring_26", "SUBSTR('ZBA', 1, 0)", evaluator.SQLVarchar("")},
			{"sql_substring_27", "SUBSTR('ZBA', 0, 0)", evaluator.SQLVarchar("")},
			{"sql_substring_28", "SUBSTRING(NULL from 4)", evaluator.SQLNull},
			{"sql_substring_29", "SUBSTRING('foobarbar' from NULL)", evaluator.SQLNull},
			{"sql_substring_30", "SUBSTRING('foobarbar' from 4 for NULL)", evaluator.SQLNull},
			{
				"sql_substring_31",
				"SUBSTRING('Quadratically' FROM 5)",
				evaluator.SQLVarchar("ratically"),
			},
			{
				"sql_substring_32",
				"SUBSTRING('Quadratically' FROM  5 for 6)",
				evaluator.SQLVarchar("ratica"),
			},
			{
				"sql_substring_33",
				"SUBSTRING('Quadratically' from 12 FOR 2)",
				evaluator.SQLVarchar("ly"),
			},
			{"sql_substring_34", "SUBSTRING('Sakila' FROM -3)", evaluator.SQLVarchar("ila")},
			{"sql_substring_35", "SUBSTRING('Sakila' from -5 for 3)", evaluator.SQLVarchar("aki")},
			{"sql_substring_36", "SUBSTRING('日本語' FROM  2)", evaluator.SQLVarchar("本語")},
			{"sql_substring_37", "SUBSTR(NULL FROM 4)", evaluator.SQLNull},
			{"sql_substring_38", "SUBSTR('foobarbar' FROM NULL)", evaluator.SQLNull},
			{"sql_substring_39", "SUBSTR('foobarbar' FROM 4 FOR NULL)", evaluator.SQLNull},
			{
				"sql_substring_40",
				"SUBSTR('Quadratically' FROM  5)",
				evaluator.SQLVarchar("ratically"),
			},
			{
				"sql_substring_41",
				"SUBSTR('Quadratically' FROM  5 for 6)",
				evaluator.SQLVarchar("ratica"),
			},
			{"sql_substring_42", "SUBSTR('Sakila' from -3)", evaluator.SQLVarchar("ila")},
			{"sql_substring_43", "SUBSTR('Sakila' from -5 for 3)", evaluator.SQLVarchar("aki")},
			{"sql_substring_44", "SUBSTR('日本語' from 2)", evaluator.SQLVarchar("本語")},
			{"sql_substring_45", "SUBSTR('five' from 2 for 2)", evaluator.SQLVarchar("iv")},
			{"sql_substring_46", "SUBSTR('nine' from 4 for  9)", evaluator.SQLVarchar("e")},
			{"sql_substring_47", "SUBSTR('five' FROM 4 FOR 3)", evaluator.SQLVarchar("e")},
			{"sql_substring_48", "SUBSTR('five' FROM -1 FOR  1)", evaluator.SQLVarchar("e")},
			{"sql_substring_49", "SUBSTR('five' FROM 4 FOR  0)", evaluator.SQLVarchar("")},
			{"sql_substring_50", "SUBSTR('ZBA' FROM 0)", evaluator.SQLVarchar("")},
			{"sql_substring_51", "SUBSTR('ZBA' FROM 0 FOR  1)", evaluator.SQLVarchar("")},
			{"sql_substring_52", "SUBSTR('ZBA' FROM 0 for  -1)", evaluator.SQLVarchar("")},
			{"sql_substring_53", "SUBSTR('ZBA' from -1 for  0)", evaluator.SQLVarchar("")},
			{"sql_substring_54", "SUBSTR('ZBA' from 1 FOR 0)", evaluator.SQLVarchar("")},
			{"sql_substring_55", "SUBSTR('ZBA' from 0 for 0)", evaluator.SQLVarchar("")},
			{"sql_substring_56", "SUBSTR('this', -5.2)", evaluator.SQLVarchar("")},
			{"sql_substring_57", "SUBSTR('this' from -5.2)", evaluator.SQLVarchar("")},
			{"sql_substring_58", "SUBSTR('this', 2.632)", evaluator.SQLVarchar("is")},
			{"sql_substring_59", "SUBSTR('this', '2.632')", evaluator.SQLVarchar("his")},
			{"sql_substring_60", "SUBSTR('this', '2.1')", evaluator.SQLVarchar("his")},
			{"sql_substring_61", "SUBSTR('this' from -2.632)", evaluator.SQLVarchar("his")},
			{"sql_substring_62", "SUBSTR('this', 2.4, 1.4)", evaluator.SQLVarchar("h")},
			{"sql_substring_63", "SUBSTR('this' from 2.4 for -1.4 )", evaluator.SQLVarchar("")},
			{"sql_substring_64", "SUBSTR('this', 1.6, 2.6)", evaluator.SQLVarchar("his")},
			{"sql_substring_65", "SUBSTR('this', 1.6, '2.6')", evaluator.SQLVarchar("hi")},
			{"sql_substring_66", "SUBSTR('this', 1.6, '2.1')", evaluator.SQLVarchar("hi")},
			{"sql_substring_67", "SUBSTR('this', -11.6)", evaluator.SQLVarchar("")},
			{"sql_substring_68", "SUBSTR(NULL, -4)", evaluator.SQLNull},
			{"sql_substring_69", "SUBSTR(NULL, -4, 2)", evaluator.SQLNull},
			{"sql_substring_70", "SUBSTR('this' FROM NULL FOR 2)", evaluator.SQLNull},
			{"sql_substring_71", "SUBSTR('this', 2, NULL )", evaluator.SQLNull},
			{"sql_substring_72", "SUBSTR('this' FROM 3 FOR NULL)", evaluator.SQLNull},
			{
				"sql_substring_index_0",
				"SUBSTRING_INDEX('www.cmysql.com', '.', NULL)",
				evaluator.SQLNull,
			},
			{
				"sql_substring_index_1",
				"SUBSTRING_INDEX('www.cmysql.com', '.', 0)",
				evaluator.SQLVarchar(""),
			},
			{
				"sql_substring_index_2",
				"SUBSTRING_INDEX('www.cmysql.com', '.', 1)",
				evaluator.SQLVarchar("www"),
			},
			{
				"sql_substring_index_3",
				"SUBSTRING_INDEX('www.cmysql.com', '.c', 1)",
				evaluator.SQLVarchar("www"),
			},
			{
				"sql_substring_index_4",
				"SUBSTRING_INDEX('www.cmysql.com', '.', 2)",
				evaluator.SQLVarchar("www.cmysql"),
			},
			{
				"sql_substring_index_5",
				"SUBSTRING_INDEX('www.cmysql.com', '.', 1000)",
				evaluator.SQLVarchar("www.cmysql.com"),
			},
			{
				"sql_substring_index_6",
				"SUBSTRING_INDEX('www.cmysql.com', '.c', 2)",
				evaluator.SQLVarchar("www.cmysql"),
			},
			{
				"sql_substring_index_7",
				"SUBSTRING_INDEX('www.cmysql.com', '.', -2)",
				evaluator.SQLVarchar("cmysql.com"),
			},
			{
				"sql_substring_index_8",
				"SUBSTRING_INDEX('www.cmysql.com', '.', -1)",
				evaluator.SQLVarchar("com"),
			},
			{"sql_tan_0", "TAN(NULL)", evaluator.SQLNull},
			{"sql_tan_1", "TAN(19)", evaluator.SQLFloat(0.15158947061240008)},
			{"sql_tan_2", "TAN(-19)", evaluator.SQLFloat(-0.15158947061240008)},
			{"sql_tan_3", "TAN('C')", evaluator.SQLFloat(0)},
			{"sql_tan_4", "TAN(0)", evaluator.SQLFloat(0)},
			{"sql_time_to_sec_0", "TIME_TO_SEC(NULL)", evaluator.SQLNull},
			{"sql_time_to_sec_1", "TIME_TO_SEC('22:23:00')", evaluator.SQLFloat(80580)},
			{"sql_time_to_sec_2", "TIME_TO_SEC('12:34')", evaluator.SQLFloat(45240)},
			{"sql_time_to_sec_3", "TIME_TO_SEC('00:39:38')", evaluator.SQLFloat(2378)},
			{"sql_time_to_sec_4", "TIME_TO_SEC(1010103)", evaluator.SQLFloat(363663)},
			{"sql_time_to_sec_5", "TIME_TO_SEC('2222')", evaluator.SQLFloat(1342)},
			{"sql_time_to_sec_6", "TIME_TO_SEC(101010)", evaluator.SQLFloat(36610)},
			{"sql_time_to_sec_7", "TIME_TO_SEC(-222)", evaluator.SQLFloat(-142)},
			{"sql_time_to_sec_8", "TIME_TO_SEC('-22:33:32')", evaluator.SQLFloat(-81212)},
			{"sql_time_to_sec_9", "TIME_TO_SEC(535911)", evaluator.SQLFloat(194351)},
			{"sql_time_to_sec_10", "TIME_TO_SEC('-850:00:00')", evaluator.SQLFloat(-3020399)},
			{"sql_time_to_sec_11", "TIME_TO_SEC('-838:59:59')", evaluator.SQLFloat(-3020399)},
			{
				"sql_time_to_sec_12",
				"TIME_TO_SEC(CONCAT('48:2','4:59'))",
				evaluator.SQLFloat(174299),
			},
			{"sql_time_to_sec_13", "TIME_TO_SEC(535959.9)", evaluator.SQLFloat(194399)},
			{"sql_time_to_sec_14", "TIME_TO_SEC(534422333)", evaluator.SQLNull},
			{"sql_time_to_sec_15", "TIME_TO_SEC(539911)", evaluator.SQLNull},
			{"sql_time_to_sec_16", "TIME_TO_SEC(8991111)", evaluator.SQLNull},
			{"sql_time_to_sec_17", "TIME_TO_SEC('-5359:11')", evaluator.SQLFloat(-3020399)},
			{"sql_time_to_sec_18", "TIME_TO_SEC('2004-07-09 10:17:35')", evaluator.SQLFloat(37055)},
			{
				"sql_time_to_sec_19",
				"TIME_TO_SEC('2004-07-09 10:17:35.238238')",
				evaluator.SQLFloat(37055),
			},
			{"sql_timediff_0", "TIMEDIFF('2000:11:11 00:00:00', NULL)", evaluator.SQLNull},
			{"sql_timediff_1", "TIMEDIFF(NULL, '2000:11:11 00:00:00')", evaluator.SQLNull},
			{
				"sql_timediff_2",
				"TIMEDIFF('2000:09:11 00:00:00', '2000:09:31 00:00:01:323211')",
				evaluator.SQLNull,
			},
			{
				"sql_timediff_3",
				"TIMEDIFF('2008-12-31 23:59:59.000001','2008-12-31 23:59:58.000001')",
				evaluator.SQLVarchar("00:00:01"),
			},
			{
				"sql_timediff_4",
				"TIMEDIFF('2000:11:11 00:00:00', '2000:11:11 10:00:00.000231')",
				evaluator.SQLVarchar("-10:00:00.000231"),
			},
			{
				"sql_timediff_5",
				"TIMEDIFF('2000:01:01 00:00:00','2000:01:01 00:00:00.000001')",
				evaluator.SQLVarchar("-00:00:00.000001"),
			},
			{
				"sql_timediff_6",
				"TIMEDIFF('2008-12-31 23:59:59.000001','2008-12-30 01:01:01.000002')",
				evaluator.SQLVarchar("46:58:57.999999"),
			},
			{
				"sql_timestampdiff_0",
				"TIMESTAMPDIFF(YEAR, DATE '2001-01-02', DATE '2002-01-02')",
				evaluator.SQLInt(1),
			},
			{
				"sql_timestampdiff_1",
				"TIMESTAMPDIFF(YEAR, DATE '2002-01-02', DATE '2001-01-02')",
				evaluator.SQLInt(-1),
			},
			{
				"sql_timestampdiff_2",
				"TIMESTAMPDIFF(YEAR, DATE '2001-01-03', DATE '2002-01-02')",
				evaluator.SQLInt(0),
			},
			{
				"sql_timestampdiff_3",
				"TIMESTAMPDIFF(YEAR, DATE '2001-01-02', DATE '2002-01-03')",
				evaluator.SQLInt(1),
			},
			{
				"sql_timestampdiff_4",
				"TIMESTAMPDIFF(QUARTER, DATE '2002-04-02', DATE '2002-01-02')",
				evaluator.SQLInt(-1),
			},
			{
				"sql_timestampdiff_5",
				"TIMESTAMPDIFF(QUARTER, DATE '2002-01-02', DATE '2002-06-02')",
				evaluator.SQLInt(1),
			},
			{
				"sql_timestampdiff_6",
				"TIMESTAMPDIFF(QUARTER, DATE '2002-01-02', DATE '2002-07-02')",
				evaluator.SQLInt(2),
			},
			{
				"sql_timestampdiff_7",
				"TIMESTAMPDIFF(QUARTER, DATE '2002-07-02', DATE '2002-01-02')",
				evaluator.SQLInt(-2),
			},
			{
				"sql_timestampdiff_8",
				"TIMESTAMPDIFF(MONTH, DATE '2002-01-02', DATE '2002-02-01')",
				evaluator.SQLInt(0),
			},
			{
				"sql_timestampdiff_9",
				"TIMESTAMPDIFF(MONTH, DATE '2002-02-01', DATE '2001-01-02')",
				evaluator.SQLInt(-12),
			},
			{
				"sql_timestampdiff_10",
				"TIMESTAMPDIFF(MONTH, DATE '2002-01-02', DATE '2002-02-02')",
				evaluator.SQLInt(1),
			},
			{
				"sql_timestampdiff_11",
				"TIMESTAMPDIFF(MONTH, DATE '2002-02-03', DATE '2002-01-02')",
				evaluator.SQLInt(-1),
			},
			{
				"sql_timestampdiff_12",
				"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-16')",
				evaluator.SQLInt(2),
			},
			{
				"sql_timestampdiff_13",
				"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-15')",
				evaluator.SQLInt(1),
			},
			{
				"sql_timestampdiff_14",
				"TIMESTAMPDIFF(WEEK, DATE '2001-01-15', DATE '2001-01-02')",
				evaluator.SQLInt(-1),
			},
			{
				"sql_timestampdiff_15",
				"TIMESTAMPDIFF(WEEK, DATE '2001-01-02', DATE '2001-01-17')",
				evaluator.SQLInt(2),
			},
			{
				"sql_timestampdiff_16",
				"TIMESTAMPDIFF(DAY, DATE '2003-01-04', DATE '2003-01-16')",
				evaluator.SQLInt(12),
			},
			{
				"sql_timestampdiff_17",
				"TIMESTAMPDIFF(DAY, DATE '2003-01-16', DATE '2003-01-04')",
				evaluator.SQLInt(-12),
			},
			{
				"sql_timestampdiff_18",
				"TIMESTAMPDIFF(HOUR, DATE '2003-01-04', DATE '2003-01-06')",
				evaluator.SQLInt(48),
			},
			{
				"sql_timestampdiff_19",
				"TIMESTAMPDIFF(MINUTE, DATE '2003-01-04', DATE '2003-01-06')",
				evaluator.SQLInt(2880),
			},
			{
				"sql_timestampdiff_20",
				"TIMESTAMPDIFF(SECOND, DATE '2003-01-04', DATE '2003-01-05')",
				evaluator.SQLInt(86400),
			},
			{
				"sql_timestampdiff_21",
				"TIMESTAMPDIFF(MICROSECOND, DATE '2003-01-04', DATE '2003-01-05')",
				evaluator.SQLInt(86400000000),
			},
			{
				"sql_timestampdiff_22",
				"TIMESTAMPDIFF(MICROSECOND, TIMESTAMP '2002-01-01 12:30:09', " +
					"TIMESTAMP '2002-01-02 13:40:33')",
				evaluator.SQLInt(90624000000),
			},
			{
				"sql_timestampdiff_23",
				"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-01-02 12:30:09', " +
					"TIMESTAMP '2003-03-04 12:45:30')",
				evaluator.SQLInt(1),
			},
			{
				"sql_timestampdiff_24",
				"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-01-02 12:30:09', " +
					"TIMESTAMP '2002-03-04 12:45:30')",
				evaluator.SQLInt(0),
			},
			{
				"sql_timestampdiff_25",
				"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2002-03-04 12:45:30', " +
					"TIMESTAMP '2002-01-02 12:30:09')",
				evaluator.SQLInt(0),
			},
			{
				"sql_timestampdiff_26",
				"TIMESTAMPDIFF(SQL_TSI_YEAR, TIMESTAMP '2003-03-04 12:30:06', DATE '2002-03-04')",
				evaluator.SQLInt(-1),
			},
			{
				"sql_timestampdiff_27",
				"TIMESTAMPDIFF(SQL_TSI_YEAR, DATE '2004-03-04', TIMESTAMP '2003-03-04 12:30:06')",
				evaluator.SQLInt(0),
			},
			{
				"sql_timestampdiff_28",
				"TIMESTAMPDIFF(SQL_TSI_QUARTER, DATE '2002-01-01', " +
					"TIMESTAMP '2002-04-01 12:30:06')",
				evaluator.SQLInt(1),
			},
			{
				"sql_timestampdiff_29",
				"TIMESTAMPDIFF(SQL_TSI_QUARTER, TIMESTAMP '2002-04-01 12:30:06', " +
					"DATE '2002-01-01')",
				evaluator.SQLInt(-1),
			},
			{
				"sql_timestampdiff_30",
				"TIMESTAMPDIFF(SQL_TSI_QUARTER, TIMESTAMP '2002-01-01 12:30:06', " +
					"DATE '2002-04-01')",
				evaluator.SQLInt(0),
			},
			{
				"sql_timestampdiff_31",
				"TIMESTAMPDIFF(SQL_TSI_QUARTER, DATE '2002-04-01', " +
					"TIMESTAMP '2002-01-01 12:30:06')",
				evaluator.SQLInt(0),
			},
			{
				"sql_timestampdiff_32",
				"TIMESTAMPDIFF(SQL_TSI_MONTH, DATE '2002-01-01', TIMESTAMP '2002-03-01 12:30:09')",
				evaluator.SQLInt(2),
			},
			{
				"sql_timestampdiff_33",
				"TIMESTAMPDIFF(SQL_TSI_MONTH, TIMESTAMP '2002-03-01 12:30:09', DATE '2002-01-01')",
				evaluator.SQLInt(-2),
			},
			{
				"sql_timestampdiff_34",
				"TIMESTAMPDIFF(SQL_TSI_MONTH, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-03-01')",
				evaluator.SQLInt(1),
			},
			{
				"sql_timestampdiff_35",
				"TIMESTAMPDIFF(SQL_TSI_MONTH, DATE '2002-03-01', TIMESTAMP '2002-01-01 12:30:09')",
				evaluator.SQLInt(-1),
			},
			{
				"sql_timestampdiff_36",
				"TIMESTAMPDIFF(SQL_TSI_WEEK, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-08')",
				evaluator.SQLInt(0),
			},
			{
				"sql_timestampdiff_37",
				"TIMESTAMPDIFF(SQL_TSI_WEEK, DATE '2002-01-01', TIMESTAMP '2002-01-08 12:30:09')",
				evaluator.SQLInt(1),
			},
			{
				"sql_timestampdiff_38",
				"TIMESTAMPDIFF(SQL_TSI_WEEK, TIMESTAMP '2002-01-08 12:30:09', DATE '2002-01-01')",
				evaluator.SQLInt(-1),
			},
			{
				"sql_timestampdiff_39",
				"TIMESTAMPDIFF(SQL_TSI_DAY, DATE '2002-01-01', TIMESTAMP '2002-01-02 12:30:09')",
				evaluator.SQLInt(1),
			},
			{
				"sql_timestampdiff_40",
				"TIMESTAMPDIFF(SQL_TSI_DAY, TIMESTAMP '2002-01-02 12:30:09', DATE '2002-01-01')",
				evaluator.SQLInt(-1),
			},
			{
				"sql_timestampdiff_41",
				"TIMESTAMPDIFF(SQL_TSI_DAY, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')",
				evaluator.SQLInt(0),
			},
			{
				"sql_timestampdiff_42",
				"TIMESTAMPDIFF(SQL_TSI_HOUR, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')",
				evaluator.SQLInt(11),
			},
			{
				"sql_timestampdiff_43",
				"TIMESTAMPDIFF(SQL_TSI_HOUR, TIMESTAMP '2002-01-01 12:30:09', " +
					"TIMESTAMP '2002-01-02 11:02:33')",
				evaluator.SQLInt(22),
			},
			{
				"sql_timestampdiff_44",
				"TIMESTAMPDIFF(SQL_TSI_MINUTE, TIMESTAMP '2002-01-01 12:30:09', " +
					"TIMESTAMP '2002-01-01 13:02:33')",
				evaluator.SQLInt(32),
			},
			{
				"sql_timestampdiff_45",
				"TIMESTAMPDIFF(SQL_TSI_MINUTE, TIMESTAMP '2002-01-01 12:30:09', DATE '2002-01-02')",
				evaluator.SQLInt(689),
			},
			{
				"sql_timestampdiff_46",
				"TIMESTAMPDIFF(SQL_TSI_SECOND, TIMESTAMP '2002-01-01 12:30:09', " +
					"TIMESTAMP '2002-01-02 14:40:33')",
				evaluator.SQLInt(94224),
			},
			{"sql_to_days_0", "TO_DAYS(NULL)", evaluator.SQLNull},
			{"sql_to_days_1", "TO_DAYS('')", evaluator.SQLNull},
			{"sql_to_days_2", "TO_DAYS('0000-00-00')", evaluator.SQLNull},
			{"sql_to_days_3", "TO_DAYS('0000-01-01')", evaluator.SQLInt(1)},
			{"sql_to_days_4", "TO_DAYS('0000-11-11')", evaluator.SQLInt(315)},
			{"sql_to_days_5", "TO_DAYS('00-11-11')", evaluator.SQLInt(730800)},
			{"sql_to_days_6", "TO_DAYS('950501')", evaluator.SQLInt(728779)},
			{"sql_to_days_7", "TO_DAYS(950501)", evaluator.SQLInt(728779)},
			{"sql_to_days_8", "TO_DAYS('1995-05-01')", evaluator.SQLInt(728779)},
			{"sql_to_days_9", "TO_DAYS('2007-10-07')", evaluator.SQLInt(733321)},
			{"sql_to_days_10", "TO_DAYS(881111)", evaluator.SQLInt(726417)},
			{"sql_to_days_11", "TO_DAYS('2006-01-02')", evaluator.SQLInt(732678)},
			{"sql_to_days_12", "TO_DAYS('1452-04-15')", evaluator.SQLInt(530437)},
			{"sql_to_days_13", "TO_DAYS('4222-12-12')", evaluator.SQLInt(1542399)},
			{"sql_to_days_14", "TO_DAYS('2000-09-23 13:45:00')", evaluator.SQLInt(730751)},
			{"sql_to_days_15", "TO_DAYS('2000-09-24 13:45:00')", evaluator.SQLInt(730752)},
			{"sql_to_days_16", "TO_DAYS('2000-10-24 13:45:00')", evaluator.SQLInt(730782)},
			{"sql_to_seconds_0", "TO_SECONDS(NULL)", evaluator.SQLNull},
			{"sql_to_seconds_1", "TO_SECONDS('')", evaluator.SQLNull},
			{"sql_to_seconds_2", "TO_SECONDS('0000-00-00')", evaluator.SQLNull},
			{"sql_to_seconds_3", "TO_SECONDS('0000-01-01')", evaluator.SQLInt(86400)},
			{"sql_to_seconds_4", "TO_SECONDS('0000-11-11')", evaluator.SQLInt(27216000)},
			{"sql_to_seconds_5", "TO_SECONDS('00-11-11')", evaluator.SQLInt(63141120000)},
			{"sql_to_seconds_6", "TO_SECONDS('950501')", evaluator.SQLInt(62966505600)},
			{"sql_to_seconds_7", "TO_SECONDS(950501)", evaluator.SQLInt(62966505600)},
			{"sql_to_seconds_8", "TO_SECONDS('1995-05-01')", evaluator.SQLInt(62966505600)},
			{"sql_to_seconds_9", "TO_SECONDS('2007-10-07')", evaluator.SQLInt(63358934400)},
			{"sql_to_seconds_10", "TO_SECONDS(881111)", evaluator.SQLInt(62762428800)},
			{"sql_to_seconds_11", "TO_SECONDS('2006-01-02')", evaluator.SQLInt(63303379200)},
			{"sql_to_seconds_12", "TO_SECONDS('1452-04-15')", evaluator.SQLInt(45829756800)},
			{"sql_to_seconds_13", "TO_SECONDS('4222-12-12')", evaluator.SQLInt(133263273600)},
			{
				"sql_to_seconds_14",
				"TO_SECONDS('2000-09-23 13:45:00')",
				evaluator.SQLInt(63136935900),
			},
			{
				"sql_to_seconds_15",
				"TO_SECONDS('2000-09-24 13:45:00')",
				evaluator.SQLInt(63137022300),
			},
			{
				"sql_to_seconds_16",
				"TO_SECONDS('2000-10-24 13:45:00')",
				evaluator.SQLInt(63139614300),
			},
			{
				"sql_to_seconds_17",
				"TO_SECONDS('2000-10-24 15:45:00')",
				evaluator.SQLInt(63139621500),
			},
			{
				"sql_to_seconds_18",
				"TO_SECONDS('2000-10-24 13:47:00')",
				evaluator.SQLInt(63139614420),
			},
			{
				"sql_to_seconds_19",
				"TO_SECONDS('2000-10-24 13:45:59')",
				evaluator.SQLInt(63139614359),
			},
			{"sql_trim_0", "TRIM(NULL)", evaluator.SQLNull},
			{"sql_trim_1", "TRIM('   bar   ')", evaluator.SQLVarchar("bar")},
			{"sql_trim_2", "TRIM(BOTH 'xyz' FROM 'xyzbarxyzxyz')", evaluator.SQLVarchar("bar")},
			{
				"sql_trim_3",
				"TRIM(LEADING 'xyz' FROM 'xyzbarxyzxyz')",
				evaluator.SQLVarchar("barxyzxyz"),
			},
			{
				"sql_trim_4",
				"TRIM(TRAILING 'xyz' FROM 'xyzbarxyzxyz')",
				evaluator.SQLVarchar("xyzbar"),
			},
			{"sql_trim_5", "TRIM('xyz' FROM 'xyzbarxyzxyz')", evaluator.SQLVarchar("bar")},
			{"sql_truncate_0", "TRUNCATE(NULL, 2)", evaluator.SQLNull},
			{"sql_truncate_1", "TRUNCATE(1234.1234, NULL)", evaluator.SQLNull},
			{"sql_truncate_2", "TRUNCATE(1 / 0, 2)", evaluator.SQLNull},
			{"sql_truncate_3", "TRUNCATE(1234.1234, 1 / 0)", evaluator.SQLNull},
			{"sql_truncate_4", "TRUNCATE(1234.1234, 3)", evaluator.SQLFloat(1234.123)},
			{"sql_truncate_5", "TRUNCATE(1234.1234, 5)", evaluator.SQLFloat(1234.1234)},
			{"sql_truncate_6", "TRUNCATE(1234.1234, 0)", evaluator.SQLFloat(1234)},
			{"sql_truncate_7", "TRUNCATE(1234.1234, -3)", evaluator.SQLFloat(1000)},
			{"sql_truncate_8", "TRUNCATE(1234.1234, -5)", evaluator.SQLFloat(0)},
			{"sql_truncate_9", "TRUNCATE(-1234.1234, 3)", evaluator.SQLFloat(-1234.123)},
			{"sql_truncate_10", "TRUNCATE(-1234.1234, -3)", evaluator.SQLFloat(-1000)},
			{"sql_ucase_0", "UCASE(NULL)", evaluator.SQLNull},
			{"sql_ucase_1", "UCASE('sdg')", evaluator.SQLVarchar("SDG")},
			{"sql_ucase_2", "UCASE(124)", evaluator.SQLVarchar("124")},
			{"sql_ucase_3", "UPPER(NULL)", evaluator.SQLNull},
			{"sql_ucase_4", "UPPER('')", evaluator.SQLVarchar("")},
			{"sql_ucase_5", "UPPER('a')", evaluator.SQLVarchar("A")},
			{"sql_ucase_6", "UPPER('AWESOME')", evaluator.SQLVarchar("AWESOME")},
			{"sql_ucase_7", "UPPER('AwEsOmE')", evaluator.SQLVarchar("AWESOME")},
			{"sql_unix_timestamp_0", "UNIX_TIMESTAMP(NULL)", evaluator.SQLNull},
			{"sql_unix_timestamp_1", "UNIX_TIMESTAMP('1923-12-12')", evaluator.SQLFloat(0)},
			/*
				These tests will fail if run on a server in a timezone
				different from EST (-05:00) - thus are flaky and commented out.
				test{
					"sql_unix_timestamp_2",
					"UNIX_TIMESTAMP('2015-11-13 10:20:19')",
					SQLUint64(1447428019),
				},
				test{
					"sql_unix_timestamp_3",
					"UNIX_TIMESTAMP('2017-03-27 03:00:00')",
					SQLUint64(1490598000),
				},
				test{
					"sql_unix_timestamp_4",
					"UNIX_TIMESTAMP('2012-11-17 12:00:00')",
					SQLUint64(1353171600),
				},
				test{"sql_unix_timestamp_5", "UNIX_TIMESTAMP('1985-03-21')", SQLUint64(480229200)},
				test{"sql_unix_timestamp_6", "UNIX_TIMESTAMP('1985')", SQLFloat(0)},
				test{"sql_unix_timestamp_7", "UNIX_TIMESTAMP('1985-12')", SQLFloat(0)},
				test{"sql_unix_timestamp_8", "UNIX_TIMESTAMP('1985-12-aa')", SQLFloat(0)},
				test{"sql_unix_timestamp_9", "UNIX_TIMESTAMP('1985-12-')", SQLFloat(0)},
				test{"sql_unix_timestamp_10", "UNIX_TIMESTAMP('1985-12-1')", SQLUint64(502261200)},
				test{"sql_unix_timestamp_11", "UNIX_TIMESTAMP('1985-12-01')", SQLUint64(502261200)},
			*/
			{"sql_week_0", "WEEK(NULL)", evaluator.SQLNull},
			{"sql_week_1", "WEEK('sdg')", evaluator.SQLNull},
			{"sql_week_2", "WEEK('2016-1-01 10:23:52')", evaluator.SQLInt(0)},
			{"sql_week_3", "WEEK(DATE '2009-1-01')", evaluator.SQLInt(0)},
			{"sql_week_4", "WEEK(DATE '2009-1-01',0)", evaluator.SQLInt(0)},
			{"sql_week_5", "WEEK(DATE '2009-1-01','str')", evaluator.SQLInt(0)},
			{"sql_week_6", "WEEK(DATE '2009-1-01',1)", evaluator.SQLInt(1)},
			{"sql_week_7", "WEEK(DATE '2009-1-01',2)", evaluator.SQLInt(52)},
			{"sql_week_8", "WEEK(DATE '2009-1-01',3)", evaluator.SQLInt(1)},
			{"sql_week_9", "WEEK(DATE '2009-1-01',4)", evaluator.SQLInt(0)},
			{"sql_week_10", "WEEK(DATE '2009-1-01',5)", evaluator.SQLInt(0)},
			{"sql_week_11", "WEEK(DATE '2009-1-01',6)", evaluator.SQLInt(53)},
			{"sql_week_12", "WEEK(DATE '2009-1-01',7)", evaluator.SQLInt(52)},
			{"sql_week_13", "WEEK(DATE '2009-1-05')", evaluator.SQLInt(1)},
			{"sql_week_14", "WEEK(DATE '2009-1-05',1)", evaluator.SQLInt(2)},
			{"sql_week_15", "WEEK(DATE '2009-1-05',2)", evaluator.SQLInt(1)},
			{"sql_week_16", "WEEK(DATE '2009-1-05',3)", evaluator.SQLInt(2)},
			{"sql_week_17", "WEEK(DATE '2009-1-05',4)", evaluator.SQLInt(1)},
			{"sql_week_18", "WEEK(DATE '2009-1-05',5)", evaluator.SQLInt(1)},
			{"sql_week_19", "WEEK(DATE '2009-1-05',6)", evaluator.SQLInt(1)},
			{"sql_week_20", "WEEK(DATE '2009-1-05',7)", evaluator.SQLInt(1)},
			{"sql_week_21", "WEEK(DATE '2009-12-31')", evaluator.SQLInt(52)},
			{"sql_week_22", "WEEK(DATE '2009-12-31',1)", evaluator.SQLInt(53)},
			{"sql_week_23", "WEEK(DATE '2009-12-31',2)", evaluator.SQLInt(52)},
			{"sql_week_24", "WEEK(DATE '2009-12-31',3)", evaluator.SQLInt(53)},
			{"sql_week_25", "WEEK(DATE '2009-12-31',4)", evaluator.SQLInt(52)},
			{"sql_week_26", "WEEK(DATE '2009-12-31',5)", evaluator.SQLInt(52)},
			{"sql_week_27", "WEEK(DATE '2009-12-31',6)", evaluator.SQLInt(52)},
			{"sql_week_28", "WEEK(DATE '2009-12-31',7)", evaluator.SQLInt(52)},
			{"sql_week_29", "WEEK(DATE '2007-12-31')", evaluator.SQLInt(52)},
			{"sql_week_30", "WEEK(DATE '2007-12-31',1)", evaluator.SQLInt(53)},
			{"sql_week_31", "WEEK(DATE '2007-12-31',2)", evaluator.SQLInt(52)},
			{"sql_week_32", "WEEK(DATE '2007-12-31',3)", evaluator.SQLInt(1)},
			{"sql_week_33", "WEEK(DATE '2007-12-31',4)", evaluator.SQLInt(53)},
			{"sql_week_34", "WEEK(DATE '2007-12-31',5)", evaluator.SQLInt(53)},
			{"sql_week_35", "WEEK(DATE '2007-12-31',6)", evaluator.SQLInt(1)},
			{"sql_week_36", "WEEK(DATE '2007-12-31',7)", evaluator.SQLInt(53)},
			{"sql_weekday_0", "WEEKDAY(NULL)", evaluator.SQLNull},
			{"sql_weekday_1", "WEEKDAY('sdg')", evaluator.SQLNull},
			{"sql_weekday_2", "WEEKDAY('2016-1-01 10:23:52')", evaluator.SQLInt(4)},
			{"sql_weekday_3", "WEEKDAY('2005-05-11')", evaluator.SQLInt(2)},
			{"sql_weekday_4", "WEEKDAY(DATE '2016-7-10')", evaluator.SQLInt(6)},
			{"sql_weekday_5", "WEEKDAY(DATE '2016-7-11')", evaluator.SQLInt(0)},
			{"sql_weekday_6", "WEEKDAY(TIMESTAMP '2016-7-13 21:22:23')", evaluator.SQLInt(2)},
			{"sql_weekofyear_0", "WEEKOFYEAR(NULL)", evaluator.SQLNull},
			{"sql_weekofyear_1", "WEEKOFYEAR('sdg')", evaluator.SQLNull},
			{"sql_weekofyear_2", "WEEKOFYEAR('2008-02-20')", evaluator.SQLInt(8)},
			{"sql_weekofyear_3", "WEEKOFYEAR('2009-01-01')", evaluator.SQLInt(1)},
			{"sql_weekofyear_4", "WEEKOFYEAR(DATE '2009-01-05')", evaluator.SQLInt(2)},
			{"sql_subtract_expr_0", "0 - 0", evaluator.SQLInt(0)},
			{"sql_subtract_expr_1", "-1 - 1", evaluator.SQLInt(-2)},
			{"sql_subtract_expr_2", "10 - 32", evaluator.SQLInt(-22)},
			{"sql_subtract_expr_3", "-10 - -32", evaluator.SQLInt(22)},
			{"sql_unary_minus_0", "- 10", evaluator.SQLInt(-10)},
			{"sql_unary_minus_1", "- a", evaluator.SQLInt(-123)},
			{"sql_unary_minus_2", "- b", evaluator.SQLInt(-456)},
			{"sql_unary_minus_3", "- null", evaluator.SQLNull},
			{"sql_unary_minus_4", "- true", evaluator.SQLInt(-1)},
			{"sql_unary_minus_5", "- false", evaluator.SQLInt(0)},
			{"sql_unary_minus_6", "- date '2005-05-11'", evaluator.SQLInt(-20050511)},
			{
				"sql_unary_minus_7",
				"- timestamp '2005-05-11 12:22:04'",
				evaluator.SQLInt(-20050511122204),
			},
			{"sql_unary_minus_8", "- '4' ", evaluator.SQLFloat(-4)},
			{"sql_unary_minus_9", "- 6.7", evaluator.SQLDecimal128(decimal.New(-67, -1))},
			{"sql_unary_minus_10", "- '3.3'", evaluator.SQLFloat(-3.3)},
			{"sql_variable_expr_0", "@@autocommit", evaluator.SQLTrue},
			{"sql_variable_expr_1", "@@global.autocommit", evaluator.SQLTrue},
			{"sql_unary_plus_expr_0", "+1", evaluator.SQLInt(1)},
			{"sql_unary_plus_expr_1", "+'string'", evaluator.SQLVarchar("string")},
			{"sql_unary_plus_expr_2", "+a", evaluator.SQLInt(123)},
			{"sql_early_eval_0", "(1, 3) > (2, 4)", evaluator.SQLFalse},
			{"sql_early_eval_1", "(1, 3) > ROW(2, 4)", evaluator.SQLFalse},
		}

		runTests(t, evalCtx, tests)

		// aggregation tests
		var t1, t2 time.Time
		t1 = time.Now()
		t2 = t1.Add(time.Hour)

		aggCtx := evaluator.NewEvalCtx(execCtx, collation.Default,
			&evaluator.Row{Data: evaluator.Values{
				{SelectID: 1, Database: "test", Table: "bar", Name: "a", Data: evaluator.SQLNull},
				{SelectID: 1, Database: "test", Table: "bar", Name: "b", Data: evaluator.SQLInt(3)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "c", Data: evaluator.SQLNull},
				{
					SelectID: 1,
					Database: "test",
					Table:    "bar",
					Name:     "g",
					Data:     evaluator.SQLDate{Time: t1},
				},
			}},
			&evaluator.Row{Data: evaluator.Values{
				{SelectID: 1, Database: "test", Table: "bar", Name: "a", Data: evaluator.SQLInt(3)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "b", Data: evaluator.SQLNull},
				{SelectID: 1, Database: "test", Table: "bar", Name: "c", Data: evaluator.SQLNull},
				{
					SelectID: 1,
					Database: "test",
					Table:    "bar",
					Name:     "g",
					Data:     evaluator.SQLDate{Time: t2},
				},
			}},
			&evaluator.Row{Data: evaluator.Values{
				{SelectID: 1, Database: "test", Table: "bar", Name: "a", Data: evaluator.SQLInt(5)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "b", Data: evaluator.SQLInt(6)},
				{SelectID: 1, Database: "test", Table: "bar", Name: "c", Data: evaluator.SQLNull},
				{SelectID: 1, Database: "test", Table: "bar", Name: "g", Data: evaluator.SQLNull},
			}},
		)

		aggTests := []test{
			{"sql_agg_expr_avg_0", "AVG(NULL)", evaluator.SQLNull},
			{"sql_agg_expr_avg_1", "AVG(a)", evaluator.SQLFloat(4)},
			{"sql_agg_expr_avg_2", "AVG(b)", evaluator.SQLFloat(4.5)},
			{"sql_agg_expr_avg_3", "AVG(c)", evaluator.SQLNull},
			{"sql_agg_expr_avg_4", "AVG('a')", evaluator.SQLFloat(0)},
			{"sql_agg_expr_avg_5", "AVG(-20)", evaluator.SQLFloat(-20)},
			{"sql_agg_expr_avg_6", "AVG(20)", evaluator.SQLFloat(20)},
			{"sql_count_expr_0", "COUNT(NULL)", evaluator.SQLInt(0)},
			{"sql_count_expr_1", "COUNT(a)", evaluator.SQLInt(2)},
			{"sql_count_expr_2", "COUNT(b)", evaluator.SQLInt(2)},
			{"sql_count_expr_3", "COUNT(c)", evaluator.SQLInt(0)},
			{"sql_count_expr_4", "COUNT(g)", evaluator.SQLInt(2)},
			{"sql_count_expr_5", "COUNT('a')", evaluator.SQLInt(3)},
			{"sql_count_expr_6", "COUNT(-20)", evaluator.SQLInt(3)},
			{"sql_count_expr_7", "COUNT(20)", evaluator.SQLInt(3)},
			{"sql_min_expr_0", "MIN(NULL)", evaluator.SQLNull},
			{"sql_min_expr_1", "MIN(a)", evaluator.SQLInt(3)},
			{"sql_min_expr_2", "MIN(b)", evaluator.SQLInt(3)},
			{"sql_min_expr_3", "MIN(c)", evaluator.SQLNull},
			{"sql_min_expr_4", "MIN('a')", evaluator.SQLVarchar("a")},
			{"sql_min_expr_5", "MIN(-20)", evaluator.SQLInt(-20)},
			{"sql_min_expr_6", "MIN(20)", evaluator.SQLInt(20)},
			{"sql_max_expr_0", "MAX(NULL)", evaluator.SQLNull},
			{"sql_max_expr_1", "MAX(a)", evaluator.SQLInt(5)},
			{"sql_max_expr_2", "MAX(b)", evaluator.SQLInt(6)},
			{"sql_max_expr_3", "MAX(c)", evaluator.SQLNull},
			{"sql_max_expr_4", "MAX('a')", evaluator.SQLVarchar("a")},
			{"sql_max_expr_5", "MAX(-20)", evaluator.SQLInt(-20)},
			{"sql_max_expr_6", "MAX(20)", evaluator.SQLInt(20)},
			{"sql_sleep_expr_0", "SLEEP(1)", evaluator.SQLInt(0)},
			{"sql_sleep_expr_1", "SLEEP(1.5)", evaluator.SQLInt(0)},
			{"sql_sleep_expr_2", "SLEEP(0)", evaluator.SQLInt(0)},
			{"sql_sum_expr_0", "SUM(NULL)", evaluator.SQLNull},
			{"sql_sum_expr_1", "SUM(a)", evaluator.SQLFloat(8)},
			{"sql_sum_expr_2", "SUM(b)", evaluator.SQLFloat(9)},
			{"sql_sum_expr_3", "SUM(c)", evaluator.SQLNull},
			{"sql_sum_expr_4", "SUM('a')", evaluator.SQLFloat(0)},
			{"sql_sum_expr_5", "SUM(-20)", evaluator.SQLFloat(-60)},
			{"sql_sum_expr_6", "SUM(20)", evaluator.SQLFloat(60)},
			{"sql_std_expr_0", "STD(NULL)", evaluator.SQLNull},
			{"sql_std_dev_expr", "STDDEV(a)", evaluator.SQLFloat(1)},
			{"sql_std_dev_pop_expr", "STDDEV_POP(b)", evaluator.SQLFloat(1.5)},
			{"sql_std_expr_1", "STD(c)", evaluator.SQLNull},
			{"sql_std_dev_samp_expr_0", "STDDEV_SAMP(NULL)", evaluator.SQLNull},
			{"sql_std_dev_samp_expr_1", "STDDEV_SAMP(a)", evaluator.SQLFloat(1.4142135623730951)},
			{"sql_std_dev_samp_expr_2", "STDDEV_SAMP(b)", evaluator.SQLFloat(2.1213203435596424)},
			{"sql_std_dev_samp_expr_3", "STDDEV_SAMP(c)", evaluator.SQLNull},
		}
		runTests(t, aggCtx, aggTests)

		// type tests
		typeTests := []typeTest{
			{"sql_coalesce_type_0", "COALESCE(NULL, 1, 'A')", schema.SQLVarchar},
			{"sql_coalesce_type_1", "COALESCE(NULL, 1, 23)", schema.SQLInt},
			{"sql_convert_type_0", "CONVERT(DATE '2006-05-11', SIGNED)", schema.SQLInt},
			{"sql_convert_type_1", "CONVERT(true, SQL_DOUBLE)", schema.SQLFloat},
			{"sql_convert_type_2", "CONVERT('16a', CHAR)", schema.SQLVarchar},
			{"sql_convert_type_3", "CONVERT('2006-05-11', DATE)", schema.SQLDate},
			{
				"sql_convert_type_4",
				"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATETIME)",
				schema.SQLTimestamp,
			},
			{
				"sql_convert_type_5",
				"CONVERT(DATE '2006-05-11', SQL_TIMESTAMP)",
				schema.SQLTimestamp,
			},
			{"sql_date_add_type_0", "DATE_ADD('2002-01-02', INTERVAL 1 YEAR)", schema.SQLTimestamp},
			{
				"sql_date_add_type_1",
				"DATE_ADD(DATE '2002-01-02', INTERVAL 1 HOUR)",
				schema.SQLTimestamp,
			},
			{
				"sql_date_add_type_2",
				"DATE_ADD(TIMESTAMP '2003-01-02 10:28:06', INTERVAL '2 2:3' DAY_SECOND)",
				schema.SQLTimestamp,
			},
			{"sql_date_sub_type_0", "DATE_SUB('2002-01-02', INTERVAL 1 YEAR)", schema.SQLTimestamp},
			{
				"sql_date_sub_type_1",
				"DATE_SUB(DATE '2002-01-02', INTERVAL 1 HOUR)",
				schema.SQLTimestamp,
			},
			{
				"sql_date_sub_type_2",
				"DATE_SUB(TIMESTAMP '2003-01-02 10:28:06', INTERVAL '2 2:3' DAY_SECOND)",
				schema.SQLTimestamp,
			},
			{
				"sql_greatest_type_0",
				"GREATEST(DATE '2005-05-11', DATE '2006-05-11', DATE '2000-05-11')",
				schema.SQLDate,
			},
			{"sql_greatest_type_1", "GREATEST(1, 123.52, 'something')", schema.SQLDecimal128},
			{"sql_if_type_0", "IF('ca.gh', 4, 5)", schema.SQLInt},
			{"sql_if_type_1", "IF('ca.gh', 4, 5.3)", schema.SQLDecimal128},
			{"sql_if_type_2", "IF('ca.gh', 'sdf', 5.2)", schema.SQLVarchar},
			{"sql_if_type_3", "IF('ca.gh', 'sdf', NULL)", schema.SQLVarchar},
			{"sql_if_null_type_0", "IFNULL(4, 5)", schema.SQLInt},
			{"sql_if_null_type_1", "IFNULL(4, 5.3)", schema.SQLDecimal128},
			{"sql_if_null_type_2", "IFNULL('sdf', NULL)", schema.SQLVarchar},
			{"sql_interval_type_0", "INTERVAL(4, 5)", schema.SQLInt64},
			{"sql_interval_type_1", "INTERVAL(4, 5.3)", schema.SQLInt64},
			{"sql_interval_type_2", "INTERVAL(NULL, 4)", schema.SQLInt64},
			{"sql_null_if_type_0", "NULLIF(3, null)", schema.SQLInt},
			{"sql_null_if_type_1", "NULLIF('abc', 'abc')", schema.SQLVarchar},
			{
				"sql_least_type_0",
				"LEAST(DATE '2005-05-11', DATE '2006-05-11', DATE '2000-05-11')",
				schema.SQLDate,
			},
			{"sql_least_type_1", "LEAST(1, 123.52, 'something')", schema.SQLDecimal128},
			{
				"sql_timestampadd_type_0",
				"TIMESTAMPADD(SQL_TSI_QUARTER, 2, DATE '2002-07-02')",
				schema.SQLTimestamp,
			},
			{
				"sql_timestampadd_type_1",
				"TIMESTAMPADD(SQL_TSI_SECOND, 1, TIMESTAMP '2003-01-02 12:30:08')",
				schema.SQLTimestamp,
			},
		}
		runTypeTests(t, evalCtx, typeTests)

		t.Run("sql_sleep_with_neg_value", func(t *testing.T) {
			subject, err := evaluator.NewSQLScalarFunctionExpr(
				"sleep", []evaluator.SQLExpr{evaluator.SQLInt(-1)})
			req.Nil(err, "unable to create scalar sleep expression")
			_, err = subject.Evaluate(evalCtx)
			req.NotNil(err, "did not return error on negative sleep value")
		})

		t.Run("sql_sleep_with_null_value", func(t *testing.T) {
			subject, err := evaluator.NewSQLScalarFunctionExpr(
				"sleep",
				[]evaluator.SQLExpr{evaluator.SQLNull},
			)
			req.Nil(err, "unable to create scalar sleep expression")
			_, err = subject.Evaluate(evalCtx)
			req.NotNil(err, "did not return error on null sleep value")
		})

		t.Run("sql_assignment_expr", func(t *testing.T) {
			e := evaluator.NewSQLAssignmentExpr(
				evaluator.NewSQLVariableExpr(
					"test",
					variable.UserKind,
					variable.SessionScope,
					schema.SQLNone,
				),
				evaluator.NewSQLAddExpr(
					evaluator.SQLInt(1),
					evaluator.SQLInt(3),
				),
			)

			result, err := e.Evaluate(evalCtx)
			req.Nil(err, "unable to evaluate sql assignment expression")
			req.Equal(
				result,
				evaluator.SQLInt(4),
				"expected value of assignment does not match actual value",
			)
		})

		t.Run("sql_divide_by_zero", func(t *testing.T) {
			subject := evaluator.NewSQLDivideExpr(
				evaluator.SQLInt(10),
				evaluator.SQLInt(0),
			)
			result, err := subject.Evaluate(evalCtx)
			req.Nil(err, "unable to evaluate sql expression")
			req.IsType(evaluator.SQLNull, result, "actual type does not match expected type")
		})

		t.Run("subject: sqlcolumnexpr", func(t *testing.T) {
			t.Run("should return the value of the field when it exists", func(t *testing.T) {
				subject := evaluator.NewSQLColumnExpr(1,
					"test",
					"bar",
					"a",
					schema.SQLInt,
					schema.MongoInt,
				)
				result, err := subject.Evaluate(evalCtx)
				req.Nil(err, "unable to evalute sql expression")
				req.Equal(
					result,
					evaluator.SQLInt(123),
					"actual value of evaluated expression does not match expected value",
				)
			})

			t.Run("should return nil when the field is null", func(t *testing.T) {
				subject := evaluator.NewSQLColumnExpr(1,
					"test",
					"bar",
					"c",
					schema.SQLInt,
					schema.MongoInt,
				)
				result, err := subject.Evaluate(evalCtx)
				req.Nil(err, "unable to evalute sql expression")
				req.IsType(
					evaluator.SQLNull,
					result,
					"actual type does not match expected type of evaluated expression",
				)
			})

			t.Run("should return nil when the field doesn't exists", func(t *testing.T) {
				subject := evaluator.NewSQLColumnExpr(1,
					"test",
					"bar",
					"no_existy",
					schema.SQLInt,
					schema.MongoInt,
				)
				result, err := subject.Evaluate(evalCtx)
				req.Nil(err, "unable to evalute sql expression")
				req.IsType(
					evaluator.SQLNull,
					result,
					"actual type does not match expected type of evaluated expression",
				)
			})
		})

		//t.Run("subject: sqlscalarfunctionexpr", func(t *testing.T) {
		//

		t.Run("subject: date", func(t *testing.T) {
			dateTime0, _ := time.Parse("2006-01-02", "2014-04-13")
			dateTime1, _ := time.Parse("15:04:05", "11:49:36")
			dateTime2, _ := time.Parse("2006-01-02 15:04:05.999999999", "1997-01-31 09:26:50.124")
			tests = []test{
				{
					"sql_date_parse_comparison_0",
					"DATE '2014-04-13'",
					evaluator.SQLDate{Time: dateTime0},
				},
				{
					"sql_date_parse_comparison_1",
					"{d '2014-04-13'}",
					evaluator.SQLDate{Time: dateTime0},
				},
				{
					"sql_date_parse_comparison_2",
					"TIME '11:49:36'",
					evaluator.SQLTimestamp{Time: dateTime1},
				},
				{
					"sql_date_parse_comparison_3",
					"{t '11:49:36'}",
					evaluator.SQLTimestamp{Time: dateTime1},
				},
				{
					"sql_timestamp_parse_0",
					"TIMESTAMP '1997-01-31 09:26:50.124'",
					evaluator.SQLTimestamp{Time: dateTime2},
				},
				{
					"sql_timestamp_parse_1",
					"{ts '1997-01-31 09:26:50.124'}",
					evaluator.SQLTimestamp{Time: dateTime2},
				},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: adddate", func(t *testing.T) {
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
				{"sql_add_date_0", "ADDDATE(NULL, INTERVAL 1 YEAR)", evaluator.SQLNull},
				{
					"sql_add_date_1",
					"ADDDATE('2002-01-02', INTERVAL 1 YEAR)",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_add_date_2",
					"ADDDATE('2003-08-31', INTERVAL 1 QUARTER)",
					evaluator.SQLTimestamp{Time: d2},
				},
				{
					"sql_add_date_3",
					"ADDDATE('2003-10-31', INTERVAL 1 MONTH)",
					evaluator.SQLTimestamp{Time: d2},
				},
				{
					"sql_add_date_4",
					"ADDDATE('2003-01-01', INTERVAL 1 DAY)",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_add_date_5",
					"ADDDATE('2003-01-02 14:30:09', INTERVAL -2 HOUR)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_6",
					"ADDDATE('2003-01-02 12:23:09', INTERVAL 7 MINUTE)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_7",
					"ADDDATE('2003-01-02 12:30:12', INTERVAL -3 SECOND)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_8",
					"ADDDATE('2003-01-02 12:32:10', INTERVAL '-2:1' MINUTE_SECOND)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_9",
					"ADDDATE('2003-01-02 05:27:06', INTERVAL '7:3:3' HOUR_SECOND)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_10",
					"ADDDATE('2003-01-02 15:32:09', INTERVAL '-3:2' HOUR_MINUTE)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_11",
					"ADDDATE('2002-12-31 10:27:05', INTERVAL '2 2:3:4' DAY_SECOND)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_12",
					"ADDDATE('2002-12-31 10:27:09', INTERVAL '2 2:3' DAY_MINUTE)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_13",
					"ADDDATE('2003-01-01 08:30:09', INTERVAL '1 4' DAY_HOUR)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_14",
					"ADDDATE('2000-09-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_15",
					"ADDDATE('2003-01-02 12:33:09', INTERVAL '-3' HOUR_MINUTE)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_16",
					"ADDDATE('2003-01-02 10:28:06', INTERVAL '2 2:3' DAY_SECOND)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_add_date_17",
					"ADDDATE('2003-01-02 10:28:06', 32)",
					evaluator.SQLTimestamp{Time: t2},
				},
				{
					"sql_add_date_18",
					"ADDDATE('2003-01-02 10:28:06', 43)",
					evaluator.SQLTimestamp{Time: t3},
				},
				{
					"sql_add_date_19",
					"ADDDATE('2003-01-02 10:28:06.000', 43)",
					evaluator.SQLTimestamp{Time: t3},
				},
				{
					"sql_add_date_20",
					"ADDDATE('2003-01-02 10:28:06.000000', 43)",
					evaluator.SQLTimestamp{Time: t3},
				},
				{"sql_add_date_21", "ADDDATE('2008-01-02', 31)", evaluator.SQLTimestamp{Time: d3}},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: convert", func(t *testing.T) {
			d, err := time.Parse("2006-01-02", "2006-05-11")
			req.Nil(err, "unable to parse time from string")
			t1, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 12:32:12")
			req.Nil(err, "unable to parse time from string")
			dt, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 00:00:00")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_convert_expr_66", "CONVERT('2006-05-11', DATE)", evaluator.SQLDate{Time: d}},
				{"sql_convert_expr_67", "CONVERT(true, DATE)", evaluator.SQLNull},
				{
					"sql_convert_expr_68",
					"CONVERT(DATE '2006-05-11', DATE)",
					evaluator.SQLDate{Time: d},
				},
				{
					"sql_convert_expr_69",
					"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATE)",
					evaluator.SQLDate{Time: d},
				},
				{"sql_convert_expr_70", "CONVERT(NULL, DATETIME)", evaluator.SQLNull},
				{"sql_convert_expr_71", "CONVERT(-3.4, DATETIME)", evaluator.SQLNull},
				{"sql_convert_expr_72", "CONVERT('janna', DATETIME)", evaluator.SQLNull},
				{
					"sql_convert_expr_73",
					"CONVERT('2006-05-11', DATETIME)",
					evaluator.SQLTimestamp{Time: dt},
				},
				{"sql_convert_expr_74", "CONVERT(true, DATETIME)", evaluator.SQLNull},
				{"sql_convert_expr_75", "CONVERT(3, SQL_TIMESTAMP)", evaluator.SQLNull},
				{
					"sql_convert_expr_76",
					"CONVERT(TIMESTAMP '2006-05-11 12:32:12', DATETIME)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_convert_expr_77",
					"CONVERT(DATE '2006-05-11', SQL_TIMESTAMP)",
					evaluator.SQLTimestamp{Time: dt},
				},
				{
					"sql_convert_expr_78",
					"CONVERT('12:32:12', TIME)",
					evaluator.SQLTimestamp{Time: time.Date(0, 1, 1, 12, 32, 12, 0, time.UTC)},
				},
				{
					"sql_convert_expr_79",
					"CONVERT('2006-04-11 12:32:12', TIME)",
					evaluator.SQLTimestamp{Time: time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)},
				},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("cot should error when out of range", func(t *testing.T) {
			subject, err := evaluator.NewSQLScalarFunctionExpr(
				"cot",
				[]evaluator.SQLExpr{evaluator.SQLFloat(0)},
			)
			req.Nil(err, "unable to create sql scalar expression")
			_, err = subject.Evaluate(evalCtx)
			req.NotNil(err, "did not return nil for out of range cot expression")
		})

		t.Run("subject: utc_date", func(t *testing.T) {
			now := time.Now().In(time.UTC)
			t0 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			tests := []test{
				{"sql_utc_date_0", "UTC_DATE()", evaluator.SQLDate{Time: t0}},
				{"sql_utc_date_1", "UTC_DATE", evaluator.SQLDate{Time: t0}},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: date", func(t *testing.T) {
			fmtString := "2006-01-02"

			d, err := time.Parse(fmtString, "2016-03-01")
			req.Nil(err, "unable to parse time from string")

			dExpected := evaluator.SQLDate{Time: d}

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
				{"sql_date_invalid_0", "DATE(NULL)", evaluator.SQLNull},
				{"sql_date_invalid_1", "DATE(23)", evaluator.SQLNull},
				{"sql_date_invalid_2", "DATE('cat')", evaluator.SQLNull},
				{"sql_date_invalid_3", "DATE(6911)", evaluator.SQLNull},
				{"sql_date_invalid_4", "DATE(2017110722040)", evaluator.SQLNull},
				{"sql_date_invalid_5", "DATE(-50)", evaluator.SQLNull},
				{"sql_date_invalid_6", "DATE('')", evaluator.SQLNull},

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
				{"sql_date_non_paddable_0", "DATE(1)", evaluator.SQLNull},
				{"sql_date_non_paddable_1", "DATE(11)", evaluator.SQLNull},

				// number inputs requiring padding
				{"sql_date_padded_nums_0", "DATE(111)", evaluator.SQLDate{Time: jan112000}},
				{"sql_date_padded_nums_1", "DATE(1110)", evaluator.SQLDate{Time: nov102000}},
				{"sql_date_padded_nums_2", "DATE(61110)", evaluator.SQLDate{Time: nov102006}},
				{"sql_date_padded_nums_3", "DATE(1161110)", evaluator.SQLDate{Time: nov100116}},
				{"sql_date_padded_nums_4", "DATE(504123025)", evaluator.SQLDate{Time: may042000}},
				{"sql_date_padded_nums_5", "DATE(1110123025)", evaluator.SQLDate{Time: nov102000}},
				{"sql_date_padded_nums_6", "DATE(61110123025)", evaluator.SQLDate{Time: nov102006}},
				{
					"sql_date_padded_nums_7",
					"DATE(61110123025.22)",
					evaluator.SQLDate{Time: nov102006},
				},
				{
					"sql_date_padded_nums_8",
					"DATE(1161110123025)",
					evaluator.SQLDate{Time: nov100116},
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
				{"sql_date_cutoff_0", "DATE('69-12-31')", evaluator.SQLDate{Time: preCutoff}},
				{"sql_date_cutoff_1", "DATE('70-01-01')", evaluator.SQLDate{Time: postCutoff}},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: date_add", func(t *testing.T) {
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
					evaluator.SQLTimestamp{Time: t0},
				},

				{
					"sql_date_add_1",
					"DATE_ADD('2003-01-02 10:28:05.500000', INTERVAL '2:2:3.5' DAY_MICROSECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_2",
					"DATE_ADD('2003-01-02 10:28:05.500000', INTERVAL '2:2:3.5' HOUR_MICROSECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_3",
					"DATE_ADD('2002-12-31 10:27:05', INTERVAL '2 2:3:4' DAY_SECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},

				{
					"sql_date_add_4",
					"DATE_ADD('2003-01-02 12:27:04.500000', INTERVAL '3:4.5' DAY_MICROSECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_5",
					"DATE_ADD('2003-01-02 12:27:04.500000', INTERVAL '3:4.5' HOUR_MICROSECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_6",
					"DATE_ADD('2003-01-02 12:27:04.500000', INTERVAL '3:4.5' MINUTE_MICROSECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_7",
					"DATE_ADD('2003-01-02 10:27:05', INTERVAL '2:3:4' DAY_SECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_8",
					"DATE_ADD('2003-01-02 10:27:05', INTERVAL '2:3:4' HOUR_SECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_9",
					"DATE_ADD('2002-12-31 10:27:09', INTERVAL '2 2:3' DAY_MINUTE)",
					evaluator.SQLTimestamp{Time: t0},
				},

				{
					"sql_date_add_10",
					"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' DAY_MICROSECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_11",
					"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' HOUR_MICROSECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_12",
					"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' MINUTE_MICROSECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_13",
					"DATE_ADD('2003-01-02 12:30:04.500000', INTERVAL '4.5' SECOND_MICROSECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_14",
					"DATE_ADD('2003-01-02 12:32:10', INTERVAL '-2:1' DAY_SECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_15",
					"DATE_ADD('2003-01-02 12:32:10', INTERVAL '-2:1' HOUR_SECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_16",
					"DATE_ADD('2003-01-02 12:32:10', INTERVAL '-2:1' MINUTE_SECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_17",
					"DATE_ADD('2003-01-02 15:32:09', INTERVAL '-3:2' DAY_MINUTE)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_18",
					"DATE_ADD('2003-01-02 15:32:09', INTERVAL '-3:2' HOUR_MINUTE)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_19",
					"DATE_ADD('2002-12-31 10:30:09', INTERVAL '2 2' DAY_HOUR)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_20",
					"DATE_ADD('2000-09-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)",
					evaluator.SQLTimestamp{Time: t0},
				},

				{
					"sql_date_add_21",
					"DATE_ADD('2002-01-02', INTERVAL NULL YEAR)",
					evaluator.SQLNull,
				},
				{
					"sql_date_add_22",
					"DATE_ADD(NULL, INTERVAL 1 YEAR)",
					evaluator.SQLNull,
				},
				{
					"sql_date_add_23",
					"DATE_ADD('2002-01-02', INTERVAL 1 YEAR)",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_date_add_24",
					"DATE_ADD('2003-08-31', INTERVAL 1 QUARTER)",
					evaluator.SQLTimestamp{Time: d2},
				},
				{
					"sql_date_add_25",
					"DATE_ADD('2003-10-31', INTERVAL 1 MONTH)",
					evaluator.SQLTimestamp{Time: d2},
				},
				{
					"sql_date_add_26",
					"DATE_ADD('2003-01-01', INTERVAL 1 DAY)",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_date_add_27",
					"DATE_ADD('2003-01-02 14:30:09', INTERVAL -2 HOUR)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_28",
					"DATE_ADD('2003-01-02 12:23:09', INTERVAL 7 MINUTE)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_29",
					"DATE_ADD('2003-01-02 12:30:12', INTERVAL -3 SECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_date_add_30",
					"DATE_ADD('2003-01-02 12:30:08.999999', INTERVAL 1 MICROSECOND)",
					evaluator.SQLTimestamp{Time: t0},
				},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: date_sub, subdate", func(t *testing.T) {
			d, err := time.Parse("2006-01-02", "2003-01-02")
			req.Nil(err, "unable to parse time from string")
			t1, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
			req.Nil(err, "unable to parse time from string")
			t2, err := time.Parse("2006-01-02 15:04:05", "2007-12-02 12:00:00")
			req.Nil(err, "unable to parse time from string")
			d2, err := time.Parse("2006-01-02", "2003-11-30")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_date_sub_0", "DATE_SUB('2004-01-02', INTERVAL NULL YEAR)", evaluator.SQLNull},
				{"sql_date_sub_1", "DATE_SUB(NULL, INTERVAL 1 YEAR)", evaluator.SQLNull},
				{
					"sql_date_sub_2",
					"DATE_SUB('2004-01-02', INTERVAL 1 YEAR)",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_date_sub_3",
					"DATE_SUB('2003-04-02', INTERVAL 1 QUARTER)",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_date_sub_4",
					"DATE_SUB('2003-12-31', INTERVAL 1 MONTH)",
					evaluator.SQLTimestamp{Time: d2},
				},
				{
					"sql_date_sub_5",
					"DATE_SUB('2003-01-03', INTERVAL 1 DAY)",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_date_sub_6",
					"SUBDATE('2004-01-02', INTERVAL 1 YEAR)",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_date_sub_7",
					"SUBDATE('2003-04-02', INTERVAL 1 QUARTER)",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_date_sub_8",
					"SUBDATE('2003-12-31', INTERVAL 1 MONTH)",
					evaluator.SQLTimestamp{Time: d2},
				},
				{
					"sql_date_sub_9",
					"SUBDATE('2008-01-02 12:00:00', 31)",
					evaluator.SQLTimestamp{Time: t2},
				},
				{
					"sql_date_sub_10",
					"SUBDATE('2016-01-02 12:00:00', 2953)",
					evaluator.SQLTimestamp{Time: t2},
				},
				{
					"sql_date_sub_11",
					"DATE_SUB('2003-01-02 10:30:09', INTERVAL -2 HOUR)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_12",
					"DATE_SUB('2003-01-02 12:37:09', INTERVAL 7 MINUTE)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_13",
					"DATE_SUB('2003-01-02 12:30:12', INTERVAL 3 SECOND)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_14",
					"DATE_SUB('2003-01-02 12:32:10', INTERVAL '2:1' MINUTE_SECOND)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_15",
					"DATE_SUB('2003-01-02 19:33:12', INTERVAL '7:3:3' HOUR_SECOND)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_16",
					"DATE_SUB('2003-01-02 15:32:09', INTERVAL '3:2' HOUR_MINUTE)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_17",
					"DATE_SUB('2003-01-04 14:33:13', INTERVAL '2 2:3:4' DAY_SECOND)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_18",
					"DATE_SUB('2003-01-04 14:33:09', INTERVAL '2 2:3' DAY_MINUTE)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_19",
					"DATE_SUB('2003-01-03 16:30:09', INTERVAL '1 4' DAY_HOUR)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_20",
					"DATE_SUB('2005-05-02 12:30:09', INTERVAL '2-4' YEAR_MONTH)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_21",
					"DATE_SUB('2003-01-02 12:33:09', INTERVAL '3' HOUR_MINUTE)",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_date_sub_22",
					"DATE_SUB('2003-01-02 14:32:12', INTERVAL '2 2:3' DAY_SECOND)",
					evaluator.SQLTimestamp{Time: t1},
				},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: from_days", func(t *testing.T) {
			t1 := time.Date(0001, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)
			t2 := time.Date(2000, 7, 3, 0, 0, 0, 0, schema.DefaultLocale)
			t3 := time.Date(10000, 3, 15, 0, 0, 0, 0, schema.DefaultLocale)
			t4 := time.Date(0005, 6, 29, 0, 0, 0, 0, schema.DefaultLocale)
			t5 := time.Date(2112, 1, 8, 0, 0, 0, 0, schema.DefaultLocale)

			tests := []test{
				{"sql_from_days_0", "FROM_DAYS(NULL)", evaluator.SQLNull},
				{"sql_from_days_1", "FROM_DAYS('sdg')", evaluator.SQLVarchar("0000-00-00")},
				{"sql_from_days_2", "FROM_DAYS(1.23)", evaluator.SQLVarchar("0000-00-00")},
				{"sql_from_days_3", "FROM_DAYS(-1.23)", evaluator.SQLVarchar("0000-00-00")},
				{"sql_from_days_4", "FROM_DAYS(-223.33)", evaluator.SQLVarchar("0000-00-00")},
				{"sql_from_days_5", "FROM_DAYS(223.33)", evaluator.SQLVarchar("0000-00-00")},
				{"sql_from_days_6", "FROM_DAYS(365.33)", evaluator.SQLVarchar("0000-00-00")},
				{"sql_from_days_7", "FROM_DAYS(3652499.5)", evaluator.SQLVarchar("0000-00-00")},
				{"sql_from_days_8", "FROM_DAYS(-771399.216)", evaluator.SQLVarchar("0000-00-00")},
				{"sql_from_days_9", "FROM_DAYS(365.93)", evaluator.SQLDate{Time: t1}},
				{"sql_from_days_10", "FROM_DAYS(343+23)", evaluator.SQLDate{Time: t1}},
				{"sql_from_days_11", "FROM_DAYS(730669)", evaluator.SQLDate{Time: t2}},
				{"sql_from_days_12", "FROM_DAYS(3652499.3)", evaluator.SQLDate{Time: t3}},
				{"sql_from_days_13", "FROM_DAYS('2006-05-11')", evaluator.SQLDate{Time: t4}},
				{"sql_from_days_14", "FROM_DAYS(771399.216)", evaluator.SQLDate{Time: t5}},
			}

			runTests(t, evalCtx, tests)
		})

		t.Run("subject: greatest", func(t *testing.T) {
			d, err := time.Parse("2006-01-02", "2006-05-11")
			req.Nil(err, "unable to parse time from string")
			t0, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 12:32:23")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_greatest_expr_0", "GREATEST(NULL, 1, 2)", evaluator.SQLNull},
				{"sql_greatest_expr_1", "GREATEST(1,3,2)", evaluator.SQLInt(3)},
				{
					"sql_greatest_expr_2",
					"GREATEST(2,2.3)",
					evaluator.SQLDecimal128(decimal.New(23, -1)),
				},
				{
					"sql_greatest_expr_3",
					"GREATEST('cats', '4', '2')",
					evaluator.SQLVarchar("cats"),
				},
				{
					"sql_greatest_expr_4",
					"GREATEST('dog', 'cats', 'bird')",
					evaluator.SQLVarchar("dog"),
				},
				{
					"sql_greatest_expr_5",
					"GREATEST('cat', 'bird', 2)",
					evaluator.SQLInt(2),
				},
				{
					"sql_greatest_expr_6",
					"GREATEST('cat', 2.2)",
					evaluator.SQLDecimal128(decimal.New(22, -1)),
				},
				{
					"sql_greatest_expr_7",
					"GREATEST(false, true)",
					evaluator.SQLTrue,
				},
				{
					"sql_greatest_expr_8",
					"GREATEST(DATE '2005-05-11', DATE '2006-05-11', DATE '2000-05-11')",
					evaluator.SQLDate{Time: d},
				},
				{
					"sql_greatest_expr_9",
					"GREATEST(DATE '2006-05-11', 14, 4235)",
					evaluator.SQLInt(20060511),
				},
				{
					"sql_greatest_expr_10",
					"GREATEST(DATE '2006-05-11', 14, 20080622)",
					evaluator.SQLInt(20080622),
				},
				{
					"sql_greatest_expr_11",
					"GREATEST(DATE '2006-05-11', 14, 20080622.1)",
					evaluator.SQLDecimal128(decimal.New(200806221, -1)),
				},
				{
					"sql_greatest_expr_12",
					"GREATEST(DATE '2006-05-11', 14, 4235.2)",
					evaluator.SQLDecimal128(decimal.New(20060511, 0)),
				},
				{
					"sql_greatest_expr_13",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', DATE '2006-05-11')",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_greatest_expr_14",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 12, 345)",
					evaluator.SQLInt(20060511123223),
				},
				{
					"sql_greatest_expr_15",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 20080923124345.3)",
					evaluator.SQLDecimal128(decimal.New(200809231243453, -1)),
				},
				{
					"sql_greatest_expr_16",
					"GREATEST(DATE '2006-05-11', 'cat', '2007-04-11')",
					evaluator.SQLVarchar("2007-04-11"),
				},
				{
					"sql_greatest_expr_17",
					"GREATEST(DATE '2006-05-11', 20080912, '2007-04-11')",
					evaluator.SQLInt(20080912),
				},
				{
					"sql_greatest_expr_18",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', TIMESTAMP '2006-05-11 10:32:45')",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_greatest_expr_19",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', 20080913, DATE '2007-08-23')",
					evaluator.SQLInt(20060511123223),
				},
				{
					"sql_greatest_expr_20",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', '2008-09-13')",
					evaluator.SQLVarchar("2008-09-13"),
				},
				{
					"sql_greatest_expr_21",
					"GREATEST(TIMESTAMP '2006-05-11 12:32:23', '2005-09-13')",
					evaluator.SQLTimestamp{Time: t0},
				},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: last_day", func(t *testing.T) {
			d1, err := time.Parse("2006-01-02", "2003-02-28")
			req.Nil(err, "unable to parse time from string")
			d2, err := time.Parse("2006-01-02", "2004-02-29")
			req.Nil(err, "unable to parse time from string")
			d3, err := time.Parse("2006-01-02", "2004-01-31")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_last_day_0", "LAST_DAY('')", evaluator.SQLNull},
				{"sql_last_day_1", "LAST_DAY(NULL)", evaluator.SQLNull},
				{"sql_last_day_2", "LAST_DAY('2003-03-32')", evaluator.SQLNull},
				{"sql_last_day_3", "LAST_DAY('2003-02-05')", evaluator.SQLDate{Time: d1}},
				{"sql_last_day_4", "LAST_DAY('2004-02-05')", evaluator.SQLDate{Time: d2}},
				{"sql_last_day_5", "LAST_DAY('2004-01-01 01:01:01')", evaluator.SQLDate{Time: d3}},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: least", func(t *testing.T) {
			d, err := time.Parse("2006-01-02", "2005-05-11")
			req.Nil(err, "unable to parse time from string")
			t0, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 00:00:00")
			req.Nil(err, "unable to parse time from string")
			t1, err := time.Parse("2006-01-02 15:04:05", "2006-05-11 10:32:23")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_least_expr_0", "LEAST(NULL, 1, 2)", evaluator.SQLNull},
				{"sql_least_expr_1", "LEAST(1,3,2)", evaluator.SQLInt(1)},
				{"sql_least_expr_2", "LEAST(2,2.3)", evaluator.SQLDecimal128(decimal.New(2, 0))},
				{"sql_least_expr_3", "LEAST('cats', '4', '2')", evaluator.SQLVarchar("2")},
				{"sql_least_expr_4", "LEAST('dog', 'cats', 'bird')", evaluator.SQLVarchar("bird")},
				{"sql_least_expr_5", "LEAST(false, true)", evaluator.SQLFalse},
				{
					"sql_least_expr_6",
					"LEAST(DATE '2005-05-11', DATE '2006-05-11', DATE '2007-05-11')",
					evaluator.SQLDate{Time: d},
				},
				{
					"sql_least_expr_7",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', DATE '2006-05-11')",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_least_expr_8",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', TIMESTAMP '2006-05-11 10:32:23')",
					evaluator.SQLTimestamp{Time: t1},
				},
				{"sql_least_expr_9", "LEAST('cat', 'bird', 2)", evaluator.SQLInt(0)},
				{"sql_least_expr_10", "LEAST('cat', 2.2)", evaluator.SQLDecimal128(decimal.Zero)},
				{"sql_least_expr_11", "LEAST(DATE '2006-05-11', 14, 4235)", evaluator.SQLInt(14)},
				{
					"sql_least_expr_12",
					"LEAST(DATE '2006-05-11', 14, 20080622.1)",
					evaluator.SQLDecimal128(decimal.New(14, 0)),
				},
				{
					"sql_least_expr_13",
					"LEAST(DATE '2006-05-11', 14, 4235.2)",
					evaluator.SQLDecimal128(decimal.New(14, 0)),
				},
				{
					"sql_least_expr_14",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', 12, 345)",
					evaluator.SQLInt(12),
				},
				{
					"sql_least_expr_15",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', 20080923124345.3)",
					evaluator.SQLDecimal128(decimal.New(20060511123223, 0)),
				},
				{
					"sql_least_expr_16",
					"LEAST(DATE '2006-05-11', 'cat', '2007-04-11')",
					evaluator.SQLVarchar("cat"),
				},
				{
					"sql_least_expr_17",
					"LEAST(DATE '2006-05-11', 20080912, '2007-04-11')",
					evaluator.SQLInt(0),
				},
				{
					"sql_least_expr_18",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', 20080913, DATE '2007-08-23')",
					evaluator.SQLInt(20070823),
				},
				{
					"sql_least_expr_19",
					"LEAST(TIMESTAMP '2006-05-11 10:32:23', '2008-09-13')",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_least_expr_20",
					"LEAST(TIMESTAMP '2006-05-11 12:32:23', '2005-09-13')",
					evaluator.SQLVarchar("2005-09-13"),
				},
			}
			runTests(t, evalCtx, tests)

		})

		t.Run("subject: makedate", func(t *testing.T) {
			d, err := time.Parse("2006-01-02", "2000-02-01")
			req.Nil(err, "unable to parse time from string")
			d1, err := time.Parse("2006-01-02", "2012-02-01")
			req.Nil(err, "unable to parse time from string")
			d2, err := time.Parse("2006-01-02", "1977-03-07")
			req.Nil(err, "unable to parse time from string")
			d3, err := time.Parse("2006-01-02", "0100-02-01")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_makedate_0", "MAKEDATE(NULL, 4)", evaluator.SQLNull},
				{"sql_makedate_1", "MAKEDATE(2004, 0)", evaluator.SQLNull},
				{"sql_makedate_2", "MAKEDATE(9999, 370)", evaluator.SQLNull},
				{"sql_makedate_3", "MAKEDATE('sdg', 32)", evaluator.SQLDate{Time: d}},
				{"sql_makedate_4", "MAKEDATE('2000.9', 32)", evaluator.SQLDate{Time: d}},
				{"sql_makedate_5", "MAKEDATE(1999.5, 32)", evaluator.SQLDate{Time: d}},
				{"sql_makedate_6", "MAKEDATE('2000.9', '32.9')", evaluator.SQLDate{Time: d}},
				{"sql_makedate_7", "MAKEDATE(1999.5, 31.5)", evaluator.SQLDate{Time: d}},
				{"sql_makedate_8", "MAKEDATE(2000, 32)", evaluator.SQLDate{Time: d}},
				{"sql_makedate_9", "MAKEDATE(12, 32)", evaluator.SQLDate{Time: d1}},
				{"sql_makedate_10", "MAKEDATE(77, 66)", evaluator.SQLDate{Time: d2}},
				{"sql_makedate_11", "MAKEDATE(99.5, 31.5)", evaluator.SQLDate{Time: d3}},
				{"sql_makedate_12", "MAKEDATE('100.9', '32.5')", evaluator.SQLDate{Time: d3}},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: str_to_date", func(t *testing.T) {
			d, err := time.Parse("2006-01-02", "2016-04-03")
			req.Nil(err, "unable to parse time from string")
			t0, err := time.Parse("2006-01-02 15:04:05", "2016-04-03 12:22:22")
			req.Nil(err, "unable to parse time from string")
			t1, err := time.Parse("2006-01-02 15:04:05", "2005-04-02 00:12:00")
			req.Nil(err, "unable to parse time from string")
			t2, err := time.Parse("2006-01-02 15:04:05", "2016-04-03 12:22:00")
			req.Nil(err, "unable to parse time from string")

			tests := []test{
				{"sql_str_to_date_0", "STR_TO_DATE(NULL, 4)", evaluator.SQLNull},
				{"sql_str_to_date_1", "STR_TO_DATE('foobarbar', NULL)", evaluator.SQLNull},
				{
					"sql_str_to_date_2",
					"STR_TO_DATE('2016-04-03','%Y-%m-%d')",
					evaluator.SQLDate{Time: d},
				},
				{
					"sql_str_to_date_3",
					"STR_TO_DATE('04,03,2016', '%m,%d,%Y')",
					evaluator.SQLDate{Time: d},
				},
				{
					"sql_str_to_date_4",
					"STR_TO_DATE('04,03,a16', '%m,%d,a%y')",
					evaluator.SQLDate{Time: d},
				},
				{
					"sql_str_to_date_5",
					"STR_TO_DATE('2016-04-03 12:22:22', '%Y-%m-%d %H:%i:%s')",
					evaluator.SQLTimestamp{Time: t0},
				},
				{
					"sql_str_to_date_6",
					"STR_TO_DATE('2016-04-03 12:22', '%Y-%m-%d %H:%i')",
					evaluator.SQLTimestamp{Time: t2},
				},
				{
					"sql_str_to_date_7",
					"STR_TO_DATE('2005-04-02 12', '%Y-%m-%d %i')",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_str_to_date_8",
					"STR_TO_DATE('Apr 03, 2016', '%b %d, %Y')",
					evaluator.SQLDate{Time: d},
				},
				{
					"sql_str_to_date_9",
					"STR_TO_DATE('Tue 2016-04-03', '%a %Y-%m-%d')",
					evaluator.SQLDate{Time: d},
				},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: timestamp", func(t *testing.T) {
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
				{"sql_timestamp_0", "TIMESTAMP(NULL)", evaluator.SQLNull},
				{"sql_timestamp_1", "TIMESTAMP(NULL, NULL)", evaluator.SQLNull},
				{"sql_timestamp_2", "TIMESTAMP(NULL, '12:22:22')", evaluator.SQLNull},
				{"sql_timestamp_3", "TIMESTAMP('2002-01-02', NULL)", evaluator.SQLNull},
				{
					"sql_timestamp_4",
					"TIMESTAMP('2010-01-01 11:11:11', '11:71:11')",
					evaluator.SQLNull,
				},
				{
					"sql_timestamp_5",
					"TIMESTAMP('2010-01-01 11:11:11', '11:23:59.5232355')",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_timestamp_6",
					"TIMESTAMP('2010-01-01 11:11:11', '12:22.4:12')",
					evaluator.SQLTimestamp{Time: t2},
				},
				{
					"sql_timestamp_7",
					"TIMESTAMP('2003-12-31 12:00:00', '12:00:00')",
					evaluator.SQLTimestamp{Time: t3},
				},
				{"sql_timestamp_8", "TIMESTAMP(20031231)", evaluator.SQLTimestamp{Time: t4}},
				{"sql_timestamp_9", "TIMESTAMP('2003-12-31')", evaluator.SQLTimestamp{Time: t4}},
				{
					"sql_timestamp_10",
					"TIMESTAMP('2003-12-31 12:00:00', '12.3:10:30')",
					evaluator.SQLTimestamp{Time: t5},
				},
				{
					"sql_timestamp_11",
					"TIMESTAMP('2003-12-31 12:23:23')",
					evaluator.SQLTimestamp{Time: t6},
				},
				{
					"sql_timestamp_12",
					"TIMESTAMP('2010-01-01 11:11:11', '12212')",
					evaluator.SQLTimestamp{Time: t7},
				},
				{
					"sql_timestamp_13",
					"TIMESTAMP('2010-01-01 11:11:11', 12212)",
					evaluator.SQLTimestamp{Time: t7},
				},
			}
			runTests(t, evalCtx, tests)
		})

		t.Run("subject: timestampadd", func(t *testing.T) {
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
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_1",
					"TIMESTAMPADD(YEAR, 0.5, DATE '2002-01-02')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_2",
					"TIMESTAMPADD(QUARTER, 1, DATE '2002-10-02')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_3",
					"TIMESTAMPADD(QUARTER, 0.5, DATE '2002-10-02')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_4",
					"TIMESTAMPADD(MONTH, 1, DATE '2002-12-02')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_5",
					"TIMESTAMPADD(MONTH, 0.5, DATE '2002-12-02')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_6",
					"TIMESTAMPADD(WEEK, 1, DATE '2002-12-26')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_7",
					"TIMESTAMPADD(WEEK, 0.5, DATE '2002-12-26')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_8",
					"TIMESTAMPADD(DAY, 1, DATE '2003-01-01')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_9",
					"TIMESTAMPADD(DAY, 0.5, DATE '2003-01-01')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_10",
					"TIMESTAMPADD(HOUR, 1, DATE '2003-01-02')",
					evaluator.SQLTimestamp{Time: dt},
				},
				{
					"sql_timestampadd_11",
					"TIMESTAMPADD(HOUR, 0.5, DATE '2003-01-02')",
					evaluator.SQLTimestamp{Time: dt},
				},
				{
					"sql_timestampadd_12",
					"TIMESTAMPADD(MINUTE, 60, DATE '2003-01-02')",
					evaluator.SQLTimestamp{Time: dt},
				},
				{
					"sql_timestampadd_13",
					"TIMESTAMPADD(MINUTE, 59.5, DATE '2003-01-02')",
					evaluator.SQLTimestamp{Time: dt},
				},
				{
					"sql_timestampadd_14",
					"TIMESTAMPADD(SECOND, 3600, DATE '2003-01-02')",
					evaluator.SQLTimestamp{Time: dt},
				},
				// No round test for SECOND, SECOND is not rounded.
				{
					"sql_timestampadd_15",
					"TIMESTAMPADD(MICROSECOND, 15000, TIMESTAMP '2003-01-02 12:30:09')",
					evaluator.SQLTimestamp{Time: t2},
				},
				{
					"sql_timestampadd_16",
					"TIMESTAMPADD(DAY, 1, TIMESTAMP '2003-01-01 12:30:09')",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_timestampadd_17",
					"TIMESTAMPADD(WEEK, 2, TIMESTAMP '2002-12-19 12:30:09')",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_timestampadd_18",
					"TIMESTAMPADD(SQL_TSI_YEAR, 2, TIMESTAMP '2001-01-02 12:30:09')",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_timestampadd_19",
					"TIMESTAMPADD(SQL_TSI_QUARTER, 2, DATE '2002-07-02')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_20",
					"TIMESTAMPADD(SQL_TSI_MONTH, 1, TIMESTAMP '2002-12-02 12:30:09')",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_timestampadd_21",
					"TIMESTAMPADD(SQL_TSI_WEEK, 1, DATE '2002-12-26')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_22",
					"TIMESTAMPADD(SQL_TSI_DAY, 1, DATE '2003-01-01')",
					evaluator.SQLTimestamp{Time: d},
				},
				{
					"sql_timestampadd_23",
					"TIMESTAMPADD(SQL_TSI_HOUR, 1, TIMESTAMP '2003-01-02 11:30:09')",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_timestampadd_24",
					"TIMESTAMPADD(SQL_TSI_MINUTE, 1, TIMESTAMP '2003-01-02 12:29:09')",
					evaluator.SQLTimestamp{Time: t1},
				},
				{
					"sql_timestampadd_25",
					"TIMESTAMPADD(SQL_TSI_SECOND, 1, TIMESTAMP '2003-01-02 12:30:08')",
					evaluator.SQLTimestamp{Time: t1},
				},
			}
			runTests(t, evalCtx, tests)
		})

		//Skipping
		//t.Run("subject: year", func(t *testing.T) {
		//	tests := []test{
		//		{"sql_year_0", "YEAR(NULL)", evaluator.SQLNull},
		//		{"sql_year_1", "YEAR('sdg')", evaluator.SQLNull},
		//		{"sql_year_2", "YEAR('2016-1-01 10:23:52')", evaluator.SQLInt(53)},
		//	}
		//	runTests(t,evalCtx, tests)
		//})
		//
		//Skipping
		//t.Run("subject: yearweek", func(t *testing.T) {
		//	tests := []test{
		//		{"sql_yearweek_0", "YEARWEEK(NULL)", evaluator.SQLNull},
		//		{"sql_yearweek_1", "YEARWEEK('sdg')", evaluator.SQLNull},
		//		{"sql_yearweek_2", "YEARWEEK('2000-01-01')", evaluator.SQLInt(199252)},
		//		{"sql_yearweek_3", "YEARWEEK('2001-01-01')", evaluator.SQLInt(200053)},
		//		{"sql_yearweek_4", "YEARWEEK('2002-01-01')", evaluator.SQLInt(200152)},
		//		{"sql_yearweek_5", "YEARWEEK('2003-01-01')", evaluator.SQLInt(200252)},
		//		{"sql_yearweek_6", "YEARWEEK('2004-01-01')", evaluator.SQLInt(200352)},
		//		{"sql_yearweek_7", "YEARWEEK('2005-01-01')", evaluator.SQLInt(200452)},
		//		{"sql_yearweek_8", "YEARWEEK('2006-01-01')", evaluator.SQLInt(200601)},
		//		{"sql_yearweek_9", "YEARWEEK('2000-01-06')", evaluator.SQLInt(200001)},
		//		{"sql_yearweek_10", "YEARWEEK('2001-01-06')", evaluator.SQLInt(200053)},
		//		{"sql_yearweek_11", "YEARWEEK('2002-01-06')", evaluator.SQLInt(200201)},
		//		{"sql_yearweek_12", "YEARWEEK('2003-01-06')", evaluator.SQLInt(200301)},
		//		{"sql_yearweek_13", "YEARWEEK('2004-01-06')", evaluator.SQLInt(200401)},
		//		{"sql_yearweek_14", "YEARWEEK('2005-01-06')", evaluator.SQLInt(200501)},
		//		{"sql_yearweek_15", "YEARWEEK('2006-01-06')", evaluator.SQLInt(200601)},
		//		{"sql_yearweek_16", "YEARWEEK('2000-01-01',1)", evaluator.SQLInt(199252)},
		//		{"sql_yearweek_17", "YEARWEEK('2001-01-01',1)", evaluator.SQLInt(200101)},
		//		{"sql_yearweek_18", "YEARWEEK('2002-01-01',1)", evaluator.SQLInt(200201)},
		//		{"sql_yearweek_19", "YEARWEEK('2003-01-01',1)", evaluator.SQLInt(200301)},
		//		{"sql_yearweek_20", "YEARWEEK('2004-01-01',1)", evaluator.SQLInt(200401)},
		//		{"sql_yearweek_21", "YEARWEEK('2005-01-01',1)", evaluator.SQLInt(200453)},
		//		{"sql_yearweek_22", "YEARWEEK('2006-01-01',1)", evaluator.SQLInt(200552)},
		//		{"sql_yearweek_23", "YEARWEEK('2000-01-06',1)", evaluator.SQLInt(200001)},
		//		{"sql_yearweek_24", "YEARWEEK('2001-01-06',1)", evaluator.SQLInt(200101)},
		//		{"sql_yearweek_25", "YEARWEEK('2002-01-06',1)", evaluator.SQLInt(200201)},
		//		{"sql_yearweek_26", "YEARWEEK('2003-01-06',1)", evaluator.SQLInt(200301)},
		//		{"sql_yearweek_27", "YEARWEEK('2004-01-06',1)", evaluator.SQLInt(200402)},
		//		{"sql_yearweek_28", "YEARWEEK('2005-01-06',1)", evaluator.SQLInt(200501)},
		//		{"sql_yearweek_29", "YEARWEEK('2006-01-06',1)", evaluator.SQLInt(200601)},
		//	}
		//	runTests(t,evalCtx, tests)
		//})

		t.Run("subject: sqlsubquerycmpexpr", func(t *testing.T) {
			t.Run("should not evaluate if the subquery returns a different number of columns "+
				"than the left expression", func(t *testing.T) {

				rows := []evaluator.Row{
					{
						Data: evaluator.Values{
							{
								SelectID: 1,
								Database: "",
								Table:    "test",
								Name:     "a",
								Data:     evaluator.SQLInt(1),
							},
							{
								SelectID: 1,
								Database: "",
								Table:    "test",
								Name:     "b",
								Data:     evaluator.SQLInt(2),
							},
						},
					},
					{
						Data: evaluator.Values{
							{
								SelectID: 1,
								Database: "",
								Table:    "test",
								Name:     "a",
								Data:     evaluator.SQLInt(2),
							},
							{
								SelectID: 1,
								Database: "",
								Table:    "test",
								Name:     "b",
								Data:     evaluator.SQLInt(4),
							},
						},
					},
				}

				cs := evaluator.NewCacheStage(0, rows, nil, nil)
				subqExpr := evaluator.NewSQLSubqueryExpr(false, false, cs)

				// Single SQLValue in left, two in subquery
				subCmpExpr := evaluator.NewSQLSubqueryCmpExpr(0, evaluator.SQLInt(1), subqExpr, "")
				_, err := subCmpExpr.Evaluate(evalCtx)
				req.NotNil(err, "expected error in evaluation")

				// Three SQLValues in left, two in subquery
				left := &evaluator.SQLValues{
					Values: []evaluator.SQLValue{
						evaluator.SQLInt(1),
						evaluator.SQLInt(2),
						evaluator.SQLInt(3),
					},
				}
				subCmpExpr = evaluator.NewSQLSubqueryCmpExpr(0, left, subqExpr, "")
				_, err = subCmpExpr.Evaluate(evalCtx)
				req.NotNil(err, "expected error in evaluation")
			})
		})

		t.Run("subject: sqltupleexpr", func(t *testing.T) {
			t.Run("should evaluate all the expressions and return sqlvalues", func(t *testing.T) {
				subject := &evaluator.SQLTupleExpr{
					Exprs: []evaluator.SQLExpr{
						evaluator.SQLInt(10),
						evaluator.NewSQLAddExpr(evaluator.SQLInt(30), evaluator.SQLInt(12)),
					},
				}
				result, err := subject.Evaluate(evalCtx)
				req.Nil(err, "unable to evaluate sql expression")
				req.IsType(
					&evaluator.SQLValues{},
					result,
					"expected type did not match actual type of evaluated expression",
				)
				resultValues := result.(*evaluator.SQLValues)
				req.Equal(
					resultValues.Values[0],
					evaluator.SQLInt(10),
					"expected value did not match evaluated value of sql expr",
				)
				req.Equal(
					resultValues.Values[1],
					evaluator.SQLInt(42),
					"expected value did not match evaluated value of sql expr",
				)
			})
			t.Run("should evaluate to a single sqlvalue if it contains only one value",
				func(t *testing.T) {
					subject := &evaluator.SQLTupleExpr{
						Exprs: []evaluator.SQLExpr{evaluator.SQLInt(10)},
					}
					sqlInt, err := subject.Evaluate(evalCtx)
					req.Nil(err, "unable to evaluate sql expression")
					intResult := sqlInt.(evaluator.SQLInt)
					req.Equal(
						intResult,
						evaluator.SQLInt(10),
						"expected value did not match evaluated value of sql expr",
					)

					subject = &evaluator.SQLTupleExpr{
						Exprs: []evaluator.SQLExpr{evaluator.SQLVarchar("10")},
					}
					sqlVarchar, err := subject.Evaluate(evalCtx)
					req.Nil(err, "unable to evaluate sql expression")
					varcharResult := sqlVarchar.(evaluator.SQLVarchar)
					req.Equal(
						varcharResult,
						evaluator.SQLVarchar("10"),
						"expected value did not match evaluated value of sql expr",
					)
				})
		})

		t.Run("evaluator should error when an unknown variable is used", func(t *testing.T) {
			subject := evaluator.NewSQLVariableExpr(
				"blah",
				variable.SystemKind,
				variable.SessionScope,
				schema.SQLNone,
			)

			_, err := subject.Evaluate(evalCtx)
			req.NotNil(err, "expected error in evaluation")
		})
		//t.Run("subject: sqlunarytildeexpr", func(t *testing.T) {
		// TODO: I'm not convinced we have this correct.
		//})
	})
}

func TestSQLLikeExprConvertToPattern(t *testing.T) {
	test := func(syntax, expected string) {
		Convey(fmt.Sprintf("XXX LIKE '%s' should convert to pattern '%s'", syntax, expected),
			func() {
				pattern := evaluator.ConvertSQLValueToPattern(evaluator.SQLVarchar(syntax), '\\')
				So(pattern, ShouldEqual, expected)
			})
	}

	Convey("Subject: SQLLikeExpr.convertToPattern", t, func() {
		test("David", "^David$")
		test("Da\\vid", "^David$")
		test("Da\\\\vid", "^Da\\\\vid$")
		test("Da_id", "^Da.id$")
		test("Da\\_id", "^Da_id$")
		test("Da%d", "^Da.*d$")
		test("Da\\%d", "^Da%d$")
		test("Sto_. %ow", "^Sto.\\. .*ow$")
	})
}

func TestMatches(t *testing.T) {
	Convey("Subject: Matches", t, func() {

		evalCtx := evaluator.NewEvalCtx(nil, collation.Default)

		tests := [][]interface{}{
			{evaluator.SQLInt(124), true},
			{evaluator.SQLFloat(1235), true},
			{evaluator.SQLVarchar("512"), true},
			{evaluator.SQLInt(0), false},
			{evaluator.SQLFloat(0), false},
			{evaluator.SQLVarchar("000"), false},
			{evaluator.SQLVarchar("skdjbkjb"), false},
			{evaluator.SQLVarchar(""), false},
			{evaluator.SQLTrue, true},
			{evaluator.SQLFalse, false},
			{evaluator.NewSQLEqualsExpr(evaluator.SQLInt(42), evaluator.SQLInt(42)), true},
			{evaluator.NewSQLEqualsExpr(evaluator.SQLInt(42), evaluator.SQLInt(21)), false},
		}

		for _, t := range tests {
			Convey(fmt.Sprintf("Should evaluate %v(%T) to %v", t[0], t[0], t[1]), func() {
				m, err := evaluator.Matches(t[0].(evaluator.SQLExpr), evalCtx)
				So(err, ShouldBeNil)
				So(m, ShouldEqual, t[1])
			})
		}
	})
}

func TestNewSQLValue(t *testing.T) {

	type test struct {
		input    interface{}
		sqlType  schema.SQLType
		sqlValue evaluator.SQLValue
	}

	runTests := func(tests []test) {
		for _, t := range tests {
			Convey(fmt.Sprintf("converting %v (%T) to '%v' should yield %v (%T)", t.input, t.input,
				t.sqlType, t.sqlValue, t.sqlValue), func() {
				val, _ := evaluator.NewSQLValue(t.input, t.sqlType, schema.SQLNone)
				So(val.String(), ShouldEqual, t.sqlValue.String())
			})
		}
	}

	getDate := func(t time.Time) time.Time {
		y, m, d := t.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, schema.DefaultLocale)
	}

	var (
		intVal              = 3
		floatVal            = 3.13
		strFloatVal         = "3.23"
		timeVal             = time.Now().In(schema.DefaultLocale)
		timeValParsed       = getDate(timeVal)
		strTimeVal          = "2006-01-01 11:10:12"
		strTimeValParsed, _ = time.Parse("2006-1-2 15:4:5", strTimeVal)
		strTimeValDate      = getDate(strTimeValParsed)
		objectIDVal         = bson.NewObjectId()
		sqlVal              = evaluator.SQLInt(0)
		zeroTime            = time.Time{}
		defaultSQLDate      = evaluator.SQLDate{Time: zeroTime}
		bsonDecimal128, _   = bson.ParseDecimal128("1.5")
	)

	Convey("Subject: NewSQLValue", t, func() {

		Convey("Subject: SQLNull", func() {
			tests := []test{
				{nil, schema.SQLBoolean, evaluator.SQLNull},
				{nil, schema.SQLDate, evaluator.SQLNull},
				{nil, schema.SQLDecimal128, evaluator.SQLNull},
				{nil, schema.SQLFloat, evaluator.SQLNull},
				{nil, schema.SQLInt, evaluator.SQLNull},
				{nil, schema.SQLInt64, evaluator.SQLNull},
				{nil, schema.SQLNumeric, evaluator.SQLNull},
				{nil, schema.SQLObjectID, evaluator.SQLNull},
				{nil, schema.SQLVarchar, evaluator.SQLNull},
			}

			runTests(tests)

		})

		Convey("Subject: SQLValue", func() {
			tests := []test{
				{sqlVal, schema.SQLBoolean, evaluator.SQLFalse},
				{sqlVal, schema.SQLDate, defaultSQLDate},
				{sqlVal, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.NewFromFloat(0))},
				{sqlVal, schema.SQLFloat, evaluator.SQLFloat(0)},
				{sqlVal, schema.SQLInt, evaluator.SQLInt(0)},
				{sqlVal, schema.SQLInt64, evaluator.SQLUint64(0)},
				{sqlVal, schema.SQLNumeric, evaluator.SQLFloat(0)},
				{sqlVal, schema.SQLObjectID, evaluator.SQLObjectID(strconv.FormatInt(int64(sqlVal),
					10))},
				{sqlVal, schema.SQLVarchar, evaluator.SQLVarchar("0")},
				{sqlVal, schema.SQLNone, sqlVal},
			}

			runTests(tests)

		})

		Convey("Subject: SQLBoolean", func() {
			tests := []test{
				{false, schema.SQLBoolean, evaluator.SQLFalse},
				{true, schema.SQLBoolean, evaluator.SQLTrue},
				{floatVal, schema.SQLBoolean, evaluator.SQLBool(floatVal)},
				{0.0, schema.SQLBoolean, evaluator.SQLFalse},
				{objectIDVal, schema.SQLBoolean, evaluator.SQLTrue},
				{intVal, schema.SQLBoolean, evaluator.SQLBool(intVal)},
				{0, schema.SQLBoolean, evaluator.SQLFalse},
				{strFloatVal, schema.SQLBoolean, evaluator.SQLBool(3.23)},
				{"0.000", schema.SQLBoolean, evaluator.SQLFalse},
				{"1.0", schema.SQLBoolean, evaluator.SQLTrue},
				{strTimeVal, schema.SQLBoolean, evaluator.SQLFalse},
				{timeVal, schema.SQLBoolean, evaluator.SQLTrue},
			}

			runTests(tests)

		})

		Convey("Subject: SQLDate", func() {
			tests := []test{
				{false, schema.SQLDate, defaultSQLDate},
				{true, schema.SQLDate, defaultSQLDate},
				{floatVal, schema.SQLDate, defaultSQLDate},
				{0.0, schema.SQLDate, defaultSQLDate},
				{objectIDVal, schema.SQLDate, evaluator.SQLDate{Time: objectIDVal.Time()}},
				{intVal, schema.SQLDate, defaultSQLDate},
				{0, schema.SQLDate, defaultSQLDate},
				{strFloatVal, schema.SQLDate, defaultSQLDate},
				{"0.000", schema.SQLDate, defaultSQLDate},
				{"1.0", schema.SQLDate, defaultSQLDate},
				{strTimeVal, schema.SQLDate, evaluator.SQLDate{Time: strTimeValDate}},
				{timeVal, schema.SQLDate, evaluator.SQLDate{Time: timeValParsed}},
			}

			runTests(tests)

		})

		Convey("Subject: SQLDecimal128", func() {
			tests := []test{
				{false, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(0, 0))},
				{true, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(1, 0))},
				{floatVal, schema.SQLDecimal128,
					evaluator.SQLDecimal128(decimal.NewFromFloat(floatVal))},
				{0.0, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(0, 0))},
				{objectIDVal, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(0, 0))},
				{intVal, schema.SQLDecimal128,
					evaluator.SQLDecimal128(decimal.NewFromFloat(float64(intVal)))},
				{0, schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(0, 0))},
				{strFloatVal, schema.SQLDecimal128,
					evaluator.SQLDecimal128(decimal.NewFromFloat(floatVal + .1))},
				{"0.000", schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(0, 0))},
				{"1.0", schema.SQLDecimal128, evaluator.SQLDecimal128(decimal.New(1, 0))},
			}

			runTests(tests)

		})

		Convey("Subject: SQLFloat, SQLNumeric", func() {
			tests := []test{

				//
				// SQLFloat
				//
				{false, schema.SQLFloat, evaluator.SQLFloat(0.0)},
				{true, schema.SQLFloat, evaluator.SQLFloat(1.0)},
				{floatVal, schema.SQLFloat, evaluator.SQLFloat(floatVal)},
				{0.0, schema.SQLFloat, evaluator.SQLFloat(0.0)},
				{intVal, schema.SQLFloat, evaluator.SQLFloat(float64(intVal))},
				{0, schema.SQLFloat, evaluator.SQLFloat(0.0)},
				{strFloatVal, schema.SQLFloat, evaluator.SQLFloat(3.23)},
				{"0.000", schema.SQLFloat, evaluator.SQLFloat(0.0)},
				{"1.0", schema.SQLFloat, evaluator.SQLFloat(1.0)},
				{bsonDecimal128, schema.SQLFloat, evaluator.SQLFloat(1.5)},

				//
				// SQLNumeric
				//
				{false, schema.SQLNumeric, evaluator.SQLFloat(0.0)},
				{true, schema.SQLNumeric, evaluator.SQLFloat(1.0)},
				{floatVal, schema.SQLNumeric, evaluator.SQLFloat(floatVal)},
				{0.0, schema.SQLNumeric, evaluator.SQLFloat(0.0)},
				{intVal, schema.SQLNumeric, evaluator.SQLFloat(float64(intVal))},
				{0, schema.SQLNumeric, evaluator.SQLFloat(0.0)},
				{strFloatVal, schema.SQLNumeric, evaluator.SQLFloat(3.23)},
				{"0.000", schema.SQLNumeric, evaluator.SQLFloat(0.0)},
				{"1.0", schema.SQLNumeric, evaluator.SQLFloat(1.0)},
			}

			runTests(tests)

		})

		Convey("Subject: SQLInt, SQLInt64", func() {
			tests := []test{
				{false, schema.SQLInt, evaluator.SQLInt(0)},
				{true, schema.SQLInt, evaluator.SQLInt(1)},
				{floatVal, schema.SQLInt, evaluator.SQLInt(int64(floatVal))},
				{0.0, schema.SQLInt, evaluator.SQLInt(0)},
				{intVal, schema.SQLInt, evaluator.SQLInt(intVal)},
				{0, schema.SQLInt, evaluator.SQLInt(0)},
				{strFloatVal, schema.SQLInt, evaluator.SQLInt(3)},
				{"0.000", schema.SQLInt, evaluator.SQLInt(0)},
				{"1.0", schema.SQLInt, evaluator.SQLInt(1)},
			}

			runTests(tests)

		})

		Convey("Subject: SQLObjectID", func() {
			tests := []test{
				{false, schema.SQLObjectID, evaluator.SQLObjectID("0")},
				{true, schema.SQLObjectID, evaluator.SQLObjectID("1")},
				{floatVal, schema.SQLObjectID,
					evaluator.SQLObjectID(strconv.FormatFloat(floatVal, 'f', -1, 64))},
				{0.0, schema.SQLObjectID, evaluator.SQLObjectID("0")},
				{objectIDVal, schema.SQLObjectID, evaluator.SQLObjectID(objectIDVal.Hex())},
				{intVal, schema.SQLObjectID,
					evaluator.SQLObjectID(strconv.FormatInt(int64(intVal), 10))},
				{0, schema.SQLObjectID, evaluator.SQLObjectID("0")},
				{strFloatVal, schema.SQLObjectID, evaluator.SQLObjectID(strFloatVal)},
				{"0.000", schema.SQLObjectID, evaluator.SQLObjectID("0.000")},
				{"1.0", schema.SQLObjectID, evaluator.SQLObjectID("1.0")},
				{strTimeVal, schema.SQLObjectID, evaluator.SQLObjectID(strTimeVal)},
				{timeVal, schema.SQLObjectID,
					evaluator.SQLObjectID(bson.NewObjectIdWithTime(timeVal).Hex())},
			}

			runTests(tests)

		})

		Convey("Subject: SQLTimestamp", func() {
			tests := []test{
				{false, schema.SQLTimestamp, evaluator.SQLTimestamp{Time: zeroTime}},
				{true, schema.SQLTimestamp, evaluator.SQLTimestamp{Time: zeroTime}},
				{floatVal, schema.SQLTimestamp, evaluator.SQLTimestamp{Time: zeroTime}},
				{0.0, schema.SQLTimestamp, evaluator.SQLTimestamp{Time: zeroTime}},
				{objectIDVal, schema.SQLTimestamp,
					evaluator.SQLTimestamp{Time: objectIDVal.Time()}},
				{intVal, schema.SQLTimestamp, evaluator.SQLTimestamp{Time: zeroTime}},
				{0, schema.SQLTimestamp, evaluator.SQLTimestamp{Time: zeroTime}},
				{strFloatVal, schema.SQLTimestamp, evaluator.SQLTimestamp{Time: zeroTime}},
				{"0.000", schema.SQLTimestamp, evaluator.SQLTimestamp{Time: zeroTime}},
				{"1.0", schema.SQLTimestamp, evaluator.SQLTimestamp{Time: zeroTime}},
				{strTimeVal, schema.SQLTimestamp, evaluator.SQLTimestamp{Time: strTimeValParsed}},
				{timeVal, schema.SQLTimestamp, evaluator.SQLTimestamp{Time: timeVal}},
			}

			runTests(tests)

		})

		Convey("Subject: SQLUint64", func() {
			tests := []test{
				{false, schema.SQLUint64, evaluator.SQLUint64(0)},
				{true, schema.SQLUint64, evaluator.SQLUint64(1)},
				{floatVal, schema.SQLUint64, evaluator.SQLUint64(uint64(floatVal))},
				{0.0, schema.SQLUint64, evaluator.SQLUint64(0)},
				{intVal, schema.SQLUint64, evaluator.SQLUint64(uint64(intVal))},
				{0, schema.SQLUint64, evaluator.SQLUint64(0)},
				{strFloatVal, schema.SQLUint64, evaluator.SQLUint64(3)},
				{"0.000", schema.SQLUint64, evaluator.SQLUint64(0)},
				{"1.0", schema.SQLUint64, evaluator.SQLUint64(1)},
			}

			runTests(tests)

		})

		Convey("Subject: SQLVarchar", func() {
			tests := []test{
				{false, schema.SQLVarchar, evaluator.SQLVarchar("0")},
				{true, schema.SQLVarchar, evaluator.SQLVarchar("1")},
				{floatVal, schema.SQLVarchar,
					evaluator.SQLVarchar(strconv.FormatFloat(floatVal, 'f', -1, 64))},
				{0.0, schema.SQLVarchar, evaluator.SQLVarchar("0")},
				{objectIDVal, schema.SQLVarchar,
					evaluator.SQLVarchar(objectIDVal.Hex())},
				{intVal, schema.SQLVarchar,
					evaluator.SQLVarchar(strconv.FormatInt(int64(intVal), 10))},
				{0, schema.SQLVarchar, evaluator.SQLVarchar("0")},
				{strFloatVal, schema.SQLVarchar, evaluator.SQLVarchar(strFloatVal)},
				{"0.000", schema.SQLVarchar, evaluator.SQLVarchar("0.000")},
				{"1.0", schema.SQLVarchar, evaluator.SQLVarchar("1.0")},
				{strTimeVal, schema.SQLVarchar, evaluator.SQLVarchar(strTimeVal)},
				{timeVal, schema.SQLVarchar,
					evaluator.SQLVarchar(timeVal.Format(evaluator.DateTimeFormat))},
			}

			runTests(tests)

		})
	})

}

func TestNewSQLValueFromSQLColumnExpr(t *testing.T) {

	Convey("When creating a SQLValue with no column type specified calling "+
		"NewSQLValueFromSQLColumnExpr on a", t, func() {

		Convey("SQLValue should return the same object passed in", func() {
			v := evaluator.SQLTrue
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLBoolean,
				schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, v)
		})

		Convey("nil value should return SQLNull", func() {
			v, err := evaluator.NewSQLValueFromSQLColumnExpr(nil, schema.SQLNull, schema.MongoBool)
			So(err, ShouldBeNil)
			So(v, ShouldResemble, evaluator.SQLNull)
		})

		Convey("bson object id should return its string value", func() {
			v := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLVarchar,
				schema.MongoObjectID)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v.Hex())
		})

		Convey("string objects should return the string value", func() {
			v := "56a10dd56ce28a89a8ed6edb"
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLVarchar,
				schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("int objects should return the int value", func() {
			v1 := int(6)
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v1, schema.SQLInt,
				schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v1)

			v2 := int32(6)
			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(v2, schema.SQLInt,
				schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v2)

			v3 := uint32(6)
			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(v3, schema.SQLInt,
				schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v3)
		})

		Convey("float objects should return the float value", func() {
			v := float64(6.3)
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLFloat,
				schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldEqual, v)
		})

		Convey("time objects should return the appropriate value", func() {
			v := time.Date(2014, time.December, 31, 0, 0, 0, 0, schema.DefaultLocale)
			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLDate,
				schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(evaluator.SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, evaluator.SQLDate{Time: v})

			v = time.Date(2014, time.December, 31, 10, 0, 0, 0, schema.DefaultLocale)
			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(v, schema.SQLTimestamp,
				schema.MongoDate)
			So(err, ShouldBeNil)

			sqlTimestamp, ok := newV.(evaluator.SQLTimestamp)
			So(ok, ShouldBeTrue)
			So(sqlTimestamp, ShouldResemble, evaluator.SQLTimestamp{Time: v})
		})
	})

	Convey("When creating a SQLValue with a column type specified calling "+
		"NewSQLValueFromSQLColumnExpr on a", t, func() {

		Convey("a SQLVarchar/SQLVarchar column type should attempt to coerce to the "+
			"SQLVarchar type", func() {

			t := schema.SQLVarchar

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(t, schema.SQLVarchar,
				schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLVarchar(t))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(6, schema.SQLVarchar,
				schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLVarchar("6"))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(6.6, schema.SQLVarchar,
				schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLVarchar("6.6"))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLVarchar,
				schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLVarchar("6"))

			_id := bson.ObjectId("56a10dd56ce28a89a8ed6edb")
			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(_id, schema.SQLVarchar,
				schema.MongoObjectID)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLObjectID(_id.Hex()))

		})

		Convey("a SQLInt column type should attempt to coerce to the SQLInt type", func() {

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(true, schema.SQLInt,
				schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(1))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int(6), schema.SQLInt,
				schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int32(6), schema.SQLInt,
				schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLInt,
				schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(float64(6.6), schema.SQLInt,
				schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(string("6"), schema.SQLInt,
				schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLInt(6))

		})

		Convey("a SQLFloat column type should attempt to coerce to the SQLFloat type", func() {

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(true, schema.SQLFloat,
				schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(1))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int(6), schema.SQLFloat,
				schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int32(6), schema.SQLFloat,
				schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLFloat,
				schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(float64(6.6), schema.SQLFloat,
				schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(6.6))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(string("6.6"), schema.SQLFloat,
				schema.MongoString)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLFloat(6.6))

		})

		Convey("a SQLDecimal column type should attempt to coerce to the SQLDecimal type", func() {

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(true, schema.SQLDecimal128,
				schema.MongoBool)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(1)))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int(6), schema.SQLDecimal128,
				schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(6)))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int32(6), schema.SQLDecimal128,
				schema.MongoInt)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(6)))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(int64(6), schema.SQLDecimal128,
				schema.MongoInt64)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(6)))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(float64(6.6), schema.SQLDecimal128,
				schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(6.6)))

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(string("6.6"), schema.SQLDecimal128,
				schema.MongoFloat)
			So(err, ShouldBeNil)
			So(newV, ShouldResemble, evaluator.SQLDecimal128(decimal.NewFromFloat(6.6)))

		})

		Convey("a SQLDate column type should attempt to coerce to the SQLDate type", func() {

			// Time type
			v1 := time.Date(2014, time.May, 11, 0, 0, 0, 0, schema.DefaultLocale)
			v2 := time.Date(2014, time.May, 11, 10, 32, 12, 0, schema.DefaultLocale)

			newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v1, schema.SQLDate,
				schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok := newV.(evaluator.SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, evaluator.SQLDate{Time: v1})

			newV, err = evaluator.NewSQLValueFromSQLColumnExpr(v2, schema.SQLDate, schema.MongoDate)
			So(err, ShouldBeNil)

			sqlDate, ok = newV.(evaluator.SQLDate)
			So(ok, ShouldBeTrue)
			So(sqlDate, ShouldResemble, evaluator.SQLDate{Time: v1})

			// String type
			dates := []string{"2014-05-11", "2014-05-11 15:04:05", "2014-05-11 15:04:05.233"}

			for _, d := range dates {

				newV, err = evaluator.NewSQLValueFromSQLColumnExpr(d, schema.SQLDate,
					schema.MongoNone)
				So(err, ShouldBeNil)

				sqlDate, ok := newV.(evaluator.SQLDate)
				So(ok, ShouldBeTrue)
				So(sqlDate, ShouldResemble, evaluator.SQLDate{Time: v1})

			}

			// invalid dates and those outside valid range
			// should return the default date
			dates = []string{"2014-12-44-44", "10000-1-1"}

			for _, d := range dates {
				newV, err = evaluator.NewSQLValueFromSQLColumnExpr(d, schema.SQLDate,
					schema.MongoNone)
				So(err, ShouldBeNil)

				_, ok := newV.(evaluator.SQLFloat)
				So(ok, ShouldBeTrue)
			}
		})

		Convey("a SQLTimestamp column type should attempt to coerce to the SQLTimestamp type",
			func() {

				// Time type
				v1 := time.Date(2014, time.May, 11, 15, 4, 5, 0, schema.DefaultLocale)

				newV, err := evaluator.NewSQLValueFromSQLColumnExpr(v1, schema.SQLTimestamp,
					schema.MongoNone)
				So(err, ShouldBeNil)

				sqlTs, ok := newV.(evaluator.SQLTimestamp)
				So(ok, ShouldBeTrue)
				So(sqlTs, ShouldResemble, evaluator.SQLTimestamp{Time: v1})

				// String type
				newV, err = evaluator.NewSQLValueFromSQLColumnExpr("2014-05-11 15:04:05.000",
					schema.SQLTimestamp, schema.MongoNone)
				So(err, ShouldBeNil)

				sqlTs, ok = newV.(evaluator.SQLTimestamp)
				So(ok, ShouldBeTrue)
				So(sqlTs, ShouldResemble, evaluator.SQLTimestamp{Time: v1})

				// invalid dates should return the default date
				dates := []string{"2044-12-40", "1966-15-1", "43223-3223"}

				for _, d := range dates {
					newV, err = evaluator.NewSQLValueFromSQLColumnExpr(d, schema.SQLTimestamp,
						schema.MongoNone)
					So(err, ShouldBeNil)
					_, ok := newV.(evaluator.SQLFloat)
					So(ok, ShouldBeTrue)
				}
			})
	})
}

func TestReconcileSQLExpr(t *testing.T) {

	type test struct {
		sql             string
		reconciledLeft  evaluator.SQLExpr
		reconciledRight evaluator.SQLExpr
	}

	runTests := func(tests []test) {
		sc := evaluator.MustLoadSchema(testSchema3)
		for _, t := range tests {
			Convey(fmt.Sprintf("Reconciliation for %q", t.sql), func() {
				e, err := evaluator.GetSQLExpr(sc, dbOne, tableTwoName, t.sql)
				So(err, ShouldBeNil)
				left, right := evaluator.GetBinaryExprLeaves(e)
				left, right, err = evaluator.ReconcileSQLExprs(left, right)
				So(err, ShouldBeNil)
				So(left, ShouldResemble, t.reconciledLeft)
				So(right, ShouldResemble, t.reconciledRight)
			})
		}
	}

	exprConv := evaluator.NewSQLConvertExpr(evaluator.SQLVarchar("2010-01-01"), schema.SQLTimestamp,
		evaluator.SQLNone)
	exprA := evaluator.NewSQLColumnExpr(1, "test", "bar", "a", schema.SQLInt, schema.MongoInt)
	exprB := evaluator.NewSQLColumnExpr(1, "test", "bar", "b", schema.SQLInt, schema.MongoInt)
	exprG := evaluator.NewSQLColumnExpr(1, "test", "bar", "g", schema.SQLTimestamp,
		schema.MongoDate)

	Convey("Subject: reconcileSQLExpr", t, func() {
		exprTime, err := evaluator.NewSQLScalarFunctionExpr("current_timestamp",
			[]evaluator.SQLExpr{})
		So(err, ShouldBeNil)
		tests := []test{
			{"a = 3", exprA, evaluator.SQLInt(3)},
			{"g - '2010-01-01'", evaluator.NewSQLConvertExpr(exprG, schema.SQLDecimal128,
				evaluator.SQLNone), evaluator.NewSQLConvertExpr(evaluator.SQLVarchar("2010-01-01"),
				schema.SQLDecimal128, evaluator.SQLNone)},
			{"a in (3)", exprA, evaluator.SQLInt(3)},
			{"a in (2,3)", exprA, &evaluator.SQLTupleExpr{Exprs: []evaluator.SQLExpr{
				evaluator.SQLInt(2), evaluator.SQLInt(3)}}},
			{"(a) in (3)", exprA, evaluator.SQLInt(3)},
			{"(a,b) in (2,3)", &evaluator.SQLTupleExpr{Exprs: []evaluator.SQLExpr{exprA, exprB}},
				&evaluator.SQLTupleExpr{Exprs: []evaluator.SQLExpr{evaluator.SQLInt(2),
					evaluator.SQLInt(3)}}},
			{"g > '2010-01-01'", exprG, exprConv},
			{"a and b", exprA, exprB},
			{"a / b", exprA, exprB},
			{"'2010-01-01' and g", exprConv, exprG},
			{"g in ('2010-01-01',current_timestamp())", exprG, &evaluator.SQLTupleExpr{
				Exprs: []evaluator.SQLExpr{exprConv, exprTime}}},
			{"g in ('2010-01-01',current_timestamp)", exprG, &evaluator.SQLTupleExpr{
				Exprs: []evaluator.SQLExpr{exprConv, exprTime}}},
		}

		runTests(tests)
	})

}

func TestCompareTo(t *testing.T) {

	var (
		diff        = time.Duration(969 * time.Hour)
		sameDayDiff = time.Duration(1)
		now         = time.Now()
		oid1        = bson.NewObjectId().Hex()
		oid2        = bson.NewObjectId().Hex()
		oid3        = bson.NewObjectId().Hex()
	)

	Convey("Subject: CompareTo", t, func() {

		type test struct {
			left     evaluator.SQLValue
			right    evaluator.SQLValue
			expected int
		}

		runTests := func(tests []test) {
			for _, t := range tests {
				Convey(fmt.Sprintf("comparing '%v' (%T) to '%v' (%T) should return %v",
					t.left, t.left, t.right, t.right, t.expected), func() {
					compareTo, err := evaluator.CompareTo(t.left, t.right, collation.Default)
					So(err, ShouldBeNil)
					So(compareTo, ShouldEqual, t.expected)
				})
			}
		}

		Convey("Subject: SQLInt", func() {
			tests := []test{
				{evaluator.SQLInt(1), evaluator.SQLInt(0), 1},
				{evaluator.SQLInt(1), evaluator.SQLInt(1), 0},
				{evaluator.SQLInt(1), evaluator.SQLInt(2), -1},
				{evaluator.SQLInt(1), evaluator.SQLUint32(1), 0},
				{evaluator.SQLInt(1), evaluator.SQLFloat(1), 0},
				{evaluator.SQLInt(1), evaluator.SQLFalse, 1},
				{evaluator.SQLInt(1), evaluator.SQLTrue, 0},
				{evaluator.SQLInt(1), evaluator.SQLNull, 1},
				{evaluator.SQLInt(1), evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLInt(1), evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLInt(1), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, 0},
				{evaluator.SQLInt(1), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLInt(1), evaluator.SQLDate{Time: now}, -1},
				{evaluator.SQLInt(1), evaluator.SQLTimestamp{Time: now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLUint32", func() {
			tests := []test{
				{evaluator.SQLUint32(1), evaluator.SQLInt(0), 1},
				{evaluator.SQLUint32(1), evaluator.SQLInt(1), 0},
				{evaluator.SQLUint32(1), evaluator.SQLInt(2), -1},
				{evaluator.SQLUint32(1), evaluator.SQLUint32(1), 0},
				{evaluator.SQLUint32(1), evaluator.SQLFloat(1), 0},
				{evaluator.SQLUint32(1), evaluator.SQLFalse, 1},
				{evaluator.SQLUint32(1), evaluator.SQLTrue, 0},
				{evaluator.SQLUint32(1), evaluator.SQLNull, 1},
				{evaluator.SQLUint32(1), evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLUint32(1), evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLUint32(1), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, 0},
				{evaluator.SQLUint32(1), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLUint32(1), evaluator.SQLDate{Time: now}, -1},
				{evaluator.SQLUint32(1), evaluator.SQLTimestamp{Time: now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLUint64", func() {
			tests := []test{
				{evaluator.SQLUint64(1), evaluator.SQLInt(0), 1},
				{evaluator.SQLUint64(1), evaluator.SQLInt(1), 0},
				{evaluator.SQLUint64(1), evaluator.SQLInt(2), -1},
				{evaluator.SQLUint64(1), evaluator.SQLUint32(1), 0},
				{evaluator.SQLUint64(1), evaluator.SQLUint64(1), 0},
				{evaluator.SQLUint64(1), evaluator.SQLFloat(1), 0},
				{evaluator.SQLUint64(1), evaluator.SQLFalse, 1},
				{evaluator.SQLUint64(1), evaluator.SQLTrue, 0},
				{evaluator.SQLUint64(1), evaluator.SQLNull, 1},
				{evaluator.SQLUint64(1), evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLUint64(1), evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLUint64(1), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, 0},
				{evaluator.SQLUint64(1), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLUint64(1), evaluator.SQLDate{Time: now}, -1},
				{evaluator.SQLUint64(1), evaluator.SQLTimestamp{Time: now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLFloat", func() {
			tests := []test{
				{evaluator.SQLFloat(0.1), evaluator.SQLInt(0), 1},
				{evaluator.SQLFloat(1.1), evaluator.SQLInt(1), 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLInt(2), -1},
				{evaluator.SQLFloat(1.1), evaluator.SQLUint32(1), 1},
				{evaluator.SQLFloat(1.1), evaluator.SQLFloat(1), 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLFalse, 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLTrue, -1},
				{evaluator.SQLFloat(0.1), evaluator.SQLNull, 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLFloat(0.0), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, -1},
				{evaluator.SQLFloat(0.1), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLFloat(0.1), evaluator.SQLDate{Time: now}, -1},
				{evaluator.SQLFloat(0.1), evaluator.SQLTimestamp{Time: now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLBool", func() {
			tests := []test{
				{evaluator.SQLTrue, evaluator.SQLInt(0), 1},
				{evaluator.SQLTrue, evaluator.SQLInt(1), 0},
				{evaluator.SQLTrue, evaluator.SQLInt(2), -1},
				{evaluator.SQLTrue, evaluator.SQLUint32(1), 0},
				{evaluator.SQLTrue, evaluator.SQLFloat(1), 0},
				{evaluator.SQLTrue, evaluator.SQLFalse, 1},
				{evaluator.SQLTrue, evaluator.SQLTrue, 0},
				{evaluator.SQLTrue, evaluator.SQLNull, 1},
				{evaluator.SQLTrue, evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLTrue, evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLTrue, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, 0},
				{evaluator.SQLTrue, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLTrue, evaluator.SQLDate{Time: now}, -1},
				{evaluator.SQLTrue, evaluator.SQLTimestamp{Time: now}, -1},
				{evaluator.SQLFalse, evaluator.SQLInt(0), 0},
				{evaluator.SQLFalse, evaluator.SQLInt(1), -1},
				{evaluator.SQLFalse, evaluator.SQLInt(2), -1},
				{evaluator.SQLFalse, evaluator.SQLUint32(1), -1},
				{evaluator.SQLFalse, evaluator.SQLFloat(1), -1},
				{evaluator.SQLFalse, evaluator.SQLFalse, 0},
				{evaluator.SQLFalse, evaluator.SQLTrue, -1},
				{evaluator.SQLFalse, evaluator.SQLNull, 1},
				{evaluator.SQLFalse, evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 0},
				{evaluator.SQLFalse, evaluator.SQLVarchar("bac"), 0},
				{evaluator.SQLFalse, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, -1},
				{evaluator.SQLFalse, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLFalse, evaluator.SQLDate{Time: now}, -1},
				{evaluator.SQLFalse, evaluator.SQLTimestamp{Time: now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLDate", func() {
			tests := []test{
				{evaluator.SQLDate{Time: now}, evaluator.SQLInt(0), 1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLInt(1), 1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLInt(2), 1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLUint32(1), 1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLFloat(1), 1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLFalse, 1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLDate{Time: now.Add(diff)}, -1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLNull, 1},
				{evaluator.SQLDate{Time: now},
					evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLDate{Time: now}, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, 1},
				{evaluator.SQLDate{Time: now}, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLDate{Time: now.Add(-diff)}, 1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLTimestamp{Time: now.Add(diff)}, -1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLTimestamp{Time: now.Add(-diff)}, 1},
				{evaluator.SQLDate{Time: now}, evaluator.SQLDate{Time: now}, 0},
			}
			runTests(tests)
		})

		Convey("Subject: SQLTimestamp", func() {
			tests := []test{
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLInt(0), 1},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLInt(1), 1},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLInt(2), 1},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLUint32(1), 1},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLFloat(1), 1},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLFalse, 1},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLNull, 1},
				{evaluator.SQLTimestamp{Time: now},
					evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLVarchar("bac"), 1},
				{evaluator.SQLTimestamp{Time: now}, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, 1},
				{evaluator.SQLTimestamp{Time: now}, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLTimestamp{Time: now},
					evaluator.SQLTimestamp{Time: now.Add(diff)}, -1},
				{evaluator.SQLTimestamp{Time: now},
					evaluator.SQLTimestamp{Time: now.Add(-diff)}, 1},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLTimestamp{Time: now}, 0},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLDate{Time: now}, 0},
				{evaluator.SQLTimestamp{Time: now.Add(sameDayDiff)},
					evaluator.SQLDate{Time: now}, 1},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLDate{Time: now.Add(diff)}, -1},
				{evaluator.SQLTimestamp{Time: now}, evaluator.SQLDate{Time: now.Add(-diff)}, 1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLNullValue", func() {
			tests := []test{
				{evaluator.SQLNull, evaluator.SQLInt(0), -1},
				{evaluator.SQLNull, evaluator.SQLInt(1), -1},
				{evaluator.SQLNull, evaluator.SQLInt(2), -1},
				{evaluator.SQLNull, evaluator.SQLUint32(1), -1},
				{evaluator.SQLNull, evaluator.SQLFloat(1), -1},
				{evaluator.SQLNull, evaluator.SQLFalse, -1},
				{evaluator.SQLNull, evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), -1},
				{evaluator.SQLNull, evaluator.SQLVarchar("bac"), -1},
				{evaluator.SQLNull, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, -1},
				{evaluator.SQLNull, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLNull, &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNull}}, 0},
				{evaluator.SQLNull, evaluator.SQLDate{Time: now}, -1},
				{evaluator.SQLNull, evaluator.SQLTimestamp{Time: now}, -1},
				{evaluator.SQLNull, evaluator.SQLNull, 0},
			}
			runTests(tests)
		})

		Convey("Subject: SQLVarchar", func() {
			tests := []test{
				{evaluator.SQLVarchar("bac"), evaluator.SQLInt(0), 0},
				{evaluator.SQLVarchar("bac"), evaluator.SQLInt(1), -1},
				{evaluator.SQLVarchar("bac"), evaluator.SQLInt(2), -1},
				{evaluator.SQLVarchar("bac"), evaluator.SQLUint32(1), -1},
				{evaluator.SQLVarchar("bac"), evaluator.SQLFloat(1), -1},
				{evaluator.SQLVarchar("bac"), evaluator.SQLFalse, 0},
				{evaluator.SQLVarchar("bac"), evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 0},
				{evaluator.SQLVarchar("bac"), evaluator.SQLVarchar("cba"), -1},
				{evaluator.SQLVarchar("bac"), evaluator.SQLVarchar("bac"), 0},
				{evaluator.SQLVarchar("bac"), evaluator.SQLVarchar("abc"), 1},
				{evaluator.SQLVarchar("bac"), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, -1},
				{evaluator.SQLVarchar("bac"), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLVarchar("bac"), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLVarchar("bac")}}, 0},
			}
			runTests(tests)
		})

		Convey("Subject: SQLValues", func() {
			tests := []test{
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLInt(0), 1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLInt(1), 0},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLInt(2), -1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLUint32(1), 0},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLUint32(11), -1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLUint32(0), 1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLFloat(1.1), -1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLFloat(0.1), 1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLFalse, 1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLObjectID("56e0750e1d857aea925a4ba1"), 1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLVarchar("abc"), 1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLNone, 1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, 0},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(-1)}}, 1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(2)}}, -1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLDate{Time: now}, -1},
				{&evaluator.SQLValues{Values: []evaluator.SQLValue{evaluator.SQLInt(1)}},
					evaluator.SQLTimestamp{Time: now}, -1},
			}
			runTests(tests)
		})

		Convey("Subject: SQLObjectID", func() {

			tests := []test{
				{evaluator.SQLObjectID(oid2), evaluator.SQLInt(0), 0},
				{evaluator.SQLObjectID(oid2), evaluator.SQLUint32(1), -1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLFloat(1), -1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLVarchar("cba"), 0},
				{evaluator.SQLObjectID(oid2), evaluator.SQLFalse, 0},
				{evaluator.SQLObjectID(oid2), evaluator.SQLTrue, -1},
				{evaluator.SQLObjectID(oid2), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLInt(1)}}, -1},
				{evaluator.SQLObjectID(oid2), &evaluator.SQLValues{
					Values: []evaluator.SQLValue{evaluator.SQLNone}}, 1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLDate{Time: now}, -1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLTimestamp{Time: now}, -1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLObjectID(oid3), -1},
				{evaluator.SQLObjectID(oid2), evaluator.SQLObjectID(oid2), 0},
				{evaluator.SQLObjectID(oid2), evaluator.SQLObjectID(oid1), 1},
			}
			runTests(tests)
		})

	})
}

func TestIsTruthyIsFalsy(t *testing.T) {

	Convey("IsTruthy, IsFalsy", t, func() {
		d, err := time.Parse("2006-01-02", "2003-01-02")
		So(err, ShouldBeNil)
		t, err := time.Parse("2006-01-02 15:04:05", "2003-01-02 12:30:09")
		So(err, ShouldBeNil)

		Convey("Subject: IsTruthy", func() {
			truthy := evaluator.IsTruthy(evaluator.SQLTimestamp{Time: t})
			So(truthy, ShouldBeTrue)

			truthy = evaluator.IsTruthy(evaluator.SQLDate{Time: d})
			So(truthy, ShouldBeTrue)

			truthy = evaluator.IsTruthy(evaluator.SQLInt(0))
			So(truthy, ShouldBeFalse)

			truthy = evaluator.IsTruthy(evaluator.SQLInt(1))
			So(truthy, ShouldBeTrue)

			truthy = evaluator.IsTruthy(evaluator.SQLVarchar("dsf"))
			So(truthy, ShouldBeFalse)

			truthy = evaluator.IsTruthy(evaluator.SQLVarchar("16"))
			So(truthy, ShouldBeTrue)
		})

		Convey("Subject: IsFalsy", func() {
			truthy := evaluator.IsFalsy(evaluator.SQLTimestamp{Time: t})
			So(truthy, ShouldBeFalse)

			truthy = evaluator.IsFalsy(evaluator.SQLDate{Time: d})
			So(truthy, ShouldBeFalse)

			truthy = evaluator.IsFalsy(evaluator.SQLInt(0))
			So(truthy, ShouldBeTrue)

			truthy = evaluator.IsFalsy(evaluator.SQLInt(1))
			So(truthy, ShouldBeFalse)

			truthy = evaluator.IsFalsy(evaluator.SQLVarchar("dsf"))
			So(truthy, ShouldBeTrue)

			truthy = evaluator.IsFalsy(evaluator.SQLVarchar("16"))
			So(truthy, ShouldBeFalse)
		})
	})
}

func TestIsUUID(t *testing.T) {
	Convey("IsUUID", t, func() {
		So(evaluator.IsUUID(schema.MongoUUID), ShouldBeTrue)
		So(evaluator.IsUUID(schema.MongoUUIDCSharp), ShouldBeTrue)
		So(evaluator.IsUUID(schema.MongoUUIDJava), ShouldBeTrue)
		So(evaluator.IsUUID(schema.MongoUUIDOld), ShouldBeTrue)
		So(evaluator.IsUUID(schema.MongoString), ShouldBeFalse)
		So(evaluator.IsUUID(schema.MongoGeo2D), ShouldBeFalse)
		So(evaluator.IsUUID(schema.MongoObjectID), ShouldBeFalse)
		So(evaluator.IsUUID(schema.MongoBool), ShouldBeFalse)
		So(evaluator.IsUUID(schema.MongoInt), ShouldBeFalse)
		So(evaluator.IsUUID(schema.MongoInt64), ShouldBeFalse)
	})
}

func TestGetBinaryFromExpr(t *testing.T) {

	Convey("GetBinaryFromExpr", t, func() {

		expected := []byte{
			0x01, 0x02, 0x03, 0x04,
			0x05, 0x06, 0x07, 0x08,
			0x09, 0x0a, 0x0b, 0x0c,
			0x0d, 0x0e, 0x0f, 0x10,
		}

		Convey("Subject: invalid SQLExpr", func() {
			_, ok := evaluator.GetBinaryFromExpr(schema.MongoUUID, evaluator.SQLVarchar("3"))
			So(ok, ShouldBeFalse)
		})

		Convey("Subject: valid SQLExpr with dashes", func() {
			b, ok := evaluator.GetBinaryFromExpr(schema.MongoUUID,
				evaluator.SQLVarchar("01020304-0506-0708-090a-0b0c0d0e0f10"))
			So(ok, ShouldBeTrue)
			So(b.Kind, ShouldEqual, 0x04)
			So(b.Data, ShouldResemble, expected)

			b, ok = evaluator.GetBinaryFromExpr(schema.MongoUUIDOld,
				evaluator.SQLVarchar("01020304-0506-0708-090a-0b0c0d0e0f10"))
			So(ok, ShouldBeTrue)
			So(b.Kind, ShouldEqual, 0x03)
			So(b.Data, ShouldResemble, expected)
		})

		Convey("Subject: valid SQLExpr without dashes", func() {
			b, ok := evaluator.GetBinaryFromExpr(schema.MongoUUIDJava,
				evaluator.SQLVarchar("0807060504030201100f0e0d0c0b0a09"))
			So(ok, ShouldBeTrue)
			So(b.Kind, ShouldEqual, 0x03)
			So(b.Data, ShouldResemble, expected)

			b, ok = evaluator.GetBinaryFromExpr(schema.MongoUUIDCSharp,
				evaluator.SQLVarchar("0403020106050807090a0b0c0d0e0f10"))
			So(ok, ShouldBeTrue)
			So(b.Kind, ShouldEqual, 0x03)
			So(b.Data, ShouldResemble, expected)
		})
	})
}

func TestNormalizeUUID(t *testing.T) {

	Convey("NormalizeUUID", t, func() {
		expected := []byte{
			0x01, 0x02, 0x03, 0x04,
			0x05, 0x06, 0x07, 0x08,
			0x09, 0x0a, 0x0b, 0x0c,
			0x0d, 0x0e, 0x0f, 0x10,
		}

		Convey("Subject: standard UUID", func() {
			bytes := []byte{
				0x01, 0x02, 0x03, 0x04,
				0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c,
				0x0d, 0x0e, 0x0f, 0x10,
			}
			So(evaluator.NormalizeUUID(schema.MongoUUID, bytes), ShouldBeNil)
			So(bytes, ShouldResemble, expected)
		})

		Convey("Subject: old UUID", func() {
			bytes := []byte{
				0x01, 0x02, 0x03, 0x04,
				0x05, 0x06, 0x07, 0x08,
				0x09, 0x0a, 0x0b, 0x0c,
				0x0d, 0x0e, 0x0f, 0x10,
			}
			So(evaluator.NormalizeUUID(schema.MongoUUIDOld, bytes), ShouldBeNil)
			So(bytes, ShouldResemble, expected)
		})

		Convey("Subject: C# Legacy UUID", func() {
			bytes := []byte{
				0x04, 0x03, 0x02, 0x01,
				0x06, 0x05, 0x08, 0x07,
				0x09, 0x0a, 0x0b, 0x0c,
				0x0d, 0x0e, 0x0f, 0x10,
			}
			So(evaluator.NormalizeUUID(schema.MongoUUIDCSharp, bytes), ShouldBeNil)
			So(bytes, ShouldResemble, expected)
		})

		Convey("Subject: Java Legacy UUID", func() {
			bytes := []byte{
				0x08, 0x07, 0x06, 0x05,
				0x04, 0x03, 0x02, 0x01,
				0x10, 0x0f, 0x0e, 0x0d,
				0x0c, 0x0b, 0x0a, 0x09,
			}
			So(evaluator.NormalizeUUID(schema.MongoUUIDJava, bytes), ShouldBeNil)
			So(bytes, ShouldResemble, expected)
		})

	})
}
