package testutils

import (
	"fmt"

	"github.com/10gen/sqlproxy/log"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongorestore"

	"gopkg.in/mgo.v2/bson"
)

func restoreInline(host, port string, inline *InlineDataSet) error {
	connection := &toolsoptions.Connection{Host: host, Port: port}
	sessionProvider, err := toolsdb.NewSessionProvider(
		toolsoptions.ToolOptions{
			Auth:       &toolsoptions.Auth{},
			Connection: connection,
			SSL:        getSslOpts(),
		},
	)
	if err != nil {
		return err
	}

	sessionProvider.SetFlags(toolsdb.DisableSocketTimeout)
	session, err := sessionProvider.GetSession()
	if err != nil {
		return err
	}

	db := session.DB(inline.Db)
	c := db.C(inline.Collection)
	c.DropCollection() // don't care about the result

	if len(inline.Collation) > 0 {
		var result bson.D
		err = db.Run(bson.D{{"create", inline.Collection}, {"collation", inline.Collation}}, &result)
		if err != nil {
			return err
		}
	}

	bulk := c.Bulk()

	for _, d := range inline.Docs {
		doc, err := bsonutil.ConvertJSONValueToBSON(d)
		if err != nil {
			panic(fmt.Sprintf("unable to parse extended json %v error: %v", d, err))
		}
		bulk.Insert(doc)
	}

	_, err = bulk.Run()
	session.Close()
	return err
}

func restoreBSON(host, port, file string) error {
	connection := &toolsoptions.Connection{Host: host, Port: port}
	sessionProvider, err := toolsdb.NewSessionProvider(
		toolsoptions.ToolOptions{
			Auth:       &toolsoptions.Auth{},
			Connection: connection,
			SSL:        getSslOpts(),
		},
	)
	if err != nil {
		return err
	}

	sessionProvider.SetFlags(toolsdb.DisableSocketTimeout)
	log.SetVerbosity(&toolsoptions.Verbosity{Quiet: true})

	restorer := mongorestore.MongoRestore{
		ToolOptions: &toolsoptions.ToolOptions{
			Connection: connection,
			Namespace:  &toolsoptions.Namespace{},
		},
		InputOptions: &mongorestore.InputOptions{Gzip: true, Archive: file},
		OutputOptions: &mongorestore.OutputOptions{
			Drop:                   true,
			StopOnError:            true,
			NumParallelCollections: 1,
			NumInsertionWorkers:    10,
			MaintainInsertionOrder: true,
		},
		NSOptions: &mongorestore.NSOptions{
			DB:         "",
			Collection: "",
		},
		SessionProvider: sessionProvider,
	}

	return restorer.Restore()
}
