# MongoDB Versions
This document describes steps that should be taken when adding or
removing MongoDB Version support to/from the BI Connector and its
evergreen project.

## Adding a new MongoDB version
- Add the version to the `mongodb_version` axis definition
- Update the `LATEST` variable at the top of `download-mongodb.sh`

## Removing a MongoDB version
- Remove the version from the `mongodb_version` axis definition
- Remove any platforms that no longer need to be supported as of this version removal
  - Follow the process documented in [`platforms.md`](./platforms.md)
- Remove any explicit handling/special casing for the removed version throught the repo.
