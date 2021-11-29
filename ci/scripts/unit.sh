#!/bin/bash -eux

pushd dp-mongodb-in-memory
  make test
popd
