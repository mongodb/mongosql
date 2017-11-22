#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

run_test() {
    # Prefix each query with a random sentinel to search for in the output of db.currentOp().
    # This is the most reliable way to identify multiple concurrent queries
    rand="it$i-job$j-rand$RANDOM"
    targetQuery=$(echo $cmd | sed "s/select /select \"$rand\" as rand,/g")
    echo "target query: $targetQuery"

    set +o errexit
    # Fork the target query and pipe the output to a file for later.
    outFile=".kill_query_test-it$i-job$j"
    echo "" > $outFile
    # The first line of output should be the connection ID we want to kill
    # Run the query as user 1.
    MYSQL_PWD=$USER1_PWD mysql $CLIENT_ARGS --disable-column-names --unbuffered -e "select connection_id(); $targetQuery" > $outFile 2>&1 &
    pid=$!

    echo "forked query for job $j pid: $pid"

    # Read a connection ID from the first line of the query output
    numSleeps=0
    while [ "$line" == "" ]; do
        # To improve test reliability (on Windows in particular), increase the maximum timeout to 30 second before failing.
        if [ $numSleeps -gt 30 ]; then
            echo "Query has not returned any results after $numSleeps seconds"
            exit 1
        fi
        read -r line < $outFile
        sleep 1
        numSleeps=$((numSleeps + 1))
    done

    connId=$line
    if [[ ! $connId =~ ^[0-9]+$ ]]; then
        echo "Connection ID for query $j isn't a number: $connId"
        exit 1
    fi

    if [ "$KILL_CONN" == "true" ]; then
        # Kill the connection
        killQuery="kill $connId"
    else
        # Kill the query
        killQuery="kill query $connId"
    fi

    # If this test was run with two different users, the second user should not be able to kill the first's query.
    if [ "$CLIENT_ARGS" != "$USER2_CLIENT_ARGS" ]; then

        echo "Attempting to kill query $j with user 2. Command: '$killQuery'"
        # Attempt to kill the query with user 2
        killResult=$(MYSQL_PWD=$USER2_PWD mysql $USER2_CLIENT_ARGS -e "$killQuery" 2>&1)
        killCode=$?

        echo "kill result for user 2: $killResult"
        if [ "$killCode" != "1" ]; then
            echo "mysql kill command '$killQuery' exited with code $killCode run as user 2. output: $killResult"
            echo "target query output: $outFile"
            exit 1
        fi
    fi

    echo "Attempting to kill query $j with user 1. Command: '$killQuery'"

    startTime=$(date +%s)
    # Kill the query with user 1
    killResult=$(MYSQL_PWD=$USER1_PWD mysql $CLIENT_ARGS -e "$killQuery" 2>&1)
    killCode=$?
    # If the the kill query failed but the exit code of the target query is 0, it completed successfully earlier than expected
    if [ "$killCode" != "0" ]; then
        echo "mysql kill command '$killQuery' exited with code $killCode. output: $killResult"
        echo "target query output: $outFile"
        exit 1
    fi


    # Wait for the target query to finish.
    echo "waiting for target query $j pid: $pid to exit"
    wait $pid
    code=$?
    set -o errexit

    endTime=$(date +%s)
    killTime=$((endTime-startTime))

    # If the exit code is 0, then the query completed too early
    # The maximum acceptable time is 30 seconds to improve test reliablility.
    if [ $killTime -gt 30 ] && [ "$code" != "0" ]; then
        echo "query $j took too long to kill: $killTime seconds"
        exit 1
    fi

    # Read the second line of output from the target query
    result=$(sed -n '2p' $outFile)
    if [ "$code" == "1" ]; then
        if [ "$result" != "$EXPECTED_ERROR" ]; then
            echo "expected '$EXPECTED_ERROR', got '$result' for target query: '$targetQuery' output: $(cat $outFile)"
            exit 1
        fi
    else
        echo "target query '$cmd' exited with code $code. Output: $result"
        echo "WARNING: the query being testing didn't take as long as expected"
    fi
    rm $outFile

    # Check that no query is running on MongoDB
    mongoBin=$ARTIFACTS_DIR/mongodb/bin/mongo
    currentOps=$($mongoBin $MONGO_CLIENT_ARGS --quiet --eval "db.currentOp({'planSummary': {\$exists: 1}})")
    if [[ $currentOps == *"\"\$literal\" : \"$rand\""* ]]; then
        echo "The following MongoDB operation(s) for job $j is/are still runnning:"
        echo $currentOps
        exit 1
    fi
}

(
    set -o errexit
    echo "running kill query test..."

    cmd="$(echo "$QUERY" | sed 's/,,/;/g')"

    USER1_PWD=$MYSQL_PWD

    if [ -z $ITERATIONS ]; then
        ITERATIONS=10
    fi
    if [ -z $PROCS ]; then
        PROCS=10
    fi
    # Allow this test to be repeated multiple times
    for ((i=0;i<$ITERATIONS;i++)); do

        jobArr=()
        # Create parallel jobs
        for ((j=0;j<$PROCS;j++)); do
            run_test &
            pid=$!
            jobArr[$j]=$pid
            echo "iteration $i: created job $j pid: $pid"
        done

        for ((j=0;j<$PROCS;j++)); do
            echo "waiting for child: ${jobArr[j]}"
            wait ${jobArr[j]}
            exitCode=$?
            if [ "$exitCode" != 0 ]; then
                echo "child exited with code $exitCode"
                exit 1
            fi
        done

        echo "Jobs: $(jobs -p)"
    done

    echo "done running kill query test"

) > $LOG_FILE 2>&1



print_exit_msg
