#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
       set -o errexit

       hash golangci-lint > /dev/null 2>&1 || $PROJECT_DIR/testdata/bin/golangci-lint.sh -b $GOBIN v1.12.2
       golangci-lint run --config $PROJECT_DIR/.golangci.yml

) > $LOG_FILE 2>&1

print_exit_msg
