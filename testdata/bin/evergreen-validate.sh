#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
      set -o errexit

      OUTPUT=$(~/evergreen validate -p sqlproxy -f .evg.yml)

      echo $OUTPUT

      if [[ $? != "0" || $(grep 'is valid with warnings' <<< $OUTPUT) ]]; then
        exit 1
      fi

) > $LOG_FILE 2>&1

print_exit_msg
