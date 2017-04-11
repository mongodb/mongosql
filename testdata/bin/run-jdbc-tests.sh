#!/bin/bash

. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    set -o verbose

    [ ! -d ${ARTIFACTS_DIR}/mongosql-java-auth ] && git clone git://github.com/jyemin/mongosql-java-auth.git ${ARTIFACTS_DIR}/mongosql-java-auth

    cd $ARTIFACTS_DIR/mongosql-java-auth
    ./gradlew -version
    ./gradlew --stacktrace --info -Dorg.mongodb.test.user=bob -Dorg.mongodb.test.password=pwd123 test

) > $LOG_FILE 2>&1

print_exit_msg
