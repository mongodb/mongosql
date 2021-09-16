#!/bin/bash

LATEST="5.0-stable"

get_latest_for_distro() {
   distro=$1
   if [ "$distro" = "debian81" ]; then
        version_for_curator="4.1.10" # Support for Debian 8 was removed in SERVER-37767, r4.1.11
   elif [ "$distro" = "ubuntu1404" ]; then
        version_for_curator="4.1.9" # Support for Debian 8 was removed in SERVER-37765, r4.1.10
   elif [ "$distro" = "ubuntu1604" ] && [ "$arch" = "ppc64le" ]; then
        version_for_curator="4.1.9" # Support for Enterprise Ubuntu 16.04 PPCLE was removed in SERVER-37774, r4.1.10
   elif [ "$distro" = "ubuntu1604" ] && [ "$arch" = "s390x" ]; then
        version_for_curator="3.6.4" # Support for Enterprise Ubuntu 16.04 s390x was removed in r3.6.5
   elif [ "$distro" = "rhel67" ] ; then
         version_for_curator="4.2.0" # Support for RHEL 6.7 was removed in r4.2.1
   elif [ "$distro" = "osx" ] ; then
         # macos currently is failing to load data for blackbox and
         #tableau, see BI-2662
         version_for_curator="4.4-stable"
   else
         version_for_curator="$LATEST"
   fi
}

set_mongodb_binaries ()
{
   # Use lowercase variable names to make sure we do not conflict with any
   # globals, it is safer
   mongodb_version=$1

   # get the distro and arch values from the PUSH_ARCH environment variable.
   distro=${MONGO_DISTRO:-$(echo $PUSH_ARCH | cut -d'-' -f2)}
   arch=${MONGO_ARCH:-$(echo $PUSH_ARCH | cut -d'-' -f1)}

   edition="enterprise"

   echo "Distro $distro, arch $arch"

   orig_dir=$(pwd)
   cache="$SQLPROXY_TEST_CACHE_DIR/mongodb-downloads"
   local_versioned_path="$cache/cached-mongodb-$mongodb_version"

   # Functions to download curator
   . $DIR/download-curator.sh
   download_curator

   # If we are on evergreen, delete the cache
   if [ "$VARIANT" != "" ]; then
	   echo "Deleting mongodb download cache ($cache)"
	   rm -Rf "$cache"
   fi

   # Only download if we do not have a local copy of this
   # specific mongo version
   if [ ! -e $local_versioned_path ]; then
       echo "Downloading mongodb binaries"

       # If version is "latest", get the latest version available for that distro.
       version_for_curator=$mongodb_version
       if [ "$mongodb_version" = "5.0" ]; then
            version_for_curator="5.0-stable"
       fi
       if [ "$mongodb_version" = "latest" ]; then
            get_latest_for_distro $distro
       fi

       # If running on Ubuntu 18.04, use the community, ie. "targeted"
       # edition. There is no enterprise edition.
       if [ "$distro" = "ubuntu1804" ]; then
           edition="targeted"
       elif [ "$distro" = "linux_x86_64" ]; then
           edition="base"
       fi

       $GOBIN/curator artifacts download \
           --target $distro \
           --arch $arch \
           --version $version_for_curator \
           --edition $edition \
           --path $cache

       cd $cache

       # Remove the compressed file, which may be either a tgz or zip file.
       rm -f mongodb*.tgz
       rm -f mongodb*.zip

       mv mongodb* $local_versioned_path
       chmod -R +x $local_versioned_path
   else
       echo "Using cached mongodb"
   fi

   cd $ARTIFACTS_DIR
   if [ "Windows_NT" = "$OS" ]; then
       # Windows orchestration cannot handle symlinks, so we must cp.
       # -T ensures that mongodb will be overwritten instead of ending up with
       # mongodb under mongodb
       cp -RT $local_versioned_path mongodb
   else
       # On *nix, a symlink to the directory will work
       ln -s $local_versioned_path mongodb || true
   fi
   mongodb/bin/mongod --version
   cd "$orig_dir"
}
