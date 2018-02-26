package schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema/mongo"
)

// mappingContext maintains state that describes the context into which a mongo
// schema should be mapped.
type mappingContext struct {

	// logger is the logger used to output warnings and other information during
	// the mapping process
	logger *log.Logger

	// db is the database into which we are mapping the current schema.
	// this will never be changed by any mapping functions.
	db *Database

	// table is the table into which we are mapping the current schema.
	// this will change whenever we need to map an array field.
	table *Table

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

	// nestedArrayDepth tracks how many levels deep in a nested array this
	// context is. For non-array contexts and top-level array contexts, its
	// value should be zero. It should be incremented once for each level of
	// array nesting inside the top-level array context.
	nestedArrayDepth int
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

	// map each prop
	for _, prop := range props {

		// get the dominant schema for this prop
		s := ctx.withSubpath(prop).getDominantSchema(js.Properties[prop])
		switch s.BsonType {
		case mongo.NoBsonType:
			// ignore column (no types)
			ctx.logger.Warnf(log.Dev, "table %q, column %q has no types: ignoring column",
				ctx.table.Name, prop)

		case mongo.Object:
			err := ctx.objectContext(prop).mapObjectSchema(s)
			if err != nil {
				return err
			}

		case mongo.Array:
			if s.SpecialType != mongo.GeoPoint {
				subctx, err := ctx.arrayContext(prop)
				if err != nil {
					return err
				}
				err = subctx.mapArraySchema(s)
				if err != nil {
					return err
				}
				break
			}

			// if this is a geo.2darray, treat it as a scalar, by falling
			// through to scalar case
			fallthrough

		default: // scalar
			err := ctx.scalarContext(prop).mapScalarSchema(s)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

// mapArraySchema maps the provided array schema into a mappingContext.
func (ctx *mappingContext) mapArraySchema(js *mongo.Schema) error {

	// get the dominant schema for the array's elements
	items := ctx.getDominantSchema(js.Items)

	// don't map null arrays
	if items.BsonType == mongo.NoBsonType {
		return nil
	}

	// calculate the name of the array index column
	indexName := ctx.path + "_idx"
	if ctx.nestedArrayDepth > 0 {
		indexName += fmt.Sprintf("_%v", ctx.nestedArrayDepth)
	}

	// create the array index column and add it to the current table
	col := &Column{
		Name:      indexName,
		MongoType: MongoInt,
		SQLName:   indexName,
		SQLType:   SQLInt,
	}

	err := ctx.table.AddColumn(col, ctx.logger)
	if err != nil {
		return err
	}

	// add the index column to the current table's primary key
	ctx.table.primaryKey = append(ctx.table.primaryKey, col)

	// add an unwind to the current table's pipeline
	unwind := bson.D{
		{Name: "$unwind", Value: bson.D{
			{Name: "path", Value: "$" + ctx.path},
			{Name: "includeArrayIndex", Value: indexName},
		}},
	}
	ctx.table.Pipeline = append(ctx.table.Pipeline, unwind)

	// map the array's elements.
	// we need a subcontext if this an a nested array (to track depth)
	// for objects and scalars, we continue to use the array's context
	switch items.BsonType {
	case mongo.Array:
		err = ctx.nestedArrayContext().mapArraySchema(items)
	case mongo.Object:
		err = ctx.mapObjectSchema(items)
	default:
		err = ctx.mapScalarSchema(items)
	}

	return err
}

// mapScalarSchema maps the provided scalar schema into a mappingContext.
func (ctx *mappingContext) mapScalarSchema(js *mongo.Schema) error {

	// create a new column
	col, err := newColumn(ctx.path, js, ctx.uuidSubtype3Encoding)
	if err != nil {
		return err
	}

	// if err and col are both nil, we don't map a column for this schema
	if col == nil {
		return nil
	}

	// add the column to the current table
	err = ctx.table.AddColumn(col, ctx.logger)
	if err != nil {
		return err
	}

	// if we are in the primary key, add this column to the table's primary key
	if ctx.inPrimaryKey {
		ctx.table.primaryKey = append(ctx.table.primaryKey, col)
	}

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
	for root.parent != nil {
		root = root.parent
	}

	// calculate the name for this array's table
	arrayTableName := root.Name + "_" + strings.Replace(newCtx.path, ".", "_", -1)

	// create the array table; add it to newCtx.db and newCtx
	arrayTable := NewTable(arrayTableName, newCtx.table.CollectionName, nil, nil, nil, nil, false)
	err := newCtx.db.AddTable(arrayTable, ctx.logger)
	if err != nil {
		return nil, err
	}
	newCtx.table = arrayTable

	ctx.logger.Debugf(log.Dev, "mapped new table %q for array at field path %q",
		arrayTableName, newCtx.path,
	)

	// set the array table's parent table to the current context's table
	arrayTable.parent = ctx.table
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

	// create a new mappingContext with the new path
	newCtx := ctx.copy()
	newCtx.path = absPath

	// if the path is mongoPrimaryKey, we have entered the primary key
	if newCtx.path == mongoPrimaryKey {
		newCtx.inPrimaryKey = true
	}

	return newCtx
}

// copy returns a new mappingContext whose fields are all equal to the current
// context's fields.
func (ctx *mappingContext) copy() *mappingContext {
	return &mappingContext{
		logger:               ctx.logger,
		db:                   ctx.db,
		table:                ctx.table,
		path:                 ctx.path,
		uuidSubtype3Encoding: ctx.uuidSubtype3Encoding,
		nestedArrayDepth:     ctx.nestedArrayDepth,
		inPrimaryKey:         ctx.inPrimaryKey,
	}
}

// getDominantSchema returns the dominant schema for the provided Schemata.
// If there were multiple candidate schemas, a warning will be logged.
func (ctx *mappingContext) getDominantSchema(s *mongo.Schemata) *mongo.Schema {
	// get the dominant schema
	dominant := s.DominantSchema()

	// if we had multiple schemas for this path, log a warning
	if len(s.Schemas) > 1 {
		bsonTypes := []string{}
		for bt := range s.Schemas {
			if bt == mongo.NoBsonType {
				// use "empty" instead of "" so that log
				// messages make sense
				bt = mongo.BsonType("empty")
			}
			bsonTypes = append(bsonTypes, fmt.Sprintf("%q", string(bt)))
		}
		ctx.logger.Warnf(log.Dev, "table %q: multiple types at field path %q: [%v] - using %q",
			ctx.table.Name, ctx.path, strings.Join(bsonTypes, ", "), dominant.BsonType,
		)
	}

	return dominant
}

// newDatabase creates a new database with the provided name and tables.
func newDatabase(name string, tables []*Table) *Database {
	return &Database{Name: name, Tables: tables}
}

// NewTable creates a new table with the provided table and collection names.
func NewTable(tableName, collectionName string, pipeline []bson.D,
	columns []*Column, parent *Table, primaryKeys []*Column, isPostProcessed bool) *Table {
	return &Table{
		Name:            tableName,
		CollectionName:  collectionName,
		Pipeline:        pipeline,
		parent:          parent,
		Columns:         columns,
		primaryKey:      primaryKeys,
		isPostProcessed: isPostProcessed,
	}
}

// newColumn creates a new column with the given name from the provided scalar
// schema, mapping the schema's BsonType and SpecialType to the appropriate
// SQLType and MongoType. If this function returns a nil column and a nil error,
// then the type represented by the provided schema was intentionally ignored.
func newColumn(name string, js *mongo.Schema, uuidSubtype3Encoding string) (*Column, error) {
	var sqlType SQLType
	var mongoType MongoType

	switch js.BsonType {
	case mongo.Int:
		sqlType = SQLInt
		mongoType = MongoInt
	case mongo.Long:
		sqlType = SQLInt64
		mongoType = MongoInt64
	case mongo.Double:
		sqlType = SQLFloat
		mongoType = MongoFloat
	case mongo.Decimal:
		sqlType = SQLDecimal128
		mongoType = MongoDecimal128
	case mongo.Boolean:
		sqlType = SQLBoolean
		mongoType = MongoBool
	case mongo.Date:
		sqlType = SQLTimestamp
		mongoType = MongoDate
	case mongo.ObjectID:
		sqlType = SQLVarchar
		mongoType = MongoObjectID
	case mongo.String:
		sqlType = SQLVarchar
		mongoType = MongoString
	case mongo.BinData:
		switch js.SpecialType {
		case mongo.UUID3:
			subtype, err := newMongoUUIDSubtype3(uuidSubtype3Encoding)
			if err != nil {
				return nil, err
			}
			sqlType = SQLVarchar
			mongoType = subtype
		case mongo.UUID4:
			sqlType = SQLVarchar
			mongoType = MongoUUID
		default:
			// ignore any non-uuid binData
			return nil, nil
		}
	case mongo.Array:
		if js.SpecialType == mongo.GeoPoint {
			sqlType = SQLArrNumeric
			mongoType = MongoGeo2D
		} else {
			return nil, fmt.Errorf("cannot create new column from array schema with SpeciaType '%s'", js.SpecialType)
		}
	case mongo.Object:
		return nil, fmt.Errorf("cannot create new column from object schema")
	case mongo.NoBsonType:
		return nil, fmt.Errorf("cannot create new column from schema with no BSON type")
	default:
		return nil, fmt.Errorf("cannot create new column: unsupported BSON type %s", js.BsonType)
	}

	return &Column{
		Name:      name,
		MongoType: mongoType,
		SQLName:   name,
		SQLType:   sqlType,
	}, nil
}

func newMongoUUIDSubtype3(uuidSubtype3Encoding string) (MongoType, error) {
	switch uuidSubtype3Encoding {
	case "old":
		return MongoUUIDOld, nil
	case "csharp":
		return MongoUUIDCSharp, nil
	case "java":
		return MongoUUIDJava, nil
	}
	return MongoNone, fmt.Errorf("cannot create new column from UUID with encoding '%s'", uuidSubtype3Encoding)
}
