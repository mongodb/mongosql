package dateutil

import (
	"testing"
	"time"
)

func TestGetIANATimezoneName(t *testing.T) {
	assertTimesEqual := func(t *testing.T, expected, actual time.Time) {
		expectedUTC := expected.UTC()
		actualUTC := actual.UTC()
		if !expectedUTC.Equal(actualUTC) {
			t.Fatalf("expected %v, but got %v", expectedUTC, actualUTC)
		}
	}

	t.Run("calling once gets back a location name that matches this machine's timezone", func(t *testing.T) {
		now := time.Now()

		tzName, err := GetIANATimezoneName(now)
		if err != nil {
			t.Fatalf("failed to get IANA Timezone Database name: %v", err)
		}

		loc, err := time.LoadLocation(tzName)
		if err != nil {
			t.Fatalf("failed to load location for IANA Timezone Database name: %v", err)
		}

		// We can't just compare loc with now.Location because it's possible
		// GetIANATimezoneName picks a different but correct name. For example,
		// Mongo HQ is in America/New_York, but America/Detroit would also be
		// a valid location since it is in the same timezone.
		alsoNow := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), loc)
		assertTimesEqual(t, now, alsoNow)
	})

	t.Run("calling multiple times gets the same result", func(t *testing.T) {
		now := time.Now()

		tzName1, err := GetIANATimezoneName(now)
		if err != nil {
			t.Fatalf("failed first call to get IANA Timezone Database name: %v", err)
		}

		tzName2, err := GetIANATimezoneName(now)
		if err != nil {
			t.Fatalf("failed second call to get IANA Timezone Database name: %v", err)
		}

		if tzName1 != tzName2 {
			t.Fatalf("got different names for different invocations: %v != %v", tzName1, tzName2)
		}

		// At this point, we know tzName1 == tzName2, so we just proceed with one of them.
		loc, err := time.LoadLocation(tzName1)
		if err != nil {
			t.Fatalf("failed to load location for IANA Timezone Database name: %v", err)
		}

		alsoNow := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), loc)
		assertTimesEqual(t, now, alsoNow)
	})
}

func TestObservesDST(t *testing.T) {
	tests := []struct {
		name   string
		hasDST bool
	}{
		{"America/New_York", true},
		{"America/Denver", true},
		{"America/Phoenix", false},
		{"Europe/Amsterdam", true},
		{"Asia/Pyongyang", false},
	}

	now := time.Now()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loc, err := time.LoadLocation(test.name)
			if err != nil {
				t.Fatalf("failed to load location: %v", err)
			}

			actual := observesDST(now.In(loc))
			if test.hasDST != actual {
				t.Fatalf("expected %v but got %v", test.hasDST, actual)
			}
		})
	}
}
