package sample_test

import (
	. "github.com/10gen/sqlproxy/internal/sample"
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewNamespace(t *testing.T) {
	versionId := bson.NewObjectId()

	Convey("When creating a new namespace", t, func() {

		ns := NewNamespace(db1, c2, versionId)

		Convey("the database, collection and version id should be correctly set", func() {
			So(ns.Database, ShouldEqual, db1)
			So(ns.Collection, ShouldEqual, c2)
			So(ns.VersionId, ShouldEqual, versionId)
		})

	})
}
