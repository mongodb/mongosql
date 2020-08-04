#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

run_tls_version_test() {
   output=$(echo "Q" | $OPENSSL_CLIENT s_client -starttls mysql $1 -connect 127.0.0.1:3307 2>&1)
   code=$?
   expected=$2

   if [ "$code" != "$expected" ]; then
        echo "expected connection to exit '$expected', but it exited '$code'"
        echo "output: $output"
        echo "tls version": $1
        exit 1
   fi
}

(
    set -o errexit
    OPENSSL_DIR="$PROJECT_DIR/testdata/artifacts/bin"

    if [ "$VARIANT" = "ubuntu1404" ]
    then
        curl -s https://noexpire.s3.amazonaws.com/sqlproxy/binary/ubuntu/openssl --output "$OPENSSL_DIR/openssl"
        curl -s https://noexpire.s3.amazonaws.com/sqlproxy/binary/ubuntu/libssl.so.1.1 --output "$OPENSSL_DIR/libssl.so.1.1"
        curl -s https://noexpire.s3.amazonaws.com/sqlproxy/binary/ubuntu/libcrypto.so.1.1 --output "$OPENSSL_DIR/libcrypto.so.1.1"
        export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:"$OPENSSL_DIR"
    else
        curl -s https://noexpire.s3.amazonaws.com/sqlproxy/binary/osx/openssl --output "$OPENSSL_DIR/openssl"
        curl -s https://noexpire.s3.amazonaws.com/sqlproxy/binary/osx/libssl.1.1.dylib --output "$OPENSSL_DIR/libssl.1.1.dylib"
        curl -s https://noexpire.s3.amazonaws.com/sqlproxy/binary/osx/libcrypto.1.1.dylib --output "$OPENSSL_DIR/libcrypto.1.1.dylib"
        export DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH:"$OPENSSL_DIR"
    fi

    echo "running tls connection test..."
    OPENSSL_CLIENT=$OPENSSL_DIR/openssl
    chmod 755 $OPENSSL_CLIENT

    # "-no_tls1" is analogous to "-ssl3" however, we don't use the latter since it's deprecated.
    tls_version_flags=("-no_tls1 -no_tls1_1 -no_tls1_2 -no_tls1_3" "-tls1" "-tls1_1" "-tls1_2")

    for index in "${!flags[@]}"
    do
        if [[ "$index" -lt "$MIN_TLS_VERSION" ]]
        then
            run_tls_version_test "${tls_version_flags[$index]}" 1
        else
            run_tls_version_test "${tls_version_flags[$index]}" 0
        fi
    done
    echo "done running tls connection test"
) > $LOG_FILE 2>&1


print_exit_msg
