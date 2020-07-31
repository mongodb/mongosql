#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    echo "Running memory limit test for $TEST_NAME"

    cd "$PROJECT_DIR"
    mongo --eval "load('$SCRIPT_DIR/js/dropall.js'); load('$SCRIPT_DIR/js/$TEST_NAME.js')"
    res_file="$ARTIFACTS_DIR/out/$TEST_NAME.out"
    [ -e $res_file ] && rm $res_file
    (GODEBUG=gctrace=1 $ARTIFACTS_DIR/bin/mongosqld &> $res_file) &

    echo "forked job $TEST_NAME"

    # We want our gctrace output to have data points from two distinct periods
    # of time: 1. during sampling and 2. after sampling is finished, and we’re
    # issuing commands. For this reason, we want to verify that sampling has
    # finished, so we can ensure we get gctrace output while mongosqld is
    # responding to mysql commands. We know sampling has finished once
    # mongosqld starts accepting connections, so we try to connect to it
    # through the mysql shell.
    i=1
    exit_code=1

    while [ $i -lt 180 ]; do # timeout after 30 mins
        echo "loop $i"

        # Connect to mongosqld using the mysql shell and issue a few basic
        # commands. For some datasets, we observed normal memory usage during
        # sampling but high memory usage when running simple commands.
        echo "running mysql commands"
        output=$(mysql -e "show databases; use memtest; describe test;" 2>&1)

        exit_code=$?
        echo "$exit_code"
        # When the sampling is not yet finished, exit_code will be 0 and the output will start with `Error`.
        # in this case, we will try again in 10 seconds, otherwise we will break out the loop with actual result
        if [[ ($exit_code = "0" && $output != ERROR*) ]]; then
            echo "break from loop";
            break;
        fi
        sleep 10
        i=$[$i+1]
    done

    echo "killing mongosqld";
    killall mongosqld

    if [ "$exit_code" == "1" ]; then
        echo "test timed out - sampling took too long";
        exit 1;
    fi

    echo "done with test for $TEST_NAME"




