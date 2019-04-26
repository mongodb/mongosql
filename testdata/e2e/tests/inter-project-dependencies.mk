
# Test the jdbc authentication plugin.
test-jdbc-auth: build-mongosqld run-mongodb run-mongosqld _restore-internal-data _run-jdbc-auth-test

# Test using scram-sha-1 user.
test-jdbc-auth-scram1: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require
test-jdbc-auth-scram1: test-jdbc-auth

# Test using scram-sha-256 user.
test-jdbc-auth-scram256: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/4.0,mongo/other-user/root,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/auth/scram-sha-256-mechanism,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/other-user-creds-scram-sha-256,client/auth/cleartext,client/ssl/require
test-jdbc-auth-scram256: MECHANISM := SCRAM-SHA-256
test-jdbc-auth-scram256: test-jdbc-auth

# The jdbc auth tests depend on the existence of a "use_test" database and
# a "foo" collection in that database. Restoring the internal data suite
# ensures these exist.
_restore-internal-data: SUITE := internal
_restore-internal-data: restore-data _create-jdbc-user

_create-jdbc-user:
	$(ENV) MECHANISM="$(MECHANISM)" testdata/bin/create-user.sh

_run-jdbc-auth-test:
	$(ENV) testdata/bin/run-jdbc-auth-tests.sh
