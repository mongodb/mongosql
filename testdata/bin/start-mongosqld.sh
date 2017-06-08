#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "starting mongosqld..."

    if [ "Windows_NT" = "$OS" ]; then
        # just to make sure these guys are stopped and not installed,
        # attempt to get rid of them,
        net stop mongosql || true
        sc.exe delete mongosql || true
        reg delete "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\EventLog\Application\mongosql" /f || true

        $ARTIFACTS_DIR/bin/mongosqld install -vvvv \
            --logPath $ARTIFACTS_DIR/log/mongosqld.log \
            --schemaDirectory "$PROJECT_DIR/testdata/resources/schema" \
            $SQLPROXY_ARGS \
            $RACE_DETECTOR

        net start mongosql
    else
        nohup $ARTIFACTS_DIR/bin/mongosqld -vvvv \
            --logPath $ARTIFACTS_DIR/log/mongosqld.log \
            --schemaDirectory "$PROJECT_DIR/testdata/resources/schema" \
            $SQLPROXY_ARGS \
            $RACE_DETECTOR &
    fi

    echo "started mongosqld"

) > $LOG_FILE 2>&1

print_exit_msg
