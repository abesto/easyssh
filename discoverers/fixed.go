package discoverers

import (
	"fmt"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type fixed struct {
	args   []interface{}
	retval []target.Target
}

func (d *fixed) Discover(input string) []target.Target {
	util.RequireArgumentsAtLeast(d, 1, d.args)
	return d.retval
}

func (d *fixed) SetArgs(args []interface{}) {
	util.RequireArgumentsAtLeast(d, 1, args)
	d.args = args
	d.retval = target.FromStrings(util.ByteToStringArray(args)...)
}

func (d *fixed) String() string {
	return fmt.Sprintf("<%s %s>", nameFixed, d.retval)
}
