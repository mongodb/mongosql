#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    echo "Running query max execution time test..."

    set +o errexit

    # Set the `max_execution_time` global variable.
    mysql $CLIENT_ARGS -e "set @@global.max_execution_time = 2000;"

    # Ensure the `max_execution_time` variable was set.
    output=$(mysql $CLIENT_ARGS -e "SELECT @@global.max_execution_time;" | grep 2000 2>&1)
    code=$?
    if [ "$code" != "0" ]
    then
      echo "unable to set max_execution_time variable"
      exit 1
    fi

    output=$(mysql $CLIENT_ARGS -e "SELECT SLEEP($SLEEP_TIME);" 2>&1)
    code=$?

    set -o errexit

    timeoutErrorMessage="ERROR 3024 (HY000) at line 1: Query execution was interrupted, maximum statement execution time exceeded"

    if [ "$EXPECT_ERROR" == "1" ] && [ "$output" != "$timeoutErrorMessage" ]
    then
      echo "expected query to return an error because it did not succeed in under the max execution time"
      exit 1
    elif [ "$EXPECT_ERROR" == "0" ] && [ "$code" != "0" ]
    then
      echo "expected query to succeed in under the max execution time"
      exit 1
    fi

    echo "Done running the query max execution time test."

) > $LOG_FILE 2>&1

print_exit_msg
