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
 * `s i-deadbeef` looks up the EC2 instance id using `aws`, and logs in to the host

## Installation

### Homebrew

```
brew tap abesto/easyssh
brew install easyssh
```

### Binary release

1. Download, extract the [release](https://github.com/abesto/easyssh/releases) for your platform
2. Add the `easyssh` executable to your `$PATH`

### Compile from source

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
| `separated-by` | Exactly one string | Splits the input at the separator provided as the argument, and uses the resulting strings as the target hosts. |
| `comma-separated` | - | `(separated-by ,)` |
| `knife` | - | Passes the discoverer argument to `knife search node`, and returns the public IP addresses provided by Chef as target hosts. |
| `first-matching` | Any number of discoverers | Runs the discoverers in its argument list in the order they were provided, and uses the first resulting non-empty target list. |
| `fixed` | At least one string | Alias: `const`. Returns its arguments as hosts, regardless of the target definition. |

### Filters

| Name      | Arguments   | Description |
|-----------|-------------|-------------|
| `id` | - | Doesn't touch the the target list. |
| `first` | - | Drops all targets in the target list, except for the first one. |
| `ec2-instance-id` | AWS region | For each target in the target list, it looks for an EC2 instance id in the target name. If there is one, it uses `aws` to look up its public DNS name and IP, and sets the target host and IP to these values. |
| `coalesce` | At least one string | The argument is a list of values from `host`, `hostname`, and `ip`. When accessing the target, the first non-empty field of the target will be used from the parameter list of `coalesce`. |
| `list` | Any number of filters | Applies each filter in its arguments to the target list. |
| `external` | At least one string | Calls the command specified in the arguments with a file containing the targets before filtering. The command must output the new targets on its STDOUT. For example: `(external percol)` |

### Executors

First, some executors that integrate "external" tools, like `ssh` and `csshx`:

| Name      | Arguments   | Command | Description |
|-----------|-------------|---------|-------------|
| `ssh-login` | - | rejects | Logs in to each target sequentially using SSH |
| `ssh-exec` | - | requires | Executes the command on each target sequentially |
| `ssh-exec-parallel` | - | requires | Executes the command on each target parallelly |
| `csshx` | - | rejects | Uses `csshx` to log in to all the targets |
| `tmux-cssh` | - | rejects | Uses `tmux-cssh` to log in to all the targets |

You can use the following combinators for run-time decisions on which executor to use:

| Name      | Arguments   | Command | Description |
|-----------|-------------|---------|-------------|
| `if-one-target` | exactly two executors | pass-through | If there's one target, it calls the executor in its first argument. Otherwise the executor in its second argument. |
| `if-command` | Exactly two executors | pass-through | Alias: `if-args`. If a command was defined, it calls the executor in its first argument. Otherwise the executor in its second argument. |

The executors calling external commands are implemented using the following low-level executors; you can use them directly if a tool you need is not supported above - feel free to send a PR to provide user-friendly support, of course.

<table>
  <tr>
    <th></th>
    <th>Single</th>
    <th>Sequential</th>
    <th>Parallel</th>
  </tr>
  <tr>
    <th>Timestamp</th>
    <td><code>execute</code></td>
    <td><code>execute-sequential</code></td>
    <td><code>execute-parallel</code></td>
  </tr>
  <tr>
    <th>Interactive</th>
    <td><code>execute-interactive</code></td>
    <td><code>execute-sequential-interactive</code></td>
    <td>-</td>
  </tr>
</table>

**Single**: A single run of the specified external command, with all the targets passed as arguments to it.<br>
**Sequential**: The command is run once for each target, sequentially.<br>
**Parallel**: The command is run once for each target, parallelly.<br>
**Timestamp**: Recommended for non-interactive tools. Each output line is prefixed with a timestamp. This is achieved by intercepting both `STDOUT` and `STDERR`. If the command does detection of terminal features, this usually results in it not emitting control sequences (ie. no colors)<br>
**Interactive**: Recommended for interactive in-terminal tools. `STDOUT` and `STDERR` are passed directly to the command.

Finally, you can use these combinators to fail early if an executor would be called incorrectly.

| Name      | Arguments   | Description |
|-----------|-------------|-------------|
| `assert-command` | Exactly one executor | Fails if no command was provided; calls its argument otherwise. |
| `assert-no-command` | Exactly one executor | Fails if a command was provided; calls its argument otherwise. |

Some examples of how these are used to provide the built-in integrations with external tools (the full list is in [executors.go](executors/executors.go), in `sexpTransforms`):

 * `ssh-login`: `(assert-no-command (external-sequential-interactive ssh))`
 * `ssh-exec-parallel`: `(assert-command (external-parallel ssh))`
 * `tmux-cssh`: `(assert-no-command (external-interactive tmux-cssh))`

## Contributing

All feedback and feature requests are welcome. Pull-requests are even more welcome :)
See [CONTRIBUTING.md](CONTRIBUTING.md) for a quick rundown of how and where to start.

## License

`easyssh` is licensed under [ISC](http://opensource.org/licenses/ISC). See the [LICENSE.txt](LICENSE.txt) file.
