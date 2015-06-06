# easyssh

Do magic like "run this command paralelly on all my machines matching this role in Chef" easily!

The syntax is slightly verbose, it's designed to be used in aliases you frequently need.

## Simple usage
```sh
# log in with an interactive shell; old-fashioned ssh
easyssh myhost.com
# sequentially run "hostname" on all nodes matching "roles:app" according to knife
easyssh -c=ssh-exec -d=knife roles:app hostname
# parallelly reload apache on all your app servers (as root, of course)
easyssh -c=ssh-exec-parallel -d=knife -l=root roles:app /etc/init.d/apache2 reload
```

## Recommended alias
This one alias will

 * look up your target hosts using `knife search node`, taking the first argument as the search query
 * if that doesn't find anything, it'll assume that you passed in a comma-separated list of hosts in the first argument
 * if there are no further arguments, then
  * if there is just one matched node, then log in
  * if there are more nodes, then it will log in using `tmux-cssh` (you can replace `tmux-cssh` with `csshx` if you want)
 * if there are further arguments, then it will run those as a command, on all matched nodes, paralelly. You can replace
   `ssh-exec-parallel` with `ssh-exec` to run the command on just a single node at a time.

```sh
alias s='easyssh -c=one-or-more:ssh-login:tmux-cssh -cc=ssh-exec-parallel -d=first-matching:knife:comma-separated'
# log in to myhost.com
s myhost.com
# reload apache on app servers (as root)
s -l root roles:app /etc/init.d/apache2 reload
```

If you frequently log in to servers as root, you can then go:

```sh
alias sr='s -l root'
# reload apache on app servers (as root)
sr roles:app /etc/init.d/apache2 reload
```