#!/bin/bash

. "$(dirname "$0")/platforms.sh"
. "$(dirname "$0")/prepare-shell.sh"


    set -o errexit
    echo "comparing benchmark results..."

    # Turn off GO111MODULE support and remove the -mod=vendor Go flag from the
    # environment. In module mode, go get will fail because it will attempt to
    # download all project dependencies (some of which are inaccessible). With
    # -mod=vendor set, go get is disabled and will therefore fail. We set that
    # flag globally because we use modules but vendor our dependencies for the
    # repository.
    export GO111MODULE="off"
    export GOFLAGS=""

    which benchstat > /dev/null 2>&1 || go get -u golang.org/x/perf/cmd/benchstat

    # download old file if we don't have it
    [ -f "$OLD" ] || curl -o "$OLD" "$S3_URI" || true

    # if we still don't have the file, then don't bother checking for regressions
    if [ ! -f "$OLD" ]; then
        echo "couldn't download old benchmark results: skipping regression detection"
        exit 0
    fi

    statfile="$ARTIFACTS_DIR/out/benchstat.out"
    benchstat "$OLD" "$NEW" > "$statfile"
    regressions="$(awk '{print $(NF-2);}' < "$statfile" | sed -e '/+/!d' -e '/+[0-4]\./d' | wc -l)"

    if [ "$regressions" != "0" ]; then
        echo "found $regressions regressions of at least 5%"
        cat "$statfile"
        exit 1
    fi
    echo "no regressions found"

    echo "done comparing benchmark results"




