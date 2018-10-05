#!/bin/bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Use ginkgo to run integration tests. If arguments are provided to the
# script, they are treated as the directories containing the tests to run.
# When no arguments are provided, all integration tests are executed.

set -e -u

fabric_chaincode_evm_dir="$(cd "$(dirname "$0")/.." && pwd)"
FABRIC_DIR=${GOPATH}/src/github.com/hyperledger/fabric

# find packages that contain "integration" in the import path
integration_dirs() {
    local packages="$1"

    go list -f {{.Dir}} "$packages" | grep -E '/integration($|/)' | sed "s,${fabric_chaincode_evm_dir},.,g"
}

main() {
    cd "$fabric_chaincode_evm_dir"

    local -a dirs=("$@")
    if [ "${#dirs[@]}" -eq 0 ]; then
        dirs=($(integration_dirs "./..."))
    fi

    if [ ! which node> /dev/null 2>&1 ]; then
        echo "No node in PATH. Check dependencies"
        exit 1
    fi

    if [ ! npm ls -g web3 | grep "web3@0.20.2"]; then
        npm install -g web3@0.20.2
    fi

    #Check if Fabric is in the gopath. Fabric needs to be in the gopath for the integration tests
    if [ ! -d "${FABRIC_DIR}" ]; then
        echo "Downloading Fabric"
        git clone https://github.com/hyperledger/fabric $FABRIC_DIR
    fi

    echo "Building CCENV image"
    pushd ${FABRIC_DIR}
        make ccenv
    popd

    echo "Running integration tests..."
    ginkgo -race -keepGoing --slowSpecThreshold 80 -r "${dirs[@]}"
}

main "$@"
