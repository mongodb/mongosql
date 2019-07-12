package catalog

import (
	"fmt"
	"math"
	"strconv"

	"github.com/10gen/mongoast/ast"

	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/mongodb"
)

// addColumnToIndex takes a MongoDB index (which can be compound) and
// checks if all members of the index key are included in a table's
// columns. If they are, we add the columns and return the index;
// otherwise, we return nil.
func addColumnToIndex(index mongodb.Index, mongoNameToColumn map[string]*results.Column) *Index {
	uniqueIndex := &Index{constraintName: index.Name}
	for _, key := range index.Key {
		column, ok := mongoNameToColumn[key.Key]
		if !ok || key.Key == mongoPrimaryKey {
			return nil
		}
		uniqueIndex.columns = append(uniqueIndex.columns, column)
	}
	return uniqueIndex
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
func getIndexKey(col *results.Column, tbl Table) string {
	colName := col.Name

	for _, pk := range tbl.PrimaryKeys() {
		if colName == pk.Name {
			return primaryKey
		}
	}

	for _, idx := range tbl.Indexes() {
		if len(idx.columns) == 1 {
			if colName == idx.columns[0].Name {
				if idx.unique {
					return uniqueKey
				}
				return multiKey
			}
		} else if !idx.unique {
			if colName == idx.columns[0].Name {
				return multiKey
			}
		}
	}

	return ""
}

// getUnwindPaths returns a list of unwind paths found in the aggregation pipeline
// and a map of the associated mongoName to its path. For a given path, if either
// the path or array index does not exist, neither is added to the returned values.
func getUnwindPaths(pipeline *ast.Pipeline) ([]string, map[string]string) {
	unwindPaths := []string{}
	pathAliases := make(map[string]string)
	for _, s := range pipeline.Stages {
		unwind, ok := s.(*ast.UnwindStage)
		if !ok {
			continue
		}

		if unwind.Path == nil || unwind.IncludeArrayIndex == "" {
			continue
		}

		pathAsString := astutil.FieldRefString(unwind.Path)

		// only consider unwindPaths where there is a defined columns
		unwindPaths = append(unwindPaths, pathAsString)
		pathAliases[pathAsString] = unwind.IncludeArrayIndex

	}
	return unwindPaths, pathAliases
}

func translateColumnType(sqlType types.EvalType, maxVarcharLength uint64) string {
	switch sqlType {
	case types.EvalBoolean:
		return "tinyint(1)"
	case types.EvalDate:
		return "date"
	case types.EvalDatetime:
		return "datetime"
	case types.EvalDecimal128:
		return "decimal(65,20)"
	case types.EvalDouble, types.EvalArrNumeric:
		return "double"
	case types.EvalInt64:
		return "bigint(20)"
	case types.EvalObjectID:
		return "varchar(24)"
	case types.EvalTimestamp:
		return "datetime(6)"
	case types.EvalUint64:
		return "bigint(20) unsigned"
	case types.EvalString:
		length := maxVarcharLength
		if length == 0 {
			length = math.MaxUint16
		}
		return fmt.Sprintf("varchar(%d)", length)
	default:
		return "<unknown>"
	}
}
