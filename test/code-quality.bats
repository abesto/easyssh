#!/usr/bin/env bats

load common

@test "go tool vet" {
    go tool vet $(go list -f '{{.Dir}}' ./... | xargs realpath --relative-to=$(pwd) | grep -v '^vendor/' | grep -v '^\.$')
}