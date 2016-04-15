package evaluator

import (
	"testing"

	"gopkg.in/mgo.v2/bson"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthProvider(t *testing.T) {

	Convey("Given a connection status response", t, func() {

		Convey("on an unauthenticated connection", func() {
			result := bson.M{
				"authInfo": bson.M{
					"authenticatedUsers": []interface{}{},
				},
			}

			provider := loadAuthProviderFromConnectionStatus(&result)

			Convey("provider should be a fixedAuthProvider", func() {
				So(provider, ShouldResemble, &fixedAuthProvider{true})
			})
		})

		Convey("on an authenticated session", func() {

			result := bson.M{
				"authInfo": bson.M{
					"authenticatedUsers": []interface{}{
						bson.M{"user": "user1", "db": "test"},
					},
					"authenticatedUserPrivileges": []interface{}{
						bson.M{
							// default for test
							"resource": bson.M{"db": "test", "collection": ""},
							"actions":  []interface{}{"a", "find", "b"},
						},
						bson.M{
							"resource": bson.M{"db": "test", "collection": "baz"},
							"actions":  []interface{}{"a"},
						},
						bson.M{
							// default for test2
							"resource": bson.M{"db": "test2", "collection": ""},
							"actions":  []interface{}{},
						},
						bson.M{
							"resource": bson.M{"db": "test2", "collection": "baz"},
							"actions":  []interface{}{"find"},
						},
						bson.M{
							// default for test3
							"resource": bson.M{"db": "test3", "collection": ""},
							"actions":  []interface{}{"a"},
						},
						bson.M{
							"resource": bson.M{"db": "test3", "collection": "baz"},
							"actions":  []interface{}{"a", "b"},
						},
					},
				},
			}

			provider := loadAuthProviderFromConnectionStatus(&result)

			Convey("provider should be a mongoAuthProvider", func() {
				So(provider, ShouldHaveSameTypeAs, &mongoAuthProvider{})
			})

			Convey("IsDatabaseAllowed should return correct value", func() {
				So(provider.IsDatabaseAllowed("test"), ShouldBeTrue)
				So(provider.IsDatabaseAllowed("test2"), ShouldBeTrue)
				So(provider.IsDatabaseAllowed("test3"), ShouldBeFalse)
			})

			Convey("IsCollectionAllowed should return correct value", func() {
				So(provider.IsCollectionAllowed("test", "foo"), ShouldBeTrue)
				So(provider.IsCollectionAllowed("test", "bar"), ShouldBeTrue)
				So(provider.IsCollectionAllowed("test", "baz"), ShouldBeFalse)
				So(provider.IsCollectionAllowed("test2", "foo"), ShouldBeFalse)
				So(provider.IsCollectionAllowed("test2", "bar"), ShouldBeFalse)
				So(provider.IsCollectionAllowed("test2", "baz"), ShouldBeTrue)
				So(provider.IsCollectionAllowed("test3", "foo"), ShouldBeFalse)
				So(provider.IsCollectionAllowed("test3", "bar"), ShouldBeFalse)
				So(provider.IsCollectionAllowed("test3", "baz"), ShouldBeFalse)
			})
		})
	})
}
