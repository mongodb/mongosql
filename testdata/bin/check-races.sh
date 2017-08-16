#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "checking whether data races were detected..."

    log_file="$ARTIFACTS_DIR/log/races.log"
    mv $log_file* $log_file > /dev/null 2>&1 || touch $log_file

    race_lines="$(cat $log_file | wc -l)"
    if [ "$race_lines" != "0" ]; then
        echo "data races were detected:"
        cat $log_file
        exit 1
    fi

    echo "done checking for data races (none found)"

) > $LOG_FILE 2>&1

print_exit_msg
