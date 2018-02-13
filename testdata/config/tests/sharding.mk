# testing joins against sharded environments
test-join-sharded-collection-3.2: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster,mongo/version/3.2
test-join-sharded-collection-3.2: test-join-on-sharded-collection

test-join-sharded-collection-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster,mongo/version/3.4
test-join-sharded-collection-3.4: test-join-on-sharded-collection

test-join-sharded-collection-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster
test-join-sharded-collection-latest: test-join-on-sharded-collection

test-join-sharded-view-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster,mongo/version/3.4
test-join-sharded-view-3.4: test-join-on-sharded-view

test-join-sharded-view-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster
test-join-sharded-view-latest: test-join-on-sharded-view

test-join-on-sharded-collection: build-mongosqld run-mongodb restore-integration-data _create-sharded-collection run-mongosqld _test-join-against-sharded
test-join-on-sharded-view: build-mongosqld run-mongodb restore-integration-data _create-sharded-collection _create-sharded-view-for-join run-mongosqld _test-join-against-sharded-view

_shard-collection:
	$(ENV) DATABASE="$(DATABASE)" COLLECTION="$(COLLECTION)" SHARD_KEY="$(SHARD_KEY)" testdata/bin/shard-collection.sh

_create-sharded-collection: DATABASE := join_test
_create-sharded-collection: COLLECTION := join_1
_create-sharded-collection: _shard-collection

_create-sharded-view-for-join: DATABASE := join_test
_create-sharded-view-for-join: SOURCE := join_1
_create-sharded-view-for-join: VIEW := join_10
_create-sharded-view-for-join: _create-view

_test-join-against-sharded: QUERY := select count(*) from join_test.bar left join join_test.foo on bar.id=foo.id;
_test-join-against-sharded: EXPECTED := 3
_test-join-against-sharded: _test-mysql-query

_test-join-against-sharded-view: QUERY := select count(*) from join_test.bar left join join_test.foo_view on bar.id=foo_view.id;
_test-join-against-sharded-view: EXPECTED := 3
_test-join-against-sharded-view: _test-mysql-query

# testing counts against sharded environments
test-count-on-sharded-collection-3.2: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster,mongo/version/3.2
test-count-on-sharded-collection-3.2: test-count-on-sharded-collection

test-count-on-sharded-collection-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster,mongo/version/3.4
test-count-on-sharded-collection-3.4: test-count-on-sharded-collection

test-count-on-sharded-collection-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster
test-count-on-sharded-collection-latest: test-count-on-sharded-collection

test-count-on-sharded-view-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster,mongo/version/3.4
test-count-on-sharded-view-3.4: test-count-on-sharded-view

test-count-on-sharded-view-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster
test-count-on-sharded-view-latest: test-count-on-sharded-view

test-count-on-sharded-view-on-view-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster,mongo/version/3.4
test-count-on-sharded-view-on-view-3.4: test-count-on-sharded-view-on-view

test-count-on-sharded-view-on-view-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster
test-count-on-sharded-view-on-view-latest: test-count-on-sharded-view-on-view

test-count-on-sharded-collection: build-mongosqld run-mongodb restore-integration-data _create-sharded-collection-for-count run-mongosqld _test-count-star-query-against-sharded
test-count-on-sharded-view: build-mongosqld run-mongodb restore-integration-data _create-sharded-collection-for-count _create-sharded-view run-mongosqld _test-count-star-query-against-sharded-view
test-count-on-sharded-view-on-view: build-mongosqld run-mongodb restore-integration-data _create-sharded-collection-for-count _create-sharded-view-on-view run-mongosqld _test-count-star-query-against-sharded-view-on-view

_create-sharded-collection-for-count: DATABASE := select_test
_create-sharded-collection-for-count: COLLECTION := select_7
_create-sharded-collection-for-count: _shard-collection

_create-sharded-view: DATABASE := select_test
_create-sharded-view: SOURCE := select_7
_create-sharded-view: VIEW := select_8
_create-sharded-view: _create-view

_create-sharded-view-on-view: DATABASE := select_test
_create-sharded-view-on-view: SOURCE1 := select_7
_create-sharded-view-on-view: VIEW1 := select_8
_create-sharded-view-on-view: SOURCE2 := select_8
_create-sharded-view-on-view: VIEW2 := select_9
_create-sharded-view-on-view: _create-view-on-view

_test-count-star-query-against-sharded: QUERY := select count(*) from select_test.foo_count;
_test-count-star-query-against-sharded: EXPECTED := 7
_test-count-star-query-against-sharded: _test-mysql-query

_test-count-star-query-against-sharded-view: QUERY := select count(*) from select_test.foo_view;
_test-count-star-query-against-sharded-view: EXPECTED := 7
_test-count-star-query-against-sharded-view: _test-mysql-query

_test-count-star-query-against-sharded-view-on-view: QUERY := select count(*) from select_test.foo_view_view;
_test-count-star-query-against-sharded-view-on-view: EXPECTED := 7
_test-count-star-query-against-sharded-view-on-view: _test-mysql-query
