
_try-alter:
	$(ENV) EXPECTED_ERROR="$(EXPECTED_ERROR)" EXPECTED_STATUS="$(EXPECTED_STATUS)" testdata/bin/test-alter.sh

_test-altered: QUERY := select count(*) from information_schema.columns where table_schema = 'test' and table_name = 'foo';
_test-altered: EXPECTED := 2
_test-altered: _test-mysql-query

_test-not-altered: QUERY := select count(*) from information_schema.columns where table_schema = 'test' and table_name = 'sample_test';
_test-not-altered: EXPECTED := 2
_test-not-altered: _test-mysql-query

_test-not-altered-flush: QUERY := select count(*) from information_schema.columns where table_schema = 'test' and table_name = 'sample_test';
_test-not-altered-flush: EXPECTED := 11
_test-not-altered-flush: _test-mysql-query

test-alter-success: EXPECTED_STATUS := 0
test-alter-success: build-mongosqld run-mongodb restore-integration-data run-mongosqld _test-schema-available _test-connect-success _try-alter _test-altered

test-alter-failure: EXPECTED_STATUS := 1
test-alter-failure: NUM_DOCS := 1
test-alter-failure: build-mongosqld run-mongodb _insert-sample-docs _write-initial-schema run-mongosqld _test-schema-available _test-connect-success _try-alter _test-not-altered

test-alter-auto-mode: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/auto,sqlproxy/schema/enable-alter
test-alter-auto-mode: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: alterations not allowed in stored-schema modes
test-alter-auto-mode: test-alter-failure

test-alter-custom-mode: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/custom,sqlproxy/schema/enable-alter
test-alter-custom-mode: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: alterations not allowed in stored-schema modes
test-alter-custom-mode: test-alter-failure

test-alter-flush: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/enable-alter
test-alter-flush: EXPECTED_STATUS := 0
test-alter-flush: build-mongosqld run-mongodb _write-initial-docs _write-initial-schema run-mongosqld _test-schema-available _test-connect-success _try-alter _test-flush _test-not-altered-flush
_test-flush:
	$(ENV) EXPECTED_STATUS='0' EXPECTED_ERROR='' testdata/bin/test-flush.sh
