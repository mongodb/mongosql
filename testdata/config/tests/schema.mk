
test-count-columns: build-mongosqld run-mongodb _insert-sample-docs run-mongosqld _test-count-columns
_test-count-columns:
	$(ENV) EXPECTED_NUM_COLUMNS="$(NUM_COLUMNS)" testdata/bin/test-count-columns.sh

_insert-sample-docs:
_insert-sample-docs:
	$(ARTIFACTS_DIR)/mongodb/bin/mongo --eval 'for(i=0;i<1001;i++){ doc={}; doc[i]=true; db.column_count.insert(doc); };' > /dev/null 2>&1

# our collection has 1001 documents, each with two fields (_id and some number between 0 and 1000).
# when we sample n documents, we expect n+1 columns in the resulting schema (_id plus the unique column from each sampled doc)

test-sample-default: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic
test-sample-default: NUM_COLUMNS = 1001
test-sample-default: test-count-columns

test-sample-all: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/sample-all
test-sample-all: NUM_COLUMNS = 1002
test-sample-all: test-count-columns

test-sample-ten: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/dynamic,sqlproxy/schema/sample-10
test-sample-ten: NUM_COLUMNS = 11
test-sample-ten: test-count-columns
