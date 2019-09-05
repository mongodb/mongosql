
# Test that both queries and connections are killable.
test-kill-queries: build-mongosqld run-mongodb run-mongosqld _restore-data _test-kill

# Test that both queries and connections are killable.
# This should be temporary until BI-2332 is completed. It is only used by
# test-kill-queries-admin-user because that test reliably fails for the last
# two queries in _test-kill. This one calls _test-kill-admin which doesn't
# include those queries. Ideally, BI-2332 will address the underlying issue
# and test-kill-queries-admin-user can go back to using test-kill-queries and
# this can be deleted.
test-kill-queries-admin: build-mongosqld run-mongodb run-mongosqld _restore-data _test-kill-admin

# Test that queries continue if a kill fails.
test-kill-queries-continue: build-mongosqld run-mongodb run-mongosqld _restore-data _test-kill-continue

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

# Test that killing queries from a different user does not work
test-kill-queries-wrong-user-continue-running: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require,mongo/other-user/read-tableau
test-kill-queries-wrong-user-continue-running: EXPECTED_CODE := "0"
test-kill-queries-wrong-user-continue-running: DO_NOT_KILL := "true"
test-kill-queries-wrong-user-continue-running: ITERATIONS := 2
test-kill-queries-wrong-user-continue-running: PROCS := 3
test-kill-queries-wrong-user-continue-running: test-kill-queries-continue

# Test that killing any queries as an admin works.
#
# The INFRASTRUCTURE_CONFIG sets an admin user (mongo/auth) as "user 1"
# and a non-admin user (mongo/other-user/read-tableau) as "user 2". This is
# necessary so the non-admin user can be created by the admin. In this test
# the users need to be swapped so that the non-admin issues the queries and
# the admin issues the kills (successfully).
# The test-kill-query.sh script has default values for all of the variables
# defined below; in this test they are set for the following reasons:
# EXPECTED_KILL_CODE - The expected exit code of the kill command issued by
#                      "user 2" (the admin) is 0 (success).
# SWAP_USERS         - The users need to be swapped after they are created.
# KILLING_USER       - The user to issue the "kill" is "user 2" (which will
#                      be the admin user since the users are swapped).
# PROCS              - Run only 1 process per iteration for this test,
#                      otherwise the results are flaky.
test-kill-queries-admin-user: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require,mongo/other-user/read-tableau
test-kill-queries-admin-user: EXPECTED_KILL_CODE := "0"
test-kill-queries-admin-user: SWAP_USERS := "true"
test-kill-queries-admin-user: KILLING_USER := "2"
test-kill-queries-admin-user: PROCS := 1
test-kill-queries-admin-user: test-kill-queries-admin

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

_test-kill:
	$(ENV) QUERY="select sleep(60)" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="false" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select sleep(60)" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="true" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select a._id,b.airport_code from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="false" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select a._id,b.airport_code from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="true" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select a._id,b.airport_code from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code order by b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="false" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select a._id,b.airport_code from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code order by b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="true" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) testdata/bin/test-kill-query.sh

_test-kill-admin:
	$(ENV) QUERY="select sleep(60)" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="false" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select sleep(60)" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="true" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select a._id,b.airport_code from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="false" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select a._id,b.airport_code from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="ERROR 1317 (70100) at line 1: Query execution was interrupted" KILL_CONN="true" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) testdata/bin/test-kill-query.sh

_test-kill-continue:
	$(ENV) QUERY="select count(sleep(60))" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="1" KILL_CONN="false" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) EXPECTED_CODE=$(EXPECTED_CODE) DO_NOT_KILL=$(DO_NOT_KILL) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select count(sleep(60))" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="1" KILL_CONN="true" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) EXPECTED_CODE=$(EXPECTED_CODE) DO_NOT_KILL=$(DO_NOT_KILL) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select count(a._id) from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="13050155" KILL_CONN="false" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) EXPECTED_CODE=$(EXPECTED_CODE) DO_NOT_KILL=$(DO_NOT_KILL) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select count(a._id) from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="13050155" KILL_CONN="true" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) EXPECTED_CODE=$(EXPECTED_CODE) DO_NOT_KILL=$(DO_NOT_KILL) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select count(a._id) from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code order by b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="13050155" KILL_CONN="false" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) EXPECTED_CODE=$(EXPECTED_CODE) DO_NOT_KILL=$(DO_NOT_KILL) testdata/bin/test-kill-query.sh
	$(ENV) QUERY="select count(a._id) from tableau.flights201406 as a inner join tableau.attendees as b on a.origin_airport_code = b.airport_code order by b.airport_code" PROCS="$(PROCS)" ITERATIONS="$(ITERATIONS)" EXPECTED_RESULT="13050155" KILL_CONN="true" EXPECTED_KILL_CODE=$(EXPECTED_KILL_CODE) SWAP_USERS=$(SWAP_USERS) KILLING_USER=$(KILLING_USER) EXPECTED_CODE=$(EXPECTED_CODE) DO_NOT_KILL=$(DO_NOT_KILL) testdata/bin/test-kill-query.sh

