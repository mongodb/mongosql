
test-mongo-ssl-failure: EXPECTED_ERROR = ERROR 1429 (HY000): Unable to connect to foreign data source: MongoDB
test-mongo-ssl-failure: test-connect-failure

# connection should fail when trying to connect to non-ssl mongod with ssl
test-mongodb-ssl-not-enabled-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/mongo-ssl/enabled
test-mongodb-ssl-not-enabled-failure: test-mongo-ssl-failure

# connection should fail when trying to connect to ssl mongod without ssl
test-mongo-ssl-not-enabled-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic
test-mongo-ssl-not-enabled-failure: test-mongo-ssl-failure

# test basic connection to ssl mongod
# NOTE: we deviate from MySQL shell behavior here, as we allow invalid certs
#       by default if --sslCAFile is not specified
test-mongo-ssl-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,sqlproxy/mongo-ssl/enabled
test-mongo-ssl-success: test-connect-success

# test connection to ssl mongod with cert verification
test-mongo-ssl-ca-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/ca
test-mongo-ssl-ca-success: test-connect-success

# test connection to ssl mongod with pem key
test-mongo-ssl-pem-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/pem
test-mongo-ssl-pem-success: test-connect-success

# connection to ssl mongod should fail with expired pem key
test-mongo-ssl-pem-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/expired-pem
test-mongo-ssl-pem-failure: test-mongo-ssl-failure

# test fips-mode connection to ssl mongod
test-mongo-ssl-fips: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/pem,sqlproxy/mongo-ssl/ca,sqlproxy/mongo-ssl/fips-mode
ifeq ($(VARIANT),macos)
test-mongo-ssl-fips: EXPECTED_STATUS = 1
test-mongo-ssl-fips: test-start-mongosqld
else ifeq ($(VARIANT),ubuntu1604)
test-mongo-ssl-fips: EXPECTED_STATUS = 1
test-mongo-ssl-fips: test-start-mongosqld
else
test-mongo-ssl-fips: test-connect-success
endif


# test connection to ssl mongod with pem key with minimum tls requirements
test-mongo-ssl-min-tls-default: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/pem
test-mongo-ssl-min-tls-default: test-connect-success

test-mongo-ssl-min-tls-1-0: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/pem,sqlproxy/mongo-ssl/min_tls_1_0
test-mongo-ssl-min-tls-1-0: test-connect-success

test-mongo-ssl-min-tls-1-1: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/pem,sqlproxy/mongo-ssl/min_tls_1_1
test-mongo-ssl-min-tls-1-1: test-connect-success

test-mongo-ssl-min-tls-1-2: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/pem,sqlproxy/mongo-ssl/min_tls_1_2
test-mongo-ssl-min-tls-1-2: test-connect-success
