package sample_test

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	. "github.com/10gen/sqlproxy/internal/sample"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewNamespace(t *testing.T) {
	versionID := bson.NewObjectId()

	Convey("When creating a new namespace", t, func() {

		ns := NewNamespace(db1, c2, versionID)

		Convey("the database, collection and version id should be correctly set", func() {
			So(ns.Database, ShouldEqual, db1)
			So(ns.Collection, ShouldEqual, c2)
			So(ns.VersionID, ShouldEqual, versionID)
		})

	})
}
