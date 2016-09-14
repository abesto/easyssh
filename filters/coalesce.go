package filters

import (
	"fmt"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type coalesce struct {
	args          []interface{}
	coalesceOrder []string
}

func (f *coalesce) Filter(targets []target.Target) []target.Target {
	util.RequireArgumentsAtLeast(f, 1, f.args)
	newTargets := make([]target.Target, len(targets))
	for i, t := range targets {
		newTargets[i] = t
		newTargets[i].CoalesceOrder = f.coalesceOrder
	}
	util.Logger.Debugf("Set CoalesceOrder of all targets to %s", f.coalesceOrder)
	return newTargets
}

func (f *coalesce) SetArgs(args []interface{}) {
	util.RequireArgumentsAtLeast(f, 1, args)
	f.args = args
	f.coalesceOrder = util.ByteToStringArray(args)
	for i, coalescer := range f.coalesceOrder {
		if _, ok := target.Coalescers[coalescer]; !ok {
			util.Panicf("Unknown target coalescer %s (index %d) in filter %s", coalescer, i, f)
		}
	}
}

func (f *coalesce) String() string {
	return fmt.Sprintf("<%s %s>", nameCoalesce, f.coalesceOrder)
}
