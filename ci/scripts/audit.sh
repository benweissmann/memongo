#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-mongodb-in-memory
  make audit
popd