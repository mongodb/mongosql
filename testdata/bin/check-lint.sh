#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

function lint_with_ignores() {
    cd "$PROJECT_DIR/$1"
    pkg=${PWD##*/}

    ignore="$2"

    echo "linting package $pkg..."

    lines=$($GOPATH/bin/golint . | grep -vE "$ignore" | wc -l | xargs)

    if [ "$lines" != "0" ]; then
        echo "found $lines lint errs:"
        $GOPATH/bin/golint . | grep -vE "$ignore"
        exit 1
    fi

    echo "done linting package $pkg"
}

(
    set -o errexit

    # install golint if missing
    which golint > /dev/null 2>&1 || go get -u github.com/golang/lint/golint

    for pkg in $(find . -name '*.go' | grep -v './vendor' | xargs dirname | uniq); do
        ignores=()
        case $pkg in
            # don't lint any of these packages
            './mongodrdl') continue ;;
            './mongodrdl/mongo') continue ;;
            './mongodrdl/relational') continue ;;
            './parser') continue ;;
            './parser/sqltypes') continue ;;

            # lint these packages, but ignore some errors
            './evaluator')
                ignores=( 'should have comment'
                          'ALL_CAPS'
                          'returns unexported type'
                          'algebrizer.go.*if block ends with a return statement' ) ;;

            './internal/json')
                ignores=( 'should have comment'
                          'ALL_CAPS'
                          'InvalidUTF8Error' ) ;;

            './internal/sample')
                ignores=( 'should have comment'
                          'stutters'
                          'returns unexported type' ) ;;

            './internal/testutils')
                ignores=( 'should have comment' ) ;;

            './internal/testutils/dbutils')
                ignores=( 'should have comment' ) ;;

            './internal/util')
                ignores=( 'should have comment' ) ;;

            './internal/util/bsonutil')
                ignores=( 'should have comment'
                          'underscores in Go names' ) ;;

            './log')
                ignores=( 'should have comment' ) ;;

            './mysqlerrors')
                ignores=( 'should have comment'
                          'ALL_CAPS' ) ;;

            './options')
                ignores=( 'should have comment' ) ;;

            './password')
                ignores=( 'should have comment' ) ;;

            './schema/mongo')
                ignores=( 'should have comment' ) ;;

            './server')
                ignores=( 'should have comment'
                          'ALL_CAPS'
                          'returns unexported type'
                          'underscores in Go names' ) ;;
        esac

        if [ "${#ignores[@]}" = 0 ]; then
            ignore='$^'
        else
            ignore="$( printf "|%s" "${ignores[@]}" )"
            ignore=${ignore#"|"}
        fi

        lint_with_ignores "$pkg" "$ignore"
    done

) > $LOG_FILE 2>&1

print_exit_msg
