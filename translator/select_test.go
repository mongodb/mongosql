package translator

import (
	"fmt"
	"github.com/erh/mongo-sql-temp/config"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestSimple(t *testing.T) {
	var testConfigData = []byte(
		`
schema :
-
  url: localhost
  db: test2
  tables:
  -
     table: bar
     collection: test.select_test1
`)

	cfg, err := config.ParseConfigData(testConfigData)
	if err != nil {
		t.Fatal(err)
	}

	eval, err := NewEvalulator(cfg)
	if err != nil {
		t.Fatal(err)
	}

	session := eval.getSession()
	defer session.Close()

	collection := eval.getCollection(session, "test.select_test1")
	collection.DropCollection()
	collection.Insert(bson.M{"_id": 5, "a": 6, "b": 7})
	collection.Insert(bson.M{"_id": 15, "a": 16, "c": 17})

	names, values, err := eval.EvalSelect("test2", "select * from bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(values) != 2 {
		t.Fatal(fmt.Printf("wrong number of rows"))
	}

	if len(names) != 4 {
		t.Fatal(fmt.Printf("names length wrong"))
	}

	if names[0] != "_id" ||
		names[1] != "a" ||
		names[2] != "b" ||
		names[3] != "c" {
		t.Fatal(fmt.Printf("names array wrong %v", names))
	}

	if values[1][0] != 15 ||
		values[1][1] != 16 ||
		values[1][2] != nil ||
		values[1][3] != 17 {
		t.Fatal(fmt.Printf("2nd row data wrong", values[1]))
	}

	for _, row := range values {
		if len(names) != len(row) {
			t.Fatal("row/name count mismatch %v", row)
		}
	}

	// push down where

	names, values, err = eval.EvalSelect("test2", "select * from bar where a = 16", nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(values) != 1 {
		t.Fatal(fmt.Printf("wrong number of rows"))
	}

}

func TestSimplePipe(t *testing.T) {
	var testConfigData = []byte(
		`
schema :
-
  url: localhost
  db: test2
  tables:
  -
     table: bar
     collection: test.select_test2
     pipeline:
     -
        $unwind : "$x"
     -
        $limit : 10
`)

	cfg, err := config.ParseConfigData(testConfigData)
	if err != nil {
		t.Fatal(err)
	}

	eval, err := NewEvalulator(cfg)
	if err != nil {
		t.Fatal(err)
	}

	session := eval.getSession()
	defer session.Close()

	collection := eval.getCollection(session, "test.select_test2")
	collection.DropCollection()
	collection.Insert(bson.M{"_id": 5, "a": 6, "b": 7, "x": []int{10, 11}})
	collection.Insert(bson.M{"_id": 15, "a": 16, "c": 17, "x": 12})

	names, values, err := eval.EvalSelect("test2", "select * from bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 5 {
		t.Fatal(fmt.Printf("names length wrong %v", names))
	}

	if len(values) != 3 {
		t.Fatal(fmt.Printf("wrong number of values"))
	}

	if names[0] != "_id" ||
		names[1] != "a" ||
		names[2] != "b" ||
		names[3] != "x" ||
		names[4] != "c" {
		t.Fatal(fmt.Printf("names array wrong %v", names))
	}

	if values[2][0] != 15 ||
		values[2][1] != 16 ||
		values[2][2] != nil ||
		values[2][3] != 12 ||
		values[2][4] != 17 {
		t.Fatal(fmt.Printf("2nd row data wrong", values[1]))
	}

	for _, row := range values {
		if len(names) != len(row) {
			t.Fatal("row/name count mismatch %v", row)
		}
	}

	// push down where

	names, values, err = eval.EvalSelect("test2", "select * from bar where x = 11", nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(values) != 1 {
		t.Fatal(fmt.Printf("wrong number of values"))
	}

	if len(names) != 4 {
		t.Fatal(fmt.Printf("names length wrong %v", names))
	}

	if values[0][0] != 5 ||
		values[0][3] != 11 {
		t.Fatal(fmt.Printf("row data wrong", values[0]))
	}

}
