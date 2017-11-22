
# Test that both queries and connections are killable.
test-kill-queries: build-mongosqld run-mongodb _restore-data run-mongosqld _test-kill

# Test that killing queries works with ssl enabled.
test-kill-queries-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/mongo-ssl/enabled,mongo/ssl/basic
test-kill-queries-ssl: build-mongosqld run-mongodb _restore-data run-mongosqld _test-kill

# Test that killing queries works with ssl and auth enabled.
test-kill-queries-auth: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require
test-kill-queries-auth: build-mongosqld run-mongodb _restore-data run-mongosqld _test-kill

# Test that killing queries from a different user does not work.
test-kill-queries-wrong-user: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require,client/auth/user2_creds
test-kill-queries-wrong-user: build-mongosqld run-mongodb _restore-data run-mongosqld _test-kill

# Test killing queries in 3.2, 3.4, 3.6, and latest
test-kill-queries-32: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/version/3.2
test-kill-queries-32: build-mongosqld run-mongodb _restore-data run-mongosqld _test-kill

test-kill-queries-34: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/version/3.4
test-kill-queries-34: build-mongosqld run-mongodb _restore-data run-mongosqld _test-kill

test-kill-queries-36: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/version/3.6
test-kill-queries-36: build-mongosqld run-mongodb _restore-data run-mongosqld _test-kill

test-kill-queries-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/version/latest
test-kill-queries-latest: build-mongosqld run-mongodb _restore-data run-mongosqld _test-kill

_restore-data: SUITE := tableau
_restore-data: restore-data _create-test-user

_create-test-user: $(INFRASTRUCTURE_CONFIG) := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require
_create-test-user:
	$(ENV) testdata/bin/create-user.sh

_test-kill: ITERATIONS := 5
_test-kill: PROCS := 5
_test-kill:
	$(ENV) QUERY="select sleep(5)" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_ERROR="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="false" testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select sleep(5)" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_ERROR="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="true" testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select a._id,b.flight_number from tableau.attendees as a inner join tableau.flights201406 as b on a.airport_id = b.origin_airport_id" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_ERROR="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="false" testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select a._id,b.flight_number from tableau.attendees as a inner join tableau.flights201406 as b on a.airport_id = b.origin_airport_id" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_ERROR="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="true" testdata/bin/test-kill-query.sh
