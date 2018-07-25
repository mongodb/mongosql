#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

create_schema_cmd_for_generation() {
    generation=$1
    protocol=$2

    other_fields=""
    for i in $(seq 0 $generation); do
        other_fields="$other_fields,
        $i: { bsonType: \"int\" }"
    done

    cmd="
        vId = new ObjectId();
        db.mongosqld.schemas.insert({
            _id: new ObjectId(),
            versionId: vId,
            database: \"test\",
            collection: \"sample_test\",
            sampleSize: NumberLong(1),
            startSampleTime: new Date(),
            endSampleTime: new Date(),
            schema: {
                bsonType: \"object\",
                properties: {
                    _id: { bsonType: \"int\" }$other_fields
                },
            }
        });
        db.mongosqld.versions.insert({
            _id: vId,
            startSampleTime: new Date(),
            endSampleTime: new Date(),
            databases: [ { name: \"test\", collections: [ \"sample_test\" ] } ],
            generation: NumberLong($1),
            protocol: \"v1\",
            processName: \"write-initial-schema.sh\"
        });
    "

    if [ "$protocol" = 'v2' ]; then
    other_fields=""
    for i in $(seq 0 $generation); do
        other_fields="$other_fields,
        $i: {
             schemas: {
                 int: { bsonType: \"int\" }
             },
             counts: {
                 int: 3
             }
        }"
    done

    cmd="
        vId = new ObjectId();
        db.mongosqld.schemas.insert({
            _id: new ObjectId(),
            versionId: vId,
            database: \"test\",
            collection: \"sample_test\",
            sampleSize: NumberLong(1),
            startSampleTime: new Date(),
            endSampleTime: new Date(),
            schema: {
                bsonType: \"object\",
                properties: {
                    _id: {
                        schemas: {
                            int: { bsonType: \"int\" }
                        },
                        counts: {
                            int: 3
                        }
                    }$other_fields
                },
            }
        });
        db.mongosqld.versions.insert({
            _id: vId,
            startSampleTime: new Date(),
            endSampleTime: new Date(),
            databases: [ { name: \"test\", collections: [ \"sample_test\" ] } ],
            generation: NumberLong($1),
            protocol: \"v2\",
            processName: \"write-initial-schema.sh\"
        });
    "
    fi
    echo $cmd
}

(
    set -o errexit

    protocol="${PROTOCOL:-v2}"

    generation=${GENERATION:-0}
    echo "writing schema for generation $generation..."

    cmd="$(create_schema_cmd_for_generation $generation $protocol)"

    set +o errexit

    output="$($ARTIFACTS_DIR/mongodb/bin/mongo mongosqld_sample_test $MONGO_CLIENT_ARGS --eval "$cmd")"
    code=$?

    set -o errexit

    if [ "$code" != "0" ]; then
        echo "failed to execute mongo command: $output"
        exit 1
    fi

    echo "done writing schema"

) > $LOG_FILE 2>&1

print_exit_msg

