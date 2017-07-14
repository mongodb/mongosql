#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "downloading tableau test data..."

    data_dir="$PROJECT_DIR/testdata/resources/data"

    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/attendees.bson.gz --output "$data_dir/attendees.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/flights201406.bson.gz --output "$data_dir/flights201406.bson.gz"

    echo "done downloading tableau test data"

) > $LOG_FILE 2>&1

print_exit_msg
