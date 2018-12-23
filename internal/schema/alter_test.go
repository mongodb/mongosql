package schema_test

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/schema"
	"github.com/stretchr/testify/require"
)

func TestAlter(t *testing.T) {
	t.Run("rename_column_simple", testRenameColumn)
	t.Run("rename_column_twice", testRenameColumnTwice)
	t.Run("rename_table_simple", testRenameTable)
	t.Run("rename_table_twice", testRenameTableTwice)
}

func testRenameColumn(t *testing.T) {
	req := require.New(t)

	rename := &schema.Alteration{
		Type:      schema.RenameColumn,
		Db:        "testDb",
		Table:     "testTable",
		Column:    "foo",
		NewColumn: "bar",
	}

	sch := createTestSchema()
	sch.AddAlterations(rename)

	altered, err := sch.Altered()
	req.NoError(err, "failed to alter schema")
	req.Equal(
		"bar",
		altered.Databases()[0].Tables()[0].Columns()[0].SQLName(),
		"incorrect SQLName for column after alterations",
	)
}

func testRenameColumnTwice(t *testing.T) {
	req := require.New(t)

	firstRename := &schema.Alteration{
		Type:      schema.RenameColumn,
		Db:        "testDb",
		Table:     "testTable",
		Column:    "foo",
		NewColumn: "bar",
	}

	secondRename := &schema.Alteration{
		Type:      schema.RenameColumn,
		Db:        "testDb",
		Table:     "testTable",
		Column:    "bar",
		NewColumn: "baz",
	}

	sch := createTestSchema()
	sch.AddAlterations(firstRename, secondRename)

	altered, err := sch.Altered()
	req.NoError(err, "failed to alter schema")
	req.Equal(
		"baz",
		altered.Databases()[0].Tables()[0].Columns()[0].SQLName(),
		"incorrect SQLName for column after alterations",
	)
}

func testRenameTable(t *testing.T) {
	req := require.New(t)

	rename := &schema.Alteration{
		Type:     schema.RenameTable,
		Db:       "testDb",
		Table:    "testTable",
		NewTable: "foo",
	}

	sch := createTestSchema()
	sch.AddAlterations(rename)

	altered, err := sch.Altered()
	req.NoError(err, "failed to alter schema")
	req.Equal(
		"foo",
		altered.Databases()[0].Tables()[0].SQLName(),
		"incorrect SQLName for table after alterations",
	)
}

func testRenameTableTwice(t *testing.T) {
	req := require.New(t)

	firstRename := &schema.Alteration{
		Type:     schema.RenameTable,
		Db:       "testDb",
		Table:    "testTable",
		NewTable: "foo",
	}

	secondRename := &schema.Alteration{
		Type:     schema.RenameTable,
		Db:       "testDb",
		Table:    "foo",
		NewTable: "bar",
	}

	sch := createTestSchema()
	sch.AddAlterations(firstRename, secondRename)

	altered, err := sch.Altered()
	req.NoError(err, "failed to alter schema")
	req.Equal(
		"bar",
		altered.Databases()[0].Tables()[0].SQLName(),
		"incorrect SQLName for table after alterations",
	)
}

func createTestSchema() *schema.Schema {
	cols := []*schema.Column{
		schema.NewColumn("foo", schema.SQLInt, "mongo_foo", schema.MongoInt),
	}
	table, _ := schema.NewTable(
		lg,
		"testTable", "testCollection",
		nil, cols,
	)
	db := schema.NewDatabase(lg, "testDb", []*schema.Table{table})
	s, _ := schema.New([]*schema.Database{db}, nil)
	return s
}
