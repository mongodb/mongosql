#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

get_aggregations_run_on_port() {
    numAggregationsRunOnNode=0
    # Get the node's total aggregations run.
    if $mongoBin "mongodb://127.0.0.1:$1" --eval "db.serverStatus().metrics.commands.aggregate" | grep -q total
    then
        REGEX_AGG_TOTAL="[0-9]"
        numAggregationsRunOnNode=$($mongoBin "mongodb://127.0.0.1:$1" --eval "db.serverStatus().metrics.commands.aggregate.total" | grep NumberLong | grep -o $REGEX_AGG_TOTAL)
    fi

    echo "$numAggregationsRunOnNode"
}

(
    set -o errexit
    
    echo "Running mongo uri read preference test..."
    
    mongoBin=$ARTIFACTS_DIR/mongodb/bin/mongo

    # Ensure sampling occurs before we query the data (2s sampling interval).
    sleep 5s

    numAggregationsToRun=3
    echo "Running $numAggregationsToRun sql queries that should result in reads on nodes..."
    
    counter=0
    # Run an sql query that results in a read pipeline.
    while [ $counter -lt $numAggregationsToRun ]; do
        mysql $CLIENT_ARGS -e "use test; select * from test1;"
        let counter=counter+1
    done
    
    echo "Evaluating which node handled the reads..."
    
    numAggregationsRunOnSecondaries=0
    numAggregationsRunOnPrimary=0
    for ip in "27017" "27018" "27019"
    do
        numAggregationsRun=$(get_aggregations_run_on_port "$ip")
        if $mongoBin "mongodb://127.0.0.1:$ip" --eval "db.isMaster()" | grep "primary" | grep -q $ip
        then
            # Get the primary node's total aggregations run.
            numAggregationsRunOnPrimary=$numAggregationsRun
        else
            # Sum the secondary nodes' total aggregations run.
            numAggregationsRunOnSecondaries=$(($numAggregationsRunOnSecondaries + $numAggregationsRun))
        fi
    done

    # Check if the count of aggregations run on the primary is expected.
    if [ "$EXPECT_PRIMARY_READS" == "1" ]
    then
      if [ "$numAggregationsRunOnPrimary" == "0" ]
      then
          echo "expected read aggregations to be run on primary"
          exit 1
      fi
    else
      if [ "$numAggregationsRunOnPrimary" != "0" ]
      then
          echo "expected read aggregations not to be run on primary"
          exit 1
      fi
    fi

    # Check if the count of aggregations run on the secondaries is expected.
    if [ "$EXPECT_SECONDARY_READS" == "1" ]
    then
      if [ "$numAggregationsRunOnSecondaries" == "0" ]
      then
          echo "expected read aggregations to be run on secondaries"
          exit 1
      fi
    else
      if [ "$numAggregationsRunOnSecondaries" != "0" ]
      then
          echo "expected read aggregations not to be run on secondaries"
          exit 1
      fi
    fi

    echo "Done running the mongo uri read preference test."
    
) > $LOG_FILE 2>&1

print_exit_msg
