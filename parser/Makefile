# Copyright 2012, Google Inc. All rights reserved.
# Use of this source code is governed by a BSD-style license that can
# be found in the LICENSE file.

MAKEFLAGS = -s

sql.go: sql.y
	echo "There are currently 8 shift/reduce errors and 398 reduce/reduce errors"
	goyacc -o sql.go sql.y
	sed -i '' '/package/,$$!d' sql.go

clean:
	rm -f y.output sql.go
