setup-kerberos:
	$(ENV) USER="$(USER)" testdata/bin/setup-kerberos-test.sh

run-mongosqld-gssapi-right-username-right-password: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-mongosql,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-correct-username-and-password
run-mongosqld-gssapi-right-username-right-password: run-mongosqld

run-mongosqld-gssapi-right-username-without-password: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-mongosql,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-correct-username-without-password
run-mongosqld-gssapi-right-username-without-password: run-mongosqld

run-mongosqld-gssapi-right-username-wrong-password: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-correct-username-wrong-password
run-mongosqld-gssapi-right-username-wrong-password: run-mongosqld _test-connect-failure

run-mongosqld-gssapi-wrong-username-wrong-password: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-wrong-username-wrong-password
run-mongosqld-gssapi-wrong-username-wrong-password: run-mongosqld _test-connect-failure

run-mongosqld-gssapi-wrong-username-without-password: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-wrong-username-without-password
run-mongosqld-gssapi-wrong-username-without-password: run-mongosqld _test-connect-failure

run-mongosqld-gssapi-right-username-right-password-with-keytab: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-mongosql,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-correct-username-and-password,
run-mongosqld-gssapi-right-username-right-password-with-keytab: run-mongosqld

run-mongosqld-gssapi-right-username-without-password-with-keytab: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-drivers,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-correct-username-without-password
run-mongosqld-gssapi-right-username-without-password-with-keytab: run-mongosqld

run-mongosqld-gssapi-right-username-wrong-password-with-keytab: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-drivers,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-correct-username-wrong-password
run-mongosqld-gssapi-right-username-wrong-password-with-keytab: run-mongosqld _test-connect-failure

run-mongosqld-gssapi-wrong-username-wrong-password-with-keytab: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-drivers,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-correct-username-wrong-password
run-mongosqld-gssapi-wrong-username-wrong-password-with-keytab: run-mongosqld _test-connect-failure

run-mongosqld-gssapi-wrong-username-without-password-with-keytab: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-drivers,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-correct-username-wrong-password
run-mongosqld-gssapi-wrong-username-without-password-with-keytab: run-mongosqld _test-connect-failure

run-mongosqld-keytab-and-username: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-drivers,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-correct-username-without-password
run-mongosqld-keytab-and-username: run-mongosqld

run-mongosqld-gssapi-wrong-service-name: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-mongosql,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism,sqlproxy/schema/gssapi-ns,sqlproxy/schema/mapping-majority,sqlproxy/auth/gssapi-correct-username-and-password,sqlproxy/auth/gssapi-wrong-service-name
run-mongosqld-gssapi-wrong-service-name: _test-wrong-service-name
_test-wrong-service-name: EXPECTED_STATUS = 0
_test-wrong-service-name:
	$(ENV) $(EXPECTED) testdata/bin/test-wrong-service-name.sh


# test gssapi with no credentials cache, just username and password
test-gssapi-with-correct-username-and-password-without-cache: build-mongosqld run-mongosqld-gssapi-right-username-right-password
	$(ENV) testdata/bin/run-gssapi-auth-tests.sh

test-gssapi-with-correct-username-wrong-password-without-cache: build-mongosqld
test-gssapi-with-correct-username-wrong-password-without-cache: EXPECTED_ERROR := $(SCHEMA_UNAVAILABLE_ERROR)
test-gssapi-with-correct-username-wrong-password-without-cache: run-mongosqld-gssapi-right-username-wrong-password

# test running sqlproxy with keytab and specified username, if ssl error message, then schema was sampled, and gssapi successfully authenticated
test-gssapi-with-keytab-and-username: build-mongosqld run-mongosqld-keytab-and-username
test-gssapi-with-keytab-and-username: EXPECTED_ERROR := ERROR 1759 (HY000): ssl is required when using cleartext authentication
test-gssapi-with-keytab-and-username: _test-connect-failure

# tests behaviour when there exists a credentials cache on the server
run-mongosqld-gssapi-cache-and-username: INFRASTRUCTURE_CONFIG := default,sqlproxy/gssapi/config,sqlproxy/mongo/gssapi-host,sqlproxy/gssapi/keytab-mongosql,sqlproxy/auth/gssapi-alt-username-without-password,sqlproxy/auth/enabled,sqlproxy/auth/gssapi-mechanism
run-mongosqld-gssapi-cache-and-username: run-mongosqld

test-gssapi-with-correct-username-and-password-wrong-cache: USER := schrödinger
test-gssapi-with-correct-username-and-password-wrong-cache: build-mongosqld setup-kerberos run-mongosqld-gssapi-right-username-right-password
ifeq ($(VARIANT),ubuntu1604)
test-gssapi-with-correct-username-and-password-wrong-cache: EXPECTED_ERROR := $(SCHEMA_UNAVAILABLE_ERROR)
else
test-gssapi-with-correct-username-and-password-wrong-cache: EXPECTED_ERROR := ERROR 1759 (HY000): ssl is required when using cleartext authentication
endif
test-gssapi-with-correct-username-and-password-wrong-cache: _test-connect-failure

test-gssapi-with-correct-username-without-password-with-cache: USER := drivers
test-gssapi-with-correct-username-without-password-with-cache: build-mongosqld setup-kerberos run-mongosqld-gssapi-right-username-without-password
	$(ENV) KEYTAB_NAME=drivers PRINCIPAL=drivers testdata/bin/run-gssapi-auth-tests.sh

test-gssapi-with-correct-username-wrong-password-with-cache: USER := drivers
ifeq ($(VARIANT),ubuntu1604)
test-gssapi-with-correct-username-wrong-password-with-cache: EXPECTED_ERROR := ERROR 1759 (HY000): ssl is required when using cleartext authentication
else
test-gssapi-with-correct-username-wrong-password-with-cache: EXPECTED_ERROR := $(SCHEMA_UNAVAILABLE_ERROR)
endif
test-gssapi-with-correct-username-wrong-password-with-cache: build-mongosqld setup-kerberos run-mongosqld-gssapi-right-username-wrong-password

test-gssapi-with-wrong-username-wrong-password-with-cache: USER := drivers
test-gssapi-with-wrong-username-wrong-password-with-cache: EXPECTED_ERROR := $(SCHEMA_UNAVAILABLE_ERROR)
test-gssapi-with-wrong-username-wrong-password-with-cache: build-mongosqld setup-kerberos run-mongosqld-gssapi-wrong-username-wrong-password

test-gssapi-with-wrong-username-without-password-with-cache: USER := drivers
test-gssapi-with-wrong-username-without-password-with-cache: EXPECTED_ERROR := $(SCHEMA_UNAVAILABLE_ERROR)
test-gssapi-with-wrong-username-without-password-with-cache: build-mongosqld setup-kerberos run-mongosqld-gssapi-wrong-username-without-password

# tests behaviour with a keytab set in KRB5_KTNAME and also supplied credentials
test-gssapi-with-correct-username-and-password-wrong-keytab: build-mongosqld run-mongosqld-gssapi-right-username-right-password-with-keytab
test-gssapi-with-correct-username-and-password-wrong-keytab: EXPECTED_ERROR := ERROR 1759 (HY000): ssl is required when using cleartext authentication
test-gssapi-with-correct-username-and-password-wrong-keytab: _test-connect-failure

test-gssapi-with-correct-username-wrong-password-with-keytab: EXPECTED_ERROR := $(SCHEMA_UNAVAILABLE_ERROR)
test-gssapi-with-correct-username-wrong-password-with-keytab: build-mongosqld run-mongosqld-gssapi-right-username-wrong-password-with-keytab

test-gssapi-with-wrong-username-wrong-password-with-keytab: EXPECTED_ERROR := $(SCHEMA_UNAVAILABLE_ERROR)
test-gssapi-with-wrong-username-wrong-password-with-keytab: build-mongosqld run-mongosqld-gssapi-wrong-username-wrong-password-with-keytab

test-gssapi-with-wrong-username-without-password-with-keytab: EXPECTED_ERROR := $(SCHEMA_UNAVAILABLE_ERROR)
test-gssapi-with-wrong-username-without-password-with-keytab: build-mongosqld run-mongosqld-gssapi-wrong-username-without-password-with-keytab

test-gssapi-wrong-service-name: USER := drivers
test-gssapi-wrong-service-name: EXPECTED_STATUS = 0
test-gssapi-wrong-service-name: build-mongosqld setup-kerberos run-mongosqld-gssapi-wrong-service-name

# test if there are changes in privilege in the case where the server has an existing credentials cache
# Sqlproxy will connect with user schrödinger to mongodb with GSSAPI.
# This creates a credentials cache on the sqlproxy server.
# A client then attempts to connect using gssapi to read from MongoDB.
# If the client successfully reads, then their privilege was not de-escalated by sqlproxy's cache.
test-gssapi-privilege-escalation: USER := schrödinger
test-gssapi-privilege-escalation: build-mongosqld setup-kerberos run-mongosqld-gssapi-cache-and-username
	$(ENV) testdata/bin/run-gssapi-auth-tests.sh

test-gssapi-admin-authenticator-plain-user: build-mongosqld run-mongosqld-gssapi-right-username-right-password
	$(ENV) testdata/bin/run-plain-auth-tests.sh
