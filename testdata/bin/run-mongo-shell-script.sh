#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

    set -o errexit

    user="$MONGO_SHELL_SCRIPT_USER"
    pass="$MONGO_SHELL_SCRIPT_PASS"
	host="$MONGO_SHELL_SCRIPT_HOST"
	script="$SCRIPT_DIR/$MONGO_SHELL_SCRIPT_NAME.js"
	uri="mongodb+srv://$user:$pass@$host"
	mongoBin="$ARTIFACTS_DIR/mongodb/bin/mongo"

	cat $script | $mongoBin --ssl $uri




