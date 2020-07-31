#!/bin/bash

. "$(dirname $0)/platforms.sh"
. "$(dirname $0)/prepare-shell.sh"


    set -o errexit
    set -o verbose

    KEY_TAB="${PROJECT_DIR}/testdata/resources/gssapi/drivers.keytab"
    kinit -k -t ${KEY_TAB} ${USER}@LDAPTEST.10GEN.CC




