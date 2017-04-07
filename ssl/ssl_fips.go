// +build -darwin

package ssl

import (
	"fmt"

	"github.com/10gen/sqlproxy/options"
	"github.com/spacemonkeygo/openssl"
)

func init() {
	sslInitializationFunctions = append(sslInitializationFunctions, setUpFIPSMode)
}

func setUpFIPSMode(opts options.Options) error {
	if err := openssl.FIPSModeSet(opts.UseFIPSMode()); err != nil {
		return fmt.Errorf("couldn't set FIPS mode: %v", err)
	}
	return nil
}
