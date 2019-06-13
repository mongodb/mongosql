package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/10gen/sqlproxy/schema/persist"
	"github.com/10gen/sqlproxy/schema/sample"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Persistor is an interface that encompasses the stored-schema-management
// behavior required by the Manager.
type Persistor interface {
	InsertSchema(context.Context, *drdl.Schema) (primitive.ObjectID, error)
	UpsertName(context.Context, string, primitive.ObjectID) error
	FindSchemaByName(context.Context, string) (*drdl.Schema, error)
}

// Sampler is an interface that encompasses the sampling behavior required by
// the Manager.
type Sampler interface {
	Sample(context.Context) (*schema.Schema, error)
}

// Manager is a type that provides high-level schema-management functionality.
type Manager struct {
	cfg Config
	lg  log.Logger

	// persistor contains the Persistor that will be used to perform all
	// stored-schema-management operations.
	persistor Persistor

	// sampler contains the Sampler that will be used to perform all schema
	// sampling operations.
	sampler Sampler

	// schema contains the most recent schema.
	schema *schema.Schema

	// schemaAltered indicates whether the current schema has been manually altered.
	schemaAltered bool

	// schemaMx guards the schema and schemaAltered fields, since the schema
	// will often be set and accessed from different threads.
	schemaMx sync.RWMutex

	// lastErr contains the last error encountered by the schema manager's main
	// loop. Storing this error will allow us to expose more detailed errors
	// when an initial schema is not available.
	lastErr error

	// lastErrMx guards the lastErr field, since it will often be set and
	// accessed from different threads.
	lastErrMx sync.RWMutex

	// rootCancelFunc is a function that will cancel the main loop's context. It
	// is nil until Start is called, and thus cannot be invoked before then. All
	// other contexts created by the Manager inherit from that root context, and
	// so will be killed by this function.
	rootCancelFunc context.CancelFunc

	// cancelMx guards the rootCancelFunc field, which must be set and accessed
	// from different goroutines.
	cancelMx sync.Mutex

	// refreshIntervalChangedChan is a channel with a buffer of size 1 that is
	// used to trigger a reset of the sample refresh loop when the refresh
	// interval is changed. This makes it possible to decrease the
	// refreshInterval without having to wait for the previous refresh interval
	// to expire for the new interval to take effect. This channel should only
	// be sent on by a single goroutine that monitors for interval changes, and
	// should only be read by the refresh loop goroutine.
	refreshIntervalChangedChan chan struct{}
}

// NewManager returns a new schema Manager that will use the provided schema
// configuration, variables, and session provider.
func NewManager(cfg Config, lg log.Logger, sp *mongodb.SessionProvider, schemaSource string) *Manager {
	component := fmt.Sprintf("%-10v [manager]", log.SchemaComponent)
	lg = log.NewComponentLogger(component, lg)
	return newManager(
		cfg, lg,
		persist.NewPersistor(sp, schemaSource),
		sample.NewSampler(cfg.SampleConfig(), lg, sp),
	)
}

func newManager(
	cfg Config,
	lg log.Logger,
	persistor Persistor,
	sampler Sampler,
) *Manager {
	return &Manager{
		cfg:                        cfg,
		lg:                         lg,
		persistor:                  persistor,
		sampler:                    sampler,
		refreshIntervalChangedChan: make(chan struct{}, 1),
	}
}

// Start kicks off the main loop used to perform background tasks. Start should
// be called before calling any other Manager methods, and should only ever be
// called once on any given Manager.
func (mgr *Manager) Start() {

	// create a cancelable context for the main loop and store its cancelFunc on
	// the Manager struct
	ctx, cancelFunc := context.WithCancel(context.Background())
	mgr.cancelMx.Lock()
	mgr.rootCancelFunc = cancelFunc
	mgr.cancelMx.Unlock()

	// run the schema refresh loop
	procutil.PanicSafeGo(func() {
		mgr.runSchemaRefreshLoop(ctx)
	}, func(err interface{}) {
		mgr.lg.Fatalf(log.Always, "panic in schema refresh loop: %v", err)
	})

	// run a loop to watch for and signal on refresh interval changes
	procutil.PanicSafeGo(func() {
		mgr.watchRefreshIntervalChanges(ctx)
	}, func(err interface{}) {
		mgr.lg.Fatalf(log.Always, "panic in refresh interval watch loop: %v", err)
	})
}

// initializeSchema does whatever is needed to get an initial schema and set
// mgr.schema to that value. If using a file-based schema, mgr.schema can be set
// directly. Otherwise, we must sample or retrieve a persisted schema. In that
// case, we retry the operation every five seconds until it succeeds.
func (mgr *Manager) initializeSchema(ctx context.Context) {

	// if we are in file-based schema mode, then we set the file-based schema as the
	// initial schema, and no further work is needed.
	if mgr.cfg.Mode() == FileBasedSchemaMode {
		mgr.setSchema(mgr.cfg.FileBasedSchema())
		return
	}

	// try to get an initial schema every five seconds until it succeeds or the
	// manager is stopped
	procutil.RetryWithDelay(ctx.Done(), 5*time.Second, true, func() bool {
		mgr.lg.Infof(log.Admin, "attempting to initialize schema")

		_, err := mgr.refreshSchema(ctx)

		// if there was an error trying to get the initial schema, return false
		// so the operation will be retried
		if err != nil {
			mgr.setLastErr(err, "error initializing schema")
			return false
		}

		// log an informational message and return true, since we got an initial
		// schema and don't need to retry any further
		mgr.lg.Infof(log.Admin, "obtained initial schema")
		return true
	})
}

// HasSchema returns whether an initial schema has been obtained.
func (mgr *Manager) HasSchema() bool {
	return mgr.getSchema() != nil
}

// persistSchema stores the provided schema in the sampleSource database and
// updates the schemaName document to point at the newly-inserted schema. This
// function does not change any state on the Manager.
func (mgr *Manager) persistSchema(ctx context.Context, sch *schema.Schema) error {
	mgr.lg.Infof(log.Admin, "persisting schema")
	oid, err := mgr.persistor.InsertSchema(ctx, sch.ToDRDL())
	if err != nil {
		return err
	}

	schemaName := mgr.cfg.SchemaName()
	err = mgr.persistor.UpsertName(ctx, schemaName, oid)
	if err != nil {
		return err
	}

	mgr.lg.Infof(log.Admin, "successfully persisted schema as %q", schemaName)
	return nil
}

// sampleSchema creates a new Schema by sampling and returns it. This function
// has no side-effects.
func (mgr *Manager) sampleSchema(ctx context.Context) (*schema.Schema, error) {
	mgr.lg.Infof(log.Admin, "sampling schema")
	sch, err := mgr.sampler.Sample(ctx)
	if err != nil {
		return nil, err
	}
	return sch, nil
}

// fetchStoredSchema attempts to fetch a stored schema and return it. This
// function has no side-effects.
func (mgr *Manager) fetchStoredSchema(ctx context.Context) (*schema.Schema, error) {
	schemaName := mgr.cfg.SchemaName()
	mgr.lg.Infof(log.Admin, "fetching stored schema with name %q", schemaName)
	drdlSchema, err := mgr.persistor.FindSchemaByName(ctx, schemaName)
	if err != nil {
		return nil, err
	}
	return schema.NewFromDRDL(mgr.lg, drdlSchema)
}

// setSchema sets the current schema to a copy of the provided schema. This
// function is thread-safe.
func (mgr *Manager) setSchema(sch *schema.Schema) {
	mgr.schemaMx.Lock()
	defer mgr.schemaMx.Unlock()
	mgr.schema = sch.DeepCopy()
}

// setSchemaAltered sets the schemaAltered field to the provided value. This
// function is thread-safe.
func (mgr *Manager) setSchemaAltered(altered bool) {
	mgr.schemaMx.Lock()
	defer mgr.schemaMx.Unlock()
	mgr.schemaAltered = altered
}

// schemaIsAltered returns the value of the schemaAltered field. This function
// is thread-safe.
func (mgr *Manager) schemaIsAltered() bool {
	mgr.schemaMx.RLock()
	a := mgr.schemaAltered
	mgr.schemaMx.RUnlock()
	return a
}

// Close shuts down the background goroutines used by the Manager.
func (mgr *Manager) Close() {
	mgr.cancelMx.Lock()
	mgr.rootCancelFunc()
	mgr.cancelMx.Unlock()
}

// Schema gets a copy of the current schema. If the manager is running in
// custom-schema mode, it will first attempt to refresh the schema. If the
// returned schema is nil, then an initial schema is not yet available. This
// function is thread-safe.
func (mgr *Manager) Schema(ctx context.Context) *schema.Schema {
	if mgr.cfg.Mode() == CustomSchemaMode {
		_, err := mgr.refreshSchema(ctx)
		if err != nil {
			mgr.setLastErr(err, "failed to refresh custom schema")
		}
	}
	return mgr.getSchema()
}

// getSchema gets a copy of the current schema. If the returned schema is nil,
// then an initial schema is not yet available. This function is thread-safe.
func (mgr *Manager) getSchema() *schema.Schema {
	mgr.schemaMx.RLock()
	defer mgr.schemaMx.RUnlock()
	return mgr.schema.DeepCopy()
}

// setLastErr logs the provided error and sets mgr.lastErr to the concatenation
// of the provided message string with the error's message string.
func (mgr *Manager) setLastErr(err error, msg string) {
	if msg != "" {
		err = fmt.Errorf("%s: %v", msg, err)
	}
	mgr.lg.Warnf(log.Admin, "%v", err)
	mgr.lastErrMx.Lock()
	defer mgr.lastErrMx.Unlock()
	mgr.lastErr = err
}

// GetLastErr returns the last error encountered during sampling or
// persisted-schema management, or nil if no such error has been encountered
// yet.
func (mgr *Manager) GetLastErr() error {
	mgr.lastErrMx.RLock()
	defer mgr.lastErrMx.RUnlock()
	return mgr.lastErr
}

// Resample refreshes the schema by sampling, if allowed by the current mode.
func (mgr *Manager) Resample(ctx context.Context) (*schema.Schema, error) {
	if !mgr.HasSchema() {
		return nil, fmt.Errorf("cannot resample before schema is initialized")
	}

	switch mgr.cfg.Mode() {
	case FileBasedSchemaMode:
		return nil, fmt.Errorf("cannot resample when using a file-based schema")
	case CustomSchemaMode:
		return nil, fmt.Errorf("cannot resample in custom-schema mode")
	case StandaloneSchemaMode, AutoSchemaMode:
		// resampling is allowed in these modes
	}

	return mgr.refreshSchema(ctx)
}

// Alter applies the provided alterations to the current schema, persisting the
// altered schema if appropriate for the current mode.
func (mgr *Manager) Alter(ctx context.Context, alts []*schema.Alteration) (*schema.Schema, error) {

	switch mgr.cfg.Mode() {
	case CustomSchemaMode, AutoSchemaMode:
		return nil, fmt.Errorf("alterations not allowed in stored-schema modes")
	}

	if !mgr.HasSchema() {
		return nil, fmt.Errorf("cannot alter schema before it is initialized")
	}

	sch := mgr.getSchema()
	altered, err := sch.Altered(alts...)
	if err != nil {
		return nil, err
	}

	mgr.setSchemaAltered(true)
	mgr.setSchema(altered)
	return altered, nil
}

// refreshSchema obtains a new schema via the mechanism indicated by the current
// mode and sets it as the Manager's current schema. If the Manager is running
// in auto schema mode, the updated schema is also persisted.
func (mgr *Manager) refreshSchema(ctx context.Context) (*schema.Schema, error) {
	var newSch *schema.Schema
	var err error

	switch mgr.cfg.Mode() {
	case CustomSchemaMode:
		// in custom mode, just get the most recent stored schema
		newSch, err = mgr.fetchStoredSchema(ctx)
	case StandaloneSchemaMode:
		// in standalone mode, just resample to get a new schema
		newSch, err = mgr.sampleSchema(ctx)
	case AutoSchemaMode:
		// in auto mode, resample and then persist the sampled schema
		newSch, err = mgr.sampleSchema(ctx)
		if err == nil {
			err = mgr.persistSchema(ctx, newSch)
		}
	case FileBasedSchemaMode:
		panic("refreshSchema should never be called in file-based schema mode")
	default:
		panic(fmt.Errorf("unknown schema mode %q", mgr.cfg.Mode()))
	}

	// if we encountered an error obtaining or persisting an updated schema,
	// return without updating the current schema
	if err != nil {
		return nil, err
	}

	mgr.setSchemaAltered(false)
	mgr.setSchema(newSch)
	return newSch, nil
}

// runSchemaRefreshLoop runs the loop that performs automatic schema updates.
func (mgr *Manager) runSchemaRefreshLoop(ctx context.Context) {

	// wait until an initial schema has been obtained
	mgr.initializeSchema(ctx)

	// If we are in file-based schema mode, no loop is necessary, so we just exit.
	if mgr.cfg.Mode() == FileBasedSchemaMode {
		return
	}

	// In any other mode, we refresh the schema every SampleRefreshIntervalSecs.
	for {
		var timeoutChan <-chan time.Time
		delay := mgr.cfg.RefreshInterval()
		if delay != 0 {
			mgr.lg.Debugf(log.Dev, "next schema refresh in %v seconds", delay.Seconds())
			timeoutChan = time.After(delay)
		} else {
			mgr.lg.Debugf(log.Dev, "refresh interval set to zero; schema will not be refreshed")
		}

		select {
		case <-ctx.Done():
			// root context was cancelled, exit the main loop
			mgr.lg.Warnf(log.Admin, "schema.Manager main loop terminating: %v", ctx.Err())
			return
		case <-mgr.refreshIntervalChangedChan:
			// refresh interval changed, restart this loop from the top
			mgr.lg.Debugf(log.Dev, "schema refresh interval changed; updating refresh loop timeout")
			continue
		case <-timeoutChan:
			// refresh interval expired, refresh schema now
		}

		if mgr.schemaIsAltered() {
			mgr.setLastErr(fmt.Errorf("schema has been altered"), "skipping automatic resample")
			continue
		}

		_, err := mgr.refreshSchema(ctx)
		if err != nil {
			mgr.setLastErr(err, "error refreshing schema")
			continue
		}

		mgr.lg.Infof(log.Admin, "refreshed schema")
	}
}

// watchRefreshIntervalChanges runs a loop that checks for changes to the schema
// refresh interval, signaling the main schema update loop whenever a change is
// detected.
func (mgr *Manager) watchRefreshIntervalChanges(ctx context.Context) {
	lastRefreshInterval := mgr.cfg.RefreshInterval()
	procutil.RepeatWithDelay(ctx.Done(), 5*time.Second, false, func() {
		currentRefreshInterval := mgr.cfg.RefreshInterval()
		if currentRefreshInterval != lastRefreshInterval {
			mgr.lg.Debugf(
				log.Dev,
				"schema refresh interval changed from %v seconds to %v seconds",
				lastRefreshInterval.Seconds(), currentRefreshInterval.Seconds(),
			)
			select {
			case mgr.refreshIntervalChangedChan <- struct{}{}:
				// put the signal into the channel
			default:
				// there is already a signal in this chan; continue
			}
		}
		lastRefreshInterval = currentRefreshInterval
	})
}
