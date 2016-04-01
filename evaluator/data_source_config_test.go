package evaluator

import (
	. "github.com/smartystreets/goconvey/convey"
	//"gopkg.in/mgo.v2/bson"
	"sort"
	"testing"
)

func TestSchemaDataSourceIter(t *testing.T) {
	env := setupEnv(t)
	cfgThree := env.cfgThree

	Convey("using config data source should iterate all columns", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgThree,
				Db:     dbOne,
			},
		}

		plan := &SchemaDataSourceStage{"columns", "", nil}

		fieldNames := []string{}
		iter, err := plan.Open(ctx)
		So(err, ShouldBeNil)

		row := &Row{}
		for iter.Next(row) {
			if col, ok := row.GetField("columns", "COLUMN_NAME"); ok {
				if colstr, ok := col.(SQLVarchar); ok {
					fieldNames = append(fieldNames, string(colstr))
				} else {
					t.Errorf("expected to get a SQLVarchar for COLUMN_NAME")
				}
			} else {
				t.Errorf("expected to find COLUMN_NAME in row")
			}
		}

		So(len(fieldNames), ShouldEqual, 10)

		names := []string{"_id", "a", "b", "c", "c", "d", "e", "f", "g", "h"}
		sort.Strings(fieldNames)
		So(names, ShouldResemble, fieldNames)
		So(iter.Err(), ShouldBeNil)
		So(iter.Close(), ShouldBeNil)
	})
}

func TestSchemaDataSourceIterTables(t *testing.T) {
	env := setupEnv(t)
	cfgOne := env.cfgOne

	Convey("using config data source should iterate tables", t, func() {

		ctx := &ExecutionCtx{
			PlanCtx: &PlanCtx{
				Schema: cfgOne,
				Db:     dbOne,
			},
		}

		plan := &SchemaDataSourceStage{"tables", "", nil}

		names := []string{}
		iter, err := plan.Open(ctx)
		So(err, ShouldBeNil)

		row := &Row{}
		for iter.Next(row) {
			if col, ok := row.GetField("tables", "TABLE_NAME"); ok {
				if colstr, ok := col.(SQLVarchar); ok {
					names = append(names, string(colstr))
				} else {
					t.Errorf("expected to get a SQLVarchar for TABLE_NAME")
				}
			} else {
				t.Errorf("expected to find TABLE_NAME in row")
			}
		}

		So(len(names), ShouldEqual, 7)

		tableNames := []string{"bar", "bar", "bar", "baz", "foo", "foo", "silly"}
		sort.Strings(names)
		So(names, ShouldResemble, tableNames)
		So(iter.Close(), ShouldBeNil)
	})
}
