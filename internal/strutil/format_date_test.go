package strutil_test

import (
	"testing"

	. "github.com/10gen/sqlproxy/internal/strutil"

	"github.com/stretchr/testify/require"
)

func TestFormatDate(t *testing.T) {
	req := require.New(t)

	_, err := FormatDate("2014-01-02T15:04:05.000Z")
	req.Nil(err)

	_, err = FormatDate("2014-03-02T15:05:05Z")
	req.Nil(err)

	_, err = FormatDate("2014-04-02T15:04Z")
	req.Nil(err)

	_, err = FormatDate("2014-04-02T15:04-0800")
	req.Nil(err)

	_, err = FormatDate("2014-04-02T15:04:05.000-0600")
	req.Nil(err)

	_, err = FormatDate("2014-04-02T15:04:05-0500")
	req.Nil(err)

	_, err = FormatDate("invalid string format")
	req.NotNil(err)

}
