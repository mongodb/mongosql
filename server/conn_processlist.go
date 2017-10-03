package server

import (
	"sync"
	"time"

	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/schema"
)

const (
	FieldNull  string = "NULL"
	FieldBlank string = ""

	// full list as per MySQL: https://dev.mysql.com/doc/refman/5.7/en/thread-commands.html
	CommandQuery string = "Query"
	CommandSleep string = "Sleep"

	StateStarting string = "starting"
)

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

func NewProcess(id uint32) *Process {
	return &Process{
		id:        id,
		db:        FieldNull,
		startTime: time.Now(),
	}
}

func (proc *Process) ComputeUptime() uint64 {
	return uint64(time.Now().Sub(proc.startTime).Nanoseconds() / 1e9)
}

func (proc *Process) SetUser(user string) {
	proc.lock.Lock()
	proc.user = user
	proc.lock.Unlock()
}

func (proc *Process) SetHost(host string) {
	proc.lock.Lock()
	proc.host = host
	proc.lock.Unlock()
}

func (proc *Process) SetDB(db string) {
	proc.lock.Lock()
	proc.db = db
	proc.lock.Unlock()
}

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
// been created. This function resides here due to dependency constraints
func (c *conn) UpdateWithProcessListTable(d *catalog.Database) error {

	t := catalog.NewDynamicTable("PROCESSLIST", catalog.SystemView, func() []*catalog.DataRow {
		var rows []*catalog.DataRow

		s := c.server

		s.activeConnectionsMx.RLock()
		for _, currConn := range s.activeConnections {
			p := currConn.process
			p.lock.RLock()
			rows = append(rows, catalog.NewDataRow(p.id, p.user, p.host, p.db, p.command, p.ComputeUptime(), p.state, p.info))
			p.lock.RUnlock()
		}
		s.activeConnectionsMx.RUnlock()

		return rows
	})

	t.AddColumn("ID", schema.SQLInt)
	t.AddColumn("USER", schema.SQLVarchar)
	t.AddColumn("HOST", schema.SQLVarchar)
	t.AddColumn("DB", schema.SQLVarchar)
	t.AddColumn("COMMAND", schema.SQLVarchar)
	t.AddColumn("TIME", schema.SQLInt64)
	t.AddColumn("STATE", schema.SQLVarchar)
	t.AddColumn("INFO", schema.SQLVarchar)

	return d.AddTable(t)

}
