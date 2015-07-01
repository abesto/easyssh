package discoverers

import (
	"fmt"

	"github.com/abesto/easyssh/util"
)

type fixed struct {
	args   []interface{}
	retval []string
}

func (d *fixed) Discover(input string) []string {
	util.RequireArgumentsAtLeast(d, 1, d.args)
	return d.retval
}
func (d *fixed) SetArgs(args []interface{}) {
	util.RequireArgumentsAtLeast(d, 1, args)
	d.args = args
	d.retval = make([]string, len(args))
	for i, x := range args {
		d.retval[i] = string(x.([]byte))
	}
}
func (d *fixed) String() string {
	return fmt.Sprintf("<%s %s>", nameFixed, d.retval)
}
