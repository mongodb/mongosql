package catalog

import (
	"fmt"
	"math"

	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"

	"github.com/10gen/mongo-go-driver/bson"

	"sort"
	"strconv"
	"strings"
)

// addColumnToIndex takes a MongoDB index (which can be compound) and
// checks if all members of the index key are included in a table's
// columns. If they are, we add the columns and return the index;
// otherwise, we return nil.
func addColumnToIndex(index mongodb.Index, mongoNameToColumn map[string]Column) *Index {
	uniqueIndex := &Index{}
	for _, key := range index.Key {
		column, ok := mongoNameToColumn[key.Name]
		if !ok || key.Name == mongoPrimaryKey {
			return nil
		}
		uniqueIndex.columns = append(uniqueIndex.columns, column)
	}
	return uniqueIndex
}

// containsSiblingPaths returns true if all elements in paths are prefixes
// of the longest (by dot-delimited component) element and false otherwise.
func containsSiblingPaths(paths []string) bool {
	sort.Slice(paths, func(i, j int) bool {
		return len(strings.Split(paths[i], ".")) < len(strings.Split(paths[j], "."))
	})

	longestPath := paths[len(paths)-1]
	for i := 0; i < len(paths)-1; i++ {
		if len(paths[i]) > len(longestPath) || paths[i] != longestPath[:len(paths[i])] {
			return true
		}
	}

	return false
}

func createForeignKeyName(db, table, foreignTable string) string {
	position := 1
	return "fk_" + db + "_" + table + "_" + "to_" + foreignTable + "_" + strconv.Itoa(position)
}

func createUniqueIndexName(database, table string, position int) string {
	return database + "_" + table + "_" + strconv.Itoa(position) + "_UNIQUE"
}

// getIndexKey returns index key associated with the given column.
//
// If the column either isn't indexed or is indexed only as a
// secondary column in a multiple-column, non-unique index, the
// key is empty.
//
// If the column is a primary key or is one of the columns in a
// multiple-column primary key, the key is primaryKey.
//
// If the column is the first column of a unique index, the key is uniqueKey.
//
// If the column is the first column of a non-unique index in which multiple
// occurrences of a given value are permitted within the column, the key is multiKey.
func getIndexKey(col Column, tbl Table) string {
	colName := string(col.Name())

	for _, pk := range tbl.PrimaryKeys() {
		if colName == string(pk.Name()) {
			return primaryKey
		}
	}

	for _, idx := range tbl.Indexes() {
		if len(idx.columns) == 1 {
			if colName == string(idx.columns[0].Name()) {
				if idx.unique {
					return uniqueKey
				}
				return multiKey
			}
		} else if !idx.unique {
			if colName == string(idx.columns[0].Name()) {
				return multiKey
			}
		}
	}

	return ""
}

// getKeyToParentTable returns the most proximal foreign key - relative
// to the given depth - from the list of foreign key candidates.
func getKeyToParentTable(foreignKeys foreignKeyCandidates, depth int) *foreignKeyCandidate {
	var keyToParent *foreignKeyCandidate
	for _, key := range foreignKeys {
		if key.depth < depth {
			keyToParent = key
			continue
		}
		break
	}
	return keyToParent
}

// getUnwindPaths returns a list of unwind paths found in the aggregation pipeline
// and a map of the associated mongoName to its path. For a given path, if either
// the path or array index does not exist, neither is added to the returned values.
func getUnwindPaths(pipeline []bson.D) ([]string, map[string]string) {
	unwindPaths := []string{}
	pathAliases := make(map[string]string)
	for _, d := range pipeline {
		unwindVal, ok := d.Map()["$unwind"]
		if !ok {
			continue
		}

		unwind, ok := unwindVal.(bson.D)
		if !ok {
			continue
		}

		path, ok := unwind.Map()["path"]
		if !ok {
			continue
		}

		pathAsString, ok := path.(string)
		if !ok {
			continue
		}

		arrayIndexNameVal, ok := unwind.Map()["includeArrayIndex"]
		if !ok {
			continue
		}

		arrayIndexNameStringVal, ok := arrayIndexNameVal.(string)
		if !ok {
			continue
		}

		// only consider unwindPaths where there is a defined columns
		unwindPaths = append(unwindPaths, pathAsString)
		pathAliases[arrayIndexNameStringVal] = pathAsString

	}
	return unwindPaths, pathAliases
}

// sortForeignKeyCandidates sorts the foreignKeyCandidates by
// how deeply nested they are.
func sortForeignKeyCandidates(foreignKeyCandidates map[string]potentialForeignKeys) {
	for collectionName := range foreignKeyCandidates {
		for _, candidates := range foreignKeyCandidates[collectionName] {
			sort.Slice(candidates, func(i, j int) bool {
				return candidates[i].depth < candidates[j].depth
			})
		}
	}
}

func translateColumnType(sqlType schema.SQLType, maxVarcharLength uint16) string {
	switch sqlType {
	case schema.SQLBoolean:
		return "tinyint(1)"
	case schema.SQLDate:
		return "date"
	case schema.SQLDecimal128:
		return "decimal(65,20)"
	case schema.SQLFloat, schema.SQLNumeric, schema.SQLArrNumeric:
		return "double"
	case schema.SQLInt, schema.SQLInt64:
		return "bigint(20)"
	case schema.SQLObjectID:
		return "varchar(24)"
	case schema.SQLTimestamp:
		return "datetime(6)"
	case schema.SQLUint64:
		return "bigint(20) unsigned"
	case schema.SQLUUID:
		return "varchar(36)"
	case schema.SQLVarchar:
		length := maxVarcharLength
		if length == 0 {
			length = math.MaxUint16
		}
		return fmt.Sprintf("varchar(%d)", length)
	default:
		return "<unknown>"
	}
}
