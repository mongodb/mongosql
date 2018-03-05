package schema

import "strings"

const (
	// MongoPrimaryKey is a constant for MongoDB's "primary key" field (i.e. "_id").
	MongoPrimaryKey = "_id"
)

type normalizedName string

func normalizeSQLName(name string) normalizedName {
	str := strings.ToLower(name)
	return normalizedName(str)
}
