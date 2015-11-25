#!/bin/bash

set -ueo pipefail

env GOOS=linux go build
mv easyssh integration-test/client/easyssh

cd integration-test
trap "rm client/easyssh" EXIT

./run.sh
