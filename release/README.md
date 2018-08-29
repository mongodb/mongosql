# Requirements

- Evergreen username and API access key (available at https://evergreen.mongodb.com/settings)
- AWS API access key and secret (you need read/write permissions on the `info-mongodb-com` S3 bucket)

# Process

Run `python scripts/release.py --version <EVERGREEN_VERSION_ID>` and follow the prompt to release a new version of the BI Connector.

Verify the information on the downloads site at ```https://www.mongodb.com/download-center#bi-connector``` is correct and that all URLs are working correctly.
