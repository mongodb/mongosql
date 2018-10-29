package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/10gen/mongo-go-driver/mongo/model"
	"github.com/10gen/mongo-go-driver/mongo/private/ops"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"
)

// Collections used to perform sampling operations.
const (
	LockCollection         = "mongosqld.lock"
	SchemasCollection      = "mongosqld.schemas"
	VersionsCollection     = "mongosqld.versions"
	VersionIDField         = "versionId"
	VersionGenerationField = "generation"
)

// DatabaseName is the name of a database.
type DatabaseName string

// CollectionName is the name of a collection.
type CollectionName string

// Info is the configuration of MongoDB at runtime.
// Info is maintained per connection, and contains information
// specific to the connected user, such as allowed privileges.
type Info struct {
	// Config is the server configuration, which is needed for
	// various authorization actions.
	Config *config.Config
	// CompatibleVersion is the version of the mongodb server we will pretend
	// we are talking to.
	CompatibleVersion string
	// CompatibleVersionArray are the components of the compatible version.
	CompatibleVersionArray []uint8
	// Databases is information about databases by name.
	Databases map[DatabaseName]*DatabaseInfo
	// Git version is the git hash of the mongodb server.
	GitVersion string
	// ClusterPrivileges are all cluster level privileges.
	ClusterPrivileges Privilege
	// Version is the version of the mongodb server.
	Version string
	// VersionArray are the components of the version.
	VersionArray []uint8
	// sampleSourcePrivileges are the privileges the user has on the sample
	// source, needed to authorize flush sample in write mode. We need
	// to keep these aside from Databases because Databases does not
	// contain the SampleSource database.
	sampleSourcePrivileges dbPrivilegeContainer
}

// SetCompatibleVersion sets the compatible version and compatible version array.
func (i *Info) SetCompatibleVersion(compatibleVersion string) error {
	var array []uint8
	if compatibleVersion != "" {
		var err error
		array, err = util.VersionToSlice(compatibleVersion)
		if err != nil {
			return err
		}
	}

	i.CompatibleVersion = compatibleVersion
	i.CompatibleVersionArray = array
	return nil
}

// VersionAtLeast indicates whether the MongoDB version is at least the required version.
func (i *Info) VersionAtLeast(version ...uint8) bool {
	if len(i.CompatibleVersionArray) > 0 {
		return util.VersionAtLeast(i.CompatibleVersionArray, version)
	}
	return util.VersionAtLeast(i.VersionArray, version)
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

// LoadInfo looks up information from MongoDB.
func LoadInfo(ctx context.Context, logger log.Logger, sp *SessionProvider, userSession *Session,
	schema *schema.Schema, config *config.Config) (i *Info, e error) {

	defer func() {
		if r := recover(); r != nil {
			i = nil
			logger.Warnf(log.Admin, "MongoDB information access session possibly closed: %v", r)
			// Make sure we return the error. Without the next line go just returns nil, nil,
			// which breaks the contract we want (namely that if the returned info is nil,
			// there should be an error).
			switch x := r.(type) {
			case string:
				e = errors.New(x)
			case error:
				e = x
			default:
				e = errors.New("Unknown panic")
			}
		}
	}()

	adminSession, err := sp.AuthenticatedAdminSession(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create admin session for loading metadata: %v", err)
	}

	// Because the driver does not directly provide the server version, check
	// out a connection from the pool to get its version information.
	c, err := userSession.Connection(context.Background())
	if err != nil {
		return nil, err
	}
	s := c.Model().Server
	if err = c.Close(); err != nil {
		return nil, err
	}

	dbs := createDatabasesFromSchema(schema)

	i = &Info{
		Databases:    dbs,
		GitVersion:   s.GitVersion,
		Version:      s.Version.Desc,
		VersionArray: s.Version.Parts,
		Config:       config,
	}

	if config.Security.Enabled {
		err = i.loadAuthInfo(ctx, logger, userSession,
			config.Schema.Sample.Source)
		if err != nil {
			return nil, err
		}
	} else {
		i.setAllPrivileges(AllPrivileges)
	}

	i.loadMetadata(ctx, logger, adminSession)

	return i, nil
}

func createDatabasesFromSchema(config *schema.Schema) map[DatabaseName]*DatabaseInfo {
	dbInfos := make(map[DatabaseName]*DatabaseInfo, len(config.Databases()))
	for _, dbSchema := range config.Databases() {
		dbName := strings.ToLower(dbSchema.Name())
		dbInfo := &DatabaseInfo{
			caseSensitiveName: dbSchema.Name(),
			Name:              DatabaseName(dbName),
			Collections:       make(map[CollectionName]*CollectionInfo),
		}

		dbInfos[dbInfo.Name] = dbInfo

		for _, table := range dbSchema.Tables() {
			name := CollectionName(table.MongoName())
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

func (i *Info) loadMetadata(ctx context.Context, logger log.Logger, s *Session) {
	for _, dbInfo := range i.Databases {
		err := dbInfo.loadMetadata(ctx, logger, s)
		if err != nil {
			logger.Warnf(
				log.Admin,
				"error while loading metadata for database %q: %v",
				dbInfo.Name, err,
			)
		}
		dbInfo.loadIndexes(ctx, logger, s)
	}
}

func (dbInfo *DatabaseInfo) loadMetadata(ctx context.Context, logger log.Logger, s *Session) error {
	logger.Debugf(log.Dev, "running listCollections on database '%v'", dbInfo.caseSensitiveName)
	iter, err := s.ListCollections(ctx, dbInfo.caseSensitiveName, ops.ListCollectionsOptions{})
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

	// This caches views and the views/collections they're based on so that it can be easy to
	// determine whether a view is sharded in loadShardingInfo.
	viewToUnderlyingCollections := make(map[string]string)

	for iter.Next(ctx, &colResult) {
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
		dbInfo.loadShardingInfo(ctx, logger, s, viewToUnderlyingCollections)
	}

	if err := iter.Close(ctx); err != nil {
		return err
	}

	return iter.Err()
}

func (dbInfo *DatabaseInfo) loadIndexes(ctx context.Context, lg log.Logger, s *Session) {
	for _, colInfo := range dbInfo.Collections {
		dbName := string(dbInfo.Name)
		colName := string(colInfo.Name)

		if colInfo.IsView {
			lg.Infof(
				log.Admin,
				"not loading indexes for %q.%q: collection is a view",
				dbName, colName,
			)
			continue
		}

		collectionIndexes, collectionIndex := []Index{}, Index{}
		cursor, err := s.ListIndexes(ctx, dbName, colName)
		if err != nil {
			lg.Warnf(log.Admin, "failed to run listIndexes on namespace %q.%q: %v",
				dbName, colName, err)
			continue
		}

		for cursor.Next(ctx, &collectionIndex) {
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
func (dbInfo *DatabaseInfo) loadShardingInfo(ctx context.Context, logger log.Logger, session *Session,
	viewToUnderlyingCollection map[string]string) {

	stats := struct {
		Sharded bool `bson:"sharded"`
	}{}

	// caching sharding results to reduce multiple round trips for same
	// collection.
	isShardedCollection := make(map[string]bool)
	for collection, collectionInfo := range dbInfo.Collections {
		collectionName := string(collection)

		// CollStats fails when run against a view. In order to get sharding
		// information on a view, we need to get the underlying collection and
		// then run a collStats on that collection. Since views can be built
		// on top of views we traverse views until we hit a base collection.
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

			err := session.Run(ctx, string(dbInfo.Name), collStatsCommand, &stats)
			if err != nil {
				logger.Warnf(
					log.Admin,
					"unable to run collStats on collection %s, %v",
					collectionName, err,
				)
			} else {
				isShardedCollection[collectionName] = stats.Sharded
				collectionInfo.IsSharded = stats.Sharded
			}
		} else {
			collectionInfo.IsSharded = isSharded
		}

	}
}
