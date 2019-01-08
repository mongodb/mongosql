test-single-bind-ip-default-port: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/address/single-bind-ip-no-port,client/connection/non-loopback-host
test-single-bind-ip-default-port: test-connect-success

test-single-bind-ip-with-port: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/address/single-bind-ip-with-port,client/connection/non-loopback-host,client/connection/non-default-port,sqlproxy/schema/drdl
test-single-bind-ip-with-port: test-connect-success

test-multiple-bind-ip-default-port: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/address/multiple-bind-ips-no-port,client/connection/non-loopback-host
test-multiple-bind-ip-default-port: test-connect-success

test-multiple-bind-ip-consistent-ports: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/address/multiple-bind-ips-consistent-ports,client/connection/non-loopback-host,client/connection/non-default-port
test-multiple-bind-ip-consistent-ports: test-connect-success

test-multiple-bind-ip-inconsistent-ports: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/address/multiple-bind-ips-inconsistent-ports,client/connection/non-loopback-host
test-multiple-bind-ip-inconsistent-ports: test-start-mongosqld-failure

test-bind-ips:
	make test-single-bind-ip-default-port
	make clean
	make test-single-bind-ip-with-port
	make clean
	make test-multiple-bind-ip-default-port
	make clean
	make test-multiple-bind-ip-consistent-ports
	make clean
	make test-multiple-bind-ip-inconsistent-ports
