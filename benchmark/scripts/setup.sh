#!/bin/bash

. ./prepare-env.sh

# Default scale factor
scale_factor="0.001"

# Check if an argument is provided
if [ "$#" -eq 1 ]; then
    case "$1" in
        0.001|0.01|0.1|1)
            scale_factor="$1"
            ;;
        *)
            echo "Invalid scale factor. Allowed values: 0.001, 0.01, 0.1, 1. Defaulting to 0.001."
            ;;
    esac
fi

rustup_url="https://sh.rustup.rs"

dataset_filename="tpch-${scale_factor}-normalized-indexed.archive.gz"
dataset_url="https://mongosql-noexpire.s3.us-east-2.amazonaws.com/tpch/${dataset_filename}"

# MongoDB setup
if [[ ! -f "$MONGODB" ]]; then
  echo "${MONGODB_VERSION} not found! Downloading..."
  download $MONGODB_URL
  tar zxvf ${MONGODB_VERSION}.tgz
fi
mkdir -p data_db
$MONGODB --dbpath data_db --logpath mongodb.log &

# MongoDB Database Tools setup
if [[ ! -f "$MONGORESTORE" ]]; then
  echo "${MONGORESTORE} not found! Downloading..."
  download $DBTOOLS_URL
  tar zxvf ${DBTOOLS_VERSION}.tgz
fi

# Rust setup and pipeline generation
curl $rustup_url -sSf | sh -s -- -y --no-modify-path
source "$HOME/.cargo/env"
(cd ../../ && cargo run --bin pipeline_generator)

# Dataset setup
echo "db.dropDatabase()" | $MONGOSH tpch
if [[ ! -f "$dataset_filename" ]]; then
  download $dataset_url
fi
$MONGORESTORE --drop --numInsertionWorkersPerCollection=8 --bypassDocumentValidation --gzip --archive=${dataset_filename}
