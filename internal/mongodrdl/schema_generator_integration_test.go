//+build integration

package mongodrdl

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/testutil/dbutils"
	"github.com/10gen/sqlproxy/internal/testutil/flags"
	mongodbutils "github.com/10gen/sqlproxy/internal/testutil/mongodb"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	t.Run("view_geo_index", testViewGeoIndex)
	t.Run("synthetic_query_field", testSyntheticQueryField)
	t.Run("polymorphic_data_field", testPolymorphicDataField)
	t.Run("uuid_subtype3_data_field", testUUIDSubtype3Field)
}

func testIgnoreSystemCollections(t *testing.T) {
	req := require.New(t)

	db := "indexed"
	opts, err := createDRDLOpts(db)
	req.NoError(err, "failed to create drdl options")

	ctx := context.Background()
	sp, err := newDrdlSessionProvider(opts)
	req.NoError(err, "failed to create SessionProvider")
	defer sp.Close()

	session, err := sp.Session(ctx)
	req.NoError(err, "failed to get MongoDB session")
	defer session.Close()
	defer dbutils.DropDatabase(session, db)
	dbutils.DropDatabase(session, db)

	documents := bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("first", "Who"),
			bsonutil.NewDocElem("second", "What"),
		),
	)

	dbutils.InsertDocuments(session, db, "test", documents)
	dbutils.CreateIndex(session, db, "test", []string{"first", "second"})

	cursor, err := session.ListIndexes(ctx, db, "test")
	req.NoError(err, "failed to list indexes")

	indexes, index := []mongodb.Index{}, mongodb.Index{}
	for cursor.Next(ctx, &index) {
		indexes = append(indexes, index)
		index.Key = bsonutil.NewD()
	}

	req.NoError(cursor.Err(), "unexpected cursor error")

	err = cursor.Close(ctx)
	req.NoError(err, "failed to close iterator")
	req.Len(indexes, 2, "got an unexpected number of indexes")

	err = GenerateSchema(ctx, logger, opts)
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

	err = GenerateSchema(context.Background(), logger, opts)
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

	ctx := context.Background()
	sp, err := newDrdlSessionProvider(opts)
	req.NoError(err, "failed to create SessionProvider")
	defer sp.Close()

	session, err := sp.Session(ctx)
	req.NoError(err, "failed to get MongoDB session")
	defer session.Close()
	defer dbutils.DropDatabase(session, db)
	dbutils.DropDatabase(session, db)

	documents := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 1), bsonutil.NewDocElem("b", 123)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 2), bsonutil.NewDocElem("b", 134)),
		bsonutil.NewD(bsonutil.NewDocElem("a", 3), bsonutil.NewDocElem("b", "s")),
	)

	dbutils.InsertDocuments(session, db, "base", documents)

	err = session.Run(ctx, db, bsonutil.NewD(
		bsonutil.NewDocElem("create", "view"),
		bsonutil.NewDocElem("viewOn", "base"),
		bsonutil.NewDocElem("pipeline", bsonutil.NewDArray(
			bsonutil.NewD(
				bsonutil.NewDocElem(
					"$match",
					bsonutil.NewD(bsonutil.NewDocElem("a", 3)),
				),
			),
		)),
	), &struct{}{})
	req.NoError(err, "failed to create view")

	_, err = session.ListIndexes(ctx, db, "view")
	req.Error(err, "should not be able to list indexes on view")

	err = GenerateSchema(ctx, logger, opts)
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

	ctx := context.Background()
	sp, err := newDrdlSessionProvider(opts)
	req.NoError(err, "failed to create SessionProvider")
	defer sp.Close()

	session, err := sp.Session(ctx)
	req.NoError(err, "failed to get MongoDB session")
	defer session.Close()
	defer dbutils.DropDatabase(session, db)
	dbutils.DropDatabase(session, db)

	documents := bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("loc", bsonutil.NewD(
				bsonutil.NewDocElem("type", "Point"),
				bsonutil.NewDocElem("coordinates", bsonutil.NewArray(-73.88, 40.78)),
			)),
		),
	)

	dbutils.InsertDocuments(session, db, "base", documents)
	dbutils.CreateIndex(session, db, "base", []string{"$2d:loc.coordinates"})

	cursor, err := session.ListIndexes(ctx, db, "base")
	req.NoError(err, "failed to list indexes")

	ok := cursor.Next(ctx, &struct{}{})
	req.True(ok, "expected base to have indexes")

	err = session.Run(ctx, db, bsonutil.NewD(
		bsonutil.NewDocElem("create", "view"),
		bsonutil.NewDocElem("viewOn", "base"),
		bsonutil.NewDocElem("pipeline", bsonutil.NewDArray()),
	), &struct{}{})
	req.NoError(err, "failed to create view")

	_, err = session.ListIndexes(ctx, db, "view")
	req.Error(err, "should not be able to list indexes on view")

	err = GenerateSchema(ctx, logger, opts)
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

	ctx := context.Background()
	sp, err := newDrdlSessionProvider(opts)
	req.NoError(err, "failed to create SessionProvider")
	defer sp.Close()

	session, err := sp.Session(ctx)
	req.NoError(err, "failed to get MongoDB session")
	defer session.Close()
	defer dbutils.DropDatabase(session, db)
	dbutils.DropDatabase(session, db)

	documents := bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("name", "John Doe"),
			bsonutil.NewDocElem("numbers", bsonutil.NewArray(1, 2, 3)),
		),
	)

	dbutils.InsertDocuments(session, db, "complete_schema", documents)

	err = GenerateSchema(ctx, logger, opts)
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

func testPolymorphicDataField(t *testing.T) {
	req := require.New(t)

	db := "polymorphicDataDb"
	opts, err := createDRDLOpts(db)
	req.NoError(err, "failed to create drdl options")

	ctx := context.Background()
	sp, err := newDrdlSessionProvider(opts)
	req.NoError(err, "failed to create SessionProvider")
	defer sp.Close()

	session, err := sp.Session(ctx)
	req.NoError(err, "failed to get MongoDB session")
	defer session.Close()
	defer dbutils.DropDatabase(session, db)
	dbutils.DropDatabase(session, db)

	documents := bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("name", "John Doe"),
			bsonutil.NewDocElem("payload", "hello"),
		),
		bsonutil.NewD(
			bsonutil.NewDocElem("name", "John Doe"),
			bsonutil.NewDocElem("payload", bsonutil.NewD(
				bsonutil.NewDocElem("subdoc1", 4),
				bsonutil.NewDocElem("subdoc2", 4),
			)),
		),
	)

	dbutils.InsertDocuments(session, db, "polymorphic_data_schema", documents)

	err = GenerateSchema(ctx, logger, opts)
	req.NoError(err, "failed to generate DRDL")

	actual, err := drdl.NewFromFile(opts.DrdlOutput.Out)
	req.NoError(err, "failed to parse generated DRDL from file")

	expected, err := drdl.NewFromFile("testdata/polymorphic_data_schema-expected.yml")
	req.NoError(err, "failed to parse expected DRDL from file")

	actualDRDL, err := actual.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")

	expectedDRDL, err := expected.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")
	req.Equal(string(expectedDRDL), string(actualDRDL), "actual drdl yml did not match expected")
}

func testUUIDSubtype3Field(t *testing.T) {
	req := require.New(t)

	db := "UUIDSubtype3Field"
	opts, err := createDRDLOpts(db)
	req.NoError(err, "failed to create drdl options")

	ctx := context.Background()
	sp, err := newDrdlSessionProvider(opts)
	req.NoError(err, "failed to create SessionProvider")
	defer sp.Close()

	session, err := sp.Session(ctx)
	req.NoError(err, "failed to get MongoDB session")
	defer session.Close()
	defer dbutils.DropDatabase(session, db)
	dbutils.DropDatabase(session, db)

	documents := bsonutil.NewDArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("name", primitive.Binary{Subtype: 0x03, Data: []byte("amOjUW1oQQ6dNsvLrQuDhg==")}),
		),
	)

	dbutils.InsertDocuments(session, db, "uuid_subtype3_schema", documents)

	err = GenerateSchema(ctx, logger, opts)
	req.NoError(err, "failed to generate DRDL")

	actual, err := drdl.NewFromFile(opts.DrdlOutput.Out)
	req.NoError(err, "failed to parse generated DRDL from file")

	expected, err := drdl.NewFromFile("testdata/uuid_subtype3_schema-expected.yml")
	req.NoError(err, "failed to parse expected DRDL from file")

	actualDRDL, err := actual.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")

	expectedDRDL, err := expected.ToYAML()
	req.NoError(err, "failed to get yaml output for drdl")
	req.Equal(string(expectedDRDL), string(actualDRDL), "actual drdl yml did not match expected")
}

func getSSLOpts() *DrdlSSL {
	sslOpts := &DrdlSSL{}

	if len(os.Getenv(mongodbutils.SSLTestKey)) > 0 {
		return drdlTestSSLOpts()
	}

	return sslOpts
}

func createDRDLOpts(db string) (DrdlOptions, error) {
	opts, err := NewDrdlOptions()
	if err != nil {
		return *opts, err
	}

	args := []string{
		"--db", db,
		"--host", host,
	}

	err = opts.Parse(args)
	if err != nil {
		return *opts, err
	}

	opts.DrdlSSL = getSSLOpts()
	opts.DrdlOutput.Out = fmt.Sprintf("out/%s.yml", db)
	opts.DrdlOutput.PreJoined = true
	opts.DrdlOutput.UUIDSubtype3Encoding = "old"
	opts.DrdlSample.Size = 1000

	return *opts, nil
}

// drdlTestSSLOpts returns the mongodrdl SSL options to use for testing.
func drdlTestSSLOpts() *DrdlSSL {
	return &DrdlSSL{
		Enabled:             true,
		SSLPEMKeyFile:       fmt.Sprintf("../../%v", *flags.ClientPEMKeyFile),
		SSLAllowInvalidCert: true,
		MinimumTLSVersion:   "TLS1_1",
	}
}
