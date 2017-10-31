#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    echo "writing $NUM_DOCS sample documents..."
    cmd="for(i=0;i<$NUM_DOCS;i++){ doc={}; doc[i]=true; db.sample_test.insert(doc); }"

    set +o errexit

    output="$($ARTIFACTS_DIR/mongodb/bin/mongo $MONGO_CLIENT_ARGS --eval "$cmd")"
    code=$?

    set -o errexit

    if [ "$code" != "0" ]; then
        echo "failed to execute mongo command: $output"
        exit 1
    fi

    echo "done writing sample docs"

) > $LOG_FILE 2>&1

print_exit_msg

