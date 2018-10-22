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
	b.Run("100_lookups", func(b *testing.B) { benchmarkViewSamplingWithLookupStages(b, 100) })
	b.Run("500_lookups", func(b *testing.B) { benchmarkViewSamplingWithLookupStages(b, 500) })
	b.Run("1000_lookups", func(b *testing.B) { benchmarkViewSamplingWithLookupStages(b, 1000) })
	b.Run("100_docs_with_match_and_10_lookups", func(b *testing.B) {
		benchmarkViewSamplingWithMatchAnd10LookupStages(b, 100)
	})
	b.Run("500_docs_with_match_and_10_lookups", func(b *testing.B) {
		benchmarkViewSamplingWithMatchAnd10LookupStages(b, 500)
	})
	b.Run("1000_docs_with_match_and_10_lookups", func(b *testing.B) {
		benchmarkViewSamplingWithMatchAnd10LookupStages(b, 1000)
	})
}

func benchmarkViewSampling(b *testing.B, req *require.Assertions, session *mongodb.Session, viewName string) {

	// Setup sampling options.
	cfg.Schema.Sample.Namespaces = []string{fmt.Sprintf("%s.%s", sampleViewDb, viewName)}
	cfg.Schema.Sample.Size = 1
	cfg.Schema.Sample.Source = ""
	cfg.Schema.Sample.OptimizeViewSampling = true

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _, err := sample.Schema(
			sample.NewSchemaSampleOptions(&cfg.Schema.Sample),
			"sampling benchmark",
			session,
			lgr,
		)
		req.NoError(err, "failed to sample schema")
	}
}

func benchmarkViewSamplingWithLookupStages(b *testing.B, numLookups int) {
	req := require.New(b)
	session := getSession(req)
	defer session.Close()
	dbutils.DropDatabase(session, sampleViewDb)
	viewName, err := createViewWithLookupStages(session, sampleViewDb, numLookups)
	req.NoError(err, "failed to create MongoDB schema")
	benchmarkViewSampling(b, req, session, viewName)
}

func benchmarkViewSamplingWithMatchAnd10LookupStages(b *testing.B, nDocs int) {
	req := require.New(b)
	session := getSession(req)
	defer session.Close()
	dbutils.DropDatabase(session, sampleViewDb)
	viewName, err := createViewWithMatchAndLookupStages(session, sampleViewDb, nDocs)
	req.NoError(err, "failed to create MongoDB schema")
	benchmarkViewSampling(b, req, session, viewName)
}

func createViewWithMatchAndLookupStages(s *mongodb.Session, db string, nDocs int) (string, error) {
	err := insertDocuments(s, db, nDocs)
	if err != nil {
		return "", err
	}

	// Add 10 lookup stages.
	pipeline := []bson.D{}
	for i := 0; i < 10; i++ {
		pipeline = append(pipeline, bson.D{
			bson.DocElem{Name: "$lookup", Value: bson.D{
				bson.DocElem{Name: "from", Value: sampleViewBaseCollection},
				bson.DocElem{Name: "localField", Value: "_id"},
				bson.DocElem{Name: "foreignField", Value: "_id"},
				bson.DocElem{Name: "as", Value: fmt.Sprintf("lookup_%d", i)},
			}},
		})
	}
	// Add the $match stage to simulate a cardinality altering stage.
	pipeline = append(pipeline, bson.D{
		bson.DocElem{
			Name: "$match", Value: bson.D{
				bson.DocElem{Name: "a", Value: 3},
			},
		},
	})
	viewName := fmt.Sprintf("%v_docs_with_match_10_lookups", nDocs)

	return viewName, createView(s, db, sampleViewBaseCollection, viewName, pipeline)
}

func createViewWithLookupStages(s *mongodb.Session, db string, numLookups int) (string, error) {
	// Insert a thousand documents.
	err := insertDocuments(s, db, 1000)
	if err != nil {
		return "", err
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

	return viewName, createView(s, db, sampleViewBaseCollection, viewName, pipeline)
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

func getSession(req *require.Assertions) *mongodb.Session {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		req.NoError(err, "failed to set up session provider to test server")
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		req.NoError(err, "failed to set up session to test server")
	}
	return session
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
