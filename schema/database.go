package schema

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema/drdl"
)

// Database represents the schema for a database.
type Database struct {
	// name is the name of the database.
	name string
	// tables is a map of normalized names to tables in the database.
	tables map[normalizedName]*Table
	// cachedSort is the cached result of the last call to TablesSorted. If it is
	// non-nil when TablesSorted is called, it will be used to avoid duplicating a
	// potentially expensive sort. cachedSort is invalidated (set to nil) whenever
	// the tables map is modified.
	cachedSort []*Table
	cacheLock  sync.RWMutex
}

// NewDatabase returns a new Database with the provided name and tables. The
// database is built by adding each of the provided tables to the database in
// order.
func NewDatabase(lg *log.Logger, name string, tables []*Table) *Database {
	db := &Database{
		name:   name,
		tables: map[normalizedName]*Table{},
	}
	for _, tbl := range tables {
		db.AddTable(lg, tbl)
	}
	return db
}

// NewDatabaseFromDRDL returns a new Database that is built from the provided DRDL
// Database. Each table in the drdl database is converted to a *Table and then
// added to the schema in order. If an error is encountered while building the
// database, it is returned along with a nil database.
func NewDatabaseFromDRDL(lg *log.Logger, drdlDb *drdl.Database) (*Database, error) {
	tbls := []*Table{}
	for _, dtbl := range drdlDb.Tables {
		tbl, err := NewTableFromDRDL(lg, dtbl)
		if err != nil {
			return nil, err
		}
		tbls = append(tbls, tbl)
	}
	return NewDatabase(lg, drdlDb.Name, tbls), nil
}

// AddTable adds the provided table to the database. If the table's name
// conflicts with the name of an existing table, its name will be changed to
// something that is unique within the database.
func (d *Database) AddTable(lg *log.Logger, t *Table) {
	tbl := d.Table(t.SQLName())
	if tbl != nil {
		initName := t.SQLName()
		t.sqlName = d.uniqueTableName(t.SQLName())
		if t.SQLName() != initName {
			lg.Warnf(log.Dev, "found 2 namespaces with the same case-insensitive "+
				"name: renamed %q to %q", initName, t.SQLName())
		}
	}

	key := normalizeSQLName(t.SQLName())
	d.tables[key] = t
	d.invalidateCachedSort()
}

// cacheSort caches the provided sorted slice of tables.
func (d *Database) cacheSort(tbls []*Table) {
	d.cacheLock.Lock()
	defer d.cacheLock.Unlock()

	d.cachedSort = make([]*Table, len(tbls))
	copy(d.cachedSort, tbls)
}

// DeepCopy returns a deep copy of this Database.
func (d *Database) DeepCopy() *Database {
	if d == nil {
		return nil
	}

	tables := map[normalizedName]*Table{}
	for key, tbl := range d.tables {
		tables[key] = tbl.DeepCopy()
	}

	return &Database{
		name:   d.name,
		tables: tables,
	}
}

// Equals checks whether the provided Database is equal to this Database.
func (d *Database) Equals(other *Database) error {
	if d == other {
		return nil
	}
	if d == nil {
		return fmt.Errorf("this db is nil, but other db is non-nil")
	}
	if other == nil {
		return fmt.Errorf("this db is non-nil, but other db is nil")
	}
	if d.Name() != other.Name() {
		return fmt.Errorf("db names %q and %q do not match", d.Name(), other.Name())
	}
	if len(d.tables) != len(other.tables) {
		return fmt.Errorf("this db has %d tables, other has %d", len(d.tables), len(other.tables))
	}
	for key, table := range d.tables {
		otherTable, ok := other.tables[key]
		if !ok {
			return fmt.Errorf("table %q missing from other schema", table.SQLName())
		}
		err := table.Equals(otherTable)
		if err != nil {
			return fmt.Errorf("tables with sqlName %q not equal: %v", table.SQLName(), err)
		}
	}

	return nil
}

// getCachedSort returns a shallow copy of this database's cached sort.
func (d *Database) getCachedSort() []*Table {
	d.cacheLock.RLock()
	defer d.cacheLock.RUnlock()

	if d.cachedSort == nil {
		return nil
	}
	tbls := make([]*Table, len(d.cachedSort))
	copy(tbls, d.cachedSort)
	return tbls
}

// invalidateCachedSort invalidate's this database's currently cached sort.
func (d *Database) invalidateCachedSort() {
	d.cacheLock.Lock()
	defer d.cacheLock.Unlock()

	d.cachedSort = nil
}

// Name returns the name of this database.
func (d *Database) Name() string {
	return d.name
}

// PostProcess removes empty tables from this database's schema, then calls
// PostProcess on all of the remaining tables in the databaese, passing in the
// provided valus of preJoin.
func (d *Database) PostProcess(lg *log.Logger, preJoin bool) {
	for key, table := range d.tables {
		if len(table.Columns()) == 0 {
			delete(d.tables, key)
		}
		table.PostProcess(lg, preJoin)
	}
	d.invalidateCachedSort()
}

// Table gets the table in this Database whose normalized SQLName matches the
// normalized form of the provided name.  If no matching table exists in the
// database, nil is returned.
func (d *Database) Table(name string) *Table {
	key := normalizeSQLName(name)
	return d.tables[key]
}

// Tables returns a slice of all the tables in this Database.
func (d *Database) Tables() []*Table {
	tbls := []*Table{}
	for _, tbl := range d.tables {
		tbls = append(tbls, tbl)
	}
	return tbls
}

// TablesSorted returns a slice of all the tables in this Database sorted in
// ascending order by name.
func (d *Database) TablesSorted() []*Table {
	cache := d.getCachedSort()
	if cache != nil {
		return cache
	}

	keys := []normalizedName{}
	for key := range d.tables {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	tbls := []*Table{}
	for _, key := range keys {
		tbls = append(tbls, d.tables[key])
	}

	d.cacheSort(tbls)
	return tbls
}

// uniqueTableName returns a version of the provided SQLName that is unique
// within this database.
func (d *Database) uniqueTableName(tableName string) string {
	retTableName := tableName
	i := 0
	for {
		tbl := d.Table(retTableName)
		if tbl != nil {
			retTableName = fmt.Sprintf("%v_%v", tableName, i)
			i++
			continue
		}
		return retTableName
	}
}

// Validate checks whether this Database is valid, returning an error if not.
func (d *Database) Validate() error {
	tmap := make(map[string]struct{})
	for _, t := range d.Tables() {
		err := t.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate table %q: %v", t.SQLName(), err)
		}

		key := strings.ToLower(t.SQLName())
		if _, ok := tmap[key]; ok {
			return fmt.Errorf("duplicated name for table %q", t.SQLName())
		}
		tmap[key] = struct{}{}
	}

	return nil
}
