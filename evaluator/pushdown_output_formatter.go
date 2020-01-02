package evaluator

import (
	"fmt"

	"github.com/10gen/mongoast/ast"
	"github.com/10gen/sqlproxy/internal/astutil"
)

type pushdownOutputFormatter func(*MongoSourceStage) (*ast.ProjectStage, error)

const (
	// NoOutputFormat is the empty string. An empty "format" means that
	// the output of pushdown does not need to be specially formatted.
	NoOutputFormat = ""

	// NoOutputVersion should be used in conjunction with NoOutputFormat.
	// If not format is specified, then the version is ignored anyway.
	NoOutputVersion = 0

	// JDBCOutputFormat is the jdbc pushdown output format name.
	JDBCOutputFormat = "jdbc"

	// ODBCOutputFormat is the odbc pushdown output format name.
	ODBCOutputFormat = "odbc"
)

var formatters = map[string]map[int]pushdownOutputFormatter{
	JDBCOutputFormat: {
		1: formatUnflattenedProject,
	},
	ODBCOutputFormat: {
		1: formatUnflattenedProject,
	},
}

// getPushdownOutputFormatter gets the formatter function for the provided
// format and formatVersion. If such a formatter does not exist, this function
// returns an error.
func getPushdownOutputFormatter(format string, formatVersion int) (pushdownOutputFormatter, error) {
	if versionedFormatters, ok := formatters[format]; ok {
		if formatter, ok := versionedFormatters[formatVersion]; ok {
			return formatter, nil
		}

		return nil, fmt.Errorf("unknown format version for %q: %v", format, formatVersion)
	}

	return nil, fmt.Errorf("unknown format: %q", format)
}

// formatPushdownOutput returns a $project stage formatting the output of the
// MongoSourceStage according to the format and formatVersion arguments.
func formatPushdownOutput(ms *MongoSourceStage, format string, formatVersion int) (*ast.ProjectStage, error) {
	formatter, err := getPushdownOutputFormatter(format, formatVersion)
	if err != nil {
		return nil, err
	}

	return formatter(ms)
}

// formatUnflattenedProject returns a rich $project stage for outputting
// the resulting column data. Typically, the final $project stage of a
// BIC translation is a flat document (i.e. no nesting). The namespace
// information is condensed into key names and values are all scalar.
//
// For example,
//     { $project: {"db_DOT_foo_DOT__id": "$_id", "db_DOT_foo_DOT_a": "$a", ... } }
// returns "$foo._id" and "$foo.a" from the database "db" in a flat manner.
//
// This function returns a $project stage with rich, nested namespace data,
// including alias information for the column and table. The $project stage
// projects only one field, "values", which is an array of all selected
// expressions, ordered by their select order in the original SQL query.
func formatUnflattenedProject(ms *MongoSourceStage) (*ast.ProjectStage, error) {
	columns := ms.Columns()
	richFieldData := make([]ast.Expr, len(columns))

	for i, c := range columns {
		ref, ok := ms.mappingRegistry.lookupFieldRef(c.Database, c.Table, c.Name)
		if !ok {
			return nil, fmt.Errorf("failed to find field ref for column '%v.%v.%v'", c.Database, c.Table, c.Name)
		}

		db := astutil.StringValue(c.Database)
		table := astutil.StringValue(c.OriginalTable)
		tableAlias := astutil.StringValue(c.Table)
		column := astutil.StringValue(c.OriginalName)
		columnAlias := astutil.StringValue(c.Name)

		// if db is empty, replace it with null
		if c.Database == "" {
			db = astutil.NullLiteral
		}

		// if table is empty, replace it with null
		if c.OriginalTable == "" {
			table = astutil.NullLiteral
		}

		// if tableAlias is empty, replace it with table
		if c.Table == "" {
			tableAlias = table
		}

		// if column is empty, replace it with null
		if c.OriginalName == "" {
			column = astutil.NullLiteral
		}

		// if columnAlias is empty, replace it with column
		if c.Name == "" {
			columnAlias = column
		}

		richFieldData[i] = ast.NewDocument(
			ast.NewDocumentElement("db", db),
			ast.NewDocumentElement("table", table),
			ast.NewDocumentElement("tableAlias", tableAlias),
			ast.NewDocumentElement("column", column),
			ast.NewDocumentElement("columnAlias", columnAlias),
			ast.NewDocumentElement("value", ref),
		)
	}

	return ast.NewProjectStage(
		ast.NewAssignProjectItem("values", ast.NewArray(richFieldData...)),
		ast.NewExcludeProjectItem(ast.NewFieldRef("_id", nil)),
	), nil
}
