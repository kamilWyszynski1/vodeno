GO := $(shell which go)
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
GOLINT := $(GOBIN)/golangci-lint
BINARY := vodeno
SRCDIRS += pkg

.PHONY: build
build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GO) build -a -o $(BINARY) ./cmd

.PHONY: docker-build
docker-build:
	docker build . -t vodeno

.PHONY: lint-install
lint-install:
	test -e $(GOLINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.42.0

.PHONY: lint
lint: lint-install
	$(GOLINT) run

.PHONY: test
test:
	$(GO) test -race -v $(addprefix ./,$(addsuffix ...,$(SRCDIRS)))

.PHONY: generate
generate:
	$(GO) generate ./...