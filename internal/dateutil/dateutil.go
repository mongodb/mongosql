package dateutil

import (
	"errors"
	"time"

	"github.com/tkuchiki/go-timezone"
)

var ianaTZCachedName string

// UnixMillis returns the Unix Epoch time in milliseconds.
// This is useful because dates in mongodb are represented as 64 bit
// integers that are unix epoch in milliseconds (including negatives
// for years before 1970).
func UnixMillis(t time.Time) int64 {
	return t.Unix()*1000 + int64(t.Nanosecond()/1e6)
}

// GetIANATimezoneName gets the IANA timezone database name for this
// machine's current timezone.
func GetIANATimezoneName(t time.Time) (string, error) {
	if ianaTZCachedName != "" {
		return ianaTZCachedName, nil
	}

	shortZone := t.Format("MST")
	hasDST := observesDST(t)

	tzs, err := timezone.GetTimezones(shortZone)
	if err != nil {
		return "", err
	}

	now := time.Now()

	for _, tz := range tzs {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return "", err
		}
		dst := observesDST(now.In(loc))
		if dst == hasDST {
			ianaTZCachedName = tz
			return tz, nil
		}
	}

	return "", errors.New("unable to get timezone info")
}

func observesDST(t time.Time) bool {
	year := t.Year()
	loc := t.Location()

	_, winterOffset := time.Date(year, 1, 1, 0, 0, 0, 0, loc).Zone()
	_, summerOffset := time.Date(year, 7, 1, 0, 0, 0, 0, loc).Zone()

	return winterOffset != summerOffset
}
