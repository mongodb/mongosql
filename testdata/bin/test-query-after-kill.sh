#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    set -o errexit

    echo "Testing to see if a query can still be run after a query is killed..."

    # Ensure sampling occurs before we query the data (2s sampling interval).
    sleep 3s

    SECONDS=0

    # Kill the query on connection 2 after waiting 10 seconds.
    mysql $CLIENT_ARGS -e "select sleep(10); kill query 2;" &

    # Sleep for 5s to ensure the command we want to kill recieves connection id 2.
    sleep 5s

    echo "Running mysql queries that we will interrupt..."
    # The force flags ensures mysql will try the 2nd query after the 1st one is interrupted.
    output=$(mysql $CLIENT_ARGS --force <<< "select sleep(10); show databases;" 2>&1)

    echo "Queries finished in $SECONDS seconds."
    
    # Get the results from our mysql queries' output.
    # The first line contain the error - the next lines the databases.
    interruptOutput=$(echo "$output" | sed -n '1p')
    databaseOutput=$(echo "$output" | sed -n '2,$p')

    interruptMessage="ERROR 1317 (70100) at line 1: Query execution was interrupted"
    databaseExpectedInOutput="INFORMATION_SCHEMA"

    if [ "$interruptOutput" != "$interruptMessage" ]
    then
        echo "failed to find query interrupted message"
        echo "found '$interruptOutput' expected '$interruptMessage'"
        exit 1
    elif ! echo "$databaseOutput" | grep -q $databaseExpectedInOutput
    then
        echo "expected to find database '$databaseExpectedInOutput' in mysql output: '$databaseOutput'"
        exit 1
    fi

    echo "Done testing to see if a query can still be run after a query is killed."




