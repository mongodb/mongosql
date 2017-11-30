package mongodb

import "flag"

// These flags allow MongoDB-specific options to be passed to an
// invocation of go test.
var (
	ServerVersion = flag.String("serverVersion", "3.4", "The version of mongodb against which these tests are being run")
	Host          = flag.String("mongoHost", "127.0.0.1", "")
	Port          = flag.String("mongoPort", "27017", "")
)
