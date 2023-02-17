#!/bin/sh

set -e

BUILD_TYPE="dev"
BUILD_VERSION=$(git rev-list -1 --abbrev-commit HEAD)
BUILD_DATE="$(date -u)"
IMAGE_DATE="$(date +%F)"

rm -rf vendor

docker_latest() {
  BUILD_TYPE="latest"
  docker build --build-arg BUILD_DATE="$BUILD_DATE" \
                --build-arg BUILD_TYPE="$BUILD_TYPE" \
                --build-arg BUILD_VERSION="$BUILD_VERSION" \
                -t signing-agent:latest \
                -f dockerfiles/Dockerfile .

  rm -rf vendor
}

# Build (and import) a docker image for the local architecture
docker_local() {
  docker build --build-arg BUILD_DATE="$BUILD_DATE" \
                --build-arg BUILD_TYPE="$BUILD_TYPE" \
                --build-arg BUILD_VERSION="$BUILD_VERSION" \
                -t signing-agent:dev \
                -f dockerfiles/Dockerfile .

  rm -rf vendor
}

# Build a docker image for unit testing
docker_test_build() {
  docker build --build-arg BUILD_DATE="$BUILD_DATE" \
                --build-arg BUILD_TYPE="$BUILD_TYPE" \
                --build-arg BUILD_VERSION="$BUILD_VERSION" \
                -t signing-agent-unittest:dev \
                -f dockerfiles/DockerfileUnitTest .

  rm -rf vendor
}

# Build a docker image for the specified architecture and store it in a tar file
docker_export() {
  if test -f signing-agent-$1-*.tar; then
      rm signing-agent-$1-*.tar
  fi
  docker buildx build \
      --build-arg BUILD_DATE="$BUILD_DATE" \
      --build-arg BUILD_TYPE="$BUILD_TYPE" \
      --build-arg BUILD_VERSION="$BUILD_VERSION" \
      --platform linux/$1 \
      --output "type=docker,push=false,name=signing-agent:dev-$1,dest=signing-agent-$1-$IMAGE_DATE.tar" \
      -f dockerfiles/Dockerfile .
}

# Build a CLI docker image for the specified architecture and store it in a tar file
docker_export_cli() {
  if test -f sa-cli-$1-*.tar; then
      rm sa-cli-$1-*.tar
  fi
  docker buildx build \
      --platform linux/$1 \
      --output "type=docker,push=false,name=sa-cli:dev-$1,dest=sa-cli-$1-$IMAGE_DATE.tar" \
      -f dockerfiles/DockerfileCLI .
}

# Build docker images for all supported architectures
docker_export_allarch() {
  # We need to build the images one by one so they can be exported (doesn't work otherwise)
  # If this command fails because of buildx, please run the following command:
  # docker buildx create --use
  for arch in amd64 arm64 ; do
      docker_export $arch
  done
  rm -rf vendor
}

# Build CLI docker images for all supported architectures
docker_export_cli_allarch() {
  for arch in amd64 arm64 ; do
      docker_export_cli $arch
  done
  rm -rf vendor
}

local_cli_build() {
  go mod tidy
  go build \
      -o sa-cli \
      github.com/qredo/signing-agent/cmd/sa-cli
}

# Build a the Go binary to run in the local environment
local_build() {
  go mod tidy
  go build \
      -tags debug \
      -ldflags "-X 'main.buildDate=$BUILD_DATE' \
                -X 'main.buildVersion=$BUILD_VERSION' \
                -X 'main.buildType=$BUILD_TYPE'" \
      -o out/signing-agent \
      github.com/qredo/signing-agent/cmd/service
}

if [ -n "$1" ]; then
  case $1 in
    docker)
      docker_local
      ;;
    docker_latest)
      docker_latest
      ;;
    docker_amd64)
      docker_export amd64
      ;;
    docker_arm64)
      docker_export arm64
      ;;
    docker_multiarch)
      docker_export_allarch
      ;;
    local_cli)
      local_cli_build
      ;;
    docker_export_cli_allarch)
      docker_export_cli_allarch
      ;;
    docker_unittest)
      docker_test_build
      ;;
  esac
else
  local_build
fi
