#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

create_schema_cmd_for_generation() {
    generation=$1

    other_fields=""
    for i in $(seq 0 $generation); do
        other_fields="$other_fields,
            {
                mongo_name: 'field$i',
                mongo_type: 'bool',
                sql_name: 'col$1',
                sql_type: 'boolean',
            }"
    done

    cmd="
        sId = UUID().toString().slice(6, 43);
        db.schemas.insert({
            _id: sId,
            created: ISODate(),
            schema: {
                databases: [{
                    name: 'test',
                    tables: [{
                        sql_name: 'sample_test',
                        mongo_name: 'sample_test',
                        pipeline: '[]',
                        columns: [
                            {
                                mongo_name: '_id',
                                mongo_type: 'bson.ObjectId',
                                sql_name: '_id',
                                sql_type: 'objectid',
                            }$other_fields
                        ],
                    }]
                }]
            }
        });
        db.names.drop();
        db.names.insert({
            _id: 'defaultSchema',
            schema_id: sId,
        });
    "
    echo $cmd
}

(
    set -o errexit

    generation=${GENERATION:-0}
    echo "writing schema for generation $generation..."

    cmd="$(create_schema_cmd_for_generation $generation)"

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

