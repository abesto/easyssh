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
		DiscovererName    string
		Discoverer        interfaces.Discoverer
		commandDefinition string
		command           interfaces.Command
		user              string
		filterNames       string
	)

	flag.StringVar(&user, "l", "",
		"Specifies the user to log in as on the remote machine. If empty, it will not be passed to the called SSH tool.")
	// TODO document what Discoverer mechanisms and command runners are available
	flag.StringVar(&DiscovererName, "d", "comma-separated", "")
	flag.StringVar(&commandDefinition, "c", "ssh-login", "")
	flag.StringVar(&filterNames, "f", "", "")
	flag.Parse()

	if flag.NArg() == 0 {
		util.Abort("Required argument for target host lookup missing")
	}

	Discoverer = discoverers.Make(DiscovererName)

	var commandArgs = []string{}
	if flag.NArg() > 0 {
		commandArgs = flag.Args()[1:]
	}

	command = commands.Make(commandDefinition)

	var targets []target.Target = []target.Target{}
	for _, host := range Discoverer.Discover(flag.Arg(0)) {
		targets = append(targets, target.Target{Host: host, User: user})
	}
	if len(targets) == 0 {
		util.Abort("No targets found")
	}
	fmt.Printf("Targets: %s\n", targets)

	if len(filterNames) > 0 {
		targets = filters.ApplyFilters(filterNames, targets)
	}

	command.Exec(targets, commandArgs)
}
