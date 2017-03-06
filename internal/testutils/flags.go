package testutils

import (
	"flag"
)

// test flags
var (
	ServerVersion = flag.String("serverVersion", "3.4", "The version of mongodb against which these tests are being run")
	MongoHost     = flag.String("mongoHost", "127.0.0.1", "")
	MongoPort     = flag.String("mongoPort", "27017", "")
	DbAddr        = flag.String("dbAddr", "127.0.0.1:3307", "")
	ClientPemFile = flag.String("clientPemFile", "testdata/resources/client.pem", "")
	RestoreData   = flag.String("restoreData", "", "Suites whose data to restore before running tests")
)
