#!/bin/bash

# Put our own installation of Cargo first in the path to make sure this is the one used.
# Otherwise, on MacOs it will use the version installed with Brew which is not as recent.
export PATH="$HOME/.cargo/bin:$PATH"
export CARGO_NET_GIT_FETCH_WITH_CLI=true

export GOPATH="$HOME/go"
export GOROOT=/opt/golang/go1.19
export GOPRIVATE=github.com/10gen/*
export GOCACHE="$HOME/gocache"

mkdir -p "$GOCACHE"
if [ "Windows_NT" = "$OS" ]; then
    set -o igncr
    export GOROOT='C:\golang\go1.19'
    export GOPATH=$(cygpath -w "$GOPATH")
    export GOCACHE=$(cygpath -w "$GOCACHE")
fi
export PATH="$GOROOT/bin:$PATH"
if [[ "${triggered_by_git_tag}" != "" ]]; then
    export release_version=$(echo ${triggered_by_git_tag} | sed -E 's/^(v|trr|dcsb|msql)//')
else
    export release_version=snapshot
fi

PROJECT_DIRECTORY="$(pwd)"
DRIVERS_TOOLS="${PROJECT_DIRECTORY}/evergreen/drivers-tools"
MONGO_ORCHESTRATION_HOME="${DRIVERS_TOOLS}/.evergreen/orchestration"
MONGODB_BINARIES="${DRIVERS_TOOLS}/mongodb/bin"
LIBRARY_PATH="${PROJECT_DIRECTORY}"
ALLOW_VULNS="${AllowVulns}"
export DB_TOOLS_VERSION="100.7.2"
export MONGOSH_VERSION="1.8.0"

cat <<EOT >expansions.yml
release_version: "$release_version"
PROJECT_DIRECTORY: "${PROJECT_DIRECTORY}"
DRIVERS_TOOLS: "${DRIVERS_TOOLS}"
MONGO_ORCHESTRATION_HOME: "${MONGO_ORCHESTRATION_HOME}"
MONGODB_BINARIES: "${MONGODB_BINARIES}"
ALLOW_VULNS: "${ALLOW_VULNS}"
LIBRARY_PATH: "${LIBRARY_PATH}"
prepare_shell: |
  set -o errexit
  set -o xtrace
  export PATH="$PATH"
  export CARGO_NET_GIT_FETCH_WITH_CLI="$CARGO_NET_GIT_FETCH_WITH_CLI"
  export GOPATH="$GOPATH"
  export GOROOT="$GOROOT"
  export GOPRIVATE="$GOPRIVATE"
  export GOCACHE="$GOCACHE"
  export release_version="$release_version"
  export SBOM_DIR="sbom_tools"
  export SBOM_VULN="sbom.merge.grype.cdx.json"
  export ALLOW_VULNS="${ALLOW_VULNS}"
  export PROJECT_DIRECTORY="${PROJECT_DIRECTORY}"
  export DRIVERS_TOOLS="${DRIVERS_TOOLS}"
  export MONGO_ORCHESTRATION_HOME="${MONGO_ORCHESTRATION_HOME}"
  export MONGODB_BINARIES="${MONGODB_BINARIES}"
  export LIBRARY_PATH="${LIBRARY_PATH}"
  export DB_TOOLS_VERSION="${DB_TOOLS_VERSION}"
  export MONGOSH_VERSION="${MONGOSH_VERSION}"
  git config --global url."ssh://git@github.com/".insteadOf "https://github.com/"
EOT

cat expansions.yml
