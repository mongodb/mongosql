#!/bin/bash

. ./prepare-env.sh

Query=$1

directories=(
  "../atlas_sql/normalized/phase/"
  "../atlas_sql/denormalized/phase/"
)

for dir in "${directories[@]}"; do
  if [[ -f "${dir}${Query}" ]]; then
    collection=$(cat "${dir}${Query}" | grep "aggregate": | cut -d: -f2 | tr -d '[:space:]')
    pipeline=$(sed -n '/pipeline:/,/cursor:/{/pipeline:/!{/cursor:/!p}}' "${dir}${Query}")

    echo "Running $Query from $dir"
    echo "collection: $collection"
    echo "$pipeline"

    start_time=$(date +%s%3N)
    
    $MONGOSH --eval "db.$collection.aggregate($pipeline)" -- tpch
    
    end_time=$(date +%s%3N)

    elapsed_time=$(echo "scale=3; ($end_time - $start_time) / 1000" | bc)

    echo "Query $Query took $elapsed_time seconds"
    break
  fi
done
