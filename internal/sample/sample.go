package sample

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/mapping"
	"github.com/10gen/sqlproxy/schema/mongo"
)

var (
	// Databases that we're excluding from sampling.
	dbSampleBlacklist = map[string]struct{}{
		"admin":  {},
		"config": {},
		"local":  {},
		"system": {},
	}
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

func (r *Record) getSchema(c SchemaSampleOptions,
	version []uint8, lg log.Logger) (*schema.Schema, error) {
	dbs := []*schema.Database{}

	sort.Slice(r.Namespaces, func(i, j int) bool {
		iLen := len(r.Namespaces[i].Collection)
		jLen := len(r.Namespaces[j].Collection)
		if iLen == jLen {
			return r.Namespaces[i].Collection > r.Namespaces[j].Collection
		}
		return iLen > jLen
	})

	seenDatabases := make(map[string]*schema.Database)
	for _, ns := range r.Namespaces {
		sampledDB, ok := seenDatabases[ns.Database]
		if !ok {
			sampledDB = schema.NewDatabase(lg, ns.Database, nil)
			dbs = append(dbs, sampledDB)
			seenDatabases[ns.Database] = sampledDB
		}

		err := mapping.Map(mapping.NewSchemaMappingConfig(
			sampledDB,
			ns.Schema,
			ns.Collection,
			false,
			c.uuidSubtype3Encoding,
			version,
			lg,
			c.schemaMappingHeuristic,
		))
		if err != nil {
			return nil, fmt.Errorf("error mapping schema version %#v, namespace %s: %v",
				r.Version, ns.QuotedString(), err)
		}
		// Mapping a schema can cause us to create significant amounts of garbage so we
		// block and allow the GC to complete before proceeding.
		runtime.GC()
	}

	return schema.New(dbs, r.Version.Alterations)
}

func (r *Record) validate() error {
	switch r.Version.Protocol {
	case Version1Protocol, Version2Protocol:
		// protocol version known
	default:
		// protocol version not known
		return fmt.Errorf(
			"cannot read stored schema of protocol version %q: current protocol is %q",
			r.Version.Protocol, CurrentProtocol,
		)
	}

	return r.validateNamespaceCount()
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
func FetchNamespaces(ctx context.Context, s *mongodb.Session, lgr log.Logger, match *util.Matcher) (NSMapping, error) {

	// If the matcher's inclusionary patterns don't include any wildcards, we can simply return the
	// namespaces that were specified without having to query MongoDB.
	if match.CanEnumerateAllNamespaces() {
		lgr.Debugf(log.Dev, "only literal namespaces provided, skipping listDatabases and "+
			"listCollections")
		var mappings NSMapping = map[string]NSCollections{}
		for db, cols := range match.Namespaces() {
			mappings[db] = cols
		}
		return mappings, nil
	}

	mappings := map[string]NSCollections{}
	dbs := []string{}

	// If the matcher's inclusionary patterns used a wildcard to specify databases then we need to
	// run listDatabases to get a list of all databases.
	if match.CanEnumerateAllDatabases() {
		lgr.Debugf(log.Dev, "only literal database names provided, skipping listDatabases")
		dbs = match.Databases()
	} else {
		if match.UsesWildcardDB() {
			lgr.Debugf(log.Dev, "wildcard database selector used: running listDatabases")
		} else {
			lgr.Debugf(log.Dev, "namespace exclusion selector used: running listDatabases")
		}

		dbIter, err := s.ListDatabases(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing databases: %v", err)
		}

		var dbResult struct {
			Name string `bson:"name"`
		}

		for dbIter.Next(ctx, &dbResult) {
			dbs = append(dbs, dbResult.Name)
		}

		if err := dbIter.Close(ctx); err != nil {
			lgr.Warnf(log.Dev, "error closing db iterator: %v", err)
		}

		if err := dbIter.Err(); err != nil {
			lgr.Warnf(log.Dev, "db iteration error: %v", err)
		}
	}

	lgr.Debugf(log.Dev, "finding namespaces in databases: %+v", dbs)

	// For each of the databases, if the collections to sample were enumerated literally, we return
	// that list of literals. if wildcards were used to specify collections, we run ListCollections
	// to get all of the collections.
	for _, db := range dbs {

		if _, ok := dbSampleBlacklist[db]; ok {
			lgr.Debugf(log.Dev, "skipping %q database", db)
			continue
		}

		if !match.HasDatabase(db) {
			lgr.Debugf(log.Dev, "exclusion database selector used for %q, skipping", db)
			continue
		}

		if match.CanEnumerateAllCollections(db) {
			lgr.Debugf(log.Dev, "only literal collection names provided for database %q, skipping "+
				"listCollections", db)
			mappings[db] = NSCollections(match.Collections(db))
			continue
		}

		if match.MustExcludeDatabase(db) {
			lgr.Debugf(log.Dev, "database %q is selected for exclusion", db)
			continue
		}

		collectionIter, err := s.ListCollections(ctx, db, ops.ListCollectionsOptions{})
		if err != nil {
			return nil, fmt.Errorf("can't get the collection names for '%v': %v", db, err)
		}

		var collectionResult struct {
			Name string `bson:"name"`
		}

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
// and collection - provided as a bson.D slice.
func getIndexes(ctx context.Context, database, collection string, session *mongodb.Session) ([]bson.D, error) {
	collectionIndexes, collectionIndex := []bson.D{}, bson.D{}
	cursor, err := session.ListIndexes(ctx, database, collection)
	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx, &collectionIndex) {
		collectionIndexes = append(collectionIndexes, collectionIndex)
	}

	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return collectionIndexes, nil
}

// InsertSampleRecord inserts the record - which includes version
// and namespace data - into the database specified in record.
func InsertSampleRecord(ctx context.Context, record *Record, session *mongodb.Session, lgr log.Logger) error {
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

		err := session.Run(ctx, record.Database, cmd, result)
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
	err := insertDocuments(mongodb.SchemasCollection, record.Namespaces)
	if err != nil {
		return err
	}

	// 2. insert the version document
	versionDocument := []interface{}{record.Version}
	err = insertDocuments(mongodb.VersionsCollection, versionDocument)
	if err != nil {
		return err
	}

	lgr.Infof(log.Admin, "schema persisted to MongoDB")

	return nil
}

// SchemaSampleOptions are a read only copy of the
// configuration SchemaSampleOptions.
type SchemaSampleOptions struct {
	source                 string
	mode                   config.SampleMode
	size                   int64
	optimizeViewSampling   bool
	preJoin                bool
	namespaces             []string
	refreshIntervalSecs    int64
	uuidSubtype3Encoding   string
	schemaMappingHeuristic config.MappingHeuristic
}

// WithOptimizeViewSampling sets the optimizeViewSampling field to the supplied value.
func (s *SchemaSampleOptions) WithOptimizeViewSampling(
	optimizeViewSampling bool) SchemaSampleOptions {
	s.optimizeViewSampling = optimizeViewSampling
	return *s
}

// WithSampleSize sets the size field to the supplied value.
func (s *SchemaSampleOptions) WithSampleSize(newSampleSize int64) SchemaSampleOptions {
	s.size = newSampleSize
	return *s
}

// NewSchemaSampleOptions creates a new read only snapshot of the SchemaSampleOptions.
func NewSchemaSampleOptions(cfg *config.SchemaSampleOptions) SchemaSampleOptions {
	nameSpaceCopy := make([]string, len(cfg.Namespaces))
	copy(nameSpaceCopy, cfg.Namespaces)
	return SchemaSampleOptions{
		source:                 cfg.Source,
		mode:                   cfg.Mode,
		size:                   cfg.Size,
		optimizeViewSampling:   cfg.OptimizeViewSampling,
		preJoin:                cfg.PreJoin,
		namespaces:             nameSpaceCopy,
		refreshIntervalSecs:    cfg.RefreshIntervalSecs,
		uuidSubtype3Encoding:   cfg.UUIDSubtype3Encoding,
		schemaMappingHeuristic: cfg.SchemaMappingHeuristic,
	}
}

// NewSchemaSampleOptionsWithHeuristic creates a new read only snapshot of the
// SchemaSampleOptions with a specified heuristic.
func NewSchemaSampleOptionsWithHeuristic(cfg *config.SchemaSampleOptions,
	heuristic config.MappingHeuristic) SchemaSampleOptions {
	nameSpaceCopy := make([]string, len(cfg.Namespaces))
	copy(nameSpaceCopy, cfg.Namespaces)
	return SchemaSampleOptions{
		source:                 cfg.Source,
		mode:                   cfg.Mode,
		size:                   cfg.Size,
		preJoin:                cfg.PreJoin,
		optimizeViewSampling:   cfg.OptimizeViewSampling,
		namespaces:             nameSpaceCopy,
		refreshIntervalSecs:    cfg.RefreshIntervalSecs,
		uuidSubtype3Encoding:   cfg.UUIDSubtype3Encoding,
		schemaMappingHeuristic: heuristic,
	}
}

// ReadSchema reads a schema stored in the configuration sampling source
// database and returns a relational representation of the schema. If no
// such schema exists, it returns nil.
func ReadSchema(ctx context.Context, cfg SchemaSampleOptions, session *mongodb.Session,
	lgr log.Logger) (*schema.Schema, error) {

	// get the latest stored schema record
	lgr.Infof(log.Admin, "retrieving latest schema")
	rec, err := LatestRecord(ctx, cfg, session)
	if rec == nil || err != nil {
		return nil, err
	}

	version, err := session.Version()
	if err != nil {
		return nil, fmt.Errorf("failed to obtain MongoDB server "+
			"version during call to ReadSchema: %s", err)
	}

	// get the schema from the record
	versionedSchema, err := rec.getSchema(cfg, version, lgr)
	if err != nil {
		return nil, err
	}

	return versionedSchema, nil
}

// Schema uses the provided mongosqld configuration and session
// to sample namespaces. It returns the relational schema generated
// and the version/schemas documents resulting from sampling.
func Schema(ctx context.Context, cfg SchemaSampleOptions, processName string,
	session *mongodb.Session, lgr log.Logger) (*schema.Schema, *Record, error) {

	namespaces := cfg.namespaces

	var err error
	var nsMatcher *util.Matcher
	nsMatcher, err = util.NewMatcher(namespaces)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid specification: %v", err)
	}

	lgr.Infof(log.Always, "sampling MongoDB for schema...")

	var mappings NSMapping
	mappings, err = FetchNamespaces(ctx, session, lgr, nsMatcher)
	if err != nil {
		return nil, nil, err
	}

	sampleVersion := NewVersion(processName)
	sampleNamespaces := []*Namespace{}

	sampledDatabases := []*schema.Database{}
	uuidSubtype3Encoding := cfg.uuidSubtype3Encoding

	// Sample source collections should not be sampled.
	nsSampleBlacklist := []string{
		NewNamespaceWithoutID(cfg.source, mongodb.SchemasCollection).QuotedString(),
		NewNamespaceWithoutID(cfg.source, mongodb.VersionsCollection).QuotedString(),
		NewNamespaceWithoutID(cfg.source, mongodb.LockCollection).QuotedString(),
	}

	sampleVersion.StartSampleTime = time.Now()

	for db, collections := range mappings {
		if _, ok := dbSampleBlacklist[db]; ok {
			lgr.Debugf(log.Dev, "skipping %q database", db)
			continue
		}

		sampledDB := schema.NewDatabase(lgr, db, nil)

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

		queryViewPerNamespace := false
		var nsViewPipelines map[string]NSViewPipeline
		nsViewPipelines, err = GetViewPipelinesInDatabase(ctx, session, db)
		if err != nil {
			lgr.Debugf(log.Dev, "unable to get view pipeline for database %q: %v", db, err)
			queryViewPerNamespace = true
		}

		for _, col := range collections {
			namespace := NewNamespaceWithoutID(db, col)
			sampleCollection := col
			quotedNs, ns := namespace.QuotedString(), namespace.String()

			if util.SliceContains(nsSampleBlacklist, quotedNs) {
				lgr.Debugf(log.Dev, "skipping sample source namespace %s", quotedNs)
				continue
			}

			if strings.HasPrefix(col, "system.") {
				lgr.Debugf(log.Dev, "skipping system collection %s", quotedNs)
				continue
			}

			if !nsMatcher.Has(ns) {
				continue
			}

			if _, ok := sampleVersion.FindDatabase(db); !ok {
				lgr.Debugf(log.Dev, "mapping schema for database %q", db)
			}

			namespace = NewNamespace(db, col, sampleVersion.ID)

			// 1. run sample command
			lgr.Debugf(log.Dev, "mapping schema for namespace %s", quotedNs)

			pipeline := getSamplingPipeline(cfg.size)

			if cfg.optimizeViewSampling {
				var viewPipeline NSViewPipeline

				if queryViewPerNamespace {
					viewPipeline, err = getViewPipelineForNamespace(ctx, session, db, col)
					if err != nil {
						lgr.Debugf(log.Dev, "unable to get view pipeline for namespace %s: %v",
							quotedNs, err)
					}
				} else {
					viewPipeline = nsViewPipelines[ns]
				}

				if len(viewPipeline.Pipeline) != 0 {
					if bsonutil.ContainsCardinalityAlteringStages(viewPipeline.Pipeline) {
						lgr.Debugf(log.Dev, "view pipeline for %s contains cardinality altering "+
							"skipping $sample optimization", quotedNs)
					} else {
						lgr.Debugf(log.Dev, "prepending $sample for view %s", quotedNs)
						pipeline = append(pipeline, viewPipeline.Pipeline...)
						sampleCollection = viewPipeline.Collection
					}
				}
			}

			namespace.StartSampleTime = time.Now()

			// 2. get sample documents
			var iter ops.Cursor
			iter, err = session.Aggregate(ctx, db, sampleCollection, pipeline)
			if err != nil {
				// If we attempted to optimize for views, it is possible that the admin user does
				// not possess "find" access on some referenced collection or view - either within
				// the pipeline or on the sampled collection. In that case, we fall back on the
				// regular sampling process.
				if len(pipeline) > 1 {
					lgr.Debugf(log.Dev, "using vanilla $sample for view %s", quotedNs)
					iter, err = session.Aggregate(ctx, db, col, getSamplingPipeline(cfg.size))
				}
				if err != nil {
					return nil, nil, fmt.Errorf("error sampling collection: %v", err)
				}
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

			if count == 0 {
				lgr.Debugf(log.Dev, "skipping namespace %s: no documents found", quotedNs)
				continue
			}

			if err = iter.Close(ctx); err != nil {
				lgr.Warnf(log.Dev, "error closing iterator: %v", err)
			}

			var indexes []bson.D

			// Index listing is not supported for views.
			if _, ok := nsViewPipelines[ns]; !ok {
				indexes, err = getIndexes(ctx, db, col, session)
				if err != nil {
					lgr.Warnf(log.Dev, "error getting indexes: %v", err)
				}
			}

			jsonSchema.AddIndexes(indexes)
			jsonSchema.InferSpecialTypes()

			namespace.SampleSize = count
			namespace.Schema = jsonSchema

			// 4. convert the JSON schema to a relational schema
			version, versionErr := session.Version()
			if versionErr != nil {
				return nil, nil,
					fmt.Errorf("failed to obtain MongoDB server version "+
						"during call to ReadSchema: %s", versionErr)
			}

			err = mapping.Map(mapping.NewSchemaMappingConfig(
				sampledDB,
				jsonSchema,
				col,
				cfg.preJoin,
				uuidSubtype3Encoding,
				version,
				lgr,
				cfg.schemaMappingHeuristic,
			))

			if err != nil {
				return nil, nil, fmt.Errorf("error mapping schema: %v", err)
			}
			// Mapping a schema can cause us to create significant amounts of garbage so we
			// block and allow the GC to complete before proceeding.
			runtime.GC()

			sampleNamespaces = append(sampleNamespaces, namespace)
			sampleVersion.AddNamespace(db, col)
			lgr.Debugf(log.Dev, "finished mapping schema for namespace %s", quotedNs)
		}

		if len(sampledDB.Tables()) != 0 {
			sampledDatabases = append(sampledDatabases, sampledDB)
		}
	}

	sampleVersion.EndSampleTime = time.Now()

	sampleData := &Record{
		Database:   cfg.source,
		Namespaces: sampleNamespaces,
		Version:    sampleVersion,
	}

	if count := len(sampleNamespaces); count != 0 {
		nsStr := util.Pluralize(count, "namespace", "namespaces")
		lgr.Infof(log.Always, "mapped schema for %v %v: %v", count, nsStr, sampleVersion.Databases)
	} else {
		lgr.Infof(log.Always, "no namespaces were sampled")
	}

	var sampledSchema *schema.Schema
	sampledSchema, err = schema.New(sampledDatabases, nil)
	return sampledSchema, sampleData, err
}

// LatestGeneration returns the most recent generation of the schema stored in MongoDB
func LatestGeneration(ctx context.Context, opts SchemaSampleOptions, session *mongodb.Session,
	lgr log.Logger) (int64, error) {

	rec, err := LatestRecord(ctx, opts, session)
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
func LatestRecord(ctx context.Context, opts SchemaSampleOptions, s *mongodb.Session) (rec *Record, err error) {
	var pipeline interface{} = []bson.D{
		{{Name: "$sort", Value: bson.D{{Name: "generation", Value: -1}}}},
		{{Name: "$limit", Value: 1}},
		{{Name: "$project", Value: bson.D{
			{Name: "_id", Value: 0},
			{Name: "version", Value: "$$CURRENT"},
		}}},
		{{Name: "$lookup", Value: bson.D{
			{Name: "from", Value: mongodb.SchemasCollection},
			{Name: "localField", Value: "version._id"},
			{Name: "foreignField", Value: mongodb.VersionIDField},
			{Name: "as", Value: "namespaces"},
		}}},
	}

	var cursor mongodb.Cursor
	cursor, err = s.Aggregate(ctx, opts.source, mongodb.VersionsCollection, pipeline)
	if err != nil {
		return nil, err
	}
	defer util.CheckDeferredFuncWithContext(context.Background(), cursor.Close, &err)

	rec = &Record{}
	if cursor.Next(ctx, rec) {
		rec.Database = opts.source
		err = rec.validate()
		if err != nil {
			return nil, err
		}
		return rec, cursor.Err()
	}

	return nil, cursor.Err()
}

// NSViewPipeline holds the pipeline used in creating a view together with the collection on which
// the view is defined.
type NSViewPipeline struct {
	Collection string
	Pipeline   []bson.D
}

// GetViewPipelinesInDatabase returns a map of namespace names to the viewPipeline for the views
// within the database, db.
func GetViewPipelinesInDatabase(ctx context.Context, s *mongodb.Session, db string) (map[string]NSViewPipeline, error) {
	type cursorCollection struct {
		Name    string `bson:"name"`
		Type    string `bson:"type"`
		Options struct {
			Pipeline []bson.D `bson:"pipeline"`
			ViewOn   string   `bson:"viewOn"`
		} `bson:"options"`
	}

	type firstBatchCursorResult struct {
		FirstBatch []cursorCollection `bson:"firstBatch"`
	}

	type cursorReturningResult struct {
		Cursor firstBatchCursorResult `bson:"cursor"`
		Ok     int                    `bson:"ok"`
	}

	result := &cursorReturningResult{}
	if err := s.Run(ctx, db, bson.D{{Name: "listCollections", Value: 1}}, result); err != nil {
		return nil, fmt.Errorf("error getting db views map: %v", err)
	}

	if result.Ok != 1 {
		return nil, fmt.Errorf("error executing db views map")
	}

	nsViewPipelines := make(map[string]NSViewPipeline)
	for _, collection := range result.Cursor.FirstBatch {
		if collection.Type == "view" {
			namespace := NewNamespaceWithoutID(db, collection.Name).String()
			nsViewPipelines[namespace] = NSViewPipeline{
				Collection: collection.Options.ViewOn,
				Pipeline:   collection.Options.Pipeline,
			}
		}
	}

	// Recursively augment initial pipelines to handle views defined on views.
	for namespace, pipeline := range nsViewPipelines {
		sourcePipeline, sourceCollection := pipeline.Pipeline, pipeline.Collection
		source, ok := nsViewPipelines[NewNamespaceWithoutID(db, sourceCollection).String()]
		for ok {
			sourcePipeline = append(append([]bson.D{}, source.Pipeline...), sourcePipeline...)
			sourceCollection = source.Collection
			source, ok = nsViewPipelines[NewNamespaceWithoutID(db, source.Collection).String()]
		}
		nsViewPipelines[namespace] = NSViewPipeline{
			Collection: sourceCollection,
			Pipeline:   sourcePipeline,
		}
	}

	return nsViewPipelines, nil
}

// getSamplingPipeline returns a slice of bson documents based on the given sampleSize.
func getSamplingPipeline(sampleSize int64) []bson.D {
	if sampleSize != 0 {
		return []bson.D{{{Name: "$sample", Value: bson.D{{Name: "size", Value: sampleSize}}}}}
	}
	return nil
}

// getViewPipelineForNamespace returns the view for the given namespace or an empty NSViewPipeline
// pipeline  if the namespace is not a view.
func getViewPipelineForNamespace(ctx context.Context, s *mongodb.Session, db, col string) (NSViewPipeline, error) {
	pipeline := NSViewPipeline{}

	type explainResult struct {
		Stages []bson.D `bson:"stages"`
		Ok     int      `bson:"ok"`
	}

	cmd := bson.D{{Name: "explain", Value: bson.D{{Name: "find", Value: col}}}}
	result := &explainResult{}
	if err := s.Run(ctx, db, cmd, result); err != nil {
		return pipeline, fmt.Errorf("error getting db views map: %v", err)
	}

	if result.Ok != 1 {
		return pipeline, fmt.Errorf("error executing db views map")
	}

	if len(result.Stages) < 1 {
		return pipeline, nil
	}

	// For views, the first stage is always $cursor.
	pipeline = NSViewPipeline{
		Collection: col,
		Pipeline:   result.Stages[1:],
	}

	return pipeline, nil
}
