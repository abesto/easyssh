#!/usr/bin/env bats

load common

@test "go tool vet" {
    godep go tool vet .
}