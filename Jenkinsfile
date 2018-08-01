// Copyright IBM Corp All Rights Reserved
//
// SPDX-License-Identifier: Apache-2.0
//
node ('hyp-x') { // trigger build on x86_64 node
     def ROOTDIR = pwd() // workspace dir (/w/workspace/<job_name>
     env.PROJECT_DIR = "gopath/src/github.com/hyperledger"
     def failure_stage = "none"
// delete working directory
     deleteDir()
      stage("Fetch Patchset") { // fetch gerrit refspec on latest commit
          try {
              dir("${ROOTDIR}"){
              sh '''
                 [ -e gopath/src/github.com/hyperledger/fabric-chaincode-evm ] || mkdir -p $PROJECT_DIR
                 cd $PROJECT_DIR
                 git clone git://cloud.hyperledger.org/mirror/fabric-chaincode-evm && cd fabric-chaincode-evm
                 git checkout "$GERRIT_BRANCH" && git fetch origin "$GERRIT_REFSPEC" && git checkout FETCH_HEAD
              '''
              }
          }
          catch (err) {
                 failure_stage = "Fetch patchset"
                 throw err
          }
      }
// clean environment, get env data and set Gopath
      stage("CleanEnv - GetEnv - SetGopath") {
          try {
                 dir("${ROOTDIR}/$PROJECT_DIR/fabric-chaincode-evm/scripts/jenkins_scripts") {
                 sh './CI_Script.sh --clean_Environment --env_Info --SetGopath'
                 }
          }
          catch (err) {
                 failure_stage = "Clean Environment - Get Env Info - SetGopath"
                 throw err
          }
      }

// Build Build Images
      stage("BuildImages") {
          try {
                 dir("${ROOTDIR}") {
                 sh '''
                    [ -e gopath/src/github.com/hyperledger/fabric ] || mkdir -p $PROJECT_DIR
                    cd $PROJECT_DIR
                    git clone git://cloud.hyperledger.org/mirror/fabric && cd fabric
                    git checkout "$GERRIT_BRANCH" && make buildenv
                    cd $PROJECT_DIR/fabric-chaincode-evm && make docker
                 }
          }
          catch (err) {
                 failure_stage = "build images"
                 throw err
          }
      }

// Run basic-checks
      stage("basic-checks") {
          try {
                 dir("${ROOTDIR}/$PROJECT_DIR/fabric-chaincode-evm") {
                 sh '''
                    make basic-checks
                 }
          }
          catch (err) {
                 failure_stage = "basic-checks"
                 throw err
          }
      }

// Run unit-tests (unit-tests)
      stage("Unit-Tests") {
          try {
                 dir("${ROOTDIR}/$PROJECT_DIR/fabric-chaincode-evm") {
                 sh '''
                    make unit-tests
                 }
          }
          catch (err) {
                 failure_stage = "unit-tests"
                 throw err
          }
      }

// Run integration tests (e2e tests)
      stage("IntegrationTests") {
          try {
                 dir("${ROOTDIR}/$PROJECT_DIR/fabric-chaincode-evm") {
                 sh '''
                    make integration-test
                 }
          }
          catch (err) {
                 failure_stage = "integration-test"
                 throw err
          }
      }
} // node block end here
