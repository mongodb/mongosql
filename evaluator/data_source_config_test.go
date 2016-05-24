package evaluator

import (
	"sort"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const numInformationSchemaColumns = 30

type fakeAuthProvider struct{}

func (p *fakeAuthProvider) IsDatabaseAllowed(dbName string) bool {
	return strings.HasPrefix(dbName, "test")
}

func (p *fakeAuthProvider) IsCollectionAllowed(dbName string, colName string) bool {
	return strings.Contains(colName, "a")
}

func TestSchemaDataSourceIter(t *testing.T) {
	env := setupEnv(t)

	gatherValues := func(table, column string, iter Iter) []string {
		values := []string{}
		row := &Row{}
		for iter.Next(row) {
			if col, ok := row.GetField(table, column); ok {
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
		ctx := &ExecutionCtx{}

		Convey("when iterating over tables", func() {
			plan := NewSchemaDataSourceStage(env.cfgOne, "tables", "")

			Convey("should return all tables when authentication is disabled", func() {

				ctx.AuthProvider = &fixedAuthProvider{true}
				iter, err := plan.Open(ctx)
				So(err, ShouldBeNil)

				names := gatherValues("tables", "TABLE_NAME", iter)
				sort.Strings(names)

				So(len(names), ShouldEqual, 10)

				expectedNames := []string{"COLUMNS", "SCHEMATA", "TABLES", "bar", "bar", "bar", "baz", "foo", "foo", "silly"}
				So(names, ShouldResemble, expectedNames)
				So(iter.Close(), ShouldBeNil)
			})

			Convey("should return allowed tables when authentication is enabled", func() {
				ctx.AuthProvider = &fakeAuthProvider{}
				iter, err := plan.Open(ctx)
				So(err, ShouldBeNil)

				names := gatherValues("tables", "TABLE_NAME", iter)
				sort.Strings(names)

				So(len(names), ShouldEqual, 6)

				expectedNames := []string{"COLUMNS", "SCHEMATA", "TABLES", "bar", "bar", "baz"}
				So(names, ShouldResemble, expectedNames)
				So(iter.Close(), ShouldBeNil)
			})
		})

		Convey("when iterating over columns", func() {
			plan := NewSchemaDataSourceStage(env.cfgOne, "columns", "")

			Convey("should return all columns when authentication is disabled", func() {
				ctx.AuthProvider = &fixedAuthProvider{true}
				iter, err := plan.Open(ctx)
				So(err, ShouldBeNil)

				names := gatherValues("columns", "COLUMN_NAME", iter)
				names = names[numInformationSchemaColumns:]
				sort.Strings(names)

				So(len(names), ShouldEqual, 22)

				expectedNames := []string{"_id", "_id", "_id", "_id", "_id", "a", "a", "a", "amount", "b", "b", "b", "c", "c", "d", "e", "e", "f", "f", "name", "orderid", "orderid"}
				So(names, ShouldResemble, expectedNames)
				So(iter.Close(), ShouldBeNil)
			})

			Convey("should return allowed columns when authentication is enabled", func() {
				ctx.AuthProvider = &fakeAuthProvider{}
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
