#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# tooling logic borrowed heavily from the talented minds of confd
# https://github.com/kelseyhightower/confd/blob/master/Makefile

cd "$(dirname "${BASH_SOURCE[0]}")/.."

./scripts/clean.sh
mkdir -p bin

VERSION=$(grep -E -o '[0-9]+\.[0-9a-z.\-]+' ./internal/redyl/version/version.go)

echo >&2 "building redyl_builder docker image"
# build a local docker image called redyl_builder: we'll use this to
# compile our release binaries
docker build -t redyl_builder -f Dockerfile.build.alpine .

echo >&2 "compiling binaries for release"
# for each of our target platforms we use the redyl_builder
#   docker container to compile a binary of our application
for platform in darwin linux windows; do
    for arch in amd64 arm64; do
        binary_name="redyl-${VERSION}-${platform}-${arch}"
        echo >&2 "compiling $binary_name"

        # * GOOS is the target operating system
        # * GOARCH is the target processor architecture
        #     see https://golang.org/cmd/go/#hdr-Environment_variables
        # * CGO_ENABLED controls whether the go compiler allows us to
        #     import C packages (we don't do this, so we set it to 0 to turn CGO off)
        #     see https://golang.org/cmd/cgo/
        docker run -it --rm \
            -v "${PWD}:/app" \
            -e "GOOS=$platform" \
            -e "GOARCH=$arch" \
            -e "CGO_ENABLED=0" \
            redyl_builder \
            go build -o "bin/$binary_name"
    done
done
