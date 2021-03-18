package evaluator_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/10gen/mongoast/astprint"
	"github.com/10gen/mongoast/parser"
	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/schema"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/require"
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
		{"scalar_abs_int_add", "abs(a+a)"},
		{"scalar_abs_neg_const", "abs(-3.2)"},
		{"scalar_abs_null", "abs(null)"},
		{"scalar_abs_string", "abs(s)"},
		{"scalar_acos", "acos(a)"},
		{"scalar_acos_arith_expr", "acos(2 * (-3 * 3))"},
		{"scalar_acos_null", "acos(null)"},
		{"scalar_acos_string", "acos(s)"},
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
		{"scalar_adddate_13", "adddate(date_format(g, '%y-%m-%d %H:%i:%s'), interval a second)"},
		{"scalar_ascii", "ascii(s)"},
		{"scalar_ascii_int", "ascii(a)"},
		{"scalar_ascii_bool", "ascii(t)"},
		{"scalar_ascii_date", "ascii(g)"},
		{"scalar_asin", "asin(a)"},
		{"scalar_asin_arith_expr", "asin('0.7'/'2')"},
		{"scalar_asin_null", "asin(null)"},
		{"scalar_asin_string", "asin(s)"},
		{"scalar_atan", "atan(a)"},
		{"scalar_atan_int_add", "atan(a+a)"},
		{"scalar_atan_null", "atan(null)"},
		{"scalar_atan_string", "atan(s)"},
		{"scalar_atan2", "atan2(a)"},
		{"scalar_atan2_arith_expr", "atan2(3*a)"},
		{"scalar_atan2_null", "atan2(null)"},
		{"scalar_atan2_string", "atan2(s)"},
		{"scalar_cast_date_to_time", "cast(g as time)"},
		{"scalar_cast_int_add_to_signed", "cast(a+a as signed)"},
		{"scalar_cast_int_add_to_unsigned", "cast(a+a as unsigned)"},
		{"scalar_cast_int_to_char", "cast(a as char)"},
		{"scalar_cast_int_to_signed", "cast(a as signed)"},
		{"scalar_cast_null_to_signed", "cast(null as signed)"},
		{"scalar_cast_string_to_char", "cast(s as char)"},
		{"scalar_cast_string_to_binary", "cast(s as binary)"},
		{"scalar_ceil", "ceil(a)"},
		{"scalar_ceil_add", "ceil(a+a)"},
		{"scalar_ceil_null", "ceil(null)"},
		{"scalar_ceil_string", "ceil(s)"},
		{"scalar_ceiling", "ceiling(a)"},
		{"scalar_ceiling_add", "ceiling(a+a)"},
		{"scalar_ceiling_null", "ceiling(null)"},
		{"scalar_ceiling_string", "ceiling(s)"},
		{"scalar_char", "char(a)"},
		{"scalar_char_add", "char(a+a)"},
		{"scalar_char_null", "char(null)"},
		{"scalar_char_string", "char(s)"},
		{"scalar_char_length", "char_length(s)"},
		{"scalar_character_length", "character_length(s)"},
		{"scalar_character_length_empty", "character_length('')"},
		{"scalar_character_length_null", "character_length(null)"},
		{"scalar_coalesce", "coalesce(null, null, 2, null, 3, null)"},
		{"scalar_coalesce_non_null", "coalesce(2, a, 3)"},
		{"scalar_coalesce_only_null", "coalesce(null)"},
		{"scalar_coalesce_mixed_types", "coalesce(null,'2', null, 3, 'foo', 3.14)"},
		{"scalar_concat", "concat(s, 'funny')"},
		{"scalar_concat_empty", "concat(s, '')"},
		{"scalar_concat_mixed_types", "concat(s, a)"},
		{"scalar_concat_null", "concat(s, null)"},
		{"scalar_concat_ws", "concat_ws(',', s)"},
		{"scalar_concat_ws_null", "concat_ws(',', s, null)"},
		{"scalar_conv_constant_bases", "conv(s, 2, 8)"},
		{"scalar_conv_column_bases", "conv(s, a, b)"},
		{"scalar_conv_null", "conv(s, null, 4)"},
		{"scalar_convert_int_to_float", "convert(a, float)"},
		{"scalar_convert_float_to_int", "convert(c, signed)"},
		{"scalar_convert_int_to_decimal", "convert(a, decimal)"},
		{"scalar_convert_decimal_to_int", "convert(d, signed)"},
		{"scalar_convert_date_to_time", "convert(g, time)"},
		{"scalar_convert_int_add_to_signed", "convert(a+a, signed)"},
		{"scalar_convert_int_to_char", "convert(a, char)"},
		{"scalar_convert_int_to_signed", "convert(a, signed)"},
		{"scalar_convert_null_to_signed", "convert(null, signed)"},
		{"scalar_convert_string_to_char", "convert(s, char)"},
		{"scalar_convert_string_to_binary", "convert(s, binary)"},
		{"scalar_cos", "cos(a)"},
		{"scalar_cos_nested_call", "cos(pi())"},
		{"scalar_cos_null", "cos(null)"},
		{"scalar_cot", "cot(a)"},
		{"scalar_cot_int_add", "cot(a+a)"},
		{"scalar_cot_null", "cot(null)"},
		{"scalar_count_0", "count(*)"},
		{"scalar_count_1", "count(a + b)"},
		{"scalar_date", "date(s)"},
		{"scalar_date_null", "date(null)"},
		{"scalar_date_redundant", "date(g)"},
		{"scalar_date_add", "date_add(g, INTERVAL 10 day)"},
		{"scalar_date_add_sec", "date_add(g, INTERVAL 10 second)"},
		{"scalar_date_add_min", "date_add(g, INTERVAL 10 minute)"},
		{"scalar_date_add_neg_min", "date_add(g, INTERVAL -10 minute)"},
		{"scalar_date_format", "date_format(g, '%% %d %f %H %k %i %j %m %s %S %T %U %Y')"},
		{"scalar_date_format_0", "date_format(g, '%Y%')"},
		{"scalar_date_format_1", "date_format(g, '%V')"},
		{"scalar_date_format_null", "date_format(g, null)"},
		{"scalar_date_sub", "date_sub(g, INTERVAL 10 day)"},
		{"scalar_date_sub_sec", "date_sub(g, INTERVAL 10 second)"},
		{"scalar_date_sub_min", "date_sub(g, INTERVAL 10 minute)"},
		{"scalar_date_sub_neg_min", "date_sub(g, INTERVAL -10 minute)"},
		{"scalar_degrees", "degrees(a)"},
		{"scalar_degrees_const", "degrees(3.14)"},
		{"scalar_degrees_null", "degrees(null)"},
		{"scalar_degrees_null_plus", "degrees(null + 3.14)"},
		{"scalar_datediff_0", "datediff(g, '2000-09-01 13:29:18')"},
		{"scalar_datediff_1", "datediff(g, h)"},
		{"scalar_datediff_mixed_types_bool", "datediff(t, '2000-09-01 13:29:18')"},
		{"scalar_datediff_mixed_types_int", "datediff(a, '2000-09-01 13:29:18')"},
		{"scalar_datediff_mixed_types_str", "datediff(s, '2000-09-01 13:29:18')"},
		{"scalar_datediff_polymorphic", "datediff(p, '2000-09-01 13:29:18')"},
		{"scalar_day", "day(g)"},
		{"scalar_day_int", "day(a)"},
		{"scalar_day_string", "day(s)"},
		{"scalar_day_null", "day(null)"},
		{"scalar_dayname", "dayname(g)"},
		{"scalar_dayname_int", "dayname(a)"},
		{"scalar_dayofmonth", "dayofmonth(g)"},
		{"scalar_dayofmonth_null", "dayofmonth(null)"},
		{"scalar_dayofweek", "dayofweek(g)"},
		{"scalar_dayofyear", "dayofyear(g)"},
		{"scalar_dayofyear_string", "dayofyear(s)"},
		{"scalar_exp", "exp(a)"},
		{"scalar_exp_zero", "exp(0)"},
		{"scalar_extract", "extract(year from g)"},
		{"scalar_extract_second", "extract(second from g)"},
		{"scalar_elt", "elt(1, a, a)"},
		{"scalar_elt_too_high", "elt(8, a, a)"},
		{"scalar_elt_too_low", "elt(0, a, a)"},
		{"scalar_field_ints", "field(a, 3, 2, 1)"},
		{"scalar_field_strings", `field(s, "3", "2", "1")`},
		{"scalar_field_mixed", `field(a, "3", 2, 1)`},
		{"scalar_floor", "floor(a)"},
		{"scalar_floor_nested", "floor(exp(2))"},
		{"scalar_floor_null", "floor(null)"},
		{"scalar_floor_null_mult", "floor(null * 1.6)"},
		{"scalar_from_unixtime", "from_unixtime(a)"},
		{"scalar_from_unixtime_w_format", "from_unixtime(a, '%Y%m%d%H%i%s')"},
		{"scalar_from_days", "from_days(a)"},
		{"scalar_from_days_null", "from_days(null)"},
		{"scalar_from_days_string", "from_days(s)"},
		{"scalar_greatest", "greatest(a, 2)"},
		{"scalar_greatest_mixed_types_0", "greatest(a, 'test')"},
		{"scalar_greatest_mixed_types_1", "greatest(a, 'test', 4)"},
		{"scalar_hour", "hour(g)"},
		{"scalar_if", "if(a, 2, 3)"},
		{"scalar_ifnull", "ifnull(a, 1)"},
		{"scalar_insert", "insert(s,a,a,s)"},
		{"scalar_insert_null", "insert(s,a,a,null)"},
		{"scalar_insert_null_idx", "insert(s,null,a,s)"},
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
		{"scalar_ln", "ln(a)"},
		{"scalar_ln_add", "ln(a+a)"},
		{"scalar_ln_null", "ln(null)"},
		{"scalar_ln_string", "ln(s)"},
		{"scalar_ln_mixed_null_plus", "ln(a + null)"},
		{"scalar_locate", "locate(s, 'funny')"},
		{"scalar_locate_three_args", "locate(s, 'funny', 3)"},
		{"scalar_log", "log(a)"},
		{"scalar_log_add", "log(a+a)"},
		{"scalar_log_null", "log(null)"},
		{"scalar_log_string", "log(s)"},
		{"scalar_log10", "log10(a)"},
		{"scalar_log2", "log2(a)"},
		{"scalar_log2_add", "log2(a+a)"},
		{"scalar_log2_null", "log2(null)"},
		{"scalar_log2_string", "log2(s)"},
		{"scalar_log2_twice", "log2(log2(16))"},
		{"scalar_lower", "lower(s)"},
		{"scalar_lpad", "lpad(s, 5, 'xy')"},
		{"scalar_ltrim", "ltrim(s)"},
		{"scalar_ltrim_blank", "ltrim(' ')"},
		{"scalar_ltrim_int", "ltrim(a)"},
		{"scalar_makedate", "makedate(a,a)"},
		{"scalar_md5", "md5(s)"},
		{"scalar_md5_int", "md5(a)"},
		{"scalar_md5_int_add", "md5(a+a)"},
		{"scalar_md5_null", "md5(null)"},
		{"scalar_microsecond", "microsecond(g)"},
		{"scalar_mid", "mid(s, 2, 4)"},
		{"scalar_minute", "minute(g)"},
		{"scalar_mod", "mod(a, 10)"},
		{"scalar_month", "month(g)"},
		{"scalar_monthname", "monthname(g)"},
		{"scalar_nopushdown", "nopushdown(a+a)"},
		{"scalar_not", "not(t)"},
		{"scalar_not_int", "not(a)"},
		{"scalar_not_null", "not(null)"},
		{"scalar_nullif", "nullif(a, 1)"},
		{"scalar_pow", "pow(a,a)"},
		{"scalar_power", "power(a,a)"},
		{"scalar_power_all_null", "power(null,null)"},
		{"scalar_power_of_null", "power(null,a)"},
		{"scalar_power_to_bool", "power(a,t)"},
		{"scalar_power_to_null", "power(a,null)"},
		{"scalar_quarter", "quarter(g)"},
		{"scalar_radians", "radians(a)"},
		{"scalar_rand", "rand()"},
		{"scalar_rand_seed", "rand(a)"},
		{"scalar_rand_seed_add_const", "rand(2+2)"},
		{"scalar_rand_string", "rand(s)"},
		{"scalar_rand_null", "rand(null)"},
		{"scalar_repeat", "repeat(s, b)"},
		{"scalar_repeat_add", "repeat('h' + 'i', 1+1)"},
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
		{"scalar_sqrt_neg", "sqrt(-2)"},
		{"scalar_str_to_date_00", "str_to_date(s, '%Y-%m-%d %H:%i:%s')"},
		{"scalar_str_to_date_01", "str_to_date(s, '%Y-%m-%d %H:%i:%s')"},
		{"scalar_str_to_date_02", "str_to_date(s, '%Y-%m-%d %H:%i:%s')"},
		{"scalar_str_to_date_03", "str_to_date(s, '%Y-%m-%d %H:%i:00')"},
		{"scalar_str_to_date_04", "str_to_date(s, '%Y-%m-%d %H:00:00')"},
		{"scalar_str_to_date_05", "str_to_date(s, '%Y-%m-%d 00:00:00')"},
		{"scalar_str_to_date_06", "str_to_date(s, '%Y-%m-%d')"},
		{"scalar_str_to_date_07", "str_to_date(s, '%Y-%m-01 00:00:00')"},
		{"scalar_str_to_date_08", "str_to_date(s, '%Y-%m-01')"},
		{"scalar_str_to_date_09", "str_to_date(s, '%Y-01-01 00:00:00')"},
		{"scalar_str_to_date_10", "str_to_date(s, '%Y-01-01')"},
		{"scalar_str_to_date_11", "str_to_date(s, '%M %d %Y')"},
		{"scalar_str_to_date_12", "str_to_date(s, null)"},
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
		{"scalar_time_to_sec", "time_to_sec(g)"},
		{"scalar_time_to_sec_int", "time_to_sec(a)"},
		{"scalar_time_to_sec_null", "time_to_sec(null)"},
		{"scalar_time_to_sec_string", "time_to_sec(s)"},
		{"scalar_timediff", "timediff(g, g)"},
		{"scalar_timediff_ints", "timediff(a, a)"},
		{"scalar_timediff_null", "timediff(g, null)"},
		{"scalar_timediff_string", "timediff(s, s)"},
		{"scalar_timediff_string_mixed", "timediff(s, g)"},
		{"scalar_timestamp", "timestamp(s)"},
		{"scalar_timestampadd", "timestampadd(month, 2, s)"},
		{"scalar_timestampdiff_microsecond", "timestampdiff(microsecond, s, s)"},
		{"scalar_timestampdiff_month", "timestampdiff(month, s, s)"},
		{"scalar_to_days", "to_days(s)"},
		{"scalar_to_seconds", "to_seconds(s)"},
		{"scalar_trim", "trim(s)"},
		{"scalar_trim_date", "trim(g)"},
		{"scalar_truncate_positive_pos", "truncate(a, 3)"},
		{"scalar_truncate_negative_pos", "truncate(a, -3)"},
		{"scalar_truncate_column_pos", "truncate(a,a)"},
		{"scalar_ucase", "ucase(s)"},
		// This test relies on the current time, so it cannot be tested without
		// at-runtime checks. Hardcoded result values don't work because the
		// correct result will be different depending on when the test is ran.
		//{"scalar_unix_timestamp", "unix_timestamp()"},
		// These tests rely on the current timezone, so they cannot be tested
		// without at-runtime checks. Hardcoded result values don't work because
		// the correct result will be different depending on where the test is ran.
		//{"scalar_unix_timestamp_int", "unix_timestamp(a)"},
		//{"scalar_unix_timestamp_null", "unix_timestamp(null)"},
		//{"scalar_unix_timestamp_string", "unix_timestamp(s)"},
		//{"scalar_unix_timestamp_date", "unix_timestamp(h)"},
		//{"scalar_unix_timestamp_timestamp", "unix_timestamp(g)"},
		{"scalar_upper", "upper(s)"},
		{"scalar_upper_useless", "upper('UPPER')"},
		{"scalar_week_0", "week(a,5)"},
		{"scalar_week_1", "week(a,7)"},
		{"scalar_week_2", "week(g)"},
		{"scalar_week_3", "week(g, 0)"},
		{"scalar_weekday", "weekday(g)"},
		{"scalar_year", "year(g)"},
		{"scalar_yearweek_0", "yearweek(a, 5)"},
		{"scalar_yearweek_1", "yearweek(a, 7)"},
		{"scalar_date_int", "date(a)"},

		// logical, arithmetic, & other binary expressions
		{"add_str", "s + s"},
		{"add_multiple_0", "1 + s + 3"},
		{"add_multiple_1", "1.3 + s + 2.4"},
		{"add_multiple_2", "1.3 + s + 2"},
		{"add_multiple_3", "s + 1.3 + 2"},
		{"add_multiple_4", "1 + 3 + s"},
		{"add_multiple_5", "-1 + 1 + s"},
		{"add_multiple_6", "1 + 2 + 3 + 4"},
		{"add_bool_to_float", "t + 3.14"},
		{"add_nulls", "null + null"},
		{"and", "a > 3 AND a < 10"},
		{"and_false_null", "false AND null"},
		{"and_true_null", "true AND null"},
		{"and_with_constant_folding_0", "1 < 2 AND a < 10 AND 2 < 3"},
		{"and_with_constant_folding_1", "a < 10 AND 1 < 2 AND 2 < 3"},
		{"and_assoc_left_int", "(a > 3 AND a < 10) AND b = 10"},
		{"and_assoc_right_int", "a > 3 AND (a < 10 AND b = 10)"},
		{"and_or_assoc_left_int", "(a > 3 AND a < 10) OR b = 10"},
		{"and_or_assoc_right_int", "a > 3 AND (a < 10 OR b = 10)"},
		{"div", "a div b"},
		{"div_str", "s div s"},
		{"div_by_float", "a div 1.2"},
		{"divide", "a / b"},
		{"divide_str", "s / s"},
		{"divide_by_bool_sum", "a / (t + t + t)"},
		{"equal_int", "a = 3"},
		{"equal", "t = 0"},
		{"equal_column", "a = b"},
		{"equal_impossible", "3 = 2"},
		{"equal_complicated_mixed", "((3 + 2)*a)/4.2 = (a*a*a)*(3 + a) + t"},
		{"equal_objectid_string_lit", "_id = '567300aad61f7baea909b3c5'"},
		{"equal_objectid_string_col", "_id = s"},
		{"gt_int", "a > 3"},
		{"gt_bool", "a > t"},
		{"gt_const", "2 > 3"},
		{"gt_column", "a > b"},
		{"gte_int", "a >= 3"},
		{"gte_bounds", "3 >= 3"},
		{"gte_expr", "a*a + 3 >= 9"},
		{"gte_column", "a >= b"},
		{"in", "a in (2,3,5)"},
		{"in_date", "h IN('2016-02-03 12:23:11.392')"},
		{"in_expr", "a*(3+a) IN (3, 4, 7*2)"},
		{"in_int", "a IN(1,3,5)"},
		{"in_null", "null IN(1,null,5)"},
		{"in_objectid_string_lit", "_id in ('567300aad61f7baea909b3c5', '567300aad61f7baea9095555')"},
		{"in_objectid_string_col", "_id in (s, s)"},
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
		{"like_normal_match", "s LIKE '%un%'"},
		{"like_cannot_match", "_id LIKE '53c2ab5e4291b17b666d742a'"},
		{"like_binary_match", "s LIKE BINARY '%UN%'"},
		{"lt_int", "a < 3"},
		{"lt_dates", "g < g + 2"},
		{"lt_column", "a < b"},
		{"lte_int", "a <= 3"},
		{"lte_strings", "s < 'hello'"},
		{"lte_column", "a <= b"},
		{"mod_str", `s % s`},
		{"mod_invalid_zero", "a % 0"},
		{"mod_multiple", "a % 3 % 6"},
		{"multiply_str", "s * s"},
		{"multiply_null", "a * null"},
		{"multiply_distributive_const", "3 * (4 + 5)"},
		{"multiply_with_constant_folding_0", "4 * (a * 5) * (2 * 3)"},
		{"multiply_with_constant_folding_1", "(a * 5) * 4 * (2 * 3)"},
		{"ne_int", "a <> 3"},
		{"ne_float_expr", "3.14*3.14*3.14 <> 3.14+(3.10 + 0.04)*3.14"},
		{"ne_mixed_types", "'hello world' <> t"},
		{"ne_null_match", "null <> null"},
		{"not_and", "NOT (a > 3 AND a < 10)"},
		{"not_eq", "NOT (a = 3)"},
		{"not_eq_column", "a != b"},
		{"not_gt", "NOT (a > 3)"},
		{"not_in_int", "a NOT IN (1,3,5)"},
		{"not_int_in", "NOT a IN (1,3,5)"},
		{"not_like_normal_match", "s NOT LIKE '%un%'"},
		{"not_like_cannot_match", "_id NOT LIKE '53c2ab5e4291b17b666d742a'"},
		{"not_like_binary_match", "s NOT LIKE BINARY '%un%'"},
		{"not_ne", "NOT (a <> 3)"},
		{"not_not_and", "NOT (NOT (a > 3 AND a < 10))"},
		{"not_not_gt", "NOT (NOT (a > 3))"},
		{"not_not_in", "NOT (a NOT IN (1,3,5))"},
		{"not_or", "NOT (a > 3 OR a < 10)"},
		{"not_regexp_0_int", "a NOT REGEXP '(a|b)'"},
		{"not_regexp_1_int", "a NOT REGEXP 'abc'"},
		{"not_regexp_2_int", "a NOT REGEXP '(.* )?'"},
		{"not_regexp_0_string", "s NOT REGEXP '(a|b)'"},
		{"not_regexp_1_string", "s NOT REGEXP 'abc'"},
		{"not_regexp_2_string", "s NOT REGEXP '(.* )?'"},
		{"nse", "a <=> 5"},
		{"nse_null", "null <=> null"},
		{"nse_one_null", "null <=> a"},
		{"or_assoc_left_int", "(a > 3 OR a < 10) OR b = 10"},
		{"or_assoc_right_int", "a > 3 OR (a < 10 OR b = 10)"},
		{"or_int", "a > 3 OR a < 10"},
		{"or_with_constant_folding_0", "5 < 10 OR a > 3 OR 3 < 8"},
		{"or_with_constant_folding_1", "a > 3 OR 5 > 10 OR 3 > 8"},
		{"regexp_0_int", "a REGEXP 'abc'"},
		{"regexp_1_int", "a REGEXP '(.* )?'"},
		{"regexp_2_int", "a REGEXP '(a|b)'"},
		{"regexp_0_string", "s REGEXP 'abc'"},
		{"regexp_1_string", "s REGEXP '(.* )?'"},
		{"regexp_2_string", "s REGEXP '(a|b)'"},
		{"subtract_str", "s - s"},
		{"subtract_const", "a - 5"},
		{"subtract_parens_term", "a - (4 - 5)"},
		{"xor", "a xor 3"},
		{"xor_bools", "t xor true"},
		{"xor_null", "null xor t"},
		{"xor_with_constant_folding_0", "true xor (a > 5) xor false"},
		{"xor_with_constant_folding_1", "(a > 5) xor true xor false"},

		// aggregate functions
		{"aggregate_date_avg", "avg(h)"},
		{"aggregate_date_std", "std(h)"},
		{"aggregate_date_stddev", "stddev(h)"},
		{"aggregate_date_stddev_pop", "stddev_pop(h)"},
		{"aggregate_date_stddev_samp", "stddev_samp(h)"},
		{"aggregate_date_sum", "sum(h)"},
		{"aggregate_group_concat", "group_concat(s)"},
		{"aggregate_group_concat_null", "group_concat(null)"},
		{"aggregate_group_concat_with_separator", "group_concat(s separator \"hello\")"},
		{"aggregate_min", "min(a + 4)"},
		{"aggregate_min_expr", "min((a + 4) * (3/4))"},
		{"aggregate_max", "max(a)"},
		{"aggregate_max_plus_null", "max(a + null)"},
		{"aggregate_max_const", "max(8*8)"},
		{"aggregate_avg_expr", "avg(a * (true + 2))"},
		{"aggregate_avg_nested", "avg(exp(3))"},
		{"aggregate_avg_string", "avg(s)"},
		{"aggregate_std", "std(a)"},
		{"aggregate_std_null", "std(null)"},
		{"aggregate_stddev", "stddev(a)"},
		{"aggregate_stddev_plus_null", "stddev(null + a)"},
		{"aggregate_stddev_pop", "stddev_pop(a)"},
		{"aggregate_stddev_samp", "stddev_samp(a)"},
		{"aggregate_stddev_samp_mixed", "stddev_samp(a + h)"},
		{"aggregate_stddev_string_reconciled", "stddev(s)"},
		{"aggregate_stddev_samp_string_reconciled", "stddev_samp(s)"},
		{"aggregate_sum_0", "sum(a * b)"},
		{"aggregate_sum_1", "sum(a < 1)"},
		{"aggregate_sum_2", "sum(a)"},
		{"aggregate_sum_string_reconciled", "sum(s)"},
		{"aggregate_count", "count(h)"},
		{"aggregate_count_nested", "count(pi())"},
		{"aggregate_count_star", "count(*)"},

		// SQLConvertExpr tests (some of these may already be tested in another way
		// in scalar funcs, see scalar_cast and scalar_convert)
		{"convert_bool_to_signed", "cast(t as signed)"},
		{"convert_bool_to_string", "cast('false' as char)"},
		{"convert_date_string_to_date", "cast('2018-09-07' as date)"},
		{"convert_datetime_string_to_date", "cast('2018-09-07 15:39:55' as date)"},
		{"convert_float_to_signed", "cast(3.14 as signed)"},
		{"convert_int_to_string", "cast(23 as char)"},
		{"convert_null_to_binary", "cast(null as binary)"},
		{"convert_signed_to_unsigned", "cast(-4 as unsigned)"},
		{"convert_string_to_datetime", "cast(s as datetime)"},
		{"convert_string_to_int", "cast('50' as unsigned)"},
		{"convert_string_to_time", "cast(s as time)"},
		{"convert_unsigned_to_signed", "cast(-1 * 4 as signed)"},

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

	// unmarshal the cached results into a four-dimensional
	// map, which is structured as follows:
	// {
	//   <expr|predicate>: {
	//     <mongodb_version>: {
	//       <expr_testcase_name>: {
	//         <type_conversion_mode>: <agg_expr_as_json_string>,
	//       }
	//     }
	//   }
	// }
	cache := make(map[string]map[string]map[string]map[string]string)
	if !*update {
		err = json.Unmarshal(data, &cache)
		req.Nil(err, "failed to unmarshal cached results json")
	}

	// define the MongoDB versions for which we want to test translation
	versions := [][]uint8{
		{3, 2, 0},
		{3, 4, 0},
		{3, 6, 0},
		{4, 0, 0},
		{4, 2, 0},
	}

	// define the type conversion modes for which we want to test translation
	sqlValueKinds := []values.SQLValueKind{
		values.MySQLValueKind,
		values.MongoSQLValueKind,
	}

	type translateFunc func(*testing.T, []uint8, string, values.SQLValueKind) string
	translators := map[string]translateFunc{
		"expr":      translateExpr,
		"predicate": translatePredicate,
	}

	// run a subtest for each expression
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if cache[test.name] == nil {
				cache[test.name] = make(map[string]map[string]map[string]string)
			}

			// run a subtest for each type of translation
			for typ, translator := range translators {
				t.Run(typ, func(t *testing.T) {
					if cache[test.name][typ] == nil {
						cache[test.name][typ] = make(map[string]map[string]string)
					}

					// run a subtest for each version
					for _, version := range versions {
						v := formatVersion(version)
						t.Run(v, func(t *testing.T) {
							if cache[test.name][typ][v] == nil {
								cache[test.name][typ][v] = make(map[string]string)
							}
							for _, sqlValueKind := range sqlValueKinds {
								mode := formatSQLValueKind(sqlValueKind)
								t.Run(mode, func(t *testing.T) {
									req = require.New(t)
									actual := translator(t, version, test.sql, sqlValueKind)
									if *update {
										cache[test.name][typ][v][mode] = actual
										return
									}
									expected, ok := cache[test.name][typ][v][mode]
									req.True(ok, "test case not found in cache")
									req.Equal(expected, actual, "result does not match cached result: %s", mode)
								})
							}
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

func formatSQLValueKind(sqlValueKind values.SQLValueKind) string {
	if sqlValueKind == values.MySQLValueKind {
		return "mysql-mode"
	}
	return "mongosql-mode"
}

func formatVersion(version []uint8) string {
	versionString := fmt.Sprintf("%d.%d", version[0], version[1])
	if version[2] != 0 {
		versionString = fmt.Sprintf("%s.%d", versionString, version[2])
	}
	return versionString
}

func translateExpr(t *testing.T, version []uint8, sql string, sqlValueKind values.SQLValueKind) string {
	req := require.New(t)

	testSchema := evaluator.MustLoadSchema(translatorTestSchema)

	db := testSchema.Database("translate_test_db")
	req.NotNil(db, "failed to get db from schema")

	execCfg := createExecutionCfg("translate_test_db", 0, version, sqlValueKind)
	optimizerCfg := createOptimizerCfg(collation.Default, execCfg)
	pushdownCfg := createPushdownCfg(version, sqlValueKind)

	translator := evaluator.NewPushdownTranslator(
		pushdownCfg,
		createFieldRefLookup(db),
	)

	e, err := evaluator.GetSQLExpr(testSchema, "translate_test_db", tableTwoName, sql, false, nil)
	req.Nil(err, "failed to get sql expr")

	n, err := evaluator.OptimizeEvaluations(optimizerCfg, e)
	req.Nil(err, "failed to optimize evaluations")

	e, ok := n.(evaluator.SQLExpr)
	req.True(ok, "node was not a SQLExpr")

	translated, pf := translator.TranslateExpr(e)

	if pf == nil {
		return astprint.ShellString(translated)
	}

	return ""
}

func translatePredicate(t *testing.T, version []uint8, sql string, sqlValueKind values.SQLValueKind) string {
	req := require.New(t)

	testSchema := evaluator.MustLoadSchema(translatorTestSchema)

	db := testSchema.Database("translate_test_db")
	req.NotNil(db, "could not find database in schema")

	execCfg := createExecutionCfg("translate_test_db", 0, version, sqlValueKind)
	optimizerCfg := createOptimizerCfg(collation.Default, execCfg)
	pushdownCfg := createPushdownCfg(version, sqlValueKind)

	translator := evaluator.NewPushdownTranslator(
		pushdownCfg,
		createFieldRefLookup(db),
	)

	e, err := evaluator.GetSQLExpr(testSchema, "translate_test_db", tableTwoName, sql, false, nil)
	req.Nil(err, "failed to get sql expr")

	n, err := evaluator.OptimizeEvaluations(optimizerCfg, e)
	req.Nil(err, "failed to optimize evaluations")

	e, ok := n.(evaluator.SQLExpr)
	req.True(ok, "node was not a SQLExpr")

	match, aggr, localMatcher, _ := translator.TranslatePredicate(e)

	ret := ""
	if localMatcher != nil {
		return ""
	}

	if match != nil {
		ret += parser.DeparseMatchExpr(match).String()
	}
	if aggr != nil {
		if ret != "" {
			ret += ","
		}
		ret += parser.DeparseMatchExpr(aggr).String()
	}

	return ret
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
		testSchema := evaluator.MustLoadSchema(translatorTestSchema)

		db := testSchema.Database("translate_test_db")
		req.NotNil(db, "could not find db in schema")

		version := []uint8{3, 4, 0}
		sqlValueKind := values.MySQLValueKind
		pushdownCfg := createPushdownCfg(version, sqlValueKind)

		translator := evaluator.NewPushdownTranslator(
			pushdownCfg,
			createFieldRefLookup(db),
		)

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				req := require.New(t)

				e, err := evaluator.GetSQLExpr(
					testSchema,
					"translate_test_db",
					tableTwoName,
					test.sql,
					false,
					nil,
				)
				req.Nilf(err, "could not get sql expr for %v", test.localDesc)

				match, _, local, _ := translator.TranslatePredicate(e)
				jsonResult := parser.DeparseMatchExpr(match).String()
				req.Equalf(test.expected, jsonResult, "actual match expr did "+
					"not match expected in %v", test.localDesc)
				req.Zerof(convey.ShouldResemble(test.local, local), "untranslated exprs "+
					"did not match in %v", test.localDesc)
			})
		}
	}

	db := "translate_test_db"

	tests := []test{
		// non-boolean types always exclude null
		{
			"0", "a", `{"$and": [{"a": {"$ne": {"$numberInt":"0"}}},{"a": {"$ne": null}}]}`, `a`,
			nil,
		},
		{
			"1", "a = 3 AND a < b", `{"a": {"$eq": {"$numberLong":"3"}}}`, "a < b",
			evaluator.NewSQLComparisonExpr(
				evaluator.LT,
				testSQLColumnExpr(1, db, tableTwoName, "a", types.EvalInt64, schema.MongoInt, false),
				testSQLColumnExpr(1, db, tableTwoName, "b", types.EvalInt64, schema.MongoInt, false),
			),
		},
		{
			"2", "a = 3 AND a < b AND b = 4", `{"$and": [{"a": {"$eq": {"$numberLong":"3"}}},{"b": {"$eq": {"$numberLong":"4"}}}]}`, "a < b",
			evaluator.NewSQLComparisonExpr(
				evaluator.LT,
				testSQLColumnExpr(1, db, tableTwoName, "a", types.EvalInt64, schema.MongoInt, false),
				testSQLColumnExpr(1, db, tableTwoName, "b", types.EvalInt64, schema.MongoInt, false),
			),
		},
		{
			"3", "a < b AND a = 3", `{"a": {"$eq": {"$numberLong":"3"}}}`, "a < b",
			evaluator.NewSQLComparisonExpr(
				evaluator.LT,
				testSQLColumnExpr(1, db, tableTwoName, "a", types.EvalInt64, schema.MongoInt, false),
				testSQLColumnExpr(1, db, tableTwoName, "b", types.EvalInt64, schema.MongoInt, false),
			),
		},
		{
			"4", "NOT (a = 3 AND a < b)", `{"$and": [{"a": {"$ne": {"$numberLong":"3"}}},{"a": {"$ne": null}}]}`, "NOT a < b",
			evaluator.NewSQLNotExpr(
				evaluator.NewSQLAndExpr(
					evaluator.NewSQLComparisonExpr(
						evaluator.EQ,
						testSQLColumnExpr(1, db, tableTwoName, "a", types.EvalInt64, schema.MongoInt, false),
						evaluator.NewSQLValueExpr(values.NewSQLInt64(values.MySQLValueKind, 3)),
					),
					evaluator.NewSQLComparisonExpr(
						evaluator.LT,
						testSQLColumnExpr(1, db, tableTwoName, "a", types.EvalInt64, schema.MongoInt, false),
						testSQLColumnExpr(1, db, tableTwoName, "b", types.EvalInt64, schema.MongoInt, false),
					),
				),
			),
		},
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
		sqlValue values.SQLValue
		expected string
	}

	datetime, _ := time.Parse("2006 Jan 02 15:04:05", "2012 Dec 07 12:15:30.918273645")
	tests := []test{
		{"SQLTrue", values.NewSQLBool(values.MySQLValueKind, true), `true`},
		{"SQLFalse", values.NewSQLBool(values.MySQLValueKind, false), `false`},
		{"SQLFloat", values.NewSQLFloat(values.MySQLValueKind, 1.1), `1.1`},
		{"SQLInt", values.NewSQLInt64(values.MySQLValueKind, 11), `NumberLong("11")`},
		{"SQLUint", values.NewSQLUint64(values.MySQLValueKind, 32), `NumberLong("32")`},
		{"SQLVarchar", values.NewSQLVarchar(values.MySQLValueKind, "vc"),
			`"vc"`},
		{"SQLNull", values.NewSQLNull(values.MySQLValueKind),
			`null`},
		{"SQLDate", values.NewSQLDate(values.MySQLValueKind, datetime),
			`ISODate("2012-12-07T00:00:00")`},
		{"SQLTimestamp", values.NewSQLTimestamp(values.MySQLValueKind, datetime),
			`ISODate("2012-12-07T12:15:30.918")`},
	}

	// Should always translate on any server version.
	pushdownCfg := createTestPushdownCfg()

	translator := evaluator.NewPushdownTranslator(
		pushdownCfg,
		createFieldRefLookup(db),
	)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			translation, pf := translator.TranslateExpr(evaluator.NewSQLValueExpr(test.sqlValue))
			req.Nil(pf)

			jsonResult := astprint.ShellString(translation)

			req.Equal(test.expected, jsonResult, "they should be equal")
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
        Name: c
        MongoType: float64
        SqlType: float64
     -
        Name: d
        MongoType: bson.Decimal128
        SqlType: decimal128
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
        Name: p
        MongoType: int
        SqlName: p
        SqlType:
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
        SqlType: objectid
`)
