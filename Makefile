# Makefile for Signing Agent
BUILD_VERSION=$(shell git rev-list -1 --abbrev-commit HEAD)
BUILD_TYPE="dev"
BUILD_DATE=$(shell date -u)
LDFLAGS = "-X 'main.buildVersion=${BUILD_VERSION}' -X 'main.buildDate=${BUILD_DATE}' -X 'main.buildType=${BUILD_TYPE}'"
UNITTESTS=$(shell go list ./... | grep -v tests/)

build:
	go mod tidy
	go build \
	    -tags debug \
	    -ldflags ${LDFLAGS} \
        -o out/signing-agent \
        github.com/qredo/signing-agent/cmd/app

test: unittest

unittest:
	@echo "running unit tests"
	go test ${UNITTESTS} -v -short=t

integrationtest:
	@echo "running integration tests"
	go test ./tests/integration -v -short=t

update-packages:
	@echo "updating all go packages"
	go get -u ./...
	go mod tidy

test-all:
	@echo "running all tests"
	go test ./... -v -count=1

lint:
	@echo "running lint"
	golangci-lint run
