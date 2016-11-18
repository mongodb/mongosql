package mongodb

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/log"
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
	// Git version is the git hash of the mongodb server.
	GitVersion string
	// Databases is information about databases by name.
	Databases map[DatabaseName]*DatabaseInfo
	// Privileges is a union of all the database privileges
	Privileges Privilege
	// Version is the version of mongodb server.
	Version string
	// VersionArray are the components of the version.
	VersionArray []int
}

// VersionAtLeast indicates whether the MongoDB version is at least the required version.
func (i *Info) VersionAtLeast(version ...int) bool {
	for idx, vi := range version {
		if idx == len(i.VersionArray) {
			return false
		}
		if ivi := i.VersionArray[idx]; ivi != vi {
			return ivi >= vi
		}
	}
	return true
}

// DatabaseInfo is the configuration of a database in MongoDB.
type DatabaseInfo struct {
	caseSensitiveName string

	// Name is the name of the database.
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
	// IsView is true if the collection is a MongoDB view
	// and false otherwise.
	IsView bool
}

// LoadInfo looks up information from MongoDB.
func LoadInfo(logger *log.Logger, session *mgo.Session, config *schema.Schema, requireAuth bool) (*Info, error) {
	buildInfo, err := session.BuildInfo()
	if err != nil {
		return nil, err
	}
	logger.Logf(log.Info, "connected to MongoDB %v, %v", buildInfo.Version, buildInfo.GitVersion)

	dbs := createDatabasesFromSchema(config)

	i := &Info{
		Databases:    dbs,
		GitVersion:   buildInfo.GitVersion,
		Version:      buildInfo.Version,
		VersionArray: buildInfo.VersionArray,
	}

	if requireAuth {
		err = i.loadAuthInfo(logger, session)
		if err != nil {
			return nil, err
		}
	} else {
		i.setAllPrivileges(AllPrivileges)
	}

	i.loadMetadata(logger, session)

	return i, nil
}

func createDatabasesFromSchema(config *schema.Schema) map[DatabaseName]*DatabaseInfo {
	dbInfos := make(map[DatabaseName]*DatabaseInfo, len(config.Databases))
	for _, dbSchema := range config.Databases {
		dbInfo := &DatabaseInfo{
			caseSensitiveName: dbSchema.Name,
			Name:              DatabaseName(strings.ToLower(dbSchema.Name)),
			Collections:       make(map[CollectionName]*CollectionInfo),
		}
		dbInfos[dbInfo.Name] = dbInfo
		for _, tblSchema := range dbSchema.Tables {
			name := CollectionName(tblSchema.CollectionName)
			if _, ok := dbInfo.Collections[name]; ok {
				// Because multiple tables can be mapped to the same collection,
				// we can skip collections we've already included.
				continue
			}

			dbInfo.Collections[name] = &CollectionInfo{
				Name: name,
			}
		}
	}

	return dbInfos
}

func (i *Info) loadMetadata(logger *log.Logger, s *mgo.Session) error {
	c := s.Clone()
	defer c.Close()
	for _, dbInfo := range i.Databases {
		err := dbInfo.loadMetadata(logger, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dbInfo *DatabaseInfo) loadMetadata(logger *log.Logger, s *mgo.Session) error {

	if (dbInfo.Privileges & ListCollectionsPrivilege) == 0 {

		logger.Warnf(log.Always, "user does not have the 'listCollections' privileges on database '%v'", dbInfo.caseSensitiveName)

		// User can't list collections on this database. This means
		// we can't determine if the collection is a view or what
		// the collation is. Hence, we need to mark the database
		// and all the collections underneath as having no
		// privileges.
		dbInfo.Privileges = NoPrivileges
		for _, colInfo := range dbInfo.Collections {
			colInfo.Privileges = NoPrivileges
		}
		return nil
	}

	var result struct {
		Collections []bson.Raw
		Cursor      struct {
			FirstBatch []bson.Raw "firstBatch"
			NextBatch  []bson.Raw "nextBatch"
			NS         string
			ID         int64
		}
	}

	logger.Logf(log.DebugHigh, "running listCollections on database '%v'", dbInfo.caseSensitiveName)
	err := s.DB(dbInfo.caseSensitiveName).Run(bson.D{{"listCollections", 1}, {"cursor", struct{}{}}}, &result)
	if err != nil {
		return fmt.Errorf("failed to run listCollections on database '%v': %v", dbInfo.caseSensitiveName, err)
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
		colInfo, ok := dbInfo.Collections[CollectionName(colResult.Name)]
		if !ok {
			continue
		}

		colInfo.Collation = colResult.Options.Collation
		colInfo.IsView = colResult.Type == "view"
	}

	if err := iter.Close(); err != nil {
		return err
	}

	return iter.Err()
}
