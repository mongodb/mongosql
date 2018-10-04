
# test that connecting to an Atlas MongoDB cluster is successful and querying works as expected.
test-atlas-connect-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/auth/atlas-prod-creds,sqlproxy/auth/enabled,sqlproxy/mongo/atlas-prod-host,sqlproxy/mongo-ssl/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/atlas-prod-creds
test-atlas-connect-success: build-mongosqld run-mongosqld _test-connect-success _test-find-city
_test-find-city: QUERY = SELECT agent_attorney_city FROM \`H1B-Visa-Applications\`.year2015 WHERE _id = '572cbdd9d2fc210e7ce696ec'
_test-find-city: EXPECTED = MINNEAPOLIS
_test-find-city: _test-mysql-query

test-alternate-user-connect-failure: start-all _create-test-user _test-connect-failure
test-alternate-user-connect-success: start-all _create-test-user _test-connect-success

# if mongodb has auth enabled, but sqlproxy does not, we expect the connection to be
# accepted as long as the user's schema comes from a drdl file
test-mongo-auth-drdl-no-creds: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth
test-mongo-auth-drdl-no-creds: test-connect-success

# when auth is enabled on mongodb and sqlproxy but the provided admin credentials
# are invalid, we expect the connection to be rejected
test-mongo-auth-drdl-wrong-admin-creds: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/wrong-admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-drdl-wrong-admin-creds: EXPECTED_ERROR := ERROR 1043 (08S01): error retrieving information from MongoDB: failed to create admin session for loading metadata: unable to authenticate conversation 0: unable to authenticate using mechanism \"SCRAM-SHA-1\": (AuthenticationFailed) Authentication failed.
test-mongo-auth-drdl-wrong-admin-creds: test-connect-failure

# when auth is enabled on mongodb and sqlproxy but the admin user does not have
# any roles, we expect the connection to be accepted as long as the user's schema
# comes from a drdl file
test-mongo-auth-drdl-no-roles: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-roles,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-drdl-no-roles: test-alternate-user-connect-success

# when auth is enabled on mongodb and sqlproxy but the admin user is missing the
# listCollections action on some databases in the schema, we expect the connection
# to be accepted
test-mongo-auth-drdl-read-some-dbs: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/read-some-dbs,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-drdl-read-some-dbs: test-alternate-user-connect-success

# when auth is enabled on mongodb and sqlproxy but the admin user does not have the listIndexes action
# on all databases in the schema, we expect the connection to be accepted
test-mongo-auth-drdl-no-listindexes: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/listcollections-only,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-drdl-no-listindexes: EXPECTED_ERROR := ERROR 1043 (08S01): error retrieving information from MongoDB: failed to run listIndexes on namespace <colname>
test-mongo-auth-drdl-no-listindexes: test-alternate-user-connect-success

# we expect the connection to succeeed if all the expected credentials are provided
test-mongo-auth-drdl-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/enabled,sqlproxy/auth/admin-creds,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-drdl-success: test-connect-success

_test-schema-empty: EXPECTED_DB_COUNT := 2
_test-schema-empty: _test-schema-dbs

_test-schema-dbs: QUERY := select count(*) from information_schema.schemata
_test-schema-dbs: EXPECTED = $(EXPECTED_DB_COUNT)
_test-schema-dbs: _test-mysql-query

_test-command-success: EXPECTED =
_test-command-success: _test-mysql-query

_test-command-failure: EXPECTED =
_test-command-failure: _test-mysql-query

test-auth-schema-available: run-mongodb restore-integration-data _create-test-user build-mongosqld run-mongosqld _test-connect-success _test-schema-dbs
test-auth-empty-schema-available: run-mongodb restore-integration-data _create-test-user build-mongosqld run-mongosqld _test-connect-success _test-schema-empty
test-auth-schema-not-available: EXPECTED_ERROR := ERROR 1043 (08S01): MongoDB schema not yet available
test-auth-schema-not-available: run-mongodb restore-integration-data _create-test-user build-mongosqld run-mongosqld _test-connect-failure
test-auth-command-success: run-mongodb _create-test-user build-mongosqld run-mongosqld _test-connect-success _test-command-success
test-auth-command-failure: run-mongodb _create-test-user build-mongosqld run-mongosqld _test-connect-success _test-command-failure
test-auth-command-data-success: run-mongodb restore-integration-data _create-test-user build-mongosqld run-mongosqld _test-connect-success _test-command-success
test-auth-command-data-failure: run-mongodb restore-integration-data _create-test-user build-mongosqld run-mongosqld _test-connect-success _test-command-failure

# when no admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-no-creds-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.4,sqlproxy/schema/mapping-majority
test-mongo-auth-sample-no-creds-3.4: test-auth-schema-not-available

# when no admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-no-creds-3.6: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.6,sqlproxy/schema/mapping-majority
test-mongo-auth-sample-no-creds-3.6: test-auth-schema-not-available

# when no admin credentials are provided and we are running against mongodb 3.7+,
# we expect the connection to fail because the schema is not yet available. This
# is different from prior mongodb versions since 3.7+ requires authentication to
# list all databases.
test-mongo-auth-sample-no-creds-4.0: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/4.0,sqlproxy/schema/mapping-majority
test-mongo-auth-sample-no-creds-4.0: test-auth-schema-not-available

# when no admin credentials are provided and we are running against mongodb 3.7+,
# we expect the connection to fail because the schema is not yet available. This
# is different from prior mongodb versions since 3.7+ requires authentication to
# list all databases.
test-mongo-auth-sample-no-creds-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority
test-mongo-auth-sample-no-creds-latest: test-auth-schema-not-available

# when incorrect admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-wrong-admin-creds-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.4,sqlproxy/schema/mapping-majority,sqlproxy/auth/wrong-admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-wrong-admin-creds-3.4: test-auth-schema-not-available

# when incorrect admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-wrong-admin-creds-3.6: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.6,sqlproxy/schema/mapping-majority,sqlproxy/auth/wrong-admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-wrong-admin-creds-3.6: test-auth-schema-not-available

# when incorrect admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-wrong-admin-creds-4.0: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/4.0,sqlproxy/schema/mapping-majority,sqlproxy/auth/wrong-admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-wrong-admin-creds-4.0: test-auth-schema-not-available

# when incorrect admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-wrong-admin-creds-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,sqlproxy/auth/wrong-admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-wrong-admin-creds-latest: test-auth-schema-not-available

# when correct admin credentials are provided but the user has no privileges,
# we expect the connection to fail because the schema is not yet available
test-mongo-auth-sample-no-roles-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.4,mongo/other-user/no-roles,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-roles-3.4: test-auth-schema-not-available

# when correct admin credentials are provided but the user has no privileges,
# we expect the connection to fail because the schema is not yet available
test-mongo-auth-sample-no-roles-3.6: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.6,mongo/other-user/no-roles,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-roles-3.6: test-auth-schema-not-available

# when correct admin credentials are provided but the user has no privileges,
# we expect the connection to fail because the schema is not yet available
test-mongo-auth-sample-no-roles-4.0: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/4.0,mongo/other-user/no-roles,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-roles-4.0: test-auth-empty-schema-available

# when correct admin credentials are provided but the user has no privileges,
# we expect the connection to fail because the schema is not yet available
test-mongo-auth-sample-no-roles-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-roles,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-roles-latest: test-auth-empty-schema-available

# when correct admin credentials are provided but the user does not have the listDatabases action,
# we expect the connection to succeed as long as none of the database selectors use wildcards
test-mongo-auth-sample-no-listdatabases-literal-db: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.6,mongo/other-user/no-listdatabases,sqlproxy/schema/mapping-majority,sqlproxy/schema/ns-literal-db,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listdatabases-literal-db: EXPECTED_DB_COUNT := 4
test-mongo-auth-sample-no-listdatabases-literal-db: test-auth-schema-available

# when correct admin credentials are provided but the user does not have the listDatabases action,
# we expect the connection to fail if any database selectors use wildcards and we are running against 3.6 or earlier
test-mongo-auth-sample-no-listdatabases-wildcard-db-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.4,mongo/other-user/no-listdatabases,sqlproxy/schema/mapping-majority,sqlproxy/schema/ns-wildcard-db,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listdatabases-wildcard-db-3.4: test-auth-schema-not-available

# when correct admin credentials are provided but the user does not have the listDatabases action,
# we expect the connection to fail if any database selectors use wildcards and we are running against 3.6 or earlier
test-mongo-auth-sample-no-listdatabases-wildcard-db-3.6: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.6,mongo/other-user/no-listdatabases,sqlproxy/schema/mapping-majority,sqlproxy/schema/ns-wildcard-db,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listdatabases-wildcard-db-3.6: test-auth-schema-not-available

# when correct admin credentials are provided but the user does not have the listDatabases action,
# we expect the connection to succeed as long as we are running against 3.7+, even when wildcards are used in db selectors
test-mongo-auth-sample-no-listdatabases-wildcard-db-4.0: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-listdatabases,sqlproxy/schema/mapping-majority,sqlproxy/schema/ns-wildcard-db,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listdatabases-wildcard-db-4.0: EXPECTED_DB_COUNT := 4
test-mongo-auth-sample-no-listdatabases-wildcard-db-4.0: test-auth-schema-available

# when correct admin credentials are provided but the user does not have the listDatabases action,
# we expect the connection to succeed as long as we are running against 3.7+, even when wildcards are used in db selectors
test-mongo-auth-sample-no-listdatabases-wildcard-db-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-listdatabases,sqlproxy/schema/mapping-majority,sqlproxy/schema/ns-wildcard-db,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listdatabases-wildcard-db-latest: EXPECTED_DB_COUNT := 4
test-mongo-auth-sample-no-listdatabases-wildcard-db-latest: test-auth-schema-available

# when correct admin credentials are provided but the user does not have the listCollections action
# on any databases, we expect the connection to succeed as long as no wildcards are used in collection selectors
test-mongo-auth-sample-no-listcollections-literal-collection: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-listcollections,sqlproxy/schema/mapping-majority,sqlproxy/schema/ns-literal-col,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listcollections-literal-collection: EXPECTED_DB_COUNT := 4
test-mongo-auth-sample-no-listcollections-literal-collection: test-auth-schema-available

# when correct admin credentials are provided but the user does not have the listCollections action
# on any databases, we expect the connection to fail if wildcards are used any collection selector
test-mongo-auth-sample-no-listcollections-wildcard-collection: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-listcollections,sqlproxy/schema/mapping-majority,sqlproxy/schema/ns-wildcard-col,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listcollections-wildcard-collection: test-auth-schema-not-available

# when correct admin credentials are provided and the user does not have the listCollections action,
# on the admin and local databases, we expect the connection to succeed even when no namespaces are specified.
test-mongo-auth-sample-no-admin-or-local-db: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-admin-or-local-db,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/schema/ns-literal-admin-local-test-db,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-admin-or-local-db: EXPECTED_DB_COUNT := 3
test-mongo-auth-sample-no-admin-or-local-db: test-auth-schema-available

# when correct admin credentials are provided, the user should be able to flush logs
test-mongo-auth-flush-logs-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-flush-logs-success: QUERY := FLUSH LOGS
test-mongo-auth-flush-logs-success: test-auth-command-success

# the admin user can flush sample because they have the proper permissions
test-mongo-auth-flush-sample-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-flush-sample-success: QUERY := FLUSH SAMPLE
test-mongo-auth-flush-sample-success: test-auth-command-data-success

# the admin user can alter tables because it has the proper permission on all tables (insert and update)
test-mongo-auth-alter-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds,sqlproxy/schema/enable-alter
test-mongo-auth-alter-success: QUERY :=  use join_test,,alter table join_1 drop column a
test-mongo-auth-alter-success: test-auth-command-data-success

# when correct admin credentials are provided, the user should be able to set global
test-mongo-auth-set-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-set-success: QUERY := set @@global.autocommit = 1
test-mongo-auth-set-success: test-auth-command-success

# when not admin user, the user should still be able to set session variables
test-mongo-auth-set-session-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,mongo/other-user/no-roles,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/other-user-creds
test-mongo-auth-set-session-success: QUERY := set @@session.autocommit = 1
test-mongo-auth-set-session-success: test-auth-command-success

# when correct admin credentials are provided, the user can see global variables
test-mongo-auth-global-variables-visible: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-global-variables-visible: QUERY := use information_schema,,select count(*) from GLOBAL_VARIABLES limit 0
test-mongo-auth-global-variables-visible: test-auth-command-success

# when correct admin credentials are provided, the user can see global status
test-mongo-auth-global-status-visible: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-global-status-visible: QUERY := use information_schema,,select count(*) from GLOBAL_STATUS limit 0
test-mongo-auth-global-status-visible: test-auth-command-success

# do all the auth-related tests for `SHOW PROCESSLIST` and `KILL CONNECTION`, since these both require having multiple connections
# note that `KILL CONNECTION` and `KILL QUERY` are identical from an auth perspective
test-mongo-auth-two-user: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,mongo/other-user/no-roles,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/second-user-creds,client/auth/creds
test-mongo-auth-two-user: run-mongodb _create-test-user build-mongosqld run-mongosqld
	$(ENV) PROCESS_COUNT_1=3 PROCESS_COUNT_2=2 USER_TO_KILL_1=alice USER_TO_KILL_2=bob EXPECTED_KILL_1='' EXPECTED_KILL_2="ERROR 1094 (HY000) at line 1: Unknown thread id:" testdata/bin/test-mongo-auth-two-user.sh

# when not admin user, the user should not be able to flush logs
test-mongo-auth-flush-logs-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,mongo/other-user/no-roles,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/other-user-creds
test-mongo-auth-flush-logs-failure: QUERY := FLUSH LOGS
test-mongo-auth-flush-logs-failure: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: only admin user can flush logs
test-mongo-auth-flush-logs-failure: test-auth-command-failure

# the user must have `find` permissions on all sampled namespaces to flush sample
test-mongo-auth-flush-sample-read-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,mongo/other-user/write-all-dbs,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/other-user-creds
test-mongo-auth-flush-sample-read-failure: QUERY := FLUSH SAMPLE
test-mongo-auth-flush-sample-read-failure: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: must have \`find\` privileges on the 'sample source' in order to flush sample
test-mongo-auth-flush-sample-read-failure: test-auth-command-data-failure

# the user must have `insert` and `update` permissions on the sample namespace
test-mongo-auth-flush-sample-write-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,mongo/other-user/no-roles,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/other-user-creds,sqlproxy/schema/clustered,sqlproxy/schema/write
test-mongo-auth-flush-sample-write-failure: QUERY := FLUSH SAMPLE
test-mongo-auth-flush-sample-write-failure: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: must have \`insert\` and \`update\` privileges on the 'sample source' or be admin user in order to flush sample
test-mongo-auth-flush-sample-write-failure: test-auth-command-data-failure

# a user cannot alter tables without insert and update privleges for the table in question when in clustered write mode
test-mongo-auth-alter-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,mongo/other-user/read-some-dbs,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/other-user-creds,sqlproxy/schema/enable-alter,sqlproxy/schema/clustered,sqlproxy/schema/write
test-mongo-auth-alter-failure: QUERY :=  use join_test,,alter table join_1 drop column a
test-mongo-auth-alter-failure: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: must have \`insert\` and \`update\` privileges for the 'sample source' or be admin user in order to alter tables
test-mongo-auth-alter-failure: test-auth-command-data-failure

# when not admin user, the user should not be able to set globals
test-mongo-auth-set-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,mongo/other-user/no-roles,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/other-user-creds
test-mongo-auth-set-failure: QUERY := set @@global.autocommit = 1
test-mongo-auth-set-failure: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: only admin user can set global variables
test-mongo-auth-set-failure: test-auth-command-failure

# when not admin user, the user should still be able to see global variables
test-mongo-auth-global-variables-still-visible: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,mongo/other-user/no-roles,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/other-user-creds
test-mongo-auth-global-variables-still-visible: QUERY := use information_schema,,select count(*) from GLOBAL_VARIABLES limit 0
test-mongo-auth-global-variables-still-visible: test-auth-command-success

# when not admin user, the user should still be able to see global status
test-mongo-auth-global-status-still-visible: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/mapping-majority,mongo/other-user/no-roles,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/other-user-creds
test-mongo-auth-global-status-still-visible: QUERY := use information_schema,,select count(*) from GLOBAL_STATUS limit 0
test-mongo-auth-global-status-still-visible: test-auth-command-success
