package mongosql

import (
	"go.mongodb.org/mongo-driver/bson"
)

// ResultSetMetadata contains information needed by data lake to
// return result set metadata along with a $sql result set.
type ResultSetMetadata struct {
	Columns []*ColumnMetadata
}

// ColumnMetadata contains information about a single column that is
// returned along with a $sql result set.
type ColumnMetadata struct {
	Database    string
	Table       string
	TableAlias  string
	Column      string
	ColumnAlias string
	BsonType    string
}

// EmptyResultSetDoc returns a BSON document specifying an $addFields
// stage that is used by data lake for handling empty result sets for
// ODBC/JDBC formatVersion 1.
func (meta *ResultSetMetadata) EmptyResultSetAddFields() ([]byte, error) {
	columnsWithValues := make([]bson.D, len(meta.Columns))
	for i, col := range meta.Columns {
		columnsWithValues[i] = bson.D{
			{"database", nullIfEmpty(col.Database)},
			{"table", nullIfEmpty(col.Table)},
			{"tableAlias", nullIfEmpty(col.TableAlias)},
			{"column", nullIfEmpty(col.Column)},
			{"columnAlias", nullIfEmpty(col.ColumnAlias)},
			{"bsonType", col.BsonType},
			{"value", nil},
		}
	}

	addFields := bson.D{
		{"$addFields", bson.D{
			{"emptyResultSet", true},
			{"values", columnsWithValues},
		}},
	}

	return bson.Marshal(addFields)
}

// MetadataDoc returns a BSON document specifying column metadata that
// is used by data lake as the first document returned for ODBC/JDBC
// formatVersion 2.
func (meta *ResultSetMetadata) MetadataDoc() ([]byte, error) {
	columns := make([]bson.D, len(meta.Columns))
	for i, col := range meta.Columns {
		columns[i] = bson.D{
			{"database", nullIfEmpty(col.Database)},
			{"table", nullIfEmpty(col.Table)},
			{"tableAlias", nullIfEmpty(col.TableAlias)},
			{"column", nullIfEmpty(col.Column)},
			{"columnAlias", nullIfEmpty(col.ColumnAlias)},
			{"bsonType", col.BsonType},
		}
	}
	return bson.Marshal(bson.D{{"columns", columns}})
}

func nullIfEmpty(name string) interface{} {
	if name == "" {
		return nil
	}
	return name
}
