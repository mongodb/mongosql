package evaluator

import (
	"context"
	"fmt"

	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/schema"
)

// DropTableCommand handles a Drop Table command.
type DropTableCommand struct {
	dbName    option.String
	tableName string
	ifExists  bool
}

// NewDropTableCommand creates a new DropTableCommand
func NewDropTableCommand(db option.String, table string, ifExists bool) *DropTableCommand {
	return &DropTableCommand{db, table, ifExists}
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
	return cfg.commandHandler.Drop(com.tableName)
}

// DropDatabaseCommand handles a DropDatabase Database command.
type DropDatabaseCommand struct {
	dbName   string
	ifExists bool
}

// NewDropDatabaseCommand creates a new DropDatabaseCommand
func NewDropDatabaseCommand(db string, ifExists bool) *DropDatabaseCommand {
	return &DropDatabaseCommand{db, ifExists}
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
	return fmt.Errorf("%T is unimplemented", com)
}

// CreateDatabaseCommand handles a CreateDatabase Database command.
type CreateDatabaseCommand struct {
	dbName      string
	ifNotExists bool
}

// NewCreateDatabaseCommand creates a new CreateDatabaseCommand
func NewCreateDatabaseCommand(db string, ifNotExists bool) *CreateDatabaseCommand {
	return &CreateDatabaseCommand{db, ifNotExists}
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
	return fmt.Errorf("%T is unimplemented", com)
}

// CreateTableCommand handles a CreateTable Table command.
type CreateTableCommand struct {
	dbName      option.String
	table       *schema.Table
	ifNotExists bool
}

// NewCreateTableCommand creates a new CreateTableCommand
func NewCreateTableCommand(dbName option.String, table *schema.Table, ifNotExists bool) *CreateTableCommand {
	return &CreateTableCommand{dbName, table, ifNotExists}
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
	return fmt.Errorf("%T is unimplemented", com)
}
