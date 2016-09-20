#!/bin/sh
set -o errexit
tags=""
if [ ! -z "$1" ]
  then
  	tags="$@"
fi

# make sure we're in the directory where the script lives
SCRIPT_DIR="$(cd "$(dirname ${BASH_SOURCE[0]})" && pwd)"
cd $SCRIPT_DIR

sed -i.bak -e "s/built-without-version-string/$(git describe)/" \
           -e "s/built-without-git-spec/$(git rev-parse HEAD)/" \
           common/version.go

# remove stale packages
rm -rf vendor/pkg
mkdir -p bin

echo "Building mongosqld..."
go build -o "bin/mongosqld" -tags "$tags" "main/sqlproxy.go"
./bin/mongosqld --version

# SSL build fails on OSX - see https://jira.mongodb.org/browse/TOOLS-1145
echo "\nBuilding mongodrdl..."
go build -o "bin/mongodrdl" -tags "$tags" "mongodrdl/main/mongodrdl.go"
./bin/mongodrdl --version

mv -f common/version.go.bak common/version.go
