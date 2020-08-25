package mongodb

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/description"
)

// Collections used to perform sampling operations.
const (
	LockCollection         = "mongosqld.lock"
	SchemasCollection      = "mongosqld.schemas"
	VersionsCollection     = "mongosqld.versions"
	VersionIDField         = "versionId"
	VersionGenerationField = "generation"
)

// Wrapper for description.Sharded for use in other sqlproxy packages.
const (
	Sharded = description.Sharded
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

// CommandRunner is an interface for running any mongodb commands
// required to get info for the Info type.
type CommandRunner interface {
	// ListCollections returns a cursor to iterate through the collections
	// present on the db database with options opts.
	ListCollections(ctx context.Context, db string, opts driver.CursorOptions) (Cursor, error)

	// ListIndexes returns a cursor to iterate through the indexes on the
	// col collection within the db database.
	ListIndexes(ctx context.Context, db, col string) (Cursor, error)

	// Run executes an arbitrary command against the given database.
	Run(ctx context.Context, db string, cmd bson.D, result interface{}) error

	// TopologyKind returns the TopologyKind for the CommandRunner.
	TopologyKind() description.TopologyKind
}

// SetCompatibleVersion sets the compatible version and compatible version array.
func (i *Info) SetCompatibleVersion(compatibleVersion string) error {
	var array []uint8
	if compatibleVersion != "" {
		var err error
		array, err = procutil.VersionToSlice(compatibleVersion)
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
		return procutil.VersionAtLeast(i.CompatibleVersionArray, version)
	}
	return procutil.VersionAtLeast(i.VersionArray, version)
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
	CaseSensitiveName string

	// Name is the name of the database.
	Name DatabaseName
	// Privileges is a union of all the collection privileges.
	Privileges Privilege
	// Collections is information about collections by name.
	Collections map[CollectionName]*CollectionInfo
}

// VersionInfo contains server version info.
type VersionInfo struct {
	Version      string  `bson:"version"`
	GitVersion   string  `bson:"gitVersion"`
	VersionArray []uint8 `bson:"-"`
}

// LoadMetadata loads metadata - such as indexes and collection information - into the Info struct.
func (i *Info) LoadMetadata(ctx context.Context, logger log.Logger, cr CommandRunner) {
	for _, dbInfo := range i.Databases {
		err := dbInfo.loadMetadata(ctx, logger, cr)
		if err != nil {
			logger.Warnf(
				log.Admin,
				"error while loading metadata for database %q: %v",
				dbInfo.Name, err,
			)
		}
		dbInfo.loadIndexes(ctx, logger, cr)
	}
}

func (dbInfo *DatabaseInfo) loadMetadata(ctx context.Context, logger log.Logger, cr CommandRunner) error {
	logger.Debugf(log.Dev, "running listCollections on database '%v'", dbInfo.CaseSensitiveName)
	cursor, err := cr.ListCollections(ctx, dbInfo.CaseSensitiveName, driver.CursorOptions{})
	if err != nil {
		return fmt.Errorf("failed to run listCollections on database '%v': %v", dbInfo.CaseSensitiveName, err)
	}

	type colResult struct {
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

	result := colResult{}
	for cursor.Next(ctx, &result) {
		colInfo, ok := dbInfo.Collections[CollectionName(result.Name)]
		if !ok {
			continue
		}

		colInfo.Collation = result.Options.Collation
		colInfo.IsView = result.Type == "view"
		if colInfo.IsView {
			viewToUnderlyingCollections[result.Name] = result.Options.ViewOn
		}

		result = colResult{}
	}

	// Check if the server is a mongos
	if cr.TopologyKind() == description.Sharded {
		dbInfo.loadShardingInfo(ctx, logger, cr, viewToUnderlyingCollections)
	}

	if err := cursor.Close(ctx); err != nil {
		return err
	}

	return cursor.Err()
}

func (dbInfo *DatabaseInfo) loadIndexes(ctx context.Context, lg log.Logger, cr CommandRunner) {
	for _, colInfo := range dbInfo.Collections {
		dbName := dbInfo.CaseSensitiveName
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
		cursor, err := cr.ListIndexes(ctx, dbName, colName)
		if err != nil {
			lg.Warnf(log.Admin, "failed to run listIndexes on namespace %q.%q: %v",
				dbName, colName, err)
			continue
		}

		for cursor.Next(ctx, &collectionIndex) {
			collectionIndexes = append(collectionIndexes, collectionIndex)
			collectionIndex.Key = bsonutil.NewD()
		}

		if cursor.Err() != nil {
			lg.Warnf(log.Admin, "cursor unable to iterate through indexes on namespace %q.%q: %v",
				dbName, colName, err)
		}

		colInfo.Indexes = collectionIndexes
	}
}

// loadShardingInfo loads sharding information for the dbInfo map.
func (dbInfo *DatabaseInfo) loadShardingInfo(ctx context.Context, logger log.Logger, cr CommandRunner,
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
			collStatsCommand := bsonutil.NewD(
				bsonutil.NewDocElem("collStats", collectionName),
			)

			err := cr.Run(ctx, string(dbInfo.Name), collStatsCommand, &stats)
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
