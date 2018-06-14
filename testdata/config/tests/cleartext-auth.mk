
# reject cleartext auth attempt with ssl disabled
test-cleartext-auth-nossl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/disabled,client/auth/creds,client/auth/cleartext,client/ssl/disable
test-cleartext-auth-nossl: EXPECTED_ERROR = ERROR 1759 (HY000): ssl is required when using cleartext authentication
test-cleartext-auth-nossl: test-connect-failure

# reject cleartext auth attempt for nossl tcp connections with sslMode=allowSSL
test-cleartext-auth-allowssl-tcp-nossl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/disable
test-cleartext-auth-allowssl-tcp-nossl: EXPECTED_ERROR = ERROR 1759 (HY000): ssl is required when using cleartext authentication
test-cleartext-auth-allowssl-tcp-nossl: test-connect-failure

# accept (on unix variants) cleartext auth attempt for nossl unix connections with sslMode=allowSSL
test-cleartext-auth-allowssl-unix-nossl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/disable,client/connection/unix-protocol
ifeq ($(VARIANT),windows)
# unix domain sockets don't exist in windows
test-cleartext-auth-allowssl-unix-nossl: EXPECTED_STATUS = 1
test-cleartext-auth-allowssl-unix-nossl: EXPECTED_ERROR = ERROR 2047 (HY000): Wrong or unknown protocol
test-cleartext-auth-allowssl-unix-nossl: test-connect-failure
else
test-cleartext-auth-allowssl-unix-nossl: test-connect-success
endif

# accept cleartext auth attempt for ssl connection
test-cleartext-auth-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/creds,client/auth/cleartext,client/ssl/require
test-cleartext-auth-ssl: test-connect-success

# accept SCRAM-SHA-256 cleartext auth attempt for ssl connection
test-cleartext-auth-ssl-scram-sha-256: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,mongo/version/4.0,mongo/other-user/root,sqlproxy/auth/admin-creds-other-user,sqlproxy/auth/enabled,sqlproxy/auth/scram-sha-256-mechanism,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/other-user-creds-scram-sha-256,client/auth/cleartext,client/ssl/require
test-cleartext-auth-ssl-scram-sha-256: MECHANISM := SCRAM-SHA-256
test-cleartext-auth-ssl-scram-sha-256: run-mongodb _create-test-user build-mongosqld run-mongosqld _test-connect-success

#  server should reject GSSAPI credentials
test-cleartext-auth-gssapi: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/ssl/allow,sqlproxy/ssl/pem,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-mongosql,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/auth/gssapi-correct-username-and-password,client/auth/gssapi-creds,client/auth/cleartext,client/ssl/require,client/ssl/pem,client/ssl/ca
test-cleartext-auth-gssapi: EXPECTED_ERROR = WARNING: no verification of server certificate will be done. Use --ssl-mode=VERIFY_CA or VERIFY_IDENTITY. ERROR 1045 (28000): Access denied for user 'drivers@LDAPTEST.10GEN.CC?mechanism=GSSAPI'
test-cleartext-auth-gssapi: test-connect-failure

# reject cleartext auth attempt with incorrect credentials
test-cleartext-auth-wrong-creds: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require
test-cleartext-auth-wrong-creds: EXPECTED_ERROR = ERROR 1045 (28000): Access denied for user '$(shell whoami)'
test-cleartext-auth-wrong-creds: test-connect-failure
