# Release Instructions

This document contains instructions for releasing various components of MongoSQL. At this time, those components are:
1. `libmongosqltranslate` - used by the ODBC and JDBC drivers for on-prem and direct cluster querying.
2. `libmongosql` - used by ADF.
3. `mongodb-schema-manager` - used for on-prem and direct cluster schema management.
4. `transition-readiness-report` - used to help inform customers what work may be required when moving from BIC to MongoSQL.

## Versioning

All projects/crates in this repository use [Semantic Versioning](https://semver.org/).

### Versioning `libmongosqltranslate`

`libmongosqltranslate` uses the following guidelines to determine when each version component will be updated:
- **major**: backwards-breaking changes to the library API
- **minor**: new features, including new server version support or new MongoSQL language constructs
- **patch**: bug fixes
- **pre-release**: which pre-release version, for use during the on-prem EAP

Tags are prepended with `libv` for `libmongosqltranslate` releases.

```sh
# git tag -am libv1.0.0-<prerelease>-<prerelease-version> libv1.0.0-<prerelease>-<prerelease-version>
git tag -am libv1.0.0-alpha-1 libv1.0.0-alpha-1
```

### Versioning `libmongosql`

`libmongosql` uses the following guidelines to determine when each version component will be updated:
- **major**: backwards-breaking changes to the library API
- **minor**: new features, including new server version support or new MongoSQL language constructs
- **patch**: bug fixes

At the moment, there are no pre-release (alpha, beta, rc, etc.) versions of `libmongosql`.

Tags are prepended with `v` for `libmongosql` releases.

```sh
git tag -am v1.6.1 v1.6.1
```

### Versioning `mongodb-schema-manager`

`mongodb-schema-manager` uses the following guidelines to determine when each version component will be updated:
- **major**: backwards-breaking changes to the library API
- **minor**: new features, including new server version support or new MongoSQL language constructs
- **patch**: bug fixes
- **pre-release**: which pre-release version, for use during the on-prem EAP

Tags are prepended with `msm` for `mongodb-schema-manager` releases.

```sh
git tag -am msm1.0.0-alpha-1 msm1.0.0-alpha-1
```

### Versioning `transition-readiness-report`

`transition-readiness-report` uses the following guidelines to determine when each version component will be updated:
- **major**: backwards-breaking changes to the library API
- **minor**: new features, including new server version support or new MongoSQL language constructs
- **patch**: bug fixes

Tags are prepended with `trr` for `transition-readiness-report` releases.

```sh
git tag -am trr1.0.0 trr.0.0
```

## Releasing

This section describes the steps for releasing new versions.

### Pre-Release Tasks

Complete these tasks before tagging a new release.

### Start Release Ticket

Move the JIRA ticket for the release to the "In Progress" state.
Ensure that its fixVersion matches the version being released.

### Complete the Release in JIRA

Go to the [SQL releases page](https://jira.mongodb.org/projects/SQL?selectedItem=com.atlassian.jira.jira-projects-plugin%3Arelease-page&status=unreleased), and ensure that all the tickets in the fixVersion to be released are closed.
Ensure that all the tickets have the correct type. Take this opportunity to edit ticket titles if they can be made more descriptive.
The ticket titles will be published in the changelog.

If you are releasing a patch version but a ticket needs a minor bump, update the fixVersion to be a minor version bump.
If you are releasing a patch or minor version but a ticket needs a major bump, stop the release process immediately.

The only uncompleted ticket in the release should be the release ticket.
If there are any remaining tickets that will not be included in this release, remove the fixVersion and assign them a new one if appropriate.

Close the release on JIRA, adding the current date (you may need to ask the SQL project manager to do this).

#### Ensure Evergreen Passing

Ensure that the build you are releasing is passing the tests on the [mongosql-rs waterfall](https://spruce.mongodb.com/commits/mongosql-rs).

#### Ensure master up to date

Ensure you have the `master` branch checked out, and that you have pulled the latest commit from `10gen/mongosql-rs`.

#### Create the tag and push

Create an annotated tag and push it:

```sh
git tag -a -m <major>.<minor>.<patch> <project-prefix><major>.<minor>.<patch>
git push --tags
```

This should trigger an Evergreen run that can be viewed on the [mongosql-rs waterfall](https://spruce.mongodb.com/waterfall/mongosql-rs).
The description for the tag triggered release starts with "Triggered From Git Tag 'vX.Y.Z"
If it does not, you may have to ask the project manager to give you the right permissions to do so.
Make sure to run the 'release' task, if it is not run automatically.

#### Set Evergreen Priorities

Some evergreen variants may have a long schedule queue.
To speed up release tasks, you can set the task priority for any variant to 101 for release candidates and 200 for actual releases.
If you do not have permissions to set priority above 100, ask someone with permissions to set the
priority.

### Post-Release Tasks

Complete these tasks after the release builds have completed on evergreen.

#### Verify Release Downloads

Make sure the libraries are available at the proper urls. They should be available in the `translators-connectors-releases` bucket,
and you can locate them in the evergreen build in the publish or upload tasks.

#### Ensure next release ticket and fixVersion created

Create a new release for the Atlas SQL Project:
- Ask the TPM or Lead to create the next fix version, if it does not yet exist.
- Select the MongoSQL Project
- Clone the previous Release Ticket.
- Go to the Cloned ticket and set the Fix Version to the next appropriate Version

#### Update and Close Release Ticket

Find the JIRA ticket tracking this release.
Edit the description to include a link to the correct JIRA release notes, and move it to the "Closed" state.
Any tickets still marked as Accepted with this release version must be updated to the next release
version.

### ADF downstream Ticket and Library Update for `libmongosql`

The following sections currently only apply to `libmongosql` releases.

#### ADF downstream Ticket

When closing the release task, a [MHOUSE](https://jira.mongodb.org/projects/MHOUSE) ticket for integrating the new release into ADF should be created automatically.
If not, create it manually and link it as "Depends On" the release ticket. The easiest way to do this is to clone the previous downstream MHOUSE ticket and update the version numbers as necessary.
Once the downstream MHOUSE ticket is created, assign it to yourself.
Follow the steps below to upgrade the version of mongosql deployed in Atlas and close the ticket once changes have been merged.

#### Library Update

Follow the instructions in [mongohouse/internal/sql/mongosql/versions.go](https://github.com/10gen/mongohouse/blob/master/internal/sql/mongosql/versions.go) to upgrade the version of mongosql.
After updating the version of mongosql via the `go get` command, your go.mod and go.sum files should have changed. Also run `go mod tidy` if needed.

This is an example of a PR for upgrading Mongosql version from 1.2.1 to 1.2.2 : https://github.com/10gen/mongohouse/commit/96bd9666fe39c134bad1048054a40ace380a0b23
Note: MHOUSE requires all commits to be signed.

Once you have made the changes locally, create a PR to merge the changes.
Follow the [ADFA guidelines](https://wiki.corp.mongodb.com/display/MMS/ADFA+Code+Review+Guidelines#ADFACodeReviewGuidelines-MergingaPullRequest) related to the process for handling PRs.
Request the ADF Query lead as a reviewer on your PR.

If you encounter issues with your PR consistently failing required tests, please check [our Wiki](https://wiki.corp.mongodb.com/display/DRIVERS/SQL+Engines+ADF+Mongosql+upgrade) for the procedure to follow.

#### ADF release Timeline

The ADF team does weekly releases on Tuesdays. For a release to be integrated into ADF on their next
Tuesday release, the library update must be merged by the previous Friday. They will then cut the
release branch on Friday and release the following Tuesday.

If there are critical fixes that we need to have integrated in the next Tuesday release, we need to
let them know either by flagging it on the #atlas-sql mongosql release announcement by tagging
@adf-query or the ADF Query lead, or by pinging the ADF Query lead directly, if the lead is not around, then by
pinging the ADF Query team on #atlas-data-federation-sql-collab-eng.

If there ever is a need for a hot fix we want to get released ASAP, reach out to any ADFA lead / @mongohouse-eng-oncall.

To monitor the status of ADF releases, check the #mongohouse-releases channel.

#### Announce Release on Slack

Use the following messsage template to announce the release in the #atlas-sql channel:

> Hello! We've released mongosql-rs version \<VERSION\>
> More information, including release notes, can be found on the Release Ticket: \<JIRA Link\>
> Please note that the release roll-out in Atlas is a separate process and it can take up to 2 weeks before the next release is available.
