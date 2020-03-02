.PHONY: all
all: vet lint test

.PHONY: vet
vet: internal/list/list32.go internal/list/list32_test.go
	go vet ./...

.PHONY: lint
lint:
	go run golang.org/x/lint/golint -set_exit_status ./...

.PHONY: test
test:
	go test -coverprofile=coverage.txt -covermode=count ./...

internal/list/list32.go: internal/list/list64.go
	sed -e 's/64/32/g' $< > $@

internal/list/list32_test.go: internal/list/list64_test.go
	sed -e 's/64/32/g' $< > $@
