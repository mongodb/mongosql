package variable_test

import (
	"testing"

	"github.com/10gen/sqlproxy/variable"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGlobalVariableContainer(t *testing.T) {
	Convey("Subject: Global Container", t, func() {

		subject := variable.NewGlobalContainer()

		Convey("Get should fail with invalid system variable name", func() {
			_, err := subject.Get(variable.Name("test"), variable.GlobalScope, variable.SystemKind)
			So(err, ShouldNotBeNil)
		})

		Convey("Get should fail with invalid scope", func() {
			_, err := subject.Get(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind)
			So(err, ShouldNotBeNil)
		})

		Convey("Get should panic with invalid kind", func() {
			f := func() { subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.UserKind) }
			So(f, ShouldPanic)
		})

		Convey("Get should get the default value of a variable", func() {
			v, err := subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind)
			So(err, ShouldBeNil)
			So(v.Value, ShouldEqual, true)
		})

		Convey("Set should fail with invalid name", func() {
			err := subject.Set(variable.Name("test"), variable.GlobalScope, variable.SystemKind, false)
			So(err, ShouldNotBeNil)
		})

		Convey("Set should fail with invalid scope", func() {
			f := func() { subject.Set(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind, false) }
			So(f, ShouldPanic)
		})

		Convey("Set should fail with invalid kind", func() {
			f := func() { subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.UserKind, false) }
			So(f, ShouldPanic)
		})

		Convey("Set should fail for an invalid type", func() {
			err := subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind, "yeahaehasdh")
			So(err, ShouldNotBeNil)
		})

		Convey("Set should succeed when variable name and scope are valid", func() {
			err := subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind, false)
			So(err, ShouldBeNil)

			v, err := subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind)
			So(err, ShouldBeNil)
			So(v.Value, ShouldEqual, false)
		})
	})
}

func TestSessionVariableContainer(t *testing.T) {
	Convey("Subject: Session Container", t, func() {

		subject := variable.NewSessionContainer(variable.NewGlobalContainer())

		Convey("Get should fail with invalid system variable name", func() {
			_, err := subject.Get(variable.Name("test"), variable.SessionScope, variable.SystemKind)
			So(err, ShouldNotBeNil)
		})

		Convey("Get should get the default value of a variable", func() {
			v, err := subject.Get(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind)
			So(err, ShouldBeNil)
			So(v.Value, ShouldEqual, true)
		})

		Convey("Get should fallback to the parent with a different scope", func() {
			v, err := subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind)
			So(err, ShouldBeNil)
			So(v.Value, ShouldEqual, true)
		})

		Convey("Get should get nil for a non-existing user variable", func() {
			v, err := subject.Get(variable.Name("test"), variable.SessionScope, variable.UserKind)
			So(err, ShouldBeNil)
			So(v.Value, ShouldBeNil)
		})

		Convey("Set should fail with invalid name", func() {
			err := subject.Set(variable.Name("test"), variable.SessionScope, variable.SystemKind, false)
			So(err, ShouldNotBeNil)
		})

		Convey("Set should fail with invalid kind", func() {
			f := func() { subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.UserKind, false) }
			So(f, ShouldPanic)
		})

		Convey("Set should fail for an invalid type", func() {
			err := subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind, "yeahaehasdh")
			So(err, ShouldNotBeNil)
		})

		Convey("Set should succeed with parent scope", func() {
			err := subject.Set(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind, false)
			So(err, ShouldBeNil)

			v, err := subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind)
			So(err, ShouldBeNil)
			So(v.Value, ShouldEqual, false)

			v, err = subject.Get(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind)
			So(err, ShouldBeNil)
			So(v.Value, ShouldEqual, true)
		})

		Convey("Set should succeed with current scope", func() {
			err := subject.Set(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind, false)
			So(err, ShouldBeNil)

			v, err := subject.Get(variable.Name("autocommit"), variable.SessionScope, variable.SystemKind)
			So(err, ShouldBeNil)
			So(v.Value, ShouldEqual, false)

			v, err = subject.Get(variable.Name("autocommit"), variable.GlobalScope, variable.SystemKind)
			So(err, ShouldBeNil)
			So(v.Value, ShouldEqual, true)
		})

		Convey("Set should succeed for when setting a user variable", func() {
			err := subject.Set(variable.Name("test"), variable.SessionScope, variable.UserKind, "yeah")
			So(err, ShouldBeNil)

			v, err := subject.Get(variable.Name("test"), variable.SessionScope, variable.UserKind)
			So(err, ShouldBeNil)
			So(v.Value, ShouldEqual, "yeah")
		})
	})
}
