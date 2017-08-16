
# reject non-ssl connections when sslMode=requireSSL
test-require-ssl-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/ssl/require,sqlproxy/ssl/pem
test-require-ssl-failure: EXPECTED_ERROR = ERROR 1043 (08S01): recv handshake response error: This server is configured to only allow SSL connections
test-require-ssl-failure: test-connect-failure

# accept ssl connections when sslMode=requireSSL
test-require-ssl-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/ssl/require,sqlproxy/ssl/pem,client/ssl/require
test-require-ssl-success: test-connect-success

# reject ssl connections when sslMode=disabled
test-disable-ssl-failure: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/ssl/disabled,client/ssl/require
test-disable-ssl-failure: EXPECTED_ERROR = ERROR 2026 (HY000): SSL connection error: SSL is required but the server doesn't support it
test-disable-ssl-failure: test-connect-failure

# accept non-ssl connections when sslMode=disabled
test-disable-ssl-success: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/ssl/disabled,client/ssl/disable
test-disable-ssl-success: test-connect-success

# accept non-ssl connections when sslMode=allowSSL
test-allow-ssl-ssl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/ssl/disable
test-allow-ssl-ssl: test-connect-success

# accept ssl connections when sslMode=allowSSL
test-allow-ssl-nossl: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/ssl/allow,sqlproxy/ssl/pem,client/ssl/require
test-allow-ssl-nossl: test-connect-success
