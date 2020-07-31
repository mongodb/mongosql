#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    set -o errexit
    echo "downloading tableau test data..."

    data_dir="$PROJECT_DIR/testdata/resources/data"

    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/attendees-new.bson.archive.gz --output "$data_dir/attendees.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/flights201406.bson.archive.gz --output "$data_dir/flights201406.bson.archive.gz"

    echo "done downloading tableau test data"




