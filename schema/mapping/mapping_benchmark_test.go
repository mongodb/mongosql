package mapping_test

import (
	"fmt"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/mapping"
	"github.com/10gen/sqlproxy/schema/mongo"
	"github.com/stretchr/testify/require"
)

func BenchmarkMapping(b *testing.B) {
	b.Run("100_columns", func(b *testing.B) { benchmarkMapWithColumnCount(b, 100) })
	b.Run("1000_columns", func(b *testing.B) { benchmarkMapWithColumnCount(b, 1000) })
	b.Run("10000_columns", func(b *testing.B) { benchmarkMapWithColumnCount(b, 10000) })
	b.Run("100000_columns", func(b *testing.B) { benchmarkMapWithColumnCount(b, 100000) })
}

func benchmarkMapWithColumnCount(b *testing.B, cols int) {
	req := require.New(b)

	mongoSchema, err := createMongoSchema(cols)
	req.NoError(err, "failed to create MongoDB schema")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		db := schema.NewDatabase(log.GlobalLogger(), "testdb", nil)
		err := mapping.Map(mapping.SchemaMappingConfig{
			Database:             db,
			Schema:               mongoSchema,
			CollectionName:       "testcol",
			UUIDSubtype3Encoding: "",
			Version:              []uint8{4, 0, 0},
			Logger:               log.GlobalLogger(),
		})
		req.NoError(err, "failed to map MongoDB schema to relational schema")
	}
}

func createMongoSchema(cols int) (*mongo.Schema, error) {
	schema := mongo.NewCollectionSchema()

	var i int
	for i < cols {
		colName := fmt.Sprintf("field_%d", i)
		doc := bson.D{{Name: colName, Value: "value"}}
		err := schema.IncludeSample(doc)
		if err != nil {
			return nil, err
		}
		i++
	}

	return schema, nil
}
