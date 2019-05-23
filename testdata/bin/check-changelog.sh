#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
	echo 'diffing against HEAD to detect uncommitted changes...'
	git diff-index --stat HEAD --
	status=$?

	# if diff-index returned 0, then there was nothing in the diff
	if [ $status -eq 0 ]; then
		echo 'found no uncommitted changes'
		if [ "$VARIANT" != '' ]; then
			merge_base_ref="$(git merge-base HEAD master)"
			echo "on evergreen, resetting to merge-base with master ($merge_base_ref)"
			git reset "$merge_base_ref"
			diff_ref='HEAD'
		else
			diff_ref='HEAD~'
		fi
	else
		echo 'found uncommitted changes'
		diff_ref='HEAD'
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
