override SHELL := /usr/bin/env bash -euxo pipefail
override .DEFAULT_GOAL := all

ifdef USE_DOCKER

%: force
	@scripts/make-with-docker.sh $@

Makefile:
	# this is a buggy target

else # ifdef USE_DOCKER

all: force vet lint test

vet: force internal/list/list32.go internal/list/list32_test.go
	@go vet ./...

lint: force
	@go run golang.org/x/lint/golint -set_exit_status ./...

test: force
	@go test -coverprofile=coverage.txt -covermode=count ./...

internal/list/list32.go: internal/list/list64.go
	@sed -e 's/64/32/g' $< > $@

internal/list/list32_test.go: internal/list/list64_test.go
	@sed -e 's/64/32/g' $< > $@

endif # ifdef USE_DOCKER

.PHONY: force
force:
