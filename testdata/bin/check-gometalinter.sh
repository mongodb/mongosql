#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    lint=gometalinter.v2
    which $lint > /dev/null 2>&1 || go get -u gopkg.in/alecthomas/gometalinter.v2
    $lint --install
    $lint $(find . -name '*.go' | grep -v './vendor' | grep -v './testdata' | xargs -L1 dirname | uniq)

) > $LOG_FILE 2>&1

print_exit_msg
