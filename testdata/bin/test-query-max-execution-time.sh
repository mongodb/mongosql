#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    echo "Running query max execution time test..."

    set +o errexit

    # Ensure sampling occurs before we query the data (2s sampling interval).
    sleep 3s

    # Set the `max_execution_time` variable.
    mysql $CLIENT_ARGS -e "set @@global.max_execution_time = $MAX_EXECUTION_TIME;"

    output=$(mysql $CLIENT_ARGS -e "$QUERY" 2>&1)
    code=$?

    set -o errexit

    timeoutErrorMessage="ERROR 3024 (HY000) at line 1: Query execution was interrupted, maximum statement execution time exceeded"

    if [ "$EXPECT_ERROR" == "1" ] && [ "$output" != "$timeoutErrorMessage" ]
    then
      echo "expected query to return an error because it did not succeed in under the max execution time"
      echo "output: $output"
      exit 1
    elif [ "$EXPECT_ERROR" == "0" ] && [ "$code" != "0" ]
    then
      echo "expected query to succeed in under the max execution time"
      echo "output: $output"
      exit 1
    fi

    echo "Done running the query max execution time test."

) > $LOG_FILE 2>&1

print_exit_msg
