#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "starting mongosqld..."

    nohup $ARTIFACTS_DIR/bin/mongosqld -vvvv \
        --logPath $ARTIFACTS_DIR/log/mongosqld.log \
        --schemaDirectory "$PROJECT_DIR/testdata/resources/schema" \
        $SQLPROXY_ARGS \
        $RACE_DETECTOR &

    echo "started mongosqld"

) > $LOG_FILE 2>&1

print_exit_msg
