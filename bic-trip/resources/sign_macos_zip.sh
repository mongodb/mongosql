#!/bin/bash
# Uses macnotary to sign and notarize the mac executable `AtlasSQLReadinessReport-macos`.
# The key id and secret were set as MACOS_NOTARY_KEY and MACOS_NOTARY_SECRET
# env vars from the expansions. The macnotary client will look for these env
# vars so we don't need to pass the credentials as CLI options.

# Turn on verbose debugging for remainder of script
set -o xtrace
set -o errexit
set -o verbose

# Turn the executable into a zip
zip -r AtlasSQLReadinessReport.zip AtlasSQLReadinessReport-macos
rm AtlasSQLReadinessReport-macos

curl -LO https://macos-notary-1628249594.s3.amazonaws.com/releases/client/v3.3.3/darwin_amd64.zip
unzip darwin_amd64.zip
chmod 0755 ./darwin_amd64/macnotary

./darwin_amd64/macnotary -v

./darwin_amd64/macnotary \
    --task-comment "Signing the Atlas SQL Transition Readiness Report" \
    --file "$PWD/AtlasSQLReadinessReport.zip" \
    --mode notarizeAndSign \
    --url https://dev.macos-notary.build.10gen.cc/api \
    --bundleId com.mongodb.mongosql \
    --out-path "$PWD/AtlasSQLReadinessReport-signed.zip"

unzip AtlasSQLReadinessReport-signed.zip
if [ ! -f "AtlasSQLReadinessReport-macos" ]; then
    echo "Error: AtlasSQLReadinessReport-macos does not exist."
    exit 1
fi
