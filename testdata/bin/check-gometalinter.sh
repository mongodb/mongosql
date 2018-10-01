#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit

    cache="$SQLPROXY_TEST_CACHE_DIR/go1.11"
    go_path="$cache/go"
    go_archive="$go_path".tar.gz

    if [ ! -e "$go_path" ]; then
        mkdir -p "$cache"
        rm -rf "$go_path" "$go_archive"

        if [ "$(uname)" = 'Darwin' ]; then
            go_url='https://dl.google.com/go/go1.11beta3.darwin-amd64.tar.gz'
            extract='tar xzvf'
        elif [ "$(uname)" = 'Linux' ]; then
            go_url='https://dl.google.com/go/go1.11beta3.linux-amd64.tar.gz'
            extract='tar xzvf'
        else
            echo "unsupported os '$(uname)', cannot download go"
            exit 1
        fi

        cd "$cache"
        curl "$go_url" --output "$go_archive"
        $extract "$go_archive"
        rm "$go_archive"
    fi

    export GOROOT="$go_path"
    export PATH="$GOROOT/bin:$PATH"

    go get -u github.com/alecthomas/gometalinter

    cd "$PROJECT_DIR"
    lint=gometalinter
    $lint --install
    $lint --deadline 200s $(find . -name '*.go' | grep -v './vendor' | grep -v 'sqlproxy-test-cache' | grep -v './testdata' | xargs -L1 dirname | uniq)

) > $LOG_FILE 2>&1

print_exit_msg
