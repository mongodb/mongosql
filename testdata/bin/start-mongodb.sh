#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "starting mongodb..."

    MONGODB_VERSION=${MONGODB_VERSION:-latest}

    DIR=$(dirname $0)
    # Functions to fetch MongoDB binaries
    . $DIR/download-mongodb.sh

    echo "downloading mongodb $MONGODB_VERSION..."
    set_mongodb_binaries "$MONGODB_VERSION"
    echo "done downloading mongodb"

    mkdir -p "$ARTIFACTS_DIR/mlaunch"
    cd "$ARTIFACTS_DIR/mlaunch"

    if [ "$VARIANT" = '' ]; then
      mlaunch_cache_dir="$SQLPROXY_TEST_CACHE_DIR/mlaunch"
      mkdir -p "$mlaunch_cache_dir"
      cd "$mlaunch_cache_dir"
    fi

    venv='venv'
    if [ "$VARIANT" = 'centos6-perf' ]; then
      venv="$PROJECT_DIR/../../../../venv"
    fi

    # Setup or use the existing virtualenv for mtools
    if [ -f "$venv"/bin/activate ]; then
      echo 'using existing virtualenv'
      . "$venv"/bin/activate
    elif [ -f "$venv"/Scripts/activate ]; then
      echo 'using existing virtualenv'
      . "$venv"/Scripts/activate
    elif virtualenv --no-site-packages "$venv" || python -m virtualenv --no-site-packages "$venv"; then
      echo 'creating new virtualenv'
      if [ -f "$venv"/bin/activate ]; then
        . "$venv"/bin/activate
      elif [ -f "$venv"/Scripts/activate ]; then
        . "$venv"/Scripts/activate
      fi

      echo 'cloning mtools...'
      rm -rf mtools
      git clone git@github.com:rueckstiess/mtools
      cd mtools
      # We should avoid checking out the master branch because it is a dev branch
      # that has occasionally had bugs committed. This commit has worked well for us.
      git checkout e544bbced1a070d7024931e7c1736ced7d9bcdd6
      echo 'installing mtools...'
      pip install .[mlaunch]
      pip freeze
      echo 'done installing mtools'
    fi

    cd "$ARTIFACTS_DIR/mlaunch"

    mlaunch_auth_args=''
    if [ "$MONGO_AUTH" = 'auth' ]; then
      mlaunch_auth_args='--auth --username bob --password pwd123'
    fi

    mlaunch_ssl_args=''
    if [ "$MONGO_SSL" = 'ssl' ]; then
      mlaunch_ssl_args="$mlaunch_ssl_args --sslMode requireSSL"
      mlaunch_ssl_args="$mlaunch_ssl_args --sslPEMKeyFile $PROJECT_DIR/testdata/resources/x509gen/server.pem"
      mlaunch_ssl_args="$mlaunch_ssl_args --sslClientPEMKeyFile $PROJECT_DIR/testdata/resources/x509gen/client.pem"
      mlaunch_ssl_args="$mlaunch_ssl_args --sslCAFile $PROJECT_DIR/testdata/resources/x509gen/ca.pem"
      mlaunch_ssl_args="$mlaunch_ssl_args --sslWeakCertificateValidation"
    fi

    if [ "$TOPOLOGY" = 'server' ]; then
      mlaunch_topology_args='--single'
    elif [ "$TOPOLOGY" = 'replica_set' ]; then
      mlaunch_topology_args='--replicaset'
    elif [ "$TOPOLOGY" = 'sharded_cluster' ]; then
      mlaunch_topology_args='--replicaset --sharded 2'
    else
      echo "invalid topology '$TOPOLOGY'"
      exit 1
    fi

    mlaunch_storage_args=''
    if [ "$STORAGE_ENGINE" != '' ]; then
      mlaunch_storage_args='--storageEngine inMemory'
    fi

    # we run mongod with compression enabled when possible
    if [ "$MONGODB_VERSION" == '3.2' ]; then
      # 3.2 does not support the --networkMessageCompressors flag
      mlaunch_compression_args=""
    elif [ "$MONGODB_VERSION" == '3.4' ]; then
      # 3.4 does not support zlib
      mlaunch_compression_args="--networkMessageCompressors snappy"
    else
      # >= 3.6 support zlib,snappy
      mlaunch_compression_args="--networkMessageCompressors zlib,snappy"
    fi

    mlaunch_args=''
    mlaunch_args="$mlaunch_args --verbose"
    mlaunch_args="$mlaunch_args --binarypath $ARTIFACTS_DIR/mongodb/bin"

    mlaunch_init_args="$mlaunch_args $mlaunch_auth_args $mlaunch_ssl_args $mlaunch_topology_args $mlaunch_storage_args $mlaunch_compression_args"
    echo "mlaunch init args: $mlaunch_init_args"

    mlaunch init $mlaunch_init_args
    mlaunch list $mlaunch_args

    echo "done starting mongodb"

) > $LOG_FILE 2>&1

print_exit_msg
