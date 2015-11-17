package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/config"
	"github.com/erh/mixer/sqlparser"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func projectTest(operator Operator, rows []bson.D, expectedRows []Values) {

	cfg, err := config.ParseConfigData(testConfig1)
	So(err, ShouldBeNil)

	session, err := mgo.Dial(cfg.Url)
	So(err, ShouldBeNil)

	collection := session.DB(dbName).C(tableOneName)
	collection.DropCollection()

	for _, row := range rows {
		So(collection.Insert(row), ShouldBeNil)
	}

	ctx := &ExecutionCtx{
		Config: cfg,
		Db:     dbName,
	}

	So(operator.Open(ctx), ShouldBeNil)

	row := &Row{}

	i := 0

	for operator.Next(row) {
		So(len(row.Data), ShouldEqual, 1)
		So(row.Data[0].Table, ShouldEqual, tableOneName)
		So(row.Data[0].Values, ShouldResemble, expectedRows[i])
		row = &Row{}
		i++
	}

	So(i, ShouldEqual, len(expectedRows))

	So(operator.Err(), ShouldBeNil)
	So(operator.Close(), ShouldBeNil)

	collection.DropCollection()

	session.Close()

}

func TestProjectOperator(t *testing.T) {

	Convey("A project operator...", t, func() {

		rows := []bson.D{
			bson.D{{"a", 6}, {"b", 9}},
			bson.D{{"a", 3}, {"b", 4}},
		}

		sExprs := SelectExpressions{
			SelectExpression{
				Column: Column{tableOneName, "a", "a", false},
				Expr:   &sqlparser.ColName{[]byte("a"), []byte(tableOneName)},
			},
			SelectExpression{
				Referenced: true,
				Column:     Column{tableOneName, "b", "b", false},
				Expr:       &sqlparser.ColName{[]byte("b"), []byte(tableOneName)},
			},
		}

		Convey("should filter out referenced columns in select expressions", func() {

			operator := &Project{
				sExprs: sExprs,
				source: &Select{
					source: &TableScan{
						tableName: tableOneName,
					},
					sExprs: sExprs,
				},
			}

			expected := []Values{{{"a", "a", SQLInt(6)}}, {{"a", "a", SQLInt(3)}}}

			projectTest(operator, rows, expected)
		})

		Convey("should not filter any results if no column is referenced", func() {
			sExprs[1].Referenced = false

			operator := &Project{
				sExprs: sExprs,
				source: &Select{
					source: &TableScan{
						tableName: tableOneName,
					},
					sExprs: sExprs,
				},
			}

			expected := []Values{{{"a", "a", SQLInt(6)}, {"b", "b", SQLInt(9)}}, {{"a", "a", SQLInt(3)}, {"b", "b", SQLInt(4)}}}

			projectTest(operator, rows, expected)
		})

	})
}
