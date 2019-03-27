package catalog

import (
	"fmt"

	"github.com/10gen/sqlproxy/evaluator/results"
)

// key is an identifier for a
// catalog table column.
type key struct {
	catalog  string
	database string
	table    string
	column   string
}

type namespace struct {
	database string
	table    string
}

type namespaces []namespace

var (
	mongoPrimaryKey  = "_id"
	primaryKey       = "PRI"
	uniqueKey        = "UNI"
	multiKey         = "MUL"
	defaultIndexType = "BTREE"
)

// Index represents an index in a SQL table.
type Index struct {
	columns        results.Columns
	unique         bool
	constraintName string
}

// foreignKeyCandidate is a struct used to represent
// all columns (with names matching an unwind path)
// that may possibly be foreign keys - more on this
// in the comment for the addForeignKeys function.
type foreignKeyCandidate struct {
	name     string
	table    string
	database string
	// depth is the number of unwinds
	// present in this candidate's table.
	depth int
}

func newForeignKeyCandidate(name, table, database string, depth int) *foreignKeyCandidate {
	return &foreignKeyCandidate{name, table, database, depth}
}

type foreignKeyCandidates []*foreignKeyCandidate
type potentialForeignKeys map[string]foreignKeyCandidates

// addForeignKeys augments the catalog with foreign key information.
// It comprises a 2-phase process:
//
// Phase 1: all candidate columns that could reference a given column
// are generated.
//
// Phase 2: all such candidate columns are then whittled down to those
// that meet the foreign key constraints - each candidate column in
// a given table:
// - must be at most one table depth greater than the referenced
// column's table (this caters to how we represent foreign keys in the
// BIC - constrained to parent/child relationships)
// - must not appear within a table containing a sibling paths i.e. if
// the base table is "foo", columns in the child tables - "foo_b" and
// "foo_c" - are not allowed as foreign keys.
func (b *catalogBuilder) addForeignKeys() {
	b.includeForeignKeys(b.generateForeignKeyCandidates())
}

// generateForeignKeyCandidates returns two maps from the namespaces
// present in the catalog:
// - one that groups namespaces by collection name and
// - another that groups potential foreign keys by the collections they
// come from.
func (b *catalogBuilder) generateForeignKeyCandidates() (map[string]namespaces,
	map[string]potentialForeignKeys) {

	collectionLineage := make(map[string]namespaces)
	candidateForeignKeys := make(map[string]potentialForeignKeys)

	for _, db := range b.catalog.Databases() {
		dbName := string(db.Name())
		for _, tbl := range db.Tables() {
			mongoTable, ok := tbl.(*MongoTable)
			if !ok {
				continue
			}

			collectionName := mongoTable.collectionName
			tableName := tbl.Name()
			ns := namespace{dbName, tableName}
			collectionLineage[collectionName] = append(collectionLineage[collectionName], ns)

			unwindPaths, pathAliases := getUnwindPaths(mongoTable.Pipeline())

			depth := len(unwindPaths)

			// Sibling paths are problematic because the number of unwinds don't correspond
			// to how deep the table is nested. Thus, such paths are ignored as candidates.
			if depth > 1 && containsSiblingPaths(unwindPaths) {
				continue
			}

			// Since the "_id" field is always a foreign key candidate,
			// we include it for subsequent consideration.
			pathAliases[mongoPrimaryKey] = mongoPrimaryKey

			// All foreign key candidate columns from mongoTable (determined
			// by column names that match an unwind path name are added to
			// the foreign key candidates for this table.
			for _, column := range mongoTable.columns {
				if unwindPath, ok := pathAliases[column.MongoName]; ok {
					fkName := column.Name
					fkCandidate := newForeignKeyCandidate(fkName, tableName, dbName, depth)
					if _, ok := candidateForeignKeys[collectionName]; !ok {
						candidateForeignKeys[collectionName] = map[string]foreignKeyCandidates{
							unwindPath: []*foreignKeyCandidate{fkCandidate},
						}
					}
					candidateForeignKeys[collectionName][unwindPath] = append(
						candidateForeignKeys[collectionName][unwindPath], fkCandidate)
				}
			}
		}
	}

	return collectionLineage, candidateForeignKeys
}

// includeForeignKeys whittles down the candidate foreign keys to those that actually meet
// the foreign key criteria:
// - must be at most one table depth greater than the referenced column's table (this caters
// to how we represent foreign keys in the BIC - constrained to parent/child relationships)
// - must not appear within a table containing a sibling paths i.e. if the base table is
// "foo", columns in the child tables - "foo_b" and "foo_c" - are not allowed as foreign keys.
// and then adds all such keys to the tables they reference.
func (b *catalogBuilder) includeForeignKeys(collectionLineage map[string]namespaces,
	candidateForeignKeys map[string]potentialForeignKeys) {

	sortForeignKeyCandidates(candidateForeignKeys)

	for collection, namespaces := range collectionLineage {
		for _, namespace := range namespaces {
			currentTable, err := b.getTableFromNamespace(namespace)
			if err != nil {
				continue
			}

			dbName, tableName := namespace.database, namespace.table
			mongoTable, ok := currentTable.(*MongoTable)

			if !ok {
				continue
			} else if len(mongoTable.Pipeline().Stages) == 0 {
				continue
			}

			unwindPaths, pathAliases := getUnwindPaths(mongoTable.Pipeline())
			depth := len(unwindPaths)
			if depth > 1 && containsSiblingPaths(unwindPaths) {
				continue
			}

			// We add the mongoPrimaryKey to the unwindPaths
			// as it's equally a foreign key candidate.
			pathAliases[mongoPrimaryKey] = mongoPrimaryKey
			tableToForeignKey := make(map[string]ForeignKey)

			for _, column := range mongoTable.columns {
				if unwindPath, ok := pathAliases[column.MongoName]; ok {
					fkCandidates, ok := candidateForeignKeys[collection][unwindPath]
					if !ok {
						continue
					}

					foreignKey := getKeyToParentTable(fkCandidates, depth)
					if foreignKey == nil {
						continue
					}

					fkColumn := foreignKey.name
					fkTable := foreignKey.table
					fkDatabase := foreignKey.database
					constraintName := createForeignKeyName(dbName, tableName, fkTable)

					// if there isn't a foreign key pointing to this table, create one.
					if foreignKey, ok := tableToForeignKey[fkTable]; !ok {
						newK := NewForeignKey(column, constraintName, fkDatabase, fkTable, fkColumn)
						tableToForeignKey[fkTable] = newK
					} else {
						// If this table already has a foreign key, add the current key
						// to generate a compound foreign key.
						foreignKey.columns = append(foreignKey.columns, column)
						foreignKey.localToForeignColumn[column.Name] = fkColumn
						tableToForeignKey[fkTable] = foreignKey
					}
				}
			}

			for _, foreignKey := range tableToForeignKey {
				mongoTable.foreignKeys = append(mongoTable.foreignKeys, foreignKey)
			}
		}
	}
}

func (b *catalogBuilder) getRowForForeignKey(tableName string, ck key, fk ForeignKey, position int, columnNames []string) results.Row {
	catalog, database, table, column := ck.catalog, ck.database, ck.table, ck.column
	constraintName, foreignColumn := fk.constraintName, fk.localToForeignColumn[column]
	foreignDatabase, foreignTable := fk.foreignDatabase, fk.foreignTable
	foreignKeyConstraintName := "FOREIGN KEY"

	checkColumnNumber := func(num int) {
		if len(columnNames) != num {
			panic(fmt.Sprintf("table: %s must be passed %d columnNames, but got %d", tableName, num, len(columnNames)))
		}
	}

	nullv, strv, intv := getValueCreators(b.variables)
	switch tableName {
	case KeyColumnUsageTable:
		checkColumnNumber(12)
		return newInfoRow(tableName,
			strv(columnNames[0], catalog),
			strv(columnNames[1], database),
			strv(columnNames[2], constraintName),
			strv(columnNames[3], catalog),
			strv(columnNames[4], database),
			strv(columnNames[5], table),
			strv(columnNames[6], column),
			intv(columnNames[7], int64(position)),
			intv(columnNames[8], int64(position)),
			strv(columnNames[9], foreignDatabase),
			strv(columnNames[10], foreignTable),
			strv(columnNames[11], foreignColumn),
		)
	case ReferentialConstraintsTable:
		checkColumnNumber(11)
		return newInfoRow(tableName,
			strv(columnNames[0], catalog),
			strv(columnNames[1], database),
			strv(columnNames[2], constraintName),
			strv(columnNames[3], catalog),
			strv(columnNames[4], database),
			strv(columnNames[5], "PRIMARY"),
			strv(columnNames[6], "NONE"),
			strv(columnNames[7], "CASCADE"),
			strv(columnNames[8], "CASCADE"),
			strv(columnNames[9], table),
			strv(columnNames[10], foreignTable),
		)
	case StatisticsTable:
		checkColumnNumber(16)
		return newInfoRow(tableName,
			strv(columnNames[0], catalog),
			strv(columnNames[1], database),
			strv(columnNames[2], table),
			intv(columnNames[3], 1),
			strv(columnNames[4], database),
			strv(columnNames[5], constraintName),
			intv(columnNames[6], int64(position)),
			strv(columnNames[7], column),
			strv(columnNames[8], "A"),
			intv(columnNames[9], 0),
			nullv(columnNames[10]),
			nullv(columnNames[11]),
			strv(columnNames[12], "YES"),
			strv(columnNames[13], defaultIndexType),
			strv(columnNames[14], ""),
			strv(columnNames[15], ""),
		)
	case TableConstraintsTable:
		checkColumnNumber(6)
		return newInfoRow(tableName,
			strv(columnNames[0], catalog),
			strv(columnNames[1], database),
			strv(columnNames[2], constraintName),
			strv(columnNames[3], database),
			strv(columnNames[4], table),
			strv(columnNames[5], foreignKeyConstraintName),
		)
	}
	panic(fmt.Sprintf("unknown foreign key table: %v", tableName))
}

func (b *catalogBuilder) getRowsForPrimaryKey(tableName string, ck key, primaryKeys results.Columns, columnNames []string) results.Rows {
	pkConstraintName, pkConstraintType := "PRIMARY", "PRIMARY KEY"
	catalog, database, table := ck.catalog, ck.database, ck.table

	var rows results.Rows

	nullv, strv, intv := getValueCreators(b.variables)
	checkColumnNumber := func(num int) {
		if len(columnNames) != num {
			panic(fmt.Sprintf("table: %s must be passed %d columnNames, but got %d", tableName, num, len(columnNames)))
		}
	}
	for position, key := range primaryKeys {
		switch tableName {
		case KeyColumnUsageTable:
			checkColumnNumber(12)
			rows = append(rows, newInfoRow(tableName,
				strv(columnNames[0], catalog),
				strv(columnNames[1], database),
				strv(columnNames[2], pkConstraintName),
				strv(columnNames[3], catalog),
				strv(columnNames[4], database),
				strv(columnNames[5], table),
				strv(columnNames[6], key.Name),
				intv(columnNames[7], int64(position+1)),
				nullv(columnNames[8]),  // position in unique constraint
				nullv(columnNames[9]),  // referenced table schema
				nullv(columnNames[10]), // referenced table name
				nullv(columnNames[11]), // referenced column name
			))
		case ReferentialConstraintsTable:
		case StatisticsTable:
			checkColumnNumber(16)
			rows = append(rows, newInfoRow(tableName,
				strv(columnNames[0], catalog),
				strv(columnNames[1], database),
				strv(columnNames[2], table),
				intv(columnNames[3], 0),
				strv(columnNames[4], database),
				strv(columnNames[5], pkConstraintName),
				intv(columnNames[6], int64(position+1)),
				strv(columnNames[7], key.Name),
				strv(columnNames[8], "A"),
				intv(columnNames[9], 0),
				nullv(columnNames[10]),
				nullv(columnNames[11]),
				strv(columnNames[12], "YES"),
				strv(columnNames[13], defaultIndexType),
				strv(columnNames[14], ""),
				strv(columnNames[15], ""),
			))
		case TableConstraintsTable:
			checkColumnNumber(6)
			// table constraints should only have one entry
			// per key (simple/compound) relationship
			rows = append(rows, newInfoRow(tableName,
				strv(columnNames[0], catalog),
				strv(columnNames[1], database),
				strv(columnNames[2], pkConstraintName),
				strv(columnNames[3], database),
				strv(columnNames[4], table),
				strv(columnNames[5], pkConstraintType),
			))
			return rows
		default:
			panic(fmt.Sprintf("unknown primary key table: %v", tableName))
		}
	}

	return rows
}

func (b *catalogBuilder) getRowsForUniqueIndexes(tableName string, ck key, indexes []Index, columnNames []string) results.Rows {
	uniqueKeyConstraint, catalog := "UNIQUE", ck.catalog
	database, table := ck.database, ck.table

	position := 0
	var rows results.Rows

	nullv, strv, intv := getValueCreators(b.variables)
	checkColumnNumber := func(num int) {
		if len(columnNames) != num {
			panic(fmt.Sprintf("table: %s must be passed %d columnNames, but got %d", tableName, num, len(columnNames)))
		}
	}

	for _, index := range indexes {
		if !index.unique {
			continue
		}
		position++
		switch tableName {
		case KeyColumnUsageTable:
			for ordinalPosition, column := range index.columns {
				checkColumnNumber(12)
				rows = append(rows, newInfoRow(tableName,
					strv(columnNames[0], catalog),
					strv(columnNames[1], database),
					strv(columnNames[2], createUniqueIndexName(ck.database, ck.table, position)),
					strv(columnNames[3], catalog),
					strv(columnNames[4], database),
					strv(columnNames[5], table),
					strv(columnNames[6], column.Name),
					intv(columnNames[7], int64(ordinalPosition+1)),
					nullv(columnNames[8]),  // position in unique constraint
					nullv(columnNames[9]),  // referenced table schema
					nullv(columnNames[10]), // referenced table name
					nullv(columnNames[11]), // referenced column name
				))
			}
		case ReferentialConstraintsTable:
		case StatisticsTable:
			for ordinalPosition, column := range index.columns {
				checkColumnNumber(16)
				rows = append(rows, newInfoRow(tableName,
					strv(columnNames[0], catalog),
					strv(columnNames[1], database),
					strv(columnNames[2], table),
					intv(columnNames[3], 0),
					strv(columnNames[4], database),
					strv(columnNames[5], column.Name),
					intv(columnNames[6], int64(ordinalPosition+1)),
					strv(columnNames[7], column.Name),
					strv(columnNames[8], "A"),
					intv(columnNames[9], 0),
					nullv(columnNames[10]),
					nullv(columnNames[11]),
					strv(columnNames[12], "YES"),
					strv(columnNames[13], defaultIndexType),
					strv(columnNames[14], ""),
					strv(columnNames[15], ""),
				))
			}
		case TableConstraintsTable:
			checkColumnNumber(6)
			rows = append(rows, newInfoRow(tableName,
				strv(columnNames[0], catalog),
				strv(columnNames[1], database),
				strv(columnNames[2], createUniqueIndexName(ck.database, ck.table, position)),
				strv(columnNames[3], database),
				strv(columnNames[4], table),
				strv(columnNames[5], uniqueKeyConstraint),
			))
		default:
			panic(fmt.Sprintf("unknown unique key table: %v", tableName))
		}
	}
	return rows
}
