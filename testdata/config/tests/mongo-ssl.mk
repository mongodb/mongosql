
test-mongo-ssl-failure: EXPECTED_ERROR = ERROR 2013 (HY000): Lost connection to MySQL server at 'reading initial communication packet'
test-mongo-ssl-failure: test-connect-failure

# connection should fail when trying to connect to non-ssl mongod with ssl
test-mongodb-ssl-not-enabled-failure: INFRASTRUCTURE_CONFIG = default,sqlproxy/mongo-ssl/enabled
test-mongodb-ssl-not-enabled-failure: test-mongo-ssl-failure

# connection should fail when trying to connect to ssl mongod without ssl
test-mongo-ssl-not-enabled-failure: INFRASTRUCTURE_CONFIG = default,mongo/ssl/basic
test-mongo-ssl-not-enabled-failure: test-mongo-ssl-failure

# test basic connection to ssl mongod
# NOTE: we deviate from MySQL shell behavior here, as we allow invalid certs
#       by default if --sslCAFile is not specified
test-mongo-ssl-success: INFRASTRUCTURE_CONFIG = default,mongo/ssl/basic,sqlproxy/mongo-ssl/enabled
test-mongo-ssl-success: test-connect-success

# test connection to ssl mongod with cert verification
test-mongo-ssl-ca-success: INFRASTRUCTURE_CONFIG = default,mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/ca
test-mongo-ssl-ca-success: test-connect-success

# test connection to ssl mongod with pem key
test-mongo-ssl-pem-success: INFRASTRUCTURE_CONFIG = default,mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/pem
test-mongo-ssl-pem-success: test-connect-success

# connection to ssl mongod should fail with expired pem key
test-mongo-ssl-pem-failure: INFRASTRUCTURE_CONFIG = default,mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/expired-pem
test-mongo-ssl-pem-failure: test-mongo-ssl-failure

# test fips-mode connection to ssl mongod
test-mongo-ssl-fips: INFRASTRUCTURE_CONFIG = default,mongo/ssl/basic,sqlproxy/mongo-ssl/enabled,sqlproxy/mongo-ssl/pem,sqlproxy/mongo-ssl/ca,sqlproxy/mongo-ssl/fips-mode
test-mongo-ssl-fips: test-connect-success
