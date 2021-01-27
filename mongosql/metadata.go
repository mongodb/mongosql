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

// EmptyResultSetAddFields returns a BSON document specifying an $addFields
// stage that is used by data lake for handling empty result sets for
// ODBC/JDBC formatVersion 1.
func (meta *ResultSetMetadata) EmptyResultSetAddFields() ([]byte, error) {
	columnsWithValues := make([]bson.D, len(meta.Columns))
	for i, col := range meta.Columns {
		columnsWithValues[i] = bson.D{
			{Key: "database", Value: nullIfEmpty(col.Database)},
			{Key: "table", Value: nullIfEmpty(col.Table)},
			{Key: "tableAlias", Value: nullIfEmpty(col.TableAlias)},
			{Key: "column", Value: nullIfEmpty(col.Column)},
			{Key: "columnAlias", Value: nullIfEmpty(col.ColumnAlias)},
			{Key: "bsonType", Value: col.BsonType},
			{Key: "value", Value: nil},
		}
	}

	addFields := bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "emptyResultSet", Value: true},
			{Key: "values", Value: columnsWithValues},
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
			{Key: "database", Value: nullIfEmpty(col.Database)},
			{Key: "table", Value: nullIfEmpty(col.Table)},
			{Key: "tableAlias", Value: nullIfEmpty(col.TableAlias)},
			{Key: "column", Value: nullIfEmpty(col.Column)},
			{Key: "columnAlias", Value: nullIfEmpty(col.ColumnAlias)},
			{Key: "bsonType", Value: col.BsonType},
		}
	}
	return bson.Marshal(bson.D{{Key: "columns", Value: columns}})
}

func nullIfEmpty(name string) interface{} {
	if name == "" {
		return nil
	}
	return name
}
