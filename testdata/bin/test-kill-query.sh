#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

check_currentop() {
    # Check that no query is running on MongoDB
    currentOps=$($mongoBin $1 $MONGO_CLIENT_ARGS --quiet --eval "db.currentOp({'planSummary': {\$exists: 1}}).inprog")
    if [[ $currentOps == *"$rand"* ]]; then
        echo "The following MongoDB operation(s) for job $j is/are still runnning:"
        echo $currentOps
        exit 1
    fi
}

get_shard_uri() {
    host=$($mongoBin $MONGO_CLIENT_ARGS --quiet --eval "db.adminCommand({listShards: 1}).shards[$1].host")
    part1=$(echo $host | awk -F "/" '{print $1}')
    part2=$(echo $host | awk -F "/" '{print $2}')

    # The URI is not in in 'replicaset/nodeList' format, use the first parameter as the host
    if [ -z $part2 ]; then
        uri="mongodb://$part1"
    else
        uri="mongodb://$part2/?replicaSet=$part1"
    fi

    # Assert that this works.
    $mongoBin $uri $MONGO_CLIENT_ARGS --quiet --eval "db.currentOp()" > /dev/null
}

run_test() {
    # Prefix each query with a random sentinel to search for in the output of db.currentOp().
    # This is the most reliable way to identify multiple concurrent queries
    rand="it$i-job$j-rand$RANDOM"
    targetQuery=$(echo $cmd | sed "s/select /select \"$rand\" as rand,/g")
    echo "target query: $targetQuery"

    set +o errexit
    # Fork the target query and pipe the output to a file for later.
    outFile="$ARTIFACTS_DIR/.kill_query_test-it$i-job$j"
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
    if [ "$CLIENT_ARGS" != "$OTHER_CLIENT_ARGS" ]; then

        echo "Attempting to kill query $j with user 2. Command: '$killQuery'"
        # Attempt to kill the query with user 2
        killResult=$(MYSQL_PWD=$MONGO_OTHER_USER_PWD mysql $OTHER_CLIENT_ARGS -e "$killQuery" 2>&1)
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
    # The maximum acceptable time is 120 seconds to improve test reliablility.
    if [ $killTime -gt 120 ] && [ "$code" != "0" ]; then
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
    rm -rf $outFile

    # Check for currentOps on the mongo (or mongos)
    check_currentop

    # Check for currentOps on shards (if they exist)
    if [ -n "$shard1Uri" ]; then
        check_currentop $shard1Uri
    elif [ -n "$shard2Uri" ]; then
        check_currentop $shard2Uri
    fi
}

(
    set -o errexit
    echo "running kill query test..."

    # These variables will stay empty on a standalone cluster
    shard1Uri=
    shard2Uri=
    mongoBin=$ARTIFACTS_DIR/mongodb/bin/mongo
    # In our current test configuration, we know we only run with two shards  
    if [ "$TOPOLOGY" == "sharded_cluster" ]; then
        get_shard_uri 0
        shard1Uri="$uri"
        get_shard_uri 1
        shard2Uri="$uri"
        echo "shard 1: $shard1Uri"
        echo "shard 2: $shard2Uri"
    fi

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
