//+build go1.13

package testutil

import "testing"

// TestingInit runs testing.Init() on go1.13 and later, and does
// nothing on earlier go versions.
func TestingInit() {
	testing.Init()
}
