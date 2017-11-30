package data

import (
	"database/sql"
	"fmt"

	"github.com/10gen/sqlproxy/internal/testutils/flags"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
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

	return flushSample()
}

func flushSample() error {
	connString := fmt.Sprintf("root@tcp(%v)/information_schema?allowNativePasswords=1", *flags.DbAddr)

	db, err := sql.Open("mysql", connString)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Query("flush sample")
	return err
}
