#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

    set -o errexit
    echo "starting mongosqld..."
    if [ "Windows_NT" = "$OS" ]; then
        # just to make sure these guys are stopped and not installed,
        # attempt to get rid of them,
        echo "stopping mongosqld..."
        net stop mongosql > /dev/null 2>&1 || true
        echo "deleting mongosqld service (sc)..."
        sc.exe delete mongosql > /dev/null 2>&1 || true
        echo "deleting mongosqld service (reg)..."
        reg delete "HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\EventLog\Application\mongosql" /f > /dev/null 2>&1 || true
        echo "printing mongosqld args..."
        echo "$SQLPROXY_ARGS"
        echo "renaming mongosqld..."
        mv $ARTIFACTS_DIR/bin/mongosqld $ARTIFACTS_DIR/bin/mongosqld.exe || true
        echo "installing mongosqld..."
        $ARTIFACTS_DIR/bin/mongosqld.exe install -vv $SQLPROXY_ARGS
        echo "starting mongosqld..."
        net start mongosql
        sleep 5
        echo $?
    else
        echo "printing mongosqld args..."
        echo "$SQLPROXY_ARGS"
        echo "starting mongosqld..."
        $ARTIFACTS_DIR/bin/mongosqld -vv \
            $SQLPROXY_ARGS &
        pid=$!

        sleep 5
        if ! kill -0 $pid; then
            echo "could not find mongosqld job after 5 seconds"
            exit 1
        fi
    fi

    echo "started mongosqld"


