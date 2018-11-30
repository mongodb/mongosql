package flags

import (
	"flag"
)

var (
	// DbAddr holds the address of the mongosqld the tests will run against.
	DbAddr = flag.String("dbAddr", "127.0.0.1:3307", "")

	// ClientPEMKeyFile holds the location of the client PEM key file tests will utilize.
	ClientPEMKeyFile = flag.String("clientPEMKeyFile", "testdata/resources/x509gen/client.pem", "")

	// Automate determines what pieces of infrastructure to restore during testing.
	Automate = flag.String("automate", "none", "Pieces of infrastructure to automate (none|data[,schema])")

	// RunSkipped will run tests marked with `skip=true` if true.
	RunSkipped = flag.Bool("all", false, "also run tests with skip=true")

	// SchemaMappingHeuristic will run tests marked with the correct SchemaMappingHeuristics only.
	SchemaMappingHeuristic = flag.String("schemaMappingHeuristic", "",
		"run tests with schema_mapping_heuristic=<lattice|majority|drdl>, skip those without")

	// DriverCompression will enable compression in the MySQL client used for testing.
	DriverCompression = flag.Bool("compress", false, "use MySQL wire compression when "+
		"running queries")

	// NoPushdown will cause the integration tests to be run with the pushdown
	// optimizer turned off when true.
	NoPushdown = flag.Bool("nopushdown", false, "run the integration tests without pushdown")
)
