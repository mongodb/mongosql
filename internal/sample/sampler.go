package sample

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/dsync"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/internal/variable"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

// NewSampler returns a Sampler using the schema sample configuration
// options and process name passed in. The returned Sampler is attached
// to the MongoDB cluster the sessionProvider is configured with.
func NewSampler(opts *config.SchemaSampleOptions, processName string,
	sessionProvider *mongodb.SessionProvider, variables *variable.Container) *Sampler {
	lgr := log.NewComponentLogger(
		fmt.Sprintf("%-10v [schemaDiscovery]", log.SamplerComponent),
		log.GlobalLogger(),
	)

	s := &Sampler{
		opts:            opts,
		lgr:             lgr,
		processName:     processName,
		sessionProvider: sessionProvider,
		variables:       variables,
	}

	return s
}

// A Sampler is used to derive a schema from a MongoDB cluster.
// It is configured via NewSampler to determine what namespaces
// should be included in the derived schema.
type Sampler struct {
	schemaLock sync.RWMutex
	schema     *schema.Schema

	opts            *config.SchemaSampleOptions
	dmtx            *dsync.DMutex
	lgr             log.Logger
	sessionProvider *mongodb.SessionProvider
	processName     string
	variables       *variable.Container
}

// Alter adds the alterations alts to the schema held by this
// Sampler.
func (s *Sampler) Alter(ctx context.Context, alts []*schema.Alteration) error {
	if s.opts.Mode == config.ReadSampleMode && s.opts.Source != "" {
		return fmt.Errorf("cannot alter schema in clustered read mode")
	} else if s.opts.Mode == config.WriteSampleMode {
		return fmt.Errorf("cannot alter schema in clustered write mode")
	}

	if s.opts.Mode == config.WriteSampleMode {
		err := s.alterAndPersistSchema(ctx, alts)
		if err != nil {
			return err
		}
	}

	s.schemaLock.Lock()
	s.schema.AddAlterations(alts...)
	s.schemaLock.Unlock()

	return nil
}

// Refresh forces this Sampler to generate a new version of
// the schema.
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

// Schema returns the schema derived by this Sampler.
func (s *Sampler) Schema(ctx context.Context) *schema.Schema {
	var newSchema *schema.Schema
	heuristic := config.GetMappingHeuristic(s.variables.GetString(variable.SchemaMappingHeuristic))
	opts := NewSchemaSampleOptionsWithHeuristic(s.opts, heuristic)
	if opts.source != "" {
		session, err := s.sessionProvider.AuthenticatedAdminSession(ctx)
		if err == nil {
			defer func() {
				err = session.Close()
				if err != nil {
					s.lgr.Warnf(log.Dev, "could not close session %v", err)
				}
			}()
			newSchema, err = ReadSchema(ctx, opts, session, s.lgr)
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

// Run is the entry function used to get a Sampler to
// to begin the process of generating a schema from a
// set of MongoDB namespaces.
func (s *Sampler) Run(ctx context.Context) {
	if s.opts.Mode == config.ReadSampleMode && s.opts.Source == "" {
		s.lgr.Infof(log.Admin, "sampler running in standalone mode")
	} else if s.opts.Mode == config.ReadSampleMode && s.opts.Source != "" {
		s.lgr.Infof(log.Admin, "sampler running in clustered read mode")
	} else {
		s.lgr.Infof(log.Admin, "sampler running in clustered write mode")
	}

	var resampleDelay = func(cancel, refreshChange <-chan struct{}) bool {
		for {
			refreshInterval := s.variables.GetInt64(variable.SampleRefreshIntervalSecs)
			if refreshInterval != 0 {
				refreshInterval := time.Duration(refreshInterval) * time.Second
				select {
				case <-cancel:
					return false
				case <-refreshChange:
					// Looks like the refresh interval has changed. If so, wait on the new
					// time.
					continue
				case <-time.After(refreshInterval):
					return true
				}
			}

			return false
		}
	}

	var sampleRecord *Record
	var err error
	var resample = func(cancel, refreshChange <-chan struct{}) {
		for {
			s.schemaLock.RLock()
			altered := len(s.schema.Alterations()) > 0
			s.schemaLock.RUnlock()
			if altered {
				s.lgr.Warnf(
					log.Admin,
					"skipping resampling schema: schema has been altered",
				)
			} else {
				s.lgr.Infof(log.Admin, "re-sampling schema")
				err = s.resampleSchema(ctx)
				if err != nil {
					s.lgr.Errf(log.Admin, "failed re-sampling schema: %v", err)
				}
			}

			if !resampleDelay(cancel, refreshChange) {
				return
			}
		}
	}

	// 1. All mongosqld's will attempt read an existing schema from the server and sample if one
	// does not exist. When sampling occurred and was successful, the sample record will be
	// returned. Until this completes successfully, we cannot move on.
	retryWaitTimeSecs := 5 * time.Second
	util.RetryWithDelay(ctx.Done(), retryWaitTimeSecs, true, func() bool {
		s.lgr.Infof(log.Always, "initializing schema")
		sampleRecord, err = s.initializeSchema(ctx)
		if err == nil {
			return true
		}

		s.lgr.Errf(log.Always, "unable to initialize schema: %v", err)
		return false
	})

	refreshChange := make(chan struct{})
	pastRefreshInterval := s.variables.GetInt64(variable.SampleRefreshIntervalSecs)
	go util.RetryWithDelay(
		ctx.Done(),
		retryWaitTimeSecs,
		false,
		func() bool {
			currentRefreshInterval := s.variables.GetInt64(variable.SampleRefreshIntervalSecs)
			if currentRefreshInterval != pastRefreshInterval {
				pastRefreshInterval = currentRefreshInterval
				refreshChange <- struct{}{}
			}

			return false
		},
	)

	type resampleFunction func(<-chan struct{}, <-chan struct{})
	var resampleLoop = func(resampleFunc resampleFunction) {
		// Even if the refresh interval is zero, it could still be changed by the
		// user later on.  Because of this possibility, we need to stay vigilant
		// every now and then to make sure we are honoring the user's desired
		// refresh interval. Therefore, if it ever changes, we will call resample()
		// to begin looping, and if that ever returns (if the user turns refresh
		// interval back to 0) then we will return false in either case and
		// continue being vigilant. When vigilant, we re-check every 5 seconds.
		util.RetryWithDelay(
			ctx.Done(),
			retryWaitTimeSecs,
			true,
			func() bool {
				if s.variables.GetInt64(variable.SampleRefreshIntervalSecs) != 0 {
					resampleFunc(ctx.Done(), refreshChange)
				}
				return false
			},
		)
	}
	// 2. if we are a reader, we just need to re-sample the schema every so often.
	if s.opts.Mode == config.ReadSampleMode {
		resampleLoop(resample)
		return
	}

	// 3. otherwise, we are a writer and need to re-sample the schema every so often
	// and persist that schema.
	s.dmtx = dsync.NewDMutex(dsync.DMutexConfig{
		Name:            "mongosqld-schema",
		DatabaseName:    s.opts.Source,
		CollectionName:  mongodb.LockCollection,
		Logger:          s.lgr,
		ProcessName:     s.processName,
		SessionProvider: s.sessionProvider,
		// Expiration time will be 5 minutes after the last refresh.
		// Every 30 seconds, we'll refresh the lock.
		HeartbeatInterval: 30 * time.Second,
		Timeout:           5 * time.Minute,
	})

	// use a different context here because if ctx is done (which is what likely prompted the
	// exit of this function and hence the invocation of the defer statement), we need a
	// different context or else the unlock wouldn't take place.
	defer func() {
		err = s.dmtx.Unlock(context.Background())
		if err != nil {
			s.lgr.Errf(log.Admin, "unable to unlock mutex: %v", err)
		}
	}()

	// 4. If we have a sample, it means that we didn't read a schema from the server.
	// Therefore, we need to persist this back to the server or, if we fail to do that,
	// read a schema that may show up in the future.
	if sampleRecord != nil && len(sampleRecord.Namespaces) > 0 {
		s.lgr.Infof(log.Admin, "writing sampled schema")
		err := s.dmtx.Lock(ctx)
		if err == nil {
			// try to do this once initially... if it doesn't work, we'll start looping
			var session *mongodb.Session
			session, err = s.sessionProvider.AuthenticatedAdminSessionPrimary(ctx)
			if err == nil {
				err = InsertSampleRecord(ctx, sampleRecord, session, s.lgr)
				_ = session.Close()
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

	var resampleAndPersist = func(cancel, refreshChange <-chan struct{}) {
		for {
			s.schemaLock.RLock()
			altered := len(s.schema.Alterations()) > 0
			s.schemaLock.RUnlock()
			if altered {
				s.lgr.Warnf(log.Admin, "skipping resampling schema: schema has been altered")
			} else {
				s.lgr.Infof(log.Admin, "re-sampling schema")
				err := s.resampleAndPersistSchema(ctx)
				if err != nil {
					s.lgr.Errf(log.Admin, "failed re-sampling schema: %v", err)
				}
			}

			if !resampleDelay(cancel, refreshChange) {
				return
			}
		}
	}

	resampleLoop(resampleAndPersist)
}

func (s *Sampler) initializeSchema(ctx context.Context) (rec *Record, err error) {
	var session *mongodb.Session
	session, err = s.sessionProvider.AuthenticatedAdminSession(ctx)
	if err != nil {
		return nil, err
	}
	defer util.CheckDeferredFunc(session.Close, &err)

	var newSchema *schema.Schema

	opts := NewSchemaSampleOptions(s.opts)
	if s.opts.Source != "" {
		newSchema, err = ReadSchema(ctx, opts, session, s.lgr)
		if err != nil {
			return nil, err
		}
	}

	if newSchema == nil {
		s.lgr.Infof(log.Admin, "stored schema not found, sampling instead")
		newSchema, rec, err = Schema(
			ctx,
			opts,
			s.processName,
			session,
			s.lgr,
		)
		if err != nil {
			return nil, err
		}
	}

	s.schemaLock.Lock()
	s.schema = newSchema
	s.schemaLock.Unlock()

	return rec, nil
}

func (s *Sampler) alterAndPersistSchema(ctx context.Context, als []*schema.Alteration) (err error) {
	err = s.dmtx.Lock(ctx)
	if err != nil {
		return err
	}

	var session *mongodb.Session
	session, err = s.sessionProvider.AuthenticatedAdminSession(ctx)
	if err != nil {
		return err
	}
	defer util.CheckDeferredFunc(session.Close, &err)

	var record *Record
	opts := NewSchemaSampleOptions(s.opts)
	record, err = LatestRecord(ctx, opts, session)
	if err != nil {
		return err
	}

	record.Alter(als)

	var writeSession *mongodb.Session
	writeSession, err = s.sessionProvider.AuthenticatedAdminSessionPrimary(ctx)
	if err != nil {
		return err
	}
	defer util.CheckDeferredFunc(writeSession.Close, &err)

	return InsertSampleRecord(ctx, record, writeSession, s.lgr)
}

func (s *Sampler) resampleSchema(ctx context.Context) error {
	session, err := s.sessionProvider.AuthenticatedAdminSession(ctx)
	if err != nil {
		return err
	}
	defer func() {
		cerr := session.Close()
		if err == nil {
			err = cerr
		}
	}()

	heuristic := config.GetMappingHeuristic(s.variables.GetString(variable.SchemaMappingHeuristic))
	opts := NewSchemaSampleOptionsWithHeuristic(s.opts, heuristic)
	opts = opts.WithOptimizeViewSampling(s.variables.GetBool(variable.OptimizeViewSampling))
	opts = opts.WithSampleSize(s.variables.GetInt64(variable.SampleSize))
	opts = opts.WithMaxNumColumnsPerTable(s.variables.GetInt64(variable.MaxNumColumnsPerTable))
	opts = opts.WithMaxNestedTableDepth(s.variables.GetInt64(variable.MaxNestedTableDepth))
	newSchema, _, err := Schema(
		ctx,
		opts,
		s.processName,
		session,
		s.lgr,
	)
	if err != nil {
		return err
	}

	s.schemaLock.RLock()
	alterations := len(s.schema.Alterations())
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

	session, err := s.sessionProvider.AuthenticatedAdminSession(ctx)
	if err != nil {
		return err
	}

	defer func() {
		cerr := session.Close()
		if err == nil {
			err = cerr
		}
	}()

	heuristic := config.GetMappingHeuristic(s.variables.GetString(variable.SchemaMappingHeuristic))
	opts := NewSchemaSampleOptionsWithHeuristic(s.opts, heuristic)
	optimizeViewSampling := s.variables.GetBool(variable.OptimizeViewSampling)
	opts = opts.WithOptimizeViewSampling(optimizeViewSampling)
	opts = opts.WithSampleSize(s.variables.GetInt64(variable.SampleSize))
	newSchema, newSampleRecord, err := Schema(
		ctx,
		opts,
		s.processName,
		session,
		s.lgr,
	)
	if err != nil {
		return err
	}

	lastGen, err := LatestGeneration(ctx, opts, session, s.lgr)
	if err != nil {
		return err
	}

	newSampleRecord.Version.Generation = lastGen + 1

	var writeSession *mongodb.Session
	writeSession, err = s.sessionProvider.AuthenticatedAdminSessionPrimary(ctx)
	if err != nil {
		return err
	}

	defer func() {
		cerr := writeSession.Close()
		if err == nil {
			err = cerr
		}
	}()

	err = InsertSampleRecord(ctx, newSampleRecord, writeSession, s.lgr)
	if err != nil {
		return err
	}

	s.schemaLock.RLock()
	alterations := len(s.schema.Alterations())
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

func (s *Sampler) writeInitialSample(ctx context.Context, initialSampleRecord *Record) (err error) {
	var session *mongodb.Session
	session, err = s.sessionProvider.AuthenticatedAdminSession(ctx)
	if err != nil {
		return err
	}
	defer util.CheckDeferredFunc(session.Close, &err)

	var newSchema *schema.Schema
	opts := NewSchemaSampleOptions(s.opts)
	newSchema, err = ReadSchema(ctx, opts, session, s.lgr)
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

	var writeSession *mongodb.Session
	writeSession, err = s.sessionProvider.AuthenticatedAdminSessionPrimary(ctx)
	if err != nil {
		return err
	}
	defer util.CheckDeferredFunc(writeSession.Close, &err)

	err = InsertSampleRecord(ctx, initialSampleRecord, writeSession, s.lgr)
	return err
}
