package mongodb

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/stretchr/testify/require"
)

func TestLoadAuthInfoFromConnectionStatus(t *testing.T) {
	req := require.New(t)

	info := &Info{
		Databases: map[DatabaseName]*DatabaseInfo{
			"test1": {
				Name: "test1",
				Collections: map[CollectionName]*CollectionInfo{
					"a": {Name: "a"},
					"b": {Name: "b"},
				},
			},
			"test2": {
				Name: "test2",
				Collections: map[CollectionName]*CollectionInfo{
					"c": {Name: "c"},
					"d": {Name: "d"},
				},
			},
		},
	}

	result := bson.M{
		"authInfo": bson.M{
			"authenticatedUsers": []interface{}{
				bson.M{"user": "user1", "db": "test"},
			},
			"authenticatedUserPrivileges": []interface{}{
				bson.M{
					"resource": bson.M{"cluster": true},
					"actions":  []interface{}{"inprog", "killop"},
				},
				bson.M{
					"resource": bson.M{"db": "", "collection": "mongosqld.lock"},
					"actions":  []interface{}{"insert", "update"},
				},
				bson.M{
					"resource": bson.M{"db": "", "collection": "mongosqld.schemas"},
					"actions":  []interface{}{"insert", "update"},
				},
				bson.M{
					"resource": bson.M{"db": "", "collection": "mongosqld.versions"},
					"actions":  []interface{}{"insert", "update"},
				},
				bson.M{
					// default for TEST1
					"resource": bson.M{"db": "TEST1", "collection": ""},
					"actions":  []interface{}{"find", "update", "listCollections"},
				},
				bson.M{
					"resource": bson.M{"db": "TEST1", "collection": "b"},
					"actions":  []interface{}{"insert"},
				},
				bson.M{
					// default for test2
					"resource": bson.M{"db": "test2", "collection": ""},
					"actions":  []interface{}{"createIndex"},
				},
				bson.M{
					"resource": bson.M{"db": "test2", "collection": "d"},
					"actions":  []interface{}{"find", "listCollections"},
				},
			},
		},
	}

	resultBson, err := bson.Marshal(result)
	req.Nil(err, "failed to marshal bson")

	var connStatusResult connectionStatusResult
	err = bson.Unmarshal(resultBson, &connStatusResult)
	req.Nil(err, "failed to unmarshall bson")

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

	result = bson.M{
		"authInfo": bson.M{
			"authenticatedUsers": []interface{}{
				bson.M{"user": "user1", "db": "test"},
			},
			"authenticatedUserPrivileges": []interface{}{
				bson.M{
					"resource": bson.M{"db": "sam", "collection": ""},
					"actions":  []interface{}{"insert", "update"},
				},
			},
		},
	}

	resultBson, err = bson.Marshal(result)
	req.Nil(err, "failed to marshal bson")

	err = bson.Unmarshal(resultBson, &connStatusResult)
	req.Nil(err, "failed to unmarshall bson")

	// Second CanUpdateSampleSource test is with update and insert on the sample.
	info.loadAuthInfoFromConnectionStatus(&connStatusResult, "SAM")
	req.True(info.IsAllowedSampleSource(InsertPrivilege|UpdatePrivilege),
		"user should be able to update sample source")

	result = bson.M{
		"authInfo": bson.M{
			"authenticatedUsers": []interface{}{
				bson.M{"user": "user1", "db": "test"},
			},
			"authenticatedUserPrivileges": []interface{}{
				bson.M{
					"resource": bson.M{"db": "", "collection": ""},
					"actions":  []interface{}{"insert", "update"},
				},
			},
		},
	}

	resultBson, err = bson.Marshal(result)
	req.Nil(err, "failed to marshal bson")

	err = bson.Unmarshal(resultBson, &connStatusResult)
	req.Nil(err, "failed to unmarshall bson")

	// Third CanUpdateSampleSource test is with update and insert on all databases.
	info.loadAuthInfoFromConnectionStatus(&connStatusResult, "SAM")
	req.True(info.IsAllowedSampleSource(InsertPrivilege|UpdatePrivilege),
		"user should be able to update sample source")
}
