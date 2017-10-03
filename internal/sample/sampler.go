package sample

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/dsync"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

func NewSampler(opts *config.SchemaSampleOptions, processName string, sessionProvider *mongodb.SessionProvider) *Sampler {
	lgr := log.NewComponentLogger(
		fmt.Sprintf("%-10v [schemaDiscovery]", log.SamplerComponent),
		log.GlobalLogger(),
	)

	s := &Sampler{
		opts:            opts,
		lgr:             lgr,
		processName:     processName,
		sessionProvider: sessionProvider,
	}

	return s
}

type Sampler struct {
	schemaLock sync.RWMutex
	schema     *schema.Schema

	opts            *config.SchemaSampleOptions
	dmtx            *dsync.DMutex
	lgr             *log.Logger
	sessionProvider *mongodb.SessionProvider
	processName     string
}

func (s *Sampler) Alter(ctx context.Context, alts []*schema.Alteration) error {
	if s.opts.Mode == config.ReadSampleMode && s.opts.Source != "" {
		return fmt.Errorf("cannot alter schema in clustered read mode")
	}

	if s.opts.Mode == config.WriteSampleMode {
		err := s.alterAndPersistSchema(ctx, alts)
		if err != nil {
			return err
		}
	}

	s.schemaLock.Lock()
	s.schema.Alterations = append(s.schema.Alterations, alts...)
	s.schemaLock.Unlock()

	return nil
}

func (s *Sampler) Refresh(ctx context.Context) error {
	if s.opts.Mode == config.ReadSampleMode && s.opts.Source != "" {
		return fmt.Errorf("cannot refresh sample in clustered read mode")
	}

	if s.opts.Mode == config.ReadSampleMode {
		err := s.resampleSchema(ctx)
		if err != nil {
			return fmt.Errorf("failed refreshing schema: %v", err)
		}
	} else {
		err := s.resampleAndPersistSchema(ctx)
		if err != nil {
			return fmt.Errorf("failed refreshing schema: %v", err)
		}
	}

	return nil
}

func (s *Sampler) Schema(ctx context.Context) *schema.Schema {
	var newSchema *schema.Schema

	if s.opts.Source != "" {
		session, err := s.sessionProvider.AdminSession(ctx)
		if err == nil {
			defer session.Close()
			newSchema, err = ReadSchema(s.opts, session, s.lgr)
		}

		if err != nil {
			s.lgr.Warnf(log.Dev, "could not fetch most recent schema: %v", err)
		}
	}

	if newSchema != nil {
		s.schemaLock.Lock()
		s.schema = newSchema
		s.schemaLock.Unlock()
	} else {
		s.schemaLock.RLock()
		newSchema = s.schema
		s.schemaLock.RUnlock()
	}

	return newSchema.DeepCopy()
}

func (s *Sampler) Run(ctx context.Context) {
	if s.opts.Mode == config.ReadSampleMode && s.opts.Source == "" {
		s.lgr.Infof(log.Admin, "sampler running in standalone mode")
	} else if s.opts.Mode == config.ReadSampleMode && s.opts.Source != "" {
		s.lgr.Infof(log.Admin, "sampler running in clustered read mode")
	} else {
		s.lgr.Infof(log.Admin, "sampler running in clustered write mode")
	}

	var sampleRecord *Record
	var err error

	// 1. All mongosqld's will attempt read an existing schema from the server and sample if one
	// does not exist. When sampling occurred and was successful, the sample record will be returned.
	// Until this completes successfully, we cannot move on.
	util.RetryWithDelay(ctx.Done(), 5*time.Second, true, func() bool {
		s.lgr.Infof(log.Always, "initializing schema")
		sampleRecord, err = s.initializeSchema(ctx)
		if err == nil {
			return true
		}

		s.lgr.Errf(log.Always, "unable to initialize schema: %v", err)
		return false
	})

	// 2. if we are a reader, we just need to re-sample the schema every so often.
	if s.opts.Mode == config.ReadSampleMode {
		if s.opts.RefreshIntervalSecs > 0 {
			util.RepeatWithDelay(ctx.Done(), time.Duration(s.opts.RefreshIntervalSecs)*time.Second, false, func() {
				s.schemaLock.RLock()
				altered := len(s.schema.Alterations) > 0
				s.schemaLock.RUnlock()
				if altered {
					s.lgr.Infof(log.Admin, "re-sampling schema")
					err := s.resampleSchema(ctx)
					if err != nil {
						s.lgr.Errf(log.Admin, "failed re-sampling schema: %v", err)
					}
				} else {
					s.lgr.Warnf(log.Admin, "skipping resampling schema: schema has been altered")
				}
			})
		}

		return
	}

	// 3. otherwise, we are a writer and need to re-sample the schema every so often and persist that schema.
	s.dmtx = dsync.NewDMutex(dsync.DMutexConfig{
		Name:            "mongosqld-schema",
		DatabaseName:    s.opts.Source,
		CollectionName:  LockCollection,
		Logger:          s.lgr,
		ProcessName:     s.processName,
		SessionProvider: s.sessionProvider,
		// Expiration time will be 5 minutes after the last refresh.
		// Every 30 seconds, we'll refresh the lock.
		HeartbeatInterval: 30 * time.Second,
		Timeout:           5 * time.Minute,
	})

	// use a different context here because if ctx is done (which is what likely prompted the exit of this function
	// and hence the invocation of the defer statement), we need a different context or else the unlock wouldn't take place.
	defer s.dmtx.Unlock(context.Background())

	// 4. If we have a sample, it means that we didn't read a schema from the server. Therefore, we need to
	// persist this back to the server or, if we fail to do that, read a schema that may show up in the future.
	if sampleRecord != nil && len(sampleRecord.Namespaces) > 0 {
		s.lgr.Infof(log.Admin, "writing sampled schema")
		err := s.dmtx.Lock(ctx)
		if err == nil {
			// try to do this once initially... if it doesn't work, we'll start looping
			var session *mongodb.Session
			session, err = s.sessionProvider.AdminSession(ctx)
			if err == nil {
				err = InsertSampleRecord(sampleRecord, session, s.lgr)
				session.Close()
			}
		}

		if err != nil {
			s.lgr.Errf(log.Admin, "unable to persist initial schema: %v", err)

			util.RetryWithDelay(ctx.Done(), 1*time.Minute, false, func() bool {
				err := s.writeInitialSample(ctx, sampleRecord)
				if err != nil {
					s.lgr.Errf(log.Admin, "unable to persist initial schema: %v", err)
					return false
				}

				return true
			})
		}
	}

	// 5. write once = done
	if s.opts.RefreshIntervalSecs <= 0 {
		return
	}

	// 6. Re-sample every writeIntervalSecs and persist the schema
	util.RepeatWithDelay(ctx.Done(), time.Duration(s.opts.RefreshIntervalSecs)*time.Second, false, func() {
		s.schemaLock.RLock()
		altered := len(s.schema.Alterations) > 0
		s.schemaLock.RUnlock()
		if altered {
			s.lgr.Infof(log.Admin, "re-sampling schema")
			err := s.resampleAndPersistSchema(ctx)
			if err != nil {
				s.lgr.Errf(log.Admin, "failed re-sampling schema: %v", err)
			}
		} else {
			s.lgr.Warnf(log.Admin, "skipping resampling schema: schema has been altered")
		}
	})
}

func (s *Sampler) initializeSchema(ctx context.Context) (*Record, error) {
	session, err := s.sessionProvider.AdminSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var newSchema *schema.Schema
	var sampleRecord *Record

	if s.opts.Source != "" {
		newSchema, err = ReadSchema(s.opts, session, s.lgr)
		if err != nil {
			return nil, err
		}
	}

	if newSchema == nil {
		s.lgr.Infof(log.Admin, "stored schema not found, sampling instead")
		newSchema, sampleRecord, err = SampleSchema(s.opts, s.processName, session, s.lgr)
		if err != nil {
			return nil, err
		}
	}

	s.schemaLock.Lock()
	s.schema = newSchema
	s.schemaLock.Unlock()

	return sampleRecord, nil
}

func (s *Sampler) alterAndPersistSchema(ctx context.Context, alts []*schema.Alteration) error {
	err := s.dmtx.Lock(ctx)
	if err != nil {
		return err
	}

	session, err := s.sessionProvider.AdminSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()

	record, err := LatestRecord(s.opts, session, s.lgr)
	if err != nil {
		return err
	}

	record.Alter(alts)

	return InsertSampleRecord(record, session, s.lgr)
}

func (s *Sampler) resampleSchema(ctx context.Context) error {
	session, err := s.sessionProvider.AdminSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()

	newSchema, _, err := SampleSchema(s.opts, s.processName, session, s.lgr)
	if err != nil {
		return err
	}

	s.schemaLock.RLock()
	alterations := len(s.schema.Alterations)
	s.schemaLock.RUnlock()
	if alterations > 0 {
		alterationStr := util.Pluralize(alterations, "alteration", "alterations")
		s.lgr.Warnf(log.Admin, "resampling overwrote %d existing %s", alterations, alterationStr)
	}

	s.schemaLock.Lock()
	s.schema = newSchema
	s.schemaLock.Unlock()

	return nil
}

func (s *Sampler) resampleAndPersistSchema(ctx context.Context) error {
	err := s.dmtx.Lock(ctx)
	if err != nil {
		return err
	}

	session, err := s.sessionProvider.AdminSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()

	newSchema, newSampleRecord, err := SampleSchema(s.opts, s.processName, session, s.lgr)
	if err != nil {
		return err
	}

	lastGen, err := LatestGeneration(s.opts, session, s.lgr)
	if err != nil {
		return err
	}

	newSampleRecord.Version.Generation = lastGen + 1

	err = InsertSampleRecord(newSampleRecord, session, s.lgr)
	if err != nil {
		return err
	}

	s.schemaLock.RLock()
	alterations := len(s.schema.Alterations)
	s.schemaLock.RUnlock()
	if alterations > 0 {
		alterationStr := util.Pluralize(alterations, "alteration", "alterations")
		s.lgr.Warnf(log.Admin, "resampling overwrote %d existing %s", alterations, alterationStr)
	}

	s.schemaLock.Lock()
	s.schema = newSchema
	s.schemaLock.Unlock()

	return nil
}

func (s *Sampler) writeInitialSample(ctx context.Context, initialSampleRecord *Record) error {
	session, err := s.sessionProvider.AdminSession(ctx)
	if err != nil {
		return err
	}
	defer session.Close()

	newSchema, err := ReadSchema(s.opts, session, s.lgr)
	if err != nil {
		return err
	}

	if newSchema != nil {
		// some other mongosqld has now written a schema, so we can abort
		s.lgr.Infof(log.Dev, "aborting initial schema write; a newer schema was discovered")

		s.schemaLock.Lock()
		s.schema = newSchema
		s.schemaLock.Unlock()

		return nil
	}

	err = s.dmtx.Lock(ctx)
	if err != nil {
		return err
	}

	err = InsertSampleRecord(initialSampleRecord, session, s.lgr)
	return err
}
