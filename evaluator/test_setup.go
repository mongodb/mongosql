package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/schema"
	"gopkg.in/mgo.v2"
)

var (
	session                      *mgo.Session
	cfgOne, cfgThree             *schema.Schema
	collectionOne, collectionTwo *mgo.Collection
)

func init() {

	var err error

	cfgOne, err = schema.ParseSchemaData(testSchema1)
	if err != nil {
		panic(fmt.Sprintf("error parsing config1: %v", err))
	}

	cfgThree, err = schema.ParseSchemaData(testSchema3)
	if err != nil {
		panic(fmt.Sprintf("error parsing config3: %v", err))
	}

	session, err = mgo.Dial(cfgOne.Url)
	if err != nil {
		panic(fmt.Sprintf("error creating evaluator: %v", err))
	}

	collectionOne = session.DB(dbOne).C(tableOneName)
	collectionTwo = session.DB(dbOne).C(tableTwoName)

}
