package translator

import (
	"context"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

func createConnectionCtx(info *mongodb.Info) *connCtx {
	return &connCtx{info: info}
}

type connCtx struct {
	variables     *variable.Container
	info          *mongodb.Info
	memoryMonitor *memory.Monitor
	server        evaluator.ServerCtx
}

func (*connCtx) LastInsertId() int64 {
	return 11
}

func (*connCtx) Logger(_ ...string) *log.Logger {
	return log.GlobalLogger()
}

func (*connCtx) RowCount() int64 {
	return 21
}

func (*connCtx) Catalog() *catalog.Catalog {
	return nil
}

func (*connCtx) UpdateCatalog(*schema.Schema) error {
	return nil
}

func (*connCtx) ConnectionID() uint32 {
	return 42
}

func (*connCtx) Context() context.Context {
	return context.Background()
}

func (*connCtx) DB() string {
	return "test"
}

func (*connCtx) GetStartupInfo() []string {
	return []string{}
}

func (*connCtx) Kill(id uint32, scope evaluator.KillScope) error {
	return nil
}

func (f *connCtx) MemoryMonitor() *memory.Monitor {
	if f.memoryMonitor == nil {
		f.memoryMonitor = memory.NewMonitor("connCtx", 0)
	}
	return f.memoryMonitor
}

func (f *connCtx) Server() evaluator.ServerCtx {
	return f.server
}

func (*connCtx) Session() *mongodb.Session {
	return nil
}

func (*connCtx) User() string {
	return "test user"
}

func (f *connCtx) Variables() *variable.Container {
	if f.variables == nil {
		gbl := variable.NewGlobalContainer(nil)
		gbl.MongoDBInfo = f.info
		ctn := variable.NewSessionContainer(gbl)
		ctn.MongoDBInfo = f.info
		f.variables = ctn
	}
	return f.variables
}

func (f *connCtx) VersionAtLeast(version ...uint8) bool {
	return f.Variables().MongoDBInfo.VersionAtLeast(version...)
}
