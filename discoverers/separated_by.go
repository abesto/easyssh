package discoverers

import (
	"fmt"
	"strings"

	"github.com/abesto/easyssh/target"
	"github.com/abesto/easyssh/util"
)

type separatedBy struct {
	args []interface{}
	sep  string
}

func (d *separatedBy) Discover(input string) []target.Target {
	strs := strings.Split(input, d.sep)
	notEmpty := make([]string, 0, len(strs))
	for _, str := range strs {
		if len(str) > 0 {
			notEmpty = append(notEmpty, str)
		}
	}
	return target.FromStrings(notEmpty...)
}

func (d *separatedBy) SetArgs(args []interface{}) {
	util.RequireArguments(d, 1, args)
	d.args = args
	d.sep = string(args[0].([]byte))
}

func (d *separatedBy) String() string {
	return fmt.Sprintf("<%s %s>", nameSeparatedBy, d.sep)
}
