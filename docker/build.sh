#!/bin/bash
set -xeuo pipefail
pushd ..
env GOOS=linux GOARCH=amd64 go build -ldflags '-s' .
mv easyssh docker
popd
docker build .
