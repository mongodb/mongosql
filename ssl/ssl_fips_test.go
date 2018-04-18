// +build !darwin

package ssl

import (
	"os"
	"testing"

	"github.com/10gen/openssl"
)

func TestSetFIPSMode(t *testing.T) {

	variant := os.Getenv("VARIANT")
	switch variant {
	case "amazon", "debian81", "ubuntu1404":
		t.Skip("skipping test on variant without FIPS-enabled OpenSSL")
	default:
		// we should run this test on any other variant
	}

	if openssl.FIPSMode() {
		t.Fatal("Expected FIPS mode to be disabled, but was enabled")
	}

	err := setFIPSMode(true)
	if err != nil {
		t.Fatal(err)
	}

	if !openssl.FIPSMode() {
		t.Fatal("Expected FIPS mode to be enabled, but was disabled")
	}

	err = setFIPSMode(false)
	if err != nil {
		t.Fatal(err)
	}

	if openssl.FIPSMode() {
		t.Fatal("Expected FIPS mode to be disabled, but was enabled")
	}
}
