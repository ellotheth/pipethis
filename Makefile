bin = pipethis
build = $(shell git describe --tags)-$(shell go env GOOS)-$(shell go env GOARCH)
goversion = $(word 3, $(shell go version))
dist = $(bin)-$(build).tar.bz2
files = $(shell go list ./... | grep -v vendor)

all: test build

.PHONY: test
test: deps lint 
	go test -cover -race $(files)

.PHONY: build
build: deps
	go build -o $(bin) -ldflags "-w -s -X main.bin=$(bin) -X main.build=$(build) -X main.builder=$(goversion)"

.PHONY: clean
clean: dist-clean
# go clean -r hits a bunch of the stdlib, which isn't ideal
	go clean -i ./...
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
	@go get honnef.co/go/tools/cmd/gosimple
	@go get honnef.co/go/tools/cmd/staticcheck
	@if gofmt -l -s . |grep -v vendor/; then \
		echo "found formatting errors. run gofmt -s -d ." && exit 1; \
	fi
	go vet $(files)
	staticcheck $(files)
	golint $(files)
	gosimple $(files)

.PHONY: deps
deps:
	@go get github.com/golang/dep/cmd/dep
	dep ensure

.PHONY: watch
watch:
	@while true; do \
		go test -cover -race $(files); \
		echo watching for changes...; \
		inotifywait -qqre modify,create,delete,move --exclude "\.git" .; \
	done
