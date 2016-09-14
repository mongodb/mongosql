package collation_test

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCompareString(t *testing.T) {
	test := func(c *collation.Collation, a, b string, r int) {
		result := c.CompareString(a, b)
		So(result, ShouldEqual, r)
	}

	Convey("Subject: CompareString", t, func() {

		Convey("utf8_bin", func() {
			subject, err := collation.Get(collation.Name("utf8_bin"))
			So(err, ShouldBeNil)

			test(subject, "aaa", "aaa", 0)
			test(subject, "aaa", "aba", -1)
			test(subject, "aba", "aaa", 1)
			test(subject, "aaa", "aaa", 0)
			test(subject, "aaa", "AAA", -1)
			test(subject, "AAA", "aaa", 1)
		})

		Convey("utf8_general_ci", func() {
			subject, err := collation.Get(collation.Name("utf8_general_ci"))
			So(err, ShouldBeNil)

			test(subject, "aaa", "aaa", 0)
			test(subject, "aaa", "aba", -1)
			test(subject, "aba", "aaa", 1)
			test(subject, "aaa", "aaa", 0)
			test(subject, "aaa", "AAA", 0)
			test(subject, "AAA", "aaa", 0)
		})
	})
}

func TestGet(t *testing.T) {
	Convey("Subject: Get", t, func() {

		Convey("With a valid Name", func() {
			subject, err := collation.Get(collation.Name("utf8_bin"))
			So(err, ShouldBeNil)
			So(subject.Name, ShouldEqual, collation.Name("utf8_bin"))
			So(subject.ID, ShouldEqual, collation.ID(83))
			So(subject.DefaultCharsetName, ShouldEqual, collation.CharsetName("utf8"))
		})

		Convey("With an invalid Name", func() {
			_, err := collation.Get(collation.Name("asdfasgwqegqweg"))
			So(err, ShouldNotBeNil)
		})
	})
}

func TestGetByID(t *testing.T) {
	Convey("Subject: GetByID", t, func() {

		Convey("With a valid ID", func() {
			subject, err := collation.GetByID(collation.ID(83))
			So(err, ShouldBeNil)
			So(subject.Name, ShouldEqual, collation.Name("utf8_bin"))
			So(subject.ID, ShouldEqual, collation.ID(83))
			So(subject.DefaultCharsetName, ShouldEqual, collation.CharsetName("utf8"))
		})

		Convey("With an invalid ID", func() {
			_, err := collation.GetByID(collation.ID(0))
			So(err, ShouldNotBeNil)
		})
	})
}

func TestMust(t *testing.T) {
	Convey("Subject: Must", t, func() {

		Convey("With a valid Name", func() {
			subject := collation.Must(collation.Get(collation.Name("utf8_bin")))
			So(subject.Name, ShouldEqual, collation.Name("utf8_bin"))
			So(subject.ID, ShouldEqual, collation.ID(83))
			So(subject.DefaultCharsetName, ShouldEqual, collation.CharsetName("utf8"))
		})

		Convey("With an invalid CharsetName", func() {
			f := func() { collation.Must(collation.Get(collation.Name("asdfasgewgqwegqweg"))) }
			So(f, ShouldPanic)
		})
	})
}
