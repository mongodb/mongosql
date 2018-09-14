#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    set -o xtrace

    plugin_dir="$ARTIFACTS_DIR/mongosql_auth"
    mkdir -p "$plugin_dir"
    cd "$plugin_dir"

    if [ "$VARIANT" = 'windows' ]; then
        curl -O 'https://s3.amazonaws.com/mciuploads/mongosql-auth-c/artifacts/full_matrix__os~windows-64/mongosql_auth_c_full_matrix__os~windows_64_compile_ff1b98818af7b8b64638323565a47e78b0666200_18_09_04_22_14_11/mongosql_auth.dll'
        chmod +x mongosql_auth.dll
    elif [ "$VARIANT" = 'macos' ]; then
        curl -O 'https://s3.amazonaws.com/mciuploads/mongosql-auth-c/artifacts/full_matrix__os~osx/mongosql_auth_c_full_matrix__os~osx_compile_ff1b98818af7b8b64638323565a47e78b0666200_18_09_04_22_14_11/mongosql_auth.so'
        chmod +x mongosql_auth.so
    elif [ "$VARIANT" = 'rhel70' ]; then
        curl -O 'https://s3.amazonaws.com/mciuploads/mongosql-auth-c/artifacts/full_matrix__os~rhel70/mongosql_auth_c_full_matrix__os~rhel70_compile_ff1b98818af7b8b64638323565a47e78b0666200_18_09_04_22_14_11/mongosql_auth.so'
        chmod +x mongosql_auth.so
    elif [ "$VARIANT" = 'ubuntu1404' ]; then
        curl -O 'https://s3.amazonaws.com/mciuploads/mongosql-auth-c/artifacts/full_matrix__os~ubuntu1404-64/mongosql_auth_c_full_matrix__os~ubuntu1404_64_compile_ff1b98818af7b8b64638323565a47e78b0666200_18_09_04_22_14_11/mongosql_auth.so'
        chmod +x mongosql_auth.so
    else
        echo "variant $VARIANT not supported"
        exit 1
    fi

    mysql \
        --host 'localhost' \
        --port '3307' \
        --default-auth 'mongosql_auth' \
        --plugin-dir "$plugin_dir" \
        --user 'drivers?mechanism=PLAIN' \
        -p'powerbook17' \
        -e 'select * from information_schema.schemata'

) > $LOG_FILE 2>&1

print_exit_msg
