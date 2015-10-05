package connector

import (
	"database/sql"
	"github.com/erh/mongo-sql-temp/config"
	"github.com/erh/mongo-sql-temp/proxy"
	"github.com/erh/mongo-sql-temp/translator"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var sampleConfig = `{
  "addr": "127.0.0.1:3307", 
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
}`

func testServer(cfg *config.Config) (*proxy.Server, error) {
	evaluator, err := translator.NewEvaluator(cfg)
	if err != nil {
		return nil, err
	}
	return proxy.NewServer(cfg, evaluator)
}

func TestRoundtrip(t *testing.T) {
	Convey("With sample config on test server", t, func() {
		cfg, err := config.ParseConfigData([]byte(sampleConfig))
		So(err, ShouldBeNil)
		srv, err := testServer(cfg)
		So(err, ShouldBeNil)
		go srv.Run()
		Convey("Running query against server should succeed", func() {
			db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3307)/test")
			So(err, ShouldBeNil)
			defer db.Close()
			var outNum int
			err = db.QueryRow("SELECT b FROM test.foo").Scan(&outNum)
			So(err, ShouldBeNil)
		})
	})
}
