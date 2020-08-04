#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "running mongosqld startup test..."

    $ARTIFACTS_DIR/bin/mongosqld -vvvv \
        $SQLPROXY_ARGS > $ARTIFACTS_DIR/mongosqld-out 2>&1 &
    pid=$!

    sleep 5

    set +o errexit

    kill -0 $pid
    started=$?

    echo "stderr:"
    cat $ARTIFACTS_DIR/mongosqld-out

    if grep "wrongServiceName/ldaptest.10gen.cc@LDAPTEST.10GEN.CC" "$ARTIFACTS_DIR"/log/mongosqld.log
    then
      echo "customized service name detected"
    else
      echo "service name was not set"
      exit 1
    fi

    echo "done running gsaapi service name test"

) > $LOG_FILE 2>&1

print_exit_msg
