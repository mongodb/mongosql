#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "downloading blackbox test data..."

    data_dir="$PROJECT_DIR/testdata/resources/data"

    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/attendees.bson.archive.gz --output "$data_dir/attendees.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/flights201406.bson.archive.gz --output "$data_dir/flights201406.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Batters.bson.archive.gz --output "$data_dir/Batters.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Calcs.bson.archive.gz --output "$data_dir/Calcs.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/DateTime.bson.archive.gz --output "$data_dir/DateTime.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Election.bson.archive.gz --output "$data_dir/Election.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Fischeriris.bson.archive.gz --output "$data_dir/Fischeriris.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Loan.bson.archive.gz --output "$data_dir/Loan.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/NumericBins.bson.archive.gz --output "$data_dir/NumericBins.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/DecimalRei.bson.archive.gz --output "$data_dir/Rei.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/SeattleCrime.bson.archive.gz --output "$data_dir/SeattleCrime.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Securities.bson.archive.gz --output "$data_dir/Securities.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/SpecialData.bson.archive.gz --output "$data_dir/SpecialData.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/DecimalStaples.bson.archive.gz --output "$data_dir/Staples.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/Starbucks.bson.archive.gz --output "$data_dir/Starbucks.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/DecimalUTStarcom.bson.archive.gz --output "$data_dir/UTStarcom.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/xy.bson.archive.gz --output "$data_dir/xy.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/bigarray.bson.archive.gz --output "$data_dir/bigarray.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/bigcoll.bson.archive.gz --output "$data_dir/bigcoll.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/bignestedarray.bson.archive.gz --output "$data_dir/bignestedarray.bson.archive.gz"
    curl -s http://noexpire.s3.amazonaws.com/sqlproxy/data/bigobjarray.bson.archive.gz --output "$data_dir/bigobjarray.bson.archive.gz"

    echo "done downloading blackbox test data"

) > $LOG_FILE 2>&1

print_exit_msg
