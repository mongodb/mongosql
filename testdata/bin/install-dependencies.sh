#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "installing mysql shell..."

    case $PUSH_NAME in
    linux)
        url="https://cdn.mysql.com//Downloads/MySQL-5.7/mysql-5.7.19-linux-glibc2.12-x86_64.tar.gz"
        ;;
    osx)
        url="https://cdn.mysql.com//Downloads/MySQL-5.7/mysql-5.7.19-macos10.12-x86_64.tar.gz"
        ;;
    win32)
        url="https://cdn.mysql.com//Downloads/MySQL-5.7/mysql-5.7.19-winx64.zip"
        ;;
    *)
        echo "Installing mysql shell for $PUSH_NAME currently unsupported"
        exit 1
        ;;
    esac

    cd $ARTIFACTS_DIR
    curl -O $url

    if [ "$PUSH_NAME" = "win32" ]; then
        unzip mysql*.zip
        rm mysql*.zip
        mv mysql* mysql
        mv mysql/bin/mysql.exe mysql/bin/mysql
    else
        tar xzvf mysql*.tar.gz
        rm mysql*.tar.gz
        mv mysql* mysql
    fi

    echo "done installing mysql shell"

) > $LOG_FILE 2>&1

print_exit_msg
