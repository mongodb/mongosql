package mongodb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/10gen/mongo-go-driver/mongo/model"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
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

// LoadInfo looks up information from MongoDB.
func LoadInfo(logger *log.Logger, sp *SessionProvider, userSession *Session, config *schema.Schema, requireAuth bool) (*Info, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warnf(log.Admin, "MongoDB information access session possibly closed: %v", r)
		}
	}()

	adminSession, err := sp.AdminSession(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create admin session for loading metadata: %v", err)
	}

	// Because the driver does not directly provide the server version, check out a connection
	// from the pool to get its version information.
	c, err := userSession.Connection(context.Background())
	if err != nil {
		return nil, err
	}
	s := c.Model().Server
	if err := c.Close(); err != nil {
		return nil, err
	}

	dbs := createDatabasesFromSchema(logger, adminSession, config)

	i := &Info{
		Databases:    dbs,
		GitVersion:   s.GitVersion,
		Version:      s.Version.Desc,
		VersionArray: s.Version.Parts,
	}

	if requireAuth {
		err := i.loadAuthInfo(logger, userSession)
		if err != nil {
			return nil, err
		}
	} else {
		i.setAllPrivileges(AllPrivileges)
	}

	i.loadMetadata(logger, adminSession)

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

			dbInfo.Collections[name] = &CollectionInfo{
				Name: name,
			}
		}
	}
	return dbInfos
}

func (i *Info) loadMetadata(logger *log.Logger, s *Session) {
	for _, dbInfo := range i.Databases {
		err := dbInfo.loadMetadata(logger, s)
		if err != nil {
			logger.Warnf(log.Admin, "error while loading metadata for database %q: %v", dbInfo.Name, err)
		}
		dbInfo.loadIndexes(logger, s)
	}
}

func (dbInfo *DatabaseInfo) loadMetadata(logger *log.Logger, s *Session) error {
	logger.Debugf(log.Dev, "running listCollections on database '%v'", dbInfo.caseSensitiveName)
	iter, err := s.ListCollections(dbInfo.caseSensitiveName, ops.ListCollectionsOptions{})
	if err != nil {
		return fmt.Errorf(
			"failed to run listCollections on database '%v': %v",
			dbInfo.caseSensitiveName, err,
		)
	}

	var colResult struct {
		Name    string `bson:"name"`
		Type    string `bson:"type"`
		Options struct {
			Collation *Collation `bson:"collation"`
			ViewOn    string     `bson:"viewOn"`
		} `bson:"options"`
	}

	// This caches views and the views/collections they're based on so that it can be easy to determine
	// whether a view is sharded in loadShardingInfo.
	viewToUnderlyingCollections := make(map[string]string)

	for iter.Next(s.ctx, &colResult) {
		colInfo, ok := dbInfo.Collections[CollectionName(colResult.Name)]
		if !ok {
			continue
		}

		colInfo.Collation = colResult.Options.Collation
		colInfo.IsView = colResult.Type == "view"
		if colInfo.IsView {
			viewToUnderlyingCollections[colResult.Name] = colResult.Options.ViewOn
		}
	}

	if s.Model().Kind == model.Mongos {
		dbInfo.loadShardingInfo(logger, s, viewToUnderlyingCollections)
	}

	if err := iter.Close(s.ctx); err != nil {
		return err
	}

	return iter.Err()
}

func (dbInfo *DatabaseInfo) loadIndexes(lg *log.Logger, s *Session) {
	for _, colInfo := range dbInfo.Collections {
		dbName := string(dbInfo.Name)
		colName := string(colInfo.Name)

		if colInfo.IsView {
			lg.Infof(log.Admin, "not loading indexes for %q.%q: collection is a view", dbName, colName)
			continue
		}

		collectionIndexes, collectionIndex := []Index{}, Index{}
		cursor, err := s.ListIndexes(dbName, colName)
		if err != nil {
			lg.Warnf(log.Admin, "failed to run listIndexes on namespace %q.%q: %v",
				dbName, colName, err)
			continue
		}

		for cursor.Next(s.Context(), &collectionIndex) {
			collectionIndexes = append(collectionIndexes, collectionIndex)
		}

		if cursor.Err() != nil {
			lg.Warnf(log.Admin, "cursor unable to iterate through indexes on namespace %q.%q: %v",
				dbName, colName, err)
		}

		colInfo.Indexes = collectionIndexes
	}
}

// loadShardingInfo loads sharding information for the dbInfo map.
func (dbInfo *DatabaseInfo) loadShardingInfo(logger *log.Logger, session *Session, viewToUnderlyingCollection map[string]string) {
	stats := struct {
		Sharded bool `bson:"sharded"`
	}{}

	// caching sharding results to reduce multiple round trips for same collection.
	isShardedCollection := make(map[string]bool)
	for collection, collectionInfo := range dbInfo.Collections {
		collectionName := string(collection)

		// CollStats fails when run against a view. In order to get sharding information on a view, we need to
		// get the underlying collection and then run a collStats on that collection. Since views can be built on top of views
		// we traverse views until we hit a base collection.
		if collectionInfo.IsView {
			var next string
			baseCollection, ok := viewToUnderlyingCollection[collectionName]
			for ok {
				if next, ok = viewToUnderlyingCollection[baseCollection]; ok {
					baseCollection = next
				}
			}
			viewToUnderlyingCollection[collectionName] = baseCollection
			collectionName = baseCollection
		}

		if isSharded, ok := isShardedCollection[collectionName]; !ok {
			collStatsCommand := struct {
				CollStats string `bson:"collStats"`
			}{collectionName}

			err := session.Run(string(dbInfo.Name), collStatsCommand, &stats)
			if err != nil {
				logger.Warnf(log.Admin, "unable to run collStats on collection %s, %v", collectionName, err)
			} else {
				isShardedCollection[collectionName] = stats.Sharded
				collectionInfo.IsSharded = stats.Sharded
			}
		} else {
			collectionInfo.IsSharded = isSharded
		}

	}
}
