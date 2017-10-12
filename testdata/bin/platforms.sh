
# set defaults for platform variables
if [ "$VARIANT" != "" ]; then
    ARCHIVE_FORMAT='tgz'
    JAVA_HOME='/opt/java/jdk8'
    GOROOT='/opt/go1.8/go'
    GOBINDIR='/opt/go1.8/go/bin'
fi

ARCHIVE_CONTENT_TYPE='x-gzip'

# set platform-specific variables
case $VARIANT in
ubuntu1404)
    PUSH_ARCH='x86_64-ubuntu-1404'
    PUSH_NAME='linux'
    ;;
macos)
    PUSH_ARCH='x86_64'
    PUSH_NAME='osx'
    GOROOT='/usr/local/go1.8/go'
    GOBINDIR='/usr/local/go1.8/go/bin'
    ;;
windows)
    PUSH_ARCH='x86_64'
    PUSH_NAME='win32'
    LIBRARY_PATH='/cygdrive/c/sasl/'
    MINGW_PATH='/cygdrive/c/mingw-w64/x86_64-4.9.1-posix-seh-rt_v3-rev1/mingw64/bin'
    ARCHIVE_FORMAT='zip'
    ARCHIVE_CONTENT_TYPE='zip'
    GOROOT='c:\go1.8\go'
    GOBINDIR='/cygdrive/c/go1.8/go/bin'
    ;;
debian71)
    PUSH_ARCH='x86_64-debian71'
    PUSH_NAME='linux'
    ;;
debian81)
    PUSH_ARCH='x86_64-debian81'
    PUSH_NAME='linux'
    ;;
amazon)
    PUSH_ARCH='x86_64-enterprise-amzn64'
    PUSH_NAME='linux'
    ;;
rhel62)
    PUSH_ARCH='x86_64-rhel62'
    PUSH_NAME='linux'
    ;;
rhel70)
    PUSH_ARCH='x86_64-rhel70'
    PUSH_NAME='linux'
    ;;
ppc)
    LIBRARY_PATH='/opt/mongodbtoolchain/v2/bin/'
    PUSH_ARCH='ppc64le-rhel71'
    PUSH_NAME='linux'
    ;;
s390x)
    LIBRARY_PATH='/opt/mongodbtoolchain/v2/bin/'
    PUSH_ARCH='s390x-enterprise-rhel72'
    PUSH_NAME='linux'
    CC='s390x-redhat-linux-gcc'
    ;;
suse11)
    PUSH_ARCH='x86_64-suse11'
    PUSH_NAME='linux'
    ;;
suse12)
    PUSH_ARCH='x86_64-suse12'
    PUSH_NAME='linux'
    ;;
other) # on evergreen, but "variant" expansion not set
    ;;
esac
