// +build !nofips

package ssl

import (
	"fmt"

	"github.com/10gen/openssl"
)

func init() {
	FipsModeSetter = setFIPSMode
}

func setFIPSMode(enabled bool) error {
	if err := openssl.FIPSModeSet(enabled); err != nil {
		return fmt.Errorf("couldn't set FIPS mode: %v", err)
	}
	if openssl.FIPSMode() != enabled {
		return fmt.Errorf("failed to change FIPS mode")
	}
	return nil
}
