VARIANT := $(VARIANT)
INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG)
ENV = VARIANT=$(VARIANT) INFRASTRUCTURE_CONFIG=$(INFRASTRUCTURE_CONFIG)
EXPECTED = EXPECTED_STATUS=$(EXPECTED_STATUS) EXPECTED_ERROR="$(EXPECTED_ERROR)"

default: test

build: build-mongodrdl build-mongosqld

build-mongosqld:
	$(ENV) testdata/bin/build-mongosqld.sh

build-mongodrdl:
	$(ENV) testdata/bin/build-mongodrdl.sh

clean:
	rm -f integration_test.go
	$(ENV) testdata/bin/reset-testing-state.sh

generate:
	$(ENV) testdata/bin/generate-tests.sh

restore-data: generate
	$(ENV) testdata/bin/restore-test-data.sh

run-mongosqld:
	$(ENV) testdata/bin/start-mongosqld.sh

run-mongodb:
	$(ENV) testdata/bin/start-orchestration.sh

shell:
	mysql -P3307

start-all: build-mongosqld run-mongosqld run-mongodb

test: test-unit test-integration

test-start-mongosqld: build-mongosqld _test-start-mongosqld
_test-start-mongosqld:
	$(ENV) $(EXPECTED) testdata/bin/test-start-mongosqld.sh

test-connect-failure: EXPECTED_STATUS = 1
test-connect-failure: start-all
	$(ENV) $(EXPECTED) testdata/bin/test-simple-connect.sh

test-connect-success: EXPECTED_STATUS = 0
test-connect-success: start-all
	$(ENV) $(EXPECTED) testdata/bin/test-simple-connect.sh

test-integration: test-connect-success restore-data
	$(ENV) testdata/bin/run-integration-tests.sh

test-unit: clean
	$(ENV) testdata/bin/run-unit-tests.sh

# include config test targets
include testdata/config/tests/cleartext-auth.mk
include testdata/config/tests/log-rotation.mk
include testdata/config/tests/mongo-ssl.mk
include testdata/config/tests/mongodrdl.mk
include testdata/config/tests/sqlproxy-ssl.mk
