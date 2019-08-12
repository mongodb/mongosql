package procutil

import (
	"fmt"
	"strconv"
	"strings"
)

// URI literals
const (
	MongoDBScheme    = "mongodb"
	MongoDBSRVScheme = "mongodb+srv"
)

const (
	// InvalidDBChars lists a number of characters that are
	// invalid for use in a database name.
	InvalidDBChars = "/\\. \"\x00$"
)

// ParseHost parses a host string into a valid mongodb uri of the form:
//     <scheme>://<host>:<port>
// or
//     <scheme>://<host>:<port>,...,<host>:<port>/?replicaSet=<replSetName>
//
// The host string may be any of these cases:
//  - <host>
//  - <host>:<port>
//  - <scheme>://<host>
//  - <scheme>://<host>:<port>
//  - <scheme>://<replSetName>/<host>:<port>,...,<host>:<port>
//  - <replSetName>/<host>:<port>,...,<host>:<port>
//
// A port may be provided. If port is non-empty, the host string is not
// a replica set seedlist and contains a port, and preferPort is true,
// then the port in the host will be replaced with the port argument.
//
// parseHost returns the uri string and a bool indicating whether or not
// it replaced the port in the host string using the provided port.
func ParseHost(host, port string, preferPort bool) (string, bool) {
	if host == "" {
		if port == "" {
			host = "localhost:27017"
		} else {
			host = "localhost"
		}
	}

	scheme := MongoDBScheme
	replSetName := ""
	hosts := host
	replacedPort := false

	// Step 1: Get the scheme
	if strings.HasPrefix(hosts, MongoDBScheme) ||
		strings.HasPrefix(hosts, MongoDBSRVScheme) {
		parts := strings.SplitN(hosts, "://", 2)
		scheme = parts[0]
		hosts = parts[1]
	}

	// Step 2: Get the replica set name
	if strings.Contains(hosts, "/") {
		parts := strings.SplitN(hosts, "/", 2)
		replSetName = parts[0]
		hosts = parts[1]
	}

	// Step 3: Get the port
	if port != "" && replSetName == "" {
		if strings.Contains(hosts, ":") {
			parts := strings.SplitN(hosts, ":", 2)
			hosts = parts[0]
			if preferPort {
				replacedPort = true
			} else {
				port = parts[1]
			}
		}

		hosts = fmt.Sprintf("%v:%v", hosts, port)
	}

	// Step 4: Build the URI string
	uri := fmt.Sprintf("%v://%v", scheme, hosts)

	if replSetName != "" {
		uri = fmt.Sprintf("%v/?replicaSet=%v", uri, replSetName)
	}

	return uri, replacedPort
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
