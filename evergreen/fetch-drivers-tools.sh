#!/bin/bash

set -o errexit

if [[ -z "$DRIVERS_TOOLS" ]]; then
    echo >&2 "\$DRIVERS_TOOLS must be set"
    exit 1
fi

rm -rf $DRIVERS_TOOLS
# until global git config is updated directly in hosts, we need this to avoid trying
# to clone over ssh
git config --global --get-regexp '^url\.' | while read -r key _; do
    git config --global --unset "$key"
done
git clone https://github.com/mongodb-labs/drivers-evergreen-tools.git $DRIVERS_TOOLS
