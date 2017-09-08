
run-mongosqld-gssapi: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/test
run-mongosqld-gssapi: run-mongosqld

test-gssapi: build-mongosqld run-mongosqld-gssapi
	$(ENV) testdata/bin/run-gssapi-auth-tests.sh
	