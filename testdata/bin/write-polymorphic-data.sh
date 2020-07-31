#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

    echo "creating column with polymorphic data types..."
    cmd="
        db.sample_test.insert({
            _id: new ObjectId(),
            sample_column: 2,
        });
        db.sample_test.insert({
            _id: new ObjectId(),
            sample_column: 2,
        });
        db.sample_test.insert({
            _id: new ObjectId(),
            sample_column: \"2\",
        });
    "

    output="$($ARTIFACTS_DIR/mongodb/bin/mongo mongosqld_sample_test $MONGO_CLIENT_ARGS --eval "$cmd")"
    code=$?

    if [ "$code" != "0" ]; then
        echo "failed to execute mongo command: $output"
        exit 1
    fi
    echo "wrote polymorphic data documents"





