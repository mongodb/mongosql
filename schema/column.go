package schema

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/10gen/sqlproxy/schema/mongo"
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
	// sampledTypes are the types for a column that have
	// been sampled from MongoDB. It only applies to columns
	// derived from MongoDB (not, for instance, dynamic columns).
	sampledTypes []string
	// hasAlteredType keeps track of if the type of this column
	// has been altered, as we must wrap references to this
	// column in converts if it has been.
	hasAlteredType bool
	// nullable is true if this column can contain NULLs.
	nullable bool
	// comment is a string added by the user during table creation.
	comment option.String
}

// NewColumn returns a new Column with the provided fields.
// If the mongoName is empty, reuse the sqlName.
func NewColumn(sqlName string, sqlType SQLType,
	mongoName string, mongoType MongoType,
	nullable bool, comment option.String) *Column {
	if mongoName == "" {
		mongoName = sqlName
	} else if sqlName == "" {
		sqlName = mongoName
	}

	if sqlType == SQLVarchar && mongoType == MongoObjectID {
		sqlType = SQLObjectID
	}

	return &Column{
		sqlName:        sqlName,
		sqlType:        sqlType,
		mongoName:      mongoName,
		mongoType:      mongoType,
		sampledTypes:   []string{},
		hasAlteredType: false,
		nullable:       nullable,
		comment:        comment,
	}
}

// NewColumnWithSampledTypes returns a new Column with the provided fields.
// If the mongoName is empty, reuse the sqlName.
func NewColumnWithSampledTypes(sqlName string, sqlType SQLType, mongoName string,
	mongoType MongoType, sampledTypes []mongo.BSONType) *Column {
	stringSampledTypes := make([]string, len(sampledTypes))
	for i, v := range sampledTypes {
		stringSampledTypes[i] = string(v)
	}
	ret := NewColumn(sqlName, sqlType, mongoName, mongoType, false, option.NoneString())
	ret.sampledTypes = stringSampledTypes
	return ret
}

// NewColumnFromDRDL creates a new Column from the provided drdl column.
func NewColumnFromDRDL(drdlCol *drdl.Column) (*Column, error) {
	sqlType, err := GetSQLType(drdlCol.SQLType)
	if err != nil {
		return nil, fmt.Errorf(`unsupported SQL type: "%v" on column "%v"`,
			drdlCol.SQLType, drdlCol.SQLName)
	}

	mongoType, err := GetMongoType(drdlCol.MongoType)
	if err != nil {
		return nil, fmt.Errorf(`unsupported Mongo type: "%v" on column "%v"`,
			drdlCol.MongoType, drdlCol.MongoName)
	}

	return NewColumn(
		drdlCol.SQLName,
		sqlType,
		drdlCol.MongoName,
		mongoType,
		false,
		option.NoneString(),
	), nil
}

// DeepCopy returns a deep copy of this Column.
func (c *Column) DeepCopy() *Column {
	copiedSampledTypes := make([]string, len(c.sampledTypes))
	copy(copiedSampledTypes, c.sampledTypes)
	newC := NewColumn(
		c.sqlName,
		c.sqlType,
		c.mongoName,
		c.mongoType,
		c.nullable,
		c.comment,
	)
	newC.sampledTypes = copiedSampledTypes
	newC.hasAlteredType = c.hasAlteredType
	return newC
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

// HasTypeAlteration returns true if this column has had a type alteration.
// This is necessary for inserting converts in the polymorphic_type_conversion
// mode 'fast'.
func (c *Column) HasTypeAlteration() bool {
	return c.hasAlteredType
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

// SampledTypes returns this Column's SampledTypes.
func (c *Column) SampledTypes() []string {
	return c.sampledTypes
}

// Nullable returns true if this column can have NULL values.
func (c *Column) Nullable() bool {
	return c.nullable
}

// Comment returns the comment for this column.
func (c *Column) Comment() option.String {
	return c.comment
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
	case MongoObjectID:
		switch c.sqlType {
		case SQLObjectID:
			err = nil
		}
	case MongoFilter, MongoString, MongoUUID,
		MongoUUIDCSharp, MongoUUIDJava, MongoUUIDOld:
		if c.sqlType == SQLVarchar {
			err = nil
		}
	default:
		err = fmt.Errorf("unsupported mongo type: '%s'", c.mongoType)
	}

	return err
}
