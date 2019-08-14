package versionutil

import (
	"fmt"
	"strings"

	"github.com/10gen/sqlproxy/internal/strutil"
)

// MySQLFixedWidthVersionCode is the version string with the minor and patch
// versions 0 padded to 2 digits. It is the code used for conditional
// comment statements/queries.
type MySQLFixedWidthVersionCode string

// GetMySQLFixedWidthVersionCode gets the fixed width MySQL version string for the given
// human representable version string:
// 5.7.12  ==> 50712
// 5.7.3   ==> 50703
// 5.12.13 ==> 51213
func GetMySQLFixedWidthVersionCode(mysqlVersion string) (MySQLFixedWidthVersionCode, error) {
	versionParts := strings.Split(mysqlVersion, ".")
	if len(versionParts) != 3 {
		return "", fmt.Errorf("valid version string must contain major, minor, and patch parts, but got %q", mysqlVersion)
	}
	if len(versionParts[0]) != 1 || !strutil.IsNumeric(versionParts[0]) {
		return "", fmt.Errorf("version major must be 1 numeric character, got %q", mysqlVersion)
	}
	for _, part := range versionParts[1:] {
		if !(len(part) == 1 || len(part) == 2) || !strutil.IsNumeric(part) {
			return "", fmt.Errorf("valid version string parts must be 1 or 2 numeric characters, but got %q", mysqlVersion)
		}
	}
	pad := func(part string) string {
		if len(part) < 2 {
			return "0" + part
		}
		return part
	}
	return MySQLFixedWidthVersionCode(
		fmt.Sprintf("%s%s%s",
			versionParts[0],
			pad(versionParts[1]),
			pad(versionParts[2])),
	), nil
}
