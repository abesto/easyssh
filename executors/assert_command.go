package executors

import (
	"fmt"

	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type assertCommand struct {
	initialArgs []interface{}
	require     bool
	child       interfaces.Executor
}

func (e *assertCommand) Exec(targets []target.Target, command []string) {
	util.RequireArguments(e, 1, e.initialArgs)
	if e.require {
		if len(command) == 0 {
			util.Panicf("%s requires a command.", e)
		}
	} else {
		if len(command) > 0 {
			util.Panicf("%s doesn't accept a command, got: %s", e, command)
		}
	}
	e.child.Exec(targets, command)
}

func (e *assertCommand) SetArgs(args []interface{}) {
	util.RequireArguments(e, 1, args)
	e.initialArgs = args
	e.child = makeFromSExp(args[0].([]interface{}))
}

func (e *assertCommand) String() string {
	var rawName string
	if e.require {
		rawName = nameAssertCommand
	} else {
		rawName = nameAssertNoCommand
	}
	return fmt.Sprintf("<%s %s>", rawName, e.child)
}

type externalMode byte
