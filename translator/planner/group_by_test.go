package planner

import (
	"fmt"
	//	"github.com/erh/mongo-sql-temp/config"
	. "github.com/smartystreets/goconvey/convey"
	//	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var (
	_ fmt.Stringer = nil
)

func TestGroupByOperator(t *testing.T) {

	Convey("With a simple test configuration...", t, func() {

		_ = []interface{}{
			bson.D{
				bson.DocElem{Name: "_id", Value: 5},
				bson.DocElem{Name: "a", Value: 6},
				bson.DocElem{Name: "b", Value: 7},
			},
			bson.D{
				bson.DocElem{Name: "_id", Value: 15},
				bson.DocElem{Name: "a", Value: 16},
				bson.DocElem{Name: "b", Value: 17},
			},
		}

		Convey("a group by operator with a single field should return columns ordered by the field", func() {

			_ = &GroupBy{
				source: &TableScan{
					tableName: tableOneName,
				},
			}

		})

	})
}
