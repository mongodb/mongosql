#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    MYSQL_VENDOR_PATH=github.com/go-sql-driver/mysql
    TARGET_DIR=$GOPATH/src/$MYSQL_VENDOR_PATH
    mkdir -p $TARGET_DIR
    cp -rf $PROJECT_DIR/vendor/$MYSQL_VENDOR_PATH/* $TARGET_DIR/

    echo "ensuring schema availability..."
    go run $(dirname $0)/ensure/schema-available.go
) > $LOG_FILE 2>&1

print_exit_msg
