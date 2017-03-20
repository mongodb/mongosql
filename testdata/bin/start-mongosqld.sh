#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "starting mongosqld..."

    nohup $ARTIFACTS_DIR/bin/mongosqld -vvvv \
        --schemaDirectory "$PROJECT_DIR/testdata/resources/schema" \
        $SQLPROXY_ARGS \
        $RACE_DETECTOR \
        > $ARTIFACTS_DIR/log/mongosqld.log 2>&1 &

    echo "started mongosqld"

) > $LOG_FILE 2>&1

print_exit_msg
