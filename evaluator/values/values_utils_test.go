package values_test

import (
	"testing"
	"time"

	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/schema"
	"github.com/stretchr/testify/require"
)

func TestParseDateTimeMongo(t *testing.T) {
	tests := []struct {
		date         string
		expectedOk   bool
		expectedTime time.Time
	}{
		// valid without rearranging
		{"20170201", true, time.Date(2017, 2, 1, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010-11-12", true, time.Date(2010, 11, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010-11-12 00:00:00", true, time.Date(2010, 11, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010-11-12T00:00:00", true, time.Date(2010, 11, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"10-11-12", true, time.Date(2010, 11, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"10-11-12 00:00:00", true, time.Date(2010, 11, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"10-11-12T00:00:00", true, time.Date(2010, 11, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010/11/12", true, time.Date(2010, 11, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"123/1/1", true, time.Date(123, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)},

		// valid with rearranging MM/DD/YYYY to YYYY-MM-DD
		{"1/1/1", true, time.Date(1, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/1/1", true, time.Date(1, 12, 1, 0, 0, 0, 0, schema.DefaultLocale)},
		{"1/12/1", true, time.Date(1, 1, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"1/1/12", true, time.Date(2012, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)},
		{"1/1/123", true, time.Date(123, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)},
		{"1/1/1234", true, time.Date(1234, 1, 1, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/1/12", true, time.Date(2012, 12, 1, 0, 0, 0, 0, schema.DefaultLocale)},
		{"1/12/12", true, time.Date(2012, 1, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/1/123", true, time.Date(123, 12, 1, 0, 0, 0, 0, schema.DefaultLocale)},
		{"1/12/123", true, time.Date(123, 1, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/1/1234", true, time.Date(1234, 12, 1, 0, 0, 0, 0, schema.DefaultLocale)},
		{"1/12/1234", true, time.Date(1234, 1, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/1", true, time.Date(1, 12, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/12", true, time.Date(2012, 12, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/123", true, time.Date(123, 12, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/1234", true, time.Date(1234, 12, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/1234 ", true, time.Date(1234, 12, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/12 ", true, time.Date(2012, 12, 12, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/12 11:11:11", true, time.Date(2012, 12, 12, 11, 11, 11, 0, schema.DefaultLocale)},
		{"12/12/12T11:11:11", true, time.Date(2012, 12, 12, 11, 11, 11, 0, schema.DefaultLocale)},

		// valid with time parts
		{"10/11/12T11:12:13Z", true, time.Date(2012, 10, 11, 11, 12, 13, 0, schema.DefaultLocale)},
		{"10/11/12T11:12:13", true, time.Date(2012, 10, 11, 11, 12, 13, 0, schema.DefaultLocale)},
		{"10/11/12 11:12:13", true, time.Date(2012, 10, 11, 11, 12, 13, 0, schema.DefaultLocale)},
		{"10/11/12 11:12:13Z", true, time.Date(2012, 10, 11, 11, 12, 13, 0, schema.DefaultLocale)},
		{"10-11-12 11:12:13Z", true, time.Date(2010, 11, 12, 11, 12, 13, 0, schema.DefaultLocale)},
		{"2010-11-12 11:12:13Z", true, time.Date(2010, 11, 12, 11, 12, 13, 0, schema.DefaultLocale)},
		{"2010/11/12 11:12:13Z", true, time.Date(2010, 11, 12, 11, 12, 13, 0, schema.DefaultLocale)},
		{"2010/11/12 1:1:1", true, time.Date(2010, 11, 12, 1, 1, 1, 0, schema.DefaultLocale)},
		{"2010/11/12 10:1:1", true, time.Date(2010, 11, 12, 10, 1, 1, 0, schema.DefaultLocale)},
		{"2010/11/12 1:10:1", true, time.Date(2010, 11, 12, 1, 10, 1, 0, schema.DefaultLocale)},
		{"2010/11/12 1:1:10", true, time.Date(2010, 11, 12, 1, 1, 10, 0, schema.DefaultLocale)},
		{"2010/11/12 1:10:10", true, time.Date(2010, 11, 12, 1, 10, 10, 0, schema.DefaultLocale)},
		{"2010/11/12 1.1:10", true, time.Date(2010, 11, 12, 1, 1, 10, 0, schema.DefaultLocale)},
		{"2010/11/12 1.1.10", true, time.Date(2010, 11, 12, 1, 1, 10, 0, schema.DefaultLocale)},
		{"2010/11/12 1:1.10", true, time.Date(2010, 11, 12, 1, 1, 10, 0, schema.DefaultLocale)},
		{"2010/11/12 1:1.10.1", true, time.Date(2010, 11, 12, 1, 1, 10, 1e8, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.1", true, time.Date(2010, 11, 12, 1, 1, 10, 1e8, schema.DefaultLocale)},
		{"2010/11/12 1.1.10.1", true, time.Date(2010, 11, 12, 1, 1, 10, 1e8, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.1234567Z", true, time.Date(2010, 11, 12, 1, 1, 10, 123456e3, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.123456Z", true, time.Date(2010, 11, 12, 1, 1, 10, 123456e3, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.12345Z", true, time.Date(2010, 11, 12, 1, 1, 10, 12345e4, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.1234Z", true, time.Date(2010, 11, 12, 1, 1, 10, 1234e5, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.123Z", true, time.Date(2010, 11, 12, 1, 1, 10, 123e6, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.12Z", true, time.Date(2010, 11, 12, 1, 1, 10, 12e7, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.1Z", true, time.Date(2010, 11, 12, 1, 1, 10, 1e8, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.Z", true, time.Date(2010, 11, 12, 1, 1, 10, 0, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.", true, time.Date(2010, 11, 12, 1, 1, 10, 0, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.1", true, time.Date(2010, 11, 12, 1, 1, 10, 1e8, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.12", true, time.Date(2010, 11, 12, 1, 1, 10, 12e7, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.123", true, time.Date(2010, 11, 12, 1, 1, 10, 123e6, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.1234", true, time.Date(2010, 11, 12, 1, 1, 10, 1234e5, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.12345", true, time.Date(2010, 11, 12, 1, 1, 10, 12345e4, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.123456", true, time.Date(2010, 11, 12, 1, 1, 10, 123456e3, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.1234567", true, time.Date(2010, 11, 12, 1, 1, 10, 123456e3, schema.DefaultLocale)},
		{"2010/11/12 1:1", true, time.Date(2010, 11, 12, 1, 1, 0, 0, schema.DefaultLocale)},
		{"2010/11/12 11:11", true, time.Date(2010, 11, 12, 11, 11, 0, 0, schema.DefaultLocale)},
		{"2010/11/12 11:11Z", true, time.Date(2010, 11, 12, 11, 11, 0, 0, schema.DefaultLocale)},

		// invalid
		{"1", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"123", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"1234", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12345", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"123456", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"1234567", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/123/1", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/12345", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/1234T", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/12T", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/12 anything", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"12/12/12Tanything", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"10$10$11 11$10$23", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"10$10$11", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"10^10^11", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"10.10.11", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"10-10/11", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"10/10-11", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010/11/12 1aa:1:10", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010/11/12 123:1:10", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010/11/12 1:123:10", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010/11/12 1:1:123", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010/11/12 1:1::1", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010/11/12 xxxx1234", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010/11/12 xxxx1234", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
		{"2010/11/12 1:1:10.12345678", false, time.Date(0, 0, 0, 0, 0, 0, 0, schema.DefaultLocale)},
	}

	for _, test := range tests {
		t.Run(test.date, func(t *testing.T) {
			req := require.New(t)

			actualTime, _, actualOk := values.ParseDateTimeMongo(test.date)

			req.Equal(test.expectedOk, actualOk, "should not have been parsed")

			if test.expectedOk {
				req.Equal(test.expectedTime, actualTime, "incorrect datetime")
			}
		})
	}
}
