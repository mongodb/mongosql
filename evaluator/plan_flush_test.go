package evaluator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/evaluator"
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

func (*fakeFlushServerCtx) Alter(context.Context, []*schema.Alteration) (*schema.Schema, error) {
	return nil, fmt.Errorf("not implemented")
}

func (*fakeFlushServerCtx) IsProcessOwner(string, uint32) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (*fakeFlushServerCtx) IsAdminUser(string, string) bool {
	return false
}

func (*fakeFlushServerCtx) Kill(uint32, uint32, evaluator.KillScope) error {
	return fmt.Errorf("not implemented")
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

	Convey("Subject: Flush Sample", t, func() {
		ctx := &evaluator.ExecutionCtx{}
		svrCtx := &fakeFlushServerCtx{}
		ctx.ConnectionCtx = &fakeConnectionCtx{
			server: svrCtx,
		}

		Convey("When flush sample is invoked and no error occurred", func() {
			cmd := evaluator.NewFlushCommand(evaluator.FlushSample)

			err := cmd.Execute(ctx).Run()
			So(err, ShouldBeNil)
			So(svrCtx.resampleCalled, ShouldBeTrue)
		})

		Convey("When flush sample is invoked and an error occurred", func() {
			svrCtx.shouldResampleError = true
			cmd := evaluator.NewFlushCommand(evaluator.FlushSample)

			err := cmd.Execute(ctx).Run()
			So(err, ShouldNotBeNil)
		})
	})
}
