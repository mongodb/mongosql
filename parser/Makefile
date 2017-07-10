# Copyright 2012, Google Inc. All rights reserved.
# Use of this source code is governed by a BSD-style license that can
# be found in the LICENSE file.

MAKEFLAGS = -s

sql.go: sql.y
	echo "There is currently 4 shift/reduce error and 321 reduce/reduce errors"
	goyacc -o sql.go sql.y

clean:
	rm -f y.output sql.go
