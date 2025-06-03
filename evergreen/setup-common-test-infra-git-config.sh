#!/bin/bash

echo "Setting global git config to use token: $common_test_infra_github_token"
git config --global url."https://x-access-token:$common_test_infra_github_token@github.com/10gen/sql-engines-common-test-infra".insteadOf https://github.com/10gen/sql-engines-common-test-infra

echo "git config:"
git config --list
