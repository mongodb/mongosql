#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

create_sample_array_document() {
    cmd="
        vId = new ObjectId();
        db.sample_test.insert({
            _id: vId,
            address : {
                building : \"351\",
                coord : [
                    -73.98513559999999,
                    40.7676919
                ],
                street : \"West   57 Street\",
                zipcode : \"10019\"
            },
            borough : \"Manhattan\",
            cuisine : \"Irish\",
            grades : [
                {
                    date : new Date(),
                    grade : \"A\",
                    score : 2
                }
            ]
        });

        db.sample_test.createIndex({
            \"address.coord\" : \"$INDEX_TYPE\"
        });
    "
    echo $cmd
}

create_sample_nested_array_document() {
    cmd="
        vId = new ObjectId();
        db.sample_test.insert({
            _id: vId,
            address : {
                building : \"351\",
                location : {
                    coord : [
                        -73.98513559999999,
                        40.7676919
                    ],
                },
                street : \"West   57 Street\",
                zipcode : \"10019\"
            },
            borough : \"Manhattan\",
            cuisine : \"Irish\",
            grades : [
                {
                    date : new Date(),
                    grade : \"A\",
                    score : 2
                }
            ]
        });

        db.sample_test.createIndex({
            \"address.location.coord\" : \"$INDEX_TYPE\"
        });
    "
    echo $cmd
}


    set -o errexit
    
    echo "creating sample array document with accompanying index..."
    
    cmd="$(create_sample_array_document)"

    if [ "$NESTED" == "1" ]; then
        echo "sample array document will contain nested field"
        cmd="$(create_sample_nested_array_document)"
    fi

    echo "sample array document will have $INDEX_TYPE index"

    output="$($ARTIFACTS_DIR/mongodb/bin/mongo mongosqld_sample_test $MONGO_CLIENT_ARGS --eval "$cmd")"
    code=$?

    set -o errexit

    if [ "$code" != "0" ]; then
        echo "failed to execute mongo command: $output"
        exit 1
    fi

    echo "done writing array document"





