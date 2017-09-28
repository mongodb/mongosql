package evaluator

import (
	"context"
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	_ fmt.Stringer = nil
)

type fakeFlushServerCtx struct {
	shouldResampleError bool

	resampleCalled bool
}

func (f *fakeFlushServerCtx) StartupInfo() []string {
	return []string{}
}

func (f *fakeFlushServerCtx) Resample(ctx context.Context) (*schema.Schema, error) {
	f.resampleCalled = true
	if f.shouldResampleError {
		return nil, fmt.Errorf("kaboom")
	}

	return &schema.Schema{}, nil
}

func TestFlushCommand(t *testing.T) {

	Convey("Subject: Flush", t, func() {
		ctx := &ExecutionCtx{}
		svrCtx := &fakeFlushServerCtx{}
		ctx.ConnectionCtx = &fakeConnectionCtx{
			server: svrCtx,
		}

		Convey("When flush sample is invoked and no error occurred", func() {
			cmd := NewFlushCommand(FlushSample)

			err := cmd.Execute(ctx).Run()
			So(err, ShouldBeNil)
			So(svrCtx.resampleCalled, ShouldBeTrue)
		})

		Convey("When flush sample is invoked and an error occurred", func() {
			svrCtx.shouldResampleError = true
			cmd := NewFlushCommand(FlushSample)

			err := cmd.Execute(ctx).Run()
			So(err, ShouldNotBeNil)
		})
	})
}
