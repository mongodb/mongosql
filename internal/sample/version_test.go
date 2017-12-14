package sample_test

import (
	"testing"

	. "github.com/10gen/sqlproxy/internal/sample"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewVersion(t *testing.T) {
	Convey("Creating a new version", t, func() {
		const processName string = "random process name"
		version := NewVersion(processName)

		Convey("should set the protocol and process name correctly", func() {
			So(version.Protocol, ShouldEqual, CurrentProtocol)
			So(version.ProcessName, ShouldEqual, processName)
		})
	})
}
