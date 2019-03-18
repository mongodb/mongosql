package mongodrdl_test

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/mongodrdl"
	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	t.Run("host_and_port", testHostAndPort)
	t.Run("invalid_ssl", testInvalidSSL)

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

func testHostAndPort(t *testing.T) {

	tests := []struct {
		description  string
		args         []string
		expectedHost string
	}{{
		"host and port",
		[]string{
			"--host", "localhost",
			"--port", "6999",
		},
		"localhost:6999",
	}, {
		"host with port and port",
		[]string{
			"--host", "localhost:34325452",
			"--port", "6999",
		},
		"localhost:6999",
	}, {
		"host with port",
		[]string{
			"--host", "localhost:6999",
		},
		"localhost:6999",
	}, {
		"host and port and misc ssl opts",
		[]string{
			"--host", "localhost",
			"--port", "6999",
			"--ssl",
			"--sslCAFile", "hello",
		},
		"localhost:6999",
	}}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req := require.New(t)
			opts, err := mongodrdl.NewDrdlOptions()
			req.NoError(err)
			req.NoError(opts.Parse(test.args))
			req.Equal(test.expectedHost, opts.DrdlConnection.Host)
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
