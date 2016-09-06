package mongodb

import (
	"testing"

	"gopkg.in/mgo.v2/bson"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoadAuthInfoFromConnectionStatus(t *testing.T) {
	Convey("Subject: loadAuthInfoFromConnectionStatus", t, func() {

		info := &Info{
			Databases: map[DatabaseName]*DatabaseInfo{
				"test1": &DatabaseInfo{
					Name: "test1",
					Collections: map[CollectionName]*CollectionInfo{
						"a": &CollectionInfo{Name: "a"},
						"b": &CollectionInfo{Name: "b"},
					},
				},
				"test2": &DatabaseInfo{
					Name: "test2",
					Collections: map[CollectionName]*CollectionInfo{
						"c": &CollectionInfo{Name: "c"},
						"d": &CollectionInfo{Name: "d"},
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
						// default for test1
						"resource": bson.M{"db": "test1", "collection": ""},
						"actions":  []interface{}{"role1", "find", "role2"},
					},
					bson.M{
						"resource": bson.M{"db": "test1", "collection": "b"},
						"actions":  []interface{}{"role1"},
					},
					bson.M{
						// default for test2
						"resource": bson.M{"db": "test2", "collection": ""},
						"actions":  []interface{}{},
					},
					bson.M{
						"resource": bson.M{"db": "test2", "collection": "d"},
						"actions":  []interface{}{"find"},
					},
				},
			},
		}

		resultBson, err := bson.Marshal(result)
		So(err, ShouldBeNil)

		var connStatusResult connectionStatusResult
		err = bson.Unmarshal(resultBson, &connStatusResult)
		So(err, ShouldBeNil)

		info.loadAuthInfoFromConnectionStatus(&connStatusResult)

		So(info.Privileges, ShouldEqual, AllPrivileges)

		So(len(info.Databases), ShouldBeGreaterThanOrEqualTo, 1)

		test1, ok := info.Databases[DatabaseName("test1")]
		So(ok, ShouldBeTrue)
		So(len(test1.Collections), ShouldEqual, 2)
		So(test1.Privileges, ShouldEqual, AllPrivileges)

		a, ok := test1.Collections[CollectionName("a")]
		So(ok, ShouldBeTrue)
		So(a.Privileges, ShouldEqual, AllPrivileges)

		b, ok := test1.Collections[CollectionName("b")]
		So(ok, ShouldBeTrue)
		So(b.Privileges, ShouldEqual, NoPrivileges)

		test2, ok := info.Databases[DatabaseName("test2")]
		So(ok, ShouldBeTrue)
		So(len(test2.Collections), ShouldEqual, 2)
		So(test2.Privileges, ShouldEqual, AllPrivileges)

		c, ok := test2.Collections[CollectionName("c")]
		So(ok, ShouldBeTrue)
		So(c.Privileges, ShouldEqual, NoPrivileges)

		d, ok := test2.Collections[CollectionName("d")]
		So(ok, ShouldBeTrue)
		So(d.Privileges, ShouldEqual, AllPrivileges)
	})
}
