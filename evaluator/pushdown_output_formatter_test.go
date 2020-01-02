package evaluator

import (
	"testing"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/sqlproxy/evaluator/results"
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
		{"jdbc v2 does not exist", "jdbc", 2, true},
		{"odbc v1 exists", "odbc", 1, false},
		{"odbc v2 does not exist", "odbc", 2, true},
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
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "a", OriginalName: "a"},
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "b", OriginalName: "b"},
			},
			ast.NewProjectStage(
				ast.NewAssignProjectItem("values", ast.NewArray(
					ast.NewDocument(
						ast.NewDocumentElement("db", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("foo")),
						ast.NewDocumentElement("column", astutil.StringValue("a")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("a")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_foo_DOT_a", nil)),
					),
					ast.NewDocument(
						ast.NewDocumentElement("db", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("foo")),
						ast.NewDocumentElement("column", astutil.StringValue("b")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("b")),
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
				{Database: "db", Table: "t1", OriginalTable: "foo", Name: "ah", OriginalName: "a"},
				{Database: "db", Table: "t2", OriginalTable: "foo", Name: "be", OriginalName: "b"},
			},
			ast.NewProjectStage(
				ast.NewAssignProjectItem("values", ast.NewArray(
					ast.NewDocument(
						ast.NewDocumentElement("db", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("t1")),
						ast.NewDocumentElement("column", astutil.StringValue("a")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("ah")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_t1_DOT_ah", nil)),
					),
					ast.NewDocument(
						ast.NewDocumentElement("db", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("t2")),
						ast.NewDocumentElement("column", astutil.StringValue("b")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("be")),
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
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "a", OriginalName: "a"},
				{Database: "db", Table: "foo", OriginalTable: "foo", Name: "b", OriginalName: "b"},
				{Database: "db", Table: "", OriginalTable: "", Name: "a+b", OriginalName: ""},
			},
			ast.NewProjectStage(
				ast.NewAssignProjectItem("values", ast.NewArray(
					ast.NewDocument(
						ast.NewDocumentElement("db", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("foo")),
						ast.NewDocumentElement("column", astutil.StringValue("a")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("a")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_foo_DOT_a", nil)),
					),
					ast.NewDocument(
						ast.NewDocumentElement("db", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.StringValue("foo")),
						ast.NewDocumentElement("tableAlias", astutil.StringValue("foo")),
						ast.NewDocumentElement("column", astutil.StringValue("b")),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("b")),
						ast.NewDocumentElement("value", ast.NewFieldRef("db_DOT_foo_DOT_b", nil)),
					),
					ast.NewDocument(
						ast.NewDocumentElement("db", astutil.StringValue("db")),
						ast.NewDocumentElement("table", astutil.NullLiteral),
						ast.NewDocumentElement("tableAlias", astutil.NullLiteral),
						ast.NewDocumentElement("column", astutil.NullLiteral),
						ast.NewDocumentElement("columnAlias", astutil.StringValue("a+b")),
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
