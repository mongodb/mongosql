VARIANT := $(VARIANT)

INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG)
ifeq ($(INFRASTRUCTURE_CONFIG),)
INFRASTRUCTURE_CONFIG := default
endif

ENV = VARIANT=$(VARIANT) INFRASTRUCTURE_CONFIG=$(INFRASTRUCTURE_CONFIG)
EXPECTED = EXPECTED_STATUS=$(EXPECTED_STATUS) EXPECTED_ERROR="$(EXPECTED_ERROR)"

SCHEMA_UNAVAILABLE_ERROR = ERROR 1043 (08S01): MongoDB schema not yet available

default: test

build: build-mongodrdl build-mongosqld

build-mongodrdl:
	$(ENV) testdata/bin/build-mongodrdl.sh

build-mongosqld:
	$(ENV) testdata/bin/build-mongosqld.sh

check:
	$(ENV) testdata/bin/check-sourcelint.sh

check-races:
	testdata/bin/check-races.sh

check-yaml:
	$(ENV) testdata/bin/check-yamllint.sh

evergreen-validate:
	$(ENV) testdata/bin/evergreen-validate.sh

clean:
	$(ENV) testdata/bin/reset-testing-state.sh

download-data:
	testdata/bin/download-blackbox-data.sh
	testdata/bin/download-tableau-data.sh
	testdata/bin/download-tpch-data.sh

restore-data:
	$(ENV) SUITE="$(SUITE)" NO_FLUSH_SCHEMA="$(NO_FLUSH_SCHEMA)" testdata/bin/restore-test-data.sh

restore-integration-data: SUITE := internal
restore-integration-data: restore-data

run-mongodb:
	$(ENV) testdata/bin/start-mongodb.sh

run-mongosqld:
	$(ENV) testdata/bin/start-mongosqld.sh

setup-hooks:
	rm -rf .git/hooks
	ln -s ../.githooks .git/hooks

shell:
	mysql -P3307

start-all: build-mongosqld run-mongodb run-mongosqld
start-mongod-second: build-mongosqld run-mongosqld run-mongodb

test: test-unit test-module-integration test-integration

test-connect-failure: start-all _test-connect-failure
_test-connect-failure: EXPECTED_STATUS = 1
_test-connect-failure:
	$(ENV) $(EXPECTED) testdata/bin/test-simple-connect.sh

test-connect-success: start-all _test-connect-success
_test-connect-success: EXPECTED_STATUS = 0
_test-connect-success:
	$(ENV) $(EXPECTED) testdata/bin/test-simple-connect.sh

# test-connect-mongod-second ensures that we can still connect when mongodb is run
# after mongosqld is started. test-simple-connect waits 15s before trying to
# connect just to ensure things stablize (since start-mongosqld already waits 5s).
test-connect-mongod-second-success: start-mongod-second _test-connect-mongod-second-success
_test-connect-mongod-second-success: EXPECTED_STATUS = 0
_test-connect-mongod-second-success:
	$(ENV) $(EXPECTED) DELAY=15s testdata/bin/test-simple-connect.sh

test-integration: SUITE := internal
test-integration: test-connect-success restore-data _test-integration
_test-integration:
	$(ENV) SUITE="$(SUITE)" testdata/bin/run-integration-tests.sh

prepare-memory-limits: run-mongodb build-mongosqld

test-memory-limits: prepare-memory-limits _test-memory-limits-control _test-memory-limits-large-documents _test-memory-limits-deeply-nested-sub-documents _test-memory-limits-many-fields _test-memory-limits-many-databases _test-memory-limits-many-collections _test-memory-limits-many-documents _test-memory-limits-many-arrays _test-memory-limits-deeply-nested-arrays

test-memory-limits-control: prepare-memory-limits _test-memory-limits-control
_test-memory-limits-control:
	$(ENV) TEST_NAME='control' testdata/bin/run-memory-tests.sh
	$(ENV) TEST_NAME='control' testdata/bin/parse-memory-results.sh

test-memory-limits-large-documents: prepare-memory-limits _test-memory-limits-large-documents
_test-memory-limits-large-documents:
	$(ENV) TEST_NAME='large-documents' testdata/bin/run-memory-tests.sh
	$(ENV) TEST_NAME='large-documents' testdata/bin/parse-memory-results.sh

test-memory-limits-deeply-nested-sub-documents: prepare-memory-limits _test-memory-limits-deeply-nested-sub-documents
_test-memory-limits-deeply-nested-sub-documents:
	$(ENV) TEST_NAME='deeply-nested-sub-documents' testdata/bin/run-memory-tests.sh
	$(ENV) TEST_NAME='deeply-nested-sub-documents' testdata/bin/parse-memory-results.sh

test-memory-limits-many-fields: prepare-memory-limits _test-memory-limits-many-fields
_test-memory-limits-many-fields:
	$(ENV) TEST_NAME='many-fields' testdata/bin/run-memory-tests.sh
	$(ENV) TEST_NAME='many-fields' testdata/bin/parse-memory-results.sh

test-memory-limits-many-databases: prepare-memory-limits _test-memory-limits-many-databases
_test-memory-limits-many-databases:
	$(ENV) TEST_NAME='many-databases' testdata/bin/run-memory-tests.sh
	$(ENV) TEST_NAME='many-databases' testdata/bin/parse-memory-results.sh

test-memory-limits-many-collections: prepare-memory-limits _test-memory-limits-many-collections
_test-memory-limits-many-collections:
	$(ENV) TEST_NAME='many-collections' testdata/bin/run-memory-tests.sh
	$(ENV) TEST_NAME='many-collections' testdata/bin/parse-memory-results.sh

test-memory-limits-many-documents: prepare-memory-limits _test-memory-limits-many-documents
_test-memory-limits-many-documents:
	$(ENV) TEST_NAME='many-documents' testdata/bin/run-memory-tests.sh
	$(ENV) TEST_NAME='many-documents' testdata/bin/parse-memory-results.sh

test-memory-limits-many-arrays: prepare-memory-limits _test-memory-limits-many-arrays
_test-memory-limits-many-arrays:
	$(ENV) TEST_NAME='many-arrays' testdata/bin/run-memory-tests.sh
	$(ENV) TEST_NAME='many-arrays' testdata/bin/parse-memory-results.sh

test-memory-limits-deeply-nested-arrays: prepare-memory-limits _test-memory-limits-deeply-nested-arrays
_test-memory-limits-deeply-nested-arrays:
	$(ENV) TEST_NAME='deeply-nested-arrays' testdata/bin/run-memory-tests.sh
	$(ENV) TEST_NAME='deeply-nested-arrays' testdata/bin/parse-memory-results.sh

test-writes: SUITE := writes
test-writes: test-connect-success restore-data _test-integration


test-option-help: build-mongosqld
	$(ARTIFACTS_DIR)/bin/mongosqld --help

test-option-version: build-mongosqld
	$(ARTIFACTS_DIR)/bin/mongosqld --version

test-start-mongosqld: build-mongosqld _test-start-mongosqld
_test-start-mongosqld:
	$(ENV) $(EXPECTED) testdata/bin/test-start-mongosqld.sh

test-start-mongosqld-failure: build-mongosqld
test-start-mongosqld-failure: EXPECTED_STATUS = 1
test-start-mongosqld-failure:
	$(ENV) $(EXPECTED) testdata/bin/test-start-mongosqld.sh

test-unit:
	$(ENV) testdata/bin/run-unit-tests.sh

test-module-integration: test-connect-success
	$(ENV) BUILD_FLAGS="$(BUILD_FLAGS) integration" testdata/bin/run-unit-tests.sh

# include e2e test targets
E2E_TEST_DIR = testdata/e2e/tests
include $(E2E_TEST_DIR)/auth.mk
include $(E2E_TEST_DIR)/cleartext-auth.mk
include $(E2E_TEST_DIR)/gssapi.mk
include $(E2E_TEST_DIR)/inter-project-dependencies.mk
include $(E2E_TEST_DIR)/kill-query.mk
include $(E2E_TEST_DIR)/log-newlines.mk
include $(E2E_TEST_DIR)/log-rotation.mk
include $(E2E_TEST_DIR)/mongo-ssl.mk
include $(E2E_TEST_DIR)/mongo-uri.mk
include $(E2E_TEST_DIR)/mongodrdl.mk
include $(E2E_TEST_DIR)/schema.mk
include $(E2E_TEST_DIR)/server.mk
include $(E2E_TEST_DIR)/sharding.mk
include $(E2E_TEST_DIR)/sqlproxy-ssl.mk
include $(E2E_TEST_DIR)/system-variables.mk
include $(E2E_TEST_DIR)/views.mk
