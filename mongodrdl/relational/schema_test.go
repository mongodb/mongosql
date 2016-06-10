package relational_test

import (
	"github.com/10gen/sqlproxy/mongodrdl/relational"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestColumnSorting(t *testing.T) {

	Convey("Columns should be sorted by name and mongo.Filters should be last", t, func() {
		columns := relational.ColumnSlice{
			&relational.Column{Name: "aaa", MongoType: "string"},
			&relational.Column{Name: "ccc", MongoType: "mongo.Filter"},
			&relational.Column{Name: "eee", MongoType: "string"},
			&relational.Column{Name: "fff", MongoType: "string"},
			&relational.Column{Name: "ddd", MongoType: "string"},
			&relational.Column{Name: "bbb", MongoType: "mongo.Filter"},
		}

		columns.Sort()

		So(columns[0].Name, ShouldEqual, "aaa")
		So(columns[1].Name, ShouldEqual, "ddd")
		So(columns[2].Name, ShouldEqual, "eee")
		So(columns[3].Name, ShouldEqual, "fff")
		So(columns[4].Name, ShouldEqual, "bbb")
		So(columns[5].Name, ShouldEqual, "ccc")
	})
}
