package schema_test

import (
	"testing"

	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestNewColumnNoSQLName(t *testing.T) {
	req := require.New(t)

	col := schema.NewColumn("", schema.SQLInt, "mongoname", schema.MongoInt)
	req.Equal("mongoname", col.SQLName(), "incorrect SQLName")
	req.Equal("mongoname", col.MongoName(), "incorrect MongoName")
}
