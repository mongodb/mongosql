setup-sample: build-mongosqld run-mongodb _write-initial-docs run-mongosqld _test-schema-available
test-basic-sample: setup-sample _test-sample-initial-schema

_test-schema-available:
	$(ENV) TIMEOUT=120 testdata/bin/test-schema-available.sh

test-sample-connect-failure: build-mongosqld run-mongodb _write-initial-docs run-mongosqld _test-connect-failure
test-sample-connect-success: build-mongosqld run-mongodb _write-initial-docs run-mongosqld _test-connect-success

test-schema-available: test-sample-connect-success

test-schema-unavailable: EXPECTED_ERROR := ERROR 1043 (08S01): MongoDB schema not yet available
test-schema-unavailable: test-sample-connect-failure

_write-initial-schema:
	$(ENV) GENERATION=0 testdata/bin/write-schema.sh

_write-v1-schema:
	$(ENV) GENERATION=0 PROTOCOL=v1 testdata/bin/write-schema.sh

_write-mixed-case-document:
	testdata/bin/write-mixed-case-document.sh

_write-polymorphic-data:
	testdata/bin/write-polymorphic-data.sh

_write-updated-schema:
	$(ENV) GENERATION=1 testdata/bin/write-schema.sh

test-read-v1-schema: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered
test-read-v1-schema: build-mongosqld run-mongodb _write-initial-docs _write-v1-schema run-mongosqld _test-connect-success _test-read-schema

test-read-schema: build-mongosqld run-mongodb _write-initial-docs _write-initial-schema run-mongosqld _test-schema-available _test-connect-success _test-read-schema
_test-read-schema: TABLE := sample_test
_test-read-schema: NUM_COLUMNS := 2
_test-read-schema: _test-count-columns

test-sample-without-prejoin: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-sample-without-prejoin: build-mongosqld run-mongodb _write-array-doc-2d-index run-mongosqld _test-schema-available _test-connect-success _test-sample-without-prejoin
_test-sample-without-prejoin: TABLE := sample_test_grades
_test-sample-without-prejoin: NUM_COLUMNS := 5
_test-sample-without-prejoin: _test-count-columns

test-sample-with-prejoin: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/prejoin
test-sample-with-prejoin: build-mongosqld run-mongodb _write-array-doc-2d-index run-mongosqld _test-schema-available _test-connect-success _test-sample-with-prejoin
_test-sample-with-prejoin: TABLE := sample_test_grades
_test-sample-with-prejoin: NUM_COLUMNS := 12
_test-sample-with-prejoin: _test-count-columns

_test-read-updated-schema: TABLE := sample_test
_test-read-updated-schema: NUM_COLUMNS := 3
_test-read-updated-schema: _test-count-columns

_write-initial-docs:
	$(ENV) NUM_DOCS=10 testdata/bin/write-sample-docs.sh

_test-sample-initial-schema: TABLE := sample_test
_test-sample-initial-schema: NUM_COLUMNS := 11
_test-sample-initial-schema: _test-count-columns

_write-array-doc-2d-index:
	$(ENV) INDEX_TYPE="2d" testdata/bin/write-array-doc.sh

_write-array-doc-2dsphere-index:
	$(ENV) INDEX_TYPE="2dsphere" testdata/bin/write-array-doc.sh

_write-nested-array-doc-2d-index:
	$(ENV) INDEX_TYPE="2d" NESTED="1" testdata/bin/write-array-doc.sh

_write-nested-array-doc-2dsphere-index:
	$(ENV) INDEX_TYPE="2dsphere" NESTED="1" testdata/bin/write-array-doc.sh

test-sample-array-2d-index: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-sample-array-2d-index: build-mongosqld run-mongodb _write-array-doc-2d-index run-mongosqld _test-schema-available _test-connect-success _test-db-table-count

test-sample-array-2dsphere-index: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-sample-array-2dsphere-index: build-mongosqld run-mongodb _write-array-doc-2dsphere-index run-mongosqld _test-schema-available _test-connect-success _test-db-table-count

_test-db-table-count: DB := mongosqld_sample_test
_test-db-table-count: TABLE := sample_test
_test-db-table-count: NUM_TABLES := 2
_test-db-table-count: _test-count-tables

test-sample-nested-array-2d-index: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-sample-nested-array-2d-index: build-mongosqld run-mongodb _write-nested-array-doc-2d-index run-mongosqld _test-schema-available _test-connect-success _test-db-table-count

test-sample-nested-array-2dsphere-index: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-sample-nested-array-2dsphere-index: build-mongosqld run-mongodb _write-nested-array-doc-2dsphere-index run-mongosqld _test-schema-available _test-connect-success _test-db-table-count

_test-db-table-count: DB := mongosqld_sample_test
_test-db-table-count: TABLE := sample_test
_test-db-table-count: NUM_TABLES := 2
_test-db-table-count: _test-count-tables

_sleep-ten:
	sleep 10

_sleep-twenty:
	sleep 20

_write-updated-docs:
	$(ENV) NUM_DOCS=20 testdata/bin/write-sample-docs.sh

_test-sample-updated-schema: TABLE := sample_test
_test-sample-updated-schema: NUM_COLUMNS := 21
_test-sample-updated-schema: _test-count-columns

_test-schema-mapping-heuristic-updated: QUERY := set @@global.schema_mapping_heuristic='lattice',,select sleep(5),,select data_type from information_schema.columns where table_schema='mongosqld_sample_test' and table_name='sample_test' and column_name='sample_column'
_test-schema-mapping-heuristic-updated: EXPECTED := varchar
_test-schema-mapping-heuristic-updated: NEW_SHELL_PER_CMD := 1
_test-schema-mapping-heuristic-updated: _test-mysql-query

# test that basic schema reading works fine
test-read-simple: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered
test-read-simple: test-read-schema

# test that reading works fine with ssl
test-read-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered,mongo/ssl/basic,sqlproxy/mongo-ssl/enabled
test-read-ssl: test-read-schema

# test that reading works fine with auth
test-read-auth: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered,mongo/auth,sqlproxy/auth/enabled,sqlproxy/auth/admin-creds,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require
test-read-auth: test-read-schema

# test that read-only mongosqlds get an updated schema for each new connection
test-read-updated: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered
test-read-updated: build-mongosqld run-mongodb restore-data _write-initial-schema run-mongosqld _test-connect-success _write-updated-schema _test-read-updated-schema

# test that basic sampling works fine
test-sample-simple: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-sample-simple: test-basic-sample

_update-sample-size-to-ten:
	$(ENV) EXPECTED_STATUS='0' EXPECTED_ERROR='' NEW_SAMPLE_SIZE=10 testdata/bin/update-sample-size.sh

# Use +1 to account for the _id field 5 + 1 = 6
_test-initial-cols-before-update: TABLE := sample_test
_test-initial-cols-before-update: NUM_COLUMNS := 6
_test-initial-cols-before-update: _test-count-columns

# Use +1 to account for the _id field 10 + 1 = 11
_test-cols-after-update: TABLE := sample_test
_test-cols-after-update: NUM_COLUMNS := 11
_test-cols-after-update: _test-count-columns2

# test that dynamically changing the same size works correctly
test-sample-size-update: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/sample-5
test-sample-size-update: setup-sample _test-initial-cols-before-update _update-sample-size-to-ten _test-flush _test-cols-after-update

_write-initial-5-docs:
	$(ENV) NUM_DOCS=5 testdata/bin/write-sample-docs.sh

_update-sample-refresh-interval-to-three:
	$(ENV) EXPECTED_STATUS='0' EXPECTED_ERROR='' NEW_REFRESH_INTERVAL=3 testdata/bin/update-sample-refresh-interval.sh

_test-sample-refresh-interval-update: build-mongosqld run-mongodb _write-initial-5-docs run-mongosqld _test-initial-cols-before-update
_test-sample-refresh-interval-update: _write-initial-docs _update-sample-refresh-interval-to-three _sleep-twenty _test-cols-after-update

test-sample-refresh-interval-update: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/sample-all
test-sample-refresh-interval-update: _test-sample-refresh-interval-update

test-sample-refresh-interval-update-write: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/sample-all,sqlproxy/schema/clustered,sqlproxy/schema/write
test-sample-refresh-interval-update-write: _test-sample-refresh-interval-update

# If our sample size global system variable was not updated to the given sample
# size of 5 and instead had the default of 1000, this would return 11 columns.
test-sample-size-respects-config: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/sample-5
test-sample-size-respects-config: build-mongosqld run-mongodb _write-initial-docs run-mongosqld
test-sample-size-respects-config: TABLE := sample_test
test-sample-size-respects-config: NUM_COLUMNS := 6
test-sample-size-respects-config: _test-count-columns

# If our refresh interval global system variable was not updated to the given
# interval of 3 seconds then in 20 seconds we should see 11 columns, but if it
# had the default of 0, we'd see no updates and 0 columns.
test-sample-refresh-interval-respects-config: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/sample-all,sqlproxy/schema/refresh-interval-3
test-sample-refresh-interval-respects-config: build-mongosqld run-mongodb run-mongosqld
test-sample-refresh-interval-respects-config: _write-initial-docs _sleep-twenty
test-sample-refresh-interval-respects-config: TABLE := sample_test
test-sample-refresh-interval-respects-config: NUM_COLUMNS := 11
test-sample-refresh-interval-respects-config: _test-count-columns

test-sample-refresh-interval-updates-quickly: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/sample-all,sqlproxy/schema/refresh-interval-10000
test-sample-refresh-interval-updates-quickly: build-mongosqld run-mongodb run-mongosqld
test-sample-refresh-interval-updates-quickly: _write-initial-docs _update-sample-refresh-interval-to-three _sleep-twenty
test-sample-refresh-interval-updates-quickly: TABLE := sample_test
test-sample-refresh-interval-updates-quickly: NUM_COLUMNS := 11
test-sample-refresh-interval-updates-quickly: _test-count-columns


# test that sampling works when customer's data contains document fields with
# mixed case across arrays, nested, and top-level scalar columns.
test-sample-mixed-case-columns: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-sample-mixed-case-columns: build-mongosqld run-mongodb _write-mixed-case-document run-mongosqld _test-connect-success _test-sample-mixed-case-columns
_test-sample-mixed-case-columns: QUERY := select count(*) from information_schema.columns where table_name like '%sample_test%'
_test-sample-mixed-case-columns: EXPECTED := 11
_test-sample-mixed-case-columns: _test-mysql-query

test-geofield-mapping1: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/geo-2darray
test-geofield-mapping1: build-mongosqld run-mongodb run-mongosqld _test-geofield-mapping1
_test-geofield-mapping1: TABLE := base
_test-geofield-mapping1: NUM_COLUMNS := 4
_test-geofield-mapping1: _test-count-columns

test-geofield-mapping2: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/geo-array
test-geofield-mapping2: build-mongosqld run-mongodb run-mongosqld _test-geofield-mapping2
_test-geofield-mapping2: TABLE := base
_test-geofield-mapping2: NUM_COLUMNS := 4
_test-geofield-mapping2: _test-count-columns

test-sample-updated: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/interval-2
test-sample-updated: build-mongosqld run-mongodb _write-initial-docs run-mongosqld _test-schema-available _test-connect-success _write-updated-docs _sleep-ten _test-sample-updated-schema

# we should be able to connect when all namespaces are empty, but nothing should be created
test-sample-empty: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-sample-empty: build-mongosqld run-mongodb run-mongosqld _test-connect-success _test-sample-empty
_test-sample-empty: NUM_DBS :=
_test-sample-empty: _test-count-dbs

# test that sampling works fine with ssl
test-sample-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,mongo/ssl/basic,sqlproxy/mongo-ssl/enabled
test-sample-ssl: test-basic-sample

# when there's an ssl problem, we expect connections to fail before even looking for a schema
test-sample-ssl-failure: EXPECTED_ERROR := ERROR 1429 (HY000): Unable to connect to foreign data source: MongoDB
test-sample-ssl-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,mongo/ssl/basic
test-sample-ssl-failure: test-sample-connect-failure

# test that sampling works fine with auth
test-sample-auth: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,mongo/auth,sqlproxy/auth/enabled,sqlproxy/auth/admin-creds,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require
test-sample-auth: test-basic-sample

# when there's an auth problem in MongoDB versions 3.2, 3.4 and 3.6, sqlproxy should give a schema-unavailable error to clients
test-sample-auth-failure-3.2: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,mongo/auth,mongo/version/3.2
test-sample-auth-failure-3.2: test-schema-unavailable

test-sample-auth-failure-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,mongo/auth,mongo/version/3.4
test-sample-auth-failure-3.4: test-schema-unavailable

test-sample-auth-failure-3.6: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,mongo/auth,mongo/version/3.6
test-sample-auth-failure-3.6: test-schema-unavailable

# when there's an auth problem in MongoDB versions 3.7+, sqlproxy fail to sample the schema
# because the schema is not yet available. This is different from prior mongodb versions
# since 3.7+ requires authentication to list all databases.
test-sample-auth-failure-4.0: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,mongo/auth
test-sample-auth-failure-4.0: test-schema-unavailable

test-sample-auth-failure-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,mongo/auth
test-sample-auth-failure-latest: test-schema-unavailable

# when there are multiple schema versions available, make sure we use the one with the highest generation
test-read-most-recent: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered
test-read-most-recent: build-mongosqld run-mongodb restore-data _write-initial-schema _write-updated-schema run-mongosqld _test-connect-success _test-read-updated-schema

# even if we sampled the first schema, we should use a stored schema when one becomes available
test-read-after-sampling: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-read-after-sampling: test-basic-sample _write-initial-schema _test-read-schema

_write-heuristic-docs:
	$(ENV) testdata/bin/write-schema-mapping-heuristic-docs.sh

# check that the sample heuristic variable works
test-schema-mapping-heuristic-majority: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-lattice
test-schema-mapping-heuristic-majority: SUITE := internal
test-schema-mapping-heuristic-majority: build-mongosqld run-mongodb _write-heuristic-docs run-mongosqld _test-schema-mapping-heuristic-majority
_test-schema-mapping-heuristic-majority:
	$(ENV) QUERY="set @@global.schema_mapping_heuristic='majority'" testdata/bin/test-mysql-query.sh
	$(ENV) QUERY="flush sample; select column_type from information_schema.columns where table_name = 'schema_mapping_heuristics' and column_name='mid'" \
		  EXPECTED="tinyint(1)" testdata/bin/test-mysql-query.sh

# check that the sample heuristic variable works
test-schema-mapping-heuristic-lattice: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-schema-mapping-heuristic-lattice: SUITE := internal
test-schema-mapping-heuristic-lattice: build-mongosqld run-mongodb _write-heuristic-docs run-mongosqld _test-schema-mapping-heuristic-lattice
_test-schema-mapping-heuristic-lattice:
	$(ENV) QUERY="set @@global.schema_mapping_heuristic='lattice'" testdata/bin/test-mysql-query.sh
	$(ENV) QUERY="flush sample; select column_type from information_schema.columns where table_name = 'schema_mapping_heuristics' and column_name='mid'" \
		  EXPECTED="decimal(65,20)" testdata/bin/test-mysql-query.sh

test-schema-mapping-heuristic-updated: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/interval-2,sqlproxy/schema/ns-sample-col
test-schema-mapping-heuristic-updated: build-mongosqld run-mongodb _write-polymorphic-data run-mongosqld _test-schema-available _test-connect-success _test-schema-mapping-heuristic-updated

_test-mysql-query:
	$(ENV) QUERY="$(QUERY)" EXPECTED="$(EXPECTED)" EXPECTED_ERROR="$(EXPECTED_ERROR)" NEW_SHELL_PER_CMD="$(NEW_SHELL_PER_CMD)" testdata/bin/test-mysql-query.sh

_test-mysql-query2:
	$(ENV) QUERY="$(QUERY)" EXPECTED="$(EXPECTED)" EXPECTED_ERROR="$(EXPECTED_ERROR)" NEW_SHELL_PER_CMD="$(NEW_SHELL_PER_CMD)" testdata/bin/test-mysql-query.sh

_test-count-columns: QUERY = select count(*) from information_schema.columns where table_name = '$(TABLE)';
_test-count-columns: EXPECTED = $(NUM_COLUMNS)
_test-count-columns: _test-mysql-query

_test-count-columns2: QUERY = select count(*) from information_schema.columns where table_name = '$(TABLE)';
_test-count-columns2: EXPECTED = $(NUM_COLUMNS)
_test-count-columns2: _test-mysql-query2

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

test-sample-size-default: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-sample-size-default: TABLE := sample_test
test-sample-size-default: NUM_COLUMNS := 1001
test-sample-size-default: test-count-columns

test-sample-size-all: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/sample-all
test-sample-size-all: TABLE := sample_test
test-sample-size-all: NUM_COLUMNS := 1002
test-sample-size-all: test-count-columns

test-sample-size-ten: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/sample-10
test-sample-size-ten: TABLE := sample_test
test-sample-size-ten: NUM_COLUMNS := 11
test-sample-size-ten: test-count-columns

test-flush-new-collection: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority
test-flush-new-collection: build-mongosqld run-mongodb run-mongosqld _test-schema-available _test-connect-success _write-initial-docs _test-flush-and-count
_test-flush-and-count: QUERY := flush sample,, select count(*) from information_schema.tables where table_name = 'sample_test'
_test-flush-and-count: EXPECTED := 1
_test-flush-and-count: _test-mysql-query
