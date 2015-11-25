#!/bin/bash

set -ueo pipefail

docker-compose build
docker-compose scale server=2
docker-compose run \
               --entrypoint /usr/local/bin/easyssh \
               client \
               -e '(if-command (ssh-exec-parallel) (if-one-target (ssh-login) (tmux-cssh)))' \
               server_1,server_2 hostname
