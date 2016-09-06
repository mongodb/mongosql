package evaluator

import (
	"sort"
	"testing"

	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/variable"
	. "github.com/smartystreets/goconvey/convey"
)

const numInformationSchemaColumns = 34

func TestSchemaDataSourceIter(t *testing.T) {
	env := setupEnv(t)

	gatherValues := func(table, column string, iter Iter) []string {
		values := []string{}
		row := &Row{}
		for iter.Next(row) {
			if col, ok := row.GetField(1, table, column); ok {
				if colstr, ok := col.(SQLVarchar); ok {
					values = append(values, string(colstr))
				} else {
					t.Errorf("expected to get a SQLVarchar for %q", column)
				}
			} else {
				t.Errorf("expected to find %q in row", column)
			}
		}

		return values
	}

	Convey("Given a SchemaDataSource", t, func() {
		variables := variable.NewSessionContainer(variable.NewGlobalContainer())

		connCtx := &fakeConnectionCtx{variables}
		ctx := &ExecutionCtx{
			ConnectionCtx: connCtx,
		}

		Convey("when iterating over tables", func() {
			plan := NewSchemaDataSourceStage(1, env.cfgOne, "tables", "")

			Convey("should return all tables when authentication is disabled", func() {

				variables.MongoDBInfo = getMongoDBInfo(env.cfgOne, mongodb.AllPrivileges)
				iter, err := plan.Open(ctx)
				So(err, ShouldBeNil)

				names := gatherValues("tables", "TABLE_NAME", iter)
				sort.Strings(names)

				So(len(names), ShouldEqual, 12)

				expectedNames := []string{"COLUMNS", "GLOBAL_VARIABLES", "SCHEMATA", "SESSION_VARIABLES", "TABLES", "bar", "bar", "bar", "baz", "foo", "foo", "silly"}
				So(names, ShouldResemble, expectedNames)
				So(iter.Close(), ShouldBeNil)
			})

			Convey("should return allowed tables when authentication is enabled", func() {
				info := getMongoDBInfo(env.cfgOne, mongodb.NoPrivileges)
				info.Privileges = mongodb.AllPrivileges
				info.Databases["test"].Privileges = mongodb.AllPrivileges
				info.Databases["test"].Collections["bar"].Privileges = mongodb.AllPrivileges
				info.Databases["test"].Collections["baz"].Privileges = mongodb.AllPrivileges
				info.Databases["test2"].Privileges = mongodb.AllPrivileges
				info.Databases["test2"].Collections["bar"].Privileges = mongodb.AllPrivileges
				variables.MongoDBInfo = info
				iter, err := plan.Open(ctx)
				So(err, ShouldBeNil)

				names := gatherValues("tables", "TABLE_NAME", iter)
				sort.Strings(names)

				So(len(names), ShouldEqual, 8)

				expectedNames := []string{"COLUMNS", "GLOBAL_VARIABLES", "SCHEMATA", "SESSION_VARIABLES", "TABLES", "bar", "bar", "baz"}
				So(names, ShouldResemble, expectedNames)
				So(iter.Close(), ShouldBeNil)
			})
		})

		Convey("when iterating over columns", func() {
			plan := NewSchemaDataSourceStage(1, env.cfgOne, "columns", "")

			Convey("should return all columns when authentication is disabled", func() {
				variables.MongoDBInfo = getMongoDBInfo(env.cfgOne, mongodb.AllPrivileges)
				iter, err := plan.Open(ctx)
				So(err, ShouldBeNil)

				names := gatherValues("columns", "COLUMN_NAME", iter)
				names = names[numInformationSchemaColumns:]
				sort.Strings(names)

				So(len(names), ShouldEqual, 23)

				expectedNames := []string{"_id", "_id", "_id", "_id", "_id", "a", "a", "a", "amount", "b", "b", "b", "c", "c", "d", "e", "e", "f", "f", "g", "name", "orderid", "orderid"}
				So(names, ShouldResemble, expectedNames)
				So(iter.Close(), ShouldBeNil)
			})

			Convey("should return allowed columns when authentication is enabled", func() {
				info := getMongoDBInfo(env.cfgOne, mongodb.NoPrivileges)
				info.Privileges = mongodb.AllPrivileges
				info.Databases["test"].Privileges = mongodb.AllPrivileges
				info.Databases["test"].Collections["bar"].Privileges = mongodb.AllPrivileges
				info.Databases["test"].Collections["baz"].Privileges = mongodb.AllPrivileges
				info.Databases["test2"].Privileges = mongodb.AllPrivileges
				info.Databases["test2"].Collections["bar"].Privileges = mongodb.AllPrivileges
				variables.MongoDBInfo = info
				iter, err := plan.Open(ctx)
				So(err, ShouldBeNil)

				names := gatherValues("columns", "COLUMN_NAME", iter)
				names = names[numInformationSchemaColumns:]
				sort.Strings(names)

				So(len(names), ShouldEqual, 9)

				expectedNames := []string{"_id", "_id", "_id", "a", "a", "amount", "b", "b", "orderid"}
				So(names, ShouldResemble, expectedNames)
				So(iter.Close(), ShouldBeNil)
			})
		})
	})
}
