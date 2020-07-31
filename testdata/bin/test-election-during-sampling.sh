#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    set -o errexit

    echo "resampling"
    mysql -e "flush sample;" > $outFile 2>&1 &
    pid=$!

    echo "discovering primary port"
    primaryPort=$(mongo --quiet --eval "db.isMaster().primary.split(':')[1]")
    echo "primary port is $primaryPort"

    echo "triggering election"
    mongo --port $primaryPort --eval "rs.stepDown()" > $outFile 2>&1 &

    samplingEnded=0
    numSleeps=150

    set +o errexit

    while [[ $numSleeps > 0 ]]; do
        kill -0 $pid
        if [[ $? == 0 ]]; then
            samplingEnded=1
            break
        fi

        sleep 1
        numSleeps=$((numSleeps - 1))
    done

    if [[ $samplingEnded == 1 ]]; then
        echo "sampling did not hang"
    else
        echo "sampling did not finish within 150 seconds"
    fi





