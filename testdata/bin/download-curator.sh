#!/bin/bash

download_curator ()
{
    orig_dir=$(pwd)

    echo "downloading curator..."

    mkdir -p $GOPATH/src/github.com/mongodb
    cd $GOPATH/src/github.com/mongodb

    if [ ! -f "$GOBIN/curator" ]; then
        # clone the repository if it doesn't already exist
        if [ ! -d "curator" ]; then
            echo "cloning curator..."
            git clone https://github.com/mongodb/curator.git
        fi
        # then build the tool
        echo "building curator..."
        cd curator
        git checkout 967f3b29ba7f5c16a9cad555b3b4ed80aa759b32
        GO111MODULE="" GOBIN="$GOBIN" go install cmd/curator/curator.go
    else
        echo "curator already exists"
    fi

    cd "$orig_dir"
}
