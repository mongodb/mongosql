#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    username="$MONGO_OTHER_USER_NAME"
    password="$MONGO_OTHER_USER_PWD"
    roles="$MONGO_OTHER_USER_ROLES"
    newrole="$MONGO_OTHER_USER_CUSTOM_ROLE"

    mongoBin=$ARTIFACTS_DIR/mongodb/bin/mongo

    if [ ! -z "$newrole" ]; then
        createRoleCmd="db.createRole($newrole)"
        echo "Creating role: $createRoleCmd"
        $mongoBin $MONGO_CLIENT_ARGS admin --eval "$createRoleCmd"
    else
        echo "Skipping custom role creation"
    fi

    if [ -z "$MECHANISM" ]; then
        MECHANISM="SCRAM-SHA-1"
    fi

    if [ ! -z "$username" ] && [ ! -z "$password" ]; then
        createUserCmd="db.createUser({user: '$username', pwd: '$password', roles: $roles})"
        if [ "$MECHANISM" == "SCRAM-SHA-256" ]; then
            createUserCmd="db.createUser({user: '$username', pwd: '$password', roles: $roles, mechanisms: [\"$MECHANISM\"]})"
        fi
        echo "Creating $MECHANISM user: $createUserCmd"
        $mongoBin $MONGO_CLIENT_ARGS admin --eval "$createUserCmd"
    else
        echo "Skipping secondary user creation"
    fi


) > $LOG_FILE 2>&1

print_exit_msg
