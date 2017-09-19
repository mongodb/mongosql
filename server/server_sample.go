package server

import (
	"context"
	"fmt"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/dsync"
	"github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

func sampleLogger() *log.Logger {
	return log.NewComponentLogger(
		fmt.Sprintf("%-10v [schemaDiscovery]", "SAMPLE"),
		log.GlobalLogger(),
	)
}

// getSchema attempts to update the server's schema from the sampleSource.
// Regardless of whether that succeeds, getSchema returns the server's schema.
func (s *Server) getSchema() *schema.Schema {
	lgr := sampleLogger()
	var newSchema *schema.Schema

	session, err := s.sessionProvider.AdminSession(s.lifetimeCtx)
	if err == nil {
		defer session.Close()
		newSchema, err = sample.ReadSchema(&s.cfg.Schema.Sample, session, lgr)
	}

	if err != nil {
		lgr.Logf(log.DebugHigh, "Could not fetch most recent schema: %v", err)
	}

	s.schemaLock.Lock()
	if newSchema != nil {
		s.schema = newSchema
	}
	localSchema := s.schema
	s.schemaLock.Unlock()

	return localSchema
}

// runSampler should be called in a goroutine. It is effectively a state machine with the following states:
// 1. Read the current schema. If a current schema doesn't exist, sample to get a schema.
// 2. If the mode is readonly, we are done.
// 3. Create a distributed lock
// 4. If we had to sample, write the initial schema back to the server.
// 		* if this fails, continue to do this until successful, or until another mongosqld has written
//		  a schema.
// 5. If we are a write once server, then we are done.
// 6. Sample every (configured amount of time) and update the stored schema.
func (s *Server) runSampler(opts *config.SchemaSampleOptions) {
	lgr := sampleLogger()

	var sampleRecord *sample.Record
	var err error

	// 1. All mongosqld's will attempt read an existing schema from the server and sample if one
	// does not exist. When sampling occurred and was successful, the sample record will be returned.
	// Until this completes successfully, we cannot move on.
	util.RetryWithDelay(s.lifetimeCtx.Done(), 5*time.Second, true, func() bool {
		lgr.Logf(log.Info, "initializing schema")
		sampleRecord, err = s.initializeSchema(opts, lgr)
		if err == nil {
			return true
		}

		lgr.Errf(log.Always, "unable to initialize schema: %v", err)
		return false
	})

	// 2. readonly = done
	if s.cfg.Schema.Sample.Mode == "read" {
		return
	}

	// 3. Since we are a writer, we need to create and use a mutex
	// for any write operations.
	mtx := dsync.NewDMutex(dsync.DMutexConfig{
		Name:            "mongosqld-schema",
		DatabaseName:    s.cfg.Schema.Sample.Source,
		CollectionName:  "mongosqld.lock",
		Logger:          lgr,
		ProcessName:     s.processName,
		SessionProvider: s.sessionProvider,
		// Expiration time will be 5 minutes after the last refresh.
		// Every 30 seconds, we'll refresh the lock.
		HeartbeatInterval: 30 * time.Second,
		Timeout:           5 * time.Minute,
	})

	// use a different context here because if s.lifetimeCtx is done (which is what likely prompted the exit of this function
	// and hence the invocation of the defer statement),  we need a different context or else the unlock wouldn't take place.
	defer mtx.Unlock(context.Background())

	// 4. If we have a sample, it means that we didn't read a schema from the server. Therefore, we need to
	// persist this back to the server or, if we fail to do that, read a schema that may show up in the future.
	if sampleRecord != nil {
		if len(sampleRecord.Namespaces) != 0 {
			lgr.Logf(log.Info, "writing sampled schema")
		}
		err := mtx.Lock(s.lifetimeCtx)
		if err == nil {
			// try to do this once initially... if it doesn't work, we'll start looping
			var session *mongodb.Session
			session, err = s.sessionProvider.AdminSession(s.lifetimeCtx)
			if err == nil {
				err = sample.InsertSampleRecord(sampleRecord, session, lgr)
				session.Close()
			}
		}

		if err != nil {
			lgr.Errf(log.DebugLow, "unable to persist initial schema: %v", err)

			util.RetryWithDelay(s.lifetimeCtx.Done(), 1*time.Minute, false, func() bool {
				err := s.writeInitialSample(mtx, lgr, sampleRecord)
				if err != nil {
					lgr.Errf(log.DebugLow, "unable to persist initial schema: %v", err)
					return false
				}

				return true
			})
		}
	}

	// 5. write once = done
	if s.cfg.Schema.Sample.WriteIntervalSecs <= 0 {
		return
	}

	// 6. Re-sample every writeIntervalSecs and persist the schema
	util.RepeatWithDelay(s.lifetimeCtx.Done(), time.Duration(s.cfg.Schema.Sample.WriteIntervalSecs)*time.Second, false, func() {
		lgr.Logf(log.Info, "re-sampling schema")
		err := s.resampleSchema(mtx, lgr, sampleRecord)
		if err != nil {
			lgr.Errf(log.Always, "failed re-sampling schema: %v", err)
		}
	})
}

func (s *Server) initializeSchema(opts *config.SchemaSampleOptions, lgr *log.Logger) (*sample.Record, error) {
	session, err := s.sessionProvider.AdminSession(s.lifetimeCtx)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	newSchema, err := sample.ReadSchema(opts, session, lgr)
	if err != nil {
		return nil, err
	}

	var sampleRecord *sample.Record
	if newSchema == nil {
		lgr.Logf(log.Info, "stored schema not found, sampling instead")
		newSchema, sampleRecord, err = sample.SampleSchema(opts, s.processName, session, lgr)
		if err != nil {
			return nil, err
		}
	}

	s.schemaLock.Lock()
	s.schema = newSchema
	s.schemaLock.Unlock()

	return sampleRecord, nil
}

func (s *Server) resampleSchema(mtx *dsync.DMutex, lgr *log.Logger, initialSampleRecord *sample.Record) error {
	err := mtx.Lock(s.lifetimeCtx)
	if err != nil {
		return err
	}

	session, err := s.sessionProvider.AdminSession(s.lifetimeCtx)
	if err != nil {
		return err
	}
	defer session.Close()

	newSchema, newSampleRecord, err := sample.SampleSchema(&s.cfg.Schema.Sample, s.processName, session, lgr)
	if err != nil {
		return err
	}

	lastGen, err := sample.LatestGeneration(&s.cfg.Schema.Sample, session, lgr)
	if err != nil {
		return err
	}

	newSampleRecord.Version.Generation = lastGen + 1

	err = sample.InsertSampleRecord(newSampleRecord, session, lgr)
	if err != nil {
		return err
	}

	s.schemaLock.Lock()
	s.schema = newSchema
	s.schemaLock.Unlock()

	return nil
}

func (s *Server) writeInitialSample(mtx *dsync.DMutex, lgr *log.Logger, initialSampleRecord *sample.Record) error {
	session, err := s.sessionProvider.AdminSession(s.lifetimeCtx)
	if err != nil {
		return err
	}
	defer session.Close()

	newSchema, err := sample.ReadSchema(&s.cfg.Schema.Sample, session, lgr)
	if err != nil {
		return err
	}

	if newSchema != nil {
		// some other mongosqld has now written a schema, so we can abort
		lgr.Logf(log.DebugLow, "aborting initial schema write; a newer schema was discovered")

		s.schemaLock.Lock()
		s.schema = newSchema
		s.schemaLock.Unlock()

		return nil
	}

	err = mtx.Lock(s.lifetimeCtx)
	if err != nil {
		return err
	}

	err = sample.InsertSampleRecord(initialSampleRecord, session, lgr)
	return err
}
