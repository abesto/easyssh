#!/usr/bin/env bats

load common

setup() {
    env GOOS=linux go build
    mv easyssh test/integration-test/client/easyssh
    
    cd test/integration-test
    
    docker-compose build
    docker-compose scale server=2
}

teardown() {
    rm client/easyssh
}

matches() {
    line_no="$1"
    regex="$2"
    line="${lines[$line_no]}"
    echo "$line" | sed -n "/^${regex}$/p" | wc -l
}

timestamp="\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}.\d{3}"
ip="[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}"
container="[a-z0-9]{12}"

@test "Smoke test, log level info" {
    run docker-compose run \
        --entrypoint /usr/local/bin/easyssh \
        client \
        -e '(ssh-exec-sequential)' \
        server_1,server_2 hostname
    echo "$output"

    [ $status -eq 0 ]
    [ ${#lines[@]} -eq 7 ]
    [ $(matches 0 "$timestamp INFO Targets: \[server_1 server_2\]") ]
    [ $(matches 1 "$timestamp INFO Executing [/usr/local/bin/ssh server_1 hostname\]") ]
    [ $(matches 2 "$timestamp NOTICE \[server_1\] (STDERR) Warning: Permanently added 'server_1,$ip' (ECDSA) to the list of known hosts.") ]
    [ $(matches 3 "$timestamp NOTICE \[server_1\] (STDOUT) $container") ]
    [ $(matches 4 "$timestamp INFO Executing \[/usr/local/bin/ssh server_2 hostname\]") ]
    [ $(matches 5 "$timestamp NOTICE \[server_2\] (STDERR) Warning: Permanently added 'server_2,$ip' (ECDSA) to the list of known hosts.") ]
    [ $(matches 6 "$timestamp NOTICE \[server_2\] (STDOUT) $container") ]
}

@test "Smoke test, log level debug" {
    run docker-compose run \
        --entrypoint /usr/local/bin/easyssh \
        client \
        -e '(ssh-exec-sequential)' \
        -log debug \
        server_1,server_2 hostname
    echo "$output"

    [ $status -eq 0 ]
    [ ${#lines[@]} -eq 17 ]
    [ $(matches 0 "$timestamp DEBUG MakeFromString (comma-separated) -> \[comma-separated\]") ]
    [ $(matches 1 "$timestamp DEBUG Transform: \[comma-separated\] -> \[separated-by ,\]") ]
    [ $(matches 2 "$timestamp DEBUG Make \[separated-by ,\] -> <separated-by ,>") ]
    [ $(matches 3 "$timestamp DEBUG MakeFromString (ssh-exec-sequential) -> \[ssh-exec-sequential\]") ]
    [ $(matches 4 "$timestamp DEBUG Transform: \[ssh-exec-sequential\] -> \[assert-command \[external-sequential ssh\]\]") ]
    [ $(matches 5 "$timestamp DEBUG Make \[external-sequential ssh\] -> <external-sequential \[ssh\]>") ]
    [ $(matches 6 "$timestamp DEBUG Make \[assert-command \[external-sequential ssh\]\] -> <assert-command <external-sequential \[ssh\]>>") ]
    [ $(matches 7 "$timestamp DEBUG MakeFromString (id) -> \[id\]") ]
    [ $(matches 8 "$timestamp DEBUG Make \[id\] -> <id>") ]
    [ $(matches 9 "$timestamp DEBUG Targets before filters: \[server_1 server_2\]") ]
    [ $(matches 10 "$timestamp INFO Targets: \[server_1 server_2\]") ]
    [ $(matches 11 "$timestamp INFO Executing [/usr/local/bin/ssh server_1 hostname\]") ]
    [ $(matches 12 "$timestamp NOTICE \[server_1\] (STDERR) Warning: Permanently added 'server_1,$ip' (ECDSA) to the list of known hosts.") ]
    [ $(matches 13 "$timestamp NOTICE \[server_1\] (STDOUT) $container") ]
    [ $(matches 14 "$timestamp INFO Executing \[/usr/local/bin/ssh server_2 hostname\]") ]
    [ $(matches 15 "$timestamp NOTICE \[server_2\] (STDERR) Warning: Permanently added 'server_2,$ip' (ECDSA) to the list of known hosts.") ]
    [ $(matches 16 "$timestamp NOTICE \[server_2\] (STDOUT) $container") ]
}

