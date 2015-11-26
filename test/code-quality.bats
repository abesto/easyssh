#!/usr/bin/env bats

load common

@test "go tool vet" {
    go tool vet .
}