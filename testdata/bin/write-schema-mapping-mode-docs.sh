#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    echo "writing $NUM_DOCS sample documents..."
	cmd="db.schema_mapping_modes.insertMany([{mid: NumberLong(1)}, {mid: true}, {mid: false}, {mid: 3.14}]);"

    set +o errexit

    output="$($ARTIFACTS_DIR/mongodb/bin/mongo $MONGO_CLIENT_ARGS --eval "$cmd")"
    code=$?

    set -o errexit

    if [ "$code" != "0" ]; then
        echo "failed to execute mongo command: $output"
        exit 1
    fi

    echo "done writing sample docs"





