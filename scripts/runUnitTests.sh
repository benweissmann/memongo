#!/bin/sh

# Runs unit tests via go test

set -e
cd `dirname "$0"`'/..'

# Suppress gin logging and startup messages
export GIN_MODE=test

go test ./... -cover -race
