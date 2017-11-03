#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    echo "getting go mysql driver..."
    go get github.com/go-sql-driver/mysql
    echo "ensuring schema availability..."
    go run $(dirname $0)/ensure-schema-available.go
) > $LOG_FILE 2>&1

print_exit_msg
