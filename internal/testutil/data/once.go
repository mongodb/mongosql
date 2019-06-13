package data

import (
	toolsoptions "github.com/mongodb/mongo-tools-common/options"
)

// OnceDataset is a dataset that should only be restored once. After the first
// time this dataset has been restored, future calls to Restore() will not do
// anything.
type OnceDataset struct {
	data     Dataset
	restored bool
}

// Once returns a new OnceDataset that wraps the provided Dataset.
func Once(data Dataset) *OnceDataset {
	return &OnceDataset{
		data:     data,
		restored: false,
	}
}

// Restore restores the OnceDataset's wrapped dataset if it has not already done
// so in the past. After the first time this dataset has been restored, future
// calls to Restore will not do anything.
func (o *OnceDataset) Restore(opts *toolsoptions.ToolOptions) error {
	var err error
	if !o.restored {
		err = o.data.Restore(opts)
		if err == nil {
			o.restored = true
		}
	}
	return err
}

// Reset guarantees that the next call to Restore will cause the wrapped dataset
// to be restored, regardless of whether previous calls to Restore were made.
func (o *OnceDataset) Reset() {
	o.restored = false
}
