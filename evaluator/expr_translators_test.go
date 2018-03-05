package evaluator_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	update = flag.Bool("update", false, "update cached results for the selected tests")
)

func TestTranslate(t *testing.T) {
	flag.Parse()
	req := require.New(t)

	type test struct {
		name string
		sql  string
	}

	// define the expressions we want to test, along with a readable name for each
	tests := []test{

		// scalar functions
		{"scalar_abs", "abs(a)"},
		{"scalar_abs_string", "abs(s)"},
		{"scalar_adddate_00", "adddate(date_format(g, '%Y-%m-%d %H:%i:%s'), interval -10 minute)"},
		{"scalar_adddate_01", "adddate(date_format(g, '%Y-%m-%d %H:%i:%s'), interval 0 second)"},
		{"scalar_adddate_02", "adddate(date_format(g, '%Y-%m-%d %H:%i:%s'), interval 10 second)"},
		{"scalar_adddate_03", "adddate(date_format(g, '%Y-%m-%d %H:%i:00'), interval 0 second)"},
		{"scalar_adddate_04", "adddate(date_format(g, '%Y-%m-%d %H:00:00'), interval 0 second)"},
		{"scalar_adddate_05", "adddate(date_format(g, '%Y-%m-%d 00:00:00'), interval 0 second)"},
		{"scalar_adddate_06", "adddate(date_format(g, '%Y-%m-%d'), interval 0 second)"},
		{"scalar_adddate_07", "adddate(date_format(g, '%Y-%m-01 00:00:00'), interval 0 second)"},
		{"scalar_adddate_08", "adddate(date_format(g, '%Y-%m-01'), interval 0 second)"},
		{"scalar_adddate_09", "adddate(date_format(g, '%Y-01-01 00:00:00'), interval 0 second)"},
		{"scalar_adddate_10", "adddate(date_format(g, '%Y-01-01'), interval 0 second)"},
		{"scalar_adddate_11", "adddate(g, INTERVAL 10 day)"},
		{"scalar_adddate_12", "adddate(g, null)"},
		{"scalar_ceil", "ceil(a)"},
		{"scalar_char_length", "char_length(s)"},
		{"scalar_character_length", "character_length(s)"},
		{"scalar_coalesce", "coalesce(null, null, 2, null, 3, null)"},
		{"scalar_concat", "concat(s, 'funny')"},
		{"scalar_concat_empty", "concat(s, '')"},
		{"scalar_concat_mixed_types", "concat(s, a)"},
		{"scalar_concat_null", "concat(s, null)"},
		{"scalar_concat_ws", "concat_ws(',', s)"},
		{"scalar_concat_ws_null", "concat_ws(',', s, null)"},
		{"scalar_cos", "cos(a)"},
		{"scalar_cot", "cot(a)"},
		{"scalar_count_0", "count(*)"},
		{"scalar_count_1", "count(a + b)"},
		{"scalar_date", "date(s)"},
		{"scalar_date_add", "date_add(g, INTERVAL 10 day)"},
		{"scalar_date_format", "date_format(g, '%% %d %f %H %k %i %j %m %s %S %T %U %Y')"},
		{"scalar_date_format_0", "date_format(g, '%Y%')"},
		{"scalar_date_format_1", "date_format(g, '%V')"},
		{"scalar_date_format_null", "date_format(g, null)"},
		{"scalar_date_sub", "date_sub(g, INTERVAL 10 day)"},
		{"scalar_degrees", "degrees(a)"},
		{"scalar_datediff_0", "datediff(g, '2000-09-01 13:29:18')"},
		{"scalar_datediff_1", "datediff(g, h)"},
		{"scalar_datediff_mixed_types_bool", "datediff(t, '2000-09-01 13:29:18')"},
		{"scalar_datediff_mixed_types_int", "datediff(a, '2000-09-01 13:29:18')"},
		{"scalar_datediff_mixed_types_str", "datediff(s, '2000-09-01 13:29:18')"},
		{"scalar_dayname", "dayname(g)"},
		{"scalar_dayname_int", "dayname(a)"},
		{"scalar_dayofmonth", "dayofmonth(g)"},
		{"scalar_dayofweek", "dayofweek(g)"},
		{"scalar_dayofyear", "dayofyear(g)"},
		{"scalar_exp", "exp(a)"},
		{"scalar_extract", "extract(year from g)"},
		{"scalar_elt", "elt(1, a, a)"},
		{"scalar_floor", "floor(a)"},
		{"scalar_from_unixtime", "from_unixtime(a)"},
		{"scalar_from_unixtime_w_format", "from_unixtime(a, '%Y%m%d%H%i%s')"},
		{"scalar_from_days", "from_days(a)"},
		{"scalar_greatest", "greatest(a, 2)"},
		{"scalar_greatest_mixed_types_0", "greatest(a, 'test')"},
		{"scalar_greatest_mixed_types_1", "greatest(a, 'test', 4)"},
		{"scalar_hour", "hour(g)"},
		{"scalar_if", "if(a, 2, 3)"},
		{"scalar_ifnull", "ifnull(a, 1)"},
		{"scalar_insert", "insert(s,a,a,s)"},
		{"scalar_instr", "instr(s,s)"},
		{"scalar_interval", "interval(a, 0, b)"},
		{"scalar_isnull", "isnull(a)"},
		{"scalar_last_day", "last_day(a)"},
		{"scalar_lcase", "lcase(s)"},
		{"scalar_least", "least(a, 2)"},
		{"scalar_least_mixed_types_0", "least(a, 'test')"},
		{"scalar_least_mixed_types_1", "least(a, 'test', 4)"},
		{"scalar_least_three_arg", "least(a, 2, b)"},
		{"scalar_left", "left(s, 2)"},
		{"scalar_left_zero", "left('abcde', 0)"},
		{"scalar_length", "length(s)"},
		{"scalar_length_int", "length(a)"},
		{"scalar_locate", "locate(s, 'funny')"},
		{"scalar_locate_three_args", "locate(s, 'funny', 3)"},
		{"scalar_log10", "log10(a)"},
		{"scalar_lower", "lower(s)"},
		{"scalar_lpad", "lpad(s, 5, 'xy')"},
		{"scalar_ltrim", "ltrim(s)"},
		{"scalar_ltrim_int", "ltrim(a)"},
		{"scalar_makedate", "makedate(a,a)"},
		{"scalar_microsecond", "microsecond(g)"},
		{"scalar_mid", "mid(s, 2, 4)"},
		{"scalar_minute", "minute(g)"},
		{"scalar_mod", "mod(a, 10)"},
		{"scalar_month", "month(g)"},
		{"scalar_monthname", "monthname(g)"},
		{"scalar_nullif", "nullif(a, 1)"},
		{"scalar_pow", "pow(a,a)"},
		{"scalar_quarter", "quarter(g)"},
		{"scalar_radians", "radians(a)"},
		{"scalar_repeat", "repeat(s, b)"},
		{"scalar_repeat_null", "repeat(s, NULL)"},
		{"scalar_repeat_null_str", "repeat(NULL, b)"},
		{"scalar_repeat_zero", "repeat(s, 0)"},
		{"scalar_replace", "replace(s,s,s)"},
		{"scalar_reverse", "reverse(s)"},
		{"scalar_right", "right(s, 2)"},
		{"scalar_round", "round(a, 5)"},
		{"scalar_round_negative", "round(a, -5)"},
		{"scalar_round_str", "round(s)"},
		{"scalar_rpad", "rpad(s, 5, 'xy')"},
		{"scalar_rtrim", "rtrim(s)"},
		{"scalar_rtrim_bool", "rtrim(t)"},
		{"scalar_second", "second(g)"},
		{"scalar_sign", "sign(a)"},
		{"scalar_sin", "sin(a)"},
		{"scalar_sleep_str", "sleep(s)"},
		{"scalar_space_str", "space(a)"},
		{"scalar_sqrt", "sqrt(a)"},
		{"scalar_subdate", "subdate(g, INTERVAL 10 day)"},
		{"scalar_subdate_null", "subdate(g, null)"},
		{"scalar_substr", "substr(s, 2)"},
		{"scalar_substr_from_for", "substr(s, 2, 4)"},
		{"scalar_substring", "substring(s, 2)"},
		{"scalar_substring_from_for", "substring(s from 2 for 4)"},
		{"scalar_substring_from_for_mixed_types", "substring(s FROM 1 FOR s)"},
		{"scalar_substring_index", "substring_index(s,s,a)"},
		{"scalar_substring_mixed_types", "substring(s, s)"},
		{"scalar_substring_negative", "substring(s, -2)"},
		{"scalar_tan", "tan(a)"},
		{"scalar_timestamp", "timestamp(s)"},
		{"scalar_timestampadd", "timestampadd(month, 2, s)"},
		{"scalar_timestampdiff_microsecond", "timestampdiff(microsecond, s, s)"},
		{"scalar_timestampdiff_month", "timestampdiff(month, s, s)"},
		{"scalar_to_days", "to_days(s)"},
		{"scalar_to_seconds", "to_seconds(s)"},
		{"scalar_trim", "trim(s)"},
		{"scalar_trim_date", "trim(g)"},
		{"scalar_truncate", "truncate(a, 3)"},
		{"scalar_truncate_negative", "truncate(a, -3)"},
		{"scalar_ucase", "ucase(s)"},
		{"scalar_upper", "upper(s)"},
		{"scalar_week_0", "week(a,5)"},
		{"scalar_week_1", "week(a,7)"},
		{"scalar_week_2", "week(g)"},
		{"scalar_week_3", "week(g, 0)"},
		{"scalar_weekday", "weekday(g)"},
		{"scalar_year", "year(g)"},
		{"scalar_yearweek_0", "yearweek(a, 5)"},
		{"scalar_yearweek_1", "yearweek(a, 7)"},

		// logical, arithmetic, & other binary expressions
		{"add_str", "s + s"},
		{"and", "a > 3 AND a < 10"},
		{"and_assoc_left_int", "(a > 3 AND a < 10) AND b = 10"},
		{"and_assoc_right_int", "a > 3 AND (a < 10 AND b = 10)"},
		{"and_or_assoc_left_int", "(a > 3 AND a < 10) OR b = 10"},
		{"and_or_assoc_right_int", "a > 3 AND (a < 10 OR b = 10)"},
		{"div", "a div b"},
		{"div_str", "s div s"},
		{"divide", "a / b"},
		{"divide_str", "s / s"},
		{"equal_int", "a = 3"},
		{"equal", "t = 0"},
		{"gt_int", "a > 3"},
		{"gte_int", "a >= 3"},
		{"in", "a in (2,3,5)"},
		{"in_date", "h IN('2016-02-03 12:23:11.392')"},
		{"in_int", "a IN(1,3,5)"},
		{"in_timestamp", "g IN('2016-02-03 12:23:11.392')"},
		{"is_false_bool", "t IS FALSE"},
		{"is_false_date", "g is false"},
		{"is_false_int", "a IS FALSE"},
		{"is_not_false_bool", "t IS NOT FALSE"},
		{"is_not_null", "a IS NOT NULL"},
		{"is_not_true_bool", "t IS NOT TRUE"},
		{"is_not_unknown", "a IS NOT UNKNOWN"},
		{"is_null", "a IS NULL"},
		{"is_true_bool", "t IS TRUE"},
		{"is_true_int", "a IS TRUE"},
		{"is_true_str", "s is true"},
		{"is_unknown", "a IS UNKNOWN"},
		{"like_int", "a LIKE '%un%'"},
		{"lt_int", "a < 3"},
		{"lte_int", "a <= 3"},
		{"mod_str", `s % s`},
		{"multiply_str", "s * s"},
		{"ne_int", "a <> 3"},
		{"not_and", "NOT (a > 3 AND a < 10)"},
		{"not_eq", "NOT (a = 3)"},
		{"not_gt", "NOT (a > 3)"},
		{"not_in_int", "a NOT IN (1,3,5)"},
		{"not_int_in", "NOT a IN (1,3,5)"},
		{"not_like_int", "a NOT LIKE '%un%'"},
		{"not_ne", "NOT (a <> 3)"},
		{"not_not_and", "NOT (NOT (a > 3 AND a < 10))"},
		{"not_not_gt", "NOT (NOT (a > 3))"},
		{"not_not_in", "NOT (a NOT IN (1,3,5))"},
		{"not_or", "NOT (a > 3 OR a < 10)"},
		{"not_regexp_0_int", "a NOT REGEXP 'a' OR 'b'"},
		{"not_regexp_1_int", "a NOT REGEXP 'abc'"},
		{"not_regexp_2_int", "a NOT REGEXP '(.* )?'"},
		{"nse", "a <=> 5"},
		{"or_assoc_left_int", "(a > 3 OR a < 10) OR b = 10"},
		{"or_assoc_right_int", "a > 3 OR (a < 10 OR b = 10)"},
		{"or_int", "a > 3 OR a < 10"},
		{"regexp_0_int", "a REGEXP 'abc'"},
		{"regexp_1_int", "a REGEXP '(.* )?'"},
		{"regexp_2_int", "a REGEXP 'a' OR 'b'"},
		{"subtract_str", "s - s"},

		// aggregate functions
		{"aggregate_date_avg", "avg(h)"},
		{"aggregate_date_std", "std(h)"},
		{"aggregate_date_stddev", "stddev(h)"},
		{"aggregate_date_stddev_pop", "stddev_pop(h)"},
		{"aggregate_date_stddev_samp", "stddev_samp(h)"},
		{"aggregate_date_sum", "sum(h)"},
		{"aggregate_min", "min(a + 4)"},
		{"aggregate_std", "std(a)"},
		{"aggregate_stddev", "stddev(a)"},
		{"aggregate_stddev_samp", "stddev_samp(a)"},
		{"aggregate_sum_0", "sum(a * b)"},
		{"aggregate_sum_1", "sum(a < 1)"},
		{"aggregate_sum_2", "sum(a)"},

		// other
		{"case", "case when a > 1 then 'gt' else 'lt' end"},
		{"column_bool", "t"},
	}

	// open the file with the cached test results
	cacheFile := "testdata/test_translate.json"
	file, err := os.Open(cacheFile)
	req.Nil(err)

	// read the contents of the cache file and close it
	data, err := ioutil.ReadAll(file)
	req.Nil(err, "failed to read cached results file")
	err = file.Close()
	req.Nil(err, "failed to close cached results file")

	// unmarshal the cached results into a three-dimensional
	// map, which is structured as follows:
	// {
	//   <expr|predicate>: {
	//     <mongodb_version>: {
	//       <expr_testcase_name>: <agg_expr_as_json_string>,
	//	   }
	//   }
	// }
	cache := make(map[string]map[string]map[string]string)
	err = json.Unmarshal(data, &cache)
	req.Nil(err, "failed to unmarshal cached results json")

	// define the MongoDB versions for which we want to test translation
	versions := [][]uint8{
		{3, 2, 0},
		{3, 4, 0},
		{3, 6, 0},
	}

	type translateFunc func(*testing.T, []uint8, string) string
	translators := map[string]translateFunc{
		"expr":      translateExpr,
		"predicate": translatePredicate,
	}

	// run a subtest for each type of translation
	for typ, translator := range translators {
		t.Run(typ, func(t *testing.T) {
			if cache[typ] == nil {
				cache[typ] = make(map[string]map[string]string)
			}

			// run a subtest for each version
			for _, version := range versions {
				v := formatVersion(version)
				t.Run(v, func(t *testing.T) {
					if cache[typ][v] == nil {
						cache[typ][v] = make(map[string]string)
					}

					// run a subtest for each expression
					for _, test := range tests {
						t.Run(test.name, func(t *testing.T) {
							req := require.New(t)
							actual := translator(t, version, test.sql)
							if *update {
								cache[typ][v][test.name] = actual
								return
							}
							expected, ok := cache[typ][v][test.name]
							req.True(ok, "test case not found in cache")
							req.Equal(expected, actual, "result does not match cached result")
						})
					}

				})
			}

		})
	}

	if *update {
		cacheBytes, err := json.MarshalIndent(cache, "", "    ")
		req.Nil(err)

		err = ioutil.WriteFile(cacheFile, cacheBytes, os.ModePerm)
		req.Nil(err)
	}
}

func formatVersion(version []uint8) string {
	versionString := fmt.Sprintf("%d.%d", version[0], version[1])
	if version[2] != 0 {
		versionString = fmt.Sprintf("%s.%d", versionString, version[2])
	}
	return versionString
}

func translateExpr(t *testing.T, version []uint8, sql string) string {
	req := require.New(t)

	testSchema := evaluator.MustLoadSchema(translatorTestSchema)

	db := testSchema.Database("translate_test_db")
	req.NotNil(db, "failed to get db from schema")

	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	ctx := createTestEvalCtx(testInfo, version...)

	translator := evaluator.NewPushDownTranslator(
		createFieldNameLookup(db),
		ctx,
	)

	e, err := evaluator.GetSQLExpr(testSchema, "translate_test_db", tableTwoName, sql)
	req.Nil(err, "failed to get sql expr")

	n, err := evaluator.OptimizeEvaluations(e, ctx, ctx.Logger(""))
	req.Nil(err, "failed to optimize evaluations")

	e, ok := n.(evaluator.SQLExpr)
	req.True(ok, "node was not a SQLExpr")

	translated, ok := translator.TranslateExpr(e)

	if ok {
		jsonResult, err := json.Marshal(translated)
		req.Nil(err, "failed to marshal pipeline to json")
		return string(jsonResult)
	}

	return ""
}

func translatePredicate(t *testing.T, version []uint8, sql string) string {
	req := require.New(t)

	testSchema := evaluator.MustLoadSchema(translatorTestSchema)

	db := testSchema.Database("translate_test_db")
	req.NotNil(db, "could not find database in schema")

	testInfo := evaluator.GetMongoDBInfo(nil, testSchema, mongodb.AllPrivileges)
	ctx := createTestEvalCtx(testInfo, version...)

	translator := evaluator.NewPushDownTranslator(
		createFieldNameLookup(db),
		ctx,
	)

	e, err := evaluator.GetSQLExpr(testSchema, "translate_test_db", tableTwoName, sql)
	req.Nil(err, "failed to get sql expr")

	n, err := evaluator.OptimizeEvaluations(e, ctx, ctx.Logger(""))
	req.Nil(err, "failed to optimize evaluations")

	e, ok := n.(evaluator.SQLExpr)
	req.True(ok, "node was not a SQLExpr")

	match, local := translator.TranslatePredicate(e)

	if local == nil {
		jsonResult, err := json.Marshal(match)
		req.Nil(err, "failed to marshal pipeline to json")
		return string(jsonResult)
	}

	return ""
}

func TestTranslateCurrentDateExpr(t *testing.T) {
	type test struct {
		name     string
		sql      string
		expected string
	}

	// pad 0 pads month, day, hour, minute, second to two digits.
	pad := func(val string) string {
		if len(val) == 1 {
			return "0" + val
		}
		return val
	}

	// currentDateLiteral gets the current date as a string or int in a literal.
	// This isn't perfect, but there is no better way to test curdate at this time.
	currentDateLiteral := func(asString bool, location *time.Location) string {
		now := time.Now().In(location)
		yearStr, monthStr, dayStr := pad(strconv.Itoa(now.Year())), pad(strconv.Itoa(int(now.Month()))), pad(strconv.Itoa(now.Day()))
		if asString {
			return fmt.Sprintf(`{"$literal":"%s-%s-%s"}`, yearStr, monthStr, dayStr)
		}
		return fmt.Sprintf(`{"$literal":%s%s%s}`, yearStr, monthStr, dayStr)
	}

	tests := []test{
		{"concat_curdate", "concat(curdate(), '')", currentDateLiteral(true, schema.DefaultLocale)},
		{"curdate", "curdate() + 0", currentDateLiteral(false, schema.DefaultLocale)},
		{"concat_utc_date", "concat(utc_date(), '')", currentDateLiteral(true, time.UTC)},
		{"utc_date", "utc_date() + 0", currentDateLiteral(false, time.UTC)},
	}

	version := []uint8{3, 6, 0}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)
			actual := translateExpr(t, version, test.sql)
			req.Equal(test.expected, actual, "actual agg expr should equal expected agg expr")
		})
	}
}

func TestTranslatePartialPredicate(t *testing.T) {
	req := require.New(t)

	type test struct {
		name      string
		sql       string
		expected  string
		localDesc string
		local     evaluator.SQLExpr
	}

	runPartialTests := func(tests []test) {
		schema := evaluator.MustLoadSchema(translatorTestSchema)

		db := schema.Database("translate_test_db")
		req.NotNil(db, "could not find db in schema")

		testInfo := evaluator.GetMongoDBInfo(nil, schema, mongodb.AllPrivileges)
		ctx := createTestEvalCtx(testInfo, 3, 4, 0)

		translator := evaluator.NewPushDownTranslator(
			createFieldNameLookup(db),
			ctx,
		)

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req := require.New(t)

				e, err := evaluator.GetSQLExpr(schema, "translate_test_db", tableTwoName, test.sql)
				req.Nil(err, "could not get sql expr")

				match, local := translator.TranslatePredicate(e)
				jsonResult, err := json.Marshal(match)
				req.Nil(err, "could not marshal json result")
				req.Equal(test.expected, string(jsonResult), "actual match expr did not match expected")
				req.Zero(ShouldResemble(test.local, local), "untranslated exprs did not match")
			})
		}
	}

	tests := []test{
		// non-boolean types always exclude null
		{"0", "a", `{"a":{"$ne":null}}`, `a`, evaluator.NewSQLColumnExpr(1, "translate_test_db", tableTwoName, "a", schema.SQLInt, schema.MongoInt)},
		{"1", "a = 3 AND a < b", `{"a":3}`, "a < b", evaluator.NewSQLLessThanExpr(evaluator.NewSQLColumnExpr(1, "translate_test_db", tableTwoName, "a", schema.SQLInt, schema.MongoInt),
			evaluator.NewSQLColumnExpr(1, "translate_test_db", tableTwoName, "b", schema.SQLInt, schema.MongoInt))},
		{"2", "a = 3 AND a < b AND b = 4", `{"$and":[{"a":3},{"b":4}]}`, "a < b", evaluator.NewSQLLessThanExpr(evaluator.NewSQLColumnExpr(1, "translate_test_db", tableTwoName, "a", schema.SQLInt, schema.MongoInt),
			evaluator.NewSQLColumnExpr(1, "translate_test_db", tableTwoName, "b", schema.SQLInt, schema.MongoInt))},
		{"3", "a < b AND a = 3", `{"a":3}`, "a < b", evaluator.NewSQLLessThanExpr(evaluator.NewSQLColumnExpr(1, "translate_test_db", tableTwoName, "a", schema.SQLInt, schema.MongoInt),
			evaluator.NewSQLColumnExpr(1, "translate_test_db", tableTwoName, "b", schema.SQLInt, schema.MongoInt))},
		{"4", "NOT (a = 3 AND a < b)", `{"$and":[{"a":{"$ne":3}},{"a":{"$ne":null}}]}`, "NOT a < b",
			evaluator.NewSQLNotExpr(evaluator.NewSQLLessThanExpr(evaluator.NewSQLColumnExpr(1, "translate_test_db", tableTwoName, "a", schema.SQLInt, schema.MongoInt),
				evaluator.NewSQLColumnExpr(1, "translate_test_db", tableTwoName, "b", schema.SQLInt, schema.MongoInt)))},
	}

	runPartialTests(tests)
}

func TestTranslateSQLValue(t *testing.T) {
	req := require.New(t)

	schema := evaluator.MustLoadSchema(translatorTestSchema)

	db := schema.Database("translate_test_db")
	req.NotNil(db, "failed to get db from schema")

	type test struct {
		name     string
		sqlValue evaluator.SQLValue
		expected string
	}

	datetime, _ := time.Parse("2006 Jan 02 15:04:05", "2012 Dec 07 12:15:30.918273645")
	tests := []test{
		{"SQLTrue", evaluator.SQLTrue, `{"$literal":true}`},
		{"SQLFalse", evaluator.SQLFalse, `{"$literal":false}`},
		{"SQLFloat", evaluator.SQLFloat(1.1), `{"$literal":1.1}`},
		{"SQLInt", evaluator.SQLInt(11), `{"$literal":11}`},
		{"SQLUint", evaluator.SQLUint32(32), `{"$literal":32}`},
		{"SQLVarchar", evaluator.SQLVarchar("vc"), `{"$literal":"vc"}`},
		{"SQLNull", evaluator.SQLNull, `{"$literal":null}`},
		{"SQLDate", evaluator.SQLDate{Time: datetime}, `{"$literal":"2012-12-07T12:15:30.918273645Z"}`},
		{"SQLTimestamp", evaluator.SQLTimestamp{Time: datetime}, `{"$literal":"2012-12-07T12:15:30.918273645Z"}`},
	}

	testInfo := evaluator.GetMongoDBInfo(nil, schema, mongodb.AllPrivileges)
	// Should always translate on any server version.
	ctx := createTestEvalCtx(testInfo, 0, 0, 0)

	translator := evaluator.NewPushDownTranslator(
		createFieldNameLookup(db),
		ctx,
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			match, ok := translator.TranslateExpr(test.sqlValue)
			req.True(ok)

			jsonResult, err := json.Marshal(match)
			req.Nil(err)

			req.Equal(test.expected, string(jsonResult), "they should be equal")
		})
	}
}

var translatorTestSchema = []byte(`
schema:
- db: translate_test_db
  tables:
  -
     table: bar
     collection: bar
     columns:
     -
        Name: a
        MongoType: int
        SqlType: int
     -
        Name: b
        MongoType: int
        SqlType: int
     -
        Name: loc.1
        MongoType: string
        SqlName: c
        SqlType: varchar
     -
        Name: g
        MongoType: date
        SqlName: g
        SqlType: timestamp
     -
        Name: h
        MongoType: date
        SqlName: h
        SqlType: date
     -
        Name: s
        MongoType: string
        SqlName: s
        SqlType: varchar
     -
        Name: t
        MongoType: bool
        SqlName: t
        SqlType: boolean
     -
        Name: _id
        MongoType: bson.ObjectId
        SqlType: varchar
`)
