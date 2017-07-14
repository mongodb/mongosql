#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "downloading blackbox test data..."

    data_dir="$PROJECT_DIR/testdata/resources/data"

    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/attendees.bson.gz --output "$data_dir/attendees.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/flights201406.bson.gz --output "$data_dir/flights201406.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Batters.bson.gz --output "$data_dir/Batters.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Calcs.bson.gz --output "$data_dir/Calcs.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/DateTime.bson.gz --output "$data_dir/DateTime.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Election.bson.gz --output "$data_dir/Election.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Fischeriris.bson.gz --output "$data_dir/Fischeriris.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Loan.bson.gz --output "$data_dir/Loan.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/NumericBins.bson.gz --output "$data_dir/NumericBins.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/DecimalRei.bson.gz --output "$data_dir/Rei.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/SeattleCrime.bson.gz --output "$data_dir/SeattleCrime.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Securities.bson.gz --output "$data_dir/Securities.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/SpecialData.bson.gz --output "$data_dir/SpecialData.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/DecimalStaples.bson.gz --output "$data_dir/Staples.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Starbucks.bson.gz --output "$data_dir/Starbucks.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/DecimalUTStarcom.bson.gz --output "$data_dir/UTStarcom.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/xy.bson.gz --output "$data_dir/xy.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/bigarray.bson.gz --output "$data_dir/bigarray.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/bigcoll.bson.gz --output "$data_dir/bigcoll.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/bignestedarray.bson.gz --output "$data_dir/bignestedarray.bson.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/bigobjarray.bson.gz --output "$data_dir/bigobjarray.bson.gz"

    echo "done downloading blackbox test data"

) > $LOG_FILE 2>&1

print_exit_msg
