// +build -darwin

package ssl

import (
	"fmt"

	"github.com/spacemonkeygo/openssl"
)

func init() {
	fipsModeSetter = setFIPSMode
}

func setFIPSMode(enabled bool) error {
	if err := openssl.FIPSModeSet(enabled); err != nil {
		return fmt.Errorf("couldn't set FIPS mode: %v", err)
	}
	return nil
}
