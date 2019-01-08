# Test various system variables.
test-query-max-execution-time-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/interval-2
test-query-max-execution-time-failure: run-mongodb build-mongosqld run-mongosqld _test-query-max-execution-time-failure

_test-query-max-execution-time-failure: QUERY := SELECT SLEEP(1)
_test-query-max-execution-time-failure:
	$(ENV) EXPECT_ERROR=1 QUERY="$(QUERY)" MAX_EXECUTION_TIME=500 testdata/bin/test-query-max-execution-time.sh

test-query-max-execution-time-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/interval-2
test-query-max-execution-time-success: run-mongodb build-mongosqld run-mongosqld _test-query-max-execution-time-success

_test-query-max-execution-time-success: QUERY := SELECT SLEEP(0.5)
_test-query-max-execution-time-success:
	$(ENV) EXPECT_ERROR=0 QUERY="$(QUERY)" MAX_EXECUTION_TIME=1000 testdata/bin/test-query-max-execution-time.sh

test-query-max-execution-time-pushdown-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/interval-2
test-query-max-execution-time-pushdown-success: run-mongodb build-mongosqld run-mongosqld restore-integration-data _test-query-max-execution-time-pushdown-success

_test-query-max-execution-time-pushdown-success: QUERY := use test; select * from test1
_test-query-max-execution-time-pushdown-success:
	$(ENV) EXPECT_ERROR=0 QUERY="$(QUERY)" MAX_EXECUTION_TIME=1000 testdata/bin/test-query-max-execution-time.sh

test-query-max-execution-time-pushdown-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/interval-2
test-query-max-execution-time-pushdown-failure: run-mongodb build-mongosqld run-mongosqld restore-integration-data _test-query-max-execution-time-pushdown-failure

_test-query-max-execution-time-pushdown-failure: QUERY := use test; select * from test1 join test2 2 where 5 = 5
_test-query-max-execution-time-pushdown-failure:
	$(ENV) EXPECT_ERROR=1 QUERY="$(QUERY)" MAX_EXECUTION_TIME=1 testdata/bin/test-query-max-execution-time.sh
