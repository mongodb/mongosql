package collation_test

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCharset(t *testing.T) {
	Convey("Subject: GetCharset", t, func() {

		Convey("With a valid CharsetName", func() {
			subject, err := collation.GetCharset(collation.CharsetName("utf8"))
			So(err, ShouldBeNil)
			So(subject.Name, ShouldEqual, collation.CharsetName("utf8"))
			So(subject.DefaultCollationName, ShouldEqual, collation.Name("utf8_general_ci"))
		})

		Convey("With an invalid CharsetName", func() {
			_, err := collation.GetCharset(collation.CharsetName("asdfagewqwre"))
			So(err, ShouldNotBeNil)
		})
	})
}

func TestMustGetCharset(t *testing.T) {
	Convey("Subject: MustGetCharset", t, func() {

		Convey("With a valid CharsetName", func() {
			subject := collation.MustCharset(collation.GetCharset(collation.CharsetName("utf8")))
			So(subject.Name, ShouldEqual, collation.CharsetName("utf8"))
			So(subject.DefaultCollationName, ShouldEqual, collation.Name("utf8_general_ci"))
		})

		Convey("With an invalid CharsetName", func() {
			f := func() { collation.MustCharset(collation.GetCharset(collation.CharsetName("asdfasdfqweg"))) }
			So(f, ShouldPanic)
		})
	})
}
