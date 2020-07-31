#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

echo 'ensuring mongosqld is available...'

set +e

code=1
for i in {0..20}; do
    echo 'calling mysql...'
    output=$(mysql $CLIENT_ARGS -e "use information_schema;" 2>&1)
    code=$?
    if [ "$code" = 0 ]; then
        echo 'successfully connected to mongosqld'
        exit 0
    fi
    echo "...could not connect to mongosqld"
    sleep 10
done
exit 1
