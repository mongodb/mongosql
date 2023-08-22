#!/bin/bash

export PATH=$PATH:.

function download {
    echo "Downloading artifact $1"
    curl -LO $1 \
     --fail \
     --max-time 60 \
     --retry 5 \
     --retry-delay 0
}

function get_version {
    grep "export $1=" ../../evergreen.yml | head -1 | cut -d\" -f2
}

MONGODB_VERSION="mongodb-linux-x86_64-ubuntu2204-$(get_version MONGODB_TEST_VERSION)"
MONGOSH_VERSION="mongosh-$(get_version MONGOSH_VERSION)-linux-x64"
DBTOOLS_VERSION="mongodb-database-tools-ubuntu2204-x86_64-$(get_version DB_TOOLS_VERSION)"

MONGODB_URL="https://fastdl.mongodb.org/linux/${MONGODB_VERSION}.tgz"
DBTOOLS_URL="https://fastdl.mongodb.org/tools/db/${DBTOOLS_VERSION}.tgz"

MONGODB=$MONGODB_VERSION/bin/mongod
MONGOSH=$MONGOSH_VERSION/bin/mongosh
MONGORESTORE=${DBTOOLS_VERSION}/bin/mongorestore

if [[ ! -f "$MONGOSH" ]]; then
  echo "${MONGOSH_VERSION} not found! Downloading..."
  download "https://downloads.mongodb.com/compass/${MONGOSH_VERSION}.tgz"
  tar zxf "${MONGOSH_VERSION}.tgz"
fi
