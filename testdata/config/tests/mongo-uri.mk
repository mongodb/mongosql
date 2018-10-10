# Test the read preference modes on a mongodb uri.
test-mongo-uri-read-preference-secondary: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/replica-set,sqlproxy/schema/interval-2,mongo/uri/read-pref-secondary
test-mongo-uri-read-preference-secondary: run-mongodb build-mongosqld run-mongosqld _test-mongo-uri-read-preference-secondary

_test-mongo-uri-read-preference-secondary: 
	$(ENV) EXPECT_PRIMARY_READS=0 EXPECT_SECONDARY_READS=1 testdata/bin/test-mongo-uri-read-preference.sh

test-mongo-uri-read-preference-primary: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),mongo/topology/replica-set,sqlproxy/schema/interval-2,mongo/uri/read-pref-primary
test-mongo-uri-read-preference-primary: run-mongodb build-mongosqld run-mongosqld _test-mongo-uri-read-preference-primary

_test-mongo-uri-read-preference-primary:
	$(ENV) EXPECT_PRIMARY_READS=1 EXPECT_SECONDARY_READS=0 testdata/bin/test-mongo-uri-read-preference.sh
