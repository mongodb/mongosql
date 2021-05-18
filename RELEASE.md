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

#### Ensure master up to date
Ensure you have the `master` branch checked out, and that you have pulled the latest commit from `10gen/mongosql-rs`.

#### Create the tag and push
Create an annotated tag and push it:
```
git tag -a -m X.Y.Z vX.Y.Z
git push --tags
```
This should trigger an Evergreen version that can be viewed on the [mongosql-rs waterfall](https://evergreen.mongodb.com/waterfall/mongosql-rs).
If it does not, you may have to ask the project manager to give you the right permissions to do so.
Make sure to run the 'release' task, if it is not run automatically.

#### Set Evergreen Priorities
Some evergreen variants may have a long schedule queue.
To speed up release tasks, you can set the task priority for any variant to 101 for release candidates and 200 for actual releases.
If you do not have permissions to set priority above 100, ask someone the project manager to set the
priority.

### Post-Release Tasks
Complete these tasks after the release builds have completed on evergreen.

#### Verify Release Downloads
Make sure the libraries are available at the proper urls:
- POSIX:
  https://translators-connectors-releases.s3.us-east-1.amazonaws.com/mongosql-rs/{rhel76,ubuntu1804,macos}/<SemVer>/libmongosql.a
- Windows:
  https://translators-connectors-releases.s3.us-east-1.amazonaws.com/mongosql-rs/{windows-64}/<SemVer>/mongosql.dll

#### Announce Release on Slack

The following template will be used for the Slack Release Announcement, which will be announced in the #enterprise-tools channel:

```
We are pleased to announce version X.Y.Z of the MongoDB MongoSQL Compiler.

<INSERT-DESCRIPTION>

The MongoSQL Compiler libraries are available at the follwing URLS:
* POSIX:
  https://translators-connectors-releases.s3.us-east-1.amazonaws.com/mongosql-rs/{rhel76,ubuntu1804,macos}/X.Y.Z/libmongosql.a
* Windows:
  https://translators-connectors-releases.s3.us-east-1.amazonaws.com/mongosql-rs/{windows-64}/X.Y.Z/mongosql.dll
```

#### Close Release Ticket
Move the JIRA ticket tracking this release to the "Closed" state.

#### Ensure next release ticket and fixVersion created
Ensure that a JIRA ticket tracking the next release has been created
and is assigned the appropriate fixVersion.

#### Ensure Downstream Tickets Created
Manually create MHOUSE ticket for integrating the new release into ADL
