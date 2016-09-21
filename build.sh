#!/bin/sh
set -o errexit

# make sure we're in the directory where the script lives
SCRIPT_DIR="$(cd "$(dirname ${BASH_SOURCE[0]})" && pwd)"
cd $SCRIPT_DIR

sed -i.bak -e "s/built-without-version-string/$(git describe)/" \
           -e "s/built-without-git-spec/$(git rev-parse HEAD)/" \
           common/version.go

# remove stale packages
rm -rf vendor/pkg
mkdir -p bin

# for users on go1.5
export GO15VENDOREXPERIMENT=1

echo "Building mongosqld..."
go build -o "bin/mongosqld" -tags "ssl sasl" "main/sqlproxy.go"
./bin/mongosqld --version

# SSL build fails on OS X 10.11 (TOOLS-995)
echo "\nBuilding mongodrdl..."
go build -o "bin/mongodrdl" -tags "ssl sasl" "mongodrdl/main/mongodrdl.go"
./bin/mongodrdl --version

mv -f common/version.go.bak common/version.go
