#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "restoring $SUITE data..."

    cd "$PROJECT_DIR"

    go test -v \
        -run $^ \
        -timeout 4h \
        $VERSION_FLAG \
        -restoreData "$SUITE"

    echo "done restoring $SUITE data"

) > $LOG_FILE 2>&1

print_exit_msg
