#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"

(
    echo "cleaning up processes..."
    pkill -9 mongo
    pkill -9 mongodump
    pkill -9 mongoexport
    pkill -9 mongoimport
    pkill -9 mongofiles
    pkill -9 mongooplog
    pkill -9 mongorestore
    pkill -9 mongostat
    pkill -9 mongotop
    pkill -9 mongod
    pkill -9 mongos
    pkill -9 mongosqld
    pkill -9 -f mongo-orchestration
    echo "done cleaning up processes"

    echo "cleaning up repo..."
    rm -rf $ARTIFACTS_DIR
    rm /tmp/mysql.sock
    echo "done cleaning up repo"

    echo "setting up repo for testing..."
    mkdir -p $ARTIFACTS_DIR/{bin,build,log,out,reports}
    echo "done setting up repo for testing"

    exit 0

) > /dev/null 2>&1

print_exit_msg
