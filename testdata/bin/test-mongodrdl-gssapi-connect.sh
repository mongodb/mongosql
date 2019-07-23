#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(

    echo "running drdl gssapi test"
    output=$($ARTIFACTS_DIR/bin/mongodrdl $DRDL_ARGS -o $ARTIFACTS_DIR/out/gssapi.drdl 2>&1)
    code=$?

    set -o errexit

    output=$(echo "$output" | sed 's/\$//g')

    if [ "$code" != "$EXPECTED_STATUS" ]; then
        echo "expected connection to exit '$EXPECTED_STATUS' but exited but '$code'"
        echo "output: $output"
        exit 1
    fi

    set +o errexit

    # if the connection succeeded, check that we got the expected drdl file
    if [ "$EXPECTED_STATUS" = '0' ]; then
        comp_output=$(cmp $ARTIFACTS_DIR/out/gssapi.drdl $PROJECT_DIR/testdata/resources/configs/gssapi.drdl 2>&1)
        comp_code=$?

        set -o errexit

        if [ "$comp_code" != 0 ]; then
            echo "output drdl file does not match expected"
            echo "output: $comp_output"
            exit 1
        fi
    fi

    echo "done testing drdl gssapi"

) > $LOG_FILE 2>&1

print_exit_msg
