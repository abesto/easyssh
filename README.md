# easyssh

[![Build Status](https://travis-ci.org/abesto/easyssh.svg?branch=master)](https://travis-ci.org/abesto/easyssh)
[![Coverage Status](https://coveralls.io/repos/abesto/easyssh/badge.svg?branch=master)](https://coveralls.io/r/abesto/easyssh?branch=master)

`easyssh` is the culmination of several years of SSH-related aliases. It's a highly configurable wrapper around `ssh`, `tmux-cssh`, `csshx`, `aws`, `knife` and whatever else.

It's for you if having a single alias that does the following makes you excited:

 * `s myhost.com` logs in to myhost.com
 * `s app.myhost.com,db.myhost.com` logs in to both hosts using `tmux-ssh`
 * `s -lroot myhost.com /etc/init.d/apache2 reload` reloads apache
 * `s app.myhost.com,db.myhost.com uptime` runs uptime on both hosts (parallelly, which is interesting if you run longer-running commands)
 * `s -lroot roles:app /etc/init.d/apache2 reload` parallelly reloads apache on all nodes that have the role `app` in Chef
 * If you provide a hostname that looks like it includes an EC2 instance id, it uses the `aws` CLI tool to look up the public IP, and uses that.

## Installation

1. Download, extract the [release](https://github.com/abesto/easyssh/releases) for your platform
2. Add the `easyssh` executable to your `$PATH`

Of course you can always compile from source:

```sh
go get -u github.com/abesto/easyssh
```

## Inline usage

You probably won't ever do this; it's only here for a basic demonstration of the syntax.

```sh
# log in with an interactive shell; old-fashioned ssh
easyssh myhost.com
# sequentially run "hostname" on all nodes matching "roles:app" according to knife
easyssh -c='(ssh-exec)' -d='(knife)' roles:app hostname
# parallelly reload apache on all your app servers (as root, of course)
easyssh -c='(ssh-exec-parallel)' -d='(knife)' -l=root roles:app /etc/init.d/apache2 reload
```

## Suggested usage

This one alias implements the use-case described in the introduction. Writing aliases like this is the suggested
way of using `easyssh`.

```sh
easyssh_executor='(if-command (ssh-exec-parallel) (if-one-target (ssh-login) (tmux-cssh)))'
easyssh_discoverer='(first-matching (knife) (comma-separated))'
easyssh_filter='(list (ec2-instance-id us-east-1) (ec2-instance-id us-west-1))'
alias s="easyssh -e='$easyssh_executor' -d='$easyssh_discoverer' -f='$easyssh_filter'"
```

If you frequently log in to servers as root:

```sh
alias sr='s -l root'
# reload apache on app servers (as root)
sr roles:app /etc/init.d/apache2 reload
```

This assumes that

 * `knife` is correctly configured for the Chef environment you want to work with
 * The `aws` CLI tool is correctly configured
 * You have `tmux-cssh` installed
 * Your EC2 nodes are in `us-east-1` and `us-west-1`.

That's a lot of assumptions; `easyssh` tries to be as general as possible, but in the end, you get to tailor it to your
specific needs. Which brings us to...

## Configuration

The behavior of `easyssh` is controlled by three components:

 * *Discoverers* produce a list of targets to run against from some input string; this input string is the first
   non-flag argument to `easyssh`
 * *Filters* mutate the list of targets produced by the discoverers
 * *Executors* do something with the targets, optionally taking the arguments not consumed by the discoverers.

```
easyssh -d='(ddef)' -f='(fdef)' -e='(edef)' targetdef [cmd to run]
          |           |           |             |           |
          ▼           ▼           ▼             |           |
      Discoverer---▶Filter----▶Executor◀--------+------------
          ▲                                     |
          |                                     |
          ---------------------------------------
```

Discoverer, filter and executor definitions are [S-Expressions](https://en.wikipedia.org/wiki/S-expression); the terms usable in them are described below. Wherever a "string" is referenced, you don't need to quote the string (but you can). For example, `(const foo "bar")` is fine.

### Discoverers

| Name      | Arguments   | Description |
|-----------|-------------|-------------|
| `comma-separated` | - | Takes the discoverer argument, splits it at commas, and uses the resulting strings as the target hostnames. |
| `knife` | - | Passes the discoverer argument to `knife search node`, and returns the public IP addresses provided by Chef as target hostnames. |
| `knife-hostname` | - | Same as `knife`, except it collects public DNS names instead of IP addresses. |
| `first-matching` | Any number of discoverers | Runs the discoverers in its argument list in the order they were provided, and uses the first resulting non-empty target list. |
| `fixed` | At least one string | Alias: `const`. Returns its arguments as hostnames, regardless of the target definition. |

### Filters

| Name      | Arguments   | Description |
|-----------|-------------|-------------|
| `id` | - | Doesn't touch the the target list. |
| `first` | - | Drops all targets in the target list, except for the first one. |
| `ec2-instance-id` | AWS region | For each target in the target list, it looks for an EC2 instance id in the target name. If there is one, it uses `aws` to look up its public IP and replace the target name with that. |
| `list` | Any number of filters | Applies each filter in its arguments to the target list. |
| `external` | At least one string | Calls the command specified in the arguments with a file containing the targets before filtering. The command must output the new targets on its STDOUT. For example: `(external percol)` |

### Executors

| Name      | Arguments   | Command | Description |
|-----------|-------------|---------|-------------|
| `ssh-login` | - | rejects | Logs in to each target sequentially using SSH |
| `ssh-exec` | - | requires | Executes the command on each target sequentially |
| `ssh-exec-parallel` | - | requires | Executes the command on each target parallelly |
| `csshx` | - | rejects | Uses `csshx` to log in to all the targets |
| `tmux-cssh` | - | rejects | Uses `tmux-cssh` to log in to all the targets |
| `if-one-target` | exactly two executors | pass-through | If there's one target, it calls the executor in its first argument. Otherwise the executor in its second argument. |
| `if-command` | Exactly two executors | pass-through | Alias: `if-args`. If a command was defined, it calls the executor in its first argument. Otherwise the executor in its second argument. |

Most of these executors are in fact constructed using the following low-level executors:

| Name      | Arguments   | Description |
|-----------|-------------|-------------|
| `assert-command` | Exactly one executor | Fails if no command was provided; calls its argument otherwise. |
| `assert-no-command` | Exactly one executor | Fails if a command was provided; calls its argument otherwise. |
| `external` | At least one string | Runs `arguments targets command`, prefixing each output line with a timestamp and the hostname. For example `(external csshx)` would run `csshx host1 host2 command`. |
| `external-interactive` | At least one string | Same as `external`, except output lines are not prefixed. |
| `external-sequential` | At least one string | Runs its arguments against each target sequentially, prefixing each output line with a timestamp and the hostname. For example `(external-sequential ssh)` with the command `hostname` would run `ssh host1 hostname; ssh host2 hostname; ...`||
| `external-sequential-interactive` | At least one string | Same as `external-sequential`, except output lines are not prefixed. |
| `external-parallel` | At least one string | Runs its arguments against each target parallelly, prefixing each output line with a timestamp and the hostname. |

For example `tmux-cssh` is `(assert-no-command (external-interactive tmux-cssh))`.

## Contributing

All feedback and feature requests are welcome. Pull-requests are even more welcome :)
See [CONTRIBUTING.md](CONTRIBUTING.md) for a quick rundown of how and where to start.

## License

`easyssh` is licensed under [ISC](http://opensource.org/licenses/ISC). See the [LICENSE.txt](LICENSE.txt) file.