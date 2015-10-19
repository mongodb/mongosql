package sqlproxy

import (
	"database/sql"
	"github.com/erh/mongo-sql-temp"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/proxy"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

var sampleConfig = []byte(`{
  "addr": "127.0.0.1:3456",
  "schema": [
    { "db": "test", "tables": [
        { "table": "foo", "collection": "test.foo", 
          "columns": [
            { "type": "int", "name": "a" }, 
            { "type": "string", "name": "b" }
          ]
        }
      ], 
    }
  ]
}`)

type testData struct {
	database   string
	collection string
	data       []bson.D
}

func testServer(cfg *config.Config) (*proxy.Server, error) {
	evaluator, err := sqlproxy.NewEvaluator(cfg)
	if err != nil {
		return nil, err
	}
	return proxy.NewServer(cfg, evaluator)
}

func populateTestData(cfg *config.Config, datasets []testData) error {
	session, err := mgo.Dial(cfg.Url)
	if err != nil {
		return err
	}
	defer session.Close()
	for _, dataset := range datasets {
		c := session.DB(dataset.database).C(dataset.collection)
		err := c.DropCollection()
		if err != nil {
			return err
		}
		for _, d := range dataset.data {
			err := c.Insert(d)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// runSQL will populate the collection for the given config with the array of data,
// then run the given SQL statement and return its results as an array of arrays.
func runSQL(statement string, cfg *config.Config) ([][]interface{}, error) {
	srv, err := testServer(cfg)
	So(err, ShouldBeNil)
	go srv.Run()

	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3456)/test")
	So(err, ShouldBeNil)
	defer db.Close()
	rows, err := db.Query(statement)
	So(err, ShouldBeNil)
	defer rows.Close()

	cols, err := rows.Columns()
	So(err, ShouldBeNil)

	result := [][]interface{}{}

	i := 0
	for rows.Next() {
		i += 1
		resultRow := make([]interface{}, len(cols))
		resultRowVals := make([]interface{}, len(cols))
		for i, _ := range resultRow {
			resultRow[i] = &resultRowVals[i]
		}
		if err := rows.Scan(resultRow...); err != nil {
			return nil, err
		}
		result = append(result, resultRowVals)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func compareResults(actual [][]interface{}, expected [][]interface{}) {
	So(actual, ShouldResemble, expected)
}

func TestRoundtrip(t *testing.T) {
	Convey("With sample config on test server", t, func() {
		cfg := config.Must(config.ParseConfigData(sampleConfig))

		err := populateTestData(cfg, []testData{
			{database: "test", collection: "foo",
				data: []bson.D{
					{{"a", "foo"}},
					{{"a", "bar"}},
					{{"a", "baz"}},
				}},
		})

		result, err := runSQL("SELECT a from foo", cfg)
		So(err, ShouldBeNil)
		So(result, ShouldResemble, [][]interface{}{
			{[]byte("foo")},
			{[]byte("bar")},
			{[]byte("baz")},
		})
	})
}
