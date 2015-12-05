package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/config"
	"gopkg.in/mgo.v2"
)

var (
	session                      *mgo.Session
	cfgOne, cfgThree             *config.Config
	collectionOne, collectionTwo *mgo.Collection
)

func init() {

	var err error

	cfgOne, err = config.ParseConfigData(testConfig1)
	if err != nil {
		panic(fmt.Sprintf("error parsing config1: %v", err))
	}

	cfgThree, err = config.ParseConfigData(testConfig3)
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
