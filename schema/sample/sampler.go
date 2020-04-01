package sample

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mongodb/provider"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/mapping"
	"github.com/10gen/sqlproxy/schema/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
)

// Sampler is a type that provides schema sampling functionality.
type Sampler struct {
	cfg Config
	lg  log.Logger
	sp  *provider.SessionProvider
}

// NewSampler returns a new Sampler with the provided Config, Logger, and
// SessionProvider.
func NewSampler(cfg Config, lg log.Logger, sp *provider.SessionProvider) Sampler {
	component := fmt.Sprintf("%-10v [sampler]", log.SchemaComponent)
	lg = log.NewComponentLogger(component, lg)
	return Sampler{
		cfg: cfg,
		lg:  lg,
		sp:  sp,
	}
}

// Sample samples collections. In write mode, it samples collections based on
// their jsonSchema validators and indexes. In other modes it samples based on
// the data in collections.
func (s Sampler) Sample(ctx context.Context) (*schema.Schema, error) {
	if s.cfg.WriteMode() {
		return s.writeModeSample(ctx)
	}
	return s.readModeSample(ctx)
}

func (s Sampler) writeModeSample(ctx context.Context) (*schema.Schema, error) {
	session, err := s.sp.AuthenticatedAdminSessionPrimary()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	formatTableNames := func(tables []*schema.Table) string {
		names := make([]string, len(tables))
		for i, t := range tables {
			names[i] = fmt.Sprintf(`"%s"`, t.SQLName())
		}
		return "[" + strings.Join(names, ", ") + "]"
	}

	s.lg.Infof(log.Always, "sampling MongoDB for schema...")

	dbs, err := session.ListDatabases(ctx)
	if err != nil {
		return nil, err
	}
	databases := make([]*schema.Database, 0, len(dbs.Databases))
	for _, db := range dbs.Databases {
		tables, err := s.getWriteModeTables(ctx, session, db.Name)
		if err != nil {
			return nil, err
		}
		if len(tables) == 0 {
			s.lg.Infof(log.Always, `skipping database "%s" because no conforming jsonSchema validators exist`, db.Name)
			continue
		}
		database := schema.NewDatabase(s.lg, db.Name, tables, false)
		databases = append(databases, database)
		s.lg.Infof(log.Always, `mapped schema for %d namespaces:  "%s": %s`,
			len(tables), db.Name, formatTableNames(tables))
	}
	s.lg.Infof(log.Always, "done mapping")

	return schema.New(databases, false)
}

func (s Sampler) getWriteModeTables(ctx context.Context, session *mongodb.Session, db string) ([]*schema.Table, error) {
	collectionCursor, err := session.ListCollections(ctx, db, driver.CursorOptions{})
	if err != nil {
		return nil, fmt.Errorf("can't get the collections for '%v': %v", db, err)
	}

	tables := []*schema.Table{}
	type collectionResultType struct {
		Name    string `bson:"name"`
		Options bson.D `bson:"options"`
	}

	collectionResult := &collectionResultType{}
	for collectionCursor.Next(ctx, collectionResult) {
		jsSchema, ok := getValidator(collectionResult.Options)
		if !ok {
			continue
		}
		indexes, err := getIndexesInfo(ctx, session, db, collectionResult.Name)
		if err != nil {
			s.lg.Infof(log.Always, `could not get serialized table indexes for collection '%s': '%s'`, collectionResult.Name, err)
			continue
		}
		table, err := s.deserializeTableSchema(collectionResult.Name, jsSchema, indexes)
		if err != nil {
			s.lg.Infof(log.Always, "could not deserialize table schema for collection '%s': '%s'", collectionResult.Name, err)
			continue
		}
		tables = append(tables, table)
		collectionResult = &collectionResultType{}
	}
	return tables, nil
}

func getValidator(options bson.D) (bson.D, bool) {
	for _, field := range options {
		if field.Key == "validator" {
			doc, ok := field.Value.(bson.D)
			if !ok {
				return bson.D{}, false
			}
			for _, innerField := range doc {
				if innerField.Key == "$jsonSchema" {
					return innerField.Value.(bson.D), true
				}
			}
		}
	}
	return bson.D{}, false
}

type indexInfo struct {
	Name    string `bson:"name"`
	Unique  bool   `bson:"unique"`
	Key     bson.D `bson:"key"`
	Weights bson.D `bson:"weights"`
}

func getIndexesInfo(ctx context.Context, session *mongodb.Session, db, col string) ([]indexInfo, error) {
	indexes := []indexInfo{}
	indexCursor, err := session.ListIndexes(ctx, db, col)
	if err != nil {
		return indexes, err
	}
	var indexResult indexInfo
	for indexCursor.Next(ctx, &indexResult) {
		// We want to skip the _id_ index, as this is opaque
		// to write mode semantics at this time.
		if indexResult.Name == "_id_" {
			continue
		}
		indexes = append(indexes, indexResult)
		indexResult = indexInfo{}
	}
	return indexes, nil
}

func (s Sampler) deserializeTableSchema(name string, jsonSchema bson.D, indexes []indexInfo) (*schema.Table, error) {
	columns, comment, err := deserializeColumnsAndComment(jsonSchema)
	if err != nil {
		return nil, err
	}
	deserializedIndexes := deserializeIndexesInfo(indexes)
	table, err := schema.NewTable(s.lg, name, name, nil, columns, deserializedIndexes, comment, false)
	if err != nil {
		return nil, err
	}
	return table, nil
}

func deserializeColumnsAndComment(jsonSchema bson.D) ([]*schema.Column, option.String, error) {
	var required []string
	var properties bson.D
	comment := option.NoneString()

	for _, field := range jsonSchema {
		switch field.Key {
		case "required":
			requiredInterfaces, ok := field.Value.(primitive.A)
			if !ok {
				return nil, option.NoneString(), fmt.Errorf("jsonSchema 'required' must be a primitive.A, not %T", field.Value)
			}
			required = make([]string, len(requiredInterfaces))
			for i, s := range requiredInterfaces {
				required[i], ok = s.(string)
				if !ok {
					return nil, option.NoneString(), fmt.Errorf("jsonSchema 'required' elements must be a string, not %T", s)
				}
			}
		case "description":
			commentStr, ok := field.Value.(string)
			if !ok {
				return nil, option.NoneString(), fmt.Errorf("jsonSchema 'description' must be a string, not %T", field.Value)
			}
			comment = option.SomeString(commentStr)
		case "properties":
			var ok bool
			properties, ok = field.Value.(bson.D)
			if !ok {
				return nil, option.NoneString(), fmt.Errorf("jsonSchema 'properties' must be a bson.D, not %T", field.Value)
			}
		case "bsonType":
			if field.Value != "object" {
				return nil, option.NoneString(), fmt.Errorf("jsonSchema must have bsonType 'object'")
			}
		}
	}
	if len(properties) < 1 {
		return nil, option.NoneString(), fmt.Errorf("jsonSchema must have at least one property for writeMode schema mapping")
	}
	if len(required) != len(properties) {
		return nil, option.NoneString(), fmt.Errorf("all properties must be required")
	}
	columns := make([]*schema.Column, len(required))
	requiredSet := strutil.StringSliceToSet(required)
	for i, property := range properties {
		col, err := deserializeSchemaColumn(required, requiredSet, property)
		if err != nil {
			return nil, option.NoneString(), err
		}
		columns[i] = col
	}
	return columns, comment, nil
}

func deserializeSchemaColumn(required []string,
	requiredSet map[string]struct{},
	property bson.E) (*schema.Column, error) {
	if _, ok := requiredSet[property.Key]; !ok {
		return nil, fmt.Errorf("found property named '%s' that is not"+
			" in the required properties: %#v", property.Key, required)
	}
	doc, ok := property.Value.(bson.D)
	if !ok {
		return nil, fmt.Errorf("property must have a bsonType object for"+
			" its value, found %T", property.Value)
	}
	sqlType := schema.SQLPolymorphic
	mongoType := schema.MongoNone
	comment := option.NoneString()
	null := false
	var err error
	getTypes := func(input interface{}) (schema.SQLType, schema.MongoType, error) {
		var bsonType string
		bsonType, ok = input.(string)
		if !ok {
			return sqlType, mongoType,
				fmt.Errorf("bsonType must be a string denoting"+
					" the type, found %T", input)
		}
		return schema.GetSQLTypeAndMongoTypeFromJSONSchemaType(bsonType)
	}
	for _, e := range doc {
		switch e.Key {
		case "description":
			var commentStr string
			commentStr, ok = e.Value.(string)
			if !ok {
				return nil, fmt.Errorf("property description must have"+
					" type string, found %T", e.Value)
			}
			comment = option.SomeString(commentStr)
		case "bsonType":
			sqlType, mongoType, err = getTypes(e.Value)
			if err != nil {
				return nil, err
			}
		case "oneOf":
			var docArr []interface{}
			docArr, ok = e.Value.(primitive.A)
			if !ok {
				return nil, fmt.Errorf("'oneOf' argument must be a primitive.A, found %T", e.Value)
			}
			if len(docArr) != 2 {
				return nil, fmt.Errorf("the only supported value for 'oneOf' is a simple bsonType with null, "+
					"thus there should be two elements, not %d", len(docArr))
			}
			sqlTypes := make([]schema.SQLType, 2)
			mongoTypes := make([]schema.MongoType, 2)

			for i, oneOfDocInterface := range docArr {
				var oneOfDoc bson.D
				oneOfDoc, ok = oneOfDocInterface.(bson.D)
				if !ok {
					return nil, fmt.Errorf("the values under 'oneOf' must be documents not %T", oneOfDocInterface)
				}
				if len(oneOfDoc) != 1 {
					return nil, fmt.Errorf("the only supported values for 'oneOf' are 'bsonType', "+
						" which means documents of size 1, but found document with length %d", len(oneOfDoc))
				}
				if oneOfDoc[0].Key != "bsonType" {
					return nil, fmt.Errorf("the only supported values for 'oneOf' are "+
						"'bsonType', not '%s'", oneOfDoc[0].Key)
				}
				sqlTypes[i], mongoTypes[i], err = getTypes(oneOfDoc[0].Value)
				if err != nil {
					return nil, err
				}
			}
			if sqlTypes[0] == schema.SQLNull {
				null = true
				sqlType, mongoType = sqlTypes[1], mongoTypes[1]
			} else if sqlTypes[1] == schema.SQLNull {
				null = true
				sqlType, mongoType = sqlTypes[0], mongoTypes[0]
			} else {
				return nil, fmt.Errorf("'oneOf' is only supported when one of the bsonTypes"+
					" is 'null', but found '%s' and '%s'", sqlTypes[0], sqlTypes[1])
			}
		}
	}
	return schema.NewColumn(strings.ToLower(property.Key), sqlType, property.Key, mongoType, null, comment), nil
}

func deserializeIndexesInfo(indexes []indexInfo) []schema.Index {
	ret := make([]schema.Index, len(indexes))
	for i, info := range indexes {
		fullText := false
		// If weights exists, this represents a full text info, and the
		// weights should be used as the key. This is a quirk of mongo
		// full text indexes.
		if len(info.Weights) != 0 {
			fullText = true
			info.Key = info.Weights
		}
		parts := deserializeIndexParts(info.Key)
		ret[i] = schema.NewIndex(info.Name, info.Unique, fullText, parts)
	}
	return ret
}

func deserializeIndexParts(indexKey bson.D) []schema.IndexPart {
	indexParts := make([]schema.IndexPart, len(indexKey))
	for i, part := range indexKey {
		switch typedDirection := part.Value.(type) {
		case int:
			indexParts[i] = schema.NewIndexPart(part.Key, typedDirection)
		case int32:
			indexParts[i] = schema.NewIndexPart(part.Key, int(typedDirection))
		case int64:
			indexParts[i] = schema.NewIndexPart(part.Key, int(typedDirection))
		case float64:
			indexParts[i] = schema.NewIndexPart(part.Key, int(typedDirection))
		default:
			panic(fmt.Sprintf("found unknown direction type %T", part.Value))
		}
	}
	return indexParts
}

// Sample samples the MongoDB namespaces indicated by the sampler config. It
// returns the relational schema generated from the sampled schema.
func (s Sampler) readModeSample(ctx context.Context) (*schema.Schema, error) {
	session, err := s.sp.AuthenticatedAdminSessionPrimary()
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

	fetchedNSes, err := fetchSortedNamespaces(ctx, session, s.lg, nsMatcher)
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

	currentNumTables := int64(0)
	for _, ns := range fetchedNSes {
		db, collections := ns.Database, ns.Collections
		if _, ok := dbSampleBlacklist[db]; ok {
			s.lg.Debugf(log.Dev, "skipping %q database", db)
			continue
		}

		sampledDB := schema.NewDatabase(s.lg, db, nil, false)

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
			if currentNumTables >= s.cfg.MaxNumGlobalTables() {
				s.lg.Infof(log.Always, fmt.Sprintf("max num global tables (%d) reached: not mapping any more tables", s.cfg.MaxNumGlobalTables()))
				break
			}

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

			ft := mongo.NewSchemaFieldTracker(s.cfg.MaxNumFieldsPerCollection(), 0, s.lg, sampleCollection)
			for cursor.Next(ctx, &doc) {
				err = jsonSchema.IncludeSample(doc, ft)
				if err != nil {
					return nil, fmt.Errorf("error including sample: %v", err)
				}
				doc = doc[:]
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
				&currentNumTables,
				s.cfg.MaxNestedTableDepth(),
				s.cfg.MaxNumTablesPerCollection(),
				s.cfg.MaxNumGlobalTables(),
				false,
			))

			if err != nil {
				return nil, fmt.Errorf("error mapping schema: %v", err)
			}

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
	sampledSchema, err = schema.New(sampledDatabases, false)
	return sampledSchema, err
}
