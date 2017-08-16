
# reject cleartext auth attempt with ssl disabled
test-cleartext-auth-nossl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth,sqlproxy/ssl/disabled,client/auth/creds
test-cleartext-auth-nossl: EXPECTED_ERROR = ERROR 1759 (HY000): ssl is required when using cleartext authentication
test-cleartext-auth-nossl: test-connect-failure

# reject cleartext auth attempt for nossl connections with sslMode=allowSSL
test-cleartext-auth-allowssl-nossl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds
test-cleartext-auth-allowssl-nossl: EXPECTED_ERROR = ERROR 1759 (HY000): ssl is required when using cleartext authentication
test-cleartext-auth-allowssl-nossl: test-connect-failure

# accept cleartext auth attempt for ssl connection
test-cleartext-auth-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require
test-cleartext-auth-ssl: test-connect-success

# reject cleartext auth attempt with incorrect credentials
test-cleartext-auth-wrong-creds: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require
test-cleartext-auth-wrong-creds: EXPECTED_ERROR = ERROR 1043 (08S01): error performing authentication: unable to authenticate conversation 0: unable to authenticate using mechanism \"SCRAM-SHA-1\": (AuthenticationFailed) Authentication failed.
test-cleartext-auth-wrong-creds: test-connect-failure
