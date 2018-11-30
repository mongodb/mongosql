
test-log-rotation: build-mongosqld run-mongodb  _write-initial-schema _write-initial-docs run-mongosqld _test-log-rotation
_test-log-rotation:
	$(ENV) MYSQL_CMD="$(CMD)" ROTATION_METHOD="$(ROTATION_METHOD)" EXPECTED_NUM_FILES="$(NUM_FILES)" testdata/bin/test-log-rotation.sh

# tests for normal log rotation

test-log-rotation-none: CMD = use test,
test-log-rotation-none: NUM_FILES = 1
test-log-rotation-none: test-log-rotation

test-log-rotation-once: CMD = use test, flush logs,
test-log-rotation-once: NUM_FILES = 2
test-log-rotation-once: test-log-rotation

test-log-rotation-twice: CMD = use test, flush logs, flush logs,
test-log-rotation-twice: NUM_FILES = 3
test-log-rotation-twice: test-log-rotation

test-log-rotation-reopen: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/log/reopen
test-log-rotation-reopen: CMD = use test, flush logs, flush logs,
test-log-rotation-reopen: NUM_FILES = 1
test-log-rotation-reopen: test-log-rotation

# tests for rotation of existing file at startup

ifeq ($(ARTIFACTS_DIR),)
ARTIFACTS_DIR := testdata/artifacts
endif

create-log-file:
	echo "content" > $(ARTIFACTS_DIR)/log/mongosqld.log

create-log-dir:
	mkdir $(ARTIFACTS_DIR)/log/mongosqld.log

test-log-rotation-startup-rotate: CMD = use test, flush logs
test-log-rotation-startup-rotate: NUM_FILES = 3
test-log-rotation-startup-rotate: create-log-file test-log-rotation

test-log-rotation-startup-append: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/log/reopen
test-log-rotation-startup-append: CMD = use test, flush logs
test-log-rotation-startup-append: NUM_FILES = 1
test-log-rotation-startup-append: create-log-file test-log-rotation

test-log-rotation-startup-dir: EXPECTED_STATUS = 1
test-log-rotation-startup-dir: create-log-dir test-start-mongosqld

test-log-rotation-sigusr1: NUM_FILES = 2
test-log-rotation-sigusr1: ROTATION_METHOD = SIGUSR1
test-log-rotation-sigusr1: test-log-rotation
