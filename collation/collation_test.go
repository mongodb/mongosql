package collation_test

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
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

			test(subject, "a", "a", 0)
			test(subject, "a", "b", -1)
			test(subject, "b", "a", 1)
			test(subject, "a", "a", 0)
			test(subject, "a", "A", 1)
			test(subject, "A", "a", -1)
		})

		Convey("utf8_general_ci", func() {
			subject, err := collation.Get(collation.Name("utf8_general_ci"))
			So(err, ShouldBeNil)

			test(subject, "a", "a", 0)
			test(subject, "a", "b", -1)
			test(subject, "b", "a", 1)
			test(subject, "a", "a", 0)
			test(subject, "a", "A", 0)
			test(subject, "A", "a", 0)
		})

		Convey("From MongoDB", func() {
			Convey("only locale", func() {
				subject, err := collation.FromMongoDB(&mgo.Collation{
					Locale: "en_US",
				})
				So(err, ShouldBeNil)

				test(subject, "a", "A", -1)
				test(subject, "A", "a", 1)
				test(subject, "a", "á", -1)
				test(subject, "A", "á", -1)
			})

			Convey("locale and strength", func() {
				subject, err := collation.FromMongoDB(&mgo.Collation{
					Locale:   "en_US",
					Strength: 1,
				})
				So(err, ShouldBeNil)

				test(subject, "a", "A", 0)
				test(subject, "A", "a", 0)
				test(subject, "a", "á", 0)
				test(subject, "A", "á", 0)
			})
		})
	})
}

func TestFromToMongoDB(t *testing.T) {
	Convey("Subject: From and To MongoDB", t, func() {

		Convey("only locale", func() {
			result, err := collation.FromMongoDB(&mgo.Collation{
				Locale: "es",
			})

			So(result.ID, ShouldEqual, collation.ID(0))
			So(result.Name, ShouldEqual, collation.Name(""))
			So(result.Default, ShouldBeFalse)
			So(result.ID, ShouldEqual, collation.ID(0))
			So(result.SortLen, ShouldEqual, 8)
			So(err, ShouldBeNil)

			mc := collation.ToMongoDB(result)

			So(mc.Locale, ShouldEqual, "es")
			So(mc.CaseLevel, ShouldBeFalse)
			So(mc.CaseFirst, ShouldEqual, "")
			So(mc.CaseLevel, ShouldBeFalse)
			So(mc.Strength, ShouldEqual, 3)
			So(mc.NumericOrdering, ShouldBeFalse)
			So(mc.Alternate, ShouldEqual, "")
			So(mc.Backwards, ShouldBeFalse)
		})

		Convey("all options", func() {
			result, err := collation.FromMongoDB(&mgo.Collation{
				Locale:          "de-AU",
				CaseLevel:       true,
				CaseFirst:       "lower",
				Strength:        2,
				NumericOrdering: true,
				Alternate:       "shifted",
				Backwards:       true,
			})

			So(result.ID, ShouldEqual, collation.ID(0))
			So(result.Name, ShouldEqual, collation.Name(""))
			So(result.Default, ShouldBeFalse)
			So(result.ID, ShouldEqual, collation.ID(0))
			So(result.SortLen, ShouldEqual, 8)
			So(err, ShouldBeNil)

			mc := collation.ToMongoDB(result)

			So(mc.Locale, ShouldEqual, "de-AU")
			So(mc.CaseLevel, ShouldBeTrue)
			So(mc.CaseFirst, ShouldEqual, "lower")
			So(mc.CaseLevel, ShouldBeTrue)
			So(mc.Strength, ShouldEqual, 2)
			So(mc.NumericOrdering, ShouldBeTrue)
			So(mc.Alternate, ShouldEqual, "shifted")
			So(mc.Backwards, ShouldBeTrue)
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
			So(subject.CharsetName, ShouldEqual, collation.CharsetName("utf8"))
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
			So(subject.CharsetName, ShouldEqual, collation.CharsetName("utf8"))
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
			So(subject.CharsetName, ShouldEqual, collation.CharsetName("utf8"))
		})

		Convey("With an invalid CharsetName", func() {
			f := func() { collation.Must(collation.Get(collation.Name("asdfasgewgqwegqweg"))) }
			So(f, ShouldPanic)
		})
	})
}
