
test-drdl-connect-success: EXPECTED_STATUS = 0
test-drdl-connect-success: build-mongodrdl run-mongodb
	$(ENV) $(EXPECTED) testdata/bin/test-drdl-connect.sh

test-drdl-connect-failure: EXPECTED_STATUS = 1
test-drdl-connect-failure: build-mongodrdl run-mongodb
	$(ENV) $(EXPECTED) testdata/bin/test-drdl-connect.sh

test-mongo-drdl-gssapi-success: EXPECTED_STATUS = 0
test-mongo-drdl-gssapi-success: build-mongodrdl
	$(ENV) $(EXPECTED) testdata/bin/test-mongodrdl-gssapi-connect.sh

test-mongo-drdl-gssapi-failure: EXPECTED_STATUS = 1
test-mongo-drdl-gssapi-failure: build-mongodrdl
	$(ENV) $(EXPECTED) testdata/bin/test-mongodrdl-gssapi-connect.sh

# test that drdl connects with no special configuration
test-drdl-simple: test-drdl-connect-success

# test that drdl connects with auth
test-drdl-auth: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,drdl/auth/creds
test-drdl-auth: test-drdl-connect-success

# drdl should fail to connect with no credentials
test-drdl-auth-no-creds: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth
test-drdl-auth-no-creds: EXPECTED_ERROR = Failed: can't get the collection names for 'test': error listing collections for 'test': (Unauthorized) command listCollections requires authentication
test-drdl-auth-no-creds: test-drdl-connect-failure

# drdl should fail to connect with incorrect credentials
test-drdl-auth-wrong-creds: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,drdl/auth/wrong-creds
test-drdl-auth-wrong-creds: EXPECTED_ERROR = Failed: unable to execute command: connection() error occured during connection handshake: auth error: sasl conversation error: unable to authenticate using mechanism \"SCRAM-SHA-1\": (AuthenticationFailed) Authentication failed.
test-drdl-auth-wrong-creds: test-drdl-connect-failure

# test that drdl connects with ssl
test-drdl-ssl-default: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,drdl/ssl/enable
test-drdl-ssl-default: test-drdl-connect-success

test-drdl-ssl-min-tls-1-0: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,drdl/ssl/enable,drdl/ssl/min_tls_1_0
test-drdl-ssl-min-tls-1-0: test-drdl-connect-success

test-drdl-ssl-min-tls-1-1: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,drdl/ssl/enable,drdl/ssl/min_tls_1_1
test-drdl-ssl-min-tls-1-1: test-drdl-connect-success

test-drdl-ssl-min-tls-1-2: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,drdl/ssl/enable,drdl/ssl/min_tls_1_2
test-drdl-ssl-min-tls-1-2: test-drdl-connect-success

test-drdl-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,drdl/ssl/enable
test-drdl-ssl: test-drdl-connect-success

test-drdl-replset-seedlist: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/replica-set,drdl/mongo/replset-host,drdl/mongo/replset-ns
test-drdl-replset-seedlist: test-drdl-connect-success

# test that drdl connects with gssapi
test-drdl-gssapi: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),drdl/mongo/gssapi-host,drdl/mongo/gssapi-ns,drdl/auth/gssapi-correct-username-and-password,drdl/auth/gssapi-mechanism
test-drdl-gssapi: test-mongo-drdl-gssapi-success

# test that drdl connects with gssapi using credentials cache
test-drdl-gssapi-using-cred-cache: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),drdl/mongo/gssapi-host,drdl/mongo/gssapi-ns,drdl/auth/gssapi-mechanism
test-drdl-gssapi-using-cred-cache: USER := drivers
test-drdl-gssapi-using-cred-cache: setup-kerberos test-mongo-drdl-gssapi-success

# test that drdl connects with gssapi
test-drdl-gssapi-wrong-service-name: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),drdl/mongo/gssapi-host,drdl/mongo/gssapi-ns,drdl/auth/gssapi-wrong-service-name,drdl/auth/gssapi-mechanism
test-drdl-gssapi-wrong-service-name: test-mongo-drdl-gssapi-failure

# test that drdl connects with gssapi
test-drdl-gssapi-wrong-host-name: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),drdl/mongo/gssapi-host,drdl/mongo/gssapi-ns,drdl/auth/gssapi-wrong-host-name,drdl/auth/gssapi-mechanism
test-drdl-gssapi-wrong-host-name: test-mongo-drdl-gssapi-failure
