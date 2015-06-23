package filters

import (
	"fmt"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type first struct{}

func (f *first) Filter(targets []target.Target) []target.Target {
	if len(targets) > 0 {
		return targets[0:1]
	}
	return targets
}
func (f *first) SetArgs(args []interface{}) {
	util.RequireNoArguments(f, args)
}
func (f *first) String() string {
	return fmt.Sprintf("<%s>", nameFirst)
}
