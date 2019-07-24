#!/bin/sh

# Runs golangci-lint to find style and correctness issues.

set -e
cd `dirname "$0"`'/..'

go mod download
golangci-lint run --deadline 5m ./...

echo 'Lint passed.'
