package main

import (
	"flag"
	"fmt"
	"github.com/abesto/easyssh/commands"
	"github.com/abesto/easyssh/discoverers"
	"github.com/abesto/easyssh/filters"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

func main() {
	var (
		discovererDefinition string
		discoverer interfaces.Discoverer
		commandDefinition string
		command           interfaces.Command
		user              string
		filterDefinition string
		filter interfaces.TargetFilter
	)

	flag.StringVar(&user, "l", "",
		"Specifies the user to log in as on the remote machine. If empty, it will not be passed to the called SSH tool.")
	// TODO document what Discoverer mechanisms and command runners are available
	flag.StringVar(&discovererDefinition, "d", "comma-separated", "")
	flag.StringVar(&commandDefinition, "c", "ssh-login", "")
	flag.StringVar(&filterDefinition, "f", "id", "")
	flag.Parse()

	if flag.NArg() == 0 {
		util.Abort("Required argument for target host lookup missing")
	}

	discoverer = discoverers.Make(discovererDefinition)
	fmt.Printf("Discoverer built: %s\n", discoverer)

	command = commands.Make(commandDefinition)
	fmt.Printf("Command built: %s\n", command)

	filter = filters.Make(filterDefinition)
	fmt.Printf("Filter built: %s\n", filter)

	var targets []target.Target = []target.Target{}
	for _, host := range discoverer.Discover(flag.Arg(0)) {
		targets = append(targets, target.Target{Host: host, User: user})
	}
	if len(targets) == 0 {
		util.Abort("No targets found")
	}
	fmt.Printf("Targets: %s\n", targets)

	targets = filter.Filter(targets)

	var commandArgs = []string{}
	if flag.NArg() > 0 {
		commandArgs = flag.Args()[1:]
	}

	command.Exec(targets, commandArgs)
}
