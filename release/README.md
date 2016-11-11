# Requirements

- jq - https://stedolan.github.io/jq/
- Evergreen username and key (available at https://evergreen.mongodb.com/settings)
- AWS CLI - http://docs.aws.amazon.com/cli/latest/userguide/installing.html
- AWS API access key and secret (you need permissions on the `info-mongodb-com` S3 bucket)

# Process

Follow the steps below to update the downloads site:

1. Find the release candidate's version identifier from Evergreen at https://evergreen.mongodb.com/waterfall/sqlproxy; let's call this `$VERSION`.

2. Run export `EVG_KEY=XXX && export EVG_USER=XXX` to export the environment variables required to pull the candidate archive URLs from Evergreen.

3. Supply the release candidate's version identifier to the fetch_urls.sh script to fetch the URLs from Evergreen, e.g. `./fetch_urls.sh $VERSION > urls.json`. This will generate the JSON data you'll need in the steps below.

4. Download the binaries from each of the URLs into a dist directory; e.g.```mkdir dist; for url in $(cat urls.json | jq -r 'to_entries[] | .value') ; do curl --url $url -o dist/$(basename $url); done;```

5. For each binary, run `aws s3 cp $BINARY s3://info-mongodb-com/mongodb-bi/v2/` to copy the release to S3 - e.g. ```for BINARY in $(ls dist); do 
  echo uploading $BINARY...; aws s3 cp dist/$BINARY s3://info-mongodb-com/mongodb-bi/v2/;
done```.

6. Update `mongodb-bi-downloads.json` to reflect the new release version and locations for each binary. e.g.```cat url.json | jq -r 'to_entries[] | "\(.key)\t\(.value)\t\t\t"' | awk '{split($0,a," ")}{print a[1]}{name=system("basename " a[2])}' |  xargs -L2 sh -c 'sed -i .bak "s/$0/$1/" mongodb-bi-downloads.json' | sed -i .bak "s/VERSION/2.0.0-rc0/" mongodb-bi-downloads.json'```.

7. Run ```aws s3 cp mongodb-bi-downloads.json s3://info-mongodb-com/mongodb-bi/``` to push the new information to the website.

8. Verify the information on the downloads site at ```https://www.mongodb.com/download-center#bi-connector``` is correct and that all URLs are working correctly.
