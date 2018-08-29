package mongodrdl

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/options"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	mongodbutils "github.com/10gen/sqlproxy/internal/testutils/mongodb"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/stretchr/testify/require"
)

const (
	host = "mongodb://localhost:27017"
)

var (
	logger = log.NewComponentLogger(log.MongodrdlComponent, log.GlobalLogger())
)

func TestMongodrdl(t *testing.T) {
	t.Run("ignore_system_collections", testIgnoreSystemCollections)
	t.Run("ignore_system_collections_admin", testIgnoreSystemCollectionsAdmin)
	t.Run("view_no_geo_index", testViewNoGeoIndex)
	// this test is currently failing. it should be uncommented when BI-1552 is complete.
	// t.Run("view_geo_index", testViewGeoIndex)
	t.Run("synthetic_query_field", testSyntheticQueryField)
}

func testIgnoreSystemCollections(t *testing.T) {
	req := require.New(t)

	db := "indexed"
	opts, err := createDRDLOpts(db)
	req.NoError(err, "failed to create drdl options")

	session, err := getSession(opts)
	req.NoError(err, "failed to get MongoDB session")
	defer session.Close()
	defer dbutils.DropDatabase(session, db)
	dbutils.DropDatabase(session, db)

	documents := []bson.M{
		{
			"first":  "Who",
			"second": "What",
		},
	}
	dbutils.InsertDocuments(session, db, "test", documents)
	dbutils.CreateIndex(session, db, "test", []string{"first", "second"})

	iter, err := session.ListIndexes(db, "test")
	req.NoError(err, "failed to list indexes")

	indexes, index := []mongodb.Index{}, mongodb.Index{}
	ctx := context.Background()
	for iter.Next(ctx, &index) {
		indexes = append(indexes, index)
	}
	err = iter.Close(ctx)
	req.NoError(err, "failed to close iterator")
	req.Len(indexes, 2, "got an unexpected number of indexes")

	err = GenerateSchema(logger, opts)
	req.NoError(err, "failed to generate DRDL")

	actual, err := drdl.NewFromFile(opts.DrdlOutput.Out)
	req.NoError(err, "failed to parse generated DRDL from file")

	expected, err := drdl.NewFromFile("testdata/indexed-expected.yml")
	req.NoError(err, "failed to parse expected DRDL from file")

	actualDRDL, err := actual.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")

	expectedDRDL, err := expected.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")
	req.Equal(string(expectedDRDL), string(actualDRDL), "actual drdl yml did not match expected")
}

func testIgnoreSystemCollectionsAdmin(t *testing.T) {
	req := require.New(t)

	db := "admin"
	opts, err := createDRDLOpts(db)
	req.NoError(err, "failed to create drdl options")

	err = GenerateSchema(logger, opts)
	req.NoError(err, "failed to generate DRDL")

	actual, err := drdl.NewFromFile(opts.DrdlOutput.Out)
	req.NoError(err, "failed to parse generated DRDL from file")

	expected, err := drdl.NewFromFile("testdata/admin-expected.yml")
	req.NoError(err, "failed to parse expected DRDL from file")

	actualDRDL, err := actual.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")

	expectedDRDL, err := expected.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")
	req.Equal(string(expectedDRDL), string(actualDRDL), "actual drdl yml did not match expected")
}

func testViewNoGeoIndex(t *testing.T) {
	req := require.New(t)

	db := "viewDB"
	opts, err := createDRDLOpts(db)
	req.NoError(err, "failed to create drdl options")

	session, err := getSession(opts)
	req.NoError(err, "failed to get MongoDB session")
	defer session.Close()
	defer dbutils.DropDatabase(session, db)
	dbutils.DropDatabase(session, db)

	documents := []bson.M{
		{"a": 1, "b": 123},
		{"a": 2, "b": 134},
		{"a": 3, "b": "s"},
	}
	dbutils.InsertDocuments(session, db, "base", documents)

	err = session.Run(db, bson.D{
		{Name: "create", Value: "view"},
		{Name: "viewOn", Value: "base"},
		{Name: "pipeline", Value: []bson.M{{"$match": bson.M{"a": 3}}}},
	}, &struct{}{})
	req.NoError(err, "failed to create view")

	_, err = session.ListIndexes(db, "view")
	req.Error(err, "should not be able to list indexes on view")

	err = GenerateSchema(logger, opts)
	req.NoError(err, "failed to generate DRDL")

	actual, err := drdl.NewFromFile(opts.DrdlOutput.Out)
	req.NoError(err, "failed to parse generated DRDL from file")

	expected, err := drdl.NewFromFile("testdata/view-expected.yml")
	req.NoError(err, "failed to parse expected DRDL from file")

	actualDRDL, err := actual.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")

	expectedDRDL, err := expected.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")
	req.Equal(string(expectedDRDL), string(actualDRDL), "actual drdl yml did not match expected")
}

// until BI-1552 is completed, this function will remain unused
// nolint: unused,megacheck
func testViewGeoIndex(t *testing.T) {
	req := require.New(t)

	db := "viewGeoDB"
	opts, err := createDRDLOpts(db)
	req.NoError(err, "failed to create drdl options")

	session, err := getSession(opts)
	req.NoError(err, "failed to get MongoDB session")
	defer session.Close()
	//defer dbutils.DropDatabase(session, db)
	dbutils.DropDatabase(session, db)

	documents := []bson.M{
		{
			"loc": bson.M{
				"type":        "Point",
				"coordinates": []interface{}{-73.88, 40.78},
			},
		},
	}
	dbutils.InsertDocuments(session, db, "base", documents)
	dbutils.CreateIndex(session, db, "base", []string{"$2d:loc.coordinates"})

	iter, err := session.ListIndexes(db, "base")
	req.NoError(err, "failed to list indexes")

	ctx := context.Background()

	ok := iter.Next(ctx, &struct{}{})
	req.True(ok, "expected base to have indexes")

	err = session.Run(db, bson.D{
		{Name: "create", Value: "view"},
		{Name: "viewOn", Value: "base"},
		{Name: "pipeline", Value: []bson.M{}},
	}, &struct{}{})
	req.NoError(err, "failed to create view")

	_, err = session.ListIndexes(db, "view")
	req.Error(err, "should not be able to list indexes on view")

	err = GenerateSchema(logger, opts)
	req.NoError(err, "failed to generate DRDL")

	actual, err := drdl.NewFromFile(opts.DrdlOutput.Out)
	req.NoError(err, "failed to parse generated DRDL from file")

	expected, err := drdl.NewFromFile("testdata/view-geo-expected.yml")
	req.NoError(err, "failed to parse expected DRDL from file")

	actualDRDL, err := actual.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")

	expectedDRDL, err := expected.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")
	req.Equal(string(expectedDRDL), string(actualDRDL), "actual drdl yml did not match expected")
}

func testSyntheticQueryField(t *testing.T) {
	req := require.New(t)

	db := "syntheticQueryDB"
	opts, err := createDRDLOpts(db)
	req.NoError(err, "failed to create drdl options")
	opts.DrdlOutput.CustomFilterField = "__MONGOQUERY"

	session, err := getSession(opts)
	req.NoError(err, "failed to get MongoDB session")
	defer session.Close()
	defer dbutils.DropDatabase(session, db)
	dbutils.DropDatabase(session, db)

	documents := []bson.M{
		{
			"name":    "John Doe",
			"numbers": []interface{}{1, 2, 3},
		},
	}
	dbutils.InsertDocuments(session, db, "complete_schema", documents)

	err = GenerateSchema(logger, opts)
	req.NoError(err, "failed to generate DRDL")

	actual, err := drdl.NewFromFile(opts.DrdlOutput.Out)
	req.NoError(err, "failed to parse generated DRDL from file")

	expected, err := drdl.NewFromFile("testdata/complete_schema_synthetic-expected.yml")
	req.NoError(err, "failed to parse expected DRDL from file")

	actualDRDL, err := actual.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")

	expectedDRDL, err := expected.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")
	req.Equal(string(expectedDRDL), string(actualDRDL), "actual drdl yml did not match expected")
}

func getSSLOpts() *options.DrdlSSL {
	sslOpts := &options.DrdlSSL{}

	if len(os.Getenv(mongodbutils.SSLTestKey)) > 0 {
		return mongodbutils.DrdlTestSSLOpts()
	}

	return sslOpts
}

func createDRDLOpts(db string) (options.DrdlOptions, error) {
	opts, err := options.NewDrdlOptions()
	if err != nil {
		return *opts, err
	}

	opts.DrdlNamespace.DB = db
	opts.DrdlConnection.Host = host
	opts.DrdlSSL = getSSLOpts()
	opts.DrdlOutput.Out = fmt.Sprintf("out/%s.yml", db)
	opts.DrdlOutput.PreJoined = true
	opts.DrdlSample.Size = 1000

	return *opts, nil
}
