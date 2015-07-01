package discoverers

import (
	"fmt"

	"github.com/abesto/easyssh/interfaces"
	"github.com/abesto/easyssh/util"
)

type firstMatching struct {
	args     []interface{}
	children []interfaces.Discoverer
}

func (d *firstMatching) Discover(input string) []string {
	util.RequireArgumentsAtLeast(d, 1, d.args)
	var hosts []string
	for _, discoverer := range d.children {
		util.Logger.Debugf("Trying discoverer %s", discoverer)
		hosts = discoverer.Discover(input)
		if len(hosts) > 0 {
			return hosts
		}
	}
	return []string{}
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
