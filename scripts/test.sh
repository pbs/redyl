#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

>&2 echo "running redyl tests"
go test -v ./...
