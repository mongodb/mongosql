package evaluator

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/schema"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
)

func TestOrderByStage(t *testing.T) {
	runTest := func(orderby *OrderByStage, collation *collation.Collation, rows []bson.D, expectedIds []int) {

		ctx := &ExecutionCtx{}

		ts := NewBSONSourceStage(1, tableOneName, collation, rows)

		orderby.source = ts
		iter, err := orderby.Open(ctx)
		So(err, ShouldBeNil)

		row := &Row{}

		i := 0

		for iter.Next(row) {
			So(row.Data[0].Data, ShouldEqual, SQLInt(expectedIds[i]))
			row = &Row{}
			i++
		}

		So(iter.Close(), ShouldBeNil)
		So(iter.Err(), ShouldBeNil)
	}

	Convey("Subject: OrderByStage", t, func() {
		Convey("default collation", func() {

			collation := collation.Default

			data := []bson.D{
				bson.D{{"_id", 1}, {"a", "a"}, {"b", 7}},
				bson.D{{"_id", 2}, {"a", "A"}, {"b", 8}},
				bson.D{{"_id", 3}, {"a", "b"}, {"b", 8}},
				bson.D{{"_id", 4}, {"a", "B"}, {"b", 7}},
			}

			Convey("single sort keys should sort according to the direction specified", func() {

				Convey("asc", func() {

					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: true},
					}

					operator := &OrderByStage{
						terms: terms,
					}

					expected := []int{2, 4, 1, 3}
					runTest(operator, collation, data, expected)

				})

				Convey("desc", func() {

					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: false},
					}

					operator := &OrderByStage{
						terms: terms,
					}

					expected := []int{3, 1, 4, 2}

					runTest(operator, collation, data, expected)

				})

			})

			Convey("multiple sort keys should sort according to the direction specified", func() {

				Convey("asc + asc", func() {
					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: true},
						{expr: NewSQLColumnExpr(1, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: true},
					}

					expected := []int{2, 4, 1, 3}

					operator := &OrderByStage{
						terms: terms,
					}

					runTest(operator, collation, data, expected)
				})

				Convey("asc + desc", func() {
					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: true},
						{expr: NewSQLColumnExpr(1, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: false},
					}

					operator := &OrderByStage{
						terms: terms,
					}

					expected := []int{2, 4, 1, 3}

					runTest(operator, collation, data, expected)

				})

				Convey("desc + asc", func() {
					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: false},
						{expr: NewSQLColumnExpr(1, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: true},
					}

					operator := &OrderByStage{
						terms: terms,
					}

					expected := []int{3, 1, 4, 2}

					runTest(operator, collation, data, expected)

				})

				Convey("desc + desc", func() {
					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: false},
						{expr: NewSQLColumnExpr(1, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: false},
					}

					operator := &OrderByStage{
						terms: terms,
					}

					expected := []int{3, 1, 4, 2}

					runTest(operator, collation, data, expected)

				})
			})

		})

		Convey("utf8_general_ci", func() {

			collation := collation.Must(collation.Get("utf8_general_ci"))

			data := []bson.D{
				bson.D{{"_id", 1}, {"a", "a"}, {"b", 7}},
				bson.D{{"_id", 2}, {"a", "A"}, {"b", 8}},
				bson.D{{"_id", 3}, {"a", "b"}, {"b", 8}},
				bson.D{{"_id", 4}, {"a", "B"}, {"b", 7}},
			}

			Convey("single sort keys should sort according to the direction specified", func() {

				Convey("asc", func() {

					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: true},
					}

					operator := &OrderByStage{
						terms: terms,
					}

					expected := []int{1, 2, 3, 4}
					runTest(operator, collation, data, expected)

				})

				Convey("desc", func() {

					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: false},
					}

					operator := &OrderByStage{
						terms: terms,
					}

					expected := []int{3, 4, 1, 2}

					runTest(operator, collation, data, expected)

				})

			})

			Convey("multiple sort keys should sort according to the direction specified", func() {

				Convey("asc + asc", func() {
					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: true},
						{expr: NewSQLColumnExpr(1, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: true},
					}

					expected := []int{1, 2, 4, 3}

					operator := &OrderByStage{
						terms: terms,
					}

					runTest(operator, collation, data, expected)
				})

				Convey("asc + desc", func() {
					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: true},
						{expr: NewSQLColumnExpr(1, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: false},
					}

					operator := &OrderByStage{
						terms: terms,
					}

					expected := []int{2, 1, 3, 4}

					runTest(operator, collation, data, expected)

				})

				Convey("desc + asc", func() {
					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: false},
						{expr: NewSQLColumnExpr(1, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: true},
					}

					operator := &OrderByStage{
						terms: terms,
					}

					expected := []int{4, 3, 1, 2}

					runTest(operator, collation, data, expected)

				})

				Convey("desc + desc", func() {
					terms := []*orderByTerm{
						{expr: NewSQLColumnExpr(1, tableOneName, "a", schema.SQLVarchar, schema.MongoString), ascending: false},
						{expr: NewSQLColumnExpr(1, tableOneName, "b", schema.SQLInt, schema.MongoInt), ascending: false},
					}

					operator := &OrderByStage{
						terms: terms,
					}

					expected := []int{3, 4, 2, 1}

					runTest(operator, collation, data, expected)

				})
			})

		})
	})

}
