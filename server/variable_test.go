package server

import (
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGlobalVariableContainer(t *testing.T) {
	Convey("Subject: globalVariableContainer", t, func() {

		subject := newGlobalVariableContainer()

		_, ok := subject.getValue("test")
		So(ok, ShouldBeFalse)

		subject.setValue("test", evaluator.SQLInt(42))
		v, ok := subject.getValue("test")
		So(ok, ShouldBeTrue)
		So(v, ShouldEqual, evaluator.SQLInt(42))

		subject.setValue("test", evaluator.SQLNull)
		v, ok = subject.getValue("test")
		So(ok, ShouldBeTrue)
		So(v, ShouldResemble, evaluator.SQLNull)
	})
}

func TestSessionVariableContainer(t *testing.T) {
	Convey("Subject: sessionVariableContainer", t, func() {

		global := newGlobalVariableContainer()
		global.setValue("test", evaluator.SQLInt(42))

		subject := newSessionVariableContainer(global)

		Convey("session variables", func() {
			_, ok := subject.getSessionVariable("blah")
			So(ok, ShouldBeFalse)

			v, ok := subject.getSessionVariable("test")
			So(ok, ShouldBeTrue)
			So(v, ShouldEqual, evaluator.SQLInt(42))

			subject.setSessionVariable("test", evaluator.SQLInt(21))
			v, ok = subject.getSessionVariable("test")
			So(ok, ShouldBeTrue)
			So(v, ShouldEqual, evaluator.SQLInt(21))

			subject.setSessionVariable("test", evaluator.SQLNull)
			v, ok = subject.getSessionVariable("test")
			So(ok, ShouldBeTrue)
			So(v, ShouldResemble, evaluator.SQLNull)
		})

		Convey("user variables", func() {
			_, ok := subject.getUserVariable("test")
			So(ok, ShouldBeFalse)

			subject.setUserVariable("test", evaluator.SQLInt(42))
			v, ok := subject.getUserVariable("test")
			So(ok, ShouldBeTrue)
			So(v, ShouldEqual, evaluator.SQLInt(42))

			subject.setUserVariable("test", evaluator.SQLVarchar("funny"))
			v, ok = subject.getUserVariable("test")
			So(ok, ShouldBeTrue)
			So(v, ShouldEqual, evaluator.SQLVarchar("funny"))

			subject.setUserVariable("test", evaluator.SQLNull)
			v, ok = subject.getUserVariable("test")
			So(ok, ShouldBeTrue)
			So(v, ShouldResemble, evaluator.SQLNull)
		})
	})
}
