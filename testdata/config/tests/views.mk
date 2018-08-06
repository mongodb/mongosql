
_create-view-test: DATABASE := test
_create-view-test: SOURCE := test2
_create-view-test: VIEW := test1
_create-view-test: _create-view

_create-view:
	$(ENV) DATABASE="$(DATABASE)" VIEW="$(VIEW)" SOURCE="$(SOURCE)" testdata/bin/create-view.sh

_create-view-on-view:
	$(ENV) DATABASE="$(DATABASE)" VIEW1="$(VIEW1)" SOURCE1="$(SOURCE1)" VIEW2="$(VIEW2)" SOURCE2="$(SOURCE2)" testdata/bin/create-view-on-view.sh

test-view: start-all _create-view-test _test-view

_test-view: QUERY := select * from test.test1;
_test-view: EXPECTED :=
_test-view: _test-mysql-query

_create_sampling_view: DATABASE := test
_create_sampling_view: SOURCE1 := sample_test
_create_sampling_view: VIEW1 := view_1
_create_sampling_view: SOURCE2 := view_1
_create_sampling_view: VIEW2 := view_2
_create_sampling_view: _create-view-on-view

test-sample-auth-view-on-collection-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/read-view-only,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/schema/dynamic,sqlproxy/schema/ns-view-only,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-sample-auth-view-on-collection-latest: NUM_DOCS=20
test-sample-auth-view-on-collection-latest: build-mongosqld run-mongodb _insert-sample-docs _create-test-user _create_sampling_view run-mongosqld _test-connect-success
