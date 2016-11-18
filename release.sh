#!/bin/bash
set -ueo pipefail

version="$1"

./test.sh

go get -u github.com/tcnksm/ghr
go get -u github.com/mitchellh/gox
gox -output "dist/{{.OS}}_{{.Arch}}_{{.Dir}}"

echo "Good, now update Homebrew!"
