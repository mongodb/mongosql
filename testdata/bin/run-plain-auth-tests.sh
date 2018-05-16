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
        curl -O 'https://s3.amazonaws.com/mciuploads/mongosql-auth-c/artifacts/full_matrix__os~windows-64/mongosql_auth_c_full_matrix__os~windows_64_compile_c650aff5061d09afa7fda51a68a20a5b630369a3_18_05_15_19_43_13/mongosql_auth.dll'
        chmod +x mongosql_auth.dll
    elif [ "$VARIANT" = 'macos' ]; then
        curl -O 'https://s3.amazonaws.com/mciuploads/mongosql-auth-c/artifacts/full_matrix__os~osx/mongosql_auth_c_full_matrix__os~osx_compile_1bc68011eb66f0fa148cb850db2c80b1819c41a4_18_05_11_17_03_42/mongosql_auth.so'
        chmod +x mongosql_auth.so
    elif [ "$VARIANT" = 'rhel62' ]; then
        curl -O 'https://s3.amazonaws.com/mciuploads/mongosql-auth-c/artifacts/full_matrix__os~rhel70/mongosql_auth_c_full_matrix__os~rhel70_compile_c650aff5061d09afa7fda51a68a20a5b630369a3_18_05_15_19_43_13/mongosql_auth.so'
        chmod +x mongosql_auth.so
    elif [ "$VARIANT" = 'ubuntu1404' ]; then
        curl -O 'https://s3.amazonaws.com/mciuploads/mongosql-auth-c/artifacts/full_matrix__os~ubuntu1404-64/mongosql_auth_c_full_matrix__os~ubuntu1404_64_compile_c650aff5061d09afa7fda51a68a20a5b630369a3_18_05_15_19_43_13/mongosql_auth.so'
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
