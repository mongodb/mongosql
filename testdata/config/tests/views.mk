
_create-view:
	$(ENV) testdata/bin/create-view.sh

test-view: start-all _create-view _test-view

_test-view: QUERY := select * from test.test1;
_test-view: EXPECTED :=
_test-view: _test-mysql-query
