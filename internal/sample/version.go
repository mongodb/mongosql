package sample

import (
	"fmt"
	"strings"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
)

type schemaProtocol string

const (
	CurrentProtocol schemaProtocol = "v1"
)

// Version represents a sampled schema version.
type Version struct {
	Id bson.ObjectId `bson:"_id"`

	// an ISODate representing the time at which sampling commenced
	// for this namespace.
	StartSampleTime time.Time `bson:"startSampleTime,omitempty"`

	// an ISODate representing the time at which sampling completed.
	EndSampleTime time.Time `bson:"endSampleTime,omitempty"`

	// an array of namespaces sampled, and therefore, the number of
	// documents in the schemas collection that are expected to
	// reference this version.
	Databases VersionDatabases `bson:"databases,omitempty"`

	// a unique int64 representing the number of times the
	// associated namespaces have been sampled.
	Generation int64 `bson:"generation"`

	// an identifier that indicates the version of the schema format.
	Protocol schemaProtocol `bson:"protocol,omitempty"`

	// a string that uniquely identifies the mongosqld process that
	// performed sampling for this version.
	ProcessName string `bson:"processName,omitempty"`
}

func NewVersion(processName string) *Version {
	return &Version{
		Id:          bson.NewObjectId(),
		Protocol:    CurrentProtocol,
		ProcessName: processName,
	}
}

func (v *Version) FindDatabase(dbName string) (int, bool) {
	for i, db := range v.Databases {
		if db.Name == dbName {
			return i, true
		}
	}
	return 0, false
}

func (v *Version) AddNamespace(dbName, collectionName string) {
	if i, ok := v.FindDatabase(dbName); ok {
		v.Databases[i].Collections = append(
			v.Databases[i].Collections,
			collectionName,
		)
	} else {
		newDatabase := VersionDatabase{
			Name:        dbName,
			Collections: []string{collectionName},
		}
		v.Databases = append(v.Databases, newDatabase)
	}
}

// VersionDatabase represents a sampled schema version database.
type VersionDatabase struct {
	// the name of the database
	Name string `bson:"name"`
	// all collections sampled in this database
	Collections []string `bson:"collections"`
}

func (d *VersionDatabase) String() string {
	var ns []string
	for _, c := range d.Collections {
		ns = append(ns, fmt.Sprintf("%q", c))
	}
	return fmt.Sprintf("%q (%v): [%s]",
		d.Name, len(d.Collections), strings.Join(ns, ", "))
}

type VersionDatabases []VersionDatabase

func (dbs VersionDatabases) String() string {
	var ns []string
	for _, db := range dbs {
		ns = append(ns, db.String())
	}
	return strings.Join(ns, "; ")
}
