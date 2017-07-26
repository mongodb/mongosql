
test-log-rotation: build-mongosqld run-mongodb run-mongosqld
	$(ENV) MYSQL_CMD="$(CMD)" EXPECTED_NUM_FILES="$(NUM_FILES)" testdata/bin/test-log-rotation.sh

test-log-rotation-none: CMD = use test,
test-log-rotation-none: NUM_FILES = 1
test-log-rotation-none: test-log-rotation

test-log-rotation-once: CMD = use test, flush logs,
test-log-rotation-once: NUM_FILES = 2
test-log-rotation-once: test-log-rotation

test-log-rotation-twice: CMD = use test, flush logs, flush logs,
test-log-rotation-twice: NUM_FILES = 3
test-log-rotation-twice: test-log-rotation
