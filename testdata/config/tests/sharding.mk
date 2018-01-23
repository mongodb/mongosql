#testing against a sharded collection
test-sharded-collection-3.2: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster,mongo/version/3.2
test-sharded-collection-3.2: test-sharded-collection

test-sharded-collection-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster,mongo/version/3.4
test-sharded-collection-3.4: test-sharded-collection


test-sharded-collection-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster
test-sharded-collection-latest: test-sharded-collection

test-sharded-collection: build-mongosqld run-mongodb restore-data _create-sharded-collection run-mongosqld _test-query-against-sharded

_shard-collection:
	$(ENV) DATABASE="$(DATABASE)" COLLECTION="$(COLLECTION)" SHARD_KEY="$(SHARD_KEY)" testdata/bin/shard-collection.sh

_create-sharded-collection: DATABASE := join_test
_create-sharded-collection: COLLECTION := join_1
_create-sharded-collection: _shard-collection

_test-query-against-sharded: QUERY := select count(*) from join_test.bar left join join_test.foo on bar.id=foo.id;
_test-query-against-sharded: EXPECTED := 3
_test-query-against-sharded: _test-mysql-query
