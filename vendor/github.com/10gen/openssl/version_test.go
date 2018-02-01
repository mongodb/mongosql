// Copyright (C) MongoDB, Inc. 2018-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package openssl

import (
	"testing"
)

func TestVersion(t *testing.T) {
	v := Version
	b := BuildVersion
	if len(v) == 0 {
		t.Fatal("Version string is empty")
	}
	if len(b) == 0 {
		t.Fatal("BuildVersion string is empty")
	}
	t.Logf("Built with headers from: %s", BuildVersion)
	t.Logf("   Tests linked against: %s", Version)
}
