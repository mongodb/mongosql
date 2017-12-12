package mongodb

import "flag"

// These flags allow MongoDB-specific options to be passed to an
// invocation of go test.
var (
	ServerVersion = flag.String("serverVersion", "3.7", "The version of mongodb against which these tests are being run, used to decide which tests should be skipped based on min_server_version")
	Host          = flag.String("mongoHost", "127.0.0.1", "")
	Port          = flag.String("mongoPort", "27017", "")
)
