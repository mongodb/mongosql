//+build integration

package persist_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/testutil/dbutils"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/10gen/sqlproxy/schema/persist"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	testDBName = "persist_test"
)

func TestPersistSchema(t *testing.T) {
	req := require.New(t)
	ctx := context.Background()

	sp := getSessionProvider()
	defer sp.Close()

	p := persist.NewPersistor(sp, testDBName)

	drdlSchema := new(drdl.Schema)
	drdlErr := drdlSchema.LoadFile("testdata/schema.drdl")
	req.NoError(drdlErr, "failed to load DRDL schema from file")

	t.Run("roundtrip schema", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		sid, err := p.InsertSchema(ctx, drdlSchema)
		req.NoError(err, "failed to insert schema")

		foundSchema, err := p.FindSchemaByID(ctx, sid)
		req.NoError(err, "failed to find previously-inserted schema by id")
		schemaEquals(req, drdlSchema, foundSchema)
	})

	t.Run("create named schema", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		sid, err := p.InsertSchema(ctx, drdlSchema)
		req.NoError(err, "failed to insert schema")

		err = p.UpsertName(ctx, "defaultSchema", sid)
		req.NoError(err, "failed to insert name")

		foundSchema, err := p.FindSchemaByName(ctx, "defaultSchema")
		req.NoError(err, "failed to find previously-inserted schema by name")
		schemaEquals(req, drdlSchema, foundSchema)
	})

	t.Run("create named schema with no pipeline in drdl", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		drdlSch := new(drdl.Schema)
		drdlErr := drdlSch.LoadFile("testdata/schema_no_pipeline.drdl")
		req.NoError(drdlErr, "failed to load DRDL schema from file")

		sid, err := p.InsertSchema(ctx, drdlSch)
		req.NoError(err, "failed to insert schema")

		err = p.UpsertName(ctx, "defaultSchema", sid)
		req.NoError(err, "failed to insert name")

		foundSchema, err := p.FindSchemaByName(ctx, "defaultSchema")
		req.NoError(err, "failed to find previously-inserted schema by name")
		schemaEquals(req, drdlSch, foundSchema)
	})

	t.Run("create name referencing nonexistent schema", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		err := p.UpsertName(ctx, "defaultSchema", primitive.NewObjectID())
		req.NoError(err, "failed to insert name")
	})

	t.Run("find schemas empty collection", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		schemas, err := p.FindSchemas(ctx)
		req.NoError(err, "failed to fetch schemas")
		req.Empty(schemas)
	})

	t.Run("find single schema", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		sid, err := p.InsertSchema(ctx, drdlSchema)
		req.NoError(err, "failed to insert schema")

		schemas, err := p.FindSchemas(ctx)
		req.NoError(err, "failed to fetch schemas")
		req.Len(schemas, 1)
		req.Equal(sid, schemas[0].ID)
		req.WithinDuration(time.Now(), schemas[0].Created, 10*time.Second)
	})

	t.Run("find names empty collection", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		names, err := p.FindNames(ctx)
		req.NoError(err, "failed to fetch names")
		req.Empty(names)
	})

	t.Run("find single name", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		err := p.UpsertName(ctx, "nameOne", primitive.NewObjectID())
		req.NoError(err, "failed to insert name")

		names, err := p.FindNames(ctx)
		req.NoError(err, "failed to fetch names")
		req.Len(names, 1)
		req.Equal("nameOne", names[0].ID)
	})

	t.Run("find schema by nonexistent name", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		_, err := p.FindSchemaByName(ctx, "defaultSchema")
		req.EqualError(err, `no schema found for name "defaultSchema"`)
	})

	t.Run("find schema by name with invalid schema id", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		err := p.UpsertName(ctx, "defaultSchema", primitive.NewObjectID())
		req.NoError(err, "failed to insert name")

		_, err = p.FindSchemaByName(ctx, "defaultSchema")
		req.EqualError(err, `no schema found for name "defaultSchema"`)
	})

	t.Run("find schema by nonexistent id", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		oid := primitive.NewObjectID()
		_, err := p.FindSchemaByID(ctx, oid)
		req.EqualError(err, fmt.Sprintf("no schema found with ObjectId %s", oid.Hex()))
	})

	t.Run("delete schema", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		firstID, err := p.InsertSchema(ctx, drdlSchema)
		req.NoError(err, "failed to insert first schema")

		secondID, err := p.InsertSchema(ctx, drdlSchema)
		req.NoError(err, "failed to insert second schema")

		err = p.DeleteSchema(ctx, firstID)
		req.NoError(err, "failed to delete first schema")

		_, err = p.FindSchemaByID(ctx, secondID)
		req.NoError(err, "failed to find second schema")

		_, err = p.FindSchemaByID(ctx, firstID)
		req.EqualError(err, fmt.Sprintf("no schema found with ObjectId %s", firstID.Hex()))
	})

	t.Run("delete nonexistent schema", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		sid, err := p.InsertSchema(ctx, drdlSchema)
		req.NoError(err, "failed to insert schema")

		err = p.DeleteSchema(ctx, primitive.NewObjectID())
		req.NoError(err, "deleting nonexistent schema should return no error")

		_, err = p.FindSchemaByID(ctx, sid)
		req.NoError(err, "failed to find schema")
	})

	t.Run("delete name", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		err := p.UpsertName(ctx, "nameOne", primitive.NewObjectID())
		req.NoError(err, "failed to insert name one")

		err = p.UpsertName(ctx, "nameTwo", primitive.NewObjectID())
		req.NoError(err, "failed to insert name two")

		err = p.DeleteName(ctx, "nameTwo")
		req.NoError(err, "failed to delete name two")

		names, err := p.FindNames(ctx)
		req.NoError(err, "failed to fetch names")
		req.Len(names, 1)
		req.Equal("nameOne", names[0].ID)
	})

	t.Run("delete nonexistent name", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		err := p.UpsertName(ctx, "defaultSchema", primitive.NewObjectID())
		req.NoError(err, "failed to insert name")

		err = p.DeleteName(ctx, "abc")
		req.NoError(err, "deleting nonexistent name should return no error")

		names, err := p.FindNames(ctx)
		req.NoError(err, "failed to fetch names")
		req.Len(names, 1)
		req.Equal("defaultSchema", names[0].ID)
	})

	t.Run("insert duplicate schema", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		firstID, err := p.InsertSchema(ctx, drdlSchema)
		req.NoError(err, "failed to insert first schema")

		secondID, err := p.InsertSchema(ctx, drdlSchema)
		req.NoError(err, "failed to insert second schema")

		req.NotEqual(secondID, firstID, "expected two separate schemas to be created")

		firstSchema, err := p.FindSchemaByID(ctx, firstID)
		req.NoError(err, "failed to find first schema")

		secondSchema, err := p.FindSchemaByID(ctx, secondID)
		req.NoError(err, "failed to find second schema")

		schemaEquals(req, firstSchema, secondSchema)
	})

	t.Run("update name", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		oid := primitive.NewObjectID()
		err := p.UpsertName(ctx, "defaultSchema", oid)
		req.NoError(err, "failed to insert name")

		newOID := primitive.NewObjectID()
		err = p.UpsertName(ctx, "defaultSchema", newOID)
		req.NoError(err, "failed to update name")

		names, err := p.FindNames(ctx)
		req.NoError(err, "failed to fetch names")
		req.Len(names, 1)
		req.Equal("defaultSchema", names[0].ID)
		req.Equal(newOID, names[0].SchemaID)
	})

	t.Run("multiple names for schema", func(t *testing.T) {
		setup(sp)
		req := require.New(t)

		sid, err := p.InsertSchema(ctx, drdlSchema)
		req.NoError(err, "failed to insert schema")

		err = p.UpsertName(ctx, "nameOne", sid)
		req.NoError(err, "failed to insert name one")

		err = p.UpsertName(ctx, "nameTwo", sid)
		req.NoError(err, "failed to insert name two")

		schemaOne, err := p.FindSchemaByName(ctx, "nameOne")
		req.NoError(err, "failed to fetch schema one")

		schemaTwo, err := p.FindSchemaByName(ctx, "nameTwo")
		req.NoError(err, "failed to fetch schema two")

		schemaEquals(req, schemaOne, schemaTwo)
	})

}

func setup(sp *mongodb.SessionProvider) {
	s, err := sp.Session(context.Background())
	if err != nil {
		panic(err)
	}
	dbutils.DropDatabase(s, testDBName)
	err = s.Close()
	if err != nil {
		panic(err)
	}
}

func schemaEquals(req *require.Assertions, expected, actual *drdl.Schema) {
	exBytes, err := expected.ToYAML()
	req.NoError(err, "failed to marshal expected schema to yaml")
	exStr := string(exBytes)

	actBytes, err := actual.ToYAML()
	req.NoError(err, "failed to marshal actual schema to yaml")
	actStr := string(actBytes)

	req.Equal(exStr, actStr, "schemas were not equal")
}

func getSessionProvider() *mongodb.SessionProvider {
	sp, err := mongodb.NewSqldSessionProvider(config.Default())
	if err != nil {
		panic(err)
	}
	return sp
}
