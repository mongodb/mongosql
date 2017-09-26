package sample

import (
	"context"
	"fmt"
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
		lgr.Logf(log.Always, "No namespaces in sample: not persisting to MongoDB")
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

	lgr.Logf(log.Always, "Mapped schema persisted to MongoDB")

	return nil
}

// ReadSchema reads a schema stored in the configuration sampling source
// database and returns a relational representation of the schema. If no
// such schema exists, it returns nil.
func ReadSchema(cfg *config.SchemaSampleOptions, session *mongodb.Session,
	lgr *log.Logger) (*schema.Schema, error) {

	// 1. Find the latest version. Version can be null even if err is also null if
	// there is no version stored.
	version, expectedNamespaceCount, err := getLatestVersion(cfg, session)
	if version == nil || err != nil {
		return nil, err
	}

	// 2. Get the schema for the latest version
	lgr.Logf(log.Info, "retrieving latest schema version %s", *version)
	versionedSchema, namespaceCount, err := getSchemaByVersion(*version, cfg, session, lgr)
	if err != nil {
		return nil, err
	}

	if namespaceCount != expectedNamespaceCount {
		return nil, fmt.Errorf("schema version %s should contain %d namespaces, but found %d", version, expectedNamespaceCount, namespaceCount)
	}

	return versionedSchema, nil
}

// getLatestVersion returns a schema (ObjectID) if a version exists; otherwise it returns nil. In addition, it returns
// the number of namespaces that are present in this version. If error is not nil, an error occurred.
func getLatestVersion(cfg *config.SchemaSampleOptions, session *mongodb.Session) (*bson.ObjectId, int, error) {
	var pipeline interface{} = []bson.D{
		{{"$sort", bson.D{{"generation", -1}}}},
		{{"$limit", 1}},
		{{"$project", bson.D{
			{"_id", 1},
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
		return nil, 0, err
	}
	defer cursor.Close(session.Context())

	result := struct {
		ID             *bson.ObjectId `bson:"_id"`
		NamespaceCount int            `bson:"namespaceCount"`
	}{}

	if cursor.Next(session.Context(), &result) {
		return result.ID, result.NamespaceCount, nil
	}

	return nil, 0, cursor.Err()
}

// getSchemaByVersion returns a schema (ObjectID) if a version exists; otherwise it returns nil. In addition, it returns
// the number of namespaces that were retrieved. If error is not nil, an error occurred.
func getSchemaByVersion(version bson.ObjectId, cfg *config.SchemaSampleOptions, session *mongodb.Session, lgr *log.Logger) (*schema.Schema, int, error) {
	var pipeline interface{} = []bson.D{
		{{"$match", bson.D{{"versionId", version}}}},
		{{"$sort", bson.D{{"database", 1}, {"collection", 1}}}},
	}

	cursor, err := session.Aggregate(cfg.Source, SchemasCollection, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.Background())

	sampledSchema := &schema.Schema{}
	namespaceCount := 0
	var ns Namespace
	var lastDB string
	var sampledDB *schema.Database
	for cursor.Next(session.Context(), &ns) {
		namespaceCount++
		if lastDB != ns.Database {
			lastDB = ns.Database

			sampledDB = &schema.Database{
				Name: ns.Database,
			}
			sampledSchema.Databases = append(sampledSchema.Databases, sampledDB)
		}

		err = sampledDB.Map(ns.Schema, ns.Collection, false, *lgr)
		if err != nil {
			return nil, 0, fmt.Errorf("error mapping schema version %s, namespace %q.%q: %v", version, ns.Database, ns.Collection, err)
		}
	}

	return sampledSchema, namespaceCount, cursor.Err()
}

// SampleSchema uses the provided mongosqld configuration and session
// to sample namespaces. It returns the relational schema generated
// and the version/schemas documents resulting from sampling.
func SampleSchema(opts *config.SchemaSampleOptions, processName string,
	session *mongodb.Session, lgr *log.Logger) (*schema.Schema, *Record, error) {

	namespaces := opts.Namespaces

	nsMatcher, err := util.NewMatcher(namespaces)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid specification: %v", err)
	}

	lgr.Logf(log.Always, "Sampling MongoDB for schema...")

	mappings, err := FetchNamespaces(session, lgr)
	if err != nil {
		return nil, nil, err
	}

	sampleVersion := NewVersion(processName)
	sampleNamespaces := []*Namespace{}
	sampledSchema := &schema.Schema{}

	// databases that we're excluding from sampling
	dbSampleBlacklist := []string{"admin", "local", "system"}
	nsSampleBlacklist := []string{
		fmt.Sprintf("%q.%q", opts.Source, SchemasCollection),
		fmt.Sprintf("%q.%q", opts.Source, VersionsCollection),
		fmt.Sprintf("%q.%q", opts.Source, LockCollection),
	}

	sampleVersion.StartSampleTime = time.Now()

	ctx := session.Context()

	for db, collections := range mappings {
		if util.SliceContains(dbSampleBlacklist, db) {
			lgr.Logf(log.Info, "Skipping %q database", db)
			continue
		}

		sampledDB := &schema.Database{Name: db}

		for _, collection := range collections {

			ns := fmt.Sprintf("%q.%q", db, collection)

			if !nsMatcher.Has(db+"."+collection) ||
				util.SliceContains(nsSampleBlacklist, ns) {
				lgr.Logf(log.Info, "Skipping namespace %s", ns)
				continue
			}

			if _, ok := sampleVersion.FindDatabase(db); !ok {
				lgr.Logf(log.Info, "Mapping schema for database %q", db)
			}

			namespace := NewNamespace(db, collection, sampleVersion.Id)

			// 1. run sample command
			lgr.Logf(log.DebugLow, "Mapping schema for namespace %s", ns)
			pipeline := []bson.M{}
			if opts.Size != 0 {
				pipeline = append(pipeline, bson.M{"$sample": bson.M{"size": opts.Size}})
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
				lgr.Errf(log.Always, "error closing iterator: %v", err)
			}

			namespace.SampleSize = count
			namespace.Schema = jsonSchema

			// 4. convert the JSON schema to a relational schema
			err = sampledDB.Map(jsonSchema, collection, false, *lgr)
			if err != nil {
				return nil, nil, fmt.Errorf("error mapping schema: %v", err)
			}

			sampleNamespaces = append(sampleNamespaces, namespace)
			sampleVersion.AddNamespace(db, collection)
			lgr.Logf(log.DebugLow, "Finished mapping schema for namespace %s", ns)
		}

		if len(sampledDB.Tables) != 0 {
			sampledSchema.Databases = append(sampledSchema.Databases, sampledDB)
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

func LatestGeneration(opts *config.SchemaSampleOptions, session *mongodb.Session, lgr *log.Logger) (int64, error) {

	// 1. Find the latest version.
	oid, _, err := getLatestVersion(opts, session)
	if oid == nil || err != nil {
		return -1, err
	}

	// 2. Get the generation for the latest version
	return getGenerationForVersion(opts, session, *oid, lgr)
}

func getGenerationForVersion(cfg *config.SchemaSampleOptions, session *mongodb.Session, version bson.ObjectId, lgr *log.Logger) (int64, error) {

	var pipeline interface{} = []bson.D{
		{{"$match", bson.D{{VersionIdField, version}}}},
		{{"$project", bson.D{{VersionGenerationField, 1}}}},
	}

	cursor, err := session.Aggregate(cfg.Source, SchemasCollection, pipeline)
	if err != nil {
		return -1, err
	}
	defer cursor.Close(context.Background())

	result := struct {
		generation int64 `bson:"generation"`
	}{}

	if cursor.Next(session.Context(), &result) {
		return result.generation, nil
	}

	if cursor.Err() != nil {
		return -1, cursor.Err()
	}

	return -1, nil
}
