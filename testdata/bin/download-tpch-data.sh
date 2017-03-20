#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "downloading tpch $TPCH_DATASET test data..."

    data_dir="$PROJECT_DIR/testdata/resources/data"
    target="$data_dir/tpch-$TPCH_DATASET.bson.gz"

    case "$TPCH_DATASET" in
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
            echo "no tpch dataset named '$TPCH_DATASET'"
            exit 1
        ;;
    esac

    curl -s "$url" --output "$data_dir/tpch-$TPCH_DATASET.bson.gz"
    ln -sf "$target" "$data_dir/tpch.bson.gz"

    echo "done downloading tpch $TPCH_DATASET test data"

) > $LOG_FILE 2>&1

print_exit_msg
