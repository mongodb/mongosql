package results

import (
	"strings"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson"
)

// ColumnType is the type of a column.
type ColumnType struct {
	EvalType    types.EvalType
	MongoType   schema.MongoType
	UUIDSubType types.EvalType
}

// Column contains information used to select data
// from a PlanStage.
type Column struct {
	ColumnType          *ColumnType `bson:"columnType"`
	SelectID            int         `bson:"selectID"`
	Table               string      `bson:"-"`
	OriginalTable       string      `bson:"originalTable"`
	Database            string      `bson:"database"`
	Name                string      `bson:"name"`
	OriginalName        string      `bson:"-"`
	MappingRegistryName string      `bson:"-"`
	MongoName           string      `bson:"mongoName"`
	PrimaryKey          bool        `bson:"primaryKey"`
	Comments            string      `bson:"comments"`
	IsPolymorphic       bool        `bson:"isPolymorphic"`
	HasAlteredType      bool        `bson:"hasAlteredType"`
	Nullable            bool        `bson:"nullable"`
}

// NewColumn is a constructor for the Column struct.
func NewColumn(selectID int, table, originalTable, database, name,
	originalName, mappingRegistryName string, evalType types.EvalType,
	mongoType schema.MongoType, primaryKey, nullable bool) *Column {
	uuidSubType := types.EvalBinary
	if mongoType == schema.MongoUUIDJava {
		uuidSubType = types.EvalJavaUUID
	} else if mongoType == schema.MongoUUIDCSharp {
		uuidSubType = types.EvalCSharpUUID
	}
	return &Column{
		ColumnType: &ColumnType{
			MongoType:   mongoType,
			EvalType:    evalType,
			UUIDSubType: uuidSubType,
		},
		SelectID:            selectID,
		Table:               table,
		OriginalTable:       originalTable,
		Database:            database,
		Name:                name,
		OriginalName:        originalName,
		MappingRegistryName: mappingRegistryName,
		PrimaryKey:          primaryKey,
		Nullable:            nullable,
	}
}

// NewColumnType returns a ColumnType with the specified types.EvalType and MongoType.
func NewColumnType(evalType types.EvalType, mongoType schema.MongoType) *ColumnType {
	return &ColumnType{
		EvalType:  evalType,
		MongoType: mongoType,
		// Because the need to set the UUIDSubType is so rare, we just use
		// the default EvalBinary encoding unless otherwise specified with the
		// other constructor.
		UUIDSubType: types.EvalBinary,
	}
}

// NewColumnTypeWithUUIDSubtype returns a ColumnType with the specified types.EvalType, MongoType, and
// UUIDSubType.
func NewColumnTypeWithUUIDSubtype(evalType types.EvalType,
	mongoType schema.MongoType,
	uuidSubType types.EvalType) ColumnType {
	return ColumnType{
		EvalType:    evalType,
		MongoType:   mongoType,
		UUIDSubType: uuidSubType,
	}
}

// Clone clones the Column.
func (c *Column) Clone() *Column {
	cb := NewColumnBuilder()
	cb.SetColumnType(NewColumnType(c.ColumnType.EvalType, c.ColumnType.MongoType))
	cb.SetSelectID(c.SelectID)
	cb.SetTable(c.Table)
	cb.SetOriginalTable(c.OriginalTable)
	cb.SetDatabase(c.Database)
	cb.SetName(c.Name)
	cb.SetOriginalName(c.OriginalName)
	cb.SetMappingRegistryName(c.MappingRegistryName)
	cb.SetMongoName(c.MongoName)
	cb.SetPrimaryKey(c.PrimaryKey)
	cb.SetComments(c.Comments)
	cb.SetIsPolymorphic(c.IsPolymorphic)
	cb.SetHasAlteredType(c.HasAlteredType)
	cb.SetNullable(c.Nullable)
	return cb.Build()
}

type marshallableColumnType struct {
	EvalType    string `bson:"sqlType"`
	MongoType   string `bson:"mongoType"`
	UUIDSubType string `bson:"uuidSubType"`
}

// MarshalBSON is a custom implementation for marshalling a ColumnType
// into raw BSON bytes. ColumnType deviates from the default BSON
// marshalling implementation by marshalling the byte fields (EvalType
// and UUIDSubType) as strings. This is necessary because single bytes
// cannot be marshaled and unmarshaled.
func (c *ColumnType) MarshalBSON() ([]byte, error) {
	mct := &marshallableColumnType{
		EvalType:    types.EvalTypeToString(c.EvalType),
		MongoType:   string(c.MongoType),
		UUIDSubType: types.EvalTypeToString(c.UUIDSubType),
	}
	return bson.Marshal(mct)
}

// UnmarshalBSON unmarshals the provided raw bytes into the Column.
func (c *ColumnType) UnmarshalBSON(b []byte) error {
	mct := &marshallableColumnType{}
	err := bson.Unmarshal(b, mct)
	if err != nil {
		return err
	}

	c.EvalType = types.EvalTypeFromString(mct.EvalType)
	c.MongoType = schema.MongoType(mct.MongoType)
	c.UUIDSubType = types.EvalTypeFromString(mct.UUIDSubType)

	return nil
}

// Columns is a slice of Column pointers.
type Columns []*Column

// FindByName searches Columns for a column of a matching name.
func (cs Columns) FindByName(name string) (*Column, bool) {
	for _, c := range cs {
		if strings.EqualFold(name, c.Name) {
			return c, true
		}
	}

	return nil, false
}

// Unique ensures that only unique columns exist in the resulting slice.
func (cs Columns) Unique() Columns {
	var results Columns
	contains := func(column *Column) bool {
		for _, c := range results {
			if c.SelectID == column.SelectID &&
				c.Name == column.Name &&
				c.Table == column.Table &&
				c.Database == column.Database {
				return true
			}
		}

		return false
	}

	for _, c := range cs {
		if !contains(c) {
			results = append(results, c)
		}
	}

	return results
}

// Names returns the slice of column names.
func (cs Columns) Names() []string {
	columnNames := make([]string, len(cs))
	for i, col := range cs {
		columnNames[i] = col.Name
	}

	return columnNames
}

// ColumnInfo keeps track of the data needed to correctly deserialize data from
// a MongoSourceStage.
type ColumnInfo struct {
	// Field is the name of the specific MongoDB field.
	Field string
	// Type is the byte corresponding to the type MongoDRDL specifies for
	// the given column. The byte corresponds to the BSON kind byte, iff
	// the column type is a BSON type. Some Column types are not BSON
	// types: e.g., Date, which needs to drop the Time portions of a
	// Timestamp for formatting purposes because BSON datetime objects
	// store both the date and the time. This is represented using
	// the type alias EvalType.
	Type types.EvalType
	// UUIDSubtype is needed to handle UUIDs written by the Java and CSharp
	// drivers, which store UUIDs using different byte orders.
	UUIDSubtype types.EvalType
}
