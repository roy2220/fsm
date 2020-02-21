.PHONY: all
all: vet lint test

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	go run golang.org/x/lint/golint -set_exit_status ./...

.PHONY: test
test:
	go test -coverprofile=coverage.txt -covermode=count ./...
