package persist

import (
	"context"
	"fmt"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema/drdl"
)

const (
	schemasCollection = "schemas"
	namesCollection   = "names"
)

// A Persistor provides methods for storing, retrieving, and manipulating stored
// schemas.
type Persistor struct {
	schemaSourceDB  string
	sessionProvider *mongodb.SessionProvider
}

// NewPersistor creates a new Persistor with the provided configuration.
func NewPersistor(sp *mongodb.SessionProvider, schemaSourceDB string) Persistor {
	return Persistor{
		schemaSourceDB:  schemaSourceDB,
		sessionProvider: sp,
	}
}

// FindNames retrieves all stored Names.
func (p Persistor) FindNames(ctx context.Context) ([]Name, error) {
	s, err := p.session(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = s.Close() }()

	pipeline := []interface{}{}
	iter, err := s.Aggregate(ctx, p.schemaSourceDB, namesCollection, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = iter.Close(ctx) }()

	names := []Name{}
	name := &Name{}
	for iter.Next(ctx, &name) {
		names = append(names, *name)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

// SchemaInfo contains the ObjectId and creation time for a schema.
type SchemaInfo struct {
	ID      bson.ObjectId `bson:"_id"`
	Created time.Time     `bson:"created"`
}

// FindSchemas retrieves info on all stored Schemas.
func (p Persistor) FindSchemas(ctx context.Context) ([]SchemaInfo, error) {
	s, err := p.session(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = s.Close() }()

	pipeline := []interface{}{}
	iter, err := s.Aggregate(ctx, p.schemaSourceDB, schemasCollection, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = iter.Close(ctx) }()

	schemas := []SchemaInfo{}
	sch := &SchemaInfo{}
	for iter.Next(ctx, &sch) {
		schemas = append(schemas, *sch)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return schemas, nil
}

// FindSchemaByName returns the drdl.Schema corresponding to the provided name,
// if the name and its corresponding schema both exist.
func (p Persistor) FindSchemaByName(ctx context.Context, name string) (*drdl.Schema, error) {
	s, err := p.session(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = s.Close() }()

	pipeline := bsonutil.NewArray(
		bsonutil.NewD(bsonutil.NewDocElem("$lookup", bsonutil.NewD(
			bsonutil.NewDocElem("from", namesCollection),
			bsonutil.NewDocElem("localField", "_id"),
			bsonutil.NewDocElem("foreignField", "schema_id"),
			bsonutil.NewDocElem("as", "name"),
		))),
		bsonutil.NewD(bsonutil.NewDocElem("$unwind", "$name")),
		bsonutil.NewD(bsonutil.NewDocElem("$match", bsonutil.NewD(
			bsonutil.NewDocElem("name._id", name),
		))),
	)

	iter, err := s.Aggregate(ctx, p.schemaSourceDB, schemasCollection, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = iter.Close(ctx) }()

	res := struct {
		Schema *drdl.Schema `bson:"schema"`
	}{}

	_ = iter.Next(ctx, &res)
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	if res.Schema == nil {
		return nil, fmt.Errorf("no schema found for name %q", name)
	}

	return res.Schema, nil
}

// FindSchemaByID returns the drdl.Schema corresponding to the provided
// ObjectId if it exists.
func (p Persistor) FindSchemaByID(ctx context.Context, schemaID bson.ObjectId) (*drdl.Schema, error) {
	s, err := p.session(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = s.Close() }()

	pipeline := bsonutil.NewArray(
		bsonutil.NewD(
			bsonutil.NewDocElem("$match", bsonutil.NewD(
				bsonutil.NewDocElem("_id", schemaID),
			)),
		),
	)

	iter, err := s.Aggregate(ctx, p.schemaSourceDB, schemasCollection, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = iter.Close(ctx) }()

	res := struct {
		Schema *drdl.Schema `bson:"schema"`
	}{}

	_ = iter.Next(ctx, &res)
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	if res.Schema == nil {
		return nil, fmt.Errorf("no schema found with ObjectId %s", schemaID.Hex())
	}

	return res.Schema, nil
}

// DeleteSchema deletes the drdl.Schema corresponding to the provided ObjectId
// if it exists.
func (p Persistor) DeleteSchema(ctx context.Context, schemaID bson.ObjectId) error {
	s, err := p.session(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	query := bsonutil.NewD(bsonutil.NewDocElem("_id", schemaID))
	return s.Delete(ctx, p.schemaSourceDB, schemasCollection, query)
}

// DeleteName deletes the provided name if it exists.
func (p Persistor) DeleteName(ctx context.Context, name string) error {
	s, err := p.session(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	query := bsonutil.NewD(bsonutil.NewDocElem("_id", name))
	return s.Delete(ctx, p.schemaSourceDB, namesCollection, query)
}

// InsertSchema inserts the provided drdl.Schema, returning the ObjectId by
// which it can be referenced in the future.
func (p Persistor) InsertSchema(ctx context.Context, drdlSchema *drdl.Schema) (bson.ObjectId, error) {
	s, err := p.session(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = s.Close() }()

	sch := newSchema(drdlSchema)

	err = s.Insert(ctx, p.schemaSourceDB, schemasCollection, []interface{}{sch})
	if err != nil {
		return "", err
	}

	return sch.ID, nil
}

// UpsertName updates the provided name to point to the schema with the
// specified ObjectId. If the provided name does not exist, a new one is
// created instead.
func (p Persistor) UpsertName(ctx context.Context, name string, schemaID bson.ObjectId) error {
	s, err := p.session(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	query := bsonutil.NewD(bsonutil.NewDocElem("_id", name))
	update := newName(name, schemaID)

	return s.Upsert(ctx, p.schemaSourceDB, namesCollection, query, update)
}

func (p Persistor) session(ctx context.Context) (*mongodb.Session, error) {
	return p.sessionProvider.AuthenticatedAdminSessionPrimary(ctx)
}
