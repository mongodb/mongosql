package data

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/10gen/sqlproxy/internal/testutil/flags"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
)

const (
	// TestClientSSL is the name of an environment variable that can be set to
	// indicate that SSL is enabled for connection to mongodb.
	TestClientSSL = "GO_TEST_CLIENT_SSL"
)

// ResamplingDataset is a wrapper around a Dataset that will issue a FLUSH
// SAMPLE command to mongosqld after restoring the wrapped Dataset.
type ResamplingDataset struct {
	data Dataset
}

// Resample wraps the provided Dataset in a ResamplingDataset.
func Resample(data Dataset) ResamplingDataset {
	return ResamplingDataset{
		data: data,
	}
}

// Restore restores the wrapped Dataset, then issues a FLUSH SAMPLE command to
// mongosqld so that it resamples the data that was just restored.
func (r ResamplingDataset) Restore(opts *toolsoptions.ToolOptions) error {
	err := r.data.Restore(opts)
	if err != nil {
		return err
	}

	return FlushSample()
}

// FlushSample issues a flush sample command to mongosqld to resample restored data.
func FlushSample() error {
	connString := fmt.Sprintf(
		"bob:pwd123@tcp(%v)/information_schema?allowNativePasswords=1&allowCleartextPasswords=1",
		*flags.DbAddr,
	)
	// For tests where SSL is enabled, append the TLS parameter to the
	// connect string.
	if len(os.Getenv(TestClientSSL)) > 0 {
		connString += "&tls=skip-verify"
	}

	db, err := sql.Open("mysql", connString)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("flush sample")
	return err
}
