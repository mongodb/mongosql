#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    echo "sharding collections..."
    cmd="sh.enableSharding(\"$DATABASE\"); sh.shardCollection(\"$DATABASE.$COLLECTION\", {_id : 1});"

    set +o errexit

    output="$($ARTIFACTS_DIR/mongodb/bin/mongo $MONGO_CLIENT_ARGS --eval "$cmd")"
    code=$?

    set -o errexit

    if [ "$code" != "0" ]; then
        echo "failed to execute mongo command: $output"
        exit 1
    fi

    echo "done sharding collections"

) > $LOG_FILE 2>&1

print_exit_msg

