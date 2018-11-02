package bsonutil

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/util"
)

var bsonDType = reflect.TypeOf(bson.D{})

// NumberedDoc gives an enumeration for bson.D's. This allows us to retain the
// original pipeline stage number for a bson.D that we have projected out (such
// as if we want all $unwinds, or all $addFields)
type NumberedDoc struct {
	number int
	doc    bson.D
}

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
		out := make(bson.D, len(typed))
		i := 0
		for key := range typed {
			out[i] = bson.DocElem{Name: key, Value: NormalizeBSON(typed[key])}
			i++
		}
		sort.Slice(out, func(i, j int) bool {
			return out[i].Name < out[j].Name
		})
		ret = out
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

// UnwindInfo contains the relevant info from an $unwind operation.
type UnwindInfo struct {
	// Position in the original pipeline.
	StageNumber int
	// Path of the $unwind.
	Path string
	// Index name.
	Index string
}

func (in *UnwindInfo) getPath() string {
	return in.Path
}

func (in *UnwindInfo) getIndex() string {
	return in.Index
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

// FindUnwindForPath finds an unwind in an []UnwindInfo that has the proper
// unwind path
func FindUnwindForPath(unwinds []UnwindInfo, path string) (UnwindInfo, bool) {
	for _, unwind := range unwinds {
		if unwind.Path == path {
			return unwind, true
		}
	}
	return UnwindInfo{StageNumber: -1, Path: "", Index: ""}, false
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

func getPipelineUnwinds(pipeline []bson.D) []*NumberedDoc {
	return getPipelineStages("$unwind", pipeline)
}

// GetPaths gets the paths from a slice of UnwindInfo
// as a slice of strings.
func GetPaths(in []UnwindInfo) []string {
	return getFields(in, (*UnwindInfo).getPath)
}

// GetIndexes gets the index name from a slice of UnwindInfo
// as a slice of strings.
func GetIndexes(in []UnwindInfo) []string {
	return getFields(in, (*UnwindInfo).getIndex)
}

func getFields(in []UnwindInfo, m func(v *UnwindInfo) string) []string {
	ret := make([]string, len(in))
	for i, v := range in {
		ret[i] = m(&v)
	}
	return ret
}

// GetPipelineUnwindFields get all the unwind fields for a pipeline, in order
func GetPipelineUnwindFields(pipeline []bson.D) []UnwindInfo {
	unwinds := getPipelineUnwinds(pipeline)
	ret := make([]UnwindInfo, len(unwinds))
	var path string
	var index string
	for i, NumberedDoc := range unwinds {
		doc := NumberedDoc.doc
		unwind := doc.Map()["$unwind"]
		unwindDoc, ok := unwind.(bson.D)
		var fields bson.M
		if ok {
			fields = unwindDoc.Map()
		} else {
			fields = unwind.(bson.M)
		}
		path = fields["path"].(string)
		if index, ok = fields["includeArrayIndex"].(string); !ok {
			index = ""
		}
		ret[i] = UnwindInfo{StageNumber: NumberedDoc.number, Path: path, Index: index}
	}
	return ret
}

// nolint: unparam
func getPipelineStages(stage string, pipeline []bson.D) []*NumberedDoc {
	ret := make([]*NumberedDoc, 0)
	for i, doc := range pipeline {
		if _, ok := doc.Map()[stage]; ok {
			ret = append(ret, &NumberedDoc{number: i, doc: doc})
		}
	}
	return ret
}

// GetUnwindSuffix will give the remaining unwinds for two slices of unwinds
// after matching on unwind path.
func GetUnwindSuffix(unwinds1, unwinds2 []UnwindInfo) ([]UnwindInfo, bool) {
	ret := make([]UnwindInfo, 0)
	end := util.MinInt(len(unwinds1), len(unwinds2))
	i := 0
	for ; i < end; i++ {
		// Prefixes are incompatible, so there is no suffix
		// don't check index, assume that is correct
		if unwinds1[i].Path != unwinds2[i].Path {
			return nil, false
		}
	}
	var tail []UnwindInfo
	if len(unwinds1) <= len(unwinds2) {
		tail = unwinds2
	} else {
		tail = unwinds1

	}
	for ; i < len(tail); i++ {
		ret = append(ret, tail[i])
	}
	return ret, true
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
