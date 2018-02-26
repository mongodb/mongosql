package flags

import "flag"

// These flags allow MongoDB-specific options to be passed to an
// invocation of go test.
var (
	// ServerVersion holds the version of MongoDB against which
	// tests are being run. It is used to decide which tests
	// should be skipped based on 'min_server_version'.
	ServerVersion = flag.String("serverVersion", "3.7", "The version of MongoDB against which these tests are being run, used to decide which tests should be skipped based on min_server_version")
	// Host is the address of the MongoDB cluster to run tests against.
	Host = flag.String("mongoHost", "127.0.0.1", "")
	// Port is the port of the MongoDB cluster to run tests against.
	Port = flag.String("mongoPort", "27017", "")
)
