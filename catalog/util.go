package catalog

import (
	"fmt"
	"math"

	"github.com/10gen/sqlproxy/schema"
)

func translateColumnType(sqlType schema.SQLType, maxVarcharLength uint16) string {
	switch sqlType {
	case schema.SQLBoolean:
		return "tinyint(1)"
	case schema.SQLDate:
		return "date"
	case schema.SQLDecimal128:
		return "decimal(65,20)"
	case schema.SQLFloat, schema.SQLNumeric, schema.SQLArrNumeric:
		return "double"
	case schema.SQLInt, schema.SQLInt64:
		return "bigint(20)"
	case schema.SQLObjectID:
		return "varchar(24)"
	case schema.SQLTimestamp:
		return "datetime"
	case schema.SQLUint64:
		return "bigint(20) unsigned"
	case schema.SQLUUID:
		return "varchar(36)"
	case schema.SQLVarchar:
		length := maxVarcharLength
		if length == 0 {
			length = math.MaxUint16
		}
		return fmt.Sprintf("varchar(%d)", length)
	default:
		return "<unknown>"
	}
}
