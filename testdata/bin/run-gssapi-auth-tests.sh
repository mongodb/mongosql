#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    set -o verbose

    [ ! -d ${ARTIFACTS_DIR}/mongosql-auth-java ] && git clone git://github.com/mongodb/mongosql-auth-java.git ${ARTIFACTS_DIR}/mongosql-auth-java

    cd $ARTIFACTS_DIR/mongosql-auth-java
    ./gradlew -version

    cat << EOF > ${ARTIFACTS_DIR}/out/java.login.drivers.config
    com.sun.security.jgss.krb5.initiate {
        com.sun.security.auth.module.Krb5LoginModule required
            doNotPrompt=true useKeyTab=true keyTab="${PROJECT_DIR}/testdata/resources/gssapi/drivers.keytab" principal=drivers;
    };
EOF

    ./gradlew --stacktrace --info --rerun-tasks \
        -Pauth.login.config="file://${ARTIFACTS_DIR}/out/java.login.drivers.config" \
        -Pkrb5.kdc="ldaptest.10gen.cc" \
        -Pkrb5.realm="LDAPTEST.10GEN.CC" \
        -Porg.mongodb.test.host="localhost" \
        -Psun.security.krb5.debug=true \
        -Porg.mongodb.test.database=kerberos \
        -Porg.mongodb.test.sql="select count(*) from test" \
        -Porg.mongodb.test.user=drivers?mechanism=GSSAPI&serviceName=mongosql2 \
        test --tests "org.mongodb.mongosql.auth.plugin.MongoSqlAuthenticationPluginFunctionalTest.testSuccessfulAuthentication"

) > $LOG_FILE 2>&1

print_exit_msg
