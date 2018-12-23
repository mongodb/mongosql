package data

import (
	"database/sql"
	"fmt"

	"github.com/10gen/sqlproxy/internal/testutil/flags"
	toolsoptions "github.com/mongodb/mongo-tools/common/options"
)

// AlteredDataset is a dataset to which we need to apply additional schema
// alterations after mapping the sampled schema to a relational schema.
type AlteredDataset struct {
	data        Dataset
	alterDB     string
	alterations []string
}

// WithAlterations returns an AlteredDataset wrapping the provided dataset
// with the specified database and alteration commands.
func WithAlterations(data Dataset, db string, alterations ...string) AlteredDataset {
	return AlteredDataset{
		data:        data,
		alterDB:     db,
		alterations: alterations,
	}
}

// Restore restores the wrapped dataset to MongoDB, then connects to mongosqld
// using the alterDB, and executes the SQL commands in alterations to modify
// the schema.
func (a AlteredDataset) Restore(opts *toolsoptions.ToolOptions) error {
	err := a.data.Restore(opts)
	if err != nil {
		return err
	}

	for _, alt := range a.alterations {
		err = alter(a.alterDB, alt)
		if err != nil {
			return err
		}
	}

	return nil
}

func alter(dbName, cmd string) error {
	connString := fmt.Sprintf("root@tcp(%v)/%s?allowNativePasswords=1", *flags.DbAddr, dbName)

	db, err := sql.Open("mysql", connString)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec(cmd)
	return err
}
