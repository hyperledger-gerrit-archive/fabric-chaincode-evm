# Copyright IBM Corp All Rights Reserved.
# Copyright London Stock Exchange Group All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# -------------------------------------------------------------
# This makefile defines the following targets
#
#   - all (default) - builds all targets and runs all tests/checks
#   - basic-checks - performs basic checks like license, spelling and linter
#   - check-deps - check for vendored dependencies that are no longer used
#   - checks - runs all non-integration tests/checks
#   - gotools - installs go tools like golint
#   - license - checks go source files for Apache license header
#   - linter - runs all code checks
#   - unit-test - runs the go-test based unit tests
#   - integration-test - runs the e2e_cli based test
#		- ccenv - pulls the latest docker ccenv image
#
ARCH=$(shell go env GOARCH)
BASEIMAGE_RELEASE=0.4.8
BASE_DOCKER_NS ?= hyperledger


PACKAGES = ./statemanager/... ./evmcc/... ./fabproxy/

EXECUTABLES ?= go git curl docker
K := $(foreach exec,$(EXECUTABLES),\
	$(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH: Check dependencies")))

all: checks integration-test

checks: basic-checks unit-test

basic-checks: license spelling linter

.PHONY: spelling
spelling:
	@scripts/check_spelling.sh

.PHONY: license
license:
	@scripts/check_license.sh

include gotools.mk

.PHONY: gotools
gotools: gotools-install

unit-test: $(PROJECT_FILES)
	@echo "Running unit-tests"
	@go test -tags "$(GO_TAGS)" $(PACKAGES)

unit-tests: unit-test

linter: check-deps
	@echo "LINT: Running code checks.."
	@scripts/golinter.sh

check-deps:
	@echo "DEP: Checking for dependency issues.."
	@scripts/check_deps.sh

changelog:
	@scripts/changelog.sh v$(PREV_VERSION) v$(BASE_VERSION)

ccenv:
	docker pull $(BASE_DOCKER_NS)/fabric-ccenv:latest
	docker tag $(BASE_DOCKER_NS)/fabric-ccenv:latest $(BASE_DOCKER_NS)/fabric-ccenv

.PHONY: integration-test
integration-test: ccenv 
	@echo "Running integration-test"
	@scripts/run-integration-tests.sh
