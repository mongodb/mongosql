package mongodrdl_test

import (
	"fmt"
	"testing"

	"github.com/10gen/sqlproxy/internal/mongodrdl"
	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	t.Run("invalid_combo_with_uri", testInvalidComboWithURI)
	t.Run("invalid_ssl", testInvalidSSL)
	t.Run("precedence", testPrecedence)
	t.Run("valid_args", testValid)

	t.Run("help", func(t *testing.T) {
		req := require.New(t)
		opts, err := mongodrdl.NewDrdlOptions()
		req.NoError(err)
		req.NoError(opts.Parse([]string{"help"}))
		req.NoError(opts.Parse([]string{"-help"}))
		req.NoError(opts.Parse([]string{"--help"}))
	})

	t.Run("unknown command", func(t *testing.T) {
		req := require.New(t)
		opts, err := mongodrdl.NewDrdlOptions()
		req.NoError(err)
		req.NoError(opts.Parse([]string{"unknown"}))
		err = opts.Validate()
		req.Error(err)
	})

	t.Run("explicit sample command", func(t *testing.T) {
		req := require.New(t)
		opts, err := mongodrdl.NewDrdlOptions()
		req.NoError(err)
		req.NoError(opts.Parse([]string{"sample", "-d", "test"}))
		err = opts.Validate()
		req.NoError(err)
	})

	t.Run("extra positional args", func(t *testing.T) {
		req := require.New(t)
		opts, err := mongodrdl.NewDrdlOptions()
		req.NoError(err)
		req.Error(opts.Parse([]string{"sample", "foo"}))
	})

	t.Run("command not at beginning", func(t *testing.T) {
		req := require.New(t)
		opts, err := mongodrdl.NewDrdlOptions()
		req.NoError(err)
		req.NoError(opts.Parse([]string{"-d", "test", "sample"}))
		err = opts.Validate()
		req.NoError(err)
	})
}

func testInvalidComboWithURI(t *testing.T) {
	tests := []struct {
		description string
		args        []string
	}{
		{
			"host",
			[]string{
				"--host", "localhost",
				"--uri", "mongodb://localhost:27017",
			},
		},
		{
			"port",
			[]string{
				"--port", "27017",
				"--uri", "mongodb://localhost:27017",
			},
		},
		{
			"username",
			[]string{
				"--username", "user",
				"--uri", "mongodb://localhost:27017",
			},
		},
		{
			"password",
			[]string{
				"--password", "pass",
				"--uri", "mongodb://user:pass@localhost:27017",
			},
		},
		{
			"db",
			[]string{
				"--db", "db",
				"--uri", "mongodb://localhost:27017/db",
			},
		},
		{
			"authenticationDatabase",
			[]string{
				"--authenticationDatabase", "authDB",
				"--uri", "mongodb://user@localhost:27017/db?authSource=authDB",
			},
		},
		{
			"authenticationMechanism",
			[]string{
				"--authenticationMechanism", "SCRAM-SHA-1",
				"--uri", "mongodb://user:pass@localhost:27017/db?authMechanism=SCRAM-SHA-1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req := require.New(t)
			opts, err := mongodrdl.NewDrdlOptions()
			req.NoError(err)
			err = opts.Parse(test.args)
			req.EqualError(err, fmt.Sprintf("illegal argument combination: cannot specify --uri and --%s", test.description))
		})
	}
}

func testInvalidSSL(t *testing.T) {
	tests := []struct {
		description string
		args        []string
	}{
		{
			"ca file without ssl",
			[]string{
				"--host", "localhost",
				"--port", "6999",
				"--sslCAFile", "hello",
				"-d", "output",
			},
		},

		{
			"pem file without ssl",
			[]string{
				"--host", "localhost",
				"--port", "6999",
				"--sslPEMKeyFile", "hello",
				"-d", "output",
			},
		},

		{
			"pem pwd without ssl",
			[]string{
				"--host", "localhost",
				"--port", "6999",
				"--sslPEMKeyPassword", "hello",
				"-d", "output",
			},
		},

		{
			"crl file without ssl",
			[]string{
				"--host", "localhost",
				"--port", "6999",
				"--sslCRLFile", "hello",
				"-d", "output",
			},
		},

		{
			"allow invalid certs without ssl",
			[]string{
				"--host", "localhost",
				"--port", "6999",
				"--sslAllowInvalidCertificates",
				"-d", "output",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req := require.New(t)
			opts, err := mongodrdl.NewDrdlOptions()
			req.NoError(err)
			req.NoError(opts.Parse(test.args))

			expectedErr := "when specifying SSL options, SSL must be enabled with --ssl"
			err = opts.Validate()
			req.EqualError(err, expectedErr)
		})
	}
}

func testPrecedence(t *testing.T) {
	t.Run("port flag has higher precedence than host", func(t *testing.T) {
		tests := []struct {
			description   string
			args          []string
			expectedHosts []string
		}{
			{
				"host and port",
				[]string{
					"--host", "localhost",
					"--port", "6999",
				},
				[]string{"localhost:6999"},
			},
			{
				"host with port and port",
				[]string{
					"--host", "localhost:34325452",
					"--port", "6999",
				},
				[]string{"localhost:6999"},
			},
			{
				"host with port",
				[]string{
					"--host", "localhost:6999",
				},
				[]string{"localhost:6999"},
			},
			{
				"replset host and port",
				[]string{
					"--host", "testReplSet/localhost:27017,localhost:27018,localhost:27019",
					"--port", "6999",
				},
				[]string{"localhost:27017", "localhost:27018", "localhost:27019"},
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				req := require.New(t)
				opts, err := mongodrdl.NewDrdlOptions()
				req.NoError(err)
				req.NoError(opts.Parse(test.args))

				cs, err := opts.ConnString()
				req.NoError(err)

				for i, expectedHost := range test.expectedHosts {
					req.Equal(expectedHost, cs.Hosts[i], "incorrect host #%v", i)
				}
			})
		}
	})

	t.Run("kerberos flags have higher precedence than uri", func(t *testing.T) {
		tests := []struct {
			description         string
			args                []string
			expectedServiceName string
			expectedHostName    string
		}{
			{
				"neither_specifies",
				[]string{
					"--uri", "mongodb://localhost:27017",
				},
				"",
				"",
			},
			{
				"only_flags_specify",
				[]string{
					"--gssapiServiceName", "flagService",
					"--gssapiHostName", "flagHost",
					"--uri", "mongodb://localhost:27017",
				},
				"flagService",
				"flagHost",
			},
			{
				"only_uri_specify",
				[]string{
					// Note: SERVICE_HOST is not a valid authMechanismProperty for GSSAPI
					"--uri", "mongodb://user@localhost:27017/db?authMechanism=GSSAPI&authMechanismProperties=SERVICE_NAME:uriService",
				},
				"uriService",
				"",
			},
			{
				"both_specify",
				[]string{
					"--gssapiServiceName", "flagService",
					"--gssapiHostName", "flagHost",

					// Note: SERVICE_HOST is not a valid authMechanismProperty for GSSAPI
					"--uri", "mongodb://user@localhost:27017/db?authMechanism=GSSAPI&authMechanismProperties=SERVICE_NAME:uriService",
				},
				"flagService",
				"flagHost",
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				req := require.New(t)
				opts, err := mongodrdl.NewDrdlOptions()
				req.NoError(err)
				req.NoError(opts.Parse(test.args))

				cs, err := opts.ConnString()
				req.NoError(err)
				req.Equal(test.expectedServiceName, cs.AuthMechanismProperties["SERVICE_NAME"])
				req.Equal(test.expectedHostName, cs.AuthMechanismProperties["SERVICE_HOST"])
			})
		}
	})

	t.Run("ssl flags have higher precendence than uri", func(t *testing.T) {
		tests := []struct {
			description                 string
			args                        []string
			expectedEnabled             bool
			expectedSSLCAFile           string
			expectedSSLPEMKeyFile       string
			expectedSSLPEMKeyPassword   string
			expectedSSLCRLFile          string
			expectedSSLAllowInvalidCert bool
			expectedSSLAllowInvalidHost bool
			expectedSSLFipsMode         bool
			expectedMinimumTLSVersion   string
		}{
			{
				description: "neither_specifies",
				args: []string{
					"--uri", "mongodb://localhost:27017",
				},
				expectedEnabled:             false,
				expectedSSLCAFile:           "",
				expectedSSLPEMKeyFile:       "",
				expectedSSLPEMKeyPassword:   "",
				expectedSSLCRLFile:          "",
				expectedSSLAllowInvalidCert: false,
				expectedSSLAllowInvalidHost: false,
				expectedSSLFipsMode:         false,
				expectedMinimumTLSVersion:   "TLS1_1",
			},
			{
				description: "only_flags_specify",
				args: []string{
					"--ssl",
					"--sslCAFile", "flagCAFile",
					"--sslPEMKeyFile", "flagPEMKeyFile",
					"--sslPEMKeyPassword", "flagPEMKeyPassword",
					"--sslCRLFile", "flagCRLFile",
					"--sslAllowInvalidCertificates",
					"--sslAllowInvalidHostnames",
					"--sslFIPSMode", "flagFIPSMode",
					"--minimumTLSVersion", "TLS1_2",
					"--uri", "mongodb://localhost:27017",
				},
				expectedEnabled:             true,
				expectedSSLCAFile:           "flagCAFile",
				expectedSSLPEMKeyFile:       "flagPEMKeyFile",
				expectedSSLPEMKeyPassword:   "flagPEMKeyPassword",
				expectedSSLCRLFile:          "flagCRLFile",
				expectedSSLAllowInvalidCert: true,
				expectedSSLAllowInvalidHost: true,
				expectedSSLFipsMode:         true,
				expectedMinimumTLSVersion:   "TLS1_2",
			},
			{
				description: "only_uri_specify",
				args: []string{
					"--uri", "mongodb://localhost:27017/db?ssl=true&sslclientcertificatekeyfile=uriPEMKeyFile&sslclientcertificatekeypassword=uriPEMKeyPassword&sslinsecure=true&sslcertificateauthorityfile=uriCAFile",
				},
				expectedEnabled:             true,
				expectedSSLCAFile:           "uriCAFile",
				expectedSSLPEMKeyFile:       "uriPEMKeyFile",
				expectedSSLPEMKeyPassword:   "uriPEMKeyPassword",
				expectedSSLCRLFile:          "",
				expectedSSLAllowInvalidCert: true,
				expectedSSLAllowInvalidHost: true,
				expectedSSLFipsMode:         false,
				expectedMinimumTLSVersion:   "TLS1_1",
			},
			{
				description: "both_specify",
				args: []string{
					"--ssl",
					"--sslCAFile", "flagCAFile",
					"--sslPEMKeyFile", "flagPEMKeyFile",
					"--sslPEMKeyPassword", "flagPEMKeyPassword",
					"--sslCRLFile", "flagCRLFile",
					"--sslAllowInvalidCertificates",
					"--sslAllowInvalidHostnames",
					"--sslFIPSMode", "flagFIPSMode",
					"--minimumTLSVersion", "TLS1_2",
					"--uri", "mongodb://localhost:27017/db?ssl=true&sslclientcertificatekeyfile=uriPEMKeyFile&sslclientcertificatekeypassword=uriPEMKeyPassword&sslinsecure=true&sslcertificateauthorityfile=uriCAFile",
				},
				expectedEnabled:             true,
				expectedSSLCAFile:           "flagCAFile",
				expectedSSLPEMKeyFile:       "flagPEMKeyFile",
				expectedSSLPEMKeyPassword:   "flagPEMKeyPassword",
				expectedSSLCRLFile:          "flagCRLFile",
				expectedSSLAllowInvalidCert: true,
				expectedSSLAllowInvalidHost: true,
				expectedSSLFipsMode:         true,
				expectedMinimumTLSVersion:   "TLS1_2",
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				req := require.New(t)
				opts, err := mongodrdl.NewDrdlOptions()
				req.NoError(err)
				req.NoError(opts.Parse(test.args))

				// test that these _are_ set in DrdlSSL and _are not_ set in ConnString()
				req.Equal(test.expectedEnabled, opts.DrdlSSL.Enabled)
				req.Equal(test.expectedSSLCAFile, opts.DrdlSSL.SSLCAFile)
				req.Equal(test.expectedSSLPEMKeyFile, opts.DrdlSSL.SSLPEMKeyFile)
				req.Equal(test.expectedSSLPEMKeyPassword, opts.DrdlSSL.SSLPEMKeyPassword)
				req.Equal(test.expectedSSLCRLFile, opts.DrdlSSL.SSLCRLFile)
				req.Equal(test.expectedSSLAllowInvalidCert, opts.DrdlSSL.SSLAllowInvalidCert)
				req.Equal(test.expectedSSLAllowInvalidHost, opts.DrdlSSL.SSLAllowInvalidHost)
				req.Equal(test.expectedSSLFipsMode, opts.DrdlSSL.SSLFipsMode)
				req.Equal(test.expectedMinimumTLSVersion, opts.DrdlSSL.MinimumTLSVersion)

				cs, err := opts.ConnString()
				req.NoError(err)
				req.False(cs.SSLSet)
				req.Zero(cs.SSL)
				req.False(cs.SSLCaFileSet)
				req.Zero(cs.SSLCaFile)
				req.False(cs.SSLClientCertificateKeyFileSet)
				req.Zero(cs.SSLClientCertificateKeyFile)
				req.False(cs.SSLClientCertificateKeyPasswordSet)
				req.Zero(cs.SSLClientCertificateKeyPassword)
				req.False(cs.SSLInsecureSet)
				req.Zero(cs.SSLInsecure)
			})
		}
	})
}

func testValid(t *testing.T) {
	tests := []struct {
		description                         string
		args                                []string
		expectedUsername                    string
		expectedPassword                    string
		expectedAuthenticationDatabase      string
		expectedAuthenticationMechanism     string
		expectedHosts                       []string
		expectedGSSAPIServiceName           string
		expectedGSSAPIHostName              string
		expectedDB                          string
		expectedCollection                  string
		expectedSSL                         bool
		expectedSSLCAFile                   string
		expectedSSLPEMKeyFile               string
		expectedSSLPEMKeyPassword           string
		expectedSSLCRLFile                  string
		expectedSSLAllowInvalidCertificates bool
		expectedSSLAllowInvalidHostnames    bool
		expectedSSLFIPSMode                 bool
		expectedMinimumTLSVersion           string
	}{
		{
			"options via flags (with auth mechanism properties)",
			[]string{
				"--username", "user",
				"--password", "pass",
				"--authenticationMechanism", "GSSAPI",
				"--host", "localhost",
				"--port", "27000",
				"--gssapiServiceName", "service",
				"--gssapiHostName", "host",
				"--db", "db",
				"--collection", "c",
				"--ssl",
				"--sslCAFile", "ca",
				"--sslPEMKeyFile", "key",
				"--sslPEMKeyPassword", "keypass",
				"--sslCRLFile", "crl",
				"--sslAllowInvalidCertificates",
				"--sslAllowInvalidHostnames",
				"--sslFIPSMode",
				"--minimumTLSVersion", "TLS1_2",
			},
			"user",
			"pass",
			"$external",
			"GSSAPI",
			[]string{"localhost:27000"},
			"service",
			"host",
			"db",
			"c",
			true,
			"ca",
			"key",
			"keypass",
			"crl",
			true,
			true,
			true,
			"TLS1_2",
		},
		{
			"options via flags (without auth mechanism properties)",
			[]string{
				"--username", "user",
				"--password", "pass",
				"--authenticationDatabase", "authDB",
				"--authenticationMechanism", "SCRAM-SHA-1",
				"--host", "localhost",
				"--port", "27000",
				"--gssapiServiceName", "service",
				"--gssapiHostName", "host",
				"--db", "db",
				"--collection", "c",
				"--ssl",
				"--sslCAFile", "ca",
				"--sslPEMKeyFile", "key",
				"--sslPEMKeyPassword", "keypass",
				"--sslCRLFile", "crl",
				"--sslAllowInvalidCertificates",
				"--sslAllowInvalidHostnames",
				"--sslFIPSMode",
				"--minimumTLSVersion", "TLS1_2",
			},
			"user",
			"pass",
			"authDB",
			"SCRAM-SHA-1",
			[]string{"localhost:27000"},
			"",
			"",
			"db",
			"c",
			true,
			"ca",
			"key",
			"keypass",
			"crl",
			true,
			true,
			true,
			"TLS1_2",
		},
		{
			"options via uri",
			[]string{
				"--uri", "mongodb://user:pass@localhost:27000/db?authSource=$external&authMechanism=GSSAPI&authMechanismProperties=SERVICE_NAME:service&ssl=true&sslcertificateauthorityfile=ca&sslclientcertificatekeyfile=key&sslclientcertificatekeypassword=keypass&sslinsecure=true",
				"--gssapiHostName", "host",
				"--collection", "c",
				"--sslCRLFile", "crl",
				"--sslFIPSMode",
				"--minimumTLSVersion", "TLS1_2",
			},
			"user",
			"pass",
			"$external",
			"GSSAPI",
			[]string{"localhost:27000"},
			"service",
			"host",
			"db",
			"c",
			true,
			"ca",
			"key",
			"keypass",
			"crl",
			true,
			true,
			true,
			"TLS1_2",
		},
		{
			"replset seedlist for host",
			[]string{
				"--username", "user",
				"--password", "pass",
				"--authenticationDatabase", "authDB",
				"--authenticationMechanism", "SCRAM-SHA-1",
				"--host", "testReplSet/localhost:27017,localhost:27018,localhost:27019",
				"--gssapiServiceName", "service",
				"--gssapiHostName", "host",
				"--db", "db",
				"--collection", "c",
				"--ssl",
				"--sslCAFile", "ca",
				"--sslPEMKeyFile", "key",
				"--sslPEMKeyPassword", "keypass",
				"--sslCRLFile", "crl",
				"--sslAllowInvalidCertificates",
				"--sslAllowInvalidHostnames",
				"--sslFIPSMode",
				"--minimumTLSVersion", "TLS1_2",
			},
			"user",
			"pass",
			"authDB",
			"SCRAM-SHA-1",
			[]string{"localhost:27017", "localhost:27018", "localhost:27019"},
			"",
			"",
			"db",
			"c",
			true,
			"ca",
			"key",
			"keypass",
			"crl",
			true,
			true,
			true,
			"TLS1_2",
		},
		{
			"replset seedlist for uri",
			[]string{
				"--uri", "mongodb://user:pass@localhost:27017,localhost:27018,localhost:27019/db?replicaSet=testReplSet&authSource=authDB&authMechanism=SCRAM-SHA-1",
				"--collection", "c",
				"--ssl",
				"--sslCAFile", "ca",
				"--sslPEMKeyFile", "key",
				"--sslPEMKeyPassword", "keypass",
				"--sslCRLFile", "crl",
				"--sslAllowInvalidCertificates",
				"--sslAllowInvalidHostnames",
				"--sslFIPSMode",
				"--minimumTLSVersion", "TLS1_2",
			},
			"user",
			"pass",
			"authDB",
			"SCRAM-SHA-1",
			[]string{"localhost:27017", "localhost:27018", "localhost:27019"},
			"",
			"",
			"db",
			"c",
			true,
			"ca",
			"key",
			"keypass",
			"crl",
			true,
			true,
			true,
			"TLS1_2",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req := require.New(t)
			opts, err := mongodrdl.NewDrdlOptions()
			req.NoError(err)
			req.NoError(opts.Parse(test.args))

			cs, err := opts.ConnString()
			req.NoError(err)

			// test values that are stored in the conn string
			req.Equal(test.expectedUsername, cs.Username, "incorrect username")
			req.Equal(test.expectedPassword, cs.Password, "incorrect password")
			req.Equal(test.expectedAuthenticationDatabase, cs.AuthSource, "incorrect auth source")
			req.Equal(test.expectedAuthenticationMechanism, cs.AuthMechanism, "incorrect auth mechanism")

			for i, expectedHost := range test.expectedHosts {
				req.Equal(expectedHost, cs.Hosts[i], "incorrect host #%v", i)
			}
			req.Equal(test.expectedGSSAPIServiceName, cs.AuthMechanismProperties["SERVICE_NAME"], "incorrect GSSAPI service name")
			req.Equal(test.expectedGSSAPIHostName, cs.AuthMechanismProperties["SERVICE_HOST"], "incorrect GSSAPI host name")
			req.Equal(test.expectedDB, cs.Database, "incorrect db (directly in conn string)")
			req.Equal(test.expectedDB, opts.DB(), "incorrect db (via method)") // the database name is also exposed through a method on opts
			req.Equal(test.expectedCollection, opts.Collection(), "incorrect collection")

			// test values that are stored in DrdlSSL
			req.False(cs.SSLSet, "expected SSL to be stored in DrdlSSL but is set in conn string")
			req.Equal(test.expectedSSL, opts.DrdlSSL.Enabled, "incorrect SSL")
			req.False(cs.SSLCaFileSet, "expected SSL CA File to be stored in DrdlSSL but is set in conn string")
			req.Equal(test.expectedSSLCAFile, opts.DrdlSSL.SSLCAFile, "incorrect SSL CA File")
			req.False(cs.SSLClientCertificateKeyFileSet, "expected SSL PEM Key File to be stored in DrdlSSL but is set in conn string")
			req.Equal(test.expectedSSLPEMKeyFile, opts.DrdlSSL.SSLPEMKeyFile, "incorrect SSL PEM Key file")
			req.False(cs.SSLClientCertificateKeyPasswordSet, "expected SSL PEM Key password to be stored in DrdlSSL but is set in conn string")
			req.Equal(test.expectedSSLPEMKeyPassword, opts.DrdlSSL.SSLPEMKeyPassword, "incorrect SSL PEM Key password")
			req.Equal(test.expectedSSLCRLFile, opts.DrdlSSL.SSLCRLFile, "incorrect SSL CRL File")
			req.False(cs.SSLInsecureSet, "expected SSL insecure flags to be stored in DrdlSSL but are set in conn string")
			req.Equal(test.expectedSSLAllowInvalidCertificates, opts.DrdlSSL.SSLAllowInvalidCert, "incorrect SSLAllowInvalidCert")
			req.Equal(test.expectedSSLAllowInvalidHostnames, opts.DrdlSSL.SSLAllowInvalidHost, "incorrect SSLAllowInvalidHost")
			req.Equal(test.expectedSSLFIPSMode, opts.DrdlSSL.SSLFipsMode, "incorrect SSL FIPS mode")
			req.Equal(test.expectedMinimumTLSVersion, opts.DrdlSSL.MinimumTLSVersion, "incorrect minimum TLS version")
		})
	}
}
