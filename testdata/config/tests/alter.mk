
_try-alter:
	$(ENV) EXPECTED_ERROR="$(EXPECTED_ERROR)" EXPECTED_STATUS="$(EXPECTED_STATUS)" testdata/bin/test-alter.sh

_rename-column: EXPECTED_STATUS := 0
_rename-column: _try-rename-column

_rename-column-fail: EXPECTED_STATUS := 1
_rename-column-fail: _try-rename-column

_rename-column-twice:
	$(ENV) TABLE="$(TABLE)" COLUMN="$(COLUMN)" NEW_COLUMN="$(SECOND_COLUMN)" EXPECTED_STATUS="0" testdata/bin/rename-column.sh
	$(ENV) TABLE="$(TABLE)" COLUMN="$(SECOND_COLUMN)" NEW_COLUMN="$(NEW_COLUMN)" EXPECTED_STATUS="0" testdata/bin/rename-column.sh

_try-rename-column:
	$(ENV) TABLE="$(TABLE)" COLUMN="$(COLUMN)" NEW_COLUMN="$(NEW_COLUMN)" EXPECTED_STATUS="$(EXPECTED_STATUS)" EXPECTED_ERROR="$(EXPECTED_ERROR)" testdata/bin/rename-column.sh

_test-altered: QUERY := select count(*) from information_schema.columns where table_schema = 'test' and table_name = 'foo';
_test-altered: EXPECTED := 5
_test-altered: _test-mysql-query

_test-not-altered: QUERY := select count(*) from information_schema.columns where table_schema = 'test' and table_name = 'test1';
_test-not-altered: EXPECTED := 5
_test-not-altered: _test-mysql-query

test-alter-success: EXPECTED_STATUS := 0
test-alter-success: build-mongosqld run-mongodb restore-data run-mongosqld _test-schema-available _test-connect-success _try-alter _test-altered

test-alter-failure: EXPECTED_STATUS := 1
test-alter-failure: build-mongosqld run-mongodb restore-data run-mongosqld _test-schema-available _test-connect-success _try-alter _test-not-altered

test-alter-drdl: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: cannot alter schema: schema was loaded from a file
test-alter-drdl: test-alter-failure

test-alter-disabled: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered
test-alter-disabled: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: cannot alter schema: alterations not enabled
test-alter-disabled: test-alter-failure

test-alter-clustered-read: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered,sqlproxy/schema/enable-alter
test-alter-clustered-read: EXPECTED_ERROR := ERROR 1105 (HY000) at line 1: cannot alter schema in clustered read mode
test-alter-clustered-read: test-alter-failure

test-alter-standalone: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/enable-alter
test-alter-standalone: test-alter-success

test-alter-clustered-write: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/clustered,sqlproxy/schema/write,sqlproxy/schema/enable-alter
test-alter-clustered-write: test-alter-success

test-alter-flush: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/enable-alter
test-alter-flush: EXPECTED_STATUS := 0
test-alter-flush: build-mongosqld run-mongodb restore-data run-mongosqld _test-schema-available _test-connect-success _try-alter _test-flush _test-not-altered
_test-flush:
	$(ENV) EXPECTED_STATUS='0' EXPECTED_ERROR='' testdata/bin/test-flush.sh

test-rename-column: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/enable-alter
test-rename-column: build-mongosqld run-mongodb restore-data run-mongosqld _test-schema-available _test-connect-success _rename-column _test-rename-success

test-rename-column-fail: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/enable-alter
test-rename-column-fail: build-mongosqld run-mongodb restore-data run-mongosqld _test-schema-available _test-connect-success _rename-column-fail

_test-rename-success: EXPECTED := 1
_test-rename-success: QUERY = select count(*) from information_schema.columns where table_name = '$(TABLE)' and column_name = '$(NEW_COLUMN)';
_test-rename-success: _test-mysql-query

test-rename-column-simple: TABLE := test1
test-rename-column-simple: COLUMN := a
test-rename-column-simple: NEW_COLUMN := aaa
test-rename-column-simple: test-rename-column

test-rename-column-to-similar-to-id: TABLE := test1
test-rename-column-to-similar-to-id: COLUMN := a
test-rename-column-to-similar-to-id: NEW_COLUMN := _id_a
test-rename-column-to-similar-to-id: test-rename-column

test-rename-column-id: TABLE := test1
test-rename-column-id: COLUMN := _id
test-rename-column-id: NEW_COLUMN := z
test-rename-column-id: test-rename-column

test-rename-column-duplicate: TABLE := test1
test-rename-column-duplicate: COLUMN := a
test-rename-column-duplicate: NEW_COLUMN := b
test-rename-column-duplicate: EXPECTED_ERROR := ERROR 1060 (42S21) at line 1: Duplicate column name 'b'
test-rename-column-duplicate: test-rename-column-fail

test-rename-column-twice: INFRASTRUCTURE_CONFIG := $(INFRASTRUCTURE_CONFIG),sqlproxy/schema/mapping-majority,sqlproxy/schema/enable-alter
test-rename-column-twice: TABLE := test1
test-rename-column-twice: COLUMN := a
test-rename-column-twice: SECOND_COLUMN := foo
test-rename-column-twice: NEW_COLUMN := bar
test-rename-column-twice: build-mongosqld run-mongodb restore-integration-data run-mongosqld _test-schema-available _test-connect-success _rename-column-twice _test-rename-success
