package mongodb

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Privilege is a bitwise enumeration of privileges.
type Privilege int

const (
	// NoPrivileges is the absense of all privileges.
	NoPrivileges = 0

	// AllPrivileges is the union of all the privileges.
	AllPrivileges = FindPrivilege
)

const (
	// FindPrivilege is the privilege to query a database.
	FindPrivilege = 1 << iota
)

var privilegeMap = map[string]Privilege{
	"find": FindPrivilege,
}

// IsAnyAllowed indicates if any privileges exist.
func (i *Info) IsAnyAllowed() bool {
	return i.Privileges != NoPrivileges
}

// IsAllowed indicates whether the privilege is granted on the Info.
func (i *Info) IsAllowed(p Privilege) bool {
	return (i.Privileges & p) == p
}

// IsAnyAllowedDatabase indicates if any privileges exist.
func (i *Info) IsAnyAllowedDatabase(name DatabaseName) bool {
	if dbInfo, ok := i.Databases[name]; ok {
		return dbInfo.Privileges != NoPrivileges
	}

	return false
}

// IsAllowedDatabase indicates whether any privileges are granted on the specified database.
func (i *Info) IsAllowedDatabase(name DatabaseName, p Privilege) bool {
	if dbInfo, ok := i.Databases[name]; ok {
		return (dbInfo.Privileges & p) == p
	}

	return false
}

// IsAnyAllowedCollection indicates whether any privileges are granted on the specified collection.
func (i *Info) IsAnyAllowedCollection(dbName DatabaseName, colName CollectionName) bool {
	if dbInfo, ok := i.Databases[dbName]; ok {
		if colInfo, ok := dbInfo.Collections[colName]; ok {
			return colInfo.Privileges != NoPrivileges
		}
	}

	return false
}

// IsAllowedCollection indicates whether the privilege is granted on the specified collection.
func (i *Info) IsAllowedCollection(dbName DatabaseName, colName CollectionName, p Privilege) bool {
	if dbInfo, ok := i.Databases[dbName]; ok {
		if colInfo, ok := dbInfo.Collections[colName]; ok {
			return (colInfo.Privileges & p) == p
		}
	}

	return false
}

// loadAuthInfo gathers the authorization information from MongoDB and propogates
// it to the Info tree.
func (i *Info) loadAuthInfo(s *mgo.Session) error {
	c := s.Clone()
	defer s.Close()
	cmd := bson.D{
		{"connectionStatus", 1},
		{"showPrivileges", 1},
	}
	var result connectionStatusResult
	err := c.Run(cmd, &result)
	if err != nil {
		return err
	}

	i.loadAuthInfoFromConnectionStatus(&result)
	return nil
}

func (i *Info) loadAuthInfoFromConnectionStatus(result *connectionStatusResult) {
	if len(result.AuthInfo.AuthenticatedUsers) == 0 {
		i.setAllPrivileges(AllPrivileges)
	}

	var container privilegeContainer

	addPrivilege := func(dbName DatabaseName, colName CollectionName, privileges Privilege) {

		if dbName == "" {
			container.defaultPrivileges |= privileges
			return
		}

		if container.databases == nil {
			container.databases = make(map[DatabaseName]*dbPrivilegeContainer)
		}

		d, ok := container.databases[dbName]
		if !ok {
			d = &dbPrivilegeContainer{}
			container.databases[dbName] = d
		}

		if colName == "" {
			d.defaultPrivileges |= privileges
		} else {
			if d.collections == nil {
				d.collections = make(map[CollectionName]Privilege)
			}
			d.collections[colName] |= privileges
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

			addPrivilege(DatabaseName(priv.Resource.DB), CollectionName(priv.Resource.Collection), privileges)
		}
	}

	for _, dbInfo := range i.Databases {
		for _, colInfo := range dbInfo.Collections {
			colInfo.Privileges = container.getPrivileges(dbInfo.Name, colInfo.Name)
			dbInfo.Privileges |= colInfo.Privileges
		}
		i.Privileges |= dbInfo.Privileges
	}
}

func (i *Info) setAllPrivileges(privileges Privilege) {
	i.Privileges = privileges
	for _, db := range i.Databases {
		db.Privileges = privileges
		for _, col := range db.Collections {
			col.Privileges = privileges
		}
	}
}

type connectionStatusResult struct {
	AuthInfo struct {
		AuthenticatedUsers          []struct{} "authenticatedUsers"
		AuthenticatedUserPrivileges []struct {
			Resource *struct {
				DB         string "db"
				Collection string
			}
			Actions []string
		} "authenticatedUserPrivileges"
	} "authInfo"
}

type privilegeContainer struct {
	defaultPrivileges Privilege
	databases         map[DatabaseName]*dbPrivilegeContainer
}

func (pc *privilegeContainer) getPrivileges(dbName DatabaseName, colName CollectionName) Privilege {
	if pc.databases == nil {
		return pc.defaultPrivileges
	}

	db, ok := pc.databases[dbName]
	if !ok {
		return pc.defaultPrivileges
	}

	col, ok := db.collections[colName]
	if !ok {
		return db.defaultPrivileges
	}

	return col
}

type dbPrivilegeContainer struct {
	defaultPrivileges Privilege
	collections       map[CollectionName]Privilege
}
