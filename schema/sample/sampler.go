package sample

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/mapping"
	"github.com/10gen/sqlproxy/schema/mongo"

	"go.mongodb.org/mongo-driver/bson"
)

// Sampler is a type that provides schema sampling functionality.
type Sampler struct {
	cfg Config
	lg  log.Logger
	sp  *mongodb.SessionProvider
}

// NewSampler returns a new Sampler with the provided Config, Logger, and
// SessionProvider.
func NewSampler(cfg Config, lg log.Logger, sp *mongodb.SessionProvider) Sampler {
	component := fmt.Sprintf("%-10v [sampler]", log.SchemaComponent)
	lg = log.NewComponentLogger(component, lg)
	return Sampler{
		cfg: cfg,
		lg:  lg,
		sp:  sp,
	}
}

// Sample samples the MongoDB namespaces indicated by the sampler config. It
// returns the relational schema generated from the sampled schema.
func (s Sampler) Sample(ctx context.Context) (*schema.Schema, error) {
	session, err := s.sp.AuthenticatedAdminSessionPrimary(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	namespaces := s.cfg.Namespaces()

	var nsMatcher *strutil.Matcher
	nsMatcher, err = strutil.NewMatcher(namespaces)
	if err != nil {
		return nil, fmt.Errorf("invalid specification: %v", err)
	}

	s.lg.Infof(log.Always, "sampling MongoDB for schema...")

	var mappings NSMapping
	mappings, err = fetchNamespaces(ctx, session, s.lg, nsMatcher)
	if err != nil {
		return nil, err
	}

	sampledNamespaces := map[string][]string{}
	sampledNamespaceCount := 0
	sampledDatabases := []*schema.Database{}
	uuidSubtype3Encoding := s.cfg.UUIDSubtype3Encoding()

	var addSampledNamespace = func(db, col string) {
		cols, ok := sampledNamespaces[db]
		if !ok {
			cols = []string{}
		}
		cols = append(cols, col)
		sampledNamespaces[db] = cols
	}

	// Sample source collections should not be sampled.
	nsSampleBlacklist := []string{
		formatNamespace(s.cfg.Source(), mongodb.SchemasCollection, true),
		formatNamespace(s.cfg.Source(), mongodb.VersionsCollection, true),
		formatNamespace(s.cfg.Source(), mongodb.LockCollection, true),
	}

	for db, collections := range mappings {
		if _, ok := dbSampleBlacklist[db]; ok {
			s.lg.Debugf(log.Dev, "skipping %q database", db)
			continue
		}

		sampledDB := schema.NewDatabase(s.lg, db, nil)

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
			if len(collections[i]) > len(collections[j]) {
				return true
			}
			// We need to be careful that two strings of the same length are
			// always sorted the same way in case mongod returns them in
			// different order. Previously, we relied on their being sorted
			// from least to greatest (X before x in ASCII order), which
			// complicates the logic, slightly.
			return len(collections[i]) == len(collections[j]) && collections[i] < collections[j]
		})

		queryViewPerNamespace := false
		var nsViewPipelines map[string]NSViewPipeline
		nsViewPipelines, err = GetViewPipelinesInDatabase(ctx, session, db)
		if err != nil {
			s.lg.Debugf(log.Dev, "unable to get view pipeline for database %q: %v", db, err)
			queryViewPerNamespace = true
		}

		for _, col := range collections {
			sampleCollection := col
			quotedNs := formatNamespace(db, col, true)
			unquotedNs := formatNamespace(db, col, false)

			if strutil.SliceContains(nsSampleBlacklist, quotedNs) {
				s.lg.Debugf(log.Dev, "skipping sample source namespace %s", quotedNs)
				continue
			}

			if strings.HasPrefix(col, "system.") {
				s.lg.Debugf(log.Dev, "skipping system collection %s", quotedNs)
				continue
			}

			if !nsMatcher.Has(unquotedNs) {
				continue
			}

			if _, ok := sampledNamespaces[db]; !ok {
				s.lg.Debugf(log.Dev, "mapping schema for database %q", db)
			}

			// 1. run sample command
			s.lg.Debugf(log.Dev, "mapping schema for namespace %s", quotedNs)

			pipeline := getSamplingPipeline(s.cfg.Size())

			if s.cfg.OptimizeViewSampling() {
				var viewPipeline NSViewPipeline

				if queryViewPerNamespace {
					viewPipeline, err = getViewPipelineForNamespace(ctx, session, db, col)
					if err != nil {
						s.lg.Debugf(log.Dev, "unable to get view pipeline for namespace %s: %v",
							quotedNs, err)
					}
				} else {
					viewPipeline = nsViewPipelines[unquotedNs]
				}

				if len(viewPipeline.Pipeline) != 0 {
					if bsonutil.ContainsCardinalityAlteringStages(viewPipeline.Pipeline) {
						s.lg.Debugf(log.Dev, "view pipeline for %s contains cardinality altering "+
							"skipping $sample optimization", quotedNs)
					} else {
						s.lg.Debugf(log.Dev, "prepending $sample for view %s", quotedNs)
						pipeline = append(pipeline, viewPipeline.Pipeline...)
						sampleCollection = viewPipeline.Collection
					}
				}
			}

			// 2. get sample documents
			var cursor mongodb.Cursor
			cursor, err = session.Aggregate(ctx, db, sampleCollection, pipeline)
			if err != nil {
				// If we attempted to optimize for views, it is possible that the admin user does
				// not possess "find" access on some referenced collection or view - either within
				// the pipeline or on the sampled collection. In that case, we fall back on the
				// regular sampling process.
				if len(pipeline) > 1 {
					s.lg.Debugf(log.Dev, "using vanilla $sample for view %s", quotedNs)
					cursor, err = session.Aggregate(ctx, db, col, getSamplingPipeline(s.cfg.Size()))
				}
				if err != nil {
					return nil, fmt.Errorf("error sampling collection: %v", err)
				}
			}

			jsonSchema := mongo.NewCollectionSchema()

			// 3. create json schema and store it
			count, doc := int64(0), bsonutil.NewD()

			for cursor.Next(ctx, &doc) {
				err = jsonSchema.IncludeSample(doc)
				if err != nil {
					return nil, fmt.Errorf("error including sample: %v", err)
				}
				doc = bsonutil.NewD()
				count++
			}

			if err = cursor.Err(); err != nil {
				return nil, fmt.Errorf("error iterating sample: %v", err)
			}

			if count == 0 {
				s.lg.Debugf(log.Dev, "skipping namespace %s: no documents found", quotedNs)
				continue
			}

			if err = cursor.Close(ctx); err != nil {
				s.lg.Warnf(log.Dev, "error closing iterator: %v", err)
			}

			var indexes []bson.D

			// Index listing is not supported for views.
			if _, ok := nsViewPipelines[unquotedNs]; !ok {
				indexes, err = getIndexes(ctx, db, col, session)
				if err != nil {
					s.lg.Warnf(log.Dev, "error getting indexes: %v", err)
				}
			}

			jsonSchema.AddIndexes(indexes)
			jsonSchema.InferSpecialTypes()

			// 4. convert the JSON schema to a relational schema
			version, versionErr := session.Version()
			if versionErr != nil {
				return nil, fmt.Errorf(
					"failed to obtain MongoDB server version during call to ReadSchema: %s",
					versionErr,
				)
			}

			err = mapping.Map(mapping.NewSchemaMappingConfig(
				sampledDB,
				jsonSchema,
				col,
				s.cfg.PreJoin(),
				uuidSubtype3Encoding,
				version.VersionArray,
				s.lg,
				s.cfg.SchemaMappingMode(),
				s.cfg.MaxNumColumnsPerTable(),
				s.cfg.MaxNestedTableDepth(),
			))

			if err != nil {
				return nil, fmt.Errorf("error mapping schema: %v", err)
			}
			// Mapping a schema can cause us to create significant amounts of garbage so we
			// block and allow the GC to complete before proceeding.
			runtime.GC()

			sampledNamespaceCount++
			addSampledNamespace(db, col)
			s.lg.Debugf(log.Dev, "finished mapping schema for namespace %s", quotedNs)
		}

		if len(sampledDB.Tables()) != 0 {
			sampledDatabases = append(sampledDatabases, sampledDB)
		}
	}

	if sampledNamespaceCount != 0 {
		nsStr := strutil.Pluralize(sampledNamespaceCount, "namespace", "namespaces")
		var dbList []string
		for db, cols := range sampledNamespaces {
			var ns []string
			for _, c := range cols {
				ns = append(ns, fmt.Sprintf("%q", c))
			}
			dbList = append(dbList, fmt.Sprintf("%q (%v): [%s]", db, len(cols), strings.Join(ns, ", ")))
		}
		s.lg.Infof(log.Always, "mapped schema for %v %v: %v", sampledNamespaceCount, nsStr, strings.Join(dbList, "; "))
	} else {
		s.lg.Infof(log.Always, "no namespaces were sampled")
	}

	var sampledSchema *schema.Schema
	sampledSchema, err = schema.New(sampledDatabases, nil)
	return sampledSchema, err
}
