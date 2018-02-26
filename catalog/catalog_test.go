package catalog_test

import (
	"testing"

	"github.com/10gen/sqlproxy/catalog"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCatalog(t *testing.T) {
	Convey("Subject: Catalog", t, func() {
		c := catalog.New("def")

		Convey("NewCatalog", func() {
			So(string(c.Name), ShouldEqual, "def")
			So(len(c.Databases()), ShouldEqual, 0)
		})

		Convey("Unknown database should return an error", func() {
			_, err := c.Database("test")
			So(err, ShouldNotBeNil)
		})

		Convey("AddDatabase", func() {
			Convey("Should add the database if it doesn't already exist", func() {
				_, err := c.AddDatabase("test")
				So(err, ShouldBeNil)
				So(len(c.Databases()), ShouldEqual, 1)
			})

			Convey("Should return an error when a database already exists with the same name", func() {
				c.AddDatabase("test")

				_, err := c.AddDatabase("test")
				So(err, ShouldNotBeNil)

				_, err = c.AddDatabase("TEST")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Database", func() {
			c.AddDatabase("test")

			Convey("Should return an error if the database doesn't exist", func() {
				_, err := c.Database("blah")
				So(err, ShouldNotBeNil)
			})

			Convey("Should return the database when it exists", func() {
				d, err := c.Database("test")
				So(err, ShouldBeNil)
				So(d, ShouldNotBeNil)
				So(string(d.Name), ShouldEqual, "test")

				d, err = c.Database("TEST")
				So(err, ShouldBeNil)
				So(d, ShouldNotBeNil)
				So(string(d.Name), ShouldEqual, "test")
			})
		})

		Convey("Databases", func() {
			c.AddDatabase("test1")
			c.AddDatabase("test2")
			c.AddDatabase("test3")

			So(len(c.Databases()), ShouldEqual, 3)
		})
	})
}

func TestDatabase(t *testing.T) {
	Convey("Subject: Database", t, func() {
		d, _ := catalog.New("def").AddDatabase("test")

		Convey("NewDatabase", func() {

			So(string(d.Name), ShouldEqual, "test")
			So(len(d.Tables()), ShouldEqual, 0)
		})

		Convey("Unknown table should return an error", func() {
			_, err := d.Table("foo")
			So(err, ShouldNotBeNil)
		})

		Convey("AddTable", func() {
			Convey("Should add the database if it doesn't already exist", func() {
				t := catalog.NewInMemoryTable("foo")
				err := d.AddTable(t)
				So(err, ShouldBeNil)
				So(len(d.Tables()), ShouldEqual, 1)
			})

			Convey("Should return an error when a table already exists with the same name", func() {
				t := catalog.NewInMemoryTable("foo")
				d.AddTable(t)

				t2 := catalog.NewInMemoryTable("foo")
				err := d.AddTable(t2)
				So(err, ShouldNotBeNil)

				t3 := catalog.NewInMemoryTable("foo")
				err = d.AddTable(t3)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Table", func() {
			d.AddTable(catalog.NewInMemoryTable("foo"))

			Convey("Should return an error if the table doesn't exist", func() {
				_, err := d.Table("blah")
				So(err, ShouldNotBeNil)
			})

			Convey("Should return the table when it exists", func() {
				t, err := d.Table("foo")
				So(err, ShouldBeNil)
				So(d, ShouldNotBeNil)
				So(string(t.Name()), ShouldEqual, "foo")

				t, err = d.Table("FOO")
				So(err, ShouldBeNil)
				So(d, ShouldNotBeNil)
				So(string(t.Name()), ShouldEqual, "foo")
			})
		})

		Convey("Tables", func() {
			d.AddTable(catalog.NewInMemoryTable("foo1"))
			d.AddTable(catalog.NewInMemoryTable("foo2"))
			d.AddTable(catalog.NewInMemoryTable("foo3"))

			So(len(d.Tables()), ShouldEqual, 3)
		})
	})
}
