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
	for _, elem := range doc {
		if mp, ok := elem.Map()["key"]; ok {
			if bsonD, ok := mp.(bson.D); ok && len(bsonD) == 1 {
				value := fmt.Sprintf("%v", bsonD[0].Value)
				if value == string(Index2DSphere) {
					indexTypeByKey[bsonD[0].Name] = Index2DSphere
				} else if value == string(Index2D) {
					indexTypeByKey[bsonD[0].Name] = Index2D
				}
			}
		}
	}
	return indexTypeByKey
}
