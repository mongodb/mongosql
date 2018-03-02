package dsync

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
)

const (
	fieldLockID         = "_id"
	fieldProcessName    = "processName"
	fieldExpirationTime = "expirationTime"
)

var errLockHeldByAnotherProcess = errors.New("lock held by another process")

// NewDMutex creates a distributed mutual exclusion runningLock.
func NewDMutex(cfg DMutexConfig) *DMutex {
	return &DMutex{
		cfg:  cfg,
		done: make(chan struct{}),
	}
}

// DMutex is a distributed mutual exclusion runningLock.
type DMutex struct {
	cfg  DMutexConfig
	done chan struct{}
	// running tracks whether we have acquired the lock and
	// are heart-beating to ensure that we keep the lock. This
	// allows the lock to be re-entrant.
	running     bool
	runningLock sync.Mutex
}

// DMutexConfig holds configuration for a DMutex.
type DMutexConfig struct {
	Name              string
	CollectionName    string
	DatabaseName      string
	HeartbeatInterval time.Duration
	Logger            *log.Logger
	ProcessName       string
	SessionProvider   *mongodb.SessionProvider
	Timeout           time.Duration
}

// Lock coordinates with other processes to acquire a lock. The request can be
// cancelled via the context and will return nil when successful. It will also
// spin up a goroutine to continually refresh the lock until Unlock is called.
// DMutex is re-entrant on a per instance basis. It is possible and encouraged
// to call Lock immediately before utilizing a resource to ensure the lock is
// still acquired.
func (d *DMutex) Lock(ctx context.Context) error {
	err := d.tryLock(ctx)
	if err != nil {
		d.cfg.Logger.Warnf(log.Dev, "failed to acquire lock %q: %v", d.cfg.Name, err)
		return err
	}

	d.cfg.Logger.Debugf(log.Dev, "acquired lock %q", d.cfg.Name)

	localUnlock := func() {
		d.runningLock.Lock()
		d.running = false
		d.runningLock.Unlock()
	}

	d.runningLock.Lock()
	defer d.runningLock.Unlock()

	if d.cfg.HeartbeatInterval > 0 && !d.running {
		d.running = true
		util.PanicSafeGo(func() {
			defer localUnlock()
			util.RetryWithDelay(d.done, d.cfg.HeartbeatInterval, false, func() bool {
				err := d.tryLock(context.Background())
				if err != nil {
					d.cfg.Logger.Warnf(log.Dev, "failed to refresh lock %q: %v", d.cfg.Name, err)
					// we failed to refresh the runningLock, so we
					// are aborting our refresh loop
					return true
				}

				d.cfg.Logger.Debugf(log.Dev, "refreshed lock %q", d.cfg.Name)
				return false
			})
		}, func(err interface{}) {
			localUnlock()
		})
	}

	return nil
}

// Unlock releases the lock.
func (d *DMutex) Unlock(ctx context.Context) (err error) {
	d.runningLock.Lock()
	defer d.runningLock.Unlock()
	if d.running {
		d.done <- struct{}{}
	}

	var session *mongodb.Session
	session, err = d.cfg.SessionProvider.AdminSession(ctx)
	if err != nil {
		return err
	}

	defer util.CheckDeferredFunc(session.Close, &err)

	cmd := bson.D{
		{Name: "findAndModify", Value: d.cfg.CollectionName},
		{Name: "query", Value: bson.M{
			fieldLockID:      d.cfg.Name,
			fieldProcessName: d.cfg.ProcessName,
		}},
		{Name: "remove", Value: true},
		{Name: "writeConcern", Value: bson.M{"w": "majority"}},
	}

	result := struct{}{}
	err = session.Run(d.cfg.DatabaseName, cmd, &result)
	if err != nil {
		d.cfg.Logger.Warnf(log.Dev, "failed to release lock %q: %v", d.cfg.Name, err)
		return err
	}

	d.cfg.Logger.Debugf(log.Dev, "released lock %q", d.cfg.Name)
	return nil
}

func (d *DMutex) tryLock(ctx context.Context) (err error) {
	var session *mongodb.Session
	session, err = d.cfg.SessionProvider.AdminSession(ctx)
	if err != nil {
		return err
	}

	defer util.CheckDeferredFunc(session.Close, &err)

	now := time.Now().UTC()
	cmd := bson.D{
		{Name: "findAndModify", Value: d.cfg.CollectionName},
		{Name: "query", Value: bson.M{
			fieldLockID: d.cfg.Name,
			"$or": []bson.M{
				{fieldProcessName: d.cfg.ProcessName},
				{fieldExpirationTime: bson.M{
					"$lte": now.Add(30 * time.Second)}, // give a cushion to account for clock skew
				},
			},
		}},
		{Name: "update", Value: bson.M{
			"$set": bson.M{
				fieldProcessName:    d.cfg.ProcessName,
				fieldExpirationTime: now.Add(d.cfg.Timeout),
			},
		}},
		{Name: "upsert", Value: true},
		{Name: "new", Value: true},
		{Name: "writeConcern", Value: bson.M{"w": "majority"}},
	}

	result := struct {
		Value *struct{}
	}{}

	err = session.Run(d.cfg.DatabaseName, cmd, &result)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error") {
			// this is to be expected when someone else holds the lock...
			return d.createLockHeldByAnotherProcessError(session)
		}
		return err
	}

	if result.Value == nil {
		return d.createLockHeldByAnotherProcessError(session)
	}

	return nil
}

// createLockHelpByAnotherProcessError is called when we know the lock is held
// by another process. It will attempt to discover extra information about the
// lock itself and include that in the error. If it fails to discover extra
// information, it will return a generic error.
func (d *DMutex) createLockHeldByAnotherProcessError(session *mongodb.Session) (err error) {
	pipeline := []bson.D{
		{{Name: "$match", Value: bson.D{{Name: "_id", Value: d.cfg.Name}}}},
		{{Name: "$limit", Value: 1}},
	}

	var cursor mongodb.Cursor
	cursor, err = session.Aggregate(d.cfg.DatabaseName, d.cfg.CollectionName, pipeline)
	if err != nil {
		return errLockHeldByAnotherProcess
	}
	defer func() {
		if cerr := cursor.Close(session.Context()); cerr != nil && err == nil {
			err = cerr
		}
	}()

	result := struct {
		ProcessName string `bson:"processName"`
	}{}

	if cursor.Next(session.Context(), &result) {
		return fmt.Errorf("lock held by process '%s'", result.ProcessName)
	}

	return errLockHeldByAnotherProcess
}
