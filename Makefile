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
        github.com/qredo/signing-agent/cmd/service

test: unittest apitest

unittest:
	@echo "running unit tests"
	go test ${UNITTESTS} -v -short=t

apitest:
	@echo "running tests in ./tests/restapi"
	go test ./tests/restapi -v -short=t

e2etest:
	@echo "environment variable for APIKEY and BASE64PKEY are needed for e2e tests"
	go test ./tests/e2e -v -short=t

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

swagger:
	@docs/swagger-generate.sh
