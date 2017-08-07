
#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "running mongosqld startup test..."

    nohup $ARTIFACTS_DIR/bin/mongosqld -vvvv \
        $SQLPROXY_ARGS > $ARTIFACTS_DIR/mongosqld-out 2>&1 &
    pid=$!

    sleep 5

    set +o errexit

    kill -0 $pid
    started=$?

    set -o errexit

    echo "stderr:"
    cat $ARTIFACTS_DIR/mongosqld-out

    if [ "$started" != "$EXPECTED_STATUS" ]; then
        echo "expected status=$EXPECTED_STATUS, got started=$started"
        exit 1
    fi

    echo "done running mongosqld startup test"

) > $LOG_FILE 2>&1

print_exit_msg
