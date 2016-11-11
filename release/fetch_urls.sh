#!/bin/bash
if [ $# -eq 0 ]; then
	printf "No EVG version supplied\n";
	exit 1;
fi

VERBOSITY=0

function log () {
    if [[ $VERBOSITY -eq 1 ]]; then
        printf "$@\n\n"
    fi
}

VERSION="$1";
CURL_CMD="curl -s -H Auth-Username:$EVG_USER -H Api-Key:$EVG_KEY";
EVG_BASE="https://evergreen.mongodb.com/rest/v1";
log "checking builds on version '$VERSION'...";

DATA=$($CURL_CMD $EVG_BASE/versions/$VERSION);
BUILDS=$(echo $DATA | jq -r '.builds[]');
BUILDSLEN=$(tr -cd ' ' <<<$BUILDS|wc -c)

if [ ${BUILDSLEN} -eq 0 ];then
	printf "No builds found for version '$VERSION'";
	exit 0;
fi;

i=0

printf "{"

for BUILD in $BUILDS; do
	TASK=$($CURL_CMD $EVG_BASE/builds/$BUILD | jq -r '.tasks.dist');
	if [ -z "$TASK" ]
		then
			log "No dist task found for build '$BUILD', skipping...";
		else
			STATUS=$(echo $TASK | jq -r '.status');
			if [ "$STATUS" != "success" ]
				then
					if [ "$STATUS" != "null" ]
						then
							log "dist not part of '$BUILD', skipping...";
						else
							log "dist $STATUS on '$BUILD', skipping...";
					fi
				else
					DIST=$($CURL_CMD $EVG_BASE/tasks/$(echo $TASK | jq -r '.task_id'));
					URL=$(echo $DIST | jq -r '.files[0].url');
					VARIANT=$(echo $DIST | jq -r '.build_variant');

					printf "\"$VARIANT\": \"$URL\"";

					if [ ${BUILDSLEN} -eq $i ]
						then
							printf "}\n"
						else
							printf ",\n"
					fi
			fi
	fi
	i=$((i + 1))
done
