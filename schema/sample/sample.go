package sample

import (
	"context"
	"fmt"
	"sort"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb/provider"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
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

// NSCollections is a list of collection names.
type NSCollections []string

// NSPair is a pair of Database name and slice of Collections in that Database.
type NSPair struct {
	Database    string
	Collections NSCollections
}

// NSMapping is a map from string to slice of Collections.
type NSMapping map[string]NSCollections

// fetchSortedNamespaces returns the fetched namespaces sorted by database name.
// We need this to ensure a stable schema in the face of hitting the global table limit.
func fetchSortedNamespaces(ctx context.Context, s *provider.Session, lgr log.Logger, match *strutil.Matcher) ([]NSPair, error) {
	namespaces, err := fetchNamespaces(ctx, s, lgr, match)
	if err != nil {
		return nil, err
	}
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].Database < namespaces[j].Database })
	return namespaces, nil
}

// fetchNamespaceMap provides the same data as fetchNamespaces in a map form
func fetchNamespaceMap(ctx context.Context, s *provider.Session, lgr log.Logger, match *strutil.Matcher) (NSMapping, error) {
	namespaces, err := fetchNamespaces(ctx, s, lgr, match)
	if err != nil {
		return nil, err
	}
	ret := make(NSMapping, len(namespaces))
	for _, ns := range namespaces {
		ret[ns.Database] = ns.Collections
	}
	return ret, nil
}

// fetchNamespaces returns a slice of database, collections pairs that exist in the MongoDB cluster
// to the collection(s) within each database.
func fetchNamespaces(ctx context.Context, s *provider.Session, lgr log.Logger, match *strutil.Matcher) ([]NSPair, error) {

	// If the matcher's inclusionary patterns don't include any wildcards, we can simply return the
	// namespaces that were specified without having to query MongoDB.
	if match.CanEnumerateAllNamespaces() {
		lgr.Debugf(log.Dev, "only literal namespaces provided, skipping listDatabases and "+
			"listCollections")
		namespaces := match.Namespaces()
		mappings := make([]NSPair, 0, len(namespaces))
		for db, cols := range match.Namespaces() {
			mappings = append(mappings, NSPair{Database: db, Collections: cols})
		}
		return mappings, nil
	}

	mappings := []NSPair{}
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

		dbResult, err := s.ListDatabases(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing databases: %v", err)
		}

		for _, dbRecord := range dbResult.Databases {
			dbs = append(dbs, dbRecord.Name)
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
			mappings = append(mappings, NSPair{Database: db, Collections: NSCollections(match.Collections(db))})
			continue
		}

		if match.MustExcludeDatabase(db) {
			lgr.Debugf(log.Dev, "database %q is selected for exclusion", db)
			continue
		}

		cursor, err := s.ListCollections(ctx, db, driver.CursorOptions{})
		if err != nil {
			return nil, fmt.Errorf("can't get the collection names for '%v': %v", db, err)
		}

		var collectionResult struct {
			Name string `bson:"name"`
		}

		collections := []string{}

		for cursor.Next(ctx, &collectionResult) {
			collections = append(collections, collectionResult.Name)
		}

		if err := cursor.Err(); err != nil {
			lgr.Warnf(log.Dev, "collection iteration error: %v", err)
		}

		if err := cursor.Close(ctx); err != nil {
			lgr.Warnf(log.Dev, "error closing collection iterator: %v", err)
		}

		mappings = append(mappings, NSPair{Database: db, Collections: NSCollections(collections)})
	}

	return mappings, nil
}

// getIndexes returns the indexes present in the namespace - database
// and collection - provided as a bson.D slice.
func getIndexes(ctx context.Context, database, collection string, session *provider.Session) ([]bson.D, error) {
	collectionIndexes, collectionIndex := bsonutil.NewDArray(), bsonutil.NewD()
	cursor, err := session.ListIndexes(ctx, database, collection)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	for cursor.Next(ctx, &collectionIndex) {
		collectionIndexes = append(collectionIndexes, collectionIndex)
		collectionIndex = bsonutil.NewD()
	}

	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return collectionIndexes, nil
}

func formatNamespace(db, collection string, quote bool) string {
	if quote {
		return fmt.Sprintf("%q.%q", db, collection)
	}
	return fmt.Sprintf("%s.%s", db, collection)
}

// NSViewPipeline holds the pipeline used in creating a view together with the collection on which
// the view is defined.
type NSViewPipeline struct {
	Collection string
	Pipeline   []bson.D
}

// GetViewPipelinesInDatabase returns a map of namespace names to the viewPipeline for the views
// within the database, db.
func GetViewPipelinesInDatabase(ctx context.Context, s *provider.Session, db string) (map[string]NSViewPipeline, error) {
	cursor, err := s.ListCollections(ctx, db, driver.CursorOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting db views map: %v", err)
	}

	nsViewPipelines := make(map[string]NSViewPipeline)

	collectionResult := struct {
		Name    string `bson:"name"`
		Type    string `bson:"type"`
		Options struct {
			Pipeline []bson.D `bson:"pipeline"`
			ViewOn   string   `bson:"viewOn"`
		} `bson:"options"`
	}{}

	for cursor.Next(ctx, &collectionResult) {
		if collectionResult.Type == "view" {
			namespace := formatNamespace(db, collectionResult.Name, false)
			nsViewPipelines[namespace] = NSViewPipeline{
				Collection: collectionResult.Options.ViewOn,
				Pipeline:   collectionResult.Options.Pipeline,
			}
		}
		collectionResult.Options.Pipeline = bsonutil.NewDArray()
	}

	// Recursively augment initial pipelines to handle views defined on views.
	for namespace, pipeline := range nsViewPipelines {
		sourcePipeline, sourceCollection := pipeline.Pipeline, pipeline.Collection
		source, ok := nsViewPipelines[formatNamespace(db, sourceCollection, false)]
		for ok {
			sourcePipeline = append(append(bsonutil.NewDArray(), source.Pipeline...), sourcePipeline...)
			sourceCollection = source.Collection
			source, ok = nsViewPipelines[formatNamespace(db, source.Collection, false)]
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
		return bsonutil.NewDArray(bsonutil.NewD(bsonutil.NewDocElem("$sample", bsonutil.NewD(bsonutil.NewDocElem("size", sampleSize)))))
	}
	return nil
}

// getViewPipelineForNamespace returns the view for the given namespace or an empty NSViewPipeline
// pipeline  if the namespace is not a view.
func getViewPipelineForNamespace(ctx context.Context, s *provider.Session, db, col string) (NSViewPipeline, error) {
	pipeline := NSViewPipeline{}

	type explainResult struct {
		Stages []bson.D `bson:"stages"`
		Ok     int      `bson:"ok"`
	}

	cmd := bsonutil.NewD(bsonutil.NewDocElem("explain", bsonutil.NewD(bsonutil.NewDocElem("find", col))))
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
