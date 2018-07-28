package schema

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/schema/drdl"
)

// Column represents the schema for a column.
type Column struct {
	// sqlName is the name of the column to be shown to users.
	sqlName string
	// sqlType is the type to be shown to users.
	sqlType SQLType
	// mongoName is the name of the field in MongoDB.
	mongoName string
	// mongoType is the type of the field in MongoDB.
	mongoType MongoType
}

// NewColumn returns a new Column with the provided fields.
// If the mongoName is empty, reuse the sqlName.
func NewColumn(sqlName string, sqlType SQLType, mongoName string, mongoType MongoType) *Column {
	if mongoName == "" {
		mongoName = sqlName
	} else if sqlName == "" {
		sqlName = mongoName
	}
	return &Column{
		sqlName:   sqlName,
		sqlType:   sqlType,
		mongoName: mongoName,
		mongoType: mongoType,
	}
}

// NewColumnFromDRDL creates a new Column from the provided drdl column.
func NewColumnFromDRDL(drdlCol *drdl.Column) *Column {
	return NewColumn(
		drdlCol.SQLName,
		GetSQLType(drdlCol.SQLType),
		drdlCol.MongoName,
		MongoType(drdlCol.MongoType),
	)
}

// DeepCopy returns a deep copy of this Column.
func (c *Column) DeepCopy() *Column {
	return &Column{
		mongoName: c.mongoName,
		mongoType: c.mongoType,
		sqlName:   c.sqlName,
		sqlType:   c.sqlType,
	}
}

// Equals checks whether this Column is equal to the provided Column.
func (c *Column) Equals(other *Column) error {
	if c == other {
		return nil
	}
	if c == nil {
		return fmt.Errorf("this table is nil, but other table is non-nil")
	}
	if other == nil {
		return fmt.Errorf("this table is non-nil, but other table is nil")
	}

	if c.mongoName != other.mongoName {
		return fmt.Errorf("mongoNames %q and %q do not match", c.mongoName, other.mongoName)
	}
	if c.mongoType != other.mongoType {
		return fmt.Errorf("mongoTypes %q and %q do not match", c.mongoType, other.mongoType)
	}
	if c.sqlName != other.sqlName {
		return fmt.Errorf("sqlNames %q and %q do not match", c.sqlName, other.sqlName)
	}
	if c.sqlType != other.sqlType {
		return fmt.Errorf("sqlTypes %q and %q do not match", c.sqlType, other.sqlType)
	}
	return nil
}

// MongoName returns this Column's MongoName.
func (c *Column) MongoName() string {
	return c.mongoName
}

// MongoType returns this Column's MongoType.
func (c *Column) MongoType() MongoType {
	return c.mongoType
}

// SQLName returns this Column's SQLName.
func (c *Column) SQLName() string {
	return c.sqlName
}

// SQLType returns this Column's SQLType.
func (c *Column) SQLType() SQLType {
	return c.sqlType
}

// Validate checks whether this Column is valid, returning an error if not.
func (c *Column) Validate() error {
	if strings.Trim(c.sqlName, " ") == "" {
		return fmt.Errorf("invalid SQLName %q", c.sqlName)
	}

	err := fmt.Errorf("cannot map mongo type '%s' to SQL type '%s'", c.mongoType, c.sqlType)
	switch c.mongoType {
	case MongoBool:
		if c.sqlType == SQLBoolean {
			err = nil
		}
	case MongoDate:
		switch c.sqlType {
		case SQLDate, SQLTimestamp:
			err = nil
		}
	case MongoDecimal128:
		switch c.sqlType {
		case SQLDecimal, SQLNumeric, SQLVarchar:
			err = nil
		}
	case MongoFloat:
		switch c.sqlType {
		case SQLFloat, SQLNumeric, SQLVarchar, SQLArrNumeric:
			err = nil
		}
	case MongoGeo2D:
		if c.sqlType == SQLArrNumeric {
			err = nil
		}
	case MongoInt:
		switch c.sqlType {
		case SQLInt, SQLNumeric, SQLVarchar:
			err = nil
		}
	case MongoInt64:
		switch c.sqlType {
		case SQLInt, SQLNumeric, SQLVarchar:
			err = nil
		}
	case MongoNumber:
		switch c.sqlType {
		case SQLInt, SQLFloat, SQLDecimal, SQLNumeric:
			err = nil
		}
	case MongoObjectID, MongoString, MongoFilter, MongoUUID,
		MongoUUIDCSharp, MongoUUIDJava, MongoUUIDOld:
		if c.sqlType == SQLVarchar {
			err = nil
		}
	default:
		err = fmt.Errorf("unsupported mongo type: '%s'", c.mongoType)
	}

	return err
}
