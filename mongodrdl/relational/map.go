package relational

import (
	"fmt"
	"github.com/10gen/sqlproxy/mongodrdl/mongo"
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2"
	"sort"
	"strings"
)

// +++++++++++++++++++++
type fieldSlice []*mongo.Field

func (slice fieldSlice) Len() int           { return len(slice) }
func (slice fieldSlice) Less(i, j int) bool { return slice[i].Name < slice[j].Name }
func (slice fieldSlice) Swap(i, j int)      { slice[i], slice[j] = slice[j], slice[i] }
func (slice fieldSlice) Sort()              { sort.Sort(slice) }

// +++++++++++++++++++++

type mappingContext struct {
	db      *Database
	table   *Table
	indexes []mgo.Index
}

func (d *Database) Map(c *mongo.Collection, idxs []mgo.Index) error {
	log.Logf(log.Info, "Adding table %q.", c.Name)
	t, err := d.AddTable(c.Name, c.Name)
	if err != nil {
		return err
	}

	ctx := &mappingContext{d, t, idxs}

	err = mapDocument(ctx, "", &c.Document)
	if err != nil {
		return err
	}

	var tables []*Table
	for _, t := range d.Tables {
		if len(t.Columns) > 0 {
			t.copyParent()
			tables = append(tables, t)
		} else {
			log.Logf(log.Info, "Removed table %q: had no columns.", t.Name)
		}
	}

	d.Tables = tables
	return nil
}

func mapDocument(ctx *mappingContext, path string, doc *mongo.Document) error {
	var fields fieldSlice
	for _, f := range doc.Fields {
		fields = append(fields, f)
	}
	fields.Sort()

	for _, f := range fields {
		fieldName := appendFieldName(path, f.Name)
		fieldType, ok := tryGetType(ctx, fieldName, &f.TypeContainer)
		if !ok {
			log.Logf(log.Always, "Table %q, column %q has no types: ignoring column.", ctx.table.Name, fieldName)
			continue // skip a field without a type
		}

		switch v := fieldType.(type) {
		case *mongo.Array:
			tableName := appendTableName(ctx.table.rootName(), fieldName)
			log.Logf(log.Info, "Adding table %q.", tableName)
			arrayTable, err := ctx.db.AddTable(tableName, ctx.table.CollectionName)
			if err != nil {
				return err
			}
			arrayTable.Parent = ctx.table

			newCtx := &mappingContext{ctx.db, arrayTable, ctx.indexes}
			err = mapArray(newCtx, fieldName, v, 0)
			if err != nil {
				return err
			}
		case *mongo.Document:
			err := mapDocument(ctx, fieldName, v)
			if err != nil {
				return err
			}
		case *mongo.Scalar:
			_, err := ctx.table.AddColumn(fieldName, v.Name())
			if err != nil {
				return err
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
		log.Logf(log.Always, "Table %q, column %q is an array that has no types: ignoring column.", ctx.table.Name, path)
		return nil
	}

	indexName := path + "_idx"
	if depth > 0 {
		indexName += fmt.Sprintf("_%v", depth)
	}

	_, err := ctx.table.AddColumn(indexName, "int")
	if err != nil {
		return err
	}

	ctx.table.addUnwind(path, indexName)

	switch v := fieldType.(type) {
	case *mongo.Array:
		err = mapArray(ctx, path, v, depth+1)
		if err != nil {
			return err
		}
	case *mongo.Document:
		err = mapDocument(ctx, path, v)
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
			log.Logf(log.Info, `Using type %s for table "%s", column "%s"`, mongo.UUIDSubtype3Map[mongo.UUIDSubtype3Encoding], ctx.table.Name, path)
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

	// Either neither were an array, or the array's item type and the other top type
	// are incompatible so we'll just return the first one.
	log.Logf(log.Info, `Type conflict for table "%s", column "%s": using type %q`, ctx.table.Name, path, types[0].Name())
	mType = types[0]
	return mType, true
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
			log.Logf(log.Info, "Type conflict for table %q, column %q: using type %q.", ctx.table.Name, path, t.Name())
		}

		return t, true
	}

	return nil, false
}

func tryGetIndex(indexes []mgo.Index, path string) (string, bool) {
	for _, index := range indexes {
		for _, key := range index.Key {
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
