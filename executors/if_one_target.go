package executors

import (
	"fmt"

	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type ifOneTarget struct {
	initialArgs []interface{}
	one         interfaces.Executor
	more        interfaces.Executor
}

func (e *ifOneTarget) Exec(targets []target.Target, command []string) {
	util.RequireArguments(e, 2, e.initialArgs)
	if len(targets) == 1 {
		util.Logger.Debugf("%s got one target, using %s", e, e.one)
		e.one.Exec(targets, command)
	} else {
		util.Logger.Debugf("%s got more than one target, using %s", e, e.more)
		e.more.Exec(targets, command)
	}
}
func (e *ifOneTarget) SetArgs(args []interface{}) {
	util.RequireArguments(e, 2, args)
	e.initialArgs = args
	e.one = makeFromSExp(args[0].([]interface{}))
	e.more = makeFromSExp(args[1].([]interface{}))
}
func (e *ifOneTarget) String() string {
	return fmt.Sprintf("<%s %s %s>", nameIfOneTarget, e.one, e.more)
}
