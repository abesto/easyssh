package executors

import (
	"fmt"

	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type ifCommand struct {
	initialArgs    []interface{}
	withCommand    interfaces.Executor
	withoutCommand interfaces.Executor
}

func (e *ifCommand) Exec(targets []target.Target, args []string) {
	util.RequireArguments(e, 2, e.initialArgs)
	if len(args) < 1 {
		util.Logger.Debugf("%s got no command, using %s", e, e.withoutCommand)
		e.withoutCommand.Exec(targets, args)
	} else {
		util.Logger.Debugf("%s got command, using %s", e, e.withCommand)
		e.withCommand.Exec(targets, args)
	}
}

func (e *ifCommand) SetArgs(args []interface{}) {
	util.RequireArguments(e, 2, args)
	e.withCommand = makeFromSExp(args[0].([]interface{}))
	e.withoutCommand = makeFromSExp(args[1].([]interface{}))
	e.initialArgs = args
}

func (e *ifCommand) String() string {
	return fmt.Sprintf("<%s %s %s>", nameIfCommand, e.withCommand, e.withoutCommand)
}
