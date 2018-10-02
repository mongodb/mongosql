package sample_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/stretchr/testify/require"
)

const (
	sampleViewDb             = "sampleTestView"
	sampleViewBaseCollection = "sampleViewBaseCollection"
)

func BenchmarkViewSampling(b *testing.B) {
	// Set mongosqld's log verbosity to quiet so benchmark results can be parsed correctly.
	log.SetVerbosity(log.Quiet)
	b.Run("100_lookups", func(b *testing.B) { benchmarkViewSamplingWithLookupCount(b, 100) })
	b.Run("500_lookups", func(b *testing.B) { benchmarkViewSamplingWithLookupCount(b, 500) })
	b.Run("1000_lookups", func(b *testing.B) { benchmarkViewSamplingWithLookupCount(b, 1000) })
}

func benchmarkViewSamplingWithLookupCount(b *testing.B, numLookups int) {
	req := require.New(b)

	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		req.NoError(err, "failed to set up session provider to test server")
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		req.NoError(err, "failed to set up session to test server")
	}
	defer session.Close()

	dbutils.DropDatabase(session, sampleViewDb)
	viewName := fmt.Sprintf("%v_lookups", numLookups)
	err = createBenchmarkView(session, sampleViewDb, numLookups)
	req.NoError(err, "failed to create MongoDB schema")

	// Setup sampling options.
	cfg.Schema.Sample.Namespaces = []string{fmt.Sprintf("%s.%s", sampleViewDb, viewName)}
	cfg.Schema.Sample.Size = 1
	cfg.Schema.Sample.Source = ""
	cfg.Schema.Sample.OptimizeViewSampling = true

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _, err = sample.Schema(sample.NewSchemaSampleOptions(&cfg.Schema.Sample),
			"sampling benchmark", session, lgr)
		req.NoError(err, "failed to sample schema")
	}
}

func createBenchmarkView(session *mongodb.Session, db string, numLookups int) error {
	// Insert a thousand documents.
	err := insertDocuments(session, db, 1000)
	if err != nil {
		return err
	}

	pipeline := []bson.D{}
	for i := 0; i < numLookups; i++ {
		pipeline = append(pipeline, bson.D{
			bson.DocElem{Name: "$lookup", Value: bson.D{
				bson.DocElem{Name: "from", Value: sampleViewBaseCollection},
				bson.DocElem{Name: "localField", Value: "_id"},
				bson.DocElem{Name: "foreignField", Value: "_id"},
				bson.DocElem{Name: "as", Value: fmt.Sprintf("lookup_%d", i)},
			}},
		})
	}

	// Add a blocking stage since $lookup streams documents.
	pipeline = append(pipeline, bson.D{bson.DocElem{Name: "$count", Value: "count"}})
	viewName := fmt.Sprintf("%v_lookups", numLookups)

	return createView(session, db, sampleViewBaseCollection, viewName, pipeline)
}

func createView(session *mongodb.Session, db, col, viewName string, pipeline []bson.D) error {
	cmd := bson.D{
		{Name: "create", Value: viewName},
		{Name: "viewOn", Value: col},
		{Name: "pipeline", Value: pipeline},
	}

	result := &struct {
		N  int `bson:"n"`
		Ok int `bson:"ok"`
	}{}

	if err := session.Run(db, cmd, result); err != nil {
		return fmt.Errorf("error creating view: %v", err)
	}

	if result.Ok != 1 {
		return fmt.Errorf("error executing view creation")
	}

	return nil
}

func insertDocuments(session *mongodb.Session, db string, numDocs int) error {
	insertHelper := func(documents interface{}) error {
		cmd := bson.D{
			bson.DocElem{Name: "insert", Value: sampleViewBaseCollection},
			bson.DocElem{Name: "documents", Value: documents},
			bson.DocElem{Name: "writeConcern", Value: bson.D{
				bson.DocElem{Name: "w", Value: "majority"}}},
		}

		result := &struct {
			N  int `bson:"n"`
			Ok int `bson:"ok"`
		}{}

		err := session.Run(db, cmd, result)
		if err != nil {
			return fmt.Errorf("error inserting documents: %v", err)
		}

		if result.Ok != 1 {
			return fmt.Errorf("error persisting documents: %v", err)
		}

		return nil
	}

	documents := []bson.M{}
	for i := 0; i < numDocs; i++ {
		documents = append(documents, bson.M{
			"_id":    i,
			"field1": "1",
			"field2": "2",
			"field3": "3",
			"field4": "4",
		})
	}
	return insertHelper(documents)
}
