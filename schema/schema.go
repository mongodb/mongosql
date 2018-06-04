package schema

import (
	"fmt"
	"sort"
	"sync"

	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema/drdl"
)

// Schema represents a relational schema.
type Schema struct {
	// alterations is a slice of alterations that should be applied to the schema
	// defined in databases in order to achieve the desired schema.
	alterations []*Alteration
	// databases is a slice of schemas for all of the databases in the schema.
	databases map[normalizedName]*Database
	// cachedSortedDatabases is the cached result of the last call to DatabasesSorted. If it
	// is non-nil when DatabasesSorted is called, it will be used to avoid
	// duplicating a potentially expensive sort. cachedSortedDatabases is invalidated (set to
	// nil) whenever the databases map is modified.
	cachedSortedDatabases []*Database
	cacheLock             sync.RWMutex
}

// New returns a new schema with the provided databases and alterations. The
// schema is built by adding each of the provided databases to the schema in
// order.
func New(dbs []*Database, alterations []*Alteration) (*Schema, error) {
	s := &Schema{
		databases: map[normalizedName]*Database{},
	}
	s.AddAlterations(alterations...)
	for _, db := range dbs {
		err := s.AddDatabase(db)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// NewFromDRDL returns a new schema that is built from the provided DRDL schema.
// Each database in the drdl schema is converted to a *Database and then added
// to the schema in order. If an error is encountered while building the schema,
// it is returned along with a nil schema.
func NewFromDRDL(lg log.Logger, drdl *drdl.Schema) (*Schema, error) {
	dbs := []*Database{}
	for _, drdlDb := range drdl.Databases {
		db, err := NewDatabaseFromDRDL(lg, drdlDb)
		if err != nil {
			return nil, err
		}
		dbs = append(dbs, db)
	}
	return New(dbs, nil)
}

// AddAlterations adds the provided alterations to the end of this Schema's list
// of alterations.
func (s *Schema) AddAlterations(alts ...*Alteration) {
	s.alterations = append(s.alterations, alts...)
}

// AddDatabase attempts to add the provided database to the schema. If the
// database's name conflicts with the name of an existing database, an error is
// returned.
func (s *Schema) AddDatabase(d *Database) error {
	db := s.Database(d.Name())
	if db != nil {
		return fmt.Errorf("database %q already exists in schema", d.Name())
	}

	key := normalizeSQLName(d.Name())
	s.databases[key] = d
	s.invalidateCachedSort()
	return nil
}

// Alterations returns this Schema's list of alterations.
func (s *Schema) Alterations() []*Alteration {
	return s.alterations
}

// Altered returns a new Schema that is equivalent to the current schema with
// its alterations applied. The returned schema will have an empty Alterations
// slice.
func (s *Schema) Altered() (*Schema, error) {
	if len(s.Alterations()) == 0 {
		return s, nil
	}

	newSchema := s.DeepCopy()
	for _, a := range s.Alterations() {
		err := a.alter(newSchema)
		if err != nil {
			return nil, fmt.Errorf("could not alter schema: %v", err)
		}
	}
	newSchema.alterations = nil
	s.invalidateCachedSort()
	return newSchema, nil
}

// cacheSortedDatabases caches the provided sorted slice of databases.
func (s *Schema) cacheSortedDatabases(dbs []*Database) {
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()

	s.cachedSortedDatabases = make([]*Database, len(dbs))
	copy(s.cachedSortedDatabases, dbs)
}

// Database gets the database in this Schema whose normalized SQLName matches
// the normalized form of the provided name. If no matching database exists in
// the schema, nil is returned.
func (s *Schema) Database(name string) *Database {
	key := normalizeSQLName(name)
	return s.databases[key]
}

// Databases returns a slice of all the databases in this Schema.
func (s *Schema) Databases() []*Database {
	dbs := []*Database{}
	for _, db := range s.databases {
		dbs = append(dbs, db)
	}
	return dbs
}

// DatabasesSorted returns a slice of all the databases in this Schema sorted in
// ascending order by name.
func (s *Schema) DatabasesSorted() []*Database {
	cache := s.getCachedSortedDatabases()
	if cache != nil {
		return cache
	}
	dbs := s.Databases()
	sort.Slice(dbs, func(i, j int) bool {
		return dbs[i].Name() < dbs[j].Name()
	})
	s.cacheSortedDatabases(dbs)
	return dbs
}

// DeepCopy returns a deep copy of this Schema.
func (s *Schema) DeepCopy() *Schema {
	if s == nil {
		return nil
	}

	dbs := map[normalizedName]*Database{}
	for key, db := range s.databases {
		dbs[key] = db.DeepCopy()
	}

	alts := []*Alteration{}
	for _, alt := range s.alterations {
		alts = append(alts, alt.DeepCopy())
	}

	return &Schema{
		alterations: alts,
		databases:   dbs,
	}
}

// Equals checks whether the provided Schema is equal to this Schema.
func (s *Schema) Equals(other *Schema) error {
	if s == other {
		return nil
	}
	if s == nil {
		return fmt.Errorf("this schema is nil, but other schema is non-nil")
	}
	if other == nil {
		return fmt.Errorf("this schema is non-nil, but other schema is nil")
	}
	if len(s.databases) != len(other.databases) {
		return fmt.Errorf(
			"this schema has %d databases, other has %d",
			len(s.databases), len(other.databases),
		)
	}
	for key, db := range s.databases {
		otherDb, ok := other.databases[key]
		if !ok {
			return fmt.Errorf("database %q missing from other schema", db.Name())
		}
		err := db.Equals(otherDb)
		if err != nil {
			return fmt.Errorf("databases with name %q not equal: %v", db.Name(), err)
		}
	}
	return nil
}

// getCachedSortedDatabases returns a shallow copy of this schema's cached sort.
func (s *Schema) getCachedSortedDatabases() []*Database {
	s.cacheLock.RLock()
	defer s.cacheLock.RUnlock()

	if s.cachedSortedDatabases == nil {
		return nil
	}
	dbs := make([]*Database, len(s.cachedSortedDatabases))
	copy(dbs, s.cachedSortedDatabases)
	return dbs
}

// invalidateCachedSort invalidate's this schema's currently cached sort.
func (s *Schema) invalidateCachedSort() {
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()

	s.cachedSortedDatabases = nil
}

// ToDRDL converts the schema to a drdl.Schema type.
func (s *Schema) ToDRDL() *drdl.Schema {
	drdlDatabases := []*drdl.Database{}
	for _, d := range s.DatabasesSorted() {
		drdlTables := []*drdl.Table{}
		for _, t := range d.TablesSorted() {
			drdlColumns := []*drdl.Column{}
			for _, c := range t.ColumnsSorted() {
				drdlColumns = append(drdlColumns, &drdl.Column{
					MongoName: c.mongoName,
					MongoType: string(c.mongoType),
					SQLName:   c.sqlName,
					SQLType:   string(c.sqlType),
				})
			}
			drdlTables = append(drdlTables, &drdl.Table{
				SQLName:   t.sqlName,
				MongoName: t.mongoName,
				Pipeline:  t.pipeline,
				Columns:   drdlColumns,
			})
		}
		drdlDatabases = append(drdlDatabases, &drdl.Database{
			Name:   d.name,
			Tables: drdlTables,
		})
	}
	return &drdl.Schema{Databases: drdlDatabases}
}

// Validate checks whether this Schema is valid, returning an error if not.
func (s *Schema) Validate() error {
	for _, d := range s.Databases() {
		err := d.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate database %q: %v", d.Name(), err)
		}
	}
	return nil
}
