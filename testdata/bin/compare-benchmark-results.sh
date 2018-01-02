#!/bin/bash

. "$(dirname "$0")/platforms.sh"
. "$(dirname "$0")/prepare-shell.sh"

(
    set -o errexit
    echo "comparing benchmark results..."

    which benchstat > /dev/null 2>&1 || go get -u golang.org/x/perf/cmd/benchstat

    statfile="$ARTIFACTS_DIR/out/benchstat.out"
    benchstat "$OLD" "$NEW" > "$statfile"
    regressions="$(awk '{print $(NF-2);}' < "$statfile" | grep '^+' | wc -l)"

    if [ "$regressions" != "0" ]; then
        echo "found $regressions regressions"
        cat "$statfile"
        exit 1
    fi
    echo "no regressions found"

    echo "done comparing benchmark results"

) > "$LOG_FILE" 2>&1

print_exit_msg
