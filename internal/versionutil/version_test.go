package versionutil_test

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/versionutil"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	req := require.New(t)

	out, err := versionutil.GetMySQLFixedWidthVersionCode("5.7.12")
	req.NoError(err)
	req.Equal(versionutil.MySQLFixedWidthVersionCode("50712"), out)

	out, err = versionutil.GetMySQLFixedWidthVersionCode("5.7.2")
	req.NoError(err)
	req.Equal(versionutil.MySQLFixedWidthVersionCode("50702"), out)

	out, err = versionutil.GetMySQLFixedWidthVersionCode("5.17.12")
	req.NoError(err)
	req.Equal(versionutil.MySQLFixedWidthVersionCode("51712"), out)

	_, err = versionutil.GetMySQLFixedWidthVersionCode(".17.12")
	req.EqualError(err, `version major must be 1 numeric character, got ".17.12"`)

	_, err = versionutil.GetMySQLFixedWidthVersionCode("5.a.12")
	req.EqualError(err, `valid version string parts must be 1 or 2 numeric characters, but got "5.a.12"`)

	_, err = versionutil.GetMySQLFixedWidthVersionCode("5.1.1c")
	req.EqualError(err, `valid version string parts must be 1 or 2 numeric characters, but got "5.1.1c"`)

	_, err = versionutil.GetMySQLFixedWidthVersionCode("55.17.12")
	req.EqualError(err, `version major must be 1 numeric character, got "55.17.12"`)

	_, err = versionutil.GetMySQLFixedWidthVersionCode("5..12")
	req.EqualError(err, `valid version string parts must be 1 or 2 numeric characters, but got "5..12"`)

	_, err = versionutil.GetMySQLFixedWidthVersionCode("5.123.12")
	req.EqualError(err, `valid version string parts must be 1 or 2 numeric characters, but got "5.123.12"`)

	_, err = versionutil.GetMySQLFixedWidthVersionCode("5.12.")
	req.EqualError(err, `valid version string parts must be 1 or 2 numeric characters, but got "5.12."`)

	_, err = versionutil.GetMySQLFixedWidthVersionCode("5.12.123")
	req.EqualError(err, `valid version string parts must be 1 or 2 numeric characters, but got "5.12.123"`)

	_, err = versionutil.GetMySQLFixedWidthVersionCode("5.12.12.9")
	req.EqualError(err, `valid version string must contain major, minor, and patch parts, but got "5.12.12.9"`)

	_, err = versionutil.GetMySQLFixedWidthVersionCode("5.12")
	req.EqualError(err, `valid version string must contain major, minor, and patch parts, but got "5.12"`)
}
