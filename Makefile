.PHONY: all
all: vet lint test

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	go run golang.org/x/lint/golint -- ./...

.PHONY: test
test:
	go test -coverprofile=coverage.txt -covermode=count ./...
