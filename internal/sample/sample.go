package sample

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/mongo"
)

// Collections used to perform sampling operations.
const (
	LockCollection         = "mongosqld.lock"
	SchemasCollection      = "mongosqld.schemas"
	VersionsCollection     = "mongosqld.versions"
	VersionIdField         = "versionId"
	VersionGenerationField = "generation"
)

// Record holds data generated as a result
// of performing sampling operations.
type Record struct {
	Database   string
	Version    *Version
	Namespaces []*Namespace
}

// nsMapping is a mapping of database names to nsCollections.
type nsMapping map[string]nsCollections

// nsCollections is a list of collection names.
type nsCollections []string

// Alter changes a Record to represent a new Version of the current record with
// the provided alterations applied.
func (r *Record) Alter(alts []*schema.Alteration) {
	id := bson.NewObjectId()
	r.Version.Id = id
	r.Version.Generation += 1
	r.Version.Alterations = append(r.Version.Alterations, alts...)
	r.Version.Protocol = CurrentProtocol
	for _, ns := range r.Namespaces {
		ns.id = bson.NewObjectId()
		ns.VersionId = id
	}
}

func (r *Record) getSchema(cfg *config.SchemaSampleOptions, lgr *log.Logger) (*schema.Schema, error) {
	sampledSchema := &schema.Schema{
		Alterations: r.Version.Alterations,
	}

	var lastDB string
	var sampledDB *schema.Database

	for _, ns := range r.Namespaces {
		if lastDB != ns.Database {
			lastDB = ns.Database

			sampledDB = &schema.Database{
				Name: ns.Database,
			}
			sampledSchema.Databases = append(sampledSchema.Databases, sampledDB)
		}

		err := sampledDB.Map(ns.Schema, ns.Collection, false, cfg.UUIDSubtype3Encoding, *lgr)
		if err != nil {
			return nil, fmt.Errorf(
				"error mapping schema version %s, namespace %q.%q: %v",
				r.Version, ns.Database, ns.Collection, err,
			)
		}
	}

	return sampledSchema, nil
}

func (r *Record) validateNamespaceCount() error {
	expected := 0
	for _, db := range r.Version.Databases {
		expected += len(db.Collections)
	}

	if len(r.Namespaces) != expected {
		return fmt.Errorf("expected %d namespaces, got %d", expected, len(r.Namespaces))
	}

	return nil
}

// FetchNamespaces returns a map of databases - that
// exist in the cluster 'session' is connected to - to
// the collection(s) within the database.
func FetchNamespaces(session *mongodb.Session, lgr *log.Logger) (nsMapping, error) {
	mappings := map[string]nsCollections{}

	ctx := session.Context()

	dbIter, err := session.ListDatabases()
	if err != nil {
		return nil, fmt.Errorf("error listing databases: %v", err)
	}

	var dbResult struct {
		Name string `bson:"name"`
	}

	for dbIter.Next(ctx, &dbResult) {
		collectionIter, err := session.ListCollections(dbResult.Name)
		if err != nil {
			return nil, fmt.Errorf("can't get the collection "+
				"names for '%v': %v", dbResult.Name, err)
		}

		var collectionResult struct {
			Name string `bson:"name"`
		}

		ctx := session.Context()

		collections := []string{}

		for collectionIter.Next(ctx, &collectionResult) {
			collections = append(collections, collectionResult.Name)
		}

		if err := collectionIter.Close(ctx); err != nil {
			lgr.Warnf(log.Dev, "error closing collection iterator: %v", err)
		}

		if err := collectionIter.Err(); err != nil {
			lgr.Warnf(log.Dev, "collection iteration error: %v", err)
		}

		mappings[dbResult.Name] = nsCollections(collections)
	}

	if err := dbIter.Close(ctx); err != nil {
		lgr.Warnf(log.Dev, "error closing db iterator: %v", err)
	}

	if err := dbIter.Err(); err != nil {
		lgr.Warnf(log.Dev, "db iteration error: %v", err)
	}

	return mappings, nil
}

// InsertSampleRecord inserts the record - which includes version
// and namespace data - into the database specified in record.
func InsertSampleRecord(record *Record,
	session *mongodb.Session, lgr *log.Logger) error {
	if len(record.Namespaces) == 0 {
		lgr.Debugf(log.Admin, "no namespaces in sample: not persisting to MongoDB")
		return nil
	}

	insertDocuments := func(collection string, documents interface{}) error {
		cmd := bson.D{
			{"insert", collection},
			{"documents", documents},
			{"writeConcern", bson.D{{"w", "majority"}}},
		}

		result := &struct {
			N  int `bson:"n"`
			Ok int `bson:"ok"`
		}{}

		err := session.Run(record.Database, cmd, result)
		if err != nil {
			return fmt.Errorf("error inserting schema: %v", err)
		}

		if result.Ok != 1 {
			return fmt.Errorf("error persisting schema: %v", err)
		}

		return nil
	}

	// 1. insert the namespaces into MongoDB before
	// inserting the corresponding version document.
	err := insertDocuments(SchemasCollection, record.Namespaces)
	if err != nil {
		return err
	}

	// 2. insert the version document
	versionDocument := []interface{}{record.Version}
	err = insertDocuments(VersionsCollection, versionDocument)
	if err != nil {
		return err
	}

	lgr.Infof(log.Admin, "schema persisted to MongoDB")

	return nil
}

// ReadSchema reads a schema stored in the configuration sampling source
// database and returns a relational representation of the schema. If no
// such schema exists, it returns nil.
func ReadSchema(cfg *config.SchemaSampleOptions, session *mongodb.Session,
	lgr *log.Logger) (*schema.Schema, error) {

	// get the latest stored schema record
	lgr.Infof(log.Admin, "retrieving latest schema")
	rec, err := LatestRecord(cfg, session, lgr)
	if rec == nil || err != nil {
		return nil, err
	}

	// get the schema from the record
	versionedSchema, err := rec.getSchema(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return versionedSchema, nil
}

// getLatestVersion returns a schema (ObjectID) if a version exists; otherwise it returns nil. In addition, it returns
// the number of namespaces that are present in this version. If error is not nil, an error occurred.
func getLatestVersion(cfg *config.SchemaSampleOptions, session *mongodb.Session) (*bson.ObjectId, []*schema.Alteration, int, error) {
	var pipeline interface{} = []bson.D{
		{{"$sort", bson.D{{"generation", -1}}}},
		{{"$limit", 1}},
		{{"$project", bson.D{
			{"_id", 1},
			{"alterations", 1},
			{"namespaceCount", bson.D{
				{"$sum", bson.D{
					{"$map", bson.D{
						{"input", "$databases"},
						{"as", "db"},
						{"in", bson.D{
							{"$size", "$$db.collections"},
						}},
					}},
				}},
			}},
		}}},
	}

	cursor, err := session.Aggregate(cfg.Source, VersionsCollection, pipeline)
	if err != nil {
		return nil, nil, 0, err
	}
	defer cursor.Close(session.Context())

	result := struct {
		ID             *bson.ObjectId       `bson:"_id"`
		Alterations    []*schema.Alteration `bson:"alterations"`
		NamespaceCount int                  `bson:"namespaceCount"`
	}{}

	if cursor.Next(session.Context(), &result) {
		return result.ID, result.Alterations, result.NamespaceCount, nil
	}

	return nil, nil, 0, cursor.Err()
}

// SampleSchema uses the provided mongosqld configuration and session
// to sample namespaces. It returns the relational schema generated
// and the version/schemas documents resulting from sampling.
func SampleSchema(cfg *config.SchemaSampleOptions, processName string,
	session *mongodb.Session, lgr *log.Logger) (*schema.Schema, *Record, error) {

	namespaces := cfg.Namespaces

	nsMatcher, err := util.NewMatcher(namespaces)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid specification: %v", err)
	}

	lgr.Infof(log.Always, "sampling MongoDB for schema...")

	mappings, err := FetchNamespaces(session, lgr)
	if err != nil {
		return nil, nil, err
	}

	sampleVersion := NewVersion(processName)
	sampleNamespaces := []*Namespace{}
	sampledSchema := &schema.Schema{}
	uuidSubtype3Encoding := cfg.UUIDSubtype3Encoding

	// databases that we're excluding from sampling
	dbSampleBlacklist := []string{"admin", "local", "system"}
	nsSampleBlacklist := []string{
		fmt.Sprintf("%q.%q", cfg.Source, SchemasCollection),
		fmt.Sprintf("%q.%q", cfg.Source, VersionsCollection),
		fmt.Sprintf("%q.%q", cfg.Source, LockCollection),
	}

	sampleVersion.StartSampleTime = time.Now()

	ctx := session.Context()

	for db, collections := range mappings {
		if util.SliceContains(dbSampleBlacklist, db) {
			lgr.Debugf(log.Dev, "skipping %q database", db)
			continue
		}

		sampledDB := &schema.Database{Name: db}

		for _, collection := range collections {

			ns := fmt.Sprintf("%q.%q", db, collection)

			if !nsMatcher.Has(db+"."+collection) ||
				util.SliceContains(nsSampleBlacklist, ns) ||
				strings.HasPrefix(collection, "system.") {
				lgr.Debugf(log.Dev, "skipping namespace %s", ns)
				continue
			}

			if _, ok := sampleVersion.FindDatabase(db); !ok {
				lgr.Debugf(log.Dev, "mapping schema for database %q", db)
			}

			namespace := NewNamespace(db, collection, sampleVersion.Id)

			// 1. run sample command
			lgr.Debugf(log.Dev, "mapping schema for namespace %s", ns)
			pipeline := []bson.M{}
			if cfg.Size != 0 {
				pipeline = append(pipeline, bson.M{"$sample": bson.M{"size": cfg.Size}})
			}
			namespace.StartSampleTime = time.Now()

			// 2. get sample documents
			iter, err := session.Aggregate(db, collection, pipeline)
			if err != nil {
				return nil, nil, fmt.Errorf("error sampling collection: %v", err)
			}

			namespace.EndSampleTime = time.Now()

			jsonSchema := mongo.NewCollectionSchema()

			// 3. create json schema and store it
			count, doc := int64(0), &bson.D{}

			for iter.Next(ctx, doc) {
				err = jsonSchema.IncludeSample(*doc)
				if err != nil {
					return nil, nil, fmt.Errorf("error including collection: %v", err)
				}
				doc = &bson.D{}
				count += 1
			}

			if err := iter.Close(ctx); err != nil {
				lgr.Warnf(log.Dev, "error closing iterator: %v", err)
			}

			namespace.SampleSize = count
			namespace.Schema = jsonSchema

			// 4. convert the JSON schema to a relational schema
			err = sampledDB.Map(
				jsonSchema,
				collection,
				false,
				uuidSubtype3Encoding,
				*lgr,
			)

			if err != nil {
				return nil, nil, fmt.Errorf("error mapping schema: %v", err)
			}

			sampleNamespaces = append(sampleNamespaces, namespace)
			sampleVersion.AddNamespace(db, collection)
			lgr.Debugf(log.Dev, "finished mapping schema for namespace %s", ns)
		}

		if len(sampledDB.Tables) != 0 {
			sampledSchema.Databases = append(sampledSchema.Databases, sampledDB)
		}
	}

	sampleVersion.EndSampleTime = time.Now()

	sampleData := &Record{
		Database:   cfg.Source,
		Namespaces: sampleNamespaces,
		Version:    sampleVersion,
	}

	if len(sampleNamespaces) != 0 {
		lgr.Infof(log.Always, "mapped schema for %v namespaces: %v",
			len(sampleNamespaces), sampleVersion.Databases)
	} else {
		lgr.Infof(log.Always, "no namespaces were sampled")
	}

	return sampledSchema, sampleData, nil
}

// LatestGeneration returns the most recent generation of the schema stored in MongoDB
func LatestGeneration(opts *config.SchemaSampleOptions, session *mongodb.Session, lgr *log.Logger) (int64, error) {
	rec, err := LatestRecord(opts, session, lgr)
	if err != nil {
		return -1, err
	}

	if rec.Version == nil {
		return -1, nil
	}

	return rec.Version.Generation, nil
}

// LatestRecord returns a Record representing the most recent generation of the
// schema stored in MongoDB. If there is no schema currently stored in MongoDB,
// LatestRecord returns a nil Record.
func LatestRecord(opts *config.SchemaSampleOptions, session *mongodb.Session, lgr *log.Logger) (*Record, error) {
	var pipeline interface{} = []bson.D{
		{{"$sort", bson.D{{"generation", -1}}}},
		{{"$limit", 1}},
		{{"$project", bson.D{
			{"_id", 0},
			{"version", "$$CURRENT"},
		}}},
		{{"$lookup", bson.D{
			{"from", SchemasCollection},
			{"localField", "version._id"},
			{"foreignField", VersionIdField},
			{"as", "namespaces"},
		}}},
	}

	cursor, err := session.Aggregate(opts.Source, VersionsCollection, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	rec := &Record{}
	if cursor.Next(session.Context(), rec) {
		rec.Database = opts.Source
		err := rec.validateNamespaceCount()
		if err != nil {
			return nil, err
		}
		return rec, cursor.Err()
	}

	return nil, cursor.Err()
}
