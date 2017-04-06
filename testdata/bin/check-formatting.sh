#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"

(
    echo "checking formatting..."

    cd $PROJECT_DIR

    unformatted="$(gofmt -e -l `find . -path ./vendor -prune -o -name '*.go' -print` 2>&1)"

    for fn in $unformatted; do
      echo >&2 "  Unformatted: $fn"
    done
    [ -z "$unformatted" ] || exit 1

    echo "done checking formatting"

) > $LOG_FILE 2>&1

print_exit_msg
