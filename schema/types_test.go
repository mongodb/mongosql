package schema_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestGetMongoType(t *testing.T) {
	req := require.New(t)

	sqlType, err := schema.GetMongoType("bool")
	req.Equal(sqlType, schema.MongoBool)
	req.Equal(err, nil)
}

func TestGetMongoTypeInvalidMongoType(t *testing.T) {
	req := require.New(t)

	_, err := schema.GetMongoType("invalidtype")
	req.Equal(fmt.Errorf(`invalid Mongo type "%v"`, "invalidtype"), err)
}

func TestGetSQLType(t *testing.T) {
	req := require.New(t)

	sqlType, err := schema.GetSQLType("date")
	req.Equal(sqlType, schema.SQLDate)
	req.Equal(err, nil)
}

func TestGetSQLTypeWithAlias(t *testing.T) {
	req := require.New(t)

	sqlType, err := schema.GetSQLType("int32")
	req.Equal(sqlType, schema.SQLInt)
	req.Equal(err, nil)
}
func TestGetSQLTypeInvalidSQLType(t *testing.T) {
	req := require.New(t)

	_, err := schema.GetSQLType("notvalidtype")
	req.Equal(err, fmt.Errorf(`invalid SQL type "notvalidtype"`))
}
