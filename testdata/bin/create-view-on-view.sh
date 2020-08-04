#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

	echo "creating view: db.createView($VIEW1, $SOURCE1, [{$project: {a: 1, b: 1}}]) $DATABASE"
	"$ARTIFACTS_DIR/mongodb/bin/mongo" $MONGO_CLIENT_ARGS --eval "db.createView('$VIEW1', '$SOURCE1', [{\$project: {'a': 1, 'b': 1}}])" "$DATABASE"

	echo "creating view: db.createView($VIEW2, $SOURCE2, [{$project: {a: 1}}]) $DATABASE"
	"$ARTIFACTS_DIR/mongodb/bin/mongo" $MONGO_CLIENT_ARGS --eval "db.createView('$VIEW2', '$SOURCE2', [{\$project: {'a': 1}}])" "$DATABASE"

	echo "done creating view"

) > $LOG_FILE 2>&1

print_exit_msg

