package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"
)

const (
	// InvalidDBChars lists a number of characters that are
	// invalid for use in a database name.
	InvalidDBChars = "/\\. \"\x00$"
)

func bsonDToMap(doc bson.D) map[string]interface{} {
	m := map[string]interface{}{}
	for _, l := range doc {
		switch typedV := l.Value.(type) {
		case bson.D:
			m[l.Name] = bsonDToMap(typedV)
		case bson.M:
			m[l.Name] = bsonMToMap(typedV)
		default:
			m[l.Name] = typedV
		}
	}
	return m
}

func bsonMToMap(doc bson.M) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range doc {
		switch typedV := v.(type) {
		case bson.D:
			m[k] = bsonDToMap(typedV)
		case bson.M:
			m[k] = bsonMToMap(typedV)
		default:
			m[k] = typedV
		}
	}
	return m
}

// ConvertBSONToMap recursively converts a bson.D/bson.M to a map[string]interface{}.
func ConvertBSONToMap(doc interface{}) map[string]interface{} {
	switch typedD := doc.(type) {
	case bson.D:
		return bsonDToMap(typedD)
	case bson.M:
		return bsonMToMap(typedD)
	}
	panic(fmt.Sprintf("Unrecognized bson type: %T", doc))
}

// ParseConnectionString extracts the replica set name and the list
// of hosts from the connection string.
func ParseConnectionString(connString string) ([]string, string) {

	// strip off the replica set name from the beginning
	slashIndex := strings.Index(connString, "/")
	setName := ""
	if slashIndex != -1 {
		setName = connString[:slashIndex]
		if slashIndex == len(connString)-1 {
			return []string{""}, setName
		}
		connString = connString[slashIndex+1:]
	}

	// split the hosts, and return them and the set name
	return strings.Split(connString, ","), setName
}

// PipelineToMapSlice converts a slice of bson.D
// to a slice of map[string]interface - with
// each element recursively converted.
func PipelineToMapSlice(pipeline []bson.D) []map[string]interface{} {
	m := make([]map[string]interface{}, 0)
	for _, stage := range pipeline {
		m = append(m, ConvertBSONToMap(stage))
	}
	return m
}

// ValidateDBName validates that a string is a valid name for a mongodb
// database. An error is returned if it is not valid.
func ValidateDBName(database string) error {
	// must be < 64 characters
	if len([]byte(database)) > 63 {
		return fmt.Errorf("db name '%v' is longer than 63 characters", database)
	}

	// check for illegal characters
	for _, illegalRune := range InvalidDBChars {
		if strings.ContainsRune(database, illegalRune) {
			return fmt.Errorf("illegal character '%c' found in db name '%v'", illegalRune, database)
		}
	}

	return nil
}

// VersionAtLeast returns true if the currentVersion contains a version greater than or equal to the
// value specified in minRequiredVersion and false otherwise.
func VersionAtLeast(currentVersion []uint8, minRequiredVersion []uint8) bool {
	for idx, vi := range minRequiredVersion {
		if idx == len(currentVersion) {
			return false
		}
		if ivi := currentVersion[idx]; ivi != vi {
			return ivi >= vi
		}
	}
	return true
}

// VersionExactly returns true if the currentVersion matches
// the requiredVersion and returns false otherwise.
func VersionExactly(currentVersion []uint8, requiredVersion []uint8) bool {
	for idx, vi := range requiredVersion {
		if idx == len(currentVersion) {
			return false
		}
		if ivi := currentVersion[idx]; ivi != vi {
			return ivi == vi
		}
	}
	return true
}

// VersionToSlice converts a version string to a uint8 slice.
func VersionToSlice(versionStr string) ([]uint8, error) {
	var slice []uint8

	if i := strings.Index(versionStr, "-"); i != -1 {
		versionStr = versionStr[:i]
	}

	parts := strings.SplitN(versionStr, ".", 3)
	for _, p := range parts {
		i, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("expected an integer, got %q: %v", p, err)
		}
		slice = append(slice, uint8(i))
	}
	return slice, nil
}
