package server

import (
	"context"

	"github.com/10gen/sqlproxy/evaluator/metrics"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/log"
)

// enqueueRecord adds the provided record to the next batch of records to be
// sent to the tracker.
func (s *Server) enqueueRecord(r metrics.Record) {
	s.recordsMx.Lock()
	defer s.recordsMx.Unlock()

	// add provided record to the current batch
	s.records = append(s.records, r)

	select {
	case s.recordsChan <- s.records:
		// sent the current batch to the tracker,
		// so we start a new batch
		s.records = nil
	default:
		// tracker is busy, so we do nothing
	}
}

// runTracker takes each batch of records sent on s.recordsChan and processes it
// by calling s.trackRecords(). This function should be called asynchronously.
func (s *Server) runTracker(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case recs := <-s.recordsChan:
			s.trackRecords(recs)
		}
	}
}

// track sends the provided records to the backend specified by the
// @@global.metrics_backend variable.
func (s *Server) trackRecords(recs []metrics.Record) {
	var tracker metrics.Tracker

	backend := s.variables.GetString(variable.MetricsBackend)
	switch backend {
	case variable.NoMetricsBackend:
		tracker = metrics.NewNoOpTracker()
	case variable.LogMetricsBackend:
		tracker = metrics.NewLogTracker()
	case variable.StitchMetricsBackend:
		stitchURL := s.cfg.Metrics.StitchURL
		if stitchURL == "" {
			s.logger.Warnf(
				log.Admin,
				"not tracking query: no url provided in config for metrics backend %q",
				backend,
			)
			tracker = metrics.NewNoOpTracker()
		} else {
			tracker = metrics.NewStitchTracker(stitchURL)
		}
	default:
		s.logger.Warnf(
			log.Admin,
			"'%s' metrics backend not supported",
			backend,
		)
		tracker = metrics.NewNoOpTracker()
	}

	tracker.Track(recs)
}
