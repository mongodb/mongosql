
# Test that both queries and connections are killable.
test-kill-queries: build-mongosqld run-mongodb run-mongosqld _restore-data _test-kill

# Test that killing queries works with ssl enabled.
test-kill-queries-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/mongo-ssl/enabled,mongo/ssl/basic
test-kill-queries-ssl: test-kill-queries

_test-query-after-kill-success:
	$(ENV) testdata/bin/test-query-after-kill.sh

# Test that when we kill a mysql that the connection stays valid.
test-query-after-kill-success: build-mongosqld run-mongodb run-mongosqld _test-query-after-kill-success

# Test that killing queries works with ssl and auth enabled.
test-kill-queries-auth: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require,client/ssl/pem
test-kill-queries-auth: test-kill-queries

# Test that killing queries from a different user does not work.
test-kill-queries-wrong-user: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require,mongo/other-user/read-tableau
test-kill-queries-wrong-user: test-kill-queries

# Test killing queries in 3.2, 3.4, 3.6, 4.0, and latest
test-kill-queries-3.2: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/version/3.2
test-kill-queries-3.2: test-kill-queries

test-kill-queries-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/version/3.4
test-kill-queries-3.4: test-kill-queries

test-kill-queries-3.6: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/version/3.6
test-kill-queries-3.6: test-kill-queries

test-kill-queries-4.0: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/version/4.0
test-kill-queries-4.0: test-kill-queries

test-kill-queries-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/version/latest
test-kill-queries-latest: test-kill-queries

# Test killing queries on a sharded cluster.
test-kill-queries-sharded: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/sharded-cluster
test-kill-queries-sharded: DATABASE := tableau
test-kill-queries-sharded: COLLECTION := flights201406
test-kill-queries-sharded: SHARD_KEY := {origin_airport_code: 1}
test-kill-queries-sharded: build-mongosqld run-mongodb _shard-collection run-mongosqld _restore-data _test-kill

# Test killing queries on a replica set.
test-kill-queries-replica-set: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/replica-set
test-kill-queries-replica-set: build-mongosqld run-mongodb run-mongosqld _restore-data _test-kill

_restore-data: SUITE := tableau
_restore-data: restore-data _create-test-user

_create-test-user: $(INFRASTRUCTURE_CONFIG) := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require
_create-test-user:
	$(ENV) MECHANISM="$(MECHANISM)" testdata/bin/create-user.sh

_test-kill: ITERATIONS := 5
_test-kill: PROCS := 5
_test-kill:
	$(ENV) QUERY="select sleep(60)" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_ERROR="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="false" testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select sleep(60)" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_ERROR="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="true" testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select a._id,b.airport_code from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_ERROR="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="false" testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select a._id,b.airport_code from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_ERROR="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="true" testdata/bin/test-kill-query.sh
