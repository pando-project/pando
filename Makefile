SHELL:=/bin/sh
.PHONY: build build_client build_server \
		test run fmt vet clean \
		mod_update vendor_from_mod vendor_clean end

export GO111MODULE=on

PROJECT_NAME := pando

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# Path Related
MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
RELEASE_DIR := ${MKFILE_DIR}bin
GO_PATH := $(shell go env | grep GOPATH | awk -F '"' '{print $$2}')

# Version
RELEASE?=v1.0.0

# Git Related
GIT_REPO_INFO=$(shell cd ${MKFILE_DIR} && git config --get remote.origin.url)
ifndef GIT_COMMIT
  GIT_COMMIT := git-$(shell git rev-parse --short HEAD)
endif

# Build Flags
GO_LD_FLAGS= "-s -w -X github.com/kenlabs/pando/pkg/version.RELEASE=${RELEASE} -X github.com/kenlabs/pando/pkg/version.COMMIT=${GIT_COMMIT} -X github.com/kenlabs/pando/pkg/version.REPO=${GIT_REPO_INFO}"

# Cgo is disabled by default
ENABLE_CGO= CGO_ENABLED=0

# Check Go build tags, the tags are from command line of make
ifdef GOTAGS
  GO_BUILD_TAGS= -tags ${GOTAGS}
  # Must enable Cgo when wasmhost is included
  ifeq ($(findstring wasmhost,${GOTAGS}), wasmhost)
	ENABLE_CGO= CGO_ENABLED=1
  endif
endif

# When build binaries for docker, we put the binaries to another folder to avoid
# overwriting existing build result, or Mac/Windows user will have to do a rebuild
# after build the docker image, which is Linux only currently.
ifdef DOCKER
  RELEASE_DIR=${MKFILE_DIR}build/bin
endif

# Targets
TARGET_SERVER=${RELEASE_DIR}/pando-server
TARGET_CLIENT=${RELEASE_DIR}/pando-client

# Rules

## build: build pando client and server cli
build: build_client build_server end

## build_client: build pando client cli
build_client:
	@echo " > Build client..."
	cd ${MKFILE_DIR} && \
	CGO_ENABLED=0 go build -v -trimpath -ldflags ${GO_LD_FLAGS} \
	-o ${TARGET_CLIENT} ${MKFILE_DIR}cmd/client

## build_server: build pando server cli
build_server:
	@echo " > Build server..."
	cd ${MKFILE_DIR} && \
	${ENABLE_CGO} go build ${GO_BUILD_TAGS} -v -trimpath -ldflags ${GO_LD_FLAGS} \
	-o ${TARGET_SERVER} ${MKFILE_DIR}cmd/server

dev_build: dev_build_client dev_build_server

dev_build_client:
	@echo " > Build dev client..."
	cd ${MKFILE_DIR} && \
	go build -v -race -ldflags ${GO_LD_FLAGS} \
	-o ${TARGET_CLIENT} ${MKFILE_DIR}cmd/client

dev_build_server:
	@echo " > Build dev server..."
	cd ${MKFILE_DIR} && \
	go build -v -race -ldflags ${GO_LD_FLAGS} \
	-o ${TARGET_SERVER} ${MKFILE_DIR}cmd/server

## test: run unit test
test:
	cd ${MKFILE_DIR}
	go mod tidy
	git diff --exit-code go.mod go.sum
	go mod verify
	go test -v ./... ${TEST_FLAGS}

## clean: remove all binaries and cache files
clean:
	rm -rf ${RELEASE_DIR}
	rm -rf ${MKFILE_DIR}build/cache
	rm -rf ${MKFILE_DIR}build/bin

run: build_server

fmt:
	cd ${MKFILE_DIR} && go fmt ./...

vet:
	cd ${MKFILE_DIR} && go vet ./...

vendor_from_mod:
	cd ${MKFILE_DIR} && go mod vendor

vendor_clean:
	rm -rf ${MKFILE_DIR}vendor

mod_update:
	cd ${MKFILE_DIR} && go get -u

end:
	@echo " > Binaries built at $(RELEASE_DIR)"
	@echo " > Done"

help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECT_NAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo