
test-basic-sample: build-mongosqld run-mongodb _write-initial-docs run-mongosqld _test-schema-available _test-connect-success _test-sample-initial-schema

_test-schema-available:
	$(ENV) TIMEOUT=60 testdata/bin/test-schema-available.sh

test-sample-connect-failure: build-mongosqld run-mongodb restore-data run-mongosqld _test-connect-failure

test-schema-unavailable: EXPECTED_ERROR := ERROR 1043 (08S01): MongoDB schema not yet available
test-schema-unavailable: test-sample-connect-failure

_write-initial-schema:
	$(ENV) GENERATION=0 testdata/bin/write-schema.sh

_write-updated-schema:
	$(ENV) GENERATION=1 testdata/bin/write-schema.sh

test-read-schema: build-mongosqld run-mongodb _write-initial-docs _write-initial-schema run-mongosqld _test-schema-available _test-connect-success _test-read-schema
_test-read-schema: TABLE := sample_test
_test-read-schema: NUM_COLUMNS := 1
_test-read-schema: _test-count-columns

_test-read-updated-schema: TABLE := sample_test
_test-read-updated-schema: NUM_COLUMNS := 2
_test-read-updated-schema: _test-count-columns

_write-initial-docs:
	$(ENV) NUM_DOCS=10 testdata/bin/write-sample-docs.sh

_test-sample-initial-schema: TABLE := sample_test
_test-sample-initial-schema: NUM_COLUMNS := 11
_test-sample-initial-schema: _test-count-columns

_sleep-ten:
	sleep 10

_write-updated-docs:
	$(ENV) NUM_DOCS=20 testdata/bin/write-sample-docs.sh

_test-sample-updated-schema: TABLE := sample_test
_test-sample-updated-schema: NUM_COLUMNS := 21
_test-sample-updated-schema: _test-count-columns

# test that basic schema reading works fine
test-read-simple: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/clustered
test-read-simple: test-read-schema

# test that reading works fine with ssl
test-read-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/clustered,mongo/ssl/basic,sqlproxy/mongo-ssl/enabled
test-read-ssl: test-read-schema

# test that reading works fine with auth
test-read-auth: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/clustered,mongo/auth,sqlproxy/auth,sqlproxy/schema/creds,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require
test-read-auth: test-read-schema

# test that read-only mongosqlds get an updated schema for each new connection
test-read-updated: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/clustered
test-read-updated: build-mongosqld run-mongodb restore-data _write-initial-schema run-mongosqld _test-connect-success _write-updated-schema _test-read-updated-schema

# test that basic sampling works fine
test-sample-simple: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic
test-sample-simple: test-basic-sample

test-sample-updated: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/interval-2
test-sample-updated: build-mongosqld run-mongodb _write-initial-docs run-mongosqld _test-schema-available _test-connect-success _write-updated-docs _sleep-ten _test-sample-updated-schema

# we should be able to connect when all namespaces are empty, but nothing should be created
test-sample-empty: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic
test-sample-empty: build-mongosqld run-mongodb run-mongosqld _test-connect-success _test-sample-empty
_test-sample-empty: NUM_DBS :=
_test-sample-empty: _test-count-dbs

# test that sampling works fine with ssl
test-sample-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,mongo/ssl/basic,sqlproxy/mongo-ssl/enabled
test-sample-ssl: test-basic-sample

# when there's an ssl problem, we expect connections to fail before even looking for a schema
test-sample-ssl-failure: EXPECTED_ERROR := ERROR 1429 (HY000): Unable to connect to foreign data source: MongoDB
test-sample-ssl-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,mongo/ssl/basic
test-sample-ssl-failure: test-sample-connect-failure

# test that sampling works fine with auth
test-sample-auth: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,mongo/auth,sqlproxy/auth,sqlproxy/schema/creds,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require
test-sample-auth: test-basic-sample

# when there's an auth problem, sqlproxy should give a schema-unavailable error
test-sample-auth-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,mongo/auth
test-sample-auth-failure: test-schema-unavailable

# when there are multiple schema versions available, make sure we use the one with the highest generation
test-read-most-recent: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/clustered
test-read-most-recent: build-mongosqld run-mongodb restore-data _write-initial-schema _write-updated-schema run-mongosqld _test-connect-success _test-read-updated-schema

# even if we sampled the first schema, we should use a stored schema when one becomes available
test-read-after-sampling: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic
test-read-after-sampling: test-basic-sample _write-initial-schema _test-read-schema

_test-mysql-query:
	$(ENV) QUERY="$(QUERY)" EXPECTED="$(EXPECTED)" testdata/bin/test-mysql-query.sh

_test-count-columns: QUERY = select count(*) from information_schema.columns where table_name = '$(TABLE)';
_test-count-columns: EXPECTED = $(NUM_COLUMNS)
_test-count-columns: _test-mysql-query

_test-count-tables: QUERY = select count(distinct table_name) from information_schema.columns where table_schema != 'mysql' and table_schema != 'information_schema' and table_schema = '$(DB)';
_test-count-tables: EXPECTED = $(NUM_TABLES)
_test-count-tables: _test-mysql-query

_test-count-dbs: QUERY = select count(distinct table_schema) from information_schema.columns where table_schema != 'mysql' and table_schema != 'information_schema';
_test-count-dbs: EXPECTED = $(NUM_DBS)
_test-count-dbs: _test-mysql-query

_insert-sample-docs:
	$(ENV) NUM_DOCS='$(NUM_DOCS)' testdata/bin/write-sample-docs.sh

# for the tests below, our collection has 1001 documents, each with two fields (_id and some number between 0 and 1000).
# when we sample n documents, we expect n+1 columns in the resulting schema (_id plus the unique column from each sampled doc)

test-count-columns: NUM_DOCS := 1001
test-count-columns: build-mongosqld run-mongodb _insert-sample-docs run-mongosqld _test-schema-available _test-count-columns

test-sample-size-default: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic
test-sample-size-default: TABLE := sample_test
test-sample-size-default: NUM_COLUMNS := 1001
test-sample-size-default: test-count-columns

test-sample-size-all: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/sample-all
test-sample-size-all: TABLE := sample_test
test-sample-size-all: NUM_COLUMNS := 1002
test-sample-size-all: test-count-columns

test-sample-size-ten: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/sample-10
test-sample-size-ten: TABLE := sample_test
test-sample-size-ten: NUM_COLUMNS := 11
test-sample-size-ten: test-count-columns

test-flush-new-collection: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic
test-flush-new-collection: build-mongosqld run-mongodb run-mongosqld _test-schema-available _test-connect-success _write-initial-docs _test-flush-and-count
_test-flush-and-count: QUERY := flush sample,, select count(*) from information_schema.tables where table_name = 'sample_test'
_test-flush-and-count: EXPECTED := 1
_test-flush-and-count: _test-mysql-query