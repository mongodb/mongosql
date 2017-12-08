package evaluator_test

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

type fakeConnectionCtx struct {
	variables *variable.Container
	info      *mongodb.Info
	server    evaluator.ServerCtx
}

func (*fakeConnectionCtx) LastInsertId() int64 {
	return 11
}
func (*fakeConnectionCtx) Logger(_ string) *log.Logger {
	lg := log.GlobalLogger()
	return &lg
}
func (*fakeConnectionCtx) RowCount() int64 {
	return 21
}
func (*fakeConnectionCtx) Catalog() *catalog.Catalog {
	return nil
}
func (*fakeConnectionCtx) UpdateCatalog(*schema.Schema) error {
	return nil
}
func (*fakeConnectionCtx) ConnectionID() uint32 {
	return 42
}
func (*fakeConnectionCtx) Context() context.Context {
	return context.Background()
}
func (*fakeConnectionCtx) DB() string {
	return "test"
}
func (*fakeConnectionCtx) GetStartupInfo() []string {
	return []string{}
}
func (*fakeConnectionCtx) Kill(id uint32, scope evaluator.KillScope) error {
	return nil
}
func (f *fakeConnectionCtx) Server() evaluator.ServerCtx {
	return f.server
}
func (*fakeConnectionCtx) Session() *mongodb.Session {
	return nil
}
func (*fakeConnectionCtx) User() string {
	return "test user"
}
func (f *fakeConnectionCtx) Variables() *variable.Container {
	if f.variables == nil {
		f.variables = variable.NewSessionContainer(variable.NewGlobalContainer(nil))
	}
	f.variables.MongoDBInfo = f.info
	return f.variables
}

func createTestConnectionCtx(info *mongodb.Info) evaluator.ConnectionCtx {
	return &fakeConnectionCtx{info: info}
}

func createTestExecutionCtx(info *mongodb.Info) *evaluator.ExecutionCtx {
	return &evaluator.ExecutionCtx{
		ConnectionCtx: createTestConnectionCtx(info),
	}
}

// getMongoDBInfo returns Info without looking up the information in MongoDB by setting
// all privileges to the specified privileges.
func getMongoDBInfo(versionArray []uint8, sch *schema.Schema, privileges mongodb.Privilege) *mongodb.Info {
	if len(versionArray) == 0 {
		versionArray = []uint8{3, 4, 0}
	}

	versionString := ""

	for _, entry := range versionArray {
		versionString = fmt.Sprintf("%v.", entry)
	}

	i := &mongodb.Info{
		Privileges:   privileges,
		Databases:    make(map[mongodb.DatabaseName]*mongodb.DatabaseInfo),
		Version:      versionString[1:],
		VersionArray: versionArray,
	}

	for _, db := range sch.Databases {
		dbInfo := &mongodb.DatabaseInfo{
			Privileges:  privileges,
			Name:        mongodb.DatabaseName(db.Name),
			Collections: make(map[mongodb.CollectionName]*mongodb.CollectionInfo),
		}

		i.Databases[dbInfo.Name] = dbInfo

		for _, col := range db.Tables {
			if _, ok := dbInfo.Collections[mongodb.CollectionName(col.Name)]; ok {
				continue
			}

			colInfo := &mongodb.CollectionInfo{
				Privileges: privileges,
				Name:       mongodb.CollectionName(col.Name),
			}

			dbInfo.Collections[colInfo.Name] = colInfo
		}
	}

	return i
}
