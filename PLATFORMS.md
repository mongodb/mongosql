# Platforms
This document describes steps that should be taken when adding,
removing, or modifying platform definitions in `.evg.yml`.

## Removing Platforms
When removing a platform, invert the steps described in the "Adding
New Platforms" section below.

## Modifying Platforms
When making changes to a platform in evergreen, carefully check all
the steps in the "Adding New Platforms" section below. Some or all of
those locations may need to be updated.

Changes that might require updates outside of `.evg.yml` include (but
are not limited to) the following:

- Changing the name of a variant
- Changing the name of a buildvariant matrix or axis
- Moving a variant to a different matrix

## Adding New Platforms
This section describes the process for adding support for a new platform.

### Add evergreen variant
Using a similar naming scheme to the existing variants, add an
evergreen variant for the new platform.

Make sure to check the [evergreen distros
page](https://evergreen.mongodb.com/distros) to confirm the distro
name, since distro naming conventions are not consistent and generally
can't be relied upon.

Generally, new variants should be added to the `os_single_variant`
matrix, which is used for most of our platforms and runs a smaller
subset of our correctness tests.

### Update `platforms.sh`
In [platforms.sh](./testdata/bin/platforms.sh), do the following:

- Add a switch case for the newly added variant id
- Set the `PUSH_NAME` and `PUSH_ARCH` according to the existing naming schemes
- If the platform is debian or ubuntu, set `BUILD_FIPS='false'`
- If the architechture is `arm64`, set `MONGO_ARCH='aarch64'`

In most cases, this should be all that's needed, but occasionally some
additional variables may need to be added. These cases won't be
discussed/documented here, since they are the exception, not the rule.

### Update `release.py`
In [release.py](./release/scripts/release.py), do the following:

- Increment the `NUM_RELEASE_PLATFORMS` constant

### Update release json template files
Add an entry for the new platform in the following template files:

- [mongodb-bi-downloads.json](./release/scripts/mongodb-bi-downloads.json)
- [mongodb-bi-releases.json](./release/scripts/mongodb-bi-releases.json)

### Mark ticket as "Docs Changes Needed"
Mark the ticket that adds the new platform as "Docs Changes Needed",
so that our documentation can be updated to reflect the newly added
platform.
