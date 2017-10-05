package testutils

import (
	"flag"
)

// test flags
var (
	ServerVersion     = flag.String("serverVersion", "3.4", "The version of mongodb against which these tests are being run")
	MongoHost         = flag.String("mongoHost", "127.0.0.1", "")
	MongoPort         = flag.String("mongoPort", "27017", "")
	DbAddr            = flag.String("dbAddr", "127.0.0.1:3307", "")
	ClientPemFile     = flag.String("clientPemFile", "testdata/resources/x509gen/client.pem", "")
	Automate          = flag.String("automate", "none", "Pieces of infrastructure to automate (none|data)")
	MaxTimeSecs       = flag.Int64("maxTimeSecs", 600, "maximum test runtime limit (seconds)")
	RunSkipped        = flag.Bool("all", false, "also run tests with skip=true")
	DriverCompression = flag.Bool("compress", false, "use MySQL wire compression when running queries")
)
