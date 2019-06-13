package bsonutil

import (
	"bytes"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// NormalizeBSONMs replaces all instances of bson.M with bson.D internally, to make
// diffing easier in tests.
func NormalizeBSONMs(input interface{}) interface{} {
	ret := input
	switch typed := input.(type) {
	case [][]bson.D:
		for i, docList := range typed {
			typed[i] = NormalizeBSONMs(docList).([]bson.D)
		}
	case []bson.D:
		for i, doc := range typed {
			typed[i] = NormalizeBSONMs(doc).(bson.D)
		}
	case bson.A:
		for i, val := range typed {
			typed[i] = NormalizeBSONMs(val)
		}
	case bson.D:
		for i, elem := range typed {
			typed[i] = NormalizeBSONMs(elem).(bson.E)
		}
		sort.Slice(typed, func(i, j int) bool {
			return typed[i].Key < typed[j].Key
		})
	case bson.M:
		out := make([]bson.E, len(typed))

		i := 0
		for key := range typed {
			out[i] = NewDocElem(key, NormalizeBSONMs(typed[key]))
			i++
		}

		sort.Slice(out, func(i, j int) bool {
			return out[i].Key < out[j].Key
		})

		ret = NewD(out...)
	case bson.E:
		typed.Value = NormalizeBSONMs(typed.Value)
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
		left, right := NormalizeBSONMs(pipeline1[i]), NormalizeBSONMs(pipeline2[i])
		leftBytes, err := bson.Marshal(left)
		if err != nil {
			return false
		}
		rightBytes, err := bson.Marshal(right)
		if err != nil {
			return false
		}
		if !bytes.Equal(leftBytes, rightBytes) {
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
