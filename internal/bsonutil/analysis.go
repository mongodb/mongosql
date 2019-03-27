package bsonutil

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
)

var bsonDType = reflect.TypeOf(NewD())

// NormalizeBSON replaces all instances of bson.M with bson.D internally, to make
// diffing easier in tests.
func NormalizeBSON(input interface{}) interface{} {
	ret := input
	switch typed := input.(type) {
	case [][]bson.D:
		for i, docList := range typed {
			typed[i] = NormalizeBSON(docList).([]bson.D)
		}
	case []bson.D:
		for i, doc := range typed {
			typed[i] = NormalizeBSON(doc).(bson.D)
		}
	case []interface{}:
		for i, val := range typed {
			typed[i] = NormalizeBSON(val)
		}
	case bson.D:
		for i, elem := range typed {
			typed[i] = NormalizeBSON(elem).(bson.DocElem)
		}
		sort.Slice(typed, func(i, j int) bool {
			return typed[i].Name < typed[j].Name
		})
	case bson.M:
		out := make([]bson.DocElem, len(typed))

		i := 0
		for key := range typed {
			out[i] = NewDocElem(key, NormalizeBSON(typed[key]))
			i++
		}

		sort.Slice(out, func(i, j int) bool {
			return out[i].Name < out[j].Name
		})

		ret = NewD(out...)
	case bson.DocElem:
		typed.Value = NormalizeBSON(typed.Value)
		ret = typed
	}
	return ret
}

// PipelineEqual returns true if both pipelines are equal and false otherwise.
func PipelineEqual(pipeline1, pipeline2 []bson.D) bool {
	if len(pipeline1) != len(pipeline2) {
		return false
	}
	for i := range pipeline1 {
		if pipeline1[i][0].Name != pipeline2[i][0].Name {
			return false
		}

		left, right := NormalizeBSON(pipeline1[i][0].Value), NormalizeBSON(pipeline2[i][0].Value)
		if !reflect.DeepEqual(left, right) {
			return false
		}
	}
	return true
}

// ContainsCardinalityAlteringStages returns true if a pipeline
// contains any stage that can change the number of documents
// returned by the pipeline.
func ContainsCardinalityAlteringStages(pipeline []bson.D) bool {
	for _, doc := range pipeline {
		for k := range doc.Map() {
			switch k {
			case "$addFields":
				continue
			case "$graphLookup":
				continue
			case "$lookup":
				continue
			case "$out":
				continue
			case "$project":
				continue
			case "$replaceRoot":
				continue
			case "$sort":
				continue
			default:
				return true
			}
		}
	}
	return false
}

// ExtractFieldByName takes a field name and document, and returns a value representing
// the value of that field in the document in a format that can be printed as a string.
// It will also handle dot-delimited field names for nested arrays or documents.
func ExtractFieldByName(fieldName string, document interface{}) (interface{}, bool) {
	dotParts := strings.Split(fieldName, ".")
	var subdoc interface{} = document

	for _, path := range dotParts {
		docValue := reflect.ValueOf(subdoc)
		if !docValue.IsValid() {
			return nil, false
		}
		docType := docValue.Type()
		docKind := docType.Kind()
		if docKind == reflect.Map {
			subdocVal := docValue.MapIndex(reflect.ValueOf(path))
			if subdocVal.Kind() == reflect.Invalid {
				return nil, false
			}
			subdoc = subdocVal.Interface()
		} else if docKind == reflect.Slice {
			if docType == bsonDType {
				// dive into a D as a document
				asD := subdoc.(bson.D)
				var err error
				subdoc, err = findValueByKey(path, &asD)
				if err != nil {
					return nil, false
				}
			} else {
				//  check that the path can be converted to int
				arrayIndex, err := strconv.Atoi(path)
				if err != nil {
					return nil, false
				}
				// bounds check for slice
				if arrayIndex < 0 || arrayIndex >= docValue.Len() {
					return nil, false
				}
				subdocVal := docValue.Index(arrayIndex)
				if subdocVal.Kind() == reflect.Invalid {
					return nil, false
				}
				subdoc = subdocVal.Interface()
			}
		} else {
			// trying to index into a non-compound type - just return blank.
			return nil, false
		}
	}
	return subdoc, true
}

// findValueByKey returns the value of keyName in document. If keyName is not found
// in the top-level of the document, ErrNoSuchField is returned as the error.
func findValueByKey(keyName string, document *bson.D) (interface{}, error) {
	for _, key := range *document {
		if key.Name == keyName {
			return key.Value, nil
		}
	}
	return nil, fmt.Errorf("no such field: %v", keyName)
}

// IsPrefix returns true if fieldToAdd is a path prefix of field
// based on the . field separator as the path separator.
func IsPrefix(field, fieldToAdd string) bool {
	fieldArr := strings.Split(field, ".")
	fieldToAddArr := strings.Split(fieldToAdd, ".")
	// if fieldToAddArr is longer than or the same length as
	// fieldArr, there is no way that fieldToAdd is a prefix
	// of field.
	if len(fieldToAddArr) >= len(fieldArr) {
		return false
	}
	for i := range fieldToAddArr {
		if fieldArr[i] != fieldToAddArr[i] {
			return false
		}
	}
	// Because they are equivalent up to i and fieldToAddArr is shorter
	// than fieldArr, it must be a prefix.
	return true
}
