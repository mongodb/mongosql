#!/bin/bash

# aws cli must be installed and configured
# aws configure
# AWS Access Key ID [None]: <release_aws_key>
# AWS Secret Access Key [None]: <release_aws_secret>

# This script downloads all s3 access logs for `translators-connectors-releases` bucket.
# Shows the daily and total download counts.

COUNTER=0
TEMPFILE=/tmp/dlcount.tmp

aws s3 sync s3://translators-connectors-releases/access-logs/s3/ . --quiet
EXIT_STATUS=$?
if [ $EXIT_STATUS -ne 0 ]; then
    echo "Failed to sync S3 bucket. Exited with error code $EXIT_STATUS."
    exit 1
fi

find . -type f | xargs grep -E \
            "AtlasSQLReadinessReport-v[0-9]+\.[0-9]+\.[0-9]+-(macos|macos-arm|linux|win.exe)" \
                | grep GET > $TEMPFILE

counts() {
    COUNT=$(grep -c "$1" $TEMPFILE)
    if [ "$COUNT" -ne "0" ]; then
        printf "%s %3s " "$1" "$COUNT"
        printf 'X%.0s' $(seq "$COUNT")
        echo
    fi
    (( "COUNTER=$COUNTER+$COUNT" ))
}

# Counts the last 500 days
for i in $(seq 500 -1 0)
do
    DATE=$(date -d "-${i} days" +%Y-%m-%d)
    counts "$DATE"
done

# Logs are in UTC time and can be a day ahead
DATE=$(date -d "+1 day" +%Y-%m-%d)
counts "$DATE"

echo ---------------------------------
echo $COUNTER total downloads

rm $TEMPFILE
