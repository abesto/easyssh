#!/usr/bin/env bats

load common

@test "go tool vet" {
    go tool vet $(for dir in $(go list -f '{{.Dir}}' ./...); do echo "${dir#$(pwd)/}"; done | grep -vE "^$(pwd)$|^vendor/")
}