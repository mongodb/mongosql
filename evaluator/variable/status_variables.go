package variable

import (
	"sync/atomic"
	"time"

	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
)

// Status Variable Names
const (
	BytesReceived    Name = "Bytes_received"
	BytesSent        Name = "Bytes_sent"
	Connections      Name = "Connections"
	Queries          Name = "Queries"
	ThreadsConnected Name = "Threads_connected"
	ThreadsCreated   Name = "Threads_created"
	Uptime           Name = "Uptime"
	MemoryAllocated  Name = "Memory_allocated"
)

func init() {
	//  Status Variable Definitions
	definitions[BytesReceived] = &definition{
		Name:             BytesReceived,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		EvalType:         types.EvalUint64,
		GetValue: func(c *Container) values.SQLValue {
			return values.NewSQLUint64(values.VariableSQLValueKind, atomic.LoadUint64(c.BytesReceived))
		},
	}

	definitions[BytesSent] = &definition{
		Name:             BytesSent,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		EvalType:         types.EvalUint64,
		GetValue: func(c *Container) values.SQLValue {
			return values.NewSQLUint64(values.VariableSQLValueKind, atomic.LoadUint64(c.BytesSent))
		},
	}

	definitions[Connections] = &definition{
		Name:             Connections,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		EvalType:         types.EvalUint64,
		GetValue: func(c *Container) values.SQLValue {
			return values.NewSQLUint64(values.VariableSQLValueKind, uint64(atomic.LoadUint32(c.Connections)))
		},
	}

	definitions[MemoryAllocated] = &definition{
		Name:             MemoryAllocated,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		EvalType:         types.EvalUint64,
		GetValue: func(c *Container) values.SQLValue {
			return values.NewSQLUint64(values.VariableSQLValueKind, c.AllocatedMemory())
		},
	}

	definitions[Queries] = &definition{
		Name:             Queries,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		EvalType:         types.EvalUint64,
		GetValue: func(c *Container) values.SQLValue {
			return values.NewSQLUint64(values.VariableSQLValueKind, atomic.LoadUint64(c.Queries))
		},
	}

	definitions[ThreadsConnected] = &definition{
		Name:             ThreadsConnected,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		EvalType:         types.EvalUint64,
		GetValue: func(c *Container) values.SQLValue {
			return values.NewSQLUint64(values.VariableSQLValueKind, uint64(atomic.LoadUint32(c.ThreadsConnected)))
		},
	}

	definitions[ThreadsCreated] = &definition{
		Name:             ThreadsCreated,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		EvalType:         types.EvalUint64,
		GetValue: func(c *Container) values.SQLValue {
			return values.NewSQLInt64(values.VariableSQLValueKind, int64(atomic.LoadUint32(c.ThreadsConnected)))
		},
	}

	definitions[Uptime] = &definition{
		Name:             Uptime,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		EvalType:         types.EvalUint64,
		GetValue: func(c *Container) values.SQLValue {
			return values.NewSQLUint64(values.VariableSQLValueKind, uint64(time.Since(c.StartTime).Nanoseconds()/1e9))
		},
	}
}
