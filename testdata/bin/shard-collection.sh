#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    set -o errexit

    echo "sharding collections..."
    if [ -z "$SHARD_KEY" ]; then
        SHARD_KEY="{_id: 1}"
    fi
    cmd="db.$COLLECTION.ensureIndex($SHARD_KEY); sh.enableSharding(\"$DATABASE\"); sh.shardCollection(\"$DATABASE.$COLLECTION\", $SHARD_KEY);"

    set +o errexit

    echo "executing mongodb query to enable sharding: $cmd"
    output="$($ARTIFACTS_DIR/mongodb/bin/mongo $MONGO_CLIENT_ARGS --eval "$cmd" $DATABASE)"
    code=$?

    set -o errexit

    if [ "$code" != "0" ]; then
        echo "failed to execute mongo command: $output"
        exit 1
    fi

    echo "done sharding collections"

    $ARTIFACTS_DIR/mongodb/bin/mongo $MONGO_CLIENT_ARGS --eval "sh.status()" $DATABASE

    echo "done waiting for balancer"





