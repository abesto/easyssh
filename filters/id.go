package filters

import (
	"fmt"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type id struct{}

func (f *id) Filter(targets []target.Target) []target.Target {
	return targets
}
func (f *id) SetArgs(args []interface{}) {
	util.RequireNoArguments(f, args)
}
func (f *id) String() string {
	return fmt.Sprintf("<%s>", nameId)
}
