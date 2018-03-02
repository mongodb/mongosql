package flags

import (
	"flag"
)

var (
	// DbAddr holds the address of the mongosqld the tests will run against.
	DbAddr = flag.String("dbAddr", "127.0.0.1:3307", "")
	// ClientPemFile holds the location of the client PEM key file
	// tests will utilize.
	ClientPemFile = flag.String("clientPemFile", "testdata/resources/x509gen/client.pem", "")
	// Automate determines what pieces of infrastructure to restore during testing.
	Automate = flag.String("automate", "none", "Pieces of infrastructure to automate (none|data)")
	// RunSkipped will run tests marked with `skip=true` if true.
	RunSkipped = flag.Bool("all", false, "also run tests with skip=true")
	// DriverCompression will enable compression in the MySQL client used for testing.
	DriverCompression = flag.Bool("compress", false, "use MySQL wire compression when "+
		"running queries")
)
