#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    echo "running logging newline test..."

    sleep 1 # wait for logs to flush
   
    #grep -c returns non-0 exit code if it counts 0, leave off set -e until after greps
    num_cr="$(grep -c $'\r' $ARTIFACTS_DIR/log/mongosqld.log)"
    num_nl="$(grep -c $'\n' $ARTIFACTS_DIR/log/mongosqld.log)"
    
    set -e

    if [ "Windows_NT" == "$OS" ]; then
        if [ "$num_cr" != "$num_nl" ]; then
            echo "On Windows, expected $num_nl carriage returns (CR) in log files, got $num_cr"
            exit 1
        fi
    else
        if [ "$num_cr" != "0" ]; then
            echo "On non-Windows OS, expected 0 carriage returns (CR) in log files, got $num_cr"
            exit 1
        fi
    fi

    echo "done running logging newline test"

) > $LOG_FILE 2>&1

print_exit_msg
