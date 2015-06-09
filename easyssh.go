package main

import (
	"flag"
	"github.com/abesto/easyssh/discoverers"
	"github.com/abesto/easyssh/executors"
	"github.com/abesto/easyssh/filters"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
	"github.com/alexcesaro/log/stdlog"
)

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

	flag.StringVar(&user, "l", "",
		"Specifies the user to log in as on the remote machine. If empty, it will not be passed to the called SSH tool.")
	// TODO document what Discoverer mechanisms and command runners are available
	flag.StringVar(&discovererDefinition, "d", "(comma-separated)", "")
	flag.StringVar(&executorDefinition, "e", "(ssh-login)", "")
	flag.StringVar(&filterDefinition, "f", "(id)", "")
	flag.Parse()

	util.Logger = stdlog.GetFromFlags()
	var logger = util.Logger

	if flag.NArg() == 0 {
		util.Abort("Required argument for target host lookup missing")
	}

	discoverer = discoverers.Make(discovererDefinition)
	executor = executors.Make(executorDefinition)
	filter = filters.Make(filterDefinition)

	var targets []target.Target = []target.Target{}
	for _, host := range discoverer.Discover(flag.Arg(0)) {
		targets = append(targets, target.Target{Host: host, User: user})
	}
	if len(targets) == 0 {
		util.Abort("No targets found")
	}

	logger.Debugf("Targets before filters: %s", targets)
	targets = filter.Filter(targets)
	logger.Infof("Targets: %s", targets)

	var command = flag.Args()[1:]
	executor.Exec(targets, command)
}
