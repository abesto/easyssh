# easyssh

Do magic like "run this command paralelly on all my machines matching this role in Chef" easily!

The syntax is slightly verbose, it's designed to be used in aliases you frequently need.

```sh
# log in with an interactive shell; old-fashioned ssh
easyssh myhost.com
# sequentially run "hostname" on all nodes matching "roles:app" according to knife
easyssh -c=ssh-exec -d=knife roles:app hostname
# parallelly reload apache on all your app servers (as root, of course)
easyssh -c=ssh-exec-parallel -d=knife -l=root roles:app /etc/init.d/apache2 reload
```
