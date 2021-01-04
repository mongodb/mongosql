package mongodb

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
)

func TestLoadAuthInfoFromConnectionStatus(t *testing.T) {
	req := require.New(t)

	info := &Info{
		Databases: map[DatabaseName]*DatabaseInfo{
			"test1": {
				CaseSensitiveName: "test1",
				Collections: map[CollectionName]*CollectionInfo{
					"a": {Name: "a"},
					"b": {Name: "b"},
				},
			},
			"test2": {
				CaseSensitiveName: "test2",
				Collections: map[CollectionName]*CollectionInfo{
					"c": {Name: "c"},
					"d": {Name: "d"},
				},
			},
		},
	}

	result := bsonutil.NewM(
		bsonutil.NewDocElem("authInfo", bsonutil.NewM(
			bsonutil.NewDocElem("authenticatedUsers", bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem("user", "user1"), bsonutil.NewDocElem("db", "test")),
			)),
			bsonutil.NewDocElem("authenticatedUserPrivileges", bsonutil.NewArray(
				bsonutil.NewM(
					bsonutil.NewDocElem("resource", bsonutil.NewM(bsonutil.NewDocElem("cluster", true))),
					bsonutil.NewDocElem("actions", bsonutil.NewArray(
						"inprog",
						"killop",
					)),
				),
				bsonutil.NewM(
					bsonutil.NewDocElem("resource", bsonutil.NewM(bsonutil.NewDocElem("db", ""), bsonutil.NewDocElem("collection", "mongosqld.lock"))),
					bsonutil.NewDocElem("actions", bsonutil.NewArray(
						"insert",
						"update",
					)),
				),
				bsonutil.NewM(
					bsonutil.NewDocElem("resource", bsonutil.NewM(bsonutil.NewDocElem("db", ""), bsonutil.NewDocElem("collection", "mongosqld.schemas"))),
					bsonutil.NewDocElem("actions", bsonutil.NewArray(
						"insert",
						"update",
					)),
				),
				bsonutil.NewM(
					bsonutil.NewDocElem("resource", bsonutil.NewM(bsonutil.NewDocElem("db", ""), bsonutil.NewDocElem("collection", "mongosqld.versions"))),
					bsonutil.NewDocElem("actions", bsonutil.NewArray(
						"insert",
						"update",
					)),
				),

				bsonutil.NewM(
					// default for TEST1
					bsonutil.NewDocElem("resource", bsonutil.NewM(bsonutil.NewDocElem("db", "TEST1"), bsonutil.NewDocElem("collection", ""))),
					bsonutil.NewDocElem("actions", bsonutil.NewArray(
						"find",
						"update",
						"listCollections",
					)),
				),
				bsonutil.NewM(
					bsonutil.NewDocElem("resource", bsonutil.NewM(bsonutil.NewDocElem("db", "TEST1"), bsonutil.NewDocElem("collection", "b"))),
					bsonutil.NewDocElem("actions", bsonutil.NewArray(
						"insert",
					)),
				),

				bsonutil.NewM(
					// default for test2
					bsonutil.NewDocElem("resource", bsonutil.NewM(bsonutil.NewDocElem("db", "test2"), bsonutil.NewDocElem("collection", ""))),
					bsonutil.NewDocElem("actions", bsonutil.NewArray(
						"createIndex",
					)),
				),
				bsonutil.NewM(
					bsonutil.NewDocElem("resource", bsonutil.NewM(bsonutil.NewDocElem("db", "test2"), bsonutil.NewDocElem("collection", "d"))),
					bsonutil.NewDocElem("actions", bsonutil.NewArray(
						"find",
						"listCollections",
					)),
				),
			)),
		)),
	)

	resultBson, err := bson.Marshal(result)
	req.Nil(err, "failed to marshal bson")

	var connStatusResult connectionStatusResult
	err = bson.Unmarshal(resultBson, &connStatusResult)
	req.Nil(err, "failed to unmarshal bson")

	info.loadAuthInfoFromConnectionStatus(&connStatusResult, "SAM")

	// First CanUpdateSampleSource test is with explicit tables and no database.
	req.True(info.IsAllowedSampleSource(InsertPrivilege|UpdatePrivilege),
		"user should be able to update sample source")

	req.True(len(info.Databases) >= 1,
		"there should be at least one database")

	test1, ok := info.Databases[DatabaseName("test1")]
	req.True(ok, "could not find 'test1' in databases")
	req.Equal(2, len(test1.Collections),
		"'test1' should contain 2 collections")
	req.Equal(FindPrivilege|UpdatePrivilege|
		ListCollectionsPrivilege|
		InprogPrivilege|KillopPrivilege,
		test1.Privileges,
		"should have find update listCollections inprog and killop on 'test1'")

	a, ok := test1.Collections[CollectionName("a")]
	req.True(ok, "could not find collection 'test1.a'")
	req.Equal(FindPrivilege|UpdatePrivilege|
		ListCollectionsPrivilege|
		InprogPrivilege|KillopPrivilege,
		a.Privileges,
		"should have all privleges of test1 on 'test1.a'")

	b, ok := test1.Collections[CollectionName("b")]
	req.True(ok, "could not find collection 'test1.b'")
	req.Equal(FindPrivilege|InsertPrivilege|UpdatePrivilege|
		ListCollectionsPrivilege|InprogPrivilege|
		KillopPrivilege,
		b.Privileges,
		"should have all 'test1' privileges + insert")

	test2, ok := info.Databases[DatabaseName("test2")]
	req.True(ok, "could not find 'test2' in databases")
	req.Equal(len(test2.Collections), 2,
		"'test2' should contain 2 collections")
	req.Equal(FindPrivilege|CreateIndexPrivilege|
		KillopPrivilege|
		InprogPrivilege,
		test2.Privileges,
		"should have find, createIndex, inprog, and killop on 'test2'")

	c, ok := test2.Collections[CollectionName("c")]
	req.True(ok, "could not find collection 'test2.c'")
	req.Equal(CreateIndexPrivilege|KillopPrivilege|
		InprogPrivilege,
		c.Privileges,
		"should have 'test2' privileges on 'test2.c'")

	d, ok := test2.Collections[CollectionName("d")]
	req.True(ok, "could not find collection 'test2.d'")
	req.Equal(CreateIndexPrivilege|KillopPrivilege|
		InprogPrivilege|FindPrivilege|ListCollectionsPrivilege,
		d.Privileges,
		"should have 'test2' privileges on 'test2.d' + "+
			"find and listCollections")

	result = bsonutil.NewM(
		bsonutil.NewDocElem("authInfo", bsonutil.NewM(
			bsonutil.NewDocElem("authenticatedUsers", bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem("user", "user1"), bsonutil.NewDocElem("db", "test")),
			)),
			bsonutil.NewDocElem("authenticatedUserPrivileges", bsonutil.NewArray(
				bsonutil.NewM(
					bsonutil.NewDocElem("resource", bsonutil.NewM(bsonutil.NewDocElem("db", "sam"), bsonutil.NewDocElem("collection", ""))),
					bsonutil.NewDocElem("actions", bsonutil.NewArray(
						"insert",
						"update",
					)),
				),
			)),
		)),
	)

	resultBson, err = bson.Marshal(result)
	req.Nil(err, "failed to marshal bson")

	err = bson.Unmarshal(resultBson, &connStatusResult)
	req.Nil(err, "failed to unmarshall bson")

	// Second CanUpdateSampleSource test is with update and insert on the sample.
	info.loadAuthInfoFromConnectionStatus(&connStatusResult, "SAM")
	req.True(info.IsAllowedSampleSource(InsertPrivilege|UpdatePrivilege),
		"user should be able to update sample source")

	result = bsonutil.NewM(
		bsonutil.NewDocElem("authInfo", bsonutil.NewM(
			bsonutil.NewDocElem("authenticatedUsers", bsonutil.NewArray(
				bsonutil.NewM(bsonutil.NewDocElem("user", "user1"), bsonutil.NewDocElem("db", "test")),
			)),
			bsonutil.NewDocElem("authenticatedUserPrivileges", bsonutil.NewArray(
				bsonutil.NewM(
					bsonutil.NewDocElem("resource", bsonutil.NewM(bsonutil.NewDocElem("db", ""), bsonutil.NewDocElem("collection", ""))),
					bsonutil.NewDocElem("actions", bsonutil.NewArray(
						"insert",
						"update",
					)),
				),
			)),
		)),
	)

	resultBson, err = bson.Marshal(result)
	req.Nil(err, "failed to marshal bson")

	err = bson.Unmarshal(resultBson, &connStatusResult)
	req.Nil(err, "failed to unmarshall bson")

	// Third CanUpdateSampleSource test is with update and insert on all databases.
	info.loadAuthInfoFromConnectionStatus(&connStatusResult, "SAM")
	req.True(info.IsAllowedSampleSource(InsertPrivilege|UpdatePrivilege),
		"user should be able to update sample source")
}
