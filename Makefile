VARIANT := $(VARIANT)

INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG)
ifeq ($(INFRASTRUCTURE_CONFIG),)
INFRASTRUCTURE_CONFIG := default
endif

ENV = VARIANT=$(VARIANT) INFRASTRUCTURE_CONFIG=$(INFRASTRUCTURE_CONFIG)
EXPECTED = EXPECTED_STATUS=$(EXPECTED_STATUS) EXPECTED_ERROR="$(EXPECTED_ERROR)"

SCHEMA_UNAVAILABLE_ERROR = ERROR 1043 (08S01): MongoDB schema not yet available; initial schema sampling still in progress

default: test

benchmark: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/in-memory,sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered,sqlproxy/schema/write,sqlproxy/schema/enable-alter
benchmark: start-all _benchmark _parse-benchmarks

benchmark-tpch: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/in-memory,sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered,sqlproxy/schema/write,sqlproxy/schema/enable-alter
benchmark-tpch: start-all _benchmark-tpch _parse-benchmarks


benchmark-evaluator: _benchmark-evaluator _parse-unit-benchmarks
_benchmark-evaluator:
	$(ENV) PACKAGE='evaluator' testdata/bin/run-unit-benchmarks.sh

_benchmark:
	$(ENV) TYPE="queries|overhead" testdata/bin/run-benchmarks.sh

_benchmark-tpch:
	$(ENV) TYPE="tpch-micro" testdata/bin/run-benchmarks.sh

_parse-benchmarks:
	$(ENV) testdata/bin/parse-benchmark-results.sh

_parse-unit-benchmarks:
	$(ENV) TYPE='unit' testdata/bin/parse-benchmark-results.sh

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

clean:
	$(ENV) testdata/bin/reset-testing-state.sh

download-data:
	testdata/bin/download-blackbox-data.sh
	testdata/bin/download-tableau-data.sh
	testdata/bin/download-tpch-data.sh

restore-data:
	$(ENV) SUITE="$(SUITE)" testdata/bin/restore-test-data.sh

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

test: test-unit test-integration

test-connect-failure: start-all _test-connect-failure
_test-connect-failure: EXPECTED_STATUS = 1
_test-connect-failure:
	$(ENV) $(EXPECTED) testdata/bin/test-simple-connect.sh

test-connect-success: start-all _test-connect-success
_test-connect-success: EXPECTED_STATUS = 0
_test-connect-success:
	$(ENV) $(EXPECTED) testdata/bin/test-simple-connect.sh


test-integration: SUITE := internal
test-integration: test-connect-success restore-data _test-integration
_test-integration:
	$(ENV) SUITE="$(SUITE)" testdata/bin/run-integration-tests.sh

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

test-unit: test-connect-success
	$(ENV) testdata/bin/run-unit-tests.sh

# include config test targets
include testdata/config/tests/alter.mk
include testdata/config/tests/auth.mk
include testdata/config/tests/cleartext-auth.mk
include testdata/config/tests/gssapi.mk
include testdata/config/tests/kill-query.mk
include testdata/config/tests/log-newlines.mk
include testdata/config/tests/log-rotation.mk
include testdata/config/tests/mongo-ssl.mk
include testdata/config/tests/mongo-uri.mk
include testdata/config/tests/mongodrdl.mk
include testdata/config/tests/schema.mk
include testdata/config/tests/server.mk
include testdata/config/tests/sharding.mk
include testdata/config/tests/sqlproxy-ssl.mk
include testdata/config/tests/system-variables.mk
include testdata/config/tests/views.mk
