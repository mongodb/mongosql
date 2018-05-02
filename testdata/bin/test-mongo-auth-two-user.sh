#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"

(
    set -o errexit
    echo "connecting both clients to mysql"

	nohup mysql $CLIENT_ARGS -p"$MYSQL_PWD" -e 'select sleep(1000000)' &> /dev/null &
	nohup mysql $SECOND_CLIENT_ARGS -p"$SECOND_MYSQL_PWD" -e 'select sleep(1000000)' &> /dev/null &

	cmd="use information_schema; select count(*) from processlist"
    set +o errexit
	process_list_result_user_1=$(mysql --skip-column-names --silent $CLIENT_ARGS -p"$MYSQL_PWD" -e "$cmd" 2> /dev/null)
    code1=$?
	process_list_result_user_2=$(mysql --skip-column-names --silent $SECOND_CLIENT_ARGS -p"$SECOND_MYSQL_PWD" -e "$cmd" 2> /dev/null)
    code2=$?
    set -o errexit

	if [ "$process_list_result_user_1" != "$PROCESS_COUNT_1" ]; then
		echo "expected '$PROCESS_COUNT_1' process, but got '$process_list_result_user_1'"
		exit 1
	fi
	if [ "$process_list_result_user_2" != "$PROCESS_COUNT_2" ]; then
		echo "expected '$PROCESS_COUNT_2' process, but got '$process_list_result_user_2'"
		exit 1
	fi

	cmd_1="use information_schema; kill connection (select Id from processlist where User = '$USER_TO_KILL_1')"
	cmd_2="use information_schema; kill connection (select Id from processlist where User = '$USER_TO_KILL_2')"
    set +o errexit
	kill_connection_result_user_1=$(mysql --skip-column-names --silent $CLIENT_ARGS -p"$MYSQL_PWD" -e "$cmd_1" 2> /dev/null)
	kill_connection_result_user_2=$(mysql --skip-column-names --silent $SECOND_CLIENT_ARGS -p"$SECOND_MYSQL_PWD" -e "$cmd_2" 2>&1)
    set -o errexit

	if [ "$kill_connection_result_user_1" != "$EXPECTED_KILL_1" ]; then
		echo "expected '$EXPECTED_KILL_2' process, but got '$kill_connection_result_user_1'"
		exit 1
	fi
	# must use regexp matching here, globbing only works with strings.
	if ! [[ "$kill_connection_result_user_2" =~ .*"$EXPECTED_KILL_2".* ]]; then
		echo "expected '$EXPECTED_KILL_2' process, but got '$kill_connection_result_user_2'"
		exit 1
	fi

) > $LOG_FILE 2>&1

print_exit_msg
