package dateutil

import "time"

// UnixMillis returns the Unix Epoch time in milliseconds.
// This is useful because dates in mongodb are represented as 64 bit
// integers that are unix epoch in milliseconds (including negatives
// for years before 1970).
func UnixMillis(t time.Time) int64 {
	return t.Unix()*1000 + int64(t.Nanosecond()/1e6)
}
