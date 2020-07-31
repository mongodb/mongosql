#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

    echo "creating conflicting table namespaces for schema generation..."
    cmd="
        db.sample_test.insert({
            _id: new ObjectId(),
            xX : [ { c : 1, c_0 : 2, C : 3 } ],
            XX : 4,
            Xx : [ { b : 5 } ],
            xX_0 : 6,
        });
    "

    output="$($ARTIFACTS_DIR/mongodb/bin/mongo mongosqld_sample_test $MONGO_CLIENT_ARGS --eval "$cmd")"
    code=$?

    if [ "$code" != "0" ]; then
        echo "failed to execute mongo command: $output"
        exit 1
    fi
    echo "wrote sample documents"





