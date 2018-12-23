package metrics

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/10gen/sqlproxy/log"
)

// A Tracker takes a Record and processes it according to
// some reporting strategy.
type Tracker interface {
	Track([]Record)
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

// Track prints the provided records to the mongosqld log in json format.
func (t *logTracker) Track(recs []Record) {
	for _, r := range recs {
		js, err := r.ToJSON()
		if err != nil {
			t.lg.Errf(log.Dev, "failed to marshal metrics record to JSON: %v", err)
			continue
		}
		t.lg.Debugf(log.Dev, "%s", js)
	}
}

func metricsComponentLogger() log.Logger {
	return log.NewComponentLogger("METRICS    [monitoring]", log.GlobalLogger())
}

type noOpTracker struct{}

// NewNoOpTracker returns a new Tracker that does nothing with
// any record it is asked to track.
func NewNoOpTracker() Tracker {
	return noOpTracker{}
}

// Track does nothing with the provided records.
func (noOpTracker) Track([]Record) {}

type stitchTracker struct {
	lg  log.Logger
	url string
}

// NewStitchTracker returns a new Tracker that sends metrics
// records to a stitch app.
func NewStitchTracker(url string) Tracker {
	return &stitchTracker{
		lg:  metricsComponentLogger(),
		url: url,
	}
}

// Track sends the provided records to a stitch app.
func (t *stitchTracker) Track(recs []Record) {
	js, err := json.Marshal(recs)
	if err != nil {
		t.lg.Errf(log.Dev, "failed to marshal metrics records to JSON: %v", err)
		return
	}
	data := bytes.NewBuffer(js)

	res, err := http.Post(t.url, "application/json", data)
	if err != nil {
		t.lg.Errf(log.Dev, "failed to send metrics records to stitch: %v", err)
	} else if res.StatusCode != 201 {
		buf := bytes.NewBuffer([]byte{})
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			t.lg.Errf(log.Dev, "unexpected response from stitch: %s", res.Status)
		} else {
			t.lg.Errf(log.Dev, "unexpected response from stitch: %s %s", res.Status, buf.String())
		}
	} else {
		t.lg.Infof(log.Dev, "sent %d metrics records to stitch", len(recs))
	}
}
