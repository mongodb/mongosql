# Test various system variables.
test-query-max-execution-time-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG)
test-query-max-execution-time-failure: run-mongodb build-mongosqld run-mongosqld _test-query-max-execution-time-failure

_test-query-max-execution-time-failure:
	$(ENV) EXPECT_ERROR=1 SLEEP_TIME=3 testdata/bin/test-query-max-execution-time.sh

test-query-max-execution-time-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG)
test-query-max-execution-time-success: run-mongodb build-mongosqld run-mongosqld _test-query-max-execution-time-success

_test-query-max-execution-time-success:
	$(ENV) EXPECT_ERROR=0 SLEEP_TIME=1 testdata/bin/test-query-max-execution-time.sh
