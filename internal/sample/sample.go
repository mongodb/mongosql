package sample

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/mongo"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/model"
)

var (
	ErrNotFound          = errors.New("sampled schema not found")
	ErrLockAcquireFailed = errors.New("lock acquisition failed")
	ErrLockReleaseFailed = errors.New("lock release failed")
)

// Collections used to perform sampling operations.
const (
	LockCollection     = "lock"
	SchemasCollection  = "schemas"
	VersionsCollection = "versions"
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

// TODO (BI-1171): AcquireLock returns nil if the database lock was
// successfully acquired. If the lock is already acquired
// by another process, it returns ErrLockAcquireFailed.
func AcquireLock(session *mongodb.Session, db string) error {
	return nil
}

// TODO (BI-1171): ReleaseLock returns nil if the database
// lock was successfully released. If the lock was acquired
// by another process, it returns ErrLockReleaseFailed.
func ReleaseLock(session *mongodb.Session, db string) error {
	return ErrLockReleaseFailed
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
			lgr.Errf(log.Always, "collection iteration close: %v", err)
		}

		if err := collectionIter.Err(); err != nil {
			lgr.Errf(log.Always, "collection iteration error: %v", err)
		}

		mappings[dbResult.Name] = nsCollections(collections)
	}

	if err := dbIter.Close(ctx); err != nil {
		lgr.Errf(log.Always, "db iteration close: %v", err)
	}

	if err := dbIter.Err(); err != nil {
		lgr.Errf(log.Always, "db iteration error: %v", err)
	}

	return mappings, nil
}

// InsertSampleRecord inserts the record - which includes version
// and namespace data - into the database specified in record.
func InsertSampleRecord(record *Record,
	session *mongodb.Session, lgr *log.Logger) error {
	if len(record.Namespaces) == 0 {
		lgr.Logf(log.Always, "No namespaces persisted to MongoDB")
		return nil
	}

	var writeConcern interface{} = "majority"

	// write concern "majority" only works for
	// replica sets and mongos-es
	if session.ClusterKind() == model.Single {
		writeConcern = 1
	}

	insertDocuments := func(collection string, documents interface{}) error {
		cmd := bson.D{
			{"insert", collection},
			{"documents", documents},
			{"writeConcern", bson.D{{"w", writeConcern}}},
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

	// first insert the namespaces into MongoDB before
	// inserting the corresponding version document.
	err := insertDocuments(SchemasCollection, record.Namespaces)
	if err != nil {
		return err
	}

	versionDocument := []interface{}{record.Version}
	err = insertDocuments(VersionsCollection, versionDocument)
	if err != nil {
		return err
	}

	lgr.Logf(log.Always, "Mapped schema persisted to MongoDB")

	return nil
}

// TODO (BI-1120): ReadSchema reads a schema stored in the configuration sampling source
// database and returns a relational representation of the schema. If no
// such schema exists, it returns ErrNotFound.
func ReadSchema(cfg *config.SchemaSampleOptions, session *mongodb.Session,
	lgr *log.Logger) (*schema.Schema, error) {
	return nil, nil
}

// SampleSchema uses the provided mongosqld configuration and session
// to sample namespaces. It returns the relational schema generated
// and the version/schemas documents resulting from sampling.
func SampleSchema(opts *config.SchemaSampleOptions, session *mongodb.Session,
	lgr *log.Logger) (*schema.Schema, *Record, error) {

	namespaces := opts.Namespaces

	nsMatcher, err := util.NewMatcher(namespaces)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid specification: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	processName := fmt.Sprintf("mongosqld-%v-%v", hostname, os.Getpid())

	lgr.Logf(log.Always, "Sampling MongoDB for schema...")

	mappings, err := FetchNamespaces(session, lgr)
	if err != nil {
		return nil, nil, err
	}

	sampleVersion := NewVersion(processName)
	sampleNamespaces := []*Namespace{}
	sampledSchema := &schema.Schema{}

	// databases that we're excluding from sampling
	dbSampleBlacklist := []string{"admin", "local", "system", opts.Source}

	sampleVersion.StartSampleTime = time.Now()

	ctx := session.Context()

	for db, collections := range mappings {
		if util.SliceContains(dbSampleBlacklist, db) {
			lgr.Logf(log.Info, "Skipping %q database", db)
			continue
		}

		sampledDb := &schema.Database{Name: db}

		for _, collection := range collections {

			ns := fmt.Sprintf("%q.%q", db, collection)

			if !nsMatcher.Has(db + "." + collection) {
				lgr.Logf(log.Info, "Skipping namespace %s", ns)
				continue
			}

			if _, ok := sampleVersion.FindDatabase(db); !ok {
				lgr.Logf(log.Info, "Mapping schema for database %q", db)
			}

			namespace := NewNamespace(db, collection, sampleVersion.Id)

			// 1. run sample command
			lgr.Logf(log.DebugLow, "Mapping schema for namespace %s", ns)
			pipeline := []bson.M{{"$sample": bson.M{"size": opts.Size}}}
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
				lgr.Errf(log.Always, "error closing iterator: %v", err)
			}

			namespace.SampleSize = count
			namespace.Schema = jsonSchema

			// 4. convert the JSON schema to a relational schema
			err = sampledDb.Map(jsonSchema, collection, false, *lgr)
			if err != nil {
				return nil, nil, fmt.Errorf("error mapping schema: %v", err)
			}

			sampleNamespaces = append(sampleNamespaces, namespace)
			sampleVersion.AddNamespace(db, collection)
			lgr.Logf(log.DebugLow, "Finished mapping schema for namespace %s", ns)
		}

		if len(sampledDb.Tables) != 0 {
			sampledSchema.Databases = append(sampledSchema.Databases, sampledDb)
		}
	}

	sampleVersion.EndSampleTime = time.Now()

	sampleData := &Record{
		Database:   opts.Source,
		Namespaces: sampleNamespaces,
		Version:    sampleVersion,
	}

	if len(sampleNamespaces) != 0 {
		lgr.Logf(log.Always, "Mapped schema for %v namespaces: %v",
			len(sampleNamespaces), sampleVersion.Databases)
	} else {
		lgr.Logf(log.Always, "No namespaces were sampled")
	}

	return sampledSchema, sampleData, nil
}
