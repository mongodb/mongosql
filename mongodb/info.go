package mongodb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/mongo-go-driver/yamgo/model"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
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
	// Version is the version of the mongodb server.
	Version string
	// VersionArray are the components of the version.
	VersionArray []uint8
	// CompatibleVersion is the version of the mongodb server we will pretend we are talking to.
	CompatibleVersion string
	// CompatibleVersionArray are the components of the compatible version.
	CompatibleVersionArray []uint8
}

// VersionAtLeast indicates whether the MongoDB version is at least the required version.
func (i *Info) VersionAtLeast(version ...uint8) bool {
	if len(i.CompatibleVersionArray) > 0 {
		return util.VersionAtLeast(i.CompatibleVersionArray, version)
	}
	return util.VersionAtLeast(i.VersionArray, version)
}

// SetCompatibleVersion sets the compatible version and compatible version array.
func (i *Info) SetCompatibleVersion(compatibleVersion string) error {
	var array []uint8
	if compatibleVersion != "" {
		parts := strings.Split(compatibleVersion, ".")
		for _, p := range parts {
			i, err := strconv.Atoi(p)
			if err != nil {
				return fmt.Errorf("expected an integer: %v", err)
			}
			array = append(array, uint8(i))
		}
	}

	i.CompatibleVersion = compatibleVersion
	i.CompatibleVersionArray = array
	return nil
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
	Collation *Collation
	// Indexes hold the indexes of the MongoDB collection.
	Indexes []Index
	// IsView is true if the collection is a MongoDB view
	// and false otherwise.
	IsView bool
	// IsSharded is true if the collection is sharded
	// and false otherwise.
	IsSharded bool
}

// addShardingInfo loads sharding information for the dbInfo map.
func addShardingInfo(logger *log.Logger, session *Session, dbs map[DatabaseName]*DatabaseInfo) {
	stats := struct {
		Sharded bool `bson:"sharded"`
	}{}

	for db, dbInfo := range dbs {
		for collection, collectionInfo := range dbInfo.Collections {
			collectionName := string(collection)
			collStatsCommand := struct {
				CollStats string `bson:"collStats"`
			}{collectionName}
			err := session.Run(string(db), collStatsCommand, &stats)
			if err != nil {
				logger.Warnf(log.Admin, "Unable to run collStats on collection %s: %v", collectionName, err)
			} else {
				collectionInfo.IsSharded = stats.Sharded
			}
		}
	}
}

// LoadInfo looks up information from MongoDB.
func LoadInfo(logger *log.Logger, session *Session, config *schema.Schema, requireAuth bool) (*Info, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warnf(log.Admin, "MongoDB information access session possibly closed: %v", r)
		}
	}()

	version := session.Model().Version

	dbs := createDatabasesFromSchema(logger, session, config)

	if session.Model().Kind == model.Mongos {
		addShardingInfo(logger, session, dbs)
	}

	i := &Info{
		Databases:    dbs,
		GitVersion:   session.Model().GitVersion,
		Version:      version.Desc,
		VersionArray: version.Parts,
	}

	if requireAuth {
		err := i.loadAuthInfo(logger, session)
		if err != nil {
			return nil, err
		}
	} else {
		i.setAllPrivileges(AllPrivileges)
	}

	i.loadMetadata(logger, session)

	return i, nil
}

func createDatabasesFromSchema(logger *log.Logger, session *Session, config *schema.Schema) map[DatabaseName]*DatabaseInfo {
	dbInfos := make(map[DatabaseName]*DatabaseInfo, len(config.Databases))
	for _, dbSchema := range config.Databases {
		dbName := strings.ToLower(dbSchema.Name)
		dbInfo := &DatabaseInfo{
			caseSensitiveName: dbSchema.Name,
			Name:              DatabaseName(dbName),
			Collections:       make(map[CollectionName]*CollectionInfo),
		}

		dbInfos[dbInfo.Name] = dbInfo

		for _, table := range dbSchema.Tables {
			name := CollectionName(table.CollectionName)
			if _, ok := dbInfo.Collections[name]; ok {
				// Because multiple tables can be mapped to the same collection,
				// we can skip collections we've already included.
				continue
			}

			collectionIndexes, collectionIndex := []Index{}, Index{}
			cursor, err := session.ListIndexes(dbName, table.CollectionName)
			if err != nil {
				logger.Warnf(log.Admin, "failed to run listIndexes on namespace %q.%q: %v",
					dbName, table.CollectionName, err)
				continue
			}

			for cursor.Next(session.Context(), &collectionIndex) {
				collectionIndexes = append(collectionIndexes, collectionIndex)
			}

			if cursor.Err() != nil {
				logger.Warnf(log.Admin, "cursor unable to iterate through indexes on namespace %q.%q: %v",
					dbName, table.CollectionName, err)
			}

			dbInfo.Collections[name] = &CollectionInfo{
				Name:    name,
				Indexes: collectionIndexes,
			}
		}
	}
	return dbInfos
}

func (i *Info) loadMetadata(logger *log.Logger, s *Session) error {
	for _, dbInfo := range i.Databases {
		err := dbInfo.loadMetadata(logger, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dbInfo *DatabaseInfo) loadMetadata(logger *log.Logger, s *Session) error {
	if (dbInfo.Privileges & ListCollectionsPrivilege) == 0 {

		logger.Warnf(log.Admin, "user does not have the 'listCollections' privileges on database '%v'", dbInfo.caseSensitiveName)

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

	logger.Debugf(log.Dev, "running listCollections on database '%v'", dbInfo.caseSensitiveName)
	iter, err := s.ListCollections(dbInfo.caseSensitiveName)
	if err != nil {
		return fmt.Errorf("failed to run listCollections on database '%v': %v", dbInfo.caseSensitiveName, err)
	}

	var colResult struct {
		Name    string
		Type    string
		Options struct {
			Collation *Collation
		}
	}

	for iter.Next(s.ctx, &colResult) {
		colInfo, ok := dbInfo.Collections[CollectionName(colResult.Name)]
		if !ok {
			continue
		}

		colInfo.Collation = colResult.Options.Collation
		colInfo.IsView = colResult.Type == "view"
	}

	if err := iter.Close(s.ctx); err != nil {
		return err
	}

	return iter.Err()
}
