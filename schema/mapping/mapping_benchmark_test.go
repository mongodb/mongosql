package mapping_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
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
		err := mapping.Map(
			mapping.NewSchemaMappingConfig(db,
				mongoSchema,
				"testcol",
				false,
				"",
				[]uint8{4, 0, 0},
				log.GlobalLogger(),
				config.MajorityMappingMode,
				1000,
				50))
		req.NoError(err, "failed to map MongoDB schema to relational schema")
	}
}

func createMongoSchema(cols int) (*mongo.Schema, error) {
	schema := mongo.NewCollectionSchema()

	var i int
	for i < cols {
		colName := fmt.Sprintf("field_%d", i)
		doc := bsonutil.NewD(bsonutil.NewDocElem(colName, "value"))
		err := schema.IncludeSample(doc)
		if err != nil {
			return nil, err
		}
		i++
	}

	return schema, nil
}
