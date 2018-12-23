package schema_test

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/schema"
	"github.com/stretchr/testify/require"
)

func TestNewColumnNoSQLName(t *testing.T) {
	req := require.New(t)

	col := schema.NewColumn("sqlname", schema.SQLInt, "", schema.MongoInt)
	req.Equal("sqlname", col.SQLName(), "incorrect SQLName")
	req.Equal("sqlname", col.MongoName(), "incorrect MongoName")
}
