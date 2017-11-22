#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    if [ ! -z $USER2_NAME ] && [ ! -z $USER2_PWD ]; then
        createUserCmd="db.createUser({user: '$USER2_NAME', pwd: '$USER2_PWD', roles: [$USER2_ROLES]})"
        echo "Creating user: $createUserCmd"

        mongoBin=$ARTIFACTS_DIR/mongodb/bin/mongo
        $mongoBin $MONGO_CLIENT_ARGS admin --eval "$createUserCmd"
    else
        echo "Skipping secondary user creation"
    fi


) > $LOG_FILE 2>&1
