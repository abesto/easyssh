package filters

import (
	"fmt"

	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type list struct {
	args     []interface{}
	children []interfaces.TargetFilter
}

func (f *list) Filter(targets []target.Target) []target.Target {
	util.RequireArgumentsAtLeast(f, 1, f.args)
	for _, child := range f.children {
		targets = child.Filter(targets)
		util.Logger.Debugf("Targets after filter %s: %s", child, targets)
	}
	return targets
}
func (f *list) SetArgs(args []interface{}) {
	util.RequireArgumentsAtLeast(f, 1, args)
	f.args = args
	f.children = make([]interfaces.TargetFilter, len(args))
	for i, def := range args {
		f.children[i] = makeFromSExp(def.([]interface{}))
	}
}
func (f *list) String() string {
	return fmt.Sprintf("<%s %s>", nameList, f.children)
}
