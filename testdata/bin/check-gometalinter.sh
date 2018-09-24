#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    which gometalinter > /dev/null 2>&1 || go get -u github.com/alecthomas/gometalinter
    gometalinter --install
    gometalinter --deadline 150s $(find . -name '*.go' | grep -v './vendor' | grep -v './testdata' | xargs -L1 dirname | uniq)

) > $LOG_FILE 2>&1

print_exit_msg