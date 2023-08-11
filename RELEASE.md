# Releasing the MongoDB MongoSQL Compiler

This document describes the version policy and release process for the MongoDB MongoSQL Compiler.
The MongoSQL Compiler is managed under the JIRA 'SQL' project.

## Versioning

The MongoDB MongoSQL Compiler uses [Semantic Versioning](https://semver.org/).

The MongoSQL Compiler uses the following guidelines to determine when each version component will be updated:
- **major**: backwards-breaking changes to the library API
- **minor**: new features, including new server version support or new MongoSQL language constructs
- **patch**: bug fixes

At the moment, there are no pre-release (alpha, beta, rc, etc.) versions of the MongoSQL compiler.

## Releasing
This section describes the steps for releasing a new version of the MongoSQL Compiler.

### Pre-Release Tasks
Complete these tasks before tagging a new release.

#### Start Release Ticket
Move the JIRA ticket for the release to the "In Progress" state.
Ensure that its fixVersion matches the version being released.

#### Complete the Release in JIRA
Go to the [SQL releases page](https://jira.mongodb.org/projects/SQL?selectedItem=com.atlassian.jira.jira-projects-plugin%3Arelease-page&status=unreleased), and ensure that all the tickets in the fixVersion to be released are closed.
Ensure that all the tickets have the correct type. Take this opportunity to edit ticket titles if they can be made more descriptive.
The ticket titles will be published in the changelog.

If you are releasing a patch version but a ticket needs a minor bump, update the fixVersion to be a minor version bump.
If you are releasing a patch or minor version but a ticket needs a major bump, stop the release process immediately.

The only uncompleted ticket in the release should be the release ticket.
If there are any remaining tickets that will not be included in this release, remove the fixVersion and assign them a new one if appropriate.

Close the release on JIRA, adding the current date (you may need to ask the SQL project manager to do this).

### Releasing

#### Ensure Evergreen Passing
Ensure that the build you are releasing is passing the tests on the [mongosql-rs waterfall](https://spruce.mongodb.com/commits/mongosql-rs).

#### Ensure master up to date
Ensure you have the `master` branch checked out, and that you have pulled the latest commit from `10gen/mongosql-rs`.

#### Create the tag and push
Create an annotated tag and push it:
```
git tag -a -m X.Y.Z vX.Y.Z
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
Make sure the libraries are available at the proper urls:
- POSIX:
  `https://translators-connectors-releases.s3.us-east-1.amazonaws.com/mongosql-rs/{rhel76,ubuntu1804,macos,macos-arm64,amazon2-arm64}/${release_version}/libmongosql.a`
- Windows:
  `https://translators-connectors-releases.s3.us-east-1.amazonaws.com/mongosql-rs/windows/${release_version}/mongosql.dll`

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

#### Ensure Downstream Tickets Created
Manually create an [MHOUSE](https://jira.mongodb.org/projects/MHOUSE) ticket for integrating the new
release into ADL, and link it as "Depends On" the release ticket. The easiest way to do this is to
clone the previous MHOUSE ticket and update the version numbers as necessary.

The ADL team does bi-weekly releases on Tuesdays. For a release to be integrated into ADL on their next
Tuesday release, the downstream ticket must be in by the previous Monday. They will then cut the
release branch on Friday and release the following Tuesday.

If there are critical fixes that we need to have integrated in the next Tuesday release, we need to
let them know either by flagging it on the #enterprise-tool mongosql release announcement by tagging
@adl-query or the ADL Query lead, or by pinging the ADL Query lead directly, if the lead is not around, then
pinging the ADL Query team on #adl-sql-collab.

#### Announce Release on Slack
Use the following messsage template to announce the release in the #enterprise-tools channel:

> Hello! We've released mongosql-rs version <VERSION>
> More information, including release notes, can be found on the Release Ticket: <JIRA Link>
