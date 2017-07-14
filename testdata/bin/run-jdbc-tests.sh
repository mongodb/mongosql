#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    set -o verbose

    [ ! -d ${ARTIFACTS_DIR}/mongosql-auth-java ] && git clone git://github.com/mongodb/mongosql-auth-java.git ${ARTIFACTS_DIR}/mongosql-auth-java

    cd $ARTIFACTS_DIR/mongosql-auth-java
    ./gradlew -version
    ./gradlew --stacktrace --info -Porg.mongodb.test.user=bob -Porg.mongodb.test.password=pwd123 test

) > $LOG_FILE 2>&1

print_exit_msg
