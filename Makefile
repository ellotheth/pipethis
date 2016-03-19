bin = pipethis
build = $(shell git describe --tags)-$(shell go env GOOS)-$(shell go env GOARCH)
goversion = $(word 3, $(shell go version))
dist = $(bin)-$(build).tar.bz2

all: test build

.PHONY: test
test: deps lint 
	@go get github.com/stretchr/testify/assert
	@go test -cover -race ./...

.PHONY: build
build: deps
	go build -o $(bin) -ldflags "-w -X main.bin=$(bin) -X main.build=$(build) -X main.builder=$(goversion)"

.PHONY: clean
clean: dist-clean
# go clean -r hits a bunch of the stdlib, so we go nuclear!
	go clean -i
	rm -rf $(GOPATH)/pkg/*
	rm -f $(bin)*

.PHONY: dist
dist: build
	tar -jcf $(dist) $(bin)
	gpg --detach-sign --armor --output $(dist).sig $(dist)

.PHONY: dist-clean
dist-clean:
	rm -f $(dist)*

.PHONY: lint
lint:
	@go get github.com/golang/lint/golint
	@go fmt ./...
	@go vet ./...
	@golint ./...

.PHONY: deps
deps:
	@go get -v

.PHONY: watch
watch:
	@while true; do \
		make test; \
		inotifywait -qqre modify,create,delete,move .; \
		echo watching for changes...; \
	done
