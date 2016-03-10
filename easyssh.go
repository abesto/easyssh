package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/abesto/easyssh/discoverers"
	"github.com/abesto/easyssh/executors"
	"github.com/abesto/easyssh/filters"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
	"github.com/alexcesaro/log/stdlog"
)

var VERSION = "HEAD"
var BUILD_DATE = "LOCAL"

func main() {
	var (
		discovererDefinition string
		discoverer           interfaces.Discoverer
		executorDefinition   string
		executor             interfaces.Executor
		user                 string
		filterDefinition     string
		filter               interfaces.TargetFilter
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %s [options] target-definition [command]

Where
  target-definition is the input to the discoverer(s) defined with -d
  command, if provided, will be run on the targets

Ideally a single alias should cover all your use-cases. For example:
  smartssh_executor='(if-command (ssh-exec-parallel) (if-one-target (ssh-login) (tmux-cssh)))'
  smartssh_discoverer='(first-matching (knife) (comma-separated))'
  smartssh_filter='(list (ec2-instance-id us-east-1) (ec2-instance-id us-west-1))'
  alias s="%s -e='$smartssh_executor' -d='$smartssh_discoverer' -f='$smartssh_filter'"

Configuration details:
  open https://github.com/abesto/smartssh/blob/master/README.md#configuration

Options:
`, os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&user, "l", "",
		"Specifies the user to log in as on the remote machine.")
	flag.StringVar(&discovererDefinition, "d", "(comma-separated)",
		fmt.Sprintf("Discoverer definition. Supported discoverers: %s", strings.Join(discoverers.SupportedDiscovererNames(), ", ")))
	flag.StringVar(&executorDefinition, "e", "(ssh-login)",
		fmt.Sprintf("Executor definition. Supported executors: %s", strings.Join(executors.SupportedExecutorNames(), ", ")))
	flag.StringVar(&filterDefinition, "f", "(id)",
		fmt.Sprintf("Filter definition. Supported filters: %s", strings.Join(filters.SupportedFilterNames(), ", ")))
	verbose := flag.Bool("v", false, "Verbose output (alias of '-log debug')")
	versionRequested := flag.Bool("V", false, "Display the version number and exit")
	flag.Parse()

	if *versionRequested {
		fmt.Printf("easyssh version %s build %s\n", VERSION, BUILD_DATE)
		return
	}

	if *verbose {
		flag.Set("log", "debug")
	}
	util.Logger = stdlog.GetFromFlags()
	logger := util.Logger

	if flag.NArg() == 0 {
		logger.Critical("Required argument for target host lookup missing")
		flag.Usage()
		return
	}

	defer func() {
		if err := recover(); err != nil {
			// discoverer, executor and filter are created in this order
			// if at least one of them is nil, then the creation of the first one that is nil has generated the error.
			if discoverer == nil {
				util.Logger.Critical("Failed to create discoverer")
			} else if executor == nil {
				util.Logger.Critical("Failed to create executor")
			} else if filter == nil {
				util.Logger.Critical("Failed to create filter")
			}
			switch err.(type) {
			case string:
				util.Logger.Critical(err.(string))
				os.Exit(1)
			default:
				panic(err)
			}
		}
	}()

	discoverer = discoverers.Make(discovererDefinition)
	executor = executors.Make(executorDefinition)
	filter = filters.Make(filterDefinition)

	targets := discoverer.Discover(flag.Arg(0))
	if len(targets) == 0 {
		util.Panicf("No targets found")
	}

	for i := range targets {
		targets[i].User = user
	}

	logger.Debugf("Targets before filters: %s", targets)
	targets = filter.Filter(targets)
	logger.Infof("Targets: %s", targets)

	command := flag.Args()[1:]
	executor.Exec(targets, command)
}
