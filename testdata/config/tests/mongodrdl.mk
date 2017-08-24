
test-drdl-connect-success: EXPECTED_STATUS = 0
test-drdl-connect-success: build-mongodrdl run-mongodb
	$(ENV) $(EXPECTED) testdata/bin/test-drdl-connect.sh

test-drdl-connect-failure: EXPECTED_STATUS = 1
test-drdl-connect-failure: build-mongodrdl run-mongodb
	$(ENV) $(EXPECTED) testdata/bin/test-drdl-connect.sh

# test that drdl connects with no special configuration
test-drdl-simple: test-drdl-connect-success

# test that drdl connects with auth
test-drdl-auth: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,drdl/auth/creds
test-drdl-auth: test-drdl-connect-success

# drdl should fail to connect with no credentials
test-drdl-auth-no-creds: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth
test-drdl-auth-no-creds: EXPECTED_ERROR = Failed: Can't get the collection names for test: (Unauthorized) not authorized on test to execute command { listCollections: 1, cursor: {}, readPreference: { mode: \"secondaryPreferred\" }, db: \"test\" }
test-drdl-auth-no-creds: test-drdl-connect-failure

# drdl should fail to connect with incorrect credentials
test-drdl-auth-wrong-creds: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,drdl/auth/wrong-creds
test-drdl-auth-wrong-creds: EXPECTED_ERROR = Failed: can't create session: no servers available: server selection failed: context deadline exceeded
test-drdl-auth-wrong-creds: test-drdl-connect-failure

# test that drdl connects with ssl
test-drdl-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,drdl/ssl/enable
test-drdl-ssl: test-drdl-connect-success
