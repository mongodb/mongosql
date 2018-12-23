package data

import (
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
)

// Dataset is an interface to be implemented by types that represent data to be
// restored to MongoDB.
type Dataset interface {
	// Restore restores the in-memory data to the MongoDB deployment specified
	// in the provided options.
	Restore(*toolsoptions.ToolOptions) error
}

// DatasetGroup is a dataset that represents the union of multiple
// other datasets.
type DatasetGroup []Dataset

// Restore restores data from all of the datasets that comprise the DatasetGroup
// to the MongoDB deployment specified in the provided options.
func (g DatasetGroup) Restore(opts *toolsoptions.ToolOptions) error {
	for _, dataset := range g {
		err := dataset.Restore(opts)
		if err != nil {
			return err
		}
	}
	return nil
}
