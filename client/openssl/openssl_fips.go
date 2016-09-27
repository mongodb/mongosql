// +build -darwin

package openssl

import (
	"github.com/10gen/sqlproxy/options"
	"github.com/spacemonkeygo/openssl"
)

func init() {
	sslInitializationFunctions = append(sslInitializationFunctions, SetUpFIPSMode)
}

func SetUpFIPSMode(opts options.Options) error {
	if err := openssl.FIPSModeSet(opts.UseFIPSMode()); err != nil {
		return fmt.Errorf("couldn't set FIPS mode: %v", err)
	}
	return nil
}
