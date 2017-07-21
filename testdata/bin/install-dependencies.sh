#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "installing mysql shell..."

    if [ "$PUSH_NAME" != "linux" ]; then
        echo "Installing mysql shell for non-linux platforms currently unsupported"
        exit 1
    fi

    cd $ARTIFACTS_DIR
    curl -O https://cdn.mysql.com//Downloads/MySQL-5.7/mysql-5.7.19-linux-glibc2.12-x86_64.tar.gz
    tar xzvf mysql*.tar.gz
    rm mysql*.tar.gz
    mv mysql* mysql

    echo "done installing mysql shell"

) > $LOG_FILE 2>&1

print_exit_msg
