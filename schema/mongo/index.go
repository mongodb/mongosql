package mongo

import (
	"fmt"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
)

type IndexType string

const (
	Index2D       IndexType = "2d"
	Index2DSphere IndexType = "2dsphere"
)

type Index struct {
	Field string
	Type  IndexType
}

// AddIndexes traverses a Schema, looking to see if any of the provided indexes
// apply to any of its fields. When a field is found that matches one of the
// provided indexes, that field is annotated with the type of the index found.
// NOTE: currently, all indexes except for 2d and 2dsphere are ignored
func (s *Schema) AddIndexes(indexes []bson.D) {
	idxs := importIndexes(indexes)
	s.addIndexes(idxs, "")
}

func (s *Schema) addIndexes(idxs indexes, path string) {
	if path != "" {
		path = path + "."
	}
	for prop, schemata := range s.Properties {
		subPath := fmt.Sprintf("%s%s", path, prop)
		for key, typ := range idxs {
			if strings.HasSuffix(key, subPath) {
				schemata.Indexes = append(schemata.Indexes, typ)
			}
		}
		for _, schema := range schemata.Schemas {
			schema.addIndexes(idxs, subPath)
		}
	}
}

type indexes map[string]IndexType

// importIndexes takes a list of bson documents and returns them as an indexes
// instance (i.e. a map of the path to the index key to the index type).
func importIndexes(doc []bson.D) indexes {
	var indexTypeByKey indexes = make(map[string]IndexType)
	for _, index := range doc {
		keys := simplifyIndexKey(index)
		for _, key := range keys {
			if strings.HasPrefix(key, "$2d:") {
				indexTypeByKey[key[4:]] = Index2D
			} else if strings.HasPrefix(key, "$2dsphere:") {
				indexTypeByKey[key[10:]] = Index2DSphere
			}
		}
	}
	return indexTypeByKey
}

// simplifyIndexKey takes an index key as a bson document, and returns the key
// represented as a string slice.
func simplifyIndexKey(realKey bson.D) (key []string) {
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
