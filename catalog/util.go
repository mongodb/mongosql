package catalog

import "github.com/10gen/sqlproxy/schema"

func translateColumnType(sqlType schema.SQLType) string {
	switch sqlType {
	case schema.SQLBoolean:
		return "tinyint(1)"
	case schema.SQLDate:
		return "date"
	case schema.SQLDecimal128:
		return "decimal"
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
		return "varchar(65535)"
	default:
		return "<unknown>"
	}
}
