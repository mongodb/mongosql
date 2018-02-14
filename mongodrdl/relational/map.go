package relational

import (
	"fmt"
	"sort"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/mongodrdl/mongo"
)

const (
	mongoPrimaryKey = "_id"
)

// +++++++++++++++++++++
type fieldSlice []*mongo.Field

func (slice fieldSlice) Len() int           { return len(slice) }
func (slice fieldSlice) Less(i, j int) bool { return slice[i].Name < slice[j].Name }
func (slice fieldSlice) Swap(i, j int)      { slice[i], slice[j] = slice[j], slice[i] }
func (slice fieldSlice) Sort()              { sort.Sort(slice) }

// +++++++++++++++++++++

type mappingContext struct {
	db           *Database
	table        *Table
	indexes      []mongodb.Index
	inPrimaryKey bool
}

func (d *Database) Map(c *mongo.Collection, idxs []mongodb.Index, preJoined bool) error {
	t, err := d.AddTable(c.Name, c.Name)
	if err != nil {
		return err
	}
	d.logger.Infof(log.Admin, "Created table %q for namespace %q.%q", c.Name, d.Name, c.Name)

	ctx := &mappingContext{d, t, idxs, false}

	err = ctx.mapDocument("", &c.Document)
	if err != nil {
		return err
	}

	var tables []*Table
	for _, t := range d.Tables {
		t.copyParent(!preJoined)
		if len(t.Columns) > 0 {
			t.Columns.Sort()
			tables = append(tables, t)
		} else {
			d.logger.Infof(log.Admin, "Removed table %q: had no columns.", t.Name)
		}
	}

	d.Tables = tables
	return nil
}

func (ctx *mappingContext) mapDocument(path string, doc *mongo.Document) error {
	var fields fieldSlice
	for _, f := range doc.Fields {
		fields = append(fields, f)
	}
	fields.Sort()

	for _, f := range fields {
		fieldName := appendFieldName(path, f.Name)
		fieldType, ok := tryGetType(ctx, fieldName, &f.TypeContainer)
		if !ok {
			ctx.db.logger.Infof(log.Dev, "Table %q, column %q has no types: ignoring column.", ctx.table.Name, fieldName)
			continue // skip a field without a type
		}

		switch v := fieldType.(type) {
		case *mongo.Array:
			tableName := appendTableName(ctx.table.rootName(), fieldName)
			arrayTable, err := ctx.db.AddTable(tableName, ctx.table.CollectionName)
			if err != nil {
				return err
			}
			ctx.db.logger.Infof(log.Admin, "Created table %q for namespace %q.%q", tableName, ctx.db.Name, ctx.table.rootName())

			arrayTable.Parent = ctx.table

			newCtx := &mappingContext{ctx.db, arrayTable, ctx.indexes, false}
			err = mapArray(newCtx, fieldName, v, 0)
			if err != nil {
				return err
			}
		case *mongo.Document:
			oldInPrimaryKey := ctx.inPrimaryKey
			if fieldName == mongoPrimaryKey {
				ctx.inPrimaryKey = true
			}
			err := ctx.mapDocument(fieldName, v)
			if err != nil {
				return err
			}
			ctx.inPrimaryKey = oldInPrimaryKey
		case *mongo.Scalar:
			switch v.Name() {
			case mongo.TimestampSchemaTypeName:
				ctx.db.logger.Infof(log.Dev, "ignoring timestamp column '%s' in table '%s'", fieldName, ctx.table.Name)
				continue
			case mongo.BinaryType0SchemaTypeName:
				ctx.db.logger.Infof(log.Dev, "ignoring binary type 0 column '%s' in table '%s'", fieldName, ctx.table.Name)
				continue
			default:
				c, err := ctx.table.AddColumn(fieldName, v.Name())
				if err != nil {
					return err
				}
				if fieldName == mongoPrimaryKey || ctx.inPrimaryKey {
					ctx.table.PrimaryKey = append(ctx.table.PrimaryKey, c)
				}
			}
		case *geo:
			fieldName := v.modifyPath(fieldName)
			_, err := ctx.table.AddColumn(fieldName, v.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func mapArray(ctx *mappingContext, path string, array *mongo.Array, depth int) error {
	fieldType, ok := tryGetType(ctx, path, &array.TypeContainer)
	if !ok {
		ctx.db.logger.Infof(log.Dev, "Table %q, column %q is an array that has no types: ignoring column.", ctx.table.Name, path)
		return nil
	}

	indexName := path + "_idx"
	if depth > 0 {
		indexName += fmt.Sprintf("_%v", depth)
	}

	c, err := ctx.table.AddColumn(indexName, "int")
	if err != nil {
		return err
	}

	ctx.table.PrimaryKey = append(ctx.table.PrimaryKey, c)
	ctx.table.addUnwind(path, indexName)

	switch v := fieldType.(type) {
	case *mongo.Array:
		err = mapArray(ctx, path, v, depth+1)
		if err != nil {
			return err
		}
	case *mongo.Document:
		err = ctx.mapDocument(path, v)
		if err != nil {
			return err
		}
	default:
		_, err := ctx.table.AddColumn(path, v.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func appendFieldName(path string, name string) string {
	if path == "" {
		return name
	}

	return path + "." + name
}

func appendTableName(tableName string, name string) string {
	return tableName + "_" + strings.Replace(name, ".", "_", -1)
}

// +++++++++++++++++++++
type sortByDescendingCountTypeSlice []mongo.Type

func (slice sortByDescendingCountTypeSlice) Len() int { return len(slice) }
func (slice sortByDescendingCountTypeSlice) Less(i, j int) bool {
	return slice[i].Count() > slice[j].Count()
}
func (slice sortByDescendingCountTypeSlice) Swap(i, j int) { slice[i], slice[j] = slice[j], slice[i] }
func (slice sortByDescendingCountTypeSlice) Sort()         { sort.Stable(slice) }

type geoKind string

const (
	geo2dArray   geoKind = "array"
	geoJsonPoint geoKind = "geoJson"
)

type geo struct {
	kind geoKind
}

func (g *geo) Name() string {
	return "geo.2darray"
}
func (g *geo) Count() int {
	// this doesn't matter
	return 0
}
func (g *geo) Combine(with mongo.Type) {
	// nothing to do
}
func (g *geo) Copy() mongo.Type {
	return &geo{g.kind}
}

func (g *geo) modifyPath(path string) string {
	switch g.kind {
	case geoJsonPoint:
		return appendFieldName(path, "coordinates")
	default:
		return path
	}
}

func tryGetType(ctx *mappingContext, path string, tc *mongo.TypeContainer) (mType mongo.Type, ok bool) {

	var types sortByDescendingCountTypeSlice = tc.Types
	types.Sort()

	defer func() {
		if mongo.UUIDSubtype3Encoding != "" {
			ctx.db.logger.Infof(log.Dev, `Using type %s for table "%s", column "%s"`, mongo.UUIDSubtype3Map[mongo.UUIDSubtype3Encoding], ctx.table.Name, path)
		}
	}()

	// even if we have no types for this field, certain types of indexes
	// dictate certain "schemas" and we can leverage that.
	if indexedType, ok := tryGetTypeFromIndex(ctx, path, types); ok {
		mType = indexedType
		return mType, true
	}

	if types.Len() == 0 {
		return mType, false
	}

	if types.Len() == 1 {
		mType = types[0]
		return mType, true
	}

	// we only care about the top 2
	t0, ok0 := types[0].(*mongo.Array)
	t1, ok1 := types[1].(*mongo.Array)

	if ok0 || ok1 {
		// one of them is an array (cause we know both can't be an array because
		// types in a TypeContainer are unique)
		var t mongo.Type
		var array *mongo.Array
		if ok0 {
			array = t0
			t = types[1]
		} else {
			array = t1
			t = types[0]
		}

		// get the predominant type for the array...
		if itemType, ok := tryGetType(ctx, path, &array.TypeContainer); ok {
			dt0, ok0 := t.(*mongo.Document)
			dt1, ok1 := itemType.(*mongo.Document)
			if ok0 && ok1 {
				// create a new array with the combined documents such that
				// each field both in the array and in the scalar both exist
				// in the created schema
				newD := mongo.CombineDocuments(dt0, dt1)
				newA := mongo.NewArray()
				newA.Types = append(newA.Types, newD)
				newA.SetCount(array.Count() + t.Count())
				mType = newA
				return mType, true
			} else if t.Name() == itemType.Name() {
				mType = array
				return mType, true
			}
		}
	}

	if isNumeric(types[0]) && isNumeric(types[1]) {
		mType = mongo.NewScalar("number")
	} else {
		// Either neither were an array, or the array's item type and the other top type
		// are incompatible so we'll just return the first one.
		ctx.db.logger.Infof(log.Dev, `Type conflict for table "%s", column "%s": using type %q`, ctx.table.Name, path, types[0].Name())
		mType = types[0]
	}
	return mType, true
}

func isNumeric(t mongo.Type) bool {
	name := t.Name()
	return strings.HasPrefix(name, "int") ||
		strings.HasPrefix(name, "float") ||
		name == "bson.Decimal128"
}

func tryGetTypeFromIndex(ctx *mappingContext, path string, types sortByDescendingCountTypeSlice) (mongo.Type, bool) {

	if indexType, ok := tryGetIndex(ctx.indexes, path); ok {

		var t mongo.Type

		if types.Len() == 0 {
			// we don't really know, so we'll ignore it
			return nil, false
		}

		switch types[0].(type) { //majority type wins
		case *mongo.Array: // use the default
			t = &geo{geo2dArray}
		case *mongo.Document:
			if indexType == "2d" {
				// 2d indexes using document form don't need a special type
				t = types[0]
			} else {
				// we don't have the sampled data at this point because we aren't keeping it, so
				// all we can do is assume a geoJson Point, which is probably predominant.
				t = &geo{geoJsonPoint}
			}
		default:
			// we have sampled types that are contrary to the index type. We shouldn't get here,
			// but if we do, we won't use a geo type
			return nil, false
		}

		if types.Len() > 1 {
			ctx.db.logger.Infof(log.Dev, "Type conflict for table %q, column %q: using type %q.", ctx.table.Name, path, t.Name())
		}

		return t, true
	}

	return nil, false
}

func SimpleIndexKey(realKey bson.D) (key []string) {
	for i := range realKey {
		field := realKey[i].Name
		vi, ok := realKey[i].Value.(int)
		if !ok {
			vf, _ := realKey[i].Value.(float64)
			vi = int(vf)
		}

		if vi > 0 {
			key = append(key, field)
			continue
		}

		if vi < 0 {
			key = append(key, "-"+field)
			continue
		}

		if vs, ok := realKey[i].Value.(string); ok {
			key = append(key, "$"+vs+":"+field)
			continue
		}

		// In 3.4 only numbers > 0, numbers < 0, and strings are allowed
		// for index keys but 3.2. allows for all sorts of index hackery
		// - including zero, dates, etc - so we'll just stringify things
		// here. This is fine since we only specially treat 2d indexes.
		key = append(key, fmt.Sprintf("%v", realKey[i].Value))
	}
	return
}

func tryGetIndex(indexes []mongodb.Index, path string) (string, bool) {
	for _, index := range indexes {
		keys := SimpleIndexKey(index.Key)
		for _, key := range keys {
			if strings.HasSuffix(key, path) {
				if strings.HasPrefix(key, "$2d:") {
					return "2d", true
				} else if strings.HasPrefix(key, "$2dsphere:") {
					return "2dsphere", true
				}
			}
		}
	}

	return "", false
}
