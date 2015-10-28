package executors

import (
	"fmt"
	"strings"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

const (
	externalModeSingleRun externalMode = iota
	externalModeSequential
	externalModeParallel
)

type external struct {
	initialArgs   []interface{}
	args          []string
	commandRunner util.InteractiveCommandRunner
	mode          externalMode
	interactive   bool
}

func (e *external) makeSingleRunJob(targets []target.Target, command []string) util.InteractiveCommandRunnerJob {
	return util.InteractiveCommandRunnerJob{
		Interactive: e.interactive,
		Label:       strings.Join(target.Strings(targets), " "),
		Argv:        append(e.args, append(target.Strings(targets), command...)...),
	}
}

func (e *external) makeJobPerTarget(targets []target.Target, command []string) []util.InteractiveCommandRunnerJob {
	jobs := make([]util.InteractiveCommandRunnerJob, len(targets))
	for i, target := range targets {
		jobs[i] = util.InteractiveCommandRunnerJob{
			Interactive: e.interactive,
			Label:       target.String(),
			Argv:        append(e.args, append([]string{target.String()}, command...)...),
		}
	}
	return jobs
}

func (e *external) Exec(targets []target.Target, command []string) {
	util.RequireArgumentsAtLeast(e, 1, e.initialArgs)
	if e.mode == externalModeSingleRun {
		e.commandRunner.Run(e.makeSingleRunJob(targets, command))
	} else if e.mode == externalModeSequential {
		for _, job := range e.makeJobPerTarget(targets, command) {
			e.commandRunner.Run(job)
		}
	} else if e.mode == externalModeParallel {
		util.Logger.Infof("Parallelly executing %s on %s", command, targets)
		e.commandRunner.RunParallel(e.makeJobPerTarget(targets, command))
	} else {
		util.Panicf("Unknown externalMode %s", e.mode)
	}
}

func (e *external) SetArgs(args []interface{}) {
	util.RequireArgumentsAtLeast(e, 1, args)
	e.initialArgs = args
	e.args = make([]string, len(args))
	for i, arg := range args {
		e.args[i] = string(arg.([]byte))
	}
}

func (e *external) String() string {
	rawName := "external"
	if e.mode == externalModeSequential {
		rawName += "-sequential"
	} else if e.mode == externalModeParallel {
		rawName += "-parallel"
	}
	if e.interactive {
		rawName += "-interactive"
	}
	return fmt.Sprintf("<%s %s>", rawName, e.args)
}
