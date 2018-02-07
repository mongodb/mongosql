
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
