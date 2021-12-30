SHELL:=/bin/sh
.PHONY: build_server

export GO111MODULE=on

# Path Related
MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))

# Version
RELEASE?=v1.0.0


# Rules
build: build_server

build_server:
	@echo "build pando server"
	cd ${MKFILE_DIR}cmd/server && \
	go mod vendor && \
	go build -o pando-server
