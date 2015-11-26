#!/bin/bash

bats test/unit-test.bats \
     test/code-quality.bats \
     test/integration-test.bats
