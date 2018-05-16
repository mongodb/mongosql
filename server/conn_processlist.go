package server

import (
	"sync"
	"time"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

// Constants representing special field values
const (
	FieldNull  string = "NULL"
	FieldBlank string = ""
)

// MySQL thread command values
// full list: https://dev.mysql.com/doc/refman/5.7/en/thread-commands.html
const (
	CommandQuery string = "Query"
	CommandSleep string = "Sleep"
)

// MySQL process states
const (
	StateStarting string = "starting"
)

// Process is a struct for wrapping the necessary data to control a process.
type Process struct {
	id        uint32
	user      string
	host      string
	db        string
	command   string
	startTime time.Time
	state     string
	info      string

	lock sync.RWMutex
}

// NewProcess is the public constructor for Process structs.
func NewProcess(id uint32) *Process {
	return &Process{
		id:        id,
		db:        FieldNull,
		startTime: time.Now(),
	}
}

// ComputeUptime computes the runtime for a running process.
func (proc *Process) ComputeUptime() uint64 {
	return uint64(time.Since(proc.startTime).Nanoseconds() / 1e9)
}

// SetUser sets the user for a process.
func (proc *Process) SetUser(user string) {
	proc.lock.Lock()
	proc.user = user
	proc.lock.Unlock()
}

// SetHost sets the host for a process.
func (proc *Process) SetHost(host string) {
	proc.lock.Lock()
	proc.host = host
	proc.lock.Unlock()
}

// SetDB sets the DB for a process.
func (proc *Process) SetDB(db string) {
	proc.lock.Lock()
	proc.db = db
	proc.lock.Unlock()
}

// UpdateProcess updates the command for a process.
func (proc *Process) UpdateProcess(command, info string) {
	proc.lock.Lock()

	if command == CommandQuery {
		proc.command = CommandQuery
		proc.state = StateStarting
		proc.info = info
	} else {
		proc.command = CommandSleep
		proc.state = FieldBlank
		proc.info = FieldNull
	}
	proc.startTime = time.Now()

	proc.lock.Unlock()
}

// UpdateWithProcessListTable adds the PROCESSLIST table to the catalog after the latter has
// been created. This function resides here due to dependency constraints.
func (c *conn) UpdateWithProcessListTable(d *catalog.Database) error {
	t := catalog.NewDynamicTable("PROCESSLIST", catalog.SystemView, func() []*catalog.DataRow {
		var rows []*catalog.DataRow

		s := c.server

		// Grab a snapshot of the active processes.
		s.activeConnectionsMx.RLock()
		processList := make([]*Process, len(s.activeConnections))
		i := 0
		for _, currConn := range s.activeConnections {
			processList[i] = currConn.process
			i++
		}
		s.activeConnectionsMx.RUnlock()

		for _, p := range processList {
			// If this is the current users process we can show it. If it is
			// not, we need to check that either security is disabled or the
			// user has the `inprog` privilege.
			p.lock.RLock()
			if p.user == c.user || !c.server.cfg.Security.Enabled ||
				c.variables.MongoDBInfo.IsAllowedCluster(mongodb.InprogPrivilege) {
				rows = append(rows, catalog.NewDataRow(p.id, p.user,
					p.host, p.db, p.command,
					p.ComputeUptime(), p.state, p.info))
			}
			p.lock.RUnlock()
		}
		return rows
	})

	t.AddColumns(
		"ID", string(schema.SQLInt),
		"USER", string(schema.SQLVarchar),
		"HOST", string(schema.SQLVarchar),
		"DB", string(schema.SQLVarchar),
		"COMMAND", string(schema.SQLVarchar),
		"TIME", string(schema.SQLInt64),
		"STATE", string(schema.SQLVarchar),
		"INFO", string(schema.SQLVarchar),
	)

	return d.AddTable(t)

}
