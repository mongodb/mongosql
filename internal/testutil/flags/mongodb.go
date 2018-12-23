package flags

import "flag"

// These flags allow MongoDB-specific options to be passed to an
// invocation of go test.
var (
	// Host is the address of the MongoDB cluster to run tests against.
	Host = flag.String("mongoHost", "127.0.0.1", "")

	// Port is the port of the MongoDB cluster to run tests against.
	Port = flag.String("mongoPort", "27017", "")
)
