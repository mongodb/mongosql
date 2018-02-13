package sample

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
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
	VersionIDField         = "versionId"
	VersionGenerationField = "generation"
)

// Record holds data generated as a result
// of performing sampling operations.
type Record struct {
	Database   string
	Version    *Version
	Namespaces []*Namespace
}

// NSMapping is a mapping of database names to NSCollections.
type NSMapping map[string]NSCollections

// NSCollections is a list of collection names.
type NSCollections []string

// Alter changes a Record to represent a new Version of the current record with
// the provided alterations applied.
func (r *Record) Alter(alts []*schema.Alteration) {
	id := bson.NewObjectId()
	r.Version.ID = id
	r.Version.Generation++
	r.Version.Alterations = append(r.Version.Alterations, alts...)
	r.Version.Protocol = CurrentProtocol
	for _, ns := range r.Namespaces {
		ns.id = bson.NewObjectId()
		ns.VersionID = id
	}
}

func (r *Record) getSchema(cfg *config.SchemaSampleOptions, lgr *log.Logger) (*schema.Schema, error) {
	sampledSchema := &schema.Schema{
		Alterations: r.Version.Alterations,
	}

	sort.Slice(r.Namespaces, func(i, j int) bool {
		iLen := len(r.Namespaces[i].Collection)
		jLen := len(r.Namespaces[j].Collection)
		if iLen == jLen {
			return r.Namespaces[i].Collection > r.Namespaces[j].Collection
		}
		return iLen > jLen
	})

	seenDatabases := make(map[string]*schema.Database, 0)
	for _, ns := range r.Namespaces {
		sampledDB, ok := seenDatabases[ns.Database]
		if !ok {
			sampledDB = &schema.Database{
				Name: ns.Database,
			}
			sampledSchema.Databases = append(sampledSchema.Databases, sampledDB)
			seenDatabases[ns.Database] = sampledDB
		}

		err := sampledDB.Map(ns.Schema, ns.Collection, false, cfg.UUIDSubtype3Encoding, lgr)
		if err != nil {
			return nil, fmt.Errorf("error mapping schema version %#v, namespace %q.%q: %v",
				r.Version, ns.Database, ns.Collection, err)
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
func FetchNamespaces(session *mongodb.Session, lgr *log.Logger, matcher *util.Matcher) (NSMapping, error) {

	// if the matcher doesn't include any wildcards, we can simply return the
	// namespaces that were specified without having to query MongoDB
	if !matcher.UsesAnyWildcardCollection() && !matcher.UsesWildcardDB() {
		lgr.Debugf(log.Dev, "only literal namespaces provided, skipping listDatabases and listCollections")
		var mappings NSMapping = map[string]NSCollections{}
		for db, cols := range matcher.Namespaces() {
			mappings[db] = cols
		}
		return mappings, nil
	}

	mappings := map[string]NSCollections{}
	dbs := []string{}

	// if the matcher used a wildcard to specify databases, then we need to run
	// listDatabases to get a list of all databases
	if matcher.UsesWildcardDB() {
		lgr.Debugf(log.Dev, "wildcard database selector used: running listDatabases")

		dbIter, err := session.ListDatabases()
		if err != nil {
			return nil, fmt.Errorf("error listing databases: %v", err)
		}

		var dbResult struct {
			Name string `bson:"name"`
		}

		for dbIter.Next(session.Context(), &dbResult) {
			dbs = append(dbs, dbResult.Name)
		}

		if err := dbIter.Close(session.Context()); err != nil {
			lgr.Warnf(log.Dev, "error closing db iterator: %v", err)
		}

		if err := dbIter.Err(); err != nil {
			lgr.Warnf(log.Dev, "db iteration error: %v", err)
		}
	} else {
		lgr.Debugf(log.Dev, "only literal database names provided, skipping listDatabases")
		dbs = matcher.Databases()
	}

	lgr.Debugf(log.Dev, "finding namespaces in databases: %+v", dbs)

	// for each of the databases, if the collections to sample were enumerated
	// literally, we return that list of literals. if wildcards were used to
	// specify collections, we run ListCollections to get all of the collections
	for _, db := range dbs {

		if !matcher.HasDatabase(db) {
			continue
		}

		if !matcher.UsesWildcardCollection(db) {
			lgr.Debugf(log.Dev, "only literal collection names provided for database '%s', skipping listCollections", db)
			mappings[db] = NSCollections(matcher.Collections(db))
			continue
		}

		lgr.Debugf(log.Dev, "wildcard collection selector used for db %s: running listCollections", db)

		collectionIter, err := session.ListCollections(db, ops.ListCollectionsOptions{})
		if err != nil {
			return nil, fmt.Errorf("can't get the collection "+
				"names for '%v': %v", db, err)
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

		mappings[db] = NSCollections(collections)
	}

	return mappings, nil
}

// getIndexes returns the indexes present in the namespace - database
// and collection - provided as a bson.D slice
func getIndexes(database, collection string, session *mongodb.Session) ([]bson.D, error) {
	collectionIndexes, collectionIndex := []bson.D{}, bson.D{}
	cursor, err := session.ListIndexes(database, collection)
	if err != nil {
		return nil, err
	}

	for cursor.Next(session.Context(), &collectionIndex) {
		collectionIndexes = append(collectionIndexes, collectionIndex)
	}

	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return collectionIndexes, nil
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
			{Name: "insert", Value: collection},
			{Name: "documents", Value: documents},
			{Name: "writeConcern", Value: bson.D{{Name: "w", Value: "majority"}}},
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

// Schema uses the provided mongosqld configuration and session
// to sample namespaces. It returns the relational schema generated
// and the version/schemas documents resulting from sampling.
func Schema(cfg *config.SchemaSampleOptions, processName string,
	session *mongodb.Session, lgr *log.Logger) (*schema.Schema, *Record, error) {

	namespaces := cfg.Namespaces

	nsMatcher, err := util.NewMatcher(namespaces)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid specification: %v", err)
	}

	lgr.Infof(log.Always, "sampling MongoDB for schema...")

	mappings, err := FetchNamespaces(session, lgr, nsMatcher)
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

		// Map the collections in descending order of length to
		// handle possible conflicts in array field names and
		// existing collection names.
		//
		// For example, if we have the following collections in
		// MongoDB: "foo" AND "foo_Xx_0".
		//
		// When mapped in the order above, if "foo" only contains
		// a document like:
		//
		//		"xX" : [ { "c" : 1 } ],
		//		"XX" : 2,
		//		"Xx" : [ { "b" : 3 } ],
		//		"xX_0" : 4
		//
		// We would naively map the entire database with the
		// following tables:
		//
		//  	+--------------------------------+
		//  	| Tables_in_test                 |
		//  	+--------------------------------+
		//  	| foo                            |
		//  	| foo_Xx_0                       |
		//  	| foo_Xx_0_0                     |
		//  	| foo_xX                         |
		//  	+--------------------------------+
		//
		// However, instead of mapping the MongoDB "foo_Xx_0"
		// collection verbatim, we map it as "foo_Xx_0_0".
		// This is because of the iteration order. In order to avoid this,
		// we sort the collection names in descending order of length - thereby
		// guaranteeing this doesn't happen.
		//
		// Note that the database tables are subsequently sorted so this update
		// does not affect the order in which table are displayed to users.
		sort.Slice(collections, func(i, j int) bool {
			return len(collections[i]) > len(collections[j])
		})

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

			namespace := NewNamespace(db, collection, sampleVersion.ID)

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
					return nil, nil, fmt.Errorf("error including sample: %v", err)
				}
				doc = &bson.D{}
				count++
			}

			if err := iter.Close(ctx); err != nil {
				lgr.Warnf(log.Dev, "error closing iterator: %v", err)
			}

			indexes, err := getIndexes(db, collection, session)
			if err != nil {
				lgr.Warnf(log.Dev, "error getting indexes: %v", err)
			}

			jsonSchema.AddIndexes(indexes)
			jsonSchema.InferSpecialTypes()

			namespace.SampleSize = count
			namespace.Schema = jsonSchema

			// 4. convert the JSON schema to a relational schema
			err = sampledDB.Map(jsonSchema, collection, false, uuidSubtype3Encoding, lgr)
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

	if l := len(sampleNamespaces); l != 0 {
		nsStr := util.Pluralize(l, "namespace", "namespaces")
		lgr.Infof(log.Always, "mapped schema for %v %v: %v",
			l, nsStr, sampleVersion.Databases)
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
	} else if rec == nil {
		lgr.Debugf(log.Dev, "no existing records found in sample source db")
		return -1, nil
	}

	return rec.Version.Generation, nil
}

// LatestRecord returns a Record representing the most recent generation of the
// schema stored in MongoDB. If there is no schema currently stored in MongoDB,
// LatestRecord returns a nil Record.
func LatestRecord(opts *config.SchemaSampleOptions, session *mongodb.Session, lgr *log.Logger) (*Record, error) {
	var pipeline interface{} = []bson.D{
		{{Name: "$sort", Value: bson.D{{Name: "generation", Value: -1}}}},
		{{Name: "$limit", Value: 1}},
		{{Name: "$project", Value: bson.D{
			{Name: "_id", Value: 0},
			{Name: "version", Value: "$$CURRENT"},
		}}},
		{{Name: "$lookup", Value: bson.D{
			{Name: "from", Value: SchemasCollection},
			{Name: "localField", Value: "version._id"},
			{Name: "foreignField", Value: VersionIDField},
			{Name: "as", Value: "namespaces"},
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
