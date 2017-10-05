#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    dataset=${TPCH_DATASET:-micro}

    echo "downloading tpch $dataset test data..."

    data_dir="$PROJECT_DIR/testdata/resources/data"
    target="$data_dir/tpch-$dataset.bson.gz"

    case "$dataset" in
        micro*)
            url="http://noexpire.s3.amazonaws.com/sqlproxy/data/tpch_small.bson.gz"
        ;;
        normalized*)
            url="http://noexpire.s3.amazonaws.com/sqlproxy/data/tpch_full_normalized.bson.gz"
        ;;
        denormalized*)
            url="http://noexpire.s3.amazonaws.com/sqlproxy/data/tpch_full_denormalized.bson.gz"
        ;;
        *)
            echo "no tpch dataset named '$dataset'"
            exit 1
        ;;
    esac

    curl -s "$url" --output "$data_dir/tpch-$dataset.bson.gz"

    if [ "Windows_NT" = "$OS" ]; then
        mv "$target" "$data_dir/tpch.bson.gz"
    else
        ln -sf "$target" "$data_dir/tpch.bson.gz"
    fi

    echo "done downloading tpch $dataset test data"

) > $LOG_FILE 2>&1

print_exit_msg
