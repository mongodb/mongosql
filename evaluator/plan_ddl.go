package evaluator

import (
	"context"
	"strings"

	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/schema"
)

type ddlCommand struct {
	catalog catalog.Catalog
	dbName  string
}

// DropTableCommand handles a Drop Table command.
type DropTableCommand struct {
	ddlCommand
	tableName string
	ifExists  bool
}

// NewDropTableCommand creates a new DropTableCommand
func NewDropTableCommand(catalog catalog.Catalog, db string, table string, ifExists bool) *DropTableCommand {
	return &DropTableCommand{ddlCommand{catalog, db}, table, ifExists}
}

// Children returns a slice of all the Node children of the Node.
func (DropTableCommand) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (DropTableCommand) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex("DropTableCommand", i, -1)
}

// Execute runs this command.
func (com *DropTableCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	// if ifExists is set, and this table does not exist, we exit early with no error.
	// In other conditions we let the commandHandler handle everything, which can return
	// an error if, for example, the table or database does not exist.
	if com.ifExists {
		// check that database exists.
		database, err := com.catalog.Database(com.dbName)
		if err != nil {
			return nil
		}
		// check that table exists.
		_, err = database.Table(com.tableName)
		if err != nil {
			return nil
		}
	}
	err := cfg.commandHandler.DropTable(ctx, com.dbName, com.tableName)
	if err != nil {
		// If we failed to Drop the Table but this table is named #Tableau... we just ignore the
		// command. This is to support the dropping of temporary tables used by Tableau, and is a
		// work-around.
		if strings.HasPrefix(com.tableName, "#Tableau") {
			return nil
		}
	}
	return err
}

// DropDatabaseCommand handles a DropDatabase Database command.
type DropDatabaseCommand struct {
	ddlCommand
	ifExists bool
}

// NewDropDatabaseCommand creates a new DropDatabaseCommand
func NewDropDatabaseCommand(catalog catalog.Catalog, db string, ifExists bool) *DropDatabaseCommand {
	return &DropDatabaseCommand{ddlCommand{catalog, db}, ifExists}
}

// Children returns a slice of all the Node children of the Node.
func (DropDatabaseCommand) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (DropDatabaseCommand) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex("DropDatabaseCommand", i, -1)
}

// Execute runs this command.
func (com *DropDatabaseCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	// if ifExists is set, and this database does not exist, we exit early with no error.
	// In other conditions we let the commandHandler handle everything, which can return
	// an error if, for example, the database does not exist.
	if com.ifExists {
		// check that database exists.
		_, err := com.catalog.Database(com.dbName)
		if err != nil {
			return nil
		}
	}
	err := cfg.commandHandler.DropDatabase(ctx, com.dbName)
	if err != nil {
		return err
	}
	// If we are dropping the current database, we need to unset the current database.
	if strutil.CaseInsensitiveEquals(com.dbName, cfg.dbName) {
		return cfg.commandHandler.UnsetDatabase()
	}
	return nil
}

// CreateDatabaseCommand handles a CreateDatabase Database command.
type CreateDatabaseCommand struct {
	ddlCommand
	ifNotExists bool
}

// NewCreateDatabaseCommand creates a new CreateDatabaseCommand
func NewCreateDatabaseCommand(catalog catalog.Catalog, db string, ifNotExists bool) *CreateDatabaseCommand {
	return &CreateDatabaseCommand{ddlCommand{catalog, db}, ifNotExists}
}

// Children returns a slice of all the Node children of the Node.
func (CreateDatabaseCommand) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (CreateDatabaseCommand) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex("CreateDatabaseCommand", i, -1)
}

// Execute runs this command.
func (com *CreateDatabaseCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	// if ifNotExists is set, and this database exists, we exit early with no error.
	// In other conditions we let the commandHandler handle everything, which can return
	// an error if, for example, the database exists.
	if com.ifNotExists {
		// check that database exists.
		_, err := com.catalog.Database(com.dbName)
		if err == nil {
			return nil
		}
	}
	return cfg.commandHandler.CreateDatabase(ctx, com.dbName)
}

// CreateTableCommand handles a CreateTable Table command.
type CreateTableCommand struct {
	ddlCommand
	table       *schema.Table
	ifNotExists bool
}

// NewCreateTableCommand creates a new CreateTableCommand
func NewCreateTableCommand(catalog catalog.Catalog, dbName string, table *schema.Table, ifNotExists bool) *CreateTableCommand {
	return &CreateTableCommand{ddlCommand{catalog, dbName}, table, ifNotExists}
}

// Children returns a slice of all the Node children of the Node.
func (CreateTableCommand) Children() []Node {
	return []Node{}
}

// ReplaceChild replaces the i'th child of the receiver Node with the Node n.
func (CreateTableCommand) ReplaceChild(i int, n Node) {
	panicWithInvalidIndex("CreateTableCommand", i, -1)
}

// Execute runs this command.
func (com *CreateTableCommand) Execute(ctx context.Context, cfg *ExecutionConfig, st *ExecutionState) error {
	// if ifNotExists is set, and this table exists, we exit early with no error.
	// In other conditions we let the commandHandler handle everything, which can return
	// an error if, for example, the table or database does not exist.
	if com.ifNotExists {
		// check that database exists. If it does not, return the err.
		database, err := com.catalog.Database(com.dbName)
		if err != nil {
			return err
		}
		// check that does not table exist. If it does, return nil.
		_, err = database.Table(com.table.SQLName())
		if err == nil {
			return nil
		}
	}
	return cfg.commandHandler.CreateTable(ctx, com.dbName, com.table)
}
