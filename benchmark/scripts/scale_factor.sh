#!/bin/bash

. ./prepare-env.sh

output=$(echo "db.partsupp.count()" | $MONGOSH tpch)
count=$(echo "$output" |  tail -2 | head -1)

if [[ $count -eq 800 ]]; then
    scale_factor=0.001
elif [[ $count -eq 8000 ]]; then
    scale_factor=0.01
elif [[ $count -eq 80800 ]]; then
    scale_factor=0.1
elif [[ $count -eq 800000 ]]; then
    scale_factor=1
else
    echo "Unexpected count value."
    exit 1
fi

echo "Scale factor $scale_factor"
