package mongodb

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/log"

	"github.com/10gen/mongo-go-driver/bson"
)

// Privilege is a bitwise enumeration of privileges.
type Privilege int

const (
	// NoPrivileges is the absence of all privileges.
	NoPrivileges Privilege = 0

	// VisibilityPrivileges are the privileges necessary to view the schema for
	// a Database or Collection.
	VisibilityPrivileges = FindPrivilege | ListCollectionsPrivilege

	// AllPrivileges is the union of all the privileges.
	AllPrivileges = CreateIndexPrivilege |
		FindPrivilege |
		InprogPrivilege |
		InsertPrivilege |
		KillopPrivilege |
		ListCollectionsPrivilege |
		UpdatePrivilege
)

const (
	// CreateIndexPrivilege allows creation of a index on a collection, applied
	// to a database or collection.
	CreateIndexPrivilege Privilege = 1 << iota
	// FindPrivilege allows querying a database, applied to either a
	// database or collection.
	FindPrivilege
	// InprogPrivilege allows for viewing all processes, without this a user
	// can only view their own processes.
	InprogPrivilege
	// InsertPrivilege allows inserting into a collection, applied to a
	// database or collection.
	InsertPrivilege
	// KillopPrivilege allows killing cursors, applied to the cluster
	// resource. Treat as applied to all databases.
	KillopPrivilege
	// ListCollectionsPrivilege allows listing the collections of a
	// database.
	ListCollectionsPrivilege
	// UpdatePrivilege allows updating collections, applied to databases or
	// collections.
	UpdatePrivilege
)

var sampleCollections = map[CollectionName]struct{}{
	CollectionName("mongosqld.lock"):     {},
	CollectionName("mongosqld.schemas"):  {},
	CollectionName("mongosqld.versions"): {},
}

var privilegeMap = map[string]Privilege{
	"createIndex":     CreateIndexPrivilege,
	"find":            FindPrivilege,
	"inprog":          InprogPrivilege,
	"insert":          InsertPrivilege,
	"killop":          KillopPrivilege,
	"listCollections": ListCollectionsPrivilege,
	"update":          UpdatePrivilege,
}

// IsAllowedCluster returns true if a set of privileges is allowed for the cluster.
func (i *Info) IsAllowedCluster(p Privilege) bool {
	return (i.ClusterPrivileges & p) == p
}

// IsAllowedCollection indicates whether the privilege is granted on the specified collection.
func (i *Info) IsAllowedCollection(dbName DatabaseName, colName CollectionName, p Privilege) bool {
	if dbInfo, ok := i.Databases[DatabaseName(strings.ToLower(string(dbName)))]; ok {
		if colInfo, ok := dbInfo.Collections[colName]; ok {
			return (colInfo.Privileges & p) == p
		}
	}

	return false
}

// IsAllowedDatabase indicates whether any privileges are granted on the
// specified database.
func (i *Info) IsAllowedDatabase(name DatabaseName, p Privilege) bool {
	if dbInfo, ok := i.Databases[DatabaseName(strings.ToLower(string(name)))]; ok {
		return (dbInfo.Privileges & p) == p
	}

	return false
}

// IsAllowedSampleSource returns true if the user has the given privileges
// for the sample collections.
func (i *Info) IsAllowedSampleSource(privileges Privilege) bool {
	if i.sampleSourcePrivileges.databasePrivileges&privileges == privileges {
		return true
	}
	for collName := range sampleCollections {
		if (i.sampleSourcePrivileges.collections[collName] &
			privileges) != privileges {
			return false
		}
	}
	return true
}

// IsVisibleCollection returns true if a Collection's schema should be
// visible based on privileges, and false otherwise. That means that either the
// Database has ListCollections privilege, or the collection itself has Find
// privilege.
func (i *Info) IsVisibleCollection(dbName DatabaseName, colName CollectionName) bool {
	return (i.IsAllowedDatabase(dbName, ListCollectionsPrivilege) ||
		i.IsAllowedCollection(dbName, colName, FindPrivilege))
}

// IsVisibleDatabase returns true if a Database's schema should be visible
// based on privileges, and false otherwise. That means that the user has find or
// listCollections on privileges to the database. Note that due to how we assign
// privileges, a database will have find permissions assigned if an underlying
// collection has find privileges. This does not allow the user to access
// all collections, but rather just allows the catalog to display the database.
func (i *Info) IsVisibleDatabase(name DatabaseName) bool {
	if dbInfo, ok := i.Databases[DatabaseName(strings.ToLower(string(name)))]; ok {
		return (dbInfo.Privileges & VisibilityPrivileges) != NoPrivileges
	}
	return false
}

// IsSecurityEnabled returns true if security is enabled.
func (i *Info) IsSecurityEnabled() bool {
	return (i.Config != nil) && i.Config.Security.Enabled
}

// loadAuthInfo gathers the authorization information from MongoDB and propagates
// it to the Info tree.
func (i *Info) loadAuthInfo(logger *log.Logger, s *Session, sampleSource string) error {
	cmd := bson.D{
		{Name: "connectionStatus", Value: 1},
		{Name: "showPrivileges", Value: 1},
	}
	var result connectionStatusResult
	logger.Infof(log.Dev, "loading privilege information for current user")
	if err := s.Run("admin", cmd, &result); err != nil {
		return fmt.Errorf("failed to load privilege information for the current user: %v", err)
	}
	i.loadAuthInfoFromConnectionStatus(&result, sampleSource)
	return nil
}

func (i *Info) loadAuthInfoFromConnectionStatus(result *connectionStatusResult,
	sampleSource string) {
	var normalizedSampleSource = DatabaseName(strings.ToLower(sampleSource))
	if len(result.AuthInfo.AuthenticatedUsers) == 0 {
		i.setAllPrivileges(AllPrivileges)
	}

	i.sampleSourcePrivileges = newSampleSourcePrivileges()

	addPrivilegesToAllDatabases := func(colName CollectionName, privileges Privilege) {
		// if colName == "", this corresponds to a cluster wide privilege.
		// We must set this permission on every database as well as the
		// sampleSource because it exists outside of i.Databases.
		if colName == "" {
			for _, dbInfo := range i.Databases {
				dbInfo.Privileges |= privileges
				// Update the permissions for all collections in all databases.
				for _, colInfo := range dbInfo.Collections {
					colInfo.Privileges |= privileges
				}
			}
			i.sampleSourcePrivileges.databasePrivileges |= privileges
			i.ClusterPrivileges |= privileges
		} else {
			// this is a privilege across all collections in any database with
			// this name. We must set the privileges for this collection
			// across all known databases, as well as the sampleSource,
			// since it exists outside of i.Databases.
			for _, dbInfo := range i.Databases {
				// If this Database does not have this collection, skip setting
				// permissions for this Database.
				if colInfo, ok := dbInfo.Collections[colName]; ok {
					colInfo.Privileges |= privileges
					// We will give find permissions to any database for which
					// collections have privileges. This is a slight mismatch
					// with mongod semantics, as find on the database would
					// normally apply to all collections in the database. We do
					// this because a user should be able to see any database
					// containing collections (tables) to which they have
					// privileges.
					dbInfo.Privileges |= FindPrivilege
				}
			}
			i.sampleSourcePrivileges.collections[colName] |= privileges
		}
	}

	addPrivilegesToSampleSource := func(colName CollectionName, privileges Privilege) {
		if colName == "" {
			i.sampleSourcePrivileges.databasePrivileges |= privileges
		} else {
			i.sampleSourcePrivileges.collections[colName] |= privileges
		}
	}

	addPrivilegesToOneDatabase := func(dbInfo *DatabaseInfo,
		colName CollectionName,
		privileges Privilege) {
		if colName == "" {
			// update the permissions for all collections in the database.
			for _, colInfo := range dbInfo.Collections {
				colInfo.Privileges |= privileges
			}
			dbInfo.Privileges |= privileges
		} else {
			// Only set privileges if the collection exists in the dbInfo.
			if colInfo, ok := dbInfo.Collections[colName]; ok {
				colInfo.Privileges |= privileges
				// We will give find permissions to any database for which
				// collections have privileges. This is a slight mismatch with
				// mongod semantics, as find on the database would normally
				// apply to all collections in the database. We do this because
				// a user should be able to see any database containing
				// collections (tables) to which they have privileges.
				dbInfo.Privileges |= FindPrivilege
			}
		}

	}

	addPrivileges := func(dbName DatabaseName, colName CollectionName, privileges Privilege) {
		switch dbName {
		// If dbName is "", we apply these privileges across all databases.
		case "":
			addPrivilegesToAllDatabases(colName, privileges)
		// If dbName is sampleSource database, we apply these privileges to
		// the sampleSource, which exists outside of i.Databases.
		case normalizedSampleSource:
			addPrivilegesToSampleSource(colName, privileges)
		// Otherwise, just apply the privileges to the proper DatabaseInfo.
		default:
			if dbInfo, ok := i.Databases[dbName]; ok {
				addPrivilegesToOneDatabase(dbInfo, colName, privileges)
			}
		}
	}

	for _, priv := range result.AuthInfo.AuthenticatedUserPrivileges {
		if priv.Resource != nil {
			var privileges Privilege
			for _, a := range priv.Actions {
				if p, ok := privilegeMap[a]; ok {
					privileges |= p
				}
			}

			addPrivileges(
				DatabaseName(strings.ToLower(priv.Resource.DB)),
				CollectionName(priv.Resource.Collection),
				privileges,
			)
		}
	}
}

// nolint: unparam
func (i *Info) setAllPrivileges(privileges Privilege) {
	for _, db := range i.Databases {
		db.Privileges = privileges
		for _, col := range db.Collections {
			col.Privileges = privileges
		}
	}
}

type connectionStatusResult struct {
	AuthInfo struct {
		AuthenticatedUsers          []struct{} `bson:"authenticatedUsers"`
		AuthenticatedUserPrivileges []struct {
			Resource *struct {
				DB         string `bson:"db"`
				Collection string
			}
			Actions []string
		} `bson:"authenticatedUserPrivileges"`
	} `bson:"authInfo"`
}

type dbPrivilegeContainer struct {
	databasePrivileges Privilege
	collections        map[CollectionName]Privilege
}

func newSampleSourcePrivileges() dbPrivilegeContainer {
	ret := dbPrivilegeContainer{}
	ret.collections = make(map[CollectionName]Privilege)
	for collName := range sampleCollections {
		ret.collections[collName] = NoPrivileges
	}
	return ret
}
