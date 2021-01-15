package evaluator

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/stretchr/testify/require"
)

// TestGetPushdownOutputFormatter tests that expected formats and formatVersions exist.
// It does not test the actual formatter functions.
func TestGetPushdownOutputFormatter(t *testing.T) {
	tests := []struct {
		name          string
		format        string
		version       int
		expectedError bool
	}{
		{"jdbc v1 exists", "jdbc", 1, false},
		{"jdbc v2 exists", "jdbc", 2, false},
		{"jdbc v99999 does not exist", "jdbc", 99999, true},
		{"odbc v1 exists", "odbc", 1, false},
		{"odbc v2 exists", "odbc", 2, false},
		{"odbc v99999 does not exist", "odbc", 99999, true},
		{"unknown does not exist", "unknown", 1, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			_, err := getPushdownOutputFormatter(test.format, test.version)
			if test.expectedError {
				req.Error(err)
			} else {
				req.NoError(err)
			}
		})
	}
}

func TestFormatUnflattenedProject(t *testing.T) {
	tests := []struct {
		name      string
		mrFields  map[string]map[string]map[string]string
		mrColumns []*results.Column
		expected  *ast.ProjectStage
	}{
		{
			"no aliases",
			map[string]map[string]map[string]string{
				"db": {
					"foo": {
						"a": "db_DOT_foo_DOT_a",
						"b": "db_DOT_foo_DOT_b",
					},
				},
			},
			[]*results.Column{
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "a", OriginalName: "a", ColumnType: &results.ColumnType{EvalType: types.EvalString}},
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "b", OriginalName: "b", ColumnType: &results.ColumnType{EvalType: types.EvalBoolean}},
			},
			ast.NewProjectStage(
				ast.NewAssignProjectItem("values", ast.NewArray(
					ast.NewDocument(
						ast.NewDocumentElement("database", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("foo")),
						ast.NewDocumentElement("column", astutil.StringValue("a")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("a")),
						ast.NewDocumentElement("bsonType", astutil.StringValue("string")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_foo_DOT_a", nil)),
					),
					ast.NewDocument(
						ast.NewDocumentElement("database", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("foo")),
						ast.NewDocumentElement("column", astutil.StringValue("b")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("b")),
						ast.NewDocumentElement("bsonType", astutil.StringValue("bool")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_foo_DOT_b", nil)),
					),
				)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
			),
		},
		{
			"table and column aliases",
			map[string]map[string]map[string]string{
				"db": {
					"t1": {
						"ah": "db_DOT_t1_DOT_ah",
					},
					"t2": {
						"be": "db_DOT_t2_DOT_be",
					},
				},
			},
			[]*results.Column{
				{Database: "db", Table: "t1", OriginalTable: "foo", Name: "ah", OriginalName: "a", ColumnType: &results.ColumnType{EvalType: types.EvalString}},
				{Database: "db", Table: "t2", OriginalTable: "foo", Name: "be", OriginalName: "b", ColumnType: &results.ColumnType{EvalType: types.EvalInt64}},
			},
			ast.NewProjectStage(
				ast.NewAssignProjectItem("values", ast.NewArray(
					ast.NewDocument(
						ast.NewDocumentElement("database", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("t1")),
						ast.NewDocumentElement("column", astutil.StringValue("a")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("ah")),
						ast.NewDocumentElement("bsonType", astutil.StringValue("string")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_t1_DOT_ah", nil)),
					),
					ast.NewDocument(
						ast.NewDocumentElement("database", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("t2")),
						ast.NewDocumentElement("column", astutil.StringValue("b")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("be")),
						ast.NewDocumentElement("bsonType", astutil.StringValue("long")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_t2_DOT_be", nil)),
					),
				)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
			),
		},
		{
			"missing table names and original name (computed column)",
			map[string]map[string]map[string]string{
				"db": {
					"foo": {
						"a": "db_DOT_foo_DOT_a",
						"b": "db_DOT_foo_DOT_b",
					},
					"": {
						"a+b": "db_DOT_foo_DOT_a+db_DOT_foo_DOT_b",
					},
				},
			},
			[]*results.Column{
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "a", OriginalName: "a", ColumnType: &results.ColumnType{EvalType: types.EvalDecimal128}},
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "b", OriginalName: "b", ColumnType: &results.ColumnType{EvalType: types.EvalInt32}},
				{Database: "db", Table: "", OriginalTable: "", Name: "a+b", OriginalName: "", ColumnType: &results.ColumnType{EvalType: types.EvalDouble}},
			},
			ast.NewProjectStage(
				ast.NewAssignProjectItem("values", ast.NewArray(
					ast.NewDocument(
						ast.NewDocumentElement("database", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("foo")),
						ast.NewDocumentElement("column", astutil.StringValue("a")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("a")),
						ast.NewDocumentElement("bsonType", astutil.StringValue("decimal")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_foo_DOT_a", nil)),
					),
					ast.NewDocument(
						ast.NewDocumentElement("database", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("foo")),
						ast.NewDocumentElement("column", astutil.StringValue("b")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("b")),
						ast.NewDocumentElement("bsonType", astutil.StringValue("int")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_foo_DOT_b", nil)),
					),
					ast.NewDocument(
						ast.NewDocumentElement("database", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.NullLiteral),
						ast.NewDocumentElement("tableAlias", astutil.NullLiteral),
						ast.NewDocumentElement("column", astutil.NullLiteral),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("a+b")),
						ast.NewDocumentElement("bsonType", astutil.StringValue("double")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_foo_DOT_a+db_DOT_foo_DOT_b", nil)),
					),
				)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			ms := &MongoSourceStage{
				mappingRegistry: &mappingRegistry{
					fields:  test.mrFields,
					columns: test.mrColumns,
				},
			}

			actual, err := formatUnflattenedProject(ms)
			req.NoError(err, "unexpected error")
			req.Equal(test.expected, actual, "actual stage different from expected stage")
		})
	}
}

func TestFormatValuesArray(t *testing.T) {
	tests := []struct {
		name      string
		mrFields  map[string]map[string]map[string]string
		mrColumns []*results.Column
		expected  *ast.ProjectStage
	}{
		{
			"no aliases",
			map[string]map[string]map[string]string{
				"db": {
					"foo": {
						"a": "db_DOT_foo_DOT_a",
						"b": "db_DOT_foo_DOT_b",
					},
				},
			},
			[]*results.Column{
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "a", OriginalName: "a", ColumnType: &results.ColumnType{EvalType: types.EvalString}},
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "b", OriginalName: "b", ColumnType: &results.ColumnType{EvalType: types.EvalBoolean}},
			},
			ast.NewProjectStage(
				ast.NewAssignProjectItem("values", ast.NewArray(
					ast.NewFieldRef("db_DOT_foo_DOT_a", nil),
					ast.NewFieldRef("db_DOT_foo_DOT_b", nil),
				)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
			),
		},
		{
			"table and column aliases",
			map[string]map[string]map[string]string{
				"db": {
					"t1": {
						"ah": "db_DOT_t1_DOT_ah",
					},
					"t2": {
						"be": "db_DOT_t2_DOT_be",
					},
				},
			},
			[]*results.Column{
				{Database: "db", Table: "t1", OriginalTable: "foo", Name: "ah", OriginalName: "a", ColumnType: &results.ColumnType{EvalType: types.EvalString}},
				{Database: "db", Table: "t2", OriginalTable: "foo", Name: "be", OriginalName: "b", ColumnType: &results.ColumnType{EvalType: types.EvalInt64}},
			},
			ast.NewProjectStage(
				ast.NewAssignProjectItem("values", ast.NewArray(
					ast.NewFieldRef("db_DOT_t1_DOT_ah", nil),
					ast.NewFieldRef("db_DOT_t2_DOT_be", nil),
				)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
			),
		},
		{
			"missing table names and original name (computed column)",
			map[string]map[string]map[string]string{
				"db": {
					"foo": {
						"a": "db_DOT_foo_DOT_a",
						"b": "db_DOT_foo_DOT_b",
					},
					"": {
						"a+b": "db_DOT_foo_DOT_a+db_DOT_foo_DOT_b",
					},
				},
			},
			[]*results.Column{
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "a", OriginalName: "a", ColumnType: &results.ColumnType{EvalType: types.EvalDecimal128}},
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "b", OriginalName: "b", ColumnType: &results.ColumnType{EvalType: types.EvalInt32}},
				{Database: "db", Table: "", OriginalTable: "", Name: "a+b", OriginalName: "", ColumnType: &results.ColumnType{EvalType: types.EvalDouble}},
			},
			ast.NewProjectStage(
				ast.NewAssignProjectItem("values", ast.NewArray(
					ast.NewFieldRef("db_DOT_foo_DOT_a", nil),
					ast.NewFieldRef("db_DOT_foo_DOT_b", nil),
					ast.NewFieldRef("db_DOT_foo_DOT_a+db_DOT_foo_DOT_b", nil),
				)),
				ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			ms := &MongoSourceStage{
				mappingRegistry: &mappingRegistry{
					fields:  test.mrFields,
					columns: test.mrColumns,
				},
			}

			actual, err := formatValuesArray(ms)
			req.NoError(err, "unexpected error")
			req.Equal(test.expected, actual, "actual stage different from expected stage")
		})
	}
}
