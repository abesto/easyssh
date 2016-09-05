# How to contribute

Use friendly open-source common sense: send pull-requests with clean, tested code. The description of the PR should
include (in decreasing order of importance) why you made the changes, what they are, and how they work. Please make
sure to run `gofmt -s -w .` and ideally `goimports -w .` before committing your changes. If you want to go the
extra mile, you can check whether you introduced any new problems according to `golint ./...`.

You can use [http://golang.org/misc/git/pre-commit](http://golang.org/misc/git/pre-commit) as your pre-commit hook
to make sure you don't forget `gofmt`.

## Dependencies, versions

`easyssh` is currently developed against Go 1.7, using [Glide](https://glide.sh/) to manage dependencies.

## Where to start

All [issues](https://github.com/abesto/easyssh/issues) are fair game. The milestones provide a rough order in which
the current maintainer (@abesto at the time of writing) plans to work on stuff, but it's all subject to change.

## Licensing

`easyssh` is [licensed](LICENSE.txt) under [ISC](http://opensource.org/licenses/ISC).

