package mapping

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/mongo"
)

const renameSeparator = "_DOT_"

// SchemaMappingConfig holds all the configuration
// necessary to perform schema mapping.
type SchemaMappingConfig struct {
	CollectionName        string
	Database              *schema.Database
	mode                  config.MappingMode
	Logger                log.Logger
	MaxNestedTableDepth   int64
	MaxNumColumnsPerTable int64
	PreJoin               bool
	Schema                *mongo.Schema
	UUIDSubtype3Encoding  string
	Version               []uint8
}

// NewSchemaMappingConfig is a constructor that builds
// SchemaMappingConfig structs.
func NewSchemaMappingConfig(
	database *schema.Database,
	schema *mongo.Schema,
	collectionName string,
	preJoin bool,
	uuidSubtype3Encoding string,
	version []uint8,
	logger log.Logger,
	mode config.MappingMode,
	maxNumColumnsPerTable int64,
	maxNestedTableDepth int64,
) SchemaMappingConfig {
	return SchemaMappingConfig{
		Database:              database,
		Schema:                schema,
		CollectionName:        collectionName,
		PreJoin:               preJoin,
		UUIDSubtype3Encoding:  uuidSubtype3Encoding,
		Version:               version,
		Logger:                logger,
		mode:                  mode,
		MaxNumColumnsPerTable: maxNumColumnsPerTable,
		MaxNestedTableDepth:   maxNestedTableDepth,
	}
}

// namedSchema stores a schema and the property name associated with it.
// A map could not be used because ordering is important.
type namedSchema struct {
	name   string
	schema *mongo.Schema
}

// getTypeNames returns a slice of all the scalar type names from a Schemata.
func getTypeNames(sc *mongo.Schemata) ([]mongo.BSONType, error) {
	ret := make([]mongo.BSONType, 0)
	for ty := range sc.Schemas {
		// Verify that the JSON schema key is a valid BSON type.
		if !mongo.IsValidType(ty) {
			return nil, fmt.Errorf("found invalid json schema key: %v", ty)
		}
		ret = append(ret, ty)
	}
	return ret, nil
}

// Map takes a mongo schema that describes a collection with the provided name
// and creates a set of tables in the Database that comprise a relational
// equivalent of that schema. If preJoined is true, the tables generated for
// array fields will include parent fields, effectively resulting in pre-joined
// tables.
func Map(cfg SchemaMappingConfig) error {

	// create the table into which we will map this collection's fields.
	// this table has the same name as the collection it is mapped from.
	// unless we have array fields, this is the only table we will create.
	t, err := schema.NewTable(cfg.Logger, cfg.CollectionName, cfg.CollectionName, nil, nil)
	if err != nil {
		return err
	}

	mongoNames := make(map[*schema.Table]map[string]string)
	mongoNamePrefixes := make(map[*schema.Table]map[string]string)
	seenFields := make(map[*schema.Table][]string)
	uniqueColumns := make(map[*schema.Table]map[string]struct{})
	uniqueFields := make(map[*schema.Table]map[string]struct{})

	mongoNames[t] = make(map[string]string)
	mongoNamePrefixes[t] = make(map[string]string)
	seenFields[t] = make([]string, 0)
	uniqueColumns[t] = make(map[string]struct{})
	uniqueFields[t] = make(map[string]struct{})
	// initialize the top-level mapping context
	ctx := newMappingContext(cfg.Logger,
		cfg.Database,
		t,
		cfg.UUIDSubtype3Encoding,
		procutil.VersionAtLeast(cfg.Version, []uint8{3, 4, 0}),
		mongoNames,
		mongoNamePrefixes,
		seenFields,
		uniqueColumns,
		uniqueFields,
		GetHeuristic(cfg.mode),
		"",
		false,
		false,
		-1,
		cfg.MaxNumColumnsPerTable,
		cfg.MaxNestedTableDepth,
	)

	// map the collection schema to a relational schema
	err = ctx.mapObjectSchema(cfg.Schema)
	if err != nil {
		return err
	}

	// add this table to the database
	cfg.Database.AddTable(cfg.Logger, t)

	// post-process the database
	cfg.Database.PostProcess(cfg.Logger, cfg.PreJoin)

	// validate the db schema
	err = cfg.Database.Validate()
	if err != nil {
		return err
	}

	cfg.Logger.Debugf(log.Dev, "mapped new table %q", cfg.CollectionName)
	return nil
}

// mappingContext maintains state that describes the context into which a mongo
// schema should be mapped.
type mappingContext struct {

	// logger is the logger used to output warnings and other information during
	// the mapping process
	logger log.Logger

	// db is the database into which we are mapping the current schema.
	// this will never be changed by any mapping functions.
	db *schema.Database

	// table is the table into which we are mapping the current schema.
	// this will change whenever we need to map an array field.
	table *schema.Table

	// path is the path of the field being mapped, relative to the collection
	// to which it belongs. A field's path is comprised of all the object
	// property names followed when accessing the field in a document.
	path string

	// uuidSubtype3Encoding is the encoding used to store UUID subtype 3 values.
	// It is used to drive how such values are decoded from MongoDB.
	uuidSubtype3Encoding string

	// inPrimaryKey tracks whether the current path is "in" the primary key.
	// this should be true whenever the path begins with "_id" (false otherwise)
	inPrimaryKey bool

	// hasConflict tracks whether there is a conflict in or above the current
	// context.
	hasConflict bool

	// heuristic is the heuristic function used to determine the dominant
	// Schema(s) for a Schemata.
	heuristic SchemataHeuristic

	// nestedArrayDepth tracks how many levels deep in a nested array this
	// context is. For non-array contexts and top-level array contexts, its
	// value should be zero. It should be incremented once for each level of
	// array nesting inside the top-level array context.
	nestedArrayDepth int

	// isAtLeastVersion34 is true if the MongoDB server version is >= 3.4.0.
	// This, in particular, means the server has the $type expression and
	// $addFields pipeline stage.
	isAtLeastVersion34 bool

	// mongoNames is a mapping from sqlColumn name to underlying mongo
	// field name, for fields that have been directly remapped, versus
	// prefixes having been renamed, which is covered below.
	mongoNames map[*schema.Table]map[string]string

	// mongoNamePrefixes is a mapping from sqlColumn name prefixes to
	// underlying mongo field name prefixes. This allows us to recover the
	// proper mongo name when we have a conflict above more deeply nested fields.
	mongoNamePrefixes map[*schema.Table]map[string]string

	// seenFields is a slice of fieldNames in the order they are traversed (depth first,
	// left to right). This order guarantees us that a path prefix will always exist
	// before the extension of that prefix, meaning that we can check for
	// prefix collisions in projects without sorting (only needed for server
	// version 3.2 where we are forced to use $project).
	seenFields map[*schema.Table][]string

	// uniqueColumns is the unique set of columns for a table so that we do
	// not map a column twice, which can happen with nested arrays.
	uniqueColumns map[*schema.Table]map[string]struct{}

	// uniqueFields is a set of all uniqueField names without a specified order,
	// so that we do not introduce field names that are already in use.
	uniqueFields map[*schema.Table]map[string]struct{}

	// maxNumColumnsPerTable is the maximum number of columns that will be mapped for any
	// given table in the schema.
	maxNumColumnsPerTable int64

	// maxNestedTableDepth is the maximum number of nested tables that will be mapped for any
	// given heritage within a table.
	maxNestedTableDepth int64
}

// newMappingContext constructs a new mappingContext.
func newMappingContext(logger log.Logger,
	db *schema.Database,
	table *schema.Table,
	uuidSubtype3Encoding string,
	isAtLeastVersion34 bool,
	mongoNames map[*schema.Table]map[string]string,
	mongoNamePrefixes map[*schema.Table]map[string]string,
	seenFields map[*schema.Table][]string,
	uniqueColumns map[*schema.Table]map[string]struct{},
	uniqueFields map[*schema.Table]map[string]struct{},
	heuristic SchemataHeuristic,
	path string,
	inPrimaryKey bool,
	hasConflict bool,
	nestedArrayDepth int,
	maxNumColumnsPerTable int64,
	maxNestedTableDepth int64,
) *mappingContext {
	return &mappingContext{
		logger:                logger,
		db:                    db,
		table:                 table,
		uuidSubtype3Encoding:  uuidSubtype3Encoding,
		isAtLeastVersion34:    isAtLeastVersion34,
		mongoNames:            mongoNames,
		mongoNamePrefixes:     mongoNamePrefixes,
		seenFields:            seenFields,
		uniqueColumns:         uniqueColumns,
		uniqueFields:          uniqueFields,
		heuristic:             heuristic,
		path:                  path,
		inPrimaryKey:          inPrimaryKey,
		hasConflict:           hasConflict,
		nestedArrayDepth:      nestedArrayDepth,
		maxNumColumnsPerTable: maxNumColumnsPerTable,
		maxNestedTableDepth:   maxNestedTableDepth,
	}
}

// addTransitiveProjects adds all the necessary field preservations from previous
// projects. This is only necessary in MongoDB 3.2 where there is no access
// to $addFields.
func (ctx *mappingContext) addTransitiveProjections(projectedFields map[string]struct{},
	projectBody bson.D) bson.D {
OUTER:
	// seenFields are already sorted in a partial order of lowest
	// depth to greatest because they are constructed in dfs order,
	// this makes prefix checking linear.
	for i := len(ctx.seenFields[ctx.table]) - 1; i >= 0; i-- {
		field := ctx.seenFields[ctx.table][i]
		// The field may have been renamed after it was added to seenFields.
		// This may result in use checking the same field twice, but is cheaper
		// than the book keeping necessary to avoid that, since it will be immediately
		// seen in the projectedFields map.
		if renamedField, ok := ctx.mongoNames[ctx.table][field]; ok {
			field = renamedField
		} else if renamedField, ok := ctx.mongoNamePrefixes[ctx.table][field]; ok {
			field = renamedField
		}
		// Only check if this is a prefix of already projected fields if we are not
		// already projecting this field.
		if _, ok := projectedFields[field]; !ok {
			for projectedField := range projectedFields {
				if bsonutil.IsPrefix(projectedField, field) {
					continue OUTER
				}
			}
			projectBody = append(projectBody, bsonutil.NewDocElem(field, true))
			projectedFields[field] = struct{}{}
		}
	}
	return projectBody
}

// getUniqueFieldName gets a unique flattened field name for a field that
// is projected out of an object in an object/non-object conflict. The prefix
// is the path to the top of the field, while the name is the name of the
// sub-field in the object.
func (ctx *mappingContext) getUniqueFieldName(prefix, name string) string {
	initial := prefix + renameSeparator + name
	current := initial
	for i := 0; ; i++ {
		if _, ok := ctx.uniqueFields[ctx.table][current]; !ok {
			ctx.uniqueFields[ctx.table][current] = struct{}{}
			return current
		}
		current = initial + "_" + strconv.Itoa(i)
	}
}

// getProjectAndSchemasForProperties returns the necessary project and schemas that
// result from the properties in an object schema. The returned $project/$addFields
// is only needed if there is an object/non-object conflict in the items of an array.
func (ctx *mappingContext) getProjectAndSchemasForProperties(js *mongo.Schema,
	props []string) (project bson.D, namedSchemas []namedSchema) {

	namedSchemas = []namedSchema{}
	projectBody := bsonutil.NewD()

	// projectedFields keeps track of already projectedFields so that
	// we do not project a field twice.
	projectedFields := make(map[string]struct{})
	for _, prop := range props {
		dottedCtx := ctx.withSubpath(prop)
		mongoRenamedPrefixPath := dottedCtx.path
		// If this path has been renamed in a previous $addFields, make sure
		// to reference the new name for the projection, but maintain the old
		// name for the column name.
		if renamedPath, ok := ctx.mongoNames[ctx.table][mongoRenamedPrefixPath]; ok {
			mongoRenamedPrefixPath = renamedPath
		}
		s := dottedCtx.getDominantSchemas(js.Properties[prop])
		namedSchemas = append(namedSchemas, namedSchema{name: prop, schema: s[0]})
		if len(s) == 1 {
			// If we are under an array context, which we determine with nestedArrayDepth > -1,
			// and there was a remapping above this path, we need to project this property
			// out of the embedded document to use the same field as the previous
			// remapping, as that is the Mongo name used in the Column of the table.
			// If the remapped field already has a non-NULLish value, we do not want
			// to project over it, this allows us to refer to the same path in multiple
			// array unwinds, e.g.: for collection COL containing:
			// {a: [[{b: 1}]]}
			// {a: [{b:2}]}
			// both 1 and 2 should appear in the same column, `a.b` of table COL_a.
			//
			// Adding this even in non array contexts is correct, checking the nestingArrayDepth
			// is only a performance optimization.
			if ctx.nestedArrayDepth > -1 && mongoRenamedPrefixPath != dottedCtx.path {
				previousField := "$" + mongoRenamedPrefixPath
				unmappedField := "$" + dottedCtx.path
				projectBody = append(projectBody, bsonutil.NewDocElem(mongoRenamedPrefixPath,
					bsonutil.WrapInCond(
						previousField,
						// We will return the value of the unmapped field as long as that
						// value is not an array. If it is an array, it belongs in another
						// descendent table rather than this table.
						ctx.buildIfNotArray(unmappedField),
						bsonutil.WrapInBinOp(bsonutil.OpGt, previousField, nil),
					)),
				)
				// Add to projectedFields so that we do not project the same field
				// twice on MongoDB server 3.2.
				projectedFields[mongoRenamedPrefixPath] = struct{}{}
			}
			continue
		}
		ctx.hasConflict = true
		// We have at least one object conflict, return needed = true meaning that
		// we have to add a $project/$addFields.
		namedSchemas = append(namedSchemas, namedSchema{name: prop, schema: s[1]})
		// Add non-Object to $project.
		projectBody = append(projectBody, bsonutil.NewDocElem(mongoRenamedPrefixPath,
			ctx.buildIfNotObject("$"+mongoRenamedPrefixPath)))
		projectedFields[mongoRenamedPrefixPath] = struct{}{}
		objectSchema := s[1]
		for name := range objectSchema.Properties {
			preimage := mongoRenamedPrefixPath + "." + name
			image := ctx.getUniqueFieldName(mongoRenamedPrefixPath, name)
			colName := dottedCtx.withSubpath(name).path
			ctx.mongoNames[ctx.table][colName] = image
			ctx.mongoNamePrefixes[ctx.table][colName] = image
			projectBody = append(projectBody, bsonutil.NewDocElem(image,
				ctx.buildIfObject("$"+mongoRenamedPrefixPath, "$"+preimage)))
			ctx.seenFields[ctx.table] = append(ctx.seenFields[ctx.table], image)
			projectedFields[image] = struct{}{}
		}
	}
	if len(projectBody) == 0 {
		return nil, namedSchemas
	}
	if ctx.isAtLeastVersion34 {
		project = bsonutil.NewD(bsonutil.NewDocElem("$addFields", projectBody))
		return project, namedSchemas
	}

	projectBody = ctx.addTransitiveProjections(projectedFields, projectBody)
	project = bsonutil.NewD(bsonutil.NewDocElem("$project", projectBody))
	return project, namedSchemas
}

// getProjectAndSchemaForItems returns the necessary project and schema that
// result from the items in an array schema. Needed tells us if the project
// is actually needed, which only occurs if there is an object/nonObject
// conflict in the sampled array items. This differs primarily from
// getProjectAndSchemasForProperties in that there is only one schema returned,
// rather than one per property, and that we have to be careful to treat
// the $unwind index that comes from arrays, which is passed as the indexName
// argument.
func (ctx *mappingContext) getProjectAndSchemaForItems(items *mongo.Schemata,
	indexName string) (project bson.D, schemas []*mongo.Schema) {

	project = bson.D(nil)
	schemas = ctx.getDominantSchemas(items)
	if len(schemas) == 1 {
		return project, schemas
	}
	ctx.hasConflict = true
	// projectedFields keeps track of already projectedFields so that
	// we do not project a field twice.
	projectedFields := make(map[string]struct{})
	var projectBody bson.D
	if ctx.isAtLeastVersion34 {
		projectBody = bsonutil.NewD()
	} else {
		// If we have to use $project instead of $addFields, make
		// sure not to drop the indexName.
		projectBody = bsonutil.NewD(bsonutil.NewDocElem(indexName, true))
		projectedFields[indexName] = struct{}{}
	}
	// Add nonObject to project.
	mongoRenamedPrefixPath := ctx.path
	if renamedPath, ok := ctx.mongoNames[ctx.table][mongoRenamedPrefixPath]; ok {
		mongoRenamedPrefixPath = renamedPath
	}
	projectBody = append(projectBody, bsonutil.NewDocElem(mongoRenamedPrefixPath,
		ctx.buildIfNotObject("$"+mongoRenamedPrefixPath)))
	projectedFields[mongoRenamedPrefixPath] = struct{}{}
	objectSchema := schemas[1]
	for name := range objectSchema.Properties {
		preimage := mongoRenamedPrefixPath + "." + name
		image := ctx.getUniqueFieldName(mongoRenamedPrefixPath, name)
		colName := ctx.withSubpath(name).path
		ctx.mongoNames[ctx.table][colName] = image
		ctx.mongoNamePrefixes[ctx.table][colName] = image
		projectBody = append(projectBody, bsonutil.NewDocElem(image,
			ctx.buildIfObject("$"+mongoRenamedPrefixPath, "$"+preimage)))
		ctx.seenFields[ctx.table] = append(ctx.seenFields[ctx.table], image)
		projectedFields[image] = struct{}{}
	}
	if ctx.isAtLeastVersion34 {
		project = bsonutil.NewD(bsonutil.NewDocElem("$addFields", projectBody))
		return project, schemas
	}

	projectBody = ctx.addTransitiveProjections(projectedFields, projectBody)
	project = bsonutil.NewD(bsonutil.NewDocElem("$project", projectBody))
	return project, schemas
}

/*
 * The following functions all take a mongo schema of a particular type and map
 * it into a relational table (or tables).
 */

// mapObjectSchema maps the provided object schema into a mappingContext.
func (ctx *mappingContext) mapObjectSchema(js *mongo.Schema) error {
	// order the props alphabetically
	props := make([]string, 0, len(js.Properties))

	for prop := range js.Properties {
		props = append(props, prop)
	}

	sort.Slice(props, func(i, j int) bool {
		// To cater to cases where we might have mixed case properties
		// within the context of an object's mapping, we sort the
		// properties in descending order of length.
		//
		// This avoids collisions that are possible when an existing
		// field name is the same as what we might map a mixed case
		// property to.
		//
		// For example, if the properties we are mapping are:
		//
		// "C", "c", "c_0"
		//
		// and we used an ascending sort, we would proceed thus:
		//
		// "C" => "C"
		// "c" => "c_0" (we haven't seen c_0 yet)
		// "c_0" => should error (we shouldn't rename existing fields)
		//
		// By sorting in descending order, we remove the possibility
		// of collisions when there are mixed case keys in users' data.
		iLen, jLen := len(props[i]), len(props[j])
		if iLen == jLen {
			return props[i] > props[j]
		}
		return iLen > jLen
	})

	// Add every property of this object to ctx.seenFields.
	for _, prop := range props {
		subPath := ctx.withSubpath(prop).path
		ctx.seenFields[ctx.table] = append(ctx.seenFields[ctx.table], subPath)
		ctx.uniqueFields[ctx.table][subPath] = struct{}{}
	}

	// Get the dominant schemas and a project, if necessary.
	project, namedSchemas := ctx.getProjectAndSchemasForProperties(js, props)
	// If we need the conflict project, add it.
	if project != nil {
		err := ctx.table.AddPipelineStage(project)
		if err != nil {
			return err
		}
	}
	// Map each schema.
	for _, namedSchema := range namedSchemas {
		schema := namedSchema.schema
		name := namedSchema.name

		switch schema.BSONType {
		case mongo.Object:
			err := ctx.objectContext(name).mapObjectSchema(schema)
			if err != nil {
				return err
			}

		case mongo.Array:
			if schema.SpecialType != mongo.GeoPoint {
				subctx, err := ctx.arrayContext(name)
				if err != nil {
					return err
				}
				if subctx == nil {
					continue
				}
				err = subctx.mapArraySchema(schema)
				if err != nil {
					return err
				}
				break
			}

			// if this is a geo.2darray, treat it as a scalar, by falling
			// through to scalar case
			fallthrough
		default: // scalar
			sampleTypes, err := getTypeNames(js.Properties[namedSchema.name])
			if err != nil {
				return err
			}
			err = ctx.scalarContext(name).mapScalarSchema(schema, sampleTypes)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ctx *mappingContext) buildIfNotObject(v interface{}) interface{} {
	if ctx.isAtLeastVersion34 {
		return bsonutil.WrapInCond(
			nil, v, bsonutil.WrapInBinOp(bsonutil.OpEq, bsonutil.WrapInType(v), "object"))
	}
	// $type does not exist in MongoDB 3.2, use type bracketing to figure out if this
	// is an object or not.
	cond := bsonutil.WrapInBinOp(bsonutil.OpOr,
		bsonutil.WrapInBinOp(bsonutil.OpLt, v, bsonutil.NewD()),
		bsonutil.WrapInBinOp(bsonutil.OpGte, v, bsonutil.NewArray()),
	)
	return bsonutil.WrapInCond(v, nil, cond)
}

func (ctx *mappingContext) buildIfObject(v interface{}, subV interface{}) interface{} {
	if ctx.isAtLeastVersion34 {
		return bsonutil.WrapInCond(
			subV, nil, bsonutil.WrapInBinOp(bsonutil.OpEq, bsonutil.WrapInType(v), "object"))
	}
	// $type does not exist in MongoDB 3.2, use type bracketing to figure out if this
	// is an object or not.
	cond := bsonutil.WrapInBinOp(bsonutil.OpAnd,
		bsonutil.WrapInBinOp(bsonutil.OpGte, v, bsonutil.NewD()),
		bsonutil.WrapInBinOp(bsonutil.OpLt, v, bsonutil.NewArray()),
	)
	return bsonutil.WrapInCond(subV, nil, cond)
}

func (ctx *mappingContext) buildIfNotArray(v interface{}) interface{} {

	if ctx.isAtLeastVersion34 {
		return bsonutil.WrapInCond(
			v, nil, bsonutil.WrapInBinOp(bsonutil.OpNeq, bsonutil.WrapInType(v), "array"))
	}
	// $type does not exist in MongoDB 3.2, use type bracketing to figure out if this
	// is an object or not.
	cond := bsonutil.WrapInBinOp(bsonutil.OpOr,
		bsonutil.WrapInBinOp(bsonutil.OpLt, v, bsonutil.NewArray()),
		// We would really like to use bson.Binary here instead of false,
		// but the go driver doesn't support bson.Binary on MongoDB server 3.2.
		bsonutil.WrapInBinOp(bsonutil.OpGte, v, false),
	)
	return bsonutil.WrapInCond(v, nil, cond)
}

// mapArraySchema maps the provided array schema into a mappingContext.
func (ctx *mappingContext) mapArraySchema(js *mongo.Schema) error {

	// calculate the name of the array index column
	indexName := ctx.path + "_idx"
	if ctx.nestedArrayDepth > -1 {
		indexName += fmt.Sprintf("_%v", ctx.nestedArrayDepth+1)
	}

	// Get the dominant schemas and a project, if necessary.
	project, itemSchemas := ctx.getProjectAndSchemaForItems(js.Items, indexName)

	// Don't map null arrays unless there is an object conflict on this field.
	if len(itemSchemas) == 1 && mongo.IsUnmappableType(itemSchemas[0].BSONType) {
		return nil
	}
	ctx.seenFields[ctx.table] = append(ctx.seenFields[ctx.table], indexName)

	// create the array index column and add it to the current table
	col := schema.NewColumn(indexName, schema.SQLInt, indexName, schema.MongoInt)
	ctx.table.AddColumn(ctx.logger, col, true)

	path := ctx.path
	if renamedPath, ok := ctx.mongoNames[ctx.table][path]; ok {
		path = renamedPath
	}

	// If we have a conflict above or in the current context, we need to
	// filter out empty arrays before we add an unwind.
	if ctx.hasConflict {
		err := ctx.table.AddPipelineStage(bsonutil.NewD(
			bsonutil.NewDocElem("$match", bsonutil.NewD(
				bsonutil.NewDocElem(path, bsonutil.NewD(
					bsonutil.NewDocElem("$ne", bsonutil.NewArray()),
				)),
			)),
		))
		if err != nil {
			return err
		}
	}

	// add an unwind to the current table's pipeline. If there is a conflict
	// in or above the current context we need to preserveNullAndEmptyArrays
	unwind := bsonutil.NewD(
		bsonutil.NewDocElem("$unwind", bsonutil.NewD(
			bsonutil.NewDocElem("path", "$"+path),
			bsonutil.NewDocElem("includeArrayIndex", indexName),
			bsonutil.NewDocElem("preserveNullAndEmptyArrays", ctx.hasConflict),
		)),
	)

	err := ctx.table.AddPipelineStage(unwind)
	if err != nil {
		return err
	}
	// If there are two itemSchemas there was a conflict, so we must
	// add the project here to tease them out.
	if len(itemSchemas) == 2 {
		err := ctx.table.AddPipelineStage(project)
		if err != nil {
			return err
		}
	}

	// Map the array's elements. Note that in the presence of object conflicts,
	// there will be two schemas that both need be mapped, the $project/$addFields
	// was already handled.
	// We need a subcontext if this an a nested array (to track depth)
	// for objects and scalars, we continue to use the array's context.
	for _, itemSchema := range itemSchemas {
		var err error
		switch itemSchema.BSONType {
		case mongo.Array:
			err = ctx.nestedArrayContext().mapArraySchema(itemSchema)
		case mongo.Object:
			err = ctx.mapObjectSchema(itemSchema)
		default:
			var sampleTypes []mongo.BSONType
			sampleTypes, err = getTypeNames(js.Items)
			if err != nil {
				return err
			}
			err = ctx.mapScalarSchema(itemSchema, sampleTypes)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// mapScalarSchema maps the provided scalar schema into a mappingContext.
// The original mongo.Schemata is passed for describe table comments.
func (ctx *mappingContext) mapScalarSchema(js *mongo.Schema, sampleTypes []mongo.BSONType) error {
	if mongo.IsUnmappableType(js.BSONType) {
		ctx.logger.Warnf(log.Dev, "table %q, column %q has unsupported type %q, will not map",
			ctx.table.SQLName(), ctx.path, js.BSONType)
		return nil
	}
	if ctx.table.NumColumns() == ctx.maxNumColumnsPerTable {
		ctx.logger.Warnf(log.Dev,
			`cannot map path %q - table %q, has reached configured column limit "%v"`,
			ctx.path, ctx.table.SQLName(), ctx.maxNumColumnsPerTable)
		return nil
	}

	// We will map columns that are entirely null as strings.
	if js.BSONType == mongo.Null {
		ctx.logger.Warnf(log.Dev, "table %q, column %q has inferred type NULL: mapping as varchar",
			ctx.table.SQLName(), ctx.path)
		js.BSONType = mongo.String
	}

	// When columns exist with the same name at different unwinding
	// depths, we end up mapping the column twice. Go ahead and
	// skip the second instance.
	if _, ok := ctx.uniqueColumns[ctx.table][ctx.path]; ok {
		return nil
	}
	ctx.uniqueColumns[ctx.table][ctx.path] = struct{}{}

	mapToName, ok := ctx.mongoNames[ctx.table][ctx.path]
	if !ok {
		// Use the prefix path if we have not remapped this directly.
		mapToName = ctx.mongoNamePrefixes[ctx.table][ctx.path]
	}
	// Create a new column.
	col, err := newColumn(ctx.path, mapToName,
		js, ctx.uuidSubtype3Encoding, sampleTypes)
	if err != nil {
		return err
	}

	// if err and col are both nil, we don't map a column for this schema
	if col == nil {
		return nil
	}

	// add the column to the current table
	ctx.table.AddColumn(ctx.logger, col, ctx.inPrimaryKey)
	return nil
}

/*
 * The functions in this group all create subcontexts from an existing mapping
 * context. They are responsible for doing the bookkeeping that is required when
 * transitioning between contexts.
 * These functions are intended for use as helpers in the ctx.map*Schema
 * functions, and should probably not be used outside of that context.
 */

// scalarContext returns a new mappingContext whose path is equal to the
// current context's path joined to the provided subpath with a '.'
func (ctx *mappingContext) scalarContext(subpath string) *mappingContext {
	return ctx.withSubpath(subpath)
}

// objectContext returns a new mappingContext whose path is equal to the
// current context's path joined to the provided subpath with a '.'
func (ctx *mappingContext) objectContext(subpath string) *mappingContext {
	return ctx.withSubpath(subpath)
}

// arrayContext returns a new mappingContext whose table is a new table
// representing the array at the specified subpath. The new table's parent
// table is the current context's table.
func (ctx *mappingContext) arrayContext(subpath string) (*mappingContext, error) {
	newCtx := ctx.withSubpath(subpath)

	// find the root of the current table heritage
	root := newCtx.table
	depth := int64(0)
	for root.Parent() != nil {
		root, depth = root.Parent(), depth+1
	}

	if depth == ctx.maxNestedTableDepth {
		ctx.logger.Warnf(log.Dev, `cannot map path %q - field %q has reached configured nested table limit "%v"`,
			newCtx.path, root.SQLName(), ctx.maxNestedTableDepth)
		return nil, nil
	}

	// calculate the name for this array's table
	arrayTableName := root.SQLName() + "_" + strings.Replace(newCtx.path, ".", "_", -1)

	// create the array table; add it to newCtx.db and newCtx
	arrayTable, err := schema.NewTableWithUnwindPath(
		ctx.logger,
		arrayTableName,
		newCtx.table.MongoName(),
		nil,
		nil,
		newCtx.path,
	)

	if err != nil {
		return nil, err
	}

	// Copy the seenFields, uniqueFields, mongoNames, and mongoNamePrefixes from the parent table.
	ctx.mongoNames[arrayTable] = make(map[string]string, len(ctx.uniqueFields[ctx.table]))
	for col, field := range ctx.mongoNames[ctx.table] {
		ctx.mongoNames[arrayTable][col] = field
	}
	ctx.mongoNamePrefixes[arrayTable] = make(map[string]string, len(ctx.uniqueFields[ctx.table]))
	for col, field := range ctx.mongoNamePrefixes[ctx.table] {
		ctx.mongoNamePrefixes[arrayTable][col] = field
	}
	ctx.seenFields[arrayTable] = make([]string, len(ctx.seenFields[ctx.table]))
	copy(ctx.seenFields[arrayTable], ctx.seenFields[ctx.table])
	ctx.uniqueColumns[arrayTable] = make(map[string]struct{},
		len(ctx.uniqueColumns[ctx.table]))
	for field := range ctx.uniqueColumns[ctx.table] {
		ctx.uniqueColumns[arrayTable][field] = struct{}{}
	}
	ctx.uniqueFields[arrayTable] = make(map[string]struct{}, len(ctx.uniqueFields[ctx.table]))
	for field := range ctx.uniqueFields[ctx.table] {
		ctx.uniqueFields[arrayTable][field] = struct{}{}
	}

	// Set the current table as the new array table's parent.
	err = arrayTable.SetParent(ctx.table)
	if err != nil {
		return nil, err
	}

	newCtx.db.AddTable(ctx.logger, arrayTable)
	newCtx.table = arrayTable

	ctx.logger.Debugf(log.Dev, "mapped new table %q for array at field path %q",
		arrayTableName, newCtx.path)

	return newCtx, nil
}

// nestedArrayContext returns a new mappingContext whose nestedArrayDepth
// value is one larger than the current context's.
func (ctx *mappingContext) nestedArrayContext() *mappingContext {
	newCtx := ctx.copy()
	newCtx.nestedArrayDepth = newCtx.nestedArrayDepth + 1
	return newCtx
}

/*
 * The remaining functions in this file are miscellaneous helper functions.
 */

// withSubpath returns a new mappingContext whose path is the specified subpath
// of the current context's path. If the new context's absolute path is
// mongoPrimaryKey, the context will also have inPrimaryKey = true.
func (ctx *mappingContext) withSubpath(subPath string) *mappingContext {

	// construct a new absolute path from the context's current path and the
	// provided subpath
	absPath := subPath
	if ctx.path != "" {
		absPath = ctx.path + "." + subPath
	}
	// update the mongoNamesPrefixes if needed due to prefix path appearing in
	// in mongoNamePrefixes.
	if renamedPath, ok := ctx.mongoNamePrefixes[ctx.table][ctx.path]; ok {
		ctx.mongoNamePrefixes[ctx.table][absPath] = renamedPath + "." + subPath
	}

	// create a new mappingContext with the new path
	newCtx := ctx.copy()
	newCtx.path = absPath

	// if the path is mongoPrimaryKey, we have entered the primary key
	if newCtx.path == schema.MongoPrimaryKey {
		newCtx.inPrimaryKey = true
	}

	return newCtx
}

// copy returns a new mappingContext whose fields are all equal to the current
// context's fields.
func (ctx *mappingContext) copy() *mappingContext {
	return newMappingContext(
		ctx.logger,
		ctx.db,
		ctx.table,
		ctx.uuidSubtype3Encoding,
		ctx.isAtLeastVersion34,
		ctx.mongoNames,
		ctx.mongoNamePrefixes,
		ctx.seenFields,
		ctx.uniqueColumns,
		ctx.uniqueFields,
		ctx.heuristic,
		ctx.path,
		ctx.inPrimaryKey,
		ctx.hasConflict,
		ctx.nestedArrayDepth,
		ctx.maxNumColumnsPerTable,
		ctx.maxNestedTableDepth,
	)
}

// getDominantSchemas returns the dominant schema for the provided Schemata.
// If there were multiple candidate schemas, a warning will be logged.
func (ctx *mappingContext) getDominantSchemas(s *mongo.Schemata) []*mongo.Schema {
	// get the dominant schema.
	dominant := ctx.heuristic(s)
	if dominant == nil {
		return []*mongo.Schema{mongo.NewEmptySchema()}
	}

	// if we had multiple schemas for this path, log a warning
	if len(s.Schemas) > 1 {
		bsonTypes := []string{}
		for bt := range s.Schemas {
			bsonTypes = append(bsonTypes, fmt.Sprintf("%q", string(bt)))
		}
		if len(dominant) == 1 {
			ctx.logger.Warnf(log.Dev, "table %q: multiple types at field path %q: [%v] - using %q",
				ctx.table.SQLName(), ctx.path, strings.Join(bsonTypes, ", "), dominant[0].BSONType,
			)
		} else {
			ctx.logger.Warnf(log.Dev, "table %q: multiple types at field path %q: [%v] - "+
				"using object conflict resolution for %q and object",
				ctx.table.SQLName(), ctx.path, strings.Join(bsonTypes, ", "),
				dominant[0].BSONType)
		}
	}

	return dominant
}

// newColumn creates a new column with the given name from the provided scalar
// schema, mapping the schema's BSONType and SpecialType to the appropriate
// SQLType and MongoType. It also records the bson types that are sampled for a
// given column.  If this function returns a nil column and a nil error, then
// the type represented by the provided schema was intentionally ignored.
func newColumn(sqlName, mongoName string, js *mongo.Schema,
	uuidSubtype3Encoding string, sampledTypes []mongo.BSONType) (*schema.Column, error) {
	var sqlType schema.SQLType
	var mongoType schema.MongoType

	switch js.BSONType {
	case mongo.Int:
		sqlType = schema.SQLInt
		mongoType = schema.MongoInt
	case mongo.Long:
		sqlType = schema.SQLInt
		mongoType = schema.MongoInt64
	case mongo.Double:
		sqlType = schema.SQLFloat
		mongoType = schema.MongoFloat
	case mongo.Decimal:
		sqlType = schema.SQLDecimal
		mongoType = schema.MongoDecimal128
	case mongo.Boolean:
		sqlType = schema.SQLBoolean
		mongoType = schema.MongoBool
	case mongo.Date:
		sqlType = schema.SQLTimestamp
		mongoType = schema.MongoDate
	case mongo.ObjectID:
		sqlType = schema.SQLObjectID
		mongoType = schema.MongoObjectID
	case mongo.String:
		sqlType = schema.SQLVarchar
		mongoType = schema.MongoString
	case mongo.BinData:
		switch js.SpecialType {
		case mongo.UUID3:
			subtype, err := newMongoUUIDSubtype3(uuidSubtype3Encoding)
			if err != nil {
				return nil, err
			}
			sqlType = schema.SQLVarchar
			mongoType = subtype
		case mongo.UUID4:
			sqlType = schema.SQLVarchar
			mongoType = schema.MongoUUID
		default:
			// ignore any non-uuid binData
			return nil, nil
		}
	case mongo.Array:
		if js.SpecialType == mongo.GeoPoint {
			sqlType = schema.SQLArrNumeric
			mongoType = schema.MongoGeo2D
		} else {
			return nil, fmt.Errorf("cannot create new column from array schema with SpeciaType"+
				" '%s'", js.SpecialType)
		}
	case mongo.Object:
		return nil, fmt.Errorf("cannot create new column from object schema")
	default:
		return nil, fmt.Errorf("cannot create new column: unsupported BSON type %s, check the definition of IsUnmappableType", js.BSONType)
	}

	return schema.NewColumnWithSampledTypes(
		sqlName,
		sqlType,
		mongoName,
		mongoType,
		sampledTypes,
	), nil
}

func newMongoUUIDSubtype3(uuidSubtype3Encoding string) (schema.MongoType, error) {
	switch uuidSubtype3Encoding {
	case "old":
		return schema.MongoUUIDOld, nil
	case "csharp":
		return schema.MongoUUIDCSharp, nil
	case "java":
		return schema.MongoUUIDJava, nil
	}
	err := fmt.Errorf("cannot create new column from UUID with encoding '%s'", uuidSubtype3Encoding)
	return schema.MongoNone, err
}
