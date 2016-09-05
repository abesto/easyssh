#!/usr/bin/env bats
# Shamelessly ripped from https://github.com/sdboyer/gogl/blob/master/test-coverage.sh
# run on CI service w/something like:
#
# go get github.com/axw/gocov/gocov
# go get github.com/mattn/goveralls
# COVERALLS="-service drone.io -repotoken $COVERALLS_TOKEN" ./test-coverage.sh
#
# NOTE: go files in the root and in the dir "utils" are ignored for purposes of
#       test coverage calculation.

load common


@test "Unit tests" {
    export PATH="$(pwd)/test/noop-external-tools:$PATH"

    echo "mode: count" > acc.out
    fail=0
    
    want-coverage-for() {
        if [[ $1 == . || $1 == ./util ]]; then
            return 1
        fi
    }
    
    for dir in $(for dir in $(go list -f '{{.Dir}}' ./...); do echo "${dir#$(pwd)/}"; done | grep -vE "^$(pwd)$|^vendor/")
    do
        if ls $dir/*.go &> /dev/null; then
            go test -covermode=count -coverprofile=profile.out ./$dir || fail=1
            if [ -f profile.out ] && want-coverage-for ./$dir
            then
                cat profile.out | grep -v "^mode: " | grep -v "test_helpers.go" >> acc.out
                rm profile.out
            fi
        fi
    done
    
    # Failures have incomplete results, so don't send
    if [ "$fail" -eq 0 -a "$COVERALLS" ]; then
        goveralls -v -coverprofile=acc.out $COVERALLS
    fi
    
    rm -f acc.out
    
    [ $fail -eq 0 ]
}
