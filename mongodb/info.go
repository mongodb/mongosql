package mongodb

import (
	"strings"

	"github.com/10gen/sqlproxy/schema"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// DatabaseName is the name of a database.
type DatabaseName string

// CollectionName is the name of a collection.
type CollectionName string

// Info is the configuration of MongoDB at runtime.
type Info struct {
	buildInfo *mgo.BuildInfo

	// Git version is the git hash of the mongodb server.
	GitVersion string
	// Databases is information about databases by name.
	Databases map[DatabaseName]*DatabaseInfo
	// Privileges is a union of all the database privileges
	Privileges Privilege
	// Version is the version of mongodb server.
	Version string
}

// VersionAtLeast indicates whether the MongoDB version is at least the required version.
func (i *Info) VersionAtLeast(version ...int) bool {
	return i.buildInfo.VersionAtLeast(version...)
}

// DatabaseInfo is the configuration of a database in MongoDB.
type DatabaseInfo struct {
	// Nme is the name of the database.
	Name DatabaseName
	// Privileges is a union of all the collection privileges.
	Privileges Privilege
	// Collections is information about collections by name.
	Collections map[CollectionName]*CollectionInfo
}

// CollectionInfo is the configuration of a collection in MongoDB.
type CollectionInfo struct {
	// Name is the name of the collection.
	Name CollectionName
	// Privileges indicates what is allowed on the collection.
	Privileges Privilege
	// Collation is the default collation of the collection.
	Collation *mgo.Collation
}

// LoadInfo looks up information from MongoDB.
func LoadInfo(s *mgo.Session, sch *schema.Schema, requireAuth bool) (*Info, error) {
	session := s.Clone()
	defer session.Close()

	buildInfo, err := session.BuildInfo()
	if err != nil {
		return nil, err
	}

	dbs, err := loadDatabases(session, sch)
	if err != nil {
		return nil, err
	}

	i := &Info{
		buildInfo:  &buildInfo,
		Databases:  dbs,
		GitVersion: buildInfo.GitVersion,
		Version:    buildInfo.Version,
	}

	if requireAuth {
		err = i.loadAuthInfo(s)
		if err != nil {
			return nil, err
		}
	} else {
		i.setAllPrivileges(AllPrivileges)
	}

	return i, nil
}

func loadDatabases(s *mgo.Session, sch *schema.Schema) (map[DatabaseName]*DatabaseInfo, error) {
	dbInfos := make(map[DatabaseName]*DatabaseInfo, len(sch.RawDatabases))
	for _, dbSchema := range sch.RawDatabases {
		cols, err := loadCollections(s, dbSchema)
		if err != nil {
			return nil, err
		}
		dbInfo := &DatabaseInfo{
			Name:        DatabaseName(strings.ToLower(dbSchema.Name)),
			Collections: cols,
		}
		dbInfos[dbInfo.Name] = dbInfo
	}

	return dbInfos, nil
}

func loadCollections(s *mgo.Session, dbSchema *schema.Database) (map[CollectionName]*CollectionInfo, error) {

	var result struct {
		Collections []bson.Raw
		Cursor      struct {
			FirstBatch []bson.Raw "firstBatch"
			NextBatch  []bson.Raw "nextBatch"
			NS         string
			ID         int64
		}
	}

	err := s.DB(dbSchema.Name).Run(bson.D{{"listCollections", 1}, {"cursor", struct{}{}}}, &result)
	if err == nil {
		colInfos := make(map[CollectionName]*CollectionInfo)
		for _, t := range dbSchema.RawTables {
			name := CollectionName(t.CollectionName)
			if _, ok := colInfos[name]; ok {
				// Because multiple tables can be mapped to the same collection,
				// we can skip collections we've already included.
				continue
			}

			colInfos[name] = &CollectionInfo{
				Name: name,
			}
		}

		ns := strings.SplitN(result.Cursor.NS, ".", 2)
		iter := s.DB(ns[0]).C(ns[1]).NewIter(nil, result.Cursor.FirstBatch, result.Cursor.ID, nil)
		var colResult struct {
			Name    string
			Type    string
			Options struct {
				Collation *mgo.Collation
			}
		}

		for iter.Next(&colResult) {
			colInfo, ok := colInfos[CollectionName(colResult.Name)]
			if !ok {
				continue
			}

			colInfo.Collation = colResult.Options.Collation
		}
		err := iter.Err()
		if err != nil {
			return nil, err
		}
		err = iter.Close()
		if err != nil {
			return nil, err
		}
		return colInfos, nil
	}

	return nil, err
}
