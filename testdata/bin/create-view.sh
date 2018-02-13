#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    echo "creating view..."

    "$ARTIFACTS_DIR/mongodb/bin/mongo" $MONGO_CLIENT_ARGS --eval "db.createView('$VIEW', '$SOURCE', [])" "$DATABASE"

    echo "done creating view"

) > $LOG_FILE 2>&1

print_exit_msg

