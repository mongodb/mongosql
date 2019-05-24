#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
	echo 'diffing against HEAD to detect uncommitted changes...'
	git diff-index --stat HEAD --
	status=$?

	# if diff-index did not exit 0, then there was something in the diff
	if [ $status -ne 0 ]; then
		has_uncommitted_changes='true'
	else
		has_uncommitted_changes='false'
	fi

	merge_base_ref="$(git merge-base HEAD master)"
	head_ref="$(git rev-parse HEAD)"

	if [ "$merge_base_ref" != "$head_ref" ]; then
		echo 'current commit not in master'
		echo 'diffing with merge-base of HEAD and master'
		diff_ref="$merge_base_ref"
	elif [ "$has_uncommitted_changes" = 'true' ]; then
		echo 'found uncommitted changes'
		echo 'diffing with previous HEAD'
		diff_ref='HEAD'
	else
		echo 'found no uncommitted changes'
		echo 'diffing with previous commit'
		diff_ref='HEAD~'
	fi

	echo "diffing against $diff_ref..."
	diff_output="$(git diff-index --stat "$diff_ref" -- CHANGELOG.md | head -n 1 | tr -d '\n')"
	echo "$diff_output"

	expected_diff_output=' CHANGELOG.md | 1 +'
	if [ "$diff_output" != "$expected_diff_output" ]; then
		echo 'incorrect changelog diff:'
		echo "       got: '$diff_output'"
		echo "  expected: '$expected_diff_output'"
		echo ''
		echo 'full diff included below:'
		git diff-index --stat "$diff_ref" --
		exit 1
	fi

	echo 'got expected diff output'

) > $LOG_FILE 2>&1

print_exit_msg
