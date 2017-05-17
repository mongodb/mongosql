package variable

import (
	"sync/atomic"
	"time"

	"github.com/10gen/sqlproxy/schema"
)

const (
	// Status Variable Names
	BytesReceived    Name = "Bytes_received"
	BytesSent             = "Bytes_sent"
	Connections           = "Connections"
	Queries               = "Queries"
	ThreadsConnected      = "Threads_connected"
	ThreadsCreated        = "Threads_created"
	Uptime                = "Uptime"
)

func init() {
	//  Status Variable Definitions
	definitions[BytesReceived] = &definition{
		Name:             BytesReceived,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLUint64,
		GetValue: func(c *Container) interface{} {
			return atomic.LoadUint64(c.BytesReceived)
		},
	}

	definitions[BytesSent] = &definition{
		Name:             BytesSent,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLUint64,
		GetValue: func(c *Container) interface{} {
			return atomic.LoadUint64(c.BytesSent)
		},
	}

	definitions[Connections] = &definition{
		Name:             Connections,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLUint64,
		GetValue: func(c *Container) interface{} {
			return atomic.LoadUint32(c.Connections)
		},
	}

	definitions[Queries] = &definition{
		Name:             Queries,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLUint64,
		GetValue: func(c *Container) interface{} {
			return atomic.LoadUint64(c.Queries)
		},
	}

	definitions[ThreadsConnected] = &definition{
		Name:             ThreadsConnected,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLUint64,
		GetValue: func(c *Container) interface{} {
			return atomic.LoadUint32(c.ThreadsConnected)
		},
	}

	definitions[ThreadsCreated] = &definition{
		Name:             ThreadsCreated,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLUint64,
		GetValue: func(c *Container) interface{} {
			return atomic.LoadUint32(c.ThreadsConnected)
		},
	}

	definitions[Uptime] = &definition{
		Name:             Uptime,
		Kind:             StatusKind,
		AllowedSetScopes: Scope(0), // not allowed to be set
		SQLType:          schema.SQLUint64,
		GetValue: func(c *Container) interface{} {
			return uint64(time.Now().Sub(*c.StartTime).Nanoseconds() / 1e9)
		},
	}
}
