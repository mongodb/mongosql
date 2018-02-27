
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

_test-schema-full: EXPECTED_DB_COUNT := 19
_test-schema-full: _test-schema-dbs

_test-schema-dbs: QUERY := select count(*) from information_schema.schemata
_test-schema-dbs: EXPECTED = $(EXPECTED_DB_COUNT)
_test-schema-dbs: _test-mysql-query

test-auth-schema-available: run-mongodb restore-integration-data _create-test-user build-mongosqld run-mongosqld _test-connect-success _test-schema-dbs
test-auth-full-schema-available: run-mongodb restore-integration-data _create-test-user build-mongosqld run-mongosqld _test-connect-success _test-schema-full
test-auth-empty-schema-available: run-mongodb restore-integration-data _create-test-user build-mongosqld run-mongosqld _test-connect-success _test-schema-empty
test-auth-schema-not-available: EXPECTED_ERROR := ERROR 1043 (08S01): MongoDB schema not yet available
test-auth-schema-not-available: run-mongodb restore-integration-data _create-test-user build-mongosqld run-mongosqld _test-connect-failure

# when no admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-no-creds-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.4,sqlproxy/schema/dynamic
test-mongo-auth-sample-no-creds-3.4: test-auth-schema-not-available

# when no admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-no-creds-3.6: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.6,sqlproxy/schema/dynamic
test-mongo-auth-sample-no-creds-3.6: test-auth-schema-not-available

# when no admin credentials are provided and we are running against mongodb 3.7+,
# we expect the connection to succeed, but for no databases to be mapped into the schema
test-mongo-auth-sample-no-creds-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/dynamic
test-mongo-auth-sample-no-creds-latest: test-auth-empty-schema-available

# when incorrect admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-wrong-admin-creds-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.4,sqlproxy/schema/dynamic,sqlproxy/auth/wrong-admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-wrong-admin-creds-3.4: test-auth-schema-not-available

# when incorrect admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-wrong-admin-creds-3.6: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.6,sqlproxy/schema/dynamic,sqlproxy/auth/wrong-admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-wrong-admin-creds-3.6: test-auth-schema-not-available

# when incorrect admin credentials are provided, we expect the connection to fail
# because the schema is not yet available
test-mongo-auth-sample-wrong-admin-creds-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/dynamic,sqlproxy/auth/wrong-admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-wrong-admin-creds-latest: test-auth-schema-not-available

# when correct admin credentials are provided but the user has no privileges,
# we expect the connection to fail because the schema is not yet available
test-mongo-auth-sample-no-roles-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.4,mongo/other-user/no-roles,sqlproxy/schema/dynamic,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-roles-3.4: test-auth-schema-not-available

# when correct admin credentials are provided but the user has no privileges,
# we expect the connection to fail because the schema is not yet available
test-mongo-auth-sample-no-roles-3.6: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.6,mongo/other-user/no-roles,sqlproxy/schema/dynamic,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-roles-3.6: test-auth-schema-not-available

# when correct admin credentials are provided but the user has no privileges,
# we expect the connection to fail because the schema is not yet available
test-mongo-auth-sample-no-roles-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-roles,sqlproxy/schema/dynamic,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-roles-latest: test-auth-empty-schema-available

# when correct admin credentials are provided but the user does not have the listDatabases action,
# we expect the connection to succeed as long as none of the database selectors use wildcards
test-mongo-auth-sample-no-listdatabases-literal-db: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.6,mongo/other-user/no-listdatabases,sqlproxy/schema/dynamic,sqlproxy/schema/ns-literal-db,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listdatabases-literal-db: EXPECTED_DB_COUNT := 4
test-mongo-auth-sample-no-listdatabases-literal-db: test-auth-schema-available

# when correct admin credentials are provided but the user does not have the listDatabases action,
# we expect the connection to fail if any database selectors use wildcards and we are running against 3.6 or earlier
test-mongo-auth-sample-no-listdatabases-wildcard-db-3.4: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.4,mongo/other-user/no-listdatabases,sqlproxy/schema/dynamic,sqlproxy/schema/ns-wildcard-db,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listdatabases-wildcard-db-3.4: test-auth-schema-not-available

# when correct admin credentials are provided but the user does not have the listDatabases action,
# we expect the connection to fail if any database selectors use wildcards and we are running against 3.6 or earlier
test-mongo-auth-sample-no-listdatabases-wildcard-db-3.6: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/3.6,mongo/other-user/no-listdatabases,sqlproxy/schema/dynamic,sqlproxy/schema/ns-wildcard-db,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listdatabases-wildcard-db-3.6: test-auth-schema-not-available

# when correct admin credentials are provided but the user does not have the listDatabases action,
# we expect the connection to succeed as long as we are running against 3.7+, even when wildcards are used in db selectors
test-mongo-auth-sample-no-listdatabases-wildcard-db-latest: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-listdatabases,sqlproxy/schema/dynamic,sqlproxy/schema/ns-wildcard-db,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listdatabases-wildcard-db-latest: EXPECTED_DB_COUNT := 4
test-mongo-auth-sample-no-listdatabases-wildcard-db-latest: test-auth-schema-available

# when correct admin credentials are provided but the user does not have the listCollections action
# on any databases, we expect the connection to succeed as long as no wildcards are used in collection selectors
test-mongo-auth-sample-no-listcollections-literal-collection: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-listcollections,sqlproxy/schema/dynamic,sqlproxy/schema/ns-literal-col,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listcollections-literal-collection: EXPECTED_DB_COUNT := 4
test-mongo-auth-sample-no-listcollections-literal-collection: test-auth-schema-available

# when correct admin credentials are provided but the user does not have the listCollections action
# on any databases, we expect the connection to fail if wildcards are used any collection selector
test-mongo-auth-sample-no-listcollections-wildcard-collection: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/other-user/no-listcollections,sqlproxy/schema/dynamic,sqlproxy/schema/ns-wildcard-col,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-no-listcollections-wildcard-collection: test-auth-schema-not-available

# when correct admin credentials are provided, the full schema should be available
test-mongo-auth-sample-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/schema/dynamic,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require,client/auth/creds
test-mongo-auth-sample-success: test-auth-full-schema-available
