package data

import (
	"io/ioutil"
	"path/filepath"

	yaml "github.com/10gen/candiedyaml"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
)

// A YMLDataset is a dataset defined in a yml file. The yml file should be
// formatted to Unmarshal into an InMemoryDataset.
type YMLDataset struct {
	File string
}

// NewYMLDataset returns a new YMLDataset with the provided file.
func NewYMLDataset(file string) YMLDataset {
	return YMLDataset{
		File: file,
	}
}

// Restore restores the data specified in the yml file to the MongoDB
// deployment specified in the provided options.
func (y YMLDataset) Restore(opts *toolsoptions.ToolOptions) error {
	fileName := filepath.Join("testdata", "resources", "data", y.File)

	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	inmems := []*InMemoryDataset{}
	err = yaml.Unmarshal(file, &inmems)
	if err != nil {
		return err
	}

	for _, set := range inmems {
		err = set.Restore(opts)
		if err != nil {
			return err
		}
	}
	return nil
}
