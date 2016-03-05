package executors

import (
	"sort"

	"github.com/abesto/easyssh/fromsexp"
	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
)

// Make creates an Executor by name
func Make(input string) interfaces.Executor {
	return fromsexp.MakeFromString(input, sexpTransforms, makeByName).(interfaces.Executor)
}

// SupportedExecutorNames returns the names Make can take
func SupportedExecutorNames() []string {
	names := make([]string, len(executorMakerMap)+len(sexpTransforms))

	// High-level executors
	for i := 0; i < len(sexpTransforms); i++ {
		names[i] = sexpTransforms[i].Name
	}

	// Low-level executors
	i := len(sexpTransforms)
	for key := range executorMakerMap {
		names[i] = key
		i++
	}

	sort.Strings(names)
	return names
}

func makeFromSExp(data []interface{}) interfaces.Executor {
	return fromsexp.Make(data, sexpTransforms, makeByName).(interfaces.Executor)
}

const (
	nameAssertCommand                 = "assert-command"
	nameAssertNoCommand               = "assert-no-command"
	nameExternal                      = "external"
	nameExternalInteractive           = "external-interactive"
	nameExternalSequential            = "external-sequential"
	nameExternalSequentialInteractive = "external-sequential-interactive"
	nameExternalParallel              = "external-parallel"
	nameIfOneTarget                   = "if-one-target"
	nameIfCommand                     = "if-command"
)

var executorMakerMap = map[string]func() interfaces.Executor{
	nameIfOneTarget:     func() interfaces.Executor { return &ifOneTarget{} },
	nameIfCommand:       func() interfaces.Executor { return &ifCommand{} },
	nameAssertCommand:   func() interfaces.Executor { return &assertCommand{require: true} },
	nameAssertNoCommand: func() interfaces.Executor { return &assertCommand{require: false} },
	nameExternal: func() interfaces.Executor {
		return &external{
			commandRunner: &util.RealInteractiveCommandRunner{},
			mode:          externalModeSingleRun,
		}
	},
	nameExternalSequential: func() interfaces.Executor {
		return &external{
			commandRunner: &util.RealInteractiveCommandRunner{},
			mode:          externalModeSequential,
		}
	},
	nameExternalParallel: func() interfaces.Executor {
		return &external{
			commandRunner: &util.RealInteractiveCommandRunner{},
			mode:          externalModeParallel,
		}
	},
	nameExternalInteractive: func() interfaces.Executor {
		return &external{
			commandRunner: &util.RealInteractiveCommandRunner{},
			mode:          externalModeSingleRun,
			interactive:   true,
		}
	},
	nameExternalSequentialInteractive: func() interfaces.Executor {
		return &external{
			commandRunner: &util.RealInteractiveCommandRunner{},
			mode:          externalModeSequential,
			interactive:   true,
		}
	},
}

var r = fromsexp.Replace
var sexpTransforms = []fromsexp.SexpTransform{
	r("(if-args)", "(if-command)"),
	r("(ssh-login)", "(assert-no-command (external-sequential-interactive ssh))"),
	r("(ssh-exec)", "(ssh-exec-sequential)"),
	r("(ssh-exec-sequential)", "(assert-command (external-sequential ssh))"),
	r("(ssh-exec-parallel)", "(assert-command (external-parallel ssh))"),
	r("(csshx)", "(assert-no-command (external-interactive csshx))"),
	r("(tmux-cssh)", "(assert-no-command (external-interactive tmux-cssh -ns))"),
}

func makeByName(name string) interface{} {
	var d interfaces.Executor
	for key, maker := range executorMakerMap {
		if key == name {
			d = maker()
		}
	}
	if d == nil {
		util.Panicf("Executor \"%s\" is not known", name)
	}
	return d
}
