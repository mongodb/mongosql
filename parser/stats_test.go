package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryStats(t *testing.T) {
	tests := []struct {
		name       string
		sql        string
		funcs      map[string]int
		joins      map[string]int
		unions     map[string]int
		subqueries map[string]int
	}{
		{
			name: "simple",
			sql:  "select * from foo",
		},
		{
			name: "funcs",
			sql:  "select md5(a), ltrim(b) from foo",
			funcs: map[string]int{
				"md5":   1,
				"ltrim": 1,
			},
		},
		{
			name: "funcs_multiple",
			sql:  "select md5(a), ltrim(b) from foo where md5(c) is not null",
			funcs: map[string]int{
				"md5":   2,
				"ltrim": 1,
			},
		},
		{
			name: "funcs_agg",
			sql:  "select sum(amount) from orders group by item",
			funcs: map[string]int{
				"sum": 1,
			},
		},
		{
			name: "joins",
			sql:  "select * from a, b join c straight_join d cross join e",
			joins: map[string]int{
				"comma":         1,
				"join":          1,
				"straight_join": 1,
				"cross join":    1,
			},
		},
		{
			name: "joins_multiple",
			sql:  "select * from a join b join c, d cross join e",
			joins: map[string]int{
				"comma":      1,
				"join":       2,
				"cross join": 1,
			},
		},
		{
			name: "joins_outer",
			sql:  "select * from a left join b on true right join c on true",
			joins: map[string]int{
				"left join":  1,
				"right join": 1,
			},
		},
		{
			name: "joins_natural",
			sql:  "select * from a natural join b natural left join c natural right join d",
			joins: map[string]int{
				"natural join":       1,
				"natural left join":  1,
				"natural right join": 1,
			},
		},
		{
			name: "unions",
			sql:  "select * from foo union select * from bar union all select * from baz",
			unions: map[string]int{
				"union all": 1,
				"union":     1,
			},
		},
		{
			name: "unions_multiple",
			sql:  "select * from foo union all select * from bar union all select * from baz",
			unions: map[string]int{
				"union all": 2,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			// build expected QueryStats
			exp := newQueryStats()
			if test.funcs != nil {
				exp.Functions = test.funcs
			}
			if test.joins != nil {
				exp.Joins = test.joins
			}
			if test.unions != nil {
				exp.Unions = test.unions
			}
			if test.subqueries != nil {
				exp.Subqueries = test.subqueries
			}

			// parse sql
			stmt, err := Parse(test.sql)
			req.NoError(err)

			// get actual QueryStats
			stats := GetQueryStats(stmt)

			req.Equal(exp, stats)
		})
	}
}
