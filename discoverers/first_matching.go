package discoverers

import (
	"fmt"

	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type firstMatching struct {
	args     []interface{}
	children []interfaces.Discoverer
}

func (d *firstMatching) Discover(input string) []target.Target {
	util.RequireArgumentsAtLeast(d, 1, d.args)
	var targets []target.Target
	for _, discoverer := range d.children {
		util.Logger.Debugf("Trying discoverer %s", discoverer)
		targets = discoverer.Discover(input)
		if len(targets) > 0 {
			break
		}
	}
	return targets
}

func (d *firstMatching) SetArgs(args []interface{}) {
	util.RequireArgumentsAtLeast(d, 1, args)
	d.args = args
	d.children = []interfaces.Discoverer{}
	for _, exp := range args {
		d.children = append(d.children, makeFromSExp(exp.([]interface{})))
	}
}

func (d *firstMatching) String() string {
	return fmt.Sprintf("<%s %s>", nameFirstMatching, d.children)
}
