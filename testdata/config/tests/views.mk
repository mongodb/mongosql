
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
