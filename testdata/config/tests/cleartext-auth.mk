
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

#  accept GSSAPI credentials
test-cleartext-auth-gssapi: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/ssl/allow,sqlproxy/ssl/pem,sqlproxy/gssapi/test,client/auth/gssapi_creds,client/auth/cleartext,client/ssl/require,client/ssl/pem,client/ssl/ca
test-cleartext-auth-gssapi: test-connect-success

# reject cleartext auth attempt with incorrect credentials
test-cleartext-auth-wrong-creds: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/auth,sqlproxy/auth/admin-creds,sqlproxy/auth/enabled,sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/auth/cleartext,client/ssl/require
test-cleartext-auth-wrong-creds: EXPECTED_ERROR = ERROR 1045 (28000): Access denied for user '$(shell whoami)'
test-cleartext-auth-wrong-creds: test-connect-failure
