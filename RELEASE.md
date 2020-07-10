# Releasing the BI Connector

This document describes the release process for the BI Connector.

## Pre-Release Tasks

### Start Release Ticket
Move the JIRA ticket for the release to the "In Progress" state.
Ensure that its fixVersion matches the version being released.

### Ensure Evergreen Passing
Ensure that the build you are releasing is passing the tests on the evergreen waterfall.
A completely green build is not mandatory, since we do have flaky tests; however, failing tests should be manually investigated to ensure they are not actual failures.

### Complete the Release in JIRA
Go to the [BI Connector releases page](https://jira.mongodb.org/projects/BI?selectedItem=com.atlassian.jira.jira-projects-plugin%3Arelease-page&status=unreleased), and ensure that all the tickets in the fixVersion to be released are closed.
The only uncompleted ticket in the release should be the release ticket.
If there are any remaining tickets that will not be included in this release, remove the fixVersion and assign them a new one if appropriate.
Close the release on JIRA, adding the current date.

### File Downstream Tickets
File downstream tickets for the release, all of which should depend on the release JIRA ticket.

Two CLOUDP tickets:
- "Release BI Connector 2.14.X to CM/OM" with a component of "BI Connector" and assigned team of "Automation"
- "Release BI Connector 2.14.X to Atlas" with a component of "BI Connector" and assigned team of "Atlas Triage"

And one DOCSP ticket:
- "BI Connector 2.14.X Release Notes", with a component of "BI Connector" and a link to the release notes.

### Send Slack Notifications
Inform stakeholders that a release is about to occur by sending slack messages in the following channels:
- `#enterprise-tools`
- `#connectors-mgmt` (include links to the CLOUDP tickets)
- `#docs-bic` (include a link to the DOCSP ticket)

## Releasing

### Major/Minor/Beta/RC Releases
The current assumption is that 2.14 will be the last non-patch release of the BI Connector.
As such, the release process as written here does not account for Major, Minor, Beta, or RC releases.
Some parts of the release process/infrastructure will have to be changed in order to support those types of releases.

### Patch Releases

#### Create the bump commit
Ensure you have the `master` branch checked out.
Create an empty commit with a commit message of the following format, where `<patch>` is replaced with the patch version being released.
```
git commit --allow-empty -m 'BUMP v2.14.<patch>'
```

#### Create the tag and push
Create an annotated tag and push it:
```
git tag -a -m 2.14.<patch> v2.14.<patch>
git push && git push --tags
```

## Post-Release Tasks

### Wait for Builds
Wait for the release builds to complete on evergreen.
You may need to bump the priority of some tasks to get them to run in a timely manner.

### Verify Release Downloads
Go to the [Download Center](https://www.mongodb.com/try/download/bi-connector) and verify that the new release is available there.
Download the package for your OS and confirm that `mongosqld --version` prints the correct version.

### Send Release Announcement Email
Send a release announcement email to the [biconnector-announcements](https://groups.google.com/a/10gen.com/g/biconnector-announcements) group, following the style of previous announcement messages.

### Close Release Ticket
Move the JIRA ticket tracking this release to the "Closed" state.
