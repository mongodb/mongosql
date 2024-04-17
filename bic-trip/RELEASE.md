# Releasing the BIC Transition Readiness Informational Profiler  
This document describes the version policy and release process for the BIC Transition Readiness Informational Profiler
(BIC TRIP), which is managed under the JIRA 'SQL' project.

## Versioning

The BIC TRIP uses [Semantic Versioning](https://semver.org/).

The BIC TRIP uses the following guidelines to determine when each version component will be updated:
- **major**: backwards-breaking changes to the library API
- **minor**: new features, including new server version support or new MongoSQL language constructs
- **patch**: bug fixes

At the moment, there are no pre-release (alpha, beta, rc, etc.) versions of the BIC TRIP.

## Releasing
This section describes the steps for releasing a new version of the BIC TRIP.

### Pre-Release Tasks
Complete these tasks before tagging a new release.

#### Start Release Ticket
Move the JIRA ticket for the release to the "In Progress" state.
Ensure that its fixVersion matches the version being released.

#### Complete the Release in JIRA
Go to the [SQL releases page](https://jira.mongodb.org/projects/SQL?selectedItem=com.atlassian.jira.jira-projects-plugin%3Arelease-page&status=unreleased), 
and ensure that all the tickets in the fixVersion to be released are closed.
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
git tag -a -m trr<version> trr<version>
example:
git tag -a -m trr1.2.3 trr1.2.3

git push --tags
```
This should trigger an Evergreen run that can be viewed on the [mongosql-rs waterfall](https://spruce.mongodb.com/waterfall/mongosql-rs).
The description for the tag triggered release starts with "Triggered From Git Tag 'trrX.Y.Z"
If it does not, you may have to ask the project manager to give you the right permissions to do so.
Make sure to run the `trr-release` task, if it is not run automatically.

#### Set Evergreen Priorities
Some evergreen variants may have a long schedule queue.
To speed up release tasks, you can set the task priority for any variant to 101 for release candidates and 200 for actual releases.
If you do not have permissions to set priority above 100, ask someone with permissions to set the priority.

### Post-Release Tasks
Complete these tasks after the release builds have completed on evergreen.

#### Verify Release Downloads
Make sure the executables are available at the proper urls:
- Mac ARM:  
  `https://translators-connectors-releases.s3.amazonaws.com/transition-readiness-report/AtlasSQLReadinessReport-macos-arm`  
  `https://translators-connectors-releases.s3.amazonaws.com/transition-readiness-report/AtlasSQLReadinessReport-macos-v${release-version}-arm`
- Mac Intel:  
  `https://translators-connectors-releases.s3.amazonaws.com/transition-readiness-report/AtlasSQLReadinessReport-macos`  
  `https://translators-connectors-releases.s3.amazonaws.com/transition-readiness-report/AtlasSQLReadinessReport-macos-v${release-version}`
- Ubuntu:  
  `https://translators-connectors-releases.s3.amazonaws.com/transition-readiness-report/AtlasSQLReadinessReport-linux`  
  `https://translators-connectors-releases.s3.amazonaws.com/transition-readiness-report/AtlasSQLReadinessReport-v${release-version}-linux`
- Ubuntu Signature:  
  `https://translators-connectors-releases.s3.amazonaws.com/transition-readiness-report/AtlasSQLReadinessReport-linux.sig`  
  `https://translators-connectors-releases.s3.amazonaws.com/transition-readiness-report/AtlasSQLReadinessReport-v-${release-version}-linux.sig`
- Windows:  
  `https://translators-connectors-releases.s3.amazonaws.com/transition-readiness-report/AtlasSQLReadinessReport-win.exe`  
  `https://translators-connectors-releases.s3.amazonaws.com/transition-readiness-report/AtlasSQLReadinessReport-v${release_version}-win.exe`  

#### Verify version in non-versioned files
Download and run one of the non-versioned executable files and verify that the version number shown on the generated HTML is correct.

#### Close Release Ticket
Move the JIRA ticket tracking this release to the "Closed" state.

#### Ensure next release ticket and fixVersion created
Ensure that a JIRA ticket tracking the next release has been created
and is assigned the appropriate fixVersion.

