package metrics

import (
	"github.com/10gen/sqlproxy/log"
)

// A Tracker takes a Record and processes it according to
// some reporting strategy.
type Tracker interface {
	Track(Record)
}

type logTracker struct {
	lg log.Logger
}

// NewLogTracker returns a new Tracker that will print all
// metrics records to the mongosqld log.
func NewLogTracker() Tracker {
	return &logTracker{
		lg: metricsComponentLogger(),
	}
}

// Track prints the provided record to the mongosqld log in json format.
func (t *logTracker) Track(r Record) {
	js, err := r.ToJSON()
	if err != nil {
		panic(err)
	}
	t.lg.Debugf(log.Dev, "%s", js)
}

func metricsComponentLogger() log.Logger {
	return log.NewComponentLogger("METRICS", log.GlobalLogger())
}

type noOpTracker struct{}

// NewNoOpTracker returns a new Tracker that does nothing with
// any record it is asked to track.
func NewNoOpTracker() Tracker {
	return noOpTracker{}
}

// Track does nothing with the provided record.
func (noOpTracker) Track(Record) {}
