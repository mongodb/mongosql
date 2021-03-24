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
git tag -a -m X.Y.Z X.Y.Z
git push --tags
```
This should trigger an Evergreen version that can be viewed on the [Database Tools Waterfall](https://evergreen.mongodb.com/waterfall/mongosql-rs).
If it does not, you may have to ask the project manager to give you the right permissions to do so.
Note that we do not use a v-prefixed tag like many projects. Make sure to run the 'release' task, if it is
not run automatically.

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
_Released YYYY-MM-DD_

We are pleased to announce version X.Y.Z of the MongoDB MongoSQL Compiler.

<INSERT-DESCRIPTION>

The MongoSQL Compiler libraries are available at the follwing URLS:
* POSIX:
  https://translators-connectors-releases.s3.us-east-1.amazonaws.com/mongosql-rs/{rhel76,ubuntu1804,macos}/X.Y.Z/libmongosql.a
* Windows:
  https://translators-connectors-releases.s3.us-east-1.amazonaws.com/mongosql-rs/{windows-64}/X.Y.Z/mongosql.dll


Bugs and feature requests can be reported in the [SQL Translator Project Jira](https://jira.mongodb.org/browse/SQL) where a list of current issues can be found.

<INSERT-LIST-OF-TICKETS>
```

- Update the release date to the date the release-json task finished on Evergreen in Eastern Time.
  You can set your timezone in "User Settings".

- Go to [Configure Release Notes](https://jira.mongodb.org/secure/ConfigureReleaseNote.jspa?projectId=18280) on JIRA.
  Choose the version you are releasing and HTML as the style.
  This will show you the list of tickets tagged with the release version.
  (If the link does not work, you can access this through the release page for the version you are releasing.)
- Go through the list of tickets and check that each ticket is categorized correctly (as a task, bugfix etc.).
  Also make sure there is nothing in the list that might have been tagged with the wrong fix version.
- Copy the HTML list of tickets from JIRA and paste it in CHANGELOG.md in place of `<INSERT-LIST-OF-TICKETS>`.
- Remove the top line of the list of tickets that says `Release Notes - MongoDB MongoSQL Compiler - Version X.Y.Z`
- Change the ticket type titles from `<h2>`s to `<h3>`s. For example,

    ```
    <h2>        Build Failure
    </h2>
    ```

    Becomes:

    ```
    ### Build Failure
    ```
- Insert a brief description of the release in place of `<INSERT-DESCRIPTION>`.
  Do not go into too much unnecessary detail.
- Submit a PR with your changes under the release ticket number, and merge once approved.

#### Close Release Ticket
Move the JIRA ticket tracking this release to the "Closed" state.

#### Ensure Downstream Tickets Created
Manually create Downstream integration ticket for Atlas Datalake
