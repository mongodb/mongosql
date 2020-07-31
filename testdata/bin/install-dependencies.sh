#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    set -o errexit
    echo "installing mysql shell..."

    case $PUSH_NAME in
    linux)
        url="http://noexpire.s3.amazonaws.com/sqlproxy/data/mysql-5.7.21-linux-glibc2.12-x86_64.tar.gz"
        ;;
    osx)
        url="http://noexpire.s3.amazonaws.com/sqlproxy/data/mysql-5.7.21-macos10.13-x86_64.tar.gz"
        ;;
    win32)
        url="http://noexpire.s3.amazonaws.com/sqlproxy/data/mysql-5.7.21-winx64.zip"
        ;;
    *)
        echo "Installing mysql shell for $PUSH_NAME currently unsupported"
        exit 1
        ;;
    esac

    cd $ARTIFACTS_DIR
    curl -O $url \
         --silent \
         --fail \
         --max-time 60 \
         --retry 5 \
         --retry-delay 0

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




