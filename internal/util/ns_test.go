package util

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMatcher(t *testing.T) {
	matchers := []string{
		`*.user*`,
		`pr\*d.*`,
		`olea.*`,
		`paren]o.*`,
		`carr\ie.*`,
		`dsla\\xp.*`,
		`tmp|.*`,
		`kathy.*bo*x*`,
	}

	Convey("when using a matcher", t, func() {
		m, err := NewMatcher(matchers)
		So(m, ShouldNotBeNil)
		So(err, ShouldBeNil)

		message := fmt.Sprintf("non-generic matchers should match selectively %v",
			strings.Join(matchers, ", "))

		Convey(message, func() {
			So(m.Has("olaa.bobb"), ShouldBeFalse)
			So(m.Has("olea.bobb"), ShouldBeTrue)
			So(m.Has(`carr\ie.bobb`), ShouldBeFalse)
			So(m.Has("carrie.bobb"), ShouldBeTrue)
			So(m.Has(`dsla\xp.bobb`), ShouldBeTrue)
			So(m.Has("kathy.box"), ShouldBeTrue)
			So(m.Has("kathy.borx"), ShouldBeTrue)
			So(m.Has("paren]o.borx"), ShouldBeTrue)
			So(m.Has("kathy.boxer"), ShouldBeTrue)
			So(m.Has("kathy.rocks.boxs"), ShouldBeTrue)
			So(m.Has("stuff.user"), ShouldBeTrue)
			So(m.Has("stuf]f.user"), ShouldBeTrue)
			So(m.Has("stuff.users"), ShouldBeTrue)
			So(m.Has("stuff.users"), ShouldBeTrue)
			So(m.Has("pr*d.users"), ShouldBeTrue)
			So(m.Has("pr*d.magic"), ShouldBeTrue)
			So(m.Has(`pr*d\.magic`), ShouldBeFalse)
			So(m.Has("prod.magic"), ShouldBeFalse)
			So(m.Has("pr*d.turbo.encabulators"), ShouldBeTrue)
			So(m.Has("st*ging.turbo.encabulators"), ShouldBeFalse)
		})

		Convey(`a generic matcher "*.*" should match everything`, func() {
			m, err := NewMatcher([]string{"*.*"})
			So(m, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(m.Has("stuff"), ShouldBeFalse)
			So(m.Has("stuff.user"), ShouldBeTrue)
			So(m.Has("stuff.users"), ShouldBeTrue)
			So(m.Has("prod.turbo.encabulators"), ShouldBeTrue)
		})
	})

	Convey("with invalid matcher", t, func() {
		Convey("'$.user$'", func() {
			_, err := NewMatcher([]string{"$.user$"})
			So(err, ShouldNotBeNil)
		})
		Convey("'*.user$'", func() {
			_, err := NewMatcher([]string{"*.user$"})
			So(err, ShouldNotBeNil)
		})
		Convey("empty matcher", func() {
			_, err := NewMatcher(nil)
			So(err, ShouldNotBeNil)
		})
	})
}
