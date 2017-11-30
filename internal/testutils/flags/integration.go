package flags

import (
	"flag"
)

// test flags
var (
	DbAddr            = flag.String("dbAddr", "127.0.0.1:3307", "")
	ClientPemFile     = flag.String("clientPemFile", "testdata/resources/x509gen/client.pem", "")
	Automate          = flag.String("automate", "none", "Pieces of infrastructure to automate (none|data)")
	MaxTimeSecs       = flag.Int64("maxTimeSecs", 600, "maximum test runtime limit (seconds)")
	RunSkipped        = flag.Bool("all", false, "also run tests with skip=true")
	DriverCompression = flag.Bool("compress", false, "use MySQL wire compression when running queries")
)
