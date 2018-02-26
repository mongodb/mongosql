package catalog

import (
	"fmt"
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

const (
	mongoPrimaryKey  string = "_id"
	primaryKey       string = "PRI"
	uniqueKey        string = "UNI"
	multiKey         string = "MUL"
	defaultIndexType string = "BTREE"
)

// Index represents an index in a SQL table.
type Index struct {
	columns        []Column
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
func (b *catalogBuilder) generateForeignKeyCandidates() (map[string]namespaces, map[string]potentialForeignKeys) {
	collectionLineage := make(map[string]namespaces)
	candidateForeignKeys := make(map[string]potentialForeignKeys)

	for _, db := range b.catalog.Databases() {
		for _, tbl := range db.Tables() {
			mongoTable, ok := tbl.(*MongoTable)
			if !ok {
				continue
			}

			collectionName, dbName, tableName := mongoTable.CollectionName, string(db.Name), string(tbl.Name())
			ns := namespace{dbName, tableName}
			collectionLineage[collectionName] = append(collectionLineage[collectionName], ns)

			unwindPaths, pathAliases := getUnwindPaths(mongoTable.Pipeline)

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
					fkName := string(column.name)
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
func (b *catalogBuilder) includeForeignKeys(collectionLineage map[string]namespaces, candidateForeignKeys map[string]potentialForeignKeys) {

	sortForeignKeyCandidates(candidateForeignKeys)

	for collection, namespaces := range collectionLineage {
		for _, namespace := range namespaces {
			currentTable, _ := b.getTableFromNamespace(namespace)
			dbName, tableName := namespace.database, namespace.table
			mongoTable, ok := currentTable.(*MongoTable)

			if !ok {
				continue
			} else if len(mongoTable.Pipeline) == 0 {
				continue
			}

			unwindPaths, pathAliases := getUnwindPaths(mongoTable.Pipeline)
			depth := len(unwindPaths)
			if depth > 1 && containsSiblingPaths(unwindPaths) {
				continue
			}

			// We add the mongoPrimaryKey to the unwindPaths
			// as it's equally a foreign key candidate.
			pathAliases[mongoPrimaryKey] = mongoPrimaryKey
			tableToForeignKey := make(map[string]ForeignKey)

			for _, column := range mongoTable.columns {
				if unwindPath, ok := pathAliases[string(column.MongoName)]; ok {
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
					constraintName := createForeignKeyName(dbName, tableName, fkTable, 1)

					// if there isn't a foreign key pointing to this table, create one.
					if foreignKey, ok := tableToForeignKey[fkTable]; !ok {
						newFk := NewForeignKey(column, constraintName, fkDatabase, fkTable, fkColumn)
						tableToForeignKey[fkTable] = newFk
					} else {
						// If this table already has a foreign key, add the current key
						// to generate a compound foreign key.
						foreignKey.columns = append(foreignKey.columns, column)
						foreignKey.localToForeignColumn[string(column.name)] = fkColumn
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

func getDataRowForForeignKey(tableType string, ck key, fk ForeignKey, position int) *DataRow {
	catalog, database, table, column := ck.catalog, ck.database, ck.table, ck.column
	constraintName, foreignColumn := fk.constraintName, fk.localToForeignColumn[column]
	foreignDatabase, foreignTable := fk.foreignDatabase, fk.foreignTable
	foreignKeyConstraintName := "FOREIGN KEY"

	switch tableType {
	case "KEY_COLUMN_USAGE":
		return NewDataRow(catalog, database, constraintName, catalog, database, table,
			column, position, position, foreignDatabase, foreignTable, foreignColumn)
	case "REFERENTIAL_CONSTRAINTS":
		return NewDataRow(catalog, database, constraintName, catalog, database,
			"PRIMARY", "NONE", "CASCADE", "CASCADE", table, foreignTable)
	case "STATISTICS":
		return NewDataRow(
			catalog,
			database,
			table,
			1,
			database,
			constraintName,
			position,
			column,
			"A",
			0,
			nil,
			nil,
			"YES",
			defaultIndexType,
			"",
			"",
		)
	case "TABLE_CONSTRAINTS":
		return NewDataRow(catalog, database, constraintName, database, table, foreignKeyConstraintName)
	}
	panic(fmt.Sprintf("unknown foreign key table type: %v", tableType))
}

func getDataRowsForPrimaryKey(tableType string, ck key, primaryKeys []Column) []*DataRow {
	pkConstraintName, pkConstraintType := "PRIMARY", "PRIMARY KEY"
	catalog, database, table := ck.catalog, ck.database, ck.table

	var rows []*DataRow

	for position, key := range primaryKeys {
		switch tableType {
		case "KEY_COLUMN_USAGE":
			rows = append(rows, NewDataRow(
				catalog,
				database,
				pkConstraintName,
				catalog,
				database,
				table,
				string(key.Name()),
				position+1,
			))
		case "REFERENTIAL_CONSTRAINTS":
		case "STATISTICS":
			rows = append(rows, NewDataRow(
				catalog,
				database,
				table,
				0,
				database,
				pkConstraintName,
				position+1,
				string(key.Name()),
				"A",
				0,
				nil,
				nil,
				"YES",
				defaultIndexType,
				"",
				"",
			))
		case "TABLE_CONSTRAINTS":
			// table constraints should only have one entry
			// per key (simple/compound) relationship
			rows = append(rows, NewDataRow(
				catalog,
				database,
				pkConstraintName,
				database,
				table,
				pkConstraintType,
			))
			return rows
		default:
			panic(fmt.Sprintf("unknown primary key table type: %v", tableType))
		}
	}

	return rows
}

func getDataRowsForUniqueIndexes(tableType string, ck key, indexes []Index) []*DataRow {
	position, uniqueKeyConstraint := 0, "UNIQUE"
	catalog, database, table := ck.catalog, ck.database, ck.table

	var rows []*DataRow

	for _, index := range indexes {
		if !index.unique {
			continue
		}
		position++
		switch tableType {
		case "KEY_COLUMN_USAGE":
			for ordinalPosition, column := range index.columns {
				rows = append(rows, NewDataRow(
					catalog,
					database,
					createUniqueIndexName(database, table, position),
					catalog,
					database,
					table,
					string(column.Name()),
					ordinalPosition+1,
				))
			}
		case "REFERENTIAL_CONSTRAINTS":
		case "STATISTICS":
			for ordinalPosition, column := range index.columns {
				rows = append(rows, NewDataRow(
					catalog,
					database,
					table,
					0,
					database,
					string(column.Name()),
					ordinalPosition+1,
					string(column.Name()),
					"A",
					0,
					nil,
					nil,
					"YES",
					defaultIndexType,
					"",
					"",
				))
			}
		case "TABLE_CONSTRAINTS":
			rows = append(rows, NewDataRow(
				catalog,
				database,
				createUniqueIndexName(database, table, position),
				database,
				table,
				uniqueKeyConstraint,
			))
		default:
			panic(fmt.Sprintf("unknown unique key table type: %v", tableType))
		}
	}
	return rows
}
