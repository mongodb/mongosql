VARIANT := $(VARIANT)

INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG)
ifeq ($(INFRASTRUCTURE_CONFIG),)
INFRASTRUCTURE_CONFIG := default
endif

ENV = VARIANT=$(VARIANT) INFRASTRUCTURE_CONFIG=$(INFRASTRUCTURE_CONFIG)
EXPECTED = EXPECTED_STATUS=$(EXPECTED_STATUS) EXPECTED_ERROR="$(EXPECTED_ERROR)"

default: test

build: build-mongodrdl build-mongosqld

build-mongodrdl:
	$(ENV) testdata/bin/build-mongodrdl.sh

build-mongosqld:
	$(ENV) testdata/bin/build-mongosqld.sh

check-races:
	testdata/bin/check-races.sh

clean:
	rm -f integration_test.go
	$(ENV) testdata/bin/reset-testing-state.sh

generate:
	$(ENV) testdata/bin/generate-tests.sh

restore-data: generate
	$(ENV) testdata/bin/restore-test-data.sh

run-mongodb:
	$(ENV) testdata/bin/start-orchestration.sh

run-mongosqld:
	$(ENV) testdata/bin/start-mongosqld.sh

shell:
	mysql -P3307

start-all: build-mongosqld run-mongosqld run-mongodb

test: test-unit test-integration

test-connect-failure: start-all _test-connect-failure
_test-connect-failure: EXPECTED_STATUS = 1
_test-connect-failure:
	$(ENV) $(EXPECTED) testdata/bin/test-simple-connect.sh

test-connect-success: start-all _test-connect-success
_test-connect-success: EXPECTED_STATUS = 0
_test-connect-success:
	$(ENV) $(EXPECTED) testdata/bin/test-simple-connect.sh

test-integration: test-connect-success restore-data
	$(ENV) testdata/bin/run-integration-tests.sh

test-option-help: build-mongosqld
	$(ARTIFACTS_DIR)/bin/mongosqld --help

test-option-version: build-mongosqld
	$(ARTIFACTS_DIR)/bin/mongosqld --version

test-start-mongosqld: build-mongosqld _test-start-mongosqld
_test-start-mongosqld:
	$(ENV) $(EXPECTED) testdata/bin/test-start-mongosqld.sh

test-unit: test-connect-success
	$(ENV) testdata/bin/run-unit-tests.sh

# include config test targets
include testdata/config/tests/alter.mk
include testdata/config/tests/cleartext-auth.mk
include testdata/config/tests/log-newlines.mk
include testdata/config/tests/log-rotation.mk
include testdata/config/tests/mongo-ssl.mk
include testdata/config/tests/mongodrdl.mk
include testdata/config/tests/schema.mk
include testdata/config/tests/sqlproxy-ssl.mk
