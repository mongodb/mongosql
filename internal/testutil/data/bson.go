package data

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/10gen/sqlproxy/internal/testutil/mongodb"
	toolsdb "github.com/mongodb/mongo-tools/common/db"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongorestore"
)

// BSONDataset is a dataset that is stored as a .bson.gz file and restored with
// mongorestore. It is defined by the url from which it can be downloaded and
// the minimum MongoDB version to which its data can be restored.
type BSONDataset struct {
	URL        string
	MinVersion string
}

// NewBSONDataset returns a BSONDataset whose URL is constructed from the
// provided name, with a MinVersion of 3.2.
func NewBSONDataset(name string) BSONDataset {
	return BSONDataset{
		URL: fmt.Sprintf("http://noexpire.s3.amazonaws.com/sqlproxy/data/%s.bson.archive.gz",
			name),
		MinVersion: "3.2",
	}
}

// NewBSONDataset34 returns a BSONDataset whose URL is constructed from the
// provided name, with a MinVersion of 3.4.
func NewBSONDataset34(name string) BSONDataset {
	data := NewBSONDataset(name)
	data.MinVersion = "3.4"
	return data
}

// Restore restores the bson data to the MongoDB deployment specified
// in the provided options.
func (b BSONDataset) Restore(opts *toolsoptions.ToolOptions) error {
	fmt.Printf(">> Downloading %s data...\n", b.fileName())
	err := b.download(false)
	if err != nil {
		return err
	}
	fmt.Printf(">> Restoring %s data...\n", b.fileName())
	return b.restoreFromFile(opts)
}

func (b BSONDataset) fileName() string {
	parts := strings.Split(b.URL, "/")
	baseName := parts[len(parts)-1]
	fileName := filepath.Join("testdata", "resources", "data", baseName)
	return fileName
}

func (b BSONDataset) download(clobber bool) error {
	fileName := b.fileName()

	info, err := os.Stat(fileName)
	missing := os.IsNotExist(err)
	if err != nil && !missing {
		return err
	}

	if !missing && !clobber && !info.IsDir() {
		return nil
	}

	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	resp, err := http.Get(b.URL)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (b BSONDataset) restoreFromFile(opts *toolsoptions.ToolOptions) error {
	opts.DB = ""
	opts.Collection = ""
	sp, err := toolsdb.NewSessionProvider(*opts)
	if err != nil {
		return err
	}

	defer sp.Close()

	ok, err := mongodb.VersionAtLeast(sp, b.MinVersion)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	restorer := mongorestore.MongoRestore{
		ToolOptions:  opts,
		InputOptions: &mongorestore.InputOptions{Gzip: true, Archive: b.fileName()},
		OutputOptions: &mongorestore.OutputOptions{
			Drop:                   true,
			StopOnError:            true,
			NumParallelCollections: 1,
			NumInsertionWorkers:    10,
			MaintainInsertionOrder: false,
		},
		NSOptions:       &mongorestore.NSOptions{},
		SessionProvider: sp,
	}

	return restorer.Restore().Err
}
