test-log-newlines: build-mongosqld run-mongosqld _test-log-newlines
_test-log-newlines:
	$(ENV) testdata/bin/test-log-newlines.sh

