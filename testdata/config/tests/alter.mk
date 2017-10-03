
_try-alter:
	$(ENV) EXPECTED_ERROR="$(EXPECTED_ERROR)" EXPECTED_STATUS="$(EXPECTED_STATUS)" testdata/bin/test-alter.sh

_test-altered: QUERY := select count(*) from information_schema.columns where table_schema = 'test' and table_name = 'foo';
_test-altered: EXPECTED := 5
_test-altered: _test-mysql-query

_test-not-altered: QUERY := select count(*) from information_schema.columns where table_schema = 'test' and table_name = 'test1';
_test-not-altered: EXPECTED := 5
_test-not-altered: _test-mysql-query

test-alter-success: EXPECTED_STATUS := 0
test-alter-success: build-mongosqld run-mongodb restore-data run-mongosqld _test-connect-success _try-alter _test-altered

test-alter-failure: EXPECTED_STATUS := 1
test-alter-failure: build-mongosqld run-mongodb restore-data run-mongosqld _test-connect-success _try-alter _test-not-altered

test-alter-drdl: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: cannot alter schema: schema was loaded from a file
test-alter-drdl: test-alter-failure

test-alter-clustered-read: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/clustered
test-alter-clustered-read: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: cannot alter schema in clustered read mode
test-alter-clustered-read: test-alter-failure

test-alter-standalone: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic
test-alter-standalone: test-alter-success

test-alter-clustered-write: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/clustered,sqlproxy/schema/write
test-alter-clustered-write: test-alter-success

test-alter-flush: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic
test-alter-flush: EXPECTED_STATUS := 0
test-alter-flush: build-mongosqld run-mongodb restore-data run-mongosqld _test-connect-success _try-alter _test-flush _test-not-altered
_test-flush:
	$(ENV) EXPECTED_STATUS='0' EXPECTED_ERROR='' testdata/bin/test-flush.sh
