#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    which yamllint > /dev/null 2>&1 || pip install --user yamllint
    export PATH="$PATH:`python -m site --user-base`/bin"
    export LANG=en_US.UTF-8
    export LC_ALL=en_US.UTF-8
    cd $PROJECT_DIR
    echo $PROJECT_DIR
    yamllint -c $PROJECT_DIR/.yamllint testdata

) > $LOG_FILE 2>&1

print_exit_msg
