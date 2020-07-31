#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    set -o errexit
    set -o verbose

    [ ! -d ${ARTIFACTS_DIR}/mongosql-auth-java ] && git clone git://github.com/mongodb/mongosql-auth-java.git ${ARTIFACTS_DIR}/mongosql-auth-java

    cd $ARTIFACTS_DIR/mongosql-auth-java

    username=${MONGO_USERNAME:-bob}
    password=${MONGO_PASSWORD:-pwd123}
    ./gradlew -version
    ./gradlew --stacktrace --info \
              -Porg.mongodb.test.user=$username \
              -Porg.mongodb.test.password=$password test




