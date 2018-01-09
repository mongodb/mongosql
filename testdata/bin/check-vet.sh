#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

function vet_with_ignores() {
    cd "$PROJECT_DIR/$1"
    pkg=${PWD##*/}

    ignore="$2"

    echo "vetting package $pkg..."

    lines=$(go tool vet *.go 2>&1 | grep -vE "$ignore" | wc -l | xargs)

    if [ "$lines" != "0" ]; then
        echo "found $lines vet errs:"
        go tool vet *.go 2>&1 | grep -vE "$ignore" || true
        exit 1
    fi

    echo "done vetting package $pkg"
}

(
    set -o errexit

    for pkg in $(find . -name '*.go' | grep -v './vendor' | xargs -L1 dirname | uniq); do
        ignores=()
        case $pkg in
            # don't vet any of these packages
            '.') continue ;;

            # vet these packages, but ignore some errors
            './evaluator')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields'
                          'expr_translators.go.*bson.RegEx composite literal uses unkeyed fields' ) ;;

            './internal/dsync')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields' ) ;;

            './internal/json')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields' ) ;;

            './internal/sample')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields' ) ;;

            './internal/testutils/bench')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields' ) ;;

            './internal/testutils/dbutils')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields' ) ;;

            './internal/util')
                ignores=( 'bson.* composite literal uses unkeyed fields' ) ;;

            './internal/util/bsonutil')
                ignores=( 'bson.* composite literal uses unkeyed fields'
                          'json.* composite literal uses unkeyed fields' ) ;;

            './mongodb')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields' ) ;;

            './mongodrdl')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields'
                          'bson.Binary composite literal uses unkeyed fields' ) ;;

            './mongodrdl/mongo')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields' ) ;;

            './mongodrdl/relational')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields'
                          'bson.Binary composite literal uses unkeyed fields' ) ;;

            './parser')
                ignores=( "parsed_query_test.go:127: unrecognized printf verb 'a'" ) ;;

            './schema')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields' ) ;;

            './schema/mongo')
                ignores=( 'bson.DocElem composite literal uses unkeyed fields' ) ;;

        esac

        if [ "${#ignores[@]}" = 0 ]; then
            ignore='$^'
        else
            ignore="$( printf "|%s" "${ignores[@]}" )"
            ignore=${ignore#"|"}
        fi

        vet_with_ignores "$pkg" "$ignore"
    done

) > $LOG_FILE 2>&1

print_exit_msg
