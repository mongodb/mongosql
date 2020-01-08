package catalog

import (
	"context"
	"fmt"
	"sort"

	"github.com/10gen/mongoast/ast"
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
	fullText       bool
	constraintName string
}

type unwindPathsAndAliases struct {
	unwindPaths []string
	pathAliases map[string]string
}

// addForeignKeys augments the catalog with foreign key information.
func (b *catalogBuilder) addForeignKeys() {

	unwindPaths := make(map[string]map[string]unwindPathsAndAliases)

	// Iterate through each SQL table in each SQL database and construct the table's
	// foreign key.
	dbs, _ := b.catalog.Databases(context.Background())
	for _, db := range dbs {
		dbName := string(db.Name())

		// Sort the tables based on the number of unwinds in their pipeline.  We do this
		// so we will have already added a child table's parent by the time we encounter
		// the child.
		tables, _ := db.Tables(context.Background())
		sortTablesByUnwinds(tables)
		for _, tbl := range tables {
			mongoTable, ok := tbl.(*MongoTable)
			if !ok {
				continue
			}

			tblName := tbl.Name()
			collectionName := mongoTable.collectionName
			tableUnwindPaths, pathAliases := getUnwindPaths(mongoTable.Pipeline())

			// Get all unwind paths from the current collection, and add current unwind
			// path. We'll use the other unwind paths to find any parent tables.
			if unwindPaths[collectionName] == nil {
				unwindPaths[collectionName] = make(map[string]unwindPathsAndAliases)
			}
			currentCollectionUnwindPaths := unwindPaths[collectionName]
			currentCollectionUnwindPaths[tblName] = unwindPathsAndAliases{tableUnwindPaths, pathAliases}

			// Iterate through all the unwind paths in the current collection and see if
			// current table is unwound from any of those tables.
			for foreignTable, foreignPathsAndAliases := range currentCollectionUnwindPaths {

				// If we find its parent table, construct the foreign key. The foreign key
				// of the current table should contain all columns in the primary key of
				// its parent. The primary key of the parent table contains the _id field
				// of the underlying Mongo collection plus indexes of any arrays it was
				// unwound from. For a given SQL table unwound from an array, the index of
				// that array is stored in a column in that table, with the name given by
				// the path alias. Therefore we add _id + the path aliases of the unwind
				// paths of the parent table to the foreign key.
				if isUnwoundFrom(tableUnwindPaths, foreignPathsAndAliases.unwindPaths) {
					var columns []*results.Column
					localToForeignColumn := make(map[string]string)

					// First add _id, if it exists.
					column, err := mongoTable.Column(mongoPrimaryKey)
					if err == nil {
						columns = append(columns, column)
						localToForeignColumn[mongoPrimaryKey] = mongoPrimaryKey
					}
					// Then add the columns from the parent's unwind paths.
					for _, fieldName := range foreignPathsAndAliases.unwindPaths {
						localColumnName := pathAliases[fieldName]
						foreignColumnName := foreignPathsAndAliases.pathAliases[fieldName]
						localToForeignColumn[localColumnName] = foreignColumnName

						column, err := mongoTable.Column(localColumnName)
						if err != nil {
							panic(err)
						}
						columns = append(columns, column)
					}

					// Add the foreignKey if it is not empty. We might miss some
					// foreign keys due to a failure to have _id in a drdl file for
					// an array table, but it is the best we can do since the user does not
					// want to map the foreign key into the table.
					if len(localToForeignColumn) != 0 {
						// Construct the foreign key.
						constraintName := createForeignKeyName(dbName, tblName, foreignTable)
						newK := ForeignKey{
							columns:              columns,
							constraintName:       constraintName,
							foreignDatabase:      dbName,
							foreignTable:         foreignTable,
							localToForeignColumn: localToForeignColumn,
						}
						// Add the foreign key to the table.
						mongoTable.foreignKeys = append(mongoTable.foreignKeys, newK)
					}
				}
			}
		}
	}
}

func isUnwoundFrom(childUnwinds []string, parentUnwinds []string) bool {
	if len(parentUnwinds) != len(childUnwinds)-1 {
		return false
	}
	lastUnwindInd := len(childUnwinds) - 1
	for i := 0; i < lastUnwindInd; i++ {
		if childUnwinds[i] != parentUnwinds[i] {
			return false
		}
	}
	return true
}

func numUnwinds(pipeline *ast.Pipeline) int {
	numUnwinds := 0
	for _, stage := range pipeline.Stages {
		if _, ok := stage.(*ast.UnwindStage); ok {
			numUnwinds++
		}
	}
	return numUnwinds
}

func sortTablesByUnwinds(tables []Table) {
	unwindCounts := make(map[*ast.Pipeline]int)
	for _, table := range tables {
		mongoTable, ok := table.(*MongoTable)
		if !ok {
			continue
		}
		unwindCounts[mongoTable.pipeline] = numUnwinds(mongoTable.pipeline)
	}

	sort.SliceStable(tables, func(i, j int) bool {
		tableI, ok := tables[i].(*MongoTable)
		if !ok {
			return false
		}
		tableJ, ok := tables[j].(*MongoTable)
		if !ok {
			return false
		}
		return unwindCounts[tableI.pipeline] < unwindCounts[tableJ.pipeline]
	})
}

func (b *catalogBuilder) getRowForForeignKey(tableName, aliasName string, ck key, fk ForeignKey, position int, columnNames []string) results.Row {
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
		return newInfoRow(aliasName,
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
		return newInfoRow(aliasName,
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
		return newInfoRow(aliasName,
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
		return newInfoRow(aliasName,
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

func (b *catalogBuilder) getRowIterForPrimaryKey(tableName, aliasName string, ck key, primaryKeys results.Columns, columnNames []string) results.RowIter {
	pkConstraintName, pkConstraintType := "PRIMARY", "PRIMARY KEY"
	catalog, database, table := ck.catalog, ck.database, ck.table

	nullv, strv, intv := getValueCreators(b.variables)
	checkColumnNumber := func(num int) {
		if len(columnNames) != num {
			panic(fmt.Sprintf("table: %s must be passed %d columnNames, but got %d", tableName, num, len(columnNames)))
		}
	}

	rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
	done := make(chan struct{})

	go func() {
		defer close(rowChan)
		for position, key := range primaryKeys {
			switch tableName {
			case KeyColumnUsageTable:
				checkColumnNumber(12)
				select {
				case rowChan <- newInfoRow(aliasName,
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
				):
				case <-done:
					return
				}
			case ReferentialConstraintsTable:
			case StatisticsTable:
				checkColumnNumber(16)
				select {
				case rowChan <- newInfoRow(aliasName,
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
				):
				case <-done:
					return
				}
			case TableConstraintsTable:
				checkColumnNumber(6)
				// table constraints should only have one entry
				// per key (simple/compound) relationship
				rowChan <- newInfoRow(aliasName,
					strv(columnNames[0], catalog),
					strv(columnNames[1], database),
					strv(columnNames[2], pkConstraintName),
					strv(columnNames[3], database),
					strv(columnNames[4], table),
					strv(columnNames[5], pkConstraintType),
				)
				return
			default:
				panic(fmt.Sprintf("unknown primary key table: %v", tableName))
			}
		}
	}()

	return results.NewRowChanIter(rowChan, done)
}

func (b *catalogBuilder) getRowIterForUniqueIndexes(tableName, aliasName string, ck key, indexes []Index, columnNames []string) results.RowIter {
	uniqueKeyConstraint, catalog := "UNIQUE", ck.catalog
	database, table := ck.database, ck.table

	position := 0

	nullv, strv, intv := getValueCreators(b.variables)
	checkColumnNumber := func(num int) {
		if len(columnNames) != num {
			panic(fmt.Sprintf("table: %s must be passed %d columnNames, but got %d", tableName, num, len(columnNames)))
		}
	}

	rowChan := make(chan results.Row, results.DefaultRowChannelBufSize)
	done := make(chan struct{})
	go func() {
		defer close(rowChan)
		for _, index := range indexes {
			if !index.unique {
				continue
			}
			position++
			switch tableName {
			case KeyColumnUsageTable:
				for ordinalPosition, column := range index.columns {
					checkColumnNumber(12)
					select {
					case rowChan <- newInfoRow(aliasName,
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
					):
					case <-done:
						return
					}
				}
			case ReferentialConstraintsTable:
			case StatisticsTable:
				for ordinalPosition, column := range index.columns {
					checkColumnNumber(16)
					select {
					case rowChan <- newInfoRow(aliasName,
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
					):
					case <-done:
						return
					}
				}
			case TableConstraintsTable:
				checkColumnNumber(6)
				select {
				case rowChan <- newInfoRow(aliasName,
					strv(columnNames[0], catalog),
					strv(columnNames[1], database),
					strv(columnNames[2], createUniqueIndexName(ck.database, ck.table, position)),
					strv(columnNames[3], database),
					strv(columnNames[4], table),
					strv(columnNames[5], uniqueKeyConstraint),
				):
				case <-done:
					return
				}
			default:
				panic(fmt.Sprintf("unknown unique key table: %v", tableName))
			}
		}
	}()

	return results.NewRowChanIter(rowChan, done)
}
